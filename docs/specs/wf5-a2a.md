# WF5 — A2A: constraints, inventory, options

**Status: PROPOSAL. Nothing here has been enabled and nothing here picks a
transport.** Every decision this document surfaces is parked for Gaetan
(§ Parked decisions). Read § Standing constraints before any enablement
material — the constraints change what the options cost.

Evidence base:

- `status/probes/wf5/a2a-hermes.md` — Hermes A2A source recon (Tars + p-Hermes),
  read-only, 2026-08-07.
- `status/probes/wf5/a2a-claude-code.md` — Claude Code vs A2A, same day.
- `docs/facts.md` — binding operational truth sheet.
- `docs/recon/DECISION.md` D2, `PLAN.md` § Amendments (D6 dev. 7), `PLAN.md`
  § Phases WF5 row — the settled A2A decisions this proposal sits under.

Source citations are `file:line` in the Tars VM tree
`~/.hermes/hermes-agent/plugins/platforms/a2a/` at Hermes **v0.20.0
(2026.8.3)**. Lines marked *(re-verified)* were read again on 2026-08-07 while
writing this spec; everything else is quoted from the probe files. Re-verify
the whole document after any Hermes upgrade — it is source-version-bound.

---

## Standing constraints

### C1 — the inbound endpoint is unauthenticated; the bind is the only gate, and bind and token are coupled by design

`security.py:78-100` `authenticate()`: when neither `A2A_BEARER_TOKEN` nor
`A2A_PEER_TOKENS` is set, **every** request is admitted and assigned identity
`ip:<client_addr>`. There is no credential check at all.
`security.py:155-170` `is_trusted_peer()`: while `localhost_only()` is true,
**every** identity is trusted and the optional `A2A_TRUSTED_PEERS` allow-list is
skipped entirely (`a2a-hermes.md` §2).

Both env vars are unset on Tars today: `grep -c '^A2A' ~/.hermes/.env` → `0`
*(re-verified)*, and `config.yaml` carries no A2A token keys
(`a2a-hermes.md` §4).

**The sole gate is the bind**, and it is verified live *(re-verified)*:

```
ss -tlnp | grep 9900
LISTEN 0 5  127.0.0.1:9900  0.0.0.0:*  users:(("hermes",pid=76255,fd=27))
```

An admitted task is injected as a `MessageEvent` into the **live gateway
process**, on the operator's own profile — `adapter.py:1-26` docstring: "the
agent that replies is the same one talking to its user — full memory and
context, **not a throwaway clone**". Concretely it runs with the same toolsets
configured for the model (`hermes-cli`, `kanban` today) and the same memory
subsystem (`memory_enabled: true`). It is a fresh per-`context_id`
conversation, not Gaetan's Slack thread, but it carries the operator's full
tool grant (`a2a-hermes.md` §2).

**Nuance — accidental unauthenticated exposure is NOT reachable by config
alone.** `security.py:108-126` `resolve_bind_host()` *(re-verified verbatim)*:

> Rule: localhost unless the operator BOTH configured a token (shared or
> per-peer) AND explicitly asked for a wider host. A token alone does not
> widen the bind — opting into remote exposure must be deliberate.

With `A2A_HOST` set to a non-loopback host and no token configured, the code
logs `A2A: A2A_HOST=%s ignored — no A2A_BEARER_TOKEN or A2A_PEER_TOKENS set;
binding to 127.0.0.1` and binds loopback anyway. So the constraint is not "the
endpoint might leak"; it is that **widening the bind and setting a token are
one decision, not two** — any transport option that moves off loopback drags a
credential decision with it.

What the loopback gate does *not* protect against: any process able to reach
loopback on the Tars VM — any local user, another Hermes plugin's subprocess, a
compromised MCP server, an `orca`/Claude Code session on the same box — can
submit A2A tasks with zero credentials and be treated as fully trusted
(`a2a-hermes.md` §2).

In-band hardening that applies regardless of auth state (`a2a-hermes.md` §2):
prompt-injection defanging (`security.py:180-201`), an explicit
"treat as untrusted peer input" wrapper on every inbound message
(`security.py:206-222`), credential-shaped redaction on replies
(`security.py:230-249`), anti-loop turn cap per `context_id` (default 5, hard
max 20 — `protocol.py:72-83`, `adapter.py:708-722`), per-identity rate limit
(60/min — `protocol.py:495-519`), 1 MB body cap (`adapter.py:65,255-258`), and
an append-only JSONL audit log (`security.py:357-372` →
`~/.hermes/a2a_audit.jsonl`).

`A2A_ALLOWED_USERS` is registered on the platform (`__init__.py:97-138`) but is
**dead config**: `adapter.py:390-403` sets `authorization_is_upstream = True`,
making the gateway's per-platform allow-list a no-op for A2A. The real switches
are `A2A_BEARER_TOKEN` / `A2A_PEER_TOKENS` / `A2A_TRUSTED_PEERS`
(`a2a-hermes.md` §2).

### C2 — the Agent Card overstates capability

`_advertised_skills()` (`adapter.py:618-641`) builds the card's `skills` array
from `tool_registry.get_registered_toolset_names()` — every toolset any plugin
**registered**, not what is **enabled** for the model. Tars enables only
`toolsets: [hermes-cli, kanban]` (`~/.hermes/config.yaml:147-149`,
*re-verified*). The card therefore advertises ~30 skills — `spotify`,
`homeassistant`, `mcp-slack`, `browser`, `code_execution`, `file`, `memory`,
`terminal`, `web`, `a2a`… — that the live agent cannot act on
(`a2a-hermes.md` §2).

Verbatim card, live from Tars
(`curl 127.0.0.1:9900/.well-known/agent-card.json`, `a2a-hermes.md` §2):

```json
{"name": "hermes-tars", "description": "Hermes Agent — a general-purpose agent reachable over A2A.", "url": "http://127.0.0.1:9900/", "version": "1.0.0", "provider": {"organization": "Hermes Agent", "url": "http://127.0.0.1:9900/"}, "supportedInterfaces": [{"url": "http://127.0.0.1:9900/", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"}], "capabilities": {"streaming": true, "pushNotifications": true, "stateTransitionHistory": false, "extendedAgentCard": false}, "defaultInputModes": ["text/plain"], "defaultOutputModes": ["text/plain"], "skills": [ /* 30 toolset skills, see above */ ]}
```

Consequences to carry into any peering decision:

- Any A2A peer reading this card will believe Tars can do far more than it can.
  A peer that routes by capability (`a2a_orchestrate`'s `capability` matching is
  exactly this pattern, `tools.py`) would route work to Tars that Tars silently
  cannot perform.
- `A2A_ADVERTISED_TOOLSETS` / `extra.advertised_toolsets` exists and restricts
  what the **card claims** — it does **not** gate actual tool access
  (`a2a-hermes.md` §4). It is the knob that would close this gap, and it is
  unset.
- The card advertises `http://127.0.0.1:9900/` as its own URL and as the
  `supportedInterfaces` JSONRPC url. Per `a2a-hermes.md` §4, `A2A_PUBLIC_URL` is
  unset and the advertised URL is derived from the `Host` header / bind, so this
  capture reflects the loopback `curl` used to take it. Before relying on any
  non-tunnel transport, verify what a **non-loopback** caller is told — a client
  that follows the card (Hermes' own does, see C3 `_rpc_url`) will POST to
  whatever the card says.

### C3 — the outbound toolset is the sharper edge, and it ships without its allow-list

`a2a` is **not** in `toolsets:` (`config.yaml:147-149`, *re-verified*), so the 5
outbound tools are registered but not exposed to the model
(`a2a-hermes.md` §3). This matches the settled decisions: `DECISION.md` D2 —
"Outbound toolset stays off until something concretely needs Tars to call
another agent. YAGNI; flipping it later is one `hermes tools` call" — and
`PLAN.md` § Amendments D6 deviation 7, "A2A inbound-only".

**This constraint corrects D2's cost estimate.** Flipping the toolset on is one
config change, but the feature has no peer allow-list on the outbound side.
What gets enabled (`tools.py:585-596`, `a2a-hermes.md` §3):

| Tool | Params | Effect |
|---|---|---|
| `a2a_discover` | `url` | fetch + summarize any peer's Agent Card |
| `a2a_call` | `agent`, `message`, `context_id?` | send a `message/send` task to a peer and return the reply |
| `a2a_list` | — | configured peers + persisted conversations + metrics |
| `a2a_history` | `context_id`, `limit?` | replay a persisted transcript |
| `a2a_orchestrate` | `capability`, `message`, `mode?`, `context_id?` | parallel fan-out to every capability-matched peer (max 6 workers) |

The specifics, *re-verified* in source:

- **`_resolve_peer` (`tools.py:53-56`) accepts raw URLs.** Any string starting
  `http://` or `https://` is returned as a peer with empty auth and the default
  timeout, without consulting `a2a_agents` at all. Only non-URL strings are
  looked up in config (and return `None` if unknown). So the model chooses the
  destination host, not the operator.
- **`_rpc_url` (`tools.py:122-130`) follows the URL advertised in the FETCHED
  card**: it prefers the card's `supportedInterfaces` JSONRPC `url`, then the
  card's top-level `url`, and only falls back to the configured base. A peer's
  own card therefore redirects where Hermes POSTs — configuring a peer URL is
  not a binding constraint on the destination.
- **`tools.py` has NO URL validation at all.** `_http_get_json` /
  `_http_post_json` (`tools.py:81-92`) hand the URL straight to
  `urllib.request.urlopen` under `# noqa: S310 (configured peers)` — a lint
  suppression whose stated justification is false the moment `_resolve_peer`
  accepts a raw URL.
- **The SSRF guard is applied at exactly ONE call site.** `grep -n
  is_safe_callback_url *.py` across all five plugin files returns exactly
  `adapter.py:1193` and the definition at `security.py:307` *(re-verified)*.
  That call site is inside `_send_push_notification` (`adapter.py:1182-1197`) —
  the **inbound** push-notification-config path, nothing to do with outbound
  calls.

State plainly: **the peer allow-list is not hardening to defer, it is the
missing half of the feature.** Enabling `a2a` in `toolsets:` without one hands
the model an unrestricted HTTP client running inside the live gateway process
that also holds Gaetan's Slack user token and the `hermes-cli`/`kanban` grants.

---

## Inventory — what exists today

### Tars (192.168.0.9, Hermes v0.20.0, tailnet `100.116.31.76`)

Plugin at `~/.hermes/hermes-agent/plugins/platforms/a2a/` — 5 files, 3219
lines, pure stdlib, no `a2a-sdk` dependency (`protocol.py:17-19`,
`__init__.py:1-9`). Protocol: **A2A v1.0**, JSON-RPC 2.0 over HTTP
(`protocol.py:5,35`), tolerant of v0.3 peer shapes on decode
(`protocol.py:1-19,285-355`). (`a2a-hermes.md` §1)

Routes (`adapter.py:215-241` GET, `243-334` POST):

| Method | Path | Purpose |
|---|---|---|
| GET | `/.well-known/agent-card.json` | canonical v1.0 Agent Card |
| GET | `/.well-known/agent.json` | legacy alias, same payload |
| GET | `/`, `/health` | liveness + (locally) `served_agents` topology |
| GET | `/metrics` | counters |
| POST | `/` | JSON-RPC 2.0 endpoint |

JSON-RPC methods handled (`adapter.py:131-161`, PascalCase v1.0 canonical and
lowercase/slash legacy aliases both accepted): `SendMessage`/`message/send`,
`SendStreamingMessage`/`message/stream` (SSE), `GetTask`/`tasks/get`,
`ListTasks`/`tasks/list`, `CancelTask`/`tasks/cancel`,
`SubscribeToTask`/`tasks/subscribe`, and the 4 push-config CRUD methods.
Live probes confirmed the error surface: `-32601` method not found, `-32001`
task not found (legacy/v1 alias parity), `-32700` parse error
(`a2a-hermes.md` §2).

Streaming: `capabilities.streaming: true`, SSE JSON-RPC-wrapped per spec §9.4
(`adapter.py:998-1028`, `protocol.py:425-448`). Push: `pushNotifications: true`,
inline via `configuration.taskPushNotificationConfig` or the CRUD methods;
payloads HMAC-SHA256 signed when `A2A_PUSH_SECRET` is set, falling back to the
bearer token, **unsigned if neither** (`security.py:256-279`). Per-context
conversations persist across restarts (`protocol.py:791-843` →
`~/.hermes/a2a_conversations/<ctx>.jsonl`).

`/metrics` live: `inbound_total: 0, outbound_total: 0, streams_started: 0,
push_sent: 0, push_failed: 0, tasks_completed: 0` — **zero traffic ever**
(`a2a-hermes.md` §2).

Config keys and current values (`a2a-hermes.md` §4; `config.yaml` has no
`a2a_agents`, `a2a_served_agents`, `platforms.a2a.extra.agents` or
`a2a.trusted_peers` key anywhere in its 203 lines, and `.env` has no `A2A_*`
key — *re-verified* `grep -c '^A2A' ~/.hermes/.env` → `0`):

| Key | Where | Effect | Tars today |
|---|---|---|---|
| `gateway.platforms.a2a.enabled` | config.yaml | inbound server on/off | `true` |
| `A2A_PORT` / `extra.port` | env / config | inbound port | `9900` |
| `A2A_HOST` | env | requested bind host; widened only with a token (C1) | unset → `127.0.0.1` |
| `A2A_AGENT_NAME` | env | card name | unset → `hermes-tars` (`adapter.py:77-85`) |
| `A2A_BEARER_TOKEN` | env | shared inbound bearer secret | unset |
| `A2A_PEER_TOKENS` | env | `"name:token,…"` per-peer credentials | unset |
| `A2A_ALLOWED_USERS` | env | dead config, bypassed by `authorization_is_upstream` (C1) | unset |
| `A2A_ALLOW_ALL_USERS` | env | forces `is_trusted_peer()` open | unset |
| `A2A_TRUSTED_PEERS` / `a2a.trusted_peers` | env / config | identity allow-list once auth exists | unset (moot in localhost-only) |
| `A2A_ADVERTISED_TOOLSETS` / `extra.advertised_toolsets` | env / config | restricts what the **card** claims, not tool access (C2) | unset |
| `A2A_AGENT_DESCRIPTION`, `A2A_PROVIDER_ORG`, `A2A_PROVIDER_URL` | env | card cosmetics | unset |
| `A2A_PUBLIC_URL` | env | overrides the advertised URL (reverse-proxy setups) | unset → derived from `Host`/bind |
| `A2A_REPLY_TIMEOUT` | env | HTTP caller block time | unset → 300 s |
| `A2A_MAX_PINGPONG_TURNS` | env | anti-loop cap per context (hard max 20) | unset → 5 |
| `A2A_RATE_LIMIT` | env | req/min per identity | unset → 60 |
| `A2A_PUSH_SECRET` | env | HMAC key for push payloads | unset |
| `A2A_HOME_CHANNEL` | env | cron/async delivery target for A2A | unset |
| `a2a_agents.<name>` | config.yaml | **outbound** peer registry read by `tools.py` | absent |
| `platforms.a2a.extra.agents` / `a2a_served_agents` | config.yaml | **inbound** multi-agent routing | absent — only the default agent is served |
| top-level `toolsets:` | config.yaml | must include `a2a` to expose outbound tools | `[hermes-cli, kanban]` — **a2a off** |

### p-Hermes (192.168.0.8, user `hermes`, VM 103, Hermes v0.19.0)

**The A2A plugin directory does not exist in this install.** `ls
~/.hermes/hermes-agent/plugins/platforms/a2a/` → *No such file or directory*.
This is **absent from source**, not disabled — v0.19.0 predates the plugin
(`a2a-hermes.md` §5). Corroborated by `DECISION.md` D2: "A2A v1.0 ships in
v0.20.0", which is why Tars was installed at v0.20.0 rather than pinned to
p-Hermes' version.

Also: `config.yaml` has no `a2a` key, `.env` has no `A2A_*` key, nothing listens
on 9900 (`ss -tln` empty; loopback `curl` refused). Version detail: v0.19.0
(2026.7.20), upstream `0a62610f`, local `d9da213a` — **+3 carried local
commits**, install method `git`. `hermes-gateway.service` is otherwise healthy
(`a2a-hermes.md` §5).

Reachability from Tars, measured: `ping` 0.2–6.9 ms, TCP/22 open, TCP/9900
**refused** (nothing listening, not a firewall block). p-Hermes is **NOT on the
tailnet** — `tailscale status` on Tars lists no `.8`/`phermes` peer
(`a2a-hermes.md` §6).

### cooper

Runs **no Hermes A2A endpoint at all** — see the next section.

---

## Why Tars ↔ cooper is not an A2A link

`a2a-claude-code.md` settles this, and the answer is negative on both directions:

- **Claude Code does not speak A2A.** No Anthropic server or client support:
  no `a2a` keyword anywhere in `claude --help` / `claude mcp --help` /
  `claude agents --help` (v2.1.223), no A2A section in the Agent SDK overview
  or headless-mode docs, no hits on `docs.anthropic.com` /
  `code.claude.com` / `platform.claude.com` site search
  (`a2a-claude-code.md` §B). The only bridges are community wrappers
  (`claude-a2a`: 10 stars, README "**WARNING: This project is not production
  ready**"; `a2claude`: single-author, early-stage) — both self-describe as not
  production ready (`a2a-claude-code.md` §B table).
- **Claude Code's own `ListAgents`/`SendMessage` mesh is not a network
  protocol.** It is a local-only, undocumented IPC bus: a supervisor daemon
  keyed by a local secret (`~/.claude/daemon/control.key`) brokering messages
  between sibling `claude` processes over Unix domain sockets
  (`/run/user/<uid>/cc-socks/<pid>.sock`, `/tmp/cc-daemon-<uid>/…/control.sock`),
  addressed by human-assigned session names, versioned by two private integers
  (`proto`, `peerProtocol`) unrelated to A2A's `A2A-Version`. No TCP/HTTP
  surface, no routable binding (`a2a-claude-code.md` §C). A remote host has no
  path in without reverse-engineering the framing **and** obtaining
  `control.key` **and** running on the same machine.

**Conclusion: Tars ↔ cooper is ssh + the `orca` CLI** — the mechanism already
specified in `docs/specs/wf5-orca-delegation.md` and already exercised
(`PLAN.md` § Phases, WF4 probe 15: Tars-initiated `ssh cooper hostname` proving
the model-loop → shell-tool → ssh chain). `a2a-claude-code.md` §D ranks this
option 1 of 5 on effort-to-working and reaches the same verdict independently.
**A2A is the Tars ↔ p-Hermes link only** (and, hypothetically, Tars ↔ some
third-party A2A agent that does not exist yet).

Two discovery-path facts that matter if anything ever does peer:

- The current path is **`/.well-known/agent-card.json`**. The legacy
  `/.well-known/agent.json` is deprecated and **fails silently** against v1.0
  servers; several still-circulating guides use it (`a2a-claude-code.md` §A).
- Hermes' own client is compatible with both: `_fetch_card` (`tools.py:105-111`,
  *re-verified*) requests `agent-card.json` first and falls back to
  `agent.json` on a 404 only. Hermes' server serves both paths
  (`adapter.py:215-241`).

---

## Push-callback reachability — a hard constraint on any push design

The SSRF filter that governs push callbacks is `is_safe_callback_url()`
(`security.py:307-341`) over `_BLOCKED_PREFIXES` (`security.py:292-304`), both
*re-verified verbatim*. The blocked set:

```
"169.254."  link-local / metadata
"127."      loopback
"10."       RFC1918
"172.16." … "172.31."  RFC1918
"192.168."  RFC1918
"0.0.0.0"   unspecified
"::1"       IPv6 loopback
"fe80:"     IPv6 link-local
"fc00:", "fd00:"  IPv6 unique-local
```

Logic, in order: non-`http(s)` scheme → reject; hostname `localhost` → allowed
**only if** `localhost_only()`; prefix match against the block list → rejected,
**except** `127.` and `::1` which are allowed while `localhost_only()`; then a
parsed-IP check rejecting `is_loopback or is_link_local or is_private or
is_reserved`, again with a loopback exemption under `localhost_only()`; anything
that is not an IP literal passes.

Classification verified under the **actual Hermes interpreter**
(`~/.hermes/hermes-agent/venv/bin/python`, 3.11.15):

```
100.116.31.76  private=False  reserved=False  loopback=False  link_local=False
192.168.0.8    private=True   reserved=False  loopback=False  link_local=False
```

Resulting matrix:

| Callback target | Allowed? | Why |
|---|---|---|
| `192.168.x.x` (LAN, incl. p-Hermes `.8` and Tars `.9`) | **never** | `"192.168."` prefix match; the `localhost_only` exemption covers only `127.`/`::1` |
| `127.0.0.1` / `localhost` (e.g. the far end of an ssh tunnel) | **only while no token is set** | the exemption is gated on `localhost_only()`, which is false as soon as `A2A_BEARER_TOKEN` or `A2A_PEER_TOKENS` exists |
| `100.x` tailnet (Tars is `100.116.31.76`) | **allowed** | `"100."` is not in the block list, and CGNAT `100.64.0.0/10` is not RFC1918 — `is_private` is `False` |
| public DNS hostname | allowed | non-IP hostnames pass the final branch untouched |

Consequences, and they are structural:

1. **The LAN link to p-Hermes can never carry an A2A push callback.** Any design
   where p-Hermes pushes task updates to `http://192.168.0.9:9900/…` is dead in
   the code, not in config.
2. **Tunnel-plus-token cannot use push either.** A tunnel puts the callback on
   `127.0.0.1`, and the loopback exemption evaporates the moment you set the
   token that C1 requires to widen anything. So "ssh tunnel + peer token" is a
   **poll-only** design.
3. **Poll-only over a tunnel remains fine** — `tasks/get` / `tasks/list` /
   `message/send` with `A2A_REPLY_TIMEOUT` blocking (300 s default) are all
   unaffected by this filter.
4. **Only the tailnet passes the filter**, and p-Hermes is not on the tailnet
   (`a2a-hermes.md` §6). Enrolling it is a real infra step, not a config change.

One nuance worth keeping: the guard runs at **delivery**, not registration —
`_send_push_notification` (`adapter.py:1182-1197`) pops the stored callback URL
and validates it just before the POST, logging `push notification for task %s
blocked — unsafe callback URL` and incrementing `push_failed`. A peer can
register a blocked callback successfully and only discover it failed via the
absent push. Any push design needs a probe that reads `/metrics`
`push_sent`/`push_failed`, not just a successful `CreateTaskPushNotificationConfig`.

---

## Transport options — PARKED, not chosen here

**This decision is Gaetan's and it is parked.** The table below is comparison
material only; nothing in it is a recommendation and nothing in it should be
enabled by a session reading this document.

| Option | Mechanism | Cost | Exposure | Push-compatible? |
|---|---|---|---|---|
| **ssh-tunnelled localhost** | `ssh -L 9900:127.0.0.1:9900 phermes` from Tars (persistent unit / `autossh`), peer URL `http://127.0.0.1:9900` | one long-lived process to babysit; reuses existing ssh key trust; zero new listening surface on p-Hermes | lowest — the far end never binds beyond loopback; only the ssh-key holder reaches it, only while the tunnel is up | **No if a token is set** (loopback exemption is gated on `localhost_only()`); poll-only |
| **direct LAN bind + peer token** | p-Hermes sets `A2A_HOST=192.168.0.8` **and** a peer token (coupled per C1), Tars calls `http://192.168.0.8:9900` | none beyond the config change | any host on `192.168.0.0/24` can attempt a connection; the peer token is the only gate | **No** — 192.168.x callbacks are blocked outright |
| **tailnet bind + token** | enrol p-Hermes on the tailnet, bind `A2A_HOST` to its `100.x` address + token | requires enrolling a new tailnet node — a real infra step, not config | narrower than LAN (tailnet ACLs), wider than loopback; also the natural shape for an external A2A peer reaching Tars | **Yes** — `100.x` passes the filter |
| **reverse proxy** (nginx/caddy in front of 9900) | proxy terminates TLS, enforces auth/IP allow-list, forwards to loopback | most operational overhead — another service to run and patch; likely needs `A2A_PUBLIC_URL` set so the card advertises the proxy (C2) | best for genuine public exposure; overkill for a two-VM home lab link | depends on the callback address the peer registers, not on the proxy |

Cross-cutting notes for whichever is picked:

- Every non-loopback option requires a token by construction (C1) — the trust
  model is not separable from the transport.
- Every non-tunnel option needs the card's advertised URL verified from a
  non-loopback caller before a peer is pointed at it (C2).
- Push is a *design* choice, not a *transport* choice, but the two are welded
  together by the block list above: **choose push and the tunnel and LAN options
  are eliminated**.
- `a2a-hermes.md` §6 offers the recon author's own read (tunnel = lowest new
  surface for Tars ↔ p-Hermes; tailnet = natural fit for an external peer). That
  is recorded here as input, not as a decision.

---

## p-Hermes proposal — WRITTEN ONLY, nothing executed

> **p-Hermes is READ-ONLY.** No step below has been run, and every one of them
> needs Gaetan's explicit go before it is. `PLAN.md` § Guardrails ("Lane B never
> writes to personal Hermes — read-only probe only") and the repo's gated-cleanup
> rule both apply. This section is a written proposal reproduced from
> `a2a-hermes.md` §5.

1. **Upgrade p-Hermes to a version carrying the A2A plugin** (≥ the release Tars'
   v0.20.0 came from), *or* cherry-pick the plugin directory — p-Hermes carries
   **+3 local commits** over upstream (`d9da213a` vs `0a62610f`), which makes a
   full upstream bump the riskier of the two. **Without this, none of the rest is
   possible: the code is not there.** This step blocks everything else in this
   document that involves p-Hermes.
2. **Enable the inbound server**: add `gateway.platforms.a2a.enabled: true`
   (+ `extra.port: 9900` or another free port) to p-Hermes' `config.yaml`, same
   shape as Tars' § Inventory. Note the config-key path is version-dependent —
   `DECISION.md` D2 flagged that live v0.19.0 has a top-level `platforms:` rather
   than `gateway.platforms:`; determine the path empirically on whatever version
   lands, do not assume Tars' shape transfers.
3. **Decide the trust model before touching `A2A_HOST`** — bind and token are
   coupled (C1), so this is one decision. The recon's minimal-exposure sketch:
   set `A2A_PEER_TOKENS="tars:<random-token>"`, leave `A2A_HOST` unset (stays
   `127.0.0.1`), reach it over an ssh tunnel from Tars. The LAN alternative is
   `A2A_HOST=192.168.0.8` with the same peer token. **Both are parked options,
   not instructions.**
4. **Add `A2A_TRUSTED_PEERS=tars`** so a leaked or guessed second token still
   fails the trust gate. Note this key is inert while `localhost_only()` is true
   (C1) — it only starts doing work once step 3 sets a token.

Handling rules that apply to any of the above if it is ever approved: the
`.bak`-then-`flock` edit discipline from `CLAUDE.md`, per-key `sops -d
--extract` for any token delivery, and never a secret on argv or in an evidence
file.

---

## Guardrail questions to answer before any enablement

These are open questions, not answered here. They are the WF5 scope item
"define guardrails (what a remote agent may ask Tars; what Tars may ask
p-Hermes)" from `PLAN.md` § Phases.

**1. What may a remote agent ask Tars to do?**
Today the honest answer is "anything the model will do with `hermes-cli` and
`kanban`, inside the live gateway, with the operator's memory" (C1). The card
promises far more than that (C2). Sub-questions: does Tars need an
A2A-specific SOUL/system-prompt clause the way the Slack surface has one
(`docs/facts.md` § SOUL — the identity-frame bug class, and the standing
"no code, no PRs, orchestrate/report only" rule)? Is a per-surface toolset
restriction available, or is the toolset grant global to the process? Does
`A2A_ADVERTISED_TOOLSETS` get set so the card stops lying?

**2. What may Tars ask p-Hermes to do?**
p-Hermes is Gaetan's personal agent with its own credentials and its own home
channel. An A2A call from Tars lands in p-Hermes' live gateway with p-Hermes'
full tool grant, by the same C1 mechanism. Sub-questions: is there a read-only
subset? Does an inbound A2A task on p-Hermes reach its Slack surface or its
memory? What does the anti-loop cap (5 turns) mean for a Tars→p-Hermes→Tars
chain, and does either side's `A2A_HOME_CHANNEL` create a delivery path nobody
intended?

**3. Is a peer allow-list mandatory before the outbound toolset is enabled?**
**Per C3, yes.** `_resolve_peer` accepts raw URLs, `_rpc_url` follows the
fetched card, and `tools.py` validates nothing — enabling `a2a` in `toolsets:`
without a peer allow-list is not "the feature with hardening deferred", it is
the feature with half of it missing. The allow-list does not exist as a config
key today (`a2a_agents` is a *registry*, not a *restriction* —
`_resolve_peer` bypasses it entirely for URL-shaped arguments), so satisfying
this would mean a source change, not a config change. That reclassifies
"flipping it later is one `hermes tools` call" (`DECISION.md` D2) as an
underestimate.

---

## Parked decisions

None of these are decided. Listed in dependency order.

| # | Decision | Owner | Blocks |
|---|---|---|---|
| P1 | **p-Hermes upgrade go/no-go** (full bump vs cherry-pick the plugin, given +3 carried commits) | Gaetan | everything Tars ↔ p-Hermes — the plugin is absent from source, so no config path exists |
| P2 | **Transport choice** (tunnel / LAN+token / tailnet+token / reverse proxy) — and, welded to it, whether push is ever wanted | Gaetan | any round-trip probe; the push answer eliminates two of the four options |
| P3 | **Whether to enable the outbound `a2a` toolset, and on what allow-list** — reverses `DECISION.md` D2 / `PLAN.md` D6 dev. 7, and per C3 needs a source-level allow-list first | Gaetan | Tars-initiated calls of any kind |
| P4 | **Whether to set a token and widen the bind on Tars** (one decision, per C1) | Gaetan | any peer reaching Tars from off-box |

Not parked, because it is settled by evidence rather than preference:
Tars ↔ cooper is **not** an A2A link and no decision is pending on it — it is
ssh + `orca` (`docs/specs/wf5-orca-delegation.md`).

---

## Corrections to earlier documents

Recorded here rather than silently carried:

- **p-Hermes host identity.** `192.168.0.8`, user `hermes`, VM 103.
  `192.168.0.3` is `pve`, the **Proxmox hypervisor** — a different machine.
  Source: the Tars VM's `~/.ssh/config` (`a2a-hermes.md` § "Correction to the
  task brief"). `CLAUDE.md` has been corrected; any runbook still saying
  "p-Hermes = .3" is stale.
- **`DECISION.md` D2 says p-Hermes "refuses direct SSH"** and routes via
  `ssh root@192.168.0.3` → `qm guest exec 103` → `runuser -u hermes`. The recon
  contradicts this: `ssh phermes` **from the Tars VM** reaches
  `hermes@192.168.0.8` directly and was used for the whole §5 inventory
  (`a2a-hermes.md` §5). The two statements may both be true from different
  origin hosts (cooper vs Tars); this has not been tested from cooper and is
  **not** resolved here.
- **`PLAN.md` § Phases WF5 row: "p-Hermes A2A state unknown"** — now known and
  answered: the plugin is absent from source at v0.19.0 (§ Inventory).
- **`DECISION.md` D2: "flipping it later is one `hermes tools` call"** — an
  underestimate; see C3 and § Guardrail questions Q3.

---

## Sources

Probe files (this repo):

- `status/probes/wf5/a2a-hermes.md` — §1 version/plugin layout, §2 inbound
  server + auth + card, §3 outbound toolset, §4 config key inventory, §5
  p-Hermes inventory + 4-step proposal, §6 transport options and measured
  reachability.
- `status/probes/wf5/a2a-claude-code.md` — §A A2A protocol state (discovery
  path rename, RPC methods, push semantics), §B no Anthropic A2A support +
  community bridge maturity, §C Claude Code's local IPC mesh, §D ranked
  delegation options.

Repo documents: `docs/facts.md` (toolset-enable-gate precedent, SOUL
identity-frame bug class), `docs/recon/DECISION.md` D2, `PLAN.md`
§ Phases (WF5 row) / § Amendments (D6 dev. 7) / § Guardrails,
`docs/specs/wf5-orca-delegation.md` (the actual Tars ↔ cooper mechanism),
`CLAUDE.md` (host map, secret-handling and config-edit hard rules).

Re-verified live on the Tars VM for this spec, 2026-08-07, all read-only:

```
ss -tlnp | grep 9900                                   # 127.0.0.1:9900 only
grep -c '^A2A' ~/.hermes/.env                          # 0
grep -n -A4 '^toolsets:' ~/.hermes/config.yaml         # 147-149: [hermes-cli, kanban]
hermes --version                                       # v0.20.0 (2026.8.3)
sed/awk over plugins/platforms/a2a/security.py         # resolve_bind_host 108,
                                                       #   _BLOCKED_PREFIXES 292-304,
                                                       #   is_safe_callback_url 307-341
grep -n 'is_safe_callback_url' plugins/platforms/a2a/*.py   # single call site adapter.py:1193
awk over plugins/platforms/a2a/adapter.py 1180-1200    # _send_push_notification guard context
awk over plugins/platforms/a2a/tools.py 53-145         # _resolve_peer, _rpc_url, _fetch_card
venv/bin/python -c 'ipaddress…'                        # 100.x not private; 192.168.x private
```

p-Hermes was **not** touched while writing this spec.
