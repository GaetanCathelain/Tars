# GCN-50 rehydrate — P3: can the question/escalation turn BE Tars via `hermes chat --resume`?

**Date:** 2026-08-14 (VM clock UTC) · **Host:** Tars VM `192.168.0.9`, Hermes v0.20.0
**Run by:** Claude Code probe subagent on cooper · Slack DM `D0BBYNM01BL`

**Verdict: RESUME-RESTRUCTURES.** `hermes chat --resume <session_id> -Q -q '<prompt>'`,
run headless over non-interactive ssh, resumed a real **gateway** (`source=slack`) DM-thread
session: it recalled the correct prior topic and answered in Tars' French DM voice, with a
single API call and no tool fishing. Output is on stdout and pipes cleanly into
`hermes send -t slack:<chat>:<thread_ts>` exactly like wire W2. No TTY, no live process, and
no still-running session required — only the session id has to exist as a row in
`~/.hermes/state.db`. VERIFIED end to end.

---

## 1. `hermes chat --help` — flags (VERIFIED, quoted from VM)

```
--resume SESSION_ID, -r SESSION_ID
                      Resume a previous session by ID (shown on exit)
--no-restore-cwd      Don't cd into a resumed session's recorded working directory.
--continue [SESSION_NAME], -c [SESSION_NAME]
                      Resume a session by name, or the most recent if no name given
```
`SESSION_ID` is the internal id (format `YYYYMMDD_HHMMSS_<hex>`), not a Slack thread ts —
confirmed against `~/.hermes/state.db` (below). `-Q` (quiet) still applies with `--resume`:
banner/resume-status line goes to **stderr**, answer only on **stdout** (source:
`cli_agent_setup_mixin.py:368-377`, comment cites issue #11793, written specifically so
`$(hermes chat -Q --resume <id> -q "...")` stays pipe-safe).

## 2. Session store + target selection (VERIFIED)

`docs/facts.md` (line 95): canonical store is **`~/.hermes/state.db`** (SQLite); the JSON
file under `~/.hermes/sessions/` is a legacy mirror only — confirmed by querying the DB
directly (`sessions` table schema pulled via `sqlite_master`, `python3` + a `mode=ro` URI
per the documented pattern, no `sqlite3` binary on the VM).

Queried `sessions where chat_id='D0BBYNM01BL' order by started_at desc` — a **real gateway
DM-thread session already existed as a benign probe artifact** from the prior gcn50-q14
run:

```
('20260814_123532_e2cc0014', 'slack', 'D0BBYNM01BL', 'dm', '1786710890.363749', ...,
 message_count=5, ended_at=None)
```

This is the `[PROBE gcn50-q14-W1]` kanban-wake-test thread (source=`slack`, i.e. a genuine
gateway-processed session, not a bare CLI session) — used as the resume target per "prefer a
benign/probe thread." Any Slack-visible artifact this probe produced was still kept in its
**own** dedicated thread (see §5) rather than posted into that older probe's thread, per the
gcn50-rehydrate prefix/dedicated-thread rule.

Anchor for this probe: `hermes send -t slack:D0BBYNM01BL '[PROBE gcn50-rehydrate] p3
anchor'` → ts `1786715161.981479`.

## 3. Resume test (VERIFIED)

```
$ ~/.local/bin/hermes chat --resume 20260814_123532_e2cc0014 -Q -q \
    "[PROBE gcn50-rehydrate] in one line, what topic/lane were we just discussing?" \
    2>stderr.txt
La lane de test de réveil/réhydratation Kanban (`gcn50`), sans action métier.
--- stderr ---
↻ Resumed session 20260814_123532_e2cc0014 "Validation réussie du test wake" (1 user
message, 4 total messages)
session_id: 20260814_123532_e2cc0014
```
Wall clock: 13:46:08.245Z → 13:46:16.720Z (**~8.5 s**), RC=0, headless over a plain
non-interactive `ssh` (no `-t`, no pty).

`agent.log` delta for `20260814_123532_e2cc0014`:
```
13:46:13,146 agent.turn_context: conversation turn: session=20260814_...
13:46:13,147 agent.conversation_loop: Stored system prompt for session...
13:46:16,146 agent.conversation_loop: API call #1: model=gpt-5.6-sol ...
13:46:16,174 agent.conversation_loop: Turn ended: reason=text_response...
```
One API call, no tool call — the answer came straight from injected history, not from a
lookup.

**(a) Context continuity — YES.** The recalled topic ("réveil/réhydratation Kanban, gcn50")
is factually correct: that session's real history (from gcn50-q14 W1) is the kanban
wake-path test on card `t_bde61b26` — `kanban_show` tool call + a French composed reply
("Probe réussi : la tâche `t_bde61b26` est *done*..."). The resume genuinely re-loaded that
transcript (stderr's "4 total messages" matches the 1 trigger + kanban_show call/result +
1 assistant reply from the original wake turn).

**(b) Persona — YES.** Reply is in French, in Tars' terse operator register — despite my
query being in **English**. SOUL.md's language rule ("I reply in the language of the
message I received") is source-independent and loads into bare `hermes chat` too (seeded via
`--ignore-rules`-gated auto-injection, not gateway-specific) — so a cold CLI turn is not
missing SOUL.md itself. What resume adds is the **conversation's own prior language/register**
dominating the reply over the literal-language-mirror default: the resumed turn answered in
French (matching the thread's established language) rather than English (matching my
query's language). This — not "SOUL.md is loaded" — is the actual persona-restoring effect
of `--resume`. INFERRED (single sample; not re-tested with a French query to isolate the two
effects).

**(c) Pipe → in-thread delivery — YES.**
```
$ ~/.local/bin/hermes chat --resume 20260814_123532_e2cc0014 -Q -q \
    "[PROBE gcn50-rehydrate] p3: resumed reply, piped to confirm in-thread delivery." \
  | ~/.local/bin/hermes send -t slack:D0BBYNM01BL:1786715161.981479 --json
{"success": true, "platform": "slack", "chat_id": "D0BBYNM01BL", "message_id": "1786715195.840549"}
```
`slack_read_thread` on the dedicated anchor (`1786715161.981479`) confirms delivery:
```
Reply 1 of 1 — From: Tars — 15:46:35 CEST
Reçu : reprise p3 confirmée dans ce fil.
```
Landed in-thread, composed (French), same mechanic as wire W2's `chat -q | send -t` pipe.
Second resume call's stderr ("2 user messages, 6 total messages" as the *pre-existing* count)
confirms session state persists correctly across repeated headless resumes.

Minor note: the composed reply text itself does not literally start with the
`[PROBE gcn50-rehydrate]` tag (Tars paraphrases rather than echoing the prompt) — the thread
root and every command I issued do carry the tag, and this stayed in its own dedicated
thread, so the hard rule is met at the thread level even though this one reply isn't
self-tagged. Flagging rather than silently treating as compliant.

## 4. Cold control — same question, no `--resume` (VERIFIED, with a confound)

```
$ ~/.local/bin/hermes chat -Q -q \
    "[PROBE gcn50-rehydrate] in one line, what topic/lane were we just discussing?" \
    2>stderr.txt
The Kanban wake/rehydration test lane (`gcn50`), with no product work involved.
--- stderr ---
session_id: 20260814_134649_633bbb
```
Wall clock: 13:46:48.613Z → 13:47:07.220Z (**~18.6 s**, more than 2× the resumed call).

`agent.log` delta — **3 API calls**, not 1:
```
13:46:53,533 agent.turn_context: conversation turn...
13:46:57,772 API call #1
13:46:57,787 WARNING tool_executor: Tool kanban_show returned error
13:47:02,767 API call #2
13:47:03,126 tool session_search completed (0.35s)
13:47:06,625 API call #3
13:47:06,640 Turn ended: reason=text_response
```

**Confound, called out explicitly:** the cold answer looks eerily similar to the resumed
one, but it is **not** genuine zero-context recall. The bare session has no injected
conversation history, but it *does* have a `session_search` tool (cross-session search over
`state.db`), and the query text I was instructed to use verbatim contains the literal string
`gcn50-rehydrate` — which is enough of a keyword hint for the model to search, find the
related probe sessions, and reconstruct an answer from them. The tool-call trace proves this:
`kanban_show` (failed — no active card in *this* session) → `session_search` (succeeded) →
answer. So "cold" here means "no --resume", not "no memory access" — a plain `hermes chat`
turn with the kanban/session toolsets enabled can self-serve context via tools if the prompt
gives it a search hook, just slower (3 API calls, ~2.2× wall time) and less reliably (depends
on the prompt leaking a searchable term — an escalation prompt phrased without a hint like
"gcn50" would very likely not self-discover the topic this way).

Language: **English** — matches the query language, no French-DM-voice effect, consistent
with §3(b)'s reading that persona-voice-carryover needs an actual resumed thread, not just
tool-based context lookup.

## 5. Coupling / footprint (VERIFIED unless noted)

- **No TTY needed.** Every command above ran over a plain non-interactive `ssh …` (no `-t`).
- **No live process needed.** The original session (`20260814_123532_e2cc0014`) had exited
  hours earlier; resume works purely by reading `state.db` — confirmed both by code
  (`_preload_resumed_session` in `hermes_cli/cli_agent_setup_mixin.py:576-655`: looks up
  `session_meta` by id, loads `get_resume_conversations`, then `UPDATE sessions SET
  ended_at=NULL, end_reason=NULL WHERE id=?` to "reopen" it — no process/socket/handle
  involved) and by the successful live test.
- **Only requirement:** the session id must exist as a non-deleted row in `state.db` with
  message history. Source (`cli` vs `slack`) is not checked anywhere in the resume path —
  resume is source-agnostic by construction, which is exactly what let this test target a
  gateway-origin session at all.
- **Side effect on the target row:** resume clears `ended_at`/`end_reason` (a no-op here,
  it was already `NULL`) and appends new turns to that session's history in `state.db`. Used
  an already-inert, already-`[PROBE gcn50-q14-W1]`-tagged thread for this, so no real
  conversation was touched — but adopting this wire for production question/escalation turns
  means every such turn **mutates the live gateway session row** it targets.
- **Untested risk (INFERRED, not measured):** if Gaetan's real gateway session for a thread
  is concurrently resumed by a delegated CLI turn (e.g. an escalation firing while Gaetan is
  actively chatting in that same thread), both write to the same `state.db` row with no
  locking observed in the read code path above. Not exercised — flag before adopting on a
  thread that might be live at question-time.
- **No gateway restart, no config/`.env` edit.** `ActiveEnterTimestamp` (13:02:18 UTC) and
  `config.yaml` mtime (13:02:08 UTC) both predate this probe (13:46Z) and are unchanged by
  it. `errors.log` delta is routine `tools.registry check_fn … returned False` capability
  probes (fire on every session start) plus the one expected `kanban_show` warning from the
  cold-control's speculative tool call — nothing new/anomalous.
- No `sops` invocation, no secret read, no `config.yaml` edit, per hard rules.

## 6. What this means for wire 2 (question/escalation turns)

`--resume <gateway_session_id>` gives a headless CLI turn the resumed thread's actual
conversation history and its established persona/register (§3a, §3b) at **roughly 2× the
speed** of letting a cold turn self-serve context via tools (§4), with a **cleaner mechanism**
(single API call from injected history vs. speculative multi-call tool search that depends on
prompt phrasing). Composable with the proven W2 pipe (`| hermes send -t
slack:<chat>:<thread_ts>`) with no changes to that half of the wire.

Trade-off vs. plain W2 (bare `hermes chat -q`, gcn50-q14): resume requires **knowing the
target session id up front** (must be looked up from `state.db`, there is no "resume the DM
with Gaetan" shorthand tested here — `--continue` by name was not exercised) and **mutates a
live session row** rather than running in an ephemeral, side-effect-free CLI session. For a
question/escalation turn that must land in a *specific existing* thread with continuity,
resume is the better fit; for a fire-and-forget milestone report into a fresh/anchor thread
(gcn50-q14's original use case), plain W2 remains simpler and side-effect-free.

## Cleanup

No cleanup needed/performed: `[PROBE gcn50-rehydrate]` anchor thread
(`D0BBYNM01BL:1786715161.981479`) left in place per instructions; resumed target session
(`20260814_123532_e2cc0014`) already carried the `[PROBE gcn50-q14-W1]` tag and needed no
further tagging. Not committed, per instructions.
