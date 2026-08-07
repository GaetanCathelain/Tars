# Lane A — credential acquisition (orchestrator session only writes here)

Names only. Values live in `secrets/tars.sops.yaml` (SOPS+age, 2 recipients) or nowhere
(self-managing OAuth). Probe = the async agent evidence that the credential works.

## Checklist

| # | Credential | SOPS key(s) | Harvested | Probed |
|---|---|---|---|---|
| A1 | GitHub — `gh auth login` device-flow ON the VM (Gaetan's call; replaces the PAT) | — (VM gh store, never SOPS) | ☑ 2026-08-07 | ☑ `repo`+`read:org`, metarepo `.private=true` from the VM |
| A1b | ChatGPT-sub OAuth (`openai-codex`, on the VM) | — (never stored) | ☐ | ☐ |
| A2 | Linear API key (paste-file → SOPS → shred) | `LINEAR_API_KEY` | ☑ 2026-08-07 | ☑ 200, viewer = Gaëtan Cathelain |
| A2 | Notion — reused from mc-kestra `.env` (Gaetan's call) | `NOTION_TOKEN_V2` `NOTION_FILE_TOKEN` `NOTION_SPACE_ID` | ☑ 2026-08-07 | ☑ token_v2 valid (loadUserContent 200 vs control 401); file_token deferred to first export. MCP-compat gate open (r8) |
| A3 | Gmail app password (Tars-labelled, 2nd one) | `GMAIL_ADDRESS` `GMAIL_APP_PASSWORD` | ☑ 2026-08-07 | ☑ IMAP list, 11 folders |
| A3 | Calendar (rides Gmail app password) | (same) | — | ☑ CalDAV 207 + inner 200 |
| A4 | Slack personal xoxc/xoxd | `SLACK_TOKEN` `SLACK_COOKIE` | ☐ | ☐ |
| A5 | Slack Tars app xoxb/xapp (held for cutover) | `SLACK_BOT_TOKEN` `SLACK_APP_TOKEN` | ☐ | ☐ |
| — | Tailscale auth key (single-use) | never stored — spent + shredded | ☑ 2026-08-07 | ☑ VM joined: `tars` 100.116.31.76, ping 1ms, LAN primary |
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
