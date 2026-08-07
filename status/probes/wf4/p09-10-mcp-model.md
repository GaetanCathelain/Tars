# WF4 probes 9 + 10 — live Hermes capability (slack-personal MCP, gpt-5.6-sol)

Agent: WF4 verification probe (probes 9, 10). Host under test: `tars` VM `192.168.0.9`.
Probe window: **2026-08-07T19:00:07+00:00 → 2026-08-07T19:04:22+00:00** (all `date -Is`, UTC).
Post-cutover: gateway `hermes-gateway.service` `ActiveEnterTimestamp=Fri 2026-08-07 18:53:52 UTC`,
`is-active=active` — every action below is after that restart.

**Verdict: Probe 9 PASS · Probe 10 PASS.**

Rule compliance: no `sops` was run (credentials sourced from `~/.hermes/.env` on the VM only);
no unit started/stopped/restarted; no config or `.env` written (mtime proof below); no `pve`
contact; no git commands; secrets never on argv and never printed — all Slack calls go through a
`curl -K` config file written by a heredoc under `umask 077` and `shred -u`'d on EXIT.

Helper written for this probe (contains **no** secrets, sources the 0600 env file):
`/home/gaetan/tars/wf4/slackapi.sh` (mode 700, created 19:01:13).

```sh
#!/bin/sh
# WF4 probe helper. Personal xoxc/xoxd read from ~/.hermes/.env, handed to curl via a
# -K config file (heredoc) so no secret ever reaches argv or the transcript.
set -e
umask 077
CONF=$(mktemp)
trap 'shred -u "$CONF" 2>/dev/null || rm -f "$CONF"' EXIT
set -a
. "$HOME/.hermes/.env"
set +a
M=$1; shift
cat > "$CONF" <<EOF
cookie = "d=$SLACK_MCP_XOXD_TOKEN"
form = "token=$SLACK_MCP_XOXC_TOKEN"
url = "https://slack.com/api/$M"
EOF
for kv in "$@"; do echo "form = \"$kv\"" >> "$CONF"; done
curl -s -w "\nHTTP=%{http_code}" -K "$CONF"
```

---

## Probe 9 — slack-personal MCP: connect, real tool call, independent cross-check, lazy lookups, token survival

### 9a — `hermes mcp test slack` connects (19:01:48 → 19:01:49)

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes mcp test slack'
  Testing 'slack'...
  Transport: stdio → docker
  Auth: none
  ✓ Connected (877ms)
  ✓ Tools discovered: 18
    channels_list  channels_me  conversations_history  conversations_join
    conversations_leave  conversations_mark  conversations_replies
    conversations_search_messages  conversations_unreads  saved_clear_completed
    saved_list  saved_update  usergroups_create  usergroups_list  usergroups_me
    usergroups_update  usergroups_users_update  users_search
exit=0
```

877 ms cold connect — same order as WF3's 855 ms, and independent runtime proof that no eager
workspace crawl happens on startup (see 9d).

### 9b — real tool call through a chat turn (19:02:16 → 19:02:50)

Channel picked from the 40-channel list in `status/probes/wf3-s4-slack.md`: **`#help-help-tech`**,
ID `C0BLPCP0APN` (ID resolved out-of-band via `users.conversations`, because `--no-cache` disables
`#name` resolution — the documented D1 trade-off).

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); timeout 300 ~/.local/bin/hermes chat -Q -q \
  "Use your Slack MCP tools to fetch the single most recent message in the Slack channel whose ID is \
   C0BLPCP0APN (#help-help-tech). Report exactly: the message text, its ts value, and the sender user ID. \
   Do not post or send anything."'
rtk: hermes plugin warning: rtk binary not found in PATH; Hermes hook not registered

session_id: 20260807_190217_789dfc
Message text: Déso pas déso
ts: 1786124296.114459
Sender user ID: U08BDJAMSRZ
exit=0
```

Tool-execution proof from `~/.hermes/logs/agent.log` (not the model's self-report):

```
1268:2026-08-07 19:02:17,923 INFO tools.mcp_tool: MCP server 'slack' (stdio): registered 20 tool(s):
      mcp__slack__channels_list, mcp__slack__channels_me, mcp__slack__conversations_history,
      mcp__slack__conversations_join, mcp__slack__conversations_leave, mcp__slack__conversations_mark,
      mcp__slack__conversations_replies, mcp__slack__conversations_search_messages,
      mcp__slack__conversations_unreads, mcp__slack__saved_clear_completed, mcp__slack__saved_list,
      mcp__slack__saved_update, mcp__slack__usergroups_create, mcp__slack__usergroups_list,
      mcp__slack__usergroups_me, mcp__slack__usergroups_update, mcp__slack__usergroups_users_update,
      mcp__slack__users_search, mcp__slack__list_resources, mcp__slack__read_resource
1292:2026-08-07 19:02:43,462 INFO [20260807_190217_789dfc] agent.tool_executor:
      tool mcp__slack__conversations_history completed (0.49s, 321 chars)
```

Session id in the log line matches the session id printed by the chat turn — the answer came from a
real `mcp__slack__conversations_history` call, 0.49 s, 321 chars of tool output.
No `"Codex response remained incomplete after 3 continuation attempts"` — the turn succeeded first try,
no retry needed.

### 9c — independent cross-check, direct `conversations.history` (19:02:57)

Not via any claude.ai connector: direct Slack Web API with the personal `xoxc`/`xoxd` from
`~/.hermes/.env`, run 7 s after the chat turn returned.

```
$ ssh gaetan@192.168.0.9 '~/tars/wf4/slackapi.sh conversations.history "channel=C0BLPCP0APN" "limit=1" | python3 -c …'
http=200
ok= True error= None
ts=1786124296.114459
utc=2026-08-07T17:38:16.114459+00:00
user=U08BDJAMSRZ
text='Déso pas déso'
```

| Field | Tars via MCP (9b) | Direct API (9c) | Match |
|---|---|---|---|
| text | `Déso pas déso` | `Déso pas déso` | **yes** |
| ts | `1786124296.114459` | `1786124296.114459` | **yes** (to the microsecond) |
| user | `U08BDJAMSRZ` | `U08BDJAMSRZ` | **yes** |

### 9d — lazy lookups confirmed (config + runtime log + container census)

**Config, read-only grep** (19:00:45 and re-grepped post-call at 19:04:22, identical):

```
$ ssh gaetan@192.168.0.9 'grep -n -A12 "^  slack:$" ~/.hermes/config.yaml'
175:  slack:
176-    command: docker
177-    args:
178-    - run
179-    - -i
180-    - --rm
181-    - --env-file
182-    - /home/gaetan/tars/slack-mcp/.env
183-    - ghcr.io/korotovsky/slack-mcp-server:v1.3.0
184-    - --transport
185-    - stdio
186-    - --no-cache
```

**Runtime**, `~/.hermes/logs/mcp-stderr.log` (redaction filter applied on the way out; no token
appeared) — one line per server start, including all three of my probe invocations:

```
{"level":"info","timestamp":"2026-08-07T18:53:55Z","message":"Cache loading disabled via --no-cache flag",…}  ← gateway start
{"level":"info","timestamp":"2026-08-07T19:01:49Z","message":"Cache loading disabled via --no-cache flag",…}  ← 9a mcp test
{"level":"info","timestamp":"2026-08-07T19:02:06Z","message":"Cache loading disabled via --no-cache flag",…}  ← probe 10c chat
{"level":"info","timestamp":"2026-08-07T19:02:17Z","message":"Cache loading disabled via --no-cache flag",…}  ← 9b chat
$ grep -ci "users.list" ~/.hermes/logs/mcp-stderr.log
0
```

Zero `users.list` bulk-crawl mentions in the whole log — korotovsky #86 (the eager user cache that
invalidates `xoxc`/`xoxd`) is not being triggered.

**Container census.** `docker ps -a` immediately before (19:01:48) and after (19:03:40, 19:04:22)
the probe calls is byte-identical apart from the uptime counter:

```
=== BEFORE (19:01:48) ===
hardcore_wiles  df0d6781d03f                                  Up 7 minutes
clever_wright   ghcr.io/korotovsky/slack-mcp-server:v1.3.0     Up 7 minutes
=== AFTER (19:03:40) ===
hardcore_wiles  df0d6781d03f                                  Up 9 minutes
clever_wright   ghcr.io/korotovsky/slack-mcp-server:v1.3.0     Up 9 minutes
=== FINAL (19:04:22) ===
hardcore_wiles  df0d6781d03f                                  Up 10 minutes
clever_wright   ghcr.io/korotovsky/slack-mcp-server:v1.3.0     Up 10 minutes
```

**No container from my calls lingers.** The three transient `--rm` stdio containers spawned at
19:01:49 / 19:02:06 / 19:02:17 all exited with their sessions — the container count never rose.

The single surviving `slack-mcp-server` instance is **gateway-owned and pre-dates the probes**:

```
$ docker inspect --format '{{.Name}} created={{.Created}} image={{.Config.Image}}' clever_wright hardcore_wiles
/clever_wright   created=2026-08-07T18:53:54.631358685Z  image=ghcr.io/korotovsky/slack-mcp-server:v1.3.0
/hardcore_wiles  created=2026-08-07T18:53:54.634440791Z  image=mcp/notion@sha256:df0d6781d03f…

$ ps -o pid=,ppid=,lstart=,comm= -p 61969,61980,61982,61984,61990
  61969   27136 Fri Aug  7 18:53:52 2026 hermes    ← hermes-gateway MainPID
  61980   61969 Fri Aug  7 18:53:54 2026 python    ← tools/mcp_stdio_watchdog.py
  61982   61969 Fri Aug  7 18:53:54 2026 python    ← tools/mcp_stdio_watchdog.py
  61984   61982 Fri Aug  7 18:53:54 2026 docker    ← mcp/notion
  61990   61980 Fri Aug  7 18:53:54 2026 docker    ← slack-mcp-server (--env-file … --no-cache)
```

Created 2 s after the gateway became active, parented by the gateway's own stdio watchdog: this is
the gateway holding its stdio MCP servers open for its lifetime (by design), not a probe leftover
and not an eager-cache daemon — its own startup line at 18:53:55 says
`Cache loading disabled via --no-cache flag`. Recorded as a note, not a defect.

### 9e — token survival: `auth.test` re-run after the MCP call (19:04:13)

Credentials sourced from `~/.hermes/.env` (per task), passed via `-K` config, never echoed.

```
$ ssh gaetan@192.168.0.9 '~/tars/wf4/slackapi.sh auth.test | python3 -c …'
http_status=200
ok=True
error=None
user=gaetan.cathelain
user_id=U08BDJAMSRZ
team=Mobile Club
team_id=T7V1UGJ82
url=https://mobileclub-squad.slack.com/
```

`ok:true`, no `invalid_auth`. The `xoxc`/`xoxd` pair survived the live MCP tool call — matches the
WF3 A4/S4 identity exactly (`U08BDJAMSRZ` / `T7V1UGJ82`).

**Probe 9 = PASS** (connect ✓ · real `mcp__slack__conversations_history` call ✓ · independent
cross-check exact match ✓ · `--no-cache` in config + runtime log, zero eager crawl ✓ · no probe
container lingers ✓ · `auth.test ok:true` after ✓).

---

## Probe 10 — model answer proves gpt-5.6-sol

All three legs required together; all three ran at 19:01:59 → 19:02:10.

### 10a — `hermes auth status openai-codex`

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes auth status openai-codex'
openai-codex: logged in
exit=0
```

### 10b — config (read-only `config get`, no file written)

```
$ ~/.local/bin/hermes config get model.default
gpt-5.6-sol
exit=0
$ ~/.local/bin/hermes config get model.base_url
https://chatgpt.com/backend-api/codex
exit=0
```

### 10c — live single-tool-free round-trip (19:02:04 → 19:02:10)

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); timeout 300 ~/.local/bin/hermes chat -Q -q "Reply with exactly one word: sol"'
rtk: hermes plugin warning: rtk binary not found in PATH; Hermes hook not registered

session_id: 20260807_190205_ec9faa
sol
exit=0
```

Exactly `sol`, 6 s wall clock, exit 0. No `"Codex response remained incomplete"` on either chat turn
in this probe run.

**Probe 10 = PASS** (logged in ✓ · `gpt-5.6-sol` + `https://chatgpt.com/backend-api/codex` ✓ · live
round-trip answered ✓ — all three, per spec).

---

## Untouched-state proof (19:04:22)

```
$ stat -c "%n mtime=%y mode=%a" ~/.hermes/config.yaml ~/.hermes/.env ~/tars/slack-mcp/.env
/home/gaetan/.hermes/config.yaml  mtime=2026-08-07 18:08:43.859587213 +0000  mode=600
/home/gaetan/.hermes/.env         mtime=2026-08-07 18:53:46.680542160 +0000  mode=600
/home/gaetan/tars/slack-mcp/.env  mtime=2026-08-07 17:50:03.173328234 +0000  mode=600
```

All three mtimes predate this probe window (19:00:07). Nothing in config or credentials was written.
`systemctl --user` was used read-only (`is-active`, `show`) — no unit lifecycle command issued.

## Observations for the orchestrator (out of my probe scope, do not treat as my FAILs)

1. **rtk plugin warning on every `hermes chat`**: `rtk: hermes plugin warning: rtk binary not found
   in PATH; Hermes hook not registered` (19:02:04 and 19:02:16). Probe 11's rtk leg asks for
   `rtk gain --history` to show a hermes-proxied entry — with the hook unregistered under a
   non-interactive shell that leg is likely to fail. **Lane-B (B5) bucket**, flagged for whoever owns
   probe 11.
2. `--no-cache` disables `channels_list`; channel IDs had to be resolved out-of-band. Already the
   documented D1 trade-off (WF3 S4), no action.
3. The gateway keeps one `slack-mcp-server` and one `mcp/notion` container up for its lifetime.
   Expected stdio-MCP behaviour; anyone re-running the "no container lingers" check should compare
   the census before/after rather than expect an empty `docker ps -a`.
