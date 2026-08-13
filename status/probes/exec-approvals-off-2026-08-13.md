# Exec approvals OFF — full-auto execution on Tars (2026-08-13)

Gaetan's order: "we shouldn't have approvals, same behaviors as
`--dangerously-skip-permissions` for Claude". Config-only change on the VM,
live-proven end to end through Slack. **PASS.**

## 1. Mechanism (why one knob covers everything)

Source of truth is `tools/approval.py` in `~/.hermes/hermes-agent` (v0.20.0,
checkout at `4d10183`, **unpatched** — no source change was needed).

- **What fired in the 15:43 CEST incident**: `_execution_flag_findings()`
  (`tools/approval.py:1642`), the structural parser for interpreter code-eval
  flags driven by `_INTERPRETER_EXEC_FLAGS` (`:1194`, `python: {"-c"}`,
  `node: {"-e","--eval",…}`). Its reason string
  `"script execution via -e/-c flag"` is a literal dict entry at `:977`.
  Sibling trigger set: `DANGEROUS_PATTERNS` (`:693-956`, ~90 regex rules).
- **The live gate** is `check_all_command_guards()` (`:3738`, imported by
  `tools/terminal_tool.py:343`) — it honours `approvals.mode == "off"` at
  `:3788`. Same for the `execute_code` path at `:4265`.
  `check_dangerous_command()` (`:3420`) has no `mode: off` check but is **dead
  code** — zero callers outside tests.
- **The provider layer is bridged into the same knob, not independent.**
  `agent/codex_runtime.py:658-677` builds
  `_ServerRequestRouting(auto_approve_exec=X, auto_approve_apply_patch=X)` with
  `X = tools.approval.is_approval_bypass_active()` — the very function
  `approvals.mode: off` feeds. `agent/transports/codex_app_server_session.py`
  then returns `"accept"` immediately. The comment at `codex_runtime.py:655-663`
  names `approvals.mode: off` explicitly ("instead of double-gating").
- Third theoretical layer — codex's own sandbox profile via
  `~/.codex/config.toml` — is dead: the profile is intentionally never sent on
  `thread/start` (`codex_app_server_session.py:330-341`) and the file does not
  exist on the VM.

**Correction to the earlier investigation**: `approvals.cron_mode: deny` is
*not* a surviving independent gate. On both live paths the `mode: off` bypass
returns before the cron branch is consulted (`:3788` vs `:3799`; `:4265` vs
`:4269`). Left at its default `deny`. The probe below is a **cron** run and it
executed — that is the live proof.

## 2. Change applied

One file, `~/.hermes/config.yaml` on 192.168.0.9. Backup:
`~/.hermes/config.yaml.bak-2026-08-13-approvals`.

```diff
@@ end of file @@
+approvals:
+  mode: "off"
```

No `approvals:` block existed before → every value came from schema defaults
(`hermes_cli/config_defaults.py:2087-2090`, `mode: smart`). Quoted `"off"`
deliberately: YAML 1.1 parses bare `off` as boolean `False`
(`_normalize_approval_mode`, `:2891`, handles that too — quoting just doesn't
lean on the fallback).

Applied under `flock ~/.hermes/.wf3.lock` by a **proving merge** script that
re-parsed the file and asserted `new == old + {"approvals": {"mode": "off"}}`:

```
OK merge clean: +approvals.mode='off', 32 other top-level keys unchanged
```

**No gateway restart.** `systemctl --user show hermes-gateway -p
ActiveEnterTimestamp` = `Thu 2026-08-13 12:27:20 UTC` before *and* after. The
config cache key is `(st_mtime_ns, st_size)` (`hermes_cli/config.py:3292-3324`)
and `_get_approval_mode()` re-reads on every check, so the running gateway
picks it up on the next approval check.

## 3. Static probe — before/after, same command

Probe chosen so the permanent-approval store cannot mask it: `chmod -R 777`
matches pattern-key `world/other-writable permissions` (`:722`), which is *not*
one of the two allowlisted keys. `hermes approvals test` evaluates the
yolo/off bypass (step 5) **before** the allowlist step (step 6), so an `allow`
verdict provably comes from the knob.

BEFORE (`approvals.mode` → `smart`):
```json
{"verdict": "ask-approval", "exit_code": 2, "rule": "world/other-writable permissions",
 "detail": "matches a dangerous-command pattern; the runtime would raise an interactive approval prompt"}
```
AFTER (`approvals.mode` → `off`):
```json
{"verdict": "allow", "exit_code": 0, "rule": null,
 "detail": "approval bypass active (--yolo or approvals.mode: off); only hardline/deny rules would block"}
```
Incident's own class (`python3 -c …`) → `allow`, same bypass detail.
Hardline floor intact (by design):
```json
{"verdict": "hardline-deny", "exit_code": 3, "rule": "recursive delete of root filesystem",
 "detail": "matches the hardline blocklist (never bypassable, blocked even under --yolo / approvals.mode=off)"}
```

## 4. Live end-to-end probe — cron → terminal → Slack DM

Marker `APPROVAL-PROBE-89bfea9b`. Command carries a `chmod -R 777` (dangerous
pattern, **not** in `command_allowlist`) so the old gate would have fired:

```
mkdir -p /tmp/approval-probe-89bfea9b && chmod -R 777 /tmp/approval-probe-89bfea9b && echo APPROVAL-PROBE-89bfea9b
```

Triggered as a one-shot cron (job `1c5ebffcb41c`, created 13:57:47 UTC):
`hermes cron create "1m" "Run exactly this shell command and reply with its
output verbatim: <cmd>" --deliver slack --repeat 1`

### (a) Slack — marker top-level, no approval prompt

```
=== Message from Tars (U0BBH85NAKH) at 2026-08-13 15:59:43 CEST ===
Message TS: 1786629583.141469
APPROVAL-PROBE-89bfea9b
```
Read via Slack MCP on D0BBYNM01BL. Top-level, **no thread replies**, and **no
"⚠️ Command Approval Required" message anywhere** — the preceding message in the
DM is the 15:32:49 digest (`1786627969.351639`), whose *thread* is where the
15:43 incident prompt lives. Nothing between it and the marker.

### (b) Logs — terminal ran in 32 ms, no approval pause

`~/.hermes/logs/agent.log`:
```
2026-08-13 13:59:28,394 INFO cron.scheduler: Running job 'approval-probe-89bfea9b' (ID: 1c5ebffcb41c)
2026-08-13 13:59:38,142 INFO […] tools.terminal_tool: Creating new local environment …
2026-08-13 13:59:38,174 INFO […] agent.tool_executor: tool terminal completed (0.04s, …)
2026-08-13 13:59:42,976 INFO cron.scheduler: Job 'approval-probe-89bfea9b' completed successfully
2026-08-13 13:59:43,215 INFO cron.scheduler: Job '1c5ebffcb41c': delivered to slack:D0BBYNM01BL via live adapter
```
`terminal completed (0.04s)` — 32 ms between environment ready and completion:
no wait state. Under the old path this would have blocked on the gateway
approval prompt (`approvals.timeout` 300 s).
`grep -inE "approval|approve|dangerous|bypass|guard|chmod"` over the whole
13:58–14:00 window returns **only the two job-name lines** (the job is *called*
`approval-probe-…`). `grep -inE "approval required|Command Approval|awaiting
approval"` over `gateway.log` → **zero hits**.

Side effect proves the dangerous half really executed:
`drwxrwxrwx 2 gaetan gaetan 4096 Aug 13 13:59 /tmp/approval-probe-89bfea9b`
(mode 0777). Removed after the probe.

### (c) Approval store gained nothing

`~/.hermes/config.yaml` key `command_allowlist` (lines 136-138) before and
after the run, byte-identical — file sha256
`a112d73ee0fa3dafa97f1e07d071787c11ead63e5ac8f60950cc81c72554bf57`, mtime
`2026-08-13 13:55:22 UTC` (i.e. the config was untouched by the 13:59 run):
```yaml
command_allowlist:
  - find -delete
  - script execution via -e/-c flag
```
Both are pre-existing (the second is Gaetan's 15:43 "Approved permanently"
click). No new entry — because the bypass returns before the allowlist is ever
consulted. Left in place: unreachable, harmless.

## 5. What still gates (unchanged, by design)

- **Hardline blocklist** — `rm -rf /`, raw block-device writes, `mkfs`,
  shutdown/reboot, fork bomb, `kill -1` (`:434-480`). Plus the sudo-stdin guard
  (`:3765-3771`) and `approvals.deny` globs (currently `[]`). All fire *before*
  the bypass — same as `--dangerously-skip-permissions`, which likewise won't
  let Claude wipe a disk.
- **`write_file`/`patch` deny** on `~/.hermes/config.yaml`, `.env`, `SOUL.md`,
  `AGENTS.md`, `CLAUDE.md`, `.cursorrules` (`tools/file_tools.py:675-724`) —
  unconditional, untouched.
- **Not touched**: `approvals.mcp_reload_confirm` / `destructive_slash_confirm`
  (they confirm *Gaetan's own* `/clear` `/new` `/reset` `/undo` `/reload-mcp`,
  not Tars' execution). `approvals.cron_mode` left at `deny` (moot, see §1).

### ⚠️ One control that DOES lapse

`DANGEROUS_PATTERNS` also carried terminal-side coverage of
`~/.hermes/config.yaml` and `~/.hermes/.env` (`_HERMES_CONFIG_PATH`,
`_HERMES_ENV_PATH`) so that `sed -i` / `tee` / `>` / `cp` at those paths were
gated too — the module comment at `:279-283` says the pairing exists so the
`write_file` deny "isn't unpaired theater". With `mode: off` that terminal-side
half is bypassed: **Tars can now edit its own `config.yaml` (including
`approvals.*`) and read/write `.env` via shell commands.** The `write_file` deny
still stands, so it is a shell-only hole. Accepted cost of
`--dangerously-skip-permissions` parity; recorded so it isn't a surprise.

## 6. Rollback

```bash
ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); \
  ~/.local/bin/hermes config set approvals.mode smart'
```
(or restore `~/.hermes/config.yaml.bak-2026-08-13-approvals`). Live-reloads, no
restart.
