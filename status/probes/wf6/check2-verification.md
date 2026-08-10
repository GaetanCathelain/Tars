# WF6 — engagement-checker v1.7.0 deploy + spec check 2 end-to-end

Session: Claude Code (Orca subagent), cooper worktree `wf6-vm-skills`.
Window: 2026-08-10 20:19Z → 20:38Z. All VM timestamps UTC unless a `+02:00`
offset is shown (Hermes writes state in Europe/Paris).

**Headline: (a) PASS, (b) PASS, (c) PASS, (d) PASS.** `GCN-12` was created by
cycle 1, correct on every field, not duplicated by cycle 2, and cleared from
the queue by cycle 3 after being closed in Linear. Nothing was posted to Slack.

**Two things the coordinator must decide on** — neither is a regression, both
are outside what this probe was scoped to do:

1. **Tars rewrote the deployed SKILL.md mid-run** (cycle 1, `skill_manage` +
   self-merged PR #41) and the VM file therefore no longer matches the repo
   working-tree file byte-for-byte. Diff is one paragraph, detailed below.
2. **`GCN-12` is a real loop, not test data.** It was filed from a genuine
   Slack commitment and then closed by step (d). See "Cleanup" at the foot.

---

## Phase 1 — deploy

### Source of truth

Repo working tree (uncommitted at the time of deploy):
`/home/gaetan/dev/orca-worktrees/Tars/wf6-vm-skills/skills/engagement-checker/SKILL.md`

```
9834b29042996a7a707f332dcddc14719f0c3d2d02d41f8645b3fc3b452c3310  skills/engagement-checker/SKILL.md
67416 bytes   version: 1.7.0
```

### Pre-deploy state on the VM

```
64e132f46c5ff5dc34329d8179c1d0ee55562d5701b90ad7cd8b27553af553a0  ~/.hermes/skills/orchestration/engagement-checker/SKILL.md
65404 bytes   version: 1.6.0   (mtime Aug 10 19:30)
```

Gateway, read 20:19:37Z, **before** anything was touched:

```
NRestarts=0
ActiveState=active
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

(That 20:00:38Z restart is the *previous* session's `config.yaml` edit —
`config.yaml.bak-toolvis-20260810-195950` — not this probe's.)

### Step 1+2 — backups

Both taken with `cp -a` before any write, one shared timestamp
`20260810T202040Z`:

| what | path | sha256 at backup time |
|---|---|---|
| skill | `~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-wf6-ec170-20260810T202040Z` | `64e132f46c5ff5dc34329d8179c1d0ee55562d5701b90ad7cd8b27553af553a0` |
| state | `~/.hermes/state/engagement-checker.json.bak-wf6-ec170-20260810T202040Z` | `ca96214e90df2f5262350d392a1b536b9d2b68a3032fdcdc0c3fffacbb0709e5` |

Backup sha256 verified equal to the live file's sha256 at the moment of copy,
for both files.

### Step 3+4 — write under flock, verify sha256 both sides

`scp` to `/tmp/ec-170-src.md`, then:

```
flock ~/.hermes/.wf3.lock -c "cat /tmp/ec-170-src.md > ~/.hermes/skills/orchestration/engagement-checker/SKILL.md"
```

`cat >` into the existing inode, so mode `0600` and ownership are preserved
(`-rw------- gaetan gaetan`). Temp source removed afterwards.

Post-write on the VM:

```
9834b29042996a7a707f332dcddc14719f0c3d2d02d41f8645b3fc3b452c3310  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
-rw------- 1 gaetan gaetan 67416 Aug 10 20:20
version: 1.7.0
```

**sha256 identical on both sides at deploy: `9834b290…c3310`.** ✅

### Step 5 — live-reload verification

Gateway re-read at 20:21:0xZ (≈40 s after the write):

```
NRestarts=0
ActiveState=active
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

and again at 20:37:2xZ, after all three cycles:

```
NRestarts=0
ActiveState=active
ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

**The gateway did NOT restart.** `ActiveEnterTimestamp` is byte-identical
before the deploy, after the deploy, and after three agent cycles. (`NRestarts`
was 0 throughout and is, as the spec records, worthless as a test — it was
still 0 across the 20:00:38Z restart.)

`~/.local/bin/hermes skills list`:

```
│ engagement-checker   │ orchestration       │ local     │ local     │ enabled │
```

**Caveat, stated plainly: `hermes skills list` has no version column**, and
`hermes skills inspect engagement-checker` only resolves registry skills
(`Error: No skill named 'engagement-checker' found in any source.`). So "shows
v1.7.0" cannot be evidenced from the CLI. What *is* evidenced:

- the on-disk frontmatter is `version: 1.7.0` and the file's sha256 equals the
  repo file's;
- each cycle is a fresh `hermes chat` process that reads the skill from that
  path at start;
- **behaviourally**, cycle 1 created a GCN issue — which the deployed v1.6.0
  provably could not do across three consecutive live cycles.

---

## Phase 2 — spec check 2

`docs/specs/wf6-linear-integration.md` §Verification item 2.

### Method

The **real** cron prompt for job `62e8cd9db637` was extracted from
`~/.hermes/cron/jobs.json` (read-only, `json.load` + write of the `prompt`
field alone to `/tmp/ec-cron-prompt.txt`, 855 bytes,
sha256 `7264b6f7b583006740383b18d6369966db17e047b2263671181239f6fca1e4c6`):

> Run `engagement-checker` end to end. Use Europe/Paris and the fixed run
> timestamp. Apply the same workday gate as `daily-work-brief` first: weekday,
> official metropolitan-France holiday, explicit PayFit leave, then the bounded
> actual-work override for a nominal day off; return exactly `[SILENT]` when the
> gate says not to run. Otherwise acquire the single-writer lock; read durable
> state; collect and exactly filter only Slack, email, and Linear deltas since
> each source cursor; reconcile user decisions and resolved loops; update each
> source cursor only after successful processing; evaluate the pending queue
> with cooldowns; persist state; release the lock; then return the compact
> reminder or exactly `[SILENT]`. This is a read-only source workflow. Do not
> spawn or prompt Claude or Orca; Cooper/Orca may only be queried read-only for
> Linear evidence.

Exactly one paragraph was appended, as the delivery guard:

> OPERATOR OVERRIDE, THIS MANUAL RUN ONLY: do NOT post, send, deliver or
> schedule anything to Slack — no message, no DM, no thread reply, no reaction.
> Slack READS are fine; every Slack WRITE is forbidden this run. Instead, return
> the final delivery text (or exactly [SILENT]) to me here as your final answer.
> Everything else about the run is unchanged.

Combined file `/tmp/ec-run-prompt.txt`, 1210 bytes. Invocation, identical for
all three cycles, run detached under `nohup` on the VM:

```
hermes chat -q "$(cat /tmp/ec-run-prompt.txt)" -s engagement-checker \
  --provider openai-codex -m gpt-5.6-sol --yolo -Q
```

- `--provider openai-codex -m gpt-5.6-sol` are the job's own
  `provider_snapshot` / `model_snapshot`, so the cycles ran on the same model
  production runs on.
- `--yolo` — the pre-approved deviation. Recorded as used.
- `-Q` (quiet) is why the captured output is the final response only; that is
  precisely the text cron would have delivered.
- `hermes chat` has no `--deliver` target, so the cron delivery path
  (`slack:C0BP2GZUFSR:1786359613.979759`) is structurally absent from these
  runs. That is the first of the two independent Slack guarantees.

**Cron-prompt drift (pre-existing, unchanged by this probe).** The live prompt
above ends "This is a read-only source workflow" and says nothing about filing
GCN issues or the close-back; the prompt the skill prescribes
(§Scheduling contract) does both. The same drift was already logged in
`deploy-and-verify.md`. It did **not** block anything: the skill's
§Permitted writes is the authority and the run filed correctly. No cron job was
touched.

### Pre-run baselines

Durable state `~/.hermes/state/engagement-checker.json`, 79186 bytes,
sha256 `ca96214e90df2f5262350d392a1b536b9d2b68a3032fdcdc0c3fffacbb0709e5`:

```
last_completed_run 2026-08-10T21:54:41+02:00
cursors: slack/email/linear all 2026-08-10T21:54:41+02:00, all last_success equal
items: 30 total — Counter({'done': 25, 'dismissed': 3, 'snoozed': 1, 'open': 1})
filed_count 0          <- no item anywhere carried `linear_issue`
```

The single `open` item, i.e. the only §7-eligible candidate:

```
slack:GQ07CQXT7:1786012625.976839
  status open  src slack  kind "explicit promise"
  short_id EC-A44C  linear_issue null
  created 2026-08-07T10:10:37+02:00  due None  snooze None
  reminders 1  last_score 84
```

GCN board before cycle 1 — 11 issues, highest key **GCN-11**, non-terminal set
= GCN-10 (Todo), GCN-3/4/5/7 (In Progress), Backlog column empty. Read with
`mcp__linear__list_issues` team GCN limit 250, `hasNextPage: false`.

---

### (a) One cycle runs end to end — **PASS**

Cycle 1, launched 20:23:26Z, exited 20:28:00Z. Session
`20260810_202328_1e7f50`. Verbatim stdout (`/tmp/ec-cycle1.log`, 2987 bytes),
with the two long diff lines elided at the marked point only for width — the
full text of both is reproduced in "Skill drift" below:

```
  ⟳ compacting context…
  ⟳ compacting context…
  ┊ review diff
a/.hermes/skills/orchestration/engagement-checker/SKILL.md → b/.hermes/skills/orchestration/engagement-checker/SKILL.md
@@ -122,7 +122,7 @@
 1. `filter_users_from=U08BDJAMSRZ` for commitments, replies, completion signals, and deferral instructions authored by Gaetan.
 2. `filter_users_with=U08BDJAMSRZ` for DMs, threads, and messages involving him that may contain a direct ask or someone waiting.

-Discard every Slack message whose sender is not a human user — … [OLD paragraph, full text below]
+Discard every Slack message whose sender is not a human user — … [NEW paragraph, full text below]

 Filter every result by exact Slack `ts` against the cursor before classifying it. Deduplicate the two views by channel and timestamp. For a new candidate only, use `mcp__slack__conversations_replies(channel_id, thread_ts)` to recover enough thread context to decide whether there is a commitment, unanswered ask, resolution, or user instruction. Do not fetch unrelated channel history. Resolve names from source data rather than guessing.


session_id: 20260810_202328_1e7f50
[SILENT]

Operating record updated: pinned Tars’s measured Slack user ID for self-message filtering. PR https://github.com/GaetanCathelain/Tars/pull/41 merged.
```

The run completed: lock acquired and released (no `engagement-checker.lock`
directory left behind — checked), all three source cursors advanced to the run
end `2026-08-10T22:23:51+02:00` with `last_success` equal to them, state
persisted, exit clean.

**`[SILENT]` here is correct, not a failure.** §8 delivery and §7 filing are
different steps. EC-A44C already had `reminders: 1` at `last_score` 84 and its
score did not rise by ≥8, so §6's "after cooldown, repeat only when the score
rose by at least 8 or the context materially changed" suppressed the reminder.
§7 filed it anyway — as designed.

### (b) A detected loop lands as a GCN issue — **PASS**

State after cycle 1 (the only slice extracted; no snippet text read):

```
filed [('slack:GQ07CQXT7:1786012625.976839', {'id': 'GCN-12', 'priority': 3})]
slack:GQ07CQXT7:1786012625.976839  open  EC-A44C  li={"id":"GCN-12","priority":3}  reminders=1  score=84
```

`linear_issue` **was** written for the filed item — `{id: GCN-12, priority: 3}`
— which is §7c's flush having happened.

**Independent board read** (`mcp__linear__get_issue GCN-12`, not the cycle's own
claim), field by field against the task's checklist:

| requirement | observed | verdict |
|---|---|---|
| team GCN | `teamId` `81e7b769-2a46-4e2a-8db5-c165a7963b0e`, `team` "Gaetan" | ✅ |
| assignee Gaetan | `assigneeId` `4951b192-e49c-4b7e-b491-58c89e66043c` | ✅ |
| exactly one workload label | `labels: ["fix"]` — one, and the correct arm (kind = explicit promise → `fix`) | ✅ |
| explicit priority | `priority: {"value": 3, "name": "Medium"}` — §7b: explicit promise, no due date → 3 | ✅ |
| state Todo, NOT Backlog | `status: "Todo"`, `statusType: "unstarted"`; `stateHistory` shows exactly one state, Todo, from creation — it was never in Backlog | ✅ |
| source handle FIRST LINE of description | line 1 is `Source: slack:GQ07CQXT7:1786012625.976839`, exactly the item's stable id | ✅ |

Full created issue:

```
GCN-12  "Create and send the Verdict Google OAuth client"
createdAt 2026-08-10T20:26:24.745Z   createdById 4951b192-…  (Gaetan's key)
url https://linear.app/mobile-club/issue/GCN-12
description:
  Source: slack:GQ07CQXT7:1786012625.976839
  Evidence: [https://mobileclub-squad.slack.com/archives/GQ07CQXT7/p1786371250299919?thread_ts=1786012625.976839](…)

  Jamil reminded Gaetan about the Verdict OAuth client; Gaetan said he would
  send it after his meeting. Jamil explicitly said it was not urgent and to
  take his time.
```

The loop is genuine — it is the same EC-A44C ("Send Jamil the Verdict Google
OAuth client") that Tars reported to Gaetan in the reporting thread at the
17:00 final pass. No test data was planted anywhere.

Scoped board read, "issues created on GCN in the last hour": **exactly one row,
GCN-12**, `hasNextPage: false`. Nothing else was created.

### (c) Nag-loop guard — **PASS**

Cycle 2, launched 20:29:01Z, exited 20:33:30Z. Session
`20260810_202903_970cda`. Verbatim stdout (`/tmp/ec-cycle2.log`, 132 bytes,
complete, nothing elided):

```
  ⟳ compacting context…
  ⟳ compacting context…
  ⟳ compacting context…

session_id: 20260810_202903_970cda
[SILENT]
```

State after cycle 2 — compared against after cycle 1:

```
last_completed_run 2026-08-10T22:29:25+02:00   (advanced)
cursors slack/email/linear all 2026-08-10T22:29:25+02:00, last_success equal
items 30 total — Counter({'done': 25, 'dismissed': 3, 'snoozed': 1, 'open': 1})   (unchanged)
filed [('slack:GQ07CQXT7:1786012625.976839', {'id': 'GCN-12', 'priority': 3})]    (unchanged — not re-filed)
EC-A44C: open, reminders=1, score=84                                              (unchanged — not re-reported)
```

- **Not re-filed**: `linear_issue` is still the single `{GCN-12, priority 3}`;
  no second entry, and no item gained a `linear_issue`.
- **Not re-reported**: `reminders` still 1, and the returned text is `[SILENT]`.
- **No duplicate on the board**: independent `mcp__linear__list_issues` team
  GCN, `createdAt: -PT2H` → **exactly one row, GCN-12**, `hasNextPage: false`.
  Highest GCN key is still 12.

Cycle 2 also made no skill edit — VM SKILL.md sha256 unchanged across it.

### (d) Pull closes the loop — **PASS**

Close performed from this session, per instruction, with `id` + `state` only
and **no `labels`** (they replace):

```
mcp__linear__save_issue  id=GCN-12  state=Done
→ status "Done", statusType "completed", completedAt 2026-08-10T20:33:57.046Z
→ labels ["fix"]   (preserved — the "omit labels" rule holds, as linear-ticketing §13 records)
→ priority still {"value":3,"name":"Medium"}, assignee/team unchanged
```

Cycle 3, launched 20:34:04Z, exited 20:36:50Z. Session
`20260810_203405_6454ee`. Verbatim stdout (`/tmp/ec-cycle3.log`, 1500 bytes,
complete, nothing elided):

```
  ⟳ compacting context…
  ┊ review diff
a//tmp/update_engagement_state.py → b//tmp/update_engagement_state.py
@@ -0,0 +1,32 @@
+import json
+from pathlib import Path
+
+p = Path('/home/gaetan/.hermes/state/engagement-checker.json')
+d = json.loads(p.read_text())
+run_end = '2026-08-10T22:34:28+02:00'
+for name in ('slack', 'email', 'linear'):
+    src = d['sources'][name]
+    src['cursor'] = run_end
+    src['last_success'] = run_end
+for item in d.get('items', {}).values():
+    linear_issue = item.get('linear_issue') or {}
+    if linear_issue.get('id') == 'GCN-12':
+        item['status'] = 'done'
+        item['updated_at'] = '2026-08-10T22:33:57.110+02:00'
+        item.setdefault('status_history', []).append({
+            'at': '2026-08-10T22:33:57.110+02:00',
+            'status': 'done',
+            'reason': 'linear:completed',
+        })
+        item['status_history'] = item['status_history'][-10:]
+        linear_issue['priority'] = 3
+        linear_issue['closed_at'] = '2026-08-10T22:33:57.046+02:00'
+        break
+else:
+    raise RuntimeError('GCN-12 routing target missing')
+seen = d['sources']['linear'].setdefault('seen', [])
+if 'linear:GCN-12' not in seen:
+    seen.append('linear:GCN-12')
+d['sources']['linear']['seen'] = seen[-500:]
+d['last_completed_run'] = run_end
+p.write_text(json.dumps(d, ensure_ascii=False, indent=2) + '\n')
  ⟳ compacting context…

session_id: 20260810_203405_6454ee
[SILENT]
```

That script *is* the mechanism working, visible in the open: the collector's
delta view pulled GCN-12 (its `updatedAt` 20:33:57Z is newer than the
20:29:25+02:00 cursor), the routing index matched `linear_issue.id == GCN-12`
back to the parent item — not a new item — and §5 applied the terminal state to
the parent. Note `raise RuntimeError('GCN-12 routing target missing')` on the
`else`: the run refused to proceed if routing had *not* matched.

State after cycle 3:

```
items 30 total — Counter({'done': 26, 'dismissed': 3, 'snoozed': 1})     <- ZERO open
PENDING slack:D08K34MA3QT:1786350553.536169  snoozed  EC-73EC  li=null   (untouched, wakes 2026-08-12)
TARGET status            done
TARGET linear_issue      {"id":"GCN-12","priority":3,"closed_at":"2026-08-10T22:33:57.046+02:00"}
TARGET status_history:
   2026-08-10T16:14:10+02:00  open  | Explicit promise retained, but reminder suppressed because the request…
   2026-08-10T22:33:57.110+02:00  done | linear:completed
```

**The item cleared from the queue.** Reason is `linear:completed` — §5's
correct routed-terminal path, not `stale:30d`. `closed_at` was set from the
*observation* (§5 step 1: the issue is already terminal, so no redundant
`save_issue` write was made) — which is why no second Linear write appears
anywhere in this probe. `linear_issue.priority` refreshed to 3, matching the
board.

---

## Skill drift — Tars rewrote the deployed file during cycle 1

This is the one thing the coordinator has to rule on. It is **expected Tars
behaviour under SOUL rule 2** (`skill_manage` + branch + `gh pr create` +
squash-merge, same turn), not a fault, but it means "sha256 matches both sides"
is true *at deploy* and false *now*.

```
at deploy 20:20:4xZ   VM = 9834b290…c3310  = repo working tree   (67416 bytes)
after cycle 1 20:27Z  VM = bb55e1efe605b34ec33ed3f30fa6fdea902510a1ce4e501d2ee08c7495bd37d9  (67313 bytes)
after cycle 2         VM = bb55e1ef…d37d9  (unchanged)
after cycle 3         VM = bb55e1ef…d37d9  (unchanged)
version frontmatter still 1.7.0 throughout
```

`diff -u` repo-working-tree vs VM: **one hunk, one paragraph, §2 Collect Slack
deltas, nothing else in the file changed.** The §7a "Empty is a value" fix is
untouched.

Removed:

> …The `bot_id` presence test is structural and carries the rule on its own;
> Tars's own literal bot/app id is **not yet measured**, so it cannot be pinned
> here the way Gaetan's `U08BDJAMSRZ` is. Pin it in this sentence the first time
> a run records it, and until then do not weaken the structural test to
> compensate.…

Added:

> …The `bot_id` presence test is structural and carries the rule on its own;
> Tars's own measured Slack user id is `U0BBH85NAKH`, so discard that sender
> explicitly as well as applying the structural `bot_id` test.…

That is the file doing exactly what it asks for ("Pin it in this sentence the
first time a run records it"), and the pin is correct — `U0BBH85NAKH` is the id
the reporting thread shows for Tars. It strictly *tightens* the §2 filter.

PR **#41** — "engagement-checker skill: pin measured Tars Slack user id for
self-message filtering", branch `tars/engagement-checker-20260810T202719Z`,
merged 2026-08-10T20:27:25Z, **one file changed:
`skills/engagement-checker/SKILL.md`, +8/−6**. So `main` now carries v1.7.0
*including* the §7a fix (Tars pushed the whole current VM file), plus the pin.
Consequence for the coordinator: **the uncommitted v1.7.0 in this worktree is
now a subset of what is already on `main`** — reconcile before committing it, or
the pin gets reverted.

Nothing was restored, per CLAUDE.md ("when it and the VM disagree, find out
which side moved before overwriting either") — the VM side moved, deliberately,
by its own sanctioned mechanism.

---

## Slack — nothing was posted

Two independent guarantees, both required:

1. **Structural**: `hermes chat` has no delivery target. `deliver:
   slack:C0BP2GZUFSR:1786359613.979759` is a cron job property; these were not
   cron fires.
2. **Observed after the fact**, reading with Gaetan's own Slack connector,
   window `oldest=1786393350.000000` (= 20:22:30Z, before cycle 1 launched):
   - reporting thread `C0BP2GZUFSR:1786359613.979759` → `No thread messsages`
   - Tars DM `D0BBYNM01BL` → no messages
   - `#tech` `GQ07CQXT7` (the source channel of the filed loop, where a stray
     reply would have been worst) → no messages

**Zero Slack writes across all three cycles.** ✅

The only writes this probe made anywhere outside the VM are the three Linear
ones it was scoped to make: GCN-12 created by cycle 1, GCN-12 closed by this
session for step (d), and PR #41 authored+merged by Tars itself.

---

## State-file integrity

```
79440 bytes, parses as JSON
top_keys ['ambiguous_instructions','daily_catchup','initialized_at','items',
          'last_completed_run','last_failure_notice','sources','version']
version 1
seen ring sizes: slack 500, email 500, linear 154  (was 152 pre-run; +2, under the 500 cap)
items 30 (unchanged count — nothing lost, nothing orphaned)
```

Both pre-existing unknown top-level keys (`daily_catchup`, `initialized_at`)
survived all three cycles — §"State writes must preserve unknown future
top-level keys" holds. No corruption. The `.bak-wf6-ec170-20260810T202040Z`
copy is intact and was never needed.

Per-cycle `linear_issue` trace for the filed item, which is the question that
mattered:

| after | `linear_issue` |
|---|---|
| baseline | `null` (and `filed_count 0` across all 30 items) |
| cycle 1 | `{"id":"GCN-12","priority":3}` |
| cycle 2 | `{"id":"GCN-12","priority":3}` — byte-identical, not re-written |
| cycle 3 | `{"id":"GCN-12","priority":3,"closed_at":"2026-08-10T22:33:57.046+02:00"}` |

No whole-file dump was taken at any point; every reading above is a structural
extract by a `python3 -c` one-liner printing keys, counts and ids only. No item
snippet text left the VM.

---

## Verdicts

| check | verdict | what proves it |
|---|---|---|
| **(a)** one cycle runs end to end | **PASS** | cycle 1 completed 20:23:26→20:28:00Z, cursors + `last_success` advanced on all three sources, lock released, state persisted, clean exit |
| **(b)** detected loop lands as a GCN issue | **PASS** | **GCN-12** created 20:26:24.745Z; independently verified on the board — team GCN, assignee Gaetan, one label `fix`, priority 3 explicit, state Todo (never Backlog per `stateHistory`), `Source: slack:GQ07CQXT7:1786012625.976839` as description line 1 |
| **(c)** nag-loop guard | **PASS** | cycle 2 → `[SILENT]`; `linear_issue` unchanged, `reminders` still 1, and a scoped board read returns exactly one issue created in the window |
| **(d)** pull closes the loop | **PASS** | GCN-12 closed Done from this session (labels preserved); cycle 3 routed it back by key, set item `done` / reason `linear:completed` / `closed_at`, leaving **zero** `open` items |

**What remains unproven, precisely:**

1. **The automatic cron path has not been exercised.** All three cycles were
   manual `hermes chat` runs with an appended delivery guard and `--yolo`. The
   next real fire is ~08:00Z 2026-08-11 (10:00 Paris). What is proven is the
   skill logic under the real prompt on the real model; what is not proven is
   the cron fire + Slack delivery of a §8 reminder.
2. **A §8 delivery was never rendered.** All three cycles returned `[SILENT]`
   (correctly — cooldown, then a resolved queue). So the `[GCN-42]`-style bullet
   rendering for a filed item, and the `Coverage:` line, are still untested
   live. This is unchanged from before the deploy and outside check 2's wording.
3. **"skills list shows v1.7.0"** cannot be shown literally — the CLI exposes no
   version. Substituted with sha256 + frontmatter + the behavioural proof.
4. **The multi-item filing path is untested.** Exactly one item was eligible, so
   §7's 10-per-run cap, its priority-ascending ordering, and the
   `N items awaiting filing` coverage note saw no exercise.
5. **§7a's three genuine blockers are untested** — no run hit an `error`
   payload, an unmakeable call, or the 8-page cap. What is proven is the thing
   that was broken: three `hasNextPage == false` calls over a board whose
   Backlog column returns zero rows no longer block the create.

**No rollback trigger fired.** No duplicates, no non-GCN write, no Slack post,
no state corruption, no gateway restart. Both `.bak`s remain in place, unused.

## Cleanup — for the coordinator to decide

- **GCN-12 is NOT disposable verification data.** It is a real open loop of
  Gaetan's (send Jamil the Verdict Google OAuth client, promised 2026-08-07,
  Jamil said not urgent). It was filed by the skill on its own judgement, and
  then **closed by step (d) while the underlying commitment may still be
  outstanding**. Step (d) required closing it, so this was unavoidable — but
  someone should decide whether to reopen GCN-12 or let it lie. Nothing was
  deleted.
- Side effect of that close: EC-A44C is now `done` in the local queue with
  reason `linear:completed`, so **the skill will not remind Gaetan about the
  Jamil OAuth client again**. If the commitment is still live, reopening GCN-12
  in Linear is the supported way back — §5's reopen rule returns the parent to
  `open` and clears `closed_at`.
- Backups left in place, both unused:
  `~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-wf6-ec170-20260810T202040Z`
  and `~/.hermes/state/engagement-checker.json.bak-wf6-ec170-20260810T202040Z`.
- Temp files removed from the VM: `/tmp/ec-170-src.md`. Left behind:
  `/tmp/ec-cron-prompt.txt`, `/tmp/ec-run-prompt.txt`, `/tmp/ec-cycle{1,2,3}.log`
  (no secrets in any of them), and `/tmp/update_engagement_state.py` written by
  cycle 3 itself.
- `daily-work-brief`, `linear-ticketing`, `config.yaml`, every cron job and the
  MCP server were **not touched**. No git command was run from this session.

## Secret handling

`LINEAR_API_KEY` and every other secret were never read, printed, logged or
placed on argv. No `sops` invocation. No `cat ~/.hermes/.env`. The collector
inherits the key inside its own process, as designed, and nothing in this file
or in any command run derives from it.
