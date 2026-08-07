# Probe — mc-metarepo as a knowledge base (P6/kb)

All commands run read-only via `ssh gaetan@192.168.0.9` against the existing
clone at `~/dev/mc-metarepo` on the Tars VM. No `git fetch`/`pull`/`checkout`/
`commit` was run against it; no `sops -d` was run; no secret value was printed.

> **Corrected 2026-08-08** after re-verification: `knowledge/` holds **40 notes
> + 1 README**, not 41 notes. `ls ~/dev/mc-metarepo/knowledge/*.md | wc -l` →
> `41`, and that count includes `knowledge/README.md`. Two claims fixed below.

## Summary

- Tiny repo: **2.2 MB working tree** (3.8 MB incl. `.git`), **169 files**, **120
  Markdown files totalling ~1.44 MB**. A grep/search-tool design is sane; the
  whole tree could *almost* fit one prompt but that is exactly the wrong
  design per the task brief — retrieval must be tool-driven, not pasted.
- Tree is **clean**, `main` up to date with `origin/main` **as of the last
  fetch** (a fresh clone made ~3.5h before this probe — reflog shows only one
  entry, `clone: from https://github.com/mobile-club/metarepo.git`). No live
  drift check was run (no fetch allowed).
- Auth is **`gh` OAuth device-flow token** (`gho_…`, scopes `gist`,
  `read:org`, `repo`) via git's `credential.helper = !/usr/bin/gh auth
  git-credential`, remote is **HTTPS**. No expiry metadata found locally;
  failure mode measured indirectly (see below).
- **`main` is PR-protected**: 1 required approving review, no CODEOWNERS file
  exists (`require_code_owner_reviews: false`), no required status checks, no
  push restrictions, admins not exempted, force-push and branch deletion both
  blocked.
- **Directory-scoped write rules are convention, not GitHub-enforced**: they
  live in `CLAUDE.md`'s "Layout & write rules" table and in prose contracts
  (`learnings/README.md`, `CONTRIBUTING.md`) — there is no CODEOWNERS, so
  nothing at the GitHub-API level stops a PR from touching the wrong
  directory; the human reviewer is the actual gate.
- `learnings/` (agent-contributed knowledge) is explicit: **"Agent-written,
  human-merged — always... agents MUST NEVER self-merge."** — this is the
  single most binding rule for a P6 push-back design.
- 10 registered git submodules (the actual product repos) are **present but
  uninitialized** (empty dirs) — none of their content counts toward the size
  above; the metarepo itself is docs + JSON + scripts only.
- The repo already ships its own **design doc for exactly this problem**:
  `MCP.md` — a "Knowledge MCP" proposal (filesystem-server + grep, explicitly
  "no vector DB, no embeddings" — KISS) that Tars' P6 design should read and
  either mirror or explicitly diverge from.

## Shape & size

```
$ du -sh ~/dev/mc-metarepo                      → 3.8M  (incl. .git)
$ du -sh --exclude=.git ~/dev/mc-metarepo       → 2.2M  (working tree)
$ du -sh ~/dev/mc-metarepo/.git                 → 1.6M
$ find … -type f | wc -l                        → 169 files (excl. .git)
$ find … -name '*.md' | wc -l                   → 120 markdown files
$ find … -name '*.md' -print0 | xargs -0 stat -c%s | awk '{s+=$1} END{print s}'
                                                 → 1,478,742 bytes (~1.44 MB) of markdown
```

Directory tree (depth 2, `.git` pruned):

```
CLAUDE.md  CONTRIBUTING.md  README.md  PLAN.md  SKILLS.md  MCP.md  .mcp.json  .sops.yaml  .gitmodules
.claude/            settings.json, skills/ (18 skills, see SKILLS.md)
.githooks/          post-checkout (submodule + .env auto-provision)
.github/workflows/  ste-lint.yml, auto-update-submodules.yml, validate-index.yml
cleaq/               6 pinned per-repo docs (backend-api, frontend-next, keycloak, backend-dev-env, cleaq-tools, risk)
mobileclub/          9 pinned per-repo docs (workspace, terraform, strapi, ganesh, Soupy_API, pim-sync, pim-migration-sync, return-app, signaturemail)
next/                9 pinned docs (vecna, wordpress-monorepo, helm-charts, terraform-infrastructure, next-pricing, next-scoring, next-subscription, NextShadow-v2, next-bdc + next-bdc.fiches)
index/               cleaq/ mobileclub/ next/ — one JSON record per system + _schema.json + README.md
knowledge/            40 cross-cutting gotcha notes + README.md
learnings/            cleaq/ mobileclub/ next/ sidecars + README.md
schemas/              cleaq/ mobileclub/ next/ — DB schema dumps + README.md
scripts/              9 scripts (extract-next-bdc-fiches.mjs, helptaker*, ste-lint.mjs, setup-hooks.sh, to-review.sh, validate_index.mjs, wt-load) + README.md
secrets/              secrets.enc.yaml (SOPS+age) + README.md
cleaq-api/ cleaq-front/ vecna-api/ vecna-front/ wordpress/ return-app/
support-engineer/ risk/ monorepo-loop/ pim-migration-sync/   ← 10 registered submodules, all EMPTY (uninitialized)
```

10 largest files (bytes, excl. `.git`):

```
155963  schemas/mobileclub/loop-postgres.full.md
151616  next/next-bdc.fiches.md
140558  schemas/next/wordpress-mysql.full.md
102964  schemas/next/vecna-mysql.full.md
 70381  learnings/mobileclub/workspace.md
 63634  learnings/mobileclub/support-engineer.md
 49711  learnings/next/vecna.md
 41355  PLAN.md
 29638  scripts/helptaker.mjs
 29307  .claude/skills/helptaker-du-jour/SKILL.md
```

The four DB-schema dumps and `next-bdc.fiches.md` (auto-generated) dominate;
everything else is well under 40 KB per file.

## Git & auth state

```
$ git -C ~/dev/mc-metarepo status --porcelain   → (empty, exit 0) — clean
$ git -C ~/dev/mc-metarepo status
On branch main
Your branch is up to date with 'origin/main'.
nothing to commit, working tree clean

$ git -C ~/dev/mc-metarepo log -1 --format="%H %ci %s"
4831253a24928e57b76fcb2870df95705706b6f2 2026-08-07 18:17:48 +0200 chore(submodules): auto-update to latest tracked branches (#181)

$ git -C ~/dev/mc-metarepo rev-parse --abbrev-ref --symbolic-full-name @{u}
origin/main

$ git -C ~/dev/mc-metarepo remote -v
origin  https://github.com/mobile-club/metarepo.git (fetch)
origin  https://github.com/mobile-club/metarepo.git (push)

$ stat -c "%y" ~/dev/mc-metarepo/.git/FETCH_HEAD
2026-08-07 19:00:54.715604986 +0000        # probe ran 22:37 UTC same day → ~3.5h old

$ git -C ~/dev/mc-metarepo reflog -10
4831253 HEAD@{0}: clone: from https://github.com/mobile-club/metarepo.git   # ONLY entry
```

**How far behind origin/main**: cannot be determined live without `git fetch`
(banned). Local `main` and the cached `origin/main` ref are identical
(`4831253a…`), but that is only as fresh as the clone/last-fetch timestamp
above (~19:00 UTC today) — report this as *last-known*, not live.

Auth:

```
$ gh auth status
github.com
  ✓ Logged in to github.com account GaetanCathelain (/home/gaetan/.config/gh/hosts.yml)
  - Active account: true
  - Git operations protocol: https
  - Token: gho_************************************
  - Token scopes: 'gist', 'read:org', 'repo'

$ grep -oE '^\s*[a-z_]+:' ~/.config/gh/hosts.yml     # keys only, no values printed
    users:
            oauth_token:
    git_protocol:
    oauth_token:
    user:
# no "expires"/refresh_token key present at all

$ git config --global --get-all credential.helper   # (none globally)
$ cat ~/.gitconfig | grep -A3 credential
[credential "https://github.com"]
        helper =
        helper = !/usr/bin/gh auth git-credential
```

- Remote is **HTTPS**, not SSH.
- Auth flows through the **`gh` CLI's stored OAuth token** (`gho_…` prefix =
  device-flow OAuth, not a fine-grained/classic PAT), via git's credential
  helper delegating to `gh auth git-credential`.
- No expiry field or refresh_token is stored locally — cannot conclusively
  say whether/when it expires without invalidating it (not tested; that would
  be destructive).
- **Measured failure mode** (fake token, does not touch the real credential):

  ```
  $ GH_TOKEN=ghp_deliberatelyInvalidTokenForProbe0000 gh api user
  {
    "message": "Bad credentials",
    "documentation_url": "https://docs.github.com/rest",
    "status": "401"
  }
  gh: Bad credentials (HTTP 401)
  ```

  And confirmed the *real* token is currently live with a read-only GET:
  `gh api rate_limit --jq '.rate'` → `{"limit":5000,"remaining":5000,...}`.
- **Not tested**: the exact `git pull`/`fetch` failure string with lapsed
  auth — `git fetch`/`pull` against this clone is banned by the task's hard
  rules regardless of token validity, so this was not exercised. Expected
  behavior (inferred from standard git+credential-helper semantics, not
  measured): `git fetch` would call the same credential helper, get a 401 from
  GitHub, and abort with `remote: Invalid username or password.` / `fatal:
  Authentication failed for 'https://github.com/mobile-club/metarepo.git/'`.
- Repo visibility: **private** (`gh api repos/mobile-club/metarepo --jq
  '{visibility}'` → `"private"`) — every pull, including the hourly re-pull
  P6 proposes, is auth-gated, not just the initial clone.

## Contribution rules (verbatim quotes)

Files read in full: root `CLAUDE.md`, `CONTRIBUTING.md`, `README.md`,
`learnings/README.md`, `knowledge/README.md` (head), `secrets/README.md`, plus
`.github/workflows/*.yml`. **No `AGENTS.md` and no `CODEOWNERS` exist anywhere
in the repo** (`find … -iname AGENTS.md -o -iname CONTRIBUTING.md -o -iname
CLAUDE.md` → only `CLAUDE.md` and `CONTRIBUTING.md`; `find … -iname
CODEOWNERS` → empty).

**Layout & write rules table, `CLAUDE.md`** (the single most load-bearing
table in the repo for a KB-push design):

> | Path | What | Write rule |
> |---|---|---|
> | `cleaq/ mobileclub/ next/` (`*.md`) | pinned per-repo docs, `/init`-generated from source@sha (header comment) | **never hand-edit** — regenerate (see [README](README.md#regenerating-a-doc)) |
> | `index/` | machine-readable JSON per system ([index/README](index/README.md)) | derived from pinned docs; never hand-edit `sources[].sha`/`generated` to appease CI |
> | `learnings/` | agent-contributed sidecars | append via PR labeled `agent-learning`, **human-merged only**; regeneration never touches it |
> | `knowledge/` | cross-cutting gotchas (auto-loaded below) | hand-edited; new note ⇒ add `@import` below |
> | `<stack>/INFRASTRUCTURE.md`, `next/next-bdc.fiches.md` | generated syntheses | regenerate, never hand-edit |
> | `.claude/skills/`, `scripts/`, `secrets/` | shared skills / scripts / SOPS vault | hand-edited; index new entries in `SKILLS.md` / `scripts/README.md` |

**`learnings/README.md` §"The contract"** (rules 1, 2, 5, 7 quoted verbatim —
this is the file most relevant to any future automated push-back):

> 1. **Agent-written, human-merged — always.** Entries arrive ONLY as pull
> requests labeled `agent-learning`, opened by an agent against `main`.
> **A human MUST review and merge every learnings PR; agents MUST NEVER
> self-merge.** The human review is the quality gate that keeps wrong or
> trap knowledge out of the source of truth.
>
> 2. **The regeneration pipeline MUST NEVER touch `learnings/`.** ... `learnings/`
> is additive, human-curated memory and is strictly OUT of the regeneration
> pipeline's write scope — manual regeneration today, and any future automated
> pipeline (PLAN.md WS1), must both exclude this directory entirely.
>
> 5. **No PII, no secrets — ever.** Schema/system facts only: table, column and
> state NAMES, vocabularies, code paths, seams, behaviours. Never a customer
> name/email/address, an IMEI or serial *value*, row data, a credential, a
> token, or a DSN. Env var NAMES are acceptable; values never.
>
> 7. **Tech debt gets a Linear issue, not an entry.** ... An entry that is a
> defect plus its workaround with no ticket is not merged — see
> [../CONTRIBUTING.md](../CONTRIBUTING.md#tech-debt-is-a-linear-issue-not-a-learning).
> Reviewers: this is a merge gate, like rules 1 and 5.

**`CONTRIBUTING.md` §"Language"** (hard rule, applies to every committed
surface and to Linear):

> **All content committed to this repo MUST be written in English** — every
> Markdown doc (pinned `/init` docs, `INFRASTRUCTURE.md`, `knowledge/`,
> `learnings/`, this file, `README.md`, `PLAN.md`), every skill under
> `.claude/skills/`, and every commit message, PR description and code comment.
> This is a hard rule, decided with the team: no French (or any other language) in
> committed content.

And, layered on top, **ASD-STE100 Simplified Technical English** is also
required and CI-gated on diffs:

> Write every artefact listed above in **ASD-STE100 Simplified Technical
> English**... CI fails a pull request on an STE **error** in an added line,
> and lets a warning through.

**`CONTRIBUTING.md` §"Doc PRs: expect conflicts, resolve by union"**:

> Parallel doc PRs (several sessions harvesting at once) always collide on the
> same three surfaces: the `@knowledge/*` import list in `CLAUDE.md`, the index
> table in `knowledge/README.md`, and section appends in `learnings/*/*.md`.
> All three are **append-only by convention**, so the resolution is always a
> **pure union — keep both sides, drop the markers** — never a pick between
> them.

**`CONTRIBUTING.md` §"Always pass `--repo` to `gh`, and prove a merge two
ways"** (relevant if a future delegated Orca session ever pushes into this
repo or a submodule):

> An agent session's shell **cwd resets between commands**, back to the metarepo
> root. So in a multi-repo worktree a bare `gh pr merge 135`, `gh pr view`, or
> `git ls-remote origin main` resolves against **`mobile-club/metarepo`** — not the
> submodule you were working in. On 2026-08-05 that merged an unrelated PR in this
> repo, one touching `learnings/`, which the [learnings contract] says is
> **human-merged only**...
> - Pass **`--repo <owner>/<name>`** on every `gh` call and **`-C <dir>`** on
> every `git` call, including read-only ones.
> - Prove a merge **two ways**: `gh pr view <n> --repo <o>/<r> --json
> state,mergeCommit` returning `MERGED` + a SHA, **and** `git -C <dir>
> ls-remote origin main` returning that same SHA.

**`secrets/README.md`** (not a contribution rule per se, but a hard boundary
any KB-reading design must respect): `secrets/secrets.enc.yaml` is SOPS+age
encrypted, shared team key, values never in plaintext in the repo — a
search/grep tool over this repo will see ciphertext blobs only, never a
secret value, which is safe by construction.

## Branch protection / PR requirements

No CODEOWNERS file exists, so protection is entirely the GitHub branch-rule
config, read via read-only `gh api` GETs (no writes):

```
$ gh api repos/mobile-club/metarepo/branches/main/protection
{
  "required_signatures": {"enabled": false},
  "enforce_admins": {"enabled": true},
  "required_linear_history": {"enabled": false},
  "allow_force_pushes": {"enabled": false},
  "allow_deletions": {"enabled": false},
  "block_creations": {"enabled": false},
  "required_conversation_resolution": {"enabled": false},
  "lock_branch": {"enabled": false},
  "allow_fork_syncing": {"enabled": false}
}

$ gh api repos/mobile-club/metarepo/branches/main/protection/required_pull_request_reviews
{"dismiss_stale_reviews": false, "require_code_owner_reviews": false,
 "require_last_push_approval": false, "required_approving_review_count": 1}

$ gh api repos/mobile-club/metarepo/branches/main/protection/required_status_checks
{"message": "Required status checks not enabled", "status": "404"}

$ gh api repos/mobile-club/metarepo/branches/main/protection/restrictions
{"message": "Push restrictions not enabled", "status": "404"}

$ gh api repos/mobile-club/metarepo --jq '{allow_merge_commit, allow_rebase_merge, allow_squash_merge, delete_branch_on_merge, default_branch, visibility}'
{"allow_merge_commit": true, "allow_rebase_merge": false, "allow_squash_merge": true,
 "delete_branch_on_merge": true, "default_branch": "main", "visibility": "private"}
```

Summary: `main` requires a PR with **≥1 approving review**, no code-owner
requirement (none exist), no required status checks, no push restrictions,
`enforce_admins: true` (admins are NOT exempt), force-push and branch-delete
both blocked. Squash-merge is the norm (`delete_branch_on_merge: true`,
`allow_rebase_merge: false`).

**Discrepancy worth flagging for P6**: `.github/workflows/auto-update-submodules.yml`
comments *"main is protected: no direct push. We open a PR and squash-merge it
right away (0 reviews / 0 status checks required)."* — but the live branch
protection above shows `required_approving_review_count: 1`. Either the rule
was tightened after that workflow was authored, or the bot's PAT-authored PR
merge has been failing/blocked since. Not verified further (would require
either triggering the workflow or reading its run history, both out of scope
for a read-only clone probe) — flag this for whoever designs the P6 push-back
path, since it directly answers "can an automated PR self-merge here" (answer:
**apparently no, today** — contra the workflow's own comment).

## Content inventory

| Area | One-line description |
|---|---|
| `cleaq/`, `mobileclub/`, `next/` | Pinned, `/init`-generated per-repo docs (24 files) — architecture/stack summary snapshotted from a source repo @ commit SHA, regenerate-only |
| `index/` | Machine-readable JSON, one record per system (stack, purpose, deploy mechanism, Datadog service, seams, provenance), schema-validated in CI |
| `knowledge/` | 40 cross-cutting operational gotchas, one fact per file, auto-loaded into `CLAUDE.md` for the whole team (Datadog sampling traps, NM-vs-MC discriminants, bastion routing, secret-logging pitfalls, org OKRs, ...) |
| `learnings/` | Agent-contributed, human-merged sidecar knowledge per repo — durable schema facts and runbook-worthy insights not auto-loaded |
| `schemas/` | Full DB schema dumps (Postgres/MySQL) for `mobileclub`, `next` stacks — the 3 largest files in the repo |
| `scripts/` | Zero-dependency Node/bash tooling: index validation, STE linter, next-bdc fiche extractor, submodule loader, PR-review notifier |
| `.claude/skills/` | 18 shared Claude Code skills (triage flows, PR shipping, HelpTech duty rotation, Orca orchestration) |
| `secrets/` | SOPS+age encrypted team vault (ciphertext only, committed) |
| `PLAN.md` | The group's tech strategy doc ("Plan SOTA 2030") — workstreams, principles, success metrics, risks |
| `MCP.md` | Design doc proposing a "Knowledge MCP" server for this exact repo — filesystem+grep, no vector DB, directly relevant prior art for P6 |
| `CONTRIBUTING.md` / `CLAUDE.md` / `README.md` | Governance: language rule (English + STE), write-rule table, repo↔directory mapping, `gh`/`git` cwd-reset trap |
| 10 submodules (`vecna-api`, `wordpress`, `cleaq-api`, ...) | Registered but **uninitialized** — the live product-repo checkouts are not present in this clone |

## Candidate test questions

1. How do you tell a Next Mobiles order apart from a Mobile.Club order in
   api-v2? → `knowledge/nm-vs-mc-order-source.md`
2. Why can a raw Datadog span search show a ~92% error rate for
   `VecnaApi.postOrderStatus` when the service is actually healthy? →
   `knowledge/datadog-apm-error-sampling.md`
3. Why does connecting to a private Mobile Club endpoint (e.g. staging
   Postgres) fail when you SSH into the Terraform-managed
   `tag:company-bastion` host first, instead of going direct from the
   tailnet? → `knowledge/mc-bastion-naming-trap.md`
4. What happens when you run `sudo env SECRET=... <cmd>` on a shared VM, and
   what is the correct remediation if it already happened? →
   `knowledge/sudo-argv-secrets-authlog.md`
5. Which MCP tool queries Vecna's **production** MySQL database, and which
   one is sandbox? → `learnings/next/vecna.md` (2026-07-20 "SE (access /
   tooling)" entry)
6. For a NextMobiles order, who dispatches the customer-facing "bon de
   retour" (BR) notification — Loop or Vecna? → `learnings/next/vecna.md`
   (2026-07-20 "MC-3637" entry)
7. Is `orders.business_entity_id` a reliable way to discriminate NM vs MC
   orders? → `knowledge/nm-vs-mc-order-source.md` (§"What not to do" —
   answer: no)
8. What is the Q3 2026 status of the OKR "50% of HelpTech tickets resolved by
   AI" (O2-KR1), and what is the one caveat about reading its underlying
   number? → `knowledge/tech-product-okr-q3-2026.md`
9. What are the three currently-sold Next Mobiles subscription formulas per
   the customer-service knowledge base? → `next/next-bdc.fiches.md` (§01
   "Offres et abonnements")
10. Which local directory does the `Next-Mobiles/api` ("Vecna") repo map to,
    and what is the one word the team uses to refer to it? →
    `CONTRIBUTING.md` (§"Repository mapping" table)
11. What command loads just the `wordpress` submodule into an empty worktree,
    and roughly how long does it take? → `CLAUDE.md` (§"Worktrees — submodules
    load on demand")
12. Which CI workflow validates `index/**/*.json` records against their
    pinned doc's SHA/date header, and on what triggers does it run? →
    `.github/workflows/validate-index.yml` (+ `index/README.md` for why)

## Not tested:

- Live drift vs `origin/main` (would require `git fetch`, banned by the
  task's hard rules) — reported last-known only, from a clone made ~3.5h
  before this probe.
- The exact `git fetch`/`pull` error string on lapsed auth (banned target;
  only the underlying `gh api` 401 "Bad credentials" was measured, via a
  deliberately fake token that never touched the real stored credential).
- Whether the real stored `gh` OAuth token has a hard expiry — no expiry/
  refresh_token field is stored locally, but this was not stress-tested
  (would require invalidating the live token).
- Whether `auto-update-submodules.yml`'s automated PR actually still
  self-merges today, given the live 1-review requirement contradicts its own
  "0 reviews required" comment — not triggered, not checked against run
  history.
- Contents of the 10 submodules (all uninitialized in this clone) — out of
  scope; this probe covers the metarepo shell only, per the task brief.
- `.claude/skills/` content beyond `SKILLS.md`'s index and the two largest
  skill files' presence — not read in full (18 skills, out of scope for a
  contribution-rules probe).
