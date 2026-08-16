# GCN-63/66 — A2A & agent-interop protocols: fit for Tars↔lane

Researched 2026-08-16. Scope, per Gaetan mid-run: (1) three lane harnesses —
Claude Code, OpenAI Codex CLI, OpenCode — not Claude Code alone; (2) a
Hermes-specific Tars side is fine, portability only matters on the lane side;
(3) survey prior art before recommending anything new-built — an orchestrator
driving interactive coding CLIs is not a novel problem.

**Verdict: A2A, with the Tars side already built and live — herdr as the
pragmatic near-term carrier for the lane side while A2A wrappers get built,
or as the implementation detail underneath them.** ACP is a poor fit (wrong
direction on the Hermes side, wrong shape for the use case). The Claude
Agent SDK only solves 1 of 3 harnesses. Nothing in the wider prior-art survey
(claude-squad, tmux-orchestrator, Conductor, Amp, Devin) beats herdr as a
cross-harness send+ack substrate; Orca's own `orchestration` mailbox is a
close second worth a follow-up look since it's already running on cooper.

---

## Table 1 — interop protocols/SDKs, ranked

| Option | Avoids same-uid/ssh injection hack? | Claude Code | Codex CLI | OpenCode | Hermes/Tars side | Maturity | Build size | Verdict |
|---|---|---|---|---|---|---|---|---|
| **A2A** (Agent2Agent, LF v1.0) | **Yes** — authenticated HTTP(S)+JSON-RPC/gRPC+SSE, routable cross-machine by design, no local-trust concept at all | No native support; needs a custom A2A-server wrapper bridging into the Agent SDK | No native support; needs a custom A2A-server wrapper bridging into `codex app-server`/`codex mcp-server` | No native support; needs a custom A2A-server wrapper (could bridge into `opencode serve`) | **Already built and live** — `plugins/platforms/a2a/` on Tars, enabled, serving Agent Card on `127.0.0.1:9900` today | High — LF-governed, v1.0.1, 6-language official SDKs, conformance kit, CrewAI + Google ADK first-party support, commits within the last day | **S** Hermes side (done); **S–M per lane** (~1–2 days each) to wrap each harness | **Recommended** |
| **ACP** (Zed Agent Client Protocol) | **No** — only documented transport is stdio-subprocess; remote HTTP/WebSocket is spec **work-in-progress**, so it still needs an SSH-spawned child process on cooper | Via third-party adapter (`@agentclientprotocol/claude-agent-acp`), wraps the Agent SDK, not native | Via third-party adapter (`@zed-industries/codex-acp`), wraps `codex app-server`, not native | **Native** — `opencode acp` built into the binary | Hermes ships an ACP **agent** (`hermes acp`), i.e. the wrong direction — Tars needs to be the *client* driving lanes, and no Python ACP client library exists | Medium-high protocol maturity (JetBrains co-lead, agent registry, fast releases) but editor-shaped: permission model is "options for the user," no agent-identity/discovery concept | **S** cooper/lane side (off-the-shelf adapters); **M** Tars side (hand-roll a Python ACP client, no library exists) | Rejected for this use case |
| **Claude Agent SDK** | **No** — SDK is in-process/subprocess-only; Anthropic's own hosting doc says to build your own HTTP/WS wrapper yourself | **Native**, and it's the *recommended* mode (`streamInput()`/`ClaudeSDKClient.query()` — mid-run message injection is first-class) | N/A — doesn't touch Codex | N/A — doesn't touch OpenCode | Would still need a hand-built wrapper server on cooper that Tars talks to over the network | High for Claude Code specifically; irrelevant elsewhere | **M** (a wrapper server: streaming bridge + permission-relay channel + session lifecycle) | Only solves 1 of 3 harnesses — not a general answer, but a strong *implementation detail* for the Claude Code leg of an A2A wrapper |
| **MCP-as-agent-server pattern** | Depends on transport chosen (usually still local/stdio for these CLIs) | `claude mcp serve` exposes Claude Code's *tools* (Bash/Read/Write…), not itself as one callable "run a task" tool | `codex mcp-server` — **native**, exposes `codex`/`codex_reply` as callable tools, closest of the three to "call this harness as an agent" | MCP client only; MCP-server role needs a third-party wrapper (`opencode-mcp`) | Hermes already has an MCP **client** config surface (`mcp_servers:` in config.yaml) — consuming, not the interop question | Real, documented pattern (mcp-agent SDK, Microsoft Agent Framework) | S per leg where native (Codex), M elsewhere | Worth using opportunistically (esp. Codex's native `mcp-server`), not as the general cross-harness answer — MCP's tool-call model lacks A2A's first-class mid-task "I need input" state |

## Table 2 — prior-art orchestrators (send+ack substrates), ranked

| Tool | Cross-harness coverage | Programmatic API / structured ack? | Still needs the same-uid/ssh hop? | Maturity | Verdict |
|---|---|---|---|---|---|
| **herdr** (`~/.config/herdr/herdr.sock`, `claude-code-herdr-plugin` v1.3.0, installed on cooper) | **Native integration hooks for all three** — plus Pi and **Hermes itself** (`~/.hermes/plugins/herdr-agent-state/__init__.py`) | **Yes** — JSON-RPC-over-Unix-socket daemon; the Claude Code plugin's `codex.py` wraps it into structured JSON verdicts (`completed`/`awaiting_clarification`/`awaiting_approval`/…), with spawn-readiness handling, verified sends, completion markers, auto-approve, self-close | **Yes** — socket is `0600`, same-OS-user only; Tars (a different machine) still needs `ssh cooper …` to reach it, landing as uid 1000, exactly the dependency `docs/gcn50-knowledge-ledger.md` §4.2 already documents for the rejected raw-socket/Go-bridge patterns | Active, versioned product (not a bespoke reverse-engineered frame format); a **live GCN-68 probe is separately validating it as a lane substrate** — this file does not re-test it | **Best available near-term carrier** — doesn't structurally escape the injection-hack shape, but replaces bespoke/version-pinned wire-format archaeology with a maintained daemon that already speaks all three target harnesses |
| **Orca `orchestration` mailbox** (`orca orchestration send/check/ask/reply/inbox/dispatch/gate-create/task-create/…`, 228-command schema, already running on cooper) | Orca already treats Codex as a first-class `--agent` value on `worktree create`; OpenCode support not confirmed in this pass | **Yes** — pull-based FIFO mailbox with blocking `ask` (blocks until answered) and `check --wait`; `dispatch --inject` can push a task straight into a terminal; decision gates (`gate-create`/`gate-resolve`) map onto "Tars needs an answer before a lane proceeds" | Same shape — it's an `orca` CLI command on cooper; Tars would call `ssh cooper 'orca orchestration ask …'` | First-class, versioned Orca product feature (schema v1), not a hack | Close second to herdr — purpose-built for orchestrator↔agent messaging and already running under everything in this repo; **worth a follow-up dig into cross-harness (Codex/OpenCode) coverage before ranking above herdr** |
| **OpenCode `opencode serve`** | OpenCode only | **Yes, natively** — `POST /session/:id/message` blocks until the model responds; SSE `/event` stream for async | N/A — real HTTP server, network-reachable in its own right | Native, OpenAPI-documented | Zero-build option for the **OpenCode leg specifically** — skip herdr/ACP/A2A for this one harness if simplicity matters more than a uniform Tars-side integration |
| **claude-squad** | Claude Code, Codex, Aider, Gemini CLI, OpenCode, Amp | No — purely an interactive TUI, no socket/HTTP/RPC | N/A (no programmatic surface to reach) | Real, active project | Rejected — nothing for an external orchestrator to call; would mean screen-scraping tmux panes, i.e. re-inventing what herdr already replaced |
| **tmux-orchestrator pattern** (Jedward23/Tmux-Orchestrator and forks) | Claude-centric | `tmux send-keys` + `tmux capture-pane` polling for a marker string; `--dangerously-skip-permissions` required | N/A, but structurally worse — screen-scraping, no schema, no atomic ack | Thin (~10 commits), lightly maintained | Rejected — this **is** the already-rejected hacky pattern (worse than the raw-socket/MCP-channel approaches this ticket is trying to move past), not an improvement on it |
| **Conductor** (conductor.build) | Claude Code, Codex, Cursor — no OpenCode | Real beta cloud API (`api.conductor.build/v0`): create workspace/session, send message, poll status (`idle`/`working`/errored) | N/A — but wrong deployment model: targets Conductor's own cloud-hosted workspaces, not a self-hosted lane machine | Real product | Validates herdr's "status enum + poll" design; not usable as-is for cooper-hosted lanes |
| **Amp / Devin** | Neither drives Claude Code/Codex/OpenCode — they run their own agent | Both have strong (Amp: webhook-push) send+ack | N/A | Mature, production APIs | Irrelevant unless swapping the lane harness entirely for Amp/Devin |

---

## 1. A2A (Agent2Agent) — recommended

**Governance & maturity.** Google announced A2A April 2025, transferred spec/IP/SDKs to the **Linux Foundation** June 2025 as the standalone Agent2Agent Project. Current spec **v1.0.1** (2026-05-28); v1.0.0 GA shipped 2026-03-12. Canonical repo `github.com/a2aproject/A2A`: 25,359 stars, 2,570 forks, last push 2026-08-15 — actively maintained, not an announcement-only project. Six official SDKs (Python/JS/Java/.NET/Go/Rust) plus a conformance kit (`a2a-tck`) and validator (`a2a-inspector`), all with commits in the last 1–3 days. `a2a-sdk` on PyPI: v1.1.2, real release cadence April–July 2026.
Sources: [Linux Foundation press release](https://www.linuxfoundation.org/press/linux-foundation-launches-the-agent2agent-protocol-project-to-enable-secure-intelligent-communication-between-ai-agents), [github.com/a2aproject/A2A](https://github.com/a2aproject/A2A), [github.com/a2aproject/A2A/releases](https://github.com/a2aproject/A2A/releases), [pypi.org/project/a2a-sdk](https://pypi.org/project/a2a-sdk/).

**Transport.** Layered spec: Layer 1 = Protocol Buffer core data structures, Layer 2 = protocol-independent semantics, Layer 3 = three concrete bindings — **JSON-RPC 2.0 over HTTP(S)**, **gRPC**, and **HTTP+JSON/REST**. All three support **SSE** streaming (`tasks/sendSubscribe`/`tasks/resubscribe`) and webhook-style push notifications as an alternative to holding a stream open.
Source: [a2a-protocol.org/latest/specification](https://a2a-protocol.org/latest/specification/), [a2a.proto](https://github.com/a2aproject/A2A/blob/main/specification/a2a.proto).

**Agent Card & task lifecycle.** Published at `/.well-known/agent-card.json`; carries identity, `capabilities` (streaming/pushNotifications/extendedAgentCard), `skills`, `supportedInterfaces` (per-binding URLs), `securitySchemes`. Task states: `SUBMITTED → WORKING → {INPUT_REQUIRED | AUTH_REQUIRED | COMPLETED | FAILED | CANCELED | REJECTED}`. **`INPUT_REQUIRED` is a near-exact semantic match for "Tars asks a lane a follow-up question mid-run"**: the executing side emits a `TaskStatusUpdateEvent` with that state, Tars answers with a new `message/send` referencing the same `taskId`+`contextId`, the task resumes — no bespoke protocol needed for that specific pattern.

**Auth.** `securitySchemes` can declare API key, HTTP Basic/Bearer, OAuth 2.0, OIDC, or mTLS — plain bearer-token-in-header is enough for a Tars↔cooper link.

**Real adoption (skeptical read).** Confirmed shipped: CrewAI ships first-class A2A delegation (crew → any A2A-compliant remote agent); Google's own ADK has a documented "expose an ADK agent as A2A server" quickstart. **Flagged as unverified hype, not cited as fact**: a marketing blog's "150+ orgs in production including Microsoft/AWS/Google/…" claim has no primary source behind it.
Sources: [CrewAI A2A tutorial](https://baeseokjae.github.io/posts/crewai-a2a-protocol-tutorial-2026/) (secondary, cross-check before quoting numbers), [google.github.io/adk-docs/a2a/quickstart-exposing](https://google.github.io/adk-docs/a2a/quickstart-exposing/).

**Build size.** The official sample ([a2a-samples/helloworld](https://github.com/a2aproject/a2a-samples/blob/main/samples/python/agents/helloworld/__main__.py)) is ~120 lines for a working streaming A2A server (`AgentCard`, `AgentExecutor.execute()`/`cancel()`, `DefaultRequestHandler`+`uvicorn`). Real work is bridging each lane's actual lifecycle into `execute()` — streaming partial output as artifacts, mapping "lane asks a question" onto `INPUT_REQUIRED`, wiring `cancel()` to kill the process, picking an auth scheme. **~1–2 days per harness (S–M)**, done once per harness, not once per session. Tars/Hermes-side client work is **zero** — see §4.

**Confirms it avoids the injection hack.** A2A is authenticated HTTP(S)/JSON-RPC/gRPC/SSE with zero notion of local trust or shared uid; every identity claim goes through declared `securitySchemes`, routable cross-machine by design — this is precisely the category of thing Google/CrewAI/ADK expect for cross-vendor, cross-host agent talk.

## 2. ACP (Zed Agent Client Protocol) — rejected for this use case

**What it standardizes.** JSON-RPC 2.0; **local agents run as subprocesses of the client over stdio** — the only transport with real, load-bearing spec text today. Remote (HTTP/WebSocket) is explicitly flagged **work-in-progress** on the docs' own overview page, not a stable part of the spec as of 2026-08. Roles are asymmetric and human-shaped: "Clients... are typically code editors," and the permission primitive `session/request_permission` is documented as options "for **the user** to choose from" — a non-human client has to fake being "the user," bending rather than using the protocol.
Sources: [agentclientprotocol.com/overview/introduction](https://agentclientprotocol.com/overview/introduction), [agentclientprotocol.com/protocol/schema](https://agentclientprotocol.com/protocol/schema).

**Adapters, per harness.**
- Claude Code: **no native support** — third-party `@agentclientprotocol/claude-agent-acp` (moved from `zed-industries/claude-code-acp` → `zed-industries/claude-agent-acp` → the neutral `agentclientprotocol` org), wraps the **Claude Agent SDK**, not the CLI's own stream-json protocol.
- Codex CLI: **no native support** — third-party `@zed-industries/codex-acp`, a translator sitting in front of Codex's own `codex app-server` JSON-RPC.
- OpenCode: **native** — `opencode acp` is built into the binary itself, no separate adapter process. The one harness of the three where ACP is a genuine zero-build option.
Sources: [github.com/agentclientprotocol/claude-agent-acp](https://github.com/agentclientprotocol/claude-agent-acp), [github.com/agentclientprotocol/codex-acp](https://github.com/agentclientprotocol/codex-acp), [opencode.ai/docs/acp](https://opencode.ai/docs/acp/), [zed.dev/acp/agent/opencode](https://zed.dev/acp/agent/opencode).

**Why the Hermes side doesn't fix this.** Hermes v0.20.0 ships a real, working ACP surface (`~/.hermes/hermes-agent/acp_adapter/`, `import acp`, real `agent-client-protocol` schema types — confirmed via `hermes acp --check` → OK on the VM) — but it makes **Hermes the ACP agent**, meant to be spawned by an editor/host (`hermes acp` launched by VS Code/Zed/JetBrains, or bridged via `buzz-acp` over a Nostr relay). That's the opposite direction from what Tars needs: Tars must be the **client** driving a Claude Code/Codex/OpenCode lane as the agent. No Python ACP client library was found; Tars would have to hand-roll JSON-RPC framing, the `initialize` handshake, a `session/new`+`session/prompt` driver, and — with no spec guidance — an auto-answer policy for permission prompts designed for a human's click.

**Net:** doesn't avoid the ssh-spawned-child-process pattern (remote transport is spec-WIP), needs a bespoke Python client with no library to lean on, and the permission model has to be reinvented for autonomous use. Momentum is real (JetBrains co-lead since Sept 2025, an ACP agent registry launched Jan 2026, Zed 1.0 in April 2026) but for editor integration — not this problem.

## 3. Claude Agent SDK — strong for the Claude Code leg, not a general answer

**Streaming mid-run input is first-class and recommended**, not a workaround: `Query.streamInput()` (TS) / repeated `ClaudeSDKClient.query()` (Python) push new user messages into an already-running session, including while tools execute — the direct SDK equivalent of "Tars injects a message into a live session." `interrupt()` + a follow-up gives full redirect capability; `canUseTool` can block indefinitely for a human/Tars decision.
Source: [code.claude.com/docs/en/agent-sdk/streaming-vs-single-mode](https://code.claude.com/docs/en/agent-sdk/streaming-vs-single-mode).

**What's lost vs a TUI:** both Orca-tab visibility and human keyboard-takeover are gone by default — no attach/watch surface exists; a human can only reach a session through code you write (`canUseTool`, `AskUserQuestion`, or streamed messages).

**Network: relocated, not eliminated.** Anthropic's own hosting doc: "the subprocess itself does not listen on the network... expose an HTTP or WebSocket port... your application handles client requests and calls the SDK internally" — you still build the wrapper server on cooper that Tars talks to. Session persistence (`continue`/`resume`/`fork`, JSONL transcripts, pluggable `SessionStore`) is strong and directly covers "lane survives Tars going away."
Source: [code.claude.com/docs/en/agent-sdk/hosting](https://code.claude.com/docs/en/agent-sdk/hosting), [code.claude.com/docs/en/agent-sdk/sessions](https://code.claude.com/docs/en/agent-sdk/sessions).

**Verdict.** Only touches Claude Code — 1 of 3 harnesses, so not a general cross-harness answer by itself. Its real value here: it's the best available *implementation detail* for the Claude Code leg of an A2A wrapper (§1) — `AgentExecutor.execute()` on the Claude Code side would internally drive a `ClaudeSDKClient`, not shell out to the CLI.

## 4. Hermes/Tars side — already solved, confirmed live on the VM (read-only, 2026-08-16)

Package root: `/home/gaetan/.hermes/hermes-agent/` (Hermes Agent v0.20.0, editable install). Confirmed via `ssh gaetan@192.168.0.9 'readlink -f ~/.local/bin/hermes'` → wrapper script exec'ing `.../hermes-agent/venv/bin/python .../hermes-agent/hermes`.

**A2A — enabled and live right now.** `/home/gaetan/.hermes/hermes-agent/plugins/platforms/a2a/` (adapter.py 1272 lines, protocol.py 842, security.py 372, tools.py 595) is a full bidirectional, stdlib-only A2A v1.0 plugin: inbound Agent Card + JSON-RPC task server routing into the **live agent session** (same memory as talks to the operator), outbound tools (`a2a_discover`/`a2a_call`/`a2a_list`/`a2a_history`/`a2a_orchestrate`). Confirmed on the VM:
```
$ grep -n -A3 "a2a:" ~/.hermes/config.yaml
    a2a:
      enabled: true
      extra:
        port: 9900
$ ss -tlnp | grep 9900
LISTEN 0 5 127.0.0.1:9900  users:(("hermes",pid=1807357,...))
$ curl -s http://127.0.0.1:9900/.well-known/agent-card.json | head -c 200
{"name": "hermes-tars", "description": "Hermes Agent — a general-purpose agent reachable over A2A.", ...
```
Bound to localhost only, no token set — matches the documented safe default. This repo's own `docs/specs/tars-profile.md:135-140,196` already records this as deliberate (inbound-only; D2 defers the outbound toolset). Security posture: prompt-injection filtering on inbound text, SSRF-guarded + HMAC-signed push callbacks, per-peer or shared bearer tokens, every exchange audit-logged to `~/.hermes/a2a_audit.jsonl`. Env vars: `A2A_HOST`/`A2A_PORT`/`A2A_PEER_TOKENS`/`A2A_BEARER_TOKEN`/etc.

**ACP — also built in, wrong direction for this ticket.** `/home/gaetan/.hermes/hermes-agent/acp_adapter/` (server.py 2510 lines, `import acp` + real `agent-client-protocol` schema types). `hermes acp --help` → "Start Hermes Agent in ACP mode for editor integration (VS Code, Zed, JetBrains)"; `hermes acp --check` → OK. Wired at `hermes_cli/main.py:11052` (`from acp_adapter.entry import main as acp_main`). This is Hermes **as the ACP agent** (stdio, spawned per editor-session) — not a network agent-to-agent surface, and not the direction Tars needs (see §2).

**Plugin system.** The A2A plugin registers via a public `PluginContext` (`ctx.register_platform(...)`, `ctx.register_tool(...)`) — a third extensibility point, but moot here since A2A/ACP already exist as named surfaces rather than needing a new plugin built.

**MCP client config.** `~/.hermes/config.yaml` has a standard `mcp_servers:` stanza — real, but for consuming tool servers, not the agent-interop question.

**Bottom line:** no interop client needs to be built on the Hermes/Tars side for A2A — it's a config/enablement question (flip on the outbound toolset per D2, and/or set `A2A_HOST`/`A2A_PEER_TOKENS` if a non-localhost peer needs to reach `hermes-tars`), not new code.

## 5. herdr — prior art, best near-term carrier (live GCN-68 probe covers deeper validation)

Installed on cooper: Rust daemon multiplexing terminal panes with first-class AI-agent awareness, JSON-RPC line-delimited over `~/.config/herdr/herdr.sock` (mode `0600`). Driven from Claude Code via the installed `claude-code-herdr-plugin` (v1.3.0), whose one tool `scripts/codex.py` wraps the raw socket into structured JSON verdicts (`completed`/`awaiting_clarification`/`awaiting_approval`/`permission_gate`/…), handling spawn-readiness races, verified single-line sends, full-width capture, marker+`--expect`-file completion, plan-never-truncated, and self-closing panes.
Sources (local, read 2026-08-16): `~/.claude/plugins/marketplaces/claude-code-herdr-plugin/README.md`, `skills/claude-to-codex/SKILL.md`, `skills/claude-to-codex/references/architecture.md`.

**Cross-harness coverage, confirmed from the architecture doc's integration-hook list:**
```
~/.pi/agent/extensions/herdr-agent-state.ts
~/.claude/hooks/herdr-agent-state.sh
~/.codex/herdr-agent-state.sh
~/.config/opencode/plugins/herdr-agent-state.js
~/.hermes/plugins/herdr-agent-state/__init__.py
```
All five agent types (Pi, Claude, Codex, OpenCode, **Hermes**) are natively recognized by the daemon — the exact three lane harnesses in scope, plus Hermes itself already has a registered hook.

**Does it avoid the injection hack? No — same shape, better substrate.** The socket is `0600`, same-OS-user only (per `architecture.md`: "any process running as the same user can talk to the server... single user, anything-with-uid-can-do-anything"). Tars, on a different machine, still needs `ssh cooper …` to reach it — landing as uid 1000, precisely the dependency `docs/gcn50-knowledge-ledger.md` §4.2 already documents for the raw-socket/Go-bridge approaches this ticket is trying to move past. herdr does **not** structurally escape that pattern. What it changes: instead of a bespoke, version-pinned, reverse-engineered frame format (the raw-socket hack's own §4.1 risk — "no compatibility promise at all"), it's a maintained, versioned product with a documented API, an active Claude Code plugin, and (per the coordinator) a live GCN-68 probe underway to validate it as a lane substrate.

**A workable shape that removes Tars' own ssh dependency:** run one persistent process on cooper (not Tars itself) that holds the herdr connection and exposes an A2A server (§1) on its network side — Tars talks A2A over the network to that process, and that process is the only thing that ever needs same-uid herdr access. This collapses "Tars needs ssh into cooper" into "one long-lived cooper-side bridge needs herdr," which is a materially smaller trust surface and composes cleanly with the A2A recommendation in §1.

## 6. Orca `orchestration` mailbox — prior art, worth a follow-up look

`orca` is already installed and running on cooper (`/home/gaetan/.local/bin/orca`, confirmed via `orca --help`; command schema pulled via `orca agent-context --json`, 228 commands, schema v1). It has a dedicated, pull-based `orchestration` command group — not a hack, a first-class product feature:

| Command | What it does |
|---|---|
| `orchestration send --to <run:id\|dispatch:id> --subject … --body …` | Send an inter-agent message |
| `orchestration check [--wait] [--ack <id>]` | Pull the bound Run's oldest unacknowledged FIFO batch; `--wait` blocks until a match or timeout, with JSON keepalives |
| `orchestration ask --question … [--to <run:id>]` | **Blocks until answered** — a direct analog of a synchronous Tars↔lane question round-trip |
| `orchestration reply --id <msg_id> --body …` | Reply to a message |
| `orchestration dispatch --task <id> --to <handle> [--inject]` | Dispatch a task to a terminal — `--inject` can push it directly into a running terminal, not just the mailbox |
| `orchestration gate-create/-resolve` | A decision gate blocking a task pending an external answer |
| `orchestration run-create/-list/-show`, `task-create/-list/-update` | Run/task bookkeeping around the above |

This is pull-based for most of the surface (`check`, `ask`, `inbox`) — the receiving side polls or blocks on its own mailbox rather than having text injected into a private session socket, which is a meaningfully different (more principled) shape than the rejected injection pattern for the parts of the API that don't use `--inject`. `orca worktree create --agent codex --prompt "hi"` (from `orca --help`'s own examples) confirms Codex is already a first-class agent choice for Orca-managed worktrees; OpenCode coverage was **not confirmed** in this pass — would need a direct check of `orca worktree create --help`/`orca agent-context --json`'s `--agent` enum values.

**Still needs the same hop as herdr:** it's a CLI on cooper; Tars would call it via `ssh cooper 'orca orchestration ask …'`. Ranked just behind herdr here only because cross-harness (Codex/OpenCode) coverage of the *mailbox* specifically (as opposed to worktree creation) wasn't fully verified in this pass, and because herdr already has a live validation probe (GCN-68) in flight — **recommend a short follow-up check of Orca's orchestration mailbox against Codex/OpenCode worktrees before ruling it out**, since it may be the least-new-code option of everything surveyed (nothing to install, nothing to enable — it's already how this repo's own worktree agents get instructed today).

## 7. Other prior-art orchestrators surveyed, briefly rejected

- **claude-squad** (`github.com/smtg-ai/claude-squad`) — real, tmux+worktrees, genuinely multi-harness (Claude/Codex/Aider/Gemini/OpenCode/Amp) but **purely an interactive TUI**: no socket, HTTP, RPC, or completion signal an external process could call. Would mean re-inventing screen-scraping that herdr already replaced.
- **tmux-orchestrator pattern** (`Jedward23/Tmux-Orchestrator` and forks, 1.8k★ but ~10 commits, thin) — `tmux send-keys` + `tmux capture-pane` polling for a marker, `--dangerously-skip-permissions` required. This **is** the already-rejected screen-scraping shape, not an improvement — no schema, no atomic ack, brittle to prompt-format changes.
- **Conductor** (`conductor.build`) — real macOS app with a genuine beta cloud API (`api.conductor.build/v0`: create workspace/session, send message, poll `idle`/`working`/errored status). Validates "status enum + poll" as the right ack shape (same as herdr's design), but targets Conductor's own cloud-hosted workspaces only, no OpenCode, no webhook — wrong deployment model for a self-hosted lane machine.
- **Amp** (Sourcegraph) and **Devin** (Cognition) — both have strong, even push/webhook-based send+ack APIs, but drive their own proprietary agents, not Claude Code/Codex/OpenCode. Only relevant if swapping the lane harness entirely.

## 8. Per-harness native surfaces worth using opportunistically

Not general cross-harness answers, but genuine zero-build options for one leg each, and reasonable implementation details inside whichever wrapper (A2A or herdr-backed bridge) ends up carrying that harness:

- **OpenCode**: `opencode serve` — native OpenAPI-documented HTTP server (the TUI is itself just a client of it); `POST /session/:id/message` blocks until the model responds, `GET /event` gives an SSE stream. [opencode.ai/docs/server](https://opencode.ai/docs/server/)
- **Codex CLI**: `codex mcp-server` — native, exposes `codex`/`codex_reply` as MCP tools other MCP clients (including Hermes, which already has an MCP-client surface) can call directly; separately, `codex exec --json` gives a one-shot JSONL event stream, and `codex app-server` (JSON-RPC over stdio or a unix socket) is Codex's own persistent bidirectional protocol underneath both the VS Code/JetBrains integrations and `codex-acp`.
- **Claude Code**: `claude mcp serve` exposes Claude Code's *tools* (Bash/Read/Write/…) as MCP, not itself as one callable "run a task" tool — less directly useful than Codex's equivalent; the Agent SDK (§3) is the stronger native hook for this harness specifically.

## 9. Other agent-interop initiatives (brief, per original scope)

- **AGNTCY** (Cisco → Linux Foundation, donated 2025-07-29) — broader than A2A: agent *discovery* (directories), *observability*, and its own SLIM transport option; explicitly designed to interoperate with, not replace, A2A and MCP. Not a competing transport choice for Tars↔lane — an adjacent discovery layer that could sit above A2A later, not something to build against instead of it.
- **LangGraph's Agent Protocol** — a real, framework-agnostic REST spec, but never got A2A's multi-vendor governance or SDK breadth, and LangChain's own newer material treats MCP as the default interop layer instead. Not a serious alternative here.
- **MCP-as-agent-server pattern** — real (mcp-agent SDK's `@app.async_tool`, Microsoft Agent Framework examples), but MCP's request/response-with-polling shape lacks A2A's native `INPUT_REQUIRED` mid-task state; would need to bolt on a second tool call or notification for "I need more input from you."

---

## Recommendation

1. **Near term**: let the GCN-68 herdr probe land, then wire a single persistent cooper-side bridge process that (a) holds herdr access to drive Claude Code/Codex/OpenCode panes and (b) exposes an A2A server on its network side. Tars talks A2A to that one process — no ssh dependency of its own, no bespoke socket-frame reverse-engineering.
2. **Medium term**: replace herdr-as-transport for each harness with that harness's own best native surface inside the same A2A wrapper shell — Claude Code via the Agent SDK's `streamInput`, Codex via `codex app-server`/`mcp-server`, OpenCode via `opencode serve` — once each is worth the dedicated build.
3. **On the Hermes/Tars side, nothing needs building** — A2A is already live; enabling the outbound toolset and setting a peer token/host is a config change, not code (D2 in `docs/specs/tars-profile.md`).
4. **Do not build against ACP** for this problem — wrong direction on the Hermes side, no Python client library, remote transport is spec-WIP, and 2 of the 3 harnesses need a third-party translator anyway.
5. **Follow-up**: confirm whether Orca's `orchestration` mailbox already covers Codex/OpenCode worktrees the same way it covers Claude Code's — if so, it may be the cheapest of everything surveyed here, since it requires installing nothing new.
