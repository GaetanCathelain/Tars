# Claude Code 2.1.232 local cross-session messaging

Version-pinned reverse-engineering notes for constructing a disposable non-Claude Linux peer. These internals are not a stable API; re-check the native bundle and native session files on other versions.

## Discovery record

Claude discovers local peers from:

```text
<Claude config>/sessions/<pid>.json
```

Observed record fields:

```json
{
  "pid": 12345,
  "sessionId": "UUID",
  "cwd": "/absolute/path",
  "startedAt": 0,
  "procStart": "linux-proc-start-value",
  "version": "2.1.232",
  "peerProtocol": 1,
  "kind": "interactive",
  "name": "peer-name",
  "nameSince": 0,
  "messagingSocketPath": "/absolute/private/peer.sock"
}
```

For the harness, `startedAt` and `nameSince` were emitted as Unix epoch milliseconds. `procStart` came from Linux `/proc/<pid>/stat` field 22; parse after the final `)` because the process name field can contain spaces or parentheses.

## Key file

The sibling key path is:

```text
<Claude config>/sessions/<pid>.<sha256(canonical-socket-path)>.key
```

Its JSON shape is:

```json
{"peerToken":"<32 lowercase hex>","procStart":"<same process-start value>"}
```

Generate the token securely, write the key as owner-only, never log it, and publish the session record only after both key and listening socket exist.

## Socket framing

The sender connects to `messagingSocketPath` and writes newline-delimited JSON. When a peer key is published, it sends an auth frame first; redact the token before persisting any frame log and validate it inside the receiver without printing it. An explicitly configured sender socket was observed logging that auth was optional for that inbox, but that does **not** establish that arbitrary discovered peer inboxes may omit authentication. Mirror the key-file handshake for an interoperability test.

A user delivery resembles:

```json
{
  "msgV": 1,
  "msg_id": "UUID",
  "type": "user",
  "message": {"role": "user", "content": "text"},
  "priority": "next",
  "from": "uds:..."
}
```

A useful PASS predicate is strictly:

```text
frame.type == "user" AND frame.message.role == "user"
```

Require and validate the auth frame when the harness publishes a key file, while ensuring the token value is redacted from all evidence.

## Sender-mode observation

A live installed Claude Code 2.1.232 probe showed that cross-session peer discovery was available in an **interactive TTY session**: the built-in agent-listing flow displayed 11 local peer sessions. A `claude -p` print-mode probe in the same workspace reported that no `ListPeers` tool was available. For decisive tests, use an interactive PTY/tmux session and verify the actual tool list there rather than assuming print mode has the same cross-session capabilities.

The CLI accepted a hidden `--messaging-socket-path <absolute-path>` option in the live interactive probe. This is useful for containing the sender's own socket inside a scratch directory, but it is version-pinned and must be re-probed on upgrades.

### Launching an independent Claude from a Claude/Orca parent

A subprocess launched from inside a Claude/Orca terminal can inherit parent-session identity and messaging variables such as `CLAUDECODE`, `CLAUDE_CODE_CHILD_SESSION`, `CLAUDE_CODE_ENTRYPOINT`, `CLAUDE_CODE_EXECPATH`, `CLAUDE_CODE_MESSAGING_SOCKET`, `CLAUDE_CODE_MESSAGING_TOKEN`, `CLAUDE_CODE_SESSION_ID`, and `CLAUDE_PID`. With those inherited, a nominally interactive 2.1.232 launch was treated as nested/non-interactive and exited with the print-mode input error instead of publishing a session record.

For a genuinely independent live probe, remove only those parent-session variables from the child environment before invoking Claude. Do not dump the environment while diagnosing this: it can contain the live messaging token. After scrubbing, a fresh workspace reached the real interactive workspace-trust dialog, confirming that the nested-session contamination was gone.

A fresh working directory may show that trust dialog **before** Claude publishes its session record. A harness that only polls the registry will therefore time out opaquely. Watch the PTY for the trust prompt and either handle it explicitly under the test's configuration-cleanup policy or fail with a clear prerequisite; do not interpret the absent record as an IPC/protocol failure.

## Safe harness pattern

- Require `--config-dir` or `--home`, pointing to a caller-created private temporary directory.
- Bind a Unix socket in a private short runtime directory.
- Use the harness process's real PID and `/proc` start value.
- Hash the resolved absolute socket path for the key filename.
- Buffer socket reads and process only complete newline-terminated frames.
- Recursively redact sensitive keys and token-like string values before printing.
- Exit successfully only after a matching user frame.
- Remove only the exact record, key, socket, and harness-owned runtime directory.

## Complementary Python-to-Claude sender pattern

A live validation also proved the reverse direction: an independently launched stdlib-only Python process can deliver to a real interactive Claude Code 2.1.232 inbox.

1. Launch Claude independently in tmux with a unique `--session-id` and a scratch `--messaging-socket-path`.
2. Locate exactly one registry record with that session ID. Validate that its PID is live and that `/proc/<pid>/stat` field 22 matches `procStart`.
3. Resolve `messagingSocketPath`, derive the sibling key filename with SHA-256 of the canonical path bytes, and validate the key's `procStart`.
4. Read `peerToken` only in the Python process. Never print it, include it in command arguments, or copy it to evidence.
5. Connect to the Unix socket and send an auth frame followed by the `type:"user"` frame above, each newline terminated.
6. Keep the Python process alive briefly so PID/PPID evidence can prove it and Claude were independently launched.
7. Use a fresh marker absent from the setup prompt. PASS requires the Claude transcript to contain the exact marker inside a completed Claude processing turn; `sendall()` success alone is insufficient.

In the validated run, Claude's safety layer sometimes declined the requested literal `RECEIVED:` acknowledgment while still quoting the exact fresh marker and producing a completed response about it. Therefore, use transcript receipt plus completed-turn evidence as the predicate, not one mandatory response phrase.

For sanitized process evidence, whitelist PID/PPID rather than dumping command lines. Tmux ancestry can include unrelated long-lived server command lines and arguments. Record a descriptive termination state when the TUI must be stopped by owned-process cleanup instead of inventing a numeric exit status.

## Evidence boundary

Both directions were validated live against Claude Code 2.1.232 with independently launched processes:

- **Claude-to-non-Claude:** a real interactive Claude session invoked its own `SendMessage` feature, the non-Claude receiver authenticated and decoded the `type:"user"` frame containing the exact fresh marker, PID ancestry showed neither process descended from the other, and all owned processes, tmux sessions, registry records, keys, and sockets were absent after cleanup.
- **Non-Claude-to-Claude:** the client sent authenticated frames successfully, Claude's transcript visibly contained and processed the exact fresh marker, PID ancestry showed process independence, the authentication value never entered evidence, and all test-owned processes and artifacts were absent after cleanup.

The validated harnesses were Python standard-library probes. A later Go port compiled and passed pure helper tests but did not complete its live run because the newly created workspace stopped at Claude's pre-registry trust dialog. Do not treat compilation or the Python result as proof that a different-language port works: every port still needs its own two-direction live markers and cleanup evidence.
