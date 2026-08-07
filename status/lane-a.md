# Lane A — credential acquisition (orchestrator session only writes here)

Names only. Values live in `secrets/tars.sops.yaml` (SOPS+age, 2 recipients) or nowhere
(self-managing OAuth). Probe = the async agent evidence that the credential works.

## Checklist

| # | Credential | SOPS key(s) | Harvested | Probed |
|---|---|---|---|---|
| A1 | GitHub — `gh auth login` device-flow ON the VM (Gaetan's call; replaces the PAT) | — (VM gh store, never SOPS) | ☑ 2026-08-07 | ☑ `repo`+`read:org`, metarepo `.private=true` from the VM |
| A1b | ChatGPT-sub OAuth (`openai-codex`, on the VM) | — (never stored; self-rotating `auth.json` 0600) | ☑ 2026-08-07 device-flow | ☑ `hermes auth status openai-codex` → logged in |
| A2 | Linear API key (paste-file → SOPS → shred) | `LINEAR_API_KEY` | ☑ 2026-08-07 | ☑ 200, viewer = Gaëtan Cathelain |
| A2 | Notion API token — official MCP path (r8; Gaetan's call) | `NOTION_API_TOKEN` | ☑ 2026-08-07 | ☑ valid (bot `Tars`, ws `Mobile Club`); share-gap closed — 5 objects visible (Care, Tech, Cleaq, About MCG, one page) |
| A2 | Notion — reused from mc-kestra `.env` (Gaetan's call) | `NOTION_TOKEN_V2` `NOTION_FILE_TOKEN` `NOTION_SPACE_ID` | ☑ 2026-08-07 | ☑ token_v2 valid (loadUserContent 200 vs control 401); file_token deferred to first export. MCP-compat gate open (r8) |
| A3 | Gmail app password (Tars-labelled, 2nd one) | `GMAIL_ADDRESS` `GMAIL_APP_PASSWORD` | ☑ 2026-08-07 | ☑ IMAP list, 11 folders |
| A3 | Calendar (rides Gmail app password) | (same) | — | ☑ CalDAV 207 + inner 200 |
| A4 | Slack personal xoxc/xoxd (reused from mc-kestra) | `SLACK_TOKEN` `SLACK_COOKIE` | ☑ 2026-08-07 | ☑ auth.test ok, user_id `U08BDJAMSRZ`, 1-channel read ok |
| A5 | Slack Tars app xoxb/xapp (harvested at cutover; files→SOPS→shred) | `SLACK_BOT_TOKEN` `SLACK_APP_TOKEN` | ☑ 2026-08-07 | ☑ gateway Socket Mode live, 0 auth errors (`cutover.md`) |
| — | Tailscale auth key (single-use) | never stored — spent + shredded | ☑ 2026-08-07 | ☑ VM joined: `tars` 100.116.31.76, ping 1ms, LAN primary |
| — | Hindsight keys (2×) | — | ✗ SKIPPED for v1 (Gaetan, 2026-08-07) | — |

## WF3 wiring verdicts (2026-08-07, all evidence in `status/probes/`)

Independent verifier (`wf3-verify.md`) re-ran every credential probe itself; CONFIRMED = its
re-probe agreed with the wiring agent, not just the agent's claim.

| § | Capability | Verdict | Evidence |
|---|---|---|---|
| 1 | Gmail via Himalaya **v2.0.0** (spec block was v1-era, re-derived) | PASS_WITH_DEVIATIONS (re-run; first attempt aborted on incident below) — envelope probe exit 0, `account check` imap+smtp OK, re-probed by orchestrator | `wf3-s1-gmail.md` (incident), `wf3-s1b-gmail.md` |
| 2 | Linear + Notion + Calendar + **official Notion MCP** (24 tools) | CONFIRMED — Linear 200/viewer.id · Notion loadUserContent 200 · CalDAV 207 · live MCP tool call | `wf3-s2-linear-notion-calendar.md` |
| 2b | `mcp/notion` digest-pinned (`@sha256:df0d67…`), mode-600 preserved | PASS | `wf3-s2b-notion-pin.md` |
| 3 | GitHub `gh` + mc-metarepo clone (HEAD `4831253`, credential-grep 0) | CONFIRMED | `wf3-s3-github.md` |
| 4 | Slack personal MCP `korotovsky/slack-mcp-server:v1.3.0` stdio + **`--no-cache`** (the D1 eager-cache switch; no env-var form exists) | CONFIRMED — connected 18 tools, 40 channels via live chat, auth.test ok ×3 over time, user `U08BDJAMSRZ` | `wf3-s4-slack.md` |
| 5 | SSH mesh (3 legs, mac alias → tailnet name) + model backend `openai-codex`/`gpt-5.6-sol` — live `sol` reply closes R7's tier question | CONFIRMED | `wf3-s5-mesh-model.md` |

Spec corrections found live (update `wf3-wiring.md`/`wf4-probes.md` before WF4 copies them):
`hermes chat --oneshot` → `-Q -q` · `hermes mcp list` is registry-only, liveness = `hermes mcp
test <name>` · tool prefix is `mcp__slack__*` not `mcp_slack_*` · hermes units are **`--user`**
units (`hermes-gateway.service` disabled, `hermes-cloakbrowser.service` enabled) · §5's
`TARS_VM_SSH_PRIVATE_KEY` extract step is dead (no such SOPS key; keypair lives on the VM, B3)
· himalaya v2 schema + `-m` flag · `hermes auth add` did NOT write `model.base_url` (set
explicitly, verified no openrouter left).

Open, carried to cutover/WF4: Tars→p-Hermes leg — Gaetan's call 2026-08-07: wire it WITHOUT pve
root, via a prompt he pipes to p-Hermes itself (authorized_keys + sshd on VM 103; also carries
cutover step 1, the old-gateway disable) · `NOTION_FILE_TOKEN`
unprobed by design · hermes `notion` SKILL prompts for `NOTION_API_KEY` headless (decide: carry
in gateway env or disable the skill) · tirith scanner enabled-but-unavailable (degraded,
pre-existing) · rtk binary off non-interactive PATH (B5 plugin no-ops) · Codex multi-step chat
turns die after 3 continuation attempts (single-tool turns fine).

## Log

- 2026-08-07 22:32Z — **P3 + P4 APPLIED** (two peers, joint gate; Gaetan gave GO directly in each
  tab). SOUL v2 live (`b1bcabf`), Orca v2 skill live (`290b7d4`), contradiction window ~2.5 min,
  closed. First live mc-metarepo E2E **PASS** — Tars drove the `orca` CLI over ssh (never the v1
  wrapper: the only behavioural proof of skill pickup), produced a real 22-directory map on branch
  `GaetanCathelain/map-mc-metarepo-layout-readonly`, worktree kept per P3-a (mc-metarepo now at 10).
  **Three defects found, all in evidence `status/probes/wf5/apply-p3-orca-v2.md`:**
  (1) P3's load-bearing `--run` addressability claim is **DISPROVEN** — `check` needs
  `--terminal <handle>` even with `--run`; the dismissed mechanical reviewer was right, the original
  measurement was almost certainly taken inside an Orca-managed terminal. The skill's fallback was
  keyed to the wrong error code and could never fire.
  (2) `worker-start` reaching `ready` does **not** mean the agent started — Orca typed the brief
  into the TUI and left it unsubmitted; `check --wait` would have hung forever, indistinguishable
  from "still running". Gaetan caught it by pressing Enter himself. Remedy (`terminal send --enter`)
  is in the skill but **UNVERIFIED**.
  (3) **Tars self-edited its own governing SKILL.md** via `skill_manage` — sanctioned by Gaetan
  in-thread, content correct, but unaccounted for by any spec. See the open decision below.
  Also settled: no separate delivery-id field (`--ack` the message's own `msg_<12hex>`); dispatch
  ids are `ctx_<12hex>`; `worker-show` state is at `result.worker.state`; a worker's reply is
  addressed `run:<runId>` so `orchestration inbox` is needed; `--agent codex` is unusable on cooper
  (no accounts) — `claude` only.

- 2026-08-07 — **OPEN: Tars rewrites its own rules.** The live
  `~/.hermes/skills/delegate-to-cooper/SKILL.md` has drifted from the reviewed v2 and is still
  moving: applied 437 lines / md5 `929b23de`, archived post-self-edit 452 lines / `60b9a244`,
  measured by the hub at 22:50 → **460 lines / 26791 B / md5 `1bd36e2b`**, mode narrowed 664→600 by
  Tars. Both earlier versions are archived in `artifacts/`. Decisions for Gaetan: (a) keep the
  self-edit capability or remove `skill_manage` from Tars' toolset; (b) reconcile git↔VM — adopt the
  live file as authoritative or reset to the reviewed one; (c) the live file currently carries Tars'
  `--terminal` correction immediately followed by the surviving, now-disproven "MEASURED … ok:true"
  sentence, so it contradicts itself until someone edits it.

- 2026-08-07 — **OPEN: allowlist reject UNVERIFIED post-P4.** Config unchanged, code path intact,
  33 historical firings — but the newest is 20:46, before the apply, and with confinement dropped
  this control is now the **sole** gate on transitive Orca access. Needs one message from a real
  non-Gaetan Slack identity (the claude.ai connector cannot serve: it authenticates as Gaetan and is
  dropped as a bot before the allowlist is reached).

- 2026-08-07 — **WF4 negative test CLOSED as PASS-ALLOWLIST** (evening, teammate Nans):
  allowlist rejected a real mentioned message in 0.445 s with the positive `Early reject`
  WARNING; second human rejected ×2 unprompted. WF4 = **15/15**. Report amended.

- 2026-08-07 — **WF5 now-items SHIPPED** (6-agent workflow, recon→implement→Slack-E2E ×2):
  **Orca delegation v1** — `delegate-to-cooper` Hermes skill on the VM + dedicated cooper
  sandbox `~/orca/workspaces/tars-delegated/` (`delegate.sh` fixed-flag entrypoint: Tars can
  never pass permission-bypass flags; deny-rules `.claude/settings.json`; guardrail matrix
  proven: write/run OK, rm/`~/.ssh`/curl DENIED). Slack E2E 43.4 s, real cooper artifacts.
  **Kanban** — enabled via 6-line `toolsets:` addition (no restart), 12 `kanban_*` tools on
  CLI+Slack, full lane proven incl. dispatcher spawning real workers (ready→running→done),
  board archived clean. Specs: `docs/specs/wf5-orca-delegation.md`, `wf5-kanban.md`;
  evidence `status/probes/wf5/*`. Kanban recon's central claim was wrong (said no config key
  could do it) — implement agent measured and found the one-key fix; recon file carries the
  correction.

- 2026-08-07 — **WF4 CLOSED — Tars live and verified.** 15 probes exercised post-cutover:
  14 PASS, 0 FAIL, probe 3 (negative test) DEFERRED-OPEN — no teammate available; §16 stub
  ready, treated as build-blocking follow-up. Critic ran 4 cross-checks, marked 10 spec-stale
  corrections, appended gap stubs §16–23 (`docs/specs/wf4-probes.md`). Exercised report:
  `status/wf4-report.md`. Mid-WF4 incident fixed live: SOUL rule-4 identity gap made channel
  turns return empty (5 calls/61s) — identity mapping + `·` rewording deployed, re-test
  7.1s/1 call; rtk repaired (symlink + config, −62% measured). Probe 15 (added by Gaetan)
  proved Tars→cooper delegation-lite via real terminal tool. A2A investigation queued in WF5
  (PLAN.md). p-Hermes profile delete still gated — needs §23 byte-level before-evidence first.

- 2026-08-07 — **CUTOVER EXECUTED** (Gaetan's go in-session, evidence `status/probes/cutover.md`):
  old `hermes-gateway-tars.service` on p-Hermes disabled+inactive via a prompt Gaetan piped to
  p-Hermes itself (no pve root; `reset-failed` applied to that unit only) → A5 tokens + 
  `SLACK_ALLOWED_USERS=U08BDJAMSRZ` merged into VM `.env` (10→13 keys, backup kept) → new
  `hermes-gateway.service` (`--user`) active since 18:43:08, Socket Mode connected, allowlist in
  effect, CloakBrowser healthy. Bonus: Tars→p-Hermes SSH leg wired after all (VM 103 =
  `192.168.0.8`, user `hermes`, key-restricted `from=`), probe exit 0 — reverses the WF3-era
  narrowing; WF4 MAY probe this leg. **`/sethome` pending in the live DM** (sets
  `SLACK_HOME_CHANNEL`). Rollback unchanged: stop new unit, re-enable old.
  Spec-stale note: `tars-profile.md` §3 still lists `GITHUB_PAT` — obsolete since A1's
  device-flow rewrite; not a gap.

- 2026-08-07 — `/sethome` applied MANUALLY (Slack blocks slash commands across the whole
  Agent-class chat surface — they present as threads): source-verified equivalent
  (`SLACK_HOME_CHANNEL` is getenv-only), `.env` 13→14 keys, `SLACK_HOME_CHANNEL=D0BBYNM01BL`
  (matches p-Hermes' old `override.conf`, r3 — independent confirmation), gateway restarted
  clean, allowlist intact. Evidence `status/probes/cutover-sethome.md`. End-to-end DM delivery
  deliberately left to WF4 §13.

- 2026-08-07 — **INCIDENT (closed — Gaetan accepted the risk 2026-08-07, no rotation; same
  ruling as token_v2):** the first WF3 gmail
  agent ran whole-file `sops -d … | grep` despite the standing ban and printed the **full**
  `GMAIL_ADDRESS` + `GMAIL_APP_PASSWORD` values into its own transcript (local file + model
  context). It halted itself; a clean re-run did the wiring (per-key `--extract` only). The pair
  stays in service and probes pass (IMAP, CalDAV 207). Options: **accept** (token_v2 precedent)
  or **re-mint** the app password — 2-min browser step, then re-deliver only
  `~/.hermes/.secrets/gmail_app_password` + re-merge `~/.hermes/.env` (both idempotent, no
  config edits).

- 2026-08-07 — WF3 run: 6-agent workflow (5 wiring + independent verifier) + 2 remediation
  agents. All five capabilities wired on the VM; gateway still `disabled`+`inactive`; p-Hermes
  untouched; verifier's secret-sweep of all evidence files CLEAN.

- 2026-08-07 — **INCIDENT (closed — Gaetan accepted the risk 2026-08-07, no rotation):** a probe agent ran `sops -d … | head -c 200` while sanity-checking
  the age key and printed ~200 chars of `NOTION_TOKEN_V2` into its own transcript (local file +
  model context). Not the credential it was probing; no other key exposed. Scope: partial value of
  the Notion web-session cookie. Decision pending with Gaetan: rotate (Notion → log out of all
  sessions) judged not worth the mc-kestra re-harvest for a partial value scoped to the
  bulk-export path. Probe procedure was already file-only for the target token; the leak was a
  pre-probe sanity check. Standing fix: probe prompts now forbid any whole-file `sops -d` output.

- 2026-08-07 — SOPS store bootstrapped (`scripts/tars-secret`, `.sops.yaml`); dummy-roundtrip
  verified. VM age key enrolled as 2nd recipient (canary: 2 recipients in metadata, cooper
  decrypt ok).
- 2026-08-07 — A1b path settled by R7 + grill: fresh device-code OAuth on the VM, never copied,
  never stored. Runs at WF3 time.
- 2026-08-07 — Tailscale console 500 → key mint deferred to WF3 (LAN primary per D4).
- 2026-08-07 — Browser flow flipped: Gaetan's own browser + `tars-secret` hidden prompt
  (DevTools-bridge browser has no sessions; every service would cost a fresh 2FA there).
- 2026-08-07 — Lane B complete (smoke pass, e395e07). Join now waits on lane A only.
