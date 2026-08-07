# WF5 — Thread behavior: mention-follow, context depth, emoji triggers

Investigation peer session (worktree `thread-behavior`), 2026-08-07.
Mandate: `docs/plans/thread-behavior-investigation.md` (+ a third question added
by Gaetan mid-flight: emoji-reaction triggers).

**Hermes is a source checkout on the VM**, not a venv install:
`/home/gaetan/.hermes/hermes-agent/` (git HEAD `6e87d43`); `~/.local/bin/hermes`
is a bash wrapper into it. Slack adapter = `plugins/platforms/slack/adapter.py`
(~9101 lines). All `adapter.py:NNNN` citations below are that file.

Companion evidence files in this directory:

| File | Covers |
|---|---|
| `source-gating.md` | inbound gate, gate order, knob inventory |
| `gating-verification.md` | adversarial refutation of the change proposal |
| `source-context.md` | session keying + what the model sees for a thread turn |
| `live-config-snapshot.md` | live `config.yaml` / `SLACK_*` / session store |
| `log-forensics-reference-thread.md` | gateway lines for the 6 reference-thread events |
| `emoji-trigger.md` | reaction_added feasibility + scopes + threading |
| `live-context-test.md` | the live cold-thread context probe |

---

## 0. Baseline: the reference thread, reconstructed from Slack

Read via the claude.ai Slack connector (authenticated as Gaetan `U08BDJAMSRZ`),
`conversations.replies` on channel `C08RWSTU9LK`, parent ts `1786135149.072829`.
Permalink: <https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786135283424339?thread_ts=1786135149.072829&cid=C08RWSTU9LK>

Times are **CEST as Slack renders them**; UTC = CEST − 2 h (the VM clock is UTC).

| # | Slack ts | UTC | Sender | Mention? | Tars replied? |
|---|---|---|---|---|---|
| parent | 1786135149.072829 | 20:39:09 | Gaetan | **yes** | yes — 6 s |
| r3 | 1786135219.118499 | 20:40:19 | Gaetan | **yes** | yes — 6 s |
| r5 | 1786135242.240479 | 20:40:42 | Gaetan | **yes** | yes — 17 s |
| **r7** | **1786135270.459679** | **20:41:10** | **Gaetan** | **NO** | **NO — silent** |
| r8 | 1786135283.424339 | 20:41:23 | Gaetan | **yes** | yes — 6 s |
| **r10** | **1786135310.012819** | **20:41:50** | **Gaetan** | **NO** | **NO — silent** |
| **r11** | **1786135439.766629** | **20:43:59** | **Nans `U05LKHLDV0A`** | NO | **NO — allowlist reject** |

Established with **zero new messages sent**: in a thread where Tars was an active
participant and had answered 13 s earlier, Gaetan's unmentioned r7 produced no
reply; same at r10. The three mentioned messages each answered in 6–17 s, so the
silence is not latency. **Thread participation buys nothing today.**

---

## 1. Q1 — Can Tars answer in-thread without a re-mention?

### 1.1 Why it is silent today

`_handle_slack_message` (`adapter.py:5239`) runs an elif ladder. With
`strict_mention: true` the flow hits a **bare `return` with no logger call** at
`adapter.py:5707-5708`. Confirmed by log forensics: r7 and r10 produced **zero
lines at any level** across `gateway.log`, `agent.log`, `errors.log`,
`mcp-stderr.log`. The gateway's file handler is at INFO with no DEBUG override —
but that is irrelevant, the drop is hard-silent *by code*, not suppressed by
level. This is why the behaviour is invisible in the logs.

### 1.2 The machinery already exists — it is dead code

`_should_wake_on_unmentioned_message` (`adapter.py:5164`) already implements
thread-follow, with these wake paths:

1. bot-sent thread root
2. mentioned-thread memory (`_mentioned_threads`)
3. active session for the thread ts
4. bot-authored root resolved via the Slack API
5. parent-text mention

It is never reached, because `strict_mention: true` returns one branch earlier.
Worse, under `strict_mention: true` the bot also never calls
`_register_mentioned_thread` (`adapter.py:5763-5768`), so no follow state has
been accumulating.

The knob is therefore **not** a new `thread_follow` key — it is **turning
`strict_mention` off** while keeping `require_mention: true`.

### 1.3 Safety — verified twice, source and live

**The allowlist runs unconditionally BEFORE any mention or thread logic.**

- Early reject: `adapter.py:5534-5550`; the WARNING is emitted at `:5546`.
- First mention/thread branch: `adapter.py:5654` — **104 lines later**.
- Two further `_is_user_authorized` checks downstream: `gateway/run.py:14610`
  (cold path) and `:8859` (busy path), explicitly present against shared-thread
  injection.

**Live proof, not just source:** Nans `U05LKHLDV0A` — a real colleague, in the
reference thread, in a thread Tars actively participates in — produced
`[Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK`
at 20:44:01 UTC (2 s after his message ts), plus 4 further occurrences. He never
reached the mention gate.

⇒ **A thread-follow change cannot leak to non-Gaetan users.** The allowlist is
not doing double duty with `require_mention`.

### 1.4 Adversarial verification of the proposed change

An independent Opus verifier was tasked to *refute* the claim
"`extra.strict_mention: false` + `require_mention: true` gives thread-follow and
changes nothing else". Verdict: **CONFIRMED WITH CAVEATS** — the three safety
assertions survive, two operational premises were wrong.

Survived:

- **Top-level containment HOLDS.** `_should_wake_on_unmentioned_message` receives
  the *raw* `event["thread_ts"]` (`adapter.py:5622`, `:5723`), **not** the
  ts-fallback used for the session key, so `:5187-5188` (`if not
  event_thread_ts: return False`) catches every top-level post. Unmentioned
  top-level channel messages stay dropped.
  - Latent defect noted, not currently reachable: wake path 2 (`_mentioned_threads`,
    `:5198-5202`) lacks an `is_thread_reply` guard. Unreachable for fresh
    top-level traffic because the set is keyed by mention-bearing ts.
- **No plain-text name matching.** `mention_patterns` is unset in both YAML and
  `.env`; `_slack_mention_patterns` (`:8423`) has no bot-name default;
  `strict_mention` never touches `is_mentioned`. Tars will **not** start waking on
  the bare word "tars".
- **Allowlist untouched** (§1.3).

Refuted / corrected:

- **A gateway RESTART is required — live-reload does not deliver this.**
  `PlatformConfig.extra` is built once at startup
  (`config.py:1441` → `:671` → `:706`, `base.py:2760`); no watcher rebuilds it for
  a connected adapter. `ExecReload` is SIGUSR1 → a full in-band restart anyway.
  Empirical proof: `config.yaml` was last written at 20:21 UTC and the gateway is
  still `MainPID=76255` from 19:40:45 UTC.
- **The reference thread would work immediately** (and survive the restart) — but
  *not* via `_mentioned_threads`, which is RAM-only (`adapter.py:975`), never
  persisted, and currently empty. It works via `_has_active_session_for_thread`
  (a session for `…:1786135149.072829` exists and `_should_reset` returns None
  under `mode: none`) and the parent-text path.

### 1.5 Blast radius

`allowed_channels` is **unset**. Read-only `users.conversations` on the bot token:

- `C08RWSTU9LK` (**#gcn-sandbox**, public, colleagues present) — the only channel
- 27 IMs
- no MPIMs, no private channels

1:1 DMs skip the whole gate block (`adapter.py:5654`), so DM behaviour is
unchanged. Effective new surface = **Gaetan's unmentioned replies in threads Tars
already touched, in one public sandbox channel.**

### 1.6 Two undisclosed caveats found by the verifier

- **Message edits become a new wake surface** (`adapter.py:5291-5297`); the dedup
  only covers already-processed messages.
- **Every past thread stays armed forever.** `session_reset: {mode: none}`
  (`config.yaml:140`) means no TTL path is reachable
  (`gateway/session.py:2210-2211`). A thread from weeks ago will still wake on an
  unmentioned reply.

### 1.7 ⛔ Why this is NOT applied — the blocking risk

**SOUL rule 4 (reply with a literal `·` instead of staying silent) only fires for
*non-Gaetan* senders.**

Every message newly admitted by this change is Gaetan's. So when the model decides
a message needs no answer, it has **no terminal**. Per `docs/facts.md` §SOUL,
Hermes has no silence terminal: deliberate model silence becomes N empty retries,
a 61 s stall, `empty_response_exhausted`, and a **user-facing error**.

Reference thread r7 is literally `"Ptin t'es nul"` and r10 is
`"Attends que je te tune toi"` — venting, not questions. Flipping the knob today
turns every such aside into a visible error, in a public channel with colleagues.

Per the mandate ("if it needs more than config (code/SOUL), PROPOSE with a diff
and stop"), and because a **restart** is also required, this is a **proposal**.
See §5.

### 1.8 Narrower knob? No.

- `thread_require_mention` exists but **tightens**, not loosens.
- `channel_overrides` affects model/prompt selection only.
- `allowed_channels` is a useful *companion cap*, not an alternative.

---

## 2. Q2 — How much of a thread does the model actually see?

### 2.1 Session keying

`build_session_key` (`gateway/session.py:1058`, non-DM branch `:1140-1173`):

```
agent:main:slack:<chat_type>:<team_id>:<channel_id>:<thread_ts>
```

The **user id is deliberately omitted when a thread ts is present**
(`thread_sessions_per_user` defaults False) — a thread is one shared session.
Top-level channel/DM messages get a **synthetic `thread_ts = ts`**
(`adapter.py:5588-5601`), so every top-level mention starts its own session.
Confirmed live against `state.db` `gateway_routing`.

### 2.2 Whole Slack thread, or session-only? — **Both, in sequence**

- **Cold start** (no session for that thread key): `_fetch_thread_context`
  (`adapter.py:7197`) calls **`conversations.replies(limit=30+1, inclusive=True)`**
  at `adapter.py:7251`, on the cold branch at `adapter.py:5800`, guarded by
  `_has_active_session_for_thread`. The root + last ≤30 replies are injected as a
  `[Thread context — …]` block built at `adapter.py:7340-7493`, with per-message
  tags `[thread parent]` / `[assistant]` / `[unverified]`, names and text
  newline-neutralised.
- **Steady state**: **zero** Slack API calls. Context is Hermes' own transcript.
- Two watermark-scoped delta re-fetches exist: an explicit @mention on a live
  thread, and a once-per-process rehydrate after restart.
- No `conversations.history` anywhere in the gateway.
- Fails **silently to `""`** if the history scope is missing.

### 2.3 A thread that predates Tars' participation — **it sees it**

Slack's event payload carries no history; the adapter enriches. Because there is
no session for that thread key, the cold-start path fires and the pre-existing
conversation is injected. So being mentioned mid-thread gives Tars the last ≤30
replies plus the root.

**Corollary, and this is the useful bit:** unmentioned messages are invisible as
*triggers* but fully visible as *context*. Gaetan's r7/r10 were never processed —
yet they are part of the thread and will be injected into any future cold read of
it.

### 2.4 What the model receives per message

`run.py:16031-16034`:

```
[{name} | Slack user <@{U…}>] {text}
```

Channel/thread turns only — **DMs get no prefix**
(`is_shared_multi_user_session` returns False for `chat_type == "dm"`). The thread
block is prepended as `…\n\n[New message]\n{prefixed}`, with `[Replying to: "…"]`
ahead of it. (This prefix is the mechanism behind the identity-frame bug class in
`docs/facts.md` §SOUL.)

### 2.5 Session lifetime

`session_reset.mode: none` (`config.yaml:140`) ⇒ **thread sessions never expire**
on Tars. No turn cap. Only token-based trimming: gateway hygiene at 0.85 of the
context window (`run.py:16829`), agent-side compressor at 0.50.

### 2.6 The 30-message bound is not a real ceiling

Tars has ~20 Slack MCP tools including `mcp__slack__conversations_replies` and
`conversations_history`, running on **Gaetan's user token** (Gaetan's visibility).
One fired at 20:39:32 UTC during the reference thread. So Tars can read past the
injected window whenever it decides to — the 30-message limit bounds what it gets
*for free*, not what it can *reach*.

### 2.7 Live confirmation

See §4 and `live-context-test.md`.

---

## 3. Q3 — Emoji reaction triggers (e.g. an "add to Kanban" emoji)

### 3.1 Verdict: **PARTIAL — the feature ships, it is switched off**

Not a missing capability. Everything is already built:

- `@self._app.event("reaction_added")` registered at `adapter.py:1996`
- `_handle_slack_reaction()` at `adapter.py:4773`
- shipped config key **`slack.reaction_triggers`**
- Kanban already enabled — `toolsets: [hermes-cli, kanban]` gives 12 `kanban_*`
  tools on CLI *and* Slack, so the **action** exists; only the **trigger** is off.

Full inbound event registry (`adapter.py:1952-2039`): `message`, `app_mention`,
`app_home_opened`, `app_context_changed`, `file_shared`, `file_created`,
`file_change`, **`reaction_added`**, **`reaction_removed`**,
`assistant_thread_started`, `assistant_thread_context_changed`, plus a `.*`
catch-all no-op ack. Transport is **Socket Mode** (`adapter.py:1188`).

### 3.2 The single blocker: a missing Slack scope

Live `auth.test` `x-oauth-scopes` shows **17 granted scopes — no `reactions:read`
and no `reactions:write`**. Without `reactions:read`, Slack never delivers the
event, so the handler can never run.

**Pre-existing lapse, flagged:** `reactions:write` is also absent, so **Tars' 👀
acknowledgement reaction has been silently failing since cutover.** The adapter
swallows the failure. This is an existing broken control, not something this
change introduces.

### 3.3 Agent-class is NOT a blocker

Slack's own AI-apps documentation
(<https://docs.slack.dev/ai/developing-ai-apps/>) states: *"You can also subscribe
to `reaction_added` events to collect feedback based on reactions."* And Hermes'
own manifest generator already puts `reactions:read` + `reaction_added` in the
**base set** for `messaging_experience: agent`
(`hermes_cli/slack_cli.py:91`, `:104`).

### 3.4 Would it answer in the right thread? — **Yes**

`reaction_added` payloads carry `item.channel` and `item.ts` but **no
`thread_ts`**. The handler already solves this: it calls `conversations.replies`
on `item.ts` and takes the returned `thread_ts` (`adapter.py:4861-4877`).

Verified live against the reference thread: passing a *reply's* ts
(`1786135283.424339`) returns `thread_ts=1786135149.072829` — the true parent.
This matches Slack's documented rule ("avoid using a reply's `ts`; use its parent
instead"). So reacting to any message in a thread lands Tars' answer in that
thread.

### 3.5 Allowlist coverage — no regression

The reaction synthesises a message with **the reactor** as `user` and routes it
through `_handle_slack_message`, whose auth chokepoint is `adapter.py:5528-5549`
— the same WF4 "Early reject of unauthorized user" line. It bypasses **only** the
mention gate, via `_hermes_force_process` (`:4928`, `:5679`). A colleague reacting
with the emoji would be rejected exactly as Nans was.

### 3.6 Alternatives considered and rejected

| Path | Verdict |
|---|---|
| `slack.reaction_triggers` (shipped feature) | **best** — zero code |
| Slack message shortcut / message action | not implemented in Hermes (zero `@app.shortcut` hits) |
| `hermes cron` polling `reactions.get` | strictly worse — latency, quota, no thread fidelity |
| Slack Workflow Builder → webhook | strictly worse — extra moving part, separate auth |

Note on the config shape: `reaction_triggers` in **list form** triggers on **any**
message, not only Tars' own — that is what makes "react to a colleague's message
to file it" work, and it is also why the allowlist check above matters.

---

## 4. Live measurement

### 4.1 What was NOT needed

Q1's gating behaviour was established from **existing** traffic (§0) — no probe
messages were sent into the reference thread.

### 4.2 The cold-thread context probe — attempt 1 FAILED (invalid method)

A fresh thread was seeded in `C08RWSTU9LK`, deliberately cold for Tars, to
discriminate *adapter-injected context* from *model tool call*. Both messages
were sent **via the claude.ai Slack connector**:

| ts | Content | Result |
|---|---|---|
| `1786136615.681079` | top-level, **no mention**, contains `TARSTHREAD-7Q4X` | no reply (expected) |
| `1786136621.808629` | in-thread, **WITH** `<@U0BBH85NAKH>` mention | **no reply — NOT expected** |

Both landed in Slack, but the gateway logged **zero lines for either**, including
the explicitly mentioned one, while it kept processing DM traffic normally.
See `live-context-test.md`.

### 4.3 Root cause — the connector cannot trigger Tars at all

`_event_declares_bot_sender`, **`adapter.py:3130`**:

```python
if event.get("app_id") and not event.get("client_msg_id"):
    return True   # comment cites issue #35777 — deliberate
```

Raw message metadata, fetched read-only via `conversations.replies`:

| Field | connector-sent | natively typed |
|---|---|---|
| `app_id` | `A08SF47R6P4` | absent |
| `client_msg_id` | **absent** | real UUID |
| `bot_id` | absent (posts on Gaetan's *user* token) | absent |
| `user` | `U08BDJAMSRZ` | `U08BDJAMSRZ` |

So connector messages are flagged `sender_is_bot=True` and dropped at
**`adapter.py:5339`** (`if allow_bots == "none": return`) — `SLACK_ALLOW_BOTS` /
`config.extra.allow_bots` are unset on the VM, so the `"none"` default applies.
The `"mentions"` bypass at `:5341` is **never reached**, which is why even an
explicit mention produced nothing. No log line fires because the entry log is
DEBUG-gated and the level is INFO.

⇒ **The claude.ai Slack connector is READ-ONLY with respect to Tars.** It cannot
trigger a reply in a channel *or* a DM (the bot-sender filter precedes any
`channel_type` branch — same single `_handle_slack_message` for all `message`
events). Any live probe needing a response must be sent **natively**, i.e. with
the VM's `SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN` user credentials.
See `connector-message-invisibility.md`.

### 4.4 The cold-thread context probe — attempt 2, native Gaetan: **PASS**

Sent with the VM's stored user credentials (tokens on stdin via `curl -K`, never
on argv). `auth.test` → `ok:true`, `user_id U08BDJAMSRZ`, team `T7V1UGJ82` —
native Gaetan, no `app_id`, no `*Sent using*` stamp. **Exactly one message sent**,
into the *same* cold thread whose root Tars had never received.

| | |
|---|---|
| probe ts | `1786137254.622319` (21:14:14.622 UTC) |
| gateway saw it | **yes** — `21:14:15,752 gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK … reply_to_id=1786136615.681079` |
| Tars' answer | ts `1786137262.280399`, text **exactly `TARSTHREAD-7Q4X`** (`response=15 chars`) |
| **tool calls** | **NONE** — `api_calls=1/500`, `tool_turns=0`, no `tool_executor` line; `mcp-stderr.log` unchanged (last entry 21:02:16, 12 min earlier) |
| session | cold, `history=0` |
| session key | `agent:main:slack:group:T7V1UGJ82:C08RWSTU9LK:1786136615.681079` (chat_type `group`) |
| latency | **7.66 s** end-to-end (6.3 s gateway-internal) |

The stored turn was read back from `state.db` (`mode=ro`). The model's **entire**
1101-char input was:

```
[Replying to: "…TARSTHREAD-7Q4X…"]
[Thread context — prior messages in this thread (not yet in conversation history):]
[thread parent] U08BDJAMSRZ: …TARSTHREAD-7Q4X…
U08BDJAMSRZ: …
[End of thread context]
[New message]
…
```

Independently confirmed in Slack by reading the thread back: Tars' reply
`TARSTHREAD-7Q4X` is present in-thread.

**Verdict — §2 proven end to end.** The adapter injects real Slack thread history
on cold start. Messages that were never *processed* — the connector-dropped root,
and by the same mechanism any unmentioned message — remain **fully visible as
context**, including history that predates Tars' participation. The connector
stamp blocks *triggering*, not *visibility*. And Tars produced the code with
**zero tool calls**, so this is adapter injection, not the model reading Slack.

Two details worth recording:

- No `[unverified]` tag appeared (both senders are allowlisted).
- Inside the thread-context block, senders render as **raw user IDs**
  (`U08BDJAMSRZ:`), not display names — distinct from the `[New message]` prefix
  shape `[{name} | Slack user <@U…>]` (`run.py:16031-16034`). Relevant to the
  identity-frame bug class in `docs/facts.md` §SOUL.
- The hydrate is `conversations.replies(limit=31)`, **one-shot per thread
  session** — a thread deeper than 30 replies is not fully visible on cold read.

---

## 5. Proposal — NOT applied

Nothing on the VM was modified during this investigation: no config edit, no
`.env` edit, no restart, no `sops -d`, no Slack write beyond the two authorised
probe messages.

### 5.1 Thread-follow (Q1) — 3 parts, in this order

**(a) SOUL amendment — MUST land first.** Give Gaetan-authored messages a silence
terminal, mirroring the existing rule 4. Without it, every unmentioned aside in a
followed thread risks a 61 s stall and a user-facing error.

**(b) Config**, under `flock ~/.hermes/.wf3.lock` with a `.bak` first, merged not
appended:

```yaml
gateway:
  platforms:
    slack:
      enabled: true
      require_mention: true        # unchanged — top-level still needs a mention
      strict_mention: false        # was true
      extra:
        strict_mention: false      # authoritative: _slack_strict_mention() reads extra first (:8268)
      # optional companion cap, recommended:
      # allowed_channels: [C08RWSTU9LK]
```

Both keys are set because the plain key alone reaches the adapter only via the
`SLACK_STRICT_MENTION` env var, stamped under a `not os.getenv(...)` guard
(`adapter.py:8994`) — already `"true"` in the running process.

**(c) Gateway restart** — required (§1.4), live-reload will not deliver it.

**Verification after applying:** post an unmentioned reply in the reference thread
and confirm an answer; confirm a *top-level* unmentioned message is still dropped;
confirm the allowlist WARNING still fires for a non-Gaetan in-thread message.

### 5.2 Emoji trigger (Q3) — 3 parts

**(a) Slack app dashboard (human action — needs the logged-in browser):** add the
`reactions:read` scope and the `reaction_added` event subscription to app
`A0BC0GXH78R`. Skip `reaction_removed`. Consider adding `reactions:write` at the
same time to repair the broken 👀 ack (§3.2).
Slack scopes are additive, so **no token change and no SOPS rotation expected** —
confirm with `auth.test` after re-install.

**(b) Config**, one line under `gateway.platforms.slack`:

```yaml
      reaction_triggers: [kanban]
```

**(c) SOUL**, one line mapping `reaction:added:kanban` to the Kanban tool.

### 5.3 Decisions for Gaetan

1. Apply thread-follow, given it needs a **restart** + a **SOUL edit** first?
2. Cap it with `allowed_channels: [C08RWSTU9LK]`, or leave it open to any channel
   Tars joins later?
3. Accept that **every past thread stays armed forever** (`session_reset: none`),
   or introduce a TTL?
4. Repair `reactions:write` (broken 👀 ack) in the same app-dashboard visit?
