# GCN-7 — WF6 R2, engagement-checker in-window measurement (2026-08-12 morning UTC)

Executed per `status/probes/gcn7-wf6-e2e.md` §9.2 (authoritative protocol), with the
2026-08-12 morning context supplied at dispatch. `status/probes/gcn30-32-checker-fix.md`
§8 (P1–P5 acceptance probe) supplied the exact SESSION-derivation query and the GCN-25
pre-R2 baseline used below.

## 0. Scope and what changed vs. the written protocol

- All timestamps below are UTC, taken from `agent.log` (established UTC in prior probes)
  or from Linear's own ISO timestamps (also UTC, `Z` suffix).
- **Deviation from §9.2 step 5 (restore):** the written protocol says to restore the
  issue closed for the 2c test. This run's dispatch prompt drops that step and its hard
  rules cap Linear writes at "only the GCN-12/GCN-24 close and whatever the checker
  itself files" — no second write is listed. Read together with the GCN-25 instruction
  ("do NOT prevent or revert it yourself; the coordinator reverts after scoring"), the
  intent this round is: leave state as measured, coordinator reverts after scoring.
  **GCN-12 was left `Done`, not restored to `Todo`.** Coordinator: revert GCN-12 to Todo
  (state id `59ed732a-f242-4eba-926f-1c0d128fe83c`) after scoring, mirroring what §9.2
  step 5 would have done inline.
- The written protocol expected one controlled `hermes cron run` pair. In practice a
  natural cron overlap forced a third invocation (§3 below) — recorded in full, not
  hidden.

## 1. Survey of today's natural fires, 08:00–10:00 UTC

Job `62e8cd9db637` ("Gaetan engagement checker", `*/30 10-16 Paris` = `08:00-14:00 UTC`)
fired 5 times before the controlled window. Read from `~/.hermes/cron/executions.db`
and `agent.log` session `cron_62e8cd9db637_20260812_<HHMMSS>`:

| UTC fire | session | outcome | new GCN issue(s) filed |
|---|---|---|---|
| 08:00:20 | `..._100020` | not silent, digest + filings | GCN-33, GCN-34, GCN-35 |
| 08:30:20 | `..._103020` | not silent, digest + filing | GCN-37 |
| 09:00:21 | `..._110021` | not silent, digest + filing | GCN-38 |
| 09:30:21 | `..._113021` | not silent, digest + filing | GCN-39 |
| 10:00:22 | `..._120022` | reconcile only, 0 Linear calls | none |

None of the 5 returned bare `[SILENT]`; all delivered scored reminders (expected — see
§9.2's note that in-window items score ~29 points lower than evening, and several of
today's items scored high on urgency/stakeholder terms regardless of window). Every
session showed exactly one `BLOCKED` line for `execute_code` (not for `terminal`):

```
Tool execute_code returned error (0.00s): {"status": "error", "error": "BLOCKED:
execute_code runs arbitrary local Python (including subprocess calls that bypass
shell-string approval checks). Cron jobs run without a user present to approve it.
Use normal tools instead, or set approvals.cron_mode: approve only if this cron
profile is intentionally trusted."}
```

This is a **different, permanent** block from the §5.2 chat-TTY gate documented in
gcn7-wf6-e2e.md — it fires under cron too, unconditionally, on the `execute_code` tool
specifically. Every session recovered by falling back to the `terminal` tool for the
same script, which succeeded. This fallback shape (`terminal` running
`python3 /tmp/engagement_linear_collector.py`) **is stage 1's §4 collector**, and it
worked cleanly every single time today — see §2.

## 2. Stage-1 first live exercise — CONFIRMED, repeatedly

`cron_62e8cd9db637_20260812_103020` (08:30 UTC), read from `lcm.db` (`tool_calls`
column), shows the exact collector invocation:

```
LINEAR_CURSOR='2026-08-12T08:01:06Z' LINEAR_RUN_END='2026-08-12T08:30:42Z' \
  python3 /tmp/engagement_linear_collector.py
```

— run via the `terminal` tool, succeeding (2.25s, 6301 chars output) immediately after
the `execute_code` BLOCKED line for the same underlying task. This is the venv-interpreter
collector shape §4 documents (stage 1, v1.7.2, live md5 `1caff77d5dd4597756dc0e5b4a54a39c`,
confirmed unchanged before and after this whole R2 — see §6).

The Linear cursor in `~/.hermes/state/engagement-checker.json` (`.sources.linear.cursor`)
advanced monotonically across every fire today:

| after fire | `.sources.linear.cursor` |
|---|---|
| pre-R2 baseline (gcn30-32 doc, 2026-08-11 evening) | `2026-08-11T14:30:26Z` |
| 08:30 UTC | `2026-08-12T08:30:42Z` |
| 09:00 UTC | `2026-08-12T09:00:36Z` |
| 10:00 UTC | `2026-08-12T10:00:53Z` |
| after controlled run 2 (10:34 UTC) | `2026-08-12T10:34:27Z` |

**P5b-equivalent: YES, cursor moved.** Stage 1's plumbing is not just live on disk — it
executed successfully on every single cron fire measured today, both natural and
controlled, with zero collector failures.

## 3. Controlled sequence

**Before-snapshot (10:24 UTC, pre Run 1):** board `list_issues(team=GCN, limit=250)` →
`hasNextPage:false`, 39 issues. GCN-25: `statusType=duplicate`, `status=Duplicate`,
`updatedAt=2026-08-11T18:08:40.395Z` (matches the gcn30-32 §8 P2 pre-R2 baseline exactly).
GCN-12: `Todo`/`unstarted`. GCN-24: `Todo`/`unstarted`. `agent.log` at 21117 lines.
State file copied (54 items, 119227 bytes). Gateway
`ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC`.

### Run 1 — `hermes cron run 759e08c598e3`, 10:25:27–10:30:42 UTC

Session `cron_759e08c598e3_20260812_122528`. `hermes cron run` reported `succeeded`.
Not `[SILENT]` — delivered a 4-item French-language digest citing existing tickets by
number (GCN-15, GCN-33, GCN-38, GCN-34) — **no new issue created** (`EC-EAF7BB`/GCN-34
was already filed by the 08:00 UTC natural run, matching the dispatch's fallback
instruction to score 2a off that natural evidence). The run also made one legitimate
self-close: `mcp__linear__save_issue(id=GCN-39, state=<Done>)`, closing an issue the
*same skill* had filed 55 minutes earlier (09:35 UTC) after finding Slack evidence the
item was resolved — a correct use of "Permitted write #3" (close an issue this skill
itself filed), not part of the 2c test but useful corroborating evidence the close-back
mechanism works in general. GCN-25 read immediately after: unchanged
(`statusType=duplicate`, `updatedAt=2026-08-11T18:08:40.395Z`). Gateway unchanged.

### Between runs — close GCN-12

`mcp__claude_ai_Linear__save_issue(id="GCN-12", state="0434e579-7b85-487a-8cf9-5aed6caaf41b")`
(the Done state id) at **2026-08-12T10:31:34.447Z**. Result: `status=Done`,
`statusType=completed`, `completedAt=2026-08-12T10:31:34.447Z`. GCN-24 left untouched
(the other candidate; GCN-12 was chosen). Pre-close local queue item state (source
`slack:GQ07CQXT7:1786012625.976839`): `status=open`, `linear_issue={id:GCN-12,priority:3}`,
no `closed_at`.

### Run 2, attempt 1 — lock collision, `[SILENT]` by lock contention, not by design

`hermes cron run 759e08c598e3` at 10:31:44 UTC hit
`~/.hermes/state/engagement-checker.lock`, still held by the **natural** `62e8cd9db637`
fire that started at 10:30:23 UTC (the */30 schedule collided with my window). Session
`cron_759e08c598e3_20260812_123145`: `terminal` returned
`{"output": "mkdir: cannot create directory '.../engagement-checker.lock': File exists", "exit_code": 1}`,
and the run correctly bailed with exactly `[SILENT]` (5 API calls, 0 Linear calls) rather
than clobbering the held lock. Waited for the natural fire to finish
(`executions.db` `status='completed'`, `finished_at=2026-08-12T10:33:44+02:00`... UTC
2026-08-12T08:33:44 — wait, corrected: local finished_at `12:33:44+02:00` = **10:33:44
UTC**), confirmed the lock directory was gone, then retried.

### Run 2, attempt 2 (the real Run 2) — `hermes cron run 759e08c598e3`, 10:34:04–10:38:46 UTC

Session `cron_759e08c598e3_20260812_123405`. `hermes cron run` reported `succeeded`.
The §4 collector ran again
(`LINEAR_CURSOR='2026-08-12T10:30:41Z' LINEAR_RUN_END='2026-08-12T10:34:27Z' python3 /tmp/engagement_linear_collector.py`,
terminal, 2.84s/14385 chars) and **fetched GCN-12's post-close state directly**: its
JSON payload includes `"state":{"id":"0434e579-...","name":"Done","type":"completed"}`,
`"completedAt":"2026-08-12T10:31:34.447Z"`, plus the issue's comment history (the
2026-08-10 close/reopen note). The reconcile step then wrote (via `jq`, atomic
tmp+mv under the `.wf3.lock`-equivalent discipline, not a heredoc) to
`~/.hermes/state/engagement-checker.json`:

```
.items["slack:GQ07CQXT7:1786012625.976839"].status = "done"
.items[...].linear_issue.closed_at = "2026-08-12T10:31:34.447Z"
.items[...].status_history += {"at":"2026-08-12T12:34:27+02:00","status":"done","reason":"linear:completed"}
```

Verified read back from the live file after the run (not from the log — ground truth):

```json
{
  "key": "slack:GQ07CQXT7:1786012625.976839",
  "status": "done",
  "linear_issue": { "id": "GCN-12", "priority": 3, "closed_at": "2026-08-12T10:31:34.447Z" },
  "status_history": [
    { "at": "2026-08-10T16:14:10+02:00", "status": "open", "reason": "Explicit promise retained, but reminder suppressed because the requester said it was not urgent and to take his time" },
    { "at": "2026-08-10T22:33:57.110+02:00", "status": "done", "reason": "linear:completed" },
    { "at": "2026-08-10T23:19:51.823+02:00", "status": "open", "reason": "linear:reopened" },
    { "at": "2026-08-12T12:34:27+02:00", "status": "done", "reason": "linear:completed" }
  ]
}
```

This run also observed (via the collector, no `mcp__linear__` MCP call in-session — the
collector reads Linear directly over its own API key, which is the whole point of §4)
that item `slack:C0BHFEARGTG:1786530588.310929` already carried a Linear link to
**GCN-40**, filed by the natural 10:30 UTC fire moments earlier (`createdAt
2026-08-12T10:33:26.472Z`, verified via `get_issue`). Run 2 only synced its local
score/reminder fields for that already-linked item — it made **zero**
`mcp__linear__save_issue` calls, i.e. it did not re-file it. Final delivered message was
a single new-item reminder for GCN-40 (score, not a re-filing).

## 4. Per-criterion verdict

**2a — engagement loop files a GCN issue: PASS (natural evidence, per dispatch's
fallback instruction).**
`EC-EAF7BB` (email source `19ff1aab2d50767d`, "Review Vercel usage after the on-demand
budget reached 100%") was filed as **GCN-34** by the natural 08:00 UTC fire before the
controlled window opened — confirmed both from the live state file
(`.items["email:19ff1aab2d50767d"].linear_issue.id == "GCN-34"`) and from Linear
(`GCN-34` exists, `Todo`/`unstarted`, labels `["investigation"]`, priority High,
`createdAt 2026-08-12T08:07:49.062Z`). The `mcp__linear__save_issue` call and result
were captured directly from `lcm.db`: description line 1 `Source: email:19ff1aab2d50767d`,
line 2 `Evidence: [...]`, matching §7's filing shape. Five more natural filings today
(GCN-33, 35, 37, 38, 39) show the same shape and are corroborating, not load-bearing,
evidence.

**2b — no re-report of an already-filed item: PASS.**
Across 7 fires today (5 natural + 2 controlled), no item with an existing
`linear_issue` link was re-filed. The one case that could have looked like a violation
(GCN-40, linked without an in-session `mcp__linear__` call in Run 2) is not one: GCN-40
was filed by the natural fire immediately prior, and Run 2's collector — reading Linear
directly, which is precisely stage 1's mechanism — correctly recognized the existing
link and only refreshed local score/reminder fields, making zero create calls.

**2c — the closed issue reconciles: PASS.**
GCN-12, closed to Done at `2026-08-12T10:31:34.447Z`, reconciled on the very next run
(the one that hit it): its queue item moved to `status: "done"`,
`status_history` gained `{"reason": "linear:completed"}` (not `slack:...`), and
`linear_issue.closed_at` was set to the exact Linear `completedAt` timestamp. This is
the literal PASS shape from §9.2 step 4 and from gcn30-32-checker-fix.md §8 P2's
secondary assertion.

## 5. GCN-25 — the wrong-write regression check

**Unchanged across the entire R2 window.** Read 3 times (before Run 1, after Run 1,
after Run 2/attempt-2): `statusType="duplicate"`, `status="Duplicate"`,
`updatedAt="2026-08-11T18:08:40.395Z"` every time — byte-identical to the
gcn30-32-checker-fix.md §8 P2 pre-R2 baseline. **No wrong write occurred.** This is
consistent with stage 1 being plumbing-only (§4 collector shape) and NOT touching the
§5/§7c reconcile semantics that stage 2 (not yet applied) is meant to fix — R2 ran with
stage 2 absent, exactly as the coordinator's protocol expected, and the known hole
(statusType=duplicate not mapped to terminal) simply never got exercised today because
no non-terminal item pointed at a `duplicate`-statusType issue during the window (GCN-25
itself has no non-terminal item attached, consistent with gcn30-32 §8's "zero live queue
items move from the duplicate fix" framing).

## 6. Gateway health

`ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC` — checked before Run 1 and after the
final run. **Unchanged.** No gateway restart occurred.

## 7. VM write audit (hard-rules compliance)

- Skill file: `md5sum` unchanged, `1caff77d5dd4597756dc0e5b4a54a39c`, 716 lines, both
  before and after — no skill/config/SOUL edit made.
- The only VM-side triggers I issued were three `hermes cron run 759e08c598e3`
  invocations (one lock-blocked no-op, one real, plus the required Run 1) — no direct
  file writes, no `sops -d`, no secrets read.
- Linear writes I made directly: one (`GCN-12` → Done). Left in that state per §0's
  reasoning — **not restored**, flagged for the coordinator.
- No new Linear tickets created by me; GCN-33/34/35/37/38/39/40 were all filed by the
  checker itself (natural and controlled runs), not by me.
- No Slack posting by me.

## 8. Surprises

- **Natural cron traffic during the controlled window is real and collided once.** The
  */30 job and the "final pass" job are independent cron entries that can and did
  overlap (10:30 UTC natural fire vs. my 10:31:44 UTC controlled trigger), producing a
  legitimate lock-busy `[SILENT]` rather than a bug. Anyone re-running this protocol
  close to a `:00`/`:30` boundary should expect the same and simply retry after the
  natural fire's `executions.db` row shows `completed`.
- **`execute_code` is blocked unconditionally under cron**, not just under the
  chat-TTY path documented in gcn7-wf6-e2e.md §5.2 — every session today hit it once and
  recovered via `terminal`. Worth folding into that doc's residual-1 write-up if anyone
  revisits it.
- **GCN-37 self-closed itself outside any run I triggered** (created 08:35:16 UTC by the
  natural 08:30 fire, `Done` by 08:39:49 UTC — after that cron session had already ended
  per `executions.db`). Not investigated further; out of scope for R2, flagged here only
  so it isn't mistaken for stray R2 evidence.
- Today's natural traffic alone (6 filings before I ever touched the VM) meant the
  "unfiled 2a candidate" framing in the dispatch was already stale by the time I started
  — the dispatch anticipated this explicitly ("unless a natural run already filed it").
