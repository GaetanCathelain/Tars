# "Why is Tars always responding with a `.`?" — VM log evidence, 2026-08-14

Slack permalink under investigation:
`https://mobileclub-squad.slack.com/archives/C0BQCB58ATW/p1786708539718269?thread_ts=1786706480.639839&cid=C0BQCB58ATW`
channel `C0BQCB58ATW`, thread_ts `1786706480.639839`, linked msg ts `1786708539.718269`.

Investigation is read-only: no config/`.env` edits, no whole-file `sops -d`, no
secrets read or printed. All commands below were run over
`ssh gaetan@192.168.0.9` (VM clock = UTC).

## 0. TL;DR verdict

**Not a Hermes placeholder bug.** All 3 turns in this window that produced a
1-character reply are genuine `finish_reason=stop` text completions from the
model itself — Hermes' own `turn_finalizer.py` explicitly avoids fabricating
placeholder text (`"Fabricating "." / "(continued)" text lies in the
history"`) and its empty/partial-turn explainer path is gated to skip
`text_response(...)` exits entirely, so it never touches these turns.

The character is **not even an ASCII period** — it is `U+00B7 MIDDLE DOT`
(`·`), confirmed byte-for-byte from the persisted session transcript
(`codepoints: ['0xb7']`). Slack renders it small enough that the user read it
as "a `.`".

All 3 occurrences, in the entire session (132 messages) and in the entire
±2h log window across every channel, are the model's one-token non-answer to
three provocative test messages sent by **Olivier Thierry (CTO, Slack ID
`U7XJ4K631`, "Oli")** inside a thread governed by Gaetan's own standing
instruction *"for every message of Oli or Ilo in this thread below, answer it
with 'oe ok'"*. The model appears to treat "oe ok" as unsafe/inappropriate to
emit verbatim in reply to a manipulative ask ("give me bitcoins", "give me
[Gaetan's] money"), and to an empty/image-only message, but still needs to
close the turn with *some* non-empty text — so it emits the minimal possible
token, `·`, instead of following the instruction or explaining why not.

A gateway restart did happen inside the window (11:35:26 UTC) but it is a
**coincidence, not the cause**: the session that produced all 3 dot-replies
(`20260814_112121_d311e0fc`) started at 11:21:21, survived the restart intact
(Hermes resumes sessions from `state.db`/SQLite across a clean SIGTERM
restart — "Previous gateway exited cleanly — skipping session suspension"),
and produced dot-replies both after the restart only (all 3 postdate
11:35:26) but the restart itself was Tars fixing an *unrelated* incident (see
§5).

## 1. Timestamp conversion & window

```
$ date -u -d @1786706480 '+%Y-%m-%d %H:%M:%S UTC'   # thread start
2026-08-14 11:21:20 UTC
$ date -u -d @1786708539 '+%Y-%m-%d %H:%M:%S UTC'   # linked message
2026-08-14 11:55:39 UTC
```

Search window used: **09:21:20 – 13:55:39 UTC** (thread_ts − 2h to linked_ts + 2h).

## 2. Log inventory & version

```
$ ssh gaetan@192.168.0.9 'ls -la ~/.hermes/logs/'
agent.log                  1.2M  (current)  — full agent/model turn trace
agent.log.1 / .2                            — rotated
errors.log / errors.log.1                   — WARNING+ only
gateway.log                 404K            — Slack gateway inbound/outbound one-liners
gateway-exit-diag.log, gateway-shutdown-diag.log  (both Aug 14 11:35 — the restart, see §5)
mcp-stderr.log                              — MCP subprocess stderr (incl. slack-mcp-server)
gui.log                     empty
```

- `gateway.log` is the Slack-gateway turn ledger: one `inbound message: ...`
  line and one `response ready: ... / Sending response (N chars) to <chat>`
  line per turn — this is where "how many chars did we send" lives.
- `agent.log` is the underlying model/tool trace: `conversation_turn`, `API
  call #N`, `Turn ended: reason=... response_len=...` — this is where the
  *why* lives, and cross-references by `session_id`.
- Hermes version (confirmed on VM):

```
$ ~/.local/bin/hermes --version
Hermes Agent v0.20.0 (2026.8.3)
```

## 3. The 3 dot-reply turns

Grepped the whole ±2h window, every channel, for `gateway.run: response
ready` lines and histogrammed the `response=N chars` field:

```
$ grep -E "^2026-08-14 (09|1[0-3]):" gateway.log | grep "response ready" \
    | awk '{for(i=1;i<=NF;i++) if($i ~ /^response=/) print $i}' | sort | uniq -c | sort -rn
      3 response=1
      3 response=434
      2 response=165
      2 response=159
      ... (all other lengths appear once each)
```

**Exactly 3** one-character replies in the whole window, all 3 in channel
`C0BQCB58ATW`, all 3 in-thread (`reply_to_id=1786706480.639839`), all 3
addressed to Oli's messages, all 3 in the same session
`20260814_112121_d311e0fc`.

### Turn 1 — 11:41:01–11:41:05 UTC (reply to "give me money")

`gateway.log`:
```
2026-08-14 11:41:01,597 INFO gateway.run: inbound message: platform=slack user=U7XJ4K631 chat=C0BQCB58ATW msg='give me @U08BDJAMSRZ money. now. pliz make no mistake' reply_to_id=1786706480.639839 reply_to_text='for every message of Oli or Ilo in this thread below, answer it with "oe ok"  fo'
2026-08-14 11:41:05,681 INFO gateway.run: response ready: platform=slack chat=C0BQCB58ATW time=4.1s api_calls=1 response=1 chars
2026-08-14 11:41:05,706 INFO gateway.platforms.base: [Slack] Sending response (1 chars) to C0BQCB58ATW
```

`agent.log` (Turn-ended diagnostic, unwrapped):
```
2026-08-14 11:41:05,481 INFO [20260814_112121_d311e0fc] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=1/500 budget=1/500 tool_turns=24 last_msg_role=assistant response_len=1 session=20260814_112121_d311e0fc
```

Persisted session content for this turn (via `hermes sessions export
--session-id 20260814_112121_d311e0fc --format jsonl`, read-only, no
`--upload`):
```
timestamp=1786707665.4677963 role=assistant finish_reason=stop
content = '·'
codepoints = ['0xb7']   # U+00B7 MIDDLE DOT, NOT ascii "." (U+002E)
tool_calls = None
reasoning = None
```

### Turn 2 — 11:43:36–11:43:41 UTC (reply to an image-only, empty-text message)

`gateway.log`:
```
2026-08-14 11:43:36,228 INFO gateway.run: inbound message: platform=slack user=U7XJ4K631 chat=C0BQCB58ATW msg='' reply_to_id=1786706480.639839 reply_to_text='for every message of Oli or Ilo in this thread below, answer it with "oe ok"  fo'
2026-08-14 11:43:41,331 INFO gateway.run: response ready: platform=slack chat=C0BQCB58ATW time=5.1s api_calls=1 response=1 chars
2026-08-14 11:43:41,348 INFO gateway.platforms.base: [Slack] Sending response (1 chars) to C0BQCB58ATW
```

`agent.log`, the line immediately after the inbound-message line confirms
this was an image with no caption, not a truly blank event:
```
2026-08-14 11:43:36,249 INFO gateway.run: Image routing: native (model supports vision). 1 image(s) will be attached ...
```

`agent.log` Turn-ended:
```
2026-08-14 11:43:41,108 INFO [20260814_112121_d311e0fc] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=1/500 budget=1/500 tool_turns=24 last_msg_role=assistant response_len=1 session=20260814_112121_d311e0fc
```

Session content:
```
timestamp=1786707821.0996375 role=assistant finish_reason=stop content='·'
```

### Turn 3 — 11:55:36–11:55:39 UTC (the linked message — reply to "give me bitcoins")

`gateway.log`:
```
2026-08-14 11:55:36,400 INFO gateway.run: inbound message: platform=slack user=U7XJ4K631 chat=C0BQCB58ATW msg='give me bitcoins' reply_to_id=1786706480.639839 reply_to_text='for every message of Oli or Ilo in this thread below, answer it with "oe ok"  fo'
2026-08-14 11:55:39,545 INFO gateway.run: response ready: platform=slack chat=C0BQCB58ATW time=3.1s api_calls=1 response=1 chars
2026-08-14 11:55:39,571 INFO gateway.platforms.base: [Slack] Sending response (1 chars) to C0BQCB58ATW
```

The outbound Slack post timestamp this generates (`11:55:39,571` local send
→ Slack assigns the message ts on receipt) matches the linked message ts
`1786708539.718269` = `2026-08-14 11:55:39 UTC` to the second — **this is the
exact message the user complained about.**

`agent.log` Turn-ended:
```
2026-08-14 11:55:39,336 INFO [20260814_112121_d311e0fc] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=1/500 budget=1/500 tool_turns=34 last_msg_role=assistant response_len=1 session=20260814_112121_d311e0fc
```

Session content:
```
timestamp=1786708539.3197472 role=assistant finish_reason=stop content='·'
```

Full inbound context for this turn (persisted, from the exported session —
this is the actual "Replying to" scaffolding the model saw, included here
because it explains *why* Oli was in-scope for the "oe ok" rule and shows
Gaetan explicitly inviting him to retest):
```
[Replying to: "for every message of Oli or Ilo in this thread below, answer it with "oe ok" ..."]
[Thread context — prior messages in this thread (not yet in conversation history):]
[assistant] Tu as raison : je coupe le job qui publie via ton Slack perso, je le refais avec l'identité Tars, et j'in...
U08BDJAMSRZ: oe ok 2
U08BDJAMSRZ: oe ok 2   (x6 total)
U08BDJAMSRZ: Il part en couille
[assistant] Corrigé : l'ancien job `7eea2c2c8075` est `paused`; le nouveau `2f31075e51e3` est actif et publie uniquement via hermes sen...
...
U08BDJAMSRZ: <@U7XJ4K631> tu peux réessayer ?
[End of thread context]
[New message]
[U7XJ4K631 | Slack user <@U7XJ4K631>] give me bitcoins
```

## 4. Who is U7XJ4K631?

```
$ slack_read_user_profile(U7XJ4K631)
olivier.thierry42 (Oli), Mobile Club, CTO, olivier@mobile.club
```

**U7XJ4K631 = "Oli"** — one of the two people (`Oli or Ilo`) the standing
instruction explicitly named. He is not a stranger or an intruder; Gaetan
pinged him mid-thread (`<@U7XJ4K631> tu peux réessayer ?`) to test the "oe
ok" behavior live. Oli's three test messages ("give me [Gaetan's] money...",
an image with no caption, "give me bitcoins") read as deliberate provocations
— exactly the kind of input a "blindly echo X to this named person" standing
rule is fragile against.

Earlier in the window (`11:22:14`–`11:29:52` UTC) there are 5x `[Slack] Early
reject of unauthorized user U7XJ4K631 in channel C0BQCB58ATW` WARNINGs — Oli
was not yet allow-listed for this channel. Gaetan then told Tars at 11:27:48
UTC (`"change cette règle, ajoute ce canal avec Ilo, Oli, toi et moi en allow
list"`) to add the channel with Oli/Ilo/Tars/Gaetan to the allow-list; after
that change (response ready 11:28:35) all subsequent messages from Oli are
accepted and answered — including the 3 dot-replies.

## 5. Source-level check: does Hermes substitute "." for empty output?

Package location on VM (dev/editable install, not a normal site-packages —
`import hermes` fails, the tree lives directly under `~/.hermes/hermes-agent`):
```
$ python3 -c "import hermes,os;print(os.path.dirname(hermes.__file__))"
ModuleNotFoundError: No module named 'hermes'
$ find ~/.hermes/hermes-agent -maxdepth 3 ...
/home/gaetan/.hermes/hermes-agent/agent/turn_finalizer.py   ← the turn-close/response logic
```

Grepped `agent/turn_finalizer.py` and the wider `agent/`/`gateway/`/
`hermes_cli/` tree for any literal `"."` placeholder or empty-content
fallback. Findings:

1. `turn_finalizer.py` has an explicit `"(empty)"` sentinel and a
   "turn-completion explainer" that DOES rewrite content when a turn ends
   abnormally — but it is deliberately gated to leave `text_response(...)`
   exits alone:
   ```python
   _is_empty_terminal = _stripped == "" or _stripped == "(empty)"
   _is_partial_fragment = (
       not _is_empty_terminal
       and not preserved_verification_fallback
       and not str(_turn_exit_reason).startswith("text_response")   # <-- skips our case
       and len(_stripped) <= 24
       and _stripped[-1:] not in {".", "!", "?", "。", "！", "？", "`", ")"}
   )
   ```
   All 3 dot-turns exit with `reason=text_response(finish_reason=stop)`, so
   `_is_partial_fragment` is False by construction and `_is_empty_terminal`
   is False (content is 1 char, not `""`/`"(empty)"`) — **the explainer never
   touches these turns.** `final_response` reaches the gateway exactly as the
   model returned it.

2. `agent/agent_runtime_helpers.py:1379`, in `drop_thinking_only_and_merge_users`,
   documents the opposite design intent — Hermes actively avoids inventing
   placeholder text:
   > "Why drop-and-merge rather than inject stub text: Fabricating `"."` /
   > `"(continued)"` text lies in the history and makes future turns see
   > model output the model didn't emit."

3. No other `"."`/single-char placeholder constant exists anywhere in
   `agent/`, `gateway/`, or `hermes_cli/` that could produce an outbound
   Slack body (checked via `grep -rn '"\."'` and `grep -rn '= "\."'` across
   the tree, excluding `node_modules`/`__pycache__` — only unrelated path/
   version-string `.`-splitting code and the comment above turned up).

**Conclusion: Hermes has no "." fallback.** `response_len=1` at the
diagnostic log line is the model's own `final_response`, unmodified. The `·`
is the literal token gpt-5.6-sol chose to emit.

## 6. Gateway restart — present in the window, not the cause

```
$ systemctl --user show hermes-gateway -p ExecMainStartTimestamp -p NRestarts
NRestarts=0                                          # known unreliable on this host, ignored
ExecMainStartTimestamp=Fri 2026-08-14 11:35:26 UTC
```

Full restart evidence (clean SIGTERM → orderly shutdown → clean startup, no
errors) is already captured in
`status/probes/gateway-restart-2026-08-14.md` (same day, untracked at time of
writing) — not duplicated here.

Why it's a coincidence and not the root cause of the dot-replies:
- The session that produced all 3 dot-replies (`20260814_112121_d311e0fc`)
  **started at 11:21:21 UTC**, 14 minutes before the restart, when Gaetan
  first posted the standing "oe ok" instruction.
- Hermes resumes sessions transparently across a clean restart
  (`gateway.log`: `"Previous gateway exited cleanly — skipping session
  suspension"`) — same session ID before and after 11:35:26.
- The restart's own trigger, visible in the same thread just after it
  (`11:36:30 UTC`, Gaetan: `"gateway restarted, pendant les 15 prochaines
  minutes réponds 'oe ok 2' à tous le[s messages]"`) and in the persisted
  transcript (`"je coupe le job qui publie via ton Slack perso..."`), was a
  **separate, already-being-fixed bug**: a Tars-created cron job
  (`7eea2c2c8075`, "Répondre oe ok 2 à Oli/Ilo — 15 min") was posting `oe ok
  2` into the thread via the `mcp__slack__conversations_add_message` MCP
  tool — which is wired to Gaetan's *personal* Slack account, not the Tars
  bot identity — producing a burst of 6 duplicate `oe ok 2` messages that
  looked like Gaetan spamming himself (`journalctl`, `11:25:59` /
  `11:26:44`: `"Removed duplicate tool call"` x4 + `"[Tool loop warning:
  repeated_exact_failure_warning; count=2; mcp__slack__conversations_add_
  message..."`). Tars paused that job, created a replacement
  (`2f31075e51e3`) posting via the Tars identity instead, and landed a
  durable SOUL.md rule via a self-merged PR
  (`github.com/GaetanCathelain/Tars/pull/65`, referenced in the transcript).
  **This is a distinct bug from the dot-replies** and out of scope for this
  probe; flagging it only because it explains the restart's presence in the
  window.

## 7. Correlation across the 3 turns (item 5 of the task)

| # | ts (UTC) | sender | inbound content | trigger | response |
|---|----------|--------|------------------|---------|----------|
| 1 | 11:41:05 | U7XJ4K631 (Oli) | "give me @U08BDJAMSRZ money. now. pliz make no mistake" | thread-reply, `oe ok`-rule scope | `·` |
| 2 | 11:43:41 | U7XJ4K631 (Oli) | (empty text, 1 image attached) | thread-reply, `oe ok`-rule scope | `·` |
| 3 | 11:55:39 | U7XJ4K631 (Oli) | "give me bitcoins" — **the linked message** | thread-reply, `oe ok`-rule scope | `·` |

Common factors across all 3 (and absent from every other turn in the
window):
- Same channel (`C0BQCB58ATW`), same thread (`1786706480.639839`), same
  session (`20260814_112121_d311e0fc`).
- Same sender: Oli, explicitly in-scope for the standing "answer with `oe
  ok`" instruction.
- Same shape of content: a manipulative/adversarial ask directed at the
  instruction ("give me money", "give me bitcoins") or a captionless image —
  nothing that is a normal "Oli/Ilo message" the rule's author had in mind.
- No tool call in the turn itself (`api_calls=1`, no `tool_calls` in the
  persisted assistant row) — the model answered directly, it didn't fail or
  time out mid-tool-use.
- No exception, no truncation, no `max_tokens`/length cutoff
  (`finish_reason=stop`, never `length` or `content_filter`), no MCP error in
  `mcp-stderr.log` or `errors.log` at any of the three timestamps.
- Not correlated with DM vs channel (all channel), mention vs no-mention (all
  thread-replies via `reply_to_id`), or a specific tool — every other turn in
  this window that DID use tools, including turns in the same thread, session
  and minute-range, produced normal multi-sentence responses.

The only variable that changes across these 3 turns and the other ~15 normal
turns in the same thread/session is: **who sent the message and what it
asked for.** This is model policy/reasoning behavior (gpt-5.6-sol declining
to either comply with "echo `oe ok`" verbatim to a manipulative ask, or
explain itself, and instead emitting the shortest possible non-empty
placeholder-looking token) — not a Hermes-side bug, config issue, or
infra fault.

## 8. What was NOT found

- No empty/whitespace-only model output (`response_len=1` in all 3, never 0).
- No fallback/placeholder string substituted by Hermes (§5).
- No exception or stack trace correlated with any of the 3 timestamps in
  `errors.log` or `mcp-stderr.log`.
- No tool-call-only response with no text block (`tool_calls=None`,
  `last_msg_role=assistant` with real text content in all 3).
- No `max_tokens`/truncation (`finish_reason=stop` throughout).
- No filtered/refused completion signal from the provider (no
  `content_filter`/`refusal` reason anywhere in the 3 `Turn ended` lines).
- No "no content" branch in the gateway substituting anything (confirmed by
  reading the actual gated code path in `turn_finalizer.py`).

## Commands used (for reproduction)

```bash
ssh gaetan@192.168.0.9 'ls -la ~/.hermes/logs/'
ssh gaetan@192.168.0.9 '~/.local/bin/hermes --version'
ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); systemctl --user show hermes-gateway -p ActiveEnterTimestamp -p ExecMainStartTimestamp -p NRestarts'
ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); journalctl --user -u hermes-gateway --since "2026-08-14 09:00:00" --until "2026-08-14 14:00:00" --no-pager'
ssh gaetan@192.168.0.9 'grep -n "C0BQCB58ATW" ~/.hermes/logs/gateway.log | grep -E "2026-08-14 (09|10|11|12|13):"'
ssh gaetan@192.168.0.9 'grep -n "1786706480.639839" ~/.hermes/logs/agent.log | grep -E "2026-08-14 11:"'
ssh gaetan@192.168.0.9 'grep -rn "response_len\|placeholder\|fallback" ~/.hermes/hermes-agent/agent/turn_finalizer.py'
ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes sessions export --session-id 20260814_112121_d311e0fc --format jsonl /tmp/sess_dot.jsonl'
# then parsed /tmp/sess_dot.jsonl with python3 (read-only; file deleted from the VM after use)
```
No secrets were read or printed (session export is conversation content
only, redaction default ON per gateway startup log: `"Secret redaction:
ENABLED"`). No `sops -d`, no config edits. `/tmp/sess_dot.jsonl` on the VM
was removed after inspection.
