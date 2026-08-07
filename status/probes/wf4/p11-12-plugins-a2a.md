# WF4 Probes 11-12 — plugins provisioning + A2A agent card

Run window: 2026-08-07T19:00Z–19:01Z, from cooper via `ssh gaetan@192.168.0.9`.

---

## Probe 11 — plugins loaded (rtk, hindsight, hermes-lcm, i-have-adhd)

### rtk — FAIL (lane B, B5)

```
$ date -Is
2026-08-07T19:00:12+00:00

$ ls -la ~/.local/bin ~/.cargo/bin /usr/local/bin | grep -i rtk
-rwxr-xr-x 1 gaetan gaetan 10326432 Aug  7 09:44 rtk      # only in ~/.local/bin

$ ls -la ~/.config/rtk/config.toml
ls: cannot access '/home/gaetan/.config/rtk/config.toml': No such file or directory

$ find / -maxdepth 6 -iname "config.toml" -path "*rtk*"
(no output — config.toml does not exist anywhere on the VM)

$ ~/.local/bin/rtk --version
rtk 0.45.0

$ ~/.local/bin/rtk gain --history
RTK Token Savings (Global Scope)
Total commands:    1
By Command
 1.  rtk proxy hermes doctor       1      0    0.0%    4.1s
Recent Commands
08-07 17:12 • rtk proxy hermes doctor   -0% (0)
```

Binary is installed (`rtk 0.45.0`) and one manual `rtk proxy hermes doctor`
invocation is recorded in history (17:12 today), but per spec's two required
sub-conditions, **`~/.config/rtk/config.toml` does not exist** — confirmed by a
targeted `ls` and a filesystem-wide `find`. This is not the documented
false-negative (`rtk init --show` misreporting) — the spec says judge overall
install status by `hermes plugins list` for *that* specific quirk. Checked
`hermes plugins list` too:

```
│ rtk-rewrite  │ enabled │ 0.1.0 │ Rewrite Hermes terminal commands through RTK before execution │ user │
```

`rtk-rewrite` shows enabled — the plugin registration is present. But live
proof it's wired into the running gateway is negative: invoking
`hermes memory status` on the VM directly emitted:

```
rtk: hermes plugin warning: rtk binary not found in PATH; Hermes hook not registered
```

Root cause confirmed by reading the gateway unit's env (read-only
`systemctl --user show`):

```
Environment=PATH=/home/gaetan/.hermes/hermes-agent/venv/bin:/home/gaetan/.hermes/hermes-agent/node_modules/.bin:...
```

`~/.local/bin` (where the only `rtk` binary lives) is absent from
`hermes-gateway.service`'s `PATH`. So the live gateway process cannot invoke
`rtk` at all — the one history entry is from a manual CLI proxy call, not the
hook firing during real gateway-mediated command execution.

**Verdict: FAIL.** Two independent gaps: (1) `~/.config/rtk/config.toml`
never created, (2) the gateway unit's `PATH` excludes the only rtk binary
location, so the hook is not live for actual gateway operation regardless of
`plugins list` showing "enabled". Owner: **lane B (B5)** — re-run rtk install
step (config.toml generation) and fix the gateway unit `PATH`/hook wiring.

### hindsight — PASS (skipped-by-decision, as specced)

```
$ ~/.local/bin/hermes memory status
Provider:  hindsight
Plugin:    installed ✓
Status:    not available ✗
Missing:
  ✗ HINDSIGHT_API_KEY  → https://ui.hindsight.vectorize.io
  ✗ HINDSIGHT_API_KEY
  ✗ HINDSIGHT_LLM_API_KEY
```

Matches Gaetan's 2026-08-07 decision: both HINDSIGHT_* keys deliberately
absent for v1. Gateway itself is up and served the A2A card and other CLI
commands throughout this session (see probe 12 below) — i.e. it runs fine
without memory. **Verdict: PASS** (memory-absent-by-design, gateway healthy).

### hermes-lcm — PASS

```
$ grep -n -A2 "context:" ~/.hermes/config.yaml
172:context:
173:  engine: lcm
174:mcp_servers:

$ ~/.local/bin/hermes plugins list | grep -A1 lcm
│ hermes-lcm  │ enabled │ 0.21.0-rc2 │ Lossless Context Management — ... │ git │
```

`context.engine: lcm` present in config.yaml (read-only grep, no edits) and
`hermes-lcm` shows `enabled` in `hermes plugins list`. **Verdict: PASS.**

### i-have-adhd — PASS

```
$ ~/.local/bin/hermes skills list
│ i-have-adhd  │  │ skills.sh │ community │ enabled │
```

Shows installed and enabled in `hermes skills list`. **Verdict: PASS.**

### Probe 11 rollup: **FAIL** (rtk sub-check fails; hindsight/lcm/adhd pass)
3 of 4 green, rtk red — overall probe 11 does not fully pass. Owner for the
rtk gap: **lane B (B5)**.

---

## Probe 12 — A2A agent card on 127.0.0.1:9900

```
$ date -Is
2026-08-07T19:01:09+00:00

$ curl -s -o /tmp/a2a-card.json -w "HTTP %{http_code}\n" \
    http://127.0.0.1:9900/.well-known/agent-card.json
HTTP 200

$ cat /tmp/a2a-card.json | python3 -m json.tool
{
    "name": "hermes-tars",
    "description": "Hermes Agent — a general-purpose agent reachable over A2A.",
    "url": "http://127.0.0.1:9900/",
    "version": "1.0.0",
    "provider": {
        "organization": "Hermes Agent",
        "url": "http://127.0.0.1:9900/"
    },
    "supportedInterfaces": [
        {"url": "http://127.0.0.1:9900/", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"}
    ],
    "capabilities": {
        "streaming": true,
        "pushNotifications": true,
        "stateTransitionHistory": false,
        "extendedAgentCard": false
    },
    "defaultInputModes": ["text/plain"],
    "defaultOutputModes": ["text/plain"],
    "skills": [ { "id": "toolset.a2a", "name": "a2a", ... } ]
}
```

200 JSON with `name` ("hermes-tars"), `capabilities` object, and `version`
("1.0.0") all present, as required.

Outbound-disabled / localhost-only check (expected, not a failure):

```
$ timeout 3 curl -s -o /dev/null -w "external HTTP %{http_code}\n" \
    http://192.168.0.9:9900/.well-known/agent-card.json
external HTTP 000   # connection refused / not bound externally — correct per D2/R1
```

**Verdict: PASS.** Localhost-bound 9900 serves a valid agent card;
non-localhost access correctly refused (by design, not a defect).

---

## Summary

| Probe | Sub-check | Verdict | Owner if not PASS |
|---|---|---|---|
| 11 | rtk | **FAIL** | lane B (B5) — config.toml never generated; gateway unit PATH excludes rtk binary, hook not live |
| 11 | hindsight | PASS | — (skipped-by-decision, confirmed not-available + gateway healthy) |
| 11 | hermes-lcm | PASS | — |
| 11 | i-have-adhd | PASS | — |
| 12 | A2A agent card | PASS | — |

Overall: **Probe 11 = FAIL (rtk sub-check), Probe 12 = PASS.**
