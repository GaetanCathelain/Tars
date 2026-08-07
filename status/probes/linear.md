# Linear credential probe — evidence

Async probe, 2026-08-07. `LINEAR_API_KEY` decrypted from `secrets/tars.sops.yaml`
on cooper via `sops -d --extract`, piped straight into a `curl -K` header
config file (0600, `umask 077`, written under the session scratchpad) —
never printed, echoed, logged, or placed on argv. Header file was `shred -u`'d
immediately after the request. Zero secret material below or at any point in
this transcript.

## Probe (D5 row A2 / mc-kestra SETUP.md §5 Linear)

```
curl -K <cfg with 'Authorization: <bare key, no Bearer prefix>'> \
  -X POST https://api.linear.app/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ viewer { id name email } }"}'
```

- **HTTP 200**
- Response body: `{"data":{"viewer":{"id":"4951b192-e49c-4b7e-b491-58c89e66043c","name":"Gaëtan Cathelain","email":"gaetan.cathelain@mobile.club"}}}`
- Matches SETUP.md §5's healthy signature exactly: a `viewer` object
  returned on the first call, no entity dump needed to confirm liveness.
- Header shape confirmed correct per D5/SETUP.md gotcha: bare key,
  no `Bearer ` prefix — a malformed/`Bearer`-prefixed header would have
  produced the documented fast 401, not this 200.

## Verdict

| Key | Verdict | One-line reason |
|---|---|---|
| `LINEAR_API_KEY` | **PASS** | HTTP 200 with a populated `viewer` object (id/name/email) — exact healthy signature per D5 and mc-kestra SETUP.md §5. |

Identifies the account (Gaëtan Cathelain, `gaetan.cathelain@mobile.club`) —
safe to record, no key material. SOPS decrypt + header-file + shred
mechanics worked cleanly; no `.cursor`/dump semantics touched, this is a
liveness probe only, not a harvest run.
