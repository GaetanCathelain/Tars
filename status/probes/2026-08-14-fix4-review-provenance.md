# Red-team review of FIX 4 — provenance methodology (2026-08-14)

Reviewer lens: attack the mtime-clustering inference behind
`status/probes/2026-08-14-skills-mirror-drift.md`. Read-only VM checks only
(`ls/stat/find/grep/md5sum/git/diff`), no writes, no `sops -d`, no config/.env reads.

**Headline: mtime clustering was the wrong instrument, and a much stronger one was
sitting in `~/.hermes/logs/`.** `skill_manage` tool completions are logged with
second precision and match SKILL.md mtimes exactly. Cross-matching all 132 live
SKILL.md mtimes against that log settles every open question in the inventory —
including the 3 "ask the owner" cases — and changes the membership of the sync list.

---

## 0. The decisive instrument the inventory missed

`~/.hermes/logs/agent.log{,.1,.2}` log every tool call:

```
2026-08-13 14:54:51,988 INFO [20260813_140745_980067c0] agent.tool_executor: tool skill_manage completed (0.02s, 158 chars)
```

Coverage is continuous from **2026-08-07 16:20:29** (agent.log.2 first line) to now —
i.e. the entire post-go-live window. 219 `skill_manage` completions total.
The args are not logged, but the **timestamp is**, and a `skill_manage` write stamps
the file at that same second. Cross-match of all 132 live SKILL.md mtimes against the
219 timestamps:

```
-- live SKILL.md whose mtime EXACTLY matches a skill_manage call (14):
2026-08-08 10:52:08 hermes-operations/hermes-orchestration/SKILL.md     [mirrored]
2026-08-08 12:11:15 delegate-to-cooper/SKILL.md                          [mirrored]
2026-08-08 12:47:48 hermes-operations/skill-library-auditing/SKILL.md    NOT mirrored
2026-08-10 09:51:44 email/durable-email-automation/SKILL.md              NOT mirrored
2026-08-10 12:37:22 research/upstream-change-research/SKILL.md           NOT mirrored
2026-08-11 12:35:50 orchestration/secure-delta-collectors/SKILL.md       [mirrored]
2026-08-11 15:28:29 helptech-duty-intake/SKILL.md                        [mirrored]
2026-08-11 16:26:39 communication/stakeholder-communication/SKILL.md     NOT mirrored
2026-08-11 18:19:56 email/himalaya/SKILL.md                              [mirrored]
2026-08-13 14:55:48 orchestration/engagement-checker/SKILL.md            [mirrored]
2026-08-14 10:53:23 software-development/local-ipc-probes/SKILL.md       NOT mirrored
2026-08-14 11:27:00 hermes-operations/hermes-skill-portability/SKILL.md  NOT mirrored
2026-08-14 11:27:49 creative/mobile-club-okr-report/SKILL.md             [mirrored]
2026-08-14 11:32:55 hermes-operations/hermes-gateway-operations/SKILL.md NOT mirrored
```

`productivity/google-workspace` (2026-08-11 19:46:43) is **absent** — no `skill_manage`
call at that second (nearest: 18:19:56 and 2026-08-12 08:31:15).

Caveat, stated up front: the converse is **not** proof. 4 of the 11 already-mirrored
(and therefore known-Tars) skills also lack a log match — `email/gmail-triage`
(08-12 15:55:56), `orchestration/slack-channel-context` (08-13 12:36:06),
`orchestration/daily-work-brief` (08-13 13:16:09), `orchestration/linear-ticketing`
(08-13 15:18:19). Each of those carries a `SKILL.md.bak-*` sibling written seconds
before (e.g. `email/gmail-triage/SKILL.md.bak` @ 15:55:51), i.e. the cooper-session
`cp .bak` + edit-over-ssh convention. So absence of a log match means
"not written by `skill_manage`", not "not Tars-related". For google-workspace the
absence combines with three other signals (below) and is decisive.

---

## 1. Does `/upgrade` touch `~/.hermes/skills`? YES — and that is why there is no 08-13 cluster

- **It does write there.** `~/.hermes/hermes-agent/hermes_cli/update_cmd.py:963-968`,
  `:4357-4363`, `:4388-4402` call `tools.skills_sync.sync_skills()` —
  *"Sync bundled skills (copies new, updates changed…)"*. `tools/skills_sync.py:127`
  does `shutil.copytree(skill_src, dest)` into the user dir; manifest at
  `~/.hermes/skills/.bundled_manifest` (70 entries, **rewritten today 2026-08-14
  13:46:49**), so the sync runs routinely, not only on upgrade.
- **`shutil.copytree` uses `copy2` → it preserves the source mtime.** A bundled skill
  copied into `~/.hermes/skills` at any date carries the *bundle's* mtime, currently
  `2026-08-07 16:17:59` (`ls -ld ~/.hermes/hermes-agent/skills` → `Aug 7 16:17`,
  untouched by the 08-13 upgrade). This is the mechanical reason no 08-13 cluster
  exists even though the sync ran: same content, same hashes, nothing rewritten, and
  anything that *had* been rewritten would still have shown 16:17:59.
- **Empirically the 08-13 upgrade left the skills tree alone.** 15 `hermes-agent/*`
  subtrees are stamped 2026-08-13T14:46–14:48; no SKILL.md under `~/.hermes/skills`
  has an mtime in that window (`find ~/.hermes/skills -newermt "2026-08-13 00:00"
  -name SKILL.md` returns only the 8 known agent/session edits).
- **`sync_skills` will not clobber a user-edited bundled skill**: `skills_sync.py:15-17`
  — bundled changed + user copy == origin hash ⇒ update; user copy differs ⇒ skip.
  So the google-workspace venv patch is protected while it stays diverged.

**Which cluster conclusions survive:**

| Inventory claim | Verdict |
|---|---|
| Cluster A (n=68 @ 08-07 16:17:59) = shipped with Hermes | **SURVIVES, on better evidence.** `~/.hermes/hermes-agent/skills` holds exactly **70** bundled SKILL.md; all 70 exist live (`comm -12` → 70, pkg-only 0). The 2 that are not in cluster A are the 2 later-modified ones: `email/himalaya` and `productivity/google-workspace`. Bundled-ness is now provable by set membership, not by mtime. |
| Cluster B (n=35 @ 08-08 12:00:37) = "a second bulk install/upgrade batch (matches the vanilla-skills reset referenced in memory)" | **FALSE.** 0 of the 35 are in the bundled tree. They are the Matt Pocock skill pack (`ask-matt`, `tdd`, `grilling`, `writing-for-agents`, and literally `setup-matt-pocock-skills`). Nothing to do with a Hermes upgrade or a vanilla reset. |
| Cluster C (n=6 ponytail, 2026-06-20) = "predates go-live entirely → pre-Tars install" | **FALSE as reasoning.** `~/.hermes-skill-import/` shows an agent-run import at **2026-08-07 21:48–21:54** whose `stage-home/skills/` holds exactly the 6 ponytail dirs + `email/rich-email-composition`, with the *same* 2026-06-20T18:44:59..18:46:50 mtimes, and staged SKILL.md md5s **identical** to live (`0a7901d0…` ponytail, `131fe8fd…` rich-email). Those files were written into `~/.hermes/skills` on 08-07 ~21:51, mtime-preserved. Same defect kills the "`email/rich-email-composition` 2026-08-02 → pre-Tars — not a candidate" row. |

**mtime is not write time on this box, and it has been demonstrated twice
(copytree-based bundle sync, and the 08-07 import).** The strong-7 list was built on
the assumption that it is.

---

## 2. Per-suspect provenance grades

`SM` = an exact-second `skill_manage` completion in `~/.hermes/logs/agent.log*`.
"refs" = non-SKILL.md files in the dir (from the probe's own
`scratchpad/vm-nonskillmd-counts.txt`).

| # | skill dir | SKILL.md mtime | SM match | bundled? | other evidence | grade |
|---|---|---|---|---|---|---|
| 1 | `email/durable-email-automation` | 08-10 09:51:44 | **YES** (+ refs 08-10 08:49:46, 09:51:23 both SM) | no | 2 refs | **STRONG — confirmed Tars/skill_manage** |
| 2 | `research/upstream-change-research` | 08-10 12:37:22 | **YES** (+ ref 12:38:05 SM) | no | 1 ref | **STRONG — confirmed** |
| 3 | `communication/stakeholder-communication` | 08-11 16:26:39 | **YES** (+ ref 08-08 10:29:59 SM) | no | 1 ref | **STRONG — confirmed** |
| 4 | `productivity/google-workspace` | 08-11 19:46:43 | **NO** | **YES — one of the 70 bundled** | all scripts/refs at bundle mtime 08-07 16:17:59; live vs bundled SKILL.md differs by 52 lines incl. `GSETUP=…/hermes-agent/venv/bin/python …` and `GAPI=…` — the documented venv-python patch; `SKILL.md.bak-gcn31-20260811` (ticket-named, cooper convention) whose mtime 08-07 21:26:26 is also not an SM timestamp | **REFUTED as Tars-authored.** Bundled skill + ssh-session patch. |
| 5 | `software-development/local-ipc-probes` | 08-14 10:53:23 | **YES** (+ ref 10:53:15 SM) | no | 1 ref | **STRONG — confirmed** |
| 6 | `hermes-operations/hermes-skill-portability` | 08-14 11:27:00 | **YES** (+ refs 11:26:50, 08-07 21:13:51, 21:56:07, 08-08 12:30:42 all SM) | no | 4 refs; dir created 08-07 21:13, not 08-14 | **STRONG — confirmed** |
| 7 | `hermes-operations/hermes-gateway-operations` | 08-14 11:32:55 | **YES** (+ refs 11:32:19, 08-10 12:21:08 SM; one ref @ 08-13 14:47:39 has no SM match) | no | 3 refs | **STRONG — confirmed** |
| 8 | `hermes-operations/skill-library-auditing` (was MEDIUM) | 08-08 12:47:48 | **YES** (+ ref 12:47:32 SM) | no | 2 refs | **STRONG — promoted from MEDIUM, settled** |
| 9 | `i-have-adhd` (was ambiguous) | 08-07 16:32:30 | **NO** — first SM call ever is 08-07 21:13:23, log starts 16:20:29 | no | `~/.hermes/config.yaml.bak-b5-adhd` @ 16:32:20, **10 s earlier** → provisioning/config event | **REFUTED — settled, no owner input needed** |
| 10 | `last30days` (was ambiguous) | 08-08 12:05:11 | **NO** (nearest 12:05:23) | no | all 122 files incl. four multi-MB assets share mtime `…12:05:11.55` → single bulk extract; `skill_manage` writes one file at a time | **REFUTED — settled, no owner input needed** |

Style/marker check: no `SKILL.md.bak-tars-selfedit`-type sibling exists for any of the
7 suspects (`find ~/.hermes/skills -name "SKILL.md.*"` → 18 baks, all in
already-mirrored dirs + google-workspace). Baks are therefore evidence of the
**ssh-edit** convention, not of `skill_manage`; the inventory's line 34 reading of them
as "Tars' own pre-edit backups per the hard-rule workflow" is inverted for the 4
mirrored dirs listed in §0's caveat, and is what makes google-workspace's bak
informative.

---

## 3. Arithmetic sanity vs the raw TSVs

| Check | Result |
|---|---|
| `wc -l vm-skillmd.tsv` = 132; distinct dirs = 132 | ✅ matches "132 live SKILL.md dirs" |
| Cluster sizes from TSV: `68 @ 2026-08-07T16:17:59`, `35 @ 2026-08-08T12:00:37` | ✅ |
| Cluster C = 6 (2026-06-20 window, 6 distinct seconds) | ✅ (report states a window, not an identical second) |
| 68+35+6 = 109; 121−109 = 12 individually listed rows | ✅ 12 rows present |
| 132 − 11 mirrored = 121 live-only | ✅ |
| Mirrored dirs in `repo-skillfiles.tsv` = 11 | ✅ |
| "**16 files** checked total across the 11 dirs" | ❌ `repo-skillfiles.tsv` has **17** rows and `find skills -type f` = **17**. Off by one. |
| md5 identity of the 11 | ✅ spot-verified 6/17 files against the VM (himalaya SKILL.md + references/configuration.md, secure-delta-collectors SKILL.md, gmail-triage SKILL.md, mobile-club templates/report-okr-template.html, daily-work-brief scripts/linear_board.py) — all match |
| "non-SKILL.md files" column | ❌ **wrong on 9 of 12 rows**, contradicting the probe's own `vm-nonskillmd-counts.txt`: durable-email-automation 2 (reported 0), upstream-change-research 1 (0), stakeholder-communication 1 (0), local-ipc-probes 1 (0), hermes-skill-portability 4 (0), hermes-gateway-operations 3 (0), skill-library-auditing 2 (0), google-workspace 9 (1), last30days 122 (1) |
| Symlink blind spot | none — `find -L ~/.hermes/skills -name SKILL.md` = 132, same as `find`; 0 symlinks in the tree |

---

## 4. Which recommendation lines survive

| Recommendation line (drift report §Recommendation) | Verdict |
|---|---|
| "the 11 already-mirrored need nothing (all MIRRORED-IDENTICAL)" | **SURVIVES** (spot-verified). |
| "land the 7 strong-confidence live-only suspects … by PR" | **SURVIVES with an amended list.** Drop `productivity/google-workspace`; add `hermes-operations/skill-library-auditing`. Still 7 dirs, and they are provably *exactly* the set of `skill_manage`-written SKILL.md not yet mirrored (§0). |
| "…pulling VM→repo is the only direction that makes sense (no repo-side edits exist to conflict with)" | **SURVIVES** — 0 REPO-ONLY, 0 diverged, verified. |
| "treat `i-have-adhd`, `last30days`, `skill-library-auditing` as needs-owner-input" | **DOES NOT SURVIVE.** All three are settled by log/bundle/bulk-mtime evidence (§2 rows 8–10). Asking the owner spends his attention on a question the VM already answers. |
| "leave the 68+35+6 clustered dirs out of this sync" | **Outcome survives; both stated reasons for B and C are false** (§1). Cluster B is a third-party skill pack, not an upgrade; Cluster C was written post-go-live by an mtime-preserving import. |
| Implicit scope "sync the 7 **SKILL.md**" | **DOES NOT SURVIVE.** The 7 dirs carry **12 Tars-authored reference files** whose mtimes also match `skill_manage` calls. SKILL.md-only PRs land 6 skills whose bodies point at `references/*.md` that do not exist in the mirror. |
| Note "`productivity/google-workspace` … confirm provenance before PR-ing it" | **Hedge was right, the table row was wrong.** Provenance is now confirmed: bundled + ssh patch. It belongs in the upgrade-patch ledger (`~/.claude/projects/-home-gaetan-dev-Tars/memory/hermes-upgrade-reapply-local-patches.md`), not in the SOUL-rule-2 self-edit mirror. |

### Amended fix-4 list (7 dirs, 7 SKILL.md + 12 reference files)

```
communication/stakeholder-communication/      SKILL.md + 1 ref
email/durable-email-automation/               SKILL.md + 2 refs
hermes-operations/hermes-gateway-operations/  SKILL.md + 3 refs
hermes-operations/hermes-skill-portability/   SKILL.md + 4 refs
hermes-operations/skill-library-auditing/     SKILL.md + 2 refs
research/upstream-change-research/            SKILL.md + 1 ref
software-development/local-ipc-probes/        SKILL.md + 1 ref
```

Out: `productivity/google-workspace` (bundled + local patch — different problem,
different ledger).

### Method note for the next drift check

Stop diffing mtimes. The check is two lines:

```sh
# Tars-written (skill_manage) live skills
cat ~/.hermes/logs/agent.log* | grep 'tool skill_manage' | cut -c1-19 | sort -u > /tmp/sm.txt
find ~/.hermes/skills -name SKILL.md -printf '%TY-%Tm-%Td %TH:%TM:%TS|%p\n' \
  | while IFS='|' read t p; do grep -qxF "${t:0:19}" /tmp/sm.txt && echo "$p"; done
# bundled (never a mirror candidate)
comm -12 <(cd ~/.hermes/hermes-agent/skills && find . -name SKILL.md | sort) \
         <(cd ~/.hermes/skills           && find . -name SKILL.md | sort)
```

Caveat that must ride along: log rotation. `agent.log.2` starts 2026-08-07 16:20:29 —
today that covers all of Tars' life, but the three files hold ~11.9 MB and rotate.
Once a rotation drops pre-cutoff history this method degrades silently. Snapshot the
`skill_manage` timestamp list (219 lines) somewhere durable now.
