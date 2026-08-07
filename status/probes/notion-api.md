# Notion public-API integration token (`NOTION_API_TOKEN`) — evidence

Async probe, 2026-08-07. Follows `docs/recon/r8-notion-mcp.md`'s recommendation
(mint an internal integration token for the official `makenotion/notion-mcp-server`
MCP path, distinct from the session-cookie trio already probed in
`status/probes/notion.md`).

`NOTION_API_TOKEN` decrypted from `secrets/tars.sops.yaml` (`sops -d --extract`)
into a shell variable only, under `umask 077`. Auth header passed to `curl` via
a `-K` config file (`chmod 600`), never on argv — `printf` (shell builtin, no
subprocess/argv exposure) wrote the header line. Config file `shred -u`'d
immediately after both calls in the same script run. Zero secret material
below or in the response bodies (Notion never echoes the bearer token).

## Call 1 — `GET /v1/users/me`

`Notion-Version: 2022-06-28` accepted on the first attempt (no fallback needed).

- **HTTP 200 — PASS**
- Bot object:
  - `name`: `Tars`
  - `type`: `bot`
  - `bot.workspace_name`: `Mobile Club`
  - `bot.owner.type`: `workspace`

## Call 2 — `POST /v1/search` (`{"page_size":1}`)

- **HTTP 200 — PASS**
- `results`: `[]` (length 0), `has_more: false`

## Verdict

| Check | Result |
|---|---|
| Token decrypts, non-empty | PASS |
| `GET /v1/users/me` | **PASS** (200, bot=`Tars`, workspace=`Mobile Club`) |
| `POST /v1/search` | **PASS** (200) |
| Notion-Version used | `2022-06-28` (accepted as-is, no need to fall back to a newer stable version) |
| Content reachable | **SHARE-GAP** — token is valid and live, but 0 pages/databases are connected to the integration yet. `results` came back empty not because the search failed, but because nothing has been shared with it. |

**Action needed (not this probe's job to do):** Gaetan connects the integration
to a root page — Notion page `···` menu → **Connect to** → `Tars` — once,
so children inherit the grant (per r8's wiring sketch). Until then, WF3's
live Notion MCP query path will authenticate fine but return nothing.

## Workspace / bot identity

- **Workspace:** Mobile Club
- **Integration/bot name:** Tars
