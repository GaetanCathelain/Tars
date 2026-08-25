---
name: hermes-gateway-operations
description: "Use when safely configuring or verifying Hermes gateways."
version: 1.0.0
---

# Hermes Gateway Operations

Use this skill for Hermes messaging-gateway configuration, platform routing, authorization, display scoping, service lifecycle, and live-runtime verification.

The official documentation is authoritative and the installed source is the authority for the exact version currently running. Never infer a key from a similarly named setting on another platform.

## Operating principles

- Use `hermes config set ...`; never hand-edit `config.yaml`.
- Preserve credentials and existing authorization/routing unless the user explicitly asks to replace them.
- Inspect current resolved values before writing and query the exact values again afterward.
- Separate config-file resolution from live-daemon state. A fresh config-loader probe proves what the next process will load; it does not prove an already-running adapter has reloaded startup-time settings.
- Treat channels, group DMs, and 1:1 DMs as distinct surfaces. Confirm each platform adapter's exemptions and gates in docs and source.
- Prefer the narrowest composition of supported controls when one setting cannot express the requested scope.
- Never claim a restart happened or a live adapter changed unless process/service evidence confirms it.

## Procedure

1. **Resolve the active installation and profile**
   - Check `HERMES_HOME`, `hermes config path`, `hermes --version`, and the installed source path.
   - Confirm whether the gateway is foreground, a user service, or a system service.

2. **Establish the supported schema**
   - Read the relevant official messaging-platform documentation.
   - Search installed source for the exact config resolver, adapter gate, env bridge, and defaults.
   - Distinguish global display settings, per-platform overrides, and surface-specific overrides.

3. **Capture only non-secret baseline values**
   - Query relevant dotted keys individually with `hermes config get ... --json`.
   - For authorization variables, print only the allowlist/control values needed for verification; never dump `.env`.
   - Record service PID/state before any restart.

4. **Design the narrowest supported policy**
   - Write a behavior matrix before changing anything. Keep five independent axes explicit: **sender authorization**, **message/mention routing**, **tool capability/schema availability**, **external integration policy** (for example an MCP server's own outbound-write allowlist), and **display visibility** (tool progress/live status). A request about one axis does not authorize changes to another.
   - For external MCP write denials, inspect the MCP process command, its exact policy variable, and its live environment snapshot. Do not assume a Hermes inbound-channel allowlist controls an MCP server's outbound posting rights.
   - First look for a direct per-surface display override.
   - **Never use a channel/user allowlist or `platform_tool_conversations` to approximate a cosmetic visibility policy.** `slack.allowed_channels` changes whether messages are accepted; `platform_tool_conversations` changes whether tool schemas exist. Both are rights/capability boundaries, not presentation controls. If Hermes cannot express the requested display matrix, report the unsupported gap or delegate a display-layer change instead of denying valid requests or removing tools.
   - Verify whether 1:1 DMs are exempt and whether group DMs count as shared surfaces.
   - Before changing any actual authorization/routing allowlist, inventory every active shared conversation and explicit delivery destination the agent is expected to serve. A non-empty allowlist can silently discard valid mentions.
   - Ensure slash commands or per-session toggles cannot silently widen the policy when that matters.
   - Test the complete matrix: a valid mention in an arbitrary shared channel still reaches the agent **with normal tools**, while interim tool-progress/live-status output follows the requested visibility policy; home channel and 1:1 DM behavior remains correct; final substantive answers remain visible everywhere the request is accepted.
   - If an earlier mistaken capability gate caused a missed request, inspect the original platform thread, recover and route the missed work, and report in that original conversation after activation; removing the gate does not replay previously discarded events.

5. **Apply through the CLI or the owning config surface**
   - Run one explicit `hermes config set KEY VALUE` command per changed `config.yaml` key.
   - If the effective control belongs to an external MCP env file or to Hermes' credential/authorization env surface, use a targeted atomic update of only the named non-credential variable; preserve all unrelated lines, credentials, and file mode.
   - Do not invent a YAML key for an env-only control, and do not rewrite unrelated sections or credentials.

6. **Verify config resolution**
   - Re-query every changed key.
   - Run `hermes config check`.
   - When useful, run a fresh installed-runtime loader probe to prove YAML-to-env bridges and effective display resolution.

7. **Handle lifecycle safely**
   - Determine whether each changed setting is read per turn or only at adapter startup.
   - Restart only when required and when the command is outside the gateway's own process tree.
   - If Hermes blocks a self-restart because it would kill the active task, do not bypass the guard. Report the applied config as pending activation and give the exact external-shell restart and verification commands.

8. **Verify the live daemon**
   - Confirm service active state and a changed PID/start timestamp after restart.
   - Inspect logs/status for platform connection success without printing tokens.
   - Where feasible, test one allowed and one denied surface; label synthetic/config-loader checks separately from live message tests.

## Desktop as a remote-only client

When Desktop runs on one computer and Hermes Agent runs on another, distinguish the local **Desktop shell** from an optional local **Agent runtime**. Verify remote-only behavior against the user's exact released tag, not only current documentation or `main`.

Before recommending or provisioning a connection path:

1. Confirm the Desktop and remote Agent versions separately.
2. Verify the tagged first-run flow offers **Connect to existing Hermes** before local bootstrap and that its tests prove the remote choice bypasses local installation.
3. Identify the transport the first-run UI actually accepts. A generic claim that Desktop “supports SSH” is insufficient when first-run onboarding accepts only an authenticated HTTP(S) Gateway URL and exposes SSH later under Settings → Connections.
4. Inspect the remote host for a real Desktop backend endpoint. A running Slack/Discord gateway does not prove that `hermes serve` is listening or reachable.
5. **Preflight the exact SSH direction before creating either lifecycle artifact.** For a conventional local forward, prove Desktop→Agent noninteractive SSH using the exact hostname, remote username, host key, and identity that the tunnel will use. Verify a newly discovered host key against the server's local public-host-key fingerprint before trusting it. A DNS/TCP connection followed by `Permission denied (publickey)` is not a usable route. If the task says to stop on SSH authentication failure, leave both lifecycle artifacts absent rather than creating a half-deployment. When Desktop→Agent auth is unavailable but an already-authorized Agent→Desktop route exists, use the equivalent reverse-tunnel design: run `hermes serve` on Agent loopback and request `-R 127.0.0.1:9119:127.0.0.1:9119` from the Agent host. Do not create a second identity or widen the listener merely to preserve the conventional direction.
6. For a persistent loopback endpoint, treat liveness auth and API/WebSocket auth as separate layers. `/api/health` and `/api/status` may be public while Desktop still classifies `auth_required: false` as token mode and requires the dashboard session token for protected REST and `/api/ws`. If a stable token is needed across service restarts, use the installed release's supported `HERMES_DASHBOARD_SESSION_TOKEN` input, keep it in a mode-0600 secret file or equivalent credential store, and never place or print the value in the service command line, plist, logs, or report.
7. For Bot Mode, verify remote connectivity and Bot Mode version support as independent gates, then ensure creation targets the intended remote connection.

Keep loopback-bound services private and prefer a loopback SSH tunnel or trusted private network over a public bind. Snapshot unrelated listeners and daemons first, and re-check them after deployment. See `references/desktop-remote-only.md` for the evidence hierarchy, token contract, and onboarding distinction.

## Outbound delivery from an active agent

Use this when the user asks the bot to post into a different configured conversation or thread.

1. Check the installed CLI first: `hermes send --help`. Do not conclude that outbound posting is unavailable merely because the current tool schema has no `chat.postMessage`-style tool.
2. Use `hermes send --to 'platform:chat_id:thread_id' 'message'`. The command reuses the gateway-owned credentials; never read, print, interpolate, or pass bot tokens yourself.
3. Confirm that the bot belongs to the destination conversation and that the user authorized this specific message. Preserve the exact requested payload—especially when they say “only this GIF.”
4. When the user asks for a test before the real send, treat the order as an acceptance criterion: post through the same transport and payload form to the named test conversation, verify the result, and only then post to the destination. Do not parallelize the test and real delivery, and do not substitute a local file inspection for the requested Slack-rendering test.
5. For animated GIFs, verify the source is genuinely multi-frame before posting, then test the same representation that will be used in production (direct URL versus uploaded `MEDIA:<path>`). A successful API response proves delivery, not animation or unfurl rendering.
6. Capture the returned message ID and read the destination thread back through the platform integration. Report plainly if a correction arrived after an irreversible send; never imply the requested sequence was followed when it was not.

## Evidence standard

A completion report should include:

- exact credential-free commands executed;
- effective values;
- the docs/source locations establishing semantics;
- service/PID evidence for any restart;
- whether verification was config-level, fresh-runtime, or live-message level;
- files modified and explicit confirmation that credential/repository files were untouched.

## Pitfalls

- A busy acknowledgment is not proof of conversation scope or session-key collision. When observed Slack behavior contradicts the expected binding model, verify the concrete message's routing key, session freshness, busy-path precondition, event multiplicity, and chronology before explaining it. Separate the verified binding behavior from the observed anomaly and label duplicate delivery or deduplication races as hypotheses until raw-event evidence confirms them.
- `config get` may expose different roots for gateway enablement, adapter behavior, and display behavior. Query rather than assuming one unified Slack/Discord subtree.
- Per-platform tool progress and ephemeral live-status text can be independent features; disabling one may not disable the other.
- Adapter YAML-to-env bridges commonly run at startup, so a correct file value may remain inactive in the existing daemon.
- Reading `/proc/<pid>/environ` is not sufficient proof of dotenv-backed authorization state; loaders may maintain values outside the original process environment snapshot.
- A restart command issued from a gateway-owned tool process can terminate its own agent turn. Respect lifecycle guards and restart externally.

## References

- `references/private-desktop-reverse-tunnel.md` — Verified reverse-tunnel architecture, credential-safe persistent token pattern, user-systemd lifecycle, and end-to-end HTTP/WebSocket/privacy checks when only Agent→Desktop SSH works.
- `references/slack-progress-scoping.md` — Slack tool-progress scoping, DM/channel semantics, and verification pattern discovered on Hermes v0.20.0.
- `references/slack-busy-ack-diagnostics.md` — Evidence hierarchy for reconciling erroneous busy acknowledgments with thread-scoped Slack sessions.
- `references/slack-mcp-write-policy.md` — Diagnose and safely change external Slack MCP outbound write allowlists without confusing them with Hermes sender or inbound-channel gates.
