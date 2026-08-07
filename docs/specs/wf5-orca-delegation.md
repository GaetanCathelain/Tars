# WF5 — Orca delegation playbook v1

> **SUPERSEDED 2026-08-07 by v2 — see "[WF5 — Orca delegation v2
> (authoritative)](#wf5--orca-delegation-v2-authoritative)" at the bottom of this
> file.** v1 is kept for the record, not for execution: its mechanism
> (`delegate.sh` wrapper, deny-rule `.claude/settings.json`, the
> `~/orca/workspaces/tars-delegated/` sandbox) is dropped, not
> deprecated-with-fallback. Read v2 before running anything.

Status when written: **implemented and verified 2026-08-07**.
Evidence: `status/probes/wf5/orca-implement.md`. Recon it builds on:
`status/probes/wf5/orca-recon.md`, `status/probes/wf4/p15-delegation-cooper.md`.

## What v1 does

Gaetan asks Tars for something in Slack that needs code written, a file
produced or a codebase looked at. Tars writes a **brief in prose**, pipes it
over ssh to a coding agent on cooper, and reports the agent's answer plus the
exact command it ran. Tars never writes the code (SOUL rule 1); this is SOUL
rule 3 — "work that needs code written is delegated to a coding agent" — made
executable.

```
Slack → Tars (VM 192.168.0.9) → ssh cooper → delegate.sh → claude -p → stdout → Tars → Slack
```

## Mechanism chosen, and why not the others

The recon ranked three candidates. v1 takes **none of them verbatim**: it takes
the ssh transport already proven in probe 15 and points it at a headless
coding agent in a dedicated workspace, rather than at the Orca session layer.

| Candidate | Verdict |
|---|---|
| `orca terminal create` / `terminal send` | Rejected for v1. Needs the Orca desktop app up, handles go stale (`terminal_handle_stale`), and has no completion signal — Tars would hand-roll polling. Three failure modes bought nothing v1 needs. |
| `orca orchestration` run/mailbox/DAG | Rejected for v1. Durable coordinator state (run id, task id, dispatch id) across ssh calls that share no session; a lost Tars turn orphans tasks. |
| hermes `SKILL.md` | **Adopted** — as the carrier, exactly as the recon framed it. Not a transport: it is how Tars learns the playbook. |
| plain ssh + headless `claude -p` in a dedicated dir | **Adopted** as the transport. Synchronous, no session state, no app dependency, result returns on stdout in the same tool call. |

`orca` is still the right layer for long supervised multi-agent work. v1 does
not need it, so v1 does not carry its failure modes.

## Artifacts

Three files, two machines. Nothing else changed.

### cooper — `/home/gaetan/orca/workspaces/tars-delegated/`

Dedicated sandbox. Not a repo, not a checkout, no git remote.

**`delegate.sh`** — the only entrypoint Tars may call. Brief on stdin, answer
on stdout, last line `[delegate.sh exit=<code> log=<path>]`.

```bash
#!/bin/bash
set -euo pipefail
d=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$d/logs"
log="$d/logs/$(date -u +%Y%m%dT%H%M%SZ).log"
cd "$d"
ALLOW='Bash(python3:*),Bash(python:*),Bash(pytest:*),Bash(node:*),Bash(npm test:*),Bash(ls:*),Bash(cat:*),Bash(head:*),Bash(tail:*),Bash(wc:*),Bash(grep:*),Bash(rg:*),Bash(find:*),Bash(diff:*),Bash(git status:*),Bash(git diff:*),Bash(git log:*)'
timeout "${DELEGATE_TIMEOUT:-240}" "$HOME/.local/bin/claude" -p \
  --permission-mode acceptEdits --allowedTools "$ALLOW" \
  --output-format text 2>&1 | tee "$log"
rc=${PIPESTATUS[0]}
echo "[delegate.sh exit=$rc log=$log]"
exit "$rc"
```

The flags are fixed **inside the script on purpose**. Tars never types
`claude`, so Tars can never type `--dangerously-skip-permissions` and unwind
the permission layer. That is the difference between a guardrail and a
sentence in a markdown file asking nicely.

**`.claude/settings.json`** — deny-only permission rules (full list below).

**`logs/`** — one log per run, so a timed-out run still leaves its partial
output behind.

### VM — `~/.hermes/skills/delegate-to-cooper/SKILL.md`

The playbook, in the one place Tars actually reads. Top-level own category,
same shape as the existing custom skill `i-have-adhd`. Deliberately **no**
`disable-model-invocation`, so the model can reach for it unprompted.

Reaching model context needs no config edit and no restart: `prompt_builder.py`
walks `~/.hermes/skills/**/SKILL.md`, and `.skills_prompt_snapshot.json` is
invalidated by a path→(mtime,size) manifest, so a dropped file rebuilds the
skills index on the next turn. Confirmed: 72 → 73 skills, entry present.

`config.yaml` was not touched. The gateway was not restarted. SOUL.md was not
edited — the skill restates SOUL's rules rather than replacing them.

## Guardrails

Three layers. Only the first is prose.

### Layer 1 — what Tars is allowed to run on cooper (in the SKILL.md)

Exactly four command shapes, nothing else:

1. `ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF' … BRIEF`
2. `ssh cooper "ls -la ~/orca/workspaces/tars-delegated/"`
3. `ssh cooper "cat ~/orca/workspaces/tars-delegated/<file>"`
4. `ssh cooper "tail -n 200 <the log path delegate.sh printed>"`

Explicitly forbidden to Tars: invoking `claude`/`codex` or any agent CLI
directly or passing agent flags; `rm`/`rmdir`/`mv`/`chmod`/`sudo`/`systemctl`;
`git push`/`git reset`/`git clean` or any git write; opening, reviewing or
merging a PR; reading or passing on `~/.ssh/`, `~/.aws/`, `~/.claude.json`,
`~/.claude/.credentials.json`, `~/.config/sops/`, `~/.pgpass`, any `*.key`, any
`.env`; touching `192.168.0.3`/pve or the p-Hermes host; any path outside
`~/orca/workspaces/tars-delegated/`; and editing `delegate.sh` or
`.claude/settings.json` themselves.

Widening the blast radius to a real repo is a **decision handed back to
Gaetan**, not a path Tars widens on its own.

### Layer 2 — enforced deny rules (`.claude/settings.json`)

Bash: `rm`, `sudo`, `ssh`, `scp`, `rsync`, `curl`, `wget`, `nc`, `sops`, `op`,
`aws`, `git push`, `git reset`, `git clean`, `systemctl`.
Read: `~/.ssh/**`, `~/.aws/**`, `~/.awsvault/**`, `~/.claude/.credentials.json`,
`~/.claude/history.jsonl`, `~/.claude.json`, `~/.config/sops/**`, `~/.pgpass`,
`~/.boto`, `~/client.ovpn*`, `~/.terraform.d/**`, `~/**/*.key`,
`~/next-db-dump/**`, `~/annual-review-handoff/**` — i.e. the machine-map deny
list from `~/dev/gaetan-metarepo/machine/INDEX.md`.

`ssh` being denied is what keeps the delegated agent off pve and p-Hermes.
`curl`/`wget`/`nc` denied is what keeps a brief from turning into exfiltration.

**Load-bearing property, verified:** an *untrusted* workspace has its
`permissions.allow` entries **ignored** but its `permissions.deny` entries
**honoured**, and deny beats allow everywhere. So the deny list holds whether
or not anyone ever trusts this directory — which is why `.claude/settings.json`
is deny-only and the allowlist lives on the `delegate.sh` command line instead.
Do not "trust" this workspace in Claude Code; untrusted is the safer state.

### Layer 3 — scope

`delegate.sh` `cd`s into the sandbox, so the agent's writes are confined to it.
`additionalDirectories` is empty. The agent cannot reach a repo checkout.

### Reporting duty

Every Tars reply that used this skill states, in this order: the verdict; the
exact command it ran, verbatim, brief included; and the
`[delegate.sh exit=… log=…]` line. "I delegated it" is not an acceptable
summary of a command.

## Known ceilings

- **Synchronous, 240s.** hermes `code_execution.timeout` is 300. A longer job
  is killed at the cap with partial output in `logs/`; the skill tells Tars to
  report the timeout and offer a split brief. Add a background form only if
  this actually bites.
- **Single shared directory.** No per-task subdir, so two concurrent
  delegations would share a working tree. v1 is one Gaetan, one request.
- **Deny rules are pattern-based.** A determined agent could reach for an
  unlisted equivalent (a Python `os.remove`, say). The real boundary is the
  sandbox `cwd` plus the fact that Tars only ever fills in a prose brief. The
  deny list raises the floor; it is not a jail.
- **No repo access.** Deliberate. Widening is v2 and needs Gaetan's call.

## v2, if and when it is wanted

Repo work via `additionalDirectories` scoped to one checkout; a background form
with a poll command for jobs over the cap; and the `orca` orchestration layer
if supervised multi-agent runs are ever actually needed. None of it is needed
to close WF5.
---

# WF5 — Orca delegation v2 (authoritative)

**Status: authoritative design, 2026-08-07. Everything above this line is v1 and
is superseded** — kept for the record, not for execution. Decided by Gaetan in
session; recon evidence in `status/probes/wf5/orca-v2-capabilities.md`,
`orca-v2-layout.md`, `orca-v2-sequence-validation.md`, `orca-v2-current-skill.md`.

## What changed, and why

Gaetan's reframe, in his terms: **Tars is an orchestrator — "a secretary of
work", a mirror/replacement of Gaetan himself.** What Gaetan does all day is
prompt Claude Code; Tars should do that prompting for him, drive the Orca
sessions, and help schedule his work. Tars may double-check facts. Tars must not
do the implementation itself.

The load-bearing correction: **"Tars doesn't do the work" is a
separation-of-concerns guardrail, not a security guardrail.** It is fine for Tars
to have root on cooper, by design, because Gaetan has root on cooper. Everything
v1 built on the opposite reading is dropped:

- no sandbox directory,
- no fixed-flag wrapper script,
- no command allowlist,
- no deny-rule settings file,
- **no rule forbidding `sudo`**, and no enumerated forbidden paths or commands.

Stated positively, because the absence of a prohibition is not the same signal
to a model as a permission: **any path, any command, `sudo` included, by
design.** Gaetan has root on cooper; so does Tars. Tars does not ask permission
to run what the job needs, and does not stop to litigate whether an exploratory
`ls` or a file read outside the target repo is "really necessary". **Refusing to
act, or stalling to ask whether it is allowed, is a failure mode exactly as real
as doing the implementation itself** — both end with the work not happening. The
skill says this in those words.

What remains of the constraint is prose in the skill, correctly framed: Tars
prompts, tracks, verifies and reports; the spawned coding agent implements. SOUL
rules 1–3 (never write code, never open/review/merge a PR, never push) are
unchanged and are exactly that division of labour.

Three refinements to that prose, each closing a hole an adversarial read found:

- **The deliverable is named by ROLE, not by file type.** "Code, patch, script,
  config" is escapable — asked for documentation, Tars writes the markdown
  itself and believes it is inside the rules. The rule now reads: whatever
  Gaetan asked to exist is the delegated agent's to produce — code, patch,
  script, config, documentation, README, migration notes, any artifact. "It's
  only markdown" is the excuse the rule exists to refuse.
- **Two carve-outs, so the rule does not induce paralysis.** Writing the *brief*
  to disk is not doing the work (it is the instruction, not the deliverable),
  and quoting the agent's output, code and errors back to Gaetan is required
  evidence, not a violation.
- **One beat before the irreversible.** Gaetan dropped confinement, not
  prudence. For an action with no undo, Tars states what it is about to do and
  leaves the beat to be stopped — a statement, not a request for permission. It
  names no forbidden command, no path list, and no capability Tars lacks.

**Secret hygiene is retained and is not confinement.** The v1 skill named
concrete credential paths (`~/.ssh/`, `~/.aws/`, `~/.claude.json`,
`~/.claude/.credentials.json`, `~/.config/sops/`, `~/.pgpass`, `*.key`, `.env`);
both first-draft replacements dropped them, leaving the list nowhere. An
abstract "don't touch credentials" does not make a model classify
`~/.claude.json` as one. The list is restored in the skill, framed as *the
contents never leave the machine* — not into a brief, a Slack reply, argv, or an
evidence file — rather than as a place Tars may not go.

Target shape: **Tars spawns work on the running Orca app on cooper, in real
worktrees of real registered repos.** Default repo `mc-metarepo`
(`id:8099e312-3232-46f2-83a9-97aeaf5de5a2`). All registered repos are fair game;
Tars may register more (`orca repo add --path`) if the work genuinely needs one.

## Mechanism

```
Slack → Tars (VM 192.168.0.9) → ssh cooper → orca CLI → Orca app
        run-create → task-create → worker-start → check --wait → worker-read → worker-release
                                                        ↑            │
                                        check --ack <delivery> ──────┘  (every wait after the first)
```

Transport is **ssh + the `orca` CLI**. Corrected from the first draft, measured:
`orca` **is** on the non-interactive PATH from the VM, as `/usr/local/bin/orca`
(PATH = `/usr/local/sbin:/usr/local/bin:…`); `~/.local/bin/orca` is a wrapper on
the same binary and also works because the remote shell expands `~`. What is
absent from that PATH is `~/.local/bin`, not `orca` — the earlier claim had the
right remedy for the wrong reason. Also measured: `ssh cooper` resolves from the
VM to `192.168.0.4`, user `gaetan`, no prompt.

Orca's web/RPC server on `0.0.0.0:6768` is E2EE + device-token authenticated and
is deliberately **not** used — it would require pairing Tars as a device. The
layer above `worktree create`/`terminal create` is what makes this work: those
two attach no task and no dispatch, so they give no completion signal and no
durable id.

**`--ack` is mandatory, not hygiene.** `check --help`, verbatim: *"A bound Run
replays the same Delivery until `--ack`; process every message before
acknowledging."* An un-acked re-attach returns the batch already processed —
Tars reports the same `worker_done` twice and never sees the next message, a
loop that looks alive and is frozen. Only the first wait of a run omits `--ack`;
every wait after a delivery carries the id of the delivery it handled, and
`check --ack <id> --wait` does acknowledge-check-wait in one call. Read-only
looks use `--peek` (never marks read); bare `check` marks the oldest batch read,
which is how a batch gets consumed without being processed.

**Answering a worker `question` is `reply --id <msg_id> --body <text>`**, not
`send --to dispatch:<id>`. The worker is blocked inside `orchestration ask`; mail
sent to the dispatch is only delivered on the worker's *next* `check`, which a
blocked worker never issues — it errors on nothing and unblocks nothing. Flags
re-read from `--help` on cooper 2026-08-07: `check`, `reply`, `send`,
`worker-start`, `worker-show`.

## The three v1 objections, refuted

v1 rejected the Orca session layer on three grounds. All three were about the
**terminal** layer; v1 never inventoried `orchestration worker-*`, one layer up,
which is where the answers live.

| v1 objection (§"Mechanism chosen") | Refutation | Evidence |
|---|---|---|
| "handles go stale (`terminal_handle_stale`)" | True of `term_*` handles, and still true. But `worker-start` returns a **durable `dispatch_id`** — a SQLite row in `worker_dispatches` — and `worker-show`/`worker-read` keep working by that id after the terminal is gone (release archives the transcript first). Worktree selectors (`id:`/`name:`/`branch:`/`path:`) are equally durable. | capabilities §2, §4 |
| "no completion signal — Tars would hand-roll polling" | False at this layer. `worker_done` is a first-class message type with explicit `--outcome succeeded\|failed`, and `orchestration check --run <id> --wait --types worker_done,escalation,question --timeout-ms <n> --json` is a **single blocking call** that returns the instant it lands. Keepalive JSON on stderr every 15s distinguishes waiting from hanging. | capabilities §3; flags re-verified live in sequence-validation §1 item 4 |
| "durable coordinator state across ssh calls that share no session" | Confirmed present *and* confirmed **not** session-bound: `--run <id>` works standalone on `task-list`/`worker-list`/`gate-list`/`check` with **no prior `run-use` bind**. `run-use`/`run-current` exist only so a long-lived coordinator terminal can omit `--run`. | capabilities §1; **`--run` on `check` measured 2026-08-07, see below** |

### `--run <id>` standalone addressability — ~~CONFIRMED~~ **DISPROVEN by the live E2E, 2026-08-07**

> **CORRECTED POST-E2E.** Everything in this subsection below the next paragraph
> was wrong. The adversarial reviewer it dismisses was **right**. Evidence:
> `status/probes/wf5/apply-p3-orca-v2.md` §9.1.
>
> Reproduced independently from the VM over plain ssh, against this
> subsection's own probe run:
> `orca orchestration check --run run_24d000a3a6b6 --peek --json` →
> `ok:false, error.code: **no_active_terminal**`, while
> `task-list --run run_24d000a3a6b6 --json` → `ok:true`. The run is real and
> resolvable; `check` specifically cannot bind it without terminal context.
>
> A terminal-less caller must pass an explicit handle:
>
> | Verb | Needs |
> |---|---|
> | `run-create` | `--from <TERMINAL_HANDLE>` — else `no_active_sender_terminal` |
> | `task-create` | `--from <TERMINAL_HANDLE>` — the run id does not supply sender context |
> | `worker-start` | `--from <TERMINAL_HANDLE>` |
> | `check` (incl. `--wait` / `--ack`) | `--terminal <TERMINAL_HANDLE>`, **even when `--run` is given** |
> | `task-list --run`, `worker-list --run`, `worker-show --dispatch` | nothing extra — these genuinely are standalone |
>
> `run-create --help` documents it: `Usage: orca orchestration run-create
> --objective <text> [--from <handle>] [--json]`.
>
> The earlier measurement was almost certainly taken from a shell that already
> carried Orca terminal context — an agent running inside an Orca-managed
> terminal inherits it — not from "exactly Tars's stateless position".
>
> **The defensive fallback below could never fire:** it is keyed to
> `run_not_found`, and the real error code is `no_active_terminal`.
>
> The design survives, but only because Tars resolves a live terminal handle
> first. Open question it raises: Tars borrowed a handle belonging to the
> `Tars/orchestrator` worktree, which couples a delegated run to whatever
> terminal happens to be open. A durable, intentional sender terminal is unbuilt.

An adversarial review argued that `check` binds to the invoking coordinator
terminal, so a terminal-less ssh caller (exactly Tars's position) could never
resolve `--run` — which would have been fatal to the design. ~~Measured, and it
is wrong.~~ (It was not wrong. See the correction above.)

From a plain non-Orca shell with zero shared state: `orca orchestration
run-create --objective … --json` produced `run_24d000a3a6b6`; then from a
**fresh separate process**, `orca orchestration check --run run_24d000a3a6b6
--peek --json` returned `ok: true` with `runId` echoed back — not
`run_not_found`. `task-list --run` and `worker-list --run` on the same run
likewise returned `ok: true` (`legacyReadOnly: false`). `check --help`'s usage
line documents `[--run <run_id>]` explicitly; the "bound Run" prose describes
only the **default** when `--run` is omitted. The earlier `run_not_found` was on
`run_legacy_local`, a legacy-adopted run with `coordinator_handle: null` — not
representative.

**Design kept as-is.** One defensive line survives in the skill and nothing
more: if `run_not_found` ever appears on a run that `task-list --run` resolves,
fall back to polling `orca orchestration worker-show --dispatch <id> --json` and
read `result.state` (terminal: `succeeded|failed|stopped|abandoned`).

What v1 got right and v2 keeps: `orca` hard-depends on the desktop runtime being
up (`orca status --json` → `runtime.reachable` is step 0 of every sequence), and
the **reporting duty** — verdict, exact command verbatim, and the run/dispatch
ids in every reply. "I delegated it" is not a summary of a command.

## Superseded v1 artifacts

Everything below is dropped, not deprecated-with-fallback:

- `/home/gaetan/orca/workspaces/tars-delegated/delegate.sh` — the fixed-flag
  wrapper. Its whole premise (Tars must never type an agent flag) is void.
- `/home/gaetan/orca/workspaces/tars-delegated/.claude/settings.json` — the
  deny-rule layer, and the "Guardrails / Layer 1–2–3" section of v1 above.
- The `~/orca/workspaces/tars-delegated/` sandbox itself, including its
  no-git-remote and single-shared-directory properties, and the v1 rule "never a
  path outside" it.
- The v1 `~/.hermes/skills/delegate-to-cooper/SKILL.md`, replaced wholesale.
  13 of its 77 lines encoded the dropped mechanism (inventory with line numbers:
  `status/probes/wf5/orca-v2-current-skill.md` §5).

Skill install is a straight content replacement at
`~/.hermes/skills/delegate-to-cooper/SKILL.md` (renaming the directory is
optional and is the orchestrator's call). **No gateway restart is required** —
proven, not assumed: gateway PID 76255 ran continuously across an edit and
`.skills_prompt_snapshot.json` regenerated with the file's post-edit
`(mtime_ns, size)`. Verify pickup by comparing `os.stat(SKILL.md)` to the
snapshot's `manifest['<relpath>/SKILL.md']` entry — `MATCH` means picked up.
~~Skill count should stay 77 / 6-local (replacement, not an add).~~

**Corrected 2026-08-07 at apply time.** Live baseline is **80 enabled / 7 local /
66 builtin / 7 hub-installed**, and it did stay put across the replacement. But
neither the count nor the manifest proves pickup:

- The count cannot move on a replacement, so it proves nothing either way.
- `.skills_prompt_snapshot.json` still held the pre-apply `(mtime_ns, size)`
  immediately after the write — it only regenerates when the model next builds a
  prompt, so `MATCH` is not available at apply time.
- The `hermes skills list` **Category** column is derived from the on-disk path
  (`skills/<category>/<name>/SKILL.md`), **not** from frontmatter — skills sitting
  directly under `~/.hermes/skills/` render a blank category, so a frontmatter
  `category:` change is invisible there and is not a pickup signal.

**The only proof of pickup is behavioural**: run the E2E and confirm Tars drives
`orca orchestration …` over ssh rather than the v1 `delegate.sh` wrapper.

## Known ceilings

- **300s command cap.** Hermes `code_execution.timeout` is 300s, so the blocking
  wait is bounded at `--timeout-ms 240000`. Real coding work runs longer. This is
  survivable *only* because the run/dispatch/delivery ids are durable: a timeout
  is a **checkpoint**, Tars reports "still running, run `<id>`" and re-attaches
  on a later turn with `check --ack <delivery> --wait` or a cheap `worker-show`.
- **`worker-start` is a mutation with no guaranteed return.** It blocks until the
  worker is `ready` and can outrun the 300s cap *after* creating the dispatch,
  losing the id Tars needs. `--timeout-ms` exists on the verb (help-verified) and
  the skill passes `200000`, but the recovery rule is the real fix: after any
  apparent timeout, **do not assume nothing was created and do not re-run it** —
  `worker-list --run <RUN_ID> --json` first, adopt any dispatch found.
- **No outbound push, but Tars can wake itself.** Orca has no webhook, callback
  or completion-push that leaves cooper; its internal agent-status push is real
  HTTP but loopback-only (`127.0.0.1:${ORCA_AGENT_HOOK_PORT}`, per-pane token).
  Tars always has to ask. What closes the loop is on the *Hermes* side, not
  Orca's: nothing gives a Slack agent a later turn on its own, so the checkpoint
  path schedules one —
  `~/.local/bin/hermes cron create "5m" "<re-attach instruction>" --deliver slack --repeat 1`
  (`docs/facts.md` §Hermes: the bare platform name delivers to
  `SLACK_HOME_CHANNEL`; `/bg` always answers the invoking surface and can never
  reach home; the hermes CLI is at `~/.local/bin/hermes` and is **not** on PATH
  over non-interactive ssh). **When Gaetan asked to be told when it is done,
  scheduling the re-check is not optional** — "I will re-attach" without it is a
  broken promise. A caller-authored Claude Code `Stop` hook in the spawned
  worktree could notify Tars directly instead — untested, not built.
- **Orca must be running.** Hard dependency, no plain-shell fallback. If the
  desktop app is closed, every command fails; `orca status --json` first, always.
- ~~**Unverified response field names.**~~ **RESOLVED by the live E2E,
  2026-08-07** (`status/probes/wf5/apply-p3-orca-v2.md` §10):

  | Thing | Field path | Shape |
  |---|---|---|
  | Run | `result.run.id` | `run_<12 hex>` |
  | Task | `tasks[].id` | `task_<12 hex>` |
  | Dispatch | `result.dispatch.id` / `workers[].dispatchId` | **`ctx_<12 hex>`** — *not* `dispatch_…` |
  | Message / delivery | `messages[].id` | `msg_<12 hex>` |

  **There is no separate delivery-id field** — the id to carry and `--ack` is the
  message's own `id`. A non-empty batch message carries `id`, `run_id`,
  `delivery_contract`, `from_handle`, `to_handle`, `subject`, `body`, `type`,
  `priority`, `thread_id`, `payload`, `read`, `sequence`, `created_at`,
  `delivered_at`, `sender_pane_key`.

  Two traps found while measuring this: (1) a worker's reply is addressed
  `to_handle: "run:<runId>"` — the run's home inbox — so `check --terminal
  <some-other-handle> --run <id> --peek` returns `count: 0` even though the
  message exists; `orchestration inbox --json` is what shows messages across
  recipients. (2) The failure-mode table's `worker-show --dispatch <id>` →
  `result.state` is **wrong**; the real path is **`result.worker.state`**
  (`result.dispatch.status` is a separate field).

- **`worker-start` reaching `ready` does NOT mean the agent started.** Found live:
  Orca typed the brief into the Claude Code TUI and left it **unsubmitted**.
  `workerState: ready` and `input_accepted` were both true while the agent had
  not begun, so `check --wait` would have timed out forever on a job that never
  started — indistinguishable from "still running". Gaetan caught it by opening
  the worktree. The remedy, now in the live skill: after spawn, check the
  terminal preview and send
  `orca terminal send --terminal "<AGENT_TERMINAL_HANDLE>" --enter --json`, then
  confirm activity with `worker-show` / `worktree ps` before waiting. **This fix
  is unverified — no run has yet exercised it end to end.**

- **The live skill is not a stable artifact — Tars rewrites it.** During the E2E
  Tars called `skill_manage` and edited its own `SKILL.md` seven times (437 → 446
  → 452 lines; mode narrowed 0664 → 0600). The edits were correct and were
  sanctioned by Gaetan in-thread, but nothing in this design accounted for the
  model mutating its own governing skill. Both versions are archived:
  `artifacts/delegate-to-cooper-SKILL.md` (canonical v2, md5 `929b23de…`) and
  `artifacts/delegate-to-cooper-SKILL-after-tars-selfedit.md` (live, md5
  `60b9a244…`). Whether git and the VM get reconciled — and whether this
  capability should exist at all — is Gaetan's call.
- **The brief file path must be unique per run.** A fixed `/tmp/tars-brief.md` is
  silently clobbered by a concurrent delegation or by a retry, and `task-create
  --spec "$(cat …)"` then ships the wrong brief with no error. One `<slug>` per
  delegation names both the brief file and the worktree.
- **Worktrees accumulate.** Nothing self-cleans: an abandoned spawn leaves a
  checkout, a branch, rows in `orchestration.db`, and a terminal-history file.
  mc-metarepo already carries 9 worktrees, several stale. v2's default is to
  *keep* the worktree (it is the work product) and report its path — so the pile
  grows unless someone prunes. Prune on Gaetan's word, not on Tars' initiative.
- ~~**Only `claude` and `codex`** are verified-usable `--agent` values on cooper.~~
  **Corrected 2026-08-07: `claude` only.** `orca account list --json` shows
  `result.codex.accounts: []` and `activeAccountId: null` — codex has no account
  to run under on cooper. Active claude account: `009bfa67-…`.

- **The landing path is not the predicted pattern, and it contains a literal
  `null`.** The E2E worktree landed at
  `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/map-mc-metarepo-layout-readonly`,
  not `/home/gaetan/orca/workspaces/mc-metarepo/<slug>`. Orca reports that path
  itself in `effects[].id` and `worker.worktree_id`, so the `null` segment is
  Orca's own path construction. The skill's standing rule — read `path` from the
  JSON, never assume the pattern — is what kept the run working.

## The trust boundary moved — it did not disappear

Dropping the sandbox and the deny rules did not dissolve the trust boundary
around Tars; it **relocated it to Slack**. With v2, anything that can get a
message to Tars can transitively drive Orca on cooper, in real repos, as
`gaetan`. `SLACK_ALLOWED_USERS` plus the adapter's early-reject
(`adapter.py:5534-5550`, WARNING at `:5546`, 104 lines before the first
mention/thread branch) is now the **only** control holding that line, with
SOUL rule 4 as a model-level backstop.

This is a note about a control that must not lapse, not an argument against the
design. Concretely: `SLACK_ALLOWED_USERS` must stay populated and the early
reject must keep firing — the positive WARNING line is the thing to grep for, and
absence-of-reply is never evidence the gate held. Any future change that widens
who can reach Tars (new channel, relaxed `strict_mention`, an added platform, a
second adapter) is a change to *this* boundary and must be evaluated as one.

## Phase 2 — not built yet

Giving Tars **Gaetan's knowledge and preferences** — his metarepo, his coding
standards, his project context, the shape of a brief he would have written
himself — is a deliberately separate later phase. v2 makes Tars able to *drive*
Orca correctly; it does not make Tars' briefs sound like Gaetan's. Scheduling
work (what to run when, what is worth spawning at all) belongs to that phase too.
Nothing in this section depends on it, and nothing in it should be smuggled in
early.

## PARKED, awaiting Gaetan

Not ruled on. Nothing below blocks installing v2 as written; each is a knob that
would change behaviour if turned, so none of them gets turned on an agent's
initiative.

- **`approvals.mode`** — `manual` | `smart` | `off`. The v2 skill does not set it
  and the spawned agent runs on whatever the repo/worktree already implies. Under
  the separation-of-concerns reading `off` is arguable (the agent is doing the
  work Gaetan asked for, on Gaetan's machine); `smart` is the conservative pick;
  `manual` strands a worker Tars cannot un-block except through `reply`. Needs
  his call before any spawn relies on one.
- **Renaming the skill to `drive-orca-on-cooper`.** `delegate-to-cooper` is the
  v1 name and describes the dropped sandbox mechanism, not what v2 does. Install
  is a content replacement at the existing path; renaming the directory is a
  separate, optional move and would change the name Tars sees in its skill list.
  Skill count should stay 77 / 6-local either way.
- **Phase 2's landing surface for Gaetan's knowledge and preferences** — metarepo
  content, coding standards, project context, the shape of a brief he would have
  written himself. Whether that arrives as skills, as SOUL/profile text, as an
  MCP-reachable store, or as repo files the spawned agent reads is undecided, and
  it determines how much of it Tars carries in-context versus looks up.
