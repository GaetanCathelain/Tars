# R8 — Notion MCP: can Hermes consume the stored trio, or does MCP want a different credential?

Recon for the WF3 §2 GATE (`docs/specs/wf3-wiring.md` line 114, Gaetan 2026-08-07). Read-only —
WebFetch/WebSearch + GitHub reads, no browser session opened, no token minted. No secret values
below. Binding upstream: `DECISION.md` D1 (MCP-server-as-consumption-layer pattern, proven for
Slack) and D5 (trio = `NOTION_TOKEN_V2`, `NOTION_FILE_TOKEN`, `NOTION_SPACE_ID`, session-cookie
auth); `status/probes/notion.md` (the trio is live — `token_v2` proven valid against
`loadUserContent`, 200 vs 401 control; `/api/v3/search` independently proven to have
schema-drifted — 400 on both real and bogus token, i.e. auth-independent drift).

## Verdict

**No — the stored trio cannot be consumed by a Notion MCP server usable by Hermes. Mint an
internal integration token; it is the WF3 path.** The official `makenotion/notion-mcp-server`
(actively maintained, Docker image, stdio + streamable-HTTP transport — fits Hermes v0.20.0's
native MCP client on either transport) hard-requires a public-API integration token
(`NOTION_TOKEN`, `ntn_...` prefix) or an OAuth bearer; it does not accept `token_v2` in any form.
Exactly one community server accepts `token_v2` directly — `kirvigen/notion-private-api-mcp` — and
it is unfit for this build: 4 stars, 6 commits, 0 issues (near-zero real usage, not "no bugs
found"), stdio-only, and its own README calls it "inherently fragile... not for production." It
talks to the same undocumented `/api/v3/*` surface our own probe just caught mid-drift
(`/api/v3/search` returns HTTP 400 for both a real and a bogus token — schema-shaped, not
auth-shaped) — adopting it imports that exact fragility into Tars' live MCP path with zero
evidence anyone else is maintaining it against Notion's churn. Minting a token is a ~2–5 minute
browser step, the same shape as the browser steps lane A already ran for GitHub/Slack/Gmail (D5's
"harvest" column) — not a new category of cost.

## Auth matrix: server × credential × capability

| Server | Auth accepted | Maintained? | Transport | Capability granted |
|---|---|---|---|---|
| **`makenotion/notion-mcp-server`** (official) | `NOTION_TOKEN` (internal integration secret, `ntn_...`) via env, or `OPENAPI_MCP_HEADERS` (`Authorization: Bearer ...` + `Notion-Version`); also supports a full OAuth app flow for multi-tenant/public integrations — irrelevant here, internal token is the fit | Yes — Notion's own repo, versioned `Notion-Version` header (`2025-09-03`, some tools `2026-03-11`), Docker image `mcp/notion` published | stdio (default) **and** streamable HTTP (`--transport http --auth-token ...`) | Only pages/databases **explicitly connected** to the integration (per-page "Connect to", or connect one root page and children inherit) |
| **`kirvigen/notion-private-api-mcp`** (community) | `token_v2` only (cookie) — no `space_id`/`file_token` needed | No — 4 stars, 6 commits, 0 open issues (signal of no users, not of stability), README self-describes as fragile/not-for-production | stdio only | Everything the logged-in session can see (same blast radius as the trio itself) |
| *(no other token_v2/internal-API MCP server found with meaningful adoption)* | — | — | — | — |
| **Tars' stored trio directly, no MCP** (current: mc-kestra-style dump/export, or a raw-curl probe as in `status/probes/notion.md`) | `token_v2` + `space_id` (+ `file_token` for exports) | N/A — not a server, a credential | N/A — direct HTTP, no MCP client involved | Everything the session can see; correct for a **bulk weekly export**, wrong shape for live "answer one question against Notion" MCP calls |

## Why the trio doesn't map onto any Hermes-usable MCP server

1. **The official server's auth model is public-API-only by design.** It wraps
   `api.notion.com/v1/*`, which has no concept of a session cookie — `NOTION_TOKEN`/OAuth are the
   only inputs its OpenAPI spec accepts. There's no flag or env var to hand it `token_v2`.
2. **The only server that speaks `token_v2` is not a "different setup," it's a different (and
   thin) project.** It doesn't even take `space_id`/`file_token` — it's a strict subset of what
   Tars already holds, built for `/api/v3/*`, the same surface `status/probes/notion.md` just
   proved is drifting under Notion's own hand (search schema changed; `loadUserContent` still
   works today but is unversioned and could move without notice, unlike the official API's
   `Notion-Version` contract). Wiring it would mean re-doing this exact recon's failure mode —
   probe passes today, breaks silently on Notion's next internal schema change, with one
   maintainer (a 4-star repo) as the only upstream fix path.
3. **Capability direction is inverted from what D5 optimized for.** D5 chose the trio because it
   is zero-mint, reuse-what-mc-kestra-already-proved. That optimization was correct for the
   **dump/export** use case (mc-kestra's `gbrain.notion/dump`: weekly full-workspace zip, session
   auth is the only way to get "everything," and there is no MCP server in this path at all — it's
   a direct API call from a Kestra flow). It does not carry over to **live MCP query access**,
   which is a different consumption layer (same distinction D1 already drew for Slack: personal
   session ≠ what an MCP server's own auth contract wants).

## Integration token: cost and capability

**Mint:** `notion.so` → workspace settings → **Connections** → **Develop or manage integrations**
(or directly `notion.so/my-integrations`) → **New integration** → type **Internal** → name it
("Tars") → default capabilities (Read content; add Update/Insert only if WF5 ever needs Tars to
write, which the pitch doesn't ask for — read+query only) → **Submit** → copy the secret
(`ntn_...`, shown once). Browser, no 2FA gate beyond the normal Notion login already logged in.
**~2–5 minutes**, same shape as the GitHub PAT / Linear key harvests already in D5's table — not a
new category of lane-A step.

**One extra step integration tokens need that the trio never did: sharing.** An internal
integration sees **nothing** until pages/databases are explicitly connected to it — either
per-page ("..." menu → **Connect to** → pick the integration) or once on a root page, whose
children inherit the grant. For "read+query Notion content as Gaetan's assistant" (PITCH.md line
18, no scoping given), the practical move is connecting Tars to whatever root page(s) constitute
"the workspace as Gaetan uses it" — a handful of clicks, done once, same session as the token mint.
This is Business/Enterprise-plan-aware: on those plans, integration creation itself needs
workspace-owner approval via Settings → Connections — a one-time approval, not a recurring gate.

**Capability delta vs. the trio:** integration token = explicit allow-list (only shared
pages/databases); `token_v2` = everything the session can see, no allow-list possible. Narrower is
the correct default for an unattended agent holding a live credential — consistent with the
security pref's posture even though this recon doesn't relitigate guardrail scoping (workflow pref:
broad-first-restrict-later applies to *guardrails*, not to *which auth mechanism exists at all*;
here the narrower mechanism is simply the only one on offer from a maintained MCP server).

## Recommendation for WF3 §2

**Both, for different jobs — mint the integration token for MCP/live-query; keep the trio for the
export/dump path it was already scoped for.**

- **Live MCP query (Hermes' own tool calls, "search/read Notion" mid-conversation):** mint
  `NOTION_API_TOKEN` (integration secret), wire `makenotion/notion-mcp-server` as a Docker-run
  stdio MCP server — same registration pattern WF3 agent 4 already uses for
  `korotovsky/slack-mcp-server` (D1's proven shape: Hermes owns the container lifecycle via
  `config.yaml`, no port exposed, no daemon caching state between calls).
- **Trio (`NOTION_TOKEN_V2`, `NOTION_FILE_TOKEN`, `NOTION_SPACE_ID`):** keep exactly as D5/WF3 §2
  already scoped it — delivered to `.env`, probed by direct curl (already done,
  `status/probes/notion.md`), reserved for a future mc-kestra-style bulk export if/when that lands
  on Tars (PITCH.md: "postponed until I finish the work on mc-kestra"). Not deleted, not
  MCP-wired — there is no MCP server it should feed.

This resolves the WF3 §2 gate: the agent does not block waiting for a token_v2-compatible MCP
server that doesn't meaningfully exist. It waits on one ~2–5 minute lane-A browser step
(`NOTION_API_TOKEN` mint + share) before it can wire live Notion MCP access; the trio's rows in the
`.env` delivery and probe steps run unchanged.

## Wiring sketch (recommended path)

**Lane A — one more credential, D5-shaped:**

| # | Credential | Harvest | Probe | SOPS key |
|---|---|---|---|---|
| A2 (addition) | Notion internal integration token | `notion.so/my-integrations` → New integration → Internal → capabilities: Read content → Submit; then connect it to the relevant root page(s) | `curl -s -H "Authorization: Bearer $NOTION_API_TOKEN" -H "Notion-Version: 2025-09-03" https://api.notion.com/v1/users/me` → 200 with the integration's bot user | `NOTION_API_TOKEN` |

**WF3 agent 2 — extend, don't replace, the existing wire:**

1. `.env` delivery unchanged (conventions block) — add `NOTION_API_TOKEN` to the same
   `secrets/tars.sops.yaml` flat map, same `sops -d --output-type dotenv | ssh ... cat > .env`
   pattern already used for the trio.
2. `flock`-guarded `config.yaml` edit, new `mcp_servers.notion` stanza — mirrors WF3 agent 4's
   Slack stanza shape:
   ```yaml
   mcp_servers:
     notion:
       command: "docker"
       args: ["run", "-i", "--rm", "-e", "NOTION_TOKEN",
              "mcp/notion"]
       env:
         NOTION_TOKEN: "${NOTION_API_TOKEN}"
   ```
   (Exact env-passthrough syntax to confirm against `hermes mcp --help` / the live `config.yaml`
   schema at wiring time — same standing caveat R1/R7 already carry for every Hermes flag.)
3. Probe, same three-step shape as the Slack agent's probe:
   ```sh
   hermes mcp list                                   # expect: notion, connected
   hermes chat --oneshot 'search my notion for <a page you know exists>'
   curl -s -H "Authorization: Bearer $NOTION_API_TOKEN" -H "Notion-Version: 2025-09-03" \
     -X POST -d '{"query":""}' https://api.notion.com/v1/search   # 200, results envelope
   ```
4. Trio's existing three-call probe in `docs/specs/wf3-wiring.md` §2 (Linear/Notion/Calendar) runs
   unchanged — it validates the export-path credential, not the MCP path; both can and should pass
   independently.

## Confidence and what stays unverified

- Official server's exact Docker/env-passthrough syntax is WebFetch-summarized vendor docs, not
  locally run — same class of risk R1 already flagged for Hermes itself; verify with
  `docker run --rm mcp/notion --help` and `hermes mcp --help` on the VM before scripting, per the
  wf3-wiring.md standing caveat.
- `kirvigen/notion-private-api-mcp`'s star/commit/issue counts are current as of this recon
  (2026-08-07) and could shift; the verdict rests on the *pattern* (single-maintainer,
  near-zero-adoption wrapper around an undocumented, actively-drifting API — the exact failure
  class `status/probes/notion.md` already caught once), not on the specific numbers holding
  forever.
- Business/Enterprise-plan approval gate for internal integrations is documented generally, not
  confirmed against Gaetan's specific Notion plan/workspace — if it applies, it is a one-time
  workspace-owner click, not a new blocker class.
