# WF4 Probe 2 — mention-gating in a channel: **PASS** (gating axis)

Spec: `docs/specs/wf4-probes.md` §2 (post-cutover only). Channel **C08RWSTU9LK**,
workspace `mobileclub-squad` (team `T7V1UGJ82`). All timestamps `date -Is`, UTC,
2026-08-07 19:0x window. Read path = Gaetan's personal xoxc/xoxd from `~/.hermes/.env`
on the VM, fed to `curl -K -` on stdin (never argv, never printed).

**Verdict: PASS.** Mentioned → dispatched and replied. Un-mentioned → silently dropped,
no dispatch, no reply, gateway provably alive throughout.

**Adjacent RED, not probe 2's:** the reply *body* to the mention is a model-failure
notice — `gpt-5.6-sol` returned empty content 3/3 retries. That is **probe 10 / WF3-owned
(A1b, model wiring)**. It does not change probe 2's verdict: the gate, the dispatch and
the outbound Slack write all worked; only the model's content was empty.

---

## Identities pinned first (spec §2 FAIL branch: "bot was simply never invited")

```
$ curl -s -K - "https://slack.com/api/conversations.members?channel=C08RWSTU9LK"
{"ok":true,"members":["U7XJ4K631","U05LKHLDV0A","U08BDJAMSRZ","U08NR5CM86Q",
                      "U0BBH85NAKH","U0BDA9P7528"]}

$ curl -s -K - "https://slack.com/api/users.info?user=U0BBH85NAKH"
id=U0BBH85NAKH name=tars real=Tars is_bot=True app_id=A0BC0GXH78R bot_id=B0BBSBVBJUB
```

- **U0BBH85NAKH = @Tars** (app `A0BC0GXH78R`, bot `B0BBSBVBJUB`) — **is a channel member**.
- **U08BDJAMSRZ = Gaetan** (sender of both probe messages).
- Gateway confirms the same identity at connect:
  `18:53:55 INFO …slack_platform.adapter: [Slack] Authenticated as @tars in workspace Mobile Club (team: T7V1UGJ82)`

## Half A — WITH mention → Tars replies (sent by Gaetan, human half)

| item | value |
|---|---|
| mention ts | `1786129267.292069` (2026-08-07 19:01:07Z) |
| text | `<@U0BBH85NAKH> hello there` |
| permalink | https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786129267292069 |
| thread | `reply_count=5`, all replies from `U0BBH85NAKH` / `B0BBSBVBJUB` / app `A0BC0GXH78R` |

`conversations.replies channel=C08RWSTU9LK ts=1786129267.292069`:

```
1786129267.292069 user=U08BDJAMSRZ  '<@U0BBH85NAKH> hello there'
1786129330.268889 user=U0BBH85NAKH  ':warning: Empty response from model — retrying (2/3) in 13s'
1786129330.270069 user=U0BBH85NAKH  ':warning: Empty response from model — retrying (3/3) in 21s'
1786129330.270839 user=U0BBH85NAKH  ':warning: Empty response from model — retrying (1/3) in 6s'
1786129330.273399 user=U0BBH85NAKH  ':x: Model returned no content after all retries. No fallback providers configured.'
1786129330.653129 user=U0BBH85NAKH  ':warning: No reply: the model returned empty content after retries and any
                                      fallback providers. Try `continue`, switch model/provider, or inspect the tool
                                      output above.'
```

Reply permalink (last message of the thread, ts `1786129330.653129`):
https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786129330653129?thread_ts=1786129267.292069&cid=C08RWSTU9LK

### Matching gateway lines — inbound event AND outbound reply

`~/.hermes/logs/gateway.log` (INFO level; the journal only carries WARNING+):

```
2026-08-07 19:01:09,377 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg='hello there' reply_to_id=None reply_to_text=''
2026-08-07 19:02:10,351 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=61.0s api_calls=5 response=160 chars
2026-08-07 19:02:10,409 INFO gateway.platforms.base: [Slack] Sending response (160 chars) to C08RWSTU9LK
```

Note `msg='hello there'` — the `<@U0BBH85NAKH>` prefix is stripped by the gate before
dispatch, which is the gate visibly doing its job on the accept path.

`journalctl --user -u hermes-gateway.service` for the same window (why the reply body is
empty — probe 10's problem, recorded here for the record):

```
2026-08-07T19:01:10Z python[61969]: WARNING tools.registry: check_fn check_bfl_requirements returned False; …
2026-08-07T19:01:21Z python[61969]: WARNING agent.conversation_loop: Empty response (no content or reasoning) — retry 1/3 in 5.8s (model=gpt-5.6-sol)
2026-08-07T19:01:30Z python[61969]: WARNING agent.conversation_loop: Empty response (no content or reasoning) — retry 2/3 in 13.3s (model=gpt-5.6-sol)
2026-08-07T19:01:46Z python[61969]: WARNING agent.conversation_loop: Empty response (no content or reasoning) — retry 3/3 in 21.1s (model=gpt-5.6-sol)
2026-08-07T19:02:10Z python[61969]: WARNING agent.conversation_loop: Empty response (no content or reasoning) after 3 retries. No fallback available. model=gpt-5.6-sol provider=openai-codex
```

**Half A = PASS** — mention accepted, dispatched, agent turn ran, outbound Slack write
landed in-thread.

## Half B — WITHOUT mention → silence (control, sent by this probe as Gaetan)

Sent via `chat.postMessage` with the personal xoxc token (so the sender is Gaetan
`U08BDJAMSRZ`, same human identity as half A — the only variable changed is the mention):

```
$ date -Is
2026-08-07T19:02:56+00:00
$ … | curl -s -K - -X POST https://slack.com/api/chat.postMessage \
      --data-urlencode channel=C08RWSTU9LK \
      --data-urlencode "text=checking in without mention — wf4 probe 2, expect silence"
ok=True ts=1786129376.997239 channel=C08RWSTU9LK
```

Permalink: https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786129376997239

**Wait window: 19:02:56Z → 19:06:28Z (3m32s, ≥ the required 3 minutes).**

### Evidence 1 — no reply exists in Slack

`conversations.history oldest=1786129376.997239 inclusive=true`, read at
**2026-08-07T19:06:37Z**:

```
ok=True count=1
1786129376.997239 user=U08BDJAMSRZ bot_id=None thread_ts=None reply_count=None
  'checking in without mention — wf4 probe 2, expect silence'
```

One message, mine. `reply_count=None` → no thread was opened on it either. No message
from `U0BBH85NAKH`/`B0BBSBVBJUB` anywhere in the window.

### Evidence 2 — the gateway never even dispatched it (explicit, not just absence)

`~/.hermes/logs/gateway.log` after 19:02:56 is **empty**; its last line is the half-A
outbound at `19:02:10,409`. Critically there is **no `inbound message: … chat=C08RWSTU9LK`
line for the control message at all** — the Slack adapter dropped it at the mention gate
*before* the dispatch logger, so this is a positive statement about the gate's behaviour,
not merely "nothing happened".

```
$ date -Is
2026-08-07T19:06:51+00:00
$ tail -2 ~/.hermes/logs/gateway.log
2026-08-07 19:02:10,351 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=61.0s api_calls=5 response=160 chars
2026-08-07 19:02:10,409 INFO gateway.platforms.base: [Slack] Sending response (160 chars) to C08RWSTU9LK
```

The drop is silent by design — `grep -i "mention|ignor|skip"` over `gateway.log` and
`agent.log` returns no gating line (only an unrelated `18:53:55 … skipping session
suspension`). Hence the pin in §"Guardrail surface" below carries the "why".

### Evidence 3 — the gateway was alive *inside* the window (silence ≠ dead process)

Unit never restarted across the whole probe (`MainPID=61969` identical before, during and
after; `ActiveEnterTimestamp` still `18:53:52 UTC`):

```
19:05:31Z   MainPID=61969  NRestarts=0  active  ActiveEnterTimestamp=Fri 2026-08-07 18:53:52 UTC
19:06:40Z   MainPID=61969  NRestarts=0  active  ActiveEnterTimestamp=Fri 2026-08-07 18:53:52 UTC
```

And it was actively *working* at **19:05:58Z**, i.e. inside the wait window — it fired
probe 13's cron job on the same process:

```
agent.log:1326  2026-08-07 19:05:58,147 INFO cron.scheduler: Running job 'wf4-p13' (ID: 004adebba8ee)
agent.log:1330  2026-08-07 19:05:58,197 INFO run_agent: OpenAI client created … provider=openai-codex model=gpt-5.6-sol
journal          2026-08-07T19:05:58Z python[61969]: WARNING tools.registry: check_fn check_bfl_requirements returned False; …
```

Same PID 61969, live scheduler tick, agent turns being created — while the un-mentioned
Slack message sat untouched. Plus the prior positive control: the same process consumed a
mention at 19:01:09 and a DM (`ping wf4-1`, probe 1) at 19:01:32 and answered it in 3.9s.

**Half B = PASS.**

---

## Guardrail surface pinned per the spec's probe trap (read-only)

### `SLACK_STRICT_MENTION` — NOT read off the parsed config object

The spec's `/proc/<pid>/environ` route **does not work here, and that is expected, not a
failure**: `/proc/<pid>/environ` is the *exec-time* environment, frozen at `execve`. The
guardrail is written into `os.environ` by the plugin *after* start, so it can never appear
there. Confirmed empirically:

```
$ PID=$(systemctl --user show -p MainPID --value hermes-gateway.service)   # 61969
$ tr '\0' '\n' < /proc/$PID/environ | cut -d= -f1 | sort | tr '\n' ' '
DBUS_SESSION_BUS_ADDRESS GSM_SKIP_SSH_AGENT_WORKAROUND HERMES_HOME HOME INVOCATION_ID
JOURNAL_STREAM LANG LOGNAME MANAGERPID MEMORY_PRESSURE_WATCH MEMORY_PRESSURE_WRITE PATH
QT_ACCESSIBILITY SHELL SSH_AUTH_SOCK SYSTEMD_EXEC_PID USER VIRTUAL_ENV XDG_DATA_DIRS
XDG_RUNTIME_DIR
$ systemctl --user show-environment | grep -i slack   →  (none)
```

(949 bytes, readable — no `SLACK_*` at all. The unit only sets `PATH`, `VIRTUAL_ENV`,
`HERMES_HOME`; no `EnvironmentFile=`.)

The exporter is `plugins/platforms/slack/adapter.py::_apply_yaml_config`, whose docstring
states the contract outright — *"The SlackAdapter reads its runtime configuration via
`os.getenv()` throughout the connect / handle code paths … this hook keeps the existing
env-driven model and owns the YAML→env translation here"*:

```python
# adapter.py:8992-8995
if "require_mention" in slack_cfg and not os.getenv("SLACK_REQUIRE_MENTION"):
    os.environ["SLACK_REQUIRE_MENTION"] = str(slack_cfg["require_mention"]).lower()
if "strict_mention" in slack_cfg and not os.getenv("SLACK_STRICT_MENTION"):
    os.environ["SLACK_STRICT_MENTION"] = str(slack_cfg["strict_mention"]).lower()
```

and the consumer defaults to **false** if unset (`adapter.py:8273`), so an unexported value
would silently disable the strict gate:

```python
return os.getenv("SLACK_STRICT_MENTION", "false").lower() in {"true","1","yes","on"}
```

**Pin taken by replaying the gateway's own exporter** against the live `config.yaml`, in a
throwaway read-only process (no write, no restart, running gateway untouched):

```
$ grep -c '^SLACK_STRICT_MENTION=\|^SLACK_REQUIRE_MENTION=' ~/.hermes/.env
0                              ← no env override, so the `not os.getenv(...)` guard passes
                                 and config.yaml's value is what got exported at 18:53:52

$ PYTHONPATH=~/.hermes/hermes-agent/plugins/platforms/slack \
  ~/.hermes/hermes-agent/venv/bin/python -c '<load adapter.py, call _apply_yaml_config
    with the live cfg["gateway"]["platforms"]["slack"]>'
SLACK_STRICT_MENTION = true
SLACK_REQUIRE_MENTION = true
SLACK_UNAUTHORIZED_DM_BEHAVIOR = None
```

`SLACK_UNAUTHORIZED_DM_BEHAVIOR = None` is **not** a gap in probe 2 — `_apply_yaml_config`
simply doesn't translate that key; `unauthorized_dm_behavior` is consumed elsewhere. It is
probe 3's surface, flagged here so probe 3 doesn't reuse this pin as if it covered it.

### No stray top-level `platforms:` (the silent-override footgun, lane B B4)

```
$ python3 -c 'import yaml; d=yaml.safe_load(open("/home/gaetan/.hermes/config.yaml")); …'
top-level keys: _config_version agent browser code_execution compression context database
  delegation display gateway group_sessions_per_user mcp_servers memory model
  platform_toolsets plugins prompt_caching session_reset skills streaming stt telemetry
  terminal tool_loop_guardrails updates
stray top-level 'platforms' present: False
top-level 'slack' present: False
gateway.platforms.slack: {'enabled': True, 'require_mention': True,
                          'strict_mention': True, 'unauthorized_dm_behavior': 'ignore'}
```

Nothing can silently override `gateway.platforms.slack`. The three guardrails are the
D2 values verbatim.

---

## Deviations

- **Env var names in the dispatch brief were wrong.** Brief said `SLACK_TOKEN` /
  `SLACK_COOKIE`; the live names on the VM are **`SLACK_MCP_XOXC_TOKEN`** /
  **`SLACK_MCP_XOXD_TOKEN`** (the A4 rename, `docs/specs/tars-profile.md` §155). Used the
  live names. Cookie sent as `Cookie: d=$SLACK_MCP_XOXD_TOKEN` as-stored (URL-encoded),
  which works — `ok:true` on every call.
- **`/proc/<pid>/environ` cannot carry `SLACK_STRICT_MENTION`** (exec-time snapshot).
  Substituted a stronger pin: replay of the gateway's *own* `_apply_yaml_config` exporter
  against the live `config.yaml` + proof `.env` holds no override. Still not a grep of the
  parsed config object — the spec's trap is respected.
- **No explicit "ignored/gated" log line exists** — the Slack adapter drops un-mentioned
  channel messages before the dispatch logger. The absence statement is therefore backed by
  (a) no `inbound message` line for the control ts, (b) in-window liveness at 19:05:58, and
  (c) the exported-guardrail pin, rather than by a gating log line the code never writes.
- **Gateway restarted at 18:53:52Z**, ~7 min before the probe (previous instance exited
  `status=1/FAILURE` during a SIGTERM shutdown). Pre-existing, not caused by this probe; no
  restart occurred during either half (`NRestarts=0`, `MainPID=61969` throughout).
- **Model-empty-response is carried out of this probe, not swallowed.** `gpt-5.6-sol` via
  `openai-codex` returned empty content 3/3 with no fallback. Routed to **probe 10 /
  WF3-owned (A1b)**.
- The 19:05:58Z in-window activity is **probe 13's cron job** (`wf4-p13`), a concurrent WF4
  agent — used here only as liveness evidence, nothing was done to it.
- No secret values read, printed or written anywhere. No `sops`. No unit stopped, started
  or restarted. No config edited. 192.168.0.3 untouched. No git.
