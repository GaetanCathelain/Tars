# Linear MCP — MEASURED request/response shapes

Run date: 2026-08-10 (VM clock UTC), 17:34–17:55Z. Method: every payload below
was produced by a live `mcp__linear__*` tool call on the Tars VM, driven with
`~/.local/bin/hermes chat -Q -q '<instruction>'` over ssh, with the agent
instructed to echo the raw tool result byte for byte. Nothing here is inferred
from a schema, a vendor doc, or an LLM summary.

Disposable subject: **GCN-11** "WF6 probe — disposable [EC-ZZ99]", created for
this probe and left **Canceled**. No other issue was written to.

**Transport envelope.** Every tool result arrives as a JSON object with exactly
one key: `result` on success, `error` on failure. The value of `result` is a
**JSON string**, not an object — it must be parsed a second time. Hermes may
additionally wrap the payload in an `<untrusted_tool_result source="...">` block
before the agent sees it; that wrapper is Hermes-side, not Linear's.

---

## Q1 — `save_issue` CREATE response

Call: `save_issue` with `title`, `team` (GCN uuid), `state` (Todo uuid),
`labels: ["fb2f3fef-…"]`, `priority: 3`, `assignee` (Gaetan uuid). No `id`.

```
{"result": "{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":null,\"priority\":{\"value\":3,\"name\":\"Medium\"},\"url\":\"https://linear.app/mobile-club/issue/GCN-11\",\"gitBranchName\":\"gcn-11-wf6-probe-disposable-ec-zz99\",\"createdAt\":\"2026-08-10T17:34:41.473Z\",\"updatedAt\":\"2026-08-10T17:34:41.473Z\",\"archivedAt\":null,\"completedAt\":null,\"startedAt\":null,\"canceledAt\":null,\"dueDate\":null,\"slaStartedAt\":null,\"slaMediumRiskAt\":null,\"slaHighRiskAt\":null,\"slaBreachesAt\":null,\"status\":\"Todo\",\"statusType\":\"unstarted\",\"labels\":[\"test-check\"],\"attachments\":[],\"documents\":[],\"createdBy\":\"Gaëtan Cathelain\",\"createdById\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"assignee\":\"Gaëtan Cathelain\",\"assigneeId\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\"}"}
```

Answering the specific questions:

- **Top-level keys of the tool result:** exactly one — `result`.
- **`id`?** Yes, but its value is **`"GCN-11"` — the human identifier, NOT a
  UUID.**
- **`identifier`?** **NO. There is no `identifier` key anywhere in the payload.**
- **`url`?** Yes: `https://linear.app/mobile-club/issue/GCN-11`.
- **Success flag?** **None.** No `success`, no `ok`, no `status: 200`. The
  `status` key holds the workflow state name (`"Todo"`).
- `priority` comes back as an **object** `{"value":3,"name":"Medium"}`, though it
  is written as a bare integer.
- `labels` comes back as an array of **names** (`["test-check"]`), though it is
  written as an array of ids.
- Resolvers worked: `state`, `labels`, `assignee`, `team` all accepted UUIDs.

---

## Q2 — is the issue UUID ever obtainable?

**Through the three tools our skills use — `save_issue`, `get_issue`,
`list_issues` — NO field ever carries the issue's UUID. `id` IS the identifier
(`GCN-11`).** The UUIDs that do appear belong to other objects:
`createdById`, `assigneeId`, `teamId`, and (in `get_issue`)
`stateHistory[].state.id`.

`get_issue` on GCN-11 with `includeRelations: true`:

```
{"result": "{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":null,\"priority\":{\"value\":3,\"name\":\"Medium\"},\"url\":\"https://linear.app/mobile-club/issue/GCN-11\",\"gitBranchName\":\"gcn-11-wf6-probe-disposable-ec-zz99\",\"createdAt\":\"2026-08-10T17:34:41.473Z\",\"updatedAt\":\"2026-08-10T17:34:41.473Z\",\"archivedAt\":null,\"completedAt\":null,\"startedAt\":null,\"canceledAt\":null,\"dueDate\":null,\"slaStartedAt\":null,\"slaMediumRiskAt\":null,\"slaHighRiskAt\":null,\"slaBreachesAt\":null,\"status\":\"Todo\",\"statusType\":\"unstarted\",\"labels\":[\"test-check\"],\"attachments\":[],\"documents\":[],\"stateHistory\":[{\"state\":{\"id\":\"59ed732a-f242-4eba-926f-1c0d128fe83c\",\"name\":\"Todo\",\"type\":\"unstarted\"},\"startedAt\":\"2026-08-10T17:34:41.473Z\",\"endedAt\":null}],\"createdBy\":\"Gaëtan Cathelain\",\"createdById\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"assignee\":\"Gaëtan Cathelain\",\"assigneeId\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\",\"relations\":{\"blocks\":[],\"blockedBy\":[],\"relatedTo\":[],\"duplicateOf\":null}}"}
```

`get_issue` adds `stateHistory` (and `relations` when asked) over the create
shape. Still no issue UUID.

### The one oblique leak — the pagination cursor

`list_issues` returns `cursor` = **the UUID of the LAST issue on the page**.
Three independent confirmations:

| page ends at | returned `cursor` |
|---|---|
| GCN-9 (`limit:3`) | `2f19a5de-e691-4519-84f7-b084f772283c` — matches GCN-9's UUID recorded in `gcn9-native-mcp.md` |
| GCN-10 (`limit:2`) | `946a75fb-1882-4397-a603-9d0f41cc5bd6` |
| GCN-11 (`limit:1`) | `78fc29f0-8ec6-4aa2-803a-b03c216f2caf` |

Proof that this is really the issue UUID — `get_issue` with `id` set to that
cursor value resolves to GCN-11:

```
{"result": "{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",…,\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"test-check\"],…}"}
```

So **`get_issue` accepts a UUID as `id`**, but the only way to *learn* a UUID
natively is to paginate one issue at a time and harvest `cursor` — one call per
issue. That is not a viable keying strategy for a board of any size.

**Verdict: for design purposes, treat the issue UUID as UNAVAILABLE on the
native path.** The raw-GraphQL collector in `engagement-checker` §4 remains the
only place a UUID is obtainable at scale (it selects `id` and `identifier`
separately).

---

## Q3 — `list_issues` default projection

Call: `team` = GCN uuid, `limit` = 3, **no `fields`**.

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":\"\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"url\":\"https://linear.app/mobile-club/issue/GCN-11\",\"gitBranchName\":\"gcn-11-wf6-probe-disposable-ec-zz99\",\"createdAt\":\"2026-08-10T17:34:41.473Z\",\"updatedAt\":\"2026-08-10T17:34:41.473Z\",\"archivedAt\":null,\"completedAt\":null,\"startedAt\":null,\"canceledAt\":null,\"dueDate\":null,\"slaStartedAt\":null,\"slaMediumRiskAt\":null,\"slaHighRiskAt\":null,\"slaBreachesAt\":null,\"status\":\"Todo\",\"statusType\":\"unstarted\",\"labels\":[\"test-check\"],\"createdBy\":\"Gaëtan Cathelain\",\"createdById\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"assignee\":\"Gaëtan Cathelain\",\"assigneeId\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\"},{\"id\":\"GCN-10\",\"title\":\"Prune the Linear MCP tool surface exposed to Tars\",\"description\":\"All 58 mcp__linear__* tools are currently enabled to the live agent, including destructive ones (delete_*, merge_diff). Before this wave Tars had no Linear write capability at all — the engagement-checker collector was read-only by construction — so this is a new capability surface, not a lapsed control. Deliberately not hardened in this push per the broad-first-restrict-later rule. The linear-ticketing skill's closed tool list (save_issue, save… (truncated, use `get_issue` for full description)\",\"priority\":{\"value\":3,\"name\":\"Medium\"},\"url\":\"https://linear.app/mobile-club/issue/GCN-10\",\"gitBranchName\":\"gcn-10-prune-the-linear-mcp-tool-surface-exposed-to-tars\",\"createdAt\":\"2026-08-10T17:16:01.938Z\",\"updatedAt\":\"2026-08-10T17:16:01.938Z\",\"archivedAt\":null,\"completedAt\":null,\"startedAt\":null,\"canceledAt\":null,\"dueDate\":null,\"slaStartedAt\":null,\"slaMediumRiskAt\":null,\"slaHighRiskAt\":null,\"slaBreachesAt\":null,\"status\":\"Todo\",\"statusType\":\"unstarted\",\"labels\":[\"access\"],\"createdBy\":\"Gaëtan Cathelain\",\"createdById\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"assignee\":\"Gaëtan Cathelain\",\"assigneeId\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\"},{\"id\":\"GCN-9\",\"title\":\"Wire Linear MCP server natively into Hermes (catalog preset linear)\",\"description\":\"This is the precursor that unblocks GCN-3, GCN-4 and GCN-5 … (truncated, use `get_issue` for full description)\",\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"url\":\"https://linear.app/mobile-club/issue/GCN-9\",\"gitBranchName\":\"gcn-9-wire-linear-mcp-server-natively-into-hermes-catalog-preset\",\"createdAt\":\"2026-08-10T16:56:23.296Z\",\"updatedAt\":\"2026-08-10T17:15:10.783Z\",\"archivedAt\":null,\"completedAt\":\"2026-08-10T17:15:10.727Z\",\"startedAt\":\"2026-08-10T16:56:45.155Z\",\"canceledAt\":null,\"dueDate\":null,\"slaStartedAt\":null,\"slaMediumRiskAt\":null,\"slaHighRiskAt\":null,\"slaBreachesAt\":null,\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"createdBy\":\"Gaëtan Cathelain\",\"createdById\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"assignee\":\"Gaëtan Cathelain\",\"assigneeId\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\"}],\"hasNextPage\":true,\"cursor\":\"2f19a5de-e691-4519-84f7-b084f772283c\"}"}
```

**Envelope keys:** `issues`, `hasNextPage`, `cursor`. (`cursor` is **absent**
when `hasNextPage` is `false`.)

**Every field name present per issue (26), in order:**

`id`, `title`, `description`, `priority`, `url`, `gitBranchName`, `createdAt`,
`updatedAt`, `archivedAt`, `completedAt`, `startedAt`, `canceledAt`, `dueDate`,
`slaStartedAt`, `slaMediumRiskAt`, `slaHighRiskAt`, `slaBreachesAt`, `status`,
`statusType`, `labels`, `createdBy`, `createdById`, `assignee`, `assigneeId`,
`team`, `teamId`

**This list is not fixed.** Three measured deviations:

1. **Fields are OMITTED, not nulled, when empty.** GCN-8 (unassigned) came back
   with **no `assignee` key at all**:
   `{"id":"GCN-8","status":"Canceled","statusType":"canceled","labels":[],"priority":{"value":4,"name":"Low"},"updatedAt":"2026-08-10T16:34:34.585Z"}`
2. **Extra fields appear on some issues.** A Software Engineering issue carried
   `cycleId` (see Q10) that no GCN issue has.
3. **`description` is silently truncated in `list_issues`**, with the literal
   marker `… (truncated, use `get_issue` for full description)` appended.
   `get_issue`/`save_issue` return the full value.

There is **no `state` object and no `state.type`**. The state is carried by two
flat strings: `status` (display name, e.g. `"In Progress"`) and `statusType`
(machine type, one of `backlog` / `unstarted` / `started` / `completed` /
`canceled` / `duplicate`).

---

## Q4 — `list_issues` + `fields`

**Field selection WORKS.** The earlier "rejected live" note in
`gcn9-close-and-prune-ticket.md` was a misdiagnosis: the array *shape* is
correct, but `fields` is a **closed enum** and the earlier attempt used invalid
member names.

Proof it works — `fields: ["id","title","status"]`:

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"status\":\"Todo\"},{\"id\":\"GCN-10\",\"title\":\"Prune the Linear MCP tool surface exposed to Tars\",\"status\":\"Todo\"}],\"hasNextPage\":true,\"cursor\":\"946a75fb-1882-4397-a603-9d0f41cc5bd6\"}"}
```

`fields: ["id"]` → `{"issues":[{"id":"GCN-11"}],"hasNextPage":true,…}`.
`fields: []` → the **full default projection** (empty array = no projection).

Rejected shapes — a bare string, an object, and an invalid member:

```
{"error": "Input validation error: Invalid arguments for tool list_issues: fields.0: Invalid input"}
```

(identical text for `fields = "id,title,status"`, `fields = {"issue":[…]}`, and
`fields = ["identifier"]`.)

### The enum, measured

A 16-element probe array returned the offending indices, which pins the
membership exactly:

```
{"error": "Input validation error: Invalid arguments for tool list_issues: fields.0: Invalid input, fields.1: Invalid input, fields.2: Invalid input, fields.3: Invalid input, fields.6: Invalid input, fields.10: Invalid input, fields.11: Invalid input, fields.12: Invalid input, fields.13: Invalid input"}
```

against `["identifier","state","stateId","labelIds","project","projectId","cycle","cycleId","parentId","estimate","relations","comments","attachments","documents","slaBreachesAt","createdBy"]`.

- **INVALID (proven):** `identifier`, `state`, `stateId`, `labelIds`, `cycle`,
  `relations`, `comments`, `attachments`, `documents`.
- **VALID (proven):** the 26 default field names above, plus `project`,
  `projectId`, `cycleId`, `parentId`, `estimate`.

**Definitive:** `fields` is usable and worth using, but only with names drawn
from the response shape in Q3. `identifier` and `state` — the two names our
skills actually ask for — are both **hard errors that fail the entire call**.

---

## Q5 — truncation / pagination / total count

**There is no total count and no "truncated" flag. The envelope carries exactly
`hasNextPage` (bool) and, when true, `cursor` (a UUID).**

| call | result |
|---|---|
| `team=GCN`, **no `limit`** | 11 issues, `{"hasNextPage":false}` |
| `team=GCN`, `limit=100` | 11 issues, `{"hasNextPage":false}` |
| `assignee=Gaetan`, **no `limit`**, no team | **50 issues**, `{"hasNextPage":true,"cursor":"89ab46a8-99f3-4b9b-a818-9183936d8701"}` |
| `team=GCN`, `limit=2`, `cursor=2f19a5de-…` | 2 issues (GCN-5, GCN-4) — resumed correctly after GCN-9 |
| `assignee=Gaetan`, `limit=251` | `Input validation error: Invalid arguments for tool list_issues: limit: Invalid input` |
| `team=GCN`, `limit=500` | same `limit: Invalid input` error |

- **Default page size is 50**, empirically: the cross-team assignee call with no
  `limit` returned exactly 50 with `hasNextPage: true`. The GCN-only calls
  returned 11 because that is the whole team.
- **Maximum `limit` is 250.** 251 is a schema error.
- **The cap is NOT silent.** `hasNextPage` is the reliable signal, and it is
  correct: it was `false` on the 11-of-11 GCN reads and `true` on the 50-of-many
  cross-team read.
- **A second, Hermes-side ceiling exists.** `assignee=Gaetan, limit=250`
  produced a response Hermes could not deliver to the agent:
  `[Truncated: tool response was 399,510 chars. Full output could not be saved to sandbox.]`
  That is a local tool-result budget, not a Linear limit, and it defeats the
  `fields`-less large read entirely. Use `fields` to stay under it.

**Consequence for the idempotency title-scan: it CAN be trusted to be complete,
but only if the skill asserts `hasNextPage == false` on the response** (or
follows `cursor` until it is). Counting results against `limit` is not the
correct test — a page can be short and still have a next page in principle, and
`limit` was never reached in the GCN case. For GCN specifically, 11 issues sits
far inside one 250-issue page, so a single call with `fields` trimmed and an
explicit `hasNextPage == false` assertion is sufficient and cheap.

---

## Q6 — `state` param arity and semantics

All calls `team=GCN, limit=50`, varying `state` only.

| variant | outcome | returned |
|---|---|---|
| `["59ed732a-…","a428dade-…"]` (array of 2 ids) | **ERROR** | `Input validation error: Invalid arguments for tool list_issues: state: Invalid input` |
| `"Todo"` (state **name**) | SUCCESS | GCN-11:Todo, GCN-10:Todo |
| `"unstarted"` (state **type**) | SUCCESS | GCN-11:Todo, GCN-10:Todo |
| `"completed"` (state type) | SUCCESS | GCN-9, GCN-6, GCN-2, GCN-1 — all Done |
| `"!Done"` (negation) | **SUCCESS, 0 issues** | *(empty)* |
| `"Todo,In Progress"` (comma list) | **SUCCESS, 0 issues** | *(empty)* |
| single uuid (e.g. Canceled `77aad3b3-…`) | SUCCESS | used throughout for writes |

- **Arity: strictly one value.** An array is a hard schema error.
- **Accepts: state ID, state NAME, and state TYPE.** All three resolve.
- **Negation: NOT supported — and it fails SILENTLY.** `"!Done"` and a
  comma-separated list both return HTTP-200 with `"issues":[]`. A skill that
  guesses either syntax renders an **empty board with no error to catch**.

**Cross-team `state`-by-type also works** (`assignee` set, no `team`):
`state="completed"` returned GCN-9/6/2/1 plus `SE-683` (team "Software
Engineering"); `state="started"` returned GCN-5/4/3/7 plus `NMC-601` (team "NM x
MC"). So a non-terminal cross-team filter is expressible server-side as two
calls (`started`, `unstarted`, plus `backlog` if wanted) — the "other teams'
state ids are unknown" obstacle does not apply to types.

**Assignee + team simultaneously: WORKS.** `team=GCN, assignee=Gaetan, limit=50`
returned 10 of the 11 GCN issues — correctly excluding GCN-8, the only
unassigned one.

### `updatedAt` filter (bonus — our skills depend on it and called it unmeasured)

`updatedAt` takes a **bare string** and behaves as a **lower bound (since)**:

| value | returned |
|---|---|
| `"2020-01-01"` | 11 (all) |
| `"2026-08-10"` | 11 (all — all were updated today) |
| `"2026-08-10T17:45:00.000Z"` | **1** — only GCN-11, `updatedAt` 17:50:29 |
| `"2027-01-01"` | **0** |
| `"-P1D"` (duration) | accepted, returned the recent set |
| `{"gte":"2026-08-10T00:00:00.000Z"}` (object) | **ERROR** — `updatedAt: Invalid input` |

Only a lower bound is expressible. Any upper bound must be applied client-side.

---

## Q7 — `save_issue` UPDATE, `id` + `state` only

Call: `id = GCN-11`, `state = a428dade-3c2d-42b0-86ed-50460344ca41` (In
Progress). **No `labels` param.**

```
{"result": "{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":null,\"priority\":{\"value\":3,\"name\":\"Medium\"},\"url\":\"https://linear.app/mobile-club/issue/GCN-11\",\"gitBranchName\":\"gcn-11-wf6-probe-disposable-ec-zz99\",\"createdAt\":\"2026-08-10T17:34:41.473Z\",\"updatedAt\":\"2026-08-10T17:47:59.208Z\",\"archivedAt\":null,\"completedAt\":null,\"startedAt\":\"2026-08-10T17:47:59.116Z\",\"canceledAt\":null,\"dueDate\":null,\"slaStartedAt\":null,\"slaMediumRiskAt\":null,\"slaHighRiskAt\":null,\"slaBreachesAt\":null,\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"test-check\"],\"attachments\":[],\"documents\":[],\"createdBy\":\"Gaëtan Cathelain\",\"createdById\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"assignee\":\"Gaëtan Cathelain\",\"assigneeId\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"team\":\"Gaetan\",\"teamId\":\"81e7b769-2a46-4e2a-8db5-c165a7963b0e\"}"}
```

The UPDATE response is **shape-identical to the CREATE response** — same keys,
no `success`, no way to tell create from update by shape alone.

Independent re-read (`get_issue`, a fresh call, not the mutation response):

```
{"result": "{\"id\":\"GCN-11\",…,\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"test-check\"],\"attachments\":[],\"documents\":[],\"stateHistory\":[{\"state\":{\"id\":\"59ed732a-f242-4eba-926f-1c0d128fe83c\",\"name\":\"Todo\",\"type\":\"unstarted\"},\"startedAt\":\"2026-08-10T17:34:41.473Z\",\"endedAt\":\"2026-08-10T17:47:59.196Z\"},{\"state\":{\"id\":\"a428dade-3c2d-42b0-86ed-50460344ca41\",\"name\":\"In Progress\",\"type\":\"started\"},\"startedAt\":\"2026-08-10T17:47:59.196Z\",\"endedAt\":null}],…}"}
```

**`labels` is `["test-check"]` — omitting `labels` PRESERVES the label set.**
Confirmed again at the end of the run, after two further state moves, still
`["test-check"]`. The "omit `labels` unless the update is about labels" rule in
`linear-ticketing` §10 and `engagement-checker` §9 is **correct as written**.

`stateHistory` in `get_issue` is a usable audit trail and is the only place a
state UUID is returned on a read.

---

## Q8 — what a FAILED write looks like

Three distinct failure shapes were produced. All share one property: the
envelope key is **`error`, and `result` is absent**.

**(a) Schema validation, rejected by the MCP client before the network:**

```
{"error": "Input validation error: Invalid arguments for tool list_issues: state: Invalid input"}
```

**(b) Tool-level resolution failure (bogus issue id):**

```
{"error": "Error: Could not find issue \"GCN-99999\". To create a new issue, omit the \"id\" parameter. The \"id\" parameter should only be used to update an existing issue."}
```

**(c) Linear API error passed through, as a nested JSON string:**

```
{"error": "{\"error\":\"invalid_request\",\"message\":\"stateId contained an entry that could not be found.\",\"status\":400,\"requestId\":\"a290cbb879f2ad3a\"}"}
```

### The dangerous case — a write that FAILS while LOOKING like a success

`save_issue` with `id = GCN-11` and an unresolvable `state` returns a **normal,
complete, success-shaped issue payload with `result`, no `error`, and the state
unchanged**. Measured twice, with two different bad values:

`state = "00000000-0000-0000-0000-000000000000"`:

```
{"result": "{\"id\":\"GCN-11\",…,\"updatedAt\":\"2026-08-10T17:47:59.208Z\",…,\"status\":\"In Progress\",\"statusType\":\"started\",…}"}
```

`state = "NotARealStateName"`:

```
{"result": "{\"id\":\"GCN-11\",…,\"updatedAt\":\"2026-08-10T17:47:59.208Z\",…,\"status\":\"In Progress\",\"statusType\":\"started\",…}"}
```

Both returned `status: "In Progress"` — the value from the *previous*
successful update — and `updatedAt` did **not advance** (`17:47:59.208` in both,
identical to the earlier successful write). A fresh `get_issue` confirmed
`stateHistory` had gained no entry. **The writes silently no-opped.**

Note the inconsistency: a *well-formed but nonexistent* UUID that partially
resembles a real one produced the hard 400 in (c), while an all-zero UUID and a
nonsense name were silently dropped. The failure mode is not predictable from
the input, so it cannot be avoided by input validation.

### The success test skills must use — measured, not guessed

1. **Necessary:** the result has a `result` key and no `error` key.
2. **Not sufficient.** Additionally, **parse `result` and assert the returned
   payload actually carries the intended value** — for a state move, assert
   `statusType` (or `status`) equals the state that was requested; for a create,
   assert `id` matches `^[A-Z]+-\d+$` and `team`/`teamId` is GCN.
3. A create/update must never be recorded as done on the strength of "no error".

---

## Q9 — `save_comment` response shape

Call: `issueId = GCN-11`, `body = "WF6 probe comment — measuring save_comment response shape."`

```
{"result": "{\"id\":\"46cfbb79-b276-45b8-9363-09d1c1e9c53e\",\"body\":\"WF6 probe comment — measuring save_comment response shape.\",\"createdAt\":\"2026-08-10T17:42:23.883Z\",\"updatedAt\":\"2026-08-10T17:42:23.871Z\",\"parentId\":null,\"resolvedAt\":null,\"quotedText\":null,\"author\":{\"id\":\"4951b192-e49c-4b7e-b491-58c89e66043c\",\"name\":\"Gaëtan Cathelain\"}}"}
```

Keys: `id`, `body`, `createdAt`, `updatedAt`, `parentId`, `resolvedAt`,
`quotedText`, `author{id,name}`. No success flag.

**Asymmetry worth noting: a comment's `id` IS a real UUID** — the only object in
this surface whose `id` is a UUID rather than a human identifier. The response
carries **no back-reference to the issue** (`issueId` is not echoed), so a caller
must remember which issue it commented on.

---

## Q10 — cross-team read

Call: `assignee = 4951b192-e49c-4b7e-b491-58c89e66043c`, **no `team`**, no limit.

- 50 issues, `{"hasNextPage":true,"cursor":"89ab46a8-99f3-4b9b-a818-9183936d8701"}`.
- Distinct teams in that page: **Gaetan** (`81e7b769-2a46-4e2a-8db5-c165a7963b0e`),
  **Software Engineering** (`139092a5-7e07-4b8e-83be-9657d180215d`),
  **NM x MC** (`11a6f6b1-bf49-4c5f-8426-190474ce22e2`),
  **Mobile club** (`cb4eb28b-ab4b-4507-b235-a002323bb0d4`).
- A second read surfaced `NMC-601` on team "NM x MC".

**Team identity is carried by two fields on every issue: `team` (display name
string, e.g. `"Software Engineering"`) and `teamId` (team UUID).** There is no
`identifier` field, so the team prefix is only available by parsing `id`
(`SE-683` → `SE`). Grouping GCN first should key on `teamId ==
81e7b769-2a46-4e2a-8db5-c165a7963b0e`.

Verbatim non-GCN issue, showing the extra `cycleId` field that GCN issues do not
carry:

```
{"id":"SE-683","title":"Orca sessions: default orchestration policy + delegate-orca single-worker skill","description":"## What\n\n* A hand-typed `claude` in an Orca tab now starts with the orchestration policy and ultracode. Orca has no default-command setting for new tabs. A `~/.zshrc` wrapper keyed on `TERM_PROGRAM=Orca` adds the flags. `handoff-orca/PROCEDURE.md` § Bootstrap documents it.\n* New skill `delegate-orca`: delegate one lane of work to a fresh Orca worker session. It creates a subworktree of the current worktree, writes a handoff brief, and wires the … (truncated, use `get_issue` for full description)","priority":{"value":3,"name":"Medium"},"url":"https://linear.app/mobile-club/issue/SE-683/orca-sessions-default-orchestration-policy-delegate-orca-single-worker","gitBranchName":"se-683-orca-sessions-default-orchestration-policy-delegate-orca","createdAt":"2026-08-10T11:40:05.856Z","updatedAt":"2026-08-10T11:57:54.955Z","archivedAt":null,"completedAt":"2026-08-10T11:57:54.940Z","startedAt":"2026-08-10T11:40:05.887Z","canceledAt":null,"dueDate":null,"slaStartedAt":null,"slaMediumRiskAt":null,"slaHighRiskAt":null,"slaBreachesAt":null,"status":"Done","statusType":"completed","labels":[],"createdBy":"Gaëtan Cathelain","createdById":"4951b192-e49c-4b7e-b491-58c89e66043c","assignee":"Gaëtan Cathelain","assigneeId":"4951b192-e49c-4b7e-b491-58c89e66043c","team":"Software Engineering","teamId":"139092a5-7e07-4b8e-83be-9657d180215d","cycleId":"c5c48357-d704-43cf-8541-6acbb43cfcfd"}
```

Note the `url` for a non-GCN issue carries a slug segment (`/SE-683/orca-…`)
that GCN issue urls did not. No write of any kind was made to a company team.

---

## Q11 — is `description` in the `list_issues` default projection?

**YES, unambiguously.** `description` is one of the 26 default fields (Q3), and
it is returned by a plain `list_issues` call with no `fields` param.

**Full when short, truncated when long.** GCN-11's description was set to a
two-line body and came back **complete, including the embedded newline**, in a
plain `list_issues` result:

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":\"WF6 probe description body. Marker: wf6-probe-marker-7Q4X\\nSource: slack:C0BP2GZUFSR:1786359613.979759\",\"priority\":{\"value\":3,\"name\":\"Medium\"},…}],\"hasNextPage\":true,\"cursor\":\"78fc29f0-8ec6-4aa2-803a-b03c216f2caf\"}"}
```

Longer descriptions are cut at roughly 400 characters with the literal marker
`… (truncated, use \`get_issue\` for full description)` — see GCN-9 and GCN-10 in
Q3. So a handle placed in the description is retrievable from a list call
**only if it sits inside the first ~400 characters**. A handle on the last line
of a long description (the shape our skills use for the `Source:` /`Spec:`
provenance line) will be **cut off and invisible to a list scan**.

Note also `description` is `null` on an issue created without one, but `""` for
the same issue as seen through `list_issues` — the two tools disagree on the
empty representation.

---

## Q12 — can we FIND an issue by content?

**Yes: `list_issues` has a `query` parameter and it searches title AND
description. But it is a relevance-ranked fuzzy search, not a filter.**

There is **no dedicated issue-search tool** among the 58. `search_documentation`
searches Linear's product docs, not the workspace. `query` on `list_issues` is
the only content-search surface.

GCN-11's description was set to contain `wf6-probe-marker-7Q4X` and
`slack:C0BP2GZUFSR:1786359613.979759`.

| call | result |
|---|---|
| `team=GCN, query="wf6-probe-marker-7Q4X"` | **11 issues** — GCN-11 first, then GCN-8, GCN-7, GCN-2, GCN-1, GCN-6, GCN-5, GCN-4, GCN-3, GCN-10, GCN-9. i.e. **the entire team**, relevance-ordered |
| `query="wf6-probe-marker-7Q4X"`, no team | payload so large it blew the agent's output limit |
| `team=GCN, query="EC-ZZ99"` | **exactly 1** — GCN-11 |
| `team=GCN, query="slack:C0BP2GZUFSR:1786359613.979759"` | **exactly 1** — GCN-11 |
| `team=GCN, query="https://example.com/wf6-probe"` | **1** — GCN-11, *which contains no such string* (fuzzy match on "wf6 probe") |

Verbatim, the precise-token search that isolated the issue:

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"description\":\"WF6 probe description body. Marker: wf6-probe-marker-7Q4X\\nSource: slack:C0BP2GZUFSR:1786359613.979759\",…,\"status\":\"Canceled\",\"statusType\":\"canceled\",\"labels\":[\"test-check\"],…}],\"hasNextPage\":true,\"cursor\":\"78fc29f0-8ec6-4aa2-803a-b03c216f2caf\"}"}
```

Three properties that matter:

1. **`query` matches description text, not just titles** — proven: the Slack
   handle exists nowhere but in the description, and searching it returned the
   issue.
2. **A non-empty result is NOT a match.** The hyphenated marker query returned
   all 11 GCN issues, and the `example.com` URL query returned an issue that
   does not contain it. Tokens get split and fuzzy-matched. **Any caller must
   re-verify the exact handle by substring on the returned `title`/`description`
   before treating a hit as a hit.**
3. **`hasNextPage` is unreliable in query mode.** Every `query` call above
   returned `"hasNextPage":true` — including the one that returned all 11 of the
   team's 11 issues, and the ones that returned a single issue. The plain
   non-query read of the same team returns `"hasNextPage":false`. **The Q5
   completeness assertion does not hold for `query` searches.**

---

## Q13 — attachments

**Dead end. Linear's URL-attachment capability is NOT exposed by this MCP
server.** `create_attachment` has **no `url` parameter at all** — it is a
base64 file-upload tool. The server returned its full schema when the call was
rejected:

```
{"error": "tool_call to 'mcp__linear__create_attachment' is missing required argument(s): issue, base64Content, filename, contentType, sha256. The tool was NOT invoked. Parameters schema: {\"type\": \"object\", \"$schema\": \"https://json-schema.org/draft/2020-12/schema\", \"properties\": {\"issue\": {\"type\": \"string\", \"description\": \"Issue ID or identifier (e.g., LIN-123)\"}, \"base64Content\": {\"type\": \"string\", \"description\": \"Deprecated base64-encoded file content to upload\"}, \"filename\": {\"type\": \"string\", \"description\": \"Filename for the upload (e.g., 'screenshot.png')\"}, \"contentType\": {\"type\": \"string\", \"description\": \"MIME type for the upload (e.g., 'image/png', 'application/pdf')\"}, \"size\": {\"description\": \"Optional expected decoded file size in bytes. Rejects the upload if it does not match.\", \"type\": \"integer\", \"exclusiveMinimum\": 0, \"maximum\": 9007199254740991}, \"sha256\": {\"type\": \"string\", \"pattern\": \"^[a-fA-F0-9]{64}$\", \"description\": \"Expected SHA-256 hex digest of the decoded file bytes.\"}, \"title\": {\"description\": \"Optional title for the attachment\", \"type\": \"string\"}, \"subtitle\": {\"description\": \"Optional subtitle for the attachment\", \"type\": \"string\"}}, \"required\": [\"issue\", \"base64Content\", \"filename\", \"contentType\", \"sha256\"], \"additionalProperties\": false}. Retry tool_call with 'arguments' matching the parameters schema above."}
```

- **Tools that exist:** `create_attachment`, `create_attachment_from_upload`,
  `delete_attachment`, `get_attachment` (param: `id` only),
  `prepare_attachment_upload`. All are the **file-upload** flow. (The agent also
  listed `kanban_attach`, `kanban_attach_url`, `kanban_attachments` — those are
  Hermes' own kanban tools, nothing to do with Linear.)
- **No URL attach was possible**, so URL dedupe could not be tested. The
  question is moot: without a `url` param there is nothing to dedupe.
- **Attachments are readable but only via `get_issue`**, as `"attachments":[]`.
  They are **NOT in the `list_issues` default projection**, and `attachments` is
  a **rejected `fields` value** (Q4). So even a populated attachment list could
  not be pulled for a whole board in one call.
- **Reverse lookup by attachment url does not exist.** The `query` search that
  appeared to find the URL matched fuzzily on unrelated words (see Q12) — the
  issue has `attachments: []`.

Uploading a base64 blob per open loop, to carry a handle that then cannot be
listed or searched, is not a mechanism. **Rule attachments out.**

---

## Q14 — title suffix survivability

**Confirmed: bracketed text in a title round-trips byte-identically.** The title
`WF6 probe — disposable [EC-ZZ99]` (bracketed tag plus a non-ASCII em dash)
survived, unchanged and unescaped, through:

- the **create** (Q1);
- **three state-only updates** — Todo → In Progress → Canceled (Q7, Cleanup);
- a **description-only update** (Q11);
- and every read: `get_issue`, `list_issues` default, and `list_issues` with
  `fields`.

In all of them: `"title":"WF6 probe — disposable [EC-ZZ99]"`.

Additionally, a `query` on the bare tag `EC-ZZ99` returned **exactly one issue**
(Q12) — the tag is distinctive enough to isolate, unlike the hyphenated marker.

Bonus, measured on the same call: **omitting `state` on an update preserves the
state**, exactly as omitting `labels` preserves labels. The description-only
update left GCN-11 `Canceled` with an unchanged 3-entry `stateHistory`.

Also note `gitBranchName` is derived from the title at **create** time
(`gcn-11-wf6-probe-disposable-ec-zz99`) and did **not** change when the
description was updated — it is a slug of the original title, not a live
mirror.

---

## Correction to the ground truth used in this probe's brief

The brief gave In Progress as `a428dade-42ad-4c86-82ed-50460344ca41`. **That id
does not exist** — using it produced failure shape (c) above. The authoritative
GCN workflow states, from `mcp__linear__list_issue_statuses` with `team` = GCN:

```
{"result": "[{\"id\":\"a428dade-3c2d-42b0-86ed-50460344ca41\",\"type\":\"started\",\"name\":\"In Progress\"},{\"id\":\"95467ad9-9bad-4942-86d9-82d65e123a7b\",\"type\":\"backlog\",\"name\":\"Backlog\"},{\"id\":\"77aad3b3-deac-49a7-a39e-1bea02d93820\",\"type\":\"canceled\",\"name\":\"Canceled\"},{\"id\":\"59ed732a-f242-4eba-926f-1c0d128fe83c\",\"type\":\"unstarted\",\"name\":\"Todo\"},{\"id\":\"22cb42de-adf9-4c0d-a136-6af48300af8b\",\"type\":\"started\",\"name\":\"In Review\"},{\"id\":\"0ea8bdc6-215f-48d3-b228-779422c6b03e\",\"type\":\"duplicate\",\"name\":\"Duplicate\"},{\"id\":\"0434e579-7b85-487a-8cf9-5aed6caaf41b\",\"type\":\"completed\",\"name\":\"Done\"}]"}
```

`skills/linear-ticketing/SKILL.md` §2 already carries the correct value
(`a428dade-3c2d-42b0-86ed-50460344ca41`) and all seven states match this
response exactly. **The skill table is right; the brief was wrong.** Note also
the response is a bare JSON **array**, not an object with a `statuses` key.

---

## CLEANUP

`save_issue` with `id = GCN-11`, `state = 77aad3b3-deac-49a7-a39e-1bea02d93820`
(Canceled). Independent re-read via `get_issue`:

```
{"result": "{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",…,\"canceledAt\":\"2026-08-10T17:50:29.663Z\",…,\"status\":\"Canceled\",\"statusType\":\"canceled\",\"labels\":[\"test-check\"],\"attachments\":[],\"documents\":[],\"stateHistory\":[{\"state\":{\"id\":\"59ed732a-f242-4eba-926f-1c0d128fe83c\",\"name\":\"Todo\",\"type\":\"unstarted\"},\"startedAt\":\"2026-08-10T17:34:41.473Z\",\"endedAt\":\"2026-08-10T17:47:59.196Z\"},{\"state\":{\"id\":\"a428dade-3c2d-42b0-86ed-50460344ca41\",\"name\":\"In Progress\",\"type\":\"started\"},\"startedAt\":\"2026-08-10T17:47:59.196Z\",\"endedAt\":\"2026-08-10T17:50:29.680Z\"},{\"state\":{\"id\":\"77aad3b3-deac-49a7-a39e-1bea02d93820\",\"name\":\"Canceled\",\"type\":\"canceled\"},\"startedAt\":\"2026-08-10T17:50:29.680Z\",\"endedAt\":null}],…}"}
```

Not deleted. Labels intact.

Full-board verification, fresh `list_issues` (`team=GCN, limit=50`):

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"status\":\"Canceled\",\"statusType\":\"canceled\",\"labels\":[\"test-check\"],\"priority\":{\"value\":3,\"name\":\"Medium\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T17:50:29.741Z\"},{\"id\":\"GCN-10\",\"status\":\"Todo\",\"statusType\":\"unstarted\",\"labels\":[\"access\"],\"priority\":{\"value\":3,\"name\":\"Medium\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T17:16:01.938Z\"},{\"id\":\"GCN-9\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T17:15:10.783Z\"},{\"id\":\"GCN-5\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"investigation\"],\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:56:36.722Z\"},{\"id\":\"GCN-4\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"fix\"],\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:56:36.722Z\"},{\"id\":\"GCN-3\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"fix\"],\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:56:36.722Z\"},{\"id\":\"GCN-7\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"test-check\"],\"priority\":{\"value\":2,\"name\":\"High\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:40:34.253Z\"},{\"id\":\"GCN-6\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"discovery\"],\"priority\":{\"value\":3,\"name\":\"Medium\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:39:47.842Z\"},{\"id\":\"GCN-8\",\"status\":\"Canceled\",\"statusType\":\"canceled\",\"labels\":[],\"priority\":{\"value\":4,\"name\":\"Low\"},\"updatedAt\":\"2026-08-10T16:34:34.585Z\"},{\"id\":\"GCN-2\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:24:50.537Z\"},{\"id\":\"GCN-1\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"priority\":{\"value\":1,\"name\":\"Urgent\"},\"assignee\":\"Gaëtan Cathelain\",\"updatedAt\":\"2026-08-10T16:24:50.167Z\"}],\"hasNextPage\":false}"}
```

`hasNextPage: false` — this is the complete board. GCN-1 Done · GCN-2 Done ·
GCN-3 In Progress · GCN-4 In Progress · GCN-5 In Progress · GCN-6 Done · GCN-7
In Progress · GCN-8 Canceled · GCN-9 Done · GCN-10 Todo — **identical to the
"after" column of `gcn9-close-and-prune-ticket.md`**, and every `updatedAt` for
GCN-1..10 is `≤ 17:16:01.938Z`, i.e. earlier than this probe's first write at
`17:34:41`. Nothing but GCN-11 was touched.

Final end-of-run read, after the Q11–Q14 experiments (18:06Z):

```
{"result": "{\"issues\":[{\"id\":\"GCN-11\",\"title\":\"WF6 probe — disposable [EC-ZZ99]\",\"status\":\"Canceled\",\"statusType\":\"canceled\",\"labels\":[\"test-check\"],\"updatedAt\":\"2026-08-10T18:00:16.535Z\"},{\"id\":\"GCN-10\",\"title\":\"Prune the Linear MCP tool surface exposed to Tars\",\"status\":\"Todo\",\"statusType\":\"unstarted\",\"labels\":[\"access\"],\"updatedAt\":\"2026-08-10T17:16:01.938Z\"},{\"id\":\"GCN-9\",\"title\":\"Wire Linear MCP server natively into Hermes (catalog preset linear)\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"updatedAt\":\"2026-08-10T17:15:10.783Z\"},{\"id\":\"GCN-5\",\"title\":\"Tars default-team rule: create tickets in GCN unless told otherwise\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"investigation\"],\"updatedAt\":\"2026-08-10T16:56:36.722Z\"},{\"id\":\"GCN-4\",\"title\":\"daily-work-brief: pull Linear board + render priority-sorted\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"fix\"],\"updatedAt\":\"2026-08-10T16:56:36.722Z\"},{\"id\":\"GCN-3\",\"title\":\"engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"fix\"],\"updatedAt\":\"2026-08-10T16:56:36.722Z\"},{\"id\":\"GCN-7\",\"title\":\"WF6 E2E verification + docs/facts + status log updates\",\"status\":\"In Progress\",\"statusType\":\"started\",\"labels\":[\"test-check\"],\"updatedAt\":\"2026-08-10T16:40:34.253Z\"},{\"id\":\"GCN-6\",\"title\":\"cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"discovery\"],\"updatedAt\":\"2026-08-10T16:39:47.842Z\"},{\"id\":\"GCN-8\",\"title\":\"WF6 T6 verification probe — disposable, ignore\",\"status\":\"Canceled\",\"statusType\":\"canceled\",\"labels\":[],\"updatedAt\":\"2026-08-10T16:34:34.585Z\"},{\"id\":\"GCN-2\",\"title\":\"LINEAR_API_KEY scope check on the VM\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"updatedAt\":\"2026-08-10T16:24:50.537Z\"},{\"id\":\"GCN-1\",\"title\":\"Bootstrap GCN team + workload labels + priority view\",\"status\":\"Done\",\"statusType\":\"completed\",\"labels\":[\"access\"],\"updatedAt\":\"2026-08-10T16:24:50.167Z\"}],\"hasNextPage\":false}"}
```

Same ten states, same ten `updatedAt` values. GCN-11's `updatedAt` advanced only
to `18:00:16.535Z` (the Q11 description write); it remains **Canceled** and is
not deleted. The only lasting change this probe made to the workspace is the
existence of GCN-11 and one comment on it.

---

## Design consequences

Each item below **invalidates something currently written in our skills**.

1. **`engagement-checker`'s entire UUID keying is unimplementable on the native
   path.** §"Durable state", §2 stable IDs, §5 routing index and §7 all say
   `linear_issue.id` is "the issue's immutable UUID", that the routing index and
   the `filed` cache key on it, and that keying on the identifier is wrong
   because "Linear re-keys an issue when it moves team". **`save_issue` returns
   `id = "GCN-11"` — the identifier. There is no UUID to record.** A native
   create can only ever store the mutable identifier, so the stated invariant is
   false the moment it is written.

2. **The two halves of `engagement-checker`'s id-map are in different
   namespaces, so routing breaks 100% of the time.** §4's raw GraphQL collector
   emits candidate ids as `linear:<issue-uuid>` (it selects `id` and
   `identifier` separately and §2 says to read `items[].issue.id`), while §7's
   native create records `linear_issue.id = "GCN-11"`. A collector candidate for
   an issue Tars itself filed will **never** match the routing index. Every
   consequence §5 promises — "issue → item, closed", reopen detection, the
   `linear:completed` / `linear:canceled` transitions in §6 — silently fails.
   Either the collector must key on `identifier` too, or §7's filing must
   harvest the UUID by the one-issue-per-page `cursor` trick, or the whole map
   moves to the identifier and accepts the team-move risk.

3. **`linear-ticketing` §11 and `engagement-checker` §7's acceptance test can
   never pass.** Both require "a returned issue carrying **both an identifier and
   an id**" before recording a write as done. **There is no `identifier` key in
   any response.** Taken literally, every successful create is classified as a
   creation failure — so `linear_issue` is never recorded, and the next cycle
   depends entirely on the §7 title scan to avoid a duplicate. Rewrite the test
   as: `result` present, `error` absent, `id` matches `^[A-Z]+-\d+$`, and
   `teamId` is GCN.

4. **"No error" is not a success test — measured.** A `save_issue` update with
   an unresolvable `state` returns a full, normal, `result`-bearing payload and
   changes nothing (§Q8). `engagement-checker` §9's close-the-loop write sends
   `id` + `state` and treats a non-error as done; if the id or state ever fails
   to resolve it will record the item closed while the GCN issue stays open —
   producing exactly the "permanently open ghost issue" that §9 exists to
   prevent, and the daily brief will keep rendering it. **Every write must
   re-assert the intended field on the returned payload.**

5. **`daily-work-brief` read 1 fails as written.** It passes `state`
   "enumerating the four non-terminal GCN state ids". **`state` rejects an array
   outright** (`state: Invalid input`). The GCN board read errors on every run.
   Replace with a single unfiltered `team=GCN` call (11 issues, one page) plus a
   client-side `statusType ∉ {completed, canceled}` filter, or with separate
   per-type calls.

6. **`daily-work-brief` reads 2 and 3 fail as written, and so does
   `engagement-checker` §7's idempotency scan** — all three ask for `fields`
   values that are not in the enum. `identifier` and `state` are both **hard
   errors** (`fields.N: Invalid input`) that kill the whole call. The scan asks
   for `["id","identifier","title"]`; since §7 says "if the scan call fails,
   file nothing this cycle", **`engagement-checker` would never file a single
   issue.** Valid substitutes: `id` (already the identifier), `status` and
   `statusType` instead of `state`.

7. **There is no `state.type` to filter on client-side.** `daily-work-brief`
   read 2 says "drop client-side every issue whose `state.type` is `completed`
   or `canceled`", and `linear-ticketing` §8 says "request `state` in `fields`
   and filter client-side on `state.type`". The response has **no `state`
   object** — it has flat `status` (name) and `statusType` (type). Every such
   predicate references a field that does not exist and evaluates to
   undefined-vs-string, i.e. it drops nothing.

8. **The "not established" hedges are now established, and one of them was
   actively dangerous.** `linear-ticketing` §8 and `daily-work-brief` both say
   "whether `state` accepts a state *type* or a negation is not established — do
   not guess a `nin`". Measured: **type works** (`unstarted`, `started`,
   `completed`), **name works**, **single id works**; **negation does not exist
   and returns HTTP-200 with zero issues.** The hedge was right to forbid
   guessing — a guessed `"!Done"` renders `Board: clear.` with no error to
   catch — but the type support means read 2's "other teams' state ids are
   unknown, so this filter cannot be expressed server-side" is **wrong**:
   `state="started"` and `state="unstarted"` work cross-team, proven against
   SE-683 and NMC-601.

9. **`daily-work-brief`'s dedup key is impossible.** "Deduplicate reads 1 and 2
   by issue id (the UUID, not the identifier)" — there is no UUID. Dedup on
   `id` (the identifier) is the only option available, and within a single
   snapshot it is safe; the team-move hazard the parenthetical warns about
   cannot be defended against natively.

10. **Both skills' priority sort is broken on the measured payload.**
    `linear-ticketing` §2 defines the sort key as `0 if p==1 else 1 if p==2 …`
    over a scalar `p`, and §9 renders `P<n>`. **`priority` comes back as an
    object** `{"value":3,"name":"Medium"}`. Applied to the object, every
    comparison is false and every issue sorts last as No-priority, and `P<n>`
    renders a dict. Must read `priority.value`. (`priority` is still written as
    a bare integer — the asymmetry is real.)

11. **Renderers must tolerate missing keys.** GCN-8 came back with **no
    `assignee` key at all** rather than `assignee: null`, and `cursor` is absent
    when `hasNextPage` is false. Any `issue["assignee"]` style access will
    raise. Neither skill says this.

12. **`engagement-checker` §7's "`limit` large enough to cover 30 days" has a
    hard ceiling of 250** (251 → schema error), and a large unprojected read is
    additionally killed by a **Hermes-side tool-result cap** (a 250-issue
    unprojected response measured 399,510 chars and was dropped with "Full
    output could not be saved to sandbox"). The scan must use `fields` to stay
    under that budget. For GCN's current 11 issues this is not binding, but the
    skill states the rule generally.

13. **The completeness check the idempotency scan needs is `hasNextPage`, and
    no skill mentions it.** There is no total count and no truncation flag.
    `daily-work-brief`'s "a read that comes back at its `limit` may be
    truncated" is the wrong test — asserting `hasNextPage == false` is the right
    one, and it is reliable (measured `false` on a complete 11-of-11 read,
    `true` on a 50-of-many read). **With that assertion added, the title scan
    CAN be trusted to be complete.** Without it, the default page size of 50
    silently caps any unbounded read.

14. **`daily-work-brief`'s `updatedAt` hedge can be dropped, with a caveat.**
    `updatedAt` is measured as a **bare string, lower bound only** (`"…T17:45:00.000Z"`
    returned exactly the one issue updated after it; `"2027-01-01"` returned
    zero). The object form `{"gte": …}` is a hard error. So read 3's window can
    be expressed server-side at its **start** only; the end of the window must be
    filtered client-side. The skill's separate claim that there is no
    `completedAt` *parameter* is correct — but `completedAt` **is** a valid
    `fields` value, so read 3 can retrieve it and filter locally as intended.

15. **`list_issues` truncates `description`** with a literal
    `… (truncated, use \`get_issue\` for full description)` marker. Nothing in
    our skills reads descriptions from a list call today, but
    `engagement-checker`'s bounded-context path must not start doing so.

16. **The "fields array was REJECTED live" note in
    `gcn9-close-and-prune-ticket.md` is wrong and should not be propagated.**
    The array shape is correct; only the member names were invalid. `fields` is
    the right tool for keeping payloads under the Hermes result cap.

17. **`engagement-checker` §7's title scan is the right mechanism but the wrong
    implementation.** It substring-scans returned titles for `[EC-<short_id>]`,
    which is sound — titles round-trip byte-identically through create and every
    update (Q14). But it asks for `fields: ["id","identifier","title"]`, and
    `identifier` is a hard error, so the scan never runs (consequence 6). Change
    the projection to `["id","title"]` and the scan works exactly as designed.

18. **`hasNextPage` cannot be used as a completeness check on a `query`
    search.** Consequence 13's assertion is valid for plain filtered reads only:
    every `query` call measured returned `hasNextPage: true`, including one that
    returned all 11 of 11 team issues. If the idempotency scan moves to `query`,
    it loses the completeness guarantee that a plain `team`-filtered scan has.

19. **A provenance line at the END of a long description is invisible to a list
    scan.** `list_issues` truncates `description` at ~400 chars (Q11). Our
    convention puts `Source:` / `Spec:` on the **last** line — precisely the part
    that gets cut. Any handle intended to be found by a list scan must go in the
    title, or in the **first** ~400 characters of the description.

20. **Attachments are not an option and should be struck from consideration.**
    `create_attachment` takes no `url` — it requires `base64Content`,
    `filename`, `contentType`, `sha256` (Q13). Linear's URL-attachment dedupe is
    simply not exposed by this MCP server, and even a populated `attachments`
    array is unreachable from `list_issues` (`attachments` is a rejected
    `fields` value).

### Where the source handle should live — ranked, with the measurement

**1. Title tag — RECOMMENDED.** `[EC-<short_id>]` in the title.
*Backing:* Q14 — byte-identical round-trip through create, three state updates
and a description update; never truncated in any projection; `fields:
["id","title"]` makes the scan cheap; `hasNextPage: false` on a plain
`team=GCN` read gives a real completeness guarantee (Q5). This is what
`engagement-checker` §7 already specifies, and it is the only option every
measurement supports. It needs one fix — drop `identifier` from the `fields`
array.
*Cost:* the tag is visible to Gaetan on the board, and the title is the only
place it can live, so the short_id must stay short.

**2. Description handle, read from the same list scan — VIABLE as a
supplement.** Put `Source: slack:<channel>:<ts>` in the **first lines** of the
description.
*Backing:* Q11 — `description` is in the default projection and returns in full
when short; Q12 — the exact Slack handle in the description was findable. This
lets the scan match on the true source handle rather than a derived short_id,
which is strictly more robust to short_id collisions.
*Cost:* it costs nothing extra — the same `list_issues` call can project
`["id","title","description"]` and scan both — but it is **fragile to length**
(consequence 19) and cannot be the sole mechanism for issues whose description
grows past ~400 chars.

**3. `query` search — USE ONLY AS AN OPTIMISATION, NEVER AS THE TEST.**
*Backing:* Q12 — `query` does search description text, and a distinctive token
(`EC-ZZ99`, the full Slack handle) isolated exactly one issue. But it is
relevance-ranked: a hyphenated marker returned the entire team, and a URL that
appears nowhere in the issue still returned it. And `hasNextPage` is always
`true` (consequence 18).
*Rule:* if used, treat the result as a **candidate set** and confirm with an
exact substring check on the returned `title`/`description`. Never treat
"`query` returned rows" as "already filed", and never treat "`query` returned
nothing" as "not filed" — fall back to the full team scan for the negative case,
because a fuzzy search's empty result is not proof of absence.

**4. Attachments — NOT VIABLE.** See consequence 20. Rule out.

**5. Issue UUID — NOT VIABLE.** See consequences 1 and 2. The native path never
returns it. Rule out.

**Net design answer:** keep the title tag as the primary key, add the source
handle to the head of the description as a secondary match, and do the
idempotency check with a single plain `list_issues` on `team=GCN` with `fields:
["id","title","description"]`, asserting `hasNextPage == false` before trusting
a negative. That is one call, complete, and every element of it is measured
above.
