# Peer mandate — apply P3 (Orca v2 delegation skill) + first live E2E

Gaetan approved applying **P3 + P4**, 2026-08-07. Read
`docs/proposals/P3-orca-v2-skill.md` in full first, then
`artifacts/delegate-to-cooper-SKILL.md` (the replacement) and
`artifacts/wf5-orca-delegation-v2-section.md`. A sibling peer applies P4 (SOUL)
concurrently — see the joint-apply gate.

## You own (nobody else touches these)

- `~/.hermes/skills/delegate-to-cooper/SKILL.md` on the Tars VM — complete
  replacement from `artifacts/delegate-to-cooper-SKILL.md`
- `docs/specs/wf5-orca-delegation.md` — append the v2 section from
  `artifacts/wf5-orca-delegation-v2-section.md`, keep v1 documented as superseded
- The cooper-side v1 leftovers: `~/orca/workspaces/tars-delegated/` (`delegate.sh`,
  `.claude/settings.json`, `NOTE.md`). v2 does not use them. **Do not delete
  anything** — v1 shipped hours ago; mark it superseded in `NOTE.md` and report
  what could be removed later, for Gaetan's word.

## Settled sub-decisions — apply the proposal's defaults

- **P3-a worktrees: KEEP.** Release the dispatch, keep the worktree, report path +
  branch. No auto-`rm` (it risks destroying delegated work).
- **P3-b rename: NO.** Ship as `delegate-to-cooper` at the existing path — a
  zero-risk drop-in. Do not rename; Gaetan has not ruled on it.
- **P3-c:** the delivery-id field name on a non-empty batch is unobserved — the
  skill tells Tars to read the raw JSON; your E2E is what settles it. Record the
  real shape in evidence and correct the spec.

## You must NOT

- Do not touch `~/.hermes/SOUL.md` (sibling peer) or `~/.hermes/config.yaml`.
- Do not restart the gateway — none is needed (a dropped `SKILL.md` rebuilds the
  skills index via the path→(mtime,size) manifest) and it would disrupt the sibling.
- Do not edit `status/lane-a.md` (single-writer: the hub).

## Procedure

1. **`cp ~/.hermes/skills/delegate-to-cooper/SKILL.md` → `.bak-p3` FIRST.** The v1
   file exists **only on the VM, never in git** — that backup is the sole rollback
   path. Verify the copy before proceeding.
2. Apply under `flock ~/.hermes/.wf3.lock`, atomic write, preserve mode.
3. Confirm registration: `~/.local/bin/hermes skills list` shows the entry and the
   skill count moved as expected.

## Joint-apply gate (mandatory)

The file you are replacing currently contains *"Never rm, rmdir, mv, chmod, sudo,
systemctl on cooper"*, which contradicts the sibling's new SOUL rule 7. Therefore:

1. Prepare everything up to the point of writing (backup taken, command ready).
2. **SendMessage the hub: "P3 READY".** Then wait.
3. The hub replies **GO** once the sibling is also ready. Apply immediately on GO,
   then report **"P3 APPLIED"** with the timestamp.
4. On failure: restore the `.bak` under `flock`, report, and tell the hub so the
   sibling can roll back too.

## Then: the first live E2E (Gaetan confirmed mc-metarepo as the target)

One **read-only** brief into `mc-metarepo` through the full v2 loop:
`run-create` → `task-create` → `worker-start --agent claude --repo mc-metarepo` →
`check --run <id> --wait --types worker_done` → `--ack` → `worker-read` → report.

- Drive it **from Tars**, not by hand — the point is that the model uses the skill.
  Trigger with one Slack message as Gaetan (native creds; the claude.ai connector
  cannot trigger Tars) or a `hermes chat -Q -q` turn, and say which you used.
- Prove Tars did it: `agent.log` must show the real terminal/tool executions for
  the `ssh cooper … orca …` commands, and cooper must show the run/worker.
- Exercise the checkpoint story if the turn exceeds ~300 s: the `hermes cron
  create … --deliver slack --repeat 1` re-attach path, and `--ack` so the same
  delivery is not replayed forever.
- Known intermittent: multi-step Codex turns can die with "response remained
  incomplete after 3 continuation attempts" — retry once before concluding failure.

## Deliverables

- Evidence → `status/probes/wf5/apply-p3-orca-v2.md` (backup path, apply
  timestamp, skills-list output, full E2E transcript with run/dispatch ids, the
  observed delivery-id field name, cooper-side artifacts).
- Commit + push (`git pull --rebase origin main && git push origin HEAD:main`).
- SendMessage the hub at: P3 READY · P3 APPLIED · E2E verdict · done.

All hard rules in `CLAUDE.md` and `docs/facts.md` bind you.
