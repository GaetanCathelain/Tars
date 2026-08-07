# WF5 inventory — surface 4: runtime-injected persona & permission surface (VM)

Read-only recon over `ssh gaetan@192.168.0.9`. No edits, no `hermes config set`,
no restart. Hermes v0.20.0 source at `~/.hermes/hermes-agent/` on the VM.
All secret VALUES redacted; only key names / presence / lengths reported.

## 1. `~/.hermes/config.yaml` (203 lines) — keys that shape persona/permissions/tools/autonomy

| Key | Current value | Controls |
|---|---|---|
| `model.default` / `model.provider` | `gpt-5.6-sol` / `openai-codex` (`base_url: https://chatgpt.com/backend-api/codex`) | Which model Tars runs on |
| `agent.max_turns` | `500` | Hard cap on tool-call turns per session |
| `agent.reasoning_effort` | `medium` | Model reasoning budget |
| `agent.personalities.*` | 12 built-in preset personas (helpful/concise/technical/…/hype) | NOT SOUL — a switchable persona-string library, unused unless a persona is explicitly selected. SOUL.md is the primary identity (see §3) |
| `terminal.backend` | `local` | Where `terminal()` tool commands actually execute — **local = cooper itself**, no container/VM isolation |
| `terminal.timeout` | `180`s | Per-command terminal() timeout |
| `terminal.container_*` | cpu 1 / mem 5120MB / disk 51200MB / persistent true | Only relevant if backend were `docker`; inert while `backend: local` |
| `browser.cdp_url` | `http://127.0.0.1:9223` | Local Chrome DevTools endpoint for browser tool |
| `tool_loop_guardrails.*` | warn 2-3x, `hard_stop_enabled: false` | Nags on repeated tool failures but **never force-stops** the agent |
| `code_execution.timeout` | `300`s | Timeout for the `execute_code` (sandboxed Python) tool |
| `code_execution.max_tool_calls` | `50` | Max nested tool calls inside one execute_code script |
| `memory.memory_enabled` / `user_profile_enabled` | both `true` | Cross-session memory + user-profile injection (see §3, volatile band) |
| `delegation.max_iterations` | `50` | Cap on `delegate_tool` (Orca/sub-agent delegation) loop iterations |
| `toolsets` | `[hermes-cli, kanban]` | Base model-facing toolsets; comment in-file: kanban key is what `_profile_has_kanban_toolset()` reads — **already enabled** |
| `platform_toolsets.slack` | `[hermes-slack]` | Slack-specific toolset added on top of `toolsets` when platform=slack |
| `plugins.enabled` | `[hermes-lcm, rtk-rewrite]` | Loaded plugins |
| `gateway.platforms.slack.enabled` | `true` | Slack adapter on/off |
| `gateway.platforms.slack.require_mention` | `true` | Must @-mention Tars in channels |
| `gateway.platforms.slack.strict_mention` | `true` | Tightens mention matching (this is the `SLACK_STRICT_MENTION` control — lives in config.yaml, NOT `.env`, see §2) |
| `gateway.platforms.slack.unauthorized_dm_behavior` | `ignore` | What happens on a DM from a non-allowed user |
| `gateway.platforms.a2a.enabled` / `.extra.port` | `true` / `9900` | A2A protocol surface, port 9900 |
| `mcp_servers.*` | slack (docker, korotovsky/slack-mcp-server), notion (docker, `NOTION_TOKEN: ${NOTION_API_TOKEN}` — env-var reference, not a literal value) | External MCP tool servers wired in |
| `session_reset.mode` | `none` (idle_minutes 1440, at_hour 4) | No forced session resets |
| **`approvals.*`** | **absent from config.yaml entirely** | See §5 — absence ⇒ code default `mode: manual` applies (evidence below) |
| **security scanner** | **no config key** | Prompt-injection scanning of context files (SOUL.md/AGENTS.md/.cursorrules) is hardcoded in `tools/threat_patterns.py`, invoked from `prompt_builder.py:74` — not a config toggle, not an approval gate on Tars's own actions |
| approval/confirmation/permission/sandbox/allow_/deny_/restricted/autonomy | **no matching key found** (`grep` empty) | Confirms config.yaml carries no explicit permission-gate config — the gate that exists is a code default, not a doc/config setting (§5) |

## 2. `~/.hermes/.env` — key names only, no values

`GMAIL_ADDRESS`, `GMAIL_APP_PASSWORD`, `HINDSIGHT_MODE`, `LINEAR_API_KEY`,
`NOTION_API_TOKEN`, `SLACK_ALLOWED_USERS`, `SLACK_APP_TOKEN`, `SLACK_BOT_TOKEN`,
`SLACK_HOME_CHANNEL`, `SLACK_MCP_XOXC_TOKEN`, `SLACK_MCP_XOXD_TOKEN` (11 keys total).

- `SLACK_ALLOWED_USERS`: **SET** (non-empty, length 11 chars — consistent with one Slack user ID).
- `SLACK_STRICT_MENTION`: **not an env var** — confirmed absent from `.env`; the control lives only as `gateway.platforms.slack.strict_mention: true` in config.yaml (§1).

## 3. System-prompt assembly — order, sources, file:line evidence

Entry point `AIAgent._build_system_prompt()` (`run_agent.py:4494`) forwards to
`agent/system_prompt.py`. Real assembly:

```
build_system_prompt_parts()  → agent/system_prompt.py:152-406
build_system_prompt()        → agent/system_prompt.py:561-580
  joined = "\n\n".join(parts["stable"], parts["context"], parts["volatile"])   # :579
```

Three bands, concatenated **stable → context → volatile**:

**stable** (identity + static behavioral guidance, cache-friendly prefix):
1. SOUL.md, loaded via `_r.load_soul_md()` (`system_prompt.py:41-46`) — falls back to hardcoded `DEFAULT_AGENT_IDENTITY` (`prompt_builder.py:144-152`) only if SOUL.md is absent/unreadable
2. `HERMES_AGENT_HELP_GUIDANCE` (`:53`) — points the model at Hermes's own docs
3. `TASK_COMPLETION_GUIDANCE` (`:61-62`, only if tools loaded)
4. `PARALLEL_TOOL_CALL_GUIDANCE` (`:72-73`)
5. Tool-conditional behavioral guidance, joined into one block: `MEMORY_GUIDANCE` if `memory` tool present, `SESSION_SEARCH_GUIDANCE` if `session_search` present, `SKILLS_GUIDANCE` if `skill_manage` present, `KANBAN_GUIDANCE` if kanban tools present (`:77-94`)
6. `STEER_CHANNEL_NOTE` (`:98-99`)
7. `computer_use_guidance()` (`:105-107`, if `computer_use` tool present)
8. `build_nous_subscription_prompt()` (`:109-111`)
9. Model-specific tool-use enforcement: `TOOL_USE_ENFORCEMENT_GUIDANCE`, optionally `GOOGLE_MODEL_OPERATIONAL_GUIDANCE` / `OPENAI_MODEL_EXECUTION_GUIDANCE` (`:119-146`)
10. Explicit model-identity string (`:179-185`)
11. `build_environment_hints()` (`:197`) — cwd/backend/OS hints
12. Coding-workspace prefix/trailing context, local-toolchain probe line, cross-profile isolation warning, and (TUI platform only, inline here) the resolved platform hint (`:206-314`)

**context** (varies with cwd/session, not cached across cwd changes):
13. Coding-workspace + trailing context copies, any caller-supplied `system_message` override (`:317-327`)
14. `build_context_files_prompt()` — AGENTS.md / `.cursorrules` / `CLAUDE.md` / `.hermes.md`, discovered from cwd up to git root, each run through `_scan_context_content()` prompt-injection scan first (`:341-345`, scanner at `prompt_builder.py:52-79`); SOUL.md is explicitly excluded here (`skip_soul=_soul_loaded`) since it already landed in the stable band

**volatile** (turn-varying, always last):
15. Skills index — `build_skills_system_prompt()` (`:361-362`, computed earlier at `:170-176`)
16. Memory block + user-profile block from `agent._memory_store` (`:364-373`)
17. External memory-provider block (Hindsight, per `memory.provider: hindsight`) (`:376-380`)
18. Timestamp + `Platform: <name>` line (`:390-401`) — always last, guarantees the tail is never cache-stable

**Precedence for "where does a redesign rule live most effectively"**: SOUL.md (band-1, position-1) has first-token priority and is the only piece that is Tars's actual identity; anything added there outranks every subsequent guidance block by recency-of-primacy and is never truncated by the volatile-band memory/skills churn. A rule placed in a skill instead lands in band-3 (volatile, position-15) — present every turn but after ~10+ other stable-band paragraphs.

## 4. Restart vs. live-reload — verdict

**Live-reload confirmed for config.yaml and .env, no gateway restart needed.**

Evidence: `hermes_cli/config.py` caches `load_config()` / `cfg_get()` results
keyed on the config file's `(mtime_ns, size)` (`config.py:236-249, 2941-2949,
3032-3039`) — any edit changes mtime, invalidating the cache so the *next*
read (next turn) picks up fresh values. `.env` is explicitly re-read
per-turn in the gateway: `gateway/run.py:1838` `_reload_runtime_env_preserving_config_authority()`,
docstring: *"Gateway processes are long-lived, so per-turn code reloads
~/.hermes/.env"* (`:1841`). This matches the ~30s figure already documented
in project CLAUDE.md/facts.md.

Keys confirmed on this live path: `approvals.*`, `toolsets`, `platform_toolsets`,
`gateway.platforms.slack.*` (all read via `cfg_get`/`_get_approval_config()`
which shares the same mtime-cache — `approval.py:2930`). SOUL.md and skill
files are read from disk fresh at every `_build_system_prompt()` call (§3),
so those are effectively live too.

**Likely restart-required** (not directly exercised this probe, flagged from
code shape, not proven): `mcp_servers.*` (docker subprocess servers spun up
at connect time — CLI has an explicit periodic watcher for this,
`cli.py:17417` `_check_config_mcp_changes()`, but no equivalent watcher was
found for the gateway process itself), `plugins.enabled` (Python import-time
loading), and `gateway.platforms.a2a.extra.port` (socket bind at startup).
Do not treat this paragraph as verified — a sibling session is live on this
gateway right now, so don't restart to test it.

## 5. Constrains-to-drop / enables-to-turn-on for the redesign

**Currently constrains Tars, redesign says drop:**
- **`approvals.mode`** — absent from config.yaml ⇒ code default is `"manual"`
  (`tools/approval.py:2936-2939` `_get_approval_mode()` → `.get("mode", "manual")`;
  confirmed no `approvals:` key anywhere in config.yaml). This is Hermes's
  built-in "dangerous command" gate — `DANGEROUS_PATTERNS` explicitly matches
  `sudo` (with flags) and blocks `sudo -S` stdin password guessing
  (`approval.py:381-516`). In manual mode a dangerous command (sudo included)
  triggers an interactive/gateway approval prompt rather than running freely.
  This is the single concrete config lever for "drop the sudo guardrail":
  setting `approvals: {mode: off}` in config.yaml (per-key edit, `.bak` +
  `flock`, live-reloads next turn per §4) removes the gate. Not touched this
  session — read-only.
- `terminal.backend: local` is already the least-restrictive option (no
  container isolation) — nothing to loosen there.

**Currently neutral / already enables the redesign:**
- `toolsets: [hermes-cli, kanban]` + `platform_toolsets.slack: [hermes-slack]`
  — kanban toolset (work-tracking) is already live per the in-file WF5
  comment, matches the git history ("Kanban enabled … worker lane proven").
- `delegation.max_iterations: 50` — delegate_tool (Orca-driving) budget
  already configured, no additional toolset flag found gating it beyond this.
- No `cron` model-facing toolset key was found in config.yaml's `toolsets`
  list or in `tools/toolsets.py`/registry greps; `hermes cron create` exists
  as a CLI-level command (per operating-tars skill), separate from a
  model-facing toolset — if the redesign wants the model itself to call a
  `cron_*` tool (vs. Tars shelling out to the `hermes` CLI via `terminal()`),
  that would need identifying/adding a toolset name, not just a config flag.
  Flagging as unresolved, not fabricating a toolset name.

**Unrelated to the redesign, left as-is per hard rules:**
- `gateway.platforms.slack.require_mention` / `strict_mention` / `unauthorized_dm_behavior`
  and `.env`'s `SLACK_ALLOWED_USERS` (confirmed SET) are the trust boundary —
  now more load-bearing, not less, per the task brief. Not touched.
- `code_execution.timeout: 300` / `max_tool_calls: 50` — unrelated to
  persona/permission redesign, noted for completeness only.
