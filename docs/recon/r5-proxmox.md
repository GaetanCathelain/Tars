# R5 — Proxmox probe: can we create the Tars VM?

Target: 8 cores, 8G RAM, 50G disk, Ubuntu + GUI, tailscale. Read-only probes only,
all via `ssh root@192.168.0.3` (host `pve`) — no `qm create`/`qm set`/any mutation run.

## Verdict

**Capacity: YES**, with comfortable RAM/CPU headroom and workable-but-watch-it disk
headroom on the thin pool. No old Tars VM exists (VMIDs in use: 100, 101, 103 only;
`pvesh get /cluster/nextid` returns **102** — free and ready to use).

**No VM template exists to clone from** (all 3 VMs have `template=None`) — template-clone
is **not available** as a mechanism today. Two real lanes exist:

1. **Cloud image + cloud-init (recommended)** — `local:iso/noble-cloudimg-amd64.img`
   (Ubuntu 24.04 "Noble" generic cloud image) imported as the VM disk on `local-lvm`,
   cloud-init drive (`ide2`, `local-lvm:vm-XXX-cloudinit`) for SSH keys/network/user
   config, customized via a `snippets`-stored user-data YAML (`cicustom`). **This is the
   exact, proven mechanism already in production for personal Hermes (VM 103)** —
   `/var/lib/vz/snippets/hermes-user.yaml` exists on this host and VM 103's config
   shows `ide2: local-lvm:vm-103-cloudinit,media=cdrom`. It ships headless/server only —
   GUI must be added post-boot (`ubuntu-desktop-minimal`/`xubuntu-desktop` + a remote
   display server such as xrdp, or just Xorg/virtio-gpu if only a local X session for
   CloakBrowser-style headed-browser automation is needed) via cloud-init `packages`/
   `runcmd`. Fully unattended and scriptable — no console interaction needed.
2. **ISO install (fallback)** — `local:iso/ubuntu-24.04.4-desktop-amd64.iso` is present
   (also 22.04 desktop and 22.04 live-server ISOs). Gives a native GUI desktop
   out of the box, but needs either manual installer interaction over VNC/SPICE console
   or a subiquity `autoinstall` NoCloud seed to be unattended — **zero precedent on this
   host** that the autoinstall path works, unlike the cloud-init lane which is proven.

**Recommendation for lane B:** follow the hermes pattern — cloud image import +
cloud-init snippet, install a minimal desktop/X stack and tailscale via cloud-init
`packages`/`runcmd`, target VMID 102, storage `local-lvm`, bridge `vmbr0`.

## Facts with evidence

### Host identity and prior art
- Proxmox host = `pve`, reached via `ssh root@192.168.0.3` (documented in
  `~/dev/gaetan-metarepo/machine/migration-decisions.md` as `MIRROR_HOST_SSH`).
  `pveversion` → `pve-manager/8.4.19/a68fb383814bb1e6 (running kernel: 6.8.12-20-pve)`.
- Single node, **not clustered**: `pvecm status` → `Error: Corosync config
  '/etc/pve/corosync.conf' does not exist - is this node part of a cluster?`.
  `pvesh get /nodes` lists exactly one node, `pve`.
- Personal Hermes is **VM 103** on this same host (`learnings/hermes-vm.md`: "personal
  Hermes (Proxmox VM 103)") — this probe treated the whole host as the p-Hermes
  read-only boundary per the task guardrails.

### Existing VMs (no old Tars VM)
```
$ qm list
      VMID NAME                 STATUS     MEM(MB)    BOOTDISK(GB)
       100 workstation-01       running    18432            120.00
       101 homeassistant        running    4096              32.00
       103 hermes               running    8192             100.00
```
`pvesh get /nodes/pve/qemu` confirms `template=None` for all three — **none is a
template**, so `qm clone` from a golden image is not an option today.
`pvesh get /cluster/nextid` → `102` (the gap is free, not a lingering old Tars VM —
no VM named "tars"/"Tars" exists anywhere in the list).

### CPU capacity
- `pvesh get /nodes/pve/status`: `cpuinfo.cores=12, cpus=24`, model AMD Ryzen 9 3900X.
  `loadavg: ["2.05","1.86","1.73"]` — ~8.6% utilization of 24 threads, host is idle.
- Configured vCPUs already allocated: workstation-01=12, homeassistant=2, hermes=8 →
  **22 vCPU already configured** on a 24-thread host. Adding Tars' 8 cores → 30
  configured vCPUs vs 24 hardware threads (125% overcommit on paper). Not a hard
  blocker (KVM oversubscription is routine and current load is near-zero), but flag:
  if workstation-01 (a 12-core interactive VM) and Tars are both CPU-bound at once,
  contention is possible.

### RAM capacity
- `free -h` on host: `total 62Gi, used 37Gi, free 931Mi, available 25Gi`.
- Currently committed to running VMs: 18432+4096+8192 = 30720 MB ≈ **30 GiB** already
  allocated to guests; host overhead (services, buff/cache reclaimable) accounts for
  the rest of the 37Gi "used".
- **8G more for Tars fits inside the 25Gi "available"** with ~17GiB headroom left over.
  Not a blocker.

### Disk capacity
- Storage config (`/etc/pve/storage.cfg`): `local` (dir) holds
  `content snippets,backup,vztmpl,iso` — **no `images` content type**, so VM disks
  cannot live here. `local-lvm` (lvmthin, vgname `pve`, thinpool `data`) holds
  `content rootdir,images` — **this is where the new VM disk must go**, same as all
  three existing VMs.
- `pvesm status`:
  ```
  Name         Type     Total(KiB)   Used(KiB)   Available(KiB)   %
  local        dir      98497780     60146952    33301280         61.06%
  local-lvm    lvmthin  379953152    314031280   65921871          82.65%
  ```
  local-lvm ≈ 362.35 GiB total, ≈ 299.5 GiB used, **≈ 62.9 GiB available**. A 50G
  virtual disk (thin-provisioned, so actual physical consumption tracks real writes —
  a fresh Ubuntu+desktop install typically uses 10-20 GiB) fits comfortably today.
  **Watch item:** the pool is already at 82.65% nominal allocation; it's not a
  blocker for this one VM but there isn't a lot of slack left in the shared thin pool
  for future growth — flag to whoever plans the next VM after Tars.
  `lvs` confirms: `data` LV (the thin pool) `362.35g`, `Data% 82.65`.

### Cloud-init support — proven, in production use
- VM 103 (hermes) config includes `ide2: local-lvm:vm-103-cloudinit,media=cdrom` — a
  live cloud-init drive on the exact storage Tars would use.
- `/var/lib/vz/snippets/hermes-user.yaml` exists (3554 bytes) — a custom cloud-init
  user-data snippet, matching `local`'s `content snippets,...` capability. This is the
  real mechanism used to provision hermes: cloud image + custom user-data via
  `cicustom`. **Not read** (may carry SSH keys/host config) — existence and use
  confirmed by directory listing and VM103's `ide2` line only, no content probed.
- Conclusion: cloud-init is not just theoretically available on this Proxmox — it is
  the exact working pattern already used for the sibling VM on this host.

### Templates / ISOs available
```
$ pvesm list local --content iso
local:iso/noble-cloudimg-amd64.img              627,923,968 B   (Ubuntu 24.04 cloud image)
local:iso/ubuntu-22.04.1-desktop-amd64.iso     3,826,831,360 B
local:iso/ubuntu-22.04.1-live-server-amd64.iso 1,474,873,344 B
local:iso/ubuntu-24.04.4-desktop-amd64.iso     6,655,619,072 B
```
`pvesm list local --content vztmpl` → empty (no LXC container templates; irrelevant —
Tars needs a VM not a container). `/var/lib/vz/template/cache/` is empty too — no
prebuilt VM templates cached anywhere on this host.

### Network / bridge
- Single bridge `vmbr0`, static `192.168.0.3/24`, gateway `192.168.0.1`, bridged to
  physical `enp36s0` (`/etc/network/interfaces` + `ip -br a`). All three existing VMs'
  `net0` use `bridge=vmbr0` (virtio). **Use `vmbr0`** for Tars — it's the only bridge
  on the host and puts Tars on the same `192.168.0.0/24` LAN as everything else
  (Pi-hole DHCP serves `.20`–`.254` per `learnings/hermes-vm.md`; fixed addresses go
  below `.20` if wanted, though tailscale will likely be the addressing plan used
  operationally per PLAN.md B2).

### Guest agent
- VM 103's config has `agent: enabled=1` — qemu-guest-agent is standard practice on
  this host and is how the personal-Hermes memory-store probe (`qm guest exec 103`)
  was done for `learnings/hermes-vm.md`. Tars should get the same for the same reason
  (scriptable in-guest commands from the host without needing SSH to be up first).

### Task history
- `pvesh get /nodes/pve/tasks` was checked for prior `qmcreate`/`qmclone`/`qmimport`
  entries to see exactly how VM103 was originally provisioned — **none found**
  (task log retention doesn't reach back that far). Not a blocker; the cloud-init
  snippet + `ide2` cloudinit drive on VM103's live config is sufficient evidence of
  the mechanism.

## Blockers

None found for capacity or mechanism. The only real constraint is the shared
local-lvm thin pool's finite headroom (≈63GiB free, already 82.65% nominally
allocated) — enough for this one 50G VM, worth a mention to whoever provisions
the *next* VM on this host.

## Open questions

- Does "Ubuntu + GUI" mean a real interactive desktop (RDP/VNC-reachable) or just an
  X/Wayland display so a headed browser (CloakBrowser/Chromium, per the hermes
  pattern) has somewhere to render? Changes whether xrdp is needed or a bare
  Xorg+virtio-gpu session suffices. Not resolvable from the Proxmox side — PLAN.md's
  B2/B4 steps own this, this probe only confirms the host can support either.
- `hermes-user.yaml`'s exact content (packages, runcmd, network mode) wasn't read
  (avoided touching a file that may carry SSH keys). If lane B wants to literally
  clone hermes' cloud-init pattern, that snippet is the reference to read (on the
  Proxmox host, read-only) at execution time — not blocking this recon verdict.
- Tailscale join is a within-guest step (B2 in PLAN.md), orthogonal to Proxmox
  capacity/mechanism — not probed here.
