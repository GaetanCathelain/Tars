# Coordinator liveness watchdog — root cause + live state (2026-09-07)

## Trigger
Gaetan flagged: after "the coordinator will continue... stopping only for a
decision that needs you" (07:28Z), a "how's it going?" 5.5h later (13:09Z) got
"it stalled — I'm not running a coordinator, and I failed to catch that."

## Root cause (mechanism)
`delegate-to-cooper/SKILL.md` backed the primary tracked mailbox-wait with a
**one-shot** cron fallback (`--repeat 1`, "once in 15m"). A one-shot guards only
its first interval: it fires ~15 min in, sees the coordinator alive, and
expires. A coordinator that dies later strands silently until Gaetan asks.

Evidence — every historical fallback in `~/.hermes/cron/jobs.json` is one-shot
(`repeat.times:1`, "once in Nm"): MC-4226, Vercel budget, Prime Agent Orca,
GCN-90 auto-implem, Hermes runtime upgrade, etc. Tars's own live "fix" for THIS
incident, cron `e419d927145f` "Mnemosyne GCN-100 continuation fallback", was
itself one-shot and its prompt literally opens "One-shot tracking..." — a
one-shot fix for a one-shot-caused stall.

## Root cause (behavior)
Nothing made "running in the background" contingent on an active watchdog. Tars
asserted progress from having launched a coordinator, not from a verified
liveness signal. Launched ≠ tracked.

## Live state at investigation time
- Coordinator process **PID 3511845 is dead** (`ps` empty).
- Resume checkpoint `~/.hermes/research/mnemosyne/resume-checkpoint.json` last
  written **13:16:08Z** — stale.
- One-shot watchdog `e419d927145f` already fired: `state: completed`, spent.
- Orca worktree `gcn-105-mnemosyne-a` terminal is live (pty open) but shows
  `zsh: no matches found: claude-opus-4-8[1m]` — an **unquoted model id**; zsh
  globbed `[1m]` to nothing and the launch failed. Probable cause of the
  coordinator's death (ties to the 2026-09-05 1M-context standing correction).
- `hermes cron edit` refuses to convert the spent one-shot (job already
  `completed`); the fixed recipe uses `create`, so this does not affect it.

## Fix landed
- `delegate-to-cooper/SKILL.md`: one-shot fallback → recurring, self-retiring
  liveness watchdog (`every 20m`, `--repeat inf`, silent-when-alive, rm-on-
  terminal). Commit `8ad92da`, mirrored byte-identical to the VM.
- `SOUL.md` standing correction (2026-09-07): "running in the background" is
  honest only while a recurring watchdog is live; never report progress from
  having dispatched.

## Not done (Gaetan's / Tars's call — active project)
- The dead coordinator was NOT revived and no watchdog was armed on a corpse.
  The run needs relaunching with a properly quoted model id, then a recurring
  watchdog per the fixed recipe.
