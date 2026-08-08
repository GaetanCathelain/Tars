---
name: secure-delta-collectors
description: "Use when building bounded read-only API delta collectors."
version: 1.1.0
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
4. Detect repeated cursors and malformed pagination metadata.
5. Exact-filter every match locally against the interval.
6. Apply a conservative local relevance filter before the retained-event cap. Count raw matches scanned and retained candidates separately; unrelated exact matches must not consume the candidate budget.
7. Deduplicate overlapping views by stable source identity.
8. Fetch bounded context only for retained candidates and only when classification cannot be made from the candidate itself. Never fetch unrelated history.
9. Sort retained events deterministically before serialization.

Keep independent caps for raw matches scanned, retained candidates, pages, and context fetches. If any cap is hit before every required view is exhausted, mark the interval incomplete. Do not call a capped run complete merely because enough candidates were found. When a wide interval hits a cap, split it into deterministic adjacent subwindows and require every subwindow to complete before advancing the aggregate cursor.

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

For the tested Slack browser-session Web API pattern, see `references/slack-web-api.md`.