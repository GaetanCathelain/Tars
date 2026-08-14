---
name: local-ipc-probes
description: "Use when probing local process discovery or IPC protocols."
version: 1.0.0
platforms: [linux]
metadata:
  hermes:
    tags: [ipc, unix-domain-socket, protocol, reverse-engineering, proof-of-concept, security]
    related_skills: [spike, systematic-debugging]
---

# Local IPC Probes

Use this skill to build a disposable, non-production harness that impersonates or observes a local process peer: filesystem-based peer discovery, Unix-domain sockets, newline-delimited protocols, local auth handshakes, and cross-process message delivery.

The goal is a falsifiable answer backed by a runnable artifact, not a speculative protocol description.

## Core workflow

1. **State one observable PASS condition.** Example: a real sender discovers the fake peer and the harness receives a user-message frame.
2. **Separate discovered facts from guesses.** Record exact paths, field names, framing, timestamp units, hashing inputs, and auth order. Label anything unverified.
3. **Use an isolated caller-supplied root.** Never publish fake discovery records into the user's normal configuration unless explicitly requested. Accept a temporary config directory or HOME as an argument.
4. **Publish atomically.** Create private directories, bind the socket, write the key/credential material with mode `0600`, then atomically publish the discoverable record last. This prevents discovery of a half-created peer.
5. **Make the receiver tolerant.** Buffer stream reads, split complete newline-delimited frames, handle multiple connections, accept optional handshake frames, and reject or report malformed frames without crashing.
6. **Redact before printing.** Recursively redact fields whose names imply tokens, authorization, cookies, passwords, secrets, credentials, or API keys. Also redact recognizable bearer/token patterns inside strings. Never print generated key values, including in tests.
7. **Clean only owned artifacts.** Track exact record, key, socket, and runtime paths. Remove them on success, timeout, interrupt, and termination. Do not recursively delete the caller-supplied config root.
8. **Test in layers, but protect the decisive run.** First use a synthetic local sender to verify framing, redaction, PASS detection, and cleanup. Then run the real live sender as soon as the minimum compatible receiver exists—before optional source archaeology, beautification, or documentation. Reserve enough execution budget for the live run, evidence collection, and cleanup. Report synthetic and live evidence as separate levels.
9. **Treat compilation as a checkpoint, never the deliverable.** A receiver that parses, compiles, or publishes plausible records has not proved interoperability. Do not conclude until the real sender has discovered the peer, invoked its own messaging feature, and the exact marker has appeared in the receiver log.
10. **For client-to-real-app probes, require transcript-level proof.** A successful socket write proves transport only. Use a fresh marker that is absent from setup prompts, then require the real application's transcript to contain that exact marker in a completed processing turn. Do not require one fixed acknowledgment phrase: safety classifiers may refuse the requested wording while still visibly quoting and processing the marker.
11. **Whitelist process evidence fields.** Persist only the facts needed for independence and cleanup (normally PID, PPID, version, non-secret socket path, marker, exit/termination state, and owned-artifact state). Avoid full command lines and environment snapshots: ancestor processes can expose unrelated paths, endpoints, or arguments.

## Minimal artifact shape

Prefer standard-library-only probes with:

- one executable receiver;
- one short sender prompt or sender harness;
- one README containing exact two-terminal usage;
- a finite timeout and nonzero exit on failure;
- an explicit `READY`, `PASS`, and `FAIL` signal.

Keep generated runtime data outside the source directory. Use a private runtime directory and a short socket path because Unix socket path lengths are limited.

## Security invariants

- Require the supplied temporary root to be private (`0700`) or fail closed.
- Key files and sockets should be owner-only (`0600`).
- Generate tokens with a cryptographically secure standard-library primitive.
- Place secrets in files or frames only; never interpolate them into logs or command output.
- Write secret-bearing files before public discovery records, and remove discovery records before or alongside teardown.
- Treat peer names, paths, and incoming JSON as untrusted display data.

## Verification checklist

A complete result demonstrates:

- the discovery record has the expected schema and types;
- the key filename derives from the exact canonical socket path when the protocol requires it;
- a synthetic sender can connect and deliver multiple newline-delimited frames;
- sensitive fields appear as `[REDACTED]` in output;
- only the target message shape triggers PASS;
- record, key, socket, and private runtime directory disappear after exit;
- the real sender can list the fake peer and send to it, or the result is clearly marked synthetic-only.

## Pitfalls

- **Parent-session environment can invalidate process independence.** When launching a second instance of the probed app from inside that same app or an orchestrator, scrub only the parent's session-identity/IPC variables from the child. Never diagnose this by dumping the whole environment; it may contain live transport credentials. Version-specific Claude variable names and trust-dialog behavior are in `references/claude-code-cross-session-v2.1.232.md`.
- **Do not claim interoperability from a synthetic sender alone.** It validates the harness, not real peer discovery.
- **Timestamp representation matters.** Match the native producer's unit and JSON type; sorting or freshness checks may silently reject otherwise-correct records.
- **Hash the canonical path bytes.** Hashing a relative path, a symlink spelling, or a path with different encoding produces the wrong key filename.
- **Stream reads are not message reads.** One `recv()` may contain half a frame or many frames.
- **Cleanup tests need evidence.** Verify the files and socket are absent after both PASS and timeout/signal paths.
- **Do not copy live credentials into a PoC config.** Let the human launch the real sender with an already-supported authentication mechanism.

## References

- `references/claude-code-cross-session-v2.1.232.md` captures the version-specific Claude Code peer-record and NDJSON findings from a validated local harness build. Treat it as a version-pinned example, not a stable public API.
