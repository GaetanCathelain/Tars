# Skill audit: Damien follow-up reporting gap (2026-08-14)

Slack permalink under investigation: channel `D0BBYNM01BL`, ts `1786658157.729519`
(2026-08-13 21:55:57 UTC). Per SOUL.md:273 `D0BBYNM01BL` is named as **the
Gaetan<->Tars home DM** ("Everything I initiate ... lands as a NEW top-level
message in our DM (`D0BBYNM01BL`)"). So the linked message is in the
Gaetan<->Tars DM, i.e. it is very likely **Tars' own deficient report**, not
the Gaetan<->Damien conversation (which is a separate DM channel this audit
did not have access to enumerate). Established from SOUL.md text, not assumed.

## Files read

- `SOUL.md` (318 lines, full)
- `skills/orchestration/linear-ticketing/SKILL.md` (full)
- `skills/orchestration/engagement-checker/SKILL.md` (partial, lines 1-536 of 740 — covers intake, classification, permitted writes; §8 reporting/rendering section at the tail was not reached)
- `skills/orchestration/daily-work-brief/SKILL.md` (grepped)
- `skills/orchestration/slack-channel-context/SKILL.md` (grepped)

## (1) Does any rule tell Tars to read the full conversation and establish who replied last before reporting on someone's feedback?

Partial, and narrowly scoped — not a fit for the observed failure:

- `SOUL.md:253-265` (rule 10): "I read the record before I answer for the past... when the question turns on what happened — who did what, what was sent where, why something is in the state it is — I fetch before I answer: the thread I was pinged in (`conversations_replies` on its `thread_ts`), the last ~20 messages of the channel or DM I was pinged on (`conversations_history`)..."
  - This rule triggers on **questions about the past** arriving as an inbound Slack ping. It is not framed as a standing pre-report check that fires proactively when Tars itself decides to narrate a ticket status change.
  - Critically, the fetch scope is **"the thread I was pinged in" and "the channel or DM I was pinged on"** — i.e. wherever the triggering message landed. If Tars was pinged in the Gaetan<->Tars DM (or by a cron/Linear delta) and the actual unanswered exchange lives in the separate Gaetan<->Damien DM, rule 10 as written does **not** direct Tars to go open that other channel. There is no instruction anywhere to "identify the stakeholder named in the ticket, find the DM/channel with them, and read it."
- `linear-ticketing` §12 ("Delegated ticket work", lines 216-224) governs reporting when Gaetan asks Tars to *work on* a ticket (state transitions, provenance, outcome comments) — it says nothing about consulting Slack history with a named stakeholder before reporting a status change.
- No skill contains logic like "before reporting a ticket switched to Done/blocked on someone, check who sent the last message in the relevant DM."

## (2) Any guidance capping or shortcutting Slack history reads?

Yes, multiple caps, all of which would truncate or misdirect a "did Damien reply" check:

- `SOUL.md:257-259` (rule 10): fetch is capped to "the last ~20 messages of the channel or DM I was pinged on" — a fixed small window, not "read until you find the last reply from each party."
- `SOUL.md:19-21`: on a cold Slack thread start, Tars is handed "the thread's ROOT message plus only the last ~30 replies," and "long transcripts get compacted" — so an older unanswered ask could already be outside the visible window even if Tars did look.
- `engagement-checker/SKILL.md:129`: "For a new candidate only, use `conversations_replies`... to recover enough thread context... Do not fetch unrelated channel history." — explicitly forbids widening beyond the specific thread, which would block following a reference from a Linear ticket into a different Slack DM.
- `engagement-checker/SKILL.md:108-116` (§1): "Normal runs must process only new events plus the compact pending queue. Do not rescan the whole day, inbox, Slack history, or all Linear issues for context." — this is a cron-delta-scan rule, but it reinforces the shallow-by-default posture: the codebase's default mode for Slack is narrow, incremental windows, never "read the full conversation."

## (3) What does linear-ticketing say about reporting status changes?

`linear-ticketing` §12 "Delegated ticket work" (lines 216-224) is the only status-reporting guidance in that file:

1. Move to In Progress before starting, verify via §10.
2. Record delegation provenance (worker/session, start time) in comments/description.
3. "Comment again with the verified outcome or blocker when the delegated session finishes. Never report the ticket work as started unless both the state transition and provenance write are confirmed."

This is entirely about Tars's own delegated-work provenance (state transitions it made, sessions it dispatched) — it says nothing about synthesizing *human* conversational context (e.g., "Damien never replied") into a status narrative. There is no field, no step, no rubric anywhere in linear-ticketing for attaching human-response context to a status report.

`engagement-checker` §5 (Classify and reconcile, lines 508-530) is the closest thing to "did someone respond" logic, but it is scoped to Gaetan's own open loops (his commitments/asks), not to reporting whether a third party (Damien) responded to Gaetan. Its Linear-terminal-state handling (lines 522-530) treats a ticket moving to Done purely as a state-type fact, with no instruction to also narrate the human Slack exchange behind it.

## (4) Any rule that would plausibly produce the observed shallow answer?

Yes — this is the most direct fit:

- **`SOUL.md` rule 10 fetch scope is wrong-channel by construction.** It fetches "the thread I was pinged in" / "the channel or DM I was pinged on" — never the channel a *named third party* (Damien) actually corresponds in. A report about a ticket "involving Damien" would need Tars to (a) recognize Damien as a stakeholder, (b) locate the Gaetan<->Damien DM specifically, and (c) read enough of it to see who sent the last message. No rule anywhere instructs step (a) or (b); rule 10 only reads the channel of the *inbound* message.
- **Fixed ~20-message cap (SOUL.md:258-259) and the "do not fetch unrelated channel history" instruction (engagement-checker:129)** together bias toward a shallow, single-channel read even when rule 10 does fire.
- **Dangling `stakeholder-communication` skill reference.** `daily-work-brief/SKILL.md:18` says "Load `stakeholder-communication` before synthesis" — but no `stakeholder-communication` skill exists anywhere under `skills/` in this mirror (`find skills -iname "*stakeholder*"` returns nothing, and no other file defines or mirrors it). This is exactly the kind of skill that would plausibly carry "read the full thread, determine who spoke last" logic for reporting on a colleague's feedback — and it is missing. Whether it is missing from the live VM too, or only from this git mirror, is unverified from this audit alone (see gap list).
- `linear-ticketing` §12's reporting step ("comment again with the verified outcome") is itself only about Tars's own delegated work, so even a diligent read of that skill gives no instruction to check Slack for Damien's silence before writing "ticket switched to done."

## Gap list — instructions that SHOULD exist but do not

1. **No rule ties a Linear ticket status report to the Slack conversation with the specific person(s) referenced in the ticket.** Nothing says "when a ticket involves a named person, before reporting status, locate and read the Gaetan<->that-person conversation and determine who sent the last message."
2. **No "last-speaker" / unanswered-message check exists anywhere.** Neither SOUL.md nor any skill computes "did the other party reply after Gaetan's last message" as a fact to surface. `engagement-checker` computes "did Gaetan reply" (for his own open loops) but never the mirror case (did the counterparty reply to Gaetan).
3. **`stakeholder-communication` is referenced but not present in the mirror** (`daily-work-brief/SKILL.md:18`). If this skill is where "read full history, identify who replied last" logic was meant to live, its absence (from this repo, and possibly from the live VM) is a direct, mechanical explanation for the shallow report.
4. **Rule 10's fetch scope is keyed to the inbound message's channel, not to entities named in the content being reported on.** No fallback like "when the report concerns a named third party, also check the DM/channel with them" exists.
5. **No guidance distinguishes "proactive status narration" (Tars deciding to report a ticket's status) from "answering a direct question about the past."** Rule 10 is framed around the latter; the Damien case is the former, and nothing covers it.
6. **No cap-override for stakeholder-context reads.** Even if a future rule adds "check the other person's DM," the existing ~20-message / ~30-reply caps and the "do not fetch unrelated channel history" instruction would still need an explicit carve-out to actually walk back far enough to find "Gaetan sent the last message, N days ago, unanswered."
