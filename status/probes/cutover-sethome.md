# Cutover — `/sethome` equivalent, applied manually

**Verdict: PASS.** `SLACK_HOME_CHANNEL=D0BBYNM01BL` is in `~/.hermes/.env` on the Tars VM
(192.168.0.9) and the gateway restarted clean onto it.

**Why manual:** the Tars Slack app is an **Agent-class** app, so Slack refuses `/sethome` in its
chat surface ("not supported in threads"). Per `docs/specs/cutover-notes.md` and
`docs/specs/tars-profile.md` §D2, `/sethome`'s only effect is writing `SLACK_HOME_CHANNEL` to
`.env` — reproduced here byte-for-byte, no `override.conf` (DECISION.md §84 holds).

Date: 2026-08-07 · VM `tars` @ 192.168.0.9 · no secret was printed, put on argv, or decrypted
(`SLACK_BOT_TOKEN` was already on the VM in `~/.hermes/.env`; it was sourced inside the remote
shell and handed to `curl` via a `-K` config file on stdin).

---

## 1 — Native mechanism check: there is none but `.env`

`hermes --help` / `hermes config --help` / `hermes gateway --help` / `hermes slack --help` were
read in full. `hermes config` has `get/set/unset` over `config.yaml`; `hermes slack` only has
`manifest`. Neither exposes a home channel.

```
$ hermes config get slack.home_channel
Config key not set: slack.home_channel

$ python3 -c "yaml.safe_load(open('~/.hermes/config.yaml'))" → keys matching -i 'home'
KEY: terminal.home_mode = 'auto'
KEY: platform_toolsets.homeassistant = ['hermes-homeassistant']
   (no slack home key; top-level 'slack:' block carries none)
```

Source confirms **env-only**, and that `/sethome` itself just writes `.env`:

- `plugins/platforms/slack/adapter.py:8968-8972` — the only writer:
  `save_env_value("SLACK_HOME_CHANNEL", home_channel)` (the same call `hermes gateway setup`'s
  interactive prompt makes; `remove_env_value(...)` on a blank answer).
- `gateway/config.py:1990-2001` — the only reader: `slack_home = getenv("SLACK_HOME_CHANNEL")`
  → `HomeChannel(chat_id=slack_home, name=getenv("SLACK_HOME_CHANNEL_NAME",""), …)`.
- `cron/scheduler.py:268` — `"slack": "SLACK_HOME_CHANNEL"` (cron delivery target).
- `adapter.py:9088-9089` — plugin descriptor `cron_deliver_env_var="SLACK_HOME_CHANNEL"`.
- `adapter.py::_apply_yaml_config` — the `config.yaml`→`SLACK_*` translation hook; grepping it
  for `home` returns only the cron comment. **No YAML key is translated into
  `SLACK_HOME_CHANNEL`.** Env is also declared authoritative there ("Env vars take precedence
  over YAML").

The only native, non-interactive path is therefore the `.env` key — which is what was done. The
one native *interactive* path (`hermes gateway setup`) writes the identical key to the identical
file and would have re-prompted for both tokens; skipped.

`hermes config env-path` → `/home/gaetan/.hermes/.env` — the file written is the file the gateway
loads.

## 2 — DM channel ID resolved: `D0BBYNM01BL`

```
$ set -a; . ~/.hermes/.env; set +a
$ printf 'header = "Authorization: Bearer %s"\n' "$SLACK_BOT_TOKEN" \
  | curl -sS -K - -X POST https://slack.com/api/conversations.open \
        -d users=U08BDJAMSRZ -d return_im=true | python3 -c '<extract ok/channel.id>'
ok= True
error= None
channel.id= D0BBYNM01BL
is_im= True user= U08BDJAMSRZ
no_op= True already_open= True
```

`ok:true`, id starts with `D`, `is_im:true`, `user` is Gaetan — this is the Tars-bot↔Gaetan DM.
`already_open:true` means the conversation pre-existed (nothing was created, no message sent).

**Cross-check:** `docs/recon/r3-p-hermes.md:170` recorded p-Hermes' Tars unit as
`Environment="SLACK_HOME_CHANNEL=D0BBYNM01BL"` — **identical**. The new VM lands on the same home
DM as the old deployment, independently derived from the API. `conversations.open` was also a
live positive check on `SLACK_BOT_TOKEN` before the restart.

## 3 — `.env` write (flock + atomic `mv`)

Whole edit ran under `flock 9` on `~/.hermes/.wf3.lock`, `umask 077`; built as `.env.new`,
key-count asserted **before** the `mv`, aborting without touching `.env` on mismatch.

```
cp -p ~/.hermes/.env ~/.hermes/.env.bak-sethome && chmod 600 …
grep -v -E '^SLACK_HOME_CHANNEL=' .env > .env.new    # exact key; *_NAME/*_THREAD_ID untouched
[ -n "$(tail -c1 .env.new)" ] && printf '\n' >> .env.new   # no line-splice if file lacked EOF NL
printf 'SLACK_HOME_CHANNEL=%s\n' D0BBYNM01BL >> .env.new
chmod 600 .env.new ; [ keys(new) -eq PRE+1 ] || abort ; mv -f .env.new .env
```

```
PRE  key count: 13    PRE  SLACK_HOME_CHANNEL occurrences: 0    PRE  mode: 600
POST key count: 14    POST SLACK_HOME_CHANNEL occurrences: 1    POST mode: 600
POST value line: SLACK_HOME_CHANNEL=D0BBYNM01BL
diff pre.keys post.keys:
  11a12
  > SLACK_HOME_CHANNEL
backup: 600 /home/gaetan/.hermes/.env.bak-sethome
```

Key-name diff is a single addition: **no other key changed, was reordered, or was lost.** 13 → 14
matches `status/probes/cutover.md` §1d, which recorded `SLACK_HOME_CHANNEL` as the one expected
absentee. Idempotent: a rerun removes then re-appends, still one occurrence.

## 4 — Gateway restart, healthy

```
$ systemctl --user restart hermes-gateway.service
before: NRestarts=0 ActiveState=active ExecMainStart=18:43:08  MainPID=60158
after:  NRestarts=0 ActiveState=active SubState=running ActiveEnter=18:53:52  MainPID=61969
+70s:   NRestarts=0 ActiveState=active SubState=running  MainPID=61969  (unchanged)
```

`NRestarts` stayed **0** across the restart — a manual `restart` does not increment it, so 0 after
+70s proves `Restart=always` never fired, i.e. **no crash loop**. The one
`Main process exited, code=exited, status=1/FAILURE` line is the *old* PID 60158 exiting under the
`SIGTERM` logged immediately above it, then `Stopped` → `Started`; the new process has run clean
since.

### Journal — complete since restart (14 lines, nothing redacted, no token fragment)

```
$ journalctl --user -u hermes-gateway.service -n 60 --no-pager -o short-iso
18:53:51 systemd:       Stopping hermes-gateway.service - Hermes Agent Gateway…
18:53:51 python[60158]: WARNING gateway.run: Shutdown context: signal=SIGTERM under_systemd=yes
18:53:52 python[60158]: ┌──── ⚕ Hermes Gateway Starting… ────┐  (banner, 6 lines)
18:53:52 systemd:       hermes-gateway.service: Main process exited, code=exited, status=1/FAILURE
18:53:52 systemd:       Stopped hermes-gateway.service …
18:53:52 systemd:       Started hermes-gateway.service - Hermes Agent Gateway - Messaging Platform…
18:53:55 python[61969]: WARNING slack_bolt.AsyncApp: As you gave `client` as well, `token` will be unused.
18:54:00 python[61969]: WARNING gateway.run: Synthetic event source unresolvable: session_key='202608…'
18:54:00 python[61969]: WARNING gateway.run: Dropping watch notification for raw session 20260807_…
```

The three post-restart WARNINGs are byte-identical in kind to the 18:43 healthy boot recorded in
`status/probes/cutover.md` §3a — same benign set, nothing new. **No journal line mentions the home
channel** (the unit does not log it at this level), so nothing to quote there.

```
$ journalctl --user -u hermes-gateway.service --since '2026-08-07 18:53:50' --no-pager \
  | grep -icE 'invalid_auth|not_authed|account_inactive|token_revoked|Refusing to start|No env user allowlists'
0
```

- **No auth error** of any kind.
- **`No env user allowlists configured` absent** → `gateway/run.py:10940-11007` found
  `os.getenv("SLACK_ALLOWED_USERS")` truthy at startup → **`SLACK_ALLOWED_USERS` survived the
  `.env` rewrite and is still in effect.** (Same inference as `cutover.md` §3c; here it doubles as
  the regression check on the edit.)

### Socket Mode connect (unit does not log it — proved by socket + DNS, as in §3b of cutover.md)

```
$ ss -tnp | grep pid=61969
ESTAB 0 0  192.168.0.9:36870  3.68.63.139:443  users:(("hermes",pid=61969,fd=32))

$ getent ahostsv4 wss-primary.slack.com
18.197.249.189 3.126.186.102 3.65.102.105 3.67.245.95 3.68.18.70 3.68.61.181 3.68.63.139
```

`3.68.63.139` ∈ the `wss-primary.slack.com` A-record set. The **new** PID holds exactly one
established outbound TLS session to Slack's Socket Mode endpoint, still held at +70s. A rejected
`xapp-` token fails `apps.connections.open` and no WSS session is ever opened — the connection is
itself the positive auth proof. (Fresh connect: different local port than the pre-restart run.)

---

## Scope discipline

Not touched: no Slack message or DM sent (`conversations.open` on an already-open IM is not a
message), p-Hermes (192.168.0.3 / VM 103) and pve untouched, no `sops` invocation, no
`status/lane-*.md` edit, no git command. Rollback stands ready and unused:
`~/.hermes/.env.bak-sethome` (0600) — restore under the same flock and restart.

## Residual

`SLACK_HOME_CHANNEL_NAME` / `SLACK_HOME_CHANNEL_THREAD_ID` are unset. Both are optional
(`getenv(…, "")` / `or None` in `gateway/config.py:1999-2001`) and a DM needs neither.

End-to-end delivery into this DM is **not** proven here — that is
`docs/specs/wf4-probes.md` §13 (`/bg` completion lands in the home DM), which is the probe this
step unblocks.
