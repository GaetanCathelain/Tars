# WF6 — Tars ⇄ personal Linear team integration

Status: ACTIVE — all three gates resolved 2026-08-10 (evidence:
`status/probes/wf6/`). Author:
coordinator session (megaultracode-orca run, 2026-08-10). Decision basis:
`docs/recon/r9-kanban-unification.md` + `docs/recon/r9a-personal-linear-team.md`
(§6 decision update: Business plan; personal team chosen).

## Goal (verbatim from Gaetan, /goal 2026-08-10)

Tars fully integrated with the new personal Linear team; daily-work-brief and
engagement-checker single source of truth is that team. engagement-checker
pushes and pulls to Linear. daily-work-brief pulls from Orca + Claude history
and Linear. Tars reads every ticket across every team but defaults to creating
tickets in the personal team unless told otherwise; same default for
Claude/Orca sessions. Workload labels (fixes, investigations, discovery,
giving-access, …) and priority-sorted tickets.

## Design decisions (settled — do not re-litigate)

- Linear is the one board. No sync engine, no kanban.db, no Linear MCP server
  on the VM — direct GraphQL only (the path that survived the 2026-08-07
  rollback, r9a §3).
- Personal team: **"Gaetan", key GCN**, id `81e7b769-2a46-4e2a-8db5-c165a7963b0e`
  — created by Gaetan himself in the UI 2026-08-10 16:08Z, private, single
  member. Adopted as-is (a second personal team would split the SSoT).
- SSoT semantics: engagement-checker's tracked items BECOME issues in GCN; its
  JSON file demotes to cache/id-mapping only. Nag-loop guard is mandatory:
  items engagement-checker itself filed must not be re-detected as new loops.
- Labels: team-scoped in GCN (not workspace — don't pollute company space):
  `fix`, `investigation`, `discovery`, `access`, `report`, `test-check`
  (Gaetan may rename/extend; these seed the board).
- Priority: Linear native priority (0–4) set on every created ticket; all
  views/renders sorted by priority.
- Cross-team read: already live via the VM's GraphQL path (engagement-checker
  reads all teams). Default-create-in-GCN is a rule, not a permission wall.
- Tars writes as Gaetan (his API key) — accepted in r9a §6; key scope is the
  blast-radius control (gate G2).

## Gates — ALL RESOLVED 2026-08-10 (evidence: `status/probes/wf6/`)

- **G1 team ✅**: team **Gaetan/GCN** exists (created by Gaetan in the UI,
  private, single-member, 0 issues at adoption).
- **G2 key ✅ write-capable**: reversible `favoriteCreate` probe succeeded from
  the VM; no scope error. Linear personal keys are unscoped → the key is
  full-write-as-Gaetan; the collector's read-only nature is enforced by its own
  code allowlist (rejects "mutation"), not by the key. Accepted risk per r9a §6;
  no second key minted.
- **G3 skills ✅**: repo mirrors and VM copies byte-identical (no
  reconciliation). engagement-checker's collector is inline Python in SKILL.md
  (urllib, 3-op allowlist, viewer-assigned issues only → must widen for
  cross-team read). daily-work-brief ALREADY reads cooper Orca/Claude history
  over `ssh cooper` (orca CLI, Claude transcripts, shell history, git log) —
  keep, don't rebuild. Default-team rule goes in skill logic, NOT SOUL (its
  Phase 2 preferences section is an explicit empty placeholder). Cooper-side
  `to-tickets` (matt-pocock plugin) is the only other ticket-creating skill;
  it has zero team-default logic today. All three crons deliver to the same
  Slack thread anchor `slack:C0BP2GZUFSR:1786359613.979759`.

## Tickets (filed in GCN, priority-sorted; implementation structure)

| Key | Title | Label | Prio | Depends |
|---|---|---|---|---|
| T1 | Bootstrap GCN team (adopted; Gaetan created it in UI) + labels + priority view | access | Urgent | — |
| T2 | LINEAR_API_KEY scope check on the VM (resolved: full write) | access | Urgent | — |
| T3 | engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard | fix | High | T1,T2 |
| T4 | daily-work-brief: pull Linear board + Orca/Claude history, priority-sorted render | fix | High | T1,T2 |
| T5 | Tars default-team rule (SOUL + affected skills): create in GCN unless told otherwise | investigation | High | T1 |
| T6 | cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default | discovery | Medium | T1 |
| T7 | E2E verification + docs/facts.md + status/lane-a.md updates | test-check | High | T3–T6 |

## Sessions (megaultracode-orca lanes)

Seam analysis: T3/T4/T5 all touch the VM (`~/.hermes/skills/orchestration/*`,
SOUL.md, config) and the same repo mirror dirs → **one session, sequenced**
(T5 → T3 → T4: the default-team rule first, skills build on it). T6 touches
only cooper-side config (gaetan-metarepo / orca) → parallel session. T1/T2 are
coordinator+operator bootstrap, before dispatch. T7 runs in the VM session
after its tickets, plus coordinator's own independent verification.

- **Session A** `wf6-vm-skills` (subworktree of Tars repo): T5, T3, T4, T7.
  House rules: VM edit discipline (`.bak` + flock, merge-not-append), per-key
  SOPS extract only, no secret ever printed; skills land in repo mirror AND on
  VM — state which side is authoritative during the change window.
- **Session B** `wf6-cooper-defaults` (subworktree of Tars repo for status,
  edits land in gaetan-metarepo): T6. Interactive-Claude-with-Gaetan authority
  covers the preferences edit (AGENTS.md single-writer rule).

Question channel: `~/.claude/handoff/gang/<ts>-wf6-linear/{questions,answers,status}/`.
Operator checkpoints pre-wired into briefs: G2 key minting (if needed), any
SOUL rule wording change, anything touching company teams' data.

## Verification (T7 pass criteria — evidence, not claims)

1. DM Tars "create a ticket to test X" → issue appears in GCN with label +
   priority; Tars replies with key/URL.
2. One engagement-checker cycle: a detected loop lands as a GCN issue; its
   next cycle does NOT re-report the item it filed; closing the issue in
   Linear clears it from the queue (pull works).
3. daily-work-brief 08:30 output (or manual run): opens with the GCN+assigned
   board, priority-sorted, and includes an Orca/Claude-history section.
4. A ticket created from a Claude/Orca session on cooper defaults to GCN.
5. All changes committed: Tars repo (skills mirrors, this spec, status log),
   gaetan-metarepo (preference line). VM live-reload confirmed, `NRestarts=0`.
