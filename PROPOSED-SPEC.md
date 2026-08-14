# PROPOSED SPEC — Tars V2 orchestration (Tars ↔ Orca sessions)

**Status:** PROPOSED — draft, 2026-08-14. Not built as automation; the round-trip
MECHANISM is now exercised live end-to-end (three E2E tests, §E2E validation).
Gaetan's call: "we'll get back to it later." This file captures the design so it
survives context loss; it is not an execution order.
**Ticket:** GCN-50 — *Engineer Tars ↔ Orca orchestration*.
**Supersedes for planning purposes:** `scratchpad/gcn50-design-final.md` (the
locked design this spec is the readable form of). All evidence lives under
`status/probes/gcn50-*.md`.

## What this is / how it was derived

Part A (Tars ↔ session communication) is a build-ready wire design; Part B (Tars
as an autonomous master coordinator) is a roadmap skeleton. The design was
reached by: a long grill of the delegation problem → three rounds of live probes
on cooper and the Tars VM (q10/q12/q14/q15 transport probes, then P1/P2/P3
rehydration probes) → three adversarial Fable reviews (Tars→session transport;
session→Tars + rehydration + dropped options; a rehydration/persona closing pass)
plus a `/megaultracode-orca` fleet-compatibility review. Every wire below is
backed by a committed probe file; every claim is tagged with what actually proved
it.

### Legend

- **[VERIFIED]** — measured live by a cited probe, with command output on disk.
  The pointer and the load-bearing number are given inline.
- **[INFERRED]** — one link in the chain is probe-adjacent but not directly
  exercised (a single sample, or a mechanism read from source but not run).
- **[PROPOSED]** — design not yet built or not yet exercised. This includes every
  discipline (a rule no probe can prove) and the whole inbound round-trip.

If a line has no tag it is descriptive glue, not a claim.

---

## E2E validation (live)

Since this spec was written the full loop was exercised live three ways — one per
Wire-3 transport — plus a teardown pass. All three closed the round-trip; each
self-cleaned or was cleaned (`gcn50-e2e-cleanup.md`).

| Test | Wire-3 exercised | Verdict | Key number | Evidence |
|---|---|---|---|---|
| 1 — SendMessage round-trip | lean `claude -p` SendMessage | **E2E-ROUNDTRIP-PROVEN** | 6.3k tok / 6.8 s send, 16.5 s receipt | `gcn50-e2e-roundtrip.md` |
| 2 — MCP Channel | file-drop into a watched inbox → channel push | **CHANNELS-E2E-PROVEN** | 0-token send, 8.18 s round-trip (French) | `gcn50-e2e-channels.md` |
| 3 — Go non-Claude peer | one resident Go process holds both wires, no spawn | **GO-BIDIRECTIONAL-PEER-PROVEN** | ~0-token socket write; same-uid/same-box only | `gcn50-e2e-go-bridge.md` |

**What this changes.** The Tars→session→operator→session round-trip MECHANISM is
now VERIFIED end-to-end (test-1), not just component-wise. What stays unbuilt is
the AUTOMATION — Wire-1 spawn hardening and the Wire-4 inbound relay (both hand-run
in the tests). Details fold into the wires below.

---

## 1. Problem — why the current delegate-to-cooper mechanism is replaced

Today Tars delegates code work by writing a prose brief and driving a coding
agent on cooper through the **Orca orchestration mailbox** (the `delegate-orca` /
`orc-tab` path, `check`/`--ack`/`--run`). That mechanism is being retired for
Tars ↔ session coordination because its failure modes are structural, not
incidental (all evidence-backed in the review set):

- **Stale handles.** `check` needs a live terminal handle even with `--run`; a
  session that has scrolled or been renamed leaves Tars addressing nothing.
- **Ack replay.** The `--ack` discipline is manual and replayable — an
  already-consumed message can be re-read as new.
- **Spawn race.** `workerState: ready` masks a type-but-never-submitted race at
  spawn (incident-logged in this repo).
- **300 s cap.** The mailbox forces a cron double-wake to survive its own
  timeout, exactly the wrong-anchor cron shape behind GCN-49.

The one property worth keeping from the mailbox — **durable, acked,
at-least-once delivery** — is re-covered inside Wire 3 (ListAgents liveness +
application-level ack), not by keeping the mailbox. See §Dropped as dominated.

---

## 2. Spine

Tars = Gaetan's personal Hermes agent (v0.20.0) on the VM `192.168.0.9`, live in
Slack DM `D0BBYNM01BL`, answers only Gaetan. It delegates coding work to Claude
Code sessions in **Orca worktrees on cooper** (`192.168.0.4`, uid 1000). Tars
reaches cooper via `ssh cooper`; sessions reach the VM via `ssh
gaetan@192.168.0.9` + `~/.local/bin/hermes`.

**Tars is a stateless coordinator.** One delegation = one Orca worktree = one
lane = one Slack DM thread. A fresh Tars turn rehydrates from durable state
(LANE.md / SESSION.md / Linear / the VM lane-index file) AND can *resume the
actual gateway thread session* so it wakes **as Tars** — French DM voice, thread
history, durable transcript — instead of as a personaless CLI turn.

The whole design is four wires + a state contract.

---

# PART A — Tars ↔ session communication (BUILD-READY)

## Principles

1. **Deterministic addressing over discovery.** Tars chooses the peer name at
   spawn (`-n tars-<ticket>`); it never guesses by cwd. Four live sessions were
   measured sharing this one worktree's cwd, so cwd-matching is ambiguous *today*
   `[VERIFIED — gcn50-review-fable-tars-to-session.md §1, live read of
   ~/.claude/sessions/]`.
2. **Inline, never reference, across the cooper↔VM boundary.** A cross-host path
   is dead on arrival: a VM turn given a cooper path burns ~154.7k tokens over 5
   API calls to fail; the same content inlined succeeds in 1 call / ~31.4k tokens
   `[VERIFIED — gcn50-rehydrate-p1-inline.md]`.
3. **Right-size the transport per event.** Routine status is a raw ~1 s post with
   no LLM; only a question/escalation/completion pays for an agent turn
   `[VERIFIED — gcn50-q14-w1-w2-comparison.md]`.
4. **Observable failure beats raw speed.** Every Tars→session send is checked
   (ListAgents + parsed JSON + application-level ack), because the fast paths
   (raw socket) return zero bytes and no ack.
5. **Files are the record; Slack is a display surface.** A warm gateway thread
   re-reads nothing from Slack, and cold hydrate caps at 31 messages
   `[VERIFIED — docs/facts.md thread rows; gcn50-rehydrate-p2-threadindex.md]`.

## Wire 1 — Spawn

**Mechanism.** Tars → `ssh cooper` → orc-tab-style launcher creates the worktree
+ a Tars-aware Claude Code session, named explicitly **`-n tars-<ticket>[-rN]`**
(deterministic addressing; the `-r` suffix bumps on relaunch). One post-spawn
registry check: assert exactly one live registration with `name == tars-<ticket>`
**AND** `cwd == <lane worktree>`, to catch the silent collision-variant rename
and the crashed-before-registering case. The Orca orchestration mailbox is not
used.

**Commands.**
- Spawn: the orc-* launcher pair (`orca worktree create …` → `orc-tab start … orc-opus …`) with `-n tars-<ticket>` passed through. `-n, --name` is a documented `claude` flag `[VERIFIED — claude --help, cited gcn50-review-fable-tars-to-session.md §4]`.
- Registry check (one `python3`/`jq` line over ssh): read `~/.claude/sessions/*.json`, match name+cwd.

**Hardening deltas from review.**
- Trade Orca's work-based tab auto-rename for a fixed peer name — for Tars lanes,
  deterministic addressing beats tab aesthetics `[review §4]`.
- Post-spawn verify is **once per spawn**, not per message.
- Tars records the *intended* name in LANE.md + the VM lane-index + the Linear
  ticket **before first use** — no self-registration race, no discovery step.

**Evidence.** Headless enumeration and addressing proven: a headless `claude -p`
called ListAgents and saw the live 12-session roster including the intended target
`[VERIFIED — gcn50-q12-headless-sendmessage.md §3]`. Registry hygiene after a
*crash* (not clean exit) is unproven — the socket dir accumulates stale entries
(59 sockets for 16 pids observed) `[INFERRED — review §1]`.

**Startup-gate hardening (E2E finding).** `--dangerously-skip-permissions` does NOT
auto-answer Claude's two startup gates — folder-trust and "Allow external CLAUDE.md
imports?"; worse, the brief-seed is dropped while the dialogs are up and the
type-but-not-submit race applies to the re-submit `[VERIFIED — gcn50-e2e-roundtrip.md
wire-1 finding]`. **Mitigation, validated:** spawn the lane worktree OUTSIDE
`$HOME`'s CLAUDE.md inheritance (e.g. a scratch repo under `/tmp`) → the
external-import gate never fires, only folder-trust remains, cleared once at spawn
`[VERIFIED — gcn50-e2e-channels.md: gate never fired]`. The launcher must
confirm-submit the brief after clearing gates. **Orca hygiene:** this orca build has
no `repo rm` verb, so each scratch-repo registration dangles after teardown — Tars
spawning one repo per lane accumulates dead registrations; reuse a small pool of
lane repos `[VERIFIED — gcn50-e2e-cleanup.md]`.

## Wire 2 — Session → Tars (three rungs)

The reviews killed "every event = a full Hermes turn": a full turn is ~15 s + 2
API calls and adds a rephrasing hop with no consumer for routine status
`[VERIFIED — gcn50-q14-w1-w2-comparison.md §W2.5]`. Three rungs:

**Rung 1 — ROUTINE STATUS / milestones.** Session composes the line itself → raw
threaded `hermes send -t slack:D0BBYNM01BL:<thread_ts>`, prefixed `[lane:<name>]`
(honest telemetry, not a fake Tars turn), ~1 s, no LLM, exact text. Same ssh trip
appends the line to SESSION.md and refreshes the VM lane-index file.
- **Evidence.** `hermes send` self-documents "no LLM, no agent loop"; the threaded
  post landed in the anchor thread, authored by the Tars bot user
  `[VERIFIED — gcn50-q14 §W2.7, gcn50-q10 §4]`.

**Rung 2 — QUESTION / ESCALATION / COMPLETION.** A resumed gateway-session turn:

```
sid=$(ssh vm 'python3 …state.db?mode=ro… "SELECT id FROM sessions
      WHERE chat_id=? AND <thread_ts col>=? AND ended_at IS NULL … LIMIT 1"')
out=$(ssh vm '~/.local/bin/hermes chat --resume '"$sid"' --no-restore-cwd -Q -q "$(cat)"' < prompt.txt)
```

`prompt.txt` = the event + **LANE.md contents inlined** (never a path) + a routing
instruction. The session routes stdout: `ESCALATE:` prefix → second ssh, `hermes
send -t slack:D:<thread_ts>`; `RESOLVED:` prefix → answer stays in the session's
hands, no Slack post (fixes the "pipe unconditionally posts" defect; the `tee`
shape in q14 already supports it). `--no-restore-cwd` because a gateway session's
recorded cwd is meaningless to a headless relay. The **ssh-stdin `… -q "$(cat)"' <
prompt.txt`** transport is mandatory — it dodges multi-KB quote-escaping; do not
let an implementer collapse it to an argv-embedded string.
- **Why resume, not a bare turn.** The resumed turn *is* Tars: it re-loads the
  thread's own transcript and answers in the French operator register even against
  an English query, headless over plain ssh, 1 API call ~8.5 s, needing only a
  `state.db` row (no TTY, no live process) `[VERIFIED — gcn50-rehydrate-p3-resume.md]`.
  A bare `hermes chat -q` turn answers in the wrong voice with no thread history
  `[VERIFIED — gcn50-q14 §W1.7/comparison]`.
- **Inlining is load-bearing.** A buried mid-file decision, a mid-file nonce, and
  a third-section blocker were all synthesized into one correct line from the
  inlined text, `tool_turns=0` `[VERIFIED — gcn50-rehydrate-p1-inline.md]`.

**Rung 3 — COLD FALLBACK.** Plain `hermes chat -Q -q "$(cat)"` (P1 inline shape,
no resume) **only** when the `state.db` lookup returns no row (a lane thread no
gateway turn has ever run in — early, when the anchor was a raw `send` and Gaetan
hasn't replied). Voice pinned in the prompt; the degraded mode P1 measured.

**Hardening deltas from review.**
- Two-tier, not one-tier: the full turn is reserved for a real decision.
- Persona: status lines are labeled telemetry and don't pretend to be Tars;
  escalation/completion turns are Tars by construction (resumed session).
- **[INFERRED]** the persona-carryover effect is a single P3 sample; the mechanism
  (thread register beats language-mirroring) is labeled inferred by the probe.
- **[INFERRED]** resume mutates the live session row (`ended_at=NULL` + appended
  turns); whether the *next gateway wake* on that thread sees the CLI-appended
  turns is proven only in the gateway→CLI direction — see §Build-time probes.
- **[PROPOSED]** ESCALATE/RESOLVED prefix-following is prompt discipline, n=0.

**Never quote 14.6 s (or 8.5 s) as the design latency** — both are trivial-compose
floors, not the cost of a real decision turn that must ssh, read, and reason.

## Wire 3 — Tars → session

**Mechanism.** A lean, throwaway headless `claude -p` that calls the SendMessage
peer tool, spawned per message:

```
claude -p '<prompt instructing SendMessage>' \
  --output-format json --dangerously-skip-permissions \
  --model haiku \
  --strict-mcp-config --mcp-config <empty {"mcpServers":{}}> \
  --system-prompt "<1-line relay; TARGET NAME PINNED here, not in the user prompt>" \
  --setting-sources "" \
  --tools "SendMessage,ListAgents"
```

run from a cwd with no CLAUDE.md.
- **Evidence.** Config G measured **3,870 tokens / 4.73 s**, 8/8 nonce-verified
  deliveries, vs the 220,054-token / 17.39 s full-orchestrator baseline (−98.2%
  tokens, −72.8 % wall) `[VERIFIED — gcn50-q15-lean-sendmessage.md]`. Headless
  `-p` can call ListAgents + SendMessage with no permission hold, `success:true` +
  `msg_id` `[VERIFIED — gcn50-q12-headless-sendmessage.md]`.

**Hardening deltas from review** (this is where the mailbox's lost "durable
delivery" property is re-covered):
- **ListAgents liveness first** (added to `--tools`); a failed send == lane-dead.
- **Per-message nonce + application-level ACK** over Wire 2. SendMessage
  `success:true` is *enqueue at a live peer*, not receipt; identical repeats are
  silently dropped as loop-protection, so a blind retry vanishes
  `[VERIFIED — the drop is documented; nonce closes it, gcn50-review §1 finding 5]`.
  The ack = the session's next Wire-2 message quoting the nonce; Tars's cron
  surface owns the timeout on silence.
- Target name pinned in `--system-prompt`; Gaetan's free-text answer passed as an
  **opaque fenced block** the relay forwards verbatim (relay-injection containment,
  since the relay can message any peer on cooper) `[review §5]`.
- Standing line "do not reply to this message; report via your lane wire."
- Parse `--output-format json` (`is_error`/`subtype` + tool result): a
  non-`success:true` or missing msg_id is a FAILED send, surfaced to the lane
  thread, never retried blind.
- **Depends on the receiver having `crossSessionInbound: accept`** — all orc-*
  launchers set it `[VERIFIED — gcn50-q12-nonclaude-socket.md]`.

**Escape hatch (documented, not default).** A raw Unix-socket write to the
session inbox: any uid-1000 process (an `ssh cooper` command qualifies) injects
into an `accept` session with **no token and no ancestry check** — near-instant,
zero tokens, but fire-and-forget (0 bytes back, no ack)
`[VERIFIED — gcn50-q12-nonclaude-socket.md, variations A/B/B2 deliver, C held]`.
Hand-rolled against an undocumented 2.1.232-shaped wire frame; use only when the
lean relay errors twice, and resolve the socket path from the same verified
registration as Wire 1's check.

**Fallback ladder (pinned):** relay errors/times-out twice → raw-socket write →
socket connect fails → lane declared dead → Tars reports lane down + offers
respawn. TUI injection stays off the ladder (manual break-glass only).

**[PROPOSED] failure semantics against a dead / renamed / busy / held receiver are
untested** — every such state is reachable in production and none was probed as a
SendMessage target. Step-3 JSON parsing + the ladder are the design's answer; they
have not been exercised. A version tripwire (re-run q15 config G after every
`claude` upgrade) guards the lean flag stack, which is newer surface.

**Transport options — three E2E-proven; the choice is OPEN (Gaetan undecided).**

| Transport | E2E | Cost / latency | Constraints | Version-risk |
|---|---|---|---|---|
| **lean `claude -p` SendMessage** (default) | test-1 | 6.3k tok / 6.8 s per msg | headless-capable; official tool | low (official surface) |
| **MCP Channel** (file-inbox) | test-2 | 0-token send, ~8 s round-trip | interactive-only receiver + persistent per-session MCP subprocess + dev-flag opt-in; `inbox/` is a raw prompt-injection surface | medium (research-preview) |
| **resident Go / non-Claude socket peer** | test-3 | no per-msg spawn; socket write | **same-uid / same-box only** (0700 dirs) → needs a cooper-resident daemon; requires `accept` | **high** (undocumented reverse-engineered contract — `peerProtocol:1`, `sha256(socketpath)` key, `/proc` procStart; silent break on upgrade) |

**Leaning, NOT a decision (the transport is an OPEN question — see §Open questions).**
Tars is on the VM, sessions on cooper, so any same-box-only transport (the Channel
MCP subprocess, the Go peer) needs a component co-located on cooper; SendMessage runs
as `ssh cooper 'claude -p …'` landing as uid 1000 with no resident piece, works
headless, and rides the official cross-session protocol — which is why it is the
recommended *starting* point, not a committed choice. The Go peer is the
*productized* form of the raw-socket escape hatch and the natural upgrade if
per-message spawn cost ever hurts — at the price of owning an undocumented protocol
(re-run the nonce self-test on every `claude` upgrade, fail loud on schema drift).
Channels win on send-cost but pay with an interactive-only receiver and an injection
surface. All three are E2E-proven
`[VERIFIED — gcn50-e2e-roundtrip.md / -channels.md / -go-bridge.md]`. **The transport
is swappable behind the `lean-send` helper** — so the pick can be made at build time
and even changed later without touching the other wires. Gaetan has not committed to
one yet.

## Wire 4 — Operator gate + `thread_ts → lane` index

**Mechanism.** Gaetan replies in-thread → the gateway wakes with only thread
context → it must resolve `thread_ts → lane`. There is **no built-in lookup**:
no CLI verb, no session-store field, no wake-payload field maps a thread to a lane
`[VERIFIED — gcn50-rehydrate-p2-threadindex.md, source-verified over build_session_key + every relevant --help]`.
A woken in-thread turn without an index answers blind (this is the GCN-49 class).

The index, ladder-minimal:
- **Primary (the wake path): one VM-local file per lane, `~/.hermes/lanes/<thread_ts>.md`.**
  Content is *inlined substance* (ticket id, worktree path for humans/relays, the
  **peer name `tars-<ticket>[-rN]`**, decision log, said-in-thread section, current
  status) — a VM-side mirror of LANE.md, same convention the repo already runs for
  `SOUL.md` / `skills/`. Never a bare cooper path (unreadable from the VM, per
  Principle 2).
- **Writers.** Tars creates it at spawn (it holds the thread_ts and just posted the
  anchor). Thereafter the lane's cooper session refreshes it over ssh in the same
  trip as every Wire-2 send — single writer per lane file after spawn. `flock` +
  tmp+mv + `chmod 600` per repo hard rules (tmp+mv drops mode; the 2026-08-13
  world-readable-config incident is the precedent). The one two-writer touchpoint
  is the gateway turn that receives Gaetan's answer appending to said-in-thread
  same-turn; flock covers it.
- **Standing SOUL rule:** any in-thread wake checks `~/.hermes/lanes/<thread_ts>.md`
  before answering; absence ⇒ say "no lane bound to this thread", never guess from
  prose.
- **Secondary (durable + enumeration): the `(thread_ts, worktree, peer-name)` triple
  on the Linear ticket** at spawn — one native `mcp__linear__*` write. Covers
  multi-lane "status of everything?" turns and survives VM loss. Not the wake path.
- **No new thread_ts→session_id machinery needed:** `state.db`'s `sessions` table
  already carries chat_id + thread_ts (P3 queried it directly) — that is the Wire-2
  rung-2 resume lookup.

**Evidence.** Cold hydrate fetches the real Slack thread once (≤31 msgs), warm
threads make zero Slack reads `[VERIFIED — gcn50-rehydrate-p2, docs/facts.md]`. The
`kanban_notify_subs.thread_id` column is a real reverse index but covers only
kanban-bound lanes and fires only on a kanban wake — correctly ruled out (kanban is
dead per §Dropped) `[VERIFIED — gcn50-rehydrate-p2 §1]`.

**[VERIFIED-demonstrated — the gap is real] the inbound composition is unbuilt.**
Test-1 showed it live: Gaetan's in-thread reply woke the gateway, which answered him
conversationally ("English it is.") and did **not** relay inward to the session,
because the lane-index lookup + Wire-3 relay logic is not built
`[VERIFIED — gcn50-e2e-roundtrip.md wire-4 finding]`. Every *component* (wake,
resume, lean relay) is individually proven; only their composition on gateway-wake
is missing — **this is the top build item.** The DoD round-trip is its first
end-to-end test. A top-level DM (not in-thread) gets a synthetic `thread_ts = ts`
and starts a new, lane-unbound session — a known trap; tell Gaetan once that answers
go in-thread.

## State & rehydration contract

- **Linear = ticket system of record** (Tars has full native `mcp__linear__*`).
  The ticket carries the `(thread_ts, worktree, peer-name)` triple.
- **LANE.md** — Tars-authored brief/decisions/answers. **SINGLE-WRITER (Tars).**
  Wire-2 turns are **read-only** on it. **[PROPOSED] enforced same-turn write-back:**
  any turn that says or hears anything decision-relevant in a lane thread MUST
  update LANE.md (+ the VM lane mirror) in that same turn. This is the one thing no
  probe can prove — it is a discipline — and its absence is exactly the GCN-49
  wrong-conversation-state replay (`2026-08-14-damien-followup-rootcause.md`). It
  must live at SOUL-rule level, and the cron nudge (Wire-3 timeout guard) must
  resolve its anchor through the lane index, never compose thread-blind.
- **SESSION.md** — session-authored status/blockers/question+why. Readers treat it
  as a claim: check mtime + ListAgents presence before reporting "in progress".
- **Atomicity:** `flock` + tmp+mv (+ `chmod 600`) on every LANE.md and lane-mirror
  write.
- **Committed to git:** deliverables + probe/evidence only. State files (LANE.md,
  SESSION.md, VM mirror) are **not** committed.
- **What rehydrates is what was written.** Prior tool results, un-distilled in-flight
  nuance, and thread text beyond the 31-message cold window are lost by construction.
  The honest bar is **"all decision-relevant context, provided the write-back
  contract is enforced"** — Gaetan must accept this restated bar explicitly; the
  literal "ALL context" bar no design of this shape can meet
  `[verdict: MET-WITH-CONDITIONS — gcn50-rehydrate-review.md §1]`.

**[PROPOSED] accepted risks (state once):** concurrent `--resume` vs a live gateway
turn on one `state.db` row — sqlite serializes writes, so the credible failure is
interleaved transcript, not corruption, and escalations fire when the session is
blocked so overlap is rare. LANE.md size ceiling — P1's marginal-cost curve is
measured only at 2.3 KB; linearity to ~23 KB is arithmetic; keep LANE.md a
distillate, not a log. Do not `--resume` into pre-2026-08-13 threads (session-key
shape drift).

## Dropped as dominated

| Option | Killed because | Killing evidence |
|---|---|---|
| **Orca orchestration mailbox** | brittle: stale handles, ack replay, spawn race, 300 s cap. Its one good property (durable acked delivery) is re-covered by Wire 3's ListAgents + app-level ack. | `gcn50-review-fable-session-to-tars.md §D`; wf5-specs-digest |
| **Kanban as wire** | RAW-POST with no lane context; the agent-wake is gated on `task.session_id`, settable only from *inside* a chat turn → requires Wire 2 as its own precondition, then adds a card + subscription + notifier + a duplicate template post to deliver **less**. Best-run kill in the set. | `gcn50-q10-kanban-smoke.md` (source `kanban_watchers.py` L637-646/L729-772 + positive control); `gcn50-q14 §W1` (W1 requires W2; posts twice; ~37 s incl. precondition) |
| **Channel MCP** | NOT dropped — **promoted to a documented Wire-3 alternative** after test-2 proved it E2E (0-token send, ~8 s round-trip). Constraints stand: interactive-only receiver, a persistent per-session MCP subprocess, dev-flag opt-in, an `inbox/` prompt-injection surface. Default stays SendMessage; see the Wire-3 transport matrix. | `gcn50-q12-channels.md`, `gcn50-e2e-channels.md` |
| **TUI injection (`orca terminal send`)** | keystroke injection with an incident-logged type-but-unsubmitted race. Last-resort break-glass only. | `gcn50-review-fable-tars-to-session.md §2 row 4` |
| **Resident warm relay (Agent SDK / long-lived `-p` daemon)** | the only alternative that wins an axis that matters (native held/denied/expired notices), but buys it with a resident stateful daemon the stateless-Tars axiom forbids. **Named upgrade path** if per-message cost or the ack gap ever hurts. | `gcn50-review-fable-tars-to-session.md §2 row 5` |

---

## Part A build tasks

1. **Skill v3: rewrite `delegate-to-cooper`** (Tars' own skill; landed via SOUL
   rule 2 self-merged PR) to the four-wire flow above. Retire the mailbox path.
2. **`tars-notify` helper** — Wire-2 rung-1: raw threaded `hermes send` +
   SESSION.md append + VM lane-mirror refresh, one ssh trip, `[lane:]` prefix.
3. **`lean-send` helper** — Wire-3 config-G invocation + ListAgents liveness +
   JSON parse + nonce + fenced-payload + pinned target; the fallback ladder.
4. **Launcher helper** — Wire-1 spawn with `-n tars-<ticket>[-rN]` + the
   post-spawn name+cwd registry check.
5. **The lane-index file** — `~/.hermes/lanes/<thread_ts>.md` writer (spawn +
   refresh), flock/tmp+mv/chmod 600; the SOUL standing "check before answering"
   rule; the Linear `(thread_ts, worktree, peer-name)` triple write.
6. **The LANE.md / SESSION.md contract** — single-writer roles, enforced same-turn
   write-back as a SOUL-level rule, read-only-on-LANE.md for Wire-2 turns.

**Two remaining build-time probes (cheap, before DoD):**
- **CLI-resume → next-wake row identity.** After a CLI `--resume` appends turns,
  does the *next gateway wake* on that thread see them? P3 proved gateway→CLI; the
  return direction is INFERRED. One send + one Slack reply + one `state.db` read.
- **ESCALATE / RESOLVED prompt discipline.** Mechanics proven (stdout capture,
  separate send); the prefix-following itself is prompt discipline, n=0.

**Definition of Done (#1) — live exercised round-trip.** Tars spawns a real
session; a mid-flight question reaches Gaetan in the lane thread; his answer flows
back (gateway wake → lane-index resolution → Wire-3 relay to cooper); completion is
reported. Evidence in `status/`. Gaetan sends exactly one real DM at acceptance.
**Status update:** the round-trip MECHANISM is now proven end-to-end (three E2E
tests, §E2E validation) — spawn, Wire-2, operator answer, Wire-3 relay, session
action, and completion all ran live with real parts. What the DoD still gates is
the AUTOMATION: the Wire-1 startup-gate hardening and, above all, the built Wire-4
inbound relay (lane-index resolution → Wire-3) on a real gateway wake — the one
composition never yet run without a human hand-playing Tars.

---

# PART B — Tars as master coordinator (ROADMAP SKELETON — not yet designed in detail)

> **[PROPOSED / ROADMAP]** — everything in Part B is direction, not design. It is
> here so the shape is on record; it is not build-ready and has no probe backing of
> its own beyond the fleet-compatibility review.

## Vision (Gaetan's words)

Long goal: Tars becomes a **master coordinator** that, **with Gaetan's approval**,
starts working on Linear tickets on its own — tickets delegated to Claude via Orca
using exactly the Tars↔session wiring in Part A. Each **lane** (like an Orca
worktree) gets its **own message + thread** in the Tars↔Gaetan DM. Tars reports
back to Slack when a Claude session asks the operator (Gaetan) a question, and Tars
understands the context: the ticket, the work the session(s) have done, where
they're at, and **why** the question is asked. Gaetan still personally drives some
Orca worktrees; Tars manages its own. A **per-ticket approval gate**: Tars proposes
to pick up a ticket, Gaetan approves, then Tars runs it.

## How Part A carries Part B

- **Lanes ↔ per-lane DM threads.** Part A already makes one lane = one worktree =
  one DM thread with a `thread_ts → lane` index. Part B is N of these running at
  once; the index's Linear triple is what enumerates "status of everything?".
- **"Understands context / where they're at / why the question."** This is exactly
  Part A's rehydration contract: the escalation turn resumes the thread's gateway
  session (persona + history) and is handed the inlined LANE.md (ticket, decision
  log, said-in-thread, current status). The "why" rides in SESSION.md's
  question+why field, written by the session on the Wire-2 rung-2 escalation.
- **Autonomous pickup behind the per-ticket gate.** Tars proposes a ticket in the
  DM (a new lane thread), waits for Gaetan's in-thread approval (Wire 4 inbound),
  then spawns (Wire 1) and runs it. The approval is per-ticket, never standing —
  the same per-message-approval shape SOUL already uses for posting.
- **Coexistence with Gaetan's manual worktrees.** Tars addresses only lanes it
  spawned, by their deterministic `tars-<ticket>` names; it never ListAgents-and-
  guesses at a session it didn't create. Gaetan's own worktrees are invisible to
  Tars's wires — no name collision because Tars-owned names are namespaced.

## Fleet integration via `/megaultracode-orca`

`/megaultracode-orca` is a project-scoped SKILL invoked *inside* an already-running
session that then becomes a fleet coordinator and spawns N workers. It is **not a
launcher**, so it cannot replace Wire 1. It composes *under* the design: Wire 1
spawns one `tars-<ticket>` session, whose LANE.md brief tells it to invoke
`/megaultracode-orca <scope>`. The coordinator IS the lane; the fleet is its
internal detail; **the fleet never surfaces to Tars**
`[analysis — gcn50-megaultracode-compat.md; verdict WORKS-WITH-ADAPTATIONS]`.

Three adaptations are **mandatory** (all [PROPOSED]):
1. **Composition rule.** A fleet-shaped delegation is still one Wire-1 spawn; the
   lane = the *scope* (milestone/epic), not per underlying ticket; the coordinator
   is the sole Tars-facing peer; workers are internal, never named in
   LANE.md / lane mirror / Linear. Member tickets are Linear children, not lanes.
2. **Worker-containment clause.** The coordinator must append to every worker brief
   an explicit deny: no `hermes`, no Slack, no LANE.md/SESSION.md/lane-mirror writes,
   no Wire-2/Wire-3 participation — report and escalate to the coordinator via
   SendMessage only. Otherwise you get N writers on the lane files, N `[lane:]`
   posts in one thread, N concurrent `--resume`s of one session row (the accepted-
   risk interleaving becomes the common case). The generic handoff-orca template
   does not contain workers by itself.
3. **Handoff-rename protocol.** A Tars-lane coordinator handoff must relaunch under
   the deterministic successor name `tars-<ticket>-rN` (or, minimum, the outgoing
   coordinator sends a Wire-2 rename notice + refreshes LANE.md + lane mirror +
   Linear triple *before* dying). Megaultracode coordinators are high-churn, so an
   auto-assigned handoff name would silently invalidate the pinned Wire-3 target on
   all four recorded surfaces and make a live lane falsely read as dead — the
   expected path on a long milestone, not an edge case.

**Side-blocker (before any fleet lane runs in this repo):** the in-repo skill copy
`.claude/skills/megaultracode-orca/SKILL.md` (331 lines) is **stale** vs the
399-line canonical (missing the merge-token protocol + later incident entries). A
coordinator spawned in a Tars worktree loads the stale copy — **sync or delete it
first** `[VERIFIED-by-inspection — gcn50-megaultracode-compat.md §1/§Secondary]`.

Untested residue (flag, don't block): no measured startup-context figure for a
megaultracode worker; coordinator inbound-`accept` verified only for orc-launched
coordinators (the composition rule guarantees that).

---

## Evidence index

Every file under `status/probes/`:

| File | One line |
|---|---|
| `gcn50-q10-kanban-smoke.md` | Kanban `notify-subscribe --thread-id` = RAW-POST-IN-THREAD (4 s), agent-wake gated on `task.session_id` which CLI `create` can't set. |
| `gcn50-q12-headless-sendmessage.md` | Headless `claude -p` called ListAgents + SendMessage, `success:true` + `msg_id`, no permission hold (under `--dangerously-skip-permissions`). |
| `gcn50-q12-channels.md` | Channel MCP push works INTERACTIVE-only; `-p` receiver FAILED 3/3 silent drop; research-preview, dev-flag dialog. |
| `gcn50-q12-nonclaude-socket.md` | Any uid-1000 non-Claude process injects into an `accept` session; token optional, ancestry not required; held under default config. |
| `gcn50-q14-w1-w2-comparison.md` | Session→Tars bake-off: `chat -Q -q \| send -t` (W2) beats kanban wake (W1); W2 = real turn, in-thread, rehydration nonce-proven, ~14.6 s; W1 needs W2 as precondition + posts twice. |
| `gcn50-q15-lean-sendmessage.md` | Leanest Wire-3 relay = config G: 3,870 tokens / 4.73 s, 8/8 delivered, −98.2 % tokens vs full-orchestrator baseline. |
| `gcn50-rehydrate-p1-inline.md` | Inlining LANE.md into the Wire-2 prompt carries full/deep context (1 call, ~31.4k tok, ~9.3 s); a cross-host path FAILS (5 calls, ~154.7k tok). |
| `gcn50-rehydrate-p2-threadindex.md` | No built-in `thread_ts → lane` lookup exists; cold hydrate once/≤31 msgs, warm = zero Slack reads ⇒ an index is REQUIRED. |
| `gcn50-rehydrate-p3-resume.md` | `hermes chat --resume <gateway_session_id>` headless resumes the thread's gateway session — history + French voice + pipe-safe, only a `state.db` row needed. |
| `gcn50-review-fable-tars-to-session.md` | Adversarial review, crux A (Wire 3): transport sound, verification narrative inflated, addressing + failure-path fixes (`-n`, nonce, ListAgents, JSON parse). |
| `gcn50-review-fable-session-to-tars.md` | Adversarial review, cruxes B/C/D: two-tier Wire 2, LANE.md write-back + atomicity, Linear triple, dropped-option kills sound. |
| `gcn50-rehydrate-review.md` | Closing adversarial pass: rehydration verdict MET-WITH-CONDITIONS; `--resume` restructures Wire 2 into three rungs; lane index required. |
| `gcn50-megaultracode-compat.md` | Fleet review: WORKS-WITH-ADAPTATIONS via composition (not substitution); three mandatory adaptations + stale in-repo skill side-blocker. |
| `gcn50-e2e-roundtrip.md` | Live E2E test-1 (SendMessage wire-3): full round-trip PROVEN (English); wire-1 startup-gate + wire-4 inbound-gap findings. |
| `gcn50-e2e-channels.md` | Live E2E test-2 (MCP Channel wire-3): PROVEN (French, 8.18 s, 0-token send); validated the spawn-outside-`$HOME` wire-1 fix. |
| `gcn50-e2e-go-bridge.md` | Live E2E test-3 (Go non-Claude peer): SEND+RECEIVE both work; resident bidirectional peer viable same-box-only; high version-fragility. |
| `gcn50-e2e-cleanup.md` | Test-1 teardown log; Orca "no `repo rm`" dangling-registration hygiene finding. |

Supporting operational truth (not GCN-50-specific): `docs/facts.md` (thread /
session-keying rows, gateway gating), `status/probes/2026-08-14-damien-followup-rootcause.md`
(GCN-49, the write-back failure class this design guards against).

---

## Open questions / decisions deferred

Noted broad-first; hardening deferred by design (workflow rule: no
security/guardrail hardening in the first push, but flag a control that would
lapse).

**0. Wire-3 messaging transport — OPEN, not yet chosen (Gaetan undecided).** All
three (+ raw socket) are E2E-proven; the pick is **deferred and swappable behind the
`lean-send` helper**, so it can be decided at build time and changed later without
touching the other wires. Trade-offs: **lean `claude -p` SendMessage** — official,
headless, works cross-host as `ssh cooper`, robust to Claude upgrades, but ~6.3k
tokens + a spawn per message; **MCP Channels** — 0-token send, fastest, but
interactive-only receiver + a persistent per-session MCP subprocess + dev-flag + an
inbox prompt-injection surface; **resident Go / custom peer** — no per-message spawn,
one bidirectional process, but same-uid/same-box-only (needs a cooper-resident
daemon) and rides an undocumented protocol (high version-fragility); **raw socket** —
cheapest, hand-rolled undocumented wire. Recommendation *if forced today*: SendMessage
to start, Go peer as the upgrade if per-message cost bites — but this is a lean, not a
decision.

1. **Exact skill-v3 structure** — one skill or a small family (`delegate`,
   `notify`, `lane-index`)? Where the helper scripts live (in-skill vs a
   `scripts/` dir mirrored to the VM/cooper). Undecided.
2. **Security hardening of the socket-injection surface.** The raw-socket escape
   hatch authorizes on same-uid only — any uid-1000 process (incl. any `ssh cooper`
   command) can inject into an `accept` session with no token/ancestry check. This
   is a real trust boundary widening; acceptable for a single-user dev box today,
   flagged as the control to revisit if cooper ever becomes multi-tenant.
3. **`crossSessionInbound: accept` as a standing default** on every orc-* session —
   convenient for Wire 3, but it is what makes the socket surface injectable. Keep
   or scope per-lane? Deferred.
4. **The two build-time probes** (resume→next-wake row identity; ESCALATE/RESOLVED
   discipline) are cheap but unrun — do them before or fold into the DoD?
5. **Restated rehydration bar** — Gaetan must explicitly accept "all
   decision-relevant context, under enforced write-back" in place of "ALL context"
   before Part A is called done.
6. **Multi-lane concurrency limits** — how many simultaneous lanes before the DM
   thread list, the lane-index directory, or the per-message ack timeouts need
   tuning? Part B question, undesigned.
7. **Version tripwire ownership** — the lean flag stack and the socket frame are
   2.1.232-shaped; who owns re-running q15 config G after each `claude` upgrade?
