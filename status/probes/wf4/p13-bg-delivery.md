# WF4 Probe 13 — home-channel delivery after `/sethome`

**Verdict: PASS** — an async job launched from the CLI (a non-home surface) delivered its
completion into the home DM `D0BBYNM01BL`, resolved at runtime from `SLACK_HOME_CHANNEL`.
Permalink: <https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129561552349>

**Spec correction (sub-finding, owner cutover-owned/spec-stale):** `/bg` exists in this Hermes
build and was exercised live, but it is **not** the home-channel mechanism. On every surface
`/bg` returns its result **to the invoking surface**, by design. `docs/specs/wf4-probes.md` §13's
instrument is wrong for this build; its *object* (does the `/sethome` target actually receive
async job completions?) is what was proven, using the only code path that reads
`SLACK_HOME_CHANNEL`: cron delivery with `--deliver slack`.

Date: 2026-08-07 · VM `tars` @ 192.168.0.9 · window 19:00:07Z → 19:10:23Z (post-cutover; cutover
`.env` write + gateway restart were 18:53Z per `status/probes/cutover-sethome.md`).
No secret printed, put on argv, or written to disk. No `sops` invocation. No unit restarted,
stopped, enabled or disabled (`NRestarts=0`, `MainPID=61969` unchanged across the whole probe).
No `config.yaml` / `.env` edit. `pve` untouched. No git command.

Credential handling: `SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN` were sourced inside the
remote shell (`set -a; . ~/.hermes/.env; set +a`) and handed to `curl` through a `-K` config on
**stdin** (`printf 'header = "Authorization: Bearer %s"\nheader = "Cookie: d=%s"\n' … | curl -K -`).
Values never appear in argv, in a file, or in any output below.

---

## 1 — `/bg` discovery: what the CLI and the gateway actually accept

### 1a. Registry — `/bg` is a real command, alias of `/background`

```
$ grep -n "/bg" ~/.hermes/hermes-agent/hermes_cli/commands.py
147:  CommandDef("background", "Run a prompt in the background", "Session",
             aliases=("bg", "btw"), args_hint="<prompt>", busy_policy="dispatch"),
```

Two implementations exist, one per surface:

| Surface | Handler | Where the result goes |
|---|---|---|
| interactive CLI REPL | `hermes_cli/cli_commands_mixin.py:1951 _handle_background_command` | printed to the CLI (`ChatConsole().print(Panel(...))`, "Background task #N complete") |
| gateway (Slack/Telegram/…) | `gateway/slash_commands.py:3314 _handle_background_command` → `gateway/run.py:19686 _run_background_task_inner` | `adapter.send(source.chat_id, …)` — the **invoking** chat |

Docstrings state the routing outright:

```
gateway/slash_commands.py:3315-3319
    """Handle /background <prompt> — run a prompt in a separate background session.
       Spawns a new AIAgent in a background thread with its own session.
       When it completes, sends the result back to the same chat …"""

hermes_cli/cli_commands_mixin.py:1951-1955
    """… When it completes, prints the result to the CLI without modifying
       the active session's conversation history."""
```

Neither reads `SLACK_HOME_CHANNEL`. `grep -rn "SLACK_HOME_CHANNEL"` reaches `/bg` from nowhere.

### 1b. CLI one-shot does **not** dispatch slash commands (run live, not inferred)

```
$ date -Is                                                            2026-08-07T19:01:58+00:00
$ hermes chat -Q -q "/bg wait 10 seconds then reply OK wf4-p13"
session_id: 20260807_190159_c4e4a7
Codex response remained incomplete after 3 continuation attempts
EXIT=0
$ date -Is                                                            2026-08-07T19:02:21+00:00
```

No "Background task started" line, no task id — a session was opened and the text was sent to the
**model** as a prompt. Slash dispatch lives in the REPL path (`cli.py:4012 is_slash_command`), which
`-Q -q` never reaches. So the spec's "run `/bg` … from the CLI" has no non-interactive form at all;
and even the interactive form prints to stdout, never to Slack (the CLI is a separate process from
the gateway).

### 1c. Gateway `/bg` — live, accepted, and routed to the invoking surface

Posted as Gaetan (`U08BDJAMSRZ`, personal xoxc) via `chat.postMessage` into `D0BBYNM01BL`:

```
2026-08-07T19:07:11Z  chat.postMessage  ok=True ts=1786129631.661329 channel=D0BBYNM01BL
```

`conversations.replies(channel=D0BBYNM01BL, ts=1786129631.661329)` → `ok=True`:

```
ts=1786129631.661329 19:07:11.661Z user=U08BDJAMSRZ  '/bg wait 10 seconds then reply exactly: wf4-p13-bg OK'
ts=1786129633.009489 19:07:13.009Z bot_id=B0BBSBVBJUB ':arrows_counterclockwise: Background task started:
                                    "wait 10 seconds…" Task ID: bg_190712_ae1e80 …'
ts=1786129652.959719 19:07:32.959Z bot_id=B0BBSBVBJUB ':white_check_mark: Background task complete
                                    Prompt: "wait 10 seconds…"  wf4-p13-bg OK'
```

- invocation: <https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129631661329?thread_ts=1786129631.661329&cid=D0BBYNM01BL>
- completion: <https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129652959719?thread_ts=1786129631.661329&cid=D0BBYNM01BL>

Gateway log for the same window (`~/.hermes/logs/agent.log`):

```
19:07:12,862 INFO gateway.platforms.base: [Slack] Sending response (159 chars) to D0BBYNM01BL   ← the ack
19:07:12,909 INFO [bg_190712_ae1e80] agent.turn_context: conversation turn: session=bg_190712_ae1e80
19:07:17,331 INFO [bg_190712_ae1e80] tools.terminal_tool: Creating new local environment …
19:07:27,589 INFO [bg_190712_ae1e80] agent.tool_executor: tool terminal completed (10.27s, 45 chars)  ← the sleep 10
19:07:32,611 INFO [bg_190712_ae1e80] agent.conversation_loop: Turn ended: reason=text_response
```

**`/bg` works.** Both the ack and the completion landed as **thread replies under the invoking
message** (`thread_ts` = the invoking `ts`), i.e. in the invoking surface. This run was invoked
from the home DM itself (deliberately — see §4 on why not the shared channel), so "invoking
surface" and "home DM" coincide here; §1a's code is what proves the routing is
`source.chat_id`, not the home target. Either way, `/bg` cannot satisfy §13's PASS text
("arrives in the `/sethome`-set DM, **not the invoking channel**") — the two are the same
destination by construction.

---

## 2 — What *does* read `SLACK_HOME_CHANNEL`: cron delivery

```
cron/scheduler.py:263-281   _HOME_TARGET_ENV_VARS = { … "slack": "SLACK_HOME_CHANNEL", … }
cron/scheduler.py:1152      def _get_home_target_chat_id(platform_name) -> str:
                                env_var = _resolve_home_env_var(platform_name)   # "SLACK_HOME_CHANNEL"
                                value = os.getenv(env_var, "")                    # resolved at delivery time
                                …
cron/scheduler.py:1268/1327 chat_id = _get_home_target_chat_id(platform_name)
```

A bare platform name in `--deliver` (no `:chat_id`) therefore resolves the chat id **from the env
var `/sethome` writes** — the exact pipe `status/probes/cutover-sethome.md` set up and left
unproven. (Same for `hermes send --to slack`, whose `--help` says: *"'platform' (home channel)"*.)

---

## 3 — The probe: async job from the CLI → home DM

### 3a. Baseline (before)

```
$ date -Is                                                            2026-08-07T19:04:12+00:00
$ conversations.history(channel=D0BBYNM01BL, limit=8)   ok=True
  ts=1786129291.039239 19:01:31Z user=U08BDJAMSRZ  'ping wf4-1'            ← probe 1, another agent
  ts=1781961770.329849 2026-06-20     … (all remaining messages are June 20, p-Hermes era)
```

Newest item: `1786129291.039239`. Poll floor for the probe set to `oldest=1786129292`.

### 3b. Trigger — from the CLI over ssh (no TTY, no Slack surface involved)

```
$ date -Is                                                            2026-08-07T19:04:21+00:00
$ hermes cron create "1m" "Reply with exactly this text and nothing else: wf4-p13-cron OK" \
      --deliver slack --name wf4-p13 --repeat 1
Created job: 004adebba8ee
  Name: wf4-p13
  Schedule: once in 1m
  Next run: 2026-08-07T19:05:22.139491+00:00

$ hermes cron list
  004adebba8ee [active]  Name: wf4-p13  Schedule: once in 1m  Repeat: 0/1  Deliver: slack
$ date -Is                                                            2026-08-07T19:04:22+00:00
```

`Deliver: slack` = bare platform = home channel (§2). The invoking process **exited at 19:04:22**,
70s before the job ran — it is structurally incapable of receiving the completion. Whatever
arrives, arrives because the gateway resolved a target on its own.

### 3c. Execution + delivery — gateway log (`~/.hermes/logs/agent.log`, lines 1326-1346)

```
19:05:58,147 INFO cron.scheduler: Running job 'wf4-p13' (ID: 004adebba8ee)
19:05:58,147 INFO cron.scheduler: Prompt: [IMPORTANT: You are running as a scheduled cron job. DELIVERY: …
19:05:58,173 INFO cron.scheduler: Job '004adebba8ee': 44 MCP tool(s) available
19:05:58,260 INFO [cron_004adebba8ee_20260807_190558] agent.turn_context: conversation turn …
19:06:01,276 INFO [cron_004adebba8ee_20260807_190558] agent.conversation_loop: API call #1: model=gpt-5.6-sol
19:06:01,309 INFO cron.scheduler: Job 'wf4-p13' completed successfully
19:06:01,622 INFO cron.scheduler: Job '004adebba8ee': delivered to slack:D0BBYNM01BL via live adapter
```

That last line is the routing evidence: the scheduler resolved `slack` → **`D0BBYNM01BL`** and sent
through the live Socket-Mode adapter. Durable record:

```
$ hermes cron runs
f99a1bdc74b74f89944499c67d237506  completed  job=004adebba8ee  source=builtin  2026-08-07T19:05:58.086734+00:00
```

`journalctl --user -u hermes-gateway.service --since '2026-08-07 19:04:00'` carries only 6 lines in
this window, all `WARNING tools.registry: check_fn … returned False` at `19:05:58` — the cron
agent's toolset registration inside gateway PID `61969`. The unit logs at WARNING; the INFO
delivery line lives in `~/.hermes/logs/agent.log` (same process), quoted above.

### 3d. Arrival in the home DM — polled with the personal xoxc session

```
$ bash poll.sh D0BBYNM01BL 1786129292 12          # conversations.history, oldest=1786129292
2026-08-07T19:04:50+00:00 poll#1 new_messages=0
2026-08-07T19:05:05+00:00 poll#2 new_messages=0
2026-08-07T19:05:20+00:00 poll#3 new_messages=0
2026-08-07T19:05:36+00:00 poll#4 new_messages=0
2026-08-07T19:05:51+00:00 poll#5 new_messages=0
2026-08-07T19:06:06+00:00 poll#6 new_messages=1
ok= True error= None
  ts=1786129561.552349 utc=2026-08-07T19:06:01.552349+00:00
  bot_id=B0BBSBVBJUB user=U0BBH85NAKH app_id=A0BC0GXH78R bot_profile=Tars
  text='Cronjob Response: wf4-p13 (job_id: 004adebba8ee) -------------  wf4-p13-cron OK
        To stop or manage this job, send me a new message (e.g. "stop reminder wf4-p13").'
```

```
$ chat.getPermalink(channel=D0BBYNM01BL, message_ts=1786129561.552349)
ok= True   channel= D0BBYNM01BL
permalink= https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129561552349
```

- sender is the Tars app (`bot_profile=Tars`, `bot_id=B0BBSBVBJUB`, `app_id=A0BC0GXH78R`)
- destination is `D0BBYNM01BL` — the channel id `/sethome` set, byte-identical
- job id `004adebba8ee` in the body ties the Slack message to the CLI invocation of §3b
- delivered **top-level in the DM**, 99s after the CLI process had exited

### 3e. Cleanup — self-cleaning

```
$ date -Is                                                            2026-08-07T19:10:23+00:00
$ hermes cron list   → No scheduled jobs.          # --repeat 1 retired the job after its single run
$ systemctl --user show hermes-gateway.service -p NRestarts -p ActiveState -p MainPID
MainPID=61969  NRestarts=0  ActiveState=active     # same PID as before the probe
```

Temp helper scripts removed from the VM (`/tmp/wf4p13-*`). No residual state.

---

## 4 — Scope calls made, and why

- **Gateway `/bg` was fired in the home DM, not in `#gcn-sandbox` (`C08RWSTU9LK`).** The agent log
  shows another WF4 probe actively driving that channel at 19:01-19:02 (`inbound message:
  platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK`, `response ready … chat=C08RWSTU9LK`) — probe 2
  (mention gating) and/or probe 3 (negative test). Injecting a `/bg` there would have contaminated
  their evidence window. Routing to `source.chat_id` is proven from source (§1a) instead.
- **No `/bg` was simulated.** Every `/bg` line above is a real invocation with its real result.
- **Cron was used only as a one-shot delivery vehicle.** Recurring scheduling remains WF5 scope
  per PLAN.md; the job carried `--repeat 1` and is gone.

## 5 — Cross-probe finding worth propagating

Tars is an **Agent/Assistant-class** Slack app: its conversational replies land as **thread
replies** (`thread_ts` = the user's message), which `conversations.history` does **not** return.
Both `/bg` messages were invisible to `conversations.history` and only showed up under
`conversations.replies`. Probe 1's reply to `ping wf4-1` is likewise absent from the history dump
in §3a. **Cron/home-channel deliveries are different** — they are posted top-level and *are* visible
to `conversations.history`. Any probe polling `conversations.history` for a bot reply to a user
message will see a false negative; use `conversations.replies(ts=<invoking ts>)`.

## 6 — Verdict

| Claim | Verdict | Evidence |
|---|---|---|
| Home DM `D0BBYNM01BL` receives async job completions triggered from a non-home surface | **PASS** | `delivered to slack:D0BBYNM01BL via live adapter` + permalink `p1786129561552349` |
| Target is resolved from `SLACK_HOME_CHANNEL`, not hardcoded | **PASS** | `_get_home_target_chat_id` → `os.getenv("SLACK_HOME_CHANNEL")`; `--deliver slack` carried no chat id |
| Completion does not go to the invoking surface | **PASS** | invoking CLI process exited 19:04:22; delivery 19:06:01 |
| `/bg` exists in this build | **PASS** | ack + completion, `Task ID: bg_190712_ae1e80` |
| `/bg` completion routes to the home DM rather than the invoking chat | **FALSE — spec-stale** | `adapter.send(source.chat_id, …)`; thread replies under the invoking message |

Probe 13 **PASS**. §13's `/bg` instrument needs rewriting to the cron `--deliver <platform>` form
(owner: cutover-owned/spec-stale) — a documentation defect, not a wiring defect: nothing in WF3 or
lane B needs re-running.
