# WF5 — current `delegate-to-cooper` skill, verbatim, plus loading mechanics

Captured read-only over ssh from the Tars VM (192.168.0.9), 2026-08-07. No
edits, no `hermes config set`, no restart performed. This is the file the
Orca-integration redesign will REPLACE.

## 1. Verbatim current SKILL.md

Path: `~/.hermes/skills/delegate-to-cooper/SKILL.md` on the Tars VM.
77 lines, 3304 bytes. `stat` on the VM: `Modify: 2026-08-07 20:21:31.919000904 +0000`.

```markdown
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
```

## 2. House-style frontmatter conventions (from other locally-authored skills)

`hermes skills list --source local` (6 local skills on this VM):

```
                             Installed Skills
┏━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━┳━━━━━━━━┳━━━━━━━┳━━━━━━━━━┓
┃ Name                     ┃ Category          ┃ Source ┃ Trust ┃ Status  ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━╇━━━━━━━━╇━━━━━━━╇━━━━━━━━━┩
│ delegate-to-cooper       │                   │ local  │ local │ enabled │
│ hermes-lcm               │                   │ local  │ local │ enabled │
│ meme-generator           │ creative          │ local  │ local │ enabled │
│ hermes-skill-portability │ hermes-operations │ local  │ local │ enabled │
│ skill-library-auditing   │ hermes-operations │ local  │ local │ enabled │
│ last30days               │ research          │ local  │ local │ enabled │
└──────────────────────────┴───────────────────┴────────┴───────┴─────────┘
0 hub-installed, 0 builtin, 6 local — 6 enabled, 0 disabled
```

Note: the "Category" column here is derived from the **directory** the skill
lives in under `~/.hermes/skills/` (e.g. `hermes-skill-portability` is nested
under `hermes-operations/`), not from the `metadata.hermes.category` YAML key.
`delegate-to-cooper` sits directly under `~/.hermes/skills/` so its Category
column is blank even though its frontmatter sets `category:
delegate-to-cooper` — that frontmatter field is cosmetic/unused by this view.

Two sibling skills under `hermes-operations/` (locally authored, same author
style as `delegate-to-cooper`):

```yaml
# ~/.hermes/skills/hermes-operations/hermes-skill-portability/SKILL.md
---
name: hermes-skill-portability
description: "Use when migrating Hermes skills between hosts/profiles."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [hermes, skills, migration, portability, verification]
---
```

```yaml
# ~/.hermes/skills/hermes-operations/skill-library-auditing/SKILL.md
---
name: skill-library-auditing
description: "Use when auditing or comparing active skill libraries."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [skills, audit, inventory, checksums, provenance, fleet]
---
```

```yaml
# ~/.hermes/skills/i-have-adhd/SKILL.md  (top-level local skill, disables auto-invocation)
---
name: i-have-adhd
description: 'Shape output for a reader with ADHD: lead with the next action,
  number multi-step work, restate state after every tool call...'
disable-model-invocation: true
license: MIT
metadata:
  hermes:
    tags: [ADHD, Output Style, Formatting, Productivity]
    category: productivity
    related_skills: []
---
```

**Two distinct frontmatter dialects observed among local skills:**
- `delegate-to-cooper` / `i-have-adhd` style: `license: MIT`, `metadata.hermes.{tags, category, related_skills}` (the `related_skills` array is present but empty in both).
- `hermes-skill-portability` / `skill-library-auditing` style: `version: 1.0.0`, `platforms: [...]`, `metadata.hermes.tags` only — no `license`, `category`, or `related_skills`.

`disable-model-invocation: true` appears only on `i-have-adhd` (an output-style
skill meant to be always-on / manually toggled, not autonomously invoked) —
`delegate-to-cooper` does NOT set it, so it is model-invocable by
description-matching like any other skill.

## 3. Loading mechanics — with evidence

**How a SKILL.md reaches the model's context:** Hermes maintains
`~/.hermes/.skills_prompt_snapshot.json` (mode `0600`, JSON, `"version": 2`).
Top-level keys: `manifest` (dict, 93 entries — one per `SKILL.md` / `DESCRIPTION.md` file, each value `[mtime_ns, size_bytes]`), `skills` (list, 77 entries — the compiled skill-description text injected into the system prompt), `category_descriptions` (dict, 15 entries). This is a cache: the `skills` list is what actually gets folded into the prompt; `manifest` is the invalidation key used to decide whether that cache is still fresh against the files on disk.

**Invalidation confirmed directly** — compared the live file to its manifest entry after editing `delegate-to-cooper/SKILL.md` (edited 20:21:31 UTC):

```
snapshot manifest entry (mtime_ns, size): [1786134091919000904, 3304]
live file (mtime_ns, size):               (1786134091919000904, 3304)
MATCH
```

i.e. `os.stat(SKILL.md).st_mtime_ns, .st_size` for the file on disk equals the `["<relpath>/SKILL.md"]` entry in the snapshot's `manifest` dict. When they diverge, the snapshot is stale and gets regenerated on next use; when they match (as here), the running gateway's in-context skill text is confirmed current for that file.

**No restart required — confirmed, not just repeated from prior evidence:**
- `hermes-gateway.service` (`--user` unit): `ActiveEnterTimestamp` / `ExecMainStartTimestamp` = **Fri 2026-08-07 19:40:45 UTC**, same PID (76255) throughout, still running 1h45m later at time of capture. One continuous process, no restart.
- `delegate-to-cooper/SKILL.md` was last modified at **20:21:31 UTC** — 41 minutes AFTER the gateway started.
- `.skills_prompt_snapshot.json` itself was last written at **21:24** — after the edit — and, per the MATCH above, its manifest entry for that file reflects the POST-edit mtime/size.
- Conclusion: the same gateway process picked up the edited file into its snapshot without any restart or config edit in between.

**Exact command that proves a changed skill was picked up** (run on the VM):

```bash
python3 -c "
import json, os
d = json.load(open(os.path.expanduser('~/.hermes/.skills_prompt_snapshot.json')))
entry = d['manifest']['delegate-to-cooper/SKILL.md']
st = os.stat(os.path.expanduser('~/.hermes/skills/delegate-to-cooper/SKILL.md'))
print('snapshot:', entry, ' live:', (st.st_mtime_ns, st.st_size))
print('MATCH' if tuple(entry) == (st.st_mtime_ns, st.st_size) else 'STALE')
"
```
`MATCH` = the running snapshot already reflects the file on disk (no restart needed). `STALE` = snapshot hasn't regenerated yet for that file (grep `skills` list contents for the new description text as a second check, or trigger any skill-related hermes CLI call, which regenerates the snapshot).

No gateway-log evidence of an explicit "skill reload" line was found (`grep -i -E 'reload|manifest|snapshot|invalidat' ~/.hermes/logs/{gateway,agent}.log` → the only "snapshot" hits were unrelated "Session snapshot created" tool-environment lines) — the mechanism is a silent stat-based cache check, not a logged event. The manifest-comparison command above is the evidence-based way to confirm pickup, not a log grep.

## 4. Current total skill count

**77 skills** total (`d['skills']` length in the snapshot; matches `find -L ~/.hermes/skills -name SKILL.md | wc -l` = 77, which is `find` without `-L` = 76 direct entries + 1 (`hermes-lcm`, a symlink to `~/.hermes/plugins/hermes-lcm/skills/hermes-lcm/`) reached only by following the symlink).

Of these, only **6 are `local`** (author-owned, not hub/builtin): `delegate-to-cooper`, `hermes-lcm`, `meme-generator`, `hermes-skill-portability`, `skill-library-auditing`, `last30days`. The other 71 are bundled/hub skills (`~/.hermes/skills/.bundled_manifest` lists them, e.g. `airtable`, `github-*`, `claude-code`, `codex`, etc.).

This is the before-count for a before/after comparison once `delegate-to-cooper` is replaced (skill count should stay 77 / 6-local if it's a straight content replacement, not an add).

## 5. Lines that mention sudo / deny / forbidden / the tars-delegated sandbox / delegate.sh

Line-numbered, from `grep -n -i -E 'sudo|deny|denied|forbidden|tars-delegated|delegate\.sh|guardrail|dangerously-skip' SKILL.md` on the VM. These are exactly what the Orca-integration redesign removes — inventoried only, not edited:

- **L20**: `ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF'` — the whole delegation mechanism is this one fixed wrapper script in the dropped sandbox path.
- **L26**: `line is \`[delegate.sh exit=<code> log=<path>]\`. Synchronous: it returns within` — the reporting contract is tied to `delegate.sh`'s own exit-code convention.
- **L35**: `1. the \`delegate.sh\` heredoc above` — enumerated as the only sanctioned command.
- **L36**: `2. \`ssh cooper "ls -la ~/orca/workspaces/tars-delegated/"\`` — sandbox path hardcoded.
- **L37**: `3. \`ssh cooper "cat ~/orca/workspaces/tars-delegated/<file>"\`` — sandbox path hardcoded.
- **L38**: `4. \`ssh cooper "tail -n 200 <the log path delegate.sh printed>"\`` — references delegate.sh's log output.
- **L45**: `  agent flags. Only \`delegate.sh\`, which fixes them. \`--dangerously-skip-` — forbids invoking `claude`/`codex` directly; only the wrapper is allowed (paired with L44's "Never invoke `claude`, `codex` or any agent CLI directly").
- **L47**: `- Never \`rm\`, \`rmdir\`, \`mv\`, \`chmod\`, \`sudo\`, \`systemctl\` on cooper.` — the explicit sudo/command-forbidding line the redesign drops per Gaetan's directive.
- **L54**: `- Never a path outside \`~/orca/workspaces/tars-delegated/\`. If Gaetan wants` — the sandbox-confinement rule, dropped entirely under the new design (real worktrees of real registered repos are now in scope).
- **L57**: `- Never edit \`delegate.sh\` or \`.claude/settings.json\` — those are the` — deny-rule/wrapper self-protection clause.
- **L58**: `  guardrails, and asking me to change them is a request I refuse in one line.` — names the wrapper+settings.json pair as "the guardrails."
- **L67**: `3. the \`[delegate.sh exit=… log=…]\` line.` — reporting format again tied to delegate.sh.
- **L75**: `- The agent answers \`DENIED\` — a guardrail fired. I report that plainly and do` — references the wrapper's deny-rule enforcement behaviour (implies `delegate.sh`/settings.json can return a `DENIED` sentinel).

13 lines total reference the dropped v1 mechanism (sandbox path, `delegate.sh`, sudo/command denials, or `DENIED` guardrail behaviour).
