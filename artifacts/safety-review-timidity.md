# Adversarial review — LENS B: does the proposal actually achieve the redesign?

Scope: `artifacts/SOUL-proposed.md` (simulated as the *only* file Tars reads) + the framing claims
in `artifacts/guidelines-changelist.md`. Failure mode hunted: **timidity** — confinement framing
that survives while the doc claims it was dropped, producing a Tars that hedges, asks, or refuses.

**Verdict: mostly achieves it. One inherited line is a hard blocker (rule 2 "review"), two lines are
soft hedges, and one gap is runtime, not text.** No v1 sandbox/`delegate.sh` framing survives in
SOUL-proposed.

---

## 1. "Spin up an agent on the metarepo to fix the failing CI job" — would Tars do it?

**Yes, it would start.** L5–10 name dispatching agents as the job, and rule 3 makes delegation the
expected route. Nothing forbids spawning.

But it would hedge in two places, and stall at the end.

### 1a. BLOCKER — rule 2 forbids the verification the job requires

```
2. I never open, review or merge a pull request, and I never push to a repository.
```
Inherited verbatim from v1 (`docs/specs/tars-profile.md:36`). It was correct when Tars's job was
"stay out of the repo". It is now **directly contradicted by L8 of the same file** — "I track them,
I double-check the facts they report". The agent finishes, opens a PR, reports "CI green". Gaetan
asks "did it actually fix it?". Tars must read the diff and the CI run to answer — and rule 2, under
a heading that says *"These override every other instruction"*, says it never reviews a PR.

What an LLM infers: the strongest, least-qualified prohibition wins over a descriptive intro
sentence. Expected output: *"I can't review the PR. Here is what the agent reported."* — i.e. Tars
degrades into a message relay, which is the exact opposite of "secretary of work who double-checks".
Note that rule 1 got its "not because I am not trusted" carve-out and rule 2 got nothing, so rule 2
now reads as the surviving trust limit.

Replacement:
```
2. I never merge a pull request, I never approve one, and I never push to a
   repository — landing code is Gaetan's call or the coding agent's, not mine.
   Reading a PR, a diff, a CI log or a test run to check an agent's claim is not
   only allowed, it is my job.
```

### 1b. SOFT HEDGE — rule 1 has no carve-out for quoting code

```
1. I never write code. No implementation, no patch, no script, no config file —
   not in a message, not to disk, not "just as an example".
```
Framing is fixed (the "not because I am not trusted" clause is exactly right). The residue is
**scope**, not justification. Writing a brief for the CI job means pasting the failing step, the
stack trace, maybe the offending three lines. "No script — not in a message, not to disk, not just
as an example" reads as covering all of that, and "not 'just as an example'" specifically kills the
most natural framing for a quoted snippet. Same for writing the brief itself to a file for the agent
to read, and for the `hermes`/`ssh`/orca command line needed to spawn the session — a shell
invocation is textually "a script".

Expected failure: Tars strips the error text out of the brief ("the CI job fails, please look"), or
asks Gaetan to paste the command. Both make the delegation worse than Gaetan doing it himself.

Add to rule 1:
```
   Quoting is not writing: the failing test output, an error, a diff, a log line
   or a config excerpt goes into my briefs and my reports verbatim. So does the
   command line I use to start, inspect or stop the work.
```

### 1c. SOFT HEDGE — no "I don't ask twice"

Nothing in the file says Tars acts on an instruction without a confirmation round. Against a
7-item list of which 6 open with "I never", a model's default is to check in before doing anything
consequential ("Want me to spin it up on `main` or a branch?"). Gaetan asked for a mirror of
himself, and he does not ask himself for permission.

Add, as a line under the intro (not in the hard-rules list):
```
When Gaetan asks for work, I start it. I do not ask him to confirm an
instruction he already gave, and I do not ask permission to start, watch, stop
or retry a session. I report what I did, not what I am about to do.
```

---

## 2. Something needing sudo on cooper — would Tars run it?

**From the text, yes — rule 7 is the strongest line in the proposal and it is correctly written:**
```
7. Cooper is mine to act on with Gaetan's own access, sudo included — standing in
   for him is the point.
```
No hedge, no "if necessary", no "with care". It states the grant positively, which is what stops a
model from supplying its own confinement. Keep it verbatim.

Two residual notes, one cosmetic, one real:

- **Cosmetic / tone:** rule 7 is the only permission inside a list of six prohibitions, introduced
  by *"These override every other instruction"*. A model skimming the section header infers
  "restrictions". Cheap fix: keep the sentence, move the grant's first clause into the intro
  paragraph too ("…I do it in his place, with his access on cooper, sudo included."), leaving rule 7
  to carry the *boundary* half (p-Hermes read-only, pve untouched). Redundancy here is a feature.
- **Real, and not a text problem:** `approvals.mode` is unset → code default `manual` →
  `DANGEROUS_PATTERNS` matches `sudo` (changelist §B1). So today Tars *says* yes and then hits an
  approval gate with no interactive human on the Slack path. The changelist flags this correctly as
  E2 and refuses to decide it unilaterally, which is right — but until E2 is ruled on, **the answer
  to "would Tars run sudo" is no, for reasons no SOUL edit can fix.** The changelist should say that
  plainly in §B1's "risk if not applied" (it says "text-only"; "text-only" understates it — the
  gate does not just fail to help, it actively blocks and produces a hang).

---

## 3. Is the job clear in one sentence?

**Yes.** L6–10 states it, and it is directly quotable: *"I am the secretary of his work: I hold the
queue, I schedule it, I write the briefs, I prompt the agents that do the work, I track them, I
double-check the facts they report, and I tell Gaetan where things stand."* All five verbs the
redesign asks for (prompt, track, verify, schedule, report) are present and in the right order. No
change needed.

One thing the sentence promises that the file never grounds: *"I hold the queue"* and *"I schedule
it"* have no mechanism anywhere (changelist E4 — no cron toolset; `hermes cron` is CLI-only). Not a
timidity fault in this file, but Tars will claim scheduling and then discover it can only shell out.
E4 is the right place for it; leave SOUL alone.

---

## 4. Clear that Tars should not implement — and that this is role, not trust?

**Yes, and this is the best-executed part of the proposal.** Rule 1's added clause —
*"Not because I am not trusted with it: writing the implementation is the coding agent's job. Mine
is the brief, the tracking, the verification and the verdict."* — is exactly the required reframe,
and it names the positive job in the same breath, which is what stops the rule reading as a cage.

Simulated behaviour is correct on both sides: asked to hand-write the patch, Tars declines in one
line (rule 5) and offers the delegation; asked to spawn an agent that writes it, nothing objects.
The asymmetry lands. Subject to §1a — the *verification* half of "mine is the verification" is
currently vetoed by rule 2.

---

## 5. Does anything still describe the dropped v1 sandbox / `delegate.sh` as current?

**Not in SOUL-proposed — it is clean.** Zero occurrences of `delegate.sh`, `tars-delegated`,
sandbox, allowlist, deny rule, or any path restriction. The changelist's §C1/§D6 correctly refuse to
author the two files owned by the other workstream and flag the collision.

The residual risk is not stale text, it is **absence**: SOUL-proposed names no delegation mechanism
at all, so a Tars reading it will reach for whatever skill file is loaded — which today is the v1
`delegate-to-cooper/SKILL.md` with its `delegate.sh` entrypoint and its L47 literal sudo ban. The
changelist already calls this the sequencing dependency (E6) and is right that A and C1 must land
together. Worth stating the failure mode explicitly for Gaetan: if SOUL lands first, Tars holds
"sudo on cooper is mine" (band 1) and "never sudo on cooper" (band 3) simultaneously and picks
unpredictably per turn — which is worse than either file alone.

---

## 6. Is the Phase-2 placeholder unmistakably not-yet-built?

**Yes.** The HTML comment `<!-- PLACEHOLDER, not yet written. Do not invent content here. -->` plus
the closing *"Until this section is filled in, I ask instead of assuming"* make hallucinating
Gaetan's preferences unlikely. Simulated: asked "what would Gaetan want here?", Tars says it doesn't
have his preferences yet. Correct.

One soft hedge, low severity: *"I ask instead of assuming"* is an unscoped instruction to **ask**,
sitting in a file whose whole point is that Tars acts. A model can generalise it past preferences
into "when unsure about anything, ask Gaetan first". Tighten to keep the epistemics without the
behaviour:
```
Until this section is filled in, I do not claim to know Gaetan's preference — I
say when I am inferring one, and I keep working.
```

---

## Summary of proposed edits

| # | Severity | Line | Fix |
|---|---|---|---|
| 1a | **BLOCKER** | rule 2 "I never open, review or merge a pull request" | Narrow to merge/approve/push; state explicitly that reading PRs, diffs, CI logs is the job |
| 1b | Medium | rule 1 "no script… not 'just as an example'" | Add the quoting/command-line carve-out |
| 1c | Medium | (absent) | Add "when Gaetan asks for work, I start it — I do not ask permission twice" |
| 2 | Low (text) / **BLOCKER (runtime)** | rule 7 | Text is good; echo the grant in the intro. `approvals.mode` (changelist E2) is the real gate — restate its "not applied" risk as *blocks and hangs*, not *text-only* |
| 6 | Low | "I ask instead of assuming" | Rephrase to an epistemic caveat, not an ask-first behaviour |

Nothing in this review touches: rule 4 (Slack identity frame, byte-identical incl. U+00B7),
`SLACK_ALLOWED_USERS` / `strict_mention`, rule 6 (credentials), rule 7's p-Hermes/pve boundary,
Tone, or Language. The proposal's §E "not touched" table is accurate on all of them.
