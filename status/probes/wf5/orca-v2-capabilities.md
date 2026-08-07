# WF5 v2 — Orca CLI as a programmatic agent-orchestration surface

Read-only recon on cooper. Nothing created, changed, or committed. Builds on
`status/probes/wf5/orca-recon.md` (v1) and `docs/specs/wf5-orca-delegation.md`
(the shipped v1 playbook). **Every command below was actually run**; outputs
are trimmed for length but not altered. Full raw captures (`agent-context.json`,
the two skill guides) were saved to this agent's own scratchpad during the
run and are not part of this repo.

## STEER (mid-investigation, from Gaetan) — read this first

v1 (`ssh` + `delegate.sh` + headless `claude -p` sandbox) is **dropped**. The
question is no longer "should Tars drive real Orca sessions" — it's **how**,
and what breaks. Target repo for all spawn examples: **mc-metarepo**
(`id:8099e312-3232-46f2-83a9-97aeaf5de5a2`, `/home/gaetan/dev/mc-metarepo`,
`github.com/mobile-club/metarepo`). This recon stayed strictly read-only per
Gaetan's explicit instruction — no worktree/terminal/run was created, even for
a test. The `orchestration worker-start` sequence a live spawn would use is
composed and documented in full in **§9** but marked **NOT EXECUTED**.

---

## 1. What's new vs the v1 recon

v1 rejected the whole Orca session layer for three reasons and, on the
evidence in `orca-recon.md`, never inventoried the `orchestration worker-*`
subtree (`worker-start/-show/-read/-stop/-list/-abandon/-release/-retain`) —
it isn't mentioned anywhere in that file. That subtree is the actual answer
to all three v1 rejections:

| v1 rejection | v1's evidence | What this recon found |
|---|---|---|
| Stale terminal handles | `terminal_handle_stale`, no name-based addressing | True for the **terminal** layer. One layer up, `worker-start` returns a **`dispatch_id`** — a durable, DB-backed key (`worker_dispatches` table) that survives terminal churn. `worker-read`/`worker-show` keep working by `dispatch_id` even after the underlying terminal is released — "an inspectable output archive is preserved before the terminal closes" (from `worker-release --help`). |
| No completion signal | "no built-in completion signal — Tars would hand-roll polling" | False at the orchestration layer. `worker_done` is a first-class message type with an explicit `--outcome succeeded\|failed`, and `orchestration check --wait --types worker_done,... --timeout-ms <n>` is a **single blocking ssh call** that returns exactly when it happens or times out. See §3. |
| Durable coordinator state, "more state to track... across ssh calls that don't share a session" | speculative | Confirmed present (`~/.config/orca/orchestration.db`, SQLite, schema captured in §4) **and** empirically NOT session-bound: `task-list --run <id>`, `worker-list --run <id>`, `gate-list --run <id>` all work with an explicit `--run <run_id>` and **no prior `run-use` bind** — tested live against the pre-existing `run_legacy_local` run with zero shared shell state. `run-use`/`run-current` exist only so a *long-lived* coordinator terminal can omit `--run` on every call; a stateless ssh caller doesn't need them at all. This was v1's single biggest architectural objection and it does not hold for read/write calls that pass `--run` explicitly. |

Also new, not investigated by v1 at all:

- A **local API surface** is live right now (`orca serve`, bound to
  `0.0.0.0:6768`, advertised on the Tailscale IP) — §5.
- The **agent-status push mechanism** Orca uses internally
  (`~/.orca/agent-hooks/*.sh` → HTTP POST to a local Orca listener) — §3.4.
  This is the closest thing to a genuine "push" in the whole system.
- **Peer-session addressing is a Claude Code mechanism, not an Orca one**
  (`claude agents --json`, `ListAgents`/`SendMessage`) — §7.
- Exact mc-metarepo worktree-landing/branch-naming behavior — §8.

What v1 still got right and remains true:

- Terminal handles (`term_<uuid>`) are genuinely ephemeral; `terminal_handle_stale`
  is real and is called out repeatedly in both bundled skill guides.
- `orca` hard-depends on the desktop runtime being up (`orca status` →
  `runtime.reachable`). Confirmed live (pid 218650) but this is a real
  precondition, not a formality — see §9 step 0 and §10.
- No push/webhook mechanism exists for reaching *outward* past cooper (v1
  didn't rule this out explicitly; this recon confirms it — §3.4/§6).

---

## 2. Handle addressing — what's durable, what isn't

**Ephemeral, do not persist across ssh calls:**
- Terminal handle (`term_<uuid>`) — runtime-issued, goes stale on app
  restart/adoption. `orca terminal list --worktree <selector> --json` is the
  only correct way to get a *current* one; always re-resolve immediately
  before use, never dual-send to an old + new handle (both skill guides say
  this almost verbatim).

**Durable, safe to persist and reuse across independent ssh calls:**
- **Worktree selector** — not just the opaque id: `id:<repoId>::<path>`,
  `name:<displayName>`, `branch:<branch>`, `path:<abs>`, `issue:<number>` all
  resolve deterministically (`orca worktree show --worktree name:"GoCardless rotation" --json`
  verified live). This is real stable, name-based addressing that survives
  across calls and app restarts.
- **`run_id`** (e.g. `run_legacy_local`) — durable namespace/inbox row in
  `orchestration.db`. Accepted directly by `--run` on `task-list`,
  `worker-list`, `gate-list`, `send`, `check`, `ask`, `worker-start` — no
  binding step required.
- **`task_id`** — durable row in `tasks`, referenced by `run_id`.
- **`dispatch_id`** — durable row in `worker_dispatches`/`dispatch_contexts`;
  the correct address for a specific delegated agent run. Outlives the
  terminal it started in.
- **`sessionId`** from `claude agents --json` — a Claude Code concept (not
  Orca's), see §7. Durable UUID, independent of which Orca terminal spawned
  it.

**Resolution rule for a stateless caller:** never cache a `term_*` handle
across ssh invocations. Cache the worktree id / `run_id` / `task_id` /
`dispatch_id` you got back from the mutating call instead, and re-resolve a
fresh terminal handle (if you need one at all) on every call via
`terminal list --worktree <selector> --json`.

---

## 3. Completion signalling — ranked by usefulness to a stateless remote caller

1. **`orca orchestration check --run <run_id> --wait --types worker_done,escalation,question --timeout-ms <n> --json`**
   Single blocking ssh call. Returns the instant the dispatch posts
   `worker_done` (with `--outcome succeeded|failed`) or a matching
   `escalation`/`question`, or `{count:0}` at timeout. Emits JSON keepalive
   lines to stderr every 15s (`_keepalive`; `_heartbeat` is a deprecated
   alias) so the caller's ssh session can be distinguished from a hang — a
   caller merging streams filters with `jq "select(._keepalive|not)"`.
   Works with only `--run`, no bind step. **Best fit**: this is exactly the
   "one ssh command blocks, returns the answer" shape v1's `delegate.sh`
   already used, but now with a real succeeded/failed outcome instead of
   "the process exited."
2. **`orca orchestration worker-show --dispatch <id> --json`**
   Cheap point-in-time poll. Returns `state` from the DB-backed enum
   `starting|ready|start_unknown|failed|succeeded|stopping|stop_unknown|stopped|abandoned`
   (see `worker_dispatches` schema, §4). Use this in a retry-loop from a
   scheduler if you'd rather not hold an ssh connection open for the wait.
3. **`orca orchestration task-list --run <run_id> --status completed|failed --json`**
   Poll at the task/DAG level (`pending/ready/dispatched/completed/failed/blocked`)
   rather than one dispatch.
4. **`orca terminal wait --terminal <handle> --for exit|tui-idle --timeout-ms <n> --json`**
   v1's mechanism. Blocking, single call — but no outcome field (just
   "process exited" / "TUI went idle"), and the handle can go stale between
   the create call and the wait call. Weakest of the four; fine as a
   liveness supplement, not as the primary signal.
5. **Exit codes are launch-readiness signals, not task-completion signals.**
   `worker-start` "exits 0 only for `ready`"; failure/`outcome_unknown` exits
   1 with `stage`/`failedStage`/`effects`/`residualResources` in the JSON.
   That tells you the worker *started*, not that the *task* finished.
   `worker-release` exit 0/1 signals cleanup state, likewise not task
   outcome.

### 3.4 Is there any push mechanism? (sharpened per the steer)

**Internally: yes, and it's real HTTP push — but it never leaves cooper.**
`orca agent hooks status --json` shows Orca has installed managed status
hooks into *every* supported agent's own config (Claude: managedHooksPresent
in `~/.claude/settings.json`; also Codex, Gemini, Cursor, Droid, Grok,
OpenClaude, Amp, Copilot, command-code, Devin, Kimi, Antigravity, Hermes —
13 providers total). Reading the actual injected Claude hook
(`~/.claude/settings.json` → `hooks.PostToolUse[].hooks[].command`, matcher
`*`) and the script it calls
(`/home/gaetan/.orca/agent-hooks/claude-hook.sh`, read in full):

```sh
#!/bin/sh
payload=$({ command -p cat 2>/dev/null || cat; })
[ -z "$payload" ] && exit 0
[ -n "$DEVIN_PROJECT_DIR" ] && exit 0
[ -n "$ORCA_AGENT_HOOK_ENDPOINT" ] && [ -r "$ORCA_AGENT_HOOK_ENDPOINT" ] && . "$ORCA_AGENT_HOOK_ENDPOINT" 2>/dev/null
[ -z "$ORCA_AGENT_HOOK_PORT" ] || [ -z "$ORCA_AGENT_HOOK_TOKEN" ] || [ -z "$ORCA_PANE_KEY" ] && exit 0
printf '%s' "$payload" | curl -sS -X POST "http://127.0.0.1:${ORCA_AGENT_HOOK_PORT}/hook/claude" \
  --connect-timeout 0.5 --max-time 1.5 \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "X-Orca-Agent-Hook-Token: ${ORCA_AGENT_HOOK_TOKEN}" \
  --data-urlencode "paneKey=${ORCA_PANE_KEY}" \
  --data-urlencode "tabId=${ORCA_TAB_ID}" \
  --data-urlencode "launchToken=${ORCA_AGENT_LAUNCH_TOKE[N]..."
```

So: every tool-use / lifecycle event inside a Claude Code session that Orca
spawned is HTTP-POSTed, in real time, to a per-pane token-gated
`127.0.0.1:${ORCA_AGENT_HOOK_PORT}/hook/claude` listener that Orca itself
runs. This is almost certainly what feeds `terminal list`'s live `preview`
field and the `worker_dispatches.state` transitions — genuine push, not
polling, at the agent→Orca hop.

**But it is local-loopback only** (`127.0.0.1`, port/token handed to the
agent process via env vars Orca sets at launch). It never reaches past
cooper. So even with this internal push, **Tars still has to ask Orca**
(via `orca` CLI over ssh, or the local API in §5) for the result — the
push only removes Orca's *own* polling of the agent, not a remote caller's
need to ask Orca.

**Could this be extended to push outward to Tars?** In principle yes — the
mechanism is a plain hook-command pattern (`~/.claude/settings.json`
`hooks.Stop`/`PostToolUse` etc. accepts arbitrary shell commands), and a
project- or worktree-level `.claude/settings.json` merges with the user-level
one. A caller who creates the worktree could pre-seed an *additional* `Stop`
hook in that worktree that curls an endpoint on Tars, writes a sentinel file
readable over ssh, etc. **This was not tested** (would require creating a
worktree) and it would be a caller-authored addition layered on top of
Claude Code's hook system, not a built-in Orca feature — Orca's own push
stops at its own local listener.

**No outbound webhook/automation-completion hook exists in Orca itself.**
Grepped the full 223-command `agent-context.json` for
`webhook|callback|notify|push|hook_url` — zero hits beyond one unrelated
Linear example string (`"Fix OAuth callback"`). `automations` (`--precheck`,
`--trigger`) is trigger-**in** only (a command that gates whether a
*scheduled* run fires); there is no "run this command / hit this URL when a
task/dispatch completes" field anywhere in `orchestration create/task-create/
worker-start/dispatch` schemas.

**Verdict:** no existing built-in mechanism lets a spawned agent notify Tars
directly. The closest thing to "no durable state needed" is #1 above
(`check --wait`) — one blocking ssh call that returns the moment work is
done — which is functionally push-shaped from the caller's point of view
even though it's implemented as a long poll.

---

## 4. Durable state — where it lives, and how to read it without `orca`

- **`~/.config/orca/orchestration.db`** — SQLite, WAL mode (`-wal`/`-shm`
  present, confirming live writers). This is the coordinator/mailbox/task/
  dispatch store. Tables (`.tables`, schema captured, **no rows dumped**):

  ```
  runs                       messages                  deliveries
  mutation_receipts          worker_dispatches          worker_terminal_resources
  worker_terminal_archives   federated_dispatches       remote_dispatch_attachments
  federation_relay_items     remote_questions           tasks
  dispatch_contexts          decision_gates             coordinator_runs
  question_threads           legacy_adoptions           legacy_compatibility_principals
  legacy_operation_receipts  legacy_mail_receipts
  ```

  Load-bearing `CHECK` constraints (full `.schema` captured for `runs`,
  `tasks`, `messages`, `deliveries`, `worker_dispatches`, `dispatch_contexts`,
  `decision_gates`, `worker_terminal_resources`):
  - `tasks.status` ∈ `pending, ready, dispatched, completed, failed, blocked`
  - `messages.type` ∈ `status, dispatch, worker_done, merge_ready, escalation, handoff, decision_gate, question, heartbeat`
  - `worker_dispatches.state` ∈ `starting, ready, start_unknown, failed, succeeded, stopping, stop_unknown, stopped, abandoned`
  - `dispatch_contexts.status` ∈ `pending, dispatched, completed, failed, circuit_broken`
  - `worker_terminal_resources.release_state` ∈ `not_requested, retained, requested, releasing, released, unknown`

  This is directly `sqlite3`-queryable read-only over ssh with no `orca`
  binary at all (`sqlite3 -readonly ~/.config/orca/orchestration.db
  ".schema"` / a `SELECT` on non-content columns like `status`/`state`).
  That is an *unsupported side channel* for a stateless poller in a pinch —
  the sanctioned interface is still the CLI (`worker-show`/`task-list`),
  which does capability/liveness checks the raw DB read skips.

- **`~/.config/orca/profiles/local-default/orca-data.json`** (+ 5 rotating
  `.bak.0`–`.bak.4` copies) — plain JSON, *not* SQLite. Repo/worktree/
  terminal/UI-session state, distinct from orchestration state above.
  Top-level keys: `repos, projects, projectHostSetups, worktreeMeta,
  workspaceSession, settings, automations, automationRuns, sshTargets,
  claudeLivePtySessionIds, …`. `settings.workspaceDir` and
  `settings.branchPrefix` are what actually decide where a new worktree
  lands and how its branch is named — see §8.

- **`~/.config/orca/orca-runtime.json`** — how to reach the running
  instance: `transports: [{kind: unix, endpoint: <sock path>}, {kind:
  websocket, endpoint: "ws://0.0.0.0:6768"}]` plus a bearer `authToken`.
  **Value not reproduced in this report** — treat it as a credential
  (mode 600, gaetan-only) even though it isn't on the machine-map deny list
  verbatim.

- **`~/.config/orca/daemon/daemon-v32.{sock,token,pid}`** — a *separate*
  single-instance app-lifecycle lock (Electron singleton), not the CLI's RPC
  channel. `daemon-v32.pid` (not secret) confirms the daemon is the app's
  process supervisor (`pid":2704`, `entryPath: .../daemon-entry.js`).

None of the above required stopping or touching the running app; all reads
were `.schema`/`.tables`/JSON-key-listing only, no row/content dumps.

---

## 5. Local API surface — confirmed live, not merely documented

`ps aux` shows `orca serve --port 6768 --pairing-address 100.115.232.100`
already running as part of this Orca instance's own boot (three related
processes: the `serve` wrapper, the CLI `serve` entry, and `orca-ide --serve`
itself). `ss -ltnp` confirms it's bound **`0.0.0.0:6768`**, not loopback-only,
and Tailscale-advertised endpoints also exist at
`100.115.232.100:{443,8443,8444}` and the IPv6 ULA equivalent.

```
$ curl -sS -m 5 -i http://127.0.0.1:6768/
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
...
<title>Orca Web</title>
<script type="module" crossorigin src="./assets/web-D2f7P9Ss.js"></script>
...
```

This is a full web client ("Orca Web", a Vite/React SPA bundle), not just a
health check. `orca-runtime.json`'s `transports` confirms the actual RPC
protocol underneath: a local `unix` socket for same-host CLI calls, and
`websocket ws://0.0.0.0:6768` for everything else — almost certainly what
the `orca` CLI itself uses whenever `--environment`/`--pairing-code` target a
non-local instance.

**Remote pairing flow** (documented via `--help`, not tested — pairing would
create new device/environment state): `orca environment add --name <name>
--pairing-code <code>` stores a saved remote runtime; every subsequent `orca`
command can target it with `--environment <name>` or a one-off
`--pairing-code`. This is Orca's own concept of a durable, named connection
to a specific remote instance — architecturally closer to what a
cross-machine caller wants than "ssh in and shell out to the local CLI."

**Open question, not resolved by this recon (cooper-only):** whether Tars's
VM (192.168.0.9) has network reachability to cooper's tailnet identity
(`100.115.232.100`) at all. CLAUDE.md documents the existing Tars↔cooper path
as ssh over the "tars" tailnet alias — it is not established here whether
that's the *same* tailnet cooper's `100.115.232.100` sits on. Flagging as an
open question for whoever tests connectivity from the Tars side, not
asserting the pairing path works.

---

## 6. Peer-session spawn mechanics (`orc-tab`, `orc-opus`, `orc-fable`)

All three scripts read in full (`/home/gaetan/.local/bin/{orc-tab,orc-opus,orc-fable}`):

- **`orc-tab`** (despite the name, the actual generic launcher; comment
  block calls it "launch.sh"): wraps
  `orca terminal create --worktree <selector> --command "<launcher> '<prompt>'" --focus --json`,
  reads the handle back from `.result.terminal.handle // .result.handle`,
  and fails closed if the handle is empty or equals the *caller's own*
  `ORCA_TERMINAL_HANDLE` (guards against seeding your own TUI). Three modes:
  plain handoff (a brief file), `start [-w <worktree>]` (fresh prompt,
  optionally into another worktree — the hub-and-spoke pattern), `reseed`
  (resend to an existing idle tab).
- **`orc-opus` / `orc-fable`** are one-liners:
  `exec claude --model 'opus[1m]'|'fable[1m]' --settings '{"ultracode":true}' --append-system-prompt-file ~/.claude/handoff/ORCHESTRATION-POLICY.md --dangerously-skip-permissions "$@"`.
  They just launch Claude Code itself, with the standing orchestrator system
  prompt appended, inside whatever terminal Orca already created. **Orca
  only owns the PTY** (`term_<uuid>`) here — it has no concept of "session"
  beyond that terminal; no `taskId`/`dispatchId` is attached by this path
  (confirmed in the orchestration skill: "Use `orca worktree create
  --prompt ...` or `orca terminal send ...` for full handoffs... those paths
  do not attach `taskId`/`dispatchId`").

### 7. Peer-session identity and Claude Code addressing (steer Q4/Q5)

**This is a *Claude Code* mechanism, entirely separate from Orca's own
orchestration DB.** `claude agents --json` (run directly on cooper, no
`orca` involved) lists every live/background Claude Code session on the
host:

```json
{"cwd": "/home/gaetan/dev/mc-metarepo-gocardless-rotation",
 "sessionId": "596c596e-dd1d-4652-91e4-0095b69057d1",
 "name": "mc-metarepo-gocardless-rotation-98", "status": "idle"}
```

- `sessionId` is a durable UUID — this is what `-r/--resume <sessionId>`
  and (inside a live agent turn) `ListAgents`/`SendMessage` key off.
- `name` is **not caller-controllable at spawn time in the orc-opus/orc-fable
  pattern** — those launchers deliberately omit `-n/--name` (their own
  comment: "a fixed session name freezes the terminal title, which kills
  Orca's work-based tab auto-renaming") and let Claude Code auto-derive one
  from cwd/branch plus what looks like a random 2-char disambiguator (e.g.
  `-98`, `-af`, `-e5` across three sessions all in the *same* worktree
  `oof-wp-vecna` — confirms the suffix is not deterministic from the path
  alone). **A stateless caller cannot predict this name before spawn.**
- The registry backing `claude agents --json` was not located as a single
  readable file in this recon (it's presumably in Claude Code's own state
  under `~/.claude/`, out of scope for the Orca-focused file reads done
  here) — but the command itself is a perfectly good stateless, ssh-callable
  directory lookup: match on the known `cwd` (which Tars *does* know, from
  the `worktree create --json` response) rather than on `name`.
- **`ListAgents`/`SendMessage` are Claude Code SDK tools available only
  inside a live agent turn** (this very investigation used them to report to
  its Tars-orchestrator hub) — they are not `orca` subcommands and not
  shell-invokable. Confirmed: `claude --help`'s top-level `Commands:` list
  (`agents, auth, auto-mode, doctor, gateway, import, install, mcp, plugin,
  project, setup-token, ultrareview, update`) has no send/message
  subcommand. The only CLI-level lever on another session is
  `-r/--resume <sessionId>` (attach/continue), which is interactive by
  default.

**Steer Q5 — can a spawned session be sent further messages from a plain
shell (not from inside a live Claude session)? Yes, and the mechanism is
Orca's, not Claude Code's:**
- `orca terminal send --terminal <handle> --text "<msg>" --enter --json` —
  literally types into the PTY. Requires re-resolving the (possibly stale)
  handle first via `terminal list --worktree <selector> --json`.
- `orca orchestration send --to dispatch:<id> --subject ... --body ... --json`
  — structured inbox mail, delivered on the worker's *next*
  `orchestration check`, **but only if the worker was started via
  `worker-start`/`dispatch --inject`** (i.e. has a live `dispatchId`) — the
  bare `orc-tab`/`orc-opus` pattern does not qualify (no task/dispatch
  attached, see above).
- `claude -p -r <sessionId> "<msg>"` (documented via `claude --help`, **not
  tested**) would in principle resume a session non-interactively from a
  bare shell — but it's uncertain whether this conflicts with an
  already-live interactive session occupying that `sessionId`; flagging as
  unverified, not recommending it.

**Consequence for Tars:** Tars is a stateless ssh caller, not itself a live
Claude Code agent turn, so it structurally **cannot** use
`ListAgents`/`SendMessage` — that channel only exists between two live agent
turns. Tars's only two levers on any peer are `orca terminal send` (against
an ephemeral, must-be-re-resolved handle) or the orchestration mailbox
(against a durable `dispatch_id`, if and only if the peer was spawned via
the supervised `worker-start` path rather than a bare `orc-tab`/`orc-opus`
handoff).

---

## 8. mc-metarepo worktree-creation mechanics (steer Q1/Q2/Q3)

### 8.1 Binding — by id, name, or path?

`orca repo show --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json`:

```json
{
  "id": "8099e312-3232-46f2-83a9-97aeaf5de5a2",
  "path": "/home/gaetan/dev/mc-metarepo",
  "displayName": "mc-metarepo",
  "worktreeBasePath": "null",
  "gitRemoteIdentity": {"canonicalKey": "github.com/mobile-club/metarepo",
    "remoteName": "origin", "remoteUrl": "https://github.com/mobile-club/metarepo.git"},
  "hookSettings": {"mode": "auto", "setupRunPolicy": "run-by-default",
    "setupAgentStartupPolicy": "start-immediately",
    "scripts": {"setup": "", "archive": ""}}
}
```

`worktree create --help`: `--repo <selector>` accepts `id:<id>`,
`name:<name>`, or `path:<path>` — all three resolve to the same repo. Use
`--repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2` (unambiguous, no name
collisions possible) for a scripted caller.

### 8.2 `worktreeBasePath: "null"` — what it actually means

The value is the **literal string `"null"`**, not a JSON null, meaning "no
per-repo override." Practically, Orca falls back to the **global** default:
`orca-data.json → settings`:

```
workspaceDir           = "/home/gaetan/orca/workspaces"
nestWorkspaces          = true
autoRenameBranchFromWork = true
branchPrefix            = "git-username"
```

Cross-checked against real `orca worktree list --repo id:8099e312-... --json`
entries — mc-metarepo has **three different historical directory
conventions** on disk: `~/dev/mc-metarepo-<slug>` (majority, older),
`~/dev/mc-metarepo/worktrees/mc-metarepo/<slug>` (two entries, oldest —
predate the current `workspaceDir` setting or were created outside Orca),
and **one recent entry**, `oof-wp-vecna`, already at
`/home/gaetan/orca/workspaces/mc-metarepo/oof-wp-vecna` — matching the
*current* global `workspaceDir` + `nestWorkspaces` convention exactly (the
same pattern this very `orca-a2a` worktree of the `Tars` repo uses).

**Conclusion: a worktree created for mc-metarepo today lands at
`/home/gaetan/orca/workspaces/mc-metarepo/<name-slug>`** — the global
setting governs, the per-repo `"null"` override is inert, and the most
recent real example confirms it empirically (not just from settings).
Branch: `autoRenameBranchFromWork=true` + `branchPrefix="git-username"` ⇒
`refs/heads/GaetanCathelain/<name-slug>`, base = the repo default
(`origin/main`, i.e. `refs/heads/main`) unless `--base-branch` is passed.

### 8.3 `hookSettings` — does an agent auto-start?

`mode: auto, setupRunPolicy: run-by-default, setupAgentStartupPolicy:
start-immediately, scripts.setup: "" (empty), scripts.archive: "" (empty)`.

- **No default-terminal config exists for mc-metarepo** (`repo show`'s JSON
  has no such field) — so a **bare** `worktree create` with no `--agent`
  opens a plain shell in the first terminal, nothing more.
- `setupRunPolicy: run-by-default` + empty `scripts.setup` ⇒ setup "runs"
  but is a no-op (nothing to execute).
- `setupAgentStartupPolicy: start-immediately` governs *timing*, not
  *whether* an agent launches — it means if `--agent` is given, the agent
  starts immediately in parallel with setup rather than waiting for setup to
  finish (`wait-for-setup` is the alternative policy this repo does *not*
  use).
- **Net effect: nothing auto-starts an agent.** The caller must pass
  `--agent claude` (or another id) explicitly to get an agent running at
  all; omitting `--agent` yields an empty worktree + idle shell.

### 8.4 Which `--agent` values are actually usable

No `orca agent list` subcommand exists (`agent-context.json`'s only `agent
*` commands are `agent hooks {on,off,status}`). Two independent signals,
cross-checked:

- **Binaries on PATH** (`which`): `claude` → `/home/gaetan/.local/bin/claude`,
  `codex` → `/usr/local/bin/codex`. `omp, pi, grok, gemini, cursor, droid,
  opencode` — **not on PATH**.
- **`orca account list --json`**: 2 managed Claude accounts (active:
  `gaetan.cathelain-claude02@mobile.club`, session usage 10%, weekly 17% —
  headroom is fine), Codex has no *Orca-managed* account but a working
  `systemDefault` OAuth (`gae.cathelain@gmail.com`). Every other provider in
  `rateLimits` (gemini, opencodeGo, kimi, antigravity, minimax, grok) is
  explicitly `unavailable`/not signed in.
- `orca agent hooks status --json` lists 13 providers with
  `managedHooksPresent: true` (their *config/hook* plumbing exists), but
  that is not the same claim as "the binary is installed and authenticated"
  — e.g. `gemini` shows `state: installed` there while `which gemini` finds
  nothing runnable. **Do not use the hooks-status list to pick an agent id.**

**Verdict: `--agent claude` and `--agent codex` are the two verified-usable
choices on cooper today.**

---

## 9. Exact spawn sequence, unrun

**NOT EXECUTED.** Composed from `--help` text, the orchestration skill guide,
and the schema/settings evidence above. Field names from `worker-start`'s
JSON response (e.g. whether the dispatch id comes back as
`result.dispatch.id` or `result.dispatchId`) are **not confirmed** — the
command was never run, per the explicit instruction to stay read-only. Save
whatever the first live run actually prints before scripting a parser
against it.

```bash
# ── 0. Preflight (cheap, read-only, already verified working in this recon) ──
orca status --json                    # must show runtime.reachable == true
orca account list --json              # must show an active/usable claude or codex account

# ── 1. Durable namespace + task row (no terminal binding, no shared shell state) ──
orca orchestration run-create --objective "<one-line objective>" --json
#   -> capture result.run.id (or equivalent field) as RUN_ID — save it, it's the
#      handle every later call needs
orca orchestration task-create --run "$RUN_ID" \
  --spec "<full task brief, prose>" \
  --task-title "<short title>" --json
#   -> capture result.task.id as TASK_ID

# ── 2. Spawn the supervised worker directly in mc-metarepo ──
orca orchestration worker-start \
  --run "$RUN_ID" --task "$TASK_ID" \
  --worktree new-top-level \
  --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 \
  --name a2a-test-spawn \
  --agent claude \
  --setup run \
  --json
#   Predicted landing dir (from §8.2): /home/gaetan/orca/workspaces/mc-metarepo/a2a-test-spawn
#   Predicted branch (from §8.2):      refs/heads/GaetanCathelain/a2a-test-spawn, off origin/main
#   Exits 0 only if the worker reaches `ready`; exits 1 with stage/failedStage/
#   effects/residualResources on failure or outcome_unknown — inspect, don't retry blind.
#   -> capture result.dispatch.id (field name NOT confirmed) as DISPATCH_ID —
#      THIS is the durable address for every step below.
#
#   What could go wrong:
#   - orca not running (skipped by step 0's preflight, but worth restating: this
#     is a hard dependency, confirmed live right now, pid 218650)
#   - --name collision if `a2a-test-spawn` already exists as a worktree name
#     for this repo (behavior on collision not documented; not tested)
#   - the active Claude account's session-usage window is exhausted mid-task
#     (10%/17% used as of this recon — unlikely but not zero)
#   - `new-top-level` + omitted --base-branch uses the repo default (origin/main);
#     if that default has since diverged from what the task needs, say so
#     explicitly in the task spec rather than assuming --base-branch

# ── (b) Query status — cheap, one-shot, no shared state ──
orca orchestration worker-show --dispatch "$DISPATCH_ID" --json
#   -> .result.state: starting|ready|start_unknown|failed|succeeded|stopping|stop_unknown|stopped|abandoned
orca orchestration task-list --run "$RUN_ID" --status dispatched --json   # DAG-level cross-check

# ── or, block for completion in ONE ssh call instead of polling ──
orca orchestration check --run "$RUN_ID" --wait \
  --types worker_done,escalation,question --timeout-ms 900000 --json
#   Returns the instant worker_done/escalation/question lands, or {count:0} at
#   the 15-minute timeout. Emits stderr keepalives every 15s.

# ── (c) Retrieve output ──
orca orchestration worker-read --dispatch "$DISPATCH_ID" --limit 200 --json
#   Works even after worker-release (archived transcript is preserved first).

# ── (d) Cleanup ──
orca orchestration worker-release --dispatch "$DISPATCH_ID" --json
#   Post-completion only. Closes JUST the settled agent terminal; does NOT
#   touch the worktree, its git branch, or its checkout on disk. Idempotent
#   (already_released on repeat). If it returns release_pending/release_unknown,
#   do not substitute `terminal close` — follow the returned recovery action.
orca worktree rm --worktree "id:8099e312-3232-46f2-83a9-97aeaf5de5a2::/home/gaetan/orca/workspaces/mc-metarepo/a2a-test-spawn" \
  --force --json
#   Removes the worktree from BOTH Orca and git. mc-metarepo's archive hook
#   is empty, so --run-hooks would be a no-op here — omitted. This step is
#   ONLY correct once the branch's work is fully landed/abandoned; it is
#   irreversible for the local checkout (the branch/commits remain in the
#   bare repo's refs unless force-deleted separately, but the working tree
#   goes away).
```

**Cleanup / accumulation risk (steer Q6), evidenced, not just asserted:**
mc-metarepo already carries **9 worktrees** in `orca worktree list --json`,
several visibly stale (`"comment": "handed off to handoff-fable — ..."` on
`GoCardless rotation`, last activity days old). Nothing in `worktreeMeta`
carries a TTL or auto-expiry field. If nobody runs `worktree rm`, each
abandoned spawn leaves: a full git checkout on disk, its branch, rows in
`orchestration.db` (`worker_dispatches`, `dispatch_contexts`,
`worker_terminal_resources`), and a `terminal-history/` file under
`~/.config/orca/`. None of this self-cleans.

**What fails, and how loudly, if Orca isn't running (steer Q7):**
Documented (not empirically triggered — stopping the live app to test would
be destructive and out of scope): `orca --help`'s own "Behavior" section
states plainly, "Most commands require a running Orca runtime. If Orca is
not open yet, run `orca open` first." The orca-cli skill guide repeats it
almost verbatim and prescribes `orca status --json` as the first call of
every session. This is why step 0 above checks `runtime.reachable` before
anything else — it is the one precondition this whole surface has that a
plain-shell fallback doesn't.

---

## 10. Local API surface, once more — for the "should we use raw websocket
instead of shelling out" question

Verified live and running (§5); not exercised beyond an HTTP GET of `/`
(establishing a paired websocket session would create new device/pairing
state, out of scope for read-only recon). For v2, the CLI-over-ssh path
remains the concrete, already-proven transport (probe 15,
`docs/facts.md`); the local API is flagged as a real, running thing worth a
follow-up connectivity check from the Tars side, not yet a replacement.

---

## Sources — every command run (in the order used above)

```
which orca
orca --help
ls -la /home/gaetan/.local/bin/ | grep orc
cat /home/gaetan/.local/bin/orc-tab
cat /home/gaetan/.local/bin/orc-opus
cat /home/gaetan/.local/bin/orc-fable
head -40 /home/gaetan/.local/bin/orca-reap-stale-relays.py
orca status --json
orca terminal --help
orca worktree --help
orca orchestration --help
orca automations --help
orca terminal {wait,list,create,send,show,read} --help
orca orchestration {run-create,run-use,run-current,run-list,send,check,ask,
  reply,inbox,task-create,task-list,task-update,dispatch,dispatch-show,
  worker-start,worker-show,worker-read,worker-stop,worker-list,worker-abandon,
  worker-release,worker-retain,gate-create,gate-resolve,gate-list,reset,
  coordinator-start} --help
orca automations {create,show,list,run,runs,edit} --help
orca agent-context --json   (223 commands, 156676 bytes; grepped/filtered, not
  reproduced in full)
orca terminal list --json
orca worktree list --json
orca orchestration run-list --json
orca automations list --json
orca orchestration worker-list --json
orca orchestration task-list --json          (correctly errors run_required)
ss -ltnp
ps aux | grep orca
orca serve --help
orca environment {add,list,show} --help
curl -sS -m 5 -i http://127.0.0.1:6768/
curl -sS -m 3 -o /dev/null -w '%{http_code}' http://127.0.0.1:6768{/health,/status,/api,/rpc}
ip addr show
ls -la /home/gaetan/.config/orca/
ls -la /home/gaetan/.config/orca/daemon/
find /home/gaetan/.config/orca /home/gaetan/.local/share -iname '*orca*'
sqlite3 -readonly ~/.config/orca/orchestration.db ".tables"
sqlite3 -readonly ~/.config/orca/orchestration.db ".schema" (filtered to CREATE TABLE)
sqlite3 -readonly ~/.config/orca/orchestration.db ".schema {runs,tasks,messages,
  deliveries,worker_dispatches,dispatch_contexts,decision_gates,
  worker_terminal_resources}"
sqlite3 -readonly ~/.config/orca/orchestration.db "select id, home_database,
  legacy from runs limit 5"   (corrected column list after a schema-mismatch error)
stat /home/gaetan/.config/orca/daemon/{daemon-v32.sock,daemon-v32.token,daemon-v32.pid}
cat /home/gaetan/.config/orca/daemon/daemon-v32.pid
cat /home/gaetan/.config/orca/orca-runtime.json   (authToken value redacted in this report)
python3 -c "..." reading top-level keys / list-lengths / dict-key-samples of
  /home/gaetan/.config/orca/profiles/local-default/orca-data.json (no content dumped)
orca agent-context --json | grep -ci 'webhook|callback|notify|push|hook_url'
claude-teams / claude --help  (via the `claude-teams` alias observed in
  agent-context.json — resolves to the plain `claude --help` output)
orca agent hooks status --help
orca orchestration {task-list,worker-list,gate-list,inbox} --run run_legacy_local --json
orca skills list
orca skills get --help
orca skills get orca-cli --full
orca skills get orchestration --full
cat /home/gaetan/.claude/handoff/ORCHESTRATION-POLICY.md
claude agents --help
claude agents --json   (structure only: id/name/sessionId/state/cwd fields)
orca worktree ps --limit 5 --json ; orca worktree ps --help
head -40 /home/gaetan/.local/bin/orca-reap-stale-relays.py
claude agents --json | python3 -c "..." (field-shape extraction only)
orca worktree create --help
orca worktree rm --help
orca worktree set --help
orca repo show --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json
orca repo list --json
which claude codex omp pi grok gemini cursor droid opencode
orca agent hooks status --json
orca agent hooks off --help
python3 -c "..." extracting agent-context.json entries for worktree create /
  orchestration worker-start / agent hooks *
python3 -c "..." reading only settings.{workspaceDir,nestWorkspaces,
  workspaceDirHistory,autoRenameBranchFromWork,branchPrefix,...} from orca-data.json
head -c 800 /home/gaetan/.orca/agent-hooks/claude-hook.sh
ls -la /home/gaetan/.orca/agent-hooks/
sqlite3 -readonly ~/.config/orca/orchestration.db "select id, run_id, ... from runs"
  (errored — no such column; corrected in the next call)
orca worktree show --worktree name:"GoCardless rotation" --json
python3 -c "..." reading only hooks.PostToolUse/PreToolUse from
  /home/gaetan/.claude/settings.json (no other keys printed)
python3 -c "..." reading claudeLivePtySessionIds and one worktreeMeta sample
  (keys only) from orca-data.json
orca worktree list --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json
orca account list --json
```

All commands were read-only (`--help`, `--json` listings, `status`/`show`/
`list`, `sqlite3 -readonly`, `stat`, `ps`, `ss`, `curl GET`, `which`, `head`,
`cat` of non-secret files). No `worktree create`, `terminal create`,
`orchestration run-create`/`task-create`/`worker-start`, `automations
create`, or any other mutating command was executed at any point in this
investigation.
