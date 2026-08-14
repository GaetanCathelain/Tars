# GCN-50 Q10 — `hermes kanban notify-subscribe --thread-id`: real agent turn or raw post?

**Date:** 2026-08-14 (VM clock UTC) · **Host:** Tars VM `192.168.0.9`, Hermes v0.20.0
**Probe run by:** Claude Code session on cooper · **Board:** `default`
**Verdict: RAW-POST-IN-THREAD**

A `notify-subscribe --thread-id` delivery is a **template notification posted by the
gateway's notifier**, correctly rooted in the target Slack thread. **No agent turn ran.**
The agent-wake path exists but is gated on `task.session_id`, which `hermes kanban create`
cannot set from the CLI.

---

## 1. CLI facts (from `--help` on the VM, not memory)

`~/.local/bin/hermes kanban --help` — subcommand `notify-subscribe`:

> `notify-subscribe    Subscribe a gateway source to a task's terminal events (used by /kanban subscribe in the gateway adapter)`

```
usage: hermes kanban notify-subscribe [-h] --platform PLATFORM --chat-id CHAT_ID
                                      [--chat-type CHAT_TYPE]
                                      [--thread-id THREAD_ID]
                                      [--user-id USER_ID]
                                      [--notifier-profile NOTIFIER_PROFILE]
                                      task_id
options:
  --chat-type CHAT_TYPE     dm / group / channel (used by wake routing)
  --notifier-profile NOTIFIER_PROFILE
                            Profile gateway that owns/delivers this subscription
                            (default: active profile)
```

```
usage: hermes kanban notify-unsubscribe [-h] --platform PLATFORM --chat-id CHAT_ID
                                        [--thread-id THREAD_ID] task_id
usage: hermes kanban notify-list [-h] [--json] [task_id]
usage: hermes kanban complete [-h] [--result RESULT] [--summary SUMMARY]
                              [--metadata METADATA] task_ids [task_ids ...]
usage: hermes kanban archive [-h] [--rm PURGE_IDS [PURGE_IDS ...]] [task_ids ...]
```

`hermes kanban create` (relevant flags only): `--body`, `--assignee`, `--created-by`,
`--initial-status {blocked,running}`, `--json`. **There is no `--session` flag on `create`**
(`list` has `--session` as a *filter* only: "Filter by originating chat/agent session id
(set on tasks created from inside an ACP loop)").

`~/.local/bin/hermes send --help`:

```
usage: hermes send [-h] [-t TARGET] [-f PATH] [-s LINE] [-l] [-q] [--json] [message]
  -t TARGET   Delivery target. Format: 'platform' (home channel), 'platform:chat_id',
              'platform:chat_id:thread_id', or 'platform:#channel-name'.
```

Note `hermes send` explicitly documents itself as **"no LLM, no agent loop"** — it is not a
turn under any flag, and its `platform:chat_id:thread_id` form is the manual thread-post path.

---

## 2. Pre-state

```
2026-08-14T09:40:25Z
Current board: default   DB: /home/gaetan/.hermes/kanban.db   Tasks: 2 total (archived=2)
hermes kanban notify-list --json  ->  []
systemctl --user show hermes-gateway -p ActiveEnterTimestamp -p ActiveState
  ActiveState=active
  ActiveEnterTimestamp=Thu 2026-08-13 15:24:11 UTC
log line counts @09:41:42Z: agent.log=3736  gateway.log=2567  errors.log=29
```

Gateway did **not** restart during the probe (same `ActiveEnterTimestamp` throughout).

**Confound present and isolated:** Gaetan was actively using Tars in the same DM during the
probe window (sessions `20260814_092125_cf9c9f23`, `20260814_093833_70218516`,
`20260814_094023_c54246b9`, all started **before** the trigger). Every agent.log turn in the
window belongs to one of those pre-existing sessions. See §5.

---

## 3. Actions

```
09:40:36Z  hermes send -t slack:D0BBYNM01BL --json '[PROBE gcn50-q10] kanban thread-rooting smoke test anchor — ignore'
           -> {"success": true, "platform": "slack", "chat_id": "D0BBYNM01BL",
               "message_id": "1786700436.819599", "mirrored": true}
           ANCHOR ts = 1786700436.819599

09:41:42Z  hermes kanban create '[PROBE gcn50-q10] kanban thread-rooting smoke' \
             --body 'probe card; do not action' --initial-status blocked \
             --created-by probe-gcn50 --json
           -> id=t_77b562e8, status=blocked, assignee=null, **session_id=null**
           (--initial-status blocked chosen so the dispatcher never spawns a worker on
            Gaetan's live agent — the probe tests notification, not execution)

09:41:50Z  hermes kanban notify-subscribe t_77b562e8 --platform slack \
             --chat-id D0BBYNM01BL --chat-type dm \
             --thread-id 1786700436.819599 --user-id U08BDJAMSRZ
           -> "Subscribed slack:D0BBYNM01BL:1786700436.819599 to t_77b562e8"

           hermes kanban notify-list t_77b562e8 --json
           -> [{"task_id":"t_77b562e8","platform":"slack","chat_id":"D0BBYNM01BL",
                "chat_type":"dm","thread_id":"1786700436.819599","user_id":"U08BDJAMSRZ",
                "notifier_profile":"default","delivery_metadata":{},
                "created_at":1786700510,"last_event_id":15}]

09:41:50Z  TRIGGER: hermes kanban complete t_77b562e8 \
             --result '[PROBE gcn50-q10] probe complete — no action taken'
           -> "Completed t_77b562e8"
```

---

## 4. Observation A — Slack (delivery landed IN the thread)

`slack_read_thread(channel_id=D0BBYNM01BL, message_ts=1786700436.819599)` — verbatim:

```
=== THREAD PARENT MESSAGE ===
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 11:40:36 CEST
Message TS: 1786700436.819599
[PROBE gcn50-q10] kanban thread-rooting smoke test anchor — ignore

=== THREAD REPLIES (1 total) ===

--- Reply 1 of 1 ---
From: Tars (U0BBH85NAKH)
Time: 2026-08-14 11:41:54 CEST
Message TS: 1786700514.626349
:heavy_check_mark: [default] Kanban t_77b562e8 done — [PROBE gcn50-q10] kanban thread-rooting smoke
[PROBE gcn50-q10] probe complete — no action taken
```

- **Thread rooting WORKS.** `--thread-id` put the reply inside the anchor thread (11:41:54
  CEST = **09:41:54Z**, 4 s after the trigger). Nothing landed at channel top level.
- The body is a **fixed template**: `✔ [<board>] Kanban <id> <status> — <title>` + the
  literal `--result` string echoed back. No prose, no reasoning, no tool use — it is the
  `--result` text verbatim, not a model rendering of it.
- After delivery the subscription auto-removed (task reached a final status):
  `hermes kanban notify-list t_77b562e8 --json -> []`.

## 5. Observation B — VM (no agent turn ran)

Every agent.log line in the 09:41:50–09:41:59 window (grep of the raw timestamps):

```
2026-08-14 09:41:50,661 INFO [20260814_094023_c54246b9] agent.tool_executor: tool terminal completed (0.79s, 110 chars)
2026-08-14 09:41:53,398 INFO [20260814_093833_70218516] agent.conversation_loop: API call #18: model=gpt-5.6-sol prov...
2026-08-14 09:41:53,423 INFO [20260814_093833_70218516] agent.tool_executor: tool cronjob completed (0.01s, 760 chars)
2026-08-14 09:41:58,401 INFO [20260814_093833_70218516] agent.conversation_loop: API call #19: model=gpt-5.6-sol prov...
2026-08-14 09:41:58,417 INFO [20260814_093833_70218516] agent.tool_executor: tool cronjob completed (0.00s, 19292 chars)
```

Distinct sessions seen 09:40–09:43 (all pre-date the 09:41:50 trigger):
`20260814_092125_cf9c9f23`, `20260814_093356_03e3c7`, `20260814_093833_70218516`,
`20260814_093857_c11699`, `20260814_094023_c54246b9`.

- **No new session id** appears at or after 09:41:50.
- **No `agent.turn_context: conversation turn:`** line for the notification.
- **No model API call** attributable to the notification.
- gateway.log for 09:41:40–09:42:10 contains **only** the unrelated Gaetan turn finishing:
  `09:42:05,604 gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=212.3s api_calls=20 response=385 chars`
  (212.3 s runtime ⇒ that turn started ~09:38:33, long before the trigger).
- `grep -c "woke agent\|wakeup injection"` over gateway.log / agent.log / errors.log:
  **0 hits on 2026-08-14.** The one lifetime hit is from a week earlier (§6).

The notifier itself logs nothing at INFO on the plain text-ping path, so the *absence* of a
kanban line in gateway.log is consistent with a successful raw post — and the *presence* of
a `woke agent` line is what a real turn would have produced.

## 6. Root cause — the wake path exists but is gated on `task.session_id`

`/home/gaetan/.hermes/hermes-agent/gateway/kanban_watchers.py` (1523 lines) implements the
notifier. Two mutually-exclusive deliveries per terminal event:

```python
# L637-646
_WAKE_KINDS = ("completed", "gave_up", "crashed", "timed_out", "blocked")
_wake_kinds = {ev.kind for ev in d["events"] if ev.kind in _WAKE_KINDS}
_is_push_adapter = _adapter_push_ok(adapter)
_session_key = ""
if _wake_kinds:
    _session_key = getattr(task, "session_id", None) or ""
```

1. **Text ping** (what we measured). Slack is a push adapter, so `adapter.send(chat_id, msg,
   metadata=metadata)` fires, with (L516-517):
   ```python
   if sub.get("thread_id") and not metadata.get("thread_id"):
       metadata["thread_id"] = sub["thread_id"]
   ```
   ⇒ `--thread-id` is honoured on the raw post. This is the only delivery a CLI-created card gets.

2. **Agent wake** (L729-772), which *is* a genuine turn — `deliver_wake` injects a synthetic
   `MessageEvent` through `handle_message`:
   ```python
   if _is_push_adapter and _wake_kinds and _session_key:
       _source = SessionSource(platform=plat, chat_id=sub["chat_id"],
                               chat_type=_chat_type,
                               thread_id=sub.get("thread_id") or None,
                               user_id=sub.get("user_id"), profile=sub_profile or None)
       await deliver_wake(adapter, text=_synth, session_id=_session_key, source=_source)
       logger.info("kanban notifier: woke agent for %s on %s/%s profile=%s events=%s", ...)
   ```
   **`and _session_key`** is the gate. Our card printed `"session_id": null` at creation and
   still `null` at completion ⇒ `_session_key == ""` ⇒ the whole block is skipped.

`session_id` is set only for "tasks created from inside an ACP loop" (per `list --session`
help) — i.e. cards Tars creates itself with the kanban tool during a chat turn. `hermes
kanban create` exposes no flag to set it.

**Positive control, same DM, same code path:** the one lifetime `woke agent` line —

```
2026-08-07 20:30:14,803 INFO gateway.run: kanban notifier: woke agent for t_f525e54b
  on slack/D0BBYNM01BL profile=default events={'completed'}
```

and that task's record:

```
{'id': 't_f525e54b', 'title': 'WF5 kanban slack smoke card', 'status': 'archived',
 'created_by': 'worker', 'session_id': '20260807_202851_d8895c62'}
```

`created_by='worker'` + non-null `session_id` ⇒ wake fired. Our `created_by='probe-gcn50'` +
`session_id=null` ⇒ wake skipped. Same notifier, same DM, opposite outcome, single differing
input. That is the causal gate.

Note the SessionSource above carries `thread_id=sub["thread_id"]`, so a wake on a
thread-scoped subscription *should* run the turn in that thread — **untested**: no card in
this DB has ever had both a `session_id` and a `--thread-id` subscription, and forcing one
would mean spawning a live Tars turn while Gaetan was mid-conversation.

## 7. Verdict

**RAW-POST-IN-THREAD.**

`notify-subscribe --thread-id` reliably roots delivery in the target Slack thread (4 s
latency, exact thread, auto-unsubscribe on final status) — but the payload is a template
notification emitted by the gateway notifier, not an agent turn. For CLI-created cards the
agent-wake branch is unreachable by construction (`session_id` is `null` and `create` has no
flag to set it). A real turn only fires for cards a Hermes agent session created itself.

### Architecture consequence

- Want a **status ping in a specific thread** → this works today, exactly as advertised.
- Want a **real agent turn in a specific thread** → `hermes kanban create` from the CLI
  cannot get you one. Either (a) have Tars create the card in-session so `session_id` is
  stamped, then subscribe with `--thread-id` (plausible per source, **unverified end-to-end**),
  or (b) drive the turn a different way and use `hermes send -t slack:<chat>:<thread_ts>`
  for plain thread posts.
- No config change was needed and none was made — the CLI works as shipped.

## 8. Cleanup / footprint

```
hermes kanban archive t_77b562e8   -> "Archived t_77b562e8"
hermes kanban show t_77b562e8      -> {'id': 't_77b562e8', 'status': 'archived'}
hermes kanban notify-list --json   -> []
```

- No writes to `~/.hermes/config.yaml` or `~/.hermes/.env`. No `sops` invocation of any form.
- No secrets read, printed or stored. No gateway restart.
- Slack probe messages left in place per instructions: anchor `1786700436.819599` and
  notifier reply `1786700514.626349` in `D0BBYNM01BL`, both tagged `[PROBE gcn50-q10]`.
