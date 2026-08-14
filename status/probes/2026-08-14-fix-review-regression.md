# Red-team review — REGRESSIONS & SIDE-EFFECTS lens (Damien follow-up fix draft, 2026-08-14)

Target: `status/probes/2026-08-14-damien-followup-fix-draft.md` (H1/H2/H3/RB).
Mode: adversarial, read-only. Nothing applied, nothing touched on the VM, no
`sops -d`, no secret read. Files read in full: the fix draft, `SOUL.md` (319 l),
`skills/orchestration/engagement-checker/SKILL.md` (740 l), the rootcause,
verify, vm-run-trace and skill-audit probes; `daily-work-brief/SKILL.md` and
`slack-channel-context/SKILL.md` grepped for the read-ban / format sites H2
touches.

Verdict: **the diagnosis is sound; the hunks as worded buy new failures.**
Three blockers, six majors, three minors. None of them argue for keeping the
current behaviour — they argue that H2's trigger and H3's placement are far
wider than the incident, and that the runbook races a file Tars writes itself.

Measured baselines used below (this session):
`md5sum SOUL.md` = `0f2d817607f813c363fe9338f17edd81`, 319 lines;
`md5sum skills/orchestration/engagement-checker/SKILL.md` = `4d73d985c52c5929d6831fe8e889d129`.

---

## 1. Cost / latency creep (question 1)

**The draft's own risk note (fix-draft:276-279) says "one extra Slack
`conversations_history` call per person-scoped report". That is the floor, not
the estimate.** H2's trigger (fix-draft:97-101) is *per named person*, per
report, and fires on the cron/brief path as well as the answering path:

- `engagement-checker` delivers **up to six bullets** per delivering run
  (SKILL.md:700) and every Slack-sourced item carries `people`/`asker`
  (SKILL.md:80). Jobs run `*/30 10-16` plus `0 17`, weekdays — **15 runs/day**
  (SKILL.md:732-733). A delivering run under H2 owes up to 6 wide DM reads.
- `daily-work-brief` names people by construction ("Name who to follow up with
  and why… Separate 'someone owes Gaetan' from 'Gaetan owes someone'",
  daily-work-brief:127) — one brief can name a dozen.
- Every ad-hoc question naming a colleague adds one more.

Each read is a **30d window** (fix-draft:105-107) against a tool whose measured
1d window on a two-person DM already returned 3,789 chars
(vm-run-trace:57, :93). Two mechanical multipliers the draft misses:

1. **Widening is a full re-download, not a continuation.** The cached schema
   states `limit` "Must be empty when `cursor` is provided"
   (vm-run-trace:93) — so "widening until I have both" re-issues the whole read
   at 90d, not a paged continuation of the 30d one.
2. **The wide read can exceed Hermes' own tool-result budget**, the failure the
   skill already documents in another context: "an unprojected large read blows
   Hermes' own tool-result budget and never reaches the agent"
   (engagement-checker:617). `conversations_history` has no field projection.
   The degradation is silent — the model reports conversation state it never
   received.

**No stop condition, and a hard ceiling it cannot see.** The same schema line
enumerates `1d, 1w, 30d, 90d — free-tier cap`. For a person who has never DM'd
Gaetan — a colleague he only meets in channels, an email-sourced `asker`, a
vendor — "widening until I have both" (fix-draft:106-108) never terminates on
its own, and it cannot get past 90d. H2 also never tests whether a DM exists
before mandating the read. The natural brake, stop-loss ("a second research
round that surfaced no new fact… I stop", SOUL.md:64-70), is an "Every turn"
bullet, and SOUL.md:14-16 says **the hard rule wins on collision** — so H2, a
hard-rule paragraph, out-ranks the only bound in the file.

## 2. Scope creep and the rule-4 interaction (question 2)

H2 is written as "I read the DM or channel between Gaetan and that person…and I
state who sent the last message, when, and whether it is still waiting on an
answer" (fix-draft:97-104). **It never says where that sentence may be
delivered.** SOUL rule 4's second paragraph (SOUL.md:217-228) lets Tars post
substantive content into a group DM or channel that includes Gaetan when the
specific message is his call.

Failure scenario: Gaetan tells Tars to post the GCN-49 status into a group DM
with Damien and Alice. H2 now *requires* the delivery to open with the state of
Gaetan's **one-on-one** DM with the named person — "tu n'as pas répondu à Alice
depuis le 2 juillet" — disclosed to Damien. Gaetan approved a status post, not
the timing metadata of his private 1:1s. The hunk needs "…in our DM" scoping (or
"never in a conversation that includes anyone but Gaetan"); it has none, and
the standing correction at SOUL.md:319 is a reminder that this Slack surface is
Gaetan's **personal** account, so the DMs H2 reaches are not all work DMs.

Nothing in SOUL.md forbids the *read* — the xoxc/xoxd credential has Gaetan's
own visibility (vm-run-trace:92) — so the new exposure is purely on the
delivery side, which is exactly where H2 is silent.

## 3. H3 — what line 129 was protecting (question 3)

Line 129 sits in **§2 "Collect Slack deltas"**, the per-candidate
classification paragraph, and it protects §1's cursor discipline one section
above: "Normal runs must process only new events plus the compact pending
queue. **Do not rescan the whole day, inbox, Slack history**, or all Linear
issues for context" (engagement-checker:112). The ban is a cost and
idempotence guard for a cron with a tight runtime, not a privacy guard.

Two regressions from the one-line carve-out (fix-draft:148):

- **Wrong section for the intent.** The carve-out's wording is about "an item
  being **reported**" (§8, ≤6 bullets), but it is planted in the paragraph that
  processes **candidates** — §2 retains up to 100 Slack events per run
  (SKILL.md:607). An implementer reading §2 top-to-bottom licenses one 30d DM
  read per candidate, 15 times a day. That is the cost blowup in §1 above,
  multiplied by ~16.
- **The file's own precedent is bounded and H3 is not.** Two lines below, the
  existing widening carve-out reads: "widening with **one bounded**
  `conversations_history` read for that DM **only when** the handoff cannot
  otherwise be resolved" (SKILL.md:131). H3 copies neither the "one", the
  "bounded", nor the "only when".
- **The contradiction survives the fix.** H3 amends line 129 but leaves line
  112's identical ban intact, so after the apply the skill still tells Tars
  both things. The fix is incomplete on its own terms.

Minor, lower confidence: a widened read returns messages **older than the
cursor**, and §1's exception ("unless its stable event ID is absent and it can
close or mutate a pending item", SKILL.md:110) is a door for pre-cursor content
to mutate items or seed a filing. The `seen[]`/cursor idempotence argument at
SKILL.md:116 was written assuming reads never reach behind the cursor.

## 4. Report bloat (question 4)

**Blocking conflict with the daily brief.** `daily-work-brief` mandates the
Linear board block as "the **first element** of the brief", pasted "byte for
byte", with "You narrate **around** the block, never inside it"
(daily-work-brief:139), and caps the prose at 350 words (daily-work-brief:159,
:142). H2 mandates conversation state "**before** I state the ticket's state"
(fix-draft:101-103). Hard rules override skills (SOUL.md:14-16, 87-89), so
every morning Tars must either push per-person preambles ahead of the frozen
board block or violate a hard rule. Neither hunk resolves it; nothing in the
draft mentions `daily-work-brief` at all, and its existing follow-up sentence
(daily-work-brief:127) already covers part of what H2 asks for.

**Format conflict inside engagement-checker.** Its delivery format is one line
per item — `• What — who; why now. Next: one concrete action. [GCN-42]`
(SKILL.md:704). "Who spoke last, when, and whether it is still waiting on an
answer" does not fit that line and must precede the ticket handle. H2 silently
rewrites a format the skill spells out.

**The trivially-closed case is the common case.** For the incident itself the
mandated preamble is "Damien a répondu « done » à 17:04, tu as clos toi-même à
17:04:15" — useful once. On a ticket whose loop closed weeks ago it is a stale
sentence in front of every mention, against "Verdicts, not logs. The outcome is
my first line" (SOUL.md:77-79). H2 has no "state it only when the loop is
open or the state is surprising" escape.

## 5. Would H1 have prevented the wrong-channel failure? (question 5)

**Partly — the naming half, not the verification half.** H1 says the prompt
"names the DM between Gaetan and that person" (fix-draft:60-62); Tars' first
draft named "notre DM" (vm-run-trace:32), so H1 does bite on that exact
authoring step.

Two gaps remain:

- **It binds only jobs about a named person.** A follow-up compiled about a
  ticket, a repo, a PR or a deploy — no person named — keeps the old failure
  mode intact.
- **H1 and H2 disagree about where a conversation lives.** H1 mandates "the DM
  between Gaetan and that person" (singular, DM only); H2 says "the DM **or
  channel**" (fix-draft:100). For a colleague Gaetan only talks to in a project
  channel, H1 forces an anchor on a DM that may not exist, and the job is
  authored against nothing. Two hunks landing in the same file must not
  disagree on the target of the same read.

## 6. Runbook (RB)

**RB-a — the pre-flight gate races a file Tars writes itself, and the draft's
own premise is already stale.** SOUL.md is appended to by Tars via
`printf … >> ~/.hermes/SOUL.md` (SOUL.md:185) — **that append takes no flock**,
so `flock` around the `mv` (fix-draft:209) protects nothing against it. The
runbook hashes both sides once at step 0 (fix-draft:161-172) and then edits,
transfers and overwrites with no re-check. This is not hypothetical: the draft
claims "no repo-side change since" (fix-draft:12), yet `cbc504a` ("SOUL
standing correction: use Tars gateway for outbound Slack (#65)",
2026-08-14 11:46:45 UTC) landed a new line 319 after verify.md's baseline
`c32c3a82…` — which is why the repo now hashes `0f2d8176…`. Failure scenario:
Gaetan corrects Tars in Slack during the apply window, Tars appends the dated
line to `~/.hermes/SOUL.md` and merges its mirror PR, and step 2's `mv`
silently destroys the live line. Fix: re-`md5sum` the VM file **inside** the
`flock` and abort on mismatch.

**RB-b — the runbook performs, from a Claude Code session, exactly the work
SOUL rule 13 reserves to Tars, and pushes straight to main.** Rule 13
(SOUL.md:279-286): work on skills and this file "I run it myself… never through
a Claude Code or Orca session". Rule 2 (SOUL.md:190-204) requires branch → PR →
squash-merge for a SOUL change; step 3 does `git push origin HEAD:main`
(fix-draft:251). The draft names this in risk 3 (fix-draft:280-284) and then
does it anyway. At minimum the repo half should go through a PR so the mirror
history stays uniform with the flow Tars itself must follow.

**RB-c — nothing re-audits the jobs already installed.** H1 constrains *future*
job authoring only. "Done means verified" requires listing every mechanism —
"config, each cron or scheduled job, the skill that fires it, memory"
(SOUL.md:41-47) — and SOUL.md:309-313 says a running session keeps the prompt
it started with. After this apply, every already-scheduled follow-up keeps its
wrong-anchored boolean and keeps delivering it, while the fix reads as landed.
The runbook has no `cron` enumeration step and no re-author pass.

**RB-d (minor) — the engagement-checker transfer can silently revert a
`skill_manage` edit if step 0 is skipped or its output misread.** The draft
says so itself (fix-draft:273-275); worth restating as a hard STOP with the
expected hashes written down, since only 5 of 55 live skill dirs are mirrored
(verify §B).

**RB-e (minor) — `chmod $SOUL_MODE` runs after the `mv`, not before.** If step
1's `stat` capture came back empty (ssh hiccup), the guard is a human reading
an `echo` line (fix-draft:179); the `mv` still happens and `chmod` fails
afterwards, leaving the file at the `umask 077` mode rather than the captured
one. `[ -n "$SOUL_MODE" ] || exit 1` costs one line.

---

## What would make these hunks safe (not a rewrite, the smallest edits)

- H2: scope the delivery to Gaetan's DM; bound the widening ("30d, then 90d
  once, then say I could not establish it"); make the read conditional on a DM
  existing; state conversation state **only when the loop is open or the ticket
  state disagrees with it**, so the daily board keeps its first position.
- H3: move the carve-out to §8 (or the reporting path), and copy line 131's
  shape — *one* bounded read, *only when* the item is being delivered. Amend
  line 112 in the same edit or the contradiction survives.
- H1: say "the DM or channel between Gaetan and that person", matching H2.
- RB: re-hash inside the flock; add the cron re-audit step; PR the repo half.
