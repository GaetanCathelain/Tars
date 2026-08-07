# Log forensics — reference thread C08RWSTU9LK / 1786135149.072829 (2026-08-07)

Read-only investigation on cooper against the Tars VM (`ssh gaetan@192.168.0.9`).
No files edited, no services restarted, no Slack messages sent. VM clock is UTC.

## Commands run

```bash
ssh gaetan@192.168.0.9 'ls -la ~/.hermes/logs/'
ssh gaetan@192.168.0.9 'head -1 / tail -1 ~/.hermes/logs/{gateway,agent,errors,mcp-stderr}.log'
ssh gaetan@192.168.0.9 "awk -F'[ ,]' '\$2>=\"20:38:00\" && \$2<=\"20:46:00\"' ~/.hermes/logs/gateway.log"
ssh gaetan@192.168.0.9 "awk -F'[ ,]' '\$2>=\"20:38:00\" && \$2<=\"20:46:00\"' ~/.hermes/logs/errors.log"
ssh gaetan@192.168.0.9 "grep -F '1786135270' / '1786135310' / '1786135439' / '1786135219' / '1786135242' / '1786135283' ~/.hermes/logs/*.log"
ssh gaetan@192.168.0.9 "grep -n 'U05LKHLDV0A' ~/.hermes/logs/{gateway,errors,agent,mcp-stderr}.log"
ssh gaetan@192.168.0.9 'grep -riE "log_level|LOG_LEVEL|logging" ~/.hermes/config.yaml ~/.hermes/.env'
ssh gaetan@192.168.0.9 'systemctl --user show hermes-gateway.service -p Environment'
ssh gaetan@192.168.0.9 'grep -iE "^level|debug|verbose" ~/.hermes/config.yaml'
# source reading (read-only, no secrets):
ssh gaetan@192.168.0.9 'sed -n "5160,5775p" ~/.hermes/hermes-agent/plugins/platforms/slack/adapter.py'
ssh gaetan@192.168.0.9 'sed -n "8244,8320p" ~/.hermes/hermes-agent/plugins/platforms/slack/adapter.py'
ssh gaetan@192.168.0.9 'grep -n -iE "require_mention|strict_mention|thread_require" ~/.hermes/config.yaml'
```

## 1. Raw lines, window 2026-08-07 20:38:00–20:46:00 UTC (`gateway.log`)

```
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
```

`errors.log` for the same window is a strict subset of the WARNING lines above
(plus unrelated tool-registry warnings around 20:39:15–20:39:19); `agent.log`
carries the same WARNING/INFO lines interleaved with per-turn agent trace.
No line in any file contains the raw Slack `ts` values
(`1786135219`, `1786135242`, `1786135270`, `1786135283`, `1786135310`,
`1786135439`) — the gateway logs human-readable `%Y-%m-%d %H:%M:%S,fff`
timestamps only, never the Slack float ts, so `grep -F` on the ts is a dead
end by design, not a gap.

Extra context outside the given table: a second unauthorized user,
`U7XJ4K631`, was also early-rejected at 20:38:19 in the same channel — not
one of the six events, noted for completeness.

## 2. The two dropped Gaetan messages (20:41:10 "Ptin t'es nul", 20:41:50 "Attends que je te tune toi")

**No log line at all, at any level, in any of the four log files.** Not
silent-because-DEBUG-is-off — genuinely no `logger` call fires for this code
path, confirmed by reading `plugins/platforms/slack/adapter.py` on the VM
(`_handle_slack_message`, `~line 5707`):

```python
elif self._slack_strict_mention() and not is_mentioned:
    return  # Strict mode: ignore until @-mentioned again
```

`config.yaml` has `strict_mention: true` (alongside `require_mention: true`,
line ~123). Per `_slack_strict_mention()`'s docstring: "When true, channel
threads require an explicit @-mention on every message. **Disables all
auto-triggers** (mentioned-thread memory, bot-message follow-up,
session-presence)." With this flag on, the `elif` above fires first and
`return`s bare — no comment-adjacent logger call, nothing to raise even at
DEBUG. The two later branches that *do* log (`thread_require_mention`'s
`logger.debug("[Slack] Ignoring thread reply without mention …")`, and the
`_should_wake_on_unmentioned_message()` no-wake path) are unreachable here
because the `strict_mention` branch is earlier in the `elif` chain and
already returned. So: **hard-silent by code, not just by log level.**

There IS one DEBUG line that fires on every inbound event before any gate
(`"[Slack] event received type=… user=… channel=… ts=… thread_ts=…"`,
guarded by `logger.isEnabledFor(logging.DEBUG)`), but it never reached the
file either way, see §4.

## 3. The Nans message (U05LKHLDV0A, no mention, 20:43:59 UTC per Slack ts)

**Yes — the allowlist WARNING fired.** Line found (twice, see below):

```
2026-08-07 20:44:01,244 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
```

That's 2s after the message's Slack ts (20:43:59.766), consistent with the
~1–2s inbound-log lag seen on the three mentioned/replied messages. A second,
earlier line for the same user fired at 20:43:23,301 (36s before), and two
more later at 20:46:05 and 20:46:09 — `grep -n U05LKHLDV0A` across the whole
log set returns 5 hits total for this session (plus one unrelated one at
20:18:16). The task's table lists one Nans message; the log shows the
allowlist gate tripping on Nans multiple times in this window, so either
Nans sent more than the one message quoted, or the events replayed — logs
don't disambiguate that, only that the reject fired every time.

**This settles the ordering question: the allowlist gate runs BEFORE the
mention gate**, unconditionally, for every non-authorized sender — mention
status never enters into it. Confirmed from source, not just log absence:
in `_handle_slack_message` the auth check

```python
# Reject unauthorized users before thread lookups, name resolution,
# or file downloads. ...
if user_id and callable(_auth_fn):
    ...
    if not _auth_fn(_source):
        logger.warning(
            "[Slack] Early reject of unauthorized user %s in channel %s",
            user_id, channel_id,
        )
        return
```

sits at ~line 5546, well before `is_mentioned` is even computed (~line
5646) and before the `strict_mention` branch that silently drops Gaetan's
unmentioned messages (~line 5707). So: Gaetan (allowlisted) sails past the
allowlist gate and only then hits the silent mention gate; Nans
(non-allowlisted) never reaches the mention gate at all — rejected earlier,
loudly, at WARNING, regardless of whether he mentioned the bot.

## 4. Current gateway log level

**INFO — no DEBUG override anywhere.**

- `~/.hermes/config.yaml` has no `logging:` section at all (`grep -n -A5
  "^logging:"` → no match).
- `hermes_logging.py`: `level_name = (log_level or cfg_level or "INFO").upper()`
  — with no config override, defaults to `INFO`.
- `config.yaml` does have `verbose: false`, but that key drives
  `setup_verbose_logging()` (console-only `-v`/`--verbose` CLI mode), not the
  `agent.log`/`gateway.log` file handler level — confirmed by reading the
  function bodies; it's unrelated to what lands in the log files.
- `systemctl --user show hermes-gateway.service -p Environment` shows no
  `LOG_LEVEL`/`HERMES_LOG_LEVEL` env var set.

So the entry-log DEBUG line (`"[Slack] event received type=…"`, §2) is
suppressed by level regardless of code path — but it's moot for the two
drops in question, since the actual drop point (`strict_mention` branch) has
no logger call in the source at all, DEBUG or otherwise.

## 5. Log window coverage

| File | Earliest | Latest | Covers 20:38–20:46 UTC? |
|---|---|---|---|
| `gateway.log` | 2026-08-07 18:43:11 | 2026-08-07 20:46:55 | Yes |
| `agent.log` | 2026-08-07 16:20:29 | 2026-08-07 20:48:09 | Yes |
| `errors.log` | 2026-08-07 16:46:56 | 2026-08-07 20:46:09 | Yes |
| `mcp-stderr.log` | 2026-08-07 17:51:27 | 2026-08-07 20:39:32 | **No** — last entry 20:39:32, before events 3–6 (20:41:10 onward) |

No rotation past the window on the files that matter (`gateway.log`,
`errors.log`, `agent.log` — the ones the Slack adapter and gateway runner
actually write to). `mcp-stderr.log` is the Slack MCP subprocess's own
stderr (tool-call plumbing, not the inbound-message/mention/allowlist gate)
and stops at 20:39:32 — it simply isn't the right log for this question, but
flagging since the task asked to check coverage on "the logs" generally.

## Per-event verdict table

| Event | Slack ts (UTC) | Sender | Mention? | Observed | Log line found | Verdict |
|---|---|---|---|---|---|---|
| 1 | 20:40:19 | Gaetan | yes | REPLIED 20:40:25 | inbound @20:40:20 → response ready @20:40:25 (INFO) | matches |
| 2 | 20:40:42 | Gaetan | yes | REPLIED 20:40:59 | inbound @20:40:54 → response ready @20:40:59 (INFO) | matches (12s inbound-log lag vs Slack ts, unexplained but harmless) |
| 3 | 20:41:10 | Gaetan | no | DROPPED | **none, any level** | confirmed: `strict_mention=true` bare `return`, no logger call in source |
| 4 | 20:41:23 | Gaetan | yes | REPLIED 20:41:29 | inbound @20:41:25 → response ready @20:41:29 (INFO) | matches |
| 5 | 20:41:50 | Gaetan | no | DROPPED | **none, any level** | same silent `strict_mention` path as event 3 |
| 6 | 20:43:59 | Nans (U05LKHLDV0A) | no | NO reply | WARNING `[Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK` @20:44:01 (also @20:43:23, 20:46:05, 20:46:09) | allowlist gate fires, independent of mention — proves allowlist runs BEFORE mention gate |
