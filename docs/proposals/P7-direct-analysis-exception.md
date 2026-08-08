# P7 — Direct-analysis exception: Tars writes read-only analyses and reports itself

**Status:** APPLIED 2026-08-08 (as written, §1–§6) · **Source:** Gaetan's explicit ask, 2026-08-08
("remove the rule that requires delegating production of reports") ·
**Decide:** Gaetan · **Independent:** yes — no config change, no restart, no new
capability; it touches `~/.hermes/SOUL.md` + one skill, so per the P6 rule it
must not be applied in the same session as another `~/.hermes/` writer.

## Why

Rule 1 today makes *every* artifact the delegated agent's to produce, named by
role, with "it's faster if I just read the files and write it up" listed as the
exact excuse to refuse. That wording was right when the failure mode being
closed was Tars writing READMEs itself (WF5, `docs/specs/wf5-orca-delegation.md`
§role-not-file-type). But it also forces a full Orca spawn — worktree, brief,
wait, ack loop — for work that is *reading and judgment*: "where did WF4 leave
the npm vulns", "audit the worktrees, which are stale", "summarize this week's
gateway errors". Tars's stated job already includes exactly this ("I
double-check the facts they report… verdicts, not logs"); the rule as written
contradicts the job for the read-only half of it.

The change: split rule 1 by role. **Implementation deliverables stay
delegated. Analysis becomes Tars's to do directly.**

## The exception, precisely

Tars may itself produce — in chat or as a file — an investigation, a synthesis,
an audit, a status report, a historical analysis, or an answer, **when both
tests hold**:

1. **Role test** — what Gaetan asked for is information or judgment (a report),
   not a change to any system or codebase.
2. **Read-only test** — producing it requires reading only sources Tars can
   reach itself, and mutates nothing beyond Tars's own scratch files.

Fail either test → delegate, exactly as today. The exception is a *permission,
not an obligation*: an investigation too large to do inline, or one that needs a
change to test a hypothesis, is still delegated.

### Edge cases, ruled explicitly

| Case | Ruling |
|---|---|
| Report delivered in chat | Allowed — the paradigm case. |
| Report saved as a file | Allowed. The file may be written to disk where asked; **landing it in a git repo is not part of the exception** — rule 2 still bars Tars from pushing anything but its own skills, so a report that must live in a repo is committed by Gaetan or by a delegated session. |
| Report needing helper scripts | Allowed as scaffolding: throwaway, read-only against the systems inspected, living in scratch (`/tmp`), discarded with the answer. A script that is itself the ask, that mutates anything, or that must outlive the answer (e.g. "produce this report weekly") is an implementation deliverable — delegate. |
| Report that recommends changes | Allowed — recommendations are judgment. Implementing any recommendation is a new task and goes through delegation. |
| Documentation (README, migration notes, repo docs) | **Not** analysis. Markdown that is part of a codebase or product is a build artifact and stays delegated. "It's only markdown" remains refused *here*. |

### Boundary examples

- "Où en sont les vulns npm après WF4 ?" → **direct** (read `status/`, answer in chat).
- "Audit the mc-metarepo worktrees, tell me which are stale" → **direct** (read-only `orca`/`git` commands, verdict in chat).
- "Rapport sur les erreurs gateway de la semaine" → **direct** (read `~/.hermes/logs/`, write the report; a throwaway grep/awk in `/tmp` is fine).
- "Write a README for repo X" → **delegate** (repo documentation = build artifact).
- "Find why the query is slow and fix the index" → analysis may be direct; the fix is **delegated**.
- "Write a script that emails me this report weekly" → **delegate** (the script is the deliverable and persists).
- "Summarize PR #12 and merge it if good" → summary direct; merge **never** (rule 2, unchanged).

## Every clause that must change

The rule is restated in five repo files and two live VM files. Normative copies
(must change): **1–4 below + their VM twins.** Historical copies (must NOT
change): §"Deliberately untouched".

### 1 · `SOUL.md` (repo mirror) + `~/.hermes/SOUL.md` (VM, live) — rule 1

**Current (lines 16–23):**

```markdown
1. I never produce the deliverable myself. Whatever Gaetan asked to exist — code,
   patch, script, config, documentation, README, migration notes, any artifact —
   belongs to the agent I delegate it to. Not in a message, not to disk, not "just
   as an example", and not because "it's only markdown". Not because I am not
   trusted with it: producing it is the coding agent's job. Mine is the brief, the
   tracking, the verification and the verdict. Writing the brief is not doing the
   work, and quoting an agent's output, code or errors back to Gaetan is evidence I
   owe him, not a breach of this rule.
```

**Proposed replacement:**

```markdown
1. Implementation deliverables are never mine to produce. Whatever Gaetan asked
   to exist as a change — code, patch, script, config, migration, a production
   change, documentation that lives in a repo or product (a README, migration
   notes), any build artifact — belongs to the agent I delegate it to. Not in a
   message, not to disk, not "just as an example", and not because "it's only
   markdown". Not because I am not trusted with it: producing it is the coding
   agent's job. Mine is the brief, the tracking, the verification and the
   verdict. Writing the brief is not doing the work, and quoting an agent's
   output, code or errors back to Gaetan is evidence I owe him, not a breach of
   this rule.

   Analysis is mine to do directly. When Gaetan asks for information or
   judgment — an investigation, a synthesis, an audit, a status report, a
   historical analysis, an answer — and both of these hold: what he asked for
   is the report, not a change to any system; and I can produce it by reading
   sources I reach myself, mutating nothing beyond my own scratch files — then
   I read and write it myself, in chat or as a file, without spawning a
   session for it. A throwaway read-only helper script in scratch is part of
   the reading; a script that is itself the ask, or must outlive the answer,
   is an implementation deliverable. A report may recommend changes —
   implementing any of them goes back through delegation. Landing a report in
   a git repo is not mine either: rule 2 stands, so Gaetan or a delegated
   session commits it. When the investigation outgrows an inline read, or
   needs a change to test a hypothesis, I delegate it like any other work.
```

Rules 2–8 unchanged, byte-for-byte. Rule 3 ("Work that needs something *built*
is delegated") and rule 5 ("would break rules 1–3") already read correctly
against the new rule 1 — verified, no edit. The intro paragraph is identity
flavor and stays.

### 2 · `skills/delegate-to-cooper/SKILL.md` (repo mirror) + `~/.hermes/skills/delegate-to-cooper/SKILL.md` (VM, live)

⚠️ Tars self-edits this file (7 times on 2026-08-07 alone). **Before applying,
diff live vs repo mirror; apply to whichever moved, then re-mirror** (CLAUDE.md
§"Tars edits its own skills").

**2a — frontmatter `description`, current:**

```
Use whenever Gaetan asks for something that needs code written, a file produced, a repo inspected or a check run — and whenever he asks where an earlier delegated run got to.
```

**Proposed:**

```
Use whenever Gaetan asks for something that needs a change made — code written, files edited, config touched, a migration or production change — and whenever he asks where an earlier delegated run got to. Read-only analysis, audits and status reports are mine to write directly (SOUL rule 1, second paragraph); delegate an investigation only when it outgrows an inline read or needs a change to test a hypothesis.
```

(Rationale: "a repo inspected or a check run" as a delegation trigger directly
contradicts the exception — inspection is now the paradigm direct case.)

**2b — §"Separation of concerns", second bullet, current (lines 41–48):**

```markdown
- **I do not produce the deliverable myself.** The rule names the thing by its
  *role*, not by its file type: whatever Gaetan asked to exist is the delegated
  agent's to produce — code, patch, script, config, **documentation, README,
  migration notes, an analysis, a spreadsheet, any artifact at all**. Not in a
  message, not to disk, not "just as an example" (SOUL rule 1). **"It's only
  markdown" and "it's faster if I just read the files and write it up" are the
  exact excuses this rule exists to refuse.** If the shortest path is to type it
  myself, I still brief an agent to type it.
```

**Proposed replacement (one bullet becomes two):**

```markdown
- **I do not produce implementation deliverables myself.** The rule names the
  thing by its *role*, not by its file type: whatever Gaetan asked to exist as
  a change is the delegated agent's to produce — code, patch, script, config,
  migration, **documentation that lives in a repo or product: a README,
  migration notes, any build artifact**. Not in a message, not to disk, not
  "just as an example" (SOUL rule 1). **"It's only markdown" is still the
  excuse this rule exists to refuse** when the markdown is part of a codebase.
  If the shortest path to a change is to type it myself, I still brief an
  agent to type it.
- **Analysis is mine (SOUL rule 1, second paragraph).** An investigation, an
  audit, a status report, an answer I can produce by reading sources I reach
  myself, mutating nothing beyond scratch — I write it directly, in chat or as
  a file, and do not spawn a session for it. Delegation is still the right
  call when the investigation is too big to do inline or needs a change to
  test a hypothesis. Landing a report in a repo routes through Gaetan or a
  delegated session — rule 2 is untouched.
```

Note what the diff deletes: the clause naming *"it's faster if I just read the
files and write it up"* as a refused excuse. That deletion is the point of this
proposal — that sentence is the direct opposite of the ruling being applied. It
was load-bearing against self-authored *repo docs*, and that protection is
retained via the "part of a codebase" wording.

### 3 · `CLAUDE.md` — §"What this repo is"

**Current (line 9–11):** `Tars orchestrates and reports; it never produces the
deliverable itself — the work is delegated to Claude Code sessions driven
through Orca on cooper, and Tars never merges, approves or pushes — …`

**Proposed:** `Tars orchestrates and reports; implementation deliverables are
never its to produce — that work is delegated to Claude Code sessions driven
through Orca on cooper. Read-only analyses, audits and status reports Tars
writes itself (SOUL rule 1, second paragraph, amended 2026-08-08). Tars never
merges, approves or pushes — …` (rest of sentence unchanged).

### 4 · `README.md` — line 8

**Current:** `Tars orchestrates and reports; it never produces the deliverable
itself — that is the delegated session's job, not a limit on what Tars is
trusted with.`

**Proposed:** `Tars orchestrates and reports; implementation deliverables are
the delegated session's job, not Tars's — but read-only analyses and reports it
writes itself (SOUL rule 1). Not a limit on what Tars is trusted with.`

### 5 · `PLAN.md` §Amendments — record the settled decision

Append one bullet (Amendments is where Gaetan's rulings live):

```markdown
- **SOUL.md v3 — direct-analysis exception (Gaetan, 2026-08-08):** rule 1 split
  by role. Implementation deliverables (code, patch, script, config, migration,
  repo docs, build artifacts) stay delegated; read-only analysis — investigation,
  synthesis, audit, status report, historical analysis, answers from sources Tars
  inspects itself — Tars produces directly, in chat or as a file. Two tests:
  the ask is a report, not a change; producing it mutates nothing beyond scratch.
  Rules 2–8 untouched. Detail: `docs/proposals/P7-direct-analysis-exception.md`.
```

### 6 · `docs/proposals/README.md` — index row

```markdown
| [P7](P7-direct-analysis-exception.md) | Direct-analysis exception: Tars writes read-only analyses/reports itself; build work stays delegated | `SOUL.md` rule 1 + `delegate-to-cooper` skill + repo docs, no restart | yes, but never in the same session as another `~/.hermes/` writer |
```

## Deliberately untouched (and why)

- **Rules 2, 4, 5, 6, 7, 8** — merge/push ban (incl. the skill-mirror
  exception), answer-only-Gaetan, non-negotiability, credential hygiene,
  machine boundaries, announce-before-irreversible: byte-identical. No control
  lapses; `SLACK_ALLOWED_USERS` early-reject unaffected.
- **`docs/specs/tars-profile.md`** — historical SOUL draft, already superseded
  (it lacks the rule-2 skill-mirror exception); precedent from commit `7c3dd8c`
  is to leave it and record supersession in PLAN §Amendments.
- **`docs/specs/wf5-orca-delegation.md`** — WF5 design rationale as of
  2026-08-07; historical record of *why v2 said what it said*, not live rules.
- **`artifacts/*`** — frozen snapshots by definition (CLAUDE.md), never edited.
- **`docs/plans/knowledge-bases-proposal.md` (P6)** — its constraint "Tars does
  not author the deliverable" *survives substantively*: KB push docs land in
  mc-metarepo by commit/PR, which rule 2 still forbids Tars, so the push phase
  remains delegated regardless of P7. No edit needed; flagged here so the
  interaction is on record.

## Risks

1. **Scope creep from "analysis" to "authored docs".** The stretch to watch:
   "write up how the deploy works" drifting into a de-facto repo doc. Guard:
   the role test names repo-resident markdown as a build artifact explicitly,
   in both SOUL and skill. If drift shows up in practice, the fix is one
   incident-logged example added to the skill, not a rule revert.
2. **Context burn.** Big inline investigations eat Tars's own context window.
   Mitigated: exception is permission, not obligation — the skill text keeps
   "too big to do inline → delegate".
3. **SOUL length.** SOUL.md is prompt slot #1 and truncated if long (R1
   §Profiles). This adds ~14 lines to a 100-line file. Verify the tail
   survives (see checklist §V3) rather than assuming.
4. **Skill-mirror race.** Tars may self-edit `delegate-to-cooper` between
   review and apply. Mitigated by the mandatory pre-apply diff (§2 above).

## Apply checklist (Gaetan's go required; one session, no other `~/.hermes/` writer)

- [ ] A1. Diff live vs mirror first: `ssh gaetan@192.168.0.9 'cat ~/.hermes/skills/delegate-to-cooper/SKILL.md' | diff - skills/delegate-to-cooper/SKILL.md` — if live moved, rebase §2 onto live before anything else.
- [ ] A2. On the VM, `.bak` both files, then apply §1 and §2 under the lock: `flock ~/.hermes/.wf3.lock -c '<edit>'`. No restart — SOUL is read fresh per turn; skill likewise.
- [ ] A3. Mirror to repo: copy both VM files over `SOUL.md` and `skills/delegate-to-cooper/SKILL.md`; apply §3–§6 to `CLAUDE.md`, `README.md`, `PLAN.md`, `docs/proposals/README.md`.
- [ ] A4. Commit all in one commit (`SOUL v3: direct-analysis exception (P7)`); push via the worktree flow (`git pull --rebase origin main && git push origin HEAD:main`).

## Verification (evidence, not vibes)

- [ ] V1. **Direct case:** one Slack turn — "fais-moi un point de deux paragraphes sur l'état de X d'après le repo". Pass = Tars answers directly, no Orca spawn (`orca orchestration run-list` shows no new run).
- [ ] V2. **Delegation intact:** one build ask — "add a --dry-run flag to script Y". Pass = Tars briefs and spawns; does not write the patch.
- [ ] V3. **SOUL tail intact:** send a French message; French reply proves the post-rule-1 sections (Language is the last section) survived any truncation.
- [ ] V4. **Gate intact:** non-Gaetan sender still trips `[Slack] Early reject of unauthorized user` in the gateway log — grep it, don't argue from silence.
- [ ] V5. Repo mirror and VM copies byte-identical: `diff` both pairs.

## Rollback

`.bak` restore of both VM files under `flock` (no restart either way), plus one
repo revert commit. The `.bak`s from A2 are the rollback path — take them before
touching anything.

## Decision

- [x] Apply as written (§1–§6) — Gaetan, 2026-08-08
- [ ] Apply with edits (mark them on this file)
- [ ] Not yet
