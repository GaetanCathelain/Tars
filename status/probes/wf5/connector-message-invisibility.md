# WF5 probe — claude.ai Slack connector messages are invisible to Tars (silent drop, not a bug)

Read-only diagnostic. VM untouched, no restarts, no Slack writes. Evidence via
`conversations.replies` (bot token, `-K` stdin, per `operating-tars` skill) and
`adapter.py` source reads on the Tars VM (`~/.hermes/hermes-agent/`).

## 1. Raw JSON field comparison

Fetched full `conversations.replies` for both threads in `C08RWSTU9LK`.

| field | connector top-level `1786136615.681079` | connector reply `1786136621.808629` (has `<@U0BBH85NAKH>` mention) | native Gaetan `1786135283.424339` (thread root `1786135149.072829`) |
|---|---|---|---|
| `user` | `U08BDJAMSRZ` | `U08BDJAMSRZ` | `U08BDJAMSRZ` |
| `bot_id` | **absent** | **absent** | absent |
| `app_id` | **`A08SF47R6P4`** | **`A08SF47R6P4`** | **absent** |
| `client_msg_id` | **absent** | **absent** | **present** (`5439e3cd-ced2-4027-b6c8-72d43d7bd432`) |
| `subtype` | absent | absent | absent |
| `bot_profile` | absent | absent | absent |
| `username` | absent | absent | absent |
| `team` | `T7V1UGJ82` | `T7V1UGJ82` | `T7V1UGJ82` |
| `user_team` / `source_team` | absent (not returned by this endpoint for either) | absent | absent |
| `parent_user_id` | n/a (thread root) | `U08BDJAMSRZ` | `U08BDJAMSRZ` (this is itself a mid-thread reply) |
| blocks | `rich_text` + a `context` block with `"*Sent using* <@U0AFH1P6EKT>"` | same pattern, mention is a `rich_text` `user` element | plain `rich_text`, no context/attribution block |

**The differentiating field is `app_id` combined with the absence of `client_msg_id`.** `bot_id` is NOT set (the connector authenticates with Gaetan's *user* token, so Slack does not attribute the post to a bot user) — only `app_id` is stamped, and unlike every native human message, `client_msg_id` (Slack's client-generated UUID, added by the real Slack composer) is missing. That combination is the only detectable signal; `bot_profile`, `username`, `user_team`, `source_team` are identical/absent on both sides and give no signal.

## 2. The filter — `adapter.py`

`plugins/platforms/slack/adapter.py` (source checkout on the VM, `~/.hermes/hermes-agent/`):

**`_event_declares_bot_sender`, line 3117-3132** — the actual trigger:

```python
def _event_declares_bot_sender(self, event: dict) -> bool:
    """Return True when the Slack event itself identifies a bot sender."""
    if event.get("bot_id") or event.get("bot_profile"):
        return True
    if event.get("subtype") == "bot_message":
        return True
    profile = event.get("user_profile")
    if isinstance(profile, dict) and bool(profile.get("is_bot")):
        return True
    # Some Slack app-originated events arrive without subtype=bot_message
    # or bot_id, but they still carry app_id and no client_msg_id. Real
    # human-authored messages normally carry client_msg_id, so treat the
    # combination as app/bot-authored (#35777).
    if event.get("app_id") and not event.get("client_msg_id"):   # line 3130
        return True
    return False
```

This is a **known, deliberately-coded case** (comment cites issue `#35777`) — exactly the connector's signature (`app_id` present, no `client_msg_id`) is what line 3130 is written to catch.

**`_handle_slack_message`, line 5239 onward** — where the drop happens:

```python
msg_user = event.get("user", "")                              # 5329
sender_is_bot = self._event_declares_bot_sender(event)         # 5330
if not sender_is_bot and msg_user and not event.get("client_msg_id"):
    sender_is_bot = await self._resolve_user_is_bot(...)       # 5332
if sender_is_bot:                                               # 5337
    allow_bots = self._slack_allow_bots()
    if allow_bots == "none":                                    # 5339
        return
    elif allow_bots == "mentions":                               # 5341
        text_check = _slack_mention_detection_text(event)
        if self._bot_user_id and f"<@{self._bot_user_id}>" not in text_check:
            logger.debug(...)                                    # debug-only
            return
    # Always ignore our own messages to prevent echo loops         # 5352
    if msg_user and self._bot_user_id and msg_user == self._bot_user_id:
        return
```

`_slack_allow_bots()` (line 3106) reads `config.extra.allow_bots` / `SLACK_ALLOW_BOTS`, defaulting to `"none"` when unset. **Confirmed unset** on the VM — `grep -i allow_bots ~/.hermes/.env ~/.hermes/config.yaml` → no match in either file, so the code default (`"none"`) is in effect.

**Which check the connector-sent pair hits:** line 3130 (`app_id` present + no `client_msg_id`) → `_event_declares_bot_sender` returns `True` → `sender_is_bot = True` → `allow_bots == "none"` (default) → **`return` at line 5339**, unconditionally, for BOTH messages — the top-level one and the one with the explicit `<@U0BBH85NAKH>` mention. The `allow_bots == "mentions"` branch (line 5341, which *would* have let the mentioned reply through) is never reached because the policy is `"none"`, not `"mentions"`.

There is **no log line on this path**. The only logging in `_handle_slack_message` before this return is the entry-log at the very top of the function, gated behind `logger.isEnabledFor(logging.DEBUG)` — and the gateway runs at INFO. So the drop is completely silent at the current log level, by design (comment: "backward-compatible" default).

## 3. Does the event even arrive?

Yes, structurally — the message event does not require a bot-marker to be *delivered*; Socket Mode delivers `message` events uniformly regardless of who authored the underlying Slack post (user token via an app vs. a plain human token). The registered listener for `message` (and `app_mention`) is `handle_message_event` → `_handle_slack_message` (adapter.py ~1949-1952) — a real, named listener, not the `.*` catch-all. The catch-all (`handle_unhandled_event`, ~line 2039) only fires for event *types* Hermes has no listener for (e.g. `user_change`, `pin_added`); `message` is always dispatched to `_handle_slack_message`, so the catch-all is irrelevant here — it does not explain the silence.

**The current INFO log level cannot distinguish "event arrived and was filtered at line 5339" from "event never arrived."** Both look identical: zero log lines. The only way to tell them apart *without applying it*:
- Raise the log level to DEBUG — the entry-log at the top of `_handle_slack_message` (before line 3130's check is even reached) would fire and prove arrival regardless of what happens after.
- Alternatively, temporarily set `SLACK_ALLOW_BOTS=mentions` (or `all`) and re-send a mention through the connector — if Tars answers, that proves both arrival and correct routing, `allow_bots` was simply the wrong policy for this sender.

Neither was applied (read-only). Given the code path is a deliberate, named check with a cited issue number (not a crash or exception), and every other event type on the gateway is flowing normally (concurrent DM traffic in `D0BBYNM01BL`), the far more likely explanation is **arrival + silent filter**, not non-delivery — but this report treats that as inference, not confirmed fact.

## 4. Operational conclusion

**No** — the claude.ai Slack connector CANNOT be used to trigger Tars in channel `C08RWSTU9LK` (or any channel/DM under the current config). Every message it posts carries `app_id` with no `client_msg_id`, which `_event_declares_bot_sender` (adapter.py:3130) classifies as bot/app-authored, and the gateway's `allow_bots` policy is `"none"` (default, confirmed unset on the VM) — so **100% of connector-sent messages are silently dropped before any processing, mention or not.** This is true even for messages that explicitly `@mention` the bot, because the `"mentions"` bypass is never reached under policy `"none"`.

**Consequence for future probes:** the connector is unusable for anything that needs to *trigger* Tars — it is read-only from Tars's perspective (fine for verifying what's visible in Slack, useless for driving live agent-response probes). Any live probe that needs Tars to answer must be sent from Gaetan's native Slack client/session, not the claude.ai connector.

## 5. DMs affected too?

**Yes, by source inspection — not separately re-tested live.** `_handle_slack_message` is the single registered handler for the `"message"` event type regardless of channel type; the DM/`channel_type` classification (`is_dm`, `is_one_to_one_dm`, adapter.py ~5509-5528) happens *after* the bot-sender filter block (~5329-5359), not before it, and there is no `channel_type == "im"` bypass anywhere ahead of line 3130/5339. So a connector-sent message into the home DM `D0BBYNM01BL` would hit the identical `app_id`-without-`client_msg_id` check and be dropped the same way, silently, regardless of `allow_bots` policy being evaluated per-message rather than per-surface.

## Redactions

No secret values appear in this report. `SLACK_BOT_TOKEN` was sourced inside the remote shell and passed to `curl` via a `-K` config file on stdin per the hard rule; only its name was ever grepped, never printed.
