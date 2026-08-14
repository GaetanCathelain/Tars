# GCN-50 rehydration/persona — final adversarial review (closes crux C)

**Date:** 2026-08-14 · **Reviewer:** Fable 5, adversarial pass.
**Inputs read in full:** locked design (`scratchpad/gcn50-design-for-review.md`); prior reviews
`gcn50-review-fable-tars-to-session.md` (crux A) and `gcn50-review-fable-session-to-tars.md`
(cruxes B/C/D); new probes `gcn50-rehydrate-p1-inline.md`, `gcn50-rehydrate-p2-threadindex.md`,
`gcn50-rehydrate-p3-resume.md`. All three probe files present and complete — none failed.

---

## 0. Do the new probes actually support their verdicts?

| Probe | Claim | Audit |
|---|---|---|
| P1 REHYDRATE-FULL | Inlining LANE.md's full text into `hermes chat -Q -q` (ssh-stdin `$(cat)` transport) carries a buried mid-file decision, a mid-file nonce, and a third-section blocker into one correct composed line; 1 API call, ~31.4k tok (~30.5k fixed floor + ~800 marginal), 9.3s. Path-only control fails honestly cross-host (5 calls, ~154.7k tok, 29.1s). | **HOLDS.** Nonce + D3 substance ("seeded from worker_id", "incident replay") could not come from a top-of-file echo; `tool_turns=0` proves no side-channel. Caveats P1 itself flags and I confirm: n=1 at 2.3 KB; the ≥23 KB extrapolation is arithmetic, not measurement; the 31-message gateway cap is a different subsystem and P1 says nothing about it. The path-only control is the single most useful negative result in the file set: **a cooper path is dead on arrival on the VM — the prior review's hole C-1 (cross-host read unproven) is closed by removing the cross-host read, not by proving it.** |
| P2 index REQUIRED | No CLI verb, session-store field, or wake-payload field maps thread_ts→lane; cold hydrate is once-per-thread-per-process, capped 31 messages; warm thread = zero Slack reads; a woken in-thread turn without an index answers blind. | **HOLDS.** Source-verified (`build_session_key`, `adapter.py:5793-5905`, seeding test docstring) and corroborated by the `history=` jump pattern across 9 sessions. This independently re-derives prior-review hole C-2/C-3 from source, upgrading them from "unstated" to "structurally impossible without an index". The kanban `kanban_notify_subs.thread_id` near-miss is correctly ruled out (kanban is dead per D, and the trigger is wrong anyway). |
| P3 RESUME-RESTRUCTURES | `hermes chat --resume <id> -Q -q` headless over plain ssh resumed a real **gateway** (`source=slack`) thread session: correct topic recall from injected history (1 API call, ~8.5s), reply in French operator register against an English query, pipes into `hermes send -t` unchanged; no TTY, no live process, only a `state.db` row. | **HOLDS, with its own honest asterisks.** (a) Persona effect is a **single sample** and P3 correctly labels the mechanism INFERRED (SOUL.md loads cold too; what resume adds is the thread's established register winning over language-mirroring). (b) The cold control is **confounded** — `session_search` self-served context off the literal "gcn50" keyword — so the 2× speed/reliability comparison is directional, not clean; P3 says so itself. (c) Resume **mutates the live session row** (`ended_at=NULL` + appended turns) and the concurrent-writer race with a live gateway turn is unprobed. None of these overturn the verdict: the load-bearing facts (headless works, gateway-origin session resumable, history injected, in-thread delivery) are all VERIFIED with command output. |

No probe overclaims its evidence. P1's transport note matters for the spec: the ssh-stdin
`'… -q "$(cat)"' < prompt.txt` shape is what dodges multi-KB quote-escaping — mandate it,
don't let an implementer "simplify" to an argv-embedded string.

---

## 1. VERDICT on Gaetan's bar — "rehydrates ALL context"

**MET-WITH-CONDITIONS.**

The prior review's NOT-MET rested on six holes. Where they stand now:

1. **Cross-host read (was the largest measured-vs-asserted gap): CLOSED — by restructure.**
   P1 proves inline works fully and path-reference cannot work at all. The design changes
   from "the VM turn reads LANE.md on cooper" to "the cooper session inlines LANE.md into
   the wire-2 prompt". No unprobed mechanism remains on this hole.
2. **thread_ts→lane reverse index: CONFIRMED REQUIRED (P2, from source), design now concrete
   (§3).** Closed on paper; the index itself is not yet built — condition.
3. **Wire-2 posts invisible to the answering Tars / 31-message cap: SUBSTANTIALLY CLOSED for
   the tier that matters.** Under §2's shape, question/escalation/completion turns run *inside*
   the thread's own gateway session via `--resume`, so their content lands in the very
   transcript the gateway uses when Gaetan replies. Residual: raw status sends stay invisible
   to warm sessions — acceptable because they are also written to SESSION.md and the lane
   mirror (the files, not Slack, are the record). One INFERRED link remains: that the
   gateway's next wake reuses the same `state.db` row the CLI resumed (P3's target row *was*
   gateway-created, so the row is shared in one direction; the return direction is unprobed —
   see §5).
4. **LANE.md currency / GCN-49 replay risk: STILL A CONDITION, not evidence.** No probe can
   prove a discipline. The same-turn write-back contract (any turn that says/hears anything
   decision-relevant in a lane thread updates LANE.md + the lane mirror in that turn) must be
   in SOUL/standing prompts, and the GCN-49 root-cause file shows what happens without it.
   The cron nudge (wire-3 timeout guard) must anchor via the lane index — the GCN-49
   wrong-anchor class is exactly a cron turn answering without conversation-state resolution.
5. **Atomicity/locking: CONDITION with known remedy** — repo doctrine (flock + tmp+mv +
   chmod 600) applied to every LANE.md and lane-mirror write. P3 adds a new instance:
   concurrent `--resume` vs live gateway turn on the same `state.db` row — sqlite serializes
   writes, so the credible failure is transcript interleaving, not corruption; degrade
   accepted, noted in §5.
6. **"ALL" is still not literal.** What rehydrates is *what was written*: prior tool results,
   in-flight nuance, and thread text beyond the 31-message cold window are lost by
   construction. The prior review's restatement stands and Gaetan must accept it explicitly:
   **"all decision-relevant context, provided the write-back contract is enforced"** — that
   bar the evidenced design now meets; the literal one no design of this shape can.

Conditions, explicitly: (a) restated bar accepted by Gaetan; (b) same-turn write-back as a
hard SOUL-level contract; (c) lane index built as §3; (d) the §5 must-probe list run or
accepted as DoD coverage.

---

## 2. Does `--resume` restructure wire 2? — YES

P3 proves the escalation turn can *be* Tars: the thread's own gateway session, its history,
its French operator register, headless, dead-process-safe, and pipe-compatible with the
proven `send -t` half. This dissolves the prior review's B-2 (two-minds-in-one-thread) for
the tier where it mattered, and folds cleanly into that review's two-tier recommendation.
**New recommended wire-2 shape (three rungs):**

1. **Routine status / milestones — raw `hermes send -t slack:D0BBYNM01BL:<thread_ts>`.**
   No agent turn, no LLM, ~1s, exact text, prefixed `[lane:<name>]` (honest telemetry, not
   fake Tars). Same ssh trip also appends the line to SESSION.md and refreshes the VM lane
   mirror (§3). Unchanged from prior review B-1.
2. **Question / escalation / completion — resumed turn:**
   ```
   sid=$(ssh vm 'python3 …state.db mode=ro… "SELECT id FROM sessions
         WHERE chat_id=? AND <thread_ts col>=? AND ended_at IS NULL … LIMIT 1"')
   out=$(ssh vm '~/.local/bin/hermes chat --resume '"$sid"' --no-restore-cwd -Q -q "$(cat)"' < prompt.txt)
   ```
   where `prompt.txt` = event + **inlined LANE.md** (P1 shape — resume restores the thread's
   conversation, not the lane's file state; both are needed) + the routing instruction
   (below). The session then routes stdout: `ESCALATE:` prefix → second ssh,
   `hermes send -t slack:D:<thread_ts>`; `RESOLVED:` prefix → the answer is already in the
   session's hands (it captured stdout), no Slack post — this fixes the prior review's
   "pipe unconditionally posts" defect with machinery q14's tee shape already proved.
   `--no-restore-cwd` because the recorded cwd of a gateway session is meaningless to a
   headless relay.
3. **Cold fallback — P1 inline shape (`hermes chat -Q -q "$(cat)"`, no resume) — only when
   the `state.db` lookup returns no row** (a lane thread no gateway turn has ever run in:
   possible early in a lane whose anchor was a raw `send` and Gaetan hasn't replied yet).
   Voice pinned in the prompt; accepted as the degraded mode P1 measured.

The operator-gate fallback question is therefore moot: resume is available and headless.
Had it not been, the fallback would have been rung 3 with LANE.md carrying a
"said-in-thread" section — still workable, measurably worse on voice and continuity.

---

## 3. thread_ts→lane index — REQUIRED; minimal concrete design

Required per P2 (source-verified: nothing else can resolve an in-thread wake to a lane, and
a cooper path is unreadable from the VM per P1). Design, ladder-minimal:

- **Primary (the wake path): one VM-local file per lane, `~/.hermes/lanes/<thread_ts>.md`.**
  Content is *inlined substance*, never a bare cooper path: ticket id, worktree path (for
  humans/relays), **peer name `tars-<ticket>[-rN]`** (crux-A review §4), decision log,
  said-in-thread section, current status — i.e. a mirror of LANE.md, same convention the
  repo already runs for `SOUL.md`/`skills/`.
- **Writers:** Tars creates it at spawn (it is on the VM; it just posted the anchor and holds
  the thread_ts). Thereafter the **lane's cooper session** refreshes it over ssh in the same
  trip as every wire-2 send (tmp+mv + chmod 600 + flock, per hard rules) — single writer per
  lane file after spawn. Gateway turns read it; the gateway turn that receives a Gaetan
  answer appends to its said-in-thread section same-turn (the one two-writer touchpoint;
  flock covers it).
- **Standing SOUL rule:** any in-thread wake checks `~/.hermes/lanes/<thread_ts>.md` before
  answering; absence ⇒ say "no lane bound to this thread", never guess from prose.
- **Secondary (durable + enumeration): the same triple (thread_ts, worktree path, peer name)
  written to the Linear ticket at spawn** — one native `mcp__linear__*` write. Covers
  multi-lane "status of everything?" turns and survives VM loss. Not the wake path (the file
  is one read_file call, no MCP round-trip).
- **Not needed:** any new thread_ts→session_id machinery — `state.db`'s `sessions` table
  already carries chat_id + thread ts (P3 §2 queried it directly); that is the resume lookup.
  The kanban `kanban_notify_subs` route stays dead with kanban.

---

## 4. Consolidated design: rehydration + persona + operator gate (spec-ready)

**State.** Linear = system of record (ticket carries thread_ts/worktree/peer-name triple).
Worktree: LANE.md (Tars-authored brief/decisions/answers), SESSION.md (session-authored
status/blockers/question+why) — uncommitted, single-writer each, flock + tmp+mv on every
write, wire-2 turns read-only on LANE.md. VM: `~/.hermes/lanes/<thread_ts>.md` mirror (§3).
Slack is a display surface, not a record: cold hydrate is once, capped at 31 messages, warm
threads re-read nothing (P2, source-verified).

**Spawn (wire 1).** Launcher passes `-n tars-<ticket>[-rN]` (crux-A §4 fix); Tars records the
intended name in LANE.md + lane mirror + Linear before first use; one post-spawn registration
check (name AND cwd match) per crux-A §6.2. Tars anchors the lane thread with a raw
`hermes send`, records thread_ts everywhere in the same turn.

**Session→Tars (wire 2), three rungs as §2:** raw threaded send for status (+SESSION.md +
mirror refresh, `[lane:]` prefix); `--resume`d gateway-session turn with inlined LANE.md for
question/escalation/completion, ESCALATE/RESOLVED stdout routing by the session;
cold P1-inline turn only when no gateway session row exists. Never quote 14.6s (or 8.5s) as
the design latency — both are trivial-compose floors.

**Tars→session (wire 3), unchanged from crux-A §6:** lean `claude -p` relay (q15 config G) +
ListAgents in `--tools`, target name pinned in `--system-prompt`, payload as opaque fenced
block, per-message nonce, "do not reply to this message" line, JSON envelope parsed,
application-level ack = the session's next wire-2 message quoting the nonce, cron-owned
timeout → one re-send (new nonce) → raw-socket rung → lane-dead + respawn offer. The cron
nudge resolves its anchor through the lane index — never composes thread-blind (GCN-49 class).

**Operator gate (wire 4).** Session hits a question → wire-2 rung 2: resumed turn poses it
in-thread, in Tars' own session/voice, and the exchange is durably in that session's
transcript. Gaetan answers **in-thread** (top-level DM = new sessionless thread — known trap,
tell Gaetan once). Gateway wakes the thread session — which now already contains the
escalation turn — resolves the lane via `~/.hermes/lanes/<thread_ts>.md`, appends the answer
to LANE.md + mirror same-turn, and relays via wire 3. Answer reaches the session; session
acks via its next wire-2 message.

**Persona.** Escalation/completion turns are Tars by construction (resumed thread session,
P3). Status lines are labeled telemetry and don't pretend. The only voice-degraded path is
the cold rung-3 fallback, which is rare and prompt-pinned.

---

## 5. Remaining must-probe / accepted-risk items

1. **MUST-PROBE (and the DoD already is it): the inbound half live** — Gaetan's real
   threaded reply → gateway wake → lane-index resolution → gateway turn runs the wire-3
   relay over ssh to cooper. Zero coverage to date; every component is individually proven,
   the composition is not. Fine to let DoD #1 be the test — but everyone must know that is
   what it is.
2. **MUST-PROBE (cheap, before DoD): row-identity round trip** — after a CLI `--resume`
   appends turns, does the *next gateway wake* on the same thread see them? P3 proves
   gateway→CLI sharing; CLI→gateway is INFERRED. One send + one Slack reply + one
   `state.db` read.
3. **CHEAP PROBE OR ACCEPT: ESCALATE/RESOLVED routing discipline** — mechanics proven
   (stdout capture, separate send), the prefix-following itself is prompt discipline, n=0.
4. **ACCEPTED RISK (state once in the spec):** concurrent resume vs live gateway turn on
   one `state.db` row — worst credible outcome is interleaved transcript, not corruption;
   escalations fire when the session is blocked, so overlap with an active Gaetan chat in
   the same thread is rare.
5. **ACCEPTED RISK:** LANE.md size ceiling — P1's marginal-cost curve is measured only at
   2.3 KB; linearity to ~23 KB is arithmetic. Keep LANE.md a distillate, not a log
   (SESSION.md and Linear hold history), and this never binds.
6. **NOTE:** P2's observed dual session-key shapes (with/without team scope, post-cutover
   drift) — irrelevant to new lanes, but don't `--resume` into pre-2026-08-13 threads.

**Bottom line:** ALL-context bar **MET-WITH-CONDITIONS** (restated as decision-relevant +
enforced write-back; cross-host hole closed by P1's inline restructure); `--resume`
**restructures wire 2** (three-rung shape, §2); the thread index is **required** and its
minimal design is a VM-local per-thread mirror file + one Linear write (§3).
