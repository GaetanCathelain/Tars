# Root cause — Damien follow-up report (GCN-49), 2026-08-14

Synthesis over three evidence artifacts (all read in full):

- `status/probes/2026-08-14-damien-followup-slack-evidence.md` (Slack + Linear ground truth)
- `status/probes/2026-08-14-damien-followup-vm-run-trace.md` (VM log/tool trace of the run)
- `status/probes/2026-08-14-damien-followup-skill-audit.md` (SOUL/skill guidance audit)

No fix applied. Read-only synthesis.

## 0. Premise correction (must be read before the root cause)

Gaetan's complaint says Tars should have noticed *"Gaetan sent the LAST message
and Damien never replied"*. Two of the three facts hold, one does not:

| Complaint claim | Evidence | Verdict |
|---|---|---|
| Gaetan sent the last message in the Damien DM | slack-evidence §b (2026-08-13 15:04:15 UTC, "C'est complete de mon côté"); vm-run-trace §6 (same, ts `1786633455.716979`) | **TRUE** |
| Damien never replied | slack-evidence §b: Damien's "done" at 15:04:00 UTC, ts `1786633440.634439`, 15s *before* Gaetan's closing line | **FALSE** — Damien replied; Gaetan's last message is an acknowledgment of that reply |
| Tars failed to report the conversation state alongside the ticket move | slack-evidence §a (verbatim report); vm-run-trace §5 | **TRUE** |

The investigated message is **Tars' own report** in the Tars↔Gaetan home DM
(`D0BBYNM01BL` = `home_channel.chat_id`, vm-run-trace §1; SOUL.md:273,
skill-audit header), not part of the Damien exchange (`D08BJ8CQLP6`).

So the deficiency Gaetan felt is real, but it is **not** "Tars missed that
Damien went silent" — reporting that would have been factually wrong
(slack-evidence §e.2). It is: **Tars never reported the state of the
conversation at all.** It reported the outcome of an internal boolean branch.

## 1. Root cause

Tars answered a **self-generated boolean** — "did Damien post anything after ts
`1786633197.554209`?" — instead of computing and reporting **conversation
state** (who spoke last, to whom, is the loop closed). The anchor ts it tested
against was **Damien's own opening message** ("Bon je suis bloqué",
slack-evidence §b, §d `Source: slack:D08BJ8CQLP6:1786633197.554209`), lifted
from GCN-49's Linear provenance handle (vm-run-trace §4 note), not Gaetan's
last message and not the "consider done" instruction ts. That check is near
tautologically true in any live back-and-forth, so it produced a technically
correct but conversationally useless verdict. It is useless *by design*,
because **no rule in SOUL.md or any skill ever derives a last-speaker /
unanswered-loop fact** (skill-audit gap list #1, #2, #4, #5) — so nothing
required, or even enabled, the report Gaetan expected.

## 2. Mechanism — what the run did vs what it should have done

Run: cron job `982c896036bc` "Close GCN-49 if Damien stays silent today",
session `cron_982c896036bc_20260813_235523`, model `gpt-5.6-sol` via
`openai-codex`, 4 tool turns (vm-run-trace §4).

**What it did**

1. **17:03:51 CEST** — posted a status bullet saying Damien "attend ton aide",
   stale by ~9–12s: Damien had already said "done" at 17:04:00 CEST(=15:04:00
   UTC)… actually 9s *after* the post; Gaetan closed the loop 24s after it
   (slack-evidence §a/§b; vm-run-trace §3). Tars had not read the Damien DM
   before posting — nothing tells it to (skill-audit §1, §4).
2. **17:09:39–17:13:10** — Gaetan intervened twice, live, to force the cron
   job's scope off "notre DM" onto `D08BJ8CQLP6` (slack-evidence §a thread
   replies 3–6; vm-run-trace §3). Tars corrected the prompt; the correction
   landed.
3. **21:55:41 UTC** — the run made exactly **one** Slack call:
   `mcp__slack__conversations_history {channel_id: D08BJ8CQLP6,
   include_activity_messages: false, limit: "1d"}` → 3,789 chars
   (vm-run-trace §4 tool table row 3, corroborated byte-for-byte from
   `mcp-stderr.log`). Correct channel, flat history, no thread-only read.
4. **21:55:49 UTC** — one `mcp__linear__get_issue` on GCN-49 (1,728 chars).
   **Zero writes** — the branch decided no `save_issue` was needed
   (vm-run-trace §4, §7).
5. **21:55:57 UTC** — delivered 294 chars to `D0BBYNM01BL`
   (`agent.log:1413`): "Damien a bien répondu après le message source …
   GCN-49 n'a donc pas été modifié … il était déjà en Done." Both claims are
   **correct** against Linear and Slack (slack-evidence §e.3).

**What it should have done** — with the *same single tool call and the same
data already in context*:

- Sort the returned DM history, identify the last message and its author →
  Gaetan, 15:04:15 UTC, "C'est complete de mon côté", unanswered.
- Identify the counterparty's last message → Damien, "done", 15:04:00 UTC.
- Classify the loop: **counterparty answered, Gaetan closed it himself, nothing
  pending on Damien.**
- Report that state next to the ticket fact: *"Damien a répondu « done » à
  17:04 ; tu as clos toi-même à 17:04:15 et il n'a rien envoyé depuis. GCN-49
  était déjà en Done depuis 17:10 (passé hors de ce job) — rien modifié."*

The delta is **zero extra tool calls**. The data was in the model's context at
21:55:41. This is a reporting/derivation gap, not a data-access gap.

## 3. Hypotheses, ranked against evidence

| # | Hypothesis | Verdict | Evidence |
|---|---|---|---|
| **1** | **(e) It had the data and drew the wrong (too-narrow) conclusion** | **PRIMARY — confirmed** | 3,789 chars of `D08BJ8CQLP6` returned at 21:55:41 (vm-run-trace §4/§7); independent re-read of the same channel shows the full 30-msg history is only ~2 days and the whole 2026-08-13 exchange fits (slack-evidence §b). The answer given is correct for the question asked; the question was the wrong one (vm-run-trace §4 note, §7 last bullet). |
| **2** | **(d) SOUL/skill guidance gap — nothing tells it to establish who replied last** | **PRIMARY (enabling cause) — confirmed** | skill-audit gap list #1/#2/#4/#5: no rule anywhere computes "did the counterparty reply to Gaetan"; `engagement-checker` computes only the mirror case, Gaetan's own open loops (skill-audit §3, gap #2). SOUL rule 10 (SOUL.md:253-265) fires only on *questions about the past* and fetches only *the channel it was pinged in* — structurally wrong-channel for a third-party stakeholder (skill-audit §4 bullet 1). `linear-ticketing` §12 (lines 216-224) covers Tars' own delegation provenance only (skill-audit §3). |
| **3** | **(b) Truncated window (limit/oldest)** | **NOT the cause here; real latent defect** | `limit: "1d"` is a **relative 1-day time window**, the tool's default, passed explicitly (vm-run-trace §4 row 3, §8 schema). At 21:55 UTC 13 Aug it covers back to ~21:55 UTC 12 Aug, which contains the entire 15:00–15:04 UTC exchange — nothing relevant was cut. **But**: had Gaetan's last message been >24h old (a real "gone silent" case — exactly the scenario the complaint describes), this default would have returned an empty/partial window and the run would have concluded from no data. Also the standing caps (~20 msgs SOUL.md:257-259; ~30 replies SOUL.md:19-21; "do not fetch unrelated channel history" engagement-checker:129) bias every such read shallow (skill-audit §2). |
| **4** | **(a) Credential structurally cannot read a Gaetan↔Damien 1:1 DM** | **DISPROVEN** | The Slack MCP is `ghcr.io/korotovsky/slack-mcp-server` on **xoxc/xoxd browser-session tokens** — a *user-session* credential with Gaetan's own account visibility, not a scoped `xoxb-` bot token (vm-run-trace §8, config.yaml:260-271). And the read empirically succeeded: 3,789 chars from `D08BJ8CQLP6` in 0.50s, no error (vm-run-trace §4). Nothing was papered over. |
| **5** | **(c) Never called a conversation-read; answered from the ticket alone** | **DISPROVEN** | One `conversations_history` on the correct channel, logged in two independent places (`agent.log:1405`, `mcp-stderr.log`) — vm-run-trace §4. |

Secondary contributing factor, worth logging separately: **the scope of the
check had to be corrected by a human, live.** Tars' first draft aimed the job
at "notre DM" (the Tars↔Gaetan channel) and only targeted `D08BJ8CQLP6` after
Gaetan asked twice (vm-run-trace §3 Finding; slack-evidence §a thread replies
3–6). The final run was correctly scoped *because a human caught the bug in
real time*, not because the model self-corrected — the same class of error as
the root cause (reasoning about the ticket's channel instead of the people's
channel).

Also unresolved but out of scope: GCN-49 flipped Todo→Done at 15:10:28 UTC,
~49s after Gaetan's "consider done", by a different session
(`20260813_140745_980067c0`, four `save_issue` calls in that window) — so the
23:55 job was retrospective confirmation of an already-closed ticket
(slack-evidence §d; vm-run-trace §9 bullet 3). Tars' report did say so.

## 4. Evidence sufficiency

Sufficient to pick. The one log-level gap — the raw Slack payload the run
received is not persisted on the VM, only the 3,789-char count (vm-run-trace
§9 bullet 1) — is closed adequately by the independent live re-read of the same
2-person DM (slack-evidence §b, vm-run-trace §6), and does not change the
ranking: whatever the exact payload, the model itself quoted Damien's "done" ts
correctly, proving the tail of the conversation was in its context.

One unverified item that does **not** decide the root cause but should be
checked before drafting the fix: `daily-work-brief/SKILL.md:18` loads a
`stakeholder-communication` skill that **does not exist anywhere in this
mirror** (skill-audit §4 bullet 3, gap #3). Whether it also does not exist on
the live VM is unverified — one command settles it:
`ssh gaetan@192.168.0.9 'ls ~/.hermes/skills/**/stakeholder-communication'`.
If it exists live, the fix belongs there and the mirror is out of sync (a SOUL
rule 2 violation); if it does not, `daily-work-brief` carries a dangling load
that silently no-ops on every daily brief.

## 5. Proposed fix (NOT applied)

Single file, single insertion. `SOUL.md` rule 10 is the right home: it is
always loaded, it already owns "read the record before you answer", and both
the cron path and the engagement path route through it. A `linear-ticketing`
edit alone would not catch the cron/engagement reports; a new skill would not
be loaded by the cron prompt at all.

**File**: `SOUL.md`, rule 10 (currently lines 253-265) — append one paragraph
at the end of the rule.

**Draft wording**:

> **Conversation state, not just events.** Whenever I report on a ticket, a
> task or a status that involves a named person other than Gaetan, I first read
> the Gaetan↔that-person DM or channel, and I state in the report **who sent
> the last message, when, and whether it is still unanswered** — before I state
> the ticket's state. I anchor that check on **Gaetan's last message**, never on
> the ticket's `Source:` ts (that one is the person's opening message and
> answering "did they post after it" is nearly always trivially yes). The
> ~20-message cap does not apply to this read: I pass an explicit `limit` wide
> enough to reach the last message of **each** party (start at `30d`, widen
> until both are found), because the whole point of the check is to catch a
> silence that may be days old — the tool's default `1d` window returns nothing
> in exactly the case that matters. If the person's channel is not reachable or
> I cannot establish who spoke last, I say so in the report instead of omitting
> it.

Two lines of collateral, both one-liners, both optional:

- `skills/orchestration/engagement-checker/SKILL.md:129` — the "Do not fetch
  unrelated channel history" ban needs a carve-out
  (`…except the DM with a person named in the item being reported`), otherwise
  the new SOUL paragraph and this line contradict each other.
- `skills/orchestration/daily-work-brief/SKILL.md:18` — resolve the dangling
  `stakeholder-communication` load: delete the line, or land the skill in the
  mirror (decide after the `ls` in §4).

Not proposed, deliberately: a new `stakeholder-communication` skill, a
last-speaker helper tool, or any change to the Slack MCP config. The fix is one
paragraph in the file that is already always in context; the run already made
the right tool call with the right data.

## 6. Confidence

**High** for the mechanism and the ranking: the tool trace is logged in two
independent places, the delivered text is byte-verified against the cron output
file and the Slack message, and the ground truth is confirmed by an independent
re-read. **High** for the root cause as stated (no rule derives conversation
state) — it is a grep-level negative over SOUL.md and every skill in the mirror.
The residual uncertainty is not about what happened but about what Gaetan wants
reported going forward: he described a "Damien went silent" case that did not
occur here, so the fix above is written to cover both the case he described and
the case that actually happened.
