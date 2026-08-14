# MIGHT-DO — deliberate non-tickets

Follow-ups we chose NOT to file in Linear (Gaetan, 2026-08-11: "don't create
new tickets, include them in MIGHT-DO.md"). One line each, source in parens.
Single writer: the coordinator session. Promote to a ticket only on Gaetan's
word; delete when done or dead.

- Skill-mirror scope: 18 mirrored dirs vs 132 live SKILL.md dirs, so ~114 are
  bundled/third-party and unmirrored (measured 2026-08-14) — decide whole-tree
  mirror vs tracked-18-only. (GCN-13 close-out)
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
- GCN-37 self-closed 4 minutes after a natural checker run filed it
  (2026-08-12) — unexplained, out of R2's scope; worth one look. (R2)
- `execute_code` is blocked UNCONDITIONALLY under cron — not only the chat-TTY
  gate an earlier probe documented; every 08-12 fire fell back to `terminal`.
  Candidate `docs/facts.md` row. (R2)
- GCN-12's checker queue item was consumed by the R2 pull-leg test (reconciled
  `linear:completed`); the board is back to Todo but the reminder loop is dead —
  the checker re-detects only on a fresh Slack/email mention of the OAuth task.
  (R2)
- GCN-25's queue item carries a latched `linear_issue.closed_at` (its Linear
  `canceledAt`), so the new `duplicate → dismissed` transition can never fire
  for that one item; cosmetic (item already terminal locally) — clear by hand
  only if the label matters. (stage-2 apply)
- `gcn30-32-checker-fix.md` §8 P4 guard defect: `wc -c > 20` aborts on a
  legitimately `[SILENT]` extraction; doc fix in the probe text. (stage-2 apply)
- D6 unexercised: the conversation-state contract (SOUL H1'/H2' + checker H3')
  was applied 2026-08-14 but never watched fire — at the next follow-up job or
  report naming a colleague, check the anchor is run-time-resolved and both
  sides' last message are stated. (2026-08-14-damien-followup-apply-report.md §9)
- C5 sibling-ban audit still open: the DM-read carve-out was checked only
  against the 18 mirrored skill dirs — grep the full 132-dir live tree for "do
  not fetch/rescan history" bans before extending it to another skill.
  (2026-08-14-damien-followup-fix-draft-v2.md §8)
- 5 paused cron jobs carry the deleted incident job's "répond…" pattern and
  were left alone pending Gaetan's keep/delete call (`dc2d0e37a2b9`,
  `8bb269c59f66`, `71164c8f953f`, `7eea2c2c8075`, `2f31075e51e3`) — ask at the
  next cron-hygiene pass. (2026-08-14-damien-followup-apply-report.md §7)
- The 33-job cron list (16 completed / 9 scheduled / 8 paused) was never
  recorded whole — 10 jobs are counted but never named; capture `hermes cron
  list --all` into a status file before the next cron-affecting fix, or an
  audit has no baseline to diff. (2026-08-14-damien-followup-apply-report.md §7)
- `context_file_max_chars` is pinned at 40000, replacing the dynamic 65280
  (`context_length × 4 × 0.06` on gpt-5.6-sol) — recheck at every Hermes
  `/upgrade` or model swap whether the dynamic cap now exceeds the pin.
  (2026-08-14-cap-raise-report.md)
- Snapshot the 140 `skill_manage` completion timestamps (of 219 log lines) from
  `~/.hermes/logs/agent.log{,.1,.2}` (~11.9 MB, rotating) before rotation eats
  them — Fix 4's provenance method rests on them; do it now, not on a trigger.
  (2026-08-14-fix4-reconciliation.md, standing caveat)
- A 30d→90d `conversations_history` read may silently blow Hermes' tool-result
  budget (`engagement-checker/SKILL.md:617`) and never reach the model, which
  then reports conversation state it never received — measure it the first time
  a wide silence case fires; fix is the message-count form of `limit`.
  (2026-08-14-fix-review.md B8)
- No prune policy for the mandated `.bak-*` copies: 9 `SOUL.md.bak-*` plus 6 in
  the engagement-checker dir, +2 from the 08-14 apply — set a retention rule
  (keep last N per file) when VM disk becomes a concern.
  (2026-08-14-fix-review-ops.md §0/§2.9)
- `stakeholder-communication` is mirrored but is a plain drafting/recap skill
  with zero last-speaker logic, while `daily-work-brief:127` carries its own
  follow-up wording — reconcile one against SOUL rule 10 once the new paragraph
  has run a few times (post-D6). (2026-08-14-damien-followup-verify.md §B)
- Mirror exclusions `google-workspace` / `i-have-adhd` / `last30days` are
  justified only in a probe (bundled+local patch, p-Hermes copy, third-party
  pack) — CLAUDE.md now states the doctrine but not the per-dir reasons, so a
  strict "non-bundled ⇒ mirror" audit would re-flag them.
  (2026-08-14-fix4-reconciliation.md §5)
- Raw Slack tool-call payloads are not persisted on the VM — only a char count
  (3,789 for the Damien read); the same-day connector re-read that rescued that
  forensic won't work on an older or busier conversation. Weigh persisting tool
  bodies against storing Slack content at rest.
  (2026-08-14-damien-followup-vm-run-trace.md §9)
- The session that actually flipped GCN-49 to Done (`20260813_140745_980067c0`,
  4 `save_issue` calls at 15:10 UTC) was never traced — chase only if a
  double-session write pattern causes a real incident.
  (2026-08-14-damien-followup-vm-run-trace.md §9)
- WF3-era residuals never explicitly closed: `NOTION_API_KEY` prompts `notion`
  headlessly, `tirith` enabled-but-unavailable, `rtk` off the non-interactive
  PATH, Codex chat dying after 3 continuation attempts — confirm status before
  ticketing; the 2026-08-13 v1 close-out may have mooted them.
  (status/lane-a.md, WF3 wiring verdicts)
- No liveness check on the drift-check cron — now that a clean verdict is
  byte-identical run to run, a dead cron reads exactly like "no drift"; revisit if the check ever
  matters more than the mirror does. (scripts/tars-mirror-drift)
- RUNBOOK v2 gate 10 cites `hermes prompt-size --json` as the whole-inclusion
  proof for SOUL.md; it is not one — its `sections[]` reports context=0 because
  SOUL rides the identity slot inside `stable`. The real offline check is
  `prompt_builder.load_soul_md()`; fold it into the next runbook revision.
  (2026-08-14-polish-report.md, surprise 1)
