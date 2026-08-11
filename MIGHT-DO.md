# MIGHT-DO — deliberate non-tickets

Follow-ups we chose NOT to file in Linear (Gaetan, 2026-08-11: "don't create
new tickets, include them in MIGHT-DO.md"). One line each, source in parens.
Single writer: the coordinator session. Promote to a ticket only on Gaetan's
word; delete when done or dead.

- Skill-mirror scope: 119/128 live skills are bundled/third-party and
  unmirrored — decide whole-tree mirror vs tracked-9-only. (GCN-13 close-out)
- CI guard rejecting empty `skills/**/SKILL.md` blobs — the enforced version of
  the GCN-13 recipe fix; blocked today because required checks would break the
  direct `push HEAD:main` flow. Needs a flow decision first. (GCN-13)
- `LINEAR_API_KEY` is unscoped full-write; the GCN-10 prune bounds Tars' reach,
  not the credential. Scope or rotate to a restricted key. (GCN-10 residual)
- Linear file-upload capability lost by the prune (`create_attachment_from_upload`,
  `prepare_attachment_upload`) — re-add to `tools.include` only when a real
  caller appears. (GCN-10, accepted loss)
- R1: inbound-Slack full-loop E2E — one literal DM from Gaetan to Tars, capture
  protocol ready at `status/probes/gcn7-wf6-e2e.md` §9.1. (GCN-7 residual)
- mc orchestration skills (`delegate-orca`, `handoff-orca`, `megaultracode-orca`)
  are local-only under git-ignored `.claude/` because this repo is PUBLIC —
  commit them, keep local, or make the repo private. (2026-08-11)
- `operating-tars` skill header cites a stale authoritative-repo path
  (`~/orca/workspaces/Tars/orchestrator`); refresh to the live layout.
  (noticed 2026-08-11)
- `hermes tools disable/enable` writes `tools.exclude`, a silent no-op while
  `tools.include` exists; `hermes mcp configure` with all boxes ticked deletes
  both filters. Recorded in `docs/facts.md`; consider an upstream Hermes issue.
  (GCN-10)
- engagement-checker reconciles a filed issue only when the collector happens to
  re-pull it; one that closes while the cursor is held is never reconciled.
  Smallest fix: per-run read-back of non-terminal items carrying `linear_issue`,
  with an explicit per-run cap (~4–6 reads today). (GCN-30 lane)
- `skills/productivity/google-workspace/` is NOT mirrored in the repo (vendored
  upstream skill): a `hermes` upgrade can silently revert the 2026-08-11
  venv-interpreter fix — mirror it, or re-apply post-upgrade (also recorded in
  the upgrade memory). (GCN-31)
- `last_failure_notice.linear` is written for a state §8 declares exempt
  ("neither consulted nor written for"); any run inside the 4h window that
  honours §8 will suppress the disclosure while cursors freeze. (GCN-31)
- Ship the §4 collector as `scripts/linear_collector.py` in the skill instead of
  a fenced block — removes the write_file materialization step entirely and
  matches §3 / daily-work-brief practice. (GCN-31)
- Instrumentation to make collection coverage decidable: log the blocked command
  string on the cron-deny branch, log a bounded `final_response`, emit a per-run
  `sources[].{attempted,succeeded,blocked_reason}` coverage record — so
  "collected nothing" is distinguishable from "empty window". (GCN-31)
- Unroot-caused: some cron sessions hit the `pending_approval` branch with
  seconds-long hangs instead of the documented instant cron-deny block. (GCN-31)
