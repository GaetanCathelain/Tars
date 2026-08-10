# R9a — Addendum: personal Linear team, cross-team visibility, and whether to sync

Addendum to `docs/recon/r9-kanban-unification.md` (2026-08-10). Read r9 first — this file does not
repeat its option catalogue, prior art or problem statement; it answers Gaetan's three follow-up
questions and revises r9 where new evidence forces it. New evidence, cited inline, all under
`docs/recon/r9a-evidence/`: `[live]` `linear-live.md` (Linear MCP, read-only) · `[vm2]`
`vm-sync-facts.md` (read-only ssh probes of 192.168.0.9) · `[prod]` `linear-product-sync.md` (Linear
docs + sync prior art). r9's bundle is still `[vm]`, `[lin]`, `[ext]`, `[repo]`, `[obs]`.

## 1. TL;DR

**Half of Gaetan's proposal is right and cheap; the other half is the part to delete.** A personal
Linear team is feasible — he is one of 7 workspace admins `[live]` §2, and any member can create a
team, private only on Business/Enterprise (*plan tier unverified* — one billing-page check) `[prod]`
§1. The Linear↔kanban.db sync is the part to drop: Hermes kanban has **zero sync primitives** (no
import/export/sync subcommand, no external-id field; `link`/`unlink` are local parent-child edges)
`[vm2]` §2, webhooks are impossible on a tailnet-only VM `[prod]` §4, and every two-way sync failure
class in the prior art applies `[prod]` §5. **Soundest design: Linear itself is the one board** —
**one native "All teams" custom view, `assignee = me`**, spanning MC + NMC + SE `[prod]` §2 `[live]`
§3, plus a **workspace-level label** for personal/Tars items (§4.v). A dedicated personal *team* is
the conditional upgrade, not the default: only if plan ≥ Business **and** privacy is non-negotiable
(§4.i). `kanban.db` is demoted to Tars-internal execution state, unused in v1.

*Caveat carried through this file:* that Linear boards group columns by the fixed status **category**
enum is **corroborated but not certain** — `[prod]` §2 refuses to state it as primary-source fact and
asks for a 2-minute visual check in Gaetan's workspace. Every claim below that the MC-QA/NMC-QA
divergence "collapses by construction" depends on it.

**This overturns r9's recommendation (option A) and upgrades its option B.** It does *not* overturn
r9's "Linear referenced, never synced" stance — the new evidence reinforces that stance and simply
moves the board to the side of the line where the data already is.

## 2. "How do I manually create a card?" — what actually works today

| Path | Works today? | Evidence |
|---|---|---|
| Slack DM to Tars: "create a card for X" → LLM calls `kanban_create` | **Yes.** `toolsets: [hermes-cli, kanban]` confirmed live; single implicit profile serves chat and crons alike | `[vm2]` §1, §3 |
| `/kanban …` slash command in Slack | **Probably not** — blocked platform-wide for Agent-class apps *per the operating-tars skill / prior recon, **not probed this session***. Conflicts with r9 §3.2 / `[vm]` §2 (`notify-subscribe` "backs `/kanban subscribe`") and `[ext]` §0; unresolved, one probe settles it | `[vm2]` §3 vs r9 §3.2 |
| `:kanban:` emoji reaction → card | **No.** `slack.reaction_triggers` is unset; config-only change + workspace emoji + one SOUL line (P2) would enable it | `[vm2]` §3, `[repo]` via r9 §3.1 |
| `ssh gaetan@192.168.0.9` → `hermes kanban create …` | **Yes**, ~30 subcommands available; terminal-only, not a board | `[vm]` r9 §3.2, `[vm2]` §2 |
| Drag a card in a visual UI | **No such UI on this install.** A GUI is claimed only by secondary docs — *unverified* | r9 §4A `[ext]` |

Two things make "manual card creation" worse than it sounds, both **now confirmed** (r9-unverified):
`dispatch_in_gateway: true` with a 60 s tick, and `auto_decompose: true` `[vm2]` §1 — so creating a
card *is* starting the work, r9 §5.1's blocking issue, now measured. **Still unverified, and r9 §5.1
said to measure it:** whether an *unassigned* card with no `orchestrator_profile` is skipped by the
dispatcher. `[vm2]` §1 confirms `orchestrator_profile: ''` / `default_assignee: ''` as live defaults
but never dispatched a test card, and the WF5 smoke cards' assignee state is unknown. This file
assumes the worst case (every card dispatches) deliberately.

Also resolved: r9 §5.1's second unknown ("which profile do the cron skills run under") — **there is
no profile system on this install** (no `profiles:` stanza, no `~/.hermes/profiles/`), so both crons
run under the single default profile and *do* have kanban access `[vm2]` §3, §5. **This overturns r9
§3.2**: the two skills *are* under `~/.hermes/skills/`, in the `orchestration/` subdirectory
(`…/skills/orchestration/{daily-work-brief,engagement-checker}/SKILL.md`) `[vm2]` §5 — `[vm]` §4's
57-skill listing was top-level-only. That wrong absence, not the profile question, is what made r9
call the render path "in question".

## 3. Gaetan's proposal analysed: personal Linear team + two-way sync to `kanban.db`

**What it would take, concretely, on this infra:**

- **Transport: polling only.** Webhooks need a public HTTPS endpoint `[prod]` §4 and the VM is
  LAN/tailnet-only — a relay is new infra for a single-user tool. Polling is what engagement-checker
  already does `[vm2]` §4; a new cron or an extension of it would work, both have kanban tools §5.
- **ID mapping has no home.** `hermes kanban create` has **no external-id/URL field**; a Linear key
  fits only in free-text `title`/`--body` or `--idempotency-key`, a bare dedup string nothing reads
  back `[vm2]` §2. The bridge must own a mapping table — a third durable state surface next to
  `kanban.db` and `engagement-checker.json`. engagement-checker's `linear:<key>` convention (r9 §2)
  is evidence a *separate* store is what happens in practice, not evidence of integration.
- **Status vocabulary is the sharp edge.** 11/13/9 workflow states; **NMC's "QA" is `started` while
  MC's "QA" is `completed`** `[live]` §3, so any name-keyed mapping files an in-flight NMC card as
  done. Keying on `statusType` fixes that and then breaks on every state added or renamed — Unito's
  schema-drift class, alongside sync loops without change-origin tracking, conflict resolution
  (last-write-wins drops data, manual review kills autonomy) and rate-limit pressure `[prod]` §5.
- **This-infra killer (worst case, assumed deliberately):** a mirrored issue landing as `todo` is
  promoted to `ready` and — *if the dispatcher does not skip unassigned cards, which is **still
  unverified**: r9 §5.1 flagged it, `[vm2]` did not test it* — executed within 60 s `[vm2]` §1.
  Mirroring ~40 company issues would attempt to *do* ~40 company tickets. Preventing it means
  `dispatch_in_gateway: false` or a parked status — which removes the dispatcher/claims/event-log
  that were `kanban.db`'s only advantage over a text file.
- **Effort:** M–L and permanently owned; zero of it configuration `[vm2]` §2.

**What the 2026-08-07 `linear-notion-cal` rollback tells us — best-supported reconstruction, not a
root cause.** A **Google Calendar API 403 "API not enabled for this project"** live-check failed at
21:29–21:35 that evening (session `20260807_205326_2f2ee68a`); current `mcp_servers` retains only
`slack` and `notion`; no Linear-specific error near that window; `status/lane-a.md:30` has Linear +
Notion + Calendar all CONFIRMED working earlier the same day `[vm2]` §4. But `[vm2]` §4 closes: "**No
corroborating rollback rationale beyond the above was found** … the causal chain above is inferred
from timestamps + the 403 error + the config diff, not from a stated decision log entry", and hedges
that Linear was "very likely pulled together with Calendar". **No rollback action is logged at all.**
The alternative the evidence does not exclude: **the Linear MCP server was dropped deliberately, as
redundant once engagement-checker's direct GraphQL covered the read path.** Both readings support the
same two lessons: (a) bundling three integrations made one fixable GCP setting cost all three — do
them one at a time; (b) the Linear path that *survived* is the narrow one, direct GraphQL from a
skill, not the MCP server. Prefer it.

**Verdict on the proposal:** the personal-*surface* half is sound (though a label beats a team unless
privacy is bought — §4.v vs §4.i); the sync half buys nothing Gaetan cannot get natively and costs
the whole documented failure catalogue. Do not build it.

## 4. Alternatives, judged on the cross-team requirement

**First, calibrate the requirement.** Gaetan already has cross-team visibility today, for free:
Linear's **built-in "My issues"** surfaces the ~40 issues across MC/NMC/SE (r9 §2 / §4B, `[lin]` §4:
"the board artefact is a cross-team assignee=me view — which already exists for free as Linear's
built-in 'My issues' view"). So "cross-team" does not discriminate between Linear-hosted options. The
**delta a custom view adds** is exactly two things: it also covers whatever surface holds
personal/Tars items, and it excludes terminal statuses. That delta is all that is worth paying for.

### (v) Workspace label + one "All teams" view — no new team  ← **the cheap default**

Personal/Tars items become issues in an existing team carrying a **workspace-level label**
(`origin:personal`/`origin:tars`); the board is one saved "All teams" view, `assignee = me`,
non-terminal, optionally label-filtered. 15 of 28 labels are already workspace-scoped and cross-team
usable `[live]` §4; `[prod]` §3 calls a workspace label "a legitimate, sound mechanism … cheaper than
any relation or sync mechanism"; `[lin]` §4-Friction, verbatim: "creating a new team for one person
is exactly the kind of infra Gaetan's rules (KISS/YAGNI) push against. A label + personal saved view
avoids a new team entirely" (`[lin]` §5's variant: one lightweight *project*, not team).

- **Cross-team: identical to (i)** — same view mechanic, same columns.
- **KISS: wins outright.** No plan gate, no admin action, no 9th team, no new membership surface.
- Single-writer: identical to (i) — Linear owns status, Tars creates and comments.
- Privacy: **none** — items are org-visible. That is the *whole* difference against (i); per `[lin]`
  §4-Friction the items (to-check/to-test, Tars ideas) are not sensitive, but decide it explicitly.
- Cost: label discipline, and personal noise in the same stream as 40 real company issues `[lin]`
  §4-Friction — a filtering problem the label itself mitigates.

### (i) Personal Linear team + the same "All teams" view  ← **conditional upgrade**

Same view mechanic as (v); personal/Tars items live in a dedicated team, private if the plan allows.
Tars writes over GraphQL — the **read** path is exercised, writes are not (§5).

- Cross-team: same as (v). Not a differentiator.
- Privacy: a private team needs Business/Enterprise (*plan unverified*) `[prod]` §1, and on
  non-Enterprise admins still see the private-team list and can join it — so on today's likely plan
  it buys **tidiness, not privacy**, and loses to (v) on KISS.
- Column grouping: **unverified** (§1 caveat). If boards do *not* group by status category, the
  9/11/13 divergent statuses do not collapse and the MC-QA(`completed`)/NMC-QA(`started`) trap
  `[live]` §3 survives here as everywhere else.
- Costs: a 9th team where **no personal/single-member team precedent exists** — all 8 map to business
  units or product surfaces `[live]` §1; an admin-gated creation; a Linear issue is heavier ceremony
  than a DM for "check if X regressed"; if the plan gate fails it is (v)'s downside without (v)'s
  cheapness.
- **Verdict: only if plan ≥ Business AND privacy is non-negotiable.** Otherwise take (v).
- *Verify live, 2 minutes*: category grouping itself, and what dragging a card between category
  columns writes back on an all-teams board `[prod]` §2.

### (ii) r9 option A unchanged — Hermes board, Linear referenced by link

- **Fails to put the company issues on the *Hermes* board** — *not* cross-team visibility, which
  Gaetan already has via built-in "My issues". r9 imports nothing from Linear (r9 §5.2 step 6), so the
  ~40 issues never appear *there*, and making them appear means the sync §3 prices. Read as "one
  surface", requirement #3 rules this out.
- Also fails "a board I can edit visually": no GUI, and slash commands blocked per the operating-tars
  skill — *not probed, and contested by r9 §3.2* (§2 above).

### (iii) Hybrid one-way — Linear is truth, cron mirrors a filtered slice into `kanban.db`

Read-only mirror for Tars's use; write-back restricted to comments (additive, no conflict surface) —
the pattern the prior art repeatedly recommends `[prod]` §5.

- Cross-team: the mirror can span teams, but concretely each ticket becomes a card whose Linear key
  can live **only** in free-text `title`/`--body` or `--idempotency-key` (a bare dedup string nothing
  reads back) `[vm2]` §2, columns mapped by `statusType`, created in a non-dispatchable status. And
  **the mirror serves Tars, not Gaetan** — he still reads the Linear view, so it adds a store without
  adding anything he sees.
- Still needs the mapping sidecar and the dispatcher parked (§3 — worst case assumed).
- **Verdict: correct shape, wrong time.** Only if/when Tars demonstrably needs local board state for
  execution (Orca run-log lane, r9 §5.2 step 4) — a v2 decision, per `kb-push-design.md`'s precedent
  of not building until the need recurs `[repo]` via r9 §3.1.

### (iv) Anything better?

Nothing beats (v) on cost; (i) beats it only under the privacy condition above. **Notion (r9 option
F)** is the answer *iff* the plan gate blocks private teams **and** privacy is non-negotiable —
private by default, native board, MCP already wired `[vm]` — but it holds none of the Linear issues,
so it sits *next to* the Linear view: two boards again. **Re-adding the Linear MCP server** is
unnecessary: direct GraphQL already works and is the path that survived the rollback `[vm2]` §4.

## 5. Recommendation

**Linear is the one board: one "All teams / assignee=me" view, no sync, `kanban.db` idle.** Host the
personal/Tars items on a **workspace label (v)** by default; promote to a **personal team (i)** only
if plan ≥ Business *and* privacy is non-negotiable.

Against the constraints: **cross-team visibility** — already free via built-in "My issues" (r9 §2,
`[lin]` §4); the saved view adds only the personal surface and the terminal-state exclusion, by
filter not code. **Single-writer per surface** — Linear owns every human-visible field, the property
two-way sync destroys `[prod]` §5. *Same disclosure r9 §5.2 step 5 made about option A:*
engagement-checker keeps a durable queue (`engagement-checker.json`, items keyed `linear:<issue-key>`,
states `open/snoozed/waiting/done/dismissed`) surfaced in Slack every 30 min `[repo]` §4, r9 §2 —
**a second Gaetan-visible status surface over the same issues**; "one board" is a UI claim, not a
storage one, and collapsing the nag queue is a v2 decision. **KISS/YAGNI** — the board exists and
holds the ~40 open company issues (`[lin]`; no count exists for the personal/to-test items — Q5); the
migration is a label, a view and a credential, not a bridge. **Tars never pushes/merges/approves** —
untouched. **What exists on the VM** — a Linear **read** path runs today `[vm2]` §4; **writes are new
and unproven**: the exercised collector is deliberately hardened, "query-allowlist only, **no
mutations**, key never leaves the process" `[repo]` §4. No write has ever run from this VM.

### Migration sketch — dispatch step 1 only; steps 3–6 are gated on the open questions

1. **Now, no gate. Gaetan, ~5 min in the Linear UI:** read the plan tier (Settings → Administration →
   Billing) `[prod]` §1; open a multi-team board and **confirm columns group by status category**
   (§1 caveat); **drag one real issue between category columns and note which per-team status it
   writes**. In parallel (Orca, read-only): **scope-check the existing `LINEAR_API_KEY`** — in SOPS
   since 2026-08-07, probed "200, viewer = Gaëtan Cathelain" (`status/lane-a.md:12`). If it is not
   read-scoped, **Tars already holds blanket write on the whole company tracker**, silently
   pre-answering Q4. Establish this before minting a second credential.
2. **Then choose the host surface:** label (§4.v) by default; personal team (§4.i) only if step 1
   shows plan ≥ Business and privacy is called non-negotiable. Either way create the "All teams"
   view (`assignee = me`, exclude completed/canceled), save and favourite it.
3. **[gated on Q3 + Q4 + step 1's scope check] Tars write identity (Orca):** simplest is a personal
   API key scoped to create-issue + create-comment on the host surface `[prod]` §4; the heavier
   distinct-identity option is the OAuth-app-user pattern this workspace already runs for Support
   Engineer (`…@oauthapp.linear.app`, team-scoped, not admin) `[live]` §5. Ingest via
   `scripts/tars-secret`, deliver per-key over ssh. **Do not re-add the Linear MCP server.**
4. **[gated on Q2] Capture path (Orca):** one skill/SOUL line — "when Gaetan asks for a card, create a
   Linear issue on the host surface, assignee Gaetan, reply with key + URL". Q2 decides whether this
   half exists at all. Optionally repoint the P2 `:kanban:` reaction at Linear `[vm2]` §3.
5. **[gated on step 1's grouping check] Render path (Orca):** `daily-work-brief` opens with the view's
   filter, grouped by `statusType` — one board, two windows. Its cron has the access `[vm2]` §5.
6. **Leave `kanban.db` alone.** No cards, no config change, dispatcher untouched. Revisit only for the
   Orca run-log lane, as Tars-internal state never shown to Gaetan (§4.iii).

*How much of step 1 is really human:* only the **billing read** and the **drag + grouping test**.
`[live]` §6 establishes only that the *MCP* exposes no `create_team` and no custom-View tooling ("has
to be verified/built in the Linear UI, not through this MCP") — **whether Linear's GraphQL API (this
addendum's preferred path) can create a team or a saved view was never checked.** Check that before
handing team/view creation back as manual; `preferences/tooling.md`: "if a tool can do it, the task
is autonomous, not a checkpoint."

### Open questions for Gaetan

1. **Plan tier, and is privacy non-negotiable?** Decides whether a personal team means private or
   merely tidy `[prod]` §1. If not private — or if org-visible personal noise is acceptable — take
   the label (§4.v) and skip the 9th team; if privacy is hard and the plan blocks it, Notion.
2. **Is a Linear issue too heavy for a half-baked "to test" note?** If yes, the answer is not a third
   board — keep those in the DM/Notion and promote to Linear only what survives a day.
3. **Tars identity:** act-as-Gaetan API key (simple, no audit separation) or an OAuth-app "Tars" user
   (distinct identity, admin install) `[prod]` §4, `[live]` §5?
4. **May Tars create issues on the 3 company teams, or only on the host surface + comments
   elsewhere?** The blast-radius decision r9 §4B flagged — possibly already answered *de facto* by
   the scope of the `LINEAR_API_KEY` in SOPS since 2026-08-07 (`status/lane-a.md:12`). Read that
   scope first (step 1); the answer scopes step 3's credential.
5. r9 §5.3 Q1 stands unchanged: **where does the "to check / to test" list physically live today?**
   Still found in neither the repo nor the vault `[repo]` `[obs]`.
6. Confirm the Calendar 403 is worth fixing separately (enable the API in the GCP console) — it is an
   orthogonal, one-setting fix, not evidence against calendar integration `[vm2]` §4.
