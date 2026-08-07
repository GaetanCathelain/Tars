# WF5 — adversarial verification of `strict_mention: false`

**Role:** refute the claim. Default verdict on uncertainty = refuted.
**Mode:** READ-ONLY. Nothing on the VM was edited, no service restarted, no Slack
write call issued. Slack API touched read-only twice (`users.conversations`,
`conversations.replies`) via `curl -K -` with the token piped on stdin, never argv.
**Date:** 2026-08-07, VM clock UTC. Hermes v0.20.0, source checkout
`/home/gaetan/.hermes/hermes-agent/`. All `adapter.py` line numbers below are
`plugins/platforms/slack/adapter.py` unless stated otherwise.

## The claim under test

> Setting `gateway.platforms.slack.extra.strict_mention: false` (keeping
> `require_mention: true`) makes Tars answer Gaetan's unmentioned replies in
> threads it already participates in, and changes NOTHING else that matters:
> unmentioned TOP-LEVEL channel messages are still dropped, the allowlist is
> untouched, and no non-Gaetan user can trigger Tars.

---

## 0. Live baseline (re-confirmed, not assumed)

`~/.hermes/config.yaml`, redacted dump of the only platform-gating stanza that exists:

```json
{"gateway": {"platforms": {
  "slack": {"enabled": true, "require_mention": true,
            "strict_mention": true, "unauthorized_dm_behavior": "ignore"},
  "a2a":   {"enabled": true, "extra": {"port": 9900}}}}}
```

There is **no** top-level `slack:` block, **no** `extra:` under `slack`, **no**
`allowed_channels`, **no** `mention_patterns`, **no** `free_response_channels`,
**no** `require_mention_channels`, **no** `channel_overrides`, **no**
`reaction_triggers`.

`.env` key names only (`grep -oE '^[A-Za-z0-9_]+=' ~/.hermes/.env`):

```
GMAIL_ADDRESS GMAIL_APP_PASSWORD HINDSIGHT_MODE LINEAR_API_KEY NOTION_API_TOKEN
SLACK_ALLOWED_USERS SLACK_APP_TOKEN SLACK_BOT_TOKEN SLACK_HOME_CHANNEL
SLACK_MCP_XOXC_TOKEN SLACK_MCP_XOXD_TOKEN
```

→ **`SLACK_STRICT_MENTION` is NOT in `.env`.** Neither is `SLACK_MENTION_PATTERNS`,
`SLACK_ALLOWED_CHANNELS`, `SLACK_FREE_RESPONSE_CHANNELS`,
`SLACK_THREAD_REQUIRE_MENTION`. This matters twice below (§2, §5).

Gateway: systemd `--user` unit `hermes-gateway.service`,
`ExecStart=… -m hermes_cli.main gateway run`, `ExecReload=/bin/kill -USR1 $MAINPID`,
`MainPID=76255`, `ActiveEnterTimestamp=Fri 2026-08-07 19:40:45 UTC`.
`~/.hermes/gateway-starts.log` last entry `1786131646.677635` (= 19:40:46 UTC).
The gateway has **not** restarted since, even though `config.yaml` was modified at
20:21 (kanban). Direct evidence that a `config.yaml` write does **not** bounce the
process.

---

## 1. Top-level containment — every return-True path of `_should_wake_on_unmentioned_message`

`adapter.py:5164-5237`, verbatim body:

```python
5164:    async def _should_wake_on_unmentioned_message(
5165:        self,
5166:        event_thread_ts,
5167:        channel_id: str,
5168:        user_id: str,
5169:        is_thread_reply: bool,
5170:        team_id: str = "",
5171:        chat_type: str = "group",
5172:    ) -> bool:
...
5187:        if not event_thread_ts:
5188:            return False
5189:        thread_marker = self._workspace_message_marker(team_id, event_thread_ts)
5190:        # Check both the workspace-scoped marker and the bare ts: entries
5191:        # recorded before a team id was learned (or by legacy paths) are bare
5192:        # strings, and a scoped-vs-bare mismatch must not silence the bot.
5193:        if is_thread_reply and (
5194:            thread_marker in self._bot_message_ts
5195:            or event_thread_ts in self._bot_message_ts
5196:        ):
5197:            return True
5198:        if (
5199:            thread_marker in self._mentioned_threads
5200:            or event_thread_ts in self._mentioned_threads
5201:        ):
5202:            return True
5203:        if is_thread_reply and self._has_active_session_for_thread(
5204:            channel_id=channel_id,
5205:            thread_ts=event_thread_ts,
5206:            user_id=user_id,
5207:            team_id=team_id,
5208:            chat_type=chat_type,
5209:        ):
5210:            return True
5211:        # 4th check: bot-initiated thread via direct chat.postMessage.
5212:        if is_thread_reply and await self._bot_authored_thread_root(
5213:            channel_id=channel_id,
5214:            thread_ts=event_thread_ts,
5215:            team_id=team_id,
5216:        ):
5217:            return True
...
5223:        if is_thread_reply:
5224:            bot_uid = self._team_bot_user_ids.get(team_id, self._bot_user_id)
5225:            if bot_uid:
5226:                parent_text = await self._fetch_thread_parent_text(
5227:                    channel_id=channel_id,
5228:                    thread_ts=event_thread_ts,
5229:                    team_id=team_id,
5230:                    strip_bot_mention=False,
5231:                )
5232:                if parent_text and f"<@{bot_uid}>" in parent_text:
5233:                    # Remember the thread so later replies skip the fetch.
5234:                    if not self._slack_strict_mention():
5235:                        self._register_mentioned_thread(event_thread_ts)
5236:                    return True
5237:        return False
```

### Enumeration of return-True paths

| # | Line | Gated on `is_thread_reply`? | Source of truth |
|---|------|------------------------------|-----------------|
| 1 | 5197 | **YES** (5193) | `self._bot_message_ts` — RAM set |
| 2 | 5202 | **NO** ← the only ungated path | `self._mentioned_threads` — RAM set |
| 3 | 5210 | **YES** (5203) | `sessions.json` via `_has_active_session_for_thread` |
| 4 | 5217 | **YES** (5212) | Slack API `conversations.replies` parent author |
| 5 | 5236 | **YES** (5223) | Slack API parent text contains `<@BOTUID>` |

**Path 2 is not gated on `is_thread_reply`. This is the one genuine code-level gap
in the claim's "top-level messages are still dropped" wording.** Whether it is
reachable for a top-level message depends entirely on what `event_thread_ts` is.

### What is actually passed

`adapter.py:5622-5623`:

```python
5622:        event_thread_ts = event.get("thread_ts")
5623:        is_thread_reply = bool(event_thread_ts and event_thread_ts != ts)
```

and at the call site, `adapter.py:5722-5729`:

```python
5722:                if not await self._should_wake_on_unmentioned_message(
5723:                    event_thread_ts=event_thread_ts,
```

It is the **RAW `event["thread_ts"]`**, *not* the session-scoped `thread_ts` local
(which falls back to `ts` for top-level channel messages at `adapter.py:5591-5593`).
That distinction is what saves the claim. For an ordinary top-level channel post
Slack sends no `thread_ts` → `event_thread_ts is None` → `5187: return False`.

### The four edge cases you asked me to consider

**(a) `thread_ts == ts` (thread-parent shape).** `is_thread_reply` is False (5623),
`event_thread_ts` is truthy, so execution reaches the ungated path 2 at 5198. To
return True, that `ts` must already be in `_mentioned_threads`. Registration is
`adapter.py:5763-5768`:

```python
5763:            if (
5764:                thread_ts
5765:                and not self._slack_strict_mention()
5766:                and not self._slack_thread_require_mention()
5767:            ):
5768:                self._register_mentioned_thread(thread_ts, team_id=team_id)
```

— reached only inside `if is_mentioned:` (5732) and keyed on that message's own ts.
So a *new* top-level message can never collide: its ts is fresh and cannot already
be in the set. **The only reachable instance is a re-delivery of a message that
already mentioned the bot** — and such a re-delivery has `is_mentioned=True`, so it
never reaches 5721 at all. One residual: see (a′).

**(a′) Message edits — a real new surface, not a top-level breach.**
`adapter.py:5291-5297`:

```python
5291:            normalized_event = dict(updated_message)
...
5297:            event = normalized_event
```

A `message_changed` event is re-normalised into the *stored message object*, which
for a thread parent carries `thread_ts == ts`, and for a thread reply carries the
real root `thread_ts`. The dedup guard at 5270-5277 only suppresses edits of
messages Tars **already processed** (`_processed_message_ts` is written at 6347,
only on dispatch). Therefore: **after the change, editing a previously-dropped
message inside a followed thread re-enters the gate and wakes Tars.** Today
`strict_mention: true` kills it at 5707. This is new noise, not a containment
breach — the edited message must still be Gaetan's and in a followed thread.

**(b) `reply_broadcast` / `subtype: thread_broadcast`.** Slack sets
`thread_ts != ts` on these, so `is_thread_reply=True` (5623) and they are handled as
ordinary thread replies. No new path.

**(c) "a channel where a session happens to exist".** Path 3 is gated *both* on
`is_thread_reply` (5203) **and** on a session key that embeds the thread ts
(`_build_thread_session_key`, 7802-7860 → `gateway/session.py:build_session_key`).
There is no channel-wide session lookup. A top-level message cannot reach it.

**(d) `thread_ts` absent / None / empty string.** All three are falsy → `5187-5188`
returns False before anything else runs. Empty string included.

### §1 verdict

**Top-level containment HOLDS.** No reachable path returns True for an unmentioned
top-level, non-thread channel message. Path 2's missing `is_thread_reply` guard is a
latent defect in the upstream code, but with `_mentioned_threads` keyed by
mention-bearing message ts it is unreachable for fresh top-level traffic. Reportable
caveats: (a′) edits, which are new.

---

## 2. Scope creep — does `strict_mention: false` switch on plain-text name matching?

`is_mentioned` is computed at `adapter.py:5615-5621`:

```python
5615:        bot_uid = self._team_bot_user_ids.get(team_id, self._bot_user_id)
5616:        # Detect mentions authored only inside Block Kit blocks too (#52387)
5617:        routing_text = _slack_mention_detection_text(event) or original_text or ""
5618:        is_mentioned = bool(
5619:            (bot_uid and f"<@{bot_uid}>" in routing_text)
5620:            or self._slack_message_matches_mention_patterns(routing_text)
5621:        )
```

The only fuzzy/plain-text arm is `_slack_message_matches_mention_patterns`
(`adapter.py:8471-8475`), backed by `_slack_mention_patterns`
(`adapter.py:8423-8469`):

```python
8438:        patterns = self.config.extra.get("mention_patterns") if self.config.extra else None
8439:        if patterns is None:
8440:            raw = os.getenv("SLACK_MENTION_PATTERNS", "").strip()
...
8471:    def _slack_message_matches_mention_patterns(self, text: str) -> bool:
8472:        """Return True when ``text`` matches a configured wake-word pattern."""
8473:        if not text:
8474:            return False
8475:        return any(pattern.search(text) for pattern in self._slack_mention_patterns())
```

Both sources are empty on this deployment (§0: no `mention_patterns` in YAML, no
`SLACK_MENTION_PATTERNS` in `.env`) → `compiled == []` → `any([])` is False, always.

`_slack_mention_patterns` has **no** implicit default: it never derives a pattern
from the bot's display name, `real_name`, `profile.display_name`, or the literal
string "tars". Grep across `adapter.py` for a name-derived wake heuristic returns
nothing outside this explicitly-configured path.

`strict_mention` appears in exactly four places in `adapter.py` — 5234, 5707, 5765,
and its own getter at 8263 — none of which touch `is_mentioned`, `routing_text`,
`mention_patterns`, or `_compiled_mention_patterns`.

### §2 verdict

**No plain-text name matching. Not enabled, not enable-able as a side effect.**
`strict_mention: false` affects only the *unmentioned*-message branch. NOT a
refutation.

---

## 3. The elif ladder, verbatim `5654-5730`

```python
5654:        if not is_one_to_one_dm and bot_uid:
5655:            # Check allowed channels — if set, only respond in these channels (whitelist)
5656:            allowed_channels = self._slack_allowed_channels()
5657:            if allowed_channels and channel_id not in allowed_channels:
5658:                logger.debug(
5659:                    "[Slack] Ignoring message in non-allowed channel: %s", channel_id
5660:                )
5661:                return
5662:
5663:            # A message that opens by @mentioning another user is directed at
5664:            # that person. Stay silent unless we are also mentioned — this
5665:            # overrides free-response and mentioned-thread auto-follow so the
5666:            # bot does not butt in on chatter aimed at someone else.
5667:            self_uids = {u for u in (bot_uid, self._bot_user_id) if u}
5668:            if (
5669:                self._slack_ignore_other_user_mentions()
5670:                and not is_mentioned
5671:                and not self._slack_message_mentions_self(routing_text, self_uids)
5672:                and self._slack_message_addressed_to_other_user(routing_text, self_uids)
5673:            ):
5674:                logger.debug(
5675:                    "[Slack] Ignoring message addressed to another user in channel %s",
5676:                    channel_id,
5677:                )
5678:                return
5679:
5680:            if force_process:
5681:                pass  # Explicit internal routing path (reaction trigger).
5682:            elif (
5683:                channel_id not in self._slack_require_mention_channels()
5684:                and (
5685:                    channel_id in self._slack_free_response_channels()
5686:                    or not self._slack_require_mention()
5687:                )
5688:            ):
...
5695:                if (
5696:                    self._slack_thread_require_mention()
5697:                    and is_thread_reply
5698:                    and not is_mentioned
5699:                ):
...
5706:                    return
5707:            elif self._slack_strict_mention() and not is_mentioned:
5708:                return  # Strict mode: ignore until @-mentioned again
5709:            elif (
5710:                self._slack_thread_require_mention()
5711:                and is_thread_reply
5712:                and not is_mentioned
5713:            ):
...
5720:                return
5721:            elif not is_mentioned:
5722:                if not await self._should_wake_on_unmentioned_message(
5723:                    event_thread_ts=event_thread_ts,
5724:                    channel_id=channel_id,
5725:                    user_id=user_id,
5726:                    team_id=team_id,
5727:                    is_thread_reply=is_thread_reply,
5728:                    chat_type="dm" if is_dm else "group",
5729:                ):
5730:                    return
```

**Ordering trace with the proposed config** (`require_mention: true`,
`strict_mention: false`, everything else unset, channel `C08RWSTU9LK`, Gaetan,
unmentioned thread reply):

- 5657 — `allowed_channels` empty → no return.
- 5668 — `ignore_other_user_mentions` defaults False → no return.
- 5680 — `force_process` False (no `reaction_triggers` configured).
- 5682 — `require_mention_channels` empty ✓, but `free_response_channels` empty and
  `require_mention()` is **True** → the `and (…)` arm is False → branch **not** taken.
- 5707 — `strict_mention()` now **False** → branch **not** taken. *(Today this is
  where the message dies.)*
- 5709 — `thread_require_mention()` defaults False → branch not taken.
- 5721 — `not is_mentioned` is True → **enters**, calls
  `_should_wake_on_unmentioned_message`.
- 5730 — a **False** result still `return`s, i.e. drops the message with no log
  output. Confirmed: the drop is a bare `return` inside the `elif`, no logger call.

### §3 verdict

**Confirmed exactly as claimed.** The flow reaches the wake function only because
5707 stops short-circuiting, and a False verdict still drops. NOT a refutation.

---

## 4. Which wake path fires for r7 / r10 — and is the memory persisted?

### Is `_mentioned_threads` RAM or disk?

`adapter.py:971-975`:

```python
971:        self._bot_message_ts: set[str] = set()
...
975:        self._mentioned_threads: set[str] = set()
```

Bounded-set eviction only (`_trim_bot_message_timestamps` 1106-1109,
`_trim_mentioned_threads` 1111-1115). `_register_mentioned_thread`
(`adapter.py:5096-5109`) writes to the set and nothing else. A full-file grep for
`_mentioned_threads` returns exactly lines 975, 1111-1115, 5106, 5177, 5199-5200 —
**no serialiser, no path, no write.**

**Both `_mentioned_threads` and `_bot_message_ts` are process-RAM only. Neither
survives a restart.** And per the claim, they are indeed never populated for new
mentions while `strict_mention: true` (guard at 5765).

### The reference thread `1786135149.072829` in `C08RWSTU9LK`

Read-only `conversations.replies` (bot token on stdin, `curl -K -`):

```
1786135149.072829  U08BDJAMSRZ  mentions_tars=True   "<@U0BBH85NAKH> tu peux suivre les instuctions d'…"   ← ROOT
1786135155.707859  U0BBH85NAKH  bot=True
1786135178.649919  U0BBH85NAKH  bot=True
1786135219.118499  U08BDJAMSRZ  mentions_tars=True
1786135225.606029  U0BBH85NAKH  bot=True
1786135242.240479  U08BDJAMSRZ  mentions_tars=True
1786135259.991959  U0BBH85NAKH  bot=True
1786135270.459679  U08BDJAMSRZ  mentions_tars=False  "Ptin t'es nul"                 ← r7  (dropped today)
1786135283.424339  U08BDJAMSRZ  mentions_tars=True
1786135289.769869  U0BBH85NAKH  bot=True
1786135310.012819  U08BDJAMSRZ  mentions_tars=False  "Attends que je te tune toi"     ← r10 (dropped today)
1786135439.766629  U05LKHLDV0A  mentions_tars=False  "#RacingClubDeLens"             ← Nans, allowlist-rejected
```

Session store — `~/.hermes/sessions/sessions.json` contains:

```
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786135149.072829
```

### Path-by-path verdict for r7 / r10 after the change

| Path | Fires? | Why |
|------|--------|-----|
| 1 `_bot_message_ts` (5193-5197) | **Yes today, NO after restart** | Populated at 2596-2599 (`# Also register the thread root so replies-to-my-replies work`) when the bot replied in-thread at 20:39. RAM — wiped by the restart the change requires. |
| 2 `_mentioned_threads` (5198-5202) | **No** | Never populated: guard 5765 has been `strict_mention: true` for the whole life of this thread, and it is RAM anyway. |
| 3 `_has_active_session_for_thread` (5203-5210) | **YES — survives restart** | Key present in `sessions.json`. `_should_reset` (`gateway/session.py:2188-2233`) returns `None` immediately for `policy.mode == "none"` (line 2210-2211) → `_has_active_session_for_thread` returns True (7972-8032, line 8030). |
| 4 `_bot_authored_thread_root` (5211-5217) | No | Root author is `U08BDJAMSRZ`, not the bot. |
| 5 parent-text mention (5218-5236) | **YES — survives restart** | Root text is `<@U0BBH85NAKH> tu peux suivre…`; fetched live from the Slack API, no process state. |

Two independent restart-surviving paths (3 and 5) hit. Note path 3 requires the
key to match: `_build_thread_session_key` (7802-7860) delegates to
`gateway/session.py:build_session_key`, and with `thread_sessions_per_user=False`
(the default; 7848-7852) the user id is **not** appended for threads — which is
exactly the shape observed in `sessions.json`.

### §4 verdict

**The reference thread works IMMEDIATELY after the change — no fresh mention
needed.** The verification plan can replay r7/r10-style messages directly in
`…:1786135149.072829`. **Correction to the brief's premise:** it is not path 2 that
carries it (that memory is RAM-only and empty), it is paths 3 + 5. Because paths 3
and 5 are disk/API-backed, the restart in §5 does not invalidate this.

---

## 5. Effect delivery — is a restart required?

### (a) Is `_slack_strict_mention()` cached?

`adapter.py:8263-8278`:

```python
8263:    def _slack_strict_mention(self) -> bool:
...
8268:        configured = self.config.extra.get("strict_mention")
8269:        if configured is not None:
8270:            if isinstance(configured, str):
8271:                return configured.lower() in {"true", "1", "yes", "on"}
8272:            return bool(configured)
8273:        return os.getenv("SLACK_STRICT_MENTION", "false").lower() in {
8274:            "true", "1", "yes", "on",
8275:        }
```

No memo, no `functools.cache`, no instance attribute. It is **evaluated per call**,
i.e. per message at 5707 / 5765 / 5234. Contrast `_slack_mention_patterns`
(8434-8436) which *is* memoised on `self._compiled_mention_patterns`.

So the value is read fresh — but from two objects that are themselves startup-frozen.

### (b) Does live-reload rebuild `PlatformConfig.extra`?

`self.config` on the adapter is assigned once, in the base adapter constructor:

- `gateway/platforms/base.py:2760` — `self.config = config`
- `plugins/platforms/slack/adapter.py:9058-9060` — `def _build_adapter(config): return SlackAdapter(config)`, registered as `adapter_factory=_build_adapter` (9068)
- `gateway/run.py:5897-5903` — `def __init__(self, config=None): … self.config = config if config is not None else load_gateway_config_for_runner()`

I searched for a rebuild path and found none:

- No mtime-watcher on `config.yaml` for platform config anywhere in `gateway/`. The
  mtime-cached reads that *do* exist (`gateway/run.py:3169 _load_gateway_config`,
  used at 320, 3154, 3271, 3768, 7573 …) return a **plain dict** for specific
  features (compression notices, max_turns bridge, kanban); none of them
  reconstructs `PlatformConfig` or reassigns `adapter.config`.
- `gateway/run.py:1838-1864 _reload_runtime_env_preserving_config_authority()`
  re-reads `.env` per turn and re-bridges *only* `HERMES_MAX_ITERATIONS`,
  `HERMES_CJK_FTS`, `HERMES_SEARCH_SLOW_MS` (1892-1902). It does **not** re-run
  `_apply_yaml_config`, so `SLACK_STRICT_MENTION` is never re-stamped mid-life.
- `apply_yaml_config_fn` is dispatched exactly once, inside `load_gateway_config()`
  (`gateway/config.py:1648-1678`) — a startup-only call.
- The unit's `ExecReload=/bin/kill -USR1 $MAINPID` maps to
  `gateway/run.py:27117-27119` `loop.add_signal_handler(signal.SIGUSR1, restart_signal_handler)`
  — a **full in-band process restart**, not an in-place config reload.

**Empirical corroboration:** `config.yaml` mtime is `Aug 7 20:21`; gateway
`ActiveEnterTimestamp` is `19:40:45` and `MainPID=76255` is unchanged. A config
write demonstrably did not bounce or re-read the platform config.

### (c) Would the *env* route work instead?

No — and it is a trap. `_apply_yaml_config` (`adapter.py:8976-8995`):

```python
8994:    if "strict_mention" in slack_cfg and not os.getenv("SLACK_STRICT_MENTION"):
8995:        os.environ["SLACK_STRICT_MENTION"] = str(slack_cfg["strict_mention"]).lower()
```

At the 19:40 start it stamped `SLACK_STRICT_MENTION=true` into the process env
(from the current `strict_mention: true`). The `not os.getenv(...)` guard means that
**even after a restart**, flipping the plain `gateway.platforms.slack.strict_mention`
key to `false` would… actually work, because a restart is a fresh process with a
clean env (`SLACK_STRICT_MENTION` is absent from `.env`, §0). But **within the
running process it can never be un-stamped**, so nothing takes effect without the
restart either way.

### (d) Does `extra.strict_mention` even reach `config.extra`?

Yes. `strict_mention` is **not** in the shared-key bridge list
(`gateway/config.py:1535-1598` — `require_mention` is at 1551-1552,
`mention_patterns` at 1563-1564, `free_response_channels` at 1561-1562; there is no
`strict_mention` entry). Its only routes are the env stamp above and an **explicit
`extra:` sub-key**, which `_merge_platform_map` deep-merges:

```python
gateway/config.py:1441:  merged_extra = {**existing.get("extra", {}), **plat_block.get("extra", {})}
gateway/config.py:1446:                  merged["extra"] = merged_extra
gateway/config.py:671:   extra = _coerce_dict(data.get("extra", {}))
gateway/config.py:706:            extra=extra,
```

and `gateway.platforms.*` is a merge source (`gateway/config.py:1449`). The live
`a2a` stanza already proves this route works — its `extra.port: 9900` is exactly
this shape. `_slack_strict_mention()` then hits 8268 first and returns
`bool(False)` = `False`, beating the stale env value. **`extra:` is the correct
placement and it out-ranks the env var** — which is why it is preferable to editing
the plain key.

### §5 verdict — **A GATEWAY RESTART IS REQUIRED.**

Not a live-reload change. `systemctl --user restart hermes-gateway.service` (or
`reload`, which is SIGUSR1 → in-band restart). This is heavier than the claim
implies and Gaetan must be told. Side effects of the restart itself:
`_mentioned_threads`, `_bot_message_ts`, `_thread_context_cache`, `_processed_message_ts`
and `_dedup` are all wiped; in-flight turns drain (`TimeoutStopSec=60`).
`gateway_restart_notification` is not disabled in config, so the restart may post a
"♻ Gateway restarted" line on Slack — check before doing it in a shared channel.

---

## 6. Blast radius

`allowed_channels` is unset (`_slack_allowed_channels`, 8387-8402, returns an empty
set from both YAML and env) → **no channel whitelist gate at 5657.**

Surfaces, after the change:

| Surface | `is_one_to_one_dm`? | Behaviour change |
|---|---|---|
| 1:1 DMs (`channel_type == "im"`) | Yes (5528) | **NONE.** The whole gating block is `if not is_one_to_one_dm and bot_uid:` (5654) — DMs skip it entirely, mentioned or not. Unchanged. |
| Group DMs (MPIM) | No (`is_dm` True but `is_one_to_one_dm` False) | **Would change** — MPIMs run the same ladder as channels. Tars is in **zero** MPIMs today (see below). |
| Public/private channels Tars is a member of | No | **Changes**, but only for allowlisted senders. |

### Channels Tars is actually in

Read-only `users.conversations` with the bot token (types=public_channel,
private_channel, mpim, im; `exclude_archived=true`), `ok=true`:

```
C08RWSTU9LK   gcn-sandbox    public
D0BL5JW56DP … D0BBH82RBGF    im  × 27
```

**Exactly one non-DM conversation: `C08RWSTU9LK` (#gcn-sandbox), public, containing
real colleagues (Nans `U05LKHLDV0A` posted in the reference thread at 20:43:59;
Olivier is also in it).** No private channels. No MPIMs.

### Can a non-Gaetan user trigger Tars?

No. The allowlist early-reject is upstream of everything in §3
(`adapter.py:5534-5550`):

```python
5536:        if user_id and callable(_auth_fn):
...
5544:            if not _auth_fn(_source):
5545:                logger.warning(
5546:                    "[Slack] Early reject of unauthorized user %s in channel %s",
5547:                    user_id,
5548:                    channel_id,
5549:                )
5550:                return
```

Line 5550 returns at file offset **5550 < 5654**, i.e. before the ladder, before
`is_mentioned` is even computed (5618). `strict_mention` is not read on this path.
Live-proven: Nans rejected with this WARNING at 20:44:01 UTC on 2026-08-07 in
`C08RWSTU9LK`. `SLACK_ALLOWED_USERS=U08BDJAMSRZ` is untouched by the change.

**§6 verdict: blast radius = Gaetan's unmentioned messages, in followed threads,
in one public channel (#gcn-sandbox) + the DM surface (already unrestricted).
The allowlist half of the claim is CONFIRMED.**

---

## 7. Failure modes — what breaks, what gets noisy

**7.1 Thinking-out-loud gets answered. This is the real cost, and it is already
evidenced.** The two messages the verification is meant to unblock are r7
`"Ptin t'es nul"` and r10 `"Attends que je te tune toi"`. Neither is a question;
both are Gaetan venting at the bot. After the change, **both wake Tars.** Any
`ok`, `merci`, `mdr`, `attends`, `non laisse tomber` typed in a followed thread
becomes a turn — with tool access, model cost, and a visible reply in a channel
where Nans and Olivier are watching.

**7.2 Old threads reactivate.** Path 3 keys off `sessions.json`, which is
persistent, and `session_reset: {mode: none}` means `_should_reset` returns `None`
unconditionally (`gateway/session.py:2210-2211`). There is **no TTL and no idle
expiry** — `policy.mode in {"idle","both"}` (2215) and `{"daily","both"}` (2220) are
both unreachable. Combined with path 5 (parent text, fetched live from the Slack
API, ageless), **every thread Tars has ever been mentioned in stays permanently
armed.** Today that is 4 sessions in `C08RWSTU9LK`
(`…1786129267.292069`, `…1786131725.413279`, `…1786134155.187189`,
`…1786135149.072829`). A comment from Gaetan in any of them, months from now,
wakes Tars.

**7.3 Edits become triggers.** Per §1(a′): editing a message Tars previously
dropped re-enters the gate (no `_processed_message_ts` entry, 6347) and wakes it.
New behaviour, easy to hit while cleaning up a typo.

**7.4 The empty-response / 61 s stall.** Documented for this deployment in
`docs/specs/wf4-probes.md` §19: Hermes has **no silent-drop terminal at the MODEL
level**, so a model that decides not to answer produces N empty retries and a
user-facing `empty_response_exhausted` error. The SOUL rule that avoids it is rule 4
(`~/.hermes/SOUL.md`, verbatim):

> 4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — a channel
>    message whose sender prefix reads "[U08BDJAMSRZ | …]" is from Gaetan. To anyone
>    else, in any channel or DM, I give no answer: I reply with the single character
>    "·" and nothing else — no content, no reaction, no explanation.

**The `·` rule does NOT cover the new wake cases.** It is scoped to *"anyone else"*
— a non-Gaetan sender. Every message the change newly admits is **from Gaetan**, so
rule 4's `·` branch never applies to them. There is nothing in SOUL.md telling Tars
what to do when Gaetan says something that isn't addressed to it. The model must
choose between (i) answering venting (7.1 noise) or (ii) deciding not to answer —
and (ii) is precisely the state that has no terminal, i.e. the 61 s dead end.
**This is the single biggest risk in the change and it is not mitigated today.**
Mitigation would be a SOUL amendment giving Gaetan-addressed-to-nobody the same `·`
terminal, applied *before* or *with* the config change.

---

## 8. Is there a narrower knob?

I enumerated every Slack config getter in the adapter
(`grep -n "    def _slack_[a-z_]*(self)"`):

| Getter | Line | Useful here? |
|---|---|---|
| `_slack_ignored_channels` | 2425 | Blunt: silences a channel entirely. |
| `_slack_allow_bots` | 3106 | Bot senders only. Irrelevant. |
| `_slack_reaction_triggers` / `_target` | 4961 / 4992 | **Alternative design, not a narrowing of this one** — an emoji on a message routes it to Tars explicitly. Opt-in per message, zero ambient wake. |
| `_slack_require_mention` | 8244 | Already true; this is what we keep. |
| `_slack_strict_mention` | 8263 | The knob under test. |
| `_slack_ignore_other_user_mentions` | 8280 | Only affects messages *leading* with another user's `@`. Doesn't scope thread-follow. |
| `_slack_thread_require_mention` | 8303 | **Verified: tightens, does not loosen.** See below. |
| `_slack_free_response_channels` | 8355 | Loosens far more (top-level too). Wrong direction. |
| `_slack_disable_dms` | 8373 | Irrelevant. |
| `_slack_allowed_channels` | 8387 | **The one genuinely useful companion** — see below. |
| `_slack_require_mention_channels` | 8404 | **Does not help.** Read only inside the free-response branch at 5683; with `require_mention: true` that branch is never taken, so this set is never consulted on our path. |
| `_slack_mention_patterns` | 8423 | Loosens. Wrong direction. |

**`thread_require_mention` — confirmed it tightens.** Docstring at 8304-8311:
*"When true, Slack thread replies require an explicit @-mention. This is narrower
than `strict_mention`: top-level channel messages can still be processed without a
mention when `require_mention` is false or the channel is listed in
`free_response_channels`."* In the ladder it appears only as `… and is_thread_reply
and not is_mentioned: return` (5695-5706 and 5709-5720). It is the **opposite** of
what is wanted and would cancel the change entirely (it also blocks
`_register_mentioned_thread` at 5766). Do not set it.

**`channel_overrides` is NOT a gating knob.** `gateway/config.py:544-574`:
`ChannelOverride` carries exactly `model`, `provider`, `system_prompt`. There is no
per-channel `strict_mention` override anywhere in the schema.

### The narrower configuration that does exist

There is **no** knob that scopes thread-follow more precisely than
`strict_mention: false`. But blast radius can be pinned with a companion whitelist:

```yaml
extra:
  strict_mention: false
  allowed_channels: ["C08RWSTU9LK"]      # optional hard cap
```

`_slack_allowed_channels` (8387-8402) gates at **5657**, above everything, and
`is_one_to_one_dm` (5654) exempts DMs — so DMs keep working. Today it is a no-op
(Tars is in one channel anyway), but it converts "unset, so whatever channel Tars is
invited to next" into an explicit list. Recommended, not required.

**Genuinely narrower alternative to the whole change:** `reaction_triggers`
(4961-4992) — Gaetan reacts with a chosen emoji to pull a message to Tars.
`force_process` short-circuits the ladder at 5680-5681 with no ambient wake at all.
Explicit rather than implicit; costs one emoji per follow-up. Worth a decision
before shipping the broader knob.

---

## VERDICT

**CONFIRMED WITH CAVEATS.**

The three load-bearing assertions all survive:
- **Top-level containment: HOLDS.** No reachable path returns True for an unmentioned
  top-level non-thread channel message (`adapter.py:5187-5188` guards it; the raw
  `event["thread_ts"]` is passed, not the ts-fallback session key). Path 2 (5198-5202)
  lacks an `is_thread_reply` guard — a latent upstream defect — but is unreachable
  for fresh top-level traffic.
- **Plain-text name matching: NOT enabled.** `mention_patterns` is unset in both YAML
  and `.env`; `_slack_mention_patterns` has no name-derived default; `strict_mention`
  does not touch `is_mentioned`.
- **Allowlist: untouched.** Early reject at 5534-5550 fires 100+ lines before the
  ladder. No non-Gaetan user can trigger Tars.

The caveats that make it "with caveats" rather than clean:
1. **It is not a live-reload change** (the claim implies a config edit is enough).
2. **Message edits become a new wake surface** (§1a′) — not in the claim.
3. **Every past thread stays armed forever** (§7.2) — `session_reset: mode: none`
   plus a disk-backed path 3 and an API-backed path 5 means no expiry, ever.
4. **SOUL rule 4's `·` does not cover the new cases** (§7.4) — the documented
   61 s empty-response dead end is live-reachable by a Gaetan message Tars decides
   not to answer.

**Restart required: YES.** `PlatformConfig.extra` is built once at startup
(`gateway/config.py:1449 → 671 → 706`; `gateway/platforms/base.py:2760`;
`gateway/run.py:5897-5903`) and nothing rebuilds it for a connected adapter. No
config watcher exists for platform config. `ExecReload` is SIGUSR1 →
`restart_signal_handler` (`gateway/run.py:27117-27119`), an in-band **restart**.
Empirically: config.yaml written at 20:21, gateway still `MainPID=76255` from 19:40.

**Reference thread works immediately: YES** — no fresh mention needed, and it
survives the restart. Two independent restart-proof paths hit:
path 3 `_has_active_session_for_thread` (`sessions.json` key
`agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786135149.072829` exists;
`_should_reset` returns `None` for `mode: none`, `gateway/session.py:2210-2211`) and
path 5 (root text contains `<@U0BBH85NAKH>`, fetched from the Slack API).
**Correction to the brief:** it is *not* the mentioned-thread memory that carries it
— that set is RAM-only (`adapter.py:975`), never persisted, and empty.

### Exact config stanza to apply

Full nesting context, `~/.hermes/config.yaml`. **Add the `extra:` block; leave the
existing `strict_mention: true` line alone** (harmless — `extra` wins at 8268 — and
keeping it means a future operator who reverts the `extra` block lands back on the
current safe behaviour rather than on the `false` default at 8273):

```yaml
gateway:
  platforms:
    slack:
      enabled: true
      require_mention: true
      strict_mention: true          # left in place; extra.strict_mention overrides it
      unauthorized_dm_behavior: ignore
      extra:
        strict_mention: false       # NEW — read first at adapter.py:8268
        # allowed_channels: ["C08RWSTU9LK"]   # OPTIONAL hard cap, see §8
```

Apply per the repo's edit protocol: `.bak` copy first, then
`flock ~/.hermes/.wf3.lock -c '<edit>'`, merge-never-append. Then, **because the
change does not live-reload**, `systemctl --user restart hermes-gateway.service`
with `XDG_RUNTIME_DIR` exported.

Note on the merge: the *current* `strict_mention: true` at the platform level is
what `_apply_yaml_config` (8994-8995) stamps into `SLACK_STRICT_MENTION` at startup.
With `extra.strict_mention: false` present that env value is never consulted (8268
short-circuits). If instead you flip the plain key to `false` and add no `extra`,
it also works after a restart — but it is one guard (`not os.getenv`) away from
being silently ignored if `SLACK_STRICT_MENTION` ever lands in `.env`. **Use
`extra:`.**

### What Gaetan must decide before this is applied

1. **Accept a gateway restart.** Not a hot edit. Drains in-flight turns
   (`TimeoutStopSec=60`), wipes `_mentioned_threads` / `_bot_message_ts` /
   `_thread_context_cache` / dedup state. `gateway_restart_notification` is not
   disabled — a "♻ Gateway restarted" line may post to Slack; disable it in the
   same edit if a public restart ping in #gcn-sandbox is unwanted.
2. **The `·` gap (§7.4) — the single biggest risk.** SOUL rule 4 only fires for
   non-Gaetan senders. Every newly-admitted message is Gaetan's, so a "no answer
   needed" decision has no terminal and can hit the documented
   `empty_response_exhausted` / 61 s stall — in a channel with colleagues.
   Decide: amend SOUL to give Gaetan-not-addressed-to-Tars the same `·` terminal
   **before** flipping the knob, or accept the stall risk.
3. **No expiry, ever.** `session_reset: {mode: none}` + a disk-backed session path
   means the four existing `C08RWSTU9LK` threads stay armed indefinitely. Accept,
   or set an idle reset policy alongside.
4. **Optional `allowed_channels: ["C08RWSTU9LK"]`.** No-op today (Tars is in exactly
   one channel), but it makes the blast radius explicit and survives Tars being
   invited somewhere else.
5. **Consider `reaction_triggers` instead** (§8). Explicit per-message pull, zero
   ambient wake, no restart-noise trade-off — if the goal is only "follow up without
   re-@-ing", it solves the same problem with a strictly smaller surface.
