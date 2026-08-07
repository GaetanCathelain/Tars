# Cutover — steps 2–4 of `docs/specs/cutover-notes.md` §1

**Date**: 2026-08-07 · **Operator**: cutover execution agent · **Target**: `gaetan@192.168.0.9` (host `tars`)
**Verdict**: **PASS**

Scope executed: token pair + `SLACK_ALLOWED_USERS` into `~/.hermes/.env` · Tars→p-Hermes SSH leg ·
`hermes-gateway.service` start + verify. **Not** in scope and **not** done: `/sethome` (Gaetan's, in
the live DM), any Slack message, any touch of p-Hermes (192.168.0.8) or pve (192.168.0.3), any git
operation, any `status/lane-*.md` edit.

**No secret value appears in this file.** Secrets were moved only by per-key
`sops -d --extract '["KEY"]'` piped straight over `ssh` — never a whole-file decrypt, never echoed,
never on argv. Verification used key **names**, byte **lengths** and documented **prefixes** only.

---

## Pre-flight — spec reads

- `docs/specs/cutover-notes.md` — runbook order, `SLACK_ALLOWED_USERS=U08BDJAMSRZ` (A4, team `T7V1UGJ82`),
  CloakBrowser stale-lock trap, config-shadowing check.
- `docs/specs/tars-profile.md` §3 — env ownership.

**Guardrail-env question (asked explicitly): no guardrail env is required at cutover.**
`tars-profile.md` §2 "Deliberately absent" table lists `SLACK_REQUIRE_MENTION` / `SLACK_STRICT_MENTION`
as *intentionally not in `.env`* — `config.yaml` is the single source. Verified live on the VM
(see §4 config sanity): all three guardrails are present under `gateway.platforms.slack`. Cutover-owned
`.env` keys per §3 are exactly `SLACK_ALLOWED_USERS` (written here) and `SLACK_HOME_CHANNEL`
(`/sethome`, Gaetan's step, correctly still absent). **Nothing listed-but-absent.**

---

## 0 · Baseline (before any change)

```
$ ssh gaetan@192.168.0.9 '...'
== whoami: gaetan@tars
== .env mode:      600 gaetan:gaetan /home/gaetan/.hermes/.env
== key count:      10
== key names:      GMAIL_ADDRESS GMAIL_APP_PASSWORD HINDSIGHT_MODE LINEAR_API_KEY
                   NOTION_API_TOKEN NOTION_FILE_TOKEN NOTION_SPACE_ID NOTION_TOKEN_V2
                   SLACK_MCP_XOXC_TOKEN SLACK_MCP_XOXD_TOKEN
== hermes-gateway.service:      disabled / inactive
== hermes-cloakbrowser.service: enabled  / active
== ~/.ssh/config:      -rw------- 449 bytes (hosts: cooper, pve, mac — no phermes)
== ~/.ssh/known_hosts: -rw------- 3849 bytes, 0 entries for 192.168.0.8
== ~/.ssh/id_ed25519:  -rw------- present
```

Old gateway on p-Hermes: confirmed disabled+inactive by the orchestrator before this run. Not
re-verified here (touching VM 103 is out of scope).

---

## 1 · `.env` merge — **PASS**

Method: backup → build `.env.new` (0600) from scratch → append → verify shape → atomic `mv`.
All four ssh legs run under `flock ~/.hermes/.wf3.lock`. Idempotent: `.env.new` is rebuilt from
`grep -vE '^(SLACK_BOT_TOKEN|SLACK_APP_TOKEN|SLACK_ALLOWED_USERS)='` each run, so a re-run replaces
rather than duplicates. `.env` itself is untouched until the final `mv`.

### 1a — backup + stage (non-secret key)

```
$ ssh gaetan@192.168.0.9 'umask 077; flock ~/.hermes/.wf3.lock -c "
    cp -p ~/.hermes/.env ~/.hermes/.env.bak-cutover && chmod 600 ~/.hermes/.env.bak-cutover
    rm -f ~/.hermes/.env.new
    grep -vE \"^(SLACK_BOT_TOKEN|SLACK_APP_TOKEN|SLACK_ALLOWED_USERS)=\" ~/.hermes/.env > ~/.hermes/.env.new
    chmod 600 ~/.hermes/.env.new
    printf \"SLACK_ALLOWED_USERS=U08BDJAMSRZ\n\" >> ~/.hermes/.env.new
  "'
backup: 600 keys=10
staged: 600 keys=11
```

### 1b/1c — the two secrets, per-key extract piped straight over ssh

```
$ sops -d --extract '["SLACK_BOT_TOKEN"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 'umask 077; flock ~/.hermes/.wf3.lock -c
      "{ printf %s SLACK_BOT_TOKEN= ; tr -d \"\n\" ; printf \"\n\" ; } >> ~/.hermes/.env.new"'
bot_write_exit=0

$ sops -d --extract '["SLACK_APP_TOKEN"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 'umask 077; flock ~/.hermes/.wf3.lock -c
      "{ printf %s SLACK_APP_TOKEN= ; tr -d \"\n\" ; printf \"\n\" ; } >> ~/.hermes/.env.new"'
app_write_exit=0
```

`tr -d "\n"` strips any trailing newline from the decrypted value so the mapping is byte-exact
(same pattern as `wf3-s4-slack.md` Wire 1).

### 1d — shape verification (names / prefix / length only, never values)

```
$ ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "... cut -d= -f1 ...; awk ..."'
== .env.new key NAMES:
GMAIL_ADDRESS GMAIL_APP_PASSWORD HINDSIGHT_MODE LINEAR_API_KEY NOTION_API_TOKEN
NOTION_FILE_TOKEN NOTION_SPACE_ID NOTION_TOKEN_V2 SLACK_ALLOWED_USERS
SLACK_APP_TOKEN SLACK_BOT_TOKEN SLACK_MCP_XOXC_TOKEN SLACK_MCP_XOXD_TOKEN
== key count: 13
== shape (prefix+length only):
SLACK_BOT_TOKEN prefix=xoxb- len=57      <- expected xoxb-/57  MATCH
SLACK_APP_TOKEN prefix=xapp- len=98      <- expected xapp-/98  MATCH
SLACK_ALLOWED_USERS=U08BDJAMSRZ          <- not a secret
== trailing byte is newline: \n
```

### 1e — atomic swap

```
$ ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "mv ~/.hermes/.env.new ~/.hermes/.env; ..."'
post-mv: 600 gaetan:gaetan keys=13
leftover .env.new: ls: cannot access '/home/gaetan/.hermes/.env.new': No such file or directory
```

| Check | Expected | Actual |
|---|---|---|
| mode | 600 | **600** (`gaetan:gaetan`) |
| key count | 10 + 3 = 13 | **13** |
| all 10 pre-existing keys survive | yes | **yes** (name-by-name diff above) |
| `xoxb-` / `xapp-` prefix + 57/98 bytes | yes | **yes** |
| no trailing-newline corruption | yes | **yes** |
| backup exists, 0600, 10 keys | yes | **yes** — `~/.hermes/.env.bak-cutover` |

---

## 2 · Tars → p-Hermes SSH leg — **PASS**

Both appends are guarded by a `grep -q` idempotence check (neither existed; both were appended).

```
$ ssh gaetan@192.168.0.9 'umask 077; ... append if absent ...'
appended phermes block
appended known_hosts entry
== config perms: 600  known_hosts: 600
== phermes block:
Host phermes
    HostName 192.168.0.8
    User hermes
    IdentityFile ~/.ssh/id_ed25519
    BatchMode yes
    ConnectTimeout 5
== known_hosts 0.8 line:
192.168.0.8 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAII3RCZUCVLoi862Q2hw6Cy5Tuu5dda3NgmoXfllnaNh9
```

known_hosts line is byte-identical to the host key supplied by the orchestrator (reported live from
VM 103). Public host key — not a secret.

Probe:

```
$ ssh gaetan@192.168.0.9 'ssh -o BatchMode=yes phermes true; echo "phermes_probe_exit=$?"'
phermes_probe_exit=0
```

Exit 0, no prompt, no host-key warning → the pre-seeded key matched and `gaetan@tars`'s public key is
already authorized for `hermes@192.168.0.8`. VM 103 and pve were not touched.

---

## 3 · Gateway start — **PASS**

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u);
    systemctl --user enable --now hermes-gateway.service; sleep 30; ...'
Created symlink /home/gaetan/.config/systemd/user/default.target.wants/hermes-gateway.service
  → /home/gaetan/.config/systemd/user/hermes-gateway.service.
enable_exit=0
== is-enabled: enabled
== is-active:  active
MainPID=60158   NRestarts=0   ExecMainStatus=0
ActiveEnterTimestamp=Fri 2026-08-07 18:43:08 UTC
```

Ordering: `.env` was written at ~18:42, the gateway started 18:43:08 — the token pair was on disk
before the process read it.

### 3a — journal (complete, only 4 lines at the unit's log level; nothing redacted needed)

```
$ journalctl --user -u hermes-gateway.service -n 80 --no-pager -o short-iso
2026-08-07T18:43:08 tars systemd: Started hermes-gateway.service - Hermes Agent Gateway - Messaging Platform Integration.
2026-08-07T18:43:12 tars python[60158]: WARNING slack_bolt.AsyncApp: As you gave `client` as well, `token` will be unused.
2026-08-07T18:43:16 tars python[60158]: WARNING gateway.run: Synthetic event source unresolvable: session_key='202608...d3d4' platform='' chat_type='' chat_id='' evt_type=async_delegation
2026-08-07T18:43:16 tars python[60158]: WARNING gateway.run: Dropping watch notification for raw session 20260807_175250_98d3d4: no api_server adapter to self-post through
```

No token fragment appeared in any journal line (nothing to mask). Grep over the full window:

```
$ journalctl --user -u hermes-gateway.service --since "-15 min" --no-pager \
  | grep -icE "invalid_auth|not_authed|account_inactive|token_revoked|Refusing to start|No env user allowlists"
0
```

The three WARNINGs are benign: `slack_bolt.AsyncApp` is Bolt noting Hermes passes a pre-built
`AsyncWebClient` (proof the Slack adapter initialized); the two `gateway.run` lines are a stale
pre-cutover session record (`20260807_175250`) with no adapter to replay through.

### 3b — Socket Mode connection is live (the unit does not log it, so proved by socket + DNS)

```
$ ss -tnp | grep 60158
ESTAB 0 0  192.168.0.9:57510  3.126.186.102:443

$ getent ahostsv4 wss-primary.slack.com
18.197.249.189  3.126.186.102  3.65.102.105  3.67.245.95  3.68.18.70  3.68.61.181  3.68.63.139
```

`3.126.186.102` is an A record of **`wss-primary.slack.com`** — Slack's Socket Mode WSS endpoint. The
gateway holds exactly one established outbound TLS session to it. A rejected `xapp-` token fails
`apps.connections.open` and no WSS session is ever opened, so this connection is itself the
positive auth proof; `NRestarts=0` with `Restart=always` confirms it did not crash-and-retry.

### 3c — `SLACK_ALLOWED_USERS` in effect

`/proc/60158/environ` does not list it — **expected and not evidence of absence**: that file is the
exec-time environment, and Hermes loads `~/.hermes/.env` into `os.environ` after exec (config
precedence CLI > config.yaml > .env > defaults). Proved instead from the gateway's own startup check:

`~/.hermes/hermes-agent/gateway/run.py:10940-11007`

```python
_builtin_allowed_vars = (..., "SLACK_ALLOWED_USERS", ...)
_any_allowlist = any(os.getenv(v) for v in _builtin_allowed_vars + _plugin_allowed_vars)
...
if not _any_allowlist and not _allow_all:
    logger.warning("No env user allowlists configured. ...")
```

That WARNING is **absent** from the journal (grep count 0 above) → `os.getenv("SLACK_ALLOWED_USERS")`
was truthy at startup → the allowlist is loaded and in effect. Consumer side confirmed at
`plugins/platforms/slack/adapter.py:6702` (`platform_allowlist = _env("SLACK_ALLOWED_USERS")`).
`GATEWAY_ALLOW_ALL_USERS` / `SLACK_ALLOW_ALL_USERS` are absent from `.env` (key-name list, §1d) — the
allow-all escape hatch is not set.

### 3d — CloakBrowser healthy, **not** restart-looping

```
== hermes-cloakbrowser.service: ActiveState=active SubState=running NRestarts=0
   MainPID=38431  ActiveEnterTimestamp=Fri 2026-08-07 16:43:21 UTC   (2h uptime, predates cutover)
== journal: {"ok": true, "cdp_url": "http://127.0.0.1:9223",
             "version": {"Browser": "Chrome/146.0.7680.177", "Protocol-Version": "1.3", ...}}
== ss -tlnp | grep 9223: LISTEN 127.0.0.1:9223  proc=chrome pid=38446   (one owner, no orphan)
```

The lane-B B7 trap (stale Chromium profile lock → chrome exit 21 → restart loop) did **not** fire:
`NRestarts=0`, a single healthy CDP owner on 9223, no `chrome exit 21` in the journal. No kill needed.

### 3e — config sanity (cutover-notes §5)

```
$ python3 -c 'yaml.safe_load(open("~/.hermes/config.yaml")) ...'
stray_top_level_platforms: False
slack: {'enabled': True, 'require_mention': True, 'strict_mention': True, 'unauthorized_dm_behavior': 'ignore'}
a2a:   {'enabled': True, 'extra': {'port': 9900}}
```

No stray top-level `platforms:` block. All three Slack guardrails present under
`gateway.platforms.slack`, matching `tars-profile.md` §2.

### 3f — stability re-check (T+3 min)

```
== gateway:       ActiveState=active SubState=running NRestarts=0 MainPID=60158 (unchanged)
== cloakbrowser:  ActiveState=active SubState=running NRestarts=0
== slack wss:     ESTAB 192.168.0.9:57510  3.126.186.102:443   (same socket, held 3+ min)
== auth errors:   0
```

`hermes gateway status` (via `~/.local/bin/hermes`, not on non-interactive PATH):

```
● hermes-gateway.service - Hermes Agent Gateway - Messaging Platform Integration
     Active: active (running) since Fri 2026-08-07 18:43:08 UTC
   Main PID: 60158 (hermes) · Tasks: 42 · Memory: 166.6M
     CGroup: ├─60158 python -m hermes_cli.main gateway run
             ├─60169/60173 mcp_stdio_watchdog → docker run ghcr.io/korotovsky/slack-mcp-server:v1.3.0 --transport stdio --no-cache
             └─60171/60175 mcp_stdio_watchdog → docker run mcp/notion@sha256:df0d678…
✓ User gateway service is running
✓ Systemd linger is enabled (service survives logout)
```

Both stdio MCP servers came up as children — `--no-cache` is present on the Slack MCP invocation
(the korotovsky #86 lazy-lookup constraint from D1 still holds).

---

## Per-step verdicts

| Step | Verdict |
|---|---|
| 1 · `.env` merge (2 secrets + `SLACK_ALLOWED_USERS`, 0600, 10→13 keys, backup taken) | **PASS** |
| 2 · SSH leg `tars → phermes` (config + known_hosts + probe exit 0) | **PASS** |
| 3 · gateway `enable --now`, active, Socket Mode up, no auth error | **PASS** |
| 3 · CloakBrowser healthy, no restart loop | **PASS** |
| 3 · config sanity (no stray top-level `platforms:`) | **PASS** |
| 4 · prohibitions honoured (no Slack message, no p-Hermes/pve, no whole-file `sops -d`, no git, no `lane-*.md`) | **PASS** |

## Deviations

None. Every step executed as written; every check returned the expected value.

## Followups (observations, not cutover failures)

1. **`/sethome` is still pending** — `SLACK_HOME_CHANNEL` is correctly absent from `.env`. Gaetan's
   step, in the live DM. Until then the gateway has no home channel.
2. **`GITHUB_PAT` is absent from `~/.hermes/.env`** — `tars-profile.md` §3 lists it (SOPS `GITHUB_PAT`,
   lane A1), and it is also absent from `secrets/tars.sops.yaml` (SOPS holds `NOTION_TOKEN_V2`,
   `NOTION_FILE_TOKEN`, `NOTION_SPACE_ID`, `LINEAR_API_KEY`, `GMAIL_ADDRESS`, `GMAIL_APP_PASSWORD`,
   `NOTION_API_TOKEN`, `SLACK_TOKEN`, `SLACK_COOKIE`, `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`).
   Pre-existing gap, unrelated to cutover, left untouched.
3. **`HINDSIGHT_API_KEY` / `HINDSIGHT_LLM_API_KEY` absent** — the B7 open item is still open;
   `HINDSIGHT_MODE` is set alone. Did not block startup.
4. **Stale pre-cutover session `20260807_175250_98d3d4`** produced two startup WARNINGs (no adapter to
   self-post through). Harmless; will stop once a new session supersedes it.
5. **`rtk` plugin warning** on the VM: `rtk binary not found in PATH; Hermes hook not registered`.
   Cosmetic, from the `hermes` CLI wrapper only.
6. **Rollback, if ever needed** (cutover-notes §3): `systemctl --user disable --now
   hermes-gateway.service` on the VM, restore `~/.hermes/.env.bak-cutover`, then re-enable
   `hermes-gateway-tars.service` on p-Hermes. The old profile was not deleted.
