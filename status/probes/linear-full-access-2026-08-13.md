# Linear MCP — expand Tars to full content read/write (2026-08-13)

Applied live on the Tars VM (`192.168.0.9`) on 2026-08-13 15:15 UTC.
`mcp_servers.linear.tools.include` in `~/.hermes/config.yaml` went from **10**
bare tool names to **56** of the 58 the server offers. **This deliberately
reverses the GCN-10 prune of 2026-08-11**
(`status/probes/gcn10-linear-prune.md`) on Gaetan's explicit instruction:

> "Tars should be able to have all read/write access (but no admin tools like
> delete teams, ...)"

**Verdict: applied and verified.** The registration-level enumeration returns
exactly the 56 kept names, `tools.mcp_tool` independently logs
`registered 56 tool(s)` (was 10), and a newly-enabled read (`list_projects`)
returned 50 projects. No Linear writes were made as testing. No `sops`, no
token read, no restart, no Slack message.

**Only two tools are excluded, and neither is an admin tool** — see §3.

## 1. Mechanism (unchanged from GCN-10)

The filter is `mcp_servers.<name>.tools.{include,exclude}` in
`~/.hermes/config.yaml`, evaluated by the registration loop in
`tools/mcp_tool.py`. **`include` takes precedence over `exclude`** — a sibling
`exclude` next to an `include` is a silent no-op. This change therefore
rewrites the single `include` list and writes no `exclude` at all.

Names are **bare** (`get_issue`, not `mcp__linear__get_issue`) — the
`mcp__linear__` prefix is a client-side dispatch convention. None of the 56
names contains an fnmatch metacharacter, so each entry is an exact match:
56 entries → 56 tools.

`hermes mcp configure linear` (the curses checklist) was **not** used: ticking
every box deletes both filters, which would fail *open* on every tool Linear
ships in future. `include` fails closed — a new upstream tool does not
auto-register; adding it is one line in the config.

## 2. Before / after

| | before | after |
|---|---|---|
| `tools.include` entries | 10 | 56 |
| Linear tools registered | 10 | 56 |
| Tools offered by the server | 58 | 58 |
| Total MCP tools (3 servers) | 55 | 101 |
| `config.yaml` lines | 284 | 330 |

Before (`~/.hermes/config.yaml.bak-20260813-linear-expand`, md5
`222bff9670b8a4008b69bd8cac2544cf`, 284 lines, mode `-rw-------`):

```yaml
mcp_servers:
  linear:
    url: https://mcp.linear.app/mcp
    headers:
      Authorization: "Bearer ${LINEAR_API_KEY}"   # env-var reference, no literal token
    tools:
      include:
        - save_issue
        - get_issue
        - list_issues
        - list_issue_statuses
        - list_issue_labels
        - list_teams
        - list_users
        - save_comment
        - list_comments
        - delete_comment
```

After (md5 `6274567b504122a07beb37012090033d`, 330 lines, mode `-rw-------`):
the same stanza with the `include:` list replaced by the 56 names of §3, in
the server's own discovery order. `url:` and `headers:` are byte-identical;
the `slack:` and `notion:` stanzas are byte-identical; every key outside
`mcp_servers.linear.tools` is byte-identical.

How it was written: `.bak` + patch inside **one**
`flock ~/.hermes/.wf3.lock`. The writer spliced text rather than
re-serialising YAML (ruamel has measurably reflowed this file before), and
asserted, in order: the 12-line anchor block occurs exactly once;
`out.replace(new, old, 1) == src` (proves *nothing else* in the file moved);
post-parse `include == keep`, `url`/`headers` unchanged, the `linear` stanza
equal outside `tools`, and the whole config equal with `mcp_servers.linear`
nulled on both sides. Written via `os.replace` (atomic, mode 0600) so a live
reload can never read a torn file.

## 3. Categorization — all 58

**Decision rule.** INCLUDE every content-level read and write — the things a
Linear *workspace user* manipulates: issues, comments, labels, projects,
milestones, documents, initiatives, status updates, attachments, cycles,
releases, diffs — including item-scoped deletes, whose *use* is governed by
SOUL's destruction rules rather than by the tool filter. EXCLUDE
admin/org-destructive surface: anything that deletes, creates or modifies
teams, workspaces, users, memberships, webhooks, API keys or integrations, or
performs a bulk-destructive operation.

**Finding: that exclusion category is empty upstream.** Linear's MCP server
ships **no** team/workspace/user/membership/webhook/API-key/integration
*mutation* tool. `list_teams`, `get_team`, `list_users`, `get_user` and
`get_workspace` are pure reads; there is no `save_team`, no `delete_team`, no
`save_user`, no `create_webhook`. So Gaetan's guard rail ("no admin tools like
delete teams") costs nothing — **there is nothing of that kind to exclude.**

### INCLUDED — 56 (the `tools.include` list, verbatim and in file order)

| family | tools | why included |
|---|---|---|
| Attachments (6) | `get_attachment`, `prepare_attachment_upload`, `create_attachment_from_upload`, `create_attachment`, `delete_attachment`, `extract_images` | Item-scoped content on an issue. `delete_attachment` is an item-scoped delete — explicitly in scope; SOUL governs its use. Restores the file-upload capability GCN-10 knowingly dropped. |
| Agent skills (2) | `list_agent_skills`, `get_agent_skill` | Reads of Linear's own agent-skill registry. Read-only, no org mutation. |
| Comments (3) | `list_comments`, `save_comment`, `delete_comment` | Full comment lifecycle. Already registered before this change. |
| Cycles (1) | `list_cycles` | Read. Named in scope. |
| Documents (3) | `get_document`, `list_documents`, `save_document` | Content read/write. Named in scope. |
| Issues (3) | `get_issue`, `list_issues`, `save_issue` | The core lifecycle. Already registered. |
| Issue statuses (2) | `list_issue_statuses`, `get_issue_status` | Workflow **reads**. `get_issue_status` was pruned by GCN-10 as redundant with `list_issue_statuses`; redundancy is no longer a reason to withhold a read. |
| Issue labels (2) | `list_issue_labels`, `create_issue_label` | Labels named in scope. Label creation is workspace taxonomy, not org admin — it creates no team, user or permission. |
| Projects (4) | `list_projects`, `get_project`, `save_project`, `list_project_labels` | Content read/write. Named in scope. |
| Releases (7) | `list_release_pipelines`, `list_releases`, `get_release`, `save_release`, `list_release_notes`, `get_release_note`, `save_release_note` | **Borderline, included** — see §4. |
| Diffs, non-authority (6) | `get_diff`, `list_diffs`, `get_diff_threads`, `save_diff_comment`, `resolve_diff_thread`, `delete_diff_comment` | **Borderline, included** — see §4. Reads plus comment-level writes on a diff; the comment lifecycle Gaetan already granted, applied to a diff instead of an issue. |
| Milestones (3) | `list_milestones`, `get_milestone`, `save_milestone` | Content read/write. Named in scope. |
| Teams (2) | `list_teams`, `get_team` | **Reads only** — no team mutation tool exists. |
| Users (2) | `list_users`, `get_user` | **Reads only** — no user/membership mutation tool exists. |
| Workspace (1) | `get_workspace` | **Read only** — workspace metadata; no workspace mutation tool exists. |
| Docs (1) | `search_documentation` | Searches Linear's public product docs. Touches no workspace data at all. |
| Initiatives (5) | `list_initiatives`, `get_initiative`, `save_initiative`, `list_initiative_labels`, `create_initiative_label` | Content read/write. Named in scope. |
| Status updates (3) | `get_status_updates`, `save_status_update`, `delete_status_update` | Project/initiative status updates, named in scope. `delete_status_update` archives one update — item-scoped. |

### EXCLUDED — 2

| tool | rationale |
|---|---|
| `merge_diff` | Merges a Linear diff (or adds it to the merge queue). **SOUL rule 2: "Tars never merges, approves or pushes"** — one scoped exception, its own skill mirror, which is a GitHub PR and not this. Not an admin tool; excluded because the tool surface must not offer an action the soul forbids. |
| `submit_diff_review` | Approves a diff, requests changes, or submits a review. Same SOUL rule 2 clause, the "approves" half. |

**Flag for Gaetan:** these two are the *only* thing between the current state
and literally-all-58. They are excluded on SOUL rule 2, not on his "no admin
tools" guard rail. If he wants Tars to be able to merge or approve Linear
diffs, that is a SOUL amendment first and a two-line config edit second — say
so and it takes a minute.

Arithmetic: 56 + 2 = 58. `set(include) ∩ set(exclude)` = ∅; the union equals
the live `hermes mcp test linear` output name-for-name. Checked
programmatically, not by eye: the include list was *generated* from the live
server output minus the two exclusions, so a typo is structurally impossible
under `tools.include`, where a typo silently drops a tool instead of erroring.

## 4. Borderline calls and their reasoning

**Releases (7) — included.** Not named in either bucket. A release, a release
note and a release pipeline are workspace *content* records, in the same
category as documents and projects: `save_release` / `save_release_note`
create or update a record, the server offers no `delete_release`, and none of
them touches a team, a user, a permission or an integration. Under "all
read/write minus org admin" they land on the read/write side. Withholding them
would be re-applying GCN-10's stricter "issue-lifecycle-only" bar, which is
precisely the bar Gaetan just lifted.

**Diff reads and diff comments (6) — included.** GCN-10 pruned the whole
8-tool diff family under SOUL rule 2. That rule names three verbs: merge,
approve, push. `get_diff`, `list_diffs` and `get_diff_threads` are reads;
`save_diff_comment`, `delete_diff_comment` and `resolve_diff_thread` are
comment-level writes — the same lifecycle Gaetan granted for issue comments,
pointed at a diff. Only `merge_diff` and `submit_diff_review` actually perform
the forbidden verbs, so the exclusion is narrowed to exactly those two rather
than to the family. Resolving a comment thread is thread hygiene, not review
authority: it neither approves nor blocks a merge.

**`create_issue_label` / `create_initiative_label` — included.** Label
*creation* was forbidden in prose by `linear-ticketing` §6 and pruned by
GCN-10 as "admin". A label is workspace taxonomy, not org structure — it grants
no access and deletes nothing — and Gaetan named labels in the read/write
bucket. Included; §1's GCN-default write policy still governs *where* a label
gets created.

**Item-scoped deletes (4: `delete_comment`, `delete_attachment`,
`delete_status_update`, `delete_diff_comment`) — included.** Each destroys one
item, none is bulk, none is org-level. Reachability is not permission: SOUL's
destruction rules govern the call, and `linear-ticketing` §6 now says so
explicitly.

**Not treated as borderline:** `get_workspace`, `get_team`, `get_user`,
`list_teams`, `list_users` read org objects but mutate nothing, so the "no
admin" rail does not touch them.

## 5. Probes

### 5a. Step 0 — dry run of the filter itself, run BEFORE any write

Hermes' own matcher, Hermes' own interpreter, against the live 58 names on the
VM (`cd ~/.hermes/hermes-agent && .venv/bin/python /tmp/linear_expand.py`):

```
MATCHER=hermes TOTAL 58 KEPT 56 DROPPED 2
DROPPED_NAMES ['merge_diff', 'submit_diff_review']
DRYRUN_OK
```

`MATCHER=hermes` means `tools.mcp_tool._normalize_name_filter` +
`matches_name_filter` imported and ran — the real matcher, not an fnmatch
stand-in. The generator asserted 58 unique `[a-z_]+` names parsed from
`hermes mcp test linear`, that both exclusion targets and all 10 previously
included names are among them, and that no kept name contains `*?[]`. Stop
condition: no `DRYRUN_OK`, no edit.

### 5b. `hermes mcp list` (config level)

```
  Name             Transport                      Tools        Status
  linear           https://mcp.linear.app/mcp     56 selected  ✓ enabled
  slack            docker run -i                  all          ✓ enabled
  notion           docker run -i                  all          ✓ enabled
```

### 5c. Registration-level enumeration (authoritative)

```
ssh gaetan@192.168.0.9 "~/.local/bin/hermes chat -Q -q 'List the exact names of every tool you can call whose name starts with mcp__linear__. Output only the bare names, one per line, nothing else.'"
```

`session_id: 20260813_151541_a9c699` → **56 names**, alphabetically:
`create_attachment`, `create_attachment_from_upload`, `create_initiative_label`,
`create_issue_label`, `delete_attachment`, `delete_comment`,
`delete_diff_comment`, `delete_status_update`, `extract_images`,
`get_agent_skill`, `get_attachment`, `get_diff`, `get_diff_threads`,
`get_document`, `get_initiative`, `get_issue`, `get_issue_status`,
`get_milestone`, `get_project`, `get_release`, `get_release_note`,
`get_status_updates`, `get_team`, `get_user`, `get_workspace`,
`list_agent_skills`, `list_comments`, `list_cycles`, `list_diffs`,
`list_documents`, `list_initiative_labels`, `list_initiatives`,
`list_issue_labels`, `list_issue_statuses`, `list_issues`, `list_milestones`,
`list_project_labels`, `list_projects`, `list_release_notes`,
`list_release_pipelines`, `list_releases`, `list_teams`, `list_users`,
`prepare_attachment_upload`, `resolve_diff_thread`, `save_comment`,
`save_diff_comment`, `save_document`, `save_initiative`, `save_issue`,
`save_milestone`, `save_project`, `save_release`, `save_release_note`,
`save_status_update`, `search_documentation`.

**Absent: `merge_diff` and `submit_diff_review`** — the two exclusions, and
nothing else is missing.

### 5d. Independent corroboration — `~/.hermes/logs/agent.log`

```
2026-08-13 14:46:50,248 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): …
2026-08-13 14:46:50,248 INFO tools.mcp_tool: MCP: registered 55 tool(s) from 3 server(s)
--- config edit 15:15:16 UTC ---
2026-08-13 15:15:44,042 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 56 tool(s): mcp__linear__get_atta…
2026-08-13 15:15:44,042 INFO tools.mcp_tool: MCP: registered 101 tool(s) from 3 server(s)
```

Same code path the gateway uses. 10 → 56, total 55 → 101.

### 5e. One benign read through a newly-enabled tool

`list_projects` was pruned by GCN-10 and is now registered.
`session_id: 20260813_151612_cd382e`, prompt: *"Call
`mcp__linear__list_projects` once with no filters and report only how many
projects came back and their names. Make no writes."*

```
50 projects:
- Automatisation de la facturation
- EB FRE #1 : Suppression systématique des FREM ≤ seuil si premier FREM
…
- BNP PF financer
```

50 rows across the company workspace. **No writes were made anywhere in this
work** — the write path was not exercised deliberately.

## 6. The workflow-status question — answered plainly

The original ask behind this change was for Tars to create a **"Blocked"**
workflow status.

**No tool in Linear's MCP server can create, rename or delete a team's
workflow status.** The complete status surface is two **reads**:

- `list_issue_statuses` — "List available issue statuses in a Linear team"
- `get_issue_status` — "Retrieve detailed information about an issue status…"

There is no `save_issue_status`, no `create_issue_status`, no
`create_workflow_state`. (`save_status_update` / `get_status_updates` /
`delete_status_update` are a *different* concept: prose status updates posted
on a **project or initiative**, not an issue's workflow state.)

**Consequence: Tars can now SET an issue's status to any state that already
exists on the team (`save_issue.state`), and can read every team's status
list — but it still cannot CREATE one.** A new state such as "Blocked" has to
be added by a human in Linear's team settings (Settings → Team → Workflow).
This is an upstream limitation of `mcp.linear.app`, not a consequence of the
tool filter: it was equally impossible before this change, and no wider filter
can fix it. Once Gaetan adds the state in the UI, Tars can use it immediately
with no further config change.

## 7. Consequences and residual risk

- **Blast radius grew, by instruction.** `LINEAR_API_KEY` was always an
  unscoped full-write personal key; the filter only ever bounded what the
  model could *reach*. Tars can now reach project, document, initiative,
  milestone and release writes, label creation, attachment upload/delete, and
  four item-scoped deletes across the whole workspace, company teams included.
  Two things still bound behaviour and neither is the filter:
  `linear-ticketing` §1 (GCN by default; a company team only on Gaetan's
  per-message instruction naming it) and SOUL's destruction rules.
- **Context cost.** 46 more tool schemas are injected into every Linear-capable
  turn. `hermes prompt-size` reports native toolsets only and does not price
  MCP schemas, so this is **unmeasured**; if turns get slower or noisier, the
  first knob is trimming families Tars never uses (agent skills, releases,
  diffs) back out of the include list.
- **Still fails closed on upstream growth and on renames.** A tool Linear ships
  tomorrow will not auto-register — deliberate. The same property means a
  *renamed* kept tool silently disappears; the `*/30` engagement-checker cron
  remains the free regression watch.
- **Rollback** is one file: `cp ~/.hermes/config.yaml.bak-20260813-linear-expand
  ~/.hermes/config.yaml` under `flock ~/.hermes/.wf3.lock` restores the GCN-10
  10-tool surface exactly (md5 `222bff9670b8a4008b69bd8cac2544cf`).

## 8. Skill updates

`skills/orchestration/linear-ticketing/SKILL.md` **1.3.0 → 1.4.0**, §6 only:

- Heading "Allowed tools — **closed list**" → "Allowed tools"; the five-tool
  list is now the *default working set*, not a "nothing else, ever" bound.
- The GCN-10 paragraph ("**10** register, not 58 … respect the list as
  policy") is replaced by the new surface: 56 of 58 register, the two
  exclusions and why, and a pointer to this file.
- New explicit statements: registered is still not permission to use (§1's
  write policy and SOUL's destruction rules govern every write); no
  workflow-status creation exists upstream; `include` fails closed and a
  sibling `exclude` would be a silent no-op.

Mirrored byte-identical to the VM's live
`~/.hermes/skills/orchestration/linear-ticketing/SKILL.md` (md5
`334fccffb975fc104654f0ab277421ab` on both sides; the VM copy was confirmed
un-drifted, md5 `547794e4e2860bed7bfc9a439c81419d`, immediately before the
overwrite).

**No other skill needed a change.** `engagement-checker`, `daily-work-brief`
and `helptech-duty-intake` were grepped for tool-count, registration and
filter claims: they carry none — the only hit, `engagement-checker`'s
"Permitted writes", is **behavioural policy and not a mirror of the filter**,
so it was deliberately left untouched and its write restrictions still stand
in full.
