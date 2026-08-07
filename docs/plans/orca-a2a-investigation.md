# Peer mandate: Orca integration v2 + A2A investigation

From Gaetan, 2026-08-07 (~20:50Z). Two lanes, one peer session; Gaetan will
also work interactively in this session — treat his in-session input as the
authority when it arrives.

## Lane 1 — better Orca integration (v2 of the delegation playbook)

v1 shipped tonight: `docs/specs/wf5-orca-delegation.md` — plain ssh into the
`~/orca/workspaces/tars-delegated/` sandbox via `delegate.sh` (fixed flags,
deny-rules). The v1 recon (`status/probes/wf5/orca-recon.md` — READ IT, do not
redo it) deliberately rejected the Orca session layers for v1: stale handles,
no completion signal, no durable coordinator state across sessionless ssh
calls. v2's job is to solve exactly those:

1. Investigate driving REAL Orca sessions from Tars: `orca worktree create`,
   `orc-tab start`, the ListAgents/SendMessage session mesh, status/result
   retrieval. What does a completion signal look like? Can Tars poll or be
   notified (its cron tool? A2A push once lane 2 lands?)?
2. Design + implement the smallest v2 that lets Tars: spawn a named Orca
   session on cooper for a task, check its status, retrieve its result, and
   report — with the v1 guardrail philosophy (enforced not requested: fixed
   entrypoints Tars cannot flag-inject; Tars still never implements code
   itself and never creates PRs — SOUL + README law).
3. Upgrade the skill `~/.hermes/skills/delegate-to-cooper/SKILL.md` and the
   spec (append a v2 section, keep v1 as fallback). Live E2E via Slack
   (authorized, as Gaetan, minimal messages: the WF4/WF5 evidence files show
   the exact working patterns — conversations.replies, env names, IDs).

## Lane 2 — A2A investigation (Tars ↔ cooper, Tars ↔ p-Hermes)

Current state (PLAN.md WF5 row + probe 12): Tars serves an inbound agent card
on `127.0.0.1:9900` (`hermes-tars`, v1.0.0); the outbound A2A toolset is
disabled (D2 YAGNI); p-Hermes A2A state unknown; cooper runs no Hermes A2A
endpoint (Orca sessions speak a different protocol — lane 1's mesh).

1. Inventory: p-Hermes's A2A surface via the `phermes` ssh leg — READ-ONLY
   (curl its localhost:9900 from within VM 103 via ssh command execution; do
   not modify anything on p-Hermes — proposals for p-Hermes changes go to
   Gaetan, e.g. as a prompt he pipes, like the cutover pattern). Tars's
   outbound toolset: what enabling it provides (source reading).
2. Decide transport per PLAN.md: SSH-tunneled localhost vs tailnet bind vs
   keep-disabled — with guardrails (what a remote agent may ask Tars; what
   Tars may ask p-Hermes). Write the decision doc.
3. Probe a real round-trip each way where safely possible TODAY (Tars→p-Hermes
   read-only card fetch through an ssh tunnel is fair game; anything needing
   p-Hermes config changes stops at a written proposal).

## Contract

- Evidence: `status/probes/wf5/orca-v2-*.md`, `status/probes/wf5/a2a-*.md`.
- Docs: `docs/specs/wf5-a2a.md` (new), v2 section appended to
  `docs/specs/wf5-orca-delegation.md`, and an "A2A" section appended to
  `docs/facts.md` (a sibling peer is concurrently appending a "Thread
  behavior" section to the same file — always `git pull --rebase origin main`
  before pushing, and push with `git push origin HEAD:main`).
- Coordination: the `thread-behavior` peer is live on the same VM/gateway.
  Config edits use `flock ~/.hermes/.wf3.lock` as everywhere; avoid gateway
  restarts unless genuinely required, keep them short, and note them in
  evidence. Never edit `status/lane-a.md` (single-writer: the hub).
- All hard rules in `CLAUDE.md` + `docs/facts.md` bind you (secrets, flock,
  never root on 192.168.0.3, `--help` before scripting, measure before
  believing recon).
- SendMessage the hub (Tars orchestrator session, via ListAgents) at
  milestones: lane-1 design chosen, lane-1 E2E verdict, lane-2 inventory,
  lane-2 decision + any p-Hermes proposal, done.
