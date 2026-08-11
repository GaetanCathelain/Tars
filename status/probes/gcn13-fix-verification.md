# GCN-13 — independent fix verification (2026-08-11)

Verified from primary state (live VM `~/.hermes/`, `origin/main`), read-only
throughout. Not derived from `status/probes/gcn13-fix-applied.md` claims —
that file was read only to know what to check. `git fetch origin main` run
first; all repo comparisons are against `origin/main` at `796a1d1`
(`796a1d133a9d2a97863079ed8a1775b5298d90ca`), which matches this worktree's
`HEAD` (already up to date, nothing to rebase).

## Overall verdict: ALL PASS (1a limited scope noted, not a defect)

## 1. Live skills vs repo mirror — byte identity

Live tree carries **128** `SKILL.md` files (`ssh gaetan@192.168.0.9 find
~/.hermes/skills -name SKILL.md`). The repo mirror at `origin/main` tracks
**9 skills / 13 files** (`git ls-tree -r origin/main --name-only skills/`) —
the skills Tars has ever self-edited via `skill_manage`, per SOUL rule 2's
mirror recipe. This scope is explicit and disclosed in the fixer's own
report §4 ("119 live skills are not in the repo and were not added ... Say
the word if the mirror should cover the whole live tree") — a flagged
decision for Gaetan, not a fix defect. Verification below therefore checks:
**every file that exists in the repo mirror is byte-identical to its live
counterpart at the correct relative path**, which is the literal claim the
fix makes.

md5 of live file vs `git show origin/main:skills/<rel> | md5sum`, all 13:

```
92505150cd08004f7049a147659cd9f5  delegate-to-cooper/SKILL.md                                    MATCH
be1ae34eda8b8dd3150fdb7987a9e718  email/himalaya/references/configuration.md                      MATCH
b22b649592a8008fa39ff2b65eb2e0e6  email/himalaya/references/message-composition.md                MATCH
e72577a7069aeeed814e3b9c979d2b18  email/himalaya/SKILL.md                                          MATCH
d9e2f232dd8d1341d914a66cf8ab8696  helptech-duty-intake/SKILL.md                                    MATCH
e13a2c6af5ac5792c2dfec97da1be8c2  hermes-operations/hermes-orchestration/SKILL.md                  MATCH
f71e4eb89c40d520ee0d6662579be83e  orchestration/daily-work-brief/scripts/linear_board.py           MATCH
00654bacc7634c4650e3fa1fc254f771  orchestration/daily-work-brief/SKILL.md                          MATCH
a117cf494ea53d18a0fd1ffbc1b2310b  orchestration/engagement-checker/SKILL.md                        MATCH
3a6303ad4ad5bd5efc21e0b2bd985631  orchestration/linear-ticketing/SKILL.md                          MATCH
34e5ba011948641932c1c990ac48ff22  orchestration/secure-delta-collectors/references/slack-oauth-user-token.md  MATCH
e83202c15af6d5a1560f4a1eaf889050  orchestration/secure-delta-collectors/references/slack-web-api.md            MATCH
7b3a6f869ce660946c275d8f7bf644bd  orchestration/secure-delta-collectors/SKILL.md                   MATCH
```

13/13 md5 pairs identical, all at the correct relative path (`skills/` +
live path relative to `~/.hermes/skills/`). No repo file exists at a *wrong*
relative path for any of these 9 skills (checked in §2). **PASS** (for the
tracked mirror set; the 119-skill scope gap is a disclosed, undecided
product question, not a verification failure — flagged here for Gaetan same
as the fixer flagged it).

## 2. No duplicate SKILL.md for the same skill name

```
$ git ls-tree -r origin/main --name-only skills/ | grep -i himalaya
skills/email/himalaya/SKILL.md
skills/email/himalaya/references/configuration.md
skills/email/himalaya/references/message-composition.md

$ git cat-file -e origin/main:skills/himalaya/SKILL.md
fatal: Not a valid object name origin/main:skills/himalaya/SKILL.md
```

`skills/himalaya/SKILL.md` (flat orphan) is gone; `skills/email/himalaya/SKILL.md`
present. No other skill name appears at two different paths under `skills/`
(cross-checked visually against the 13-file listing in §1 — one path per
skill). **PASS**

## 3. secure-delta-collectors mirror matches live, version line included

```
-- repo (origin/main:skills/orchestration/secure-delta-collectors/SKILL.md) --
version: 1.4.0
-- live (~/.hermes/skills/orchestration/secure-delta-collectors/SKILL.md) --
version: 1.4.0
```

Full-file md5 also matches (§1 row `7b3a6f869ce660946c275d8f7bf644bd`, both
sides). **PASS**

## 4. SOUL.md: byte-identical, today's `.bak`, rule 2 content

```
$ ssh gaetan@192.168.0.9 md5sum ~/.hermes/SOUL.md
b7e8d78356807a8cf9abb84696979054  /home/gaetan/.hermes/SOUL.md
$ git show origin/main:SOUL.md | md5sum
b7e8d78356807a8cf9abb84696979054  -
```
Byte-identical. **PASS**

```
$ ssh gaetan@192.168.0.9 ls -la ~/.hermes/SOUL.md.bak*
...
-rw-rw-r-- 1 gaetan gaetan 7909 Aug 11 16:37 /home/gaetan/.hermes/SOUL.md.bak-20260811
```
`.bak-20260811` present, dated today (16:37, just before the 16:38-16:40 fix
commits land). **PASS**

Rule 2 new text (`origin/main:SOUL.md`) contains all five required elements,
grep excerpts:

- single-match path resolution: `[ "$(echo $REL | wc -w)" = 1 ] || { echo "STOP: $N resolves to [$REL]"; exit 1; }`
- `test -s` / `pipefail` / tmp+mv guards: `set -euo pipefail`; `test -s ~/.hermes/skills/"$REL"`; `cat > ~/dev/Tars/skills/$REL.new && test -s ... && mv ... $REL`
- whole-skill (not diff) pre-merge review: "Before I merge I read the full `SKILL.md` as it will exist after the merge — the whole skill, not the diff."
- post-merge byte-equality against live: `git show origin/main:skills/$REL" | diff - ~/.hermes/skills/"$REL" && echo MIRRORED`; prose: "must be byte-identical to the live file; if it is not, I fix it forward immediately, in the same turn."
- same-turn mandate: "So immediately after any `skill_manage` write, in the same turn:" / "I fix it forward immediately, in the same turn."

All five present. **PASS**

## 5. Docs: flat mapping no longer stated as the convention

`CLAUDE.md` at `origin/main` (repo map section):
```
  SOUL.md           live mirror of ~/.hermes/SOUL.md on the VM
  skills/<rel>       live mirror of ~/.hermes/skills/<rel> — SAME relative path,
```
States the relative-path convention, not flat `skills/<n>/SKILL.md`. **PASS**

`grep -rn 'skills/<' docs/` at `origin/main` turned up 4 hits total, none of
which assert the flat layout as *this repo's* current convention:

- `docs/facts.md:21` — already corrected, states the relative-path rule
  explicitly ("The repo mirror path is `skills/` + that same relative path
  (SOUL rule 2, amended 2026-08-11 after GCN-13)").
- `docs/specs/wf5-orca-delegation.md:393` — already correct:
  `skills/<category>/<name>/SKILL.md`, category-qualified.
- `docs/proposals/P6-knowledge-bases.md:206` — `Tars' own skills/<n>/SKILL.md
  mirror` used generically inside a sentence about rule 2's authorship
  carve-out (not a directory-layout claim); pre-existing proposal prose, not
  reworded by this fix. Borderline but not a factual claim about the
  convention.
- `docs/recon/r9-evidence/external-prior-art.md:179` — quotes the *upstream*
  NousResearch `hermes-agent` project's own documented layout
  (`~/.hermes/skills/<name>/SKILL.md`), unrelated to Tars' repo-mirror
  convention; recon evidence snapshot, not meant to be kept current.

No genuine leftover claim that this repo's mirror convention is flat.
**PASS**

## 6. hermes cron — drift-watchdog job

```
$ ssh gaetan@192.168.0.9 ~/.local/bin/hermes cron list
...
  45c5c8ba10ee [active]
    Name:      Skill mirror drift check
    Schedule:  0 18 * * *
    Repeat:    ∞
    Next run:  2026-08-12T18:00:00+02:00
    Deliver:   slack
```
Job exists, active, schedule matches claim (`0 18 * * *`). **PASS**

## 7. No new empty SKILL.md blob committed after the fix

```
$ git log --format="%H %ai %s" origin/main -5
796a1d1 2026-08-11 16:40:44 +0000 docs(status): GCN-13 probe evidence and applied-fix record
5f29295 2026-08-11 16:39:02 +0000 docs: record the relative-path skill mirror convention (GCN-13)
d31dc8b 2026-08-11 16:38:45 +0000 fix(soul): rule 2 mirrors skills by live relative path, guarded (GCN-13)
bbb21fa 2026-08-11 16:38:32 +0000 fix(skills): resync mirrored skills to live VM content (GCN-13)
46b56c3 2026-08-11 16:38:09 +0000 refactor(skills): mirror skills at their live category paths (GCN-13)
```
The newest commit on `origin/main` (`796a1d1`) is at **16:40:44Z** — before
the 17:00Z threshold. There are **zero commits on `origin/main` after
2026-08-11T17:00Z** (`TZ=UTC git log --since="2026-08-11 17:00:00"` returns
nothing), so no post-fix commit could have introduced a regression.

Independently confirmed no empty blob exists anywhere in the *current*
`skills/` tree at `origin/main` (belt-and-suspenders, not just "since a
timestamp"):

```
$ git ls-tree -r origin/main skills/ | awk '{print $3}' > hashes.txt
$ git cat-file --batch-check='%(objectname) %(objectsize) %(rest)' < hashes.txt
961ea267... 34794
16350ca7... 7319
5ccba6cb... 5906
2dbd7a99... 3799
bef61857... 10475
256e597e... 9753
b84ef69e... 18304
03202fc0... 17626
14a2bbe4... 69654
76e3baf0... 25022
7fc69982... 11296
e8d10179... 3641
6239a533... 5274
```
All 13 blobs have real, non-zero sizes; none is `e69de29bb2d1d6434b8b29ae775ad8c2e48c5391`
(git's canonical empty-blob hash). **PASS**

(Note: an earlier attempt at this check via a `while read` loop piping
`git ls-tree` into per-line `git cat-file -s` calls produced spurious
"command not found: git" errors and false "EMPTY" lines for all 13 files —
an artifact of the local `rtk` shell-hook rewriting nested `git` invocations
inside the loop, not a real finding. Discarded; the `--batch-check` form
above reads from a file and is not subject to that interference.)

## Summary

| # | Check | Verdict |
|---|-------|---------|
| 1 | Live↔repo mirror byte-identity (tracked set) | PASS (119/128 live skills intentionally unmirrored — disclosed scope decision, not a defect) |
| 2 | No duplicate himalaya SKILL.md | PASS |
| 3 | secure-delta-collectors mirror + version line | PASS |
| 4 | SOUL.md identity, `.bak`, rule 2 content (5/5 elements) | PASS |
| 5 | Docs no longer state flat convention | PASS |
| 6 | hermes cron drift-watchdog job live | PASS |
| 7 | No new empty SKILL.md blob after the fix | PASS |
