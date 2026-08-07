# Operational facts — verified live 2026-08-07

Everything below was measured on the Tars VM (Hermes v0.20.0, Ubuntu 24.04)
during WF3–WF5. Evidence pointers are files under `status/probes/`. Re-verify
after any Hermes upgrade.

## Hermes CLI / runtime

| Fact | Evidence |
|---|---|
| Oneshot chat is `hermes chat -Q -q '<query>'` — `--oneshot` does not exist | `wf3-s2`, `wf3-s4` |
| `hermes mcp list` is registry-only (never connects); liveness = `hermes mcp test <name>` | `wf3-s4` |
| MCP tool prefix is `mcp__<server>__<tool>` (double underscore) | `wf3-s4` |
| Systemd units are `--user`: `hermes-gateway.service`, `hermes-cloakbrowser.service`; `export XDG_RUNTIME_DIR=/run/user/$(id -u)` over ssh | `wf3-s5` |
| journald carries WARNINGs only; INFO lives in `~/.hermes/logs/{gateway,agent,errors,mcp-stderr}.log` | `wf4/p01` |
| `config.yaml` live-reloads (~30 s); no restart needed for most changes (kanban enable needed none) | `wf5/kanban-implement` |
| `hermes memory setup` and `hermes skills install` (no `--yes`) CANCEL headless and still exit 0 — use `hermes config set` | lane B handoff |
| `hermes auth add … --type oauth` does NOT write `model.base_url` — set it explicitly and verify no openrouter default survives | `wf3-s5` |
| `tools/mcp_tool.py _build_safe_env()` filters subprocess env: `.env` keys are NOT visible to MCP servers — only the server's own `env:` block + PATH/HOME allowlist | `wf3-s2` |
| Async delivery to the home channel: `hermes cron create "1m" "…" --deliver slack --repeat 1` (bare platform name ⇒ `SLACK_HOME_CHANNEL`). `/bg` always replies to the INVOKING surface — it can never deliver to home | `wf4/p13` |
| Skills live at `~/.hermes/skills/<name>/SKILL.md` and are loaded into model context unprompted when relevant | `wf5/orca-implement` |
| Kanban: single top-level `toolsets: [hermes-cli, kanban]` key unlocks 12 `kanban_*` tools on CLI **and** Slack; dispatcher spawns real workers that drive cards ready→running→done | `wf5/kanban-*` |
| Known intermittent: multi-step Codex turns can die with "response remained incomplete after 3 continuation attempts"; plain/single-tool turns unaffected. Retry once before concluding failure | `wf4/p13`, triage: later |
| tirith security scanner: enabled but not installed — command vetting is pattern-matching only (pre-existing; triage: later) | `wf3-s5` |

## Slack platform

| Fact | Evidence |
|---|---|
| Tars is an **Agent-class** app: Slack blocks ALL slash commands in its chat surface (it is threads under the hood). `/sethome` equivalent = write `SLACK_HOME_CHANNEL` to `.env` + restart (source-verified getenv-only) | `cutover-sethome` |
| Tars replies **in-thread**: `conversations.history` is blind to its replies — always poll `conversations.replies` on the sent message ts | `wf4/p01` |
| The allowlist reject is a **positive WARNING**: `[Slack] Early reject of unauthorized user <id> in channel <id>` (adapter, pre-dispatch, pre-media-fetch) — grep for it, never settle for absence-of-reply | `wf4/p16` |
| `SLACK_STRICT_MENTION`/`SLACK_ALLOWED_USERS` are written to `os.environ` by the plugin AFTER exec — `/proc/<pid>/environ` shows NO `SLACK_*` on a healthy gateway | `wf4/p03` |
| Bots don't auto-join channels — invite manually. One Slack token pair = one live gateway | `cutover-notes`, D2 |
| The claude.ai `claude_ai_Slack` MCP connector authenticates **as Gaetan** (user token `U08BDJAMSRZ`) — never usable as a non-Gaetan sender | `wf4/p03` |
| **`slack.allow_bots` is security-critical, one line, and non-obvious.** Setting it to `mentions`/`all` does TWO things at once: it stops the adapter's bot-sender drop (`adapter.py:5339`) **and** arms a gateway branch (`gateway/authz_mixin.py:499-501`) that returns authorized for any bot **with no `SLACK_ALLOWED_USERS` comparison at all** (upstream intent, #4466, with its own test). Default is `none`, and it is unset here — dormant. The adapter's startup INFO log contradicts this: it says the other bot's id must be in the allowlist. Treat this key on the same do-not-change footing as `SLACK_ALLOWED_USERS`. **What keeps the connector from triggering Tars is this default, not the check ordering** — the allowlist alone would admit a connector post, since it carries `user=U08BDJAMSRZ` | `wf5/audit-slack-allowlist-bypass` |
| Allowlist integrity otherwise verified: it keys on Slack-asserted `event["user"]` over the authenticated Socket-Mode websocket (not payload-forgeable) and fails **closed** on empty/unset/no-user/error. Every external branch (DM, public/private channel, MPIM, thread, app_mention, edited, file_share, slash, interactive, reaction) hits a check; only `is_internal` (cron, `/bg`) skips it, and that is agent-authored, not Slack-reachable | `wf5/audit-slack-allowlist-bypass` |
| IDs: Gaetan `U08BDJAMSRZ` · Tars bot user `U0BBH85NAKH` · app `A0BC0GXH78R` · team `T7V1UGJ82` · home DM `D0BBYNM01BL` · test channel `C08RWSTU9LK`. VM env names: `SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN` (cookie sent as-stored, URL-encoded) | `wf4/p01` |

## SOUL / persona design

**Identity-frame bug class** (`wf4/diag-empty-response`, fixed same day): any
hard rule of the form "answer only X" MUST name X's platform identity for every
surface. Channel (multi-user) turns prefix messages with `[<user-id> | …]`;
without an ID→person mapping the model cannot recognize X and obeys the rule by
staying silent. Hermes has **no silence terminal** — deliberate model silence
becomes N empty retries, a 61 s stall and a user-facing error. Rules must
prescribe a minimal literal reply (Tars uses `·`) instead of "ignore in
silence". The adapter allowlist rejects strangers before the model, so the
model-level rule is a backstop only.

## Thread behavior

Verified 2026-08-07 (WF5). Hermes on the VM is a **source checkout**, not a venv
install: `/home/gaetan/.hermes/hermes-agent/` (git `6e87d43`); `~/.local/bin/hermes`
is a bash wrapper into it. All `adapter.py:N` below =
`plugins/platforms/slack/adapter.py`. Evidence: `status/probes/wf5/`.

### Gating — what wakes Tars

| Fact | Evidence |
|---|---|
| **Gate order is allowlist → mention.** `SLACK_ALLOWED_USERS` early-reject at `adapter.py:5534-5550` (WARNING at `:5546`) runs **104 lines before** the first mention/thread branch (`:5654`); two more `_is_user_authorized` checks downstream (`gateway/run.py:14610` cold, `:8859` busy) exist against shared-thread injection. Live: a colleague's in-thread message produced the reject WARNING in 2 s | `wf5/log-forensics-reference-thread`, `wf5/source-gating` |
| **Unmentioned messages are dropped with ZERO log output**, at any level — `strict_mention: true` hits a bare `return` with no logger call (`adapter.py:5707-5708`). Not a log-level artifact; never diagnose this from log silence | `wf5/log-forensics-reference-thread` |
| **Thread participation buys nothing today.** In a thread where Tars had answered 13 s earlier, Gaetan's unmentioned replies were dropped; the mentioned ones answered in 6–17 s | `wf5/thread-behavior` §0 |
| **Thread-follow already exists as dead code**: `_should_wake_on_unmentioned_message` (`adapter.py:5164`) implements 5 wake paths (bot-sent root, mentioned-thread memory, active session by thread ts, bot-authored root via API, parent-text mention), unreachable because `strict_mention` returns first. The knob is turning `strict_mention` **off**, keeping `require_mention: true` — top-level stays gated (`:5187-5188` needs a real `thread_ts`) | `wf5/source-gating`, `wf5/gating-verification` |
| `strict_mention`/`require_mention` are **NOT env vars** — they live in `config.yaml` under `gateway.platforms.slack`. `strict_mention` is absent from the config.py shared-key bridge and reaches the adapter via `SLACK_STRICT_MENTION`, stamped under a `not os.getenv()` guard (`adapter.py:8994`); `_slack_strict_mention()` (`:8268`) reads `extra` first | `wf5/live-config-snapshot`, `wf5/source-gating` |
| **A gateway RESTART is required to change it** — `PlatformConfig.extra` is built once at startup (`config.py:1441→671→706`, `base.py:2760`); no watcher rebuilds it for a connected adapter, and `ExecReload` is SIGUSR1 → full in-band restart. Live-reload does NOT cover this (it does cover most other keys) | `wf5/gating-verification` |
| No plain-text name matching: `mention_patterns` unset, `_slack_mention_patterns` (`:8423`) has no bot-name default. Turning off `strict_mention` will not make Tars wake on the bare word "tars" | `wf5/gating-verification` |
| `_mentioned_threads` is **RAM-only** (`adapter.py:975`), never persisted, and is not populated at all while `strict_mention: true` (`:5763-5768`) | `wf5/gating-verification` |
| Message **edits** are a wake surface (`adapter.py:5291-5297`); `session_reset: {mode: none}` means no TTL path is reachable (`gateway/session.py:2210-2211`), so every past thread stays armed forever | `wf5/gating-verification` |
| `allowed_channels` is unset. Read-only `users.conversations`: Tars is in exactly **one** channel (`C08RWSTU9LK` #gcn-sandbox, public) + 27 IMs — no MPIMs, no private channels. 1:1 DMs skip the gate block entirely (`:5654`) | `wf5/gating-verification` |

### Context — what the model sees in a thread

| Fact | Evidence |
|---|---|
| Session key (`gateway/session.py:1058`, non-DM `:1140-1173`): `agent:main:slack:<chat_type>:<team>:<channel>:<thread_ts>`. **User id is omitted when a thread ts is present** — a thread is one shared session. Top-level messages get a synthetic `thread_ts = ts` (`adapter.py:5588-5601`), so each starts its own session | `wf5/source-context`, `wf5/live-context-test-native` |
| **Cold start fetches the real Slack thread**: `_fetch_thread_context` (`adapter.py:7197`) → `conversations.replies(limit=30+1, inclusive=True)` (`:7251`) on the cold branch (`:5800`), guarded by `_has_active_session_for_thread`; injected as a `[Thread context — …]` block (`:7340-7493`). **Steady state makes ZERO Slack calls** — context is Hermes' own transcript. No `conversations.history` anywhere in the gateway. Fails silently to `""` if the history scope is missing | `wf5/source-context` |
| **A thread that predates Tars IS visible**, and so are messages Tars never *processed*. Proven live: a thread root Tars never received (it was connector-dropped) was recited verbatim by Tars on a cold mention — `tool_turns=0`, `history=0`, 7.66 s. Unmentioned ≠ unseen: invisible as a **trigger**, fully visible as **context** | `wf5/live-context-test-native` |
| Injected block shape: `[Replying to: "…"]` + `[Thread context — prior messages in this thread (not yet in conversation history):]` / `[thread parent] <UID>: …` / `[End of thread context]` / `[New message]`. Inside the block senders render as **raw user IDs**, unlike the `[New message]` prefix `[{name} \| Slack user <@U…>]` (`run.py:16031-16034`). `[unverified]` tags non-allowlisted senders. **DMs get no prefix at all** | `wf5/live-context-test-native`, `wf5/source-context` |
| Hydrate is **one-shot per thread session** at 31 messages — a deeper thread is not fully visible on cold read. Two watermark-scoped delta re-fetches exist (explicit @mention on a live thread; once-per-process restart rehydrate) | `wf5/source-context`, `wf5/live-context-test-native` |
| `session_reset.mode: none` ⇒ **thread sessions never expire**; no turn cap. Only token-based trim: gateway hygiene at 0.85 of context (`run.py:16829`), agent compressor at 0.50 | `wf5/source-context` |
| Canonical session store is **`~/.hermes/state.db`** (SQLite: `sessions`, `messages`, `messages_fts*`, `gateway_routing`). `~/.hermes/sessions/sessions.json` is a **legacy routing mirror only** — do not read it as truth. No `sqlite3` binary on the VM; use `python3 -c` with a `file:…?mode=ro` URI | `wf5/source-context`, `wf5/live-context-test-native` |
| The 31-message window is not a real ceiling: Tars holds ~20 Slack MCP tools (`mcp__slack__conversations_replies`, `conversations_history`, search) on **Gaetan's user token**, and can read further whenever it decides to | `wf5/source-context` |

### The claude.ai Slack connector CANNOT trigger Tars

**Read-only for probing — channels and DMs alike.** `_event_declares_bot_sender`
(`adapter.py:3130`) returns True for `app_id` set **and** `client_msg_id` absent
(deliberate; the comment cites issue #35777) — exactly the shape of a connector
post (`app_id=A08SF47R6P4`, no `client_msg_id`, no `bot_id`, `user=U08BDJAMSRZ`).
It is then dropped at `adapter.py:5339` (`allow_bots == "none"`, the unset
default); the `"mentions"` bypass at `:5341` is never reached, so **even an
explicit `<@U0BBH85NAKH>` mention produces nothing**, and no log line is emitted
(entry log is DEBUG-gated, level is INFO). The bot-sender filter precedes any
`channel_type` branch, so DMs are affected too.

⇒ Any live probe that needs a *response* must be sent **natively**, with the VM's
`SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN` user credentials (tokens on stdin
via `curl -K`, never argv). Verify with `auth.test` → `user_id U08BDJAMSRZ` first.
The connector remains correct for *reading* Slack. Evidence:
`wf5/connector-message-invisibility`, `wf5/live-context-test-native`.

### Emoji / reaction triggers — the feature ships, it is switched off

| Fact | Evidence |
|---|---|
| `reaction_added` **is already handled**: registered at `adapter.py:1996` → `_handle_slack_reaction()` (`:4773`). Shipped config key **`slack.reaction_triggers`** (list form triggers on ANY message, not only Tars' own). Kanban's 12 `kanban_*` tools are already live, so the *action* exists — only the *trigger* is off | `wf5/emoji-trigger` |
| Full inbound event registry (`adapter.py:1952-2039`): `message`, `app_mention`, `app_home_opened`, `app_context_changed`, `file_shared`, `file_created`, `file_change`, `reaction_added`, `reaction_removed`, `assistant_thread_started`, `assistant_thread_context_changed`, + a `.*` catch-all no-op ack. Transport = **Socket Mode** (`adapter.py:1188`) | `wf5/emoji-trigger` |
| **Sole blocker: `reactions:read` is not granted.** Live `auth.test` `x-oauth-scopes` shows 17 scopes, no `reactions:read`. Without it Slack never delivers the event | `wf5/emoji-trigger` |
| **Pre-existing lapse: `reactions:write` is also absent**, so Tars' 👀 acknowledgement reaction has been silently failing since cutover | `wf5/emoji-trigger` |
| Agent-class is **not** a blocker — Slack's AI-apps docs explicitly endorse subscribing to `reaction_added`, and Hermes' own manifest generator puts `reactions:read` + `reaction_added` in the base set for `messaging_experience: agent` (`hermes_cli/slack_cli.py:91`, `:104`) | `wf5/emoji-trigger` |
| **It would answer in the right thread**: the handler resolves `item.ts` → true parent via `conversations.replies` (`adapter.py:4861-4877`) — reaction payloads carry no `thread_ts`. Verified live: passing a *reply's* ts returns the parent | `wf5/emoji-trigger` |
| **Allowlist covers it** — the reaction synthesizes a message with the **reactor** as `user`, routed through the same `_handle_slack_message` auth chokepoint (`adapter.py:5528-5549`). It bypasses only the mention gate, via `_hermes_force_process` (`:4928`, `:5679`) | `wf5/emoji-trigger` |
| Slack scopes are additive ⇒ adding one needs an app re-install but **no token change and no SOPS rotation expected**; confirm with `auth.test` after | `wf5/emoji-trigger` |
