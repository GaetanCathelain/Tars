# FIX 4 reconciliation — settling provenance vs policy by direct evidence (2026-08-14)

Both red-team reviews returned FIX-PLAN and disagreed on the list:

- `status/probes/2026-08-14-fix4-review-provenance.md` → **7 dirs**, i-have-adhd and
  last30days settled as install-time.
- `status/probes/2026-08-14-fix4-review-policy.md` → **9 dirs** ("1 false positive,
  5 false negatives"), i-have-adhd / last30days = owner's call.

This file does not average them. Every dir in the union of both lists plus the three
disputed ones was re-tested from scratch against the same four instruments.

Read-only throughout. VM commands: `ssh gaetan@192.168.0.9` + `find/stat/grep/md5sum/
head/ls/cut/wc` only, plus `ssh phermes` (read-only) for origin tracing. No `sops -d`,
no `~/.hermes/config.yaml` or `.env` read (the one config-adjacent check is an `ls -la`
on a `.bak` **filename**, no content), no credential printed. Nothing written on either
VM; nothing written in the repo except this file.

VM clock at probe time: `Fri Aug 14 02:02:05 PM UTC 2026`.

---

## 0. Decision rule (stated once, applied uniformly)

> A dir belongs in the mirror sync **iff**
> `(NOT in ~/.hermes/skills/.bundled_manifest)`
> `AND (NOT third-party-pack bulk)`
> `AND (skill_manage completion log match OR unambiguous Tars self-edit evidence)`.

Operational definitions used:

| Term | Test |
|---|---|
| bundled | `grep -qx "<basename>:.*" ~/.hermes/skills/.bundled_manifest` (70 entries) |
| third-party-pack bulk | SKILL.md mtime in a bulk-write cluster: Matt Pocock pack `2026-08-08T12:00:37` (35 dirs), ponytail pack `2026-06-20` (6 dirs), or any dir whose whole file set shares one mtime second |
| skill_manage log match | SKILL.md mtime equals (±2 s) the timestamp of a `tool skill_manage **completed**` line in `~/.hermes/logs/agent.log{,.1,.2}` |
| Tars self-edit evidence | `SKILL.md.bak-*` sibling written seconds before, mode 600, cooper ssh-edit convention |

### 0.1 Instrument correction the provenance review got wrong

The provenance review matched mtimes against **219** `skill_manage` lines. Only **140**
are completions; the other **79** are `WARNING … Tool skill_manage returned error`, which
write **no file**. Matching against error lines is a false-positive generator.

```
$ grep -c "tool skill_manage completed" agent.log{,.1,.2}   → 140
$ grep -c "skill_manage"                 agent.log{,.1,.2}   → 219   (140 completed + 79 errors)
```

Re-running the cross-match against **completions only**, over all 132 live SKILL.md:

```
SKILL.md matching a skill_manage COMPLETION                  → 14
SKILL.md matching ONLY an error line (would-be false positive) → 0
```

**The provenance review's 14 survives** — but by luck, not by method. Its `219`
denominator is corrected here; future runs must filter on `completed`.

### 0.2 Log coverage — no gaps, so absence is informative in-window

```
agent.log.2  2026-08-07 16:20:29 → 2026-08-10 11:39:13
agent.log.1  2026-08-10 11:39:13 → 2026-08-13 14:54:12
agent.log    2026-08-13 14:54:39 → 2026-08-14 14:02:33
```

Continuous from 2026-08-07 16:20:29. The **first `skill_manage` line of any kind is
2026-08-07 21:13:23**; first completion is the same second. So any file stamped between
16:20:29 and 21:13:23 provably was not written by `skill_manage`. Files stamped **before**
16:20:29 are outside log coverage — for those the log is silent, not exculpatory, and a
second instrument is required (used below for meme-generator / rich-email-composition).

---

## 1. Per-dir evidence table

12 dirs = provenance-7 ∪ policy-9 ∪ {i-have-adhd, last30days, productivity/google-workspace}.

| dir | in `.bundled_manifest` | pack bulk | SKILL.md mtime | skill_manage completion match | SKILL.md mode | backup / pycache siblings | files w/ completion match |
|---|---|---|---|---|---|---|---|
| `email/durable-email-automation` | **n** | n | 2026-08-10T09:51:44 | **YES ±0 s** | 600 | none | 3 / 3 |
| `research/upstream-change-research` | **n** | n | 2026-08-10T12:37:22 | **YES ±0 s** | 600 | none | 2 / 2 |
| `communication/stakeholder-communication` | **n** | n | 2026-08-11T16:26:39 | **YES ±0 s** | 600 | none | 2 / 2 |
| `software-development/local-ipc-probes` | **n** | n | 2026-08-14T10:53:23 | **YES ±0 s** | 600 | none | 2 / 2 |
| `hermes-operations/hermes-skill-portability` | **n** | n | 2026-08-14T11:27:00 | **YES ±0 s** | 600 | none | 5 / 5 |
| `hermes-operations/hermes-gateway-operations` | **n** | n | 2026-08-14T11:32:55 | **YES ±0 s** | 600 | none | 3 / 4 (see §4) |
| `hermes-operations/skill-library-auditing` | **n** | n | 2026-08-08T12:47:48 | **YES ±0 s** | 600 | none | 3 / 3 |
| `creative/meme-generator` | n | n | 2026-07-31T16:31:01 | **none** (pre-log; nearest completion +621 742 s) | 600 | none | 0 / 2 |
| `email/rich-email-composition` | n | n | 2026-08-02T15:24:18 | **none** (pre-log; nearest +452 945 s) | 600 | none | 0 / 2 |
| `i-have-adhd` | n | n | 2026-08-07T16:32:30 | **none** — in-window, and no `skill_manage` line existed yet (first ever 21:13:23) | 664 | none | 0 / 1 |
| `last30days` | n | **YES** (123 files, one mtime second) | 2026-08-08T12:05:11 | **none** (nearest +12 s) | 664 | none | 0 / 123 |
| `productivity/google-workspace` | **YES** | n | 2026-08-11T19:46:43 | **none** (nearest −5 207 s) | 664 | `SKILL.md.bak-gcn31-20260811`, `scripts/__pycache__/` (3 `.pyc`) | 0 / 10 |

Quoted log line, representative shape (only tool name / timestamp / duration / length are
logged — `skill_manage` args are **not** logged, so nothing is redacted beyond what the
line already omits):

```
2026-08-14 11:32:55,xxx INFO [<session-id>] agent.tool_executor: tool skill_manage completed (…s, … chars)
```

The seven ACCEPTED dirs each have a completion line at exactly their SKILL.md second
(delta +0 s, not ±2 s — the tolerance was never needed).

Cluster arithmetic re-derived independently: 132 live SKILL.md, 70 bundled, 35 Matt Pocock
(`2026-08-08T12:00:37`), 6 ponytail (`2026-06-20`).

---

## 2. The policy review's "5 false negatives" — tested one by one

| claimed false negative | rule outcome | evidence |
|---|---|---|
| `hermes-operations/skill-library-auditing` | **ACCEPTED** | not bundled, not pack, SKILL.md completion match `2026-08-08 12:47:48` ±0 s, mode 600, both references also completion-matched (`21:20:40`, `12:47:32`). Both reviews in fact agree here — the provenance review promoted it from MEDIUM. Not a real disagreement. |
| `creative/meme-generator` | **REJECTED** | Third instrument settles it: the identical file lives on **p-Hermes** — `md5 e322950838806f150394f2986287c865` on both `192.168.0.9:~/.hermes/skills/creative/meme-generator/SKILL.md` and `phermes:~/.hermes/skills/creative/meme-generator/SKILL.md`, same 2026-07-31 16:31 stamp. It is also listed in `~/.hermes-skill-import/destination_before.json`, i.e. already on the Tars VM before the 08-07 21:49 import ran. Frontmatter says `author: Hermes Agent` — but that Hermes is p-Hermes: the Tars VM did not exist on 2026-07-31 (bundled skills stamped 2026-08-07T16:17:59, logs open 16:20:29). Mtime-preserved copy of a p-Hermes skill, not a Tars self-edit. |
| `email/rich-email-composition` | **REJECTED** | Staged in the 2026-08-07 21:48–21:54 import: `~/.hermes-skill-import/stage-home/skills/email/rich-email-composition/` holds both files with **identical md5s** to live (`131fe8fd…` SKILL.md, `81de6eb3…` reference). Imported, mtime-preserved. The policy review's rebuttal ("go-live is the *Slack* date; the VM predates it") is wrong on fact — the VM's Hermes install is 2026-08-07 16:17:59, *later* than this file's 2026-08-02 mtime, which is exactly why it had to arrive by copy. |
| `i-have-adhd` | **REJECTED — settled, not owner's call** | see §3 |
| `last30days` | **REJECTED — settled, not owner's call** | see §3 |

And the claimed **1 false positive**, `productivity/google-workspace`: **CONFIRMED
REJECTED**, on which both reviews already agree. `grep -c '^google-workspace:'
.bundled_manifest → 1`; every script/reference still at bundle mtime `2026-08-07T16:17:59`
while only SKILL.md moved 08-11; mode 664; `SKILL.md.bak-gcn31-20260811` present. Bundled
skill + local venv-python patch → belongs to
`~/.claude/projects/-home-gaetan-dev-Tars/memory/hermes-upgrade-reapply-local-patches.md`,
never to the SOUL-rule-2 self-edit mirror.

**Net: of the policy review's 5 claimed false negatives, 1 is real (and was not actually
disputed), 4 are wrong.** Its manifest-plus-mode heuristic cannot distinguish "written
here by Tars" from "copied here, mtime preserved" — the same defect it correctly diagnosed
in the original mtime-clustering method, one level up.

---

## 3. i-have-adhd and last30days — SETTLED, no owner call needed

The provenance review reached the right verdict; the reasoning is upgraded here so it
rests on identity, not on adjacency.

**`i-have-adhd` — REJECTED, decisive.** Three independent facts:

1. `md5 829e24e6048347b60dfdf9ed4208da86` (6848 B) is **identical** on the Tars VM and on
   p-Hermes (`phermes:~/.hermes/skills/i-have-adhd/SKILL.md`, dir stamped Aug 2 09:26).
   The file was authored on p-Hermes and copied in.
2. Its Tars mtime `2026-08-07 16:32:30` sits **inside** log coverage (opens 16:20:29) and
   **before the first `skill_manage` call ever made on this box** (21:13:23). Not
   `skill_manage`, provably.
3. `ls -la ~/.hermes/config.yaml.bak-b5-adhd` (filename + stat only, content not read) →
   `-rw------- 5476  2026-08-07 16:32:20.506` — **10 s before** the skill file. A
   provisioning step that touched config and dropped the skill in the same breath.

Frontmatter carries `license: MIT`, `disable-model-invocation: true` and no author — a
third-party output-style skill riding along with the p-Hermes profile. Not Tars' work,
not the mirror's business.

**`last30days` — REJECTED, decisive.** Frontmatter is unambiguous:

```
name: last30days
version: "3.18.4"
homepage:   https://github.com/mvanhorn/last30days-skill
repository: https://github.com/mvanhorn/last30days-skill
author: mvanhorn
license: MIT
```

A **published third-party skill package**, version-tagged, by a named external author.
Corroborating: all **123** files share mtime `2026-08-08T12:05:11` (mode 664/775, ~14.4 MB
including five multi-MB assets — `dog-original.jpeg` 4.0 MB, `aging-portrait.jpeg` 2.8 MB,
`swimmom-mockup.jpeg` 2.8 MB, `dog-as-human.png` 2.5 MB, `claude-code-rap.mp3` 2.4 MB) and
a vendored dependency tree with its own LICENSE (`scripts/lib/vendor/bird-search/`).
`skill_manage` writes one file per call; 123 files in one second with a vendored package is
a tarball extraction. Nearest completion is +12 s away, matching nothing. Absent from
p-Hermes, so it was installed directly on the Tars VM from upstream.

**Neither needs Gaetan's attention.** The policy review's "genuinely owner's-call" rests on
the mode-664 heuristic alone; identity evidence (md5 against p-Hermes, upstream repo
metadata) answers the question the heuristic could only shrug at. Its recommendation to
exclude `last30days` for size was right, but for the weaker reason — it is excluded because
it is somebody else's published package, not because it is big.

**No OWNER-CALL items remain in Fix 4.**

---

## 4. Where the two reviews genuinely conflict, and how it was ruled

The one place the evidence genuinely pulls in two directions is
**`creative/meme-generator` and `email/rich-email-composition`** — the policy review's only
substantive additions. Both pass the policy rule cleanly: non-bundled, mode 600 (the exact
signature the policy review identified as "confirmed Tars self-edit"), and outside both
named packs — so under the rule *as that review wrote it* they must be mirrored, and its
argument that the inventory dismissed them on a bad premise ("pre-go-live" conflates the
Slack date with the VM date) is a fair hit on the original inventory. The provenance review
never examined them at all, so it offers no counter-argument, only silence: their mtimes
predate `agent.log.2`, and its log instrument returns "no match" for files it structurally
cannot see. Two reviews, one saying include on a passing rule, one mute — averaging would
have landed them. The tiebreak came from an instrument neither review applied to these two
dirs: **content identity against the machine they came from**. `rich-email-composition` is
byte-identical to its copy staged in `~/.hermes-skill-import/` during the 2026-08-07
21:48–21:54 profile import, and `meme-generator` is byte-identical to the live p-Hermes copy
and is named in that import run's `destination_before.json`. Mode 600 turns out to mark
"restrictive umask", not "Tars wrote it" — the p-Hermes profile carries 600 across the wire,
so the policy review's central heuristic is confounded by exactly the migration this repo
exists to document. Ruled **REJECTED**: mirroring them would import p-Hermes' skill library
into a repo whose stated contract (`CLAUDE.md:43-45`) is a mirror of what *Tars* maintains,
and would recreate the divergence trap the same review correctly identified for
google-workspace — a file the repo would then own but neither Tars nor `skill_manage` ever
writes, so "which side moved" becomes unanswerable at the next drift check.

Secondary, smaller conflict: **file counts**. The provenance review's headline says the 7
dirs carry "12 Tars-authored reference files"; its own per-dir table sums to 14. The policy
review's §4 says 14. **14 is correct** (2+1+1+1+4+3+2), verified by `find -type f`. The
provenance headline is an arithmetic slip, not a scope disagreement.

One residual anomaly, disclosed rather than smoothed over:
`hermes-operations/hermes-gateway-operations/references/slack-progress-scoping.md`
(576 B, mtime `2026-08-13 14:47:39`, mode 600) has **no** `skill_manage` completion at its
second. The log has no gap there — `agent.log.1` covers it, and the minute shows a live Tars
turn (`session=20260813_140745_980067c0`) streaming between tool calls at 14:47:34 and
14:47:51. It is not an `/upgrade` artifact: `sync_skills` only touches dirs in
`.bundled_manifest`, and this dir is not in it. Most likely a `terminal`-tool or ssh-side
write in the same working session. **Included anyway** — the decision rule qualifies *dirs*,
not individual files, and this dir qualifies on its SKILL.md plus two other
completion-matched references; the copy scope is then the whole-dir allowlist, per the
policy review's correct §4 point that scoping by per-file provenance is what produced the
partial-coverage lie in the first place.

---

## 5. FINAL sync list — 7 dirs, 21 files

Identical in membership to the provenance review's amended list, now independently
re-derived from completions-only matching, and with the file scope corrected to 14
references (not 12).

Exclusion filters applied to every dir: keep only `SKILL.md`, `references/**`, `scripts/**`,
`templates/**`; drop anything matching `*.bak*`, `*wf6-fabricating*`, `__pycache__/`, and
anything outside the allowlist. **In these 7 dirs the filters remove nothing — a
`find … \( -name '*.bak*' -o -name '*wf6-fabricating*' -o -name '__pycache__' \)` over all
12 candidate dirs returns hits only under `productivity/google-workspace` (already
excluded).** The filters still ship, because they are what keeps
`orchestration/daily-work-brief/SKILL.md.wf6-fabricating-board-20260810T201445Z` (a backup
that does not match `*.bak*`) out of any future sync.

Sizes and copy-time md5s are recorded so a pre-commit re-`md5sum` can abort on drift
(policy review §7 — three of these dirs were written today).

```
communication/stakeholder-communication/                                         2 files
  02efed76fea84d67ace94d9e4bc66167   5029 B  SKILL.md
  6326fabb2588c10c5a08b9b9d8699c23   1931 B  references/conversation-vs-project-overview.md

email/durable-email-automation/                                                  3 files
  6c1e799bbe075f73bc545524ac6bbe01   6748 B  SKILL.md
  3d85d40cf0cd3d0be82efda4a89bf1dd   4145 B  references/gmail-forwarding-idempotency.md
  1ec9df2fb7365427d612b90c947bc8aa   2965 B  references/historical-backfill.md

hermes-operations/hermes-gateway-operations/                                     4 files
  5ba519d4ea738e92e90fcc380cfdf0f8   9648 B  SKILL.md
  69c6d453a04878e06335db1ba6803d58   2502 B  references/slack-busy-ack-diagnostics.md
  1cb4580f67584ba26284c26c83853b45   2861 B  references/slack-mcp-write-policy.md
  4edb5f47627fc621c0b4f024e12c7a2f    576 B  references/slack-progress-scoping.md

hermes-operations/hermes-skill-portability/                                      5 files
  398b0e72380d261d21b730df61991257   5440 B  SKILL.md
  cde88dca19f35822e69ab5b3057bfccf   3401 B  references/deleted-package-forensics.md
  588538560ecacdde823c367dadd474eb   3097 B  references/deterministic-package-verification.md
  2cc75977d7c5d7b72f218701b9627ef1   3424 B  references/migration-checklist.md
  724a4a40fb23896e70711efd83ed59ae   1898 B  references/shared-cross-agent-install.md

hermes-operations/skill-library-auditing/                                        3 files
  a4d927e91ee4011ad08958d6e30fa933   4232 B  SKILL.md
  bcc47f91386c04519b493c7f283792d3   3979 B  references/checksum-diff-procedure.md
  84cdf6a1abcac10198c919ab214b71f1   2337 B  references/exact-collision-recovery.md

research/upstream-change-research/                                               2 files
  4235623ebf922843d14a0d3904524f37   5286 B  SKILL.md
  17f03949d7bb904efd39917fe1040b07   5649 B  references/github-evidence-workflow.md

software-development/local-ipc-probes/                                           2 files
  0869cd98b477ee5f633e8db865038ea9   6713 B  SKILL.md
  1135be2ba0c8007cf0d423e84c4ccb51   7966 B  references/claude-code-cross-session-v2.1.232.md
```

Total: **7 dirs, 21 files, ~90 KB.** No binaries, no backups, no `__pycache__`.

**Out, with reason:**

| dir | reason |
|---|---|
| `productivity/google-workspace` | bundled (`.bundled_manifest`); local venv patch → reapply-note ledger |
| `creative/meme-generator` | p-Hermes-origin, md5-identical to `phermes:` copy |
| `email/rich-email-composition` | import-staged 2026-08-07, md5-identical to `~/.hermes-skill-import/` copy |
| `i-have-adhd` | p-Hermes-origin, md5-identical; installed 10 s after a config provisioning step |
| `last30days` | third-party published package (`author: mvanhorn`, v3.18.4, MIT), 123-file bulk extract, ~14.4 MB |
| 35 Matt Pocock dirs | third-party pack, bulk `2026-08-08T12:00:37` |
| 6 ponytail dirs | third-party pack, import-staged, `2026-06-20` mtimes preserved |
| 70 bundled dirs | in `.bundled_manifest`, rewritten by `/upgrade` |

Direction check: no repo-side files exist under `skills/communication`, `skills/research`,
`skills/software-development`, and the 7 target dirs are absent from `skills/email` (holds
`gmail-triage`, `himalaya`) and `skills/hermes-operations` (holds `hermes-orchestration`
only). **0 REPO-ONLY, 0 diverged** — VM→repo is the only direction, confirmed.

---

## 6. Secrets gate on the final list — PASS

The policy review's §1 scan covered the *original* 7, which included
`productivity/google-workspace` but **not** `hermes-operations/skill-library-auditing`. Its
3 files were therefore never scanned. Re-run over all 7 final dirs:

```
grep -rlE "xox[abprse]-…|sk-…|ghp_…|github_pat_|AKIA…|AGE-SECRET-KEY|BEGIN … PRIVATE KEY|
           192\.168\.0\.[0-9]|100\.[0-9]+\.[0-9]+\.[0-9]+|\.ts\.net"  → 0 files
```

Zero hits, including for internal IPs and tailnet hostnames. Nothing to strip. Re-run this
immediately before commit, since three of the seven dirs are actively written.

---

## 7. Corrections this file makes to the two reviews

| claim | source | correction |
|---|---|---|
| "219 `skill_manage` completions" | provenance §0 | 140 completions; the other 79 are error returns that write nothing. Verdicts unaffected (0 error-only matches) but the method must filter on `completed`. |
| "7 dirs carry 12 reference files" | provenance §4 headline | 14. Its own per-dir table already sums to 14; the policy review's §4 is right. |
| "1 false positive, 5 false negatives" | policy §3 | 1 false positive confirmed; of 5 false negatives, 1 real and already agreed (`skill-library-auditing`), 4 wrong. |
| mode 600 ⇒ Tars self-edit | policy §2/§3 | Confounded: the p-Hermes profile carries 600 across the wire. `meme-generator`, `rich-email-composition` are 600 and not Tars'. Mode is corroborating at best, never decisive. |
| "`i-have-adhd`/`last30days` are owner's call" | policy §3/§6 | Settled by identity evidence. No owner input required. |
| mirror rule = "not bundled, minus named packs" | policy §5 | Insufficient as written — it admits any mtime-preserved import. The rule that holds is §0 above: non-bundled **and** non-pack **and** positively attributed to a write on this box. |

Standing caveat, carried forward from the provenance review and confirmed: the log
instrument depends on rotation. `agent.log.2` opens 2026-08-07 16:20:29 — today that covers
all of Tars' life, but the three files hold ~11.9 MB and rotate. The 140 completion
timestamps should be snapshotted somewhere durable before the window closes, otherwise this
method degrades silently and the next audit falls back to mtimes.
