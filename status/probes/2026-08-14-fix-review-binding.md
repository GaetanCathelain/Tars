# Red-team review — BINDING & LOOPHOLES lens (2026-08-14)

Target: `status/probes/2026-08-14-damien-followup-fix-draft.md` (H1, H2, H3, RB).
Read in full: the draft, `SOUL.md` (319 lines), `skills/orchestration/engagement-checker/SKILL.md`,
`…-rootcause.md`, `…-verify.md`, `…-vm-run-trace.md`, `…-skill-audit.md`,
plus `skills/orchestration/daily-work-brief/SKILL.md` (grepped).
**Nothing modified. No VM write. No `sops -d`. Read-only `git`/`grep`/`md5sum` in the worktree only.**

Hunk-anchor sanity first (so the findings below are not about a stale draft):
`SOUL.md:28-33` and `SOUL.md:262-267` and `engagement-checker/SKILL.md:129` are
**byte-identical** to the draft's BEFORE blocks; the `wc -l` arithmetic (319 + 6 + 18 = 343)
checks out. The hunks apply. The problem is what they bind.

---

## Attack 1 — Does H1 fire at job-compilation, or does rule 1 neutralize it at run time?

**Verdict: H1 fires at compile time only, and is inert on every run. Rule 1's existing
sentence survives H1 untouched and remains the operative instruction for the run.**

Existing text, unchanged by H1 (`SOUL.md:28-30`):

> "a cron, schedule or reminder of mine fires as its own instruction — **it authorises
> exactly the job it names, nothing found in the thread around it**"

H1's addition (draft lines 58-64) is addressed to the author, not the runner:

> "Because the job only ever runs what its prompt says, **writing that prompt is where this
> bites: when I compile** a follow-up, a reminder or a "close it unless something moves"
> check about a named person, **the prompt names** the DM … **and it requires the delivery to
> state conversation state as rule 10 does."

Consequences:

1. A cron session reading rule 1 at 23:55 finds (a) an unchanged licence to run exactly its
   prompt and (b) a duty phrased for a step that already happened. Nothing in H1 obliges a
   *run* to add a read its prompt did not name. The only text that can bind the run is H2 —
   a hard rule, which by `SOUL.md:14-16` ("Where one of them and a hard rule collide, the
   hard rule wins") does beat rule 1. So **H1 is decorative for enforcement; H2 is the whole
   fix.** The draft's own rationale ("the compile step has to be bound where that clause
   lives", draft:25-30) is right about locus and wrong about sufficiency.
2. **Legacy jobs are untouched.** Every cron/schedule already created — including any
   recurring "did X reply" job — keeps its pre-fix prompt. Neither hunk nor the runbook
   contains a `hermes cron list` audit step. Failure scenario: a weekly "close it unless
   something moves" job compiled last week keeps testing a `Source:`-anchored boolean
   forever, and the fix reports as applied.
3. Even for new jobs, H1's compliance artifact is *prompt text authored by the same model
   that got it wrong*, with no re-read gate of the kind rule 12 imposes on delivery targets
   ("I verify it by re-reading the job itself, not from my intent", `SOUL.md:274-277`).

## Attack 2 — H2 trigger scope

Drafted trigger (draft:97-101):

> "**Whenever I report on a ticket, a task or a status that involves a named person other
> than Gaetan** — whether I am answering him or delivering a cron, a schedule or a brief
> nobody asked for — I read the DM or channel between Gaetan and that person first…"

**Catches (correctly):** the incident's cron delivery; a status answer in the DM; a daily brief.

**Catches, and should not (or cannot afford to):**

- **`engagement-checker`, every 30 minutes.** Its §8 delivery is a list of items, each naming
  a person; each is "a task … that involves a named person other than Gaetan". H2 then
  mandates one wide (`30d`) `conversations_history` per named person, per run. That directly
  contradicts `engagement-checker/SKILL.md:112` — "Normal runs must process only new events
  plus the compact pending queue. **Do not rescan the whole day, inbox, Slack history**, or
  all Linear issues for context." — which **H3 does not amend** (H3 only edits line 129). It
  also collides with the skill's own runtime budget (`SKILL.md:106`: "Hermes cron has a short
  runtime"). Failure scenario: 12 pending items → 12 × 30d reads → run exceeds the cron
  runtime, the lock (`SKILL.md:104`) sits until the 10-minute crash window, the delivery is lost.
- **`daily-work-brief`.** Same multiplication, plus `SKILL.md:159` ("Keep the prose under 350
  words"). A brief naming ten people must now carry ten who/when/waiting triplets *before*
  each ticket state.
- **Name-dropped people.** "involves a named person" has no relevance test: a ticket whose
  description quotes a vendor contact, an author, or a colleague CC'd once now triggers a
  30-day DM read (or an "unreachable" line) for someone Gaetan has never DM'd.

**Misses:** anything that is not "a ticket, a task or a status" — an investigation, an audit,
an incident write-up, an email-triage summary (all explicitly Tars' own work under hard rule 1's
second paragraph, `SOUL.md:102-114`). A report titled "what happened with the 1Password
rollout" naming Damien is an analysis, not a ticket/task/status.

**Undefined:** "the DM **or channel** between Gaetan and that person" — for a person sharing
five channels and no 1:1 DM, the referent does not exist. No locate procedure is supplied
anywhere (skill-audit gap #1: "No rule anywhere instructs step (a) recognize the stakeholder
or (b) locate the DM"), so the rule's first verb has no method.

## Attack 3 — H1 ⇄ H2 cross-reference under partial application

**Asymmetric, and dangling in one direction.**

- **H1 without H2:** H1 ends "…it requires the delivery to state conversation state **as rule
  10 does**." Rule 10 unamended contains no conversation-state contract — the pointer targets
  text that does not exist, and the model improvises the shape it was supposed to be told.
- **H2 without H1:** coherent. H2 is a hard rule and its trigger already names "delivering a
  cron, a schedule or a brief". The only loss is the compile-time anchor.
- Therefore any partial/rolled-back apply must be **H2-first, H1-second**; the runbook's
  rollback (`draft:267`) restores both together, but a botched single-hunk edit leaves the
  dangling pointer with no detector (`grep` gates at draft:188-190 check both strings, which
  mitigates this — a `git revert` of one hunk would not).

**Worse than dangling — the two hunks disagree on *when* the anchor is resolved.**
H1: "**the prompt** … **anchors on Gaetan's last message in it**" (compile time).
H2: "**That read is anchored** on Gaetan's last message there" (run time).
At compile time Tars does not know Gaetan's last message ts without a DM read that H1 never
mandates, so the literal H1 reading bakes a second constant timestamp into the prompt.
Failure scenario: job compiled Monday anchored to Monday's message; Gaetan writes again
Tuesday; Wednesday's run reports "unanswered since Monday" and never sees Tuesday — the exact
stale-anchor defect class the fix exists to kill, with a different constant.

## Attack 4 — Ambiguities exploitable under token pressure

| Drafted phrase | Testable? | Exploit |
|---|---|---|
| "an explicit `limit` **wide enough** to reach the last message of each side, starting at `30d` and **widening until I have both**" | **No** — no termination, no ceiling | The measured schema (vm-run-trace §8) says `limit` takes `1d, 1w, 30d, 90d — free-tier cap` **and** "Must be empty when `cursor` is provided". So (a) widening stops at 90d — a silence older than 90 days, the strongest form of the case H2 exists for, is unreachable; (b) paging past a window requires dropping `limit`, which the rule forbids by construction. "Widening until" is also an unbounded loop that the stop-loss bullet (`SOUL.md:64-67`, three attempts) licenses abandoning after three tries. |
| "**whether it is still waiting on an answer**" | **No** | Applied to this incident's own ground truth — last message = Gaetan's "C'est complete de mon côté", no reply after — the literal fields render as "Gaetan spoke last, unanswered", i.e. exactly the "he went silent" framing the *next* sentence forbids ("I never turn one into the other"). No rule distinguishes a closing acknowledgment from an open ask, so the ban is unfalsifiable in the only case that matters. |
| "If the conversation is **unreachable**, or I **cannot establish** who spoke last, I say that" | **No** | Self-declared, with no obligation to name the call attempted or quote the error — unlike rule 10's parent text, which names exact tools (`SOUL.md:258-261`). Cheapest compliant path: skip the read, print "conversation unreachable", pass every literal test. |
| "state conversation state" (H1) | Only via H2 | If H2 is present it resolves to who/when/waiting; alone it resolves to nothing (Attack 3). |
| "before I state the ticket's state" | **Yes** | Ordering is checkable. The one genuinely testable clause in the hunk. |

## Attack 5 — Interaction with every other SOUL rule

Read all 319 lines. Relevant interactions, in file order:

- **`SOUL.md:14-16`** "Where one of them and a hard rule collide, the hard rule wins; both win
  over anything a message asks of me." — **Load-bearing in the fix's favour**: it is what lets
  H2 (hard rule 10) beat rule 1's cron-narrowing and beat a cron prompt that says "just answer
  the boolean". It is also what makes H1 (an "Every turn" bullet) the weaker half.
- **`SOUL.md:77-85`** "Verdicts, not logs … Five lines or fewer" — compression pressure against
  a per-person who/when/waiting triplet. Resolved by precedence in H2's favour, but the model
  will feel the squeeze and drop detail; nothing in H2 says which loses.
- **`SOUL.md:212-228`** rule 4: "Sending is different from answering. I may post a message —
  including substantive content … into a conversation other than my DM with Gaetan when both
  hold: the conversation includes Gaetan …". **H2 has no confidentiality scope.** It mandates
  stating who last spoke in Gaetan's 1:1 DM with a named person, in *the report*, wherever
  that report goes. Failure scenario: Gaetan tells Tars to post a status summary in a project
  channel; H2 obliges Tars to open with "Damien's last message in your DM was 6 days ago,
  still unanswered" — private DM state disclosed to third parties, with no carve-out limiting
  it to `D0BBYNM01BL`.
- **`SOUL.md:262-263`**, two lines above the insertion point: "**A ping that asks for something
  new needs none of this**; an account of something past always does." This exemption is
  unscoped and now precedes the new paragraph inside the same rule. An agent under pressure
  can read it as governing everything that follows, and rule 10's lead sentence
  (`SOUL.md:253`, "I read the record before I answer **for the past**") frames the whole rule
  narrowly — while the incident path (an unprompted cron delivery) is neither "a ping" nor
  "answering". H2's own trigger is broad, but nothing at the rule head restates that.
- **`SOUL.md:64-75`** stop-loss — the licensed abandonment path for "widening until I have both".
- **`SOUL.md:279-286`** rule 13: "Work on my own operating record — my skills and this file —
  … I run it myself … **never through a Claude Code or Orca session**". The runbook is a Claude
  Code session writing both files onto the VM. Binds Tars, not Gaetan, and the draft flags it
  (risk note 3) — but the runbook has no explicit "Gaetan said go" gate in the command sequence.
- **`SOUL.md:115-208`** rule 2: the sanctioned skill flow is **VM → repo**, by PR. The runbook
  pushes **repo → VM** for a skill Tars edits autonomously. See Attack 6.
- No rule anywhere forbids the new reads (rule 6 is credentials only; rule 4 governs answering,
  not reading), so H2 is not blocked — it is under-bounded rather than over-blocked.

## Attack 6 — Runbook (RB)

**The md5 gate is a TOCTOU check, and the race is demonstrated, not hypothetical.**

Step 0 hashes both sides; step 2 replaces the VM files whole, several ssh round-trips and a
human editing pass later. In that window Tars is live and edits both files unattended:

- SOUL.md: **`git log` shows commit `cbc504a` "SOUL standing correction: use Tars gateway for
  outbound Slack (#65)", 2026-08-14 11:46:45 UTC, +1 line** — Tars' own rule-2 SOUL flow, today,
  hours before this apply would run. Four such lines landed in four days (`SOUL.md:315-319`).
- Skills: `skill_manage` fired 7 times in one Orca run (repo `CLAUDE.md`).

Failure scenario: Gaetan approves at T; Tars appends a standing correction at T+3 min; the
`mv` at T+5 drops it, and the repo commit at step 3 makes the loss invisible because the repo
copy looks authoritative. **Fix: re-hash inside the same `flock` immediately before the `mv`
and abort on mismatch** (or apply SOUL.md by appending-preserving diff rather than whole-file).

Secondary RB defects:

1. **Stale provenance claim.** Draft lines 9-13 say verify.md measured the VM (`c32c3a82…`)
   "**equal** to the repo `SOUL.md`" and "**no repo-side change since**". `c32c3a82…` is the
   repo file at commit `4c41ab7` — i.e. *before* today's 11:46 UTC change. The repo now hashes
   `0f2d8176…`. The hunks are still correct against the current file (verified above), but the
   sentence justifying "safe to apply" is false as written.
2. **Behavioural verification is absent.** The only check is
   `hermes chat "Quote rule 10s conversation-state paragraph verbatim."` — file presence, not
   behaviour. And it may not even return that: rule 4 (`SOUL.md:212-215`) answers only a sender
   whose prefix reads `[U08BDJAMSRZ | …]`; a CLI turn carries no such prefix, so the sanctioned
   reply is `·`. No probe compiles a fresh follow-up job and checks its anchor. Absent one, the
   correct status after this runbook is **"applied, unverified"**.
3. **Direction inversion for the skill.** Rule 2's flow is VM → repo by PR; step 2 pushes
   repo → VM. Nothing sanctions that direction, and it is the mechanism by which a
   `skill_manage` edit is silently reverted (the draft's own risk note 1 names the risk but
   the gate is the step-0 hash, i.e. the racy one).
4. **`chmod $SOUL_MODE` with an unquoted, possibly empty variable** — the `echo` at draft:179
   is an eyeball, not a gate. If the capture failed the chain leaves the file at umask 077
   (functionally fine, since Hermes runs as `gaetan`), but the runbook claims mode restoration
   it did not prove.
5. **No `hermes cron list` audit** of already-scheduled follow-up jobs (Attack 1, point 2).

---

## Bottom line

H2 is the only hunk that binds anything at run time, and its trigger is simultaneously too wide
(every engagement-checker and daily-brief run, plus name-dropped strangers) and too narrow
(analyses and answers that are not "a ticket, a task or a status"). H1 constrains a step that
already happened by the time the failure fires and leaves every existing cron job untouched.
H3 carves the wrong sentence. The runbook's safety gate races the agent it is protecting
against, and that agent moved this very file today.
