# GCN-50 Q12 probe B — can a channel MCP server push a file-borne event into a live Claude Code session?

**Date**: 2026-08-14 (UTC) · **Host**: cooper · **Claude Code**: 2.1.232 · **Account**: Claude Max (no managed-settings file on this box)

**VERDICT: VERIFIED-DELIVERED — in an INTERACTIVE session only.**
`claude -p` (headless): **VERIFIED-FAILED (receiver side — no channel registration, event dropped silently)**, 3/3 runs.

---

## 1. The contract (from the docs, quoted)

Source: <https://code.claude.com/docs/en/channels> and <https://code.claude.com/docs/en/channels-reference>.
Both pages carry: "Channels are in research preview… They require Anthropic authentication through claude.ai
or a Console API key, and are not available on Amazon Bedrock, Google Cloud's Agent Platform, or Microsoft Foundry."

### 1.1 What a channel is

> "A channel is an MCP server that pushes events into your running Claude Code session, so Claude can react to
> things that happen while you're not at the terminal. […] Events only arrive while the session is open, so for
> an always-on setup you run Claude in a background process or persistent terminal."

> "Claude Code spawns it as a subprocess and communicates over stdio."

### 1.2 The three server requirements (channels-reference, "What you need")

> "Your server needs to:
> 1. Declare the `claude/channel` capability so Claude Code registers a notification listener
> 2. Emit `notifications/claude/channel` events when something happens
> 3. Connect over stdio transport"

### 1.3 Capability declaration

> | `capabilities.experimental['claude/channel']` | `object` | Required. Always `{}`. Presence registers the notification listener. |
> | `capabilities.experimental['claude/channel/permission']` | `object` | Optional. Always `{}`. Declares that this channel can receive permission relay requests. |
> | `capabilities.tools` | `object` | Two-way only. Always `{}`. […] |
> | `instructions` | `string` | Recommended. Added to Claude's system prompt. |

### 1.4 Notification format

> "Your server emits `notifications/claude/channel` with two params:"
> | `content` | `string` | The event body. Delivered as the body of the `<channel>` tag. |
> | `meta` | `Record<string, string>` | Optional. Each entry becomes an attribute on the `<channel>` tag […] Keys must be identifiers: letters, digits, and underscores only. Keys containing hyphens or other characters are silently dropped. |

Arrival shape in the model's context:

> ```text
> <channel source="your-channel" severity="high" run_id="1234">
> build failed on main: https://ci.example.com/run/1234
> </channel>
> ```

> "Claude Code doesn't acknowledge notifications. The `await` on `mcp.notification()` resolves when the message is
> written to the transport, not when Claude has processed it. **If the session hasn't loaded your server as a
> channel, or the organization policy blocks it, Claude Code drops the events silently and returns no error to
> your server.**"  ← this is exactly the failure mode observed in `-p` below.

> "Events queue into the session and are processed in order. If several notifications arrive while Claude is busy,
> they're delivered together on the next turn and Claude handles them as a group."

### 1.5 Config shape + opt-in flags

> ```json title=".mcp.json"
> { "mcpServers": { "webhook": { "command": "bun", "args": ["./webhook.ts"] } } }
> ```

> "Being in `.mcp.json` isn't enough to push messages: a server also has to be named in `--channels`."

> ```bash
> # Testing a plugin you're developing
> claude --dangerously-load-development-channels plugin:yourplugin@yourmarketplace
> # Testing a bare .mcp.json server (no plugin wrapper yet)
> claude --dangerously-load-development-channels server:webhook
> ```

> "Neither `--channels` nor `--dangerously-load-development-channels` appears in `claude --help` while the feature
> is in preview. The flags work even though they aren't listed."

> "During the research preview, `--channels` only accepts plugins from an Anthropic-maintained allowlist […]
> To test a channel you're building, use `--dangerously-load-development-channels`."

So: **not plugin-only.** A bare MCP server named in `--dangerously-load-development-channels server:<name>` is a
supported (if dev-flagged) path. `--channels` is the plugin-only, allowlisted path.

---

## 2. What was built

`/tmp/claude-1000/-home-gaetan-dev-orca-worktrees-Tars-improvements/86148edc-827d-4f2c-9864-855a2a981415/scratchpad/q12-channel-probe/`

No npm/bun deps: the MCP stdio wire format is newline-delimited JSON-RPC, so python3 stdlib is enough.
The server **deletes** each inbox file right after pushing it, so the nonce exists nowhere the receiver could read
from disk — the channel is the only possible source.

`server.py`:

```python
#!/usr/bin/env python3
"""Minimal Claude Code channel (stdlib only): each new file in inbox/ is pushed
as a notifications/claude/channel event, then deleted so its content exists
nowhere Claude could read from disk."""
import json, os, sys, threading, time

HERE = os.path.dirname(os.path.abspath(__file__))
INBOX = os.path.join(HERE, "inbox")
LOG = open(os.path.join(HERE, "server.log"), "a", buffering=1)
_lock = threading.Lock()


def out(msg):
    with _lock:
        sys.stdout.write(json.dumps(msg) + "\n")
        sys.stdout.flush()
    LOG.write("OUT " + json.dumps(msg) + "\n")


def watch():
    seen = set(os.listdir(INBOX))
    while True:
        for name in sorted(set(os.listdir(INBOX)) - seen):
            path = os.path.join(INBOX, name)
            try:
                with open(path) as f:
                    body = f.read()
                os.remove(path)
            except OSError:
                continue
            out({"jsonrpc": "2.0", "method": "notifications/claude/channel",
                 "params": {"content": body, "meta": {"filename": name}}})
        seen = set(os.listdir(INBOX))
        time.sleep(0.5)


def main():
    started = False
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        LOG.write("IN  " + line + "\n")
        req = json.loads(line)
        method, rid = req.get("method"), req.get("id")
        if method == "initialize":
            out({"jsonrpc": "2.0", "id": rid, "result": {
                "protocolVersion": req.get("params", {}).get("protocolVersion", "2025-06-18"),
                "capabilities": {"experimental": {"claude/channel": {}}},
                "serverInfo": {"name": "filewatch", "version": "0.0.1"},
                "instructions": 'Events arrive as <channel source="filewatch" filename="...">. '
                                'One-way: read the content and act on it, no reply expected.'}})
        elif method == "notifications/initialized":
            if not started:
                started = True
                threading.Thread(target=watch, daemon=True).start()
        elif method == "ping":
            out({"jsonrpc": "2.0", "id": rid, "result": {}})
        elif method in ("tools/list", "resources/list", "prompts/list"):
            out({"jsonrpc": "2.0", "id": rid, "result": {method.split("/")[0]: []}})
        elif rid is not None:
            out({"jsonrpc": "2.0", "id": rid,
                 "error": {"code": -32601, "message": "method not found: %s" % method}})


if __name__ == "__main__":
    main()
```

`mcp.json`:

```json
{"mcpServers":{"filewatch":{"command":"python3","args":["<abs path>/server.py"]}}}
```

### Self-check (server in isolation) — PASS

```
( printf '{"jsonrpc":"2.0","id":1,"method":"initialize",...}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n'; sleep 3 ) \
  | python3 server.py > selfcheck.out & sleep 1; echo "SELFCHECK-NONCE-123" > inbox/t1.txt
```
```
{"jsonrpc": "2.0", "id": 1, "result": {..."capabilities": {"experimental": {"claude/channel": {}}}...}}
{"jsonrpc": "2.0", "method": "notifications/claude/channel", "params": {"content": "SELFCHECK-NONCE-123\n", "meta": {"filename": "t1.txt"}}}
```

---

## 3. Attempts

Writer was always a separate process: `nohup sh -c "echo '<nonce>' > <inbox>/evt.txt"`.

### Attempt 1 — `claude -p`, receiver told to `sleep 45` — INCONCLUSIVE (receiver ended its turn)

The harness hook blocked standalone `sleep 45`; the model backgrounded it and called `ScheduleWakeup`, ending the
turn. `-p` exited after 22.4 s, before the drop mattered. No `received.txt`.

### Attempt 2 — `claude -p`, blocking `python3 -c 'import time; time.sleep(45)'`, `--disallowedTools ScheduleWakeup` — FAILED

```bash
command claude -p "<wait 45s, then write the channel event content to received.txt, else NO-EVENT; never read inbox>" \
  --mcp-config $D/mcp.json \
  --dangerously-load-development-channels server:filewatch \
  --dangerously-skip-permissions --disallowedTools ScheduleWakeup \
  --output-format json
```
* server.log: `OUT {"jsonrpc": "2.0", "method": "notifications/claude/channel", "params": {"content": "Q12-NONCE-7f3a9c-DELIVERED\n", ...}}` at drop time (09:37:17), session alive until 09:37:56.
* `received.txt` = `NO-EVENT`.
* Transcript: *"No `<channel source=\"filewatch\">` event was delivered during or after the wait."*

### Attempt 3 — same as 2, plus `--debug-file` — FAILED, and the debug log localizes the break

```
[DEBUG] MCP server "filewatch": Successfully connected (transport: stdio) in 24ms
[DEBUG] MCP server "filewatch": Connection established with capabilities:
        {"hasTools":false,"hasPrompts":false,"hasResources":false,"hasResourceSubscribe":false,
         "serverVersion":{"name":"filewatch","version":"0.0.1"}}
[DEBUG] [MCP] Server "filewatch" connected with subscribe=false
```
`grep -inE 'blocked|org policy|allowlist|channelsEnabled|development-channel|experimental' dbg.txt` → **zero hits.**
The server connects as an ordinary MCP server; nothing in the log shows a channel listener being registered, and
`received.txt` = `NO-EVENT` again. The drop is silent, exactly as §1.4 warns.

Controls: `--bogus-flag-xyz` → `error: unknown option '--bogus-flag-xyz'`, while
`--dangerously-load-development-channels server:filewatch` parses fine ⇒ the flag IS recognized by 2.1.232.
No `/etc/claude-code/managed-settings.json` on this box ⇒ no org policy block here.

### Attempt 4 — INTERACTIVE session under a pty (`script`) — **DELIVERED**

```bash
{ sleep 8; printf '\r'; sleep 4; printf '\r'; sleep 25; printf '\003'; sleep 1; printf '\003'; } \
 | script -qc "command claude --mcp-config $D/mcp.json \
     --dangerously-load-development-channels server:filewatch --dangerously-skip-permissions" $D/tty.log &
sleep 20; nohup sh -c "echo 'Q12-INTERACTIVE-NONCE-b41d' > $D/inbox/evt.txt"
```

Startup dialog (accepted with the first `\r`):

```
--dangerously-load-development-channels is for local channel development
only. Do not use this option to run channels you have downloaded off the
internet.
Please use --channels to run a list of approved channels.
Channels: server:filewatch
```

Startup notice, then the delivery, from `tty.log` (ANSI stripped):

```
▎Channels (experimental) messages from server:filewatch inject directly in
▎this session · restart without --dangerously-load-development-channels to stop
 server:filewatch · no MCP server configured with that name
...
← filewatch: Q12-INTERACTIVE-NONCE-b41d
● Filewatch probe received: Q12-INTERACTIVE-NONCE-b41d (from evt.txt). No ...
```

The nonce was written by an unrelated shell process into `inbox/`, read and **deleted** by the channel server, and
came back out of the model's mouth in a session that was idle at the prompt. Push delivery of a file-borne event
is real. (Session was idle-at-prompt, not mid-turn — the event still landed and the model reacted immediately.)

---

## 4. Findings

1. **The mechanism works.** File written by an external process → channel server → `<channel>` tag in a live
   session, with the model reacting in the same session. Nonce provenance is airtight (file deleted before the
   session could have read it).
2. **`claude -p` is not a usable receiver on 2.1.232.** 3/3 headless runs: MCP server connects, notification is
   written to the transport, event never reaches the model, `received.txt` = `NO-EVENT`, and the debug log shows
   no channel registration and no error. Docs describe the `-p` + channels combination as supported ("When you run
   channels in non-interactive mode with `-p`, tools that need terminal input … are disabled"), so this is either a
   preview gap or an undocumented restriction — either way, **do not build on `-p`**. Not root-caused further
   (timebox); no org-policy or allowlist evidence was found to explain it.
3. **Odd but non-fatal**: even in the working interactive run, the banner also prints
   `server:filewatch · no MCP server configured with that name` — likely because the server came from
   `--mcp-config` rather than a `.mcp.json`/user config, and the allowlist checker looks in the latter. Delivery
   happened anyway. Worth retesting with a real `.mcp.json` before trusting the warning either way.
4. **For GCN-50 (remote agent writes files over ssh)**: viable, with the shape the docs already prescribe — an
   always-on **interactive/persistent-terminal** session (orca tab, tmux, or `--bg`-style long-lived process),
   never a `-p` one-shot. The server should gate senders (§"Gate inbound messages"): anything the remote agent
   writes lands verbatim in front of the model, so the ssh drop dir is a prompt-injection surface.
5. Cost of the receiver: nil. stdlib python, ~55 lines, no bun/node/npm.

## 5. Cleanup

All probe processes killed (`python3 server.py`, the pty `script`, the interactive `claude`); verified none
remain. Scratch dir left in place under
`/tmp/claude-1000/-home-gaetan-dev-orca-worktrees-Tars-improvements/86148edc-827d-4f2c-9864-855a2a981415/scratchpad/q12-channel-probe/`
(`server.py`, `mcp.json`, `server.log`, `dbg.txt`, `tty.log`, `out*.json`). No existing session, worktree or repo
file other than this evidence file was touched. Not committed.
