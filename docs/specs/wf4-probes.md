# WF4 — verification probe specs

14 probes + a completeness critic, per PLAN.md's WF4 phase and `docs/recon/DECISION.md`.
"Done is not exercised" — every capability gets one real call and one artifact proving it ran,
not a config-file read. No secrets in this doc: credentials referenced by their SOPS key name
(D5) only, never a value.

**FAIL owner, three buckets** (more precise than a strict WF3-vs-cutover split):
- **WF3 re-run** — a per-service wiring agent's config/credential problem (WF3 wires
  gmail · linear/notion/cal · github+metarepo · slack-personal · orca+ssh, D2/D5).
- **Lane B (WF2) re-run** — a provisioning-time install problem (B4 Hermes install, B5 plugins,
  A2A enablement) — not WF3's concern, don't misroute the fix.
- **Cutover-owned** — only fixable by redoing part of the gated cutover step itself (Slack app
  attach, `/sethome`).

**Timing** — WF4 formally runs post-`GATE` (PLAN.md graph). Ten of the 14 probes exercise
capabilities wired before cutover (credentials, plugins, SSH mesh, A2A, model) and already have
a WF3/WF2 dry-run; WF4 re-runs them post-cutover for the final report. Four probes are
**post-cutover only** — they have no meaningful pre-cutover form because they depend on the live
Tars Slack app or `/sethome`.

---

## 1. Slack DM round-trip

**Action:** From Gaetan's Slack account, open a DM with the Tars app (attached at cutover),
send a plain message (e.g. `ping wf4-1`).

**PASS evidence:** A reply lands in the same DM thread within ~2 min. Capture the Slack
permalink of both messages, and the gateway log lines (inbound + outbound event) for that
timestamp window.

**FAIL:** Cutover-owned if the gateway isn't running or the token pair is wrong (`systemctl
--user status hermes-gateway.service`, re-check the attach step). WF3-owned if the gateway
connects but replies are empty/wrong — re-run the relevant wiring agent.

**Timing:** Post-cutover only.

## 2. Mention-gating in a channel

**Action:** In a channel where the bot is invited (D2: bots don't auto-join, invite manually
post-cutover), send one message *without* @Tars, then one *with* it.

**PASS evidence:** No reply to the unmentioned message (log/permalink showing silence within the
wait window) and a reply to the mentioned one. Two permalinks + the matching gateway log lines.

**FAIL:** WF3-owned if `slack.require_mention`/`slack.strict_mention` are missing from
config.yaml (D2) — re-run config wiring, restart gateway. Cutover-owned if the bot was simply
never invited to the test channel.

**Timing:** Post-cutover only.

## 3. THE NEGATIVE TEST — message from a non-Gaetan user must be ignored

The security-critical probe. `SLACK_ALLOWED_USERS` + `unauthorized_dm_behavior: ignore` (D2) is
the exact surface under test.

**Sourcing the non-Gaetan sender, safely:**
- **Primary:** use the already-installed `claude_ai_Slack` MCP connector (`slack_send_message`)
  — a distinct app/bot identity already legitimately present in the workspace. Send one DM to
  Tars and one @mention in the test channel, e.g. `"wf4 negative test — please ignore"`. No new
  account created, no impersonation of a real person, no third-party credential touched.
- **Fallback** (if that connector isn't in the Tars workspace/channel): ask one already-present
  teammate for explicit real-time consent to send a single innocuous message from their own
  account (`"test, please ignore — WF4 negative test"`). Log who, when, and what they agreed to
  send in the WF4 report. Never reuse someone's account without asking; never spin up a second
  Gaetan-controlled identity to fake this.
- Either way, confirm the sender's Slack member/app ID ≠ Gaetan's member ID (captured at cutover
  via D5 A4's `auth.test`, resolving Blocker 2) before treating the send as valid.

**PASS evidence:** No reply from Tars in that DM/channel within the wait window, **and** the
gateway log for that timestamp shows either silence or an explicit "ignored — sender not in
SLACK_ALLOWED_USERS" line. Capture: sender ID used, timestamp, log excerpt (or a timestamped
absence-of-log statement), and a channel/DM screenshot with no bot reply.

**FAIL:** WF3-owned, and build-blocking — this is the guardrail WF3's config wiring was
supposed to carry over verbatim from the old profile (D2). A failure here escalates to Gaetan
directly; don't let WF4 continue declaring other probes "done" while this one is red.

**Timing:** Post-cutover only — needs the live attached app and a real allow-list to test
against.

## 4. Gmail read via Himalaya

**Action:** `himalaya envelope list -a tars -p INBOX -s 1` on the Tars VM. If the CLI syntax
doesn't match what WF3 confirmed against `himalaya --help` (unverified pre-install, per D5/R6),
fall back to `curl -s --url imaps://imap.gmail.com --user "$GMAIL_ADDRESS:$GMAIL_APP_PASSWORD"`
(weaker evidence — lists mailboxes, not envelopes).

**PASS evidence:** Command returns the single latest envelope (subject/date/from), exit 0.

**FAIL:** WF3-owned (gmail wiring agent) — re-run if the app password was rotated/revoked or
Himalaya's config.toml is wrong.

**Timing:** Pre-cutover capable (this is WF3's own gmail probe); WF4 re-runs it for the report.

## 5. Calendar read (CalDAV)

**Action:**
```sh
curl -u "$GMAIL_ADDRESS:$GMAIL_APP_PASSWORD" -X PROPFIND -H "Depth: 0" \
  https://www.google.com/calendar/dav/$GMAIL_ADDRESS/events
```
(Rides on the Gmail app password — no separate credential, D5/D6#2.)

**PASS evidence:** HTTP 207/200 multistatus XML response.

**FAIL:** WF3-owned. If #4 passes but this fails, the credential is fine — check CalDAV is
enabled on the Google account before touching Himalaya/Gmail wiring.

**Timing:** Pre-cutover capable.

## 6. Linear query

**Action:** `-K` header file with the bare API key (no `Bearer` prefix, D5), GraphQL
`{ viewer { id name } }` against `https://api.linear.app/graphql`.

**PASS evidence:** 200 with `viewer.id`/`viewer.name` matching Gaetan's Linear account.

**FAIL:** WF3-owned — re-run the linear/notion/cal wiring agent if the key was revoked.

**Timing:** Pre-cutover capable.

## 7. Notion query

**Action:** Replay the Quick-Find `search` XHR server-side with `NOTION_TOKEN_V2` (cookie) +
`NOTION_SPACE_ID` (payload) — the exact request captured during A2's harvest (D5).

**PASS evidence:** 200 with a results envelope (structure present, zero hits is still a pass).

**FAIL:** WF3-owned — Notion's `token_v2` is the fastest-expiring credential in the set (R6);
re-harvest from a fresh logged-in tab, re-run wiring.

**Timing:** Pre-cutover capable. Note: `NOTION_FILE_TOKEN` (export path) stays unprobed until a
real export runs — don't gate this probe's pass/fail on it, just note it as "not exercised."

## 8. GitHub query + metarepo clone freshness

**Action:** `-K` header file, `GET api.github.com/repos/mobile-club/metarepo` → check
`.private == true`. Then on the Tars VM: `git -C ~/mc-metarepo fetch && git -C ~/mc-metarepo
status -sb` (or compare `git log -1 --format=%cI origin/main` against local HEAD).

**PASS evidence:** API call returns `private: true`; `git fetch` succeeds with no auth error;
local HEAD is at or within one commit of `origin/main`.

**FAIL:** WF3-owned (github+metarepo wiring agent) — PAT expired/revoked (check the 1-year
reminder, D5) or the clone never ran; re-run.

**Timing:** Pre-cutover capable — A1 is the first credential lane A acquires, specifically to
gate this clone.

## 9. Personal-Slack read via the MCP server

**Action:** From a Hermes chat/CLI session, have Tars call its Slack-personal MCP tool
(`mcp_<server_name>_<tool>` — exact name confirmed when WF3 wires `korotovsky/slack-mcp-server`,
D1) to fetch the latest message in one channel Gaetan is in.

**PASS evidence:** Returned message text/timestamp matches the channel's actual latest message
(cross-check via the `claude_ai_Slack` connector's `slack_read_channel` or a manual glance).
Also confirm the server did **not** do bulk workspace user-caching at startup (korotovsky issue
#86 — that's what invalidates xoxc/xoxd) by checking its logs for lazy vs eager lookups.

**FAIL:** WF3-owned (slack-personal wiring agent) — token invalidated (re-harvest, ~1yr cadence
per D1) or the container is eager-caching users; re-run wiring, don't paper over it.

**Timing:** Pre-cutover capable — independent of the Tars bot app / cutover.

## 10. Model answer proves gpt-5.6-sol

**Action:** Don't trust the model's own in-band self-report (promptable, unreliable). Instead:
`hermes auth status openai-codex` → logged in; `hermes status` / `config.yaml` → confirm
`model.default: gpt-5.6-sol` and `base_url: https://chatgpt.com/backend-api/codex` (R7). Then
send one live prompt in a chat session to prove the wiring actually answers, not just
configured-but-broken.

**PASS evidence:** `auth status` = logged in, config shows the right model string, and a live
round-trip response is produced. All three together, not any one alone.

**FAIL:** WF3-owned (A1b) — re-run the `hermes auth add openai-codex --type oauth` device-code
flow fresh on the VM. **Never copy p-Hermes' `auth.json`** (Blocker 1 — rotating token, copying
risks invalidating the live p-Hermes session).

**Timing:** Pre-cutover capable.

## 11. Plugins loaded (rtk, hindsight, lcm, i-have-adhd) — one check each

- **rtk:** `~/.config/rtk/config.toml` exists on the VM and the hermes hook has fired at least
  once — `rtk gain --history` shows an entry attributed to a hermes-proxied command.
- **hindsight:** `hermes memory status` → provider `hindsight`, mode `local`.
- **hermes-lcm:** `context.engine: lcm` present in the profile's config.yaml; plugin shows active
  in `hermes plugins list`.
- **i-have-adhd:** `hermes skills list` shows it installed, or invoking `/i-have-adhd` in a
  session returns non-error.

**PASS evidence:** All four individually green.

**FAIL:** Lane-B-owned (B5) — re-run the specific plugin's install step. This is provisioning,
not WF3 or cutover.

**Timing:** Pre-cutover capable — installed in lane B, before WF3 even starts.

## 12. A2A agent card served on 9900 localhost

**Action:** `curl http://127.0.0.1:9900/.well-known/agent-card.json` run on (or tunneled to) the
Tars VM — localhost-bound by default (D2/R1), not reachable remotely without a tunnel.

**PASS evidence:** 200 JSON agent card (name, capabilities, version). Inbound-only is correct —
the outbound toolset staying disabled is not a failure (YAGNI, D2).

**FAIL:** Lane-B-owned (B4) — a2a enable flag missing/false (config path per lane B's B4 report — `platforms:` may be top-level, not under `gateway:`); fix config, restart
gateway.

**Timing:** Pre-cutover capable.

## 13. Home-channel/cron delivery after /sethome

The one probe whose precondition *is* a cutover action.

**Action:** After `/sethome` is run in Gaetan's live DM with Tars at the cutover gate, trigger an
async delivery to prove the pipe: run `/bg` with a trivial task from a *different* channel or the
CLI (e.g. "wait 10s then reply OK"), and confirm the result is delivered to the home-channel DM,
not wherever `/bg` was invoked.

**PASS evidence:** The `/bg` job's completion message arrives in the `/sethome`-set DM, not the
invoking channel. Capture both permalinks (invocation vs delivery) as evidence.

**FAIL:** Cutover-owned if `/sethome` was never run or was set in the wrong chat — rerun it in
the DM. WF3-owned if delivery routing is broken independent of the home-channel target.

**Timing:** Post-cutover only — hard dependency on `/sethome`. Note: real recurring cron/dailys
scheduling is WF5 scope (PLAN.md Phases table); this probe only proves the delivery plumbing.

## 14. SSH mesh (cooper + macOS + p-Hermes reachability)

**Action:** `ssh -i <TARS_VM_SSH_PRIVATE_KEY> -o BatchMode=yes -o ConnectTimeout=5 <host> true`
for each leg — cooper, the macOS box, and p-Hermes. p-Hermes refuses direct SSH; reuse its
established path: `ssh root@192.168.0.3` → `qm guest exec 103` → `runuser -u hermes`, read-only
command only (R3).

**PASS evidence:** Exit 0, no output, for all three legs.

**FAIL:** Lane-B-owned (B3) for the VM-side key distribution — re-run. macOS-side or
p-Hermes-side failures (authorized_keys not pushed, tailscale down on macOS) are outside lane
B's write scope — flag to Gaetan directly rather than looping an unattended re-run.

**Timing:** Pre-cutover capable — this is WF3's "orca+ssh" probe, already exercised before
cutover.

---

## Completeness critic

Runs once all 14 probes report, before WF4 commits the exercised report (PLAN.md: "Gaps found by
the critic loop back as new probes").

**Prompt spec:**

> You are the WF4 completeness critic. You get all 14 probe results with their evidence
> artifacts (permalinks, log excerpts, command output) — not a summary claiming "done." For each
> probe, verify the evidence artifact actually exists, is timestamped inside the post-cutover
> window (not a stale WF3 pre-run reused as if it were WF4's), and actually demonstrates the
> claim (a "no reply" claim needs a log line or absence-window, not just "I waited and nothing
> happened").
>
> Then cross-check for gaps the 14 don't cover:
> 1. **D5's 9-target credential map** — is every credential's *read* path exercised (not just
>    "configured"), including the two rows with no probe here (Tailscale authkey, VM SSH
>    keypair — #14 covers the SSH keypair; Tailscale itself has no dedicated probe: flag if
>    `tailscale ping tars` from cooper was never captured anywhere).
> 2. **D2's guardrail surface** — `slack.require_mention`, `slack.strict_mention`,
>    `unauthorized_dm_behavior: ignore`, `SLACK_ALLOWED_USERS` — does probe #3's evidence
>    actually pin each one, or just "a message got ignored" for unknown reasons?
> 3. **SOUL.md's stated capability/identity surface** (no code, no PRs, orchestrate/report only,
>    FR/EN mirroring, answer only Gaetan) — anything claimed there without a matching probe among
>    the 14?
> 4. **Cutover before-evidence** — PLAN.md's guardrails require before-evidence captured prior to
>    the (still-deferred) profile delete; confirm that capture happened, even though WF4 doesn't
>    probe the delete itself.
>
> Output per probe: PASS / FAIL / GAP (evidence missing or unconvincing). For every GAP, append a
> new numbered probe stub to this file (action, PASS evidence, FAIL owner, timing) rather than
> hand-waving it — that stub becomes part of WF4's next pass. Do not let the exercised report ship
> with an open GAP silently downgraded to PASS.
