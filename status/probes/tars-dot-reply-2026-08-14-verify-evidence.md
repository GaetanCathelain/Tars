# Adversarial verification — EVIDENCE lens on `tars-dot-reply-2026-08-14-RCA.md`

Verifier: independent subagent, 2026-08-14. Read-only: no VM file written, no
config touched, no session/model turn run, no repo file changed except this one.
Question asked: **does every load-bearing claim in the RCA cite a log line,
`file:line` or Slack message that actually exists?**

**Verdict: NOT REFUTED.** The root cause survives. I re-ran 12 of the RCA's
citations against the VM and the repo; 9 are exact, 3 are defective. All 3
defects are citation-level, and one is a factual error — none of them moves the
root cause.

---

## 1. Re-run citations — VERIFIED

| # | RCA claim | What I ran | Result |
|---|---|---|---|
| V1 | E1 sender partition, `gateway.log` | `grep -E "^2026-08-14 1[12]:" gateway.log \| grep C0BQCB58ATW` | **Exact.** Every row of the RCA's table is a real log line, timestamps to the second. |
| V2 | "3/3 non-Gaetan → `·`, 0/13 Gaetan" | day-wide `user=…chat=…` histogram + all `response=1 chars` lines | **Stronger than claimed.** Whole day, every channel: inbound = 13 Gaetan/`C0BQCB58ATW` + 25 Gaetan/`D0BBYNM01BL` + **3 Oli/`C0BQCB58ATW`**; `response=1 chars` occurs exactly 3× and they are Oli's 3. Partition is 3/3 vs **0/38**. |
| V3 | live `~/.hermes/SOUL.md:212-215` rule 4, codepoint `0xb7` | `sed -n 215p \| python3` codepoint dump; `md5sum` of lines 212-215 on both sides | **Byte-identical.** Live and repo lines 212-215 share md5 `c1f2caf0…`; line 215 non-ASCII = `['0xb7','0x2014']`. Live file 319 lines vs repo 318 → the claimed "one line of drift" is exactly one line. |
| V4 | `.env` / `config.yaml` values + mtimes | `grep ^SLACK_ALLOWED_USERS`, `grep -n allowed_channels`, `ls --time-style=full-iso` | **Exact.** `SLACK_ALLOWED_USERS=U08BDJAMSRZ,U7XJ4K631,U0BQTJUK2F2,U0BBH85NAKH` (`.env` mtime `11:30:27Z`); `allowed_channels: C0BQCB58ATW` at `config.yaml:136` (mtime `11:31:01Z`, mode `600`). |
| V5 | `docs/facts.md:59` invariant | `sed -n 40,62p docs/facts.md` | **Verbatim.** "The adapter allowlist rejects strangers before the model, so the model-level rule is a backstop only." |
| V6 | `wf4-report.md:97`, `lane-a.md:349-355`, `tars-profile.md:52-55` | `sed -n` each range | **All verbatim**, including the `❌ never exercised → probe 19` row and the "identity mapping + `·` rewording deployed" log entry. |
| V7 | `plugins/platforms/slack/adapter.py:8387-8399` docstring + gate `:5652-5658` | `sed -n` both ranges on the VM | **Exact, path included.** The quoted whitelist docstring is verbatim; the gate really is nested under `if not is_one_to_one_dm and bot_uid:` — the "Gaetan's DM is unaffected" claim is source-true, not assumed. |
| V8 | `agent_runtime_helpers.py:1379` "Fabricating `"."` … lies in the history" | `sed -n 1375,1382p` | **Verbatim, at that line.** |
| V9 | H1 refutation: `turn_finalizer.py` explainer skips `text_response` exits | `grep -n _is_partial_fragment\|_is_empty_terminal\|text_response` | **Present:** `:519 and not str(_turn_exit_reason).startswith("text_response")`, plus `:502` comment "``text_response(...)`` exits never produce an explanation". |

## 2. Gaps in the RCA that I closed for it (asserted, not cited — now evidenced)

- **"the live system prompt" contains SOUL.md.** The RCA asserts it; nothing in
  the three probes evidences it. Verified: `agent/prompt_builder.py:1961-2015`
  loads `SOUL.md` from `HERMES_HOME` and injects it as the identity slot
  (`:44` "…SOUL.md before they get injected into the system prompt").
- **The sender prefix rule 4 keys on is really constructed.**
  `gateway/run.py:16032`: `f"{_safe_user_name} | Slack user <@{source.user_id}>"`
  — matches the `[U7XJ4K631 | Slack user <@U7XJ4K631>]` form the RCA quotes.
- **The widening actually took effect, and gate 1 is otherwise intact.** Oli's
  *only* other Early rejects today are `GQ07CQXT7` at `08:56:43` and `10:35:15`
  — both pre-change; **zero** rejects of `U7XJ4K631` anywhere after 11:30. Other
  unlisted users are still rejected after the restart (`11:43:37`, `12:27:39`,
  `12:36:52`), so the allowlist was widened, not disabled. This is the single
  strongest corroboration of the causal chain and neither the RCA nor the probes
  ran it.
- **`deleg_97b39490` has no repo artifact** — `grep -rn` over the worktree
  returns nothing outside the four probe files. RCA's own caveat is true.

## 3. Defects found

**D1 — factual error. E3's "`allowed_channels` ← NEW key; absent from every
backup through 2026-08-13 15:15" is FALSE.**

```
$ grep -l allowed_channels ~/.hermes/config.yaml*
/home/gaetan/.hermes/config.yaml
/home/gaetan/.hermes/config.yaml.bak-boundaries-20260810-143638
      226-slack:
      227:  allowed_channels: C0BP2GZUFSR,C0BFQ5WFYTB
```

The key existed on 2026-08-10 (reporting channel + the team channel), was dropped
somewhere around the 08-12 DM cutover (absent from the 08-12 17:10 and all 08-13
backups), and was **re-introduced** today with a different value. So today's edit
is a re-introduction, not a novel key, and the "nobody asked for this narrowing"
framing loses its force — narrowing by `allowed_channels` is an established
pattern in this deployment (Tars was equally deaf in `#gcn-sandbox` between
08-10 and 08-12). The *fix* is unaffected: deleting the key restores the state
that has held since 08-12.

**D2 — wrong pointer. `docs/specs/wf4-probes.md:100` does not show probe 19 was
stubbed.** Line 100 reads "probe tests (that is probe 19)." — it is inside probe
3's prose. The real evidence is `## 19. SOUL rule 4's ``·`` terminal at the model
layer` at **`wf4-probes.md:431-450`**, which is indeed an unrun stub (it even
predicts this exact situation: "`·` is defence-in-depth that Slack can never
reach"). Claim true, citation wrong; inherited verbatim from the `-repo.md` probe.

**D3 — E1's table is one row short.** The log has 13 Gaetan inbounds and 13
responses in `C0BQCB58ATW`; the table lists 12 (missing `11:29:24 → response=457
chars`). Cosmetic — the `0 out of 13` count in the prose is right.

## 4. Claims I could not close (bounded, flagged)

- **"the 11:35:26 gateway restart loaded it" is unmeasured.** `.env` mtime
  `11:30:27` and restart `11:35:26` are facts; no non-Gaetan message arrived
  between 11:31 and 11:35, so nothing in the logs distinguishes "the restart
  activated it" from "a live reload picked it up at 11:31". The only support is
  Tars' own in-thread claim (Slack #31, 11:32:36, "live PID still on old policy")
  — an agent assertion, not a log line. **Does not touch the root cause**: the
  widening is the enabler under either timing.
- **"the only place in repo or VM that prescribes `·`"** is broader than what was
  checked: U+00B7 appears in ~19 other files under `~/.hermes` (skills, HTML
  templates, PDFs) as a separator glyph. None *prescribes it as a reply*, and
  skills are not injected wholesale into the system prompt — but the uniqueness
  claim as written was not tested at that scope.
- **The persisted session row** (`content='·'`, `codepoints=['0xb7']`,
  `reasoning=None`) is second-hand from `-vm-logs.md`; re-running
  `hermes sessions export` would have written to the VM, which my brief forbids.
  It is corroborated indirectly by `[Slack] Sending response (1 chars)` and by
  the byte-identical literal in the prompt.
- **No chain-of-thought proof** — the RCA already states this and bounds it at
  ~3%. I agree with the bound: three different inputs (two adversarial asks and a
  captionless image), three identical U+00B7 emissions, from a prompt that
  literally instructs U+00B7 for this sender class, is not coincidence.

## 5. Bottom line

Every load-bearing leg of the root cause — the sender partition, the byte-exact
glyph in the live persona file, the config mtimes, and the source-level
elimination of a Hermes-side fabricator — is backed by evidence that exists and
that I reproduced independently. The two supporting details that are wrong (D1,
D2) sit in the *collateral finding* and the *3-whys narrative*, not in the causal
chain. **Not refuted.**

Recommended edits before the RCA is treated as final: correct E3's "absent from
every backup" sentence (D1), repoint `wf4-probes.md:100` → `:431` (D2), and
downgrade "the 11:35:26 restart loaded it" to "activated no later than 11:35:26"
unless someone can produce a reload log line.
