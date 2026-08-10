# Linear product facts + sync prior art (primary sources)

Researched 2026-08-10 for Gaetan's pushback on r9-kanban-unification.md. All
facts below are from linear.app official docs / developers.linear.app /
official Linear GitHub or changelog, fetched live. Quotes are paraphrased
from WebFetch extraction of the live pages (tool has no raw-HTML cache; treat
quoted fragments as close paraphrase, not byte-exact).

## 1. Private teams

Source: https://linear.app/docs/private-teams , https://linear.app/pricing ,
https://linear.app/docs/teams

- **Plan gate**: Private teams require **Business or Enterprise** plan. Not
  on Free/Basic. → Mobile Club's plan tier must be checked (not verified here
  — check Settings > Administration > Billing) before this path is even
  available to Gaetan inside the company workspace.
- **Visibility**: non-members of a private team **cannot see its issues** —
  not in search, not in cross-team views, nothing. Projects created under a
  private team are visible only to team members; if later shared with a
  public team, the project becomes visible to that team too.
- **API access**: a personal API key inherits the key owner's access — if
  the user has private-team access, the key can read it. No separate
  API-level bypass.
- **Admin/owner visibility (non-Enterprise)**: workspace admins can see the
  list of private teams (Settings > Administration > Teams) and can join
  (with a warning), after which they see its issues. On **Enterprise**,
  owners get this too, plus the ability to share individual issues from a
  private team with non-members.
- **Who can create a team**: by default **any workspace member** can create
  a new team (private or not) via Settings > Your teams > "+". Workspace
  admins can restrict team creation to admins only (Settings >
  Administration > Security). **Changing an existing team's visibility**
  (public↔private) is restricted to workspace owners, admins, and team
  owners.
- **Sub-teams**: a private parent team can only have private sub-teams.

**Implication for Gaetan's option 2** (personal team inside Mobile Club's
single workspace): if Business/Enterprise is active, Gaetan (a member) can
almost certainly self-serve create a private "Gaetan" team, and it would be
invisible to other Mobile Club members by default — closest thing to a
"personal Linear space" without a second workspace. Needs a one-line plan
check first.

## 2. Cross-team views / boards

Source: https://linear.app/docs/custom-views , https://linear.app/docs/teams,
WebSearch of Linear docs (user-views, workflows pages 404'd directly —
content below is corroborated via docs excerpts surfaced in search, treat
board-drag mechanics as **not fully confirmed from a single primary page**).

- **A single custom view CAN span multiple teams.** Docs: "To share a view
  across multiple teams but limit it to projects or issues within a specific
  team or teams, create an *All teams* view and use filters to refine the
  list." So: yes, one view, filter `assignee = me` (or a label), across all
  teams Gaetan is a member of.
- **Workflows are defined per team** — each team has its own ordered list of
  statuses ("Backlog, Todo, In Progress, Done" can differ per team, e.g.
  design team vs eng team). Docs: "Workflows can be customized per team...
  each team independently configures issue statuses."
- **Status *categories* are the fixed umbrella every custom status belongs
  to**: Triage (if enabled), Backlog, Unstarted, Started, Completed,
  Canceled. This is Linear's underlying conceptual model
  (linear.app/docs/conceptual-model) — every team-specific status is tagged
  with exactly one category.
- **Board grouping across teams**: not found spelled out on one exhaustive
  primary page in this session's fetch, but the mechanism implied by
  "workflows per team" + "status category" is the standard Linear behavior:
  an "All teams" board groups by **status category** (5-6 fixed columns),
  not by each team's raw per-team status label — because raw statuses differ
  per team and can't be column-merged 1:1. **This needs a 2-minute visual
  confirmation in Gaetan's own workspace** before designing around it — flag
  as "corroborated but not certain," not a hard primary-source citation.
- **Dragging a card across category-columns in a cross-team board**: this
  detail (does it write back the underlying per-team status, and which
  status if the team has several statuses in that category) is **not
  documented on any page this session reached**. Recommend: verify live by
  dragging one real issue, don't build the design assuming a specific
  behavior.

## 3. Workspace-level labels

Source: https://linear.app/docs/labels

- Labels can be created **at workspace level or at team level**. Workspace
  labels are visible/usable from every team.
- Team-level labels are scoped to that team, BUT: "team-specific labels act
  like workspace labels when filtering all-teams / multi-team views" as
  long as the label **name matches** across teams — i.e., if three teams
  each independently have a label literally named "urgent", an all-teams
  filter on "urgent" unions them, even though they're not the same label
  object.
- **API exception**: this name-based union is a UI/filter convenience only.
  API calls need the actual per-team label ID — same-named labels across
  teams are different objects with different IDs.
- Sub-teams inherit parent-team labels read-only (can't edit at sub-team
  level).

**Implication**: a workspace label "personal-board" applied consistently
(or a label named identically per team) is a legitimate, sound mechanism to
pull a cross-team subset into one filtered view — cheaper than any relation
or sync mechanism.

## 4. Write access for an agent

Source: https://developers.linear.app/docs/ (agents, oauth-actor-authorization,
webhooks pages), https://linear.app/docs/mcp

Three distinct identity mechanisms, ranked by fit for a personal bot like Tars:

1. **Personal API key** (Settings > Account > Security & Access). Simplest.
   Scopable to Read/Write/Admin/Create-issues/Create-comments, and to
   specific teams. Acts AS the creating user (issues show as created by
   Gaetan, not as a bot). **Best fit for a single-person, single-workspace,
   non-multi-tenant tool like Tars** — no OAuth dance, no app review, no
   webhook infra required if you poll instead.
2. **OAuth application (`actor=app`)** — creates a dedicated bot user inside
   the workspace representing the app, not impersonating a person. Needs
   workspace **admin** approval to install/authorize the app (the
   `actor=app` param on the OAuth URL requires admin rights). Scopes
   include `app:assignable`, `app:mentionable`. Heavier to set up than an
   API key; buys a distinct "Tars" identity in Linear's activity log/mentions.
3. **Linear "Agents"** (developers.linear.app/agents) — a purpose-built
   framework on top of OAuth `actor=app`: the agent can be @mentioned,
   delegated issues (shows as "delegate" not classic assignee), comment,
   and participate in "agent sessions" with built-in session-state
   lifecycle. Free to develop; agent users don't count as billable seats.
   This is Linear's recommended shape for a bot that should behave like a
   teammate (get pinged, work an issue, reply) — heavier than needed for a
   passive kanban mirror, but the natural fit **if** Tars is ever meant to
   act *inside* Linear conversations (comment, get @mentioned) rather than
   just read/write issue rows.

**Webhook delivery**: Linear webhooks require a **publicly-accessible
HTTPS, non-localhost URL**; 5-second response timeout, non-200 or slow
response triggers retries at +1min/+1h/+6h, then the webhook is auto-disabled
until manually re-enabled. **The Tars VM (192.168.0.9) is LAN/tailnet-only —
no public HTTPS endpoint exists today**, so live webhook push is not viable
without standing up a public relay (e.g. a small reverse tunnel or a cloud
relay function) purely to catch webhooks — extra infra Gaetan's own
constraint (avoid infra for what a script covers) argues against for a
personal single-user tool. **Polling** (Tars calls Linear's GraphQL API on
an interval, as engagement-checker already does) is the realistic,
zero-extra-infra path and matches what's already proven working.

## 5. Sync prior art — field ownership, conflict rules, known failure modes

Sources: https://linear.app/docs/mcp, https://unito.io/connectors/linear/,
https://unito.io/blog/bidirectional-sync/, GitHub (calcom/synclinear.com,
jtormey/linear-sync), https://linear.app/docs/github-integration (fetch
returned only a stub header — Linear's own GitHub-sync mechanics could not be
confirmed from primary source in this session; only the community/Unito
material below is solid).

- **Official Linear MCP server** (`https://mcp.linear.app/mcp`, hosted,
  OAuth in-browser, no local install): read-write by default, or hit
  `/mcp/readonly` for a hard read-only mode, or scope an OAuth/API key to
  `read` only. Tools: find/create/update issues, projects, comments,
  relations. **It is call-response, not event-driven** — no push/sync
  primitive of its own; a sync tool built on it still needs its own polling
  or webhook loop layered on top. No cross-workspace session — each
  workspace needs separate auth.
- **Unito (commercial no-code sync)** — the general pattern documented for
  2-way sync: per-field, per-direction mapping (you choose 1-way or 2-way
  per field, not per issue) — e.g. status can be 2-way while attachments are
  forced 1-way because the target API can't accept them back cleanly.
  Unito's own bidirectional-sync explainer names concrete failure classes:
  - **Sync loops**: A updates → sync to B → B's update event syncs back to A
    → loop, unless the tool tracks change-origin to suppress echo-writes.
    This is the single most commonly cited two-way-sync bug class.
  - **Conflict resolution**: both sides edited before the next sync tick —
    documented mitigations are last-write-wins (silently drops the other
    edit), manual-review queue (kills the "autonomous" property), or
    field-level merge (adds real complexity, only works for fields that are
    safely mergeable, e.g. free text is not).
  - **Schema mismatch**: differing status vocabularies (Linear's per-team
    custom statuses vs a different tool's fixed set) need a mapping table
    that breaks silently whenever either side adds/renames/removes a status.
  - **API rate limits**: aggressive 2-way polling to feel "live" competes
    with the rate budget, forcing a real-time-vs-quota tradeoff.
- **One-way mirror is the documented safe pattern**: source of truth stays
  in exactly one system; the mirror is disposable/rebuildable; no conflict
  resolution needed because only one writer exists per field. The
  established compromise for "I still want to comment from the other side"
  is **one-way state mirror + comment write-back only** (comments are
  additive/append-only, so writing a comment on the mirrored side and
  relaying it back cannot conflict with anything — there's no "current
  value" to fight over, unlike a status or title field).
- **GitHub↔Linear specific prior art**: Linear ships a **native GitHub
  integration** offering both 1-way and 2-way issue sync (per Linear's own
  marketplace/integration copy — the docs page itself didn't load in this
  session, so cite with caveat). Community project **calcom/synclinear.com**
  (open source) deliberately narrows scope: it syncs only issues that carry
  a **specific label**, one-way-triggered by adding that label, explicitly
  to avoid giving external GitHub contributors access to the internal Linear
  team — i.e. even Linear's own ecosystem defaults to label-gated, scoped
  sync rather than blanket two-way, echoing the "smallest sync surface that
  works" lesson.

**Bottom line for Tars' design**: given Tars can only poll (webhook is
blocked by no public endpoint) and Gaetan wants KISS/single-writer-per-surface,
the safe pattern per the prior art above is: **pick ONE system as the writer
for issue state** (title/status/assignee), let the other be a read
projection built by polling, and if two-way "feels" needed, restrict it to
comments only (additive, no conflict surface) — never mirror status
bidirectionally without loop-suppression and a conflict rule, both of which
are exactly the maintenance burden Gaetan is trying to avoid re-creating
(note: the rolled-back `config.yaml.bak-linear-notion-cal` /
`.env.bak-linear-notion-cal` pair is circumstantial evidence something in
that direction was tried and abandoned on 2026-08-07 — cause unconfirmed,
but consistent with hitting exactly this class of problem).

## 6. Sub-issues / issue relations across teams

Source: https://linear.app/docs/parent-and-sub-issues,
https://linear.app/docs/issue-relations

- **Sub-issues do NOT cross teams.** Docs: "Sub-issues inherit the parent
  issue's team, priority, and project." A sub-issue is always in the same
  team as its parent — so a personal-team issue cannot have a sub-issue that
  actually lives in another company team. (Linear's own docs suggest using
  a **project** instead when work needs to span teams.)
- **Issue relations** (`related`, `blocks`/`blocked by`, `duplicate`) ARE
  available via UI shortcuts (`M`+`R`, `M`+`B`, `M`+`X`) and the GraphQL API
  (`relatedIssues`, relation types: blocks/blocked_by/related/duplicate/
  duplicate_of). Docs do **not explicitly state a same-team restriction**,
  and relations are commonly used cross-team in practice (referencing an
  issue by identifier works across teams) — but whether a related/blocking
  issue from another team **surfaces automatically inside a filtered or
  custom view of the personal team** was **not confirmed by any primary
  page reached this session**. The issue's *sidebar* on its own detail page
  will show the relation (that's documented), but that is not the same as
  the related other-team issue appearing as a *card* on the personal-team
  board — no evidence that relations feed board/view membership; view
  membership is governed by team/filter membership (§2), not by relation
  edges. **Do not assume relations = board inclusion — treat as unconfirmed
  and design as if it does NOT surface the other-team ticket automatically.**

## Answer to Gaetan's three questions, given the above

1. **Manual card creation + visual sync**: Hermes kanban has no comparable
   "drag a card" UI — it's tool-driven (12 kanban_* tools) or nothing. If
   the visual, editable board is the requirement, **Linear's own board IS
   that board** — a second bespoke visual UI to keep in sync with it is the
   exact 2-way-sync trap in §5. Simplest: make Linear the visual board,
   Hermes kanban either dropped or kept as Tars' internal scratch state
   only (not shown to Gaetan, not synced).

2. **Personal Linear team inside Mobile Club workspace**: feasible if
   plan ≥ Business (verify), member-creatable, private by default → nobody
   else at Mobile Club sees it. This satisfies "personal space" without a
   second workspace/account. Combine with §2's all-teams custom view.

3. **Other-team tickets visible on the personal board**: don't try to copy
   or relation-link them into the personal team (sub-issues can't cross
   teams; relations don't confirm-surface on boards). Instead use §2's
   built-in mechanism: **one "All teams" custom view, filtered by
   `assignee = me`**, spanning the personal team + the 3 company teams
   Gaetan is already a member of. That view already groups by status
   category and needs zero sync code — it's Linear's native cross-team
   read model. This is the soundest FEASIBLE design: no new integration, no
   webhook infra, no conflict logic, reuses the exact mechanism Linear ships
   for this. Tars' role reduces to: poll Linear GraphQL (as
   engagement-checker already does) to read/report on that same
   assignee=me filter, comment-write-back only if any bot writes are needed,
   and never attempt to own Linear's status field.

## Open items / not confirmed from primary sources this session

- Whether Mobile Club's Linear plan is Business/Enterprise (private teams
  gate) — check Settings > Administration > Billing.
- Exact behavior of dragging a card between status-category columns on an
  "All teams" board (which per-team status it writes) — verify live.
- Whether an issue relation (blocks/related) causes the related issue to
  appear as a card in the other issue's team board/view — not found
  documented; assume no.
- Root cause of the 2026-08-07 Linear+Notion+Calendar integration rollback —
  circumstantial only, not investigated here (would need git blame /
  Tars' own status/lane-a.md log, out of scope for this product-facts task).
