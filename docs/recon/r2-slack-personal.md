# R2 — Slack personal-account access method for Tars/Hermes

Recon probe, WF1. Read-only. No secret values captured — shapes and item names only.

## Verdict

The prior-art harvest method (`xoxc-` token + `xoxd-` cookie from a logged-in Slack **web**
tab, per `~/dev/mc-kestra/SETUP.md` §5) **transfers to Tars as-is for credential harvest** — no
alternative method needed there. The gap is one layer down: **Hermes' built-in Slack
integration is bot-OAuth-only** (`xoxb-`/`xapp-` via Bolt SDK + Socket Mode) and cannot consume
personal session tokens at all. That's a different lane in PLAN.md anyway (A5, "Slack Tars app
tokens," held for cutover) — it doesn't compete with A4.

For A4 (personal-account access), the fix is not a different credential type, it's a different
**consumption layer**: Hermes has a native MCP client (`~/.hermes/config.yaml` → `mcp_servers`,
stdio or HTTP transport, auto-reloads on edit). Point it at an MCP Slack server that supports
the xoxc/xoxd ("stealth") auth mode — `korotovsky/slack-mcp-server` is the concrete existing
match, Docker-deployable, env vars `SLACK_MCP_XOXC_TOKEN`/`SLACK_MCP_XOXD_TOKEN` — same token
values mc-kestra's harvest step already produces. No custom HTTP client code needed. This is a
reuse of an existing tool, not new development.

Confidence: high on the harvest method and the Hermes-gateway gap (both directly evidenced
below); medium on rate limits (Slack does not publicly document limits for the internal
web-client API path xoxc/xoxd calls use).

## Facts with evidence

### 1. Harvest method — confirmed, unchanged from prior art

`~/dev/mc-kestra/SETUP.md` lines 110–118 (§5 "Per-source setup" → Slack), verbatim table:

| Var | Where |
|---|---|
| `SLACK_TOKEN` | Application → Local Storage → `localConfig_v2` → `teams.<T…>.token` (starts `xoxc-`). Or: the `token` form field on any XHR to `/api/`. |
| `SLACK_COOKIE` | Application → Cookies → `slack.com` → the **`d`** cookie value (starts `xoxd-`). |

Must be the logged-in Slack **web** tab in devtools — not the desktop app. This is the exact
step lane A runs in Gaetan's browser for A4; no changes needed for Tars.

Independent corroboration (web search, korotovsky/slack-mcp-server auth docs): same two values,
same DOM paths — `JSON.parse(localStorage.localConfig_v2).teams[...].token` for xoxc, the `d`
cookie in Application → Cookies for xoxd. Two independent tools extract the identical pair the
identical way — the method is not mc-kestra-specific.

Gotcha carried over: `xoxd-` cookie must be URL-encoded (`%2F` not `/`) for Slack's HTTP API;
mc-kestra's `env-secrets.sh` auto-encodes a decoded paste. Whatever stores it for Tars
(1Password/SOPS per PLAN.md) needs to either store it URL-decoded and encode at consumption, or
confirm the chosen MCP server does its own encoding — flagged as open question below.

### 2. Lifetime & invalidation triggers

mc-kestra doc (SETUP.md line 146-148): "Practical lifetime is long (~5 years) — it dies on
logout, a password change, or an admin-forced session reset, not on a schedule."

Independent web research (2026): xoxc/xoxd are tied to the browser session; a Dec 2025 Slack
change **shortened the cookie's max lifetime from ~10 years to just over 1 year**, but
"expiration date over a year's time" — still long. Same three invalidation triggers confirmed:
logout, password change, admin-forced session reset. The two sources agree on trigger
mechanics; the absolute duration figure has moved (10y → ~1y as of the Dec 2025 change) — treat
mc-kestra's "~5 years" as stale and plan re-harvest around the ~1-year mark, not 5.

Practical implication for Tars: same operational posture as mc-kestra's — a failure means
"re-harvest," not "something's broken." No cron-based rotation needed.

### 3. Hermes-side consumption — concrete gap found, concrete fix found

**Gap**: Hermes' documented Slack integration (`hermes-agent.nousresearch.com/docs/user-guide/messaging/slack`,
fetched) is bot-token-only:

> Hermes uses bot tokens via Slack App OAuth... "Bot Token (xoxb-) + App-Level Token (xapp-)
> needed" for Socket Mode connectivity... Hermes operates exclusively as a bot, not a personal
> user account. The system uses the Slack Bolt SDK with Socket Mode.

Config vars documented: `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`, `SLACK_ALLOWED_USERS`,
`SLACK_HOME_CHANNEL`. None of these accept `xoxc-`/`xoxd-`. This is PLAN.md's A5 lane (Tars
app tokens), confirmed to be a genuinely separate credential/config path from A4 — good, matches
the plan's own split.

**Fix**: Hermes ships a native MCP client independent of the gateway
(`hermes-agent.nousresearch.com/docs/user-guide/features/mcp`, fetched):

> Config stored in `~/.hermes/config.yaml` under `mcp_servers`. Stdio transport example:
> ```yaml
> mcp_servers:
>   my_slack_server:
>     command: "npx"
>     args: ["-y", "@modelcontextprotocol/server-slack"]
>     env:
>       SLACK_BOT_TOKEN: "xoxb-..."
> ```
> Tools auto-prefixed `mcp_<server_name>_<tool_name>`. Editing config.yaml live-reloads MCP
> connections (30s timeout; `hermes mcp login <server>` for interactive OAuth, 5-min window).

`korotovsky/slack-mcp-server` (github.com/korotovsky/slack-mcp-server, fetched README + auth
docs) explicitly supports this exact credential pair as its "stealth mode":

> You need one of: `xoxp-*` User OAuth token, `xoxb-*` Bot token, or both `xoxc-*` and
> `xoxd-*` session tokens... priority is xoxp > xoxb > xoxc/xoxd.
> Env vars: `SLACK_MCP_XOXC_TOKEN`, `SLACK_MCP_XOXD_TOKEN`.

Deployable via Docker (repo ships `Dockerfile`/`docker-compose.yml` — fits Gaetan's
docker-first tooling pref) or `go run`. So the Hermes-side "HTTP client setup" is: no custom
client, one `mcp_servers` stanza in the Tars profile's `config.yaml` pointing at this server
(or an equivalent), fed the same two values lane A already harvests. This is a config change,
not new code.

**Operational gotcha specific to this live/MCP-server consumption pattern** (does not apply to
mc-kestra's slackdump, which is a one-shot/nightly bulk archiver, not a long-running server):
`korotovsky/slack-mcp-server` issue #86 (github.com/korotovsky/slack-mcp-server/issues/86,
fetched) — "Slack invalidates XOXC/XOXD tokens when the MCP server tries to cache users in the
workspace." Reporter's workaround: this doesn't happen with `xoxp` tokens, and the trigger is
specifically **bulk workspace user enumeration/caching at startup**, not normal message
read/post. Whatever Slack MCP server Tars runs, confirm it does lazy/on-demand user lookups, not
eager full-workspace caching on init — that's the one thing that can burn the harvested token
early. mc-kestra's slackdump never hits this because it isn't a long-running interactive server.

### 4. Rate limits

Slack's officially documented 2025→2026 tightening (`docs.slack.dev/changelog/2025/05/29/rate-limit-changes-for-non-marketplace-apps`,
fetched) — `conversations.history`/`conversations.replies` cut to 1 request/min, 15
objects/request for **non-Marketplace registered Slack Apps** (new installs from 2025-05-29;
ToS applies to pre-existing apps from 2025-06-30). Scope language talks about "apps" and
"applications" throughout — a registered OAuth app with a client_id. xoxc/xoxd personal-session
calls are not an "app" in Slack's model at all; they're the same auth the web client itself
uses. This tightening is evidenced as **not applicable** to the A4 personal-account path with
high confidence (it targets a different auth surface entirely), though Slack does not publish
an explicit rate-limit number for the xoxc/xoxd/web-client path to cite directly — noted as an
open question below.

Corroborating empirical evidence from mc-kestra's own operation (SETUP.md lines 149, 166-168):
slackdump using this exact xoxc/xoxd pair bulk-archived ~290 conversations (member channels +
group-DMs + every DM) and only hit soft rate-limit **pauses**, not hard failures — "you'll see
rate-limit pause messages in the log — that's Slack, not a bug." If bulk archival of the entire
personal workspace only produces soft pauses, Tars' expected usage (read/post a handful of
messages, targeted at Gaetan's own DMs/mentions) is well inside whatever budget governs this
path.

### 5. Alternatives considered — no gap in the harvest method itself, so none needed

Per task instructions, only pursue an alternative harvest method if the prior-art one has a
concrete gap for Tars' use case. It doesn't — the gap found is one layer down (Hermes' gateway
consumption), and it has a fix that reuses the same xoxc/xoxd values. For completeness:

- **Legacy full-workspace tokens**: Slack stopped issuing new legacy tokens years ago (not
  obtainable for a fresh Tars setup) — not viable, not needed.
- **Browser-session export / UI automation** (e.g. a persisted Playwright/Puppeteer session
  driving the Slack web UI instead of its API): heavier, fragile to Slack UI changes, and
  unnecessary — the MCP-server route above gets API-level access with the same two harvested
  values. Not recommended.

## Blockers

None. The harvest step is unblocked (Gaetan's browser, lane A, whenever he's ready). Choosing
and wiring the actual MCP Slack server binary into the Tars Hermes profile is a lane-A/WF3
task, not a recon blocker — flagged as an open question for whoever picks it up next.

## Open questions

- Which MCP Slack server does Tars actually run — `korotovsky/slack-mcp-server` (verified to
  support xoxc/xoxd + Docker) vs. writing a thinner one? Recommend the existing one; no
  evidence found that a custom client is needed.
- Does the chosen MCP server expect the `xoxd-` cookie URL-encoded or raw? mc-kestra's
  `env-secrets.sh` auto-encodes; confirm the Slack-MCP-server's own env parsing before assuming
  the same convention, or just store both known-good forms in 1Password and test once.
  Not independently confirmed in this probe — needs a 30-second check against whichever server
  is chosen, once harvested (can be folded into A4's own probe agent per PLAN.md's "each
  credential lands an async probe agent" step).
- Exact numeric rate limit for the xoxc/xoxd web-client auth path: not published by Slack.
  Treated as "generous enough" on the strength of mc-kestra's bulk-archive experience (§4
  above) and the March-2026 tightening's explicit non-applicability, but not a cited number —
  low risk given Tars' low expected message volume, not zero risk.
- Token/cookie storage location for Tars: PLAN.md says 1Password/SOPS, item *names* only in
  `status/lane-a.md` — this probe did not decide the exact item-naming scheme; that's a lane-A
  execution detail, not a recon fact.
