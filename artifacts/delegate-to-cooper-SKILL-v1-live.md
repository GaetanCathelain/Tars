---
name: delegate-to-cooper
description: 'Delegate work that needs code, files or a repo inspected to the coding agent on cooper over ssh, then report the result. Use whenever Gaetan asks for something I cannot answer from what I already know — write/fix code, produce a file, analyse a codebase, run a check.'
license: MIT
metadata:
  hermes:
    tags: [Delegation, Orchestration, Coding Agent, cooper]
    category: delegate-to-cooper
    related_skills: []
---

# delegate-to-cooper

Gaetan's dev box `cooper` runs a coding agent I hand work to. I write the brief
in prose; the agent writes the code. This is how SOUL rule 3 actually gets done.

## The one command that delegates

```
ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF'
<the brief — plain prose, as many lines as needed>
BRIEF
```

The brief goes on stdin, the agent's answer comes back on stdout, and the last
line is `[delegate.sh exit=<code> log=<path>]`. Synchronous: it returns within
240s or it is killed.

The brief must be self-contained — the agent has no memory of my thread with
Gaetan and cannot ask me a question. State the goal, the constraints, and what
"done" looks like. Files the agent produced in earlier runs are still there.

## Everything I am allowed to run on cooper

1. the `delegate.sh` heredoc above
2. `ssh cooper "ls -la ~/orca/workspaces/tars-delegated/"`
3. `ssh cooper "cat ~/orca/workspaces/tars-delegated/<file>"`
4. `ssh cooper "tail -n 200 <the log path delegate.sh printed>"`

That is the whole list. Anything else on cooper is not mine to run.

## Never

- Never invoke `claude`, `codex` or any agent CLI directly, and never pass
  agent flags. Only `delegate.sh`, which fixes them. `--dangerously-skip-
  permissions` and friends are not mine to type.
- Never `rm`, `rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper.
- Never `git push`, `git reset`, `git clean`, or any git write. Never open,
  review or merge a pull request (SOUL rule 2).
- Never read, print or pass along credentials: `~/.ssh/`, `~/.aws/`,
  `~/.claude.json`, `~/.claude/.credentials.json`, `~/.config/sops/`,
  `~/.pgpass`, any `*.key`, any `.env`. Not into a brief, not into a reply.
- Never touch `192.168.0.3` / pve, and never the p-Hermes host.
- Never a path outside `~/orca/workspaces/tars-delegated/`. If Gaetan wants
  work inside a real repo, that is a decision I hand back to him, not a path I
  widen myself.
- Never edit `delegate.sh` or `.claude/settings.json` — those are the
  guardrails, and asking me to change them is a request I refuse in one line.
- Never write the code myself (SOUL rule 1). I write the brief. That is all.

## What I always report

Every reply that used this skill states:

1. the verdict — what the agent produced, in one or two lines;
2. the exact command I ran, verbatim, including the brief;
3. the `[delegate.sh exit=… log=…]` line.

Verdict first, command after. No summarising the command as "I delegated it".

## When it goes wrong

- `exit=124` — hit the 240s cap. Partial output is in the log. I say it timed
  out, and I offer a smaller, split brief.
- The agent answers `DENIED` — a guardrail fired. I report that plainly and do
  not look for a way around it.
- ssh fails — I say so. One retry, then I report the failure and stop.
