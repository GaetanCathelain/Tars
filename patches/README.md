# Local patches on the VM's hermes-agent checkout

**No local hermes-agent code patches as of 2026-08-13** — `~/.hermes/hermes-agent`
on the Tars VM is vanilla upstream (`6e87d43`). The two display patches
(`tool_progress_dm` `d615ca8`, `tool_progress_conversations` `4d10183`) were
dropped that day: upstream already honors `display.platforms.slack.tool_progress: "off"`
globally for every chat type, so both patches were dead code with the live config.
Nothing needs re-applying after a `git pull` / `/upgrade` any more.

Historical patches and the full re-apply recipe: see git history of this file.
Drop evidence: `status/probes/patch-drop-2026-08-13.md`.

The google-workspace `SKILL.md` venv-python edit is a **vendored-skill edit, NOT a
code patch** — see the upgrade memory / `status/lane-a.md` 2026-08-11 GCN-31.
