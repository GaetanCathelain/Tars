# GCN-50 rehydrate P1 — does inlining LANE.md into the wire-2 prompt carry full context?

**Date:** 2026-08-14 (VM clock UTC) · **Host:** Tars VM `192.168.0.9`, Hermes v0.20.0,
gateway `ActiveEnterTimestamp=Fri 2026-08-14 13:02:18 UTC` (unchanged throughout — confirmed
`gateway.log` line count identical before/after: 2841/2841, so nothing in this probe touched
Slack or the gateway at all; no `[PROBE gcn50-rehydrate]` Slack post was needed).

**Scope:** the design under test is: a cooper-side session composes a `hermes chat -Q -q`
prompt that has the FULL contents of its LANE.md **substituted into the query text itself**
(no path reference), then runs that command over ssh on the VM. This tests whether that
inlining actually survives contact with the model — does Tars's reply prove a deep read of a
buried, non-obvious decision, or just a top-of-file echo? A second control confirms the
premise that a bare path reference cannot work cross-host.

---

## 0. Scratch LANE.md (cooper-side only)

Created at
`/tmp/claude-1000/-home-gaetan-dev-orca-worktrees-Tars-improvements/86148edc-827d-4f2c-9864-855a2a981415/scratchpad/p1-lane/LANE.md`
(cleaned up after the probe, full text reproduced below since the path no longer exists).

Contents (3119 bytes with preamble+event wrapper, LANE.md alone was 2298 bytes / 60 lines):
a ticket summary (GCN-77 backoff rewrite), a brief, a 4-item decision log (D1–D4), a buried
nonce `P1-NONCE-82d01d5d` placed mid-document (inside the decision log, not the header), a
"Gaetan answered in-thread" section with a made-up prior answer + `thread_ts
1755168420.004100`, and a Status section.

**Decision D3, the substance being probed** (buried ~40 lines in, between D2 and D4):

> Decision D3: chose a deterministic PRNG seeded from `worker_id` (not `random.random()`
> unseeded) for the jitter, specifically so replaying a captured incident log reproduces the
> exact same backoff timings the worker actually used — unseeded jitter would make
> post-incident replay non-deterministic and useless for verifying a fix against a past
> trace. Do NOT revisit this without re-checking the replay tooling in
> `tools/incident_replay.py`, which assumes seeded-by-worker_id.

A naive top-of-doc echo would produce something generic ("GCN-77 backoff rewrite, decision
pending") — it would NOT be able to name "seeded from worker_id", "incident replay",
"non-deterministic", or "tools/incident_replay.py", none of which appear anywhere near the
top of the file.

---

## 1. INLINE shape (the design under test)

**Mechanism used to get the file content into a single ssh-transported `-q` argument**
without local/remote quoting hazards:

```bash
ssh gaetan@192.168.0.9 '~/.local/bin/hermes chat -Q -q "$(cat)"' < full_prompt.txt
```

`full_prompt.txt` = preamble ("You are Tars, relaying a delegated coding session's status to
Gaetan. The FULL lane state follows verbatim:") + the entire LANE.md text + an event ask
("the session just reached decision point D3 above and is asking whether to proceed to
opening the PR. Compose ONE line for Gaetan that (a) names what D3 actually decided, (b)
says what it's waiting on, and (c) includes the lane nonce verbatim."). This is functionally
identical to embedding `$(cat LANE.md)` straight into the `-q` string — the ssh stdin
indirection only exists to dodge local vs. remote quote-escaping of a multi-KB multi-line
payload; the wire still carries the full content in-band, no path.

**Command + timing:**

```
START=2026-08-14T13:33:40.653Z
RC=0
END=2026-08-14T13:33:49.908Z
```
Elapsed ≈ **9.26 s**.

**stdout (the model's actual answer):**

```
GCN-77 uses a deterministic PRNG seeded from `worker_id` so incident replays reproduce exact
backoff timings; it's waiting on your go/no-go to open the PR — P1-NONCE-82d01d5d.
```

**stderr:** `session_id: 20260814_133341_f94b0d`

**agent.log (full lines, untruncated, via VM):**

```
13:33:44,972 INFO [...] agent.turn_context: conversation turn: session=20260814_133341_f94b0d
  model=gpt-5.6-sol provider=openai-codex platform=cli history=0
  msg="You are Tars, relaying a delegated coding session's status to Gaetan. The FULL l..."
13:33:49,325 INFO [...] agent.conversation_loop: API call #1: model=gpt-5.6-sol
  provider=openai-codex in=31332 out=55 total=31387 latency=4.3s
13:33:49,351 INFO [...] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop)
  model=gpt-5.6-sol api_calls=1/500 budget=1/500 tool_turns=0 last_msg_role=assistant
  response_len=175 session=20260814_133341_f94b0d
13:33:49,564 INFO [...] cli: CLI cleanup calling memory shutdown, 2 message(s)
```

**Verification against the three checks:**

| Check | Result |
|---|---|
| (a) nonce present | **YES** — `P1-NONCE-82d01d5d` verbatim in the answer |
| (b) reflects D3's SUBSTANCE (deep read, not top-of-doc echo) | **YES** — "deterministic PRNG seeded from `worker_id`" + "incident replays reproduce exact backoff timings" is D3's actual reasoning, word-adjacent to the source text, not a paraphrase of the ticket title or brief |
| (c) coherent, one-line, operator-facing status | **YES** — names the decision, states the blocker (go/no-go to open the PR — pulled correctly from the Status section, a THIRD part of the document distinct from both D3 and the nonce), single line |

**Cost:** single API call. **in=31332 out=55 total=31387 tokens, latency 4.3s** (model call
only; wall time incl. ssh+cleanup 9.26s). No tool calls (`tool_turns=0`) — the whole answer
came from the inlined text in the turn context, no file/search tool was invoked or needed.

**Fixed-overhead finding:** the path-only control below (§2) shows an *empty-file* query
still costs `in=30590` tokens on its first call. So the ~31.3k input on the inline run is
**~30.5–30.6k fixed floor (SOUL.md + tool schemas + system prompt) + ~740–800 tokens
marginal** for the actual LANE.md content (2298 bytes ≈ 580 words ≈ consistent with a
~4 chars/token estimate). **Implication for "how large can LANE.md get before this is
costly": the marginal cost is cheap and linear relative to the fixed floor** — a LANE.md
10x this size (~23 KB) would add roughly 7–8k tokens, still under the fixed floor, not yet a
practical problem for a single oneshot call. The real ceiling is the model's total context
window, not this fixed floor — untested here since no LANE.md that large was probed.

---

## 2. PATH-ONLY control (proves inline is not optional)

```
ssh gaetan@192.168.0.9 "~/.local/bin/hermes chat -Q -q \"Read the lane file at
<cooper-absolute-path>/p1-lane/LANE.md and summarize decision D3, including any nonce
found in the file.\""
```

```
START=2026-08-14T13:34:13.160Z
RC=0
END=2026-08-14T13:34:42.277Z
```
Elapsed ≈ **29.12 s** (3.1x the inline run).

**stdout:**

```
I can't summarize D3: the lane file no longer exists at that path, and neither
`/tmp/claude-1000` nor the UUID `86148edc-827d-4f2c-9864-855a2a981415` is present under
`/tmp`. No nonce could be recovered.
```

No hallucination — the model reported the absence honestly rather than fabricating D3's
content or a nonce. (Note: at the moment of this run the scratch dir *did* still exist on
cooper — the failure is real cross-host non-existence, not a race with cleanup, which
happened afterward in §3.)

**agent.log — 5 API calls, 4 tool_turns (vs. inline's 1 call / 0 tool_turns):**

```
13:34:22,902 API call #1: in=30590 out=61  total=30651 latency=4.7s
  [WARNING] Tool kanban_show returned error (0.00s)
13:34:26,840 API call #2: in=30694 out=86  total=30780 latency=3.9s cache=30208/30694 (98%)
  [tools.file_tools] Creating new local environment for task default
  [WARNING] Tool read_file returned error (0.03s)
13:34:29,836 API call #3: in=30897 out=69  total=30966 latency=2.9s cache=30208/30897 (98%)
  [WARNING] Tool search_files returned error (0.05s)
13:34:37,000 API call #4: in=31015 out=73  total=31088 latency=7.1s cache=30208/31015 (97%)
  [tool] search_files completed (0.08s, 18 chars)
13:34:41,380 API call #5: in=31106 out=71  total=31177 latency=4.3s cache=30208/31106 (97%)
  Turn ended: reason=text_response(finish_reason=stop) api_calls=5/500 tool_turns=4
  response_len=203
```

The agent tried `read_file` and `search_files` against its own **VM-local** environment
(`tools.file_tools: Creating new local environment for task default`) — there is no
mechanism by which a VM-side tool call reaches a cooper-side path. It correctly concluded
the file doesn't exist (from the VM's point of view, it genuinely doesn't) rather than
inventing D3's content.

**Cost:** 5 API calls, **total tokens summed ≈ 154,662** (5x the inline run's 31,387) for a
FAILED read, vs. 1 call / 31,387 tokens for a SUCCESSFUL inline read. Cache hit rate ~97–98%
on calls 2–5 (same fixed floor context reused turn-to-turn within the one CLI session).

**Confirms the premise:** a bare path in the prompt cannot cross cooper→VM. Inlining is not
an optimization, it's the only mechanism that works at all for this design.

---

## 3. Cleanup

```
rm -rf /tmp/claude-1000/.../scratchpad/p1-lane
```
Scratch dir removed after both runs completed and this evidence file was drafted.

---

## Verdict

**REHYDRATE-FULL.**

- Inlining LANE.md's full contents into the `hermes chat -Q -q` argument (transported over
  ssh, no path reference) makes it into the model's context intact: a decision buried
  ~40 lines into a 60-line document, a nonce buried in the same section, and a blocker
  pulled from a THIRD, separate section (Status) were all correctly synthesized into one
  coherent operator-facing line. This is deep read, not top-of-file echo — VERIFIED, not
  inferred, against the three checks in the task.
- Path-only reference (no inlining) FAILS cleanly and honestly (no hallucination) because the
  VM has no access to a cooper-side filesystem path — VERIFIED. It also costs 5x the tokens
  and 3.1x the wall time to arrive at that failure, vs. 1 call to arrive at a correct answer
  via inlining.
- Nothing was lost from the inlined document in this test — no partial-read symptom
  observed. This does NOT generalize to arbitrarily large LANE.md files: this single-call
  `hermes chat -Q -q` oneshot is a DIFFERENT code path from the "warm gateway thread session,
  cold-hydrate-once" behavior documented in `docs/facts.md:93` (`wf5/source-context`), which
  governs the Slack **gateway's own** conversation loop reading a Slack **thread's** history
  (one-shot per thread session at 31 messages, with two narrow re-fetch triggers). That
  finding is about a different subsystem (gateway thread rehydration from Slack) than what
  P1 exercises (CLI `chat -q` argument content) — INFERRED that they are unrelated
  mechanisms based on the code-path description in that doc line and the absence of any
  `gateway.log` activity in this probe (line count 2841→2841); NOT separately re-verified
  here that the gateway's Slack-thread hydrate-once limit doesn't also apply if this same
  inlined-LANE.md prompt were later relayed onward through a Slack thread rather than
  answered directly by `chat -q`.
- Fixed per-call overhead is large (~30.5–30.6k tokens, SOUL.md + tool schemas) and dwarfs
  LANE.md's marginal cost (~740–800 tokens for a 2.3 KB file) — cost scaling with LANE.md
  size is cheap and linear up to at least the model's context window; that ceiling itself
  was not probed.

**Bottom line:** inline-vs-path — VERIFIED, path-only is a hard fail, inline is the only
mechanism that carries context cross-host, and it carries it FULLY and DEEPLY at this
document size. Token/time cost of inline: 1 API call, ~31.4k tokens, ~9.3s wall. Cost of
the (failing) path-only alternative: 5 API calls, ~154.7k tokens, ~29.1s wall.
