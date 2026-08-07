# Inputs — push-phase design + P-series house style (P6/kb)

Local-only extraction, no ssh. Read in full: `docs/proposals/{P1,P2,P3,P4,P5,README}.md`,
`docs/specs/{wf5-orca-delegation,wf5-kanban}.md`,
`status/probes/wf5/{orca-v2-capabilities,orca-v2-current-skill}.md`,
`artifacts/SOUL-proposed.md` (read in addition — P4.md describes rule 1 but
does not quote it verbatim; the proposed SOUL replacement does).

---

## 1. Orca v2 mechanics (exact commands)

Source: `docs/specs/wf5-orca-delegation.md` §"WF5 — Orca delegation v2
(authoritative)" (the only non-superseded half of that file — everything
above its `---` is v1, dropped, kept for the record only).

Transport: **ssh + the `orca` CLI**, talking to the Orca desktop app already
running on cooper (`orca` is on the non-interactive ssh PATH at
`/usr/local/bin/orca`; `~/.local/bin/orca` also works, it's a wrapper on the
same binary). No sandbox, no wrapper script, no deny-rule settings file — v1's
entire guardrail machinery is dropped by Gaetan's ruling (separation of
concerns, not confinement).

Full sequence, in order (composed from the spec's "Mechanism" section +
`orca-v2-capabilities.md` §9 — **§9's sequence is explicitly marked NOT
EXECUTED**, composed from `--help` text and schema evidence, field names like
`result.dispatch.id` are unconfirmed):

```bash
# 0. Preflight — cheap, read-only
orca status --json                    # must show runtime.reachable == true
orca account list --json              # must show an active/usable claude or codex account

# 1. Durable namespace + task row (no terminal binding, no shared shell state)
orca orchestration run-create --objective "<one-line objective>" --json
#   -> capture result.run.id as RUN_ID
orca orchestration task-create --run "$RUN_ID" \
  --spec "<full task brief, prose>" \
  --task-title "<short title>" --json
#   -> capture result.task.id as TASK_ID

# 2. Spawn the supervised worker directly in the target repo's worktree
orca orchestration worker-start \
  --run "$RUN_ID" --task "$TASK_ID" \
  --worktree new-top-level \
  --repo id:<repoId> \
  --name <slug> \
  --agent claude \
  --setup run \
  --json
#   -> capture result.dispatch.id (field name NOT confirmed) as DISPATCH_ID

# 3a. Block for completion in ONE ssh call (preferred)
orca orchestration check --run "$RUN_ID" --wait \
  --types worker_done,escalation,question --timeout-ms 240000 --json
#   First wait omits --ack. Every wait AFTER a delivery carries that
#   delivery's id: `check --ack <delivery_id> --run "$RUN_ID" --wait ...`
#   (an un-acked re-attach replays the same already-processed batch forever
#   — --ack is mandatory, not hygiene, per `check --help` verbatim).

# 3b. Or poll cheaply instead of holding the ssh connection open
orca orchestration worker-show --dispatch "$DISPATCH_ID" --json
orca orchestration task-list --run "$RUN_ID" --status dispatched --json

# 3c. Answering a worker's mid-task question
orca orchestration reply --id <msg_id> --body "<text>"
#   NOT `send --to dispatch:<id>` — that mails the dispatch's *next* check,
#   which a blocked worker never issues; it errors on nothing.

# 4. Retrieve output (works even after release — transcript archived first)
orca orchestration worker-read --dispatch "$DISPATCH_ID" --limit 200 --json

# 5. Cleanup — post-completion only, keep-by-default (see §4 below)
orca orchestration worker-release --dispatch "$DISPATCH_ID" --json
orca worktree rm --worktree "id:<repoId>::<worktree-path>" --force --json   # only on Gaetan's word
```

Checkpoint / re-attach across Hermes' 300 s turn cap (the whole point of v2
over v1): `run_id`/`task_id`/`dispatch_id` are durable DB rows
(`~/.config/orca/orchestration.db`), so a lost turn is a checkpoint, not a
lost job. Because nothing pushes a later turn to a Slack agent on its own,
"I will re-attach" requires actually scheduling one:

```bash
~/.local/bin/hermes cron create "5m" "<re-attach instruction>" --deliver slack --repeat 1
```

(bare platform name ⇒ `SLACK_HOME_CHANNEL`; `/bg` always answers the invoking
surface and can never reach home — `docs/facts.md`).

### Brief shape it sends

- The brief is passed via `task-create --spec "<full task brief, prose>"`
  (not stdin heredoc — that was v1's `delegate.sh` shape, dropped).
- No structured schema is specced — "prose" is the whole contract, same as
  v1: state the goal, the constraints, and what "done" looks like. The v1
  `SKILL.md` (superseded, quoted in full in `orca-v2-current-skill.md` §1)
  added "self-contained — the agent has no memory of my thread with Gaetan
  and cannot ask me a question" for its synchronous `delegate.sh` shape; v2's
  worker CAN ask a question via `orchestration ask`/`reply`, so that
  constraint is relaxed but not explicitly re-stated anywhere in v2.
- P3 (`docs/proposals/P3-orca-v2-skill.md`) adds: "unique brief path
  `/tmp/tars-brief-<slug>.md`" — a leftover reference to a file-based brief
  that doesn't fully reconcile with `--spec` taking the brief inline; the
  spec's "Known ceilings" section separately states **"The brief file path
  must be unique per run"** and ties both the brief file and the worktree
  name to one `<slug>`. Read together: a brief file may still be written to
  disk and then `--spec "$(cat …)"`'d in, per-run-unique naming required
  either way.

### How Tars observes the session

- **Blocking wait** (`check --run --wait --types worker_done,...`) — single
  ssh call, returns the instant a matching message lands, or `{count:0}` at
  timeout. Stderr keepalive JSON every 15 s distinguishes waiting from
  hanging.
- **Cheap poll** (`worker-show --dispatch`) — point-in-time state:
  `starting|ready|start_unknown|failed|succeeded|stopping|stop_unknown|stopped|abandoned`.
- **Output retrieval** (`worker-read --dispatch --limit`) — works after
  release, archived transcript preserved first.
- No push notification reaches Tars on its own — Orca's internal
  agent-status push (`~/.orca/agent-hooks/claude-hook.sh` → HTTP POST) is
  **loopback-only** (`127.0.0.1:${ORCA_AGENT_HOOK_PORT}`), never leaves
  cooper. Tars always has to ask (`orca-v2-capabilities.md` §3.4, confirmed:
  grepped 223-command `agent-context.json` for
  `webhook|callback|notify|push|hook_url` — zero real hits).

### What it can and cannot do

**Can:** any path, any command, `sudo` included, on cooper, by design (SOUL
rule 7, once P4 lands) — Tars acts with Gaetan's own access. Spawn workers in
any **Orca-registered** repo (`--repo id:<id>`/`name:`/`path:`), register a
new repo itself if the work needs one (`orca repo add --path`, mentioned in
the spec, **never tested in any probe read for this task**). Reply to a
blocked worker's question. Read a PR/diff/CI log to verify what it delegated
(SOUL rule 2, once P4 lands).

**Cannot:** produce the deliverable itself (SOUL rule 1 — code, patch,
*or documentation*); merge, approve or push (SOUL rule 2); use
`ListAgents`/`SendMessage` — that channel exists only between two live agent
turns, and Tars over ssh is a stateless caller, not itself a live Claude Code
turn (`orca-v2-capabilities.md` §7); rely on any built-in outbound
push/webhook from Orca (none exists); trust the local-loopback web API at
`0.0.0.0:6768` as a substitute for the CLI transport (not exercised, flagged
only as a "worth a follow-up connectivity check", not adopted).

### Every stated limitation (from the spec's "Known ceilings")

- **300 s command cap** — the blocking wait is bounded at
  `--timeout-ms 240000`; real coding work runs longer. Survivable only
  because ids are durable — a timeout is a checkpoint, not a loss.
- **`worker-start` is a mutation with no guaranteed return** — can outrun the
  cap after creating the dispatch. Recovery rule: do NOT assume nothing was
  created and do NOT re-run; `worker-list --run <RUN_ID> --json` first, adopt
  any dispatch found.
- **No outbound push** — see above; Tars always has to ask, and closing the
  "tell me when it's done" loop is Hermes' `cron create` side, not Orca's.
- **Orca must be running** — hard dependency, no plain-shell fallback;
  `orca status --json` first, always.
- **Unverified response field names** — `result.run.id`/`result.task.id`/
  `result.dispatch.id`, and the delivery-id field name on a non-empty `check`
  batch, were never observed live (only the empty-batch shape
  `{messages, count, acknowledged, runId}` is confirmed). First live run must
  log and correct this.
- **Brief file path must be unique per run** — a fixed path is silently
  clobbered by a concurrent delegation or retry.
- **Worktrees accumulate** — nothing self-cleans; mc-metarepo already carries
  9, several stale. v2 default is *keep* (it's the work product) and report
  the path. Prune only on Gaetan's word.
- **Only `claude` and `codex`** are verified-usable `--agent` values on
  cooper today (binaries on PATH + `orca account list` cross-checked;
  `gemini`/`cursor`/etc. show hook plumbing but are not installed/runnable).

---

## 2. SOUL constraints (verbatim rules 1 and 8)

Source: `artifacts/SOUL-proposed.md` (the proposed replacement text; P4.md
paraphrases rule 1, this is the actual proposed wording). **Status: proposed,
NOT applied** — P4 is a review-shaped proposal, not yet decided by Gaetan.

> 1. I never produce the deliverable myself. Whatever Gaetan asked to exist —
>    code, patch, script, config, documentation, README, migration notes, any
>    artifact — belongs to the agent I delegate it to. Not in a message, not
>    to disk, not "just as an example", and not because "it's only markdown".
>    Not because I am not trusted with it: producing it is the coding agent's
>    job. Mine is the brief, the tracking, the verification and the verdict.
>    Writing the brief is not doing the work, and quoting an agent's output,
>    code or errors back to Gaetan is evidence I owe him, not a breach of
>    this rule.

> 8. Before an action that cannot be undone, I say what I am about to do and
>    leave Gaetan the beat to stop it. This is not asking permission — on
>    cooper I act with his access — it is not surprising him with something
>    he cannot reverse.

Both are directly load-bearing for a knowledge-push design: rule 1 means
Tars can never write the pushed knowledge content itself — even a
"just append this one learning" edit to mc-metarepo must go through a
delegated agent, no exception for markdown. Rule 8 means an irreversible
step in the push flow (plausibly: opening a PR, or any git write against
mc-metarepo) needs an announce-then-act beat — a statement of intent with a
window to stop it, not a permission request.

### Other rules bearing on writing to repos, approvals, delegation

> 2. I never merge, approve or push. Reading a pull request, a diff or a CI
>    log is how I verify what I delegated — that is my job, not a breach of
>    it.
> 3. I orchestrate and I report. Work that needs something built is
>    delegated to an agent or handed back to Gaetan as a decision.
> 6. I never read, print, echo or pass along a credential, token or key —
>    not in a message, not into a file, not on a command line. If work needs
>    a secret, I name which one and let Gaetan or the agent that owns it
>    supply it.
> 7. Cooper is mine to act on with Gaetan's own access, sudo included —
>    standing in for him is the point. I do not ask permission to run what
>    the job needs; refusing to act is as much a failure as doing the
>    implementation myself. Other machines are not mine: p-Hermes
>    (192.168.0.8) is read-only to me, and the pve hypervisor (192.168.0.3,
>    a different machine) is not mine to touch at all.

P4.md itself adds framing not in the SOUL text: **"Tars never creates a
PR"** is explicitly reworded to **"Tars does not do the work itself"** — a
PR resulting from a Tars-initiated delegated task is fine; Tars authoring
one is not (P3.md, "Scope settled by your ruling"). The `approvals:` mode
(B1 in P4) is the actual code-level sudo gate — separate from any prose rule
— and is **PARKED**, Gaetan's call, not decided. Under `off`, the
"dangerous patterns" tier (`git reset --hard`, writes to `.env`/credential
files, etc.) is released for the *spawned Claude Code session*, not Tars
itself — Tars' own command surface stays ssh/orca/hermes only.

**Not yet decided (both P3 and P4 have empty checkboxes):** neither proposal
has been applied. Any push-phase design built on top of this must say so —
it presumes P3+P4 land, and per `docs/proposals/README.md`, P3+P4 "must land
together" and widen *what a wake can do*, while P1 (not read for this task,
noted only from README) widens *what wakes* Tars; the two together are a
recommended-not-same-day combination, stated as a risk to carry forward, not
a blocker to this design task.

---

## 3. Delegation brief shape as specced

Covered inline in §1 above (brief = prose via `task-create --spec`, per-run
unique naming). Restating the parts specific to "as specced" rather than "as
measured":

- **Repo scope, as decided by Gaetan** (P3.md, "Scope settled by your
  ruling"): mc-metarepo is the **default** repo; all Orca-registered repos
  are fair game; Tars may register more. **gaetan-metarepo is deliberately
  NOT registered** — i.e. explicitly out of scope for any Tars-initiated
  delegation, push-phase included.
- **Layout the skill teaches Tars**, per P3: Orca's model of
  projects → worktrees (with parent/child) → tabs (browser / Claude Code /
  etc.) — the push-phase brief would need to reason in these terms
  (which worktree, which repo id) exactly like any other delegated task.
- **Reporting duty**, unchanged from v1 and restated for v2 (spec, "What v1
  got right and v2 keeps"): every reply states the verdict, the exact
  command verbatim, and the run/dispatch ids. "I delegated it" is not an
  acceptable summary.
- **Knowledge/preferences transfer is explicitly a separate, later,
  design-only phase** — both P3 ("Knowledge/preferences transfer is a later
  phase, to be planned not built") and the spec's own "Phase 2 — not built
  yet" section say this in near-identical words: giving Tars Gaetan's
  knowledge/preferences (so its briefs "sound like Gaetan's") is deferred,
  and nothing in v2 depends on it. **This is precisely the P6 push-phase's
  neighborhood** — v2 makes Tars able to *drive* Orca correctly; it does not
  make Tars' briefs read like a human author's, which matters if a pushed
  learnings-file entry is expected to read naturally.

---

## 4. Measured Orca capabilities

Source: `status/probes/wf5/orca-v2-capabilities.md` — read-only recon on
cooper, **every command in the report body was actually run**; §9's spawn
sequence is the sole explicit exception (composed, not executed).

**Creating a worktree for an arbitrary repo:**
- Repo binding is by **Orca-registered repo only** — `--repo id:<id>` /
  `name:<name>` / `path:<path>` all resolve to the same repo record
  (measured: `orca repo show --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json`).
- **A repo not yet registered with Orca cannot be targeted directly** —
  `orca repo add --path <path>` is the documented mechanism (cited in the
  spec) but **was not run or verified in any file read for this task**. If
  the push-phase target is mc-metarepo, this is moot (already registered,
  `id:8099e312-3232-46f2-83a9-97aeaf5de5a2`,
  `github.com/mobile-club/metarepo`, `/home/gaetan/dev/mc-metarepo` on
  cooper — a **separate clone from the Tars-VM read-only clone** used for
  search/pull).
- Landing path is governed by the **global** `orca-data.json → settings`
  (`workspaceDir=/home/gaetan/orca/workspaces`, `nestWorkspaces=true`,
  `branchPrefix=git-username`), not any per-repo override — mc-metarepo's
  own `worktreeBasePath` is the literal string `"null"`, confirmed inert by
  cross-checking real `worktree list` entries. **Measured conclusion:** a
  worktree created for mc-metarepo today lands at
  `/home/gaetan/orca/workspaces/mc-metarepo/<name-slug>`, branch
  `refs/heads/GaetanCathelain/<name-slug>` off `origin/main` unless
  `--base-branch` is passed.
- **Nothing auto-starts an agent.** `--agent claude` (or another id) must be
  passed explicitly; mc-metarepo's `hookSettings` has empty setup/archive
  scripts, so setup "runs" as a no-op. Measured: `claude` and `codex` are
  the only two verified-usable `--agent` values (binaries on PATH +
  `orca account list --json` cross-check; other providers show hook
  plumbing only, not installed).

**Starting a session with a prompt / getting a result back:**
- The full mutating sequence (`run-create` → `task-create` → `worker-start`
  → `check --wait` → `worker-read` → `worker-release`) is **composed from
  `--help` text and schema evidence, explicitly marked NOT EXECUTED** — no
  worktree, terminal, or orchestration run was created in this recon, per
  Gaetan's explicit read-only instruction. Field names in the JSON responses
  (`result.dispatch.id` etc.) are unconfirmed.
- What WAS measured and is load-bearing: `--run <run_id>` resolves
  **standalone**, with **no prior `run-use` bind**, from "a plain non-Orca
  shell with zero shared state" — i.e. exactly Tars's structural position as
  a stateless ssh caller. `orca orchestration run-create --objective … --json`
  produced `run_24d000a3a6b6`; a **fresh separate process** then ran
  `orca orchestration check --run run_24d000a3a6b6 --peek --json` and got
  `ok: true` with `runId` echoed, not `run_not_found`. This settles the one
  objection ("durable coordinator state across ssh calls that share no
  session") that would have been fatal to the whole design if true.
- `worker-start`/`worker-show`/`worker-read` durability by `dispatch_id`
  (survives terminal churn) is confirmed by direct reading of
  `worker-release --help` ("an inspectable output archive is preserved
  before the terminal closes") and the `worker_dispatches` SQLite schema —
  not by an actual live worker run.
- **Peer-session addressing** (`ListAgents`/`SendMessage`) is a live-Claude-
  Code-turn-only mechanism, structurally unavailable to Tars (a stateless
  ssh caller). Confirmed via `claude --help`'s top-level command list having
  no send/message subcommand — the only CLI-level lever on another session
  is `-r/--resume <sessionId>`, interactive by default, untested for `-p`
  non-interactive resume.

**Cleanup / accumulation risk, evidenced not asserted:** mc-metarepo
carries 9 worktrees today (`orca worktree list --json`, several visibly
stale, e.g. a `"comment": "handed off to handoff-fable — ..."` days-old
entry). Nothing in `worktreeMeta` carries a TTL. A recurring push-phase
that keeps every worktree by default (v2's stated default) would add to
this pile every run.

---

## P-series house style skeleton

Derived from all five proposals (`P1`–`P5`) plus `docs/proposals/README.md`
as the index pattern. Section order below is the pattern to match for a P6
file; length across P1–P5 ranges ~65–125 lines, dense, no filler.

1. **Title** — `# P<N> — <short name>: <one-line framing of the change or
   decision>`. The framing after the colon states the *shape* of the ask,
   e.g. "transport PARKED", "reduces to one go/no-go", "separation of
   concerns, not confinement".

2. **Metadata line** (single line, `·`-separated, right under the title):
   `**Status:** proposed, NOT applied` (or "investigation complete, nothing
   proposed for application tonight" for P5) `· **Source:** <peer>, <date>
   · **Evidence:**`/`**Full text:**`/`**Full spec:**` pointers to
   `artifacts/`/`status/probes/` files · `**Decide:** Gaetan` · dependency
   note if any (`**Must land together with [P#](...)**` or
   `**Independent of [P#](...)**`).

3. **Framing section** — heading varies by content, always answers "why is
   this here / what's the verdict": `## Problem` (P1), `## Verdict: ...`
   (P2), `## What changes, and why v2 is *smaller* than v1` (P3), `## Why`
   (P4), `## The two findings that reshape the question` (P5). One short
   paragraph or a compact table, states the conclusion before the mechanism.

4. **Mechanism / what changes** — the substantive middle, heading and shape
   vary freely with content: a fenced diagram (P3's `Mechanism` code block),
   a YAML diff with inline comments (P1, P2), a ranked table of candidates
   with verdicts (wf5-orca-delegation's own style, reused inside P3), or
   named sub-findings each with **bold lead-in** + evidence citation (P5's
   "Three standing constraints"). Always cites the evidence file backing
   each claim inline, in parens or a trailing clause — never a bare
   assertion.

5. **Decisions already settled vs. still open** — near-universal pattern:
   a short prose block for what Gaetan already ruled on ("Scope settled by
   your ruling" in P3), then a **table of open sub-decisions**, each row
   `# | Question | Note`, numbered `P<N>-a`, `P<N>-b`, ... so they can be
   referenced and checked off individually later.

6. **Order · restart · verify · rollback** — P3/P4 fold these into one
   combined `##` section with `- **Order:** ... - **Restart:** ...
   - **Verify:** ... - **Rollback:** ...` bullets. P1/P2 instead give
   "Order of operations" as a numbered list (P1, because a SOUL step must
   land *first*) and separate `## Verification after applying` / `##
   Rollback` headings. Choose combined-vs-separate by whether ordering
   itself is a hazard (P1: yes, SOUL-first-or-a-visible-error) or not.
   Every proposal has all four concerns somewhere, never omitted.

7. **Risk, stated plainly** — a short `## Risk` (P2) or `## Risk to state
   plainly` (P4) paragraph, separate from the open-decisions table: names
   what changes in the *security/blast-radius* frame, not what could go
   technically wrong. Explicitly cross-references sibling proposals when
   the combination compounds risk (P4 ↔ P1: "the blast radius is the
   product, not the sum").

8. **Footprint / cost paid so far** (P3 only, optional) — one paragraph
   naming any actual mutation the investigation itself made (an empty
   orchestration run, in P3's case) — recorded, not chased, so review
   doesn't have to wonder what's dirty.

9. **Decision** — always last, always a markdown checkbox list, always
   unchecked (nothing is decided by writing the proposal):
   ```
   - [ ] <primary option 1>
   - [ ] <primary option 2, often "Apply with edits">
   - [ ] Not yet / No-go
   - P<N>-a <label>: [ ] <option> · [ ] <option>
   ```
   One line per open sub-decision from §5's table, each with its own
   inline checkbox pair — so Gaetan can tick individual sub-calls without
   re-writing prose.

**Index file (`README.md`) pattern**, for the P6 line to add there (not
part of P6.md itself, noted for completeness): a table `# | Proposal |
Needs | Independent?`, a numbered "Reading order" recommending which to
read first and why, and a "Cross-cutting risk, stated once" closing
section — do not duplicate a proposal's own risk section here, only the
*combination* risk across proposals.

**Tone, held constant across all five:** concise, declarative, no hedging
adjectives ("might", "could" are used only where genuinely unverified and
flagged as such — e.g. P5's "untested from cooper"). Every load-bearing
claim carries its evidence file inline. Verbatim source quotes (`--help`
text, code comments, SOUL rule text) are marked as verbatim and quoted
exactly, not paraphrased, when the exact wording is the point (P3's `--ack`
quote, P4's rule 8 quote).

---

## Open gaps — what the existing docs do NOT answer about pushing to a third-party repo

1. **No trigger design exists anywhere.** P3/P4/wf5-orca-delegation design
   the *mechanism* of delegating to Orca on cooper for arbitrary work; none
   of them say when a push-back to mc-metarepo would fire — on Gaetan's
   explicit word only, on Tars' own initiative after some
   threshold/heuristic, or scheduled (à la the hourly *pull* refresh this
   same P6 proposal covers)? P6 must decide this; nothing here rules any
   option in or out.
2. **No design for how Tars decides WHAT to push.** The whole delegation
   machinery assumes a human (Gaetan, via Slack) supplies the brief content.
   A push-phase brief instead needs Tars to have *noticed* something worth
   writing back — but Hindsight memory is deliberately off (v1, no keys, per
   the task's starting facts) and Phase 2 (Gaetan's knowledge/preferences in
   SOUL) is explicitly "not built yet" in both P3 and the spec. Nothing here
   says what feeds a push brief's content in the interim.
3. **Two parallel delegation mechanisms exist and neither doc reconciles
   them for this use case.** `docs/specs/wf5-orca-delegation.md` (ssh + orca
   CLI, spawning a real Claude Code worker in a worktree) and
   `docs/specs/wf5-kanban.md` (12 `kanban_*` tools, a dispatcher that spawns
   workers to drive board cards ready→running→done) are both live surfaces
   Tars can reach. Kanban is oriented at task-board workflow, not
   named/scoped to any particular repo in the spec read. Nothing decides
   whether a scheduled/recurring push-phase task should be a Kanban card
   (dispatcher-driven, board-visible) or a direct Orca `run-create` (ad hoc,
   Slack-reported) — this is a real fork P6 has to resolve, not a detail.
4. **The VM's own mc-metarepo clone vs. cooper's Orca-registered clone are
   two different checkouts, and no doc says explicitly that push must never
   go through the VM's clone.** The task's own starting facts describe
   `~/dev/mc-metarepo` on the **Tars VM** (gh device-flow auth, HEAD tracking
   origin/main) as the search/pull-refresh target. Cooper's Orca-registered
   mc-metarepo is a **separate** clone at `/home/gaetan/dev/mc-metarepo`
   (repo `id:8099e312-3232-46f2-83a9-97aeaf5de5a2`). None of the read docs
   states the design intent this implies (push always via a delegated Orca
   worker on cooper, never via the VM's read-only pull clone) as an explicit
   rule — it's inferable from the separation-of-concerns frame, not written
   down.
5. **`orca repo add --path` (registering a new repo) is undocumented by
   measurement** — cited in the spec, never run in any probe read here. Moot
   for mc-metarepo specifically (already registered) but relevant if the
   push-phase design ever needs to reach a repo not already known to Orca.
6. **No rule ties SOUL rule 8 (announce-then-act) to a specific push-phase
   step.** It's phrased generically ("an action that cannot be undone");
   whether opening a PR against mc-metarepo counts, and whether "announce"
   means a Slack message before *delegating* the push task or before the
   delegated agent's own `git push`/PR-open step (which Tars does not
   control turn-by-turn once dispatched), is not resolved anywhere.
7. **B1 (`approvals:` mode) is PARKED** — the actual code-level gate on
   what the *spawned* Claude Code session can do (including `git push`,
   which sits in the "dangerous patterns" tier, released under `off`) is
   undecided. A push-phase design built assuming `off` is landed would be
   building on an unmade decision.
8. **No design for push failure/retry/conflict handling** — what happens if
   mc-metarepo has diverged since the worktree was created, if the PR
   conflicts, or if the delegated agent's push is rejected — is out of
   scope for every doc read (P3's "Known ceilings" covers timeout/mutation-
   with-no-return/worktree accumulation, not push-specific failure modes).
9. **Worktree keep-by-default compounds for a recurring push-phase.**
   v2's stated default is to keep every worktree (it's "the work product");
   mc-metarepo already has 9, several stale. A push-phase that runs
   hourly/periodically and keeps every worktree needs its own cleanup
   policy — not addressed, and explicitly "prune only on Gaetan's word" as
   written today, which does not scale to an unattended cadence.
10. **P3+P4 are themselves undecided.** Every mechanism this file describes
    (sudo-capable Tars, no sandbox, sudo-including cooper access) is a
    **proposal**, not shipped state — both checkboxes in P3.md and P4.md are
    unchecked. Any push-phase design is provisional on those landing as
    written, or must separately specify what changes if they land "with
    edits" instead.

---

## Sources read (this file only)

- `docs/proposals/P1-thread-follow.md`, `P2-emoji-trigger.md`,
  `P3-orca-v2-skill.md`, `P4-soul-guidelines-redesign.md`, `P5-a2a.md`,
  `README.md`
- `docs/specs/wf5-orca-delegation.md` (v1 + v2 authoritative sections),
  `wf5-kanban.md`
- `status/probes/wf5/orca-v2-capabilities.md`,
  `orca-v2-current-skill.md`
- `artifacts/SOUL-proposed.md` (for verbatim SOUL rule text — P4.md
  paraphrases rule 1, does not quote it)
- `docs/facts.md` (background context, per task instructions this repo's
  hard rules require reading before touching anything)
