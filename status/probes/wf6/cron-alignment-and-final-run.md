# WF6 — engagement-checker env-var declaration, cron-prompt alignment, and the real-cron-path question

Agent: Claude Code session on cooper, worktree
`/home/gaetan/dev/orca-worktrees/Tars/wf6-vm-skills` at HEAD `c50ec4f`.
VM: `gaetan@192.168.0.9`, clock UTC. All times below UTC unless a `+02:00`
offset is shown.

Verdict, three lines:

- **Step 1 done and re-proved green.** engagement-checker is v1.7.1 with
  `required_environment_variables: [LINEAR_API_KEY]`; repo and VM sha256 match;
  the skill still loads enabled; the gateway did not restart; one full
  engagement-checker cycle on the real cron prompt completed clean, returned
  `[SILENT]`, advanced all three cursors and released its lock.
- **Step 2 done.** All three cron prompts now carry the text their skill
  prescribes. A field-by-field before/after over every stored job field shows
  `prompt` as the only field that changed on any of the three.
- **Step 3 STOPPED, deliberately.** `hermes cron run` **does deliver to the
  job's configured target**, there is no dry-run or no-deliver flag, and the
  target is Gaetan's reporting thread. Running it would have posted a full
  daily brief into that thread at ~23:20 Paris. Not run. Evidence in §3.

**Nothing was posted to Slack by this session** — verified against the thread
and the DM in §1.6.

---

## 1. Step 1 — engagement-checker declares `LINEAR_API_KEY`

### 1.1 The defect this closes

`engagement-checker`'s §4 collector needs `LINEAR_API_KEY` in its process
environment. Hermes' `execute_code` sandbox scrubs every environment variable
whose name contains `KEY` (`_SECRET_SUBSTRINGS`); the terminal/shell tool keeps
it. The skill declared no `required_environment_variables` and said nothing
about which tool to run the collector through, so a run that happened to pick
`execute_code` would have the collector fail closed on its own prerequisite
check. Check 2 passed only because that run used the terminal tool.

`daily-work-brief` v1.5.0 already carries the same frontmatter key and was
proven not to break skill loading, which is what made this safe to do.

Independent corroboration that `execute_code` really is the tool Tars reaches
for on this workflow: the reporting thread shows Tars performing
engagement-checker state edits with `:snake: Running code import json, os,
tempfile from dateti...` on 2026-08-10. The tool choice is not hypothetical.

### 1.2 The change — diff, repo side, `skills/engagement-checker/SKILL.md`

Two hunks. Nothing else in the file was touched: §7a, `sanitized()`, the `OPS`
allowlist, the `"mutation" in query.lower()` guard at the top of `_post()`, and
the `U0BBH85NAKH` pin in §2 are all byte-identical to v1.7.0.

```diff
--- SKILL.md.bak-wf6-envvar-20260810T210155Z	2026-08-10 20:27:01.119008462 +0000
+++ SKILL.md	2026-08-10 21:01:54.872010083 +0000
@@ -1,7 +1,8 @@
 ---
 name: engagement-checker
 description: "Use for incremental follow-up and commitment reminders."
-version: 1.7.0
+version: 1.7.1
+required_environment_variables: [LINEAR_API_KEY]
 metadata:
   hermes:
     tags: [engagement, reminders, slack, email, linear, orchestration]
@@ -145,6 +146,8 @@

 Require `LINEAR_API_KEY` to be present in the collector process environment without reading or displaying its value during prerequisite checks. Set `LINEAR_CURSOR` to the stored source cursor and `LINEAR_RUN_END` to the run's fixed end timestamp, then run the collector below locally with `python3`.

+**Run it through the shell/terminal tool — never `execute_code`.** Measured: the `execute_code` sandbox scrubs every environment variable whose name contains `KEY`, so `LINEAR_API_KEY` is absent there and the collector can only fail closed on its own prerequisite check; the terminal tool keeps it. That is also why the frontmatter declares `required_environment_variables`.
+
 **The fenced block is the read collector and nothing else — it contains no write path, and none is to be added.** ...
```

(The second hunk's trailing context line is the unchanged §4 collector
paragraph, elided here for length; the diff was generated on the VM as
`diff -u SKILL.md.bak-wf6-envvar-20260810T210155Z SKILL.md` and had exactly
these two hunks.)

Frontmatter key placement and YAML style match `daily-work-brief` exactly: the
key sits between `version:` and `metadata:`, flow-sequence form,
`required_environment_variables: [LINEAR_API_KEY]`.

### 1.3 Deploy — backup, lock, sha256 both sides

Backup taken before the write:
`~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-wf6-envvar-20260810T210155Z`
(the v1.7.0 file, 67313 bytes, sha256 `bb55e1ef…`).

Write path: content streamed to `SKILL.md.wf6new`, then moved into place under
the lock —

```
flock ~/.hermes/.wf3.lock -c "mv .../SKILL.md.wf6new .../SKILL.md"
```

sha256, both sides, after the deploy:

```
repo  973cdec80c01c483cf93d18f984cbf64b9faf60245ed2ea3d550599fccaa58b2  skills/engagement-checker/SKILL.md
VM    973cdec80c01c483cf93d18f984cbf64b9faf60245ed2ea3d550599fccaa58b2  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md
```

**MATCH.** File is 67740 bytes, mode `0600`, owner `gaetan`.

`daily-work-brief` was not touched: repo and VM both remain
`e55278b0322c799d94d19171c28a48712917353c6c767ee8ad79f853e741314c`.

### 1.4 Skill still loads, and loads enabled

```
$ COLUMNS=200 hermes skills list --enabled-only | grep engagement-checker
│ engagement-checker            │ orchestration        │ local     │ local     │ enabled │
```

`--enabled-only` hides disabled skills, so appearing in that view *is* the
enabled proof. `daily-work-brief` and `linear-ticketing` also still list
`enabled`.

### 1.5 Gateway did not restart

`NRestarts` is unreliable and is ignored. `ActiveEnterTimestamp` is the test:

```
before deploy   ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC   ActiveState=active  SubState=running
after  deploy   ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC   ActiveState=active  SubState=running
```

Identical. The gateway live-reloaded the skill without restarting.

### 1.6 Re-proof of check 2 — one full cycle on the real cron prompt

Method, identical to the original check-2 probe so the two are comparable. The
prompt was read **out of live `~/.hermes/cron/jobs.json`** for job
`62e8cd9db637` (i.e. the pre-step-2 text, reproduced verbatim in §2.1), with
exactly one paragraph appended:

> OPERATOR OVERRIDE, THIS MANUAL RUN ONLY: do NOT post, send, deliver or
> schedule anything to Slack — no message, no DM, no thread reply, no reaction.
> Slack READS are fine; every Slack WRITE is forbidden this run. Instead, return
> the final delivery text (or exactly [SILENT]) to me here as your final answer.
> Everything else about the run is unchanged.

Combined prompt `/tmp/ec-run-prompt-wf6b.txt`, 1208 bytes, sha256
`7adbd02efbe5673805c6d98a9cffefdded806ef26b49fa057db6a150d8451c82`.

Invocation, detached on the VM:

```
hermes chat -q "$(cat /tmp/ec-run-prompt-wf6b.txt)" -s engagement-checker \
  --provider openai-codex -m gpt-5.6-sol --yolo -Q
```

`--provider`/`-m` are the job's own `provider_snapshot`/`model_snapshot`.
`hermes chat` carries no `--deliver` target, so the cron delivery path is
structurally absent — the first of the two Slack guarantees, the operator
override being the second. Session id `20260810_210250_306801`. Started
21:02:50, finished ~21:05. Exit clean; `stderr` contained only
`session_id: 20260810_210250_306801`.

**Captured output, verbatim and complete** (`/tmp/ec-run-wf6b.out`):

```
  ┊ review diff
a//home/gaetan/.hermes/skills/email/himalaya/SKILL.md → b//home/gaetan/.hermes/skills/email/himalaya/SKILL.md
@@ -275,9 +275,10 @@

 ## Output Formats

-Most commands support `--output` for structured output:
-
-```bash
+Most commands support `--json` for structured output (Himalaya v2 uses `--json`, not `--output json`):
+
+```bash
+himalaya account list --json
 himalaya envelope list --json
 himalaya envelope list
 ```
  ⟳ compacting context…
  ⟳ compacting context…
[SILENT]

Operational note: I corrected the `himalaya` skill to document its v2 `--json` flag. The first sync mistakenly produced an empty file and merged as PR https://github.com/GaetanCathelain/Tars/pull/42; I immediately restored the full skill with the correction in https://github.com/GaetanCathelain/Tars/pull/43, which also merged.
```

**Verdict: clean.** `[SILENT]` is the required outcome for a drained queue.
No error, no crash, no held cursor:

| durable state `~/.hermes/state/engagement-checker.json` | before | after |
|---|---|---|
| `last_completed_run` | 2026-08-10T22:34:28+02:00 | 2026-08-10T23:03:11+02:00 |
| slack cursor / `last_success` | 22:34:28 / 22:34:28 | 23:03:11 / 23:03:11 |
| email cursor / `last_success` | 22:34:28 / 22:34:28 | 23:03:11 / 23:03:11 |
| linear cursor / `last_success` | 22:34:28 / 22:34:28 | 23:03:11 / 23:03:11 |
| items | 30 — done 26, dismissed 3, snoozed 1 | 30 — done 26, dismissed 3, snoozed 1 |
| items carrying `linear_issue` | 1 | 1 |
| size / sha256 | 79440 / `5386df51…` | 79440 / `301d0f32…` |

Every source's `last_success` equals the run's end instant, so §8's
"source failed this run" test (`last_success < run end`) is false for all
three — **no held cursor**. The lock directory
`~/.hermes/state/engagement-checker.lock` does not exist after the run:
cleanup ran. No new item, no new filed issue — the queue was already drained,
which is the expected and stated-acceptable outcome.

Two non-fatal tool errors were logged inside the run, both handled by the skill
rather than escalating:

- `kanban_show` — `{"error": "task_id is required (or set HERMES_KANBAN_TASK in the env)"}`. Irrelevant to this workflow.
- `terminal` — a traceback from `~/.hermes/skills/productivity/google-workspace/scripts/google_api.py`, i.e. the primary Gmail path failed. The skill's own instruction is to fall back to `himalaya` before declaring email unavailable, and that is what the run did — hence the email cursor advancing normally.

**Side effect worth flagging to the coordinator.** Chasing that fallback, Tars
found `himalaya`'s SKILL.md documented the wrong flag and exercised its SOUL
rule 2 self-merge path: PR #42 (merged 21:04:36Z) landed an **empty** file and
PR #43 (merged 21:04:58Z) restored it with the correction, +305 lines. Both PRs
touch **only** `skills/himalaya/SKILL.md` — verified with
`gh pr view {42,43} --json files`. Neither touches
`skills/engagement-checker/SKILL.md`, so the step-1 change is unaffected, but
`main` moved under this worktree while the cycle ran. This session ran no git
command; the change is Tars's own, on its own skill mirror, which SOUL rule 2
sanctions.

**Slack: nothing posted.** The reporting thread
`C0BP2GZUFSR:1786359613.979759` ends at the 17:00 final-pass reminder
(`• Send Jamil the Verdict Google OAuth client … [EC-A44C]`), with
`There are no more messages in this thread` — nothing at all after 15:04Z, and
the cycle ran 21:02–21:05Z. The DM `D0BBYNM01BL` has nothing later than
20:09:34Z. Both read-only, via the Slack connector.

---

## 2. Step 2 — the three cron prompts now match their skills

### 2.1 THE ORIGINAL PROMPTS, VERBATIM — this is the only backup

Cron prompts have no `.bak` on disk. These three blocks are the restore source.
Captured from `~/.hermes/cron/jobs.json` **before** any edit.

#### `e231e5faf180` — Gaetan daily work brief — ORIGINAL prompt (407 bytes)

```
Prepare Gaetan's morning daily by following the attached daily-work-brief skill completely. Use Europe/Paris for all date/time decisions. For the Monday 2026-08-10 run only, include in the Today section a reminder to try the idea from this Instagram reel: https://www.instagram.com/reel/DY42w69iOb4/?igsh=MXh4Z3kyYjE3ZTQ0ZA== . On later runs, do not repeat this reminder unless Gaetan explicitly asks again.
```

#### `62e8cd9db637` — Gaetan engagement checker — ORIGINAL prompt (855 bytes)

```
Run `engagement-checker` end to end. Use Europe/Paris and the fixed run timestamp. Apply the same workday gate as `daily-work-brief` first: weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off; return exactly `[SILENT]` when the gate says not to run. Otherwise acquire the single-writer lock; read durable state; collect and exactly filter only Slack, email, and Linear deltas since each source cursor; reconcile user decisions and resolved loops; update each source cursor only after successful processing; evaluate the pending queue with cooldowns; persist state; release the lock; then return the compact reminder or exactly `[SILENT]`. This is a read-only source workflow. Do not spawn or prompt Claude or Orca; Cooper/Orca may only be queried read-only for Linear evidence.
```

#### `759e08c598e3` — Gaetan engagement checker final pass — ORIGINAL prompt

Byte-identical to `62e8cd9db637`'s original above.

### 2.2 The new prompts

Extracted programmatically from the deployed skills — the single `> `-quoted
block in `daily-work-brief` §6 and in `engagement-checker` §Scheduling contract
— with the `> ` marker stripped and nothing else altered, so the stored prompt
is byte-for-byte the text the skill prescribes.

#### `e231e5faf180` — NEW prompt (1877 bytes, sha256 `a150b59bf58b64c7…`)

```
Run `daily-work-brief` end to end. Use Europe/Paris. Establish the window from the last successful delivery, then apply the workday gate — weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off — and return exactly `[SILENT]` when the gate says not to run. Otherwise collect Slack, Cooper (Orca state, Claude transcripts, shell history, git), GitHub, Linear, email and calendar evidence in parallel, keep raw results out of the message, and deliver the compact brief. Do Linear LAST. The brief must OPEN with the board block, and **the board block is the stdout of `python3 ${HERMES_HOME:-$HOME/.hermes}/skills/orchestration/daily-work-brief/scripts/linear_board.py`, run through the shell/terminal tool (never `execute_code`, which scrubs `LINEAR_API_KEY`) and included byte for byte.** Do not write, correct, re-order, re-title, filter or re-type a single row: the script already sorted, capped, computed `+N more` and emitted the `Coverage:` line. If it printed `board unavailable (coverage unproven)`, print exactly that one line as the board and nothing else. If its output is not in front of you — never run, scrolled away, or the context compacted — run it again rather than recalling rows; a fabricated board is worse than no board. Take `Linear tickets completed: N` from a separate native `mcp__linear__list_issues` read (`state` `"completed"`, `updatedAt` at the window start, `fields` `["id","title","completedAt"]`), proven complete by paging `cursor` until `hasNextPage == false` — never by narrowing the filter. `Since the last daily` must end with the labelled `Claude/Orca:` line covering the Cooper Claude/Orca history, or `Claude/Orca: none.` Every source is read-only: write nothing to Linear, and do not spawn, prompt, or delegate to another agent. Return only the brief.
```

This is the one the task called out specifically: the live prompt predated
SKILL.md §6 and carried **no board instruction at all**. It now carries the
byte-for-byte board rule, the `never execute_code` clause, the
`board unavailable (coverage unproven)` rule, and the "run it again rather than
recalling rows" rule — the three things the 2026-08-10 fabricated-board
incident turned on. It also drops the one-off Instagram-reel line, which by its
own terms was for the 2026-08-10 run only (already past) and was not to repeat.

#### `62e8cd9db637` and `759e08c598e3` — NEW prompt (1693 bytes, sha256 `ae335880ed94d729…`, identical on both)

```
Run `engagement-checker` end to end. Use Europe/Paris and the fixed run timestamp. Apply the same workday gate as `daily-work-brief` first: weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off; return exactly `[SILENT]` when the gate says not to run. Otherwise acquire the single-writer lock; read durable state; collect and exactly filter only Slack, email, and Linear deltas since each source cursor; reconcile user decisions and resolved loops; update each source cursor only after successful processing; evaluate the pending queue with cooldowns; then file retained open items as GCN issues **and close back the GCN issues this skill filed whose items were explicitly resolved elsewhere — never one dismissed merely for going stale**; persist state; release the lock; then return the compact reminder or exactly `[SILENT]`. Sources stay read-only, and the permitted writes are exactly the skill's "Permitted writes" section — that list, not this prompt, is the authority. Do not spawn or prompt another agent. Linear transport is mixed on purpose: run the delta scan through the audited local collector with the inherited `LINEAR_API_KEY`, because its coverage verdict gates the cursor, and take every other Linear read and every Linear write through the native `mcp__linear__*` tools per `linear-ticketing`. Never expose the key; never write outside GCN. Prove every native read complete by paging `cursor` until `hasNextPage` is false, never by narrowing the filter. If a required read cannot be proven complete or a write cannot be confirmed, say so in the delivery rather than proceeding as if it worked.
```

This closes the drift logged in `check2-verification.md` and
`deploy-and-verify.md`: the old prompt said "This is a read-only source
workflow" and mentioned neither the GCN filing nor the close-back, both of
which the skill's §Permitted writes has required since v1.6.

### 2.3 How the edit was made

`hermes cron edit --help` was read first (CLI-facts rule). The relevant surface:

```
usage: hermes cron edit [-h] [--schedule SCHEDULE] [--prompt PROMPT] [--name NAME]
                        [--deliver DELIVER] [--repeat REPEAT] [--skill SKILLS] ... job_id
```

Only `--prompt` was passed. To keep backticks, `**`, and
`${HERMES_HOME:-$HOME/.hermes}` from being touched by a shell, the call was made
from Python with an argv list and no shell:

```python
subprocess.run(["/home/gaetan/.local/bin/hermes","cron","edit",jid,"--prompt",text])
```

All three returned rc 0.

### 2.4 Field-by-field before/after — only `prompt` changed

Every stored field of all three jobs was dumped before and after, excluding only
the run-bookkeeping fields that move on their own (`last_run_at`, `last_status`,
`next_run_at`, `fire_claim`, `run_claim`, `last_error`, `last_delivery_error`,
`monitor_state`). Compared key by key:

```
e231e5faf180 changed fields: ['prompt']
62e8cd9db637 changed fields: ['prompt']
759e08c598e3 changed fields: ['prompt']
```

That is the whole diff — `base_url`, `context_from`, `created_at`, `deliver`,
`enabled`, `enabled_toolsets`, `id`, `model`, `model_snapshot`,
`monitor_script`, `monitor_url`, `name`, `no_agent`, `origin`, `paused_at`,
`paused_reason`, `provider`, `provider_snapshot`, `repeat`, `schedule`,
`schedule_display`, `script`, `skill`, `skills`, `state`, `workdir` all
unchanged on all three. No revert was needed.

Post-edit state read back out of `jobs.json`, with the stored prompt compared
against the intended text as an exact string:

```
e231e5faf180 match: True | stored bytes 1877 | sha256 a150b59bf58b64c7 | enabled True | deliver slack:C0BP2GZUFSR:1786359613.979759 | sched 30 8 * * 1-5        | skills ['daily-work-brief']   | state scheduled
62e8cd9db637 match: True | stored bytes 1693 | sha256 ae335880ed94d729 | enabled True | deliver slack:C0BP2GZUFSR:1786359613.979759 | sched */30 10-16 * * 1-5  | skills ['engagement-checker'] | state scheduled
759e08c598e3 match: True | stored bytes 1693 | sha256 ae335880ed94d729 | enabled True | deliver slack:C0BP2GZUFSR:1786359613.979759 | sched 0 17 * * 1-5        | skills ['engagement-checker'] | state scheduled
```

`hermes cron list` agrees — all three `[active]`, schedules `30 8 * * 1-5`,
`*/30 10-16 * * 1-5`, `0 17 * * 1-5`, all delivering to
`slack:C0BP2GZUFSR:1786359613.979759`, next runs 2026-08-11 08:30 / 10:00 /
17:00 +02:00.

---

## 3. Step 3 — `hermes cron run` delivers to Slack. NOT RUN.

### 3.1 The finding

**`hermes cron run <job-id>` delivers to the job's configured delivery target.
There is no dry-run, no no-deliver, and no local-only flag.**

Three independent pieces of evidence on the VM's own Hermes source
(`~/.hermes/hermes-agent/`, v0.20.0):

1. **The CLI surface has nothing to suppress delivery with.**

   ```
   usage: hermes cron run [-h] [--accept-hooks] job_id
   ```

   `--accept-hooks` only auto-approves shell hooks. `hermes_cli/cron.py:580`
   maps the subcommand straight to `_job_action("run", args.job_id, "Triggered")`.

2. **The run path resolves and installs the job's delivery target.**
   `cron/scheduler.py:3576` states outright that
   `cronjob(action="run")` calls `run_one_job() -> run_job()`; inside `run_job`,
   `cron/scheduler.py:3613`:

   ```python
   delivery_target = _resolve_delivery_target(job)
   if delivery_target:
       _VAR_MAP["HERMES_CRON_AUTO_DELIVER_PLATFORM"].set(delivery_target["platform"])
       _VAR_MAP["HERMES_CRON_AUTO_DELIVER_CHAT_ID"].set(str(delivery_target["chat_id"]))
       _VAR_MAP["HERMES_CRON_AUTO_DELIVER_THREAD_ID"].set(...)
   ```

   and `_resolve_delivery_targets` (`cron/scheduler.py:1398`) reads
   `job.get("deliver", "local")`, returning an empty target list **only** when
   the value is literally `local`. Job `e231e5faf180`'s value is
   `slack:C0BP2GZUFSR:1786359613.979759` — Gaetan's reporting thread. It
   delivers there, whether the trigger fires inline ("Ran now") or on the next
   scheduler tick; both go through the same `run_job`.

3. **The only delivery suppression in the whole scheduler is the agent's own
   `[SILENT]` marker** (`cron/scheduler.py:295`, `:312`, `:2695`). A grep for
   `dry[_-]?run|no[_-]deliver|skip[_-]deliver|suppress.*deliver` across
   `cron/scheduler.py` and `hermes_cli/cron.py` returns those `[SILENT]`
   comments and nothing else.

`[SILENT]` is not a usable guard here: 2026-08-11 is a weekday, the
daily-work-brief workday gate would pass, and the skill would produce a real
brief — a post, not silence.

### 3.2 Decision

**Not run.** Per the task's own instruction, an unexpected post into Gaetan's
reporting thread is not this session's to risk. Triggering `cron run
e231e5faf180` at ~21:20Z would have posted a full, unrequested daily brief into
`C0BP2GZUFSR:1786359613.979759` at ~23:20 Paris.

There is no safe variant available from the CLI:

- `hermes cron tick` runs *due* jobs, and would hit the same delivery path for
  anything it fires.
- `hermes cron pause` + `run` does not help: pausing removes the job from
  scheduling, it does not make an explicit trigger local.
- Temporarily editing `--deliver local` and back would mutate a field the task
  explicitly forbids changing, and would leave a window in which a real 08:30
  brief could fire to nowhere.

Whatever the operator decides, the closest safe rehearsal that already exists is
the §1.6 pattern: `hermes chat -q "<the job's stored prompt>" -s <skill>` with
the operator-override paragraph appended. That exercises the same skill, the
same model and the same stored prompt text, and is structurally incapable of
delivering — but it is not the cron path, which is exactly what the coordinator
asked to verify, so it is offered as a fallback, not as a substitute.

### 3.3 What is therefore still unverified

The board block has **not** been observed coming out of the real cron
invocation path. It was verified by check 3 (board byte-identical to
`scripts/linear_board.py` output) and the script is now named byte-for-byte in
the job's prompt, but the end-to-end `cron run → agent → shell tool →
linear_board.py → Slack` chain has not been exercised. The next natural,
zero-risk opportunity is the scheduled 08:30 +02:00 run on 2026-08-11, whose
output lands in the reporting thread anyway; compare its opening block against
a direct run of the script at that moment.

---

## Appendix — files touched

Repo (this worktree), one file:

- `skills/engagement-checker/SKILL.md` — v1.7.0 → v1.7.1, two hunks, §1.2.

VM:

- `~/.hermes/skills/orchestration/engagement-checker/SKILL.md` — deployed,
  sha256 `973cdec8…`; backup `SKILL.md.bak-wf6-envvar-20260810T210155Z`.
- `~/.hermes/cron/jobs.json` — `prompt` field of `e231e5faf180`,
  `62e8cd9db637`, `759e08c598e3`; originals in §2.1.
- `~/.hermes/state/engagement-checker.json` — advanced by the §1.6 cycle, as
  any engagement-checker run does.
- `/tmp/ec-run-prompt-wf6b.txt`, `/tmp/ec-run-wf6b.out`, `/tmp/ec-run-wf6b.err`,
  `/tmp/prompt-dwb.txt`, `/tmp/prompt-ec.txt` — scratch, no secrets.

No git command was run by this session. No secret was read, printed, logged or
written to any file here.
