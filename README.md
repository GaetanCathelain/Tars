# Tars

Personal pro assistant for Gaetan — a dedicated [Hermes agent](https://hermes-agent.nousresearch.com/)
running on its own Proxmox VM, interacting through Slack (DM-only — the reporting channel
retired 2026-08-13, responds only to Gaetan),
wired into everything Gaetan touches professionally: Slack, Gmail, Linear, GitHub (incl.
mc-metarepo), Notion, Calendar, the machines on the tailnet, and Orca on the cooper VM.

Tars orchestrates and reports; implementation deliverables are the delegated session's job, not
Tars's — but read-only analyses and reports it writes itself (SOUL rule 1). Not a limit on what
Tars is trusted with. Work is delegated to Claude Code sessions
driven through Orca on cooper, which Tars drives, tracks and verifies. Tars never merges, approves
or pushes: a PR authored by a session Tars delegated to is fine; Tars authoring one is not.

**Status: LIVE** since 2026-08-07 — built, cut over, and verified 15/15
(`status/wf4-report.md`). Repo guide for working sessions: `CLAUDE.md`;
operational facts: `docs/facts.md`.

| Doc | What |
|---|---|
| [PITCH.md](PITCH.md) | the original project pitch (verbatim) |
| [PLAN.md](PLAN.md) | the run graph: phases, lanes, execution model, coordination contract |
| [docs/graph.html](docs/graph.html) | visual of the run graph (self-contained page) |
| [docs/graph-engineering-research.md](docs/graph-engineering-research.md) | research: graph-run orchestration landscape, 2026-08 |
| [docs/conversation-2026-08-07.md](docs/conversation-2026-08-07.md) | the design conversation that produced this plan |

`status/` will hold the per-lane execution logs once the build starts (single writer per file —
see PLAN.md § Coordination).
