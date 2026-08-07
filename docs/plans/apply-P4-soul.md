# Peer mandate — apply P4 (SOUL / guidelines redesign)

Gaetan approved applying **P3 + P4**, 2026-08-07. Read `docs/proposals/P4-soul-guidelines-redesign.md`
in full first, then `artifacts/SOUL-proposed.md` and `artifacts/guidelines-changelist.md`.
A sibling peer applies P3 (the Orca v2 skill) concurrently — see the joint-apply gate.

## You own (nobody else touches these)

- `~/.hermes/SOUL.md` on the Tars VM (complete replacement from `artifacts/SOUL-proposed.md`)
- The repo-doc changes D1–D5 from the change list (repo side, no VM impact)
- `README.md`'s stale absolute *"No PR is ever created by Tars"* → reconcile to the
  settled reading: **Tars does not do the work itself**; a PR resulting from a
  Tars-initiated Orca task is fine.

## You must NOT

- **`approvals:` (change B1) is PARKED** — Gaetan has not ruled between
  `manual` / `smart` / `off` / route-into-Slack. Do not add the block, do not
  change approval behavior. Leave it for his decision.
- Do not touch `~/.hermes/skills/` (sibling peer owns the skill) or
  `~/.hermes/config.yaml`.
- Do not restart the gateway — **no restart is needed** (SOUL.md is read fresh at
  every system-prompt build) and a restart would disrupt the sibling.
- Do not edit `status/lane-a.md` (single-writer: the hub).

## Procedure

1. Diff `artifacts/SOUL-proposed.md` against the live `~/.hermes/SOUL.md` and
   confirm what actually changes. **Rule 4 (identity frame + the minimal reply
   `·`, U+00B7) must be byte-identical to live** — verify programmatically, not by
   eye; that rule is the fix for today's empty-response incident.
2. `cp ~/.hermes/SOUL.md ~/.hermes/SOUL.md.bak-p4` (0600 not required; SOUL carries
   no secret, but note it is currently mode 664).
3. Apply under `flock ~/.hermes/.wf3.lock` — write the new file atomically
   (temp + `mv`), and **preserve the existing mode** (an earlier agent's
   `os.replace` silently dropped a file from 600 to 664; do not repeat it).
4. Confirm the model actually sees it: one `~/.local/bin/hermes chat -Q -q` turn
   whose answer can only come from the new text (e.g. ask Tars to state, in its
   own words, whether reading a PR diff is allowed — new rule 2 says yes, the old
   said no).

## Joint-apply gate (mandatory)

The live skill still contains *"Never rm, rmdir, mv, chmod, sudo, systemctl on
cooper"*, which contradicts new SOUL rule 7. Therefore:

1. Prepare everything up to the point of writing (backup taken, diff verified,
   command ready).
2. **SendMessage the hub: "P4 READY".** Then wait.
3. The hub replies **GO** once the sibling is also ready. Apply immediately on GO,
   then report **"P4 APPLIED"** with the timestamp.
4. If anything fails, restore the `.bak` under `flock` and report — do not leave a
   half-applied state, and tell the hub so the sibling can roll back too.

## Verify after applying

- One Slack turn as Gaetan (native creds — the claude.ai connector cannot trigger
  Tars, see `docs/facts.md`): Tars still answers, i.e. the identity frame survived.
- A non-Gaetan sender still trips `[Slack] Early reject of unauthorized user …` —
  grep the adapter log; never settle for absence-of-reply. **This control is now
  the sole gate on transitive Orca access; it must not lapse.**
- No `Empty response` warnings in `~/.hermes/logs/errors.log` after the change.

## Deliverables

- Evidence → `status/probes/wf5/apply-p4-soul.md` (diff summary, backup path,
  apply timestamp, verification output).
- Commit + push (`git pull --rebase origin main && git push origin HEAD:main`).
- SendMessage the hub at: P4 READY · P4 APPLIED · verification verdict · done.

All hard rules in `CLAUDE.md` and `docs/facts.md` bind you.
