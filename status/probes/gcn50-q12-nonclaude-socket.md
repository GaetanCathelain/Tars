# GCN-50 Q12 — Can a NON-Claude process inject into a running Claude Code session's messaging inbox?

**Date:** 2026-08-14  **Box:** cooper (Linux 5.15, uid 1000 `gaetan`)  **Claude Code:** 2.1.232
**Method:** live system + official docs only (no repo docs, no Slack). Throwaway receivers only.

## VERDICT: NONCLAUDE-INJECT-VERIFIED

A plain non-Claude process (python here; equally bash+socat), NOT a descendant of the
target `claude`, delivered a message straight into a running session's inbox with **no
SendMessage tool and no Claude hop**. The write is authorized by **same-OS-user socket
access only**. The auth token is *optional* and process ancestry is *not required* when
the target has `crossSessionInbound: accept` — which every orc-* / interactive Orca
session on cooper sets. `ssh cooper '<cmd>'` runs as uid 1000 and therefore qualifies.

---

## 1. Docs — the wire frame (quoted)

From https://code.claude.com/docs/en/cross-session-messaging (§"The session's inbox socket" / "own-child messages"):

> Alongside the path, Claude Code exports a per-session token as `CLAUDE_CODE_MESSAGING_TOKEN`.
> A script posting to its own session's socket can send `{"type":"auth","token":"<token>"}`
> as the first line of its connection.

> **Own-child messages**: when no `crossSessionInbound` value applies, Claude Code delivers a
> message it verifies came from the session's own child processes … On Linux … Claude Code can
> verify by process evidence even for a child that has already exited … Where that process
> evidence is missing … Claude Code verifies a child that sent the session's exported
> `CLAUDE_CODE_MESSAGING_TOKEN` as its first-line auth frame. When Claude Code can verify
> neither way, it treats the message like any other that asserts no permission class, so a
> session that bypasses permission prompts holds it for your approval.

`crossSessionInbound` values (§"Control inbound messages"): `accept` = deliver each message;
`hold` = show a notice, don't deliver; `refuse` = drop.

### The exact frame — pulled from the claude binary's own debug log (2.1.232)

The uds-messaging listener log literally embeds the injection recipe:

```
[uds-messaging] Inject messages (auth line ${ob.authRequired?"REQUIRED":"optional"} here):
{ echo '{"type":"auth","token":"'"$CLAUDE_CODE_MESSAGING_TOKEN"'"}';
  echo '{"type":"user","message":{"role":"user","content":"hello"}}'; } | socat - UNIX-CONNECT:<sock>
```

**Wire = newline-delimited JSON.** Optional first line `{"type":"auth","token":"<KEY>"}`,
then the message line `{"type":"user","message":{"role":"user","content":"<text>"}}`.
`ob.authRequired` is **false** on this build (the auth line is "optional") — proven
empirically by variation B2 below (delivered with no auth line at all).

## 2. Registry / socket facts (measured)

Each messaging session writes:
- `~/.claude/sessions/<pid>.json` — mode `-rw-rw-r--`, holds `messagingSocketPath`, `sessionId`, `procStart`.
- `~/.claude/sessions/<pid>.<sha>.key` — mode `-rw-------`, 71 bytes, JSON `{"peerToken":<32 hex>,"procStart":<8>}`.
- the socket `/run/user/1000/cc-socks/<pid>.sock` — mode `srw-------`.

Both `~/.claude/sessions/` and `/run/user/1000/cc-socks/` are `drwx------` (owner-only).
**Consequence: reaching a socket requires the same OS user (uid 1000). A different user cannot.**
`CLAUDE_CODE_MESSAGING_SOCKET`/`_TOKEN` in a session's `/proc/environ` are the *inherited
parent* values; the session binds its own socket at its own pid and writes its own key file.

## 3. Throwaway receiver (RECEIVER-1)

Spawned under a pty with a scrubbed env (`env -u CLAUDE_CODE_*`) so it is a top-level
session, not a child session, in `…/scratchpad/q12-socket-probe/`:

- `claude --model 'sonnet[1m]' --settings '{"crossSessionInbound":"accept"}' --dangerously-skip-permissions '<passive receiver prompt>'`
- Registered: **pid 4039191**, socket `/run/user/1000/cc-socks/4039191.sock`, key file
  `4039191.<sha>.key` (peerToken 32B + procStart 8B), status idle.
- (First attempt inherited `CLAUDE_CODE_CHILD_SESSION` from the launching agent's claude and
  did NOT register as an independent peer — hence the env scrub. Documented confound.)

## 4. Sender = plain python, NOT a descendant of the receiver (proven)

`sender.py` run via `env -i PATH=… HOME=… python3 …` (models an ssh-arriving command).
Every run captured the sender's full ppid chain; the receiver pid never appears in it:

```
sender ppid_chain: python3 → zsh → claude(3588220 = MY agent session) → zsh → orca-ide(x3) → bash → systemd → 1
receiver_pid_in_chain: false    (receiver = 4039191, absent from chain)
```

## 5. Controlled variations (each labeled; artifacts in scratchpad/q12-socket-probe/)

| # | receiver | inbound | sender | token | RESULT |
|---|----------|---------|--------|-------|--------|
| A | 4039191 | accept | non-child python | **correct** peerToken | **DELIVERED** — received.txt written, nonce in transcript ×3 |
| B | 4039191 | accept | non-child python | **wrong** (deadbeef×4) | **DELIVERED** — received.txt written, transcript ×3 |
| B2 | 4039191 | accept | non-child python | **omitted** (no auth line) | **DELIVERED** — received.txt written, transcript ×3 |
| C | 4044597 | **default** | non-child python | omitted | **HELD, not delivered** — received.txt empty; hold dialog shown |

Frames sent (varA.json), token redacted:
```
{"type": "auth", "token": "<KEY>"}
{"type": "user", "message": {"role": "user", "content": "SOCKET-NONCE-dd4bded3"}}
```
Server sent **0 bytes back** in every case (fire-and-forget; no ack on the wire).

**Isolation of the gate:** A→B→B2 hold the sender (non-child) and receiver (accept) fixed and
vary only the token: correct / wrong / absent all DELIVER. ⇒ **the token is not the gate and
ancestry is not the gate under `accept`.** C holds the sender fixed and flips only the
receiver's inbound config to default ⇒ message is HELD. ⇒ **the flipping variable is
`crossSessionInbound` (accept vs default), plus the same-uid socket permission.**

Variation C receiver logged verbatim:
> Held peer message from an unidentified session **[verified pid 4046332]**; … not delivered
> to Claude (held). The sender did not attest its permission mode and this session bypasses
> prompts. Review it below, or set "crossSessionInbound" to "accept".

So Claude Code reads the connecting process's pid via SO_PEERCRED, finds it is not an
own-child ("unidentified"), and — because this default receiver bypasses permission prompts —
**holds it for the human**. It is neither silently delivered nor dropped; a person must click
Deliver. (Variation c "sender-as-child" was unnecessary: the non-child sender already
succeeded under accept, so ancestry was never the blocker.)

## 6. The real target scenario: `ssh cooper '<cmd>'`

`ssh cooper '<cmd>'` runs as **gaetan / uid 1000** — the same user that owns every session
socket. It is a plain process, not a descendant of any target `claude`. It can:
1. read `~/.claude/sessions/<pid>.json` to get `messagingSocketPath` (same-uid traverses the 0700 dir),
2. `connect()` the `srw-------` socket (same-uid),
3. `printf '{"type":"user","message":{"role":"user","content":"…"}}\n' | socat - UNIX-CONNECT:<sock>` — **no auth token, no Claude process on the sending side.**

Outcome is set entirely by the TARGET session's config:
- **Target set `crossSessionInbound: accept`** → injected **silently** as a user turn. **This
  is the state of every orc-sonnet/opus/fable session and every interactive `claude` typed in
  an Orca tab** (the `claude()` shell function sets accept). ⇒ ssh-injection succeeds unattended.
- **Target is bypass-perms without accept** → **held** for Gaetan's manual approval (variation C).
- **Target is an ordinary prompting session without accept** → delivered (docs: asserts no
  permission class; not separately tested here).

**Bottom line for Q12:** yes — a remote agent over ssh (as gaetan) can deliver into a running
session with no Claude hop and no SendMessage, and against cooper's accept-configured
orchestrator sessions it lands silently. The only hard requirement is **same OS user**;
uid-separation (a different user, or a container with its own filesystem) is the real
boundary, not any key or ancestry check.

## Artifacts (scratchpad/q12-socket-probe/)
sender.py, varA.json, varB.json, varB2.json, varC.json, nonce_*.txt,
received.txt (RECEIVER-1), recv2/received.txt (RECEIVER-2, empty = held), receiver.log(s).
All spawned processes killed after the run.
