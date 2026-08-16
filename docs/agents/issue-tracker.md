# Issue tracker: Linear (team GCN)

Issues and specs for this repo live in **Linear**, team **GCN** (`Gaetan`,
personal). Use the Linear MCP tools (`mcp__claude_ai_Linear__*`) for all
operations — never scrape the web UI. Fetch tool schemas via ToolSearch
(`select:mcp__claude_ai_Linear__save_issue,...`) before calling them.

Default every new issue to team **GCN** unless a skill or the user names another
team.

## Conventions

- **Create an issue**: `save_issue` with `title`, `description` (markdown), and
  `teamId` for GCN. Omit `id` to create.
- **Read an issue**: `get_issue` by id or identifier (e.g. `GCN-59`). Use
  `list_comments` for the discussion thread.
- **List / find issues**: `list_issues` filtered by `teamId`, `stateId`,
  `assigneeId`, or `labelIds`. Use `search_documentation` only for Linear's own
  docs, not repo issues.
- **Comment**: `save_comment` with `issueId` and `body`.
- **Apply / remove labels**: `list_issue_labels` to resolve names → ids, then
  `save_issue` with the full desired `labelIds` set (Linear replaces, so include
  labels you want to keep). Create a missing label with `create_issue_label`.
- **Change state (incl. close)**: `list_issue_statuses` for the team → `save_issue`
  with the target `stateId` (e.g. a `Done`/`Canceled` workflow state).

## Triage labels

Roles map to Linear labels per `docs/agents/triage-labels.md`. Resolve the label
name to its id with `list_issue_labels` before applying; create it with
`create_issue_label` on the GCN team if it doesn't exist yet.

## PRs as a request surface

**No.** This repo does not treat external PRs as feature requests; `triage`
ignores the PR queue. (Linear has no native PR-request surface — flip this only
if you wire GitHub PRs into Linear via an integration and want them triaged.)

## When a skill says "publish to the issue tracker"

Create a Linear issue on team GCN with `save_issue`.

## When a skill says "fetch the relevant ticket"

`get_issue` by identifier, then `list_comments` for the thread.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a parent issue; **tickets** are its
sub-issues.

- **Map**: one Linear issue labelled `wayfinder:map`, its description holding the
  Notes / Decisions-so-far / Fog body.
- **Child ticket**: an issue with `parentId` set to the map. Label
  `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`). Assign to the
  driving dev on claim.
- **Blocking**: Linear **issue relations** — add a `blocks`/`blocked-by` relation
  between issues (the UI-visible, canonical representation). A ticket is
  unblocked when every blocker reaches a completed/canceled state.
- **Frontier query**: `list_issues` for the map's open children (`parentId` =
  map), drop any with an open blocker relation or an assignee; first in map order
  wins.
- **Claim**: `save_issue` setting `assigneeId` to the current user — the
  session's first write.
- **Resolve**: `save_comment` with the answer, move the issue to a `Done` state,
  then append a context pointer to the map's Decisions-so-far.
