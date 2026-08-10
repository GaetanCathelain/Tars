# `fields` union + `TeamIssues` — EXECUTED, not inferred

Run date: 2026-08-10 (VM clock UTC), 18:33–18:47Z, on the Tars VM
(`192.168.0.9`). Closes two review findings that flagged inferred-but-never-run
behaviour:

1. the `fields` members our skills request were pinned by an **enum/error
   probe** (`mcp-linear-measured-shapes.md` §Q4) — the offending-index trick —
   but the members our skills actually pass were never run together as a real
   projection;
2. `engagement-checker` §4's `TeamIssues` GraphQL operation had **never been
   executed once**, and its collector fails closed, so a bad filter would
   degrade silently.

Method. Probe 1: `~/.local/bin/hermes chat -Q -q '<instruction>'`, agent
instructed to make exactly one `mcp__linear__list_issues` call and echo the raw
tool result byte for byte. Probe 2: raw HTTPS to `https://api.linear.app/graphql`
with the query string lifted verbatim from `skills/engagement-checker/SKILL.md`,
credential passed to `curl` via `-K` on stdin (never argv, never printed).

**Read-only. No issue was created, updated, deleted or commented on. No
mutation was sent.** GCN's 11 issues and their `updatedAt` values are unchanged
from the end-of-run state recorded in `mcp-linear-measured-shapes.md` §Cleanup.

---

## Probe 1 — every `fields` member our skills request

### The exact lists our skills pass

| file | call site | `fields` |
|---|---|---|
| `skills/engagement-checker/SKILL.md` | §7a dedupe scan | `["id","title","description","url","createdAt"]` |
| `skills/daily-work-brief/SKILL.md` | read 1, GCN board | `["id","title","priority","status","statusType","teamId"]` |
| `skills/daily-work-brief/SKILL.md` | read 2, cross-team assigned | same six as read 1 |
| `skills/daily-work-brief/SKILL.md` | read 3, completed in window | `["id","title","completedAt"]` |
| `skills/linear-ticketing/SKILL.md` | — | issues no call; §8 documents the enum only |

**Union our skills request = 10 distinct members:** `id`, `title`,
`description`, `url`, `createdAt`, `priority`, `status`, `statusType`,
`teamId`, `completedAt`.

### (a) The full union, all at once — SUCCEEDS

Call: `team = 81e7b769-2a46-4e2a-8db5-c165a7963b0e`, `limit = 11`,
`fields = ["id","title","description","url","createdAt","priority","status","statusType","teamId","completedAt"]`.

Raw tool result (session `20260810_183344_9e1bc2`):

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":\"WF6 probe description body. Marker: wf6-probe-marker-7Q4X\\nSource: slack:C0BP2GZUFSR:1786359613.979759\",\"url\":\"https://linear.app/mobile-club/issue/GCN-11\",\"createdAt\":\"2026-08-10T17:34:41.473Z\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"status\":\"Canceled\",\"statusType\":\"canceled\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-10\",\"title\":\"Prune the Linear MCP tool surface exposed to Tars\",\"description\":\"All 58 mcp__linear__* tools are currently enabled to the live agent, including destructive ones (delete_*, merge_diff). Before this wave Tars had no Linear write capability at all — the engagement-checker collector was read-only by construction — so this is a new capability surface, not a lapsed control. Deliberately not hardened in this push per the broad-first-restrict-later rule. The linear-ticketing skill's closed tool list (save_issue, save… (truncated, use `get_issue` for full description)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-10\",\"createdAt\":\"2026-08-10T17:16:01.938Z\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"status\":\"Todo\",\"statusType\":\"unstarted\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-9\",\"title\":\"Wire Linear MCP server natively into Hermes (catalog preset linear)\",\"description\":\"This is the precursor that unblocks GCN-3, GCN-4 and GCN-5 (engagement-checker push+pull, daily-work-brief board, default-team rule).\\n\\nAll three of those tickets assume Tars can read and write Linear from inside the agent loop. Today it cannot: the only path is ad-hoc raw HTTP against [api.linear.app/graphql](<http://api.linear.app/graphql>), with the API key sourced from ~/.hermes/.env at call time and hand-written GraphQL in a shell heredoc fo… (truncated, use `get_issue` for full description)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-9\",\"createdAt\":\"2026-08-10T16:56:23.296Z\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"status\":\"Done\",\"statusType\":\"completed\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":\"2026-08-10T17:15:10.727Z\"},{\"id\":\"GCN-5\",\"title\":\"Tars default-team rule: create tickets in GCN unless told otherwise\",\"description\":\"Rule lives in the ticket-creating skill logic (NOT SOUL Phase 2 placeholder); cross-team read stays.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-5\",\"createdAt\":\"2026-08-10T16:24:51.510Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"status\":\"In Progress\",\"statusType\":\"started\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-4\",\"title\":\"daily-work-brief: pull Linear board + render priority-sorted\",\"description\":\"Open the 08:30 brief with the GCN + assigned-across-teams board sorted by priority; cooper Orca/Claude history section already exists - keep it.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-4\",\"createdAt\":\"2026-08-10T16:24:51.186Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"status\":\"In Progress\",\"statusType\":\"started\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-3\",\"title\":\"engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard\",\"description\":\"Tracked items become GCN issues; JSON demotes to cache/id-map; must not re-detect its own filings; closing an issue in Linear clears the item.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-3\",\"createdAt\":\"2026-08-10T16:24:50.861Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"status\":\"In Progress\",\"statusType\":\"started\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-7\",\"title\":\"WF6 E2E verification + docs/facts + status log updates\",\"description\":\"Pass criteria in the spec Verification section; evidence required, not claims.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-7\",\"createdAt\":\"2026-08-10T16:24:52.100Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"status\":\"In Progress\",\"statusType\":\"started\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-6\",\"title\":\"cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default\",\"description\":\"to-tickets skill (matt-pocock plugin) + a preference line in gaetan-metarepo.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-6\",\"createdAt\":\"2026-08-10T16:24:51.819Z\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"status\":\"Done\",\"statusType\":\"completed\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":\"2026-08-10T16:39:47.822Z\"},{\"id\":\"GCN-8\",\"title\":\"WF6 T6 verification probe — disposable, ignore\",\"description\":\"Disposable probe created by the WF6 Session B (T6) session to verify the [claude.ai](<http://claude.ai>) Linear MCP connector can write to team GCN from a Claude Code session on cooper.\\n\\nContext: `status/probes/wf6/linear-live2.md` recorded \\\"session expired\\\" on every live call through this connector, so its write capability was unproven. This ticket closes that gap.\\n\\nWill be moved to Canceled immediately after creation. Safe to delete.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-8\",\"createdAt\":\"2026-08-10T16:34:28.223Z\",\"priority\":{\"value\":4,\"name\":\"Low\"},\"status\":\"Canceled\",\"statusType\":\"canceled\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":null},{\"id\":\"GCN-2\",\"title\":\"LINEAR_API_KEY scope check on the VM\",\"description\":\"Resolved 2026-08-10 - key confirmed full-write via reversible favoriteCreate probe; collector remains read-only by code allowlist.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-2\",\"createdAt\":\"2026-08-10T16:24:50.537Z\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"status\":\"Done\",\"statusType\":\"completed\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":\"2026-08-10T16:24:50.587Z\"},{\"id\":\"GCN-1\",\"title\":\"Bootstrap GCN team + workload labels + priority view\",\"description\":\"Created by the WF6 bootstrap agent from the Tars VM. The GCN team itself was created by Gaetan in the Linear UI on 2026-08-10 16:08 UTC; this bootstrap added the six workload labels and the seven WF6 tickets. No PER team was created - GCN is the personal team WF6 integrates with.\\n\\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)\",\"url\":\"https://linear.app/mobile-club/issue/GCN-1\",\"createdAt\":\"2026-08-10T16:24:50.167Z\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"status\":\"Done\",\"statusType\":\"completed\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"completedAt\":\"2026-08-10T16:24:50.215Z\"}],\"hasNextPage\":false}"}
```

**Verdict: the full union succeeds.** 11 of 11 GCN issues, `hasNextPage: false`,
`result` present, `error` absent. **(b) is moot — no bisect was needed, no
member was rejected.**

### The three §(d) members no skill currently requests

Second call, exercising `labels`, `team`, `updatedAt` — plus `assignee`, added
deliberately because GCN-8 is the board's only unassigned issue and is the
control for the "empty fields are omitted" rule under an *explicit* projection.

Call: `team = 81e7b769-2a46-4e2a-8db5-c165a7963b0e`, `limit = 11`,
`fields = ["id","labels","status","statusType","team","teamId","url","updatedAt","priority","assignee","completedAt"]`.

Raw tool result (session `20260810_183458_f59212`):

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"labels\":[\"test-check\"],\"status\":\"Canceled\",\"statusType\":\"canceled\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-11\",\"updatedAt\":\"2026-08-10T18:00:16.535Z\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":null},{\"id\":\"GCN-10\",\"labels\":[\"access\"],\"status\":\"Todo\",\"statusType\":\"unstarted\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-10\",\"updatedAt\":\"2026-08-10T17:16:01.938Z\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":null},{\"id\":\"GCN-9\",\"labels\":[\"access\"],\"status\":\"Done\",\"statusType\":\"completed\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-9\",\"updatedAt\":\"2026-08-10T17:15:10.783Z\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":\"2026-08-10T17:15:10.727Z\"},{\"id\":\"GCN-5\",\"labels\":[\"investigation\"],\"status\":\"In Progress\",\"statusType\":\"started\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-5\",\"updatedAt\":\"2026-08-10T16:56:36.722Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":null},{\"id\":\"GCN-4\",\"labels\":[\"fix\"],\"status\":\"In Progress\",\"statusType\":\"started\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-4\",\"updatedAt\":\"2026-08-10T16:56:36.722Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":null},{\"id\":\"GCN-3\",\"labels\":[\"fix\"],\"status\":\"In Progress\",\"statusType\":\"started\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-3\",\"updatedAt\":\"2026-08-10T16:56:36.722Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":null},{\"id\":\"GCN-7\",\"labels\":[\"test-check\"],\"status\":\"In Progress\",\"statusType\":\"started\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-7\",\"updatedAt\":\"2026-08-10T16:40:34.253Z\",\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":null},{\"id\":\"GCN-6\",\"labels\":[\"discovery\"],\"status\":\"Done\",\"statusType\":\"completed\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-6\",\"updatedAt\":\"2026-08-10T16:39:47.842Z\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":\"2026-08-10T16:39:47.822Z\"},{\"id\":\"GCN-8\",\"labels\":[],\"status\":\"Canceled\",\"statusType\":\"canceled\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-8\",\"updatedAt\":\"2026-08-10T16:34:34.585Z\",\"priority\":{\"value\":4,\"name\":\"Low\"},\"completedAt\":null},{\"id\":\"GCN-2\",\"labels\":[\"access\"],\"status\":\"Done\",\"statusType\":\"completed\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-2\",\"updatedAt\":\"2026-08-10T16:24:50.537Z\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":\"2026-08-10T16:24:50.587Z\"},{\"id\":\"GCN-1\",\"labels\":[\"access\"],\"status\":\"Done\",\"statusType\":\"completed\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"url\":\"https://linear.app/mobile-club/issue/GCN-1\",\"updatedAt\":\"2026-08-10T16:24:50.167Z\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"assignee\":\"Gaëtan Cathelain\",\"completedAt\":\"2026-08-10T16:24:50.215Z\"}],\"hasNextPage\":false}"}
```

### (c) + (d) — the member table

Accepted = the call did not fail. Present = the key appears on returned issues.
"Sample" is the observed value shape, not a schema claim.

| member | requested by | accepted? | present in output? | sample / note |
|---|---|---|---|---|
| `id` | EC §7a, DWB 1·2·3 | ✅ | ✅ 11/11 | `"GCN-11"` — the **identifier**, never a UUID |
| `title` | EC §7a, DWB 1·2·3 | ✅ | ✅ 11/11 | byte-identical, em dash and brackets intact |
| `description` | EC §7a | ✅ | ✅ 11/11 | **truncated ~400 ch** with the literal `… (truncated, use \`get_issue\` for full description)` on GCN-9/GCN-10; full on the short ones |
| `url` | EC §7a | ✅ | ✅ 11/11 | `https://linear.app/mobile-club/issue/GCN-11` |
| `createdAt` | EC §7a | ✅ | ✅ 11/11 | ISO-8601 Z string |
| `priority` | DWB 1·2 | ✅ | ✅ 11/11 | **object** `{"value":3,"name":"Medium"}` |
| `status` | DWB 1·2 | ✅ | ✅ 11/11 | flat display name, e.g. `"In Progress"` |
| `statusType` | DWB 1·2 | ✅ | ✅ 11/11 | `unstarted` · `started` · `completed` · `canceled` observed |
| `teamId` | DWB 1·2 | ✅ | ✅ 11/11 | GCN UUID on every row |
| `completedAt` | DWB 3 | ✅ | ✅ 11/11 | **`null`, NOT omitted**, on the 7 non-completed issues; ISO string on the 4 Done |
| `labels` | *(none)* | ✅ | ✅ 11/11 | array of **names**; **`[]`, not omitted**, on GCN-8 |
| `team` | *(none)* | ✅ | ✅ 11/11 | display name `"Gaetan"` |
| `updatedAt` | *(none)* | ✅ | ✅ 11/11 | ISO-8601 Z string |
| `assignee` | *(none — added as the omission control)* | ✅ | ⚠️ **10/11 — OMITTED on GCN-8** | `"Gaëtan Cathelain"`; the key is entirely absent on the unassigned issue |

**No member our skills request is rejected. No member our skills request is
accepted-but-never-returned.** Every one of the ten arrives on every row.

### The one behaviour that refines the prior measurement

`mcp-linear-measured-shapes.md` consequence 11 states the rule as a blanket
"**Empty fields are OMITTED, not null**", generalising from a single observation
(GCN-8's missing `assignee` in the *default* projection). Under an explicit
`fields` projection the rule is measured to be **field-specific**:

| empty value | representation | evidence |
|---|---|---|
| `assignee`, unassigned | **key absent** | GCN-8 |
| `completedAt`, not completed | **present, `null`** | 7 rows in both calls |
| `labels`, no labels | **present, `[]`** | GCN-8 |

Consequence: `row.get("completedAt")` and `"completedAt" in row` disagree —
every row has the key, so a presence test is not a completed test. Defensive
`.get()` is correct for all three; a `key in row` test is not.

---

## Probe 2 — `TeamIssues`, executed

### Query text — lifted verbatim from `skills/engagement-checker/SKILL.md` §4 `OPS`

```graphql
query TeamIssues($teamId: ID!, $since: DateTimeOrDuration!, $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after, includeArchived: true,
        filter: {team: {id: {eq: $teamId}}, updatedAt: {gte: $since}}) {
        nodes { id identifier title url description priority dueDate createdAt updatedAt
          canceledAt completedAt state { id name type } assignee { id name }
          team { id key name } }
        pageInfo { hasNextPage endCursor }
      }
    }
```

Sent as `{"operationName":"TeamIssues","query":<the above>,"variables":{…}}` —
the same envelope `_post()` builds. Variables, run 1:

```json
{"teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","since":"2026-08-05T00:00:00Z","first":5,"after":null}
```

Request body 684 bytes. Response **HTTP 200, 5687 bytes.**

### Raw response (run 1)

```json
{"data":{"viewer":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c"},"issues":{"nodes":[{"id":"78fc29f0-8ec6-4aa2-803a-b03c216f2caf","identifier":"GCN-11","title":"WF6 probe — disposable [EC-ZZ99]","url":"https://linear.app/mobile-club/issue/GCN-11","description":"WF6 probe description body. Marker: wf6-probe-marker-7Q4X\nSource: slack:C0BP2GZUFSR:1786359613.979759","priority":3,"dueDate":null,"createdAt":"2026-08-10T17:34:41.473Z","updatedAt":"2026-08-10T18:00:16.535Z","canceledAt":"2026-08-10T17:50:29.663Z","completedAt":null,"state":{"id":"77aad3b3-deac-49a7-a39e-1bea02d93820","name":"Canceled","type":"canceled"},"assignee":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain"},"team":{"id":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","key":"GCN","name":"Gaetan"}},{"id":"946a75fb-1882-4397-a603-9d0f41cc5bd6","identifier":"GCN-10","title":"Prune the Linear MCP tool surface exposed to Tars","url":"https://linear.app/mobile-club/issue/GCN-10","description":"All 58 mcp__linear__* tools are currently enabled to the live agent, including destructive ones (delete_*, merge_diff). Before this wave Tars had no Linear write capability at all — the engagement-checker collector was read-only by construction — so this is a new capability surface, not a lapsed control. Deliberately not hardened in this push per the broad-first-restrict-later rule. The linear-ticketing skill's closed tool list (save_issue, save_comment, list_issues, get_issue) is currently the only blast-radius bound and it is policy, not enforcement. Investigate hermes mcp configure linear as the prune mechanism.\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":3,"dueDate":null,"createdAt":"2026-08-10T17:16:01.938Z","updatedAt":"2026-08-10T17:16:01.938Z","canceledAt":null,"completedAt":null,"state":{"id":"59ed732a-f242-4eba-926f-1c0d128fe83c","name":"Todo","type":"unstarted"},"assignee":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain"},"team":{"id":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","key":"GCN","name":"Gaetan"}},{"id":"2f19a5de-e691-4519-84f7-b084f772283c","identifier":"GCN-9","title":"Wire Linear MCP server natively into Hermes (catalog preset linear)","url":"https://linear.app/mobile-club/issue/GCN-9","description":"This is the precursor that unblocks GCN-3, GCN-4 and GCN-5 (engagement-checker push+pull, daily-work-brief board, default-team rule).\n\nAll three of those tickets assume Tars can read and write Linear from inside the agent loop. Today it cannot: the only path is ad-hoc raw HTTP against [api.linear.app/graphql](<http://api.linear.app/graphql>), with the API key sourced from ~/.hermes/.env at call time and hand-written GraphQL in a shell heredoc for every operation. There is no introspectable tool surface, no parameter schema the model can reason about, and the secret-handling rules make each ad-hoc call fragile.\n\nThis ticket wires the Linear MCP server natively into Hermes so the agent gets a first-class `mcp__linear__*` tool surface instead of raw HTTP.\n\nImportant: no Linear MCP server has ever existed on this VM. `mcp_servers:` in ~/.hermes/config.yaml currently holds exactly two entries, `slack` and `notion`. This is fresh wiring, not a re-add, so there is no prior stanza to restore or migrate.\n\nBlocks: GCN-3, GCN-4, GCN-5.\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":1,"dueDate":null,"createdAt":"2026-08-10T16:56:23.296Z","updatedAt":"2026-08-10T17:15:10.783Z","canceledAt":null,"completedAt":"2026-08-10T17:15:10.727Z","state":{"id":"0434e579-7b85-487a-8cf9-5aed6caaf41b","name":"Done","type":"completed"},"assignee":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain"},"team":{"id":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","key":"GCN","name":"Gaetan"}},{"id":"3925f93c-813e-4b10-904d-5dd07ece1f6c","identifier":"GCN-8","title":"WF6 T6 verification probe — disposable, ignore","url":"https://linear.app/mobile-club/issue/GCN-8","description":"Disposable probe created by the WF6 Session B (T6) session to verify the [claude.ai](<http://claude.ai>) Linear MCP connector can write to team GCN from a Claude Code session on cooper.\n\nContext: `status/probes/wf6/linear-live2.md` recorded \"session expired\" on every live call through this connector, so its write capability was unproven. This ticket closes that gap.\n\nWill be moved to Canceled immediately after creation. Safe to delete.\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":4,"dueDate":null,"createdAt":"2026-08-10T16:34:28.223Z","updatedAt":"2026-08-10T16:34:34.585Z","canceledAt":"2026-08-10T16:34:34.568Z","completedAt":null,"state":{"id":"77aad3b3-deac-49a7-a39e-1bea02d93820","name":"Canceled","type":"canceled"},"assignee":null,"team":{"id":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","key":"GCN","name":"Gaetan"}},{"id":"75579da7-27e9-401e-a34f-1b4351667cc3","identifier":"GCN-7","title":"WF6 E2E verification + docs/facts + status log updates","url":"https://linear.app/mobile-club/issue/GCN-7","description":"Pass criteria in the spec Verification section; evidence required, not claims.\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":2,"dueDate":null,"createdAt":"2026-08-10T16:24:52.100Z","updatedAt":"2026-08-10T16:40:34.253Z","canceledAt":null,"completedAt":null,"state":{"id":"a428dade-3c2d-42b0-86ed-50460344ca41","name":"In Progress","type":"started"},"assignee":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain"},"team":{"id":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","key":"GCN","name":"Gaetan"}}],"pageInfo":{"hasNextPage":true,"endCursor":"75579da7-27e9-401e-a34f-1b4351667cc3"}}}}
```

### Does it parse? Does it return data? Does the filter work?

- **Parses.** No syntax or validation error. Top-level keys are exactly
  `["data"]`; **`errors` is absent** — `_post()`'s `payload.get("errors")` guard
  passes.
- **Returns data.** 5 nodes for `first: 5`.
- **The team filter selects GCN and only GCN.** Every node carries
  `team.key == "GCN"` / `team.id == 81e7b769-…`. `$teamId: ID!` feeds
  `TeamFilter.id: IDComparator.eq` without a coercion error.
- **`viewer { id }` resolves** to Gaetan's UUID, so `scan_issues`' "authenticated
  viewer shape is unavailable" guard passes.
- **The `since` lower bound is real, not silently ignored.** Three follow-up
  runs, `first: 50`, same query:

  | `since` | nodes | identifiers | `pageInfo` |
  |---|---|---|---|
  | `2026-08-10T17:20:00Z` | **1** | `GCN-11` (updatedAt 18:00:16) | `{"hasNextPage": false, "endCursor": "78fc29f0-…"}` |
  | `2027-01-01T00:00:00Z` | **0** | — | `{"hasNextPage": false, "endCursor": null}` |
  | `-P1D` (duration form) | **11** | GCN-11…GCN-1 | `{"hasNextPage": false, "endCursor": "c57fc81e-…"}` |

  GCN-10 (17:16:01) and GCN-9 (17:15:10) are correctly excluded at the 17:20
  bound. `DateTimeOrDuration` accepts both an ISO-Z instant (what the collector
  sends) and an ISO-8601 duration.
- **Pagination works and does not stall.** Page 2 with
  `after: "75579da7-…"` (page 1's `endCursor`) returned a **disjoint** set —
  `GCN-6, GCN-5, GCN-4, GCN-3, GCN-2` — with a **new** `endCursor`
  `a5a5380e-…`. Page 1 ∪ page 2 ∪ the remaining `GCN-1` = the whole team, no
  overlap, no repeat. `scan_issues`' `next_after == after` stall guard is never
  tripped.
- An empty result yields `endCursor: null` **with** `hasNextPage: false`, so
  `connection()`'s `hasNextPage and not endCursor` guard does not misfire on
  the zero-row case.

### Node shape vs `AssignedIssues` — measured, not compared textually

`AssignedIssues` was executed with the same `since` and `first: 2` (HTTP 200,
2154 bytes, `errors` absent, `viewer.id` = Gaetan). Both node sets, read off the
live responses:

| | `AssignedIssues` | `TeamIssues` (page 2) |
|---|---|---|
| node keys | `id, identifier, title, url, description, priority, dueDate, createdAt, updatedAt, canceledAt, completedAt, state, assignee, team` | **identical, same order** |
| `state` sub-keys | `id, name, type` | `id, name, type` |
| `team` sub-keys | `id, key, name` | `id, key, name` |
| `priority` type | `int` | `int` |
| `pageInfo` | `{hasNextPage, endCursor}` | `{hasNextPage, endCursor}` |

**The shapes match exactly.** Everything the collector's downstream code touches
is present and correctly typed:

- `issue["id"]` — a real **UUID** (`78fc29f0-…`), used by the `seen_ids` dedupe
  of the two views. ✅
- `issue["identifier"]` — `"GCN-11"`, which becomes `"linear:" + identifier`,
  the same namespace as the native `id`. ✅ (§4's malformed-node guard requires
  both and both are always present.)
- `issue["updatedAt"]` — ISO-Z, parseable by `instant()`, used for the exact
  cursor window and the deterministic newest-first cap sort. ✅
- `state.type` — §5 reconciles on `completed` / `canceled`; both observed
  live. ✅
- `assignee` — **can be `null`** (GCN-8). Harmless: the node guard checks only
  `id`/`identifier`, and §5's assignee-change rule reads the *history* view's
  `fromAssignee`/`toAssignee`, not this field.
- `description` — returned **in full, never truncated** on this transport
  (GCN-10's 620-char body arrives whole, where native `list_issues` cuts it at
  ~400 with a marker). `sanitized(…, 2000)` is the only cap on this path.

### Probe 2 verdict

**`TeamIssues` executes correctly, as written. No GraphQL error, no correction
needed to the query string.** The review's concern — a bad filter degrading
silently behind a fail-closed collector — is disproved: the filter is exercised,
selective on both axes (team and `since`), and the node shape is byte-for-byte
the contract `AssignedIssues` already established.

---

## Corrections required in the skills

Nothing here invalidates a `fields` list or the `TeamIssues` query. Four
wording-level corrections, one of which is a real render bug.

1. **`skills/engagement-checker/SKILL.md` §8 — `priority` shape is wrong for the
   collector path.** The line reads: *"a filed item uses its issue's priority
   (`priority.value` — priority reads back as an object)"*. Measured: the §4 raw
   GraphQL collector — the **only** Linear read an ordinary run performs —
   returns `priority` as a **bare integer** (`"priority":3`). `priority.value`
   on an int is an attribute error / undefined, so every filed bullet sorts as
   if it had no priority. Change to name both transports: *native
   `list_issues`/`get_issue`/`save_issue` return `{"value":n,"name":…}`; the §4
   collector returns a bare int; read
   `p["value"] if isinstance(p, dict) else p`.*

2. **`skills/linear-ticketing/SKILL.md` §9 — the blanket omission rule is
   too broad.** *"Empty fields are OMITTED, not null"* is measured true for
   `assignee` only. Under an explicit `fields` projection `completedAt` comes
   back **`null`** and `labels` comes back **`[]`** — both present. Reword to:
   *some empty fields are omitted (`assignee`, measured), others come back
   `null` (`completedAt`) or empty (`labels`) — never assume either; use a
   defaulting read, and never use key-presence as a value test.*

3. **`skills/daily-work-brief/SKILL.md` §3 — same blanket claim, plus a
   concrete consequence for read 3.** It repeats *"empty fields are omitted,
   not null, so never assume a key exists"*. Apply the same reword, and state
   explicitly that read 3's client-side "keep only issues whose `completedAt`
   falls inside the window" must test the **value**, not the key: with
   `fields: ["id","title","completedAt"]` **every** row carries `completedAt`,
   `null` on the ones that are not completed. A `"completedAt" in row` test
   keeps all of them.

4. **`skills/linear-ticketing/SKILL.md` §9 — scope the description-truncation
   note to the native transport.** It currently says `description` is truncated
   at ~400 chars in `list_issues` and returned whole by `get_issue`/`save_issue`.
   Add: the `engagement-checker` §4 raw GraphQL collector also returns it
   **whole** (measured, GCN-10's 620-char body). This matters because
   `engagement-checker` §7a's head-placement rationale ("the head placement
   exists to survive truncation") is a property of the **native** dedupe scan
   only — the collector never sees the truncated form.

Explicitly **not** requiring a change:

- Every `fields` list in all three skills — all ten members measured valid and
  returned on every row. `engagement-checker` §7a's
  `["id","title","description","url","createdAt"]` and `daily-work-brief`'s
  three lists all execute clean.
- The `TeamIssues` query string in `engagement-checker` §4 — executes as
  written, filters correctly on both axes, paginates without stalling, and
  returns the same node shape as `AssignedIssues`.
- `engagement-checker` §4's malformed-node, viewer, stall and cursor guards —
  all four exercised against live responses without a false trip.
