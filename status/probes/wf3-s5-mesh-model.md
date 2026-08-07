# WF3 §5 — Orca/SSH mesh + model backend — **PASS**

Agent: WF3 wiring agent 5 ("mesh-model"). Date: 2026-08-07.
Spec: `docs/specs/wf3-wiring.md` §5. Placeholders from `status/lane-b.md` (B1–B7).

## Placeholders resolved

| Token | Value | Source |
|---|---|---|
| `<TARS_IP>` | `192.168.0.9` | lane-b B1/B2 |
| `<TARS_USER>` | `gaetan` | lane-b B1/B2 |
| `<TARS_HOST>` | `tars.tail6e788b.ts.net` / `100.116.31.76` | lane-b B2 (tailscale) |
| `<MACOS_HOST>` | `macbook-pro-de-gaetan.tail6e788b.ts.net` / `100.78.160.92` | **not in lane-b**; derived from `tailscale status` on the VM (the only macOS node owned by `gaetan.cathelain@`), then *proven* the same host as the old DHCP alias by host-key fingerprint (below) |

All hermes calls used `~/.local/bin/hermes` (not on PATH over non-interactive ssh).

---

## Part 1 — Model backend (done first; agents 2 and 4 depend on it)

### 1.1 `--help` verification before scripting (mandatory, R1/R7)

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes config set --help'
usage: hermes config set [-h] [--force] [key] [value]
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes auth status --help'
usage: hermes auth status [-h] provider
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes chat --help'
usage: hermes chat [-h] [-q QUERY] ... [-Q] ...
  -q QUERY, --query QUERY   Single query (non-interactive mode)
  -Q, --quiet               Quiet mode for programmatic use ... Only output the final response
```

**Spec CLI fact is wrong: there is no `--oneshot` flag** in v0.20.0. Non-interactive single query
is `-q/--query`, with `-Q/--quiet` for programmatic output. Probe adjusted accordingly (allowed:
"adjusted only by verified `--help` facts"). Verdict: **PASS (deviation recorded)**.

### 1.2 Baseline (before any edit)

```
model.provider = auto
model.default  = anthropic/claude-opus-4.6
model.base_url = https://openrouter.ai/api/v1
```

**The OpenRouter `base_url` SURVIVED lane A's A1b OAuth** — `hermes auth add openai-codex` did
*not* rewrite it. The spec's contingency ("if the installer default survives, `config set` it
explicitly") is the branch that actually fired. An add-if-absent implementation would have left
Tars on OpenRouter with no key, exactly as lane-b B7 warned.

### 1.3 Backup + flock-guarded override

```
$ ssh ... 'cp -n ~/.hermes/config.yaml ~/.hermes/config.yaml.bak-mesh-model'
backup=created   -rw------- 5509 Aug 7 17:43 /home/gaetan/.hermes/config.yaml.bak-mesh-model

$ ssh ... 'flock ~/.hermes/.wf3.lock -c "$H config set model.provider openai-codex"'
✓ Set model.provider = openai-codex in /home/gaetan/.hermes/config.yaml    exit=0
$ ssh ... 'flock ~/.hermes/.wf3.lock -c "$H config set model.default gpt-5.6-sol"'
✓ Set model.default = gpt-5.6-sol in /home/gaetan/.hermes/config.yaml      exit=0
$ ssh ... 'flock ~/.hermes/.wf3.lock -c "$H config set model.base_url https://chatgpt.com/backend-api/codex"'
✓ Set model.base_url = https://chatgpt.com/backend-api/codex               exit=0
```

Every edit wrapped in `flock ~/.hermes/.wf3.lock -c`; one-time `.bak-mesh-model` taken first.
`~/.hermes/.env` was **not** touched by this agent. Post-edit integrity:

```
model.provider = openai-codex
model.default  = gpt-5.6-sol
model.base_url = https://chatgpt.com/backend-api/codex
python3 yaml.safe_load(config.yaml) -> parse=ok keys= 24        # not half-written
```

Verdict: **PASS**.

### 1.4 Probe — auth

```
$ ~/.local/bin/hermes auth status openai-codex
openai-codex: logged in                                   exit=0

$ ~/.local/bin/hermes auth list
copilot (1 credentials):
  #1  gh auth token        api_key gh_cli ←
openai-codex (1 credentials):
  #1  openai-codex-oauth-1 oauth   device_code ←          exit=0
```

`openai-codex`: exactly one credential, `source: device_code` — matches the pass line. (The
`copilot` entry is a *different* provider, registered by another WF3 agent's `gh` login; it does
not violate "one credential" for openai-codex.) Verdict: **PASS**.

### 1.5 Probe — live inference (the actual pass condition)

```
$ ~/.local/bin/hermes chat -Q -q "Reply with exactly one word: sol"
rtk: hermes plugin warning: rtk binary not found in PATH; Hermes hook not registered
  ⚠ tirith security scanner enabled but not available — command scanning will use pattern matching only
session_id: 20260807_174352_719067
sol
exit=0

$ ~/.local/bin/hermes dump | grep -iE '^ *(model|provider)'
model:            gpt-5.6-sol
provider:         openai-codex
```

**First real inference on the Tars VM.** Reply is exactly `sol`, served by `gpt-5.6-sol` over
`openai-codex`. This closes R7's open question: **the ChatGPT subscription tier does reach Sol.**
No rollback needed, `.bak-mesh-model` left in place. Verdict: **PASS**.

---

## Part 2 — SSH mesh

### 2.1 Key material — no decrypt was run

`TARS_VM_SSH_PRIVATE_KEY` **is not in `secrets/tars.sops.yaml`**. The store holds 8 keys
(`NOTION_FILE_TOKEN`, `NOTION_SPACE_ID`, `LINEAR_API_KEY`, `GMAIL_ADDRESS`, `GMAIL_APP_PASSWORD`,
`NOTION_API_TOKEN`, `SLACK_TOKEN`, `SLACK_COOKIE`) — none SSH. This is consistent, not a gap:
lane-b B3 generated the keypair **on the VM** and states the private half never left it.

```
$ ssh gaetan@192.168.0.9 'ls -l ~/.ssh/; ssh-keygen -lf ~/.ssh/id_ed25519.pub'
-rw------- 398 authorized_keys
-rw------- 409 config
-rw------- 411 id_ed25519            <- already present, 0600
-rw-r--r--  98 id_ed25519.pub
-rw------- 2934 known_hosts
256 SHA256:5shawRKsUq5JzYXazx5shAAhwBXyRGzc2euLaQv3FYk tars@192.168.0.9 (ED25519)
```

Fingerprint matches lane-b B3 byte-for-byte. **No `sops -d` of any form was executed by this
agent** — no per-key extract, and certainly no whole-file decrypt. Verdict: **PASS (deviation:
spec step 1's key delivery was already satisfied by B3 and is a no-op)**.

### 2.2 Leg probes run FIRST (before any authorized_keys change)

```
# from the VM
$ ssh -o BatchMode=yes -o ConnectTimeout=5 cooper true; echo "cooper:$?"     -> cooper:0
$ ssh -o BatchMode=yes -o ConnectTimeout=5 mac true;    echo "macos-lan:$?"  -> macos-lan:0
# from cooper
$ ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 true; echo "tars:$?" -> tars:0
```

All three legs passed on the first probe — lane B had already keyed every target. Therefore
**no `authorized_keys` line was appended anywhere**, on cooper, on the Mac, or on the VM.
**Nothing under cooper's `~/.ssh/` was read, printed, or written** (deny-list respected; the
append-only `>>` path was never needed).

### 2.3 MANDATORY EDIT — `mac` alias repointed off the DHCP lease

Same-host proof before touching anything (fingerprints, not keys):

```
$ ssh-keygen -F 192.168.0.23 -f ~/.ssh/known_hosts | ssh-keygen -lf -
256 SHA256:NEV0aBbhcthXTf2/98sk75HPyU5A/bRN8XFDa++1oFI (ED25519)     # already trusted
$ ssh-keyscan -T 5 -t ed25519 192.168.0.23 | ssh-keygen -lf -
256 SHA256:NEV0aBbhcthXTf2/98sk75HPyU5A/bRN8XFDa++1oFI 192.168.0.23 (ED25519)
$ ssh-keyscan -T 5 -t ed25519 macbook-pro-de-gaetan.tail6e788b.ts.net | ssh-keygen -lf -
256 SHA256:NEV0aBbhcthXTf2/98sk75HPyU5A/bRN8XFDa++1oFI macbook-pro-de-gaetan.tail6e788b.ts.net (ED25519)
```

Identical host key on both addresses ⇒ same machine, not a guess.

```
$ cp -n ~/.ssh/config ~/.ssh/config.bak-mesh-model          -> ssh-config backup=created
$ ssh-keyscan -T 5 macbook-pro-de-gaetan.tail6e788b.ts.net >> ~/.ssh/known_hosts   # grep-guarded
known_hosts: appended 3 lines                               # 3 key types, no interactive prompt possible
$ sed -i ... ~/.ssh/config
Host mac
    HostName macbook-pro-de-gaetan.tail6e788b.ts.net        # was 192.168.0.23 (DHCP)
    User gcath
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
```

### 2.4 Final leg probes, post-repoint

| Leg | Command | Result |
|---|---|---|
| VM → cooper | `ssh -o BatchMode=yes -o ConnectTimeout=5 cooper true` | `cooper:0` **PASS** |
| VM → macOS (tailnet) | `ssh -o BatchMode=yes -o ConnectTimeout=5 mac true` | `macos-tailnet:0` **PASS** |
| VM → macOS identity | `ssh mac hostname` | `MacBook-Pro-de-Gaetan.local`, exit 0 |
| cooper → VM | `ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 true` | `tars:0` **PASS** |

Exit 0 on every configured leg, no prompt, no output on the bare `true` runs. Verdict: **PASS**.

### 2.5 Tars → p-Hermes: deliberately UNWIRED

Untouched. `192.168.0.3` was not contacted in any form by this agent — no ssh, no `qm guest exec`,
no keyscan. Spec's flag to Gaetan stands: **confirm before WF4 writes a probe assuming this leg
exists.**

---

## Invariants re-checked after the run

| Invariant | Observed |
|---|---|
| gateway not enabled / not started | `systemctl --user is-enabled hermes-gateway` → `disabled`; `is-active` → `inactive` |
| `auth.json` never read/copied/mirrored | present, untouched — existence checked with `ls >/dev/null`, contents never read |
| p-Hermes (192.168.0.3 / VM 103) | not contacted |
| `~/.hermes/.env` | not touched by this agent |
| `config.yaml` parses | `yaml.safe_load` ok, 24 top-level keys |
| no OAuth / browser flow | none attempted |

---

## Deviations

1. **`hermes chat --oneshot` does not exist** (v0.20.0). Used the verified `-Q -q "…"`. The spec's
   probe line needs correcting for WF4.
2. **`TARS_VM_SSH_PRIVATE_KEY` is not in the SOPS store** and does not need to be — B3 generated
   the keypair on the VM. Spec §5 Wire step 1 (extract-pipe the private half) is a **dead step**;
   executing it as written would have overwritten a working key with a nonexistent one.
3. **No `authorized_keys` append on any target** — all legs already passed. Wire step 1's public-half
   distribution was already done by B3.
4. **`base_url` was NOT written by `hermes auth add`** — the OpenRouter installer default survived
   A1b. Set explicitly. Treat "OAuth writes base_url" as false for v0.20.0.
5. **`<MACOS_HOST>` was not in lane-b.md** in resolvable form (B3 gives only the DHCP IP). Resolved
   from `tailscale status` on the VM — an explicitly permitted source — and pinned by host-key
   fingerprint rather than by name matching.

## Surprises worth flagging

- **The hermes units are `--user` units, not system units.** `systemctl is-enabled hermes` returns
  `not-found`; the real names are `hermes-gateway.service` (disabled) and
  `hermes-cloakbrowser.service` (enabled), under `systemctl --user`. Any WF4 probe using system
  `systemctl` will silently report the wrong thing.
- **Two warnings on every `hermes chat` run** (non-fatal, inference still succeeded):
  `rtk: hermes plugin warning: rtk binary not found in PATH; Hermes hook not registered` — the
  rtk-rewrite plugin is enabled but its binary is not on the non-interactive PATH; and
  `⚠ tirith security scanner enabled but not available — command scanning will use pattern
  matching only`. The second is a **security control running in degraded mode** and predates this
  agent; flagged, not fixed.
- `hermes auth list` now shows a `copilot` credential (`api_key`, `gh_cli`) alongside openai-codex —
  a side effect of another WF3 agent's `gh` login, harmless here but it means "one credential"
  must be read per-provider.

## Rollback state

None applied — nothing failed. Left in place for the human:
`~/.hermes/config.yaml.bak-mesh-model`, `~/.ssh/config.bak-mesh-model` (both on the VM).
