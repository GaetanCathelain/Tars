# WF5 v2 — spawn→wait→read→cleanup sequence validation

Read-only validation of `status/probes/wf5/orca-v2-capabilities.md` §9 ("Exact
spawn sequence, unrun") against the live `orca` CLI on cooper (app pid 218650,
`appVersion 1.4.176`, `orca status --json` → `runtime.reachable: true`,
confirmed at validation time — same instance the §9 sequence was composed
against). Method: `--help` on every command in the sequence, plus read-only
`--json` list/show calls against real objects (mc-metarepo's 9 worktrees,
the `run_legacy_local` run, the 2 managed Claude accounts) to cross-check
field names. **No create/send/spawn/remove command was executed.** Two
genuinely negative-path reads were done (`repo show`/`worker-show`/`check`
against ids that don't exist) — these are read-only lookups, not mutations,
and were used specifically to capture the failure-mode JSON shapes below.

## 1. Command-by-command verdict

| # | Command in §9 | Verdict | Detail |
|---|---|---|---|
| 0a | `orca status --json` | VERIFIED | Ran live. `result.runtime.reachable: true`, `result.app.pid: 218650`. Shape matches. |
| 0b | `orca account list --json` | VERIFIED | Ran live. Active Claude account `gaetan.cathelain-claude02@mobile.club`, session 10%/weekly 17% — unchanged from the capabilities recon, still ample headroom. |
| 1a | `orca orchestration run-create --objective <text> [--from <handle>] [--json]` | VERIFIED (flags); response field UNVERIFIABLE-WITHOUT-EXECUTION | `--help` confirms `--objective`, `--from`, `--retry-request`, `--json` exactly as used. `result.run.id` (the field §9 tells the caller to capture) could not be cross-checked: the only existing run (`run_legacy_local`) was legacy-adopted, not created via `run-create`, and `run-list --json` returns it as a **flat** `result.runs[].id` (not nested under a `run` key) — circumstantial support for `result.run.id` on a single-object response, not proof. Capture and log the real shape on first execution. |
| 1b | `orca orchestration task-create --run <id> --spec <text> --task-title <text> --json` | VERIFIED (flags); response field UNVERIFIABLE-WITHOUT-EXECUTION | `--help` confirms `--spec`, `--task-title`, `--run`, `--json` exactly. `result.task.id` unverifiable — `task-list --run run_legacy_local` returns `count:0`, no real task object exists anywhere on the host to inspect. |
| 2 | `orca orchestration worker-start --run --task --worktree new-top-level --repo id:… --name --agent claude --setup run --json` | VERIFIED, exact | `--help`'s usage line is `--worktree <current\|selector\|new-child\|new-top-level>` — `new-top-level` is a literal enum value, spelled correctly. `--repo`, `--name`, `--agent`, `--setup <run\|skip\|inherit>` all confirmed present with the exact spelling used. Exit-code contract confirmed verbatim in `--help` Notes (see §3 below). `result.dispatch.id` unverifiable — **zero rows** in `worker_dispatches` (`sqlite3 -readonly … "select count(*) from worker_dispatches" → 0`), so no live dispatch object exists anywhere to inspect. |
| 3a | `orca orchestration worker-show --dispatch <id> --json` | VERIFIED (flags); `.result.state` enum VERIFIED via schema | Only flag is `--dispatch`. The 9-value enum documented in §9 (`starting…abandoned`) matches the live `CHECK` constraint on `worker_dispatches.state` read via `sqlite3 -readonly .schema worker_dispatches` — schema-level proof, not a live object (none exist). |
| 3b | `orca orchestration task-list --run <id> --status dispatched --json` | VERIFIED, exact | Ran live against `run_legacy_local` with `--status dispatched` combined with `--run`: returns `200` shape, `result.{runId,legacyReadOnly,tasks,count}`. Flag combo works mechanically. |
| 4 | `orca orchestration check --run <id> --wait --types worker_done,escalation,question --timeout-ms 900000 --json` | **CORRECTED — flag exists, but `--run` resolution failed on the only real run on the host** | All flags (`--run`, `--wait`, `--types`, `--timeout-ms`, `--json`) are confirmed present in `--help`. **But**: `orca orchestration check --run run_legacy_local --json` (and with `--wait --timeout-ms 1500`) both returned `{"ok":false,"error":{"code":"run_not_found","message":"Run run_legacy_local was not found."}}` exit 1 — on the exact same run id that `task-list`/`worker-list`/`gate-list --run run_legacy_local` all resolve successfully (`ok:true`) in the same session. `run_legacy_local` has `coordinator_handle: null, coordinator_pane_key: null, legacy: 1` in `run-list --json` — plausible explanation is `check` requires a bound coordinator handle that legacy-adopted runs never got, so this may not reproduce on a freshly `run-create`d run. **Not proven either way without executing `run-create`.** Flag it as a checkpoint: verify `check --run "$RUN_ID"` resolves (even a cheap `--timeout-ms 500` no-op) immediately after step 1a, before trusting the long `--wait` in step (b). |
| 5 | `orca orchestration worker-read --dispatch <id> --limit 200 --json` | VERIFIED, exact | `--dispatch`, `--limit` confirmed. `--source`/`--cursor` also exist (optional, correctly unused). |
| 6 | `orca orchestration worker-release --dispatch <id> --json` | VERIFIED, exact | Only mutating flags are `--dispatch`/`--retry-request`. Exit-code contract (`already_released`, `release_unknown` is the only exit-1 case) confirmed verbatim in `--help` Notes. |
| 7 | `orca worktree rm --worktree "id:<repo>::<path>" --force --json` | VERIFIED, exact | `--worktree`, `--force`, `--run-hooks` all confirmed. §9's decision to omit `--run-hooks` (empty archive script ⇒ no-op) still holds — `repo show` confirms `hookSettings.scripts.archive: ""` unchanged. |

**Score: 0 flags wrong. 1 real correction (item 4, `check --run` on the only
inspectable run failed where three sibling commands with the identical `--run`
succeeded) — everything else that could be checked without executing a
mutation is VERIFIED; response-body field names for `run-create`/
`task-create`/`worker-start` remain UNVERIFIABLE-WITHOUT-EXECUTION because no
real run/task/dispatch object exists anywhere on the host to inspect (`run
run_legacy_local` is legacy-adopted, not `run-create`d; `worker_dispatches`
has 0 rows).**

## 2. Read-only inspection commands (beyond the §9 sequence itself)

| Purpose | Command | Verdict |
|---|---|---|
| List repos | `orca repo list --json` | VERIFIED, ran live |
| Show one repo | `orca repo show --repo id:<id> --json` | VERIFIED, ran live (mc-metarepo shape unchanged since capabilities recon) |
| List worktrees for a repo | `orca worktree list --repo id:<id> --json` | VERIFIED, ran live — **still 9 worktrees** for mc-metarepo, matching the capabilities doc exactly (no drift) |
| Show one worktree | `orca worktree show --worktree <selector> --json` | Flags VERIFIED via `--help` (not re-run; already live-verified in the source recon against `name:"GoCardless rotation"`) |
| Show a Run (bonus, not in §9) | `orca orchestration run-show --id <run_id> --json` | Note: flag is **`--id`, not `--run`** — `run-show --run run_legacy_local` errors `invalid_argument`/`Unknown flag --run`. Not used by §9, but a skill author reaching for it by analogy would hit this immediately. |
| List/show a worker/dispatch | `worker-show --dispatch`, `worker-list [--run] [--terminal-state]` | VERIFIED. `worker-list`'s only filters are `--run` and `--terminal-state <active\|reclaimable\|retained\|release_pending\|release_unknown\|released>` — **no `--status` and no `--all` flag**; both were tried and rejected with `invalid_argument` (see §3). |
| List tasks for a run | `orca orchestration task-list --run <id> [--status] [--ready] [--brief] --json` | VERIFIED, ran live |
| Read a worker's output | `orca orchestration worker-read --dispatch <id> [--source] [--cursor] [--limit] --json` | VERIFIED via `--help` |
| Release/cleanup a worker | `orca orchestration worker-release --dispatch <id> --json` | VERIFIED via `--help` |
| Remove a worktree | `orca worktree rm --worktree <selector> --force [--run-hooks] --json` | VERIFIED via `--help` |

Side note (not part of §9, flagged because the capabilities doc's §2 claims
it): **`orca orchestration inbox` does NOT accept `--run`** — only
`--terminal`, `--full`, `--limit`. `inbox --run run_legacy_local --json`
returns `invalid_argument`/`Unknown flag --run for command: orchestration
inbox` live today. The capabilities doc's Sources list bundles `inbox` into
the same brace-expanded line as `task-list,worker-list,gate-list --run
run_legacy_local` — that bundling appears to be imprecise; `inbox` was not
actually exercised with `--run` successfully. Doesn't affect §9 (which never
calls `inbox`), noted for whoever writes the skill's inspection helpers.

## 3. Corrected end-to-end sequence

Only one line changes from §9 — insert a cheap `check --run` verification
checkpoint right after the run is created, and treat the wait-step's
`run_not_found` as ambiguous rather than fatal until that checkpoint has
passed once for real:

```bash
# 0. Preflight — unchanged, verified.
orca status --json
orca account list --json

# 1. Durable namespace + task row — unchanged, verified.
orca orchestration run-create --objective "<one-line objective>" --json
#   -> capture RUN_ID; LOG THE FULL RAW RESPONSE before parsing — result.run.id
#      is a plausible field name (cross-checked against run-list's flat shape)
#      but was never observed live.
orca orchestration task-create --run "$RUN_ID" --spec "<brief>" --task-title "<title>" --json
#   -> capture TASK_ID; same logging caveat.

# 1.5 NEW — cheap checkpoint, catches the item-4 correction before the 15-min wait.
orca orchestration check --run "$RUN_ID" --timeout-ms 500 --json
#   Expect ok:true (even with 0 messages). If this comes back run_not_found,
#   STOP: check cannot resolve this run and the blocking --wait in step (b)
#   below will also fail — do not proceed to worker-start on this RUN_ID.

# 2. Spawn — unchanged, verified exact.
orca orchestration worker-start --run "$RUN_ID" --task "$TASK_ID" \
  --worktree new-top-level --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 \
  --name a2a-test-spawn --agent claude --setup run --json
#   -> capture DISPATCH_ID; log full raw response, field name unconfirmed.

# (b) Block for completion — unchanged, verified exact (contingent on 1.5 passing).
orca orchestration check --run "$RUN_ID" --wait \
  --types worker_done,escalation,question --timeout-ms 900000 --json

# (c) Retrieve output — unchanged, verified exact.
orca orchestration worker-read --dispatch "$DISPATCH_ID" --limit 200 --json

# (d) Cleanup — unchanged, verified exact.
orca orchestration worker-release --dispatch "$DISPATCH_ID" --json
orca worktree rm --worktree "id:8099e312-3232-46f2-83a9-97aeaf5de5a2::/home/gaetan/orca/workspaces/mc-metarepo/a2a-test-spawn" --force --json
```

## 4. Failure-mode table

Exit codes and JSON shapes below are either **measured live** (bad repo id,
bad dispatch id, bad run id, unknown flag) or **quoted verbatim from
`--help`** where measuring would require a mutation (worker-start failure,
worker-release edge states, orca-not-running). Column 4 is the literal
recognition rule a skill should code against.

| Condition | Exit code | JSON shape | How a skill should recognise it |
|---|---|---|---|
| Orca app not running | non-zero (not measured — stopping the live app is destructive and out of scope) | `--help`'s own Behavior section, verbatim: *"Most commands require a running Orca runtime. If Orca is not open yet, run `orca open` first."* | Always run `orca status --json` first; check `result.runtime.reachable == true` before any other call, rather than parsing a specific error code from a failed call. |
| Bad repo id | 1 (measured) | `{"ok":false,"error":{"code":"repo_not_found","message":"repo_not_found"}}` | Match `error.code == "repo_not_found"`. |
| Bad dispatch id | 1 (measured) | `{"ok":false,"error":{"code":"dispatch_not_found","message":"Worker Dispatch <id> was not found."}}` | Match `error.code == "dispatch_not_found"`. |
| Bad run id | 1 (measured) | `{"ok":false,"error":{"code":"run_not_found","message":"Run <id> was not found."}}` | Match `error.code == "run_not_found"` — **but see item 4 above**: this same code fired for a run that genuinely exists (`run_legacy_local`, confirmed via `run-list`) when the caller was `check`/`inbox` specifically. A skill must not conclude "the run id was typo'd" purely from this code on a `check` call — cross-check with `task-list --run <id>` (which resolved the same id fine) before treating it as a bad-id error. |
| Unknown/misspelled flag | 1 (measured, e.g. `worker-list --status …`, `run-show --run …`) | `{"ok":false,"error":{"code":"invalid_argument","message":"Unknown flag --x for command: …","data":{"validFlags":[…],"suggestions":[…],"nextSteps":[…]}}}` | Match `error.code == "invalid_argument"`; `error.data.validFlags` is a ready-made fix list — surface it directly in the skill's error message. |
| Worker fails to start | 1 (quoted from `worker-start --help` Notes, not triggered) | *"The call exits 0 only for ready. Failed or outcome_unknown exits 1 and JSON includes stage/failedStage, setup, effects, residualResources, and recovery commands when needed."* | Match on `exit_code != 0`, then read `stage`/`failedStage`/`effects`/`residualResources` for the reason — do not blind-retry; the Notes text explicitly says "inspect, don't retry blind" language is echoed in §9 too. |
| `check --wait` times out with nothing done | 0, `{count:0}` (quoted from the capabilities recon §3, item 1 — this recon did not re-derive it since it requires either a live worker or a very long real wait to distinguish from item-4's `run_not_found` failure mode) | `{"ok":true,"result":{"count":0,...}}` (shape inferred from `--wait`'s documented default-return contract; not independently re-measured this pass) | Treat as a **checkpoint, not a failure**: the `dispatch_id`/`run_id` are durable (§2 of the capabilities doc), so re-issue the same `check --wait` in a later turn rather than treating timeout as terminal. Distinguish from the item-4 case by exit code: timeout is `ok:true` exit 0; the run-resolution failure above is `ok:false` exit 1 with `run_not_found` — they will not be confused if the skill checks `ok` first. |
| `worker-release` non-terminal states | 0 for `retained`, `release_pending`, `already_released`; 1 only for `release_unknown` (quoted verbatim from `worker-release --help` Notes) | Not measured (no dispatch exists to release) | Exit 0 does not always mean "fully released" — read the returned state string itself (`retained`/`release_pending`/`already_released`/`released`) rather than trusting exit code alone; only `release_unknown` (exit 1) is an actual error requiring the "recovery action" the Notes reference. |

## 5. Sources — every command run in this validation pass

```
which orca
orca status --json
orca orchestration run-create --help
orca orchestration task-create --help
orca orchestration worker-start --help
orca orchestration worker-show --help
orca orchestration task-list --help
orca orchestration check --help
orca orchestration worker-read --help
orca orchestration worker-release --help
orca worktree rm --help
orca account list --help
orca repo list --help
orca repo show --help
orca worktree list --help
orca worktree show --help
orca terminal list --help
orca orchestration run-list --json
orca orchestration worker-list --json
orca orchestration task-list --run run_legacy_local --json
orca orchestration worker-list --help
orca orchestration dispatch-show --help
orca orchestration run-current --help
orca orchestration worker-list --all --json          (rejected, invalid_argument — evidence for §4)
orca orchestration worker-list --status succeeded,failed,stopped,abandoned --json  (rejected, invalid_argument)
sqlite3 -readonly ~/.config/orca/orchestration.db "select count(*) from worker_dispatches;"   (0 rows)
sqlite3 -readonly ~/.config/orca/orchestration.db ".schema worker_dispatches"
orca repo show --repo id:00000000-0000-0000-0000-000000000000 --json   (repo_not_found — evidence for §4)
orca orchestration worker-show --dispatch dispatch_does_not_exist --json   (dispatch_not_found — evidence for §4)
orca orchestration check --run run_does_not_exist_xyz --json   (run_not_found — evidence for §4)
orca orchestration check --run run_legacy_local --wait --types worker_done --timeout-ms 1500 --json   (run_not_found on a REAL run — the item-4 correction)
orca --help   (full, incl. Behavior section)
orca orchestration run-show --help
orca orchestration run-show --run run_legacy_local --json   (invalid_argument — --id not --run)
orca orchestration inbox --run run_legacy_local --json   (invalid_argument — contradicts capabilities doc §2)
orca orchestration gate-list --run run_legacy_local --json
orca orchestration worker-list --run run_legacy_local --json
orca orchestration task-list --run run_legacy_local --json   (retest)
orca account list --json
orca orchestration task-list --run run_legacy_local --status dispatched --json
orca worktree list --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json   (count re-check: still 9)
```

All commands above are read-only: `--help`, `status`, `--json` list/show
calls, and negative-path lookups against ids/flags known or expected to be
invalid (never a create/send/spawn/remove verb). No worktree, terminal, run,
task, dispatch, or automation was created, sent, spawned, or removed at any
point in this validation pass.
