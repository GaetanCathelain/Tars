# R4 — Orca control surface on cooper

Recon only. Read-only probes (`--help`, `--json` list/show/status commands, file reads). Nothing
created, started, stopped, or deleted.

## Verdict

Orca ships two distinct coordination layers relevant to Tars:

1. **`orca-cli` primitives** (worktree/terminal create, `orc-tab`/`orc-opus` launchers) — already
   the mechanism PLAN.md's "Coordination contract" is written against, and already working today
   (evidence: 6 live peer sessions visible via `ListAgents`, one of them `orchestrator-11`).
2. **`orca orchestration` layer** (Run/Task/Dispatch/Gate, `worker-start`, `gate-create`) — a
   separate, heavier, structured-DAG system with its own local DB. It is reachable and
   functional on this runtime (capability flags present, RPCs return `ok:true`), but **no Run is
   currently bound** for this project — `task-list`/`gate-list` both fail with `run_required`
   right now, meaning adopting it means standing up new state, not reusing something already in
   motion.

The orchestration layer's own bundled skill guide explicitly tells the caller to use plain
worktree+prompt ("full handoff") — not Task/Dispatch/Gate — for exactly the shape PLAN.md
describes for lane B ("hand off to another worktree", unsupervised). Gates require a Task to
attach to, so a bare cutover gate is not a standalone primitive — it comes bundled with adopting
the whole Task/Dispatch model.

No decision made here — pros/cons for the synthesis agent below.

## Facts with evidence

### 1. Binary and PATH

- `which orca` → `/home/gaetan/.local/bin/orca` (99-byte wrapper file, not the GNOME
  screen-reader `/usr/bin/orca` — `orc-tab`'s own comment flags this exact collision risk and
  hardcodes the absolute path for that reason).
- `orca status --json` → `app.running: true`, `runtime.state: "ready"`, `runtime.reachable: true`,
  `appVersion: "1.4.176"`. Runtime capability list includes `orchestration.contract.v1`,
  `orchestration.federation.v1`, `orchestration.worker-launch-preferences.v1` — the orchestration
  RPC surface is live on this runtime (not gated off).

### 2. `orca --help` — full command surface (top-level groups)

Startup/diagnostics, Accounts, Skills, Environments, Automations, **Projects**, **Repos**,
**Worktrees**, Files, **Terminals**, **Orchestration**, Computer Use, Linear, Mobile Emulator,
**Browser Automation** (`tab`/`snapshot`/`click`/`fill`/... — full Playwright-style surface).
Full text captured during recon; key groups detailed below.

### 3. `orca worktree create` — flags (from `--help`)

```
orca worktree create --name <name>
  [--repo <selector>|--project <id> [--host <host-id>]|--project-host-setup <id>]
  [--agent <id>] [--prompt <text>] [--setup run|skip|inherit]
  [--base-branch <ref>] [--issue <number>] [--linear-issue <id|url>]
  [--comment <text>] [--parent-worktree <selector>] [--no-parent]
  [--run-hooks] [--activate] [--json]
```

- `--agent <id>` launches a known TUI agent (codex, claude, ...) in the first terminal;
  `--prompt` seeds it.
- Selectors: `--repo id:<id>|name:<name>|path:<path>`; `--parent-worktree
  active|id:<repo>::<path>|branch:<b>|issue:<n>|path:<p>|folder:<id>|worktree:<id>`.
- `--no-parent` forces independent lineage (used for genuinely unrelated top-level work).
- Default behavior does not switch the active Orca view; `--activate` reveals it.
- This is exactly what PLAN.md's coordination contract already specifies for the lane-B spawn:
  `orca worktree create --name lane-b --repo path:<repo-root> --base-branch main --json`.

### 4. `orc-tab` / `orc-opus` — what they are, where they live

- `/home/gaetan/.local/bin/orc-tab` (POSIX sh, 6.4 KB) — a wrapper around `orca terminal create`
  / `orca terminal send`. Three modes:
  - `orc-tab <launcher> <title> <brief-path>` — **handoff**: opens a tab in the *current*
    worktree seeded with a one-line pointer to a brief file (avoids embedding a newline that
    would submit early in a TUI).
  - `orc-tab start [-w <worktree-selector>] <launcher> <title> [prompt]` — **new work**: opens a
    tab, optionally in *another* Orca worktree via `-w path:<abs>|id:<id>|branch:<name>` — this
    is the hub-and-spoke spawn path PLAN.md's contract calls out by name.
  - `orc-tab reseed <handle> <brief-path>` — re-sends the seed only if the tab sat down without
    consuming the initial prompt.
  - Safety guards baked in: refuses an empty terminal handle (`assert_handle`) so it can never
    accidentally type into its own session; hardcodes `$ORCA=/home/gaetan/.local/bin/orca` to
    dodge the `/usr/bin/orca` (GNOME screen reader) name collision; treats CLI exit status as
    unreliable and instead branches on the JSON `.ok` field.
- `/home/gaetan/.local/bin/orc-opus` (POSIX sh, 935 bytes) — launches
  `claude --model 'opus[1m]' --settings '{"ultracode":true}' --append-system-prompt-file
  ~/.claude/handoff/ORCHESTRATION-POLICY.md --dangerously-skip-permissions "$@"`.
  Sibling launchers exist for other models: `orc-fable` (same, `--model 'fable[1m]'`); no
  `orc-sonnet` on PATH (`which` reported not found).
  - Explicitly documented gotcha in the script's own comments: never pipe/redirect it (switches
    Claude to print-mode and exits), never insert a flag before `"$@"` (swallows the positional
    prompt), no `-n/--name` (a fixed session name freezes the tab title and kills Orca's
    activity-based auto-renaming).
- **`ORCHESTRATION-POLICY.md`** (`~/.claude/handoff/ORCHESTRATION-POLICY.md`, read in full) is
  what `orc-opus`/`orc-fable` append as extra system prompt. Its §9 "Peer sessions
  (hub-and-spoke)" is close to word-for-word what PLAN.md's coordination contract already
  encodes: spawn via `worktree create` + `orc-tab start -w <path> orc-opus <lane> '<prompt>'`,
  prompt is a pointer to a plan file committed in the repo, single writer per status file,
  commit-and-push as the durable fallback, peers `SendMessage` milestones/blockers as they
  happen not just a final done. **PLAN.md's contract is this policy applied to Tars, not a new
  invention** — confirms lane-B spawn mechanics are the well-trodden path on this machine.

### 5. Session listing / messaging (cross-session, not Orca-specific)

- `ListAgents` (harness tool, not `orca` CLI) — right now shows **6 live peer sessions**,
  including `orchestrator-11 [769ff3]` (idle, started 27m ago) and four `mc-metarepo-*` sessions.
  Confirms cross-session `SendMessage`/`ListAgents` addressing works today on this machine
  without any Orca orchestration Run being bound.
- `orca worktree ps --json` on the current worktree shows a prior agent pane already exists here:
  prompt `/start-orca`, state `done`, `lastAssistantMessage` referencing a tab `start-fable`
  already spawned via `orc-fable` — i.e. evidence this exact spawn mechanism was just exercised
  live in this repo (not hypothetical).

### 6. Orca orchestration primitives — Run / Task / Dispatch / Gate / Worker

Full subcommand list (`orca --help` "Orchestration" section):
`run-create, run-use, run-current, run-list, run-show, send, check, ask, reply, inbox,
task-create, task-list, task-update, dispatch, dispatch-show, worker-start, worker-show,
worker-read, worker-stop, worker-abandon, worker-release, worker-retain, worker-list,
coordinator-start, coordinator-stop, gate-create, gate-resolve, gate-list, reset`.

Model (from `orca skills get orchestration --full`, the bundled canonical guide — fetched via
`orca skills get`, not assumed):

- **Run** = durable namespace + coordinator inbox only; never schedules or places workers.
- **Task** = a work item (`--spec`, `--deps`, `--parent`); statuses `pending/ready/dispatched/
  completed/failed/blocked`.
- **Dispatch** = one Task attempt bound to a terminal; holds lifecycle authority (not the Run,
  not the terminal handle).
- **Gate**: `orchestration gate-create --task <task_id> --question <text> [--options <json>]` /
  `gate-resolve --id <gate_id> --resolution <text>` / `gate-list [--task <id>] [--status <s>]`.
  **A gate blocks a Task** — there is no gate primitive independent of the Task/Dispatch model;
  using `gate-create` for the cutover means the cutover must first exist as an Orca Task.
- **worker-start** (preferred supervised path) composes worktree+terminal+dispatch in one call;
  supports `--worktree current|new-child|new-top-level|<selector>`, `--agent`, `--model`,
  `--effort`, `--on <remote-environment>`. Exits 0 only on `ready`; failed/unknown exits 1 with
  `stage`/`effects`/`residualResources` for recovery.
- **Messaging**: `send --type worker_done --outcome succeeded|failed --task-id --dispatch-id
  --files-modified --report-path`, `check --wait --types worker_done,escalation,question
  --timeout-ms <n>` (blocking wait loop, no sleep/poll), `ask`/`reply` for worker→coordinator
  blocking questions, group addresses `@all @idle @claude @codex ... @worktree:<id>`.
- **The guide's own tool-boundary rule** (verbatim intent): use plain `orca-cli` (worktree
  create + `--agent`/`--prompt`, no lifecycle tracking) for **"full handoffs"** — explicitly
  including the phrase **"another worktree"** — unless the user *explicitly* asks to supervise,
  monitor, wait for `worker_done`, track completion, coordinate a DAG, or use a decision gate.
  It says outright: **do not run `task-create`, `dispatch --inject`, or `check --wait` for full
  handoffs.**
- Housekeeping calls out `coordinator-start`/`coordinator-stop`/legacy `run`/`run-stop` as
  **"retired scheduler commands... perform no effects"** — i.e. parts of this layer have already
  been deprecated/replaced once.

### 7. Current live orchestration state (evidence, not assumption)

```
$ orca orchestration run-list --json
→ ok:true, runs: [{ id: "run_legacy_local", objective: "Legacy orchestration state
  (inspect only)", legacy: 1, ... }]   # only a legacy tombstone, nothing active

$ orca orchestration task-list --json   (no --run bound)
→ ok:false, error.code: "run_required",
  "No Run is bound. Use orchestration run-create or run-use first."

$ orca orchestration gate-list --json   (no --run bound)
→ ok:false, error.code: "run_required"  (same)

$ orca orchestration worker-list --json
→ ok:true, workers: [], counts: {}      # nothing running
```

**Nothing in the orchestration layer is currently set up for Tars.** Using it means creating a
fresh Run from scratch, not plugging into existing state.

## Pros/cons for the synthesis agent (no decision made here)

### (a) Lane-B rendezvous (VM created · Hermes installed · smoke pass pings)

**Plain `SendMessage` + committed status file** (what PLAN.md's contract already specifies):
- Pros: zero new mechanism — matches ORCHESTRATION-POLICY.md §9 verbatim; already proven working
  right now (`ListAgents` shows 6 live addressable peer sessions); no Run/Task bootstrapping;
  durable fallback already designed (git commit to `status/lane-b.md`); message delivery is
  async/enqueued, no polling needed on either side.
- Cons: no structured status enum (must read/grep the status file or ask), no built-in
  worker-liveness registry (no equivalent of `worker-list` showing lane B is still alive) beyond
  `ListAgents`' `idle`/... state.

**Orca `orchestration` Task/Dispatch + `worker_done`/`check --wait`**:
- Pros: typed lifecycle (`worker_done` with `--outcome succeeded|failed`, `--files-modified`,
  `--report-path`), `check --wait --types worker_done,escalation,question --timeout-ms <n>` gives
  a native blocking-wait loop instead of a hand-rolled poll, `worker-list`/`worker-show` give a
  resource-accounting view across dispatches.
- Cons: lane B, per PLAN.md, is spawned as an unattended peer via `orca worktree create` +
  `orc-tab start -w ... orc-opus` — exactly the shape the orchestration guide classifies as
  **"full handoff"** by its own rule ("another worktree" → full handoff by default), for which it
  explicitly says not to use `task-create`/`dispatch --inject`/`check --wait`. Adopting it here
  means going against the bundled guide's own classification, and bootstrapping a Run from
  scratch (none exists today — see §7) as a second coordination channel alongside the
  already-designed status-file contract.

### (b) Cutover gate (destructive Slack-app cutover, human go/no-go)

**Plain inline stop-and-wait** (current PLAN.md design: "STOP at the cutover... Gaetan's explicit
'go' in this session"):
- Pros: the orchestrator session is already interactive with Gaetan (same session running lane
  A's browser work) — a plain conversational pause needs no tooling at all; evidence-before-go is
  already the stated discipline (workflow.md pref: "Evidence before assertions"); no dependency
  on any Orca state being bound.
- Cons: no structured, listable audit record of the gate itself (only chat transcript + whatever
  gets committed to the repo/report).

**Orca `gate-create`/`gate-resolve`**:
- Pros: purpose-built for exactly this — a named `--question`, optional `--options`, a durable
  `gate-list [--status pending]` you (or a probe) can query independently of scrollback, and a
  `gate-resolve --id <id> --resolution <text>` that leaves a typed record. Semantically the
  closest primitive Orca ships to "blocking decision point."
- Cons: **a gate blocks a Task** — there's no standalone gate, so using it means first modeling
  the cutover as an Orca Task (`task-create`) under a bound Run, adding the same
  bootstrap-from-nothing cost as (a); the gate only blocks Orca's own task state, it does not
  stop the orchestrator session's own turn from proceeding — the actual enforcement is still "the
  orchestrator session chooses not to act until it sees the resolution," same as the plain
  stop-and-wait, just with extra bookkeeping around it. The orchestrator session *is* Gaetan's
  session (same browser, same terminal) — there is no separate worker process a `gate-create`
  would need to block for this specific gate, since the human and the executor are in the same
  seat.

## Blockers

None — recon completed without needing anything destructive or unavailable.

## Open questions

- Whether the "orchestration experimental feature" the bundled guide says must be enabled in
  Settings > Experimental is actually toggled on: not confirmed via a Settings-UI probe (read-only
  recon didn't open the app UI), but indirect evidence says yes — `orchestration.contract.v1` and
  related capability flags are present in `orca status --json`, and `run-list`/`worker-list` both
  returned `ok:true` (a disabled feature would be expected to error, not silently succeed).
- Whether `--effort` on `worker-start`/`orca orchestration ...` maps to the same `low/high/xhigh`
  vocabulary ORCHESTRATION-POLICY.md §6 uses for the MODEL RULE — not tested (would require an
  actual spawn, out of scope for read-only recon).
- No `orc-sonnet` launcher exists on PATH — only `orc-opus` and `orc-fable`. If a future lane
  wants a Sonnet-tier sub-orchestrator via a fixed launcher (vs. per-agent `model: "sonnet"` on
  Task tool calls), one doesn't exist yet; unclear if that's intentional (Sonnet workers are
  meant to be plain subagents, not peer sessions) or a gap.
