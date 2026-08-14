# Red-team review of Fix 4 (skills mirror sync) — POLICY & SAFETY lens (2026-08-14)

Reviewer: red-team subagent. Read-only. Nothing written to the VM, nothing written
in the repo except this file. All VM commands were `ssh gaetan@192.168.0.9` +
`find/stat/grep/md5sum/file/wc` only. No `sops -d`, no `config.yaml`/`.env` read,
no credential value printed.

Target under review: **Fix 4 — one-way VM→repo sync of the 7 strong-confidence
live-only skill dirs**, per
`status/probes/2026-08-14-skills-mirror-drift.md:63-71` and `:105`.

**Verdict: FIX-PLAN.** The secrets gate passes, but the candidate set is wrong in
both directions and the file scope is wrong.

---

## 1. SECRETS GATE — PASS (no stripping required)

Every file in the 7 candidate dirs was scanned. 3 files are non-text (all
`.pyc` under `productivity/google-workspace/scripts/__pycache__/`); everything
else is UTF-8 and was grepped line by line.

Patterns scanned: `xox[abprs]-`, `sk-…`, `ghp_…`, `github_pat_`, `AKIA…`,
`AGE-SECRET-KEY`, `BEGIN * PRIVATE KEY`, `(api_key|secret|token|password|
bearer|client_secret)\s*[:=]\s*<12+ chars>`, email addresses, RFC1918 /
CGNAT / `*.ts.net` / `*.internal` hosts, Slack `U…/C…/D…/T…` IDs, and
credential file paths.

| Check | Result |
|---|---|
| Key / token / private-key literals | **0 hits** |
| Assignment-shaped hits | 1, benign — see below |
| Real email addresses | **0** (all placeholders) |
| Internal IPs / tailnet hosts | **0** |
| Slack user / channel / team IDs | **0** (one false positive: the literal `CLAUDECODE` in `software-development/local-ipc-probes/references/claude-code-cross-session-v2.1.232.md:82`) |
| Credential *paths* named (no values) | `token.json`, `.credentials`, `config.yaml`, `.env` — names only |

The single assignment-shaped hit, redacted:

```
productivity/google-workspace/scripts/gws_bridge.py:102:    <REDACTED> = <REDACTED>()
```

It is an identifier assigned the result of a function call, not a literal — a
bundled upstream file (see §2), unchanged since Cluster A. Not hot.

Placeholder emails only, e.g. `productivity/google-workspace/SKILL.md:169`
`user@example.com`, `:193` `alice@co.com`/`bob@co.com`,
`references/gmail-search-syntax.md:53` `accounting@company.com`. No
`gaetan*`, no `mobile.club`, no `cleaq`/`nextmobiles` string anywhere in the 7.

Credential-path mentions are documentation, e.g.
`hermes-operations/hermes-gateway-operations/SKILL.md:15,51` and
`references/slack-mcp-write-policy.md:7,8,19` name `config.yaml` / `.env` as
paths with no values. That is the same shape the repo already accepts:
`skills/orchestration/secure-delta-collectors/references/slack-oauth-user-token.md:9-10`
documents `xoxp-<REDACTED>…` / `xoxe.xoxp-<REDACTED>…` as truncated prefixes.

**Gate verdict: nothing needs stripping. The gate does not block Fix 4.** The
blockers below are semantic, not secret-leakage.

---

## 2. BLOCKER — `productivity/google-workspace` is a BUNDLED Hermes skill; it must not enter the mirror

The VM carries an authoritative provenance oracle the inventory never opened:
`~/.hermes/skills/.bundled_manifest` — 70 lines of `<skill-basename>:<hash>`,
rewritten by Hermes (mtime 2026-08-14 13:46, i.e. live).

```
$ grep -c '^google-workspace:' .bundled_manifest   → 1
$ grep -c '^himalaya:' .bundled_manifest           → 1
$ grep -c '^durable-email-automation:' …           → 0   (same for the other 6 candidates)
```

Corroborating signals, all pointing the same way:

- mode `664`, while **every** confirmed Tars self-edit is `600`
  (`find . -name SKILL.md -perm 600` → 18 files; google-workspace is not one).
- `scripts/google_api.py`, `gws_bridge.py`, `_hermes_home.py`, `setup.py`,
  `references/gmail-search-syntax.md` all mtime `2026-08-07T16:17:59` = Cluster A
  (bulk install) — only `SKILL.md` moved on 08-11.
- `SKILL.md.bak-gcn31-20260811` (mode 664, content mtime 2026-08-07T21:26) — a
  patch applied *on top of* a shipped file.

This is exactly the venv-python local patch already documented in memory
`hermes-upgrade-reapply-local-patches.md`, not a `skill_manage` self-edit.

**Failure scenario:** we land google-workspace in `skills/`. The next
`/upgrade` — one already ran 2026-08-13 and dropped local patches, reverting to
vanilla `6e87d43` — rewrites the VM file from upstream. The repo copy silently
becomes stale. A later drift audit reports `MIRRORED-DIVERGED`, and the
documented resolution ("find out which side moved before overwriting either",
`CLAUDE.md:56`) leads someone to push the repo copy back onto the VM, reapplying
a patch on top of a changed upstream base. The mirror has then manufactured the
exact regression the memory note exists to prevent.

**Drop it from the sync.** The venv-python patch stays where it already lives: a
reapply note, not a mirrored file.

Blast-radius extra for the same dir, if it were synced naively: 3 binary
`__pycache__/*.pyc` (~72 KB) and `SKILL.md.bak-gcn31-20260811`. Repo `.gitignore`
covers `__pycache__/` but nothing else here.

---

## 3. BLOCKER — the "strong 7" is the wrong set: 1 false positive, 5 false negatives

The inventory inferred provenance from `SKILL.md` mtime clustering alone. The
manifest gives a hard membership test. Classifying all 132 live dirs by it:

| Class | Count |
|---|---|
| Bundled with Hermes (in `.bundled_manifest`) | **70** |
| Non-bundled | **62** |
| …of which the mattpocock pack (Cluster B, 2026-08-08T12:00:37) | 35 |
| …of which the ponytail pack (Cluster C, 2026-06-20) | 6 |
| **Personal / agent-authored (non-bundled, non-pack)** | **21** |

The 21, with mode and mtime (`grep -q "^$basename:" .bundled_manifest ||` …):

```
2026-07-31 16:31:01|600|creative/meme-generator              NOT MIRRORED
2026-08-02 15:24:18|600|email/rich-email-composition         NOT MIRRORED
2026-08-07 16:32:30|664|i-have-adhd                          NOT MIRRORED
2026-08-08 10:52:08|600|hermes-operations/hermes-orchestration   mirrored
2026-08-08 12:05:11|664|last30days                           NOT MIRRORED
2026-08-08 12:11:15|600|delegate-to-cooper                       mirrored
2026-08-08 12:47:48|600|hermes-operations/skill-library-auditing  NOT MIRRORED
2026-08-10 09:51:44|600|email/durable-email-automation       NOT MIRRORED *
2026-08-10 12:37:22|600|research/upstream-change-research    NOT MIRRORED *
2026-08-11 12:35:50|600|orchestration/secure-delta-collectors    mirrored
2026-08-11 15:28:29|600|helptech-duty-intake                     mirrored
2026-08-11 16:26:39|600|communication/stakeholder-communication  NOT MIRRORED *
2026-08-12 15:55:56|600|email/gmail-triage                       mirrored
2026-08-13 12:36:06|664|orchestration/slack-channel-context      mirrored
2026-08-13 13:16:09|600|orchestration/daily-work-brief           mirrored
2026-08-13 14:55:48|600|orchestration/engagement-checker         mirrored
2026-08-13 15:18:19|664|orchestration/linear-ticketing           mirrored
2026-08-14 10:53:23|600|software-development/local-ipc-probes    NOT MIRRORED *
2026-08-14 11:27:00|600|hermes-operations/hermes-skill-portability NOT MIRRORED *
2026-08-14 11:27:49|600|creative/mobile-club-okr-report           mirrored
2026-08-14 11:32:55|600|hermes-operations/hermes-gateway-operations NOT MIRRORED *
```

(`*` = in the proposed 7. The 11th "mirrored" dir, `email/himalaya`, is **bundled** —
the repo already mirrors an upstream-owned skill, same category error as §2.)

So the correct gap is **11 non-bundled unmirrored dirs**, not 7:
the 6 correct candidates above, plus `hermes-operations/skill-library-auditing`
(the "medium" one — non-bundled, mode 600, same signature as the strong six;
its mtime proximity to Cluster B is coincidence, the bulk-install dirs are all
in the manifest and it is not), plus `creative/meme-generator` and
`email/rich-email-composition` (non-bundled, mode 600, dismissed by the
inventory purely as "pre-2026-08-07 go-live" — but go-live is the *Slack* date;
the VM predates it), plus `i-have-adhd` and `last30days` (non-bundled, mode 664
→ third-party installs, genuinely owner's-call, and `last30days` drags ~14 MB of
jpeg/mp3/png assets, so exclude it explicitly rather than leave it undecided).

**Failure scenario:** we land the 7, declare the mirror fixed, and the next
skill audit still reads a mirror missing 5 non-bundled skills — including
`skill-library-auditing`, the skill about auditing the skill library. Same
false-negative class as today's 5/55 miss, one iteration later, but now with a
"we fixed it" commit in the log making the next reviewer trust it.

---

## 4. MAJOR — the file scope in the inventory is wrong: 12 reference files would be dropped

`status/probes/2026-08-14-skills-mirror-drift.md:46-57` has a column
"non-SKILL.md files" reporting **0** for six of the seven candidates. Actual
`find <dir> -type f`:

| candidate | probe says | actually |
|---|---|---|
| `email/durable-email-automation` | 0 | **2** (`references/gmail-forwarding-idempotency.md`, `references/historical-backfill.md`) |
| `research/upstream-change-research` | 0 | **1** (`references/github-evidence-workflow.md`, mtime 12:38:05 — *later* than its SKILL.md) |
| `communication/stakeholder-communication` | 0 | **1** |
| `software-development/local-ipc-probes` | 0 | **1** (7966 B, written 8 s before the SKILL.md) |
| `hermes-operations/hermes-skill-portability` | 0 | **4** |
| `hermes-operations/hermes-gateway-operations` | 0 | **3** (incl. `slack-mcp-write-policy.md`, 2861 B, mtime 11:32 — same edit turn) |
| `hermes-operations/skill-library-auditing` | 0 | **2** |

**Failure scenario:** the sync is scoped from that column, lands `SKILL.md`
alone, and the next drift check reports these dirs `MIRRORED-IDENTICAL` while 14
reference files exist only on the VM. That is precisely the partial-coverage lie
that produced today's false negative — reproduced inside the fix for it. Scope
the sync by a positive allowlist over the whole dir, not by that column.

---

## 5. MAJOR — mirror doctrine: pick a machine-checkable rule and write it into CLAUDE.md

`CLAUDE.md:43-45` says `skills/<rel>` is a "live mirror of `~/.hermes/skills/<rel>`".
Today the word means "the subset somebody once landed" — 11 of 132 — and that
ambiguity is what let a skill audit conclude 5/55 coverage without anyone
noticing the denominator was fiction.

Honest alternatives:

- **Sync all 132.** Rejected: 70 are bundled and rewritten by every `/upgrade`
  (one already ran 08-13) → permanent churn, and the repo becomes a competing
  source of truth for upstream files. Plus `last30days` alone is ~14 MB of
  binary assets.
- **Sync only the 7 (as proposed).** Rejected: §3 — differently wrong, and now
  wrong with a commit that reads as authoritative.
- **Leave it and document the subset.** Rejected: a subset defined by "whoever
  remembered" is not auditable; there is no predicate a checker can evaluate.
- **PICK THIS: `mirror = every skill dir NOT in `~/.hermes/skills/.bundled_manifest`,
  excluding installed third-party packs (mattpocock, ponytail).`** 21 dirs today,
  10 already landed, 11 to land (minus whatever the owner excludes from
  `i-have-adhd` / `last30days`). Machine-checkable in one line on the VM, so a
  drift check can assert it instead of a human re-deriving intent from mtimes.
  Write the rule and the two named pack exclusions into `CLAUDE.md` next to
  line 43, and note there that a bundled skill's local patch (google-workspace)
  is tracked as a reapply note, never mirrored.

---

## 6. MAJOR — who lands it: recommend (c), and close the detection gap

**Recommend (c): we land them, with a provenance note in each commit, plus a
separate detection fix. Do not hand the retroactive landing to Tars.**

Against (b) — Tars runs its own rule-2 flow for each:

- It **writes to the VM** (resolve the live file, tmp+mv) and therefore hits the
  `chmod 600` hard rule (`CLAUDE.md:95-100`) — a rule added 2026-08-13 precisely
  because a tmp+mv left a credential-bearing file world-readable. Seven
  unattended repeats of a write path with a one-day-old guardrail is the wrong
  place to practise.
- GCN-13 precedent: the previous flat-path recipe **merged six empty SKILL.md
  files** (`CLAUDE.md:60-66`). Seven retroactive self-merged PRs, unwatched, is
  the same shape of run.
- Rule 2 is scoped to "the file this turn edited". Retroactively re-landing an
  edit from 08-10 is not that turn; it invites Tars to rewrite the live file to
  produce a diff.

Against (a) alone: landing on Tars' behalf without a detection fix means the
next silent miss is found by the next manual audit, whenever someone thinks to
ask.

The detection gap is real and narrow. Counter-evidence that the flow itself
works: `creative/mobile-club-okr-report` SKILL.md mtime 2026-08-14T11:27:49,
repo commit 11:28:49 — a 60 s lag, and 10 of the 11 mirrored dirs show the same
sub-3-minute pattern. So rule 2 fires when Tars is in the loop; what is missing
is anything that notices when it does not. Cheapest closure consistent with the
§5 rule:

```
ssh tars 'cd ~/.hermes/skills && while read -r p; do d=$(dirname "$p"); \
  grep -q "^$(basename "$d"):" .bundled_manifest || md5sum "$p"; done \
  < <(find . -name SKILL.md)'
```
diffed against the repo's `skills/` on a daily cron; any non-bundled live path
absent from the repo, or md5-mismatched, is a rule-2 miss. One command, no new
module. Ship it with the sync, not as its own project.

Separately worth a ticket, not a blocker: *why* 7 misses were silent. Both
plausible causes are cheap to check — the skill dirs missing are ones Tars
created (not edited), so a rule-2 recipe keyed on "edit" may simply not cover
"create".

---

## 7. MAJOR — race/state: the plan has no md5 gate, and the tree is live right now

The proposal as written does not md5-gate at copy time. It must.

- VM clock at review time: `Fri Aug 14 01:49:08 PM UTC 2026`.
- Three of the seven were written **today**: `local-ipc-probes` 10:53:23,
  `hermes-skill-portability` 11:27:00, `hermes-gateway-operations` 11:32:55.
- `~/.hermes/skills/.usage.json` mtime **13:48** and `.bundled_manifest` **13:46**
  — Hermes was writing into the skills tree two minutes before this check.

**Failure scenario:** Tars edits `hermes-gateway-operations/SKILL.md` at 14:05
and lands its own rule-2 PR; our sync, computed from the 11:32 read, opens a
second PR on the same path minutes later and squash-merges the older bytes over
the newer. The repo now holds a version that never existed on the VM, and the
next drift check reports `DIVERGED` with no way to tell which side moved.

Required: capture md5 at copy time, re-`md5sum` the VM file immediately before
commit, abort that dir on mismatch; **one PR per skill dir** so a collision is
confined to one path; and record the copy-time md5 in the commit message so a
future audit can tell which VM state was landed.

**.bak litter — exclude it.** 17 `*.bak*` files live in the tree. Backups are
pre-edit snapshots; git history is the history, so mirroring them duplicates it
badly and makes every future md5 comparison noisier. But exclude by *allowlist*,
not by `*.bak*`: `orchestration/daily-work-brief/SKILL.md.wf6-fabricating-board-20260810T201445Z`
is a backup that does **not** match `*.bak*` — the inventory's own filter would
have swept it in. Allowlist: `SKILL.md`, `references/**`, `scripts/**`,
`templates/**`.

---

## 8. VM writes — the plan is read-only as written; two ways it stops being

Nothing in Fix 4 as proposed writes to the VM. Flags:

- Option (b) in §6 **does** write to the VM. That alone disqualifies it from
  being the low-risk option it sounds like.
- The implementation must put the VM strictly on the **source** side:
  `ssh gaetan@192.168.0.9 'cd ~/.hermes/skills && tar c <dirs>' | tar x -C <repo>/skills`,
  or `scp -r` / `rsync` with the VM as source and **never** `--delete`. A
  reversed `rsync` argument order writes the repo onto the VM and, with
  `--delete`, removes live skills.
- Do **not** "normalise" the 600/664 mode split on the VM. It is a write, and
  the mode is load-bearing provenance evidence (§2, §3). Git stores only the
  executable bit anyway, so there is nothing to gain.

---

## Recommended Fix 4′

1. Drop `productivity/google-workspace` (bundled; patch tracked as a reapply note).
2. Land the other 6 **plus** `hermes-operations/skill-library-auditing`,
   `creative/meme-generator`, `email/rich-email-composition` — 9 dirs, whole-dir
   allowlist (`SKILL.md`, `references/**`, `scripts/**`, `templates/**`), no backups.
3. Owner's call on `i-have-adhd` and `last30days` only (third-party installs;
   `last30days` is ~14 MB of assets — recommend excluding, with the reason
   recorded in CLAUDE.md).
4. We land them (option c), one PR per dir, md5 re-checked against the VM
   immediately before each commit, provenance + copy-time md5 in each message.
5. Same PR series: write the mirror rule into `CLAUDE.md` (§5) and add the daily
   md5 drift check (§6).
6. Secrets gate: PASS, re-run the same scan against whatever the final dir list
   is before committing.
