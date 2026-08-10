# GCN-3 / GCN-5 — moved to Done

Two precise Linear state moves on the Tars VM, driven natively through
`~/.local/bin/hermes chat -Q -q '<instruction>'` → `mcp__linear__*` tools (per
`status/probes/wf6/gcn9-native-mcp.md`). Run date: 2026-08-10 (VM clock UTC).

Method: each phase (baseline read, comment, state move, re-read, board pull)
was driven by a small local Python script (`subprocess.run([...])`, argv list
— no shell quoting of comment text) copied to the VM, executed with
`python3 <script>.py`, then deleted. No secret was printed, echoed, or
persisted; no `sops` command was run; no git command was run; no file outside
this deliverable was touched.

---

## GCN-3 — "engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard"

`id: 174ff5fd-2253-43a8-a00a-7971b2bde482`

### Before (baseline `get_issue`)

```json
{"identifier":"GCN-3","status":"In Progress","statusType":"started","labels":["fix"],"priority":2,"assignee":"Gaëtan Cathelain"}
```

### Comment posted (`save_comment`, `issueId` + `body` only)

commentId: `1d920f9d-54ca-4b2e-b790-76a2dbf41216`

> Spec check 2 passes all four sub-steps — a detected loop landed as GCN-12
> (verified by independent re-read: team GCN, assignee Gaetan, one label,
> explicit priority, state Todo never Backlog, source handle at the head of
> the description); the second cycle filed no duplicate; closing the issue
> cleared the item. Separately, the `closed_at` reopen rule was verified
> live — reopening GCN-12 returned the item to open, deleted the latch,
> re-emitted the reminder, and minted no new issue. Deployed v1.7.1 on the
> VM, sha256-matched to the repo.

Verified by a separate `list_comments` read-back (not the mutation
response) — body matched verbatim, character for character.

### State move (`save_issue`, `id` + `state` ONLY — no `labels`)

```json
{"id":"GCN-3","state":"Done","success":true}
```

### After (independent re-read, `get_issue`, fresh tool call)

```json
{"identifier":"GCN-3","status":"Done","statusType":"completed","labels":["fix"],"priority":2,"assignee":"Gaëtan Cathelain"}
```

**Confirmed**: `statusType` `started` → `completed`. `labels` (`["fix"]`),
`priority` (`2`), `assignee` (`Gaëtan Cathelain`) all byte-identical to
baseline.

---

## GCN-5 — "Tars default-team rule: create tickets in GCN unless told otherwise"

`id: 2e2d0eea-60e2-4a7f-b72d-cc6cfb1e78ed`

### Before (baseline `get_issue`)

```json
{"identifier":"GCN-5","status":"In Progress","statusType":"started","labels":["investigation"],"priority":2,"assignee":"Gaëtan Cathelain"}
```

### Comment posted (`save_comment`, `issueId` + `body` only)

commentId: `b257933d-32d5-4e82-a47c-8141ef6364a2`

> The `linear-ticketing` skill is live (v1.3.0), carrying the default-team
> rule (create in GCN unless Gaetan names another team per message, never
> standing), the never-close-what-Tars-didn't-file rule, required fields
> with explicit state (Backlog is the silent default), label/priority
> rubrics and the closed tool list. Exercised end to end by check 2's
> creation of GCN-12.

Verified by a separate `list_comments` read-back — body matched verbatim.

### State move (`save_issue`, `id` + `state` ONLY — no `labels`)

```json
{"id":"GCN-5","state":"Done","success":true}
```

### After (independent re-read, `get_issue`, fresh tool call)

```json
{"identifier":"GCN-5","status":"Done","statusType":"completed","labels":["investigation"],"priority":2,"assignee":"Gaëtan Cathelain"}
```

**Confirmed**: `statusType` `started` → `completed`. `labels`
(`["investigation"]`), `priority` (`2`), `assignee` (`Gaëtan Cathelain`) all
byte-identical to baseline.

---

## Final GCN board (`list_issues`, team=GCN, includeArchived=false)

| Key | Title | Status | statusType | Priority |
|---|---|---|---|---|
| GCN-1 | Bootstrap GCN team + workload labels + priority view | Done | completed | 1 |
| GCN-2 | LINEAR_API_KEY scope check on the VM | Done | completed | 1 |
| GCN-3 | engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard | **Done** | **completed** | 2 |
| GCN-4 | daily-work-brief: pull Linear board + render priority-sorted | In Progress | started | 2 |
| GCN-5 | Tars default-team rule: create tickets in GCN unless told otherwise | **Done** | **completed** | 2 |
| GCN-6 | cooper defaults: Claude/Orca sessions create Linear tickets in GCN by default | Done | completed | 3 |
| GCN-7 | WF6 E2E verification + docs/facts + status log updates | In Progress | started | 2 |
| GCN-8 | WF6 T6 verification probe — disposable, ignore | Canceled | canceled | 4 |
| GCN-9 | Wire Linear MCP server natively into Hermes (catalog preset linear) | Done | completed | 1 |
| GCN-10 | Prune the Linear MCP tool surface exposed to Tars | Todo | unstarted | 3 |
| GCN-11 | WF6 probe — disposable [EC-ZZ99] | Canceled | canceled | 3 |
| GCN-12 | Create and send the Verdict Google OAuth client | Todo | unstarted | 3 |

### Scope confirmation

- **Only GCN-3 and GCN-5 changed** (both `In Progress`/`started` →
  `Done`/`completed`). Every other issue's status matches what a pre-change
  read would show for untouched work.
- **GCN-12 is non-terminal**: `status: Todo`, `statusType: unstarted` — not
  closed, per the instruction that it is a live commitment.
- **GCN-4, GCN-7, GCN-10 untouched**: still `In Progress` / `In Progress` /
  `Todo` respectively — no tool call in this run referenced any of them.
