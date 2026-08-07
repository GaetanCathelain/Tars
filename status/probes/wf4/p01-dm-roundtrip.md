# WF4 Probe 1 — Slack DM round-trip

**Verdict: PASS.** `ping wf4-1` sent as Gaetan into the Tars DM `D0BBYNM01BL` at
`2026-08-07T19:01:31Z`; Tars replied `pong wf4-1` **6.04 s later** at `19:01:37Z`, in the same DM
(as a thread reply on the sent message). Gateway journal/log carries both the inbound event and the
outbound send for that window.

**Window: post-cutover.** Gateway `MainPID=61969`, `ActiveEnterTimestamp=2026-08-07 18:53:52 UTC`
(the `/sethome` restart, `status/probes/cutover-sethome.md` §4). Every action below is timestamped
after that. Local wall clock at probe start: `2026-08-07T19:00:06+00:00`.

**Secret handling.** No `sops` was run. Credentials were sourced *inside the remote shell*
(`set -a; . ~/.hermes/.env; set +a`) and handed to `curl` only through a `-K` config file written
with `printf` under `umask 077`, `shred -u`'d in an EXIT trap. Never on argv, never echoed. Only
prefixes and byte lengths appear below.

**Key-name correction (matters for the other probes):** the task brief named `SLACK_TOKEN` /
`SLACK_COOKIE`. Those are the **SOPS** key names; on the VM the same pair is stored as
`SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN`. Verified before use:

```
$ ssh gaetan@192.168.0.9 'for k in SLACK_TOKEN SLACK_COOKIE SLACK_MCP_XOXC_TOKEN SLACK_MCP_XOXD_TOKEN \
    SLACK_BOT_TOKEN SLACK_HOME_CHANNEL SLACK_ALLOWED_USERS; do printf "%s=%s\n" "$k" "$(grep -c "^$k=" ~/.hermes/.env)"; done'
SLACK_TOKEN=0            SLACK_COOKIE=0
SLACK_MCP_XOXC_TOKEN=1   SLACK_MCP_XOXD_TOKEN=1
SLACK_BOT_TOKEN=1        SLACK_HOME_CHANNEL=1        SLACK_ALLOWED_USERS=1
```

---

## 0 · Baseline — gateway state before the send (`19:01:25Z`)

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u);
    systemctl --user show hermes-gateway.service -p ActiveState -p SubState -p NRestarts -p MainPID -p ActiveEnterTimestamp'
MainPID=61969
NRestarts=0
ActiveState=active
SubState=running
ActiveEnterTimestamp=Fri 2026-08-07 18:53:52 UTC
```

Journal tail at that moment showed the gateway already busy with **another** WF4 agent's session in
channel `C08RWSTU9LK` (probe 2's traffic) — recorded so the log excerpts below can be told apart.

## 1 · Identity sanity + send (`19:01:30Z` → `19:01:31Z`)

Script piped to the VM over `ssh … 'sh -s'` (no secret in argv on either side):

```sh
set -a; . "$HOME/.hermes/.env"; set +a
: > "$CFG"; chmod 600 "$CFG"
printf 'header = "Authorization: Bearer %s"\n' "$SLACK_MCP_XOXC_TOKEN" >> "$CFG"
printf 'header = "Cookie: d=%s"\n'            "$SLACK_MCP_XOXD_TOKEN" >> "$CFG"   # as-stored, percent-encoded (A4)
curl -sS -K "$CFG" -X POST https://slack.com/api/chat.postMessage \
     --data-urlencode 'channel=D0BBYNM01BL' --data-urlencode 'text=ping wf4-1'
```

```
== host: tars user: gaetan
== T_start: 2026-08-07T19:01:30+00:00  epoch=1786129290
== xoxc prefix: xoxc-  len=111
== xoxd prefix: xoxd-  len=235
--- auth.test (identity sanity) ---
ok= True error= None user_id= U08BDJAMSRZ team_id= T7V1UGJ82 user= gaetan.cathelain
--- chat.postMessage -> D0BBYNM01BL ---
== T_post: 2026-08-07T19:01:30+00:00
ok= True error= None
channel= D0BBYNM01BL ts= 1786129291.039239
msg.user= U08BDJAMSRZ msg.text= 'ping wf4-1' msg.bot_id= None
--- chat.getPermalink (sent) ---
ok= True error= None
permalink= https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129291039239
== T_end: 2026-08-07T19:01:31+00:00
```

Sender is Gaetan's own account (`U08BDJAMSRZ` = `SLACK_ALLOWED_USERS`, team `T7V1UGJ82`), his own
DM with the Tars app, plain text — exactly spec §1's action. Cookie sent **as stored** (no
re-encoding), confirming `status/probes/slack-personal.md` probe 2 still holds post-cutover.

## 2 · The trap — `conversations.history` shows an empty DM for 2m24s

Eight polls, `oldest=<sent_ts>`, `19:01:52Z` → `19:03:55Z`, every ~15 s:

```
--- poll #1 at 2026-08-07T19:01:52+00:00 ---   ok= True n= 1
  ts=1786129291.039239 user=U08BDJAMSRZ bot_id=None text='ping wf4-1'
…
--- poll #8 at 2026-08-07T19:03:39+00:00 ---   ok= True n= 1
  ts=1786129291.039239 user=U08BDJAMSRZ bot_id=None text='ping wf4-1'
== NO REPLY within window, last check 2026-08-07T19:03:55+00:00      (exit 3)
```

**This is not a failure — it is a probe trap.** Tars replies **in-thread** (`thread_ts` = the
inbound message ts), and `conversations.history` returns only top-level messages. The reply had
already been posted at `19:01:37Z`, i.e. before poll #1 even ran.

## 3 · Reply located — `conversations.replies` (`19:05:04Z`)

```
--- conversations.replies channel=D0BBYNM01BL ts=1786129291.039239 ---
ok= True error= None n= 2
  ts=1786129291.039239 thread_ts=1786129291.039239 user=U08BDJAMSRZ bot_id=None      app_id=None       text='ping wf4-1'
  ts=1786129297.075679 thread_ts=1786129291.039239 user=U0BBH85NAKH bot_id=B0BBSBVBJUB app_id=A0BC0GXH78R text='pong wf4-1'
--- chat.getPermalink (reply) ---
ok= True error= None
permalink= https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129297075679?thread_ts=1786129291.039239&cid=D0BBYNM01BL
--- last 5 top-level messages in the DM (context) ---
  ts=1786129291.039239 user=U08BDJAMSRZ  bot_id=None       reply_count=1  text='ping wf4-1'
  ts=1781961770.329849 user=U0BBH85NAKH  bot_id=None       reply_count=12 text='You still there ?'
  ts=1781960554.662989 user=U0BBH85NAKH  bot_id=B0BBSBVBJUB reply_count=4 text=':warning: Gateway shutting down …'
```

| Item | Value |
|---|---|
| sent ts | `1786129291.039239` (2026-08-07T19:01:31Z) |
| sent permalink | https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129291039239 |
| reply ts | `1786129297.075679` (2026-08-07T19:01:37Z) |
| reply permalink | https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786129297075679?thread_ts=1786129291.039239&cid=D0BBYNM01BL |
| reply text | `pong wf4-1` |
| reply identity | bot user `U0BBH85NAKH` · `bot_id=B0BBSBVBJUB` · `app_id=A0BC0GXH78R` (the attached Tars app) |
| round-trip | **6.04 s** (well inside the ~2 min window) |
| same DM/thread | yes — `thread_ts` == sent ts, `channel=D0BBYNM01BL`, `reply_count=1` |

Reply content is responsive to the input (`ping wf4-1` → `pong wf4-1`), not empty/garbled — the
"WF3-owned if the gateway connects but replies are empty/wrong" branch of spec §1 does not apply.

## 4 · Gateway-side log — inbound event + outbound reply

The unit's journal only carries WARNING-level lines, so `journalctl` for the window shows nothing
about this exchange (it shows only the *other* session's model retries):

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u);
    journalctl --user -u hermes-gateway.service --since "2026-08-07 19:01:20" --no-pager -o short-iso'
2026-08-07T19:01:21 tars python[61969]: WARNING agent.conversation_loop: Empty response … retry 1/3 in 5.8s (model=gpt-5.6-sol)
2026-08-07T19:01:30 tars python[61969]: WARNING agent.conversation_loop: Empty response … retry 2/3 in 13.3s (model=gpt-5.6-sol)
2026-08-07T19:01:46 tars python[61969]: WARNING agent.conversation_loop: Empty response … retry 3/3 in 21.1s (model=gpt-5.6-sol)
2026-08-07T19:02:10 tars python[61969]: WARNING agent.conversation_loop: Empty response … after 3 retries. No fallback available. model=gpt-5.6-sol provider=openai-codex
```

Those four lines belong to session `20260807_190109_14287ca4` (chat `C08RWSTU9LK`), **not** to this
probe — see §5. The INFO-level record of this probe lives in `~/.hermes/logs/gateway.log` and
`~/.hermes/logs/agent.log`:

```
$ ssh gaetan@192.168.0.9 'grep -aE "^2026-08-07 19:0[1-5]" ~/.hermes/logs/gateway.log'
2026-08-07 19:01:09,377 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg='hello there' …
2026-08-07 19:01:32,889 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='ping wf4-1' reply_to_id=None reply_to_text=''   <-- INBOUND
2026-08-07 19:01:36,826 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=3.9s api_calls=1 response=10 chars                                  <-- OUTBOUND
2026-08-07 19:01:36,854 INFO gateway.platforms.base: [Slack] Sending response (10 chars) to D0BBYNM01BL                                                            <-- OUTBOUND
2026-08-07 19:02:10,351 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=61.0s api_calls=5 response=160 chars
2026-08-07 19:02:10,409 INFO gateway.platforms.base: [Slack] Sending response (160 chars) to C08RWSTU9LK
```

Agent-side detail for the same session:

```
$ ssh gaetan@192.168.0.9 'grep -aE "^2026-08-07 19:01:3[0-9]" ~/.hermes/logs/agent.log'
19:01:32,889 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='ping wf4-1'
19:01:32,943 INFO run_agent: OpenAI client created (agent_init) provider=openai-codex base_url=https://chatgpt.com/backend-api/codex model=gpt-5.6-sol
19:01:32,954 INFO [20260807_190132_acf22e5f] agent.turn_context: conversation turn: session=20260807_190132_acf22e5f model=gpt-5.6-sol provider=openai-codex platform=slack history=0 msg='ping wf4-1'
19:01:36,524 INFO [20260807_190132_acf22e5f] agent.conversation_loop: API call #1: model=gpt-5.6-sol provider=openai-codex in=24029 out=31 total=24060 latency=3.5s
19:01:36,539 INFO [20260807_190132_acf22e5f] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) api_calls=1/500 tool_turns=0 last_msg_role=assistant response_len=10
19:01:36,826 INFO gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=3.9s api_calls=1 response=10 chars
19:01:36,854 INFO gateway.platforms.base: [Slack] Sending response (10 chars) to D0BBYNM01BL
```

`response_len=10` == `len("pong wf4-1")`. Independent third artifact — the session store binds the
session to the exact message ts:

```
$ ssh gaetan@192.168.0.9 '<python3 walk of ~/.hermes/sessions/sessions.json>'
/agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786129291.039239/session_id  = 20260807_190132_acf22e5f
/agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786129291.039239/created_at  = 2026-08-07T19:01:32.891352
/agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786129291.039239/platform    = slack
/agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786129291.039239/origin/chat_id = D0BBYNM01BL
/agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786129291.039239/origin/user_id = U08BDJAMSRZ
```

Slack ts → session key → agent turn → outbound send → Slack reply ts: chain closed end to end.

## 5 · Note — concurrent traffic, not this probe's

Session `20260807_190109_14287ca4` (chat `C08RWSTU9LK`, msg `'hello there'`, `19:01:09`) is another
WF4 agent's. It hit `gpt-5.6-sol` empty-response retries 1/3→3/3 and eventually answered in **61.0 s
over 5 API calls**. This probe's DM turn hit **none** of that (1 API call, 3.5 s). Recorded because
(a) the journal lines in the probe window belong to it, and (b) it is a real reliability signal for
the WF4 report even though it does not change probe 1's verdict.

## Deviations / discipline

- No unit restarted, stopped, enabled or disabled. No `config.yaml` or `.env` edit. No `sops`. No
  git command. No `status/lane-*.md` edit. pve/p-Hermes untouched.
- Scratch files on the VM (`~/.hermes/.p01.*`) removed after the run; `-K` config `shred -u`'d by
  its trap. Verified: `ls: cannot access '/home/gaetan/.hermes/.p01*': No such file or directory`.

## Followups for the other WF4 probes

1. **`conversations.history` is blind to Tars' replies** — Tars answers in-thread. Probes 2, 3 and
   13 must read `conversations.replies` (or set `include_all_metadata`/check `reply_count`) before
   claiming silence. A `history`-only "no reply" is **not** valid negative evidence for probe 3 —
   it would have marked this PASS probe as FAIL.
2. **Tars app identity, for probe 3's ≠-Gaetan check:** app `A0BC0GXH78R`, bot user `U0BBH85NAKH`,
   `bot_id=B0BBSBVBJUB`. Gaetan is `U08BDJAMSRZ` (team `T7V1UGJ82`).
3. **VM env key names** are `SLACK_MCP_XOXC_TOKEN` / `SLACK_MCP_XOXD_TOKEN`, not
   `SLACK_TOKEN` / `SLACK_COOKIE` (those are the SOPS names). Probe 9's wiring reads the same pair.
4. **`gpt-5.6-sol` empty-response retries** observed on a sibling session (§5) — flag for probe 10.
