# GCN-50 rehydrate P2 — in-thread wake with NO LANE.md pointer: can it find its lane?

**Date:** 2026-08-14 (VM clock UTC) · **Host:** Tars VM `192.168.0.9`, Hermes v0.20.0 ·
**Method:** read-only source/log/DB inspection on the VM. No `hermes send`, no `hermes chat`,
no Slack write of any kind was run for this probe — **no `[PROBE gcn50-rehydrate]` Slack
artifact exists for P2**, none was needed. No `sops -d`, no `config.yaml` edit, no secret
value read or printed.

**Question:** when a Tars turn wakes from an in-thread event (Gaetan replying inside a lane
thread) it gets no LANE.md pointer in the trigger payload. Can it find its lane and recover
the Q&A anyway? Is a `thread_ts → lane` reverse index required?

---

## 1. Session store + CLI: is there a thread_ts lookup?

**`~/.hermes/sessions/sessions.json`** (VERIFIED, key names only, no content read) — 103
entries, keyed literally as
`agent:main:slack:dm:<team_scope?>:<chat_id>:<thread_ts>`, e.g.:
```
agent:main:slack:dm:D0BBYNM01BL:1786710890.363749
agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786134530.441039
```
So Hermes' **own** session store is already, structurally, a `thread_ts → session` index —
but a *session* is a conversation-memory handle, not a *lane* (no worktree path, no ticket
id, no LANE.md pointer lives in this key or its value beyond the running transcript).
`gateway/session.py::build_session_key` (read on VM, `~/.hermes/hermes-agent/gateway/session.py`)
confirms the key construction: DM thread key = `[ns, platform, "dm", scope_id?, chat_id,
thread_id]` joined by `:` — **VERIFIED from source**, matches `docs/facts.md:89`.

Note in passing, not otherwise relevant to P2: two key SHAPES coexist for the *same*
`thread_ts` (`1786134530.441039` appears both with and without the `T7V1UGJ82` scope
segment) — consistent with the 2026-08-13 DM-cutover config change altering whether
`scope_id` is populated. Session continuity for threads created before that change would be
broken by key-format drift alone; **INFERRED**, not probed further (out of P2 scope).

**`hermes` CLI** (`--help`, `chat --help`, `gateway --help`, `debug --help`, `kanban --help`,
`kanban list --help`, `kanback notify-list --help` — all read on VM): **no verb resolves
`thread_ts → anything`.** `chat --resume SESSION_ID` / `--continue [NAME]` take a session id
or name, never a thread ts. `kanban list --session SESSION` filters by *originating agent
session id*, not thread. `kanban notify-list [task_id]` is forward-keyed (task→subscribers).
**VERIFIED: no lookup-by-thread_ts CLI verb exists anywhere in the `hermes` binary.**

**Adjacent existing mechanism (reuse candidate, ladder step 2):** `~/.hermes/kanban.db` has a
table **`kanban_notify_subs`** (schema read via read-only sqlite3, VERIFIED) with columns
`task_id, platform, chat_id, chat_type, thread_id, user_id, notifier_profile,
delivery_metadata, created_at, last_event_id` — `thread_id` is a plain column, so
`SELECT task_id FROM kanban_notify_subs WHERE thread_id=?` is a real, already-persisted
`thread_id → task_id` reverse index **today**, for lanes that are bound to a kanban card via
`kanban notify-subscribe ... --thread-id`. This is exactly the shape P2 asks for, already
built — but it only covers kanban-card-bound lanes, and only fires automatically on a
**kanban-notifier wake** (a card's terminal-status event), which the prior W1/W2 probe showed
delivers a synthetic `[kanban] Task <id> completed. Title: …` trigger message the woken turn
then resolves via a `kanban_show` tool call (observed in that probe's `agent.log`). **P2's
scenario is different and NOT covered by this**: an *ordinary* Slack thread reply from
Gaetan (no kanban event involved) goes through the Slack adapter's thread-reply path, which
never touches `kanban.db` at all — confirmed by reading the call site (§2 below): nothing in
`_fetch_thread_context` / `_has_active_session_for_thread` references kanban.

---

## 2. Warm vs cold hydrate — VERIFIED, source + log evidence

`docs/facts.md:89-93` claim: session key omits user id when threaded (a thread is one shared
session); cold start fetches the real Slack thread once via `_fetch_thread_context` →
`conversations.replies(limit=31, inclusive=True)`, guarded by `_has_active_session_for_thread`;
steady state makes zero Slack calls; two narrow re-fetch triggers (explicit @mention delta,
once-per-process-restart delta).

**Read the actual source** (`~/.hermes/hermes-agent/plugins/platforms/slack/adapter.py`,
`~/.hermes/hermes-agent/tests/gateway/test_session_dm_thread_seeding.py`) to pin down the
exact mechanism, since agent.log's `msg=` field is truncated to ~80 chars by the app's own
log formatter and cannot show an injected context block directly (confirmed by measuring a
raw on-disk line length = 283 bytes vs. the visible `msg=` slice — the metadata fields
`session=/model=/provider=/platform=/history=` are NOT truncated, only the trailing message
text is):

- **DM thread sessions start with an EMPTY transcript, by explicit design** — no
  parent-DM-transcript seeding. `test_session_dm_thread_seeding.py`'s own docstring: "DM
  thread sessions must start empty — no parent transcript seeding... Session-level seeding
  was removed because it copied the ENTIRE parent DM transcript, causing unrelated
  conversations to bleed across threads." **VERIFIED from source + tests.**
- Instead, thread context is injected **once**, at `adapter.py:5793-5829`: on a thread-reply
  event where `_has_active_session_for_thread()` is False (no existing, non-reset-pending
  session entry for this exact `thread_ts` key), it calls `_fetch_thread_context(...)`
  (→ `conversations.replies`) and folds the result into `channel_context` for that one turn,
  then marks `_mark_thread_rehydration_checked(...)` so it never fires again for this
  thread+process. **VERIFIED from source.**
- Two guarded re-fetches confirmed in source: an **explicit @mention on an already-active
  thread** forces a `force_refresh=True` delta fetch of everything past the last watermark
  (`adapter.py:~5836-5864`); and the **first ordinary reply after a gateway process
  restart** does one watermark-delta fetch, gated by `rehydration_key not in
  self._thread_rehydration_checked` so it costs at most one `conversations.replies` call per
  thread per process lifetime (`adapter.py:~5875-5905`, comment cites #63530/#33215).

**Corroborating log evidence** (agent.log, `agent.turn_context: conversation turn:
session=... history=N`, parsed across all 89 sessions with a `turn_context` line): 9 sessions
had ≥3 turns. Every one starts `history=0` on turn 1, then a **sharp one-time jump** on an
early turn (e.g. `[0, 66, 66, 91, 95, ...]`, `[0, 82, 82, 105, 107]`, `[0, 20, 34, 41, 41,
61]`), then only small incremental growth thereafter. This matches the source mechanism
exactly: turn 1 = thread root / first message (not a thread reply, no fetch); the jump = the
one-shot cold hydrate firing on the first genuine reply; flat/small deltas afterward = steady
state, zero further Slack calls, growth is just the session's own accumulating transcript.
**VERIFIED** (source) and **corroborated** (log pattern matches, though the log's `history=`
counter is circumstantial evidence, not a direct trace of the fetch call itself — the fetch
call has no dedicated log line at INFO level).

**Bottom line for P2:** a woken in-thread turn on a **brand-new** thread session DOES see
that thread's own prior Slack messages once (up to ~31, root + replies) — so if a lane's
anchor/status posts (wire W2 in the sibling probe) are in-window, the woken turn can read
their literal text. On a thread that is **already warm** (this process already ran ≥1 turn
in it), no further Slack read happens on ordinary replies — the turn only has its own
accumulated transcript, which itself was seeded, once, from that same capped 31-message
window plus everything it personally said/heard since.

---

## 3. Conclusion — is a thread_ts→lane index REQUIRED?

**Yes.** Seeing prior thread text (§2) is not the same as finding the lane, for three
source/probe-backed reasons:

1. **Prose ≠ pointer.** The cold-hydrated (or self-accumulated) thread content is whatever
   was literally posted — free text. There is no structured field anywhere in the session
   store, the Slack event, or the hydrate mechanism that carries a lane id, ticket id, or
   file path. If a lane's identity was never spelled out in-thread (or scrolled out of the
   31-message cap on a long-running lane thread), the woken turn has **nothing** to resolve
   it from — it must either guess from prose or fail. This is the general form of the
   already-logged GCN-49 failure (`status/lane-a.md`: fresh turn answered a wrong-anchored
   boolean, root-caused to no conversation-state rule) — the same structural hole, on any
   thread that isn't a same-turn continuation.
2. **Even a correctly-parsed path is a dead end.** The sibling probe P1
   (`status/probes/gcn50-rehydrate-p1-inline.md`, §2 PATH-ONLY control) proved this directly,
   VERIFIED, not theoretical: a `hermes chat` turn on the VM given a *cooper-hosted* LANE.md
   path tried `read_file`/`search_files` against its own **VM-local** environment (log:
   `tools.file_tools: Creating new local environment for task default`) and correctly, but
   uselessly, reported the file doesn't exist — 5 API calls / ~154.7k tokens burned to fail,
   vs. 1 call / ~31.4k tokens to succeed when the content was inlined instead of pathed. A
   thread_ts→lane index that resolves to a **bare cooper path** inherits this exact failure
   for any lane living in a cooper worktree (which per the repo map, `LANE.md` always is).
3. **The existing kanban reverse-index (§1) proves the pattern works — but only covers
   kanban-bound lanes and only fires on a kanban-notifier wake**, a different trigger than
   the ordinary in-thread Slack reply P2 asks about. Nothing wires the two together today.

### Where the index should live, and what writes it

Given (2), the index cannot be "a path" alone if the lane lives off-VM — it must resolve to
something the woken turn (VM-local tools only, confirmed by P1) can actually read. Cheapest
fix, reusing what the repo already does for `SOUL.md`/`skills/`:

- **A VM-local file per lane, keyed by thread_ts**, e.g. `~/.hermes/lanes/<thread_ts>.md` (or
  one JSON map `~/.hermes/lanes/index.json` for cheap enumeration + a per-thread `.md` for
  content — the index is a directory listing away either way). Content = the same thing P1
  proved works: **inlined** lane state (ticket, decision log, current status), not a path —
  i.e. a VM-side mirror of LANE.md's substance, same shape as the repo's existing
  `SOUL.md`/`skills/<path>` live-mirror convention (`CLAUDE.md` §Repo map).
- **Written by whichever turn posts into a lane thread for the first time** (the wire-W2
  anchor post) and **refreshed by every subsequent wire-W2 status turn in the same turn it
  posts** — same-turn discipline, mirroring the fix the sibling review already proposed for
  hole #4 (LANE.md currency) and the repo's own `flock` + tmp+mv + `chmod 600` doctrine
  (`CLAUDE.md` hard rules) so a crash mid-write can't hand a later reader a truncated file.
- A woken **in-thread ordinary-reply** turn needs a standing instruction (SOUL-level, not
  per-prompt, since the trigger payload carries nothing — that's the whole premise of P2) to:
  check `~/.hermes/lanes/<this thread_ts>.md` before answering, fall back to
  `kanban_notify_subs WHERE thread_id=?` if the lane happens to also be kanban-bound (§1), and
  only then fall back to whatever thread prose the one-time cold hydrate gave it.

### Exact failure if absent

Gaetan replies "go ahead" (or asks a question) in a lane thread. The gateway wakes a turn
keyed on that `thread_ts`. With no index:
- **Cold thread, in-window:** turn parses free prose from ≤31 messages, must *guess* which
  lane/decision this answers — no verification, no way to distinguish two lanes that happen
  to share vague phrasing.
- **Warm thread, or content aged out of the 31-message cap:** turn has only its own partial
  transcript (steady state, zero further Slack calls — §2) or literally nothing usable. It
  answers blind: composes a plausible reply with no access to the lane's actual current
  LANE.md/decision log (even if it somehow named the right path, §3.2 shows that path is
  unreadable from the VM). This is the precise "wakes in-thread, cannot find LANE.md, answers
  blind" failure the task asked P2 to identify — **CONFIRMED as the real failure mode**, not
  hypothetical: both preconditions (no structured pointer in the wake payload, no cross-host
  file access) are independently source/probe-verified above.

---

## Evidence provenance summary

| Claim | Status |
|---|---|
| Session key format `agent:main:slack:dm:...:<thread_ts>` | VERIFIED (source: `gateway/session.py::build_session_key`, cross-checked against live `sessions.json` keys) |
| No CLI verb does thread_ts→anything lookup | VERIFIED (`--help` on every relevant subcommand) |
| `kanban_notify_subs` table has a `thread_id` column, usable as a reverse index for kanban-bound lanes only | VERIFIED (read-only schema query) |
| DM thread sessions start empty, no parent-transcript seeding | VERIFIED (source + `test_session_dm_thread_seeding.py`) |
| One-shot cold hydrate on first thread reply, guarded by `_has_active_session_for_thread`, zero Slack calls in steady state | VERIFIED (source read, `adapter.py`); corroborated by `history=` counter jump pattern across 9 multi-turn sessions in `agent.log` |
| agent.log `msg=` field truncation is the app's own log formatter, not a display artifact | VERIFIED (measured raw line length 283B vs. visible `msg=` slice, via a python one-liner run directly on the VM) |
| Path-only cross-host LANE.md read fails; inlined content succeeds | VERIFIED — carried over from sibling probe `gcn50-rehydrate-p1-inline.md`, not re-run here |
| thread_ts→lane index is REQUIRED for the operator-gate | Conclusion drawn from the above VERIFIED facts, consistent with the independent finding already logged in `gcn50-review-fable-session-to-tars.md` hole #2 |
