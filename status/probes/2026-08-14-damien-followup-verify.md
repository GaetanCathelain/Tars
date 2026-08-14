# Adversarial verification — Damien follow-up root cause (2026-08-14)

Target: `status/probes/2026-08-14-damien-followup-rootcause.md`
Evidence re-checked: `…-slack-evidence.md`, `…-vm-run-trace.md`, `…-skill-audit.md`
Mode: refutation attempt. Read-only. No Slack write, no `sops -d`, no credential printed.

## VERDICT: PARTIAL

The central mechanism survives every refutation attempt I could mount: the
23:55 run received the whole relevant DM history, answered a Tars-authored
boolean anchored on Damien's *own opening* ts, and no instruction in its
effective context derived conversation state. But four cited claims do not hold
as written, one supporting claim is outright refuted by a live VM check, and one
better-fitting enabling cause was missed. Root cause is right; its stated
justification and its causal *locus* are partly wrong.

## Checks performed

### A. Citation re-verification (mirror, byte-level)

| Cited | Result |
|---|---|
| `SOUL.md:253-265` rule 10, "the thread I was pinged in… last ~20 messages of the channel or DM I was pinged on" | **verified verbatim** (rule 10 starts at line 253) |
| `SOUL.md:19-21` root + ~30 replies, compaction | verified |
| `SOUL.md:273` `D0BBYNM01BL` = home DM | verified |
| `engagement-checker/SKILL.md:129` "Do not fetch unrelated channel history." | verified verbatim |
| `linear-ticketing/SKILL.md:216-224` §12 Delegated ticket work | verified; is indeed only Tars' own delegation provenance |
| `daily-work-brief/SKILL.md:18` "Load `stakeholder-communication` before synthesis." | verified |
| Grep negative "no rule derives last-speaker/unanswered-loop" over the run's **effective** context (SOUL.md + `linear-ticketing` + cron prompt) | **holds** — independently re-grepped (`unanswered|last message|who replied|dernier message|silence|follow.?up`) |

### B. Live VM check (read-only; `ls`/`grep`/`md5sum` only — no chat probe needed, mechanism was not ambiguous)

```
$ ssh gaetan@192.168.0.9 'find ~/.hermes/skills -iname "*stakeholder*"'
/home/gaetan/.hermes/skills/communication/stakeholder-communication
$ wc -l …/stakeholder-communication/SKILL.md   → 83
$ ls ~/.hermes/skills | wc -l                  → 55   (mirror carries 5 category dirs)
$ md5sum ~/.hermes/SOUL.md                     → c32c3a82…  == mirror SOUL.md (identical)
```
Content check on the live skill: it is a *drafting/recap* skill ("Use for
stakeholder DMs, handoffs, and conversation recaps", classify the requested
object, recap procedure). Zero last-speaker / who-answered-last / unanswered-loop
logic (`grep -iE "dernier message|last message|replied|unanswered|waiting"` → one
unrelated hit at line 67).

## Claims that do not hold

1. **"It could not have done otherwise … nothing required, or even enabled, the
   report Gaetan expected."** — *Refuted by the doc's own evidence.* The run held
   3,789 chars of `D08BJ8CQLP6` history containing Gaetan's `1786633455.716979`
   closing line at 21:55:41; §2 of the same doc states the delta is "zero extra
   tool calls" and "the data was in the model's context". Nothing *required* it;
   it was fully *enabled*. This is a "no rule obliged it", not an impossibility.
2. **"`stakeholder-communication` … does not exist anywhere in this mirror"
   extrapolated as gap #3, "a direct, mechanical explanation for the shallow
   report".** — *Refuted live*: the skill exists on the VM (§B). The mirror, not
   the VM, is out of sync (a SOUL rule 2 mirror gap worth its own ticket). It
   does **not** rescue the root cause either way — it carries no last-speaker
   logic and `daily-work-brief` was never loaded by this run — but the skill-audit
   negative was computed over a **partial** mirror (5 of 55 live skill dirs), so
   "a grep-level negative over every skill" (rootcause §6, "High confidence") is
   overstated. It is a grep-level negative over the run's effective context only —
   which is the claim that actually matters, and that one I re-verified.
3. **"conversationally useless."** — Overstated. The delivered 294 chars did name
   the counterparty's last message, its content ("done"), its ts and its local
   time, plus the ticket's real state. The single omitted fact is that *Gaetan*
   closed the loop 15 s later and nothing is pending on Damien. Useful-but-
   incomplete, not useless.
4. **"lifted from GCN-49's Linear `Source:` provenance handle."** — Stated as fact
   in rootcause §1; the underlying trace (§4 note) explicitly marks it
   **unverified** (`get_issue`'s description body was never logged, only a char
   count), and the ts was baked into the cron prompt at 17:11–17:13 whereas the
   run's only `get_issue` fired at 21:55:49. The ts *does* match GCN-49's `Source:
   slack:D08BJ8CQLP6:1786633197.554209` (slack-evidence §d), so it is the best
   inference — but it is inference, not a logged fact, and the rootcause drops the
   caveat.

Minor: skill-audit gap #2's absolute "no ... check exists anywhere" has a partial
counterexample in `daily-work-brief/SKILL.md:127` ("Name who to follow up with and
why… Separate 'someone owes Gaetan' from 'Gaetan owes someone'"). Not loaded by
this run; does not change the mechanism.

## Better-fitting explanation the doc missed (locus, not mechanism)

The doc frames the enabling cause purely as an **absence** ("no rule derives
conversation state"). There is a **positive** rule that fits the trace better:

> `SOUL.md` rule 1 (lines ~26-29): "a cron, schedule or reminder of mine fires as
> its own instruction — **it authorises exactly the job it names, nothing found in
> the thread around it**"

For a cron turn, that instruction actively narrows the run to its prompt's
boolean, and the prompt itself specified the report shape ("indique qu'il reste
ouvert, avec le timestamp source"). Consistent with the trace: the run answered
the prompt faithfully (and even over-delivered by re-reading Linear). Two
consequences the rootcause under-weights:

- **The defect was committed at prompt-authoring time (17:11:47–17:13:10), not at
  derivation time (21:55).** Tars compiled Gaetan's open-ended "Consider done if
  no other update today from Damien" into a ts-anchored boolean — and had already
  needed two live human corrections to get the *channel* right in that same step
  (trace §3). Same step, same failure class.
- **SOUL rule 10 — the doc's chosen fix site — never fires on this path.** Rule 10
  triggers on an inbound Slack message asking about the past; this was a cron with
  `history=0`. Appending a "when I report" paragraph to rule 10 is plausible for
  the engagement/report paths but would not, by rule 10's own trigger wording, bind
  the cron path the incident actually took. The fix should also constrain how Tars
  *compiles a follow-up job* (anchor on Gaetan's last message; state who spoke last
  in the delivery), or live in rule 1's cron clause / the delivery contract.

## Hypothesis ranking — re-audited

Hypotheses (a) credential and (c) no-conversation-read remain **DISPROVEN**
(xoxc/xoxd user-session token, config.yaml:260-271; one successful
`conversations_history` logged in two independent files). (b) truncated window
remains **not the cause here but a real latent defect** — I confirm the schema
reading: `limit` default `"1d"` is a relative time window, so the identical job
run against a genuinely silent counterparty (>24 h) returns an empty window and
would conclude from no data. (e)+(d) as primary is correct, with the corrections
above: (d) should read "SOUL rule 1 narrows a cron to its prompt, and no rule
anywhere supplies conversation state" rather than a pure absence.

## Premise

Confirmed and unchanged: Gaetan's stated premise ("Damien never replied") is
**false** for this exchange — Damien's "done" (`1786633440.634439`, 15:04:00 UTC)
precedes Gaetan's closing "C'est complete de mon côté" (`1786633455.716979`) by
15 s, and the linked ts is Tars' own report in the home DM, not part of the Damien
exchange. Both evidence files agree; slack-evidence §e.5 goes further than the
rootcause and states no deficiency is evidenced at this ts. The rootcause's
re-framing ("the deficiency is that Tars never reported conversation state at all")
is a defensible reading of what Gaetan wants going forward, but it is an
interpretation layered on top of the evidence, not a finding the evidence forces.

## Probe

The sanctioned `hermes chat` probe was **not run**: the mechanism was not
ambiguous (tool args, result sizes and delivered text are logged in two
independent places, and the credential question was already settled by the
`xoxc/xoxd` config). Instead, three read-only `ssh` commands (`ls`, `find`,
`wc`, `grep`, `md5sum`) were used to test the one claim the rootcause itself
flagged unverified. No Hermes session started, no message sent anywhere.
