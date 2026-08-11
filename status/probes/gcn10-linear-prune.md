# GCN-10 — prune the Linear MCP tool surface exposed to Tars (2026-08-11)

Applied live on the Tars VM (`192.168.0.9`) on 2026-08-11 ~16:58 UTC. The
`mcp_servers.linear` stanza in `~/.hermes/config.yaml` gained a
`tools.include` whitelist of **10 bare tool names**; the other **48** of the
58 tools Linear's MCP server offers are no longer registered into the agent's
toolset. Additive 12-line diff, `.bak` taken under the same `flock`, no
`sops`, no restart, no Slack message. **Verdict: applied and enforced.**
The registration-level probe enumerates exactly the 10 kept names, the
`tools.mcp_tool` log line independently reads `registered 10 tool(s)` (was 58
on every prior boot), a kept tool still works (`list_issues` → GCN-24 /
GCN-10 / GCN-12), and both `list_projects` and `merge_diff` are gone.
`hermes mcp test linear` still reads 58 — that is expected and **inconclusive,
not a failure**; see §7 before touching anything.

## 1. Mechanism (DECISION 1)

The per-tool filter is **`mcp_servers.<name>.tools.include`** (whitelist) and
**`mcp_servers.<name>.tools.exclude`** (blacklist), exact names or fnmatch
globs, nested under the existing per-server block in `~/.hermes/config.yaml`
as a sibling of `url:` / `headers:`.

The enforcement point is not a UI convention — it is the runtime registration
loop in `tools/mcp_tool.py` on the VM (`~/.hermes/hermes-agent`, Hermes Agent
v0.20.0), verbatim:

```python
    # Selective tool loading: honour include/exclude lists from config.
    # Rules (matching issue #690 spec, extended with glob support):
    #   tools.include — whitelist: only matching tool names are registered
    #   tools.exclude — blacklist: all tools EXCEPT matching ones are registered
    #   entries may be exact names or fnmatch globs (e.g. "*_radar_*")
    #   include takes precedence over exclude
    #   Neither set → register all tools (backward-compatible default)
    tools_filter = config.get("tools") or {}
    include_set = _normalize_name_filter(
        tools_filter.get("include"), f"mcp_servers.{name}.tools.include"
    )
    exclude_set = _normalize_name_filter(
        tools_filter.get("exclude"), f"mcp_servers.{name}.tools.exclude"
    )
    def _should_register(tool_name: str) -> bool:
        if include_set:
            return matches_name_filter(tool_name, include_set)
        if exclude_set:
            return not matches_name_filter(tool_name, exclude_set)
        return True
```

The same `f"mcp_servers.{name}.tools.include"` / `.exclude"` key pair is built
3× in `tools/mcp_tool.py` (registration and the rebuild-on-reload path), so it
governs both the first load and every live reload.

**The ticket's guessed mechanism was wrong.** GCN-10's body and its parent
record (`status/probes/wf6/gcn9-native-mcp.md:433-435`) name
`hermes mcp configure linear` as "the prune mechanism". Measured on the VM:

```
$ ~/.local/bin/hermes mcp configure --help
usage: hermes mcp configure [-h] name
positional arguments:
  name        Server name to configure
options:
  -h, --help  show this help message and exit
```

No headless mode, no tool arguments — it is a **curses checklist UI** whose
docstring says it "Writes changes back as ``tools.exclude`` entries in
config.yaml" (in fact the current code writes `include`). It is a *writer* for
the config key above, not the mechanism itself, and it cannot be driven over
non-interactive ssh at all. The mechanism is the config key; the CLI is one of
several ways to type it.

**Names are BARE.** `_should_register(tool_name)` is fed the name the server
reports (`get_issue`, `save_issue`, …) — the server name is already the config
key, and the `mcp__linear__` prefix is a client-side display/dispatch
convention. Writing `mcp__linear__save_issue` leaves `include_set` non-empty
with **zero matches → zero Linear tools register, silently**, killing
engagement-checker (`*/30` cron), daily-work-brief (`30 8` cron),
helptech-duty-intake and linear-ticketing, and presenting to Gaetan as "I
don't have that tool" rather than as an error. Step 0 (§6) exists to turn that
from inference into evidence *before* the config is touched.

Related but distinct: `agent.disabled_toolsets` is a coarser flat list that
kills a whole toolset/server. Per-tool filtering is exclusively the
`mcp_servers.<name>.tools.{include,exclude}` shape.

## 2. Why `include`, not `exclude`

**Fail-closed.** `_should_register` returns `matches_name_filter(tool_name,
include_set)` — a tool Linear ships *tomorrow* is not in the list and is not
registered. Under `exclude`, every new Linear tool auto-mounts on Tars. That
is not hypothetical: this surface already grew a diff-review family
(`merge_diff`, `submit_diff_review`, `resolve_diff_thread`, …), which is
exactly the category GCN-10 exists to remove. An auto-opening blacklist
re-contradicts **SOUL rule 2** ("I never merge, approve or push") on Linear's
release schedule rather than on Gaetan's.

**The codebase says so itself**, in a comment written for this exact question
(`hermes_cli/tools_config.py`): *"We standardize on tools.include across the
codebase (catalog installs, hermes mcp configure, and this UI) so a server's
on-disk config shape doesn't depend on which UI the user touched last."* A
48-entry `exclude` list would also be rewritten to `include` the first time
anyone saves the curses screen.

**Audit size**: 10 lines a human reads in a diff, versus 48 lines where one
missing entry is an invisible fail-open.

Accepted cost: `include` fails closed on a **rename** too — if Linear renames
a kept tool, it silently disappears. The `*/30` engagement-checker cron is the
free regression watch (see §8).

No glob risk: none of the 10 names contains an fnmatch metacharacter
(`*?[]`), and a metacharacter-free fnmatch pattern is an **exact** match — so
`get_issue` does not also admit `get_issue_status`. 10 entries → 10 tools.
Do **not** "simplify" to globs: `save_*` would re-admit `save_project`,
`save_release`, `save_release_note`, `save_document`, `save_initiative`,
`save_milestone`, `save_status_update` and `save_diff_comment`.

## 3. Decision rule for the table

KEEP iff the tool (a) performs a create / read / search / update / close /
cancel / comment / delete on an **issue or a comment**, or (b) is the
resolution read that makes such a mutation *confirmable* under
`skills/orchestration/linear-ticketing/SKILL.md` §10, which measured that a
`save_issue` with an unresolvable `state` "returns a full, normal,
`result`-bearing payload, changes nothing, and does not advance `updatedAt`"
(`:170`) and that a write is confirmed only by asserting the **returned
representation** (`statusType`, `assigneeId`, label *names*, `teamId`) against
the ids that were sent (`:172-180`). That clause is what earns the four
metadata reads their place — they are the only way to obtain the ids a
lifecycle write must be checked against.

PRUNE everything else, per Gaetan's bar "prune only tools with no lifecycle
use", plus a **separate and stricter** reason: any singleton by-ID lookup
whose ID can only come from a kept `list_*` that already returns the full
record adds **zero capability** — a harder disqualification than "no call
site", and one that does not engage the keep-when-torn tie-break.

## 4. Full tool table — all 58

### KEEP — 10 (the `tools.include` list, verbatim and in file order)

| # | tool | rationale |
|---|---|---|
| 1 | `save_issue` | The entire issue lifecycle: create, edit, assign, relate, close/cancel via `state`. 18 references; live call sites `skills/orchestration/engagement-checker/SKILL.md:533,606`, `skills/helptech-duty-intake/SKILL.md:118`, `skills/orchestration/linear-ticketing/SKILL.md:63,75,107,225`. Non-negotiable. |
| 2 | `get_issue` | Only read path for the full issue record incl. `attachments` — helptech's Slack-Ask provenance gate (`skills/helptech-duty-intake/SKILL.md:102-109`) and engagement-checker's post-create confirmation (`:608,656`). Measured live 2026-08-11 on MC-4228 and GCN-24. |
| 3 | `list_issues` | Dedupe scan before every create (`engagement-checker/SKILL.md:590`), helptech's eligible-ticket scan (`:82`), daily-work-brief's completed count (`daily-work-brief/SKILL.md:98`). 18 references. |
| 4 | `list_issue_statuses` | Resolves a team's valid state names/ids. §10 measured that an unresolvable `state` **silently no-ops** and cannot be pre-empted by validation (`linear-ticketing/SKILL.md:170`) — resolving the state before a close is lifecycle-critical, not convenience. |
| 5 | `list_issue_labels` | `save_issue.labels` **REPLACES the full set** (`linear-ticketing/SKILL.md:124`) and §10 asserts the returned label *names*; a stale name silently wipes labels. A read of the valid set is a safety on a mutation path. |
| 6 | `list_teams` | `team` is required on create and takes a name or ID (`:118`); §10's create assertion is `teamId` matching (`:176`). Name → UUID is the only way to confirm a create landed on the intended team — the GCN-vs-company gate. |
| 7 | `list_users` | `save_issue.assignee` takes a name/email, but §10 states "`assigneeId` equals it; `assignee` is a display name and **never matches**" (`:175`) — an assign by name is otherwise *unconfirmable*. Name/email → UUID is what makes an assignment write verifiable. |
| 8 | `save_comment` | Comment lifecycle (create + update). Gaetan: "the same for comments". `linear-ticketing/SKILL.md:210-211` requires a provenance comment and an outcome comment on **every** delegated ticket; a live call is recorded at `status/probes/wf6/gcn12-reopened.md:94`. |
| 9 | `list_comments` | Read comments on an issue. In linear-ticketing's blessed closed list (`:91`); the read half of the comment lifecycle Gaetan kept by name. |
| 10 | `delete_comment` | Comment lifecycle: delete. Gaetan keeps delete for comments explicitly. |

### PRUNE — 48, grouped by family

**Diff review (8)** — contradicts SOUL rule 2; wholly unused. Named by name in
GCN-10's own ticket body as the concern.

| # | tool | rationale |
|---|---|---|
| 11 | `merge_diff` | Merges a Linear diff. SOUL rule 2: "I never merge, approve or push" — the tool surface must stop contradicting the soul. Named by Gaetan. |
| 12 | `submit_diff_review` | Approves / requests changes on a diff. Same SOUL rule 2 contradiction. Named by Gaetan. |
| 13 | `resolve_diff_thread` | Resolves/reopens a diff comment thread — code-review authority Tars must not hold. |
| 14 | `save_diff_comment` | Comments on a **diff**, not on an issue; outside the comment lifecycle Gaetan kept. |
| 15 | `delete_diff_comment` | Deletes a diff comment; same — not an issue comment. |
| 16 | `get_diff` | Read of Linear's PR/diff surface. No issue or comment lifecycle use. |
| 17 | `get_diff_threads` | Same, thread-level read. No lifecycle use. |
| 18 | `list_diffs` | Same, listing. No lifecycle use. |

**Releases (7)** — release management, not issue lifecycle.

| # | tool | rationale |
|---|---|---|
| 19 | `get_release` | Release read. Issue↔release linkage lives on `save_issue` (`setReleases`/`addReleases`/`removeReleases`, `linear-ticketing/SKILL.md:135`), not here. |
| 20 | `list_releases` | Same; no skill reads releases. |
| 21 | `save_release` | Release write — workspace-admin surface named in the governing decision. |
| 22 | `get_release_note` | Release-note read. No lifecycle use. |
| 23 | `list_release_notes` | Same. No lifecycle use. |
| 24 | `save_release_note` | Release-note write — admin surface. |
| 25 | `list_release_pipelines` | Pipeline read. No lifecycle use, no skill reference. |

**Attachments (6)** — write path is `save_issue.links`, read path is `get_issue`.

| # | tool | rationale |
|---|---|---|
| 26 | `create_attachment` | Deprecated base64 fallback. Attachments Tars needs are minted by `save_issue.links` (`[{url,title}]`) and read back by `get_issue` — measured live 2026-08-11: MC-4228 → `{title:"Ask from Apolline", url:".../archives/CC397R0HY/..."}`, GCN-24 → `{title:"Source email — …", url:"https://mail.google.com/…"}`. |
| 27 | `create_attachment_from_upload` | Links an uploaded asset to an issue. **Genuine capability loss: file upload to an issue goes away** (`links` is URL+title only). No skill uploads files; accepted. |
| 28 | `prepare_attachment_upload` | Other half of that upload flow; same accepted loss. |
| 29 | `delete_attachment` | Destructive, zero call sites, and would let Tars destroy the provenance record `helptech-duty-intake` §5 gates every company-team write on. |
| 30 | `get_attachment` | Fetches attachment **content** by ID; the ID can only come from `get_issue`, whose array already carries `title` and `url` — the only fields any step tests. Adds nothing even in principle. |
| 31 | `extract_images` | Image extraction from markdown. No lifecycle use, no skill reference. |

**Initiatives (5)** — workspace-admin surface named in the governing decision.

| # | tool | rationale |
|---|---|---|
| 32 | `get_initiative` | Initiative read. No issue/comment lifecycle use. |
| 33 | `list_initiatives` | Same. No skill reference. |
| 34 | `save_initiative` | Initiative write — admin surface Gaetan named. |
| 35 | `list_initiative_labels` | Initiative label read. No lifecycle use. |
| 36 | `create_initiative_label` | Label **creation** = admin; already forbidden in prose (`linear-ticketing/SKILL.md:91`). |

**Projects (4)** — workspace-admin surface named in the governing decision.

| # | tool | rationale |
|---|---|---|
| 37 | `get_project` | Project read. `save_issue.project` takes "name, ID, identifier, or slug" (`:128`), so setting an issue's project survives without it. |
| 38 | `list_projects` | Same. Removes *value discovery* for `save_issue.project` and the `project` filter on `list_issues` (`:141`); no current use. |
| 39 | `save_project` | Project write — admin surface Gaetan named. |
| 40 | `list_project_labels` | Project label read. No issue lifecycle use. |

**Milestones (3)** — named in the governing decision.

| # | tool | rationale |
|---|---|---|
| 41 | `get_milestone` | Milestone read. No skill sets `save_issue.milestone` (grep over `skills/`). |
| 42 | `list_milestones` | Same. No caller. |
| 43 | `save_milestone` | Milestone write — admin surface Gaetan named. |

**Documents (3)** — named in the governing decision.

| # | tool | rationale |
|---|---|---|
| 44 | `get_document` | Document read. Not an issue, not a comment. |
| 45 | `list_documents` | Same. No skill reference. |
| 46 | `save_document` | Document write — admin surface; `linear-ticketing/SKILL.md:91` already forbids document writes in prose. |

**Status updates (3)** — project/initiative updates, not issue state.

| # | tool | rationale |
|---|---|---|
| 47 | `get_status_updates` | Project/initiative status updates. Issue state lives in `save_issue.state`. |
| 48 | `save_status_update` | Write to a project/initiative status update — admin surface. |
| 49 | `delete_status_update` | Destructive admin write, zero lifecycle use. |

**Singleton by-ID lookups redundant with a kept `list_*` (3)** — reclassified
from KEEP during the adversarial review; zero capability added, which is a
stricter disqualification than "no call site".

| # | tool | rationale |
|---|---|---|
| 50 | `get_issue_status` | Needs a status ID only `list_issue_statuses` (KEPT) can supply, and that call already returns every status with its name and type. Strict subset. Zero call sites; engagement-checker hardcodes the GCN state ids (`engagement-checker/SKILL.md:533`). |
| 51 | `get_team` | Needs a team ID/name only `list_teams` (KEPT) can supply. The GCN-vs-company gate runs off `get_issue`'s `id`/`teamId` — "there is no team-key field; the prefix is only obtainable by parsing `id`" (`linear-ticketing/SKILL.md:186`). Strict subset. |
| 52 | `get_user` | Needs a user ID only `list_users` (KEPT) can supply, which already returns the workspace's users. The draft's justification ("helptech assigns") is false: helptech passes a hardcoded UUID (`helptech-duty-intake/SKILL.md:121`), daily-work-brief filters on the same literal (`daily-work-brief/SKILL.md:100`). Strict subset. |

**Agent skills (2)**

| # | tool | rationale |
|---|---|---|
| 53 | `get_agent_skill` | Linear's agent-skill registry. Unrelated to Tars's own `skill_manage`. No lifecycle use. |
| 54 | `list_agent_skills` | Same. No skill reference. |

**Singles (4)**

| # | tool | rationale |
|---|---|---|
| 55 | `create_issue_label` | Label **creation** is admin; `linear-ticketing/SKILL.md:91` forbids it in prose ("no label/team/state creation") — this makes it enforcement. Reading labels stays. |
| 56 | `list_cycles` | GCN is a personal team with no cycle workflow; no skill sets `save_issue.cycle` or filters by cycle. Removes value discovery for those params; no current use. |
| 57 | `get_workspace` | Workspace metadata. No issue/comment lifecycle use. |
| 58 | `search_documentation` | Searches Linear's own **product docs**. No lifecycle use whatsoever. |

**Arithmetic: 10 keep + 48 prune = 58.** `set(keep) ∩ set(prune)` = ∅;
`set(keep) ∪ set(prune)` equals the canonical 58 exactly, which itself equals
the live `hermes mcp test linear` output name-for-name and the earlier
`status/probes/wf6/gcn9-native-mcp.md:263-291` enumeration. Checked
programmatically, not by eye — no invented, misspelled or duplicated name.
That is the check that matters most under `tools.include`, where a typo
silently drops a tool instead of erroring.

## 5. Before / after config diff

Additive only. `diff ~/.hermes/config.yaml.bak-gcn10-20260811 ~/.hermes/config.yaml`:

```
170a171,182
>     tools:
>       include:
>       - save_issue
>       - get_issue
>       - list_issues
>       - list_issue_statuses
>       - list_issue_labels
>       - list_teams
>       - list_users
>       - save_comment
>       - list_comments
>       - delete_comment
```

`diff ... | grep -c '^<'` → `0`. Nothing removed; the `slack:` and `notion:`
stanzas are byte-identical.

Resulting `mcp_servers.linear` stanza (verbatim — the file holds only an
env-var reference, no literal token anywhere in it):

```yaml
mcp_servers:
  linear:
    url: https://mcp.linear.app/mcp
    headers:
      Authorization: "Bearer ${LINEAR_API_KEY}"
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
  slack:
```

File facts:

```
478ab1721c3169b100b6610caa3ba3f4  /home/gaetan/.hermes/config.yaml
debdf1f36c85208b314897ff9697792d  /home/gaetan/.hermes/config.yaml.bak-gcn10-20260811
  223 /home/gaetan/.hermes/config.yaml
  211 /home/gaetan/.hermes/config.yaml.bak-gcn10-20260811
-rw------- 1 gaetan gaetan 6616 Aug 11 16:58 /home/gaetan/.hermes/config.yaml
-rw------- 1 gaetan gaetan 6375 Aug 10 19:59 /home/gaetan/.hermes/config.yaml.bak-gcn10-20260811
```

211-line `debdf1f36c85208b314897ff9697792d` → 223-line
`478ab1721c3169b100b6610caa3ba3f4`, mode `-rw-------` unchanged on both.
Rollback target: `~/.hermes/config.yaml.bak-gcn10-20260811`, whose md5 equals
the measured baseline exactly — the backup is known-good, not merely present.

How it was written (why the diff is this quiet): backup + patch inside **one**
`flock ~/.hermes/.wf3.lock`; the writer asserted the baseline md5 before
touching anything (a sibling agent's `mcp_servers.slack` write once landed
between another agent's backup and its edit,
`status/probes/wf3-s2-linear-notion-calendar.md:139-142`), asserted anchor
uniqueness and "no existing `tools:` block", spliced text rather than
re-serialising YAML (ruamel measurably reflowed this same file on 2026-08-10,
`status/probes/wf6/slack-tool-visibility-restore.md:207-212`), parsed YAML only
as a **post-condition** (include == KEEP; `url`/`headers` unchanged; everything
outside `mcp_servers.linear.tools` byte-equal after deleting the new key),
asserted `len(out) == len(src) + 12`, and wrote via `os.replace` (atomic, mode
0600) so a live reload can never read a torn file.

## 6. Probes

### 6a. Step 0 — dry run of the filter itself, run BEFORE any write

Hermes' own matcher, Hermes' own interpreter, against the live 58 names on the
VM (`cd ~/.hermes/hermes-agent && .venv/bin/python /tmp/gcn10_dryrun.py`):

```
MATCHER=hermes
TOTAL 58 KEPT 10
KEPT_SET_OK True
DROPPED 48
DRYRUN_OK
```

`MATCHER=hermes` means `tools.mcp_tool._normalize_name_filter` +
`matches_name_filter` imported successfully — the real matcher was exercised,
not the fnmatch fallback. The generator that built `/tmp/gcn10_all58.txt`
asserted 58 unique `[a-z_]+` names with all 10 KEEP names present
(`LOCAL_OK 58 keep_all_present`). Stop condition was: no `DRYRUN_OK`, no edit.

### 6b. `hermes mcp test linear` (tier 1 — server discovery)

```
  Testing 'linear'...
  Transport: HTTP → https://mcp.linear.app/mcp
    Authorization: <REDACTED>
  ✓ Connected (2041ms)
  ✓ Tools discovered: 58
    get_attachment                       Retrieve an attachment's content by ID.
    prepare_attachment_upload            Prepare a direct Linear file upload for an existing iss...
    …
    delete_status_update                 Delete (archive) a project or initiative status update.
```

(The elided lines are the remaining 55 `name  description` rows; the 58 names
printed are exactly the canonical set tabulated in §4, name-for-name, in the
server's own order.) **Reads 58 → INCONCLUSIVE by design, not a failure.**
See §7.

### 6c. Registration-level enumeration (tier 2 — authoritative)

```
ssh gaetan@192.168.0.9 "~/.local/bin/hermes chat -Q -q 'List the exact names of every tool you can call whose name starts with mcp__linear__. Output only the bare names, one per line, nothing else.'"
```
```
session_id: 20260811_165943_e1791a
mcp__linear__delete_comment
mcp__linear__get_issue
mcp__linear__list_comments
mcp__linear__list_issue_labels
mcp__linear__list_issue_statuses
mcp__linear__list_issues
mcp__linear__list_teams
mcp__linear__list_users
mcp__linear__save_comment
mcp__linear__save_issue
```

Exactly the 10 kept names. Absent: `merge_diff`, `submit_diff_review`,
`list_projects`, `save_project`, `delete_attachment` and the other 43. This is
the same enumeration path that produced the original 58 in
`status/probes/wf6/gcn9-native-mcp.md:263-291`.

Independently-verified corroboration from `~/.hermes/logs/agent.log`, same
code path the gateway uses:

```
2026-08-11 17:01:03,725 INFO tools.mcp_tool: MCP server 'linear' (HTTP): registered 10 tool(s): mcp__linear__list_comments, mcp__linear__save_comment, mcp__linear__delete_comment, mcp__linear__get_issue, mcp__linear__list_issues, mcp__linear__save_issue, mcp__linear__list_issue_statuses, mcp__linear__list_issue_labels, mcp__linear__list_teams, mcp__linear__list_users
```

Before/after on that same log line:

```
2026-08-10 22:28:48 ... 'linear' (HTTP): registered 58 tool(s)
2026-08-11 07:45:05 ... 'linear' (HTTP): registered 58 tool(s)
2026-08-11 16:59:45 ... 'linear' (HTTP): registered 10 tool(s)
2026-08-11 17:00:19 ... 'linear' (HTTP): registered 10 tool(s)
2026-08-11 17:01:03 ... 'linear' (HTTP): registered 10 tool(s)
```

Siblings unchanged in the same registration pass: `notion` 24 tools, `slack`
20 tools, and the totals line reads `MCP: registered 54 tool(s) from 3
server(s)` (10 + 24 + 20). The blast radius of the edit is exactly the Linear
server.

### 6d. POSITIVE probe — a kept tool still works (read-only)

```
ssh gaetan@192.168.0.9 "~/.local/bin/hermes chat -Q -q 'Using mcp__linear__list_issues, list the issues on team GCN with state Todo. Report just their identifiers and titles.'"
```
```
session_id: 20260811_170006_b18147
GCN-24 — Review Yacine's Verdict staging Terraform PR (#65)
GCN-10 — Prune the Linear MCP tool surface exposed to Tars
GCN-12 — Create and send the Verdict Google OAuth client
```

Real GCN identifiers; `agent.log` records `tool mcp__linear__list_issues
completed` at 17:00:44. Nothing was created, modified, closed or deleted.

### 6e. NEGATIVE probe — pruned tools are gone

```
ssh gaetan@192.168.0.9 "~/.local/bin/hermes chat -Q -q 'Do you have a tool named mcp__linear__list_projects? Answer yes or no, then do the same for mcp__linear__merge_diff. Do not call them.'"
```
```
session_id: 20260811_170057_96e9a0
mcp__linear__list_projects: No
mcp__linear__merge_diff: No
```

### 6f. NO-RESTART proof

```
NRestarts=0
ActiveState=active
SubState=running
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

`ActiveEnterTimestamp` is byte-identical to the pre-change baseline — the
gateway was never bounced; the change went live on Hermes' own ~30 s reload,
as GCN-9 established for `mcp_servers` edits
(`status/probes/wf6/gcn9-audit.md`, Claim 3). `NRestarts=0` is recorded as a
secondary datum only; it is known to miss restarts and is not cited as proof.

### 6g. Logs

`tail -30 ~/.hermes/logs/errors.log` in the post-change window:

```
2026-08-11 16:51:43,933 WARNING [20260811_165128_cdf37061] agent.tool_executor: Tool kanban_show returned error (0.00...
2026-08-11 16:59:46,368 WARNING tools.registry: check_fn check_bfl_requirements returned False; dependent tools will ...
2026-08-11 16:59:46,685 WARNING tools.registry: check_fn check_computer_use_requirements returned False; dependent to...
2026-08-11 16:59:46,691 WARNING tools.registry: check_fn check_image_generation_requirements returned False; dependen...
2026-08-11 16:59:46,800 WARNING tools.registry: check_fn check_web_api_key returned False; dependent tools will be un...
2026-08-11 16:59:52,523 WARNING [20260811_165943_e1791a] agent.tool_executor: Tool kanban_show returned error (0.00s)...
2026-08-11 17:00:20,252 WARNING tools.registry: check_fn check_bfl_requirements returned False; dependent tools will ...
2026-08-11 17:00:25,479 WARNING [20260811_170006_b18147] agent.tool_executor: Tool kanban_show returned error (0.00s)...
2026-08-11 17:01:04,055 WARNING tools.registry: check_fn check_bfl_requirements returned False; dependent tools will ...
2026-08-11 17:01:04,274 WARNING tools.registry: check_fn check_web_api_key returned False; dependent tools will be un...
```
```
$ grep -iE "linear|mcp" ~/.hermes/logs/errors.log | grep -E "2026-08-11 1[67]:"
(no output)
```

Zero MCP/Linear errors introduced. The `check_fn …` and `kanban_show`
warnings are **pre-existing and unrelated**: `check_bfl_requirements` alone
occurs 203 times in the file, earliest 2026-08-07 16:46; they fire once per
agent construction, including the three probe sessions above.

### 6h. Verdict table

| probe | verdict | note |
|---|---|---|
| step-0 dry run | **PASS** | `MATCHER=hermes`, `TOTAL 58 KEPT 10`, `KEPT_SET_OK True`, `DRYRUN_OK` — run before any write |
| baseline | **PASS** | md5 `debdf1f36c85208b314897ff9697792d`, 211 lines |
| patch | **PASS** | `PATCH_OK 478ab1721c3169b100b6610caa3ba3f4 223` |
| diff | **PASS** | one `170a171,182` hunk, 12 added lines, `grep -c '^<'` = 0 |
| md5 / wc / ls | **PASS** | 223 lines, mode `-rw-------` unchanged, `.bak` md5 == baseline |
| `mcp test linear` (tier 1) | **INCONCLUSIVE** (expected) | reports 58; discovery-level, blind to `tools.include`. Not a failure |
| registration enumeration (tier 2) | **PASS** | exactly the 10 kept names; corroborated by `registered 10 tool(s)` in `agent.log` (was 58) |
| positive probe | **PASS** | `list_issues` → GCN-24 / GCN-10 / GCN-12 |
| negative probe | **PASS** | "No" for both `list_projects` and `merge_diff` |
| no-restart | **PASS** | `ActiveEnterTimestamp` unchanged, `active/running` |
| logs | **PASS** | zero MCP/Linear errors post-change |

## 7. The two-tier acceptance test — read this before acting on a "58"

`hermes mcp test linear` connects to the MCP server and enumerates **what the
server offers**. There is no evidence, in code or in this repo's precedent,
that it applies `tools.include` before printing — no server on this VM had
ever carried a `tools:` block before today, so no precedent existed either
way, and after the change it still prints 58. `_should_register` runs at
**tool-registration time inside the agent**, which is a different layer.

So the acceptance test has two tiers:

- **Tier 1 — `hermes mcp test linear` (discovery).** `10` would confirm at
  both layers. `58` is **INCONCLUSIVE**, not failure.
- **Tier 2 — registration (authoritative).** What Tars can actually call:
  `hermes chat -Q -q` enumeration (§6c), corroborated by the
  `tools.mcp_tool: … registered N tool(s)` line in `~/.hermes/logs/agent.log`.

This run resolved at tier 2: tier 1 read 58 exactly as the pre-implementation
audit predicted, tier 2 read 10.

**A future operator who reads `Tools discovered: 58` and concludes the prune
failed must NOT roll back and must NOT restart the gateway.** Fall through to
tier 2 first. A restart on a misread tier-1 is the expensive mistake here: it
bounces the live Slack agent for a state that was never wrong. This is the
single most reusable finding in this file.

## 8. Scope caveats — what this does NOT bound

1. **`LINEAR_API_KEY` remains unscoped full-write.** This prune shrinks what
   the *model can reach*, not what the *credential can do*. Anything holding
   the key — including anything that reads it off the VM — still has the full
   Linear API. Gate G2 of `docs/specs/wf6-linear-integration.md` is unchanged.
2. **Two raw-HTTP Linear paths are untouched by MCP tool scoping, not one.**
   `skills/orchestration/engagement-checker/SKILL.md:155-480` §4's urllib
   GraphQL collector (with its own mutation-rejecting allowlist) **and**
   `scripts/linear_board.py`, run through the terminal tool for
   daily-work-brief's board and explicitly exempted from the native-tool rule
   (`skills/orchestration/daily-work-brief/SKILL.md:88-94`). Both are
   read-only — that is the mitigating fact, and it is stated rather than
   implied.
3. **The claude.ai connector (`mcp__claude_ai_Linear__*`) is a different
   server**, used from cooper Claude Code sessions and authenticating as
   Gaetan via claude.ai. It lives outside Tars's `config.yaml` entirely — out
   of scope and not prunable here.
4. **There is no `delete_issue` tool on this surface at all.** Verified
   against the canonical 58: the only `delete_*` tools are
   `delete_attachment`, `delete_comment`, `delete_diff_comment`,
   `delete_status_update`. Gaetan's "must remain able to close and delete
   tickets" therefore maps to `save_issue` (`state` → Canceled/Done) for the
   issue and `delete_comment` for comments — both kept. **Nothing capable of
   deleting an issue was pruned, because none is offered.**
5. **The kept comment tools still reach non-issue parents.** `save_comment`
   takes `documentId`, `initiativeId`, `projectId`, `milestoneId`,
   `statusUpdateId` (`skills/orchestration/linear-ticketing/SKILL.md:158`),
   `list_comments` lists comments on "a Linear issue, project, initiative,
   d…", and `delete_comment` deletes any of them. So "workspace-admin surface
   removed" is **not literally true** — comment-shaped access to
   projects/initiatives/milestones/documents/status updates survives. That
   does not violate the directive (Gaetan kept the comment lifecycle whole),
   but the claim must be phrased honestly.
6. **`include` is fail-closed, so the list needs revisiting on server
   upgrades.** Any tool Linear adds later is auto-blocked — deliberate — but a
   Linear MCP upgrade that *renames* a kept tool silently removes it. The
   `*/30` engagement-checker cron is the free regression watch: a
   `registered N tool(s)` line for `linear` in `~/.hermes/logs/agent.log` and
   any Linear-tool loss in `errors.log` surface within 30 minutes.
7. **Policy stays tighter than enforcement, deliberately.**
   `linear-ticketing` §6's closed list of 5 tools now coexists with a
   technical control allowing 10. That is the governing decision, not drift.
8. **Registration was proven on freshly-constructed agents** (headless
   `hermes chat`, same `tools.mcp_tool` code path as the gateway). The
   long-running gateway process (up since 2026-08-10 20:00:38 UTC) was not
   independently re-probed through the Slack path — sending a Slack message
   was out of scope for this run. Precedent (GCN-9,
   `status/probes/wf6/gcn9-audit.md` Claim 3) is that an `mcp_servers` config
   edit goes live on reload alone.
9. **Pruning `list_projects` / `list_cycles` removes value discovery** for
   `save_issue.project` / `.cycle` and the `project` / `cycle` / `release`
   filters on `list_issues`. Those params accept names, so a value Gaetan
   states still works; nothing uses them today. Noted so it is not met as a
   surprise.

## 9. Operator traps — record these in `docs/facts.md` next to the prune

Both are silent. Both are now live consequences of shipping `include`.

1. **`hermes tools disable linear:<tool>` / `hermes tools enable
   linear:<tool>` is a SILENT NO-OP.** `_apply_mcp_change()` in
   `hermes_cli/tools_config.py` writes only `tools_cfg["exclude"]` ("Add or
   remove specific MCP tools from a server's exclude list"), and
   `_should_register` checks `if include_set:` **first** and returns before
   ever looking at `exclude_set`. The operator gets a success message and zero
   effect. While `tools.include` is present, the only working edit paths are
   `hermes mcp configure linear` (curses, needs a TTY) or a hand edit of
   `config.yaml`.
2. **`hermes mcp configure linear` with every box ticked DELETES the guard.**
   ```python
   if len(chosen) == len(tools):
       tools_cfg.pop("exclude", None)
       tools_cfg.pop("include", None)
   ```
   "Select all" is not "keep 58 explicitly" — it removes both filters,
   silently reverting GCN-10 and restoring all 58, plus auto-enabling
   everything Linear ships thereafter.

## 10. Downstream corrections — both DONE in this lane

Both were initially flagged as follow-ups and then ruled in-lane by the
coordinator, so they ship in this same commit series.

### 10a. `docs/facts.md`

Three rows added/corrected in the Linear section: the `linear` server row no
longer claims 58 tools; a new row records the `tools.include` mechanism (bare
names, and the silent-zero foot-gun of a prefixed name); a new row records
that `hermes mcp test` is a discovery probe while **registration** is the
enforcement layer; a new row records the two §9 operator traps.

### 10b. `skills/orchestration/linear-ticketing/SKILL.md` — corrected LIVE-FIRST

The old §6 sentence read "All **58** `mcp__linear__*` tools are registered and
enabled … so this list is the **only** bound on blast radius". That went
factually stale the moment this landed.

Corrected under the GCN-13 mirror invariant — **the live tree is the source of
truth and the repo mirrors it byte-identical**, so the VM was edited first and
the repo copy taken from the live bytes. A repo-only edit would have tripped
the 18:00Z drift cron.

| step | evidence |
|---|---|
| drift check before touching anything | live and repo both `3a6303ad4ad5bd5efc21e0b2bd985631` — no divergence, safe to edit |
| live edited first | `.bak` at `~/.hermes/skills/orchestration/linear-ticketing/SKILL.md.bak-gcn10-20260811`; whole file read before writing; transferred tmp + `mv`, never in place |
| repo mirrored from live bytes | live `0e5bf9968bc095ae37915a969108da45` == repo `0e5bf9968bc095ae37915a969108da45` |
| shape preserved | 232→234 lines, 25022→25627 bytes, mode `-rw-------`, trailing newline intact |

What the correction says: 10 tools register, not 58, so the closed list is no
longer the only bound; the registered set is deliberately **wider** than the
behavioral closed list (it adds `list_issue_statuses`, `list_issue_labels`,
`list_teams`, `list_users`, `delete_comment`); **registered is not permission
to use** — the closed list of 5 is unchanged by GCN-10 and still governs
behaviour; and `LINEAR_API_KEY` remains unscoped full-write.

This is a factual correction, not a policy change. No behavioral rule was
touched, no section renumbered.

## 11. Rollback

One command, no restart:

```
flock ~/.hermes/.wf3.lock -c 'cp -p ~/.hermes/config.yaml.bak-gcn10-20260811 ~/.hermes/config.yaml'
```

Then wait ~45 s for Hermes' live reload and confirm:

- `md5sum ~/.hermes/config.yaml` → `debdf1f36c85208b314897ff9697792d`
- `wc -l ~/.hermes/config.yaml` → `211`
- the **registration-level** probe (§6c) → 58 names, and
  `~/.hermes/logs/agent.log` → `registered 58 tool(s)`
- `systemctl --user show hermes-gateway -p ActiveEnterTimestamp` →
  `Mon 2026-08-10 20:00:38 UTC`, unchanged

**Never restart the gateway** as part of a rollback, and never roll back on a
tier-1 `58` reading alone (§7). Confirm at the registration layer, which is
the only layer that ever measured the filter.
