---
name: linear-ticketing
description: "Use whenever creating a ticket, updating, commenting on, or listing Linear issues — carries the write policy (tickets go to GCN unless Gaetan names another team), the required fields, the measured tool shapes, and the only sanctioned Linear tools."
version: 1.4.0
metadata:
  hermes:
    tags: [linear, tickets, gcn, priority, orchestration]
    category: orchestration
---

# Linear ticketing

The write contract for Linear. Every ticket Tars creates, updates, or comments on obeys this file, whether the request arrives as a plain DM or from another skill. `engagement-checker` and `daily-work-brief` reference it by name for ids, rubrics, tools, response shapes, and rendering.

Every tool fact below §8 was **measured live** on 2026-08-10 (`status/probes/wf6/mcp-linear-measured-shapes.md`), not read off a schema. The declared schema is wrong in several places. Where a behaviour is not recorded here, it is unknown — say so rather than assuming it.

## 1. Write policy — READ FIRST

- **Default: create in GCN.** Always an **explicit state** (Backlog is the silent default trap), **exactly one workload label**, and an **explicit priority**.
- **Writes to any company team (MC, NMC, SE, NXT, AGE, IE, DES, CLE): ONLY on Gaetan's explicit instruction naming the team, per message — his instruction is the approval, never standing.** Mirrors SOUL rule 4's semantics. "Explicitly" means Gaetan said which team; a team inferred from context is not enough.
- **Dedicated project channels supply routing context.** Before creating a ticket, inspect the conversation/channel context rather than treating the text alone as the request. Slack channel `C0BFQ5WFYTB` is dedicated to the **Support Engineer** project: tickets originating there go directly to team **NMC** and project **Support Engineer**, unless Gaetan says otherwise. This channel mapping counts as explicit routing authority for that create; do not create in GCN first and move it afterward.
- **When no dedicated-channel mapping or explicit team is available, create in GCN and say so in the reply.**
- A comment is a write. There is no standing authorization to comment on a company-team issue, not even one already assigned to Gaetan.
- **Age or silence never authorizes closing or canceling an issue.** Outside explicit user instructions, the only status-write authority is `engagement-checker`'s current **Permitted writes**: close-back for its own filed GCN issues (§5), and evidence-backed GCN organization including Done (§5a). The latter is not restricted to issues it created, but requires direct proof of the requested outcome and no conflicting evidence. Do not extend this exception to other workflows, teams, cancellation of human-filed issues, or inferred resolution. The engagement reminder pause remains in force and does not erase authorized knowledge filing.
- Reads across every team stay free and unrestricted.

## 2. Reference

Team **Gaetan / GCN** `81e7b769-2a46-4e2a-8db5-c165a7963b0e`. Viewer (Gaetan) `4951b192-e49c-4b7e-b491-58c89e66043c`.

GCN workflow states:

| name | type | id |
|---|---|---|
| Backlog | backlog | `95467ad9-9bad-4942-86d9-82d65e123a7b` |
| Todo | unstarted | `59ed732a-f242-4eba-926f-1c0d128fe83c` |
| In Progress | started | `a428dade-3c2d-42b0-86ed-50460344ca41` |
| In Review | started | `22cb42de-adf9-4c0d-a136-6af48300af8b` |
| Done | completed | `0434e579-7b85-487a-8cf9-5aed6caaf41b` |
| Canceled | canceled | `77aad3b3-deac-49a7-a39e-1bea02d93820` |
| Duplicate | duplicate | `0ea8bdc6-215f-48d3-b228-779422c6b03e` |

All seven verified live 2026-08-10 against the workspace's own status list. On **reads**, prefer the `statusType` values (`backlog`, `unstarted`, `started`, `completed`, `canceled`, `duplicate`) over these ids — they work cross-team, where other teams' ids are unknown. On **GCN writes** use the ids above: the one earlier objection to raw state UUIDs traced to a mistyped In Progress id in a briefing note, and that id was never in this table.

GCN team-scoped labels:

| name | id |
|---|---|
| fix | `bd21ce6f-2ff9-4850-9f81-d54fc48524f0` |
| investigation | `a0973525-a091-412d-abf1-aed30bedf60a` |
| discovery | `07362d67-916b-4b82-a248-2ff2d0dddbda` |
| access | `a4fd96c1-ee2d-4c03-8461-f8a15cca372a` |
| report | `bb8da686-e6f7-4691-a030-e11f98c3f8c6` |
| test-check | `fb2f3fef-0180-45a4-9de4-e4f69cbb428b` |

Priority integers: 0 = No priority, 1 = Urgent, 2 = High, 3 = Medium, 4 = Low.

"Priority-sorted" everywhere means the sort key `0 if p==1 else 1 if p==2 else 2 if p==3 else 3 if p==4 else 4` — Urgent, High, Medium, Low, then No-priority last.

**`p` is an integer, normalised at the boundary — never sort on the raw field.** Three producers, two shapes: the native tools return `priority` as an object `{"value":3,"name":"Medium"}`, `engagement-checker` §4's raw collector returns a **bare int**, and a stored priority is already an int. Read `p = p["value"] if isinstance(p, dict) else p`, treat absent or null as 0, then sort. Sorting on the object makes every comparison false and lands everything last.

## 3. Required on every create

One `mcp__linear__save_issue` call with **no `id`**, carrying all of these together:

- `team` `81e7b769-2a46-4e2a-8db5-c165a7963b0e` (GCN);
- `assignee` `4951b192-e49c-4b7e-b491-58c89e66043c` (Gaetan);
- `state` `59ed732a-f242-4eba-926f-1c0d128fe83c` (Todo);
- `labels` with **exactly one** workload label id from §4;
- an **explicit** `priority` integer from §5 — never omit it, never leave 0 unless the work is genuinely unprioritised;
- a `title`;
- a `description` whose **first line** is the provenance handle `Source: <what created it>`. It goes at the head, not the foot: `list_issues` truncates `description` at ~400 characters, so a handle on the last line is cut away and invisible to any scan.

**Never omit `state` on create. Linear's silent default is Backlog, not Todo** — the issue lands there permanently and the failure surfaces as a wrong column, never as an error. The same applies to `assignee`, to exactly one `labels` entry, and to an explicit `priority`: a create missing any of the four is malformed — refuse it and say so rather than sending a partial payload.

**All of it goes in ONE call. Never create then correct.** `save_issue` takes `team`, `state`, `labels`, `priority` and `assignee` as first-class params, so there is no window in which the issue exists in Backlog before a follow-up fixes it.

Then apply §10 before recording the create as done.

## 4. Label rubric

`fix` a broken thing to repair · `investigation` find out why · `discovery` explore or scope something new · `access` credentials, permissions, accounts · `report` produce a written output · `test-check` verify something works.

Pick exactly one. If two fit, pick the one naming the *outcome*.

## 5. Priority rubric

1 Urgent: blocks Gaetan or someone waiting on him today · 2 High: committed, dated, or blocking others this week · 3 Medium: real work, no date · 4 Low: nice-to-have · 0 only when genuinely unprioritised.

## 6. Allowed tools

**Default working set** — what an ordinary ticketing turn uses: `mcp__linear__save_issue` (create when `id` is omitted, update when it is passed), `mcp__linear__save_comment`, `mcp__linear__list_issues`, `mcp__linear__get_issue`, `mcp__linear__list_comments`. Reach past it when the task genuinely needs the wider surface, not by default.

`mcp__linear__list_comments` is the read-only path for issue comments. Measured live 2026-08-11: pass `issueId`, `limit` (max 250), and `orderBy` (`createdAt` or `updatedAt`). The response envelope is `comments` / `hasNextPage` / optional `cursor`; page exactly as for `list_issues`. Each comment includes `id`, `body`, `createdAt`, `updatedAt`, `parentId`, `resolvedAt`, `quotedText`, and `author` (`id`, `name`). Sort client-side when a requested display order is not guaranteed by the server response.

The reachable surface is enforced at registration by `mcp_servers.linear.tools.include` in `~/.hermes/config.yaml`. Since 2026-08-13, on Gaetan's explicit instruction ("all read/write access, but no admin tools like delete teams"), that list registers **56** of the server's **58** `mcp__linear__*` tools — every content-level read and write: issues, comments, labels, projects, milestones, documents, initiatives, status updates, attachments, cycles, releases and release notes, diff reads and diff comments, plus the item-scoped deletes `delete_comment`, `delete_attachment`, `delete_status_update`, `delete_diff_comment`. This deliberately **reverses the GCN-10 prune** of 2026-08-11, which registered 10. Evidence: `status/probes/linear-full-access-2026-08-13.md`.

**Excluded, and the only two: `merge_diff` and `submit_diff_review`.** SOUL rule 2 — Tars never merges, approves or pushes — so the tool surface must not offer them. Linear's MCP server ships **no** team, user, workspace, membership, webhook, API-key or integration mutation tool at all, so "no admin tools" costs nothing here: there was nothing else to exclude.

**Registered is still not permission to use.** §1's write policy governs every write unchanged — GCN by default, a company team only on Gaetan's per-message instruction naming it. The same for destruction: SOUL's destruction rules govern the `delete_*` tools, and being able to reach one is not a reason to call it. `LINEAR_API_KEY` remains an unscoped full-write personal key — the filter bounds what the model can *reach*, never what the credential can *do*.

**No workflow-status creation exists upstream.** `list_issue_statuses` and `get_issue_status` are reads; the server offers no create or update for a team's workflow states. Tars can *set* an issue's status (`save_issue.state`, §10) but cannot *create* one — a new state such as "Blocked" has to be added by a human in Linear's team settings.

`include` is a whitelist and fails closed: a tool Linear ships tomorrow does **not** auto-register. Adding it is one line under `mcp_servers.linear.tools.include`, taken with a `.bak` under `flock ~/.hermes/.wf3.lock`. Never write a sibling `tools.exclude` — `include` takes precedence and silently makes `exclude` a no-op.

## 7. Company teams

MC, NMC, SE, NXT, AGE, IE, DES, CLE: read freely. Every write — create, update, comment alike — needs §1's per-message instruction naming the team. `save_issue` accepts a bare `id` from any team, so the tool itself stops nothing: before any update or comment, confirm the target's team from the prefix of its `id` (`SE-683` → SE) or from `teamId` on a `get_issue`, and refuse a non-GCN target absent the instruction.

## 8. Call patterns

Linear is reached through the Hermes-native `mcp__linear__*` tools. A read or a write is a **tool call, not code**: no endpoint, no opener, no key handling, no pagination loop. The key is never printed, echoed, logged, hashed, persisted, or placed in **any process's argv** — the native path satisfies this by construction: it lives in `~/.hermes/.env`, reaches the server stanza only as the `${LINEAR_API_KEY}` placeholder, and never lands in `config.yaml`.

**No raw HTTP to Linear from this skill, ever** — no `curl`, no `urllib`, no GraphQL. The tool list of §6 is the blast-radius bound and a raw call voids it. `curl -H "Authorization: $LINEAR_API_KEY"` is doubly banned: shell expansion puts the plaintext key in curl's argv, visible in `/proc/<pid>/cmdline`.

The transport is a **manual `mcp_servers.linear` stanza** with a static `Authorization: "Bearer ${LINEAR_API_KEY}"` header. Do not "fix" it back to `hermes mcp install linear`: the catalog preset is remote OAuth 2.1 + DCR and registers a VM-loopback `redirect_uri`, so the browser flow can never complete on a headless gateway. One deliberate exception to native-only exists — `engagement-checker`'s cursor-gated delta collector keeps an audited raw GraphQL implementation, because its coverage verdict gates durable cursor advancement. It is annotated in its own file. Do not extend it, do not "clean it up".

### `mcp__linear__save_issue` — create AND update in one tool

There is no `create_issue`. Omit `id` to create; pass `id` to update. Nothing is required globally; `title` and `team` are required when creating. Exercised live: `id`, `title`, `description`, `team`, `state`, `labels`, `priority`, `assignee`. The rest of the table is the server's declared schema, unmeasured — treat it as such.

| param | type | notes |
|---|---|---|
| `id` | string | **update only** — the issue key (e.g. `GCN-9`); a UUID also resolves |
| `title` | string | **required when creating**; round-trips byte-identically, brackets and non-ASCII included |
| `description` | string | Markdown, real newlines |
| `patch` | array | partial atomic edits to existing content |
| `team` | string | **required when creating** — team name or ID |
| `state` | string | state type, name, or ID — **an unresolvable value silently no-ops (§10)** |
| `priority` | number | bare integer `0=None, 1=Urgent, 2=High, 3=Medium, 4=Low` |
| `assignee` | string\|null | user ID, name, email, or `"me"`; null removes |
| `labels` | array | label names or IDs — **REPLACES the full set** (declared schema, never exercised); omitting it is *measured* to preserve the set |
| `project` / `cycle` / `milestone` | string\|null | name, ID, identifier, or slug |
| `parentId` / `duplicateOf` | string\|null | issue ID or identifier |
| `estimate` / `dueDate` | number\|null / string\|null | `dueDate` is ISO |
| `blocks` / `blockedBy` / `relatedTo` / `links` | array | **APPEND-ONLY** (`links` is `[{url, title}]`) |
| `removeBlocks` / `removeBlockedBy` / `removeRelatedTo` | array | inverse ops |
| `slaBreachesAt`, `slaType`, `delegate`, `setReleases`, `addReleases`, `removeReleases` | | out of scope |

### `mcp__linear__list_issues`

Required: none. Params: `assignee`, `createdAt`, `cursor`, `cycle`, `delegate`, `fields`, `includeArchived`, `label`, `limit`, `orderBy`, `parentId`, `priority`, `project`, `query`, `release`, `state`, `team`, `updatedAt`. `assignee` and `team` combine correctly.

**`fields` works and must be used** — an unprojected 250-issue read measured 399,510 characters and blew Hermes' own tool-result budget, so the payload never reached the agent. It is an **array of strings** and a **closed enum**; a bare string, an object, or one invalid member fails the entire call with `Input validation error: … fields.N: Invalid input` and returns no data. `fields: []` means no projection — the full default set.

- **VALID** (measured): `id` · `title` · `description` · `priority` · `url` · `gitBranchName` · `createdAt` · `updatedAt` · `archivedAt` · `completedAt` · `startedAt` · `canceledAt` · `dueDate` · `slaStartedAt` · `slaMediumRiskAt` · `slaHighRiskAt` · `slaBreachesAt` · `status` · `statusType` · `labels` · `createdBy` · `createdById` · `assignee` · `assigneeId` · `team` · `teamId` · `project` · `projectId` · `cycleId` · `parentId` · `estimate`.
- **INVALID — HARD ERRORS, never request these**: `identifier` · `state` · `stateId` · `labelIds` · `cycle` · `relations` · `comments` · `attachments` · `documents`.

`identifier` and `state` are the two names an implementer reaches for by reflex, and both kill the call. Use `id` instead of `identifier`, `status`/`statusType` instead of `state`.

**`state` is single-valued, and a wrong value returns zero issues with no error.** It accepts one state id, name, or type; an **array is a hard error**. **Negation (`"!Done"`) and comma-lists (`"Todo,In Progress"`) return a normal response with zero issues** — a whole board silently rendered empty, with nothing to catch. Never construct one. For non-terminal work use the type: `state: "started"` and `state: "unstarted"` (plus `"backlog"` if wanted), which work cross-team and need no other team's ids. "Not completed" is not expressible in one call — use those calls, or read unfiltered and drop `statusType ∈ {completed, canceled}` client-side.

**`updatedAt` is a bare string and a LOWER bound.** The object form `{"gte": …}` is a hard error. Any upper bound is applied client-side.

**Completeness — `hasNextPage` is the only signal, and the remedy is paging.** The envelope is `issues` / `hasNextPage` / `cursor`; `cursor` is absent when `hasNextPage` is false. There is no total count and no truncation flag. Default page size is **50**; `limit` maxes at **250** (251 is a schema error). **A scan is complete only when `hasNextPage == false`** — counting rows against `limit` is not the test.

When a scan that must be complete returns `hasNextPage: true`, **page it**: re-issue the identical call with `cursor` set to the returned value, and repeat until `hasNextPage` is false. Union the pages and deduplicate on `id`. Resuming from a returned `cursor` is measured to work. Bound the loop with an explicit **page cap of 8**; if the cap is reached while `hasNextPage` is still true, the read is **INCOMPLETE** — name it in `Coverage:` and never act on it or render it as whole.

**Never narrow the filter to make `hasNextPage` go false.** Narrowing changes the question, so the retry's `false` certifies a different and smaller read: silently incomplete and self-certifying at the same time. Change a filter only to change *what is being asked for* — a `state`-type-scoped call is a different read, deliberately — never as a completeness remedy.

**`query` is fuzzy — an optimisation, never the correctness test.** It does search description text as well as titles, but it is relevance-ranked: a hyphenated marker returned all 11 issues of the team, and a URL that appears nowhere in an issue still matched it. `hasNextPage` is **always `true`** in query mode, so a query result carries no completeness guarantee. Confirm every hit by exact substring on the returned `title`/`description`, and never read an empty `query` result as proof of absence — fall back to a plain filtered scan for the negative case.

### `mcp__linear__get_issue`

Required: `id`. Params: `id`, `includeCustomerNeeds`, `includeRelations`, `includeReleases`. `id` takes the issue key (`GCN-11`); a UUID also resolves. Adds `stateHistory` over the list shape — the only place a state UUID appears on a read — and `relations` when asked. Use it for the full `description` when a list result shows the truncation marker.

### `mcp__linear__save_comment`

Required: `body`. Params: `body`, `documentId`, `id`, `initiativeId`, `issueId`, `milestoneId`, `parentId`, `projectId`, `statusUpdateId`, `statusUpdateType`. A comment on an issue is `body` + `issueId`. The response carries **no back-reference to the issue**, so remember which issue was commented on.

## 9. Response shapes — measured

- Every tool result is a JSON object with exactly one key: **`result` on success, `error` on failure**. **The value of `result` is a JSON string — parse it a second time.**
- **There is no success flag.** No `success`, no `ok`. `status` holds the workflow state *name* (`"Todo"`), not an HTTP status.
- **There is no issue UUID and no `identifier` field. `id` IS the key** — `"GCN-11"`. Routing and display both key on `id`. (`list_issues`' `cursor` is the last row's UUID, but harvesting it costs one call per issue; it is not a keying strategy.) Any test requiring both an `identifier` and an `id` can never pass — do not write one.
- **There is no `state` object.** The state is two flat strings: `status` (display name) and `statusType` (`backlog` · `unstarted` · `started` · `completed` · `canceled` · `duplicate`). A predicate on `state.type` reads undefined and drops nothing.
- **`priority` comes back as an object** `{"value":3,"name":"Medium"}` on every native path, though it is written as a bare integer; `engagement-checker` §4's raw collector returns it as a **bare int**. Two paths, two shapes — normalise to an int at the boundary (§2). `labels` comes back as **names** though it is written as ids.
- **Empty values are field-specific: key presence is not a value test.** Measured under an explicit `fields` projection: an unassigned issue carries **no `assignee` key at all**; `completedAt` is **present and `null`** on every issue that is not completed; `labels` is **present and `[]`** when there are none. `cursor` is absent when `hasNextPage` is false. Handle all three shapes — absent, null, empty collection — with a defaulting read, and never decide a value from `key in row`. Extra fields also appear on some issues (`cycleId` on company-team issues).
- `description` is truncated in `list_issues` at ~400 characters, with a literal "… (truncated, use get_issue for full description)" marker appended; `get_issue`, `save_issue` and `engagement-checker` §4's raw collector all return it whole — truncation is a `list_issues` property, not a Linear one. An empty description reads `null` from `get_issue`/`save_issue` but `""` from `list_issues`.
- A create response and an update response are **shape-identical**. Nothing in the payload distinguishes them.
- Team identity is `team` (display name) plus `teamId` (UUID). There is no team-key field; the prefix is only obtainable by parsing `id`.

## 10. The success test — "no error" is not success

A `save_issue` update with an unresolvable `state` returns a full, normal, `result`-bearing payload, changes nothing, and does not advance `updatedAt`. Measured twice, with two different bad values. It cannot be pre-empted by input validation: a nonsense state name and an all-zero UUID were both silently dropped, while a well-formed-but-nonexistent UUID produced a hard 400. The failure mode is not predictable from the input.

A write is confirmed only when **all** of these hold:

1. the result carries `result` and no `error`;
2. `result` parses as JSON;
3. **the parsed payload shows the intent that was sent** — not merely that an issue exists. Compare against the representation the response actually uses, which is not the one the request used:

| sent | assert on the parsed payload |
|---|---|
| `state` id | `statusType` equals that state's type (`unstarted` for Todo, `completed` for Done, `canceled` for Canceled); `status` equals its name |
| `priority` int | the returned `priority` **object**'s `value` equals the int sent |
| `labels` id | the returned `labels` array — **names**, not ids — contains that label's name from §2 |
| `assignee` uuid | `assigneeId` equals it; `assignee` is a display name and never matches |
| create | additionally `id` matches `^[A-Z]+-\d+$` and `teamId` is GCN's |
| comment `body` | the returned `body` equals it |

Assert every field the call set, not just identity: a create that asserts only `id` and `teamId` passes while a silently no-opped `state` leaves the issue in Backlog forever. After every write, read back the exact target independently: `mcp__linear__get_issue` for issues, `mcp__linear__list_comments` for the target issue's comment. Assert the intended fields/body there as well as on the write response; a successful response alone is not final verification.

Anything short of that is a **failure**. Never record a write as done, and never report it to Gaetan as done, on the strength of "no error". Fail loudly: say the write could not be confirmed and leave it for the next cycle.

The same discipline applies to reads: a read that must be complete is complete only when `hasNextPage == false` after paging the `cursor` to exhaustion (§8). An incomplete required read is a failure, not a smaller answer, and never a reason to narrow the filter until it looks complete.

## 11. Rendering

Any list of issues shown to a human is priority-sorted by the §2 rule, on the priority normalised to an int, one line per issue:

`P<n> <id> <title> — <status>`

`<n>` is the priority **normalised to an int** per §2; a missing, null or 0 priority renders `P–`. `<status>` is the flat `status` string. Within the same priority, GCN issues come first — key that on `teamId == 81e7b769-2a46-4e2a-8db5-c165a7963b0e`, not on a parsed prefix. Tolerate absent keys (§9): an issue with no `assignee` is normal, not an error.

## 12. Delegated ticket work

When Gaetan asks Tars to work on a ticket:

1. Move the ticket to **In Progress** before starting the delegated work, and verify the returned `status` / `statusType` under §10.
2. Record delegation provenance in the ticket's comments or description: worker/session type (for example Orca), the session identifier or other durable handle, and the exact UTC start time.
3. Comment again with the verified outcome or blocker when the delegated session finishes. Never report the ticket work as started unless both the state transition and provenance write are confirmed.

Company-team tickets still require Gaetan's explicit per-message instruction naming that team for each write under §1; if that authority is absent, do not start the session under the fiction that the ticket has been updated.

## 13. Update semantics

`labels` **replaces** on `save_issue`: always re-supply every label the issue should keep, and **omit the param entirely** when the update is not about labels — passing it empty wipes the set. Omitting `labels` is measured to preserve them, as is omitting `state`. `blocks` / `blockedBy` / `relatedTo` / `links` are the opposite — **append-only**; undo with `removeBlocks` / `removeBlockedBy` / `removeRelatedTo`.

**Duplicate updates are relation-first and relation-only.** Measured 2026-08-11: sending `duplicateOf` and the Duplicate `state` together fails with `Issues can only be moved to a duplicate state when a duplicate issue relation exists.` Send only `id` + `duplicateOf`; Linear creates the relation and automatically moves the issue to `statusType: "duplicate"`. Then verify both the returned state and `relations.duplicateOf.id` with `get_issue(includeRelations: true)`. Do not send a second state update.

Send strings raw: real newlines in markdown, never a literal `\n` — the server's own instruction. Machine paths use the §2 ids with the human name in a trailing comment: `state: "59ed732a-f242-4eba-926f-1c0d128fe83c"  # Todo`. A chat-driven create may take a name Gaetan typed, but resolve it against §2 before sending and echo which team and state were used in the reply. Never mix a name and an ID for the same field in one file.

## 14. `linear_write` — the contract every write obeys

`op` is `issue_create` · `issue_update` · `comment_create`, mapping onto exactly two tools: `save_issue` without `id`, `save_issue` with `id`, and `save_comment`. Return shapes differ because the measured payloads differ:

- `issue_create` / `issue_update` → `{id, url, status, priority}` — the issue key, its url, the flat state name, and the priority normalised to an int (§2).
- `comment_create` → `{id, created_at}` only. The comment payload carries no `url`, no `status`, and **no back-reference to the issue**, so the caller remembers which issue it commented on.

Every assertion lives inside this contract, never in a caller: §1's team policy, applied to all three ops; an explicit `state` on create; exactly one workload label; an explicit `priority`; `assignee` present. **No `save_issue` call (with or without `id`) and no `save_comment` call may target a team other than GCN unless the current turn carries Gaetan's explicit instruction naming that team.** Absent that instruction, a non-GCN target is refused — create, update and comment equally.

A write counts as done only when §10's three-part test passes. A tool `error`, an unparseable `result`, or a payload that does not show the intended value is a **failure**: fail open to the caller, which logs a bounded coverage note and retries next cycle. Reads fail closed. Every documented call site states which op it uses and on which team.
