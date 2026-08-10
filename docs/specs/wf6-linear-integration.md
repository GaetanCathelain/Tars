# WF6 — Tars ⇄ personal Linear team integration

Status: DRAFT (gates pending `linear-integration-recon` workflow). Author:
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
- Personal team: private (Business plan), single member. Working name
  **"Personal"**, key **PER** (rename in UI is free).
- SSoT semantics: engagement-checker's tracked items BECOME issues in PER; its
  JSON file demotes to cache/id-mapping only. Nag-loop guard is mandatory:
  items engagement-checker itself filed must not be re-detected as new loops.
- Labels: team-scoped in PER (not workspace — don't pollute company space):
  `fix`, `investigation`, `discovery`, `access`, `report`, `test-check`
  (Gaetan may rename/extend; these seed the board).
- Priority: Linear native priority (0–4) set on every created ticket; all
  views/renders sorted by priority.
- Cross-team read: already live via the VM's GraphQL path (engagement-checker
  reads all teams). Default-create-in-PER is a rule, not a permission wall.
- Tars writes as Gaetan (his API key) — accepted in r9a §6; key scope is the
  blast-radius control (gate G2).

## Gates (from recon — resolve before dispatch)

- **G1 team**: does a personal team already exist? If not, create via GraphQL
  `teamCreate` (VM key) or MCP if capable; else operator step.
- **G2 key**: is the VM `LINEAR_API_KEY` write-capable (favoriteCreate probe)?
  If read-only → operator mints a scoped write key (create-issue,
  create-comment; team-restricted if the UI allows) → SOPS ingest
  (`scripts/tars-secret LINEAR_API_KEY_RW`) → per-key deliver to VM.
- **G3 skills**: exact current mechanics of both skills (two-sided diff repo
  mirror vs VM) — which side moved; how the VM reaches cooper for Orca/Claude
  history.

## Tickets (filed in PER, priority-sorted; implementation structure)

| Key | Title | Label | Prio | Depends |
|---|---|---|---|---|
| T1 | Bootstrap PER team + labels + priority view | access | Urgent | — |
| T2 | LINEAR_API_KEY scope check / scoped RW key on VM | access | Urgent | — |
| T3 | engagement-checker: Linear push+pull, SSoT=PER, nag-loop guard | fix | High | T1,T2 |
| T4 | daily-work-brief: pull Linear board + Orca/Claude history, priority-sorted render | fix | High | T1,T2 |
| T5 | Tars default-team rule (SOUL + affected skills): create in PER unless told otherwise | investigation | High | T1 |
| T6 | cooper defaults: Claude/Orca sessions create Linear tickets in PER by default | discovery | Medium | T1 |
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

1. DM Tars "create a ticket to test X" → issue appears in PER with label +
   priority; Tars replies with key/URL.
2. One engagement-checker cycle: a detected loop lands as a PER issue; its
   next cycle does NOT re-report the item it filed; closing the issue in
   Linear clears it from the queue (pull works).
3. daily-work-brief 08:30 output (or manual run): opens with the PER+assigned
   board, priority-sorted, and includes an Orca/Claude-history section.
4. A ticket created from a Claude/Orca session on cooper defaults to PER.
5. All changes committed: Tars repo (skills mirrors, this spec, status log),
   gaetan-metarepo (preference line). VM live-reload confirmed, `NRestarts=0`.
