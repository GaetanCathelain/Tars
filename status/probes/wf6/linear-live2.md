# Linear live facts — BLOCKED (MCP session expired)

Task: read-only Linear MCP calls (list_teams, list_issue_labels,
list_issue_statuses, save_issue schema, list_issues) to answer 5 questions
about a new personal Linear team, labels, save_issue schema, priority
representation, and label/team creation ability.

## What happened

1. `ToolSearch("select:mcp__claude_ai_Linear__list_teams,...")` succeeded and
   returned full schemas for `list_teams`, `list_issue_labels`,
   `list_issue_statuses`, `save_issue`, `list_issues`, `create_issue_label`
   (captured below — item 3 and 5 answered from schema alone, no live call
   needed for those).
2. Every actual invocation — `list_teams`, `list_issue_labels`, `list_issues`
   (tried twice each in parallel, then `list_teams` alone a third time) —
   returned: `MCP server "claude.ai Linear" session expired`.
3. No tool available to this subagent re-authenticates that connector
   session. It is a claude.ai-side OAuth/session connector (like Slack/Gmail
   in this same tool family), not something `ssh`/`sops`/local config can
   restart.

**Items 1, 2 (team-scoping check), 4 could NOT be answered — no live Linear
read succeeded.** Re-run this task in a fresh session/turn where the Linear
connector session is valid (e.g. after the user re-triggers the connector in
claude.ai, or in a session where a prior Linear call already worked this
turn).

## What IS answered, from the tool schema alone (ToolSearch, no live call)

### Q3 — `save_issue` schema (mcp__claude_ai_Linear__save_issue)

Full input schema returned by ToolSearch (see tool-search result in this
transcript). Key facts:

- **Create vs update**: single tool does both. `id` omitted → **create**
  (requires `title` + `team`); `id` present → update existing issue. So yes,
  it CAN create, not just update.
- **team**: `team` param — "Team name or ID (required when creating)".
- **labels**: `labels` param — array of label names or IDs. **Replaces the
  full label set** on save (not additive) — existing labels not included are
  removed. This matters for the workload-label design: every save_issue call
  that sets one label must re-supply all labels the issue should keep.
- **priority**: `priority` param, type number, explicit enum in description:
  **`0=None, 1=Urgent, 2=High, 3=Medium, 4=Low`**.
- **assignee**: `assignee` param (NOT `assigneeId`) — accepts user ID, name,
  email, or `"me"`; `null` removes it.
- **state**: `state` param — "State type, name, or ID" (workflow status).
- Other notable fields: `project`, `cycle`, `milestone`, `estimate`,
  `dueDate`, `parentId`, `delegate` (agent assignment), `blockedBy`/`blocks`/
  `relatedTo` (issue relations, append-only), `links` (append-only),
  `description` (Markdown, literal newlines not `\n` escapes — per the
  claude.ai Linear MCP server-level instruction surfaced this session),
  `patch` (anchor-based text edits to description, update-only).

Exact param name list relevant to the goal: `id, title, team, state,
priority, assignee, labels, project, cycle, milestone, estimate, dueDate,
description, parentId, delegate`.

### Q5 — label/team/view creation ability (from ToolSearch schema)

- **`create_issue_label`** exists (confirmed in this session's tool list and
  schema loaded): params `name` (required), `color` (hex), `description`,
  `isGroup`, `parent` (parent label group name), and critically
  **`teamId`: "Team UUID (omit for workspace label)"**. So the same tool
  creates either a **workspace-level label** (omit teamId) or a
  **team-scoped label** (pass teamId) — confirms team-scoped labels are a
  real Linear concept the MCP can create into.
- No `create_team` tool in the loaded/available Linear tool set (session's
  full Linear tool list, visible in the system reminder, has ~60 Linear
  tools — team creation is not among them: only `get_team`/`list_teams`
  read team data). **Team creation is not exposed via this MCP** — would
  need Linear's web UI or app-level API access outside this MCP surface.
- No `create_issue_status`/state-creation tool either (only
  `get_issue_status`/`list_issue_statuses` read). Workflow states, like
  teams, are not MCP-creatable.
- No project/view creation tool distinct from `save_project`
  (create-or-update, same pattern as `save_issue`) — but no dedicated
  "create a Linear *view* (saved filter/board)" tool exists in the listed
  set. Views are a Linear UI-level construct with no corresponding MCP tool
  here.

### Q3 answers `list_issue_labels` schema (structural only, not live data)

`list_issue_labels` takes an optional `team` param ("Team name or ID") —
confirms the tool itself supports scoping the label listing to one team,
consistent with labels being either workspace-wide or team-scoped in
Linear's data model. Cannot confirm from schema alone whether any
team-scoped labels currently exist in this workspace — needs the blocked
live call.

## Unanswered — needs a working Linear MCP session

- **Q1**: Does a new personal/single-member team already exist beyond the 8
  known company teams? Name+key+id of every team. NOT ANSWERED.
- **Q2**: Full workspace-level label list (name+id); do team-scoped labels
  exist on any team today? NOT ANSWERED (only the schema-level capability is
  confirmed above).
- **Q4**: Confirm from `list_issues` output shape how priority appears in
  practice and whether Gaetan's existing issues use it (the 0–4 enum is
  confirmed from the `save_issue` schema doc string, but not cross-checked
  against real issue data). NOT ANSWERED.

## Recommended next step

Re-run this exact task (same file, same tool list) in a session where a
Linear MCP call has already succeeded, or after the user reopens/re-auths
the claude.ai Linear connector. Do not attempt to work around the session
expiry (no local credential/token for this connector exists in this repo —
it's a hosted OAuth session, out of scope for `sops`/ssh secret flows in
this repo's CLAUDE.md).
