# P5 — A2A: what it can and cannot be, transport PARKED

**Status:** investigation complete, **nothing proposed for application tonight** ·
**Source:** orca-a2a peer, 2026-08-07 · **Full spec:** `docs/specs/wf5-a2a.md`
(564 lines) · **Evidence:** `status/probes/wf5/a2a-{hermes,claude-code}.md` ·
**Decide:** Gaetan (one go/no-go, the rest follows)

## The two findings that reshape the question

**1. Tars ↔ cooper is not an A2A link and never will be.** Claude Code doesn't
speak A2A at all — no Anthropic client or server; its `ListAgents`/`SendMessage`
mesh is a **local Unix-socket bus** (`/run/user/<uid>/cc-socks/`) with no network
binding, unreachable from the VM. That link stays **ssh + the `orca` CLI** ([P3](P3-orca-v2-skill.md)).

**2. Tars ↔ p-Hermes is blocked by version, not configuration.** p-Hermes runs
Hermes **v0.19.0**, where the A2A plugin is **absent from source** — not disabled,
not present. Nothing listens on 9900 there. **No flag can enable this; it needs a
p-Hermes upgrade.** A 4-step proposal is written; nothing was executed.

So A2A reduces to exactly one question for you: **upgrade p-Hermes, or leave it.**
Your provisional call was no-go. Everything below only matters if that flips.

## Three standing constraints (they open the spec, above any "how to enable")

**C1 — inbound is unauthenticated.** With `A2A_BEARER_TOKEN`/`A2A_PEER_TOKENS`
unset (both are), `security.authenticate()` admits every request as `ip:<addr>`,
and inbound tasks execute **in the live gateway process with the operator's full
tool and memory grant** — not a throwaway clone. The sole gate is the
`127.0.0.1:9900` bind.
*Correcting my own earlier overstatement to you:* `resolve_bind_host()` **refuses
to widen the bind without a token** — set a wider host with no token and it warns
and binds localhost anyway. Bind and token are coupled **in code**, so accidental
unauthenticated exposure is impossible.

**C2 — the agent card oversells Tars.** It advertises ~30 *registered* toolsets
regardless of what's actually enabled for the model (`toolsets: [hermes-cli,
kanban]`). Any peer reading the card sees capabilities Tars won't honour.

**C3 — the outbound toolset has no peer allow-list, and no URL validation.**
`is_safe_callback_url()` has exactly **one** call site (inbound push-config);
`tools.py` has none, and `_rpc_url` follows the URL advertised **in the fetched
card** — a peer redirects where Tars POSTs. Enabling the outbound toolset without
an allow-list hands the model an unrestricted HTTP client (LAN services, cloud
metadata endpoints). **The allow-list is the missing half of the feature, not
hardening to defer.**
*Peer's own correction:* `_resolve_peer` treats a string as a URL only when it
starts `http(s)://` (its earlier "any unrecognised string" was too strong) — the
no-allow-list conclusion is unchanged. And the SSRF guard fires at push
**delivery**, not registration, so a blocked callback registers fine and surfaces
later as `push_failed` in `/metrics`.

## Transport — parked, but the constraint is counter-intuitive

Push-callback reachability, verified under the real Hermes interpreter:

| Target | Push callback |
|---|---|
| `192.168.x` LAN (incl. p-Hermes at .8) | **Blocked outright** (RFC1918) |
| `127.0.0.1` (SSH tunnel) | Allowed **only while no token is set** |
| tailnet `100.x` | **Allowed** — CGNAT is not RFC1918 |

So "SSH tunnel is always safest" is **false where push is concerned**: the LAN
link to p-Hermes can never carry an A2A push callback, and Tars's tailnet address
is the only address the guard permits. **Poll-over-tunnel remains fine** and stays
the conservative shape. Tailnet + tokens is the only push-capable design.

## Flag for you — a documented fact contradicted by measurement

`docs/recon/DECISION.md` D2 says p-Hermes *"refuses direct SSH"* (route via pve +
`qm guest exec 103`). But the recon reached `hermes@192.168.0.8` **directly** with
`ssh phermes` **from Tars** — which is also how tonight's cutover leg works.
Possibly both true from different origins (untested from cooper). Left unresolved
in the spec rather than asserted; worth a one-line correction to D2 when someone
tests it from cooper.

## Decision

- [ ] p-Hermes upgrade v0.19.0 → v0.20.0: **go** (unblocks everything below)
- [ ] **No-go** — A2A stays parked (current provisional call); Tars↔cooper is unaffected
- *Only if go:*
  - Transport: [ ] poll-over-tunnel (conservative) · [ ] tailnet + tokens (required if push wanted)
  - Outbound toolset: [ ] leave disabled · [ ] enable **only together with** a peer allow-list
  - [ ] Fix C2 (card overstating toolsets) before any peer can read it
