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

Use multiple narrow search views when necessary, then exact-filter and deduplicate locally by `channel_id + message_ts`. Search date operators should be widened around the exact UTC interval because they are coarse retrieval filters.

Support both `response_metadata.next_cursor` and legacy `messages.paging` when fixtures or deployed behavior require them. Detect repeated cursors. A page, scan, or event cap must produce `truncated=true`, `incomplete=true`, and `fail_closed=true`; do not return candidates from an incompletely covered interval when downstream code could advance its cursor.

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
3. dedup across overlapping views;
4. Unicode snippet caps;
5. exact and pattern-based token redaction;
6. HTTP failures and unexpected exceptions return only safe codes;
7. probe discards private fields from `auth.test` and emits no events.

Finish with a live `--probe`; it should report authentication and coverage booleans only, with an empty event array.