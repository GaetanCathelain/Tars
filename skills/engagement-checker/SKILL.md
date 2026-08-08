---
name: engagement-checker
description: "Use for incremental follow-up and commitment reminders."
version: 1.1.0
metadata:
  hermes:
    tags: [engagement, reminders, slack, email, linear, orchestration]
    category: orchestration
---

# Engagement checker

A half-hourly, source-backed monitor for Gaetan's open loops: explicit commitments he made, direct asks he has not answered, and blockers, reviews, or decisions waiting on him. It is not a daily summary. `daily-work-brief` owns the broad outcome view; this skill processes deltas and re-evaluates only a compact durable pending queue.

Tars owns and executes this workflow directly. Do not spawn, prompt, resume, modify, or delegate through Claude, Orca, an Orca worktree/session, or another non-Tars agent system. Cooper and Orca may be queried read-only as evidence sources. Never send Slack messages, email, reactions, or Linear updates from source integrations; the only write outside Tars's state is the scheduled reminder delivered to Gaetan.

All human times use `Europe/Paris`. Scheduled runs are restricted to 10:00–17:00 on workdays. Most runs must return exactly `[SILENT]`.

## Load only what is needed

- Use the Slack MCP tools directly. Gaetan's Slack user ID is `U08BDJAMSRZ`.
- Load `google-workspace` for Gmail. Its existing OAuth is the primary mailbox path; if its auth check fails, load and check `himalaya` before declaring email unavailable.
- Query Linear read-only on Cooper with `ssh cooper '~/.local/bin/orca linear … --json'`. Never use a Linear mutation command.
- Load `hermes-agent` only when installing or troubleshooting the cron jobs.
- Consult `daily-work-brief` only for its workday-gate semantics. Do not load or run its broad collection workflow during ordinary checks; the objectives and state are separate.

## Workday gate

Apply the same workday decision as `daily-work-brief` before normal delta collection:

1. Today must be Monday–Friday in Europe/Paris.
2. Query the official metropolitan-France holiday calendar at `https://calendrier.api.gouv.fr/jours-feries/metropole/<YEAR>.json` and match today's Paris-local ISO date.
3. Search existing email access for PayFit messages that explicitly cover today. Read the relevant body; a subject or snippet alone is not proof. Approved leave, RTT, or absence is off evidence; a generic PayFit notification is not.

If today is a weekend, metropolitan-France bank holiday, or explicit PayFit day off, make the same bounded actual-work probe as the daily: Gaetan-authored Slack, merged PRs, and Cooper Claude activity. If there is no substantive work signal, return exactly `[SILENT]` without acquiring the engagement lock or advancing source cursors. If there is substantive work, continue; actual work overrides the nominal day-off gate, exactly as for the daily.

If holiday or leave evidence cannot be checked, continue rather than silently skipping. Mention the coverage failure only if it materially weakens a reminder judgment. The cron schedule itself is weekday-only; the gate also protects manual runs and excludes holidays and leave.

## Durable state

The source of truth is `~/.hermes/state/engagement-checker.json`. It is operational state, not model memory and not git. Create the parent directory if needed. Read it at the start and overwrite it after every successfully reconciled source, so one later source failure cannot lose earlier progress.

Use this bounded shape:

```json
{
  "version": 1,
  "initialized_at": "ISO timestamp",
  "last_completed_run": "ISO timestamp or null",
  "sources": {
    "slack":  {"cursor": "ISO timestamp", "last_success": "ISO or null", "failures": 0, "seen": []},
    "email":  {"cursor": "ISO timestamp", "last_success": "ISO or null", "failures": 0, "seen": []},
    "linear": {"cursor": "ISO timestamp", "last_success": "ISO or null", "failures": 0, "seen": []}
  },
  "items": {},
  "last_failure_notice": {},
  "ambiguous_instructions": []
}
```

Each `items` value contains only:

- `id`, `short_id`, `source`, `kind`, `title`, `people`, `asker`;
- `created_at`, `updated_at`, `due_at`, `due_phrase`, `due_inferred`;
- `status`: `open`, `snoozed`, `waiting`, `done`, or `dismissed`;
- `snooze_until`, `stakeholder`, `someone_waiting`;
- a source evidence handle and a snippet of at most 240 characters;
- `last_score`, up to 10 reminder records, and up to 10 status-history records.

Never persist raw Slack threads, email bodies, Linear descriptions/comments, credentials, tokens, headers, or private transcript dumps. Redact obvious secret forms; when uncertain, omit the snippet. Keep at most 500 stable event IDs per source. Delete done/dismissed items after 14 days; dismiss open items untouched for 30 days. State writes must preserve unknown future top-level keys.

Stable item IDs:

- Slack: `slack:<channel_id>:<thread_ts-or-message_ts>`
- Email: `email:<threadId>`; use message ID only when no thread ID exists
- Linear: `linear:<issue-key>`

`short_id` is a stable four-to-six character uppercase digest or suffix used in reminders, for example `EC-A1B2`. Re-seeing an event updates the existing item without changing its status or original `created_at`.

## Single-writer lock

After the workday gate passes, atomically create `~/.hermes/state/engagement-checker.lock` as a directory. If it already exists and is less than 10 minutes old, return exactly `[SILENT]`. If it is older than 10 minutes, treat it as a crashed run, remove only that empty lock directory, and retry once. Always remove the lock directory in final cleanup. Do not remove any other path.

The interval is 30 minutes and Hermes cron has a short runtime; the lock is crash protection, not a reason to let a run linger.

## 1. Establish the incremental window

Get the current timestamp with a tool and convert it to Europe/Paris. On first run only, initialize each source cursor to two hours before now. Otherwise use each source's own cursor. Add a five-minute retrieval overlap when the source supports it, but discard any event whose exact timestamp is not newer than the stored cursor unless its stable event ID is absent and it can close or mutate a pending item.

Normal runs must process only new events plus the compact pending queue. Do not rescan the whole day, inbox, Slack history, or all Linear issues for context. When an API has only date-granular search, widen to the cursor's local date, cap the response, then filter locally by exact timestamp before analysis. The widened retrieval is transport overlap, not permission to reprocess old content.

Advance a source cursor to the run's fixed end timestamp only after all of that source's pages and candidate context were successfully processed. On failure, increment its failure count, keep the cursor unchanged, persist the other state changes, and continue. Stable IDs plus each source's bounded `seen` ring make overlap idempotent.

## 2. Collect Slack deltas

Use `mcp__slack__conversations_search_messages` with a bounded date filter derived from the Slack cursor and paginate until `next_cursor` is empty or 100 exact delta events have been retained.

Run two views:

1. `filter_users_from=U08BDJAMSRZ` for commitments, replies, completion signals, and deferral instructions authored by Gaetan.
2. `filter_users_with=U08BDJAMSRZ` for DMs, threads, and messages involving him that may contain a direct ask or someone waiting.

Filter every result by exact Slack `ts` against the cursor before classifying it. Deduplicate the two views by channel and timestamp. For a new candidate only, use `mcp__slack__conversations_replies(channel_id, thread_ts)` to recover enough thread context to decide whether there is a commitment, unanswered ask, resolution, or user instruction. Do not fetch unrelated channel history. Resolve names from source data rather than guessing.

Also inspect new messages in the origin DM with Tars for decisions about existing engagement items. A reply in a reminder thread is authoritative when the parent reminder names one `short_id`; when it contains several items, require the reply to name a short ID, person, or unique topic.

## 3. Collect email deltas

Use the configured Google Workspace script read-only:

```bash
GAPI="python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
$GAPI gmail search "after:YYYY/MM/DD" --max 100
```

Widen from the email cursor's Paris-local date, then filter returned message dates strictly newer than the cursor. Inspect the full body only for new candidate messages where sender/subject/snippet cannot establish whether a response is expected. Use thread IDs to combine sent and received events and to close an item when Gaetan has replied.

Ignore newsletters, automated notifications, receipts, promotions, cold outreach, passive CC/FYI mail, calendar noise, and threads where Gaetan already answered or no response is expected. Never infer urgency from `UNREAD` alone. Never call Gmail send, reply, modify, label, or delete operations.

## 4. Collect Linear deltas

Use the source cursor directly when possible:

```bash
ssh cooper '~/.local/bin/orca linear list-issues --updated-at <ISO_CURSOR> --assignee me --workspace all --limit 100 --json'
```

Filter exact `updatedAt` values newer than the cursor. For a new candidate issue only, inspect bounded context with:

```bash
ssh cooper '~/.local/bin/orca linear issue <ISSUE_KEY> --comments --activity --json'
```

Track direct asks, review/decision requests, issues explicitly blocking someone, security/customer/incident urgency, due dates, and status changes that close loops. An assigned issue is not automatically an engagement reminder; routine backlog and progress updates are noise. Linear coverage is limited to configured workspaces and discoverable assigned/relevant issue updates; do not claim global comment completeness.

## 5. Classify and reconcile

Valid kinds and starting weights:

- explicit promise: 45
- direct unanswered ask: 40
- review or decision: 30
- blocker, security, customer, or incident: 55
- FYI/noise: 0 and never retained as an open item

Never invent a commitment from participation, an emoji, a tentative idea, or a message where somebody else owns the action.

Reconcile new events before scoring:

- Gaetan replied in the Slack/email thread, explicitly completed the action, or Linear reached a terminal state: `done`.
- “Done”, “ignore this”, or an equivalent explicit decision: `done` or `dismissed`.
- “Waiting for Alice”, “delegated to Bob”, or equivalent: `waiting` until a later event returns ownership to Gaetan.
- “Tomorrow instead”, “remind me at 16:00”, or equivalent: `snoozed` with a Paris-local `snooze_until`.

Apply an instruction only when it resolves to exactly one pending item by reminder thread, `short_id`, explicit source link, or unique person-plus-topic match. If it matches none or several, do not mutate state. Store a bounded ambiguity record and mention it only alongside a future substantive reminder; never send a standalone clarification nag.

Relative-time conventions are conservative and always preserve the original phrase with `due_inferred=true`:

- début de matinée → 10:30
- fin de matinée / ce matin → 12:00
- midi → 14:00
- début d'après-midi / début d'aprèm → 15:00
- cet après-midi → 16:30
- fin de journée / EOD → 17:00
- ce soir → 19:00
- demain / next business day → next weekday at 12:00; for a snooze with no time, wake at 10:00
- demain matin → next weekday at 10:30
- semaine prochaine / next week → next Monday at 12:00

Do not silently choose among multiple plausible dates. Explicit calendar dates and times override these conventions. “Next business day” uses the workday gate, not weekdays alone: skip verified metropolitan-France holidays and explicit PayFit leave when calculating the wake date.

## 6. Judge urgency

Re-evaluate only `open` items and snoozes whose wake time has arrived. Compute a private score:

- kind weight above;
- +25 if overdue, or +10 if due within 60 minutes;
- +10 for a demonstrated stakeholder, not title-based guessing;
- +10 when another person is visibly waiting;
- +2 per elapsed Paris work hour since 10:00, capped at +14;
- +15 from 15:45 onward for a non-FYI item due today or with no explicit due date.

Threshold: 60. Context can suppress a mathematically eligible item when evidence shows Gaetan is waiting on someone else, already replied, is on approved leave, or the ask no longer matters. Context can raise a blocker/security/customer/incident item when the evidence is explicit; record the reason.

Cooldowns:

- ordinary items: 120 minutes;
- score 80 or above: 60 minutes;
- blocker/security/customer/incident: 45 minutes.

After cooldown, repeat an ordinary item only when its score rose by at least 8 or the context materially changed. Critical items may repeat after their shorter cooldown while the risk remains explicit. A snooze that wakes may remind once immediately. Do not remind every half hour merely because an item remains open.

At 16:00, 16:30, and the final 17:00 pass, escalate genuinely forgotten direct replies and explicit promises that should not roll overnight. Do not turn the end-of-day rule into an inbox dump.

## 7. Deliver or remain silent

If nothing crosses the threshold, persist source progress, set `last_completed_run`, release the lock, and return exactly:

`[SILENT]`

Otherwise send at most six items, grouped when they share a person or topic. One line each:

`• What — who; why now. Next: one concrete action. [EC-A1B2]`

Include promised/due wording when relevant. Add `+N lower-urgency items retained` only when the cap hid reminder-worthy items. A user may reply in the reminder thread with `tomorrow`, `16:00`, `done`, `ignore`, `waiting for …`, or a short ID.

Mark only delivered items with reminder timestamp and score before final state persistence. Include one `Coverage:` line only when a failed mandatory source materially weakens an item-level judgment and no equivalent evidence closes the gap. Rate-limit identical coverage notices to once every four hours.

On an internal failure, preserve unadvanced source cursors. If no trustworthy reminder can be produced, return `[SILENT]`; never fabricate substitute evidence.

## Scheduling contract

Hermes global timezone must remain `Europe/Paris`. Use two recurring weekday jobs with this skill attached and delivery to the origin conversation:

- `*/30 10-16 * * 1-5` — 10:00, 10:30, …, 16:30
- `0 17 * * 1-5` — the single final 17:00 pass; no 17:30 run

Both jobs use the same self-contained prompt:

> Run `engagement-checker` end to end. Use Europe/Paris and the fixed run timestamp. Apply the same workday gate as `daily-work-brief` first: weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off; return exactly `[SILENT]` when the gate says not to run. Otherwise acquire the single-writer lock; read durable state; collect and exactly filter only Slack, email, and Linear deltas since each source cursor; reconcile user decisions and resolved loops; update each source cursor only after successful processing; evaluate the pending queue with cooldowns; persist state; release the lock; then return the compact reminder or exactly `[SILENT]`. This is a read-only source workflow. Do not spawn or prompt Claude or Orca; Cooper/Orca may only be queried read-only for Linear evidence.

The two jobs share the same state, so do not create separate per-job cursors.