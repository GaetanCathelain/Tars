# status — lane B (WF2 provision)

Single writer: the lane-B session. Verdicts only, newest section last.
Spec: `PLAN.md` §Amendments + `docs/recon/DECISION.md` D2 / D4 / D6.

## Chain state

| Step | What | State |
|---|---|---|
| B1 | Proxmox VM 102 `tars` — 8c / 8G / 50G, cloud-image + cloud-init, `local-lvm`, `vmbr0` | **PASS** |
| B2 | Ubuntu 24.04 + minimal X + **Docker** (D6-5) + tailscale (additive, D4) | **PASS** except tailscale (awaiting lane-A key) |
| B3 | SSH mesh: cooper · macOS · p-Hermes (read-only) | **PASS** |
| B4 | Hermes install — base profile `~/.hermes` **is** Tars (D2/D6-4) | **PASS** |
| B5 | plugins ∥: rtk · hindsight · hermes-lcm · i-have-adhd | pending |
| B6 | CloakBrowser | pending |
| B7 | smoke check → verdict | pending |

Milestone pings to the hub: VM created · Hermes installed · smoke pass. Blockers ping immediately.

## 2026-08-07 — session start

**Hub contact established.** Introduction received; ack sent. Ping target recorded.

**Pre-flight probe, live (evidence):**

```
$ ssh root@192.168.0.3 'pveversion; pvesh get /cluster/nextid; qm list'
pve-manager/8.4.19/a68fb383814bb1e6 (running kernel: 6.8.12-20-pve)
102
 VMID NAME            STATUS   MEM(MB)  BOOTDISK(GB)
  100 workstation-01  running    18432        120.00
  101 homeassistant   running     4096         32.00
  103 hermes          running     8192        100.00
```

Verdict: pve reachable from cooper on key auth; **VMID 102 still free** (D4 holds, no drift);
no pre-existing Tars VM anywhere.

**Open item raised to hub, not blocking:** D5 lists a tailscale auth key as a lane-B input
("mint fresh, pipe into lane B") but minting needs the browser, which is lane A's. D4 makes
tailscale additive and LAN primary, so B1–B7 proceed on `vmbr0`/LAN. If no key arrives before
B7, tailscale is recorded as a deviation rather than a failure.

## 2026-08-07 — B1 + most of B2: **PASS** (milestone: VM created)

**VM 102 `tars` is up.** `192.168.0.9/24` on `vmbr0`, Ubuntu 24.04.4 LTS, `systemctl
is-system-running` → `running`, cloud-init `status: done`, `errors: []`. Cloud-init finished
~100 s after `qm start`.

Config as built (matches D4 exactly):

```
cores 8 · sockets 1 · cpu host · memory 8192 · balloon 4096 · machine q35 · ostype l26
agent enabled=1 · scsihw virtio-scsi-single · onboot 1 · serial0 socket · vga std
scsi0  local-lvm:vm-102-disk-0,discard=on,iothread=1,size=50G,ssd=1
ide2   local-lvm:vm-102-cloudinit,media=cdrom
net0   virtio=BC:24:11:76:E0:F1,bridge=vmbr0
cicustom user=local:snippets/tars-user.yaml · ipconfig0 ip=192.168.0.9/24,gw=192.168.0.1
```

**Mechanism note:** PVE 8.4.19 has dropped `qm importdisk`; the image was imported with
`qm set 102 --scsi0 local-lvm:0,import-from=…,format=raw` and grown 3.5G → 50G. Guest rootfs
auto-expanded: `/dev/sda1 48G, 3.3G used, 45G free`.

**B2 content verified on the VM (evidence):**

| Want | Got |
|---|---|
| Docker (D6-5, for the Slack MCP server) | `Docker version 29.1.3`, `docker compose version 2.40.3`, `gaetan` in group `docker` |
| minimal X (D4) | `/usr/bin/Xvfb`, `/usr/bin/openbox`, `xserver-xorg` — no xrdp/VNC, per D4 |
| guest agent | `qemu-guest-agent` enabled, `qm guest exec 102` works |
| gateway survives logout | `loginctl show-user gaetan` → `Linger=yes` |
| SSH from cooper | `ssh gaetan@192.168.0.9` → exit 0, key auth, `ssh_pwauth: false` |
| tooling | git · curl · jq |

**Decisions taken during B1, worth knowing:**

1. **Cloud-init user-data was assembled on the pve host with cooper's public key piped in**,
   never rendered into a transcript or this repo — `~/.ssh/` is on the global never-print list.
   The snippet lives at `/var/lib/vz/snippets/tars-user.yaml` (host-only, never copied here).
2. **Fixed an ordering bug before first boot:** the drafted user-data put `gaetan` in
   `groups: [sudo, docker]`, but cloud-init creates users in the *init* stage — before
   `packages:` installs Docker, so the `docker` group does not exist yet. Moved to
   `usermod -aG docker gaetan` in `runcmd` (final stage). Verified: `groups=…,112(docker)`.
3. **Added `docker-compose-v2`** — D1 calls for `docker compose up` of the Slack MCP server and
   Ubuntu's `docker.io` package does not ship the compose v2 plugin.
4. **Added `xvfb`** as cheap insurance for headless rendering (see the B6 correction below).
5. **Corrected a recon miscall:** B1 recon reported the pve host had "only 2.0Gi truly free RAM"
   and was at OOM risk. That read the `free` column. The host has **64 GiB total and ~25 GiB
   *available***; D4's capacity finding stands. Not a blocker, and no VM was resized for it.

**Watch item, not a failure — ballooning.** `info balloon` on VM 102 reports
`actual=6231 max_mem=8192 total_mem=5978`: the guest currently sees ~5.9 G, not 8 G, because
`balloon: 4096` lets the host reclaim under pressure (host is at ~39 G used / 10 G swap).
This is **byte-identical to VM 103's config** (`memory: 8192`, `balloon: 4096`) — the live
p-Hermes box that already runs Hermes fine — so it is the working precedent, not a deviation.
If Hermes turns out to be memory-starved at B7, the one-line fix is `qm set 102 --balloon 0`.

**Thin pool** after adding the 50 G thin volume: sum of thin volumes now 496 G against a 465 G
VG. Thin-provisioned and fine today (real usage 3.3 G), but the host has **no autoextend
threshold set** — flagged for Gaetan, outside lane B's mandate to change.

## 2026-08-07 — B3 SSH mesh: **PASS** (all five legs, independently re-verified)

One ed25519 keypair generated **on the Tars VM** (`tars@192.168.0.9`, fingerprint
`SHA256:5shawRKsUq5JzYXazx5shAAhwBXyRGzc2euLaQv3FYk`). Public half appended — one line, grep-guarded
for idempotency — to `authorized_keys` on cooper, pve and the Mac. No private key ever left the VM;
no existing `authorized_keys` line was rewritten or removed.

Verdicts re-run by this session directly, not taken from the agent's report:

| Leg | Direction | Probe | Result |
|---|---|---|---|
| Tars → cooper | `ssh cooper true` | `-o BatchMode=yes -o ConnectTimeout=5` | exit 0 **PASS** |
| Tars → pve | `ssh pve true` | idem | exit 0 **PASS** |
| Tars → macOS | `ssh mac true` | idem | exit 0 **PASS** |
| Tars → p-Hermes | `ssh pve 'qm guest exec 103 -- /bin/true'` | JSON `exitcode` | `0` **PASS** |
| cooper → Tars | `ssh gaetan@192.168.0.9 true` | idem | exit 0 **PASS** (unbroken) |

On the VM: `~/.ssh` 700, `id_ed25519` / `known_hosts` / `config` all 600; `known_hosts` pre-seeded
via `ssh-keyscan` (9 lines = 3 key types × 3 hosts) so **no leg can ever hit an interactive
fingerprint prompt**; `~/.ssh/config` carries three aliases — `cooper` / `pve` / `mac`.

**p-Hermes remains untouched.** The only thing sent through `qm guest exec 103` was `/bin/true`.
There is no SSH key on p-Hermes and none was installed — that leg is `root@pve` → `qm guest exec`
by design (VM 103 refuses direct SSH), which is why lane B's read-only constraint holds structurally
rather than by good behaviour.

**Two durability notes, neither blocking:**

1. **The Mac's TEMP key was still valid** — its first probe failed with `Host key verification
   failed`, which is a `known_hosts` gap on cooper, *not* revocation. Worth knowing because the
   failure mode looks identical to a revoked key at a glance. Now moot either way: the Tars VM has
   its **own** key on the Mac, so revoking cooper's temp key no longer breaks this leg.
2. **The Mac is at `192.168.0.23` by DHCP and has already drifted once** (cooper's `~/.ssh/config`
   still points at a dead `192.168.0.34`). The stable fix is the tailnet address, which is gated on
   the auth key deferred to WF3. Until then this leg is LAN-DHCP-fragile — if it breaks later, the
   cause is almost certainly a new lease, not a credential.

## 2026-08-07 — B4 Hermes install + base profile: **PASS** (milestone: Hermes installed)

**`Hermes Agent v0.20.0 (2026.8.3)`** on Python 3.11.15, installed unattended via the vendor
`install.sh` (exit 0; wizard auto-skipped, no TTY). Not pinned to p-Hermes' v0.19.0, per D2.

Re-verified by this session directly on the VM:

| D2 requirement | Observed |
|---|---|
| base profile IS Tars, no `hermes profile create` | `~/.hermes/profiles/` **absent**, `HERMES_HOME` **unset** |
| identity in `SOUL.md` | `diff` vs `docs/specs/tars-profile.md` §1 → **IDENTICAL**, byte-for-byte |
| Slack guardrails | `require_mention: true`, `strict_mention: true`, `unauthorized_dm_behavior: ignore` |
| A2A inbound-only, port 9900 | `gateway.platforms.a2a` → `enabled: true`, `extra.port: 9900` |
| `.env` 0600, empty at B4 | `600`, **0 bytes** |
| no credentials touched | `~/.hermes/auth.json` **absent**; no OAuth flow run |
| unit created, **not** enabled, **not** started | `is-enabled` → `disabled`, `is-active` → `inactive`, no enable symlink, no `override.conf` |

### The A2A config-path question is settled: **use `gateway.platforms.a2a`**

D2 said nested, live v0.19.0 looked top-level, and DECISION marked it UNCONFIRMED. The answer is
that it was never a contradiction — **both paths parse, and top-level silently wins a conflict.**
Established four independent ways: the v0.20.0 loader source (`gateway/config.py`) merges
`gateway.platforms` first, then top-level `platforms`, with a comment stating top-level keeps
precedence; an A/B/C experiment through the real loader in throwaway `HERMES_HOME`s (nested→8888,
top-level→7777, both-and-conflicting→**top-level won**); the live profile resolving correctly; and
the docs shipped inside v0.20.0 using the nested form.

**Consequence worth carrying into WF3/WF4:** the nested form is the documented one and is what is
written here, but a stray top-level `platforms:` block would override it *silently, with no error*.
That is a real footgun for the cutover — if a guardrail ever appears not to apply, check for a
top-level `platforms:` before debugging anything else.

Same mechanism settles the spec's open Slack-nesting question. Related find: **`strict_mention`
never appears in `PlatformConfig.extra`** — the Slack plugin exports it as the env var
`SLACK_STRICT_MENTION` (verified `=true`). Anyone grepping the parsed config object for
`strict_mention` will wrongly conclude the guardrail is missing.

### Deviations and corrections

1. **`config.yaml` was appended to, not replaced.** The installer ships a ~92 KB default with 21
   sections of real defaults; a pristine copy is kept at `~/.hermes/config.yaml.installer-default`.
2. **`model.*` is NOT absent — it holds installer defaults**, which is a stronger statement than
   "lane B didn't write it". Live right now: `provider: auto`, `default: anthropic/claude-opus-4.6`,
   `base_url: https://openrouter.ai/api/v1`. **WF3 must OVERRIDE these three keys, not add them** —
   a WF3 step written as "add the model block" would leave OpenRouter/Opus-4.6 in place and Tars
   would run on the wrong backend (or fail, since no OpenRouter key exists). Lane B deliberately
   left them untouched per the profile spec's §4 table.
3. **`[DOC]` facts checked: 12; wrong: 1.** The runbook's `hermes gateway install` flag list was
   incomplete — `--no-start-now` and `--no-start-on-login` also exist, which is precisely what let
   the "create but do not enable or start" constraint be satisfied cleanly rather than by
   post-hoc `systemctl disable`.
4. **p-Hermes was never contacted during B4.** The v0.20.0 tree ships its own source and docs on the
   new VM, which is better evidence than reading a v0.19.0 config off the live box. The read-only
   constraint held by not needing to read at all.
