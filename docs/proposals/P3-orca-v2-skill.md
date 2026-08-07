# P3 — Orca v2: Tars drives real Orca sessions on cooper

**Status:** proposed, NOT applied · **Source:** orca-a2a peer, 2026-08-07 ·
**Full text:** `artifacts/delegate-to-cooper-SKILL.md` (437 lines, complete
replacement for `~/.hermes/skills/delegate-to-cooper/SKILL.md`) ·
**Spec section:** `artifacts/wf5-orca-delegation-v2-section.md` (append to
`docs/specs/wf5-orca-delegation.md`) · **Reviews:**
`artifacts/review-{mechanical,usability,law}.md` · **Evidence:**
`status/probes/wf5/orca-v2-*.md` · **Decide:** Gaetan ·
**Must land together with [P4](P4-soul-guidelines-redesign.md)** — see Order.

## What changes, and why v2 is *smaller* than v1

v1 (shipped tonight, hours old) was built to make Tars **unable** to act:
`delegate.sh` fixed-flag wrapper, deny-rule `.claude/settings.json`, a
no-git-remote sandbox. Your separation-of-concerns ruling deletes the rationale
for all of it. v2 is **one `SKILL.md` on the VM** — no script on cooper, no
settings file, no sandbox. Tars gets the `orca` CLI over ssh and prose telling it
what its job is.

## Mechanism

```
ssh cooper → orca orchestration run-create
           → task-create → worker-start --agent claude --repo mc-metarepo
           → check --run <id> --wait --types worker_done
           → worker-read → report
```

The durable `run_id`/`dispatch_id` carry across turns. **That is the whole v2
win:** Hermes' 300 s turn cap becomes a *checkpoint* instead of a lost job.

**The load-bearing fact, measured** (a mechanical reviewer claimed `check --run`
needs a bound Run and would have stranded the design — it was wrong, generalising
from a stale id and weighting a bundled prose doc over `--help`). From a plain
non-Orca shell, i.e. exactly Tars's stateless position:

```
$ orca orchestration run-create --objective "tars v2 addressability probe" --json
  → "run": { "id": "run_24d000a3a6b6", … }
# then from a FRESH separate process:
$ orca orchestration check --run run_24d000a3a6b6 --peek --json
  → {"ok":true,"result":{"messages":[],"count":0,"acknowledged":null,"runId":"run_24d000a3a6b6"}}
$ orca orchestration task-list --run run_24d000a3a6b6 --json    → ok:true
$ orca orchestration worker-list --run run_24d000a3a6b6 --json  → ok:true
```

No `run_not_found`; `check --help` documents `[--run <run_id>]` in its usage line.
Independently re-measured. A `worker-show --dispatch` polling fallback stays in
the skill in case it ever appears.

## Three real defects the adversarial reviews caught (all fixed)

1. **`--ack` was missing entirely.** `check --help`: *"A bound Run replays the
   same Delivery until `--ack`"* — without it, every post-300 s re-attach replays
   the same completion message forever. The multi-turn loop, i.e. the entire v2
   win, was broken.
2. **The separation rule enumerated file types**, so asked for documentation Tars
   would have written it itself and believed it compliant. Now scoped by role
   (same fix as SOUL rule 1).
3. **"I will re-attach" was a promise with no mechanism** — nothing gives a Slack
   agent a later turn. Now wired to
   `hermes cron create "5m" "…" --deliver slack --repeat 1` (bare platform name ⇒
   `SLACK_HOME_CHANNEL`; `/bg` cannot reach home). Not optional when you asked to
   be told when it's done.

Also corrected: questions use `reply --id <msg_id> --body <text>` (`send --to
dispatch:` fails silently); unique brief path `/tmp/tars-brief-<slug>.md`; the
credential path list (`~/.ssh/`, `~/.aws/`, `~/.claude.json`, `~/.config/sops/`,
`~/.pgpass`, `*.key`, `.env`) is kept and **framed as secret hygiene, explicitly
not a place-restriction**.

Two of the peer's own earlier claims, corrected by measurement: `orca` **is** on
PATH over non-interactive ssh (`/usr/local/bin/orca` — it's `~/.local/bin` that
isn't), and `worker-start` **does** accept `--timeout-ms`.

## Scope settled by your ruling

- **Repos:** mc-metarepo is the default; all Orca-registered repos are fair game;
  Tars may register more. gaetan-metarepo deliberately not registered.
- **Delivery shape:** you prompt Claude Code and the session does what it does —
  branch, commits, PR included. *"Tars never creates a PR"* survives as **"Tars
  does not do the work itself"**, not "no PR may result from a Tars-initiated
  task". (I still owe `README.md` a wording fix for its old absolute — one edit,
  when this lands.)
- **Layout:** the skill teaches Tars Orca's model — projects → worktrees
  (with parent/child) → tabs of type browser / Claude Code / etc.
- Knowledge/preferences transfer is a **later phase, to be planned not built**.

## Open sub-decisions

| # | Question | Note |
|---|---|---|
| P3-a | **Cleanup default:** release the dispatch, **keep the worktree**, report path + branch — in v2 the checkout *is* the work product you pick up in Orca. Cost named honestly: mc-metarepo is already at 9 worktrees, nothing self-cleans, the pile grows | Auto-`rm` risks silently destroying delegated work — pruning only on your word |
| P3-b | **Rename** the skill to `drive-orca-on-cooper`? Ships as `delegate-to-cooper` (zero-risk drop-in at the existing path); renaming = one frontmatter line + removing the old VM directory | Cosmetic but the old name now misdescribes it |
| P3-c | The delivery-id **field name** on a non-empty batch is unobserved (an empty batch returns `messages/count/acknowledged/runId`). The skill tells Tars to read the raw JSON and report the real shape | First real run corrects the spec — no action needed |

## Order · restart · verify · rollback

- **Order:** must land **with P4**. Right now the live skill still says *"Never
  rm, rmdir, mv, chmod, sudo, systemctl on cooper"* (`SKILL.md:47`) — SOUL alone
  would grant sudo while a loaded skill forbids it.
- **Restart:** none — a dropped `SKILL.md` rebuilds the skills index via the
  path→(mtime,size) manifest. Do not restart while the other peer is live.
- **Verify:** `hermes skills list` shows the entry and the count moves; then the
  **E2E**: one read-only brief into mc-metarepo (spawn → `check --wait` →
  `worker-read` → report), which also settles P3-c.
- **Rollback:** the v1 `SKILL.md` exists **only on the VM, never in git** — take
  its `.bak` before touching anything; that's the only rollback path.

## Footprint from the investigation

One empty orchestration run on cooper (`run_24d000a3a6b6`: no tasks, no workers,
no worktree) from the addressability measurement. The CLI has no `run-delete`, so
it stays as an inert row — recorded, not chased. It is the only mutation made
anywhere tonight.

## Decision

- [ ] Apply v2 skill + P4 SOUL together, then run the mc-metarepo E2E
- [ ] Apply with edits (mark them on `artifacts/delegate-to-cooper-SKILL.md`)
- [ ] Not yet
- P3-a worktrees: [ ] keep (proposed) · [ ] prune after report
- P3-b rename: [ ] `drive-orca-on-cooper` · [ ] keep `delegate-to-cooper`
