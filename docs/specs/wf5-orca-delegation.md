# WF5 — Orca delegation playbook v1

Authoritative spec. Status: **implemented and verified 2026-08-07**.
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
