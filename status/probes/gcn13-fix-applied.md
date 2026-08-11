# GCN-13 — fix applied (2026-08-11)

Root cause and evidence: `status/probes/gcn13-git-archaeology.md` (GitHub side)
and `status/probes/gcn13-vm-skill-drift.md` (VM side). Both converge on SOUL
rule 2's mirror recipe: `$N` is a bare skill name, most live skills sit under a
category dir, so `cat ~/.hermes/skills/$N/SKILL.md` read nothing while the
remote `cat >` truncated the destination — six empty SKILL.md files merged to
`main` across three skills.

**Settled principle of the fix:** the live tree `~/.hermes/skills/` is the single
source of truth; the repo path is `skills/` + the file's path relative to
`~/.hermes/skills/`. No name-to-path mapping anywhere.

## 1. SOUL rule 2 — old vs new

Backup of the pre-amendment file on the VM: `~/.hermes/SOUL.md.bak-20260811`
(md5 `858d1f0aa6f4891d73a00bf615e0db30`). New live file md5
`b7e8d78356807a8cf9abb84696979054`, byte-identical to repo `SOUL.md` at
`d31dc8b`. Repo `SOUL.md` and the VM copy were verified identical *before* the
edit (`diff` clean), so nothing else had moved on either side.

### OLD (verbatim, the amended part only)

````
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
````

### NEW (verbatim, live on the VM and at repo `SOUL.md`)

````
   ```bash
   set -euo pipefail
   N=<skill-name>

   # Resolve the LIVE file first — its own path is the mirror path.
   REL=$(cd ~/.hermes/skills && find . -path "*/$N/SKILL.md" -printf '%P\n')
   [ "$(echo $REL | wc -w)" = 1 ] || { echo "STOP: $N resolves to [$REL]"; exit 1; }
   test -s ~/.hermes/skills/"$REL"

   ssh cooper "cd ~/dev/Tars && git checkout -q main && git pull --rebase -q origin main \
     && mkdir -p skills/$(dirname "$REL")"
   cat ~/.hermes/skills/"$REL" | ssh cooper "cat > ~/dev/Tars/skills/$REL.new \
     && test -s ~/dev/Tars/skills/$REL.new && mv ~/dev/Tars/skills/$REL.new ~/dev/Tars/skills/$REL"

   ssh cooper "cd ~/dev/Tars && git checkout -q -b tars/$N-\$(date -u +%Y%m%dT%H%M%SZ) \
     && git add skills/$REL \
     && git commit -q -m '$N skill: <what I changed and why, one line>' \
     && git push -q -u origin HEAD \
     && gh pr create --fill"
   # ← I read the whole file here, then merge and prove what landed:
   ssh cooper "cd ~/dev/Tars && gh pr merge --squash --delete-branch \
     && git checkout -q main && git pull --rebase -q origin main \
     && git show origin/main:skills/$REL" | diff - ~/.hermes/skills/"$REL" && echo MIRRORED
   ```

   The mirror path IS the live path: `skills/` + the file's path relative to
   `~/.hermes/skills/`, so `orchestration/linear-ticketing/SKILL.md` lands at
   `skills/orchestration/linear-ticketing/SKILL.md`. No name-to-path mapping
   anywhere; zero or several matches means I stop and say so rather than guess.
   Pull FIRST, then copy the file: copying before the pull leaves the tree
   dirty and `git pull --rebase` refuses it (measured 2026-08-07, twice). Never
   a bare `>` onto the destination — it truncates it before a single byte
   arrives.

   **Before I merge I read the full `SKILL.md` as it will exist after the merge
   — the whole skill, not the diff.** The diff is what I meant to change; the
   file is what I will be running on. After the merge,
   `git show origin/main:skills/$REL` must be byte-identical to the live file;
   if it is not, I fix it forward immediately, in the same turn.

   `~/dev/Tars` on cooper is the Tars repo with a push remote and `gh` logged
   in, and the pull request is the reviewable copy. I say in my reply that I
   changed the skill, what I changed, and give the PR URL and that it merged.
   If any step fails I say so plainly, never force anything, and never merge
   with `--admin`. (Amended 2026-08-11, GCN-13: the old recipe assumed a flat
   `skills/<name>/` layout, so for a categorized skill the local `cat` read
   nothing while the remote `>` emptied the destination — six empty files
   merged before it was caught.)
````

The resolver was exercised read-only on the VM before the rule was written:

```
OK   linear-ticketing        -> orchestration/linear-ticketing/SKILL.md
OK   himalaya                -> email/himalaya/SKILL.md
OK   delegate-to-cooper      -> delegate-to-cooper/SKILL.md
OK   hermes-orchestration    -> hermes-operations/hermes-orchestration/SKILL.md
OK   secure-delta-collectors -> orchestration/secure-delta-collectors/SKILL.md
STOP nosuchskill             -> []
STOP SKILL.md                -> []
```

## 2. Repo changes (branch `GaetanCathelain/improvements` → `main`)

| commit | change |
|---|---|
| `46b56c3` | `refactor(skills): mirror skills at their live category paths` — `git mv` of `daily-work-brief`, `engagement-checker`, `linear-ticketing`, `secure-delta-collectors` (with their `scripts/`, `references/`) from flat `skills/<n>/` to `skills/orchestration/<n>/`; `git rm skills/himalaya/SKILL.md` (the wrong-path orphan). History preserved (`R` in `git status`). |
| `bbb21fa` | `fix(skills): resync mirrored skills to live VM content` — 6 files, +553/−16. himalaya v2 `--json` fix that had only ever landed on the orphan path now on `skills/email/himalaya/SKILL.md`; `secure-delta-collectors` v1.1.0 → v1.4.0 (4 days unbackpropagated); adds `email/himalaya/references/{configuration,message-composition}.md` and `secure-delta-collectors/references/slack-oauth-user-token.md`, which exist live but were missing here. |
| `d31dc8b` | `fix(soul): rule 2 mirrors skills by live relative path, guarded` — repo mirror of the amended VM `SOUL.md`. |
| `5f29295` | `docs: record the relative-path skill mirror convention` — `CLAUDE.md` repo map + "Tars edits its own skills" paragraph, `docs/facts.md` skills row. |
| (this commit) | probe evidence `gcn13-git-archaeology.md`, `gcn13-vm-skill-drift.md`, and this file. |

`docs/specs/wf6-linear-integration.md` already writes `~/.hermes/skills/orchestration/*`
(line 82) — correct as-is, not touched.

### Post-push verification — every mirrored file byte-identical to live

`git show origin/main:<path> | md5sum` vs `md5sum ~/.hermes/skills/<rel>` on the VM,
all 13 tracked files, after `5f29295` landed on `main`:

```
MATCH delegate-to-cooper/SKILL.md
MATCH email/himalaya/SKILL.md
MATCH email/himalaya/references/configuration.md
MATCH email/himalaya/references/message-composition.md
MATCH helptech-duty-intake/SKILL.md
MATCH hermes-operations/hermes-orchestration/SKILL.md
MATCH orchestration/daily-work-brief/SKILL.md
MATCH orchestration/daily-work-brief/scripts/linear_board.py
MATCH orchestration/engagement-checker/SKILL.md
MATCH orchestration/linear-ticketing/SKILL.md
MATCH orchestration/secure-delta-collectors/SKILL.md
MATCH orchestration/secure-delta-collectors/references/slack-oauth-user-token.md
MATCH orchestration/secure-delta-collectors/references/slack-web-api.md
```

Backups on the VM (`SKILL.md.bak-*`, `SKILL.md.wf6-*`) are deliberately excluded
from the mirror.

## 3. Drift watchdog — created

`hermes cron create --help` run on the VM first (never trusting recon syntax);
a daily schedule is expressible. Created:

```
Created job: 45c5c8ba10ee
  Name: Skill mirror drift check
  Schedule: 0 18 * * *
  Next run: 2026-08-12T18:00:00+02:00
  Deliver: slack   (bare platform ⇒ SLACK_HOME_CHANNEL)
```

Prompt: fetch `origin/main` on cooper, `git ls-tree -r origin/main --name-only skills/`,
compare each `skills/REL` blob byte-for-byte (md5) against live `~/.hermes/skills/REL`,
flag repo paths with no live file, backpropagate every drifted file per SOUL rule 2
(live wins) in the same turn, DM Gaetan the drift list with PR URLs; one line
"skills in sync" when clean. No script file was added on the VM — the agent
prompt carries the whole check.

## 4. Flags / needs-Gaetan-decision

- **119 live skills are not in the repo and were not added.** The brief said
  "every live skill on the VM"; the VM carries 128 `SKILL.md` files, of which
  only the 9 Tars has ever self-edited were ever mirrored. The rest are bundled
  hub/third-party skills (`creative/*`, `ponytail*`, `mattpocock-*`,
  `productivity/*`, …). Importing them would bulk-add third-party content to a
  public repo for no operational gain, so the sync was scoped to the tracked
  mirror. Say the word if the mirror should cover the whole live tree.
- **No repo skill file lacks a live counterpart** other than the deleted
  `skills/himalaya/SKILL.md` orphan — nothing was left needing a delete decision.
- **No CI, no branch protection on `main`** (probe §4 of the archaeology file).
  The new rule-2 verification is the only gate against an empty mirror commit;
  it is prompt-level, not enforced. A one-line `pre-receive`-equivalent does not
  exist for this repo — worth a GitHub Action rejecting empty `skills/**/SKILL.md`
  blobs if this recurs.
- **Old branch names / merged PRs are untouched.** The six empty→restore PR pairs
  stay in history as-is; no history rewrite was attempted.
