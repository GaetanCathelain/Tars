# GCN-59/GCN-61 probe — resume→wake row identity + session→Tars alternatives

Method: read-only source inspection on the Tars VM (`ssh gaetan@192.168.0.9`),
Hermes v0.20.0 install at `~/.hermes/hermes-agent`
(`~/.local/bin/hermes` → `~/.hermes/hermes-agent/venv/bin/python
~/.hermes/hermes-agent/hermes`). No config/env edits, no gateway restart, no
`sops`, no `hermes send`, no kanban cards. All claims below are INSPECTED
(source-level) unless tagged otherwise. Ledger background:
`docs/gcn50-knowledge-ledger.md` §1.3, §4.6, §4.7, §6 (esp. §6's open item
"CLI→gateway row-identity round trip … INFERRED — gcn50-rehydrate-review.md
§5.2").

## VERDICT — Q1: does the next gateway wake see CLI-`--resume`-appended turns?

**YES — INSPECTED (mechanism proven from source; not re-confirmed live in this
probe).** Upgrades the ledger's §6 "INFERRED" tag to INSPECTED. Two
independent, redundant mechanisms both point the same direction:

1. **Primary — unconditional per-turn reload.** Every gateway turn (cache-hit
   or cache-miss) calls `history = await
   self.async_session_store.load_transcript(session_entry.session_id)`
   (`gateway/run.py:16800`) *before* touching the agent cache. `load_transcript`
   (`gateway/session.py:3401-3417`) reads via
   `SessionDB.get_messages_as_conversation(session_id, repair_alternation=True)`
   against `~/.hermes/state.db` — the same file and the same `sessions`/message
   rows the CLI's `_preload_resumed_session` reads (`get_resume_conversations`)
   and `_flush_messages_to_session_db` writes
   (`append_messages_batch`). This reload result (`ctx.history` →
   `_build_gateway_agent_history` → `agent_history`) is what actually gets
   passed as `conversation_history` into `agent.run_conversation()`
   (`gateway/run.py:5124-5161`, `:5447-5459`) — including for a **warm,
   cache-reused** agent, not just a freshly built one. The only override is a
   corruption guard (`gateway/run.py:5127-5161`,
   `_select_cached_agent_history`) that prefers the *live in-memory* history
   over disk **only when disk is shorter** than memory (an FTS-write-failure
   detector). A CLI append makes disk **longer**, so the guard does not
   suppress it.
2. **Secondary — explicit cross-process-write cache invalidation.** Even
   before reaching the reload above, the agent-cache lookup
   (`gateway/run.py:4658-4663`, comment references issue #45966) compares the
   session row's current `message_count` against the count snapshotted when
   the agent was cached; on mismatch it evicts the cached `AIAgent` and forces
   a fresh one to be built from disk (`gateway/run.py:4816-4863`). This exists
   *specifically* for "another process (e.g. hermes dashboard) appends to the
   same session in the shared SessionDB" — the CLI is exactly such a process.
   `append_messages_batch` (`hermes_state.py:6758-6825`) increments
   `sessions.message_count` on every write
   (`UPDATE sessions SET message_count = message_count + ?…`,
   `hermes_state.py:6815/6821`), so a CLI-resume append reliably trips this
   guard too.

Both CLI and gateway construct `SessionDB()` against the same
`DEFAULT_DB_PATH = get_hermes_home() / "state.db"` (`hermes_state.py:260`) —
there is no per-surface DB split. Resume being source-agnostic
(`cli_agent_setup_mixin.py`, ledger §4.7) means the row the CLI mutated is
literally the row the gateway's `session_id` for that thread points at (same
`build_session_key` → `session_entry.session_id` → state.db row), not a
lookalike.

**What is *not* re-verified here (residual gap):** end-to-end composition —
that a genuine Slack thread reply after a real CLI resume append actually
shows the appended content in the model's next answer — is still only
INSPECTED, not PROVEN-live for this exact combination. §6 already listed this
as zero-coverage; the mechanism gap is closed, the live-composition gap is
not. See the confirmation recipe below.

## VERDICT — Q2: session→Tars alternatives beyond the two-rung shape

| Mechanism | Real agent turn in the right DM thread? | Cost / requirement | Verdict |
|---|---|---|---|
| `hermes send` raw (proven) | No — "no LLM, no agent loop" (ledger §1.3) | none | baseline, not a turn |
| `hermes chat --resume \| hermes send` (proven) | **Yes** — composed reply lands in-thread | ssh + session_id lookup | baseline, proven |
| `hermes cron create … --deliver slack` (proven, ledger §1.3/§5) | Yes, but delivers to **home DM**, not an arbitrary thread (`/bg` note) | cron entry | already-known alternative, not new |
| `kanban notify-subscribe --thread-id` (proven, ledger §1.3) | Yes — reply lands in the exact subscribed thread | a kanban card + subscription | already-known alternative, not new |
| **`hermes webhook`** | **Conditionally yes, but NOT into the existing thread session** | HTTP POST + per-route HMAC secret; see below | new, partial |
| **MCP server-pushed notifications** | **No** | n/a | new, negative |
| Everything else in `hermes --help` | Not turn-producing for this purpose | — | scanned, nothing new found |

### (a) `hermes webhook` — full shape

`hermes webhook {subscribe|list|remove|test}` (`hermes_cli/webhook.py`,
`gateway/platforms/webhook.py`, route handler `_handle_webhook` at
`webhook.py:584`).

- **`subscribe <name> [--prompt P] [--events E] [--description D]
  [--skills S] [--deliver TARGET] [--deliver-chat-id ID] [--secret S]
  [--deliver-only] [--script PATH]`** creates route `/webhooks/<name>`
  (also multiplex-scoped at `/p/{profile}/webhooks/{route_name}`,
  `webhook.py:291,298`).
- **Auth model**: mandatory per-route HMAC secret, validated at startup —
  a route with no secret refuses to serve (`webhook.py:252-268`,
  `:657-668`). `_validate_signature` checks the request signature
  (`webhook.py:1029`, `_hmac_str_equal` at `:158-169`). Escape hatch
  `secret: "INSECURE_NO_AUTH"` only accepted on a loopback bind
  (`webhook.py:30, 266-268`).
- **Can a webhook event carry a payload into an agent turn?** Yes —
  `--prompt` supports `{dot.notation}` payload refs, rendered by
  `_render_prompt` (`webhook.py:1197-1206`) into the user message for a
  **brand-new** agent turn (unless `--deliver-only`, which skips the LLM
  entirely and posts the rendered template verbatim, `webhook.py:16,
  814-820, 1260-1269`).
- **Crucial limitation — it is a NEW session, not the DM thread's session.**
  On a non-`deliver_only` route, the handler builds
  `session_chat_id = f"webhook:{route_name}:{delivery_id}"`
  (`webhook.py:864`) and runs the agent turn against that fresh, isolated
  session — it does **not** resume or extend the existing
  `agent:main:slack:dm:...:<thread_ts>` session. There is no history
  continuity with an ongoing Tars DM thread; only what's templated from the
  POST payload reaches the model.
- **Delivery targeting**: `--deliver-chat-id` only ever sets
  `deliver_extra.chat_id` (`hermes_cli/webhook.py:197-198`).
  `_deliver_cross_platform` (`webhook.py:1358-1409`) *does* support a
  `deliver_extra.thread_id` / `.message_thread_id` key, forwarded as
  `metadata={"thread_id": …}` to the platform adapter's `send()` — so
  landing in a specific Slack thread is mechanically possible — but no CLI
  flag exposes `thread_id`; reaching it requires hand-editing the route's
  `deliver_extra` in `config.yaml` (out of scope for a read-only probe, and
  itself a hard-ruled edit needing `.bak` + `flock` per this repo's rules).
- **Currently not provisioned**: `grep -n webhook ~/.hermes/config.yaml` on
  this VM returned nothing — no webhook routes exist on Tars today; this is
  greenfield if pursued.
- **Verdict**: real LLM turn, real Slack delivery, but a *disconnected*
  one-shot context (like the CLI two-rung shape minus session continuity)
  triggered by HTTP instead of ssh — no advantage over `chat -Q -q | send`
  for carrying a cooper session's *thread* context in, and it adds an HTTP
  attack surface + HMAC secret to manage. Only wins if the trigger truly
  needs to be an inbound HTTP POST (e.g. a third-party service), not a
  cooper-session push.

### (b) MCP client — server-pushed notifications

`tools/mcp_tool.py`. Hermes' `ClientSession` construction wires both
`message_handler` and `logging_callback` when the installed MCP SDK supports
them (`mcp_tool.py:284-319`, wired at `:2537,2539`). But the dispatch is
narrow:

- `_make_logging_callback` (`mcp_tool.py:2085-2114`): routes
  `notifications/message` (MCP structured log notifications) into
  **Hermes' own `agent.log`/`hermes_logging`** only — `logger.log(level,
  "MCP server log [%s]: %s", origin, data)`. Never touches conversation
  history, never triggers a turn.
- `_make_message_handler` (`mcp_tool.py:2119-2160`): pattern-matches on
  notification type. **Only** `ToolListChangedNotification` does anything
  (`_schedule_tools_refresh()` — refreshes the tool registry in the
  background). `PromptListChangedNotification` and
  `ResourceListChangedNotification` are explicitly logged and ignored
  (`"… (ignored)"`, `mcp_tool.py:2154-2156`). No other notification type is
  handled (`case _: pass`).
- No code path turns an inbound MCP notification into a new agent turn or a
  delivered message on any platform.

**Verdict: No.** The MCP client can receive server-pushed events, but they
are plumbing-only (log line, or silent tool-list refresh) — not a channel
for getting content into a live conversation or a Slack thread. Not viable
as a session→Tars carrier.

### (c) Rest of `hermes --help` — nothing new

Full top-level command list captured live (`hermes --help`, VM). Beyond
`chat`, `send`, `cron`, `kanban`, `webhook`, and `gateway/mcp` already
covered: `pairing` (DM pairing codes — auth, not messaging), `console`
("safe Hermes command console" — operator shell, not a turn producer),
`acp`/`serve`/`dashboard`/`proxy` (alternate front-ends to the same agent
loop, not new delivery paths), `sync`, `plugins`, `skills`, `insights`,
`monitoring`, `checkpoints`, `backup`/`import`, `profile`,
`prompt-size` — none produce or route a turn into an existing DM thread.
No further candidates found.

---

## Evidence — file:line index (Q1)

- `gateway/run.py:16800` — `history = await
  self.async_session_store.load_transcript(session_entry.session_id)`,
  unconditional per-turn reload before agent-cache lookup.
- `gateway/session.py:3401-3417` — `AsyncSessionStore/SessionStore
  .load_transcript` → `self._db.get_messages_as_conversation(session_id,
  repair_alternation=True)`; docstring: "state.db is the canonical store."
- `gateway/session.py:1257-1260` — `self._db = SessionDB()` (default ctor,
  same `~/.hermes/state.db` as CLI).
- `hermes_state.py:260` — `DEFAULT_DB_PATH = get_hermes_home() /
  "state.db"`.
- `gateway/run.py:4658-4663` and `:4816-4863` — cross-process
  message_count-mismatch cache invalidation ("Detect cross-process writes …
  invalidate the cache so a fresh agent re-reads from disk", issue #45966);
  fresh-agent construction path when cache misses/invalidates.
- `gateway/run.py:5124-5161` — `agent_history` built from `ctx.history`
  every turn via `_build_gateway_agent_history`; `_select_cached_agent_history`
  only prefers live memory when disk is *shorter* (write-failure guard),
  never suppresses a longer (CLI-appended) disk copy.
- `gateway/run.py:5447-5459` — `_conversation_kwargs["conversation_history"]
  = agent_history` fed into `agent.run_conversation()`.
- `hermes_state.py:6758-6825` — `SessionDB.append_messages_batch`:
  `UPDATE sessions SET message_count = message_count + ? …` — confirms CLI
  writes increment the counter the cache-invalidation guard reads.
- `hermes_cli/cli_agent_setup_mixin.py:576-655` (already in ledger) —
  `_preload_resumed_session`: `get_resume_conversations` read +
  `UPDATE sessions SET ended_at=NULL, end_reason=NULL`.
- `run_agent.py:2077-2087, 2281-2282` — `_flush_messages_to_session_db_unlocked`
  → `self._session_db.append_messages_batch(session_id=self.session_id, …)`,
  the CLI's write side, confirmed to hit the exact table/counter above.

## Evidence — file:line index (Q2)

- `hermes_cli/webhook.py:197-198`, `hermes_cli/subcommands/webhook.py:44` —
  `--deliver-chat-id` → `deliver_extra.chat_id` only.
- `gateway/platforms/webhook.py:252-268, 657-668, 1029, 158-169` — HMAC
  auth model, startup validation, `INSECURE_NO_AUTH` loopback-only escape.
- `gateway/platforms/webhook.py:584` — `_handle_webhook` route entry.
- `gateway/platforms/webhook.py:814-864` — `deliver_only` branch (no LLM)
  vs. agent-turn branch (`session_chat_id = f"webhook:{route_name}:
  {delivery_id}"`, a fresh session).
- `gateway/platforms/webhook.py:1197-1206, 1260-1269` — `_render_prompt`
  dot-notation templating; `deliver_only` template-becomes-message note.
- `gateway/platforms/webhook.py:1358-1412` — `_deliver_cross_platform`:
  chat_id resolution, `thread_id`/`message_thread_id` → `metadata`
  forwarded to `adapter.send()`.
- `tools/mcp_tool.py:284-319` — SDK capability probes for
  `message_handler` / `logging_callback`.
- `tools/mcp_tool.py:2085-2114` — `_make_logging_callback`: log-only sink.
- `tools/mcp_tool.py:2119-2160` — `_make_message_handler`: only
  `ToolListChangedNotification` acted on; prompt/resource notifications
  explicitly ignored; default case `pass`.

---

## Live confirmation Gaetan could trigger (Q1) — written, NOT run (~30 s)

**Goal:** prove, not just infer, that a genuine Slack thread reply picks up
content a CLI `--resume` append put into state.db moments earlier.

1. Pick a live DM thread with Tars and get its current gateway session id
   (read-only): on the VM,
   `~/.local/bin/hermes sessions list` (or the `state.db` `gateway_routing`
   table lookup already used in `gcn50-rehydrate-p2-threadindex.md`) to find
   `<gateway_session_id>` for the target `thread_ts`.
2. From the VM (or via ssh from cooper), append a marked turn with a fresh
   nonce, output discarded (no `hermes send` — this only touches state.db):
   ```
   ssh gaetan@192.168.0.9 '~/.local/bin/hermes chat --resume <gateway_session_id> -Q -q "Silently remember the codeword ZEBRA-<random4digits>. Do not mention it unless asked." ' >/dev/null
   ```
3. Within the same thread, Gaetan sends a real Slack message from his own
   client (not the probing session): "What was the codeword I just gave
   you?"
4. Confirm in two independent ways:
   - **Behavioral**: Tars' reply in-thread contains `ZEBRA-<same digits>`.
     Read via `slack_read_thread` (per skill: bot replies land in-thread,
     not at channel top level).
   - **Log corroboration** (only fires if the wake hit a warm cached
     agent — harmless if it doesn't):
     `grep 'message_count changed' ~/.hermes/logs/gateway.log` — presence of
     a line naming this session confirms the cross-process-write guard
     (mechanism 2 above) fired and forced a disk reread.
5. Negative control (optional, rules out coincidence): repeat steps 1-4
   without step 2 (no CLI append) — the reply should NOT know a codeword.

Total elapsed: one `ssh` CLI call (~5-10 s) + Gaetan's own Slack reply +
Tars' turn (~10-15 s per ledger's measured turn latencies) + one
`slack_read_thread` check. Nothing here touches config.yaml, sends via
`hermes send`, or requires a gateway restart.
