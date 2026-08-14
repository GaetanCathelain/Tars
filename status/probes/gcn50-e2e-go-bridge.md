# GCN-50 — Go bidirectional peer to Claude Code sessions (native, no `claude -p` spawn)

**Date:** 2026-08-14  **Box:** cooper (Linux 5.15, uid 1000 `gaetan`)  **Claude Code:** 2.1.232
**Question:** can a plain **Go** process be a *first-class cross-session peer* — SEND into a
running Claude session's inbox AND be discovered + messaged BACK by that session's own
`SendMessage` tool — so Tars/Hermes could talk to Claude sessions from one resident process,
no `claude -p` relay, no ssh-chat, no human in the loop?

## VERDICT: **GO-BIDIRECTIONAL-PEER-PROVEN**

- **SEND (Go → session): WORKS.** Go wrote the auth+user NDJSON frames to the target session's
  socket; the session rendered it as a real inbound peer turn ("Another Claude session sent a
  message: [E2E gcn50-go] ping 7f3a91"). Direct port of the proven python injector.
- **RECEIVE (session → Go): WORKS.** The Go process published a `sessions/<pid>.json` + `.key`
  registration; the session's `ListAgents` discovered it as `go-bridge-peer`, its `SendMessage`
  connected to the Go socket and delivered — Go logged 3 real SendMessage frames. This is the
  novel part the prior Go attempt never reached (it died at the trust dialog).
- **Round-trip: WORKS.** Go asked "reply with the single word BRIDGE"; the session `SendMessage`d
  `BRIDGE` back to `go-bridge-peer`; Go printed it. Full Tars↔session role, both wires, one
  resident process, autonomous.

A resident Go peer is **viable** as the Tars↔session transport. It requires only same-uid
filesystem access; no Claude process on the Go side. Top risk: the whole thing rides an
**undocumented, version-pinned** registration + socket contract (details + fragility below).

---

## 1. Reverse-engineered registration format (Claude Code 2.1.232)

Confirmed firsthand against a live session (`96610.json`) and reproduced by the Go writer.

**Discovery record** `~/.claude/sessions/<pid>.json` (dir mode `0700`, file `0664`). The Go peer
wrote this verbatim and was discovered:
```json
{"pid":158026,"sessionId":"733f80c1-5022-46da-8dd7-ce6121e3d8de","cwd":"/run/user/1000/gobridge-158026",
 "startedAt":1786723266339,"procStart":"75099360","version":"2.1.232","peerProtocol":1,
 "kind":"interactive","entrypoint":"cli","messagingSocketPath":"/run/user/1000/gobridge-158026/peer.sock",
 "name":"go-bridge-peer","nameSource":"derived","nameSince":1786723266339,
 "status":"idle","updatedAt":1786723266339,"statusUpdatedAt":1786723266339}
```
Field notes (what actually mattered for discovery + liveness):
- `pid` — the **Go process's real pid**. Claude cross-checks liveness against it.
- `procStart` — `/proc/<pid>/stat` **field 22** (starttime). Parse after the *final* `)` (comm can
  contain spaces/parens), then take index 19 of the remaining fields. **Must equal** the live
  value or a discoverer treats the record as a dead/reused pid and skips it.
- `messagingSocketPath` — absolute path to the Go peer's own listening socket. Can live **anywhere
  same-uid**; it need not be under `cc-socks/`. Used a short private `/run/user/1000/gobridge-<pid>/`
  dir (unix socket path length limit).
- `name` — the address `SendMessage to:` resolves. Set it to what the session should call.
- `peerProtocol:1`, `version:"2.1.232"` — compatibility gate; mismatch is the likely breakage on
  upgrade.
- `kind:"interactive"`, `entrypoint:"cli"`, `status:"idle"`, the three `*At`/`*Since` epoch-**ms**
  timestamps — included to match the live shape; `status:"idle"` presents as reachable.
- `startedAt`/`nameSince`/`updatedAt`/`statusUpdatedAt` are Unix epoch **milliseconds**.

**Key file** `~/.claude/sessions/<pid>.<sha256hex(socketpath)>.key`, mode `0600`, 71 bytes:
```json
{"peerToken":"<32 lowercase hex>","procStart":"<same as record>"}
```
- Filename hash = `sha256(messagingSocketPath bytes)` hex. **Verified**: computed name matched the
  live session's actual key file exactly (`exists: True`).
- `peerToken` is what a *sender* (including Claude) reads and presents in its auth frame when
  delivering to this peer. Generated with `crypto/rand`; never logged.

**Publish order** (matches the local-ipc-probes skill invariant): bind socket → write key `0600`
→ write record last via `tmp`+`rename` (atomic; no discovery of a half-built peer).

## 2. Wire protocol (NDJSON over the unix socket), both directions observed

**SEND (Go as sender), frames written (token redacted):**
```
{"type":"auth","token":"<KEY>"}
{"type":"user","message":{"role":"user","content":"[E2E gcn50-go] ping 7f3a91"}}
```
Server sends **0 bytes back** (fire-and-forget; no wire ack). Auth is *optional* under the
receiver's `accept` mode (proven separately in gcn50-q12-nonclaude-socket.md) but the Go sender
mirrors the key handshake for a faithful interop test.

**RECEIVE (Go as receiver), a real `SendMessage` frame the session delivered (redacted, verbatim
from recv.log):**
```json
{"from":"uds:/run/user/1000/cc-socks/158634.sock","type":"user","msgV":1,
 "msg_id":"6c52dff7-6073-41f2-a0df-7497d510e495","priority":"next",
 "message":{"role":"user","content":
   "<cross-session-message from=\"uds:/run/user/1000/cc-socks/158634.sock\" hop-chain=\"d4b746ae0ced483d6365c9a7\" from-name=\"probe-cwd-6b\" from-mode=\"bypass\">\n[E2E gcn50-go] pong 9b2c4d\n</cross-session-message>"}}
```
An auth frame with the Go peer's own `peerToken` arrives first (Go validated `matches=true`,
value never printed). Key learnings for the receiver:
- Claude wraps the payload in a **`<cross-session-message …>` envelope** carrying `from` (uds
  path), `hop-chain`, `from-name`, `from-mode`. A real Tars-side receiver would strip this
  envelope to recover the human text.
- Extra top-level fields vs. the SEND minimum: `msgV`, `msg_id` (UUID), `priority` ("next"),
  `from`. PASS predicate held: `type=="user" && message.role=="user"`.
- SO_PEERCRED on the accepted conn showed `pid=158634 uid=1000` = the session's pid — a receiver
  can attribute the sender at the kernel level.

## 3. Live test log (nonces prove causation)

| Step | Action | Evidence |
|---|---|---|
| setup | `go-bridge serve --name go-bridge-peer` resident (pid 158026), registered | recv.log `SERVE`/`READY`; json+key on disk |
| session | `claude-probe-go` spawned in tmux, env-scrubbed, `crossSessionInbound:accept`, `--dangerously-skip-permissions`, trust dialog accepted | registered pid 158634 name `probe-cwd-6b`, idle |
| **SEND** | Go `send` → session socket: `[E2E gcn50-go] ping 7f3a91` | session pane: *"Another Claude session sent a message: [E2E gcn50-go] ping 7f3a91"* |
| **RECEIVE** | session `SendMessage to go-bridge-peer` (triggered by a 2nd Go send) | recv.log PASS `pong 9b2c4d`; pane: *"→ sent to go-bridge-peer … ● Sent."* |
| **round-trip** | Go asks "reply with BRIDGE"; session answers via SendMessage | recv.log PASS `BRIDGE` |

recv.log captured **3 PASS** SendMessage frames from the session (`pong 7f3a91`, `pong 9b2c4d`,
`BRIDGE`), each with a valid auth handshake. No message was ever held or refused — the Go peer is
indistinguishable from a Claude peer to the sending session (it even labels it "another Claude
session on this machine").

## 4. Wire-1 startup gate handled

Fresh cwd under `$HOME` → Claude's **workspace-trust dialog** fired before the session published
its record (as the reference warned; this is exactly what stopped the prior Go attempt). Handled
by launching in tmux and sending `Enter` on "1. Yes, I trust this folder". Also scrubbed the
parent-session vars (`CLAUDE_CODE_MESSAGING_SOCKET/_TOKEN`, `CLAUDE_CODE_SESSION_ID`,
`CLAUDE_CODE_CHILD_SESSION`, `CLAUDECODE`, `CLAUDE_CODE_ENTRYPOINT`, `CLAUDE_CODE_EXECPATH`,
`CLAUDE_PID`, `CLAUDE_EFFORT`) so the child registered as an **independent top-level peer**, not a
nested child session.

## 5. Is a resident Go peer viable as the Tars↔session transport? — YES, with caveats

**What it buys:** one long-lived Go process holds *both* wires. To push to a session: connect to
its `cc-socks/<pid>.sock`, write 2 lines. To be reachable: keep a `sessions/<pid>.json`+`.key`
published and accept on a socket. **No `claude -p` spawn per message** (the current Wire-3 relay
costs ~6.3k tokens + ~6.8 s each — see gcn50-e2e-roundtrip.md); the Go path is a socket write,
~0 tokens, sub-ms. Natively models Tars as a peer in `ListAgents`.

**Hard requirements:**
1. **Same OS user (uid 1000).** `sessions/` and `cc-socks/` are `0700`. This is the real
   boundary — a different uid or a container with its own fs cannot reach in. Tars runs on its own
   VM as user `gaetan`? No — Tars is a *different* machine. So this transport is **same-box only**:
   it works for a Go process co-located with the Claude sessions (e.g. an orchestrator daemon on
   cooper), **not** for the Tars VM reaching cooper. Cross-machine still needs Remote Control or an
   ssh hop that lands as uid 1000 on the box that hosts the sessions.
2. **Receiver posture:** to deliver *into* a session unattended, that session must be
   `crossSessionInbound:accept` (every orc-* / interactive-Orca session is). Otherwise the message
   is *held* for human approval (variation C, gcn50-q12).
3. **Correct registration:** real pid + matching `procStart` + socket path + `peerProtocol:1` +
   sha256-named key. Liveness is re-checked by discoverers on every `ListAgents`.

**Robustness / effort:** the receiver must strip the `<cross-session-message>` envelope, buffer
partial NDJSON reads, tolerate reconnects (each SendMessage is a fresh short-lived connection),
and self-clean its record/key/socket on exit. All demonstrated in ~150 lines of stdlib Go.

## 6. TOP VERSION-FRAGILITY RISK (flagged explicitly)

**The entire transport depends on an undocumented, reverse-engineered contract with no API
stability guarantee.** Concretely, any Claude Code upgrade can silently break it by changing:
`peerProtocol` (currently `1` — the single most likely gate to bump), the record schema/required
fields, the key-filename hashing input, the `procStart` liveness check, the socket framing, or the
`cc-socks`/`sessions` paths. Failure mode is **silent**: the Go peer simply stops appearing in
`ListAgents` or its sends stop landing — no error, no log. A prior Go port already failed to
complete once (trust dialog). **Mitigation if productionized:** pin/measure `version`+`peerProtocol`
on the box, re-probe both wires with a nonce self-test on every Claude upgrade (this file is that
test), and fail loud when the schema drifts. Treat it as an internal integration to re-validate
per upgrade, never a stable interface.

## 7. Go source (final, built with go 1.26.5)

Single file, stdlib only. `serve` (register+receive), `send` (inject), `find` (locate a live
session by name). See inline — the load-bearing pieces are `procStart` (field-22 parse),
`keyfilePath` (sha256 of socket path), the publish order in `serve`, and the
`<cross-session-message>`-carrying frame the `handle` loop decodes.

```go
// procStart returns /proc/<pid>/stat field 22 (starttime), parsed after the final ')'.
func procStart(pid int) string {
	data, _ := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	s := string(data)
	rp := strings.LastIndex(s, ")")
	rest := strings.Fields(s[rp+2:])
	return rest[19] // field3=state is rest[0]; field22 = rest[19]
}
func keyfilePath(sessDir string, pid int, sock string) string {
	h := sha256.Sum256([]byte(sock))
	return filepath.Join(sessDir, fmt.Sprintf("%d.%s.key", pid, hex.EncodeToString(h[:])))
}
// serve: bind socket -> write key(0600) -> publish record(tmp+rename) last; accept NDJSON; on
//        SIGTERM/timeout remove record+key+socket+rundir.
// send:  dial socket; write {"type":"auth","token":<peerToken from key>}\n then
//        {"type":"user","message":{"role":"user","content":<msg>}}\n ; read 0 bytes back.
// find:  scan sessions/*.json for name match with live pid && matching procStart.
```
Full source: `scratchpad/gcn50-go-bridge/bridge.go` (session scratchpad; not committed). It matches
the two registration/framing recon points verified live in §1–§2.

## 8. Self-clean (verified)

- go-bridge serve SIGTERM'd → `sessions/<pid>.json` ABSENT, `.key` ABSENT, `/run/user/1000/gobridge-<pid>/`
  ABSENT, proc dead (self-cleans in its own shutdown path — a race where `main` returned before the
  cleanup goroutine finished was found and fixed; cleanup now runs on the main goroutine).
- `claude-probe-go` TUI exited → its `sessions/158634.json` ABSENT, proc dead.
- tmux `probego` killed; final sweep: 0 stray rundirs, 0 gobridge/probe session jsons, 0 stray
  processes, no `.tmp` files. No worktree or git scratch repo was created (tmux + a plain dir).

## Artifacts (session scratchpad, NOT committed)
`scratchpad/gcn50-go-bridge/`: `bridge.go`, `bridge` (binary), `recv.log` (3 PASS frames),
`serve.stdout/err`, `send-nonce.txt` (7f3a91), `recv-nonce.txt` (9b2c4d). Redaction held: no
`peerToken` value appears in any evidence.
