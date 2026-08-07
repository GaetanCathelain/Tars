# Orca Web Surface (cooper:6768) — Security Posture Probe

Scope: local review of Gaetan's own dev box `cooper`. Non-mutating, localhost-only
testing per mandate. No other host touched, no port scan, no mutating RPC call made.

## Bind & reachability (facts)

- `ss -tlnp`: `orca-ide` (pid 218650) LISTEN on `0.0.0.0:6768` — all interfaces.
- Confirmed independently by the app's own state file, `~/.config/orca/orca-runtime.json`:
  `"transports": [{"kind":"unix","endpoint":".../o-218650-e55e.sock"}, {"kind":"websocket","endpoint":"ws://0.0.0.0:6768"}]`.
- Process tree (`ps aux`) shows this instance was launched as:
  `/usr/local/bin/orca serve --port 6768 --pairing-address 100.115.232.100`
  → spawns `orca-ide --serve --serve-port --serve-pairing-address 6768 100.115.232.100`.
- `ufw status` → **inactive** (no host firewall on cooper at all — not Orca-specific).
- cooper's tailnet IP `100.115.232.100` and LAN `192.168.0.0/24` can both reach the
  socket in principle (bind + no firewall); **not verified by an actual off-box
  connection**, per mandate (reasoned from bind address + ufw state only).
- Root cause of the `0.0.0.0` bind, found in source (see below): running via the
  `orca serve` CLI path unconditionally sets `exposeNetworkByDefault = true`, and
  `resolveInitialWebSocketBindHost()` returns `WS_BIND_HOST_ALL_INTERFACES` whenever
  that flag is true — independent of whether any device has actually paired
  remotely. `--pairing-address` only changes the address embedded in the pairing
  QR/link, not the bind decision (confirmed by the CLI's own help text: *"changes
  only the client-advertised address"*).

## API surface (from source)

Single combo HTTP+WebSocket server (`WebSocketTransport`, `out/main` bundle inside
`/opt/Orca/resources/app.asar`, class `WebSocketTransport`/`OrcaRuntimeRpcServer`):

- `node:http` (or `https` if TLS configured) server with a static-file request
  handler (`createStaticWebClientHandler(webClientRoot)`) serving the SPA shell
  (`index.html` + `/assets/*.js`) — this explains why `/api`, `/api/health`,
  `/health`, `/status` all 404: there is no REST tree, only static assets plus one
  attached `ws.WebSocketServer` with no path restriction (upgrades on any path).
- All real functionality is one JSON-RPC-style protocol tunneled over that single
  WebSocket. The method registry (`ALL_RPC_METHODS`, asar `out/main` bundle) is
  built from ~35 category arrays including `REPO_METHODS`, `WORKTREE_METHODS`,
  `AGENT_SESSION_METHODS`, `TERMINAL_METHODS`, `ORCHESTRATION_METHODS`,
  `COMPUTER_METHODS`, `BROWSER_*_METHODS`, `GIT_METHODS`, `GITHUB_METHODS`,
  `SSH_METHODS`, `PLUGIN_METHODS`, `SKILL_METHODS`, `ACCOUNT_METHODS`, etc. Sample
  method names seen directly: `worktree.create`, `worktree.rm`,
  `worktree.forceDeleteBranch`, `worktree.ps`, `terminal.wait`,
  `terminal.updateViewport`, `orchestration.ask`, `orchestration.check`,
  `ui.get`/`ui.set`, `status.get`. This is exactly the surface the mandate worried
  about (repos, worktrees, terminals, orchestration, agent spawning) — all reached
  only through this one WS RPC channel.
- A second, separate transport exists: a Unix domain socket
  (`~/.config/orca/o-<pid>-<hash>.sock`, mode implied 600 by directory perms) used
  for local CLI / desktop-IPC calls. That channel is filesystem-permission
  protected and not reachable over TCP 6768 at all.

## Auth mechanism

Two independent, non-overlapping auth tiers exist in source
(`OrcaRuntimeRpcServer` class):

**Tier A — local-only channel (Unix socket / desktop IPC), not exposed on :6768:**
`this.authToken = randomBytes(24).toString('hex')`, checked in `parseAndAuth()`:
missing or mismatching token → `{error:"unauthorized"}`. Persisted to
`~/.config/orca/orca-runtime.json` (mode `600`, owner `gaetan` only — verified).
Electron main↔renderer IPC and local CLI use hardcoded sentinel values
(`"desktop-ipc"`, `"remote-cli"`) but those only travel over the local socket/IPC,
never over the network listener.

**Tier B — the network WebSocket on :6768 (the one reachable from LAN/tailnet):**
Every inbound WS connection is wrapped by `MobileSocketWiring`, which requires a
successful **E2EE handshake** (`e2ee_hello`) against the server's per-runtime
`e2eeKeypair` *before* any plaintext RPC frame is even parsed. Only after that
handshake does `handleWebSocketMessage()` run, which then requires a
**`deviceToken`** validated against `this.deviceRegistry.validateToken(token)`
(persisted in `~/.config/orca/orca-devices.json`, mode `600`, owner-only).
Unknown/absent tokens → `{unauthorized, "Missing device token"}` /
`{unauthorized, "Invalid device token"}`. Mobile-scoped devices are further capped
to a `MOBILE_RPC_METHOD_ALLOWLIST` subset.

Device tokens are minted only via the pairing flow (`createPairingOffer` /
`createMobilePairingOffer`): a pairing URL/QR code is generated and shown locally
(terminal QR via `renderTerminalPairingQr`, or the desktop pairing UI) — the token
is never broadcast; it only leaves the machine if a human deliberately shares or
scans it. I did **not** find any code path on the TCP :6768 listener that bypasses
either the E2EE handshake or the device-token check (no "trust 127.0.0.1" / "trust
same-origin" shortcut located in the reviewed `OrcaRuntimeRpcServer`,
`WebSocketTransport`, or `MobileSocketWiring` code).

## Read-only probe results (verbatim)

All from `127.0.0.1`, no credentials presented, nothing created/modified/deleted.

1. `curl http://127.0.0.1:6768/`
   → `HTTP/1.1 200 OK`, `Content-Type: text/html; charset=utf-8`,
   `Content-Length: 3566`, body contains `<title>Orca Web</title>` (SPA shell only,
   no application data).
2. `curl http://127.0.0.1:6768/assets/web-D2f7P9Ss.js`
   → `HTTP 200`, 146,582 bytes (client JS bundle; no application data, but does leak
   the RPC method surface — recon value, not access).
3. `/api`, `/api/health`, `/health`, `/status` → all `404` (pre-established, re-
   confirmed no REST tree exists — everything routes through the WS RPC channel).
4. Raw WebSocket handshake to `ws://127.0.0.1:6768/` (custom Python socket script,
   no library): `GET / HTTP/1.1` + `Upgrade: websocket` headers →
   `HTTP/1.1 101 Switching Protocols` (the upgrade itself is unauthenticated —
   anyone who can reach the port can open a raw WS connection).
5. Immediately after the upgrade, sent **one** masked text frame containing a
   read-only RPC request with **no** `authToken`/`deviceToken`/E2EE envelope:
   `{"id":"probe-1","method":"status.get"}`. Server's entire response was a close
   frame (opcode `0x88`), payload `Invalid e2ee_hello` (close code bytes `0x0fa1`
   = 4001, app-defined policy-violation range). The connection was then closed by
   the server. **No RPC data of any kind was returned** — the request was rejected
   before it ever reached `handleWebSocketMessage`/`deviceRegistry.validateToken`,
   i.e. before the device-token check even runs. This empirically confirms the
   source-level finding: the outer E2EE envelope gate is real and fires on the
   first unauthenticated message.

## What I did NOT test, and why

- No call to any state-changing method (`worktree.create/rm`, `terminal.*`,
  `orchestration.*`, etc.) — forbidden by mandate; also unreachable anyway since
  `status.get` (read-only) was already rejected pre-dispatch, so no read method
  would have gotten further either.
- No access attempt from another host (LAN peer, phone, etc.) — forbidden by
  mandate; reachability is reasoned from bind address (`0.0.0.0`) + `ufw inactive`
  only.
- No port scan of LAN/tailnet.
- Did not attempt to complete a real pairing (would require the QR/link + E2EE key
  exchange with material only the legitimate desktop app holds) — out of scope for
  a "no credentials" test and not needed to answer the question.
- Did not verify whether a `deviceToken`, once paired, is bound to a specific
  source IP or is a bearer-style credential usable from any reachable address.
  Source review didn't show explicit IP binding in the paths read; this is a real
  residual unknown worth a follow-up read of `MobileSocketWiring`/`deviceRegistry`
  if precise blast-radius of a leaked pairing matters.
- Did not enumerate every one of the ~35 RPC method categories exhaustively (listed
  the categories and representative method names found; full names weren't needed
  to answer the auth question).

## Severity assessment

**Not** the "anyone on LAN/tailnet can spawn agents and read source" scenario the
mandate flagged as the worst case. The state-reading/state-changing RPC surface is
gated by two stacked, unbypassed checks (E2EE handshake + per-message device
token against a locally-persisted, owner-only-readable registry), and a live probe
against a real read-only method with zero credentials was rejected before
dispatch.

Residual risk, in descending order:
1. **Open static asset server** on `0.0.0.0:6768` with no auth: low-severity
   information disclosure (confirms Orca is running, discloses the RPC method
   name surface / bundle hashes). Useful recon for an attacker, not access.
2. **Single point of failure**: all real protection rests on one code path
   (E2EE + device-token check inside one server class). No defense-in-depth —
   `ufw` is fully inactive on cooper, so if that one check ever regresses (a
   future Orca release, a config change, a bug), there is nothing else in front of
   it. This is exactly the kind of "existing control that would lapse" worth
   flagging even though nothing is broken today.
3. **Pairing-token capture**: if an attacker obtains a valid `deviceToken` (e.g.
   shoulder-surfing/screenshotting the pairing QR, or reading
   `orca-devices.json` if they ever get any code-execution/read access on cooper),
   they could plausibly impersonate that device from anywhere the network reaches
   `0.0.0.0:6768` — IP-binding of tokens was not confirmed either way (see "not
   tested" above).

## Remediation options, ranked by effort

1. **Cheapest — stop forcing `--pairing-address <tailnet-ip>` / the always-network
   `orca serve` invocation for an always-on process.** The `0.0.0.0` bind is a
   direct, unconditional consequence of `orca serve`'s `exposeNetworkByDefault`
   flag (source-confirmed) — it is not a settings toggle, so the fix is at the
   launch-command layer (whatever supervises `orca-ide --serve …`, e.g. a systemd
   unit or shell wrapper). Note `--no-pairing` does **not** appear to change
   `exposeNetworkByDefault` in the reviewed code, so it alone won't narrow the
   bind — don't rely on it for this.
2. **Reasonable — scoped firewall rule instead of touching the launch mode**, since
   remote/mobile pairing over the tailnet is presumably wanted: e.g.
   `ufw allow in on tailscale0 to any port 6768` plus a default-deny for
   `192.168.0.0/24` → `6768/tcp`, rather than enabling `ufw` wholesale (bigger
   blast radius, out of scope here — but flagged: today NOTHING is firewalled on
   cooper, Orca or otherwise).
3. **Most correct if remote pairing is rarely needed**: don't expose the port
   network-wide at all; have remote devices reach it via `ssh -L 6768:localhost:6768
   gaetan@cooper` or a Tailscale Serve ACL scoped to specific tailnet identities,
   and drop `--pairing-address`/serve's forced exposure entirely.
4. **No further app-level change strictly required** given the code-level auth
   gate held under test — but consider periodically auditing
   `~/.config/orca/orca-devices.json` for unexpected paired devices, since that
   file is the trust root for tier B once a token exists.

## Sources

- `ss -tlnp`, `ps aux`, `ufw status` (live cooper state).
- `~/.config/orca/orca-runtime.json` (bind confirmation; secret values redacted in
  this report).
- `/opt/Orca/resources/app.asar.unpacked/out/cli/specs/serve.js`,
  `.../out/cli/handlers/core.js`, `.../out/cli/runtime/launch.js` (CLI → child
  process argv wiring for `orca serve`).
- `/opt/Orca/resources/app.asar` (packed Electron main bundle, read via
  `strings -a -n 6`): `OrcaRuntimeRpcServer` class (`parseAndAuth`,
  `handleWebSocketMessage`, `createPairingOffer`, `resolveInitialWebSocketBindHost`,
  `ensureNetworkExposure`), `WebSocketTransport` class, `ALL_RPC_METHODS` /
  `MOBILE_RPC_METHOD_ALLOWLIST` registries.
- `http://127.0.0.1:6768/` and `/assets/web-D2f7P9Ss.js` (curl, unauthenticated).
- Custom raw-socket Python script (`/tmp/ws_probe.py`, not part of this repo) doing
  a manual WS upgrade + one masked text frame with a read-only, credential-free
  RPC request, to observe the server's actual rejection behavior.
