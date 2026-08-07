# P1 — Thread-follow: let Tars answer unmentioned replies in threads it is already in

**Status:** proposed, NOT applied · **Source:** thread-behavior peer, 2026-08-07 ·
**Evidence:** `status/probes/wf5/{thread-behavior,source-gating,gating-verification,log-forensics-reference-thread}.md`
· **Decide:** Gaetan

## Problem

In a thread where Tars had answered 13 s earlier, unmentioned replies from Gaetan
were dropped silently; only re-mentioning `@Tars` gets an answer. Asked for in
thread `C08RWSTU9LK` / `1786135149.072829`.

## What changes

`strict_mention` off, `require_mention` stays on. This unlocks machinery that
**already exists but is unreachable**: `_should_wake_on_unmentioned_message`
(`adapter.py:5164`) with 5 wake paths — bot-sent root, mentioned-thread memory,
active session by thread ts, bot-authored root via API, parent-text mention.
`strict_mention: true` returns before it (`:5707-5708`).

Top-level messages stay gated: the wake check needs a real `thread_ts`
(`:5187-5188`), so an unmentioned message in the channel body is still dropped.

```yaml
gateway:
  platforms:
    slack:
      enabled: true
      require_mention: true        # unchanged
      strict_mention: false        # was true
      extra:
        strict_mention: false      # authoritative — _slack_strict_mention() reads extra first (:8268)
      # allowed_channels: [C08RWSTU9LK]   # optional cap, see P1-b
```

Both keys are set because the plain key reaches the adapter only via the
`SLACK_STRICT_MENTION` env var, stamped under a `not os.getenv()` guard
(`adapter.py:8994`) — already `"true"` in the running process.

## Order of operations (the SOUL step is not optional)

1. **SOUL amendment — MUST land first.** Rule 4's `·` fallback only fires for
   *non-Gaetan* senders. Every message this change newly admits is Gaetan's, so
   "decide not to answer" has no terminal → `empty_response_exhausted`, 61 s
   stall, user-facing error (the exact failure mode fixed earlier today,
   `wf4/diag-empty-response.md`). The reference thread contains "Ptin t'es nul" —
   venting, not a question. Without this step, every aside becomes a visible
   error in a public channel.
2. Config edit: `.bak` first, merged not appended, under `flock ~/.hermes/.wf3.lock`.
3. **Gateway restart — required.** `PlatformConfig.extra` is built once at startup
   (`config.py:1441→671→706`, `base.py:2760`); no watcher rebuilds it for a
   connected adapter. Live-reload does not cover this key.

## Safety — settled, not assumed

The allowlist early-reject (`adapter.py:5534-5550`, WARNING at `:5546`) runs
**104 lines before** the first mention/thread branch (`:5654`), with two further
`_is_user_authorized` checks downstream (`gateway/run.py:14610`, `:8859`) added
explicitly against shared-thread injection. Live-proven: a colleague's in-thread
message rejected in 2 s. **Thread-follow cannot leak to non-Gaetan users.**

No plain-text name matching either: `mention_patterns` unset and
`_slack_mention_patterns` (`:8423`) has no bot-name default — Tars will not wake
on the bare word "tars".

## Risks / open sub-decisions

| Risk | Detail | Sub-decision |
|---|---|---|
| Blast radius | `allowed_channels` unset. Today: exactly **one** channel (`C08RWSTU9LK` #gcn-sandbox) + 27 IMs, no MPIMs, no private channels. 1:1 DMs skip the gate entirely | **P1-b:** cap to `[C08RWSTU9LK]` (recommended — no-op today, prevents surprise in channels joined later) or leave open |
| Threads never expire | `session_reset: {mode: none}`, no TTL path reachable (`gateway/session.py:2210-2211`) — every past thread stays armed forever | **P1-c:** accept (recommended, matches how a colleague would behave) or introduce a TTL (own investigation, parks this proposal) |
| Message edits | Edits are a wake surface (`adapter.py:5291-5297`) — editing an old thread message can wake Tars | Accept, or fold into the TTL work |
| Seeding | `_mentioned_threads` is RAM-only (`:975`) and never populated while strict mode is on. The reference thread still works immediately post-change via the active-session path; other old threads may need one fresh mention to arm | Informational |

## Verification after applying

1. Unmentioned reply in the reference thread → Tars answers.
2. Unmentioned **top-level** message in the same channel → still dropped.
3. Non-Gaetan in-thread message → allowlist WARNING still fires.

(Live probes must be sent **natively** with `SLACK_MCP_XOXC_TOKEN`/`XOXD` —
the claude.ai connector cannot trigger Tars at all, see `docs/facts.md`.)

## Rollback

Restore `config.yaml.bak-*`, restart. The SOUL amendment is independent and
harmless to keep — it only gives Tars a way to acknowledge without answering.

## Decision

- [ ] Apply full package (SOUL + config + restart)
- [ ] SOUL amendment only now, config flip later
- [ ] Don't apply
- P1-b channel cap: [ ] cap to #gcn-sandbox · [ ] leave open
- P1-c thread TTL: [ ] accept no-TTL · [ ] investigate a TTL first
