# GCN-9 close + GCN-10 prune ticket

Run date: 2026-08-10 (VM clock UTC). Method reference:
`status/probes/wf6/gcn9-native-mcp.md`.

**VERDICT: both actions done via the native path, no fallback needed.**

- GCN-9 → **Done** (comment added first, then state-only update)
- New ticket **GCN-10** — "Prune the Linear MCP tool surface exposed to Tars"
  — Todo, assignee Gaetan, label `access`, priority 3 (Medium)

---

## Method used

**Native only**, both actions and both verification reads. Driven via
`~/.local/bin/hermes chat -Q -q '<instruction>'` over ssh, instructing the
agent to call `mcp__linear__save_comment` / `mcp__linear__save_issue` /
`mcp__linear__list_issues` with explicit parameters, then report the raw tool
result verbatim. The argv-safe `curl -K` fallback was never invoked — not
needed.

No secret was read, echoed, or placed on argv at any point in this session;
`LINEAR_API_KEY` was never touched (the native MCP path resolves auth
internally on the VM).

---

## ACTION 1 — close GCN-9

### Step 1: comment (`mcp__linear__save_comment`)

Instruction sent (verbatim body requested):

> The Linear MCP server is installed as a manual mcp_servers.linear stanza
> with static Bearer auth; the catalog preset was correctly rejected as
> remote OAuth 2.1 + DCR, unusable headless. hermes mcp test linear connects
> with 58 tools, and mcp__linear__* visibility was proven twice through the
> agent loop. NRestarts=0, and the slack and notion MCP configs were left
> untouched.

Raw tool result:

```json
{"id":"0ecc5dad-1740-4a8f-9bf6-a1efa93f6ea3","body":"The Linear MCP server is installed as a manual mcp_servers.linear stanza with static Bearer auth; the catalog preset was correctly rejected as remote OAuth 2.1 + DCR, unusable headless. hermes mcp test linear connects with 58 tools, and mcp__linear__* visibility was proven twice through the agent loop. NRestarts=0, and the slack and notion MCP configs were left untouched.","createdAt":"2026-08-10T17:15:03.775Z","updatedAt":"2026-08-10T17:15:03.621Z","parentId":null,"resolvedAt":null,"quotedText":null,"author":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain"}}
```

### Step 2: state update (`mcp__linear__save_issue`, id + state ONLY)

Raw tool result (relevant fields):

```json
{"id":"GCN-9","...","status":"Done","statusType":"completed","completedAt":"2026-08-10T17:15:10.727Z","labels":["access"],"assignee":"Gaëtan Cathelain","priority":{"value":1,"name":"Urgent"},"team":"Gaetan"}
```

`labels` came back as `["access"]` — unchanged, confirming the "pass only id
and state" instruction was honored (no label wipe).

---

## ACTION 2 — file GCN-10

Created via `mcp__linear__save_issue` (no `id` → create), single call, all
target fields passed at once (`team`, `assignee: "me"`, `title`, `state`,
`labels`, `priority`, `description`).

Raw tool result:

```json
{"id":"GCN-10","title":"Prune the Linear MCP tool surface exposed to Tars","description":"All 58 mcp__linear__* tools are currently enabled to the live agent, including destructive ones (delete_*, merge_diff). Before this wave Tars had no Linear write capability at all — the engagement-checker collector was read-only by construction — so this is a new capability surface, not a lapsed control. Deliberately not hardened in this push per the broad-first-restrict-later rule. The linear-ticketing skill's closed tool list (save_issue, save_comment, list_issues, get_issue) is currently the only blast-radius bound and it is policy, not enforcement. Investigate hermes mcp configure linear as the prune mechanism.\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":{"value":3,"name":"Medium"},"url":"https://linear.app/mobile-club/issue/GCN-10","createdAt":"2026-08-10T17:16:01.938Z","status":"Todo","statusType":"unstarted","labels":["access"],"assignee":"Gaëtan Cathelain","assigneeId":"4951b192-e49c-4b7e-b491-58c89e66043c","team":"Gaetan","teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e"}
```

Description's last line is exactly `Spec: Tars repo docs/specs/wf6-linear-integration.md (WF6)` — confirmed in
the raw JSON above.

**GCN-10 = new ticket key.**

---

## VERIFY — independent re-read (fresh `mcp__linear__list_issues`, not the mutation responses)

Queried `team="GCN", limit=50` (no state/assignee filter) twice: once
**before** either action, once **after** both. `hasNextPage: false` both
times — full board, 9 issues before, 10 after.

### Before/after board table

| id | title (short) | status before | status after | priority | labels | changed? |
|---|---|---|---|---|---|---|
| GCN-1 | Bootstrap GCN team | Done | Done | Urgent | access | no |
| GCN-2 | LINEAR_API_KEY scope check | Done | Done | Urgent | access | no |
| GCN-3 | engagement-checker Linear push+pull | In Progress | In Progress | High | fix | no |
| GCN-4 | daily-work-brief pull board | In Progress | In Progress | High | fix | no |
| GCN-5 | default-team rule | In Progress | In Progress | High | investigation | no |
| GCN-6 | cooper defaults | Done | Done | Medium | discovery | no |
| GCN-7 | WF6 E2E verification | In Progress | In Progress | High | test-check | no |
| GCN-8 | WF6 T6 verification probe | Canceled | Canceled | Low | (none) | no |
| GCN-9 | Wire Linear MCP natively | **In Progress** | **Done** | Urgent | access | **yes — target of Action 1** |
| GCN-10 | Prune the Linear MCP tool surface | (did not exist) | **Todo** | Medium | access | **yes — created by Action 2** |

All `updatedAt` timestamps for GCN-1 through GCN-8 are identical between the
before and after reads (e.g. GCN-3/4/5/7 all `2026-08-10T16:...Z`, unchanged) —
confirmed nothing besides GCN-9 and the new GCN-10 moved.

---

## Note on the native agent-loop path (evidence for T7)

Clean, with one early friction point:

- `mcp__linear__list_issues` rejected a `fields: [...]` array on the first
  attempt (`Invalid input, fields.0/fields.2`) — dropping the `fields`
  parameter entirely fixed it immediately. The tool's `fields` parameter
  shape isn't what a plain string-array guess produces; worth noting in the
  schema reference for next use, but a one-line fix in practice.
- `save_comment` and `save_issue` both worked first try with exactly the
  parameters from the referenced probe's schema table (`id`, `state` only for
  the state-only update; full param set for the create).
- The "pass only id and state, don't touch labels" instruction was followed
  correctly — labels came back unchanged in both the mutation response and
  the independent re-read.
- Multi-line markdown descriptions with an em-dash and no escaped `\n`
  round-tripped correctly (matches the server's own `instructions`: send real
  newlines, not `\n` literals).
- Every call returned in single-digit seconds; no retries, no session drops.

Overall: reliable enough to drive multi-step Linear writes (comment → state
update → create → re-read-verify) in one directive without hand-holding
individual tool calls beyond specifying exact parameters.
