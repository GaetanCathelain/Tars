# WF6 probe — deterministic rendering of the Linear board block

Measured on the Tars VM (`gaetan@192.168.0.9`, Hermes v0.20.0, clock UTC) on
2026-08-10. Read-only: no config edit, no install, no restart, nothing written
on the VM. No secret was read, printed or persisted; the only credential-shaped
measurement is a *presence count*, never a value.

Question: the 08:30 brief's Linear board must be rendered deterministically, not
LLM-recalled. Can a non-agentic step invoke an MCP tool?

**Answer up front: NO.** The Hermes CLI has no command that calls one MCP tool
with given arguments and prints its result. The fallback (a script + raw Linear
GraphQL) is viable and its crux — does the script actually get `LINEAR_API_KEY`
— is **measured yes, but only on the terminal/shell tool path**; the
`execute_code` sandbox scrubs it.

---

## Q1 — Does the Hermes CLI expose a non-agentic MCP tool call?

### `hermes mcp --help`

```
usage: hermes mcp [-h] [--accept-hooks]
                  {serve,add,remove,rm,list,ls,test,configure,config,login,reauth,picker,catalog,install}
                  ...
Manage MCP server connections and run Hermes as an MCP server. MCP servers
provide additional tools via the Model Context Protocol. Use 'hermes mcp add'
to connect to a new server, or 'hermes mcp serve' to expose Hermes
conversations over MCP.
positional arguments:
  {serve,add,remove,rm,list,ls,test,configure,config,login,reauth,picker,catalog,install}
    serve               Run Hermes as an MCP server (expose conversations to
                        other agents)
    add                 Add an MCP server (discovery-first install)
    remove (rm)         Remove an MCP server
    list (ls)           List configured MCP servers
    test                Test MCP server connection
    configure (config)  Toggle tool selection
    login               Force re-authentication for an OAuth-based MCP server
    reauth              Re-authenticate one OAuth MCP server, or all of them
                        (--all)
    picker              Interactive catalog picker (also the default for
                        `hermes mcp`)
    catalog             List Nous-approved MCPs available for one-click
                        install
    install             Install a catalog MCP by name (e.g. `hermes mcp
                        install n8n`)
```

No `call` / `invoke` / `run` / `exec` / `tool`.

### Every `hermes mcp` subcommand, verbatim

```
=== mcp serve --help ===
usage: hermes mcp serve [-h] [-v] [--accept-hooks]
=== mcp add --help ===
usage: hermes mcp add [-h] [--url URL] [--command MCP_COMMAND] [--args ...]
                      [--auth {oauth,header}] [--preset PRESET]
                      [--connect-timeout CONNECT_TIMEOUT] [--env [ENV ...]]
                      name
=== mcp remove --help ===
usage: hermes mcp remove [-h] name
=== mcp list --help ===
usage: hermes mcp list [-h]
=== mcp test --help ===
usage: hermes mcp test [-h] name
positional arguments:
  name        Server name to test
options:
  -h, --help  show this help message and exit
=== mcp configure --help ===
usage: hermes mcp configure [-h] name
=== mcp login --help ===
usage: hermes mcp login [-h] name
=== mcp reauth --help ===
usage: hermes mcp reauth [-h] [--all] [name]
=== mcp picker --help ===
usage: hermes mcp picker [-h]
=== mcp catalog --help ===
usage: hermes mcp catalog [-h]
=== mcp install --help ===
usage: hermes mcp install [-h] identifier
```

**`hermes mcp test` takes a server name and nothing else** — no tool name, no
arguments. It is a connectivity/handshake probe, not an invoker. This kills the
"some implementations allow it" hypothesis.

### `hermes tools --help` and subcommands

```
usage: hermes tools [-h] [--summary] {list,disable,enable,post-setup} ...
Enable, disable, or list tools for CLI, Telegram, Discord, etc. Built-in
toolsets use plain names (e.g. web, memory). MCP tools use server:tool
notation (e.g. github:create_issue). Run 'hermes tools' with no subcommand for
the interactive configuration UI.

=== tools list --help ===
usage: hermes tools list [-h] [--platform PLATFORM]
=== tools disable --help ===
usage: hermes tools disable [-h] [--platform PLATFORM] NAME [NAME ...]
=== tools enable --help ===
usage: hermes tools enable [-h] [--platform PLATFORM] NAME [NAME ...]
=== tools post-setup --help ===
usage: hermes tools post-setup [-h] KEY
Run the install/bootstrap hook a tool backend declares … Keys: agent_browser,
camofox, cua_driver, kittentts, piper, ddgs, spotify, langfuse, xai_grok.
```

`hermes tools` is a *configuration* surface (enable/disable per platform). It
never executes a tool. `server:tool` notation appears only as a config key.

### Other plausible surfaces (help only, nothing started)

- `hermes mcp serve` — "Run Hermes as an MCP server (expose conversations to
  other agents)". Wrong direction: it makes Hermes an MCP *server*, it does not
  call another server's tool.
- `hermes proxy {start,status,providers}` — "local HTTP server that forwards
  OpenAI-compatible requests to an OAuth-authenticated provider". Inference
  proxy, not MCP.
- `hermes acp` — ACP server for editors (VS Code, Zed, JetBrains). Agent loop.
- `hermes serve` — "the JSON-RPC/WebSocket gateway the desktop app and remote
  clients connect to". Long-lived backend; drives the agent, not a raw tool.
- `hermes debug {share,delete}` — uploads a debug report.
- `hermes dump [--show-keys]` — plain-text setup summary.
- `hermes console` — "Open a curated Hermes command REPL. This is not a raw
  shell and does not expose the full Hermes CLI."
- `hermes send` — delivers text to a messaging platform. Useful as an *output*
  step; it fetches nothing.
- `hermes -z/--oneshot PROMPT`, `hermes chat -Q -q` — full LLM agent loop.

### Verdict

**NO.** There is no command anywhere in `hermes` v0.20.0 that invokes one named
MCP tool with JSON arguments and prints its raw result outside the LLM agent
loop. Every MCP tool call on this box goes through the model.

Corollary worth recording: the standing "Linear access must be Hermes-native"
rule *cannot* be satisfied by a deterministic step. `mcp__linear__*` exists only
inside the agent loop, which is exactly the thing that fabricated the rows.

## Q2 — Exercising it

Not applicable: Q1 is No. Nothing to exercise, nothing to diff.

## Q3 — The fallback, established precisely

### 3.1 Python

```
$ python3 -V
Python 3.12.3
```

(System Python. The Hermes gateway venv is `~/.hermes/hermes-agent/venv`, whose
`bin` sits first on the child `PATH`, so bare `python` inside a tool subprocess
resolves to the venv interpreter — the existing precedent below invokes
`python`, not `python3`. Both are irrelevant to correctness here: the collector
uses only stdlib `urllib`/`json`/`ssl`.)

### 3.2 `LINEAR_API_KEY` reaches a subprocess — measured, with a caveat that matters

`~/.hermes/.env` variable **names** (no values read):

```
GMAIL_ADDRESS
GMAIL_APP_PASSWORD
HINDSIGHT_MODE
LINEAR_API_KEY
NOTION_API_TOKEN
SLACK_ALLOWED_USERS
SLACK_APP_TOKEN
SLACK_BOT_TOKEN
SLACK_HOME_CHANNEL
SLACK_MCP_XOXC_TOKEN
SLACK_MCP_XOXD_TOKEN
```

`.env` reaches the process environment: `gateway/run.py:1835`, `run_agent.py:129`
and `cron/` all call `hermes_cli/env_loader.load_hermes_dotenv(hermes_home=…)`,
which loads `~/.hermes/.env` into `os.environ` with `override=True`. So the
gateway, the CLI and cron runs all carry `LINEAR_API_KEY` in `os.environ`.
(It will not show in `/proc/<pid>/environ` — same post-exec `os.environ` write
as the `SLACK_*` vars.)

But Hermes **strips secrets from tool child processes by default**, and the two
sandboxes strip differently. From `tools/env_passthrough.py`:

```
Skills that declare ``required_environment_variables`` in their frontmatter
need those vars available in sandboxed execution environments (execute_code,
terminal).  By default both sandboxes strip secrets from the child process
environment for security.  This module provides a session-scoped allowlist
so skill-declared vars (and user-configured overrides) pass through.
```

- **terminal backend** (`tools/environments/local.py`, `_make_run_env` /
  `_sanitize_subprocess_env`) strips only `_HERMES_PROVIDER_ENV_BLOCKLIST`
  (derived from the provider registry + `OPTIONAL_ENV_VARS` entries whose
  `category` is `tool` or `messaging`) plus dynamically-named Hermes-internal
  secrets (`AUXILIARY_*_API_KEY`, `GATEWAY_RELAY_*_SECRET|KEY|TOKEN`).
  `LINEAR_API_KEY` is registered in `hermes_cli/config_defaults.py` as
  `"category": "skill"`, so it is **not** on that blocklist (`grep -n LINEAR
  tools/environments/local.py` → no match).
- **`execute_code` sandbox** (`tools/code_execution_tool.py`) scrubs by
  *substring*: `_SECRET_SUBSTRINGS = ("KEY", "TOKEN", "SECRET", "PASSWORD",
  "CREDENTIAL", … "CREDS", "BEARER", "APIKEY")`, applied in `_scrub_child_env`.
  `LINEAR_API_KEY` contains `KEY` → **scrubbed**, unless explicitly registered
  via `env_passthrough`.

Measured on the VM (single `hermes chat -Q -q` run, prompt asked for a *count*
and a *word*, never a value):

```
session_id: 20260810_202200_431a4d
terminal=1 execute_code=absent
```

i.e. `env | grep -c '^LINEAR_API_KEY='` printed `1` inside the terminal tool,
and `'LINEAR_API_KEY' in os.environ` was **False** inside `execute_code`.

**The script approach works — on the terminal/shell tool path only.** The VM has
`terminal.backend: local` (`~/.hermes/config.yaml`) and no
`terminal.env_passthrough` list.

Two consequences to write down:

1. Any deterministic collector must be run through the **shell/terminal tool**
   (`python3 /path/script.py …`), never `execute_code`.
2. Belt and braces: declare `required_environment_variables: [LINEAR_API_KEY]`
   in the skill frontmatter. `LINEAR_API_KEY` is not a Hermes-managed provider
   credential, so `register_env_passthrough` accepts it (the
   GHSA-rhgp-j443-p4rf blocklist only refuses Hermes provider keys), and the
   var then survives in both sandboxes.

**Existing-skill defect found while measuring.** `engagement-checker` §4 says the
collector uses "the inherited `LINEAR_API_KEY`", and its frontmatter (v1.6.0)
declares **no** `required_environment_variables`:

```yaml
---
name: engagement-checker
description: "Use for incremental follow-up and commitment reminders."
version: 1.6.0
metadata:
  hermes:
    tags: [engagement, reminders, slack, email, linear, orchestration]
    category: orchestration
---
```

So that collector silently depends on the agent choosing the terminal tool; if a
run picks `execute_code`, `os.environ["LINEAR_API_KEY"]` raises `KeyError` and
the skill's own §4 guard (`RuntimeError("LINEAR_API_KEY is unavailable")`) fires.
Worth a one-line frontmatter fix in the same PR.

### 3.3 Script on disk — the precedent

Yes, a skill can instruct running a script file, and the convention is
`~/.hermes/skills/<category>/<skill>/scripts/<name>.py`. 21 skills ship a
`scripts/` directory; `engagement-checker` already has
`~/.hermes/skills/orchestration/engagement-checker/scripts/` (currently only an
empty `tests/`).

The precedent invocation, verbatim from `engagement-checker` §3:

```bash
GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
$GAPI gmail search "after:YYYY/MM/DD" --max 100
```

That is the pattern to copy: an `HERMES_HOME`-relative absolute path, a
sub-command + flags argv, run through the shell tool, output consumed as text.
Note the credential never appears on argv — `google_api.py` reads its own
OAuth state, exactly as the board script must read `os.environ["LINEAR_API_KEY"]`.

## Q4 — Determinism of inclusion, and the size limits

Hermes hands a tool result to the model as ordinary message text. There is **no
primitive that injects a tool result verbatim into the reply** — the agent
retypes it, which is precisely the failure mode being fixed. The mechanism
therefore buys *correct data*, and the skill wording plus block size must buy
*correct copying*.

Three measured limits shape that:

1. **Terminal-tool stdout cap — 50,000 chars.**
   `tools/tool_output_limits.py`: `DEFAULT_MAX_BYTES = 50_000` (=
   `terminal_tool.MAX_OUTPUT_CHARS`), overridable via `tool_output.max_bytes` in
   `config.yaml`. The VM config has **no** `tool_output` section, so 50,000 is
   live. Over the cap, `terminal_tool.py:3155-3163` keeps **40 % head + a
   truncation notice + 60 % tail** — the *middle* is deleted, which would
   silently amputate board rows without any visible row-count error. A board
   block is ~1–3 KB, so this is a wide margin, but the collector must never be
   asked to dump a full issue list.

2. **Per-result persistence threshold — 100,000 chars; per-turn budget —
   200,000 chars; preview — 1,500 chars.**
   `tools/budget_config.py`:
   ```
   DEFAULT_RESULT_SIZE_CHARS: int = 100_000
   DEFAULT_TURN_BUDGET_CHARS: int = 200_000
   DEFAULT_PREVIEW_SIZE_CHARS: int = 1_500
   ```
   Past the per-result threshold, `tools/tool_result_storage.maybe_persist_tool_result`
   replaces the result with a **1,500-char preview plus a spill-file handle** —
   the agent then sees a truncated block and a path. If that ever happened to the
   board block, byte-for-byte inclusion becomes impossible by construction.
   (These scale down for small-context models: 15 % of the window per result,
   30 % per turn, floors 8,000 / 16,000 chars.)

3. **MCP tool results are not capped like terminal output.** `tools/mcp_tool.py`
   has no char cap on a tool result (only `_MCP_RESOURCE_MAX_BYTES = 50 MB` for
   binary resources), so an MCP result lands unabridged against the 100 K /
   200 K budgets — consistent with the previously observed 399,510-char
   tool-result blowout. A shell script that prints exactly the rows it was asked
   for is *structurally* smaller and more predictable than any
   `mcp__linear__list_issues` response.

Practical implication: the collector should print **only** the finished markdown
block (nothing else on stdout), keeping the result 2–3 orders of magnitude under
every limit above, and the skill should fence it and require a byte-for-byte
copy. A trailing machine-checkable marker (`rows:<n>`) that the agent must copy
along with the block is a cheap post-hoc detector of a mangled paste; skip it if
the reviewer thinks it clutters the brief.

---

## Recommended mechanism

**A Python script on disk, run through the shell/terminal tool, that queries
Linear's GraphQL API directly and prints the finished board block verbatim to
stdout; the agent includes that stdout byte-for-byte and narrates around it.**

Concretely:
`~/.hermes/skills/orchestration/daily-work-brief/scripts/linear_board.py`,
invoked with the engagement-checker precedent
(`python ${HERMES_HOME:-$HOME/.hermes}/skills/.../scripts/linear_board.py --team GCN`),
reading `os.environ["LINEAR_API_KEY"]` — never argv, never `curl -H`, and mirroring
the engagement-checker collector's guards (fixed endpoint, hardcoded read-only
query, `mutation` rejected, bounded response size).

Evidence for it:

- Q1: no non-agentic MCP invocation exists in `hermes` v0.20.0 — `mcp test`
  takes a server name only; `tools` only enables/disables; `mcp serve` /
  `acp` / `serve` / `proxy` are the wrong direction or the agent loop.
- Q3: `LINEAR_API_KEY` is measurably present in a terminal-tool subprocess
  (`terminal=1`) and measurably absent in `execute_code` (`execute_code=absent`),
  so the script gets its credential on the shell path without the key ever
  touching argv, a log, or an evidence file.
- Q3: the script-on-disk pattern is already the house style
  (`google-workspace/scripts/google_api.py`, 21 skills with `scripts/`).
- Q4: a script's stdout is small and bounded by construction, well inside the
  50 K terminal cap and the 100 K/200 K budgets that an uncapped MCP result can
  blow through.

### Trade-off against the native-transport rule

This **requires raw HTTP** and is a documented, measured exception to
"Linear access should be Hermes-native (`mcp__linear__*`)". The rule and the
coordinator ruling are in direct conflict here: native Linear access exists only
inside the LLM agent loop, and the agent loop is what fabricated rows under
compaction. Determinism wins for the board; nothing else changes.

The split should be identical in shape to the one `engagement-checker` §"Load
only what is needed" already documents and calls load-bearing:

> **Linear transport is mixed on purpose. Delta scan: the audited raw collector
> in §4, coverage-gated. Everything else: native tools.**

For `daily-work-brief`: **the board block only** goes over raw GraphQL; every
other Linear read, and every Linear write, stays on `mcp__linear__*` per
`linear-ticketing`.

### What must be documented if raw HTTP is used (it is)

1. The measured limitation, in the skill itself: no `hermes` command invokes an
   MCP tool outside the agent loop (`hermes mcp` has no `call`; `mcp test` takes
   only a server name) — so a deterministic step cannot use `mcp__linear__*`.
   Cite this file.
2. Scope fence: raw HTTP covers the board block and nothing else; all other
   Linear reads/writes remain native. No write path in the script, `mutation`
   rejected, fixed endpoint `https://api.linear.app/graphql`.
3. Credential handling: `os.environ["LINEAR_API_KEY"]`, never argv, never
   `curl -H "Authorization: $VAR"`, never printed/logged/hashed/persisted —
   same wording as `engagement-checker` §4.
4. Execution path: **shell/terminal tool, not `execute_code`** (measured: the
   `execute_code` sandbox scrubs any var whose name contains `KEY`), plus
   `required_environment_variables: [LINEAR_API_KEY]` in the frontmatter so the
   dependency is declared rather than implicit.
5. Inclusion contract: the script prints only the block; the agent reproduces it
   byte-for-byte and may not re-derive, re-order, re-title, or summarise rows;
   any "+N more" is emitted by the script or not at all.
6. Revisit trigger: if a future Hermes version ships a non-agentic MCP tool
   invocation, this exception is retired and the block moves to `mcp__linear__*`.

### Rejected alternatives

- **`mcp__linear__*` inside the agent loop.** The thing being fixed.
- **Speak MCP directly to `https://mcp.linear.app/mcp` from the script**
  (the server is a remote HTTP MCP with a static `Authorization` header in
  `config.yaml`). Technically possible and marginally more "native", but it is
  still raw HTTP — with an added `initialize` handshake, session id, and SSE
  framing to parse — and it would pull a second copy of a credential out of
  `config.yaml`. Strictly worse than GraphQL for the same guarantee.
- **A `hermes cron` job writing the block to a file for the brief to read.**
  Adds a second scheduled component and a staleness window for zero gain; the
  brief's own shell step is one call.
