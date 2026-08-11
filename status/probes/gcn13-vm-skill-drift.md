# GCN-13 — VM-side investigation: skill self-merge empty-file incidents

Investigated 2026-08-11 from the Tars VM (`ssh gaetan@192.168.0.9`, clock UTC,
confirmed via `date -u` = `Tue Aug 11 04:18:52 PM UTC 2026` at investigation
start). Read-only throughout: no file on the VM was created, edited, or
deleted; no `sops -d`; no secret file opened. Companion probe
`status/probes/gcn13-git-archaeology.md` covers the GitHub/PR side in more
depth and independently converges on the same root cause found here from the
VM/log side — the two corroborate each other.

## 1. Tars' local checkout on the VM — there isn't one

`ls -la ~` and a full `find /home/gaetan -name .git` (node_modules/.cache
excluded) turned up only unrelated repos: `~/dev/mc-metarepo`,
`~/.hermes/plugins/hermes-lcm`, `~/.hermes/hermes-agent`
(`NousResearch/hermes-agent` upstream source, not the Tars ops repo), plus
throwaway clones in `/tmp` (`hermes-agent-docs`, `mattpocock-skills-verify-*`,
`smcp`). None has `origin` pointing at `GaetanCathelain/Tars`:

```
$ ssh gaetan@192.168.0.9 'find /home/gaetan -name ".git" -type d -not -path "*/node_modules/*" -not -path "*/.cache/*"'
/home/gaetan/dev/mc-metarepo/.git
/home/gaetan/.hermes/plugins/hermes-lcm/.git
/home/gaetan/.hermes/hermes-agent/.git
```

A filesystem-wide `find / -xdev -name .git` (single device, `/proc` excluded)
adds nothing more. `~/.hermes/sandboxes/singularity` (the one persistent
"sandbox" dir referenced by the terminal tool) is empty. There is **no
persistent checkout to run `git status --porcelain` or `git reflog` against**
— so step 1's literal ask (reflog bracketing the two incidents) is
unanswerable as posed, and that absence is itself the finding.

Why: `agent.log` shows the terminal/file tools create an **ephemeral, torn-down-per-session**
working directory, not a durable checkout:

```
2026-08-10 19:54:41,492 INFO [...] tools.terminal_tool: Creating new local environment for task default...
2026-08-10 19:54:41,5xx INFO [...] tools.terminal_tool: local environment ready for task default
...
2026-08-10 19:59:06,523 INFO [...] tools.terminal_tool: Manually cleaned up environment for task: default
2026-08-10 19:59:06,524 INFO [...] tools.terminal_tool: Cleaned 1 environments
```

Every session in both incident windows follows this create→use→"Manually
cleaned up"/"Cleaned 1 environments" lifecycle. `git` and `gh` **are**
installed locally on the VM (`git version 2.43.0`, `gh version 2.97.0`), so
the ephemeral env could in principle run git locally — but the actual
mechanism Tars uses (found in `~/.hermes/SOUL.md`, see §3) is to **shell out to
`cooper` over ssh** and operate on `~/dev/Tars` there, not on the VM at all.
Corroborating evidence: `~/.local/share/rtk/tee/1786462153_ssh_cooper_cd___dev_Tars____gh_pr_view_5.log`
on the VM (rtk's own command-tee cache, itself proof the VM's `gaetan` shell
runs through the same `rtk` proxy as the interactive user) contains:
```json
{"mergeCommit":"2b28885...","mergedAt":"2026-08-11T15:28:48Z","state":"MERGED","url":"https://github.com/GaetanCathelain/Tars/pull/51"}
```
— matching PR #51 exactly (see §4/§5 below), confirming Tars really does run
`gh pr view` against a checkout that lives on **cooper**, not on the VM. The
rtk tee cache only retains today's entries (oldest 15:08Z), so it doesn't
reach back to either incident window — not useful for reconstructing them,
only for confirming the mechanism.

## 2. Reconstructing the two named incidents from `~/.hermes/logs/agent.log`

`agent.log` (current, not the `.1` rotation) covers both windows end-to-end.
Log design limits what's recoverable: **INFO-level tool completions only log
a byte count** (`tool X completed (Ns, N chars)`), not the arguments or
output; only **WARNING-level failures** print the actual JSON payload
(itself truncated ~300 chars by Hermes' own logger — confirmed by finding a
line that cuts off mid-URL with no trailing quote/brace). So exact
`skill_manage`/`terminal` command text is not recoverable from this log for
successful calls — order, timing, and sizes are.

### Incident 1 — himalaya, session `20260810_210250_306801` (21:02:54–21:06:25Z)

This is the actual empty-file session; the earlier 19:57–19:59Z session
(`20260810_195424_fc192c`) that I initially flagged as "incident 1" is a
**different, successful** first-ever himalaya commit (PR #40, ADDED,
304/0 lines — see §4) and is not the incident.

```
2026-08-10 21:03:11,517 INFO [...] tools.terminal_tool: Creating new local environment for task default...
2026-08-10 21:04:17,482 INFO [...] agent.tool_executor: tool terminal completed (0.05s, 4530 chars)
2026-08-10 21:04:17,532 INFO [...] agent.tool_executor: tool skill_manage completed (0.04s, 449 chars)
2026-08-10 21:04:28,356 INFO [...] agent.conversation_loop: API call #7 ...
2026-08-10 21:04:40,196 INFO [...] agent.tool_executor: tool terminal completed (11.82s, 701 chars)
2026-08-10 21:04:50,339 INFO [...] agent.conversation_loop: API call #8 ...
2026-08-10 21:05:01,976 INFO [...] agent.tool_executor: tool terminal completed (11.63s, 612 chars)
```
`skill_manage` completes at **21:04:17** (write to the *live* file). GitHub:
PR #42 created **21:04:34**, merged **21:04:36** — inside the 21:04:28→21:04:40
(11.82s) terminal call, which is long enough to be the whole
`ssh cooper "checkout+add+commit+push+gh pr create+gh pr merge"` one-liner
from SOUL.md §3. The next terminal call (21:04:50→21:05:01, 11.63s)
correlates with the restore, PR #43, created 21:04:55, merged 21:04:58 — i.e.
**skill_manage write → git mirror (empty) → self-detected → git mirror
(restored), all inside one continuous session, ~40 seconds total.** Order is
write-then-push in both rounds; there is no visible truncate-then-write
*inside* `skill_manage` itself (see §3 for where the truncation actually
happens — it's in the mirror step, not the tool).

### Incident 2 — linear-ticketing, session `20260811_073904_7d8ab5` (07:39:07–07:40:41Z)

```
2026-08-11 07:39:34,937 INFO [...] agent.tool_executor: tool mcp__linear__list_comments completed ...
2026-08-11 07:39:46,242 INFO [...] agent.tool_executor: tool skill_manage completed (0.02s, 530 chars)
2026-08-11 07:39:54,231 INFO [...] tools.terminal_tool: Creating new local environment for task default...
2026-08-11 07:40:05,733 INFO [...] agent.tool_executor: tool terminal completed (11.52s, 714 chars)
2026-08-11 07:40:24,793 INFO [...] agent.tool_executor: tool terminal completed (10.67s, 642 chars)
2026-08-11 07:40:33,266 INFO [...] agent.tool_executor: tool terminal completed (2.87s, 325 chars)
2026-08-11 07:40:41,620 INFO [...] agent.conversation_loop: Turn ended...
```
Same shape: `skill_manage` completes once (07:39:46), *then* the ephemeral
terminal env is created (07:39:54) and three terminal calls run the mirror
(matches PR #44 created/merged 07:40:00/07:40:02, and the restore PR #45
created/merged 07:40:19/07:40:21 — both inside this one session, ~50 seconds
total). Same write-before-push ordering as incident 1.

### This is not a two-incident story — it recurred twice more today

`gh pr list` shows the identical empty→restore pair happened **twice more**
after the ticket was filed, both for `linear-ticketing`, both self-corrected
within ~20–30s the same way:

| wipe PR | merged | restore PR | merged |
|---|---|---|---|
| #44 (0 add / 218 del) | 07:40:02Z | #45 (220 add / 0 del) | 07:40:21Z |
| #49 (0 add / 220 del) | 15:11:33Z | #50 (222 add / 0 del) | 15:11:58Z |
| #54 (0 add / 222 del) | 16:13:23Z | #55 (232 add / 0 del) | 16:13:44Z |

Sizes above from `gh pr view <n> --repo GaetanCathelain/Tars --json additions,deletions,files`.
The bug is **live and unfixed as of this investigation**, not a closed
two-occurrence incident.

## 3. Root cause — found in `~/.hermes/SOUL.md` itself, read-only, no secrets

`hermes skills --help` (the CLI) only covers `browse/search/install/...` — it
does not touch the agent-internal `skill_manage` tool. That tool is
implemented in the vendored `hermes-agent` source on the VM,
`~/.hermes/hermes-agent/tools/skill_manager_tool.py` (`NousResearch/hermes-agent`
checkout). `_create_skill`/`_edit_skill` write the **live** file via
`atomic_write_text(skill_md, content)` and validate frontmatter/size first —
an empty `content` argument would fail `_validate_frontmatter` and
`skill_manage` would return an *error*, not the `completed (N chars)` success
we see in both incidents. So the live `~/.hermes/skills/.../SKILL.md` write
is atomic and was never the empty one. `grep -n '^\s*skill' ~/.hermes/config.yaml`
returns only the section header `105:skills:` — no git/mirror config there;
the mirroring is not a coded feature at all, it's an **ad hoc shell recipe
the agent re-runs from its own prompt every turn**, and that recipe is
SOUL.md rule 2, verbatim:

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

**The bug:** SOUL.md's own prose claims "the repo path `skills/<name>/SKILL.md`
mirrors my live `~/.hermes/skills/<name>/SKILL.md`" — a **flat** 1:1 mapping.
That's false for any categorized skill: live skills actually live at
`~/.hermes/skills/<category>/<name>/SKILL.md` (confirmed: `orchestration/linear-ticketing`,
`orchestration/engagement-checker`, `orchestration/daily-work-brief`,
`orchestration/secure-delta-collectors`, `email/himalaya`,
`hermes-operations/hermes-orchestration` — only ungrouped skills like
`delegate-to-cooper` and `helptech-duty-intake` sit flat). When `$N` is set
to the bare skill name for a categorized skill:

1. `cat ~/.hermes/skills/$N/SKILL.md` reads the **wrong, nonexistent** local
   path and fails, printing nothing to stdout — but the script has no
   `set -o pipefail` and never checks the `cat`'s exit code.
2. The pipe proceeds anyway: `ssh cooper "... cat > .../skills/$N/SKILL.md"`
   still runs, with **empty stdin**. Shell opens (`>`) and truncates the
   destination *before* `cat` reads anything; `cat` on empty stdin is not an
   error, it just writes zero bytes.
3. `git add && git commit` commits that genuinely empty file.

This exactly explains the two observed failure shapes:
- **himalaya**: the live skill is `email/himalaya`, no flattened repo path
  ever existed, so the empty write **creates a brand-new file** at the wrong
  flattened path `skills/himalaya/SKILL.md` — matches PR #42
  (`changeType: ADDED`, 0/0 lines) exactly, blob hash confirmed below.
- **linear-ticketing** (and engagement-checker, per PR #53's title "restore
  and preserve DM handoffs" — same pattern, not separately traced here): the
  live skill is `orchestration/linear-ticketing`, but a flattened
  `skills/linear-ticketing/SKILL.md` already existed in the repo from before
  category nesting was introduced — so the empty write **wipes the existing
  file in place** — matches PR #44/#49/#54 (`changeType: MODIFIED`, 0
  additions / N deletions == N, i.e. every line removed, nothing added)
  exactly.

Confirmed at the blob level — both wipe/empty-add commits point at git's
canonical empty-blob hash, not a partial write:
```
$ git ls-tree -r 241c9d5 skills/himalaya/SKILL.md          # PR #42
100644 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391 skills/himalaya/SKILL.md
$ git ls-tree -r f3acd6c skills/linear-ticketing/SKILL.md   # PR #44
100644 blob e69de29bb2d1d6434b8b29ae775ad8c2e48c5391 skills/linear-ticketing/SKILL.md
```
`e69de29b...` is literally `git hash-object` of a zero-byte file — this is
`cat` of empty stdin, not truncation of partial content, not a size/timeout
limit. Self-correction (Tars catching its own empty commit ~15–30s later
and pushing a restore) is presumably a post-hoc verify step baked into the
same turn's prompt flow, not visible in `agent.log` (no distinct log line
marks a verification read).

## 4. Full drift check — live VM skills vs `origin/main`

Live: 128 `SKILL.md` files under `~/.hermes/skills/` (recursive; the naive
`~/.hermes/skills/*/SKILL.md` glob used in the task brief misses every
categorized skill — 58 matched, 128 exist). Repo (`origin/main`, fetched)
tracks only **9** `SKILL.md` files under `skills/` — i.e. only the skills
Tars has actually self-edited via `skill_manage` ever get mirrored; the other
~119 live skills (bundled/hub skills, skills authored by Claude Code sessions
elsewhere, etc.) were never expected to land here and aren't drift.

Of those 9 tracked files, comparing live `md5sum` against
`git show origin/main:<path> | md5sum`:

| repo path | status |
|---|---|
| `skills/daily-work-brief/SKILL.md` | in sync |
| `skills/delegate-to-cooper/SKILL.md` | in sync |
| `skills/engagement-checker/SKILL.md` | in sync |
| `skills/helptech-duty-intake/SKILL.md` | in sync |
| `skills/hermes-operations/hermes-orchestration/SKILL.md` | in sync |
| `skills/linear-ticketing/SKILL.md` | in sync (post PR #55 restore) |
| `skills/email/himalaya/SKILL.md` | **stale** — see below |
| `skills/himalaya/SKILL.md` (orphan, wrong path) | matches live, but shouldn't exist |
| `skills/secure-delta-collectors/SKILL.md` | **drifted, unrelated to the empty-file bug** |

**himalaya duplicate/stale path.** Repo `skills/email/himalaya/SKILL.md` (md5
`a99ac16a...`) no longer matches live `~/.hermes/skills/email/himalaya/SKILL.md`
(md5 `e72577a7...`). The live file's current content instead matches the
*wrong* orphan path `skills/himalaya/SKILL.md` (md5 `e72577a7...`, byte-identical).
Diff (correct path vs live):
```diff
278c278
< Most commands support `--output` for structured output:
---
> Most commands support `--json` for structured output (Himalaya v2 uses `--json`, not `--output json`):
280a281
> himalaya account list --json
```
One-line fix, live since 21:04:17Z (see §5) — it only ever landed on the
wrong path (PR #43), never on the semantically-correct path. The repo now
permanently carries a duplicate: one correct-path file that's one edit stale,
one wrong-path file that's current. Neither has been cleaned up since PR #43
(2026-08-10 21:04:58Z).

**secure-delta-collectors — large unrelated drift, not backpropagated at
all.** Live is `v1.4.0`; repo is still `v1.1.0` (last touched 2026-08-08,
PR #26). `diff` is +66/-30/~17 — a new "HTTP transport hardening" section,
a rewritten "Testing and verification" section, and a second reference doc
(`references/slack-oauth-user-token.md`) that exists live but isn't in the
repo at all. This is not the empty-file bug — it's a normal live edit that
was simply never run through the rule-2 mirror flow. Flagging it because
step 5 asks for exactly this ("an edit that never got backpropagated").

## 5. Timestamps — live mtime vs last mirroring commit

```
$ ssh gaetan@192.168.0.9 stat -c "%y %n" <8 tracked live paths>
2026-08-10 20:38:42  orchestration/daily-work-brief/SKILL.md
2026-08-08 12:11:15  delegate-to-cooper/SKILL.md
2026-08-10 21:04:17  email/himalaya/SKILL.md
2026-08-11 15:31:33  orchestration/engagement-checker/SKILL.md
2026-08-11 15:28:29  helptech-duty-intake/SKILL.md
2026-08-08 10:52:08  hermes-operations/hermes-orchestration/SKILL.md
2026-08-11 16:13:07  orchestration/linear-ticketing/SKILL.md
2026-08-11 12:35:50  orchestration/secure-delta-collectors/SKILL.md

$ git log -1 --format='%ad %h %s' --date=iso origin/main -- <same paths>
2026-08-10 20:59:45  c50ec4f  daily-work-brief: ...                    (after live edit — in sync)
2026-08-08 12:12:04  dae0a16  delegate-to-cooper: ...                  (after live edit — in sync)
2026-08-10 19:58:46  3c8c8f3  himalaya: use the v2 JSON output flag    (BEFORE the 21:04:17 live edit — stale)
2026-08-11 15:32:24  6ec9f33  engagement-checker: restore and preserve DM handoffs  (after — in sync)
2026-08-11 15:28:48  2b28885  helptech-duty-intake: ...                (after — in sync)
2026-08-08 10:52:40  d2d9f4d  hermes-orchestration: ...                (after — in sync)
2026-08-11 16:13:44  64c027b  linear-ticketing: restore file and record delegated session provenance (after — in sync)
2026-08-08 15:14:25  3095df7  secure-delta-collectors: filter before candidate caps  (~4 days BEFORE the 12:35:50 live edit — badly stale)
```
The live himalaya mtime (21:04:17Z) lines up exactly with the `skill_manage
completed` log line for incident 1 (§2) and with when PR #42/#43 fired
(21:04:34–58Z) — confirms that edit is the one that only landed on the wrong
path. `secure-delta-collectors` is the clearest "edit that never got
backpropagated": last commit predates the live edit by ~4 days, and no PR for
it exists at all as of investigation time.

## Bottom line

- **No persistent local checkout exists on the VM** — Tars mirrors skills by
  shelling out to `cooper` (`ssh cooper "cd ~/dev/Tars && ..."`) per SOUL.md
  rule 2, using an ephemeral terminal-tool sandbox on the VM side only to run
  that ssh. Step 1's reflog ask doesn't apply; the ephemeral-env log lines are
  the evidence for why.
- **Root cause of the empty commits, found in `~/.hermes/SOUL.md` (not a code
  bug):** rule 2's one-liner `cat ~/.hermes/skills/$N/SKILL.md | ssh cooper
  "... cat > .../skills/$N/SKILL.md"` assumes a flat skill layout that most
  skills don't have (they live under a category subdirectory). No
  `set -o pipefail`, no exit-code check. When `$N` lacks the category prefix,
  the local `cat` fails silently, and the remote `cat > file` still truncates
  the destination to zero bytes and commits it — confirmed at the blob level
  (both incidents' commits point at git's canonical empty blob
  `e69de29b...`, not a partial write).
- **The bug is still live**, recurring for `linear-ticketing` three times
  since the ticket was filed (PRs #44/45, #49/50, #54/55, each pair ~20s
  apart) and at least once for `engagement-checker` (PR #53 title). Fix
  belongs in SOUL.md rule 2's script (resolve `$N` to its full
  `<category>/<name>` path, or add `set -o pipefail` plus an explicit
  "abort if local cat produced zero bytes" guard before the remote write) —
  out of scope for this read-only investigation, flagging for the ticket.
- **Drift beyond the named incidents:** himalaya has a permanent orphan
  duplicate at the wrong path holding the current content, while the
  "correct" path is one edit stale; `secure-delta-collectors` has a large,
  unrelated, ~4-day-old unbackpropagated edit (v1.1.0 in repo vs v1.4.0 live).
  Five of nine tracked skills are in clean sync as of investigation time.
