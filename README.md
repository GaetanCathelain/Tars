# Tars

Personal pro assistant for Gaetan — a dedicated [Hermes agent](https://hermes-agent.nousresearch.com/)
running on its own Proxmox VM, interacting through Slack (DM-first, responds only to Gaetan),
wired into everything Gaetan touches professionally: Slack, Gmail, Linear, GitHub (incl.
mc-metarepo), Notion, Calendar, the machines on the tailnet, and Orca on the cooper VM.

Tars orchestrates and reports; it never implements. Coding work is delegated to cooper/Orca.
No PR is ever created by Tars.

**Status: planning.** Build not started.

| Doc | What |
|---|---|
| [PITCH.md](PITCH.md) | the original project pitch (verbatim) |
| [PLAN.md](PLAN.md) | the run graph: phases, lanes, execution model, coordination contract |
| [docs/graph.html](docs/graph.html) | visual of the run graph (self-contained page) |
| [docs/graph-engineering-research.md](docs/graph-engineering-research.md) | research: graph-run orchestration landscape, 2026-08 |
| [docs/conversation-2026-08-07.md](docs/conversation-2026-08-07.md) | the design conversation that produced this plan |

`status/` will hold the per-lane execution logs once the build starts (single writer per file —
see PLAN.md § Coordination).
