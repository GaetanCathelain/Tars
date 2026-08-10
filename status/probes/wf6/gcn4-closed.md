# GCN-4 — moved to Done

`id: 23b14dd2-f1c3-4ed3-8cd7-20bebc49762b` (issue UUID; the MCP surface itself
addresses it as `"GCN-4"`)

"daily-work-brief: pull Linear board + render priority-sorted"

Driven natively on the Tars VM through `~/.local/bin/hermes chat -Q -q
'<instruction>'` → `mcp__linear__*` tools, per
`status/probes/wf6/gcn9-native-mcp.md`. Run date: 2026-08-10 (VM clock UTC).

Method: each phase (baseline read, comment + read-back, state move, independent
re-read, board pull) was one oneshot `hermes chat` turn, dispatched by a small
Python script (`subprocess.run([...])`, argv list — no shell quoting of the
comment text) copied to the VM, run with `python3`, then deleted. No secret was
printed, echoed or persisted; no `sops` command was run; no git command was run;
no repo file outside this deliverable and `docs/facts.md` was touched.

---

## Before (baseline `get_issue`, fresh call)

```json
{"id":"GCN-4","title":"daily-work-brief: pull Linear board + render priority-sorted","status":"In Progress","statusType":"started","priority":{"value":2,"name":"High"},"assignee":"Gaëtan Cathelain","labels":["fix"],"url":"https://linear.app/mobile-club/issue/GCN-4"}
```

## Comment posted (`save_comment`, `issueId` + `body` only)

commentId: `3bb725dc-905c-489c-8caa-868e426d5a39`

> The agent-rendered board fabricated content (7 of 12 titles invented,
> terminal issues shown as live, "+2 more" against a true remainder of ~46)
> after context compaction evicted the raw reads; the VM copy was rolled back
> the same evening as mitigation. The fix moves rendering out of the LLM:
> `scripts/linear_board.py` computes the block and the agent includes it
> byte-for-byte, with a Coverage gate that prints "board unavailable (coverage
> unproven)" rather than any partial board. Verified twice under the exact
> failure condition - spec check 3 via `hermes chat` (compacted twice) and
> through the real cron path via a temporary mirror job (compacted once), both
> byte-identical to the script's direct output; 0 of 205 terminal issues
> leaked. Production cron jobs verified unchanged field-by-field.

Verified by a separate `list_comments` read-back (not the mutation response),
compared programmatically against the sent string:

```
VERBATIM MATCH: True
len sent/readback: 797 797
```

## State move (`save_issue`, `id` + `state` ONLY — no `labels`, no priority, no assignee)

`state` passed as the Done state id `0434e579-7b85-487a-8cf9-5aed6caaf41b`.
Raw tool return (abridged to the decisive fields):

```json
{"id":"GCN-4","status":"Done","statusType":"completed","completedAt":"2026-08-10T22:13:25.297Z","startedAt":"2026-08-10T16:40:23.787Z","canceledAt":null,"priority":{"value":2,"name":"High"},"labels":["fix"],"assignee":"Gaëtan Cathelain","assigneeId":"4951b192-e49c-4b7e-b491-58c89e66043c","team":"Gaetan","teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e"}
```

Per the standing rule, this payload is **not** treated as proof — a normal
payload is exactly what an unresolvable `state` also returns.

## After (independent re-read, `get_issue`, separate fresh session)

session `20260810_221350_afe85f`:

```json
{"id":"GCN-4","title":"daily-work-brief: pull Linear board + render priority-sorted","status":"Done","statusType":"completed","priority":{"value":2,"name":"High"},"assignee":"Gaëtan Cathelain","labels":["fix"]}
```

**Confirmed**: `statusType` `started` → `completed` (`status` `In Progress` →
`Done`). `labels` (`["fix"]`), `priority` (`{"value":2,"name":"High"}`),
`assignee` (`Gaëtan Cathelain`) and `title` all byte-identical to baseline.

---

## Final GCN board (`list_issues`, team=GCN, includeArchived=false, limit=100)

`HASNEXTPAGE=false COUNT=12` — completeness proven by the only valid signal.

| Key | Title | Status | statusType | Priority |
|---|---|---|---|---|
| GCN-1 | Bootstrap GCN team + workload labels + priority view | Done | completed | 1 |
| GCN-2 | LINEAR_API_KEY scope check on the VM | Done | completed | 1 |
| GCN-3 | engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard | Done | completed | 2 |
| GCN-4 | daily-work-brief: pull Linear board + render priority-sorted | **Done** | **completed** | 2 |
| GCN-5 | Tars default-team rule: create tickets in GCN unless told otherwise | Done | completed | 2 |
| GCN-6 | cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default | Done | completed | 3 |
| GCN-7 | WF6 E2E verification + docs/facts + status log updates | In Progress | started | 2 |
| GCN-8 | WF6 T6 verification probe — disposable, ignore | Canceled | canceled | 4 |
| GCN-9 | Wire Linear MCP server natively into Hermes (catalog preset linear) | Done | completed | 1 |
| GCN-10 | Prune the Linear MCP tool surface exposed to Tars | Todo | unstarted | 3 |
| GCN-11 | WF6 probe — disposable [EC-ZZ99] | Canceled | canceled | 3 |
| GCN-12 | Create and send the Verdict Google OAuth client | Todo | unstarted | 3 |

### Scope confirmation

- **Only GCN-4 changed.** Every other row is identical — key, status and
  statusType — to the board recorded at the end of
  `status/probes/wf6/gcn3-gcn5-closed.md`, which is the last full read before
  this run. GCN-4 is the single diff (`In Progress`/`started` →
  `Done`/`completed`).
- **GCN-12 is non-terminal**: `status: Todo`, `statusType: unstarted` — not
  closed, per the instruction that it is a live commitment.
- **GCN-7 and GCN-10 untouched**: still `In Progress`/`started` and
  `Todo`/`unstarted`. No tool call in this run referenced either.
