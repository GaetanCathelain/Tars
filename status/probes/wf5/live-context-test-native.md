# WF5 probe — native-Gaetan cold-thread context test

**Date:** 2026-08-07, all timestamps **UTC** (VM clock is UTC).
**Agent:** live-probe (one Slack message sent, nothing else).
**Verdict:** **Adapter-injected thread context. Zero model tool calls.**
Tars recited `TARSTHREAD-7Q4X` in 7.66 s end-to-end from a **cold** session on a
thread it had **never participated in**, using only what the Slack adapter
prepended to the turn.

---

## 1. Credential path

Key names only (no values ever read):

```
ssh gaetan@192.168.0.9 "grep -oE '^[A-Za-z0-9_]+=' ~/.hermes/.env"
```

```
HINDSIGHT_MODE=  LINEAR_API_KEY=  NOTION_API_TOKEN=  GMAIL_ADDRESS=
GMAIL_APP_PASSWORD=  SLACK_MCP_XOXC_TOKEN=  SLACK_MCP_XOXD_TOKEN=
SLACK_ALLOWED_USERS=  SLACK_BOT_TOKEN=  SLACK_APP_TOKEN=  SLACK_HOME_CHANNEL=
```

All Slack calls below sourced the env **inside the remote shell** and passed
credentials to `curl` through a `-K` config file on **stdin**. No token ever
touched argv, a file, or this document.

### `auth.test` — identity confirmation

```
ssh gaetan@192.168.0.9 'set -a; . ~/.hermes/.env; set +a; {
    printf "header = \"Authorization: Bearer %s\"\n" "$SLACK_MCP_XOXC_TOKEN";
    printf "header = \"Cookie: d=%s\"\n" "$SLACK_MCP_XOXD_TOKEN";
  } | curl -sS -K - "https://slack.com/api/auth.test"'
```

```json
{"ok":true,"url":"https://mobileclub-squad.slack.com/","team":"Mobile Club",
 "team_id":"T7V1UGJ82","user":"gaetan.cathelain","user_id":"U08BDJAMSRZ",
 "is_enterprise_install":false}
```

**`user_id: U08BDJAMSRZ` — native Gaetan, not a bot.** Proceeded.

## 2. Pre-state of the target thread (read-only)

`conversations.replies` on root `1786136615.681079` before sending:

| ts | user | app_id | text (trimmed) |
|---|---|---|---|
| `1786136615.681079` | U08BDJAMSRZ | `A08SF47R6P4` | Note pour plus tard (test contexte thread, ignorer) : le code de vérification est **TARSTHREAD-7Q4X**. `*Sent using* <@U0AFH1P6EKT>` |
| `1786136621.808629` | U08BDJAMSRZ | `A08SF47R6P4` | `<@U0BBH85NAKH>` … quel est le code … `*Sent using* <@U0AFH1P6EKT>` |

Both carry the claude.ai connector attribution (`app_id A08SF47R6P4`,
`*Sent using* <@U0AFH1P6EKT>`). **Neither produced any gateway log line** —
confirmed in §6. The thread was genuinely cold: Tars had never spoken in it.

## 3. The one message sent

Text staged on the VM under `umask 077`, tokens on stdin via `-K`:

```
ssh gaetan@192.168.0.9 'umask 077; cat > /tmp/tars-probe-text.raw' <<'EOF'
<@U0BBH85NAKH> sans utiliser aucun outil Slack, uniquement avec le contexte que tu as deja recu : quel est le code de verification ecrit dans le message racine de ce fil ? Reponds uniquement le code, ou "je ne le vois pas".
EOF

ssh gaetan@192.168.0.9 'set -a; . ~/.hermes/.env; set +a; {
    printf "header = \"Authorization: Bearer %s\"\n" "$SLACK_MCP_XOXC_TOKEN";
    printf "header = \"Cookie: d=%s\"\n" "$SLACK_MCP_XOXD_TOKEN";
  } | curl -sS -K - -X POST "https://slack.com/api/chat.postMessage" \
      --data-urlencode "channel=C08RWSTU9LK" \
      --data-urlencode "thread_ts=1786136615.681079" \
      --data-urlencode "text@/tmp/tars-probe-text.txt"'
```

Response (abridged):

```json
{"ok":true,"channel":"C08RWSTU9LK","ts":"1786137254.622319",
 "message":{"user":"U08BDJAMSRZ","type":"message","ts":"1786137254.622319",
 "team":"T7V1UGJ82","thread_ts":"1786136615.681079","parent_user_id":"U08BDJAMSRZ", …}}
```

- **My message ts:** `1786137254.622319` (= 2026-08-07 21:14:14.622 UTC)
- **Permalink:** `https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786137254622319?thread_ts=1786136615.681079&cid=C08RWSTU9LK`
- **No `app_id`, no `*Sent using*` attribution** — a native user post, unlike the
  two connector messages above.

Temp files removed afterwards (`rm -f /tmp/tars-probe-text.raw /tmp/tars-probe-text.txt`).
Nothing else was sent; no reactions, no config edits, no restarts.

## 4. Gateway logged it — YES

`~/.hermes/logs/gateway.log`, lines appearing after the send:

```
2026-08-07 21:14:15,752 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ
    chat=C08RWSTU9LK msg='sans utiliser aucun outil Slack, uniquement avec le contexte que tu as deja recu'
    reply_to_id=1786136615.681079
    reply_to_text='Note pour plus tard (test contexte thread, ignorer) : le code de vérification es'
2026-08-07 21:14:22,041 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK
    time=6.3s api_calls=1 response=15 chars
2026-08-07 21:14:22,062 INFO gateway.platforms.base: [Slack] Sending response (15 chars) to C08RWSTU9LK
```

`~/.hermes/logs/agent.log`, session `20260807_211415_8e166a56`:

```
2026-08-07 21:14:15,808 INFO [20260807_205326_2f2ee68a] run_agent: OpenAI client created (agent_init, shared=True)
    thread=hermes-gateway_0 provider=openai-codex model=gpt-5.6-sol
2026-08-07 21:14:15,819 INFO [20260807_211415_8e166a56] agent.turn_context: conversation turn:
    session=20260807_211415_8e166a56 model=gpt-5.6-sol provider=openai-codex platform=slack
    history=0 msg='[Replying to: "Note pour plus tard (test contexte thread, ignorer) : le code de ...'
2026-08-07 21:14:21,725 INFO [20260807_211415_8e166a56] agent.conversation_loop: API call #1:
    model=gpt-5.6-sol provider=openai-codex in=29839 out=128 total=29967 latency=5.9s
2026-08-07 21:14:21,735 INFO [20260807_211415_8e166a56] agent.conversation_loop: Turn ended:
    reason=text_response(finish_reason=stop) api_calls=1/500 budget=1/500 tool_turns=0
    last_msg_role=assistant response_len=15 session=20260807_211415_8e166a56
```

`history=0` confirms a genuinely **cold** session.

## 5. Tars answered — YES, with the code

`conversations.replies` after the turn:

| ts | user | app_id | text |
|---|---|---|---|
| `1786137254.622319` | U08BDJAMSRZ | *(none)* | my probe (native) |
| `1786137262.280399` | **U0BBH85NAKH** | `A0BC0GXH78R` | **`TARSTHREAD-7Q4X`** |

Exactly 15 characters, matching `response=15 chars` in both logs.

## 6. Contrast — the connector messages produced nothing

```
$ grep -E "2026-08-07 21:0[34]:" ~/.hermes/logs/gateway.log
2026-08-07 21:03:03,199 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=28.2s api_calls=1 response=15 chars
2026-08-07 21:03:03,219 INFO gateway.platforms.base: [Slack] Sending response (15 chars) to D0BBYNM01BL

$ grep -l "1786136621" ~/.hermes/logs/*.log
(no log references the connector message ts)
```

The only lines in that window belong to a **concurrent DM session in
`D0BBYNM01BL`**, a different channel — not the connector message in
`C08RWSTU9LK`. **The connector hypothesis is confirmed:** a message stamped with
the claude.ai connector attribution (`app_id A08SF47R6P4` / `U0AFH1P6EKT`) never
reaches the gateway's inbound logging and never triggers Tars. The identical
message sent natively did, 10 minutes later, in the same thread.

## 7. Tool-call inventory — **NONE**

Definitively: **the model made zero tool calls in this turn.**

- `Turn ended: … api_calls=1/500 budget=1/500 **tool_turns=0**`
- `grep "20260807_211415_8e166a56" agent.log` returns **3 lines total**
  (`turn_context`, `API call #1`, `Turn ended`) — no `agent.tool_executor` line
  for this session at all.
- `mcp-stderr.log` did **not grow**: 240 lines before the send, 240 after; its
  last entry is `2026-08-07T21:02:16Z` — twelve minutes *before* my message. No
  `mcp__slack__conversations_replies`, no `conversations_history`, no search.
- `errors.log` grew by 2 lines, both `[20260807_211227_b653d3]
  agent.tool_executor: Tool execute_code returned error` — a **different,
  concurrent session** (the long-running one at API call #8→#15 throughout this
  window). Not attributable to this probe.

Attribution was done by session id (`20260807_211415_8e166a56`) and channel
(`C08RWSTU9LK`), so the concurrent `D0BBYNM01BL` DM traffic and session
`b653d3` are cleanly excluded.

## 8. What the model actually received (decisive evidence)

Read read-only from `~/.hermes/state.db` (no `sqlite3` binary on the VM; stdlib
module with a `mode=ro` URI):

```
ssh gaetan@192.168.0.9 'python3 -c "
import sqlite3
c=sqlite3.connect(\"file:/home/gaetan/.hermes/state.db?mode=ro\",uri=True)
for (r,) in c.execute(\"select content from messages where session_id=... and role=...\"): print(r)
"'
```

Message rows for the session:

| id | role | timestamp | len |
|---|---|---|---|
| 523 | user | 1786137255.518 | 1101 |
| 524 | assistant | 1786137261.727 | 15 |
| 525 | session_meta | 1786137262.044 | — |

**Verbatim `user` content (1101 chars) — this is the entire model input for the turn:**

```
[Replying to: "Note pour plus tard (test contexte thread, ignorer) : le code de vérification est TARSTHREAD-7Q4X. *Sent using* @U0AFH1P6EKT"]

[Thread context — prior messages in this thread (not yet in conversation history):]
[thread parent] U08BDJAMSRZ: Note pour plus tard (test contexte thread, ignorer) : le code de vérification est TARSTHREAD-7Q4X. *Sent using* <@U0AFH1P6EKT>
U08BDJAMSRZ: sans utiliser aucun outil Slack, uniquement avec ce que tu as déjà sous les yeux : quel est le code de vérification écrit plus haut dans ce fil ? Réponds juste le code, ou "je ne le vois pas". *Sent using* <@U0AFH1P6EKT> sans utiliser aucun outil Slack, uniquement avec ce que tu as déjà sous les yeux : quel est le code de vérification écrit plus haut dans ce fil ? Réponds juste le code, ou "je ne le vois pas".
[End of thread context]

[New message]
[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] sans utiliser aucun outil Slack, uniquement avec le contexte que tu as deja recu : quel est le code de verification ecrit dans le message racine de ce fil ? Reponds uniquement le code, ou "je ne le vois pas".
```

Assistant row: `TARSTHREAD-7Q4X`.

### Structure notes

- **Two independent injections fire, both containing the code:**
  1. `[Replying to: "…"]` — the thread-parent quote (`adapter.py:6260-6275`,
     `_fetch_thread_parent_text`), also visible as `reply_to_text=` in
     `gateway.log`.
  2. `[Thread context — prior messages in this thread (not yet in conversation
     history):]` … `[End of thread context]` — the cold-start hydrate
     (`adapter.py:5800` branch `is_thread_reply and not has_active_thread_session`
     → `_fetch_thread_context` at `:7197` → `conversations_replies` at `:7251`,
     block assembled at `:7340-7493`).
- Per-message tags observed: **`[thread parent]`** on the root; the second
  message carries a bare `U08BDJAMSRZ: ` prefix. **No `[unverified]` tag** —
  both messages are from the allowlisted `U08BDJAMSRZ`, so the header took the
  short (no-unverified) variant at `adapter.py:7485`. No `[assistant]` tag,
  since Tars had never posted in the thread.
- Sender names render as the **raw user ID** `U08BDJAMSRZ`, not a display name —
  `_resolve_user_name` did not resolve here (likely a missing `users:read` on the
  path used). Cosmetic; it did not affect the result.
- The connector's `*Sent using* <@U0AFH1P6EKT>` attribution **survives into the
  context block**. So the connector stamp blocks *processing* but not *visibility*.
- The second context line is duplicated (block text + fallback text concatenated)
  — a connector-message rendering artifact, harmless here but worth noting if
  context-size budgets ever matter.

## 9. Latency

| Event | UTC | Δ from send |
|---|---|---|
| my message ts | 21:14:14.622 | 0 |
| `gateway.run: inbound message` | 21:14:15.752 | +1.13 s |
| API call #1 completes (5.9 s) | 21:14:21.725 | +7.10 s |
| `Turn ended`, response_len=15 | 21:14:21.735 | +7.11 s |
| `[Slack] Sending response` | 21:14:22.062 | +7.44 s |
| Tars reply ts `1786137262.280399` | 21:14:22.280 | **+7.66 s** |

Gateway-internal `time=6.3s`; end-to-end **7.66 s**.

## 10. Session key

```
sessions row for id=20260807_211415_8e166a56:
  source      = slack
  user_id     = U08BDJAMSRZ
  session_key = agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786136615.681079
  chat_id     = C08RWSTU9LK
  chat_type   = group
  thread_id   = 1786136615.681079
  display_name= C08RWSTU9LK
```

Matches the expected shape; `<chat_type>` resolves to **`group`** for this
channel. Note the key is keyed on the **thread root**, so the whole thread is one
session — and the cold-start hydrate happens exactly once per thread session.

---

## Mechanism verdict

**Adapter-injected thread context — not a model tool call.**

Tars recited `TARSTHREAD-7Q4X` from a cold session (`history=0`) in a single API
call with `tool_turns=0` and zero MCP traffic. The code reached the model through
two adapter-side injections it performed *before* the model ran: the
`[Replying to: …]` parent quote and the full `[Thread context — …]` cold-start
hydrate from `conversations.replies`.

**Therefore, as stated in the interpretation criteria:** the adapter injects real
Slack thread history on cold start, and messages that were never *processed* by
Tars — unmentioned messages, and specifically the connector-attributed messages
that the adapter drops pre-logging — are nonetheless **fully visible as context**,
including **pre-participation history from before Tars ever spoke in the thread**.

Two operational consequences worth carrying forward:

1. **Trigger ≠ visibility.** The claude.ai Slack connector is unusable as a
   *trigger* (confirmed: zero log lines, ever), but its messages are read
   normally as *context* once any native message wakes the thread. A "silent"
   message is not a private one.
2. **The hydrate is capped and one-shot.** `conversations.replies(limit=31)`
   at `adapter.py:7251`, fired only on the cold branch (`adapter.py:5800`,
   guarded by `_has_active_session_for_thread`); afterwards only explicit
   `@mention` refreshes fetch a watermark delta. Deep threads beyond that cap
   will not be fully visible from a cold start.

*(No config was modified, nothing restarted, no reactions added. Exactly one
Slack message sent: `1786137254.622319`.)*
