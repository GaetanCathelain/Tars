# GCN-50 adversarial review — crux A: is lean `claude -p` + SendMessage really the best Tars→session transport?

**Reviewer:** Fable 5, adversarial pass. **Date:** 2026-08-14. **Scope:** wire 3 of the locked
design (`scratchpad/gcn50-design-for-review.md`), attacked against every considered and
un-considered alternative. Grounded in: the six committed probe files, the scratchpad
artifacts, live reads of `~/.claude/sessions/` and `/run/user/1000/cc-socks/` on cooper
(this box, Claude Code 2.1.232), `claude --help`, and the official cross-session-messaging
doc (fetched 2026-08-14). No live sessions spawned, no Slack touched.

## VERDICT: CONDITIONAL

Lean `claude -p` calling SendMessage **survives the head-to-head** — every alternative loses
on at least one axis it wins on — but the design as locked **overclaims its verification
state and leaves the two hardest problems unspecified**: addressing (how Tars learns the
peer name without a race) and failure semantics (what Tars observes when the send does NOT
land). Both have cheap fixes. Confirm the transport only with the conditions in §6.

---

## 1. Evidence audit — is each design claim actually supported?

| Design claim | Evidence | Verdict |
|---|---|---|
| "Lean launcher measured at 3,870 input tokens / ~4.7s … delivery intact" | q15 config G, delivery proven by nonce in `received.txt` (q15 L41, L89-91) | **SUPPORTED — but n=1 per config.** One timed run per configuration, 8 runs total. 4.73s is a single wall-clock sample, no variance. Acceptable for an order-of-magnitude claim; do not treat 4.7s as an SLO. |
| Headless `-p` can call SendMessage (VERIFIED-SENT) | q12-headless L133-143 | **SUPPORTED, honestly labeled.** q12 itself proved only `success:true` + msg_id ("Whether improvements-c8 actually received … was not independently verified", L127-129). **q15 closed the receipt gap** — every A–H run nonce-verified in the receiver's file. So the design's receipt story rests on q15, not q12. Correctly so. |
| Raw socket escape hatch "verified: any uid-1000 process can inject … no token/ancestry check" | q12-nonclaude-socket variations A/B/B2/C (L85-104) | **SUPPORTED.** Clean isolation: token correct/wrong/absent all deliver under `accept`; default config holds. Well-run probe. |
| Channel MCP "works for interactive sessions (verified) but adds a component" | q12-channels attempt 4 delivered, attempts 1–3 (`-p`) failed 3/3 | **SUPPORTED.** |
| Kanban dropped as dominated | q10 (RAW-POST, `session_id` gate at `kanban_watchers.py` L637-646) + q14 W1 (gate is non-empty check, not a handle; template + composed = 2 messages) | **SUPPORTED.** The strongest kill in the file set — W1 provably requires a `hermes chat` turn (W2) as its own precondition. |
| "All probe-verified" (design L20) | — | **OVERCLAIM.** What was verified is the **happy path in a lab shape**: idle receiver, `crossSessionInbound:accept`, receiver launched with a **fixed `-n q15-receiver`** (q15 L10-14), same box, one CC version. Production lanes will have **derived names**, will be **mid-turn busy** for long stretches, and will sometimes be **dead or restarted**. None of those states was probed as a SendMessage target (q12 attempt 1 did hit a derived-name idle session once — the only production-shaped datum). |

### Unverified states that the design silently assumes away

Measured live on cooper during this review:

- `~/.claude/sessions/*.json`: 16 registrations, 16 live `claude` pids — registry currently
  self-consistent. But `/run/user/1000/cc-socks/`: **59 sockets for 16 processes** — the
  socket dir demonstrably accumulates stale entries (cooper-messaging-facts L88 saw 51 vs
  ~13). Registry hygiene after a *crash* (not clean exit) is unproven.
- **Four sessions concurrently registered with cwd = this very worktree**
  (`improvements-c8/-18/-00/-03`), all `nameSource:"derived"`. A "match the lane by cwd"
  addressing scheme is ambiguous **today, on this box** — probe receivers, forks and peer
  sessions pile into the same directory.

Never measured, all reachable in production:

1. SendMessage to a **name that doesn't exist** (lane crashed, registration gone) — what does
   the relay's JSON envelope show? Presumably a tool error; unprobed.
2. SendMessage to a **duplicate name**. Docs: interactively "Claude asks you which one you
   mean" — a `-p` relay **cannot ask**. Behavior unprobed. (Docs also say Claude Code renames
   a colliding live session to a variant, so exact live-live duplicates are rare — but the
   rename itself silently invalidates a name Tars recorded, see §4.)
3. SendMessage to a **busy** receiver. Docs: queued, read "between tool calls", drains at next
   tool round. Fine on paper; a receiver hung in a stuck tool call never reads it, and
   nothing tells Tars. Unprobed.
4. SendMessage to a **held** receiver (bypass-perms without `accept` — q12-socket variation C
   proved this state exists). Doc, delivery section: *"a notice appears there [sender side]
   when the message is held, and a follow-up reports the outcome when the receiver later
   delivers, denies, or expires it."* — **the platform's delivery-outcome channel is a
   sender-side async notice, and the throwaway relay is dead 5 seconds after sending.** The
   lean-spawn shape structurally discards the only native ack mechanism. `success:true`
   means *enqueued at a live peer*, not delivered — the SendMessage tool doc's own words are
   "messages enqueue and drain at the receiver's next tool round" (cooper-messaging-facts L97).
5. **Retry dedupe trap** (doc, Limitations): *"Claude Code rate-limits repeated messages per
   sender, **drops identical repeats arriving within a short window**"*. Tars's natural
   recovery move — re-spawn the relay with the same text after a silence — can be **silently
   dropped as loop protection**. Cheap fix: a per-message nonce in the text. Nobody probed this.

---

## 2. Head-to-head: every transport, scored

Axes: **V** robustness to Claude Code version changes · **A** addressing/discovery ·
**B** busy/held/dead behavior · **O** observability of delivery · **L** latency/cost at N lanes ·
**S** operational simplicity.

| Transport | V | A | B | O | L | S | Kill or keep |
|---|---|---|---|---|---|---|---|
| **1. Lean `claude -p` + SendMessage** (locked) | ~ documented tool; but the lean flag stack (`--tools`, `--setting-sources ""`, `--strict-mcp-config`) is newer surface, re-validate per CC upgrade | needs a name; design leaves discovery unspecified (§4) | queue on busy = correct; held/dead = untested; native outcome-notice lost (sender dies) | JSON envelope parseable (`is_error`, tool result verbatim) — best of field | 4.7s + ~3.9k haiku tokens/msg; a spawn per message; trivially parallel across lanes | 1 ssh + 1 spawn, zero resident parts | **KEEP, conditioned** |
| **2. Raw socket write over ssh** | **worst** — frame format harvested from 2.1.232's own debug log (q12-socket L38-47); no compatibility promise at all | needs `sessions/<pid>.json` lookup (cwd-ambiguous, see §1) | delivers silently under `accept`; **0 bytes back in every probe run** (q12-socket L98) — fire-and-forget | **none.** No ack, no error, no envelope. A dead socket = connect error at best; a held message = nothing | near-instant, zero tokens — best | 1 printf \| socat over ssh — simplest wire, but hand-rolled protocol maintenance | **Escape hatch only — correct call.** For an unattended coordinator, observable failure beats raw speed. |
| **3. File-inbox channel MCP** | preview feature; `--dangerously-load-development-channels` undocumented in `--help`, allowlist pending; **startup confirmation dialog needed `\r` to accept** (q12-channels attempt 4) — a blocker for unattended spawns | n/a (per-session inbox dir) | interactive-only receiver (`-p` fails 3/3); "Claude Code doesn't acknowledge notifications … drops the events silently" (q12-channels §1.4) | silent drop by design | fine | resident MCP server per lane session + drop-dir which is a prompt-injection surface | **DROP — design's reasoning confirmed and understated** (the dev-flag dialog alone kills it) |
| **4. TUI injection (`orca terminal send`)** | Orca-version dependent, not CC | terminal handle from spawn — actually the *best-addressed* option | requires `tui-idle` wait; typing into a busy TUI interleaves with output; the old mailbox's type-but-don't-submit race is incident-logged in this repo | can read the tab after — weak, scrape-based | fast | no new parts but keystroke-fragile | **Last resort — correct call** |
| **5. Resident warm relay (Agent SDK / long-lived `-p --input-format stream-json` under systemd --user)** — *not considered by the design* | SDK spawns the `claude` CLI underneath — same version surface as #1 plus a daemon | same as #1 | **only option that RECEIVES the sender-side held/denied/expired notices** — a persistent sender fixes finding §1.4 | best-in-field once warm | amortizes the 4.7s to ~0 | **a resident stateful component to babysit** — contradicts the design's own stateless-Tars axiom; auth note: `ANTHROPIC_API_KEY` unset on cooper (q15 L49-51), OAuth flows through the CLI, so any SDK path must ride the CLI anyway | **REJECT for now; the named upgrade path** if per-message cost or the ack gap starts to hurt. 4.7s is noise next to wire 2's 14.6s and human answer latency. |
| **6. Session-side inbox drain (SessionStart hook / background watcher posting to OWN socket with `CLAUDE_CODE_MESSAGING_TOKEN`)** — *not considered* | auth frame documented; the `{"type":"user"…}` message frame is only in the debug log — half-documented | trivial (file path, no name) | works even without `accept` (own-child rule) — the one axis where it beats #1 | silent | instant | a hand-rolled resident watcher **per lane session** | **REJECT** — rebuilds channel-MCP out of string; dominated by #1 in the `accept`-everywhere fleet cooper actually runs |
| **7. Kanban wake** | — | — | — | — | — | — | **DROP — conclusively killed by q10+q14** (W1 requires W2 as its own precondition; q14 L406-408) |

**No un-considered option beats the locked choice.** The resident relay (#5) is the only one
that wins an axis that matters (native delivery-outcome notices), and it buys that with a
daemon the design's own doctrine forbids. The raw socket (#2) wins latency, which doesn't
matter, and loses observability, which does.

---

## 3. Does `claude -p` confirm delivery, or only that it SENT?

Precise answer, because the design blurs it:

- **q12 alone proves SENT** (`success:true` + msg_id, explicitly "not independently verified"
  receiver-side — q12 L127-129).
- **q15 proves DELIVERED-AND-ACTED-ON** — but only for an **idle, accept-configured,
  fixed-name receiver**, verified out-of-band by nonce grep, "not by trusting the CLI's own
  `"result":"Sent."` text" (q15 L17-19). The probe author knew the tool result is not receipt.
- **In production the out-of-band verifier doesn't exist.** Tars gets exactly the JSON
  envelope. Per the docs, `success:true` = enqueued at a live peer; the Delivered/Held/Refused
  decision happens on arrival at the receiver, and its outcome is reported via an async
  sender-side notice **that outlives the relay**. So wire 3 as designed has **no delivery
  confirmation at all beyond "a live peer enqueued it"**.
- Consequence: the design's own DoD (#1 round-trip) can pass while the ack problem stays
  invisible, because the happy path masks it. The real ack has to be **application-level**:
  the session's next wire-2 report referencing the relayed content, with a Tars-side timeout
  (cron nudge) on silence. This is workable — but it must be **written into the design**, not
  left implicit.

---

## 4. Addressing: the weakest link in the locked design

The design says LANE.md (Tars-written) holds "the session's peer-name for addressing" — and
never says where Tars gets it. Attack surface:

1. **First-message race.** Spawn (wire 1) returns an Orca terminal handle, not a peer name.
   The name is derived by Claude Code seconds later (`nameSource:"derived"`, folder name +
   2-hex suffix). If Tars's first relay fires before anyone records the derived name, there
   is nothing to address. If the plan is "session writes its name into SESSION.md/LANE.md
   first" — that's a self-registration race *and* contradicts LANE.md's single-writer=Tars rule.
2. **Restart invalidation.** A lane session that dies and is relaunched in the same worktree
   derives a **new** suffix → the recorded name is stale → SendMessage to a name that no
   longer exists (untested failure mode #1 in §1).
3. **cwd matching doesn't rescue it**: four live sessions share this worktree's cwd right now (§1).
4. **The orc-tab "no `-n`" rule is cosmetic, not architectural.** Its own comment says fixed
   names are avoided because they freeze Orca's work-based *tab auto-renaming*
   (cooper-messaging-facts L31-32). The *peer name*, meanwhile, is stable once derived
   (`nameSince == startedAt` on inspected registrations). For Tars lanes, trading tab
   aesthetics for deterministic addressing is obviously correct.

**Fix (kills all four at once):** the lane launcher passes `-n tars-<ticket>[-r<attempt>]`
(`-n, --name` is a documented flag, `claude --help` L128). Tars *chooses* the name before
spawn — zero discovery, zero race, restart bumps the attempt suffix. One residual trap the
docs name: if a live session already answers to that name, Claude Code **silently renames the
newcomer to a variant** → Tars would address the zombie. Hence the post-spawn verification in
§6 step 2.

---

## 5. Two more concrete nits

- **Relay prompt injection.** The relay's user prompt embeds Gaetan's free-text in-thread
  answer, and the relay can message **any** peer on cooper (q12 listed 12 sessions across
  unrelated repos). A weird/hostile payload could redirect `to:`. Cheap containment: pin the
  target session name inside `--system-prompt` (not the user prompt) and pass the payload as
  an opaque fenced block the relay is told to forward verbatim.
- **No replies to a ghost.** The receiver "can reply to the sender the same way" (docs) — to a
  relay that exited. The relay message template must say "do not reply to this message;
  report via your lane wire" or lane sessions will burn turns replying into the void.

---

## 6. Recommended Tars→session design (concrete)

Keep wire 3's transport; specify what the lock left open:

1. **Spawn** with `-n tars-<ticket>` (bump `-r2`, `-r3` on relaunch). Record the *intended*
   name in LANE.md at spawn time — no discovery step, no race.
2. **Verify addressing once per spawn** (not per message): `ssh cooper` → read
   `~/.claude/sessions/*.json`, assert exactly one live registration with
   `name == tars-<ticket>` AND `cwd == <lane worktree>`. Catches the collision-variant
   rename and the crashed-before-registering case. One `python3 -c`/`jq` line.
3. **Send** via the q15 config-G invocation unchanged, plus:
   - `--output-format json`, and Tars parses `is_error`/`subtype` and the tool result — a
     non-`success:true` or missing msg_id is a FAILED send, surfaced to the lane thread,
     never retried blind;
   - a **unique nonce in every message text** (defeats the documented identical-repeat drop
     on retries, and gives the wire-2 ack something to quote);
   - target name pinned in `--system-prompt`; payload as opaque fenced block;
   - the standing line "do not reply to this message; report via your lane wire".
4. **Delivery confirmation is application-level, explicitly:** the message instructs the
   session to acknowledge in its next wire-2 report (quoting the nonce). Tars's existing cron
   surface owns the timeout: no wire-2 activity within N minutes of a relayed answer → Tars
   re-sends once (new nonce), then falls back.
5. **Fallback ladder, pinned with triggers:** (a) relay send errors or times out twice →
   raw-socket write (resolve socket path from the *same verified registration* in step 2) —
   accepting its fire-and-forget nature because step 4's ack still guards the outcome;
   (b) socket connect fails → the lane session is dead → Tars reports the lane down and
   offers respawn. TUI injection stays off the ladder (manual-only).
6. **Version tripwire:** the lean flag stack and the socket frame are both 2.1.232-shaped.
   Add a 30-second smoke (spawn `-n` receiver, config-G send, nonce check — i.e. re-run q15
   config G only) to the post-`claude`-upgrade checklist. Cheaper than discovering it in a
   live lane.
7. **Do not build the resident relay now.** It is the named upgrade path if (and only if)
   per-message cost or the missing native ack starts hurting; it is the only alternative
   that wins an axis this design loses.

---

## 7. Answers to the review's stress questions, one line each

- **HELD?** Sender gets an async notice the dead relay never sees; in the accept-everywhere
  fleet this shouldn't occur, but nothing detects it if it does → step 4's ack-timeout is the guard.
- **Session died?** Untested; expected tool error in the envelope → step 3 parsing + step 5
  ladder. Crash-stale registrations unproven (socket dir provably accumulates them, 59 vs 16).
- **Two messages race?** Docs: queued in order, delivered as a group at next tool round —
  acceptable; unmeasured.
- **Stale/duplicate peer-name?** The real risk (§4); killed by `-n` + post-spawn verify.
- **Does `-p` confirm delivery?** No — enqueue only (§3). q15's receipt proof was out-of-band
  and doesn't exist in production. Ack must be application-level.

**Bottom line: transport choice sound, verification narrative inflated, addressing and
failure-path design missing. Apply §6 and lock it.**
