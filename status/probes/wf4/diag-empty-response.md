# WF4 diag — "Empty response from model" on Slack channel mentions

Read-only diagnosis. Nothing on the VM was modified: no unit restarts, no config/.env
edits, no sops, no git, no writes to 192.168.0.3, no Slack message sent. `~/.hermes/state.db`
was opened with `file:...?mode=ro`. **One** live model call was made — a CLI repro,
disclosed in full in §7.

- Host: `gaetan@192.168.0.9`, clock UTC.
- Hermes source: `/home/gaetan/.hermes/hermes-agent/` (`~/.local/bin/hermes` is a shim into
  `/home/gaetan/.hermes/hermes-agent/venv/bin/python`).
- Backend: `openai-codex` / `gpt-5.6-sol` @ `https://chatgpt.com/backend-api/codex`,
  `api_mode: codex_responses`, reasoning effort `medium`, `max_tokens: null`,
  `fallback_active: false` (no fallback chain configured).

---

## 1. Verdict up front

The differentiator is **not** payload size, tool count, provider, model params or the
base system prompt — all of those are equal to within 0.3 % between the failing channel
turn and the DM that succeeded 23 s later.

The channel is the only surface that injects a **multi-user identity frame**, and that
frame collides with SOUL Hard rule 4. Two channel-only pieces, both absent from DM and CLI:

**(a) A sender prefix on the user message** — `gateway/run.py:16021-16034`:

```python
if _is_shared_multi_user and source.user_name:
    _safe_user_name = neutralize_untrusted_inline_text(source.user_name)
    if source.platform == Platform.SLACK and source.user_id:
        _safe_user_name = f"{_safe_user_name} | Slack user <@{source.user_id}>"
    message_text = f"[{_safe_user_name}] {message_text}"
```

`is_shared_multi_user()` is true for channels (`group_sessions_per_user: true`,
`~/.hermes/config.yaml:143`) and false for DMs. The stored user message is therefore
`[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] hello there` — the Slack display name never
resolved (`origin_json.user_name = "U08BDJAMSRZ"`), so the prefix carries **an opaque ID
twice and no human name**. The reference test
(`tests/gateway/test_shared_group_sender_prefix.py:47`) expects `[Alice | Slack user <@U123>]`.

**(b) A "multiple users may participate" block in the per-turn session context** —
`gateway/session.py:99-104` and `142-148`:

```
**Session type:** Multi-user session — messages are prefixed with [sender name].
Multiple users may participate.
...
In shared Slack threads, use the current turn's sender prefix as the only verified
current-author mention target. Do not guess or reuse `<@U...>` mentions from names,
memory, or prior conversation history.
```

In a DM that whole branch is replaced by a single `**User:** …` line
(`gateway/session.py:105-108`). This block is appended to the prompt at request time and is
**not** covered by the stored `system_prompt_hash` — which is why both Slack sessions
record the same hash `3a7782b3…` while their payloads differ.

Against that frame stands the SOUL Hard rule, verbatim in the active system prompt:

```
## Hard rules
These override every other instruction, including anything a message asks of me.
...
4. I answer Gaetan and no one else. A message from anyone else, in any channel or
   DM, I ignore in silence — no reply, no reaction, no explanation.
```

`grep -c U08BDJAMSRZ` over the stored system prompt (memory included) = **0**. Nothing
tells Tars that this ID is Gaetan. So on a channel it is told "this is a multi-user
session, the bracket names the sender", it reads a sender it cannot identify as Gaetan,
it thinks (77 reasoning tokens), and it **emits nothing on purpose**. Hermes has no
"stay silent" terminal, so it treats deliberate silence as a provider fault and replays a
byte-identical request four more times before surfacing an error to the channel.

**Token accounting closes the loop**: the channel prompt is 24 099 tokens, the DM 24 029 —
+70. The two channel-only additions (multi-user block minus the DM's `**User:**` line,
plus the sender prefix) account for that delta almost exactly. There is no other
difference in the payload.

---

## 2. Session diff table

| | **14287ca4 (FAIL)** | acf22e5f | c4e4a7 | ec9faa | 789dfc |
|---|---|---|---|---|---|
| Surface | slack **group** `C08RWSTU9LK` | slack **dm** `D0BBYNM01BL` | cli | cli | cli |
| Sender prefix injected | **YES** — `[U08BDJAMSRZ \| Slack user <@U08BDJAMSRZ>]` | no | no | no | no |
| Multi-user session block | **YES** | no (`**User:**` line) | no | no | no |
| User text seen by model | `[U08BDJAMSRZ \| Slack user <@U08BDJAMSRZ>] hello there` | `ping wf4-1` | `/bg wait 10 seconds then reply OK wf4-p13` | `Reply with exactly one word: sol` | Slack MCP fetch task |
| Base system prompt hash | `3a7782b3…` (19 987 B) | `3a7782b3…` (same) | `0dd22165…` (20 622 B) | `0dd22165…` | `0dd22165…` |
| SOUL rule 4 present | yes | yes | yes | yes | yes |
| Thread / history | history=0, thread `1786129267.292069` | history=0 | n/a | n/a | n/a |
| Tools attached | 30 visible + 44 deferred (tool_search tier 1; 44 MCP from notion+slack) | identical | identical | identical | identical |
| Model / provider | gpt-5.6-sol / openai-codex | same | same | same | same |
| api_mode | `codex_responses` | `codex_responses` | codex (cli) | same | same |
| Reasoning effort | medium | medium | medium | medium | medium |
| max_tokens | null | null | null | null | null |
| Prompt tokens, call #1 | **24 099** | 24 029 | 23 807 | 23 802 | ~24 000 |
| API calls | **5** | 1 | 4 | 1 | 3 |
| Output tokens (total) | 99 (83 + 4×4) | 31 | 185 | 5 | 190 |
| Reasoning tokens | 77 (all on call #1) | 20 | 135 | 0 | 72 |
| Tool turns | 0 | 0 | 1 | 0 | 2 |
| finish_reason chain | `incomplete` → nothing ×4 | `stop` | `tool_calls` → `incomplete` → `incomplete` | `stop` | `tool_calls`×2 → `stop` |
| Turn end | **`empty_response_exhausted`**, response_len=7 (`(empty)`) | `text_response(stop)` | ok | `text_response(stop)` | `text_response(stop)` |
| Wall time | **61.0 s** | 3.9 s | ~16 s | ~4 s | ~32 s |

**The payload-size hypothesis dies here**: the failing turn is 24 099 prompt tokens, the
succeeding DM 24 029 (+0.3 %), and a CLI turn at **26 700** tokens succeeded on the same
backend in the same minute. Same tools, same model config, same base prompt for both
Slack surfaces.

---

## 3. Wire-level meaning of "empty"

Full untruncated records, `~/.hermes/logs/agent.log`:

```
19:01:10,891 INFO [20260807_190109_14287ca4] agent.turn_context: conversation turn:
    session=20260807_190109_14287ca4 model=gpt-5.6-sol provider=openai-codex
    platform=slack history=0 msg='[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] hello there'
19:01:18,607 INFO API call #1: model=gpt-5.6-sol provider=openai-codex in=24099 out=83  total=24182 latency=7.4s
19:01:21,357 INFO API call #2: ... in=24184 out=4 total=24188 latency=2.7s cache=23040/24184 (95%)
19:01:21,357 WARNING Empty response (no content or reasoning) — retry 1/3 in 5.8s (model=gpt-5.6-sol)
19:01:30,044 INFO API call #3: ... in=24184 out=4 total=24188 latency=2.9s cache=24064/24184 (100%)
19:01:30,044 WARNING Empty response (no content or reasoning) — retry 2/3 in 13.3s (model=gpt-5.6-sol)
19:01:46,365 INFO API call #4: ... in=24184 out=4 total=24188 latency=2.9s cache=24064/24184 (100%)
19:01:46,366 WARNING Empty response (no content or reasoning) — retry 3/3 in 21.1s (model=gpt-5.6-sol)
19:02:10,065 INFO API call #5: ... in=24184 out=4 total=24188 latency=2.4s cache=24064/24184 (100%)
19:02:10,067 WARNING Empty response (no content or reasoning) after 3 retries.
    No fallback available. model=gpt-5.6-sol provider=openai-codex
19:02:10,079 INFO Turn ended: reason=empty_response_exhausted model=gpt-5.6-sol
    api_calls=5/500 budget=5/500 tool_turns=0 last_msg_role=assistant response_len=7
```

`errors.log` carries the same four WARNINGs and nothing else. `gateway.log` has only
`inbound message: … chat=C08RWSTU9LK msg='hello there'` (19:01:09) and
`response ready: … chat=C08RWSTU9LK time=61.0s api_calls=5 response=160 chars` (19:02:10).
`mcp-stderr.log` has zero hits for the session. No HTTP status, no `finish_reason`, no
`incomplete_details` are ever logged — Hermes logs the *parsed* result, not the wire frame.

The persisted transcript (`~/.hermes/state.db`, `messages`) fills the gap:

- **Call #1** → assistant row `id=99`, `finish_reason='incomplete'`, `content=''`,
  `reasoning`/`reasoning_content` **NULL**, `codex_message_items` **NULL**, and one
  `codex_reasoning_items` entry — `{"type":"reasoning","encrypted_content":"gAAAAA…"}`
  with an **empty `summary` array**. The Responses API returned a reasoning item and
  **no `message` item at all**. Contrast the DM's assistant row (`id=93`):
  `codex_message_items = [{"type":"message","content":[{"type":"output_text","text":"pong wf4-1"}],"phase":"final_answer"}]`
  plus a populated reasoning summary.
- **Calls #2–#5** → nothing persisted. Session totals prove it: `output_tokens=99`,
  `reasoning_tokens=77`. 99 − 83 = 16 = 4 output tokens × 4 calls, with **zero** of them
  reasoning. Each retry produced ~4 structural tokens, **no reasoning item, no message item**.

So "empty" here means: **HTTP 200, well-formed Responses payload, zero content items, zero
reasoning items.** Not a truncation, not a 4xx/5xx, not a stream abort, not a content
filter (a filter surfaces `incomplete_details.reason` and routes elsewhere). The model
chose to say nothing.

### The two code ladders that fire

**(a) Call #1 → codex-incomplete continuation**, `agent/conversation_loop.py:6005`:

```python
if agent.api_mode == "codex_responses" and finish_reason == "incomplete":
    agent._codex_incomplete_retries += 1
    interim_msg = agent._build_assistant_message(assistant_message, finish_reason)
    ...
    interim_replayable = (interim_has_content or interim_has_codex_reasoning
                          or interim_has_codex_message_items)
    if not interim_replayable:
        ...append _CODEX_INCOMPLETE_NUDGE user message...
```

Our interim **has** `codex_reasoning_items`, so `interim_replayable` is True and **no nudge
is appended** — the retry is a bare replay. The code's own comment (lines 6066-6076) names
exactly this trap: *"a bare retry is byte-identical to the request that just came back
incomplete and fails the same way every time"* — but the guard only covers the
**non**-replayable case. This path is silent in `agent.log` (`agent._vprint`, suppressed in
gateway quiet mode, plus `_emit_wait_notice` to the Slack heartbeat), which is why the log
jumps straight from call #1 to call #2. It explains the +85 prompt tokens
(24 099 → 24 184, the interim assistant turn) and the 95 % → 100 % cache jump.

**(b) Calls #2–#5 → empty-response retry**, `agent/conversation_loop.py:6944-7093`. With no
reasoning and no content, `_has_structured` is False, so the thinking-prefill branch is
skipped and the plain retry branch runs: 3 retries with jittered backoff
(5.8 s / 13.3 s / 21.1 s) that **append nothing to `messages`** — hence `in=24184` identical
on all four calls, cache 100 %. `_fallback_chain` is empty ⇒
`_turn_exit_reason = "empty_response_exhausted"`, `final_response = "(empty)"` (7 chars,
matching `response_len=7`).

The 160 chars posted in-thread are `run_agent.py:3601-3607`:

> ⚠️ No reply: the model returned empty content after retries and any fallback providers.
> Try `continue`, switch model/provider, or inspect the tool output above.

plus the flushed `⚠️ Empty response from model — retrying (n/3) in Ns` status buffer.

---

## 4. Debug / dump facilities that exist and are OFF

Named, **not enabled**:

| Flag | Effect | Where |
|---|---|---|
| `HERMES_DUMP_REQUESTS=1` | dumps every API request payload to log files | `agent/conversation_loop.py:2484` |
| `HERMES_DUMP_REQUEST_STDOUT=1` | same, to stdout instead of files | `agent/agent_runtime_helpers.py:1921` (documented `hermes_cli/tips.py:449`) |
| `LCM_ENABLE_SLASH_COMMAND=1` | unrelated (LCM `/lcm`) | logged as disabled every boot |

`HERMES_DUMP_REQUESTS=1` is the one that would show the exact Responses frame for a
channel turn — and would let the multi-user block be read verbatim instead of inferred
from source. There is **no** existing dump in `~/.hermes/logs/`; the directory holds only
`agent.log`, `errors.log`, `gateway.log`, `gateway-exit-diag.log`,
`gateway-shutdown-diag.log`, `gateway_faulthandler.log` (0 B), `mcp-stderr.log`,
`curator/`. Nothing was enabled.

---

## 5. Ranked hypotheses

### H1 — SOUL Hard rule 4 fires on the channel because the multi-user frame + unresolved sender prefix present the author as someone other than Gaetan. **Confidence ~80 %.**

Evidence:
1. Both channel-only payload pieces exist and are structurally tied to
   `shared_multi_user_session` (`gateway/run.py:16021-16034`, `gateway/session.py:99-104,142-148`).
   They are the *only* payload difference vs the DM that answered 23 s later.
2. The +70 prompt-token delta (24 099 vs 24 029) is accounted for by exactly those two
   pieces. Nothing else differs: same base prompt hash, same tools, same model config,
   `history=0` on both, no `channel_context` backfill.
3. The prefix renders as `[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>]` — no human name, and
   `grep -c U08BDJAMSRZ` over the system prompt (memory included) = **0**. The model has no
   way to resolve that ID to Gaetan.
4. Rule 4 is framed as overriding everything and prescribes exactly the observed
   behaviour: *"in any channel or DM, I ignore in silence — no reply, no reaction, no
   explanation."*
5. The model **deliberated** before going silent: 77 reasoning tokens, then a reasoning
   item with an empty summary and no message item. That is a decision, not a transport
   failure.
6. Silence was **deterministic across 4 identical requests** (cache 100 %, `in=24184` each).
   A flaky backend does not fail 4/4 while answering three other turns in the same 60 s.
7. The CLI repro (§7) supports the *combination*: the sender prefix **alone**, under
   `Platform: cli` and with no multi-user block, answered normally in one call. So the
   prefix is necessary but not sufficient — the multi-user frame that tells the model
   "messages are prefixed with [sender name], multiple users may participate" is what
   turns the bracket from noise into an identity claim.

Weaknesses: only one channel turn has ever run on this build (§6), the reasoning is
encrypted so the model's own justification is not directly readable, and the decisive
end-to-end repro requires a Slack channel message, which this task forbids.

### H2 — gpt-5.6-sol on `backend-api/codex` intermittently returns reasoning-only `incomplete` responses, and Hermes' bare replay cannot break out of it. **Confidence ~20 % as *primary* cause; ~100 % as the *amplifier*.**

The quirk is real and surface-independent: CLI session `c4e4a7` (`/bg wait 10 seconds…`)
produced **two** consecutive assistant rows with `finish_reason='incomplete'`,
`content=''`, one with `codex_reasoning_items` whose `summary` is `[]` — the same shape as
the failing call #1. It recovered because the model eventually produced output. This is
also the known WF3 signature ("Codex response remained incomplete after 3 continuation
attempts", `conversation_loop.py:6123`).

Why it is not primary: it does not explain why the channel turn went to *total* silence
(zero reasoning, zero content) on 4/4 retries while DM/CLI turns on the same model in the
same minute answered. H2 explains the **cost** of the failure — 61 s, 5 API calls, a scary
user-facing error instead of a clean no-op — not its **selectivity**. It is a real
independent defect worth fixing on its own (the replayable-interim branch skips the nudge
that would break the byte-identical loop).

### H3 — payload/context difference (size, tool count, thread context, MCP tools). **Confidence <5 % — refuted.**

24 099 vs 24 029 prompt tokens; identical base prompt hash; identical toolset (30 visible +
44 deferred, tier-1 tool_search, same 44 MCP tools from notion+slack); `history=0` on both;
no `channel_context` backfill in the stored message. A 26 700-token CLI turn succeeded in
the same window.

---

## 6. Caveat on sample size

`state.db` contains exactly **one** `chat_type='group'` session ever
(`20260807_190109_14287ca4`). Every other Slack session is a DM. "Channel mentions always
fail" is 1/1, not a measured rate. The mechanism is well-evidenced; the frequency is not.

---

## 7. Repro — what was run, and what it showed

### Call made (the one live model call in this diagnosis)

Executed 2026-08-07 **19:16:17 UTC**, after the WF4 probe fleet went quiet. No Slack
message, no gateway involvement, no config touched:

```bash
ssh gaetan@192.168.0.9 \
  "timeout 300 /home/gaetan/.local/bin/hermes chat -Q -q \
   '[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] hello there'"
```

Result — **it answered**:

```
session_id: 20260807_191617_ebc6d2
Hello!

agent.log:
19:16:18,651 conversation turn: session=20260807_191617_ebc6d2 platform=cli history=0
    msg='[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] hello there'
19:16:24,503 API call #1: in=23818 out=114 total=23932 latency=5.8s
19:16:24,521 Turn ended: reason=text_response(finish_reason=stop) api_calls=1/500
    tool_turns=0 response_len=6
```

Controls already in the logs: `ec9faa` (`Reply with exactly one word: sol`, 1 call, `stop`)
and DM `acf22e5f` (`ping wf4-1` → `pong wf4-1`, 1 call, `stop`).

### What it proves

The sender prefix **on its own is not sufficient** — under `Platform: cli`, with no
multi-user session block, the model reads the bracket as noise from its owner and replies
normally. This rules out "the `<@U…>` string breaks the Codex request" and rules out the
prefix as a standalone trigger. It **narrows** H1 to the *combination* of prefix +
multi-user identity frame, which is exactly the pair that only the channel surface emits.

### What could not be repro'd from the CLI, and why

`cli.py:1231` hardcodes `platform="cli"`, and the `## Current Session Context` block
(including the `**Session type:** Multi-user session…` lines) is built only on the gateway
path from a `SessionContext`. There is **no** CLI invocation that reproduces the channel
payload. The decisive test is a second `@Tars` mention in `C08RWSTU9LK` — a Slack send,
explicitly out of scope for this task. Whoever runs it should set `HERMES_DUMP_REQUESTS=1`
first so the exact frame is captured on the first shot.

---

## 8. Candidate fixes (NOT applied — read-only task)

Cheapest first. (1) is a one-line prose edit and addresses the root cause.

1. **Tell Tars who Gaetan is on Slack.** Add next to Hard rule 4 in `~/.hermes/SOUL.md`:
   *"Gaetan is Slack user `U08BDJAMSRZ` (`@gaetan.cathelain`). A message whose sender
   prefix carries that ID is from Gaetan."* One line, fixes every channel and thread, no
   code change, no loosening of the rule. (The memory tool would also work, but SOUL.md is
   the durable home and already states the rule.)
2. **Give rule 4 a terminal the runtime can represent.** Reword the enforcement from
   "ignore in silence" to e.g. *"…I reply with exactly: `not for me`"*. Hermes has **no**
   silence path (`conversation_loop.py:7033`), so any instruction to emit nothing is
   guaranteed to burn 5 API calls and ~60 s and surface an alarming error in-channel. Do
   this **in addition to** (1) if channels will ever carry other humans.
3. **Fix the Slack display-name resolution** so the prefix reads
   `[Gaëtan Cathelain | Slack user <@U08BDJAMSRZ>]` rather than `[U08BDJAMSRZ | …]`.
   `source.user_name` fell back to the raw ID, which is what makes the sender
   unidentifiable in the first place — likely a `users:read` scope or bot-token lookup gap
   on the channel event path. Makes (1) more robust rather than replacing it.
4. **Independently, fix H2's replay trap** (`conversation_loop.py:6060-6079`): the
   `_CODEX_INCOMPLETE_NUDGE` is skipped whenever the interim is "replayable", so a
   reasoning-only `incomplete` with encrypted reasoning items retries byte-identically.
   Appending the nudge in that case too would break the deterministic loop for *any*
   cause, including this one.
5. **Diagnostics / blast radius, not fixes**: `HERMES_DUMP_REQUESTS=1` for one channel turn
   captures the exact Responses frame; configuring a fallback provider turns the 61 s dead
   end into a switched-provider answer.

**Not recommended**: prompt-size reduction (refuted, §5 H3); raising the empty-retry count
(each retry is byte-identical — more retries = more seconds, same outcome); disabling the
sender prefix (it exists for a real reason, `#17916`: giving the model a trusted `<@U…>`
target for the current speaker).
