# Audit — Slack allowlist bypass on the claude.ai connector path

**Scope:** adversarial audit of the claim *"the claude.ai connector authenticates
as Gaetan and is dropped as a bot sender before the allowlist check is reached."*
**Target:** live Hermes source checkout on the Tars VM (`192.168.0.9`),
`/home/gaetan/.hermes/hermes-agent/`. **Mode:** read-only. **Date:** 2026-08-07.

All line numbers below are from the live source (adapter.py is 9100 lines). Every
load-bearing claim carries a `file:line` citation and a verbatim excerpt.

---

## TL;DR VERDICT

**Can the allowlist be bypassed? — CONDITIONAL. Currently NO (fail-closed).**

- As configured **right now** (`SLACK_ALLOW_BOTS` unset → `"none"`,
  `SLACK_ALLOWED_USERS=[U08BDJAMSRZ]`, `require_mention: true`,
  `strict_mention: true`, no `*_ALLOW_ALL_USERS`), **there is no inbound Slack
  branch that reaches model dispatch without passing the allowlist.** Every
  external branch routes through `_is_user_authorized`, which is fail-closed
  default-deny keyed on the Slack-asserted `user` id. The connector post
  specifically dies *even earlier*, at the bot-sender drop.
- The property is **config-dependent, not structural.** Flipping `SLACK_ALLOW_BOTS`
  to `mentions`/`all` (or setting `SLACK_ALLOW_ALL_USERS` /
  `GATEWAY_ALLOW_ALL_USERS`) opens real paths — one of which is a genuine
  allowlist *bypass* branch (#4466) that is presently dormant.

**Was the claim accurate?** The mechanism is real and the ordering is correct, but
the phrasing invites a wrong conclusion. See "Verdict on the claim" at the end.

---

## Current effective config values

| Setting | Effective value | Source |
|---|---|---|
| `allow_bots` | **`none`** (unset in config.yaml AND `.env`) | `_slack_allow_bots()` adapter.py:3106; `SLACK_ALLOW_BOTS` absent from `.env` |
| `require_mention` | **`true`** | `~/.hermes/config.yaml:130` |
| `strict_mention` | **`true`** | `~/.hermes/config.yaml:131` |
| `unauthorized_dm_behavior` | **`ignore`** | `~/.hermes/config.yaml:132` |
| `allowed_channels` | **empty / unrestricted** | `allowed_channels` absent from config; `SLACK_ALLOWED_CHANNELS` absent from `.env`; `_slack_allowed_channels()` returns `set()` → no channel restriction (adapter.py:8387) |
| `SLACK_ALLOWED_USERS` | **SET, 1 entry: `U08BDJAMSRZ`** (Gaetan) | `.env`; IDs are not secret |
| `SLACK_ALLOW_ALL_USERS` | absent | `.env` |
| `GATEWAY_ALLOW_ALL_USERS` | absent | `.env` |
| `GATEWAY_ALLOWED_USERS` | absent | `.env` |
| `SLACK_REACTION_TRIGGERS` | absent → reaction routing disabled | `.env`; `_slack_reaction_triggers()` returns `None` (adapter.py:4939) |

(`.env` values were confirmed by presence/count only; no token value was read or printed.)

---

## 1. Ordered gate list — real inbound path

Entry points (`adapter.py:1963-2035`): `message`, `app_mention`, `reaction_added/removed`,
`file_shared` all funnel into **`_handle_slack_message`** (`adapter.py:5239`).
`app_mention` calls it directly (`:1964`); reactions and file-shares synthesize a
`message`-shaped event and re-enter it (`:5010`, `:5093`). Slash commands and Block-Kit
actions take separate handlers but still hit the same gateway auth gate.

Ordered gates inside `_handle_slack_message`:

1. **`message_changed` normalization** (`:5265`) — an edited message is unwrapped into
   a normal `message` event so an @mention added by edit still wakes the bot once. Not a
   gate; it means edits are NOT a distinct code path — they merge into this same flow.
2. **Dedup** (`:5303`) — `self._dedup.is_duplicate(...)` → `return`. Not security.
3. **Ignored-channel drop** (`:5310`) — `_is_ignored_channel` → `return`.
4. **★ BOT-SENDER FILTER** (`:5330-5362`) — the gate the claim relies on:
   ```python
   sender_is_bot = self._event_declares_bot_sender(event)          # :5330
   if not sender_is_bot and msg_user and not event.get("client_msg_id"):
       sender_is_bot = await self._resolve_user_is_bot(...)        # users.info probe
   if sender_is_bot:
       allow_bots = self._slack_allow_bots()                       # :5338
       if allow_bots == "none":
           return                                                  # :5340  ← DROP
       elif allow_bots == "mentions":
           ... if f"<@{self._bot_user_id}>" not in text_check: return   # :5341-5352
       # "all" falls through
       if msg_user and self._bot_user_id and msg_user == self._bot_user_id:
           return                                                  # ignore own echo
   ```
5. **`message_deleted` drop** (`:5357`).
6. text/blocks/attachment extraction (`:5367-5470`) — not gates.
7. **DM determination + `disable_dms`** (`:5500-5510`): `is_dm = channel_type in {"im","mpim"}`;
   `is_one_to_one_dm = channel_type == "im"`.
8. **★ ALLOWLIST EARLY-REJECT (adapter side)** (`:5528-5548`):
   ```python
   _runner = getattr(getattr(self, "_message_handler", None), "__self__", None)
   _auth_fn = getattr(_runner, "_is_user_authorized", None)
   if user_id and callable(_auth_fn):
       _source = self.build_source(chat_id=channel_id, chat_name="",
           chat_type="dm" if is_dm else "group", user_id=user_id, user_name="")
       if not _auth_fn(_source):
           logger.warning("[Slack] Early reject of unauthorized user %s in channel %s",
                          user_id, channel_id)
           return
   ```
   Note: `build_source` here does **not** pass `is_bot`, so `_source.is_bot` defaults
   `False` (`gateway/platforms/base.py:6601`, `is_bot: bool = False`). Note also the
   guard `if user_id and ...`: an event with **no** `user` field skips this early-reject.
9. thread/session keying (`:5550-5620`) — not a gate.
10. mention detection → `is_mentioned` (`:5628`); `force_process = event.get("_hermes_force_process")` (`:5636`).
11. **Second bot-user resolution** (`:5637-5652`) — re-applies `allow_bots` to a user that
    `users.info` reveals as a bot; `none`→`return`, `mentions`+unmentioned→`return`.
12. **Channel controls block** `if not is_one_to_one_dm and bot_uid:` (`:5698`) — **skipped
    for 1:1 DMs**. Inside: `allowed_channels` whitelist, then
    `require_mention`/`strict_mention`/`thread_require_mention`/free-response wake logic
    (`:5707-5730`, e.g. `elif self._slack_strict_mention() and not is_mentioned: return`).
13. **Dispatch build** (`:6213-6303`): `build_source(..., is_bot=bool(event.get("bot_id"))
    or event.get("subtype")=="bot_message")` then `MessageEvent(...)` → handed to
    `self._message_handler`.
14. **★ GATEWAY FINAL AUTH** (`gateway/run.py:14607` cold path, `:8859` busy-session path):
    ```python
    elif not self._is_user_authorized(source):
        logger.warning("Unauthorized user: %s (%s) on %s", ...)   # run.py:14610-14611
    ```
    This runs `_is_user_authorized` a second time, on the source whose `is_bot` was set at
    step 13. `is_internal` events skip it (`run.py:14598 if is_internal: pass`); a
    `user_id is None` event still defers to `_is_user_authorized` (`:14601-14608`).
15. Model / agent dispatch.

**Ordering proof of the claim:** the bot drop (step 4, `:5340`) is at line 5340; the
allowlist early-reject (step 8, `:5546`) is at line 5546. **5340 < 5546 → the bot drop is
strictly earlier.** Confirmed from code, not comments.

### `_event_declares_bot_sender` — the connector's shape (adapter.py:3117)
```python
def _event_declares_bot_sender(self, event: dict) -> bool:
    if event.get("bot_id") or event.get("bot_profile"): return True
    if event.get("subtype") == "bot_message": return True
    profile = event.get("user_profile")
    if isinstance(profile, dict) and bool(profile.get("is_bot")): return True
    # #35777: app-originated events carry app_id and no client_msg_id
    if event.get("app_id") and not event.get("client_msg_id"): return True   # :3130
    return False
```
A claude.ai connector post (`app_id=A08SF47R6P4`, no `client_msg_id`, no `bot_id`,
`user=U08BDJAMSRZ`) matches the **`:3130`** clause → `True`. **Background fact verified.**

---

## 2. THE KEY QUESTION — per-branch: does any dispatch path skip the allowlist?

`_is_user_authorized` is enforced on **every external branch** at step 8 (adapter) and/or
step 14 (gateway). Table:

| Branch | Path | Allowlist gate on it? |
|---|---|---|
| **1:1 DM (`im`)** | `_handle_slack_message` | **YES** — early-reject (:5546) + gateway (:14607). (require/strict-mention exempt for 1:1, but NOT auth.) |
| **Public channel** | `_handle_slack_message` | **YES** — early-reject + gateway; also allowed_channels/require/strict-mention. |
| **Private channel** | `_handle_slack_message` (channel_type not im/mpim) | **YES** — same as public. |
| **MPIM (group DM)** | `_handle_slack_message` (`is_dm` true, `is_one_to_one_dm` false) | **YES** — auth + full channel controls (:5698 block runs). |
| **Thread reply** | `_handle_slack_message` | **YES** — same handler. |
| **app_mention** | `handle_app_mention → _handle_slack_message` (:1964) | **YES**. |
| **message_changed / edited** | normalized at :5265 then same flow | **YES**. |
| **file_share (message subtype)** | normal `message` | **YES**. |
| **file_shared event** | `_handle_slack_file_shared` synthesizes `message`/`file_share` with `user=…` and re-enters `_handle_slack_message` (:5093) | **YES**. |
| **slash command** | `_handle_slash_command` builds source+MessageEvent (:7735) — no adapter early-reject, but **gateway final auth (:14607) runs** | **YES**. |
| **slash-confirm / clarify / approval buttons** | `_is_interactive_user_authorized` (:6728, :6872, :7032) | **YES** — same `_is_user_authorized` (:6663), env fallback also allowlist-keyed. |
| **reaction_added/removed (emoji trigger)** | synthesizes `message` with `user=<reactor>`, `_hermes_force_process=True`, re-enters `_handle_slack_message` (:5010) | **YES** — `force_process` skips the **mention** requirement only; code comment at :4925 and re-entry through :5546 confirm "User authorization and allowed_channels still apply". (Currently moot: reaction routing disabled, triggers `None`.) |
| **Scheduled / cron / `/bg` re-entry** | gateway-internal synthesized events | **AUTH-EXEMPT** via `is_internal` (:14598) — but these are agent-authored, **not attacker-reachable from Slack ingress**. Busy-session re-entry of a *real* user message re-checks auth at :8859. |

**No external Slack branch reaches the model without an allowlist check under current
config.** The only auth-exempt path (`is_internal`) is not reachable by an inbound Slack
sender.

---

## 3. The `allow_bots` counterfactual

**Current effective value: `allow_bots = "none"` (unset — say so explicitly).**
`_slack_allow_bots()` (adapter.py:3106): reads `config.extra["allow_bots"]`, else
`os.getenv("SLACK_ALLOW_BOTS","none")`; unknown values coerce to `"none"` (:3112). Both
config.yaml and `.env` lack it → `"none"`.

There are **two** distinct consequences of flipping it, and they are different in kind:

### (a) Connector post SATISFIES the allowlist as Gaetan (not a bypass, but an opening)
Under `allow_bots="all"`:
- Step 4 no longer drops the connector post (`"all"` falls through, :5354).
- Step 8 early-reject: source built with `user_id=U08BDJAMSRZ`, `is_bot` **unset→False**.
  `_is_user_authorized` matches the allowlist on the `user` field → **passes as Gaetan**.
- Step 13 sets `is_bot=False` (connector has no `bot_id`/`subtype=bot_message`), step 14
  passes again as Gaetan → **reaches the model.**

So enabling `allow_bots` does not let the connector *bypass* the allowlist — it lets the
connector *satisfy* it, because the post genuinely carries Gaetan's `user` id. **Consequence:
anything able to post to that workspace with a user token bound to an allowlisted user
(`U08BDJAMSRZ`) would then drive the agent as Gaetan.** The claude.ai connector is exactly
such a principal (it holds Gaetan's user token). Under `allow_bots="mentions"` the same is
true **only if the post @mentions the bot** (:5341-5352).

### (b) A genuine, currently-dormant allowlist BYPASS branch (#4466)
`gateway/authz_mixin.py:646-651`, inside `_is_user_authorized`:
```python
# Bots admitted by {PLATFORM}_ALLOW_BOTS bypass the human allowlist (#4466).
if getattr(source, "is_bot", False):
    allow_bots_var = platform_allow_bots_map.get(source.platform)   # SLACK_ALLOW_BOTS
    if allow_bots_var and _platform_gate_env(allow_bots_var, "none").lower().strip() in {"mentions", "all"}:
        return True
```
This returns `True` **without any allowlist comparison** when the source is marked `is_bot`
and `SLACK_ALLOW_BOTS ∈ {mentions, all}`. It is reachable at the gateway final gate (step 14)
for events carrying `subtype=bot_message`/`bot_id` (that is what sets `source.is_bot=True` at
:6227). Such events with **`user=None`** (Slack Workflow-Builder / bot posts) skip the adapter
early-reject (step 8 is guarded by `if user_id`), so #4466 is the deciding gate → **any bot
in the workspace could drive the agent, no allowlist entry required** (subject to
`strict_mention` in channels; 1:1-DM-shaped bot traffic would be mention-exempt).

**Exact config change that opens each:** set `SLACK_ALLOW_BOTS=all` (or `mentions`) in
`~/.hermes/.env` or `platforms.slack.extra.allow_bots`. `SLACK_ALLOW_ALL_USERS=true` /
`GATEWAY_ALLOW_ALL_USERS=true` would open it even wider (authorize everyone;
authz_mixin.py:706, run.py has the allow-all shortcuts). All are currently absent.

---

## 4. What the allowlist is keyed on, and spoofability

**Keyed on `source.user_id`, i.e. the event's `user` field.** Final comparison,
`gateway/authz_mixin.py:718-779`:
```python
allowed_ids = set()
if platform_allowlist:  # SLACK_ALLOWED_USERS
    allowed_ids.update(uid.strip() for uid in platform_allowlist.split(",") if uid.strip())
...
if "*" in allowed_ids: return True
check_ids = {user_id}
...
return bool(check_ids & allowed_ids)      # :779
```
`user_id = source.user_id`, and `source.user_id` is `event.get("user")` (adapter.py:5484,
`user_id = event.get("user") or assistant_meta.get("user_id","")`). So the allowlist gate is
the raw Slack `user` field vs `SLACK_ALLOWED_USERS`.

**Is the `user` field attacker-influenceable?** Practically no:
- This gateway runs **Socket Mode** (comments at :5246 reference "Socket Mode will not
  deliver events the app manifest hasn't subscribed to"). Events arrive over Slack's
  app-authenticated WebSocket; the `user` field is **set server-side by Slack** to the
  authenticated author, not a client-supplied field an off-workspace attacker can forge.
- Slack request-signature (HMAC `X-Slack-Signature`) verification is an **HTTP Events API**
  concern; under Socket Mode the authenticated WS is the trust boundary. Either way the
  envelope is Slack-asserted, so these checks run against a trusted `user` value. There is
  no adapter code that trusts an inbound-payload-supplied identity in place of Slack's.
- The **only** way to present `user=U08BDJAMSRZ` is to actually author a message as that
  Slack user — Gaetan himself, or a token/app authorized to post **as** him (a user token).
  A random workspace *bot* posting via a bot token gets its **own** bot user id in `user`,
  which is not in the 1-entry allowlist → rejected. The claude.ai connector is the notable
  holder of Gaetan's user token; that is why it (and only it, plus Gaetan) carries the
  allowlisted id.

**`SLACK_ALLOWED_USERS` entry count: 1 — `U08BDJAMSRZ`.**

---

## 5. Fail-open vs fail-closed on error / empty / malformed

**Fail-CLOSED.** `_is_user_authorized` default path (`gateway/authz_mixin.py`):
- Empty/unset allowlist, Slack does not "enforce its own access policy", no allow-all →
  the no-allowlist branch ends at `return _auth_env("GATEWAY_ALLOW_ALL_USERS").lower() in
  {"true","1","yes"}` (:695) → **`False`** when unset.
- Allowlist set but user not in it → `bool(check_ids & allowed_ids)` on disjoint sets →
  **`False`** (:779).
- No user id at all → `if not user_id: return False` (:502) — except the documented
  chat-scoped (Telegram/QQ group) and `is_bot`+allow_bots pre-guards above it, none of
  which apply to a plain Slack user event.
- Lookup throws: the adapter early-reject wraps nothing special, but the interactive path
  explicitly falls back to an **env-allowlist** re-check on exception
  (adapter.py:6677-6707) — still allowlist-keyed, still deny-by-default. The
  `delivered_via_upstream_relay` / role / MagicMock guards use `is True` identity checks
  specifically to **refuse to fail-open** on non-bool stand-ins (authz_mixin.py comments at
  :438, :700). No branch returns `True` on error.

Conclusion: absent explicit opt-in (`*` in allowlist, `*_ALLOW_ALL_USERS`, pairing grant,
trusted-upstream marker, or `SLACK_ALLOW_BOTS` for bot-shaped events), the answer is deny.

---

## Verdict on the claim

**The claim is ACCURATE in mechanism and ordering, but the framing is imprecise and can
mislead.**

- ✔ "authenticates as Gaetan" — true, the connector post carries `user=U08BDJAMSRZ`.
- ✔ "dropped as a bot sender" — true, `_event_declares_bot_sender` matches on
  `app_id` + no `client_msg_id` (adapter.py:3130) and the `allow_bots="none"` gate returns
  at adapter.py:5340.
- ✔ "before the allowlist check is reached" — true, 5340 < 5546; the drop is strictly
  earlier. The `mentions` bypass at :5341 is never reached because the effective value is
  `none`.
- ✔ "fail-closed / strictly more restrictive" — true **for the current config**: the
  connector post dies at the bot drop and never reaches the model.

**Where the phrasing misleads:** it presents "dropped before the allowlist" as the *safety
property*, when the actual load-bearing fact is `allow_bots="none"`. Two caveats the phrasing
hides:
1. The bot drop being "before" the allowlist is not a fail-closed guarantee independent of
   config — flip `SLACK_ALLOW_BOTS` and the drop disappears; the event then reaches the
   allowlist and **passes as Gaetan** (it carries his id). The allowlist would **not**
   independently reject a connector post. So the protection is the `allow_bots=none`
   default, not the allowlist, and not the ordering per se.
2. Enabling `allow_bots` also arms a genuine allowlist-*bypass* branch (#4466,
   authz_mixin.py:646) for `bot_message`/`bot_id`+`user=None` events, which would admit
   arbitrary workspace bots regardless of the allowlist.

Net: **no hole today (fail-closed NO)**; the mechanism the claim names is real; but "dropped
before the allowlist, therefore fail-closed" should be stated as "dropped by the
`allow_bots=none` default before the allowlist; do not change `SLACK_ALLOW_BOTS`, and note
the allowlist alone would admit a connector post as Gaetan."

---

## Orchestrator verification of this audit (independent re-check on the VM)

The findings above were spot-checked against source rather than accepted. Result:
**substance confirmed, one citation wrong, and one detail sharper than reported.**

### Correction — the bypass branch is at `authz_mixin.py:499-501`, not `:646`

`:646` is an unrelated adapter-policy branch. The real code, read verbatim from
`~/.hermes/hermes-agent/gateway/authz_mixin.py`:

```python
# Bots admitted by {PLATFORM}_ALLOW_BOTS bypass the human allowlist (#4466).
platform_allow_bots_map = { ... Platform.SLACK: "SLACK_ALLOW_BOTS" }
if getattr(source, "is_bot", False):
    allow_bots_var = platform_allow_bots_map.get(source.platform)
    if allow_bots_var and _platform_gate_env(allow_bots_var, "none").lower().strip() in {"mentions", "all"}:
        return True
if not user_id:
    return False
```

`return True` = authorized, with no allowlist comparison. The comment states the bypass
as intent (#4466), and upstream ships a test for it:
`tests/gateway/test_slack_bot_auth_bypass.py`. This is designed behaviour, not a defect.

### Sharper than reported — it is ONE config line, not an env-only knob

The audit frames the trigger as the `SLACK_ALLOW_BOTS` env var. But the Slack adapter
bridges YAML into that env var at `plugins/platforms/slack/adapter.py:9006-9007`:

```python
if "allow_bots" in slack_cfg and not os...:
    os.environ["SLACK_ALLOW_BOTS"] = str(slack_cfg["allow_bots"]).lower()
```

So setting `slack.allow_bots: mentions` in `~/.hermes/config.yaml` does **both** things at
once: it stops the adapter drop at `adapter.py:5339` *and* arms the gateway authz bypass at
`authz_mixin.py:501`. There is no second, separate opt-in. Env wins if already set; config
fills it in otherwise.

### Doc/code mismatch worth knowing before anyone flips it

The operator-facing INFO line the adapter logs when `allow_bots != none`
(`adapter.py:2211-2221`) tells the operator that *"the other bot's Slack user id [must be]
in SLACK_ALLOWED_USERS or GATEWAY_ALLOW_ALL_USERS=true"* — i.e. it implies the human
allowlist still applies to bot traffic. `authz_mixin.py:499-501` says it does not. The
guidance is more conservative than the code, so an operator reading only the startup log
would underestimate what the flag opens.

### Current state re-confirmed independently

```
SLACK_ALLOW_BOTS / SLACK_ALLOW_ALL_USERS / GATEWAY_ALLOW_ALL_USERS
  → 0 of 3 present in ~/.hermes/.env  (defaults apply)
slack.allow_bots → unset → "none" → adapter drops bot events at adapter.py:5339
```

So the bypass branch is **dormant today** and the deployment is fail-closed, as the audit
concludes. What protects it is the `allow_bots: none` default — **not** the check ordering.
That distinction is the correction owed to the orchestrator's original phrasing.

### Standing risk note

`SLACK_ALLOWED_USERS` (1 entry) is, since P4, the sole gate on transitive Orca access.
`slack.allow_bots` is now a **one-line, non-obvious way to widen that gate to every bot in
the workspace** — including anything that can post via Slack Workflow Builder, which arrives
`user=None` and is exactly the traffic the bypass exists to admit. Treat that key as
security-critical config: it belongs on the same do-not-change footing as
`SLACK_ALLOWED_USERS` itself.
