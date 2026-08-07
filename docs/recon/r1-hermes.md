# R1 — Hermes platform deep-dive

WF1 recon probe. Read-only throughout: no installs, no config changes, no VM touched except
read-only inspection of `~/.hermes` on cooper (a stub/leftover, see below). Personal Hermes
(p-Hermes) itself was **not** probed live here — R3 owns that; this doc leans on the
already-ingested `learnings/hermes-vm.md` for p-Hermes facts and on official docs
(`hermes-agent.nousresearch.com`, `github.com/NousResearch/hermes-agent`) for the install/config
mechanics lane B needs.

## Verdict

Hermes Agent (NousResearch, current pinned release **v0.20.0 / tag `v2026.8.3`**, published
2026-08-03) is a well-documented, actively maintained (226,961 GH stars) CLI-first agent
framework with first-class multi-profile support, a systemd-service-per-profile gateway model,
native A2A v1.0 (shipped in this exact release), and a plugin/skill ecosystem that covers all
four requested add-ons (rtk, hindsight, hermes-lcm, i-have-adhd) via three *different*
mechanisms (native CLI hook, bundled memory provider, third-party plugin repo, third-party skill
tap, respectively). Lane B has everything it needs to script the install unattended. The one
real design decision left open — not a blocker — is *how* "Tars is the default profile" gets
realized (see Open questions).

## Facts with evidence

### What Hermes is / how it's installed

- Open-source agent framework, repo `NousResearch/hermes-agent`, "The agent that grows with
  you". `gh api repos/NousResearch/hermes-agent` → 226,961 stars, default branch `main`, last
  push 2026-08-07. Latest 3 releases (`gh api .../releases`): `v2026.8.3` "v0.20.0 — " (2026-08-03),
  `v2026.7.30` "v0.19.1" (2026-07-30), `v2026.7.20` "v0.19.0 — The Quicksilver Release"
  (2026-07-20). **A2A shipped in this exact pinned range**: GH issue #514 "Feature: A2A protocol
  support" was closed by PR #77109 "feat(a2a): Agent-to-Agent protocol plugin, A2A v1.0 (closes
  #514)" — consistent with the v0.20.0 changelog.
- Linux install (official docs, `/docs/getting-started/installation`):
  ```
  curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash
  ```
  Prereqs: git, curl, xz-utils (`sudo apt install curl xz-utils`), build-essential for the
  desktop app. Installer provisions **uv** (Python pkg mgr), **Python 3.11** (via uv, no sudo),
  **Node.js v22**, **ripgrep**, **ffmpeg** — no manual dependency install needed.
- Layout: source `~/.hermes/hermes-agent/`, binary symlink `~/.local/bin/hermes`, config/data
  `~/.hermes/`.
- First run: `source ~/.bashrc && hermes setup --portal` — one command that logs in, sets Nous
  as provider, enables the Tool Gateway. Or granular: `hermes model`, `hermes tools`,
  `hermes config set`. Three setup modes exist: Quick (Nous Portal OAuth, no API keys), Full
  (manual provider config), Blank Slate (minimal, opt-in features).
- Platform support (`/docs/getting-started/platform-support`, verbatim): *"We test on the
  latest Ubuntu and WSL2. If your distro has glibc, systemd, and follows the Filesystem
  Hierarchy Standard, it's likely to work pretty well."* — **systemd is a stated dependency**,
  matches the gateway-as-systemd-service model below. Ubuntu with GUI (as the pitch specifies)
  is squarely in the tested target.
- `hermes doctor` (diagnose config/deps), `hermes status` (agent/auth/platform status), `hermes
  dump` (copy-pasteable support summary) are the built-in smoke-check commands — natural fit
  for lane B's terminal "smoke check → verdict" step.

### Profiles — creation, default, system prompt

- `~/.hermes` itself **is** the default profile; there is no separate "default" directory —
  it's the base install. Named profiles live at `~/.hermes/profiles/<name>/`.
- `hermes profile` subcommands (docs `/docs/user-guide/profiles`): `create <name>`, `list`,
  `show <name>`, `rename <old> <new>`, `delete <name>`, `use <name>` (sticky default —
  *"like `kubectl config use-context`"*), `use default` (revert to base), `export`/`import`
  (`.tar.gz`), `describe`.
- `-p <name>` / `--profile=<name>` targets any profile per-invocation on any `hermes` command.
  `hermes profile use <name>` also drops a command alias at `~/.local/bin/<name>` — e.g. after
  `hermes profile use tars`, `tars chat` / `tars gateway start` work with no flag.
- **`HERMES_HOME`** is the real mechanism underneath all of this: it points at a profile's home
  dir; unset → `~/.hermes`. Confirmed independently by the p-Hermes memory-ingestion note
  (`learnings/hermes-vm.md`, 2026-07-29, live VM read): `$HERMES_HOME` is **not exported in any
  shell**, only set per-unit inside each `~/.config/systemd/user/hermes-gateway*.service`; a
  bare `runuser` shell sees it empty and `$HERMES_HOME/...` paths silently collapse to `/...`.
  On p-Hermes: default profile = `/home/hermes/.hermes`; the (to-be-deleted) `tars` profile =
  `/home/hermes/.hermes/profiles/tars`.
- System prompt / identity = **`SOUL.md`** at `$HERMES_HOME/SOUL.md` (docs
  `/docs/user-guide/features/personality`). Slot #1 in the system prompt — injected raw (after
  prompt-injection scanning + size truncation), before tool guidance / memory / skills / other
  context files. Falls back to a built-in default identity if missing/empty. Full precedence
  stack: SOUL.md → tool-aware guidance → memory/user context → skills guidance → context files
  (AGENTS.md, `.cursorrules`) → runtime `/personality <name>` overlays (`/personality none`
  resets to base SOUL.md). SOUL.md is for durable persona/tone; project-specific instructions
  belong in AGENTS.md instead — direct implication for the Tars profile draft: "no coding, no
  PRs, orchestrate/report only" belongs in SOUL.md as identity, not as a context file.
- Config precedence generally: CLI args > `~/.hermes/config.yaml` (non-secrets: model, terminal
  backend, compression, toolsets) > `~/.hermes/.env` (secrets: API keys, bot tokens) > built-in
  defaults. `${VAR}` substitution works in config.yaml. `hermes config {edit,set,get,check}`.
- Everything profile-scoped: config.yaml, .env/auth.json, sessions, memories, skills, plugins,
  toolsets, platform credentials — each profile is a fully isolated Hermes instance sharing only
  the install.

### Plugins/skills — where rtk, hindsight, hermes-lcm, i-have-adhd come from

Four different install mechanisms — confirmed this is deliberate, not sloppy naming in the pitch:

1. **rtk** — already installed on cooper (`brew info rtk` → `rtk 0.44.1 → 0.44.2`, from
   `Homebrew/homebrew-core`, upstream `https://www.rtk-ai.app/`, Apache-2.0). It is **not** a
   Hermes plugin repo — it's a standalone Rust CLI-proxy binary with a generic
   hook-installer subcommand: `rtk init --agent hermes` is a **first-class supported target**
   (confirmed via `rtk init --help` → `--agent <AGENT>` accepts `hermes` alongside `claude`,
   `cursor`, `windsurf`, `cline`, etc.). So for Tars: `brew install rtk` (or the RTK install
   script) then `rtk init --agent hermes` (or `-g --agent hermes` for a global/all-profiles
   hook) — same shape as the existing Claude Code hook on cooper (`~/.claude/RTK.md`). Config
   lives at `~/.config/rtk/{config.toml,filters.toml}` (user-global, not per-Hermes-profile).
2. **hindsight** — a **bundled** external memory provider, not a separate repo to fetch. Docs
   (`/docs/user-guide/features/memory`): *"Hermes ships with 8 external memory provider
   plugins — including Honcho, OpenViking, Mem0, Hindsight, Holographic, RetainDB, ByteRover,
   and Supermemory."* Setup: `hermes memory setup` (interactive, pick "hindsight") or manual:
   `hermes config set memory.provider hindsight` + `.env` keys. **Fully local mode exists**
   (matches pitch: "local, basic"): `HINDSIGHT_MODE=local` + an LLM API key
   (`llm_provider`/`llm_api_key`) — no cloud account, no `HINDSIGHT_API_KEY` needed; an embedded
   Postgres daemon starts automatically on first use (~1 min cold start). Config at
   `~/.hermes/hindsight/config.json` (env vars override). Verify with `hermes memory status`.
   This runs **alongside**, not instead of, the builtin `MEMORY.md`/`USER.md` (2,200/1,375-char
   files in `$HERMES_HOME/memories/`, frozen snapshot injected at session start).
3. **hermes-lcm** — third-party GitHub repo, confirmed: `stephenschoettler/hermes-lcm`,
   "Lossless Context Management plugin for Hermes Agent — DAG-based context engine that never
   loses a message." It's a **context-engine plugin** (`plugins/context_engine/<name>/`
   discovery path per the plugin docs) — these are single-select via `context.engine` in
   config.yaml and **never auto-activate**. Install per its own README: clone/checkout under
   the target profile's plugin dir, e.g. `git clone https://github.com/stephenschoettler/hermes-lcm
   ~/.hermes/profiles/<name>/plugins/hermes-lcm` (or `~/.hermes/plugins/hermes-lcm` for the
   default/base profile) then run its `./scripts/install.sh` (profile-aware via `HERMES_PROFILE=
   <name> ./scripts/install.sh`); then set `context.engine: lcm` in that profile's config.yaml.
   Generic Hermes plugin install also applies: `hermes plugins install stephenschoettler/hermes-lcm
   --enable`.
4. **i-have-adhd** — third-party GitHub repo `ayghri/i-have-adhd`, "A skill to stop your coding
   agent from burying the answer. ADHD-friendly output." This is a **skill**, not a plugin —
   installed through the skills-tap mechanism: `hermes skills tap add ayghri/i-have-adhd` then
   `hermes skills install ayghri/i-have-adhd/skills/i-have-adhd` (or the one-shot
   `hermes skills install ayghri/i-have-adhd/skills/i-have-adhd`, which auto-registers the tap).
   Exposed as `/i-have-adhd` at next session start; alternatively its behavior can be folded
   permanently into AGENTS.md or the profile's SOUL.md so it's always-on without invoking the
   slash command.

General plugin mechanics (docs `/docs/user-guide/features/plugins`, cross-checked against a
real installed plugin on cooper — see next section): manifest `plugin.yaml` (`name`, `version`,
`description`, optional `requires_env`), Python `__init__.py` with `register(ctx)` calling
`ctx.register_hook(event, fn)` / `register_tool` / `register_command` / `register_skill` / etc.
Discovery paths: bundled (`<repo>/plugins/`), user (`~/.hermes/plugins/`), project
(`./.hermes/plugins/`, needs `HERMES_ENABLE_PROJECT_PLUGINS=true`), pip entry points, NixOS
module. General (non-infrastructure) plugins are opt-in via `config.yaml`:
```yaml
plugins:
  enabled: [my-tool-plugin]
  disabled: [noisy-plugin]
```
CLI: `hermes plugins {list,install user/repo [--enable|--no-enable],update,remove,enable,disable}`.

**Live evidence of the plugin shape**, found on cooper at `~/.hermes/plugins/orca-status/`
(this `~/.hermes` on cooper is a **minimal stub** — only `cache/`, `mnemosyne/` (backs the
`mcp__mnemosyne__*` memory MCP tools available in this session, unrelated to Hindsight),
`plugins/orca-status/`, and a 3-line `config.yaml` with `plugins.enabled: [orca-status]` — it is
**not** a full Hermes profile, just evidence Orca auto-drops this hook plugin into any
Hermes-managed session dir it supervises on this machine):
  - `plugin.yaml`: `name: orca-status`, `kind: standalone`, `provides_hooks:` listing all 9
    lifecycle hooks (`on_session_start`, `pre_llm_call`, `post_llm_call`, `pre_tool_call`,
    `post_tool_call`, `pre_approval_request`, `post_approval_response`, `on_session_end`,
    `on_session_finalize`, `on_session_reset`) plus a header comment `# Managed by Orca. Do not
    edit; changes may be overwritten.`
  - `__init__.py`: `register(ctx)` loops the hook list calling `ctx.register_hook(event,
    handler)`; each handler POSTs a size-bounded JSON payload to
    `http://127.0.0.1:{ORCA_AGENT_HOOK_PORT}/hook/hermes` with header
    `X-Orca-Agent-Hook-Token`, all params read from env (`ORCA_PANE_KEY`,
    `ORCA_AGENT_HOOK_TOKEN`, `ORCA_AGENT_LAUNCH_TOKEN`, `ORCA_TAB_ID`, `ORCA_WORKTREE_ID`) or a
    file at `$ORCA_AGENT_HOOK_ENDPOINT`. Fire-and-forget (0.75s timeout, swallows errors) — this
    is Orca's generic cross-agent status-reporting shim, confirming Orca already knows how to
    supervise a Hermes session the same way it supervises Claude Code, which is directly
    relevant if lane B/Orca wants status visibility into the Tars gateway.

### A2A (agent-to-agent) support

Native, shipped in the pinned release, bidirectional (docs
`/docs/user-guide/messaging/a2a` + GH issue #514 / PR #77109):
- Protocol: A2A v1.0 (Linux Foundation-stewarded, Apache 2.0, complementary to MCP — *"MCP
  answers 'what tools can I use?', A2A answers 'who can help me?'"*). 3-layer: protobuf data
  model → abstract ops → JSON-RPC 2.0/HTTP bindings (+ optional gRPC, SSE streaming).
- Enable: `hermes gateway setup` → pick A2A, or directly in config.yaml:
  ```yaml
  gateway:
    platforms:
      a2a:
        enabled: true
        extra: { port: 9900 }
  ```
  The **outbound** toolset (calling other agents) ships disabled by default — enable via
  `hermes tools`.
- Env vars: `A2A_BEARER_TOKEN`, `A2A_PEER_TOKENS` (`name:token,…`), `A2A_HOST` (default
  `127.0.0.1`), `A2A_PORT` (default `9900`), `A2A_AGENT_NAME`, `A2A_PUBLIC_URL`,
  `A2A_TRUSTED_PEERS`, `A2A_RATE_LIMIT` (default 60/min), `A2A_MAX_PINGPONG_TURNS` (default 5,
  max 20 — anti-loop cap), `A2A_REPLY_TIMEOUT` (default 300s).
- Discovery: `GET /.well-known/agent-card.json` (canonical v1.0 path, legacy `agent.json` also
  served). Known peers declared in config.yaml under `a2a_agents.<name>: {url, auth, timeout,
  capabilities}`.
- Outbound tools once enabled: `a2a_discover(url)`, `a2a_call(agent, message, context_id?)`,
  `a2a_list()`, `a2a_history(context_id)`, `a2a_orchestrate(capability, message, mode?)` (fan
  out to peers advertising a capability, `all`/`first`/`best`).
- Inbound serving: Agent Card + JSON-RPC 2.0 `POST /` (`SendMessage`, `SendStreamingMessage`,
  `GetTask`, `ListTasks`, `CancelTask`, `SubscribeToTask`), SSE streaming, HMAC-SHA256-signed
  push-notification webhooks. Inbound tasks inject into the **live gateway session** — shares
  memory/tools across all channels.
- Security, secure-by-default: no token → localhost-only bind; remote exposure needs a bearer
  token **and** explicit `A2A_HOST`; remote peers **cannot invoke slash commands**;
  credential-shaped strings are scrubbed from outbound replies; every exchange logged to
  `~/.hermes/a2a_audit.jsonl`.
- **Not profile-scoped** — one gateway session's A2A surface is shared across all its channels,
  unlike Slack/Telegram/etc. which are strictly per-profile.
- Quick test: `curl http://host:9900/.well-known/agent-card.json` and a `SendMessage` JSON-RPC
  POST — directly usable as a WF4 verify probe once Tars is live.
- Pitch's "hermes chat CLI ... or using A2A protocol" reading confirmed: `hermes chat` is the
  interactive/one-shot local CLI entry point (separate feature, not A2A); A2A is the
  machine-to-machine HTTP surface. Both are valid ways to "talk to Tars" programmatically for
  WF3/WF4 probes — A2A additionally lets **other agents** (e.g. an Orca-run Claude session) call
  Tars as a tool, which is relevant to "orchestrate work on Cooper VM" from the pitch.

### How a Slack app attaches to a profile

Docs `/docs/user-guide/messaging/slack` + `/docs/user-guide/messaging/` (general) +
`/docs/user-guide/multi-profile-gateways`:
- Hermes uses **Socket Mode** (Slack Bolt SDK, WebSocket) — no public URL/reverse proxy needed,
  works behind a firewall or on a plain VM. Needs two tokens: Bot Token (`xoxb-`) and App-Level
  Token (`xapp-`, scope `connections:write`).
- App creation: `hermes slack manifest --agent-view --write` generates
  `~/.hermes/slack-manifest.json` for **Create New App → From an app manifest** at
  api.slack.com/apps — or manual setup with a documented scope list (`chat:write`,
  `app_mentions:read`, `channels:history`, `channels:read`, `groups:history`, `im:history`,
  `im:read`, `im:write`, `mpim:history`, `mpim:read`, `users:read`, `files:read`,
  `files:write`; docs flag `channels:history`/`groups:history` as the #1 missed-setup gotcha —
  without them the bot only ever sees DMs). Also needs Event Subscriptions
  (`message.im`/`message.mpim`/`message.channels`/`message.groups`/`app_mention`) and, easy to
  miss, **App Home → Messages Tab ON** + "allow users to send messages" — otherwise Slack shows
  "Sending messages to this app has been turned off" on DM attempts.
- Binding to a profile is purely **credential placement**: tokens go in that profile's
  `.env` (`$HERMES_HOME/.env` → `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`, `SLACK_ALLOWED_USERS`
  comma-separated Slack member IDs, optional `SLACK_HOME_CHANNEL`/`SLACK_HOME_CHANNEL_NAME`).
  **One token pair can only be "live" on one profile's gateway at a time** — the docs are
  explicit: *"if two profiles share a Telegram, Discord, Slack, WhatsApp, or Signal token, the
  second gateway refuses to start with an error naming the conflicting profile."* This is the
  hard technical reason the pitch's plan (delete the old p-Hermes `tars` profile before
  attaching the same Slack app on the new VM) is **not optional politeness** — it's a load-bearing
  constraint: the new profile's gateway will not start against a still-live old one holding the
  same token. Matches PLAN.md's gated-cutover design exactly.
- **Reuse path for an already-existing Slack app** (exactly the pitch's scenario): copy the
  existing `xoxb-`/`xapp-` tokens into the new profile's `.env`; no Slack-side reinstall or
  re-manifest needed — scopes/events/Socket-Mode config live in Slack itself and are shared;
  only `SLACK_ALLOWED_USERS` may need updating per profile.
- Bot does not auto-join channels — `/invite @<app-name>` is required per channel even after the
  gateway is running; relevant since the pitch wants Tars to see "all channels I have access
  to" — that access has to be granted channel-by-channel post-cutover, or via being invited by
  an admin, not implied by DM setup alone.
- Access control beyond tokens: `SLACK_ALLOWED_USERS` (flat allowlist) or
  `GATEWAY_ALLOW_ALL_USERS=true` (docs discourage this for a bot with terminal access); DM
  pairing flow (`hermes pairing approve/list/revoke slack <code|user-id>`) as an alternative to
  manually listing user IDs; admin tiering via
  `gateway.platforms.slack.{allow_admin_from, user_allowed_commands}` in config.yaml — directly
  useful for enforcing "responds only to Gaetan" (WF4's negative test) at the platform-scoping
  layer, on top of whatever SOUL.md says.
- Prior-art corroboration from `learnings/hermes-vm.md` (live p-Hermes read, 2026-07-29): the
  current `tars` bot there already runs "least-privilege scope (owner-only DMs + invited-channel
  mentions; no local network, credentials, shell/files or Home Assistant)" — i.e. today's
  Slack-side scoping on the app to be reused is already DM+invited-channel shaped, which lines
  up with what the new profile needs.

### Persistence — gateway as a systemd service

- Every profile that runs a messaging platform gets **its own systemd user service**:
  `~/.config/systemd/user/hermes-gateway-<profile>.service` (Linux) /
  `~/Library/LaunchAgents/ai.hermes.gateway-<profile>.plist` (macOS). Created automatically by
  a `<profile> gateway install`-style command (exact subcommand not fully quoted by the docs
  fetch — verify with `hermes gateway --help` during actual install, flagged below).
  Auto-restarts on crash and on login.
- **Directly corroborated on the actual p-Hermes VM** by `learnings/hermes-vm.md`: units matched
  glob `~/.config/systemd/user/hermes-gateway*.service`, and critically, `$HERMES_HOME` is set
  **inside the unit file**, not exported anywhere else — this is *the* mechanism that lets one
  systemd instance run one profile's gateway with correct isolation.
- Cross-VM confirmation from a different Hermes deployment (mc-agents `support-engineer`,
  `~/dev/se-nmc496/systemd/hermes-gateway-support-engineer.service` +
  `hermes-linear-shim.service` exist as real unit files in that repo) — same naming convention,
  independent deployment, same platform.
- For the units to survive SSH logout/reboot without an active login session:
  `sudo loginctl enable-linger "$USER"` (user lingering) — required on a headless-ish VM unless
  someone stays logged in via the GUI session.
- A wrapper is mentioned (`hermes-gateways start`) that launches every configured profile's
  gateway via its service manager in one shot — worth checking for on the actual install; not
  independently verified.

### /sethome

- **Messaging-only** slash command (works in-chat on Slack/Telegram/Discord/WhatsApp/
  Signal/Matrix/Mattermost — NOT a CLI command). Alias `/set-home`. Verbatim doc quote: *"Mark
  the current chat as the platform home channel for deliveries."*
- Purpose: designates the default delivery location for automated job results / session
  handoffs on that platform. Example dependency: `/handoff` (CLI→messaging session transfer)
  requires "a home channel configured for the target platform" to succeed.
- **Directly satisfies the pitch's requirement** ("`/sethome` should be a DM conversation
  between me and him"): once the Slack app is attached and Gaetan DMs the bot, running
  `/sethome` in that DM thread sets it as Tars's home channel — this is the literal mechanism
  for "cron/scheduled reports land in my DM, not some channel." No config-file equivalent is
  documented; it must be run from inside the live chat, which puts it at the gated-cutover step
  in PLAN.md, not something lane B can pre-script headlessly (it needs Gaetan's own DM to exist
  as "the current chat").
- Related commands surfaced alongside it in the slash-command reference: `/platform` (list/
  pause/resume gateway adapters), `/profile` (show active profile name + home dir), `/whoami`
  (show the caller's access tier — admin vs user), `/topic` (Telegram multi-session mode).

## Blockers

None for recon itself. For lane B execution, none of the above blocks scripting the install —
everything needed (install script, profile mechanics, plugin/skill install commands, systemd
unit pattern, A2A config, Slack token placement) is documented and cross-corroborated against a
real deployment. The one true blocker is out of R1's scope: `/sethome` cannot be run until
Gaetan's Slack DM with the (cutover-attached) app exists — that's already correctly placed at
the gated cutover step in PLAN.md, not in lane B.

## Open questions

1. **How literally should "Tars is the default profile" be read?** Two valid readings:
   (a) configure the **base** `~/.hermes` directly as Tars — no `hermes profile create` at all,
   since the base install already *is* the default profile and this VM will run nothing else;
   or (b) `hermes profile create tars && hermes profile use tars`, which makes a `tars` alias
   sticky-default but still physically lives under `~/.hermes/profiles/tars/`. (a) is the
   simpler, more literal match to the pitch and to how p-Hermes's *default* profile already
   works (`/home/hermes/.hermes`, no `profiles/` subpath) — recommend lane B do (a) unless a
   reason to keep `~/.hermes` as an empty base + named profile surfaces later (e.g. wanting a
   second throwaway profile for testing without touching Tars's live state).
   **Not yet resolved — needs a decision before lane B writes SOUL.md/config.yaml.**
2. Exact `hermes gateway install`/`<profile> gateway install` subcommand syntax for the
   systemd-unit-generation step was not directly quoted by the docs fetch (summarized, not
   verbatim) — confirm with `hermes gateway --help` on the actual VM before scripting it
   unattended; likely candidates from the CLI list: `hermes gateway setup` (interactive) is
   confirmed, a non-interactive `install`/`enable` variant is inferred but not directly quoted.
3. `hermes-lcm`'s exact plugin name string for `context.engine:` was reported as `lcm` by one
   source and the repo is literally named `hermes-lcm` — confirm the precise config value from
   the plugin's own `plugin.yaml`/README at install time rather than trusting the search
   summary verbatim.
4. Whether RTK's `rtk init --agent hermes` needs to run once per profile or once globally
   (`-g`) to cover a single-profile VM — likely moot under Open Q1(a) since there's only one
   profile, but worth a one-line check during lane B's plugin fan-out step.
5. Docs describe A2A's outbound toolset as "disabled by default, enable via `hermes tools`" —
   worth deciding during WF2/profile-draft whether Tars should have outbound A2A enabled at
   install time (to call other agents, e.g. an Orca-run session) vs. leaving it inbound-only
   until a concrete need appears (YAGNI-flavored call, not a blocker).
6. All Hermes-specific facts in this doc come from the vendor's own docs site fetched via
   WebFetch (an LLM-summarized read, not a byte-for-byte scrape) cross-checked against one live
   local artifact (`~/.hermes/plugins/orca-status` on cooper) and one previously-ingested live
   VM read (`learnings/hermes-vm.md`, p-Hermes, 2026-07-29). Nothing here was verified by
   actually running `hermes` — appropriate given the read-only recon mandate, but lane B should
   treat exact flag spellings as "verify with `--help`" rather than copy-paste gospel.

## Sources

- `hermes-agent.nousresearch.com/docs/getting-started/{installation,quickstart,platform-support}`
- `hermes-agent.nousresearch.com/docs/user-guide/{cli,configuration,profiles,multi-profile-gateways}`
- `hermes-agent.nousresearch.com/docs/user-guide/messaging/{,slack,a2a}`
- `hermes-agent.nousresearch.com/docs/user-guide/features/{plugins,memory,skills,personality}`
- `hermes-agent.nousresearch.com/docs/reference/{cli-commands,slash-commands}`
- `github.com/NousResearch/hermes-agent` (repo metadata, releases, issue #514 / PR #77109)
- `github.com/stephenschoettler/hermes-lcm`, `github.com/ayghri/i-have-adhd/blob/main/INSTALL.md`
- `hindsight.vectorize.io/sdks/integrations/hermes`
- `brew info rtk`, `rtk --help`, `rtk init --help`, `~/.config/rtk/` (live, cooper)
- `~/.hermes/{config.yaml,plugins/orca-status/*}` (live, cooper — a stub, not a real profile)
- `~/dev/gaetan-metarepo/learnings/hermes-vm.md` (previously-ingested live p-Hermes read,
  2026-07-29)
- `~/dev/gaetan-metarepo/repos/mc-metarepo/knowledge/hermes-{platform-primitives,agents-core-design}.md`
  (separate mc-agents Hermes deployment, ~v0.19 — session-model/toolset-scoping/approval-gate
  facts that still apply platform-wide, cited where relevant above)
- `~/dev/se-nmc496/systemd/hermes-gateway-support-engineer.service` (independent deployment,
  corroborates the systemd unit naming convention)
- `~/dev/Tars/{PITCH.md,PLAN.md,README.md,docs/conversation-2026-08-07.md}` (project context)
