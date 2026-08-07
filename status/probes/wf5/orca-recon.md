# WF5 recon — Tars delegating to Orca sessions on cooper

Read-only recon. Nothing changed anywhere. Builds on probe 15 (plain ssh
command execution from Tars/VM to cooper works:
`status/probes/wf4/p15-delegation-cooper.md`).

## What's on cooper

- `orca` CLI at `/home/gaetan/.local/bin/orca` (also `/usr/local/bin/orca`,
  `/opt/Orca/orca-ide`). 223 commands (`orca agent-context --json` for full
  schema). Talks to the locally-running Orca runtime app.
- `orc-tab` at `/home/gaetan/.local/bin/orc-tab` — thin launcher wrapper only
  (`usage: launch.sh <launcher> <title> <brief-path>`), not a general
  scripting surface. Not useful for delegation directly.
- `/home/gaetan/orca/workspaces/` currently has two repos checked out:
  `Tars/` and `mc-metarepo/` (this orchestrator's own tree, and the
  metarepo). `orca worktree list --json` shows the full set of Orca-managed
  worktrees across repos (mc-metarepo main + several feature worktrees) with
  live terminals attached (`orca terminal list --json`).
- Bundled skill guides ship with the CLI (`orca skills list` /
  `orca skills get <name>`) and are the closest thing to "docs for
  programmatic use": `orca-cli` (terminal/worktree spawn+send, the simple
  path) and `orchestration` (structured multi-agent: mailboxes, ask/reply,
  task DAGs, worker_done). No standalone README/docs tree under
  `~/orca`, `/opt/Orca`, or `/usr/share/doc/orca-ide` beyond these skill
  guides.

## How a running Orca session can be messaged

Two built-in mechanisms, both driven entirely through the `orca` CLI (no
sockets, no ad hoc files):

1. **`orca terminal send`** — write text (optionally `--enter`) into a live
   PTY handle. Pair with `orca terminal create` (spawn a new terminal running
   an agent in a worktree) or target an existing handle from
   `orca terminal list --json`. Read results with `orca terminal read` /
   `orca terminal wait --for tui-idle|exit`.
2. **`orca orchestration send/ask/task-create`** — structured inter-agent
   mailbox: threaded messages, blocking `ask` with timeout, task DAGs,
   `worker_done` completion signals, `orchestration check`/`inbox` to poll.
   Needs a bound Run (`orchestration run-create`/`run-use`) first.

## Candidate delegation mechanisms, ranked by simplicity

### 1. ssh + `orca terminal create` / `orca terminal send` (simplest)

Tars, from the VM, ssh's into cooper and drives an existing or fresh
Orca-managed terminal directly — the same shape as probe 15's plain command
execution, just aimed at the `orca` CLI instead of a bare shell.

```bash
# spawn an agent in a worktree and hand it a prompt
ssh gaetan@192.168.0.9-equivalent-cooper-user \
  "orca worktree create --name <task> --no-parent --agent codex --prompt '<brief>' --json"
# or target an existing terminal
ssh <cooper> "orca terminal send --terminal <handle> --text '<prompt>' --enter --json"
# check back later
ssh <cooper> "orca terminal read --terminal <handle> --json"
```

- **Failure modes**: terminal handle can go stale (`terminal_handle_stale` —
  must re-`terminal list` and never dual-send to old+new handle); Orca app
  must actually be running on cooper (`orca status`) or `terminal create`
  fails; no built-in completion signal — polling `terminal read`/
  `terminal wait --for tui-idle` is the only way to know it's done; ssh
  session itself can drop mid-command with no retry built in.
- **Guardrails needed**: Tars must never be allowed to choose the launched
  agent's *command* freely — only fill a prompt/brief string. It must never
  create/attach terminals outside `~/orca/workspaces/` scoped worktrees, must
  never target a worktree it doesn't own the naming convention for, and
  must never pass `--command` (arbitrary shell) to `terminal create`
  itself — only `--agent <id> --prompt <text>`. No `orca worktree rm`,
  `repo add`, or `account` subcommands should ever reach Tars's playbook.

### 2. ssh + `orca orchestration` (more structure, more moving parts)

Same ssh transport, but uses the Run/mailbox layer: `run-create`, then
`orchestration send --to run:<id>` / `orchestration ask` (blocking,
timeout-bounded) / `task-create` for DAGs, and `orchestration check`/`inbox`
to read back. Gives Tars a native "wait for result" and "escalation" path
instead of hand-rolled polling.

- **Failure modes**: more state to track (Run id, task id, dispatch id)
  across ssh calls that don't share a session; `ask` blocks the *calling*
  process for up to `--timeout-ms`, which is awkward from a one-shot ssh
  command — needs a background/async invocation pattern; task-create records
  durable coordinator state that outlives the ssh call, so a crashed/lost
  Tars turn can leave orphaned tasks.
- **Guardrails needed**: same command-allowlist problem as (1) — Tars gets
  `orchestration send/ask/task-create/check` only, never `task-create`
  chained into `dispatch --inject` with arbitrary payload, and never a
  `--payload` JSON blob that could smuggle a `--command`-equivalent. The
  skill guide itself flags `orchestration` vs `orca-cli` as mutually
  exclusive framings (handoff-and-stop vs supervised); mixing them in one
  playbook risks Tars both dispatching *and* holding tracking state it
  can't clean up.

### 3. hermes skill carrying the playbook (how Tars *learns* the above, not a transport)

Not a delegation transport by itself — this is how Tars would carry
"delegate to Orca on cooper" as reusable instructions, parallel to existing
custom skills. Confirmed mechanism, read-only inspected:

- `~/.hermes/skills/<category>/<name>/SKILL.md` — same format as the
  existing custom skill `~/.hermes/skills/i-have-adhd/SKILL.md` (YAML
  frontmatter: `name`, `description`, optional `disable-model-invocation`,
  `metadata.hermes.tags/category`). `orca skills list`-equivalent on the
  hermes side is `hermes` skill listing (shown via `hermes skills list`,
  not run here to stay read-only beyond `--help`); population happens via
  `hermes sync` (Skill Sync) or a manually authored SKILL.md dropped in that
  tree.
- No sibling instruction files exist next to `~/.hermes/SOUL.md` — SOUL.md
  is the only profile-level file at that layer (checked `ls ~/.hermes/*.md`).
  SOUL.md's own hard rules are directly relevant: rule 1 forbids Tars from
  writing code/config/scripts "not even as an example" — a delegation
  playbook must stay pure orchestration (ssh + `orca` CLI calls with a
  prompt string), never construct or edit a file of code on Tars's own
  behalf. Rule 3 already frames this exact use case: "Work that needs code
  written is delegated to a coding agent."
- **Failure modes**: a SKILL.md is inert documentation — it doesn't grant
  Tars the ssh key, the orca CLI, or any capability; it only shapes what
  commands the LLM chooses. `disable-model-invocation` would have to be
  *unset* for Tars to actually invoke it in response to a request. Skill
  Sync pulls from wherever the team's skill source is — same
  supply-chain-trust question as any other community/synced skill.
- **Guardrails needed**: the SKILL.md itself is the guardrail surface — it
  should enumerate the exact allowed `orca` subcommands (from mechanism 1 or
  2) and explicitly restate SOUL.md's "never write code" / "never push" /
  "Gaetan-only" rules so they survive being read out of a skill file rather
  than the system prompt.

## What Tars must never be able to run on cooper (applies to all three)

- Any `orca` subcommand that mutates repo/account/environment state:
  `account add`, `repo add`/`repo set-base-ref`, `environment add/rm`,
  `worktree rm`, `automations create/remove/run`.
- `terminal create --command <arbitrary>` — only the `--agent <id> --prompt`
  form; a free-form `--command` is equivalent to arbitrary shell exec on
  cooper via ssh, twice removed.
- Anything outside `~/orca/workspaces/` — no path arguments pointing at
  `~/dev/`, `~/.ssh/`, or any path on the machine-map deny list
  (`~/dev/gaetan-metarepo/machine/INDEX.md`).
- Direct `orca orchestration dispatch --inject` with raw `--payload` JSON
  (bypasses the safer `--task-id`/`--dispatch-id` flags the CLI itself
  recommends to avoid quoting/injection issues).

## Sources (all read-only)

- `orca --help`, `orca agent-context`, `orca orchestration send/ask/task-create --help`,
  `orca terminal create/send --help`, `orca worktree list --json`,
  `orca terminal list --json`, `orca skills list`, `orca skills get orca-cli`
  — run on cooper.
- `hermes --help`, `hermes skills list`, `ls ~/.hermes/skills/`,
  `ls ~/.hermes/*.md`, `cat ~/.hermes/SOUL.md`,
  `head ~/.hermes/skills/i-have-adhd/SKILL.md` — run over ssh on the VM.
- `status/probes/wf4/p15-delegation-cooper.md` — prior probe establishing
  plain ssh command execution VM→cooper works.
