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

- 2026-08-12 ~11:10Z — **WF6 closed (GCN-7 Done) + engagement-checker v1.8.0
  live: the 2026-08-11 batch is fully landed.** R2 measured in-window on the
  UNMODIFIED semantics per Gaetan's split-diff ruling: check 2 **PASS on all
  three legs** — 2a EC-EAF7BB→GCN-34 (natural 08:00Z fire), 2b no re-report
  across 7 fires, 2c GCN-12 close reconciled `linear:completed` (board restored
  to Todo after; its checker reminder loop is consumed — noted in MIGHT-DO).
  Stage-1 plumbing (v1.7.2, `d5a051d`) had its first live exercise: §4
  collector ran 7/7 from cron (8+ blocks the day before), Linear cursor
  advanced 14:30:26Z(-1d)→11:00:37Z; email collection live post-interpreter
  fix (GCN-34 came from email). GCN-25 was NEVER wrong-written (predicted
  close-back path did not trigger; item's queue entry already terminal).
  Stage-2 semantics applied 10:44Z (v1.8.0, `02d9c96`; live == origin/main ==
  staged, md5 `d0adb891…`; gateway `ActiveEnterTimestamp` unchanged through
  everything, still Mon 2026-08-10 20:00:38 UTC). First natural 1.8.0 run
  11:00Z clean: P1–P5 + F2 regression all PASS, close-back guard exercised in
  production (GCN-38). Closures: GCN-7, GCN-30, GCN-31, GCN-32 all Done with
  evidence comments; GCN-33/35/37/38/39/40 are real checker-filed business
  items (untouched). Residual on record: R1 inbound-DM full loop (protocol
  `gcn7-wf6-e2e.md` §9.1, MIGHT-DO). Evidence: `status/probes/gcn7-r2-run.md`
  (`105c000`), `gcn30-32-stage2-applied.md` (`aed3987`),
  `gcn30-32-checker-fix.md`, `gcn31-collector.md`. Overnight, independently:
  Tars self-merged a NEW skill via the fixed rule-2 recipe (PR #58,
  slack-channel-context) — second in-the-wild proof of the GCN-13 fix.

- 2026-08-11 ~18:25Z — **WF6 T7/GCN-7 §Verification pass — first measurement
  AFTER the GCN-10 prune; supersedes earlier WF6 evidence.** Every 2026-08-11
  cron fire (15 checker + 1 brief) predates the prune instant 16:59:45Z, so all
  prior WF6 PASSes measured the old 58-tool surface. Established this lane:
  agent.log timestamps are **UTC** (simultaneous `date -u`/`date` + 3 cron
  cross-checks). Counts now notion 24 / slack 20 / linear 10 = 54 from 3
  servers; `ActiveEnterTimestamp` unchanged at `Mon 2026-08-10 20:00:38 UTC`
  across the whole lane, measured by 4 agents — the 58→10 change went live by
  reload, no restart. **ZERO `mcp__linear__*` denials anywhere: no GCN-10
  regression.** Verdicts (`status/probes/gcn7-wf6-e2e.md`): check 1 **PARTIAL**
  — agent+tool leg PASS post-prune (GCN-28: team GCN, label `test-check`, P3,
  reply carried key+URL, `save_issue completed` 17:24:34.451Z), inbound Slack
  leg **never stimulated** (Gaetan held off after seeing GCN-28; `gateway.log`
  inbound stops 16:51:28Z, no `platform=slack` session after 17:00Z — 3
  independent negatives). check 2 **BLOCKED, NOT MEASURED** — 2a/2b/2c
  unexercised: `cron run` at 19:30/19:43 Paris hit the skill's own
  10:00–17:00 gate (`[SILENT]` in ~55 s, queue byte-identical), and manual
  `hermes chat` cleared that gate but died on a second, undocumented one (§4
  collector is a terminal command; no TTY → `BLOCKED … timed out` after 302.81 s;
  does NOT fire under cron). No SKILL.md edit, no issue closed — fully
  re-runnable. check 3 **PASS** (3a/3b/3c) — brief delivered ts
  1786470753.300289, 12/12 board rows real vs live Linear, no live P1 omitted,
  held through a mid-run compaction. check 4 **PASS** on a clean rerun (GCN-29,
  no team/id/framing in the prompt); the first run (GCN-27) is recorded as
  **contaminated by this lane's own prompt and not scored**. check 5 **PASS** —
  both repos clean and == origin/main (`26e4b91` / `3065f7d`), 8/8 skill
  mirrors byte-identical. Attribution answered: the engagement-checker did NOT
  file GCN-28 (its only save_issue was 15:04:13Z → GCN-25). Side findings
  filed: **GCN-30** (§5 misses `statusType=duplicate` → real nag-loop hole,
  live instance GCN-25), **GCN-31** (collector-path coverage risk) and
  **GCN-32** (engagement-checker records carry invented provenance: a cited
  "09:00–19:00 window" that exists in no file, and `linear_issue: GCN-26`
  written into durable state on a turn with zero Linear calls). Probe artifacts
  GCN-27/28/29 all **Canceled** 18:32:29–18:32:31Z, comment first on each. Two
  residuals with executable protocols in the probe file: Gaetan's literal DM,
  and an in-window (`08:00–15:00Z`) `hermes cron run 759e08c598e3` for check 2.

- 2026-08-11 ~07:50Z — **WF6 first production fire GREEN — goal's last
  observation point closed.** Brief cron fired 06:30:55Z (schedule is
  Paris-local), ran 3m43s, delivered to C0BP2GZUFSR. Board block row-verified
  against a fresh `linear_board.py` run: all deltas are genuine board movement
  (MC-4179/NMC-622 closed ~50 min post-delivery), all other rows verbatim;
  `Coverage: 2 views, both complete` present; zero terminal-state leakage.
  One anomaly, likely pre-existing: delivery dropped the origin thread_id and
  posted top-level instead of into the anchor thread (a "report-thread
  reconciler" cron already exists for this pattern — not a WF6 regression;
  delivery target config verified unchanged). Also this morning: Tars
  self-evolved linear-ticketing (PR #44/#45) reproducing the empty-file merge
  defect — recurrence #2 commented onto GCN-13.
- 2026-08-11 ~00:15Z — **WF6 check 1 PASS by proxy** (`check1-proxy.md`,
  main `a71985e`): bare ask via `hermes chat` → GCN-14 created with one label
  (`test-check`), explicit P4, direct-to-Todo (stateHistory clean), assignee
  Gaetan, key in the reply; canceled on instruction, verified. Same agent
  loop/skill/tools as a DM; the Slack transport itself is proven by live
  daily use (Gaetan's deploy-check exchange ts 1786372813/1786372819).
  Gaetan's literal DM + the first scheduled cron fire remain as gold
  acceptance / natural-use confirmation.
- 2026-08-10 ~23:30Z — **WF6 build complete; goal open on two observation
  points.** All three skills live on the VM, sha-matched to main (evidence:
  `status/probes/wf6/`, 15 files). Spec §Verification: check 2 PASS (GCN-12
  full lifecycle incl. the reopen edge — closed_at latch deleted, no duplicate
  minted); check 3 PASS twice under forced context compaction (deterministic
  script render + Coverage gate; the first agent-rendered attempt fabricated
  7/12 titles and was rolled back same evening before any cron fire); check 4
  PASS as a composite (Session B: preference line `3065f7d` + 3-session
  default-resolution tests + real GCN-8 write via the account connector);
  check 5 PASS (`ActiveEnterTimestamp` constant — NRestarts now documented as
  invalid restart proof). OPEN: check 1 (Gaetan's live DM — only he can
  trigger Tars) and the first scheduled fire of the realigned cron prompts
  (weekday 08:30+02:00 brief, 10:00–16:00 engagement runs). Deferred by
  design: GCN-10 (prune the 58-tool mcp__linear__* surface — `delete_*`
  exposed, policy-only bound). Incidents this run: coordinator mis-derived an
  abort deadline (worker corrected + rolled back unilaterally — right call);
  coordinator inverted a Slack channel id (worker's name-lookup refusal
  caught it); Ahmad Moussa impersonation attempt at Tars rejected by the
  sender guard pre-evaluation (escalated to Gaetan); Tars self-merged PRs
  #40–43 mid-run, #42 landing an EMPTY skill file restored 22 s later by #43
  — guard ticket filed in GCN.
- 2026-08-10 ~16:50Z — **WF6 underway: Tars ⇄ Linear GCN integration** (spec
  `docs/specs/wf6-linear-integration.md`, /goal-driven; coordinator = the
  kanban-research session, megaultracode-orca gang). Landed so far: r9/r9a
  recon + decision docs; team **Gaetan/GCN** (Gaetan-created, adopted) seeded
  with 6 workload labels + GCN-1..7 (T1/T2 Done at filing); cooper default
  team=GCN via metarepo `preferences/tooling.md` (`3065f7d`, 3-session
  headless verification, GCN-6 Done); Hermes-NATIVE transport mandated by
  Gaetan mid-flight → spec amended `639fa0b`, r9a rollback story corrected
  (**no Linear MCP ever existed on the VM** — .bak was pre-edit), GCN-9 filed
  as precursor (wire `linear` catalog preset). Session A (`wf6-vm-skills`)
  executing GCN-9→5→3→4→7; rulings on channel: delta collector stays raw
  (coverage-gated cursor doctrine), all other reads/writes native
  `mcp__linear__*`; GCN-only write guard is instruction-level BY GOAL DESIGN
  (cross-team writes on explicit per-message instruction, SOUL-rule-4
  semantics). Session B done + torn down. Evidence: `status/probes/wf6/`.
- 2026-08-10 ~11:50Z — **P6 applied: mc-metarepo knowledge base live** (Gaetan's go in chat).
  SOUL.md rule 9 added (KB pointer; `.bak`+flock, backup `~/.hermes/SOUL.md.bak-p6-20260810`),
  `~/.hermes/scripts/kb-refresh.sh` installed, hermes cron `mc-metarepo-refresh` created
  (id `3eb53a322510`, every 60m, no-agent, deliver slack); clone refreshed `4831253`→`2d3d014`.
  First apply run was interrupted mid-flight on Gaetan's hold (live Tars work running) — stop
  landed between the SOUL edit and the cron create; resumed and finished on his go. The 4
  pre-existing cron jobs intact. Verify (`status/probes/wf5/p6-verify.md`): **PARTIAL** —
  lookup-style trial grounded+correct (`search_files`+`read_file` fired, answer matched
  `knowledge/nm-vs-mc-order-source.md` incl. the business_entity_id trap); judgment-style trial
  (sudo-argv key rotation) reached the right conclusion WITHOUT touching the KB (zero retrieval
  calls logged) — retrieval *triggering* on scenario-shaped questions is the residual weak
  point, wiring itself works. Reduced verification per Gaetan's right-sizing: 2 of the
  proposal's 5 §Verify questions re-fired, no Slack-surface trial, forced-failure refresh tick
  (Slack alert path) NOT tested. P6-c map stays deferred; P6-d/e/f push-back design-only.
  P6 considered DONE per Gaetan.

- 2026-08-08 14:45Z — **Tool-call display split: DMs verbose, channels quiet** (Gaetan's ask after
  Tars dumped its tool trace into C04LZBBNVNY). Hermes has per-platform but not per-chat-type
  display resolution, so: (1) local patch in the VM's `~/.hermes/hermes-agent` checkout
  (commit `d615ca8`, gateway/run.py) adding a DM-scoped override key
  `display.platforms.<platform>.tool_progress_dm`; (2) config.yaml (`.bak` + flock) now sets
  `display.platforms.slack: {tool_progress: off, tool_progress_dm: all}`; global
  `display.tool_progress: all` untouched. Gateway restarted 14:28:04Z — this interrupted the live
  staging-fix run in C04LZBBNVNY; auto-resume recovered both sessions (~1 min lost, Tars cleaned its
  own orphan lock). **Verified on the resumed sessions**: DM "Continue" thread still shows tool
  bubbles post-restart; the channel turn at 14:32:16Z ran tools (memory-tool error for session
  `142314_3c38a71d` in errors.log) yet posted text only. Patch is local-only — re-apply/upstream it
  before any `git pull` of hermes-agent.
  Note: three probes this session (plus Gaetan's 13:38 CEST DM) were sent via the claude.ai
  connector and were inert — connector messages cannot trigger Tars, as already established in
  `status/probes/wf5/connector-message-invisibility.md`; verification used the resumed live
  sessions instead. Three orphan probe messages left in the Tars DM.

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

- 2026-08-07 23:10Z — **Self-edit question RESOLVED by Gaetan: keep the capability, require the
  commit.** *"The model should still hold the skill manage possibility, it's just that it should
  commit it too."* Tars' self-edits were correct; the defect was that they lived nowhere reviewable.
  Implemented (`ed7b33c`, `6c078a8`): SOUL rule 2 ("I never merge, approve or push") was
  **unimplementable** as written — committing a skill edit *is* a push, so rule 5 would have made
  Tars refuse in one line. Rule 2 now carries exactly one exception, scoped to Tars' own operating
  record, with the verbatim command sequence and an explicit boundary: *"This exception covers my own
  skills and nothing else. It is not licence to push in a repo I was delegated work in."* Put in SOUL
  rather than a skill on purpose — `skill_manage` can fire in any turn, and SOUL is the only text
  guaranteed to be in context. Verified by **running** the flow as Tars (VM → cooper → commit → push
  → PUSHED), which is what seeded `skills/delegate-to-cooper/SKILL.md`.
  **Repo layout separated** (two paths were doing one job — the drift's root cause):
  `SOUL.md` + `skills/<name>/SKILL.md` = **live mirrors**, equal to the VM, Tars writes here;
  `artifacts/*-SKILL*.md` = **frozen snapshots** (restored to `929b23de` = v2 exactly as P3
  approved, so the proposal and mandates are not retroactively falsified). Hub-verified parity:
  SOUL `13655395…` and SKILL `a6e0a4b5…` identical on VM and in git.
  **Residual: this is a RULE, not an ENFORCEMENT** — nothing stops Tars editing a skill and not
  committing it; the guarantee is model compliance, same as rules 1–8. A hook on `skill_manage`
  would make it structural. Not built, not proposed.
  Note: the P3 peer edited SOUL.md, which its mandate assigned to the sibling — done on Gaetan's
  direct ruling and self-flagged. Backup `~/.hermes/SOUL.md.bak-skillcommit` alongside `.bak-p4`.

- 2026-08-07 23:0xZ — **Self-edit reconciled (question then still open, now resolved above).** git↔VM parity restored on
  Gaetan's word (`2087890`, `6e972b9`): both at md5 `a6e0a4b5…`, 512 lines, mode back to 0664,
  hub-verified. Base was Tars' own self-edited text (its structural fixes were right and were
  independently checked); the stale claims it had left standing beside its own corrections were
  stripped — notably the disproven "`--run` addressing works standalone" sentence sitting directly
  beside Tars' contradicting `--terminal` fix. **A live race was caught and handled**: Tars edited
  the file again at 22:55 mid-reconcile; the peer diffed the missed generation, folded both new
  findings in (self-check gates before `worktree rm --force`; the worktree `preview` is unreliable —
  use `terminal read`, bottom `❯` authoritative), and guarded the write with an md5 check that would
  have aborted rather than clobber. Rollback ladder intact on the VM: current (`a6e0a4b5`) ·
  `.bak-tars-selfedit` (`5cc32dfc`) · `.bak-p3` = v1 (`d61888ec`).
  **Parity is a SNAPSHOT, not an invariant** — Tars holds `skill_manage` and rewrote this file seven
  times in one turn; it can drift again on any run. The design question stands: *should the model be
  able to rewrite its own governing skill?* Not closed by this lane.
  Cooper v1 leftovers deleted 23:05Z, order deliberate (archive → verify → commit → push → delete);
  every file survives verbatim in `status/probes/wf5/p3-cooper-v1-leftovers.md` (399 lines,
  hub-verified) and `~/orca/workspaces/` now holds only `mc-metarepo/` and `Tars/`.
  Still unverified: the `terminal send --enter` fix (no run has exercised it end to end).

- 2026-08-07 — **OPEN (superseded above, kept for the record): Tars rewrites its own rules.** The live
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
