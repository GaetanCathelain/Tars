# Graph engineering — research report (2026-08-07)

Context: before designing the Tars build plan, we researched "graph engineering" — orchestrating
LLM agent runs as explicit graphs (nodes = agent/tool steps, edges = control flow, fan-out/fan-in,
checkpointing) — to decide what to build the Tars run graph on. Verdict first:

**Use Claude Code's native Workflow tool for session-scale graphs; Kestra for durable/scheduled
ones; skip third-party orchestrators.** The native tool is graph-as-code (`agent()` nodes,
`pipeline()`/`parallel()` edges, JSON-schema structured outputs, worktree isolation, journaled
checkpoint-resume, 16 concurrent / 1000 total agents per run). Kestra (already running gbrain
flows on cooper) covers graphs that must outlive a session. Third-party Claude Code orchestrators
are increasingly thin wrappers over the native features.

## 1. Is "graph engineering" a real emerging term? — yes, coined ~July 2026, mid-hype

- "Loop engineering" entered mainstream dev discussion June 2026; "graph engineering" followed
  ~6 weeks later. A widely-shared Peter Steinberger X post (2026-07-18) framed it as "designing
  the multi-agent organization as a programmable structure, not just programming one agent's
  behavior cycle". Provenance unresolved per MarkTechPost.
- LangChain co-opted it immediately: "3 Years of Graph Engineering with LangGraph" (2026-07-22).
- Consensus definition: **loop engineering tunes one agent's observe-act-verify cycle; graph
  engineering wires several such loops into a topology** — "you don't rewrite the loop, you put
  it in a node."
- HN traction thin (biggest thread: 24 points). Heavy blog/SEO churn, near-zero practitioner
  debate — the practice predates the label by years.

Sources: marktechpost.com/2026/07/29 (prompt vs loop vs graph engineering),
langchain.com/blog/3-years-of-graph-engineering-with-langgraph, explainx.ai graph-engineering
posts, HN Algolia search.

## 2. Framework survey (state mid-2026)

| Framework | Lang | Graph model | Checkpoint/resume | Maturity |
|---|---|---|---|---|
| LangGraph | Py/TS | cyclic StateGraph, `Send` dynamic fan-out | strongest in class: per-superstep checkpointer, threads, time-travel, Postgres/DynamoDB | de-facto standard, 65M+ dl/mo |
| Mastra | TS | step-composed (`.then/.branch/.parallel`) | `suspend()`/`resume()` storage-backed, indefinite pause | the TS batteries-included pick |
| Pydantic AI / pydantic-graph | Py | typed FSM, edges via return-type annotations | state persistence + resumable | active, deliberately low-level |
| CrewAI Flows | Py | event-driven `@start/@listen/@router` (implicit graph) | `@persist` decorator | popular; Flows is a retrofit |
| Microsoft Agent Framework | Py/.NET | explicit graph workflows (successor to AutoGen + SK) | built-in checkpoint/resume + HITL | 1.0 GA 2026-04; enterprise pick |
| OpenAI Agents SDK | Py/TS | handoffs (dynamic delegation), no declared graph | none native — pair with Temporal (GA 2026-03) or DBOS | minimal by design |
| Google ADK | Py/TS/Go/Java | Sequential/Parallel/Loop agents + LLM routing | session state via Vertex; criticized as not true durable execution | Vertex-anchored |
| Temporal | many | workflow-as-code, agents = workflows | gold standard: event-sourced replay | the durability layer others wrap |
| Inngest AgentKit | TS | network = loop + shared state + router | via Inngest durable functions | solid, smaller |
| Vercel Workflows, DBOS | TS/Py | workflow-as-code | strong | rising durability substrates |

## 3. Claude Code specifically — third-party graph layers

- **ruvnet/claude-flow → renamed `ruflo`** (~67k stars, v3 Rust/WASM rebuild): MCP server + 35
  plugins + CLI harness, GOAP planning, vector memory. Biggest and noisiest; long-standing
  community skepticism about benchmark claims.
- **oh-my-claudecode** (~38k stars): staged pipeline (plan → prd → exec → verify → fix) over
  native Agent Teams + tmux workers. Pragmatic, but a fixed pipeline, not arbitrary graphs.
- Smaller: barkain/claude-code-workflow-orchestration (~80★, wave-based plugin),
  zircote/claude-team-orchestration, mbruhler/claude-orchestration, claude-squad (~7.8k★,
  session-level not graph), open-multi-agent.
- Context: Anthropic shipped natively (Agent Teams 2026-02, Dynamic Workflows ~2026-05), so most
  third-party layers became thin wrappers. The native path ages best.

## 4. Native Claude Code / Agent SDK primitives

- **Subagents**: nodes with isolated context, own prompt/tools/model.
- **Agent Teams** (2026-02): lead + 2–16 peers, own worktrees, peer messaging, shared task list.
- **Dynamic Workflows** (research preview ~v2.1.154, with Opus 4.8): deterministic JS
  orchestration scripts — `agent(prompt, {model, schema, isolation:'worktree'})`, `parallel()`,
  `pipeline()`, nested `workflow()`. 16 concurrent / 1000 total agents. Checkpoint-based resume
  (journaled cached results) — **same-session only**. Determinism enforced (`Date.now()` /
  `Math.random()` throw).
- **SDK session control**: `resume`, `fork_session` (cheap per-node context seeding),
  `continue`; hooks for deterministic edges. Known gap: programmatic subagent resume is limited.

## What this meant for Tars

- Session-scale fan-outs (recon, verify) → native Workflow tool, in-process.
- Long-lived unattended lane (VM provisioning) → a spawned Orca session, because workflow resume
  being same-session-only makes an in-flight workflow hostage to its session's context budget.
- Human-in-the-loop steps (credential auth, destructive cutover) → never inside a workflow;
  they live in the orchestrator session, and the destructive one behind an explicit gate.
- Durable/scheduled graphs later (WF5 dailys/reminders) → Kestra territory, or Hermes-native
  scheduling — decide at WF5 design time.

Evidence caveat: the "graph engineering" content wave is heavily SEO-driven; framework facts
cross-checked against official docs/repos, adoption claims treated as directional.
