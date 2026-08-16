# GCN-50 knowledge ledger — what works, what doesn't, what it costs

Distilled from the 17 `status/probes/gcn50-*.md` files (probes, three adversarial
reviews, three live end-to-end runs, one teardown log), all read in full.
`PROPOSED-SPEC.md` §"E2E validation (live)" and §"Evidence index" were used only to
cross-check that the file set is complete — the probe files are the sources of truth.

**This file records facts, not design.** Where a probe is only intelligible via its
original wire number, the mechanism is named in plain words instead. No
recommendation, no chosen option, no architecture is carried over.

## Confidence tags

Used exactly as the probes themselves tagged their claims — nothing is upgraded here.

- **PROVEN-live** — exercised on real machines, outcome observed out-of-band
  (nonce, file mtime, Slack read, log line).
- **INSPECTED** — read from source, `--help`, docs, a DB schema, or file modes;
  not exercised. (Probes call this "VERIFIED from source"/"VERIFIED".)
- **INFERRED** — a probe's own reasoned conclusion from the above, explicitly
  labelled as such in the probe.

All measurements are Claude Code **2.1.232** on cooper (Linux 5.15, uid 1000
`gaetan`) and Hermes **v0.20.0** on the VM `192.168.0.9`, dated 2026-08-14, unless
stated.

---

## 1. WHAT WORKS

### 1.1 Delivering a message into a running Claude Code session

**A headless one-shot `claude -p` can call ListAgents and SendMessage.**
PROVEN-live — `gcn50-q12-headless-sendmessage.md` §2–§3, §5. ListAgents returned
the live 12-session peer roster; SendMessage returned
`{"success":true, …,"msg_id":"3983426e-…"}`, `permission_denials:[]`, exit 0, no
permission hold. Conditions: run with `--dangerously-skip-permissions`; that probe
did **not** verify receipt on the receiver side (§4 says so explicitly).

**Receipt was closed separately, 8/8.** PROVEN-live —
`gcn50-q15-lean-sendmessage.md` §Method, §Caveats. Every configuration A–H was
verified by grepping a per-run nonce out of the receiver's `received.txt`, "not by
trusting the CLI's own `"result":"Sent."` text". Receiver conditions: interactive
`claude` under a tmux pty, `--settings '{"crossSessionInbound":"accept"}'`, fixed
name `-n q15-receiver`, scratch cwd with no CLAUDE.md, **idle** at measurement time.

**The leanest still-delivering invocation is config G** — PROVEN-live,
`gcn50-q15-lean-sendmessage.md` §Results/§Recommended:
`claude -p '<prompt>' --output-format json --dangerously-skip-permissions --model haiku
--strict-mcp-config --mcp-config <empty-mcp.json> --system-prompt "<1-line override>"
--setting-sources "" --tools "SendMessage"`, run from a cwd with no CLAUDE.md.
Every flag stacked additively; delivery survived all the way down.

**A plain non-Claude process, not a descendant of the target, can inject straight
into a session's inbox socket — no SendMessage tool, no Claude hop.** PROVEN-live —
`gcn50-q12-nonclaude-socket.md` §4–§5. Sender was `env -i … python3` (probe notes
bash+socat is equivalent); the receiver pid never appeared in the sender's ppid
chain. Wire is newline-delimited JSON: optional `{"type":"auth","token":"…"}` then
`{"type":"user","message":{"role":"user","content":"…"}}`. Controlled variations
isolated the gate: correct token / wrong token / **no auth line at all** all
DELIVERED (A, B, B2) against a `crossSessionInbound: accept` receiver. The only
hard requirement is same-OS-user socket access.

**A resident Go process can be a first-class bidirectional peer.** PROVEN-live —
`gcn50-e2e-go-bridge.md` §VERDICT, §1–§3. It (a) wrote auth+user frames into a
session's socket and the session rendered a real inbound peer turn; (b) published
`~/.claude/sessions/<pid>.json` + `.key` and was discovered by that session's own
`ListAgents` as `go-bridge-peer`; (c) received 3 PASS `SendMessage` frames back,
including a full ask/answer round-trip. ~150 lines of stdlib Go, go 1.26.5. The
session labels it "another Claude session on this machine" — indistinguishable from
a Claude peer to the sender.

Registration facts that made discovery work (INSPECTED against a live session's
`96610.json`, then reproduced) — `gcn50-e2e-go-bridge.md` §1:
`pid` must be the real pid; `procStart` must equal `/proc/<pid>/stat` field 22
(parsed after the final `)`, index 19 of the remainder) or discoverers skip the
record as a dead/reused pid; `messagingSocketPath` may live anywhere same-uid;
`name` is what `SendMessage to:` resolves; `peerProtocol:1` + `version:"2.1.232"`
are the compatibility gate; the key file is `<pid>.<sha256hex(socketpath)>.key`,
mode 0600, 71 bytes (the computed name matched the live session's file exactly).
Publish order: bind socket → write key 0600 → write record last via tmp+rename.

**Inbound frames from Claude carry an envelope.** PROVEN-live (captured verbatim in
recv.log) — `gcn50-e2e-go-bridge.md` §2: payload is wrapped in
`<cross-session-message from="uds:…" hop-chain="…" from-name="…" from-mode="…">`,
with extra top-level fields `msgV`, `msg_id` (UUID), `priority:"next"`, `from`.
`SO_PEERCRED` on the accepted connection showed `pid=…  uid=1000` — a receiver can
attribute the sender at kernel level.

**An MCP "channel" server can push a file-borne event into a live session's
context.** PROVEN-live — `gcn50-q12-channels.md` §3 attempt 4, re-proven in
`gcn50-e2e-channels.md` §1–§3. A ~55-line stdlib-python stdio server declaring
`capabilities.experimental["claude/channel"]` and emitting
`notifications/claude/channel` delivered a nonce written by an unrelated shell
process; the server **deleted** the file before the session could read it, so the
channel is the only possible source. Receiver flags:
`--mcp-config <mcp.json> --dangerously-load-development-channels server:<name>`.
Delivery landed while the session was idle at the prompt and the model reacted
immediately.

### 1.2 Full operator loops, closed live

**Loop 1 — relay via a lean `claude -p` SendMessage.** PROVEN-live —
`gcn50-e2e-roundtrip.md` §Phase 2, §OVERALL. Sequence: session spawned in an Orca
worktree → session posted its question into the Slack DM thread → operator (Gaetan)
answered → orchestrator ran the config-G relay carrying the answer → the session
wrote `greeting.txt` = `Hello!` (English, **not** `Bonjour!`) → posted
`[lane:gcn50-e2e] done — wrote greeting.txt in ENGLISH` into the thread.
Causation locked by a pre-relay baseline showing `greeting.txt` ABSENT.
Caveat carried by the probe itself: the operator→session hop was **hand-run**, not
built.

**Loop 2 — relay via a file drop into a watched channel inbox.** PROVEN-live —
`gcn50-e2e-channels.md` §VERDICT, §3. Same shape, French answer, `greeting.txt` =
`Bonjour !` (10 bytes, `cat -A` → `Bonjour !$`). Provenance airtight: the nonce
`E2E-CH-f1da94` existed only in the file the server deleted before the session could
read it. **Honesty note the probe states up front:** this run was autonomous — the
operator role was played by the probing session, not by Gaetan.

**Loop 3 — Go peer.** PROVEN-live for both socket directions and an ask/answer
round-trip between the Go process and a session (`gcn50-e2e-go-bridge.md` §3). It
did **not** exercise a Slack/operator leg.

### 1.3 Session → operator reporting on the Hermes side

**`hermes chat -Q -q '<prompt>' | hermes send -t slack:<chat>:<thread_ts>` produces a
real agent turn, in the right thread, with composed content.** PROVEN-live —
`gcn50-q14-w1-w2-comparison.md` §W2.5–§W2.8. `agent.log` shows
`agent.turn_context: conversation turn` + 2 model API calls + `Turn ended`; the
reply landed as the single reply inside the anchor thread, nothing at channel top
level; it rewrote a manifest into an operator-facing one-liner and obeyed "one line".

**`-Q` is pipe-safe as shipped.** PROVEN-live (measured with `cat -A`, not assumed)
— `gcn50-q14-w1-w2-comparison.md` §W2.2: stdout carries only the answer
(`OK$` = `OK\n`, no ANSI, no banner, no trailing blank line); the session-info line
goes to **stderr**.

**Context can be carried into that turn by inlining a file's full text into the
`-q` argument, transported over ssh stdin.** PROVEN-live —
`gcn50-rehydrate-p1-inline.md` §1. Shape:
`ssh gaetan@192.168.0.9 '~/.local/bin/hermes chat -Q -q "$(cat)"' < full_prompt.txt`.
The reply correctly named a decision buried ~40 lines into a 60-line document
("deterministic PRNG seeded from `worker_id`", "incident replays"), reproduced a
nonce buried in the same section, and pulled the blocker from a third, separate
section — with `tool_turns=0`, so no file/search tool could have been the source.
This is deep read, not top-of-file echo.

**A headless CLI turn can resume a real gateway (Slack-origin) DM-thread session.**
PROVEN-live — `gcn50-rehydrate-p3-resume.md` §1–§3.
`hermes chat --resume <session_id> -Q -q '…'` over a plain non-interactive ssh
(no `-t`, no pty) recalled the correct prior topic from injected history (1 API
call, no tool fishing) and answered in French — the thread's established register —
against an **English** query. Output pipes into `hermes send -t slack:<chat>:<ts>`
unchanged; delivery in-thread confirmed by `slack_read_thread`. Requirements:
**only** that the session id exist as a non-deleted row with history in
`~/.hermes/state.db`; no TTY, no live process (the target had exited hours earlier).
`-Q` still routes banner/resume-status to stderr under `--resume` (INSPECTED:
`cli_agent_setup_mixin.py:368-377`, written specifically to keep the pipe safe).

**Raw thread posting with no LLM.** INSPECTED + PROVEN-live —
`gcn50-q10-kanban-smoke.md` §1 (`hermes send` self-documents "no LLM, no agent
loop"; target form `platform:chat_id:thread_id`) and used live throughout the
probes to post anchors.

**Kanban `notify-subscribe --thread-id` roots delivery in the target thread.**
PROVEN-live — `gcn50-q10-kanban-smoke.md` §4: reply landed inside the anchor thread
4 s after the trigger, nothing at channel top level, and the subscription
auto-removed once the task reached a final status.

**The kanban agent-wake branch does fire when the card carries a `session_id`.**
PROVEN-live (first end-to-end confirmation of a branch previously only read in
source) — `gcn50-q14-w1-w2-comparison.md` §W1.5. `gateway.log`:
`kanban notifier: woke agent for t_bde61b26 …` then a new session ran the turn.
Both the template post and the composed agent reply landed in the anchor thread, so
`--thread-id` propagates through `SessionSource(thread_id=…)` to the agent's own
reply.

**Chat sessions carry the kanban toolset**, including `kanban_create`, which accepts
an initial `blocked` status — PROVEN-live, `gcn50-q14-w1-w2-comparison.md` §W1.2
(tool list quoted verbatim, card `t_bde61b26` created from inside a chat turn with
`session_id` stamped).

### 1.4 Spawn handling that worked

**Spawning the target repo outside `$HOME` removes the external-CLAUDE.md-import
gate entirely.** PROVEN-live, and explicitly flagged as a *validated* fix —
`gcn50-e2e-channels.md` §2. In that run the folder-trust dialog and the
channel-dev warning fired and were accepted; "Allow external CLAUDE.md imports?"
**did not fire**, because the scratch repo lived under `/tmp`, outside `~/CLAUDE.md`'s
inheritance (in `gcn50-e2e-roundtrip.md` the worktree lived under `$HOME` and the
import gate did fire).

**Startup dialogs can be cleared programmatically.** PROVEN-live —
`gcn50-e2e-channels.md` §2 (`orca terminal send --enter`), `gcn50-e2e-go-bridge.md`
§4 (tmux `Enter` on "1. Yes, I trust this folder"), `gcn50-q12-channels.md` §3
attempt 4 (`printf '\r'` into a `script` pty).

**Scrubbing inherited `CLAUDE_CODE_*` env vars makes a spawned session register as
an independent top-level peer.** PROVEN-live — `gcn50-e2e-go-bridge.md` §4 (lists
`CLAUDE_CODE_MESSAGING_SOCKET/_TOKEN`, `CLAUDE_CODE_SESSION_ID`,
`CLAUDE_CODE_CHILD_SESSION`, `CLAUDECODE`, `CLAUDE_CODE_ENTRYPOINT`,
`CLAUDE_CODE_EXECPATH`, `CLAUDE_PID`, `CLAUDE_EFFORT`) and
`gcn50-q12-nonclaude-socket.md` §3, where the *first* attempt inherited
`CLAUDE_CODE_CHILD_SESSION` and did **not** register — documented as a confound.

**Self-clean works.** PROVEN-live — `gcn50-e2e-go-bridge.md` §8: on SIGTERM the Go
peer removed its record, key, socket and rundir; a race where `main` returned before
the cleanup goroutine finished was found and fixed (cleanup moved onto the main
goroutine). An exited Claude TUI's `sessions/<pid>.json` was likewise absent.

---

## 2. WHAT FAILED / DOESN'T WORK

**`claude -p` is not a usable channel receiver on 2.1.232 — 3/3 silent drops.**
PROVEN-live (negative) — `gcn50-q12-channels.md` §3 attempts 2–3, §4.2. The MCP
server connected (`Successfully connected (transport: stdio) in 24ms`), the
notification was written to the transport at drop time, the session stayed alive
past the drop, and the model reported `NO-EVENT` every time. `--debug-file` shows
**no** channel listener being registered and zero hits for
`blocked|org policy|allowlist|channelsEnabled|development-channel|experimental`.
Controls: a bogus flag errors while `--dangerously-load-development-channels`
parses fine, so the flag *is* recognized; no `/etc/claude-code/managed-settings.json`
exists on the box, so no org policy explains it. **Not root-caused** (timeboxed).

**A cross-host file path in a prompt is dead on arrival.** PROVEN-live (negative) —
`gcn50-rehydrate-p1-inline.md` §2. Given a cooper-side absolute path, the VM turn
tried `read_file`/`search_files` against its own VM-local environment
(`tools.file_tools: Creating new local environment for task default`) and correctly
reported the file doesn't exist — no hallucination, but no content either. It burned
5 API calls / ~154,662 tokens / 29.12 s to arrive at that failure. The file *did*
still exist on cooper at run time, so this is genuine cross-host non-existence, not
a cleanup race.

**A CLI-created kanban card can never trigger an agent turn.** PROVEN-live +
INSPECTED — `gcn50-q10-kanban-smoke.md` §5–§6. `hermes kanban create` prints
`session_id=null` and exposes no flag to set it; the notifier's wake block is gated
`if _is_push_adapter and _wake_kinds and _session_key` (`kanban_watchers.py`
L729-772), so the branch is skipped by construction. Observation side: no new
session id, no `agent.turn_context` line, no model API call after the trigger, and
`grep -c "woke agent\|wakeup injection"` = **0 hits on 2026-08-14**. Positive
control from 2026-08-07 (`created_by='worker'` + non-null `session_id`) fired the
wake — same notifier, same DM, single differing input.

**The stamped `session_id` is a non-empty gate, not a live-session handle.**
PROVEN-live — `gcn50-q14-w1-w2-comparison.md` §W1.6. The stamping session was
already dead at trigger time; the gateway opened a **brand-new** session and ran the
turn there, and the stamped id appears nowhere in `agent.log` afterwards.
Consequences the probe draws: a stale or foreign id still fires the wake (no
dead-session failure mode), and the woken agent inherits **none** of the originating
session's conversation state.

**The kanban wake path carries no prompt surface.** PROVEN-live —
`gcn50-q14-w1-w2-comparison.md` §W1.7: the wake payload is the fixed template
`[kanban] Task <id> completed. Title: <title> Assign…`; rehydration from a file was
**NOT** demonstrated for this path. One operator-facing event also produced **two**
Slack messages (raw template post *and* the composed agent turn) — the two
deliveries are not mutually exclusive in practice.

**The operator→session inbound relay was not built and had to be hand-run.**
PROVEN-live (demonstrated gap) — `gcn50-e2e-roundtrip.md` §Step 4. Gaetan replied
"English" in the thread at 17:36:57 — *before* the session's own question posted at
17:39:28. The Hermes gateway woke on that reply and merely **chatted back to the
operator** ("English it is.", 17:37:07); it did not relay the answer inward to the
waiting session. The loop was closed by a human/orchestrator-driven relay instead.

**A message to a receiver without `crossSessionInbound: accept` is HELD, not
delivered.** PROVEN-live — `gcn50-q12-nonclaude-socket.md` §5 variation C. The
receiver logged: "Held peer message from an unidentified session [verified pid …];
… not delivered to Claude (held). The sender did not attest its permission mode and
this session bypasses prompts." Claude Code reads the connecting pid via SO_PEERCRED,
finds it is not an own-child, and — because that receiver bypasses prompts — parks it
for a human to click Deliver. Not silently delivered, not dropped.

**An earlier Go peer attempt failed at the workspace-trust dialog** (the fresh cwd
was under `$HOME`) — recorded as prior art in `gcn50-e2e-go-bridge.md` §VERDICT/§4.

**Operational misfire worth not repeating:** `/usr/bin/time -v command claude …`
tried to exec the shell builtin `command`, rc 127 in 0.00 s, no message sent —
`gcn50-e2e-roundtrip.md` §Step 2. Fixed by timing the real binary path.

---

## 3. MEASURED COSTS

### 3.1 Sending one message into a session

| Mechanism | Tokens | Wall | Source |
|---|---|---|---|
| `claude -p` relay, full orchestrator baseline (config A: `sonnet[1m]`, full CLAUDE.md, all MCP servers, policy appended) | **220,054** total prompt (6 raw + 53,232 cache_creation + 166,816 cache_read) | 17.39 s | `gcn50-q15-lean-sendmessage.md` §Results |
| config B (+ `--strict-mcp-config --mcp-config {}`) | 180,281 | 8.02 s | ibid. |
| config C (+ `--model haiku`) | 159,379 | 13.16 s | ibid. |
| config D (+ scratch cwd, no CLAUDE.md) | 105,345 | 9.13 s | ibid. |
| config E (`--system-prompt` 1-line override) | 80,734 | 9.43 s | ibid. |
| config F (+ `--setting-sources ""`) | 57,980 | 7.38 s | ibid. |
| **config G (+ `--tools "SendMessage"`)** | **3,870** (all raw input; 0 cache) | **4.73 s**, 2 turns | ibid. |
| config H (G + `--effort low`) | 3,887 | 4.92 s | ibid. |
| Same lean relay, as run in the live loop | **6,278** total prompt (output 366), cost $0.00878, `num_turns:3` | **6.83 s** | `gcn50-e2e-roundtrip.md` §Step 2 |
| File drop into a watched channel inbox | ~0 (no model, no tokens) | send is a `printf` | `gcn50-e2e-channels.md` §4 |
| Raw socket write | ~0 tokens, sub-ms | — | `gcn50-e2e-go-bridge.md` §5 |

Reduction config A → G: **−216,184 tokens (−98.2 %)** and **−12.66 s (−72.8 %)** —
`gcn50-q15-lean-sendmessage.md` §Reduction. Biggest single lever `--tools
"SendMessage"` (F→G alone: −54,110 tokens); second `--setting-sources ""` (E→F:
−22,754).

Token-metric caveat the probe insists on: the metric is
`input + cache_creation + cache_read`. Raw `input_tokens` alone is misleading —
heavy configs hit warm caches from other live sessions sharing the prefix, so their
raw input reads as 6–38 while they still occupied 150k+ tokens of cache read.

### 3.2 End-to-end round-trip latency

| Loop | send → file written | send → completion posted | Source |
|---|---|---|---|
| lean `claude -p` SendMessage relay | 12.4 s | **16.5 s** | `gcn50-e2e-roundtrip.md` §Step 3 |
| channel file-drop | **4.20 s** | **8.18 s** | `gcn50-e2e-channels.md` §3, §4 |

### 3.3 Hermes-side turns

| Operation | Cost | Source |
|---|---|---|
| `chat -Q -q` trivial query ("reply with exactly OK") | ≈ 9 s | `gcn50-q14-w1-w2-comparison.md` §W2.2 |
| `chat -Q -q \| send -t` composed status into a thread | **14.56 s** end-to-end (12:34:08.536→12:34:23.093Z), 2 API calls; ~4 s of it a speculative `kanban_show` error | `gcn50-q14-w1-w2-comparison.md` §W2.5–§W2.8 |
| kanban wake: trigger → template post | 1.8 s | ibid. §W1.8 |
| kanban wake: trigger → composed agent reply | 10.0 s (gateway self-reported `time=7.7s api_calls=2`) | ibid. |
| kanban wake **including** its mandatory card-creation turn | ~37 s (card creation alone 26 s) | ibid. |
| kanban `notify-subscribe --thread-id` raw post | 4 s after trigger | `gcn50-q10-kanban-smoke.md` §4 |
| inlined-file prompt (2,298-byte file, 3,119-byte prompt) | **1 API call, in=31,332 out=55, total=31,387**, model latency 4.3 s, wall 9.26 s | `gcn50-rehydrate-p1-inline.md` §1 |
| same question by cross-host **path** (fails) | 5 API calls, ~154,662 tokens summed, 29.12 s | ibid. §2 |
| `chat --resume` recall turn | 1 API call, ~8.5 s wall | `gcn50-rehydrate-p3-resume.md` §3 |
| cold (`no --resume`) equivalent — **confounded**, see §6 | 3 API calls, ~18.6 s | ibid. §4 |

**Fixed floor on a Hermes CLI turn: ~30.5–30.6k input tokens** (SOUL.md + tool
schemas + system prompt), measured because an effectively-empty query still cost
`in=30590` on its first call — `gcn50-rehydrate-p1-inline.md` §1. A 2.3 KB inlined
file added only ~740–800 marginal tokens. Cache hit rate on later calls within one
CLI session: 97–98 %.

### 3.4 Message and process counts

- Slack messages per operator-facing event: **1** via the compose-and-post pipe;
  **2** via the kanban wake (raw template + composed reply) —
  `gcn50-q14-w1-w2-comparison.md` §W1.7, §Comparison.
- Persistent processes: the channel path keeps **one filewatch MCP subprocess alive
  for the whole session**; the `claude -p` relay keeps **nothing** — spun on demand,
  exits — `gcn50-e2e-channels.md` §4.
- Peer roster measured on cooper during one probe: **12** live peer sessions
  (`gcn50-q12-headless-sendmessage.md` §3); during the crux-A review: **16**
  registrations / 16 live pids, but **59** sockets in `/run/user/1000/cc-socks/`
  (`gcn50-review-fable-tars-to-session.md` §1).

---

## 4. CONSTRAINTS & FRAGILITY

### 4.1 Version-shaped behavior

- **Everything measured is Claude Code 2.1.232.** Stated in the headers of
  `gcn50-q12-channels.md`, `gcn50-q12-nonclaude-socket.md`,
  `gcn50-e2e-go-bridge.md`, `gcn50-e2e-channels.md`.
- **The lean flag stack (`--tools`, `--setting-sources ""`, `--strict-mcp-config`)
  is newer surface and must be re-validated per Claude upgrade** — judgement in
  `gcn50-review-fable-tars-to-session.md` §2 (table row 1) and §6.6, which proposes
  re-running config G as a 30-second post-upgrade smoke.
- **The socket frame is not a contract.** It was harvested from the claude binary's
  own debug log, which literally embeds the injection recipe
  (`[uds-messaging] Inject messages (auth line … optional here): { echo '{"type":"auth"…}'; echo '{"type":"user"…}'; } | socat - UNIX-CONNECT:<sock>`)
  — `gcn50-q12-nonclaude-socket.md` §1. The review scores it "worst" on version
  robustness: "no compatibility promise at all"
  (`gcn50-review-fable-tars-to-session.md` §2).
- **The Go peer contract is undocumented and version-pinned, and fails silently.**
  `gcn50-e2e-go-bridge.md` §6 flags this as the top risk: any upgrade can break
  `peerProtocol` (currently `1`, "the single most likely gate to bump"), the record
  schema, the key-filename hashing input, the `procStart` liveness check, the socket
  framing, or the `cc-socks`/`sessions` paths. Failure mode: the peer simply stops
  appearing in `ListAgents` or its sends stop landing — **no error, no log**.
- `ob.authRequired` is **false on this build** — the auth line is optional here, and
  that is a build property, not a guarantee (`gcn50-q12-nonclaude-socket.md` §1).

### 4.2 Same-uid / same-box limits

- `~/.claude/sessions/` and `/run/user/1000/cc-socks/` are `drwx------`; session
  records are `0664`, key files `0600`, sockets `srw-------`. **Reaching a socket
  requires the same OS user.** INSPECTED — `gcn50-q12-nonclaude-socket.md` §2.
- Therefore the socket/Go transports are **same-box only**. The Go probe spells out
  the consequence: Tars is a *different machine*, so this works for a process
  co-located with the sessions, "**not** for the Tars VM reaching cooper";
  cross-machine needs an ssh hop that lands as uid 1000 —
  `gcn50-e2e-go-bridge.md` §5.1.
- `ssh cooper '<cmd>'` **does** run as uid 1000 and therefore qualifies —
  `gcn50-q12-nonclaude-socket.md` §6.
- The messaging env vars visible in a session's `/proc/environ` are the *inherited
  parent* values; a session binds its own socket at its own pid and writes its own
  key — `gcn50-q12-nonclaude-socket.md` §2.

### 4.3 Interactive-only receivers and channel preview gating

- **Channel events reach interactive sessions only** (`-p` drops them silently, 3/3)
  — `gcn50-q12-channels.md` §VERDICT, §4.2; restated as a receiver constraint in
  `gcn50-e2e-channels.md` §4.
- Channels are a **research preview** requiring Anthropic auth via claude.ai or a
  Console API key, "not available on Amazon Bedrock, Google Cloud's Agent Platform,
  or Microsoft Foundry"; `--channels` accepts only an Anthropic-maintained plugin
  allowlist during preview; **neither flag appears in `claude --help`** — quoted
  docs, `gcn50-q12-channels.md` §1.
- The dev flag prints a **confirmation dialog that must be accepted** before the
  session starts — `gcn50-q12-channels.md` §3 attempt 4; the crux-A review calls
  that dialog alone "a blocker for unattended spawns"
  (`gcn50-review-fable-tars-to-session.md` §2, row 3).
- Cosmetic-but-persistent oddity: even in the working runs the banner also prints
  `server:filewatch · no MCP server configured with that name`, likely because the
  server came from `--mcp-config` rather than a `.mcp.json`. Delivery happened
  anyway; the probe says retest with a real `.mcp.json` before trusting the warning
  either way — `gcn50-q12-channels.md` §4.3, echoed in `gcn50-e2e-channels.md` §2.
- Docs (quoted, `gcn50-q12-channels.md` §1.4): Claude Code **does not acknowledge**
  notifications; if the session hasn't loaded the server as a channel or policy
  blocks it, events are dropped **silently with no error to the server**. Queued
  events are delivered together on the next turn and handled as a group.
- **The inbox directory is a raw prompt-injection surface** — whatever is dropped
  there lands verbatim in the model's context inside a `<channel>` tag, so the drop
  dir must be access-controlled and senders gated — `gcn50-q12-channels.md` §4.4,
  `gcn50-e2e-channels.md` §4.

### 4.4 Startup trust gates

- Gates observed at spawn: **folder-trust** ("Is this a project you trust?"), the
  **channel-dev warning**, and — only when the working directory sits under `$HOME`
  and inherits `~/CLAUDE.md`'s external `@imports` — **"Allow external CLAUDE.md
  imports?"** — `gcn50-e2e-channels.md` §2 vs `gcn50-e2e-roundtrip.md` (wire-1
  finding referenced there).
- **A seed prompt can be dropped while a dialog is up.** The channel run therefore
  spawned bare (no prompt), cleared the gates, then sent the brief pointer
  separately — `gcn50-e2e-channels.md` §2.4.
- The trust dialog fires **before** the session publishes its registration record —
  `gcn50-e2e-go-bridge.md` §4. That is what killed a prior Go attempt.

### 4.5 Addressing and delivery semantics

- `success:true` means **enqueued at a live peer**, not delivered. The
  Delivered/Held/Refused outcome is reported via an **async sender-side notice** —
  and a throwaway relay is dead seconds later, so that native ack channel is
  structurally discarded. `gcn50-review-fable-tars-to-session.md` §1.4, §3.
- **Claude Code drops identical repeats from the same sender within a short window**
  (documented rate-limit/loop protection). A blind retry of the same text can be
  silently swallowed; a per-message nonce defeats it —
  `gcn50-review-fable-tars-to-session.md` §1.5. Doc-sourced, **unprobed**.
- **cwd is not a usable address.** Four live sessions were concurrently registered
  with cwd = the same worktree, all `nameSource:"derived"` —
  `gcn50-review-fable-tars-to-session.md` §1, §4.3.
- **A derived name is not stable across restarts**: a session relaunched in the same
  worktree derives a new suffix, invalidating any recorded name — ibid. §4.2. And if
  a live session already answers to a chosen name, Claude Code silently renames the
  newcomer to a variant — ibid. §4 (docs).
- **The socket directory accumulates stale entries** (59 sockets for 16 live
  processes; an earlier reading saw 51 vs ~13). Registry hygiene after a *crash*
  rather than a clean exit is unproven — ibid. §1.
- **A fleet coordinator handoff mints a brand-new peer identity** (new auto-assigned
  name and ref), silently invalidating any recorded name —
  `gcn50-megaultracode-compat.md` §Wire 3.
- Receiver posture is the flipping variable: `accept` → silent delivery;
  bypass-perms without `accept` → held for a human; an ordinary prompting session
  without `accept` → delivered per docs, **not separately tested** —
  `gcn50-q12-nonclaude-socket.md` §5–§6.

### 4.6 Slack / gateway reading behavior

- **DM thread sessions start with an empty transcript by explicit design** —
  session-level parent-transcript seeding was removed because it copied the entire
  parent DM transcript and bled unrelated conversations across threads. INSPECTED
  from source + `test_session_dm_thread_seeding.py` docstring —
  `gcn50-rehydrate-p2-threadindex.md` §2.
- **Thread context is injected once**: on a thread-reply event where
  `_has_active_session_for_thread()` is false, the adapter calls
  `_fetch_thread_context` → `conversations.replies(limit=31, inclusive=True)` and
  folds it into that one turn, then marks the thread checked so it never fires again
  for that thread+process. INSPECTED (`adapter.py:5793-5829`) and corroborated by a
  one-time `history=` jump across 9 multi-turn sessions in `agent.log` — ibid.
- **Two guarded re-fetches only**: an explicit @mention on an already-active thread
  forces a watermark-delta fetch; the first ordinary reply after a gateway process
  restart does one delta fetch, at most one `conversations.replies` call per thread
  per process lifetime. INSPECTED (`adapter.py:~5836-5905`) — ibid.
- **Consequence:** once a lane thread's gateway session is warm, further bot-authored
  posts into that thread are invisible to the turn that later answers there; threads
  longer than 31 messages are partially invisible even cold —
  `gcn50-review-fable-session-to-tars.md` §C-3.
- **A top-level DM message gets a synthetic `thread_ts = ts` and starts a new
  session with no lane binding** — so a natural top-level reply lands in the wrong
  session. `gcn50-review-fable-session-to-tars.md` §B-4 (citing `docs/facts.md`).
- **`hermes send --json` to a threaded target returns no `"mirrored"` key**, while a
  top-level send returns `"mirrored": true`. PROVEN-live —
  `gcn50-q14-w1-w2-comparison.md` §W2.8. Whether that means threaded posts never
  enter any Hermes transcript is **unresolved and load-bearing** —
  `gcn50-review-fable-session-to-tars.md` §0, §B-4.
- **No `thread_ts → anything` lookup verb exists in the `hermes` CLI** — checked
  across `--help` for `chat`, `gateway`, `debug`, `kanban`, `kanban list`,
  `kanban notify-list`. `--resume` takes a session id, `--continue` a name;
  `kanban list --session` filters by originating agent session id. INSPECTED —
  `gcn50-rehydrate-p2-threadindex.md` §1.
- The Hermes session store *is* structurally thread-keyed —
  `agent:main:slack:dm:[<team_scope>:]<chat_id>:<thread_ts>`, 103 entries, key
  construction confirmed in `gateway/session.py::build_session_key` — but a session
  is a conversation handle, carrying no worktree path, ticket id, or file pointer.
  INSPECTED — ibid.
- **Two key shapes coexist for the same `thread_ts`** (with and without the team
  scope segment), consistent with the 2026-08-13 DM-cutover config change; session
  continuity for pre-change threads would be broken by key-format drift alone.
  INFERRED, not probed — ibid. The closing review turns this into "don't resume into
  pre-2026-08-13 threads" (`gcn50-rehydrate-review.md` §5.6).
- `kanban.db` has a `kanban_notify_subs` table with a plain `thread_id` column, so
  `SELECT task_id … WHERE thread_id=?` is a real reverse index **today** — but only
  for kanban-bound cards, and it only fires on a kanban-notifier wake, not on an
  ordinary thread reply. INSPECTED (read-only schema query) — ibid. §1.

### 4.7 Resume side effects

- `--resume` **mutates the target row**: `_preload_resumed_session`
  (`cli_agent_setup_mixin.py:576-655`) does `UPDATE sessions SET ended_at=NULL,
  end_reason=NULL WHERE id=?` and then appends the new turns. INSPECTED + observed
  live — `gcn50-rehydrate-p3-resume.md` §5.
- Resume is **source-agnostic by construction** — nothing in the resume path checks
  whether the session originated from `cli` or `slack`, which is what allowed a
  gateway-origin session to be targeted at all. INSPECTED — ibid.
- Concurrent `--resume` against a live gateway turn on the same row: no locking
  observed in the read code path. Flagged INFERRED/unmeasured by the probe (ibid.),
  and accepted in `gcn50-rehydrate-review.md` §5.4 with the reasoning that sqlite
  serializes writes so the credible failure is transcript interleaving, not
  corruption.

### 4.8 Document-size and context limits

- The inline-carry result is **n=1 at 2,298 bytes**. The "a 10× larger file would add
  ~7–8k tokens" figure is **arithmetic extrapolation, not measurement**, and the real
  ceiling is the model's context window, which was not probed —
  `gcn50-rehydrate-p1-inline.md` §1/§Verdict, confirmed by the audit in
  `gcn50-rehydrate-review.md` §0 and listed as accepted risk §5.5.
- The gateway's 31-message thread-hydrate limit is a **different subsystem** from the
  CLI `chat -q` argument path. That they are unrelated is **INFERRED** (from the
  code-path description plus zero `gateway.log` activity during the probe, line
  count 2841→2841), not separately verified — `gcn50-rehydrate-p1-inline.md`
  §Verdict.

### 4.9 Fleet-shaped work (skill-composition review)

`gcn50-megaultracode-compat.md`, an inspection-based review, not a live run:

- `/megaultracode-orca` is a project-scoped **skill** with
  `disable-model-invocation: true`, invoked *inside* an already-running session which
  then becomes a coordinator — it is **not a launcher** and cannot spawn the first
  session (§0).
- Coordinator and workers receive the *same* appended orchestration policy, and the
  skill's brief arrives as a **user message**, not a system prompt, so it coexists
  rather than overwriting — verified from the launcher scripts (§Wire 1).
- **Worker session names are auto-assigned**, not deterministic (§Wire 1).
- Every orc-launched session (coordinator and workers) carries
  `crossSessionInbound:"accept"`; a coordinator started as a *bare* interactive
  session would not (§Wire 3).
- Side-blocker: the in-repo copy `.claude/skills/megaultracode-orca/SKILL.md` is
  **331 lines vs 399 canonical**, missing the merge-token protocol and later incident
  entries — a coordinator spawned in a Tars worktree loads the stale one (§Wire 1,
  §Secondary).
- No measured startup-context figure for a fleet worker exists (digest: NOT FOUND) —
  §Untested residue.

---

## 5. OPERATIONAL FINDINGS

**Orca CLI**

- **There is no `repo rm` / `repo remove` verb** in this Orca build:
  `orca repo <command> --help` lists only list/add/show/set-base-ref/search-refs, and
  `orca repo rm` errors "Unknown command". A registration pointing at a deleted path
  is left dangling — harmless, since no worktree can be created under a missing
  directory. `gcn50-e2e-cleanup.md` §Not removed; same finding independently in
  `gcn50-e2e-channels.md` §5.
- `orca terminal create` **rejects an unregistered path** with `selector_not_found`;
  `orca repo add --path <repo>` must come first — `gcn50-e2e-channels.md` §2.2.
- `orca terminal close` may need **two calls**: the first SIGHUPs the underlying
  claude (`ptyKilled: false`), the second finishes the job (`ptyKilled: true`) —
  `gcn50-e2e-cleanup.md` §Removed.
- `orca worktree rm --force --json` returns `{"removed": true}` and moves the
  directory into Orca's own `.orca-worktree-trash` — ibid.
- Verification of teardown relied on `orca terminal list` (0 terminals) and
  `orca worktree list` (worktree absent) because no `ListAgents`-equivalent tool was
  available in that session — ibid. §Unrelated infra.

**Launcher / environment facts on cooper**

- `orc-sonnet` (read in full) is
  `claude --model 'sonnet[1m]' --settings '{"ultracode":true,"crossSessionInbound":"accept"}'
  --append-system-prompt-file <policy> --dangerously-skip-permissions "$@"`, and its
  own header says never to pipe or redirect it —
  `gcn50-q12-headless-sendmessage.md` §1.
- The `claude()` shell function used for interactive Orca tabs sets `accept`, so
  every orc-* and interactive-Orca session on cooper is inbound-accepting —
  `gcn50-q12-nonclaude-socket.md` §6.
- **No `/etc/claude-code/managed-settings.json`** on the box ⇒ no org-policy block
  in play for any of these measurements — `gcn50-q12-channels.md` §3.
- **`ANTHROPIC_API_KEY` is unset**; this account authenticates via OAuth/subscription,
  so `--bare` (which accepts only `ANTHROPIC_API_KEY`/`apiKeyHelper`) would break
  auth and was not tested — `gcn50-q15-lean-sendmessage.md` §Results note.
- `--setting-sources ""` strips user/project/local settings **including this repo's
  SubagentStart hook injection** — fine for a throwaway relay, not for a session
  meant to do real work under repo rules — ibid. §Caveats.
- `--output-format json` surfaces only the model's summary string, not the raw tool
  envelope, so a machine `msg_id` is not exposed by the lean config —
  `gcn50-e2e-roundtrip.md` §Step 2.
- Haiku reasoning was not the bottleneck (100 % delivery, 8/8) and `--effort low`
  bought nothing: wall time is dominated by process/model cold-start. Before `--tools`
  was narrowed, haiku was *slower* than sonnet (4 turns vs 3) because it "thought"
  more with the full tool surface exposed; that inversion disappeared once `--tools`
  was restricted — `gcn50-q15-lean-sendmessage.md` §Caveats.

**Hermes VM facts**

- A `hermes chat` CLI session **does not go through the gateway at all** —
  `gateway.log` had zero lines in the window while `agent.log` showed the full turn —
  `gcn50-q14-w1-w2-comparison.md` §W2.6.
- CLI chat sessions and gateway sessions are different species living in the same
  `state.db`; the CLI session carries no conversation history with Gaetan and its
  voice differs (English vs the DM's French) — `gcn50-q14-w1-w2-comparison.md`
  §W1.7/§Comparison/§Practical notes.
- Chat turns **speculatively call `kanban_show`** and log a `WARNING … returned
  error` before answering — a wasted round trip costing ~4 s of a 14.5 s turn, seen
  again in the resume probe's cold control — `gcn50-q14-w1-w2-comparison.md` §W2.8,
  `gcn50-rehydrate-p3-resume.md` §4.
- `agent.log`'s `msg=` field is truncated (~80–100 chars) by the application's own
  log formatter, not by display: measured raw on-disk line = 283 bytes vs the visible
  slice. The metadata fields (`session=`, `model=`, `history=`) are **not** truncated
  — `gcn50-rehydrate-p2-threadindex.md` §2.
- **There is no `sqlite3` binary on the VM**; DB reads were done with `python3` and a
  `mode=ro` URI — `gcn50-rehydrate-p3-resume.md` §2.
- `~/.hermes/state.db` is the canonical session store; the JSON under
  `~/.hermes/sessions/` is a legacy mirror — INSPECTED, ibid.
- A bare `hermes chat` turn is **not context-free**: it has a `session_search` tool
  over `state.db` and will self-serve context if the prompt leaks a searchable term.
  This confounded the resume probe's cold control (the query literally contained
  "gcn50-rehydrate") — `gcn50-rehydrate-p3-resume.md` §4.
- SOUL.md's language rule loads into bare `hermes chat` too (auto-injected, not
  gateway-specific), so a cold CLI turn is not missing SOUL.md — what resume adds is
  the *thread's* established register winning over language-mirroring. INFERRED,
  single sample — ibid. §3(b).
- Kanban CLI hygiene used across probes: `--initial-status blocked` so the dispatcher
  never spawns a worker on a probe card; `archive` after; subscriptions auto-clear on
  final status — `gcn50-q10-kanban-smoke.md` §3/§8,
  `gcn50-q14-w1-w2-comparison.md` §Cleanup.
- No probe restarted the gateway or edited `config.yaml`/`.env`, and none invoked
  `sops` — `ActiveEnterTimestamp` identical at start and end of each run
  (`gcn50-q10-kanban-smoke.md` §2/§8, `gcn50-q14-w1-w2-comparison.md` §0/§Cleanup,
  `gcn50-rehydrate-p1-inline.md` header, `gcn50-rehydrate-p3-resume.md` §5).

**Probe hygiene that mattered to the results**

- Causation in every live loop was locked by a **pre-action baseline** showing the
  target artifact absent, plus a nonce that existed nowhere the receiver could read
  from disk — `gcn50-e2e-roundtrip.md` §Pre-relay baseline,
  `gcn50-e2e-channels.md` §VERDICT, `gcn50-q12-channels.md` §2.
- Delivery was never accepted on the CLI's own word; it was verified receiver-side —
  `gcn50-q15-lean-sendmessage.md` §Method.
- Sequential windows with no overlap were used to keep attribution clean when two
  mechanisms were compared on one gateway — `gcn50-q14-w1-w2-comparison.md` header.
- One probe ran while Gaetan was actively using the DM; the confound was identified
  and isolated by enumerating every pre-existing session id —
  `gcn50-q10-kanban-smoke.md` §2, §5.

---

## 6. WHAT THE EVIDENCE DOES NOT ESTABLISH

Flagged by the probes themselves. Listed so nobody mistakes silence for coverage.

**Never measured at all**

- SendMessage to a **name that doesn't exist**, to a **duplicate name**, to a
  **busy** receiver, or to a **held** receiver — all four unprobed; a `-p` relay
  cannot answer the interactive "which one do you mean?" disambiguation —
  `gcn50-review-fable-tars-to-session.md` §1, §7.
- The documented **identical-repeat drop** on retries — doc-sourced, "nobody probed
  this" — ibid. §1.5.
- **Registry hygiene after a crash** (as opposed to a clean exit) — ibid. §1.
- **SendMessage under default permission mode from a headless process** — only
  demonstrated under `--dangerously-skip-permissions`; a `-p` run without the flag
  would have no TTY to answer a prompt — `gcn50-q12-headless-sendmessage.md` §4, §5.3.
- **`--tools ""`** (known-negative control) and **`--bare`** (would break OAuth auth)
  — `gcn50-q15-lean-sendmessage.md` §Results note.
- **The inbound half live** — a real operator threaded reply waking the right session,
  resolving which delegated lane it belongs to, and relaying onward — has **zero
  probe coverage**. Every component is individually proven; the composition is not.
  `gcn50-review-fable-session-to-tars.md` §B-4;
  `gcn50-rehydrate-review.md` §5.1.
- **The CLI→gateway row-identity round trip**: after a CLI `--resume` appends turns,
  whether the *next gateway wake* on that thread sees them. Gateway→CLI sharing is
  proven; the return direction is INFERRED — `gcn50-rehydrate-review.md` §5.2.
- **Concurrent `--resume` vs a live gateway turn** on one `state.db` row —
  `gcn50-rehydrate-p3-resume.md` §5; accepted-risk, not measured.
- **`--continue` by name** was not exercised, and there is no tested "resume the DM
  with Gaetan" shorthand — the session id must be looked up from `state.db` —
  `gcn50-rehydrate-p3-resume.md` §6.
- **Whether threaded sends enter any Hermes transcript** — the missing `"mirrored"`
  key is unresolved and load-bearing — `gcn50-review-fable-session-to-tars.md` §0, §B-4.
- **Startup context cost of a fleet worker** — digest NOT FOUND,
  `gcn50-megaultracode-compat.md` §Untested residue.
- A **kanban wake into a thread-scoped subscription** was untested at the time of
  `gcn50-q10-kanban-smoke.md` §6; `gcn50-q14-w1-w2-comparison.md` §W1.5 later ran it
  — noting the sequence so the earlier "untested" is not read as still open.

**Not root-caused**

- Why `claude -p` silently drops channel events (timeboxed; no org-policy or
  allowlist evidence found) — `gcn50-q12-channels.md` §4.2.
- Why the interactive banner prints `no MCP server configured with that name` while
  delivery works — hypothesis only (`--mcp-config` vs `.mcp.json`), retest advised —
  ibid. §4.3.

**Measured under easier conditions than production**

- Every delivery measurement used an **idle, `accept`-configured, fixed-name
  receiver on one box at one Claude version**; production sessions have derived
  names, are busy for long stretches, and are sometimes dead —
  `gcn50-review-fable-tars-to-session.md` §1 ("OVERCLAIM" row).
- The composed-status latency (14.56 s) was a **trivial "compose one line from a
  local file"** task including a wasted tool call; it is a floor, not a design
  number — `gcn50-review-fable-session-to-tars.md` §0, §Consolidated #7.
- The kanban-file rehydration nonce was proven for a **VM-local** file whose absolute
  path was handed to the model — that is not the same as a cross-host read, which the
  later probe showed fails outright — `gcn50-review-fable-session-to-tars.md` §0,
  `gcn50-rehydrate-p1-inline.md` §2.
- The resume probe's **cold control is confounded** by `session_search`, so its
  "2× faster / more reliable" comparison is directional, not clean —
  `gcn50-rehydrate-p3-resume.md` §4, audited in `gcn50-rehydrate-review.md` §0.
- The persona-carryover effect is a **single sample**, mechanism labelled INFERRED,
  not re-tested with a French query to separate the two effects —
  `gcn50-rehydrate-p3-resume.md` §3(b).
- One of the three live loops was **autonomous**: the operator role was played by the
  probing session, not by Gaetan — `gcn50-e2e-channels.md` header.
- In the other, the operator's answer arrived **before** the question was posted,
  so the loop did not test a genuine wait-then-answer ordering —
  `gcn50-e2e-roundtrip.md` §Step 4.
- Per-configuration timings in the lean-relay table are **n=1** — 4.73 s is a single
  wall-clock sample with no variance; not an SLO —
  `gcn50-review-fable-tars-to-session.md` §1.

**Not knowable from a probe at all**

- Any claim that rests on *discipline* rather than mechanism — write-back rules,
  single-writer conventions, prefix-following in output routing — has n=0 and cannot
  be proven by a probe. Stated explicitly in `gcn50-rehydrate-review.md` §1.4, §5.3.
- "All context" is not literal in any shape measured: what rehydrates is *what was
  written*. Prior tool results, in-flight nuance, and thread text beyond the
  31-message cold window are lost by construction —
  `gcn50-review-fable-session-to-tars.md` §C-7, `gcn50-rehydrate-review.md` §1.6.

**Outside this evidence base**

- Slack read gotchas such as `include_bots`, and any session-limit/pacing behavior,
  appear **nowhere in the 17 GCN-50 probe files**. Whatever is known about them comes
  from other files in this repo and is deliberately not carried into this ledger,
  which cites only sources read for it.
