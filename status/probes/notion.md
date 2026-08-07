# Notion credential probe — evidence

Async probe, 2026-08-07. Keys decrypted from `secrets/tars.sops.yaml`
(`NOTION_TOKEN_V2`, `NOTION_FILE_TOKEN`, `NOTION_SPACE_ID`) into shell
variables only — never printed, echoed, or logged. Headers/body passed to
`curl` via `-K` config files and `--data @file`, never on argv. All temp
files holding decrypted values were shredded after use. Zero secret
material below.

## Probe 1 — literal prior-art replay (D5 / r6-credentials.md §3)

```
curl -s -K <cfg with 'Cookie: token_v2=$NOTION_TOKEN_V2'> \
  --data '{"query":"","spaceId":"$NOTION_SPACE_ID"}' \
  https://www.notion.so/api/v3/search
```

- **HTTP 400**, `time_total` 0.264s
- Envelope keys: `isNotionError, errorId, name, debugMessage, message`
- `name: "ValidationError"`, `debugMessage: "Invalid input."` (generic, no
  field named)
- No token/secret material present anywhere in the response body.

Not the documented healthy signature (200 + results envelope).

## Probe 2 — fuller current-schema payload (diagnostic, same endpoint/creds)

Added the fields Notion's `/api/v3/search` is known to expect today
(`type`, `filters`, `sort`, `limit`, `source`) on top of the same
`token_v2` cookie + `spaceId`.

- **HTTP 400**, `time_total` 0.240s
- Identical envelope shape, `name: "ValidationError"`, `debugMessage:
  "Invalid input."`

Still not a pass. Payload shape guessing stops here (time-boxed, per
workflow pref — this is a diagnostic aid, not a new standing probe).

## Probe 3 — control test: bogus cookie, minimal (prior-art) payload

Same request as Probe 1, but `Cookie: token_v2=bogus-invalid-token-for-control-test`
and a dummy all-zero `spaceId` — no real credential material involved.

- **HTTP 400**, `time_total` 0.235s
- **Identical** envelope/error shape to Probe 1 (`ValidationError`,
  `debugMessage: "Invalid input."`, same 5 envelope keys).

## Interpretation

Probe 3 proves the 400 is **payload/schema-shaped, not auth-shaped**: a
provably-bogus token produces the exact same error as the real
`NOTION_TOKEN_V2`/`NOTION_SPACE_ID` pair. Notion's internal
`/api/v3/search` schema has drifted since mc-kestra's SETUP.md §5 / R6
captured the minimal `{"query","spaceId"}` shape — the endpoint now
rejects that body before it would ever reach an auth check, so this probe
cannot currently distinguish "credential valid" from "credential invalid."
The prior art's documented FAIL signature (401 = re-harvest) never
triggered in any of the three attempts; a real 401 would still be
diagnostic if it occurs later with an updated payload.

**Do not route this to WF3 as "re-harvest token_v2"** — that fix target is
wrong per the control test. The actual gap is the probe's request schema
needing an update to match Notion's current API contract, which is outside
this probe's scope to fully reverse-engineer (only a partial guess was
tried in Probe 2, itself not a pass either).

## Verdict

| Key | Verdict | One-line reason |
|---|---|---|
| `NOTION_TOKEN_V2` | **INCONCLUSIVE (reported as FAIL against the stated PASS bar)** | Documented probe returns HTTP 400 for both the real token and a bogus control token — the credential's validity is not determined by this test; Notion's search API schema has drifted from the prior-art shape. |
| `NOTION_SPACE_ID` | **INCONCLUSIVE (reported as FAIL against the stated PASS bar)** | Same request/response as above — payload including `spaceId` never reaches a state where auth/space validity would be distinguishable. |
| `NOTION_FILE_TOKEN` | **DEFERRED** | Per D5: not cheaply probeable; first real export is its probe. No probe attempted, none invented. |

All three keys decrypted successfully from SOPS (non-empty, confirmed by
length only, never by value) — the SOPS store itself is healthy.

## Probe 4 — coordinator follow-up: `loadUserContent`, real token_v2

Needs no request-body schema knowledge (`{}`), so it sidesteps the
`/api/v3/search` schema-drift problem above.

```
curl -s -K <cfg with 'Cookie: token_v2=$NOTION_TOKEN_V2'> \
  --data '{}' \
  https://www.notion.so/api/v3/loadUserContent
```

- **HTTP 200**, `time_total` 0.420s
- Envelope: `{"recordMap": {...}}` — single top-level key `recordMap`,
  body 126077 bytes.

## Probe 5 — control: `loadUserContent`, bogus token

Same request, `Cookie: token_v2=bogus-invalid-token-for-control-test`.

- **HTTP 401**, `time_total` 0.202s
- Full body (no secret material — this is the bogus placeholder string,
  not a real credential):
  ```
  {"isNotionError":true,"errorId":"ec80af27-517c-4d4c-a40e-8dce895cbb4a",
   "name":"UnauthorizedError","clientData":{"type":"login_try_again"},
   "debugMessage":"Token was invalid or expired.","message":"Something went wrong. (401)"}
  ```

## Settled verdict

Probes 4 vs 5 diverge in an unambiguously auth-shaped way — status code
(200 vs 401) **and** error identity (`UnauthorizedError` /
`"Token was invalid or expired."` only on the bogus token). The real
`NOTION_TOKEN_V2` returns a populated `recordMap` for a live session.

**`NOTION_TOKEN_V2`: PASS (valid).** `NOTION_SPACE_ID` is not exercised by
this endpoint, so it stays unconfirmed by this probe specifically, but
`token_v2`'s validity is no longer in question. `/api/v3/search`'s 400 in
Probes 1–3 is confirmed to be pure schema drift on Notion's side (WF3
follow-up: update the search probe's payload to match Notion's current
`/api/v3/search` contract — not a re-harvest). `NOTION_FILE_TOKEN` stays
**DEFERRED** per D5.
