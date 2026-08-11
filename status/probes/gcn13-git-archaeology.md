# GCN-13 — git archaeology: skill self-merge empty-file incidents

Repo: `git@github.com:GaetanCathelain/Tars` (owner `GaetanCathelain`, name `Tars`,
public, default branch `main`). All evidence below is from `origin/main` after
`git fetch origin main` in this worktree (worktree itself is behind and was
never checked out or modified for this investigation).

Investigated 2026-08-11. Scope grew beyond the two named incidents: the same
empty→restore pattern recurs **6 times** across **3 skills** through the
latest commit on `origin/main` at investigation time (PR #55, 16:13:44Z),
i.e. it is still live and unfixed, not a one-off.

## 1. PRs #40–#45 — metadata

All six were authored and merged by the same GitHub account:
**`GaetanCathelain` (Gaetan's own personal account)**, commit author email
`gaetan.cathelain@mobile.club`, merge-commit author line
`GaetanCathelain <7223061+GaetanCathelain@users.noreply.github.com>`. There is
**no separate bot/service account** — Tars' self-merges are indistinguishable
in the git graph from Gaetan personally merging to `main` (`gh pr view --json author`
confirms `is_bot: false`, same account for all 6). All PR bodies are empty
(`gh pr create --fill` inherits the single commit's empty body).

```
$ gh pr view <n> -R GaetanCathelain/Tars --json title,body,mergedAt,headRefName,commits,author
```

| PR | title | headRefName | mergedAt (UTC) | commit oid |
|----|-------|-------------|-----------------|------------|
| 40 | himalaya skill: use the v2 JSON output flag | tars/email-himalaya-20260810T195802Z | 2026-08-10T19:58:46Z | 73ab373 |
| 41 | engagement-checker skill: pin measured Tars Slack user id for self-message filtering | tars/engagement-checker-20260810T202719Z | 2026-08-10T20:27:25Z | 591bf2c |
| 42 | himalaya skill: document Himalaya v2 JSON flag | tars/himalaya-20260810T210431Z | 2026-08-10T21:04:36Z | 0cdf719 |
| 43 | himalaya skill: restore content and document v2 JSON flag | tars/himalaya-20260810T210452Z | 2026-08-10T21:04:58Z | 44b40f5 |
| 44 | linear-ticketing skill: allow measured read-only comment listing | tars/linear-ticketing-20260811T073956Z | 2026-08-11T07:40:02Z | f565995 |
| 45 | linear-ticketing skill: restore file and allow measured comment reads | tars/linear-ticketing-repair-20260811T074016Z | 2026-08-11T07:40:21Z | 2b26296 |

Note PR #40's branch is `tars/email-himalaya-…` (correct nested path,
`skills/email/himalaya/`) while PR #42/#43's branches are `tars/himalaya-…`
(flat, wrong path) — the branch names alone reveal the path drift documented
in §2.

## 2. The two named incidents — byte-exact evidence

### Incident 1: himalaya (PR #42 → #43) — NOT a truncation, a **wrong-path duplicate file**

```
$ git show --stat 241c9d5e   # PR #42 merge commit
 skills/himalaya/SKILL.md | 0
 1 file changed, 0 insertions(+), 0 deletions(-)

$ git show 241c9d5e -- skills/himalaya/SKILL.md
diff --git a/skills/himalaya/SKILL.md b/skills/himalaya/SKILL.md
new file mode 100644
index 0000000..e69de29
```

`e69de29` is git's well-known empty-blob hash — **0 bytes, not whitespace**.
Critically this is a `new file mode 100644`, not a modification: the parent
commit (`c50ec4fe`, main tip immediately before this merge) has **no file at
all** at `skills/himalaya/SKILL.md`. The real, populated himalaya skill file
lives at a **different path**: `skills/email/himalaya/SKILL.md` (category
subdirectory `email/`), last touched by PR #40 (19:58:46Z) and untouched by
#42/#43. `git log --follow` on `skills/himalaya/SKILL.md` misleadingly reports
a 2026-08-07 origin commit — that's git's rename-detection false-positive
(empty/near-empty blobs look "similar" to any nearby deletion); `git log`
without `--follow` confirms only 2 commits ever touched that flat path (#42,
#43).

PR #43's restore commit (`4e4d2dc`) added **305 lines to the flat path**
`skills/himalaya/SKILL.md`, again leaving `skills/email/himalaya/SKILL.md`
untouched:
```
$ git show --stat 4e4d2dc
 skills/himalaya/SKILL.md | 305 +++++++++++++++++++++++++++++++++++++++++++++++
 1 file changed, 305 insertions(+)
```

**Current state on `origin/main`: two divergent copies of the himalaya
skill exist side by side**, and they have already drifted:
```
$ git ls-tree -r origin/main --name-only | grep -i himalaya
skills/email/himalaya/SKILL.md   (7243 bytes, last write PR #40 19:58:46Z)
skills/himalaya/SKILL.md         (7319 bytes, last write PR #43 21:04:58Z)

$ diff flat.md nested.md
278c278
< Most commands support `--json` for structured output (Himalaya v2 uses `--json`, not `--output json`):
---
> Most commands support `--output` for structured output:
```
The flat (wrong-path) copy has the v2-JSON-flag fix; the nested (correct,
originally-tracked) copy still documents the old `--output json` behaviour.
Whichever path Hermes actually loads on the VM determines whether that fix is
live — unresolved by this investigation, flagged for follow-up.

Repo-wide, 2 of 10 skills use a category subdirectory (`skills/email/himalaya/`,
`skills/hermes-operations/hermes-orchestration/`) while everything else,
and the flow in SOUL.md (§3 below), assumes flat `skills/<name>/SKILL.md`.

### Incident 2: linear-ticketing (PR #44 → #45) — genuine truncation of an existing file

```
$ git show --stat f3acd6c   # PR #44 merge commit
 skills/linear-ticketing/SKILL.md | 218 ---------------------------------------
 1 file changed, 218 deletions(-)

$ git show f3acd6c -- skills/linear-ticketing/SKILL.md | head -6
diff --git a/skills/linear-ticketing/SKILL.md b/skills/linear-ticketing/SKILL.md
index e2237da..e69de29 100644
```
Real truncation this time: existing blob `e2237da` (218 lines) → empty blob
`e69de29` (0 bytes). Restore (PR #45, `26f5580`) put 220 lines back:
```
$ git show --stat 26f5580
 skills/linear-ticketing/SKILL.md | 220 +++++++++++++++++++++++++++++++++++++++
 1 file changed, 220 insertions(+)
```

### Gap between empty-merge and restore-merge (all instances, see §5)

| pair | empty merged (UTC) | restore merged (UTC) | gap |
|------|---------------------|------------------------|-----|
| himalaya #42→#43 | 21:04:36Z | 21:04:58Z | **22s** |
| linear-ticketing #44→#45 | 07:40:02Z | 07:40:21Z | **19s** |
| engagement-checker #47→#48 | 09:11:45Z | 09:12:07Z | **22s** |
| linear-ticketing #49→#50 | 15:11:33Z | 15:11:58Z | **25s** |
| engagement-checker #52→#53 | 15:31:49Z | 15:32:24Z | **35s** |
| linear-ticketing #54→#55 | 16:13:23Z | 16:13:44Z | **21s** |

Consistently 19–35s: Tars runs the full SOUL-rule-2 script twice back to back
— the flow has no verification step that would let it check the first
attempt's outcome from further away, so this is almost certainly the same
turn immediately re-running after the agent (or a fast follow-up read)
notices the file it just wrote is empty.

## 3. SOUL rule 2 — verbatim (origin/main `SOUL.md`, unchanged since 2026-08-10T10:53:54Z,
i.e. **before** all 6 incidents — never amended in response)

```
2. I never merge, approve or push — with exactly one exception, spelled out
   below. Reading a pull request, a diff or a CI log is how I verify what I
   delegated — that is my job, not a breach of it.

   **The exception is my own operating record.** My skills are mine to change
   with `skill_manage` when a run teaches me something the file gets wrong —
   that is how I stop repeating a mistake. But an edit is **not finished until it
   is committed**: an uncommitted change to my own instructions is drift nobody
   can review, and the next person who reconciles the file silently destroys it.
   So immediately after any `skill_manage` write, in the same turn:

   ```bash
   N=<skill-name>
   ssh cooper "cd ~/dev/Tars && git checkout -q main && git pull --rebase -q origin main"
   cat ~/.hermes/skills/$N/SKILL.md | ssh cooper "mkdir -p ~/dev/Tars/skills/$N && cat > ~/dev/Tars/skills/$N/SKILL.md"
   ssh cooper "cd ~/dev/Tars && git checkout -q -b tars/$N-\$(date -u +%Y%m%dT%H%M%SZ) \
     && git add skills/$N/SKILL.md \
     && git commit -q -m '$N skill: <what I changed and why, one line>' \
     && git push -q -u origin HEAD \
     && gh pr create --fill \
     && gh pr merge --squash --delete-branch \
     && git checkout -q main && echo MERGED"
   ```

   Pull FIRST, then copy the file: copying before the pull leaves the tree
   dirty and `git pull --rebase` refuses it (measured 2026-08-07, twice).
   `~/dev/Tars` on cooper is the Tars repo with a push remote and `gh` logged
   in; the repo path `skills/<name>/SKILL.md` mirrors my live
   `~/.hermes/skills/<name>/SKILL.md`, and the pull request is the reviewable
   copy. I say in my reply that I changed the skill, what I changed, and give
   the PR URL and that it merged. If any step fails I say so plainly, never
   force anything, and never merge with `--admin`.

   This exception covers **my own skills and nothing else.** It is not licence to
   push in a repo I was delegated work in — there, rule 2 stands whole.
```

Mechanically relevant: `cat ~/.hermes/skills/$N/SKILL.md | ssh cooper "... cat > ..."`
truncates the destination the instant the remote shell opens the file
(`cat >` is `O_TRUNC`), *before* any bytes have to arrive over the pipe/ssh
hop, and the script has **no check between the copy step and the
add/commit/push/merge step** that the copied content is non-empty (no
`test -s`, no diff, no `git diff --stat` gate). If the local read
(`~/.hermes/skills/$N/SKILL.md`) is momentarily empty/mid-write on the VM
side, or the ssh pipe stalls, the destination is committed and merged empty
with nothing to catch it. This matches every truncation incident. It does
**not** explain the himalaya path-drift incident (§2, Incident 1) — that one
is `N=himalaya` being the wrong `$N` for a skill actually tracked at
`skills/email/himalaya/`, so the script created a sibling flat file instead
of touching the real one.

`CLAUDE.md` at `origin/main` (§Repo map) documents the same flat convention:
```
skills/<n>/SKILL.md  live mirror of ~/.hermes/skills/<n>/SKILL.md
```
and (§Tars edits its own skills…):
```
**Tars edits its own skills, and lands them here.** `skill_manage` lets Tars
rewrite its own `SKILL.md` when a run teaches it something the file gets wrong —
it did so 7 times during the first live Orca run. **SOUL rule 2 requires it to
land that edit on `skills/<name>/SKILL.md` via a self-merged PR (branch →
`gh pr create` → squash-merge), in the same turn.**
```
`docs/facts.md` has no dedicated skill-self-merge entry; the only adjacent
fact is line 21: "Skills live at `~/.hermes/skills/<name>/SKILL.md` and are
loaded into model context unprompted when relevant."

## 4. CI / branch protection

```
$ git ls-tree -r origin/main --name-only | grep '^\.github'
(no output — no .github/workflows, no CI at all)

$ gh api repos/GaetanCathelain/Tars/branches/main/protection
{"message":"Branch not protected","documentation_url":"...","status":"404"}
```
**No CI, no branch protection on `main`.** Nothing would have blocked either
empty-file merge; there is no required check, no PR review requirement, and
`main` accepts direct pushes. Repo is public (`gh api repos/... -q .private` → `false`).

## 5. Other skill-mirror commits on `origin/main` since 2026-08-09 (backprop volume)

Full self-merge PR list, `main`, chronological (`gh pr list --state all`):

| PR | merged (UTC) | title | pattern |
|----|--------------|-------|---------|
| 34 | 2026-08-09T11:42:50Z | engagement-checker: use direct Linear GraphQL reads | normal |
| 35 | 2026-08-10T11:01:29Z | daily-work-brief: follow configured reporting conversation | normal |
| 36 | 2026-08-10T11:01:59Z | engagement-checker: read decisions from reporting conversation | normal |
| 37 | 2026-08-10T13:26:29Z | Daily report threading (not a Tars self-merge — different flow) | — |
| 38 | 2026-08-10T13:34:10Z | Rename reconciler script (not a Tars self-merge) | — |
| 39 | 2026-08-10T19:58:48Z | Slack tool visibility correction (not a Tars self-merge) | — |
| 40 | 2026-08-10T19:58:46Z | himalaya: use the v2 JSON output flag | normal (correct path `skills/email/himalaya/`) |
| 41 | 2026-08-10T20:27:25Z | engagement-checker: pin Tars Slack user id | normal |
| **42→43** | 21:04:36Z→21:04:58Z | himalaya: document v2 JSON flag / restore | **empty→restore (wrong path, §2 Incident 1)** |
| **44→45** | 07:40:02Z→07:40:21Z | linear-ticketing: allow read-only comment listing / restore | **empty→restore (truncation, §2 Incident 2)** |
| 46 | 2026-08-11T08:30:28Z | helptech-duty-intake: capture verified B2C duty intake workflow | normal |
| **47→48** | 09:11:45Z→09:12:07Z | engagement-checker: attach Gmail permalinks / restore | **empty→restore (truncation, 710→0→712 lines)** |
| **49→50** | 15:11:33Z→15:11:58Z | linear-ticketing: document duplicate relation semantics / restore | **empty→restore (truncation, 220→0→220 lines)** |
| 51 | 2026-08-11T15:28:48Z | helptech-duty-intake: prevent duplicate worktrees/races | normal |
| **52→53** | 15:31:49Z→15:32:24Z | engagement-checker: preserve unthreaded DM handoffs / restore | **empty→restore (truncation, 712→0→714 lines)** |
| **54→55** | 16:13:23Z→16:13:44Z | linear-ticketing: record delegated session provenance / restore | **empty→restore (truncation, 222→0→222 lines)** |

**Six empty→restore pairs total (not two)**, across three skills
(himalaya ×1, linear-ticketing ×3, engagement-checker ×2), spanning
2026-08-10T21:04Z through 2026-08-11T16:13Z — the most recent one is the
*latest commit on `origin/main` at investigation time*. The bug is currently
live and unfixed; SOUL.md has not been touched since before the first
occurrence (§3).

## Summary of what each empty-PR's commit message claimed

None of the 6 "empty" commit messages mention emptying, deleting, or any
data-loss risk — every one reads as a normal, additive skill update
("document Himalaya v2 JSON flag", "allow measured read-only comment
listing", "attach Gmail permalinks to mail-sourced tickets", "document
duplicate relation update semantics", "preserve unthreaded DM handoffs",
"record delegated session provenance and start state"). The commit message
is generated by the same turn that wrote the (accidentally empty) content, so
it describes the *intended* change, not what actually landed — there is no
self-detection of the truncation until the very next commit, seconds later.
