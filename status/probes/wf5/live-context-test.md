# Live probe — thread-context injection vs. model tool call (WF5)

Read-only log forensics. VM: `ssh gaetan@192.168.0.9` (clock UTC). Source:
`~/.hermes/hermes-agent/plugins/platforms/slack/adapter.py`. Logs:
`~/.hermes/logs/{gateway,agent,errors,mcp-stderr}.log`.

## Probe under test

Channel `C08RWSTU9LK`, fresh cold thread:

- Root, ts `1786136615.681079` (= **2026-08-07 21:03:35 UTC**) — top-level,
  NO mention: "Note pour plus tard (test contexte thread, ignorer) : le code
  de vérification est TARSTHREAD-7Q4X."
- Reply, ts `1786136621.808629` (= **2026-08-07 21:03:41 UTC**) — in-thread,
  WITH `<@U0BBH85NAKH>` mention, asking Tars to recite the code without using
  any Slack tool.

Timestamp conversion used:
```
ssh gaetan@192.168.0.9 'date -u -d @1786136615 +"%Y-%m-%d %H:%M:%S UTC"; date -u -d @1786136621 +"%Y-%m-%d %H:%M:%S UTC"'
→ root ts -> 2026-08-07 21:03:35 UTC
→ reply ts -> 2026-08-07 21:03:41 UTC
```
(Note: this is ~20 min later than the "20:43:40 UTC" figure named in the task
brief — the brief's figure appears stale; all polling below used the
ts-derived time as ground truth.)

## What was checked and how

1. Exact-ts grep, repeatedly, across `gateway.log` / `agent.log` / `errors.log`:
   ```
   grep -n "1786136615.681079\|1786136621.808629" ~/.hermes/logs/{gateway,agent,errors}.log
   ```
2. Alternate-format grep (in case of formatting drift): `136615\.68|136621\.80|136615681079|136621808629` — no hits, any format.
3. Broad grep for the bare epoch prefix `1786136615|1786136621` across all of
   `~/.hermes/logs/` (all files) — **zero hits, anywhere, in any file.**
4. Broad grep for `TARSTHREAD` across all of `~/.hermes/logs/` — **zero hits.**
5. Channel-scoped grep for `C08RWSTU9LK` across `gateway.log` (full history) to
   establish the channel's baseline activity pattern.
6. Full unfiltered tail of `gateway.log`, `agent.log`, `errors.log`,
   `mcp-stderr.log` around the probe window.
7. `systemctl --user status hermes-gateway.service` — confirmed the gateway
   process was up, healthy, and actively processing (unrelated DM traffic)
   throughout the whole window.
8. Repeated polling loops (two passes, ssh-side `while` loop with 10–15s
   sleep, condition-checked each iteration — not a blind sleep) spanning
   **2026-08-07 21:04:20 UTC through 21:10:34 UTC**, i.e. from ~45s before
   the reply ts to **6m53s after** the reply ts (more than double the ~3 min
   allowance in the task brief).

## Findings

### 1. Did Tars answer? — NO, not within the observed window

No `inbound message`, no response, no error, no trace of either probe
message's timestamp appeared in any log file at any point from before the
root message was sent through 6m53s after the reply. The last log activity
attributable to channel `C08RWSTU9LK` at all is:

```
2026-08-07 20:46:05,116 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05L... in channel C08RWSTU9LK
2026-08-07 20:46:09,178 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05L... in channel C08RWSTU9LK
2026-08-07 20:46:55,351 INFO gateway.run: Agent cache idle-TTL evict: session=agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786131725.413279 (idle=3882s)
```

— all **17+ minutes before** the root probe message's ts (21:03:35). Nothing
channel-related appears after that until the polling window ended at 21:10:34.

The gateway itself was healthy and busy the whole time — it processed roughly
a dozen DM turns (`chat=D0BBYNM01BL`) in the same window, so this is not a
crashed/stuck process; it simply never logged anything for `C08RWSTU9LK`
after 20:46:55.

**Caveat / limitation of this forensics pass:** this agent has no Slack
read access (no MCP Slack tools were granted to this session) and made no
Slack calls per the read-only mandate, so it cannot independently confirm
whether the two probe messages actually reached Slack / this channel at all.
The complete absence of *any* log trace (not even a mention-gate reject, which
DOES get logged for unauthorized users) is consistent with either (a) the
Slack event never being delivered to the gateway for this channel in this
window, or (b) some other silent failure upstream of the adapter's normal
logging points. Log evidence alone cannot distinguish these from here.

### 2. Did `_fetch_thread_context` fire for thread `1786136615.681079`? — INDETERMINATE (no evidence either way)

No log line referencing this thread ts, `conversations.replies`, `[Thread
context`, or any `_has_active_session_for_thread` outcome appears anywhere.
Since there is no evidence the adapter processed this thread at all (see §1),
there is no basis to say the cold-start hydrate path ran. **Not confirmed
fired, not confirmed skipped — simply never observed.**

### 3. Did the MODEL call any Slack MCP tool during this turn? — NO TURN OCCURRED; NO TOOL CALLS OBSERVED

`agent.log` and `mcp-stderr.log` were checked end-to-end for the window. The
only `mcp__slack__*` tool activity in `agent.log`/`mcp-stderr.log` near this
time belongs to the **unrelated DM session** (`agent:main:slack:dm:...`,
session `20260807_205326_2f2ee68a`), e.g.:

```
agent.log: 2026-08-07 21:02:17,482 INFO [20260807_205326_2f2ee68a] agent.tool_executor: tool mcp__slack__conversations_history c...
mcp-stderr.log: {"level":"info","timestamp":"2026-08-07T21:02:16Z","message":"Request received","app":"slack-mcp-server","tool":"conv...
mcp-stderr.log: {"level":"info","timestamp":"2026-08-07T21:02:17Z","message":"Request finished","app":"slack-mcp-server","tool":"conv...
```

That call is timestamped **before** the probe reply ts (21:02:17 < 21:03:41)
and belongs to a different chat (`D0BBYNM01BL`, the DM), not to a turn
answering the channel probe — there is no evidence it is related to
`TARSTHREAD-7Q4X`. **No tool call attributable to a turn processing the probe
thread was found, because no turn processing that thread was found at all.**

### 4. Tars' final answer text — NOT PRODUCED

No response was sent to channel `C08RWSTU9LK` in the observed window. Cannot
quote an answer or check for `TARSTHREAD-7Q4X` in it.

### 5. Root message (`1786136615.681079`) zero log lines — CONFIRMED, but not diagnostic on its own here

Confirmed zero log lines for the root ts, consistent with the known bare-return
silent drop for unmentioned messages at `adapter.py:5707-5708`. However, given
that the REPLY (which *does* carry a mention and should not hit that silent-drop
path) also produced zero log lines, the root's silence cannot by itself be
attributed with confidence to the mention-gate code path specifically — see the
caveat in §1: it's equally consistent with neither message having reached the
adapter's logging points at all.

### 6. Injected `[Thread context — …]` block — NOT OBSERVED

No such block, or any per-message tag (`[thread parent]`, `[assistant]`,
`[unverified]`), appears in any log for this window. Nothing to quote.

### 7. End-to-end latency — N/A

No answer was produced within the observation window (reply ts 21:03:41 UTC
→ polling ended 21:10:34 UTC with nothing, i.e. **at least 6m53s with no
response**, more than double the task's ~3 min allowance).

## Verdict

**Inconclusive — insufficient evidence to attribute the code to either
adapter-injected context or a model tool call, because no evidence exists
that Tars processed this thread at all within the observed window.** This
is a negative result, not a null hypothesis confirmation: the expected
`inbound message` log line for the mentioned reply never appeared, which is
itself the primary anomaly and should be investigated (e.g. confirm the two
probe messages actually posted to Slack / to this exact channel and thread,
and check gateway Slack event-delivery health beyond what log-only forensics
can show).
