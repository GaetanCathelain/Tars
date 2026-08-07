# status — lane B (WF2 provision)

Single writer: the lane-B session. Verdicts only, newest section last.
Spec: `PLAN.md` §Amendments + `docs/recon/DECISION.md` D2 / D4 / D6.

## Chain state

| Step | What | State |
|---|---|---|
| B1 | Proxmox VM 102 `tars` — 8c / 8G / 50G, cloud-image + cloud-init, `local-lvm`, `vmbr0` | in progress |
| B2 | Ubuntu 24.04 + minimal X + **Docker** (D6-5) + tailscale (additive, D4) | pending |
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
