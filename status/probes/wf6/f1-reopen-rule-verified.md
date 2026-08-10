# F1 — `closed_at` latch clearing (reopen rule) — VERIFIED ON THE LIVE VM

**Verdict: PASS on all four criteria (a)(b)(c)(d).**

Date: 2026-08-10, 21:22–21:32 UTC on the Tars VM (192.168.0.9).
Fix under test: engagement-checker §5 (line 519 of the live `SKILL.md`), skill v1.7.1:

> A routed filed issue in a **non-terminal** state whose parent is `done` or
> `dismissed` with a `linear:*` reason: return the parent to `open` with a fresh
> cooldown, **and delete `linear_issue.closed_at` in the same step**.

This path had never been executed. Every prior verification closed an issue and
stopped. The reopen edge — the one that reinstates the original blocker — was
untested until this run.

Live skill path on the VM: `~/.hermes/skills/orchestration/engagement-checker/SKILL.md`
(note: **not** `~/.hermes/skills/engagement-checker/`; the skill sits under the
`orchestration/` group).

---

## 1. Backup taken before anything ran

```
BAK = /home/gaetan/.hermes/state/engagement-checker.json.bak-f1-20260810T212230Z
sha256(bak)                  = 301d0f322b0a3792db981fb8064d02fc9a1c59c460125ec80099f2555135ed4b
sha256(live, at backup time) = 301d0f322b0a3792db981fb8064d02fc9a1c59c460125ec80099f2555135ed4b
```

Copy made under `flock ~/.hermes/.wf3.lock`. Identical hashes = faithful copy.
Not restored — the run passed.

## 2. BEFORE state (precondition — HELD)

Structural fields only; no snippet/title/body text was read out of the state file.

| field | value |
|---|---|
| state file | valid JSON, 79 440 bytes, sha256 `301d0f32…5ed4b` |
| top-level keys | `ambiguous_instructions, daily_catchup, initialized_at, items, last_completed_run, last_failure_notice, sources, version` (8) |
| `items` count | **30** |
| item (short_id `EC-A44C`) `status` | **`done`** |
| last `status_history` entry | `{at: 2026-08-10T22:33:57.110+02:00, status: done, reason: linear:completed}` |
| prev `status_history` entry | `{at: 2026-08-10T16:14:10+02:00, status: open, reason: …}` |
| `status_history` length | 2 |
| `linear_issue` keys | `['closed_at', 'id', 'priority']` |
| `linear_issue.id` | `GCN-12` |
| `linear_issue.priority` | `3` |
| **`linear_issue.closed_at`** | **PRESENT** — `2026-08-10T22:33:57.046+02:00` |
| `snooze_until` | `None` |
| `reminders` | 1 entry: `{at: 2026-08-10T17:00:34.600661+02:00, score: 84}` |
| items carrying a `linear_issue` | 1 (only this one) |

Linear side, same moment: **GCN-12 = Todo / `statusType: unstarted`** (non-terminal),
priority Medium (3), assignee Gaëtan Cathelain, labels `["fix"]`.
Its `stateHistory` records the full arc: Todo → Done (20:33:57Z) → Todo (21:18:25Z).

So: item `done` + `linear:*` reason + `closed_at` set, while the issue is
non-terminal. **The reopen rule's trigger, exactly. Precondition HELD.**

Board before the run: 12 issues, GCN-1 … GCN-12, highest key **GCN-12**.

Gateway before: `ActiveEnterTimestamp = Mon 2026-08-10 20:00:38 UTC`, active/running.

## 3. The cycle

Ran through the real agent loop with the **live cron prompt for job `62e8cd9db637`**
("Gaetan engagement checker", `*/30 10-16 * * 1-5`, skill `engagement-checker`),
read verbatim from `~/.hermes/cron/jobs.json` — i.e. the scheduling-contract text as
just updated. The job carries `model: null` / `provider: null`, so the CLI defaults
are the same inference path cron itself takes.

Appended override (the only deviation from the cron text):

> VERIFICATION RUN OVERRIDE: do not post to Slack; return the text to me instead.
> Do not call any Slack write/post tool (no chat.postMessage, no send_message, no
> draft) for any reason this run. Slack READS are fine and expected. Everything
> else in this prompt and in the skill stays exactly as written, including all
> Linear reads and writes and the durable-state persist.

Invocation:

```
hermes chat -Q --yolo -s engagement-checker --source tool -q "$(cat /tmp/f1_prompt.txt)"
```

- `--yolo`: pre-approved deviation, recorded here — needed so the dangerous-command
  approval prompt could not block the audited local Linear collector in a headless run.
- `hermes chat` has no Slack delivery path at all (the `deliver:` field is a cron
  property, not a chat one), so delivery was structurally prevented, not merely asked for.
- Session `20260810_212430_de4268`. Started 21:24:30 UTC; state file written 21:27:30 UTC.

**Returned text (stdout, in full):**

```
  ⟳ compacting context…
  ⟳ compacting context…
• Create and send the Verdict Google OAuth client — Jamil; the ticket was reopened because the commitment is still live. Next: create the client and send it to Jamil. [GCN-12]
```

stderr: `session_id: 20260810_212430_de4268` only. No errors.
A secret scan (`lin_api|xox[bposd]|sk-`) over both streams returned **0 matches**;
no secret was read, printed or persisted at any point in this probe.

## 4. AFTER state

| field | value | vs before |
|---|---|---|
| state file | valid JSON, 79 610 bytes, sha256 `eee4ae25…bb0807` | changed (expected) |
| top-level keys | same 8, unchanged | ✅ preserved |
| `items` count | **30** | ✅ unchanged |
| `status` | **`open`** | `done` → `open` |
| last `status_history` entry | `{at: 2026-08-10T23:19:51.823+02:00, status: open, reason: linear:reopened}` | **new entry appended** |
| prev `status_history` entry | `{at: 2026-08-10T22:33:57.110+02:00, status: done, reason: linear:completed}` | old head, retained |
| `status_history` length | 3 | 2 → 3, append-only |
| `linear_issue` keys | **`['id', 'priority']`** | `closed_at` **key removed entirely** |
| `linear_issue.id` | `GCN-12` | ✅ unchanged |
| `linear_issue.priority` | `3` | ✅ unchanged |
| **`linear_issue.closed_at`** | **ABSENT** | was set → deleted |
| `snooze_until` | `None` | unchanged |
| `reminders` | 2 entries; new: `{at: 2026-08-10T23:24:55+02:00, score: 84}` | **fresh cooldown stamped** |
| `last_score` | 84 | unchanged |
| `last_completed_run` | `2026-08-10T23:24:55+02:00` | advanced |
| items carrying a `linear_issue` | 1 | ✅ unchanged — nothing else routed |

## 5. Board read — no duplicate, no new issue

Full GCN team read after the cycle (`includeArchived: true`, `hasNextPage: false`,
so the read is complete): **12 issues, GCN-1 … GCN-12. Highest key still GCN-12.**

No second "Create and send the Verdict Google OAuth client". No GCN-13. GCN-12 itself
is untouched by the run: still Todo / `unstarted`, priority Medium (3), assignee
Gaëtan Cathelain, labels `["fix"]`. The reopen adopted the existing issue rather than
minting one — the dedupe scan found the item's stable id
(`Source: slack:…` head line) in GCN-12's description and matched it.

## 6. Pass criteria

| # | criterion | verdict | evidence |
|---|---|---|---|
| **a** | item `status` returned to `open` | **PASS** | `status='open'`; history head `{status: open, reason: linear:reopened}` |
| **b** | `linear_issue.closed_at` now ABSENT | **PASS** | `linear_issue` keys went `['closed_at','id','priority']` → `['id','priority']`. The key is **deleted**, not nulled. A set-only latch would have left it. **This is the fix, and it fired.** |
| **c** | fresh cooldown applied | **PASS** | new `reminders` entry `{at: 2026-08-10T23:24:55+02:00, score: 84}` — the cooldown clock restarts from this run, not from 17:00 |
| **d** | `linear_issue.id` still GCN-12, priority preserved, no new issue | **PASS** | `id='GCN-12'`, `priority=3` both unchanged; board still tops out at GCN-12 with no duplicate row |

## 7. Also checked

- **Reminder eligibility restored — YES, and demonstrated, not inferred.** The item
  did not merely become `open`: this very cycle emitted its reminder line
  (`… [GCN-12]`) and stamped a new `reminders` entry. Gaetan's real commitment is
  being tracked again, which is the entire point of the rule.
- **Gateway: NOT restarted.** `ActiveEnterTimestamp = Mon 2026-08-10 20:00:38 UTC`
  before **and** after, byte-identical; `ActiveState=active`, `SubState=running`.
  (`NRestarts` deliberately ignored — it is known unreliable and missed the observed
  20:00Z restart.)
- **State file integrity: intact.** Valid JSON, item count still 30, all 8 top-level
  keys preserved (`ambiguous_instructions`, `daily_catchup`, `initialized_at`,
  `items`, `last_completed_run`, `last_failure_notice`, `sources`, `version`).
  `status_history` grew by append; no entry was rewritten or dropped.
- **Nothing posted to Slack.** Three independent confirmations:
  1. `~/.hermes/logs/gateway.log` last modified **21:21:54 UTC** — before the run
     started at 21:24:30. The gateway did nothing for the whole window.
  2. Reporting conversation `C0BP2GZUFSR` thread `1786359613.979759`: no replies.
  3. Home DM `D0BBYNM01BL`: no messages in the window.
- No `sops` invocation, no `.env` read, no secret on argv, in a log, or in this file.

## 8. What this closes

The adversarial review's finding was that `closed_at` was a set-only latch: once
stamped, a reopened issue could never be closed again, so the loop would sit
permanently open on the board — the original blocker re-entering through the reopen
edge. The clearing step is now proven live: the latch cleared, the item went back to
`open` with a fresh cooldown, the reminder resumed, and no ghost or duplicate issue
was produced.

The state-file `.bak` at
`~/.hermes/state/engagement-checker.json.bak-f1-20260810T212230Z` was left in place
and **not** restored — the run passed, and the post-run state is the correct one.
Nothing in the state file was hand-edited at any point.
