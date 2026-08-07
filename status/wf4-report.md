# WF4 — exercised report

Completeness critic, 2026-08-07. Spec: `docs/specs/wf4-probes.md` §Completeness critic.
Evidence read: `status/probes/wf4/*`, `status/probes/cutover.md`, `status/probes/cutover-sethome.md`.
Cutover: gateway first started **18:43:08Z**, `/sethome`-equivalent restart **18:53:52Z** (PID 61969),
SOUL/rtk fix restart **19:40:45Z** (PID 76255). Every probe artifact is timestamped inside that
window — no WF3 pre-run was reused as WF4 evidence.

**Bottom line: Tars is live and verified, with one open security probe.** 14 of 15 probes PASS on
evidence that survives an adversarial read — permalinks with matching ts, gateway INFO log lines
tying Slack ts → session id → tool call → outbound send, and independent cross-checks rather than
assertions. The one exception is **probe 3, the negative test**, deferred by Gaetan's decision for
lack of a non-Gaetan sender: its configuration half is fully pinned and clean on all four D2
controls, but no message from a foreign identity has ever hit the gate. **The residual risk is
therefore precisely one thing: the allow-list has never been shown rejecting anyone.** Everything
else — transport, gating on the mention axis, all nine credential rows, model, plugins, A2A, mesh,
home-channel delivery, and Tars' own model→tool→ssh chain — is exercised. Ship, and treat probe
16 (probe 3's re-run) as a build-blocking follow-up, not a nice-to-have.

---

## Per-probe verdicts

| # | Probe | Verdict | Evidence pointer | Fail owner |
|---|---|---|---|---|
| 1 | Slack DM round-trip | **PASS** | `p01` — `ping wf4-1` 19:01:31Z → `pong wf4-1` 19:01:37Z (6.04 s), both permalinks, `gateway.log` inbound+outbound, session store binds Slack ts → `20260807_190132_acf22e5f` | — |
| 2 | Mention-gating | **PASS** | `p02` — mentioned msg dispatched (`inbound message: … chat=C08RWSTU9LK`); un-mentioned control produced **no inbound line at all** over 3m32s while the same PID ran probe 13's cron at 19:05:58 | — |
| 3 | Negative test (non-Gaetan ignored) | **DEFERRED-OPEN** | `p03` — no send made; `claude_ai_Slack` connector proven to authenticate *as* `U08BDJAMSRZ`, so it would have been a positive test mislabelled. Config half pinned in full | WF3, build-blocking **if** it fails |
| 4 | Gmail / Himalaya | **PASS** | `p04-08` §4 — 1 envelope, exit 0, himalaya v2.0.0 (`-m` not `-p`) | — |
| 5 | Calendar CalDAV | **PASS** | `p04-08` §5 — HTTP **207** multistatus, never a bare 200 | — |
| 6 | Linear | **PASS** | `p04-08` §6 — 200, `viewer.id` + `viewer.name` = Gaetan | — |
| 7 | Notion | **PASS (partial)** | `p04-08` §7 — `loadUserContent` 200 + `recordMap` (`search` XHR 400s unconditionally). Exercises `NOTION_TOKEN_V2` only | — (gap → probe 20) |
| 8 | GitHub + metarepo | **PASS (map divergence)** | `p04-08` §8 — `private: true`, `git fetch` clean, HEAD == `origin/main` at `~/dev/mc-metarepo`. Auth is `gh` device-flow, **not** the `GITHUB_PAT` D5 lists | — (gap → probe 21) |
| 9 | Slack-personal MCP | **PASS** | `p09-10` §9 — `mcp__slack__conversations_history` completed 0.49 s in the session id the turn printed; text/ts/user match a direct API read **to the microsecond**; `--no-cache` in config *and* runtime log, zero `users.list`; `auth.test ok` after | — |
| 10 | Model = gpt-5.6-sol | **PASS** | `p09-10` §10 — `auth status` logged in · `model.default`/`base_url` correct · live round-trip returned exactly `sol`. All three legs | — |
| 11 | Plugins (rtk/hindsight/lcm/adhd) | **PASS (repaired)** | rtk FAILed on first pass, **fixed and re-measured**: `fix-soul-rtk.md` §3/§6 — `rtk gain --history` shows `rtk ls -la /etc/hostname` −62 % at **19:43, post-restart**, hook firing on a real gateway-mediated command. hindsight not-available-by-design, lcm + adhd enabled | (was lane B / B5 — closed) |
| 12 | A2A agent card :9900 | **PASS** | `p11-12` §12 — 200 JSON (`hermes-tars`, capabilities, v1.0.0) on 127.0.0.1; external bind refused (correct per D2/R1) | — |
| 13 | Home-channel delivery | **PASS** | `p13` — CLI-created cron `--deliver slack` (bare platform ⇒ `SLACK_HOME_CHANNEL`); invoking process exited 19:04:22, delivery 19:06:01 into `D0BBYNM01BL`; `delivered to slack:D0BBYNM01BL via live adapter` + permalink | — |
| 14 | SSH mesh + tailscale | **PASS** | `p14` — 4 legs exit 0 (cooper→VM, VM→cooper, VM→mac, VM→phermes direct); `tailscale ping tars` → pong 1 ms from cooper. 192.168.0.3 never touched | — |
| 15 | Delegation-lite to cooper | **PASS** | `p15` — reply carried cooper's `hostname`/`uptime -p`; `agent.log` for `20260807_194604_3befdf` shows `tools.terminal_tool` create → `tool terminal completed (0.89s)` → `tool_turns=1`; matches a direct run on cooper | — |

**Tally: 14 PASS · 1 DEFERRED-OPEN · 0 FAIL · 8 GAPs (cross-check level, stubs §16–23).**

### Notes that survived scrutiny

- Probe 2's mentioned reply *body* was the empty-response error notice, not real content. The gating
  axis still passes (gate → dispatch → agent turn → outbound Slack write all worked); the content
  failure was the incident below, and the same channel, same sender prefix, was re-run green after
  the fix.
- Probe 13's PASS rests on the correct instrument, not the spec's. `/bg` was still exercised live
  and works — it simply always answers the invoking surface, so it cannot prove home routing.
- Probe 2/3's guardrail pins were taken by **replaying the gateway's own exporter**
  (`_apply_yaml_config`) against the live `config.yaml`, not by grepping a parsed config object.
  That respects the spec's trap and is stronger than the `/proc/environ` route the spec asked for
  (which is structurally incapable of showing `SLACK_*`).

## Cross-checks

**1 · D5's credential map — 9 of 10 rows exercised post-cutover.**

| Row | Exercised | By |
|---|---|---|
| A1 GitHub PAT | functionally yes, **name divergent** | probe 8 via `gh` device-flow; `GITHUB_PAT` exists in neither `.env` nor `secrets/tars.sops.yaml` → **probe 21** |
| A1b ChatGPT OAuth | yes | probe 10 (all three legs) |
| A2 Linear | yes | probe 6 |
| A2 Notion (3 fields) | **partial** | `NOTION_TOKEN_V2` yes (probe 7); `NOTION_SPACE_ID` + `NOTION_API_TOKEN` **never read**; `NOTION_FILE_TOKEN` unexercised by design → **probe 20** |
| A3 Gmail app password | yes | probe 4 |
| A3 Calendar (rides A3) | yes | probe 5 (207) |
| A4 Slack personal xoxc/xoxd | yes | probe 9 (+ `auth.test` after the MCP call) |
| A5 Slack app xoxb/xapp | yes | probes 1/2/13 + Socket Mode WSS session (`cutover.md` §3b) |
| LB Tailscale authkey | yes | probe 14e — `tailscale ping tars` → pong via `192.168.0.9:41641`, the spec's named gap, now closed |
| LB VM SSH keypair | yes | probe 14 legs 1–4 |

**2 · D2's guardrail surface — configuration pinned to a high standard, behaviour half-covered.**

| Control | Config pin | Behaviour |
|---|---|---|
| `require_mention: true` | ✅ `config.yaml:122`, exporter replay → `SLACK_REQUIRE_MENTION=true` | ✅ both directions (probe 2 A + B) |
| `strict_mention: true` | ✅ `config.yaml:123`, exporter replay → `true` | ❌ every tested mention sat at message start → **probe 22** |
| `unauthorized_dm_behavior: ignore` | ✅ `config.yaml:124`; `authz_mixin.py:785-880` precedence traced to a genuine drop (per-platform wins; allow-list fallback agrees) | ❌ → probe 3 / **16** |
| `SLACK_ALLOWED_USERS` | ✅ single value `U08BDJAMSRZ` in `.env`; runtime-loaded (startup warning `No env user allowlists configured` absent, count 0); no `SLACK_ALLOW_BOTS` / `SLACK_ALLOW_ALL_USERS` / `GATEWAY_ALLOW_ALL_USERS` anywhere | ❌ → probe 3 / **16** |

No stray top-level `platforms:` block (verified twice, independently). Config surface: **clean**.
Answering the spec's question directly — probe 3's evidence pins each control individually by
value, source line and consumer, not "a message got ignored for unknown reasons". What is missing
is the message, not the pin. Re-verified live at 19:53Z on PID 76255: guardrails unchanged,
`config.yaml`/`.env` mtimes still 18:08:43 / 18:53:46, zero auth or allow-list warnings since the
restart.

**3 · SOUL.md capability surface vs the 15 probes** — run against the **current** file
(`~/.hermes/SOUL.md`, md5 `c18c1ac4…`, 1472 B, byte-identical to `docs/specs/tars-profile.md` §1,
confirmed 19:52Z).

| SOUL claim | Probed? |
|---|---|
| Orchestrates, dispatches to other machines, reports back | ✅ probe 15 (model → shell tool → ssh → cooper), probe 13 (async delivery) |
| Rule 4 "answers Gaetan and no one else" — adapter layer | ⚠️ probe 3 **deferred** → 16 |
| Rule 4 identity mapping + `·` terminal — model layer | ❌ never exercised → **probe 19** |
| Rules 1–3 (no code, no PRs, orchestrate/report only) + rule 5 refusal | ❌ never exercised (`tars-profile.md` §1 excluded it deliberately; it is one cheap turn) → **probe 18** |
| Language: FR in → FR out, EN in → EN out | ❌ never exercised → **probe 17** |
| Tone (concise, verdicts not logs) | not probeable objectively — out of scope, not a gap |

**Pre- vs post-amendment.** SOUL rule 4 was amended at 19:39:40Z and the gateway restarted 19:40:45Z.
Probes 1, 2, 4–14 ran on PID 61969 (**pre**-amendment); probe 15 and the fix re-test ran on PID
76255 (**post**). This does **not** threaten the adapter-level claims — mention gating, allow-list
loading, guardrail pins, transport and routing all execute before the model is ever called, and
`config.yaml`/`.env` were not touched by the fix (mtimes prove it). Agreed and stated explicitly.
The model-level claims were all re-demonstrated post-amendment: channel mention → real content reply
(19:42), CLI round-trip (19:41), terminal-tool turn (19:43), probe 15 (19:46). Two model-level
results were **not** individually re-run post-restart — the DM round-trip and the cron home
delivery. Both re-read the same unchanged `.env`, so the risk is low; listed under residual, not
re-opened.

**4 · Cutover before-evidence — PARTIAL.** PLAN.md §Guardrails requires before-evidence prior to the
(still-deferred) profile delete. What exists is **inventory-level**: `docs/recon/r3-p-hermes.md`
§3–5 captured the profile dir listing, the unit file and `override.conf` verbatim, `.env` key names,
`config.yaml` guardrail values, `auth.json` top-level keys, the Slack manifest identifiers, plus
sizes and modes — enough to rebuild the wiring. `cutover.md` confirms the old profile was **not**
deleted and rollback is re-enabling the old unit. What does **not** exist is any byte-level capture
(no archive, no checksum manifest), so a delete today would be irreversible beyond what a human
transcribed. Not a blocker for declaring Tars live; it **is** a blocker for the delete → **probe 23**.

## Mid-WF4 incident and fix

**Symptom.** Every `@Tars` channel mention returned `⚠️ Empty response from model` after 3 retries —
61 s, 5 API calls, an alarming notice posted in-channel. DMs and CLI turns in the same minute
answered normally.

**Root cause** (`diag-empty-response.md`, ~80 % confidence, later confirmed by the fix). Not payload
size, tools, provider or model params — the failing channel turn was 24 099 prompt tokens vs the
succeeding DM's 24 029, same base prompt hash, same toolset, and a 26 700-token CLI turn succeeded
on the same backend in the same window. Channels are the only surface that injects a **multi-user
identity frame**: a sender prefix `[U08BDJAMSRZ | Slack user <@U08BDJAMSRZ>]` (`gateway/run.py`) plus
a "multiple users may participate" session block (`gateway/session.py`). `grep -c U08BDJAMSRZ` over
the stored system prompt was **0** — nothing told Tars that ID was Gaetan. Reading an unidentifiable
sender under SOUL rule 4 ("I ignore in silence"), the model spent 77 reasoning tokens and then
deliberately emitted nothing. Hermes has no silence terminal, so it read the decision as a provider
fault and replayed a byte-identical request four times (cache 100 %, `in=24184` each).

**Fix** (`fix-soul-rtk.md`): one prose edit to SOUL rule 4 — add the identity mapping
(`U08BDJAMSRZ` is Gaetan) and give the rule a terminal the runtime can represent (reply `·` instead
of silence). 2 lines removed, 4 added, nothing else in the file; mirrored byte-exact into
`docs/specs/tars-profile.md` §1 (uncommitted — orchestrator owns the commit). Gateway restarted
19:40:45Z, clean, `NRestarts=0` sustained.

**Re-test, same channel, same sender prefix, byte-for-byte the same frame:**

| | before (`14287ca4`) | after (`2dc9b6e6`) |
|---|---|---|
| API calls | 5 | **1** |
| Wall time | 61.0 s | **7.1 s** |
| `Empty response` warnings | 4 | **0** |
| Turn end | `empty_response_exhausted`, `(empty)` | `text_response(finish_reason=stop)` |
| Posted to channel | ⚠️ model-returned-no-content notice | **`Hi Gaetan 👋`** |

`grep -c "Empty response"` over `agent.log` since the restart = **0** (re-verified 19:53Z).
The rtk lane-B FAIL was repaired in the same pass and measured, not assumed.

## Open items

1. **Probe 3 — DEFERRED-OPEN (build-blocking if it fails).** Needs a sender that is genuinely not
   `U08BDJAMSRZ`; a teammate will send the consented message. Ready-to-run stub at
   `docs/specs/wf4-probes.md` §16, with PASS evidence upgraded from the spec's weak
   timestamped-absence to a **positive grep** for the WARNING the build actually emits
   (`[Slack] Early reject of unauthorized user`). Until it runs, "Tars ignores strangers" is a
   configuration claim, not an exercised one.
2. **GAPs → new probe stubs**, appended to `docs/specs/wf4-probes.md` — TRIAGED by Gaetan
   2026-08-07 evening:
   - **§16 probe-3 re-run: LATER** (teammate message pending) — **§18 (rules 1–3/5 refusal) and
     §19 (`·` terminal) run WITH it** in the same later pass.
   - **§21 (`GITHUB_PAT` row reconcile): done via cleanup agent** same evening.
   - **§17 FR/EN mirroring: might-do.**
   - **§20 Notion runtime read path: DROPPED.** §22 `strict_mention` behavioural: **covered by
     probe 2**, closed.
   - **§23 byte-level before-evidence: DROPPED** — the gated p-Hermes profile delete will rest on
     the inventory-level evidence in `r3-p-hermes.md` §3–5 only (Gaetan's call).
   - **Notion export path CANCELLED** (Gaetan): Notion is MCP live-fetching only; the `token_v2`
     trio (`NOTION_TOKEN_V2`/`NOTION_FILE_TOKEN`/`NOTION_SPACE_ID`) is removed from the VM `.env`
     and the SOPS store by the cleanup agent; re-harvest if ever needed. `NOTION_API_KEY`
     (bundled-skill credential) also dropped.
   - tirith degraded mode + Codex multi-step continuations: **LATER** (Gaetan; still worth
     noting in soak observations if they recur, but no active work).
3. **Spec corrections applied** (each marked *(corrected post-WF4, 2026-08-07)*): `/bg` → cron
   `--deliver slack` for §13 · himalaya v2 `-m` for §4 · `conversations.replies` not `.history` for
   §1/§3 · `/proc/environ` blindness for §2 · journald is WARNING-only, INFO lives in
   `~/.hermes/logs/` · Notion `loadUserContent` for §7 · `~/dev/mc-metarepo` + `gh` device-flow for
   §8 · direct `phermes` leg for §14 · rtk repair notes for §11 · §15 transcribed into the file.
4. **Residual, accepted:** the DM round-trip and cron home-delivery were proven on PID 61969 and not
   individually re-run after the 19:40:45Z restart (same unchanged `.env`; low risk). Channel
   behaviour with a *second real human* present is untested — probes 16/19/22 are exactly that
   surface. `NOTION_FILE_TOKEN` stays unexercised until a real export runs, as the spec allows.
   `~/.hermes/SOUL.md` is mode 664; it carries no secret, but it is the prompt that enforces the
   hard rules.
