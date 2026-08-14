# GCN-50 E2E round-trip — spawn → Wire-2 question → operator answer → Wire-3 relay → session acts → completion

Machine: cooper. Date: 2026-08-14. Live test of the full session↔operator↔session
loop for a delegated lane, using the leanest working `claude -p` SendMessage
relay (config G, see `gcn50-q15-lean-sendmessage.md`).

Fixed coordinates:
- Lane peer session: **`tars-gcn50-e2e`** (ref `63417c`), Orca worktree
  `/home/gaetan/orca/workspaces/e2e-scratch-repo/gcn50-e2e`, branch
  `GaetanCathelain/gcn50-e2e`, scratch repo `e2e-scratch-repo`.
- Slack DM channel `D0BBYNM01BL`, lane thread anchor ts `1786721723.183519`.
- Briefed session behaviour: wait for a peer message carrying the language,
  write `greeting.txt` in it (English ⇒ `Hello!`), then post
  `[lane:gcn50-e2e] done — wrote greeting.txt in <LANG>` into the thread.

## Phase 1 — spawn + Wire-2 question + operator answer (recap, verified live)

Established/verified before this phase (context from the orchestrator):
- Session spawned into the worktree and registered as a cross-session peer.
- Wire-2 (session → operator) question posted to the DM thread at
  `1786721968.195059` (17:39:28 CEST): `[E2E gcn50] Should greeting.txt be
  written in FRENCH or ENGLISH?`
- Operator (Gaetan) answered **ENGLISH** — thread reply `1786721817.388129`
  (17:36:57 CEST), i.e. *before* the session's question even posted.

## Phase 2 — Wire-3 relay + receipt

### Pre-relay baseline (proves causation)
At 15:43 UTC the worktree held only `.git`, `README.md` — **`greeting.txt`
ABSENT**. Any later `greeting.txt` is caused by this relay, not a stale artifact.

### Step 1 — confirm peer name/status (ListAgents via throwaway `claude -p`)
Lean throwaway (`--model haiku --tools "ListAgents"`, empty MCP, scratch cwd)
returned exactly one match:
```json
[{"name":"tars-gcn50-e2e","status":"idle","ref":"63417c","model":null}]
```
No collision-variant rename; session listed and **idle/waiting**. Confirmed.

### Step 2 — Wire-3 relay (lean SendMessage, config G)
Command (run from a no-CLAUDE.md scratch cwd; `empty-mcp.json` = `{"mcpServers":{}}`):
```sh
/usr/bin/time -v /home/gaetan/.local/bin/claude -p \
  'Send exactly one message via the SendMessage tool. to: "tars-gcn50-e2e". message: "[E2E gcn50 nonce=cbf75c02] operator answer: ENGLISH — write greeting.txt now and post completion." Then stop.' \
  --output-format json --dangerously-skip-permissions --model haiku \
  --strict-mcp-config --mcp-config <empty-mcp.json> \
  --system-prompt "You are a one-shot cross-session message relay. The target peer is pinned: 'tars-gcn50-e2e'. Call SendMessage exactly once with that recipient and the user's message verbatim, then stop. Do not add commentary." \
  --setting-sources "" --tools "SendMessage"
```
- Nonce: `cbf75c02`.
- Relay-send epoch: `1786722283.104419` (15:44:43.10 UTC / 17:44:43 CEST).
- Envelope `result`: `"Message sent successfully to tars-gcn50-e2e."`  `is_error: false`, `num_turns: 3`, exit 0.
- Tokens: input 6278, cache_creation 0, cache_read 0, output 366 → **TOTAL prompt 6278 tokens** (cost $0.00878).
- Wall time: **6.83s** (self-reported `duration_ms` 5266 / `duration_api_ms` 6112).

Caveat: `--output-format json` surfaces only the model's summary string, not the
raw SendMessage tool envelope, so a machine `msg_id` is not exposed by this lean
config. Delivery is instead proven receiver-side (below) — stronger than an
envelope field.

> First attempt mis-fired: `/usr/bin/time -v command claude …` tried to exec the
> shell builtin `command` (rc 127, 0.00s) — no message sent. Corrected by timing
> the real binary `/home/gaetan/.local/bin/claude` directly. Noted so the recipe
> is right: don't wrap a shell builtin in `/usr/bin/time`.

### Step 3 — receipt (the point of this phase)
a. **Session woke and acted.** `greeting.txt` present in the worktree,
   mtime `1786722295.493623` (15:44:55.49 UTC), contents exactly:
   ```
   Hello!
   ```
   English, **not** French (`Bonjour!`). ✓
b. **Completion line landed** in the DM thread (application-level ACK closing
   Wire-3) — `slack_read_thread` on `1786721723.183519`, reply 4/4 from Tars
   (U0BBH85NAKH), ts `1786722299.603149` (17:44:59 CEST):
   `[lane:gcn50-e2e] done — wrote greeting.txt in ENGLISH` ✓
c. **Latencies** (from relay-send `1786722283.10`):
   - relay-send → greeting.txt written: **12.4 s**
   - relay-send → completion post (receipt-latency): **16.5 s**

### Step 4 — ordering-race finding (unbuilt Wire-4 inbound relay)
Thread order (all CEST):
| ts | from | text |
|---|---|---|
| 17:35:23 `…723.183519` | Tars | `[E2E gcn50] Lane thread — greeting.txt task. (probe)` (parent) |
| 17:36:57 `…817.388129` | Gaetan | `English` |
| 17:37:07 `…827.817779` | Tars | `English it is.` |
| 17:39:28 `…968.195059` | Tars | `[E2E gcn50] Should greeting.txt be written in FRENCH or ENGLISH?` (session's Wire-2 Q) |
| 17:44:59 `…299.603149` | Tars | `[lane:gcn50-e2e] done — wrote greeting.txt in ENGLISH` |

Gaetan pre-replied **"English" at 17:36:57 — before** the session's own question
posted (17:39:28). The Hermes gateway woke on Gaetan's reply and merely
**chatted back "English it is." (17:37:07)** to the operator; it did **not**
relay the answer inward to the waiting session. That inbound (operator → session)
hop — **Wire-4** — is **not built** (spec marks it PROPOSED). This phase closed
the loop by **hand-running Wire-4**: a human/orchestrator-driven lean `claude -p`
SendMessage (Step 2) carried the answer to the session. Expected gap, cleanly
demonstrated.

## OVERALL E2E VERDICT: **E2E-ROUNDTRIP-PROVEN**

spawn → Wire-2 question (session→operator, Slack) → operator answer → Wire-3
relay (orchestrator→session, lean `claude -p` SendMessage) → session acted
(`greeting.txt` = `Hello!`, English) → completion posted
(`[lane:gcn50-e2e] done … ENGLISH`) — full round-trip observed end to end, with
the single known caveat that the **Wire-4 inbound (operator→session) relay was
hand-run, not built** (PROPOSED in spec). Wire-3 relay cost: 6278 tokens /
6.83 s; receipt-latency 16.5 s. Causation locked by the pre-relay
`greeting.txt`-absent baseline. Session/worktree/scratch repo left standing for
cleanup coordination.
