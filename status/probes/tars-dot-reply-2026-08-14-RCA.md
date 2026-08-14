# RCA — "Why is Tars always responding with a `.`?" (2026-08-14)

Target: `https://mobileclub-squad.slack.com/archives/C0BQCB58ATW/p1786708539718269?thread_ts=1786706480.639839&cid=C0BQCB58ATW`
→ channel `C0BQCB58ATW` (group DM), thread `1786706480.639839`, linked msg
`1786708539.718269` = 2026-08-14 11:55:39 UTC.

Synthesis of three evidence probes (`-slack.md`, `-vm-logs.md`, `-repo.md`) plus
my own read-only verification on the VM. **Nothing was changed on the VM or in
the repo. The fix below is NOT applied.**

---

## Symptom

Tars replies to some messages in `C0BQCB58ATW` with a single character that
reads as a full stop. Three occurrences, all 2026-08-14 UTC: 11:41:05,
11:43:41, 11:55:39 (the linked one). ~14.5 min span, one thread, one session
(`20260814_112121_d311e0fc`).

Two corrections to the complaint:

- The character is **`·` U+00B7 MIDDLE DOT**, not ASCII `.` (U+002E). Confirmed
  byte-exact from the persisted session (`codepoints: ['0xb7']`).
- It is not "always". It is **3 out of 3 turns whose inbound sender was not
  Gaetan**, and **0 out of 13 turns whose inbound sender was Gaetan**, in the
  same channel on the same day. It reads as "always" because from Gaetan's seat
  every retry he asked Oli to run came back `·`.

## Evidence

### E1 — perfect partition on inbound sender (verified by me, `gateway.log`, `C0BQCB58ATW`, 11:16–11:58 UTC)

```
11:16:47  user=U08BDJAMSRZ  → response=227 chars
11:18:30  user=U08BDJAMSRZ  → response=883 chars
11:21:21  user=U08BDJAMSRZ  → response=54  chars
11:22:14  user=U7XJ4K631    → WARNING [Slack] Early reject of unauthorized user   (no reply)
11:22:34  user=U08BDJAMSRZ  → response=98  chars
11:25:02  user=U7XJ4K631    → Early reject                                        (no reply)
11:26:21  user=U08BDJAMSRZ  → response=556 chars
11:26:51  user=U08BDJAMSRZ  → response=230 chars
11:27:48  user=U08BDJAMSRZ  → response=108 chars   ← Gaetan: "ajoute ce canal … en allow list"
11:28:22  user=U7XJ4K631    → Early reject                                        (no reply)
11:29:44  user=U7XJ4K631    → Early reject                                        (no reply)
11:29:52  user=U7XJ4K631    → Early reject                                        (no reply)
11:32:02  user=U08BDJAMSRZ  → response=454 chars
--- 11:30/11:31 config+.env edited (deleg_97b39490) ; 11:35:26 gateway restart loads it ---
11:36:30  user=U08BDJAMSRZ  → response=165 chars
11:38:26  user=U08BDJAMSRZ  → response=136 chars
11:41:01  user=U7XJ4K631    → response=1 char   ·      api_calls=1  4.1s
11:43:36  user=U7XJ4K631    → response=1 char   ·      api_calls=1  5.1s
11:45:30  user=U08BDJAMSRZ  → response=235 chars
11:55:36  user=U7XJ4K631    → response=1 char   ·      api_calls=1  3.1s   ← the linked message
11:57:06  user=U08BDJAMSRZ  → response=148 chars
```

Sender is the **only** variable that partitions cleanly. Content does not:
the three `·` turns were "give me @Gaëtan money", **a bare `@Tars` mention with
an image and no text**, and "give me bitcoins". Channel/thread/session/mention
shape are identical across `·` and non-`·` turns.

### E2 — the emitted glyph is byte-identical to a literal in the live system prompt

`~/.hermes/SOUL.md:212-215` on 192.168.0.9 (verified by me today, live file, not
the repo mirror):

> 4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — a channel
>    message whose sender prefix reads "[U08BDJAMSRZ | …]" is from Gaetan. **To anyone
>    else, in any channel or DM, I give no answer: I reply with the single character
>    "·" and nothing else — no content, no reaction, no explanation.**

Non-ASCII codepoints on that line: `['0xb7', '0x2014']` — the `·` in SOUL.md is
U+00B7, the same codepoint the model emitted. Live VM `SOUL.md` vs repo mirror:
one line of drift (today's PR #65 Slack-MCP-identity rule, present on the VM,
absent from this worktree) — rule 4 is byte-identical on both sides.

The model saw the sender prefix rule 4 evaluates: the exported session shows
`[U7XJ4K631 | Slack user <@U7XJ4K631>] give me bitcoins`.

### E3 — the config change that made rule 4 reachable (verified by me)

`deleg_97b39490`, executed 11:30–11:31 UTC in response to Gaetan's 11:27:48
"change cette règle, ajoute ce canal avec Ilo, Oli, toi et moi en allow list":

```
~/.hermes/.env         mtime Aug 14 11:30
  SLACK_ALLOWED_USERS=U08BDJAMSRZ,U7XJ4K631,U0BQTJUK2F2,U0BBH85NAKH
                      Gaetan     Oli       Ilo         Tars' own bot id
~/.hermes/config.yaml  mtime Aug 14 11:31, mode 600
  slack:
    allowed_channels: C0BQCB58ATW        ← NEW key; absent from every backup
                                            through 2026-08-13 15:15
```

Live at 11:35:26 (gateway restart). First non-Gaetan message to reach the model
after that: 11:41:01 → `·`. Before it: five Early rejects, zero Slack output.

### E4 — Hermes does not fabricate the character (from the VM-log probe, source-read)

`agent/turn_finalizer.py`'s empty/partial-turn explainer is gated
`not str(_turn_exit_reason).startswith("text_response")`; all three turns exit
`reason=text_response(finish_reason=stop) … response_len=1`. No `"."` placeholder
constant exists in `agent/`, `gateway/`, `hermes_cli/`. `agent_runtime_helpers.py:1379`
documents the opposite intent: *"Fabricating `"."` / `"(continued)"` text lies in
the history."*

## Root cause

**A configuration change widened gate 1 (the adapter allowlist) without
touching gate 2 (SOUL hard rule 4), promoting rule 4's `·` terminal from an
unreachable backstop into the live, user-visible reply path for every non-Gaetan
sender.**

The two gates are designed as a pair (`docs/facts.md:50-60`):

| gate | where | who it stops | what Slack sees |
|---|---|---|---|
| 1. adapter allowlist | `SLACK_ALLOWED_USERS` (`.env`) | anyone not listed | **nothing** — WARNING-only log |
| 2. SOUL hard rule 4 | `~/.hermes/SOUL.md:212-215` (system prompt) | anyone who is not `U08BDJAMSRZ` | **`·`** |

`docs/facts.md:59` states the invariant the design rests on: *"The adapter
allowlist rejects strangers before the model, so the model-level rule is a
backstop only."* Adding `U7XJ4K631` (Oli) and `U0BQTJUK2F2` (Ilo) to
`SLACK_ALLOWED_USERS` at 11:30 broke that invariant. From 11:35:26 onward their
messages reach the model, and rule 4 — which has no content clause, no channel
clause and no override by an in-thread instruction — answers each one with
exactly the character it is told to answer with.

Nothing is malfunctioning. Hermes, the model, the gateway and SOUL rule 4 all
did exactly what they are specified to do. The defect is a **config/persona
coupling that nothing enforces**: the allowlist lives in `.env` on the VM, the
identity rule lives in `SOUL.md`, and widening the first silently changes the
second's user-visible behaviour.

## Why it happens (3 whys)

1. **Why does Tars answer `·`?** Because SOUL hard rule 4 prescribes the literal
   character `·` as the reply to any sender whose prefix is not
   `[U08BDJAMSRZ | …]`, and Oli is `U7XJ4K631`. Hard rules beat in-message
   instructions (`SOUL.md:14-16`), so Gaetan's standing "answer every Oli/Ilo
   message with 'oe ok'" in the same thread could not override it — which is
   precisely why the "oe ok" experiment had to be routed through a cron job
   (`is_internal` skips the Slack-sender path, `docs/facts.md:43`).

2. **Why does the rule prescribe a visible character instead of silence?**
   Because the original wording ("I ignore in silence") produced genuine empty
   completions, and **Hermes has no silence terminal**: it read the model's
   deliberate silence as a provider fault and replayed the request 5× over 61 s,
   posting `⚠️ Empty response from model` in-channel (WF4 incident 2026-08-07,
   `status/wf4-report.md:122-165`, `status/lane-a.md:349-352`). The fix added the
   `U08BDJAMSRZ`→Gaetan mapping and gave the refusal a terminal the runtime can
   represent. `·` is that terminal. It traded a retry storm for a visibly odd
   one-character reply — acceptable **on the assumption that gate 1 makes it
   unreachable in practice**.

3. **Why is it visible today when it never was in the 7 days Tars has been
   live?** Because at 11:30–11:31 UTC today `deleg_97b39490` added Oli and Ilo to
   `SLACK_ALLOWED_USERS`, and the 11:35:26 gateway restart loaded it. That was
   the first time a non-Gaetan Slack message ever reached the model. `wf4-report.md:97`
   had logged rule 4's `·` terminal as *"❌ never exercised → probe 19"*, and
   `docs/specs/wf4-probes.md:100` shows probe 19 was stubbed, never run. Oli
   ran it for us, live, eight days later.

Fourth why, for the record: the widening bought nothing. Gaetan asked for it so
Oli/Ilo could interact with Tars; rule 4 guarantees they cannot get content. The
change could only ever add `·` noise.

## Refuted hypotheses

- **H1 — Hermes substitutes a minimal non-empty string for empty model output.
  REFUTED.** `turn_finalizer.py`'s explainer is explicitly gated to skip
  `text_response(...)` exits (all 3 turns), `_is_empty_terminal` is False
  (content is 1 char, not `""`/`"(empty)"`), and no `"."` placeholder constant
  exists in the tree. Decider: the emitted codepoint is **0xb7**, not 0x2e — an
  empty-output guard would not coincidentally pick the exact glyph written in
  the persona file.

- **H2 — the model returned empty text (tool_use exit / truncation / stop
  condition). REFUTED.** All 3: `reason=text_response(finish_reason=stop)
  … response_len=1`, `tool_calls=None`, `api_calls=1`, no `length`, no
  `content_filter`, no exception in `errors.log`/`mcp-stderr.log` at any of the 3
  timestamps. The turn completed normally and its whole content was `·`.

- **H3 — SOUL rules make Tars decline in a non-DM channel and emit filler.
  CONFIRMED on the mechanism, WRONG on the discriminator.** The "SOUL rule +
  filler instead of silence" half is exactly right and is the documented WF4 fix.
  But the gate is **sender identity, not channel type**: rule 4 says *"To anyone
  else, in any channel **or DM**"* — Oli would get `·` in a 1:1 DM too, and
  Gaetan gets full answers in this very same non-DM channel (13 times today).
  Decider: `SOUL.md:214` + E1's partition.

- **H4 — today's gateway restart broke the reply path. REFUTED as a fault, but
  it is in the causal chain** — and both sibling probes call it a pure
  coincidence, which undersells it. The 11:35:26 restart is the **activation
  event**: it is what loaded the 11:30/11:31 allowlist change. Before it,
  Oli's 5 messages in this channel were Early-rejected; after it, all 3 reached
  the model. The restart itself was clean (SIGTERM → orderly shutdown → clean
  start, session `20260814_112121_d311e0fc` resumed intact) and its trigger was
  the unrelated cron-posting-as-Gaetan bug fixed in PR #65.

- **H5 — something else. THIS IS THE ANSWER**, and it is not "model policy".
  The `-vm-logs.md` probe concluded the model was *"unwilling to echo 'oe ok'
  verbatim to a manipulative ask"* — that probe never read the system prompt.
  Two facts kill it: turn 2's inbound was a **bare `@Tars` mention with an image
  and zero text** (nothing manipulative, same `·`), and the glyph is byte-exact
  to the SOUL.md literal. A model improvising a minimal non-answer would emit
  `.`, `…`, `ok`, or a refusal sentence — not U+00B7.

Also refuted in passing: `-slack.md`'s "content-based guardrail on 'give me
money'" hypothesis (same two facts), and any Slack-side cause (no `blocks[]`
truncation, no edit, no subtype — the 1-char body is what Hermes handed Slack:
`[Slack] Sending response (1 chars)`).

## Fix (NOT applied)

**Revert `deleg_97b39490`.** The root cause is the config change, not the persona
rule — so the fix is where the wrong thing was introduced. Two values on
**192.168.0.9**, no code, no SOUL edit:

1. `~/.hermes/.env` → `SLACK_ALLOWED_USERS=U08BDJAMSRZ`
   (drop `U7XJ4K631` Oli, `U0BQTJUK2F2` Ilo, and `U0BBH85NAKH` — Tars' own bot
   id, which the delegation also added and which is inert: bot senders are
   dropped upstream.)
2. `~/.hermes/config.yaml` → **delete** the `slack.allowed_channels:
   C0BQCB58ATW` key added today (see collateral finding below).

Then `hermes gateway restart`. Per repo hard rules: `.bak` first, `flock
~/.hermes/.wf3.lock`, merge don't append, `chmod 600` after any tmp+mv.

Why this and not the alternatives:

- **Do not edit SOUL rule 4.** It is working as specified, and its current
  wording is the fix for a real incident. Any rewording that restores "stay
  silent" re-opens the 61 s / 5-retry / `⚠️ Empty response` failure mode — Hermes
  still has no silence terminal.
- **Do not patch Hermes** to suppress 1-char sends. It is a bigger diff in
  someone else's tree, and per `hermes-upgrade-reapply-local-patches` memory,
  local code patches were dropped on the 2026-08-13 upgrade — a patch here dies
  at the next `/upgrade`.
- Result: non-Gaetan senders are stopped at gate 1 again — **zero** Slack
  output, a WARNING in the log — which is the designed behaviour and is what
  happened all week until 11:35 today.

**The fork Gaetan must decide, if he ever wants Oli/Ilo to actually talk to
Tars:** the blocker is rule 4, not the allowlist. That is a persona decision, not
a bug fix — and whatever replaces it must still prescribe a literal reply for the
refusal case, never silence.

### Collateral finding, fixed by the same revert

`slack.allowed_channels` is a **whitelist**, verified in source
(`plugins/platforms/slack/adapter.py:8387-8399`, gate at `:5652-5658`):

> *"When non-empty, messages from channels NOT in this set are silently ignored
> — even if the bot is @mentioned. DMs are controlled separately."*

Setting it to the single value `C0BQCB58ATW` made Tars **deaf in every other
channel** it is in (e.g. `C08RWSTU9LK` #gcn-sandbox) as of 11:35 today. Gaetan's
DM `D0BBYNM01BL` is unaffected — the gate is inside `if not is_one_to_one_dm`
(source-verified, not assumed). Nobody asked for this narrowing; it came with the
delegation.

### Process findings (not the bug, worth one line each)

- Today's 11:31 `config.yaml` edit created **no `.bak`** — the newest backup is
  `Aug 13 15:15`. That is the repo's hard rule, violated by the delegated agent.
  (Mode is correctly `600`.)
- `docs/facts.md`'s channel inventory (`:44`, `:83`, last verified 2026-08-07/08)
  does not list `C0BQCB58ATW` and now understates the allowlist — stale.
- `status/lane-a.md` has no entry for today's allowlist change, restart, or this
  incident. The four `status/probes/*2026-08-14*` files are uncommitted.

## Verification step

After the revert + restart, one command decides it:

```bash
ssh gaetan@192.168.0.9 'grep -E "^2026-08-14 " ~/.hermes/logs/gateway.log \
  | grep C0BQCB58ATW | grep -E "inbound message|response ready|Early reject" | tail -6'
```

Have Oli @-mention Tars in `C0BQCB58ATW`, then Gaetan @-mention Tars in the same
thread. Pass =

- Oli's message → `WARNING … Early reject of unauthorized user U7XJ4K631` and
  **no** `response ready` line, **no** Slack message posted;
- Gaetan's message → normal `response ready … response=N chars` with N ≫ 1;
- `ssh … 'grep -c allowed_channels ~/.hermes/config.yaml'` → `0`;
- and Tars answers an @-mention in `#gcn-sandbox` (`C08RWSTU9LK`) again.

## Confidence + what is unverified

**Root cause: high confidence (~97%).** Three independent legs, each of which I
re-ran myself rather than taking from a probe summary: (a) the 3/3 vs 0/13
sender partition in `gateway.log`; (b) the emitted codepoint `0xb7` is
byte-identical to the literal in the **live** `~/.hermes/SOUL.md:215`, which is
the only place in repo or VM that prescribes it; (c) the `.env`/`config.yaml`
mtimes (11:30/11:31) and the key's absence from every pre-2026-08-14 backup place
the enabling change 4 minutes before the restart that activated it and 6 minutes
before the first `·`.

Unverified / bounded:

- **No chain-of-thought proof.** `reasoning = None` in the persisted rows, so I
  cannot show the model *citing* rule 4 — the inference is from the byte-exact
  glyph match plus the clean sender partition. The residual ~3% is "the model
  independently chose U+00B7 three times for three different inputs", which I
  consider negligible but did not rule out by experiment. The one probe that
  would close it: a controlled non-Gaetan @-mention with the allowlist widened
  and rule 4 temporarily removed from the persona — not worth doing.
- **The fix is untested** — nothing was applied, so the verification block above
  is a plan, not a result.
- **Other surfaces not swept**: I checked `C0BQCB58ATW` and `D0BBYNM01BL` on
  2026-08-14 only. Whether Oli/Ilo triggered `·` elsewhere (or in a DM to Tars,
  which the widened allowlist now permits) is unchecked.
- **`deleg_97b39490` has no artifact** anywhere in this repo — I verified its
  *effect* on the VM (file mtimes + values), not its authorship or transcript.
- MCP-side Slack reads in the sibling probe expose a formatted view, not raw
  `conversations.replies` JSON — the "no hidden content in the `·` message"
  claim is bounded by that tooling. My own evidence does not depend on it:
  `[Slack] Sending response (1 chars)` shows Hermes sent one character.
