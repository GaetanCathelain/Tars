# GCN-7 — WF6 end-to-end verification (T7 §Verification pass, 2026-08-11 evening UTC)

The T7 / GCN-7 pass over `docs/specs/wf6-linear-integration.md` §Verification,
run 2026-08-11 between **17:23 and 18:22 UTC** by five probe agents against the
live Tars VM (`192.168.0.9`), Linear workspace `mobile-club` and cooper. It is
the **first measurement taken after the GCN-10 prune took effect at 2026-08-11
16:59:45 UTC** — every pre-existing WF6 evidence file measured the old 58-tool
Linear surface, because every 2026-08-11 cron fire (15 engagement-checker + 1
daily-work-brief) happened *before* that instant (§2.3). This file therefore
**supersedes** earlier WF6 verification evidence rather than repeating it: where
an older probe says PASS, that PASS was measured on a tool surface that no
longer exists. Verdicts: **criterion 1 PARTIAL** (agent+tool leg PASS, inbound
Slack leg never stimulated), **criterion 2 BLOCKED — never exercised**,
**criteria 3, 4, 5 PASS**. Two residuals, both with executable protocols (§9).
All times UTC unless a line says Paris/CEST.

## 1. What was measured, and what was not

| Criterion | Verdict | One-line basis |
|---|---|---|
| 1 — DM → GCN issue with label+priority, reply carries key/URL | **PARTIAL** | agent+tool leg PASS (GCN-28, §3); inbound Slack leg **never stimulated** (§4) — the DM was not sent |
| 2 — engagement-checker cycle: file / no re-file / close clears | **BLOCKED — NOT MEASURED** | two attempts, two different gates, neither a WF6 defect (§5) |
| 3 — daily-work-brief opens with priority-sorted board + Orca/Claude history | **PASS** (3a, 3b, 3c) | delivered brief ts `1786470753.300289`, 12/12 board rows real (§6) |
| 4 — cooper Claude/Orca session defaults to GCN | **PASS** | clean rerun → GCN-29; first run GCN-27 recorded as contaminated and **not scored** (§7) |
| 5 — everything committed; live-reload via `ActiveEnterTimestamp` | **PASS** | both repos clean and == origin/main; 8/8 skill mirrors byte-identical; timestamp unchanged (§8) |

**Read the difference between BLOCKED and FAIL.** Criterion 2 is *not* a
measured failure. Nothing about the push/no-re-file/pull mechanism was observed
to misbehave; the mechanism never ran. Criterion 1's inbound leg is likewise
not a failure — the stimulus (a DM typed by Gaetan) never existed, confirmed by
the operator after the fact. Anything not measured is written "not measured"
below and never softened into an implied pass.

## 2. System state at measurement

### 2.1 Versions and clocks

```
Hermes Agent v0.20.0 (2026.8.3)
Install directory: /home/gaetan/.hermes/hermes-agent
Python: 3.11.15
OpenAI SDK: 2.24.0
```

The VM's **system clock is UTC**; the Hermes config carries `timezone:
Europe/Paris`, so `hermes cron list` prints `+02:00` while the logs do not.

### 2.2 `agent.log` timestamps are UTC — established, not assumed

Two independent methods, both run this lane.

Simultaneous `date -u` / `date` against a fresh log line:

```
=== A1: clock readings ===
2026-08-11 17:23:58 UTC
2026-08-11 17:23:58 UTC
=== A1: tail agent.log ===
2026-08-11 17:23:32,360 INFO run_agent: Shared OpenAI client retired (cache_evict, tcp_shutdown=0, fd_release=deferre...
```

`date -u` and local `date` print the identical string — the VM's own local zone
is UTC — and the log line sits 26 s earlier. Cross-checked against three cron
completions, matching `hermes cron list`'s Paris output to the second:

| Job | `cron list` "Last run" (+02:00) | → UTC | `agent.log` completion line |
|---|---|---|---|
| Gaetan daily work brief | `2026-08-11T08:34:38.926554+02:00` | 06:34:38 | `2026-08-11 06:34:38,536 INFO cron.scheduler: Job 'Gaetan daily work brief' completed successfully` |
| Gaetan engagement checker | `2026-08-11T16:34:47.174128+02:00` | 14:34:47 | `2026-08-11 14:34:46,565 INFO cron.scheduler: Job 'Gaetan engagement checker' completed successfully` |
| …final pass | `2026-08-11T17:05:34.648078+02:00` | 15:05:34 | `2026-08-11 15:05:34,039 INFO cron.scheduler: Job 'Gaetan engagement checker final pass' completed successfully` |

**ANSWER: `~/.hermes/logs/*.log` timestamps are UTC.** Every log quotation in
this file is UTC. This mattered: without it, the 17:24 probe and the 19:24
Paris wall clock read as two different events.

### 2.3 The prune instant — and why it invalidates earlier WF6 evidence

Config write (read-only `stat`, no content read):

```
2026-08-11 16:58:38.075177373 +0000 /home/gaetan/.hermes/config.yaml
```

First post-edit registration — **the moment the prune took effect**:

```
2026-08-11 16:59:45,976 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): mcp__linear__list_com...
```

67 s after the config write, consistent with the documented ~30 s live-reload.
Last pre-prune registration, for contrast:

```
2026-08-11 07:45:05,644 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 58 tool(s): mcp__linear__get_atta...
```

Every 2026-08-11 cron fire of both WF6 jobs is hours **earlier** than
16:59:45:

- daily-work-brief, the only fire today: `06:30:55` start / `06:34:38` complete
  → **BEFORE**, by ~8.5 h.
- engagement-checker, **all 15 fires** 08:00 → 15:00 (`08:00:56, 08:30:56,
  09:00:57, 09:30:57, 10:00:58, 10:30:58, 11:00:59, 11:31:00, 12:00:00,
  12:30:01, 13:00:01, 13:30:02, 14:00:02, 14:30:03, 15:00:03`) → **BEFORE**,
  every single one; the latest (the `0 17 * * 1-5` final pass at 15:00:03) is
  still ~2 h before the prune.

⇒ **All pre-existing WF6 evidence measured the 58-tool world.** This lane is the
first post-prune measurement of every criterion it scores.

### 2.4 Registered tool counts — the enforcement surface

```
2026-08-11 17:24:08,545 INFO tools.mcp_tool: MCP server 'notion' (stdio): registered 24 tool(s): mcp__notion__API_get...
2026-08-11 17:24:08,754 INFO tools.mcp_tool: MCP server 'slack' (stdio): registered 20 tool(s): mcp__slack__channels_...
2026-08-11 17:24:10,595 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): mcp__linear__list_com...
2026-08-11 17:24:10,596 INFO tools.mcp_tool: MCP: registered 54 tool(s) from 3 server(s)
```

**notion 24 + slack 20 + linear 10 = 54 from 3 servers.** The Linear surface,
in full (from the untruncated line captured at 17:29:35):

```
mcp__linear__list_comments, mcp__linear__save_comment, mcp__linear__delete_comment,
mcp__linear__get_issue, mcp__linear__list_issues, mcp__linear__save_issue,
mcp__linear__list_issue_statuses, mcp__linear__list_issue_labels,
mcp__linear__list_teams, mcp__linear__list_users
```

Per `gcn10-linear-prune.md` §7, **registration is the enforcement layer**;
`hermes mcp test linear` (discovery) still reads 58 and is inconclusive by
design. `hermes mcp test` was not run and is not cited anywhere in this file.

### 2.5 No restart, across the whole lane

```
NRestarts=0
ActiveState=active
SubState=running
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

`ActiveEnterTimestamp` was read by **four independent agents at six separate
moments** spanning 17:24 → 18:22 UTC (baseline, criterion-1 probe ×2 at 17:31
and 17:32:51, criterion-2 probe before/after both cycles, criterion-3 probe at
17:47:32 and 18:00:28, criterion-2 retry before and after) — **identical every
time**. The 58→10 registration change therefore happened by **live-reload on the
same process**, not a bounce. Per the spec's own correction, `NRestarts=0` is
recorded as a secondary datum and is **not** cited as proof.

### 2.6 GCN-10 regression check — negative everywhere

Across every probe in this lane — 3 headless `hermes chat` sessions, 3 cron
sessions, 1 long manual cycle — there were **ZERO `tool-denied` /
`tool-not-found` / `unknown-tool` errors on any `mcp__linear__*` tool**. Grep
over `errors.log` for the whole window:

```
$ grep -E "^2026-08-11 1[678]:" ~/.hermes/logs/errors.log | grep -iE "denied|not found|unknown tool|not allowed|unauthorized|no such tool"
2026-08-11 16:14:40,763 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U09SS9Q1W82 in channel GQ07CQXT7
2026-08-11 16:15:00,666 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U09SS9Q1W82 in channel GQ07CQXT7
```

Two Slack allowlist rejects (see §4.3), nothing Linear. **No GCN-10 regression
observed anywhere in this lane.**

### 2.7 WF6 ticket state at measurement

T1→GCN-1, T2→GCN-2, T3→GCN-3, T4→GCN-4, T5→GCN-5, T6→GCN-6 — **all Done**,
closed 2026-08-10. T7→**GCN-7 In Progress** (`startedAt
2026-08-10T16:40:34.149Z`, `completedAt: null`), zero comments on the ticket;
all working state lives in this repo.

---

## 3. Criterion 1 — the agent + tool leg

> **Criterion, verbatim from the spec:** *"DM Tars 'create a ticket to test X' →
> issue appears in GCN with label + priority; Tars replies with key/URL."*
>
> Spec correction already on record (post-GCN-10/GCN-13): *"No agent can
> exercise this criterion's inbound leg. […] The autonomous substitute is
> `hermes chat -Q -q '<query>'` […] which proves the agent + tool leg only […]
> and leaves the Slack inbound leg unproven."*

### 3.1 The probe

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes chat -Q -q "create a ticket to test the WF6 end-to-end verification path (GCN-7 probe, 2026-08-11)"' 2>&1

session_id: 20260811_172408_865ddb
Created GCN-28: Test the WF6 end-to-end verification path (GCN-7 probe, 2026-08-11)
Todo · P3 Medium · test-check · assigned to Gaetan  
https://linear.app/mobile-club/issue/GCN-28
=== EXIT: 0 ===
```

Complete captured output, nothing elided. Exit 0, ~30 s wall (session start
17:24:07, turn end 17:24:38). `-Q -q` was confirmed against `hermes chat
--help` on the VM first — there is no `--oneshot` flag.

### 3.2 Read-back — `mcp__claude_ai_Linear__get_issue(id="GCN-28")`

```json
{"id":"GCN-28",
 "title":"Test the WF6 end-to-end verification path (GCN-7 probe, 2026-08-11)",
 "priority":{"value":3,"name":"Medium"},
 "url":"https://linear.app/mobile-club/issue/GCN-28",
 "createdAt":"2026-08-11T17:24:33.637Z","updatedAt":"2026-08-11T17:24:33.637Z",
 "completedAt":null,"canceledAt":null,
 "status":"Todo","statusType":"unstarted",
 "labels":["test-check"],
 "stateHistory":[{"state":{"id":"59ed732a-f242-4eba-926f-1c0d128fe83c","name":"Todo","type":"unstarted"},"startedAt":"2026-08-11T17:24:33.637Z","endedAt":null}],
 "createdBy":"Gaëtan Cathelain","createdById":"4951b192-e49c-4b7e-b491-58c89e66043c",
 "assignee":"Gaëtan Cathelain","assigneeId":"4951b192-e49c-4b7e-b491-58c89e66043c",
 "team":"Gaetan","teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e"}
```

### 3.3 Sub-criteria

**(a) Issue in team GCN — PASS.** `"team":"Gaetan"`,
`"teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e"`, key prefix `GCN-`. No other
team touched.

**(b) Label — PASS.** `"labels":["test-check"]`. The probe prompt named no
label; the agent chose the board's probe/verification workload label unprompted.

**(c) Priority — PASS, `3` / Medium.** `"priority":{"value":3,"name":"Medium"}`
— non-zero, i.e. explicitly set. Control: sibling issue GCN-27, created 35 s
earlier by an agent that passed no priority, reads `{"value":0,"name":"No
priority"}` — that is what unset looks like in this workspace, so `3` could not
have arrived by default.

**(d) Reply carries key and/or URL — PASS, both.** Exact substrings of stdout:
`Created GCN-28:` and `https://linear.app/mobile-club/issue/GCN-28`. The reply
also echoed `Todo · P3 Medium · test-check · assigned to Gaetan`, matching the
Linear record field-for-field.

### 3.4 Log evidence

```
2026-08-11 17:24:10,595 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): mcp__linear__list_comments, mcp__linear__save_comment, mcp__linear__delete_comment, mcp__linear__get_issue, mcp__linear__list_issues, mcp__linear__save_issue, mcp__linear__list_issue_statuses, mcp__linear__list_issue_labels, mcp__linear__list_teams, mcp__linear__list_users
2026-08-11 17:24:34,451 INFO [20260811_172408_865ddb] agent.tool_executor: tool mcp__linear__save_issue completed (1.44s, 1328 chars)
```

Session id matches the probe's stdout, so the attribution is exact.
`completed`, not `returned error`. The only tool error in the whole probe window
is unrelated to Linear and is a known headless artefact:

```
2026-08-11 17:24:16,103 WARNING [20260811_172408_865ddb] agent.tool_executor: Tool kanban_show returned error (0.00s): {"error": "task_id is required (or set HERMES_KANBAN_TASK in the env)"}
```

Whether a Slack-delivered turn hits the same `kanban_show` artefact: **not
measured.**

**VERDICT — criterion 1, agent+tool leg: PASS (a, b, c, d), post-prune.**

### 3.5 Disposition

GCN-28 was left in place, state Todo. See §12 for cleanup.

---

## 4. Criterion 1 — the inbound Slack leg: RESIDUAL, never stimulated

**This is not "attempted and failed". The stimulus never existed.**
Operator-confirmed after the fact: Gaetan pasted the test DM text to the
coordinator, then held off sending it after seeing GCN-28 appear on the board —
which was this lane's own probe (§3), not a response to him.

Three independent negatives were captured *before* that confirmation.

### 4.1 Slack side — polling and search

`D0BBYNM01BL` (the Gaetan DM channel) polled **12 times over 5 m 29 s**
(17:29:40 → 17:35:09). Newest message unchanged on every poll:

```
=== Message from Gaetan Cathelain <gaetan.cathelain@mobile.club> (U08BDJAMSRZ) at 2026-08-10 22:09:21 CEST ===
Message TS: 1786392561.474909
read ~/tars-probe.txt one more time and reply with its exact contents
```

That is **2026-08-10** — the DM channel had no traffic at all on 2026-08-11.
Slack-wide search:

```
mcp__claude_ai_Slack__slack_search_public_and_private
  query="create a ticket to test the WF6"  sort=timestamp  limit=20
→ # Search Results for: "create a ticket to test the WF6"
  No results found.
```

### 4.2 VM side — the decisive evidence

`gateway.log` records every inbound Slack message. The last three of the day,
verbatim:

```
2026-08-11 16:23:28,811 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C0BP2GZUFSR msg='Got it. Redraft the message for Oli, Jerem and Nans. You should be able to add S' reply_to_id=1786464083.177089 …
2026-08-11 16:24:45,543 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C0BP2GZUFSR msg="It's a female, update your draft" reply_to_id=1786464083.177089 …
2026-08-11 16:51:28,042 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C0BP2GZUFSR msg='<https://mobileclub-squad.slack.com/archives/C0BFQ5WFYTB/p1786442960691639?threa' reply_to_id=None reply_to_text=''
```

**Last inbound Slack event of any kind: 16:51:28 UTC**, and it was in
`C0BP2GZUFSR`, not the DM. Everything in `gateway.log` from 17:15 → 17:33 is
idle housekeeping (`Agent cache idle-TTL evict`). Session tags in `agent.log`
from 17:00 onward:

```
[20260811_170006_b18147]        platform=cli
[20260811_170057_96e9a0]        platform=cli
[20260811_172408_865ddb]        platform=cli
[cron_759e08c598e3_20260811_192932]  platform=cron
[SILENT]
```

**No `platform=slack` session exists after 17:00 UTC.**

### 4.3 Two POSITIVE partials — inbound ingestion is alive today

These are stated because "nothing arrived" is otherwise an argument from
silence:

1. **A real inbound was ingested normally at 16:51:28 UTC** in `C0BP2GZUFSR`
   (line quoted above) — the gateway's inbound path was working ~33 minutes
   before the probe window.
2. **The allowlist guard fired on real traffic today**, twice:

```
2026-08-11 16:14:40,763 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U09SS9Q1W82 in channel GQ07CQXT7
2026-08-11 16:15:00,666 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U09SS9Q1W82 in channel GQ07CQXT7
```

   A genuine third party hit the guard and was rejected. **No sender was
   simulated** — per `docs/facts.md`, the claude.ai Slack connector authenticates
   as Gaetan and its posts are dropped as bot-sender, so it can never serve as
   either a positive or a negative-sender stimulus.

### 4.4 The outbound Slack leg IS proven

Criterion 3's delivery (§6.2) is a live Tars→Slack post: ts
`1786470753.300289`, channel `C0BP2GZUFSR`, sender Tars `U0BBH85NAKH`, 3365
chars, server-side confirmed `cron.scheduler: Job 'e231e5faf180': delivered to
slack:C0BP2GZUFSR`.

**VERDICT — criterion 1, inbound Slack leg: RESIDUAL, not proven. The full loop
`authorized DM → agent → save_issue → reply in Slack` remains unexercised for
today's post-prune surface.** Protocol to close it: §9.1.

---

## 5. Criterion 2 — engagement-checker cycle: BLOCKED / NOT MEASURED

> **Criterion, verbatim:** *"One engagement-checker cycle: a detected loop lands
> as a GCN issue; its next cycle does NOT re-report the item it filed; closing
> the issue in Linear clears it from the queue (pull works)."*

| Sub-criterion | Verdict |
|---|---|
| 2a PUSH — a detected loop lands as a GCN issue | **BLOCKED — not exercised** |
| 2b NAG-LOOP GUARD — next cycle does not re-file | **BLOCKED — not exercised** |
| 2c PULL — closing in Linear clears the queue | **BLOCKED — not exercised** |
| (incidental) whitelist survivability under the checker | **PASS on registration, UNPROVEN on call** |

**None of 2a/2b/2c was measured. None was measured and failed.** Two attempts
hit two *different* gates, and neither gate is a WF6 defect.

### 5.1 Attempt 1 — `hermes cron run`, out of window

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes cron run 759e08c598e3'
Triggered job: Gaetan engagement checker final pass (759e08c598e3)
  Next run: 2026-08-12T17:00:00+02:00
  Ran now: succeeded.
EXIT=0
```

Run twice: agent turns 17:29:32→17:30:28 and 17:42:44→17:43:54 UTC (= 19:30 and
19:43 Paris). Both hit the skill's own guardrail, `engagement-checker` SKILL.md
v1.7.1 line 18, verbatim:

> All human times use `Europe/Paris`. **Scheduled runs are restricted to
> 10:00–17:00 on workdays.** Most runs must return exactly `[SILENT]`.

The clock read, and the model's own reasoning headers on the final message:

```
=== role= tool ts= 1786469418.97791
{"output": "2026-08-11T17:30:18+00:00\n2026-08-11T19:30:18+02:00\n2026-08-11\n2026", "exit_code": 0, "error": null}

=== id 16124 ts 1786469428.4134042
REASONING: **Enforcing time-based run restriction**
          **Finalizing silent exit due to schedule**
TOOL_CALLS: None
CONTENT: [SILENT]
```

```
2026-08-11 17:30:28,458 INFO cron.scheduler: Job 'Gaetan engagement checker final pass' completed successfully
2026-08-11 17:30:28,484 INFO cron.scheduler: Job '759e08c598e3': agent returned [SILENT] — skipping delivery
```

Cycle 2 identical (`REASONING: **Enforcing scheduled run time restrictions**` →
`[SILENT]`). ~55 s each, **before** collection, reconcile, filing or delivery.
No Slack post, no board change, and the queue file was never written:

```
$ ssh gaetan@192.168.0.9 'ls -la ~/.hermes/state/engagement-checker.json'
-rw-rw-r-- 1 gaetan gaetan 107964 Aug 11 15:05 /home/gaetan/.hermes/state/engagement-checker.json
```

mtime still 15:05 UTC (the 17:00 Paris cron run), size unchanged. The
`~/.hermes/state/engagement-checker.lock` directory — created only *"After the
workday gate passes"* — never appeared.

### 5.2 Attempt 2 — manual `hermes chat`, testing the "Scheduled" wording

A second agent read the whole 714-line SKILL.md before invoking anything, to
test the reading that line 18's gate binds only *scheduled* runs. **That reading
was confirmed correct.** Four reasons from the file's own text: the sentence
names its subject ("Scheduled runs"); the skill extends a gate to manual runs
**explicitly** where it means to, and did so exactly once, for a *different*
gate (line 52: *"The cron schedule itself is weekday-only; the gate also
protects manual runs and excludes holidays and leave."*); the 10:00–17:00 window
is implemented as two crontab expressions, not a runtime check; and the
authoritative run prompt's gate clause enumerates weekday / holiday / leave with
no time-of-day term.

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes chat -Q -q "Run the engagement-checker skill now as an explicit manual verification cycle. This is a direct operator-initiated request, not a scheduled cron run."'
```

The manual run was **not** schedule-gated — proof, mid-run at 18:18:58Z:

```
--- state mtime ---
2026-08-11 15:05:09.026128993 +0000 107964
/home/gaetan/.hermes/state/engagement-checker.lock
LOCK_HELD
```

It took the single-writer lock, advanced the Slack and email cursors, reconciled
4 items, flushed state and exited 0 after **~13.5 minutes** (18:08:49 →
~18:22), versus ~55 s for the gated cron runs.

**But it hit a second, undocumented gate.** The §4 Linear collector is a
`terminal` command, and non-interactive `hermes chat` has no TTY to answer the
approval prompt:

```
⚠ Approval: set -euo pipefail export LINEAR_CURSOR='2026-08-11T14:30:26Z' export LINEAR_RUN_END='2026-08-11T18:09:11Z…
  ⏱ Timeout — denying command
```

```
2026-08-11 18:16:15,069 WARNING [20260811_180849_14f453] agent.tool_executor: Tool terminal returned error (302.81s): {"output": "", "exit_code": -1, "error": "BLOCKED: Command timed out without user response. The user has NOT consented to this action. Do NOT retry this command, do NOT rephrase it, and do NOT attempt
```

Tars recorded the failure honestly in durable state:

```json
"last_failure_notice": {"linear": {"at": "2026-08-11T20:09:11+02:00", "message": "Linear delta collection could not complete because the read-only collector command was blocked by the execution approval gate."}}
```

**This gate does not fire under cron.** A cron session called Linear MCP
successfully the same evening:

```
2026-08-11 17:50:12,320 INFO [cron_e231e5faf180_20260811_194748] agent.tool_executor: tool mcp__linear__list_issues completed (0.59s, 3837 chars)
```

Cursor bookkeeping behaved exactly as the skill prescribes on a source failure:

```
slack   cursor 2026-08-11T17:00:32+02:00 -> 2026-08-11T20:09:11+02:00   ADVANCED
email   cursor 2026-08-11T17:00:32+02:00 -> 2026-08-11T20:09:11+02:00   ADVANCED
linear  cursor 2026-08-11T14:30:26Z -> 2026-08-11T14:30:26Z   HELD
```

**Zero `mcp__linear__*` calls in the entire manual cycle** — no `list_issues`
(§7a dedupe), no `get_issue` (§5 confirm), no `save_issue` (§7 create / §5
close). Four items did move to `done`/`waiting`, but **every `status_history`
reason string begins `slack:`**, none `linear:completed`/`linear:canceled` —
those are §5's *instruction* path following Gaetan's own Slack messages
(18:00:22 CEST *"Switch the ticket to cancelled"* → GCN-25; 18:13:15 CEST
*"…GCN-23 — I believe this was done, double check"*), **not** the Linear-state
pull that 2c tests.

### 5.3 What was NOT done, deliberately

**No SKILL.md edit was made or considered.** Working around the operator's own
time gate was explicitly out of bounds; the second agent was instructed to abort
if the skill's text forbade manual runs, and it verified the file was untouched
afterwards (`engagement-checker/SKILL.md` mtime `2026-08-11 15:31:33 UTC`, md5
`a117cf494ea53d18a0fd1ffbc1b2310b` — both predate the session). No config.yaml
edit, no cron schedule change.

**No issue was closed** to set up the pull test. The only candidates were live
human tickets (GCN-12 Verdict OAuth client; GCN-15 — Urgent GoCardless token
revocation, which Gaetan raised with Tars at ~18:09 CEST tonight; GCN-24
Yacine's PR). Flipping one for ~15 minutes to measure nothing was not a trade
worth making. **Nothing was consumed: the pull leg is fully re-runnable.**

### 5.4 What criterion 2 DID establish

The pruned surface registers cleanly in the cron path, on both post-prune
startups and on the manual session:

```
2026-08-11 17:29:35,440 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): mcp__linear__list_comments, mcp__linear__save_comment, mcp__linear__delete_comment, mcp__linear__get_issue, mcp__linear__list_issues, mcp__linear__save_issue, mcp__linear__list_issue_statuses, mcp__linear__list_issue_labels, mcp__linear__list_teams, mcp__linear__list_users
2026-08-11 17:29:35,441 INFO cron.scheduler: Job '759e08c598e3': 54 MCP tool(s) available
2026-08-11 17:42:46,907 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): …
2026-08-11 18:08:51,931 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): …
```

with **zero denials**. The three tools the checker's write paths need
(`save_issue`, `get_issue`, `list_issues`) are all in the whitelist.
⇒ **PASS on registration, UNPROVEN on call** for the checker path specifically.

One deviation worth recording even though it is not a criterion: a fresh item
`EC-EAF7BB` ("Review Vercel usage after the on-demand budget reached 100%",
score 80, `status=open`, `source != linear`, no `linear_issue`) met §7's filing
condition in the manual cycle and was **left unfiled** — Tars generalised
"Linear collection failed" into "do not touch Linear at all", although §7a lists
exactly three things that block filing and a §4 source failure is not one of
them. Whether this reproduces under cron is **not measured** (under cron §4
succeeds, so the branch may never be taken).

**VERDICT — criterion 2: BLOCKED. 2a, 2b, 2c all NOT MEASURED.** Protocol to
close it: §9.2.

---

## 6. Criterion 3 — daily-work-brief

> **Criterion, verbatim:** *"daily-work-brief 08:30 output (or manual run):
> opens with the GCN+assigned board, priority-sorted, and includes an
> Orca/Claude-history section."*
>
> Spec correction on record: v1.5.0 emits a mandatory labelled `*Claude/Orca:*`
> **line**, not a `###` section; *"the implementation satisfies the intent."*

### 6.1 The probe

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); date -u; ~/.local/bin/hermes cron run e231e5faf180; echo "EXIT=$?"; date -u'
Tue Aug 11 05:47:48 PM UTC 2026
  ⟳ compacting context…
  ⟳ compacting context…
  ⟳ compacting context…
Triggered job: Gaetan daily work brief (e231e5faf180)
  Next run: 2026-08-12T08:30:00+02:00
  Ran now: succeeded.
EXIT=0
Tue Aug 11 05:52:34 PM UTC 2026
```

First fire of this job **after** the prune. Session registered the pruned
surface at 17:47:51 (`registered 10 tool(s)`).

### 6.2 Delivery

```
2026-08-11 17:52:32,969 INFO [cron_e231e5faf180_20260811_194748] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=14/500 budget=13/500 tool_turns=10 last_msg_role=assistant response_len=3365 session=cron_e231e5faf180_20260811_194748
2026-08-11 17:52:33,081 INFO cron.scheduler: Job 'Gaetan daily work brief' completed successfully
2026-08-11 17:52:33,402 INFO cron.scheduler: Job 'e231e5faf180': delivered to slack:C0BP2GZUFSR
```

Channel `C0BP2GZUFSR` (`#gcn-tars-reporting`), sender Tars `U0BBH85NAKH`,
message ts **`1786470753.300289`** (2026-08-11 17:52:33 UTC = 19:52:33 CEST),
posted as a reply in the day's report thread, parent `1786430078.708099`.

### 6.3 Delivered text — the board block, verbatim

```
*Board* (priority-sorted)
P1 GCN-15 Revoke the old GoCardless E1 production token and verify rejection — Backlog
P1 MC-4227 Absence d'alertes Slack pour les Non Compliant Loop — In Progress
P1 NMC-601 Rotate the leaked GoCardless E1 production token (prod log leak follow-through) — In Progress
P2 GCN-7 WF6 E2E verification + docs/facts + status log updates — In Progress
P2 MC-4112 Datadog observability E2E for Verdict (ECS + Vercel + Cleaq distributed tracing) — Todo
P2 MC-4228 Bouton de facturation absent sur Loop — In Progress
P2 NMC-497 Block manual "stop invoicing" on Loop for non-tech operators (167 leaking cases) — In Progress
P2 NMC-498 Gate manual stop-invoicing: Hasura permission revoke + api-v2 resolver guard + admin UI — In Progress
P2 NMC-606 A low-confidence investigation automatically triggers an adversarial review pass — In Review
P2 NMC-615 SE never opened a single metarepo learnings PR despite the capability being Done — investigate + fix — In Review
P2 NMC-623 Pre-flight askback-need gate: the model reports evidence unconditionally, code decides to park — In Spec &amp; Design
P2 NMC-624 verdict.js can mint duplicate "Slack draft" fields — the prefill picks the wrong one — In Spec &amp; Design
+43 more
Coverage: 2 views, both complete (55 issues)
```

and the mandatory closing line of the `Since the last daily` block:

```
- *Claude/Orca:* `HelpTechs/MC-4226-2`, `MC-4227`, `mc-4228-rc`, `Tars/gcn7-wf6-e2e` and the NMC-649 lane moved; evidence includes `api#1509/#1510`, `metarepo#217` and open `metarepo#214`.
```

### 6.4 Sub-criteria

**3a — opens with the GCN + assigned board: PASS.** `*Board* (priority-sorted)`
is the first byte of the message; nothing precedes it. 12 rows (the script's
`MAX_ROWS = 12`) + computed `+43 more` + `Coverage: 2 views, both complete (55
issues)`. Three teams present — GCN 2 rows, MC 3, NMC 7 — i.e. both views (every
non-terminal GCN issue; every non-terminal issue assigned to Gaetan on any team)
are represented.

**3b — priority-sorted: PASS.** Rank sequence over the 12 rows is
`0,0,0,1,1,1,1,1,1,1,1,1` — monotonic non-decreasing, no inversion. Within each
band the GCN row leads (GCN-15 before MC-4227; GCN-7 before MC-4112), then team
key ascending (`MC` < `NMC`), then issue number ascending
(4112<4228; 497<498<606<615<623<624). That is exactly `linear_board.py`:

```python
PRIORITY_RANK = {1: 0, 2: 1, 3: 2, 4: 3, 0: 4}
...
def sort_key(row):
    return (
        PRIORITY_RANK.get(row["priority"], 4),
        0 if row["team_id"] == GCN_TEAM_ID else 1,
        row["team_key"],
        row["number"],
        row["identifier"],
    )
```

`12 rendered + 43 more = 55`, internally consistent with the `Coverage:` line.

**3c — Orca/Claude history: PASS.** The labelled `*Claude/Orca:*` line above is
present, naming four Orca worktrees plus a lane with evidence handles. §5 of the
skill mandates exactly that: *"one required labelled line closing the block […]
It is a named line, not an optional fold — the brief must be checkable for it."*
Evidence the Cooper collection actually ran: 20 `terminal` calls in the run,
several returning 16–24 KB (`17:48:58 tool terminal completed (0.68s, 24561
chars)`, `17:49:18 tool terminal completed (1.47s, 16719 chars)`). The command
strings themselves are **not measured** — `agent.log` records only tool name,
duration and output size.

### 6.5 Accuracy — the load-bearing check

On 2026-08-10 an LLM-rendered board fabricated 7 of 12 rows after context
compaction. **This run: zero fabrication.** Independent read of live Linear
through a different credential and code path (the claude.ai Linear connector,
from cooper):

| Delivered row | Exists? | Priority matches? | Status matches? |
|---|---|---|---|
| P1 GCN-15 — Backlog | yes | 1 | Backlog |
| P1 MC-4227 — In Progress | yes | 1 | In Progress |
| P1 NMC-601 — In Progress | yes | 1 | In Progress |
| P2 GCN-7 — In Progress | yes | 2 | In Progress |
| P2 MC-4112 — Todo | yes | 2 | Todo |
| P2 MC-4228 — In Progress | yes | 2 | In Progress |
| P2 NMC-497 — In Progress | yes | 2 | In Progress |
| P2 NMC-498 — In Progress | yes | 2 | In Progress |
| P2 NMC-606 — In Review | yes | 2 | In Review |
| P2 NMC-615 — In Review | yes | 2 | In Review |
| P2 NMC-623 — In Spec & Design | yes | 2 | In Spec & Design |
| P2 NMC-624 — In Spec & Design | yes | 2 | In Spec & Design |

**12/12 real, live, non-terminal, correct priority, correct status name,
matching title.** No hallucinated identifier, no invented title, no terminal
issue rendered as live.

**Omission check: none.** The live P1s assigned to Gaetan workspace-wide are
exactly three — `GCN-15`, `MC-4227`, `NMC-601` (`hasNextPage:false` on that
query) — and they occupy rows 1–3. Every other P1 that query returned is
`completed`/`canceled`.

The verifying agent recomputed the sort from the MCP data **before** reading the
delivered message and got an identical top-12 in identical order. And it held
through a mid-run context compaction — the exact condition that produced the
2026-08-10 fabrication:

```
17:51:18 agent.conversation_compression: context compression done: session=cron_… messages=52->34 rough_tokens=~103,869
17:51:56 tool terminal completed (0.86s, 1340 chars)
```

the script was re-run *after* the compaction (17:51:56), honouring the skill's
own rule. **The `linear_board.py` fix holds.**

### 6.6 Prune impact on criterion 3: none

Exactly **one** `mcp__linear__*` call in the whole run:

```
2026-08-11 17:50:12,320 INFO [cron_e231e5faf180_20260811_194748] agent.tool_executor: tool mcp__linear__list_issues completed (0.59s, 3837 chars)
```

In-whitelist, succeeded — it is the skill's `Linear tickets completed: N` native
read. Zero denials. Structurally, **the board itself cannot be affected by the
prune**: it comes from `scripts/linear_board.py` over raw HTTPS to
`api.linear.app/graphql` through the terminal tool, never through an
`mcp__linear__*` tool. Script confirmed present:

```
-rw-rw-r-- 1 gaetan gaetan 17626 Aug 10 20:36 linear_board.py
```

Four non-Linear tool errors occurred (a Slack MCP `channel not found` quirk, two
cron dangerous-command `BLOCKED` results, one `execute_code` `BLOCKED`); the
agent routed around all of them and still produced a correct board. None is a
Linear denial.

### 6.7 Not measured, stated plainly

- The `55 issues` total: the assigned read paginates and returned
  `hasNextPage:true`; ≥45 distinct live non-terminal issues were enumerated.
  45 ≤ 55 is consistent; "exactly 55" is **not measured**.
- The prose sections (`Since the last daily`, `Stats`, `Today`) — `PRs merged:
  17`, `Linear tickets completed: 25` — **not measured**, out of scope.
- Whether the single `list_issues` read was paged to `hasNextPage == false` as
  the skill requires — **not measured** (only one call was made).

**VERDICT — criterion 3: PASS on 3a, 3b, 3c, plus accuracy, post-prune.**

---

## 7. Criterion 4 — cooper Claude/Orca session defaults to GCN

> **Criterion, verbatim:** *"A ticket created from a Claude/Orca session on
> cooper defaults to GCN."*

### 7.1 The first run is recorded as CONTAMINATED and is NOT scored

The first probe filed **GCN-27** and returned PASS. It is not scored, and the
reason is a defect in **this lane's methodology**, not in the system: the
orchestrator's own prompt handed that agent the GCN team id
(`Linear team GCN id 81e7b769-2a46-4e2a-8db5-c165a7963b0e`), plus the criterion
framing and a prescribed title naming "default-team probe". The agent disclosed
all three contaminations itself:

> **Consequence: this run proves DELIVERY (§2, hard evidence) and proves the end
> state is correct (§4, hard evidence). It does NOT cleanly isolate the prose
> rule as the *cause*.**

It did establish one thing worth keeping — the rule text reaches a spawned
subagent's prompt on cooper, first-person, under an *"IMPORTANT: These
instructions OVERRIDE any default behavior"* preamble.

### 7.2 The clean rerun — GCN-29

A second agent was given a plain chore ("Rotate the Beszel agent token on the
monitoring host") with **no team, no GCN identifier and no criterion framing
anywhere in its prompt**. It filed:

```json
{"id":"GCN-29","title":"Rotate the Beszel agent token on the monitoring host",
 "priority":{"value":4,"name":"Low"},
 "url":"https://linear.app/mobile-club/issue/GCN-29",
 "createdAt":"2026-08-11T17:28:51.755Z",
 "status":"Backlog","statusType":"backlog","labels":[],
 "createdBy":"Gaëtan Cathelain","createdById":"4951b192-e49c-4b7e-b491-58c89e66043c",
 "team":"Gaetan","teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e"}
```

and cited its reason verbatim, unprompted:

> Linear tickets an agent creates default to team **GCN** (`Gaetan`, personal)
> unless another team is named — in a company repo, that repo's own context
> names the team.

Import chain: `~/.claude/CLAUDE.md:9` → `@~/dev/gaetan-metarepo/preferences/tooling.md`
— one hop, no conditionals, no per-project gating. It applied the rule's own
branch logic explicitly (the Tars repo is personal and names no Linear team, so
the company-repo carve-out does not fire), and resolved the key `GCN` to an id
by `list_teams` rather than assuming one. Nine teams were reachable — including
`Agents`, a plausible wrong attractor for an agent-filed ticket. Not chosen.

### 7.3 The mechanism — the honest finding

**Shipped at the policy layer, unenforced at the tooling layer.** From the
loaded `mcp__claude_ai_Linear__save_issue` schema, verbatim:

> "Create or update a Linear issue. […] **When creating, `title` and `team` are
> required.**"
> `"team": {"description": "Team name or ID (required when creating)", "type": "string"}`

So `team` is a required argument **with no default** — the value must come from
the model's context, i.e. from the prose rule. And equally: **nothing validates
it.** Any of the nine reachable teams would have been accepted silently. There
is no hook, no settings.json entry, no MCP wrapper, no env var carrying a Linear
`teamId`. The failure mode is silent, and it fails **outward**: a personal
ticket landing in `Mobile club` / `Cleaq` / `Next Mobiles` is colleague-visible.
Blast radius is a wrong-team ticket, not a data leak, but it is not zero.

**Not measured:** whether the default survives context compaction in a long
session; whether Tars-on-VM holds the same default (`tooling.md` does not reach
the VM — `~/.hermes/SOUL.md` has no team-default rule at all; the GCN default
lives only inside the two orchestration skills, hardcoded).

**VERDICT — criterion 4: PASS, on the clean rerun (GCN-29). GCN-27 not scored.**

---

## 8. Criterion 5 — commits and live-reload

> **Criterion, verbatim:** *"All changes committed: Tars repo (skills mirrors,
> this spec, status log), gaetan-metarepo (preference line). VM live-reload
> confirmed via `ActiveEnterTimestamp` unchanged — NOT `NRestarts`, which
> provably misses restarts."*

### 8.1 Tars repo

```
$ git status --porcelain=v1 -unormal
(empty output)

$ git fetch origin main
ok fetched (1 new refs)
HEAD:        26e4b91db49d6d67da3a8a0d1585a75a87e282fb
origin/main: 26e4b91db49d6d67da3a8a0d1585a75a87e282fb
git rev-list --left-right --count HEAD...origin/main => 0	0
```

Clean tree, **HEAD == origin/main exactly**, 0 ahead / 0 behind. The WF6
artifacts are inside that pushed HEAD:

- skill mirrors: `e4dda26` ("wf6 T5/T3/T4: Linear GCN integration via native
  mcp__linear__* tools") → … → `26e4b91` ("feat(vm): prune Linear MCP tool
  surface 58 -> 10 (GCN-10)");
- the spec: `f6be527` (draft) → `f599904` → `639fa0b` → `d0c37ef` ("wf6 spec:
  restart proof = ActiveEnterTimestamp, not NRestarts");
- `status/lane-a.md`: unbroken trail through `36d9c4f`, `5942086`, `c5c763a`.

### 8.2 gaetan-metarepo

```
$ git status --porcelain=v1 -unormal
(empty output)
HEAD:        3065f7db96192cc2838b3a34400b900f93be3851
origin/main: 3065f7db96192cc2838b3a34400b900f93be3851
git branch -r --contains 3065f7d => origin/main
git log origin/main..HEAD --oneline => (empty)
```

```
commit 3065f7db96192cc2838b3a34400b900f93be3851
Date:   Mon Aug 10 16:36:00 2026 +0000
    preferences/tooling: agent-created Linear tickets default to team GCN
 preferences/tooling.md | 1 +
 1 file changed, 1 insertion(+)
```

Clean, and the preference commit **is pushed** — not an unpushed-preference
finding.

### 8.3 Skill mirror integrity (GCN-13 convention)

Repo `skills/<rel>` vs live `~/.hermes/skills/<rel>`, md5 both sides:

| Relative path | md5 (identical both sides) |
|---|---|
| `delegate-to-cooper/SKILL.md` | `92505150cd08004f7049a147659cd9f5` |
| `email/himalaya/SKILL.md` | `e72577a7069aeeed814e3b9c979d2b18` |
| `helptech-duty-intake/SKILL.md` | `d9e2f232dd8d1341d914a66cf8ab8696` |
| `hermes-operations/hermes-orchestration/SKILL.md` | `e13a2c6af5ac5792c2dfec97da1be8c2` |
| `orchestration/daily-work-brief/SKILL.md` | `00654bacc7634c4650e3fa1fc254f771` |
| `orchestration/engagement-checker/SKILL.md` | `a117cf494ea53d18a0fd1ffbc1b2310b` |
| `orchestration/linear-ticketing/SKILL.md` | `0e5bf9968bc095ae37915a969108da45` |
| `orchestration/secure-delta-collectors/SKILL.md` | `7b3a6f869ce660946c275d8f7bf644bd` |

**8/8 byte-identical, zero mismatches, no file missing on either side.** The VM
carries 128 `SKILL.md` against the repo's 8 — expected and by design: the mirror
carries only skills Tars has self-edited and landed, not the built-in Hermes
library. Not flagged as drift.

### 8.4 Live-reload

```
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

Unchanged across the entire lane (§2.5), including across the 58→10
registration change. `NRestarts` not cited.

**VERDICT — criterion 5: PASS.**

---

## 9. Residuals — protocols a fresh session with ZERO context can execute

### 9.1 Residual 1 — criterion 1's inbound Slack DM loop

**What is missing:** today's post-prune full loop `Gaetan's authorized DM →
gateway ingest → agent → mcp__linear__save_issue → Tars replies in Slack with
key/URL`. Everything except the *inbound Slack* hop is already proven (§3 for
the agent+tool leg, §4.4 for outbound delivery).

**Only Gaetan can execute the stimulus.** No agent can: the claude.ai Slack
connector authenticates as Gaetan but its posts are dropped as bot-sender and
can never trigger Tars. Do **not** attempt to simulate a sender.

**Protocol:**

1. **Ask Gaetan to send, from his own Slack client, a DM to Tars** in
   `D0BBYNM01BL` — text of the form *"create a ticket to test the WF6 E2E path"*.
   Note the wall-clock minute he sends it.
2. Within the following 5 minutes, on the VM:
   ```
   ssh gaetan@192.168.0.9 'grep -n "inbound message: platform=slack" ~/.hermes/logs/gateway.log | tail -5'
   ```
   Expect a line `user=U08BDJAMSRZ chat=D0BBYNM01BL msg='create a ticket…'`
   with a UTC timestamp after the send.
3. Find the turn and its write:
   ```
   ssh gaetan@192.168.0.9 'grep -E "platform=slack" ~/.hermes/logs/agent.log | tail -3'
   ssh gaetan@192.168.0.9 'grep "mcp__linear__save_issue" ~/.hermes/logs/agent.log | tail -3'
   ```
   The `agent.turn_context` line must read `platform=slack` (contrast: this
   lane's probe read `platform=cli`). The `save_issue` completion timestamp must
   bracket the new issue's `createdAt`.
4. Read the reply **with `include_bots=true`** — see §11, `slack_read_channel`
   hides Tars' own messages otherwise. Assert the reply text contains the issue
   key and/or `https://linear.app/mobile-club/issue/GCN-<n>`.
5. Read the created issue back (`get_issue`) and score (a) `teamId ==
   81e7b769-2a46-4e2a-8db5-c165a7963b0e`, (b) `labels` non-empty, (c) `priority.value != 0`,
   (d) key/URL in the reply. All four → **PASS**.

**Do not touch:** GCN-15 (live Urgent, GoCardless token revocation), GCN-12,
GCN-24, GCN-7. **Close afterwards:** only the new probe issue this protocol
creates (Canceled is fine).

### 9.2 Residual 2 — criterion 2, in-window re-run

**Exact job id: `759e08c598e3`** ("Gaetan engagement checker final pass"), or
`62e8cd9db637` ("Gaetan engagement checker", `*/30 10-16 * * 1-5`) — either
runs the same `engagement-checker` skill.

**Exact time window:** a **workday**, **10:00–17:00 Europe/Paris = 08:00–15:00
UTC**. The natural **11:00 Paris = 09:00 UTC** fire is ideal — reading a natural
cron fire needs no invocation at all.

**Ready-made candidate, no setup needed:** item `EC-EAF7BB` ("Review Vercel
usage after the on-demand budget reached 100%", score 80) is already queued
`status=open`, `source != linear`, no `linear_issue` — it satisfies §7's filing
condition today.

**Protocol — read step 0 before typing anything:**

> **0. Use `hermes cron run <job_id>` — NOT `hermes chat -Q -q`.** The chat path
> is *proven* (§5.2) to trip the terminal TTY approval gate: `Tool terminal
> returned error (302.81s): {"output": "", "exit_code": -1, "error": "BLOCKED:
> Command timed out without user response…`. The §4 Linear collector is a
> `terminal` command and a non-interactive chat session has no TTY to approve
> it, so the collector never runs, the Linear cursor is HELD and the cycle makes
> **zero** `mcp__linear__*` calls — nothing criterion 2 tests is exercised.
> **That gate does not fire under cron** (a cron session called
> `mcp__linear__list_issues` successfully at 17:50:12 the same evening). Either
> `hermes cron run` in-window, or read the natural cron fire.

1. **Before:** snapshot the board (`list_issues`, team `GCN`, `limit=250`,
   assert `hasNextPage:false`), copy `~/.hermes/state/engagement-checker.json`,
   and record `wc -l ~/.hermes/logs/agent.log` and
   `systemctl --user show hermes-gateway.service -p ActiveEnterTimestamp`
   (expect `Mon 2026-08-10 20:00:38 UTC`; `export XDG_RUNTIME_DIR=/run/user/$(id -u)` first).
2. **Run 1, in-window** → score **2a**:
   ```
   ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes cron run 759e08c598e3'
   ```
   PASS requires, in `agent.log` for that cron session: `mcp__linear__list_issues`
   (the §7a dedupe scan) followed by `mcp__linear__save_issue completed`, and a
   new GCN issue whose description line 1 is `Source: <the item's stable id>`
   and line 2 `Evidence: …`.
3. **Between runs** → set up **2c**: close **GCN-12 or GCN-24** to **Done**
   (`statusType: completed`). **Never `Duplicate`** — see GCN-30, §11. **Never
   GCN-15** (live Urgent GoCardless token revocation Gaetan raised with Tars on
   2026-08-11), never GCN-7, never GCN-27/28/29/30/31.
4. **Run 2, still in-window** → score **2b** and **2c**.
   - 2b PASS: no *new* issue duplicating an item that already carries a
     `linear_issue` handle.
   - 2c PASS: the closed issue's queue item reaches `done` with a status-history
     reason of **`linear:completed`** (not `slack:`) and its
     `linear_issue.closed_at` gets set.
5. **Restore** the issue closed in step 3 to its prior state, then re-verify
   `ActiveEnterTimestamp` is still `Mon 2026-08-10 20:00:38 UTC`.

**Scoring rule that will otherwise mislead you:** score 2a/2b/2c **off the
`agent.log` and the board, never off whether a Slack message appears.** §6's
urgency scoring is clock-dependent (`+2 per elapsed Paris work hour capped at
+14`, `+15 from 15:45`), so an in-window run scores items **~29 points lower**
than an evening run against a threshold of 60. A legitimate `[SILENT]` delivery
is expected and is not a failure.

**Do not** edit `SKILL.md`, `config.yaml`, or the cron schedule to get around
either gate.

---

## 10. Attribution — who filed what (a question Gaetan actually asked)

Gaetan believed the engagement-checker filed GCN-28. **It did not.**

**Method note — read before the table.** Several writers share the *same* Linear
API key, so `createdBy` is `"Gaëtan Cathelain"` (`createdById
4951b192-e49c-4b7e-b491-58c89e66043c`) for **all** of them and cannot
disambiguate anything. The only sound discriminator is correlation against the
VM's `agent.log`: an issue written by the VM leaves an
`agent.tool_executor: tool mcp__linear__save_issue completed` line, session-tagged,
whose timestamp minus the printed duration brackets the Linear `createdAt`. An
issue written by a **cooper Claude session** through the claude.ai connector
leaves **no VM log entry at all** — that absence *is* the discriminator.

Ground truth: **after 17:00 UTC there is exactly ONE `save_issue` execution on
the entire VM**, at 17:24:34,451.

| Key | createdAt (UTC) | VM correlate | Therefore |
|---|---|---|---|
| GCN-27 | 17:23:58.604Z | **none** | cooper Claude session — this lane's contaminated criterion-4 probe |
| GCN-28 | 17:24:33.637Z | `20260811_172408_865ddb`, `save_issue completed (1.44s)` at 17:24:34.451 | this lane's headless `hermes chat -Q` probe |
| GCN-29 | 17:28:51.755Z | **none** | cooper Claude session — this lane's clean criterion-4 rerun |
| GCN-30 | 18:08:40.395Z | **none** | cooper Claude session — this lane's side-finding ticket |
| GCN-31 | 18:09:07.103Z | **none** | cooper Claude session — this lane's side-finding ticket |

**GCN-28 arithmetic.** Log line = tool *completion* at 17:24:34.451, printed
duration 1.44 s ⇒ call start ≈ 17:24:33.01. Linear `createdAt` 17:24:33.637Z
falls inside `[17:24:33.01, 17:24:34.451]`. Unique, and the only candidate.
Session origin:

```
2026-08-11 17:24:11,499 INFO [20260811_172408_865ddb] agent.turn_context: conversation turn: session=20260811_172408_865ddb model=gpt-5.6-sol provider=openai-codex platform=cli history=0 msg='create a ticket to test the WF6 end-to-end verification path (GCN-7 probe, 2026-...'
```

`platform=cli` ⇒ `hermes chat -Q`. Not `platform=slack`, not `platform=cron`.

**Where the checker actually was.** Its `0 17 * * 1-5` job fired 15:00:03 UTC
and made a single `save_issue`:

```
2026-08-11 15:00:04,413 INFO [cron_759e08c598e3_20260811_170003] agent.turn_context: … platform=cron …
2026-08-11 15:04:13,797 INFO [cron_759e08c598e3_20260811_170003] agent.tool_executor: tool mcp__linear__save_issue completed (1.68s, 1367 chars)
```

which created **GCN-25** (`createdAt 15:04:13.277Z`) — **2 h 20 m before**
GCN-28. The `*/30` job `62e8cd9db637` last ran 14:34:47 UTC. The manual checker
re-run at 17:29:32 — **4 m 58 s after GCN-28 already existed** — made **zero**
`save_issue` calls.

---

## 11. Side findings — filed, and NOT WF6 pass criteria

- **GCN-30** (`fix`, High) — <https://linear.app/mobile-club/issue/GCN-30> —
  **engagement-checker §5 misses Linear's `duplicate` status type.** The
  reconciliation treats an item as closed only on `state.type == "completed"` or
  `"canceled"`. Marking an issue **Duplicate** is a way Gaetan legitimately
  clears one, but the checker does not see it as cleared: the parent item stays
  `open`, keeps being scored and reminded about every 30 minutes — **the exact
  nag loop the guard exists to prevent**, reached by a status he actually uses.
  Live instance: **GCN-25** (`statusType: "duplicate"`, `completedAt: null`,
  `canceledAt: 2026-08-11T15:10:41.624Z`), itself filed by the checker at
  15:04:13 UTC. The same two-value test is **duplicated** in §5 step 1 (the
  "skip the write when already terminal" guard), so a fix must land in **both**
  places.

- **GCN-31** (`investigation`, Medium) — <https://linear.app/mobile-club/issue/GCN-31>
  — **collector-path coverage risk.** The cron dangerous-command guard blocked
  `python3 -c` and heredoc terminal shapes during the checker cycles (17:30:08,
  17:43:32, 17:43:49 ×2) and three times during the work-brief run (terminal
  heredoc 17:48:27, terminal `-c` 17:50:24, `execute_code` 17:51:28); agents
  routed around all of them. It is intermittent **by command shape**, and §4
  mandates the collector run through the terminal tool. Separately,
  `google-workspace setup.py --check` fails on a uv/pip install error
  (`No module named pip` / externally-managed interpreter), so **email-sourced
  loops may be silently undetected** — the dangerous failure mode, where the
  checker looks healthy and reports nothing.

- **GCN-32** (`investigation`, priority 3/Medium, Backlog, created
  `2026-08-11T18:34:14.927Z`) — <https://linear.app/mobile-club/issue/GCN-32> —
  **engagement-checker records carry invented provenance.** Findings F1 and F2
  filed as **one failure family: right-ish behaviour, fabricated record.** Both
  from the same manual cycle, session `20260811_180849_14f453`.
  - **F1 — a policy citation that does not exist.** Tars justified correct
    silence with `No reminders were sent because the cycle ran outside the
    09:00–19:00 window.` That window is in **neither** `~/.hermes/SOUL.md`,
    `~/.hermes/config.yaml` **nor** the SKILL.md: grepping `09:00|19:00` across
    all three returns exactly one hit, and it is an unrelated French relative-time
    convention —
    `…/skills/orchestration/engagement-checker/SKILL.md:548:- ce soir → 19:00`.
    The real gate reads *"Scheduled runs are restricted to 10:00–17:00 on
    workdays."* Behaviour right (it did not nag at 20:09 Paris), citation
    invented.
  - **F2 — a state record with no corresponding action.** The state file got
    `+ slack:C09H8CHTHMZ:1786460726.573869 short=EC-D43FE6 status=done
    linear=GCN-26` on a turn that made **zero** `mcp__linear__*` calls — no
    `list_issues` (§7a dedupe), no `get_issue` (§5 confirm), no `save_issue`
    (§7 create). §7c allows exactly two write paths for that field, a confirmed
    create or a §7a adoption; neither happened, so the id is an unsourced
    assertion persisted into durable state. It is correct by luck — GCN-26 really
    is that commitment — inferred from conversation, not from a response payload.
  - **Why it matters, plainly.** Today F1 only excused a no-op and F2 only wrote
    a plausible-looking line into durable state; neither caused harm. The risk is
    that the audit trail is unreliable exactly when it is needed: the day an
    invented justification excuses a *wrong* action, or a phantom `linear_issue`
    id makes reconciliation skip a real item, the record will look plausible and
    be wrong. A confident wrong record stops the check that a missing record
    would have prompted.

- **PROBE GOTCHA, high value.** `slack_read_channel` **hides bot messages**
  unless `include_bots=true`. Tars' own deliveries are invisible without it —
  `slack_read_channel(C0BP2GZUFSR, oldest=1786470000)` returned *zero* messages
  while the brief was already posted. A future probe reading that channel with
  `slack_read_channel` alone will wrongly conclude "Tars delivered nothing".
  This burned time in this lane. (Note also: the `slack_read_channel` schema
  exposed to one agent this lane had **no** `include_bots` parameter at all —
  use `slack_search_public_and_private(..., include_bots=true)`.)

### 11.1 GCN-13's skill-mirror fix, confirmed in the wild (cross-lane)

Not a WF6 criterion, and not this lane's subject — which is precisely what makes
it worth its own section.

During the manual engagement-checker cycle (session `20260811_180849_14f453`),
**unsupervised and mid-cycle**, Tars hit `error: unrecognized subcommand
'folder'` at 18:10:15Z, called `skill_manage` at 18:19:56Z to rewrite
`~/.hermes/skills/email/himalaya/SKILL.md` for Himalaya v2's `folder`→`mailbox`
rename, and landed it per SOUL rule 2 — branch → PR → squash-merge, same turn,
~2 minutes after the edit:

```
PR56 state=MERGED merged=2026-08-11T18:21:52Z title=himalaya skill: document v2 mailbox command
files: [{"path":"skills/email/himalaya/SKILL.md","additions":12,"deletions":2,"changeType":"MODIFIED"}]
```

Both halves of the GCN-13 fix hold: the mirror path is the **correct live
relative path** `skills/email/himalaya/SKILL.md`, category dir included, not the
old flat form; and the diff is **+12/−2 on a 315-line file — non-empty**. The
pre-GCN-13 flat-path recipe merged **six EMPTY `SKILL.md` files**; "non-empty,
at the correct relative path" is exactly the failure that rules out.

**Why this outranks the in-lane verification: the lane that wrote a fix cannot
be the only witness to it.** GCN-13's own lane verified the recipe by running
it deliberately, knowing what it was testing. This sighting is **cross-lane and
incidental** — a different lane, a different ticket, a trigger (a Himalaya CLI
error) with no relation to GCN-13, no operator watching the merge, and nobody
setting up for the observation. That is the only kind of evidence that shows the
recipe works when nobody is checking it.

Bounding it honestly: **one** observation, on **one** skill file, on the same
day the fix landed. It is confirmation, not a regression suite.

---

## 12. Cleanup record

This lane's **disposable probe artifacts**, all three closed on 2026-08-11. An
explanatory comment was landed on each ticket **before** its state change, so
the record reads correctly from the ticket alone:

| Key | Created | What it is | Disposition |
|---|---|---|---|
| **GCN-27** | 17:23:58.604Z | contaminated criterion-4 probe (not scored, §7.1) | **Canceled** — comment 18:32:22.503Z, `canceledAt 2026-08-11T18:32:29.187Z`, `status "Canceled"` / `statusType "canceled"` |
| **GCN-28** | 17:24:33.637Z | criterion-1 agent+tool-leg probe (§3) | **Canceled** — comment 18:32:23.519Z, `canceledAt 2026-08-11T18:32:30.376Z`, `status "Canceled"` / `statusType "canceled"` |
| **GCN-29** | 17:28:51.755Z | clean criterion-4 rerun (§7.2) | **Canceled** — comment 18:32:24.955Z, `canceledAt 2026-08-11T18:32:31.467Z`, `status "Canceled"` / `statusType "canceled"` |

**Canceled, deliberately not Done.** `statusType: canceled`, never `completed`:
these are test artifacts, not completed work, and closing them as Done would put
three fake deliveries into the board's own history — and into every count that
reads it.

**GCN-29's body is not a real chore.** Its "Rotate the Beszel agent token on the
monitoring host" text was **invented as neutral probe content** — deliberately
plausible so the agent would file it without a team hint — and is **NOT** a
tracked task. Its comment says so verbatim, so nobody later mistakes it for a
dropped backlog item: *"The Beszel token-rotation content was invented as neutral
probe text and is NOT a real tracked chore; if Gaetan wants it tracked, it needs
filing on its own merits."*

**Do NOT touch:** **GCN-30**, **GCN-31** and **GCN-32** — real side-finding
tickets (§11), not probe artifacts. **GCN-15** (live Urgent, GoCardless token
revocation), **GCN-12**, **GCN-24** (live human commitments), **GCN-7** (this
ticket).

**Verified untouched by the cleanup — no write call was issued against any of
them:** GCN-7, GCN-12, GCN-15, GCN-16…GCN-26, GCN-30, GCN-31.

---

## 13. For `status/lane-a.md` (orchestrator to land — single-writer)

Ready to paste as the newest entry at the top of the `## Log` section. This file
does not edit `lane-a.md`.

```markdown
- 2026-08-11 ~18:25Z — **WF6 T7/GCN-7 §Verification pass — first measurement
  AFTER the GCN-10 prune; supersedes earlier WF6 evidence.** Every 2026-08-11
  cron fire (15 checker + 1 brief) predates the prune instant 16:59:45Z, so all
  prior WF6 PASSes measured the old 58-tool surface. Established this lane:
  agent.log timestamps are **UTC** (simultaneous `date -u`/`date` + 3 cron
  cross-checks). Counts now notion 24 / slack 20 / linear 10 = 54 from 3
  servers; `ActiveEnterTimestamp` unchanged at `Mon 2026-08-10 20:00:38 UTC`
  across the whole lane, measured by 4 agents — the 58→10 change went live by
  reload, no restart. **ZERO `mcp__linear__*` denials anywhere: no GCN-10
  regression.** Verdicts (`status/probes/gcn7-wf6-e2e.md`): check 1 **PARTIAL**
  — agent+tool leg PASS post-prune (GCN-28: team GCN, label `test-check`, P3,
  reply carried key+URL, `save_issue completed` 17:24:34.451Z), inbound Slack
  leg **never stimulated** (Gaetan held off after seeing GCN-28; `gateway.log`
  inbound stops 16:51:28Z, no `platform=slack` session after 17:00Z — 3
  independent negatives). check 2 **BLOCKED, NOT MEASURED** — 2a/2b/2c
  unexercised: `cron run` at 19:30/19:43 Paris hit the skill's own
  10:00–17:00 gate (`[SILENT]` in ~55 s, queue byte-identical), and manual
  `hermes chat` cleared that gate but died on a second, undocumented one (§4
  collector is a terminal command; no TTY → `BLOCKED … timed out` after 302.81 s;
  does NOT fire under cron). No SKILL.md edit, no issue closed — fully
  re-runnable. check 3 **PASS** (3a/3b/3c) — brief delivered ts
  1786470753.300289, 12/12 board rows real vs live Linear, no live P1 omitted,
  held through a mid-run compaction. check 4 **PASS** on a clean rerun (GCN-29,
  no team/id/framing in the prompt); the first run (GCN-27) is recorded as
  **contaminated by this lane's own prompt and not scored**. check 5 **PASS** —
  both repos clean and == origin/main (`26e4b91` / `3065f7d`), 8/8 skill
  mirrors byte-identical. Attribution answered: the engagement-checker did NOT
  file GCN-28 (its only save_issue was 15:04:13Z → GCN-25). Side findings
  filed: **GCN-30** (§5 misses `statusType=duplicate` → real nag-loop hole,
  live instance GCN-25), **GCN-31** (collector-path coverage risk) and
  **GCN-32** (engagement-checker records carry invented provenance: a cited
  "09:00–19:00 window" that exists in no file, and `linear_issue: GCN-26`
  written into durable state on a turn with zero Linear calls). Probe artifacts
  GCN-27/28/29 all **Canceled** 18:32:29–18:32:31Z, comment first on each. Two
  residuals with executable protocols in the probe file: Gaetan's literal DM,
  and an in-window (`08:00–15:00Z`) `hermes cron run 759e08c598e3` for check 2.
```
