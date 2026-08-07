# WF5 — Kanban enabled for Tars

Date: 2026-08-07. Host: Tars VM (192.168.0.9). Hermes `hermes_agent 0.20.0`,
source `~/.hermes/hermes-agent/`.

## The change

One top-level key added to `~/.hermes/config.yaml`, inserted immediately before
`platform_toolsets:`. Nothing else was touched — no `kanban:` block, no
`platform_toolsets` edit, no restart.

```yaml
# WF5: unlocks the model-facing kanban_* tools. The kanban toolset is already
# in enabled_toolsets for cli+slack; this key is what _profile_has_kanban_toolset()
# (tools/kanban_tools.py check_fn) reads. Keep hermes-cli or CLI loses its tools.
toolsets:
- hermes-cli
- kanban
```

`hermes-cli` must stay in the list: this key *replaces* the default
`["hermes-cli"]` rather than extending it, so dropping it would strip every
normal tool.

## Why this key and no other

`tools/kanban_tools.py:52` is the whole gate:

```python
def _profile_has_kanban_toolset() -> bool:
    cfg = load_config()
    toolsets = cfg.get("toolsets", [])
    return "kanban" in toolsets
```

Both `_check_kanban_mode` (12 lifecycle tools) and
`_check_kanban_orchestrator_mode` (`kanban_list`, `kanban_unblock`) return this
same value for any process that is not a dispatcher-spawned worker. It reads the
**top-level** `toolsets:` key — not `platform_toolsets:`, and not a per-platform
setting. So one key flips the whole model-facing surface, on every platform at
once.

The `kanban` toolset was **already** present in `enabled_toolsets` for both `cli`
and `slack` before the change — the tail of `_get_platform_tools()`
(`hermes_cli/tools_config.py:2380+`) recovers non-configurable toolsets whose
tools are a subset of the platform composite, and the `kanban_*` names live in
`_HERMES_CORE_TOOLS` (`toolsets.py:81-85`). `check_fn` was the only thing
filtering them out. This corrects `status/probes/wf5/kanban-recon.md` §2, which
concluded the Slack session could never receive these tools; measured evidence to
the contrary is in `status/probes/wf5/kanban-implement.md`.

## Live-reload, no restart

`load_config()` is mtime-cached and the gateway recomputes `enabled_toolsets` per
turn (`gateway/run.py:19723`); `check_fn` results are TTL-cached 30s
(`tools/registry.py:216`). The running gateway picked the change up on its own —
`NRestarts=0`, active since 19:40:45 UTC, edit applied ~20:21 UTC, Slack turn
succeeded at 20:24 UTC.

## Storage

| What | Path | State |
|---|---|---|
| Board database | `~/.hermes/kanban.db` (+ `-wal`, `-shm`, `.init.lock`, `.dispatch.lock`) | Pre-existing, SQLite/WAL, one board `default`, 0 tasks |
| Dispatcher lock | `~/.hermes/kanban/.dispatcher.lock` | Present — dispatcher is live in the gateway |
| Worker logs | `~/.hermes/kanban/logs/` | Not created yet; appears on first spawned worker (`hermes kanban log`) |

## What was deliberately NOT enabled

No `kanban:` block was written, so every `kanban.*` default from
`hermes_cli/config_defaults.py:2308` remains in force untouched:

- `orchestrator_profile: ""` — **not set**. Task decomposition still falls back
  to the default profile; no orchestration topology configured.
- `auto_decompose: true` (default) — left alone. It only fires on tasks in
  `triage`, and both `kanban_create` and `hermes kanban create` default
  `triage=False`, landing new tasks in `todo`. Nothing auto-decomposes unless a
  task is explicitly created with `--triage`.
- `dispatch_in_gateway: true`, `dispatch_interval_seconds: 60`,
  `failure_limit: 2`, `dispatch_stale_timeout_seconds: 14400` — all pre-existing
  defaults, already in effect before this change.

Tool-level "orchestrator mode" (`kanban_list`, `kanban_unblock`) **is** now on
and cannot be separated from the rest: `_check_kanban_orchestrator_mode` reads
the same single config key. `kanban_list` is required to see the board at all,
so this is basic board use, not an orchestration opt-in.

## Rollback

```bash
flock ~/.hermes/.wf3.lock -c 'cp ~/.hermes/config.yaml.bak-kanban ~/.hermes/config.yaml'
```

Takes effect within ~30s, no restart.
