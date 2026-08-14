# HANDOFF — GCN-50 Tars↔Orca orchestration (Tars V2)

**Date:** 2026-08-14 · **Ticket:** GCN-50 (team GCN, High) · **Branch:** pushed to `main`.
Written by the design session for whoever continues — human or agent.

## TL;DR
The **design phase is complete, adversarially reviewed, and the mechanism is proven
end-to-end live**. **Nothing is built into Tars yet.** #1 (Tars↔session
communication) is fully specced and build-ready. #2 (Tars as autonomous master
coordinator) is a roadmap skeleton, not yet designed in detail.

## Read these first (canonical, committed)
- **`PROPOSED-SPEC.md`** (repo root) — THE design. Part A = build-ready four-wire
  Tars↔session comms; Part B = #2 roadmap. Carries VERIFIED/INFERRED/PROPOSED tags
  and an evidence index. **Start here.**
- `scratchpad/gcn50-design-final.md` — the locked design the spec is the readable
  form of (scratchpad may be ephemeral; the spec is authoritative).
- `status/probes/gcn50-*.md` — all evidence (probes + 3 adversarial reviews + 3 E2E
  runs + cleanup). Indexed at the bottom of `PROPOSED-SPEC.md`.
- `ORCHESTRATION-POLICY.md` + `CLAUDE.md` (repo root) — how to work here
  (orchestrate/delegate, Opus-vs-Sonnet rule) and the Tars hard rules.
- Skill `operating-tars` + `skills/delegate-to-cooper/SKILL.md` — the CURRENT (v1)
  delegation mechanism this work replaces.

## The design in one breath (settled — see spec, don't re-litigate)
Tars = **stateless coordinator**. One delegation = one Orca worktree = one lane =
one Slack DM thread. Four wires:
- **Wire 1 Spawn** — orc-launcher, `-n tars-<ticket>`, Tars-aware brief.
- **Wire 2 Session→Tars** — 3 rungs: raw `hermes send` (routine status) /
  `hermes chat --resume <gw_session> … | hermes send` with **LANE.md inlined**
  (question/escalation/completion — the turn *is* Tars) / cold `chat -q` (no gateway
  session yet).
- **Wire 3 Tars→session** — lean `claude -p` SendMessage (**default**); MCP Channels
  and a resident Go non-Claude peer are **proven alternatives** (transport matrix in
  spec); raw-socket escape hatch.
- **Wire 4 Operator gate** — Gaetan replies in-thread → gateway wakes → reads the
  `~/.hermes/lanes/<thread_ts>.md` index → relays inward via Wire 3.
State: Linear = ticket record; LANE.md (Tars, single-writer, enforced same-turn
write-back) + SESSION.md (session). Only deliverables + evidence are committed.

## Proven vs UNBUILT (the actual remaining work)
**Proven live:** every wire's mechanism; the full round-trip end-to-end via all 3
Wire-3 transports (E2E tests 1/2/3); rehydration MET-WITH-CONDITIONS; `--resume`
restores Tars persona + thread context headless.

**Unbuilt — this is what "build #1" means:**
1. **Wire-4 inbound relay — TOP ITEM.** gateway-wake → lane-index lookup → Wire-3
   relay. A human hand-played this in every test; it has never been composed in code.
2. **Wire-1 spawn-gate handling.** Startup trust gates (folder-trust +
   external-CLAUDE.md-import) block autonomous spawn; `--dangerously-skip-permissions`
   does not cover them and the brief-seed drops. Validated fix: spawn the worktree
   **outside `$HOME`** + confirm-submit. Build into the launcher.
3. **LANE.md same-turn write-back** as a SOUL-level rule (the GCN-49 replay guard).
4. The **lane-index writer** + the Linear `(thread_ts, worktree, peer-name)` triple.

## Build tasks (spec §Part A build tasks)
1. **Skill v3:** rewrite `delegate-to-cooper` to the four-wire flow, retire the
   mailbox path. (Tars lands its own skills via SOUL-rule-2 self-merged PR; other
   code goes through a delegated session.)
2. `tars-notify` helper (Wire-2 rung-1).
3. `lean-send` helper (Wire-3 config-G + ListAgents liveness + nonce + app-ack + ladder).
4. Launcher helper (Wire-1 + spawn-gate mitigation + name/cwd registry check).
5. Lane-index writer + SOUL "check before answering" rule + Linear triple.
6. LANE.md / SESSION.md contract + write-back SOUL rule.
Plus **2 cheap build-time probes** first: CLI-resume→next-gateway-wake row identity;
ESCALATE/RESOLVED prompt discipline.
**DoD:** a live round-trip using the BUILT Wire-4 relay (not hand-run); Gaetan sends
one real DM at acceptance.

## Needs Gaetan (decisions before/at build)
- **Accept the restated rehydration bar:** "all *decision-relevant* context, under
  enforced write-back" (literal "ALL context" is unachievable by this shape). Spec
  open-question #5.
- Other spec open-questions: skill-v3 structure; socket-injection hardening (deferred
  broad-first); `crossSessionInbound: accept` default scope; multi-lane concurrency
  limits; version-tripwire ownership.
- **#2 (Part B) is not yet designed** — it needs its own grill/design pass before build.

## Gotchas / operational facts
- **Secrets/config hard rules:** never whole-file `sops -d` (per-key `--extract` over
  ssh only); `.bak` + `flock ~/.hermes/.wf3.lock` + `chmod 600` on every VM config
  edit; no secret on argv/echo/evidence. See `CLAUDE.md`.
- **Reach:** `ssh cooper` (192.168.0.4 — Orca sessions) · `ssh gaetan@192.168.0.9`
  (Tars VM), Hermes CLI `~/.local/bin/hermes` (not on PATH non-interactively). Slack
  DM `D0BBYNM01BL`, Gaetan `U08BDJAMSRZ`, Tars bot `U0BBH85NAKH`.
- **Measure before scripting:** CLI facts in specs are LLM-summarized — run `--help`
  on the box first. Read Slack via the MCP with `include_bots=true` or bot/thread
  messages are invisible.
- **Orca:** no `repo rm` verb → scratch-repo registrations dangle; reuse a small
  lane-repo pool instead of one-per-ticket.
- **Go bridge (test-3) scratch:** `scratchpad/gcn50-go-bridge/` (may be cleaned; full
  Go source is inside `status/probes/gcn50-e2e-go-bridge.md`). Same-uid/same-box only,
  high version-fragility — nonce self-test on every `claude` upgrade if productized.
- **Fleet (`/megaultracode-orca`)** composes UNDER Wire 1 (it's a skill run inside a
  session, not a launcher); 3 mandatory adaptations in spec Part B; a stale in-repo
  skill copy (331 vs 399 canonical lines) must be synced before any fleet lane.

## How to work in this repo
Orchestrate, don't implement (`ORCHESTRATION-POLICY.md`): delegate to **Opus**
(judgment/impl/review) and **Sonnet** (grunt/probes), set model explicitly on every
spawn. Every delegated unit writes a file + returns a short summary; verify artifacts
before claiming done. Commit + push evidence as you go:
`git pull --rebase origin main && git push origin HEAD:main` (this worktree can't
check out `main`).

## Session note
The account **session limit** was hit once during this work (reset ~17:20 UTC) — pace
long build sessions accordingly.

## Suggested next step
Get Gaetan's sign-off on the restated rehydration bar (open-q #5), then start the
build with the two load-bearing items — **Wire-4 inbound relay** and the **Wire-1
launcher** — delegated per the orchestration policy, running the 2 build-time probes
first.
