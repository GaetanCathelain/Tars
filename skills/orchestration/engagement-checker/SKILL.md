---
name: engagement-checker
description: "Use for incremental follow-up and commitment reminders."
version: 2.3.0
required_environment_variables: [LINEAR_API_KEY]
metadata:
  hermes:
    tags: [engagement, reminders, slack, email, linear, orchestration]
    category: orchestration
---

# Engagement checker

A half-hourly, source-backed monitor for Gaetan's open loops: explicit commitments he made, direct asks he has not answered, and blockers, reviews, or decisions waiting on him. It is not a daily summary. `daily-work-brief` owns the broad outcome view; this skill processes deltas and re-evaluates only a compact durable pending queue.

Tars owns and executes this workflow directly. Do not spawn, prompt, resume, modify, or delegate through Claude, a worktree/session, or another non-Tars agent system. Outside Tars's own state the permitted writes are **exactly the list in "Permitted writes" below** — that section is the authority, and nothing else in this file, in the cron prompt, or in a reminder restates or extends it.

**Reminders remain paused under SOUL's 2026-08-19 correction.** Do not create,
resume or send engagement reminders unless Gaetan explicitly unpauses them.
Pausing alerts does not stop authorized knowledge filing: preserve tracked
items in Linear and persist material source-backed conversation facts, owner
and next step on the existing ticket, deduplicating comments. Do not delete
knowledge or close issues merely to silence notifications.

All human times use `Europe/Paris`. The scheduling and delivery rules below
apply only after an explicit unpause; a manual knowledge-only run keeps alerts
silent, including routine Coverage notices, and records failures in its task
result/state without pretending the filing succeeded.

**Empty is a value, not a failure — this rule is stated once, here, and every step obeys it.** An empty result set, a zero count, an omitted optional field (`assignee`), an explicit `null` (`completedAt`), and an empty collection (`labels`, `seen`, `items`, the routing index) are all legitimate readings of a call that worked. Only three things are failures: a response carrying `error`/`errors`, a call that could not be made at all, and a read that hit its page cap with `hasNextPage` still `true`. Nothing may block a write, hold a cursor, or claim incompleteness on anything else, and every one of those three must surface in §8's `Coverage:` line.

## Permitted writes

Five, and no others:

1. **The scheduled reminder** to Gaetan (§8), only after explicit unpause; currently disabled by policy.
2. **Create a GCN issue** for a retained open item — `mcp__linear__save_issue` with no `id` (§7).
3. **Close a GCN issue this skill itself filed**, whose item was explicitly resolved outside Linear — `mcp__linear__save_issue` with `id` and `state` only (§5). Never an issue Tars did not file, and never on age alone.
4. **Organize a GCN issue into Todo, In Progress, Waiting for answer, Delayed, or Done** when §5a has direct, source-backed evidence and no conflict — `mcp__linear__save_issue` with `id` and `state` only. Done requires explicit proof of resolution and may apply to a non-terminal issue; the other four targets apply only while the issue is non-terminal. Uncertainty means no write, not a best guess.
5. **Link material source-backed knowledge to its existing GCN issue as a Linear comment** — `mcp__linear__save_comment` with `issueId` and `body`. This applies whenever a filed item's conversation or source reveals a material update: who replied or spoke last and when, whether a reply followed Gaetan's message, a changed owner or waiting party, investigation progress, a decision, a blocker, or a concrete next step. It runs whether or not the item crosses the reminder threshold and whether or not reminders are currently scheduled. Never bury such an update only in Tars's reminder or local state.

For write 5, first read the issue's existing comments with `mcp__linear__list_comments`. If an existing comment already carries the same source event or equivalent update, do not duplicate it. Otherwise create one concise comment with: a heading `Engagement update — <Europe/Paris timestamp>`; the material facts with named people and times; `Next:` with the concrete next step or `ownership unresolved`; and `Source:` with the Slack permalink, email permalink, or Linear source handle. Do not paste raw transcripts or private-message content beyond the minimum operational fact. Parse the create result, then re-read the comments and require an exact matching body on that issue before treating the knowledge as filed. A failed or unverified comment is named in §8's `Coverage:` line and is never reported as saved.

Never a Slack message, email, or reaction from a source integration. Never a write to any team but GCN: a cron run carries no instruction from Gaetan, so `linear-ticketing` §1's company-team escape hatch never applies here. Writes 2–5 obey `linear-ticketing` §13 and are recorded only after its §10 success test passes.

## Load only what is needed

- Run the bundled `scripts/gcn73_helper.py` through `terminal` for Slack delta discovery and every durable-state read/write. It loads the existing protected Slack MCP credentials inside the process, calls only the read-only Slack methods it allowlists, and emits bounded NDJSON instead of filling the agent context with raw pages. Never print or pass either Slack token in argv. Native Slack tools remain the path for the bounded per-candidate thread and delivery-time conversation reads.
- Load `google-workspace` for Gmail. Its existing OAuth is the primary mailbox path. Run `setup.py --check` before Gmail collection; record the bounded auth verdict through the state helper. If it fails, load and check `himalaya` before declaring email unavailable. The email cursor remains held on auth failure, and the first later authenticated, successfully reconciled Gmail run reports recovery once.
- **Linear transport is mixed on purpose. Delta scan: the audited raw collector in §4, coverage-gated. Everything else: native tools.** The collector stays raw because the coverage verdict it emits — complete, truncated, incomplete — is what gates durable cursor advancement, and a tool that paginates internally would make that verdict unknowable. Its scope is exactly the delta scan and the per-candidate bounded context. Every other Linear read, and every write in "Permitted writes", goes through the native `mcp__linear__*` tools per `linear-ticketing`. Do not unify the two; the split is load-bearing.
- The collector queries the fixed endpoint `https://api.linear.app/graphql` with Python stdlib `urllib` and the inherited `LINEAR_API_KEY`. The key must remain inside the process: never print, log, hash, inspect, persist, or place it in argv, and never use `curl` with an expanded authorization header. It reads only the hardcoded query operations in §4, on a path that rejects anything containing `mutation`, and it holds no write path at all.
- Load `linear-ticketing` before any Linear write. It carries the team, state, and label ids, the label and priority rubrics, the closed tool list, the measured request and response shapes, the success test, and the priority-sort rule used for rendering. Where a Linear behaviour is not recorded there, it is unknown — say so rather than assuming it.
- Load `hermes-agent` only when installing or troubleshooting the cron jobs.
- Consult `daily-work-brief` only for its workday-gate semantics. Do not load or run its broad collection workflow during ordinary checks; the objectives and state are separate.

## Workday gate

Apply the same workday decision as `daily-work-brief` before normal delta collection:

1. Today must be Monday–Friday in Europe/Paris.
2. Query the official metropolitan-France holiday calendar at `https://calendrier.api.gouv.fr/jours-feries/metropole/<YEAR>.json` and match today's Paris-local ISO date.
3. Search existing email access for PayFit messages that explicitly cover today. Read the relevant body; a subject or snippet alone is not proof. Approved leave, RTT, or absence is off evidence; a generic PayFit notification is not.

If today is a weekend, metropolitan-France bank holiday, or explicit PayFit day off, make the same bounded actual-work probe as the daily: Gaetan-authored Slack, merged PRs, and local Hermes activity. If there is no substantive work signal, return exactly `[SILENT]` without acquiring the engagement lock or advancing source cursors. If there is substantive work, continue; actual work overrides the nominal day-off gate, exactly as for the daily.

If holiday or leave evidence cannot be checked, continue rather than silently skipping. Mention the coverage failure only if it materially weakens a reminder judgment. The cron schedule itself is weekday-only; the gate also protects manual runs and excludes holidays and leave.

## Durable state

The source of truth is `~/.hermes/state/engagement-checker.json`. It is operational state, not model memory and not git. Create the parent directory if needed. **Never read or rewrite this file through `read_file`, an inline Python snippet, or hand-spliced pages.** Export its compact NDJSON view with `gcn73_helper.py state-export`; read that bounded file in chunks. Apply one compact mutation document with `state-apply --expected-hash <exported sha256>` after every successfully reconciled source. The helper reads JSON directly, preserves unknown top-level keys, trims `seen` rings, writes mode `0600` through temp + fsync + replace, and rereads the exact object before reporting success. A hash mismatch means another writer won: make no write, release the lock, and fail closed.

Linear/GCN is the source of truth for any item that has been filed. `~/.hermes/state/engagement-checker.json` is a working queue, a cache, and an id-map — never the authority for a filed item's status. When the two disagree about a filed item, Linear wins.

Use this bounded shape:

```json
{
  "version": 1,
  "last_completed_run": "ISO timestamp or null",
  "sources": {
    "slack":  {"cursor": "ISO timestamp", "last_success": "ISO or null", "seen": []},
    "email":  {"cursor": "ISO timestamp", "last_success": "ISO or null", "seen": []},
    "linear": {"cursor": "ISO timestamp", "last_success": "ISO or null", "seen": []}
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
- `last_score`, up to 10 reminder records, and up to 10 status-history records;
- `linear_issue`: `{id, priority, closed_at}`, present only once the item has been filed as a GCN issue. `id` is the issue key (`GCN-42`) — **no tool on any path returns an issue UUID** (`linear-ticketing` §9), so nothing keys on one. `priority` is the integer §7b computed, refreshed from a routed candidate's bare int whenever §5 sees one, and read only by §8's bullet sort. `closed_at` is set once the issue is known terminal — **observed** (§5 step 1) or **confirmed** (§5 step 4) — and **cleared by §5's reopen rule**, so it latches a fact that is still true rather than a call that once happened. Three fields, three named consumers: routing (§4), sorting (§8), close-back (§5). Nothing else is stored — a url or a create timestamp would have no reader.

The routing index is the whole mechanism and it is derived from `items{}` at load: every item carrying `linear_issue` contributes `linear_issue.id` → item id. There is no second cache. A `filed{}` map keyed on deleted items existed here and was deleted: every read site's rule was "drop the entry and treat the candidate as fresh", so it changed no behaviour, and the `source == "linear"` items it produced are never filed (§7). A dedupe that must survive local state loss is §7a's scan of Linear itself, which needs no cache.

Never persist raw Slack threads, email bodies, Linear descriptions/comments, credentials, tokens, headers, or private transcript dumps. Redact obvious secret forms; when uncertain, omit the snippet. Keep at most 500 stable event IDs per source. Delete done/dismissed items after 14 days. **One exception: a done/dismissed item whose `linear_issue` has no `closed_at` is kept past 14 days** — §5's close-back retries only while the item survives, and deleting it strands an open GCN issue nobody will ever close. Drop it at 30 days and name the still-open issue once in §8's `Coverage:` line. State writes must preserve unknown future top-level keys.

**The 30-day staleness sweep REPORTS; it never writes.** An open item untouched for 30 days is dropped from the local queue with status-history reason `stale:30d`, and that reason **never triggers §5's close-back**. Going quiet is not resolution: silence is the normal life of a low-priority backlog ticket, and a cron job that Cancels one is cancelling a live human ticket with nothing in the delivery to say so. If such an item carried a `linear_issue`, count it and surface `N filed tickets untouched for 30+ days` in §8's `Coverage:` line — that is the whole of the sweep's output. `linear-ticketing` §1 names the exact close-back and evidence-backed organization exceptions; neither permits a close or cancellation on age or silence.

Stable item IDs:

- Slack: `slack:<channel_id>:<thread_ts-or-message_ts>`
- Email: `email:<threadId>`; use message ID only when no thread ID exists
- Linear: `linear:<issue key>`, e.g. `linear:GCN-42` — the collector reports that key as `issue.identifier`, every native tool reports it as `id`, and it is the same string. **One namespace across the collector and the filer**; a scheme that differed between them would break routing on every filed issue. **Look the index up on `issue.identifier`, never on the collector's prefixed item `id`**: `linear:GCN-42` matches nothing, forever, with no error. The raw `issue.id` UUID exists only inside the collector, where it dedupes the two views within a single run; it never leaves it and never keys durable state.

`short_id` is a stable **six**-character uppercase digest used as the reply-routing handle for an item that is **not yet filed**, for example `EC-A1B2C3`. Six, not four: four collide by birthday across a few hundred live items. It is never sent to Linear, never part of an issue title or description, and never a lookup key — a filed item's identity is its GCN key and nothing else. Re-seeing an event updates the existing item without changing its status or original `created_at`.

## Single-writer lock

After the workday gate passes, atomically create `~/.hermes/state/engagement-checker.lock` as a directory. If it already exists and is less than 10 minutes old, return exactly `[SILENT]`. If it is older than 10 minutes, treat it as a crashed run, remove only that empty lock directory, and retry once. Always remove the lock directory in final cleanup. Do not remove any other path.

The interval is 30 minutes and Hermes cron has a short runtime; the lock is crash protection, not a reason to let a run linger.

## 1. Establish the incremental window

Get the current timestamp with a tool and convert it to Europe/Paris. **First run means `last_completed_run` is null** — then, and only then, initialize each source cursor to two hours before now. Otherwise use each source's own cursor. Add a five-minute retrieval overlap when the source supports it, but discard any event whose exact timestamp is not newer than the stored cursor unless its stable event ID is absent and it can close or mutate a pending item.

Normal runs must process only new events plus the compact pending queue. Do not rescan the whole day, inbox, Slack history, or all Linear issues for context; the single exception is §2's bounded delivery-time read. When an API has only date-granular search, widen to the cursor's local date, cap the response, then filter locally by exact timestamp before analysis. The widened retrieval is transport overlap, not permission to reprocess old content.

Advance a source cursor to the run's fixed end timestamp only after all of that source's pages and candidate context were successfully processed. **On success, set that source's `last_success` to the run's fixed end timestamp. On failure, leave `last_success` unchanged, keep the cursor unchanged, persist the other state changes, and continue.** That single write is the only one the field has, and §8's mandatory "source failed this run" coverage condition is exactly the test `last_success < this run's end`. A `failures` counter lived here and was deleted: nothing ever read it, and a field named `failures` sitting in the state shape reads as retry or backoff state that does not exist. Add one only together with its consumer.

**Append every processed event's stable ID to that source's `seen` ring**, oldest evicted past 500. Stable IDs plus each source's bounded `seen` ring are what make the 5-minute retrieval overlap idempotent; §1's "unless its stable event ID is absent" test reads that ring and nothing else writes it.

## 2. Collect Slack deltas

Run the audited helper, writing its bounded NDJSON to a mode-0600 scratch file:

```bash
HELPER="${HERMES_HOME:-$HOME/.hermes}/skills/orchestration/engagement-checker/scripts/gcn73_helper.py"
python3 "$HELPER" slack --cursor '<stored Slack cursor>' --run-end '<fixed run end>' --output "${HERMES_HOME:-$HOME/.hermes}/state/engagement-slack-delta.ndjson"
```

The helper calls Slack's fixed Web API endpoint with the protected `xoxc`/`xoxd` values loaded inside the process, never accepted on argv or emitted. It runs the `from:<@U08BDJAMSRZ>` and `with:<@U08BDJAMSRZ>` views for each Paris date, exhausts cursor or legacy page pagination, exact-filters `stored_cursor < ts <= run_end`, removes bots and Tars, deduplicates across views, sorts oldest-first, and returns at most the oldest 100 exact events including edge ties. It emits a `proposed_cursor`: the edge timestamp when more events remain, otherwise the fixed run end. A page cap, repeated cursor, auth/HTTP/API error, malformed response, redirect, or missing completeness verdict exits nonzero with `complete:false`, `incomplete:true`, `fail_closed:true`, no events, and no advancing cursor.

Read the NDJSON scratch file in bounded line chunks. Line 1 is coverage; each later line contains one minimal event. Do not reconstruct raw Slack pages in model context. Only after every retained event and required candidate thread read is classified and its item mutation is ready may the proposed cursor be included in a `state-apply` mutation with all processed stable IDs in `append_seen.slack`. Delete the scratch file after that confirmed state write; on any failure keep the source cursor unchanged and remove the scratch file without using its partial contents.

For a new candidate only, use `mcp__slack__conversations_replies(channel_id, thread_ts)` to recover enough thread context to decide whether there is a commitment, unanswered ask, resolution, or user instruction. Do not fetch unrelated channel history — with one exception: **one bounded** `conversations_history` read, per person named in an item **this run is delivering**, of Gaetan's conversation with them (DM or channel), and only when the delivery owes conversation state (SOUL rule 10). Resolve names from source data rather than guessing.

**Unthreaded DM handoffs are one conversational unit, not independent messages.** When a new unthreaded DM turn offers a choice about who will act (for example, “do you want to do X or shall I?”) and the next human reply selects Gaetan (“do it please”), classify the pair together as a direct ask owned by Gaetan. Use only the bounded adjacent DM turns already returned by the delta scan, widening with one bounded `conversations_history` read for that DM only when the handoff cannot otherwise be resolved. Do not append either event to `seen` until the pair has been classified and any retained item has been persisted; marking both individual rows seen while dropping the combined handoff permanently loses the ask.

Also inspect new messages in the configured Slack reporting conversation for decisions about existing engagement items. A reply in a reminder thread is authoritative when the parent reminder names one `short_id`; when it contains several items, require the reply to name a short ID, person, or unique topic.

## 3. Collect email deltas

Use the configured Google Workspace scripts read-only. Check auth first:

```bash
GSETUP="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/setup.py"
GAPI="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
$GSETUP --check
$GAPI gmail search "after:YYYY/MM/DD" --max 100
```

`AUTHENTICATED` with exit 0 is the only successful auth verdict. On `invalid_grant`, `REFRESH_FAILED`, `NOT_AUTHENTICATED`, a nonzero exit, or an unrecognized response: make no Gmail collection call, hold the email cursor and `last_success`, and persist `auth.email={"ok":false,"at":"<run end>"}` through `state-apply`. Never persist stderr or OAuth details. On the first later authenticated run, collect from the still-held cursor; only after the whole Gmail window reconciles successfully persist `auth.email={"ok":true,"at":"<run end>"}` together with the advancing email cursor. When the helper returns `recovered_sources:["email"]`, include `Gmail authentication recovered; no email interval was skipped` once in Coverage. A successful auth probe without a successful collection is not recovery and does not advance anything.

Widen from the email cursor's Paris-local date, then filter returned message dates strictly newer than the cursor. Inspect the full body only for new candidate messages where sender/subject/snippet cannot establish whether a response is expected. Use thread IDs to combine sent and received events and to close an item when Gaetan has replied.

Ignore newsletters, automated notifications, receipts, promotions, cold outreach, passive CC/FYI mail, calendar noise, and threads where Gaetan already answered or no response is expected. Never infer urgency from `UNREAD` alone. Never call Gmail send, reply, modify, label, or delete operations.

## 4. Collect Linear deltas

Require `LINEAR_API_KEY` in the collector process environment without reading or displaying its value. Set `LINEAR_CURSOR` to the stored source cursor and `LINEAR_RUN_END` to the fixed run end.

Run the bundled audited collector through `terminal`, never `execute_code`:

```bash
LINEAR_CURSOR='…' LINEAR_RUN_END='…' python3 "${HERMES_HOME:-$HOME/.hermes}/skills/orchestration/engagement-checker/scripts/linear_collector.py"
```

Use that exact installed script; do not materialize, rewrite, inline, pipe, or delete a temporary copy. The terminal path is required because `execute_code` scrubs environment variables containing `KEY`. The command keeps the credential out of argv and output.

The collector is the read-only, coverage-gated exception to native Linear transport. It contains only hardcoded GraphQL `query` operations, rejects `mutation` at the shared network choke point and name-resolution path, disables redirects, pins the response URL, caps bytes before JSON parsing, bounds pages/candidates/context, and redacts secret forms before clipping output. It has no write path. All permitted Linear writes remain native `mcp__linear__save_issue` calls governed by `linear-ticketing`.

The read pass is transient and read-only: do not persist its full output. Candidate context is intentionally bounded to 40 comments and 40 activity records per issue, with `source_has_more` making truncation explicit. Treat a nonzero exit, HTTP error, GraphQL `errors`, malformed response, stalled pagination, or a completeness flag that is missing **or false** as a Linear source failure, and keep the Linear cursor unchanged in all such cases. **An empty `items` array is not one of them**: a window in which no issue was touched exits 0 with `exact_candidates: 0`, `truncated: false` and an advancing cursor. That is a complete, quiet window — zero rows never means the read failed.

**A cap overflow truncates instead of raising, and the two truncations behave differently because they know different things.** Both report `truncated`, `candidates_dropped` and `issue_connection_exhausted`, and both are named in `Coverage:` per §8.

- **Candidate-budget overflow** (the shared `MAX_CANDIDATES` pool, deliberately shared across both delta views, not per view): the window *was* fetched whole, so it can be ordered. Candidates are sorted by `updatedAt` **ascending** and the oldest budget-worth are retained, ties with the edge event included. That batch is every event between the cursor and the edge, contiguous and complete, so the cursor advances to the edge and skips nothing; everything newer is re-fetched next run. The backlog therefore **drains one bounded batch per run** instead of the source wedging while the window widens forever.
- **Page-cap truncation** (`MAX_ISSUE_PAGES`): the window was *not* fetched whole and the connection carries no `orderBy`, so the fetched subset is server-ordered and the true oldest events may be unseen — no client-side ordering can rescue it. `drain()` therefore applies `secure-delta-collectors`' prescription verbatim: **when a wide interval hits a cap, split it into deterministic adjacent subwindows and require every subwindow to complete before advancing the aggregate cursor.** The interval `(cursor − 5 min, run_end]` is cut into `SUBWINDOWS` equal adjacent parts, scanned oldest-first with a closed `lte` upper bound, and the view is complete through the end of the last subwindow that finished under the page cap. The aggregate cursor advances to the **minimum** `complete_through` of the two views, so it moves only over ground both proved whole. A wide window therefore **drains** across runs instead of wedging, and nothing is skipped.

Bounds on that path. The subwindow queries (`AssignedIssuesWindow`, `TeamIssuesWindow`) are the unbounded ones plus `lte: $until` on the same measured `updatedAt` comparator — **the `lte` form itself is not yet measured; probe it in T7**. They run only after a cap was already hit, so if `lte` were rejected the collector fails closed exactly where it would have held the cursor anyway. The residual ceiling is a subwindow that *still* caps — more than `ISSUE_PAGE * MAX_ISSUE_PAGES` issues touched inside one sixteenth of the window, per view: then `complete_through` stops at the last good boundary, `truncated` is reported, and the next run resumes there. Report it in `Coverage:` per §8 rather than letting it degrade quietly, and treat a persistent one as worth its own GCN ticket (raise the caps), the way GCN-10 carries the tool-surface deferral. Advance the cursor to `proposed_cursor` once every candidate's required bounded context calls have succeeded, classification and reconciliation are complete, and state persistence succeeds; when `cursor_advances` is false, `proposed_cursor` is the unchanged cursor and that write is a no-op by construction.

Track direct asks, review/decision requests, issues explicitly blocking someone, security/customer/incident urgency, due dates, and status changes that close loops. An assigned issue is not automatically an engagement reminder; routine backlog and progress updates are noise. Coverage includes every GCN issue touched in the window plus cross-team issues assigned to the authenticated viewer, deduplicated by issue id, and the explicitly bounded context reported above; never claim global issue, comment, or activity completeness. `TeamIssues` exists so a GCN issue Gaetan creates by hand, or reassigns away, still pulls; `AssignedIssues` keeps the cross-team coverage.

### Filed-issue routing — the nag-loop guard

Build the routing index at state load from `items{}`, and from nothing else: every item carrying `linear_issue` contributes `linear_issue.id` → item id. It keys on the issue **key** (`GCN-42`), which the collector reports as `issue.identifier` and every native tool reports as `id`: one namespace on both sides of the skill. Look up `issue.identifier`, not the collector's `linear:`-prefixed item id. Keying on a UUID is not an option — no tool returns one.

A Linear candidate whose key is in that index never becomes a new item. Route it to the parent item and apply every state change there. A candidate that routes to nothing — no item carries that key — is simply a fresh candidate; the item it becomes has `source` `linear`, which §7 never files, so that path cannot loop.

Termination argument, edge by edge:

- **Item → issue, at most once.** §7's filing condition requires `linear_issue` to be absent; §7c records it and flushes to disk before any further call. If the response is lost or the run dies before that flush, §7a's next scan finds the item's own stable id at the head of the existing issue's description and adopts it instead of creating a second issue — and if that scan cannot be proven complete, nothing is created at all. Known ceiling: an issue a human moves off GCN leaves the GCN-scoped scan's view, so its item can re-file once. Bounded, human-initiated, and preferable to a key that does not exist.
- **Issue → item, closed.** A filed issue routes back to its parent through the index above while the item exists. Once the item is deleted the route is gone by design, and the issue can become a fresh item — whose `source` is `linear`, which §7 never files, so the edge cannot cycle.
- **Reminder → item, closed.** §2 discards every Slack message carrying a `bot_id` or authored by Tars, and §7 refuses any item whose snippet reproduces Tars's own bullet shape. Tars's own reminders cannot re-enter as candidates, so their fresh timestamps have nothing to feed.

## 5. Classify and reconcile

Valid kinds and starting weights:

- explicit promise: 45
- direct unanswered ask: 40
- review or decision: 30
- blocker, security, customer, or incident: 55
- FYI/noise: 0 and never retained as an open item

Never invent a commitment from participation, an emoji, a tentative idea, or a message where somebody else owns the action.

Reconcile new events before scoring:

- Gaetan replied in the Slack/email thread or explicitly completed the action: `done`. A Linear terminal state resolves through the terminal set in the next bullet and never through this one — that set maps each type to its own status, and a blanket `done` here would record a Canceled or Duplicate issue as work Gaetan finished.
- A candidate reporting a terminal `state.type` for an issue an item already carries — routed to it through `linear_issue`, or matching an item whose own id is `linear:<key>`. §8 keys its bullet handle on those same two shapes, a Linear identity rather than Tars having filed it, and reconcile keys on them for the same reason. **The terminal set is these three types, stated here once and referenced everywhere else**: `completed` → parent item `status = "done"`, status-history reason `linear:completed`; `canceled` → `status = "dismissed"`, reason `linear:canceled`; `duplicate` → `status = "dismissed"`, reason `linear:duplicate`. Closing the issue in Linear is how Gaetan clears the item, and Duplicate is one of the ways he closes one: the work continues under the canonical issue, not this one. The remainder — `triage`, `backlog`, `unstarted`, `started` — is the whole non-terminal set: three terminal types, four live ones, the entire `statusType` enum. `triage` is not configured on GCN, which is why §7a covers the whole non-terminal board in three calls. So the set is closed and an eighth type is an edit to this file, never a silent fall-through to "not terminal". **Key on the type enum, never on the status name**: a workspace renames a status freely, and GCN's names agreeing with their types today guarantees nothing about tomorrow or about another team. **Never read terminality off `canceledAt`**: Linear stamps that field for the whole cancel-like group, so GCN-25 carries `canceledAt` while its type is `duplicate` — the field neither implies `canceled` nor is required by it. `state.type` is the **collector's** raw GraphQL shape and is correct here; a native read has no `state` object at all — it carries flat `status` and `statusType` (`linear-ticketing` §9). Same enum, two shapes.
- A routed filed issue in a **non-terminal** state whose parent is `done` or `dismissed` with a `linear:*` reason: return the parent to `open` with a fresh cooldown, **and delete `linear_issue.closed_at` in the same step**. `closed_at` is what suppresses the close-back; leaving it set on a reopened item means the loop can never be closed again and the issue becomes a permanently open ghost. A latch with no clearing step is a dead mechanism — this is its clearing step. Reopening an issue must make it visible again, not bury it until the 14-day sweep.
- A routed filed issue, any state: **refresh `linear_issue.priority` from the candidate's `issue["priority"]`** — a bare int on the collector path (`linear-ticketing` §2). Without it §8 sorts filed bullets forever on Tars's original guess and disagrees with the daily board about the same issue. This is the only consumer of the collector's `priority` field.
- A routed candidate matching no item: treat it as a fresh candidate.
- An assignee change away from the viewer, read from the history view's `fromAssignee`/`toAssignee`: `waiting`. The team view keeps pulling the issue, which is correct, but Tars does not nag about work Alice now owns.
- “Done”, “ignore this”, or an equivalent explicit decision: `done` or `dismissed`.
- “Waiting for Alice”, “delegated to Bob”, or equivalent: `waiting` until a later event returns ownership to Gaetan.
- “Tomorrow instead”, “remind me at 16:00”, or equivalent: `snoozed` with a Paris-local `snooze_until`.

### 5a. Organize GCN issues from evidence

This is permitted write 4. It applies only to GCN issues already seen in the Linear delta or already present in the compact pending queue. It never scans or rewrites the whole board merely to make it look tidy.

Resolve the GCN state IDs at the start of a run with the native `mcp__linear__list_issue_statuses` read. Require exactly one status for each exact name: `Todo`, `In Progress`, `Waiting for answer`, `Delayed`, `Done`. If any name is absent or duplicated, organize nothing and report the coverage failure. Never infer an ID from memory, a previous run, or a similarly named status.

For each candidate, build a bounded evidence packet from the issue description/comments/history, the source Slack thread or DM, the Gmail thread when the item is email-backed or mail is explicitly referenced, and recent read-only Cooper Claude/Orca evidence when it can show whether investigation actually began. Reuse the bounded Cooper inspection semantics in `daily-work-brief` §3: inspect current Orca session/worktree state and relevant Claude transcript turns, never spawn, prompt, resume, or modify a session. Absence of evidence is unknown, not evidence of inactivity.

Map only these direct signals:

- `In Progress`: Gaetan explicitly says he started/is working on it, or a matching Claude/Orca session contains substantive investigation or implementation activity for this issue. Merely opening a session, mentioning the topic, browsing, or having an old branch is insufficient.
- `Waiting for answer`: a named other person owns the next response/decision and Gaetan has already sent the ask or handoff. A draft, an intended ask, or “I should ask” is insufficient.
- `Delayed`: Gaetan explicitly defers/snoozes the work to a later time/date or says it cannot proceed until a later condition. Generic low priority, silence, or no recent activity is insufficient.
- `Todo`: the issue is live and owned by Gaetan, but use this only when direct evidence shows the prior blocking/waiting/delay condition ended or Gaetan explicitly puts it back in the queue. Never downgrade `In Progress`, `Waiting for answer`, or `Delayed` to `Todo` merely because their positive evidence is absent from the current bounded window.
- `Done`: Gaetan explicitly says the task is done/resolved, or source evidence proves the requested outcome landed. A reply alone, stopped activity, elapsed time, a merged but unrelated change, or a person no longer waiting is insufficient.

Evidence precedence is explicit newer user direction, then source-backed newer Slack/email/Linear/Claude-Orca activity. An already-terminal Linear state always wins and is never rewritten by organization. When two states remain plausible, evidence conflicts, identity/topic matching is not unique, a required source failed, or the evidence packet is incomplete in a way that could change the verdict: **make no state write**. Keep the current Linear state and record the ambiguity; do not choose the safer-looking status.

Write only when the target differs from the current flat `status`, using one native `mcp__linear__save_issue` call with `id` and resolved `state` only. Never pass labels, priority, assignee, description, or relations. Parse the result and require both returned `status` to equal the exact target name and returned `statusType` to match the resolved status's type. Then make one `mcp__linear__get_issue` read and verify the same `id`, GCN `teamId`, `status`, and `statusType` before recording the organization as done. An unconfirmed or silently no-op write is a failure and is reported under §8; never retry it as another status in the same run.

Close in the item → Linear direction too, or every loop Gaetan resolves in Slack or email leaves a permanently open ghost issue on the board — and the daily brief's board renders those ghosts ahead of real work. This is permitted write 3. It runs for any item carrying `linear_issue` without `closed_at` whose status is `done` or `dismissed` **by an explicit resolution — a user instruction, a reply that closed the thread, or a terminal Linear state — and never by the 30-day staleness sweep**, including one that reached that status on an earlier run. Check the status-history reason, not the bare status: a `stale:30d` dismissal is Tars losing sight of the loop, not Gaetan resolving it, and closing on it would Cancel a live human ticket silently.

1. **Skip the write when the issue is already terminal.** If this run's routed candidate reports a `state.type` in the terminal set above — that is how the item reached `done`/`dismissed` in the first place — set `linear_issue.closed_at` from that observation and make no call. Writing Done to an already-Done issue only bumps `updatedAt` and pulls it back into the next window for nothing.
2. **Confirm the issue is still on GCN, and still not terminal.** Read it with `mcp__linear__get_issue`, `id` = `linear_issue.id`, and check `teamId` is `81e7b769-2a46-4e2a-8db5-c165a7963b0e`. **If it is not GCN, or the read fails, do not write.** Record the fact on the item and surface one line in §8 naming the issue and the team it now sits on. A company-team write needs Gaetan's per-message instruction; a cron run has none, and "it is only a cleanup" is not an exception. **Then test that same payload's flat `statusType` against the terminal set — zero extra calls, the read is already in hand.** If it is terminal the issue is already resolved: set `linear_issue.closed_at` from that read, make no write, and never report it as a close. This backstop exists because the reconcile above is event-driven only — it fires for a routed candidate inside the run's cursor window and nowhere else — so an issue that goes terminal while the cursor is held is never routed, step 1 never sees it, and step 3 overwrites a resolution a human made. It reaches only the items the close-back already reaches: an item still `open`, `waiting` or `snoozed` is not read back anywhere, and reconciles on the first successful collector run instead. Concretely: Gaetan marks the issue Duplicate, no run routes it, step 3 writes Done over that, and step 4's assertion passes on the state it has just written and reports the close as confirmed. A wrong write that certifies itself is the one failure this list exists to prevent.
3. On GCN, call `mcp__linear__save_issue` with **`id` and `state` only** — `state` `0434e579-7b85-487a-8cf9-5aed6caaf41b` (Done) for `done`, `77aad3b3-deac-49a7-a39e-1bea02d93820` (Canceled) for `dismissed`. Never pass `labels`, not even empty: it replaces the whole set, and omitting it is measured to preserve it.
4. **Confirm the move on the returned payload**: parse the `result` string and assert `statusType` is `completed` (or `canceled`). An unresolvable `state` returns a normal payload with the old state and changes nothing — "no error" is not success (`linear-ticketing` §10). Only a confirmed move sets `linear_issue.closed_at`.

An unconfirmed close fails open: the item keeps its local status, `closed_at` stays absent, the attempt is retried on the next run while the item survives, and it is named in `Coverage:` — never reported to Gaetan as closed.

Apply an instruction only when it resolves to exactly one pending item by reminder thread, `short_id`, GCN key, explicit source link, or unique person-plus-topic match. If it matches none or several, do not mutate state. Store a bounded ambiguity record and mention it only alongside a future substantive reminder; never send a standalone clarification nag.

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

## 7. File retained items as GCN issues

**This runs after §6 scoring and before §8 delivery, in that order, every run** — the priority below is computed from this run's classified facts, never from `last_score`, which §6 has only just produced and only for delivered items. It is a native tool call the agent makes — `mcp__linear__save_issue` — not code; §4's collector has no write path. Load `linear-ticketing` first: its §13 `linear_write` contract and §10 success test are what this section obeys.

File an item that is retained with `status == "open"`, whose `source` is not `linear`, and which has no `linear_issue`. Refuse to file an item whose snippet reproduces **Tars's own reminder bullet shape** — a leading `• `, a ` Next: ` clause, and a trailing `[EC-…]` or `[GCN-…]` handle. That shape is Tars's own output, not a loop, and this is the structural backstop behind §2's sender filter. A bare mention of a ticket key is not a match: "can you look at GCN-42 today?" is a genuine ask and must still file.

**File at most 10 issues per run.** Order the eligible items by the §7b `priority` integer ascending — most urgent first — ties broken by oldest `created_at`, file the first 10, and file **none** of the rest this cycle. §2 retains up to 100 Slack events and §3 pulls up to 100 mails, so a single run after a multi-day outage can classify ~200 candidates and nothing else bounds the creates; without a cap that lands as a hundred issues on Gaetan's board and in the daily board in one cron run, with nothing in the delivery to say so. A deferred item is left **untouched**: §7c writes `linear_issue` only for an issue actually created or adopted, so §7's filing condition still selects a deferred item next run and the queue drains 10 per run with no new state and no extra dedupe risk — §7a's scan runs before every create regardless. Name the deferral in §8's `Coverage:` line as `N items awaiting filing`, on the ordinary rate-limited path — a self-healing, non-lossy backlog is not a must-never-hide state, and §8's rate-limit paragraph says why.

### 7a. Dedupe scan — first, every time, before any create

The dedupe key is the item's own stable id — `slack:<channel>:<ts>` or `email:<threadId>` — carried as the **first line of the issue description** (§7b). No synthetic namespace and no title tag: the issue's identity is its GCN key, and the tie back to the item is the source handle we already hold.

Call `mcp__linear__list_issues` with `team` `81e7b769-2a46-4e2a-8db5-c165a7963b0e` (GCN), `limit: 250`, `fields: ["id", "title", "description", "statusType", "priority"]`, **one call per non-terminal type**: `state` `"started"`, then `"unstarted"`, then `"backlog"`. Union the three and deduplicate on `id`. Exactly as `daily-work-brief` read 1, and for the same reason: those three types are the whole non-terminal set on GCN (`linear-ticketing` §2 lists all seven states), so scoping by type keeps every Done, Canceled and Duplicate issue — including every issue this skill ever filed and closed — out of the page budget entirely, and the scan does not start failing its own completeness test the year GCN crosses 2000 lifetime issues. **Do not read the team unfiltered and drop terminal rows client-side**: step 3 discards them anyway, and they are paid for in the one currency this scan cannot afford, its own completeness budget. This is not the forbidden narrowing of step 1 — a `state`-type-scoped call is a different read, deliberately (`linear-ticketing` §8), and it is the read this scan is *for*, since a terminal hit was never an answer.

All three type values are the same measured enum. **A successful call that returns zero rows is a complete, empty view — it is never evidence that the read failed, and it never blocks a create.** GCN's Backlog column is legitimately empty and returns `{"issues":[],"hasNextPage":false}`; the earlier rule here — "backlog returned zero while the other two returned rows, so treat the scan as unproven" — was a permanent stop on this board, and it filed nothing for three consecutive cycles with two eligible items queued. **Never infer failure from a row count, and never compare one view's row count against another's.** Completeness is proven per call, by `hasNextPage == false` inside the page cap (step 1), and by nothing else. The real silent-zero trap is a malformed filter, not an empty column: **never build a negated (`"!Done"`) or comma-list (`"Todo,In Progress"`) `state` filter, and never pass `state` an array**: an array is a hard error, and the other two return a normal response carrying zero issues — a scan that finds no duplicate because it looked at nothing.

All five `fields` are measured valid and returned on every row; `identifier` and `state` are the two an implementer reaches for by reflex and both are hard errors that kill the whole call and return no data (`linear-ticketing` §8). Project deliberately: an unprojected large read blows Hermes' own tool-result budget and never reaches the agent. `statusType` and `priority` are not decoration — step 3 asserts the first and §7c stores the second.

Then, in order:

1. **Page each of the three calls until `hasNextPage == false`.** It is the only completeness signal that exists — there is no total count and no truncation flag, and counting rows against `limit` is not the test. While it is `true`, re-issue the **identical** call with `cursor` set to the returned value and union the pages, up to an explicit cap of **8 pages per call** (`linear-ticketing` §8). **Never narrow the filter to make it go false**: an `updatedAt` bound changes the question, so the retry's `false` proves a *different* scan complete while the issues outside the bound — exactly the older duplicates this scan exists to find — are never looked at. Wrong and self-certifying at once.
2. **Exactly three things block filing, and each is a genuine failure**: a call that returned an `error` payload, a call that could not be made at all, or a call that hit the 8-page cap with `hasNextPage` still `true`. On any of them, file nothing this cycle and say so in §8's `Coverage:` line, naming which view failed and which of the three it was. **Nothing else blocks.** Three calls that each ended `hasNextPage == false` prove the scan complete *even when all three returned zero rows*: an empty non-terminal board is a board with nothing to dedupe against, so filing proceeds normally. Creating on top of an unproven scan is how one lost response mints a duplicate issue every 30 minutes indefinitely; refusing to create on top of a proven-empty one is how a board that is merely quiet never receives its first issue.
3. Scan the returned `description` values for the item's stable id as an **exact substring**. Read the field defensively: it may be absent, `null`, or `""` depending on the path, and it is truncated at ~400 characters, which the head placement exists to survive. A hit counts **only when that issue is non-terminal**; the scan can no longer return anything else, so keep `statusType` and assert it — `statusType` in `completed`, `canceled`, `duplicate` on a row this read returned means the type scoping is not doing what it says, and the scan is not to be trusted for dedupe. A terminal issue is a *closed past loop on a re-used thread or mail thread*, not this loop, and it is correctly invisible here: a new loop on a re-used thread finds no live hit and files a new issue, rather than being silently swallowed by an old Done ticket. On a live hit the item is already filed: **do not create** — adopt that issue's `id` and its priority normalised to an int, and record them through §7c, which is the record path for an adoption as much as for a create.

`query` also searches description text, but it is relevance-ranked and fuzzy — it returns issues that do not contain the string, and `hasNextPage` is always `true` in query mode, so it carries no completeness guarantee. Use it as an optimisation over the scan above if at all; never as the test, and never read an empty `query` result as proof of absence.

### 7b. The create — one call

**One `mcp__linear__save_issue` call with no `id`**, carrying every field below together. Never create then correct.

**Email-source deep link.** For every item whose `source == "email"`, derive a Gmail permalink from its stable id: `email:<threadId>` becomes `https://mail.google.com/mail/u/0/#all/<threadId>`. The Gmail thread id is an opaque identifier, not a credential. Prefer the thread id already used by the stable item id; only when Gmail supplied no thread id may the stable message id occupy that position. On the same create call, pass `links: [{"url": "<derived Gmail permalink>", "title": "Source email — <item title>"}]`, with the title trimmed to 255 characters. This is part of the atomic create, not a later correction. For non-email items, omit `links`. Also store this direct permalink as the email item's `evidence_handle` when first classifying it, so the `Evidence:` line below is clickable; legacy `gmail:<id>` handles may remain in old state, but the create still derives the direct link from the stable item id. A confirmed email-source create must additionally return an `attachments` entry whose `url` and `title` exactly match the values sent; if it does not, apply §7c's one confirming `get_issue`, and if the attachment is still absent treat the create as unconfirmed.

- `description`, in this order, real newlines, never a literal `\n`:

  ```
  Source: <the item's stable id>
  Evidence: <evidence handle>

  <sanitized snippet>
  ```

  The `Source:` line is §7a's dedupe key and **must stay on line 1**. `list_issues` truncates descriptions at ~400 characters, so the old convention of a provenance line at the foot is cut away and invisible to any scan.
- `title`: the item title, trimmed to 255 characters. No tag, no short_id, no synthetic id.
- `team` `81e7b769-2a46-4e2a-8db5-c165a7963b0e` (GCN); `assignee` `4951b192-e49c-4b7e-b491-58c89e66043c` (Gaetan).
- `state` `59ed732a-f242-4eba-926f-1c0d128fe83c` (Todo). **Never omit it. Linear's silent default is Backlog, not Todo**, the issue stays there permanently, and the failure surfaces as a wrong column rather than an error.
- `labels`: exactly one id, from the table below. `labels` replaces the whole set — harmless on a create, load-bearing on §5's close update.
- `priority`: a bare integer 0–4, from the table below. Never 0, never omitted.

| when — the wording arm is checked first and wins over the kind arms | `labels` | id |
|---|---|---|
| access, credential, or permission wording | `access` | `a4fd96c1-ee2d-4c03-8461-f8a15cca372a` |
| kind: explicit promise | `fix` | `bd21ce6f-2ff9-4850-9f81-d54fc48524f0` |
| kind: blocker, security, customer, or incident | `fix` | `bd21ce6f-2ff9-4850-9f81-d54fc48524f0` |
| kind: review or decision | `report` | `bb8da686-e6f7-4691-a030-e11f98c3f8c6` |
| kind: direct unanswered ask, and the default | `investigation` | `a0973525-a091-412d-abf1-aed30bedf60a` |

Precedence is what makes "exactly one" true: "I'm blocked, I need AWS credentials" is kind `blocker` and access wording at once, and it files as `access`. The full label table is in `linear-ticketing` §2.

`priority` comes from **time-invariant facts only** — the item's kind and its due-date state, read from the item, never from `last_score`, which drifts with the clock by design and would file the same item P4 at 10:00 and P1 at 16:00 with nothing ever re-pricing it:

- blocker, security, customer, or incident → 1 if overdue or due today, else 2;
- explicit promise → 2 with a due date, else 3;
- direct unanswered ask → 3, raised to 2 when someone is visibly waiting;
- review or decision → 3;
- anything else → 4.

### 7c. Confirm, record, flush — the only place `linear_issue` is written

`save_issue` returns one key, `result`, holding a **JSON string**: parse it. There is no success flag, and "no error" is not success — an unresolvable field returns a normal, complete, `result`-bearing payload and silently changes nothing (`linear-ticketing` §10). There is **no `identifier` field on any response** — never test for one; a test requiring both an identifier and an id can never pass, and would classify every successful create as a failure.

Confirm the **intent**, not just that an issue exists. Identity plus all four mandatory fields, each against the representation the response uses:

- `id` matches `^[A-Z]+-\d+$` and `teamId` is `81e7b769-2a46-4e2a-8db5-c165a7963b0e`;
- `statusType` is `unstarted` (`status` `"Todo"`) — a `state` that fails to resolve is dropped silently and the issue sits in Backlog forever, which surfaces as a wrong column and never as an error;
- the `priority` **object**'s `value` equals the integer sent;
- the `labels` array of **names** contains the name of the label id sent (`linear-ticketing` §2 maps id → name);
- `assigneeId` equals `4951b192-e49c-4b7e-b491-58c89e66043c` — `assignee` is a display name and never matches the uuid.

If any of them cannot be compared on the response, do **one** `mcp__linear__get_issue` on the returned `id` and assert there; never skip an assertion.

On confirmation — or on an adoption from §7a, which needs no confirmation because it created nothing — in one step, before issuing any further call:

1. set `items[<item id>].linear_issue = {id, priority}` — `id` is the GCN key straight off the response, `priority` the integer §7b computed (or, for an adoption, the adopted issue's priority normalised to an int). §8 sorts on that stored integer; nothing else ever reads a filed issue's priority back;
2. **flush state to disk.**

That is the one named place it is written. Never batch the flush to the end of the run: the crash window between a confirmed create and the flush is precisely what §7a's scan absorbs, and it can only absorb one.

Anything short of confirmation is a creation **failure**: fail open, leave the item unfiled, log a bounded coverage note, retry next cycle — which §7a makes safe. Never record `linear_issue` from an unconfirmed response and never report the issue to Gaetan as filed. Reads still fail closed.

**Provenance is a rule about the value, not about the mechanism that writes it, and it binds every site that records one — §5's `closed_at` latch and its `priority` refresh included, not only this section.** Every value in `linear_issue` that Linear owns — the `id` always, a `priority` refreshed from a candidate, a latched `closed_at` — is copied from a Linear tool response this run actually received and still holds: the confirmed `save_issue` payload above, the one confirming `get_issue`, a §7a adoption's read, §5 step 2's team-and-terminality read, a §5 step 4 confirmed close, or a routed collector candidate. (A create's starting `priority` is the integer §7b computed, exactly as the checklist above says; it is Tars's own judgment, not a Linear reading, and the id is the field a wrong value silently mis-routes.) An id that is remembered, inferred from the conversation, carried over from an earlier run's transcript, or typed as a literal into code is **not a source**, and writing state through `execute_code` is no exemption — a dict assembled by hand is the same write with the section heading removed. Measured: a run persisted `linear_issue: {"id": "GCN-26", …}` as a hardcoded Python literal inside an `execute_code` call, in a cycle that made zero Linear tool calls, from the same script whose `last_failure_notice.linear` recorded Linear as unreachable. If Linear was unreachable or the call failed this run, `linear_issue` stays absent and the item stays unfiled — the fail-open path in the paragraph above, unchanged. A believed-but-unsourced key is at most a note on the item and never a field of `linear_issue`: §4's routing index and §8's bullet handle are both keyed on `linear_issue.id`, so a wrong key mis-routes both, silently and for as long as the item lives.

## 8. Deliver or remain silent

If nothing crosses the threshold, persist source progress, set `last_completed_run`, release the lock, and return exactly:

`[SILENT]`

The one exception is the coverage break below: any of the seven must-never-hide states is reported even on an otherwise silent run. A silent run owes no explanation beyond that; when one is given anyway, this section's last paragraph governs what it may cite.

Otherwise send at most six items, grouped when they share a person or topic. One line each:

`• What — who; why now. Next: one concrete action. [EC-A1B2C3]`

An item with a Linear identity shows its GCN key instead, and only that: `• What — who; why now. Next: one concrete action. [GCN-42]`. **The branch keys on having a Linear identity, not on Tars having filed it**: an item carrying `linear_issue`, *or* whose own id is `linear:<key>`, renders the key. A `linear:GCN-42` item is never filed by §7 and would otherwise show `[EC-A1B2C3]` for an issue whose real name is on the board — the two-identity problem this skill exists to avoid. Once an item has a Linear identity it has one handle, not two. Bullets sort on a single key: a filed item uses the integer `linear_issue.priority` recorded at §7c — never a live read, and never `priority.value`, which is the shape of a native payload this step does not hold; if it is somehow absent, normalise whatever priority is in hand with `p["value"] if isinstance(p, dict) else p` and treat null as 0. An unfiled item maps its urgency score through the same thresholds — 80 or above → 1, 60 or above → 2, 40 or above → 3, otherwise 4 — and both then sort by the `linear-ticketing` rule: Urgent, High, Medium, Low, then No-priority last.

Include promised/due wording when relevant. Add `+N lower-urgency items retained` only when the cap hid reminder-worthy items. A user may reply in the reminder thread with `tomorrow`, `16:00`, `done`, `ignore`, `waiting for …`, or a short ID.

Mark only delivered items with reminder timestamp and score before final state persistence. Include one `Coverage:` line when a failed mandatory source materially weakens an item-level judgment and no equivalent evidence closes the gap, and **always, on any run that delivers**, for the seven states this skill must never hide:

1. a dedupe scan that could not be proven complete (§7a);
2. a create or close that could not be confirmed (§7c, §5);
3. a close skipped because the issue is no longer on GCN (§5);
4. a truncated Linear window (§4) — `truncated`, with `complete_through` short of the run end;
5. a **source that failed this run**: its `last_success` is older than this run's fixed end timestamp. §1 writes that field on every success and nowhere else, so the test is exact rather than historical. For Linear it means a held cursor behind a run that otherwise looks healthy. Zero candidates on their own are not this condition: a refreshed `last_success` with an empty window is a quiet half hour, and reporting it as a failure is the same mistake §7a step 2 bans;
6. a done item dropped at 30 days with its GCN issue still open (durable state);
7. `N filed tickets untouched for 30+ days` from the staleness sweep, which reports and never writes (durable state).

Say plainly what was not done; never let a run read as if the write had worked. When one of those is the only thing to report and nothing crosses the threshold, break silence with that bare `Coverage:` line alone.

**Rate limit, and its one exemption, stated here once.** Judgment-weakening coverage notes — a source gap that merely softens an item-level call — are rate-limited to once every four hours per distinct condition, keyed on `last_failure_notice[<condition>]`, which exists for exactly that and is written whenever such a notice is emitted. **The seven states above are EXEMPT: they are reported on every delivering run, however recently they last appeared, and `last_failure_notice` is neither consulted nor written for them.** A rate limit that can suppress "the ticket was never created" is a rate limit that makes a broken run read as healthy — the exact failure the list exists to prevent.

§7's filing backlog — `N items awaiting filing` — is an ordinary rate-limited note on this path, **not** an eighth exempt state. The exempt seven are conditions where silence would mislead Gaetan about correctness: a source that failed, a scan that could not be proven complete, a write that could not be confirmed. A filing backlog is none of those. It is self-healing (10 drain per run with no intervention), non-lossy (nothing is dropped; the items stay eligible), and bounded. Exempting it would break silence every 30 minutes for five hours on a 100-item drain — the nagging this skill exists to prevent, and its default is `[SILENT]`. Once every four hours still tells Gaetan it is happening.

Do not restate the rate limit anywhere else; point at this paragraph.

On an internal failure, preserve unadvanced source cursors. If no trustworthy reminder can be produced, return `[SILENT]`; never fabricate substitute evidence. **That ban covers the explanation as much as the evidence.** A delivery-or-silence decision cites only a gate this run actually evaluated, and quotes that gate's live text. The run gate is the "Workday gate" section and the schedule is the "Scheduling contract" section; neither is restated here, so read the live text there rather than recalling it. Measured: a run justified its silence with "the cycle ran outside the 09:00–19:00 window", a window no live text anywhere defines — the live restriction is 10:00–17:00. An invented reason reads to the next person exactly like a real rule and is the same failure as an invented reminder. **When nothing crossed §6's threshold there is no gate to cite**: the run is the bare `[SILENT]` of this section's first rule, and if it is ever asked why it was silent the whole answer is that nothing crossed the threshold. Reaching for a gate is where the invention starts.

## Scheduling contract

**Inactive while reminders are paused.** The recipe below documents a future
explicitly authorized unpause; loading this skill or fixing it does not
activate jobs. Preserve existing jobs/state rather than deleting or recreating
them, and verify every job's active state and target before claiming unpause.

Hermes global timezone must remain `Europe/Paris`. Use two recurring weekday jobs with this skill attached and `--deliver slack` — the bare platform name, which resolves to `SLACK_HOME_CHANNEL` (Gaetan's DM) and posts a **new top-level message**. Never `--deliver origin` and never `--deliver slack:<chat_id>`: both inherit the job's origin `thread_id`, so the reminder lands as a reply buried inside whatever thread the job was created in (measured 2026-08-13 — reminders were landing in an 8 August DM thread). A reminder is never a thread reply, and never goes to the reporting channel `C0BP2GZUFSR`.

- `*/30 10-16 * * 1-5` — 10:00, 10:30, …, 16:30
- `0 17 * * 1-5` — the single final 17:00 pass; no 17:30 run

Both jobs use the same self-contained prompt:

> Run `engagement-checker` end to end. Use Europe/Paris and the fixed run timestamp. Apply the same workday gate as `daily-work-brief` first: weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off; return exactly `[SILENT]` when the gate says not to run. Otherwise acquire the single-writer lock; read durable state; collect and exactly filter only Slack, email, and Linear deltas since each source cursor; reconcile user decisions and resolved loops; update each source cursor only after successful processing; evaluate the pending queue with cooldowns; then file retained open items as GCN issues, close back the GCN issues this skill filed whose items were explicitly resolved elsewhere, and organize GCN issues into Todo, In Progress, Waiting for answer, Delayed, or Done only from direct, conflict-free source evidence under §5a — Done requires explicit proof of resolution, and uncertainty always means no write; persist state; release the lock; then return the compact reminder or exactly `[SILENT]`. Sources stay read-only except for the permitted Linear writes, and the permitted writes are exactly the skill's "Permitted writes" section — that list, not this prompt, is the authority. Do not spawn or prompt another agent; Claude/Orca are read-only evidence sources only. Linear transport is mixed on purpose: run the delta scan through the audited local collector with the inherited `LINEAR_API_KEY`, because its coverage verdict gates the cursor, and take every other Linear read and every Linear write through the native `mcp__linear__*` tools per `linear-ticketing`. Never expose the key; never write outside GCN. Prove every native read complete by paging `cursor` until `hasNextPage` is false, never by narrowing the filter. If a required read cannot be proven complete or a write cannot be confirmed, say so in the delivery rather than proceeding as if it worked.

The two jobs share the same state, so do not create separate per-job cursors.
