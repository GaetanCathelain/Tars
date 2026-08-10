# GCN-9 — native Linear MCP on the Tars VM

**VERDICT: GCN-9 DONE**

Linear MCP wired natively into Hermes on 192.168.0.9. `hermes mcp test linear`
green (58 tools), `mcp__linear__*` visibility **PROVEN** in the agent loop by an
end-to-end tool call, gateway `NRestarts=0` (no restart at all), `slack` and
`notion` byte-identical and their containers never recycled.

Run date: 2026-08-10 (VM clock UTC). Agent: SESSION A.

---

## Headline finding — the catalog preset was NOT used, and that was the right call

The `linear` catalog preset is a **remote OAuth 2.1 + DCR** entry pointing at
`https://mcp.linear.app/mcp`. On a headless gateway like this VM that flow is
**broken by construction**: Hermes registers `redirect_uri =
http://127.0.0.1:<port>/callback` and listens on the VM's loopback, so an
operator's browser 302s to their *own* laptop and the code never arrives.
Hermes ships a whole skill (`mcp-oauth-remote-gateway`) documenting this exact
failure.

Running `hermes mcp install linear` would therefore have written an `auth: oauth`
stanza that could never complete unattended, on a live production agent.

**Instead I probed whether the same endpoint accepts a static API key — it
does.** `Authorization: Bearer <LINEAR_API_KEY>` returns HTTP 200 and a valid
MCP `initialize`. So the OAuth STOP condition in the directive did **not** apply:
no operator step, no browser, no question file. Hermes' own skill doc agrees
this is the preferred shape:

> **Servers that accept a static Bearer token (API key)** — always prefer
> `headers.Authorization: "Bearer <token>"`.

Note the auth shape is picky: the **raw key** as `Authorization` gets `401
invalid_token`; it must be prefixed with `Bearer `. (The Linear *GraphQL* API is
the opposite — it takes the raw key with no prefix.)

---

## TASK 1 — file GCN-9

### Call-shape validation (sanctioned `curl -K` stdin form)

```
{"data":{"viewer":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain"}}}
```

Viewer id matches the ground-truth Gaetan id.

### issueCreate

Sent `teamId` GCN, `assigneeId` Gaetan, `stateId` Todo (explicit), `labelIds`
`[access]`, `priority: 1`, exact title, description ending in the required
`Spec:` line.

```
{"data":{"issueCreate":{"success":true,"issue":{"id":"2f19a5de-e691-4519-84f7-b084f772283c","identifier":"GCN-9","title":"Wire Linear MCP server natively into Hermes (catalog preset linear)","priority":1,"url":"https://linear.app/mobile-club/issue/GCN-9","state":{"id":"59ed732a-f242-4eba-926f-1c0d128fe83c","name":"Todo"},"assignee":{"name":"Gaëtan Cathelain"},"labels":{"nodes":[{"id":"a4fd96c1-ee2d-4c03-8461-f8a15cca372a","name":"access"}]}}}}}
```

GCN-9 id: `2f19a5de-e691-4519-84f7-b084f772283c`

### issueRelationCreate — relations WERE created

`type: blocks` worked first try, all three in one aliased mutation. No fallback
to description-only was needed.

```
{"data":{"r3":{"success":true,"issueRelation":{"id":"ec585aad-88a2-4eae-9307-6cf127c2d4f3","type":"blocks","issue":{"identifier":"GCN-9"},"relatedIssue":{"identifier":"GCN-3"}}},"r4":{"success":true,"issueRelation":{"id":"403c1a35-194e-4171-98cb-f511a62aa890","type":"blocks","issue":{"identifier":"GCN-9"},"relatedIssue":{"identifier":"GCN-4"}}},"r5":{"success":true,"issueRelation":{"id":"0fa5f7b7-f0c9-4490-8f36-e6a476408ae8","type":"blocks","issue":{"identifier":"GCN-9"},"relatedIssue":{"identifier":"GCN-5"}}}}}
```

The blocking relationship is *also* stated in GCN-9's description ("Blocks:
GCN-3, GCN-4, GCN-5"), so the two agree.

### issueUpdate → In Progress

```
{"data":{"issueUpdate":{"success":true,"issue":{"identifier":"GCN-9","state":{"name":"In Progress"}}}}}
```

### Independent verification — re-read via `team(id:).issues`

Not the mutation responses; a fresh query.

```json
{
    "identifier": "GCN-9",
    "title": "Wire Linear MCP server natively into Hermes (catalog preset linear)",
    "priority": 1,
    "state": { "name": "In Progress" },
    "assignee": { "name": "Gaëtan Cathelain" },
    "labels": { "nodes": [ { "name": "access" } ] },
    "relations": {
        "nodes": [
            { "type": "blocks", "relatedIssue": { "identifier": "GCN-3" } },
            { "type": "blocks", "relatedIssue": { "identifier": "GCN-4" } },
            { "type": "blocks", "relatedIssue": { "identifier": "GCN-5" } }
        ]
    }
}
```

All five attributes confirmed: In Progress, priority 1 (Urgent), assignee
Gaetan, label `access`, blocks GCN-3/4/5.

---

## TASK 2 — install

### CLI facts (verbatim, run before any change)

```
usage: hermes mcp [-h] [--accept-hooks]
                  {serve,add,remove,rm,list,ls,test,configure,config,login,reauth,picker,catalog,install}
    add                 Add an MCP server (discovery-first install)
    test                Test MCP server connection
    catalog             List Nous-approved MCPs available for one-click install
    install             Install a catalog MCP by name (e.g. `hermes mcp install n8n`)
```

```
usage: hermes mcp install [-h] identifier
positional arguments:
  identifier  Catalog entry name (or `official/<name>`)
```

```
usage: hermes mcp add [-h] [--url URL] [--command MCP_COMMAND] [--args ...]
                      [--auth {oauth,header}] [--preset PRESET]
                      [--connect-timeout CONNECT_TIMEOUT] [--env [ENV ...]]
                      name
```

```
  MCP Catalog + configured servers:

  Name               Status                   Description
  ------------------ ------------------------ -----------
  blender            available                Drive a live Blender session — modeling, scenes, and renders.
  comfy-cloud        available                Generate images, video, audio, and 3D on Comfy Cloud.
  figma              available                Official Figma remote MCP …
  linear             available                Find, create, and update Linear issues, projects, and comments.
  n8n                available                Manage and inspect n8n workflows from Hermes …
  unreal-engine      available                Drive the Unreal Engine 5.8 editor over its local MCP server.
  notion             custom — enabled         docker
  slack              custom — enabled         docker
```

### What the catalog install would have DONE

`~/.hermes/hermes-agent/optional-mcps/linear/manifest.yaml`:

```yaml
transport:
  type: http
  url: https://mcp.linear.app/mcp
auth:
  type: oauth
  # No `provider:` — this is native MCP OAuth (case 1) …
  # The MCP client triggers the browser flow on the first probe / first connect.
post_install: |
  On first connection, Hermes will open a browser to authenticate with Linear.
```

Rejected: browser OAuth with a VM-loopback `redirect_uri`, unattendable here.

### Auth probe (read-only, no config change)

```
=== A: raw key as Authorization ===
{"error":"invalid_token","error_description":"Missing or invalid access token"}
[http 401]
=== B: Bearer <key> ===
event: message
data: {"result":{"protocolVersion":"2025-06-18","capabilities":{"tools":{"listChanged":true}},"serverInfo":{"name":"Linear MCP","title":"Linear","version":"1.0.0",…}}}
[http 200]
```

### Why `${LINEAR_API_KEY}` in `headers` is safe (verified in source)

`tools/mcp_tool.py::_interpolate_env_vars` recurses over **dict, list and str**
across the whole server config — not just `env:` — and the loader docstring
states `${ENV_VAR}` placeholders resolve from `os.environ`, "which includes
`~/.hermes/.env` loaded at startup". So the key stays in `.env` and never lands
in `config.yaml`. Same placeholder pattern the existing `notion` entry uses.

The documented config shape for this case, from `mcp_tool.py`:

```yaml
remote_api:
  url: "https://my-mcp-server.example.com/mcp"
  headers:
    Authorization: "Bearer sk-..."
```

### Backup

```
BACKUP=/home/gaetan/.hermes/config.yaml.bak-wf6-20260810T165931Z
-rw------- 1 gaetan gaetan 6441 Aug 10 14:36 /home/gaetan/.hermes/config.yaml.bak-wf6-20260810T165931Z
```

Pre-change baseline: `md5 e2f4d47f7e4e3e0ec1910d7389e03949`, gateway
`NRestarts=0 ActiveState=active`, `mcp list` = slack + notion only.

### The edit — surgical insert under flock, guarded by asserts

Applied with `flock ~/.hermes/.wf3.lock -c '<python>'`. The script refused to
write unless, on the *in-memory* result: exactly one top-level `mcp_servers:`
existed, `linear` was not already present, `slack` and `notion` parsed
**byte-equal to before**, and everything outside `mcp_servers` was unchanged.
Write was atomic (`os.replace`) with mode preserved at 0600.

```
MERGE OK: linear inserted; slack/notion byte-equal; rest of config unchanged
```

### Config diff

```diff
--- /home/gaetan/.hermes/config.yaml.bak-wf6-20260810T165931Z	2026-08-10 14:36:38 +0000
+++ /home/gaetan/.hermes/config.yaml	2026-08-10 16:59:47 +0000
@@ -186,6 +186,10 @@
 context:
   engine: lcm
 mcp_servers:
+  linear:
+    url: https://mcp.linear.app/mcp
+    headers:
+      Authorization: "Bearer ${LINEAR_API_KEY}"
   slack:
     command: docker
     args:
```

Four added lines, nothing else touched. The credential appears only as the
placeholder — there is no secret in `config.yaml` to redact.

---

## TASK 2 VERIFICATION

### 1. `hermes mcp test linear` — GREEN

```
  Testing 'linear'...
  Transport: HTTP → https://mcp.linear.app/mcp
    Authorization: <REDACTED>
  ✓ Connected (4097ms)
  ✓ Tools discovered: 58
```

(Hermes prints a partially-masked header here; fully redacted above.) The fact
it resolved at all proves `${LINEAR_API_KEY}` interpolation worked.

### 2. `mcp__linear__*` visibility — PROVEN, two ways

**(a) Enumeration** via `hermes chat -Q -q "List the exact names of every tool
you can call whose name starts with mcp__linear__ …"` — 58 names returned:

```
mcp__linear__create_attachment            mcp__linear__list_issue_statuses
mcp__linear__create_attachment_from_upload mcp__linear__list_issues
mcp__linear__create_initiative_label      mcp__linear__list_milestones
mcp__linear__create_issue_label           mcp__linear__list_project_labels
mcp__linear__delete_attachment            mcp__linear__list_projects
mcp__linear__delete_comment               mcp__linear__list_release_notes
mcp__linear__delete_diff_comment          mcp__linear__list_release_pipelines
mcp__linear__delete_status_update         mcp__linear__list_releases
mcp__linear__extract_images               mcp__linear__list_teams
mcp__linear__get_agent_skill              mcp__linear__list_users
mcp__linear__get_attachment               mcp__linear__merge_diff
mcp__linear__get_diff                     mcp__linear__prepare_attachment_upload
mcp__linear__get_diff_threads             mcp__linear__resolve_diff_thread
mcp__linear__get_document                 mcp__linear__save_comment
mcp__linear__get_initiative               mcp__linear__save_diff_comment
mcp__linear__get_issue                    mcp__linear__save_document
mcp__linear__get_issue_status             mcp__linear__save_initiative
mcp__linear__get_milestone                mcp__linear__save_issue
mcp__linear__get_project                  mcp__linear__save_milestone
mcp__linear__get_release                  mcp__linear__save_project
mcp__linear__get_release_note             mcp__linear__save_release
mcp__linear__get_status_updates           mcp__linear__save_release_note
mcp__linear__get_team                     mcp__linear__save_status_update
mcp__linear__get_user                     mcp__linear__search_documentation
mcp__linear__get_workspace                mcp__linear__submit_diff_review
mcp__linear__list_agent_skills            mcp__linear__list_documents
mcp__linear__list_comments                mcp__linear__list_initiative_labels
mcp__linear__list_cycles                  mcp__linear__list_initiatives
mcp__linear__list_diffs                   mcp__linear__list_issue_labels
```

**(b) End-to-end functional call** — listing is not proof the tool *works*, so
the agent was told to actually invoke one:

```
$ hermes chat -Q -q "Use the mcp__linear__get_issue tool to fetch issue GCN-9. …"
session_id: 20260810_170104_8ee90f
{
  "identifier": "GCN-9",
  "title": "Wire Linear MCP server natively into Hermes (catalog preset linear)",
  "state name": "In Progress",
  "priority": "Urgent",
  "assignee name": "Gaëtan Cathelain",
  "label names": ["access"]
}
```

The agent loop read live Linear data through the MCP surface. Closed loop:
GCN-9 was filed by raw HTTP and read back natively.

### 3. Gateway — NO regression

```
=== FINAL gateway ===
NRestarts=0
ActiveState=active
SubState=running
ActiveEnterTimestamp=Mon 2026-08-10 14:37:08 UTC
```

`NRestarts=0` and `ActiveEnterTimestamp` is **identical to the pre-change
baseline** — the gateway did not restart, it live-reloaded. No rollback needed.

### 4. slack / notion survived

```
mcp_servers keys: ['linear', 'notion', 'slack']
```

```
  MCP Servers:
  Name             Transport                      Tools        Status
  ──────────────── ────────────────────────────── ──────────── ──────────
  linear           https://mcp.linear.app/mcp     all          ✓ enabled
  slack            docker run -i                  all          ✓ enabled
  notion           docker run -i                  all          ✓ enabled
```

Stronger than a config check — the same container PIDs are still alive post-change:

```
├─677174 /usr/bin/docker run -i --rm --env-file /home/gaetan/tars/slack-mcp/.env ghcr.io/korotovsky/slac...
├─677176 /usr/bin/docker run -i --rm -e NOTION_TOKEN mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f28838...
```

PIDs 677174 / 677176 were the same before and after, so neither server was even
recycled. The merge script additionally asserted both stanzas parsed byte-equal.

---

## Tool schemas for the next tickets

Naming: **`save_issue` is both create and update** — omit `id` to create, pass
`id` to update. There is no separate `create_issue`.

### `mcp__linear__save_issue` — required: *none* (title + team required on create)

Resolvers accept **names or IDs**, so the GCN-3/4/5 work does not need to carry
UUIDs around.

| param | type | notes |
|---|---|---|
| `id` | string | Only to update. Issue ID or identifier (e.g. `GCN-9`) |
| `title` | string | **required when creating** |
| `description` | string | Markdown. Literal newlines, do not escape |
| `patch` | array | Partial atomic edits to existing content |
| `team` | string | **required when creating** — team *name or ID* |
| `state` | string | State **type, name, or ID** (so `"Todo"` works) |
| `priority` | number | `0=None, 1=Urgent, 2=High, 3=Medium, 4=Low` |
| `assignee` | string\|null | User ID, name, email, or `"me"`; null to remove |
| `labels` | array | Label **names or IDs**; **replaces the full label set** |
| `project` | string\|null | Name, ID, identifier, or slug |
| `cycle` | string\|null | Name, number, or ID |
| `milestone` | string | Name or ID |
| `parentId` | string\|null | Parent issue ID/identifier |
| `estimate` | number\|null | |
| `dueDate` | string\|null | ISO |
| `blocks` | array | Issue IDs/identifiers this blocks — **append-only** |
| `blockedBy` | array | Issues blocking this — append-only |
| `relatedTo` | array | append-only |
| `duplicateOf` | string\|null | |
| `removeBlocks` / `removeBlockedBy` / `removeRelatedTo` | array | inverse ops |
| `links` | array | `[{url, title}]`, append-only |
| `slaBreachesAt`, `slaType`, `delegate`, `setReleases`, `addReleases`, `removeReleases` | | |

**Answering the coordinator's question directly: yes — `save_issue` exposes
`team`, `state`, `labels` and `priority` as first-class params, plus `assignee`
and even the `blocks` relations that needed a separate `issueRelationCreate`
mutation over raw GraphQL.** The entire GCN-9 filing (issue + 3 relations +
state move), which took three GraphQL round-trips, is one `save_issue` call.

Two gotchas for GCN-3: `labels` **replaces** the whole set (read-modify-write if
preserving), while `blocks`/`blockedBy` are **append-only**.

### Read tools

```
## get_issue            | required: ['id']
   params: id, includeCustomerNeeds, includeRelations, includeReleases
## list_issues          | required: none
   params: assignee, createdAt, cursor, cycle, delegate, fields, includeArchived,
           label, limit, orderBy, parentId, priority, project, query, release,
           state, team, updatedAt
## list_issue_statuses  | required: ['team']
## list_issue_labels    | required: none — cursor, limit, name, orderBy, team
## list_teams           | required: none — createdAt, cursor, includeArchived,
                          limit, orderBy, query, updatedAt
## save_comment         | required: ['body']
   params: body, documentId, id, initiativeId, issueId, milestoneId, parentId,
           projectId, statusUpdateId, statusUpdateType
```

`list_issues` covers the GCN-4 board pull directly: filter by `team` + `state` +
`assignee`, sort with `orderBy`, and `fields` trims the payload.
`save_comment` with `issueId` covers the GCN-3 push side.

Server `instructions` returned at initialize, which Hermes injects into the
agent context:

> When passing string values to tools, send the content directly without escape
> sequences. For example, use real newlines in markdown content rather than
> literal backslash-n (\n) characters.

---

## Notes / residual risk

- **Token lifetime**: this is a static Personal API key, not OAuth, so there is
  no refresh and no expiry-driven breakage — but also no scope reduction. The
  key is whatever GCN-2 provisioned; the MCP surface can do anything the key can
  (including `merge_diff` and `delete_*` tools).
- The 58-tool surface is large and entirely enabled. `hermes mcp configure
  linear` can prune it if the tool budget becomes a problem.
- No secret was written to disk, echoed, placed on argv, or included in this
  file. All API calls used the sanctioned `curl -K -` stdin form.
- No git command was run and no other repo file was touched.
