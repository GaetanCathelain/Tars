# Audit — side effects of the P6 baseline probe turns (audit only, nothing cleaned up)

Scope: audit only. Nothing was changed, deleted, killed, or closed. All commands
below are read-only (`git status`, `git ls-remote`, `gh pr list`/`gh api` GETs,
`orca … list/show`, `ps`, `ls`). Machine: cooper. VM (192.168.0.9) was not
touched for this audit — everything needed lives in the probe file and in
cooper's own Orca/git state.

---

## What the probe did

Source: `status/probes/wf5/kb-retrieval-trials.md`, "Baseline — what Tars
answers TODAY" section. Five oneshots (`hermes chat -Q -q`) sent 2026-08-07
22:45–23:15 UTC, no map/skill, exactly as Gaetan would DM them.

| Q | session | wallclock | verdict | what happened |
|---|---|---|---|---|
| Q1 (NM vs MC order) | `20260807_225044_b6d355` | **422 s** | partial (missed the trap) | Tars delegated to an **Orca worker on cooper**, which checked out `mc-metarepo` and read live product source (`packages/api-v2/...`) instead of reading the 767-byte note that already answers it |
| Q2 (Datadog 92%) | `20260807_225810_7d8488` | **414 s** | correct | Tars delegated to a second **Orca worker on cooper**, which ran a **live production Datadog investigation** (spans, unsampled logs, 5 monitors) instead of reading the 971-byte note |
| Q3 (API-key rotate) | `20260807_225812_fd1da1` | 22 s | **wrong** | zero retrieval, answered from priors, contradicts the KB's explicit "don't clean the logs" rule |
| Q4 (learning → metarepo) | `20260807_225816_2a3dac` | 113 s | correct | ~42 KB of blind `terminal` grepping, no delegation |
| Q5 (bon de retour) | `20260807_225817_764c60` | 62 s | correct | answered from `lcm.db` conversation memory, not the KB, no delegation |

Only Q1 and Q2 delegated to Orca and are the source of the flagged side
effects. Per the probe file's own flag: `write_file` calls by these two turns
landed in `/tmp` only, and the probe author verified nothing under
`~/.hermes/` or the VM's mc-metarepo clone was written.

---

## Orca runs/tasks/workers

Measured with `orca --help` first (flags below are as the installed CLI
actually accepts — `run-show` takes `--id`, not `--run`; `task-list` takes
`--run`).

| Run id | Objective | Task | Status | Created (UTC) | Completed (UTC) | Coordinator terminal / workspace |
|---|---|---|---|---|---|---|
| `run_33222715552e` | "Identify how api-v2 distinguishes Next Mobiles from Mobile.Club orders" (**= Q1**) | `task_4f9e05d18e4e` | **completed / succeeded** | 22:51:55 (run) / 22:52:00 (task) | 22:55:46 | `term_4a2f5ec0…` in `/home/gaetan/orca/workspaces/Tars/apply-p3-orca` |
| `run_9a6a5e637304` | "Investigate Vecna API error alert" (**= Q2**) | `task_dd61dc750848` | **completed / succeeded** | 22:59:25 (run) / 22:59:31 (task) | 23:03:49 | task's `created_by_terminal_handle` = `term_479c7b7b…` in `/home/gaetan/orca/workspaces/Tars/orchestrator` (run itself shows `coordinator_handle: null` in `run-list`, i.e. currently unbound) |

Both orchestration-layer tasks are marked `completed` / `outcome: succeeded`.
**Neither run/task is "still running" at the orchestration layer** — but see
"Still running" below: the underlying OS processes and Orca terminals (ptys)
were never stopped.

Worker reports (both `outcome: succeeded`, both explicitly read-only per their
task spec):
- Q1 worker report: `/tmp/claude-1000/-home-gaetan-dev-mc-metarepo-null-mc-metarepo-api-v2-order-origin/46ad85c3-.../scratchpad/order-origin-report.md` (4190 B, present)
- Q2 worker report: `/tmp/claude-1000/-home-gaetan-dev-mc-metarepo-null-mc-metarepo-vecna-post-order-status/35969e34-.../scratchpad/vecna-postorderstatus-verdict.md` (4263 B, present)
- VM-side brief files named in the probe are also present: `/tmp/tars-brief-api-v2-order-origin.md` (774 B), `/tmp/tars-brief-vecna-post-order-status.md` (1727 B), `/tmp/verify-api-v2-order-origin.py` (664 B) — all under `/tmp`, none under `~/.hermes/` or the clone.

**Two more Orca runs exist in the same time window, NOT attributable to the
probe's 5 questions** (found while enumerating `orca orchestration run-list`;
flagging for completeness since the task asked for "any run/task/worker
attributable" and these are adjacent enough to be confusable):
- `run_b08c7d3d6571` "Read-only map of every top-level directory in mc-metarepo", 22:35:16–22:41:28, created by `term_479c7b7b…` (Tars/orchestrator workspace) — before the probe's 22:45 start, no worktree/branch created (task spec said "write no report file").
- `run_a0a34f5ec451` "Summarize the five most recent mc-metarepo commits…", 23:21:47–23:24:43, created by the same `term_479c7b7b…` — after the probe's 23:15 end, produced worktree `latest-five-commits-readonly-2` (branch `GaetanCathelain/latest-five-commits-readonly-2`). Both timestamps sit outside the probe's own stated 22:45–23:15 UTC window and both were created by the Tars/orchestrator hub terminal, not by a baseline-Q1/Q2-style delegation from a `hermes chat` turn. Included below in the worktree table for completeness, but it is a separate artifact, not one the probe caused.

---

## Worktrees & branches

| Path | Branch | Dirty? | Pushed to origin? |
|---|---|---|---|
| `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/api-v2-order-origin` | `GaetanCathelain/api-v2-order-origin` | **No** — `git status --short` clean, `git diff --stat` empty, HEAD = `4831253` (tip of `main`, zero unique commits) | **No** — `git ls-remote origin refs/heads/GaetanCathelain/api-v2-order-origin` returns nothing |
| `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/vecna-post-order-status` | `GaetanCathelain/vecna-post-order-status` | **No** — same, clean, HEAD = `4831253` | **No** — `git ls-remote` returns nothing |
| `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/latest-five-commits-readonly-2` (separate cause, see above) | `GaetanCathelain/latest-five-commits-readonly-2` | **No** — clean, HEAD = `4831253` | **No** — `git ls-remote` returns nothing |

All three branches were cut from `main`'s current tip and never advanced —
each worker was genuinely read-only, matching its task spec's "do not modify
files or create commits" constraint. `orca worktree ps` confirms `api-v2-order-origin`
and `vecna-post-order-status` still have a **live pty** each; `latest-five-commits-readonly-2`
does not appear in `worktree ps` at all — no live terminal remains for it.

---

## Did anything reach a remote?

**No.**

- `git ls-remote origin refs/heads/GaetanCathelain/api-v2-order-origin refs/heads/GaetanCathelain/vecna-post-order-status refs/heads/GaetanCathelain/latest-five-commits-readonly-2` → **empty output**, i.e. none of the three branches exist on `origin` (`mobile-club/metarepo`).
- `gh pr list --repo mobile-club/metarepo --state all --search "head:GaetanCathelain/api-v2-order-origin"` → `[]`
- `gh pr list --repo mobile-club/metarepo --state all --search "head:GaetanCathelain/vecna-post-order-status"` → `[]`
- `gh pr list --repo mobile-club/metarepo --state all --search "head:GaetanCathelain/latest-five-commits-readonly-2"` → `[]`
- Sanity check, the 15 most recent PRs on `mobile-club/metarepo` (`gh pr list --state all --limit 15`): the latest is **#182**, opened 2026-08-07 20:56:30 UTC by `Robin-Koenig` — **before** the probe's 22:45 UTC start. No PR numbered after #182 exists. No PR/issue references either probe branch or worktree name.
- `gh auth status` confirms the token used is the interactive `GaetanCathelain` GitHub account (repo/workflow scopes), and `gh repo view` confirms the local clone targets `mobile-club/metarepo` — so this is the right remote to check.

Nothing was pushed, no PR or issue was opened. This is not an incident.

---

## Datadog

Per the probe file (baseline Q2 section) and the Q2 worker's own completion
report (`task_dd61dc750848` result, `orca orchestration task-list --run
run_9a6a5e637304`):

- The task spec explicitly constrained the worker to read-only investigation:
  *"Do not request or reveal credentials. This is an investigation only: make
  no code, config, deployment, infrastructure, or data changes… Do not submit
  or mutate an order."*
- The worker's reported actions, all reads: span counts (594 OK / 1 error over
  15 days), unsampled log counts (2 "Vecna API error" lines in 15 days, both
  deterministic 404s), and a check of all 5 Vecna monitors (all reporting OK).
  It also **noticed** (did not fix) that monitor `110136431` still filters
  `env:production` and is likely blind — reported as a follow-up suggestion,
  not acted on.
- Tool trace cited in the probe file: 16 calls, all `terminal` plus one
  `browser_navigate` and one `write_file` (the write landed in `/tmp`, per the
  side-effects note at the top of the probe file).

**It only read. No monitor, dashboard, or any other Datadog object was
created, edited, muted, or deleted**, per both the task's own constraints and
its self-reported actions. Caveat, already flagged in the probe file's own
"Not tested" section and repeated here rather than re-litigated: the specific
span/log counts were not independently re-queried by the probe author or by
this audit — I have no direct read access to that Datadog org from this
session, so this section relies on the worker's self-report plus the
mandate's terms, not on an independent Datadog-side check.

---

## Still running

Orchestration layer says both tasks are `completed`/`succeeded` — but the
underlying OS processes and Orca terminals were never stopped:

| PID | Cmd | cwd | Elapsed at check | Orca terminal | `connected` |
|---|---|---|---|---|---|
| 1011020 | `claude --dangerously-skip-permissions` | `.../null/mc-metarepo/api-v2-order-origin` | ~42 min | `term_2714788a-4262-4a27-947f-f52484e3db83` | `true` |
| 1044130 | `claude --dangerously-skip-permissions --resume 35969e34-3c6b-470c-b776-a82d3b04f644` | `.../null/mc-metarepo/vecna-post-order-status` | ~28 min | `term_742e71a7-fb5b-4562-b174-c05449c9d927` | `true` |

Both terminal previews show the Claude Code TUI idle at its bottom-line
prompt, past the point where they sent `worker_done` — i.e. they finished
their work and are sitting parked, not looping. `ps aux` CPU% for both is
≤1%. Nothing was killed. No live process/terminal exists for
`latest-five-commits-readonly-2` — that one already has no pty.

No tmux server is running on cooper (`tmux list-sessions` → "no such file or
directory" — this box doesn't use tmux for Orca; Orca manages its own ptys
directly, shown above).

---

## Recommended cleanup (for Gaetan to approve — NOT performed)

1. **Stop the two idle Claude Code sessions** — `orca terminal stop` (or
   `terminal close`) for `term_2714788a-4262-4a27-947f-f52484e3db83` and
   `term_742e71a7-fb5b-4562-b174-c05449c9d927` (equivalently, kill PID
   1011020 and PID 1044130). They're idle and harmless but each is a live,
   `--dangerously-skip-permissions` Claude Code session sitting in a real
   git checkout for no ongoing reason.
2. **Remove the two (or three, including the unrelated
   `latest-five-commits-readonly-2`) worktrees** with `orca worktree rm`,
   which also deletes the local-only branch. Safe: nothing pushed, nothing
   uncommitted, HEAD == `main` tip in all three, so this is a no-loss,
   fully-local cleanup.
3. **No remote cleanup needed** — confirmed nothing reached
   `mobile-club/metarepo` (no branch, no PR, no issue).
4. **/tmp scratch files** (`order-origin-report.md`,
   `vecna-postorderstatus-verdict.md`, `verify-api-v2-order-origin.py`,
   `tars-brief-*.md`) can be left — they're already outside any tracked repo
   and outside `~/.hermes/`, per the probe's own verification.
5. Optional, not urgent: the probe file itself already recommends fixing
   `SOUL.md`'s dead `~/dev/gaetan-metarepo` pointer and building the
   `SKILL.md`-map retrieval mechanism — those are P6 design decisions, not
   cleanup, and are out of scope for this audit.
