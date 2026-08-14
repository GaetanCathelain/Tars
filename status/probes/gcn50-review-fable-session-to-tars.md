# GCN-50 adversarial review (Fable) — cruxes B, C, D: Session→Tars wire, rehydration, dropped options

**Date:** 2026-08-14 · **Reviewer:** adversarial Fable session (read-only; no Slack posts, no live sessions, no Tars turns fired)
**Inputs read in full:** the locked design (`scratchpad/gcn50-design-for-review.md`); committed probes
`gcn50-q10-kanban-smoke.md`, `gcn50-q12-headless-sendmessage.md`, `gcn50-q12-channels.md`,
`gcn50-q12-nonclaude-socket.md`, `gcn50-q14-w1-w2-comparison.md`, `gcn50-q15-lean-sendmessage.md`;
scratchpad artifacts `hermes-push-surfaces.md`, `cooper-messaging-facts.md`,
`cross-session-messaging-facts.md`, `wf5-specs-digest.md`, `q10-wake-verdict.md`; plus
`docs/facts.md` (thread/session-keying rows) and `status/probes/2026-08-14-damien-followup-rootcause.md`
(prior-art failure for crux C).

**One-line verdicts**

- **B — REVISE.** The `chat -Q -q | send -t` primitive beat kanban fair and square, but "every
  session→Tars event = a full Hermes turn" is the wrong default: adopt two-tier (raw threaded
  `hermes send` for status — already proven in the same probe — full turn only for
  question/escalation/completion), and stop calling the CLI turn "Tars": it is measurably not
  the DM agent (wrong voice, no history, different session store).
- **C — NOT MET as stated.** Rehydration was proven from a VM-local `/tmp` file named explicitly
  in the prompt; the design's LANE.md lives in a cooper worktree, the cross-host read was never
  probed, there is no thread_ts→lane index, no write-back enforcement, no atomicity/locking —
  and the repo has already shipped exactly this failure class (GCN-49 Damien follow-up).
- **D — SOUND, with two footnotes.** All four kills are evidence-backed. But the Orca mailbox's
  one good property (durable, acked delivery) is silently lost and must be re-covered in wire 3;
  channel MCP should be *parked* (documented external-push path, preview-gated today), not buried.

---

## 0. Claims-vs-evidence audit (design §"Four wires")

| Design claim | Evidence | Audit |
|---|---|---|
| Wire 2 "fires a REAL Hermes agent turn" | q14 §W2.6: `agent.turn_context` + 2 API calls + `Turn ended` | **SUPPORTED** |
| "Measured: 14.6s end-to-end" | q14 §W2.5: 14.556 s | **SUPPORTED but non-generalizable** — that was a trivial "compose one line from a local file" task, ~4 s of which was a wasted speculative `kanban_show` error. A real turn that must ssh to cooper, read LANE.md+SESSION.md, reason, possibly fire wire 3, will be materially slower. Do not quote 14.6 s as the design's latency. |
| "rehydration from file proven by nonce" | q14 §W2.7: `W2-NONCE-3ea8` only ever on disk | **OVERCLAIMED for the design.** Proven for `/tmp/probe-lane/LANE.md` **on the VM**, with the absolute path handed to the model in the prompt. The design's LANE.md is in a **cooper worktree**: the chat turn must ssh cooper→read a remote file mid-turn. Plausible (hermes-cli toolset has a terminal tool) but **never exercised**. This is the single largest measured-vs-asserted gap in the design. |
| "`-Q` is pipe-safe (answer stdout, session_id stderr)" | q14 §W2.2, `cat -A` | **SUPPORTED** |
| Composed message reached the right thread | q14 §W2.7: `slack_read_thread` on the anchor, one reply, nothing top-level | **SUPPORTED**, and it renders as **From: Tars (U0BBH85NAKH)** — a human sees "Tars said this". Caveats: (a) threaded `send --json` omits `"mirrored"` (top-level returns `mirrored:true`) — whether that means threaded posts are NOT mirrored into any Hermes session transcript is **unresolved and load-bearing** (see C-3); (b) the probe's anchor was minutes old — a lane thread lives for hours/days, and the *inbound* half (Gaetan replying inside a bot-anchored DM thread waking the right session) was never exercised in any GCN-50 probe. |
| Wire 3 lean `-p`: 3,870 tok / 4.7 s, delivery intact | q15 table, nonce-verified 8/8 | **SUPPORTED** |
| Socket escape hatch: any uid-1000 process injects, no token/ancestry check under `accept` | q12-nonclaude variations A/B/B2/C | **SUPPORTED** |
| Kanban = RAW-POST, wake gated on session_id only settable in-turn | q10 §6 (source + positive control), q14 §W1.2/W1.6 | **SUPPORTED** |

Design line "Tars reasons, rehydrating from LANE.md" is therefore one-third measured (a turn
ran), one-third measured-under-easier-conditions (local-file rehydration), one-third asserted
(that the turn is *Tars* in any sense Gaetan would recognize — see B-2).

---

## B — Is `hermes chat -Q -q | hermes send` the best Session→Tars transport?

### B-1. Every event does NOT need a full agent turn. Two-tier wins.

The design routes *all* session→Tars events through a full Hermes turn. Attack:

- **Cost/latency per event, measured:** ~15 s wall + 2 model API calls on the VM
  (`gpt-5.6-sol`) per event (q14 §W2.5–W2.6), plus a demonstrated wasted tool round-trip. Per
  lane, per milestone ("spawned", "25%", "tests green", "pushed"), across concurrent lanes,
  this is a standing tax with no consumer: routine status needs no decision.
- **What does the turn add for routine status? Nothing that survives the turn.** The composing
  session is *stateless and exits* (q14 comparison row "coupling"). It doesn't durably advance
  Tars' awareness — the next fresh Tars turn knows only what's in LANE.md/SESSION.md, which the
  Claude session can (and under the design's own state model, must) write anyway. The Hermes
  turn is a pure rephrasing hop for status lines — and a Claude Code session is at least as
  good a one-line-status author as a blind CLI chat turn reading the same file.
- **The rephrasing hop adds failure modes, removes none:** empty chat stdout posts an empty
  Slack message (q14 practical notes); the LLM can garble facts in transit; the speculative
  tool-call latency is nondeterministic.
- **The cheap alternative is already proven by the same probe.** Threaded raw
  `hermes send -t slack:D0BBYNM01BL:<thread_ts>` is exactly the second half of W2's pipe —
  exercised live, ~1 s, exact text control, no LLM (`hermes send` self-documents "no LLM, no
  agent loop", hermes-push-surfaces §5).

**Recommendation (B-1): two-tier.**
- *Routine status / milestones:* session composes the line itself → raw
  `hermes send -t slack:D:<thread_ts>` → **and appends the same line to SESSION.md** (durable
  record; see C-3 for why Slack alone is not a record Tars can read back).
- *Questions, escalations, completion:* full `chat -Q -q | send -t` turn (or better — B-2's
  variant), because there Tars genuinely has a decision to make (answer from Linear/LANE.md
  itself vs. surface to Gaetan).
- Note the probes never actually measured the decision case: q14's prompt was "compose a
  one-line status". "Tars decides whether to bother Gaetan" — including the awkward fact that
  the pipe **unconditionally posts stdout to the thread**, so a turn that decides *not* to
  escalate still ships something — is unprobed. Before locking wire 2 for escalations, run one
  probe of the question flow (turn reads LANE.md, decides, output routed conditionally —
  e.g. session inspects stdout and only pipes to `send` when the turn says ESCALATE, which the
  `tee` shape in q14 already supports).

### B-2. The voice/history problem is structural, not cosmetic — but it is containable.

Measured facts: the W2 turn answered in terse English while the same probe's gateway wake
(W1 reply 2) answered in French, "Tars' operator voice" (q14 §W1.7, comparison row
"persona"). The CLI chat session "is *not* the gateway DM session — it carries no conversation
history with Gaetan" (q14 practical notes, verbatim). Deeper, from `docs/facts.md`:

- Gateway sessions are keyed `agent:main:slack:<chat_type>:<team>:<channel>:<thread_ts>` —
  **the lane thread has its own shared gateway session**; CLI chat sessions
  (`20260814_…`) are a different species in the same `state.db`.
- So the design puts **two different "Tars minds" in one thread**: CLI-composed posts (no
  history, drifting voice) interleaved with genuine gateway thread-session replies whenever
  Gaetan writes. Persona continuity across a lane thread is not guaranteed by anything.

Is prompting for voice enough? It's untested (q14 only *suggests* "say so in the prompt"), and
even a perfect French one-liner doesn't give the CLI turn the thread's history — it can
contradict or repeat what "Tars" said two messages up, because it never saw those messages.

**Recommendations (B-2):**
1. Under two-tier this mostly dissolves: routine status lines are honest *lane telemetry* — give
   them a fixed prefix (e.g. `[lane:<name>]`) so they visibly aren't conversational Tars turns.
   Don't dress a history-less CLI turn up as the DM agent.
2. For the escalation turn, pin voice AND supplied history in the prompt (LANE.md must carry a
   "said so far in this thread" section — see C-4) — or, better:
3. **One un-run probe that could dissolve the problem entirely:** `hermes chat --resume
   <SESSION_ID>` exists (q14 §W2.1). Nobody has tried resuming **the thread's own gateway
   session key** from the CLI. If `--resume 'agent:main:slack:dm:…:<thread_ts>'` (or whatever id
   `state.db` shows for the lane thread) works, the escalation turn runs *in* the thread's
   session — history, persona, and transcript continuity all at once, and its output mirrors
   into the very session that will later process Gaetan's reply. Risks a two-writer race with a
   concurrent gateway turn, so probe it; but it is the only variant in which "the thing that
   talks to Gaetan" actually *is* Tars-the-DM-agent. Cheap to test, high payoff either way.

### B-3. Better surface entirely? No — the alternatives lose, for probe-backed reasons.

- **A2A inbound (loopback:9900):** the only surface where a turn runs *inside the live gateway
  process* on the real profile (wf5-specs-digest C1) — the strongest theoretical fix for B-2.
  But: zero traffic ever recorded, session/thread semantics unknown, no documented bridge from
  an A2A task to a Slack thread (its output returns to the RPC caller — you'd still need
  `hermes send`), Agent Card overstates capability (C2), and an A2A task almost certainly gets
  its own contextId-keyed session — i.e. the same fresh-session property as `chat`. It adds an
  unproven protocol to reach the same place. Skip; re-examine only if B-2's `--resume` probe
  fails AND persona continuity is declared blocking.
- **Webhook:** disabled, absent from config.yaml entirely (hermes-push-surfaces §2), and Slack
  thread targeting appears simply not implemented upstream (q10-wake-verdict row 7 — Telegram
  got `message_thread_id`, Slack visibly didn't). Enabling it = config edit + gateway restart +
  a new listener, to obtain less than what `send` does today. Correctly not chosen.
- **Linear comments Tars polls:** polling = cron, and cron delivery is channel-level with a
  60 s tick — the exact shape behind the already-root-caused wrong-anchor incident
  (q10-wake-verdict row 6; GCN-49). Not a wake path. (Linear comments as an *additional durable
  event log* alongside two-tier is cheap and fits "Linear = system of record" — optional, not a
  transport.)
- **SESSION.md + Tars notices next turn:** as the *only* mechanism it's polling with human-scale
  latency and no push at all; as the durable substrate under push it's already in the design.
  Correct as substrate, insufficient as wire.

### B-4. Thread-rooting re-check (asked explicitly)

Yes — q14 §W2.7 proves the composed message landed as a reply **inside** the anchor thread in
`D0BBYNM01BL`, authored by the Tars bot user, nothing at channel top level. Human-visible
attribution to Tars: yes. **Two rooting caveats stand:** (1) the missing `"mirrored"` key on
threaded sends leaves it unknown whether these posts enter any Hermes transcript (C-3 makes
this matter); (2) the anchor is bot-authored and the whole operator gate assumes Gaetan replies
*in-thread* in a DM — per facts.md, a top-level DM message gets a synthetic `thread_ts = ts`
and starts a **new session with no lane binding**, so a natural top-level "oui vas-y" lands in
the wrong session. The inbound half of wire 4 (Gaetan's threaded reply → correct thread session
→ Tars relays via wire 3) has **zero probe coverage**; the DoD round-trip is currently the
first-ever test of it. Acceptable only if everyone knows that's what the DoD is.

---

## C — Does stateless fresh-turn + LANE.md/SESSION.md rehydrate ALL context?

**No — and "ALL" is not just unproven, it is unenforceable as specified.** Holes, concrete:

1. **Cross-host rehydration is unproven** (see audit table). Every fresh Tars turn — gateway
   thread turn or wire-2 chat turn — runs on the VM; LANE.md/SESSION.md live in a cooper
   worktree. The one measured rehydration used a VM-local file with its path in the prompt.
   The design's actual read path (in-turn `ssh cooper cat <worktree>/LANE.md`) has never run.
2. **No thread→lane index.** LANE.md maps lane→thread_ts, but the wake direction is the
   reverse: a gateway turn triggered by Gaetan's in-thread answer knows `thread_ts` and must
   find the worktree. Nothing specified provides that lookup (grep every worktree's LANE.md
   over ssh? unstated). Same hole for multi-lane awareness: no lane registry exists, so a fresh
   top-level "status of everything?" turn has nothing to enumerate.
   *Fix:* store `thread_ts` + worktree path + peer-name on the Linear ticket (Tars has full
   native `mcp__linear__*` — design already names Linear the system of record); that one write
   kills holes 2 and the multi-lane gap simultaneously.
3. **The DM thread back-and-forth is NOT reliably re-readable — this is measured, not
   hypothetical.** From `docs/facts.md` (wf5/source-context, live-verified): a thread session
   cold-start fetches the real Slack thread **once**, capped at 31 messages; **steady state
   makes ZERO Slack calls** — context is Hermes' own transcript only. Wire-2/raw-send posts are
   bot-authored Slack messages that never traverse the gateway inbound path, and threaded sends
   show no `mirrored` flag. Consequence: once the lane thread's gateway session is warm, **every
   subsequent wire-2 status post is invisible to the Tars that answers Gaetan in that thread**,
   and threads longer than 31 messages are partially invisible even cold. So "is the thread
   re-read from Slack?" — only once, only 31 messages, and never again. LANE.md is the *only*
   continuity mechanism, which makes hole 4 fatal rather than theoretical.
4. **LANE.md currency is guaranteed by nobody.** The design says LANE.md holds "Gaetan's
   in-thread answers so far" — but no rule forces the gateway turn that *receives* an answer to
   write it back before ending, and nothing verifies it happened. The repo has already shipped
   this exact failure: GCN-49 (`2026-08-14-damien-followup-rootcause.md`) — a fresh turn
   reported an internal boolean and "never reported the state of the conversation at all",
   root-caused to the absence of a conversation-state rule. Stateless-by-design makes that
   class *structural* unless write-back is a hard contract.
   *Fix:* a SOUL-rule-2-style same-turn discipline: any Tars turn that says or hears anything
   decision-relevant in a lane thread MUST update LANE.md in that same turn, and wire-2
   escalation prompts MUST include "append your posted text to LANE.md §said-in-thread".
   Two-tier (B-1) already forces the session to log its own status lines into SESSION.md.
5. **Single-writer is asserted, not real.** "Tars-written" ≠ one process: a gateway thread
   turn, a cron turn, and a wire-2 chat turn can all be live simultaneously, each writing
   LANE.md over ssh. No lock, no atomic-write discipline is specified. A crash mid-write leaves
   a truncated file the next fresh turn treats as ground truth. The repo already owns the
   remedies (flock doctrine; the 2026-08-13 tmp+mv-drops-mode lesson): mandate
   `flock` + tmp+mv (+`chmod` if perms matter) for every LANE.md write, and make wire-2 turns
   **read-only** on LANE.md (they report; the coordinator-side turns decide/write).
6. **Staleness is indistinguishable from truth.** If the session dies, SESSION.md's "current
   step" reads as live forever. *Fix:* readers must treat SESSION.md as a claim: check mtime
   and the session's presence in ListAgents (wire-3's lean relay can carry that check) before
   reporting "in progress" to Gaetan.
7. **What is lost by construction:** prior tool results (logs read, diffs reviewed), in-flight
   nuance not distilled into the files, and the exact wording of past posts beyond the 31-message
   cold window. No file protocol recovers "ALL" context — it recovers *what was written*.

**Verdict C:** the honest claim is "rehydrates all **decision-relevant** context **provided**
write-back rules 2/4/5/6 are enforced". That is a weaker bar than Gaetan's stated one — say so
to him explicitly and get the restated bar accepted, rather than asserting Q11 is met. As
specified today, it is not.

---

## D — Were the dropped options killed for sound reasons?

- **Orca orchestration mailbox — sound kill,** thoroughly evidence-backed (wf5-specs-digest:
  `check` needs a live terminal handle even with `--run` — disproven-standalone live; `--ack`
  replay discipline; 300 s cap forcing the cron double-wake; the type-but-never-submitted spawn
  race that `workerState: ready` masks). **But name what dies with it:** the mailbox was the
  only mechanism with durable, acked, at-least-once delivery. Wire 3's SendMessage enqueues to
  a live peer; against a dead/renamed session the outcome is unprobed and the message is simply
  gone. *Re-cover it:* the lean relay must run ListAgents first (it already loads only
  SendMessage — add ListAgents to `--tools`), report the tool result verbatim, and Tars must
  treat non-delivery as lane-dead → respawn-from-LANE.md. Durability then lives in LANE.md,
  where the design wants it anyway.
- **Kanban — sound and DEAD, do not resurrect.** This is the best-run kill in the file set:
  q10 proved RAW-POST + the `session_id` wake gate from source *and* a positive control; the
  q10-wake-verdict judge said "keep kanban but run the smoke test"; the smoke test (q14 W1) ran
  and showed W1 *requires* a `hermes chat` turn anyway, posts twice, carries no prompt surface,
  and the stamped session_id is bookkeeping the notifier never reads (W1.6). The steelman
  ("kanban is the durable store that knows the thread") was already refuted upstream: Slack
  itself is the thread-keyed store (q10-wake-verdict, verdict section). Nothing to return.
- **Channel MCP — sound today; park, don't bury.** Killed as "a component wire 3 makes
  unnecessary" — true, and the probe adds harder reasons: dead under `-p` 3/3 with a silent
  drop, research-preview, `--channels` allowlisted, dev flag prints a warning banner, and the
  ssh-writable inbox dir is a prompt-injection surface the probe itself flags. **Steelman:** it
  is the *documented* mechanism for exactly this job ("push external events into a session"),
  costs a 55-line stdlib server, no per-message process spawn, and queues cleanly into busy
  sessions. At q15's measured 3.9k tokens/4.7 s per wire-3 spawn the spawn-elimination argument
  is worth little today — but if Tars→session traffic ever becomes high-frequency, or channels
  GA with `-p` support, this is the first thing to re-evaluate. Record it as parked-with-trigger,
  not dominated-forever.
- **TUI injection last-resort — correct.** It is keystroke injection with an incident-logged
  failure mode (typed-but-unsubmitted); break-glass status is exactly right.

**Should any return? No.** The two real gaps D leaves (durable delivery, persona continuity)
are better closed inside the chosen wires (ListAgents check; `--resume` probe / two-tier) than
by resurrecting a dropped component.

---

## Consolidated recommended changes (ranked)

1. **Two-tier the session→Tars wire** (B-1): raw threaded `hermes send` + SESSION.md append for
   status; full turn only for question/escalation/completion. Removes most of the cost, the
   empty-post failure mode, and most of the voice problem at once.
2. **Make LANE.md write-back a hard same-turn contract + atomic** (C-4/C-5): flock + tmp+mv;
   gateway turn records Gaetan's answer before relaying; wire-2 turns read-only. This is the
   difference between "stateless coordinator" and "GCN-49 at scale".
3. **Put the thread_ts→worktree→peer-name mapping on the Linear ticket** (C-2): kills the
   reverse-lookup hole and the multi-lane blindness with a tool Tars already has natively.
4. **Run two cheap probes before DoD:** (a) wire-2 turn rehydrating LANE.md *over ssh from
   cooper* (the actually-designed path — currently unmeasured); (b) `hermes chat --resume
   <thread-session-id>` — if it attaches to the lane thread's gateway session, the voice/history
   problem largely disappears for escalations.
5. **Wire 3: add ListAgents to the lean relay's `--tools` and treat SendMessage failure as
   lane-dead** (D): restores the delivery guarantee the mailbox kill silently dropped.
6. **Re-state the Q11 bar honestly to Gaetan** (C verdict): "all decision-relevant context,
   under enforced write-back" — not "all context". Get the restated bar accepted.
7. Stop quoting 14.6 s as the wire-2 latency; it is the floor for a trivial local-file compose
   including a wasted tool call, not the design's number.
