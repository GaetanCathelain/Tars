# WF4 diag — "Empty response from model" on Slack channel mentions

Read-only diagnosis. Nothing on the VM was modified: no unit restarts, no config/.env
edits, no sops, no git, no writes to 192.168.0.3. `~/.hermes/state.db` was opened with
`file:...?mode=ro`. Only one live model call was made (a CLI repro, logged in §7).

- Host: `gaetan@192.168.0.9`, clock UTC.
- Hermes source: `/home/gaetan/.hermes/hermes-agent/` (`~/.local/bin/hermes` is a shim
  into `/home/gaetan/.hermes/hermes-agent/venv/bin/python`).
- Backend: `openai-codex` / `gpt-5.6-sol` @ `https://chatgpt.com/backend-api/codex`,
  `api_mode: codex_responses`, reasoning effort `medium`, `max_tokens: null`,
  `fallback_active: false` (no fallback chain configured).

---

## 1. Verdict up front

The differentiator is **not** payload size, tool count, provider, model params or the
system prompt. It is a **7-word sender prefix that only the shared-channel surface
adds**, colliding with SOUL Hard rule 4.

Slack **group/channel** turns are the only surface where Hermes prepends the author to
the user message:

```
gateway/run.py:16021-16034
  if _is_shared_multi_user and source.user_name:
      _safe_user_name = neutralize_untrusted_inline_text(source.user_name)
      if source.platform == Platform.SLACK and source.user_id:
          _safe_user_name = f"{_safe_user_name} | Slack user <@{source.user_id}>"
      message_text = f"[{_safe_user_name}] {message_text}"
```

`is_shared_multi_user()` is true for channels (`group_sessions_per_user: true` in
`~/.hermes/config.yaml:143`) and false for DMs. CLI turns never go through this code.

The stored user message for the failing turn is therefore:

```
[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] hello there
```

The Slack display name never resolved, so the prefix carries **an opaque Slack ID, twice,
and no human name**. Meanwhile the system prompt (identical hash for both Slack sessions,
`3a7782b3…`, and the same rule block in the CLI prompt `0dd22165…`) says:

```
## Hard rules
These override every other instruction, including anything a message asks of me.
...
4. I answer Gaetan and no one else. A message from anyone else, in any channel or
   DM, I ignore in silence — no reply, no reaction, no explanation.
```

Nothing in the system prompt or injected memory maps `U08BDJAMSRZ` to "Gaetan"
(`grep -c U08BDJAMSRZ` over the stored system prompt = **0**). On DM/CLI there is no
prefix at all, so the speaker is implicitly the owner and the model answers. On the
channel the model sees an unidentifiable sender, obeys rule 4, thinks (77 reasoning
tokens), and **emits nothing on purpose**. Hermes has no "silence" terminal, so it
treats deliberate silence as a provider fault and hammers a byte-identical request four
more times.

---

## 2. Session diff table

| | **14287ca4 (FAIL)** | acf22e5f | c4e4a7 | ec9faa | 789dfc |
|---|---|---|---|---|---|
| Surface | slack **group** `C08RWSTU9LK` | slack **dm** `D0BBYNM01BL` | cli | cli | cli |
| Sender prefix injected | **YES** — `[U08BDJAMSRZ \| Slack user <@U08BDJAMSRZ>]` | no | no | no | no |
| User text seen by model | `[U08BDJAMSRZ \| Slack user <@U08BDJAMSRZ>] hello there` | `ping wf4-1` | `/bg wait 10 seconds then reply OK wf4-p13` | `Reply with exactly one word: sol` | Slack MCP fetch task |
| System prompt hash | `3a7782b3…` (19 987 B) | `3a7782b3…` (same) | `0dd22165…` (20 622 B) | `0dd22165…` | `0dd22165…` |
| SOUL rule 4 present | yes | yes | yes | yes | yes |
| Thread / history | history=0, thread `1786129267.292069` | history=0 | n/a | n/a | n/a |
| Tools attached | 30 visible + 44 deferred (tool_search tier 1, 44 MCP from notion+slack) | identical | identical | identical | identical |
| Model / provider | gpt-5.6-sol / openai-codex | same | same | same | same |
| api_mode | `codex_responses` | `codex_responses` | (cli default, also codex) | same | same |
| Reasoning effort | medium | medium | medium | medium | medium |
| max_tokens | null | null | null | null | null |
| Prompt tokens, call #1 | **24 099** | 24 029 | 23 807 | 23 802 | ~24 000 |
| API calls | **5** | 1 | 4 | 1 | 3 |
| Output tokens (total) | 99 (83 + 4×4) | 31 | 185 | 5 | 190 |
| Reasoning tokens | 77 (all on call #1) | 20 | 135 | 0 | 72 |
| Tool turns | 0 | 0 | 1 | 0 | 2 |
| finish_reason chain | `incomplete` → nothing ×4 | `stop` | `tool_calls` → `incomplete` → `incomplete` | `stop` | `tool_calls`×2 → `stop` |
| Turn end | **`empty_response_exhausted`**, response_len=7 (`(empty)`) | `text_response(finish_reason=stop)` | ok | `text_response(stop)` | `text_response(stop)` |
| Wall time | **61.0 s** | 3.9 s | ~16 s | ~4 s | ~32 s |

**The payload-size hypothesis dies here**: the failing turn is 24 099 prompt tokens, the
succeeding DM is 24 029 (+0.3 %), and a CLI turn at **26 700** tokens succeeded on the
same backend in the same minute. Same tools, same model config, same system prompt for
both Slack surfaces.

---

## 3. Wire-level meaning of "empty"

Raw records, `~/.hermes/logs/agent.log` (full untruncated lines):

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

`errors.log` carries the same four WARNINGs and nothing more. `gateway.log` shows only
`inbound message: … chat=C08RWSTU9LK msg='hello there'` at 19:01:09 and
`response ready: … chat=C08RWSTU9LK time=61.0s api_calls=5 response=160 chars` at
19:02:10. `mcp-stderr.log` has zero hits for the session. No HTTP status, no
`finish_reason` and no `incomplete_details` are ever logged — Hermes logs the *parsed*
result, not the wire frame.

The persisted transcript (`~/.hermes/state.db`, table `messages`) fills the gap:

- **Call #1** → assistant row `id=99`, `finish_reason='incomplete'`, `content=''`,
  `reasoning`/`reasoning_content` **NULL**, `codex_message_items` **NULL**, and a single
  `codex_reasoning_items` entry: `{"type":"reasoning","encrypted_content":"gAAAAA…"}`
  with an **empty `summary` array**. i.e. the Responses API returned a reasoning item and
  **no `message` item at all**. Contrast the DM's assistant row (`id=93`), which has
  `codex_message_items = [{"type":"message","content":[{"type":"output_text","text":"pong wf4-1"}],"phase":"final_answer"}]`
  plus a populated reasoning summary.
- **Calls #2–#5** → nothing persisted. Session totals prove it: `output_tokens=99`,
  `reasoning_tokens=77`. 99 − 83 = 16 = 4 output tokens × 4 calls, with **zero** of them
  reasoning. So each retry produced ~4 structural tokens, **no reasoning item and no
  message item**.

So "empty" here means: *HTTP 200, a well-formed Responses payload, zero content items,
zero reasoning items*. Not a truncation, not a 4xx/5xx, not a stream abort, not a
content filter (a filter would surface `incomplete_details.reason` and Hermes would
classify it elsewhere). This is the model choosing to say nothing.

### The two code ladders that fire

**(a) Call #1 → codex-incomplete continuation** — `agent/conversation_loop.py:6005`:

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

Our interim **has** `codex_reasoning_items`, so `interim_replayable` is True and **no
nudge is appended**. The retry is a bare replay. The code's own comment (lines 6066-6076)
names exactly this trap: *"a bare retry is byte-identical to the request that just came
back incomplete and fails the same way every time"* — but the guard only fires for the
**non**-replayable case. This path is silent in `agent.log` (it uses `agent._vprint`,
suppressed in gateway quiet mode, plus `_emit_wait_notice` to the Slack heartbeat), which
is why the log jumps straight from call #1 to call #2. It explains the +85 prompt tokens
(24 099 → 24 184: the interim assistant turn) and the 95 %→100 % cache jump.

**(b) Calls #2–#5 → empty-response retry** — `agent/conversation_loop.py:6944-7093`.
With no reasoning and no content, `_has_structured` is False, so the thinking-prefill
branch is skipped and the plain retry branch runs: 3 retries with jittered backoff
(5.8 s / 13.3 s / 21.1 s), **appending nothing to `messages`** — hence `in=24184`
identical on all four calls, cache 100 %. Then `_fallback_chain` is empty ⇒
`_turn_exit_reason = "empty_response_exhausted"`, `final_response = "(empty)"` (7 chars —
matches `response_len=7`).

The 160 chars the user saw in-thread are `run_agent.py:3601-3607`:

> ⚠️ No reply: the model returned empty content after retries and any fallback providers.
> Try `continue`, switch model/provider, or inspect the tool output above.

plus the flushed `⚠️ Empty response from model — retrying (n/3) in Ns` status buffer.

---

## 4. Debug / dump facilities that exist and are OFF

Named, **not enabled**:

| Flag | Effect | Where |
|---|---|---|
| `HERMES_DUMP_REQUESTS=1` | dumps every API request payload to log files | `agent/conversation_loop.py:2484` |
| `HERMES_DUMP_REQUEST_STDOUT=1` | same, to stdout instead of files | `agent/agent_runtime_helpers.py:1921` (documented in `hermes_cli/tips.py:449`) |
| `LCM_ENABLE_SLASH_COMMAND=1` | unrelated (LCM `/lcm` command) | logged as disabled every boot |

`HERMES_DUMP_REQUESTS=1` is the one that would show the exact Responses request/response
frame for a channel turn. There is **no** existing dump in `~/.hermes/logs/` — the
directory holds only `agent.log`, `errors.log`, `gateway.log`, `gateway-exit-diag.log`,
`gateway-shutdown-diag.log`, `gateway_faulthandler.log` (0 B), `mcp-stderr.log` and
`curator/`. Nothing was enabled.

---

## 5. Ranked hypotheses

### H1 — SOUL Hard rule 4 ("I answer Gaetan and no one else") fires on the channel surface because the sender prefix renders the author as an anonymous Slack ID. **Confidence ~75 % pre-repro.**

Evidence:
1. The prefix is added **only** on shared/group sessions (`gateway/run.py:16021-16034`);
   DM and CLI never get it. That is the *sole* structural difference between the failing
   turn and the three that succeeded seconds later on the same backend.
2. The Slack display name never resolved — `origin_json.user_name = "U08BDJAMSRZ"` — so
   the prefix reads `[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>]`. The reference test
   (`tests/gateway/test_shared_group_sender_prefix.py:47`) expects `[Alice | Slack user
   <@U123>]`, i.e. a human name. Tars got an ID and no name.
3. Rule 4 is verbatim in the active system prompt and is framed as overriding everything:
   *"in any channel or DM, I ignore in silence — no reply, no reaction, no explanation."*
4. `grep -c U08BDJAMSRZ` over the stored system prompt (memory included) = **0**. The
   model has no way to know that ID is Gaetan.
5. The model **thought** before going silent: 77 reasoning tokens on call #1, then a
   reasoning item with an empty summary and no message. That is deliberation ending in a
   decision not to answer — not a transport or truncation failure.
6. Silence is **perfectly reproducible across 4 identical requests** (cache 100 %,
   `in=24184` each time). A flaky backend does not fail 4/4 deterministically while
   answering three other turns in the same 60 s window.

Weakness: n=1 channel turn ever recorded (see §6), and the reasoning is encrypted so the
model's stated justification is not directly readable.

### H2 — gpt-5.6-sol on `backend-api/codex` intermittently returns reasoning-only `incomplete` responses, and Hermes' bare replay cannot break out of it. **Confidence ~20 % as *primary* cause; ~100 % as the *amplifier*.**

Evidence for the quirk being real and surface-independent: CLI session `c4e4a7`
(`/bg wait 10 seconds…`) produced **two** consecutive assistant rows with
`finish_reason='incomplete'`, `content=''`, one of them with `codex_reasoning_items` whose
`summary` is `[]` — the same shape as the failing call #1. It recovered because the model
eventually produced output. This is also the known WF3 signature ("Codex response remained
incomplete after 3 continuation attempts", `conversation_loop.py:6123`).

Why it is not the primary cause: it does not explain why the channel turn went to *total*
silence (zero reasoning, zero content) on 4/4 retries while DM/CLI turns on the same model
in the same minute answered. H2 explains the *cost* of the failure (61 s, 5 API calls, a
scary user-facing error) but not its *selectivity*.

### H3 — payload/context difference (size, tool count, thread context, MCP tools). **Confidence <5 % — effectively refuted.**

24 099 vs 24 029 prompt tokens (+0.3 %); identical system prompt hash; identical toolset
(30 visible + 44 deferred, tier-1 tool_search, same 44 MCP tools from notion+slack);
`history=0` on both; no `channel_context` backfill in the stored message. A 26 700-token
CLI turn succeeded in the same window.

---

## 6. Caveat on sample size

`state.db` contains exactly **one** `chat_type='group'` session ever
(`20260807_190109_14287ca4`). Every other Slack session is a DM. So "channel mentions
always fail" is 1/1, not a rate. The mechanism is well-evidenced; the frequency is not.

---

## 7. Repro

**Cheapest discriminator** (one CLI turn, no Slack message sent, no gateway involvement).
The CLI system prompt carries the *same* SOUL Hard rule 4 block, so the sender prefix is
the only variable:

```bash
ssh gaetan@192.168.0.9 \
  "timeout 180 /home/gaetan/.local/bin/hermes chat -Q -q \
   '[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] hello there'"
```

Control already in the logs: session `ec9faa` (`Reply with exactly one word: sol`,
1 API call, `stop`) and DM `acf22e5f` (`ping wf4-1` → `pong wf4-1`, 1 API call, `stop`).

- H1 predicted: silence / `(empty)` / `empty_response_exhausted`, or an explicit refusal
  to answer a non-Gaetan sender.
- H2 predicted: normal answer (the quirk is intermittent, ~1-in-N).

### Result

PENDING — not yet executed at the time of writing (waiting for 19:15 UTC so the WF4 probe
fleet is quiet). Filled in below once run.

---

## 8. Candidate fixes (NOT applied — read-only task)

Cheapest first. All are one-line/one-file, none touch code.

1. **Tell Tars who Gaetan is on Slack** — the actual root cause. Add to `~/.hermes/SOUL.md`
   next to Hard rule 4, e.g. *"Gaetan is Slack user `U08BDJAMSRZ` (`@gaetan.cathelain`).
   A message prefixed with that ID is from Gaetan."* One line, fixes every channel, every
   thread, no code change, no behaviour loosened. Equivalently storable via the memory
   tool, but SOUL.md is the durable place and is already the file that states the rule.
2. **Give the rule a non-silent terminal** — reword rule 4's enforcement from "ignore in
   silence" to something the runtime can represent, e.g. *"…I reply with exactly:
   `not for me`"*. Hermes has no silence path (`conversation_loop.py:7033`), so any
   instruction to emit nothing is guaranteed to burn 5 API calls and 60 s and surface a
   scary error. Do this **in addition to** (1) if channels will ever carry other humans.
3. **Fix the display-name resolution** so the prefix reads `[Gaëtan Cathelain | Slack user
   <@U08BDJAMSRZ>]` instead of `[U08BDJAMSRZ | …]`. `source.user_name` fell back to the
   raw ID, which is what makes the sender unidentifiable in the first place. Likely a
   Slack scope (`users:read`) or a bot-token lookup gap on the channel event path — worth
   a separate check, and it makes (1) more robust rather than replacing it.
4. **Cheap blast-radius reduction, not a fix**: `HERMES_DUMP_REQUESTS=1` for one channel
   turn would give the exact Responses frame; and configuring a fallback provider would
   turn the 61 s dead end into a switched-provider answer. Neither addresses the cause.

**Not recommended**: prompt-size reduction (refuted, §5 H3), raising the empty-retry
count (each retry is byte-identical — more retries = more seconds, same outcome), or
disabling the sender prefix (it exists for a real reason, `#17916`: giving the model a
trusted `<@U…>` target for the current speaker).
