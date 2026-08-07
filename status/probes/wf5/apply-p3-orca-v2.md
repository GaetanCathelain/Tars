# P3 — apply Orca v2 delegation skill + first live E2E

Peer session: `apply-p3-orca` (worktree
`/home/gaetan/orca/workspaces/Tars/apply-p3-orca`, branch
`GaetanCathelain/apply-p3-orca`). Mandate: `docs/plans/apply-P3-orca-v2.md`.
Hub: the Tars orchestrator session. Sibling peer applied P4 (SOUL) concurrently
under a joint-apply gate.

All times UTC (the VM clock is UTC).

---

## 1. Backup — taken FIRST, before anything was touched

The v1 `SKILL.md` existed **only on the VM, never in git**. It is the sole
rollback path, so it was copied before any other action.

```
$ ssh gaetan@192.168.0.9 'cp -p ~/.hermes/skills/delegate-to-cooper/SKILL.md \
                             ~/.hermes/skills/delegate-to-cooper/SKILL.md.bak-p3'
```

| | |
|---|---|
| Backup path | `~/.hermes/skills/delegate-to-cooper/SKILL.md.bak-p3` (on the VM, 192.168.0.9) |
| v1 md5 | `d61888ecb34a2342663a7122414aecdb` |
| v1 size / lines | 3304 bytes / 77 lines |
| Mode / owner | `-rw-rw-r-- gaetan gaetan`, preserved by `cp -p` |
| `diff SKILL.md SKILL.md.bak-p3` | empty — **IDENTICAL** |

The write refused to proceed if `.bak-p3` already existed (guard in the command),
so an existing backup could not be clobbered by a re-run.

**Index-pollution check (not in the mandate — verified anyway).** A stray file in
a skills directory could in principle be picked up as a second skill. It cannot
be: every scanner in the Hermes source matches the filename exactly —
`agent/learning_graph.py:80`, `agent/curator_backup.py:185`,
`tools/skill_usage.py:359,551,1235,1318`, `tools/skills_sync_client.py:533,1713`,
`tools/skills_hub.py:3376,3391,3701` all use `rglob("SKILL.md")`, and
`scripts/build_skills_index.py:178` tests `path.endswith("/SKILL.md")`.
`SKILL.md.bak-p3` matches none of them. The backup is safe where it sits.

**Second rollback copy (beyond the mandate).** Because "only on the VM" is a
single point of failure, the v1 file was also pulled into git, byte-identical:

```
artifacts/delegate-to-cooper-SKILL-v1-live.md   md5 d61888ecb34a2342663a7122414aecdb
```

Rollback is therefore possible from either the VM `.bak-p3` or from git.

### Rollback command (if ever needed)

```bash
ssh gaetan@192.168.0.9 'D=~/.hermes/skills/delegate-to-cooper
flock ~/.hermes/.wf3.lock -c "cp -p $D/SKILL.md.bak-p3 $D/.SKILL.md.tmp && mv $D/.SKILL.md.tmp $D/SKILL.md"'
# then verify: md5sum must be d61888ecb34a2342663a7122414aecdb
```

---

## 2. What was replaced, and the contradiction that forced the joint gate

v1 `SKILL.md:47`, verbatim:

> `- Never `rm`, `rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper.`

This is what P4's new SOUL rule 7 contradicts: SOUL would grant `sudo` on cooper
while a loaded skill forbade it. Confirmed live before the apply — the gate was
justified, not ceremonial.

v1 also confined Tars to `~/orca/workspaces/tars-delegated/` (`SKILL.md:54-56`)
and to a four-command allowlist (`:33-40`). v2 drops all of it.

---

## 3. Staging + rehearsal (before the gate)

v2 source: `artifacts/delegate-to-cooper-SKILL.md`, 437 lines,
md5 `929b23de94fd0697bf3ae2709f2dca42`.

Staged to the VM over stdin (never through argv or a heredoc, so nothing in the
file — it contains `$(cat …)` and a `BRIEF` heredoc — could be shell-expanded):

```
$ ssh gaetan@192.168.0.9 'cat > /tmp/p3-skill.new' < artifacts/delegate-to-cooper-SKILL.md
$ ssh gaetan@192.168.0.9 'md5sum /tmp/p3-skill.new'
929b23de94fd0697bf3ae2709f2dca42  /tmp/p3-skill.new    # matches source exactly
```

The apply command was then **rehearsed on a scratch directory** so that the GO
would be instant and already proven:

```
before: -rw-rw-r-- gaetan gaetan 3304  md5 d61888ecb34a2342663a7122414aecdb
after:  -rw-rw-r-- gaetan gaetan 25050 md5 929b23de94fd0697bf3ae2709f2dca42
=== REHEARSAL OK, scratch removed ===
```

Mode and owner preserved; the rename is atomic; the content is verified *before*
it goes live. Scratch dir removed.

---

## 4. Baseline before the apply

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes skills list'
7 hub-installed, 66 builtin, 7 local — 80 enabled, 0 disabled
```

**Spec correction.** `docs/specs/wf5-orca-delegation.md` (v2 section, §Superseded
v1 artifacts) predicts *"Skill count should stay 77 / 6-local"*. That number is
**stale** — measured baseline on 2026-08-07 is **80 enabled / 7 local / 66
builtin / 7 hub-installed**. The prediction's substance still holds: this is a
content replacement, not an add, so the count must not move.

The count therefore cannot by itself prove the new file parsed. The signal that
does: v1 frontmatter carries `category: delegate-to-cooper`, v2 carries
`category: orchestration`. A category move proves Hermes re-read the file.

---

## 5. E2E preflight (read-only, run before the gate)

| Check | Result |
|---|---|
| `ssh cooper` from the VM | OK — user `gaetan`, `192.168.0.4`, no prompt |
| `orca` on non-interactive PATH | `/usr/local/bin/orca` — confirms the spec's correction |
| `orca status --json` | `ok:true`, `runtime.reachable: **true**`, app running pid 218650 |
| `orca repo list --json` | `8099e312-3232-46f2-83a9-97aeaf5de5a2` → `/home/gaetan/dev/mc-metarepo` — **matches the id hardcoded in the skill** |
| | `a71db494-6e32-437d-be39-acff38061723` → `/home/gaetan/dev/Tars` |
| `orca account list --json` | claude: 2 accounts, active `009bfa67-…` (`gaetan.cathelain-claude02@mobile.club`, subscription-oauth) |

**Correction to a spec claim.** The v2 spec says *"Only `claude` and `codex` are
verified-usable `--agent` values on cooper."* Measured: `result.codex.accounts`
is **empty** and `activeAccountId` is `null`. `--agent codex` has no account to
run under right now — `claude` is the only presently usable value.

---

## 6. Joint-apply gate

Everything above was completed with **nothing written to `~/.hermes/`**.
`P3 READY` was sent to the hub, reporting the backup, the staged file, the
rehearsal, the baseline, and the confirmed v1/SOUL contradiction.

Not touched, per mandate: `~/.hermes/SOUL.md` (sibling peer),
`~/.hermes/config.yaml`, `status/lane-a.md` (hub single-writer). No gateway
restart at any point.


---

## 7. APPLIED — 2026-08-07T22:32:48Z

Gaetan said "Go" directly in the peer session. I applied on his word rather than
on the hub's GO, and told the hub so immediately in the `P3 APPLIED` message so
it could sequence the sibling's P4. Recording that deviation plainly: the gate
was satisfied by the higher authority, not by the mechanism the mandate named.

```bash
D=~/.hermes/skills/delegate-to-cooper
flock ~/.hermes/.wf3.lock -c "
  cp -p $D/SKILL.md $D/.SKILL.md.tmp &&
  cat /tmp/p3-skill.new  > $D/.SKILL.md.tmp &&
  mv $D/.SKILL.md.tmp      $D/SKILL.md
"
```

`cp -p` seeds the temp file with the original's mode/owner, the `cat` overwrites
content in place, and `mv` is an atomic same-filesystem rename. The lock is held
for milliseconds and the content was md5-verified before it went live.

| | As applied |
|---|---|
| md5 | `929b23de94fd0697bf3ae2709f2dca42` — exact match to `artifacts/delegate-to-cooper-SKILL.md` |
| Size / lines | 25050 bytes / 437 lines |
| Mode / owner | `-rw-rw-r-- gaetan gaetan` — preserved |
| Backup after apply | `SKILL.md.bak-p3`, md5 `d61888ec…` — still intact |
| v1 prohibition `Never rm, rmdir…` | **0 occurrences** — gone |
| Gateway | **not restarted** |
| `SOUL.md`, `config.yaml` | not touched |

### Registration

```
$ hermes skills list
│ delegate-to-cooper   │                     │ local     │ local     │ enabled │
7 hub-installed, 66 builtin, 7 local — 80 enabled, 0 disabled
```

Entry present and enabled; totals unchanged, which is correct for a content
replacement.

**My own P3 READY prediction was wrong, corrected here.** I said the Category
column would move `delegate-to-cooper` → `orchestration` and prove the file
parsed. It does not. The Category column is derived from the **on-disk path**
(`skills/<category>/<name>/SKILL.md`), not from frontmatter — `delegate-to-cooper`
and `hermes-lcm` both render blank because they sit directly under
`~/.hermes/skills/`. The frontmatter `category:` change is invisible there.

`.skills_prompt_snapshot.json` still held the old `(mtime_ns, size)`
`[1786134091919000904, 3304]` against the live `(1786141968918890863, 25050)`
immediately after the write — expected; the snapshot regenerates when the model
next builds a prompt. **So neither the count nor the manifest proves pickup at
apply time. The only real proof is behavioural, and the E2E provided it.**

---

## 8. E2E — the first live run, into mc-metarepo

**Trigger: a native Slack DM sent as Gaetan** (`chat.postMessage` to the home DM
`D0BBYNM01BL`, user token verified first with `auth.test` → `user_id
U08BDJAMSRZ`). Tokens were read from `~/.hermes/.env` into shell variables and
emitted through a `printf`-built `curl -K -` config on **stdin** — never on argv,
never written to disk. The claude.ai connector cannot trigger Tars
(`docs/facts.md`), so native was the only option that gets a response.

Message ts `1786142044.365879`, sent `2026-08-07T22:34:04Z`:

> Spin up a Claude session on cooper in mc-metarepo and have it map the repo
> layout for me: every top-level directory, one line each on what it holds, based
> on actually reading them. Strictly read-only - nothing written, nothing
> committed, nothing pushed. Tell me when it is done.

### Proof Tars did it — not me

`~/.hermes/logs/agent.log`, turn `20260807_223406_1a7fb8aa`, and the Slack
progress trace show, in order: `Reading skill delegate-to-cooper` → repeated
`terminal` tool executions of `rtk ssh cooper '~/.local/bin/orca orc…'` /
`'…orca wor…'` / `'…orca ter…'`. **Tars drove the `orca` CLI over ssh; it did not
call the v1 `delegate.sh` wrapper anywhere.** That is the behavioural proof that
v2 was loaded and in force. The model behind the turn is `gpt-5.6-sol` via
`openai-codex`.

### Ids and artifacts

| | |
|---|---|
| Run | `run_b08c7d3d6571` — objective *"Read-only map of every top-level directory in mc-metarepo"* |
| Task | `task_362a5c486ef9` — *"Map mc-metarepo top-level layout read-only"* |
| Dispatch | `ctx_fce35d74fac4` |
| Agent terminal | `term_02c27f84-e6a4-4056-b3c0-f6d832b6f5cb` |
| Worktree | `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/map-mc-metarepo-layout-readonly` |
| Branch | `GaetanCathelain/map-mc-metarepo-layout-readonly` |
| Outcome | `result.worker.state: **succeeded**`, `stage: settled`, dispatch `completed` at 22:41:28 |
| Reply message | `msg_1442a8fdefa5` — *"Full mc-metarepo map + audit (22 dirs)"* |

**Worktree KEPT, per the P3-a decision.** Nothing was pruned. mc-metarepo is now
at 10 worktrees.

The brief Tars wrote itself is good evidence the skill's "Writing the brief"
section works — it has Goal / Where / explicit read-only constraints / a
5-point "Done means" including *"Run `git status --short` at the start and at the
end and report both verbatim"*. Full text is in `task-list --run` output.

### Verdict: the loop closed, and the work was real

The delegated Claude agent mapped all 22 top-level directories, cited the
representative files it read per directory, and returned both `git status
--short` runs clean. Tars independently re-verified the checkout was untouched
(status, unstaged diff, staged diff all empty) before reporting. Repository
unmodified — the read-only constraint held.

---

## 9. Three defects the E2E found — the reason to run it

### 9.1 The proposal's load-bearing addressability claim is WRONG

P3 dismissed a mechanical reviewer's objection — that `check` binds to the
invoking coordinator terminal and a terminal-less ssh caller could never resolve
`--run` — as *"wrong, generalising from a stale id"*. **The reviewer was right.**

Tars hit it live, and I reproduced it independently from the VM over plain ssh
against the proposal's own probe run:

```
$ orca orchestration check --run run_24d000a3a6b6 --peek --json
  → { "ok": false, "error": { "code": "no_active_terminal" } }

$ orca orchestration task-list --run run_24d000a3a6b6 --json
  → { "ok": true, ... }          # the run IS real and resolvable
```

`run-create --help` confirms the missing piece: `Usage: orca orchestration
run-create --objective <text> [--from <handle>] [--json]`.

Measured requirements for a plain-ssh (terminal-less) caller:

| Verb | Needs |
|---|---|
| `run-create` | `--from <TERMINAL_HANDLE>` — else `no_active_sender_terminal` |
| `task-create` | `--from <TERMINAL_HANDLE>` — the run id alone does not supply sender context |
| `worker-start` | `--from <TERMINAL_HANDLE>` |
| `check` (incl. `--wait`, `--ack`) | `--terminal <TERMINAL_HANDLE>` — **even when `--run` is supplied** |
| `task-list --run`, `worker-list --run`, `worker-show --dispatch` | nothing extra — these DO work standalone |

The earlier measurement was almost certainly taken from a shell that already had
Orca terminal context (an agent running inside an Orca-managed terminal inherits
it), not from "exactly Tars's stateless position" as claimed.

**The skill's defensive fallback existed but could not fire.** It was keyed to
`run_not_found`:

> *"Only if `run_not_found` ever comes back on a run that `task-list --run` DOES
> resolve…"*

The actual error code is `no_active_terminal`. A fallback keyed to the wrong
error string is not a fallback.

### 9.2 `worker-start` reaching `ready` does NOT mean the agent started

The sharpest finding, and it came from **Gaetan watching in Slack**:

> *"Note: I opened the worktree you created on Orca, your prompt is there but
> it'd still need to be pressed 'enter' for it to start working"*

Orca had typed the brief into the Claude Code TUI and left it **unsubmitted**.
`workerState: ready` / `input_accepted` were both true and the agent had not
started. Left alone, `check --wait` would have timed out forever on a job that
had never begun — indistinguishable from "still running".

The fix, now in the skill: after spawn, check the terminal preview and send

```bash
orca terminal send --terminal "<AGENT_TERMINAL_HANDLE>" --enter --json
```

then confirm real activity via `worker-show` / `worktree ps` **before** waiting.
Once Enter was sent, the agent went to `state: working` and completed in ~5 min.

### 9.3 Tars self-edits its own SKILL.md — nobody planned for this

During the run Tars called `skill_manage` and **rewrote the file I had just
applied**, four times mid-run and three more times afterwards. This is a real
capability that neither the proposal, the spec, nor the mandate accounts for:
**a live skill file is not a stable artifact — the model mutates it.**

| Moment | md5 | Lines | Mode |
|---|---|---|---|
| As applied by me, 22:32:48Z | `929b23de94fd0697bf3ae2709f2dca42` | 437 | `-rw-rw-r--` |
| After mid-run self-edits, 22:35 | `e2d8aa9c7d16681114808d35686a7413` | 446 | `-rw-------` |
| After Gaetan's audit prompt, 22:46 | `60b9a24465eacda4736a0bb60fdf270e` | 452 | `-rw-------` |

Note `skill_manage` also **narrowed the mode from 0664 to 0600**.

The edits were **sanctioned in-thread** — Gaetan asked *"Did you update your orca
skill with those findings?"* and Tars audited the whole sequence and folded in
the `--from` / `--terminal` / `terminal send --enter` corrections. The content is
correct (I reproduced 9.1 independently). Both drifted versions are archived in
git for the record:

- `artifacts/delegate-to-cooper-SKILL.md` — canonical v2, what P3 approved
- `artifacts/delegate-to-cooper-SKILL-after-tars-selfedit.md` — live now, md5 `60b9a244…`, byte-identical to the VM

**Left exactly as Tars left it.** Reconciling git↔VM, and whether the model
should be able to rewrite its own governing skill at all, is Gaetan's call, not
mine. One concrete wart to fix if it is reconciled: the file now carries Tars'
correction *"Current Orca requires `--terminal` over plain SSH"* immediately
followed by the surviving, now-disproven sentence *"MEASURED 2026-08-07 from a
plain shell … answers `check --run … --peek` … with `ok:true`"*. Adjacent
contradictory claims will confuse the next read.

---

## 10. P3-c settled — the message/delivery shape

The mandate asked for the delivery-id field name on a **non-empty** batch. Read
from `orca orchestration inbox --json` (`check --peek` on the coordinator
terminal returned `count: 0` — see the gotcha below):

```json
{ "id": "msg_1442a8fdefa5",
  "run_id": "run_b08c7d3d6571",
  "delivery_contract": "current_delivery",
  "from_handle": "term_02c27f84-e6a4-4056-b3c0-f6d832b6f5cb",
  "to_handle": "run:run_b08c7d3d6571",
  "subject": "Full mc-metarepo map + audit (22 dirs)",
  "type": "status", "priority": "normal", "thread_id": null,
  "payload": "{\"taskId\":\"task_362a5c486ef9\",\"dispatchId\":\"ctx_fce35d74fac4\"}",
  "read": 1, "sequence": 2,
  "created_at": "2026-08-07T22:42:54Z", "delivered_at": null,
  "sender_pane_key": "65a1b252-…:041ef4ae-…" }
```

**There is no separate "delivery id" field.** The id to carry and `--ack` is the
message's own `id`, shaped `msg_<12 hex>`. Confirmed id shapes:

| Thing | Field path | Shape |
|---|---|---|
| Run | `result.run.id` | `run_<12 hex>` |
| Task | `tasks[].id` | `task_<12 hex>` |
| Dispatch | `result.dispatch.id` / `workers[].dispatchId` | **`ctx_<12 hex>`** — *not* `dispatch_…` as the skill's prose guessed |
| Message / delivery | `messages[].id` | `msg_<12 hex>` |
| Terminal | `agent_terminal_handle` | `term_<uuid>` |

**Gotcha found while settling this:** `check --terminal <coordinator-handle>
--run <runId> --peek` returned `count: 0` **despite the message existing**,
because the message is addressed `to_handle: "run:run_b08c7d3d6571"` — the run's
home inbox, not that terminal. `check` binds by terminal identity; the recipient
must match. `orchestration inbox --json` is what shows messages across
recipients.

**Also wrong in the skill:** its failure-mode table says to poll `worker-show
--dispatch <id> --json` → `result.state`. The real path is
**`result.worker.state`** (`result.dispatch.status` carries a separate
`completed` value). `result.state` does not exist.

---

## 11. Other measured corrections

- **`--agent codex` is not usable on cooper.** `orca account list --json` →
  `result.codex.accounts: []`, `activeAccountId: null`. Only `claude` has
  accounts (active `009bfa67-…`). The spec's *"only `claude` and `codex` are
  verified-usable"* overstates it.
- **The worktree landed at a path containing a literal `null` segment:**
  `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/map-mc-metarepo-layout-readonly`.
  Reported by Orca in `effects[].id` and in `worktree_id`, so it is Orca's own
  path construction, not a transcription slip. The skill's *"Predicted landing:
  `/home/gaetan/orca/workspaces/mc-metarepo/<slug>`"* is wrong, and its own
  advice — *read `path` from the JSON, never assume the pattern* — is what saved
  the run.
- **Skill-count prediction stale:** spec says 77 / 6-local; live is 80 enabled /
  7 local / 66 builtin / 7 hub-installed, unchanged across the apply.
- **`retainedReason: "user_takeover"`** appeared on the worker resource once
  Gaetan opened the tab in Orca — a state worth knowing before reading
  `terminalState: retained` as a fault.
- **Tars borrowed the hub's terminal as sender context.** `task_362a5c486ef9`
  records `created_by_process_incarnation:
  a71db494-…::/home/gaetan/orca/workspaces/Tars/orchestrator@@…` — to satisfy
  `--from`, Tars resolved a handle belonging to the **Tars/orchestrator**
  worktree. It works, but it couples a delegated run to whatever terminal
  happened to be open; a durable, intentional sender terminal is the open design
  question this raises.

---

## 12. Cooper-side v1 leftovers — nothing deleted

`~/orca/workspaces/tars-delegated/` is a plain directory (not a git repo).
`NOTE.md` was **appended** with a SUPERSEDED section (11 → 26 lines, the original
11 intact); `delegate.sh` and `.claude/settings.json` were read only and are
byte-identical. Full inventory, the verbatim appended text, and the
removable-later list with per-path risk notes are in
`status/probes/wf5/p3-cooper-v1-leftovers.md`. **Removal needs Gaetan's word.**

One flagged discrepancy, not investigated (read-only mandate): log
`20260807T202032Z.log` claims it wrote `wf5-selftest.md` and that a `rm` on it
was denied — so it should still exist — but the file is nowhere in the tree.

---

## 13. Final state

| | |
|---|---|
| Live skill | `~/.hermes/skills/delegate-to-cooper/SKILL.md`, md5 `60b9a24465eacda4736a0bb60fdf270e`, 452 lines, mode `0600` (narrowed by `skill_manage`) |
| Rollback to v1 | `SKILL.md.bak-p3` on the VM, md5 `d61888ec…`, intact — **and** `artifacts/delegate-to-cooper-SKILL-v1-live.md` in git |
| Rollback to canonical v2 | `artifacts/delegate-to-cooper-SKILL.md`, md5 `929b23de…` |
| Gateway | never restarted |
| `SOUL.md` / `config.yaml` / `lane-a.md` | never touched |
| Worktrees | kept (P3-a); mc-metarepo now at 10 |
| Orca footprint added | `run_b08c7d3d6571`, `task_362a5c486ef9`, `ctx_fce35d74fac4`, one worktree + branch |

**E2E verdict: PASS, with the loop completing only after a human pressed Enter.**
The skill loaded, Tars drove Orca over ssh, a real worktree and worker were
created in mc-metarepo, the delegated agent did real read-only work and reported
back, and `worker-read` retrieved it. The unattended path was broken by 9.2 and
is now fixed in the live skill — **that fix is unverified: no second run has
exercised `terminal send --enter` end to end.**

---

## 14. Follow-up on Gaetan's word: reconcile git↔VM, remove the leftovers

Both authorised explicitly after the E2E report.

### 14.1 git↔VM reconciled — `a6e0a4b53a11b9f79d635c2ecf6cadc2`

The live file and `artifacts/delegate-to-cooper-SKILL.md` now hold the same 512
lines. Written under `flock ~/.hermes/.wf3.lock` by atomic rename; **mode
restored `0600` → `0664`** (`skill_manage` had narrowed it). No gateway restart.
`hermes skills list` after: entry enabled, totals still 80 enabled / 7 local.

Base = Tars' self-edited version, because its structural fixes were right and I
verified the central one independently. Removed from it the stale claims it left
standing beside its own corrections:

| Fixed | Was |
|---|---|
| Deleted *"a freshly created run answers `check --run … --peek` … `--run` addressing works standalone; that is the design"* | Sat **directly beside** Tars' contradicting `--terminal` correction — the single most dangerous line in the file |
| Fallback trigger `no_active_terminal` | `run_not_found` — the wrong string, so the fallback could never fire |
| `result.worker.state` | `result.state` — does not exist |
| Id shapes marked CONFIRMED: `run_`/`task_`/`msg_` + **dispatch = `ctx_<hex>`** | "field name NOT confirmed live", and prose implying `dispatch_…` |
| "the delivery id IS the message's own `id`, `msg_<hex>`" | "has not been observed live — take whatever names the delivery" |
| Measured landing path (with its literal `null` segment) | A predicted `workspaces/mc-metarepo/<slug>` that is simply wrong |
| `claude` is the only usable `--agent` | "only `claude` and `codex`" |
| Failure table: added `no_active_terminal`, `no_active_sender_terminal`, the `count:0`-wrong-recipient case, and the unsubmitted-prompt trap | — |

**A live race, caught.** Between my archive copy (22:50, `60b9a244…`) and the
first reconcile write, Tars edited the file **again** at 22:55 (`5cc32dfc…`). I
had based the reconcile on the stale copy. Rather than clobber, I diffed the
generation I had missed and folded both of its findings back in:

- **Three self-check gates before `orca worktree rm --force`** (clean
  `--porcelain=v2 --untracked-files=all`, no unique commits via `rev-list
  --left-right --count`, worker settled), plus the measured fact that
  `worktree rm --force` can return `preservedBranch` and does **not** necessarily
  delete the local branch.
- **The worktree-level `preview` is unreliable** — it mixed stale
  dispatch-prompt lines with later output while the input line was empty and
  `agents[].state` was `done`. **This contradicted my own text**, which had told
  Tars to judge the unsubmitted-prompt case from the preview. The remedy now
  reads `terminal read --terminal <h> --limit 200 --json`, bottom `❯` line
  authoritative.

The second write carried an **md5 guard** that would abort rather than overwrite
a further concurrent edit. It passed; live was still my own `967aa3ee…`.

Rollback ladder on the VM, all three intact:

| File | md5 | What it is |
|---|---|---|
| `SKILL.md` | `a6e0a4b5…` | reconciled canonical, = git |
| `SKILL.md.bak-tars-selfedit` | `5cc32dfc…` | Tars' own final self-edit, 22:55 |
| `SKILL.md.bak-p3` | `d61888ec…` | v1, pre-P3 |

**Standing caveat: this can drift again at any time.** Tars holds `skill_manage`
and rewrote this file seven times in one turn. git↔VM parity is a snapshot, not
an invariant. Whether the model should be able to rewrite its own governing skill
is an open design question, not something this lane settled.

### 14.2 cooper v1 leftovers — REMOVED

`~/orca/workspaces/tars-delegated/` is gone (`rm -rf`, 2026-08-07T23:05Z).

Order of operations, deliberately: **archive → verify → commit → push → delete.**
Nothing was removed until its content was durable in git.

- Everything not already quoted verbatim in
  `status/probes/wf5/p3-cooper-v1-leftovers.md` was appended verbatim first: all
  six run logs, `first_10_primes.py`, and `NOTE.md` as it stood at removal.
  (`delegate.sh` and `.claude/settings.json` were already there verbatim.)
- A programmatic check confirmed **every** file's content is present in the doc
  before the delete ran — `ALL PRESENT`.
- Pre-delete checks: not a git repo (nothing tracked, no unpushed work), and
  `orca worktree list` had **0** references to it.
- Committed and pushed as `2087890` **before** the deletion.

Removed: `delegate.sh`, `.claude/settings.json`, `first_10_primes.py`, six
`logs/*.log`, `NOTE.md`, and the directory itself. `~/orca/workspaces/` now holds
only `mc-metarepo/` and `Tars/`.

One thing that stays unresolved, and was already flagged: log
`20260807T202032Z.log` claimed it wrote `wf5-selftest.md` and that a `rm` on it
was denied — so the file should have existed — but it was absent from the tree
before the deletion. That discrepancy is now unfalsifiable on disk; the log text
survives verbatim in the probe file.
