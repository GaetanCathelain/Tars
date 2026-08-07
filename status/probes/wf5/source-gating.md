# WF5 probe — inbound Slack THREAD-REPLY gating (source read)

**Date**: 2026-08-07 · **Mode**: READ-ONLY on the Tars VM (no edits, no restarts,
no Slack sends). Files were `scp`-copied to cooper for reading; the VM copies were
not touched.

## 0. Resolved package path

`~/.local/bin/hermes` is a 159-byte bash wrapper (not a pipx/uv tool venv):

```bash
#!/usr/bin/env bash
unset PYTHONPATH
unset PYTHONHOME
exec "/home/gaetan/.hermes/hermes-agent/venv/bin/python" "/home/gaetan/.hermes/hermes-agent/hermes" "$@"
```

So Hermes runs from a **source checkout**, not site-packages:

| Thing | Path (on the VM) |
|---|---|
| Package root | `/home/gaetan/.hermes/hermes-agent/` |
| Git HEAD | `6e87d43a5765d4480359cfa57b574e97bc5eacf4` (Sun Jul 12 02:38:01 2026 +0530) |
| **Slack adapter** | `/home/gaetan/.hermes/hermes-agent/plugins/platforms/slack/adapter.py` (402 419 B, 9 101 lines) |
| Gateway runner | `/home/gaetan/.hermes/hermes-agent/gateway/run.py` |
| Allowlist logic | `/home/gaetan/.hermes/hermes-agent/gateway/authz_mixin.py` |
| Config loader | `/home/gaetan/.hermes/hermes-agent/gateway/config.py` |
| Upstream test proving the thread knob | `/home/gaetan/.hermes/hermes-agent/tests/test_slack_thread_require_mention.py` |

All `file:line` citations below are against those VM paths (byte-identical copies were read).

---

## 1. Where the inbound handler / event gate lives

**Registration** — `plugins/platforms/slack/adapter.py:1951-1965`:

```python
            # Register message event handler
            @self._app.event("message")
            async def handle_message_event(event, say, body):
                await self._handle_slack_message(event, body)
            ...
            @self._app.event("app_mention")
            async def handle_app_mention(event, say, body):
                await self._handle_slack_message(event, body)
```

**The gate itself** — `SlackAdapter._handle_slack_message`,
`plugins/platforms/slack/adapter.py:5239` (`async def _handle_slack_message(self, event, payload=None)`).
Everything in §3 happens inside this one function. Reactions
(`_handle_slack_reaction`, :4773) and slash commands (`_handle_slash_command`, :7642)
are separate entry points and are out of scope here.

The helper that owns *all* thread-awareness is
`SlackAdapter._should_wake_on_unmentioned_message`, `adapter.py:5164`.

---

## 2. Thread-awareness of the mention requirement — YES, extensively

There **is** a `thread_ts` check, a bot-participation check, a session lookup keyed
by thread ts, and a dedicated config knob. Every gating-relevant hit:

### 2a. `thread_ts` is read and classified — `adapter.py:5622-5623`

```python
        event_thread_ts = event.get("thread_ts")
        is_thread_reply = bool(event_thread_ts and event_thread_ts != ts)
```

(`thread_ts == ts` is deliberately *not* a reply — a thread-root shape.)

### 2b. The "was I already in this thread" check — `adapter.py:5164-5237`

Five independent wake conditions, verbatim (trimmed):

```python
    async def _should_wake_on_unmentioned_message(
        self, event_thread_ts, channel_id, user_id, is_thread_reply,
        team_id="", chat_type="group",
    ) -> bool:
        """Return True if the bot should wake on an un-mentioned message.

        Combines the four wake checks:
          1. _bot_message_ts           (thread root was sent by us via send())
          2. _mentioned_threads        (someone @-mentioned us earlier)
          3. _has_active_session...    (there's already an agent session)
          4. _bot_authored_thread_root (#63530: the bot posted the thread root
             via direct chat.postMessage, outside the gateway send() path ...)
        """
        if not event_thread_ts:
            return False                                      # 5187-5188
        thread_marker = self._workspace_message_marker(team_id, event_thread_ts)
        if is_thread_reply and (
            thread_marker in self._bot_message_ts
            or event_thread_ts in self._bot_message_ts
        ):
            return True                                       # 5193-5197
        if (
            thread_marker in self._mentioned_threads
            or event_thread_ts in self._mentioned_threads
        ):
            return True                                       # 5198-5202
        if is_thread_reply and self._has_active_session_for_thread(
            channel_id=channel_id, thread_ts=event_thread_ts,
            user_id=user_id, team_id=team_id, chat_type=chat_type,
        ):
            return True                                       # 5203-5210
        if is_thread_reply and await self._bot_authored_thread_root(
            channel_id=channel_id, thread_ts=event_thread_ts, team_id=team_id,
        ):
            return True                                       # 5212-5217
        if is_thread_reply:                                   # 5223-5236
            bot_uid = self._team_bot_user_ids.get(team_id, self._bot_user_id)
            if bot_uid:
                parent_text = await self._fetch_thread_parent_text(...)
                if parent_text and f"<@{bot_uid}>" in parent_text:
                    if not self._slack_strict_mention():
                        self._register_mentioned_thread(event_thread_ts)
                    return True
        return False
```

- `_bot_message_ts` — thread roots we sent (in-memory).
- `_mentioned_threads` — in-memory set, cap 5000 (`adapter.py:976`), registered at
  `adapter.py:5096` / written at `adapter.py:5768`.
- `_has_active_session_for_thread` — `adapter.py:7972`; **existing-session lookup
  keyed by thread ts**, survives gateway restarts (persistent session store).
- `_bot_authored_thread_root` — `adapter.py:5111`; derived from the Slack API,
  so it also survives restarts.
- Parent-text mention probe — Slack API read of the thread root.

**This entire helper is unreachable in the current Tars configuration** (see §3).

### 2c. The config knob that exists: `thread_require_mention`

Accessor — `adapter.py:8303-8322`:

```python
    def _slack_thread_require_mention(self) -> bool:
        """When true, Slack thread replies require an explicit @-mention.

        This is narrower than ``strict_mention``: top-level channel messages can
        still be processed without a mention when ``require_mention`` is false
        or the channel is listed in ``free_response_channels``. ...
        """
        configured = self.config.extra.get("thread_require_mention")
        if configured is not None:
            if isinstance(configured, str):
                return configured.lower() in {"true", "1", "yes", "on"}
            return bool(configured)
        return os.getenv("SLACK_THREAD_REQUIRE_MENTION", "false").lower() in {
            "true", "1", "yes", "on",
        }
```

Note the polarity: `thread_require_mention` **tightens** gating (it is the knob to
make threads *stricter*). It is `false` by default and is **not set** on Tars.
There is no inverse knob — see §5.

### 2d. Grep sweep for the hypothesised knob names

`grep -rn 'thread_follow|follow_threads|mention_once_per_thread|respond_to_thread_replies|thread_reply_without_mention|require_mention_in_thread|auto_follow'`
over the whole checkout (`*.py`, `*.yaml`, `*.md`, excluding `node_modules`) returns
**one** file: `tests/gateway/test_slack.py` (incidental). **None of those knobs exist.**

### 2e. Other thread-touching config (not gating, listed for completeness)

| Key | file:line | Effect |
|---|---|---|
| `reply_in_thread` (default `True`) | `adapter.py:3092`, `3158`, `5590` | Outbound threading + session scoping: when true a top-level channel message becomes its own synthetic thread root (`thread_ts = ts`, :5593). Does **not** gate inbound. |
| `dm_top_level_threads_as_sessions` (default `True`) | `adapter.py:3045` | DM-only session keying. |
| `assistant_thread_titles` (default `True`) | `adapter.py:4563` | Cosmetic. |
| `reply_broadcast` (default `False`) | `adapter.py:2543` | Outbound only. |
| `thread_sessions_per_user` (default `False`) | `gateway/config.py:1208` | Gateway-wide session keying, not a gate. |

---

## 3. GATE ORDER for an inbound message

All inside `_handle_slack_message` unless noted. Ordered, with line numbers:

| # | Gate | file:line | Notes |
|---|---|---|---|
| 0 | DEBUG entry log (fires before all filtering) | `adapter.py:5250-5264` | metadata only |
| 1 | `subtype == "message_changed"` normalisation | `adapter.py:5265-5297` | edits re-enter the pipeline once |
| 2 | **Dedup** (workspace-scoped event ts) | `adapter.py:5303-5308` | Socket-Mode redelivery |
| 3 | **Ignored-channel blacklist** (`_is_ignored_channel`, :2436) | `adapter.py:5311-5313` | `SLACK_IGNORED_CHANNELS` |
| 4 | **Bot/app-sender filter #1** (`allow_bots`, default `"none"`) | `adapter.py:5329-5351` | `return` on `"none"` |
| 5 | **Self-message filter** (`msg_user == self._bot_user_id`) | `adapter.py:5353-5354` | inside the bot branch |
| 6 | `subtype == "message_deleted"` | `adapter.py:5359-5360` | |
| 7 | text/blocks/attachments enrichment | `adapter.py:5362-5478` | no gating |
| 8 | **DM vs channel branch** (`is_dm`, `is_one_to_one_dm`) | `adapter.py:5509-5528` | `channel_type in {"im","mpim"}`; MPIM is treated as a channel for gating |
| 9 | **`disable_dms`** | `adapter.py:5513-5519` | |
| 10 | **★ ALLOWLIST — early reject** | **`adapter.py:5534-5550`** | emits the `[Slack] Early reject of unauthorized user %s in channel %s` WARNING |
| 11 | thread_ts computed for session keying | `adapter.py:5559-5601` | |
| 12 | `is_mentioned` / `is_thread_reply` computed | `adapter.py:5615-5623` | |
| 13 | **Bot/app-sender filter #2** (resolved bot *users*) | `adapter.py:5639-5652` | |
| 14 | *(channel-only block starts)* `if not is_one_to_one_dm and bot_uid:` | `adapter.py:5654` | 1:1 DMs skip everything below |
| 15 | **`allowed_channels` whitelist** | `adapter.py:5656-5661` | |
| 16 | **`ignore_other_user_mentions`** | `adapter.py:5668-5678` | |
| 17 | `force_process` (internal reaction routing) | `adapter.py:5680-5681` | bypasses mention gate only |
| 18 | **free-response branch** (`free_response_channels` or `require_mention=false`, unless `require_mention_channels`) — then `thread_require_mention` sub-gate | `adapter.py:5682-5706` | |
| 19 | **`strict_mention` and not mentioned → `return`** | **`adapter.py:5707-5708`** | ← **the gate Tars currently hits** |
| 20 | `thread_require_mention` and thread reply and not mentioned → `return` | `adapter.py:5709-5720` | |
| 21 | **not mentioned → `_should_wake_on_unmentioned_message(...)`** (all thread-awareness) | `adapter.py:5721-5730` | **unreachable while #19 is on** |
| 22 | mention stripping + `_register_mentioned_thread` (skipped under strict/thread-gated) | `adapter.py:5732-5768` | |
| 23 | thread context hydration | `adapter.py:5793-5900+` | post-gate |
| 24 | *(gateway)* **allowlist re-check, cold path** `_handle_message` | `gateway/run.py:14483`, check at `14607` / `14610` | `_is_user_authorized` |
| 25 | *(gateway)* **allowlist re-check, busy path** `_handle_active_session_busy_message` | `gateway/run.py:8853`, check at `8859` | "Dropping message from unauthorized user in active session" |

### The critical ordering answer

**YES — the allowlist is evaluated strictly BEFORE any mention or thread logic.**

Adapter-side early reject, verbatim `adapter.py:5530-5550`:

```python
        # Reject unauthorized users before thread lookups, name resolution,
        # or file downloads.  The final gateway runner auth check happens
        # after MessageEvent construction, so adapter-side media fetches need
        # the same auth chain up front.
        _runner = getattr(getattr(self, "_message_handler", None), "__self__", None)
        _auth_fn = getattr(_runner, "_is_user_authorized", None)
        if user_id and callable(_auth_fn):
            _source = self.build_source(
                chat_id=channel_id, chat_name="",
                chat_type="dm" if is_dm else "group",
                user_id=user_id, user_name="",
            )
            if not _auth_fn(_source):
                logger.warning(
                    "[Slack] Early reject of unauthorized user %s in channel %s",
                    user_id, channel_id,
                )
                return
```

Line 5550 (`return`) is **104 lines before** the first mention/thread branch
(line 5654). The mention chain is an `elif` ladder that can only *drop* messages —
it never re-admits a sender. So no change confined to lines 5654-5730 can widen who
gets through: `SLACK_ALLOWED_USERS` is a hard precondition.

Defence in depth, three layers:
1. `adapter.py:5544` (pre-dispatch, adapter).
2. `gateway/run.py:14610` (cold path, after `MessageEvent` construction).
3. `gateway/run.py:8859` (busy path — explicitly added so "unauthorized users in
   shared threads can [not] inject messages into an active session they don't own").

The allowlist resolution itself is `gateway/authz_mixin.py:386` `_is_user_authorized`;
the Slack env var is wired at `authz_mixin.py:512` (`Platform.SLACK: "SLACK_ALLOWED_USERS"`)
and re-declared by the plugin at `adapter.py:9086` (`allowed_users_env="SLACK_ALLOWED_USERS"`,
`allow_all_env="SLACK_ALLOW_ALL_USERS"`). Default is **deny** (`authz_mixin.py:395`,
and `if not user_id: return False` at `:504-505`).

One residual note: the adapter early reject is skipped when `user_id` is falsy
(`if user_id and callable(_auth_fn)`, :5536). Those events are still caught by
`run.py:14607`/`14610` and, for bot senders, by the `allow_bots == "none"` return at
`adapter.py:5339-5340`. No hole.

---

## 4. Knob inventory — every Slack config key the adapter reads

Two-tier precedence in every accessor: **`PlatformConfig.extra[key]` wins over the
env var**, env var wins over the built-in default. Separately,
`_apply_yaml_config` (`adapter.py:8976-9042`) translates `config.yaml` → `SLACK_*`
env vars, but **only when the env var is not already set** (`not os.getenv(...)`),
so a real env var in `~/.hermes/.env` beats `config.yaml`.

### Gating knobs

| YAML key (`gateway.platforms.slack.*`) | Env var | Default | Accessor file:line | YAML→env bridge | Effect |
|---|---|---|---|---|---|
| `require_mention` | `SLACK_REQUIRE_MENTION` | **`true`** (explicit-false parsing) | `adapter.py:8244-8261` | `:8992-8993` | channel messages need a mention |
| `strict_mention` | `SLACK_STRICT_MENTION` | `false` | `adapter.py:8263-8278` | `:8994-8995` | **kills all auto-triggers** (thread memory, bot-message follow-up, session presence) |
| `thread_require_mention` | `SLACK_THREAD_REQUIRE_MENTION` | `false` | `adapter.py:8303-8322` | `:9000-9005` | thread replies need a mention (narrower than strict) |
| `ignore_other_user_mentions` | `SLACK_IGNORE_OTHER_USER_MENTIONS` | `false` | `adapter.py:8280-8301` | `:8996-8999` | stay silent on messages opening with someone else's `@` |
| `free_response_channels` | `SLACK_FREE_RESPONSE_CHANNELS` | `""` (empty set) | `adapter.py:8355-8371` | `:9008-9012` | per-channel mention exemption |
| `require_mention_channels` | `SLACK_REQUIRE_MENTION_CHANNELS` | `""` | `adapter.py:8404-8421` | `:9013-9017` | per-channel force-mention (overrides free-response) |
| `mention_patterns` | `SLACK_MENTION_PATTERNS` | `None` | `adapter.py:8423-8469` (used at `:8471`) | bridged into `extra` at `config.py:1563-1564` | regex wake-words counted as a mention |
| `allowed_channels` | `SLACK_ALLOWED_CHANNELS` | `""` | `adapter.py:8387-8402` | `:9031-9035` | channel whitelist |
| `ignored_channels` | `SLACK_IGNORED_CHANNELS` | `""` | `adapter.py:2429` (`_is_ignored_channel`, `:2436`) | `:9036-9041` | channel blacklist |
| `disable_dms` | `SLACK_DISABLE_DMS` | `false` | `adapter.py:8373-8385` | `:9029-9030` | drop all DMs |
| `allow_bots` | `SLACK_ALLOW_BOTS` | `"none"` | `adapter.py:3106-3110` | `:9006-9007` | `none` / `mentions` / `all` |
| — (env only) | `SLACK_ALLOWED_USERS` | unset → deny | `authz_mixin.py:512`, registered `adapter.py:9086` | — | **the allowlist** |
| — (env only) | `SLACK_ALLOW_ALL_USERS` | `""` | `adapter.py:6686`, `authz_mixin.py:539`/`9087` | — | open workspace access |
| `unauthorized_dm_behavior` | — | `"pair"` | `config.py:1535-1539`, read at `config.py:1228` | bridged into `extra` | pair vs ignore on unauthorized DM |

### Non-gating knobs the adapter reads

| YAML key | Env var | Default | file:line |
|---|---|---|---|
| `reply_in_thread` | — (extra only) | `true` | `adapter.py:3092`, `3158`, `5590`; bridged `config.py:1547-1548` |
| `dm_top_level_threads_as_sessions` | — | `true` | `adapter.py:3045` |
| `assistant_thread_titles` | — | `true` | `adapter.py:4563` |
| `reply_broadcast` | — | `false` | `adapter.py:2543` |
| `reactions` | `SLACK_REACTIONS` | `"true"` | `adapter.py:3755`; bridge `:9018-9019` |
| `reaction_triggers` | `SLACK_REACTION_TRIGGERS` | unset | `adapter.py:4977` (`:4961`); bridge `:9020-9024` |
| `reaction_trigger_target` | `SLACK_REACTION_TRIGGER_TARGET` | `""` | `adapter.py:5002` (`:4992`); bridge `:9025-9027` |
| `channel_skill_bindings` | — | — | bridged `config.py:1585-1586` |
| — | `SLACK_DEDUP_TTL_SECONDS` | `""` | `adapter.py:759` |
| — | `SLACK_BOT_TOKEN` / `SLACK_APP_TOKEN` | required | `adapter.py:1784`, `:9045-9055`, `:9072` |
| — | `SLACK_HOME_CHANNEL` | unset | `adapter.py:9089` (cron delivery) |

### What is actually live on Tars (read-only inspection)

`~/.hermes/config.yaml`, `gateway.platforms.slack` — only four keys (lines 120-124):

```
120:    slack
121:      enabled
122:      require_mention: true
123:      strict_mention: true
124:      unauthorized_dm_behavior
```

`~/.hermes/.env` — KEY NAMES only (no values read or printed):
`SLACK_ALLOWED_USERS`, `SLACK_APP_TOKEN`, `SLACK_BOT_TOKEN`, `SLACK_HOME_CHANNEL`,
`SLACK_MCP_XOXC_TOKEN`, `SLACK_MCP_XOXD_TOKEN` (+ non-Slack keys).

⇒ No `SLACK_STRICT_MENTION` / `SLACK_THREAD_REQUIRE_MENTION` env override exists,
so the YAML bridge is what sets them: `SLACK_STRICT_MENTION=true`,
`SLACK_THREAD_REQUIRE_MENTION` unset (→ `false`), `free_response_channels` unset,
`allowed_channels` unset, `allow_bots` = `"none"`.

**Consequence with the live config**: in the §3 ladder, branch 18 is False
(`require_mention=true`, no free-response channels), so **branch 19 fires** —
`strict_mention=true and not is_mentioned → return`. Branch 21
(`_should_wake_on_unmentioned_message`, i.e. every thread-awareness check) is
**dead code on Tars today**. That is exactly why unmentioned thread replies are
dropped as silently as unmentioned top-level messages.

Also note `adapter.py:5763-5768`: under `strict_mention=true` the bot does **not**
even register a mentioned thread, so there is no state to follow later.

---

## 5. VERDICT

### ✅ YES — there is a safe, config-only knob: turn `strict_mention` OFF.

There is no positive "follow threads" switch; the mechanism already exists and is
being suppressed by `strict_mention: true`. Removing that suppression re-enables
branch 21 while leaving `require_mention: true` in force.

Resulting behaviour with `require_mention: true`, `strict_mention: false`,
`thread_require_mention` absent (false):

| Inbound | Path | Outcome |
|---|---|---|
| Unmentioned **top-level** channel msg | 18 F → 19 F → 20 F → 21: `_should_wake_on_unmentioned_message(event_thread_ts=None)` → `adapter.py:5187-5188` `return False` | **still dropped** ✔ (unchanged) |
| Unmentioned **thread reply** in a thread Tars posted / was mentioned in / has a session for | 21 → wake check 1/2/3/4/5 → `True` | **answered** ✔ (the goal) |
| Unmentioned thread reply in a thread Tars never touched | 21 → all five checks False | dropped ✔ |
| Any DM | branch 14 (`if not is_one_to_one_dm`) skips the whole ladder | unchanged ✔ |
| Non-Gaetan sender, anywhere | killed at `adapter.py:5544` — **104 lines earlier** | rejected ✔ |

Blast radius is confined to `is_one_to_one_dm == False` channel/MPIM traffic from an
already-allowlisted user. `allowed_channels` is unset, so Tars would follow its own
threads in any channel it is a member of — tighten with `allowed_channels` if that
matters. The bot's own echo is still blocked (`adapter.py:5353-5354`) and other bots
still blocked by `allow_bots: none` (`adapter.py:5339-5340`, `5649-5650`).

#### Recommended stanza (option A — most robust, bypasses the env bridge)

`~/.hermes/config.yaml`:

```yaml
gateway:
  platforms:
    slack:
      enabled: true
      require_mention: true
      unauthorized_dm_behavior: <unchanged>
      extra:
        strict_mention: false
```

Why `extra:` — `PlatformConfig.from_dict` reads the `extra:` sub-key verbatim
(`config.py:671`, `:706`), and `_slack_strict_mention()` checks
`self.config.extra.get("strict_mention")` **first** (`adapter.py:8268-8272`), before
consulting `SLACK_STRICT_MENTION`. Note `strict_mention` is **not** in the
shared-key bridge list at `config.py:1535-1598` (only `require_mention`,
`free_response_channels`, `mention_patterns`, `reply_in_thread`, … are), so the
`extra:` route is the only way to put it into `PlatformConfig.extra`.

#### Option B — simply flip the value

```yaml
      strict_mention: false
```

⚠️ **Restart caveat, verify before trusting a live reload.** `_apply_yaml_config`
writes the env var only when it is not already set:

```python
    if "strict_mention" in slack_cfg and not os.getenv("SLACK_STRICT_MENTION"):
        os.environ["SLACK_STRICT_MENTION"] = str(slack_cfg["strict_mention"]).lower()
```
`adapter.py:8994-8995`

The running gateway process already has `SLACK_STRICT_MENTION="true"` stamped into
`os.environ` from the first load. A live config reload in the **same process** will
hit the `not os.getenv(...)` guard and skip the write, and nothing unsets the
variable — so option B is likely a **no-op until the gateway is restarted**.
Option A does not have this problem (it does not go through env at all), but still
depends on the reload actually rebuilding `PlatformConfig` and handing the adapter
the new `config.extra` — **not verified in this probe**. Safest sequence regardless
of option: edit under `flock`, with a `.bak` first, then restart the gateway unit
and confirm with an @-mention followed by an unmentioned thread reply.

#### If more surgical scoping is wanted

`free_response_channels: <channel-id>` + `thread_require_mention: false` is *not* the
answer — that would make top-level messages in that channel free-response too
(branch 18), which is the behaviour you explicitly do not want.

### Code change — NOT needed, and NOT applied

For the record, if you ever wanted "unmentioned thread-follow while keeping
`strict_mention: true` for everything else", the minimal diff is one line in
`plugins/platforms/slack/adapter.py::_handle_slack_message` at line 5707:

```python
-            elif self._slack_strict_mention() and not is_mentioned:
+            elif self._slack_strict_mention() and not is_mentioned and not is_thread_reply:
                 return  # Strict mode: ignore until @-mentioned again
```

…plus dropping `and not self._slack_strict_mention()` from the
`_register_mentioned_thread` guard at `adapter.py:5763-5768` so the thread memory is
actually populated. **Not applied** — and unnecessary, since the config-only route
above achieves the same end state.

---

## Evidence / method notes

- All file reads were `ssh`/`scp` copies; nothing on the VM was written, and no
  service was touched.
- No `sops -d` was run. Secret KEY NAMES only were listed from `~/.hermes/.env` via
  `grep -oE '^[A-Z_]+='`; no value was read, printed, or copied.
- `config.yaml` was inspected key-name-only (`grep -nE '^[[:space:]]*[a-zA-Z_]+:' | cut -d: -f1,2`),
  plus a targeted grep for the gating keys whose values are booleans.
- Not verified in this probe (would need a live test): whether a Hermes config
  live-reload rebuilds `PlatformConfig.extra` for an already-connected Slack adapter.
