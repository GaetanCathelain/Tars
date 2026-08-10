# GCN-9 ground-truth audit — Tars VM (192.168.0.9)

Run: 2026-08-10, ~17:00-17:10 UTC. All checks live via ssh, read-only.

## Claim 1 — linear: mcp_servers stanza + preceding .bak

VERIFIED. `~/.hermes/config.yaml` mcp_servers block:

```
mcp_servers:
  linear:
    url: https://mcp.linear.app/mcp
    headers:
      Authorization: "Bearer ${LINEAR_API_KEY}"
  slack: ...
  notion: ...
```

Static header auth via `${LINEAR_API_KEY}` env-var reference, no `auth: oauth`
stanza anywhere in the file (`grep oauth` = no hits).

Two `.bak-*` files carry today's date:
- `config.yaml.bak-boundaries-20260810-143638` (mtime 14:36:38.765, unrelated
  edit — slack `allowed_channels`/`require_mention` persona formatting, no
  `linear:` content)
- `config.yaml.bak-wf6-20260810T165931Z` (mtime 14:36:38.903 — inherited via
  `cp -p` from config.yaml's mtime at copy time, i.e. config.yaml was
  untouched between the boundaries edit and this backup)

Diffing `bak-wf6-20260810T165931Z`'s mcp_servers section against current
config.yaml shows **only** the 4-line `linear:` stanza added — nothing else
changed. This is the `.bak` that immediately precedes the edit.

## Claim 2 — `hermes mcp test linear`

VERIFIED. Ran live://
```
Transport: HTTP → https://mcp.linear.app/mcp
  Authorization: Bear***C6XB
✓ Connected (1973ms)
✓ Tools discovered: 58
```
5 example tools: `get_attachment`, `prepare_attachment_upload`,
`create_attachment_from_upload`, `delete_attachment`, `list_agent_skills`.
Tool count matches claim exactly (58).

## Claim 3 — gateway untouched

VERIFIED. `systemctl --user show hermes-gateway.service -p NRestarts,ActiveState`:
```
NRestarts=0
ActiveState=active
```
No restarts, still active — the config live-reload did not bounce the process.

## Claim 4 — nothing deployed to skills yet

VERIFIED. Pulled both live VM skill files and diffed against the repo mirror
at `/home/gaetan/dev/Tars/.claude/worktrees/kanban-research/skills/*/SKILL.md`
(that checkout is on `main`, confirmed via `git branch --show-current` /
`git log -1`):
- `orchestration/engagement-checker/SKILL.md` — byte-identical
- `orchestration/daily-work-brief/SKILL.md` — byte-identical

`find ~/.hermes/skills -iname '*linear-ticketing*' -o -iname '*linear*ticket*'`
on the VM returned nothing — no `linear-ticketing` skill dir exists yet, as
expected.

## Claim 5 — Linear board reality

VERIFIED via the VM's argv-safe GraphQL pattern (staged script over scp,
`.env` sourced remotely, key handed to curl via `-K` stdin — never on argv).
Team `81e7b769-2a46-4e2a-8db5-c165a7963b0e` (GCN), full issue list:

| Identifier | Title | State |
|---|---|---|
| GCN-11 | WF6 probe — disposable [EC-ZZ99] | Canceled |
| GCN-10 | Prune the Linear MCP tool surface exposed to Tars | Todo |
| **GCN-9** | Wire Linear MCP server natively into Hermes (catalog preset linear) | **Done** |
| GCN-8 | WF6 T6 verification probe — disposable, ignore | Canceled |
| GCN-7 | WF6 E2E verification + docs/facts + status log updates | In Progress |
| GCN-6 | cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default | Done |
| GCN-5 | Tars default-team rule: create tickets in GCN unless told otherwise | In Progress |
| GCN-4 | daily-work-brief: pull Linear board + render priority-sorted | In Progress |
| GCN-3 | engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard | In Progress |
| GCN-2 | LINEAR_API_KEY scope check on the VM | Done |
| GCN-1 | Bootstrap GCN team + workload labels + priority view | Done |

GCN-9 exists and is Done, matching the claim. GCN-3/4/5 are all still
"In Progress" (not started/claimed as done by this worker — consistent, since
those are separate follow-on tickets, not part of GCN-9's scope). GCN-7 is
also In Progress (this audit's own tracking ticket, presumably).

## Claim 6 — no cron/config changes beyond mcp_servers

PARTIAL / mostly VERIFIED, with one note.

`config.yaml.bak-*` full listing — only the two files above are dated today;
the rest (`bak-b5-*`, `bak-kanban`, `bak-linear-notion-cal`, `bak-mesh-model`,
`bak-notionpin`, `bak-slack`) are all from 2026-08-07, unrelated to this
ticket.

`hermes cron list` shows **6** active jobs, not 3:
1. Gaetan daily work brief — `30 8 * * 1-5` (known)
2. Gaetan engagement checker — `*/30 10-16 * * 1-5` (known)
3. Gaetan engagement checker final pass — `0 17 * * 1-5` (known)
4. Forward new Claude invoices to Olivier — `30 9 * * *`
5. mc-metarepo-refresh — `every 60m`
6. Daily report thread reconciler — `25,55 7-16 * * 1-5`

The 3 known jobs are present with unchanged schedules. Jobs 4-6 are
pre-existing, unrelated automations (invoice forwarding, metarepo KB refresh,
report-thread reconciliation) — none reference Linear or GCN-9, and none look
newly added for this ticket. Flagging only because the worker's evidence
apparently described "3 known jobs" without mentioning these 3 others exist;
they are out of scope for GCN-9 but the auditor should not assume "3 jobs
total" as ground truth going forward.

## Summary

All 6 claims verified against live VM state. No secrets printed (Authorization
header shown only as VM-masked `Bear***C6XB`; env var names only elsewhere).
