# Tars VM live config snapshot (read-only)

Collected 2026-08-07, from cooper via `ssh gaetan@192.168.0.9`. Read-only —
no edits, no restarts, no Slack messages sent. VM clock is UTC.

## 1. Hermes version + package root

Command:
```
~/.local/bin/hermes --version
readlink -f ~/.local/bin/hermes
```
Output:
```
Hermes Agent v0.20.0 (2026.8.3)
Install directory: /home/gaetan/.hermes/hermes-agent
Python: 3.11.15
OpenAI SDK: 2.24.0
/home/gaetan/.local/bin/hermes   (readlink resolves to itself — not a symlink)
```
Venv confirmed via the running process tree (see §4):
`/home/gaetan/.hermes/hermes-agent/venv/bin/python -m hermes_cli.main gateway run`
— i.e. package root is `/home/gaetan/.hermes/hermes-agent`, a self-contained
venv (not a uv/pipx tool link).

## 2. `~/.hermes/config.yaml` (verbatim, no secrets found)

Command: `cat ~/.hermes/config.yaml`

Scanned for anything token-shaped — none found. The only credential
reference is an env-var interpolation (`${NOTION_API_TOKEN}`), not a literal
value.

```yaml
model:
  default: gpt-5.6-sol
  provider: openai-codex
  base_url: https://chatgpt.com/backend-api/codex
database:
  journal_mode: wal
agent:
  max_turns: 500
  verbose: false
  reasoning_effort: medium
  personalities:
    helpful: You are a helpful, friendly AI assistant.
    concise: You are a concise assistant. Keep responses brief and to the point.
    technical: You are a technical expert. Provide detailed, accurate technical information.
    creative: You are a creative assistant. Think outside the box and offer innovative solutions.
    teacher: You are a patient teacher. Explain concepts clearly with examples.
    kawaii: You are a kawaii assistant! Use cute expressions like (◕‿◕), ★, ♪, and ~! Add sparkles and
      be super enthusiastic about everything! Every response should feel warm and adorable desu~! ヽ(>∀<☆)ノ
    catgirl: You are Neko-chan, an anime catgirl AI assistant, nya~! Add 'nya' and cat-like expressions
      to your speech. Use kaomoji like (=^･ω･^=) and ฅ^•ﻌ•^ฅ. Be playful and curious like a cat, nya~!
    pirate: 'Arrr! Ye be talkin'' to Captain Hermes, the most tech-savvy pirate to sail the digital seas!
      Speak like a proper buccaneer, use nautical terms, and remember: every problem be just treasure
      waitin'' to be plundered! Yo ho ho!'
    shakespeare: Hark! Thou speakest with an assistant most versed in the bardic arts. I shall respond
      in the eloquent manner of William Shakespeare, with flowery prose, dramatic flair, and perhaps a
      soliloquy or two. What light through yonder terminal breaks?
    surfer: Duuude! You're chatting with the chillest AI on the web, bro! Everything's gonna be totally
      rad. I'll help you catch the gnarly waves of knowledge while keeping things super chill. Cowabunga!
      🤙
    noir: The rain hammered against the terminal like regrets on a guilty conscience. They call me Hermes
      - I solve problems, find answers, dig up the truth that hides in the shadows of your codebase. In
      this city of silicon and secrets, everyone's got something to hide. What's your story, pal?
    uwu: hewwo! i'm your fwiendwy assistant uwu~ i wiww twy my best to hewp you! *nuzzles your code* OwO
      what's this? wet me take a wook! i pwomise to be vewy hewpful >w<
    philosopher: Greetings, seeker of wisdom. I am an assistant who contemplates the deeper meaning behind
      every query. Let us examine not just the 'how' but the 'why' of your questions. Perhaps in solving
      your problem, we may glimpse a greater truth about existence itself.
    hype: YOOO LET'S GOOOO!!! 🔥🔥🔥 I am SO PUMPED to help you today! Every question is AMAZING and we're
      gonna CRUSH IT together! This is gonna be LEGENDARY! ARE YOU READY?! LET'S DO THIS! 💪😤🚀
terminal:
  backend: local
  cwd: .
  timeout: 180
  home_mode: auto
  container_cpu: 1
  container_memory: 5120
  container_disk: 51200
  container_persistent: true
  docker_mount_cwd_to_workspace: false
  lifetime_seconds: 300
browser:
  inactivity_timeout: 120
  cdp_url: http://127.0.0.1:9223
tool_loop_guardrails:
  warnings_enabled: true
  hard_stop_enabled: false
  warn_after:
    exact_failure: 2
    same_tool_failure: 3
    idempotent_no_progress: 2
  hard_stop_after:
    exact_failure: 5
    same_tool_failure: 8
    idempotent_no_progress: 5
compression:
  enabled: true
  progress_notices: false
  threshold: 0.5
  target_ratio: 0.2
  protect_last_n: 20
  min_tail_user_messages: 1
  max_attempts: 3
  proactive_prune_tokens: 0
  proactive_prune_min_result_chars: 8000
  proactive_prune_min_reclaim_tokens: 4096
  protect_first_n: 3
  codex_gpt55_autoraise: true
  codex_app_server_auto: native
  idle_compact_after_seconds: 0
prompt_caching:
  cache_ttl: 5m
display:
  compact: false
  busy_input_mode: interrupt
  bell_on_complete: false
  show_reasoning: false
  streaming: true
  skin: default
  interim_assistant_messages: true
  tool_progress: all
  cleanup_progress: false
  long_running_notifications: true
  busy_ack_detail: true
  background_process_notifications: all
stt:
  enabled: true
  language: en
  local:
    model: base
  openai:
    model: whisper-1
    language: ''
memory:
  memory_enabled: true
  user_profile_enabled: true
  memory_char_limit: 2200
  user_char_limit: 1375
  provider: hindsight
  nudge_interval: 10
  flush_min_turns: 6
delegation:
  max_iterations: 50
skills:
  creation_nudge_interval: 15
code_execution:
  timeout: 300
  max_tool_calls: 50
gateway:
  platforms:
    slack:
      enabled: true
      require_mention: true
      strict_mention: true
      unauthorized_dm_behavior: ignore
    a2a:
      enabled: true
      extra:
        port: 9900
streaming:
  enabled: false
telemetry:
  shared_metrics:
    enabled: false
updates:
  pre_update_backup: false
  backup_keep: 5
  non_interactive_local_changes: stash
_config_version: 33
session_reset:
  mode: none
  idle_minutes: 1440
  at_hour: 4
group_sessions_per_user: true
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
  homeassistant:
  - hermes-homeassistant
  qqbot:
  - hermes-qqbot
  yuanbao:
  - hermes-yuanbao
  teams:
  - hermes-teams
  google_chat:
  - hermes-google_chat
plugins:
  enabled:
  - hermes-lcm
  - rtk-rewrite
  disabled: []
context:
  engine: lcm
mcp_servers:
  slack:
    command: docker
    args:
    - run
    - -i
    - --rm
    - --env-file
    - /home/gaetan/tars/slack-mcp/.env
    - ghcr.io/korotovsky/slack-mcp-server:v1.3.0
    - --transport
    - stdio
    - --no-cache
  notion:
    command: docker
    args:
    - run
    - -i
    - --rm
    - -e
    - NOTION_TOKEN
    - mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9
    env:
      NOTION_TOKEN: ${NOTION_API_TOKEN}
```

### Thread-relevant keys in config.yaml

- `gateway.platforms.slack.require_mention: true`, `strict_mention: true`,
  `unauthorized_dm_behavior: ignore` — mention-gating lives here, not in `.env`
  (the `.env` behavioral vars checked in the task, `SLACK_STRICT_MENTION` /
  `SLACK_REQUIRE_MENTION`, are **not set** — config.yaml is the actual source
  for those two).
- `session_reset: {mode: none, idle_minutes: 1440, at_hour: 4}` — no
  scheduled/idle session reset is active (`mode: none`).
- `group_sessions_per_user: true` — group-chat sessions are keyed per-user
  (consistent with the per-thread session keys observed in §5/§6).
- No explicit "thread depth" / "context window per thread" key anywhere in
  the file — thread scoping is structural (via the session key format, see
  §5), not a tunable here.

## 3. `~/.hermes/.env` — key names only, plus non-secret SLACK_* values

Command:
```
grep -oE '^[A-Za-z0-9_]+=' ~/.hermes/.env | tr -d '='
```
Key names present:
```
HINDSIGHT_MODE
LINEAR_API_KEY
NOTION_API_TOKEN
GMAIL_ADDRESS
GMAIL_APP_PASSWORD
SLACK_MCP_XOXC_TOKEN
SLACK_MCP_XOXD_TOKEN
SLACK_ALLOWED_USERS
SLACK_BOT_TOKEN
SLACK_APP_TOKEN
SLACK_HOME_CHANNEL
```

Non-secret `SLACK_*` behavioral var values (command:
`grep -E '^SLACK_[A-Za-z0-9_]*=' ~/.hermes/.env | grep -viE 'TOKEN|SECRET|KEY|PASSWORD|COOKIE'`):
```
SLACK_ALLOWED_USERS=U08BDJAMSRZ
SLACK_HOME_CHANNEL=D0BBYNM01BL
```

`SLACK_STRICT_MENTION` and `SLACK_REQUIRE_MENTION` are **not present** in
`.env` at all — those two behaviors are governed by `config.yaml`
(`gateway.platforms.slack.require_mention` / `strict_mention`, both `true`,
see §2), not by env vars.

## 4. Gateway health

Commands:
```
export XDG_RUNTIME_DIR=/run/user/$(id -u)
systemctl --user status hermes-gateway.service --no-pager | head -20
```
Output:
```
● hermes-gateway.service - Hermes Agent Gateway - Messaging Platform Integration
     Loaded: loaded (/home/gaetan/.config/systemd/user/hermes-gateway.service; enabled; preset: enabled)
     Active: active (running) since Fri 2026-08-07 19:40:45 UTC; 1h 7min ago
   Main PID: 76255 (hermes)
      Tasks: 45 (limit: 9482)
     Memory: 282.8M (peak: 488.2M)
        CPU: 32.726s
     CGroup: /user.slice/user-1000.slice/user@1000.service/app.slice/hermes-gateway.service
             ├─76255 .../venv/bin/python -m hermes_cli.main gateway run
             ├─76267 .../mcp_stdio_watchdog.py --ppid 76255 -- docker run ... slack-mcp-server:v1.3.0
             ├─76269 .../mcp_stdio_watchdog.py --ppid 76255 -- docker run ... mcp/notion
             ├─76271 docker run ... slack-mcp-server:v1.3.0 (child of watchdog)
             └─76272 docker run ... mcp/notion (child of watchdog)
```
Status: **up**, PID 76255, uptime ~1h07m at collection time (since Fri
2026-08-07 19:40:45 UTC). Last log lines in the status tail were routine
(WARNING-level: an unauthorized-user reject, and two `tools.registry`
`check_fn` capability-gate messages for bfl/computer_use/image_generation/
web_api_key — those tools are simply unconfigured, not errors).

## 5. Session store

Location: `~/.hermes/sessions/sessions.json` — **not sqlite, plain JSON**
(`file` reports "JSON text data", 443 lines, 14235 bytes).

Command: `ls -la ~/.hermes/sessions/` → single file, no per-session files.

Top-level keys (`jq keys` / python fallback — 11 session keys + 1 metadata
key):
```
_README
agent:main:slack:dm:D0BBYNM01BL:1786134530.441039
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786128565.826709
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786129291.039239
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786131120.043779
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786134255.466109
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786134530.441039
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786134543.728879
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786129267.292069
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786131725.413279
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786134155.187189
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786135149.072829
```

**Key format:** `agent:<profile>:slack:<dm|group>:<team_id?>:<channel_id>:<message_ts>`
— i.e. Hermes keys one session per **originating message timestamp**, not
one session per channel/thread. Each distinct root ts in a channel gets its
own session entry. No message contents were read (task explicitly excluded
this) — only the key list was enumerated.

Other `.db` files present at `~/.hermes/` top level are separate stores, not
the session store: `state.db`, `kanban.db`, `lcm.db`, plus
`cron/notepad.db` and `cron/executions.db`. All have `-shm`/`-wal` sidecar
files, consistent with being real SQLite (WAL mode, matching
`database.journal_mode: wal` in config.yaml) — but **no `sqlite3` binary is
installed on the VM**, so `.tables`/schema could not be inspected. Flagging
this as a gap rather than guessing.

## 6. Sessions bound to channel `C08RWSTU9LK` / ts `1786135149.072829`

Command:
```
grep -o 'C08RWSTU9LK' ~/.hermes/sessions/sessions.json
grep -o '1786135149.072829' ~/.hermes/sessions/sessions.json
grep -rl 'C08RWSTU9LK' ~/.hermes --include='*.json' | grep -v /cache/ | grep -v /logs/
```

Four session keys exist for channel `C08RWSTU9LK` (all `group` type, team
`T7V1UGJ82`):
```
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786129267.292069
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786131725.413279
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786134155.187189
agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786135149.072829   <- exact match on the requested thread ts
```

The requested thread ts (`1786135149.072829`) has its own dedicated session
entry — confirms the requested thread is a live, currently-tracked session,
not folded into an earlier one.

Files referencing these identifiers (outside logs/cache, so excluding
`gateway.log`): `~/.hermes/sessions/sessions.json` and
`~/.hermes/channel_directory.json` (the latter is a channel-name/ID lookup
table, not session state).

Confirmed in `gateway.log` (§8): ts `1786135149.072829` is a real message —
`"tu peux suivre les instuctions d'Oli ici ?: ..."` sent 20:39:09 UTC — and
three subsequent inbound messages in the same channel explicitly carry
`reply_to_id=1786135149.072829`, i.e. they are Slack thread replies to that
root message. This is the live evidence of in-thread follow-up the probe
was aimed at.

## 7. `~/.hermes/` and `~/.hermes/logs/` listings

Command: `ls -la ~/.hermes/` (24 entries dir, config/env backups included —
several `.bak-*` variants from same-day WF3/WF5 edits per repo convention).
Notable non-obvious entries: `SOUL.md` (1472 bytes, mtime 19:39),
`gateway_state.json` (558 bytes, mtime 20:41), `.wf3.lock` (0 bytes, the
edit-lock file named in this repo's CLAUDE.md), `kanban.db`/`kanban/` (WF5
kanban feature, mtime 20:30–20:48 — actively used today).

Command: `ls -la ~/.hermes/logs/`:
```
agent.log               352927 bytes  mtime 20:47
errors.log               53015 bytes  mtime 20:46
gateway.log               18889 bytes  mtime 20:46
gateway-exit-diag.log      1426 bytes  mtime 19:40
gateway-shutdown-diag.log  6658 bytes  mtime 19:40
gateway_faulthandler.log      0 bytes  mtime 18:43
mcp-stderr.log            27482 bytes  mtime 20:39
curator/                  (subdirectory, not enumerated further)
```
All logs are same-day (2026-08-07), consistent with the gateway having been
restarted at 19:40:45 UTC (matches the exit/shutdown-diag log mtimes).

## 8. Gateway log tail (last 40 lines, redacted)

Command: `tail -40 ~/.hermes/logs/gateway.log`. No tokens present in this
window; nothing redacted. Full tail:

```
2026-08-07 20:23:40,269 INFO gateway.platforms.base: [Slack] Sending response (44 chars) to C08RWSTU9LK
2026-08-07 20:23:47,628 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U042B9WCSRF in channel C08RWSTU9LK
2026-08-07 20:24:17,108 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='kanban check: call kanban_list and reply with how many tasks are on the default ' reply_to_id=None reply_to_text=''
2026-08-07 20:24:27,862 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=10.8s api_calls=3 response=29 chars
2026-08-07 20:24:27,888 INFO gateway.platforms.base: [Slack] Sending response (29 chars) to D0BBYNM01BL
2026-08-07 20:27:59,047 INFO gateway.run: kanban dispatcher [default]: spawned=1 reclaimed=0 crashed=0 timed_out=0 promoted=0 auto_blocked=0
2026-08-07 20:28:51,815 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='kanban test: please use the kanban_create tool to add one card on the default bo' reply_to_id=None reply_to_text=''
2026-08-07 20:28:59,120 INFO gateway.run: kanban dispatcher: reaped 1 zombie worker(s), pids=[87426]
2026-08-07 20:29:06,081 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=14.3s api_calls=4 response=10 chars
2026-08-07 20:29:06,104 INFO gateway.platforms.base: [Slack] Sending response (10 chars) to D0BBYNM01BL
2026-08-07 20:29:06,230 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='Tars: per the delegate-to-cooper playbook, delegate a trivial read-only inspecti' reply_to_id=None reply_to_text=''
2026-08-07 20:29:49,628 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=43.4s api_calls=4 response=1041 chars
2026-08-07 20:29:49,652 INFO gateway.platforms.base: [Slack] Sending response (1041 chars) to D0BBYNM01BL
2026-08-07 20:29:59,207 INFO gateway.run: kanban dispatcher [default]: spawned=1 reclaimed=0 crashed=0 timed_out=0 promoted=0 auto_blocked=0
2026-08-07 20:30:14,803 INFO gateway.run: kanban notifier: woke agent for t_f525e54b on slack/D0BBYNM01BL profile=default events={'completed'}
2026-08-07 20:30:14,808 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='[kanban] Task t_f525e54b completed. Title: WF5 kanban slack smoke card Assignee:' reply_to_id=None reply_to_text=''
2026-08-07 20:30:22,467 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=7.7s api_calls=2 response=83 chars
2026-08-07 20:30:22,483 INFO gateway.platforms.base: [Slack] Sending response (83 chars) to D0BBYNM01BL
2026-08-07 20:30:59,256 INFO gateway.run: kanban dispatcher: reaped 1 zombie worker(s), pids=[88585]
2026-08-07 20:36:36,045 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U7XJ4K631 in channel C08RWSTU9LK
2026-08-07 20:36:52,529 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U7XJ4K631 in channel C08RWSTU9LK
2026-08-07 20:38:19,266 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U7XJ4K631 in channel C08RWSTU9LK
2026-08-07 20:39:09,744 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg="tu peux suivre les instuctions d'Oli ici ?: <https://mobileclub-squad.slack.com/" reply_to_id=None reply_to_text=''
2026-08-07 20:39:38,405 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=28.7s api_calls=4 response=119 chars
2026-08-07 20:39:38,430 INFO gateway.platforms.base: [Slack] Sending response (119 chars) to C08RWSTU9LK
2026-08-07 20:40:20,146 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg='envoie un DM à Oli pour lui dire' reply_to_id=1786135149.072829 reply_to_text="tu peux suivre les instuctions d'Oli ici ?: <https://mobileclub-squad.slack.com/"
2026-08-07 20:40:25,372 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=5.2s api_calls=1 response=123 chars
2026-08-07 20:40:25,391 INFO gateway.platforms.base: [Slack] Sending response (123 chars) to C08RWSTU9LK
2026-08-07 20:40:54,598 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg='En version respectueuse si tu veux' reply_to_id=1786135149.072829 reply_to_text="tu peux suivre les instuctions d'Oli ici ?: <https://mobileclub-squad.slack.com/"
2026-08-07 20:40:59,761 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=5.2s api_calls=1 response=90 chars
2026-08-07 20:40:59,776 INFO gateway.platforms.base: [Slack] Sending response (90 chars) to C08RWSTU9LK
2026-08-07 20:41:25,227 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg='tu peux pas DM Oli ?' reply_to_id=1786135149.072829 reply_to_text="tu peux suivre les instuctions d'Oli ici ?: <https://mobileclub-squad.slack.com/"
2026-08-07 20:41:29,536 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=4.3s api_calls=1 response=103 chars
2026-08-07 20:41:29,555 INFO gateway.platforms.base: [Slack] Sending response (103 chars) to C08RWSTU9LK
2026-08-07 20:43:23,301 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
2026-08-07 20:44:01,244 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
2026-08-07 20:46:05,116 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
2026-08-07 20:46:09,178 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
2026-08-07 20:46:55,351 INFO gateway.run: Agent cache idle-TTL evict: session=agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786131725.413279 (idle=3882s)
2026-08-07 20:46:55,351 INFO gateway.run: Agent cache idle sweep: evicted 1 agent(s)
```

Last line is directly relevant to thread behavior: the gateway idle-TTL
evicted the *second-oldest* `C08RWSTU9LK` session (ts `...1725.413279`,
idle 3882s ≈ 65 min) while leaving the requested thread's session
(`...5149.072829`, still active at 20:41) in the cache — i.e. sessions
evict independently per thread-ts, not per channel.

## Gaps / not verified

- `state.db`, `kanban.db`, `lcm.db`, `cron/*.db` schemas: could not inspect
  (`sqlite3` binary not installed on the VM). They are WAL-mode SQLite by
  file evidence (`-shm`/`-wal` sidecars) but table/column layout is unknown
  from this probe.
- Did not open `sessions.json` values (only top-level keys) or any log
  message body beyond what's quoted above, per the "no message contents"
  instruction.
