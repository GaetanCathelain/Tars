# Red-team review — INCIDENT REPLAY of the conversation-state fix draft (2026-08-14)

Target: `status/probes/2026-08-14-damien-followup-fix-draft.md` (H1, H2, H3, apply runbook).
Mode: adversarial replay. Read-only. Nothing applied, nothing on the VM touched,
no `sops -d`, no secret read. Citations `draft:N` are line numbers of the fix draft;
`SOUL.md:N` / `SKILL.md:N` are the repo mirrors as of md5
`0f2d817607f813c363fe9338f17edd81` (SOUL.md, 319 lines) and
`4d73d985c52c5929d6831fe8e889d129` (engagement-checker/SKILL.md, 739 lines), both
measured this session.

**Verdict: DO NOT APPLY AS DRAFTED.** The patched text does not deterministically
produce the report Gaetan asked for in scenario A, silently changes a *write*
predicate, dead-ends in scenario B (the case the fix exists for), and its runbook
can destroy a standing correction Tars appended between the draft's stale
measurement and the apply.

---

## 0. What the patched text actually binds

| Clause | draft lines | Fires when |
|---|---|---|
| H1 authoring clause | 58–64 | Tars *compiles* a cron/schedule/reminder about a named person |
| H2 trigger | 97–100 | Tars reports a ticket/task/status involving a named person — answering, **or** delivering a cron/schedule/brief |
| H2 read + state | 100–103 | read DM or channel; state last sender, when, still-waiting; **before** ticket state |
| H2 anchor | 103–105 | anchor on Gaetan's last message, never a `Source:` ts |
| H2 window | 105–109 | explicit `limit`, start `30d`, widen "until I have both" |
| H2 anti-manufacture | 109–111 | "he answered and Gaetan closed it himself" ≠ "he has not answered" |
| H2 escape hatch | 111–113 | unreachable / cannot establish → say so |
| H3 carve-out | 148 | engagement-checker may fetch "the DM between Gaetan and a person named in an item being reported" |

Note for all three scenarios: SOUL.md **is** in a cron run's effective context
(verify.md:29 greps "SOUL.md + linear-ticketing + cron prompt"), and H2's trigger
is explicitly widened to cron deliveries (draft:99–100). So H2 fires at run time on
its own. The draft's stated rationale for H1 — "rule 10 alone never fires on that
path" (draft:26–30) — is no longer true once H2's trigger is widened. What H1 adds
is therefore not reach; it is a change to the job's *content*, which is where its
two defects live (§1.2, §1.3).

---

## 1. Scenario A — replay of the actual 2026-08-13 incident

Ground truth (vm-run-trace:29, 46, 70–75, 79): Damien's last = "done"
`1786633440.634439` (15:04:00 UTC); Gaetan's closing ack "C'est complete de mon
côté" `1786633455.716979` (15:04:15 UTC, last overall); GCN-49 Done since 15:10:28Z;
job authored 17:11–17:13 CEST, run 23:55 CEST.

### 1.1 Authoring step (17:11), under H1

- draft:60–62 "the prompt names the DM between Gaetan and that person" → `D08BJ8CQLP6`.
  In the real run this took two live human corrections (vm-run-trace:33–36); H1
  states the requirement but adds no resolution or verification step (see §3).
- draft:62–63 "anchors on Gaetan's last message in it — never a ticket's `Source:` ts"
  → the anchor becomes `1786633455.716979` instead of `1786633197.554209`. **This is
  the one clause that fixes the incident's stated defect.**
- draft:63–64 "it requires the delivery to state conversation state as rule 10 does".

### 1.2 The predicate flip (defect, not a fix)

The incident job was not only a report; it carried a **write branch**
(vm-run-trace:46): *"S'il n'y a aucun autre message de Damien aujourd'hui, mets
GCN-49 dans l'état Done… Si Damien a envoyé un autre message aujourd'hui, ne modifie
pas GCN-49."*

H1 does not scope the anchor to the reporting half of the prompt — draft:61–62 says
the *prompt* anchors on Gaetan's last message. Re-running the job with the new
anchor:

- old predicate: "Damien posted after `1786633197`?" → **yes** → no write, ticket left alone.
- new predicate: "Damien posted after Gaetan's last message `1786633455`?" → **no** →
  branch taken → `save_issue(GCN-49, Done)`.

Here that write is harmless (already Done). In general it is not: "Gaetan spoke last"
is the *normal* end state of any exchange, so the patched anchoring makes the silence
branch fire on nearly every "close it unless something moves" job. The draft claims
its locus is the report ("A boolean I invented is not a state I read", draft:64) but
the text it adds moves the closing predicate too, with no line saying the anchor
governs the narration only. **Blocker.**

### 1.3 Delivery step (23:55), under H2 — two legal outputs, opposite meanings

The run reads `D08BJ8CQLP6` (H2 draft:100–101), with `limit=30d` (draft:105–107),
and holds both facts: Damien "done" 15:04:00, Gaetan ack 15:04:15.

*Output A1*, following the anchor clause (draft:103–105) — measure from Gaetan's
last message:

> Depuis ton dernier message (13/08 17:04, « C'est complete de mon côté »), Damien n'a
> pas répondu. GCN-49 est en Done depuis 17:10.

*Output A2*, following the anti-manufacture clause (draft:109–111) — "he answered and
Gaetan closed it himself" and "he has not answered" are different findings and I never
turn one into the other:

> Damien a répondu « done » à 17:04:00 ; tu as clôturé toi-même 15 s plus tard. Rien
> n'attend de réponse de sa part. GCN-49 est en Done depuis 17:10.

Both are literally compliant. A1 is the sentence Gaetan asked for
("Damien didn't answer since your last message, ticket switched to done") and is the
exact framing draft:109–111 was written to forbid — verify.md:125–133 records that
Gaetan's premise was **false** for this exchange. A2 is true but is not the sentence
he asked for. The patched rules do not decide between them: whether Tars manufactures
the silence Gaetan expects, or refuses to, is left to sampling. A rule that permits
both readings of the very incident it was written for has not fixed the incident.
**Blocker.**

Would A2 satisfy Gaetan anyway? Probably yes — it carries who spoke last, when,
whether anything is pending, and the ticket state, which is the information he was
missing. But that is an argument for deleting the anchor-framing clause (draft:103–105)
in favour of "state who spoke last and whether anything is pending on either side",
not for shipping both clauses and hoping.

### 1.4 Secondary A-defects

- "whether it is still waiting on an answer" (draft:102) has **no criterion**. The
  incident turns entirely on "a closing ack is not an open ask" (vm-run-trace:79),
  and the draft never says so. Same clause now also gates the auto-close of §1.2.
- Ordering mandate "before I state the ticket's state" (draft:102–103) collides with
  the Every-turn rule "The outcome is my first line. Five lines or fewer"
  (`SOUL.md:77–79`) whenever the outcome *is* the ticket state.

---

## 2. Scenario B — genuine silence (Gaetan asked 5 days ago, zero reply)

Tool semantics of record (vm-run-trace:93): `limit` is a **string**, default `"1d"`,
"format of maximum ranges of time (e.g. 1d, 1w, 30d, 90d — free-tier cap) **or**
number of messages (e.g. 50). Must be empty when `cursor` is provided."

Step-by-step against the draft:

1. draft:105–107 "I pass an explicit `limit` … starting at `30d`" → the `1d`-default
   trap (the one the fix names) is closed. ✔ This is the draft's real win.
2. draft:107–109 "widening until I have both [the last message of each side]" — here
   it breaks:
   - Gaetan's last message (5 d ago) is inside 30d. ✔
   - Damien's last message may be **months** old, or may not exist (first outreach).
     The terminator is "until I have both", so the run is instructed to widen past
     30d → 90d → and then it has nowhere to go: 90d is the documented cap
     (vm-run-trace:93), and on a free-tier workspace history beyond it does not
     exist. The draft never states the cap and never states a stop rule.
   - The only exit left is draft:111–113 "if I cannot establish who spoke last, I say
     that" — i.e. the fix routes its own flagship case into "I could not establish
     it", instead of the correct and available report: *Gaetan asked on 09/08, no
     reply in 30 d.* The loop terminator is simply the wrong one: what the report
     needs is Gaetan's last message plus "did the counterparty post after it", not
     both parties' last messages. **Blocker.**
3. Where it can still read empty and narrate anyway:
   - The 30d window returns messages but the server truncates. 1 day of this DM was
     already 3,789 chars (vm-run-trace:57); 30 days of an active DM is far more.
     Nobody has measured whether the MCP server truncates, paginates, or in which
     order it returns — and the draft's own widen instruction is incompatible with
     the documented pagination path (`limit` "must be empty when `cursor` is
     provided", vm-run-trace:93). If truncation drops the oldest end, Gaetan's 5-day
     -old ask disappears and Tars reports on whatever survived, with no signal that
     anything was dropped. The repo's own hard rule ("measure before believing recon
     claims", CLAUDE.md) is not satisfied for the single mechanism this fix rests on,
     and the runbook adds no measurement. **Major.**
   - A person Gaetan talks to only in a channel: H2 says "the DM **or channel**"
     (draft:100–101) with no rule for *which* channel. An empty DM read plus an
     unbounded channel choice is exactly where an invented answer enters.

---

## 3. Scenario C — no Gaetan↔person DM (or the token cannot see it)

- **Run time is covered, textually.** draft:111–113 mandates "if the conversation is
  unreachable … I say that in the report instead of dropping it". An empty/erroring
  `conversations_history` therefore yields a stated non-answer rather than an
  omission. ✔ (Credential-wise the token is Gaetan's own xoxc/xoxd session,
  vm-run-trace:92, so visibility ≈ Gaetan's — a DM that Gaetan never opened simply
  does not exist rather than 403s.)
- **Authoring time is not covered.** H1 draft:60–62 makes "the prompt names the DM
  between Gaetan and that person" an unconditional element of the compiled job, with
  no fallback when there is no such DM and no step for resolving the id. The incident
  itself is the proof this step is failure-prone (two live corrections,
  vm-run-trace:33–36). A job compiled with a guessed or absent channel id fails at
  23:55 with nobody watching — and H1 offers no authoring-time equivalent of the
  say-so clause (e.g. "if I cannot identify the conversation, I say so before
  scheduling the job, or I anchor the job on the ticket alone"). **Major.**
- Second-order: H3's carve-out (draft:148) permits fetching *the DM*; finding *which*
  DM (a users/conversations lookup) is not history-fetching so it is not blocked —
  but H2's "or channel" case is blocked, see §4.

---

## 4. Cross-cutting defects

1. **H3 is narrower than H2 requires.** H2 mandates "the DM **or channel** between
   Gaetan and that person" (draft:100–101); H3 carves out only "the DM"
   (draft:148). Under engagement-checker the channel read stays banned, so the
   contradiction the draft claims to remove survives for every person Gaetan only
   talks to in channels.
2. **H3 patches one of two bans.** `engagement-checker/SKILL.md:112` ("Normal runs
   must process only new events … **Do not rescan the whole day, inbox, Slack
   history**, or all Linear issues for context") is the stronger, run-level ban and
   is untouched — skill-audit.md:66 (gap 6) named both caps; the draft addresses
   only :129.
3. **Blast radius of H2's trigger.** "Whenever I report on a ticket, a task or a
   status that involves a named person other than Gaetan" (draft:97–99) fires on
   every brief, kanban render and status line that names a colleague: N people ⇒ N
   `30d` history reads per report, N conversation-state sentences ahead of the
   verdict. Risk note 2 (draft:276–279) counts "one extra call per person-scoped
   report" — an undercount by a factor of N — and does not mention that the mandated
   ordering fights `SOUL.md:77–79`. It also loosens, for every named person, the one
   control that currently keeps colleague DM content out of context; that is a
   deliberate trade, but it should be stated as one.
4. **Contradiction audit was run over a partial mirror.** verify.md:60 measured 5 of
   55 live skill dirs mirrored. Sibling "don't widen history" bans in the 50
   unmirrored VM skills are unaudited, so H3 cannot be claimed to have removed the
   conflicts — only the one visible in the mirror.

---

## 5. Runbook (RB)

1. **No STOP is defined for a `SOUL.md` mismatch.** draft:170–172 reads "Both pairs
   must match. `SOUL.md` matched on 2026-08-14 (verify §B). **If the `SKILL.md` pair
   differs, STOP**" — the explicit stop covers the skill only, and the sentence
   before it pre-announces the SOUL pair as already matching. That measurement is
   **stale and provably so**: verify.md was committed 2026-08-14 07:58 UTC with VM
   md5 `c32c3a82…`; at 11:46 UTC commit `cbc504a` ("SOUL standing correction: use
   Tars gateway for outbound Slack (#65)") added line 319, and repo SOUL.md now
   hashes `0f2d8176…`. The draft asserts "no repo-side change since" (draft:9–13),
   which is false. SOUL.md has a **live append-only writer** (Tars, `SOUL.md:299–313`
   + rule 2's flow) that lands corrections the same day they are given; step 2's
   whole-file `cat SOUL.md | ssh … cat > ~/.hermes/SOUL.md` overwrite therefore
   **deletes any `## Standing corrections` line appended between the operator's pull
   and the transfer** — precisely the destruction rule 2 exists to prevent
   ("the next person who reconciles the file silently destroys it", `SOUL.md:122–123`).
   Required before apply: a hard STOP on SOUL mismatch, a re-check of both md5 pairs
   immediately before step 2, and a tail-merge procedure (rebase the hunks onto the
   VM copy) rather than STOP-only. **Blocker.**
2. **The `flock` does not serialize the writer it needs to.** `~/.hermes/.wf3.lock`
   is the config-edit convention; `skill_manage` and rule 2's SOUL append do not take
   it, so step 0 → step 2 is a plain TOCTOU window of minutes to hours for both
   files.
3. **Verification proves loading, not behaviour.** Step 2's probe (draft:236–238)
   asks Tars to *quote* the new paragraph. That is a recitation test: it cannot
   distinguish output A1 from A2 (§1.3), cannot exercise scenario B's widen loop, and
   cannot show the write predicate of §1.2. The apply ships with zero behavioural
   check; at minimum, one scheduled-job dry authoring plus one silent-counterparty
   read should be exercised before this is called done.
4. Minor mechanics: `$STAMP` in the rollback line (draft:267) is dead in any new
   shell — pin the backup names in the report; `hermes chat` (draft:238) is used
   without the `--help` check CLAUDE.md mandates for non-interactive invocations; if
   step 1's `ssh` for `SOUL_MODE` fails, step 2's `chmod $SOUL_MODE` runs empty
   *after* the `mv`, and the "must be non-empty" note (draft:179) is prose, not a guard.
5. Not a defect, verified: line arithmetic is right (H1 +6, H2 +18 ⇒ 343), both
   BEFORE hunks match the current repo file byte-for-byte (SOUL.md:28–33, 262–267),
   indentation matches, and the H3 diff is exactly the one sentence claimed.

---

## 6. Smallest change that would survive this replay

Not a request, stated once for the orchestrator: drop the anchor-framing sentence
(draft:103–105) down to "never a ticket's `Source:` ts"; replace the terminator
(draft:107–109) with "wide enough to cover Gaetan's last message and everything
after it, `30d` then `90d` (the cap); if his last message is not in a 90d window I
say so"; scope H1's anchor to the *narration* and leave any close/no-close predicate
to Gaetan's own words; add the ack-vs-open-ask sentence; extend H3 to "the DM or
channel" and to `SKILL.md:112`; add the SOUL STOP + re-check + tail-merge to step 0.
