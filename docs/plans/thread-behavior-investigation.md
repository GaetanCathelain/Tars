# Investigation: Tars thread behavior (mention-follow + context depth)

Mandate from Gaetan, 2026-08-07 (~20:45Z), reference thread:
`https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786135283424339?thread_ts=1786135149.072829&cid=C08RWSTU9LK`
(channel `C08RWSTU9LK`, thread_ts `1786135149.072829`).

Two questions, then a change if safely possible:

1. **Thread-follow:** can Tars be set to answer Gaetan's messages in a thread
   where Tars was already mentioned/participating, WITHOUT re-mentioning it?
2. **Thread context:** when Tars replies in a thread, how much of the thread
   does the model actually see? Whole thread from Slack, or only its own
   Hermes session history for that thread? What happens for a thread that
   predates its participation?

## Ground truth already established (do not re-derive)

Read first: `CLAUDE.md`, `docs/facts.md`, then evidence
`status/probes/wf4/p02-mention-gating.md` (unmentioned TOP-LEVEL channel
messages from Gaetan are dropped at the adapter gate, silently, pre-model),
`p03-negative.md` (gate order + guardrail pinning; `require_mention: true`,
`strict_mention: true` live), `p16-negative-rerun.md` (allowlist rejects
non-Gaetan mentions in 0.445 s). Sessions bind to message/thread ts
(`p01-dm-roundtrip.md`).

## Work plan

1. **Source:** in the installed Hermes package on the VM (locate via
   `readlink -f ~/.local/bin/hermes`), read the Slack adapter's thread
   handling: how inbound thread replies are gated (does `require_mention`
   consult thread participation? any knob like thread-follow /
   `mention_once_per_thread` / session-existence check?), and how model
   context is assembled for a thread turn (Slack `conversations.replies`
   fetch vs Hermes session store only).
2. **Live measurement** (authorized: send as Gaetan via the VM env creds, in
   the reference thread or a fresh one in `C08RWSTU9LK`; keep message count
   minimal): (a) mention Tars in a thread, get a reply; (b) follow up in the
   same thread WITHOUT a mention → observe answer/silence + the adapter log
   line; (c) for context depth: seed a thread with a distinctive fact in an
   EARLY unmentioned message, then ask Tars later in-thread whether it can
   see it — distinguishes Slack-thread fetch from session-only memory.
3. **Change, if a safe knob exists:** enable answer-without-re-mention for
   threads where Tars participates. Constraints: the ALLOWLIST must remain
   intact (a non-Gaetan in-thread message must still be rejected — verify at
   minimum by source reading of gate order; a live non-Gaetan test needs a
   teammate and is optional), `strict_mention`/`require_mention` semantics for
   top-level messages must not weaken, config edit under
   `flock ~/.hermes/.wf3.lock` with a `.bak`, live-reload preferred over
   restart. If it needs more than config (code/SOUL), PROPOSE with a diff and
   stop — do not apply.
4. **Verify** whatever was changed, live, in the reference thread.

## Deliverables & contract

- Evidence: `status/probes/wf5/thread-behavior.md` (source paths with
  file:line, live message ts/permalinks, log lines, config diff if any).
- Answers doc: append a "Thread behavior" section to `docs/facts.md` (both
  questions answered with evidence pointers).
- Commit early, push with `git pull --rebase origin main && git push origin
  HEAD:main`. Do not edit `status/lane-a.md` (single-writer: the hub).
- Message the hub session (find it via ListAgents — the Tars orchestrator
  session) at milestones: source verdict, live-test verdict, change
  applied/proposed, done. The committed repo is the fallback if messaging
  fails.
- All hard rules in `CLAUDE.md` apply (secrets, flock, never 192.168.0.3,
  `--help` before scripting).
