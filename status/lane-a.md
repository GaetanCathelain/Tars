# Lane A — credential acquisition (orchestrator session only writes here)

Names only. Values live in `secrets/tars.sops.yaml` (SOPS+age, 2 recipients) or nowhere
(self-managing OAuth). Probe = the async agent evidence that the credential works.

## Checklist

| # | Credential | SOPS key(s) | Harvested | Probed |
|---|---|---|---|---|
| A1 | GitHub — `gh auth login` device-flow ON the VM (Gaetan's call; replaces the PAT) | — (VM gh store, never SOPS) | ◐ staged, awaiting authorize | ☐ |
| A1b | ChatGPT-sub OAuth (`openai-codex`, on the VM) | — (never stored) | ☐ | ☐ |
| A2 | Linear API key (paste-file `linear_api_key.secret` → SOPS → shred) | `LINEAR_API_KEY` | ☐ awaiting paste | ☐ |
| A2 | Notion — reused from mc-kestra `.env` (Gaetan's call) | `NOTION_TOKEN_V2` `NOTION_FILE_TOKEN` `NOTION_SPACE_ID` | ☑ 2026-08-07 | ◐ probe running |
| A3 | Gmail app password (Tars-labelled, 2nd one) | `GMAIL_ADDRESS` `GMAIL_APP_PASSWORD` | ☐ | ☐ |
| A3 | Calendar (rides Gmail app password) | (same) | — | ☐ |
| A4 | Slack personal xoxc/xoxd | `SLACK_TOKEN` `SLACK_COOKIE` | ☐ | ☐ |
| A5 | Slack Tars app xoxb/xapp (held for cutover) | `SLACK_BOT_TOKEN` `SLACK_APP_TOKEN` | ☐ | ☐ |
| — | Tailscale auth key (single-use, → lane B/WF3) | not stored (0600 handoff file) | ☐ blocked: console 500 | — |
| — | Hindsight keys (2×) | — | ✗ SKIPPED for v1 (Gaetan, 2026-08-07) | — |

## Log

- 2026-08-07 — SOPS store bootstrapped (`scripts/tars-secret`, `.sops.yaml`); dummy-roundtrip
  verified. VM age key enrolled as 2nd recipient (canary: 2 recipients in metadata, cooper
  decrypt ok).
- 2026-08-07 — A1b path settled by R7 + grill: fresh device-code OAuth on the VM, never copied,
  never stored. Runs at WF3 time.
- 2026-08-07 — Tailscale console 500 → key mint deferred to WF3 (LAN primary per D4).
- 2026-08-07 — Browser flow flipped: Gaetan's own browser + `tars-secret` hidden prompt
  (DevTools-bridge browser has no sessions; every service would cost a fresh 2FA there).
- 2026-08-07 — Lane B complete (smoke pass, e395e07). Join now waits on lane A only.
