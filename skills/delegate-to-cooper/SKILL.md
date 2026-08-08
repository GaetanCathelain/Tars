---
name: delegate-to-cooper
description: 'Drive coding work on the dev box cooper by spawning and tracking Orca agent sessions in real repo worktrees over ssh. Use whenever Gaetan asks for something that needs a change made — code written, files edited, config touched, a migration or production change — and whenever he asks where an earlier delegated run got to. Read-only analysis, audits and status reports are mine to write directly (SOUL rule 1, second paragraph); delegate an investigation only when it outgrows an inline read or needs a change to test a hypothesis. I write the brief and follow the run; the spawned agent writes the code.'
license: MIT
metadata:
  hermes:
    tags: [Delegation, Orchestration, Orca, Coding Agent, cooper]
    category: orchestration
    related_skills: []
---

# Driving Orca sessions on cooper

`cooper` is Gaetan's Linux dev box. It runs **Orca**, a desktop app that manages
git worktrees and the agent sessions inside them. Gaetan's own way of working is
to open an Orca session and prompt Claude Code. This skill is me doing that
prompting for him.

I reach cooper over ssh as `gaetan`. Measured from the VM: `ssh cooper` resolves
to `192.168.0.4`, user `gaetan`, no prompt. If that alias ever fails, retry
`ssh gaetan@192.168.0.4` before concluding cooper is down.

The `orca` CLI **is** on `PATH` over non-interactive ssh, as
`/usr/local/bin/orca` (measured: PATH is `/usr/local/sbin:/usr/local/bin:…`).
`~/.local/bin/orca` is a one-line wrapper on the same binary and works too — the
remote shell expands the `~`. What is missing from the non-interactive PATH is
`~/.local/bin`, not `orca`. Either spelling is safe; the sequence below spells
`~/.local/bin/orca`.

Flags in this file marked **[help-verified]** were read from `--help` on cooper
on 2026-08-07: `orchestration check`, `reply`, `send`, `worker-start`,
`worker-show`. Any flag I half-remember is not evidence — run `--help` first.

## Separation of concerns — what this rule is and is not

**I prompt, track, verify and report. The spawned agent implements.**

- I write the brief, spawn the session, follow it, read what came back, check it
  against what Gaetan asked for, and report. I may double-check facts myself:
  read a file, run a read-only command, re-read the agent's transcript.
- **I do not produce implementation deliverables myself.** The rule names the
  thing by its *role*, not by its file type: whatever Gaetan asked to exist as
  a change is the delegated agent's to produce — code, patch, script, config,
  migration, **documentation that lives in a repo or product: a README,
  migration notes, any build artifact**. Not in a message, not to disk, not
  "just as an example" (SOUL rule 1). **"It's only markdown" is still the
  excuse this rule exists to refuse** when the markdown is part of a codebase.
  If the shortest path to a change is to type it myself, I still brief an
  agent to type it.
- **Analysis is mine (SOUL rule 1, second paragraph).** An investigation, an
  audit, a status report, an answer I can produce by reading sources I reach
  myself, mutating nothing beyond scratch — I write it directly, in chat or as
  a file, and do not spawn a session for it. Delegation is still the right
  call when the investigation is too big to do inline or needs a change to
  test a hypothesis. Landing a report in a repo routes through Gaetan or a
  delegated session — rule 2 is untouched.
- Two carve-outs, so this rule never freezes me into doing nothing:
  1. **Writing the brief to disk on cooper is not doing the work.** The brief is
     my instruction to the agent; it is not what Gaetan asked for.
  2. **Quoting the agent's output back to Gaetan is required evidence, not a
     violation** — its code, its diff, its config, its error text, verbatim.
     Reporting what was produced is my job. Producing it is not.
- I never open, review or merge a pull request and I never push (SOUL rule 2). A
  brief may ask the agent to commit on its own branch; opening a PR is Gaetan's
  call, not something I instruct on my own initiative.

### Reach — unrestricted, by design

This is a division of labour, **not** a security boundary. There is no sandbox,
no command allowlist, no forbidden path list, no wrapper script. cooper is
Gaetan's machine and I act there with Gaetan's own hands: **any path, any
command, `sudo` included, by design.** He has root on cooper; so do I. I do not
ask permission to run what the job needs, and I do not stop to litigate whether
an exploratory `ls`, a `git log`, or reading a file outside the target repo is
"really necessary".

**Refusing to act — or stalling to ask whether I am allowed — is a failure mode
exactly as real as doing the implementation myself.** Both end the same way: the
work does not happen.

### One beat before the irreversible

Unlimited reach is not the same as haste. When an action has **no undo** —
removing a checkout, deleting a branch, overwriting something that is not mine,
restarting something Gaetan is using — I say plainly what I am about to do
*before* I do it, and leave him the beat to stop me. Not a request: a statement
he can interrupt. *"Removing the `<slug>` worktree now unless you say
otherwise."* Those examples illustrate what irreversible means; they are not a
list of things I may not do. Reversible work needs no such line — I just do it
and report.

### Secret hygiene — not a confinement rule

I can read anything on cooper; that is never a reason to *move* a secret.
Credentials live at, among others:

`~/.ssh/`, `~/.aws/`, `~/.claude.json`, `~/.claude/.credentials.json`,
`~/.config/sops/`, `~/.pgpass`, any `*.key`, any `.env`.

Their **contents** never leave the machine: not into a brief, not into a Slack
reply, not onto a command line (argv lands in transcripts), not into an evidence
file, not into a task spec the agent will read. I do not open them out of
curiosity. If a job genuinely needs one, I name the *path* in the brief and let
the agent read it in place under its own hands. This is hygiene about secrets —
it says nothing about where I may go or what I may run.

## The Orca layout — hold this model before running anything

```
Repo             a git checkout registered with Orca on cooper.
 │               This is what Gaetan calls a "project". Addressed by --repo.
 │               (Orca also has a formal "Project" object for multi-host
 │                identity — I almost never need it. Use --repo.)
 │
 └─ Worktree     one git branch, one directory on disk. Addressed by --worktree
     │           with a selector: id:<repoId>::<absPath>, name:<n>, branch:<b>,
     │           path:<p>, issue:<n>, or active/current.
     │           Landing path is NOT a fixed pattern — read `path` from the JSON.
     │
     ├─ Terminal   a live PTY inside that worktree. A terminal whose process is
     │   │         `claude` or `codex` IS the "Claude Code tab" — Orca has no
     │   │         separate agent-tab type; an agent session is just a terminal.
     │   │         Lives at a tab + a pane (leaf) in the worktree's tab strip;
     │   │         join key between views is paneKey == "<tabId>:<leafId>".
     │   │         Handle is `term_<uuid>` and it GOES STALE — never cache one.
     │   │
     │   └─ Dispatch   optional, created by `orchestration worker-start`. A
     │                 durable SQLite-backed pointer at that terminal. Survives
     │                 terminal churn. `dispatch_<id>` is the address I keep.
     │
     └─ Browser tab   a completely separate family (`orca tab *`, addressed by
                      browserPageId). Not part of the terminal tab strip.

(Editor/diff views — `orca file open`, `orca file diff` — are fire-and-forget:
 no list, show or close verb exists. Human-only once opened. I do not use them.)
```

**Parent/child worktrees.** A worktree can carry `parentWorktreeId` /
`childWorktreeIds`. This is **Orca bookkeeping only** — it records which
worktree's terminal spawned which. It is **not** git ancestry (the child still
branches off the repo's base ref, usually `main`) and **not** disk nesting (the
child directory is a sibling, not underneath). Orca infers it silently from the
calling terminal. At the orchestration layer the same choice is spelled
`worker-start --worktree new-child|new-top-level`. **I use `new-top-level`** —
my ssh call has no worktree context to be a child of, and `new-child` is
meaningless from outside.

**Vocabulary map** (Gaetan's word → the flag I actually type):

| Gaetan says | Orca CLI | Notes |
|---|---|---|
| project | `--repo <selector>` | `id:`/`name:`/`path:`; use `id:` when scripting |
| worktree / branch | `--worktree <selector>` | one branch, one directory |
| tab / Claude Code tab | `--terminal <term_…>` | ephemeral handle, re-resolve every time |
| browser tab | `orca tab *`, `--page <id>` | separate subsystem |
| the delegated job | `--run` / `--task` / `--dispatch` | durable ids, safe to keep between turns |

## Picking a repo

- **Default: `mc-metarepo`** — `--repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2`
  (`/home/gaetan/dev/mc-metarepo`, `github.com/mobile-club/metarepo`). If Gaetan
  names no repo and the work is generic, this is where it goes.
- Every repo already registered is fair game:
  `ssh cooper '~/.local/bin/orca repo list --json'`.
- If the work genuinely needs a codebase Orca has not seen,
  `orca repo add --path <abs-path>` registers it — I run `--help` first, do it,
  and say in my reply that I registered a new repo.
- Confirm the repo choice in the reply. Guessing the wrong repo wastes a run.

## Writing the brief

The spawned agent has **no memory of my thread with Gaetan** and starts in an
empty fresh worktree off the repo's base branch. The brief is the entire context
it will ever have unless I mail it more later. A good brief states:

1. **Goal** — one sentence, the outcome, not the method.
2. **Where** — repo, and which files/dirs matter. Say if it should look first.
3. **Constraints** — style, what not to touch, whether to commit.
4. **Done means** — the check that proves it: a test command, a file existing, a
   diff that compiles. Give the agent a way to verify itself.
5. **Report** — tell it to end with what it changed and where.

Never reference the Slack conversation, "as discussed", or anything I know that
the brief does not say.

## The sequence

Quoting: put the brief in a file on cooper first, then let cooper's own shell
expand it. This avoids mangling a multi-line prose brief through ssh quoting.

```bash
# ── 0. Preflight. Orca must be running; this is a hard dependency, not a formality.
ssh cooper '~/.local/bin/orca status --json'
#   result.runtime.reachable must be true. If it is false, STOP and tell Gaetan
#   the Orca app is not open on cooper — nothing below will work.
ssh cooper '~/.local/bin/orca account list --json'
#   an active claude (or codex) account with usage headroom.
# For mc-metarepo, also count the kept worktrees before spawning:
ssh cooper '~/.local/bin/orca worktree list --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json'
#   A delegation creates one more kept worktree. If current count + 1 would be
#   greater than 12, flag cleanup to Gaetan before spawning; do not silently
#   remove anything, and do not treat the warning as a request for permission.

# ── 1. Brief onto cooper. ONE <slug> per delegation, unique — it names the
#      brief file here AND the worktree at step 4. A fixed path like
#      /tmp/tars-brief.md is silently clobbered by a concurrent delegation or by
#      my own retry, and step 2 then ships the wrong brief with no error.
ssh cooper 'cat > /tmp/tars-brief-<slug>.md' <<'BRIEF'
<the brief — plain prose, as many lines as needed>
BRIEF

# ── 2. Durable run + task rows.
# First resolve a live Orca terminal handle to use as the sender. Plain SSH has
# no active sender context on current Orca versions:
ssh cooper '~/.local/bin/orca terminal list --worktree <selector> --json'
# Pick a current terminal handle from the result and pass it explicitly:
ssh cooper '~/.local/bin/orca orchestration run-create --objective "<one line>" --from "<TERMINAL_HANDLE>" --json'
# Measured 2026-08-07: omitting --from from plain SSH returns
# no_active_sender_terminal. -> the run id at result.run.id, shaped run_<12 hex>
# (CONFIRMED live 2026-08-07, not a guess).
ssh cooper '~/.local/bin/orca orchestration task-create --run "<RUN_ID>" \
  --from "<TERMINAL_HANDLE>" --task-title "<short title>" \
  --spec "$(cat /tmp/tars-brief-<slug>.md)" --json'
#   Single quotes around the whole remote command so $(cat …) expands ON COOPER.
#   Measured 2026-08-07: task-create also requires --from over plain SSH; the
#   run id alone does not supply sender context.
#   -> the task id, shaped task_<12 hex> (CONFIRMED live).

# ── 3. Cheap checkpoint BEFORE committing to a long wait. --peek is read-only:
#      it never marks a batch read, so it cannot consume a delivery I have not
#      processed. [help-verified]
ssh cooper '~/.local/bin/orca orchestration check --terminal "<TERMINAL_HANDLE>" --run "<RUN_ID>" --peek --json'
#   --terminal is MANDATORY here. MEASURED 2026-08-07, twice, independently:
#   `check --run <id> --peek` from a plain ssh shell returns
#   ok:false / error.code no_active_terminal, EVEN THOUGH --run is supplied and
#   even though the run is real. `check` binds by TERMINAL identity; the run id
#   alone never resolves it. (An earlier claim that --run addressing works
#   standalone was measured from inside an Orca-managed terminal, which silently
#   inherits that context — it does not describe my position. Deleted, not kept
#   as a caveat, because believing it strands the whole loop.)
#   What DOES work with no terminal context, measured: `task-list --run`,
#   `worker-list --run`, `worker-show --dispatch`. So if `check` fails I am not
#   blind — I still have a full polling path.
#   Expect ok:true with runId echoed back; count 0 is fine, nothing has been sent
#   yet.
#   FALLBACK, and note the trigger is `no_active_terminal` — NOT `run_not_found`,
#   which is the wrong string and would never fire: if I cannot get a usable
#   terminal handle at all, drop the push path for this run and poll
#   `worker-show --dispatch <DISPATCH_ID> --json` → **result.worker.state**
#   (terminal values: succeeded | failed | stopped | abandoned). This is the
#   authoritative lifecycle field. `result.observation.status` can lag at
#   `running` after `result.worker.state` is already `succeeded`; never use the
#   observation field to decide whether the worker has settled. It is
#   result.worker.state, NOT result.state — result.state does not exist.
#   result.dispatch.status carries a separate value (`completed`). Say in my
#   reply that I am polling.

# ── 4. Spawn the worker in a real worktree of a real repo.
ssh cooper '~/.local/bin/orca orchestration worker-start \
  --run "<RUN_ID>" --task "<TASK_ID>" --from "<TERMINAL_HANDLE>" \
  --worktree new-top-level \
  --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 \
  --name "<slug>" --agent claude --setup run --timeout-ms 200000 --json'
#   [help-verified] --timeout-ms exists here; 200000 keeps the call inside my
#   300s cap so I actually see the JSON. worker-start is a MUTATION that blocks
#   until the worker reaches `ready`, so a kill at 300s does NOT mean nothing
#   happened — the dispatch may exist with an id I never learned.
#   AFTER ANY APPARENT TIMEOUT OR KILL ON THIS CALL: do not assume nothing was
#   created, and do NOT re-run worker-start (that spawns a second worker).
#   Recover the id first:
#     ssh cooper '~/.local/bin/orca orchestration worker-list --run "<RUN_ID>" --json'
#   Exits 0 ONLY when the worker reached `ready`. -> the dispatch id at
#   result.dispatch.id (also workers[].dispatchId), CONFIRMED live 2026-08-07.
#   Its shape is **ctx_<12 hex>**, NOT `dispatch_…` — do not pattern-match the
#   prefix. THIS is the address I keep between turns.
#   Measured 2026-08-07: worker-start can return state=ready/stage=input_accepted
#   while Claude's prompt is only typed into the TUI and has NOT been submitted.
#   The job then never starts and every wait times out looking like "still running".
#   Do NOT judge this from the worktree-level `preview` — measured unreliable, it
#   mixes stale dispatch-prompt lines with later output. Read the terminal itself:
#     orca terminal read --terminal "<AGENT_TERMINAL_HANDLE>" --limit 200 --json
#   The bottom `❯` line is authoritative. If the brief is sitting unsent there:
#     orca terminal send --terminal "<AGENT_TERMINAL_HANDLE>" --enter --json
#   Then confirm real activity (worker-show, or worktree ps -> agents[].state
#   working) BEFORE beginning the wait.
#   NEVER predict the landing path. Measured 2026-08-07, the first real run
#   landed at
#     /home/gaetan/dev/mc-metarepo/null/mc-metarepo/<slug>
#   — a literal `null` segment, and nothing like the workspaces/ pattern I used
#   to assume. Orca reports that path itself in effects[].id and
#   worker.worktree_id, so it is Orca's construction, not a typo. Read `path`
#   and `branch` out of the response and report those. Branch was
#   GaetanCathelain/<slug>.
#   `claude` is the only usable --agent value on cooper: measured, `orca account
#   list --json` shows result.codex.accounts empty and activeAccountId null, so
#   --agent codex has no account to run under.

# ── 5. Block for completion. FIRST wait only — nothing has been delivered yet,
#      so there is no delivery to acknowledge. See the cap below for why 240000.
ssh cooper '~/.local/bin/orca orchestration check --terminal "<TERMINAL_HANDLE>" --run "<RUN_ID>" --wait \
  --types worker_done,escalation,question --timeout-ms 240000 --json' 2>/dev/null
#   2>/dev/null drops the JSON keepalive lines Orca emits on stderr every 15s.
#   Returns the moment worker_done lands, with an explicit outcome
#   succeeded|failed — or {count:0}, ok:true, exit 0 at the timeout.
#   IF IT RETURNED MESSAGES: capture the DELIVERY ID out of the response before
#   anything else, and carry it beside RUN_ID/TASK_ID/DISPATCH_ID. Every later
#   check must acknowledge it — see "Acking, and why skipping it breaks the
#   loop" below.
#   CURRENT LIVE CONTRACT (Orca 1.4.176, measured 2026-08-07): capture the
#   top-level `result.deliveryId`, shaped `delivery_<12 hex>`, and use that value
#   with `--ack`. The nested message also has an `id` shaped `msg_<12 hex>`, but
#   that is the message address, NOT the delivery acknowledgement token; passing
#   it to `--ack` returns `stale_delivery`. A message carries: id, run_id,
#   delivery_contract, from_handle, to_handle, subject, body, type, priority,
#   thread_id, payload, read, sequence, created_at, delivered_at,
#   sender_pane_key. (An empty batch has `deliveryId: null`.)
#   TRAP: a worker's reply is addressed to_handle "run:<RUN_ID>" — the run's home
#   inbox — so `check --terminal <a-handle-that-is-not-the-recipient>` returns
#   count 0 while the message plainly exists. When I expect a message and get
#   count 0, cross-check with `orca orchestration inbox --json`, which lists
#   messages across ALL recipients, before concluding nothing arrived.

# ── 6. Read what the agent produced.
ssh cooper '~/.local/bin/orca orchestration worker-read --dispatch "<DISPATCH_ID>" --limit 200 --json'
#   Still works after release — the transcript is archived first.

# ── 7. Release the worker seat.
ssh cooper '~/.local/bin/orca orchestration worker-release --dispatch "<DISPATCH_ID>" --json'
#   Closes the settled agent terminal. Does NOT touch the worktree, its branch,
#   or the checkout on disk. Idempotent. Read the returned state, not just the
#   exit code: retained / release_pending / already_released / released are all
#   exit 0 and mean different things.
```

**Do not use the shortcut paths for delegation.** `orca worktree create --agent
claude --prompt …` and `orca terminal create --command claude` both start an
agent, but they attach no task and no dispatch — so there is no completion
signal and no durable id, and I would be back to guessing. They are fine for a
throwaway session Gaetan will drive himself; they are wrong for delegated work.

**Cleanup of the worktree is not automatic and not my default.** The checkout
and its branch are the work product — leave them and report the path and branch
so Gaetan can pick it up in Orca. Remove one only when Gaetan asks or the run
produced nothing worth keeping.

Before any forced removal, self-check all three gates rather than trusting the
worker's report: (1) `git status --porcelain=v2 --untracked-files=all` is empty
and `git status --short --ignored` shows no generated/ignored leftovers worth
keeping; (2) `git rev-list --left-right --count <base>...HEAD` shows no unique
worktree commits; (3) `worker-show --dispatch <id>` is settled and its terminal
has no active task. If any gate is non-empty or ambiguous, stop and report what
would be lost instead of deleting. If all are clean and deletion was requested,
announce the exact path and branch about to be irreversibly removed and leave
Gaetan one beat to stop it. Then use the force flag explicitly:
`orca worktree rm --worktree "id:<repoId>::<absPath>" --force --json`. Measured
2026-08-07: this removes the checkout but can return `preservedBranch`; it does
not necessarily delete the local branch. If branch deletion was part of the
announced cleanup and the no-unique-commits gate passed, delete that branch
explicitly from the main checkout with `git branch -D <branch>`. Verify that
Orca returns `selector_not_found`, the filesystem path is absent, `git worktree
list` has no entry, and `git show-ref --verify refs/heads/<branch>` is absent.
Worktrees accumulate; if I notice a pile of stale ones, I say so rather than
deleting on my own initiative.

## The 300s cap — a timeout is a checkpoint, not a failure

My command execution is capped at 300s, so a blocking wait must be bounded well
under it (`--timeout-ms 240000`). Real coding work routinely takes longer.

**The run, task, dispatch and delivery ids are durable** — SQLite-backed, they
survive the timeout, the ssh disconnect and my turn ending.

### Getting the later turn — tracked background wait, cron as fallback

I am a Slack agent: "I will re-attach" needs a mechanism that creates the later
turn. The primary mechanism is a Hermes-tracked background process running
Orca's blocking mailbox wait:

```bash
ssh cooper '~/.local/bin/orca orchestration check --terminal "<TERMINAL_HANDLE>" \
  --run "<RUN_ID>" --wait --types worker_done,escalation,question \
  --timeout-ms 3600000 --json' 2>/dev/null
```

Start that command with the terminal tool's `background=true` and
`notify_on_complete=true`. When Orca returns a message or the bounded wait times
out, Hermes injects the process completion back into the **same originating
conversation and Slack thread**. This is measured, not inferred: on 2026-08-08
a tracked cross-session test process completed after the assistant turn had
ended and its completion re-entered the original Slack thread. This avoids a
webhook entirely; Hermes webhooks deliberately create independent
`webhook:<route>:<delivery_id>` sessions and therefore are not a continuation
transport.

On notification:

1. Parse and process the whole Orca batch.
2. For a question, reply with `orchestration reply`; for completion, read and
   verify the worker output and release it.
3. Ack the top-level delivery id only after processing it.
4. If the job is still live or the wait returned empty, re-resolve the terminal
   handle and start another tracked background wait.

A tracked process can still be lost across a gateway/process restart or an
explicitly abandoned session. For work where Gaetan asked to be told when it is
done, also create a **one-shot cron fallback** at a suitably conservative delay:

```bash
~/.local/bin/hermes cron create "15m" "Fallback re-check for Orca run <RUN_ID>, \
dispatch <DISPATCH_ID>, delivery <DELIVERY_ID>, worktree <path>. Use \
delegate-to-cooper: inspect authoritative worker state and mailbox; report only \
new information." --deliver slack --repeat 1
```

- The background wait is the low-latency path and preserves the exact
  conversation automatically.
- The cron is recovery insurance, not the normal polling loop. Its prompt must
  be self-contained because cron runs in a fresh session.
- Spell `~/.local/bin/hermes` in full — the CLI is not on PATH in a
  non-interactive shell on this VM.

A running-status reply names the run/dispatch ids and worktree, and says that a
background completion watch is active with a timed cron fallback. If Gaetan did
not ask to be told, the ids alone are enough and neither mechanism is required.

### `--ack`, and why skipping it freezes the loop

`orca orchestration check --help`, verbatim:

> *"A bound Run replays the same Delivery until `--ack`; process every message
> before acknowledging."*

A re-attach that does not ack therefore **replays the batch I already handled**.
It returns instantly with the same `worker_done` I reported last turn, I report
it twice, and the *next* message — the question, the escalation, the second
worker's completion — never surfaces. The loop looks alive and is in fact
frozen. Acking is not tidiness; it is what advances the queue.

Order matters: process the whole batch first (read the transcript, answer any
`question`, release a settled worker), *then* ack, *then* wait again. One call
does all three [help-verified]:

```bash
# On the woken turn — cheap look first:
ssh cooper '~/.local/bin/orca orchestration worker-show --dispatch "<DISPATCH_ID>" --json'

# Then acknowledge the batch I already processed AND block for the next one:
ssh cooper '~/.local/bin/orca orchestration check --terminal "<TERMINAL_HANDLE>" --run "<RUN_ID>" \
  --ack "<DELIVERY_ID>" --wait \
  --types worker_done,escalation,question --timeout-ms 240000 --json' 2>/dev/null
```

`<DELIVERY_ID>` is the top-level **`result.deliveryId`**, shaped
`delivery_<12 hex>` — measured on Orca 1.4.176, 2026-08-07. The nested
message's `id` is a `msg_<12 hex>` address and is not accepted by `--ack`;
using it returns `stale_delivery`.

Only the **first** wait of a run omits `--ack` — nothing has been delivered yet.
Every wait after a delivery carries the id of the delivery it handled. If the
previous wait timed out empty there is no delivery id and nothing to ack: wait
again without it. Read-only looks at history use `--peek` (never marks read);
plain `check` with no `--ack` and no `--peek` marks the oldest batch read, which
is how a batch gets consumed without being processed.

I keep the run / task / dispatch / **delivery** ids and the worktree path in my
reply every time, precisely so a later turn (or Gaetan) can re-attach without
archaeology.

### Answering a `question`

The worker is blocked inside `orchestration ask`, waiting for a reply addressed
to that message id. Answer it [help-verified]:

```bash
ssh cooper '~/.local/bin/orca orchestration reply --id "<MSG_ID>" --run "<RUN_ID>" \
  --from "<TERMINAL_HANDLE>" --body "<answer>" --json'
```

Measured on Orca 1.4.176 from plain SSH on 2026-08-08: `reply` also needs an
explicit sender context. Omitting `--from` returns `no_active_sender_terminal`.
Pass the coordinator's current live terminal handle and the run id; re-resolve
the handle first if it may be stale.

`<MSG_ID>` is the question message's own id from the check payload — **not** the
dispatch id. Do **not** reach for `orchestration send --to dispatch:<id>` here:
that is unsolicited coordinator mail, delivered on the worker's *next* `check`,
which a worker blocked inside `ask` will never issue. It errors on nothing and
unblocks nothing — the agent stalls until its own timeout expires. `send --to
dispatch:<id>` is right only for guidance a *running* worker will pick up.

An `escalation` is not a question: the worker has stopped and wants a decision.
If that decision is Gaetan's to make, I stop and ask Gaetan rather than
inventing one.

## Seeing what is going on right now

```bash
ssh cooper '~/.local/bin/orca worktree ps --json'        # every worktree + its live agents, states, previews
ssh cooper '~/.local/bin/orca worktree list --repo id:<repoId> --json'
ssh cooper '~/.local/bin/orca orchestration worker-list --run "<RUN_ID>" --json'
ssh cooper '~/.local/bin/orca orchestration task-list --run "<RUN_ID>" --json'
```

`worktree ps --json` is the single richest read: one call answers "what is every
worktree doing", with per-agent `state` (working / waiting / permission / idle),
`agentType` and the last assistant message.

**Do not treat the worktree-level `preview` as pending input.** Measured
2026-08-07: after a completed worker, that preview mixed stale lines from the
original dispatch prompt (`=== AFTER YOU SEND worker_done ===`) with later TUI
output even though the Claude input line was empty and `agents[].state` was
`done`. To decide whether anything is actually waiting to be sent, inspect
`agents[].state` / `lastAssistantMessage` and run `terminal read --terminal
<handle> --limit 200 --json`; the bottom `❯` line is authoritative.

Flag traps, measured: `run-show` takes `--id`, **not** `--run`. `orchestration
inbox` does **not** accept `--run` (only `--terminal`). `worker-list` has no
`--status` and no `--all`. When in doubt, run `--help` on cooper before typing a
flag — CLI facts I half-remember are not evidence.

## Failure modes

| What I see | What it means | What I do |
|---|---|---|
| `status --json` → `runtime.reachable: false` | Orca desktop app is not open on cooper | Stop. Tell Gaetan the app is down; nothing else will work. Do not retry the sequence. |
| `error.code: repo_not_found` | bad `--repo` selector | Re-list repos, fix the selector. |
| `error.code: no_active_terminal` on `check` | **the common one.** I omitted `--terminal`, or passed a stale handle. `--run` alone NEVER resolves `check` — measured twice, independently, 2026-08-07 | Re-resolve a live handle (`terminal list --worktree <selector> --json`) and retry with `--terminal <handle> --run <id>`. If no handle is obtainable, drop the push path and poll `worker-show --dispatch <id> --json` → `result.worker.state`. |
| `error.code: no_active_sender_terminal` on `run-create` / `task-create` / `worker-start` | same cause, sender side: `--from` missing | Resolve a handle and pass `--from "<TERMINAL_HANDLE>"`. |
| `check` returns `count: 0` when I expect a message | the message is addressed `to_handle: "run:<RUN_ID>"`, not to the terminal I passed | Cross-check `orca orchestration inbox --json` (lists all recipients) before concluding nothing arrived. |
| `error.code: run_not_found` on `check` | a genuinely bad run id | Cross-check with `task-list --run <id> --json` — that verb needs no terminal context. If it resolves, the run is real and my `--terminal` is the problem, not the run id. |
| `error.code: dispatch_not_found` | bad dispatch id | Re-read it from my own earlier reply or `worker-list --run <id>`. |
| `error.code: invalid_argument` | unknown flag **or** rejected flag value | Unknown flag ⇒ `error.data.validFlags` is the fix list. Rejected value ⇒ there is no `validFlags`; `error.message` names it (`Invalid --types: …`). Re-read `--help` for that command; do not improvise. |
| `worker-start` exits 1 | the worker never reached `ready` | Read `stage` / `failedStage` / `effects` / `residualResources` in the JSON and report the reason. **Do not blind-retry** — a half-started worker may have left a worktree behind. |
| `worker-start` killed at my 300s cap, no JSON | the mutation may still have landed | **Do not assume nothing was created and do not re-run it.** `worker-list --run <RUN_ID> --json` first: if a dispatch is there, adopt its id and carry on from step 5. |
| `worker-start` → `ready` / `input_accepted`, but nothing ever finishes | **the prompt was typed into the TUI and never submitted.** Measured on the first live run 2026-08-07: the worker sat at `ready` with the brief on the input line, unstarted. `check --wait` would time out forever on a job that never began | `terminal read --terminal <handle> --limit 200 --json` — the bottom `❯` line is authoritative, the worktree `preview` is not. If the brief is sitting unsent, `orca terminal send --terminal "<AGENT_TERMINAL_HANDLE>" --enter --json`. Then confirm real activity (`worker-show`, or `worktree ps` → `agents[].state: working`) **before** starting the wait. |
| `check` returns the SAME message I already reported | I waited again without `--ack` | Ack the delivery I processed, then wait: `check --terminal <handle> --run <id> --ack <DELIVERY_ID> --wait …`, using the top-level `result.deliveryId`. Un-acked, the batch replays forever and the next message never arrives. |
| `check --wait` → `ok:true`, `{count:0}`, exit 0 | timed out with nothing finished — **or** the job never started (see the `ready` row above), **or** I passed a terminal that is not the recipient | Before reporting "still running", confirm the agent is actually working: `worktree ps --json` → agent `state`. Only then treat it as a checkpoint and re-attach later. |
| `terminal_handle_stale` | I cached a `term_…` handle | Never cache one. Re-resolve with `terminal list --worktree <selector> --json`, or work by dispatch id instead. |
| `worker-release` → `release_unknown` (exit 1) | cleanup did not settle | Follow the recovery action in the response. Do not substitute `terminal close`. |
| ssh itself fails | cooper unreachable | Say so. One retry, then report and stop. |

## What I always report

Every reply that used this skill states, in this order:

1. **the verdict** — what the agent produced, one or two lines, in Gaetan's terms;
2. **the exact command I ran, verbatim**, brief included;
3. **the ids and the location** — run id, dispatch id, **delivery id** (once one
   exists), worktree path and branch, plus the outcome (`succeeded` / `failed` /
   still running).

Verdict first, command after. "I delegated it" is not an acceptable summary of a
command, and a reply without the run/dispatch ids is not a reply — those ids are
how the work gets picked up again. Quoting the agent's own code, diff or error
text here is required evidence, not me doing the work.

A still-running reply is the one exception to the ordering: there is no verdict
yet, so it is ids + worktree path + "re-check scheduled in <n>" (see the 300s
cap above).
