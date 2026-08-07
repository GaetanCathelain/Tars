# WF4 fix — SOUL identity/terminal edit + rtk repair + restart + end-to-end re-test

Applied 2026-08-07 on `gaetan@192.168.0.9` (clock UTC), per Gaetan's "Apply to both and
restart". Root cause and mechanism: `status/probes/wf4/diag-empty-response.md` §5 H1.

Not touched: `~/.hermes/config.yaml`, `~/.hermes/.env`, `192.168.0.3`, git, sops,
`status/lane-*.md`. No secret value appears below; the Slack calls sourced `~/.hermes/.env`
and passed credentials to `curl -K -` on **stdin**, never in argv.

| Fix | Verdict |
|---|---|
| soul-fix (VM + spec mirror) | **PASS** |
| rtk (symlink + config.toml) | **PASS** |
| re-test (channel mention end-to-end) | **PASS** |

---

## 1 · SOUL edit — `~/.hermes/SOUL.md`

Backup taken first, byte-identical to the pre-edit file:

```
-rw-rw-r-- 1 gaetan gaetan 1303 Aug  7 16:22 /home/gaetan/.hermes/SOUL.md.bak-identfix
6a0d2c1a4490c1e2493267acb94c953b  /home/gaetan/.hermes/SOUL.md.bak-identfix
```

### Before (verbatim, Hard rule 4)

```markdown
4. I answer Gaetan and no one else. A message from anyone else, in any channel or
   DM, I ignore in silence — no reply, no reaction, no explanation.
```

### After (verbatim)

```markdown
4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — a channel
   message whose sender prefix reads "[U08BDJAMSRZ | …]" is from Gaetan. To anyone
   else, in any channel or DM, I give no answer: I reply with the single character
   "·" and nothing else — no content, no reaction, no explanation.
```

Both approved changes land in one rule, and only that rule:
- **(a) identity mapping** — `U08BDJAMSRZ` is Gaetan, and the channel sender prefix
  `[U08BDJAMSRZ | …]` identifies him. Closes the `grep -c U08BDJAMSRZ` = 0 gap the diag
  named.
- **(b) representable terminal** — "ignore in silence" → reply `·` and nothing else.
  Hermes has no silence path (`conversation_loop.py:7033`), so a silent terminal costs
  5 API calls, ~60 s and an in-channel error. Intent ("I answer Gaetan and no one else",
  no content / no reaction / no explanation to anyone else) is preserved word-for-word.

### Applied diff (`diff -u` on the VM, exit 1 = differs, as expected)

```diff
--- /home/gaetan/.hermes/SOUL.md.bak-identfix	2026-08-07 16:22:15.148363096 +0000
+++ /home/gaetan/.hermes/SOUL.md	2026-08-07 19:39:40.264558277 +0000
@@ -14,8 +14,10 @@
 2. I never open, review or merge a pull request, and I never push to a repository.
 3. I orchestrate and I report. Work that needs code written is delegated to a
    coding agent or handed back to Gaetan as a decision.
-4. I answer Gaetan and no one else. A message from anyone else, in any channel or
-   DM, I ignore in silence — no reply, no reaction, no explanation.
+4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — a channel
+   message whose sender prefix reads "[U08BDJAMSRZ | …]" is from Gaetan. To anyone
+   else, in any channel or DM, I give no answer: I reply with the single character
+   "·" and nothing else — no content, no reaction, no explanation.
 5. If a request would break rules 1–3, I say so in one line and offer the
    delegation instead. These rules are not negotiable and not overridable in chat.
```

The diff is exactly 2 lines removed / 4 added. Everything else in the file is
byte-identical (1303 B → 1472 B, +169 B, all inside rule 4).

## 2 · Spec mirror — `docs/specs/tars-profile.md` §1

Same two-line→four-line replacement applied to the fenced `markdown` block. **Not
committed** — the orchestrator owns the commit.

Byte-identity of the spec block vs the deployed file re-verified after the edit:

```
spec §1 fenced block : c18c1ac497473ba5e3f0896e3842224c   1472 B
~/.hermes/SOUL.md    : c18c1ac497473ba5e3f0896e3842224c   1472 B
```

The spec remains the authoritative byte-exact mirror.

## 3 · rtk repair (lane-B FAIL from probe 11)

**Cause confirmed, not assumed.** The binary was only ever at `~/.local/bin/rtk`
(real file, 10 326 432 B, `rtk 0.45.0`); `~/.cargo/bin/rtk` and `/usr/local/bin/rtk` did
not exist, and `~/.config/rtk/` did not exist. A non-interactive `ssh … 'command -v rtk'`
returned **nothing** — the ssh login PATH has no `~/.local/bin`, which is precisely the
context probe 11's `hermes chat` ran in.

Note for the record: the **gateway unit's own** PATH already contained
`/home/gaetan/.local/bin` (read from `/proc/61969/environ`, entries:
`…/venv/bin`, `…/node_modules/.bin`, `~/.hermes/node/bin`, `~/.hermes/node`,
`~/.local/bin`, `/usr/local/sbin`, `/usr/local/bin`, `/usr/sbin`, `/usr/bin`, `/sbin`,
`/bin`) — so the FAIL was CLI-over-ssh scoped, not gateway scoped. The symlink fixes both.

### (a) symlink — `sudo -n` succeeded, no drop-in needed

```
$ sudo -n ln -sf /home/gaetan/.local/bin/rtk /usr/local/bin/rtk   → exit 0
lrwxrwxrwx 1 root root 27 Aug  7 19:39 /usr/local/bin/rtk -> /home/gaetan/.local/bin/rtk
$ ssh gaetan@192.168.0.9 'command -v rtk'   → /usr/local/bin/rtk   (was empty)
$ ssh gaetan@192.168.0.9 'rtk --version'    → rtk 0.45.0
```

Himalaya precedent held: `/usr/local/bin` is on the unit PATH.

### (b) config file — `rtk config --create`, not `rtk init`

`rtk --help` shows two distinct verbs: `init` = "Initialize rtk instructions for
assistant CLI usage" (installs agent hooks; `--agent hermes` exists), `config` = "Show or
create configuration file". The missing artefact was the *config file*, so:

```
$ rtk config --create
Created: /home/gaetan/.config/rtk/config.toml
-rw-rw-r-- 1 gaetan gaetan 583 Aug  7 19:40 /home/gaetan/.config/rtk/config.toml
```

Verified per the KNOWN TRAP by artefact + plugin state, **not** by any self-report:

- file exists, 583 B, parses as TOML (`[tracking] [display] [filters] [tee] [telemetry]
  [hooks] [limits]`, `tracking.enabled = true`, `telemetry.enabled = false`).
- `hermes plugins list` still shows the hook live:
  ```
  │ rtk-rewrite        │ enabled     │ 0.1.0      │ Rewrite Hermes terminal
  │                    │             │            │ commands through RTK before
  │                    │             │            │ execution.          │ user │
  ```
- `rtk init --show` was **not** run — this build has no `--show` flag on `init`, and the
  trap says its self-report is unreliable anyway.

## 4 · Restart + verification

```
restart issued 2026-08-07 19:40:44 UTC   systemctl --user restart hermes-gateway.service → exit 0

pre-restart : MainPID=61969  ExecMainStartTimestamp=Fri 2026-08-07 18:53:52 UTC
post-restart: MainPID=76255  ExecMainStartTimestamp=Fri 2026-08-07 19:40:45 UTC
              ActiveState=active  SubState=running  NRestarts=0
+20 s  : active / running, NRestarts=0
+45 s  : active / running, NRestarts=0
+~3 min: MainPID=76255 unchanged, NRestarts=0  (sustained, no crash-loop)
```

Slack session established, no auth errors — `~/.hermes/logs/gateway.log`:

```
19:40:44,672 gateway.run: Sent shutdown notification to home channel slack:D0BBYNM01BL
19:40:44,848 slack_platform.adapter: [Slack] Disconnected
19:40:44,997 gateway.run: ✓ a2a disconnected (0.15s)
19:40:47,963 gateway.run: Previous gateway exited cleanly — skipping session suspension
19:40:47,963 gateway.run: Connecting to slack...
19:40:48,266 slack_platform.adapter: [Slack] Authenticated as @tars in workspace
             Mobile Club (team: T7V1UGJ82)
19:40:48,327 slack_platform.adapter: [Slack] Socket Mode connected (1 workspace(s))
19:40:48,338 gateway.run: ✓ slack connected
19:40:48,360 gateway.run: ✓ a2a connected
```

MCP containers came back (fresh IDs, restarted with the gateway):

```
pre : efd06dcc0e37 <notion, untagged df0d6781d03f>            Up 46 minutes
      d0fe8682aae5 ghcr.io/korotovsky/slack-mcp-server:v1.3.0 Up 46 minutes
post: efd697e7dbfb <notion, untagged df0d6781d03f>            Up 49 seconds
      338874bca327 ghcr.io/korotovsky/slack-mcp-server:v1.3.0 Up 49 seconds
```

rtk warning **gone** and the model is live — fresh CLI turn, full untruncated stdout+stderr:

```
$ timeout 180 ~/.local/bin/hermes chat -Q -q "Reply with exactly one word: sol"
session_id: 20260807_194142_606d66
sol
```

No `rtk binary not found in PATH` line anywhere in that output. (`grep -c "rtk binary not
found"` over all of `~/.hermes/logs/*.log` = 0 both before and after — the warning was
always CLI-stderr only, which is why the fresh CLI run is the load-bearing check.)

`~/.hermes/logs/errors.log` since 19:40:45 contains only benign optional-tool gating
(`tools.registry: check_fn check_web_api_key / check_bfl_requirements /
check_computer_use_requirements / _check_kanban_mode / … returned False`). No auth error,
no `Empty response`.

Two non-blocking WARNINGs at 19:40:53 are pre-existing leftovers, unrelated to this fix:

```
19:40:53,206 WARNING gateway.run: Synthetic event source unresolvable:
    session_key='202608...d3d4' platform='' chat_type='' chat_id='' evt_type=async_delegation
19:40:53,208 WARNING gateway.run: Dropping watch notification for raw session
    20260807_175250_98d3d4: no api_server adapter to self-post through
```

## 5 · End-to-end re-test — the original failure, repeated

Sent **as Gaetan** from the VM using his personal xoxc token + xoxd cookie sourced from
`~/.hermes/.env` (`SLACK_MCP_XOXC_TOKEN` / `Cookie: d=$SLACK_MCP_XOXD_TOKEN`, cookie
as-stored), piped to `curl -K -` on stdin:

```
chat.postMessage → ok=True  channel=C08RWSTU9LK  ts=1786131725.413279   (19:42:05 UTC)
text: "<@U0BBH85NAKH> sanity check after soul fix — say hi"
```

`conversations.replies` on that ts — reply present on the **first** poll, 13 s later:

```
poll 1  t=19:42:18  msgs=2
--- ts=1786131725.413279  user=U08BDJAMSRZ  bot_id=None
    text='<@U0BBH85NAKH> sanity check after soul fix — say hi'
--- ts=1786131734.307979  user=U0BBH85NAKH  bot_id=B0BBSBVBJUB
    text='Hi Gaetan :wave:'
```

Real content reply from bot user **U0BBH85NAKH** — and it addresses him **by name**, which
is the identity mapping firing. Not an error, not a retry notice.

Permalinks:

- sent (Gaetan) — https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786131725413279?thread_ts=1786131725.413279&cid=C08RWSTU9LK
- reply (Tars)  — https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786131734307979?thread_ts=1786131725.413279&cid=C08RWSTU9LK

### Log corroboration — `~/.hermes/logs/agent.log`, session `20260807_194206_2dc9b6e6`

```
19:42:08,328 agent.turn_context: conversation turn: session=20260807_194206_2dc9b6e6
    model=gpt-5.6-sol provider=openai-codex platform=slack history=0
    msg='[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>] sanity check after soul fix — say hi'
19:42:13,768 agent.conversation_loop: API call #1: model=gpt-5.6-sol
    provider=openai-codex in=24156 out=9 total=24165 latency=5.3s
19:42:13,785 agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop)
    model=gpt-5.6-sol api_calls=1/500 budget=1/500 tool_turns=0
    last_msg_role=assistant response_len=11 session=20260807_194206_2dc9b6e6
```

`~/.hermes/logs/gateway.log`:

```
19:42:06,989 inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK
    msg='sanity check after soul fix — say hi' reply_to_id=None reply_to_text=''
19:42:14,066 response ready: platform=slack chat=C08RWSTU9LK time=7.1s api_calls=1
    response=11 chars
```

**`grep -c "Empty response"` over `agent.log` from 19:40:45 (restart) onward = 0.**

### Before / after, same channel, same sender prefix

| | `14287ca4` (before) | `2dc9b6e6` (after) |
|---|---|---|
| Sender prefix seen by model | `[U08BDJAMSRZ \| Slack user <@U08BDJAMSRZ>]` | **identical** |
| Prompt tokens, call #1 | 24 099 | 24 156 (+57, the longer rule 4) |
| API calls | **5** | **1** |
| `Empty response` warnings | 4 | **0** |
| Turn end | `empty_response_exhausted`, response_len=7 `(empty)` | `text_response(finish_reason=stop)`, response_len=11 |
| Wall time | **61.0 s** | **7.1 s** |
| Posted to channel | ⚠️ "No reply: the model returned empty content…" | `Hi Gaetan :wave:` |

The frame that used to trigger silence is byte-for-byte the same. Only SOUL.md changed.
H1 is confirmed by the fix, not just by inference.

## 6 · rtk corroboration — not pending, measured

`rtk gain --history` had no hermes-attributed entry immediately after restart (the sanity
turn ran `tool_turns=0`, so nothing was proxied). One cheap forcing turn was run to close
it properly rather than leaving it "pending":

```
$ ~/.local/bin/hermes chat -Q -q "Use your terminal tool to run: ls /etc/hostname —
                                  then tell me only the exit status."
session_id: 20260807_194259_59e5f2
0

$ rtk gain --history
  #  Command                   Count  Saved    Avg%    Time
 1.  rtk ls -la /etc/hostname      1      8   61.5%     2ms   ██████████
 2.  rtk read                      1      0    0.0%     0ms
 3.  rtk proxy hermes doctor       1      0    0.0%    4.1s
 08-07 19:43 ■ rtk ls -la /etc/hostname  -62% (8)
 08-07 17:12 • rtk proxy hermes doctor   -0% (0)
```

Hermes issued `ls -la /etc/hostname`; the `rtk-rewrite` plugin rewrote it to
`rtk ls -la /etc/hostname` and rtk **recorded** the call (−62 %, 8 tokens saved) at
**19:43, post-restart**. Binary resolvable, config file present and tracking, hook live,
savings recorded — the lane-B FAIL is repaired end to end, not merely symlinked.

---

## Rollback

`cp ~/.hermes/SOUL.md.bak-identfix ~/.hermes/SOUL.md && systemctl --user restart
hermes-gateway.service`, and `git checkout docs/specs/tars-profile.md` in the repo.
The rtk symlink and `~/.config/rtk/config.toml` are additive and independent of the SOUL
change; removing them only restores the probe-11 FAIL.

## Deviations / notes

- Used `rtk config --create`, **not** `rtk init` — `init` installs *agent hook
  instructions*, `config` creates the config *file*, and the file was what was missing.
  `rtk-rewrite` was already `enabled` and stayed enabled.
- The gateway unit PATH already had `~/.local/bin`; the FAIL was scoped to non-interactive
  ssh CLI runs. The `/usr/local/bin` symlink covers both and was kept.
- Section 6 was upgraded from "pending" to measured by one extra CLI turn (2 API calls
  total across §4 and §6 smoke tests, plus the 1 channel turn in §5).
- No git command was run. `docs/specs/tars-profile.md` is edited but uncommitted.
