# Daily report threading — one parent per Paris day, replies in-thread

Status: spec settled 2026-08-10 (this file), implemented by
`scripts/tars-report-thread.py` + one new no-agent cron job on the VM.

## Problem

The three report jobs (daily-work-brief `e231e5faf180`, engagement-checker
`62e8cd9db637` + `759e08c598e3`) all deliver to the static target
`slack:C0BP2GZUFSR:1786359613.979759` — every day's reports pile into one
old Aug-8 thread. Wanted: **one top-level reporting message per Europe/Paris
calendar day** (normally the 08:30 daily brief itself), with every later
material report that day threaded under it. `[SILENT]` runs must keep
posting nothing.

## Design: a deterministic thread reconciler (no core changes, no skill changes)

Everything that already works stays untouched:

- **Posting** stays core cron delivery (`cron/scheduler.py _deliver_result`
  → Slack adapter). Thread targets `slack:CH:TS` are first-class there.
- **`[SILENT]` suppression** is core (`_is_cron_silence_response`, gate in
  `run_one_job`) — a silent run never reaches delivery. Zero Slack writes.
- **Schedules** are evaluated in Europe/Paris (`config.yaml timezone` via
  `hermes_time.py`), so cron expressions below are Paris wall-clock.
- **Skills** (`daily-work-brief`, `engagement-checker`) are not modified;
  their prompts, sources, cursors, shared engagement state and failure
  semantics are unchanged.

The only new piece is `scripts/tars-report-thread.py`, run as a **no-agent
`--script` cron job** (same shape as the existing mc-metarepo-refresh job)
at `25,55 7-16 * * 1-5` Paris. Each tick it decides the correct delivery
target for the day and applies it to the three report jobs via the official
CLI — it **never posts any Slack message itself**:

1. `today` = current date in Europe/Paris.
2. Fast path: if the state file already has today's `parent_ts` and all
   three jobs already target it → exit (no Slack call).
3. Otherwise reconcile against Slack: `conversations.history` for
   `C0BP2GZUFSR` since Paris midnight; the day's parent is the **earliest
   top-level message authored by the Tars bot** (`U0BBH85NAKH`;
   top-level ⇔ no `thread_ts` or `thread_ts == ts`; messages with a
   `subtype` — e.g. `channel_join` — are never parents).
4. Desired deliver value:
   - parent found → `slack:C0BP2GZUFSR:<parent_ts>` (same-day reports
     become thread replies);
   - no parent yet → `slack:C0BP2GZUFSR` (the next material report posts
     top-level and *becomes* the day's parent).
5. For each of the three jobs whose current `deliver` (read from
   `~/.hermes/cron/jobs.json`, read-only) differs from the desired value:
   `hermes cron edit <id> --deliver <desired>`. Schedules never touched.
6. Persist state atomically (`~/.hermes/state/daily-report-thread.json`,
   tmp+rename, under a flock): `{channel, date, parent_ts, updated_at}`.
7. Print nothing on stdout (empty output is the scheduler's silence
   marker, same as the kb-refresh script job); diagnostics go to stderr.
   On any failure exit nonzero — the scheduler delivers a failure alert to
   the reconciler job's own `--deliver` target, which is set to the home
   DM `slack:D0BBYNM01BL` so breakage reaches Gaetan without touching the
   reporting channel.

## Why this satisfies the constraints

- **Brief as parent, atomically**: at the first tick after Paris midnight
  the jobs are flipped to top-level `slack:C0BP2GZUFSR`; the 08:30 brief is
  then posted by core delivery as a single ordinary message — it *is* the
  parent. No placeholder/heading message exists.
- **Daily missed → first material engagement report becomes parent** for
  free: deliver is still top-level until a parent exists.
- **No-op stays no-op**: `[SILENT]` runs are dropped by core before
  delivery; on a fully silent day (e.g. French holiday) no parent and no
  placeholder is ever created.
- **No duplicate parents**: Slack itself is the source of truth — the
  reconciler adopts an existing parent before ever leaving the jobs
  top-level. The send-succeeded/state-write-failed window collapses because
  state is derived from Slack, not the other way round. Ticks at :25/:55
  bracket every report fire (:00/:30, 08:30, 17:00) so at most one material
  report can post top-level per day in normal operation.
- **Rollover**: first tick of a new Paris day sees no parent for `today`
  and flips the jobs back to top-level. Old threads are never reused.
- **Degradation, not loss**: if the reconciler dies, reports keep flowing
  to the last-set target (worst case: yesterday's thread — the current bug,
  as the failure mode instead of the steady state). Nothing is ever lost or
  double-posted by the new code because it posts nothing.

## Credentials

The reconciler reads Slack with the bot token (`SLACK_BOT_TOKEN`) the
gateway already holds. Script cron jobs run with a **sanitized env** —
`_sanitize_subprocess_env` strips messaging-category vars — so the script
checks `os.environ` first (future-proof for `terminal.env_passthrough`)
and otherwise parses `~/.hermes/.env` in-process. The token value is never
printed, logged, persisted, or placed on argv; it exists only in process
memory as the Authorization header.

## Testing

`tests/test_tars_report_thread.py` (stdlib unittest, runs on cooper or the
VM) drives the script end-to-end against a temp `~/.hermes` tree, a local
fake Slack HTTP server, and a stub `hermes` binary that records `cron edit`
invocations. Covered: first tick of day (top-level flip), parent discovery
(thread flip), fast-path no-op (no HTTP), Paris rollover, adoption of an
existing parent when state was lost (retry/reconciliation), ignoring
non-bot and threaded messages, dry-run, `[SILENT]`-only stdout.

## Ops

- Disable: `hermes cron edit <reconciler-id> --disable` (or delete the job);
  then optionally pin the three jobs' `--deliver` back to a static target.
- Observe: reconciler stderr/log + `hermes cron list` (job status), state
  file `~/.hermes/state/daily-report-thread.json`.
- Known edges (accepted): a non-report top-level bot message posted in
  C0BP2GZUFSR after midnight and before the brief would be adopted as
  parent (conversational replies are threaded and lifecycle notifications
  are disabled, so this is rare); a brief longer than the adapter's 39 000
  char chunk limit posts sibling top-level chunks and the first chunk
  becomes the parent; history discovery reads a single page (limit 999) —
  fine for a channel with a handful of messages per day.

## Transition (2026-08-10 deploy day)

C0BP2GZUFSR was created 2026-08-10; the three jobs were hand-pointed at
that day's thread (`1786359613.979759`, parented by a *human* message from
Gaetan, which bot-only discovery would rightly ignore). At deploy time the
state file is seeded with that ts so the reconciler no-ops for the rest of
the day and the jobs keep threading where Gaetan put them; the first Paris
rollover tick (07:25 next weekday) flips to top-level and the 08:30 brief
becomes the first true bot parent.
