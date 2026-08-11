# Slack Web API delta collector notes

Validated in August 2026 for an engagement-checker source adapter.

## Authentication boundary

For a Slack browser-session integration with both protected values available:

- load `SLACK_MCP_XOXC_TOKEN` and `SLACK_MCP_XOXD_TOKEN` from the protected env file inside the process;
- send the xoxc value only as `Authorization: Bearer …`;
- send the xoxd value only as the `d=…` cookie;
- never accept either token on argv or include request/exception details in output;
- validate expected token prefixes and require the credential file to have no group/other permission bits.

Do not generalize this authentication scheme to ordinary Slack apps: bot/user OAuth integrations may use different token types and should follow their own supported authentication model.

## Read-only collection shape

Use an explicit allowlist such as:

- `auth.test` for a metadata-only probe;
- `search.messages` for bounded candidate discovery;
- `conversations.replies` only when a retained candidate needs thread context.

Use multiple narrow search views when necessary, then exact-filter and deduplicate locally by `channel_id + message_ts`. For Slack's user-ID search form, use `from:<USER_ID>` for messages authored by the user and `with:<@USER_ID>` for threads and DMs involving the user; a bare `<@USER_ID>` query finds explicit mentions only and is not an involving-user substitute. Search date operators should be widened around the exact UTC interval because they are coarse retrieval filters.

For engagement collection, the Slack search modifiers have distinct scopes and should not be collapsed into one view:

- `from:<@USER_ID>` finds messages authored by the person;
- `with:<@USER_ID>` finds DMs and threads with the person; this exact ID form was accepted by a live `search.messages` probe;
- a bare `<@USER_ID>` term finds explicit mentions, including standalone channel mentions that are not guaranteed to be covered by `with:`.

Deduplicate across all required views before counting candidates. Otherwise one message returned by both `from:` and `with:` can consume multiple slots and trigger a false event-cap failure.

Support both `response_metadata.next_cursor` and legacy `messages.paging` when fixtures or deployed behavior require them. Detect repeated cursors. A page, scan, or event cap must produce `truncated=true`, `incomplete=true`, and `fail_closed=true`; do not return candidates from an incompletely covered interval when downstream code could advance its cursor.

## Coarse-date and rate-limit interaction

Slack date operators are date-granular transport filters. Splitting one day into tiny exact-time subwindows still causes each subwindow to scan much of the same date-filtered result set before local timestamp filtering. Do not assume recursive time splitting reduces API work; budget raw pages and requests independently from retained candidates.

On HTTP 429, parse `Retry-After` as a non-negative decimal integer only. Use fixed limits for both HTTP attempts and cumulative wait; invalid, missing, or over-budget values fail closed immediately. Reuse the original read-only request, sleep only through an injected sleeper, and never expose headers, response bodies, request URLs, or exception reasons. Deterministic tests should cover success after one retry, exhausted attempts, excessive wait, malformed metadata, and credentials planted separately in the error reason, headers, and body. A retry that eventually succeeds still must satisfy normal pagination and interval-coverage proof.

For broad historical catch-up, deterministic adjacent subdivision plus throttled retries can recover complete coverage, but failed parent attempts are not coverage. The aggregate is complete only when every final leaf succeeds, the leaf intervals exactly cover the target, and the deduplicated union is reconciled before the durable cursor advances.

## Minimal event shape

A bounded Slack candidate can contain:

- stable ID: `slack:<channel_id>:<message_ts>`;
- `channel_id`, `thread_ts`, and `message_ts`;
- author user ID;
- validated Slack permalink or fallback evidence handle;
- a few conservative classification hints;
- a whitespace-normalized snippet capped at 240 characters.

Do not retain channel/workspace names, raw matches, full threads, or nested source objects.

## Synthetic verification pattern

Inject an HTTP opener into the client rather than monkeypatching global network behavior. Queue compact fixture responses and inspect requests only inside tests. Verify:

1. cursor and page pagination;
2. exact start-exclusive/end-inclusive filtering;
3. all required query views (`from:`, `with:`, and explicit mention where applicable);
4. relevance filtering followed by cross-view deduplication before the unique-candidate cap;
5. deterministic ordering;
6. Unicode snippet caps;
7. exact and pattern-based token redaction;
8. HTTP/API failures and unexpected exceptions return only safe codes;
9. probe discards private fields from `auth.test` and emits no events.

Finish with a live `--probe`; it should report authentication and coverage booleans only, with an empty event array. Validate search syntax with count-one, metadata-only probes; never print raw message bodies while probing.