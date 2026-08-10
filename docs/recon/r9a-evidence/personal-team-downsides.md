# Personal team downsides — Linear Business plan (primary sources only)

Researched 2026-08-10. All citations are linear.app/docs or developers.linear.app
pages, fetched live via WebFetch this session. Quotes are WebFetch's extraction
(no raw-HTML cache), treat as close paraphrase/verbatim-where-quoted, not
byte-exact HTML diff. Builds on prior evidence files `linear-product-sync.md`
and `linear-live.md` in this same scratchpad dir — not re-derived here.

## 1. Private team semantics on Business plan

Source: https://linear.app/docs/private-teams

- **Plan gate**: "Available to workspaces on our Business and Enterprise
  plans." Mobile Club is confirmed reachable for this (Business or above,
  since private teams exist in the live workspace per `linear-live.md`
  precedent pattern — plan tier itself still not directly confirmed by any
  tool, but the feature-gate text is primary-sourced).
- **Who sees the team exists**: on Enterprise, workspace **owners** can view
  private teams under Settings > Administration > Teams. On other paid plans
  (i.e. **Business**), "workspace admins can see all private teams in
  Settings > Administration > Teams." So on Business specifically: admins
  see it exists; non-admin members do not.
- **Who sees its issues**: "Those who are not a member of the private team
  will not be able to see issues associated with the team" — flat statement,
  no admin exception carved out for issue *content*, only for team
  *existence/joining* (next bullet).
- **Can admins see/join without invitation?** Yes, and the docs give it
  precisely: on Business, "workspace admins can see all private teams... or
  join a private team by adding themselves as a member" — self-service, not
  invitation-gated. But: "If an admin or owner attempts to join a private
  team, they will receive a pop-up warning before confirming." So an admin
  CAN self-add without asking Gaetan, and only gets a confirmation dialog,
  not a block. Once joined, they become a member and by the rule above can
  then see issues.
- **My Issues / cross-team custom views**: **UNVERIFIED.** Neither
  https://linear.app/docs/private-teams nor https://linear.app/docs/my-issues
  mentions private teams at all. `my-issues` only says "My issues is a
  curated view that shows your most pertinent issues" (assigned/created/
  subscribed/recent-activity tabs) with no inclusion/exclusion language tied
  to team privacy. Do not assume either way from docs; the only sound
  inference is structural — My Issues and an "All teams" custom view are
  scoped by the viewing user's own membership/assignment, and Gaetan would be
  the private team's only member and its issues' likely assignee, so nothing
  in the private-team access rule (member-only visibility) would exclude the
  *owning member himself* from seeing his own team's issues in his own
  curated views. That is inference from the access-control text, not a
  documented guarantee.
- **Sub-teams**: "Private parent teams can only have private sub-teams,"
  with two sub-flavors when creating under a private parent: "Restricted"
  (parent members can see and join) or "Private" (only explicitly added
  members). Not relevant to a single-member team but confirms no
  public-sub-under-private escape hatch.

**Verdict**: Admins (7 at Mobile Club) can see the team exists and can
self-join to see its issues, with only a warning dialog as friction — not a
hard barrier. "Private" does not mean "admin-proof."

## 2. Moving an issue between teams

Source: https://linear.app/docs/editing-issues

- **Identifier**: "When you move an issue to a new team, we generate a new
  issue ID and unique URL for the issue." Confirmed: PER-12 moved to NMC
  becomes a new NMC-xxx identifier, it is not preserved.
- **Old URL/ID**: "Old URLs will still work and redirect to the new issue
  URL. Searching for old issue IDs will also bring up the current issue."
  So links and bookmarks survive via redirect; the old ID becomes a pointer,
  not dead.
- **What the docs say breaks/changes on move**:
  - Team-scoped **labels**: "Removed"
  - **Projects**: "Removed"
  - **Cycle**: "May be cleared"
  - **Status**: "Changed" (remapped into the destination team's own workflow)
- **Git branch names, external references**: **UNVERIFIED** — not mentioned
  on this page. Since Linear's suggested git branch name is derived from the
  issue identifier at the time of creation (per general Linear convention),
  and the identifier changes on move, a previously-created branch name would
  keep the OLD identifier string (branches aren't rewritten retroactively) —
  but this is inference, not something the docs state explicitly for the
  move case. Flag as unconfirmed, do not build assuming either behavior.
- **Undo**: "You can undo the move with Cmd/Ctrl Z. Most fields are
  restored, but changes to labels, subscribers, estimates, or access-related
  assignments may remain" — i.e. undo is not perfectly clean even
  immediately after.

**Verdict**: identifier always changes, old links redirect (no dead links),
but labels/projects/cycle are actively stripped on move — moving an issue
into or out of a personal team is lossy, not a free relabeling.

## 3. OAuth-app agent user vs personal API key — billed seat?

Sources: https://linear.app/docs/agents-in-linear ,
https://linear.app/docs/billing-and-plans ,
https://linear.app/docs/members-roles

- **Agent/app user (the `@oauthapp.linear.app` pattern)**: "Agents are not
  counted as billable seats in Linear." Explicit, direct statement on the
  agents-in-linear page. The page adds: "The services that provide the agent
  may have their own pricing structure" — i.e. Linear itself doesn't charge
  a seat for the app-user identity, but whatever service hosts the agent
  (e.g. an OAuth app's own SaaS pricing) is out of Linear's scope.
- **General seat-billing rule**: billing-and-plans states "Customers are
  billed for the number of unsuspended users within a workspace" / "charged
  ... for the number of unsuspended users in any role on your workspace" —
  role-agnostic (member/admin/owner all count). Guests are the one
  explicitly-named exception that still counts: members-roles says "Guest
  accounts are only available on Business and Enterprise plans, and are
  billed as regular members" — i.e. guests count too, contrary to a common
  assumption; agents are the only named carve-out.
- **Personal API key**: **UNVERIFIED as an explicit statement**, but
  structurally not a new user — https://linear.app/docs/api-and-webhooks
  says only "Admins and permitted Members can create personal API keys from
  Settings > Account > Security & Access," i.e. the key is issued to an
  *existing* member's account, not a new identity. No page states "personal
  API keys do not affect billing" in so many words, but since billing counts
  *users*, not *keys*, and a personal API key does not create a user, there
  is no seat-billing difference between using Linear via UI vs via a
  personal API key tied to the same account. Treat the "no incremental seat
  cost" conclusion as a strong structural inference, not a directly quoted
  guarantee.

**Verdict**: OAuth-app agent user — confirmed zero seat cost, explicit docs
quote. Personal API key — zero incremental seat cost by construction (reuses
an existing paid seat), but that specific sentence doesn't exist in the docs;
mark the "no incremental cost" claim UNVERIFIED-BY-EXPLICIT-QUOTE though
structurally sound.

## 4. Linear's own guidance on team granularity

Source: https://linear.app/docs/teams , https://linear.app/docs/sub-teams,
https://linear.app/docs/how-to-use-linear-small-teams

- **Direct granularity guidance found**: "If you are unsure how to structure
  your teams, start with one or two. It is easy to add more teams in the
  future." And: "Keeping everyone on one team is the simplest approach (best
  for small teams)."
- **Org-chart guidance**: the small-teams page frames team creation around
  function, not individuals: "Mimic your org chart and establish teams by
  function — such as Engineering, Design, and Marketing — and use Sub-teams
  to reflect specialized teams under those." This implicitly argues for
  functional/departmental teams over per-person teams, but it is a
  recommendation-by-omission, not a stated prohibition.
- **No explicit "don't create many small/single-person teams" warning found
  anywhere** in linear.app/docs across the three pages fetched. **UNVERIFIED
  as a named anti-pattern** — Linear's docs bias toward fewer/bigger teams by
  suggesting the default and the org-chart framing, but never say "avoid
  single-person teams" or name a concrete overhead cost from having many.
  The overhead argument (triage/cycles/workflows-to-maintain/notifications)
  is not one Linear's docs make explicitly anywhere found; it would need to
  be constructed by the reader from the per-team feature list in §5, not
  quoted from a "don't do this" passage.

**Verdict**: docs nudge toward fewer/functional teams ("start with one or
two", org-chart-based) but never explicitly warn against single-person
teams or name a cost. The overhead case has to be inferred, not cited.

## 5. Per-team overhead a new team drags in by default

Source: https://linear.app/docs/teams , https://linear.app/docs/use-cycles ,
https://linear.app/docs/triage , https://linear.app/docs/configuring-workflows ,
https://linear.app/docs/labels (already fetched in prior research, see
`linear-product-sync.md` §3)

- **On creation, a new team automatically gets**: a team page, an Issues
  section, a Projects section, and a Views section (linear.app/docs/teams).
  These are structural, not "maintenance," and can't be turned off.
- **Cycles**: NOT on by default — "opt-in and need to be enabled in team
  settings" (linear.app/docs/teams); confirmed by use-cycles: "Cycles are
  configured under Team Settings > Cycles, where you turn on the toggle for
  Enable cycles." So a personal team starts with zero cycle overhead unless
  Gaetan explicitly turns it on.
- **Triage**: also opt-in — "To enable triage, go to your Team Settings >
  Triage and toggle it on, after which Triage will appear under the team
  name in the sidebar" (linear.app/docs/triage). Off by default, no
  maintenance burden unless enabled.
- **Workflow states**: NOT opt-in — every team gets a default workflow
  automatically: "Linear creates a default workflow for you that you can
  further customize in Team Settings," default set/order "Backlog > Todo >
  In Progress > Done > Canceled" (linear.app/docs/configuring-workflows).
  This one IS mandatory per-team state that exists whether or not you touch
  it — it just starts sane and requires zero manual setup to be usable, but
  it is still a distinct per-team object (not shared with other teams) that
  a "team explosion" would multiply, one config surface per team.
- **Labels**: team-scoped labels are optional and separate from workspace
  labels — a new team has none by default; nothing forces label duplication
  (per `linear-product-sync.md` §3, workspace-level labels are usable
  workspace-wide without any per-team setup).

**Verdict**: the concrete "drag-in by default" list is short — a page/
issues/projects/views shell plus one auto-generated default workflow.
Cycles, triage, and custom labels are all opt-in, not automatic overhead.
Docs do not name notifications-per-team as an overhead item anywhere found.
