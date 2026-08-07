# Probe: Claude Code / Claude Agent SDK vs the A2A protocol

Date: 2026-08-07. Research task, read-only — no code changed, no commits.

## One-line answer

**Claude Code does not speak A2A today.** No official Anthropic server or
client support exists. Everything that connects the two is a small,
early-stage, third-party wrapper (single-digit-to-low-hundreds GitHub stars,
explicitly "not production ready"). Anthropic's own mesh for agent-to-agent
messaging (`SendMessage`/`ListAgents`) is a separate, undocumented,
local-machine-only proprietary bus — not A2A, not network-reachable, no public
protocol surface.

---

## A. A2A protocol — current state (web research, 2026-08-07)

### Discovery path — THIS CHANGED, use the new one

- **Current (v1.0, live today): `/.well-known/agent-card.json`** — per RFC 8615
  well-known-URI convention.
- **Old/deprecated path: `/.well-known/agent.json`** (pre-1.0, ≤0.2.x). The
  spec source itself flags this as a breaking rename; several
  still-in-the-wild guides and tutorials reference the old path and will
  silently fail discovery against a v1.0 agent. If Hermes ever needs to guess
  a path against an unknown A2A server, try `agent-card.json` first,
  `agent.json` as a fallback for pre-1.0 servers.
- Two other discovery mechanisms exist besides well-known: **curated
  registries** (a directory service agents query by skill/tag/provider) and
  **direct configuration** (hardcoded URL, env var, config file) — no
  registry is standard/mandatory.

### Core RPC methods (transport-independent; A2A v1.0.0)

- `message/send` — send a message, get back a `Task` or a `Message`
- `message/stream` — same, with real-time streaming updates
- `tasks/get` — fetch current task state + artifacts
- `tasks/list` — filtered, cursor-paginated task listing
- `tasks/cancel` — request cancellation
- `tasks/subscribe` — open a streaming connection to an existing task
- `tasks/pushNotificationConfig/set` (create), `/get`, `/list`, `/delete` —
  manage webhook registrations for async task updates
- `agent/getAuthenticatedExtendedCard` — fetch the authenticated (fuller)
  agent card

### Transport bindings

Spec mandates all three be exposable with identical semantics:
JSON-RPC 2.0 (spec §9), gRPC (§10), HTTP+JSON/REST (§11), plus a
"custom binding" escape hatch (§12) with interoperability requirements.

### Auth model

Declared per-agent in the Agent Card's `securitySchemes`: API key, HTTP
Basic/Bearer, OAuth 2.0 (multiple flows), OIDC, mutual TLS. Spec text: "Agents
MUST reject requests with invalid or missing authentication credentials."
Fetching the *extended* agent card additionally requires the caller to
authenticate.

### Streaming

SSE on HTTP bindings. Contract: "The stream MUST begin with the Task object,
followed by zero or more `TaskStatusUpdateEvent` or `TaskArtifactUpdateEvent`
objects. The stream MUST close when the task reaches a terminal state."
Multiple concurrent subscriber streams per task are allowed independently.

### Push notifications / webhooks

Client registers a webhook (`tasks/pushNotificationConfig/set`) with a target
URL + optional auth token/`AuthenticationInfo`. On task state change, the
agent does a plain HTTP POST of a `StreamResponse` payload to that URL —
this happens over plain HTTP *regardless* of the primary transport binding
the task used. This is the mechanism a remote, non-polling Hermes would use
to get results back without holding a connection open.

### Protocol maturity/version note (staleness flag)

v1.0.0 is current; prior versions 0.3.0, 0.2.6, 0.1.0 existed with material
breaking changes (the well-known path rename above is one). Version is
negotiated per-request via an `A2A-Version` parameter. Given the churn from
0.x → 1.0, **treat anything written before ~mid-2026 about A2A specifics as
suspect** — several blog posts found in search results still describe the old
path or pre-1.0 method names.

---

## B. Does Anthropic/Claude Code ship A2A support?

**No official support found.** Searched:
- `docs.anthropic.com` / `code.claude.com` / `platform.claude.com` site search
  for "A2A" / "Agent2Agent" — no hits.
- Agent SDK overview page (`code.claude.com/docs/en/agent-sdk/overview`,
  fetched 2026-08-07) — full capability table (tools, hooks, subagents, MCP,
  permissions, sessions, skills, plugins) with zero mention of A2A anywhere.
- Headless-mode doc (`code.claude.com/docs/en/headless`) — same, no mention.
- `claude --help`, `claude mcp --help`, `claude agents --help` (local, this
  machine, v2.1.223) — no `a2a` subcommand, flag, or reference anywhere in
  CLI surface.
- Anthropic's one webinar that pairs "Claude" with "A2A"
  (`anthropic.com/webinars/deploying-multi-agent-systems-using-mcp-and-a2a-with-claude-on-vertex-ai`)
  is about using Claude models *as an agent implementation* inside **Google's**
  Vertex AI Agent Builder stack, which supplies the A2A layer itself — it is
  not Anthropic shipping A2A support in Claude Code or the Agent SDK.

Absence of evidence, stated precisely: no `a2a` keyword anywhere in Claude
Code's CLI help, no A2A section in the Agent SDK or headless-mode docs, no
A2A entry in either SDK's public GitHub changelog links surfaced, no MCP
server-mode doc mentioning A2A.

## Existing bridges (community, not Anthropic)

| Repo/Project | Direction | What it wraps | Maturity |
|---|---|---|---|
| [ericabouaf/claude-a2a](https://github.com/ericabouaf/claude-a2a) | Exposes Claude Code **as** an A2A server (agent card at `http://localhost:3008/.well-known/agent-card`, note: no `.json` suffix shown in their README — verify against current spec before relying on it) | Claude Code SDK (not raw `claude -p`) | 10 stars, 2 forks, 2 open issues. README: **"WARNING: This project is not production ready. Use it at your own risk."** No auth, no config file, no persistent storage, no Docker, no WebSocket — all listed as TODO. |
| a2claude ([dev.to writeup](https://dev.to/kanywst/a2claude-turn-claude-code-into-a-server-other-ai-agents-can-call-1mf6)) | Exposes Claude Code **as** an A2A server, over JSON-RPC/gRPC/REST | Claude Agent SDK via a normalized event abstraction (translation layer doesn't import the SDK directly); has an "echo" backend for offline testing | Early-stage, single-author project; maps A2A's `input-required` task state to Claude Code's permission-approval pauses — a genuinely useful design (mid-task tool-permission prompts become an A2A state transition) but unverified in production. |
| [GongRzhe/A2A-MCP-Server](https://mcpservers.org/servers/GongRzhe/A2A-MCP-Server) | **Opposite direction**: lets an MCP client (e.g. Claude Desktop, Claude Code as MCP *client*) call *out* to remote A2A agents. Does not expose Claude Code as an A2A server. | MCP↔A2A bridge, flow is `MCP Client → FastMCP Server → A2A Client → A2A Agent` | Early-stage, docs reference Claude Desktop config, no mention of Claude Code or the Agent SDK specifically. |
| [steipete/claude-code-mcp](https://github.com/steipete/claude-code-mcp) | Exposes Claude Code as a one-shot **MCP** server (not A2A) so another MCP client can invoke it | Wraps the CLI | Community, "one-shot" (task-and-exit) design, useful precedent for the "MCP server exposing Claude Code" option below but is MCP not A2A. |

None of these are Anthropic-maintained, none have meaningful adoption
signals (stars/forks/issue activity), and the two genuine "Claude Code as A2A
server" projects both explicitly self-describe as not production-ready.

### A2A SDK maturity (for anyone building a bridge from scratch)

- [a2aproject/a2a-python](https://github.com/a2aproject/a2a-python) — 1.4k
  stars, official Linux Foundation project (A2A was Google-originated, donated
  to Linux Foundation), `pip install a2a-sdk`, now at full v1.0 protocol
  compatibility with a 0.3 compat mode. Actively maintained.
- Official SDKs claimed for 6 languages: Python, JS, Java, Go, .NET, Rust.
  Python is the most mature; did not independently verify JS/other-language
  parity depth in this pass.

### A2A vs MCP — how they relate (a2a-protocol.org's own framing)

MCP = **vertical** (agent → its own tools/data, single agent's toolbox). A2A
= **horizontal** (peer agent → peer agent, delegation and collaboration
between autonomous systems with their own internal state/reasoning/tools).
Official recommended pattern: use both, layered — "An agentic application
might primarily use A2A to communicate with other agents. Each individual
agent internally uses MCP to interact with its specific tools and resources."
An A2A server *can* expose a "skill" that is really just a stateless
tool-shaped operation via A2A too, but that's discouraged when MCP already
fits (A2A's value is in stateful, multi-turn, opaque-to-caller
collaboration — MCP is not designed for that).

For the Hermes↔Claude-Code case specifically: MCP would be the wrong layer
if Hermes wants to hand Claude Code an open-ended task and get structured
task/artifact/status back over time (that's exactly A2A's shape); MCP would
be right if Hermes just wants to call one well-defined tool-like operation
Claude Code exposes.

---

## C. Claude Code's OWN inter-agent mesh (local inspection, this machine)

This machine: `cooper`, Claude Code v2.1.223 (per `~/.claude/sessions/*.json`
and `daemon.log`), inspected 2026-08-07.

### What `SendMessage`/`ListAgents` actually is

Fetched the live tool schema for `SendMessage` (via `ToolSearch
select:SendMessage,ListAgents`). Key facts straight from the schema:

- Addressing is **by name**, not by network address: `"to": "researcher"`,
  `"to": "main"`, or `"to": "worker [3fa9c1]"` (name + disambiguating ref
  shown by `ListAgents`). There is no host:port, no URL, no session token in
  the addressing scheme.
- "Your plain text output is NOT visible to other agents — to communicate,
  you MUST call this tool." — confirms it's a dedicated message-passing
  channel, not shared memory/transcript.
- Cross-session delivery wraps the payload as
  `<cross-session-message from="...">` — i.e. there's a sender identity tag,
  but it's assigned by the local harness (the `from` is a locally-known
  session/agent name), not any external/verifiable credential.
- Explicit permission-boundary rule baked into the tool description: a
  session must never ask a peer to do something its own permission settings
  would block — "cross-session permission laundering" is a named, guarded-
  against failure mode. This only makes sense as a same-machine, same-trust-
  domain concept; it is not designed with an untrusted remote peer in mind.

### What's under the hood (local files inspected)

- `~/.claude/daemon/roster.json` — tracks live workers under a supervisor:
  ```json
  {"proto": 1, "supervisorPid": 3667589, "updatedAt": 1785765169866, "workers": {}}
  ```
  `"proto": 1` is the daemon's own internal roster-protocol version — separate
  from the `peerProtocol` field seen per-session (below). Both are private,
  undocumented, and unrelated to A2A's `A2A-Version`.
- `~/.claude/daemon/control.key` — 32-byte ASCII secret (not printed, per this
  repo's and the global security rules); this is almost certainly an
  authentication token gating the daemon's local control channel.
- `~/.claude/daemon.log` — confirms the daemon:
  - binds a **Unix domain control socket**: `/tmp/cc-daemon-<uid>/<hash>/control.sock`
  - manages worker processes ("bg spawned", "bg spare spawned host pid=…")
  - self-restarts on binary upgrade, idle-exits after no clients
  - is versioned in lockstep with the CLI (`version=2.1.217` → `2.1.220` in
    the log, matching `claude --version` upgrades)
- `~/.claude/sessions/<pid>.json` — one file per live/recent session:
  ```json
  {"pid":2546713,"sessionId":"38714606-...","cwd":"/home/gaetan/dev/...",
   "startedAt":..., "procStart":"9592646", "version":"2.1.223",
   "peerProtocol":1, "kind":"interactive", "entrypoint":"cli",
   "name":"mc-metarepo-ht-console-gang-b7", "nameSource":"derived",
   "status":"idle", ...}
  ```
  `peerProtocol: 1` — an internal wire-protocol version for the mesh itself.
- `ss -xlp` on this machine shows **per-session Unix-domain listening
  sockets**: `/run/user/1000/cc-socks/<pid>.sock`, one per live `claude`
  process (confirmed 11 live sockets, one per interactive/background
  session PID currently running).

### Verdict on the mesh

It is a **local-only, proprietary, undocumented IPC bus**: a supervisor
daemon (keyed by a local secret, `control.key`) brokers messages between
sibling `claude` processes over Unix domain sockets in `/run/user/<uid>/` and
`/tmp/cc-daemon-<uid>/`, addressed by human-assigned session names, versioned
by two separate private integers (`proto`, `peerProtocol`) that have nothing
to do with A2A. No TCP/HTTP surface, no listening socket bound to a routable
interface was found for this mechanism specifically (`ss -tlnp` shows other
unrelated services on this box — `orca-ide`, `agent-browser-l`, various
`127.0.0.1:81xx` — but nothing under `claude`/`cc-daemon` bound to anything
but a Unix socket or `127.0.0.1`-only loopback-adjacent paths).

**Could a non-Claude-Code agent inject a message into this mesh?** Only by
reverse-engineering an undocumented Unix-socket wire protocol, obtaining
`control.key` (a local file, mode-restricted to the owning user), and running
*on the same machine* — since the socket paths are under `/run/user/<uid>`
and `/tmp`, not network-bound. A remote host (Hermes, on another machine) has
**no path in** at all without an SSH tunnel to a Unix socket plus reverse-
engineered framing — not a realistic integration point. This is the honest
"no" the task asked me to be precise about: this mesh is not, and does not
pretend to be, a network protocol.

### Claude Code's actual server-shaped surfaces

Three real ones, none of them A2A:

1. **`claude -p` (headless/print mode)** — one-shot or session-resumable
   subprocess. Confirmed capabilities from `claude --help` + official docs
   (`code.claude.com/docs/en/headless`, fetched 2026-08-07):
   - `--output-format json|stream-json`, `--input-format stream-json` for
     structured, streamable I/O
   - `--resume <session-id>` / `--continue` / `--fork-session` for
     multi-turn continuation — **works cross-directory as of v2.1.223**, so a
     bridge process can hold a session ID and drive it turn-by-turn
   - `--include-partial-messages` for token-level streaming events
   - a `system/init` event reports loaded tools/MCP servers/plugins and a
     `capabilities` array for feature detection
   - exit code 0/non-zero for scripted branching; SIGTERM triggers a clean
     `SessionEnd` hook + exit 143
   - This is exactly the substrate `claude-a2a` and `a2claude` both build on
     (or the SDK equivalent of it) — spawn/pipe/parse-stream-json is the
     entire integration surface.
2. **`claude mcp serve`** — starts Claude Code's own tools (Read/Edit/Bash/
   etc.) as an MCP server so *other* MCP clients (Claude Desktop, Cursor,
   etc.) can call them. This exposes Claude Code's **tools**, not a "run an
   agentic task and get a result back" surface — wrong shape for what Hermes
   wants (task delegation), right shape only if Hermes wanted individual
   file/bash primitives.
3. **Claude Agent SDK (Python/TypeScript only)** — a library, not itself a
   server: "same tools, agent loop, and context management that power Claude
   Code, programmable in Python and TypeScript." Per the official overview
   (`code.claude.com/docs/en/agent-sdk/overview`, fetched 2026-08-07):
   - No A2A/Agent2Agent mention anywhere on the page.
   - Sessions "maintain context across exchanges, resume or fork later" —
     same session model as headless mode, just in-process instead of
     subprocess.
   - Explicitly the recommended layer for "building production AI agents
     with Claude Code as a library" — i.e. the right foundation for anyone
     writing an A2A wrapper today, since `claude-a2a`/`a2claude` both already
     chose it over raw subprocess piping.
   - Adjacent Anthropic product worth flagging: **Managed Agents** (hosted
     REST API, "Anthropic runs the agent and the sandbox") — a different
     product from the Agent SDK, for teams that don't want to run their own
     session infra. Not A2A either, but closer in spirit to "a server Hermes
     could just call" if self-hosting the bridge is unwanted. Not
     investigated further — out of scope for this probe (would need its own
     recon pass on auth model and network reachability from the Tars VM).

---

## D. Ranked options for Hermes → Claude Code task delegation

Ranked by realistic effort-to-working-today, most to least attractive given
this repo's existing constraints (Tars/Hermes already talks to `cooper`/Orca
over SSH per `CLAUDE.md`):

1. **SSH + Orca/Claude Code headless (`claude -p`) — lowest effort, do this
   first.** Hermes SSHes into `cooper` (or wherever a Claude Code checkout
   lives), runs `claude -p "<task>" --output-format stream-json --resume
   <id>`, parses newline-delimited JSON off stdout. This is *exactly* the
   pattern this repo's own `CLAUDE.md` already uses for delegating to
   Orca/cooper, and it's the same substrate the community A2A bridges wrap —
   you'd just be skipping the A2A envelope. Failure modes: no auth beyond
   SSH-key trust (fine here, same trust domain as today), no standard task/
   artifact schema (Hermes has to define its own), no push-notification
   equivalent (must poll stream output or long-hold the SSH session), no
   protocol-level cross-vendor interop if Hermes ever needs to talk to a
   *third* agent vendor's stack. Effort: near-zero if reusing existing SSH
   wiring; a thin JSON-line parser on the Hermes side.

2. **Thin custom HTTP wrapper around `claude -p`/Agent SDK, NOT A2A-shaped**
   — a small persistent process (Python/Node, using the Agent SDK for
   session lifecycle) that exposes a couple of REST endpoints
   (`POST /task`, `GET /task/:id`) backed by `--resume`/session IDs, with a
   webhook callback to Hermes on completion (hand-rolled, not A2A's
   `pushNotificationConfig` schema). Effort: low-to-medium — a day or two;
   reuses the same session/stream-json substrate as option 1 but gives
   Hermes a stable network address instead of an SSH exec each time, and a
   real completion callback instead of holding a stream open. Failure mode:
   another bespoke protocol only Hermes and this wrapper understand — no
   interop payoff, but also no A2A spec compliance burden.

3. **Adopt/harden an existing community A2A wrapper (`claude-a2a` or
   `a2claude`) — gets you the real A2A envelope, but you inherit
   unmaintained code.** Effort: medium — both are small enough to read
   end-to-end in an afternoon, and `a2claude`'s design (Agent SDK +
   normalized event abstraction, permission-approval mapped to A2A
   `input-required`) is the more thoughtful of the two. But both explicitly
   warn they're not production-ready: no auth, no persistent storage, no
   hardened error handling. You'd be doing the productionization work
   Anthropic hasn't done and the original authors haven't finished. Worth it
   only if interop with a *third* A2A-speaking agent (not just Hermes) is a
   real near-term need — otherwise it's spec-compliance overhead with no
   consumer to justify it yet.

4. **Write an A2A server from scratch on `a2a-python`, wrapping the Agent
   SDK — highest effort, most "correct."** `a2a-python` (1.4k stars,
   Linux-Foundation-maintained, v1.0-compliant) gives you the JSON-RPC/SSE/
   push-notification machinery for free; the actual agent execution logic is
   a thin adapter over the Agent SDK's `query()`/session API, similar in
   shape to what `a2claude` already sketched. Effort: medium-high — a solid
   week to get message/send, tasks/get, streaming, and push-notification
   webhooks all correctly implemented and tested against a real client;
   ongoing maintenance burden as A2A itself is still moving (0.x→1.0 breaking
   changes happened once already this year). Only justified if Hermes needs
   to be a generic A2A client that also talks to *other* vendors' agents
   through the same code path — if Claude Code is the only peer Hermes will
   ever have, this is over-engineering relative to option 1/2.

5. **`claude mcp serve` / MCP-shaped exposure — wrong tool for this job.**
   Exposes Claude Code's individual tools (Read/Edit/Bash), not
   "delegate a task, get a result." Only relevant if Hermes wants granular
   tool calls rather than task delegation; not a fit for "task Claude Code
   and get a result back" as framed in this brief.

**Bottom line for this repo, given ponytail-style triage:** option 1 is
almost certainly the right next step — it reuses trust and tooling this repo
already has, and is a same-day change. Escalate to option 2 only if polling/
holding an SSH session proves operationally painful. Options 3–4 (real A2A)
only earn their cost the day a *non-Hermes, non-Anthropic* agent needs to
talk to this Claude Code instance too — that's the actual reason A2A's
interop guarantee would matter here, and that need doesn't exist yet.

---

## Sources

### Fetched (web), 2026-08-07

- https://a2a-protocol.org/latest/topics/agent-discovery/ — well-known path, alternate discovery mechanisms
- https://a2a-protocol.org/latest/specification/ — RPC methods, transports, auth, streaming, push notifications, version 1.0.0
- https://a2a-protocol.org/latest/topics/a2a-and-mcp/ — official A2A/MCP relationship framing
- https://github.com/ericabouaf/claude-a2a — community Claude-Code-as-A2A-server wrapper
- https://dev.to/kanywst/a2claude-turn-claude-code-into-a-server-other-ai-agents-can-call-1mf6 — a2claude design writeup
- https://mcpservers.org/servers/GongRzhe/A2A-MCP-Server — MCP→A2A client bridge (opposite direction)
- https://code.claude.com/docs/en/headless — official headless-mode docs (redirected from docs.claude.com)
- https://code.claude.com/docs/en/agent-sdk/overview — official Agent SDK overview (redirected via platform.claude.com)
- WebSearch: "A2A protocol agent card well-known path 2026 a2a-protocol.org"
- WebSearch: "Anthropic Claude Code A2A protocol support Agent2Agent"
- WebSearch: "A2A protocol MCP relationship agent-to-agent tools comparison"
- WebSearch: "site:docs.anthropic.com A2A protocol OR Agent2Agent" (no hits from that domain)
- WebSearch: "\"agent.json\" renamed \"agent-card.json\" A2A protocol breaking change well-known"
- WebSearch: "a2a-python SDK github a2aproject maturity stars" — 1.4k stars, v1.0-compliant, Linux Foundation
- WebSearch: "Claude Agent SDK build long-running agent server headless mode documentation"
- WebSearch: "\"claude mcp serve\" Claude Code MCP server mode expose tools documentation" — surfaced steipete/claude-code-mcp

### Local commands run (this machine, `cooper`), 2026-08-07

```
claude --help
claude mcp --help
claude agents --help
ls -la ~/.claude/
find ~/.claude -maxdepth 3 (daemon, ide, sessions dirs)
tail -60 ~/.claude/daemon.log
python3 -c "..." (parsed ~/.claude/daemon/roster.json structure)
file ~/.claude/daemon/control.key ; wc -c ~/.claude/daemon/control.key   # size/type only, content never read/printed
cat ~/.claude/sessions/2546713.json   # own-user session state, no secrets
ss -tlnp   # TCP listeners, no claude/cc-daemon network binding found
ss -xlp | grep claude   # Unix-domain sockets: /run/user/1000/cc-socks/<pid>.sock, one per live session
```

Plus the live `SendMessage` tool schema (fetched via `ToolSearch
select:SendMessage,ListAgents` — `ListAgents` itself is a different deferred
tool not separately fetched, but its behavior is fully described in
`SendMessage`'s own schema text, which was sufficient for this probe).

### Staleness flags

- A2A protocol is mid-migration from 0.x → 1.0; several search-result blog
  posts (not cited above as primary sources) still show the deprecated
  `/.well-known/agent.json` path — don't trust any A2A tutorial dated before
  ~2026 without cross-checking against `a2a-protocol.org/latest/`.
- Community bridge repos (`claude-a2a`, `a2claude`) are both small,
  single/few-author projects — re-check their READMEs before relying on any
  specific claim above (e.g. the exact agent-card URL `claude-a2a` serves),
  since they may have changed since this probe.
- Anthropic ships CLI/SDK updates roughly weekly (`daemon.log` on this
  machine shows 4 version bumps — 2.1.217→2.1.220→2.1.223 — in under a
  month); re-run `claude --help` before trusting any flag/behavior claim
  above if this probe is read more than a few weeks after 2026-08-07.
