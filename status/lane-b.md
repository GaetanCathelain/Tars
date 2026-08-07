# status — lane B (WF2 provision)

Single writer: the lane-B session. Verdicts only, newest section last.
Spec: `PLAN.md` §Amendments + `docs/recon/DECISION.md` D2 / D4 / D6.

## Chain state

| Step | What | State |
|---|---|---|
| B1 | Proxmox VM 102 `tars` — 8c / 8G / 50G, cloud-image + cloud-init, `local-lvm`, `vmbr0` | **PASS** |
| B2 | Ubuntu 24.04 + minimal X + **Docker** (D6-5) + tailscale (additive, D4) | **PASS** except tailscale (awaiting lane-A key) |
| B3 | SSH mesh: cooper · macOS · p-Hermes (read-only) | pending |
| B4 | Hermes install — base profile `~/.hermes` **is** Tars (D2/D6-4) | pending |
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
