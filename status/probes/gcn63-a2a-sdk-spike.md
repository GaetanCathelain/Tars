# GCN-70 — A2A Claude-lane wrapper: live spike

Ran on **cooper**, 2026-08-16 10:10–10:18 UTC. All scratch under `/tmp/gcn70*`;
nothing written under `$HOME` except one accidental file, removed (see §Surprises).
VM work was strictly read-only.

**Environment measured, not assumed**
- `claude` CLI `2.1.233`, `/home/gaetan/.local/bin/claude`
- Python `3.14.6`; `claude-agent-sdk` **0.2.139** in `/tmp/gcn70-venv` (PyPI, clean install)
- npm `@anthropic-ai/claude-agent-sdk` **0.3.233** exists (version-checked only, not exercised — the Python SDK installed cleanly first and both wrap the same CLI)
- `env | grep -i anthropic` → **no matches**; every run additionally forced through `env -u ANTHROPIC_API_KEY`

---

## Verdict table

| # | Question | Verdict | Confidence |
|---|---|---|---|
| 1 | Agent SDK works headless on the OAuth subscription, no `ANTHROPIC_API_KEY` | **YES** — completed in 4.6 s, `subtype=success`, zero auth stderr | **PROVEN-live** |
| 2a | Mid-task user message lands *during* an in-flight foreground tool | **Accepted mid-turn, acted on at the next sampling step** — never rejected, never lost, but it cannot preempt a running tool. Same run, one `ResultMessage`, `num_turns=3` | **PROVEN-live** |
| 2b | `client.interrupt()` actually stops the current turn | **YES, hard** — 10 ms to terminate, in-flight `Bash` child killed, completion file never written, `subtype=error_during_execution`, no orphan process | **PROVEN-live** |
| 3 | Stdlib A2A facade over one SDK session: `message/send` → poll → `completed`, incl. `input_required` round trip | **YES, both halves** — nonce file written; and `working → input_required → deny+steer → input_required → allow → completed` driven entirely over curl | **PROVEN-live** |
| 4 | Hermes A2A inbound session identity | **Isolated `a2a` namespace, caller-keyed by `contextId`** — cannot bind to a Slack DM-thread session; but a caller reusing a `contextId` deterministically reuses one A2A session, and `contextId` is unauthenticated caller input | **INSPECTED** |

**One-line spike verdict:** the A2A-wrapped Agent SDK lane is mechanically sound —
subscription auth, hard interrupt, and a full `input_required` round trip all work
live — with two real constraints: steering queues behind the in-flight tool, and the
Hermes side gives an A2A caller its own session, never Tars' DM session.

---

## §1 Subscription auth headless — PROVEN-live

Script `/tmp/gcn70/t1_auth.py`, plain `query()`, `max_turns=1`, `allowed_tools=[]`.

```
ANTHROPIC env keys: NONE
TEXT: AUTH-OK-gcn70
RESULT subtype=success is_error=False cost=0.34317 dur=4.6s
```

- No `ANTHROPIC_API_KEY`, no `ANTHROPIC_AUTH_TOKEN`, no auth-related stderr at all.
- The SDK spawns the `claude` CLI as a subprocess, so it inherits the CLI's stored
  OAuth login. **A lane host therefore needs a logged-in `claude` CLI for that uid —
  not an API key.** That is a provisioning fact for whatever runs the lane: it is a
  human-interactive `claude` login once per host/uid, not a secret we can hand it.
- `total_cost_usd` is still populated (0.343) on a subscription run. It is a computed
  equivalent, not a charge — do not read it as spend. **INFERRED** from the run alone.

## §2a Steer semantics — PROVEN-live

`/tmp/gcn70/t2a_steer.py`. Foreground CPU busy loop (`while time.time()-t<30`, ~1.2e8
iterations — genuinely foreground, explicitly not backgrounded), then a second
`client.query(...)` pushed **at t+22.98 s, mid-loop**. Timestamps relative to process start:

```
 0.97s  sent BUSY task
 9.08s  TOOL_USE Bash   (busy loop starts)
22.98s  STEER MESSAGE SENT          <-- injected while the tool is running
39.45s  busy-done.txt written       (loop finishes)
44.11s  TOOL_USE Write steer-n7a2.txt
45.54s  TEXT "BUSY-FINISHED"
45.57s  RESULT success num_turns=3
```

Reading, honestly:

- The injected message is **accepted while a tool is executing** — no error, no block,
  no dropped message. `ClaudeSDKClient.query()` just writes a `{"type":"user"...}` line
  to the CLI's stdin, so it never contends with the running turn.
- It is **not a preemption**. The busy loop ran to completion (39.45 s) and the steer
  file was only written 4.7 s later. The message is queued and consumed at the next
  model sampling step, which is the first moment after the in-flight tool returns.
- Crucially it is still the **same run**: one `ResultMessage`, `num_turns=3`. It does
  not wait for the current task to *finish and return* before being seen — it is
  mid-run, just not mid-tool. An earlier control run (steer sent at t+8.98 s, before
  the tool started) behaved identically and the model narrated it as
  "Handling the mid-turn instruction now".

**Consequence for Tars:** "steer the lane" is real but its latency floor is the
duration of whatever tool call is in flight (a 20-minute test run = 20-minute steer
latency). To stop something *now*, §2b is the mechanism, not §2a.

## §2b Interrupt — PROVEN-live

`/tmp/gcn70/t2b_interrupt.py`, same busy loop, `await client.interrupt()` at t+20 s.

```
 1.51s  sent BUSY task
13.68s  TOOL_USE Bash   (busy loop starts, would end ~43.7s)
21.52s  INTERRUPT() CALLED
21.53s  RESULT error_during_execution is_error=True num_turns=3
21.53s  stream done
21.97s  FILE busy-done.txt ABSENT
```

- The turn stops in ~10 ms. The stream ends with `ResultMessage(subtype="error_during_execution", is_error=True)` — that subtype is how a caller distinguishes an interrupt from a failure, and it is not obviously distinguishable from a genuine error. **Map it carefully in an A2A `canceled` state.**
- The completion file was **never written**: the in-flight `python3` child was actually
  killed, not merely detached. Verified independently — `pgrep -af "time.time()-t<30"`
  returned nothing afterwards. This is the property an A2A `tasks/cancel` needs.
- Not tested for time: whether the same client can be re-queried after an interrupt
  and continue on the same session. **UNTESTED** — assume it works (the docs' redirect
  pattern is interrupt-then-query) but prove it before building cancel-and-redirect.

## §3 Minimal A2A facade — PROVEN-live

`/tmp/gcn70/a2a_facade.py`, 129 lines, stdlib + `claude_agent_sdk` only, one SDK
session, `ThreadingHTTPServer` on `127.0.0.1`. Source in §Facade source below.

**Happy path** (port 9911, nonce `f9k2`):

```
10:15:38 SEND    -> taskId=24035c3c58e1  {"status":{"state":"submitted"}}
10:15:38 poll1 state=working
10:15:44 poll4 state=completed
         artifacts[0].parts[0].text == "DONE"
         facade.log: TOOL_USE Write {"file_path": "facade-f9k2.txt", "content": "DONE\n"}
         /tmp/gcn70/f/facade-f9k2.txt exists, 5 bytes, mtime 10:15:41
```

**`input_required` round trip** (port 9912, nonce `f4q8`) — the interesting half, and it
is **reachable**, not specced-not-proven:

```
10:17:18 SEND   -> taskId=6bbc69a396af
10:17:22 poll3  state=input_required
         status.message.parts[0].text = 'approve tool Write: {"file_path": "/home/gaetan/facade-f4q8.txt", ...}'
10:17:36 message/send #2: "deny - never write outside /tmp; use /tmp/gcn70/f/facade-f4q8.txt"
10:17:38 poll2  state=input_required
         status.message.parts[0].text = 'approve tool Write: {"file_path": "/tmp/gcn70/f/facade-f4q8.txt", ...}'
10:17:38 message/send #3: "allow"
10:17:40 poll3  state=completed
         /tmp/gcn70/f/facade-f4q8.txt exists, 5 bytes, mtime 10:17:38
```

That is the full A2A shape exercised over curl: task submitted → working →
`input_required` with an agent question → operator answer denies **and steers** →
agent retries with the corrected input → second `input_required` → allow → completed,
with the side effect verified on disk. The deny message text reaches the model as
usable feedback (it corrected the path from the deny reason alone).

**How to actually get `input_required` — the non-obvious bit.** `can_use_tool` alone
does **not** fire headless. Two separate shadowing effects, both measured:

1. `allowed_tools=["Bash","Write","Read"]` emits `CanUseToolShadowedWarning` — a
   whole-tool grant auto-approves before the callback is consulted.
2. With `allowed_tools` removed entirely, `can_use_tool` *still* never fired.
   `/tmp/gcn70/t3b_perm.py` ran a Bash call twice — with `setting_sources=None`
   (default) and `setting_sources=[]` — and both printed
   `can_use_tool calls = []` while the tool executed normally.

The working recipe is a **`PreToolUse` hook returning `permissionDecision: "ask"`**,
which routes the call into `can_use_tool`; the facade then blocks on an
`asyncio.Future` there and that block *is* the `input_required` state
(`types.py:223-224` documents the ask→callback forwarding; the SDK's own warning text
recommends the hook). Give the `HookMatcher` a long `timeout` (600 s here) or the
human-in-the-loop wait dies at the 60 s default.

## §4 Hermes A2A inbound session identity — INSPECTED (read-only, VM 192.168.0.9)

Source: `~/.hermes/hermes-agent/plugins/platforms/a2a/` plus the shared gateway.

- `adapter.py:705` — `context_id = protocol.extract_context_id(params) or protocol.new_context_id()`
- `adapter.py:764-773` — the `MessageEvent` is built with
  `source=self.build_source(chat_id=context_id, chat_name=f"a2a:{peer}", chat_type="dm", user_id=peer, ...)`, `message_id=task_id`
- `adapter.py:779` — dispatch: `asyncio.run_coroutine_threadsafe(self.handle_message(event), self._loop)`
- `gateway/platforms/base.py:5584-5588` — `handle_message` turns that source into the key via `build_session_key(event.source, ...)`
- `gateway/session.py:1103-1113` — DM branch yields **`agent:main:a2a:dm:<context_id>`** (no Slack `scope_id`, because platform ≠ SLACK)

**Answers:**

1. **Isolated.** Every inbound A2A message lands in the `a2a` platform namespace. There
   is no path by which it binds to an existing Slack DM-thread session — same mechanism
   family as `hermes webhook`'s isolated sessions (`gateway/platforms/webhook.py:891-892`
   builds its source the same way, different namespace), confirming the
   gcn59-resume-row-identity read.
2. **But it is caller-keyed, not per-request.** `protocol.py:358-362`:
   ```python
   def extract_context_id(params: dict) -> str:
       msg = params.get("message") or {}
       ctx = str(msg.get("contextId") or "") if isinstance(msg, dict) else ""
       return ctx or str(params.get("contextId") or "")
   ```
   The caller's `contextId` is trusted verbatim with no auth binding. Reusing it across
   `message/send` calls **deterministically reuses the same Hermes session** — so a
   long-lived A2A conversation with memory is available today, for free. There is no
   `session_id` parameter; `contextId` is the only handle.
3. **Security note, flagged not fixed:** because `contextId` is unauthenticated caller
   input, any peer that can reach the A2A port can address *any* `a2a:dm:*` session,
   including one another peer is using. Today that is inert — the listener is bound to
   `127.0.0.1` with no token (recorded in `docs/specs/tars-profile.md:135-140,196`). It
   stops being inert the moment `A2A_HOST` leaves localhost; a peer token would then be
   doing double duty as the only thing separating sessions.
4. Registration: `plugin.yaml` declares the platform plugin, `__init__.py` is the entry
   point, `adapter.py:139` registers the JSON-RPC method table including
   `"message/send": ("send", False)`.
5. A separate mechanism exists for forwarded/served-agent profiles —
   `adapter.py:852-893`, a DB-title-based resumption keyed on
   `(profile, agent_slug, context_id)`. Not the local-agent path above; noted so nobody
   confuses the two.

---

## Headless permissions, in one place

- **Coarse:** `ClaudeAgentOptions(permission_mode=...)`, values
  `"default" | "acceptEdits" | "plan" | "bypassPermissions" | "dontAsk" | "auto"`
  (`types.py:25-27`). `bypassPermissions` is the "just run it" lane setting and is what
  §2a/§2b used; it also disables `can_use_tool` outright.
  Changeable mid-session via `await client.set_permission_mode(mode)`.
- **Per-call:** `can_use_tool(tool, input, ctx) -> PermissionResultAllow(updated_input=...)
  | PermissionResultDeny(message=..., interrupt=bool)`. Note `updated_input` — the
  approver can *rewrite* the tool call, not just yes/no it. And `PermissionResultDeny`
  carries `interrupt: bool` — deny-and-stop is available without a separate interrupt.
- **The gate that actually fires headless:** a `PreToolUse` hook returning
  `permissionDecision: "ask"|"deny"|"allow"|"defer"`. Anything less specific gets
  auto-approved, as measured in §3.
- Practical lane shape: `bypassPermissions` for a trusted sandboxed lane; the
  hook+`can_use_tool` pair only where a real human gate is wanted, because it costs a
  network round trip per tool call.

## What a headless SDK lane loses vs a TUI lane

Honestly: the pane and the human in front of it. A TUI lane is a place Gaetan can look
at mid-run — scrollback, the diff about to be applied, the tool it is grinding on — and
where he can simply type into the same box the agent is reading, take the keyboard, or
Ctrl-C. The headless lane has none of that surface: the only view is whatever the
wrapper chooses to accumulate and expose (this facade exposes concatenated assistant
text and nothing else — no tool inputs, no diffs, no partial output, no token/context
telemetry), and the only intervention is whatever the wrapper's API offers. Takeover
degrades from "sit down at the terminal" to "send another `message/send` and hope the
lane is at a sampling step", with §2a's tool-duration latency floor. Against that, the
headless lane gains the things a pane never had: it is programmatically addressable from
another machine, its permission gate is a callback that can route to Slack instead of a
keyboard, its interrupt is an API call rather than a keystroke, and it does not die with
a tmux server. Neither is strictly better — the pane is for work Gaetan wants to watch,
the headless lane is for work Tars needs to drive.

---

## Surprises / gotchas worth carrying into the build

1. **`can_use_tool` is a trap.** It looks like *the* permission API and it silently never
   fires. Two independent shadowing paths, and `setting_sources=[]` does not rescue it.
   Anyone building the wrapper will lose an hour here. Use the `PreToolUse` hook.
2. **`cwd` does not tell the model where it is.** With `cwd="/tmp/gcn70/f"`, relative
   paths resolve correctly (`facade-f9k2.txt` landed in `/tmp/gcn70/f`), but when the
   model chose to absolutize "the current directory" it produced **`/home/gaetan/...`**.
   One earlier run therefore wrote `/home/gaetan/facade-f7c3.txt` — **removed at 10:18
   UTC**, and it is the one violation of the keep-everything-in-/tmp rule in this spike.
   A lane wrapper must pin the working directory in the prompt or the system prompt, not
   only in `ClaudeAgentOptions.cwd`.
3. **A task can report `completed` having done nothing.** The very first facade run
   (port 9910, `allowed_tools` set) returned `state=completed` with artifact text
   `"DONE"` and logged no tool use — the model asserted success without writing the
   file. Completion state is the model's claim, not evidence. Any real lane needs a
   verifiable side-effect check, exactly as the orchestration policy already demands of
   agents.
4. **Interrupt and error look alike.** Both surface as `is_error=True`; only
   `subtype="error_during_execution"` distinguishes the interrupt, and that subtype is
   plausibly reachable by genuine failures too. Track "did I call interrupt" wrapper-side
   rather than inferring it from the result.
5. **Auth is a login, not a secret.** Nothing to put in SOPS, and nothing a systemd unit
   can be handed — the lane host needs an interactively logged-in `claude` CLI for the
   uid that runs the wrapper. That is a provisioning step with a human in it.
6. **`contextId` is the whole session identity on the Hermes side** and it is
   unauthenticated. Free session continuity today; a real access-control question the
   day A2A leaves localhost.

## Cleanup

- Facade servers on 9910/9911/9912: terminated (`ss -ltnp` shows no listeners; 0
  processes matching the venv python).
- Interrupt test left no orphan child (`pgrep` verified empty).
- `/home/gaetan/facade-f7c3.txt` removed.
- **Left in place deliberately:** `/tmp/gcn70/` (scripts, logs, nonce files) and
  `/tmp/gcn70-venv/` (~; a re-run needs it and it is under `/tmp`). Nothing under `$HOME`.

## Not done / open

- TS SDK (`@anthropic-ai/claude-agent-sdk` 0.3.233) exists but was **not exercised** —
  version-checked only. **INFERRED** that it behaves the same, since both drive the same
  `claude` CLI subprocess; unproven.
- Re-query after `interrupt()` on the same client: **UNTESTED**.
- Facade is single-session, in-memory, no auth, no TLS, no SSE streaming, no
  `tasks/cancel`. It proves mechanics; it is not the build.
- The real A2A `tasks/cancel` → `client.interrupt()` binding was proven separately (§2b)
  but not wired into the facade.

---

## Facade source (`/tmp/gcn70/a2a_facade.py`, as run)

```python
#!/usr/bin/env python3
"""Minimal A2A-shaped facade over ONE Claude Agent SDK session. Stdlib + claude_agent_sdk only.

POST /            {"method":"message/send","params":{"message":{"parts":[{"text":...}]}}}
                  -> {"result":{"id":<taskId>,"status":{"state":"submitted"}}}
GET  /tasks/<id>  -> {"result":{"id":..,"status":{"state":"working|input_required|completed|failed"},
                                "artifacts":[{"parts":[{"kind":"text","text":<accumulated>}]}]}}
A pending tool-permission ask becomes state=input_required; the next message/send whose text
starts with "allow" or "deny" answers it (A2A's INPUT_REQUIRED round trip, one session deep).
ponytail: one session, in-memory tasks, no auth/TLS/SSE. Add those when it leaves localhost.
"""
import asyncio, json, sys, threading, uuid
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from claude_agent_sdk import (ClaudeSDKClient, ClaudeAgentOptions, AssistantMessage,
                              TextBlock, ToolUseBlock, ResultMessage,
                              PermissionResultAllow, PermissionResultDeny, HookMatcher)

CWD = sys.argv[1] if len(sys.argv) > 1 else "/tmp/gcn70/f"
PORT = int(sys.argv[2]) if len(sys.argv) > 2 else 9910
TASKS = {}          # taskId -> {"state":..,"text":..,"ask":..}
CURRENT = None      # taskId of the in-flight task
PENDING = None      # asyncio.Future awaiting an allow/deny answer
LOOP = None
CLIENT = None


async def can_use_tool(tool, tool_input, ctx):
    """Block the SDK turn and surface the ask as A2A input_required."""
    global PENDING
    PENDING = LOOP.create_future()
    t = TASKS[CURRENT]
    t["state"], t["ask"] = "input_required", f"approve tool {tool}: {json.dumps(tool_input)[:200]}"
    answer = await PENDING
    PENDING, t["ask"] = None, None
    t["state"] = "working"
    if answer.lower().startswith("allow"):
        return PermissionResultAllow(updated_input=tool_input)
    return PermissionResultDeny(message=answer)


async def ask_hook(input_data, tool_use_id, context):
    """PreToolUse -> "ask" is what actually routes a tool call into can_use_tool headless."""
    return {"hookSpecificOutput": {"hookEventName": "PreToolUse", "permissionDecision": "ask",
                                   "permissionDecisionReason": "A2A facade gate"}}


async def pump():
    """Single reader over the one session; folds messages into task state."""
    async for m in CLIENT.receive_messages():
        t = TASKS.get(CURRENT)
        if t is None:
            continue
        if isinstance(m, AssistantMessage):
            for b in m.content:
                if isinstance(b, TextBlock):
                    t["text"] += b.text
                elif isinstance(b, ToolUseBlock):
                    print("TOOL_USE", b.name, json.dumps(b.input)[:120], flush=True)
        elif isinstance(m, ResultMessage):
            t["state"] = "failed" if m.is_error else "completed"


async def send(text):
    """Either answers a pending ask, or starts a new task on the same session."""
    global CURRENT
    if PENDING is not None and not PENDING.done():
        PENDING.set_result(text)
        return CURRENT
    CURRENT = uuid.uuid4().hex[:12]
    TASKS[CURRENT] = {"state": "working", "text": "", "ask": None}
    await CLIENT.query(text)
    return CURRENT


def part_text(params):
    msg = params.get("message") or {}
    return "".join(p.get("text", "") for p in msg.get("parts", [])) or msg.get("text", "")


class H(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _reply(self, code, body):
        raw = json.dumps(body).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def log_message(self, *a):
        pass

    def do_POST(self):
        req = json.loads(self.rfile.read(int(self.headers["Content-Length"] or 0)) or b"{}")
        if req.get("method") != "message/send":
            return self._reply(404, {"error": {"code": -32601, "message": "method not found"}})
        tid = asyncio.run_coroutine_threadsafe(send(part_text(req.get("params", {}))), LOOP).result(30)
        self._reply(200, {"jsonrpc": "2.0", "id": req.get("id"),
                          "result": {"id": tid, "kind": "task", "status": {"state": "submitted"}}})

    def do_GET(self):
        tid = self.path.rsplit("/", 1)[-1]
        t = TASKS.get(tid)
        if t is None:
            return self._reply(404, {"error": {"code": -32001, "message": "task not found"}})
        st = {"state": t["state"]}
        if t["ask"]:
            st["message"] = {"role": "agent", "parts": [{"kind": "text", "text": t["ask"]}]}
        self._reply(200, {"jsonrpc": "2.0", "result": {
            "id": tid, "kind": "task", "status": st,
            "artifacts": [{"parts": [{"kind": "text", "text": t["text"]}]}]}})


async def main():
    global LOOP, CLIENT
    LOOP = asyncio.get_running_loop()
    opts = ClaudeAgentOptions(
        cwd=CWD, can_use_tool=can_use_tool,  # no allowed_tools: a whole-tool grant shadows it
        hooks={"PreToolUse": [HookMatcher(matcher="Write|Bash", hooks=[ask_hook], timeout=600)]})
    async with ClaudeSDKClient(options=opts) as CLIENT_:
        CLIENT = CLIENT_
        srv = ThreadingHTTPServer(("127.0.0.1", PORT), H)
        threading.Thread(target=srv.serve_forever, daemon=True).start()
        print(f"facade on http://127.0.0.1:{PORT} cwd={CWD}", flush=True)
        await pump()

asyncio.run(main())
```

Drive it:

```bash
python a2a_facade.py /tmp/lane 9912          # one lane, one session
curl -s -X POST http://127.0.0.1:9912/ -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"message/send","params":{"message":{"role":"user","parts":[{"kind":"text","text":"..."}]}}}'
curl -s http://127.0.0.1:9912/tasks/<taskId>
```
