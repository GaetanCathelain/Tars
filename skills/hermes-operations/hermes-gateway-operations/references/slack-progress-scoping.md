# Slack tool-progress scoping — SUPERSEDED 2026-08-13

The local display patches (`tool_progress_dm`, `tool_progress_conversations`) were
dropped on 2026-08-13; `~/.hermes/hermes-agent` is vanilla upstream (`6e87d43`).
Tool-call narration is now governed solely by `display.platforms.slack.tool_progress: "off"`
(quoted) in `config.yaml`, which applies uniformly to all chat types — DM, group DM and
channel alike. There is no per-DM or per-conversation display override any more.
Verify with: `hermes config get display.platforms.slack.tool_progress --json` → `"off"`.
