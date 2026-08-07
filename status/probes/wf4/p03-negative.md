# WF4 Probe 3 — THE NEGATIVE TEST (non-Gaetan sender must be ignored)

**Verdict: DEFERRED** — the send was never made, on purpose. The spec's designated
non-Gaetan sender does not exist: the `claude_ai_Slack` connector authenticates **as Gaetan**
(`U08BDJAMSRZ`), not as a distinct app identity. Using it would have been a positive test wearing
a negative test's label, and the explicit instruction is to never fall back to a Gaetan-controlled
identity or to message a human teammate unasked.

Not a FAIL: nothing was observed failing. The guardrail surface it targets was pinned in full
from the **running** gateway (read-only), and the evidence design for the eventual send is now
known to be a positive greppable WARNING rather than an absence argument.

- **Probe window:** 2026-08-07T19:00:07+00:00 → 2026-08-07T19:03:22+00:00 (all timestamps `date -Is`)
- **Post-cutover:** yes — cutover restart was 2026-08-07T18:53:52+00:00, gateway PID **61969**,
  `active`, unchanged and untouched throughout.
- **Mutations:** none. No message sent, no unit restarted, no `config.yaml`/`.env` edit, no `sops`,
  no git, no pve/p-Hermes contact.

---

## 1 · Why the send did not happen — connector identity is Gaetan

The spec (§3, "Primary") assumes `claude_ai_Slack` is "a distinct app/bot identity already
legitimately present in the workspace." It is not. It is a **user-token connector bound to
Gaetan's own account**.

Two independent confirmations, both before any send attempt:

**(a) The tool contract states it outright.** `mcp__claude_ai_Slack__slack_send_message`'s own
description: *"If the user wants to send a message to themselves, the current logged in user's
user_id is U08BDJAMSRZ."* Same string in `slack_search_users`.

**(b) `slack_read_user_profile` with no `user_id` (defaults to the caller's own identity),
2026-08-07T19:00:13+00:00:**

```
User ID: U08BDJAMSRZ
Username: gaetan.cathelain
Display Name: Gaëtan
Real Name: Gaëtan Cathelain
Title: DevOps
Email: gaetan.cathelain@mobile.club
Organization Name: Mobile Club
Timezone: Europe/Brussels
Admin: No   Owner: No   Bot: No   Restricted: No
```

`Bot: No`, and the user ID is byte-identical to `SLACK_ALLOWED_USERS` (`U08BDJAMSRZ`, A4 /
`status/lane-a.md:17`, DECISION Blocker 2). The instruction gate was *"first confirm the connector
identity is not U08BDJAMSRZ"* — it **is** U08BDJAMSRZ, so the gate closed and no message was sent.

This is a **stronger** stop than the anticipated one. The briefing predicted "Slack refuses
bot-to-bot DM"; the reality is that Slack would have accepted the send happily and delivered a
message **from the allow-listed user**. Tars would then have been *correct* to reply, and a green
"no reply" would have been unobtainable while a reply would have been misread as a guardrail
breach. Either outcome would have been a false reading of the security-critical probe.

No fallback was taken: no second Gaetan-controlled identity was created (spec §3 forbids it), and
no teammate was messaged — the spec's consented-teammate fallback requires *Gaetan's* real-time
human consent, which is not mine to grant.

---

## 2 · Guardrail surface pinned from the RUNNING gateway (read-only)

Gateway identity for every capture below — 2026-08-07T19:00:34+00:00:

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); \
    systemctl --user is-active hermes-gateway.service; \
    systemctl --user show -p MainPID --value hermes-gateway.service'
active
61969
```

### 2a · `SLACK_ALLOWED_USERS` — present, single value, and it is Gaetan

```
$ 2026-08-07T19:01:40+00:00 — cut -d= -f1 ~/.hermes/.env | sort
GMAIL_ADDRESS GMAIL_APP_PASSWORD HINDSIGHT_MODE LINEAR_API_KEY NOTION_API_TOKEN
NOTION_FILE_TOKEN NOTION_SPACE_ID NOTION_TOKEN_V2 SLACK_ALLOWED_USERS SLACK_APP_TOKEN
SLACK_BOT_TOKEN SLACK_HOME_CHANNEL SLACK_MCP_XOXC_TOKEN SLACK_MCP_XOXD_TOKEN
key count: 14

$ grep -E '^SLACK_ALLOWED_USERS=' ~/.hermes/.env
SLACK_ALLOWED_USERS=U08BDJAMSRZ          <- not a secret (per cutover.md §1d)
```

14 keys = the 13 from the cutover `.env` rewrite + `SLACK_HOME_CHANNEL` added by `/sethome`.
Exactly one ID in the allowlist; `_coerce_allow_set` (`authz_mixin.py:73-84`) splits on commas, so
a one-element set is what the gateway holds.

### 2b · Trap cleared — the allowlist is NOT in `/proc/<pid>/environ`, and that is correct

```
$ 2026-08-07T19:01:05+00:00 — tr '\0' '\n' < /proc/61969/environ | cut -d= -f1 | sort
DBUS_SESSION_BUS_ADDRESS GSM_SKIP_SSH_AGENT_WORKAROUND HERMES_HOME HOME INVOCATION_ID
JOURNAL_STREAM LANG LOGNAME MANAGERPID MEMORY_PRESSURE_WATCH MEMORY_PRESSURE_WRITE PATH
QT_ACCESSIBILITY SHELL SSH_AUTH_SOCK SYSTEMD_EXEC_PID USER VIRTUAL_ENV XDG_DATA_DIRS
XDG_RUNTIME_DIR                                            (20 keys — zero SLACK_*)

$ systemctl --user show-environment | grep -i '^SLACK'
(no SLACK_* in user manager environment)
```

**Do not read this as a missing guardrail.** `/proc/<pid>/environ` is the environment *as of
`execve`* and never reflects later `os.environ` writes. The unit file carries no
`EnvironmentFile=` — only `PATH`, `VIRTUAL_ENV`, `HERMES_HOME`:

```
ExecStart=/home/gaetan/.hermes/hermes-agent/venv/bin/python -m hermes_cli.main gateway run
WorkingDirectory=/home/gaetan/.hermes
Environment="PATH=…"  Environment="VIRTUAL_ENV=…"  Environment="HERMES_HOME=/home/gaetan/.hermes"
```

so `~/.hermes/.env` is loaded by the process itself after exec. The runtime proof that it landed
is §2d below. **A future probe must not "verify `SLACK_ALLOWED_USERS` in the gateway environment"
by reading `/proc/<pid>/environ` — it will read empty on a healthy gateway.**

### 2c · `SLACK_STRICT_MENTION` — spec §2 trap confirmed, with the exact mechanism

Not in `.env`, not in the exec environ. It is written into `os.environ` at plugin load from
config, `plugins/platforms/slack/adapter.py:8992-8995`:

```python
if "require_mention" in slack_cfg and not os.getenv("SLACK_REQUIRE_MENTION"):
    os.environ["SLACK_REQUIRE_MENTION"] = str(slack_cfg["require_mention"]).lower()
if "strict_mention" in slack_cfg and not os.getenv("SLACK_STRICT_MENTION"):
    os.environ["SLACK_STRICT_MENTION"] = str(slack_cfg["strict_mention"]).lower()
```

Read back at `adapter.py:8273` as `os.getenv("SLACK_STRICT_MENTION", "false").lower() in {…}`.
With `strict_mention: true` in config (§2e), the exported value is the string `"true"`.

The docstring at that site states the design explicitly: *"no extras are seeded into
`PlatformConfig.extra` directly (everything flows through env)"* — which is precisely why grepping
the parsed config object finds nothing.

### 2d · Runtime proof the allowlist is actually in effect

`run.py:10943-11007` warns at startup **only when no allowlist env var is set**:

```python
_builtin_allowed_vars = (…, "SLACK_ALLOWED_USERS", …)
if not _any_allowlist and not _allow_all:
    logger.warning("No env user allowlists configured. Messaging platforms default to "
                   "pairing/allowlist policies and will deny unknown senders unless you …")
```

```
$ journalctl --user -u hermes-gateway.service --since '2026-08-07 18:53:50' --no-pager \
    | grep -c 'No env user allowlists configured'
0
```

Zero occurrences since the cutover restart ⇒ `os.getenv("SLACK_ALLOWED_USERS")` was truthy in
PID 61969 at startup ⇒ **the allowlist survived the `.env` rewrite and `/sethome`, and is live in
the running process.** (Same inference used by `cutover.md` §3c and `cutover-sethome.md`; re-run
here inside the WF4 window.)

Auth health in the same window, 2026-08-07T19:02:52+00:00:

```
$ journalctl … --since '2026-08-07 18:53:50' \
    | grep -icE 'invalid_auth|not_authed|account_inactive|token_revoked|Refusing to start'
0
```

### 2e · `config.yaml` — all three D2 guardrails, correctly nested, no stray override

Grepped for the guardrail keys only, never dumped (mode 600, holds secrets):

```
$ grep -nE 'require_mention|strict_mention|unauthorized_dm_behavior|^gateway:|^ *slack:' ~/.hermes/config.yaml
118:gateway:
120:    slack:
122:      require_mention: true
123:      strict_mention: true
124:      unauthorized_dm_behavior: ignore
153:  slack:
175:  slack:
```

Spec §2's "stray top-level `platforms:` block silently overrides `gateway.platforms.*`" trap —
**cleared**. Top-level keys of `config.yaml`:

```
model database agent terminal browser tool_loop_guardrails compression prompt_caching display
stt memory delegation skills code_execution gateway streaming telemetry updates _config_version
session_reset group_sessions_per_user platform_toolsets plugins context mcp_servers
```

No top-level `platforms:`. The other two `slack:` keys are unrelated and correctly parented:

| line | parent | verdict |
|---|---|---|
| 120 | `gateway:` → `platforms:` (indent 4) | the guardrail block — authoritative |
| 153 | `platform_toolsets:` (indent 2) | toolset scoping, not a guardrail override |
| 175 | `mcp_servers:` (indent 2) | the personal-Slack MCP server (probe 9) |

### 2f · No bypass switch is set anywhere

`authz_mixin.py:497` grants bots a pass when `SLACK_ALLOW_BOTS ∈ {mentions, all}`; `:539` does the
same for `SLACK_ALLOW_ALL_USERS`; `run.py` honours `GATEWAY_ALLOW_ALL_USERS`.

```
$ grep -nE 'allow_bots|allow_all|ALLOW_ALL|ALLOW_BOTS|dm_policy|group_policy' ~/.hermes/config.yaml
(none present)
$ grep -cE '^(SLACK_ALLOW_BOTS|SLACK_ALLOW_ALL_USERS|GATEWAY_ALLOW_ALL_USERS)=' ~/.hermes/.env
0
```

All three unset ⇒ defaults hold (`_platform_gate_env("SLACK_ALLOW_BOTS", "none")`), so no identity
— human or app — bypasses the allowlist. Worth noting for the eventual send: had a real bot been
available, `SLACK_ALLOW_BOTS` being unset means it would have been rejected on the bot branch
*before* the allowlist branch. Both gates are shut; a passing negative test proves the pair.

### 2g · `unauthorized_dm_behavior: ignore` resolves to a genuine drop

`authz_mixin.py:785-880`, precedence order: explicit **per-platform** `unauthorized_dm_behavior` in
config *always wins* (`:794`, `:812-816`). Tars sets it per-platform at
`gateway.platforms.slack.unauthorized_dm_behavior: ignore`, so the first rule fires — no pairing
code, no reply. The allowlist-aware fallback at `:851-880` (`SLACK_ALLOWED_USERS` non-empty →
`"ignore"`) would independently reach the same answer. Belt and braces.

---

## 3 · What the eventual send will look like — evidence design, settled

The reject site is `logger.**warning**`, not debug — `plugins/platforms/slack/adapter.py:5544-5549`:

```python
if not _auth_fn(_source):
    logger.warning(
        "[Slack] Early reject of unauthorized user %s in channel %s",
        user_id, channel_id,
    )
    return
```

Every line the gateway currently emits to the journal is WARNING level, so **this line will
appear**. Consequences for whoever completes probe 3:

1. PASS evidence is a **positive grep**, not the weaker "timestamped absence" the spec allows:
   `journalctl --user -u hermes-gateway.service --since '<T>' | grep 'Early reject of unauthorized user'`
   → expect one line naming the sender's ID and the channel/DM ID.
2. The reject happens in the adapter **before** thread lookup, name resolution and file download
   (comment at `:5529-5532`), so a passing probe also proves no media fetch was triggered.
3. Absence of that line **with** a non-Gaetan message delivered would be the FAIL — and the
   distinction is only readable if the sender ID is captured at send time. Record it.

---

## 4 · What is still needed to close this probe

A sender that is genuinely not `U08BDJAMSRZ`. Three options, Gaetan's call — all need a human
decision, none is mine to take unattended:

1. **A second app/bot identity** in `T7V1UGJ82` with `chat:write` (any existing workspace bot).
   Cleanest: no human involved, and `SLACK_ALLOW_BOTS` being unset means it exercises both gates.
2. **The spec's consented-teammate fallback** (§3) — one teammate, explicit real-time consent,
   one innocuous message, logged. Requires Gaetan to ask.
3. **A channel variant** — only helps if it originates from a non-Gaetan identity. A channel run
   from the `claude_ai_Slack` connector is still Gaetan and still proves nothing; the DM-vs-channel
   axis is not what blocks this probe.

Until then probe 3's *behavioural* half is unexercised. Its *configuration* half is fully pinned
above and clean on every one of D2's four controls.

---

## 5 · Out-of-scope observation (not mine to adjudicate)

The journal window shows, at 2026-08-07T19:01:21→19:02:10, PID 61969:

```
WARNING agent.conversation_loop: Empty response (no content or reasoning) — retry 1..3/3 (model=gpt-5.6-sol)
WARNING agent.conversation_loop: Empty response (no content or reasoning) after 3 retries.
                                 No fallback available. model=gpt-5.6-sol provider=openai-codex
```

Concurrent with another WF4 probe's live prompt, not with anything I ran (I sent no message and
opened no session). Flagging for whoever owns probes 1 and 10 — it looks like a model round-trip
returning empty, which is exactly what probe 10's "configured-but-broken" clause is meant to
catch. Untouched by me.

## 6 · Scope discipline

No Slack message sent (the entire point). No unit restarted/stopped/enabled/disabled. No
`config.yaml` or `.env` edit — reads only, and only greps for named non-secret keys. No `sops`
invocation. No secret echoed: the one value printed, `SLACK_ALLOWED_USERS=U08BDJAMSRZ`, is a
member ID already recorded as non-secret in `cutover.md` §1d and `DECISION.md`. No git command.
No `status/lane-*.md` edit. pve (192.168.0.3) and p-Hermes untouched.
