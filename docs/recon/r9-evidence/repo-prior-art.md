# Kanban unification — prior art inside the Tars repo

Repo checked out at `/home/gaetan/dev/Tars/.claude/worktrees/kanban-research` (worktree of the
Tars build+ops repo). All paths below are relative to that root unless a full path is given.

## 1. The original ask (PITCH.md)

`PITCH.md:66-68` — this is the founding requirement, word for word:

> Be my personnal assistant. I've been having issues with reporting and overall structuring and
> orchestrating my work, especially since I'm going so fast with AI now. I want to have a personal
> Kanban board somewhere I can follow, managed mostly by Tars and Tars should remind me of what I
> should do, prepare my dailys, create reports, orchestrate Orca on my Cooper VM to start working
> on tasks, ...

So "one kanban Tars operates" was the original vision, not a new idea — it just never got built
as its own workstream; the WF5 work (below) built the *mechanism* (Hermes kanban toolset) but not
the *policy* of what goes on the board or how it reconciles with Linear / dailies / personal lists.

## 2. `docs/specs/wf5-kanban.md` — what it decided (full read)

Dated 2026-08-07. This spec documents a narrow, already-completed **enablement** change, not a
board-merging design:

- **The change**: one top-level key added to `~/.hermes/config.yaml`:
  ```yaml
  toolsets:
  - hermes-cli
  - kanban
  ```
  This flips `_profile_has_kanban_toolset()` (`tools/kanban_tools.py:52`) to `True`, which gates
  both `_check_kanban_mode` (12 lifecycle tools: `kanban_create`, `kanban_show`, `kanban_complete`,
  `kanban_block`, `kanban_comment`, `kanban_link`, `kanban_attach*`, `kanban_heartbeat`) and
  `_check_kanban_orchestrator_mode` (`kanban_list`, `kanban_unblock`) for **every platform at
  once** — this single key, not a per-platform setting, unlocks Hermes's model-facing kanban tools
  on both CLI and the live Slack session.
- **Why this key specifically**: the `kanban` toolset was already present in `enabled_toolsets`
  for `cli` and `slack` (recovered via a non-configurable-toolset fallback in
  `hermes_cli/tools_config.py:2380+`, since `kanban_*` names live in `_HERMES_CORE_TOOLS`); the gate
  function was the only thing filtering the tools back out. This *corrects* an earlier recon
  document (`status/probes/wf5/kanban-recon.md`) that had concluded no config key could ever grant
  the live Slack session these tools — that recon's central claim was wrong, per
  `status/probes/wf5/kanban-implement.md` (measured evidence) and noted as a correction in
  `status/lane-a.md:177`.
- **Live-reload**: `load_config()` is mtime-cached, gateway recomputes `enabled_toolsets` per turn,
  `check_fn` results TTL-cached 30s — no restart needed. Confirmed live (`NRestarts=0`).
- **Storage**: `~/.hermes/kanban.db` (SQLite/WAL) — pre-existing before this change, one board
  named `default`, 0 tasks at spec time. Dispatcher lock `~/.hermes/kanban/.dispatcher.lock`
  present (dispatcher live in the gateway process). Worker logs land in `~/.hermes/kanban/logs/`
  on first spawned worker.
- **Deliberately NOT enabled**: no `kanban:` config block was written, so all `kanban.*` defaults
  stand: `orchestrator_profile: ""` (no orchestration topology), `auto_decompose: true` (harmless —
  only fires on `triage`-status tasks, and both `kanban_create` and `hermes kanban create` default
  new tasks to `todo`, not `triage`), `dispatch_in_gateway: true` / `dispatch_interval_seconds: 60`
  (pre-existing defaults, already active).
- **Rollback**: `flock ~/.hermes/.wf3.lock -c 'cp ~/.hermes/config.yaml.bak-kanban ~/.hermes/config.yaml'`.

### Was it implemented? Yes — and E2E tested, then cleaned up

Cross-checked against `status/probes/wf5/kanban-implement.md` and `status/probes/wf5/kanban-test.md`:

- Implement probe (2026-08-07 ~20:21 UTC): backup taken (`config.yaml.bak-kanban`), diff applied,
  confirmed via `hermes config get` that `toolsets(effective) = ['hermes-cli','kanban']` and both
  check functions flip to `True`. CLI got 12/41 tools tagged kanban; Slack session too. A live
  Slack turn (`hermes chat -Q`) successfully called `kanban_list`/`kanban_show`.
- E2E test (`kanban-test.md`): created a real card via CLI (`kanban_create` → task `t_4f772ab0`),
  confirmed `kanban stats` moved it `todo→ready`(auto-promoted by `recompute_ready()`, a spec
  correction: `kanban_create` does **not** leave cards sitting in `todo`, the schema docstring is
  stale — the dispatcher promotes and picks it up within its 60s tick), dispatcher spawned a real
  worker subprocess that drove the card `ready→running→done`, worker log confirmed lifecycle. Same
  full lifecycle then repeated from the **live Slack session** (Gaetan DM'd Tars: "kanban test:
  ... kanban_create ..." → task `t_f525e54b`, completed). Finally both test cards were cleaned up:
  `hermes kanban archive t_4f772ab0 t_f525e54b` (soft-delete — rows remain in the DB, logs kept).
- **Net state as of the spec/tests**: Kanban tooling is fully live and proven end-to-end (dispatcher
  ready→running→done, CLI and Slack both), but the **board itself is empty** (archived test cards
  only) — nothing productive has been put on it. No policy exists yet for what goes on it, when,
  or how it reconciles with Linear/dailies/personal lists.
- Two follow-on ideas were scoped but **explicitly not built**:
  - `docs/proposals/P2-emoji-trigger.md` / `status/probes/wf5/emoji-trigger.md`: react `:kanban:`
    to a Slack message → routes as `reaction:added:kanban` → one SOUL line maps that token to
    "create a Kanban card from the reacted-to message". Needs `slack.reaction_triggers: [kanban]`
    config (feature already shipped and handled at `adapter.py:1996`, just needs enabling) plus a
    custom `:kanban:` workspace emoji. **Not implemented** — this is the closest existing design to
    "quick capture into the kanban from Slack" and directly relevant to merging the personal
    to-check/to-test list into the board.
  - `status/probes/wf5/kb-push-design.md:495-496` — a "Kanban-card-driven push" trigger for the
    unrelated mc-metarepo knowledge-base push workflow was scored **not worth building for v1**
    ("only earns its keep once the trigger is recurring, which v1 does not have") — a real
    precedent of Gaetan/the reviewing agent explicitly rejecting kanban-as-trigger machinery until
    there's recurring demand. Relevant precedent for "don't over-engineer the merge, either."

## 3. Cross-checks

### `docs/specs/wf5-orca-delegation.md`
No kanban-merge content — this spec is entirely about the separate `delegate-to-cooper` mechanism
(Orca sandbox, `delegate.sh` fixed-flag entrypoint, guardrail matrix). Confirms WF5 shipped two
independent things together (Orca delegation + Kanban enablement), not a combined design.

### `PLAN.md` — Amendments (`:185-218`) and Phases (`:60-87`)
- Line 60: `WF5["WF5 · later — kanban · dailys · reminders · Orca playbooks"]` — kanban was grouped
  with dailies/reminders/Orca playbooks as one deferred "behavior layer," not separately scoped.
- Line 82 (WF5 triage table, 2026-08-07): "**NOW (workflows dispatched):** Orca delegation playbook
  ... · Kanban (Hermes ships gated kanban tooling — `_check_kanban_mode` currently False)." i.e.
  at triage time the ask was just "unlock the tooling," which is exactly what shipped. The
  **LATER** bucket in the same line is "behavior-layer spec (dailys, reminders, reports)" — this is
  where "what goes on the kanban and how the sources merge" was implicitly deferred to and never
  picked back up as its own item.
- No amendment anywhere mentions merging Linear + daily-brief + engagement-checker + a personal
  list into the kanban. The Amendments section (`:185-218`) covers model backend, cutover, secrets
  store, profile strategy, SOUL versions v1-v3 — nothing kanban-specific beyond what's in the
  triage table.

### `DECISION.md`
Does not exist in this repo (`find . -iname "DECISION*"` returns nothing at top level). PLAN.md's
Amendments section says "Detail in `docs/recon/DECISION.md`" but that path is also absent —
apparently superseded/removed, or the decisions got folded directly into PLAN.md's Amendments
instead. Not a blocker for this research, just noting the doc trail has a dangling reference.

### `docs/facts.md`
Three kanban-relevant fact rows, all corroborating the spec:
- `:16` — config.yaml live-reloads ~30s, no restart needed (kanban enable needed none).
- `:22` — "Kanban: single top-level `toolsets: [hermes-cli, kanban]` key unlocks 12 `kanban_*`
  tools on CLI **and** Slack; dispatcher spawns real workers that drive cards ready→running→done."
- `:109` — `reaction_added` is already wired in the gateway (`adapter.py:1996`); the shipped config
  key `slack.reaction_triggers` (list form) triggers on ANY message; "Kanban's 12 `kanban_*` tools
  are already live, so the *action* exists — only the *trigger* is off." This is the load-bearing
  fact for cheaply wiring "react to add to kanban" later.

### `status/lane-a.md` (recent entries)
Line 169-179 (2026-08-07, "WF5 now-items SHIPPED"): confirms both Orca delegation v1 and Kanban
shipped same day, kanban "full lane proven incl. dispatcher spawning real workers
(ready→running→done), board archived clean." No later lane-a entry revisits kanban policy/usage —
searched the whole file for "kanban" and these are the only hits. **The board has sat empty and
unused since 2026-08-07**; nothing in the log shows Tars or anyone putting real work items on it
since the smoke tests.

## 4. Skills mirrors — how work is tracked/reported *today* (bypassing the kanban entirely)

`skills/` contains: `daily-work-brief/`, `delegate-to-cooper/`, `engagement-checker/`,
`hermes-operations/hermes-orchestration/`, `secure-delta-collectors/`. Both relevant skills are
read-only report generators — **neither reads from nor writes to `~/.hermes/kanban.db`**:

- **`daily-work-brief/SKILL.md`** (v1.0.0): scheduled/on-demand workday brief. Builds a private
  evidence ledger (`time | source | outcome | people | follow-up | evidence handle`) from Slack
  (Gaetan-authored, `filter_users_from=U08BDJAMSRZ`), Cooper (Claude transcripts, git logs, shell
  history over `ssh cooper`), GitHub (merged PRs via `gh`), Linear ("any already-configured Linear
  integration ... If no integration is reachable, omit the metric"), and email/calendar. Classifies
  each row Done/Moved/Open-loop/Noise, then produces a **stateless** message (no durable file
  mentioned) with sections "Since the last daily," "Stats," "Today." Finds its own window by
  searching for "the timestamp of the last successful daily-work-brief delivery **in the configured
  Slack reporting conversation** or cron history" (`:29`) — i.e. **Slack itself is the durable
  store for daily-brief history**, not any file or the kanban DB. This "configured Slack reporting
  conversation" phrase was introduced by commit `9c8dc1a` ("follow configured reporting
  conversation (#35)", 2026-08-10) replacing the prior wording "the origin Slack conversation" —
  i.e. as of this commit the reporting conversation became Hermes-cron-configurable rather than
  hardcoded to wherever the cron message originated. No channel ID is recorded anywhere in this
  repo (that's a live Hermes cron-job config value on the VM, not repo state).
- **`engagement-checker/SKILL.md`** (v1.2.0): half-hourly monitor for open loops (commitments,
  unanswered asks, blockers). Unlike daily-brief, this one **does** have durable state:
  `~/.hermes/state/engagement-checker.json` (JSON, versioned, per-source cursors, `items{}` keyed
  by stable IDs `slack:<channel>:<ts>` / `email:<threadId>` / `linear:<issue-key>`, statuses
  `open/snoozed/waiting/done/dismissed`, up to 500 stable IDs retained, done/dismissed items purged
  after 14 days, stale-open dismissed after 30). Reads Linear directly via a hardcoded Python
  `urllib` GraphQL collector (no CLI/MCP — a deliberate hardening: query-allowlist only, no
  mutations, key never leaves the process, no redirects, byte-capped responses) — companion commit
  `2a75d93` "use direct Linear GraphQL reads." Latest commit `b12d802` "read decisions from
  reporting conversation (#36)" — makes engagement-checker inspect new messages in "the configured
  Slack reporting conversation" for decisions about existing items (a reply naming a `short_id`
  authoritatively resolves it). So **both skills now share the concept of one configured "reporting
  conversation"** as of 2026-08-10, but its literal channel ID lives only in Hermes's live config
  on the VM, not in this repo — I could not find it by grep.
  This skill also has its own **separate durable queue** (`items{}` in the JSON state file) that is
  conceptually a mini kanban (statuses, IDs, snippets, due dates) but is deliberately **not** the
  Hermes kanban — it's private to the skill, never touches `kanban.db`.

Net: today Gaetan effectively has **three separate "boards"** already, none unified: (1) Hermes's
real (but empty/unused) `kanban.db`, (2) engagement-checker's private JSON queue of open loops, (3)
Linear as the actual work tracker, plus daily-work-brief which is a stateless narrative digest with
no queue of its own (it re-derives from sources each run). The personal "to check / to test" list
is not referenced anywhere in this repo by name or by any grep for `backlog`, `to.check`, `to.test`,
or `task list` (only unrelated hits: a generic "the risk to check", the graph-engineering-research
doc's mention of a generic "shared task list" concept in an unrelated context about multi-agent
frameworks, and log-forensics' "the task asked to check coverage"). **The personal list is not
digitized/referenced anywhere Tars-side; it lives entirely outside this repo/VM** (Gaetan's own
notes, not surfaced to Tars).

## 5. Slack "reporting conversation" — what it actually looks like today

Could not find a documented channel ID for the daily-brief/engagement-checker "reporting
conversation" anywhere in the repo (grepped for "reporting conversation", "reporting channel", and
Slack channel-ID patterns `C0[A-Z0-9]{8,}` — only unrelated channel IDs turned up, e.g.
`C08RWSTU9LK` used throughout WF4/WF5 probes as a *test* channel `#gcn-sandbox`, and `C04LZBBNVNY`
as a stray tool-trace dump target). This is a live Hermes cron-job setting on the VM, not repo
state.

Per instructions I loaded the Slack MCP tool and read the last ~30 messages of Gaetan's direct DM
with Tars (bot user `U0BBH85NAKH`, DM channel `D0BBYNM01BL`) since that's the one channel ID this
repo's own probes consistently use for live human↔Tars traffic. Contents of that window
(2026-08-07 21:06 through 2026-08-10 14:06): gateway restart/shutdown notices, a cronjob-reminder
delivery ("wf4-p13-cron OK"), the two live kanban smoke-test exchanges quoted in §2 above, several
ad-hoc delegate-to-cooper requests from Gaetan, and general Q&A threads (SOUL rule 4 quote, A2A
questions, ruflo install question). **No daily-work-brief or engagement-checker delivery message
appears in this 30-message window** — either their actual delivery target is a different
conversation (most likely, given the skills call it "the configured Slack reporting conversation"
as a distinct concept from ad-hoc chat), or none has fired successfully in the covered period. This
is a real gap: I could not confirm, from this repo or this one Slack read, what a live daily brief
or engagement reminder message currently looks like in practice. Recommend a follow-up read of
whatever channel the Hermes cron jobs for `daily-work-brief`/`engagement-checker` are actually
configured to deliver to (check `hermes cron list`/`hermes profile show` on the VM, not this repo).

## 6. Summary for the research question

- The **mechanism** for a Tars-operated kanban already exists, is enabled, and is proven working
  end-to-end (Slack- and CLI-triggered card creation, dispatcher-driven ready→running→done). This
  is the cheapest possible "board" to reuse — it needs zero new infrastructure, per Gaetan's
  KISS/no-new-infra preference.
  - This mechanism is Hermes's SQLite-backed kanban (`~/.hermes/kanban.db`) plus its CLI/Slack
    tool surface (`kanban_create`, `kanban_list`, `kanban_show`, `kanban_complete`, etc.) and an
    in-gateway dispatcher that can auto-spawn a worker per ready task.
- What's **missing** is entirely policy, not plumbing: (a) a decision to route personal
  to-check/to-test items and Linear items into it (no ingestion path built — the emoji-trigger
  design in P2 is the nearest sketched mechanism, unbuilt), (b) reconciliation with Linear (today
  fully separate — daily-work-brief and engagement-checker both read Linear directly, neither
  writes back to it or to the kanban), (c) reconciliation with engagement-checker's own private
  JSON queue (a de facto second board), (d) any use of the kanban board as the daily-brief's
  window/history store (today that's Slack itself).
- Explicit precedent exists in this repo for **not** over-building a kanban-trigger mechanism
  until there's proven recurring need (`kb-push-design.md` rejection of a "kanban-card-driven
  push" trigger) — worth citing back to Gaetan as house doctrine when scoping the merge design.
- The board has been empty/idle since the 2026-08-07 smoke test; nothing since has attempted real
  use, so this genuinely is greenfield policy work on top of already-shipped, already-tested
  plumbing.

## Sources cited
- `docs/specs/wf5-kanban.md` (full read)
- `docs/specs/wf5-orca-delegation.md` (grep, cross-check)
- `PLAN.md:60,82,185-218`
- `docs/facts.md:16,22,109`
- `status/lane-a.md:165-179`
- `status/probes/wf5/kanban-recon.md`, `kanban-implement.md`, `kanban-test.md`, `emoji-trigger.md`,
  `kb-push-design.md:495-496`
- `docs/proposals/P2-emoji-trigger.md`
- `skills/daily-work-brief/SKILL.md`, `skills/engagement-checker/SKILL.md`
- `PITCH.md:63-68`
- git log: commits `9c8dc1a`, `b12d802`, `2a75d93`
- Slack MCP `slack_read_channel(channel_id=U0BBH85NAKH)` — Tars DM `D0BBYNM01BL`, last 30 messages
  (2026-08-07 21:06 to 2026-08-10 14:06 CEST)
- `find . -iname "DECISION*"` (no result — dangling reference in PLAN.md)
- `grep -rniE "kanban|backlog|to.check|to.test|task list"` (full repo, capped at 200 hits)
