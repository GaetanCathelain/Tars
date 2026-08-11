---
name: secure-delta-collectors
description: "Use when building bounded read-only API delta collectors."
version: 1.4.0
metadata:
  hermes:
    tags: [api, delta, security, pagination, orchestration]
    category: orchestration
---

# Secure delta collectors

Build small source adapters that retrieve an exact time-bounded delta from an external API without exposing credentials, raw pages, or unrelated content. The collector should be safe to invoke from an agent workflow: bounded input, bounded output, deterministic coverage semantics, and fail-closed behavior.

## Contract first

Before implementation, define:

- an exact interval convention, normally `start < event_timestamp <= end`;
- stable event identity and local deduplication keys;
- the smallest event fields downstream reconciliation needs;
- explicit page, scan, event, output, and time-window caps;
- coverage fields that distinguish complete, truncated, incomplete, and fail-closed runs;
- whether partial candidates are safe to return. For reminder/reconciliation pipelines, default to returning no events when coverage fails.

Use a fixed run end supplied by the caller. Never advance a durable cursor merely because retrieval started; only complete coverage permits cursor advancement.

## Security boundary

1. Load credentials internally from a protected file or secret provider. Never accept tokens as CLI arguments.
2. Reject credential files readable by group or others when local file permissions are meaningful.
3. Keep an allowlist of read-only API methods. Reject all other methods before constructing a request.
4. Use HTTPS with normal certificate verification and fixed API hosts.
5. Never emit headers, request objects, exception strings, response bodies, or raw API pages.
6. Redact both known credential values and recognizable token/secret forms from retained snippets.
7. Treat permalinks as untrusted input: require the expected HTTPS host and remove query strings and fragments.
8. Convert all failures, including unexpected exceptions, to short enumerated error codes. Exception text can contain headers, URLs, or response bodies.

## Retrieval workflow

1. Parse timezone-aware ISO timestamps and normalize them to UTC.
2. Widen coarse server-side date filters only enough to guarantee recall.
3. Fully paginate each required view until the server signals completion or a deterministic cap is reached.
4. Detect repeated cursors and malformed pagination metadata. When several endpoint flows use cursor pagination, route cursor extraction, type validation, repeat detection, seen-set mutation, and page advancement through one small strict helper. Parameterize endpoint-specific safe error codes rather than duplicating the state transition. Keep fundamentally different legacy page-number pagination separate instead of forcing it through the cursor abstraction.
5. Exact-filter every match locally against the interval.
6. Apply a conservative local relevance filter. Count raw exact matches separately; unrelated exact matches must not consume the retained-candidate budget.
7. Deduplicate overlapping views by stable source identity before enforcing the retained-event cap. The same event appearing in several required views must consume one candidate slot, not one slot per view.
8. Enforce the retained-event cap on unique candidates.
9. Fetch bounded context only for retained candidates and only when classification cannot be made from the candidate itself. Never fetch unrelated history.
10. Sort retained events deterministically before serialization.

Keep independent caps for raw matches scanned, retained candidates, pages, and context fetches. If any cap is hit before every required view is exhausted, mark the interval incomplete. Do not call a capped run complete merely because enough candidates were found. When a wide interval hits a cap, split it into deterministic adjacent subwindows and require every subwindow to complete before advancing the aggregate cursor.

## HTTP transport hardening

For credential-bearing requests, make redirect and response-size behavior explicit security boundaries rather than relying on client defaults:

- Disable all redirects in the HTTP transport. In Python `urllib`, build an explicit opener with `ssl.create_default_context()`, an `HTTPSHandler` using that context, and a custom `HTTPRedirectHandler` whose `redirect_request` raises. This prevents an `Authorization` header from being forwarded to a redirect target.
- After opening the response, require `response.geturl()` to equal the allowlisted endpoint exactly. Treat any mismatch as a transport failure even when redirects are disabled; this defense-in-depth check pins scheme, host, path, query, and fragment.
- Bound each response before parsing: read exactly `MAX_RESPONSE_BYTES + 1`, fail closed when the result exceeds `MAX_RESPONSE_BYTES`, decode UTF-8 strictly, then parse JSON. Never pass the response stream directly to an unbounded parser such as `json.load(response)`.
- Choose the byte cap from the maximum legitimate page shape, independently of pagination and emitted-output caps. A bounded number of pages does not bound bytes per page.
- Before clipping emitted descriptions, comments, or snippets, redact the exact loaded credential, its URL-encoded representation, labeled secret assignments, bearer tokens, and conservative recognizable token prefixes. Apply clipping after redaction so replacement text cannot violate the output bound. Keep recognizable surrounding prose so classification remains useful.

Verify these controls without real credentials by injecting a fake opener/response and asserting: TLS hostname checking and `CERT_REQUIRED`; every redirect raises; a mismatched final URL fails before body reads; reads request exactly cap-plus-one; an oversized body fails; valid bounded UTF-8 JSON succeeds; and secret forms disappear while ordinary context remains. Then run only a bounded read-only live probe and report aggregate metadata. For stateful workflows, hash durable state and record lock presence before and after the probe.

## Safe retries and rate limits

- Retry only explicitly transient failures. For HTTP 429, accept `Retry-After` only after strict bounded parsing; reject missing, negative, fractional, malformed, or excessive values.
- Bound both request attempts and cumulative sleep. Exhaustion must return the same compact fail-closed error shape as an immediate transport failure.
- Never include response bodies, headers, request URLs, or exception reasons in retry diagnostics; those surfaces can carry credentials or private source data.
- Inject both the HTTP transport and sleeper. Tests must cover success-after-retry, budget exhaustion, malformed retry metadata, and secret non-disclosure without real sleeping.
- Test the cumulative-wait inequality at the exact configured boundary, not only below and above it. Use independently chosen waits that sum exactly to the cap (for example 10 then 20 for a 30-second budget), followed by success; assert the exact sleep sequence, exact attempt count, successful payload, and the cap constant itself. This pins that `total_wait + retry_after <= budget` allows the boundary while any excess fails closed.
- A successful retry does not weaken coverage checks: every required view and page must still complete.

## Reconcile before committing the cursor

Complete retrieval is necessary but not sufficient when a downstream workflow derives durable items or transitions. Treat retrieval, chronological reconciliation, and cursor advancement as one transaction:

1. Prove the union of completed leaves covers the exact interval, with unique event counts and no partition overlap or gaps.
2. Reconcile all candidates chronologically across partition boundaries. Later replies, completion signals, delegation, and terminal states may close earlier apparent work.
3. Independently challenge proposed mutations when false positives are costly. Use bounded context only for unresolved retained candidates; ambiguity must never be promoted to a durable open item by guesswork.
4. Validate proposed additions against current durable state for stable-ID collisions and cross-source duplicates, and validate every transition's expected prior status.
5. Guard the write with the state hash and fixed target used during reconciliation. If either changed, abort and recompute.
6. Write atomically with owner-only permissions, preserve unknown state keys, retain a bounded dedup ring, and verify the resulting cursor, completion flags, item counts, and lock state.

Advance the cursor only after this semantic commit succeeds. A fully retrieved but unreconciled interval remains incomplete operationally.

## Output contract

Stdout is exactly one compact JSON document. Put diagnostics in bounded metadata, not stderr logs containing source data. A useful envelope is:

```json
{"coverage":{"source":"…","start":"…","end":"…","complete":false,"truncated":false,"incomplete":true,"fail_closed":true,"pages":0,"matches_scanned":0,"exact_matches":0,"deduplicated":0,"returned":0,"errors":[],"caps":{}},"events":[]}
```

Each event should contain only stable IDs, source timestamps, author/owner IDs, a validated evidence handle, bounded classification hints, and a short sanitized snippet. Do not include channel names, workspace names, full transcripts, raw objects, or unrelated bodies.

A `--probe` mode should call the smallest read-only authentication/metadata endpoint and return metadata only, with an empty events list. Do not reflect account, workspace, or private source metadata unless the caller explicitly needs it.

## Testing and verification

Use injected HTTP transports and synthetic fixtures. Cover at least:

- cursor pagination and any supported legacy page pagination;
- exact boundary filtering;
- overlapping-view deduplication;
- deterministic ordering;
- snippet length and secret redaction;
- HTTP/API/malformed-response failures;
- repeated cursor and cap-triggered incomplete states;
- unexpected exceptions converted to safe JSON;
- probe output containing no source message or private account content.

After tests pass, run the real metadata-only probe. Report only the compact result fields needed to verify authentication and completeness; never reproduce credentials or private content.

## Pitfalls

- API error bodies and exception messages are not safe diagnostics.
- Search date operators are transport filters, not exact cursor semantics.
- Returning the first N events while claiming complete coverage can skip later events permanently.
- A token-pattern regex alone is insufficient; also redact the exact loaded secret values and URL-encoded forms.
- Search results can expose private names in nested objects even when those fields are irrelevant. Build a new event object field-by-field instead of pruning raw responses.

## Provider notes

- For the tested Slack browser-session Web API pattern, see `references/slack-web-api.md`.
- For the supported OAuth V2 user-token alternative, exact scopes, lifecycle, visibility semantics, and current distribution/rate-limit constraints, see `references/slack-oauth-user-token.md`. Prefer this supported OAuth path over browser-session credentials when `search.messages` must run on a member's behalf.