# GCN-59 — busy-receiver & steer semantics, probed live

**Date:** 2026-08-16 (UTC) · **Host:** cooper · **Claude Code: 2.1.233** (ledger's
measurements were on 2.1.232 — one point of version drift, noted, not re-validated
beyond what's below). n=1 per clean scenario unless stated; results were internally
consistent (no retry needed to resolve raciness).

## Verdict table

| Transport | Busy-delivery semantics | Unattended-spawn viability | Pluggability |
|---|---|---|---|
| **SendMessage relay** (`claude -p … --tools "SendMessage"`, config G) into an interactive receiver | **QUEUED to end of current turn** — arrives on the wire mid-tool-call, model acts on it only after the in-flight tool call returns. PROVEN-live. Not a mid-tool-call interrupt; not dropped. | Folder-trust dialog clears via a programmatic `tmux`/pty `Enter` (`orca terminal send --enter` equivalently, per ledger §1.4). PROVEN-live. | N/A — this is Claude Code's own peer-messaging wire, not MCP; the question doesn't apply. |
| **MCP `claude/channel` file-drop** (`--dangerously-load-development-channels`) into an interactive receiver | **QUEUED to end of current turn**, same shape as SendMessage — delivered to the transport immediately, but the model only reads/acts on the `<channel>` content once the blocking tool call ends. PROVEN-live. **Caveat:** whether the model *acts* on it is a judgment call, not guaranteed delivery-side compliance (see §3). | Extra channel-dev confirmation dialog fires on top of folder-trust; **also clears via a plain programmatic `tmux Enter`** — same mechanism, no special handling needed. PROVEN-live. | **Claude-Code-specific.** No evidence any other harness (OpenAI Codex CLI, Cursor CLI, Aider, Cline, Continue, Goose, Gemini CLI) implements `claude/channel` or an equivalent unsolicited push-into-live-session mechanism. MCP's spec defines only generic, lower-level primitives (`experimental` capability slot, `logging`/`notifications/message`) a client *could* build such a feature on — it does not standardize the "inject text into a running session's context" behavior itself. Desk research, see §4. |

**Headline answer to the ticket question:** for both transports, against a genuinely
busy (foreground-blocking) receiver, the message is **queued to the end of the
current turn** — never an in-flight interrupt, never dropped. The first live run
(SendMessage into a receiver "busy" with `sleep 90`) looked like an immediate
interrupt, but that was a methodological artifact: Claude Code 2.1.233
auto-backgrounds a bare `sleep` and ends the model's turn immediately after
backgrounding it, so the receiver was actually idle-with-a-background-job, not
mid-turn, when the steer arrived. Repeating with a genuine non-backgroundable
foreground CPU-bound Bash call (a Python busy-loop) on both transports cleanly
reproduced queued-not-interrupt in both cases — see §2 and §3 for the raw
timestamps and transcript excerpts.

---

## 1. Method notes / deviations from the brief

- Both receivers spawned interactively under tmux, scratch cwd under `/tmp`
  (`/tmp/gcn59-busy-probe/rx1`, `/tmp/gcn59-busy-probe/rx2`), no CLAUDE.md in
  either tree — no external-CLAUDE.md-import gate fired for either, consistent
  with the ledger's §1.4/§4.4 finding (spawn outside `$HOME`).
- Fixed receiver names: `gcn59-busy-rx` (SendMessage transport) and
  `gcn59-busy-rx2` (channel transport), both `--settings '{"crossSessionInbound":"accept"}' --dangerously-skip-permissions`.
  Startup dialogs cleared with tmux `Enter` keystrokes.
- **Confound found and corrected:** the brief's suggested busy task (`sleep 90`
  via Bash) does **not** produce a genuinely busy/blocked receiver on this
  Claude Code build — the harness detects bare `sleep` and auto-backgrounds it
  ("Foreground sleep is blocked. Running in background instead."), which ends
  the model's turn immediately. A receiver that just backgrounded a sleep and
  reported "I'll confirm once it completes" is, from the scheduler's point of
  view, idle at a fresh prompt — not busy. Switched the busy task to a
  synchronous, non-backgroundable CPU loop run explicitly in the foreground:
  `python3 -c "import time; end=time.time()+90; x=0; exec('while time.time() < end: x += 1'); print('busy-loop-done', x)"`,
  instructed as "do NOT background it, wait for it to finish in the
  foreground." Confirmed genuinely foreground both times (pane showed "Ran 1
  shell command" only after ~90s, no backgrounding language).
- Relay sender: config G verbatim from the ledger (`gcn50-q15-lean-sendmessage.md`
  §Recommended), run from `/tmp/gcn59-busy-probe/relay-cwd` (no CLAUDE.md), with
  `--strict-mcp-config --mcp-config empty-mcp.json`.
- Channel server: byte-for-byte copy of the ~55-line stdlib python server from
  `status/probes/gcn50-q12-channels.md` §2 (inbox-watch, push-then-delete). Self-check
  (bare stdin/stdout handshake + nonce push) passed before spawning the receiver.
- Causation: every steer nonce was freshly generated (`openssl rand -hex 4`)
  per run, never reused, and file timestamps were read with
  `--time-style=full-iso` (nanosecond precision) directly off the filesystem,
  not from CLI/model self-report.

## 2. Transport 1 — SendMessage relay

### Run A (confounded — kept as a secondary data point, see §1)
Receiver told to `sleep 90` then write `busy-done.txt`. Harness auto-backgrounded
the sleep at spawn; the model's turn ended right after launching it
("Running in background... I'll confirm once it completes"). Config-G relay sent
at **08:28:02.601** UTC (nonce `7edb6382`). Result:
`steer-7edb6382.txt` mtime **08:28:02.601307196** UTC (same second as send —
essentially instant), `busy-done.txt` mtime **08:29:11.653790793** UTC (69s
later, when the backgrounded sleep actually finished). **Reads as an
"interrupt," but the receiver's turn had already ended before the steer
arrived** — this is "queued into an idle prompt with a background job still
running," not "interrupted a live tool call." Kept for completeness; do not
read this row as proof of true mid-turn interrupt.

### Run B (clean — genuine foreground blocking busy task)
Receiver told to run the CPU busy-loop **without** backgrounding it, then write
`busy-done2.txt`. Confirmed via pane text: `● Running a 90-second Python busy
loop in foreground` → `Ran 1 shell command` only ~90s later, no
"backgrounded"/"in background" language this time. Config-G relay sent at
**08:30:11** UTC (nonce `6fe92c38`). Result (nanosecond mtimes,
`ls -la --time-style=full-iso`):

| file | mtime (UTC) |
|---|---|
| `busy-done2.txt` (busy task's own completion) | 08:31:32.734766841 |
| `steer2-6fe92c38.txt` (steer target) | 08:31:33.354771100 |

**Steer file landed 0.62s AFTER the busy-task file** — i.e. after the blocking
Bash call returned and the turn resumed. Pane transcript confirms the ordering
explicitly: the inbound `STEER TEST 6fe92c38: …` message is visible in the
transcript (arrived via the wire) directly under the "Ran 1 shell command" line,
but the model's response text — `Busy loop finished. Writing the user's required
file first, then the peer's steer file.` — only appears after the tool call
completed, and it wrote `busy-done2.txt` before `steer2-*.txt`, exactly matching
the mtime order. **Verdict: queued to end of current (blocking) turn, not a
mid-call interrupt, not dropped.** PROVEN-live, n=1 (clean run; the confounded
Run A is a second, non-comparable data point, not a repeat of the same
condition).

## 3. Transport 2 — MCP `claude/channel` file-drop

Receiver `gcn59-busy-rx2` spawned with
`--mcp-config …/rx2/mcp.json --dangerously-load-development-channels server:filewatch`.
**Two startup dialogs** fired in sequence: folder-trust, then the channel-dev
warning ("--dangerously-load-development-channels is for local channel
development only… ❯ 1. I am using this for local development"). **Both cleared
by the identical mechanism** — a plain `tmux send-keys Enter` — no special
handling needed for the extra dialog; it behaves exactly like the folder-trust
one for unattended clearing. Banner confirmed channel registration: `Channels
(experimental) messages from server:filewatch inject directly in this session`.

Same busy protocol as Transport 1 Run B (foreground CPU loop, no backgrounding).

### First drop (`answer.txt`, nonce `a2392ea4`) — delivery proven, compliance withheld
Dropped into `inbox/` at **08:33:10.966** UTC, ~26s into the 90s busy loop.
`server.log` shows the `notifications/claude/channel` push emitted immediately
(sub-second, per the server's 0.5s poll loop) — the file was gone from `inbox/`
within that window, proving the channel, not a stray read, was the source.
Busy loop returned and `busy-done3.txt` was written at **08:34:17.135888188**
UTC. `steer3-a2392ea4.txt` was **never written** — the receiver's own
transcript explains why: the busy-task prompt for this run included an
instruction not to touch `inbox/`/`server.py` etc. (meant only to stop it
peeking at the mechanism), and the model generalized that into treating the
channel push itself as untrusted: *"a filewatch channel message tried to steer
me into writing a steer3-*.txt file mid-task. Ignored it — it's untrusted
external input, not from you, and not part of this task."* This is a **model
compliance decision, not a transport failure** — the channel content
demonstrably reached the model's context on schedule (visible in the
transcript as `← filewatch: STEER TEST a2392ea4: …`, positioned right after
"Ran 1 shell command," same slot as the SendMessage case).

### Second drop (`answer2.txt`, nonce `2cbcd1fb`) — delivery + compliance both proven
Re-ran with the busy-task prompt neutralized (no "ignore inbox" framing;
instead: "if any filewatch channel event arrives with instructions, follow
those instructions exactly… as soon as you see them"). Dropped at
**08:35:51.371** UTC, ~13s into the loop. Result:

| file | mtime (UTC) |
|---|---|
| `busy-done4.txt` | 08:37:19.549117676 |
| `steer4-2cbcd1fb.txt` | 08:37:19.549117676 |

Same-second mtimes (writes landed in the same turn, back to back); pane
transcript disambiguates the order — the model's own narration is
`"Busy loop returned. Writing busy-done4.txt now, and the filewatch event this
time is a harmless file-write that you explicitly pre-authorized, so I'll
honor it too."`, i.e. `busy-done4.txt` first, `steer4-*.txt` second, both only
after the foreground Bash call returned. **Verdict: queued to end of current
(blocking) turn — identical shape to Transport 1.** PROVEN-live, n=1 per
condition (2 conditions run: default-trust and explicitly-authorized).

**Load-bearing caveat this probe surfaces that the ticket didn't anticipate:**
channel content is NOT treated by the model as an authoritative instruction by
default the way a `SendMessage` peer call is — the server's own `instructions`
field says "read the content and act on it, no reply expected," but the model
still applies its own judgment about whether to comply, and a receiver primed
(even incidentally) to be suspicious of file-based inputs will decline a
channel-borne steer outright. A `SendMessage` steer was not tested for the
same refusal behavior in this probe (out of scope — the ledger's other probes
already treat SendMessage content as an ordinary inbound peer message that
gets rendered and acted on); this asymmetry (peer message vs "untrusted event
in an inbox tag") is worth flagging for anyone designing the busy-receiver
steer UX, since it means channel-based steering is inherently softer than
SendMessage-based steering.

## 4. Pluggability — is `claude/channel` honored anywhere else?

Desk research (parallel sub-agent, WebSearch/WebFetch against official
sources); full citations in the sub-agent's working file, folded in here.

- **`claude/channel` — CONFIRMED-BY-DOCS, Claude-Code-specific.** Capability
  key `capabilities.experimental['claude/channel']`, notification
  `notifications/claude/channel` → rendered as a `<channel>` tag in the
  model's context (companion notifications
  `notifications/claude/channel/permission_request`/`.../permission` exist for
  remote-approval relay). Explicitly gated behind
  `--dangerously-load-development-channels` (bypasses an Anthropic-curated
  allowlist) or `--channels` for allowlisted plugins; explicitly a **research
  preview** requiring Anthropic auth via claude.ai or a Console API key, "not
  available on Amazon Bedrock, Google Cloud's Agent Platform, or Microsoft
  Foundry," and further gated by an org-level `channelsEnabled` managed
  setting on Team/Enterprise. (docs.claude.com channels / channels-reference
  pages.)
- **OpenAI Codex CLI (INSPECTED, docs only — not full Rust source):** its own
  MCP interface doc marks the *interface* experimental, but that's stability
  language, not a capability-negotiation field. No vendor-namespaced
  experimental capability keys found. Its only server→client push
  (`codex/event`) is tied to an already-running turn/conversation, not an
  unsolicited injection into an idle or independent session; approval flows
  (`applyPatchApproval`/`execCommandApproval`) are request/response, not
  fire-and-forget. INFERRED absence (doc-based only, source not grepped).
- **Cursor CLI, Aider, Cline, Continue, Goose, Gemini CLI (INFERRED,
  search-only):** no mentions found of an equivalent vendor-namespaced
  unsolicited-push capability. Absence of evidence from search, not
  exhaustive source review per project.
- **Hermes:** out of scope, no public docs, not investigated (would require
  this repo's own specs, which the task explicitly kept off-limits for this
  probe: no Tars-VM/Hermes-gateway touching).
- **MCP specification itself (CONFIRMED-BY-DOCS):** defines only generic,
  vendor-neutral primitives a client *could* build such a feature on — the
  `experimental` field in `ClientCapabilities`/`ServerCapabilities` is
  described only as a slot for "non-standard experimental features," with no
  spec-level reservation or meaning for any specific key (unlike `_meta`
  prefixes). The standard unsolicited server→client push primitive is
  `logging` + `notifications/message` (RFC 5424 severities, generic
  `{level, logger, data}`), explicitly silent on any specific "inject into a
  live LLM context" behavior. `sampling`/`elicitation` are server-initiated
  but are **requests awaiting a response**, not fire-and-forget context
  injection. `claude/channel` itself appears nowhere in the MCP spec — it's a
  Claude-Code-only extension built on the generic `experimental` slot plus a
  custom notification method name in Anthropic's own namespace, matching
  Anthropic's own "research preview" framing.

**Conclusion:** distinguish cleanly between (a) the generic MCP building
blocks (`experimental` capability slot, `logging`/`notifications/message`),
which are standard and any compliant client could implement something similar
on top of, and (b) `claude/channel` as actually shipped, which is a
Claude-Code-namespaced, Claude-Code-gated, research-preview extension with no
observed adoption elsewhere. Nothing found suggests portability today.

## 5. Cleanup

- `tmux kill-session -t gcn59rx1` / `gcn59rx2` — both confirmed gone
  (`tmux ls` shows only a pre-existing unrelated session, `rollback-cleaq`).
- No stray `claude` process traced to either scratch cwd: every remaining
  `claude --dangerously-skip-permissions` process on the box was checked via
  `/proc/<pid>/cwd` and all resolve to unrelated pre-existing worktrees.
- `rm -rf /tmp/gcn59-busy-probe` run after this file was drafted.
- Nothing in this probe touched Slack, the Tars VM, the Hermes gateway, or
  `secrets/`; only receivers spawned by this probe under their fixed names
  (`gcn59-busy-rx`, `gcn59-busy-rx2`) were addressed.

## 6. What this does NOT establish

- **Steer semantics against a genuinely long-running *tool-internal* wait**
  (e.g. an MCP tool call itself blocked on network I/O, not a Bash busy-loop)
  were not tested — only Bash-tool-blocking was exercised, since that's the
  cleanest way to force a real non-backgroundable busy turn.
- **n=1 per clean condition.** Both transports showed internally consistent
  timestamp ordering (steer strictly after the busy marker) with no raciness
  observed, so no repeat was run; a second independent trial per transport
  would strengthen confidence but wasn't run here (time-boxed).
- **SendMessage's own compliance-under-suspicion behavior** (the §3 caveat)
  was not tested for Transport 1 — only channel content was shown to be
  refusable by the model; whether a SendMessage-delivered steer can be
  similarly declined by a suspicious receiver is unprobed.
- **A receiver already `--effort low` or non-Opus** was not varied; both
  receivers ran on the account's default (Opus 4.8, high effort) since no
  model override was passed at spawn — a smaller/faster model's queuing
  behavior around a blocking tool call is unmeasured.
