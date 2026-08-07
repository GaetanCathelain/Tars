# DECISION — WF1 synthesis

Settles the six open calls from WF1 recon (R1–R6). Binding for lane A, lane B, WF3, the cutover
and WF4. Evidence is cited to the recon artifact that produced it; nothing here is new research.

---

## ⛔ Blockers — unresolved, read first

1. ~~Model backend has no recon coverage and no credential path.~~ **RESOLVED (grill + R7,
   2026-08-07).** Gaetan corrected the premise: p-Hermes already runs on the ChatGPT 5.6
   subscription. R7 (`r7-model-backend.md`) confirmed live: provider **`openai-codex`**,
   ChatGPT-subscription OAuth via device-code flow — `hermes auth add openai-codex --type oauth`,
   verify `hermes auth status openai-codex` → logged in; `model.default: gpt-5.6-sol`,
   `base_url: https://chatgpt.com/backend-api/codex`, credential lands in the base profile's
   `auth.json` and self-manages. **Do NOT copy p-Hermes' `auth.json` and do NOT store this in
   SOPS** — rotating refresh token; copying risks invalidating the live p-Hermes session. Lane A
   runs the OAuth flow fresh on the new VM (A1b), records only "flow run, verified" in
   `status/lane-a.md`. Lane B still does not block on it; wired in WF3.
2. ~~Which Slack workspace member ID is Gaetan's~~ **RESOLVED 2026-08-07 by A4's probe:
   `SLACK_ALLOWED_USERS=U08BDJAMSRZ`** (user `gaetan.cathelain`, team `Mobile Club` / `T7V1UGJ82`,
   `mobileclub-squad.slack.com`). Cutover writes it into the new profile's `.env`.

Everything else the recon flagged as a blocker is resolved below (op-signin → dropped by D5;
tailnet ambiguity → resolved by D4; `/sethome` needing a live DM → correctly placed at the gate;
Slack-app-transport question → resolved by D2/D5).

---

## D1 — Slack personal-account access: harvest xoxc/xoxd, consume via an MCP server, not the gateway

**Decision:** Lane A harvests the `xoxc-` token + `xoxd-` cookie from Gaetan's logged-in Slack
**web** tab exactly as `~/dev/mc-kestra/SETUP.md` §5 documents. Tars consumes them through
**`korotovsky/slack-mcp-server`, run as a Docker container on the Tars VM**, registered in the Tars
Hermes profile's `config.yaml` under `mcp_servers` (`SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN`).
No custom HTTP client, no browser automation, no legacy token.

**Rationale (R2):** the harvest transfers unchanged and is independently corroborated by the MCP
server's own auth docs (identical DOM paths). Hermes' *gateway* Slack integration is bot-OAuth-only
(Bolt + Socket Mode, `xoxb-`/`xapp-`) and cannot consume personal session tokens at all — so the
personal account is a different consumption layer, not a competing credential. Hermes ships a
native MCP client that live-reloads on `config.yaml` edit, which makes this a config change rather
than code. Legacy workspace tokens are no longer issuable; UI automation is strictly worse.

**Binding constraints carried forward:**
- The MCP server must do **lazy user lookups, not eager full-workspace caching at startup** —
  korotovsky issue #86: bulk user enumeration is what gets xoxc/xoxd invalidated. Verify the
  deployed config before declaring the wiring done.
- Plan re-harvest at **~1 year**, not mc-kestra's stale "~5 years" (Slack cut cookie max TTL in
  Dec 2025). Failure mode is "re-harvest", never "debug".
- **CORRECTED by A4's probe (2026-08-07): the stored cookie is URL-ENCODED and works as-stored**
  in a `Cookie: d=` header. D1's original "store decoded, consumer encodes" convention was wrong.
  WF3 must confirm `korotovsky/slack-mcp-server` does not encode again on top (double-encoding
  risk) — if it does, store a decoded copy instead.
- Rate limits: the 2026 tightening scopes to registered apps, not the web-client auth path. Treated
  as ample for Tars' volume; inferential, not a cited number.

**Consequence for lane B:** B2 must install Docker on the VM, so WF3 can `docker compose up` this
server. That is one package, not a new step.

## D2 — Hermes: install latest release, Tars **is** the base profile; cutover deletes exactly five paths

**Install (lane B, B4):**
```
curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash
```
on Ubuntu 24.04 (systemd + glibc + FHS = the officially tested target). Take the current release
(**v0.20.0 / tag `v2026.8.3`**) — no pin to p-Hermes' older v0.19.0; A2A v1.0 ships in v0.20.0 and
is wanted. Smoke check = `hermes doctor` + `hermes status` + `hermes dump`.

**Profile mechanics — decision on R1's open Q1: option (a), configure the base `~/.hermes`
directly. No `hermes profile create`.** This VM runs one agent and nothing else; `~/.hermes` *is*
the default profile, so a named profile buys an extra path segment and an alias for zero benefit.
Concretely:
- `HERMES_HOME` unset → `~/.hermes`; identity in `~/.hermes/SOUL.md`; non-secrets in
  `~/.hermes/config.yaml`; secrets in `~/.hermes/.env` (0600).
- Gateway unit is therefore `~/.config/systemd/user/hermes-gateway.service` (no `-tars` suffix),
  plus `sudo loginctl enable-linger` so it survives logout/reboot.
- "No coding, no PRs, orchestrate/report only" goes in **SOUL.md** (durable identity, prompt slot
  #1), not AGENTS.md — R1's documented precedence stack.
- Guardrails copied from the live old profile (R3): `slack.require_mention: true`,
  `slack.strict_mention: true`, `slack.unauthorized_dm_behavior: ignore`, plus
  `SLACK_ALLOWED_USERS`. This is the exact surface WF4's negative test targets.
- **`SLACK_HOME_CHANNEL` lives in `.env` only** — do not reproduce p-Hermes' systemd
  `override.conf`. R3 found it set in two places with unconfirmed precedence; one source of truth
  removes the question instead of answering it. `/sethome` in the live DM at cutover sets/confirms
  it (messaging-only command, cannot be pre-scripted).
- Plugins (B5): **as-executed spellings (lane B B5 verdict 2026-08-07 — 5 of 10 documented
  spellings were wrong, two cancel-with-exit-0 headless):** rtk via its direct `curl | sh`
  installer then `rtk init --agent hermes` (no `-g` — proven no-op; verify via `hermes plugins
  list`, NOT `rtk init --show` which false-negatives); hindsight via `hermes config set
  memory.provider hindsight` (NOT `hermes memory setup` — cancels headless with exit 0),
  `HINDSIGHT_MODE=local_embedded` canonical (`local` is a legacy alias); `hermes plugins install
  stephenschoettler/hermes-lcm --enable` + `context.engine: lcm` (not `hermes-lcm`); `hermes
  skills install ayghri/i-have-adhd/skills/i-have-adhd --yes` (without `--yes`: exit-0 cancel).
- A2A: enable **inbound only** at install (config key path UNCONFIRMED — live v0.19.0 has `platforms:` top-level, not under `gateway:`; lane B determines the v0.20.0 path empirically at B4. Port 9900,
  localhost-bound by default). Outbound toolset stays off until something concretely needs Tars to
  call another agent. YAGNI; flipping it later is one `hermes tools` call.

**Cutover target on p-Hermes — the exact and complete delete-set (R3):**
Reached via `ssh root@192.168.0.3` → `qm guest exec 103` → `runuser -u hermes` (VM 103 `hermes`,
192.168.0.8; the VM refuses direct SSH).
```
/home/hermes/.hermes/profiles/tars/
/home/hermes/.config/systemd/user/hermes-gateway-tars.service
/home/hermes/.config/systemd/user/hermes-gateway-tars.service.d/override.conf
/home/hermes/.config/systemd/user/hermes-gateway-tars.service.d/          (the dir, once empty)
/home/hermes/.config/systemd/user/default.target.wants/hermes-gateway-tars.service  (enable symlink)
```
Preceded by `systemctl --user disable --now hermes-gateway-tars.service` (live pid 1268,
Slack state `connected` as of 2026-08-05).
**Do not touch:** `hermes-gateway.service` (default/Telegram), `hermes-gateway-justine.service`,
both `hermes-cloakbrowser*` units (no tars-specific instance exists), `hermes-dashboard-lan.service`,
`gbrain-*`, `hermes-reconcile-*-20260724.*`.

**Why the order is load-bearing, not politeness:** one Slack token pair can only be live on one
profile's gateway at a time — a second gateway holding the same token refuses to start and names
the conflicting profile (R1, vendor docs). Delete-then-attach is a technical requirement.
Reuse needs no Slack-side reinstall: `token_rotation_enabled: false` and Socket Mode mean the
`xoxb-`/`xapp-` pair is static and host-independent — copy the values into the new profile's `.env`,
update `SLACK_ALLOWED_USERS`, done. Re-invite the bot per channel afterwards (bots do not auto-join).

## D3 — Rendezvous and cutover gate: plain SendMessage + committed status files. No Orca orchestration layer.

**Decision:** keep PLAN.md's coordination contract exactly as written. Lane B pings the orchestrator
via `SendMessage` at VM-created / Hermes-installed / smoke-pass and immediately on blockers;
`status/lane-b.md` committed to `main` is the durable fallback. The cutover gate is a plain inline
stop-and-wait for Gaetan's "go" in the orchestrator session. **No `run-create`, no `task-create`,
no `dispatch`, no `gate-create`.**

**Rationale (R4):** three independent reasons, any one sufficient.
1. Orca's own bundled orchestration guide classifies "hand off to another worktree, unsupervised" —
   verbatim the shape of lane B — as a **full handoff**, and says outright not to use
   `task-create` / `dispatch --inject` / `check --wait` for those.
2. No Run is bound: `task-list` and `gate-list` both return `run_required`, `worker-list` is empty.
   Adopting the layer means bootstrapping fresh state as a *second* coordination channel alongside
   the status-file contract that already exists.
3. A gate is not a standalone primitive — `gate-create` requires `--task`, so the cutover would
   first have to be modelled as an Orca Task. And it would gate nothing real: the human and the
   executor are the same seat (the orchestrator session is Gaetan's session), so enforcement is
   still "the session chooses not to act", with bookkeeping bolted on.

Cross-session `SendMessage`/`ListAgents` is proven live right now (6 peer sessions, including this
project's own). PLAN.md's contract is ORCHESTRATION-POLICY.md §9 applied to Tars — the well-trodden
path on this machine, not an invention. Spawn stays: `orca worktree create --name lane-b --repo
path:<repo-root> --base-branch main --json` → `orc-tab start -w path:<new> orc-opus lane-b '<prompt>'`.

Accepted cost: no typed `worker_done` and no liveness registry beyond `ListAgents`' idle state. The
committed status file covers the durability need; that is the whole reason it exists.

## D4 — Proxmox: GO. Node `pve`, VMID 102, cloud image + cloud-init, `local-lvm`, `vmbr0`.

**Decision:** provision on node **`pve` (192.168.0.3)** — single node, not clustered — as **VMID
102** (`pvesh get /cluster/nextid` → 102; no old Tars VM exists anywhere). 8 cores / 8G RAM / 50G
disk as specified.

**Mechanism: cloud image + cloud-init.** `local:iso/noble-cloudimg-amd64.img` (Ubuntu 24.04)
imported as the disk on **`local-lvm`** (the only storage with `content images`), a cloud-init drive
(`ide2: local-lvm:vm-102-cloudinit`), custom user-data as a snippet under `/var/lib/vz/snippets/`
via `cicustom`, `agent: enabled=1`, net0 on **`vmbr0`** (the only bridge). **No template-clone** —
all three existing VMs report `template=None` and no vztmpl/template cache exists on the host.
**No ISO install** — the 24.04 desktop ISO is present but autoinstall has zero precedent here,
whereas cloud-init is the exact mechanism already in production for VM 103 (`hermes-user.yaml`
exists, VM103 carries a live cloudinit drive). Read `hermes-user.yaml` on the host at execution time
if lane B wants to mirror it; it may carry keys, so read it there, never copy it into this repo.

**Capacity, verified:** RAM 8G fits inside 25GiB available with ~17GiB spare. Disk: `local-lvm` has
~62.9GiB free — a thin-provisioned 50G volume (real usage 10–20GiB for Ubuntu+desktop) fits.
**Watch items, not blockers:** the thin pool is already 82.65% nominally allocated (little slack for
the VM *after* Tars); configured vCPUs go to 30 on a 24-thread host (routine overcommit, host load
avg ~2.05, but workstation-01 and Tars both CPU-bound at once will contend).

**GUI scope — settling R5's open question: minimal X, not a remote desktop.** Install
`xserver-xorg` + a minimal WM/`xubuntu-desktop-minimal` for a local display so headed-browser
automation (CloakBrowser, B6) has somewhere to render. **No xrdp/VNC** — nothing in the pitch needs
an interactive desktop session, and the Proxmox SPICE/VNC console already exists for the rare
hands-on case. Add xrdp later if a real need appears.

**Networking — settling R6's tailnet blocker: LAN is the primary path; tailscale is additive.**
p-Hermes and Proxmox are **not on any tailnet** (R3: zero hits for hermes in `tailscale status`;
plain-LAN only) and cooper (192.168.0.4) shares `192.168.0.0/24` with them. So the Tars VM gets a
LAN address on `vmbr0` and cooper↔Tars↔pve all work over LAN with no tailnet dependency — lane B is
**not blocked** on the tailnet question. Still run `tailscale up` and join **the same tailnet cooper
is on** (`workstation-vm`, owner `gaetan.cathelain@`) — it is the only tailnet Gaetan has an
identity on from this machine, and it is what gives the macOS leg of the SSH mesh a path when it is
off-LAN. If Tars turns out not to need off-LAN reach at all, dropping tailscale is a later
simplification, not a blocker.

## D5 — Credential map, lane A order, one store: SOPS/age

**Storage decision: SOPS + age is the single canonical store. 1Password is not a build dependency.**
`secrets/tars.sops.yaml` in this repo, `.sops.yaml` recipients = cooper's age public key + the Tars
VM's own age key (generated at provisioning, public half enrolled with
`grep -oE 'age1[0-9a-z]{58}'`). Rationale: Hermes reads secrets from a profile-local `.env`/`auth.json`
(0600) at runtime — R6 evidenced this on a real deployment — so a decrypt-to-local-file deploy step
is required *whatever* the at-rest store is. SOPS does that with a file and a key already in
Gaetan's practice; 1Password on a headless VM would need a service-account token, i.e. one more
credential minted to fetch credentials. **This removes R6's `op signin` blocker entirely** — lane A
no longer depends on it. 1Password's role stays what the security pref actually says it is: sharing
a secret with a *person*, via `op item share`. Not needed here.
Naming: SOPS keys reuse mc-kestra's env-var names verbatim so the walkthrough probes map 1:1.

**Order and content — PLAN order, with two corrections (see D6):**

| # | Credential | Harvest | Probe (secret never on argv) | SOPS key(s) |
|---|---|---|---|---|
| A1 | GitHub PAT (classic, `repo`+`read:org`) — gates the VM's mc-metarepo clone | github.com/settings/tokens | `curl -K` with an `Authorization` header file → `api.github.com/repos/mobile-club/metarepo` returns `.private == true`; then `git ls-remote` via `GIT_ASKPASS` | `GITHUB_PAT` |
| A1b | **ChatGPT-sub OAuth / model backend** (R7) | `hermes auth add openai-codex --type oauth` on the new VM — device-code flow completed in Gaetan's browser | `hermes auth status openai-codex` → logged in; then `model.default: gpt-5.6-sol` in config.yaml | **none — do not store** (self-rotating OAuth in `auth.json`; never copy from p-Hermes) |
| A2 | Linear API key | linear.app → Settings → Security & access → Personal API keys | `-K` header file, GraphQL `{ viewer { id name } }` → 200. **Bare key, no `Bearer` prefix** | `LINEAR_API_KEY` |
| A2 | Notion (3 fields, one item) | devtools on the Quick-Find `search` XHR: payload `spaceId`, cookie `token_v2`; `file_token` only appears **after one manual export** | replay the `search` XHR server-side → 200 with a results envelope. `file_token` is not cheaply probeable — first real export is its probe | `NOTION_TOKEN_V2`, `NOTION_FILE_TOKEN`, `NOTION_SPACE_ID` |
| A3 | Gmail app password (**mint a second one, labelled for Tars** — independent blast radius from mc-kestra's) | myaccount.google.com/apppasswords, needs 2FA already on | `curl -s --url imaps://imap.gmail.com --user "$ADDR:$PW"` — cheapest auth check, works without Himalaya installed | `GMAIL_ADDRESS`, `GMAIL_APP_PASSWORD` |
| A3 | Google Calendar | **none — rides on the Gmail app password** | `curl -u … -X PROPFIND -H "Depth: 0" https://www.google.com/calendar/dav/$ADDR/events` → **207 Multi-Status with inner 200** (PROPFIND never returns bare 200 — RFC 4918; probe-corrected 2026-08-07) | (same as above) |
| A4 | Slack personal (`xoxc-` + `xoxd-`) | logged-in Slack **web** tab devtools, per D1 | `curl -b "d=$COOKIE" -F "token=$TOKEN" slack.com/api/auth.test` → `ok:true`. Also returns Gaetan's member ID → feeds Blocker 2 | `SLACK_TOKEN`, `SLACK_COOKIE` |
| A5 | Slack Tars app: `xoxb-` bot + `xapp-` app-level. **Socket Mode. No signing secret.** Held until cutover | api.slack.com/apps → Tars → OAuth & Permissions (bot); Basic Information → App-Level Tokens, scope `connections:write` | `auth.test` with the bot token; `POST apps.connections.open` with the app token → `ok:true`, **do not connect to the returned `wss://`** | `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN` |
| LB | Tailscale auth key | login.tailscale.com/admin/settings/keys | inseparable from consumption — `tailscale up --authkey` succeeding, then `tailscale ping tars` from cooper | mint fresh, pipe into lane B; **do not store** (single-use, spent on first use) |
| LB | VM SSH keypair (mesh) | generated, not harvested: `ssh-keygen -t ed25519` | `ssh -i <key> -o BatchMode=yes -o ConnectTimeout=5 <host> true` → exit 0 | `TARS_VM_SSH_PRIVATE_KEY` |

R6's open question 4 (Slack transport) is **settled**: Socket Mode, `xoxb-` + `xapp-`, no signing
secret — R1's vendor docs and R3's live `socket_mode_enabled: true` agree. Open question 6 (Orca
pairing vs raw SSH) is **settled by D3**: raw SSH only; no Orca pairing code, so no 9th item.
GitHub PAT expiry: **1 year with a calendar reminder**, not no-expiry — an unattended agent holding
a never-expiring `repo` token is the one standing risk here worth one reminder to avoid.

## D6 — Deviations from PLAN.md

Six. None fatal; four are corrections, two are additions.

1. **The credential map is 9 targets, not 8.** PLAN.md commits to GPT 5.6 Sol via ChatGPT OAuth
   (§Model backend) but R6's brief enumerated 8 and omitted it. See Blocker 1. Lane A owns it.
2. **Lane A's A2/A3 order is backwards for Calendar.** PLAN puts Calendar in A2 and Gmail in A3, but
   Calendar has no credential of its own — it rides on the Gmail app password (R6 §4). A2 = Linear +
   Notion; the Gmail app password is minted in A3 and the Calendar CalDAV probe runs immediately
   after it.
3. **"SOPS/1Password" resolves to SOPS only** (D5). PLAN left it as an either/or; picking both would
   mean two stores drifting. 1Password is not on the critical path and `op signin` stops being a
   gate.
4. **`hermes profile create tars` is not run** (D2). PLAN and the pitch say "Tars is the default
   profile"; the literal reading is the base `~/.hermes`, which also changes the unit name from
   `hermes-gateway-tars.service` to `hermes-gateway.service` on the *new* VM. The old p-Hermes unit
   keeps its `-tars` name and is deleted as named in D2 — do not confuse the two.
5. **Lane B gains Docker in B2** (D1) — the personal-Slack path needs a container host, which is not
   in PLAN's B-chain. One package at provision time; WF3's slack-personal agent deploys the server.
   Also B6 CloakBrowser has no p-Hermes precedent for the tars profile specifically (R3: no
   `hermes-cloakbrowser-tars.service` exists) — it is genuinely new on the Tars side, so budget it
   as new work rather than a copy.
6. **Proxmox has no template** (R5) — PLAN's R5 node says "capacity · template"; the answer is that
   template-clone is unavailable and the mechanism is cloud image + cloud-init (D4). Also **no
   tailnet dependency for lane B** — PLAN's B2 implies tailscale is the addressing plan; LAN is, and
   tailscale is additive (D4).

Not a deviation, recorded for completeness: PLAN's guardrail asked WF1 to *evaluate* Orca's
orchestration layer for the cutover gate and lane dispatch rather than assume. It was evaluated and
rejected — D3.

---

## Confidence and what stays unverified

- Hermes vendor facts (R1) are LLM-summarized WebFetch reads cross-checked against one live local
  plugin and two independent live deployments — **verify exact flag spellings with `--help` on the
  VM** rather than copy-pasting. This applies to `hermes gateway install`/`setup`, the
  `context.engine:` value for hermes-lcm, and whether `rtk init --agent hermes` wants `-g`.
- Everything in R3 (cutover delete-set, paths, identifiers) and R5 (capacity, VMIDs, storage) is
  live-probed on the real machines. Highest-confidence material in this document.
- No numeric rate limit exists for the xoxc/xoxd path; "generous enough" is inferential (D1).
- Himalaya's exact probe syntax is unverified (not installed anywhere); the `curl imaps://` fallback
  in D5 needs no install and is the probe of record until Himalaya is on the VM.
