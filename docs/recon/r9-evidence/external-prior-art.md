# External prior art: merging N task sources into one kanban

Research date: 2026-08-10. Sources are cited inline; WebSearch summaries flagged as such (not primary-doc quotes) where I could not fetch the primary page directly.

## 0. Headline finding: Hermes already has a native kanban

The Hermes framework Tars runs on (NousResearch/hermes-agent, ~/.hermes layout,
SOUL.md, profiles, gateway, skills — this is confirmed the same project: its
docs describe exactly Tars' architecture, see §4) **ships a built-in
multi-agent kanban board**, added around v0.12 and hardened in v0.17
("Tenacity" release: durable multi-agent kanban, subagent delegation,
background curator).

- Doc: https://hermes-agent.nousresearch.com/docs/user-guide/features/kanban
  (mirror: https://github.com/NousResearch/hermes-agent/blob/main/website/docs/user-guide/features/kanban.md)
- Feature name: **"Hermes Kanban" / "Kanban (Multi-Agent Board)"**
- Storage: SQLite, `~/.hermes/kanban.db` (default board) or
  `~/.hermes/kanban/boards/<slug>/kanban.db` (named boards). Attachments under
  `<hermes-home>/kanban/attachments/<task_id>/`. Append-only `task_events`
  table for history, `task_runs` for per-attempt metadata.
- Two front doors, same DB:
  - **Agent-driven**: dedicated tool namespace — `kanban_show`, `kanban_list`,
    `kanban_complete`, `kanban_block`, `kanban_heartbeat`, `kanban_comment`,
    `kanban_attach`, `kanban_create`, `kanban_link`, `kanban_unblock`.
  - **Human-driven**: CLI (`hermes kanban ...`), Slack/chat slash commands
    (`/kanban ...`), bundled dashboard GUI.
- Columns/statuses: `triage → todo → ready → running → done` or `blocked →
  archived`. Triage rows expose two LLM actions: Decompose (fan out into
  child tasks routed to specialist profiles) and Specify (spec rewrite).
- A dispatcher loop inside the gateway reclaims stale claims (60s default,
  configurable), promotes ready tasks, spawns assigned profiles as OS
  processes — i.e. it's built to let *multiple Hermes profiles* work one
  board, not just one agent talking to itself.

**Implication for Gaetan's problem**: this is very likely the "operated by
Tars, works well with Hermes" backbone he's after — a board Tars can drive
through its own toolset (no MCP needed, no external service), that a human
can also touch via CLI/slash commands/GUI, backed by one SQLite file that's
easy to inspect/backup. It is NOT the "personal to-check list" or Linear or
the daily-brief output — those still need to feed into it (see §5).
**Caveat**: I did not find the version Tars' Hermes install (v0.20.0) ships
with confirmed as containing kanban — v0.12/v0.17 predate v0.20 in the
version-number sense the docs use, so on the surface it should already be
present, but **verify with `hermes kanban --help` on the VM before
committing to this** — CLI facts here are LLM-summarized secondary sources
(dev.to/Medium posts), not the changelog itself, and this repo's CLAUDE.md
already mandates measuring instead of trusting recon claims.

## 1. Classic integration patterns

### (a) Single source of truth — move everything into one tool
Definition per project-management literature: a System of Record stores data
for one function; a Single Source of Truth aggregates/reconciles multiple
systems of record into one trusted view (Nulab, Wrike, Kohezion — see
sources). Applied here: pick one tool (e.g. Linear, or Hermes Kanban) and
migrate the other four sources' *items* into it as the only place status
lives.
- Failure mode: migration tax (someone re-enters existing "to check" items),
  and it only works if every future item is created there too — otherwise a
  fifth source reappears immediately (the exact problem being solved).
- Best where write frequency is high and stale forks are costly (this is
  the option that structurally *cannot* re-fragment).

### (b) Two-way sync / federation — Unito, Zapier, n8n style
Unito ("Sync Platform") is the canonical vendor: two-way sync across
Jira/Asana/Trello/Azure DevOps Kanban etc., field-level and sub-field-level
mapping, each field chosen as one-way / two-way / no sync
(https://unito.io/platform-overview/, https://unito.io/blog/bidirectional-sync/).
- **Known failure mode named directly in Unito's own materials**: "when both
  systems modify the same record between sync cycles, the sync process must
  decide which change to keep" — i.e. conflict resolution is not solved, it's
  *configured*, and the vendor's own docs concede this is "the primary
  challenge in bidirectional sync." Unito's mitigation: automatic loop
  detection, field/sub-field-level conflict rules, sync retries — but this is
  inherent complexity, not eliminated complexity.
- n8n/Zapier are the DIY version of the same pattern: an Aggregate node (n8n)
  or Zap merges items from multiple sources into one record on a schedule.
  n8n's docs show this as one-way aggregation into a unified view more than
  true bidirectional sync; Zapier's linear step model makes true 2-way sync
  clunkier than Unito's purpose-built field mapper
  (https://n8n.io/integrations/aggregate/, https://community.n8n.io/t/how-to-aggregate-items-from-multiple-executions-into-a-single-execution-in-n8n-node/72910).
- Failure modes generally: sync-cycle conflicts (both sides edited between
  polls), silent field-mapping drift when one side adds a custom field the
  other can't represent, and doubled webhooks/loops if not guarded.
- Given the KISS/YAGNI mandate and "prefer existing tools over new
  infra," standing up Unito or an n8n workflow to bidirectionally sync five
  sources is very likely over-engineering for a single user's personal
  backlog — flag but do not default to this.

### (c) Read-only aggregation — one virtual board, state stays at source
State never moves; a board is *rendered* by reading N sources and merging
the view. No conflict resolution needed because nothing is written back to
the sources by the aggregator — only the source's own native UI/API writes
to it.
- Failure mode: the aggregated view is only as fresh as the last read/poll
  (staleness), and "closing" a card in the aggregated view either (a) is
  disabled — user must go edit the real source, defeating "one board", or
  (b) requires the aggregator to become a writer after all, which collapses
  this pattern back into (b) two-way sync for at least that one operation.
- This is the cheapest pattern to build and the one most aligned with "Tars
  orchestrates and reports" — Tars could poll Linear + daily-brief output +
  engagement-checker output + a personal file, and render one merged kanban
  view, while writes still go to whichever source actually owns that item.
  The open design question is exactly the "closing a card" case above: does
  Tars get write-back permission into Linear/the personal file (turning this
  into limited two-way sync for just the "move card" action), or is the
  merged view genuinely read-only and the human/Tars closes items at the
  source.

## 2. Plain-text / file-based kanbans an agent can operate by editing files

| Format | Storage | Diffable/greppable | Agent-friendly notes |
|---|---|---|---|
| **Obsidian Kanban plugin** | Plain Markdown per board: YAML frontmatter (`kanban-plugin: basic`) + `##` headings as lists + `- [ ]`/`- [x]` items as cards | Yes — plain markdown, git-diffable | Compatible with Obsidian Tasks/Dataview plugins; no proprietary format, no JSON blob (github.com/obsidian-community/obsidian-kanban README) |
| **Backlog.md** | `.backlog/` folder in a git repo, one Markdown file per task (YAML frontmatter for ID/status/assignee + free-form markdown body) | Yes, explicitly designed for it — README states "git becomes your state machine," `git diff`/`git checkout` to review/rollback | Built *specifically* for AI-agent collaboration: ships a CLI agents can drive ("Claude, please take over task 33"), works with Claude Code/Gemini CLI/Codex/MCP-compatible assistants; terminal Kanban view + web UI on top of the same files (github.com/MrLesk/Backlog.md, HN discussion). Strongest fit in this table for "agent edits files, human also edits files." |
| **Taskwarrior** | SQLite-like local DB (not plain text at rest — `.data` files are a binary-ish format, edited via `task` CLI), Taskserver for sync | Not diffable by design — the appeal is the CLI, not the file | Steeper CLI learning curve; conflict-free sync only via Taskserver infra |
| **todo.txt** | One flat plain-text file, one task per line, tag/priority prefix convention | Yes — trivially greppable/diffable, plain text | No subtasks/nesting, no due-date/recurrence support natively — too flat for "backlog with structure," but the simplest possible thing that works for a linear list |
| **GitHub Projects** | Backed by GitHub's own DB/API, not files in the repo | No — it's a hosted product, not a file you diff | Has a real API/webhooks an agent can drive, but it's another external service, not a "just edit files" option |
| **Simple markdown table / YAML in a git repo** | Whatever schema you pick | Yes | Zero-dependency baseline; loses any board/CLI tooling — you'd be rebuilding Backlog.md's job by hand |

Takeaway: **Backlog.md is the closest existing tool to "one file-based kanban
an agent can operate with plain file edits," already built for exactly this
audience** (AI coding agents + human, git-native, diffable). Obsidian Kanban
is the next-best if Gaetan already lives in Obsidian for the personal
to-check list — same diffability, less agent tooling out of the box (an
agent just edits markdown checkboxes, no dedicated CLI/MCP).

## 3. MCP servers for kanban/task management

- **vibe-kanban** (BloopAI/vibe-kanban, Rust) — mature-leaning, actively
  maintained, real npx install (`npx vibe-kanban`). Built to orchestrate
  *coding agents* (Claude Code, Cursor, Codex, etc.) across a Kanban board:
  track task status, review diffs, start dev servers, centralize MCP config
  for the agents it drives. It is simultaneously an MCP *client* (calls out
  to Postgres/Brave Search MCP servers on the agents' behalf) and an MCP
  *server* (exposes itself to be driven). A second project,
  `yigitkonur/mcp-better-vibe-kanban`, wraps it as a pure MCP server.
  Numerous fan forks exist (Horikawaer, StrayDragon, chanta093) — sign of
  real community traction, but also sign the core repo (BloopAI) is the one
  to trust, forks are noise.
  Fit check: vibe-kanban's job is "coordinate coding-agent sessions," which
  overlaps with Tars/Orca's existing delegation flow more than with a
  personal life/work kanban — worth a look for the Orca one-shot lane
  specifically, less obviously for the whole-kanban merge problem.
- **Linear MCP** — Linear ships an **official remote MCP server**
  (`https://mcp.linear.app/sse`), which supersedes community servers.
  `jerhadf/linear-mcp-server` is explicitly deprecated in favor of it.
  Several community servers still exist (`tacticlaunch/mcp-linear`,
  `dvcrn/mcp-server-linear`, `mkusaka/mcp-server-linear`) but with Linear's
  own server live, there's no reason to run a third-party one — and this
  research task's context already notes Tars' engagement-checker uses
  **direct Linear GraphQL**, not an MCP server, so Linear-the-source is
  already solved without new infra.
- **Generic "task management MCP" search** surfaced mostly toy/one-off repos
  and Backlog.md's own MCP compatibility claim (works with "other
  MCP-compatible assistants") rather than a dedicated Backlog.md MCP server
  — its primary interface is its CLI.
- No mature, widely-adopted **generic kanban MCP server** (i.e., one that
  isn't tied to a specific product like Linear or vibe-kanban) turned up.
  The realistic choices are: use Linear's own official MCP for Linear items,
  use Hermes' native `kanban_*` toolset for the Hermes-side board (no MCP
  needed — it's first-party), and treat the personal-list / daily-brief /
  engagement-checker sources as plain files or scripted reads rather than
  hunting for MCP servers that don't exist yet.

## 4. Hermes framework — confirming project identity and the kanban feature

Multiple unrelated GitHub projects are named "hermes" + "kanban" (noise found
during search: `HSIANG-LIN/hermes-kanban-arch`, `PriuS2/HermesKanban`,
`amanning3390/hermes-agent-kanban`) — these are NOT the framework Tars runs
on; they read as unrelated/personal repos riffing on the same name.

**Identity confirmation for the real project**: `NousResearch/hermes-agent`
docs describe: a `~/.hermes` home directory containing `SOUL.md` ("the
personality that follows you everywhere"), `MEMORY.md`, `USER.md`, skills,
profiles (each with its own `config.yaml`, `.env`, `SOUL.md`, skills, cron
jobs, state DB), and a gateway process per profile integrating with
Telegram/Discord/Slack with its own bot token — this matches Tars'
documented layout (`~/.hermes/SOUL.md`, `~/.hermes/skills/<name>/SKILL.md`,
`~/.hermes/logs/`, `~/.hermes/config.yaml`, Slack gateway, `hermes` CLI at
`~/.local/bin/hermes`) closely enough to treat as confirmed, not guessed.
Docs root: https://hermes-agent.nousresearch.com/docs/user-guide/ ,
repo: https://github.com/NousResearch/hermes-agent .

Given that match, §0's kanban feature is the real, exact answer to "does
Hermes have a built-in kanban": **yes, `hermes kanban`, doc name "Kanban
(Multi-Agent Board)."** A third-party dashboard extension,
`amanning3390/hermes-agent-kanban`, adds a nicer Web UI tab on top of the
same feature but is not required to use it.

## 5. How other personal-AI-assistant setups handle a shared human+agent board

General multi-agent-coordination literature (arXiv survey material, Tacnode
blog) converges on the same discipline Gaetan's own workflow.md already
states independently: **single-writer ownership per resource** is the
concrete pattern that avoids race conditions when multiple writers (human +
one or more agents) touch shared state — "only one entity can write to a
specific resource at a time." Applied to a board: either (a) one column/file
per writer (e.g. Tars only appends to a "Tars inbox" column, human only
edits their own column, Linear items are read-only mirrors) so no
row/file is ever contended, or (b) a real backing store with
row-level locking/optimistic concurrency (which is exactly what Hermes
Kanban's SQLite + `task_events`/heartbeat/stale-claim-reclaim design already
gives you for the Hermes-native part of the board).

No public write-up of a *specific* "personal Hermes/Tars-style assistant +
human shared kanban" case study beyond Hermes' own docs turned up — the
closest is Hermes Kanban's own design doc language ("every handoff is a row
anyone can read and write; every worker is a full OS process with its own
identity"), which is Hermes solving exactly this concurrency problem
internally via per-task claims + heartbeats rather than a global lock.

## Recommendation shape (not a full decision — that's for the main research doc)

1. Hermes Kanban is very likely the right backbone: zero new infra (it's
   already in the framework running Tars), gives Tars a first-party toolset
   (no MCP glue), has human surfaces (CLI/slash/GUI) so Gaetan isn't locked
   out, and already solves the single-writer/concurrency problem via
   per-task claims. **Verify it's present in the v0.20.0 install before
   committing** (`hermes kanban --help` / check for `~/.hermes/kanban.db` on
   the Tars VM).
2. Linear stays the system of record for Linear-created items — read via
   the official Linear MCP (or the existing direct-GraphQL approach the
   engagement-checker skill already uses) — do not migrate Linear issues
   into Hermes Kanban; instead have Tars mirror/reference them as read-only
   cards or link out (pattern 1c), avoiding two-way sync conflict handling.
3. The personal "to check/to test" list and daily-work-brief output are
   currently informal — cheapest onboarding is to have Tars create Hermes
   Kanban cards *from* those instead of merging file formats — i.e. push
   items in, don't build a bidirectional bridge (pattern 1a for those two
   sources, applied one-way to avoid Unito-style conflict logic).
4. Skip vibe-kanban / Unito / n8n sync infra for now under YAGNI — they
   solve team/cross-org integration problems Gaetan's single-user setup does
   not have, and the existing framework already offers a native answer.
