# WF6 rollback — daily-work-brief SKILL.md reverted to pre-WF6 on the live VM

Date: 2026-08-10, ~20:14Z. Host: `gaetan@192.168.0.9` (Tars VM, clock UTC).
Scope: exactly ONE file. No config, no cron, no state file, no MCP server touched.

## Why

`~/.hermes/skills/orchestration/daily-work-brief/SKILL.md` was deployed today
(v1.3.0) with a Linear "Board" section. A live verification run proved the
board's CONTENT is fabricated: 7 of 12 issue titles invented or paraphrased,
terminal issues rendered as live rows, and a "+N more" count wrong by ~44.
The 08:30 Europe/Paris cron (06:30Z) fires tomorrow morning and would deliver
that plausible-but-wrong board to Gaetan, who plans his day from it. The fixed
version is not ready, so the last-known-good copy goes back and a corrected
version is redeployed later. Failing safe beats shipping it.

## Backup confirmed before acting

`SKILL.md.bak-wf6-20260810T193012Z` — present, 9362 bytes, mtime
2026-08-10 11:00:49 UTC.

## sha256 (full)

| file | sha256 |
|---|---|
| CURRENT (bad, v1.3.0, 18039 B) before rollback | `e8cc68e5d7199a29d21d437a23d1f3459593c92ed303cf02728fd7b658d6ad03` |
| BACKUP `SKILL.md.bak-wf6-20260810T193012Z` (9362 B) | `69a3c1285e6b54da02c2d5e07f521dc51f7130176c5c9b4dd72d47c3cdd7f992` |
| RESTORED `SKILL.md` after rollback (9362 B) | `69a3c1285e6b54da02c2d5e07f521dc51f7130176c5c9b4dd72d47c3cdd7f992` |

restored == backup: **YES** (byte-identical).

## Action

Run under the mandated lock:

```
flock ~/.hermes/.wf3.lock -c "cp -p SKILL.md SKILL.md.wf6-fabricating-board-$TS \
  && cp SKILL.md.bak-wf6-20260810T193012Z SKILL.md"
```

Preserved bad copy (not lost, available to diff against the corrected version):
`~/.hermes/skills/orchestration/daily-work-brief/SKILL.md.wf6-fabricating-board-20260810T201445Z`
(18039 B, sha256 `e8cc68e5…`, i.e. identical to the bad file that was live).

## Grep verdicts — restored file is the PRE-WF6 version

| pattern | bad (v1.3.0) | restored (v1.0.0) |
|---|---|---|
| `Board` | 3 | **0** |
| `mcp__linear__` | 2 | **0** |

`grep -n -i "board\|mcp__linear__" SKILL.md` on the restored file returns nothing.
No Board block, not even incidental prose. (`linear` still appears only as a
frontmatter tag — see below.)

Restored frontmatter:

```
---
name: daily-work-brief
description: "Use for concise workday dailies from all activity sources."
version: 1.0.0
metadata:
  hermes:
    tags: [daily, status, slack, github, linear, cooper, payfit]
    category: orchestration
---
```

## Version

**1.3.0 → 1.0.0.**

## Gateway evidence

Before rollback (20:14:36Z):

```
NRestarts=0
ActiveState=active
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

After rollback, at t+45 s (20:15:41Z):

```
NRestarts=0
ActiveState=active
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

`ActiveEnterTimestamp` is **unchanged** — the gateway did **NOT** restart. It is
still the same unit instance that came up at 20:00:38 UTC. (`NRestarts` is not
usable as evidence here: it stayed 0 through the observed 20:00Z restart
earlier today; the timestamp is the real test.)

`~/.local/bin/hermes skills list`: `daily-work-brief | orchestration | local |
local | enabled` — still registered and enabled (124 enabled, 0 disabled).
`hermes skills inspect` only resolves hub sources, so the registered version is
the file's frontmatter: **1.0.0**. `engagement-checker` and `linear-ticketing`
remain enabled and untouched.

## What is live vs what is in the repo

Live on the VM is the **pre-WF6 daily-work-brief v1.0.0** — no Linear board, so
tomorrow's 06:30Z cron cannot deliver a fabricated board.

The repo mirror `skills/daily-work-brief/SKILL.md` is **v1.4.0** (5 `Board`
hits, 2 `mcp__linear__` hits, sha256 `0049cb32…`) — i.e. the repo is ahead of
the VM on purpose and still carries a Board section. It was NOT modified by this
rollback (no git command run, nothing under `skills/` touched). The bad v1.3.0
that was actually live is preserved only on the VM as
`SKILL.md.wf6-fabricating-board-20260810T201445Z`; diff the corrected board
implementation against that, not against the repo's 1.4.0, when redeploying.
