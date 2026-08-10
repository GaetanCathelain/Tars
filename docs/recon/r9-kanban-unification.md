# R9 — One kanban for Gaetan: merging five scattered work sources

Recon date 2026-08-10. Read-only: repo reads, live read-only probes of the Tars VM
(`ssh gaetan@192.168.0.9`), Linear MCP reads, Slack MCP read of the Tars DM, WebSearch, and a
tarball snapshot of the Obsidian vault. Nothing was written, enabled or configured.

Evidence bundle, cited inline by tag — all under `docs/recon/r9-evidence/`:
`[repo]` `repo-prior-art.md`, `[vm]` `hermes-capabilities.md`, `[lin]` `linear-landscape.md`,
`[ext]` `external-prior-art.md`, `[obs]` `obsidian-setup.md`.

## 1. TL;DR

**Use the Hermes kanban that is already installed, was enabled as of 2026-08-07 and was proven
end-to-end on the Tars VM (`hermes kanban`, `~/.hermes/kanban.db`) as the single board, and make
Slack the place Gaetan sees it** — zero new infrastructure, Tars drives it with first-party
`kanban_*` tools. Its atomic claims arbitrate *which worker executes a task*, not concurrent
human-vs-Tars field edits; that concurrency is unmeasured and low-risk for one user `[vm]`.
**"Enabled" is a 2026-08-07 spec/probe fact, not a live reading:** `hermes config get toolsets` was
never run, and `[vm]` records ten later `config.yaml.bak-*` snapshots — re-probe before acting.
**Linear stays the system of record for company work: referenced by key/link, never synced
two-way** — bidirectional sync is exactly what this repo's precedent says not to build until
recurring need is proven `[lin]` `[ext]` `[repo]`. **Settle before the first real card: the
in-gateway dispatcher auto-promoted and *executed* the WF5 smoke cards within 60 s** — today "add to
the board" means "start doing it", against Tars-delegates-never-implements; capture must land in a
non-dispatchable status or with dispatch off `[repo]` `[vm]`.

## 2. Problem

Gaetan tracks work in five places that never meet: (1) `daily-work-brief`, a stateless 08:30
narrative that re-derives everything from Slack/cooper/GitHub/Linear/email each run and uses the
Slack reporting conversation as its only history store; (2) `engagement-checker` (every 30 min
10:00–16:00 plus 17:00), which *does* keep durable state — `~/.hermes/state/engagement-checker.json`,
items keyed `slack:<ch>:<ts>` / `email:<thread>` / `linear:<key>` with
`open/snoozed/waiting/done/dismissed` — a de-facto second board that never touches `kanban.db`;
(3) a personal "to check / to test" list referenced nowhere in the repo or the vault snapshot,
invisible to Tars; (4) Linear, ~40 currently-open issues assigned to Gaetan across three teams
(NM x MC ~15, Software Engineering ~18, Mobile club ~3) — surfaced by Linear's built-in cross-team
"My issues" view, but mixed in there with company work `[lin]`; (5) ad-hoc "Orca one-shot" sessions
dispatched by DM through `delegate-to-cooper`, tracked only in the DM scrollback.
`[repo]` `[vm]` `[lin]` `[obs]`

## 3. Prior art

### 3.1 What this repo already decided

- **The ask is original, not new.** `PITCH.md:66-68` asks in so many words for "a personal Kanban
  board somewhere I can follow, managed mostly by Tars" `[repo]`.
- **`docs/specs/wf5-kanban.md` (2026-08-07) shipped the mechanism, never the policy.** One key —
  `toolsets: [hermes-cli, kanban]` — flips `_profile_has_kanban_toolset()` and unlocks 12 `kanban_*`
  tools on **both** CLI and the live Slack session at once, live-reloaded, no restart. Verified
  live, `NRestarts=0` `[repo]`, corroborated by `docs/facts.md:22`.
- **It was E2E tested, then the board was emptied.** `kanban_create` → `t_4f772ab0` auto-promoted
  `todo→ready` via `recompute_ready()` (the schema docstring saying cards sit in `todo` is stale),
  the in-gateway dispatcher spawned a real worker and drove it `ready→running→done`; repeated from
  Gaetan's live Slack DM (`t_f525e54b`). Both archived; **empty and unused since** `[repo]` `[vm]`.
- **The merge policy was deferred and never picked up.** `PLAN.md:60,82` bundled kanban with
  "dailys · reminders · Orca playbooks" as a later behaviour layer; only the tooling unlock was
  scoped `[repo]`.
- **Nearest built design for capture is unbuilt:** `docs/proposals/P2-emoji-trigger.md` — react
  `:kanban:` → `reaction:added:kanban` → SOUL line creates a card. The gateway already handles
  `reaction_added` (`adapter.py:1996`); needs `slack.reaction_triggers: [kanban]` plus a workspace
  emoji. Config-only `[repo]`.
- **House precedent against over-building triggers:** `status/probes/wf5/kb-push-design.md:495-496`
  scored a kanban-card-driven trigger as not worth v1 — "only earns its keep once the trigger is
  recurring" `[repo]`. (`DECISION.md` is referenced by `PLAN.md` but does not exist — dangling
  trail, not a blocker.)

### 3.2 What Hermes actually offers on the VM (measured, not marketing)

`hermes kanban --help` on v0.20.0 (2026.8.3): "Durable SQLite-backed task board shared across Hermes
profiles. Tasks are claimed atomically, can depend on other tasks, and are executed by a named
profile in an isolated workspace." ~30 subcommands, incl. `create, list, show, assign, claim, block,
schedule, promote, dispatch, notify-subscribe, specify, decompose, gc, repair` `[vm]`. Statuses:
`triage, todo, scheduled, ready, running, blocked, done, review, archived`. Live state: board
`default`, `archived=2`, everything else 0 `[vm]`.

Also on the VM: a `hermes project` layer that can `bind-board` a board 1:1 to a multi-folder
workspace (unused, "No projects yet"); `notify-subscribe`, which backs Slack's `/kanban subscribe`;
`specify`/`decompose`, LLM steps that turn a raw `triage` card into a spec or a child-task graph —
**and `kanban.auto_decompose: true` is a live default that fires on exactly `triage`-status tasks**
`[repo]`, so `triage` is not a quiet parking status; `gc`/`repair` for hygiene. MCP servers wired:
**`slack` and `notion` only — no Linear MCP**; Linear reaches Tars solely through
engagement-checker's direct GraphQL `[vm]`. Six live cron jobs, delivering to Slack channel
**`C0BP2GZUFSR`** *or* `local` `[vm]`; the two report producers — `daily-work-brief` and
`engagement-checker` — are the ones targeting `C0BP2GZUFSR`, so that is the best candidate for the
"configured reporting conversation" ID `[repo]` could not find; *unverified* that cron target and
skill setting are the same thing.

**Neither of those two skills is among the 57 under `~/.hermes/skills/`** `[vm]` §4 — they resolve
under a different profile. Kanban tools are gated per-profile by `toolsets:` `[repo]`, so that
profile may not have them at all, which puts the whole render path in question (see §5.1).

### 3.3 What the outside world does

Three classic patterns `[ext]`: **(a) single source of truth** — cannot re-fragment, costs a
migration; **(b) two-way sync** (Unito/n8n/Zapier) — Unito's own materials concede conflict
resolution is *configured*, not solved, when both sides edit between cycles; **(c) read-only
aggregation** — cheapest, stale between polls, and collapses into (b) the moment you close a card
from the merged view.

File-based `[ext]`: **Backlog.md** is purpose-built for "agent and human both edit git-diffable
markdown"; **Obsidian Kanban** is its equivalent for someone already in Obsidian; todo.txt too flat,
Taskwarrior not diffable. MCP: Linear ships an official remote MCP server; **no mature
product-agnostic kanban MCP server exists**; vibe-kanban targets coding-agent orchestration, not a
personal board. Multi-agent literature converges on `preferences/workflow.md`'s rule —
**single-writer ownership per resource**. Hermes
Kanban's claims + heartbeats + stale-claim reclaim implement that *for task execution* (measured
help text `[vm]`); `[ext]`'s broader "single-writer discipline for editing" reading is secondary
blog material and is not taken as fact here.

Obsidian ground truth `[obs]`: the vault lives on Gaetan's Mac, Syncthing folder `c5efu-kvd7c`
("Obsidian", 26.7 MB, `sendreceive`, shared with mc-mbp / Pixel 6 / the hub). **The Tars VM has no
Syncthing at all**; cooper is not currently a target for that folder **but already runs Syncthing and
already trusts the hub and mc-mbp device IDs, so adding it is a config-only change, not a new
pairing.** The Kanban plugin (mgmeyers v2.0.51) is installed and enabled but has no `data.json` and
zero board files — dormant, never used; `obsidian-tasks-plugin` and `dataview` also enabled. Format
confirmed from the plugin's `main.js`: `kanban-plugin: board` frontmatter, `##` columns, list-item
cards. The vault has a documented sync-divergence history.

## 4. Options

### A — Hermes Kanban as the one board, Slack as the window  (effort: **S**)

State lives in `~/.hermes/kanban.db` on the VM. **Writers:** Tars (`kanban_*` tools) and Gaetan
(Slack DM / `/kanban` / `hermes kanban` CLI). The engine's atomic claims arbitrate worker execution,
not simultaneous field edits `[vm]` — not a real risk for one user. Sources flow in: *personal list*
→ Slack DM, or `:kanban:` reaction (P2, config-only) `[repo]`; *Orca one-shots* →
`delegate-to-cooper` creates a card first, the card ID is the run handle; *daily brief* → renders the
board into its 08:30 post; *engagement-checker* → keeps its private JSON but *optionally, and only on
Gaetan's say-so (§5.3 Q3)* promotes a surviving open loop to a card; *Linear* → **not imported**,
referenced by key/link. Gaetan sees a rendered list in the reporting conversation.

Pros: zero new infrastructure; installed, enabled as of 2026-08-07 and E2E-proven on this VM;
first-party tools, no MCP glue; execution concurrency solved in the engine; `notify-subscribe`
already pushes terminal events to Slack; `specify`/`decompose` offer triage automation (which cuts
both ways — §5.1). Cons: no native visual board (a GUI is claimed by secondary docs `[ext]` —
**unverified on this install**), so the human view is text in Slack; state is one SQLite file on one
VM, a new thing to back up; **the dispatcher will try to execute cards** (§5.1). Constraint fit:
board writes are not repo writes, delegation preserved, KISS maximal — the tool is already there.

### B — Linear as the one board  (effort: **S–M**)

State lives in Linear. Personal + Tars items become issues in an existing team, segmented by labels
(`origin:personal`, `origin:tars`) or one lightweight project; the board artefact is a cross-team
assignee=me view — **which already exists for free as Linear's built-in "My issues"** `[lin]`, so the
board costs nothing to build; only the label taxonomy is real work. Tars writes via the GraphQL path
it already exercises for reads, or as a Linear Agent identity — a pattern this workspace already
built for the Support Engineer agent (NMC-461/465) `[lin]`.

Pros: a good kanban UI Gaetan already uses daily on desktop and phone; states, labels, due dates,
priority, cycles, cross-team views for free; mature API/webhooks; no new infra; ~40 open company
items need no migration `[lin]`. Cons: **no private space** — Mobile Club is a single company
workspace, so personal to-test items are org-visible; mixing personal noise into 40 real company
issues costs legibility; Tars gains issue-*write* authority in the company tracker, a larger blast
radius than a local SQLite file; half-baked ideas become permanent company records. Constraint fit:
acceptable, but it grows Tars' write surface into shared company state.

### C — Obsidian Kanban board, synced  (effort: **M–L** on the VM, **S** via cooper)

State lives in one or two `.md` files in the vault (`kanban-plugin: board`, `##` columns, `- [ ]`
cards `[obs]`); Tars edits the markdown, Gaetan drags cards on Mac/phone. Single-writer by splitting
files: `Inbox.md` Gaetan-only, `Board.md` Tars-only. Two routes: **VM-direct** (install Syncthing on
the Tars VM, pair it to the hub, share `c5efu-kvd7c`) or — cheaper — **cooper-mediated**: cooper
already runs Syncthing and already trusts the hub and mc-mbp device IDs, so sharing the folder to
cooper is a config-only change, not a new pairing `[obs]`; implementation is delegated to cooper Orca
sessions anyway, so Tars' `Board.md` can be written cooper-side and reach Mac/phone over the existing
mesh with **nothing installed on the VM**.

Pros: the only option giving a real drag-and-drop board on the devices Gaetan uses, in a tool he
already runs; plain markdown, no lock-in; plugin already installed; `obsidian-tasks`/`dataview` query
the vault for free. Cons: the VM-direct route adds infrastructure to that host — the cooper variant
removes that cost entirely; the vault has a documented divergence history, and eventually-consistent
sync with two writers produces conflict copies (the file split mitigates, does not eliminate); Tars
drives it by raw text munging — no claims, no event log, no heartbeat, throwing away exactly what A
gets free; the plugin has zero boards here, so the workflow is unproven. Constraint fit: fine on
delegation and pushing; weakest on KISS.

### D — Read-only aggregated view, no unified store  (effort: **S–M**)

Nothing moves. Tars renders one merged board into the reporting conversation, refreshed by the
existing crons, reading Linear (GraphQL), engagement-checker's JSON, `kanban.db`, and a personal file
if one ever becomes visible. Every write stays at the source `[ext]`.

Pros: cheapest; no migration, no new store, no conflict semantics; extends existing crons. Cons: it
does not solve the stated problem — five sources stay five, and moving a card from the merged view
collapses into two-way sync `[ext]`; the personal list stays invisible for lack of a
machine-readable home `[obs]` `[repo]`; staleness between polls. Constraint fit: perfect, by doing
less than asked.

### E — Git-file board (Backlog.md / markdown files)  (effort: **M**)

`.backlog/` markdown files; git is the state machine; Orca sessions and Gaetan both edit; every state
change is a commit `[ext]`. **Verdict: it duplicates the already-installed engine** — no claims, no
event log, no dispatcher, all of which `kanban.db` already gives for free. That, not a rule, is why
it loses; it fails for the same real reason C does. The "Tars never pushes" hard rule bans
pushing/merging/approving, not local file edits or commits, so it **only becomes constraint-relevant
if you want the board pushed to a remote**.

### F — Notion database with a board view  (effort: **S–M**)

State lives in a Notion database in Gaetan's private space, rendered as a native board view. **The
Notion MCP server is one of only two wired on Tars today (`slack`, `notion`, both enabled,
`NOTION_TOKEN` already in config) `[vm]`** — the tool path exists and needs no new integration.

Pros: kills B's headline con (a personal Notion page is private by default — no company visibility)
*and* A's headline con (native drag-and-drop board views on Mac, iOS and web) at once; zero new
infrastructure; Tars writes through an already-wired MCP server. Cons: no claims, no heartbeats, no
event log, no dispatcher — the same duplication objection as C and E, so anything Tars *executes*
still needs `kanban.db` underneath or a second store appears; a remote API in the write path (rate
limits, latency, network dependency) where A is a local SQLite file; **unverified** whether the wired
Notion token can create databases and whether Tars' Notion access is scoped to a personal space or a
shared team one. Constraint fit: fine on delegation and pushing; it is the strongest runner-up to A
on *human ergonomics*, and the right answer if the Slack text view proves insufficient.

## 5. Recommendation

**Take option A: Hermes Kanban is the one board; Slack is the window; Linear is referenced, never
synced.** Only A is already installed, enabled as of 2026-08-07, proven E2E on this VM, and carries a
real dispatcher/event log. **F (Notion) is the strongest runner-up** — private by default, native
board views, MCP already wired — and is the right upgrade if the Slack text view disappoints; it does
not replace `kanban.db` for anything Tars executes. B's board costs nothing (built-in My-issues view)
but puts personal noise in a company workspace and hands Tars write authority over company records.
C is worth doing **later, as a Tars-written read-only `Board.md` mirror via cooper** — not by
installing Syncthing on the VM (S, not M–L). D under-delivers; E duplicates the installed engine.

### 5.1 Blocking issues to settle first

On today's config, `kanban_create` does **not** park a card: `recompute_ready()` promotes
`todo→ready` and the in-gateway dispatcher (`dispatch_in_gateway: true`, 60 s tick) spawned a real
worker that ran the WF5 smoke card to `done` `[repo]`. So capturing an idea would start executing it
— directly against "Tars orchestrates, implementation is delegated to Orca". Lead with
`kanban.dispatch_in_gateway: false`, or capture into `blocked`/`scheduled`. **Do not reach for
`triage` first:** `kanban.auto_decompose: true` is a live default and fires on exactly
`triage`-status tasks `[repo]`, fanning the card into a child-task graph whose children land in
todo/ready and *are* dispatched — the fix would re-create the problem. `triage` is only safe with
`kanban.auto_decompose: false` set at the same time. **Unverified:** whether an unassigned card with
no `orchestrator_profile` is skipped by the dispatcher; measure on the VM before deciding.

Second unknown, equally blocking for the render path: **which profile the `daily-work-brief` and
`engagement-checker` cron skills resolve under.** Neither is among the 57 skills in
`~/.hermes/skills/` `[vm]` §4, and the kanban tools are gated per-profile by `toolsets:` `[repo]` —
so the skills that §5.2 steps 3 and 5 have calling `kanban_list`/`kanban_create` may not have the
toolset at all. Locate the profile and confirm its effective `toolsets` includes kanban.

### 5.2 Migration sketch

1. **Measure first**, all in one pass on the VM: `hermes config get toolsets` (is kanban still
   enabled after ten config churns since 2026-08-07?), the current value of `kanban.auto_decompose`
   and `kanban.dispatch_in_gateway`, the dispatcher behaviour on an unassigned card as in §5.1, and
   the profile the two cron skills run under plus its effective `toolsets`. Then pick the capture
   status and apply any `kanban:` config change under the standard `.bak` +
   `flock ~/.hermes/.wf3.lock` discipline. Delegated to an Orca session, not done by Tars.
2. **Capture path**: enable `slack.reaction_triggers: [kanban]`, add the `:kanban:` workspace emoji,
   add the one SOUL line from `P2-emoji-trigger.md`. Plain DM capture already works today.
3. **Render path**: extend `daily-work-brief` to open with the board (`kanban_list`, grouped by
   status) in the 08:30 post to the reporting conversation. Gated on step 1 confirming that skill's
   profile actually has the kanban toolset. This is the "one place Gaetan follows".
4. **Orca lane**: `delegate-to-cooper` creates the card before dispatching, comments the result on
   it. Cards, not DM scrollback, become the run log.
5. **Engagement-checker**: leave its JSON queue alone in v1 — it works — and only promote an item to
   a card when Gaetan says so (§5.3 Q3). **So A as recommended still runs two state surfaces in v1 —
   `kanban.db` and `engagement-checker.json`; "one board" is a UI claim, not yet a storage one.**
   Collapsing them is a v2 decision, per the `kb-push-design.md` precedent of not building until the
   need recurs.
6. **Linear**: no writes, no mirror job. Cards that relate to an issue carry its key in the title and
   a link in the body; the brief keeps reading Linear as it does now.

### 5.3 Open questions for Gaetan

1. Where does the "to check / to test" list physically live today? It is in neither the repo nor the
   vault snapshot `[repo]` `[obs]`. Is it a one-time paste into the board, or a file Tars should read?
2. Is a text board in Slack enough, or do you want a real visual board from day one — Notion (F, MCP
   already wired) or an Obsidian mirror (C)? Note the Obsidian route no longer implies Syncthing on
   the VM: cooper can hold the share and write `Board.md`, a config-only change `[obs]`.
3. Should Tars create cards on its own initiative (from the brief, from engagement-checker), or only
   when you tell it to — mirroring SOUL rule 4's "your instruction is the approval, per message"?
4. Confirm `C0BP2GZUFSR` is the "configured Slack reporting conversation" both skills mean — the
   `daily-work-brief`/`engagement-checker` cron targets matched (other crons post to `local`), but
   the skills' own setting was not read `[vm]` `[repo]`.
5. One board or several? `hermes kanban boards` supports per-workstream boards and `hermes project
   bind-board` `[vm]`. YAGNI says one board with labels until it hurts.
6. Backup: `kanban.db` becomes real state on one VM. Add it to whatever backs up `~/.hermes/`, or
   accept the loss risk? (**Unverified** what, if anything, backs up `~/.hermes/` today.)
