# WF4 Probe 15 — Tars-initiated delegation-lite to cooper

Spec: PLAN.md WF4 row, "Probe 15 added 2026-08-07 (Gaetan)": prove the full chain
model turn → Tars shell tool → ssh mesh → cooper → output back in the reply — not
just transport (transport itself was already proven in probe 14).

**Verdict: PASS**, first attempt — no retry needed, the known
`"Codex response remained incomplete after 3 continuation attempts"` failure mode
(see p13-bg-delivery.md, p01-dm-roundtrip.md) did not occur this run.

Date 2026-08-07, VM `tars` @ 192.168.0.9, window 19:46:03Z → 19:46:19Z. Read-only
throughout: no sops, no unit lifecycle command, no config/.env edit, no Slack
message sent, 192.168.0.3 never touched, no git command.

## 1 — the chat turn (run on the VM)

```
$ date -Is
2026-08-07T19:46:03+00:00
$ ssh -o BatchMode=yes -o ConnectTimeout=10 gaetan@192.168.0.9 \
    'export XDG_RUNTIME_DIR=/run/user/$(id -u); timeout 300 ~/.local/bin/hermes chat -Q -q \
     "Run this exact command via your shell tool and report its raw output verbatim: \
      ssh cooper \"hostname && uptime -p\""'

session_id: 20260807_194604_3befdf
cooper
up 1 day, 20 hours, 21 minutes
EXIT=0
$ date -Is
2026-08-07T19:46:15+00:00
```

Reply contains cooper's hostname (`cooper`) and an `uptime -p` line
(`up 1 day, 20 hours, 21 minutes`) — criterion (a) met.

## 2 — agent.log proof the MODEL invoked the tool (criterion (b))

Session id from the chat output (`20260807_194604_3befdf`) grepped straight out of
`~/.hermes/logs/agent.log` on the VM:

```
$ date -Is
2026-08-07T19:46:19+00:00
$ ssh gaetan@192.168.0.9 'grep -n "20260807_194604_3befdf" ~/.hermes/logs/agent.log'
1690:2026-08-07 19:46:06,450 INFO [20260807_194604_3befdf] agent.turn_context: conversation turn: session=20260807_194604_3befdf model=gpt-5.6-sol provider=openai-codex platform=cli history=0 msg='Run this exact command via your shell tool and report its raw output verbatim: s...'
1692:2026-08-07 19:46:09,977 INFO [20260807_194604_3befdf] agent.conversation_loop: API call #1: model=gpt-5.6-sol provider=openai-codex in=23870 out=49 total=23919 latency=3.5s
1693:2026-08-07 19:46:10,003 INFO [20260807_194604_3befdf] tools.terminal_tool: Creating new local environment for task default...
1694:2026-08-07 19:46:10,018 INFO [20260807_194604_3befdf] tools.environments.base: Session snapshot created (session=84bfe108590f, cwd=/home/gaetan)
1695:2026-08-07 19:46:10,018 INFO [20260807_194604_3befdf] tools.terminal_tool: local environment ready for task default
1696:2026-08-07 19:46:10,874 INFO [20260807_194604_3befdf] agent.tool_executor: tool terminal completed (0.89s, 83 chars)
1697:2026-08-07 19:46:14,462 INFO [20260807_194604_3befdf] agent.conversation_loop: API call #2: model=gpt-5.6-sol provider=openai-codex in=23969 out=19 total=23988 latency=3.6s cache=23040/23969 (96%)
1698:2026-08-07 19:46:14,477 INFO [20260807_194604_3befdf] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=2/500 budget=2/500 tool_turns=1 last_msg_role=assistant response_len=37 session=20260807_194604_3befdf
1699:2026-08-07 19:46:14,492 INFO [20260807_194604_3befdf] tools.terminal_tool: Manually cleaned up environment for task: default
1700:2026-08-07 19:46:14,492 INFO [20260807_194604_3befdf] tools.terminal_tool: Cleaned 1 environments
```

Reading: API call #1 returns a tool-call, `tools.terminal_tool` spins up a fresh
local shell environment (`session=84bfe108590f`), `agent.tool_executor` reports
`tool terminal completed (0.89s, 83 chars)`, then API call #2 (fed the tool output,
`cache=96%` — same context reused) produces the final text reply.
`Turn ended: … tool_turns=1` confirms exactly one real tool turn occurred inside
this session — the model invoked the shell ("terminal") tool, which is what ran
the nested `ssh cooper "hostname && uptime -p"`; it did not hallucinate the output.
Same tool name (`terminal`) and log shape as the prior confirmed-real tool call in
`p13-bg-delivery.md` (`tool terminal completed (10.27s, 45 chars)` for the `sleep
10` background job) and the MCP-tool pattern in `p09-10-mcp-model.md` (session id
in the log line matches the session id printed by the chat turn).

## 3 — cross-check, direct on cooper

```
$ date -Is
2026-08-07T19:45:59+00:00
$ hostname; uptime -p
cooper
up 1 day, 20 hours, 21 minutes
```

and again after the probe:

```
$ date -Is
2026-08-07T19:46:31+00:00
$ hostname; uptime -p
cooper
up 1 day, 20 hours, 21 minutes
```

Direct cooper output matches the chat reply exactly (`up 1 day, 20 hours, 21
minutes`, same rounding bucket before/during/after the probe) — no drift beyond
the expected "minutes" granularity of `uptime -p`.

## Summary

| Check | Result |
|---|---|
| (a) reply has cooper hostname + uptime line | PASS — `cooper` / `up 1 day, 20 hours, 21 minutes` |
| (b) agent.log shows real tool execution for this session | PASS — `tools.terminal_tool` create→exec→cleanup, `agent.tool_executor: tool terminal completed`, `tool_turns=1`, session id matches |
| cross-check vs direct cooper `uptime -p` | PASS — exact match |
| Known "incomplete after 3 continuation attempts" failure | Did not occur — single attempt, PASS first try |

No sops invoked. No unit lifecycle command run. No config/.env edited. No Slack
message sent. 192.168.0.3 never touched. No git command run.

**Verdict: PASS.**
