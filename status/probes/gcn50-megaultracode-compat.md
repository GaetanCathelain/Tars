# GCN-50 × `/megaultracode-orca` — compatibility review

**Date:** 2026-08-14 · **Reviewer:** Fable 5, adversarial pass.
**Inputs read in full:** locked design (`scratchpad/gcn50-design-final.md`, v2 post-rehydrate),
`/megaultracode-orca` discovery digest (`scratchpad/megaultracode-orca-digest.md`), probes
`gcn50-q12-headless-sendmessage.md`, `gcn50-q12-nonclaude-socket.md`, `gcn50-rehydrate-review.md`.

## 0. The category error to clear first

`/megaultracode-orca` is **not a launcher**. It is a project-scoped SKILL
(`disable-model-invocation: true`) invoked **inside an already-running Claude Code
session**, which then becomes a fleet **coordinator** and spawns N worker sessions via
`delegate-orca` → `orc-tab start … orc-opus …`. So "spawn the lane via /megaultracode-orca
instead of orc-tab" is not a substitution that exists: something must already be a session
before the skill can run. The only coherent composition is:

> **Wire 1 as designed** (Tars → ssh cooper → launcher, `-n tars-<ticket>`) spawns ONE
> session; the LANE.md brief tells THAT session to invoke `/megaultracode-orca <scope>`.
> The coordinator IS the lane session. The fleet is its internal implementation detail.

This maps surprisingly well onto the skill's own stated design — "THIS session staying as
the single coordinator that relays questions to the operator" is exactly the single-funnel
shape GCN-50 needs. Everything below is assessed under this composition; the alternative
(Tars spawning workers directly through the skill's machinery) is DOA on naming and
Tars-awareness and is not further considered.

## 1. Wire-by-wire

### Wire 1 — Spawn: **NEEDS-ADAPTATION**

- The skill can't be the launcher (above). Under the compose shape, Wire 1 is untouched:
  Tars spawns `-n tars-<ticket>` via the orc-* pair, deterministic name, post-spawn
  registry check — all preserved, because the coordinator is a perfectly ordinary
  orc-opus session.
- **No clobbering:** workers and coordinator get the *same*
  `~/.claude/handoff/ORCHESTRATION-POLICY.md` appended (verified from the launcher
  scripts); the megaultracode brief is a *user message*, not a system prompt, so the
  Tars lane brief coexists rather than being overwritten.
- **But the workers are not Tars-aware and are not deterministically named.** Worker
  session names are auto-assigned (`slug → session name [ref]` roster kept by the
  coordinator). That is fine *only if* no Tars wire ever targets a worker. The design
  must say so explicitly.
- **N-tickets→M-sessions vs 1-lane=1-worktree=1-thread:** a `/megaultracode-orca` scope
  is a milestone/epic. The mapping survives only if the LANE is defined as the *scope*
  (one Tars delegation = one coordinator = one thread = one LANE.md in the coordinator's
  worktree), not per underlying ticket. Workers live in child subworktrees — outside the
  lane's state files. Declare this in the spec: fleet lane = scope-level lane; the
  member tickets are tracked in Linear as children, not as lanes.
- **Stale skill copy blocker:** the copy checked into THIS repo
  (`.claude/skills/megaultracode-orca/SKILL.md`) is 331 lines vs 399 canonical — missing
  the merge-token protocol and later incident entries. A coordinator spawned in a Tars
  worktree loads the STALE copy. Sync (or delete) it before any fleet lane runs here.

### Wire 3 — Tars→session: **NEEDS-ADAPTATION**

- **Inbound accept:** verified — every orc-opus/orc-fable worker AND the coordinator
  (spawned via wire 1's orc launcher) carry `crossSessionInbound:"accept"`. The
  q12 socket probe's finding holds for the whole fleet. No hold-dialog risk on the
  coordinator, contra the skill's own Phase-0 warning (that warning applies only to
  coordinators started as bare interactive sessions — ours isn't).
- **Single addressable peer:** yes, IF the rule is "Tars messages only the coordinator."
  The fleet does not break "message the lane session" because the skill itself is
  hub-and-spoke: workers already report to the coordinator over SendMessage; the
  coordinator's inbox is the lane's inbox. Tars must never ListAgents-and-guess at
  workers — a milestone fleet plus 12 unrelated peers (measured roster) makes
  name-guessing exactly the ambiguity `-n tars-<ticket>` was introduced to kill.
- **The real break: coordinator handoff.** The digest verifies a mid-run handoff creates
  a brand-new peer identity (new auto-assigned name+ref). The GCN-50 design pins the
  target name in the wire-3 relay's `--system-prompt`, in LANE.md, in the lane mirror,
  and on the Linear ticket. A handoff silently invalidates all four; the next wire-3
  send fails the ListAgents liveness check and Tars declares a live lane dead. And
  megaultracode coordinators are *high-churn* — the skill has its own context-ceiling/
  handoff sections, so this is the expected path on a long milestone, not an edge case.
  **Adaptation (mandatory):** the handoff procedure for a Tars-lane coordinator must
  relaunch under the deterministic successor name `tars-<ticket>-rN` (the design's
  existing respawn convention) — or, minimum, the outgoing coordinator sends a wire-2
  rename notice and refreshes lane mirror + LANE.md + Linear triple *before* dying.
  Without this, wire 3 is broken by design on any lane long enough to need the skill.

### Wire 2 — Session→Tars: **COMPATIBLE**

- The coordinator is a normal Claude Code session with full Bash; nothing in the skill
  or the launchers removes shell/ssh. `ssh gaetan@192.168.0.9 '… hermes chat --resume …
  -q "$(cat)"' < prompt.txt` and the raw `hermes send -t` rung work identically. It is
  not headless-only; workers have shells too (irrelevant — they must not use wire 2,
  see Wire 4).
- One semantic remap needed in the brief: the skill's "relay questions to the operator"
  assumes a human at the invoking terminal. In a Tars lane there is none — the
  coordinator's operator gate IS wire-2 rung 2 (resumed gateway turn → Slack thread).
  One paragraph in the lane brief; mechanics unchanged.
- Escalation funnel actually *improves*: workers escalate to the coordinator (skill's
  native flow), the coordinator triages and escalates the residue via wire 2 — fewer,
  better-formed questions reaching Gaetan than N independent sessions would produce.

### Wire 4 + lane index + LANE.md single-writer: **NEEDS-ADAPTATION**

- **Single-writer holds only by explicit containment.** Lane state (LANE.md Tars-written,
  SESSION.md session-written, VM lane mirror refreshed on every wire-2 send) must be
  touched by the **coordinator only**. The generic handoff-orca brief template is not
  Tars-aware (digest, INFERRED-from-absence but consistent with the skill text): if the
  Tars lane brief is naively forwarded to workers, you get N writers on SESSION.md/lane
  mirror, N `[lane:]` posts into one thread, N concurrent `--resume`s of the same
  gateway session row (the accepted-risk interleaving becomes the common case, not the
  rare one). **Adaptation (mandatory):** every worker brief the coordinator generates
  must carry an explicit deny: no `hermes`, no Slack, no LANE.md/SESSION.md/lane-mirror
  writes, no wire-2/wire-3 participation; report and escalate to the coordinator via
  SendMessage only. The coordinator aggregates into SESSION.md and speaks with one voice.
- **1-lane=1-thread survives** as 1-*scope*=1-thread (Wire 1 note). The lane index entry
  `~/.hermes/lanes/<thread_ts>.md` gains one field: the roster snapshot is NOT mirrored
  there (churny, worker names are internal); only the coordinator peer name goes in —
  same schema as today. Gaetan's in-thread reply resolves to the lane → wire-3 to the
  coordinator → coordinator routes to the relevant worker. Composition unchanged.
- Per-ticket threads for one fleet (N lanes ↔ 1 session) would violate the design's
  lane↔peer mapping and force ESCALATE routing to carry lane tags — reject that shape;
  one thread per invocation.

### Rehydration / state: **COMPATIBLE** (one pressure point)

- Stateless Tars, inline-LANE.md wire-2 prompts, same-turn write-back: all unaffected —
  the coordinator does exactly what a solo session would, just about a bigger scope.
- Pressure point: LANE.md as "distillate, not log" (rehydrate review, accepted risk #5)
  is much easier to blow through when one lane covers a milestone's worth of decisions.
  The size ceiling was measured at 2.3 KB with arithmetic-only extrapolation to ~23 KB.
  Keep per-ticket detail in Linear/SESSION.md; LANE.md carries scope-level decisions
  only. Restate this in the fleet-lane brief, don't rely on the general rule.
- Completion semantics: wire-2 completion fires once, when the *fleet* is done (or the
  coordinator hands off/aborts) — not per ticket. Per-ticket closure is Linear traffic.

## 2. Bottom line

**WORKS-WITH-ADAPTATIONS — via composition, not substitution.** `/megaultracode-orca`
cannot replace the wire-1 launcher (it isn't one), but it slots *under* the design
cleanly because it is itself hub-and-spoke with a single relaying coordinator: spawn the
lane session exactly as designed, let it invoke the skill, and treat the coordinator as
the lane peer for all four wires. The fleet never surfaces to Tars. It is NOT a
fundamentally different shape needing a distinct integration — provided the three
adaptations below are made mandatory; without #3 in particular, every long fleet lane
will falsely go lane-dead at the coordinator's first context handoff.

### Required adaptations (ranked)

1. **Composition rule in the spec + lane brief:** a fleet-shaped delegation is still one
   wire-1 spawn (`-n tars-<ticket>`); the brief instructs the session to run
   `/megaultracode-orca <scope>`; lane = scope; coordinator = sole Tars-facing peer;
   workers are internal and never named in LANE.md/lane mirror/Linear.
2. **Worker containment clause:** coordinator must append to every worker brief an
   explicit prohibition on hermes/Slack/lane-file/wire participation (single-writer,
   single-voice); the generic handoff-orca template does not do this by itself, and the
   coordinator's operator gate is remapped from "human at terminal" to wire-2 rung 2.
3. **Handoff rename protocol:** a Tars-lane coordinator handoff must yield the
   deterministic successor name `tars-<ticket>-rN` (or at minimum a pre-death wire-2
   rename notice + same-turn refresh of LANE.md, lane mirror, and the Linear triple).
   Auto-assigned handoff names break the pinned wire-3 target on all four recorded
   surfaces.

Secondary before first use: **sync or delete the stale in-repo skill copy** (331-line
`.claude/skills/megaultracode-orca/SKILL.md` vs 399-line mc-metarepo canonical) — a
coordinator in a Tars worktree loads the stale one.

Untested residue (flag, don't block): no measured startup-context figure for a
megaultracode worker (digest NOT FOUND); Claude Code ≥ 2.1.224 precondition already met
on cooper (2.1.232 per q12 probe); coordinator inbound-accept verified only for
orc-launched coordinators — the composition rule guarantees that, keep it.
