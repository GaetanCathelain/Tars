---
name: durable-email-automation
description: "Use when automating durable inbound email actions."
version: 1.0.0
created_by: agent
metadata:
  hermes:
    tags: [email, automation, idempotency, gmail, cron, mime]
    category: email
---

# Durable email automation

Build unattended email workflows that discover narrowly scoped inbound messages and perform an outbound or mailbox mutation without replaying history, losing attachments, leaking credentials, or duplicating side effects after retries.

## Establish the real contract first

1. Confirm the exact mailbox and explicitly authorized recipients/actions.
2. Inspect configured integrations using authentication checks and account metadata only. Never print credential files, tokens, auth commands, or secret-bearing config.
3. Before inventing a filter, search the mailbox read-only and emit only the minimum metadata needed to identify real senders, subjects, dates, and attachment shapes.
4. Convert the observed pattern into a strict positive filter: exact sender, stable subject phrase, expected attachment condition, and mailbox scope. Avoid broad product-name searches in the live action path.
5. Count and clearly separate pre-existing matches from messages arriving after activation.

## Safe activation without backlog surprise

Use a two-phase activation:

- **Dry run while uninitialized:** report the exact filter and backlog count, but create no outbound side effect.
- **First live run:** atomically record a source-native cutoff (for Gmail, compare `internalDate` milliseconds locally), create explicit state labels, and treat every pre-cutoff match as backlog. Do not forward backlog unless the user separately authorizes it.

A server-side date query is only a recall optimization. Recheck the exact source timestamp locally against the persisted cutoff.

## Durable idempotency protocol

Email send APIs generally do not provide a transactional coupling between sending and labeling the source. Close the retry gap with several independent controls:

1. Acquire a single-writer file lock for the whole run.
2. Maintain an owner-only local journal keyed by immutable source message ID.
3. Use explicit source labels, normally `Pending` before send and `Forwarded` after successful send.
4. Give every outgoing action a deterministic marker derived from the source ID and place it in a recipient-safe, provider-searchable part of the outgoing message body.
5. Persist `sending` before calling the send API, then `sent` with the returned ID, then label the source and persist `complete`.
6. On retry from `sending`, reconcile the Sent mailbox by exact marker before considering another send.
7. If reconciliation is temporarily inconclusive, defer rather than resend. Use a bounded grace period for provider indexing; only retry the side effect after that period and another reconciliation attempt.
8. Treat `sent` but not `complete` as a label-repair operation, never as permission to resend.

Do not claim mathematical exactly-once semantics from a non-transactional email API. Report the concrete crash-recovery protocol and remaining provider assumptions.

## Bounded candidate processing

- Exclude the completed label in the server-side candidate query, then verify labels locally.
- Keep pending/uncertain records eligible for reconciliation.
- Apply a strict per-run maximum and a time bound below the scheduler timeout.
- Paginate or query so processed messages cannot occupy the bounded result window forever and starve newer or older unprocessed candidates.
- Process deterministically, usually oldest eligible first.
- Emit compact aggregate JSON; avoid dumping subjects, bodies, attachment contents, or raw API responses on routine cron runs.

## MIME preservation

Prefer an API that can retrieve raw RFC 5322/MIME and modify labels. Parse the original message, create a new forward from the authenticated mailbox, retain minimal original provenance in the body, and reattach each original attachment with its filename and content type. Verify with a real read-only sample that the expected attachment count and MIME types survive message construction.

Do not spoof the original sender. Set `From` to the authenticated mailbox and `To` only to the explicitly authorized recipient.

## Separately authorized historical backfills

When a user later authorizes an exact pre-cutoff backlog, do not weaken or reuse the routine post-cutoff query ad hoc. Freeze the exact authorized source IDs in a SQLite manifest, pin the expected count and recipient, use dedicated backfill labels/workflow state, and process it through bounded restartable batches. The scheduled invocation must remain on the routine post-cutoff path.

Fail closed if the locally checked pre-cutoff count differs from either the initialization backlog count or the newly authorized total. Final verification must reconcile the frozen manifest against Gmail Sent by deterministic body markers, exact recipient, unique Sent IDs, source labels, and aggregate attachment counts. Then run one no-op retry to prove completed items are not selected again.

See `references/historical-backfill.md` for the full execution and verification pattern.

## Scheduler deployment

For Hermes, prefer a local `no-agent` cron script for deterministic workflows:

1. Put the script under `~/.hermes/scripts/` with owner-only permissions.
2. Register `--script` using the filename relative to that directory, even when reporting the absolute path to the user.
3. Use local delivery unless the user requested notifications; keep routine output compact.
4. Verify the active job, ID, schedule, next run, and gateway heartbeat.
5. Verify at least one scheduler-owned execution in `hermes cron runs` and inspect its local output. A direct script run proves script behavior but not scheduler wiring.

## Verification checklist

- Authentication check names the intended mailbox without exposing secrets.
- Read-only discovery proves the live sender/subject/attachment pattern.
- Dry run reports backlog separately.
- First live run creates the cutoff and sends no pre-cutoff item.
- Labels and journal exist with owner-only local permissions.
- MIME construction retains expected attachments.
- A second run does not duplicate completed work.
- Sent-mail reconciliation finds the deterministic marker used by the implementation.
- `hermes cron list` and `hermes cron runs` show an active, successful scheduler execution.
- Final report includes job ID, schedule, absolute script path, integration, observed filter, backlog count, real outcomes, recipient scope, and blockers.

## Provider-specific references

- Gmail API behavior and the validated retry/search pattern: `references/gmail-forwarding-idempotency.md`.
