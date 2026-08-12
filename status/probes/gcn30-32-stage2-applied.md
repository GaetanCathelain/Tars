# GCN-30 + GCN-32 — STAGE 2 APPLIED to the live Tars VM (2026-08-12, UTC)

Executed per `status/probes/gcn30-32-checker-fix.md` §9 (apply protocol) and §8 (acceptance
probes P1–P5). The R2 gate (§9.2) was lifted before this run: `status/probes/gcn7-r2-run.md`
records R2 complete and green.

Deviation from the letter of §9: the protocol's paths name the worktree
`/home/gaetan/dev/orca-worktrees/Tars/gcn30-32-checker-fix`; this apply ran from
`/home/gaetan/dev/orca-worktrees/Tars/improvements` — same repo, same files, same
`origin/main`. Nothing else in §9 was altered. `status/staged/…` and
`status/probes/gcn30-32-checker-fix.md` were already committed on `origin/main`, so §9.6's
`git add` of those two was a no-op here; only the mirror moved.

## 1. §9.1 PRE-APPLY GUARD — PASS

```
$ ssh gaetan@192.168.0.9 'md5sum …/SKILL.md; wc -l …/SKILL.md; grep -c "^version: 1.7.2" …/SKILL.md; date -u'
1caff77d5dd4597756dc0e5b4a54a39c  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
716 /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
1
Wed Aug 12 10:43:50 AM UTC 2026
```

Stage 1 exactly (`1caff77d…`, 716 lines, `version: 1.7.2`). Not `a117cf49…`, not an
unexplained third hash — no Tars self-edit to reconcile. Proceeded.

## 2. §9.2 R2-completion gate — PASS

```
$ ssh gaetan@192.168.0.9 python3 -  (executions.db, read-only URI)
('759e08c598e3', 'direct', 'completed', '2026-08-12T12:34:05.334764+02:00', '2026-08-12T12:38:45.643795+02:00')
('759e08c598e3', 'direct', 'completed', '2026-08-12T12:31:45.054367+02:00', '2026-08-12T12:32:56.121549+02:00')
('62e8cd9db637', 'builtin', 'completed', '2026-08-12T12:30:23.611936+02:00', '2026-08-12T12:33:44.837880+02:00')
('759e08c598e3', 'direct', 'completed', '2026-08-12T12:25:28.465532+02:00', '2026-08-12T12:30:41.677860+02:00')
('3cee7db96367', 'builtin', 'completed', '2026-08-12T12:25:23.034616+02:00', '2026-08-12T12:25:23.116763+02:00')
```

R2's controlled runs (`759e08c598e3`, 10:25/10:31/10:34 UTC) all `completed` with non-null
`finished_at`; `status/probes/gcn7-r2-run.md` carries the verdicts (2a/2b/2c PASS, GCN-25
untouched, gateway unchanged). No lock held (`~/.hermes/state/engagement-checker.lock`
absent) and no run in flight at apply time (next natural fire 11:00 UTC).

## 3. §9.3 backup — dated, `cp -a`

```
-rw------- 1 gaetan gaetan 71567 Aug 11 20:28 SKILL.md.bak-gcn30-32-stage2-20260812T104444Z
$ md5sum …/SKILL.md.bak-gcn30-32-stage2-20260812T104444Z
1caff77d5dd4597756dc0e5b4a54a39c
```

This backup is **stage 1**. The pre-fix base remains
`SKILL.md.bak-gcn30-32-stage1-20260811T202839Z` (`a117cf49…`) — two different rollbacks,
§9.7 vs §5.6.

## 4. §9.4 transfer + swap

```
$ cat status/staged/gcn30-32-engagement-checker-SKILL.md | ssh … 'umask 077; cat > …/SKILL.md.new'
$ ssh … 'md5sum …/SKILL.md.new; wc -l …/SKILL.md.new'
d0adb8917fa7d793465bbf9a96b6bb87  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md.new
718 /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md.new

$ ssh … 'flock ~/.hermes/.wf3.lock -c "mv …/SKILL.md.new …/SKILL.md"'
mv rc=0
Wed Aug 12 10:44:54 AM UTC 2026
```

`.new` md5-verified **before** the `mv`, swap under `~/.hermes/.wf3.lock`, no heredoc.

## 5. §9.5 post-swap verification — PASS

```
d0adb8917fa7d793465bbf9a96b6bb87  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
718 /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
1                     <- grep -c "^version: 1.8.0"
0                     <- count of *.new leftovers in the directory
600 /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
d0adb8917fa7d793465bbf9a96b6bb87  …/status/staged/gcn30-32-engagement-checker-SKILL.md
```

Mode `600` preserved, no stray `.new`.

Live diff stage 1 → stage 2 (on the VM, backup vs live — 16 changed lines, the four §6.3
hunks and the version bump):

```
4c4
< version: 1.7.2
---
> version: 1.8.0
521,522c521,522   (terminal-state set: Linear terminal resolved per statusType, duplicate included)
533,534c533,534   (close-back guard: skip when terminal, and re-read must confirm not terminal)
668a669,670       (provenance: a rule about the value, binding every site that records linear_issue)
675c677 / 703c705 (coverage-break + [SILENT] wording carried over)
```

## 6. Gateway — NOT restarted

```
before apply (10:43 UTC): ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
after  apply (10:45 UTC): ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

Identical, and identical to the value R2 recorded. Hermes live-reloads the skill from disk;
nothing was restarted. (`NRestarts` is deliberately not cited — it is known to miss restarts.)

## 7. §9.6 mirror commit

```
$ cp status/staged/gcn30-32-engagement-checker-SKILL.md skills/orchestration/engagement-checker/SKILL.md
d0adb8917fa7d793465bbf9a96b6bb87  skills/orchestration/engagement-checker/SKILL.md
$ git commit … && git pull --rebase origin main && git push origin HEAD:main
   105c000..02d9c96  HEAD -> main
02d9c96 fix(engagement-checker): stage 2 semantics — statusType duplicate terminal,
        close-back guard, linear_issue provenance (GCN-30, GCN-32)
```

## 8. Acceptance probes (§8)

**Which run they score.** §8's P2–P5 were written for the R2 run (stage 2 absent). R2 is
already scored in `status/probes/gcn7-r2-run.md`. §9.8 additionally requires P2–P5 on **the
first in-window run after the apply** — so that is what is scored below: the natural cron
fire of job `62e8cd9db637`, **11:00:24 → 11:03:02 UTC**, session
`cron_62e8cd9db637_20260812_130024` (the id embeds Paris local time), the first run ever to
execute version 1.8.0. No run was triggered by hand (`hermes cron run` is out of scope for
this session); the fire was waited for. The apply landed 10:44:54 UTC, ~15 min before it —
well past the ~30 s live-reload window.

```
SESSION=cron_62e8cd9db637_20260812_130024
STAGE 2 APPLIED AT RUN TIME?  yes

P1 byte-equality                 PASS   (d0adb8917fa7d793465bbf9a96b6bb87, 718 lines)
P2 GCN-25 still duplicate        PASS   <- most important
P2b item dismissed + closed_at   NOT-REACHED (item was already terminal `done`; see below)
P3 F2 provenance                 PASS
P4 F1 gate citation              PASS   (run was `[SILENT]`; see the guard note)
P5 run health                    PASS
P5b Linear cursor moved past 2026-08-11T14:30:26Z   YES (now 2026-08-12T11:00:37Z)
```

### P1 — byte equality live == staged == repo == origin/main — PASS

```
live   d0adb8917fa7d793465bbf9a96b6bb87
d0adb8917fa7d793465bbf9a96b6bb87  …/status/staged/gcn30-32-engagement-checker-SKILL.md
d0adb8917fa7d793465bbf9a96b6bb87  …/skills/orchestration/engagement-checker/SKILL.md
origin/main  d0adb8917fa7d793465bbf9a96b6bb87   (718 lines)
```

Four-way equality (§8's three plus `origin/main` read through `git show`), live file 718
lines, `version: 1.8.0`, mode 600. Re-checked after the 11:00 run: still
`d0adb8917fa7d793465bbf9a96b6bb87`, no `.new` leftover.

### P2 — GCN-25 unchanged — PASS

`mcp__claude_ai_Linear__get_issue(id="GCN-25")`, read twice: 10:45 UTC (post-apply,
pre-run) and 11:05 UTC (post-run). Byte-identical both times, and identical to §8's pre-R2
baseline:

```
statusType  = "duplicate"                                   ✓
status      = "Duplicate"                                   ✓
stateHistory length = 2, last .state.id = 0ea8bdc6-215f-48d3-b228-779422c6b03e, endedAt null  ✓
completedAt = null                                          ✓
canceledAt  = 2026-08-11T15:10:41.624Z    updatedAt = 2026-08-11T18:08:40.395Z   (both unmoved)
```

No third `stateHistory` entry, no `completedAt`. **No wrong write.** §9.9's hand-revert is
not needed and was not run.

### P2b — the item side — NOT-REACHED (expected)

```json
// item slack:CC397R0HY:1786458544.180249 — identical before and after the run
{ "status": "done",
  "linear_issue": { "id": "GCN-25", "priority": 3, "closed_at": "2026-08-11T17:10:41+02:00" },
  "last_reason": "slack: Gaetan marked the engagement-checker item duplicate of the Help Tech investigation and closed the request" }
```

`NOT-REACHED`, per §8's own framing ("zero non-terminal items are linked to a
`duplicate`-statusType issue" — the item is already terminal `done`, so stage 2's
`duplicate → dismissed/linear:duplicate` transition has nothing to move). It is **not** the
FAIL shape: `status` is not `dismissed`, and in particular not `dismissed` with reason
`linear:canceled` (the `canceledAt` trap of §4 did not fire).

**One deviation from §8's written NOT-REACHED shape, worth recording:** §8 describes
NOT-REACHED as "still `done` **with no** `closed_at`", but `linear_issue.closed_at` is now
set — `2026-08-11T17:10:41+02:00`, i.e. GCN-25's own `canceledAt` (15:10:41Z) in Paris
local. It was absent in the 2026-08-11 measurement, so it was latched by one of the R2-day
runs (stage 1's §4 collector reads Linear directly and records the close time). Consequence:
§9.9's second half — "clear the wrongly-latched `linear_issue.closed_at` … or the close-back
can never run for it again" — now applies to this item on its own merits, independently of
any corruption. Flagged for the coordinator; **not** touched here (writing
`~/.hermes/state/engagement-checker.json` is outside this session's mandate).

### P3 — F2 regression: no `linear_issue` change without a Linear tool call — PASS

Snapshots pulled read-only either side of the run (guards ran, neither was empty):

```
ec-before.json items=55 bytes=121769   (10:45 UTC, post-apply, pre-run)
ec-after.json  items=55 bytes=122323   (11:05 UTC, post-run)
```

Items whose `linear_issue` changed:

```
slack:GQ07CQXT7:1786012625.976839   {"id":"GCN-12","priority":3}
slack:C0BQHP903LY:1786524823.516609 {"id":"GCN-38","priority":2,"closed_at":"2026-08-12T13:02:42.050+02:00"}
```

Linear calls in the same session:

```
SESSION=cron_62e8cd9db637_20260812_130024
linear_tool_calls 7
```

**PASS** — every changed item is accompanied by `linear_tool_calls >= 1` (7 ≫ 1). The calls
are real MCP calls, not a bare count: 3 × `mcp__linear__get_issue` and
1 × `mcp__linear__save_issue` are visible in `tool_calls`, plus 3 rows the regex matched on
`tool_name`. Both changes are explicable and provenance-backed:

- **GCN-12** — its `closed_at` (set by R2 at 10:31:34.447Z) was **cleared** and the item went
  `done → open` with reason `linear:reopened` at 13:00:37+02:00. That is the coordinator's
  post-R2 revert of GCN-12 to Todo (R2 §0 asked for it) being reconciled correctly.
- **GCN-38** — item went `open → done`, `closed_at` set to `2026-08-12T13:02:42.050+02:00`,
  reason `slack:resolved — Gaetan supplied the sandbox/development endpoints…`; the
  in-session `mcp__linear__save_issue` is the matching close of an issue this same skill
  filed at 09:00 UTC today (Permitted write #3). This exercised the **stage-2 close-back
  guard** path in production on the first run, without incident.

Reference contrast: the F2 failure session `20260811_180849_14f453` scored
`linear_tool_calls == 0` while writing `GCN-26`. (§8's stated limit still holds: P3 is a
session-level proxy, not per-key attribution.)

### P4 — F1 regression: no gate citation absent from the live files — PASS

The run's final assistant message is exactly `[SILENT]` — 14 assistant rows, 8 content bytes
total, one non-empty:

```
('assistant', 8, '[SILENT]')
```

```
check 1: grep -nE '09:00[^0-9]{0,3}19:00|quiet hour|working hour|business hour'  -> no output (rc=1)  PASS
check 2: clock ranges cited by the run -> none extracted; no range with live_hits=0            PASS
```

The run cites no window at all, so there is no invented gate; the measured F1 sentence does
not recur. **PASS.**

**Guard note (a real defect in §8's P4, not in the fix):** the mandated guard
`[ "$(wc -c < /tmp/run-text.txt)" -gt 20 ] || ABORT` **fires on a legitimately silent run** —
`[SILENT]` is 8 bytes plus a newline. Run verbatim, P4 aborts with "no assistant text" on
exactly the outcome §8 calls "the account of the silence". The abort here was a false
negative from the guard, resolved by reading the extraction directly (shown above), not by
skipping the guard. Suggested correction for a future edit of §8: keep the guard but treat a
whole-body `[SILENT]` as a valid, passing extraction.

Cross-check, R2's Run 2 session `cron_759e08c598e3_20260812_123405` (stage 2 absent):
225 chars of assistant text, check 1 no match, zero clock ranges → P4 also PASS there.
Reference live-file hit for a real range: `10:00–17:00` → `live_hits=2` (with
`--exclude='*.bak*'`; the six `SKILL.md.bak-*` copies are correctly excluded).

### P5 — run health — PASS

```
$ stat -c '%y %s' ~/.hermes/state/engagement-checker.json
2026-08-12 11:02:58.729596884 +0000 122323          <- flushed during the run (11:00:24–11:03:02)

$ jq '{items, nonterminal}'
{ "items": 55, "nonterminal": 10 }                  <- valid JSON, count intact (was 50 on 08-11, 55 pre-run)

$ grep -F '[cron_62e8cd9db637_20260812_130024]' ~/.hermes/logs/agent.log | grep -cE 'ERROR|Traceback'
0                                                   <- zero errors attributable to this run
```

**PASS** on all three. `executions.db` records the run `completed`,
`13:00:24.149116+02:00 → 13:03:02.944266+02:00`.

### P5b — Linear cursor — YES

```
pre-R2 (2026-08-11):        {"cursor":"2026-08-11T14:30:26Z", …}
pre-run today (10:45 UTC):  {"cursor":"2026-08-12T10:34:27Z","last_success":"2026-08-12T12:34:27+02:00"}
post-run     (11:05 UTC):   {"cursor":"2026-08-12T11:00:37Z","last_success":"2026-08-12T13:00:37+02:00"}
```

Cursor advanced again on the first stage-2 run. Stage 1's collector keeps working under 1.8.0.

## 9. F2 regression check — PASS

P3 **is** the F2 regression check the document defines ("§8 P3 — F2 regression: no
`linear_issue` change without a Linear tool call in the same session"). Verdict above:
**PASS**, on a run that did change two `linear_issue` values, with 7 in-session Linear calls
including a real `mcp__linear__get_issue`/`save_issue` pair. The underlying F2 defect
(§2.3) — a `linear_issue` written from a fabricated key with zero Linear calls — did not
recur.

## 10. Compliance and what was NOT touched

- No `config.yaml`, `.env` or `SOUL.md` edit. No skill file other than
  `skills/orchestration/engagement-checker/SKILL.md` (repo) and its VM mirror.
- No `sops -d` in any form; no secret read, echoed or written.
- No `hermes chat`, no `hermes cron run`, no `hermes cron create`, no gateway restart —
  the scored run is a natural cron fire, waited for.
- No Linear ticket created. One Linear **read** (`get_issue GCN-25`, twice) — no write.
- No Slack post. `status/lane-a.md` and `MIGHT-DO.md` untouched.
- `~/.hermes/state/engagement-checker.json` not written by this session (only read).

## 11. Open items for the coordinator

1. **GCN-25's item carries a latched `linear_issue.closed_at`** (§8/P2b above). Until it is
   cleared, the stage-2 `duplicate → dismissed` transition can never fire for
   `slack:CC397R0HY:1786458544.180249`. §9.9's second half describes the clearing; it is a
   state-file write and was deliberately not performed here.
2. **§8's P4 guard aborts on a `[SILENT]` run** (details above) — a documentation fix, not a
   code fix.
3. **§9.8 ticket closure** (GCN-30, GCN-32 → Done, citing 1.8.0 and md5
   `d0adb8917fa7d793465bbf9a96b6bb87`) is the coordinator's call, not this session's; no
   ticket was moved.
4. GCN-12 is back to non-terminal (`linear:reopened` at 11:00:37 UTC) — the R2 revert landed
   and reconciled cleanly. Recorded here only so it is not mistaken for drift.
