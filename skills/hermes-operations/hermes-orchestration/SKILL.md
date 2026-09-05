---
name: hermes-orchestration
version: 1.0.0
description: "Use when choosing Hermes orchestration primitives or limits."
---

# Hermes orchestration field guide

Use this skill to select, explain, or audit Hermes Agent orchestration. The public docs are authoritative; re-check them when current behavior matters.

## Source baseline

Verified 2026-08-08 against Hermes Agent docs/repository commit `52920747e17dc5e8ab18a41e4084ed28508b4e6c`.

Primary references:
- https://hermes-agent.nousresearch.com/docs/user-guide/features/delegation
- https://hermes-agent.nousresearch.com/docs/user-guide/features/kanban
- https://hermes-agent.nousresearch.com/docs/user-guide/features/goals
- https://hermes-agent.nousresearch.com/docs/user-guide/features/cron
- https://hermes-agent.nousresearch.com/docs/user-guide/messaging/slack

## Choose the primitive

| Need | Primitive | Key boundary |
|---|---|---|
| One direct action | Normal tool call | No separate worker |
| 3+ mechanical calls with loops/filtering | `execute_code` | No LLM reasoning inside script |
| Reasoning-heavy isolated subtask or parallel batch | `delegate_task` | Process/session-local execution; child starts with fresh context |
| Separate ad-hoc session while chat remains free | `/background` (`/bg`, `/btw`) | Fresh isolated session; not a durable workflow queue |
| Keep one chat iterating until acceptance criteria pass | `/goal` | Single session; no fan-out or board task |
| Long shell command | `terminal(background=true, notify=true)` | Shell process, not an agent workflow |
| Scheduled or restart-independent fresh agent run | `cronjob_manage` (discover current schema) | Fresh session; scheduling authority remains scoped |
| Durable multi-profile workflow with dependencies/retries/handoffs | Kanban | Single-host SQLite board |
| External event starts work | Webhook | Trigger surface; pair with cron/Kanban/agent logic as needed |

## Delegation

- `delegate_task` spawns isolated children; only final summaries return to the parent context.
- Top-level dispatches return immediately and post completion later where async delivery is supported. Stateless endpoints may fall back to synchronous execution.
- Batch children run concurrently; default cap is 3 via `delegation.max_concurrent_children`. Oversized batches error instead of truncating.
- Children inherit the parent's enabled toolsets and cannot widen access per call.
- Leaf children cannot call `delegate_task`, `clarify`, `memory`, `send_message`, or `cronjob`; both leaf and orchestrator roles retain `execute_code`.
- Default depth is flat (`max_spawn_depth: 1`). `role="orchestrator"` only gains effect when depth is raised. Cost grows multiplicatively with width × depth.
- Default child iteration budget is 50. Default wall-clock timeout is disabled; optional `child_timeout_seconds` adds one.
- Background children have progress-based stall detection (quiet thresholds differ inside/outside tools), live transcripts, `/agents` monitoring, kill/pause controls in the TUI, and scoped steering support.
- Durability distinction: completion delivery is persisted if the child already finished, but in-progress execution does not resume after process/session death. A lost attempt becomes `unknown`; side effects may already have happened.
- Verify child claims independently (diff, tests, URL, file stat). Child summaries are reports, not proof.

## Persistent goals

- `/goal` is a Ralph-style continuation loop in the current session, judged after every turn.
- It does not create, assign, or move Kanban cards.
- Default budget is 20 turns; pause/resume/clear are supported.
- Prefer explicit completion contracts: outcome, verification, constraints, boundaries, stop condition.
- Deterministic shell quality gates must pass before the LLM judge can mark done.
- Goal loops can park on background process/session completion or a timed wait instead of polling wastefully.
- A Kanban card with goal mode borrows the same loop inside that card's worker; it remains separate from the chat's `/goal` state.

## Kanban and graph engineering

Hermes can perform graph-style engineering workflows as a durable task DAG:

- Tasks have one assignee/profile and statuses `triage | todo | ready | running | blocked | done | archived`.
- Parent→child links are cycle-checked dependencies. Children auto-promote when all parents are done.
- Fan-out, fan-in, pipelines, voting/quorum, reviewer gates, human-in-the-loop, journals, and fleets are supported patterns.
- `hermes kanban swarm` builds root/blackboard → parallel workers → verifier → synthesizer.
- Triage auto-decomposition can turn a one-line objective into an LLM-produced task graph routed by profile descriptions. Auto mode is on by default; manual decompose/specify is available.
- The gateway-embedded dispatcher automatically sweeps boards (default every 60s), promotes eligible work, atomically claims tasks, and spawns the assigned profile. Thus ready tasks can be picked up automatically without a human polling.
- Workers use `kanban_*` tools, not shell CLI calls. They must orient with `kanban_show`, work in the assigned workspace, heartbeat on long work, and end with complete or block.
- Durable handoffs include comments, parent summaries/metadata, run history, events, attachments, artifacts, and retry history.
- Failure handling includes crash detection, stale reclaim, runtime limits, protocol-violation retries, spawn-failure circuit breakers, recurrent-block escalation, and respawn guards.
- Concurrency can be capped globally and per profile. Scheduled start times are supported.
- Boards are hard isolation boundaries with separate DB/workspaces/logs. Tenants are soft filters/namespaces.
- Workspaces: disposable scratch, preserved directory, or preserved git worktree.
- Human control works through dashboard, CLI, `/kanban`, REST, comments, block/unblock, assignment, and live events.
- `/kanban` is explicitly allowed mid-run and can inspect or mutate the board without interrupting the active agent.

### Kanban limits

- Built-in Kanban is single-host. The local SQLite board and PID-based crash detection do not coordinate workers across hosts.
- Tasks must resolve to installed profiles; profile descriptions drive auto-routing quality.
- External CLI lanes (Codex/Claude/OpenCode as native dispatcher lanes) are not yet paved; they require a plugin/integration around the lane contract.
- Scratch workspaces are deleted on completion except declared artifacts, which are copied to durable attachment storage.
- Dashboard plugin routes rely on localhost binding; exposing the dashboard on `0.0.0.0` exposes task data and mutation routes to the network.
- Auto-decomposition is model judgment, not a formal workflow compiler. Explicit dependencies and acceptance criteria remain important.

## Cron and durable automation

- Cron supports one-shot, interval, cron-expression, and ISO schedules; pause/resume/edit/run/remove.
- Jobs can load ordered skills, set a project workdir, restrict toolsets, run scripts, monitor URLs/scripts for changes, chain prior outputs via `context_from`, and deliver to origin/local/platform targets.
- `no_agent=true` runs a script only and delivers stdout verbatim; empty stdout is silent.
- Workdir-bound jobs serialize because workdir state is process-global; workdir-less jobs may run in parallel.
- Continuable jobs seed their delivery into a chat/thread so replies have the brief in context.
- Preflight checks credentials, skills, and delivery configuration; misconfigured jobs block without spending model tokens.
- Cron sessions cannot recursively create cron jobs.

## Slack interaction limits relevant to orchestration

Separate Slack UI limitations from Hermes gateway behavior:

- Slack Agent/Assistant DM UI on mobile may disable the composer while an AI response is in progress. Hermes cannot override Slack's client UI. This is consistent with Slack's own in-progress Agent view, not evidence that Hermes cannot receive concurrent input.
- At the gateway layer, messages arriving while a turn runs are serialized by session guards; depending on busy-input mode they interrupt, steer, or queue. The default CLI mode is interrupt, but platform UI may prevent the message from being sent at all.
- Use `/btw <prompt>` (`/background`) for isolated ad-hoc work so the main session is not occupied by that job; use Kanban/Cron for durable detached work.
- `/kanban` bypasses the running-agent guard, so board inspection and mutation can execute mid-turn when Slack allows the slash command to be sent.
- Native Slack slash commands do not work inside thread replies; use `!queue`, `!stop`, `!kanban ...`, etc. as regular thread messages.
- Slash-command responses are ephemeral and capped/chunked; long Kanban output is truncated around gateway message limits.
- In channels, Hermes normally needs an initial mention and replies in a thread; established thread sessions may continue without a fresh mention. 1:1 DMs are mention-exempt; group DMs are shared surfaces and follow channel controls.
- Messages from other bots are ignored by default. `allow_bots: mentions` is the recommended loop-safe peer-agent mode; `all` risks bot loops.

## Capability preflight and completion evidence

Before saying “unavailable” or asking for login, inspect current tool discovery,
installed CLI/help and existing authenticated API/CLI access. Prefer a working
supported route over a new login; on “retry”, recheck because tools and auth may
have changed. Name the actual failed check if blocked. Preserve human-only
authentication steps; never bypass access controls or change provider/model
defaults, profiles or permissions to make preflight pass.

For delegation, check the selected runtime is reachable, its account usable,
and the effective host/model complies with current restrictions. On Cooper,
Codex is unavailable by standing policy; use the authorized Claude workflow.
Inspect board/dispatcher/profile limits only when that primitive is involved.
Distinguish source support, installed release support, configured state and
actual delivered behavior; source code alone does not prove a binary feature.

Keep these evidence states in the existing task record, not another tracker:

- **queued**: dispatch/input accepted; no task execution observed yet;
- **executing**: substantive task work observed beyond a submitted prompt;
- **blocked**: the actual failed check or pending decision is identified;
- **implemented**: requested artifact exists and its relevant checks passed;
- **deployed**: intended installation received and activated that artifact;
- **verified**: every requested acceptance item passed on the delivered system.

Before dispatch/retry, reconcile current request identity, scope, existing
worker and completed effects. Status questions, old roots and result events do
not authorize duplicate execution. Unknown effects require inspection first.
Before “done”, independently read back exact external targets and verify all
acceptance items, including requested ticket state, cleanup and durability.
Preserve material worker warnings as unverified until checked. Verify exact
links before supplying them; if access blocks verification, say so rather than
guessing a URL. Procedural checks do not prove runtime replay protection.
