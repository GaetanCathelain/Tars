# A2A recon — Hermes inbound server + outbound toolset (Tars + p-Hermes)

Read-only recon, 2026-08-07. No config/service on either VM was changed.
All claims below cite either a source `file:line` or a live command run
from the Tars VM. VM clock is UTC.

## Correction to the task brief — flag first

The task described p-Hermes as "VM 103, `192.168.0.3`". **This is wrong.**
On the Tars VM's `~/.ssh/config`:

- `Host phermes` → `HostName 192.168.0.8`, user `hermes`
- `Host pve` → `HostName 192.168.0.3`, user `root` — this is the **Proxmox
  hypervisor**, not p-Hermes.

`ssh phermes` from Tars correctly resolves to `.8` (verified below), so no
wrong-host commands were run. But `192.168.0.3` in CLAUDE.md's "never root
on 192.168.0.3" now reads as protecting the *hypervisor*, and any future
doc/runbook that says "p-Hermes = .3 = VM 103" is stale and should be fixed
before someone acts on it.

## 1. Source location & version

- Tars: Hermes v0.20.0 (2026.8.3), install method not git-tracked, tree at
  `~/.hermes/hermes-agent` (`hermes --version` on the VM).
- A2A plugin: `~/.hermes/hermes-agent/plugins/platforms/a2a/` — 4 files,
  3219 lines total (`adapter.py` 1272, `protocol.py` 842, `tools.py` 595,
  `security.py` 372, `__init__.py` 138). Pure stdlib, no `a2a-sdk` dependency
  (`protocol.py:17-19`, `__init__.py:1-9`).
- Protocol implemented: **A2A v1.0**, JSON-RPC 2.0 binding over HTTP
  (`protocol.py:5,35`). The code explicitly tracks a spec migration — v1.0
  SCREAMING_SNAKE_CASE task states, member-presence-discriminated Parts (no
  `kind` field), `SendMessageResponse` oneof wrapper, JSON-RPC-wrapped SSE
  frames — while staying tolerant of v0.3 peer shapes on decode
  (`protocol.py:1-19`, `285-355` `extract_text`).

## 2. Inbound server

### Routes (`adapter.py:215-241` GET, `243-334` POST)

| Method | Path | Purpose |
|---|---|---|
| GET | `/.well-known/agent-card.json` | canonical v1.0 Agent Card |
| GET | `/.well-known/agent.json` | legacy alias, same payload |
| GET | `/`, `/health` | liveness + (locally) served-agent topology |
| GET | `/metrics` | counters (see below) |
| POST | `/` (or any routed agent's path prefix) | JSON-RPC 2.0 endpoint |

JSON-RPC methods handled (`adapter.py:131-161` `_method_info`, PascalCase =
v1.0 canonical, lowercase/slash = legacy alias, both accepted):
`SendMessage`/`message/send`, `SendStreamingMessage`/`message/stream` (SSE),
`GetTask`/`tasks/get`, `ListTasks`/`tasks/list`, `CancelTask`/`tasks/cancel`,
`SubscribeToTask`/`tasks/subscribe`, and the 4 push-config CRUD methods
(`CreateTaskPushNotificationConfig`, `Get…`, `List…`, `Delete…`).

### Agent Card — verbatim, live from Tars (`curl 127.0.0.1:9900/.well-known/agent-card.json`)

```json
{"name": "hermes-tars", "description": "Hermes Agent — a general-purpose agent reachable over A2A.", "url": "http://127.0.0.1:9900/", "version": "1.0.0", "provider": {"organization": "Hermes Agent", "url": "http://127.0.0.1:9900/"}, "supportedInterfaces": [{"url": "http://127.0.0.1:9900/", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"}], "capabilities": {"streaming": true, "pushNotifications": true, "stateTransitionHistory": false, "extendedAgentCard": false}, "defaultInputModes": ["text/plain"], "defaultOutputModes": ["text/plain"], "skills": [ /* 30 toolset skills — see note below */ ]}
```

Full skills array (30 toolset entries incl. `a2a`, `kanban`, `mcp-slack`,
`homeassistant`, `spotify`, `browser`, `code_execution`, `file`, `memory`,
`terminal`, `web`…) is saved verbatim at
`/tmp/claude-1000/…/scratchpad/agent-card.json` in this session (not copied
into this repo — it's a full local capture, reproducible with the curl
above).

**Card/reality mismatch (source-grounded, contradicts what the card
implies):** `_advertised_skills()` (`adapter.py:618-641`) builds the skills
list from `tool_registry.get_registered_toolset_names()` — every toolset any
plugin *registered*, not what's actually *enabled* for the model. Tars'
`config.yaml` enables only `toolsets: [hermes-cli, kanban]`
(`~/.hermes/config.yaml:147-149`). So the card advertises ~30 skills
(`spotify`, `homeassistant`, `mcp-slack`, `browser`, …) that the live agent
cannot currently act on. A peer reading the card would over-estimate Tars'
real capability by a wide margin.

`/health` (localhost, no `Authorization` header) returned
`served_agents` topology:
```json
{"status": "ok", "agent": "hermes-tars", "served_agents": [{"slug": "default", "name": "hermes-tars", "url": "http://127.0.0.1:9900/", "tenant": null, "profile": "default", "local": true}]}
```
This is intentional — `adapter.py:228-236` only hides `served_agents` from
*remote unauthenticated* callers; `security.localhost_only()` being `True`
(no token configured) always satisfies the disclosure condition locally.

`/metrics` (live): `{"uptime_seconds": 4525.2, "inbound_total": 0, "outbound_total": 0, "streams_started": 0, "push_sent": 0, "push_failed": 0, "tasks_completed": 0, "tasks_failed": 0, "anti_loop_triggers": 0, "rate_limit_triggers": 0, "avg_latency_ms": 0.0}` — zero traffic ever, as expected (nothing has called it).

### Auth — is the inbound endpoint authenticated? **No, by default.**

`security.py:78-100` `authenticate()`: if neither `A2A_BEARER_TOKEN` nor
`A2A_PEER_TOKENS` is set (Tars' current state — confirmed, see §3), **every**
request is accepted and assigned identity `ip:<client_addr>` — there is no
credential check at all. `security.py:155-170` `is_trusted_peer()`: when
`localhost_only()` is true, **every** identity is trusted — the optional
`A2A_TRUSTED_PEERS` allow-list is skipped entirely.

The real gate is **not auth, it's the bind address**: `resolve_bind_host()`
(`security.py:108-126`) forces `127.0.0.1` unless a token is configured *and*
`A2A_HOST` is explicitly widened. Confirmed live —
`ss -tlnp | grep 9900` → `LISTEN 127.0.0.1:9900 … hermes`. So today: any
process able to reach loopback on the Tars VM (any local user, any other
Hermes plugin's subprocess, a compromised MCP server, `orca`/Claude Code
sessions on the same box) can submit A2A tasks with zero credentials and
have them treated as fully trusted.

One more doc/code gap worth flagging: `__init__.py:97-138` registers
`allowed_users_env="A2A_ALLOWED_USERS"` on the platform (implying an
allow-list knob), but `adapter.py:390-403` `authorization_is_upstream = True`
makes the *gateway's* per-platform allow-list check a no-op for A2A — the
comment says this is deliberate (A2A does its own auth), but it means
`A2A_ALLOWED_USERS` is effectively dead config; the real switch is
`A2A_TRUSTED_PEERS` / `A2A_BEARER_TOKEN` / `A2A_PEER_TOKENS`.

Inbound tasks are still hardened in-band even with no auth: prompt-injection
markers are defanged (`security.py:180-201` — ChatML tags, role-prefix
lines, "ignore previous instructions" phrasing), every inbound message is
wrapped with an explicit "treat as untrusted peer input" prefix
(`security.py:206-222`), and outbound replies get credential-shaped strings
redacted (`security.py:230-249`, covers `sk-…`, `sk-ant-…`, `ghp_…`,
`xox[bap]-…`, AKIA keys, JWTs, bearer headers, emails).

**What a successful (i.e. any) inbound caller gets:** the task is injected
via `MessageEvent` into the **live gateway process** on the same profile as
the operator's own sessions (`adapter.py:1-26` docstring: "the agent that
replies is the same one talking to its user — full memory and context, not
a throwaway clone"). Concretely this means the *same toolsets* configured
for the model (`hermes-cli`, `kanban` on Tars today) and the *same memory
subsystem* (`memory_enabled: true` in config.yaml) are reachable from an
anonymous loopback caller. It is a new per-`context_id` conversation, not
literally Gaetan's Slack thread, but it runs with the operator's full tool
grant.

Streaming: yes, `capabilities.streaming: true` — `message/stream` returns
SSE, JSON-RPC-wrapped per §9.4 (`adapter.py:998-1028`, `protocol.py:425-448`).
Push notifications: yes, `capabilities.pushNotifications: true` — inline in
`message/send` via `configuration.taskPushNotificationConfig`, or the 4
CRUD methods; payloads are HMAC-SHA256 signed when a secret is configured
(falls back to the bearer token; unsigned if neither is set —
`security.py:256-279`), and callback URLs are SSRF-filtered
(`security.py:286-341`, blocks RFC1918/loopback/link-local/metadata unless
`localhost_only()`).

Other hardening present regardless of auth state: anti-loop turn cap per
`context_id` (default 5, env `A2A_MAX_PINGPONG_TURNS`, hard cap 20 —
`protocol.py:72-83`, `adapter.py:708-722`), per-identity sliding-window rate
limit (default 60/min, `A2A_RATE_LIMIT` — `protocol.py:495-519`), 1 MB max
request body (`adapter.py:65,255-258`), append-only JSONL audit log of every
inbound/outbound/push exchange (`security.py:357-372`,
`~/.hermes/a2a_audit.jsonl`), and per-context conversation persistence
surviving restarts/compaction (`protocol.py:791-843`,
`~/.hermes/a2a_conversations/<ctx>.jsonl`).

### Live RPC probes (all read-only, non-mutating)

```
POST / {"method":"bogus/method"}        -> {"error":{"code":-32601,"message":"method not found: bogus/method"}}
POST / {"method":"tasks/get","taskId":"task-doesnotexist0000"}  -> {"error":{"code":-32001,"message":"task not found: …"}}
POST / {"method":"GetTask", same id}    -> same -32001 (legacy/v1 alias parity confirmed)
POST / with body "not json"             -> {"error":{"code":-32700,"message":"parse error"}}
```
Error codes match `protocol.py:62-70` (`-32001` = A2A spec `TaskNotFoundError`,
`-32601`/`-32700` = standard JSON-RPC). No task was created by any of these
(bogus lookups short-circuit before `TaskStore.create`).

## 3. Outbound toolset (`tools.py`)

5 tools registered in the `a2a` toolset (`tools.py:585-596`,
`_SCHEMAS`/`_HANDLERS`):

| Tool | Params | What it does |
|---|---|---|
| `a2a_discover` | `url` | fetch + summarize a peer's Agent Card (name, protocol, streaming/push flags, skills) |
| `a2a_call` | `agent`, `message`, `context_id?` | send a `message/send` task to a peer (config name from `a2a_agents` or a raw URL) and return the reply text; multi-turn via `context_id` |
| `a2a_list` | — | list configured peers + persisted conversations + live metrics |
| `a2a_history` | `context_id`, `limit?` | replay a persisted conversation transcript from `~/.hermes/a2a_conversations/` |
| `a2a_orchestrate` | `capability`, `message`, `mode? (all\|first\|best)`, `context_id?` | fan-out one task to every peer in `a2a_agents` advertising a matching `capabilities` tag, parallel (`ThreadPoolExecutor`, max 6 workers) |

Peers are resolved from `config.yaml` → `a2a_agents.<name>: {url, auth:
{type: bearer, token}, timeout, capabilities}` (`tools.py:11-22`), or the
model can pass a raw `http(s)://` URL directly to `a2a_call`/`a2a_discover`
— **there is no peer allow-list on the outbound side**; if the toolset is
enabled the model can call *any* URL it likes, not just configured peers.
Outbound requests redact credential-shaped strings before sending
(`tools.py:167`, reuses `security.redact_outbound`) and are audited/persisted
identically to inbound (`tools.py:182-184`).

**Is this toolset actually available to the model on Tars right now? No.**
`config.yaml:147-149` sets the top-level `toolsets: [hermes-cli, kanban]` —
`a2a` is not in that list. The plugin still registers the 5 tools in the
registry (which is *why* the Agent Card can list them as a skill, per the
mismatch noted in §2), but per the same enable-gate pattern documented for
kanban in `docs/facts.md` ("single top-level `toolsets:` key unlocks …"),
an unlisted toolset is not exposed to the model's tool-calling. **If
enabled** (adding `a2a` to that list), Tars would gain: discover any A2A
peer's card, send it arbitrary natural-language tasks and read replies,
fan a task out to multiple capability-matched peers in parallel, and resume
multi-turn peer conversations — all via plain URLs, no pre-registered peer
required unless the model wants named/`capabilities`-based routing.

## 4. Config keys & current values

`~/.hermes/config.yaml` (verbatim relevant section, `config.yaml:118-149`):
```yaml
gateway:
  platforms:
    a2a:
      enabled: true
      extra:
        port: 9900
...
toolsets:
- hermes-cli
- kanban
```
No `a2a_agents`, `a2a_served_agents`, `platforms.a2a.extra.agents`, or
`a2a.trusted_peers` keys present anywhere in the 203-line file (grepped, no
matches). `~/.hermes/.env` has **no `A2A_*` keys at all** (grepped, empty).
So every A2A env-var default from the source applies as-is on Tars today.

Full key inventory, source-grounded:

| Key | Where | Effect | Tars current value |
|---|---|---|---|
| `gateway.platforms.a2a.enabled` | config.yaml | turns the inbound server on/off | `true` |
| `A2A_PORT` / `extra.port` | env / config.yaml | inbound bind port | `9900` |
| `A2A_HOST` | env | requested bind host (widened only if a token is also set — `security.py:108-126`) | unset → `127.0.0.1` |
| `A2A_AGENT_NAME` | env | name advertised in the card / hostname-derived default | unset → `hermes-tars` (hostname-derived, `adapter.py:77-85`) |
| `A2A_BEARER_TOKEN` | env | shared inbound bearer secret; presence alone doesn't widen bind | unset |
| `A2A_PEER_TOKENS` | env | `"name:token,…"` per-peer credentials | unset |
| `A2A_ALLOWED_USERS` | env (registered, effectively dead — see §2) | would restrict which platform-level identities can reach A2A, but bypassed by `authorization_is_upstream` | unset |
| `A2A_ALLOW_ALL_USERS` | env | forces `is_trusted_peer()` open even with a trust list configured | unset |
| `A2A_TRUSTED_PEERS` / `a2a.trusted_peers` (config.yaml) | env / config | allow-list of authenticated identities once auth is configured | unset (moot — localhost-only mode trusts everyone anyway) |
| `A2A_ADVERTISED_TOOLSETS` / `extra.advertised_toolsets` | env / config | restricts which toolsets the *card* claims (does not gate actual tool access) | unset → card shows all registered toolsets |
| `A2A_AGENT_DESCRIPTION`, `A2A_PROVIDER_ORG`, `A2A_PROVIDER_URL` | env | card cosmetic fields | unset → defaults |
| `A2A_PUBLIC_URL` | env | overrides the URL the card/RPC-interface advertise (for reverse-proxy setups) | unset → derived from `Host` header / bind |
| `A2A_REPLY_TIMEOUT` | env | seconds to block an HTTP caller waiting for the agent's reply (default 300) | unset → 300s |
| `A2A_MAX_PINGPONG_TURNS` | env | anti-loop cap per context (default 5, hard max 20) | unset → 5 |
| `A2A_RATE_LIMIT` | env | requests/min per identity (default 60) | unset → 60 |
| `A2A_PUSH_SECRET` | env | HMAC key for push payloads (falls back to bearer token; unsigned if neither set) | unset |
| `A2A_HOME_CHANNEL` | env | cron/async delivery target for A2A (mirrors Slack's home-channel pattern) | unset |
| `a2a_agents.<name>` (config.yaml, top-level) | config | **outbound** peer registry consumed by `tools.py` | absent — no peers configured |
| `platforms.a2a.extra.agents` / `a2a_served_agents` | config | **inbound** multi-agent/multi-tenant routing (serve several profiles behind one port) | absent — only the default/root agent is served |
| top-level `toolsets:` list | config.yaml | must include `a2a` to expose the outbound tools to the model | currently `[hermes-cli, kanban]` — **a2a not enabled** |

## 5. p-Hermes inventory — READ-ONLY

Reached via `ssh gaetan@192.168.0.9 'ssh phermes "<cmd>"'` (alias resolves to
`192.168.0.8`, confirmed by `~/.ssh/config`, not `.3`).

- Version: **Hermes Agent v0.19.0** (2026.7.20) · upstream `0a62610f` · local
  `d9da213a` (+3 carried commits), install method `git`, tree at
  `/home/hermes/.hermes/hermes-agent`.
- **A2A plugin does not exist in this install**: `ls
  ~/.hermes/hermes-agent/plugins/platforms/a2a/` → `No such file or
  directory`. This isn't "disabled", it's absent from source — v0.19.0
  predates (or was built without) the A2A plugin entirely.
- `~/.hermes/config.yaml` has **no `a2a` key at all** (grep empty), `.env`
  has no `A2A_*` keys.
- Port 9900: **nothing listening** (`ss -tln | grep 9900` → empty).
  `curl --max-time 3 http://127.0.0.1:9900/.well-known/agent-card.json` on
  p-Hermes returned nothing (connection refused, consistent with no
  listener).
- `hermes-gateway.service` (`--user`) is otherwise healthy: active/running
  1d22h, PID 1269, normal WARNING-level noise (an MCP server `gbrain`
  retry-failing repeatedly — pre-existing, unrelated to A2A).

### Proposal (not executed) — what would be needed to light up p-Hermes A2A

This is a written proposal only; nothing below was run.

1. **Upgrade p-Hermes to a version carrying the A2A plugin** (≥ whatever
   release Tars' v0.20.0 came from, or cherry-pick the plugin directory if
   p-Hermes's "+3 carried commits" local delta makes a full upstream bump
   risky). Without this, none of the rest is possible — the code isn't
   there.
2. After the plugin exists: add `gateway.platforms.a2a.enabled: true` (+
   `extra.port: 9900` or another free port) to p-Hermes's `config.yaml`,
   same shape as Tars §4.
3. Decide the trust model before touching `A2A_HOST` — per `security.py`,
   binding beyond loopback requires **both** a token (`A2A_BEARER_TOKEN` or
   `A2A_PEER_TOKENS`) **and** an explicit `A2A_HOST`. Given both hosts sit on
   the same private LAN (see §6), the minimal-exposure option is: set
   `A2A_PEER_TOKENS="tars:<random-token>"` on p-Hermes, keep `A2A_HOST`
   unset (stays `127.0.0.1`), and reach it over an SSH tunnel from Tars —
   or, if LAN-only exposure is judged acceptable, set `A2A_HOST=192.168.0.8`
   with the same peer token so only Tars' pre-shared credential is accepted.
4. Add `A2A_TRUSTED_PEERS=tars` so even a leaked/guessed second token
   couldn't get past the trust gate without also being on the allow-list.

## 6. Transport options — Tars ↔ p-Hermes, and Tars ↔ external

Facts first: both today bind `9900` to `127.0.0.1` only. Tars is on the
tailnet (`100.116.31.76`, alias `tars`); **p-Hermes is NOT on the tailnet**
— `tailscale status` on Tars lists no `.8`/`phermes`/`cooper`-named peer.
Both hosts ARE on the same private LAN/24 (`192.168.0.9` and `192.168.0.8`)
and Tars can already reach p-Hermes: `ping` succeeds (0.2–6.9ms), TCP/22 is
open, TCP/9900 is refused (nothing listening, not a firewall block) —
confirmed live from Tars via `/dev/tcp` probes.

| Option | Mechanism | Cost | Exposure |
|---|---|---|---|
| **SSH-tunneled localhost** | `ssh -L 9900:127.0.0.1:9900 phermes` (or a persistent `autossh`/systemd unit) from Tars, then `a2a_agents.p-hermes.url: http://127.0.0.1:9900` on Tars | one more long-lived process/unit to babysit; reuses existing SSH key trust, zero new listening surface on p-Hermes | lowest — p-Hermes A2A never binds beyond loopback; only Tars (holder of the SSH key) can reach it, and only while the tunnel is up |
| **Direct LAN bind** | p-Hermes sets `A2A_HOST=192.168.0.8` + a peer token, Tars calls `http://192.168.0.8:9900` directly | none beyond the config change itself (§5 proposal) | any host on `192.168.0.0/24` can *attempt* a connection; the peer token is the only gate — must not skip step 3/4 of the proposal |
| **Tailnet bind** | join p-Hermes to the tailnet, bind `A2A_HOST` to its `100.x` tailnet IP + token | requires enrolling a new tailnet node (out of scope of this recon — a real infra step) | narrower than LAN (tailnet ACLs apply), but wider than loopback; matches how Tars itself would be reached by an *external* A2A peer |
| **Reverse proxy** (e.g. nginx/caddy in front of 9900) | proxy terminates TLS, enforces auth/IP-allow, forwards to loopback | most operational overhead (another service to run/patch), only pays off if A2A needs to be reachable from outside the tailnet/LAN entirely (true "external" A2A peers) | best for public exposure; overkill for a two-VM home lab link |

For **Tars ↔ p-Hermes** specifically (both already on the same trusted LAN,
p-Hermes not on tailnet): SSH tunnel is the lowest-new-surface option and
needs no p-Hermes network reconfiguration beyond the plugin/config work in
§5's proposal. For **Tars ↔ an external A2A peer** (anywhere off the LAN):
tailnet bind is the natural fit since Tars is already enrolled — avoids
standing up a reverse proxy for what would otherwise be point-to-point
agent traffic.

## Sources

- `~/.hermes/hermes-agent/plugins/platforms/a2a/{__init__,adapter,protocol,security,tools}.py` (Tars VM, v0.20.0) — read in full, cited by line above.
- `~/.hermes/config.yaml` (Tars VM) — read in full (203 lines).
- `~/.hermes/.env` (Tars VM) — grepped for `^A2A`, empty.
- Live curls: `http://127.0.0.1:9900/{.well-known/agent-card.json,.well-known/agent.json,health,metrics}` and 4 non-mutating JSON-RPC POSTs, run from Tars.
- `ss -tlnp | grep 9900` (Tars) — bind confirmation.
- `~/.ssh/config` (Tars VM) — host alias table, source of the §0 correction.
- `tailscale status` (Tars VM).
- `ssh phermes` leg (from Tars): `hermes --version`, `ls .../plugins/platforms/a2a/`, `grep a2a config.yaml/.env`, `curl` to its loopback 9900, `ss -tln`, `systemctl --user status hermes-gateway.service`.
- `ping` / `/dev/tcp` reachability probes, Tars → `192.168.0.8`.
- `docs/facts.md` (this repo) — read first per task instructions, cited for the toolset-enable-gate precedent (kanban).
