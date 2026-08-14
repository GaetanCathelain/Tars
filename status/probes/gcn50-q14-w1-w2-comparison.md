# GCN-50 Q14 — session→Tars wire bake-off: W2 (`chat -q | send -t`) vs W1 (kanban wake)

**Date:** 2026-08-14 (VM clock UTC) · **Host:** Tars VM `192.168.0.9`, Hermes v0.20.0
**Run by:** Claude Code subagent on cooper · **DM:** `D0BBYNM01BL` · **Board:** `default`
**Gateway:** `ActiveState=active`, `ActiveEnterTimestamp=Fri 2026-08-14 11:35:26 UTC` — unchanged
throughout, no restart during the probe.
**Run strictly sequentially:** W2 window 12:33:26–12:34:23Z, W1 window 12:34:49–12:35:40Z.
No overlap, clean attribution.

**Verdict: W2 WINS. Both wires produce a real agent turn in the right thread. W1's
precondition is a `hermes chat` turn — i.e. W1 is W2 plus a kanban card, a subscription
row, a notifier hop and a duplicate template post.**

---

## 0. Baseline

```
2026-08-14T12:33:26Z
agent.log=6737  gateway.log=2737  errors.log=189
systemctl --user show hermes-gateway -p ActiveEnterTimestamp -p ActiveState
  ActiveState=active
  ActiveEnterTimestamp=Fri 2026-08-14 11:35:26 UTC
```

---

# TEST W2 — `hermes chat -Q -q '<prompt>' | hermes send -t slack:<DM>:<thread_ts>`

## W2.1 — CLI facts (`hermes chat --help`, run on the VM)

Oneshot flags confirmed present:

```
-q QUERY, --query QUERY   Single query (non-interactive mode)
-Q, --quiet               Quiet mode for programmatic use: suppress banner, spinner, and
                          tool previews. Only output the final response and session info.
```

Other flags of interest for a delegation wire (not exercised):
`-m MODEL`, `-t TOOLSETS`, `--reasoning LEVEL`, `-s SKILLS`, `--max-turns N`,
`--resume SESSION_ID`, `--continue [SESSION_NAME]`, `--pass-session-id`, `--source SOURCE`.

## W2.2 — Pipe-safety, measured (not assumed)

```
$ date -u +%FT%TZ; ~/.local/bin/hermes chat -Q -q "reply with exactly OK" \
    2>/tmp/probe-lane/chat_stderr.txt | cat -A | head -20; echo "EXIT=$?"
2026-08-14T12:33:40Z
OK$
EXIT=0
--- stderr ---
session_id: 20260814_123341_2bf5ba
2026-08-14T12:33:49Z
```

**stdout carries ONLY the answer** (`OK$` = `OK\n` — no ANSI, no banner, no trailing blank
line). The "session info" the help text mentions goes to **stderr**. `-Q` is pipe-safe as
shipped. Trivial-query cost ≈ 9 s.

## W2.3 — Lane manifest (the rehydration target)

```
$ N=$(openssl rand -hex 2); mkdir -p /tmp/probe-lane; cat > /tmp/probe-lane/LANE.md <<EOF
...
EOF
NONCE=W2-NONCE-3ea8
--- manifest ---
# Lane: probe-lane
- Lane name: probe-lane
- Linear ticket: GCN-999 — [PROBE gcn50-q14] fake ticket for wire test
- Decision log:
  - D1: chose wire W2 (chat -q | send -t) for the milestone report
  - D2: nonce carried in this manifest to prove file rehydration
- Nonce: W2-NONCE-3ea8
```

The nonce `W2-NONCE-3ea8` exists **only in that file** — it is never in the prompt. Any
occurrence of it in Slack proves the agent opened the file.

## W2.4 — Anchor

```
$ ~/.local/bin/hermes send -t slack:D0BBYNM01BL --json "[PROBE gcn50-q14-W2] anchor"
{"success": true, "platform": "slack", "chat_id": "D0BBYNM01BL",
 "message_id": "1786710836.117089", "mirrored": true}
```
W2 anchor ts = **1786710836.117089** (12:33:56Z).

## W2.5 — The wire, timed

```bash
~/.local/bin/hermes chat -Q -q "A delegated coding session reports: milestone reached on lane
probe-lane. Read /tmp/probe-lane/LANE.md and compose a one-line status update for the operator
that includes the nonce from that file." 2>w2_chat_stderr.txt \
  | tee w2_chat_stdout.txt \
  | ~/.local/bin/hermes send -t slack:D0BBYNM01BL:1786710836.117089 --json
```

```
AGENT_LOG_LINES_BEFORE=6823
START=2026-08-14T12:34:08.536Z
END=2026-08-14T12:34:23.093Z
RC=0
ELAPSED_S=14.555745928
AGENT_LOG_LINES_AFTER=6898
--- chat stdout (piped payload, cat -A) ---
probe-lane milestone reached (GCN-999); nonce: W2-NONCE-3ea8.$
--- chat stderr ---
session_id: 20260814_123409_bf539f
--- send json ---
{"success": true, "platform": "slack", "chat_id": "D0BBYNM01BL",
 "message_id": "1786710862.843289"}
--- send stderr ---   (empty)
```

## W2.6 — (a) Real agent turn? YES

`agent.log`, session `20260814_123409_bf539f` (8 lines, verbatim, log-formatter-truncated at
~100 chars as it is on disk):

```
12:34:12,480 INFO [20260814_123409_bf539f] agent.turn_context: conversation turn: session=20260814_123409_...
12:34:16,756 INFO [20260814_123409_bf539f] agent.conversation_loop: API call #1: model=gpt-5.6-sol provide...
12:34:16,776 WARNING [20260814_123409_bf539f] agent.tool_executor: Tool kanban_show returned error (0.00s)...
12:34:21,493 INFO [20260814_123409_bf539f] agent.conversation_loop: API call #2: model=gpt-5.6-sol provide...
12:34:21,523 INFO [20260814_123409_bf539f] agent.conversation_loop: Turn ended: reason=text_response(finis...
12:34:21,529 INFO [20260814_123409_bf539f] tools.terminal_tool: Manually cleaned up environment for task: ...
12:34:21,530 INFO [20260814_123409_bf539f] tools.terminal_tool: Cleaned 1 environments
12:34:21,719 INFO [20260814_123409_bf539f] cli: CLI cleanup calling memory shutdown for session 20260814_1...
```

`conversation turn:` + 2 model API calls + tool execution + `Turn ended` ⇒ genuine turn.
`gateway.log` has **no** lines in the 12:34 window — the CLI chat session does not go through
the gateway at all.

## W2.7 — (b) What landed in Slack

`slack_read_thread(D0BBYNM01BL, 1786710836.117089)` — verbatim:

```
=== THREAD PARENT MESSAGE ===
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 14:33:56 CEST
Message TS: 1786710836.117089
[PROBE gcn50-q14-W2] anchor

=== THREAD REPLIES (1 total) ===

--- Reply 1 of 1 ---
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 14:34:22 CEST
Message TS: 1786710862.843289
probe-lane milestone reached (GCN-999); nonce: W2-NONCE-3ea8.
```

- **In the thread.** Nothing at channel top level (confirmed by `slack_read_channel`, §W1.8).
- **Composed, not templated.** It rewrote the manifest into an operator-facing sentence, pulled
  `GCN-999` out of the ticket line and obeyed "one line".
- **Rehydration PROVEN.** `W2-NONCE-3ea8` was only ever on disk.

## W2.8 — (c) Latency, (d) pollution

- **End-to-end 14.56 s** (12:34:08.536 → 12:34:23.093Z), one process pair, no polling.
- **No stdout pollution.** `cat -A` shows a single clean line, one trailing `\n`. Slack rendered
  it exactly. No banner, no session id, no ANSI.
- Two cosmetic notes, neither blocking:
  - The agent speculatively called `kanban_show` and got a `WARNING … returned error` before
    answering — one wasted tool round trip (~4 s of the 14.5 s). A tighter prompt or
    `-t <toolsets>` would drop it.
  - `hermes send --json` to a **threaded** target returns no `"mirrored"` key; the top-level
    anchor send returns `"mirrored": true`. Cosmetic difference in the JSON shape only.

---

# TEST W1 — kanban wake with a session-stamped card + `--thread-id`

## W1.1 — Anchor

```
$ ~/.local/bin/hermes send -t slack:D0BBYNM01BL --json "[PROBE gcn50-q14-W1] anchor"
{"success": true, "platform": "slack", "chat_id": "D0BBYNM01BL",
 "message_id": "1786710890.363749", "mirrored": true}
```
W1 anchor ts = **1786710890.363749** (12:34:50Z).

## W1.2 — Path A: card created FROM a session (Path B not needed)

```
$ ~/.local/bin/hermes chat -Q -q "List the exact names of any kanban tools you have available.
Then, if you have a card-creation tool, use it to create a card titled \"[PROBE gcn50-q14-W1]
wake test\" with body \"probe card; do not action\" and initial status blocked if that parameter
is supported. Reply with: the tool names, and the created card id."

Kanban tools: kanban_attach, kanban_attach_url, kanban_attachments, kanban_block,
kanban_comment, kanban_complete, kanban_create, kanban_heartbeat, kanban_link,
kanban_list, kanban_show, kanban_unblock

Created card ID: t_bde61b26

--- stderr ---
session_id: 20260814_123451_f9d0b9
```
(12:34:50Z → 12:35:16Z, **26 s**.)

Chat sessions **do** carry the kanban toolset, including `kanban_create`, and it accepts an
initial blocked status (so the dispatcher never spawns a worker on the probe card).

**Row after creation — `session_id` IS stamped (Path A succeeded; no direct sqlite writes were
ever needed, so no before/after mutation to record):**

```json
{"id": "t_bde61b26", "title": "[PROBE gcn50-q14-W1] wake test",
 "body": "probe card; do not action", "assignee": "default", "status": "blocked",
 "created_by": "worker", "created_at": 1786710906, "started_at": null, "completed_at": null,
 "workspace_kind": "scratch", "result": null,
 "session_id": "20260814_123451_f9d0b9", "block_kind": null, "block_recurrences": 0}
```

Same shape as the 2026-08-07 positive control (`created_by='worker'` + non-null `session_id`),
confirming the q10 root cause: the gate is on `session_id`, and only an in-session
`kanban_create` sets it.

**The stamped session was DEAD before the trigger** — it was a `-q` oneshot and exited:

```
12:35:15,024 INFO [20260814_123451_f9d0b9] agent.conversation_loop: Turn ended: reason=text_response(finis...
12:35:15,031 INFO [20260814_123451_f9d0b9] tools.terminal_tool: Cleaned 1 environments
12:35:15,243 INFO [20260814_123451_f9d0b9] cli: CLI cleanup calling memory shutdown for session 20260814_1...
```
Trigger was 15 s later. This is the coupling test the task asked for — see §W1.6.

## W1.3 — Subscribe + pre-trigger state

```
2026-08-14T12:35:29Z
PRE_AGENT=7007 PRE_GATEWAY=2737 PRE_ERRORS=204

$ ~/.local/bin/hermes kanban notify-subscribe t_bde61b26 --platform slack \
    --chat-id D0BBYNM01BL --chat-type dm --thread-id 1786710890.363749 --user-id U08BDJAMSRZ
Subscribed slack:D0BBYNM01BL:1786710890.363749 to t_bde61b26

$ ~/.local/bin/hermes kanban notify-list t_bde61b26 --json
[{"task_id":"t_bde61b26","platform":"slack","chat_id":"D0BBYNM01BL","chat_type":"dm",
  "thread_id":"1786710890.363749","user_id":"U08BDJAMSRZ","notifier_profile":"default",
  "delivery_metadata":{},"created_at":1786710929,"last_event_id":18}]
```

## W1.4 — Trigger

```
=== TRIGGER ===
2026-08-14T12:35:30.238Z
$ ~/.local/bin/hermes kanban complete t_bde61b26 \
    --result "[PROBE gcn50-q14-W1] wake-path probe complete — no action taken"
Completed t_bde61b26
2026-08-14T12:35:30.574Z
```

## W1.5 — (a) Real agent turn? YES

Wake landed **1.8 s** after the trigger. `gateway.log` lines 2738-2741, **full untruncated**:

```
2026-08-14 12:35:32,346 INFO gateway.run: kanban notifier: woke agent for t_bde61b26 on slack/D0BBYNM01BL profile=default events={'completed'}
2026-08-14 12:35:32,352 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='[kanban] Task t_bde61b26 completed. Title: [PROBE gcn50-q14-W1] wake test Assign' reply_to_id=None reply_to_text=''
2026-08-14 12:35:40,102 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=7.7s api_calls=2 response=136 chars
2026-08-14 12:35:40,131 INFO gateway.platforms.base: [Slack] Sending response (136 chars) to D0BBYNM01BL
```

`agent.log` delta (from line 7008), the wake turn:

```
12:35:32,346 INFO gateway.run: kanban notifier: woke agent for t_bde61b26 on slack/D0BBYNM01BL profile=def...
12:35:32,352 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='[kan...
12:35:32,435 INFO [20260814_112121_d311e0fc] run_agent: OpenAI client created (agent_init, shared=True) th...
12:35:32,438 INFO [20260814_123532_e2cc0014] agent.turn_context: conversation turn: session=20260814_12353...
12:35:32,498 INFO run_agent: OpenAI client created (codex_stream_request, shared=False) thread=Thread-252 ...
12:35:35,512 INFO [20260814_123532_e2cc0014] agent.conversation_loop: API call #1: model=gpt-5.6-sol provi...
12:35:35,535 INFO [20260814_123532_e2cc0014] agent.tool_executor: tool kanban_show completed (0.00s, 2096 ...
12:35:39,847 INFO [20260814_123532_e2cc0014] agent.conversation_loop: API call #2: model=gpt-5.6-sol provi...
12:35:39,883 INFO [20260814_123532_e2cc0014] agent.conversation_loop: Turn ended: reason=text_response(fin...
12:35:39,899 INFO agent.auxiliary_client: Auxiliary auto-detect: using main provider openai-codex (gpt-5.6...
12:35:40,102 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=7.7s api_calls=2 response=136 chars
12:35:40,131 INFO gateway.platforms.base: [Slack] Sending response (136 chars) to D0BBYNM01BL
```

`errors.log` delta over the whole W1 window: **empty** (204 → 204 lines).

This is the first end-to-end confirmation of the branch q10 could only read in source
(`kanban_watchers.py` L729-772) — and the first time a card in this DB has had both a
`session_id` and a `--thread-id` subscription.

## W1.6 — CORE FINDING: `session_id` is a pure non-empty gate, NOT a live-session handle

The wake did **not** resume `20260814_123451_f9d0b9` and did **not** error on it being dead.
The gateway opened a **brand-new session `20260814_123532_e2cc0014`** and ran the turn there.
Post-trigger, the stamped session id appears **nowhere** in `agent.log` (its last line is the
12:35:15 shutdown).

So `deliver_wake(..., session_id=_session_key, ...)` uses the value only to pass the
`if _wake_kinds and _session_key` guard; the turn itself is routed by `SessionSource`
(platform/chat/thread/user), not by the stamped id. Consequences:

- A **stale or foreign** `session_id` is fine — the wake still fires. No dead-session failure
  mode exists. (This *removes* a coupling risk, but it also means the stamp is bookkeeping
  the notifier does not honour: the woken agent has **none of the originating session's
  conversation state**.)
- Since a **CLI** `kanban create` cannot stamp it (q10) and only an in-session `kanban_create`
  can, **a value that is never read back still has to be minted by a live `hermes chat` turn.**
  That is W1's whole cost, for nothing the wake actually uses.

## W1.7 — (b) Delivery: in-thread, but TWO messages — template + composed

`slack_read_thread(D0BBYNM01BL, 1786710890.363749)` — verbatim:

```
=== THREAD PARENT MESSAGE ===
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 14:34:50 CEST
Message TS: 1786710890.363749
[PROBE gcn50-q14-W1] anchor

=== THREAD REPLIES (2 total) ===

--- Reply 1 of 2 ---
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 14:35:32 CEST
Message TS: 1786710932.041549
:heavy_check_mark: [default] @default Kanban t_bde61b26 done — [PROBE gcn50-q14-W1] wake test
[PROBE gcn50-q14-W1] wake-path probe complete — no action taken

--- Reply 2 of 2 ---
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 14:35:40 CEST
Message TS: 1786710940.288419
Probe réussi : la tâche `t_bde61b26` est *done* et le wake-path s'est exécuté comme prévu, sans aucune action. Aucun suivi nécessaire.
```

- **Both** deliveries land in the anchor thread — `--thread-id` propagates through
  `SessionSource(thread_id=…)` to the agent's own reply, confirming the q10 "untested"
  hypothesis.
- The two paths are **not** mutually exclusive in practice: the raw template post (reply 1) fires
  **and** the agent turn (reply 2). One operator-facing event ⇒ **two Slack messages**.
- Reply 2 **is** composed output (French, Tars' DM persona, `*done*` markdown) — but it is
  composed from what `kanban_show` returned, i.e. from the card. It answers "what happened to
  the card", not "what is the lane status".
- **Rehydration NOT demonstrated for W1.** The synthetic wake payload is a fixed template —
  `[kanban] Task <id> completed. Title: <title> Assign…` — with no room for a file path or lane
  pointer. Any context must be smuggled through the card's `title` / `body` / `--result`, which
  the agent then reads via `kanban_show`. Nothing in the wire carries "read /path/LANE.md".

## W1.8 — (c) Latency

| leg | wall |
|---|---|
| card creation via `hermes chat` (the `session_id` precondition) | 26 s (12:34:50 → 12:35:16Z) |
| `notify-subscribe` + `notify-list` | ~1 s |
| trigger → template post in thread | **1.8 s** (12:35:30.24 → 12:35:32) |
| trigger → composed agent reply in thread | **10.0 s** (12:35:30.24 → 12:35:40.13) |
| **trigger → composed reply, incl. precondition** | **~37 s** |

Gateway self-reported turn cost: `time=7.7s api_calls=2`.

No stray top-level posts — `slack_read_channel(D0BBYNM01BL)` shows both anchors carrying their
replies as threads (`Thread: 1 replies` / `Thread: 2 replies`) and no probe message at channel
level. The only other Tars messages in the window are unrelated cron nudges (GCN-34, GCN-56).

---

# COMPARISON

| | **W2 — `chat -Q -q … \| send -t slack:DM:ts`** | **W1 — kanban wake + `--thread-id`** |
|---|---|---|
| Real agent turn? | **YES** — session `20260814_123409_bf539f`, 2 API calls, `Turn ended` | **YES** — session `20260814_123532_e2cc0014`, 2 API calls, `woke agent` line |
| Lands in the target thread? | **YES**, exactly one message | **YES**, but **two** messages (template + composed) |
| Composed content? | **YES** — obeyed "one line", extracted GCN-999 | **YES** for reply 2; reply 1 is the fixed `✔ [board] Kanban … done` template |
| Rehydration from a file proven? | **YES** — `W2-NONCE-3ea8`, present only on disk | **NO** — wake payload is a fixed `[kanban] Task … completed` string; context can only ride in card title/body/result |
| Arbitrary prompt/context? | **YES** — full free-text prompt, any path, any instruction | **NO** — no prompt surface at all |
| Latency (trigger → Slack) | **14.6 s** total | **10.0 s** after trigger, **~37 s** incl. the mandatory card-creation turn |
| Moving parts | 2 processes, 1 pipe. No DB, no gateway, no subscription | live chat turn → `kanban_create` → sqlite row → `notify-subscribe` row → `complete` → notifier → gateway wake → new DM session |
| Persistent state left behind | none (session exits) | card row + subscription row in `~/.hermes/kanban.db` (subscription auto-clears on final status) |
| Failure surface | `hermes chat` non-zero → pipe carries empty/partial; `send` reports `success:false` in `--json` | needs a card whose `session_id` was stamped **by a live chat turn** — CLI `create` can't; notifier is fire-and-forget with no ack to the caller |
| Coupling to a live session | none — CLI session is ephemeral and self-contained | **none observed either**: stamped session was already dead, wake fired anyway into a *fresh* session. The stamp is a gate, not a handle — so the woken agent inherits **no** state from the reporting session |
| Runs through the gateway / Tars DM persona? | **NO** — bare CLI agent session, output as instructed (terse English) | **YES** — full DM session, replied in Tars' operator voice (French) |
| Caller gets the text back? | **YES** — it is on stdout, inspectable/loggable before it ships | **NO** — the agent's reply exists only in Slack |

## Failure modes actually observed

- **W2:** one speculative `kanban_show` tool error (`WARNING … returned error`) before the
  answer — cost ≈ 4 s, no impact on output. `--json` on a threaded `send` omits `"mirrored"`.
- **W1:** duplicate delivery (template + composed) for a single event. Wake payload is
  uncustomisable. `session_id` precondition can only be met by a `hermes chat` turn, which is
  the very thing W1 was meant to avoid.

---

# RECOMMENDATION — use W2

`hermes chat -Q -q '<prompt>' | hermes send -t slack:D0BBYNM01BL:<thread_ts>` is the wire for
session→Tars reporting.

1. **It is strictly simpler and strictly more capable.** Two processes and a pipe, versus a
   card + a subscription row + the notifier + a gateway wake — for the same single outcome, a
   composed message in a thread.
2. **W1 cannot avoid W2.** Its `session_id` gate is only satisfiable from inside a `hermes chat`
   turn. Anyone who can do that already has W2 — and W2 delivers the message directly instead of
   using the card as a doorbell.
3. **Only W2 can carry the lane.** Rehydration from `/tmp/probe-lane/LANE.md` is proven by
   nonce. W1's wake text is a fixed kanban template with no prompt surface; lane context would
   have to be stuffed into a card body.
4. **One message, not two.** W1 posts the raw `✔ [default] Kanban … done` template alongside the
   agent's reply. Noisy for an operator-facing DM.
5. **The output is inspectable before it ships.** W2's composed text is on stdout — the calling
   session can log it, gate on it, or retry. W1's reply only ever exists in Slack.
6. **`-Q` is pipe-safe as shipped** — measured, not assumed: answer on stdout, `session_id` on
   stderr, no ANSI, single trailing newline.

Practical notes for building on W2:
- Post an anchor once per lane (`hermes send -t slack:D0BBYNM01BL --json`, take `message_id`),
  then thread every milestone under it — that is exactly the shape probed here.
- Budget ~15 s per report; consider `-t <toolsets>` or a tighter prompt to skip the speculative
  kanban tool call.
- Check `RC` of the pipeline **and** `"success"` in the `send --json` output; an empty chat
  stdout posts an empty message.
- W2's session is *not* the gateway DM session — it carries no conversation history with
  Gaetan, and its persona/voice differs from Tars-in-Slack (English vs French here). If the
  update must read in Tars' DM voice, say so in the prompt.

**Keep W1 for what it is for:** notifying on cards a Hermes worker genuinely owns, where the
kanban lifecycle *is* the event. It is not a session→Tars transport.

---

# Cleanup / footprint

```
$ hermes kanban archive t_bde61b26      -> Archived t_bde61b26
$ hermes kanban notify-list --json      -> []
   card row after: {"id": "t_bde61b26", "status": "archived"}
$ rm -rf /tmp/probe-lane
   ls: cannot access '/tmp/probe-lane': No such file or directory
$ errors.log delta since line 204       -> (empty)
```

- **No direct sqlite writes.** Path A stamped `session_id` natively, so Path B (manual UPDATE)
  was never executed. Every kanban mutation went through the `hermes kanban` CLI or the agent's
  own `kanban_create`.
- No `~/.hermes/config.yaml` or `.env` edit, no read of `.env`, **no `sops` invocation of any
  form**. No secret or token-like value appears in this file.
- No gateway restart (`ActiveEnterTimestamp` identical at start and end).
- Slack artifacts left in place per instructions, all tagged `[PROBE gcn50-q14-…]`:
  - W2 anchor `1786710836.117089` + composed reply `1786710862.843289`
  - W1 anchor `1786710890.363749` + template reply `1786710932.041549` + composed reply
    `1786710940.288419`
