# Gmail app password + Calendar credential probe — evidence

Async probe, 2026-08-07. `GMAIL_ADDRESS` and `GMAIL_APP_PASSWORD` decrypted from
`secrets/tars.sops.yaml` on cooper via `sops -d --output-type json`, piped
straight into a restricted (0600, `umask 077`) file under the session
scratchpad, never printed/echoed/logged. Credentials passed to `curl` via
`--netrc-file` (never on argv). Netrc + decrypted-secrets files `shred -u`'d
immediately after each probe run. No values, no mailbox/calendar contents
below or at any point in this transcript.

## Probe 1 — IMAP (D5 row A3, Gmail)

```
curl -s --netrc-file <netrc, 0600, shredded after> --url imaps://imap.gmail.com
```

- Exit code: `0`
- stderr: empty
- Mailbox list returned: **11 folders** (line count of the default `LIST` response)

**Verdict: PASS** — mailbox list returned, exit 0, no auth error.

## Probe 2 — CalDAV (D5 row A3, Calendar — rides on the Gmail app password)

```
curl -s -X PROPFIND -H "Depth: 0" --netrc-file <same netrc> \
  --url https://www.google.com/calendar/dav/<address>/events
```

- HTTP status: **207** (Multi-Status)
- Response envelope: WebDAV `<D:multistatus>` root, one embedded per-property
  status line reading `HTTP/1.1 200 OK` — the standard successful-PROPFIND
  shape (RFC 4918), not an error wrapped at 207. Body size 945 bytes,
  inspected only for tag names/status strings, never rendered in full.
- Note vs D5's literal criterion ("PASS = HTTP 200"): PROPFIND replies are
  defined by the WebDAV spec to come back as `207 Multi-Status` on success,
  never a bare `200`; the inner status line is where the real 200 lives. This
  is the endpoint behaving correctly, not the documented endpoint-retirement
  risk (`401` alongside a passing IMAP probe) — no `CALDAV-RETIRED` flag.

**Verdict: PASS** — 207 envelope with an inner `200 OK`, i.e. a healthy
authenticated PROPFIND response.

## Summary

| Probe | Verdict | One-line reason |
|---|---|---|
| Gmail app password (IMAP) | **PASS** | `curl imaps://` exit 0, 11-folder mailbox list returned. |
| Google Calendar (CalDAV) | **PASS** | PROPFIND → HTTP 207 envelope, inner status `200 OK`; not the 401 retirement signature. |

Mechanics: SOPS decrypt → restricted-file → `--netrc-file` → `shred -u` worked
cleanly on both probes; no credential ever touched argv, stdout, or this
transcript.
