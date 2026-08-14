# Revert of `deleg_97b39490` — allowlist widening + `allowed_channels` (2026-08-14)

Approved by Gaetan ("Yeah revert everything") after
`status/probes/tars-dot-reply-2026-08-14-RCA.md` §Fix. Target: **192.168.0.9**
(Tars VM, clock UTC). Two values, no code, no `SOUL.md` edit.

**Verdict: APPLIED AND VERIFIED.** Both files reverted, gateway restarted at
13:02:18 UTC and healthy. The one thing this run cannot prove by itself is the
end-to-end behaviour — that needs a live non-Gaetan message (see §7).

No secret was decrypted, echoed, or written to this file. No `sops` invoked.
`SLACK_ALLOWED_USERS` values are Slack user IDs, not secrets; every other `.env`
key appears here by **name only**.

---

## 1. Pre-state established BEFORE any edit (13:01 UTC, read-only)

```
$ ssh gaetan@192.168.0.9 'date -u; ls -la --time-style=full-iso ~/.hermes/.env* ~/.hermes/config.yaml*'
Fri Aug 14 01:01:09 PM UTC 2026

-rw------- 908  2026-08-14 11:30:27 .env                      ← changed by deleg_97b39490
-rw------- 1423 2026-08-07 17:50:50 .env.bak-cutover
-rw------- 21   2026-08-07 16:30:41 .env.bak-linear-notion-cal
-rw------- 1675 2026-08-07 20:13:49 .env.bak-notiontrim
-rw------- 1644 2026-08-07 18:42:35 .env.bak-sethome
-rw------- 8115 2026-08-14 11:31:01 config.yaml               ← changed by deleg_97b39490
-rw------- 9994 2026-08-13 15:15:05 config.yaml.bak-2026-08-13-lcm-removal   (newest pre-change)
   … 18 older config backups …
```

**Confirmed: `deleg_97b39490` created NO `.bak` for either file** (newest `.env`
backup Aug 7, newest `config.yaml` backup Aug 13 15:15). The RCA's process
finding is reproduced. The backups this run created are therefore snapshots of
the *widened* state, not of the pre-11:30Z state — the pre-state is proven by
the Aug 7 / Aug 13 backups below instead.

### 1a. `.env` — what the prior value was

```
$ ssh … 'for f in .env .env.bak-cutover .env.bak-linear-notion-cal .env.bak-notiontrim .env.bak-sethome; do grep -h "^SLACK_ALLOWED_USERS" ~/.hermes/$f || echo "(key absent)"; done'
.env                       SLACK_ALLOWED_USERS=U08BDJAMSRZ,U7XJ4K631,U0BQTJUK2F2,U0BBH85NAKH
.env.bak-cutover           (key absent)
.env.bak-linear-notion-cal (key absent)
.env.bak-notiontrim        SLACK_ALLOWED_USERS=U08BDJAMSRZ      ← 2026-08-07 20:13
.env.bak-sethome           SLACK_ALLOWED_USERS=U08BDJAMSRZ      ← 2026-08-07 18:42
```

**No ambiguity.** Two independent backups prove the key has only ever held
`U08BDJAMSRZ` alone; the three additions are all from today's 11:30 edit:

| id | who | disposition |
|---|---|---|
| `U08BDJAMSRZ` | Gaetan | **kept** — the only allowed sender |
| `U7XJ4K631` | Oli | removed (added 11:30) |
| `U0BQTJUK2F2` | Ilo | removed (added 11:30) |
| `U0BBH85NAKH` | **Tars' own bot user id** | removed (added 11:30; inert either way — bot senders are dropped upstream of the allowlist) |

`U0BBH85NAKH` needed no lookup: it is Tars' bot id per the `operating-tars`
skill identity table and RCA §Fix.1. No Slack API call was made.

Bounded: there is no `.env` backup between 2026-08-07 20:13 and today's 11:30
edit. Corroboration that nothing else was allowed in between is behavioural —
RCA E1 shows Oli Early-rejected five times up to 11:29:52, i.e. the allowlist
was still Gaetan-only minutes before the change.

### 1b. `config.yaml` — `slack.allowed_channels`

```
$ ssh … 'grep -n -B4 -A2 allowed_channels ~/.hermes/config.yaml'
132-slack:
133-  require_mention: true
134-  strict_mention: true
135-  unauthorized_dm_behavior: ignore
136:  allowed_channels: C0BQCB58ATW
137-command_allowlist:

$ ssh … 'for f in ~/.hermes/config.yaml.bak*; do echo "$f: $(grep -c allowed_channels $f)"; done'
… bak-2026-08-13-lcm-removal: 0   ← newest pre-change backup, Aug 13 15:15
… bak-2026-08-13-approvals:   0
… bak-20260813-linear-expand: 0
… bak-20260813-narration:     0
… bak-20260813-dm-cutover:    0   (Aug 12 17:14)
… bak-gcn10-20260811:         0   (Aug 10 19:59)
… bak-boundaries-20260810-143638: 1   ← the ONE exception, see below
… all 12 others:              0
```

The single non-zero hit is historical and *different*:

```
$ ssh … 'grep -n -A1 allowed_channels ~/.hermes/config.yaml.bak-boundaries-20260810-143638'
226-slack:
227:  allowed_channels: C0BP2GZUFSR,C0BFQ5WFYTB
```

That is the retired reporting-channel era (2026-08-10); the key was dropped
during the 2026-08-13 DM-only cutover and is absent from **every** backup from
2026-08-10 19:59 onward. Pre-11:30Z state = **key absent** → delete it entirely,
as instructed. (This also un-deafens Tars in `#gcn-sandbox` etc. — RCA
§Collateral finding.)

---

## 2. The edits (13:02:08 UTC)

Executed as one script piped to `ssh gaetan@192.168.0.9 'bash -s'`. Backups
first, both `sed -i` calls inside a single `flock` critical section, in-place
edits (no tmp+mv, so no mode loss — `chmod 600` applied anyway per the hard rule).

```bash
H="$HOME/.hermes"

cp -p "$H/.env"        "$H/.env.bak-20260814-revert"
cp -p "$H/config.yaml" "$H/config.yaml.bak-20260814-revert"

flock "$H/.wf3.lock" -c "
  sed -i 's/^SLACK_ALLOWED_USERS=.*/SLACK_ALLOWED_USERS=U08BDJAMSRZ/' '$H/.env'
  sed -i '/^  allowed_channels: C0BQCB58ATW\$/d' '$H/config.yaml'
"

chmod 600 "$H/.env" "$H/config.yaml" "$H/.env.bak-20260814-revert" "$H/config.yaml.bak-20260814-revert"
```

Output:

```
=== T0 2026-08-14 13:02:08 UTC ===
backups created:
-rw------- 1 gaetan gaetan 8115 Aug 14 11:31 /home/gaetan/.hermes/config.yaml.bak-20260814-revert
-rw------- 1 gaetan gaetan  908 Aug 14 11:30 /home/gaetan/.hermes/.env.bak-20260814-revert
edits applied at 2026-08-14 13:02:08 UTC
=== T1 2026-08-14 13:02:08 UTC ===
```

(`cp -p` kept the source mtimes on the `.bak`s — 11:30/11:31 — which is exactly
the state they are meant to preserve.)

### Blast radius — proof nothing else moved

```
.env    diff lines vs pre-edit backup: 4     (= "8c8" / "<" / "---" / ">", one line changed)
config  diff lines vs pre-edit backup: 2     (= "136d135" / "<", one line deleted)
config  '>' lines (additions):         0     (deletion only — nothing added or reordered)
```

Diff *content* deliberately not printed (both files carry credentials); the line
counts plus the targeted greps below are the evidence.

```
before: .env 908 B / config.yaml 8115 B
after:  .env 874 B / config.yaml 8083 B
        Δ.env    = -34  = len("U7XJ4K631,U0BQTJUK2F2,U0BBH85NAKH") + 1 comma  ✓
        Δconfig  = -32  = len("  allowed_channels: C0BQCB58ATW\n")           ✓
```

`.env` still holds all 11 keys, same order (names only, values never read):
`HINDSIGHT_MODE LINEAR_API_KEY NOTION_API_TOKEN GMAIL_ADDRESS GMAIL_APP_PASSWORD
SLACK_MCP_XOXC_TOKEN SLACK_MCP_XOXD_TOKEN SLACK_ALLOWED_USERS SLACK_BOT_TOKEN
SLACK_APP_TOKEN SLACK_HOME_CHANNEL` — identical to the pre-edit listing. No
sibling key dropped (the `mcp_servers` near-miss class of failure did not recur).

`config.yaml` slack stanza after the edit — byte-identical to the 2026-08-13
15:15 backup's stanza:

```
132:slack:
133-  require_mention: true
134-  strict_mention: true
135-  unauthorized_dm_behavior: ignore
136-command_allowlist:
```

```
$ python3 -c "import yaml; yaml.safe_load(open('~/.hermes/config.yaml')); print('YAML OK')"
YAML OK
```

---

## 3. Restart (`.env` is read only at process start)

```
$ ssh … 'export XDG_RUNTIME_DIR=/run/user/$(id -u); systemctl --user list-units --type=service | grep -i herm'
  hermes-cloakbrowser.service loaded active running
  hermes-gateway.service      loaded active running

BEFORE (13:02:14 UTC)   MainPID=1784687  ActiveEnterTimestamp=Fri 2026-08-14 11:35:26 UTC  active/running
$ systemctl --user restart hermes-gateway.service     → exit=0   (issued 13:02:17)
AFTER  (13:02:38 UTC)   MainPID=1807357  ActiveEnterTimestamp=Fri 2026-08-14 13:02:18 UTC  active/running
```

Restart proven by `ActiveEnterTimestamp` moving forward **and** `MainPID`
changing (`NRestarts` deliberately not used — unreliable on this host).
Ordering that matters: edits 13:02:08 → new process start 13:02:23 → the
reverted `.env` is what the live gateway loaded.

### `gateway.log` — clean shutdown + clean start

```
13:02:17,767  Shutdown phase: post-interrupt tool kill done at +0.32s
13:02:17,902  [Slack] Disconnected   /  ✓ slack disconnected
13:02:18,095  ✓ a2a disconnected (0.19s)
13:02:18,110  WARNING Shutdown (final-cleanup): marked 1 in-flight cron job(s) interrupted: 62…
13:02:18,125  Gateway stopped (total teardown 0.68s)
13:02:23,508  Starting Hermes Gateway...
13:02:23,515  Secret redaction: ENABLED
13:02:23,765  [Slack] Authenticated as @tars in workspace Mobil…
13:02:23,816  [Slack] Socket Mode connected (1 workspace(s))
13:02:23,824  ✓ slack connected
13:02:23,839  ✓ a2a connected
13:02:23,846  Gateway running with 2 platform(s)
13:02:24,407  Channel directory built: 107 target(s)
13:02:25,421  Gateway housekeeping started (interval=60s)
13:02:30,418  kanban dispatcher: embedded in gateway (interval=60.0s)
```

No startup errors. Channel directory 107 targets (was 105 at the 11:35 restart —
directory growth, unrelated to the allowlist which is a *sender* gate).

### `errors.log`, 13:02 onward

```
13:02:12  WARNING [cron_62e8cd9db637_…] check_fn check_bfl/computer_use/image_generation/web_api_key returned False   ← pre-existing noise
13:02:17  WARNING Shutdown context: signal=SIGTERM under_systemd=yes …
13:02:17  WARNING Gateway drain timed out after 0.0s with 0 active agent(s), 1 in-flight cron …
13:02:18  WARNING Shutdown (post-interrupt/final-cleanup): marked 1 in-flight cron job(s) interrupted: 62…
13:02:23  WARNING slack_bolt.AsyncApp: As you gave `client` as well, `token` will be unused.   ← known benign
13:02:25  WARNING cron.scheduler_provider: Marked 1 interrupted cron execution(s) unknown after restart
```

Only side effect worth naming: **cron job `62e8cd9db637` was mid-execution at
13:02:17 and was interrupted**, now marked `unknown`. Collateral of restarting
during a cron run, not of the config change; the job runs on its own schedule
again. Everything else is the same tool-capability-check noise documented in
`gateway-restart-2026-08-14.md`.

---

## 4. Post-checks (13:02:56 UTC)

```
$ grep -n "^SLACK_ALLOWED_USERS" ~/.hermes/.env
8:SLACK_ALLOWED_USERS=U08BDJAMSRZ                                   ✅ matches the Aug 7 backups exactly

$ grep -n allowed_channels ~/.hermes/config.yaml
(no match, exit=1)                                                  ✅ key gone

$ ls -l
-rw------- 8083 Aug 14 13:02 config.yaml                            ✅ 600
-rw------- 8115 Aug 14 11:31 config.yaml.bak-20260814-revert        ✅ 600, .bak exists
-rw-------  874 Aug 14 13:02 .env                                   ✅ 600
-rw-------  908 Aug 14 11:30 .env.bak-20260814-revert               ✅ 600, .bak exists

$ systemctl --user is-active hermes-gateway.service
active                                                              ✅
MainPID=1807357  ActiveEnterTimestamp=Fri 2026-08-14 13:02:18 UTC

$ ~/.local/bin/hermes chat -Q -q "reply with the single word: alive"
session_id: 20260814_130257_079a5b
alive
EXIT=0                                                              ✅ agent path healthy
```

All six required post-checks pass.

---

## 5. Rollback (if this revert ever needs undoing)

```
cp -p ~/.hermes/.env.bak-20260814-revert        ~/.hermes/.env
cp -p ~/.hermes/config.yaml.bak-20260814-revert ~/.hermes/config.yaml
chmod 600 ~/.hermes/.env ~/.hermes/config.yaml
# + restart, per §3
```
Those two `.bak`s are the exact 11:30/11:31 widened state.

---

## 6. Scope — what was NOT touched

- `~/.hermes/SOUL.md` — unchanged. Rule 4 and its `·` terminal stay as specified
  (RCA §Fix: rewording re-opens the `⚠️ Empty response` retry storm).
- No Hermes code patched. No other `config.yaml` key, no other `.env` key.
- `192.168.0.3` (pve) and `192.168.0.8` (p-Hermes) untouched.
- Nothing posted to Slack by this run (the gateway's own restart notice to the
  home DM `D0BBYNM01BL` is automatic, not ours).

---

## 7. What is still unverified

**End-to-end behaviour.** The revert is proven at the file + process level, not
yet in Slack. Closing it needs a live message and cannot be self-served:

```bash
ssh gaetan@192.168.0.9 'grep -E "^2026-08-14 " ~/.hermes/logs/gateway.log \
  | grep C0BQCB58ATW | grep -E "inbound message|response ready|Early reject" | tail -6'
```

Pass = Oli's next `@Tars` in `C0BQCB58ATW` produces `WARNING … Early reject of
unauthorized user U7XJ4K631` with **no** `response ready` and **no** Slack post;
Gaetan's produces a normal multi-hundred-char response; and Tars answers an
`@`-mention in `#gcn-sandbox` (`C08RWSTU9LK`) again — the collateral fix.

Secondary, from the RCA's own open list: other surfaces were not swept for `·`
occurrences during the 11:35→13:02 window when the widened allowlist was live
(Oli/Ilo could also have DM'd Tars directly). Anything found there is historical
now — the gate is closed.
