# Slack personal account (xoxc/xoxd) probe — evidence

Async probe, 2026-08-07. Per D1 + D5 row A4 (`docs/recon/DECISION.md`). `SLACK_TOKEN` and
`SLACK_COOKIE` decrypted **individually** from `secrets/tars.sops.yaml` via
`sops -d --extract '["SLACK_TOKEN"]'` / `'["SLACK_COOKIE"]'` — never a whole-file `sops -d`.
Values held only in shell variables, never printed/echoed/logged, never placed on argv. `curl`
read them from a `-K` config file written with the `printf` builtin under `umask 077`; the config
file was `shred -u`'d in an EXIT trap immediately after each call. No token/cookie material appears
below or anywhere else in this transcript.

## Probe 1 — `auth.test` (D1, D5 row A4)

`POST https://slack.com/api/auth.test`, token as a `data` form field, cookie as `Cookie: d=<value>`
header, both via `-K`.

- HTTP status: **200**
- Response: `ok:true`
- **user:** `gaetan.cathelain`
- **user_id:** `U08BDJAMSRZ` — the value Blocker 2 / `SLACK_ALLOWED_USERS` needs
- **team:** `Mobile Club` (`team_id: T7V1UGJ82`)
- **url:** `https://mobileclub-squad.slack.com/`

**Verdict: PASS.**

## Probe 2 — cookie encoding check (D1 open question)

- Stored `SLACK_COOKIE` value **does contain percent-escapes** (checked with a pattern match on
  the shell variable, value never printed).
- Sent to Slack **exactly as stored** (no additional encoding applied) in probe 1's `Cookie: d=`
  header — this is the form that returned `ok:true` on the first attempt. No fallback/re-encode
  attempt was needed.

**Verdict: the as-stored form works directly.** This **deviates from D1's stated convention**
("the store holds it URL-decoded by convention; the consumer encodes"): the value in
`secrets/tars.sops.yaml` is already percent-encoded, not decoded. Flag for whoever wires
`korotovsky/slack-mcp-server`: if that server assumes an unencoded input and applies its own
URL-encoding step, it will **double-encode** this value and fail auth. Either strip the encoding
before storing (to match the documented convention) or confirm the MCP server accepts an
already-encoded cookie as-is before deploying it unchanged.

## Probe 3 — `conversations.list` cheap read (D1 binding constraint: no `users.list`)

`POST https://slack.com/api/conversations.list`, `limit=1`, `types=public_channel`, same token +
cookie delivery as probe 1. No user enumeration performed anywhere in this probe.

- HTTP status: **200**
- Response: `ok:true`
- **Channels returned:** 1 (as requested via `limit=1`)

**Verdict: PASS** — proves a working authenticated session (not just a syntactically valid
token), without touching `users.list`.

## Summary

| Probe | Verdict | One-line reason |
|---|---|---|
| `auth.test` | **PASS** | `ok:true`, HTTP 200, session identity resolved. |
| Cookie encoding | **as-stored form works** | Stored value already percent-encoded (deviates from D1's stated decoded-by-convention); no re-encoding needed for this raw curl path — flagged for MCP server wiring. |
| `conversations.list` (limit=1) | **PASS** | `ok:true`, 1 channel returned, no `users.list` call made. |

**Recorded for Blocker 2 / `SLACK_ALLOWED_USERS`:** `user_id = U08BDJAMSRZ` (not a secret).
