# Linear live facts — Mobile Club workspace (via claude.ai Linear MCP)

Pulled live 2026-08-10 with `mcp__claude_ai_Linear__*` tools (get_workspace,
list_teams, get_user, list_users, list_issue_labels, list_issue_statuses,
list_issues). All calls read-only, no mutation tools invoked.

## 1. Workspace

```json
{"id":"b2985795-2b8d-46be-8014-7e1d67a887e6","name":"Mobile Club","url":"https://linear.app/mobile-club"}
```
`get_workspace` exposes only id/name/url — **no plan/tier field**. Plan tier is
not visible through this MCP; would need the Linear billing UI or the
`subscription`/`organization` GraphQL fields the MCP doesn't surface.

**8 teams exist today** (`list_teams`, limit 100, `hasNextPage:false` — this is
the full list):

| Team | id | created |
|---|---|---|
| Mobile club (MC) | cb4eb28b… | 2020-07-01 |
| Intelligence Engineering (IE) | 1813aa39… | 2023-10-03 |
| NM x MC (NMC) | 11a6f6b1… | 2026-02-24 |
| Software Engineering (SE) | 139092a5… | 2023-03-10 |
| Cleaq | c6bb5692… | 2022-08-16 |
| Next Mobiles | 79cf278b… | 2026-06-23 |
| Agents (AGE) | a08b0839… | 2026-06-23 |
| Cleaq Design | 6d7f0a13… | 2022-11-24 |

**No precedent of a personal/single-member team.** Every team name maps to a
business unit or engineering discipline (Mobile Club, Next Mobiles, Cleaq,
Cleaq Design, Software Engineering, Intelligence Engineering) or to an
agent/product surface (Agents, NM x MC — the cross-company support-engineer
initiative). Nothing named after a person. `get_team("Mobile club")` returns
only id/icon/name/timestamps — no member-list or team-settings fields exposed
by this MCP (see §6).

## 2. Gaetan's user — role

```json
{"id":"4951b192-…","name":"Gaëtan Cathelain","email":"gaetan.cathelain@mobile.club",
 "isAdmin":true,"isGuest":false,"isActive":true,
 "teams":[Agents, NM x MC, Intelligence Engineering, Software Engineering, Mobile club]}
```

**`isAdmin: true`.** Cross-checked against the full `list_users` dump (29
users): admins are Olivier Thierry, Robin Koenig, Nans Brun Dumortier, Grégoire
Segretain, Jeremy Pinto, Nicolas Duchon, and Gaëtan — a small admin set, not
everyone. Workspace admin in Linear can create teams (Settings → Teams → New
team is an admin-gated action) — **so yes, he can create a team himself**,
personal or otherwise, no request needed. This is a live fact, not inferred
from docs (see caveat in §6: the MCP itself has no `create_team` tool, so this
is a UI/API capability claim I could not exercise, only infer from the
`isAdmin` flag + Linear's documented permission model).

## 3. Workflow states for the 3 teams where his ~40 open issues live

Confirmed via two pages of `list_issues(assignee: me, limit 100)` (200 issues
scanned, `hasNextPage:true` beyond that — not exhaustive, but enough to
confirm which teams and to pull each team's full status list). Open issues
(non-Done/Canceled/Duplicate) appeared on exactly 3 teams: **Mobile club
(MC)**, **NM x MC (NMC)**, **Software Engineering (SE)** — matching Gaetan's
"~40 issues across 3 company teams" claim. Sample open issues: MC-4179 (Triage,
Urgent), MC-4092 (Triage), MC-4000/MC-3975 (In Progress), MC-4112 (Todo);
NMC-601/NMC-622/NMC-498/NMC-497 (In Progress), NMC-606/NMC-589/NMC-615 (In
Review), NMC-624/NMC-623/NMC-594/NMC-539 (In Spec & Design/backlog);
SE-643/SE-609/SE-610 (In Progress), SE-612/SE-611/SE-405/SE-587 (Todo),
SE-679/SE-641/SE-589/SE-369/SE-665 (Backlog).

`list_issue_statuses(team)` per team:

**Mobile club (MC)** — 11 states: Triage(triage), Backlog(backlog), In Spec &
Design(backlog), Todo(unstarted), In Progress(started), Blocked(started), In
Review(started), QA(completed), Done(completed), Canceled(canceled),
Duplicate(duplicate).

**NM x MC (NMC)** — 13 states: Triage(triage), Ready For Grooming(backlog),
Backlog(backlog), In Spec & Design(backlog), Todo(unstarted), In
Progress(started), Blocked(started), In Review(started), QA(started — note:
type `started`, not `completed`, unlike MC's QA), QA Done(completed),
Done(completed), Canceled(canceled), Duplicate(duplicate).

**Software Engineering (SE)** — 9 states: Triage(triage), Backlog(backlog),
Todo(unstarted), In Progress(started), Blocked(started), In
Review(started), Done(completed), Canceled(canceled), Duplicate(duplicate).

**Divergence assessment:** low-to-moderate. All 3 share the same backbone —
Triage → Backlog → Todo → In Progress/Blocked/In Review → Done, plus
Canceled/Duplicate. Divergence is additive, not structural:
- MC and NMC both insert a **backlog-type "In Spec & Design"** step SE lacks.
- NMC adds **"Ready For Grooming"** (backlog) that MC/SE lack.
- NMC's **QA is `started`-type** (still "in flight") while MC's QA is
  `completed`-type — the same status name means a different lane on the two
  teams. A cross-team kanban column keyed on status **name** would silently
  misplace NMC's QA card into a "done-ish" column if it copied MC's semantics.
- SE has no QA state at all.

**Practical implication for a cross-team board:** group columns by `statusType`
(triage/backlog/unstarted/started/completed/canceled/duplicate — Linear's own
6-value enum, present on every team) rather than by status **name**. That
collapses the 9–13 team-specific names into ~5 visible columns (Triage,
Backlog, Todo, In Progress, Done) with `canceled`/`duplicate` filtered out or
merged into an "Archived" swimlane. This is the one design fact this recon
step was meant to settle.

## 4. Labels — workspace-level vs team-level

`list_issue_labels(limit 100)` returned 28 labels, `hasNextPage:false`. Most
carry a `"parent"` field, meaning **team/project-scoped label groups**
(Cleaq's own label group: "Cleaq ID Provider (Keycloak)", "Hubspot", "Cleaq
Back-office (Retool)", "Cleaq Front", "Cleaq Back"; "Risk App" group: "Risk App
Front/BO/API"; "HelpTech" group: "B2C (Next)", "B2C (MC)", "B2B (Cleaq)", "B2B
Mobile Club", "Other").

**Workspace-level labels (no `parent`, so usable/visible across all teams)**:
Tech, Fixed By AI, NiceToHave, Vecna, Monito, Security Incident, Content,
Feature, Bug, Defect, Test, OPS, Documentation, Client feedback, Improvement —
**15 of the 28**. These are the ones a cross-team personal board could rely on
for a shared taxonomy (e.g. filter/tag by "Feature"/"Bug"/"OPS" across MC, NMC
and SE alike) without needing per-team label mapping.

## 5. Existing write-identity pattern: "Support Engineer"

`list_users` / `get_user("supportengineer")`:
```json
{"id":"18f02260-…","name":"Support Engineer",
 "email":"945576c7-6729-4851-b7e2-2a1051c0b384@oauthapp.linear.app",
 "isAdmin":false,"isGuest":false,"isActive":true,
 "teams":[{"name":"Agents","key":"AGE"},{"name":"Mobile club","key":"MC"}]}
```
Its email is an `@oauthapp.linear.app` synthetic address — this is Linear's
**OAuth application user** pattern: a workspace can install an OAuth app
(Linear's own agent-app framework) and Linear mints a first-class `User`
identity for it, scoped to specific teams (here: Agents + Mobile club), not a
guest and not an admin. Two other bot-style users of the same shape exist:
"Cursor" (`…@oauthapp.linear.app`) and "ChatPRD" (`…@oauthapp.linear.app`),
plus a `"linear"` system user (`linear-<workspaceid>@linear.linear.app`, a
Linear-native automation user, different pattern — not an OAuth app).

**Reusable pattern for Tars**: this is exactly the precedent engagement-checker
already uses/needs to reason about — the Support Engineer's Linear identity is
an OAuth-app user with team-scoped membership, not a personal API key
impersonating Gaetan. Giving Tars its own write identity on Linear would mean
registering (or reusing an existing) OAuth app in this workspace and granting
it membership on whichever team(s) it needs to write to — the same shape, no
new pattern to invent.

## 6. What this MCP does NOT expose

- **No `create_team` / team-settings mutation tool.** Team creation, team
  privacy, and per-team membership management are not in the Linear MCP tool
  surface at all (confirmed by the tool list: create_* tools exist only for
  issue/project/initiative/document/attachment/comment/label/status-update/
  release — nothing team-shaped). Any claim about "can Gaetan click-create a
  personal team" rests on the `isAdmin` flag + Linear's documented UI
  permission model, **not** on an MCP call I could exercise. Needs the Linear
  docs/UI to confirm the exact gate (e.g. whether team creation is
  admin-only vs any member, and whether workspace plan tier caps team count).
- **No custom-view / saved-filter tooling.** Linear's "Views" (cross-team
  saved filters, which is Linear's own native answer to "one board across
  teams") are not exposed by any tool here — `list_issues` supports ad-hoc
  filters (assignee/team/label/state/etc.) but there is no `list_views` /
  `create_view` tool to inspect or build a persistent custom view. This
  matters directly for Gaetan's question #3 (cross-team board) — Linear
  *does* natively support a custom view spanning arbitrary teams filtered by
  assignee=me, which may be a zero-build answer to "see my ~40 issues from 3
  teams in one place" without any Hermes/sync work at all — but it has to be
  verified/built in the Linear UI, not through this MCP.
- **No privacy/visibility flags** on teams or workspace (private team? guest
  access scoping?) — `get_team`/`list_teams` return only
  id/icon/name/timestamps, nothing about member visibility rules.
- **No plan/tier/billing field** (see §1).
- **No webhook/integration-config inspection** — can't see from here what the
  rolled-back Linear+Notion+Calendar integration (`config.yaml.bak-linear-notion-cal`)
  actually wired up; that's a VM-side artifact, not visible from Linear's side
  through this MCP.

Net: for the synthesis, treat "personal Linear team, admin-creatable" and
"native cross-team custom view exists" as **plausible, doc/UI-verifiable
claims**, not MCP-confirmed facts — everything else above (workspace identity,
team list, Gaetan is admin, the 3 teams' workflow shapes, label scoping, the
OAuth-app user pattern) is MCP-confirmed live data as of 2026-08-10.
