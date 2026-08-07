# WF3 agent 2 (linear-notion-cal) — evidence

Date 2026-08-07. Target VM `tars` = `gaetan@192.168.0.9` (placeholders resolved from
`status/lane-b.md`: `<TARS_IP>`=192.168.0.9 (B1 PASS), `<TARS_USER>`=gaetan (B2 PASS),
`~/.hermes` base profile = Tars (B4 PASS)). Hermes always invoked as `~/.local/bin/hermes`.

**No secret value appears in this file.** Evidence is exit codes, HTTP codes, booleans, key
*names*, and hash-equality verdicts only.

---

## 0 · Rules honoured

- **Zero whole-file `sops -d`.** Every decrypt was `sops -d --extract '["KEY"]'`, piped straight
  over ssh into a `umask 077` file on the VM. No secret was written unencrypted on cooper, echoed,
  or placed in a shell argument on cooper.
- Key *names* enumerated without decrypting: `grep -oE '^[A-Za-z_][A-Za-z0-9_]*:' secrets/tars.sops.yaml`
  → `NOTION_TOKEN_V2 NOTION_FILE_TOKEN NOTION_SPACE_ID LINEAR_API_KEY GMAIL_ADDRESS
  GMAIL_APP_PASSWORD NOTION_API_TOKEN SLACK_TOKEN SLACK_COOKIE` (+ `sops:` metadata).
- Every write to `~/.hermes/{config.yaml,.env}` was `flock ~/.hermes/.wf3.lock -c '<edit>'`, after
  a one-time backup.
- p-Hermes (192.168.0.3 / VM 103) never contacted. Gateway unit never enabled or started. No
  browser, no OAuth/device flow. `~/.hermes/auth.json` never read or copied.

## 1 · CLI facts verified on the VM before scripting (`--help` first)

| Spec/recon said | Live v0.20.0 says | Consequence |
|---|---|---|
| `hermes chat --oneshot '...'` | **No such flag.** `hermes chat -q QUERY` (single query), `-Q` quiet | probe adjusted to `hermes chat -Q -q` |
| `hermes mcp list` → "connected" | `hermes mcp list` reports **config** state (`✓ enabled`), not liveness. `hermes mcp test <name>` is the live-connection probe | both run; `test` is the real evidence |
| `mcp_servers.<n>.env` passthrough syntax unconfirmed (r8) | `tools/mcp_tool.py` docstring + `_build_safe_env()` + `_interpolate_env_vars()`: `env:` dict is merged into the stdio subprocess env; `${VAR}` resolved from `~/.hermes/.env` loaded at startup | token reaches the server via env, never argv |
| `mcp/notion` image env var | `docker run --rm mcp/notion --help` → `NOTION_TOKEN` (recommended), stdio is the default transport | stanza uses `-e NOTION_TOKEN` + `env:` |

```
$ docker run --rm mcp/notion --help
Usage: notion-mcp-server [options]
  --transport <type>   Transport type: 'stdio' or 'http' (default: stdio)
Environment Variables:
  NOTION_TOKEN         Notion integration token (recommended)
  OPENAPI_MCP_HEADERS  JSON string with Notion API headers (alternative)
$ docker pull mcp/notion   → Status: Downloaded newer image for mcp/notion:latest
```

## 2 · `~/.hermes/.env` — built per-key (I am its sole writer)

Backups first (flock-guarded):

```
flock ~/.hermes/.wf3.lock -c 'cp -p ~/.hermes/config.yaml ~/.hermes/config.yaml.bak-linear-notion-cal
                              && cp -p ~/.hermes/.env      ~/.hermes/.env.bak-linear-notion-cal'
BACKUP_EXIT=0
```

Pre-flight charset classification (counts only, no values printed) — proves unquoted dotenv lines
are safe for both `set -a; . .env` and a python dotenv parser, and that no value carries a
trailing newline:

```
LINEAR_API_KEY      risky_chars=0 trailing_newlines=0
NOTION_TOKEN_V2     risky_chars=0 trailing_newlines=0
NOTION_FILE_TOKEN   risky_chars=0 trailing_newlines=0
NOTION_SPACE_ID     risky_chars=0 trailing_newlines=0
NOTION_API_TOKEN    risky_chars=0 trailing_newlines=0
GMAIL_ADDRESS       risky_chars=0 trailing_newlines=0
GMAIL_APP_PASSWORD  risky_chars=0 trailing_newlines=0
SLACK_TOKEN         risky_chars=0 trailing_newlines=0
SLACK_COOKIE        risky_chars=0 trailing_newlines=0
```
("risky" = anything outside `[A-Za-z0-9_.@%:/+=-]`, measured through `tr -d`/`wc -c`, never printed.)

Delivery, once per key:

```sh
sops -d --extract '["<SOPS_KEY>"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 "umask 077; { printf '%s=' '<ENV_KEY>'; cat; echo; } >> ~/.hermes/.env.new"
# 9/9 exit=0
```

Atomic merge that preserves pre-existing keys (lane B's `HINDSIGHT_MODE`) and cannot clobber a
concurrent writer:

```sh
flock ~/.hermes/.wf3.lock -c 'umask 077;
  awk -F= "NR==FNR{k[\$1];next} !(\$1 in k)" ~/.hermes/.env.new ~/.hermes/.env > ~/.hermes/.env.merged
  && cat ~/.hermes/.env.new >> ~/.hermes/.env.merged
  && chmod 600 ~/.hermes/.env.merged && mv ~/.hermes/.env.merged ~/.hermes/.env
  && rm -f ~/.hermes/.env.new'
MERGE_EXIT=0
```

Final state — `cut -d= -f1 ~/.hermes/.env`, mode `600`, 10 keys:

```
HINDSIGHT_MODE          (lane B B5, preserved untouched)
LINEAR_API_KEY
NOTION_TOKEN_V2
NOTION_FILE_TOKEN
NOTION_SPACE_ID
NOTION_API_TOKEN
GMAIL_ADDRESS
GMAIL_APP_PASSWORD
SLACK_MCP_XOXC_TOKEN    (from SOPS SLACK_TOKEN — renamed per profile spec §3)
SLACK_MCP_XOXD_TOKEN    (from SOPS SLACK_COOKIE — renamed per profile spec §3)
```
Structure check on the staging file: 9 lines, `grep -cvE '^[A-Z_][A-Z0-9_]*=[^[:space:]]+$'` → **0**
malformed lines, mode 600.

**Byte-identity verified** (sha256 of the SOPS value vs sha256 of the value read back out of the
VM's `.env`; hashes compared in-shell, never printed) — this is the guard for `SLACK_COOKIE`, which
is stored **URL-encoded** and must survive `%XX` intact:

```
LINEAR_API_KEY     -> LINEAR_API_KEY       : BYTE-IDENTICAL
NOTION_TOKEN_V2    -> NOTION_TOKEN_V2      : BYTE-IDENTICAL
NOTION_FILE_TOKEN  -> NOTION_FILE_TOKEN    : BYTE-IDENTICAL
NOTION_SPACE_ID    -> NOTION_SPACE_ID      : BYTE-IDENTICAL
NOTION_API_TOKEN   -> NOTION_API_TOKEN     : BYTE-IDENTICAL
GMAIL_ADDRESS      -> GMAIL_ADDRESS        : BYTE-IDENTICAL
GMAIL_APP_PASSWORD -> GMAIL_APP_PASSWORD   : BYTE-IDENTICAL
SLACK_TOKEN        -> SLACK_MCP_XOXC_TOKEN : BYTE-IDENTICAL
SLACK_COOKIE       -> SLACK_MCP_XOXD_TOKEN : BYTE-IDENTICAL
```

**Deliberately excluded** (ownership table, `docs/specs/tars-profile.md` §4): `SLACK_BOT_TOKEN`,
`SLACK_APP_TOKEN`, `SLACK_HOME_CHANNEL`, `SLACK_ALLOWED_USERS` — cutover-held. Also absent by
upstream decision, not omission: `GITHUB_PAT` (replaced by the VM's `gh` device-flow store, A1),
`HINDSIGHT_API_KEY`/`HINDSIGHT_LLM_API_KEY` (skipped for v1, lane A 2026-08-07).

## 3 · `config.yaml` — `mcp_servers.notion` (official makenotion server)

Edit applied through a load→merge→`os.replace` python script under flock, deliberately **not** an
append: a raw `mcp_servers:` append would have produced a duplicate top-level key and silently
dropped agent 4's Slack stanza (last-key-wins in YAML).

```
flock ~/.hermes/.wf3.lock -c 'python3 /tmp/add_notion_mcp.py'   → EDIT_EXIT=0, mode 600
leaf-key diff vs config.yaml.bak-linear-notion-cal:
  added  : ['.mcp_servers.notion.args[]', '.mcp_servers.notion.command',
            '.mcp_servers.notion.env.NOTION_TOKEN',
            '.mcp_servers.slack.args[]', '.mcp_servers.slack.command']
  removed: []
```
(The two `mcp_servers.slack.*` leaves are **agent 4's concurrent write**, which landed between my
backup and my edit — the merge preserved it. Nothing was removed from the config.)

Live block:

```yaml
mcp_servers:
  slack:            # agent 4's, untouched
    command: docker
    args: [run, -i, --rm, --env-file, /home/gaetan/tars/slack-mcp/.env,
           ghcr.io/korotovsky/slack-mcp-server:v1.3.0, --transport, stdio, --no-cache]
  notion:
    command: docker
    args: [run, -i, --rm, -e, NOTION_TOKEN, mcp/notion]
    env:
      NOTION_TOKEN: ${NOTION_API_TOKEN}
```

`-e NOTION_TOKEN` with **no value** forwards the variable from the `docker` process env, which
Hermes populates from the `env:` block after `${NOTION_API_TOKEN}` interpolation from `.env`.
Proven live from the VM's own process table during the probe — the token is nowhere in argv:

```
$ ps -eo pid,etimes,cmd | grep 'docker run -i'
53978  147  /usr/bin/docker run -i --rm -e NOTION_TOKEN mcp/notion
```

The token_v2 trio stays in `.env` for the separate bulk-export path and is **not** MCP-wired (r8).

## 4 · Probes

### 4a · The three specced credential calls — run FROM THE VM, sourcing `~/.hermes/.env`

Script written to the VM over stdin (contains variable *references* only); `set -a; . ~/.hermes/.env;
set +a`; Linear key handed to curl through a `umask 077` `-K` config file that is `shred -u`'d
immediately; no value ever on cooper's argv.

```
--- P1 Linear (bare key, -K config file, no argv)
http_code=200
viewer_id_present=true
viewer_id_len=36            (UUID; value withheld)
viewer_name=Gaëtan Cathelain
errors=none
--- P2 Notion loadUserContent (token_v2, export-path credential)
http_code=200
recordMap_present=true
--- P3 Calendar CalDAV PROPFIND (rides Gmail app password)
http_code=207
inner_200_count=1
body_bytes=945
PROBE_EXIT=0
```

| Probe | Pass line in the spec | Got | Verdict |
|---|---|---|---|
| Linear | 200 with `viewer.id` | 200, `viewer.id` present (36 chars), no `errors` | **PASS** |
| Notion `loadUserContent` | bare `200` (401 = dead token) | 200, `recordMap` present | **PASS** |
| Calendar CalDAV | `207` Multi-Status, inner `HTTP/1.1 200 OK`, never bare 200 | 207 + exactly one inner `HTTP/1.1 200 OK` | **PASS** |

`/api/v3/search` was **not** used (schema-drifted, 400s on any token — spec + `status/probes/notion.md`).

### 4b · MCP-level probe (the r8 path)

```
$ ~/.local/bin/hermes mcp list
  Name     Transport        Tools   Status
  slack    docker run -i    all     ✓ enabled
  notion   docker run -i    all     ✓ enabled

$ ~/.local/bin/hermes mcp test notion
  Testing 'notion'...
  Transport: stdio → docker
  Auth: none
  ✓ Connected (595ms)
  ✓ Tools discovered: 24
    API-get-user, API-get-users, API-get-self, API-post-search,
    API-get-block-children, API-retrieve-a-page, API-query-data-source, … (24 total)
  TEST_EXIT=0
```

Real tool call through Hermes (`hermes chat -Q -q`, model backend = agent 5's openai-codex /
gpt-5.6-sol):

```
$ hermes chat -Q -q "Use the notion MCP tools (API-post-search) to search my Notion workspace
                     for the word Cleaq. Reply with ONLY the title of the top result."
  ⏭ Secret entry skipped            (see surprise 2)
  session_id: 20260807_175205_a6e866
  Cleaq
  CHAT_EXIT=0
```

That first reply is **weak evidence** — "Cleaq" appears in the prompt, so a model could echo it
without touching the tool. Two follow-ups were run to settle it:

```
$ hermes chat -Q -q "Call API-post-search with an empty query, list every object my integration
                     can see, titles + count."
  ⏭ Secret entry skipped   (×2)
  session_id: 20260807_175250_98d3d4
  [subagent-0] ⚡ Interrupted during API call.
  Codex response remained incomplete after 3 continuation attempts
  CHAT2_EXIT=0        → INCONCLUSIVE (backend continuation failure, not an MCP failure)

$ hermes chat -q "Call exactly one tool: the notion MCP tool API-get-self. Then reply with only
                  the bot user name and the workspace name it returns."
  ┊ ⚡ mcp__noti   0.3s                     ← the MCP tool call itself, in the tool preview
  ╭─ Hermes ─────────╮
   Tars
   Mobile Club
  ╰──────────────────╯
  Duration: 13s · Messages: 4 (1 user, 2 tool calls)
  CHAT3_EXIT=0        → PASS
```

`Mobile Club` was **not** in the prompt and matches lane A's recorded workspace for the
`NOTION_API_TOKEN` bot (`status/lane-a.md` A2: bot `Tars`, ws `Mobile Club`). Together with the
visible `mcp__noti…` tool-call preview and the `2 tool calls` counter, this is a real round trip
through the MCP server to the live Notion API — not a hallucinated answer.

## 5 · Deviations

1. **`hermes chat --oneshot` does not exist** in v0.20.0 — used the verified `-Q -q` form. Same
   class of stale-CLI-fact as lane B's 5-of-10 wrong flag spellings.
2. **`hermes mcp list` is not a liveness probe** — it prints configured/enabled state. The
   connection evidence is `hermes mcp test notion` (`✓ Connected`, 24 tools), which is what the
   spec's "connected" intent actually requires.
3. **`config.yaml` merged, not appended.** Deliberate: agent 4 wrote `mcp_servers.slack`
   concurrently and a plain append would have created a duplicate top-level `mcp_servers:` key,
   silently dropping one of the two servers. Same footgun family as lane B's B4 top-level
   `platforms:` finding.
4. **`GMAIL_ADDRESS` / `GMAIL_APP_PASSWORD` delivered as currently stored in SOPS**, even though
   agent 1 (`status/probes/wf3-s1-gmail.md`) declared them transcript-exposed and asked for a
   re-mint that has **not** been recorded in `status/lane-a.md`. The Calendar probe needs the pair
   and it works (207). If lane A rotates it, `.env` must be re-delivered — one `--extract` per key,
   the merge step is idempotent.
5. `NOTION_API_TOKEN` was delivered under its SOPS name; Hermes' `${NOTION_API_TOKEN}` reference in
   `config.yaml` resolves it. No second copy under any other name (see surprise 2).

## 6 · Surprises worth carrying forward

1. **`_build_safe_env()` filters the subprocess environment.** `tools/mcp_tool.py` passes only a
   PATH/HOME-style allowlist, `XDG_*`, external-secret-source values, and the server's own `env:`
   block to a stdio MCP server. So `.env` keys are **not** automatically visible to MCP servers —
   anything an MCP server needs must be named explicitly in its `env:` block. Relevant to WF4/WF5.
2. **Hermes prompts for `NOTION_API_KEY` (not `_TOKEN`) on a Notion-flavoured turn.** It is the
   bundled `notion` **skill**'s credential (`hermes_cli/config_defaults.py`: category `skill`,
   "used by the `notion` skill"), a different consumption path from the MCP server. Headless it
   auto-skips (`⏭ Secret entry skipped`, exit 0) and the MCP call succeeds regardless. Left unset:
   it is not in the profile spec's `.env` map and duplicating a secret under a second name is not
   this agent's call. **Cutover follow-up:** decide whether the gateway should carry
   `NOTION_API_KEY` too, or whether the bundled skill should be disabled, so the prompt never fires
   in the live gateway.
3. `mcp/notion` was **not** present on the VM and had to be pulled (`docker pull mcp/notion` → ok).
   No image tag pin is in the config — it runs `:latest`. Worth pinning before cutover, same
   discipline agent 4 applied (`slack-mcp-server:v1.3.0`).
4. Non-interactive `hermes` prints `rtk: hermes plugin warning: rtk binary not found in PATH` —
   cosmetic here, but it means B5's `rtk-rewrite` plugin is a no-op under a non-login PATH. Not
   this agent's to fix; flagged for whoever owns the gateway unit's environment.
5. **A multi-step chat turn died with `Codex response remained incomplete after 3 continuation
   attempts`** (it had spawned `[subagent-0]`), while a single-tool turn on the same backend
   completed in 13s. Both `mcp test` and the single-tool call pass, so this is not an MCP or
   credential fault — it is a first live sign of the openai-codex/gpt-5.6-sol backend (agent 5's
   PASS_WITH_DEVIATIONS) truncating long delegated turns. Not debugged here per instruction; worth
   one WF4 probe before the gateway carries real work.

## 7 · Not covered, by design

- **`NOTION_FILE_TOKEN` is deliberately unprobed.** No cheap probe exists (r6); the first real
  bulk export is its probe, in WF4/WF5. It is delivered to `.env` and byte-verified, nothing more.
- Slack app tokens, `SLACK_HOME_CHANNEL`, `SLACK_ALLOWED_USERS` — cutover-owned, excluded from
  `.env` on purpose.

## 8 · Rollback state

Nothing to undo. Both probes passed, so none of §2's rollback branches (Linear 401 → re-mint,
Notion 401 → re-harvest, Calendar 401-while-IMAP-passes → blocked capability) fired. Restore points
if needed: `~/.hermes/config.yaml.bak-linear-notion-cal`, `~/.hermes/.env.bak-linear-notion-cal`.

## VERDICT

**PASS_WITH_DEVIATIONS.** All four required probes pass on the section's own pass line: Linear
`200` + `viewer.id`; Notion `loadUserContent` bare `200` + `recordMap`; CalDAV `207` with one inner
`HTTP/1.1 200 OK`; and the MCP layer — `hermes mcp test notion` → `✓ Connected`, 24 tools, plus a
real `mcp__notion` tool call returning `Tars` / `Mobile Club`. `~/.hermes/.env` is built per-key,
0600, 10 keys, all nine SOPS values hash-verified byte-identical (including the URL-encoded
`SLACK_COOKIE`), with lane B's `HINDSIGHT_MODE` preserved and the four cutover-held keys excluded.
`NOTION_FILE_TOKEN` is delivered but **deliberately unprobed** — the first real export is its probe
(WF4/WF5). Deviations are CLI-fact corrections (`-Q -q` not `--oneshot`; `mcp test` not `mcp list`
for liveness), a merge-not-append config edit that saved agent 4's concurrent Slack stanza, and the
still-unrotated Gmail pair inherited from agent 1's incident.

