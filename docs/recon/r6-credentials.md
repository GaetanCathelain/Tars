# R6 — Credential map: every credential Tars needs

Recon probe for WF1. Read-only throughout — no signin, no token minted, no file written to
personal Hermes. Prior art read first per task: `~/dev/mc-kestra/SETUP.md` §5 and
`~/dev/mc-kestra/setup-calendar-section.md` (see PLAN.md § "Prior art"). All 6 of those
credential shapes carry over near-verbatim; only storage differs (SOPS/1Password, never
`.env`).

## Verdict

Prior art settles harvest + probe shape for **6 of 8** targets (GitHub, Linear, Notion, Calendar,
Gmail-transport, Slack-personal) — Lane A can follow those walkthroughs directly, storage swapped
to SOPS/1Password. The **2 Tars-specific** targets (Slack Tars app tokens, Orca/SSH mesh) have no
mc-kestra precedent and are worked out below from first principles + local evidence. Storage
mechanics have a real gap: `op` CLI is installed but **not signed in** (interactive step, belongs
to Lane A, not this recon); no vault list could be enumerated, so proposed item names below are
provisional pending Lane A confirming the target vault. A second real gap: Hermes' own runtime
reads secrets from a profile-local `.env`/`auth.json` (evidence below), so whatever lands in
SOPS/1Password still needs a decrypt-to-local-file step on the VM at deploy time — this is not
optional plumbing, it's how Hermes profiles consume credentials at all.

## Facts with evidence

### Storage substrate (applies to all 8)

- `op` (1Password CLI) and `sops` both installed: `which op sops` → `/home/linuxbrew/.linuxbrew/bin/op`,
  `/usr/local/bin/sops` (v3.10.2).
- `op vault list` / `op whoami` → `[ERROR] You are not currently signed in.` `op account list` shows
  one account: shorthand `mobileclubgroup`, URL `https://mobileclubgroup.1password.eu`, email
  `gaetan.cathelain@mobile.club`. **Signin is interactive (password/2FA) — a Lane A step, not
  recon.** No vault name confirmed yet.
- mc-kestra's `.gitignore` (`~/dev/mc-kestra/.gitignore`) ignores `.env`/`.env.kestra` — i.e. even
  the prior art keeps plaintext off git, but it's still plaintext-on-disk, unencrypted at rest.
  PLAN.md explicitly rejects that pattern for Tars ("never `.env`").
- **Hermes profiles store secrets locally, not centrally** — evidence from a real Hermes
  profile-distribution build on this machine (`~/hermes-setup-prompt.md`, a different agent
  ("support-engineer") on a different VM, but same Hermes mechanics): "SECRETS (`.env` /
  `auth.json`) stay LOCAL and OUT of the repo," confirmed live in
  `~/hermes-aws-provisioning-handoff.md`: `/data/hermes/profiles/support-engineer/.env` (0600,
  owned by the service user, gitignored). **Implication for Tars:** SOPS/1Password is the at-rest
  canonical store; a deploy step must still decrypt/fetch into a profile-local `.env`/`auth.json`
  under Tars' Hermes profile dir on the VM — mirrors mc-kestra's own `env-secrets.sh` (decrypt →
  local file → restart consumer) rather than replacing that pattern.
- Proposed naming convention (provisional — vault TBD by Lane A):
  - 1Password items, one per row in the table below, named `Tars — <label>` in a vault (new
    `Tars` vault recommended over dumping into a personal catch-all — cheap to create, keeps
    revocation/rotation scoped).
  - SOPS: one age-encrypted YAML, `secrets/tars.sops.yaml` (new `secrets/` dir in this repo,
    `.sops.yaml` recipient rule pointing at the machine key(s) that need to decrypt — VM's own age
    key, cooper's if Lane A stages values there first). Keys reuse mc-kestra's own env-var names
    where they overlap (`GITHUB_PAT`, `LINEAR_API_KEY`, …) — zero-friction mapping from the
    walkthroughs already written.

### 1. GitHub (+ mc-metarepo clone access)

**Harvest** (browser, unchanged from mc-kestra): `https://github.com/settings/tokens` → Generate
new token (classic) → scopes `repo` + `read:org` → pick expiry → copy into storage.

**Confirmed fit for Tars**, evidence gathered live:
- `gh auth status` → already logged in as `GaetanCathelain`, token scopes
  `admin:org, gist, repo, workflow` (Gaetan's own gh CLI session — not Tars' credential, but
  proves the account/org relationship Tars' own PAT will need).
- `gh repo view mobile-club/metarepo --json visibility,name` → `{"name":"metarepo","visibility":"PRIVATE"}`
  — confirms mc-metarepo is a **private** repo under org `mobile-club`, reachable from this
  GitHub identity.
- `gh api user/orgs --jq '.[].login'` → `mobile-club`, `cleaq`, `Next-Mobiles` — matches
  mc-kestra's own owner scope list exactly (`.env.example`'s `vars.orgs` comment), so the same
  scopes (`repo` + `read:org`) that already work for mc-kestra's GITHUB_PAT will work for Tars'.

**Probe** (non-destructive, secret never on argv — mirrors mc-kestra's own `curl -K` pattern,
never `-H "Bearer $TOKEN"`):
```sh
printf 'header = "Authorization: Bearer %s"\n' "$GITHUB_PAT" > /tmp/gh.cfg
curl -s -K /tmp/gh.cfg https://api.github.com/repos/mobile-club/metarepo | jq .private
shred -u /tmp/gh.cfg   # or rm -f if shred unavailable
```
`true` back = token can read the metarepo. Clone probe (also read-only, safe): same technique
feeding a `GIT_ASKPASS` helper script instead of a URL-embedded token, `git ls-remote
https://github.com/mobile-club/metarepo.git HEAD`.

**Storage name:** `Tars — GitHub PAT` / SOPS key `GITHUB_PAT`.

**Gotchas (from prior art, unchanged):** never build a `https://<token>@github.com/...` URL —
that's the one path that persists a credential into a clone's `.git/config`. Classic PAT only;
fine-grained needs one token per owner (4 tokens vs 1). Expiry is your own choice at mint time —
**open question below**, since an unattended agent needs either a long expiry + calendar reminder
or a rotation job, and mc-kestra's docs don't resolve which.

### 2. Linear

**Harvest** (unchanged): `linear.app` → Settings → Security & access ("API") → Personal API keys
→ create → copy `lin_api_...`.

**Probe:**
```sh
printf 'header = "Authorization: %s"\nheader = "Content-Type: application/json"\n' "$LINEAR_API_KEY" > /tmp/lin.cfg
curl -s -K /tmp/lin.cfg -d '{"query":"{ viewer { id name } }"}' https://api.linear.app/graphql
shred -u /tmp/lin.cfg
```
**Bare key, no `Bearer ` prefix** — mc-kestra's own gotcha, proved live against a deliberately-bad
key (clean 401, not a malformed-header error). A 200 with `viewer.id` = healthy.

**Storage name:** `Tars — Linear API Key` / SOPS key `LINEAR_API_KEY`.

**Gotchas:** never expires on a schedule — dies only on manual revocation. Comments paginate at
250/issue (script-level limitation in mc-kestra's dumper, not relevant if Tars only reads live via
API, but worth knowing if Tars ever bulk-fetches issue history).

### 3. Notion

**Harvest** (browser devtools, unchanged): Quick Find (Ctrl/Cmd+P) → any query → inspect the
`search` XHR → request payload `spaceId` → `NOTION_SPACE_ID`; same request's Cookies →
`token_v2` → `NOTION_TOKEN_V2`; **`file_token` only appears after one manual export** (right-click
a page → Export, or Settings & members → Settings → Export content) — do that once first.

**Probe:** repeat the same `search` XHR shape server-side (cheapest live check, no export
triggered):
```sh
curl -s -b "token_v2=$NOTION_TOKEN_V2" -H "Content-Type: application/json" \
  -d '{"query":"","spaceId":"'"$NOTION_SPACE_ID"'"}' \
  https://www.notion.so/api/v3/search
```
A 200 with a results envelope (even empty) = `token_v2` + `spaceId` are live; a 401 = re-harvest.
`file_token` can't be probed this cheaply — it's only exercised by an actual export download, so
its probe is necessarily the first real `gbrain.notion/dump`-equivalent run, not a standalone
check (same limitation mc-kestra lives with).

**Storage name:** `Tars — Notion (token_v2 / file_token / space_id)` — one item, three fields, not
three items (they're always used together) / SOPS keys `NOTION_TOKEN_V2`, `NOTION_FILE_TOKEN`,
`NOTION_SPACE_ID`.

**Gotchas:** ~1 year lifetime, dies on a web logout (rare re-auth, same class as Slack's cookie).

### 4. Google Calendar

**Harvest: none — reuses the Gmail app password** (see #5). No OAuth client, no GCP project.

**Probe** (cheaper than mc-kestra's own full dump — PROPFIND instead of GET, no event body
pulled):
```sh
curl -s -u "$GMAIL_ADDRESS:$GMAIL_APP_PASSWORD" -X PROPFIND -H "Depth: 0" \
  "https://www.google.com/calendar/dav/$GMAIL_ADDRESS/events"
```
200 = credential authenticates CalDAV; 401 = re-mint the Gmail app password (check Gmail/IMAP
first — same credential, so if IMAP still works and only this fails, it's the endpoint-retirement
risk noted below, not a bad password).

**Storage name:** none of its own — rides on `Tars — Gmail App Password`.

**Gotchas (unchanged, and the sharpest one in the whole map):** the working endpoint
(`google.com/calendar/dav/.../events`) is Google's **undocumented legacy pre-2013 DAV host** — the
modern documented endpoint refuses Basic auth outright. Google can retire it without notice; the
symptom is a 401 here while Gmail IMAP keeps working. No mitigation exists short of an OAuth
Calendar-API-v3 rebuild (mc-kestra tried and retired it once CalDAV was proven live — the code
isn't there to fall back on). Primary calendar only (no `calendarList.list` over CalDAV). No
cursor — every run/probe is a full snapshot, so failure costs nothing to retry.

### 5. Gmail (via Himalaya)

**Harvest** (unchanged): `https://myaccount.google.com/apppasswords` — **requires 2FA already
enabled**, else the page 404s/redirects instead of refusing. Name it, copy the 16 chars (spaces
optional) into storage.

**What Himalaya specifically needs, beyond the raw credential** — checked and **not installed
anywhere on this machine** (`which himalaya` → not found; no binary, no `~/.config/himalaya/`, no
docker image tagged himalaya). From Himalaya's own config model (not locally verifiable without
installing, which is out of scope for read-only recon — flagged as an open question below):
a TOML config (`~/.config/himalaya/config.toml` by default) per account, with an `imap` backend
block (`host`, `port`, `login`, and an `auth` section — either a literal password or a shell
command that prints one) and a separate `smtp` block for sending. **The Gmail app password
already harvested for #4/#5 covers both** — Google app passwords are account-scoped, not
protocol-scoped (imap.gmail.com:993 / smtp.gmail.com:465, same 16 chars). Reuse avoids a second
harvest; the only decision is whether to reuse the literal same 1Password item or mint a second
app password labeled for Himalaya specifically (Google allows several per account) so Himalaya's
credential can be rotated/revoked independently of mc-kestra's. Lean toward the second (cheap,
and the two consumers should not share a blast radius) — flagged as open question.

**Probe** (works today without Himalaya installed, and doubles as the fallback probe if the CLI
syntax below turns out wrong):
```sh
curl -s --url imaps://imap.gmail.com --user "$GMAIL_ADDRESS:$GMAIL_APP_PASSWORD"
```
Lists mailboxes with no message fetch — cheapest possible IMAP auth check. Once Himalaya is
actually installed on the VM, the intended probe is `himalaya envelope list -a tars -p INBOX -s
1` (list the single latest envelope, no send) — **syntax unverified locally, flag for whoever
wires WF3's Gmail agent to confirm against `himalaya --help` post-install.**

**Storage name:** `Tars — Gmail App Password (Himalaya)` if minted separately, else this item
*is* the same one Calendar rides on / SOPS key `GMAIL_ADDRESS` + `GMAIL_APP_PASSWORD` (or
`GMAIL_ADDRESS` + `GMAIL_APP_PASSWORD_HIMALAYA` if split).

**Gotchas (unchanged from prior art):** displayed as 4 space-separated groups, spaces optional.
IMAP download cap ~2.5GB/day (irrelevant to Himalaya's live-read use, relevant if Tars ever bulk
syncs). Folder names localize — resolved via the IMAP `\All` RFC 6154 tag, nothing to hardcode.
Advanced Protection Program enrollment disables app passwords entirely (would force OAuth,
unverified whether the account has APP enabled).

### 6. Slack — personal account

**Settled by prior art**, this is the R2 verdict too: **`xoxc-` token + `xoxd-` cookie**, harvested
from a logged-in Slack **web** tab via devtools (Application → Local Storage →
`localConfig_v2` → `teams.<T…>.token` for the token; Application → Cookies → `slack.com` → `d`
cookie for the cookie). Confirmed as the right method for Tars' runtime too: it's a plain
HTTP-API credential, not tied to Kestra's containers — works identically wherever the HTTP
request originates, including a Hermes plugin on the new VM.

**Probe** (standard Slack self-check, cheaper than mc-kestra's own full-archive first run):
```sh
curl -s -b "d=$SLACK_COOKIE" -F "token=$SLACK_TOKEN" https://slack.com/api/auth.test
```
`{"ok":true,"user":"...","team":"..."}` = healthy; `{"ok":false,"error":"invalid_auth"}` =
re-harvest.

**Storage name:** `Tars — Slack Personal (xoxc/xoxd)` / SOPS keys `SLACK_TOKEN`, `SLACK_COOKIE`.

**Gotchas (unchanged):** cookie must be URL-encoded for raw API calls (`%2F` not `/`) — write the
storage/decrypt step to auto-encode a decoded paste, same as mc-kestra's `env-secrets.sh`, so
Lane A can paste either form. Practical lifetime ~5 years — dies on logout / password change /
admin-forced session reset, never on a schedule; treat any failure as "re-harvest."

### 7. Slack — Tars app tokens (held for cutover)

**No mc-kestra precedent** — this is Tars-specific. Key fact from `PITCH.md` line 54: **the Tars
Slack app already exists** ("I previously created a Tars Slack app, it's used by another profile
on my personal Hermes") — harvest is retrieving tokens for an **existing** app, not creating one.

**Harvest** (browser): `https://api.slack.com/apps` → select the Tars app → **OAuth & Permissions**
tab → Bot User OAuth Token (`xoxb-...`). If the app uses Socket Mode (typical for a
long-running bot with no public HTTP endpoint, which fits "new VM behind Tailscale, no inbound
webhook") — **Basic Information** → App-Level Tokens → an existing token or generate one scoped
`connections:write` (`xapp-...`). **Basic Information** → App Credentials → Signing Secret, needed
only if the integration verifies inbound HTTP requests rather than using Socket Mode exclusively.
**Exact transport (Socket Mode vs Events API) is Hermes' Slack plugin's decision, not
independently knowable from this app-console page — flag for R1 (Hermes deep-dive) / whoever
wires WF3's Slack-Tars agent to confirm which tokens Hermes' plugin config actually asks for.**
Checked locally for a shortcut: this machine's own `~/.hermes/plugins/` has only `orca-status`
installed (`find ~/.hermes/plugins -maxdepth 4` → one entry) — **no local Slack plugin to read
config shape from**, so the exact env-var names Hermes expects remain unconfirmed by this probe.

**Probe** (validates tokens without opening a persistent connection — no side effects):
```sh
curl -s -H "Authorization: Bearer $SLACK_BOT_TOKEN" https://slack.com/api/auth.test
curl -s -X POST -H "Authorization: Bearer $SLACK_APP_TOKEN" https://slack.com/api/apps.connections.open
```
Bot probe: same `auth.test` shape as #6. App-level probe: `apps.connections.open` returns a
one-time `wss://` URL on success — the probe should **not connect to it**, just confirm `"ok":true`
in the response, which alone proves the app token is live.

**Storage name:** `Tars — Slack App Tokens (bot + app + signing)` — held, not attached, until
cutover (per PLAN.md's explicit sequencing: harvested/probed like every other credential in Lane
A, but never wired into the new VM's running Hermes config until the gated cutover step deletes
the old profile). **Why held, inferred from PLAN.md's own graph** (not stated outright, but the
only reading that makes the sequencing make sense): Slack Socket Mode allows multiple concurrent
app-token connections consuming the same event stream — if the new VM's Hermes were wired live
while the old p-Hermes profile is still running, both could receive and act on the same Slack
events, producing duplicate/racing responses. Holding the tokens until the old profile is deleted
avoids that, rather than any technical inability to harvest them earlier.

**Gotchas:** none from prior art (no precedent); the two identified above are transport-shape
(Socket Mode vs Events API — confirm before wiring) and the held-not-attached sequencing (already
enforced by PLAN.md's graph, just noting *why*).

### 8. Orca / SSH mesh access for the VM

**No mc-kestra precedent, and the shape is different from every credential above** — not one
secret harvested from a website, but **infrastructure trust material** established at
provisioning time. Two distinct pieces:

**(a) Tailscale auth key** — needed for the new VM to join the tailnet non-interactively (PLAN.md
Lane B step B2, "tailscale join"). **Harvest** (browser): `https://login.tailscale.com/admin/settings/keys`
→ Generate auth key → choose reusable vs single-use, ephemeral vs persistent, tags, expiry (default
90 days) → copy `tskey-auth-...`. **Probe is inseparable from consumption** — unlike every other
credential here, an auth key (especially single-use) can't be tested without spending it; its real
"probe" is Lane B's own `tailscale up --authkey=$TSKEY` succeeding during provisioning, followed by
`tailscale ping tars` from cooper post-join. **Storage name:** `Tars — Tailscale Authkey` / SOPS
key `TAILSCALE_AUTHKEY` — but given it's typically single-use, storing it long-term at rest may
be pointless (it's spent at first use); worth deciding whether this even belongs in the
credential store vs. being minted fresh and piped directly into Lane B's provisioning step.

Evidence this machine's own tailnet identity does **not** currently see a Hermes/Proxmox host:
`tailscale status` → this device (`workstation-vm`, owner `gaetan.cathelain@`) lists only
work-shaped peers (`bastion-cleaq`, `bastion`, various `desktop-*`/`envirowipes-*` Windows/macOS
hosts) — **no host matching `hermes`/`proxmox`/`tars` anywhere in the list.** Either p-Hermes and
Proxmox live on a **separate, personal tailnet/account** not joined from cooper, or they're
tagged/hidden from this view. **This is a blocker for whoever runs Lane B's provisioning** (not
resolved here — squarely R3/R5's probes, flagged for the synthesis to reconcile): confirm which
tailnet account the new Tars VM should join before minting the auth key above.

**(b) SSH keypair for the mesh** (cooper ↔ macOS ↔ p-Hermes ↔ new VM, per PLAN.md B3) — not
harvested from anywhere, **generated locally**: `ssh-keygen -t ed25519 -f tars_vm_ed25519 -N ""`
on whichever side initiates trust (simplest: one keypair minted during VM provisioning, its
private half stored, public half pushed to `~/.ssh/authorized_keys` on the hosts that need to
reach the VM; the reverse direction — VM reaching out — needs its own key added to each target's
`authorized_keys`, but PLAN.md's guardrails keep the VM **read-only** on personal Hermes, so that
direction may not need a key at all, only an existing read-only mechanism). This machine's own
`~/.ssh/config` has exactly one host alias today (`mac`, a temp entry from 2026-07-30) — **no
existing entry for a Tars VM or p-Hermes**, confirming this mesh is being built from scratch, not
extended from something already there.

**Probe:** `ssh -i <key> -o BatchMode=yes -o ConnectTimeout=5 <host> true` — exit 0, no output,
fully read-only/no-op.

**Storage name:** `Tars — VM SSH keypair (mesh)` — private key as the item's file attachment (1Password
supports SSH-key item type natively) or SOPS-encrypted PEM block under key `TARS_VM_SSH_PRIVATE_KEY`.

**Possibly also relevant — Orca's own remote-pairing mechanism**, distinct from raw SSH: `orca
--help` lists `orca environment add --name <name> --pairing-code <code>` and `orca serve
--pairing-address <host>` — i.e. Orca has a built-in way to attach to a remote headless runtime via
a short-lived pairing code generated by `orca serve` on the target host, rather than (or alongside)
raw SSH. **Whether Tars/lane dispatch should use this instead of/in addition to SSH is R4's call
(Orca control node) — flagged here only because it's credential-shaped** (a pairing code is a
short-lived secret) and would need its own storage entry (`Tars — Orca Pairing Code`) if adopted,
though its short lifetime likely makes long-term storage moot, same caveat as the Tailscale
auth key above.

## Blockers

- `op` CLI not signed in on this machine — Lane A must `op signin` interactively (password + 2FA)
  before any item can be created; no vault enumerated yet, so item-vault placement is unconfirmed.
- Himalaya is not installed anywhere reachable from this machine (no binary, no config, no docker
  image) — its exact CLI probe syntax (`himalaya envelope list ...`) is asserted from general
  knowledge of the tool, not verified against a running instance. Falls back cleanly to the raw
  `curl imaps://` probe, which needs no Himalaya install.
- This machine's tailnet does not show a Hermes/Proxmox host — the Tailscale auth key and SSH mesh
  targets (item 8) can't be fully pinned down until R3/R5 settle which tailnet/account the new VM
  joins. Flagged for synthesis, not resolved here.
- Slack Tars app's exact token requirements (Socket Mode vs Events API, whether a signing secret
  is needed at all) depend on Hermes' Slack plugin's own config contract, which isn't installed
  locally to inspect. Flagged for R1 / WF3's Slack-Tars wiring agent.

## Open questions

1. **1Password vault for Tars items** — new dedicated `Tars` vault, or an existing personal vault?
   (Needs `op signin` first to even list options.)
2. **GitHub PAT expiry** — pick a date now (needs a rotation reminder) or mint no-expiry classic
   PAT (simpler, but a standing risk mc-kestra's own docs don't take a position on)?
3. **Himalaya's Gmail app password** — reuse the exact same app password already minted for
   mc-kestra's Gmail dump + Calendar CalDAV, or mint a second one labeled for Tars/Himalaya so its
   blast radius is independent? (Recommend the second — cheap, and Google allows multiple app
   passwords per account.)
4. **Slack Tars app transport** — Socket Mode (`xapp-` app-level token) vs Events API (public
   endpoint + signing secret)? Given the VM sits behind Tailscale with no public inbound path
   planned elsewhere in PLAN.md, Socket Mode looks like the natural fit, but this should be
   confirmed against Hermes' actual Slack plugin config rather than assumed here.
5. **Tailscale account/tailnet** — does the new Tars VM join the same tailnet as this machine
   (`workstation-vm`, work-shaped) or a separate personal one where p-Hermes/Proxmox actually live?
   This machine's own `tailscale status` shows no evidence of the latter being reachable from here.
6. **Orca pairing vs raw SSH** for the mesh — does Lane B wire raw SSH keys, Orca's own
   `environment add --pairing-code` mechanism, or both? Affects whether a 9th storage item
   (`Tars — Orca Pairing Code`) is needed.
