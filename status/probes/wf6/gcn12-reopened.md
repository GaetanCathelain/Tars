# GCN-12 — reopened per Gaetan's instruction

**Why**: GCN-12 ("Create and send the Verdict Google OAuth client" — a real
open commitment, Jamil's OAuth client for Gaetan) was filed by
engagement-checker, then closed to Done purely to verify that closing a
Linear issue clears it from the reminder queue (WF6 spec-check-2). Gaetan
instructed it be reopened afterward: the commitment is still live, the
closure was a test artifact, not real completion.

Run date: 2026-08-10 (VM clock UTC). Tool: native `mcp__linear__*` via
`~/.local/bin/hermes chat -Q -q '<instruction>'` on 192.168.0.9, per
`status/probes/wf6/gcn9-native-mcp.md`.

---

## Before state (verbatim, via `mcp__linear__get_issue`)

```json
{
  "identifier": "GCN-12",
  "title": "Create and send the Verdict Google OAuth client",
  "state": {
    "name": "Done",
    "id": "0434e579-7b85-487a-8cf9-5aed6caaf41b",
    "statusType": "completed"
  },
  "labels": [
    { "name": "fix", "id": "bd21ce6f-2ff9-4850-9f81-d54fc48524f0" }
  ],
  "priority": { "number": 3, "label": "Medium" },
  "assignee": {
    "name": "Gaëtan Cathelain",
    "id": "4951b192-e49c-4b7e-b491-58c89e66043c"
  },
  "team": { "name": "Gaetan", "id": "81e7b769-2a46-4e2a-8db5-c165a7963b0e" }
}
```

## Mutation call

`mcp__linear__save_issue` with **exactly two parameters**, no others:

```
id: "GCN-12"
state: "Todo"
```

No `labels` param passed (would have replaced/wiped the set), no `priority`,
no `assignee`, no `team`.

### Mutation response (raw)

```json
{
  "id": "GCN-12",
  "title": "Create and send the Verdict Google OAuth client",
  "priority": { "value": 3, "name": "Medium" },
  "status": "Todo",
  "statusType": "unstarted",
  "completedAt": null,
  "labels": ["fix"],
  "assignee": "Gaëtan Cathelain",
  "assigneeId": "4951b192-e49c-4b7e-b491-58c89e66043c",
  "team": "Gaetan",
  "teamId": "81e7b769-2a46-4e2a-8db5-c165a7963b0e"
}
```

## After state — independent re-read (not the mutation response)

Fresh `mcp__linear__get_issue` call, separate `hermes chat -Q` session
(`session_id: 20260810_211849_65870d`):

```
Identifier: GCN-12
Title: Create and send the Verdict Google OAuth client
State: Todo
  - ID: 59ed732a-f242-4eba-926f-1c0d128fe83c
  - statusType: unstarted
Labels: fix — bd21ce6f-2ff9-4850-9f81-d54fc48524f0
Priority: 3 — Medium
Assignee: Gaëtan Cathelain — 4951b192-e49c-4b7e-b491-58c89e66043c
Team: Gaetan — 81e7b769-2a46-4e2a-8db5-c165a7963b0e
```

**Confirmed**: `statusType` is `unstarted` (was `completed`). Label set
(`fix`, same id), priority (`3`/`Medium`), assignee (Gaëtan Cathelain, same
id) and team (Gaetan, same id) are **unchanged** from the before state — the
re-read proves the mutation did not silently no-op and did not touch any
other field.

## Comment added

`mcp__linear__save_comment`, `issueId: "GCN-12"`, comment id
`c96dddd4-5fdc-4b39-92f1-8921a6dd1753`:

> This issue was closed on 2026-08-10 solely as part of WF6 spec-check-2
> verification, to confirm that closing a Linear issue clears it from the
> reminder queue. The underlying commitment (sending the Verdict OAuth
> client to Jamil) is still live, so Gaetan instructed it be reopened.

## Rest of board unchanged

Full `mcp__linear__list_issues` read of team Gaetan (12 issues,
`hasNextPage: false`) after the mutation — GCN-1..GCN-11 all at their
expected, untouched states:

```
GCN-1  — Bootstrap GCN team + workload labels + priority view      — Done       (completed)
GCN-2  — LINEAR_API_KEY scope check on the VM                      — Done       (completed)
GCN-3  — engagement-checker: Linear push+pull, SSoT=GCN, nag-loop   — In Progress (started)
GCN-4  — daily-work-brief: pull Linear board + render priority-sort — In Progress (started)
GCN-5  — Tars default-team rule: create tickets in GCN by default   — In Progress (started)
GCN-6  — cooper defaults: sessions create Linear tickets in GCN     — Done       (completed)
GCN-7  — WF6 E2E verification + docs/facts + status log updates     — In Progress (started)
GCN-8  — WF6 T6 verification probe — disposable, ignore              — Canceled   (canceled)
GCN-9  — Wire Linear MCP server natively into Hermes                — Done       (completed)
GCN-10 — Prune the Linear MCP tool surface exposed to Tars           — Todo       (unstarted)
GCN-11 — WF6 probe — disposable [EC-ZZ99]                            — Canceled   (canceled)
GCN-12 — Create and send the Verdict Google OAuth client             — Todo       (unstarted)  ← reopened
```

Only GCN-12 was touched by this session. No other issue's state, labels,
priority, or assignee was modified.

## Constraints observed

No secret printed/echoed/logged; no `sops` invoked; no secret on argv. No
git command run. No file touched under `skills/`. No engagement-checker
cycle run. Only this one new file written.
