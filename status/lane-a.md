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
| A5 | Slack Tars app xoxb/xapp (held for cutover) | `SLACK_BOT_TOKEN` `SLACK_APP_TOKEN` | ☐ | ☐ |
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
