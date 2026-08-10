# Hermes-native Linear MCP — probe findings (read-only, 2026-08-10)

## 1. config.yaml.bak-linear-notion-cal vs live config.yaml

**Key finding: this backup does NOT contain a linear mcp_servers entry — it never
did.** Ordered by mtime, the `.hermes/config.yaml.bak-*` chain on 2026-08-07 is:

```
16:22:15 config.yaml.installer-default   (94KB, full commented template, no mcp_servers set)
16:29:29 config.yaml.pre-b5
16:30:40 config.yaml.bak-b5-memory
16:31:35 config.yaml.bak-b5-lcm
16:32:20 config.yaml.bak-b5-adhd
17:43:37 config.yaml.bak-mesh-model        mcp_servers: {}  (empty)
17:43:42 config.yaml.bak-linear-notion-cal mcp_servers: {}  (empty)
17:50:53 config.yaml.bak-slack             mcp_servers: {}  (empty)
18:08:21 config.yaml.bak-notionpin         mcp_servers: {slack, notion}
20:21:40 config.yaml.bak-kanban            mcp_servers: {slack, notion}
2026-08-08 14:27:49 bak-20260808-142749    mcp_servers: {slack, notion}
2026-08-10 14:36:38 bak-boundaries-...     mcp_servers: {slack, notion}
2026-08-10 14:36:38 config.yaml (live)     mcp_servers: {slack, notion}
```

These `.bak-<name>` files follow a "snapshot taken right before an edit named
after the upcoming change" convention (confirmed against SOUL/config edit
hard rules). `bak-linear-notion-cal` is the **pre**-edit snapshot for a
linear+notion+cal change — it captures the state *before* that edit, which
already had zero mcp_servers. The edit that actually landed added **slack**
and **notion** only (between 17:50:53 and 18:08:21). No surviving file on
this VM — not one of the ~13 config.yaml.bak-* variants, not gateway.log,
not agent.log/.1, not errors.log — contains a genuine Linear MCP stanza or a
non-false-positive "linear" MCP mention. (`grep -il linear` hits on
`installer-default`/`pre-b5`/`bak-b5-memory` are the English word "linearly"
in a delegation-cost warning comment, not the integration.)

**Conclusion: no evidence a Linear MCP server was ever wired into
config.yaml on this VM, on 2026-08-07 or since.** The "rolled back due to
Calendar-403 collateral" Linear MCP, if it existed, left no trace in any
config backup or log — it was either never persisted to config.yaml, or
persisted and reverted within a window not covered by any surviving
snapshot/log (gateway.log itself only starts at 2026-08-07 18:43:11, i.e.
*after* the slack/notion mcp_servers edit already landed).

### Live config.yaml mcp_servers stanza (structure only, values redacted)

```yaml
mcp_servers:
  slack:
    command: docker
    args: [run, -i, --rm, --env-file, /home/gaetan/tars/slack-mcp/.env,
           ghcr.io/korotovsky/slack-mcp-server:v1.3.0, --transport, stdio, --no-cache]
  notion:
    command: docker
    args: [run, -i, --rm, -e, NOTION_TOKEN, <redacted>]
    env:
      NOTION_TOKEN: ${NOTION_API_TOKEN}
```

Diff to re-add linear: add a third top-level key `linear:` under
`mcp_servers:` — installer-default's documented example forms are:

```yaml
# stdio form
linear:
  command: <cmd>
  args: [...]
  env:
    LINEAR_API_KEY: ${LINEAR_API_KEY}

# or HTTP form (like notion's url-only style)
linear:
  url: https://mcp.linear.app/mcp   # or whatever official endpoint
  headers:
    Authorization: ${LINEAR_API_KEY}
```

Exact shape (stdio vs HTTP, OAuth vs header) is NOT recoverable from local
files — `hermes mcp catalog` lists `linear` as a Nous-approved one-click
preset ("Find, create, and update Linear issues, projects, and comments")
but the preset definition itself is fetched remotely by the catalog/install
command, not cached anywhere in `~/.hermes/hermes-agent/` (grepped the whole
installed tree, only false-positive CSS/icon "linear" hits). Figma's catalog
entry is explicitly OAuth; Linear's catalog line has no auth hint visible
locally. Running `hermes mcp install linear` would reveal it but was not run
(would mutate config — out of scope for a read-only probe).

## 2. `hermes mcp --help` / test

```
hermes mcp {serve,add,remove,rm,list,ls,test,configure,config,login,reauth,picker,catalog,install}
```

- `hermes mcp add <name> [--url URL | --command CMD --args ...] [--auth {oauth,header}] [--preset NAME] [--env KEY=VALUE ...]`
  — config-editing command, not config-only manual editing (though manual
  YAML edit under `mcp_servers:` + reload also works, same as slack/notion
  were presumably added).
- `hermes mcp catalog` — lists Nous-approved one-click installs. `linear` is
  in it: `linear   available   Find, create, and update Linear issues,
  projects, and comments.` (also figma, n8n, unreal-engine, comfy-cloud).
- `hermes mcp install <identifier>` — installs a catalog entry by name
  (`hermes mcp install linear`).
- `hermes mcp test <name>` tests connection for an already-configured
  server (per its `--help`, just `name` — no flags). Per the operating-tars
  skill, `mcp list` never actually connects; `mcp test` does. Not run here
  since linear isn't wired (per probe scope).

## 3. Toolsets vs MCP — are MCP tools gated?

`hermes config get toolsets` → `[hermes-cli, kanban]`. This is the
**platform_toolsets** mechanism (per-platform tool bundles: web, terminal,
file, browser, vision, image_gen, skills, todo, tts, cronjob, plus presets
like hermes-cli/hermes-telegram/hermes-slack, plus composites
debugging/safe/all). `kanban` is a **custom** toolset name added to this
list (not one of the documented built-ins) — i.e. kanban tools genuinely are
gated behind an explicit platform_toolsets entry, matching the premise in
the task.

**MCP servers are a separate, ungated mechanism.** installer-default's
documentation for the `mcp_servers:` block states: "Each server's tools are
automatically discovered and registered" — no toolset entry required. This
is corroborated by live behavior: neither `slack` nor `notion` appears
anywhere in `platform_toolsets` or the `toolsets` list, yet their tools are
used directly by skills (`mcp__slack__conversations_search_messages` is
called from the `engagement-checker` skill with no toolset gate). So: **MCP
server tools are NOT gated per-profile/toolset like kanban was** — once a
server is defined under `mcp_servers:` and connects successfully, its tools
are unconditionally available to the agent loop.

Tool name prefix: `mcp__<server-name>__<tool>` — e.g. `mcp__slack__*`,
`mcp__notion__*` (implied, not directly greped but consistent with the
`mcp__slack__conversations_search_messages` reference in
engagement-checker/SKILL.md). A re-wired linear server named `linear` would
expose tools as `mcp__linear__*`.

There's also a distinct `inherit_mcp_toolsets` setting under `delegation:`
(installer-default, commented, default true): "When explicit child toolsets
are narrowed, also keep the parent's MCP toolsets... Set false for strict
intersection." This only matters for **subagent delegation** (child agents
spawned by the main loop) — it does not gate the top-level agent loop's
access to MCP tools.

## 4. Cron/skill context — full agent loop or restricted?

`hermes cron --help`: `{list,create,add,edit,pause,resume,run,remove,rm,
delete,status,runs,history,notepad,tick}`.

`hermes cron list` shows live entries, e.g.:
- `Gaetan engagement checker` — schedule `*/30 10-16 * * 1-5`, Skills:
  `engagement-checker`, deliver `slack:C0BP2GZUFSR:...`, last run "ok".
- `Gaetan daily work brief` — Skills: `daily-work-brief`, same delivery
  channel, last run "ok".
Both are `Skills:`-mode jobs (as opposed to the `Script:`/no-agent jobs like
`claude_invoice_forward.py` or `kb-refresh.sh`, which run a script directly
with no agent loop at all).

Evidence the Skills-mode cron jobs run through the **full agent loop with
MCP tools available**:
- `errors.log` line `2026-08-10 10:31:06 ... cron_62e8cd9db637_20260810_123056
  ... Tool kanban_show returned erro...` — a toolset-gated tool
  (`kanban_show`) executing inside a cron-invoked session, proving cron
  sessions get the same tool surface as interactive ones, not a stripped-down
  context.
- `engagement-checker/SKILL.md` (the skill that IS this cron job's payload)
  explicitly instructs: "Use the Slack MCP tools directly. Gaetan's Slack
  user ID is `U08BDJAMSRZ`." and later "Use
  `mcp__slack__conversations_search_messages` with a bounded date filter
  derived from the Slack cursor..." — a skill prompt can and does name an
  MCP tool by its `mcp__<server>__<tool>` name and the cron job actually
  calls it (cron history shows "ok" completions, not tool-not-found errors).

So: **yes, a cron-invoked skill runs through the full agent loop with MCP
tools available** — the skill prompt can say "use the linear tools" and it
would resolve to `mcp__linear__*` the same way it resolves
`mcp__slack__conversations_search_messages` today, IF a `linear` MCP server
were wired. Currently it isn't, which is exactly why
`engagement-checker/SKILL.md` §4 "Collect Linear deltas" instead hand-rolls
raw `urllib` GraphQL calls against `https://api.linear.app/graphql` guarded
by `os.environ["LINEAR_API_KEY"]` — a deliberate workaround for the missing
MCP wiring, not a cron/agent-loop restriction. (Matches the recent commit
"engagement-checker skill: use direct Linear GraphQL reads".)

`daily-work-brief/SKILL.md` mentions Linear only in prose ("Use any
already-configured Linear integration, CLI, API, email, or Slack links...")
— no MCP tool call, no raw GraphQL code. It doesn't touch Linear directly at
all; it defers entirely to whatever's available.

## 5. Auth — env var name

`~/.hermes/.env` var names present (values never read): `GMAIL_ADDRESS`,
`GMAIL_APP_PASSWORD`, `HINDSIGHT_MODE`, `LINEAR_API_KEY`,
`NOTION_API_TOKEN`, `SLACK_ALLOWED_USERS`, `SLACK_APP_TOKEN`,
`SLACK_BOT_TOKEN`, `SLACK_HOME_CHANNEL`, `SLACK_MCP_XOXC_TOKEN`,
`SLACK_MCP_XOXD_TOKEN`.

**`LINEAR_API_KEY` already exists** in `.env` (also referenced by
`engagement-checker/SKILL.md` lines 122/205/256/267/269-270 as
`os.environ["LINEAR_API_KEY"]`, `LINEAR_CURSOR`, `LINEAR_RUN_END`). This is a
personal API key, not an OAuth flow — the skill sends it as a raw
`Authorization` header value (no `Bearer` prefix), which is Linear's
documented convention for personal API keys. There is no separate
`LINEAR_OAUTH_*`/`LINEAR_CLIENT_ID`/`LINEAR_CLIENT_SECRET` anywhere in
`.env` or in any config.yaml.bak. So: if Hermes's `linear` catalog preset
uses `--auth header`, `LINEAR_API_KEY` is directly reusable; if the catalog
preset forces `--auth oauth` (like figma's catalog entry), that would be a
**new** auth path independent of the existing key — undetermined without
running `hermes mcp install linear`, out of scope for this read-only probe.
Also present: 4 unrelated `.env.bak-*` files (cutover, linear-notion-cal,
notiontrim, sethome) — same pre-edit-snapshot convention as config.yaml;
none inspected for content beyond var names since the ask was structure
only.

## 6. 2026-08-07 logs — Linear MCP connecting successfully?

**No evidence found.** `grep -ic linear ~/.hermes/logs/gateway.log` → `0`
(zero occurrences, ever, in the current gateway.log). The file's first line
is `2026-08-07 18:43:11,560 INFO gateway.run: Starting Hermes Gateway...` —
i.e. gateway.log only covers from 18:43:11 onward on 08-07, which is
*already after* the mcp_servers edit window (slack+notion landed between
17:50:53–18:08:21 per the bak timestamps above). If a linear MCP server ever
connected before 18:43:11, no log capturing that window survives (no rotated
gateway.log.1, no archived copy — `ls ~/.hermes/logs/` shows only
`agent.log`/`agent.log.1`, `errors.log`, `gateway.log` (single, current),
`gateway-exit-diag.log`, `gateway_faulthandler.log` (empty),
`gateway-shutdown-diag.log`, `mcp-stderr.log`). `journalctl --user -u
hermes-gateway.service --since 2026-08-07 --until 2026-08-08 | grep -i
linear` also returned nothing. `mcp-stderr.log` (per-MCP-subprocess stderr, 164KB) also checked: `grep -ic
linear` → `0`. Every log surface on this VM (gateway.log, agent.log,
agent.log.1, errors.log, mcp-stderr.log, journald) is consistent with **the
Linear MCP server never actually running on this VM**, contrary to the
task's stated premise.
