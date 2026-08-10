# Linear landscape for Gaetan (research for the unified-kanban question)

Source: `mcp__claude_ai_Linear__*` tools (claude.ai Linear connector), queried 2026-08-10.

## 1. Workspace / teams / projects — no personal space

- `get_workspace` → single workspace: **"Mobile Club"** (`linear.app/mobile-club`), id
  `b2985795-2b8d-46be-8014-7e1d67a887e6`. This is the company workspace — there is
  no separate personal Linear org for `gaetan.cathelain@mobile.club`.
- `list_teams` → 8 teams, all company-scoped, none personal:
  Mobile club (MC), Intelligence Engineering (IE), NM x MC (NMC), Software
  Engineering (SE), Cleaq, Next Mobiles, Agents (AGE), Cleaq Design.
  **No "Gaetan personal" or "Tars" team exists.**
- `list_users` filtered on "gaetan" → exactly one user, Gaëtan Cathelain
  (`4951b192-e49c-4b7e-b491-58c89e66043c`), `isAdmin: true`. Confirms he's a normal
  seat-holding admin in the company org, not a guest.
- `list_projects` (member=me) → 7 active/backlog projects he's a member of, all
  under company teams (SE, IE, MC, NMC, AGE): e.g. "Integration PIM Quable",
  "🫰 Infrastructure costs", "Agile environments", "Agent Support Technique".
  Nothing resembling a personal to-do project.

**Conclusion for Q1: no personal team/project exists or is implied by the data model.
Anything "personal" (Tars-created items, his to-check/to-test list) would have to
live inside the Mobile Club company workspace — visible to the whole org unless
scoped with team/project/label visibility controls (Linear has no truly private
per-user space short of "only me" issue assignment + restricted team membership).**

## 2. Board states (kanban columns) per relevant team

`list_issue_statuses` — states are per-team workflows, not global:

**Software Engineering (SE)** — 9 states:
Triage → Backlog → Todo → In Progress → In Review → Blocked → Done / Canceled / Duplicate

**Agents (AGE)** — 7 states (leaner):
Backlog → Todo → In Progress → In Review → Done / Canceled / Duplicate

(NM x MC team, where most of the Tars/Hermes work actually lives, showed richer
custom states in the issue list: Backlog, **In Spec & Design**, Todo, In Progress,
In Review, Done, Canceled, Duplicate, Triage — teams customize columns freely.)

**Takeaway:** Linear's per-team kanban already gives Backlog/Todo/In Progress/In
Review/Done/Blocked out of the box, customizable per team. This is a superset of
what a personal to-check/to-test list needs (that maps cleanly to
Backlog→Todo→In Progress→Done, or reuses "In Review" as a "to check" state).

## 3. Issues assigned to Gaetan — rough counts

`list_issues(assignee=me)`, first page of 250 (of a larger total, `hasNextPage: true`,
ordered by updatedAt desc, includeArchived both tried):

By status (250-issue sample, includeArchived=true):
- Done: 181, Canceled: 24, Todo: 15, In Progress: 9, Backlog: 9,
  In Spec & Design: 5, In Review: 3, Duplicate: 2, Triage: 2

By team (same sample): NM x MC 133, Mobile club 62, Software Engineering 54,
Intelligence Engineering 1.

With `includeArchived=false`, scanning the returned (non-archived) issues for
non-terminal states (Todo/In Progress/In Review/Backlog/In Spec & Design/Triage)
gives **~40 currently-open issues** assigned to Gaetan, split roughly:
- NM x MC (the Tars/Hermes/HT-console project): ~15 open — In Progress (NMC-601
  GoCardless token rotation, NMC-622 staging deploy breakage, NMC-498/497 stop-
  invoicing gate), In Review (NMC-606, NMC-615, NMC-589), Backlog/"In Spec &
  Design" (NMC-624, NMC-623, NMC-594, NMC-539, NMC-478, NMC-475, NMC-435).
- Software Engineering: ~18 open — mostly long-lived infra Todo/Backlog items
  (Datadog tracing setup SE-609/610/611/612, Vecna Azure→AWS migration SE-589,
  Crowdsec SE-587, ELB cleanup SE-405, on-call rotation SE-665, weekly `/insights`
  reminder SE-641).
- Mobile club: ~2-3 open (MC-4179 Triage customer ticket, MC-4000 risk-app CI/CD
  In Progress, MC-4112 Datadog E2E Todo).

**This confirms Linear already holds a real, non-trivial personal workload for
Gaetan** — comparable in volume to what a manual "to check/to test" list would
hold, but scattered across 3 teams' boards with no single saved view surfaced by
default (would need a personal filtered view: assignee=me, all teams, non-
terminal states — doable today with existing Linear features, see §4).

## 4. What Linear gives for free vs. the friction, for hosting personal + Tars items

### Free (native Linear model, zero new infra)
- **States/workflow per team** — already covers Backlog/Todo/In Progress/In
  Review/Done/Blocked; a to-check/to-test list maps directly onto this.
- **Assignee** — issues can be assigned to Gaetan regardless of who created them
  (human, Tars, or Linear's own agent framework).
- **Labels** — cheap, filterable, no schema migration; `origin:tars`,
  `origin:personal`, `to-check`, `to-test` labels would let one saved view
  (assignee=me, any label) become "the one kanban" without inventing a new team.
- **Due dates, priority, cycles** — already used elsewhere in the org (MC-4112
  has priority; cycles exist on other teams) — free scheduling primitives.
- **Personal saved views / filters** — Linear supports per-user saved views
  across teams (e.g. "My issues" already exists as a built-in cross-team view);
  a custom view (assignee=me, state≠Done/Canceled) is a **no-code** way to get
  "one kanban merging personal + Tars + Linear-native items" for anything that
  is a Linear issue.
- **API + webhooks + GraphQL** — mature, well-documented; Tars' engagement-checker
  skill already reads Linear via **direct GraphQL** (not just the MCP tools used
  here), so the credential/auth path for Tars-on-the-VM to read/write Linear
  is a solved, exercised problem already — no new integration surface needed
  for Tars to create/update issues.
- **Agents framework** — Linear has a first-class "Agents" team/session concept
  (`list_agent_skills`, delegate field on issues) already used in this workspace
  (NMC-461 "Decide the Linear service identity and token grant", NMC-465 "Own
  the Linear app-token manager") — Tars could plausibly act as a Linear agent
  session rather than a bare API bot, reusing infra already built for the
  Support Engineer agent.

### Friction
- **No personal/private space.** Everything lives in the Mobile Club company
  workspace. A personal to-check/to-test list or Tars-created noise (e.g. "check
  if X regressed") would be visible org-wide unless put in a team with
  restricted membership — and creating a new team for one person is exactly the
  kind of infra Gaetan's rules (KISS/YAGNI) push against. A label + personal
  saved view avoids a new team entirely.
- **Seats/licensing for a bot identity.** Tars creating/editing issues under its
  own identity (vs. impersonating Gaetan) needs a Linear seat or the existing
  "Agent" app-user mechanism (NMC-461/465 already solved this for the Support
  Engineer agent — reusable pattern, not new spend, if that seat/agent identity
  can be shared).
- **API auth for Tars on the VM.** Already solved: engagement-checker skill
  reads Linear via direct GraphQL today. Extending to writes (creating issues
  from Tars, moving states) is incremental, not a new integration.
- **Mixing signal with company noise.** Gaetan's own assigned-issue stream is
  already ~40 open items of real company work (infra, Help Tech tickets,
  incidents) mixed with whatever personal/Tars items would be added — a single
  "assignee=me" view becomes a company-priorities-plus-personal-todos blend
  unless labels/filters segment it. This is a filtering problem, not a Linear
  limitation — solvable with the label approach above, worth flagging as a
  real usability cost of merging into Linear vs. a separate personal board.
- **No org-wide separation of "Tars-authored" from "human-authored" without
  discipline.** Nothing stops issue-title bloat if Tars creates issues liberally;
  would want a `source:tars` label + a lightweight convention, not new tooling.

## 5. Bottom line for the merge-options comparison

Linear is a strong **candidate host** for the unified kanban:
- Kanban columns: already there, per-team, customizable.
- Merge personal + Tars + Linear-native: **no new infra required** — a label
  taxonomy (`origin:personal`, `origin:tars`, `origin:linear`) plus one saved
  cross-team view (assignee=me, filtered by state) does it.
- API path for Tars: **already built and exercised** (engagement-checker's
  direct GraphQL reads); writes are incremental from there.
- Real gap: no personal/private space — anything added is company-visible.
  Given Gaetan is workspace admin and the items in question (to-check/to-test,
  Tars-created improvement ideas) are not sensitive in nature, this is likely an
  acceptable tradeoff, but it is the one point worth deciding explicitly rather
  than defaulting into.
- Alternative to a label taxonomy: a single new lightweight project (not team)
  scoped to Gaetan + Tars, inside an existing team (e.g. "Agents" or "Software
  Engineering") — cheaper than a new team, still filterable, still uses native
  states/assignee/labels.
