# GCN-30 + GCN-32 — engagement-checker SKILL.md fix, split in two stages (2026-08-11)

## 1. What this is, and the two-stage split

**What this is.** The single evidence file for the GCN-30 / GCN-32 lane. It carries the
re-verification of both tickets from primary state, the root cause, the Linear
`statusType` audit, **stage 1 exactly as applied tonight**, the complete staged stage-2
edit with per-hunk rationale, the adversarial review results, a copy-paste post-apply
acceptance probe, the stage-2 apply protocol and both rollbacks. A coordinator can apply
stage 2 — or roll stage 1 back — from this file alone.

**The single diff was SPLIT on Gaetan's ruling into two stages.**

| stage | content | status | who applies |
|---|---|---|---|
| **stage 1 — plumbing** | §4's collector invocation shape + the interpreter on line 139 + version bump | **APPLIED to the live VM and pushed, 2026-08-11 (commit `d5a051d`)** | this lane, tonight |
| **stage 2 — semantics** | the reconcile terminal set, the close-back backstop, `linear_issue` provenance, the §8 citation rule | **STAGED ONLY. Not on the VM, not in `skills/`.** | the coordinator, **after** the WF6 R2 run |

**Why the split exists.** §4's wording and the interpreter are what make the collector
**runnable from cron at all** — plumbing. Under `approvals.cron_mode: deny` the previous
shape could not run: the collection silently did not happen and the run went `[SILENT]`
looking healthy (that is the operational cause in section 3.4, and it is what has held the
Linear cursor all day). Fixing it before tomorrow's run is what gives R2 its best chance of
a collector that actually collects — expected, not measured, and section 11.1 says why: no
run has yet exercised the new shape. The **reconcile semantics are the thing R2 measures**; changing them the night
before would change the behaviour under measurement, so they wait.

**The lane's one rule now applies to stage 2 only: STAGED TONIGHT, APPLIED BY THE
COORDINATOR AFTER THE WF6 R2 RUN.** Stage 2 exists only at

```
status/staged/gcn30-32-engagement-checker-SKILL.md
```

and must be moved into place by the protocol in section 9, never by re-typing it. Nothing
of stage 2 was written to the VM, to Linear, or to `skills/`.

### 1.1 Version scheme

Decided by this lane, ratified by the coordinator. **The live `version:` always names which
stages are in**, which is the whole point of having one:

| version | stage | digit moved | why |
|---|---|---|---|
| `1.7.1` | pre-fix base | — | the file as it stood before this lane |
| `1.7.2` | **stage 1** | patch | plumbing only: how the collector is invoked and by which interpreter. No decision path in the skill changes; nothing the skill *decides* moves. |
| `1.8.0` | **stage 2** | minor | behavioural: the terminal set, the close-back backstop, a new provenance rule and a new citation rule. The skill decides differently after it. |

Rationale for the patch/minor call: across the file's whole git history the minor digit
moved for every deliberate behavioural bump (workday gate, collector-contract bump, GraphQL
switch, native Linear-tool rewrite, Slack self-id pin) and the patch digit moved exactly
once, for a metadata-field addition (`9731ac5`, 1.7.0→1.7.1). Stage 1 is that second kind;
stage 2 is the first.

If a reader ever finds the live file at `1.7.2`, stage 2 is not applied. At `1.8.0`, it is.

### 1.2 Artefacts and their hashes

| file | md5 | lines | bytes |
|---|---|---|---|
| pre-fix base (`d5a051d^`, and every backup named below) | `a117cf494ea53d18a0fd1ffbc1b2310b` | 714 | 69654 |
| **stage 1** — VM live, repo `skills/orchestration/engagement-checker/SKILL.md`, `origin/main` | **`1caff77d5dd4597756dc0e5b4a54a39c`** | **716** | 71567 |
| **stage 2 — staged** `status/staged/gcn30-32-engagement-checker-SKILL.md` | **`d0adb8917fa7d793465bbf9a96b6bb87`** | **718** | 77072 |

**Do not confuse the two 716s.** 716 lines is *stage 1*. The staged stage-2 file is **718**
lines. An earlier staged file (`cb33422c0e8462a74e55c98c73f7b488`, 716 lines) was cut from
the pre-fix base and is **superseded and dead** — see section 6.1; if that hash appears
anywhere outside this sentence and section 6.1, the document is stale.

If the coordinator edits the staged file before applying, recompute the md5 and use the new
value everywhere in sections 8 and 9.

---

## 2. Re-verification of both tickets from primary state

All measurement was read-only: `ssh gaetan@192.168.0.9` with `cat`/`stat`/`md5sum`/`grep`,
Python's stdlib `sqlite3` on `mode=ro` URIs (the VM has no `sqlite3` CLI), and Linear MCP
reads (`get_issue`, `list_issues`, `list_issue_statuses`, `get_team`, `list_comments`).
VM clock is UTC; JSON state and cron records carry `+02:00` Paris offsets. Every
conversion below is explicit. Line numbers in this section are the **pre-fix base**
(714 lines); stage 1 shifted everything from base line 149 onward by +2, so base 519/520/531
are live lines 521/522/533 today. The quoted text itself is unchanged — stage 1 does not
touch §5.

### 2.1 GCN-30 — "statusType=duplicate is not treated as closed" — **CONFIRMED**

**Claim A: the terminal test recognises exactly two `state.type` values.** CONFIRMED,
string-for-string. Base line 520 and base line 531 (live lines 522 and 533), verbatim:

```
- A routed filed issue reporting `state.type == "completed"`: parent item `status = "done"`, status-history reason `linear:completed`. `state.type == "canceled"`: `status = "dismissed"`, reason `linear:canceled`. Closing the issue in Linear is how Gaetan clears the item. `state.type` is the **collector's** raw GraphQL shape and is correct here; a native read has no `state` object at all — it carries flat `status` and `statusType` (`linear-ticketing` §9).
```
```
1. **Skip the write when the issue is already terminal.** If this run's routed candidate reports `state.type` `completed` or `canceled` — that is how the item reached `done`/`dismissed` in the first place — set `linear_issue.closed_at` from that observation and make no call. Writing Done to an already-Done issue only bumps `updatedAt` and pulls it back into the next window for nothing.
```

`duplicate` appears in neither branch.

**Claim B: GCN-25 really is `statusType: "duplicate"`.** CONFIRMED. Raw
`mcp__linear__get_issue(id="GCN-25")`:

```json
{
  "id": "GCN-25",
  "title": "Check the missing Non Compliant Slack alerts and report back",
  "status": "Duplicate",
  "statusType": "duplicate",
  "canceledAt": "2026-08-11T15:10:41.624Z",
  "completedAt": null,
  "startedAt": null,
  "stateHistory": [
    {"state": {"id":"59ed732a-f242-4eba-926f-1c0d128fe83c","name":"Todo","type":"unstarted"},
     "startedAt":"2026-08-11T15:04:13.277Z","endedAt":"2026-08-11T15:10:41.449Z"},
    {"state": {"id":"0ea8bdc6-215f-48d3-b228-779422c6b03e","name":"Duplicate","type":"duplicate"},
     "startedAt":"2026-08-11T15:10:41.449Z","endedAt":null}
  ]
}
```

**Claim C: the parent item is stuck `open` right now.** **REFUTED as stated, CONFIRMED as a
past fact.** The live item (`/home/gaetan/.hermes/state/engagement-checker.json`, mtime
`2026-08-11 18:22:35 +0000`, md5 `f9b7dc639e38a26c5d3aa8facbd62b5e`) reads:

```json
{
  "id": "slack:CC397R0HY:1786458544.180249",
  "short_id": "EC-2B1E44",
  "title": "Check the missing Non Compliant Slack alerts and report back",
  "status": "done",
  "status_history": [
    {"at":"2026-08-11T17:00:32+02:00","status":"open",
     "reason":"Fresh explicit promise; retained and filed, but reminder suppressed while Gaetan is actively investigating"},
    {"at":"2026-08-11T18:10:01+02:00","status":"done",
     "reason":"slack: Gaetan marked the engagement-checker item duplicate of the Help Tech investigation and closed the request"}
  ],
  "linear_issue": {"id": "GCN-25", "priority": 3}
}
```

The prior snapshot `engagement-checker.json.bak-manual-20260811T180911Z` (taken at run
start, 18:09:11Z) carries the **same item with one status-history entry, `status: "open"`**
— i.e. the item genuinely sat `open` for **2 h 58 min** after GCN-25 flipped to Duplicate at
15:10:41Z (15:10:41Z → the backup's 18:09:11**Z** stamp; an earlier draft read that stamp as
Paris local and said "58 minutes"). The bug fired in a live run. It stopped being visible only because Gaetan
*also* said "duplicate" in Slack and the Slack-explicit-completion branch closed the item
(`reason` prefixed `slack:`, not `linear:*`). `linear_issue` carries **no `closed_at`** —
proof the Linear-terminal branch never ran for it. See section 3: that rescue is exactly what
arms the wrong write.

**Claim D: the ticket's stated queue line `status= open`.** PARTLY — true when the ticket
was filed, superseded by the 18:10:01+02:00 Slack close. The gap it names is unpatched.

**Ticket verdict: CONFIRMED**, with claim C re-stated as "was open, now `done` via an
unrelated branch, with the hole intact and a wrong write armed."

### 2.2 GCN-32 F1 — "a policy citation that does not exist" — **CONFIRMED**

Search, verbatim command:

```
grep -rniE "09:00|19:00|9am|7pm|quiet hour|working hour|business hour" \
  ~/.hermes/skills/ ~/.hermes/SOUL.md ~/.hermes/config.yaml
grep -niE "<same pattern>" ~/.hermes/config.yaml.bak* ~/.hermes/SOUL.md.bak*
```

7 raw hits, 5 unique files, all under `skills/**`. **Zero in `SOUL.md`, zero in
`config.yaml`, zero in their backups (grep exit 1).** None of the 7 describes a delivery
gate: `engagement-checker/SKILL.md:548` is `- ce soir → 19:00`, a French relative-time
mapping for due-date inference; `apple-reminders/SKILL.md:68` is a reminder due-date
example; the rest are cron-syntax examples (`"0 9 * * *"`, `"every monday 9am"`).

The invented sentence itself is **not** in `agent.log` — that log truncates prompts and
records only `response_len` for turns. It was recovered from the primary conversation
store `~/.hermes/lcm.db` (`messages`, `store_id=19718`,
`session_id=20260811_180849_14f453`, `role=assistant`,
`timestamp=1786472565.1246026` = 2026-08-11T18:22:45.12 UTC, `len(content)==1020`,
matching `agent.log`'s `Turn ended … response_len=1020`). Verbatim, the load-bearing line
inside that message:

```
- No reminders were sent because the cycle ran outside the 09:00–19:00 window.
```

The live restriction, line 18 (unmoved by stage 1), verbatim:

```
All human times use `Europe/Paris`. Scheduled runs are restricted to 10:00–17:00 on workdays. Most runs must return exactly `[SILENT]`.
```

`09:00–19:00` is a near-recall of that line with both bounds wrong. **F1 CONFIRMED.**

### 2.3 GCN-32 F2 — "a state record with no corresponding action" — **CONFIRMED**

Session `20260811_180849_14f453`, two independent counts:

```
grep "20260811_180849_14f453" ~/.hermes/logs/agent.log | grep -c "agent.tool_executor"   → 48
grep "20260811_180849_14f453" ~/.hermes/logs/agent.log | grep -niE "linear"               → exit 1 (no match)

SELECT COUNT(*), role FROM messages WHERE session_id='20260811_180849_14f453' GROUP BY role
   → [(53,'assistant'), (126,'tool'), (1,'user')]
SELECT COUNT(*) FROM messages WHERE session_id=? AND tool_name LIKE '%linear%'
   → (0,)
```

**Zero Linear MCP calls, 0 of 48 (agent.log) and 0 of 126 (lcm.db).** The state write, from
`lcm.db` `store_id=19702`, the `execute_code` call whose arguments contain "GCN-26":

```python
items[commit_id]={
  ...
  'linear_issue':{'id':'GCN-26','priority':3,'closed_at':'2026-08-11T18:16:00+02:00'},
  ...
}
```

A hardcoded Python literal, written straight to `~/.hermes/state/engagement-checker.json`
(`grep -n "GCN-26" …json` → `3221: "id": "GCN-26",`; absent from the 18:09:11Z backup, so
the write is this session's). Two lines above it, the same script sets:

```python
d.setdefault('last_failure_notice',{})['linear'] = {'message':'Linear delta collection could not
  complete because the read-only collector command was blocked by the execution approval gate.'}
```

The run recorded Linear as unreachable and fabricated a Linear id in the same script.
**F2 CONFIRMED.**

---

## 3. ROOT CAUSE

### 3.1 The mechanism, plainly

**The Linear→item reconcile in §5 is event-driven only.** It fires exclusively for a §4
candidate inside the run's cursor window that routes to a parent item. Greps over all 714
base lines for `sweep|periodic|re-read|staleness|every item|all items|linear_issue|get_issue|…`
returned 29 hits, all inspected: the only 30-day sweep is explicitly report-only and never
reads Linear (base line 91, unmoved by stage 1), and the sole `get_issue` in §5 (base line
532) is reached only from inside the close-back and tests `teamId` alone. **There is no
sweep, no periodic re-read, no staleness re-check.**

`duplicate` is a distinct workflow-state type, not an alias of `canceled` (section 4 below). The
two terminal tests in §5 do not include it. That is the ticketed nag-loop hole. **Stage 1
does not touch any of this** — it is exactly what stage 2 changes.

### 3.2 The escalation: this is a WRONG WRITE, not a nag loop

The nag-loop framing understates it. On the current live state the consequence chain, if
stage 2 does not ship and the collector recovers on the 10:00 Paris run:

1. **The cursor is held.** The Linear cursor has not moved since the 16:30 Paris run —
   live state and the 18:09:11Z backup both read
   `"linear": {"cursor":"2026-08-11T14:30:26Z","last_success":"2026-08-11T16:30:26+02:00"}`.
   Every run since failed the Linear source and correctly held it. GCN-25's
   `updatedAt` (`2026-08-11T18:08:40.395Z`) is therefore **inside the next successful
   window** — the first healthy run routes it.
2. **The item is `done` with `linear_issue` and no `closed_at`** (section 2.1 above). That is
   precisely the close-back's precondition at base line 529: an item carrying
   `linear_issue` without `closed_at` whose status is `done`/`dismissed` by an explicit
   resolution. **The close-back fires.**
3. **Step 1 does not skip.** Its test is `state.type` `completed` or `canceled`;
   `duplicate` matches neither, so the guard passes the candidate through.
4. **Step 2 checks only `teamId`.** GCN-25 is on GCN
   (`81e7b769-2a46-4e2a-8db5-c165a7963b0e`), the read succeeds, nothing tests the
   `statusType` the native payload already carries.
5. **Step 3 writes Done over Gaetan's Duplicate** — `save_issue` with `state`
   `0434e579-7b85-487a-8cf9-5aed6caaf41b`.
6. **Step 4 "confirms" success on the state it just wrote**: it asserts `statusType ==
   completed` on the returned payload, which is now `completed` *because of step 3*, and
   latches `closed_at`.

A silent, self-certifying write that destroys a human's Duplicate classification, with an
audit trail saying it succeeded. **Stage 1 makes the collector runnable again, which makes
this chain more likely to fire, not less** — the guard against it is stage 2, and until it
ships, P2 in section 8 is the detector and section 9.9 is the hand-revert.

### 3.3 The self-rescue nuance

**The item left `open` only because Gaetan ALSO said it in Slack.** That bounds this
instance — the queue is not currently nagging — but it does not bound the hole, and it is
what **armed** the wrong write: the Slack branch set the item to `done` without setting
`linear_issue.closed_at` (only the Linear-terminal path latches it), and `done` + no
`closed_at` is exactly the close-back's trigger. Had the item stayed `open`, the close-back
would not run at all and the failure would have stayed a nag loop. The rescue converted a
nag loop into a pending overwrite.

### 3.4 The operational cause behind both misses — **this is what stage 1 fixes**

No run ever saw GCN-25 as a routed candidate after it closed. Run inventory from cron
`executions.db` (read-only, `mode=ro`), Paris offsets as stored:

```
16:00:02→16:06:14  62e8cd9db637 builtin completed
16:30:03→16:34:47  62e8cd9db637 builtin completed
17:00:03→17:05:34  759e08c598e3 builtin completed   <- last scheduled run of the day
19:29:32→19:30:28  759e08c598e3 direct  completed   <- aborted, [SILENT], no Linear call
19:42:44→19:43:54  759e08c598e3 direct  completed   <- aborted, [SILENT], no Linear call
```

GCN-25 went Duplicate at 17:10:41 Paris — **5 min 07 s after the day's final pass
finished**. The two later `direct` runs died on
`BLOCKED: Command flagged as dangerous (script execution via -e/-c flag)`; the 20:08 manual
run had its collector blocked by the approval gate
(`BLOCKED: Command timed out without user response`, 302.81 s) and recorded exactly that in
`last_failure_notice.linear`. The same `pending_approval` block appears on the 11:00, 11:30,
12:00, 12:30, 13:00, 15:30 and 17:00 Paris runs — **7 runs whose collection actually
failed**, plus the 20:08 manual run. The 16:30 run carries the block too, but it is the run
that holds the cursor's `last_success` (`2026-08-11T16:30:26+02:00`), so its Linear
collection did complete — it is the day's **last successful collection**, not a failure, and
is not counted above.

That chronic collector failure is the operational cause of both misses. It is a
VM/config-shaped defect that **stage 1 addresses in the skill text** by replacing the
invocation shape with one the guard lets through (section 5). It is not a semantics defect
and no stage-2 hunk touches it. The residue that no SKILL.md edit can fix is in section 10(b).

---

## 4. Linear `statusType` AUDIT — team GCN

This audit is what makes stage 2 **land once** rather than per status type: it enumerates
the whole enum, so the terminal set can be stated as a closed set and a future type cannot
fall through silently.

Raw `list_issue_statuses(team="GCN")` (team confirmed by `get_team(query="GCN")` →
`{"id":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","name":"Gaetan"}`):

```json
[{"id":"a428dade-3c2d-42b0-86ed-50460344ca41","type":"started","name":"In Progress"},
 {"id":"95467ad9-9bad-4942-86d9-82d65e123a7b","type":"backlog","name":"Backlog"},
 {"id":"77aad3b3-deac-49a7-a39e-1bea02d93820","type":"canceled","name":"Canceled"},
 {"id":"59ed732a-f242-4eba-926f-1c0d128fe83c","type":"unstarted","name":"Todo"},
 {"id":"22cb42de-adf9-4c0d-a136-6af48300af8b","type":"started","name":"In Review"},
 {"id":"0ea8bdc6-215f-48d3-b228-779422c6b03e","type":"duplicate","name":"Duplicate"},
 {"id":"0434e579-7b85-487a-8cf9-5aed6caaf41b","type":"completed","name":"Done"}]
```

7 statuses, 6 distinct types (`started` used twice).

| statusType | verdict for this skill | justification |
|---|---|---|
| `triage` | **LIVE** | undecided ≠ resolved. **Not configured on GCN** — no status of this type exists in the team today; listed so the enum is closed |
| `backlog` | **LIVE** | accepted, not started; still on Gaetan's plate |
| `unstarted` | **LIVE** | queued (Todo); still live |
| `started` | **LIVE** | actively worked (In Progress / In Review) |
| `completed` | **TERMINAL** | done — nothing left to act on → item `done`, reason `linear:completed` |
| `canceled` | **TERMINAL** | explicitly dropped — stop tracking → item `dismissed`, reason `linear:canceled` |
| `duplicate` | **TERMINAL** | work continues under the canonical issue, not this one — stop tracking this one → item `dismissed`, reason `linear:duplicate` |

**Three terminal, four live, seven total — the entire enum.**

**Evidence that `duplicate` is a distinct enum value and not an alias of `canceled`:**

- `list_issue_statuses` returns them as two separate rows with two ids and two `type`
  strings: `{"…77aad3b3…","type":"canceled","name":"Canceled"}` and
  `{"…0ea8bdc6…","type":"duplicate","name":"Duplicate"}`.
- GCN-25's live state id is `0ea8bdc6-215f-48d3-b228-779422c6b03e` — the Duplicate row
  exactly — and `get_issue` reports `"status":"Duplicate","statusType":"duplicate"`.
- On the **collector** transport, GCN-25's `stateHistory` carries the very object shape
  §4's GraphQL selection (`state { id name type }`) returns:
  `{"state":{"id":"0ea8bdc6-…","name":"Duplicate","type":"duplicate"},…}`. So the
  collector will report `state.type == "duplicate"`.
  *Honest limit:* no collector payload for a duplicate-typed issue has ever been captured
  end to end. The inference is strong but one measurement short — see section 11.
- Population check, `list_issues(team="GCN", limit=250, includeArchived=true)`,
  `hasNextPage:false`, 32 issues: Done/completed 18, Backlog/backlog 4, Canceled/canceled 6,
  Todo/unstarted 2, In Progress/started 1, Duplicate/duplicate 1 (GCN-25), In Review 0.
  18+4+6+2+1+1 = 32. Every type except `triage` is met in practice.

**The `canceledAt` trap.** GCN-25 carries `"canceledAt":"2026-08-11T15:10:41.624Z"` with
`"completedAt":null` **while its statusType is `duplicate`**. Linear stamps `canceledAt`
for the whole cancel-like group (canceled ∪ duplicate). A checker keying on "is
`canceledAt` non-null" conflates the two types; a checker keying on the status **name**
breaks on any rename (names are freely editable per team). **Key on the type enum only.**
GCN's names all happening to agree with their types today (checked across all 7) is the
reassuring case, not a licence to string-match.

---

## 5. STAGE 1 — PLUMBING, APPLIED 2026-08-11

**This section describes a change that is live.** Live VM == worktree == `origin/main`, all
three at md5 `1caff77d5dd4597756dc0e5b4a54a39c`, 716 lines, 71567 bytes, mode 600, version
`1.7.2`. Mirror commit `d5a051d` on `origin/main`. Its rollback is section 5.6.

### 5.1 What changed, and why

Exactly **three** changes against the pre-fix base `a117cf494ea53d18a0fd1ffbc1b2310b`:

1. **§4, base line 149 — the collector invocation paragraph replaced by the cron-stable
   two-step shape.** Spliced **byte-identical** from the GCN-31 lane's verified text
   (`b97dff6`, `status/probes/gcn31-collector.md` §4.4; fence body 1953 bytes, md5
   `ca5c42cd4d44e2fcd134ac291fce094a`) — one clause excepted, section 5.3. The old text told
   the run to "run the collector below locally with `python3`", which an agent reaches for
   as `python3 -c` / heredoc / pipe-to-interpreter — every one of those shapes is what the
   command guard flags. Under `approvals.cron_mode: deny` a flagged command becomes a
   `status: "blocked"` tool error that never reaches the delivered report: **the collection
   simply does not happen and the run goes `[SILENT]` looking healthy.** The new text
   mandates `write_file` then run the file by absolute path, which is the one invocation the
   guard's exec-flag scan stops before reading.
2. **Line 139 — the Google Workspace interpreter.** `GAPI="python …google_api.py"` →
   `GAPI="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python …google_api.py"`. Bare
   `python` is not the interpreter that has the skill's dependencies.
3. **Front matter — `version: 1.7.1` → `1.7.2`** (section 1.1).

**Nothing in §5, §7 or §8 moved.** Stage 1 changes how work is invoked, never what the skill
decides.

### 5.2 The diff actually applied (base → stage 1)

```diff
--- base (d5a051d^, a117cf494ea53d18a0fd1ffbc1b2310b, 714 lines)
+++ skills/orchestration/engagement-checker/SKILL.md (1caff77d5dd4597756dc0e5b4a54a39c, 716 lines)
@@ -1,7 +1,7 @@
 ---
 name: engagement-checker
 description: "Use for incremental follow-up and commitment reminders."
-version: 1.7.1
+version: 1.7.2
 required_environment_variables: [LINEAR_API_KEY]
 metadata:
   hermes:
@@ -136,7 +136,7 @@
 Use the configured Google Workspace script read-only:
 
 ```bash
-GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
+GAPI="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
 $GAPI gmail search "after:YYYY/MM/DD" --max 100
 ```
 
@@ -146,7 +146,9 @@
 
 ## 4. Collect Linear deltas
 
-Require `LINEAR_API_KEY` to be present in the collector process environment without reading or displaying its value during prerequisite checks. Set `LINEAR_CURSOR` to the stored source cursor and `LINEAR_RUN_END` to the run's fixed end timestamp, then run the collector below locally with `python3`.
+Require `LINEAR_API_KEY` to be present in the collector process environment without reading or displaying its value during prerequisite checks. Set `LINEAR_CURSOR` to the stored source cursor and `LINEAR_RUN_END` to the run's fixed end timestamp.
+
+**Materialize the fenced block below to a file with the `write_file` tool, then run that file by absolute path — `LINEAR_CURSOR='…' LINEAR_RUN_END='…' python3 /tmp/engagement_linear_collector.py`. That two-step shape is mandatory and the shell command carries nothing else.** A scheduled run has no user present to answer an approval prompt, so `approvals.cron_mode: deny` turns any flagged command into a `status: "blocked"` tool error that never reaches the delivered report: the collection simply does not happen and the run goes `[SILENT]` looking healthy. The guard flags exactly the shapes an agent reaches for first, and each one is forbidden here: `python3 -c '…'` and `python3 - <<'PY'` (*script execution via -e/-c flag*, *via heredoc*), `perl -ne`/`ruby -e` line extraction (same rule, interpreter family), piping the block into an interpreter — `… | python3 -` (Tirith *pipe_to_interpreter*, HIGH) — and `rm -f /tmp/<file>` afterwards, which matches *delete in root path* because the path is absolute. **Leave the extracted file where it is; do not clean it up.** `write_file` is not command-guarded and needs no approval, and a bare interpreter followed by a path is the one invocation the guard's exec-flag scan stops before reading — env-var prefixes and `&&` chaining of two such invocations are inside the safe shape. Measured 2026-08-11 in cron run `cron_62e8cd9db637_20260811_140000`: two `python3 -c` extraction attempts were stopped at 12:02:58 and 12:03:16 UTC — logged `"status": "pending_approval", "approval_pending": true`, the approval-fallback branch rather than the `cron_mode: deny` block, which changes the branch but not the outcome: the command never ran, then `write_file` at 12:03:50 and `python3 /tmp/extract_engagement_collector.py && LINEAR_CURSOR='…' LINEAR_RUN_END='…' python3 /tmp/engagement-linear-collector.py` at 12:04:01 returned `exit_code: 0`.
 
 **Run it through the shell/terminal tool — never `execute_code`.** Measured: the `execute_code` sandbox scrubs every environment variable whose name contains `KEY`, …
```

`diff -u` over the two files reports exactly **3 hunks** — reproducible with
`git show d5a051d^:skills/orchestration/engagement-checker/SKILL.md`.

### 5.3 NAMED DEVIATION — the `pending_approval` clause

**Ratified by the coordinator. This is the one place stage 1 is not a byte-identical splice
of the GCN-31 text**, and the block's own author instructed the correction in
`gcn31-collector.md` §4.5.

OLD (as it stood in `b97dff6`):

```
two `python3 -c` extraction attempts were blocked at 12:02:58 and 12:03:16 UTC
```

NEW (as shipped):

```
two `python3 -c` extraction attempts were stopped at 12:02:58 and 12:03:16 UTC — logged
`"status": "pending_approval", "approval_pending": true`, the approval-fallback branch
rather than the `cron_mode: deny` block, which changes the branch but not the outcome: the
command never ran
```

Verified read-only on the VM tonight — `agent.log` carries:

```
"status": "pending_approval", "approval_pending": true, "command": "python3 -c \"from pathlib import Path; s=Path('/home/gaetan/.hermes/skills/orchestratio...
```

**Rationale for the record:** shipping a known-false measured claim into a live skill is the
same failure family as GCN-32 — an invented or wrong measured statement reads to the next
person exactly like a real one. The correction changes the named branch, not the conclusion:
the command never ran either way, so the mandated two-step shape is unaffected.

### 5.4 The apply sequence as executed

In order, all of it tonight, 2026-08-11:

1. **Backup on the VM, before any write** —
   `cp -p ~/.hermes/skills/orchestration/engagement-checker/SKILL.md \
    …/SKILL.md.bak-gcn30-32-stage1-20260811T202839Z`.
   The backup is the pre-apply file: md5 `a117cf494ea53d18a0fd1ffbc1b2310b`, 69654 bytes.
2. **Transfer to a temp file, never in place** —
   `cat <stage-1 file> | ssh gaetan@192.168.0.9 'umask 077; cat > …/SKILL.md.new'`.
   No heredoc (GCN-13 merged six empty SKILL.md files that way).
3. **md5 verified on `.new` BEFORE the swap** — `1caff77d5dd4597756dc0e5b4a54a39c`.
4. **Atomic swap under the lock** —
   `ssh … 'flock ~/.hermes/.wf3.lock -c "mv …/SKILL.md.new …/SKILL.md"'`.
5. **Re-verified after the swap** — md5 `1caff77d5dd4597756dc0e5b4a54a39c`, 716 lines,
   71567 bytes, mode `600`.
6. **Mirror commit** at the real path
   `skills/orchestration/engagement-checker/SKILL.md` → `d5a051d`.
7. **Pushed** — `git pull --rebase origin main && git push origin HEAD:main`.
8. **Byte-equality proven live vs `origin/main`** — all three md5s identical.

Re-verified read-only after the fact (this is the current state, not a plan):

```
$ ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
    wc -lc …; stat -c "%a %y" …'
1caff77d5dd4597756dc0e5b4a54a39c  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
  716  71567 /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
600 2026-08-11 20:28:40 +0000

-rw------- 1 gaetan gaetan 71567 Aug 11 20:28 SKILL.md
-rw------- 1 gaetan gaetan 69654 Aug 11 15:31 SKILL.md.bak-gcn30-32-stage1-20260811T202839Z
-rw------- 1 gaetan gaetan 26089 Aug 10 11:01 SKILL.md.bak-wf6-20260810T193012Z
-rw------- 1 gaetan gaetan 65404 Aug 10 19:30 SKILL.md.bak-wf6-ec170-20260810T202040Z
-rw------- 1 gaetan gaetan 67313 Aug 10 20:27 SKILL.md.bak-wf6-envvar-20260810T210155Z

$ md5sum skills/orchestration/engagement-checker/SKILL.md   # worktree, == origin/main
1caff77d5dd4597756dc0e5b4a54a39c
```

(`cp -p` preserves the source mtime, which is why the backup's `ls` timestamp reads 15:31 —
that is the pre-apply file's own mtime, not the backup time. The backup **name** carries the
apply time, `20260811T202839Z`.)

### 5.5 Gateway evidence — a live-reload, not a restart

```
$ systemctl --user is-active hermes-gateway            → active
$ systemctl --user show hermes-gateway -p ActiveEnterTimestamp
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

**Unchanged across the write.** Hermes picked the new skill up by live-reload (~30 s); the
unit was never restarted, so nothing about tomorrow's queue or run history was disturbed.

### 5.6 STAGE 1 ROLLBACK — self-contained

Someone may need this at 07:00 without reading anything else. Same discipline as the apply:
copy → verify → `flock` + `mv`.

```bash
# 1. VM: stage the restore from the dated backup (copy first — never cp over the live file;
#    a live-reload landing mid-copy reads a truncated skill).
ssh gaetan@192.168.0.9 'cp -a ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-gcn30-32-stage1-20260811T202839Z \
  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.restore \
  && md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.restore'
# MUST read a117cf494ea53d18a0fd1ffbc1b2310b BEFORE the mv. If not, stop.

# 2. atomic swap under the lock
ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "mv ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.restore \
  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md"'

# 3. verify the restore, and clear any stray .new
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  wc -l ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  rm -f ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new'
# MUST be a117cf494ea53d18a0fd1ffbc1b2310b and 714 lines.

# 4. repo: revert the mirror commit and push
cd /home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix
git revert --no-edit d5a051d
git pull --rebase origin main && git push origin HEAD:main
```

Then wait **at least 35 s** before triggering, measuring or reasoning about any run — a run
started inside the reload window may still be executing 1.7.2 and is not evidence either way.

**If stage 2 has already been applied, do not run this** — it would restore the pre-fix base
and silently revert stage 2 as well. Roll stage 2 back first (section 9.7), then this.

---

## 6. STAGE 2 — SEMANTICS, STAGED, NOT APPLIED

`status/staged/gcn30-32-engagement-checker-SKILL.md` · md5
`d0adb8917fa7d793465bbf9a96b6bb87` · **718 lines** · 77072 bytes · `version: 1.8.0`.

### 6.1 It was REBASED onto stage 1 — and why that mattered

The staged file was originally cut from the pre-fix base `a117cf49`. Section 9 applies it as
a **whole-file replacement**. Applied unchanged after stage 1, it would have **silently
reverted stage 1** — §4's cron-runnable shape and the venv interpreter would have gone back
to the text that cannot run from cron, with no error, no conflict, and a version line that
said 1.8.0. That is the exact failure mode a whole-file transfer invites, and it is why the
rebase is recorded here rather than assumed.

The staged file has been re-cut on top of stage 1. Three-way proof, all four commands run
tonight:

```
$ diff -u <base a117cf49> skills/orchestration/engagement-checker/SKILL.md | grep -c '^@@'
3          <- base → stage 1: the 3 plumbing hunks (version, line 139, §4 line 149)

$ diff -u skills/orchestration/engagement-checker/SKILL.md \
          status/staged/gcn30-32-engagement-checker-SKILL.md | grep -c '^@@'
5          <- stage 1 → stage 2: the 5 semantic hunks, and nothing else

$ diff -u <base a117cf49> status/staged/… | grep -c '^@@'
7          <- base → stage 2: both sets (two adjacent hunks coalesce)

$ grep -c 'venv/bin/python'                       <stage 1> <stage 2>   → 1 / 1
$ grep -c 'Materialize the fenced block below'    <stage 1> <stage 2>   → 1 / 1
$ grep -c 'pending_approval'                      <stage 1> <stage 2>   → 1 / 1

$ sed -n '147,506p' <stage 1> | md5sum   → 9d0b9dce27edc1fbc458ed181d64ed0e
$ sed -n '147,506p' <stage 2> | md5sum   → 9d0b9dce27edc1fbc458ed181d64ed0e
```

So: **stage 2 carries both plumbing changes and the corrected `pending_approval` clause, and
§4 (lines 147–506) is byte-identical between stage 1 and stage 2.** Applying stage 2 cannot
revert stage 1.

The pre-rebase staged file — md5 `cb33422c0e8462a74e55c98c73f7b488`, 716 lines — is
**superseded and must not be applied**; it is the one that would have reverted stage 1. An
earlier draft still (`35f66f755a896e2f22789b6041c034ff`) appears in `review-voice.md` and is
likewise superseded. Only `d0adb8917fa7d793465bbf9a96b6bb87` / 718 lines is live.

### 6.2 Complete unified diff — stage 1 → stage 2

Regenerated verbatim tonight with

```
diff -u /home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix/skills/orchestration/engagement-checker/SKILL.md \
        /home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix/status/staged/gcn30-32-engagement-checker-SKILL.md
```

```diff
--- skills/orchestration/engagement-checker/SKILL.md	2026-08-11 20:29:10.403503062 +0000
+++ status/staged/gcn30-32-engagement-checker-SKILL.md	2026-08-11 20:29:41.683695671 +0000
@@ -1,7 +1,7 @@
 ---
 name: engagement-checker
 description: "Use for incremental follow-up and commitment reminders."
-version: 1.7.2
+version: 1.8.0
 required_environment_variables: [LINEAR_API_KEY]
 metadata:
   hermes:
@@ -518,8 +518,8 @@
 
 Reconcile new events before scoring:
 
-- Gaetan replied in the Slack/email thread, explicitly completed the action, or Linear reached a terminal state: `done`.
-- A routed filed issue reporting `state.type == "completed"`: parent item `status = "done"`, status-history reason `linear:completed`. `state.type == "canceled"`: `status = "dismissed"`, reason `linear:canceled`. Closing the issue in Linear is how Gaetan clears the item. `state.type` is the **collector's** raw GraphQL shape and is correct here; a native read has no `state` object at all — it carries flat `status` and `statusType` (`linear-ticketing` §9).
+- Gaetan replied in the Slack/email thread or explicitly completed the action: `done`. A Linear terminal state resolves through the terminal set in the next bullet and never through this one — that set maps each type to its own status, and a blanket `done` here would record a Canceled or Duplicate issue as work Gaetan finished.
+- A candidate reporting a terminal `state.type` for an issue an item already carries — routed to it through `linear_issue`, or matching an item whose own id is `linear:<key>`. §8 keys its bullet handle on those same two shapes, a Linear identity rather than Tars having filed it, and reconcile keys on them for the same reason. **The terminal set is these three types, stated here once and referenced everywhere else**: `completed` → parent item `status = "done"`, status-history reason `linear:completed`; `canceled` → `status = "dismissed"`, reason `linear:canceled`; `duplicate` → `status = "dismissed"`, reason `linear:duplicate`. Closing the issue in Linear is how Gaetan clears the item, and Duplicate is one of the ways he closes one: the work continues under the canonical issue, not this one. The remainder — `triage`, `backlog`, `unstarted`, `started` — is the whole non-terminal set: three terminal types, four live ones, the entire `statusType` enum. `triage` is not configured on GCN, which is why §7a covers the whole non-terminal board in three calls. So the set is closed and an eighth type is an edit to this file, never a silent fall-through to "not terminal". **Key on the type enum, never on the status name**: a workspace renames a status freely, and GCN's names agreeing with their types today guarantees nothing about tomorrow or about another team. **Never read terminality off `canceledAt`**: Linear stamps that field for the whole cancel-like group, so GCN-25 carries `canceledAt` while its type is `duplicate` — the field neither implies `canceled` nor is required by it. `state.type` is the **collector's** raw GraphQL shape and is correct here; a native read has no `state` object at all — it carries flat `status` and `statusType` (`linear-ticketing` §9). Same enum, two shapes.
 - A routed filed issue in a **non-terminal** state whose parent is `done` or `dismissed` with a `linear:*` reason: return the parent to `open` with a fresh cooldown, **and delete `linear_issue.closed_at` in the same step**. `closed_at` is what suppresses the close-back; leaving it set on a reopened item means the loop can never be closed again and the issue becomes a permanently open ghost. A latch with no clearing step is a dead mechanism — this is its clearing step. Reopening an issue must make it visible again, not bury it until the 14-day sweep.
 - A routed filed issue, any state: **refresh `linear_issue.priority` from the candidate's `issue["priority"]`** — a bare int on the collector path (`linear-ticketing` §2). Without it §8 sorts filed bullets forever on Tars's original guess and disagrees with the daily board about the same issue. This is the only consumer of the collector's `priority` field.
 - A routed candidate matching no item: treat it as a fresh candidate.
@@ -530,8 +530,8 @@
 
 Close in the item → Linear direction too, or every loop Gaetan resolves in Slack or email leaves a permanently open ghost issue on the board — and the daily brief's board renders those ghosts ahead of real work. This is permitted write 3. It runs for any item carrying `linear_issue` without `closed_at` whose status is `done` or `dismissed` **by an explicit resolution — a user instruction, a reply that closed the thread, or a terminal Linear state — and never by the 30-day staleness sweep**, including one that reached that status on an earlier run. Check the status-history reason, not the bare status: a `stale:30d` dismissal is Tars losing sight of the loop, not Gaetan resolving it, and closing on it would Cancel a live human ticket silently.
 
-1. **Skip the write when the issue is already terminal.** If this run's routed candidate reports `state.type` `completed` or `canceled` — that is how the item reached `done`/`dismissed` in the first place — set `linear_issue.closed_at` from that observation and make no call. Writing Done to an already-Done issue only bumps `updatedAt` and pulls it back into the next window for nothing.
-2. **Confirm the issue is still on GCN.** Read it with `mcp__linear__get_issue`, `id` = `linear_issue.id`, and check `teamId` is `81e7b769-2a46-4e2a-8db5-c165a7963b0e`. **If it is not GCN, or the read fails, do not write.** Record the fact on the item and surface one line in §8 naming the issue and the team it now sits on. A company-team write needs Gaetan's per-message instruction; a cron run has none, and "it is only a cleanup" is not an exception.
+1. **Skip the write when the issue is already terminal.** If this run's routed candidate reports a `state.type` in the terminal set above — that is how the item reached `done`/`dismissed` in the first place — set `linear_issue.closed_at` from that observation and make no call. Writing Done to an already-Done issue only bumps `updatedAt` and pulls it back into the next window for nothing.
+2. **Confirm the issue is still on GCN, and still not terminal.** Read it with `mcp__linear__get_issue`, `id` = `linear_issue.id`, and check `teamId` is `81e7b769-2a46-4e2a-8db5-c165a7963b0e`. **If it is not GCN, or the read fails, do not write.** Record the fact on the item and surface one line in §8 naming the issue and the team it now sits on. A company-team write needs Gaetan's per-message instruction; a cron run has none, and "it is only a cleanup" is not an exception. **Then test that same payload's flat `statusType` against the terminal set — zero extra calls, the read is already in hand.** If it is terminal the issue is already resolved: set `linear_issue.closed_at` from that read, make no write, and never report it as a close. This backstop exists because the reconcile above is event-driven only — it fires for a routed candidate inside the run's cursor window and nowhere else — so an issue that goes terminal while the cursor is held is never routed, step 1 never sees it, and step 3 overwrites a resolution a human made. It reaches only the items the close-back already reaches: an item still `open`, `waiting` or `snoozed` is not read back anywhere, and reconciles on the first successful collector run instead. Concretely: Gaetan marks the issue Duplicate, no run routes it, step 3 writes Done over that, and step 4's assertion passes on the state it has just written and reports the close as confirmed. A wrong write that certifies itself is the one failure this list exists to prevent.
 3. On GCN, call `mcp__linear__save_issue` with **`id` and `state` only** — `state` `0434e579-7b85-487a-8cf9-5aed6caaf41b` (Done) for `done`, `77aad3b3-deac-49a7-a39e-1bea02d93820` (Canceled) for `dismissed`. Never pass `labels`, not even empty: it replaces the whole set, and omitting it is measured to preserve it.
 4. **Confirm the move on the returned payload**: parse the `result` string and assert `statusType` is `completed` (or `canceled`). An unresolvable `state` returns a normal payload with the old state and changes nothing — "no error" is not success (`linear-ticketing` §10). Only a confirmed move sets `linear_issue.closed_at`.
 
@@ -666,13 +666,15 @@
 
 Anything short of confirmation is a creation **failure**: fail open, leave the item unfiled, log a bounded coverage note, retry next cycle — which §7a makes safe. Never record `linear_issue` from an unconfirmed response and never report the issue to Gaetan as filed. Reads still fail closed.
 
+**Provenance is a rule about the value, not about the mechanism that writes it, and it binds every site that records one — §5's `closed_at` latch and its `priority` refresh included, not only this section.** Every value in `linear_issue` that Linear owns — the `id` always, a `priority` refreshed from a candidate, a latched `closed_at` — is copied from a Linear tool response this run actually received and still holds: the confirmed `save_issue` payload above, the one confirming `get_issue`, a §7a adoption's read, §5 step 2's team-and-terminality read, a §5 step 4 confirmed close, or a routed collector candidate. (A create's starting `priority` is the integer §7b computed, exactly as the checklist above says; it is Tars's own judgment, not a Linear reading, and the id is the field a wrong value silently mis-routes.) An id that is remembered, inferred from the conversation, carried over from an earlier run's transcript, or typed as a literal into code is **not a source**, and writing state through `execute_code` is no exemption — a dict assembled by hand is the same write with the section heading removed. Measured: a run persisted `linear_issue: {"id": "GCN-26", …}` as a hardcoded Python literal inside an `execute_code` call, in a cycle that made zero Linear tool calls, from the same script whose `last_failure_notice.linear` recorded Linear as unreachable. If Linear was unreachable or the call failed this run, `linear_issue` stays absent and the item stays unfiled — the fail-open path in the paragraph above, unchanged. A believed-but-unsourced key is at most a note on the item and never a field of `linear_issue`: §4's routing index and §8's bullet handle are both keyed on `linear_issue.id`, so a wrong key mis-routes both, silently and for as long as the item lives.
+
 ## 8. Deliver or remain silent
 
 If nothing crosses the threshold, persist source progress, set `last_completed_run`, release the lock, and return exactly:
 
 `[SILENT]`
 
-The one exception is the coverage break below: any of the seven must-never-hide states is reported even on an otherwise silent run.
+The one exception is the coverage break below: any of the seven must-never-hide states is reported even on an otherwise silent run. A silent run owes no explanation beyond that; when one is given anyway, this section's last paragraph governs what it may cite.
 
 Otherwise send at most six items, grouped when they share a person or topic. One line each:
 
@@ -700,7 +702,7 @@
 
 Do not restate the rate limit anywhere else; point at this paragraph.
 
-On an internal failure, preserve unadvanced source cursors. If no trustworthy reminder can be produced, return `[SILENT]`; never fabricate substitute evidence.
+On an internal failure, preserve unadvanced source cursors. If no trustworthy reminder can be produced, return `[SILENT]`; never fabricate substitute evidence. **That ban covers the explanation as much as the evidence.** A delivery-or-silence decision cites only a gate this run actually evaluated, and quotes that gate's live text. The run gate is the "Workday gate" section and the schedule is the "Scheduling contract" section; neither is restated here, so read the live text there rather than recalling it. Measured: a run justified its silence with "the cycle ran outside the 09:00–19:00 window", a window no live text anywhere defines — the live restriction is 10:00–17:00. An invented reason reads to the next person exactly like a real rule and is the same failure as an invented reminder. **When nothing crossed §6's threshold there is no gate to cite**: the run is the bare `[SILENT]` of this section's first rule, and if it is ever asked why it was silent the whole answer is that nothing crossed the threshold. Reaching for a gate is where the invention starts.
 
 ## Scheduling contract
```

Shape of the change: **5 diff sites, 9 insertions / 7 deletions, 716 → 718 lines.** Sites:
H7 (front matter) · H1+H2 (contiguous, one hunk) · H3+H4 (contiguous, one hunk) · H6+H5b
(one hunk) · H5. The staged file itself is authoritative.

### 6.3 Per-hunk OLD → NEW with rationale

Line numbers below are **stage-1 (live) lines**; the pre-fix base number is given in
parentheses where they differ.

**H7 — front matter, line 4.**
OLD `version: 1.7.2` → NEW `version: 1.8.0`. Minor digit because H1–H6 change what the skill
decides and what it may write — see the version scheme in section 1.1. Nothing else in the
front matter changed; it carries no date or `last_updated` field (verified: `name`,
`description`, `version`, `required_environment_variables`,
`metadata.hermes.{tags,category}` only).

**H1 — §5, line 521 (base 519).**
OLD: `- Gaetan replied in the Slack/email thread, explicitly completed the action, or Linear reached a terminal state: `done`.`
NEW: see diff.
The old bullet asserted `done` for *any* terminal Linear state, contradicting the next line
of the same file, which correctly maps `canceled` → `dismissed`. GCN-25 is the measured case
where that would record a Duplicate as work Gaetan finished. The Slack/email half is
unchanged; only the Linear clause moves, and it now defers to the terminal set instead of
restating it.

**H2 — §5, line 522 (base 520) — the reconcile bullet.**
The terminal set is now stated **once, completely, by type enum**: `completed` → `done` /
`linear:completed`; `canceled` → `dismissed` / `linear:canceled`; `duplicate` →
`dismissed` / `linear:duplicate`. The non-terminal remainder (`triage`, `backlog`,
`unstarted`, `started`) is named explicitly so the set is closed and an eighth type cannot
fall through silently. The two kept sentences survive verbatim in substance — "Closing the
issue in Linear is how Gaetan clears the item" and the collector-vs-native shape note
(`state.type` vs flat `statusType`, `linear-ticketing` §9). Two measured traps added in the
file's voice: key on the type enum never the status name; never read terminality off
`canceledAt` (GCN-25 carries it while typed `duplicate`).

**H3 — §5, line 533 (base 531) — close-back step 1.**
OLD tested `state.type` `completed` or `canceled`; NEW tests "a `state.type` in the terminal
set above" — widened **by reference**, so the set stays stated once. Behaviour, the
`closed_at` latch from that observation, the make-no-call rule and the `updatedAt`
rationale are unchanged.

**H4 — §5, line 534 (base 532) — close-back step 2, the hunk that prevents the wrong write.**
The existing teamId rule, its fail-closed clause ("If it is not GCN, or the read fails, do
not write") and its §8 reporting are preserved verbatim; the terminal re-check is
**additive**, on the payload the step already fetched — **zero extra calls** — using the
flat `statusType` the native read carries. If terminal: set `closed_at` from that read,
make no write, do not report it as a close. The rationale states plainly why the backstop
exists (event-driven reconcile + held cursor ⇒ step 1 never sees it ⇒ step 3 overwrites a
human resolution while step 4 confirms on the state it just wrote), and bounds its reach so
nobody reads it as covering `open`/`waiting`/`snoozed` items.

**H5 — §8, line 703 (base 701) — the F1 fix.**
Anchored to the existing no-fabrication sentence, because inventing the reason is the same
failure class as inventing the evidence — the new text says so ("That ban covers the
explanation as much as the evidence"). A delivery-or-silence decision may cite only a gate
the run actually evaluated, quoting that gate's live text; the two real gates are pointed
at by section name ("Workday gate", "Scheduling contract") and **not restated**. When
nothing crossed §6's threshold there is no gate to cite and the whole answer is the
threshold.

**H5b — §8, line 675 (base 673) — placement pointer, from an accepted review finding.**
One sentence at the silence decision pointing at the paragraph that governs it, in the idiom
the file already uses for its rate limit. States no new rule. H5 required both "anchor it to
the closing sentence" and "place it where a reader deciding to be silent will meet it";
those pull apart, and this satisfies the second without moving the first.

**H6 — §7c, inserted after line 667 (base 665) — the F2 fix.**
A provenance rule that cannot be read as §7c-only: it binds "every site that records one —
§5's `closed_at` latch and its `priority` refresh included". Values Linear owns are copied
from a Linear tool response this run received and still holds (sources enumerated,
including §5 step 2's new read and §5 step 4's confirmed close). A remembered / inferred /
carried-over / hard-coded id is **not a source**, and `execute_code` is **no exemption** —
the rule is about where the value came from, not which mechanism writes it, with the
measured GCN-26 incident folded in as evidence. If Linear was unreachable this run,
`linear_issue` stays absent and the item stays unfiled — pointing at §7c's existing
fail-open path, not duplicating it. A believed-but-unsourced id is at most a note on the
item, because §4's routing index and §8's bullet handle are both keyed on it. §7c's
confirmation checklist and flush rules are untouched.

### 6.4 Departures from the tickets' suggested fixes

| # | Departure | One-line rationale |
|---|---|---|
| (a) | **No changelog line added.** | The file has no changelog section and `grep -rEli 'changelog\|change log\|version history' skills/ --include=SKILL.md` returns zero across the whole repo — the version bump is the change record. |
| (b) | **The fix extends to §5 step 2**, the `get_issue` the file already makes, which neither ticket names. | It is the hunk that actually prevents the wrong write: the routed path is event-driven and misses terminal transitions for hours while the cursor is held, so widening only the routed tests (H2/H3) leaves step 3 free to write Done over a Duplicate the moment the close-back fires for any other reason — as it is about to for GCN-25. It costs zero extra calls. |
| (c) | **H1 fixes the pre-existing line-519 contradiction**, which neither ticket names. | Base line 519 asserted `done` for any terminal Linear state while base line 520 mapped `canceled` → `dismissed`; leaving it would have made the new three-type set self-contradicting on the very first bullet. |
| (d) | **H2's subject widened to both Linear-identity shapes** (`linear_issue`, and an item whose own id is `linear:<key>`) — beyond either ticket. | H1 deleted the only §5 rule that closed a `linear:<key>` item on a terminal state, and §4 builds the routing index from `linear_issue` alone; without this the change would fix GCN-30 for filed items and reintroduce it for Linear-sourced ones (live cases tonight: `linear:NMC-649`, `linear:MC-4226`). §8's bullet-handle rule already defines the two-shape identity in these words. |
| (e) | **GCN-32's suggested "`save_issue`/`get_issue` return value or `null`" narrowed to "values Linear owns".** | §7c's own untouched checklist says a create's starting `priority` is the integer §7b computed; the literal wording would forbid the file's documented create path. The `id` — the only field whose wrong value silently mis-routes §4 and §8 — is bound absolutely, and every clause about literals, remembered ids, `execute_code` and unreachable-Linear is unchanged. Declared as a deviation from the design brief; reverting to the literal wording is a one-clause edit. |
| (f) | **GCN-30's "worth checking whether any other status type can be terminal" answered by an audit, not by adding one type.** | Section 4 above enumerates the whole enum and closes the set, which is what makes the fix land once. |
| (g) | **No per-run read-back sweep.** | Deliberately deferred — see section 10(a). |

---

## 7. VERIFICATION of stage 2

**Nothing about stage 2 is measured post-apply. No run was executed with it. No live file
carries it. No Linear write was made.** Everything below is static review and on-paper
walkthrough. (Stage 1's verification is section 5.4–5.5, and it is byte-level, not
behavioural — see section 11.)

### 7.1 The three adversarial reviews

| lens | what it checked | verdict |
|---|---|---|
| **bytes** (`scratchpad/review-bytes.md`) | every diff hunk authorised; §4 untouched; whitespace/unicode/CR drift; no clause silently dropped; base and `skills/` unmodified | no BLOCKER, 1 MAJOR, 2 MINOR |
| **semantics** (`review-semantics.md`) | six scenario walkthroughs S1–S6 traced literally through the staged text | 4 findings (1 MAJOR, 3 MINOR) |
| **voice** (`review-voice.md`) | whole file read, not only the hunks: register match, contradictions with untouched text | 1 BLOCKER, 3 MAJOR, 6 MINOR |

**Adjudication (`scratchpad/adjudication.md`): 9 ACCEPTED, 4 REJECTED, 5 DEFERRED.**

These reviews were run against the staged file as cut from the pre-fix base. The rebase onto
stage 1 (section 6.1) changed no reviewed hunk — the 5 semantic hunks are byte-identical
across the rebase; only their surrounding context moved by +2 lines.

Accepted and folded in:

- **BLOCKER (voice)** — H5's draft clause "**say exactly that and name no gate at all**" is
  an imperative to *speak* in exactly the case §8's first rule says the whole response is
  `[SILENT]`; a cron run obeying it literally posts "nothing crossed the threshold" into
  the reporting conversation every 30 minutes and corrupts any run-count keyed on
  `[SILENT]`. Rewritten to bind the run's *account of itself* (where the measured F1
  sentence appeared), preserving H5's semantics exactly. **The bytes lens declared "no
  BLOCKER" and was wrong** — its lens never compared the new §8 text against §8's own first
  rule. Cross-lens adjudication caught it; one reviewer would not have.
- **MAJOR (semantics)** — the `linear:<key>` coverage regression, fixed in H2 (departure
  (d) above). The only coverage regression any lens found.
- **MAJOR (voice)** — "every field of `linear_issue`" contradicts §7c's untouched checklist
  (departure (e)).
- **MAJOR (voice)** — the provenance rule was still readable as §7c-only, the exact failure
  H6 exists to close; lead-in rewritten to bind every recording site.
- **MAJOR (voice + bytes, same clause)** — a false citation (`linear-ticketing` §2 lists
  seven *states* carrying six types and never mentions `triage`) and an apparent
  contradiction with §7a's "those three types are the whole non-terminal set on GCN".
  Citation dropped, reconciliation stated in place.
- **MINORs** — provenance source list completed with §5 step 2 and step 4; step 2's
  rationale bounded so it does not over-claim coverage; H5b placement pointer added;
  ambiguous antecedent fixed; the real 10:00–17:00 window named next to the fabricated one.

Rejected (4): a deliberate sentence join read as a lost full stop; a curly-quote request
against the file's own convention; citing a line **number** inside the skill text (the file
cross-references by §-number and section name, and a line number rots on the next edit);
and a new "the cron prompt is not an authority" rule that no hunk asked for and line 16
already covers.

Deferred (5): the unknown-eighth-`statusType` fail-closed rule; §7c's factually wrong
heading ("the only place `linear_issue` is written" — §5 writes it four times, pre-existing);
the un-clearable `closed_at` on an item with a non-`linear:*` reason (pre-existing);
the per-run read-back sweep (section 10(a)); the blocked collector (section 10(b)).

### 7.2 On-paper walkthroughs

| # | scenario | verdict |
|---|---|---|
| **S1** | **The wrong write** — GCN-25 (`duplicate`) routes tomorrow, item `done`, no `closed_at` | **PASS, stopped at step 1.** §5's terminal set matches `duplicate` → item `dismissed`, reason `linear:duplicate`; the close-back still runs (precondition unchanged) but step 1 latches `closed_at` and makes no call. No `save_issue`. Note for tomorrow: the item's status flips `done` → `dismissed` with a Linear-sourced reason replacing the Slack-sourced one. Both are terminal, neither nags — do not read the flip as a bug. |
| **S2** | **The nag loop as ticketed** — filed issue routes `duplicate`, item `open` | **PASS.** Item → `dismissed`/`linear:duplicate`; step 1 latches `closed_at`, no call; §6 re-evaluates only `open` items and woken snoozes, so the item leaves the reminder pool. Loop closed. |
| **S3** | **Cursor-held reconcile** — GCN-21 `completed` 15:28:52Z, cursor held at 14:30:26Z, item `waiting` | **Unchanged, as intended.** No candidate is produced, every reconcile bullet is inapplicable, and the close-back needs the item to be `done`/`dismissed` — `waiting` is neither, so step 2's backstop is never reached. GCN-21 reconciles on the first successful collector run. **This change does not fix GCN-21**; the wording was tightened so it does not claim to. |
| **S4** | **F2 replay** — hardcoded `GCN-26` written through `execute_code` with zero Linear calls | **Forbidden, unambiguously.** Four independent grounds in the new §7c paragraph: the value-not-mechanism lead-in; "received **and still holds**" (this run holds none); "typed as a literal into code is **not a source**, and writing state through `execute_code` is no exemption"; and "if Linear was unreachable … `linear_issue` stays absent". No sentence is scoped by "here" or "in this step". |
| **S5** | **F1 replay** — "the cycle ran outside the 09:00–19:00 window" | **Forbidden.** Banned twice: the gate was never evaluated by that run, and it has no live text to quote. The reviewer walked the pre-adjudication draft; the BLOCKER rewrite changed only the no-threshold clause, which now requires the bare `[SILENT]` and, if asked, the threshold as the whole answer. **The rewritten clause was not re-walked by a reviewer** (see section 11). |
| **S6** | **False-positive hunt** — 7 constructed cases where the widened set could make the skill do something the old text got right | **No false positive found.** `triage`/`backlog`/`unstarted`/`started` never reach the terminal mapping; no state exists where a terminal issue still needs the Done/Canceled write; a live issue cannot be read as terminal without a name or `canceledAt` heuristic, both banned in text; the reopen path still works and works *better* (the old text could wrongly reopen a `canceled`-dismissed item later moved to Duplicate); step 4's `completed`/`canceled`-only assertion is correct because the skill only ever writes those two. The one real regression found — `linear:<key>` items — was the accepted MAJOR, fixed. |

---

## 8. POST-APPLY ACCEPTANCE PROBE

**Execution status of the commands below.** Every command in this section was **dry-run
executed read-only on 2026-08-11 evening, pre-apply** (VM read-only, no Linear write, no VM
write) and corrected where it did not execute or could pass vacuously. Real pre-apply output
is pasted inline, marked *measured 2026-08-11 pre-apply* — that is the shape to expect, not
the post-run result. One command was found **hard broken** and is fixed here: P4's
`while read … ssh` loop needed `ssh -n` (without it `ssh` drains the loop's stdin and only
the first range is ever checked). Four more could pass vacuously and now carry guards.

Two things could not be truly run, because they read state that exists only after the run:
P3's before/after `jq` diff — validated instead by building synthetic *after* files from
tonight's real 50-item snapshot, where it caught both a latched `closed_at` and a replay of
the fabricated `GCN-26` item and stayed quiet on the identity case — and P2's optional `jq`
equivalent, which needs the response on disk. Nothing in section 9 was executed.

**Read this first: no live queue item moves from the duplicate fix.** Measured against the
live state file (50 items, 9 non-terminal): **zero non-terminal items are linked to a
`duplicate`-statusType issue.** GCN-25 is the only GCN issue in `duplicate`, and its one
linked item is already `done`. The four non-terminal items carrying a `linear_issue` are
GCN-12 (`unstarted`), GCN-15 (`backlog`), GCN-21 (`completed`) and GCN-24 (`unstarted`).
So this is a **static/dry check plus regression checks, NOT a "watch it move" probe.**
(GCN-21's item *would* move on the next healthy run — but because of the pre-existing
`completed` coverage gap, not because of this fix; it is not valid evidence for it.)

### P1 — byte equality after apply (live == staged == repo)

```bash
STAGED=/home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix/status/staged/gcn30-32-engagement-checker-SKILL.md
REPO=/home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix/skills/orchestration/engagement-checker/SKILL.md
LIVE=$(ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md' | cut -d" " -f1)
echo "live   $LIVE"
md5sum "$STAGED" "$REPO"
```

**P1's mechanics are proven, not merely drafted: stage 1 passed this exact check tonight** —
after its apply, live == mirror == `origin/main` at `1caff77d5dd4597756dc0e5b4a54a39c`, 716
lines, which is what section 5.4 step 8 records. The same three-way comparison is what
stage 2 must satisfy at `d0adb8917fa7d793465bbf9a96b6bb87`.

Current state, measured 2026-08-11 after stage 1 (this is what a stage-2 FAIL looks like —
stage 2 is not applied yet, so live and repo are stage 1 and staged is the target):

```
live   1caff77d5dd4597756dc0e5b4a54a39c
d0adb8917fa7d793465bbf9a96b6bb87  …/status/staged/gcn30-32-engagement-checker-SKILL.md
1caff77d5dd4597756dc0e5b4a54a39c  …/skills/orchestration/engagement-checker/SKILL.md
```

**PASS** if all three md5s are identical and equal `d0adb8917fa7d793465bbf9a96b6bb87`, with
the live file at **718** lines (recompute from the staged file if it was edited after this
document was written).
**FAIL** on any mismatch — do not proceed to the mirror commit; re-run the transfer in
section 9.

### P2 — the GCN-25 guard (the wrong-write regression check) — **the single most important assertion**

> ### PRE-R2 BASELINE — measured live 2026-08-11 evening. Compare against THIS.
>
> ```
> GCN-25  status="Duplicate"   statusType="duplicate"
>         stateHistory length 2, last entry state.id=0ea8bdc6-215f-48d3-b228-779422c6b03e,
>                                startedAt=2026-08-11T15:10:41.449Z, endedAt=null
>         completedAt=null   canceledAt=2026-08-11T15:10:41.624Z
>         updatedAt=2026-08-11T18:08:40.395Z
>         team=Gaetan (81e7b769-2a46-4e2a-8db5-c165a7963b0e)
> ```
>
> **Any deviation from this after the R2 run is the wrong write.** `updatedAt` moving on its
> own is not proof of corruption (a comment moves it too); a third `stateHistory` entry, a
> non-null `completedAt`, or `statusType != "duplicate"` is.

After the next in-window run, **GCN-25 must still be `statusType: "duplicate"`**. If it is
`completed`, Tars overwrote Gaetan's classification. **R2 runs with stage 2 NOT applied**, so
this is the run where the hole is still open — P2 is the detector, and section 9.9 is the
hand-revert if it fires.

```
# from a Claude Code session on cooper:  mcp__claude_ai_Linear__get_issue(id="GCN-25")
# from Tars on the VM (Hermes)        :  mcp__linear__get_issue(id="GCN-25")
```

The two names are the same read through different MCP servers; the VM name does not exist on
cooper and vice versa. Read the returned JSON and assert:

```
.statusType == "duplicate"
.status     == "Duplicate"
.stateHistory[-1].state.id == "0ea8bdc6-215f-48d3-b228-779422c6b03e"
.completedAt == null
```

All four fields were confirmed present in the tool's real response 2026-08-11 — no
assertion names a field the payload does not carry. If the JSON is on disk, one runnable
equivalent:

```bash
jq -e '.statusType=="duplicate" and .status=="Duplicate"
       and (.stateHistory|length)==2
       and .stateHistory[-1].state.id=="0ea8bdc6-215f-48d3-b228-779422c6b03e"
       and .completedAt==null' gcn25.json && echo P2=PASS || echo P2=FAIL
```

**PASS** = `statusType` still `duplicate` and `stateHistory` has gained no entry.
**FAIL** = any `completed`/Done entry appended → the wrong write happened; execute the
hand-revert in section 9.9 and treat the run as corrupting.

Secondary assertion on the same run, on the item side (stage 2's intended effect — it can
only be seen after stage 2 is applied):

```bash
ssh gaetan@192.168.0.9 'cat ~/.hermes/state/engagement-checker.json' \
| jq '.items["slack:CC397R0HY:1786458544.180249"]
      | {status, linear: .linear_issue, last_reason: .status_history[-1].reason}'
```

Measured 2026-08-11 (the documented pre-state — this is the "not reached" shape):

```json
{
  "status": "done",
  "linear": { "id": "GCN-25", "priority": 3 },
  "last_reason": "slack: Gaetan marked the engagement-checker item duplicate of the Help Tech investigation and closed the request"
}
```

**PASS** = `status` is `dismissed` with `last_reason` starting `linear:duplicate`, **and**
`linear_issue.closed_at` is now set.
**NOT-REACHED** = still `done` with no `closed_at` — the reconcile never ran, the collector
was blocked again, or stage 2 is not applied yet; a coverage miss, not a fix failure.
**FAIL** = any other combination, in particular `dismissed` with reason `linear:canceled`:
the terminal set was applied but `duplicate` was mapped onto the wrong reason — the
`canceledAt` trap of section 4 firing in practice.

### P3 — F2 regression: no `linear_issue` change without a Linear tool call in the same session

Snapshot the state **before** the run (read-only pull, no VM write):

```bash
ssh gaetan@192.168.0.9 'cat ~/.hermes/state/engagement-checker.json' > /tmp/ec-before.json
```

Measured 2026-08-11: exit 0, 111246 bytes, `.items|length = 50`.

After the run:

```bash
ssh gaetan@192.168.0.9 'cat ~/.hermes/state/engagement-checker.json' > /tmp/ec-after.json

# guard first: an empty or missing snapshot makes the filter below print nothing and
# exit 0 — i.e. pass vacuously. Proven, so do not skip this.
for f in /tmp/ec-before.json /tmp/ec-after.json; do
  n=$(jq -r '.items|length' "$f" 2>/dev/null) || { echo "P3 ABORT: $f unreadable"; exit 1; }
  [ "${n:-0}" -ge 40 ] || { echo "P3 ABORT: $f has only $n items"; exit 1; }
done

# every item whose linear_issue changed
jq -s -r '
  (.[0].items // {}) as $b | (.[1].items // {}) as $a
  | $a | to_entries[]
  | select((.value.linear_issue // null) != (($b[.key].linear_issue) // null))
  | [.key, (.value.linear_issue|tostring)] | @tsv
' /tmp/ec-before.json /tmp/ec-after.json
```

This filter iterates the *after* snapshot, so an item that **loses** its `linear_issue`
between snapshots is invisible to it. That is a known blind spot, not the F2 failure mode;
it is not worth widening the filter for.

For each changed item, the run's session must contain at least one Linear tool call. Derive
the session id of the run (do not hand-type it — a wrong id makes P3, P4 and P5 all pass
vacuously), then count its Linear calls (the VM has **no `sqlite3` CLI** — the same SELECT is
run through Python's stdlib on a read-only URI):

```bash
# derive the most recent cron session. NOTE: LIKE 'cron_%' is wrong here ('_' is a LIKE
# wildcard); GLOB is correct. Verified executing 2026-08-11.
SESSION=$(ssh gaetan@192.168.0.9 python3 - <<'PY'
import sqlite3
c = sqlite3.connect("file:/home/gaetan/.hermes/lcm.db?mode=ro", uri=True)
q = """SELECT session_id FROM messages WHERE session_id GLOB 'cron_*'
       GROUP BY session_id ORDER BY MAX(timestamp) DESC LIMIT 1"""
print(c.execute(q).fetchone()[0])
PY
)
echo "SESSION=$SESSION"   # measured 2026-08-11: cron_e231e5faf180_20260811_194748
# Sanity-check it is TOMORROW's R2 run before using it (the id embeds the local date-time:
# cron_<jobid>_<YYYYMMDD>_<HHMMSS>). A manual run has a bare 20260812_HHMMSS_<hex> id
# instead; take that one from the run's own output if R2 is run by hand.

ssh gaetan@192.168.0.9 python3 - "$SESSION" <<'PY'
import sqlite3, sys
sid = sys.argv[1]
c = sqlite3.connect("file:/home/gaetan/.hermes/lcm.db?mode=ro", uri=True)
q = """SELECT COUNT(*) FROM messages
       WHERE session_id = ?
         AND (tool_name LIKE '%linear%' OR IFNULL(tool_calls,'') LIKE '%mcp__linear__%')"""
print("linear_tool_calls", c.execute(q, (sid,)).fetchone()[0])
PY
```

**PASS** = the jq output is empty (no `linear_issue` changed), **or** every changed item is
accompanied by `linear_tool_calls >= 1`.
**FAIL** = at least one `linear_issue` changed while `linear_tool_calls == 0` — that is the
F2 failure reproducing. (Reference failure: session `20260811_180849_14f453` had 126 tool
rows and `linear_tool_calls == 0` while writing `GCN-26`.)

The predicate was run read-only 2026-08-11 and **discriminates on real data**: the
F2 reference session scores `linear_tool_calls 0`, while cron sessions that did reach Linear
score 16–60 (`20260811_160123_f9a59109` 60, `cron_62e8cd9db637_20260811_130059` 22,
`cron_759e08c598e3_20260811_170003` 16). Every column it names exists in `messages`.

**Honest limit — P3 is a session-level proxy.** It catches the exact measured F2 shape: a
`linear_issue` written in a cycle with **zero** Linear calls. It does not catch a run that
makes one legitimate Linear call and hardcodes a different key in an `execute_code` block;
that session scores ≥ 1 and passes. Per-key attribution is a bigger query than this lane
wants.

### P4 — F1 regression: the delivery/silence text cites no gate absent from the live files

Pull the run's final assistant message (the delivery or the account of the silence) from
`lcm.db`, read-only:

```bash
# $SESSION as derived in P3 — same value, do not re-type it
ssh gaetan@192.168.0.9 python3 - "$SESSION" <<'PY' > /tmp/run-text.txt
import sqlite3, sys
sid = sys.argv[1]
c = sqlite3.connect("file:/home/gaetan/.hermes/lcm.db?mode=ro", uri=True)
rows = c.execute("""SELECT content FROM messages
                    WHERE session_id=? AND role='assistant' AND IFNULL(content,'') <> ''
                    ORDER BY timestamp""", (sid,)).fetchall()
print("\n".join(r[0] for r in rows))
PY

# 0. guard: an empty extraction (wrong $SESSION, ssh failure, no assistant turn) makes
# check 1 exit 1, which reads as PASS. Proven. Do not skip.
[ "$(wc -c < /tmp/run-text.txt)" -gt 20 ] || { echo "P4 ABORT: no assistant text for $SESSION"; exit 1; }

# 1. the measured fabrication must be absent
grep -nE '09:00[^0-9]{0,3}19:00|quiet hour|working hour|business hour' /tmp/run-text.txt
# expect: no output, exit 1

# 2. every clock range the run cites must exist in a live file.
# ssh -n is REQUIRED here: without it ssh drains the loop's stdin (the ranges file) and the
# loop runs exactly ONE iteration, silently, making the FAIL unreachable. Proven.
grep -ohE '[0-9]{1,2}:[0-9]{2}[[:space:]]*[-–—][[:space:]]*[0-9]{1,2}:[0-9]{2}' /tmp/run-text.txt \
  | sort -u > /tmp/run-ranges.txt
cat /tmp/run-ranges.txt
while read -r r; do
  n=$(ssh -n gaetan@192.168.0.9 "grep -rF --include='*.md' --include='*.yaml' --exclude='*.bak*' -- '$r' \
        ~/.hermes/skills/ ~/.hermes/SOUL.md ~/.hermes/config.yaml | wc -l")
  echo "$r  live_hits=$n"
done < /tmp/run-ranges.txt
```

**⚠ Do NOT propagate `ssh -n` to the two heredoc calls above (P3's counter, P4's extraction).**
`-n` redirects stdin from `/dev/null`, which silently kills the heredoc: `python3 -` reads
nothing, prints nothing and exits 0 — turning both into always-passes. Verified both ways.
`-n` belongs only on the `ssh` inside the `while read` loop.

**`--exclude='*.bak*'` is load-bearing.** Four `SKILL.md.bak-*` copies now sit next to the
live skill — the three stale `.bak-wf6-*` plus stage 1's
`SKILL.md.bak-gcn30-32-stage1-20260811T202839Z` — and section 9.3 adds a fifth. Without the
exclusion, text a fix *removed* still answers "exists in a live file". Measured 2026-08-11,
with the exclusion:

```
09:00–19:00  live_hits=0     <- the fabricated gate: correctly absent
10:00–17:00  live_hits=1     <- the real gate, in the live SKILL.md only
```

(the same grep without `--exclude` reported `live_hits=4` for `10:00–17:00`, 3 of them
backups — and there are more backups now, so the unexcluded number is only going up.)

Check 1 was run against the real F1 session `20260811_180849_14f453` 2026-08-11 and
**fires on the genuine fabricated sentence**:

```
11:- No reminders were sent because the cycle ran outside the 09:00–19:00 window.
```

**Check 1 is a specific-recurrence check, not a general fabrication detector** — it catches
the measured sentence and three fixed phrases; a freshly worded invention ("9am to 7pm
delivery window") sails past it. Check 2 is the general one: it extracts *any* clock range
and demands live text for it. Read the two criteria in that order.

**PASS** = check 1 produces no output (grep exit 1) **and** every range in
`/tmp/run-ranges.txt` reports `live_hits >= 1`. Realistically the only range that should
appear is `10:00–17:00`, from the Scheduling contract / line 18.
**FAIL** = any hit on check 1, or any range with `live_hits=0` — an invented gate.

### P5 — the run still behaves (no functional break from the edit)

```bash
# state file: sub-second mtime, size, real item counts (jq also fails loudly on corruption)
ssh gaetan@192.168.0.9 "stat -c '%y %s' ~/.hermes/state/engagement-checker.json"
ssh gaetan@192.168.0.9 'cat ~/.hermes/state/engagement-checker.json' \
| jq '{items:(.items|length),
       nonterminal:([.items[]|select(.status=="open" or .status=="waiting" or .status=="snoozed")]|length)}'

# errors, SCOPED TO THE RUN. An unscoped `grep ERROR … | tail -20` returns a wall of
# pre-existing execute_code tracebacks spanning two days and answers nothing.
ssh gaetan@192.168.0.9 "grep -F '[$SESSION]' ~/.hermes/logs/agent.log | grep -cE 'ERROR|Traceback'"
ssh gaetan@192.168.0.9 "grep -F '[$SESSION]' ~/.hermes/logs/agent.log | grep -E 'ERROR|Traceback' | head -20"
```

Measured 2026-08-11 — state file, and the error count for the F2 reference session:

```
2026-08-11 18:22:35.815828422 +0000 111246
{ "items": 50, "nonterminal": 9 }
2                     <- for SESSION=20260811_180849_14f453; a bounded number, not a wall
```

**PASS** = mtime advanced past the run, `items` still ≈ 50 with valid JSON, and the run's own
error count is 0. **FAIL** = state never flushed, the item count collapsed, jq refuses the
file, or the run's own count is non-zero (read the lines and judge; the log is dense with
pre-existing `execute_code` tracebacks, so only lines carrying `[$SESSION]` count).

**P5 is also stage 1's first behavioural evidence.** The R2 run is the first run to use the
new §4 shape; if the collector now actually collects, the Linear cursor advances past
`2026-08-11T14:30:26Z`. Worth recording alongside P5:

```bash
ssh gaetan@192.168.0.9 'cat ~/.hermes/state/engagement-checker.json' | jq '.sources.linear'
# pre-R2: {"cursor":"2026-08-11T14:30:26Z","last_success":"2026-08-11T16:30:26+02:00"}
# a moved cursor = stage 1 worked; an unmoved one = it did not, and section 10(b) is still live.
```

### PASS/FAIL summary line to record

```
SESSION=<derived in P3; every one of P3/P4/P5 keys on it — a wrong id passes all three vacuously>
STAGE 2 APPLIED AT R2 TIME?  yes/no    <- P1 and P2b mean different things either way

P1 byte-equality        PASS/FAIL      (target d0adb8917fa7d793465bbf9a96b6bb87, 718 lines)
P2 GCN-25 still duplicate   PASS/FAIL   <- most important
   (pre-R2 baseline 2026-08-11: statusType=duplicate, status=Duplicate,
    stateHistory len 2 last id 0ea8bdc6-215f-48d3-b228-779422c6b03e, completedAt=null)
P2b item dismissed + closed_at  PASS/FAIL/NOT-REACHED (collector blocked, or stage 2 not applied)
P3 F2 provenance        PASS/FAIL
P4 F1 gate citation     PASS/FAIL
P5 run health           PASS/FAIL
P5b Linear cursor moved past 2026-08-11T14:30:26Z   YES/NO   <- stage 1's first real test
```

---

## 9. STAGE 2 APPLY PROTOCOL — exact and ordered

**None of the commands in this section has ever been executed** — they were composed against
measured facts (paths, hashes, the flock discipline) and read-only-audited, not run.
Everything here writes. (Stage 1 was applied by the same shape of sequence, which is recorded
as executed in section 5.4 — that is evidence the shape works, not that these exact commands
ran.)

**9.1 PRE-APPLY GUARD — the live file must be stage 1. Do this first, before anything else.**

```bash
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  wc -l ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  grep -c "^version: 1.7.2" ~/.hermes/skills/orchestration/engagement-checker/SKILL.md'
# MUST be 1caff77d5dd4597756dc0e5b4a54a39c, 716 lines, version 1.7.2.
```

- **`1caff77d5dd4597756dc0e5b4a54a39c`** → stage 1 is in place. Proceed to 9.2.
- **`a117cf494ea53d18a0fd1ffbc1b2310b`** → stage 1 was rolled back or reverted on the VM.
  **STOP.** Do not write: the staged file is cut from stage 1 and applying it here would
  re-introduce stage 1's changes without anyone deciding to. Find out who rolled it back
  and why first.
- **anything else** → Tars edited its own skill (it does that; SOUL rule 2). **STOP and
  resolve which side moved before writing**, per CLAUDE.md: diff live against stage 1, find
  out what Tars changed, and re-cut the stage-2 hunks on top of the real live file. Never
  overwrite an unexplained live file — that is how a self-authored skill edit gets silently
  destroyed.

**9.2 R2-completion gate.** Run **after** the WF6 R2 measurement run has completed, never
before it. Confirm completion rather than assuming it — read the cron store read-only
(verified executing 2026-08-11):

```bash
ssh gaetan@192.168.0.9 python3 - <<'PY'
import sqlite3
c = sqlite3.connect("file:/home/gaetan/.hermes/cron/executions.db?mode=ro", uri=True)
for r in c.execute("""SELECT job_id, source, status, started_at, finished_at
                      FROM executions ORDER BY started_at DESC LIMIT 5"""):
    print(r)
PY
```

The R2 row must read `status='completed'` **with a non-null `finished_at`**, and the R2 entry
must be in `status/lane-a.md`. Only then run 9.3.

**9.3 Back up the live file first.**

```bash
ssh gaetan@192.168.0.9 'cp -a ~/.hermes/skills/orchestration/engagement-checker/SKILL.md \
  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-gcn30-32-stage2-$(date -u +%Y%m%dT%H%M%SZ) \
  && ls -la ~/.hermes/skills/orchestration/engagement-checker/'
```

This backup is **stage 1** (`1caff77d…`), not the pre-fix base. Stage 1's own backup
(`SKILL.md.bak-gcn30-32-stage1-20260811T202839Z`) is the pre-fix base and is what section 5.6
restores. Two different rollbacks, two different backups — read the names.

**9.4 Transfer by tmp + mv, under the repo's flock discipline.** Never an in-place edit,
never a heredoc (a heredoc can be silently truncated — that is exactly how GCN-13 merged
six empty SKILL.md files).

```bash
STAGED=/home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix/status/staged/gcn30-32-engagement-checker-SKILL.md
cat "$STAGED" | ssh gaetan@192.168.0.9 'umask 077; cat > ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new'
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new'
# must equal d0adb8917fa7d793465bbf9a96b6bb87 BEFORE the mv. If it does not:
#   ssh gaetan@192.168.0.9 'rm ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new'
# and re-transfer. Never mv an unverified file, and never leave the stray .new behind.

ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "mv ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new \
  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md"'
```

Hermes live-reloads in ~30 s. **Wait at least 35 s after the `mv` before triggering,
measuring or reasoning about any run** — a run started inside the reload window may still be
executing 1.7.2, and its result is not evidence about stage 2 either way.
(`.wf3.lock` is the repo's documented serialisation point for `~/.hermes` edits; CLAUDE.md
mandates `.bak` first, then `flock`, and merge-never-append — this is a whole-file
replacement of a file whose pre-state was verified in 9.1, which satisfies the merge rule by
construction. That verification is what makes the whole-file shape safe; without 9.1 it is
the silent-revert trap of section 6.1.)

**9.5 Verify byte equality live vs staged.**

```bash
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  wc -l ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  grep -c "^version: 1.8.0" ~/.hermes/skills/orchestration/engagement-checker/SKILL.md'
md5sum "$STAGED"
# both md5 = d0adb8917fa7d793465bbf9a96b6bb87 ; 718 lines ; version 1.8.0. Any mismatch →
# roll back with section 9.7 before anything else, then re-cut and re-transfer. Do not
# commit the mirror.
```

**9.6 Mirror commit, at the REAL repo path.**

```bash
cd /home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix
cp "$STAGED" skills/orchestration/engagement-checker/SKILL.md
md5sum skills/orchestration/engagement-checker/SKILL.md   # d0adb8917fa7d793465bbf9a96b6bb87
git add skills/orchestration/engagement-checker/SKILL.md status/probes/gcn30-32-checker-fix.md \
        status/staged/gcn30-32-engagement-checker-SKILL.md
git commit -m "fix(engagement-checker): stage 2 semantics — statusType duplicate terminal, close-back guard, linear_issue provenance (GCN-30, GCN-32)"
git pull --rebase origin main && git push origin HEAD:main
```

`status/staged/…` is committed alongside it because every md5 assertion in sections 8 and 9
points at that artefact; leaving it untracked keeps the document's authority in a directory
git does not carry. (Deleting it in the same commit is the acceptable alternative — but then
say so, and the md5 assertions become historical.)

The path is `skills/orchestration/engagement-checker/SKILL.md` — the file's path relative
to `~/.hermes/skills/`, per SOUL rule 2 and the GCN-13 correction; stage 1 already landed at
exactly that path in `d5a051d`. **The commit is not optional**: a `hermes cron` job checks
the mirror invariant daily — job `45c5c8ba10ee`, "Skill mirror drift check", schedule
`0 18 * * *` (18:00 **Europe/Paris**, i.e. 16:00Z — the "18:00Z" shorthand in the lane brief
is Paris local, not UTC), delivering to Slack. An un-mirrored live edit will be reported as
drift.

**9.7 Stage-2 rollback — back to STAGE 1, not to the base.**

```bash
# VM: restore the backup taken in 9.3 — by copy-then-mv, never a cp over the live file
# (a live-reload landing mid-copy reads a truncated skill; same reason 9.4 uses tmp+mv).
# Resolves the backup in-shell — no placeholder to hand-substitute at 07:00.
ssh gaetan@192.168.0.9 'D=~/.hermes/skills/orchestration/engagement-checker; \
  B=$(ls -t $D/SKILL.md.bak-gcn30-32-stage2-* | head -1); echo "restoring from: $B"; \
  cp -a "$B" $D/SKILL.md.restore && md5sum "$B" $D/SKILL.md.restore'
# must read 1caff77d5dd4597756dc0e5b4a54a39c BEFORE the mv
ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "mv ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.restore \
  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md"'
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  wc -l ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; \
  rm -f ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new'
# must be back to 1caff77d5dd4597756dc0e5b4a54a39c and 716 lines — STAGE 1, version 1.7.2

# repo: revert the stage-2 mirror commit only
cd /home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix
git revert --no-edit <stage-2-mirror-commit-sha>
git pull --rebase origin main && git push origin HEAD:main
```

To roll back **both** stages, run this first and then section 5.6.

**9.8 Ticket closure — only after P1 has passed at apply time AND P2–P5 have been
recorded on the first in-window run after it.** P2–P5 do not exist until that run has
happened; closing on 9.1–9.7 alone closes both tickets on unverified work. GCN-30 and
GCN-32 → Done, with a comment naming this file
(`status/probes/gcn30-32-checker-fix.md`), the new version `1.8.0`, the live md5, and the
acceptance-probe result lines from **section 8**. State in the GCN-30 comment that the nag
loop and the wrong write are both closed but the **coverage** miss (a terminal transition
that never routes) is not — that is the deferred sweep in section 10(a). Stage 1 is a
GCN-31-adjacent plumbing fix and is reported there, not as GCN-30/32 closure evidence.

### 9.9 GCN-25 hand-revert — **ONLY IF R2 CORRUPTED IT**

R2 runs before stage 2, so the wrong write is still possible on it. If the R2 run fires the
close-back and writes Done over GCN-25 (detect with P2: `statusType` becomes `completed`, a
Done entry appended to `stateHistory`), put it back:

```
# from a Claude Code session on cooper:
#   mcp__claude_ai_Linear__save_issue(id="GCN-25", state="0ea8bdc6-215f-48d3-b228-779422c6b03e")
# from Tars on the VM (Hermes):
#   mcp__linear__save_issue(id="GCN-25", state="0ea8bdc6-215f-48d3-b228-779422c6b03e")
```

`0ea8bdc6-215f-48d3-b228-779422c6b03e` is the Duplicate state on team GCN. Pass **`id` and
`state` only** — never `labels`, not even empty, or the whole label set is replaced.
Confirm on the returned payload that `statusType == "duplicate"` ("no error" is not
success). Then clear the wrongly-latched `linear_issue.closed_at` on item
`slack:CC397R0HY:1786458544.180249`, or the close-back can never run for it again.
**Do not run this unless P2 shows the corruption.**

That second half is a write to `~/.hermes/state/engagement-checker.json` — the same file a
running checker rewrites wholesale — so it takes the same discipline as any `~/.hermes` edit:
`.bak` first, then `flock`, then tmp + atomic replace, and **never while a run is in flight**
(check `~/.hermes/logs/agent.log` for an active session first). Transfer the script the way
9.4 transfers the skill rather than nesting quotes inside `flock -c`:

```bash
ssh gaetan@192.168.0.9 'cp -a ~/.hermes/state/engagement-checker.json \
  ~/.hermes/state/engagement-checker.json.bak-gcn30-32-revert-$(date -u +%Y%m%dT%H%M%SZ)'

ssh gaetan@192.168.0.9 'umask 077; cat > /tmp/clear-closed-at.py' <<'PY'
import json, os
p = "/home/gaetan/.hermes/state/engagement-checker.json"
d = json.load(open(p))
d["items"]["slack:CC397R0HY:1786458544.180249"]["linear_issue"].pop("closed_at", None)
t = p + ".tmp"
json.dump(d, open(t, "w"), ensure_ascii=False, indent=2)
os.replace(t, p)
PY

ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "python3 /tmp/clear-closed-at.py"'
ssh gaetan@192.168.0.9 'cat ~/.hermes/state/engagement-checker.json' \
| jq '.items["slack:CC397R0HY:1786458544.180249"].linear_issue'   # closed_at must be gone
```

(9.4's "never a heredoc" bans heredoc-transferring the 718-line skill, whose truncation would
be silent; this is a 7-line script and its effect is verified by the read-back on the last
line. `indent=2` matches the live file's own formatting, verified read-only, so the rewrite
is not a whole-file reformat.)

Never executed. Restore the `.bak` if the replace errors; the file is the checker's entire
queue.

---

## 10. FOLLOW-UPS — MIGHT-DO candidates for the coordinator

**This lane files no tickets.** Both are candidates, not commitments.

**(a) The deferred per-run read-back sweep for non-routed filed items.** On each run, for
every item carrying `linear_issue`, non-terminal, that did not route this run: one
`mcp__linear__get_issue` and reconcile on the flat `statusType`. Measured cost on tonight's
queue: **4 extra reads/run** on the tight definition (items with `linear_issue` whose item
status is `open`/`waiting`/`snoozed`: GCN-12, GCN-15, GCN-21, GCN-24), **6** on the
`closed_at` definition (those four plus GCN-23 and GCN-25) — minus any that routed, so
fewer in practice. Not explicitly bounded by anything in the file; the file's cap idiom is
line 586 ("**File at most 10 issues per run.**"), and a sweep would need a cap written
in that idiom — e.g. read at most 20 per run, oldest `updated_at` first, remainder named in
§8's `Coverage:` line. That wording is a proposal, not a quote from the skill.
**Why deferred:** nothing is lost — the cursor is held at `2026-08-11T14:30:26Z`, so
GCN-21 and GCN-25 are both inside the next successful window and reconcile on the first
healthy collector run; the measured cause of the miss is the blocked collector, which stage 1
attacks directly; and tomorrow morning is a **measurement run** — adding 4–6 new
Linear calls and a new fail-open path to the critical path the night before makes the
measurement harder to read, not easier. Known risks if it is built: a failed read must fail
open (never read as "not terminal"); it is on the native `statusType` path, never
`state.type`; and the reopen rule must stay keyed on routed candidates or hand-closed
issues start ping-ponging.

**(b) The chronically approval-gate-blocked collector — partly addressed by stage 1.** The
read-only Linear collector was blocked on **7 runs today whose collection actually failed**
(11:00, 11:30, 12:00, 12:30, 13:00, 15:30, 17:00 Paris) plus the 20:08 manual run, holding
the Linear cursor at `2026-08-11T14:30:26Z` since the 16:30 Paris run — that run shows the
block too but is the day's last **successful** collection (it holds
`last_success 2026-08-11T16:30:26+02:00`). Two distinct block messages appear:
`BLOCKED: Command flagged as dangerous (script execution via -e/-c flag)` on the aborted
`direct` runs, and `BLOCKED: Command timed out without user response` (302.81 s) on the
manual run. **This is the operational cause behind both misses in GCN-30 and behind
GCN-21's stale item.** Stage 1 rewrites the skill's invocation shape to one the guard lets
through, which is the SKILL.md-side half of the fix; the config-side half (why a read-only
collector is guard-flagged at all, and why `cron_mode` surfaces as `pending_approval`) is
untouched and still belongs to GCN-31 ("collector path: dangerous-command guard reroutes +
google-workspace setup failure"). **Whether stage 1's half is enough is measured by P5b in
section 8**, not asserted here.

---

## 11. WHAT IS NOT VERIFIED

Stated plainly, so nothing here is read as measured.

1. **Stage 1 is applied and byte-verified, but nothing about it is behaviourally
   exercised.** Every claim in section 5 is a hash, a file listing, a commit or a
   `systemctl` field — all measured. **No cron run has yet used the new §4 shape or the
   venv interpreter on this skill.** "The collector now runs from cron" is **EXPECTED, not
   measured**: the first in-window run is the test, and P5b (a Linear cursor moving past
   `2026-08-11T14:30:26Z`) is its evidence. Until then, stage 1 is a correctly transferred
   piece of prose whose effect on a model's behaviour is inferred from the GCN-31 lane's
   measurement of a *different* run, not from this skill.
2. **Stage 2 is entirely unexercised.** No run was executed with it, no live file carries
   it, no Linear write was made, no VM write of any kind. Every verdict in section 7 is
   static review or an on-paper walkthrough of the staged text.
3. **No collector payload for a `duplicate`-typed issue has ever been captured.** H2/H3
   widen the **collector** path (`state.type`) on the strong inference that §4's
   `state { id name type }` selection returns `"duplicate"` — supported by GCN-25's
   `stateHistory` carrying exactly that object from a native read, but one measurement
   short of proven. H4's backstop keys on the native `statusType`, where the evidence is
   direct. The first successful run will contain such a payload; logging it costs nothing.
4. **The BLOCKER rewrite in H5 was not re-walked by any reviewer.** The semantics lens
   walked S5 against the pre-adjudication draft; the clause it quoted was replaced. The
   replacement is the adjudicator's own text.
5. **The staged file's effect on a real run is unknown.** It is prose consumed by a model,
   not code: "the rule now forbids X" is a claim about the text, not about the run. The same
   caveat applies to stage 1's §4 paragraph.
6. **Stage 2 does not fix GCN-21**, and does not close the coverage gap at all — only
   the wrong-write hole and the routed nag loop. GCN-21's item is `waiting`, so the
   close-back never runs and step 2's backstop is never reached.
7. **An unrecognised eighth `statusType` still falls through step 2 into the write.** H2
   declares the set closed as intent but gives the run no fail-closed instruction.
   Deferred deliberately (a fail-closed default could equally stall a legitimate close);
   it requires Linear to add an enum value to the workspace.
8. **§7c's heading, "the only place `linear_issue` is written", remains factually wrong** —
   §5 writes the field four times (priority refresh, two `closed_at` latches, one
   `closed_at` delete). Pre-existing, out of scope, worked around by governing the value
   rather than the location. Reported, not fixed.
9. **A `closed_at` latched by step 1/2/4 on an item whose status-history reason is not
   `linear:*` can never be cleared**, because §5's reopen rule requires a `linear:*` reason.
   Pre-existing; H4 widens the set of items that can reach it. Items expire at 14/30 days.
10. **Step 2's terminal skip leaves the item's local status disagreeing with Linear** in one
    case (an item `done` from a Slack reason whose issue is `duplicate` stays `done` where
    §5's mapping says `dismissed`). Left as designed — mutating item status from a *read*
    rather than a routed event is the deferred sweep, and the held cursor guarantees the
    routed candidate applies the mapping on the first healthy run.
11. **The staged md5 `d0adb8917fa7d793465bbf9a96b6bb87` is only valid for the file as it
    stands now.** Any further edit invalidates every md5 assertion in sections 8 and 9.
12. **Section 9's stage-2 apply protocol has never been executed; section 8's probe was
    dry-run executed read-only and nothing more.** On 2026-08-11 every section-8 command was
    run read-only against the live VM and Linear (real output pasted inline, marked
    *measured*), which is how the broken `while read … ssh` loop and four vacuous-pass paths
    were found and corrected — but a dry run proves the commands execute and discriminate,
    **not** that the fix works. Section 9's transfer, rollback, mirror commit and 9.9
    hand-revert are unexercised: they were composed against measured facts and audited by
    reading, never run. Section 5.6's stage-1 rollback is equally unexercised — its *forward*
    counterpart ran successfully tonight, which is the closest thing to evidence it has.
