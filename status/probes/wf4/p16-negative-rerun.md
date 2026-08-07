# WF4 Probe 16 — Probe 3 re-run: THE NEGATIVE TEST, with a genuine non-Gaetan sender

**Verdict: PASS-ALLOWLIST** — a real non-Gaetan human (`U05LKHLDV0A`, Nans Brun Dumortier)
@mentioned Tars in `C08RWSTU9LK`; the gateway logged the positive reject WARNING 0.445 s later,
never dispatched the message to the agent, and never replied. All three of §16's PASS conditions
are met, and the strongest form (positive grep, not an absence argument) is the one obtained.

- **Probe window:** 2026-08-07T20:20:48+00:00 → 2026-08-07T20:24:17+00:00 (`date -Is`, VM clock UTC)
- **Message under test:** `C08RWSTU9LK`, ts `1786133895.750309` = **2026-08-07T20:18:15.750Z**
  ([permalink](https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786133895750309))
- **Gateway:** `hermes-gateway.service` `active`, MainPID **76255**, unchanged throughout
- **Mutations by this probe: none.** No Slack message sent, no unit touched, no `config.yaml`/`.env`
  write, no `sops`, no git, no contact with 192.168.0.3. Every Slack call is a read
  (`conversations.history`, `conversations.replies`, `users.info`) through the existing
  `/home/gaetan/tars/wf4/slackapi.sh` helper on the VM (mode 700, `curl -K` config from a heredoc
  under `umask 077`, `shred -u` on EXIT) — tokens never on argv, never printed.

---

## 1 · Sender identity — genuinely non-Gaetan, genuinely human

`conversations.history channel=C08RWSTU9LK latest=oldest=1786133895.750309 inclusive=true`, HTTP 200,
`ok: true`:

```json
{ "ts": "1786133895.750309", "user": "U05LKHLDV0A",
  "bot_id": null, "app_id": null, "subtype": null,
  "text": "<@U0BBH85NAKH> test WF4 negative test" }
```

`users.info user=U05LKHLDV0A`:

| field | value |
|---|---|
| id | **U05LKHLDV0A** |
| name | `nans` |
| real_name | Nans Brun Dumortier |
| title | Software Engineer |
| email | nans@mobile.club |
| team_id | `T7V1UGJ82` (the Tars workspace) |
| is_bot / is_app_user / deleted | `false` / `false` / `false` |

Gate checks, both satisfied:

- `U05LKHLDV0A` ≠ **`U08BDJAMSRZ`** (Gaetan / the sole entry of `SLACK_ALLOWED_USERS`) → this is
  the sender p03 §4 was blocked on. The p03 §1 disqualification of the `claude_ai_Slack` connector
  (it authenticates *as* `U08BDJAMSRZ`) does not apply: this is a real teammate's own account.
- `U05LKHLDV0A` ≠ **`U0BBH85NAKH`** (Tars' own bot user) → not Tars talking to itself.

This is spec §3/§16 route **(b)**, the consented-teammate fallback: Nans Brun Dumortier, message
`"@Tars test WF4 negative test"`, sent 2026-08-07T20:18:15.750Z into `C08RWSTU9LK`. The consent was
obtained by Gaetan out-of-band; this probe only observed the result and sent nothing.

## 2 · The text carries a REAL mention tag — so the allowlist gate is what fired

Not a literal string `"@Tars"`: the message's `blocks` contain a `rich_text` → `rich_text_section`
whose first element is a genuine mention entity —

```json
{"type": "rich_text", "block_id": "Y2cbz", "elements": [{"type": "rich_text_section", "elements": [
   {"type": "user", "user_id": "U0BBH85NAKH"},
   {"type": "text", "text": " test WF4 negative test"} ]}]}
```

`user_id: U0BBH85NAKH` is Tars' bot user ID. So `require_mention: true` / `strict_mention: true`
(config lines 122-123) are **satisfied** by this message, and the drop cannot be attributed to the
mention gate. The gate that must fire — and did — is the allowlist. This is the branch p03 §3 said
would produce a positive greppable WARNING, and it is the branch §16 calls the strongest evidence.

## 3 · The positive reject line — quoted, from all three sinks

Latency: message `20:18:15.750` → reject logged `20:18:16.195` = **0.445 s**.

```
$ journalctl --user -u hermes-gateway.service --since '2026-08-07 20:17:00' --no-pager | grep 'Early reject'
Aug 07 20:18:16 tars python[76255]: WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK

$ grep 'Early reject' ~/.hermes/logs/errors.log
275:2026-08-07 20:18:16,195 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK

$ grep -n 'Early reject' ~/.hermes/logs/gateway.log
108:2026-08-07 20:18:16,195 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
```

The line names **the exact sender ID** and **the exact channel** — the discriminator p03 §3.3
demanded. Source site `plugins/platforms/slack/adapter.py:5544-5549`, unchanged.

### 3a · The drop is pre-dispatch: the agent never saw the message

```
$ grep -n 'U05LKHLDV0A' ~/.hermes/logs/agent.log
1772:2026-08-07 20:18:16,195 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U05LKHLDV0A in channel C08RWSTU9LK
```

**One** occurrence in the whole of `agent.log`, and it is the reject WARNING itself (WARNINGs fan
out to all sinks). No `inbound message: platform=slack user=U05LKHLDV0A …`, no `agent.turn_context`,
no `conversation_loop`, no `response ready`, no `[Slack] Sending response` — anywhere. Consistent
with p03 §3.2: the reject precedes thread lookup, name resolution and file download, so no media
fetch was triggered either.

For contrast, `gateway.log` shows what an *accepted* message looks like — the allow-listed user, in
the same channel, before and after (§5).

### 3b · Config still holds the guardrails, re-read at 20:24:17 (after the event)

```
$ grep -nE 'require_mention|strict_mention|unauthorized_dm_behavior' ~/.hermes/config.yaml
122:      require_mention: true
123:      strict_mention: true
124:      unauthorized_dm_behavior: ignore
$ grep -E '^SLACK_ALLOWED_USERS=' ~/.hermes/.env
SLACK_ALLOWED_USERS=U08BDJAMSRZ                       (member ID, non-secret per cutover.md §1d)
$ grep -cE '^(SLACK_ALLOW_BOTS|SLACK_ALLOW_ALL_USERS|GATEWAY_ALLOW_ALL_USERS)=' ~/.hermes/.env
0
```

Unchanged from p03 §2a/§2e/§2f, and now proven *behaviourally* rather than only by inspection.

## 4 · Silence half — `conversations.replies`, window satisfied

Polled four times; the 3-minute floor (`20:18:15.750 + 180 s = 20:21:15.750`) is cleared from the
second poll onward:

| poll | wall clock | elapsed since ts | `messages` returned | from `U0BBH85NAKH` |
|---|---|---|---|---|
| 1 | 20:21:49Z | 213.8 s | 1 (the parent only) | **0** |
| 2 | 20:22:33Z | 258.0 s | 1 (the parent only) | **0** |
| 3 | 20:23:17Z | 301.5 s | 1 (the parent only) | **0** |
| 4 | 20:23:58Z | 342.9 s | 1 (the parent only) | **0** |

`conversations.replies(channel=C08RWSTU9LK, ts=1786133895.750309)` returns HTTP 200 / `ok: true`
with exactly one message — the parent itself, `user: U05LKHLDV0A`. **Zero thread replies, from
anyone, including zero from Tars.** Final margin **5 min 43 s** against a 3-minute requirement.

Cross-check on the channel top level (`conversations.history oldest=1786133895.750309`, 20:23:58Z),
because a bot could in principle answer outside the thread — 2 messages, both from `U042B9WCSRF`,
**zero with `user: U0BBH85NAKH`, zero with a `bot_id` or `app_id`**:

```
1786134112.637739  20:21:52Z  U042B9WCSRF  subtype=channel_join
1786134155.187189  20:22:35Z  U042B9WCSRF  "<@U0BBH85NAKH> t'as une photo des pieds de <@U08BDJAMSRZ> à me partager ?"
```

Corroborated log-side: `awk '$0>="2026-08-07 20:18:16,196"' ~/.hermes/logs/gateway.log | grep -c
'Sending response'` → **0** for the whole window after the reject, until the unrelated Gaetan turn
in §5.

### 4a · Free second data point — a second non-Gaetan human, same reject

Unprompted and not requested by this probe: **Jeremy Pinto (`U042B9WCSRF`**, `jeremy438`, Fullstack
Developper, `T7V1UGJ82`, `is_bot: false`) joined the channel at 20:21:52 and at 20:22:35 sent a real
mention of Tars (`blocks` mention tags: `['U0BBH85NAKH', 'U08BDJAMSRZ']`). Both events were rejected
identically:

```
2026-08-07 20:21:53,378 WARNING … [Slack] Early reject of unauthorized user U042B9WCSRF in channel C08RWSTU9LK
2026-08-07 20:22:35,645 WARNING … [Slack] Early reject of unauthorized user U042B9WCSRF in channel C08RWSTU9LK
```

Three reject lines total in `errors.log`, two distinct non-Gaetan member IDs, zero replies to either
of them. The allowlist is not keyed to one unlucky sender.

## 5 · Liveness — the silence is a decision, not a dead gateway

`systemctl --user is-active hermes-gateway.service` → `active`, MainPID `76255`, at both 20:22:08
and 20:23:37. Activity strictly **after** the reject at 20:18:16.195:

**(a) The same gateway answering the allow-listed user, in the same channel, 5 min later** — the
cleanest possible control (accept path works ⇒ the reject was selective):

```
2026-08-07 20:23:26,582 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=C08RWSTU9LK msg='envois une photo de mes pieds' reply_to_id=1786134155.187189 …
2026-08-07 20:23:40,243 INFO gateway.run: response ready: platform=slack chat=C08RWSTU9LK time=13.7s api_calls=2 response=44 chars
2026-08-07 20:23:40,269 INFO gateway.platforms.base: [Slack] Sending response (44 chars) to C08RWSTU9LK
```

Same channel, same PID, same Socket-Mode connection: `U08BDJAMSRZ` in → answered; `U05LKHLDV0A` and
`U042B9WCSRF` in → rejected. That contrast is the guardrail, live.

**(b) Two further Slack events consumed** (the 20:21:53 and 20:22:35 rejects) — the socket was
receiving throughout the silence window.

**(c) Agent + housekeeping alive** (`agent.log`, concurrent WF4/WF5 traffic, not mine):

```
2026-08-07 20:21:43,054 INFO … agent.turn_context: conversation turn: session=20260807_202141_ad41c4 model=gpt-5.6-sol provider=openai-codex platform=cli …
2026-08-07 20:21:48,294 INFO … agent.conversation_loop: API call #1: … latency=5.2s
2026-08-07 20:22:00,324 INFO … agent.conversation_loop: API call #2: … latency=12.0s cache=23040/25244 (91%)
2026-08-07 20:22:02,479 INFO hermes_cli.mem_trim: memory trim: reason=messaging gateway housekeeping …
```

## 6 · What this closes, and what it does not

**Closed — §16 PASS conditions, all three:**

1. Sender ID captured, `U05LKHLDV0A` ≠ `U08BDJAMSRZ`, human, in-workspace (§1).
2. `journalctl … | grep 'Early reject of unauthorized user'` returns the line naming that sender and
   that channel (§3).
3. `conversations.replies(ts=…)` shows no Tars message in the thread, 5 min 43 s window (§4).

Probe 3's *behavioural* half — deferred since 2026-08-07T19:03 — is now exercised, and against a
**mentioned** message, so the allowlist gate is proven on its own rather than shadowed by the
mention gate. `unauthorized_dm_behavior: ignore` resolved to a genuine drop (p03 §2g): no pairing
code, no "you are not authorized" reply, no reply of any kind.

**Not closed:**

- **DM channel-type.** §16 asked for one DM **and** one channel mention; only the channel mention
  was sent. The reject site (`adapter.py:5544`) is shared by both surfaces and its log line prints
  the DM ID in the `channel` slot, so the DM path is *expected* to behave identically — but it is
  inferred, not observed. Anyone wanting it observed needs one teammate DM to Tars.
- **The bot gate.** `SLACK_ALLOW_BOTS` is unset, and §16 route (a) (a second app/bot identity) would
  have exercised the bot branch *before* the allowlist branch. Route (b) was taken, so only the
  allowlist branch is proven live; the bot branch remains config-only evidence (p03 §2f).

## 7 · Out-of-scope observation

`~/.hermes/config.yaml` mtime is `2026-08-07 20:21:48.974` — it was `18:08:43.859` when p09 recorded
it. Something in the concurrent WF4/WF5 traffic rewrote or re-serialised it during this window; it
was **not** this probe (reads only, greps for named keys, no editor, no `sops`). The three guardrail
keys and the allowlist were re-read *after* that mtime (§3b) and are intact, so it does not affect
this verdict — flagging it for whoever owns the config-drift question.

## 8 · Scope discipline

No Slack message sent (the whole point — the message under test was Nans', sent before this probe
started). No unit started/stopped/restarted/enabled/disabled. No `config.yaml` or `.env` write. No
`sops`. No secret echoed: the only credential-adjacent value printed is
`SLACK_ALLOWED_USERS=U08BDJAMSRZ`, a member ID already recorded as non-secret. No git command. No
`status/lane-*.md` edit. 192.168.0.3 and p-Hermes untouched.
