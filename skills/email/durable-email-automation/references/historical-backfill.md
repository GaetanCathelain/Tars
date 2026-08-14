# Authorized historical Gmail backfills

Use this pattern when the normal automation intentionally excludes a pre-activation backlog and the user later authorizes an exact historical batch.

## Freeze the authorized set before sending

1. Keep the routine post-cutoff path and schedule unchanged.
2. Discover all messages with the same strict positive filter as the routine workflow.
3. Recheck Gmail `internalDate` locally and retain only messages strictly before the persisted cutoff.
4. Fail closed unless the count equals both the user-authorized total and the initialization backlog count.
5. Persist the exact source IDs, timestamps, and deterministic order in a SQLite manifest before the first send. Subsequent retries must read the manifest rather than rediscovering the mailbox.
6. Pin the authorized recipient and expected count in metadata. Refuse execution if either differs on retry.

This prevents later mailbox arrivals, query drift, or a post-cutoff message from entering the backfill.

## Separate state from the routine workflow

Use dedicated Pending/Forwarded labels and a workflow discriminator in the journal. Store, at minimum:

- immutable source ID and source `internalDate`;
- workflow name and exact recipient;
- deterministic body marker;
- `sending`, `sent`, or `complete` status;
- attempt timestamp and returned Gmail Sent ID;
- attachment count and bounded last error.

The source ID can remain the journal primary key when one source may only ever be acted on once. If multiple authorized actions per source are possible, key by `(workflow, source_id, recipient)` instead.

## Bounded execution

Expose a hard maximum per invocation and reject values above it. Process oldest-first from the frozen manifest, add a small delay between successful sends, and run as multiple restartable batches. A retry after completion must select zero candidates.

Do not add backfill flags to the scheduled cron invocation. Backfill is a separately invoked one-shot mode; the daily scheduler continues to run the post-cutoff path.

## Verification

Verify both local state and Gmail Sent, without printing bodies or attachments:

- manifest count equals the authorized total;
- every manifest row is `complete`, with no pending status or error;
- every row records the exact authorized recipient;
- Sent IDs and deterministic markers are unique;
- each marker occurs exactly once in Sent;
- every matching Sent message has exactly one allowed `To` address and no CC/BCC recipients;
- source messages carry the dedicated completion label and not the pending label;
- aggregate attachment count reconstructed from sources equals the aggregate count in Sent;
- no manifest timestamp is at or after the cutoff;
- the routine live path reports zero candidates for the already-processed post-cutoff item.

Finally execute the backfill mode once more. A successful zero-candidate retry is concrete evidence that the normal completion path is idempotent.
