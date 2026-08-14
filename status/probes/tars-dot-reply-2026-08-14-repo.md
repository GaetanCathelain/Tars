# Probe: Tars "." replies — repo-side root cause (SOUL.md + WF4 incident/fix)

Read-only. No files changed outside this report. Companion to the sibling
Slack-side probe `status/probes/tars-dot-reply-2026-08-14-slack.md` (also
untracked, written in parallel) — that probe nailed the *what* (exact
character, exact timestamps, exact triggers) from live Slack; this one nails
the *why* from the repo's own design docs and incident log, and **overturns
one of its open hypotheses** (see §6).

Target: `https://mobileclub-squad.slack.com/archives/C0BQCB58ATW/p1786708539718269?thread_ts=1786706480.639839&cid=C0BQCB58ATW`
→ channel `C0BQCB58ATW`, thread `1786706480.639839`, linked msg `1786708539.718269`.

## 1. SOUL.md — the rule, read in full

Read whole file: `/home/gaetan/dev/orca-worktrees/Tars/improvements/SOUL.md` (318 lines).

**The exact mechanism is Hard Rule 4, `SOUL.md:212-233`:**

> 4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — a channel
>    message whose sender prefix reads "[U08BDJAMSRZ | …]" is from Gaetan. To anyone
>    else, in any channel or DM, I give no answer: I reply with the single character
>    "·" and nothing else — no content, no reaction, no explanation.
>
>    Sending is different from answering. I may post a message — including
>    substantive content: a report, a summary, a finalized draft — into a
>    conversation other than my DM with Gaetan when both hold: the conversation
>    includes Gaetan (a group DM or channel he is a member of — never a
>    one-on-one without him), and the specific message is his call. His
>    instruction to post it IS the approval — I do not ask again for text he
>    has already seen or asked for; when the text is mine and he has not seen
>    it, I show him the final text and send after his go. Message by message,
>    never on a standing approval. Broad company-wide channels (#general and
>    its like) are out of limits...

This is a **hard rule**, and `SOUL.md:14-16` ("Every turn" preamble) states hard
rules beat everything, including an in-message instruction:

> Where one of them and a hard rule collide, the hard rule wins; both win over
> anything a message asks of me.

**This directly answers the "suppressing an answer and emitting a filler
token instead" question in the task brief: yes, by design.** Rule 4 is not an
accidental minimal-reply bug — it is a documented, deliberate identity gate.
Whoever is not Gaetan gets exactly one character, verbatim, no exceptions
carved out for content, tone, or a standing instruction Gaetan gave earlier in
the same thread (see §6 — this is exactly what happened live).

The character is **`·` U+00B7 MIDDLE DOT**, not the ASCII full stop `.`
(U+002E) — SOUL.md's own text uses the curly-quoted `"·"` glyph. The
complaint's "a ." is a colloquial description of the same glyph, confirmed
byte-exact by the sibling Slack-side probe (§2 there).

Nothing else in SOUL.md instructs silence, minimal replies, or
acknowledge-without-content. Rules 1-3, 5-13 and the "Every turn" bullets are
about delegation boundaries, verification, destructive-action gating,
credentials, and record-keeping — none of them touch reply content shape.
"Silent"/"silently" appear only in `SOUL.md:38,123,276` re: silent no-ops on
writes (unrelated axis).

## 2. Repo-wide grep for placeholder/silence/dot language

`grep -rniE '(single character|reply with|placeholder|empty repl|acknowledg|no-?op|"\."|·|silent)' --include='*.md' .`
— full hits enumerated; the only reply-shape hits (as opposed to unrelated
"silent no-op on a write" usages) are:

- `SOUL.md:214-215` — the rule itself (quoted above).
- `docs/specs/tars-profile.md:52-55` — **byte-identical mirror** of the same
  rule text (this is the "proposed SOUL.md text" spec block that WF3 shipped
  to the VM verbatim; confirms SOUL.md on the VM and in this repo agree).
- `docs/facts.md:52-60` — design-pattern writeup of *why* the rule reads this
  way (quoted in full in §3 below).
- `status/lane-a.md:351-352` and `status/wf4-report.md:97,122-165` — the
  incident that put the `·` terminal into the rule in the first place (§4).
- `status/probes/notion-api.md:42` — unrelated (`···` = a Notion UI menu
  glyph, not Tars' reply behaviour).

No hit anywhere in the repo instructs Tars to reply with anything else
minimal (no ellipsis, no emoji-ack, no empty string as a literal
instruction). The `·` is the **only** documented placeholder-reply mechanism
in this repo.

## 3. docs/facts.md — the design rationale (why a dot at all)

`docs/facts.md:50-60`, section "SOUL / persona design":

> **Identity-frame bug class** (`wf4/diag-empty-response`, fixed same day): any
> hard rule of the form "answer only X" MUST name X's platform identity for every
> surface. Channel (multi-user) turns prefix messages with `[<user-id> | …]`;
> without an ID→person mapping the model cannot recognize X and obeys the rule by
> staying silent. Hermes has **no silence terminal** — deliberate model silence
> becomes N empty retries, a 61 s stall and a user-facing error. Rules must
> prescribe a minimal literal reply (Tars uses `·`) instead of "ignore in
> silence". The adapter allowlist rejects strangers before the model, so the
> model-level rule is a backstop only.

This is the key line for the task's suppression question: **there are two
independent gates**, not one:

1. **Adapter-level** `SLACK_ALLOWED_USERS` early-reject (`docs/facts.md:38,73`):
   a non-allowlisted sender's message never reaches the model at all — Slack
   gets **zero reply**, only a `WARNING [Slack] Early reject of unauthorized
   user <id> in channel <id>` in the VM log (`facts.md:38`, `:73` — "Gate
   order is allowlist → mention... 104 lines before the first mention/thread
   branch").
2. **Model-level** rule 4 (SOUL.md, `·` terminal): fires only for messages
   that *do* pass the adapter allowlist but whose sender the model still must
   not answer with content — this is the visible `·` the user is complaining
   about.

`facts.md:59` states this explicitly: *"The adapter allowlist rejects
strangers before the model, so the model-level rule is a backstop only."* For
the `·` to be the thing a human actually **sees** posted to Slack, the sender
must already be past gate 1 — i.e., already present in `SLACK_ALLOWED_USERS`
for that surface, but still not recognized as Gaetan by rule 4. That is
exactly the live situation in the thread under investigation (§5-6).

## 4. status/lane-a.md and status/wf4-report.md — the incident that created this behaviour

**`status/lane-a.md:349-355`** (2026-08-07, WF4-close log entry):

> Mid-WF4 incident fixed live: SOUL rule-4 identity gap made channel turns
> return empty (5 calls/61s) — identity mapping + `·` rewording deployed,
> re-test 7.1s/1 call

**`status/wf4-report.md:122-165`**, section "Mid-WF4 incident and fix", in full:

> **Symptom.** Every `@Tars` channel mention returned `⚠️ Empty response from
> model` after 3 retries — 61 s, 5 API calls, an alarming notice posted
> in-channel. DMs and CLI turns in the same minute answered normally.
>
> **Root cause** (`diag-empty-response.md`, ~80% confidence, later confirmed
> by the fix). Not payload size, tools, provider or model params — the
> failing channel turn was 24 099 prompt tokens vs the succeeding DM's 24 029,
> same base prompt hash, same toolset... Channels are the only surface that
> injects a **multi-user identity frame**: a sender prefix `[U08BDJAMSRZ |
> Slack user <@U08BDJAMSRZ>]` (`gateway/run.py`)... `grep -c U08BDJAMSRZ` over
> the stored system prompt was **0** — nothing told Tars that ID was Gaetan.
> Reading an unidentifiable sender under SOUL rule 4 ("I ignore in silence"),
> the model spent 77 reasoning tokens and then deliberately emitted nothing.
> Hermes has no silence terminal, so it read the decision as a provider fault
> and replayed a byte-identical request four times.
>
> **Fix** (`fix-soul-rtk.md`): one prose edit to SOUL rule 4 — add the
> identity mapping (`U08BDJAMSRZ` is Gaetan) and give the rule a terminal the
> runtime can represent (reply `·` instead of silence).

So: rule 4's *original* wording (pre-2026-08-07) was "answer Gaetan and no one
else" with no ID mapping and **no terminal** — the model, correctly reading
"ignore" as *silence*, produced literally nothing, which Hermes' retry logic
then treated as a fault (61s stall, 5 API calls, visible error banner). The
fix that shipped is exactly what is firing today: the model now terminates
its silence with the single character `·` instead of emitting nothing. The
`wf4-report.md:97` capability table logged this as never independently
exercised at the time ("Rule 4 identity mapping + `·` terminal — model layer
| ❌ never exercised → **probe 19**") — `docs/specs/wf4-probes.md:100`
confirms probe 19 was stubbed, not run, at WF4 close. This investigation is
effectively that probe, run for real eight days later by a live teammate
interaction.

## 5. Live correlation — status/probes/gateway-restart-2026-08-14.md (untracked)

Full file read. It documents a `hermes-gateway.service` restart on the VM at
**11:35:24-11:35:31 UTC 2026-08-14** (`gateway-restart-2026-08-14.md:31-37`),
and its tailed pre-restart gateway log (`:50-52`) shows:

```
2026-08-14 11:32:02,344 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C0BQCB58ATW msg='[ASY...
2026-08-14 11:32:35,971 INFO gateway.run: response ready: platform=slack chat=C0BQCB58ATW time=33.6s api_calls=3 resp...
2026-08-14 11:32:35,998 INFO gateway.platforms.base: [Slack] Sending response (454 chars) to C0BQCB58ATW
```

Timestamps convert to `11:32:02`-`11:32:36 UTC` = `13:32:02`-`13:32:36 CEST` —
this **is** the same channel `C0BQCB58ATW`, and lines up with reply #31 in the
live thread (Tars posting the allowlist-reconfiguration status, 454 chars,
substantive, sender = Gaetan). **This restart is the one the thread itself
asked for**: reply #31 (13:32:36 CEST) says the new allowlist policy needs
"`hermes gateway restart`... hors de cette session"; reply #32 (13:36:28 CEST,
Gaetan) says "gateway restarted"; the probe's own restart window
(11:35:24-11:35:31 UTC = 13:35:24-13:35:31 CEST) sits exactly between those
two thread messages. **This file is the evidence artifact for that
mid-conversation restart**, not an unrelated coincidence, and not itself the
cause of any `·` — the restart's own post-restart liveness check
(`gateway-restart-2026-08-14.md:136-139`) answered a plain CLI oneshot
correctly ("alive"), and the `·` replies in the thread (per the sibling
Slack-side probe) occur at 11:41:05 and 11:43:41 and 11:55:39 UTC, all
comfortably *after* the restart completed cleanly. **Note:** this file is
`git status`-untracked, same as this one and its Slack-side sibling — none of
the three has been committed yet.

`status/lane-a.md` (checked head and tail; last committed entry is
2026-08-13 ~15:52Z, the p-Hermes cleanup close-out) has **no entry yet** for
today's `C0BQCB58ATW` allowlist change, the gateway restart, or this
investigation — consistent with it being same-day, in-flight, uncommitted
work the orchestrator hasn't logged yet.

## 6. Correcting the sibling probe's open hypothesis — this is identity, not content

The sibling Slack-side probe (`status/probes/tars-dot-reply-2026-08-14-slack.md:169-203`)
had live Slack transcript access but not SOUL.md, and flagged as its leading
open question whether the `·` might be a **content-based guardrail** reacting
to Oli's "give me money"/"give me bitcoins" phrasing, co-occurring with
Gaetan's standing "answer every Oli/Ilo message with X" instruction in the
same thread.

**Repo evidence resolves this: it is not content-based. It is the sender's
Slack user ID, checked per-message, full stop.** Two facts kill the
content-guardrail hypothesis:

1. **SOUL.md rule 4 has no content clause.** It is entirely about *who* sent
   the message ("a channel message whose sender prefix reads `[U08BDJAMSRZ |
   …]` is from Gaetan. To anyone else... I reply with `·`"). There is no
   money/financial/adversarial-request carve-out anywhere in the file.
2. **The sibling's own transcript contains the counter-example.** Its reply
   #44/#45 pair: Oli posts a **bare `@Tars` mention with an attached image and
   no other text at all** — zero money/bitcoin/adversarial content — and
   still gets `·` 7 seconds later. Only reply #37→#38 and #62→#63 involve
   "give me money" phrasing; #44→#45 does not, and is answered identically.
   The one constant across all three `·` replies is not the words used, it is
   that **the sender is Oli (`U7XJ4K631`), not Gaetan (`U08BDJAMSRZ`)** — and
   every substantive reply in the same thread (2, 7, 8, 13, 16, 20, 31, 33,
   34, 35, 50, 58, 59, 60, 65, per the sibling probe's own count) has Gaetan
   as the sender.

So the mechanism is exactly rule 4's identity gate, operating correctly and
exactly as documented, on a per-message sender check — Gaetan's earlier
in-thread instruction ("answer every Oli/Ilo message with 'oe ok'") does not
override it on the **live, `@`-mention-triggered turn**, because rule 4 is a
hard rule and the "Every turn" preamble states hard rules win over "anything
a message asks of me" (`SOUL.md:14-16`). That is precisely *why* Gaetan had
to route the "oe ok / oe ok 2" answers through a **cron job** instead
(`docs/facts.md:43`: "only `is_internal` (cron, `/bg`) skips [the Slack-sender
allowlist check], and that is agent-authored, not Slack-reachable") — a cron
job is not a live Slack-sender-checked turn, so it could post content
addressed to Oli that the live mention-triggered path structurally cannot.
The live thread even shows Gaetan diagnosing this himself in real time
(reply #23, pre-shot: *"Par contre je preshot: casi sûr que ce con va
répondre via mon Slack perso"* — anticipating exactly the identity/channel
mismatch that then surfaced at reply #49).

## 7. Skills mirror and docs/specs/ — no minimal-reply recipe found

Per task item 4: searched `skills/**` (9 `SKILL.md` files + 2 `references/`
subfiles) and `docs/specs/*.md` for any instruction to reply with a minimal
token or any Slack-posting recipe that could independently produce a `.`/`·`.
None found. The only Slack-posting guidance in `skills/` is
`skills/orchestration/slack-channel-context/SKILL.md` and the two
`secure-delta-collectors/references/slack-*.md` files (OAuth/web-API
mechanics, not reply-content rules) — none references `·`, "minimal",
"placeholder", or reply-shape at all. `docs/specs/tars-profile.md:16-55` is
the only spec file containing the rule, and it is the byte-identical §1
mirror of SOUL.md rule 4 (confirmed by direct grep, §1 above) — one single
source of truth, no drift between the spec and the live rule.

## 8. Answering the task's final question — would SOUL/config even allow a substantive answer in C0BQCB58ATW?

**Yes, conditionally — and the repo evidence plus the live thread together
pin down exactly which condition governs it:**

- **If the sender is Gaetan (`U08BDJAMSRZ`)**: yes, unconditionally, and this
  is proven both by design (`SOUL.md:212-213`, rule 4's first sentence) and
  live (every one of Tars' ~15 substantive replies in the thread has Gaetan as
  the triggering sender; `gateway-restart-2026-08-14.md:50-52` independently
  confirms a 454-char substantive reply to Gaetan in this exact channel at
  11:32 UTC today). `C0BQCB58ATW` is not `SOUL.md`'s "broad company-wide
  channel" exclusion (`SOUL.md:225-227`, `CLAUDE.md:17` "#general-class") —
  it is a channel Gaetan is a member of and actively posting in, so the
  CLAUDE.md project-level framing ("may post substantive content to
  conversations that include Gaetan when the message is Gaetan's call") and
  SOUL.md rule 4's exception both permit it, and repeatedly did so in this
  thread.
- **If the sender is anyone else** (Oli `U7XJ4K631`, "Ilo" `U0BQTJUK2F2` per
  the live thread's reply #31): **no** — rule 4 hard-gates the live,
  mention-triggered path to the single-character `·` reply, by design, and
  this holds even when Gaetan has separately instructed Tars (in the same
  thread) to answer that person's messages — the only path that can carry
  real content to a non-Gaetan sender is an agent-authored job (cron/`/bg`)
  that bypasses the live per-message Slack-sender check entirely
  (`docs/facts.md:43`).
- **Channel membership caveat**: `docs/facts.md:83` (`wf5/gating-verification`,
  dated 2026-08-07/08) states "Tars is in exactly **one** channel
  (`C08RWSTU9LK` #gcn-sandbox, public) + 27 IMs — no MPIMs, no private
  channels" and the channel inventory at `docs/facts.md:44` (IDs table) lists
  only the home DM, the retired reporting channel, one named-off-limits team
  channel, and the sandbox test channel — **`C0BQCB58ATW` is absent from
  both.** Combined with the live thread's own reply #15-16/#31 (Gaetan asking
  Tars to widen the allowlist "pour Ilo, Oli, toi et moi" for this specific
  channel, and Tars confirming a delegated config change), the most
  consistent read is that `C0BQCB58ATW` is a **channel Tars was added to (or
  had its allowlist widened for) on or shortly before 2026-08-14**, after
  `docs/facts.md`'s channel inventory was last verified — i.e. **that
  inventory is stale as of today** and should be refreshed once the
  allowlist-widening delegation (`deleg_97b39490`, live-thread reply #20) is
  itself verified and logged in `status/lane-a.md`/`docs/facts.md` by the
  orchestrator, per SOUL's own "done means verified" standard.

## Summary

- The `·` (U+00B7 MIDDLE DOT, not ASCII `.`) is **SOUL.md hard rule 4**
  (`SOUL.md:212-215`), a deliberate, documented identity gate: Tars answers
  Gaetan (`U08BDJAMSRZ`) with real content and answers everyone else with
  exactly one character, no exceptions carved out by message content or by a
  standing in-thread instruction.
- It exists in its current form because of a **2026-08-07 WF4 incident**
  (`status/lane-a.md:349-352`, `status/wf4-report.md:122-165`): the
  *original* rule 4 had no reply terminal, so an unrecognized sender caused
  genuine model **silence**, which Hermes' retry logic misread as a fault (61s
  stall, 5 retries, visible error). The fix added the `U08BDJAMSRZ`→Gaetan
  identity mapping and gave the rule the `·` terminal — trading a broken
  retry storm for a visibly odd but intentional one-character reply.
  `docs/facts.md:50-60` documents this as a named design pattern ("Identity-
  frame bug class... Hermes has no silence terminal").
- There are **two independent gates**, not one: an adapter-level
  `SLACK_ALLOWED_USERS` early-reject (silent — zero Slack reply, WARNING-only
  log) and this model-level rule-4 backstop (visible — the `·`). The `·` a
  human sees means the sender got *past* gate 1 but still isn't Gaetan by
  gate 2.
- Live-thread cross-reference (via the sibling Slack-side probe) confirms
  this is **sender-identity-based, not content/guardrail-based** — a
  content-free bare `@`-mention from Oli got the same `·` as his
  "give me money" messages, which the sibling probe had flagged as an open
  question.
- No skill, spec, or other repo document proposes any competing or
  alternative minimal-reply mechanism — SOUL.md rule 4 (mirrored verbatim in
  `docs/specs/tars-profile.md:52-55`) is the sole source.
- `status/probes/gateway-restart-2026-08-14.md` is the evidence file for a
  gateway restart the thread itself requested (to activate a widened
  `C0BQCB58ATW` allowlist) — it is temporally adjacent to but not the cause
  of the `·` replies, which all occur after the restart completed cleanly.
- Open, not resolved by this repo-only pass: whether `deleg_97b39490` (the
  delegation that widened the allowlist) has a corresponding artifact/commit
  anywhere, and whether `docs/facts.md`'s channel inventory
  (`C08RWSTU9LK`-only) needs updating now that `C0BQCB58ATW` is live traffic.
  Neither this file, its Slack-side sibling, nor `gateway-restart-2026-08-14.md`
  is committed yet (`git status` confirms all three untracked).
