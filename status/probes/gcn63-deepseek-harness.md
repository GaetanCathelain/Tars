# GCN-69 / GCN-63 probe — deepseek-ai/deepseek-harness as a plugin-native lane transport

Researched 2026-08-16. Primary sources only: GitHub API (`gh api`), raw file
fetches from `raw.githubusercontent.com`, npm registry. No WebFetch-summarized
claims are relied on below without being cross-checked against the raw source
— an initial WebFetch pass reported a plausible-sounding but wrong star count
and no multi-model detail; every number and claim here is instead read
straight from `gh api` / raw file content.

## Verdict

**Exists as described: PARTIALLY.**

- The repo is real, large, MIT-licensed, genuinely plugin-native (Cordis:
  "everything is a plugin"), and genuinely new — this matches Gaetan's
  framing.
- **Multi-model via subscriptions is only half true**, and the two halves
  must not be conflated:
  - As dsh's own model backend, every provider (including the DeepSeek-native
    adapter) authenticates via **API key** (`ANTHROPIC_API_KEY`,
    `OPENAI_API_KEY`, `DEEPSEEK_API_KEY` env vars). There is **zero** code or
    doc reference to OAuth or "subscription" anywhere in the repo (`gh api
    search/code` for `oauth`, `subscription`, `kimi`, scoped to this repo,
    each returned `total_count: 0`).
  - Separately, dsh ships **subagent-delegation** providers
    (`subagent-claude-code`, `subagent-codex`) that spawn the *native*
    `claude` / `codex` CLI via the official Claude Agent SDK / Codex
    app-server, and explicitly **inherit the host machine's existing CLI
    login state** (OAuth/subscription) rather than requiring an API key. This
    is the mechanism that would let the Claude leg ride Gaetan's existing
    subscription — but it is architecturally identical to what herdr/orca
    already do (drive the native CLI, inherit its session): dsh adds a plugin
    wrapper around the same trick, not a new capability.
- **Kimi: not supported at all.** No `kimi` reference anywhere in the repo
  (code search, `packages/llm/*`, `packages/subagent/*`). Contradicts the
  pitch's "Claude / ChatGPT / DeepSeek V4 / Kimi k3" claim.
- **Push return path: real primitive, wrong transport for this use case.**
  The plugin API genuinely has push-out extension points (`agent.inject()`,
  ACP `session/update` notifications) — but every one of them assumes a
  **live in-process or piped connection**, not a message arriving from an
  external process/machine. There is no documented "receive an external
  socket/HTTP/file message and turn it into a session event" surface. Using
  dsh's push events across the Tars-VM ↔ cooper boundary would require
  building a stdio↔network bridge around the ACP server — new code, not a
  drop-in.

## 1. What it is

- Repo: `github.com/deepseek-ai/deepseek-harness`, org `deepseek-ai`
  (verified via `gh api repos/deepseek-ai/deepseek-harness`).
- `created_at: 2026-08-13T11:56:32Z`, `pushed_at: 2026-08-13T13:00:21Z` — as a
  *public GitHub object* it is 3 days old at research time (2026-08-16).
- **But** it already has 12,293 commits and 22 contributors
  (`gh api repos/.../commits` Link header page count; `gh api
  repos/.../contributors`) — this is an internal repo dropped onto GitHub
  wholesale, not something built from scratch in public over 3 days. Treat
  "age" as "public age," not "engineering maturity." No GitHub Releases and
  no tags exist yet (`gh api repos/.../releases` → `[]`, `.../tags` → `[]`) —
  there is no versioned artifact to pin against on GitHub itself.
- Stars: 122,877; forks: 12,143; watchers: 501 subscribers (as of research
  time) — unusually fast traction for a 3-day-old public repo, consistent
  with a hyped big-lab drop rather than organic growth. Flagging as an
  anomaly worth discounting slightly, not a red flag on the code itself.
- `homepage`: `https://deepseek.com/harness`. `has_issues: false`,
  `has_pull_requests: false` — GitHub-native issue/PR intake is closed;
  README instead directs feedback to GitHub **Discussions** and a Discord.
- License: MIT (top-level `LICENSE` / GitHub API `license.key: mit`); one
  early pre-release npm manifest (`0.0.1-rc.1`) self-declared
  `BSD-3-Clause` — likely stale metadata from before the license settled,
  worth re-checking before depending on the license, not a real blocker.
- Install: `npx @deepseek-ai/dsh web` (published on npm — verified via
  `registry.npmjs.org/@deepseek-ai/dsh`, `dist-tags.latest: 0.1.0-rc.6`,
  first version `0.0.1-rc.1`), or build from source
  (`pnpm install && pnpm run build && pnpm dsh web`).
- Modes: **Web UI** (`dsh web`, serves `http://127.0.0.1:3080`) and
  **headless one-shot** (`dsh --profile headless "task"` — see §4). No
  TUI/pane mode is documented; the two shipped profile templates are `web`
  and `headless`.
- Explicit status: **"developer preview... THERE WILL BE
  COMPATIBILITY-BREAKING CHANGES."** (README, verbatim.)

Sources:
`https://raw.githubusercontent.com/deepseek-ai/deepseek-harness/master/README.md`,
`gh api repos/deepseek-ai/deepseek-harness`,
`gh api repos/deepseek-ai/deepseek-harness/releases|tags`,
`https://registry.npmjs.org/@deepseek-ai/dsh`.

## 2. Plugin API depth

Framework: [Cordis](https://github.com/cordiverse/cordis) — plugins
contribute services, typed events, and reversible effects to a shared
`Context`. Per `docs/architecture.md`: "There is no privileged core to
patch: you extend dsh by mounting a plugin beside the others."

**(a) Inject an external message as a user turn / steering input into a
running session — yes, but in-process only.**
- `agent.inject()` — documented extension point ("Add model-facing context
  → call `agent.inject()`; it lands in the next admitted request",
  `docs/architecture.md`).
- The `hooks-claude-code` bridge (a Claude-Code-hooks-config compatibility
  shim, see below) uses exactly this for `SessionStart`/`SubagentStart`
  context injection, and its own doc is explicit about the limit: *"a remote
  child has no local injection target."* Injection is a same-process API
  call, not a message received over a socket.
- The ACP server (`packages/acp/acp`, `@deepseek-ai/dsh-acp`) is the closest
  thing to an external-message inlet: a JSON-RPC-over-**stdio** server.
  `session/prompt` concatenates text blocks into one user message on a live
  agent and blocks until the agent goes idle. This is a genuine "inject a
  turn from outside the process" path — but the outside is another process
  on the *other end of a stdio pipe you spawned*, not a network peer.

**(b) Emit events outward on turn completion/question — yes, via two
mechanisms, both in-process/piped:**
- ACP `session/update` — emits one `agent_message_chunk` notification per
  committed assistant message over the *same* open JSON-RPC/stdio
  connection. `session/request_permission` similarly offers a one-shot
  question channel back to the client. This is a real push channel, but it
  requires holding a live stdio connection to a spawned dsh process — there
  is no HTTP/webhook/socket variant documented.
- `ctx.jobs` (background-job family, `packages/jobs/`) gives long-running
  tools "observation, cancellation, waiting, and completion notices" — again
  an in-process capability consumed by other in-process plugins/tools, not
  an external push.
- Nothing resembling "POST to a configured URL when a turn completes" exists
  in the repo (no webhook/callback-URL config surface found in
  `docs/config-catalog.md` skim or package tree).

**(c) Run background services — yes.** `ctx.jobs` is exactly this: a
documented extension point ("Add background work → register on `ctx.jobs`;
`job_*` tools collect or stop it," `docs/architecture.md`). Also `ctx.commands`
("Add a human command → register on `ctx.commands`; it dispatches without a
model turn") is a second non-model-turn entry point, again in-process.

**Net for GCN's push-return-path question:** dsh does have a real,
documented push-events primitive (ACP `session/update`), which is one more
than any currently-evaluated candidate (A2A wrapper, herdr, `orca terminal
send`, lean SendMessage all lack this). But it is scoped to a stdio JSON-RPC
connection to a process you spawned and keep a pipe open to — using it from
Tars (remote VM) into a lane on cooper means either (i) running dsh directly
on cooper and bridging its stdio ACP server over ssh/a socket proxy into
something Tars can poll/push to, or (ii) writing a new network-facing plugin
against `ctx.agents`/`ctx.commands` that dsh doesn't ship. Either way this is
new integration work, not a drop-in transport.

Sources: `docs/architecture.md`, `packages/acp/acp/README.md`,
`packages/jobs/README.md`,
`packages/hooks/hooks-claude-code/README.md`.

## 3. Model / subscription support

Checked via `gh api search/code` scoped to this repo (`oauth`,
`subscription`, `kimi` → all `total_count: 0`) plus direct reads of the
`packages/llm/*` and `packages/subagent/*` package READMEs.

| Provider | Mechanism in dsh | Auth |
|---|---|---|
| DeepSeek | Native adapter `packages/llm/llm-deepseek` (`deepseek-official` route) | `DEEPSEEK_API_KEY` env var, resolved via `ctx.credentials` |
| Claude/Anthropic (as dsh's own model) | Generic `packages/llm/llm-pi-ai` adapter (wraps `@earendil-works/pi-ai`), route `anthropic` | `ANTHROPIC_API_KEY` env var — API key only, no OAuth |
| OpenAI/ChatGPT (as dsh's own model) | Same `llm-pi-ai` adapter, route `openai` | `OPENAI_API_KEY` env var — API key only |
| Kimi | **No adapter, no route, no reference anywhere in the repo.** | N/A — not supported |
| Claude, as a **delegated subagent task** | `packages/subagent/subagent-claude-code` — spawns the official `@anthropic-ai/claude-agent-sdk` (pinned `0.3.220`) driving the native `claude` CLI resolved from `PATH` | **Inherits the host's existing `claude` CLI login/account state** ("reads the host's normal user, project, and local Claude settings... including native account state... does not create or modify login state"). No API key needed if `claude` is already OAuth-logged-in on that host. |
| ChatGPT, as a **delegated subagent task** | `packages/subagent/subagent-codex` — spawns `codex app-server --stdio` | Same pattern: "uses the host's native Codex configuration and authentication." Inherits Codex CLI's ChatGPT-subscription login if present. |

**Claude-subscription-auth verdict (load-bearing for Gaetan):** dsh's own
model-routing layer (`llm-pi-ai`) does **not** support Anthropic subscription
OAuth — API key only, same constraint as any other library wrapping the
Anthropic Messages API. The *only* way dsh touches an OAuth-subscription
Claude session is via `subagent-claude-code`, which is explicitly a
one-shot, no-continuation, "final text only" task delegation to a
**separately-installed, separately-logged-in** `claude` CLI on the same
host — architecturally the same trick orca/herdr already use (drive the
native CLI process, ride its login), just repackaged as a Cordis plugin.
It buys nothing new on the auth question; the constraint ("Gaetan's Claude
access is OAuth subscription, no ANTHROPIC_API_KEY") is respected only
insofar as dsh delegates out to a `claude` CLI that is itself already
logged in — same requirement as today.

ToS angle: not separately investigated beyond what's implied by the docs —
`subagent-claude-code`'s own README calls out "The project owner's
identity-scoped distribution authorization covers the official SDK and the
official CLI/platform payloads declared by each SDK version," i.e. DeepSeek
is relying on the official Anthropic SDK's own terms, the same terms orca/
herdr already operate under. No new ToS exposure beyond what's already
accepted for the proven candidates.

Sources: `packages/llm/llm-pi-ai/README.md`, `packages/llm/llm-deepseek/README.md`,
`packages/subagent/subagent-claude-code/README.md`,
`packages/subagent/subagent-codex/README.md`.

## 4. Ops fit

- **Headless one-shot mode exists and is directly comparable to `claude
  -p`:** `dsh --profile headless "task"` runs `dsh-headless` bundle
  (`packages/bundle/headless`) — no HTTP server, no listening port, submits
  the task as one user message, waits for quiescence, writes the last
  non-empty assistant text to stdout, exits 0/1 on success/failure. This is
  the mode that would actually be launchable in an Orca pane like any other
  CLI, matching Gaetan's framing.
- **Web mode** (`dsh web`) is a local server + browser app, port 3080 by
  default — not pane-friendly, not the relevant mode for a lane.
- Config: `cordis.yml` (bundle base) + `cordis.patch.yml` layered overlays
  (profile → home → `--patch` flag), YAML. `dsh --profile web
  --dump-config` prints the resolved composition.
- Instability: no releases/tags on GitHub at all; the npm package is on
  `0.1.0-rc.6` (still release-candidate). Combined with the explicit
  "compatibility-breaking changes" warning and zero GitHub Releases to pin
  a stable point against, this is pre-alpha-grade churn risk for anything
  built against it today.
- **Smoke install attempted, not trivial — abandoned per the ticket's own
  "skip freely if not trivial" clause.** `npx --yes
  @deepseek-ai/dsh@0.1.0-rc.6 --version`, isolated `NPM_CONFIG_CACHE` under
  the scratchpad, no global install, no sudo. Two attempts, each ran past
  60-90s without printing a version string; `ps` during the second attempt
  showed `npm exec @deepseek-ai/dsh@0.1.0-rc.6` pinned at ~103% CPU and
  ~1.9GB RSS, still resolving/building minutes in (this is a 117k-line
  monorepo package, not a single binary) — killed rather than let it run
  further. **Unverified**, not failed: the package is confirmed real and
  installable via npm registry metadata (§1), but a version-print smoke
  test is not the "trivial, single binary or npx into /tmp" case the ticket
  scoped in — it's a real build/resolve step, consistent with §4's
  "developer preview" / pre-release-churn read. Spec for a real smoke test:
  clone from source (`pnpm install && pnpm run build`) or accept a multi-
  minute cold `npx` resolve, then run `dsh --profile headless "echo hi"`
  and confirm stdout/exit code — budget 5-10 minutes, not attempted here.

Sources: `packages/bundle/headless/README.md`, `docs/architecture.md`
("Profiles and bundles" section), `registry.npmjs.org/@deepseek-ai/dsh`.

## 5. Comparison row

| Candidate | Push return path | Auth for Claude leg | Multi-model | Maturity / risk | What it uniquely costs |
|---|---|---|---|---|---|
| A2A wrapper | No (poll/`hermes send`) | N/A (Hermes-side) | N/A | Hermes already serves it — proven | — |
| herdr raw CLI | No | Rides host `claude` CLI OAuth | Claude-only (pane driver) | Proven | — |
| `orca terminal send` | No | Rides host `claude` CLI OAuth | Claude-only | Proven | — |
| lean `claude -p` SendMessage | No | Rides host `claude` CLI OAuth | Claude-only | Proven | — |
| **deepseek-harness (dsh)** | **Partial**: real ACP push-event primitive (`session/update`), but stdio-only — needs a new stdio↔network bridge to reach a remote Tars | Its own model layer: API-key only (no OAuth). Its `subagent-claude-code` path: rides host `claude` CLI OAuth, same as the proven candidates | DeepSeek native + Claude/OpenAI via API key (own model layer); Claude/ChatGPT via subagent-delegation (rides native CLI login). **No Kimi**, contra the pitch | 3-day-old public repo (but 12k commits of prior private history), no GitHub releases/tags yet, npm still at rc, explicit "breaking changes" warning | One more moving part (a whole Cordis harness + a plugin to write for the network bridge); unproven in this repo; the "N models via one harness" promise is weaker than pitched (no Kimi, and the Claude/ChatGPT legs are subagent delegation, not native routing) |

**Bottom line for GCN-69:** dsh is real and does have a genuinely deeper
plugin/event model than any proven candidate — but the one property this
research track cares about most (a push return path Tars could use across
the VM↔cooper boundary) is not turnkey: it exists as an in-process/stdio
primitive that still needs a bridge to be reachable at all, which is exactly
the kind of new-moving-part cost the ticket asked to weigh. The multi-model
pitch is also weaker than described once model-routing (API-key-only) is
separated from subagent-delegation (rides native CLI logins, same trick
already proven) — and Kimi isn't supported at all. Recommend: do not adopt
for V2 (Claude-only, needs the transport now); worth a second look if V2+
multi-model work happens AND dsh reaches a tagged release, since the
subagent-delegation pattern for Claude/ChatGPT costs nothing extra on the
auth question over what's already proven.
