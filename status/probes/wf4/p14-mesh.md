# WF4 Probe 14 — SSH mesh, all legs + tailscale capture

Spec: docs/specs/wf4-probes.md §14. Superseded per WF4 dispatch instructions: the
direct `phermes` leg (VM 103 = 192.168.0.8) replaces the spec's stale
pve-root+`qm guest exec` path — 192.168.0.3 root was NOT touched.

All timestamps `date -Is`, run 2026-08-07 in the ~19:00Z window (post-cutover).

## Leg 1 — cooper -> VM (192.168.0.9)

```
$ date -Is
2026-08-07T19:00:09+00:00
$ ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 true; echo "exit=$?"
exit=0
```

**Verdict: PASS** (exit 0, no output, as spec'd).

## Leg 2 — VM -> cooper (tailnet/LAN alias `cooper`)

Run remotely on the VM via nested ssh:

```
$ ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 \
    "date -Is; ssh -o BatchMode=yes -o ConnectTimeout=5 cooper true; echo exit=\$?"
2026-08-07T19:00:12+00:00
exit=0
```

VM's `~/.ssh/config` entry for `cooper`:
```
Host cooper
    HostName 192.168.0.4
    User gaetan
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
```

**Verdict: PASS**.

## Leg 3 — VM -> macOS box (tailnet alias `mac`)

```
$ ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 \
    "date -Is; ssh -o BatchMode=yes -o ConnectTimeout=5 mac true 2>&1; echo exit=\$?"
2026-08-07T19:00:15+00:00
exit=0
```

VM's `~/.ssh/config` entry for `mac`:
```
Host mac
    HostName macbook-pro-de-gaetan.tail6e788b.ts.net
    User gcath
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
```

Mac was awake/reachable — exit 0, no deviation needed.

**Verdict: PASS**.

## Leg 4 — VM -> p-Hermes (direct `phermes` leg, VM 103 = 192.168.0.8)

Per dispatch: the spec's pve-root(192.168.0.3)+`qm guest exec 103` path is
SUPERSEDED by the direct `phermes` alias wired at cutover. 192.168.0.3 root was
NOT touched in this probe.

```
$ ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 \
    "date -Is; ssh -o BatchMode=yes -o ConnectTimeout=5 phermes true 2>&1; echo exit=\$?; grep -A5 '^Host phermes' ~/.ssh/config"
2026-08-07T19:00:19+00:00
exit=0
Host phermes
    HostName 192.168.0.8
    User hermes
    IdentityFile ~/.ssh/id_ed25519
    BatchMode yes
    ConnectTimeout 5
```

**Verdict: PASS**.

## Tailscale capture (proactive, completeness-critic requirement)

Run from cooper.

```
$ date -Is
2026-08-07T19:00:22+00:00
$ tailscale ping tars
pong from tars (100.116.31.76) via 192.168.0.9:41641 in 1ms
$ ping -c1 100.116.31.76
PING 100.116.31.76 (100.116.31.76) 56(84) bytes of data.
64 bytes from 100.116.31.76: icmp_seq=1 ttl=64 time=2.72 ms

--- 100.116.31.76 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 2.722/2.722/2.722/0.000 ms
```

`tailscale ping tars` resolved via direct UDP path (`192.168.0.9:41641`, LAN
short-circuit of the tailnet — expected on same subnet) and ICMP ping both
succeeded. Satisfies the completeness critic's D5 gap-check item ("Tailscale
itself has no dedicated probe: flag if `tailscale ping tars` from cooper was
never captured anywhere") — it is now captured here.

**Verdict: PASS**.

## Summary

| Leg | Verdict | FAIL owner (if not PASS) |
|---|---|---|
| cooper -> VM | PASS | n/a |
| VM -> cooper | PASS | n/a |
| VM -> mac (tailnet alias) | PASS | n/a |
| VM -> phermes (direct leg) | PASS | n/a |
| tailscale ping tars + ICMP ping (cooper) | PASS | n/a |

No 192.168.0.3 (pve root) contact made — direct phermes leg used per
superseding instruction. No units restarted/stopped/enabled/disabled. No
config/.env edited. No sops invoked. No secrets echoed.
