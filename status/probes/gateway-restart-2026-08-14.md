# Hermes gateway restart — 2026-08-14

Task: restart `hermes-gateway.service` (--user unit) on the Tars VM
(192.168.0.9) and prove it came back healthy. Restart only — no config
edits, no `sops -d`.

## 1. BEFORE state

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); date -u; systemctl --user show hermes-gateway.service -p ActiveEnterTimestamp -p ActiveState -p SubState -p MainPID'
Fri Aug 14 11:35:22 AM UTC 2026
MainPID=1578393
ActiveState=active
SubState=running
ActiveEnterTimestamp=Thu 2026-08-13 15:24:11 UTC
```

(NRestarts intentionally NOT captured/used per instructions — unreliable on
this host.)

## 2. Restart command

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); systemctl --user restart hermes-gateway.service'
restart exit=0
```

## 3. AFTER state (captured ~21s post-restart-command, after a 15s wait)

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); date -u; systemctl --user show hermes-gateway.service -p ActiveEnterTimestamp -p ActiveState -p SubState -p MainPID'
Fri Aug 14 11:35:43 AM UTC 2026
MainPID=1784687
ActiveState=active
SubState=running
ActiveEnterTimestamp=Fri 2026-08-14 11:35:26 UTC
```

**Proof of restart:**
- `ActiveEnterTimestamp` moved forward: `2026-08-13 15:24:11 UTC` →
  `2026-08-14 11:35:26 UTC`.
- `MainPID` changed: `1578393` → `1784687`.
- `ActiveState=active` / `SubState=running` after restart.

## 4. Logs after restart

### `~/.hermes/logs/gateway.log` (tail 40, spans pre- and post-restart)

```
2026-08-14 11:32:02,344 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C0BQCB58ATW msg='[ASY...
2026-08-14 11:32:35,971 INFO gateway.run: response ready: platform=slack chat=C0BQCB58ATW time=33.6s api_calls=3 resp...
2026-08-14 11:32:35,998 INFO gateway.platforms.base: [Slack] Sending response (454 chars) to C0BQCB58ATW
2026-08-14 11:32:57,922 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='Rest...
2026-08-14 11:33:20,028 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=22.1s api_calls=4 resp...
2026-08-14 11:33:20,055 INFO gateway.platforms.base: [Slack] Sending response (166 chars) to D0BBYNM01BL
2026-08-14 11:35:24,872 INFO gateway.run: Received SIGTERM — initiating shutdown
2026-08-14 11:35:24,873 WARNING gateway.run: Shutdown context: signal=SIGTERM under_systemd=yes parent_pid=27136 pare...
2026-08-14 11:35:24,874 INFO gateway.run: Stopping gateway...
2026-08-14 11:35:25,187 INFO gateway.run: Sent shutdown notification to home channel slack:D0BBYNM01BL
2026-08-14 11:35:25,187 INFO gateway.run: Shutdown phase: notify_active_sessions done at +0.31s
2026-08-14 11:35:25,192 INFO gateway.run: Shutdown phase: drain done at +0.32s (drain took 0.00s, timed_out=False, ac...
2026-08-14 11:35:25,204 INFO hermes_plugins.slack_platform.adapter: [Slack] Disconnected
2026-08-14 11:35:25,204 INFO gateway.run: ✓ slack disconnected (0.00s)
2026-08-14 11:35:25,613 INFO gateway.run: ✓ a2a disconnected (0.41s)
2026-08-14 11:35:25,613 INFO gateway.run: Shutdown phase: all adapters disconnected at +0.74s
2026-08-14 11:35:25,615 INFO gateway.run: Shutdown phase: final-cleanup tool kill done at +0.74s
2026-08-14 11:35:25,623 INFO gateway.run: Shutdown phase: SessionDB close done at +0.75s
2026-08-14 11:35:25,624 INFO gateway.run: Gateway stopped by an unexpected signal — persisting gateway_state=running ...
2026-08-14 11:35:25,633 INFO gateway.run: Gateway stopped (total teardown 0.76s)
2026-08-14 11:35:25,634 INFO gateway.run: Gateway housekeeping stopped
2026-08-14 11:35:25,990 INFO gateway.run: Exiting with code 1 (signal-initiated shutdown without restart request) so ...
2026-08-14 11:35:29,310 INFO gateway.run: Starting Hermes Gateway...
2026-08-14 11:35:29,312 INFO gateway.run: Session storage: /home/gaetan/.hermes/sessions
2026-08-14 11:35:29,318 INFO gateway.run: Agent budget: max_iterations=500 (agent.max_turns from config.yaml, or HERM...
2026-08-14 11:35:29,319 INFO gateway.run: Secret redaction: ENABLED (tool output, logs, and chat responses are scrubb...
2026-08-14 11:35:29,335 INFO gateway.run: Previous gateway exited cleanly — skipping session suspension
2026-08-14 11:35:29,335 INFO gateway.run: Connecting to slack...
2026-08-14 11:35:29,563 INFO hermes_plugins.slack_platform.adapter: [Slack] Authenticated as @tars in workspace Mobil...
2026-08-14 11:35:29,611 INFO hermes_plugins.slack_platform.adapter: [Slack] Socket Mode connected (1 workspace(s))
2026-08-14 11:35:29,619 INFO gateway.run: ✓ slack connected
2026-08-14 11:35:29,620 INFO gateway.run: Connecting to a2a...
2026-08-14 11:35:29,629 INFO hermes_plugins.a2a_platform.adapter: A2A: serving Agent Card + JSON-RPC on http://127.0....
2026-08-14 11:35:29,634 INFO gateway.run: ✓ a2a connected
2026-08-14 11:35:29,641 INFO gateway.run: Gateway running with 2 platform(s)
2026-08-14 11:35:30,602 INFO gateway.run: Channel directory built: 105 target(s)
2026-08-14 11:35:31,615 INFO gateway.run: Press Ctrl+C to stop
2026-08-14 11:35:31,619 INFO gateway.run: kanban dispatcher: holding singleton dispatcher lock (/home/gaetan/.hermes/...
2026-08-14 11:35:31,622 INFO gateway.run: Gateway housekeeping started (interval=60s)
2026-08-14 11:35:36,621 INFO gateway.run: kanban dispatcher: embedded in gateway (interval=60.0s)
```

Read: clean SIGTERM at 11:35:24 (systemd-initiated, matches the restart
command), orderly shutdown (slack + a2a disconnected, SessionDB closed,
teardown 0.76s), then a fresh startup sequence at 11:35:29: "Starting Hermes
Gateway...", slack connected, a2a connected, "Gateway running with 2
platform(s)", channel directory + kanban dispatcher initialized. No startup
errors.

### `~/.hermes/logs/errors.log` (tail 20, spans pre- and post-restart)

```
2026-08-14 11:31:16,022 WARNING [cron_62e8cd9db637_20260814_133038] tools.registry: check_fn check_image_generation_r...
2026-08-14 11:31:16,049 WARNING [cron_62e8cd9db637_20260814_133038] tools.registry: check_fn check_web_api_key return...
2026-08-14 11:32:01,208 WARNING tools.registry: check_fn check_bfl_requirements returned False; dependent tools will ...
2026-08-14 11:32:01,222 WARNING tools.registry: check_fn check_computer_use_requirements returned False; dependent to...
2026-08-14 11:32:01,225 WARNING tools.registry: check_fn check_image_generation_requirements returned False; dependen...
2026-08-14 11:32:01,242 WARNING tools.registry: check_fn check_web_api_key returned False; dependent tools will be un...
2026-08-14 11:32:03,283 WARNING [cron_62e8cd9db637_20260814_133038] agent.tool_executor: Tool execute_code returned e...
2026-08-14 11:32:17,718 WARNING [20260814_112121_d311e0fc] agent.tool_executor: Tool terminal returned error (0.06s):...
2026-08-14 11:32:20,283 WARNING [cron_62e8cd9db637_20260814_133038] agent.tool_executor: Tool execute_code returned e...
2026-08-14 11:32:28,681 WARNING [20260814_112820_7d1a68] agent.tool_executor: Tool skill_manage returned error (0.01s...
2026-08-14 11:32:32,875 WARNING [cron_62e8cd9db637_20260814_133038] agent.tool_executor: Tool execute_code returned e...
2026-08-14 11:33:02,310 WARNING [20260814_113257_60346fe7] agent.tool_executor: Tool kanban_show returned error (0.00...
2026-08-14 11:33:13,109 WARNING [20260814_113257_60346fe7] agent.tool_executor: Tool terminal returned error (0.01s):...
2026-08-14 11:33:55,129 WARNING tools.registry: check_fn check_bfl_requirements returned False; dependent tools will ...
2026-08-14 11:33:55,151 WARNING tools.registry: check_fn check_computer_use_requirements returned False; dependent to...
2026-08-14 11:33:55,152 WARNING cron.scheduler: Job '62e8cd9db637': origin has thread_id=1786191017.826669 but delive...
2026-08-14 11:33:55,156 WARNING tools.registry: check_fn check_image_generation_requirements returned False; dependen...
2026-08-14 11:33:55,174 WARNING tools.registry: check_fn check_web_api_key returned False; dependent tools will be un...
2026-08-14 11:35:24,873 WARNING gateway.run: Shutdown context: signal=SIGTERM under_systemd=yes parent_pid=27136 pare...
2026-08-14 11:35:29,402 WARNING slack_bolt.AsyncApp: As you gave `client` as well, `token` will be unused.
```

Read: everything before 11:35:24 is pre-existing tool-capability-check noise
(unrelated to the restart — `check_bfl_requirements` / computer-use /
image-generation / web-api-key checks returning False for unconfigured
optional tool integrations, plus unrelated tool-executor errors from earlier
sessions/cron jobs). The only line after the restart (11:35:29) is a benign
`slack_bolt.AsyncApp` SDK warning ("As you gave `client` as well, `token`
will be unused") — a known harmless duplicate-init notice, not a startup
failure. No new errors attributable to the restart itself.

## 5. Liveness check

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes chat -Q -q "reply with the single word: alive"'
session_id: 20260814_113553_883920
alive
EXIT=0
```

Answered correctly, well within the 120s timeout budget (ran in a 130s wall
clock allowance, exit 0, no hang).

## Verdict

Gateway restarted successfully and is healthy:
- `ActiveEnterTimestamp` moved forward, `MainPID` changed, `ActiveState=active/running`.
- Clean SIGTERM → orderly shutdown → fresh startup in the log, both platforms
  (slack, a2a) reconnected, no startup errors.
- Liveness oneshot answered "alive" correctly.

No secrets were read, echoed, or written. No config/`.env` edits made. No
`sops -d` invoked. 192.168.0.3 and 192.168.0.8 untouched.
