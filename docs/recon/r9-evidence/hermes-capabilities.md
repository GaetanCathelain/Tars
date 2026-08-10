# Hermes v0.20.0 live capabilities on Tars VM (192.168.0.9) — kanban research

Probed read-only via `ssh gaetan@192.168.0.9`, `~/.local/bin/hermes ...`. No
sops, no file reads, no edits. 2026-08-10.

## 1. `hermes --version`

```
Hermes Agent v0.20.0 (2026.8.3)
Install directory: /home/gaetan/.hermes/hermes-agent
```

## 2. Hermes HAS a built-in kanban — this is the headline finding

`hermes --help` top-level subcommand list includes `kanban` and `project`
alongside `cron`, `mcp`, `memory`, `skills`, etc.

`hermes kanban --help`:
> Durable SQLite-backed task board shared across Hermes profiles. Tasks are
> claimed atomically, can depend on other tasks, and are executed by a named
> profile in an isolated workspace.
> See https://hermes-agent.nousresearch.com/docs/user-guide/features/kanban
> or docs/hermes-kanban-v1-spec.pdf for the full design.

Subcommands (abridged, all present in `--help`):
`init, boards, create, swarm, list/ls, show, assign, set-model, reclaim,
reassign, diagnostics/diag, link, unlink, claim, comment, attach,
attachments, attach-rm, complete, edit, block, schedule, unblock, promote,
archive, tail, dispatch, daemon (DEPRECATED — dispatcher now runs in the
gateway), watch, stats, notify-subscribe, notify-list, notify-unsubscribe,
log, runs, heartbeat, assignees, context, specify, decompose, gc, repair`.

Notable primitives already built in:
- **Boards**: `hermes kanban boards` — one board per project/workstream,
  multi-tenant (`--tenant`), switchable (`HERMES_KANBAN_BOARD` env or
  `boards switch`).
- **Task graph**: `link`/`unlink` parent→child dependencies; `swarm` creates
  a whole graph (parallel workers → verifier → synthesizer) in one call.
- **Statuses**: triage, todo, scheduled, ready, running, blocked, done,
  review, archived (from `kanban list --status {...}` choices).
- **Assignment to profiles**: `assign`/`reassign`/`assignees` — tasks are
  claimed and executed by a named Hermes profile in an isolated workspace
  (scratch, worktree, or existing dir).
- **Notifications**: `notify-subscribe` backs `/kanban subscribe` in the
  gateway/Slack adapter — i.e. Slack can already subscribe to a task's
  terminal events.
- **Triage automation**: `specify` (flesh out a triage card into a concrete
  spec via `auxiliary.triage_specifier` LLM) and `decompose` (fan a triage
  card into a child-task graph via `auxiliary.kanban_decomposer`) — this is
  effectively an LLM-driven "convert raw idea into structured tickets" step,
  built in.
- **Ops hygiene**: `gc` (garbage-collect archived workspaces/events/logs),
  `repair` (SQLite integrity auto-repair), `dispatch` (one dispatcher pass;
  the docs note the dispatcher itself now runs inside the gateway process,
  not as a separate daemon — so it's live whenever the gateway is up).

### Current live state (read-only probe)

```
$ hermes kanban boards list
    SLUG      NAME       COUNTS
●   default   Default    archived=2
Current board: default

$ hermes kanban stats
triage 0, todo 0, scheduled 0, ready 0, running 0, blocked 0, done 0

$ hermes kanban list --status archived
— t_4f772ab0  archived  default  WF5 kanban smoke card
— t_f525e54b  archived  default  WF5 kanban slack smoke card
```

`hermes kanban show t_4f772ab0` confirms this board was already smoke-tested
during WF5 (per repo's `docs/specs/wf5-kanban.md`): task created → claimed →
spawned worker → completed → archived, full event log intact. **So the
kanban engine is not hypothetical — it already round-tripped a task on this
exact VM.** Board is currently empty/idle (0 active tasks) — nothing is
using it as the real work tracker yet.

### `hermes project --help`

> Projects are human-named workspaces that can span multiple folders/repos.
> They anchor desktop session grouping and, when bound to a kanban board,
> give tasks a deterministic worktree + branch convention.

`bind-board` subcommand exists — a kanban board can be bound 1:1 to a named
project (multi-folder workspace). Currently: `hermes project list` → "No
projects yet."

### Docs on disk

- `/home/gaetan/.hermes/hermes-agent/docs/hermes-kanban-v1-spec.pdf` — full
  design spec, present locally (also cached at
  `/tmp/hermes-agent-docs/docs/hermes-kanban-v1-spec.pdf`).
- `/home/gaetan/.hermes/hermes-agent/docs/kanban/multi-gateway.md` — a doc
  specifically about kanban across multiple gateways (Slack/WhatsApp/etc?),
  not read in full (read-only structural probe only; worth a follow-up read
  if kanban is chosen as the answer).
- `/home/gaetan/.hermes/hermes-agent/hermes-already-has-routines.md` — a
  repo-root doc whose filename alone is a strong signal Hermes core team
  already pre-empted "don't reinvent cron/routines" — relevant precedent for
  not reinventing kanban either. Not read in full (out of scope for
  read-only structural probe, but the filename is itself evidence worth
  citing verbatim).

## 3. MCP servers currently wired on Tars

```
$ hermes mcp list
  Name     Transport        Tools   Status
  slack    docker run -i    all     enabled
  notion   docker run -i    all     enabled
```

Confirmed via `config.yaml` `mcp_servers:` block (names/keys only, no
secrets printed):
```
mcp_servers:
  slack:
    command: docker
  notion:
    command: docker
    env:
      NOTION_TOKEN: ${NOTION_API_TOKEN}
```

**No Linear MCP wired on Tars today.** The engagement-checker skill reads
Linear via direct GraphQL (per its own description/memory note), not via an
MCP server — so Linear access exists on Tars but outside the MCP layer.
Backup filenames in `~/.hermes/` (`config.yaml.bak-linear-notion-cal`,
`.env.bak-linear-notion-cal`) show a Linear+Notion+Calendar integration was
attempted/rolled back at some point (2026-08-07) — filenames only, contents
not read.

## 4. `~/.hermes/` state stores (`ls -la`, structure only)

Relevant to hosting a kanban:
- `kanban.db` (+ `-shm`/`-wal`, WAL mode) — the SQLite kanban board itself,
  118 KB currently (idle/near-empty).
- `kanban/` directory — task workspaces + logs (`kanban/logs/`,
  `kanban/workspaces/<task_id>/`, confirmed by the smoke-test task's
  `workspace:` path).
- `projects.db` — separate SQLite db for `hermes project`.
- `state.db` (92 MB, WAL) — general agent state, not kanban-specific.
- `lcm.db` (56 MB, WAL) — "large context memory" / LCM store (there's an
  `hermes-lcm` skill installed — see below).
- `memories/` — directory for built-in memory (MEMORY.md/USER.md per
  `hermes memory --help`).
- `cron/` — cron job store backing `hermes cron list`.
- `skills/` — 57 skills installed today (see full list below), no
  kanban/todo/task-specific skill among them.
- Many `config.yaml.bak-*` / `SOUL.md.bak-*` snapshots — evidence of heavy
  recent iteration (b5-memory, b5-adhd, kanban, linear-notion-cal, mesh-model,
  notionpin, slack, prflow, skillcommit, p4, p6, identfix) — none read.

`hermes memory --help`:
> Available providers: honcho, openviking, mem0, hindsight, holographic,
> retaindb, byterover. Only one external provider can be active at a time.
> Built-in memory (MEMORY.md/USER.md) is always active.

This is a memory/recall subsystem, not a task tracker — not a kanban
candidate, but confirms Hermes already separates "memory" from "kanban" as
distinct built-in subsystems.

### Installed skills (57, `ls ~/.hermes/skills/`)

```
apple, ask-matt, autonomous-ai-agents, claude-handoff, codebase-design,
code-review, communication, creative, delegate-to-cooper,
diagnosing-bugs, domain-modeling, email, git-guardrails-claude-code, github,
grilling, grill-me, grill-with-docs, handoff, hermes-lcm, hermes-operations,
i-have-adhd, implement, improve-codebase-architecture, last30days, loop-me,
mattpocock-research, media, migrate-to-shoehorn, mlops, note-taking,
orchestration, productivity, prototype, research, resolving-merge-conflicts,
scaffold-exercises, setup-matt-pocock-skills, setup-pre-commit,
setup-ts-deep-modules, smart-home, social-media, software-development, tdd,
teach, to-questionnaire, to-spec, to-tickets, triage, wait-what, wayfinder,
wizard, writing-beats, writing-for-agents, writing-fragments, writing-shape
```

Note: `daily-work-brief` and `engagement-checker` (the two skills already
producing "kanban-like" Slack output per the task brief) are NOT in this
list — they must live under a different profile's skills dir or be delivered
purely via cron (`hermes cron list` shows both running with
`Skills: daily-work-brief` / `Skills: engagement-checker`, so they exist
somewhere in the skill-resolution path even if not in this profile's
`~/.hermes/skills/`). Worth a follow-up if pursuing the kanban route, since
whichever skill ends up "owning" the merged board will need to sit alongside
(or replace) these two.

## 5. `hermes cron list` — live jobs (names/schedules only, no tokens)

6 active jobs, all delivering to Slack channel `C0BP2GZUFSR` (or `local`):
- `Gaetan daily work brief` — 08:30 Mon-Fri, skill `daily-work-brief`
- `Gaetan engagement checker` — every 30 min, 10:00-16:00 Mon-Fri, skill
  `engagement-checker`
- `Gaetan engagement checker final pass` — 17:00 Mon-Fri, skill
  `engagement-checker`
- `Forward new Claude invoices to Olivier` — 09:30 daily, script-mode
- `mc-metarepo-refresh` — every 60 min, script-mode
- `Fallback: Tars Slack channel boundaries` — one-shot, 15 min

Cron is a real scheduling primitive already in daily use, separate from
kanban. If kanban becomes the merged store, these two report-producing crons
are the natural candidates to be rewired to read/write kanban tasks instead
of (or in addition to) posting free-text Slack messages.

## Answer to the grounding question

Tars/Hermes v0.20.0 has a **first-class, already-tested, SQLite-backed
kanban engine built in** (`hermes kanban`), plus a **projects layer** that
can bind 1:1 to a board, task dependency graphs, profile-based assignment,
Slack event-subscription hooks, and LLM-driven triage-to-spec automation
(`specify`/`decompose`). It was smoke-tested successfully on this exact VM
during WF5 and is currently idle (0 live tasks, board `default` exists).
Cron (6 active jobs) and MCP (slack, notion — no Linear MCP) are the other
two primitives available; Linear access today goes through direct GraphQL
in the engagement-checker skill, not MCP.

This means the research question ("do we need new infrastructure to merge
5 kanbans?") has a strong "no" candidate sitting unused: `hermes kanban`
already satisfies KISS — it's an existing, installed, working feature, not
something to build. The open design question is not "build vs. buy" but
"how do daily-work-brief, engagement-checker (Linear), the personal
to-check list, and ad-hoc Orca one-shots all funnel INTO `hermes kanban`
boards" — e.g. cron jobs `create`/`comment` cards instead of just posting to
Slack, and `notify-subscribe` pushes terminal events back to Slack for
visibility. That merge design is out of scope for this read-only probe.
