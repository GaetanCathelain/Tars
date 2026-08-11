# GCN-31 — engagement-checker collection: liveness, guard rerouting, the google-workspace interpreter repair (applied), cron-stable collector shape (2026-08-11)

GCN-31 asked four things about the Tars engagement-checker: **(Q1)** is the
queue's email collection actually live or historical residue; **(Q2)** can the
cron dangerous-command guard reroute a collector into a *silently truncated*
collection that the run still reports as healthy; **(Q3)** what is the root cause
of the google-workspace collector's dependency failure and what is the minimal
repair; **(Q4)** what command shape is stable under the guard from cron, with the
`SKILL.md` §4 replacement text to encode it.

Evidence was gathered read-only on `gaetan@192.168.0.9` on **2026-08-11 between
~19:10 and ~19:2x UTC** (anchor `date -u` = `Tue Aug 11 19:10:42 UTC 2026`;
latest logged instant anywhere in the evidence: `agent.log` tail
`2026-08-11 19:11:58,643`). One repair followed at **19:46–19:47 UTC** — a
two-line edit to `~/.hermes/skills/productivity/google-workspace/SKILL.md`, the
**only** write this lane made anywhere (§3.2, §6). **Every timestamp below is
UTC.** The VM clock is UTC; state files written by the skill carry `+02:00`
(Europe/Paris) and are converted inline. No checker, cron or chat run was
triggered (§6).

**Verdicts:**

> **Q1 — LIVE BUT NOT AUTONOMOUS: email collection is genuinely live over IMAP/himalaya — 22 new IMAP UIDs (`imap:35832`…`imap:35853`) and one new email item were ingested at 2026-08-11T18:09:11 UTC and persisted at 18:22:35 UTC — but that ingest came from the interactive session `20260811_180849_14f453`, so if LIVE must mean scheduled, unattended collection the answer is STALLED SINCE 2026-08-11T15:00:32 UTC: every cron engagement-checker run after that (17:29:32 and 17:42:44 UTC) ended `[SILENT]` without reaching any collector.**

**Q2 UNPROVABLE-FROM-LOGS** (split: durable state provably safe for the one
traced run, visibility unprovable), **Q3 APPLIED-AND-PROVEN BUT INCOMPLETE BY
DESIGN** (google-workspace fixed and exercised; the line the cron *email*
collector actually runs belongs to a sibling lane and is untouched),
**Q4 STABLE-SHAPE-PROVEN — empirical for 2026-08-11's rule set.**

| Question | Verdict | One-line basis |
|---|---|---|
| Q1 — email collection LIVE or residue? | **LIVE (manual) / STALLED SINCE 15:00:32 UTC (cron)** | 22 new IMAP UIDs + one new email item ingested 18:09:11, persisted 18:22:35 — by the interactive session `20260811_180849_14f453`; both later cron runs ended `[SILENT]` reaching no collector (§1) |
| Q2 — can guard rerouting silently truncate? | **UNPROVABLE-FROM-LOGS** | durable state provably safe for the one traced failure; the *disclosure* half is structurally unloggable — `agent.log` never records a response body (§2). Structural, and separate: the §4 collector **as written cannot run from cron at all** (§2.4c) |
| Q3 — google-workspace root cause + repair | **APPLIED-AND-PROVEN, INCOMPLETE BY DESIGN** | 2 of 3 invocation strings converted to the venv python and proven green (`Dependencies already installed.`/`OK`/`AUTHENTICATED`, all exit 0); the third — `engagement-checker/SKILL.md:139`, the one the cron email collector runs — is a **sibling lane's file, untouched** (§3) |
| Q4 — cron-stable collector shape | **STABLE-SHAPE-PROVEN** | `write_file` + `[VAR=…] python3 /abs/path.py` ran to `exit_code: 0` in a real cron run under `cron_mode: deny` (§4) — empirical for 2026-08-11's rule set; Tirith's binary rule set was not enumerated, so this is not a proof against future rules (§4.5) |

**Cross-cutting structural finding (Q2 §5c / Q4 §1.3).** Under the default
`approvals.cron_mode: deny`, the §4 Linear collector **as currently written
cannot run from cron at all**: it exists only as a fenced Python block inside
`SKILL.md` (`scripts/` holds nothing but an empty `tests/`), and both of its only
two shell shapes — `python3 -c` and heredoc — are unconditionally denied. Every
future cron engagement-checker run therefore registers a Linear source failure.
Non-lossy per run (the cursor is held), but the window widens each cycle toward
the `MAX_ISSUE_PAGES=4 × ISSUE_PAGE=50` cap.

---

## 1. Q1 — VERDICT: LIVE BUT NOT AUTONOMOUS

> **Q1 — LIVE BUT NOT AUTONOMOUS: email collection is genuinely live over IMAP/himalaya — 22 new IMAP UIDs (`imap:35832`…`imap:35853`) and one new email item were ingested at 2026-08-11T18:09:11 UTC and persisted at 18:22:35 UTC — but that ingest came from the interactive session `20260811_180849_14f453`, so if LIVE must mean scheduled, unattended collection the answer is STALLED SINCE 2026-08-11T15:00:32 UTC: every cron engagement-checker run after that (17:29:32 and 17:42:44 UTC) ended `[SILENT]` without reaching any collector.**

Read both halves. The queue is not historical residue — but nothing scheduled
has collected since 15:00:32 UTC.

### 1.1 The before/after pair is exact

`~/.hermes/state/` happens to hold a backup taken by the run in question at its
own start, which makes the diff exact rather than inferred. From `ls -la`:

```
-rw-------  1 gaetan gaetan 111246 2026-08-11 18:22:35.815828422 +0000 engagement-checker.json
-rw-rw-r--  1 gaetan gaetan 107964 2026-08-11 15:05:09.026128993 +0000 engagement-checker.json.bak-manual-20260811T180911Z
```

The backup's filename stamp `20260811T180911Z` = **18:09:11 UTC** matches the
live file's `last_completed_run` to the second, and the copy still carries the
*previous* cursor:

```
bak-manual-20260811T180911Z   last_completed_run / email.cursor / email.last_success:
                              "2026-08-11T17:00:32+02:00"   (= 15:00:32 UTC)
engagement-checker.json       same three fields:
                              "2026-08-11T20:09:11+02:00"   (= 18:09:11 UTC)
```

Everything that differs between the two files was therefore added by the
18:09:11-UTC run.

### 1.2 The diff is real new email material, not a cursor bump

```
$ jq '[.sources.email.seen[]|select(startswith("imap:"))]|max'  bak-manual-20260811T180911Z
imap:35802
$ jq '[.sources.email.seen[]|select(startswith("imap:"))]|max'  engagement-checker.json
imap:35853
$ comm -23 <live item ids> <bak item ids>
email:vercel-budget-2026-08-11
slack:C09H8CHTHMZ:1786460726.573869
```

`email.seen` gained **22 contiguous new IMAP UIDs, `imap:35832` … `imap:35853`**,
all strictly above the previous high-water mark `imap:35802` (2026-08-10). Dedup
keys are only written when a collector enumerates messages. One new
email-sourced queue item was created, with message-specific content and an IMAP
evidence handle inside the new UID range — excerpted from the live state file
(values verbatim, fields elided):

```json
{"id":"email:vercel-budget-2026-08-11","source":"email",
 "title":"Review Vercel usage after the on-demand budget reached 100%",
 "snippet":"Vercel reports that MobileClub used 100% of its $200.02 on-demand spend budget and asks the team to review usage or adjust spend controls.",
 "asker":"Vercel","evidence_handle":"imap:35847",
 "created_at":"2026-08-11T18:32:03+02:00",
 "status_history":[{"at":"2026-08-11T20:09:11+02:00","status":"open",
   "reason":"email: fresh billing threshold alert with an explicit usage-review action"}]}
```

`created_at` is the **message** time (16:32:03 UTC); `status_history[0].at` is
the **ingest** time (18:09:11 UTC) — the same ordering every other email item
shows (e.g. `email:19ff1193a65b10c3`: created 13:53:20 UTC, opened 14:00:41
UTC). A message that arrived at 16:32 UTC cannot be residue of a run whose last
email touch was 15:00:32 UTC.

*(Evidence-file erratum, recorded so nobody "corrects" it downstream:
`A-state-cursors.md:116-117` calls this tail "the last 57 entries" while naming
the range `imap:35832`..`imap:35853`. The raw array in the same file
(`:1266-1287`) holds **22** entries. 22 is the number.)*

### 1.3 Which process did it — an interactive session, not cron

`agent.log` shows **no cron engagement-checker invocation at 18:09**. The last
two cron checker runs both ended `[SILENT]` without reaching any collector:

```
2026-08-11 17:29:32,497 INFO cron.scheduler: Running job 'Gaetan engagement checker final pass' (ID: 759e08c598e3)
2026-08-11 17:30:28,484 INFO cron.scheduler: Job '759e08c598e3': agent returned [SILENT] — skipping delivery
2026-08-11 17:42:44,464 INFO cron.scheduler: Running job 'Gaetan engagement checker final pass' (ID: 759e08c598e3)
2026-08-11 17:43:54,456 INFO cron.scheduler: Job '759e08c598e3': agent returned [SILENT] — skipping delivery
```

The 18:09 write came from the interactive session `20260811_180849_14f453`:

```
2026-08-11 18:09:01,214 INFO [20260811_180849_14f453] agent.tool_executor: tool skill_view completed (0.04s, 71434 chars)
2026-08-11 18:09:11,102 INFO [20260811_180849_14f453] agent.tool_executor: tool terminal completed (0.07s, 100 chars)
2026-08-11 18:10:15,972 WARNING [20260811_180849_14f453] agent.tool_executor: Tool terminal returned error (0.1... [himalaya]
2026-08-11 18:22:35,902 INFO [20260811_180849_14f453] agent.tool_executor: tool execute_code completed (0.12s, 254 chars)
```

The 18:22:35,902 `execute_code` matches the live file's mtime
(`18:22:35.815828422`) to the millisecond band. Ingest window: **18:09:11 →
18:22:35 UTC**.

### 1.4 The mechanism: the Gmail-API path is dead, email now arrives over IMAP

Stated plainly because it is the finding behind both Q1 and Q3: **the
google-workspace (Gmail API) collector cannot run — that is §3's failure — and
email is reaching the queue through `himalaya` over IMAP instead.** No
`mcp__email__*` server exists (`MCP: registered 54 tool(s) from 3 server(s)` =
notion/slack/linear only); all email access is `terminal` → `himalaya`.

The fingerprint is the `email.seen` id-format transition: bare and
`email:`/`email-message:`-prefixed 16-hex **Gmail** ids up to
`email-message:19ff15471a42ef74`, then a pure **`imap:<UID>`** tail of 22
entries. The format changes exactly where the transport changed.

### 1.5 What would falsify Q1

1. **Proof the 22 IMAP UIDs were never fetched from the mailbox.** The `terminal`
   command text is truncated out of `agent.log` by design and no transcript for
   `14f453` survives in `~/.hermes/sessions/`. If `himalaya envelope list` on the
   `tars` account shows the mailbox UID range does not cover 35832–35853, or no
   message at UID 35847 matching the Vercel alert, the verdict flips to
   **STALLED SINCE 2026-08-11T15:00:32 UTC** outright. *(Not run — mailbox reads
   were outside the read-only-file remit.)*
2. **A later state write that reverts this.** If the email cursor drops back
   below `2026-08-11T20:09:11+02:00`, the advance was a transient hand-edit.
3. **Redefining LIVE as "autonomous"** — already carried in the verdict above.
   The last cron-driven cursor advance is `2026-08-11T17:00:32+02:00`, written at
   15:05:09 UTC by the 15:00:03 UTC `final pass` run. Every cron checker run
   after that reached no collector at all.

---

## 2. Q2 — VERDICT: UNPROVABLE-FROM-LOGS (split; the visibility half cannot be settled from logs at all)

The question has two halves and the evidence answers only one.

- **Durable-state half — positively SAFE for the one traced failure.** The Linear
  cursor and `last_success` were held unchanged across a run whose Linear
  collector was blocked, and an explicit `last_failure_notice.linear` was
  written. The delta is **deferred, not lost**.
- **Visibility half — UNPROVABLE.** Whether the run's *delivered report* named
  the failure cannot be established: `agent.log` records `response_len` and
  delivery target, **never the response body**. The one measurable proxy points
  the **wrong** way (§2.4b).

PROVEN-SAFE would require positive evidence that the §8 coverage gate *fires*,
and the gate's firing is structurally unloggable. Hence UNPROVABLE-FROM-LOGS —
not "safe", not "unsafe".

### 2.1 The traced BLOCKED → reroute sequence

**Cron run A** (`cron_759e08c598e3_20260811_192932`), block then reroute — the
block line is truncated in the source log, marked as such:

```
2026-08-11 17:30:08,909 WARNING [cron_759e08c598e3_20260811_192932] agent.tool_executor: Tool terminal returned error (0.03s): {"output": "", "exit_code": -1, "error": "BLOCKED: Command flagged as dangerous (script execution via -e/-c flag) but cron jobs run without a user present to approve it. Find an alternative approach t   [TRUNCATED IN SOURCE LOG]
2026-08-11 17:30:08,945 INFO … tool search_files completed (0.03s, 83 chars)
2026-08-11 17:30:09,024 INFO … tool tool_describe completed (0.07s, 2953 chars)
2026-08-11 17:30:18,977 INFO … tool terminal completed (0.09s, 115 chars)
2026-08-11 17:30:19,208 WARNING … Tool terminal returned error (0.23s): {"output": "/usr/bin/python3: No module named pip\nerror: The interpreter at /usr is externally managed…
2026-08-11 17:30:19,435 INFO … tool terminal completed (0.22s, 45 chars)
2026-08-11 17:30:19,543 INFO … tool read_file completed (0.10s, 82762 chars)
```

**Cron run B** (`cron_759e08c598e3_20260811_194244`) — three blocks, same
pattern: heredoc at 17:43:32,746; `-e/-c` at 17:43:49,019 and 17:43:49,200; the
reroutes are 133-char and 56-char terminal outputs.

**The reroute is not a narrower collection — it is zero collection.**
`mcp__linear__list_issues` appears in neither run. The reroute outputs (115 / 45
/ 133 / 56 chars) are two to three orders of magnitude below the 3837-char
`mcp__linear__list_issues` and 17–38 kchar Slack collector results seen in the
same evening's daily-work-brief run. What the agent did after each block was
introspection and probing (`search_files`, `tool_describe`, `read_file`), not
collection.

The blocked scope is not recoverable from the log but is bounded by the skill's
own contract: Linear, window `(LINEAR_CURSOR − 5 min, LINEAR_RUN_END]`, queries
`AssignedIssues` (assignee isMe) + `TeamIssues` (GCN team `81e7b769-…`),
`ISSUE_PAGE=50`, `MAX_ISSUE_PAGES=4`, `MAX_CANDIDATES=100` shared.

**Manual run 18:08:49 — Linear blocked by the *other* guard branch**, email and
slack collected:

```
2026-08-11 18:16:15,069 WARNING [20260811_180849_14f453] agent.tool_executor: Tool terminal returned error (302.81s): {"output": "", "exit_code": -1, "error": "BLOCKED: Command timed out without user response. The user has NOT consented to this action. Do NOT retry this command, do NOT rephrase it, and do NOT attempt…
```

302.81 s = `approvals.timeout` 300 s — the gateway-approval branch, not
`cron_mode: deny`. Same net effect: the §4 Linear collector never ran. **This is
also the only datapoint the evidence has on the visibility half, and it is a
lucky one: it was a CLI session, so Gaetan saw its 1020-char reply. A cron run in
that state would have had no witness at all.**

### 2.2 What actually landed in durable state

```
== engagement-checker.json.bak-manual-20260811T180911Z   (state as of 18:09:11 UTC)
 last_completed_run: 2026-08-11T17:00:32+02:00
 last_failure_notice: {}
   slack  cursor= 2026-08-11T17:00:32+02:00  last_success= 2026-08-11T17:00:32+02:00  seen=500
   email  cursor= 2026-08-11T17:00:32+02:00  last_success= 2026-08-11T17:00:32+02:00  seen=500
   linear cursor= 2026-08-11T14:30:26Z       last_success= 2026-08-11T16:30:26+02:00  seen=193

== engagement-checker.json                               (live, written 18:22:35 UTC)
 last_completed_run: 2026-08-11T20:09:11+02:00
 last_failure_notice: {"linear": {"at": "2026-08-11T20:09:11+02:00", "message": "Linear delta collection could not complete because the read-only collector command was blocked by the execution approval gate."}}
   slack  cursor= 2026-08-11T20:09:11+02:00  last_success= 2026-08-11T20:09:11+02:00  seen=500
   email  cursor= 2026-08-11T20:09:11+02:00  last_success= 2026-08-11T20:09:11+02:00  seen=500
   linear cursor= 2026-08-11T14:30:26Z       last_success= 2026-08-11T16:30:26+02:00  seen=193
```

Two hard conclusions:

1. **Cron runs A and B wrote nothing.** The 18:09:11 snapshot still carries
   `last_completed_run: 2026-08-11T17:00:32+02:00` (the 15:00 UTC run), while the
   scheduler logged `completed successfully` and `[SILENT]` with
   `response_len=8` — the literal 8-char token. **No `Coverage:` line fits in 8
   characters.**
2. **The manual run held the Linear cursor exactly as §1/§4 require** — cursor
   `2026-08-11T14:30:26Z`, `last_success 16:30:26+02:00`, both unchanged across a
   run whose fixed end was `20:09:11+02:00`, while email and slack advanced.
   Partial delta, run marked complete, **no data skipped**. That is the spec's
   designed behaviour, not a defect.

### 2.3 The gate wording, quoted from the live VM SKILL.md (read-only, not edited)

`~/.hermes/skills/orchestration/engagement-checker/SKILL.md` (69654 B, md5
`a117cf494ea53d18a0fd1ffbc1b2310b`, mtime 2026-08-11 15:31 UTC).

§1, line 113 — the cursor rule the manual run obeyed:

> Advance a source cursor to the run's fixed end timestamp only after all of that source's pages and candidate context were successfully processed. **On success, set that source's `last_success` to the run's fixed end timestamp. On failure, leave `last_success` unchanged, keep the cursor unchanged, persist the other state changes, and continue.** That single write is the only one the field has, and §8's mandatory "source failed this run" coverage condition is exactly the test `last_success < this run's end`.

§8, lines 683-693 — the gate, condition 5 (the source quote is itself partial):

> 5. a **source that failed this run**: its `last_success` is older than this run's fixed end timestamp. §1 writes that field on every success and nowhere else, so the test is exact rather than historical. For Linear it means a held cursor behind a run that otherwise looks healthy. …

§8, line 695 — the rate limit and its exemption:

> **The seven states above are EXEMPT: they are reported on every delivering run, however recently they last appeared, and `last_failure_notice` is neither consulted nor written for them.**

**Applying condition 5 to the live state:** `linear.last_success =
2026-08-11T16:30:26+02:00` is older than the run's fixed end
`2026-08-11T20:09:11+02:00`. **Condition 5 is TRUE right now**, and remains true
for every subsequent run until Linear collection succeeds.

### 2.4 Three paths where a degraded run reads as healthy

None is proven to have *lost* data; all three make a degraded run look fine.

**(a) Zero-collection run reported as success.** Runs A and B: guard-blocked, no
collector call, no state write, `response_len=8`, scheduler `completed
successfully`. *Unproven part:* the log cannot distinguish this from a legitimate
exit at the workday gate or the single-writer lock, both of which SKILL.md
sanctions as a bare `[SILENT]`. Their blocked `python3 -c`/heredoc attempts plus
a `pip install` attempt are strong circumstantial evidence they were trying to
run the §4 collector, but the command text that would settle it is absent by
design.

**(b) The `last_failure_notice` misfile — the concrete suppression mechanism.**
The live state carries `last_failure_notice.linear` for a condition that is
exactly §8 state 5, which line 695 says is *"neither consulted nor written for"*.
Writing it means the run classified a must-never-hide state as an ordinary
rate-limited note. Any run within 4 h that consults that key by the stated rule
will **suppress the disclosure** while `last_completed_run` keeps advancing and
the Linear cursor keeps freezing. *Unproven part:* the suppression has not been
observed happening; only the misfiled key is observed.

**(c) Structural: the §4 collector cannot run from cron at all.** No on-disk
script (`.../engagement-checker/scripts/` holds only an empty `tests/`), the
collector is an inline fenced block, and both of its only two shell shapes
(`python3 -c`, heredoc) are denied under `cron_mode: deny`. Every future cron EC
run registers a Linear source failure. Non-lossy per run (cursor held), but the
window widens each cycle toward the `MAX_ISSUE_PAGES=4 × ISSUE_PAGE=50` cap —
i.e. it converts into §8 condition 4 (`truncated`) over time, which the same
misfiled rate-limit key could then suppress. **Q4's §4 replacement text is the
proposed fix for this leg — drafted, not applied, and guard-stable only against
2026-08-11's rule set.**

### 2.5 What would falsify Q2

**Flip to PROVEN-SAFE if:** a cron EC run is observed in which the Linear
collector is blocked AND the delivered message body is captured and contains a
`Coverage:` line naming the Linear source failure. Requires either logging the
agent's `final_response` (bounded, redacted) at `cron/scheduler.py:4183/:4197`
alongside the existing `response_len`, or reading the Slack `C0BP2GZUFSR`
transcript for the 15:05:34 UTC delivery and any future one.

**Flip to PROVEN-UNSAFE if:** a cron run shows `last_success` older than that
run's `last_completed_run` **and** the delivered body carries no `Coverage:`
line. The `last_failure_notice.linear` misfile makes this the expected outcome on
the next EC run inside the 4-hour window, so **the next scheduled cron run is the
natural test.**

---

## 3. Q3 — VERDICT: APPLIED AND PROVEN, AND INCOMPLETE BY DESIGN

> **Read this first.** The repair landed on `google-workspace/SKILL.md` only. The
> third invocation string —
> `~/.hermes/skills/orchestration/engagement-checker/SKILL.md:139`, `GAPI="python
> …/google_api.py"` — **is the line the cron email collector actually executes,
> and it is still bare `python`.** That file belongs to the sibling lane
> `gcn30-32-checker-fix`; this session never wrote it. **§3 email collection is
> therefore NOT yet repaired.** The one-line change was handed to the coordinator
> as text (§3.6).

### 3.1 Root cause

Three documented invocation strings in two live `SKILL.md` files point the Google
scripts at a bare `python`/`python3` that resolves to the PEP-668 system
interpreter, while the exact pinned dependencies already sit in the venv Hermes
itself runs under. The caller strings, verbatim as they were before the repair:

```
~/.hermes/skills/productivity/google-workspace/SKILL.md
  41:GSETUP="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/setup.py"
 154:GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"

~/.hermes/skills/orchestration/engagement-checker/SKILL.md   (§3 "Collect email deltas")
 139:GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
```

`/usr/bin/python` does not exist on this box (`command -v python` → rc=1), so the
literal string cannot run; the agent substitutes `python3` → `/usr/bin/python3`
(3.12.3, `EXTERNALLY-MANAGED`, no `pip` module). `setup.py` then installs into
`sys.executable` — twice, with no override available:

```python
137      # First choice: pip in the current interpreter. Works for most installs.
139          subprocess.check_call(
140              [sys.executable, "-m", "pip", "install", "--quiet"] + missing,
...
155      # needing pip present. Targeting sys.executable keeps us on the venv the
156      # script is actually running under, rather than guessing.
161                  [uv, "pip", "install", "--python", sys.executable, "--quiet"]
```

Whole-file read confirms **no `--python` flag** (argparse offers only
`--check/--check-live/--client-secret/--auth-url/--auth-code/--revoke/--install-deps`),
**no env-var override** (the file's only `os.environ` write is
`OAUTHLIB_RELAX_TOKEN_SCOPE` at line 421), **no venv creation or detection.**

Observed failures, verbatim (13 occurrences in `agent.log` between 2026-08-10
19:34 and 2026-08-11 18:10 UTC):

```
2026-08-11 17:30:19,208 WARNING [cron_759e08c598e3_20260811_192932] agent.tool_executor: Tool terminal
  returned error (0.23s): {"output": "/usr/bin/python3: No module named pip\nerror: The interpreter at
  /usr is externally managed, and indicates the following:\n\n  To install Python packages system-wide, try apt install…
2026-08-11 17:48:12,405 WARNING [cron_e231e5faf180_20260811_194748] agent.tool_executor: … "TASK=\nWORKSPACE=\n
  2026-08-11T19:48:12+02:00\n/usr/bin/python3: No module named pip\nerror: The interpreter at /usr is externally managed…
```

The deps are already present, at the exact pins, in the interpreter Hermes runs
as:

```
~/.hermes/hermes-agent/venv/bin/python -B -c "import google_api, setup; ..."
  → module import OK under /home/gaetan/.hermes/hermes-agent/venv/bin/python
  → missing pkgs: []          # setup._missing_required_packages()
google-api-python-client 2.194.0 / google-auth 2.55.1 / google-auth-oauthlib 1.3.1
google-auth-httplib2 0.3.1 / httplib2 0.32.0 / pyasn1 0.6.4   ← exact REQUIRED_PACKAGES pins
/usr/bin/python3 -c "import googleapiclient" → ModuleNotFoundError
```

```
$ cat ~/.local/bin/hermes
#!/usr/bin/env bash
unset PYTHONPATH
unset PYTHONHOME
exec "/home/gaetan/.hermes/hermes-agent/venv/bin/python" "/home/gaetan/.hermes/hermes-agent/hermes" "$@"
```

**Root cause, one sentence:** the documented invocation hands the scripts to a
bare `python`/`python3` that resolves to the externally-managed system
interpreter, while the exact pinned dependencies already sit in
`~/.hermes/hermes-agent/venv` — a caller/invocation defect, not a missing
dependency and not a broken script.

Two "obvious" alternatives were checked and rejected: there is no venv logic in
`setup.py` to point at; an env-var override traps on
`_missing_required_packages()` (line 110) checking `importlib.metadata` in the
*running* interpreter, not the install target, so install-into-B-while-running-as-A
re-detects "missing" forever; and pointing uv at a uv-managed standalone python
was measured to fail identically (`2026-08-11 18:10:07 UTC`: `error:
externally-managed-environment … This Python installation is managed by uv and
should not be modified.`).

### 3.2 What was applied — one file, two lines, 19:46–19:47 UTC

Applied to `~/.hermes/skills/productivity/google-workspace/SKILL.md` **only**,
lines 41 and 154, bare `python ` → `${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python `.
Verified twice: once by the applying agent, once independently by the
orchestrator via a fresh `ssh` read.

```diff
--- /home/gaetan/.hermes/skills/productivity/google-workspace/SKILL.md.bak-gcn31-20260811	2026-08-07 21:26:26.300792111 +0000
+++ /home/gaetan/.hermes/skills/productivity/google-workspace/SKILL.md	2026-08-11 19:46:43.816260965 +0000
@@ -38,7 +38,7 @@
-GSETUP="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/setup.py"
+GSETUP="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/setup.py"
@@ -151,7 +151,7 @@
-GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
+GAPI="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
```

| file | md5 | size | mode |
|---|---|---|---|
| `…/google-workspace/SKILL.md` (after) | `2255d9b1e7199890de89410edca1edec` | 12883 B | 664 |
| `…/google-workspace/SKILL.md.bak-gcn31-20260811` (backup) | `96a0fbd4495db33325ee7f7683a711f4` | 12779 B | 664 (preserved by `cp -p`) |
| `~/.hermes/google_token.json.bak-gcn31-20260811` (token backup, contents never read) | — | 1153 B | 664 |

Leftover check after the edit: `grep -n '^GSETUP="python \|^GAPI="python '`
exits 1 — **no bare-python invocation remains in that file**;
`grep -c 'hermes-agent/venv/bin/python'` = 2.

No `flock` was needed (neither `config.yaml` nor `.env`), no service restart
(skills are read per request), no install, no venv creation, no apt, no global
site-packages.

### 3.3 Proof — three checks, all green

```
$ ~/.hermes/hermes-agent/venv/bin/python ~/.hermes/skills/productivity/google-workspace/scripts/setup.py --install-deps; echo "exit=$?"
Dependencies already installed.
exit=0
```

The literal `Dependencies already installed.` with **zero pip/uv output** is the
proof that `install_deps()` short-circuited at line 131 — nothing was installed
or downloaded.

```
$ cd ~/.hermes/skills/productivity/google-workspace/scripts && \
  ~/.hermes/hermes-agent/venv/bin/python -B -c \
  "import sys;sys.path.insert(0,'.');import google_api;from googleapiclient.discovery import build;print('OK')"; echo "exit=$?"
OK
exit=0
```

```
$ ~/.hermes/hermes-agent/venv/bin/python ~/.hermes/skills/productivity/google-workspace/scripts/setup.py --check; echo "exit=$?"
AUTHENTICATED: Token refreshed at /home/gaetan/.hermes/google_token.json
exit=0
```

**The open credential question is closed.** Before this repair, `check_auth()`
died inside `_ensure_deps()` before it ever tested the token against Google, so
the stored OAuth material had never been exercised at all. It has now been
exercised end-to-end: status word `AUTHENTICATED` (not `AUTHENTICATED (partial)`,
not `NOT_AUTHENTICATED`), exit 0, and the live call refreshed and rewrote the
token as predicted (`google_token.json` mtime moved 14:30:27 → 19:46:59 UTC; the
pre-refresh copy is preserved in the `.bak`). No credential value was read,
printed or copied anywhere — paths, sizes and modes only.

### 3.4 Negative control — isolates the interpreter, but is not a byte-identical reproduction

```
$ /usr/bin/python3 ~/.hermes/skills/productivity/google-workspace/scripts/setup.py --install-deps; echo "exit=$?"
/usr/bin/python3: No module named pip
Installing Google API dependencies...
ERROR: Failed to install dependencies: Command '['/usr/bin/python3', '-m', 'pip', 'install', '--quiet', …]' returned non-zero exit status 1.
exit=1
```

Same script, same args, only the interpreter differs → exit 1 under system python
vs exit 0 under the venv python. **Reported honestly: this is not identical to
the cron-observed failure.** The cron failures carried the
`error: The interpreter at /usr is externally managed…` line; this control does
not, because that line comes from the `uv pip install` fallback (`setup.py:157`,
`shutil.which("uv")`) and `uv` is not on `PATH` in a non-interactive ssh shell —
so the fallback never fired and the script errored out one step earlier. Same
failure class, same exit code, nothing installed; it isolates interpreter
selection as the only delta, but it does not reproduce the original byte-for-byte.

### 3.5 Boundary — what was NOT touched

`~/.hermes/skills/orchestration/engagement-checker/SKILL.md` was never opened for
write. Re-checked after the repair at 19:47:06 UTC:

```
a117cf494ea53d18a0fd1ffbc1b2310b  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
… | 2026-08-11 15:31:33.833054684 +0000 | 69654 | 600
```

md5, size (69654 B), mtime (15:31:33 UTC) and mode (0600) all predate this
session and are byte-identical to baseline. Nothing else was touched either: no
`config.yaml`, no `.env`, no `SOUL.md`, no install, no venv creation, no restart,
no queue/cron/run-history write, no `sops`. No new `__pycache__` writes — all
`.pyc` mtimes still Aug 7 / Aug 10.

Rollback, if ever needed (not needed; recorded for completeness):

```bash
mv -f ~/.hermes/skills/productivity/google-workspace/SKILL.md.bak-gcn31-20260811 \
      ~/.hermes/skills/productivity/google-workspace/SKILL.md
# only if the refreshed token must be reverted:
mv -f ~/.hermes/google_token.json.bak-gcn31-20260811 ~/.hermes/google_token.json
```

### 3.6 The remaining third of the repair — sibling lane, handed over as text

`~/.hermes/skills/orchestration/engagement-checker/SKILL.md:139` still reads:

```
139:GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
```

**That is the line the cron email collector runs.** Until it changes, the §3
email leg of the checker stays broken exactly as measured. The change is one
string, identical in form to the two already applied:

```
GAPI="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
```

**For the lane that owns that file (`gcn30-32-checker-fix`), not for this lane.**
This session wrote nothing under `~/.hermes/skills/orchestration/**` and did not
touch the repo mirror. That file **is** mirrored here at
`skills/orchestration/engagement-checker/SKILL.md` and is currently byte-identical
to the live copy (md5 `a117cf494ea53d18a0fd1ffbc1b2310b`, 69654 B), so the same
one-line diff is for the executing lane to land by branch → PR → squash-merge.

`skills/productivity/google-workspace/SKILL.md` is **not** mirrored (no
`skills/productivity/` in this repo): it is a vendored upstream skill, so this
edit is VM-only and a `hermes` upgrade may re-copy the file from
`~/.hermes/hermes-agent/skills/…` and **silently revert it** — same class as the
`tool_progress_dm` patch (§5). Do **not** edit `~/.hermes/hermes-agent/skills/…`
(upstream clone; not what the agent loads).

---

## 4. Q4 — VERDICT: STABLE-SHAPE-PROVEN (empirical for 2026-08-11's rule set)

**The stable shape:** `write_file` the collector to an absolute path, then run it
as `[VAR=… ] python3 /abs/path/collector.py` — **no `-c`, no heredoc, no
pipe-into-interpreter, and no `rm` of an absolute path afterwards.**

### 4.1 Why each rule family misses that shape

Derived from `~/.hermes/hermes-agent/tools/approval.py`. In a cron run, a command
is safe iff it clears *both* `detect_dangerous_command` (47 `DANGEROUS_PATTERNS`
regexes + `_execution_flag_findings`) **and** Tirith; there is no third gate and
no allowlist the shape needs. The decisive line is the exec-flag scanner's loop
(`:1520-1558`):

```python
for token in args:
    ...
    if token == "--": break
    if family != "powershell" and not token.startswith("-"):
        break        # <-- first non-flag token stops the scan
```

The first argument is a path, does not start with `-`, so the scan stops before
it can find any `-c`. Heredoc detection requires an argument starting with `<<`.
The pipe-to-shell regex family only models *shell* targets. Env-var prefixes and
`&&` chaining of two such invocations stay inside the shape — only
`DOCKER_HOST=`/`CONTAINER_HOST=` assignments are modelled by any pattern.
`~/.hermes/config.yaml` has **no `approvals:` key at all**, so `cron_mode` is at
its hardcoded default `"deny"` (`:2989`) — **the proof below was produced under
that default, and nothing here asks for a config change.**

### 4.2 The second-order trap — materializing the collector

The §4 collector is not a file on disk: it is a fenced `python` block inside
`SKILL.md` (lines 155–480 of 714), and
`~/.hermes/skills/orchestration/engagement-checker/scripts/` contains only an
empty `tests/` directory. Every *shell* way of materializing it is flagged:
`python3 - <<'PY'` (heredoc), `python3 -c '…write_text(…)'` (`-e/-c` flag),
`perl -ne … SKILL.md` (same rule, perl family), and piping the block into an
interpreter — Tirith **HIGH**, verbatim from the transcript store:

```
"description": "Security scan — [HIGH] Pipe to interpreter: python3 | python3: Command pipes local output into interpreter 'python3'. …", "pattern_key": "tirith:pipe_to_interpreter"
```

Even the *cleanup* is flagged: `rm -f /tmp/engagement-linear-collector.py`
matches `DANGEROUS_PATTERNS[0]` (`r'\brm\s+(-[^\s]*\s+)*/'`, "delete in root
path") because the path is absolute — confirmed in the store with
`"pattern_key": "delete in root path"` for a command ending
`…; rc=$?; rm -f /tmp/engagement-linear-collector.py; exit $rc`.

**Resolution: do not write the file from the shell at all.** The `write_file`
tool never reaches `check_all_command_guards` (that function is called from the
terminal/exec path only) and works in cron.

### 4.3 The log proof — a run of this shape that PASSED, under `cron_mode: deny`

All from cron session `cron_62e8cd9db637_20260811_140000` (hourly
engagement-checker), `agent.log` lines 15309–15390 region:

```
2026-08-11 12:02:58,129 WARNING [cron_62e8cd9db637_20260811_140000] agent.tool_executor: Tool terminal returned error (4.24s): {"output": "", "exit_code": -1, "error": "", "status": "pending_approval", "approval_pending": true, "command": "python3 -c 'impor…
2026-08-11 12:03:16,747 WARNING [cron_62e8cd9db637_20260811_140000] agent.tool_executor: Tool terminal returned error (4.83s): {"output": "", "exit_code": -1, "error": "", "status": "pending_approval", "approval_pending": true, "command": "python3 -c 'impor…
2026-08-11 12:03:50,476 INFO [cron_62e8cd9db637_20260811_140000] agent.tool_executor: tool write_file completed (0.06s, 446 chars)
2026-08-11 12:04:01,266 INFO [cron_62e8cd9db637_20260811_140000] agent.tool_executor: tool terminal completed (4.05s, 7725 chars)
```

**`agent.log` never records a successful command's text** — successes log only
duration and output length. The command behind the 12:04:01 success comes from
`~/.hermes/lcm.db`, the LCM transcript store, which keeps full tool-call
arguments and results (byte offset 54420734, same session tag):

```
cron_62e8cd9db637_20260811_140000assistant[{"id": "call_TqRem8IwmSKCRDGvFUMMQESk", "call_id": "call_TqRem8IwmSKCRDGvFUMMQESk", … "function": {"name": "terminal", "arguments": "{\"command\":\"python3 /tmp/extract_engagement_collector.py && LINEAR_CURSOR='2026-08-11T11:31:25Z' LINEAR_RUN_END='2026-08-11T12:00:15Z' python3 /tmp/engagement-linear-collector.py\",\"timeout\":300}"}}]
```

Its result tail (byte offset 54575049):

```
… "context_completeness": {"all_required_calls_succeeded": true, …}}]}", "exit_code": 0, "error": null}call_TqRem8IwmSKCRDGvFUMMQESkterminal
```

So in one cron run under `cron_mode: deny`: two `python3 -c` extraction attempts
stopped by the guard, then `write_file` with no approval, then a **plain-path
`python3` invocation — twice over, `&&`-chained, with env-var prefixes — to
`exit_code: 0`.** No new run was triggered to produce this; it is a read of an
existing 12:00 UTC run. Supporting negative control: `grep` for `linear_board.py`
across all of `agent.log` returns **zero hits** — the sibling skill's documented
plain-path collector never once appears in a guard warning, because it never
trips one.

### 4.4 The §4 replacement text — TEXT HANDED TO THE COORDINATOR, not applied

**Not applied by this session.** Neither the live
`~/.hermes/skills/orchestration/engagement-checker/SKILL.md` nor its repo mirror
was edited by this lane; the block below is **output handed to the coordinator
for the lane that owns that file**, reproduced byte-for-byte from
`Q4-stable-shape.md` §3.2.

Current text it replaces: live file **line 149**, the first paragraph under
`## 4. Collect Linear deltas` (between the heading on 147 and the `**Run it
through the shell/terminal tool — never `execute_code`.**` paragraph on 151,
which stays exactly as it is):

> Require `LINEAR_API_KEY` to be present in the collector process environment without reading or displaying its value during prerequisite checks. Set `LINEAR_CURSOR` to the stored source cursor and `LINEAR_RUN_END` to the run's fixed end timestamp, then run the collector below locally with `python3`.

Replacement, verbatim:

```markdown
Require `LINEAR_API_KEY` to be present in the collector process environment without reading or displaying its value during prerequisite checks. Set `LINEAR_CURSOR` to the stored source cursor and `LINEAR_RUN_END` to the run's fixed end timestamp.

**Materialize the fenced block below to a file with the `write_file` tool, then run that file by absolute path — `LINEAR_CURSOR='…' LINEAR_RUN_END='…' python3 /tmp/engagement_linear_collector.py`. That two-step shape is mandatory and the shell command carries nothing else.** A scheduled run has no user present to answer an approval prompt, so `approvals.cron_mode: deny` turns any flagged command into a `status: "blocked"` tool error that never reaches the delivered report: the collection simply does not happen and the run goes `[SILENT]` looking healthy. The guard flags exactly the shapes an agent reaches for first, and each one is forbidden here: `python3 -c '…'` and `python3 - <<'PY'` (*script execution via -e/-c flag*, *via heredoc*), `perl -ne`/`ruby -e` line extraction (same rule, interpreter family), piping the block into an interpreter — `… | python3 -` (Tirith *pipe_to_interpreter*, HIGH) — and `rm -f /tmp/<file>` afterwards, which matches *delete in root path* because the path is absolute. **Leave the extracted file where it is; do not clean it up.** `write_file` is not command-guarded and needs no approval, and a bare interpreter followed by a path is the one invocation the guard's exec-flag scan stops before reading — env-var prefixes and `&&` chaining of two such invocations are inside the safe shape. Measured 2026-08-11 in cron run `cron_62e8cd9db637_20260811_140000`: two `python3 -c` extraction attempts were blocked at 12:02:58 and 12:03:16 UTC, then `write_file` at 12:03:50 and `python3 /tmp/extract_engagement_collector.py && LINEAR_CURSOR='…' LINEAR_RUN_END='…' python3 /tmp/engagement-linear-collector.py` at 12:04:01 returned `exit_code: 0`.
```

### 4.5 Caveats on Q4

- Tirith is an external, self-updating binary; its Python wrapper was read but the
  binary's rule set **was not enumerated** and was deliberately not executed
  against candidate strings. The verdict is **empirical for 2026-08-11's rule
  set, not a proof against future rules.**
- `/tmp` is ephemeral and shared — the proving run's `write_file` result carried
  `"_warning": "… was modified by sibling subagent … but this agent never read it"`.
  The *shape* is guard-stable; the *location* is not durable.
- These cron sessions reach the `is_gateway`/`is_ask` fallback branch
  (`pending_approval`, seconds-long hangs) rather than the documented instant
  cron-deny block — not root-caused (§5). It does not change the verdict: both
  branches fire only when `detect_dangerous_command` or Tirith produced a
  finding, so a shape producing neither avoids both.
- **For the lane that owns §4, an inherited wording defect in the block above,
  left unpatched on purpose** (the block must stay byte-identical to what its
  author returned): it says the two 12:02:58 / 12:03:16 attempts were "blocked",
  while the log shows both as `"status": "pending_approval", "approval_pending":
  true` — the fallback branch, not the cron-deny block. The distinction does not
  change the shape verdict, but fix it in the same edit if you touch the text.

---

## 5. Open / MIGHT-DO candidates

**Candidates for the coordinator, one line each. This lane created no Linear
tickets and none of these is a ticket.**

- **`engagement-checker/SKILL.md:139` — the bare-`python` `GAPI=` line** (§3.6):
  the line the cron email collector actually runs; sibling lane
  `gcn30-32-checker-fix` owns the file, change text already handed over.
- **Mirror `skills/productivity/google-workspace/SKILL.md` into this repo** — it
  is not mirrored, so a `hermes` upgrade can silently revert today's fix (same
  class as the `tool_progress_dm` patch); mirror it, or add it to the
  re-apply-after-upgrade list.
- **Fix the `last_failure_notice` misfile** (§2.4b): §8's seven must-never-hide
  states are declared exempt from that key, yet the 18:09:11 run wrote
  `last_failure_notice.linear` for state 5; next EC run inside 4 h is the test.
- **Log `final_response` (bounded, redacted)** at `cron/scheduler.py:4183/:4197`
  — instrumentation gap 1 of 3 from Q2 §6; today only `response_len` and delivery
  target are recorded, which is the single reason Q2 is UNPROVABLE.
- **Log the blocked command string on the `cron_mode: deny` branch**
  (`approval.py:3808-3817`) as the `pending_approval` branch already does — gap 2
  of 3; today the scope of a blocked collection is unrecoverable from the logs.
- **Emit a per-run coverage record into state** (`runs[].sources[].{attempted,
  succeeded, blocked_reason}`) — gap 3 of 3; today "collected nothing" and
  "collected an empty window" look identical: no state write at all.
- **Ship the collector as `scripts/linear_collector.py`** inside the
  engagement-checker skill and invoke it by `$HERMES_HOME`-relative path — the
  precedent already used by `daily-work-brief/scripts/linear_board.py` and by
  §3's `google_api.py`; removes the per-run materialization and the `/tmp`
  durability caveat, strictly better than §4's two-step but a larger change.
- **Root-cause the `pending_approval` branch firing in cron sessions**, which
  `_is_gateway_approval_context()`'s own docstring says is impossible for cron
  (D-guard-rules §5) — these approvals are never answered and age out: a second,
  silent way a collector invocation fails to run.
- **Re-exec guard in `setup.py`/`google_api.py`** (sketched, untested) for an
  agent improvising `python3 …/google_api.py` outside the documented string —
  only if a post-fix run shows that failure again; it is code and needs its own
  test run.
- **`slack.seen` id-format drift** — the tail is bare epoch timestamps with no
  `slack:` prefix and a `slack-event:` run sits mid-array; noted, not
  investigated, out of GCN-31's scope.

---

## 6. Method + limits

**Read-only except for one two-line edit.** Every probe in this lane was
file/log/state reads over `ssh gaetan@192.168.0.9` plus `--version` /
`-c "import …"` interpreter probes run with `python -B` (bytecode writes
suppressed; `__pycache__` mtimes verified unchanged afterwards — Aug 7 / Aug 10).
The only mutation is §3.2: two lines of
`~/.hermes/skills/productivity/google-workspace/SKILL.md`, backed up first, plus
the token refresh that `--check` performs by design (backed up first). **No
`config.yaml`, `.env`, `SOUL.md` or `~/.hermes/skills/orchestration/**` write; no
package install; no venv creation; no `systemctl`.** No secret was read, printed
or copied; credential material is reported by path/size/perms only
(`~/.hermes/google_client_secret.json` 462 B `-rw-rw-r--`,
`~/.hermes/google_token.json` 1153 B, mode 664).

**What was deliberately NOT run, and why.** Tomorrow morning's WF6 R2 measurement
needs a pristine queue and run history, so nothing that enqueues or dequeues work
was invoked: **the engagement-checker was never run, no `hermes chat`, no
`hermes cron run`, no `hermes cron create`, no gateway restart.** Queue and
run-history are intact. Every run cited above is a run that had already happened;
§4's positive proof was recovered from `~/.hermes/lcm.db` (which stores full
tool-call arguments and results) rather than from `agent.log`, precisely so that
**no new run had to be triggered to manufacture it.** `hermes mcp test` was not
run; Tirith was not executed against candidate command strings.

**Known blind spots in the evidence itself, not in the method.**

- Every `BLOCKED:` line is truncated **in the source log** at `_err_text[:200]`
  (`agent/tool_executor.py:1440-1441`) — 328 bytes per BLOCKED line as measured
  (332 for the `execute_code` one), the JSON wrapper eating ~90–100 of the 200.
  The full cron-deny message quoted anywhere in this file comes from the source
  code, never reconstructed from a log line.
- `agent.log` logs a *successful* command's duration and output length, never its
  text. The only verbatim successful command in this file comes from `lcm.db`.
- Delivered report bodies are not logged for any run — only `response_len` and
  delivery target. That is the whole of why Q2 cannot be closed.
- `agent.log` covers 2026-08-10 11:39:13 → 2026-08-11 19:11:58 UTC; `agent.log.1`
  is contiguous before it. No rotation beyond `.1` exists, so history before
  ~2026-08-10 11:39 UTC is unavailable.
- No session transcript survives for `20260811_180849_14f453`
  (`~/.hermes/sessions/` holds only a 93 KB `sessions.json` last written 16:51
  UTC, zero hits for `14f453`).
