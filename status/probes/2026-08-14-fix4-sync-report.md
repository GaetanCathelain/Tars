# FIX 4 sync execution — VM→repo copy report (2026-08-14)

Executed the reconciled FINAL list from
`status/probes/2026-08-14-fix4-reconciliation.md` §5. **Copy only** — no git
command was run in this session (add/commit/push/pull all deliberately
skipped; orchestrator commits separately). VM (`ssh gaetan@192.168.0.9`)
touched read-only: `md5sum`/`find`/`tar cf -` as source only, no writes, no
`sops -d`, no `config.yaml`/`.env` read.

## Step 1 — re-md5 on VM vs reconciliation's recorded copy-time md5s

All 21 allowlisted files re-hashed on the VM now; **0 drift** since the
reconciliation ran (every md5 identical to the recorded value). `find <dir>
-type f` per dir also re-run: each of the 7 dirs contains **exactly** the
allowlisted files, nothing extra (no `.bak*`, `__pycache__`, or
`wf6-fabricating` siblings) — confirms the reconciliation's "filters remove
nothing" claim independently.

## Step 2 — per-dir / per-file table

| dir | file | VM md5 (now) | repo md5 (after copy) | match | drift since reconciliation |
|---|---|---|---|---|---|
| `communication/stakeholder-communication` | SKILL.md | `02efed76fea84d67ace94d9e4bc66167` | `02efed76fea84d67ace94d9e4bc66167` | YES | none |
| | references/conversation-vs-project-overview.md | `6326fabb2588c10c5a08b9b9d8699c23` | `6326fabb2588c10c5a08b9b9d8699c23` | YES | none |
| `email/durable-email-automation` | SKILL.md | `6c1e799bbe075f73bc545524ac6bbe01` | `6c1e799bbe075f73bc545524ac6bbe01` | YES | none |
| | references/gmail-forwarding-idempotency.md | `3d85d40cf0cd3d0be82efda4a89bf1dd` | `3d85d40cf0cd3d0be82efda4a89bf1dd` | YES | none |
| | references/historical-backfill.md | `1ec9df2fb7365427d612b90c947bc8aa` | `1ec9df2fb7365427d612b90c947bc8aa` | YES | none |
| `hermes-operations/hermes-gateway-operations` | SKILL.md | `5ba519d4ea738e92e90fcc380cfdf0f8` | `5ba519d4ea738e92e90fcc380cfdf0f8` | YES | none |
| | references/slack-busy-ack-diagnostics.md | `69c6d453a04878e06335db1ba6803d58` | `69c6d453a04878e06335db1ba6803d58` | YES | none |
| | references/slack-mcp-write-policy.md | `1cb4580f67584ba26284c26c83853b45` | `1cb4580f67584ba26284c26c83853b45` | YES | none |
| | references/slack-progress-scoping.md | `4edb5f47627fc621c0b4f024e12c7a2f` | `4edb5f47627fc621c0b4f024e12c7a2f` | YES | none |
| `hermes-operations/hermes-skill-portability` | SKILL.md | `398b0e72380d261d21b730df61991257` | `398b0e72380d261d21b730df61991257` | YES | none |
| | references/deleted-package-forensics.md | `cde88dca19f35822e69ab5b3057bfccf` | `cde88dca19f35822e69ab5b3057bfccf` | YES | none |
| | references/deterministic-package-verification.md | `588538560ecacdde823c367dadd474eb` | `588538560ecacdde823c367dadd474eb` | YES | none |
| | references/migration-checklist.md | `2cc75977d7c5d7b72f218701b9627ef1` | `2cc75977d7c5d7b72f218701b9627ef1` | YES | none |
| | references/shared-cross-agent-install.md | `724a4a40fb23896e70711efd83ed59ae` | `724a4a40fb23896e70711efd83ed59ae` | YES | none |
| `hermes-operations/skill-library-auditing` | SKILL.md | `a4d927e91ee4011ad08958d6e30fa933` | `a4d927e91ee4011ad08958d6e30fa933` | YES | none |
| | references/checksum-diff-procedure.md | `bcc47f91386c04519b493c7f283792d3` | `bcc47f91386c04519b493c7f283792d3` | YES | none |
| | references/exact-collision-recovery.md | `84cdf6a1abcac10198c919ab214b71f1` | `84cdf6a1abcac10198c919ab214b71f1` | YES | none |
| `research/upstream-change-research` | SKILL.md | `4235623ebf922843d14a0d3904524f37` | `4235623ebf922843d14a0d3904524f37` | YES | none |
| | references/github-evidence-workflow.md | `17f03949d7bb904efd39917fe1040b07` | `17f03949d7bb904efd39917fe1040b07` | YES | none |
| `software-development/local-ipc-probes` | SKILL.md | `0869cd98b477ee5f633e8db865038ea9` | `0869cd98b477ee5f633e8db865038ea9` | YES | none |
| | references/claude-code-cross-session-v2.1.232.md | `1135be2ba0c8007cf0d423e84c4ccb51` | `1135be2ba0c8007cf0d423e84c4ccb51` | YES | none |

**21/21 files copied, 21/21 md5-verified, 0 drift, 0 skips.**

## Step 3 — copy method

`tar cf -` over ssh (source read only, `~/.hermes/skills/` on the VM),
piped to a scratch file, extracted into
`/home/gaetan/dev/orca-worktrees/Tars/improvements/skills/` with `tar xf`.
Byte-exact (tar preserves content; no newline/encoding transform possible via
this path). Destination dirs `communication/`, `research/`,
`software-development/` did not exist in the repo before this run (confirmed
by the reconciliation's §5 direction check — 0 REPO-ONLY) and were created
fresh; `email/durable-email-automation` and the three
`hermes-operations/*` dirs are new subdirs under existing category dirs. Repo
file modes: default from `tar x` (644), irrelevant per the task's own
allowance.

## Step 4 — secrets gate on copied repo files

```
grep -rlE 'xox[abprse]-|sk-…|ghp_…|github_pat_|AKIA…|AGE-SECRET-KEY|
           BEGIN …PRIVATE KEY|192\.168\.0\.[0-9]+|100\.x\.x\.x|\.ts\.net'
  over all 7 copied dirs → 0 hits
```

Matches the reconciliation's own §6 result on the same file set. No file
deleted (nothing to delete).

## Junk-file check

`find <7 dirs> \( -name '*.bak*' -o -name '*wf6-fabricating*' -o -name
'__pycache__' \)` → 0 hits, both on VM source and in the copied repo tree.

## Totals

21 files, 89,827 B (~87.7 KiB) on disk in the repo — consistent with the
reconciliation's "~90 KB" estimate (its number was file-size-only from the
table; on-disk total via `du -cb` above).

## Skips

**None.** All 7 dirs from the FINAL list qualified on re-check and were
copied in full; no file had drifted since the reconciliation ran.

## Scope note

Per the ROLE constraint, this session copied files only. `git status` was
not run and no `git add`/`commit`/`push`/`pull` was issued — the orchestrator
owns staging/commit to avoid index races with a concurrent agent.
