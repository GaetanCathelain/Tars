# Probe — Hermes retrieval mechanisms (P6/kb)

All commands run against `gaetan@192.168.0.9` (Tars VM), read-only, 2026-08-07
~22:37–22:41 UTC. Hermes source checkout: `/home/gaetan/.hermes/hermes-agent/`.
No writes made anywhere under `~/.hermes/` or `~/dev/mc-metarepo`; the one
"write-shaped" action taken was a single `hermes chat -Q -q` oneshot turn
(the project's own sanctioned probe pattern, see `docs/facts.md` /
`operating-tars` skill) used purely to prove tool-driven search end to end.

## Summary — ranked mechanisms that actually exist

1. **`search_files` + `read_file` (the `file` toolset)** — already live today.
   Part of `_HERMES_CORE_TOOLS`, which is the tool list for *both* the
   `hermes-cli` toolset (CLI) and `hermes-slack` toolset (Slack) — the two
   platforms share one list. Ripgrep-backed regex content search + glob file
   search + arbitrary file read, with **no path jail** (only `write_file`/
   `patch` carry a soft cross-Hermes-profile guard, opt-outable). Proven live
   end-to-end against `~/dev/mc-metarepo` (§1 below, §3 evidence). **This is
   the mechanism for P6 item 1 — no config change needed.**
2. **`terminal` (shell exec)** — also in `_HERMES_CORE_TOOLS`, so also already
   live on both CLI and Slack. Full subprocess exec as user `gaetan`,
   `backend: local` (config.yaml), no container/sandbox. Model can run
   `git -C ~/dev/mc-metarepo log/show/diff`, `rg`, anything. `tirith` command
   vetting is enabled-but-not-installed (pattern-matching only, pre-existing,
   per `docs/facts.md`) — i.e. there is effectively no guard rail on this tool
   today. **This is the mechanism for P6 item 2's hourly `git pull`/`fetch`
   (via `hermes cron create` → a `terminal` command), and a fallback search
   path.**
3. **`hermes-lcm` plugin (`lcm_grep`/`lcm_recall`/`lcm_retrieve`/…, 15 tools)**
   — a real hybrid FTS+embedding vector store exists (`vector_store.py`,
   `embedding_provider.py`, backing `~/.hermes/lcm.db`), but it is the
   **context-engine's own conversation memory** (current session + past
   sessions), not a generic document index. Its own docs say to use
   `session_search` / these tools for "Hermes-tracked history" — nothing in
   its tool surface takes an arbitrary directory. **Not usable for indexing
   mc-metarepo without separately building an ingestion path into `lcm.db`** —
   out of scope for a search-on-demand design that should just call
   `search_files`.
4. **Skills (`~/.hermes/skills/<name>/SKILL.md`)** — auto-injected into model
   context when judged relevant (`skills_list`/`skill_view`/`skill_manage`,
   already in the core toolset). A prose/pointer mechanism, not a search
   mechanism — a skill could tell the model "grep mc-metarepo via
   `search_files`" but does no searching itself.
5. **Built-in `memory` tool** — session notes + user profile,
   `memory_char_limit: 2200` / `user_char_limit: 1375` (config.yaml). Config
   shows `memory.provider: hindsight` **actively set**, which contradicts the
   task brief's "hindsight is deliberately skipped for v1 (no keys)" — flagged
   as a live discrepancy to verify with whoever owns that config edit (see §4
   caveat; two peers are editing `~/.hermes/` concurrently right now). Either
   way, this is not a document-retrieval mechanism for mc-metarepo.
6. **MCP** — only `slack` and `notion` servers registered
   (`hermes mcp list`); the Nous-approved catalog (`hermes mcp catalog`) has
   **no filesystem/docs/RAG server** on offer. `node`/`npx` are **absent from
   PATH** on the VM, so a stock `@modelcontextprotocol/server-filesystem`
   MCP server can't be run without first installing Node — an unnecessary new
   dependency given #1 already covers this.
7. **`cronjob` tool + `hermes cron create`** — already in the core toolset;
   this is the scheduling primitive for P6 item 2 (hourly re-pull), not a
   retrieval mechanism itself.
8. **`kanban` toolset (12 `kanban_*` tools)** — coordination/handoff only,
   relevant to P6 item 3 (push-back via a delegated Orca session tracked as a
   card), unrelated to search or retrieval.

**Model can run shell commands today: YES**, unconditionally, on both CLI and
Slack, via the `terminal` tool already in the enabled `hermes-cli`/`kanban`
toolset stack — no new toolset, no MCP server, no skill needed for P6 item 1.

---

## 1. CLI surface

`~/.local/bin/hermes --help` (top-level subcommand list, verbatim from
positional-arguments block):

```
{chat,model,moa,fallback,secrets,egress,migrate,gateway,proxy,lsp,setup,
whatsapp,whatsapp-cloud,slack,send,logout,auth,status,pause,resume,cron,
sync,webhook,portal,kanban,project,hooks,doctor,security,approvals,dump,
debug,backup,checkpoints,import,import-agent,config,skin,console,pairing,
skills,bundles,plugins,curator,pets,journey,memory,tools,computer-use,mcp,
sessions,insights,monitoring,claw,version,update,uninstall,acp,profile,
completion,dashboard,serve,desktop,logs,prompt-size}
```

`hermes chat --help` (relevant flags):
```
usage: hermes chat [-h] [-q QUERY] [--image IMAGE] [-m MODEL] [-t TOOLSETS]
                   [--reasoning LEVEL] [-s SKILLS] [--provider PROVIDER] [-v]
                   [-Q] [--resume SESSION_ID] ...
  -q QUERY, --query QUERY   Single query (non-interactive mode)
  -t TOOLSETS, --toolsets TOOLSETS   Comma-separated toolsets to enable
  -Q, --quiet   Quiet mode for programmatic use
```
(Confirms `docs/facts.md`'s note that oneshot is `hermes chat -Q -q '<query>'`,
`--oneshot` is a top-level `-z` flag on the bare `hermes` command, not on `chat`.)

`hermes tools --help`:
```
usage: hermes tools [-h] [--summary] {list,disable,enable,post-setup} ...
Built-in toolsets use plain names (e.g. web, memory). MCP tools use
server:tool notation (e.g. github:create_issue).
```

`hermes mcp --help`:
```
usage: hermes mcp [-h] [--accept-hooks]
                  {serve,add,remove,rm,list,ls,test,configure,config,login,
                   reauth,picker,catalog,install} ...
```

`hermes skills --help`:
```
usage: hermes skills [-h]
   {browse,search,install,inspect,list,check,update,audit,uninstall,reset,
    list-modified,diff,opt-out,opt-in,repair-official,publish,snapshot,
    tap,config} ...
```

`hermes memory --help`:
```
usage: hermes memory [-h] {setup,status,off,reset} ...
Available providers: honcho, openviking, mem0, hindsight, holographic,
retaindb, byterover. Only one external provider can be active at a time.
Built-in memory (MEMORY.md/USER.md) is always active.
```

`hermes cron --help`:
```
usage: hermes cron [-h] [--accept-hooks]
   {list,create,add,edit,pause,resume,run,remove,rm,delete,status,runs,
    history,notepad,tick} ...
```

`hermes config --help`:
```
usage: hermes config [-h]
   {show,edit,get,set,unset,path,env-path,check,migrate} ...
```

No `hermes toolsets` subcommand exists — toolset management is under
`hermes tools`.

## 2. Toolset registry in source (`toolsets.py`)

File: `/home/gaetan/.hermes/hermes-agent/toolsets.py` (top-level, not under
`tools/`).

Shared base list `_HERMES_CORE_TOOLS` (used verbatim, or +extras, by
`hermes-cli`, `hermes-slack`, `hermes-signal`, `hermes-bluebubbles`,
`hermes-homeassistant`, `hermes-email`, `hermes-mattermost`, `hermes-matrix`,
`hermes-dingtalk`, `hermes-feishu`, `hermes-weixin`, `hermes-telegram`,
`hermes-discord`, `hermes-whatsapp`, `hermes-qqbot`, `hermes-teams`,
`hermes-google_chat`, `hermes-api-server`, `hermes-acp`, `coding`) — this is
the "narrow waist" the source comments describe:

```python
_HERMES_CORE_TOOLS = [
    "web_search", "web_extract",
    "terminal", "process",
    "read_file", "write_file", "patch", "search_files",
    "vision_analyze", "image_generate",
    "bfl_flux3_text_to_video", ... ,
    "skills_list", "skill_view", "skill_manage",
    "browser_navigate", ... , "browser_dialog",
    "text_to_speech",
    "todo", "memory",
    "session_search",
    "clarify",
    "execute_code", "delegate_task",
    "cronjob",
    "ha_list_entities", "ha_get_state", "ha_list_services", "ha_call_service",
    "kanban_show", "kanban_list", "kanban_complete", "kanban_block",
    "kanban_heartbeat", "kanban_comment", "kanban_create", "kanban_link",
    "kanban_unblock", "kanban_attach", "kanban_attach_url", "kanban_attachments",
    "computer_use",
]
```

Every toolset name and tool list, as literally defined (`grep -n` line refs
into `toolsets.py`):

| Toolset name | Tools (summary) |
|---|---|
| `web` | `web_search`, `web_extract` |
| `search` | `web_search` only |
| `x_search` | `x_search` (opt-in, xAI creds) |
| `vision` | `vision_analyze` |
| `video` | `video_analyze` (off by default) |
| `image_gen` | `image_generate` |
| `video_gen` | `video_generate`, `xai_video_edit`, `xai_video_extend` |
| `bfl` | 6 `bfl_flux3_*` tools |
| `computer_use` | `computer_use` |
| `terminal` | **`terminal`, `process`** |
| `skills` | `skills_list`, `skill_view`, `skill_manage` |
| `browser` | 12 `browser_*` + `web_search` |
| `cronjob` | `cronjob` |
| **`file`** | **`read_file`, `write_file`, `patch`, `search_files`** |
| `tts` | `text_to_speech` |
| `todo` | `todo` |
| `memory` | `memory` |
| `context_engine` | `[]` — "Runtime tools exposed by the active context engine" (populated dynamically; see §8) |
| `session_search` | `session_search` |
| `project` | `project_list/create/switch` (desktop GUI only) |
| `desktop_ui` | GUI-only pane/terminal affordances (desktop app only, not CLI/Slack) |
| `clarify` | `clarify` |
| `code_execution` | `execute_code` |
| `delegation` | `delegate_task` |
| `homeassistant` | 4 `ha_*` tools |
| `kanban` | 12 `kanban_*` tools, gated on `HERMES_KANBAN_TASK` env or profile enabling it |
| `discord` / `discord_admin` | Discord-specific |
| `yuanbao`, `feishu_doc`, `feishu_drive`, `spotify` | platform-specific |
| `debugging` | `terminal`, `process` + includes `web`, `file` |
| `safe` | includes `web`, `vision`, `image_gen` — no terminal |
| `coding` | full core minus messaging/tts/image_gen/spotify/HA/cron/computer-use (posture toolset, auto-selected in a code workspace) |
| **`hermes-cli`** | **= `_HERMES_CORE_TOOLS`** verbatim ("Full interactive CLI toolset") |
| `hermes-cron` | = `_HERMES_CORE_TOOLS` (mirrors hermes-cli) |
| **`hermes-slack`** | **= `_HERMES_CORE_TOOLS`** verbatim — `"Slack bot toolset - full access for workspace use (terminal has safety checks)"` |
| `hermes-signal`, `hermes-bluebubbles`, `hermes-homeassistant`, `hermes-email`, `hermes-mattermost`, `hermes-matrix`, `hermes-dingtalk` | all = `_HERMES_CORE_TOOLS` verbatim |
| `hermes-feishu` | `_HERMES_CORE_TOOLS` + 5 feishu tools |
| `hermes-api-server` | near-`_HERMES_CORE_TOOLS` minus kanban/messaging bits, for the OpenAI-compatible HTTP server |
| `hermes-acp` | coding-focused subset (editor integration) |

Direct answers to the specific questions asked:
- **Shell/bash/exec toolset**: yes — `terminal` toolset (`terminal`, `process`
  tools), and it is inside `_HERMES_CORE_TOOLS` so it ships by default in
  `hermes-cli` and `hermes-slack` alike.
- **Filesystem/file-read toolset**: yes — `file` toolset (`read_file`,
  `write_file`, `patch`, `search_files`), also inside `_HERMES_CORE_TOOLS`.
- **Grep/search toolset**: `search_files` (inside `file`) IS the grep/find/ls
  replacement — its own schema description: *"Search file contents or find
  files by name. Use this instead of grep/rg/find/ls in terminal.
  Ripgrep-backed, faster than shell equivalents."* (verbatim,
  `tools/file_tools.py:2504`). There is no separate toolset named `grep` or
  `search`-for-files; `search_files` is it.
- **Memory/RAG toolset**: `memory` toolset = built-in session notes/profile
  (not RAG). The only real embedding/vector-backed retrieval system in the
  source tree is the `hermes-lcm` plugin (§8) — not a toolset entry in
  `toolsets.py` at all; its tools are injected by the plugin itself into the
  `context_engine` toolset slot when LCM is the active context engine.

## 3. What `hermes-cli` and `kanban` (currently enabled) expose to the model

`toolsets: [hermes-cli, kanban]` in `~/.hermes/config.yaml` resolves to:
- `hermes-cli` → all of `_HERMES_CORE_TOOLS` listed in full in §2 above —
  including `terminal`, `process`, `read_file`, `write_file`, `patch`,
  `search_files`, `memory`, `session_search`, `todo`, `clarify`,
  `execute_code`, `delegate_task`, `cronjob`, web/vision/image/video-gen/tts,
  browser automation, skills tools, HA tools, computer_use.
- `kanban` → `kanban_show`, `kanban_list`, `kanban_complete`, `kanban_block`,
  `kanban_heartbeat`, `kanban_comment`, `kanban_create`, `kanban_link`,
  `kanban_unblock`, `kanban_attach`, `kanban_attach_url`, `kanban_attachments`
  (12 tools, matches `docs/facts.md`'s existing note).

Because `hermes-slack` (the toolset Slack actually loads per
`platform_toolsets.slack`) is **also** `= _HERMES_CORE_TOOLS` verbatim, Slack
gets the identical file/terminal/search surface as the CLI — confirmed by the
live oneshot test in §7/Summary #1 (which ran through the `cli` platform path,
same toolset shape as Slack).

`hermes tools --summary` (interactive-only; `hermes tools list` works
non-interactively):

```
$ COLUMNS=200 ~/.local/bin/hermes tools list
Built-in toolsets (cli):
  ✓ enabled  web  🔍 Web Search & Scraping
  ✓ enabled  browser  🌐 Browser Automation
  ✓ enabled  terminal  💻 Terminal & Processes
  ✓ enabled  file  📁 File Operations
  ✓ enabled  code_execution  ⚡ Code Execution
  ✓ enabled  vision  👁️  Vision / Image Analysis
  ✗ disabled  video  🎬 Video Analysis
  ✓ enabled  image_gen  🎨 Image Generation
  ✗ disabled  video_gen  🎬 Video Generation
  ✓ enabled  bfl  🎬 BFL FLUX 3 Video
  ✗ disabled  x_search  🐦 X (Twitter) Search
  ✓ enabled  tts  🔊 Text-to-Speech
  ✗ disabled  stt  🎙️ Speech-to-Text
  ✓ enabled  skills  📚 Skills
  ✓ enabled  todo  📋 Task Planning
  ✓ enabled  memory  💾 Memory
  ✓ enabled  context_engine  🧩 Context Engine
  ✓ enabled  session_search  🔎 Session Search
  ✓ enabled  clarify  ❓ Clarifying Questions
  ✓ enabled  delegation  👥 Task Delegation
  ✓ enabled  cronjob  ⏰ Cron Jobs
  ✗ disabled  homeassistant  🏠 Home Assistant
  ✗ disabled  spotify  🎵 Spotify
  ✗ disabled  yuanbao  🤖 Yuanbao
  ✓ enabled  computer_use  🖱️  Computer Use (macOS/Windows/Linux)

MCP servers:
  slack  all tools enabled
  notion  all tools enabled
```
(This lists the `cli` platform's toolset composition; `file` — containing
`search_files`/`read_file` — is enabled. No per-platform breakdown flag was
found in `hermes tools list --help`'s options; not pursued further, time-boxed.)

## 4. Current live config (read-only)

Read at `2026-08-07 22:41 UTC` (`stat -c '%y' ~/.hermes/config.yaml` reported
mtime `2026-08-07 20:21:48 UTC` — i.e. last modified ~2h20m before this read,
not mid-write at read time, but sibling peers are actively working in this
project so treat as a snapshot, not a permanent fact).

Exact `toolsets:` stanza:
```yaml
# WF5: unlocks the model-facing kanban_* tools. The kanban toolset is already
# in enabled_toolsets for cli+slack; this key is what _profile_has_kanban_toolset()
# (tools/kanban_tools.py check_fn) reads. Keep hermes-cli or CLI loses its tools.
toolsets:
- hermes-cli
- kanban
platform_toolsets:
  cli:
  - hermes-cli
  telegram:
  - hermes-telegram
  discord:
  - hermes-discord
  whatsapp:
  - hermes-whatsapp
  slack:
  - hermes-slack
  signal:
  - hermes-signal
  ... (homeassistant, qqbot, yuanbao, teams, google_chat — each own hermes-<platform>)
```

`memory:` stanza:
```yaml
memory:
  memory_enabled: true
  user_profile_enabled: true
  memory_char_limit: 2200
  user_char_limit: 1375
  provider: hindsight
  nudge_interval: 10
  flush_min_turns: 6
```
**Caveat**: this shows `provider: hindsight` actively configured, which
conflicts with the task brief's stated fact "Hindsight memory is deliberately
skipped for v1 (no keys)." Not reconciled here — flag for whoever owns that
key; two sibling peers are editing this file live.

`context:` stanza: `context: {engine: lcm}` — confirms LCM is the active
context engine (relevant to §8).

`plugins:` stanza:
```yaml
plugins:
  enabled:
  - hermes-lcm
  - rtk-rewrite
  disabled: []
```

`mcp_servers:` stanza (only 2 servers, neither filesystem/RAG):
```yaml
mcp_servers:
  slack:
    command: docker
    args: [run, -i, --rm, --env-file, /home/gaetan/tars/slack-mcp/.env,
           ghcr.io/korotovsky/slack-mcp-server:v1.3.0, --transport, stdio, --no-cache]
  notion:
    command: docker
    args: [run, -i, --rm, -e, NOTION_TOKEN,
           mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9]
    env:
      NOTION_TOKEN: ${NOTION_API_TOKEN}
```

No `tools:`/`skills:` override stanza was present in the file (nothing to
report there — toolset composition is entirely via `toolsets:` +
`platform_toolsets:` above).

No `gateway.platforms.slack` change beyond what `docs/facts.md` already
records (`require_mention: true`, `strict_mention: true`,
`unauthorized_dm_behavior: ignore`) — read but not re-quoted here since
already-verified elsewhere.

## 5. Skills — existing skill directories (ls only, read-only)

```
$ ls ~/.hermes/skills/
apple  autonomous-ai-agents  creative  delegate-to-cooper  email  github
hermes-lcm  hermes-operations  i-have-adhd  media  mlops  note-taking
productivity  research  smart-home  social-media  software-development
```

17 entries. Per `docs/facts.md`, each `<name>/SKILL.md` is loaded into model
context unprompted when relevant. `hermes-lcm` here is the plugin-shipped
skill that carries LCM's "recall policy" doc referenced in §8 (not inspected
further — read-only names only, per instructions; peers are editing this dir).

## 6. MCP

`hermes mcp list` (registry-only, per `docs/facts.md`):
```
  MCP Servers:
  Name             Transport                      Tools        Status
  ──────────────── ────────────────────────────── ──────────── ──────────
  slack            docker run -i                  all          ✓ enabled
  notion           docker run -i                  all          ✓ enabled
```

`hermes mcp catalog` (Nous-approved one-click installs):
```
  Name               Status                   Description
  ------------------ ------------------------ -----------
  blender            available                Drive a live Blender session
  comfy-cloud        available                Generate images/video/audio/3D
  figma              available                Official Figma remote MCP
  linear             available                Find/create/update Linear issues
  n8n                available                Manage n8n workflows
  unreal-engine      available                Drive Unreal Engine 5.8 editor
  notion             custom — enabled         docker
  slack              custom — enabled         docker
```
**No filesystem/docs/RAG MCP server** is registered or offered in the catalog.

Node/npx availability:
```
$ node --version
bash: line 1: node: command not found
$ npx --version
bash: line 1: npx: command not found
$ which npm
(no output)
$ find / -maxdepth 4 -iname node -type f 2>/dev/null
(no results — only /proc/irq/*/node false positives)
```
`node`/`npx` are **absent from PATH and from the filesystem** at shallow
search depth on this VM — a stock `npx @modelcontextprotocol/server-filesystem`
MCP server is not runnable today without first installing Node (an added
dependency/attack surface the design doesn't need, given §1/§2 already work).

## 7. Base tooling on the VM

```
$ rg --version | head -1
ripgrep 14.1.0
$ git --version
git version 2.43.0
$ fd --version
bash: line 1: fd: command not found
$ ctags --version
bash: line 1: ctags: command not found
$ python3 --version
Python 3.12.3
```
`rg` and `git` are installed; `fd` and `ctags` are not.

**Is there a shell/bash tool the model can call TODAY?** Yes —
confirmed by source (`toolsets.py`, `tools/terminal_tool.py`) AND live test.
`terminal_tool.py`'s schema: `"description": "The command to execute on the
VM"`, `terminal.backend: local` in `config.yaml` (no docker/sandbox layer by
default) — genuine unrestricted subprocess exec as `gaetan`.

`search_files` is **not path-jailed**: `tools/path_security.py`'s
`validate_within_dir()` helper is used only by `skill_manager_tool`,
`skills_tool`, `skills_hub`, `cronjob_tools`, `credential_files` — not by
`file_tools.py`. The only guard in `file_tools.py` is `_check_cross_profile_path`
(`file_tools.py:991`), and it only fires for `write_file`/`patch` (soft
warning, opt-out via `cross_profile=true`) — **`read_file` and `search_files`
have no path restriction at all.**

**Live end-to-end proof** — oneshot chat, forced to use only `search_files`+
`read_file`, searching the real `~/dev/mc-metarepo` clone that the VM already
holds:
```
$ ~/.local/bin/hermes chat -Q -q "Using only your search_files and read_file \
  tools (no other means), find the phrase RECIPIENTS inside ~/dev/mc-metarepo \
  and tell me which file it is in and quote one matching line. Do not write \
  or execute git commands, just search_files + read_file."

session_id: 20260807_224013_140e09
No exact uppercase `RECIPIENTS` match exists. Case-insensitive matches occur in
three files; one is:

`~/dev/mc-metarepo/secrets/README.md:96`

> `# 3. Rotate the DATA key and swap recipients — `sops rotate`, NOT `sops
updatekeys`:`
```
This ran through the `hermes-cli` toolset (same tool list as `hermes-slack`),
zero config changes, and correctly found+quoted a real line in the real repo.

Sanity check that the clone is real and current (read-only, `git log`/`status`
only, per hard rules):
```
$ git -C ~/dev/mc-metarepo log -1 --oneline
4831253 chore(submodules): auto-update to latest tracked branches (#181)
$ git -C ~/dev/mc-metarepo status -sb
## main...origin/main
```
Clean, tracking `origin/main`, matches the task brief's starting fact.

## 8. Native document-retrieval / RAG in Hermes source

`grep -rli 'rag\|embed\|vector\|chroma\|faiss\|retriev\|index'` across the
Hermes source tree turned up hits concentrated in exactly one place outside
test/build noise: the **`hermes-lcm` plugin**, installed at
`~/.hermes/plugins/hermes-lcm/` (NOT under the `hermes-agent` source checkout
— it's a separately-installed plugin, version `0.21.0-rc2` per its manifest).

`~/.hermes/plugins/hermes-lcm/plugin.yaml`:
```yaml
name: hermes-lcm
version: 0.21.0-rc2
description: "Lossless Context Management — standalone Hermes plugin with a
  DAG-based context engine that never loses a message"
provides_tools:
  - lcm_grep
  - lcm_recall
  - lcm_query_state
  - lcm_compute
  - lcm_compile_evidence
  - lcm_evidence_pack
  - lcm_retrieve
  - lcm_recent
  - lcm_load_session
  - lcm_describe
  - lcm_expand
  - lcm_expand_query
  - lcm_status
  - lcm_inspect
  - lcm_doctor
```
Real files back this up: `vector_store.py` (125 KB), `embedding_provider.py`
(71 KB), `~/.hermes/lcm.db` (SQLite, live on disk) — this is not a stub.

**But it is scoped to conversation memory, not documents.** From
`docs/retrieval-tools.md` (the plugin's own reference, read read-only):
> `lcm_grep` — "Search current-session raw messages and summaries."
> `lcm_recall` — "Search the agent's entire memory across ALL conversations
> and all time by meaning... Runs three arms over the whole **local
> database**."
> "Use `session_search` for earlier separate sessions or broad cross-session
> recall [outside `lcm.db`]."

Every tool operates over `lcm.db` (raw messages, summaries, embedded
verbatim chunks of **past conversations**) — none takes an arbitrary
filesystem path or directory as input. There is no ingestion path from an
external git clone into this store today. Confirmed a stub/non-fit for
indexing `mc-metarepo`: real infrastructure, wrong corpus.

`context_engine` toolset in `toolsets.py` is declared with `"tools": []` —
a deliberate empty placeholder, populated at runtime by whichever plugin is
the active `context.engine` (here, `hermes-lcm`, per config.yaml `context:
{engine: lcm}`), matching the comment: *"Runtime tools exposed by the active
context engine."*

No other rag/embed/vector/chroma/faiss hits exist in the `hermes-agent`
source tree outside this plugin, `node_modules/` build tooling, and unrelated
matches (e.g. `liblcms2` — a color-management library, not context/LCM).

---

## Not tested

- **Slack-specific `hermes tools list --platform slack`** (or equivalent) —
  no such flag surfaced in `--help`; inferred Slack's tool set from
  `toolsets.py` source (`hermes-slack` = `_HERMES_CORE_TOOLS` verbatim) plus
  the CLI-path live test, not a live Slack-channel probe. Confidence is high
  (source is unambiguous) but not empirically fired over the Slack surface
  itself in this probe.
- **Whether `search_files`/`terminal` are actually reachable from a live
  Slack turn today** (vs. just resolved into the schema) — not fired over
  Slack; the live proof in §7 used the `cli` platform toolset path
  (`hermes chat -Q -q`), which resolves to the identical tool list per source,
  not a native Slack message. Time-boxed: firing a real Slack turn needs the
  VM's user token per `docs/facts.md`'s connector-invisibility finding, which
  felt like scope creep for a recon-only task.
- **`lcm_grep`/`lcm_recall` actually invoked live** — read from source +
  docs only, not exercised with a real query, since scoping (§8) already
  rules the mechanism out for this proposal without needing a live call.
  `hermes mcp test <name>` was not run either (only `hermes mcp list`/
  `catalog`, both registry-only) — not needed since neither registered server
  is retrieval-relevant.
- **Whether any other MCP server could be added later that provides
  filesystem/RAG search** (e.g. installing Node first) — noted as a design
  option in §6 but not prototyped; out of scope for "what exists today".
- **Full content of `docs/retrieval-tools.md` beyond the excerpted sections**
  (it continues past what's quoted here, covering hybrid-mode internals,
  timeouts, etc.) — read in full locally but only the retrieval-scope-defining
  parts are quoted; the rest is genuinely about internal ranking mechanics,
  not relevant to "can this search mc-metarepo".
- **Exact behavior of `search_files` under `.gitignore`** (does it honor
  `.gitignore` like real ripgrep, or walk everything?) — description says
  "Ripgrep-backed" but a direct `subprocess`/`Popen` call to an `rg` binary
  was not found in `file_tools.py` by grep (came back empty); whether it
  shells out (a `subprocess` import may exist elsewhere in the file, not
  isolated) or reimplements ripgrep-like semantics in pure Python was not
  conclusively resolved — flagged as unverified sub-detail, doesn't change
  the top-line finding that the tool exists and works (proven live in §7).
