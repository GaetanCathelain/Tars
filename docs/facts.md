# Operational facts — verified live 2026-08-07

Everything below was measured on the Tars VM (Hermes v0.20.0, Ubuntu 24.04)
during WF3–WF5. Evidence pointers are files under `status/probes/`. Re-verify
after any Hermes upgrade.

## Hermes CLI / runtime

| Fact | Evidence |
|---|---|
| Oneshot chat is `hermes chat -Q -q '<query>'` — `--oneshot` does not exist | `wf3-s2`, `wf3-s4` |
| `hermes mcp list` is registry-only (never connects); liveness = `hermes mcp test <name>` | `wf3-s4` |
| MCP tool prefix is `mcp__<server>__<tool>` (double underscore) | `wf3-s4` |
| Systemd units are `--user`: `hermes-gateway.service`, `hermes-cloakbrowser.service`; `export XDG_RUNTIME_DIR=/run/user/$(id -u)` over ssh | `wf3-s5` |
| journald carries WARNINGs only; INFO lives in `~/.hermes/logs/{gateway,agent,errors,mcp-stderr}.log` | `wf4/p01` |
| `config.yaml` live-reloads (~30 s); no restart needed for most changes (kanban enable needed none) | `wf5/kanban-implement` |
| `hermes memory setup` and `hermes skills install` (no `--yes`) CANCEL headless and still exit 0 — use `hermes config set` | lane B handoff |
| `hermes auth add … --type oauth` does NOT write `model.base_url` — set it explicitly and verify no openrouter default survives | `wf3-s5` |
| `tools/mcp_tool.py _build_safe_env()` filters subprocess env: `.env` keys are NOT visible to MCP servers — only the server's own `env:` block + PATH/HOME allowlist | `wf3-s2` |
| Async delivery to the home channel: `hermes cron create "1m" "…" --deliver slack --repeat 1` (bare platform name ⇒ `SLACK_HOME_CHANNEL`). `/bg` always replies to the INVOKING surface — it can never deliver to home | `wf4/p13` |
| Skills live at `~/.hermes/skills/<name>/SKILL.md` and are loaded into model context unprompted when relevant | `wf5/orca-implement` |
| Kanban: single top-level `toolsets: [hermes-cli, kanban]` key unlocks 12 `kanban_*` tools on CLI **and** Slack; dispatcher spawns real workers that drive cards ready→running→done | `wf5/kanban-*` |
| Known intermittent: multi-step Codex turns can die with "response remained incomplete after 3 continuation attempts"; plain/single-tool turns unaffected. Retry once before concluding failure | `wf4/p13`, triage: later |
| tirith security scanner: enabled but not installed — command vetting is pattern-matching only (pre-existing; triage: later) | `wf3-s5` |

## Slack platform

| Fact | Evidence |
|---|---|
| Tars is an **Agent-class** app: Slack blocks ALL slash commands in its chat surface (it is threads under the hood). `/sethome` equivalent = write `SLACK_HOME_CHANNEL` to `.env` + restart (source-verified getenv-only) | `cutover-sethome` |
| Tars replies **in-thread**: `conversations.history` is blind to its replies — always poll `conversations.replies` on the sent message ts | `wf4/p01` |
| The allowlist reject is a **positive WARNING**: `[Slack] Early reject of unauthorized user <id> in channel <id>` (adapter, pre-dispatch, pre-media-fetch) — grep for it, never settle for absence-of-reply | `wf4/p16` |
| `SLACK_STRICT_MENTION`/`SLACK_ALLOWED_USERS` are written to `os.environ` by the plugin AFTER exec — `/proc/<pid>/environ` shows NO `SLACK_*` on a healthy gateway | `wf4/p03` |
| Bots don't auto-join channels — invite manually. One Slack token pair = one live gateway | `cutover-notes`, D2 |
| The claude.ai `claude_ai_Slack` MCP connector authenticates **as Gaetan** (user token `U08BDJAMSRZ`) — never usable as a non-Gaetan sender | `wf4/p03` |
| IDs: Gaetan `U08BDJAMSRZ` · Tars bot user `U0BBH85NAKH` · app `A0BC0GXH78R` · team `T7V1UGJ82` · home DM `D0BBYNM01BL` · test channel `C08RWSTU9LK`. VM env names: `SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN` (cookie sent as-stored, URL-encoded) | `wf4/p01` |

## SOUL / persona design

**Identity-frame bug class** (`wf4/diag-empty-response`, fixed same day): any
hard rule of the form "answer only X" MUST name X's platform identity for every
surface. Channel (multi-user) turns prefix messages with `[<user-id> | …]`;
without an ID→person mapping the model cannot recognize X and obeys the rule by
staying silent. Hermes has **no silence terminal** — deliberate model silence
becomes N empty retries, a 61 s stall and a user-facing error. Rules must
prescribe a minimal literal reply (Tars uses `·`) instead of "ignore in
silence". The adapter allowlist rejects strangers before the model, so the
model-level rule is a backstop only.
