# Adversarial verification — alternative-explanation lens

Target: `status/probes/tars-dot-reply-2026-08-14-RCA.md`
Mandate: find the best competing explanation the RCA dismissed or never
considered and argue for it; default to REFUTED if the RCA cannot rule it out.

Read-only. Nothing changed on 192.168.0.9 or in the repo except this file.

## Verdict

**NOT REFUTED.** I built four competing explanations, ran my own VM checks to
feed them, and three died on evidence the RCA did not have. The fourth
survives — but it does not touch the root cause or the fix; it bounds one
*corollary* the RCA states with more confidence than its N=3 sample can carry.

Root cause as stated (rule 4's `·` terminal, made reachable by
`deleg_97b39490`'s 11:30 `.env` widening, activated by the 11:35:26 restart)
stands, and is now better evidenced than when it was written.

## What I verified myself (new evidence, none of it in the three probes)

| # | check | result |
|---|---|---|
| V1 | `grep -c 'response=1 chars'` over **all** rotated `gateway.log*` | **3** — the only 1-char Slack sends in Tars' entire life |
| V2 | `user=U…` histogram over **all** `gateway.log*` | `343 U08BDJAMSRZ`, `3 U7XJ4K631`, **nobody else ever** |
| V3 | `Early reject` by user, whole history | 14 distinct users, 826 rejects; Oli `U7XJ4K631` 134 of them, in `C0BFQ5WFYTB` (30), `C0BQCB58ATW` (5), `C08RWSTU9LK` (3) |
| V4 | Oli's early rejects on 2026-08-14 | 08:56, 10:35, 11:22:14, 11:25:02, 11:28:22, 11:29:44, 11:29:52 — **last one 5.5 min before the restart, zero after** |
| V5 | `grep -rIn '·'` (U+00B7) over `agent/ gateway/ hermes_cli/ plugins/` in `~/.hermes/hermes-agent` | ~40 hits, **all** UI separators (`f"… · …"`), `_FREE_GLYPH` in `context_breakdown.py:175`, `"missing": "·"` in `lsp/cli.py:161`, one regex char-class. **No chat/reply path** |
| V6 | `grep -n '·' ~/.hermes/SOUL.md` | exactly **one** line, `:215` (rule 4) |
| V7 | rule 4 in `~/.hermes/SOUL.md.bak-soulv2-2026-08-13` (`:105-108`) vs live (`:212-215`) | **byte-identical** |
| V8 | `stat` on live SOUL.md | mtime **2026-08-14 11:46:24** — SOUL.md was rewritten *mid-incident*, between `·`#2 and `·`#3 |
| V9 | key-only diff of live `config.yaml` vs newest backup (`…bak-2026-08-13-lcm-removal`, Aug 13 15:15) | only `allowed_channels` added; other deltas are the Aug-13 lcm removal + engine/mode cutover. No second behaviour-affecting key from today |
| V10 | `grep '^SLACK_ALLOWED_USERS=' ~/.hermes/.env`, mtimes | `U08BDJAMSRZ,U7XJ4K631,U0BQTJUK2F2,U0BBH85NAKH`; `.env` 11:30:27, `config.yaml` 11:31:01, both mode 600 |

## The four alternatives

### ALT-1 — Hermes substitutes `·` for an empty/suppressed completion. DEAD, and the RCA's own kill was aimed at the wrong character.

This is the alternative with the most room to live, because **nobody had
actually looked for it.** The `-vm-logs.md` probe grepped `'"\."'` and
`'= "\."'` — **ASCII U+002E only** — and the RCA inherited that grep as E4
("no `"."` placeholder constant exists"). A U+00B7 constant in the Hermes tree
would have satisfied every observation the RCA uses (byte-exactness included,
if SOUL.md's author had copied a Hermes convention) and would have moved the
root cause into vendor code.

I ran the grep the probes did not: U+00B7 does occur ~40 times in the Hermes
tree (V5) — and every occurrence is a display separator, a context-bar glyph
(`_FREE_GLYPH`), an LSP "missing" marker, or a punctuation char-class. None is
reachable from a gateway reply body. Combined with the persisted assistant row
(`content='·'`, `finish_reason=stop`, written before the Slack send) the
substitution story is dead. **Net: the RCA's conclusion survives, its E4
evidence was weaker than stated and is now repaired.**

### ALT-2 — the enabling change was something other than the `.env` widening (mention gate, `allowed_channels`, restart-only, engine cutover). DEAD.

V4 is the decider: Oli hit the *user* gate 7 times on 2026-08-14, the last at
11:29:52, and **zero** times after 11:35:26 — while his messages started
reaching the model at 11:41. `allowed_channels` cannot explain the pre-11:35
rejects (it did not exist before 11:31, and an absent whitelist allows all),
and V9 shows no other behaviour key changed today. V2 shows the widening is
also *necessary*: in 7 days and 343 inbound Slack turns, no non-Gaetan sender
had ever reached the model before today.

### ALT-3 — content guardrail / model policy on "give me money" (the `-slack.md` and `-vm-logs.md` conclusion). DEAD, but the RCA's second killer fact is soft.

The glyph provenance (V5+V6+V7) kills it: a model improvising a minimal
non-answer does not land on U+00B7 three times when U+00B7 is the exact literal
its system prompt hands it for this exact case.

The RCA's *other* killer — "turn 2 was a bare `@Tars` + image, **nothing
manipulative**" — is not evidence. Nobody opened `image.png` (`F0BR3659MJ4`,
473.9 KB), and the image *was* fed to the model (`Image routing: native …
1 image(s) will be attached`). "Nothing manipulative" is an assumption about
unread content. Delete that leg; the glyph leg carries the claim alone.

### ALT-4 — SURVIVES: `·` is the resolution of an instruction conflict, not an unconditional sender gate.

The competing reading: rule 4 supplied the *character*, but the *trigger* was
not merely "sender ≠ Gaetan". All three `·` turns landed inside a live conflict
— Gaetan's standing "answer every message of Oli or Ilo with 'oe ok'" (thread
root, quoted back to the model in every one of the three inbounds as
`reply_to_text=`) versus a hard rule forbidding any answer to Oli. `·` is the
only exit the persona offers from that conflict. Under this reading a *normal*
question from Oli might have been answered, and the RCA's fourth why — "the
widening bought nothing; rule 4 guarantees they cannot get content" — would be
false.

**The RCA cannot rule this out, and neither can I.** V2 makes it structural:
there are exactly **3** non-Gaetan model-reaching turns in Tars' entire
history, and all 3 sit inside the "oe ok" window. Sender-identity and
instruction-conflict are 100% confounded in the sample; the perfect 3/3 vs 0/13
partition the RCA leans on is equally predicted by both readings. The RCA's
discriminator (turn 2 = bare mention) does not separate them — turn 2 is inside
the same window, same sender, same standing instruction.

**Why this does not refute the RCA:** it leaves the mechanism (rule 4 is the
source of the glyph), the enabling event (the 11:30 widening + 11:35 restart),
and the fix (revert both values; do not touch SOUL; do not patch Hermes)
untouched. Reverting gate 1 restores zero output either way. What it demotes is
a *scope claim* — "every non-Gaetan sender gets `·`" and "the widening bought
nothing" are extrapolations from a fully confounded N=3, not findings. They
should be labelled as such, and the one probe that would settle it is cheap and
worth doing if Gaetan ever revisits the Oli/Ilo fork: with the allowlist
widened, have Oli ask Tars an ordinary question with the "oe ok" instruction
absent from the thread.

## Corrections and additions the RCA should absorb

1. **E4 was grepped for the wrong codepoint.** ASCII-only. Repaired by V5 —
   state it as U+00B7-verified, not `"."`-verified.
2. **"Nothing manipulative" (H5 kill, turn 2) is unverified** — the image was
   attached and read by the model, and never inspected. Drop the leg.
3. **"Other surfaces not swept" (Confidence §) is now closed, not open:** V2
   shows only 3 non-Gaetan inbounds ever, all in `C0BQCB58ATW`, and V1 shows
   only 3 1-char sends ever. Oli/Ilo triggered `·` nowhere else, including
   DMs. Same greps also confirm the incident is *over-scoped* by the complaint:
   3 of 346 lifetime Slack turns.
4. **SOUL.md was rewritten mid-incident** (V8, mtime 11:46:24, between `·`#2 and
   `·`#3 — PR #65's standing-corrections append). The RCA verified rule 4 in the
   *post*-edit file and attributed the 11:41 and 11:43 turns to it. V7 closes
   that gap: rule 4 is byte-identical in the 2026-08-13 backup. Worth stating,
   because "I read the live file" was not sufficient for two of the three turns.
5. **Blast radius of the collateral finding is bigger than reported, and the
   two config changes partially cancel.** V3: Oli alone was early-rejected 134
   times, 30 of them in `C0BFQ5WFYTB` — a channel absent from `docs/facts.md`
   and from the RCA. With `SLACK_ALLOWED_USERS` widened, every future Oli/Ilo
   message in *any* channel Tars sits in would reach the model and draw a `·`;
   the `allowed_channels: C0BQCB58ATW` narrowing is currently the only thing
   containing that. **Order matters in the fix: revert `.env` first (or in the
   same locked edit), never delete `allowed_channels` alone.** The RCA's fix
   does both, so it is correct — but it does not flag that the halves are not
   independently safe.
6. 14 distinct users have been early-rejected (V3), i.e. gate 1 is load-bearing
   daily, not a theoretical backstop. Reverting it is the right call.

## Confidence

High that the RCA's root cause holds. Its three legs are independently
reproducible and I reproduced all three plus five checks it did not run. The
residual is not "wrong cause" — it is ALT-4's scope confound and the
acknowledged absence of chain-of-thought proof (`reasoning=None`).
