# Design — the push-back phase (P6/kb, designed not built)

**Nothing here was built.** No branch, no PR, no Orca run, no worktree, no
session. The only commands run for this file are read-only `--help` /
`show` / `gh api` GETs, quoted verbatim where they back a claim.

Two inputs are taken as given (both are peer probe files on disk, cited
inline): `kb-push-inputs.md` (Orca v2 mechanics, SOUL rules, house style) and
`kb-mc-metarepo.md` (the target repo's contribution rules, branch protection).

**Headline finding, and the reason this design is small:** mc-metarepo already
ships the whole push-back procedure as one of its own Claude Code skills.
`.claude/skills/retro/SKILL.md` — "Harvest a finished piece of work into the
metarepo: extract facts + frictions from the sessions, map coverage, ship
approved docs by PR, and propose new skills" — encodes the placement table, the
quality bar, the append-only rule, the PR shape and the conventional-commit
format. `.claude/skills/ship-pr/SKILL.md` line 20 already encodes the
no-self-merge rule for `learnings/` diffs. So the push phase is **not a new
mechanism to build**: it is Tars handing a delegated session one fact and the
instruction "run this repo's own `retro` on it, stop at the PR".

Measured:

```
$ cd /home/gaetan/dev/mc-metarepo && ls .claude/skills/
driving-orca-browser/  handoff-orca/  handoff-orca-opus/  help-ops-request/
helptaker-du-jour/  ht-wrapup/  megaultracode-orca/  notify-tech-pr/  pr-review/
relaunch-vecna-return-order/  retro/  review-diff/  ship-pr/  start-orca/
start-orca-opus/  triage-common/  triage-helptech/  … (22 total)

$ grep -n -i "merge\|learnings" .claude/skills/ship-pr/SKILL.md | head
20:- Diff touches `learnings/` → add `--label agent-learning` at `gh pr create`
   (or `gh pr edit --add-label agent-learning`) and STOP after step 5: learnings
   PRs are always human-merged (see `learnings/README.md`).
```

And the label the contract requires already exists, so no label creation is
needed:

```
$ gh api repos/mobile-club/metarepo/labels --jq '.[].name'
agent-learning
bug
documentation
duplicate
enhancement
good first issue
help wanted
invalid
question
wontfix
```

Prior art for the PR shape (read-only GET, five most recent `agent-learning`
issues/PRs, any state):

```
$ gh api 'repos/mobile-club/metarepo/issues?labels=agent-learning&state=all&per_page=5' \
    --jq '.[] | {number, title, state, user: .user.login}'
{"number":182,"state":"closed","title":"learnings(next/vecna): local environment reality and doc drift","user":"Robin-Koenig"}
{"number":177,"state":"open","title":"docs(knowledge, learnings): night-gang 2026-08-07 harvest — CI zero-step traps + SE run-state facts","user":"GaetanCathelain"}
{"number":162,"state":"closed","title":"docs(knowledge, learnings): an ORDER_ERROR relaunch cannot fix Loop — check the flow's start states first (MC-4164)","user":"NansD"}
{"number":160,"state":"open","title":"docs(learnings): api-v2 jest has a permanent Postgres deadlock rate","user":"NansD"}
{"number":155,"state":"closed","title":"docs(knowledge, learnings): harvest the M1 gang run — Linear team scope, label groups, deploy-gate deadlock, console seams","user":"GaetanCathelain"}
```

**Provisional on P3 + P4.** Every mechanism below assumes P3 (Orca v2 skill)
and P4 (SOUL redesign) land as written; both are unchecked proposals today
(`kb-push-inputs.md` §"Open gaps" #10). If they land with edits, re-read §3
and §4 here.

---

## 1. Trigger — what causes a push

**Recommendation: Gaetan says so, explicitly, in Slack. Nothing else fires a
push in v1.**

Trigger phrase, one of: "remember this", "push that to the metarepo", "that's a
learning", "retro that". Tars treats the message it is replying to (or the text
Gaetan quotes) as the material and starts the delegation in §3.

Why this one, and not the alternatives:

| Candidate trigger | Verdict |
|---|---|
| **Gaetan says "remember this"** | **Chosen.** Zero new machinery, zero heuristics, and the human is already in the loop at the point where value is judged. |
| End of a delegated Orca task ("harvest what that session learned") | **Not v1.** Plausible later, and it is the shape `retro` was written for. But the delegated session's transcript lives on cooper, not on the VM, and Tars would have to decide *whether* the session learned something — a judgement call it has no evidence base for. Revisit once §8's Phase-2 knowledge transfer exists. |
| Periodic harvest (hourly/daily, à la the pull refresh) | **Rejected.** This is the noise generator (see §6). Nothing feeds it: Hindsight memory is off for v1 (no keys), so Tars has no durable record of what it learned since the last harvest. A scheduled harvest with no source of new facts produces filler, on a cadence, into a team repo. |
| Tars' own initiative, on some threshold | **Rejected for v1.** Same missing evidence base, plus an unbounded PR rate against a repo Gaetan does not own alone. |

Note the trigger is also the primary brake — see §6. It is deliberately the
same decision.

**SOUL rule 8 (announce-then-act) applies here**, and the resolution to gap #6
in `kb-push-inputs.md` is: Tars announces **before dispatching the worker**, not
before the worker's `git push`. Tars does not control the worker turn-by-turn
once dispatched, so the only beat it can actually give Gaetan is the one before
`worker-start`. Concretely — one Slack line: "Delegating a learnings PR to
mc-metarepo with this text: `<quote>`. Target `learnings/<stack>/<repo>.md`,
PR only, no merge. Say stop if not." Then proceed; this is not a permission
request. Opening a PR is not irreversible anyway (it can be closed), which is
the second reason the beat sits before dispatch rather than around the push.

---

## 2. The brief — literal text

Passed via `task-create --spec`. `--spec` takes text inline; there is no
`--spec-file`. Measured:

```
$ orca orchestration task-create --help
Usage: orca orchestration task-create --spec <text> [--task-title <text>]
  [--display-name <text>] [--deps <json_array>] [--parent <task_id>]
  [--run <run_id>] [--from <handle>] [--json]
```

Per `kb-push-inputs.md` §1 ("Brief shape it sends") the brief file must be
unique per run, so: write the filled-in text to `/tmp/tars-brief-<slug>.md` on
cooper, then `--spec "$(cat /tmp/tars-brief-<slug>.md)"`. `<slug>` =
`kb-push-<YYYYMMDD-HHMM>`.

Three placeholders Tars fills: `<VERBATIM_QUOTE>`, `<CONTEXT_ONE_LINE>`,
`<SOURCE>`. **Tars fills them by quoting Gaetan, not by composing prose** —
quoting Gaetan's own words back is evidence, not authorship (SOUL rule 1's own
carve-out). Everything else in the block is fixed text, and this fixed text is
part of the P6 proposal, not something Tars writes per run.

```markdown
You are a delegated Claude Code session working in the Mobile Club metarepo
(`mc-metarepo`, github.com/mobile-club/metarepo). Your whole job is to record
ONE piece of knowledge into this repo and open a pull request. Do nothing else.

## The knowledge to record — verbatim, from Gaetan

> <VERBATIM_QUOTE>

Context Gaetan gave around it: <CONTEXT_ONE_LINE>
Where the fact comes from: <SOURCE>

## What to do

1. Run this repo's own `retro` skill (`.claude/skills/retro/SKILL.md`) on the
   single item above. That skill is the authority on placement and quality —
   follow it over anything in this brief if the two disagree, and say so in
   your report. Three deviations, all deliberate:
   - Skip retro's step 1 "Extract". The raw material is the quote above. There
     is no session transcript to mine and you must not go looking for one.
   - Skip retro's step 6 "Skill pass" entirely. No new skills in this task.
   - Skip retro's step 4 AskUserQuestion approval gate. The human gate for this
     task is the pull-request review. Do not block waiting for an answer.
2. Map coverage FIRST (retro step 2). If this fact is already covered by
   `knowledge/*.md`, `learnings/<stack>/<repo>.md` or a pinned stack doc:
   STOP, open no PR, and report `already covered by <path>` with the covering
   quote. A duplicate entry is worse than no entry.
3. Place it per retro's placement table. The default home is
   `learnings/<stack>/<repo>.md`, **append-only** — never rewrite, reorder or
   edit an existing entry. `knowledge/` is allowed only if the fact is
   genuinely cross-cutting (it costs every team session context, so the bar is
   high); if you choose it, justify that choice in the PR body.
4. Allowed paths, hard limit: `learnings/**`. Plus, ONLY if step 3 chose
   `knowledge/`: the new `knowledge/<slug>.md`, its `@import` line in
   `CLAUDE.md`, and its row in `knowledge/README.md`. Touch nothing else.
   Never the pinned docs in `cleaq/ mobileclub/ next/`, never `index/`, never
   `schemas/`, never `scripts/`, never `secrets/`, never `.claude/skills/`,
   never a submodule.

## This repo's own rules — they bind you. Read them before you write.

- `learnings/README.md` "The contract" rule 1, verbatim: entries arrive ONLY
  as pull requests labeled `agent-learning`, opened by an agent against
  `main`. **"A human MUST review and merge every learnings PR; agents MUST
  NEVER self-merge."**
- rule 5: **no PII, no secrets, ever.** Table, column, state and env-var NAMES
  are fine. Values are not — no customer name/email/address, no IMEI or serial
  value, no row data, no credential, no token, no DSN. If the quote above
  contains anything of that shape, STOP, open no PR, and report it.
- rule 7: tech debt gets a Linear issue, not an entry. A defect plus its
  workaround with no ticket is not merged. If this item is that, write no
  entry — report that it needs a Linear issue instead.
- `CONTRIBUTING.md` §Language: everything committed here is in **English**, in
  **ASD-STE100 Simplified Technical English**. CI (`ste-lint`) fails the PR on
  an STE error in an added line.
- `CONTRIBUTING.md` §"Always pass `--repo`": your shell cwd resets between
  commands, back to the metarepo root. Pass `--repo mobile-club/metarepo` on
  EVERY `gh` call and `-C <dir>` on EVERY `git` call, read-only ones included.
  On 2026-08-05 skipping this merged an unrelated PR in this repo.

## PR, not a merge

- Branch off `origin/main`. One commit. Conventional message:
  `docs(learnings): <hook>` (or `docs(knowledge, learnings): <hook>`).
- `gh pr create --repo mobile-club/metarepo --label agent-learning`.
- **STOP THERE.** Do not merge. Do not approve. Do not enable auto-merge. Do
  not push to `main`. If you use `/ship-pr`, it already stops after PR
  creation for a `learnings/` diff — do not override that behaviour.
- Never `git push --force`. Never touch a branch other than your own.

## Done means

Report, in one message, all of:
- the PR URL and number,
- the exact target file path(s) and the number of ADDED lines,
- the output of `gh pr view <n> --repo mobile-club/metarepo --json number,url,state,labels,files`,
- the STE lint result — green with the check output, or the failure text.

If you stopped early (already covered / secret-shaped content / tech debt),
report which of the three and why, with the quote or path that made you stop.
Nothing else counts as done.
```

---

## 3. Mechanics — the exact Orca sequence

Runs on cooper over ssh, from Tars. Every mutating command below is
**NOT EXECUTED** anywhere — the sequence is composed from `--help` text (mine
and `kb-push-inputs.md`'s) and has never been run end to end. Flag markers:
`[VERIFIED]` = I ran that exact command read-only for this file; `[HELP-ONLY]`
= its signature is confirmed from `--help` but it was never invoked;
`[UNVERIFIED — run --help first]` = neither.

```bash
SLUG="kb-push-$(date -u +%Y%m%d-%H%M)"
REPO_ID=8099e312-3232-46f2-83a9-97aeaf5de5a2      # [VERIFIED], see below

# 0. Preflight — read-only, cheap                                [HELP-ONLY]
orca status --json                 # runtime.reachable == true
orca account list --json           # an active claude account

# 1. Brief to a per-run-unique path, then the durable rows
cat > /tmp/tars-brief-$SLUG.md <<'BRIEF'
<the §2 block, placeholders filled>
BRIEF

orca orchestration run-create --objective "Record one learning in mc-metarepo ($SLUG)" --json
#   -> RUN_ID from result.run.id            [UNVERIFIED field name]

orca orchestration task-create --run "$RUN_ID" \
  --spec "$(cat /tmp/tars-brief-$SLUG.md)" \
  --task-title "learnings PR: <hook>" --json
#   -> TASK_ID from result.task.id          [UNVERIFIED field name]
#   --spec/--task-title signature           [VERIFIED via --help]

# 2. Worker in a fresh top-level worktree of mc-metarepo         [HELP-ONLY]
orca orchestration worker-start \
  --run "$RUN_ID" --task "$TASK_ID" \
  --worktree new-top-level \
  --repo "id:$REPO_ID" \
  --name "$SLUG" \
  --agent claude \
  --setup run \
  --json
#   -> DISPATCH_ID from result.dispatch.id  [UNVERIFIED field name]
#   Lands at /home/gaetan/orca/workspaces/mc-metarepo/$SLUG,
#   branch refs/heads/GaetanCathelain/$SLUG off origin/main
#   (measured in kb-push-inputs.md §4, not re-measured here).

# 3. Observe. First wait carries no --ack; every later wait must. [HELP-ONLY]
orca orchestration check --run "$RUN_ID" --wait \
  --types worker_done,escalation,question --timeout-ms 240000 --json
# on a question:
orca orchestration reply --id <msg_id> --body "<text>"
# cheap poll instead of holding ssh open:
orca orchestration worker-show --dispatch "$DISPATCH_ID" --json

# 4. Read the transcript / the worker's report                   [HELP-ONLY]
orca orchestration worker-read --dispatch "$DISPATCH_ID" --limit 200 --json

# 5. Close the loop across Hermes' 300 s turn cap
~/.local/bin/hermes cron create "5m" \
  "Re-attach to Orca run $RUN_ID dispatch $DISPATCH_ID: check, then report the PR URL to Gaetan." \
  --deliver slack --repeat 1

# 6. Release the worker once the PR exists. KEEP the worktree.   [HELP-ONLY]
orca orchestration worker-release --dispatch "$DISPATCH_ID" --json
```

`worker-start`'s flag set is confirmed:

```
$ orca orchestration worker-start --help
Usage: orca orchestration worker-start --task <task_id> [--on <saved-environment>]
  [--worktree <current|selector|new-child|new-top-level>] (--agent <agent> | --terminal <handle>)
  [--model <id>] [--effort <level>] [--name <name>] [--repo <selector>] [--base-branch <ref>]
  [--display-name <text>] [--comment <text>] [--setup <run|skip|inherit>] [--retry-of <dispatch_id>]
  [--timeout-ms <n>] [--run <run_id>] [--from <handle>] [--retry-request <id>] [--json]
…
Notes:
  The call exits 0 only for ready. Failed or outcome_unknown exits 1 and JSON includes
  stage/failedStage, setup, effects, residualResources, and recovery commands when needed.
```

Repo id and path, verified — and note this is **cooper's clone, a different
checkout from the VM's read-only `~/dev/mc-metarepo` search clone**:

```
$ orca repo show --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json
  "id": "8099e312-3232-46f2-83a9-97aeaf5de5a2",
  "path": "/home/gaetan/dev/mc-metarepo",
  "displayName": "mc-metarepo",
  "gitUsername": "GaetanCathelain",
  "gitRemoteIdentity": { "canonicalKey": "github.com/mobile-club/metarepo",
                         "remoteUrl": "https://github.com/mobile-club/metarepo.git" },
  "hookSettings": { "scripts": { "setup": "", "archive": "" } }
```

**Rule this design states explicitly, because no existing doc does** (gap #4 in
`kb-push-inputs.md`): *a push NEVER goes through the VM's `~/dev/mc-metarepo`
clone.* That clone is read-only, search-and-refresh only. Every write to
mc-metarepo happens in an Orca worktree on cooper, created by a delegated
worker. If Tars ever finds itself about to run `git commit` on the VM, it is
wrong twice — wrong clone and SOUL rule 1.

**Recovery rule, inherited and load-bearing:** `worker-start` can outrun the
300 s cap *after* creating the dispatch. Do not assume nothing was created and
do not re-run it — `orca orchestration worker-list --run "$RUN_ID" --json`
first and adopt any dispatch found (`kb-push-inputs.md` §1, "Known ceilings").
Re-running it here would mean two workers writing two PRs for one fact.

**Orca CLI, not Kanban** (gap #3). A push is a one-shot ad-hoc job reported to
Gaetan in the Slack thread where he asked for it; there is no board state worth
maintaining and no queue to drain. If the trigger ever becomes automatic
(§1 alternatives), revisit — a recurring harvest is exactly what a board is for.

---

## 4. PR vs branch vs direct-to-main

**PR, labeled `agent-learning`, merged by a human. This is not a preference —
it is the repo's stated contract and it is also enforced by GitHub.**

The argument, from the repo's own rules (all quoted in `kb-mc-metarepo.md`):

1. `learnings/README.md` rule 1: *"Agent-written, human-merged — always.
   Entries arrive ONLY as pull requests labeled `agent-learning`, opened by an
   agent against `main`. A human MUST review and merge every learnings PR;
   agents MUST NEVER self-merge."* This design's target directory is exactly
   the one that rule governs. There is no reading of it under which
   direct-to-main is permitted.
2. `CLAUDE.md` write-rule table, `learnings/` row: *"append via PR labeled
   `agent-learning`, **human-merged only**"*.
3. Branch protection makes it moot anyway: `main` requires ≥1 approving review,
   `enforce_admins: true` (admins are not exempt), force-push and deletion
   blocked. A direct push from a worker would be rejected by GitHub, not just
   by convention.

**Who reviews and merges: a Mobile Club human — Gaetan or a teammate — through
the normal GitHub review flow.** Not Tars, and not the delegated session. This
is doubly bound: the repo's rule 1, and the standing project rule that Tars
never merges, approves or pushes. A PR *authored by a session Tars delegated
to* is fine; Tars authoring or merging one is not.

Direct-to-main is therefore **not an open decision** — it is closed by the
target repo. The only thing that would reopen it is Gaetan changing
mc-metarepo's own contract, which is a team decision, not a P6 one.

One flag carried forward from `kb-mc-metarepo.md`, relevant because it is the
one place in this repo where an automated PR merges itself:
`.github/workflows/auto-update-submodules.yml` comments *"0 reviews / 0 status
checks required"*, but live protection says
`required_approving_review_count: 1`. Either the rule was tightened after that
workflow was written, or that bot's merges have been failing. Not investigated
further. It does not change this design — it confirms it: **an automated PR
cannot self-merge in this repo today.**

---

## 5. How Tars verifies, without doing the work

Reading a PR and a diff is verification, which is Tars' job (SOUL rule 2's own
carve-out). Writing or fixing the content is not.

What Tars checks, after the worker reports done:

```bash
gh pr view <n> --repo mobile-club/metarepo --json number,url,state,labels,files,additions,deletions
gh pr diff <n> --repo mobile-club/metarepo --name-only
gh pr checks <n> --repo mobile-club/metarepo        # ste-lint
```

Four assertions, each a hard pass/fail:

| Check | Pass condition |
|---|---|
| PR exists and is open | `state: OPEN`, URL resolves |
| Label | `labels` contains `agent-learning` |
| Path scope | every entry of `--name-only` is under `learnings/`, or is one of the three `knowledge/` surfaces the brief allows. Anything else = fail. |
| Not merged | `state != MERGED`. If it merged, that is an incident, not a success — report it as such. |

What Tars reports to Gaetan, in one Slack message: the PR URL, the target
path(s), added-line count, the CI verdict, and the run/dispatch ids. Per the
standing reporting duty, the exact commands verbatim, not "I delegated it".

If the session failed or produced junk:

- **Failed / `outcome_unknown`** — report the failure, the `failedStage` from
  `worker-start`'s JSON, and the dispatch id. Do not retry silently; a retry is
  a new decision Gaetan makes.
- **Stopped early on purpose** (already covered / secret-shaped / tech debt) —
  that is a **successful** outcome. Report which of the three, with the
  worker's quote. No PR is the right answer in all three cases.
- **PR opened but out of scope** (touched a forbidden path) — Tars does not fix
  it. It reports the offending paths and asks Gaetan whether to close the PR or
  hand it back to a new session with a corrective brief. Tars closing the PR
  itself is on the wrong side of rule 2; ask.
- **Content is junk but in-scope** — Tars is not the quality gate. Say so
  plainly ("PR open, content looks thin: `<quote>`") and let the human reviewer
  decide. Do not rewrite it.

---

## 6. The noise problem

The failure this design most has to prevent is not a bad PR. It is **thirty
mediocre PRs**, each individually defensible, landing in a repo four other
people read, until `learnings/` is padding and nobody trusts it. The repo's own
`retro` skill says the quiet part directly in its placement table — "Already
derivable from code, git history, or CLAUDE.md → **skip**" and
"Session-specific narrative → **skip**".

**The single strongest brake: the trigger is human-only. Nothing pushes unless
Gaetan says "remember this".** That is why §1 rejects the periodic harvest even
though it is the more impressive design. A push rate that cannot exceed the
rate at which a human deliberately asks for one cannot leak. Every other brake
below is a second line for the day the trigger is loosened.

Ranked, and all cheap:

1. **Human trigger** (§1) — chosen, primary. Bounds the rate at the source.
2. **Human merge** (§4) — the repo's own gate, already enforced by GitHub, and
   the only one that catches *wrong* knowledge rather than merely excessive
   knowledge. Costs this design nothing: it exists.
3. **Scoped directory** — `learnings/**` by default, `knowledge/` only with a
   written justification in the PR body. `knowledge/` is auto-imported into
   `CLAUDE.md` for the whole team, so a note there costs every teammate's
   context on every session. That is the surface where noise actually hurts;
   `learnings/` is opt-in reading and append-only, so it tolerates volume.
4. **Coverage check before writing** (brief step 2) — the duplicate/contradiction
   brake. "Already covered → STOP, open no PR" makes *not writing* an explicit
   success state, which is the only way an agent ever chooses it.
5. **One open Tars-originated learnings PR at a time.** Before dispatching,
   Tars runs `gh pr list --repo mobile-club/metarepo --label agent-learning
   --state open --json number,title` and, if one from a previous push is still
   open, it says so and does not dispatch. Backpressure is the reviewer's
   queue. One line of preflight, no state to keep. *(Command not run for this
   file — [UNVERIFIED], but `gh pr list --json` is standard.)*
6. **Append-only single file per repo** — already true by construction:
   `learnings/<stack>/<repo>.md` is one file per repo, appended, not a new file
   per note. No design work needed; just do not let the worker create new
   `learnings/` files.

**Rejected as brakes:** a rate limit in wall-clock terms (a limit nobody hits
is theatre, and one that bites blocks a genuinely valuable note); a numeric
"is this worth a file" score (unfalsifiable, and `retro`'s placement table is
already the qualitative version of it); a staging area that batches notes
(state to maintain, and batching is what the human trigger already does).

---

## 7. What could go wrong

| Failure | Mitigation |
|---|---|
| **Session writes to the wrong repo.** `CONTRIBUTING.md` records this happening for real on 2026-08-05 — cwd resets between commands, a bare `gh pr merge 135` resolved against the metarepo and merged an unrelated PR, one touching `learnings/`. | The brief quotes that rule and mandates `--repo mobile-club/metarepo` on every `gh` call and `-C <dir>` on every `git` call. Tars' §5 path-scope check catches it after the fact. |
| **A secret or PII lands in a private-but-team repo.** The material is a quote from a Slack conversation; Slack conversations contain tokens and customer identifiers. | Two layers. Brief: quotes `learnings/` rule 5 and makes "STOP, open no PR, report it" the required action on any credential-shaped or customer-shaped value. Human PR review is the backstop. Tars itself never handles the secret — SOUL rule 6 means it does not paste one into a brief in the first place. Residual risk is real: Tars quotes Gaetan verbatim, so if Gaetan's own message contains a value, it reaches the brief. **Tars must refuse to dispatch and say why** if the quote contains something credential-shaped. |
| **Duplicate or contradictory notes** accumulating over months. | Brief step 2's coverage map runs before any write, and "already covered → STOP" is an explicit success state. Append-only placement means an old entry is never silently overwritten by a newer, wronger one — a contradiction shows up as two dated entries a reviewer can see. |
| **PR spam.** | §6 brakes 1 and 5: human trigger, and at most one open Tars-originated learnings PR. |
| **Tars misreports success** — says "PR opened" when the worker died, or "merged" reading a stale state. | §5's four assertions are `gh` calls against GitHub, not the worker's own claim. The worker's report is a hint; `gh pr view --json` is the evidence. Reporting duty requires the verbatim command and the ids, so a bare "done" is not an acceptable answer. |
| **`worker-start` outruns the 300 s cap after creating the dispatch; Tars retries; two workers write two PRs for one fact.** | Never re-run `worker-start`. `worker-list --run "$RUN_ID"` first, adopt any dispatch found (§3). |
| **The worker self-merges.** | `ship-pr` already stops at PR creation for a `learnings/` diff; the brief forbids merge/approve/auto-merge explicitly; branch protection requires an approving review with `enforce_admins: true`. Three layers, two of them outside Tars' control. Tars' §5 check treats `state: MERGED` as an incident. |
| **Worktrees accumulate.** mc-metarepo already carries 9, several stale, and v2's default is keep. | v1's push rate is bounded by a human trigger, so this grows slowly. `worker-release` yes; `worktree rm` only on Gaetan's word. Flagged as an open decision (P6-e) rather than solved, because pruning is not this design's call to make. |
| **The delegated session asks a question and blocks forever.** | `check --wait --types worker_done,escalation,question` surfaces it; `reply --id <msg_id>` answers it. Tars answers questions about *scope and intent* — it does not answer by supplying content, which would be authoring the deliverable through a side channel. |
| **STE lint fails the PR.** | Not a failure mode of this design — it is the repo's own quality gate doing its job. The brief requires reporting the lint result either way; a red lint is reported, not hidden, and fixing it is the worker's job in the same session or a follow-up delegation. |

---

## 8. Explicitly out of scope for v1

- **Automatic triggers of any kind** — end-of-task harvest, periodic harvest,
  Tars' own initiative. Reason: no evidence base. Hindsight memory is off (no
  keys), so Tars has no durable record of what it learned; a scheduled harvest
  with no source produces filler. Revisit when memory exists.
- **Writing skills back** (`retro` step 6). One PR per skill, a much higher bar,
  and a skill is code-shaped. Explicitly skipped in the brief.
- **Any repo except mc-metarepo.** `gaetan-metarepo` is deliberately not
  Orca-registered (P3, "Scope settled by your ruling") and stays out.
  `orca repo add --path` has never been run or measured by any probe.
- **Editing existing entries, or `knowledge/` as the default target.** Both
  raise the blast radius of a wrong note. Append-only to `learnings/` is the
  reversible shape.
- **Tars closing, editing or merging a PR** it caused to exist. It reports and
  asks. Rule 2.
- **Kanban-card-driven push** (`kb-push-inputs.md` gap #3). Only earns its
  keep once the trigger is recurring, which v1 does not have.
- **Push failure/conflict/retry automation.** If mc-metarepo diverged and the
  PR conflicts, the human reviewer sees a conflicted PR and says so. Automating
  conflict resolution for a docs append — where `CONTRIBUTING.md` already says
  the resolution is always "pure union, keep both sides" — is machinery for a
  problem that has not happened yet.

---

## Open decisions for Gaetan

| # | Question | Recommendation |
|---|---|---|
| P6-a | Trigger: human-only, or also end-of-delegated-task? | **Human-only for v1.** §1. |
| P6-b | Does SOUL rule 8's announce-then-act beat sit before `worker-start`, or is a learnings PR reversible enough to skip the beat? | **Announce before dispatch**, one Slack line, no wait. §1. |
| P6-c | May Tars target `knowledge/` at all, or is `learnings/**` the hard ceiling in v1? | **Allow `knowledge/` with written justification** — the repo's own placement table already sets the bar, and forbidding it outright would push cross-cutting facts into the wrong home. |
| P6-d | The "one open Tars-originated learnings PR at a time" brake: on or off? | **On.** One preflight `gh pr list`, no state, and it is the brake that matters if P6-a ever loosens. |
| P6-e | Worktree cleanup for push runs: keep (v2 default), or auto-`worktree rm` after the PR is merged? | **Keep in v1** (low volume, and the worktree is evidence). Revisit with P6-a. |
| P6-f | If the worker opens an out-of-scope PR, may Tars close it, or must it ask? | **Ask.** Closing is a write against a team repo; rule 2. |

---

## Sources

- `status/probes/wf5/kb-push-inputs.md` — Orca v2 mechanics, SOUL rules 1/2/3/6/7/8 verbatim, house style, the 10 open gaps.
- `status/probes/wf5/kb-mc-metarepo.md` — `learnings/README.md` contract, `CLAUDE.md` write-rule table, `CONTRIBUTING.md` language + `--repo` rules, live branch protection.
- Measured for this file, read-only, on cooper: `orca orchestration task-create --help`, `orca orchestration worker-start --help`, `orca repo show --repo id:8099e312-… --json`, `gh api repos/mobile-club/metarepo/labels`, `gh api 'repos/mobile-club/metarepo/issues?labels=agent-learning&state=all&per_page=5'`, `ls /home/gaetan/dev/mc-metarepo/.claude/skills/`, `head -80 .claude/skills/retro/SKILL.md`, `grep -n -i "merge\|learnings" .claude/skills/ship-pr/SKILL.md`.
- **Not tested:** the entire `run-create → task-create → worker-start → check → worker-read → worker-release` sequence (no Orca run, no worker, no worktree was created); every JSON response field name (`result.run.id`, `result.task.id`, `result.dispatch.id`); `gh pr list --label agent-learning --state open --json`; whether `retro` behaves correctly with its Extract and AskUserQuestion steps skipped (it has never been run in that mode); whether an Orca worker's `gh` calls authenticate as `GaetanCathelain` from a fresh worktree.
