# Tailscale join — evidence

Closes lane B's deferred B2 item (`status/lane-b.md` B2: "**PASS** except tailscale
(awaiting lane-A key)"). Additive per D4/PLAN.md Amendments — LAN (`192.168.0.9`) stays
primary; tailnet is a second path, not a replacement. No key material appears below or at
any point in this transcript.

## Steps run (VM: `gaetan@192.168.0.9`)

1. **Install** — official `install.sh` piped through `sudo sh`:
   `Tailscale 1.102.2` installed from `pkgs.tailscale.com/stable/ubuntu noble`;
   `tailscaled.service` enabled via symlink at install time.
2. **Key transfer** — `cat /home/gaetan/.tars-tskey | ssh gaetan@192.168.0.9 'umask 077; cat >
   /tmp/.tskey'`. Never touched argv or any log; landed `0600`, 61 bytes (matches source).
3. **Flag check** — `tailscale up --help` on the VM confirmed `--auth-key value` accepts a
   `file:`-prefixed path (current binary, not the older `--authkey`).
4. **Join** — `sudo tailscale up --auth-key file:/tmp/.tskey --hostname tars`. Exit 0. Only
   output was a benign health-check note (`Some peers are advertising routes but
   --accept-routes is false`) — expected, unrelated to this join.
5. **Key destroyed both sides, unconditionally:**
   - VM: `shred -u /tmp/.tskey` → exit 0, file confirmed gone.
   - cooper: `shred -u /home/gaetan/.tars-tskey` → exit 0, file confirmed gone.
   - Auth key was single-use and pre-authorized — spent by the `up` call either way, so both
     shreds ran regardless of step 4's outcome (it succeeded).

## Verification

| Check | Command | Result |
|---|---|---|
| Node identity | `tailscale status --json` on VM | `HostName: tars`, `DNSName: tars.tail6e788b.ts.net.`, `TailscaleIPs: 100.116.31.76` (+ IPv6 `fd7a:115c:a1e0::9139:1f4d`), `Online: true`, owner `gaetan.cathelain@` |
| Tailnet | same | `MagicDNSSuffix: tail6e788b.ts.net` — lands on the existing mobile.club org tailnet (same one `workstation-vm`, `bastion`, `mac`, etc. are already on), not a fresh/isolated net |
| Reachability | `tailscale ping tars` from cooper | `pong from tars (100.116.31.76) via 192.168.0.9:41641 in 1ms` — direct path over LAN, mesh healthy |
| Gateway untouched | `systemctl --user is-enabled hermes-gateway.service` on VM | `disabled` (exit 1); `is-active` → `inactive` (exit 3) — unchanged from B4, not started/enabled by this probe |
| `~/.hermes` | not touched | no command in this probe read, wrote, or referenced any path under `~/.hermes` |

## Verdict

**PASS.**

- Node: **`tars`**
- Tailscale IP: **`100.116.31.76`** (DNS: `tars.tail6e788b.ts.net`)
- Tailnet: **`tail6e788b.ts.net`** (mobile.club org tailnet, shared with `workstation-vm`,
  `bastion`, `mac`, `macbook-pro-de-gaetan`, etc.)
- Gateway: confirmed `disabled`/`inactive`, not touched.
- `~/.hermes`: not touched.
- No key material printed, echoed, or logged at any point.
