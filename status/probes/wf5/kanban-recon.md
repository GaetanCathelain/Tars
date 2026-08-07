# WF5 Kanban recon — read-only findings

Date: 2026-08-07. Hermes source: `~/.hermes/hermes-agent/` (via `readlink -f ~/.local/bin/hermes` → `/home/gaetan/.local/bin/hermes`, editable install `hermes_agent 0.20.0`). No config or state changed.

## TL;DR

- Kanban **infrastructure is already live and unlocked** on the VM: `~/.hermes/kanban.db` exists (empty, `default` board only), the dispatcher is embedded in the gateway and enabled by default (`kanban.dispatch_in_gateway: true`), and the `/kanban` **slash command** works today in Slack with **zero config change** — it's registered unconditionally in the gateway's command table, not gated by `_check_kanban_mode`.
- What's actually gated ("registered but disabled" per the task) is only the **LLM-facing tool-call surface** — `kanban_show/list/create/complete/block/...` as model tools Tars can invoke autonomously mid-conversation. That gate cannot be satisfied by editing `platform_toolsets` (see below); it only opens two ways: (a) the process is a dispatcher-spawned worker (`HERMES_KANBAN_TASK` env), or (b) the process's top-level `toolsets:` config list contains `"kanban"` (CLI-invocation path only, not gateway sessions).
- No gateway restart needed for any of this — `kanban.*` config and `toolsets:`/`platform_toolsets:` are read fresh via `load_config()` (mtime-cached) and `check_fn` results are TTL-cached ~30s, matching the documented live-reload window.

## 1. What the two check functions test

Both live in `tools/kanban_tools.py`.

```python
def _check_kanban_mode() -> bool:
    if _is_delegated_child_context():
        return False
    if os.environ.get("HERMES_KANBAN_TASK") and _is_dispatcher_owned_worker():
        return True
    return _profile_has_kanban_toolset()

def _check_kanban_orchestrator_mode() -> bool:
    if _is_delegated_child_context():
        return False
    if os.environ.get("HERMES_KANBAN_TASK") and _is_dispatcher_owned_worker():
        return False
    return _profile_has_kanban_toolset()
```

- `_profile_has_kanban_toolset()` → `load_config()`, then `"kanban" in cfg.get("toolsets", [])` — the **top-level** `toolsets:` key (default `["hermes-cli"]`, confirmed live via `hermes config get toolsets`), not `platform_toolsets`.
- `_is_delegated_child_context()` / `_is_dispatcher_owned_worker()` (from `agent/delegation_context.py`) distinguish: a `delegate_task` subagent (always denied — can't touch the board), a genuine dispatcher-spawned worker (`HERMES_KANBAN_TASK` set, owned), vs. a cron job that merely inherited a stale `HERMES_KANBAN_TASK` from its parent worker process (not owned, denied).
- Net effect: `_check_kanban_mode` gates the 12 lifecycle/worker tools (`kanban_show`, `kanban_complete`, `kanban_block`, `kanban_heartbeat`, `kanban_comment`, `kanban_create`, `kanban_link`, `kanban_attach*`); `_check_kanban_orchestrator_mode` gates the 2 board-routing tools (`kanban_list`, `kanban_unblock`) — deliberately hidden from task workers, only available to whatever profile has the `kanban` toolset without being a scoped worker.

## 2. What check_fn gating actually controls — and its limit

`check_fn` only decides, for tool names **already present in the candidate set**, whether they survive into the model schema (`tools/registry.py: get_definitions`). The candidate set itself (`tools_to_include`) is built in `model_tools.py: _compute_tool_definitions()` purely from `enabled_toolsets` (the caller-supplied toolset name list) via `resolve_toolset()` — **`kanban` never enters that set unless it's already a member of `enabled_toolsets`**, at which point `check_fn` is almost redundant (belt-and-suspenders).

Two independent code paths build `enabled_toolsets`, and they behave differently for kanban:

- **CLI (`hermes chat` / `hermes -p <profile> chat`)**: `enabled_toolsets` = the top-level `toolsets:` config list directly. Adding `"kanban"` there makes a CLI-invoked session eligible (subject to `_check_kanban_mode`'s cache).
- **Gateway sessions (Slack, Telegram, etc. — i.e. Tars's live chat)**: `enabled_toolsets = sorted(_get_platform_tools(user_config, platform_key))` (`gateway/run.py:19725` and `:24712`, both the foreground-turn and background-task paths). `_get_platform_tools()` (`hermes_cli/tools_config.py`) only ever returns names from `CONFIGURABLE_TOOLSETS` — a fixed registry used for the `hermes tools` checklist. **`kanban` is explicitly excluded from that registry** (comment in `tools_config.py:290-291`: *"Non-configurable toolsets that `_get_platform_tools` resolves at read time — `kanban` and other check_fn-gated toolsets"*). So **adding `kanban` to `platform_toolsets.slack` in `config.yaml` has no effect** — it's silently dropped by `_get_platform_tools`, never reaching `enabled_toolsets` for the gateway session.
- **The one exception**: `_compute_tool_definitions()` force-appends `"kanban"` to `effective_enabled_toolsets` whenever `HERMES_KANBAN_TASK` is set on a dispatcher-owned worker process — this is how a spawned worker (a *separate* `hermes -p <assignee> chat -q` subprocess the dispatcher launches per ready task, not the live Slack session) gets the lifecycle tools regardless of its profile's normal toolset config.

**Consequence**: there is no config key that gives the live interactive Tars-in-Slack session the `kanban_*` model tools. The board is reachable from that session today only via the terminal tool shelling out to `hermes kanban ...`, or via the human using `/kanban ...` slash commands directly (bypasses the model entirely). True autonomous in-conversation kanban tool calls require Tars to run as a dispatcher-spawned worker for a specific task, or via a CLI-invoked profile with `toolsets: [..., kanban]`.

## 3. What already works with **no config change** (verified live, read-only)

```
$ hermes kanban boards
    SLUG   NAME      COUNTS
●   default  Default   (empty)
Current board: default

$ hermes kanban list        → (no matching tasks)
$ hermes kanban stats       → all statuses = 0
```

- `~/.hermes/kanban.db` (118 KB, SQLite, WAL mode — `.db-wal`/`.db-shm`/`.db.init.lock`/`.db.dispatch.lock` siblings present) already exists with a `default` board, empty. It was created lazily by some prior `hermes kanban ...` invocation (`init` is idempotent and auto-runs on first touch) — not by this recon.
- `/kanban` is registered unconditionally in the gateway's slash-command table (`gateway/run.py:14300`, dispatched via `gateway/slash_commands.py:_handle_kanban_command` → `hermes_cli/kanban.py:run_slash`). No `_check_kanban_mode`/toolset check wraps it — any user who can reach a runnable-commands slash dispatch in the gateway can already run `/kanban create ...`, `/kanban list`, `/kanban show <id>`, etc. against `~/.hermes/kanban.db` today. `/kanban create` auto-subscribes the originating Slack thread to the task's terminal events (complete/block/crash), so replies land back in-thread.
- The dispatcher itself: `kanban.dispatch_in_gateway` defaults to `True` (`hermes_cli/config_defaults.py:2322`) and is **not overridden** in the live `config.yaml` (absent = default), so it is presumably already ticking every `dispatch_interval_seconds` (default 60s) inside the running gateway process, ready to claim/spawn workers for any task that reaches `ready`.

## 4. Minimal config diff to unlock the **model-facing** kanban tool surface

Two independently-decidable knobs, pick based on what "enable kanban for Tars" is meant to mean:

**(a) Let a CLI-invoked profile act as kanban orchestrator** (list/create/link/unblock tasks by calling tools, not slash commands) — top-level `toolsets:`:
```yaml
toolsets:
  - hermes-cli
  - kanban
```
This only affects `hermes chat` / `hermes -p <profile> chat` CLI invocations (per §2), not the live Slack session. Live-reloads, no restart (config re-read via mtime-cached `load_config()`; `check_fn` cache TTL ~30s).

**(b) Let the live Tars Slack session itself call `kanban_*` tools** — not achievable via `platform_toolsets` (silently dropped, §2). The only supported mechanism is `HERMES_KANBAN_TASK`, which is set by the dispatcher when it spawns a worker subprocess for a claimed task — i.e. this requires routing the work *through* the board (`/kanban create` a task assigned to Tars's profile, or `kanban.dispatch_in_gateway` picking it up), not a static config flip. There is no minimal config key that grants an always-on interactive session the tool surface; it's scoped to task-bound worker processes by design (see the module docstring's "no shell-quoting footguns" / worker-isolation rationale).

Neither (a) nor (b) requires a gateway restart. `kanban.*` block (`dispatch_in_gateway`, `dispatch_interval_seconds`, `orchestrator_profile`, `default_assignee`, `auto_decompose`, etc.) is entirely absent from the live `config.yaml` (all defaults apply) and from `config.yaml.installer-default` (zero `kanban` hits) — the installer's generated commented template does **not** document the kanban block at all, so a fresh install gives no discoverability hint that it exists short of reading source or `hermes kanban --help`.

## 5. Storage paths this will create

- `~/.hermes/kanban.db` (+ `-wal`/`-shm`, `.init.lock`, `.dispatch.lock`) — **already exists**, created lazily, default board, currently empty (0 tasks in every status).
- `hermes kanban log` docstring references `<kanban-root>/kanban/logs/` for worker stdout/stderr — `~/.hermes/kanban/logs/` (not yet created; no tasks have run).
- Multi-board support exists (`hermes kanban boards`, `--board <slug>`) — additional boards presumably get their own sqlite file under the same root, unconfirmed since only `default` exists.

## 6. `hermes --help` / `hermes config get` surface

- `hermes --help` lists `kanban` as a top-level subcommand ("Multi-profile collaboration board (tasks, links, ...)") and `pause` ("Emergency stop: pause cron/kanban dispatch and new ...").
- `hermes kanban --help` exposes the full CLI: `init, boards, create, swarm, list/ls, show, assign, set-model, reclaim, reassign, diagnostics/diag, link, unlink, claim, comment, attach, attachments, attach-rm, complete, edit, block, schedule, unblock, promote, archive, tail, dispatch, daemon (deprecated), watch, stats, notify-subscribe, notify-list, notify-unsubscribe, log` — all of this is CLI/human-usable today regardless of the model tool-call gate.
- `hermes config get toolsets` → `- hermes-cli` (default, confirms the key is live and settable).
- `hermes config get platform_toolsets` → per-platform composite bundles (`slack: [hermes-slack]`, etc.) — confirms `kanban` is absent and, per §2, adding it here would have no effect.

## Sources (all on VM, read-only)

- `tools/kanban_tools.py` (gating functions, docstrings, tool registrations ~L2149-2248)
- `toolsets.py` (`"kanban"` composite definition ~L308-322, description references `kanban.dispatch_in_gateway`)
- `hermes_cli/tools_config.py` (`_get_platform_tools`, `CONFIGURABLE_TOOLSETS` exclusion comment ~L290-291, ~L2223-2360)
- `model_tools.py` (`_compute_tool_definitions`, `HERMES_KANBAN_TASK` force-include ~L389-410)
- `hermes_cli/config_defaults.py` (`"toolsets": ["hermes-cli"]` ~L12; `kanban:` block ~L2303-2359)
- `gateway/run.py` (`enabled_toolsets` computation ~L19725, ~L24712; slash command table ~L14300)
- `gateway/slash_commands.py` (`_handle_kanban_command` ~L432-538)
- `agent/delegation_context.py` (referenced, not read in full — `is_delegated_child_context` / `is_dispatcher_owned_worker_context`)
- Live VM state: `hermes config get toolsets|platform_toolsets`, `hermes kanban boards|list|stats`, `hermes --help`, `hermes kanban --help`, `ls ~/.hermes/kanban.db*`, `~/.hermes/config.yaml` (197 lines, no `kanban:` block, no top-level `toolsets:` block), `~/.hermes/config.yaml.installer-default` (1735 lines, zero `kanban` hits).
