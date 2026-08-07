# WF3 — wiring specs

Five agents, parallel, after the join (VM up ∧ creds proven). Each ends in its own probe.
Binding upstream: `PLAN.md` §Amendments, `docs/recon/DECISION.md` D1–D6, `r2`, `r6`, `r7`.
Secrets appear here as **SOPS key names only**.

## Placeholders — resolve from `status/lane-b.md` before running anything

| Token | Source |
|---|---|
| `<TARS_IP>` | B1/B2 verdict — LAN address on `vmbr0` (VMID 102) |
| `<TARS_HOST>` | B2 verdict — tailscale name, if the tailnet leg happened (D4: additive) |
| `<TARS_USER>` | B2 verdict — cloud-init user owning `~/.hermes` |
| `<TARS_AGE_PUB>` | B1/B2 verdict — VM age public key, enrolled in `.sops.yaml` |
| `<MACOS_HOST>` | B3 verdict — macOS leg of the mesh |

An agent that cannot resolve its placeholders **stops and reports**; it does not guess an address.

## Conventions — read once, applied by all five agents

**Secret delivery.** Canonical store is `secrets/tars.sops.yaml` (D5), recipients = cooper's age
key + `<TARS_AGE_PUB>`. Never on argv, never echoed, never written unencrypted on cooper:

```sh
# whole env, atomic, 0600, never touches disk in clear on either side
sops -d --output-type dotenv secrets/tars.sops.yaml \
  | ssh <TARS_USER>@<TARS_IP> 'umask 077; cat > ~/.hermes/.env.new && mv ~/.hermes/.env.new ~/.hermes/.env'

# single value, for a consumer that wants its own file
sops -d --extract '["GMAIL_APP_PASSWORD"]' secrets/tars.sops.yaml \
  | ssh <TARS_USER>@<TARS_IP> 'umask 077; cat > ~/.hermes/.secrets/gmail_app_password'
```

`--output-type dotenv` requires a flat map — D5's table already is one. Probes that need a secret
in a `curl` call read it from the env or a `-K` config file (r6 pattern), never `-H "Bearer $X"`.

**Concurrent-edit hazard — mandatory.** All five agents touch `~/.hermes/config.yaml` and/or
`~/.hermes/.env`. Every read-modify-write is wrapped:

```sh
flock ~/.hermes/.wf3.lock -c '<edit command>'
```
plus `cp config.yaml config.yaml.bak-<agent>` before the first edit of the run. Hermes live-reloads
config.yaml, so a half-written file is a live failure, not a deferred one.

**Split.** Non-secrets → `config.yaml`. Secrets → `.env` (0600). No systemd `override.conf` (D2).

**Status writing.** WF3 runs in the orchestrator session; agents return verdicts, the orchestrator
appends to `status/lane-a.md`. Agents write no status file (single-writer contract).

**Standing caveat (R1/R7).** Hermes CLI facts are LLM-summarized vendor docs. Every agent runs
`<cmd> --help` on the VM before scripting a non-interactive invocation of it.

---

## 1 — Gmail via Himalaya

**Preconditions**
- SOPS: `GMAIL_ADDRESS`, `GMAIL_APP_PASSWORD` (A3; second app password, minted for Tars — D5).
- Lane B: B2 done (VM reachable), B4 done (`~/.hermes` exists).
- Not installed anywhere today (r6 blocker) — syntax is unverified until it is on the VM.

**Wire**
1. Install: `sudo snap install himalaya` or the release binary — pick whichever `himalaya --version`
   confirms first; record which.
2. Land the password as its own 0600 file (Himalaya reads a command, not a literal):
   `~/.hermes/.secrets/gmail_app_password` via the `--extract` pattern above.
3. `~/.config/himalaya/config.toml`, account name `tars`:
   ```toml
   [accounts.tars]
   default = true
   email = "<GMAIL_ADDRESS value, injected — not literal in this repo>"
   backend.type = "imap"
   backend.host = "imap.gmail.com"
   backend.port = 993
   backend.login = "<same address>"
   backend.auth.type = "password"
   backend.auth.cmd = "cat ~/.hermes/.secrets/gmail_app_password"
   message.send.backend.type = "smtp"
   message.send.backend.host = "smtp.gmail.com"
   message.send.backend.port = 465
   message.send.backend.login = "<same address>"
   message.send.backend.auth.type = "password"
   message.send.backend.auth.cmd = "cat ~/.hermes/.secrets/gmail_app_password"
   ```
   Confirm key spelling against `himalaya account configure --help` / the shipped example before
   accepting a parse error as a credential failure.
4. No MCP server, no plugin: Tars reaches Gmail through its shell tool. Nothing to register.

**Probe (ends the agent)**
```sh
himalaya envelope list -a tars -f INBOX -s 1
```
Pass evidence: one envelope row (sender/subject/date) on stdout, exit 0. No message fetched, no
send. Do not hardcode folder names — resolve via the IMAP `\All` tag if `INBOX` localizes (r6).

**Fallback probe** (credential-vs-config discriminator, needs no Himalaya):
```sh
curl -s --url imaps://imap.gmail.com --user "$GMAIL_ADDRESS:$GMAIL_APP_PASSWORD"
```
Mailbox list back = the credential is fine and the failure is Himalaya config.

**Rollback**
- Probe fails, fallback passes → `rm ~/.config/himalaya/config.toml`, re-derive from
  `himalaya --help`; the credential is untouched.
- Both fail → the app password is dead or Advanced Protection is on the account. Do **not** debug:
  re-mint in lane A, re-encrypt, redeliver. Delete `~/.hermes/.secrets/gmail_app_password` on the
  way out so a stale value cannot half-work.

---

## 2 — Linear + Notion + Calendar

**GATE (Gaetan, 2026-08-07): before wiring Notion, verify the chosen Notion MCP server can
consume the stored trio (token_v2 / file_token / space_id — internal web-session auth). Most MCP
servers want a public-API integration token instead. Recon r8-notion-mcp.md answers this; if the
verdict is integration-token, lane A mints one (2-min browser step, notion.so → integrations)
and it lands in SOPS as NOTION_API_TOKEN before this agent runs.**


**Preconditions**
- SOPS: `LINEAR_API_KEY`; `NOTION_TOKEN_V2`, `NOTION_FILE_TOKEN`, `NOTION_SPACE_ID`;
  Calendar rides `GMAIL_ADDRESS` + `GMAIL_APP_PASSWORD` (D6-2 — no credential of its own).
- Lane B: B4 done. Coordinate with agent 1: both need the Gmail pair present in `.env`.

**Wire**
1. Deliver the full `.env` once (conventions block). All five keys are consumed as env vars by
   Tars' shell tool — no CLI, no MCP server, no plugin for any of the three.
2. `flock`-guarded: nothing to add to `config.yaml`. If a later agent needs these visible to a
   subprocess, they already are — the gateway unit inherits `~/.hermes/.env`.
3. Record in the WF3 verdict that `NOTION_FILE_TOKEN` is **not covered** by this agent's probe
   (r6: only a real export exercises it) — its first real use is its probe, in WF4 or WF5.

**Probe (ends the agent)** — three calls, secret never on argv:
```sh
# Linear — BARE key, no "Bearer" prefix (r6 gotcha, proven live)
printf 'header = "Authorization: %s"\nheader = "Content-Type: application/json"\n' "$LINEAR_API_KEY" > /tmp/lin.cfg
curl -s -K /tmp/lin.cfg -d '{"query":"{ viewer { id name } }"}' https://api.linear.app/graphql; shred -u /tmp/lin.cfg

# Notion — loadUserContent with empty body, NOT /api/v3/search (its schema drifted; the search
# probe 400s on ANY token. loadUserContent is auth-shaped: 200+recordMap valid, 401 invalid —
# proven live 2026-08-07, status/probes/notion.md probes 4-5)
curl -s -o /dev/null -w '%{http_code}\n' -b "token_v2=$NOTION_TOKEN_V2" \
  -H "Content-Type: application/json" -d '{}' https://www.notion.so/api/v3/loadUserContent

# Calendar (legacy DAV host — the only one that accepts Basic auth)
curl -s -o /dev/null -w '%{http_code}\n' -u "$GMAIL_ADDRESS:$GMAIL_APP_PASSWORD" \
  -X PROPFIND -H "Depth: 0" "https://www.google.com/calendar/dav/$GMAIL_ADDRESS/events"
```
Pass evidence: Linear → 200 with `viewer.id`; Notion → 200 with a results envelope (empty counts);
Calendar → `200`. All three run **from the VM**, not from cooper — the point is that the VM's copy
of the secrets works.

**Rollback**
- Linear 401 → key revoked (they never expire on a schedule); re-mint in lane A.
- Notion 401 → `token_v2` died on a web logout; re-harvest, do not debug.
- Calendar 401 **while Gmail IMAP still passes** → this is the endpoint-retirement risk (r6's
  sharpest gotcha), not a bad password. Report it as a blocked capability; there is no fallback
  code to switch to. Do not re-mint the app password on this signal.
- Any failure: `.env` stays as delivered; nothing to undo on the VM.

---

## 3 — GitHub via `gh` CLI + mc-metarepo clone

**REWRITTEN 2026-08-07 (Gaetan's call): no PAT, no SOPS entry.** `gh auth login` device-flow ran
directly on the VM (lane A staged it; Gaetan authorized in his browser). The credential lives in
the VM's gh store, self-managed, like A1b's OAuth.

**Preconditions**
- `gh auth status` on the VM → logged in, scopes include `repo` (+ org access).
- Lane B: B2 done (`git` present). The completed device-flow is the gate for this agent.

**Wire**
1. `gh auth setup-git` — routes git HTTPS auth through gh's credential helper; no token in any
   URL, `.git/config`, or plaintext helper file.
2. `git clone https://github.com/mobile-club/metarepo.git ~/dev/mc-metarepo`.
3. Do **not** add `credential.helper store` and do not create askpass files — gh IS the helper.

**Probe (ends the agent)**
```sh
gh auth status --hostname github.com
gh api repos/mobile-club/metarepo --jq .private
git -C ~/dev/mc-metarepo rev-parse --short HEAD
grep -c 'ghp_\|github_pat_\|x-access-token' ~/dev/mc-metarepo/.git/config
```
Pass evidence: logged-in status · `true` · a commit SHA · `0`. The last line stays non-optional —
it is the check that the clone did not persist a raw credential.

**Rollback**
- `.private` not `true` / 404 → the gh login lacks org access; re-run the device flow in lane A
  (`gh auth refresh -s repo,read:org` first — cheaper than a full re-login). Nothing to undo.
- Clone fails after the API call passes → network/host-key issue, retry; do not re-auth.
- Any grep hit ≠ 0 → `rm -rf ~/dev/mc-metarepo`, `gh auth logout` + fresh device flow, re-clone.

---

## 4 — Slack personal via `korotovsky/slack-mcp-server` (Docker)

**Preconditions**
- SOPS: `SLACK_TOKEN` (`xoxc-`), `SLACK_COOKIE` (`xoxd-`, **stored URL-decoded** — D1).
- Lane B: B2 done **with Docker** (D6-5), B4 done.
- D1 binding constraint: **lazy user lookups, no eager full-workspace caching at startup**
  (korotovsky #86 — bulk user enumeration is what gets xoxc/xoxd invalidated).

**Wire**
1. `.env` delivery (conventions block), plus a container-local env file:
   `~/tars/slack-mcp/.env` (0600) mapping `SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN` from
   `SLACK_TOKEN` / `SLACK_COOKIE`.
2. **Before starting anything**, read the image's own surface for a users-cache switch:
   `docker run --rm ghcr.io/korotovsky/slack-mcp-server --help` (and its README pinned at the tag
   used). If an eager-cache/user-prefetch option exists, disable it explicitly and **record the
   exact flag or env var name in the verdict** — D1 says verify the deployed config before
   declaring the wiring done, and the flag name is not in recon.
3. Register as a **stdio** MCP server in `config.yaml` — Hermes owns the container lifecycle, no
   port exposed, no long-running process to cache anything between calls:
   ```yaml
   mcp_servers:
     slack:
       command: "docker"
       args: ["run", "-i", "--rm", "--env-file", "/home/<TARS_USER>/tars/slack-mcp/.env",
              "ghcr.io/korotovsky/slack-mcp-server:<TAG>"]
   ```
   `flock`-guarded edit. Tools surface as `mcp_slack_<tool>`. Config.yaml live-reloads (30s).
   Fallback only if the image is SSE/HTTP-only: `docker compose up -d` bound to `127.0.0.1:<port>`
   and an HTTP-transport stanza — same env file, same cache check.
4. Cookie encoding: store decoded, let the server encode. Confirm which the server expects with the
   probe below rather than assuming (r2's explicit 30-second open question).

**Probe (ends the agent)** — three steps, in this order:
```sh
# a. server sees the workspace
hermes mcp list            # expect: slack, connected  (verify subcommand with `hermes mcp --help`)
# b. one real, minimal tool call through Hermes
hermes chat --oneshot 'list my Slack channels'   # verify flag with `hermes chat --help`
# c. the token survived the call — this is the #86 guard, not a formality
curl -s -b "d=$SLACK_COOKIE" -F "token=$SLACK_TOKEN" https://slack.com/api/auth.test
```
Pass evidence: (a) `slack` connected with `mcp_slack_*` tools listed; (b) a non-empty channel list
in the reply; (c) `{"ok":true,...}` **after** (b) — and capture the `user` field, it is DECISION
Blocker 2 (Gaetan's member ID for `SLACK_ALLOWED_USERS` at cutover). Run (c) at least 5 minutes
after (b) as well if the container was left running.

**Rollback**
- (a) fails → remove the `mcp_servers.slack` stanza (`flock`; live-reload drops it),
  `docker rm -f` any container, fix the stanza. No credential impact.
- (c) returns `invalid_auth` → the tokens are burned. Stop the container, remove the stanza, and
  **re-harvest in lane A** (D1: the failure mode is "re-harvest", never "debug"). Before the second
  attempt, prove the cache switch from step 2 is actually in effect — a second burn on the same
  config means the image caches unconditionally and the server choice has to be revisited.
- Cookie-encoding mismatch (401 from the server but `auth.test` on the raw pair passes) → flip the
  stored form once, retest. One flip, not a search.

---

## 5 — Orca / SSH mesh + model backend

**Preconditions**
- SOPS: `TARS_VM_SSH_PRIVATE_KEY` (mesh keypair, generated at provisioning — D5).
- Lane B: B3 verdict (which legs exist, fingerprints), B4 done.
- Lane A: **A1b already run on the VM** — `hermes auth add openai-codex --type oauth`, device-code
  flow completed in Gaetan's browser (R7). This agent does **not** run it and does **not** store
  anything for it; the credential self-manages in `~/.hermes/auth.json`.
- D3: no Orca Run/Task/Gate/Worker. "Orca wiring" here means SSH reachability only.

**Wire — mesh**
1. Public half into `~/.ssh/authorized_keys` on each target that Tars must reach; private half to
   the VM via the `--extract` pattern, `~/.ssh/id_ed25519`, 0600.
2. `~/.ssh/config` on the VM: one `Host` block per leg, `IdentityFile ~/.ssh/id_ed25519`,
   `BatchMode yes`. Legs:
   - **cooper ↔ Tars** — LAN, `192.168.0.4` ↔ `<TARS_IP>`, both directions. Required.
   - **Tars → `<MACOS_HOST>`** — over tailscale if the tailnet leg landed (D4: additive). Optional;
     record as a deviation if absent, not a failure.
   - **Tars → p-Hermes — deliberately NOT wired.** The only path to VM 103 is
     `ssh root@192.168.0.3` → `qm guest exec 103` (r3: the VM refuses direct SSH), i.e. hypervisor
     **root**. Handing an unattended agent root on `pve` to satisfy a read-only probe is not a
     trade this spec makes. p-Hermes access stays on cooper, where the cutover runs anyway.
     **Flag to Gaetan** — this narrows PLAN's B3 wording; confirm before WF4 writes a probe that
     assumes the leg exists.

**Wire — model backend**
3. `flock`-guarded edit of `~/.hermes/config.yaml`. **OVERRIDE, do not add-if-absent** (lane B
   B4 verdict): the installer pre-fills `model.provider: auto`, `model.default:
   anthropic/claude-opus-4.6`, `model.base_url: https://openrouter.ai/api/v1` — an "add the
   block" step no-ops and leaves Tars pointed at OpenRouter with no key. Use `hermes config set`
   (never assume a key is absent; the installer ships a ~92 KB default config — pristine copy at
   `~/.hermes/config.yaml.installer-default`). Set all three:
   ```yaml
   model:
     provider: openai-codex
     default: gpt-5.6-sol
   ```
   `base_url` (`https://chatgpt.com/backend-api/codex`) is written by `hermes auth add` — **confirm
   it REPLACED the openrouter value after the OAuth step; if the installer default survives,
   `config set` it explicitly** (R7 step 3 + lane B B4). Do not copy p-Hermes' `tars` profile
   block: it pins the stale `gpt-5.5`.
4. Never copy p-Hermes' `auth.json`, never mirror it into SOPS (R7 §5 — rotating refresh token; a
   copied token can invalidate the live p-Hermes session).

**Probe (ends the agent)**
```sh
# mesh — one line per configured leg, from the VM
ssh -o BatchMode=yes -o ConnectTimeout=5 cooper true; echo "cooper:$?"
ssh -o BatchMode=yes -o ConnectTimeout=5 <MACOS_HOST> true; echo "macos:$?"   # if the leg exists
# and the reverse, from cooper
ssh -o BatchMode=yes -o ConnectTimeout=5 <TARS_USER>@<TARS_IP> true; echo "tars:$?"

# model backend — auth, then a real call
hermes auth status openai-codex
hermes auth list
hermes chat --oneshot 'Reply with exactly one word: sol'
```
Pass evidence: exit `0` on every configured leg, no prompt, no output. `openai-codex: logged in`;
one credential, `source: device_code`. And a live model reply — this last one is the point: it is
the first actual inference on the new VM and the only thing that closes R7's open question about
whether the subscription tier reaches Sol. `auth status` alone is not a pass.

**Rollback**
- SSH leg fails → fix `authorized_keys` / `~/.ssh/config` on that leg only; the other legs stand.
  Never fall back to password auth to "unblock".
- `auth status` not `logged in` → lane A's A1b did not take. Re-run it **in lane A, in Gaetan's
  browser** — this agent must not attempt an OAuth flow, and must not reach for p-Hermes' token.
- Model call fails while `auth status` says logged in → restore `config.yaml.bak-model` (model id
  or provider wrong), retry once. If it still fails, the ChatGPT tier does not reach Sol: report it
  as a live finding, leave `auth.json` untouched, and let Gaetan pick the fallback model id. Do not
  silently downgrade the config.

---

## What this spec does not cover

- `NOTION_FILE_TOKEN` — no cheap probe exists (r6); first real export is its probe.
- Slack **app** tokens (`SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`) — held for the cutover, never wired
  by WF3 (one token pair, one live gateway — D2).
- `SLACK_HOME_CHANNEL` — set by `/sethome` in the live DM at cutover, `.env` only, no
  `override.conf` (D2).
- SOUL.md / guardrails (`slack.require_mention`, `strict_mention`, `unauthorized_dm_behavior`,
  `SLACK_ALLOWED_USERS`) — profile draft + cutover, targeted by WF4's negative test.
