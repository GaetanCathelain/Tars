# Patch drop — hermes-agent back to vanilla upstream (2026-08-13)

Executed on Gaetan's go, after the read-only verification pass that proved the
two local display patches were dead code with the live config. No secrets read,
no `sops`, no `~/.hermes/.env` values touched, no `config.yaml` or `SOUL.md` edit
(config needed none — the patch-added keys were never set).

## Before

```
$ cd ~/.hermes/hermes-agent && git status --short
 M package-lock.json
 M package.json
 M ui-tui/package.json          # unrelated npm drift, nothing else dirty

$ git log --oneline -4
4d10183 feat(display): per-conversation tool_progress override (display.platforms.<platform>.tool_progress_conversations)
d615ca8 feat(display): DM-scoped tool_progress override (display.platforms.<platform>.tool_progress_dm)
6e87d43 fix(tools): lazily bring up sandbox for vision_analyze reads

$ git branch --show-current
main
```

## Drop

```
$ git stash push -- package-lock.json package.json ui-tui/package.json
Saved working directory and index state WIP on main: 4d10183 ...
$ git status --short                     # clean
$ git reset --hard 6e87d43
HEAD is now at 6e87d43 fix(tools): lazily bring up sandbox for vision_analyze reads
$ git log --oneline -3
6e87d43 fix(tools): lazily bring up sandbox for vision_analyze reads
$ git stash pop
Dropped refs/stash@{0} (9113caf4ff390a997622fa54bd8aaf455b032341)
$ git status --short
 M package-lock.json
 M package.json
 M ui-tui/package.json          # drift restored intact
```

`git stash pop` prints "Your branch and 'origin/main' have diverged, and have 1
and 1 different commits each" — the known shallow-clone graft artifact (git
cannot walk past the `.git/shallow` boundary, so it misreports `6e87d43`).
The GitHub API check in the verification pass proved `6e87d43` is a genuine
ancestor of `origin/main` (`behind_by: 0`). Not a real divergence.

## Restart

Code change (`gateway/run.py`), so live-reload does not cover it. Waited for a
quiet window: last inbound Slack message 14:25:13 UTC, response delivered
14:25:46, agent cache idle-evicted 14:43:35 — only `mem_trim` housekeeping
after that. Restart issued at 14:46 from a plain ssh shell, outside the
gateway's process tree.

```
BEFORE  ActiveEnterTimestamp=Thu 2026-08-13 12:27:20 UTC
AFTER   ActiveEnterTimestamp=Thu 2026-08-13 14:46:45 UTC   (changed)
        systemctl --user is-active hermes-gateway → active
```

Startup log is clean — full sequence, no tracebacks:

```
14:46:50 Starting Hermes Gateway...
14:46:50 Secret redaction: ENABLED
14:46:50 [Slack] Authenticated as @tars in workspace Mobil...
14:46:50 [Slack] Socket Mode connected (1 workspace(s))  → ✓ slack connected
14:46:50 ✓ a2a connected
14:46:50 Gateway running with 2 platform(s)
14:46:51 Channel directory built: 95 target(s)
14:46:52 kanban dispatcher: holding singleton dispatcher lock
14:46:57 kanban dispatcher: embedded in gateway (interval=60.0s)
```

The gateway sent its own automatic shutdown notification to the home DM
(`14:46:44 Sent shutdown notification to home channel slack:D0BBYNM01BL`) —
built-in unit behavior on SIGTERM, not a message this session composed.
Nothing was sent to Slack by hand.

## Post-restart checks

```
$ ~/.local/bin/hermes config get display.platforms.slack.tool_progress --json
"off"

# gateway.log / agent.log / errors.log since 14:46:50:
grep -iE "error|traceback|exception|critical|warning"     → (none) in gateway.log, (none) in agent.log
grep -E "tool_progress_dm|tool_progress_conversations|platform_tool_conversations"  → (none)
```

Only post-restart line in `errors.log` is
`WARNING slack_bolt.AsyncApp: As you gave 'client' as well, 'token' will be unused.`
— pre-existing and benign: 13 occurrences in the file, one per gateway startup
(also present 2026-08-12 17:14:50 and 2026-08-13 12:27:24).

Narration verdict: with the patches gone, upstream resolves
`display.platforms.slack.tool_progress: "off"` → `progress_mode = "off"` →
`tool_progress_enabled = False` for every chat type (the un-patched resolver has
no chat-type branch at all). Same observable behavior as before the drop.

## VM-side doc edit

`~/.hermes/skills/hermes-operations/hermes-gateway-operations/references/slack-progress-scoping.md`
documented `tool_progress_dm` as a live, queryable key. Content replaced with a
6-line supersession note (file kept, path may be referenced by skills; mode
stays `0600`): patches dropped, checkout vanilla upstream `6e87d43`, narration
governed solely by `display.platforms.slack.tool_progress: "off"` for all chat
types, with the `hermes config get` verification one-liner.

This skill directory is VM-only — it has never been mirrored into this repo
(only `skills/hermes-operations/hermes-orchestration/SKILL.md` is). The VM side
moved here; if Tars later lands its own mirror of it, this is the newer content.

`SKILL.md` in the same directory still mentions `platform_tool_conversations`
inside a "never use this to approximate display" warning — left as is, that
note stays correct as a don't-do-this.

## Repo files changed

- `patches/0001-tool-progress-dm.patch` — deleted
- `patches/0002-tool-progress-conversations.patch` — deleted
- `patches/README.md` — rewritten: no local code patches, re-apply recipe in git
  history, google-workspace `SKILL.md` venv-python edit called out as a
  vendored-skill edit rather than a code patch
- `docs/facts.md` — supersession sentence appended in place to the "Capability
  and narration are separate axes" row (the axis separation itself still holds)
- `docs/specs/tars-profile.md` — new dated correction line in the existing
  standing-corrections blockquote
- `docs/plans/apply-slack-channel-boundaries.md` — SUPERSEDED banner at the top;
  its display half is undone, its mention-routing half stands
- `docs/specs/soul-v2-rationale.md`, `docs/specs/soul-v2-changemap.md` — the
  `tool_progress_conversations` flag they pass upstream marked moot
- `status/probes/patch-drop-2026-08-13.md` — this file
