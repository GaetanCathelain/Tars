# Adversarial review — Lens 2: would an LLM actually succeed reading this?

Target: `artifacts/delegate-to-cooper-SKILL.md`
Simulated actor: Tars, holding only that file + ssh to cooper.
Simulated prompt from Gaetan (Slack DM): *"get the metarepo's CI workflow files
documented, and tell me when it's done"*.

Method: I walked the skill as written, stopping at every point where I would
have had to guess. Two facts were checked live rather than assumed (marked
**[measured]**).

---

## Part 1 — The walkthrough, as I would actually run it

### Turn 1, before any command: do I even use this skill?

The `description` frontmatter says *"needs code written, **a file produced**, a
repo inspected"*. "Documented" = a file produced. I invoke the skill. ✅

But three lines into the body I hit the first real fork. §Separation says:

> I do not write the code, the patch, the script or the config myself

Four nouns, all of them code artifacts. The deliverable here is **markdown
prose about YAML files**. The cheapest path is obvious and available to me:
`ssh cooper 'cat /home/gaetan/dev/mc-metarepo/.github/workflows/*.yml'`, read
them, write the doc in my reply. That is *inspecting a repo* (explicitly
allowed as a double-check: *"I may double-check facts myself: read a file, run
a read-only command"*) followed by writing prose, which the enumerated list
does **not** forbid.

**This is where I would most likely go wrong**, and I would go wrong while
believing I was inside the rules. The description saves the skill only if I
weight frontmatter over body text — and body text is more specific, so I would
normally weight it *higher*. Fix 1.

Assume I resist. Continue.

### "the metarepo" — which repo?

§Picking a repo says default `mc-metarepo`, `id:8099e312-…`. Gaetan's own
CLAUDE.md world is full of `gaetan-metarepo`, and he said "the metarepo", not
"the mc metarepo". I would run `repo list --json`, see `mc-metarepo` present
and `gaetan-metarepo` absent, and pick `mc-metarepo`. Recoverable — but only
because of a lucky registration fact, not because the skill told me. Fix 9.

### Step 0 — preflight

```
ssh cooper '~/.local/bin/orca status --json'
```
No hesitation. `runtime.reachable` is a named field with a named consequence
(stop, tell Gaetan). Clean.

**[measured]** `ssh cooper` from the Tars VM resolves: `~/.ssh/config` on
192.168.0.9 has `Host cooper → HostName 192.168.0.4, User gaetan`. The alias is
real, so the skill is not wrong. But it is *unverified in the text*, and the
failure table maps any ssh failure to "cooper unreachable → report and stop" —
so a config drift would make me report a **false** infrastructure outage. Fix 5.

### Step 1 — brief onto cooper

```
ssh cooper 'cat > /tmp/tars-brief.md' <<'BRIEF'
```
Hesitation #2, small but real: §Separation said *"not to disk"*. I am about to
write a file to disk on cooper. I can talk myself past it (prose ≠ code), but
the sentence and the sequence contradict each other on their face. Fix 1 fixes
this too.

Second hesitation: `/tmp/tars-brief.md` is a **fixed path**. Two concurrent
delegations, or a re-run after a timeout, silently clobber it. Nothing in the
skill says to uniquify. Minor, Fix 10.

I write the brief. §Writing the brief's 5-point structure is genuinely good —
"Done means" and "no reference to the Slack conversation" are exactly the two
things I would otherwise get wrong.

### Step 2 — run-create / task-create

```
… run-create --objective "Document mc-metarepo CI workflows" --json
```
The skill pre-warns that `result.run.id` is unconfirmed and tells me to read
the JSON and report the real shape. That is the right instruction and I would
follow it. ✅ Honest uncertainty, correctly handled.

`task-create --spec "$(cat /tmp/tars-brief.md)"` inside single quotes: the
comment explains *why* the quoting is that way. Correct, and I would not have
got it right unaided.

### Step 3 — cheap checkpoint. Fine, and the `run_not_found` cross-check with
`task-list` is a nice save.

### Step 4 — worker-start

```
--worktree new-top-level --repo id:8099e312-… --name <slug> --agent claude --setup run
```
Hesitation #3, and I would have to **invent**: what is `<slug>`? Is it a shell
slug (`ci-workflow-docs`) or a display name (`"CI workflow docs"`)? Orca
worktrees demonstrably carry spaced display names elsewhere in this repo's
evidence (`name:"GoCardless rotation"`). The skill's own predicted path
`/home/gaetan/orca/workspaces/mc-metarepo/<slug>` implies a filesystem-safe
token, but never says so, never gives a charset, and never says whether a name
collision with one of mc-metarepo's 9 existing worktrees is an error or a
silent reuse. I would guess `ci-workflow-docs` and probably be right — a guess
is a guess. Fix 4.

### Step 5 — the wait, and the 300s cap

```
… check --run … --wait --types worker_done,escalation,question --timeout-ms 240000 --json' 2>/dev/null
```
Documenting CI workflows in a 9-worktree monorepo takes an agent well over 4
minutes. So: `{count:0}`, exit 0.

**Does the skill make the next move obvious? Yes — this is its strongest
section.** §"The 300s cap" states the principle (checkpoint ≠ failure), gives
the literal sentence to send, and gives both re-attach commands. I would not
report failure. Unambiguous win. ✅

One flaw in the mechanics: `2>/dev/null` is applied to the **local ssh
process**, so it also swallows ssh's own error text. A network blip during the
240 s wait gives me a nonzero exit and *no output at all* — the least
diagnosable state possible. Fix 7.

### Turn 1 ends. And here the skill fails the actual request.

Gaetan asked **"tell me when it's done"**. I reply:

> Still running. Run `run_…`, dispatch `dispatch_…`, worktree `/home/…`. I will
> re-attach.

**I have just promised something I have no mechanism to do.** I am a Slack
agent. I do not get a turn unless someone speaks to me. The skill says "on any
later turn, re-attach" — it never asks *what creates a later turn*. Nothing in
this file wakes me. Gaetan's request is answered with a lie of omission, and
the run completes 20 minutes later into silence.

The mechanism exists and is documented in `docs/facts.md` (async delivery to
the home channel: `hermes cron create "…" "…" --deliver slack --repeat 1`) but
it is **not in this skill**, and this skill is what I am reading. Fix 2 — this
is the single most consequential omission for this exact prompt.

### Steps 6–7, on the hypothetical later turn

`worker-read --limit 200` — no guidance on what 200 transcript messages does to
my context, or that I should read the tail. Minor, Fix 10.
`worker-release` — the "read the returned state, not the exit code, all four
are exit 0" note is excellent and I would have got it wrong without it. ✅

---

## Part 2 — The layout-model self-test

Asked *without re-reading*: **what is a child worktree for?**

My answer from memory: *nothing I need.* It is Orca's own bookkeeping recording
which worktree's terminal spawned which — provenance, not git ancestry (the
child still branches off `main`) and not disk nesting (the child directory is a
sibling). Orca infers it silently from the calling terminal, so it is
meaningless from an ssh call that has no terminal context. I always use
`new-top-level`.

That is correct, and I retained it because the section states the negatives
explicitly ("not git ancestry", "not disk nesting") instead of just defining
the term. **The layout model holds.** I would not conflate:

- terminal vs agent session — the box says outright there is no separate
  agent-tab type, an agent session *is* a terminal whose process is `claude`;
- terminal vs dispatch — dispatch is a durable pointer *at* a terminal, and
  the ephemeral/durable split is repeated in the vocabulary table;
- browser tab vs terminal tab — flagged as a separate family with a separate
  verb namespace;
- repo vs Project — "Orca also has a formal Project object… I almost never
  need it. Use `--repo`."

Residual soft spot, not a conflation: "a terminal lives at a tab + a pane
(leaf), join key `paneKey == <tabId>:<leafId>`" is the one sentence I retained
*worst*, because nothing in the sequence ever uses it. It is inert detail.

---

## Part 3 — Am I too timid? Am I too free?

**Clear enough that I must not implement:** yes for code, **no for prose** —
see Fix 1. That is the one hole.

**Clear enough that I am otherwise unrestricted:** *mostly*. The paragraph
"This is a division of labour, not a security boundary. There is no sandbox, no
command allowlist, no forbidden path list, no wrapper script" is direct and
does the job. Two residual sources of timidity:

1. **`sudo` is never mentioned, positively or negatively.** The absence of a
   prohibition is not the same signal as a permission. My default posture on an
   unmentioned `sudo` on someone else's machine is to stop and ask. Gaetan
   explicitly wants that reflex dead. One affirmative sentence kills it. Fix 6.
2. **"The limit on what I run is what the job actually needs — nothing more,
   and nothing enumerated."** The clause reads as a soft minimality rule and
   invites me to litigate whether an exploratory `ls`, a `git log`, or reading a
   file outside the target repo is "what the job actually needs". That is
   exactly the over-cautious dithering the design is trying to prevent. Fix 6.
3. Third, narrower: *"I do not write the code… not in a message"* would, read
   strictly, stop me from **quoting the agent's YAML in my report** — while
   §What I always report demands verbatim evidence. The two sections disagree.
   Fix 6.

---

## Part 4 — Would I actually quote the command and the run id?

**Yes.** "a reply without the run/dispatch ids is not a reply" is the kind of
absolute phrasing that survives compression into a Slack answer, and the
ordering is numbered. This section works.

Two defects, both about *which* text:

- *"the exact command I ran, verbatim, brief included"* — I ran **seven**
  commands. Verbatim-all is 40 lines of bash in a Slack DM; verbatim-some means
  I choose, and I would choose inconsistently between turns. Fix 3.
- The §300s "still running" template contains **no command at all**, directly
  contradicting "Every reply that used this skill states, in this order: 1
  verdict, 2 exact command, 3 ids". The interim reply needs its own named
  shape. Fix 3.

---

## Part 5 — Numbered wording fixes

### Fix 1 — (highest impact) prose and documentation are not exempt

**Current** (§Separation of concerns):
> - I do not write the code, the patch, the script or the config myself — not in a
>   message, not to disk, not "just as an example" (SOUL rule 1). If the shortest
>   path is to type the fix, I still brief an agent to type it.

**Proposed:**
> - I do not produce the deliverable myself — code, patch, script, config,
>   **documentation, README, migration notes, any artifact Gaetan asked to
>   exist**. Not in a message, not "just as an example" (SOUL rule 1). "It's
>   only markdown" and "it's faster if I just read the files and write it up"
>   are the two excuses this rule exists to refuse. If the shortest path is to
>   type it, I still brief an agent to type it.
> - Writing the brief itself to a file on cooper (`/tmp/tars-brief-…md`) is not
>   me doing the work — the brief is my instruction, not the deliverable.

### Fix 2 — (highest impact) make "I will re-attach" a mechanism, not a promise

**Current** (§The 300s cap):
> > Still running. Run `<RUN_ID>`, dispatch `<DISPATCH_ID>`, worktree
> > `<path>`. I will re-attach.
>
> and on any later turn, re-attach with either:

**Proposed:**
> Nothing gives me a later turn on its own — I only speak when spoken to. So a
> checkpoint reply is **always** two actions, never one:
>
> ```bash
> hermes cron create "5m" "Re-attach to Orca run <RUN_ID> dispatch <DISPATCH_ID> \
>   using delegate-to-cooper: check --wait, then report the outcome, the worktree \
>   path and the branch." --deliver slack --repeat 1
> ```
>
> then reply:
>
> > Still running. Run `<RUN_ID>`, dispatch `<DISPATCH_ID>`, worktree `<path>`.
> > I've scheduled a re-check in 5 minutes and will report back here.
>
> If Gaetan asked to be told when it is done, the cron is **not optional** — a
> reply that promises to re-attach without scheduling one is a broken promise.
> On the woken turn, re-attach with either:

### Fix 3 — say *which* command goes in the reply, and give the interim shape

**Current** (§What I always report):
> 2. **the exact command I ran, verbatim**, brief included;

**Proposed:**
> 2. **the `worker-start` command verbatim, plus the brief in full** — that one
>    command and that text are what actually determined the work; the preflight,
>    run-create and check calls are plumbing and stay out of the reply unless one
>    of them is what failed. If a command failed, quote *that* one instead.

**Add, after the numbered list:**
> An interim (still-running) reply is the exception to the ordering: it has no
> verdict yet, so it is ids + worktree path + "re-check scheduled", and it
> repeats the brief only if Gaetan has not seen it yet.

### Fix 4 — define `--name`, and check for a collision first

**Current** (§The sequence, step 4):
> `--name <slug> --agent claude --setup run --json'`

**Proposed:**
> `--name <slug> --agent claude --setup run --json'`
> `#   <slug>: lowercase, hyphens, no spaces — it becomes the directory name and`
> `#   the branch suffix. Derive it from the objective (ci-workflow-docs).`
> `#   mc-metarepo already carries ~9 worktrees; collision behaviour is NOT`
> `#   documented, so check first:`
> `#     orca worktree list --repo id:<repoId> --json  → grep the names`
> `#   and add a short suffix rather than reusing an existing name.`

### Fix 5 — give the ssh fallback so I never invent an outage

**Current:**
> I reach cooper over ssh as `gaetan` (`ssh cooper`), and I drive Orca through its
> CLI.

**Proposed:**
> I reach cooper over ssh as `gaetan` (`ssh cooper`, an alias for
> `gaetan@192.168.0.4` — if the alias ever fails to resolve, retry with the IP
> before concluding anything). I drive Orca through its CLI.

*(Alias verified live on the VM: `Host cooper → 192.168.0.4`, user `gaetan`.)*

### Fix 6 — kill the residual timidity, affirmatively

**Current:**
> This is a division of labour, **not** a security boundary. There is no sandbox,
> no command allowlist, no forbidden path list, no wrapper script. cooper is
> Gaetan's machine and I act there as Gaetan does. The limit on what I run is what
> the job actually needs — nothing more, and nothing enumerated.

**Proposed:**
> This is a division of labour, **not** a security boundary. There is no sandbox,
> no command allowlist, no forbidden path list, no wrapper script. cooper is
> Gaetan's machine and I act there with Gaetan's own hands: **any path, any
> command, `sudo` included, by design** — he has root there, so do I, and I never
> ask permission to run something that the job needs. Refusing to act, or
> stopping to ask whether I am allowed to look at a file, is a failure mode just
> as real as doing the implementation myself. The only thing I hold back from is
> *producing the deliverable* (above). Quoting the agent's output — including its
> code and config — back to Gaetan in my report is required evidence, not a
> violation of that rule.

### Fix 7 — stop blinding myself to ssh errors

**Current** (step 5):
> `--types worker_done,escalation,question --timeout-ms 240000 --json' 2>/dev/null`
> `#   2>/dev/null drops the JSON keepalive lines Orca emits on stderr every 15s.`

**Proposed:**
> `--types worker_done,escalation,question --timeout-ms 240000 --json' \`
> `  2>&1 | jq -c 'select(._keepalive|not)'`
> `#   Orca emits a JSON keepalive on stderr every 15s; the jq filter drops those`
> `#   and KEEPS real error text. (2>/dev/null also works but silences ssh's own`
> `#   failure messages — a dropped connection then looks like empty success.)`
> `#   If the output is empty and the exit code is nonzero, re-run without the`
> `#   filter and read what stderr actually said before reporting anything.`

### Fix 8 — mark `repo add` as unverified, like the other unconfirmed calls

**Current** (§Picking a repo):
> - If the work genuinely needs a codebase Orca has not seen,
>   `orca repo add --path <abs-path>` registers it — I run `--help` first, do it,
>   and say in my reply that I registered a new repo.

**Proposed:**
> - If the work genuinely needs a codebase Orca has not seen, registering it is
>   fine — **but `orca repo add` has never been run on cooper and its flags are
>   unverified**: run `orca repo --help` and `orca repo add --help` first, use
>   what they actually say, and report in my reply both that I registered a repo
>   and what the real command turned out to be.

### Fix 9 — disambiguate "metarepo"

**Current** (§Picking a repo, first bullet):
> - **Default: `mc-metarepo`** — `--repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2`

**Proposed:**
> - **Default: `mc-metarepo`** — `--repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2`.
>   When Gaetan says "the metarepo" unqualified, he means this one. (His personal
>   `~/dev/gaetan-metarepo` is a different repo and is not registered with Orca —
>   it is not the target of delegated work.)

### Fix 10 — two small hygiene notes

**Current** (step 1): `ssh cooper 'cat > /tmp/tars-brief.md' <<'BRIEF'`
**Proposed:** `ssh cooper 'cat > /tmp/tars-brief-<slug>.md' <<'BRIEF'`
> `#   Unique per delegation — a fixed path is clobbered by a concurrent run or`
> `#   by my own retry after a timeout.`

**Current** (step 6): `worker-read --dispatch "<DISPATCH_ID>" --limit 200 --json`
**Proposed:** append
> `#   200 messages is a lot of transcript. The agent's final report is at the`
> `#   END — read the tail first, and only go further back if the outcome was`
> `#   `failed` and I need to know why.`

---

## Verdict

I would have **spawned the run correctly** and **handled the 300s timeout
correctly** — the two things the skill was most obviously designed to get right,
it gets right.

I would have **failed the sentence Gaetan actually wrote** ("tell me when it's
done"), because nothing in the skill wakes me back up (Fix 2) — and there is a
material chance I never spawn anything at all, because "documentation" does not
appear in the list of things I am told not to produce myself (Fix 1).
