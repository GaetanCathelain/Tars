# T6 — cooper-side Linear default team (GCN): surface map + verification

Ticket: **GCN-6** (`cooper defaults: Claude/Orca sessions create Linear tickets
in GCN by default`, label `discovery`, prio Medium, depends T1).
Session B `wf6-cooper-defaults`, 2026-08-10. All work on cooper; **no ssh to the
Tars VM**. This session wrote nothing to `~/.claude` or `~/dev/gaetan-metarepo`
— the metarepo edit was applied by the coordinator (ruling A on question T6-1).

## Verdict

**T6's primary goal is achieved and behaviourally verified.** The brief's
assumed surface (`to-tickets`, matt-pocock-skills plugin) is **not** the path
cooper sessions take; the only live create-path is the account-level claude.ai
Linear connector, whose `team` argument is required with no default. T6
therefore reduced to **one line in `~/dev/gaetan-metarepo/preferences/tooling.md`**,
the file imported into every session on this machine. Applied by the coordinator
as metarepo commit **`3065f7d`**. A fresh session now answers `GCN` unprompted.

**One known limitation, not a defect.** The line's company-repo carve-out binds
on a capable model but not on a weak one: from `cleaq-api`, Sonnet correctly
resolves to the **Cleaq** team, while Haiku falls through to GCN. The clause
works — it just relies on inference rather than an explicit mapping. Evidence
and both transcripts in §Verification test 3; optional hardening filed as
**T6-2**. Not applied here: `preferences/` is single-writer and out of scope.

## Surfaces checked (18 examined; only these can mint a Linear issue)

| Surface | Creates Linear issues? | Team default |
|---|---|---|
| `mcp__claude_ai_Linear__save_issue` — claude.ai **account** connector, no file on disk, present in *every* session on this account | **YES — the live path** | none; `team` required on create |
| `linear-cli` (`https://mcp.linear.app/mcp`) in 3 project `.mcp.json`: `mc-metarepo-everything-else/cleaq-api`, `orca-worktrees/mc-metarepo/graphify-mc-metarepo/{cleaq-api,cleaq-front}` | YES, when that project is loaded | none |
| `to-tickets` / `to-issues` (mattpocock-skills 1.2.3) | indirect, **inert today** | n/a — see §to-tickets |
| `mobile-club/workspace/.claude/settings.local.json:161` → `mcp__linear-server__save_issue` | **dead** — no MCP server by that name exists anywhere | n/a |
| `qa`, `request-refactor-plan` | **No** — `gh issue create`, GitHub only | n/a |
| `retro` | **No** — references an existing Linear ticket, never writes | n/a |
| Tars' own `engagement-checker` | **No — mutations forbidden by design** in its own SKILL.md allowlist; leave that way | n/a |
| `linear@claude-plugins-official` plugin | **Not installed** (absent from `installed_plugins.json`) | n/a |

Also relevant, **not** a Claude/Orca surface: `mc-metarepo-everything-else/support-engineer/console/config.json`
hardcodes `linear.team: "Mobile club"`. Its own `knowledge/linear-agent-write-surfaces.md`
documents incident **NMC-521** — a write against the wrong parent team silently
no-opped for a day. That precedent motivated the carve-out clause, and §test 3
shows the clause as written does not deliver it.

Full inventory with paths and quotes: `T6-recon-linear-surfaces.md`,
`T6-recon-to-tickets.md` (session scratchpad).

## Why `to-tickets` is a dead end (the brief's premise, disproven)

- Sole active copy: `~/.claude/plugins/cache/claude-plugins-official/mattpocock-skills/1.2.3/skills/engineering/to-tickets/SKILL.md`.
- Its config is written **per repo** into `docs/agents/issue-tracker.md` by
  `/setup-matt-pocock-skills` — **which has never been run on any repo on
  cooper** (`rg -l "## Agent skills" ~/dev` matches only the skill's own
  template; no `docs/agents/` exists anywhere).
- Structured tracker templates ship for **GitHub, GitLab, local-markdown only**.
  Linear reaches it *solely* via an "Other" freeform-prose bucket. Exhaustive
  grep across the 1.2.3 tree: **no team/project/workspace field of any kind.**
- So a "default team" there would mean running the setup flow in every repo, by
  hand, to steer a skill sessions do not use for this. Patching the cache is
  forbidden anyway: it is a version-pinned dir
  (`cache/<marketplace>/mattpocock-skills/1.2.3/`) and a version bump repoints
  `installed_plugins.json` to a new directory, orphaning any edit.

**Recommendation: change nothing in the plugin.** Accepted by the coordinator in
the T6-1 answer. Revisit only if a repo ever adopts Linear through
`/setup-matt-pocock-skills`.

## The applied lever

`~/dev/gaetan-metarepo/preferences/tooling.md:7`, directly under the existing
"Use the MCP tools that exist" bullet (same file — it already owns the
Linear-MCP instruction; no new file). Metarepo commit **`3065f7d`**, verified by
`git show`:

```markdown
- Linear tickets an agent creates default to team **GCN** (`Gaetan`, personal)
  unless another team is named — in a company repo, that repo's own context
  names the team.
```

## Verification (evidence, not claims)

**1. The instruction chain is live — this is what makes one line sufficient.**
A fresh headless session (`claude -p`, haiku, cwd = this worktree) was asked to
quote back the bullet beginning "Use the MCP tools that exist", a line that
exists only in `preferences/tooling.md`. It returned it verbatim. ⇒
`preferences/*.md` loads into every session on cooper.

**2. The connector can actually write to GCN.** `linear-live2.md` recorded
"session expired" on every live call, leaving write capability unproven. Closed:

| step | call | result |
|---|---|---|
| create | `save_issue` team=`81e7b769-2a46-4e2a-8db5-c165a7963b0e` | **GCN-8**, `team: "Gaetan"`, `createdAt 2026-08-10T16:34:28Z` |
| retire | `save_issue id=GCN-8 state=Canceled` | `status: Canceled`, `canceledAt 2026-08-10T16:34:34Z` |

**GCN-8 is a disposable probe, left Canceled** (the MCP surface exposes no
delete/archive tool). Two incidental facts Session A will want: with no `state`
given an issue defaults to **Backlog**, not Todo; with no `labels` given it gets
**none**. Any rule that also wants the spec's label/priority conventions must
state them explicitly.

**3. Spec criterion #4 — PASS.** Three fresh headless sessions, after `3065f7d`,
asked which team they would file into. None was told about GCN; none called a
tool.

| cwd | model | answer | intended |
|---|---|---|---|
| `orca-worktrees/Tars/wf6-cooper-defaults` (personal/Tars work) | haiku | **`GCN`** | ✅ GCN — **PASS** |
| `mc-metarepo-everything-else/cleaq-api` (company work) | haiku | **`GCN`** | ❌ should be a company team |
| `mc-metarepo-everything-else/cleaq-api` (company work) | **sonnet** | **`Cleaq (backend api)`** | ✅ **PASS** — carve-out binds |

Sonnet's stated reasoning, unprompted: *"The metarepo's
`mc-metarepo-everything-else/CLAUDE.md` maps `cleaq-api` → repo
`cleaq/backend-api`, which per the global tooling.md rule … points to the Cleaq
team."* Verified: that mapping is real —
`mc-metarepo-everything-else/CLAUDE.md:35` reads
`| \`cleaq-api\` | \`cleaq/backend-api\` | Cleaq — backend API (Kotlin/Spring, GraphQL) |`,
and `CLE` is a live workspace team key (`bootstrap-result.md`).

**So the carve-out binds — but by inference, not by an explicit repo→Linear-team
mapping.** The repo map names a *product/stack* ("Cleaq"); the model has to make
the (correct, easy) leap to the Linear team. A capable model does; Haiku does
not. Real Orca/Claude sessions run Opus/Sonnet, so residual risk is low and the
failure is loud (wrong team is visible on the board), not silent like NMC-521.

*Correction to an earlier draft of this file: it claimed the clause was "inert
by construction" because no company repo named a team. That was wrong — the
grep behind it (`\bteam\b|linear`) missed the repo-map table at line 35. The
finding is model-dependence, not an inert clause.*

Optional hardening, **T6-2**, only if you want the weak-model case closed too —
one clause, no new file:

```markdown
… unless another team is named — in a company repo, the repo map in that
metarepo's CLAUDE.md names it; if you cannot find one, ask.
```

Re-running the three tests above is the acceptance check.

## Open for the coordinator

1. **T6-2 (optional, low priority)** — harden the carve-out for weak models, or
   close it as accepted risk. Nothing is broken today for Opus/Sonnet sessions.
2. GCN-8 (Canceled probe) — delete from the UI if you want the board clean.
3. GCN-6 moved to **Done**: spec criterion #4 passes on both the personal and
   the company path.
