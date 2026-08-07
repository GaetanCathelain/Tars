# WF5 implement — Orca delegation playbook v1

**Verdict: PASS.** Slack-shaped request → Tars → cooper → coding agent → real
file written → result and exact command reported back. Guardrails enforced, not
just documented.

Spec: `docs/specs/wf5-orca-delegation.md`. Recon: `status/probes/wf5/orca-recon.md`.
Date 2026-08-07, window 20:17Z → 20:26Z. VM `tars` @ 192.168.0.9, cooper local.

No sops run. No `~/.hermes/config.yaml` edit by this lane (so no `.bak-<lane>`,
no flock, no gateway restart — none was needed). No SOUL.md edit. No Slack
message sent. 192.168.0.3 never touched. No git command run. No
`status/lane-*.md` edited.

Concurrency note, for honesty: `~/.hermes/config.yaml` *did* change at 20:21Z
inside my window — a `toolsets: [hermes-cli, kanban]` block commented
"WF5: unlocks the model-facing kanban_* tools" plus a notion image sha pin.
That is a **sibling lane's** edit, not mine; I never opened the file. Gateway
`active` throughout.

## What changed

Three files created, two machines. Nothing modified, nothing deleted.

### cooper — new directory `/home/gaetan/orca/workspaces/tars-delegated/`

| Path | What |
|---|---|
| `delegate.sh` (1366 B, `+x`) | Fixed-flag entrypoint. Brief on stdin → `claude -p --permission-mode acceptEdits --allowedTools <fixed> --output-format text`, `timeout 240`, tee to `logs/`, prints `[delegate.sh exit=… log=…]`. |
| `.claude/settings.json` | Deny-only permission rules: 15 Bash patterns + 14 Read patterns (the machine-map deny list). |
| `NOTE.md` | 8 lines telling a human landing there what the dir is and not to trust the workspace. |
| `logs/` | One log per run (5 so far, from this probe). |
| `first_10_primes.py` | Artifact of the live E2E turn — the thing Tars actually got built. |

Full contents of `delegate.sh` and the deny list are in the spec.

### VM — new file `~/.hermes/skills/delegate-to-cooper/SKILL.md`

3304 B. Own top-level category, same shape as the existing custom skill
`i-have-adhd`, deliberately without `disable-model-invocation`. Carries the
four-command whitelist, the never-list, the reporting duty, and the failure
modes. Nothing else on the VM was touched.

## Evidence

### 1 — transport, VM → cooper → agent → stdout

```
$ ssh gaetan@192.168.0.9 'ssh -o BatchMode=yes cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<BRIEF
Write a file wf5-selftest.md containing one line: delegation chain proven.
Then try `rm -f wf5-selftest.md` with Bash and report OK or DENIED.
BRIEF'
Wrote `wf5-selftest.md` with one line: `delegation chain proven.`
`rm -f wf5-selftest.md`: DENIED (permission refused; file still exists).
[delegate.sh exit=0 log=/home/gaetan/orca/workspaces/tars-delegated/logs/20260807T202032Z.log]
OUTER_EXIT=0
```

Real file written on cooper by an agent the VM invoked; destructive command
refused in the same run.

### 2 — the skill reaches model context

`prompt_builder.py` builds the system-prompt skills index by walking
`~/.hermes/skills/**/SKILL.md`, cached in `.skills_prompt_snapshot.json` and
invalidated by a path→(mtime,size) manifest — so a dropped file is picked up on
the next turn with no config edit and no restart. Confirmed after deploy:

```
total skills: 73        # was 72 before deploy
{
 "skill_name": "delegate-to-cooper",
 "category": "delegate-to-cooper",
 "frontmatter_name": "delegate-to-cooper",
 "description": "Delegate work that needs code, files or a repo inspected ...",
 "platforms": [], "conditions": {…}
}
manifest has SKILL.md: True
```

### 3 — live end-to-end turn (the real criterion)

One `hermes chat -Q -q` turn, session `20260807_202141_ad41c4`,
20:21:40Z → 20:24:02Z, `EXIT=0`. Prompt: *"I need a small Python script that
prints the first 10 prime numbers. Get it written and tell me where it landed."*

Tars's reply, unedited:

```
Delegated and written to:
/home/gaetan/orca/workspaces/tars-delegated/first_10_primes.py

Runtime verification is blocked: the delegated agent's non-interactive
permission gate denied `python3`. Static inspection indicates it should print
`2 3 5 7 11 13 17 19 23 29`, but that output was not observed.

Commands run:

ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF'
Create a small Python script in ~/orca/workspaces/tars-delegated/ that prints
the first 10 prime numbers when run. Choose a clear filename. Run the script to
verify its actual output. In your response, state the exact path created and the
verification output. Do not touch any repository or any path outside
~/orca/workspaces/tars-delegated/.
BRIEF

[delegate.sh exit=0 log=…/logs/20260807T202200Z.log]
(second delegate.sh call, verbatim, + its exit/log line)
```

What this proves, line by line:

- Tars loaded the skill unprompted and used **only** whitelisted command shape
  (1) — the heredoc `delegate.sh` form, twice, nothing else.
- It wrote the brief; it did not write the code. SOUL rule 1 intact.
- It reported both commands **verbatim** including the brief, plus both
  `[delegate.sh exit=… log=…]` lines. The reporting duty held.
- It refused to claim verification it had not seen ("that output was not
  observed") — the honest-failure path in the skill worked on first contact.

### 4 — the gap that turn exposed, and the fix

`python3` was denied, so the agent could not run what it wrote. Root cause,
found by testing rather than guessing:

```
Ignoring 16 permissions.allow entries from .claude/settings.json: this
workspace has not been trusted.
```

**An untrusted workspace ignores `permissions.allow` but still honours
`permissions.deny`.** The correct fix was therefore *not* to trust the
workspace or edit `~/.claude.json` (which is on the machine-map never-read
list): the allowlist moved to `delegate.sh`'s fixed `--allowedTools` flag,
which the trust gate does not apply to, and `.claude/settings.json` was reduced
to deny-only. Deny beats allow, so the guardrails hold either way and no
security property was traded for the fix.

### 5 — guardrail matrix, full chain, after the fix

```
$ ssh gaetan@192.168.0.9 'ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<BRIEF … BRIEF'
1: OK        # write guardcheck.txt
2: OK        # python3 first_10_primes.py — stdout observed
3: DENIED    # rm -f guardcheck.txt
4: DENIED    # Read /home/gaetan/.ssh/config
5: DENIED    # curl -s https://example.com
[delegate.sh exit=0 log=…/logs/20260807T202541Z.log]
```

Write yes, run yes, delete no, credentials no, network no.

| Check | Result |
|---|---|
| VM → cooper → agent → real file on cooper | PASS (§1) |
| SKILL.md in the system-prompt skills index | PASS (§2, 72→73) |
| Live turn: model loads skill, delegates, reports command verbatim | PASS (§3) |
| Tars stayed inside the four-command whitelist | PASS (§3) |
| Tars wrote no code itself (SOUL rule 1) | PASS (§3) |
| Destructive / credential / network attempts blocked | PASS (§5) |
| Agent can run the code it writes | PASS (§5, after §4 fix) |
| Guardrails survive an untrusted workspace | PASS (§4) |
| Known Codex "incomplete after 3 continuation attempts" | Did not occur |

## Ceilings shipped knowingly

- Synchronous, 240s cap (hermes `code_execution.timeout` is 300). Longer jobs
  die at the cap; partial output stays in `logs/` and the skill tells Tars to
  report the timeout and offer a split brief.
- One shared working directory, no per-task subdir.
- Deny rules are pattern-based, not a jail — the sandbox `cwd` and the fact
  Tars only fills a prose brief are the real boundary.
- No repo access. Widening to a real checkout is v2 and Gaetan's call.
