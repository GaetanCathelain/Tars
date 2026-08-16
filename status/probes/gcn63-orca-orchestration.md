# GCN-63/67 — Orca orchestration surface as a Tars→session transport

Question: could Orca itself carry messages from Tars (Hermes agent, remote VM,
reaching cooper over ssh as uid 1000) into running Claude Code sessions?
Run on cooper, Orca CLI `/home/gaetan/.local/bin/orca`, app version 1.4.182,
2026-08-16.

## Verdict

**Yes, `orca terminal send` works as a transport and is viable — but it is a
different mechanism from the ledger's, with a different fragility shape, not a
strict upgrade.** It trades "reverse-engineered wire protocol, session-to-session,
no middleware" for "official CLI verb, but hard-dependent on the Orca desktop
app/runtime being alive, and scoped to terminals Orca itself is tracking."

| Axis | `orca terminal send` (this probe) | Ledger's proven mechanisms (§1.1, §4.5) |
|---|---|---|
| Delivery confirmed live | YES — nonce files written, `PROVEN-live` this session | YES — PROVEN-live, 8/8 configs |
| Arrives as clean user turn | YES, with `--enter` | YES (SendMessage/socket) |
| Needs separate `--enter` | Text-only send lands in input box unsent; `--text ... --enter` in one call submits it — confirmed both ways | N/A (socket write is inherently one shot) |
| Send latency (CLI call itself) | **<1s**, synchronous `bytesWritten` ack | `claude -p` relay 4.7–17.4s; socket write sub-ms; channel drop ~0 |
| Busy-session semantics | Queued cleanly, zero corruption of the running turn, auto-processed after (measured) | Not directly comparable — ledger's mechanisms all target **idle** receivers only |
| Addressing | Runtime-issued `--terminal` handle; **not stable across restart** (confirmed: new handle after close+recreate in same worktree) — re-discover via stable `--worktree path:<abs>` selector + `terminal list` | Derived session name, **also not stable across restart** (§4.5); cwd not usable as address either |
| Official vs reverse-engineered | Official, documented, versioned CLI verb (`--help` on every subcommand) | Reverse-engineered wire format (socket JSON frames, `.key` file hashing, `procStart` matching) — version-shaped, breaks on Claude Code internals changing (§4.1) |
| Cross-harness (Codex, OpenCode) | **Pluggable by construction** — it is raw pty keystroke injection, harness-agnostic. NOT separately tested this probe (would need a live Codex/OpenCode pane); the Claude-Code-specific part observed (clean input-box queueing while busy, trust-dialog clearing) is a property of Claude Code's TUI, not of Orca — a different TUI could interleave, ignore, or reject input differently. `orca terminal wait --for tui-idle` also likely depends on Orca's own per-harness idle-detection heuristics, unverified for Codex/OpenCode. | Claude-Code-specific by construction — session sockets, `SendMessage` tool, `ListAgents` discovery all only exist inside Claude Code's own process; **zero** portability to Codex/OpenCode |
| Delivery confirmation semantics | `{accepted:true, bytesWritten:N}` synchronous — confirms bytes reached the pty, **not** that the harness processed it as an instruction | `success:true` means "enqueued at a live peer", not delivered either (§4.5) — same class of weak ack; ledger's async Delivered/Held/Refused notice is "structurally discarded" for a throwaway sender anyway |
| Return channel | **None** — must poll `orca terminal read` (optionally after `orca terminal wait --for tui-idle`, which blocks the caller) | None either for the lean relay; the resident go-bridge peer (§1.1) *can* receive pushed replies — genuinely bidirectional, something Orca's terminal channel cannot do |
| Single point of failure | **The Orca desktop app + runtime must be running** (`orca status` showed `app.running:true`, `runtime.reachable:true`) — if the GUI app is down, the whole transport is unreachable regardless of whether the target Claude Code session is alive | None — socket/SendMessage work session-to-session with no middleware app required |
| Non-interactive / no-TTY | Confirmed — every command in this probe ran via the Bash tool with `tty` reporting "not a tty"; `orca` itself needs no controlling terminal | N/A, ledger mechanisms are also non-interactive |
| Runs over ssh from Tars | Not directly tested (ssh localhost hit an unrelated host-key mismatch in `~/.ssh/known_hosts`, not touched per hard rules) — but since every orca call in this probe was already non-interactive/no-TTY, an ssh-exec of the same `orca terminal send …` command is architecturally the same shape and should work | Ledger mechanisms are same-uid, same-box; Tars reaching cooper over ssh as uid 1000 satisfies that already |

**What Orca gives that the ledger's mechanisms don't:** an official, stable,
documented verb; a genuine orchestration surface beyond raw send (see below) with
mailboxes, tasks, gates, dispatch; and — the real prize — pluggability across
**any** TUI harness in a pane (Codex, OpenCode, raw shell), because it operates at
the pty/keystroke layer, not inside Claude Code's own IPC.

**What it lacks that the ledger's mechanisms have:** a push-based return channel
(the go-bridge peer can receive pushed replies; Orca's terminal channel cannot),
independence from a running GUI app/runtime, and the ability to address a Claude
Code session Orca did not itself spawn/track.

**Recommendation for GCN-67:** worth keeping on the table specifically for its
cross-harness angle if Tars ever needs to drive Codex/OpenCode panes, not as a
wholesale replacement for the proven Claude-Code-native mechanisms in §1.1 of the
ledger. Both require the Orca app to be alive; that dependency should be weighed
against the ledger mechanisms' independence from any middleware process.

---

## 1. Surface map

`orca --help`, `--help` on relevant subcommands, `orc-tab` (found to be a thin
shell launcher wrapping `orca` calls, not a distinct API — see below).

### Top-level command groups
Startup, Diagnostics, Agent Discovery, Accounts, Skills, Environments,
Environment Recipes, **Automations** (scheduled agent runs: list/show/create/
edit/remove/run/runs), Projects, **Repos**, **Worktrees**, Files, **Terminals**,
**Orchestration**, Computer Use, Linear, Mobile Emulator, Browser Automation.

### Orchestration group — kanban/queue/messaging-shaped, in full
```
orchestration run-create        Create and bind a lightweight orchestration Run
orchestration run-use           Bind this coordinator terminal to an existing Run
orchestration run-current       Show this terminal's bound Run
orchestration run-list          List lightweight orchestration Runs
orchestration run-show          Show one lightweight orchestration Run
orchestration send              Send an inter-agent message
orchestration check             Check the bound Run mailbox
orchestration ask               Ask the coordinator a blocking question
orchestration reply             Reply to a message
orchestration inbox             Show all messages across recipients
orchestration task-create       Create an orchestration task
orchestration task-list         List orchestration tasks
orchestration task-update       Update a task status
orchestration dispatch          Dispatch a task to a terminal
orchestration dispatch-show     Show dispatch context for a task
orchestration worker-start      Start a supervised worker locally or remote
orchestration worker-show       Inspect one supervised worker
orchestration worker-read       Read bounded output from one supervised worker
orchestration worker-stop       Fence one Dispatch; stop only its supervised worker
orchestration worker-abandon    Fence an uncertain worker without claiming it stopped
orchestration worker-release    Release a settled worker's terminal, archive output
orchestration worker-retain     Keep a worker terminal live for debugging
orchestration worker-list       Report worker terminal resource accounting
orchestration coordinator-start Start the legacy automatic coordinator loop
orchestration coordinator-stop  Stop the legacy automatic coordinator loop
orchestration gate-create       Create a decision gate blocking a task
orchestration gate-resolve      Resolve a pending decision gate
orchestration gate-list         List decision gates
orchestration reset             Reset orchestration state
```
This is a real mailbox/task/gate system — a Run is "a namespace and home inbox,
[it] never schedules or places workers" (per `--help`). `orchestration send`
supports `--to run:id|dispatch:id|legacy_handle`, priorities, threads, JSON
payloads, task/dispatch linkage — structurally comparable to the ledger's
SendMessage tool. **Not exercised live this probe** (out of scope per the task's
explicit method — it says evaluate `terminal send`) — flagged here because it is
the closest thing to a first-class "messaging" verb Orca has, and a stronger
candidate than raw `terminal send` if Tars could itself hold a Run-bound handle.
The catch: `orchestration check`/`inbox` are **pull-based** — a receiving session
has to actively poll its mailbox, which an ordinary Claude Code session run by a
human/Tars-relay is not doing unless explicitly instructed to loop on it. Raw
`terminal send` pushes text straight into the pty with no polling required on the
receiver's part, which is why it — not `orchestration send` — is the more direct
analog to "Tars pokes a running session."

### Terminal group (the transport under test)
```
terminal list      terminal show      terminal read     terminal send
terminal wait      terminal stop      terminal create   terminal rename
terminal split      terminal switch    terminal focus    terminal close
```

### `orc-tab`
`/home/gaetan/.local/bin/orc-tab` is a POSIX shell script (`launch.sh`), not a
distinct binary — it wraps `orca terminal create`/`orca terminal send` to open a
new tab seeded with a handoff brief or starting prompt, with a `-w` flag for
hub-and-spoke spawns into another worktree. It is a convenience layer on top of
exactly the same `orca terminal` verbs tested below, not a separate transport.

---

## 2. Live evidence (`orca terminal send`)

Setup (per the required ledger recipe, never under `$HOME`):
```
mkdir -p /tmp/orca-scratch-gcn67 && git init -q && ... commit
orca repo add --path /tmp/orca-scratch-gcn67 --json      # repo id 3fff7a60-...
orca terminal create --worktree path:/tmp/orca-scratch-gcn67 \
  --command "claude" --title "gcn67-probe" --json         # handle term_5629c61a-...
```

### 2a. Trust-dialog clear + idle send
`orca terminal read` immediately showed Claude Code's folder-trust prompt
("Quick safety check… ❯ 1. Yes, I trust this folder"). `orca terminal send
--terminal <h> --enter --json` cleared it in one call (`bytesWritten:1`); the
next read showed a normal idle prompt.

**Text without `--enter`**: sent `write a file named steer-a7f3k9.txt containing
just the word done` with no `--enter`. Read-back showed it sitting in the input
box (`❯ write a file named steer-a7f3k9.txt …`) — **not submitted**. A separate
`orca terminal send --terminal <h> --enter --json` call submitted it; 4s later
the read showed `Write(steer-a7f3k9.txt)` / `Wrote 1 line` / `Done. Churned for
3s`. File verified on disk: `/tmp/orca-scratch-gcn67/steer-a7f3k9.txt` = `done`.
**PROVEN-live.**

**Combined `--text "..." --enter` in one call**: sent + submitted a second nonce
(`steer-b2q8mz.txt`) in a single `terminal send` invocation (`bytesWritten:66`).
File appeared within ~4s, confirming the combined form works and is the leaner
recipe (one round trip instead of two). **PROVEN-live.**

Latency: the `terminal send` CLI call itself returns synchronously in well under
1s (`bytesWritten` ack); end-to-end idle-send-to-file-written was ~3–4s, i.e. the
model's own turn time for a trivial write, not transport overhead.

### 2b. Busy-session semantics (mid-task send)
Instructed the session to run a genuine foreground CPU busy-loop via its own
Bash tool (`python3 -c '...while time.time()-s<25: x+=1...'`, no `&`, no `sleep`
— confirmed non-backgrounded by the pane showing "(ctrl+b to run in background)"
mid-execution, i.e. Claude Code offered but did not apply auto-backgrounding).

At T+13s into the 25s loop, sent a second instruction with nonce `midbusy77`
(`orca terminal send --text "MIDTASK PROBE midbusy77: …" --enter --json`,
`bytesWritten:111`, accepted synchronously). The pane, read <1s later, showed:
```
✽ Manifesting… (14s · ↓ 160 tokens)
❯ MIDTASK PROBE midbusy77: when you finish the current busy loop, also write steer-midbusy77.txt containing done
❯ Press up to edit queued messages
```
i.e. Claude Code's own input-box **queued** the message — it did not interleave
into or corrupt the running turn's output, and did not get lost. 20s later (loop
had reached its full 25s), the pane showed:
```
Ran 1 shell command
❯ MIDTASK PROBE midbusy77: …
● The loop ran the full 25s (busylooped 100650477). Now the mid-task request:
● Write(steer-midbusy77.txt)
  ⎿ Wrote 1 line to steer-midbusy77.txt
     1 done
● Both done: busy-loop ran the full 25s, and steer-midbusy77.txt written.
```
File verified on disk (`steer-midbusy77.txt` = `done`). **PROVEN-live**: the busy
loop ran its full, uninterrupted 25s (confirmed by the printed loop count and
Claude's own "ran the full 25s" statement), the mid-task send neither aborted nor
corrupted it, and the queued instruction was auto-processed as the very next turn
once the busy turn ended. This queueing behavior is a property of Claude Code's
own TUI input handling, not something `orca terminal send` itself controls — it
only injects raw keystrokes; the receiving harness decides whether to queue,
interleave, or drop them. **Not verified for Codex/OpenCode this probe.**

### 2c. Addressing and non-interactive shape
- `orca terminal show --terminal <h>` returns `ptyId`, `worktreeId`,
  `worktreePath`, `branch`, `title`, `connected`, `writable`, `lastOutputAt` — no
  stable cross-restart identifier beyond the worktree path.
- **Handle stability test**: closed the pane (`orca terminal close --terminal
  <h1> --json` → `ptyKilled:true`), then created a fresh terminal in the *same*
  worktree (`orca terminal create --worktree path:/tmp/orca-scratch-gcn67 …`) →
  new handle `term_820d42cd-...`, confirming **the handle does not survive a
  session restart**; the durable address is `--worktree path:<abs>` (or
  `id:`/`branch:`), from which `orca terminal list --worktree <selector>`
  re-discovers the live handle.
- **Non-interactive**: every command in this entire probe ran through this
  session's own Bash tool, which reports `tty` → "not a tty". `orca` required no
  controlling terminal for any call (repo add, terminal create/send/read/
  show/close, worktree rm). This directly demonstrates the "simulate the remote
  shape" requirement — an ssh-exec of the same commands is architecturally
  identical.
- **ssh localhost**: attempted `ssh -o BatchMode=yes localhost 'echo ok'` to more
  literally simulate Tars's ssh-in shape; it failed on an unrelated stale
  `~/.ssh/known_hosts` host-key mismatch for `localhost` (pre-existing, not
  caused by this probe). Per the hard rule against touching `~/.ssh/`, did not
  run `ssh-keygen -R localhost` to fix it. This is an environment fact, not a
  finding about Orca — the no-TTY point above already establishes non-interactive
  viability without needing the ssh hop specifically.

### Known wart confirmed
- Second `orca terminal close` call (per the ledger's "may need two calls" note):
  in this run, the **first** close already fully removed the pane/tab
  (`terminal list` returned empty immediately after); a follow-up `--tab` close
  on the same handle returned `runtime_error: tab_not_found`. On the *second*
  terminal, two closes in a row both returned `ok:true, ptyKilled:true`
  (idempotent that time). Timing/state-dependent, not deterministic — noted
  honestly rather than asserted as "always needs two."
- `orca worktree rm --worktree path:/tmp/orca-scratch-gcn67 --force` refused:
  `"Refusing to delete protected worktree path"` — this was the repo's *main*
  worktree (registered via `orca repo add`, not `orca worktree create`), which
  Orca protects from removal. **No `orca repo rm` verb exists** — the repo
  registration for `/tmp/orca-scratch-gcn67` remains dangling in `orca repo list`
  after this probe. Flagged, not removed, per the task's instruction.

---

## 3. Cleanup performed
- Closed both terminals created (`term_5629c61a-...`, `term_820d42cd-...`) —
  confirmed via `orca terminal list --worktree path:/tmp/orca-scratch-gcn67`
  returning empty.
- `orca worktree rm --force` attempted, refused as a protected main worktree
  (expected — it was a repo-add main worktree, not an `orca worktree create`
  worktree).
- No stray processes found (`ps aux | grep orca-scratch-gcn67` empty).
- `rm -rf /tmp/orca-scratch-gcn67` — confirmed removed.
- **Dangling registration left behind**: repo id `3fff7a60-9ae2-427c-b980-614b44ee0f89`
  (`/tmp/orca-scratch-gcn67`) remains in `orca repo list` — no removal verb
  exists. This matches the documented known wart; not something this probe could
  fix.
- Nothing touched: Slack, the Tars VM, `secrets/`, other live sessions,
  `~/.ssh/`.

## Confidence tags
Following the ledger's convention: PROVEN-live for everything under §2 (nonce
files, timestamps, pane captures all observed out-of-band); INSPECTED for the
surface map in §1 (read from `--help` output, not exercised beyond `terminal
send`/`show`/`close`/`list`/`create`); INFERRED for the cross-harness
(Codex/OpenCode) pluggability claim and the ssh-exec equivalence claim — neither
was directly exercised this probe, both follow from the no-TTY evidence and the
pty-level (not Claude-Code-internal) nature of `terminal send`.
