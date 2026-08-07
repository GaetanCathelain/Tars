# P3 — Tars self-commit rule: end-to-end live test (2026-08-07)

Orchestrated by the apply-p3-orca session; every action delegated to agents,
nothing on the commit path touched by hand. All times UTC (VM clock).

## Verdict

**PASS — twice, under two different flows.**

- **Phase A (rule as written this morning: direct push to main):** Tars
  self-edited `delegate-to-cooper` **4 times** during/after a second live
  delegation run and landed **one commit per edit, same turn, unprompted**
  (`d0d7091`, `eca0ec3`, `432a2c0`, `727c519`). Every commit's blob is
  byte-identical (md5) to the VM file at edit time, and Tars stated in Slack
  that it changed the skill and pushed. All three PASS criteria met.
- **Phase B (Gaetan's mid-test ruling: no direct push — branch → PR →
  auto-merge):** SOUL rule 2 was rewritten live (backup + flock + md5 guard,
  no gateway restart). In a fresh conversation Tars ran the new flow first
  try: branch `tars/delegate-to-cooper-20260807T234230Z` →
  **[PR #12](https://github.com/GaetanCathelain/Tars/pull/12)** →
  squash-merge `3b94c8f` (committer `GitHub <noreply@github.com>`), reply
  carried "Merged: <PR URL>". Parity held.
- **The `terminal send --enter` fix was NOT exercised** — worker-start
  submitted the prompt correctly first try, the premature-ready state never
  occurred. The fix **remains unverified end-to-end**; nothing failed.

## What was tested

SOUL rule 2's exception: immediately after any `skill_manage` write, in the
same turn, Tars must mirror the file to cooper's `~/dev/Tars` and land it on
`main` of github.com/GaetanCathelain/Tars — and say so in its reply. PASS
required (1) a new commit on `skills/delegate-to-cooper/SKILL.md` produced by
Tars's own flow, (2) VM md5 == git md5 afterwards, (3) the Slack claim.

Trigger: a second read-only delegation into mc-metarepo through the full v2
loop (summarize the five most recent commits), sent as Gaetan via native
`chat.postMessage` (tokens sourced from `~/.hermes/.env` on the VM into
`curl -K -` on stdin; `auth.test` verified `U08BDJAMSRZ` first).

## Timeline

| UTC | Event |
|---|---|
| 23:17:56 | Baseline: 5-way parity — VM/origin-main/worktree identical for SOUL (`13655395…`) and skill (`a6e0a4b5…`); origin/main `c574585`; `~/dev/Tars` clean |
| 23:20:36 | Trigger DM sent (ts `1786144836.409589`) |
| 23:20:43 | Tars starts: reads skill, orca preflight |
| 23:24:07 | `stale_delivery` on `msg_927cdf53c134` — the learning moment (ack wants top-level `delivery_*`, not `msg_*`) |
| 23:25:21–23:26:33 | 3× `skill_manage` → commits `d0d7091` 23:25:51, `eca0ec3` 23:26:14, `432a2c0` 23:26:40 — one commit per edit, same turn |
| 23:25:30, 23:26:08 | `git pull --rebase` fails twice ("unstaged changes") — flaw in the rule's verbatim flow (copy-before-pull); **Tars self-recovered both times** |
| 23:27:14 | "Done." reply: five-commit summary + "The correction was committed and pushed to Tars." Run `run_a0a34f5ec451`, task `task_7f2d40fa8a7a`, dispatch `ctx_08510f2221df` (follow-up `ctx_dd87a0699290`); worktree `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/latest-five-commits-readonly-2`, branch `GaetanCathelain/latest-five-commits-readonly-2`; `worker.state: succeeded`, `stage: settled` |
| ~23:33 | **Gaetan's mid-test ruling**: no direct push — Tars should commit, PR and merge automatically for its own repo |
| 23:36:41 | New SOUL rule 2 applied on VM (backup `SOUL.md.bak-prflow` md5 `13655395…`; flock; md5-guarded swap; new md5 `6d36025f…`; rules 1–8 intact; no restart) |
| 23:37:59 | Hub mirror commit `7c3dd8c` (SOUL.md + CLAUDE.md) — touches no `skills/` file |
| 23:39:46 | In-thread nudge: fold worker-show `observation.status` lag into the skill |
| 23:40:20 | Edit #4 (`abc5a443…`) lands as **direct push** `727c519` — old flow, despite new SOUL: **SOUL is snapshotted per conversation** (same thread = old rule). Claim: "Committed and pushed to the Tars repo." — truthful for the rule it had |
| 23:42:03 | Fresh top-level DM (new conversation, ts `1786146123.180139`): worktree pile-up note |
| 23:42:30–49 | Edit #5 (`cbccbf34…`) → branch `tars/delegate-to-cooper-20260807T234230Z` → PR #12 → squash-merge `3b94c8f` (+5/−0). Reply: "Updated `delegate-to-cooper`: … It warns only—no automatic deletion or blocking." + "Merged: <https://github.com/GaetanCathelain/Tars/pull/12>" |
| 23:47 | Final parity: VM skill `cbccbf34…` == `origin/main` blob; VM SOUL `6d36025f…` == mirror; `~/dev/Tars` clean on `main` @ `3b94c8f`, `tars/*` branch deleted |

## Edit → commit ledger (all blobs verified byte-identical to VM state)

| # | VM md5 after edit | Commit | Flow | Subject |
|---|---|---|---|---|
| 1 | `4d2e6fec756d2d108df3b8f3e7f93a74` | `d0d7091` | direct push | fix Orca delivery acknowledgement token contract |
| 2 | `03f019145b1c5457b5472eb176ac6087` | `eca0ec3` | direct push | correct ack token in reattach guidance |
| 3 | `318d2039ef7c2ee52e6926bebb560ee7` | `432a2c0` | direct push | fix stale MSG_ID ack example |
| 4 | `abc5a443eb991c73171704592ec5754b` | `727c519` | direct push (old-rule conversation) | trust worker state over lagging observation status |
| 5 | `cbccbf345260f0a0c72959e9e7db3157` | `3b94c8f` (PR #12) | branch → PR → squash-merge | flag mc-metarepo cleanup above 12 worktrees |

No commit on `skills/` was made by the test harness; the hub's only pushes
were `7c3dd8c` (SOUL/CLAUDE.md mirror) and this report.

## Findings beyond the verdict

1. **SOUL is snapshotted per conversation.** A SOUL edit does not reach an
   in-flight thread: the 23:36 rewrite was ignored at 23:40 in the run-2
   thread (old flow used) and honored at 23:42 in a fresh conversation. Same
   pattern explains the morning: 23:10 rule change → obeyed by the 23:20
   conversation. Operational consequence: after any SOUL change, behavior
   changes only for NEW conversations — don't expect it mid-thread, and don't
   diagnose non-compliance from an old thread.
2. **The original verbatim flow had a real bug Tars routed around**: it copies
   the file before `git pull --rebase`, so a moved origin/main makes the pull
   refuse the dirty tree. Hit twice (23:25:30, 23:26:08), both times the `&&`
   chain stopped safely and Tars recovered unprompted. The new rule pulls
   first; my own report push hit the identical failure mode, confirming it.
3. **The known suspect (`git commit` with nothing to commit breaking the `&&`
   chain) was never hit** — all five edits had real diffs. It remains latent
   in the new flow too (an unchanged file would die at `git commit` and no PR
   would be created); the rule's "if any step fails I say so plainly" covers
   the reporting side.
4. **`worker-show` returns `observation.status: "running"` after
   `result.worker.state: "succeeded"`** — stale secondary field; `worker.state`
   is authoritative. (Found by this test's evidence sweep; now folded into the
   skill by Tars itself as commit `727c519`.)
5. **Unprompted learning is real**: edits 1–3 were entirely spontaneous — a
   live `stale_delivery` error at 23:24:07 became three concrete skill
   corrections with commits within ~2.5 minutes, before the delegation even
   finished. The deterministic nudge path (planned fallback) was never needed
   for the original verdict.
6. **mc-metarepo is now at 11 kept worktrees** (run 2 added
   `latest-five-commits-readonly-2`; kept per P3-a). Tars's own new skill note
   flags cleanup when a delegation would push past 12.

## Flow change record (Phase B ruling)

- New SOUL rule 2 exception flow (verbatim now in `SOUL.md`, VM + mirror,
  md5 `6d36025f9c535e4e36749d502f81c64e`): pull main first → copy the skill
  file → `git checkout -b tars/<skill>-<utc>` → commit → push branch →
  `gh pr create --fill` → `gh pr merge --squash --delete-branch` → `MERGED`
  sentinel. Reply must name the change and give the PR URL. Never `--admin`.
- Preconditions verified before applying: cooper `gh` authed as
  GaetanCathelain (ADMIN), `main` unprotected, squash allowed.
- Backup on VM: `~/.hermes/SOUL.md.bak-prflow` (pre-change content,
  md5 `13655395899b9708dc638472be6b4e19`).
- `CLAUDE.md` updated in the same commit (`7c3dd8c`): the "never merges,
  approves or pushes" line now carries the scoped own-skill-PR exception.
- Executor stays cooper (`~/dev/Tars` + gh). The VM holds no GitHub
  credentials; moving git/gh onto the VM would be a separate provisioning
  decision.

## Appendix — Slack transcripts (verbatim, native API `conversations.replies`)

Sender `U08BDJAMSRZ` = Gaetan (native token, verified), `U0BBH85NAKH` = Tars.

### Thread 1 — run-2 trigger + phase-A/B nudge (root ts 1786144836.409589)

# run-2 thread transcript (channel D0BBYNM01BL, thread_ts 1786144836.409589)

Fetched: 2026-08-07T23:44:48Z
Messages: 7

## ts=1786144836.409589 user=U08BDJAMSRZ

Second run, same rules: spin up a Claude session on cooper in mc-metarepo and have it summarize the five most recent commits for me: one line each on what changed and why, based on actually reading the diffs. Strictly read-only - nothing written, nothing committed, nothing pushed. Tell me when it is done.


---

## ts=1786144843.588829 user=U0BBH85NAKH

:clipboard: kanban_show...
:books: Reading skill delegate-to-cooper
:books: Reading skill claude-code
:computer: terminal
```
rtk ssh cooper '~/.local/bin/orca sta...
```
```
rtk ssh cooper '~/.local/bin/orca acc...
```
```
rtk ssh cooper '~/.local/bin/orca wor...
```
```
rtk ssh cooper '~/.local/bin/orca ter...
```
```
rtk ssh cooper '~/.local/bin/orca wor...
```
```
rtk ssh cooper '~/.local/bin/orca ter...
```
```
rtk ssh cooper '~/.local/bin/orca orc...
``` (×4)
```
rtk ssh cooper '~/.local/bin/orca ter...
``` (×2)
```
rtk ssh cooper '~/.local/bin/orca orc...
```
```
rtk ssh cooper '~/.local/bin/orca ter...
```
```
rtk ssh cooper '~/.local/bin/orca orc...
``` (×2)
:book: Reading 1786145005_ssh_cooper____local_bin_or...
:computer: terminal
```
rtk ssh cooper '~/.local/bin/orca ter...
```
```
rtk ssh cooper '~/.local/bin/orca orc...
``` (×7)
```
rtk ssh cooper '~/.local/bin/orca ter...
```
```
rtk ssh cooper '~/.local/bin/orca orc...
```
:book: Reading 1786145083_ssh_cooper____local_bin_or...
:computer: terminal
```
rtk ssh cooper 'cd /home/gaetan/dev/m...
```
```
rtk ssh cooper '~/.local/bin/orca orc...
``` (×2)
:memo: Updating skill delegate-to-cooper
:computer: terminal
```
N=delegate-to-cooper; cat ~/.hermes/s...
```
```
rtk ssh cooper 'cd ~/dev/Tars &amp;&amp; git ...
```
```
rtk ssh cooper "cd ~/dev/Tars &amp;&amp; git ...
```
:memo: Updating skill delegate-to-cooper
:computer: terminal
```
N=delegate-to-cooper; cat ~/.hermes/s...
```
```
rtk ssh cooper "cd ~/dev/Tars &amp;&amp; git ...
```
:mag_right: Searching files for ack.*MSG|message's own|no separate de...
:memo: Updating skill delegate-to-cooper
:computer: terminal
```
N=delegate-to-cooper; cat ~/.hermes/s...
```

---

## ts=1786145017.869529 user=U0BBH85NAKH

:hourglass_flowing_sand: Working — 6 min — iteration 38/500, receiving stream response

---

## ts=1786145234.711819 user=U0BBH85NAKH

Done. Claude read the five diffs and reported:

- `4831253` *chore(submodules): auto-update to latest tracked branches (#181)​* — bumps the `monorepo-loop` submodule pointer `ed91de0a5d → 7970da7abb` (picks up a Loop box-aggregate feature and an STE docs commit); routine bot-driven tracked-branch sync.
- `8f88d6f` *chore(helptaker): seal duty days 2026-08-05 → 2026-08-07 (#180)​* — appends three entries to `scripts/helptaker.schedule.json` (`ahmad`, `nans`, `alexis`); freezes the last three HelpTech rotation days so the schedule is a durable record.
- `131a9ea` *chore(submodules): auto-update to latest tracked branches (#179)​* — bumps seven submodule pointers (`cleaq-api`, `cleaq-front`, `monorepo-loop`, `pim-migration-sync`, `return-app`, `risk`, `support-engineer`), most pulling each repo’s Simplified Technical English docs commit; propagates the STE rule rollout across the stacks.
- `52cefc9` *feat(scripts, skills): enforce ASD-STE100 Simplified Technical English (#175)​* — adds `scripts/ste-lint.mjs`, a diff-scoped CI gate, the `write-ste` skill, and supporting documentation; makes STE enforceable on new prose without blocking on 1,906 existing violations.
- `7171d1b` *chore(submodules): auto-update to latest tracked branches (#176)​* — bumps `monorepo-loop` for accessory-provisioning tests and the MC-4174 RIB fix, plus `support-engineer` for the NMC-613 automatic Care/Ops ask; routine tracked-branch sync.

Verified independently after completion: HEAD remained `4831253a24928e57b76fcb2870df95705706b6f2`, status was empty, and Claude wrote, committed, and pushed nothing. Orca created only the session worktree/branch infrastructure.

*Delegation command and brief:​*
```
ssh cooper '~/.local/bin/orca orchestration task-create --run "run_a0a34f5ec451" --from "term_479c7b7b-7978-4c35-8a98-ad9db0904eb2" --task-title "Read-only summary of latest five commits" --spec "Goal: In mc-metarepo, summarize the five most recent commits on the checked-out HEAD, one line per commit, based on actually reading each full diff. Work strictly read-only. Do not edit, create, delete, rename, stage, commit, push, fetch, pull, checkout, switch, reset, clean, install, format, or run any command that mutates files, refs, the index, or repository state. Use only read-only inspection commands, with GIT_OPTIONAL_LOCKS=0 for Git commands. Record HEAD and git status --porcelain=v2 --untracked-files=all before inspection and again afterward; they must be identical. Inspect exactly HEAD through HEAD~4 using git show/diff, including commit messages and full patches, and infer why from the diff plus message without inventing intent. Output exactly five commit-summary lines, newest first, each formatted: &lt;short SHA&gt; &lt;subject&gt; — &lt;what changed&gt;; &lt;why&gt;. Then add a separate verification sentence stating whether HEAD and status remained unchanged. Do not write any report file; return the result only in the Claude session." --json'

ssh cooper '~/.local/bin/orca orchestration worker-start --run "run_a0a34f5ec451" --task "task_7f2d40fa8a7a" --from "term_479c7b7b-7978-4c35-8a98-ad9db0904eb2" --worktree new-top-level --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --name "latest-five-commits-readonly-2" --agent claude --setup run --timeout-ms 200000 --json'
```

- Run: `run_a0a34f5ec451`
- Primary dispatch: `ctx_08510f2221df`
- Exact-output follow-up dispatch, same Claude session: `ctx_dd87a0699290`
- Deliveries: `delivery_82d6f8351da6`, `delivery_7c978c4a0c9c`
- Worktree: `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/latest-five-commits-readonly-2`
- Branch: `GaetanCathelain/latest-five-commits-readonly-2`
- Outcome: `succeeded`; terminal released

I also corrected the `delegate-to-cooper` skill: current Orca requires the top-level `delivery_*` token for acknowledgements, not the nested `msg_*` ID. The correction was committed and pushed to Tars.

---

## ts=1786145986.202339 user=U08BDJAMSRZ

Nice. One correction to fold into your delegate-to-cooper skill: worker-show's observation.status can lag at "running" even when result.worker.state is already "succeeded" - worker.state is the authoritative field, don't trust observation.status. Update the skill.


---

## ts=1786145996.980119 user=U0BBH85NAKH

:mag_right: Searching files for observation|worker\.state|worker-show...
:book: Reading SKILL.md L225-259
:memo: Updating skill delegate-to-cooper
:computer: terminal
```
N=delegate-to-cooper; cat ~/.hermes/s...
```

---

## ts=1786146027.240749 user=U0BBH85NAKH

Updated `delegate-to-cooper`:

- `result.worker.state` is now explicitly authoritative.
- `result.observation.status` may lag at `running` after the worker has succeeded and must not determine completion.

Committed and pushed to the Tars repo.

---


### Thread 2 — fresh conversation, PR-flow verification (root ts 1786146123.180139)

# B-2 thread transcript (channel D0BBYNM01BL, thread_ts 1786146123.180139)

Fetched: 2026-08-07T23:44:48Z
Messages: 3

## ts=1786146123.180139 user=U08BDJAMSRZ

Separate small thing: your delegate-to-cooper skill should track worktree pile-up - mc-metarepo is at 11 kept worktrees after your last run. Fold a note into the skill: flag cleanup to me when a delegation would push it past 12.


---

## ts=1786146128.867669 user=U0BBH85NAKH

:clipboard: kanban_show...
:books: Reading skill delegate-to-cooper
:memo: Updating skill delegate-to-cooper
:computer: terminal
```
N=delegate-to-cooper ...
```
```
rtk ssh cooper 'command -v gh || for ...
```
```
rtk ssh cooper 'cd ~/dev/Tars &amp;&amp; B=$(...
```
```
rtk ssh cooper 'cd ~/dev/Tars &amp;&amp; /hom...
```

---

## ts=1786146184.221049 user=U0BBH85NAKH

Updated `delegate-to-cooper`: every `mc-metarepo` delegation now checks kept worktrees first and flags cleanup when adding one would exceed 12. It warns only—no automatic deletion or blocking.

Merged: <https://github.com/GaetanCathelain/Tars/pull/12>

---

