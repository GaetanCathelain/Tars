# Empty-file guard ticket — GCN-13

Filed via native `mcp__linear__*` tools in the Hermes agent loop on the Tars
VM (`hermes chat -Q -q '<instruction>'`), per `docs/specs/wf6-linear-integration.md`
and the tool schemas measured in `status/probes/wf6/gcn9-native-mcp.md`.
Run date: 2026-08-10 (VM clock UTC).

---

## Create call

Instruction driven through `hermes chat -Q -q`: call `mcp__linear__save_issue`
once, no `id` (create), with `team`, `assignee`, `state`, `labels`, `priority`
all passed explicitly by ID.

Params sent:

```
team:      81e7b769-2a46-4e2a-8db5-c165a7963b0e   (GCN)
assignee:  4951b192-e49c-4b7e-b491-58c89e66043c   (Gaetan)
state:     59ed732a-f242-4eba-926f-1c0d128fe83c   (Todo, EXPLICIT)
labels:    [a0973525-a091-412d-abf1-aed30bedf60a] (investigation)
priority:  3
title:     Guard Tars skill self-merge path against empty-file commits
description: (see raw result below — ends in the required Spec: line)
```

Raw tool result (verbatim, `save_issue` response):

```json
{"id":"GCN-13","title":"Guard Tars skill self-merge path against empty-file commits","description":"On 2026-08-10, during a WF6 verification cycle, Tars used skill_manage to rewrite its own skills/himalaya/SKILL.md and self-merged PR #42 containing an EMPTY file, then self-merged PR #43 restoring it (+305 lines) 22 seconds later.\n\nTars self-corrected, but the self-edit path can commit an empty skill file to main — a window in which the live skill mirror is empty. SOUL rule 2 authorises Tars to land its own skill edits by self-merged PR, so the guard belongs in that path (e.g. refuse to open or merge a PR whose skill file is empty or drastically shorter than its parent).\n\nThis was observed incidentally during WF6 and is not a WF6 regression; four self-merges happened that session (#40–#43).\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":{"value":3,"name":"Medium"},"url":"https://linear.app/mobile-club/issue/GCN-13","gitBranchName":"gcn-13-guard-tars-skill-self-merge-path-against-empty-file-commits","createdAt":"2026-08-10T22:21:20.280Z","updatedAt":"2026-08-10T22:21:20.280Z","archivedAt":null,"completedAt":null,"startedAt":null,"canceledAt":null,"dueDate":null,"slaStartedAt":null,"slaMediumRiskAt":null,"slaHighRiskAt":null,"slaBreachesAt":null,"status":"Todo","statusType":"unstarted","labels":["investigation"],"attachments":[],"documents":[],"createdBy":"Gaëtan Cathelain","createdById":"4951b192-e49c-4b7e-b491-58c89e66043c","assignee":"Gaëtan Cathelain","assigneeId":"4951b192-e49c-4b7e-b491-58c89e66043c","team":"Gaetan","teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e"}
```

**GCN-13** — `statusType: unstarted` (Todo, not Backlog), priority 3 (Medium),
label `investigation`, assignee/team match the required IDs.

---

## Independent verification — `get_issue` re-read

A **separate** `hermes chat -Q -q` call, new session, calling
`mcp__linear__get_issue` with `id="GCN-13"` — not the mutation response above.

Raw tool result (verbatim):

```json
{"id":"GCN-13","title":"Guard Tars skill self-merge path against empty-file commits","description":"On 2026-08-10, during a WF6 verification cycle, Tars used skill_manage to rewrite its own skills/himalaya/SKILL.md and self-merged PR #42 containing an EMPTY file, then self-merged PR #43 restoring it (+305 lines) 22 seconds later.\n\nTars self-corrected, but the self-edit path can commit an empty skill file to main — a window in which the live skill mirror is empty. SOUL rule 2 authorises Tars to land its own skill edits by self-merged PR, so the guard belongs in that path (e.g. refuse to open or merge a PR whose skill file is empty or drastically shorter than its parent).\n\nThis was observed incidentally during WF6 and is not a WF6 regression; four self-merges happened that session (#40–#43).\n\nSpec: Tars repo docs/specs/wf6-linear-integration.md (WF6)","priority":{"value":3,"name":"Medium"},"url":"https://linear.app/mobile-club/issue/GCN-13","gitBranchName":"gcn-13-guard-tars-skill-self-merge-path-against-empty-file-commits","createdAt":"2026-08-10T22:21:20.280Z","updatedAt":"2026-08-10T22:21:20.280Z","archivedAt":null,"completedAt":null,"startedAt":null,"canceledAt":null,"dueDate":null,"slaStartedAt":null,"slaMediumRiskAt":null,"slaHighRiskAt":null,"slaBreachesAt":null,"status":"Todo","statusType":"unstarted","labels":["investigation"],"attachments":[],"documents":[],"stateHistory":[{"state":{"id":"59ed732a-f242-4eba-926f-1c0d128fe83c","name":"Todo","type":"unstarted"},"startedAt":"2026-08-10T22:21:20.280Z","endedAt":null}],"createdBy":"Gaëtan Cathelain","createdById":"4951b192-e49c-4b7e-b491-58c89e66043c","assignee":"Gaëtan Cathelain","assigneeId":"4951b192-e49c-4b7e-b491-58c89e66043c","team":"Gaetan","teamId":"81e7b769-2a46-4e2a-8db5-c165a7963b0e","relations":{"blocks":[],"blockedBy":[],"relatedTo":[],"duplicateOf":null}}
```

Confirmed independently: state `id` `59ed732a-f242-4eba-926f-1c0d128fe83c`
("Todo", `type: unstarted`), team `81e7b769-2a46-4e2a-8db5-c165a7963b0e`,
assignee `4951b192-e49c-4b7e-b491-58c89e66043c`, exactly one label
(`investigation`), priority `3` (Medium). No relations (none requested).

---

## Final board — `list_issues(team=GCN, limit=100)`, verbatim summary

Pulled in a third, independent `hermes chat -Q -q` call after the create.

| id | title | status | statusType | priority | label(s) | updatedAt |
|---|---|---|---|---|---|---|
| GCN-13 | Guard Tars skill self-merge path against empty-file commits | Todo | unstarted | 3 Medium | investigation | 2026-08-10T22:21:20.280Z |
| GCN-12 | Create and send the Verdict Google OAuth client | Todo | unstarted | 3 Medium | fix | 2026-08-10T21:19:51.823Z |
| GCN-11 | WF6 probe — disposable [EC-ZZ99] | Canceled | canceled | 3 Medium | test-check | 2026-08-10T18:00:16.535Z |
| GCN-10 | Prune the Linear MCP tool surface exposed to Tars | Todo | unstarted | 3 Medium | access | 2026-08-10T17:16:01.938Z |
| GCN-9 | Wire Linear MCP server natively into Hermes (catalog preset linear) | Done | completed | 1 Urgent | access | 2026-08-10T17:15:10.783Z |
| GCN-8 | WF6 T6 verification probe — disposable, ignore | Canceled | canceled | 4 Low | (none) | 2026-08-10T16:34:34.585Z |
| GCN-7 | WF6 E2E verification + docs/facts + status log updates | In Progress | started | 2 High | test-check | 2026-08-10T16:40:34.253Z |
| GCN-6 | cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default | Done | completed | 3 Medium | discovery | 2026-08-10T16:39:47.842Z |
| GCN-5 | Tars default-team rule: create tickets in GCN unless told otherwise | Done | completed | 2 High | investigation | 2026-08-10T22:05:55.297Z |
| GCN-4 | daily-work-brief: pull Linear board + render priority-sorted | Done | completed | 2 High | fix | 2026-08-10T22:13:25.366Z |
| GCN-3 | engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard | Done | completed | 2 High | fix | 2026-08-10T22:05:52.598Z |
| GCN-2 | LINEAR_API_KEY scope check on the VM | Done | completed | 1 Urgent | access | 2026-08-10T16:24:50.537Z |
| GCN-1 | Bootstrap GCN team + workload labels + priority view | Done | completed | 1 Urgent | access | 2026-08-10T16:24:50.167Z |

**Nothing else changed.** GCN-13 is the only issue whose `updatedAt` falls in
the window of this run (~22:21–22:22 UTC). GCN-12 is `Todo`/`unstarted` —
non-terminal, as required. GCN-7 (`In Progress`, updated 16:40:34) and GCN-10
(`Todo`, updated 17:16:01 — identical to its `createdAt`) both carry
pre-existing `updatedAt` timestamps well before this run and were not touched
by it.

---

## Method notes

- No `sops`, no `.env` cat, no secret on argv — only `hermes chat -Q -q`
  driving native `mcp__linear__*` tools already wired into the agent (per
  GCN-9). The instruction text (non-secret) was piped to a scratch file on the
  VM (`~/wf6-ticket-query.txt` etc.) and read via `$(cat …)` to avoid shell
  quoting issues with the multi-paragraph description; no secret ever touched
  that path.
- No git command was run. No file under `skills/` was touched. No other
  Linear issue was created, updated, or deleted.
