# SOUL rule 10 — read the record before answering for the past

2026-08-13, orchestrator session (Fable) in worktree `verify-slack-messages-dm`.
Incident: `C0BP2GZUFSR` thread `1786619456.260179`, flagged reply
`1786619650.634919` (13:14:10 CEST). Full diagnosis: workflow `wf_a8d1f331-193`
(4 agents); design artifact archived in session scratchpad (`FIX-DESIGN.md`).

## Root cause — doctrine, not mechanism

- The gateway can only ever inject the current thread (`adapter.py:5623/7197/7495`;
  `docs/facts.md` l.88: no `conversations.history` anywhere in the gateway). The
  evidence Tars needed sat in a neighbouring thread (`1786618059`, its own "I'll
  return the verified verdict in our DM"), another session's tool trace, and
  channel history — none reachable by injection.
- Tars had the tools (log census: `conversations_search_messages` ×240,
  `conversations_replies` ×111, `conversations_history` ×50) and used none in the
  flagged turn: `api_calls=1`, zero `tool_executor` lines. It answered from
  narrative memory.
- No SOUL line or any of the 118 skills required reading before answering; the one
  skill it did open (`slack-channel-context` rule 3) actively rationed history reads.

## Fix landed (both mirrors, byte-identical)

1. `SOUL.md` — new hard rule 10 inserted after rule 9: fetch the pinged thread
   (`conversations_replies`), the last ~20 messages of the pinged surface
   (`conversations_history`), and — when the act in question is Tars' own — the
   session that performed it (`session_search`/`lcm_grep`), before answering any
   question about what happened. Gate: "either I checked, or I say I have not
   checked." Scoped to questions about the past; forward-looking pings cost nothing.
2. `skills/orchestration/slack-channel-context/SKILL.md` → v0.2.0 — rule 3's
   history-rationing scoped to channel-purpose classification only (it contradicted
   rule 10 in the one document Tars consulted during the incident).

Evidence:

- VM landing: `.bak-20260813` copies of both files, then tmp+mv.
- `ssh tars cat ~/.hermes/SOUL.md | diff - SOUL.md` → empty; same for the skill.
- Doctrine probe (fresh session `20260813_113706_5d1b17`, `hermes chat -Q`):
  quotes rule 10 verbatim, paraphrases the gate correctly.
- System prompt is built once per session and `session_reset.mode: none`, so
  **existing threads keep the old SOUL forever** — any behavioral test must be a
  new thread/DM turn.

## Misroute cause recovered (was confabulated in the apology)

`hermes sessions export --session-id 20260813_104740_69c4ca59` shows exactly two
`conversations_add_message` calls: `channel_id D0BBYNM01BL` (Gaetan's DM —
correct) and `channel_id D08K34MA3QT` (reached Oli). `D08K34MA3QT` occurs exactly
once in the exported transcript — in the call arguments, never in any prior tool
result — i.e. Tars sent to a destination ID it never looked up in that session.
The flagged apology's "home channel presented as our DM" story is disproven.

## Open items

- **Send-side gap, not covered by rule 10** (which gates answering): nothing yet
  requires verifying a `channel_id` before sending to it. Proposal for Gaetan:
  one more SOUL sentence ("I never send to a channel_id I have not resolved this
  session"). Awaiting his call.
- The 2026-08-13 FAIL verdict's follow-ups remain open: `config.yaml` l.137/l.243
  and four enabled cron jobs + reconciler still target `C0BP2GZUFSR`, contradicting
  `USER.md` l.13.
- Latent watermark bug recorded, not fixed: thread rooted in a top-level post
  never sets a watermark, so `adapter.py:5885` rehydrate cannot fire after a
  gateway restart mid-thread. Patch only if observed.
- Real-behavior test needs Gaetan (claude.ai connector is dropped as bot-sender):
  in a NEW thread, a "what happened / who sent what" question whose answer sits
  in a different thread → pass = `conversations_history`/`replies` tool lines
  between `inbound message` and `response ready` in `agent.log`. Negative test:
  a trivial forward-looking ping must not grow Slack reads.
