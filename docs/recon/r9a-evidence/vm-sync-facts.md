# VM probe: kanban/Linear sync ground truth (2026-08-10)

All commands run read-only over `ssh gaetan@192.168.0.9`. No `.env`/sops touched. No edits made.

## 1. Kanban toolset still enabled

```
$ ~/.local/bin/hermes config get toolsets
- hermes-cli
- kanban
```

Confirmed live (matches r9). Only two toolsets exist system-wide — there is no per-platform/per-profile
toolset override in effect (see §3: no `profiles:` stanza exists at all).

Kanban config section (`hermes config get kanban`):

```
auto_subscribe_on_create: true
dispatch_in_gateway: true
dispatch_interval_seconds: 60
failure_limit: 2
worker_log_rotate_bytes: 2097152
worker_log_backup_count: 1
orchestrator_profile: ''
default_assignee: ''
max_in_progress_per_profile: null
auto_decompose: true
auto_decompose_per_tick: 3
dispatch_stale_timeout_seconds: 14400
reconcile_orphans: true
```

`dispatch_in_gateway: true` and `auto_decompose: true` — both r9-unverified items now CONFIRMED true.

## 2. `hermes kanban --help` — full subcommand list

```
init, boards, create, swarm, list (ls), show, assign, set-model, reclaim, reassign,
diagnostics (diag), link, unlink, claim, comment, attach, attachments, attach-rm,
complete, edit, block, schedule, unblock, promote, archive, tail, dispatch, daemon,
watch, stats, notify-subscribe, notify-list, notify-unsubscribe, log, runs, heartbeat,
assignees, context, specify, decompose, gc, repair
```

**No sync/import/export subcommand exists.** The closest name-matches are:
- `link` / `unlink` — these are **parent→child task DEPENDENCY edges within the same board**
  (`hermes kanban link parent_id child_id`), not external-system links. Verified via
  `hermes kanban link --help`: only two positional args, `parent_id child_id`, nothing else.
- `boards` — manages multiple **local** boards ("one board per project/workstream"), not
  external systems.

`hermes kanban create --help` fields: `title` (positional), `--body`, `--assignee`, `--parent`,
`--workspace`, `--branch`, `--project` (id/slug — "Link to a project", but this refers to
`hermes project list`, a *local* project registry, not Linear), `--tenant`, `--priority`,
`--triage`, `--idempotency-key`, `--max-runtime`, `--created-by`, `--skill`, `--max-retries`,
`--model`, `--provider`, `--goal`, `--goal-max-turns`, `--initial-status`, `--json`.

**No external-id/URL/Linear-key field exists on `create`.** The only place a Linear identifier
could live today is unstructured: stuffed into `title`/`--body` as free text, or into
`--idempotency-key` (a dedup string with no schema — you could put a Linear issue key there to
prevent duplicate cards, but nothing reads it back as a Linear reference; it's a bare TEXT
column purely for de-dup, `hermes kanban show`/`list` will not surface it as a link).

`hermes kanban list --help` filters: `--mine`, `--assignee`, `--status`, `--tenant`, `--session`,
`--archived`, `--json`, `--sort`, `--workflow-template-id`, `--step-key`. No external-source
filter.

**Conclusion for Q1/Q2 (design feasibility):** Hermes kanban has zero native sync primitives.
Any Linear↔kanban link has to be built as an out-of-band script (webhook or poll) that creates/
updates kanban cards via the CLI/DB and writes the Linear issue key into the free-text body or
the idempotency-key — there is nothing to "turn on."

## 3. Manual creation path / which profile serves Slack

**There is no `profiles:` stanza in `config.yaml` at all** — confirmed by:
```
$ grep -nE "^profiles:|^  [a-zA-Z_-]+:$|toolsets:" ~/.hermes/config.yaml
```
returning only `toolsets:`, `platform_toolsets:`, and generic 2-space keys under other top-level
sections (`display.platforms`, `gateway.platforms`, `telemetry.shared_metrics`, `plugins.enabled`,
etc.) — no `profiles:` key, and `ls ~/.hermes/profiles/` → `No such file or directory`.

So Tars runs as a **single implicit default profile** for everything (chat, crons, kanban
dispatch). Its toolsets ARE `hermes-cli` + `kanban` (§1) — kanban tools are reachable from the
live Slack session today. Manual card creation from Slack chat would rely on the LLM invoking
the `kanban_*` tool surface conversationally (e.g. "create a kanban card for X") — there's no
slash command (`/kanban`) route: Slack Agent-class apps have slash commands blocked platform-wide
(per operating-tars skill / prior recon), so that channel is confirmed closed, not just "likely."

`slack.reaction_triggers`: **not set** — `grep -n reaction_triggers config.yaml` returns nothing.
The only `slack:` top-level key present is `allowed_channels: C0BP2GZUFSR,C0BFQ5WFYTB`.

Full `platform_toolsets` block (config.yaml lines 156-226, key names/structure only):
```
toolsets: [hermes-cli, kanban]
platform_toolsets: {cli:[hermes-cli], telegram:[hermes-telegram], discord:[hermes-discord],
  whatsapp:[hermes-whatsapp], slack:[hermes-slack], signal:[hermes-signal],
  homeassistant:[hermes-homeassistant], qqbot:[hermes-qqbot], yuanbao:[hermes-yuanbao],
  teams:[hermes-teams], google_chat:[hermes-google_chat]}
plugins: {enabled:[hermes-lcm, rtk-rewrite], disabled:[]}
context: {engine: lcm}
mcp_servers: {slack: {...docker/korotovsky slack-mcp-server v1.3.0...}, notion: {...docker mcp/notion pinned digest...}}
onboarding: {seen: {busy_input_prompt: true}}
timezone: Europe/Paris
cron: {wrap_response: false}
platforms: {slack: {home_channel: {chat_id: C0BP2GZUFSR, name: Home}, gateway_restart_notification: true}}
slack: {allowed_channels: C0BP2GZUFSR,C0BFQ5WFYTB}
```
Note `platform_toolsets.slack: [hermes-slack]` is a DIFFERENT, apparently-unused/legacy toolset
name than the live `toolsets: [hermes-cli, kanban]` that actually governs the session — worth a
flag but out of scope here.

## 4. The Linear+Notion+Calendar rollback

### Backups on disk (`ls -la ~/.hermes/config.yaml.bak*`)

```
config.yaml.bak-b5-memory          Aug 7 16:30   94271 bytes
config.yaml.bak-b5-lcm             Aug 7 16:31    5457 bytes
config.yaml.bak-b5-adhd            Aug 7 16:32    5476 bytes
config.yaml.bak-linear-notion-cal  Aug 7 17:43    5512 bytes
config.yaml.bak-mesh-model         Aug 7 17:43    5509 bytes
config.yaml.bak-slack              Aug 7 17:50    5512 bytes
config.yaml.bak-notionpin          Aug 7 18:08    5836 bytes
config.yaml.bak-kanban             Aug 7 20:21    5908 bytes
config.yaml.bak-20260808-142749    Aug 8 14:27    6148 bytes
```
Per repo convention (`.bak` taken BEFORE the named change), `bak-linear-notion-cal` is the
snapshot from **immediately before** Linear+Notion+Calendar was wired in on 2026-08-07 17:43.

### Key-level diff (bak-linear-notion-cal → current config.yaml)

`diff <(grep -E '^[a-zA-Z_-]+:|^  [a-zA-Z_-]+:' bak) <(same on current)` shows current gained,
relative to that pre-change snapshot: `platforms:` (agent display sub-key), `toolsets:`,
`mcp_servers:` with **only `slack:` and `notion:`** children, `onboarding:`, `timezone:`,
`cron:`, `platforms:` (top-level slack home_channel), `slack: {allowed_channels}`.

**Current `mcp_servers` has NO `linear` and NO `calendar` entry** — only `slack` and `notion`
survived. This is hard structural confirmation: Notion MCP is still live (matches
`config.yaml.bak-notionpin` existing separately, taken 25 min later same evening); Linear and
Calendar MCP servers were removed from `config.yaml` at some point after 17:43 on 2026-08-07.

### Why — found in `~/.hermes/logs/errors.log`

Two WARNING lines, same session `20260807_205326_2f2ee68a`, 2026-08-07 21:29:55 and 21:35:12:
```
Tool terminal returned error (0.63s): {"output": "LIVE_CHECK_FAILED: <HttpError 403 when
requesting https://www.googleapis.com/calendar/v3/users/me/calendarList?maxResults=1&alt=json
returned \"Google Calendar API has not been used in pro[ject...]" ...
```
(truncated by log line length; message is a standard Google "API not enabled for this GCP
project" 403.) **Root cause of the Calendar rollback: the Google Calendar API was never enabled
on the backing GCP project — a live-check smoke test failed at setup time, not a design
decision.** This is a fixable prerequisite (enable the API in Google Cloud Console), not evidence
against a calendar integration in principle.

No comparable Linear-specific failure appears in `errors.log` around the same rollback window —
i.e. no evidence Linear MCP itself was broken; it was very likely pulled together with Calendar
as part of the same rollback action rather than failing on its own merits. (One unrelated Linear
GraphQL error does appear the next day, 2026-08-08 14:46:56, `linear_network_error: Variable
"$filter" got invalid v[alue...]` — this is from engagement-checker's own direct-GraphQL script,
a day after the MCP rollback, and is unrelated to the MCP-based integration that got rolled back.)

`~/.hermes/logs/gateway.log` grep for linear/notion/calendar around 2026-08-07 surfaces the human
side of the story — Gaetan asking about Notion access and floating the idea of adding more APIs:
```
20:16:36 msg='Do you have acccess to Notion ?'
20:17:10 msg='Do a summary of the "1Password migration" page'
20:50:47 msg='Nope, not that page :confused: <https://app.notion.com/p/mobileclub/1Password-mi'
21:36:07 msg='Any other APIs that could be interesting to setup ? Calendar ? Docs ?'
```
So the sequence that evening was: Notion wired in and used successfully → Gaetan asks about
adding Calendar/other APIs → Calendar setup attempted and failed live-check (403, API not
enabled) around 21:29-21:35 → rollback bundled Linear out along with the broken Calendar,
leaving only Slack+Notion in current `mcp_servers`.

Repo grep, `status/lane-a.md`, only one relevant line (line 30):
```
| 2 | Linear + Notion + Calendar + **official Notion MCP** (24 tools) | CONFIRMED — Linear
200/viewer.id · Notion loadUserContent 200 · CalDAV 207 · live MCP tool call |
`wf3-s2-linear-notion-calendar.md` |
```
This confirms the WF3 probe DID get all three working and CONFIRMED (Linear 200, Notion 200,
CalDAV 207) at build/probe time on 2026-08-07 — the later rollback in `config.yaml` therefore
happened AFTER that successful wiring, consistent with the errors.log timeline (probe pass,
then live Slack-driven Calendar API 403 later that evening, then a same-evening rollback that
took Linear down with it rather than isolating just the broken piece).

**No corroborating rollback rationale beyond the above was found** in errors.log/gateway.log
(no explicit "rolling back" log message) — the causal chain above is inferred from timestamps +
the 403 error + the config diff, not from a stated decision log entry.

## 5. Crons: profile-per-job, SKILL.md locations

`hermes cron list` output — **no per-job "Profile:" field exists in the output at all** (fields
shown: Name, Schedule, Repeat, Next run, Deliver, Skills, Last run, Execution, or Script/Mode for
no-agent jobs). This is consistent with §3: there is no profile system in this Hermes install, so
"which profile does the cron run under" has a trivial answer — **the single default profile**,
same one that serves live Slack chat and has `toolsets: [hermes-cli, kanban]`. Both
daily-work-brief and engagement-checker crons therefore DO have kanban tool access, structurally
identical to the interactive session.

Relevant cron entries:
```
e231e5faf180  Gaetan daily work brief          30 8 * * 1-5     Skills: daily-work-brief
62e8cd9db637  Gaetan engagement checker        */30 10-16 * * 1-5  Skills: engagement-checker
759e08c598e3  Gaetan engagement checker final  0 17 * * 1-5     Skills: engagement-checker
137e6c516ddf  Forward new Claude invoices      30 9 * * *       Script (no-agent)
3eb53a322510  mc-metarepo-refresh              every 60m        Script (no-agent)
```

SKILL.md locations on disk:
```
~/.hermes/skills/orchestration/engagement-checker/SKILL.md
~/.hermes/skills/orchestration/daily-work-brief/SKILL.md
```
`find ~/.hermes -iname SKILL.md` also confirms **no `~/.hermes/profiles/` directory exists**
anywhere (the earlier `ls` on that path returned ENOENT) — the "union of profiles/ dir and
current board assignees" language in `hermes kanban assignees --help` is therefore vestigial
today: with no profiles/ dir, the only assignee names that can exist are whatever ad-hoc profile
strings get passed to `hermes kanban create --assignee` / `kanban swarm --worker PROFILE:...` at
task-creation time — there's no pre-registered roster.

## Design implications (for the write-up, not asked to draft here but noted since load-bearing)

- Hermes kanban has zero sync primitives (§2) — a Linear↔kanban bridge is 100% bespoke script,
  no built-in webhook/import to lean on.
- Single-profile install (§3, §5) means kanban tool access is global, not scoped — no
  natural place to wall off "Gaetan's personal board" from anything else Tars does.
- The Notion/Slack MCP servers are proven stable in `mcp_servers`; Linear MCP was proven working
  once (WF3, 200/viewer.id) but is currently absent from config — re-adding a Linear MCP server
  is a config edit + smoke test, not new engineering, and the Calendar 403 that likely triggered
  the bundled rollback is an orthogonal, separately-fixable GCP project setting.
- engagement-checker already does direct Linear GraphQL from cron today (§4, ongoing
  `LINEAR_CURSOR`/`LINEAR_RUN_END` calls, ok as of 2026-08-10 15:05) — so a read path to Linear
  (all 3 company teams + any personal team) already exists and works without any MCP server at
  all, independent of the rolled-back MCP integration.
