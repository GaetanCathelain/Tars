# Gmail forwarding idempotency

Validated against Gmail API in August 2026 for a raw-MIME forwarding workflow.

## Integration choice

When both IMAP/SMTP and Gmail OAuth are configured, Gmail API is preferable for this class of workflow because one integration provides:

- Gmail-native search and `internalDate`;
- raw RFC 5322 retrieval;
- attachment payloads;
- custom label creation and message modification;
- send and Sent-mail reconciliation.

Use an OAuth authentication check first and verify `users.getProfile("me")` matches the intended mailbox. Do not print token JSON or configuration containing auth commands.

## Filter discovery

Start with broad, read-only metadata searches, then narrow the action filter to the observed sender and stable subject phrase. Inspect one representative message's MIME tree but emit only headers, filenames, MIME types, attachment-presence booleans, and sizes. Do not emit bodies or encoded attachment data.

A useful strict action-query shape is:

```text
from:EXACT_SENDER subject:"STABLE SUBJECT PREFIX" has:attachment
```

For post-activation work, add a server-side cutoff and exclude the completion label:

```text
... after:CUTOFF_EPOCH_SECONDS -label:"Workflow/Forwarded"
```

Still compare each result's `internalDate` milliseconds against the exact persisted cutoff locally. Gmail's second-resolution `after:` query is only a coarse recall filter.

## Message-ID pitfall

Do **not** rely only on a caller-supplied `Message-ID` for Gmail send reconciliation. In a live Gmail API send, Gmail replaced the deterministic `Message-ID` with a Gmail-generated value. Consequently:

```text
in:sent rfc822msgid:DETERMINISTIC_MARKER
```

returned no match, while an exact search for the same deterministic marker embedded in the outgoing plain-text body did find the sent message:

```text
in:sent "DETERMINISTIC_MARKER"
```

Therefore embed a recipient-safe deterministic marker such as `workflow-SOURCE_ID@local-domain` in the body and reconcile by exact body search. A custom header may be retained, but Gmail search does not offer a dependable arbitrary-header lookup, so it is not sufficient by itself.

Allow time for Gmail indexing. If local state says `sending` and marker search is empty, defer the retry for a bounded grace period instead of immediately resending.

## State sequence

Recommended sequence under a single-writer lock:

1. Add `Pending` to source.
2. Persist source ID, internal date, marker, `sending`, and attempt time.
3. Build and send the forward.
4. Persist returned Gmail sent ID and `sent`.
5. Add `Forwarded`, remove `Pending`.
6. Persist `complete`.

Recovery:

- `complete` or source already carrying `Forwarded`: no send.
- `sent`: repair labels, then mark complete.
- `sending`: exact-search Sent by body marker; if found, persist sent ID and repair labels; if not found and still within the indexing grace period, defer.

## Attachment preservation

Fetch the source in `format=raw`, URL-safe-base64 decode it, and parse with Python's `BytesParser(policy=policy.default)`. Build a new `EmailMessage` from the authenticated mailbox, then copy each `iter_attachments()` payload with its original content type and filename. Test message construction read-only against a representative source and assert attachment count and MIME types before enabling send.

## Bounded-query starvation pitfall

Do not fetch a small page containing already-forwarded messages, filter them locally, and stop. After enough completed messages, they can occupy the entire result page and hide eligible candidates. Exclude the completion label in the Gmail query itself, then verify labels locally. Keep pending records unexcluded so reconciliation still occurs.

## Hermes cron detail

A Hermes cron script must live under `~/.hermes/scripts/`. Current CLI creation expects `--script` to receive the filename relative to that directory, not an absolute path. Use `--no-agent` for deterministic execution, owner-only permissions for script/state/lock, and confirm one scheduler-owned run via both `hermes cron runs JOB_ID` and the local output under `~/.hermes/cron/output/JOB_ID/`.
