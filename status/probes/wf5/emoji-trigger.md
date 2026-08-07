# WF5 probe — emoji / reaction triggers ("react :kanban: → Tars adds it")

Agent: emoji-trigger (read-only source + read-only Slack API). Date 2026-08-07,
VM clock UTC. **Nothing was modified. No Slack write method was called. No
reaction added or removed.**

Scope: emoji/reaction triggers only. Sibling agents own the message-path
mention gate, session/thread assembly, config snapshot, log forensics.

---

## TL;DR

**PARTIAL — the feature is fully implemented and shipped in the installed
Hermes; it is switched off at three independent places, none of which is a
code change.**

Hermes v0.20.0 ships a complete, documented `slack.reaction_triggers` feature:
`reaction_added`/`reaction_removed` Bolt listeners, an emoji allowlist mode
explicitly designed for "emoji-handoff workflows", correct thread-parent
resolution, and the allowlist/auth gate applied at the shared chokepoint.
What is missing on Tars:

1. Slack app has **no `reactions:read` scope** (verified live) ⇒ Slack will not
   deliver `reaction_added` over Socket Mode at all.
2. `reaction_added` is therefore not in the app's event subscriptions.
3. `reaction_triggers` is **unset** in `~/.hermes/config.yaml` and
   `SLACK_REACTION_TRIGGERS` is not in `~/.hermes/.env` ⇒ the handler acks and
   drops (`_slack_reaction_triggers()` returns `None`).

Agent-class is **not** a blocker: Slack's own AI-apps guide tells agent/assistant
apps to subscribe to `reaction_added`.

Fix = **Slack dashboard (1 scope + 1 event) + 2 config lines. Zero code.**
And per Slack's additive-scope model the existing `xoxb` token is **not**
invalidated ⇒ **no SOPS rotation** (verify with `auth.test` after re-install).

---

## 0. Resolved package path

`~/.local/bin/hermes` is a 4-line wrapper:

```
#!/usr/bin/env bash
unset PYTHONPATH
unset PYTHONHOME
exec "/home/gaetan/.hermes/hermes-agent/venv/bin/python" "/home/gaetan/.hermes/hermes-agent/hermes" "$@"
```

The package is an **editable install** — there is no copy under
`site-packages`; `__editable__.hermes_agent-0.20.0.pth` points back at the
source tree. All `file:line` citations below are relative to:

**`/home/gaetan/.hermes/hermes-agent/`** (on the VM, `192.168.0.9`).

Slack adapter: `plugins/platforms/slack/adapter.py` — 9100 lines, 402 KB.

---

## 1. Does the Hermes Slack adapter handle reaction events at all? — **YES**

`plugins/platforms/slack/adapter.py:1990-2002`:

```python
            # Forward reaction_added events through the normal message
            # pipeline (see _handle_slack_reaction). Skills that present
            # confirmation-style proposals ("react 👍 to proceed") then work
            # end-to-end. Registered explicitly so high-traffic channels do
            # not fill gateway.error.log with Slack Bolt "Unhandled request"
            # warnings.
            @self._app.event("reaction_added")
            async def handle_reaction_added(event, say):
                await self._handle_slack_reaction(event)

            @self._app.event("reaction_removed")
            async def handle_reaction_removed(event, say):
                await self._handle_slack_reaction(event, removed=True)
```

Handler: `plugins/platforms/slack/adapter.py:4773` `_handle_slack_reaction()`.

### Grep results table (package-wide, excluding `venv/`, `node_modules/`)

| Term | file:line | Classification |
|---|---|---|
| `reaction_added` | `plugins/platforms/slack/adapter.py:1996` | **INBOUND-TRIGGER** — Bolt listener registration |
| `reaction_removed` | `plugins/platforms/slack/adapter.py:2000` | **INBOUND-TRIGGER** — Bolt listener registration |
| `_handle_slack_reaction` | `plugins/platforms/slack/adapter.py:4773` | **INBOUND-TRIGGER** — synthesizes a message event, calls `_handle_slack_message` |
| `_slack_reaction_triggers` | `plugins/platforms/slack/adapter.py:4961` | INBOUND-TRIGGER config gate (`slack.reaction_triggers` / `SLACK_REACTION_TRIGGERS`) |
| `_slack_reaction_trigger_target` | `plugins/platforms/slack/adapter.py:4992` | INBOUND-TRIGGER routing (`slack.reaction_trigger_target`) |
| `_REACTION_EMOJI_MAP` | `plugins/platforms/slack/adapter.py:4754-4769` | INBOUND-TRIGGER — Slack name → unicode for 15 common emoji |
| `reaction_triggers` env plumbing | `plugins/platforms/slack/adapter.py:9020-9024` | INBOUND-TRIGGER config→env bridge |
| `reaction_trigger_target` env plumbing | `plugins/platforms/slack/adapter.py:9025-9027` | INBOUND-TRIGGER config→env bridge |
| `reactions_add(...)` | `plugins/platforms/slack/adapter.py:3729` | **OUTBOUND-ACTION** — `_add_reaction()`, the 👀 lifecycle ack |
| `reactions_remove(...)` | `plugins/platforms/slack/adapter.py:3739` | **OUTBOUND-ACTION** — `_remove_reaction()` |
| `SLACK_REACTIONS` gate | `plugins/platforms/slack/adapter.py:3755`, `9018-9019` | OUTBOUND-ACTION gate, **default `true`** |
| `app_mention` | `plugins/platforms/slack/adapter.py:1963` | INBOUND-TRIGGER (message path — sibling agent's scope) |
| `message_changed` | `plugins/platforms/slack/adapter.py:5265` | INBOUND — edited-message normalisation, unrelated to reactions |
| reaction hook contract | `gateway/platforms/base.py:3364` | INBOUND observer — "(currently the Slack adapter's ``reaction_added``/``reaction_removed``)" |
| manifest scopes/events | `hermes_cli/slack_cli.py:104-105` (`reaction_added`, `reaction_removed`), `:91` (`reactions:read`) | Manifest generator — see §5 |
| shipped user doc | `website/docs/user-guide/messaging/slack.md:641-684` | Documentation of the whole feature |
| `star_added` | *(no hits)* | Not handled |
| `reactions.get` / `reactions_get` | *(no hits)* | Never called |
| `@app.shortcut` / `.shortcut(` | *(no hits package-wide)* | **No message shortcut / message action exists** — see §8b |
| `message_action` | *(no hits)* | idem |
| `interactivity` | `hermes_cli/slack_cli.py:156` | Manifest key for Block Kit buttons only |

Other platforms have their own reaction paths (`plugins/platforms/photon/adapter.py:1265`,
`plugins/platforms/matrix/adapter.py:2044`, `plugins/platforms/feishu/adapter.py:3260`) —
listed for completeness, not in play for Tars.

---

## 2. The adapter's complete inbound event dispatch surface

Bolt registrations, in registration order — **order matters**: bolt dispatches
to the first matching listener, so the catch-all at the end never shadows a
named handler (`adapter.py:2033-2038` explains this).

| # | Event type | file:line | Handler |
|---|---|---|---|
| 1 | `message` | `adapter.py:1952` | `_handle_slack_message(event, body)` |
| 2 | `app_mention` | `adapter.py:1963` | `_handle_slack_message(event, body)` |
| 3 | `app_home_opened` | `adapter.py:1967` | `_handle_app_home_opened` |
| 4 | `app_context_changed` | `adapter.py:1971` | `_handle_app_context_changed` |
| 5 | `file_shared` | `adapter.py:1978` | `_handle_slack_file_shared` |
| 6 | `file_created` | `adapter.py:1982` | no-op ack |
| 7 | `file_change` | `adapter.py:1986` | no-op ack |
| 8 | **`reaction_added`** | `adapter.py:1996` | `_handle_slack_reaction(event)` |
| 9 | **`reaction_removed`** | `adapter.py:2000` | `_handle_slack_reaction(event, removed=True)` |
| 10 | `assistant_thread_started` | `adapter.py:2004` | `_handle_assistant_thread_lifecycle_event` |
| 11 | `assistant_thread_context_changed` | `adapter.py:2008` | `_handle_assistant_thread_lifecycle_event` |
| 12 | `re.compile(r".*")` catch-all | `adapter.py:2039` | debug-log + 200 ack, nothing else |

Non-event registrations: `@self._app.command(_slash_pattern)` (`adapter.py:2076`,
dead on Tars — Slack blocks slash on the Agent surface), and Block Kit
`self._app.action(...)` for approval / slash-confirm / feedback / clarify
buttons (`adapter.py:2092`, `2101`, `2103`, `2109`, `2112`) plus plugin-registered
action ids (`adapter.py:2156`).

**Definitive answer to "can a reaction reach Tars at all": yes at the code
level — a handler is registered and waiting.** What blocks it is the Slack
side. The adapter says so itself at `adapter.py:5244-5248`:

> `# bot-to-bot interop, allow_bots config, or SLACK_ALLOWED_USERS`
> `# drops can confirm whether the event actually arrived from Slack`
> `# (vs. being silently filtered upstream by the app's event`
> `# subscriptions — Socket Mode will not deliver events the app`
> `# manifest hasn't subscribed to). See #30091.`

---

## 3. Transport — **Socket Mode**

- `plugins/platforms/slack/adapter.py:27` — `from slack_bolt.adapter.socket_mode.async_handler import AsyncSocketModeHandler`
- `plugins/platforms/slack/adapter.py:1183-1196` — `_start_socket_mode_handler()` constructs `AsyncSocketModeHandler(...)` at `:1188` and runs it as a task.
- `plugins/platforms/slack/adapter.py:1287-1400` — reconnect/watchdog (`_restart_socket_mode`).
- Env: `SLACK_APP_TOKEN` (`xapp-…`, `connections:write`) is present in `~/.hermes/.env`, confirming Socket Mode.

**Consequence:** there is no request URL to configure, and no HTTP endpoint to
expose. Enabling a reaction trigger is therefore a **Slack-app-dashboard change
only** (scope + event subscription) — *not* a code change and *not* an
infrastructure change. Socket Mode still respects the event-subscription list:
Slack only pushes down the socket what the app is subscribed to.

---

## 4. Slack scopes — **`reactions:read` is NOT granted**

Live read-only probe (token sourced inside the remote shell, handed to curl via
`-K -` on stdin; never on argv, never printed):

```
ssh gaetan@192.168.0.9 'set -a; . ~/.hermes/.env; set +a;
  printf "header = \"Authorization: Bearer %s\"\n" "$SLACK_BOT_TOKEN" |
  curl -sS -D - -o /dev/null -K - "https://slack.com/api/auth.test" |
  grep -i "^x-oauth-scopes"'
```

`auth.test` → `{"ok":true, "url":"https://mobileclub-squad.slack.com/",
"team":"Mobile Club", "user":"tars", "team_id":"T7V1UGJ82", ...}` (bot token
`<redacted>`).

**Full granted bot scope list** (from the `x-oauth-scopes` response header,
17 scopes):

```
app_mentions:read   assistant:write   channels:history   channels:read
chat:write          commands          files:read         files:write
groups:history      groups:read       im:history         im:read
im:write            mpim:history      mpim:read          mpim:write
users:read
```

| Scope | Granted? | Consequence |
|---|---|---|
| `reactions:read` | **NO** | `reaction_added` cannot be subscribed to, and would not be delivered. **This is the hard blocker.** |
| `reactions:write` | **NO** | `_add_reaction()` (`adapter.py:3729`) fails silently — logged at DEBUG only (`adapter.py:3735`: *"Don't log as error — may fail if already reacted or missing scope"*). So Tars' 👀 lifecycle ack has been a no-op all along, invisibly. |
| `channels:history` | yes | `conversations.replies` works — used by the reaction thread-parent resolution (§6). |

Not guessed — read from the live Slack response header.

`apps.permissions.scopes.list` was not needed and was not called; the
`x-oauth-scopes` header is authoritative and cheaper. Reading the app manifest
(`apps.manifest.export`) requires a **configuration token**, not a bot token, so
the subscribed-events list cannot be read from the VM. It does not need to be:
Slack requires the backing scope for an event subscription
(https://docs.slack.dev/reference/events/reaction_added — *"Required OAuth
scope: `reactions:read`"*), and `reactions:read` is absent ⇒ `reaction_added` is
provably not subscribed.

---

## 5. Agent-class constraint — **NOT a blocker**

Slack's own AI-apps guide, in the section on event subscriptions for
agent/assistant apps:

> "You can also subscribe to [`reaction_added`](/reference/events/reaction_added)
> events to collect feedback based on reactions."

Source: **https://docs.slack.dev/ai/developing-ai-apps/**

Required events for the `agent` messaging experience are `app_home_opened`,
`app_context_changed`, `message.im`; for `assistant`, `assistant_thread_started`,
`assistant_thread_context_changed`, `message.im`. Those are the *minimum*, not a
closed set — the doc's own `reaction_added` suggestion proves an Agent-class app
is not restricted to the assistant event set. The only restriction the page
states is that **workspace guests cannot access apps with the Agents feature
enabled** — irrelevant here (Gaetan is a full member).

Corroborating evidence from the installed Hermes: its manifest generator puts
`reactions:read` and `reaction_added`/`reaction_removed` in the **base** scope and
event lists shared by *all three* messaging experiences (`none`, `assistant`,
`agent`), adding the agent-specific keys on top:

`hermes_cli/slack_cli.py:78-121`:

```python
    bot_scopes = [
        "app_mentions:read", "channels:history", "channels:read", "chat:write",
        "commands", "files:read", "files:write", "groups:history", "groups:read",
        "im:history", "im:read", "im:write", "mpim:history", "mpim:read",
        "reactions:read",          # ← line 91
        "users:read",
    ]
    bot_events = [
        "app_mention", "message.channels", "message.groups", "message.im",
        "message.mpim",
        "reaction_added",          # ← line 104
        "reaction_removed",        # ← line 105
    ]
    ...
    elif messaging_experience == "agent":
        features["agent_view"] = {...}
        bot_scopes.append("assistant:write")
        bot_events.extend(["app_context_changed", "app_home_opened"])
```

So the Hermes upstream ships an Agent-class manifest that *includes*
`reactions:read` + `reaction_added`. The Tars app was installed with a
hand-narrowed manifest (it has `assistant:write` and `commands` but not
`reactions:read`), which is why the scope is missing.

**Slash commands are the exception, not the rule.** Slack blocks slash commands
on the Agent chat surface (established, DECISION.md) even though the `commands`
scope is granted — that is a surface-level UI restriction on one interaction
type, and it does not generalise into "Agent apps may only subscribe to
assistant events". Nothing in Slack's docs restricts Agent-app event
subscriptions.

*Residual risk, stated honestly:* this cannot be 100 % proven without actually
adding the scope in the dashboard and watching Slack accept it. The doc quote
above is Slack telling AI-app developers to do exactly this, which is as strong
as evidence gets short of doing it. If Slack's manifest validator rejects
`reaction_added` on an app with `agent_view`, that will be visible immediately
in the dashboard, cost nothing, and change nothing — the app stays as it is.

---

## 6. The threading answer — **it lands in the right thread; already solved in code**

### Slack's documented rule

`chat.postMessage`, `thread_ts` argument
(https://docs.slack.dev/reference/methods/chat.postMessage):

> "Provide another message's `ts` value to make this message a reply. **Avoid
> using a reply's `ts` value; use its parent instead.**"

And on identifying a parent
(https://docs.slack.dev/messaging/retrieving-messages/):

> "Identify parent messages by comparing the `thread_ts` and `ts` values. If they
> are _equal_, the message is a parent message."
> "Threaded replies are also identified by comparing the `thread_ts` and `ts`
> values. If they are _different_, the message is a reply."

The `reaction_added` payload carries only `item.channel` + `item.ts` and **no
`thread_ts`** (confirmed at https://docs.slack.dev/reference/events/reaction_added
— *"Embedded `item` nodes are more lightweight than the structures you'll find in
`reactions.list`"*). So a naive `thread_ts = item.ts` on a reaction to a *thread
reply* would ask Slack to thread under a reply — the exact thing the doc says to
avoid. In practice Slack silently re-parents it to the real thread rather than
erroring, but the returned `ts` and any `reply_broadcast` semantics get muddy,
and it is explicitly documented as wrong.

### What Hermes actually does — resolves the real parent

`plugins/platforms/slack/adapter.py:4861-4877`:

```python
        thread_ts: Optional[str] = msg_ts
        if client is not None:
            try:
                history = await client.conversations_replies(
                    channel=channel_id, ts=msg_ts, limit=1, inclusive=True,
                )
                messages = (history or {}).get("messages") or []
                if messages:
                    first = messages[0]
                    thread_ts = first.get("thread_ts") or first.get("ts") or msg_ts
```

This is precisely the "call `conversations.replies` on `item.ts` to resolve the
true parent" implementation. **Verified empirically, read-only, on the reference
thread `C08RWSTU9LK` / `1786135149.072829`:**

`conversations.replies?channel=C08RWSTU9LK&ts=1786135270.459679&limit=3`
(where `1786135270.459679` is a *reply*, not the parent) →

```
ok= True
n= 1
1786135270.459679 thread_ts=1786135149.072829 user=U08BDJAMSRZ
```

Slack returns the reply itself as `messages[0]` with `thread_ts` = the **true
parent** `1786135149.072829`. Hermes' `first.get("thread_ts")` therefore yields
the parent, and the reply lands **in the same thread the reaction was added
to** — which is exactly what Gaetan asked for. No new thread, no error.

Fallback behaviour if the lookup fails (`adapter.py:4885-4890`) is `thread_ts =
msg_ts`, documented at `adapter.py:4856-4860` as *"correct for top-level messages
and degrades gracefully for in-thread reactions where we lose the parent
linkage"*. On Tars the lookup will succeed — `channels:history` is granted.

Optional override: `slack.reaction_trigger_target` (`adapter.py:4992-5006`,
used at `:4946-4956`) can redirect the answer to a fixed channel/thread instead
(`C123` = top-level in that channel, `C123:<thread_ts>` = that thread). Not
needed here; leave it unset so the answer stays in-thread.

**Answer to Gaetan's second question: yes, Tars answers in the thread where the
emoji was added.**

---

## 7. Allowlist interaction — **covered, at the shared chokepoint. No regression.**

`_handle_slack_reaction` does not implement its own auth. It builds a synthetic
message event and hands it to the same function every real message goes through
(`plugins/platforms/slack/adapter.py:4958`):

```python
        await self._handle_slack_message(synthetic)
```

with `"user": user_id` = **the reactor** (`adapter.py:4917`). The allowlist check
lives inside `_handle_slack_message` at
**`plugins/platforms/slack/adapter.py:5528-5549`**:

```python
        # Reject unauthorized users before thread lookups, name resolution,
        # or file downloads.  The final gateway runner auth check happens
        # after MessageEvent construction, so adapter-side media fetches need
        # the same auth chain up front.
        _runner = getattr(getattr(self, "_message_handler", None), "__self__", None)
        _auth_fn = getattr(_runner, "_is_user_authorized", None)
        if user_id and callable(_auth_fn):
            _source = self.build_source(
                chat_id=channel_id, chat_name="",
                chat_type="dm" if is_dm else "group",
                user_id=user_id, user_name="",
            )
            if not _auth_fn(_source):
                logger.warning(
                    "[Slack] Early reject of unauthorized user %s in channel %s",
                    user_id, channel_id,
                )
                return
```

That is the exact WARNING line WF4 uses as allowlist proof. A second,
independent gateway-runner auth check exists at `adapter.py:6665`
(`auth_fn = getattr(runner, "_is_user_authorized", None)`) after `MessageEvent`
construction — defence in depth.

**Verdict: a reaction trigger would NOT bypass `SLACK_ALLOWED_USERS`.** A
teammate reacting with the trigger emoji produces the same
`[Slack] Early reject of unauthorized user <id> in channel <id>` WARNING as a
teammate typing a message. Hermes' own doc states it
(`website/docs/user-guide/messaging/slack.md:671-675`):

> "The reactor becomes the message's user, so **user authorization and
> `allowed_channels` gating apply exactly as for typed messages** — a random
> user's reaction cannot trigger the agent anywhere their message couldn't."

**What a reaction trigger DOES bypass — deliberately — is the mention gate.**
`adapter.py:4928` sets `"_hermes_force_process": True`, consumed at
`adapter.py:5627` and `:5679`:

```python
        # Internal routing paths (reaction triggers) are pre-authorized as
        # "addressed to the bot" — they skip the mention requirement but NOT
        # the allowed_channels whitelist or user authorization above.
        force_process = bool(event.get("_hermes_force_process"))
        ...
            if force_process:
                pass  # Explicit internal routing path (reaction trigger).
```

So `require_mention: true` + `strict_mention: true` (both live) do **not** block
a reaction trigger. That is the point — an emoji IS the addressing act. The
security perimeter that matters (`SLACK_ALLOWED_USERS`) is untouched.

Three further built-in safeties, all in `_handle_slack_reaction`:
- self-reactions dropped (`adapter.py:4820-4822`) — no 👀-ack feedback loop;
- non-`message` items (files) dropped (`adapter.py:4808-4809`);
- with `reaction_triggers: true` (boolean form), only reactions on the **bot's
  own** messages route (`adapter.py:4878-4894`); an explicit emoji allowlist
  lifts that restriction *only for the listed emoji*.

`allowed_channels` is currently unset in `~/.hermes/config.yaml`, so a trigger
emoji would work in any channel Tars is a member of. That is a policy choice,
not a bug — `SLACK_ALLOWED_USERS` still means only Gaetan can fire it.

---

## 8. Alternative trigger paths needing no code change — cheapest first

### (a) Config-only: `slack.reaction_triggers` — **THE ANSWER**

Not really an "alternative": it is the shipped, intended feature. Costs 2 config
lines + 1 scope + 1 event subscription, no code. Full treatment in §9.

Works today? **No** — blocked only by the missing `reactions:read` scope.
Breaks? Nothing. It is opt-in, default-off, and gated by the existing allowlist.

### (b) Slack "Add to <app>" message shortcut / message action — **does not exist**

Grep for `@app.shortcut`, `.shortcut(`, `message_action` across the entire
package (excluding `venv/`, `node_modules/`): **zero hits**. Hermes registers
Bolt `@app.event`, `@app.command` and `@app.action` (Block Kit buttons) only.
`interactivity` appears once, at `hermes_cli/slack_cli.py:156`, purely to enable
Block Kit button callbacks in the manifest.

A message shortcut would additionally need a `view_submission`/`shortcut`
listener that does not exist ⇒ **real code change, larger than (a)**. Rejected.

### (c) `hermes cron` poller calling `reactions.get` / `conversations.history`

Would work with **no Slack app change at all**? No — `reactions.get` also
requires `reactions:read`. `conversations.history` (granted) returns a
`reactions` array per message, so a poller *could* technically detect new
reactions with today's scopes.

Cost: a cron job per watched channel, a persisted "already seen" set (nothing
ships one), polling latency (minimum ~1 min vs. sub-second for Socket Mode),
N API calls/minute forever, and it re-implements the dedup, thread-parent
resolution and allowlist enforcement that §6/§7 already give for free. It also
misses the reactor identity cleanly only for the most recent reactions.

**Strictly worse than (a) on every axis, and it is *more* code, not less.**
Rejected.

### (d) Slack Workflow Builder triggered by an emoji reaction

Workflow Builder does offer an emoji-reaction trigger. Its steps would have to
reach Hermes — but Tars is **Socket Mode**, so there is no inbound HTTP endpoint
for a webhook step to call, and nothing on the VM is exposed to Slack. It would
require standing up a webhook receiver (new service, new port, new exposure,
tailnet/ingress decision) to convert the workflow into a `hermes cron`/CLI
invocation.

That is a whole new network surface to replace two config lines. Also: Workflow
Builder emoji triggers are a paid-plan feature and the workflow's identity is
the *installer's*, not Tars', which muddies the allowlist story.

**Rejected — most expensive option by a wide margin.**

### (e) Anything else already shipped

The gateway hook surface (`gateway/platforms/base.py:3364`, fired at
`adapter.py:4835-4854`) emits `reaction:added` / `reaction:removed` to hook
consumers **regardless of the `reaction_triggers` opt-in**. It still needs the
`reactions:read` scope to receive the event in the first place, so it does not
avoid the dashboard change — and it produces no agent turn. Useful for
observability, not for "make Tars do the thing". Noted, not proposed.

### Comparison

| Path | Works today | Slack dashboard change | Code change | Latency | Verdict |
|---|---|---|---|---|---|
| (a) `reaction_triggers` config | No — scope only | 1 scope + 1 event | **none** | sub-second (socket) | **CHOSEN** |
| (b) message shortcut | No | scope + shortcut decl | new Bolt listener | sub-second | rejected — code |
| (c) cron + `conversations.history` | Partially | none | poller + state store | ≥1 min | rejected — more code, worse |
| (d) Workflow Builder → webhook | No | workflow + new endpoint | new HTTP receiver | seconds | rejected — new network surface |
| (e) gateway hooks | No — scope | 1 scope + 1 event | hook consumer | sub-second | observability only |

---

## 9. Verdict + minimal proposal

### Verdict: **PARTIAL**

An emoji trigger does **not** work today. It is not missing — it is switched
off. The implementation, the emoji-allowlist mode built for exactly this
use case, the in-thread reply, and the allowlist enforcement are all already in
the installed package and already documented by upstream Hermes at
`website/docs/user-guide/messaging/slack.md:641-684`:

> "Or an explicit emoji allowlist — only these names route, and they may target
> ANY message (**emoji-handoff workflows, e.g. `:task:` to capture**)"

That is Gaetan's question, answered by the vendor's own example.

Since Kanban is already enabled (12 `kanban_*` tools on the Slack surface), the
action side needs nothing.

### The minimal change — no code

**Step 1 — Slack app dashboard** (api.slack.com/apps → `A0BC0GXH78R`):

1. *OAuth & Permissions* → Bot Token Scopes → add **`reactions:read`**.
   (Optionally also `reactions:write` — unrelated to this feature, but it would
   make the existing 👀 lifecycle ack actually work; it has been silently
   failing at `adapter.py:3729` for lack of scope. Separate decision.)
2. *Event Subscriptions* → Subscribe to bot events → add **`reaction_added`**.
   **Do not add `reaction_removed`** — un-reacting should not fire a second
   agent turn. The adapter registers a listener for it either way; not
   subscribing simply means Slack never sends it. One less thing to reason about.
3. Re-install to the workspace when prompted.

Do **not** paste the output of `hermes slack manifest` — it would also (re)declare
slash commands that Slack blocks on the Agent surface and could disturb the
working `agent_view` config. Two additive edits in the UI, nothing else.

**Step 2 — config** (`~/.hermes/config.yaml`, under the existing
`gateway.platforms.slack` block at lines 118-124). Per the repo hard rule:
`.bak` copy first, then `flock ~/.hermes/.wf3.lock -c '<edit>'`, **merge, never
append**. Hermes live-reloads in ~30 s.

```diff
--- a/home/gaetan/.hermes/config.yaml
+++ b/home/gaetan/.hermes/config.yaml
@@ -118,10 +118,14 @@
 gateway:
   platforms:
     slack:
       enabled: true
       require_mention: true
       strict_mention: true
       unauthorized_dm_behavior: ignore
+      # Emoji-handoff: reacting with :kanban: routes the reacted-to message
+      # into the agent loop as `reaction:added:kanban`, threaded under the
+      # reacted-to message's parent. Explicit allowlist => may target ANY
+      # message (not just Tars' own). SLACK_ALLOWED_USERS still gates the
+      # reactor (adapter.py:5528).
+      reaction_triggers: [kanban]
     a2a:
       enabled: true
       extra:
         port: 9900
```

Use the **list** form, not `true`. `reaction_triggers: true` would only route
reactions on Tars' *own* messages (`adapter.py:4878-4894`) — useless for "I see a
message, I want it captured". The list form is what makes the emoji work on any
message, and simultaneously means every *other* emoji in every channel stays
inert. Both properties come from `adapter.py:4858-4860`:

```python
        explicit_allowlist = bool(triggers)
        if explicit_allowlist and reaction_name.strip(":") not in triggers:
            return
```

`:kanban:` must exist as a workspace custom emoji (Slack sends the emoji *name*;
unknown names pass through verbatim as `reaction:added:kanban` per
`adapter.py:4907` — `_REACTION_EMOJI_MAP` has no entry, so no translation).
Any name works; `kanban` is unambiguous and unlikely to be used casually.

**Step 3 — make the model act on it.** The agent receives a turn whose text is
literally `reaction:added:kanban`, threaded under the reacted-to message, with
the target message's text injected as `reply_to_text` (`adapter.py:4787-4791`).
One line in SOUL/system prompt maps that token to the Kanban tool, e.g.:

> When a turn's text is `reaction:added:kanban`, create a Kanban card from the
> message being replied to and confirm in one line.

That is prompt text, not code.

### Re-auth / secret-rotation implication — **none expected**

Slack's OAuth scoping is **additive**
(https://docs.slack.dev/authentication/installing-with-oauth):

> "Any subsequent time(s) you send that same user through the OAuth flow, any new
> scopes you request will be added to that initial set." … "each scope you
> request is **additive** to the scopes you've already been awarded."

Tokens are invalidated only on explicit revocation / uninstall / deactivation.
So adding `reactions:read` and re-installing should **not** change the `xoxb`
value ⇒ **no `scripts/tars-secret SLACK_BOT_TOKEN` re-ingest, no SOPS rotation,
no re-delivery to the VM.** `SLACK_APP_TOKEN` (`xapp-…`) is unaffected by bot
scope changes.

Verify rather than assume — after re-install, one read-only call:

```
ssh gaetan@192.168.0.9 'set -a; . ~/.hermes/.env; set +a;
  printf "header = \"Authorization: Bearer %s\"\n" "$SLACK_BOT_TOKEN" |
  curl -sS -D - -o /dev/null -K - "https://slack.com/api/auth.test" |
  grep -i "^x-oauth-scopes"'
```

`ok:true` + `reactions:read` present in the header ⇒ nothing to rotate. If it
returns `invalid_auth`, the token did change: re-ingest via
`scripts/tars-secret SLACK_BOT_TOKEN` and deliver with the sanctioned per-key
`sops -d --extract` pipe, then restart `hermes-gateway.service`.

### Unified diff against the installed package

**None required.** No file under `/home/gaetan/.hermes/hermes-agent/` needs to
change. The only diff is the `config.yaml` hunk above.

### Verification once enabled (read-only, for whoever applies this)

1. `grep "reaction" ~/.hermes/logs/gateway.log` after reacting once — the
   handler is reached.
2. Negative test: have a second (non-Gaetan) member react with `:kanban:` and
   grep for `[Slack] Early reject of unauthorized user` — the WF4 allowlist
   proof line, from `adapter.py:5546`.
3. Thread test: react to a message that is itself a **thread reply**, then poll
   `conversations.replies` on the *parent* ts — Tars' answer must appear there,
   not as a new thread. (`conversations.history` will show nothing; that is
   expected — see repo facts.)

---

## What was NOT done (deliberate)

- No Slack write method called. No message sent. No reaction added or removed.
- No file on the VM read or written outside `~/.hermes/{config.yaml,.env}`
  (key names only from `.env`) and the read-only package source.
- No `sops -d` of any kind, whole-file or per-key. No secret on argv, echoed,
  or written here.
- No config or service change. The Slack app manifest was not modified and
  could not be read (needs a configuration token, not a bot token) — the
  missing-scope evidence is used instead, which is conclusive.
