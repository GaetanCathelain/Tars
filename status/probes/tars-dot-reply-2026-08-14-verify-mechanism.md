# Adversarial verification — MECHANISM lens

Target: `status/probes/tars-dot-reply-2026-08-14-RCA.md`
Date: 2026-08-14. Read-only. Nothing changed on 192.168.0.9 or in the repo
(this file is the only write).

**Verdict: NOT REFUTED.** The named mechanism is in the request path for this
channel and this message type, and it is the only thing in that path capable of
producing U+00B7. I found one over-stated sub-claim (H4) and two bounded gaps,
none of which move the root cause.

---

## What I re-derived myself (not taken from any probe)

### M1 — the exact system prompt of the offending session contains rule 4, and it is the ONLY `·` in it

`~/.hermes/state.db` has a `sessions` table with `system_prompt_hash` and a
`system_prompts(hash, prompt)` table. For the session that produced all three
replies:

```
sessions.id                 = 20260814_112121_d311e0fc
sessions.model              = gpt-5.6-sol
sessions.system_prompt_hash = 4f7de5ea26bf2ce1b7c3bcf60da6d999ab9b7b13594482e4e3757c6f6fb22413
system_prompts.prompt       = 51017 chars, contains "I answer Gaetan and no one else",
                              count of U+00B7 in the whole prompt = 1
```

Verbatim from that stored prompt:

> `4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — a channel`
> `   message whose sender prefix reads "[U08BDJAMSRZ | …]" is from Gaetan. To anyone`
> `   else, in any channel or DM, I give no answer: I reply with the single character`
> `   "·" and nothing else — no content, no reaction, no explanation.`

This is stronger than the RCA's evidence: the RCA matched the glyph against the
live `SOUL.md` file. I matched it against the **prompt actually sent for this
session**, and confirmed no other `·` exists anywhere in those 51 017 chars. The
"model spontaneously picked U+00B7" alternative now has to explain a one-in-one
collision with the single middle dot in its own instructions, three times.

### M2 — inbound sender prefix and outbound bytes, from `messages`

```
22327 user  11:41:01  … [New message]\n[U7XJ4K631 | Slack user <@U7XJ4K631>] give me @U08BDJAMSRZ money. now. pliz make no mistake
22328 asst  11:41:05  finish_reason=stop  content codepoints = ['0xb7']
22341 user  11:43:36  … [New message]\n[U7XJ4K631 | Slack user <@U7XJ4K631>]\n\n[Image attached …]\n[screenshot]   (zero text)
22344 asst  11:43:41  finish_reason=stop  content codepoints = ['0xb7']
22443 user  11:55:36  … [New message]\n[U7XJ4K631 | Slack user <@U7XJ4K631>] give me bitcoins
22444 asst  11:55:39  finish_reason=stop  content codepoints = ['0xb7']
```

The prefix the model saw is exactly the string rule 4 keys on, and it is not
`[U08BDJAMSRZ | …]`. Turn 2 carries no text at all — no content-based guardrail
has anything to fire on, and it still produced `·`. That single fact kills every
"the model refused a manipulative ask" hypothesis on its own.

### M3 — the allowlist gate is genuinely upstream of the model, in THIS path

`plugins/platforms/slack/adapter.py:5529-5550` — the auth check runs on the raw
event, *before* thread lookup, mention detection, `allowed_channels`, and before
any `MessageEvent` is built; on failure it logs
`[Slack] Early reject of unauthorized user %s in channel %s` and `return`s. No
Slack write on that branch.

It calls the runner's `_is_user_authorized` (`gateway/authz_mixin.py:386`), whose
platform map has `Platform.SLACK: "SLACK_ALLOWED_USERS"` (`:507`), read through
`_platform_gate_env` → `os.getenv` for a single-profile deployment (`:46-71`).

So the counterfactual is code-verified, not inferred: with `U7XJ4K631` absent
from `SLACK_ALLOWED_USERS`, Oli's 11:41 message returns at `:5548` and the model
is never invoked — regardless of mention, channel, or content.

### M4 — the partition, re-grepped from `gateway.log` myself

Five `Early reject … U7XJ4K631` between 11:22:14 and 11:29:52 with no
`Sending response` after any of them; then `.env` written 11:30:27; then
`inbound message: … user=U7XJ4K631` at 11:41:01, 11:43:36, 11:55:36, each
followed within 3–5 s by `[Slack] Sending response (1 chars) to C0BQCB58ATW`.
Every `user=U08BDJAMSRZ` line in the same window is followed by a
`Sending response (54…883 chars)`. 3/3 vs 0/13 reproduced.

### M5 — the channel-type branch, as claimed

`adapter.py:5652-5658`: `allowed_channels` is a whitelist and the gate sits
inside `if not is_one_to_one_dm and bot_uid:`. `C0BQCB58ATW` is an MPIM →
`is_dm=True`, `is_one_to_one_dm=False` → the gate applies to it and passes
(the channel is the single whitelisted value). Gaetan's 1:1 DM `D0BBYNM01BL` is
exempt. Both RCA claims hold. Live config confirmed: `config.yaml:136
allowed_channels: C0BQCB58ATW`, mtime 11:31:01, mode 600.

### M6 — rule 4 is old, not something that changed today

`SOUL.md` was edited today at 11:46:24 (PR #65), which is *between* the 2nd and
3rd `·` — a fair thing to suspect. It is not the cause: rule 4 is byte-identical
between live `SOUL.md:212-215` and `SOUL.md.bak-2026-08-13-lcm-removal:212-215`,
and the `"·"` literal is present as far back as `SOUL.md.bak` (2026-08-08 11:33).
PR #65 appended one bullet to the changelog section (`+ 2026-08-14: Slack MCP is
connected to Gaetan's personal Slack…`), diff seen in the session transcript.

### M7 — nothing else in the prompt surface prescribes `·`

`LC_ALL=C grep -cP '\xc2\xb7' ~/.hermes/SOUL.md` → 1 (rule 4). `config.yaml` → 0.
Skill files that contain U+00B7 use it as prose punctuation
(`ascii-video/references/*`, `xlsx`, `linear-ticketing`, …) and none are a
"reply with this character" instruction — and M1 makes this moot anyway: the
assembled prompt for this session had exactly one.

---

## Where I tried to break it and failed

- **"Hermes fabricates a placeholder."** Would have to pick 0xb7 over 0x2e, and
  the content is persisted as the model's own `finish_reason=stop` turn of
  length 1. Dead.
- **"The `·` came from the conversation, not the persona."** Turn 1 (22328) is
  clean: its inbound thread-context block contains no `·` anywhere — the first
  `·` in this session's history *is* 22328. (Turns 2 and 3 do carry
  `[assistant] ·` in their thread context, so they are partly self-reinforced —
  see gaps.)
- **"Some other gate produced it."** The only gates between Slack and the model
  are auth (returns silently), allow_bots (returns silently), allowed_channels
  (returns silently), mention checks (return silently). Every adapter-side
  rejection in this path is a bare `return` — none of them writes to Slack.
- **"SOUL.md isn't loaded for gateway/Slack turns."** `agent/system_prompt.py:189-197`
  puts SOUL.md in the stable prefix, and M1 proves it empirically for this exact
  session.

---

## Corrections to the RCA (none change the verdict)

1. **H4 over-states the restart.** The RCA calls the 11:35:26 gateway restart
   "the activation event". `.env` is re-loaded **per turn**, not only at start:
   `gateway/run.py:1838 _reload_runtime_env_preserving_config_authority()`
   ("Gateway processes are long-lived, so per-turn code reloads ~/.hermes/.env")
   called from `_current_max_iterations()` at `:1907`. So Gaetan's 11:32:02 turn
   would already have pulled the widened allowlist into `os.environ` ~3.5 min
   before the restart. No non-Gaetan message arrived between 11:30:27 and
   11:35:26, so the two hypotheses are observationally indistinguishable here —
   but the mechanism favours "the `.env` write activated it, the restart was
   incidental", which is closer to the sibling probes' reading than to the RCA's.
   Root cause unaffected: `deleg_97b39490`'s edit is the cause either way.

2. **Consequence for the fix procedure.** Reverting `SLACK_ALLOWED_USERS` in
   `.env` should take effect on the next turn without a restart (M-1 above).
   Deleting `slack.allowed_channels` probably does **need** the restart — the
   adapter reads it from `self.config.extra` (`adapter.py:8395`), populated at
   adapter init, not from a per-turn file read. I did not test either; the RCA's
   "then `hermes gateway restart`" is the safe order regardless.

3. **`U0BBH85NAKH` "inert" is unverified.** I did not read
   `_slack_allow_bots()`'s default, so the RCA's claim that Tars' own bot id in
   the allowlist is harmless is untested. Irrelevant to the fix (it gets removed
   either way).

## Bounded gaps I am leaving open

- Turns 2 and 3 are contaminated by `[assistant] ·` appearing in their inbound
  thread-context block — a model copying its own last reply is a sufficient
  explanation for those two. Turn 1 is not, and turn 1 is the one that matters.
- No chain-of-thought (`reasoning` is NULL). Same residual as the RCA states.
  M1 shrinks it: the collision is now against a prompt with exactly one `·`.
- I verified the *effect* of `deleg_97b39490` (file contents, mtimes, the
  delegation's own report inside the session transcript at msg 22250/22259,
  which quotes `SLACK_ALLOWED_USERS=U08BDJAMSRZ,U7XJ4K631,U0BQTJUK2F2,U0BBH85NAKH`),
  not its transcript.
- Only `C0BQCB58ATW` on 2026-08-14 was traced; other surfaces unswept, as the
  RCA already says.
