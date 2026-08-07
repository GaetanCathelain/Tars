# R3 — personal-Hermes probe: how the Tars Slack app is wired TODAY

Read-only recon for the WF1 barrier. All commands run against the live
machine; no writes, installs, restarts, or deletes performed.

## Verdict

The Tars Slack app runs **today** as the `tars` profile of Hermes Agent
v0.19.0 on personal Hermes (Proxmox VM 103, `hermes`, on hypervisor `pve` at
`192.168.0.3`), driven by one `systemd --user` unit
(`hermes-gateway-tars.service`, enabled, currently running, Slack platform
state `connected`). The cutover's destructive step is precisely scoped: one
profile directory, one unit file, one override drop-in — no shared state with
the other five profiles or with the default/justine gateways. No Docker
involved anywhere in the Hermes stack; everything is a Python venv process
under systemd user units with `Linger=yes`. Access to the VM is LAN-only via
the Proxmox hypervisor's `qm guest exec` (the VM refuses direct SSH); no
tailscale hop is needed because `cooper` (this box, `192.168.0.4`) shares the
`192.168.0.0/24` LAN with `pve` (`192.168.0.3`) directly.

## Facts with evidence

### 1. p-Hermes' address — LAN, not tailscale

- `~/dev/gaetan-metarepo/hermes/README.md` **does not exist**. The authoritative
  doc is `~/dev/gaetan-metarepo/docs/specs/2026-07-28-gaetan-metarepo-design.md`
  §3.5 ("Hermes integration"), line 475-477:
  > "Confirmed access (finder, 2026-07-28): personal Hermes runs in Proxmox VM
  > `hermes` (VMID 103, 192.168.0.8) on hypervisor `pve` = 192.168.0.3; path in
  > = key-based root SSH to the hypervisor + `qm guest exec 103` (the VM
  > refuses direct SSH)."
- Also corroborated by `~/dev/gaetan-metarepo/learnings/hermes-vm.md`: prior
  ingestion used the identical path `ssh root@192.168.0.3` → `qm guest exec
  103` → `runuser -u hermes`.
- No tailscale hostname for Hermes exists in any doc searched (`machine/*.md`,
  `docs/specs/*.md`, `docs/plans/*.md`) or in live `tailscale status` output
  (`grep -i hermes` on the full status list: zero hits). `pve`/`hermes` VM is
  plain-LAN only.
- Verified reachable and used for every probe below:
  ```
  $ ip route
  192.168.0.0/24 dev enp6s18 proto kernel scope link src 192.168.0.4
  $ ssh -o BatchMode=yes root@192.168.0.3 "echo SSH_OK; qm list"
  SSH_OK
        VMID NAME                 STATUS     MEM(MB)    BOOTDISK(GB) PID
         100 workstation-01       running    18432            120.00 1384818
         101 homeassistant        running    4096              32.00 1312175
         103 hermes               running    8192             100.00 1274904
  ```

### 2. Hermes install — version, binary, runtime

Probe: `qm guest exec 103 -- /bin/bash -lc 'runuser -u hermes -- bash -c "hermes --version"'`

```
Hermes Agent v0.19.0 (2026.7.20) · upstream 0a62610f · local d9da213a (+3 carried commits)
Install directory: /home/hermes/.hermes/hermes-agent
Install method: git
Python: 3.11.15
OpenAI SDK: 2.24.0
```

- `whoami` / `$HOME` / `hostname` / `uname -a` inside the VM: user `hermes`,
  `HOME=/home/hermes`, hostname `hermes`, `Linux hermes 6.8.0-136-generic
  #136-Ubuntu SMP … x86_64`.
- Binary invoked by the gateway service: `/home/hermes/.hermes/hermes-agent/venv/bin/python -m hermes_cli.main --profile <name> gateway run` — a Python venv process, **not a container**. No Docker daemon or `docker`/`podman` references found anywhere in the systemd unit set for Hermes.

### 3. Profiles — six total, `tars` confirmed

Probe: `ls -la /home/hermes/.hermes/profiles/`

```
drwx------ 18 hermes hermes  immichbackup
drwx------ 18 hermes hermes  immichresearch
drwx------ 31 hermes hermes  justine
drwx------ 18 hermes hermes  papainventory
drwx------ 17 hermes hermes  privatelocal
drwx------ 22 hermes hermes  tars
```

Plus the **default** profile at `/home/hermes/.hermes` itself (not under
`profiles/`) — per the design doc, the default profile is surfaced over
Telegram, not Slack, and is out of scope for this cutover.

`tars` is the only profile wired to the "Tars" Slack app (see §5). `justine`
has its own separate gateway + CloakBrowser instance (`hermes-gateway-justine.service`,
`hermes-cloakbrowser-justine.service`) — a different Slack/messaging surface,
untouched by this cutover.

`profile.yaml` inside the tars profile dir:
```
description: Restricted Slack bot profile for the Tars bot. Only the owner should
  be allowed to interact; tool access should remain minimal.
description_auto: false
```

### 4. How the service runs — systemd --user, not Docker

Probe: `systemctl --user list-unit-files | grep -i hermes` (as user `hermes`, via `runuser -l hermes`)

```
hermes-cloakbrowser-justine.service     enabled
hermes-cloakbrowser.service             enabled
hermes-dashboard-lan.service            enabled
hermes-gateway-justine.service          enabled
hermes-gateway-tars.service             enabled     <-- Tars
hermes-gateway.service                  enabled     <-- default profile
hermes-reconcile-kill-old-20260724.service   static   (unrelated: one-off reconcile from a past version bump)
hermes-reconcile-restart-20260724.service    static
hermes-reconcile-restart-20260724.timer      disabled
```

`systemctl --user is-enabled hermes-gateway-tars.service` → `enabled`.

Live process (confirms it's actually running, not just enabled):
```
$ ps -eo pid,cmd | grep tars
1268 /home/hermes/.hermes/hermes-agent/venv/bin/python -m hermes_cli.main --profile tars gateway run
```

`gateway_state.json` (profile dir) confirms Slack connectivity, last updated
2026-08-05:
```
{"pid":1268,"kind":"hermes-gateway","gateway_state":"running",
 "platforms":{"slack":{"state":"connected","error_code":null,"error_message":null,
 "updated_at":"2026-08-05T22:41:59.993628+00:00"}}, ...}
```

Also running for OTHER profiles (do not touch): `hermes-gateway.service`
(default/Telegram), `hermes-gateway-justine.service`, two CloakBrowser
instances, `hermes-dashboard-lan.service`, plus unrelated `gbrain-mail-collect`,
`gbrain-sms-collect`, `gbrain-whatsapp-webhook` services/timers (a separate
GBrain project, per-profile `.service` files owned by user `hermes`, mode 600,
in the same unit directory — cosmetically adjacent but functionally
independent).

### 5. Slack app attachment — config file locations, app/bot identifiers

**Exact paths to touch at cutover (and nothing else):**

| Path | What it is |
|---|---|
| `/home/hermes/.hermes/profiles/tars/` | entire profile home dir (config, `.env`, `auth.json`, memories, sessions, state.db, skills, cron, etc.) |
| `/home/hermes/.config/systemd/user/hermes-gateway-tars.service` | the gateway unit file (`ExecStart=... --profile tars gateway run`) |
| `/home/hermes/.config/systemd/user/hermes-gateway-tars.service.d/override.conf` | drop-in setting `SLACK_HOME_CHANNEL` |
| `/home/hermes/.config/systemd/user/hermes-gateway-tars.service.d/` | the drop-in dir itself, once the override is gone |
| `/home/hermes/.config/systemd/user/default.target.wants/hermes-gateway-tars.service` | the systemd-enable symlink → the unit file above |

No other file anywhere on the VM referenced `tars` in a way that looked
gateway-related (no separate CloakBrowser service for `tars`, no separate
dashboard entry).

**Unit file** (`systemctl --user cat hermes-gateway-tars.service`):
```
[Unit]
Description=Hermes Agent Gateway - Messaging Platform Integration
After=network-online.target
[Service]
Type=simple
ExecStart=/home/hermes/.hermes/hermes-agent/venv/bin/python -m hermes_cli.main --profile tars gateway run
WorkingDirectory=/home/hermes/.hermes/profiles/tars
Environment="HERMES_HOME=/home/hermes/.hermes/profiles/tars"
Restart=always
RestartSec=5
[Install]
WantedBy=default.target

# .d/override.conf
[Service]
Environment="SLACK_HOME_CHANNEL=D0BBYNM01BL"
```
(`D0BBYNM01BL` is a Slack DM/channel ID — not a secret, kept here as the
"home channel" identifier the `/sethome` DM in the cutover step will need to
reproduce or replace on the new side.)

**App/bot identifiers** — from `slack-manifest.tars.redacted.json`
(13,326 bytes, world-readable 664, meant to be shared) inside the profile dir:
- App display name: `Tars`; bot user display name: `Tars`; background color `#1a1a2e`
- `socket_mode_enabled: true`, `token_rotation_enabled: false`, `org_deploy_enabled: false` — Socket Mode, so no public inbound webhook URL is actually load-bearing (the manifest lists `https://hermes-agent.local/slack/commands` as the slash-command URL, but that host does not resolve publicly — Socket Mode carries the traffic).
- Bot OAuth scopes: `app_mentions:read, assistant:write, channels:history, channels:read, chat:write, commands, files:read, files:write, groups:history, groups:read, im:history, im:read, im:write, mpim:read, users:read`
- Event subscriptions: `app_mention, assistant_thread_context_changed, assistant_thread_started, message.channels, message.groups, message.im`
- ~50 slash commands registered under `/tars`, `/bg`, `/new`, `/sethome`, etc. (full list in the manifest; the exhaustive command menu is what the new VM's app must replicate for parity).
- A non-redacted copy also exists at the same size (`slack-manifest.json`, same 13,326 bytes) — sizes matching strongly suggests the manifest itself carries no secret (Slack app manifests never embed tokens; scopes/commands only). Not opened, to stay strictly on the safe side.

**Credential files present (existence/shape only, values never read):**
- `/home/hermes/.hermes/profiles/tars/.env` (518 bytes, mode 600) — key names only, values redacted server-side before ever reaching this session:
  ```
  SLACK_BOT_TOKEN=[REDACTED]
  SLACK_APP_TOKEN=[REDACTED]
  SLACK_ALLOWED_USERS=[REDACTED]
  SLACK_ALLOW_ALL_USERS=[REDACTED]
  SLACK_REQUIRE_MENTION=[REDACTED]
  SLACK_STRICT_MENTION=[REDACTED]
  SLACK_ALLOW_BOTS=[REDACTED]
  SLACK_HOME_CHANNEL=[REDACTED]
  ```
  (`SLACK_BOT_TOKEN` + `SLACK_APP_TOKEN` is the standard Slack Bolt Socket-Mode
  pair — bot token + app-level token — consistent with `socket_mode_enabled:
  true` in the manifest.)
- `/home/hermes/.hermes/profiles/tars/auth.json` (2,573 bytes, mode 600) — top-level keys only: `version, providers, credential_pool, updated_at`. This looks like the LLM-provider credential store (OpenAI/Anthropic/etc.), separate from the Slack tokens in `.env`. Not opened beyond key names.
- `slack.require_mention: true`, `slack.strict_mention: true`, `slack.unauthorized_dm_behavior: ignore` in `config.yaml` (12,499 bytes, mode 600) — the behavioural guardrail config (mention required, unknown DMs ignored). This is the config surface WF4's negative test (non-Gaetan message ignored) should target/reproduce on the new VM.

### 6. Confirmed NOT relevant to this cutover (adjacent, do not touch)

- `hermes-gateway.service` (default profile, Telegram) and
  `hermes-gateway-justine.service` (own Slack-like surface) — separate
  profiles, separate unit files, separate `.env`/tokens.
- `hermes-cloakbrowser.service` / `hermes-cloakbrowser-justine.service` — no
  `hermes-cloakbrowser-tars.service` exists; the `tars` profile does not run
  its own CloakBrowser instance.
- `gbrain-mail-collect.{service,timer}`, `gbrain-sms-collect.{service,timer}`,
  `gbrain-whatsapp-webhook.service` — unrelated GBrain project units living in
  the same `~/.config/systemd/user/` directory.
- `hermes-reconcile-kill-old-20260724.*` / `hermes-reconcile-restart-20260724.*`
  — generic one-off reconciliation units from a past version bump, not
  tars-specific, timer already disabled.

## Blockers

None. All probes succeeded read-only; no access denied, no missing tooling.

## Open questions

- The manifest lists a slash-command URL at `https://hermes-agent.local/slack-commands`-style
  hostname that will not resolve from the new VM either — confirm during WF3
  wiring whether Socket Mode alone is sufficient for 100% of app functionality
  (it appears to be, given `socket_mode_enabled: true`) or whether that URL
  field is genuinely dead weight in the manifest.
- `SLACK_HOME_CHANNEL` is set in **two** places (`.env` key, present but
  redacted-empty-looking in the file dump above since the visible value was
  masked, and the systemd override `D0BBYNM01BL`) — confirm which one wins at
  runtime (systemd `Environment=` in a `.d/override.conf` overrides the
  process env before `.env` is loaded by the app, typically making the
  override authoritative) before deciding what the new VM's `/sethome` needs
  to reproduce.
- Not verified: whether the Slack app itself (the App ID in the Slack API
  dashard) is single-workspace-installed only to this bot, or whether
  reinstalling scopes on a new backend (new VM) requires reauthorizing in the
  Slack UI — that's a Slack-side action for lane A / the cutover step, out of
  scope for this read-only VM probe.
