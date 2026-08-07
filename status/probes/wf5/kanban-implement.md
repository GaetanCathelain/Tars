# WF5 Kanban implement — evidence

Date: 2026-08-07, Tars VM (192.168.0.9). Spec: `docs/specs/wf5-kanban.md`.
Verdict: **PASS** — kanban model tools live on CLI *and* Slack, no restart, board clean.

## 0. Recon correction (found before editing)

`kanban-recon.md` §2 concluded that `kanban` never enters `enabled_toolsets` for
gateway/Slack sessions, and that only CLI sessions could ever get the tools. Both
halves are wrong:

- The CLI path does **not** read top-level `toolsets:` either — `oneshot.py:396`
  and `cli_commands_mixin.py:764` both call `_get_platform_tools(cfg, "cli")`,
  the same resolver the gateway uses.
- The `tools_config.py:291` comment the recon read as an exclusion
  ("Non-configurable toolsets that `_get_platform_tools` resolves at read time —
  `kanban` …") says the opposite: the recovery block at `tools_config.py:2380+`
  adds `kanban` back, because the `kanban_*` names are in `_HERMES_CORE_TOOLS`
  (`toolsets.py:81-85`) and thus a subset of every platform composite.

Measured baseline, **before** any edit (venv python, read-only):

```
toolsets(effective): ['hermes-cli']
cli   kanban in enabled_toolsets: True
slack kanban in enabled_toolsets: True
_check_kanban_mode: False
_check_kanban_orchestrator_mode: False
```

So `check_fn` was the only gate, on both platforms. That made the fix one key
instead of the "no config key can do this" dead end the recon described.

## 1. Change applied

Backup taken once: `~/.hermes/config.yaml.bak-kanban` (5908 bytes, 20:21 UTC).
Edit executed under `flock ~/.hermes/.wf3.lock`:

```
OK: inserted top-level toolsets before platform_toolsets
```

`diff config.yaml.bak-kanban config.yaml` — the entire change:

```
143a144,149
> # WF5: unlocks the model-facing kanban_* tools. The kanban toolset is already
> # in enabled_toolsets for cli+slack; this key is what _profile_has_kanban_toolset()
> # (tools/kanban_tools.py check_fn) reads. Keep hermes-cli or CLI loses its tools.
> toolsets:
> - hermes-cli
> - kanban
```

`grep -c '^kanban:' ~/.hermes/config.yaml` → `0` (no `kanban:` block written;
all `kanban.*` defaults untouched).

## 2. check_fn now passes

```
YAML parses OK
toolsets(effective): ['hermes-cli', 'kanban']
_check_kanban_mode: True
_check_kanban_orchestrator_mode: True
```

## 3. Tools actually in the model schema

Built the definitions exactly as `oneshot.py` / `gateway/run.py` do
(`get_tool_definitions(enabled_toolsets=_get_platform_tools(cfg, plat))`):

```
cli   total tools=41 kanban tools=12
    ['kanban_attach', 'kanban_attach_url', 'kanban_attachments', 'kanban_block',
     'kanban_comment', 'kanban_complete', 'kanban_create', 'kanban_heartbeat',
     'kanban_link', 'kanban_list', 'kanban_show', 'kanban_unblock']
slack total tools=41 kanban tools=12
    [... identical 12 ...]
```

## 4. Fresh CLI invocation — tools really execute

```
$ hermes chat -Q -q "Call the kanban_list tool and report exactly how many tasks
                     are on the default board. Do not create anything."
session_id: 20260807_202223_d077aa
0 tasks.
```

Not a hallucinated answer — `~/.hermes/logs/agent.log` for that session:

```
20:22:31 WARNING [20260807_202223_d077aa] agent.tool_executor: Tool kanban_show returned error (0...
20:22:45 INFO    [20260807_202223_d077aa] agent.tool_executor: tool kanban_list completed (0.00s, 94...
```

(`kanban_show` erroring is correct — the model called it with no task id.) The
last `check_fn _check_kanban_mode returned False` line in `agent.log` is
20:21:42, i.e. pre-edit; no such warning appears for the 20:22:23 session.

## 5. Slack E2E — the live gateway session, no restart

One DM sent as Gaetan (U08BDJAMSRZ) to the Tars home DM D0BBYNM01BL via the VM's
personal xoxc/xoxd creds (`curl -K` stdin, no secret in argv), `ok=True
ts=1786134255.466109`. Polled `conversations.replies` on that ts:

```
--- U08BDJAMSRZ 'kanban check: call kanban_list and reply with how many tasks are on the default board (one line).'
--- U0BBH85NAKH ':clipboard: kanban_show...\n:clipboard: kanban_list...'
--- U0BBH85NAKH '0 tasks on the default board.'
```

Tars (U0BBH85NAKH) invoked the kanban tools inside the live Slack session. This
is the direct disproof of recon §2's "there is no config key that gives the live
interactive Tars-in-Slack session the `kanban_*` model tools".

## 6. No restart, live-reload confirmed

```
NRestarts=0
ActiveState=active
ActiveEnterTimestamp=Fri 2026-08-07 19:40:45 UTC
now:                 Fri 2026-08-07 20:24:38 UTC
```

Gateway has been up continuously since 19:40:45 — before the ~20:21 edit and
before the 20:24 Slack turn. Restart protocol never invoked because it was never
needed.

## 7. Board left clean

```
$ hermes kanban stats
  triage 0  todo 0  scheduled 0  ready 0  running 0  blocked 0  done 0
```

Storage: `~/.hermes/kanban.db` (118784 B, pre-existing, unchanged mtime 18:43),
`~/.hermes/kanban/.dispatcher.lock` present (dispatcher live in gateway),
`~/.hermes/kanban/logs/` still absent (no worker has run).

## Caveats

- Enabling lifecycle tools necessarily enables the two orchestrator tools
  (`kanban_list`, `kanban_unblock`) — same config key, no separation available.
  See spec for why that is basic board use.
- The recon's claim that `/kanban` slash commands work unconditionally was not
  re-verified here; it was not needed for this change.
- Scratch files used on the VM (`/tmp/wf5_*`) contained no secrets and were
  removed after verification.
