# GCN-68 (filed as GCN-63 probe artifact) — herdr as the lane substrate

Live probe run on cooper, 2026-08-16 ~08:55–09:00 UTC. Plugin: `claude-code-herdr-plugin`
v1.3.0, loaded copy at
`/home/gaetan/.claude/plugins/cache/claude-code-herdr-plugin/claude-code-herdr-plugin/1.3.0/`
(marketplace source mirror at `~/.claude/plugins/marketplaces/claude-code-herdr-plugin/` —
identical `skills/` content). Skill: `skills/claude-to-codex/SKILL.md` + `references/*.md` +
`scripts/codex.py`. Installed `herdr` binary: `/home/gaetan/.local/bin/herdr`, client v0.7.5,
protocol 17 — **newer than the skill docs, which self-report "verified on herdr v0.6.0"**;
several CLI shapes have drifted (see Lifecycle/ssh-shape rows).

## Verdict table

| Axis | Verdict | Confidence |
|---|---|---|
| Lifecycle (spawn/send/read/end a Claude lane) | **Works.** Full loop proven: workspace→pane→`agent start --kind claude`→`pane run`(nonce)→file on disk→`pane read`→`pane close`. | PROVEN-live |
| ssh shape (no TTY, no `HERDR_PANE_ID`) | **Split verdict.** Raw `herdr` CLI (pane/agent/workspace verbs) works fine with no controlling terminal and outside a herdr pane — confirmed via `setsid bash -c '...' </dev/null`. But the recommended one-tool driver `scripts/codex.py` **hard-requires `HERDR_PANE_ID`** and fails (`NAMING_FAILED`, exit 3) when it's unset — no flag to override. Also: **the herdr *server* itself was down** and the skill/traps doc claims it "can't be started from a script — needs a TUI" (`pitfalls-and-traps.md` trap 17); that claim is **false on this build** — `setsid herdr server </dev/null &` started it headlessly, no TTY, and it came up serving instantly. | PROVEN-live (both the works-without-TTY part and the doc's wrong claim); INSPECTED (no override flag in codex.py) |
| Busy semantics (mid-task nonce send) | **Same shape as `orca terminal send`.** A message sent to a pane mid-genuine-CPU-busy-loop (Bash tool, 20s Python spin, confirmed `agent_status: working` mid-loop) does not interleave or get lost — it lands in Claude Code's own input queue ("Press up to edit queued messages") and auto-processes as the next turn once the busy tool call returns. herdr adds nothing here; it's Claude Code's own composer queueing, herdr is just the byte-pipe. | PROVEN-live |
| Return path (push vs poll) | **No push channel.** `agent wait` (blocking, backgroundable) and `events.subscribe` (long-lived socket stream) are the only two primitives, and both require a live *driver-side* process parked on the socket — there is no webhook/callback config. `herdr notification show` is a **local desktop toast for the TUI human**, not a remote signal. For a stateless, wake-based Tars (no long-lived process), herdr changes nothing about the lane→Tars leg: Tars still needs either its own cron poll (`ssh cooper 'herdr agent get <pane>'` / `pane read`) or the lane's own agent to call out (`hermes send`) on completion — exactly the prior wire-4 finding, herdr only reshapes the Tars→lane leg. | INSPECTED (read `waiting-and-async.md`, `events-and-subscribe.md`, CLI `--help`); no live push test possible since there's nothing to push to |
| Orca coexistence | **No conflict.** herdr is its own Rust daemon/PTY multiplexer ("like tmux but not tmux" — `architecture.md`), own socket `~/.config/herdr/herdr.sock`, independent of Orca's own terminal manager. Live test: split a herdr pane with `--cwd` pointed at *this* Orca-managed worktree (`/home/gaetan/dev/orca-worktrees/Tars/improvements`), spawned Claude in it, ran a read-only `git status`/`pwd`/branch check — clean, no lock contention, no interference with the concurrently-running orchestrator session in the same tree. | PROVEN-live |
| Cross-harness (codex / opencode) | **codex present, opencode absent.** `scripts/codex.py start` failed outside a herdr pane (see ssh-shape row) — never got to smoke-test Codex through the polished driver. Fell back to raw herdr (`agent start --kind codex --pane <id>`): spawn worked, first send raced Codex's TUI init and was silently dropped (mirrors the skill's documented "spawn-readiness" trap — exactly what `codex.py` exists to paper over), second send landed and submitted correctly, but the installed `codex` CLI itself then failed with `stream error: Failed to refresh token: 401 Unauthorized` — an environment credential problem on cooper, unrelated to herdr. herdr-level mechanics (spawn, send, submit) are proven; end-to-end Codex task completion is untested because of the auth failure. `opencode` binary not found on cooper — untested. | PROVEN-live (herdr↔codex wiring) / UNTESTED (codex task completion, opencode) |

## Bottom line for the respec

herdr is a legitimate substrate for the **Tars→lane** leg (spawn, send, read, busy-queue
semantics all work, including headless over the ssh-cooper shape) **if you drive it via the
raw `herdr` CLI**, not via `scripts/codex.py` as shipped — that script assumes it's called
from inside a herdr pane and has no override for a bare-ssh caller. It does **not** solve the
**lane→Tars** leg any better than the raw-Claude-session option already probed: no push
channel exists, so Tars still needs a poller or the lane needs to `hermes send` on completion.
And the herdr *server* has to be alive on cooper for any of this to work — it was down at
probe start and had to be started headlessly first, which the shipped docs incorrectly say is
impossible from a script.

## Evidence log (with timestamps, confidence tags)

All work was scoped to a throwaway `/tmp/gcn68-herdr-probe/` tree and one throwaway herdr
workspace (`wK`); nothing under `$HOME` outside `~/.config/herdr/` (the socket/log dir,
untouched in content) was created. No Slack, no Tars VM, no `secrets/`, no pre-existing
panes/workspaces (`w1`, `w2`, `wF`, `wG`, `wJ` — 5 pre-existing Claude panes from a prior
session, left completely alone).

1. **Plugin location** (INSPECTED) — `find ~ -iname '*herdr*'` located the live plugin copy at
   `~/.claude/plugins/cache/claude-code-herdr-plugin/claude-code-herdr-plugin/1.3.0/` (the copy
   actually loaded by Claude Code) and the git-backed marketplace mirror at
   `~/.claude/plugins/marketplaces/claude-code-herdr-plugin/`. Read `SKILL.md` in full plus
   references: `herdr-cli.md`, `agent-vs-pane.md`, `multi-agent-patterns.md`,
   `fake-and-custom-agents.md`, `waiting-and-async.md`, `events-and-subscribe.md`,
   `pane-lifecycle.md`, `pitfalls-and-traps.md` before running anything.

2. **Server was down** (PROVEN-live, ~08:53 UTC) — `herdr status` → `server: status: not
   running`. `ps aux | grep herdr` → no process. Log tail (`~/.config/herdr/herdr-server.log`)
   showed last activity 2026-07-27; server had not been running for ~3 weeks.

3. **Headless server start** (PROVEN-live, ~08:55 UTC) — `setsid bash -c 'nohup herdr server
   > /tmp/.../server-boot.log 2>&1 &' </dev/null`, then `sleep 2; herdr status` → `server:
   status: running`. Directly contradicts `pitfalls-and-traps.md` trap 17 ("You can't start
   the server yourself from a script — it requires an interactive TUI invocation"). The
   server's own boot message even nudges toward the TUI ("did you mean to open the Herdr TUI?
   run `herdr`; you do not need `herdr server`") but does not refuse to run headless.

4. **Pre-existing state surfaced on restart** — `herdr workspace list` / `pane list` /
   `agent list` immediately showed 5 workspaces (`w1`,`w2`,`wF`,`wG`,`wJ`) each with one
   `claude` agent pane, all `idle`, persisted from before the server outage. Confirms herdr
   persists workspace/pane state across a server restart (a "named persistent session"
   concept per `herdr session <subcommand>` in `--help`). None of these were touched.

5. **Lifecycle — spawn** (PROVEN-live) —
   `herdr workspace create --cwd /tmp/gcn68-herdr-probe/lane1 --label gcn68-probe --no-focus`
   → new workspace `wK`, root pane `wK:p1`. `herdr pane run wK:p1 "claude
   --dangerously-skip-permissions"` → Claude booted, hit the one-time "trust this folder"
   prompt, `herdr pane send-keys wK:p1 Enter` cleared it, composer ready
   (`agent_status: idle`) within ~2s.

6. **CLI-shape drift vs. the skill docs** (PROVEN-live / doc-vs-reality mismatch) —
   `herdr agent wait wK:p1 --status idle` → `unknown option: --status`; the flag on this
   build is **`--until`**, not `--status` as used throughout every reference doc and
   `multi-agent-patterns.md`'s recipes. Separately, `herdr agent start <name> --split right
   --no-focus -- <cli>` (the docs' "canonical path", `pane-lifecycle.md`) → `unknown option:
   --split`; the current syntax is `herdr agent start <NAME> --kind <KIND> --pane <ID>`
   (agent-start-into-an-*existing* pane, not spawn-and-split atomically) — confirmed via
   `herdr agent start --help`. Both mismatches were worked around live; documenting them here
   since a scripted Tars driver built against the shipped docs would break on first call
   against this herdr version.

7. **Lifecycle — nonce send/verify/read/end** (PROVEN-live, ~08:56 UTC) —
   `herdr pane run wK:p1 "Write a file named steer-1786870566-32452.txt ... OK ..."` →
   Claude wrote the file in ~4s. Verified on disk: `cat
   /tmp/gcn68-herdr-probe/lane1/steer-1786870566-32452.txt` → `OK`. `herdr pane read wK:p1
   --source recent` showed the full turn transcript ("● Write(...) ⎿ Wrote 1 line ... ● Done").
   `herdr agent get wK:p1` → `agent_status: done`. End: `herdr pane close wK:p1` → `{"type":
   "ok"}`; confirmed gone from `pane list` afterward.

8. **ssh shape — no TTY** (PROVEN-live, ~08:57 UTC) — re-ran a `pane run` + `pane read` cycle
   inside `setsid bash -c '...' </dev/null`; output confirmed `tty` → `not a tty` and
   `HERDR_ENV=` / `HERDR_PANE_ID=` both empty (i.e. genuinely outside any herdr pane, the
   shape `ssh cooper '<cmd>'` would produce), and the send still landed and Claude executed it
   (`Update(steer-....txt)` appended a second line). Raw `herdr` CLI has no TTY dependency —
   it's a pure Unix-socket JSON-RPC client.

9. **Busy-loop mid-task send** (PROVEN-live, ~08:57–08:58 UTC) — sent a genuine 20s Python
   CPU busy-loop as a Bash-tool task (`while time.time()<end: x+=1`) to `wK:p1`. Confirmed
   truly busy 3s in: `pane read` showed `● Running a 20-second busyloop in Python · 3s` and
   `agent get` → `agent_status: working`. Sent a second `pane run` mid-loop (filename fell back
   to `steer-.txt` because the nonce var didn't survive across separate Bash-tool invocations —
   noted, doesn't affect the finding). `pane read` immediately after showed the message queued
   under the busy tool call: `❯ Write a file named steer-.txt ...` with a **"Press up to edit
   queued messages"** footer — i.e. queued, not interleaved, not lost. After the loop finished
   (~15s later), one continuous turn executed both: `Ran 1 shell command` then
   `Write(steer-.txt)` then `● Both done: - Busyloop finished ... - steer-.txt written`.
   Confirms the file: both `steer-1786870566-32452.txt` and `steer-.txt` present on disk.
   Matches the prior `orca terminal send` finding exactly.

10. **Return path** (INSPECTED) — read `waiting-and-async.md` (two wait commands, both
    driver-blocking; `done→idle` has no event at all, must poll) and
    `events-and-subscribe.md` (`events.subscribe` is a long-lived Unix-socket stream, dies
    when the socket closes, one subscription per socket, no persistence, no webhook). Checked
    `herdr notification --help` → `notification show <TITLE>` is a **local desktop toast**
    (position/sound options for a human at the TUI), not a network/webhook primitive. No live
    push test was possible or meaningful — there is nothing herdr can push *to* outside its
    own socket.

11. **Orca coexistence** (PROVEN-live, ~08:59 UTC) — `architecture.md`: herdr is "A Rust
    daemon that multiplexes terminal panes — like tmux — but [not tmux itself]", own socket.
    `which orca tmux` → both exist; `ps aux | grep tmux` showed one unrelated tmux session
    (`rollback-cleaq`, a pre-existing SSH tunnel, untouched). Live test: `herdr pane split
    wK:p1 --direction down --no-focus --cwd /home/gaetan/dev/orca-worktrees/Tars/improvements`
    (this exact GCN-68 worktree, live and in-use by this very orchestrator session) → clean
    split, no error. `herdr agent start orca-coexist --kind claude --pane wK:p2` → registered
    `interactive_ready: true` immediately. Ran a strictly read-only check (`pwd && git status
    --short && git branch --show-current`, explicitly instructed no edits) → clean working
    tree, correct branch (`GaetanCathelain/improvements`), no lock/conflict with the
    concurrently-running orchestrator process in the same tree. Closed via `pane close wK:p2`.

12. **Cross-harness — codex.py fails outside a herdr pane** (PROVEN-live, ~08:59 UTC) —
    `python3 scripts/codex.py start --task "..." --slug gcn68-probe-a --cwd
    /tmp/gcn68-herdr-probe/codex-lane --expect ...` → first attempt rejected the slug
    (`BAD_SLUG`, "codex" is a reserved word — worked as documented). Retried with a clean slug
    → `{"ok": false, "error": {"code": "NAMING_FAILED", "message": "HERDR_PANE_ID is not
    set", ...}, "exit_code": 3}`. Traced to `_core.py:_caller_workspace_id` /
    `name_herdr_tab.py:build_label`, which raise `NamingError("HERDR_PANE_ID is not set")`
    with no override flag exposed by `codex.py`'s CLI. This is a hard blocker for exactly the
    Tars ssh-cooper shape this probe is testing.

13. **Cross-harness — raw herdr↔codex smoke test** (PROVEN-live mechanics / UNTESTED
    completion, ~09:00 UTC) — fell back to raw CLI: `herdr pane split wK:p1 --direction down
    --no-focus --cwd /tmp/gcn68-herdr-probe/codex-lane` → `wK:p3`. `herdr agent start
    codex-raw --kind codex --pane wK:p3` → `interactive_ready: true`, `agent: codex`. First
    `pane run` with the nonce task was silently swallowed (composer still showed Codex's
    ghost placeholder text, not the sent task) — reproduces the documented "spawn-readiness
    race" the skill's `codex.py` exists to solve. Second `pane run` (after a further ~8s
    settle) landed and submitted correctly (composer showed the actual sent text). Codex then
    failed with `stream error: Failed to refresh token: 401 Unauthorized` repeated 5x and
    `■ Failed to refresh token: 401 Unauthorized` — a `codex` CLI auth/credential problem on
    cooper, not a herdr fault. Did not attempt to fix Codex auth (out of scope, no secrets
    touched). `opencode` — `which opencode` → not found; untested, no fallback attempted.

## Cleanup performed

- Closed all 3 probe panes (`wK:p1`, `wK:p2`, `wK:p3`) via `pane close`; workspace `wK`
  auto-removed once its last pane closed (confirmed empty from `workspace list`).
- Confirmed the 5 pre-existing panes/workspaces (`w1`,`w2`,`wF`,`wG`,`wJ`) untouched throughout
  — never addressed by any `pane`/`agent` verb in this probe.
- Stopped the herdr server (`herdr server stop`) to restore the exact pre-probe state (it was
  not running when the probe started); confirmed `herdr status` → `not running` afterward.
- Deleted `/tmp/gcn68-herdr-probe/` entirely (`rm -rf`); confirmed gone.
- No changes made under `$HOME` outside the herdr socket/log directory's normal runtime files
  (server log gained routine entries; no config files edited).
- Never touched Slack, the Tars VM, `secrets/`, or any session/pane not created by this probe.
