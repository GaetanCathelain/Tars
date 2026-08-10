# WF6 — daily brief board survives the REAL cron invocation path

**Date:** 2026-08-10 (VM clock UTC) · **Verdict: PASS (a–f)**

The last open WF6 question: spec check 3 proved the board block byte-identical
under `hermes chat`, but the production path is different — cron prompt
delivery, `HERMES_CRON_AUTO_DELIVER_*`, a real Slack delivery target. Running
`hermes cron run e231e5faf180` would have posted to Gaetan's production
reporting thread, so it was deliberately never run.

This probe ran the real cron path through a **temporary mirror job** delivering
to the **channel top level** of `#gcn-tars-reporting`, leaving the production
job and the production anchor thread untouched.

---

## 1. CLI facts (measured on the VM before scripting)

`~/.local/bin/hermes cron --help` subcommands:
`list, create|add, edit, pause, resume, run, remove|rm|delete, status, runs|history, notepad, tick`

```
usage: hermes cron create [-h] [--name NAME] [--deliver DELIVER]
                          [--repeat REPEAT] [--skill SKILLS] [--script SCRIPT]
                          [--no-agent] [--monitor-script MONITOR_SCRIPT]
                          [--monitor-url MONITOR_URL] [--workdir WORKDIR]
                          [--model MODEL] [--provider MODEL_PROVIDER]
                          schedule [prompt]
  --deliver DELIVER   Delivery target: origin, local, telegram, discord,
                      signal, or platform:chat_id
  --skill SKILLS      Attach a skill. Repeat to add multiple skills.

usage: hermes cron run [-h] [--accept-hooks] job_id
  job_id          Job ID to trigger

usage: hermes cron remove [-h] job_id
```

There is no `show`/`detail` subcommand — `cron list` does not print prompts.
The full job records live in `~/.hermes/cron/jobs.json` (`{"jobs": [...]}`),
which is where the verbatim prompt and every comparable field were read.

## 2. Production job read (step 1) — `e231e5faf180`

| field | value |
|---|---|
| `name` | `Gaetan daily work brief` |
| `schedule_display` / `schedule.expr` | `30 8 * * 1-5` (kind `cron`) |
| `deliver` | `slack:C0BP2GZUFSR:1786359613.979759` |
| `skills` / `skill` | `["daily-work-brief"]` / `daily-work-brief` |
| `enabled` | `true` (state `scheduled`) |
| `no_agent` / `script` | `false` / `null` |
| `model` / `provider` | `null` / `null` (snapshots `gpt-5.6-sol` / `openai-codex`) |
| `origin` | slack `D0BBYNM01BL`, thread `1786191017.826669`, user `U08BDJAMSRZ` |
| `prompt` sha256 (first 16) | `a150b59bf58b64c7` |

Prompt verbatim (as stored, unescaped):

> Run `daily-work-brief` end to end. Use Europe/Paris. Establish the window from the last successful delivery, then apply the workday gate — weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off — and return exactly `[SILENT]` when the gate says not to run. Otherwise collect Slack, Cooper (Orca state, Claude transcripts, shell history, git), GitHub, Linear, email and calendar evidence in parallel, keep raw results out of the message, and deliver the compact brief. Do Linear LAST. The brief must OPEN with the board block, and **the board block is the stdout of `python3 ${HERMES_HOME:-$HOME/.hermes}/skills/orchestration/daily-work-brief/scripts/linear_board.py`, run through the shell/terminal tool (never `execute_code`, which scrubs `LINEAR_API_KEY`) and included byte for byte.** Do not write, correct, re-order, re-title, filter or re-type a single row: the script already sorted, capped, computed `+N more` and emitted the `Coverage:` line. If it printed `board unavailable (coverage unproven)`, print exactly that one line as the board and nothing else. If its output is not in front of you — never run, scrolled away, or the context compacted — run it again rather than recalling rows; a fabricated board is worse than no board. Take `Linear tickets completed: N` from a separate native `mcp__linear__list_issues` read (`state` `"completed"`, `updatedAt` at the window start, `fields` `["id","title","completedAt"]`), proven complete by paging `cursor` until `hasNextPage == false` — never by narrowing the filter. `Since the last daily` must end with the labelled `Claude/Orca:` line covering the Cooper Claude/Orca history, or `Claude/Orca: none.` Every source is read-only: write nothing to Linear, and do not spawn, prompt, or delegate to another agent. Return only the brief.

## 3. Temporary mirror job (step 3)

Created by reading the production prompt straight out of `jobs.json` and passing
it as argv — no retyping, no interpolation:

```python
j = [v for v in json.load(open('~/.hermes/cron/jobs.json'))['jobs']
     if v['id'] == 'e231e5faf180'][0]
cmd = ['/home/gaetan/.local/bin/hermes', 'cron', 'create', '0 4 1 1 *', j['prompt'],
       '--name', 'wf6-verify-temp daily brief mirror',
       '--deliver', 'slack:C0BP2GZUFSR',
       '--repeat', '1']
for s in j['skills']:
    cmd += ['--skill', s]
```

```
Created job: 2734e0e5f2fa
  Name: wf6-verify-temp daily brief mirror
  Schedule: 0 4 1 1 *
  Skills: daily-work-brief
  Next run: 2027-01-01T05:00:00+01:00
```

Stored definition verified against production:

```
prompt_identical True a150b59bf58b64c7
temp_deliver slack:C0BP2GZUFSR
temp_skills ['daily-work-brief'] daily-work-brief
temp_enabled True repeat {'times': 1, 'completed': 0} sched 0 4 1 1 *
temp_model None None noagent False script None
prod_deliver_unchanged slack:C0BP2GZUFSR:1786359613.979759
```

Same prompt (byte-identical), same skill, same model/provider defaults, same
`no_agent`/`script`. Only the delivery target differs: channel top level
`slack:C0BP2GZUFSR`, **no** `:1786359613.979759` thread suffix. Schedule
`0 4 1 1 *` (1 January 04:00) plus `--repeat 1` so it could never fire on its
own inside the probe window.

## 4. Reference board — `linear_board.py` direct, immediately pre-trigger (step 4)

```
ssh gaetan@192.168.0.9 'set -a; . ~/.hermes/.env; set +a; \
  python3 ~/.hermes/skills/orchestration/daily-work-brief/scripts/linear_board.py'
```

exit 0, empty stderr, 1299 bytes, **sha256
`0392f6923a3bbc0098c0cda84aad8e22b62e47a4d7929bd1f1925c320146ddb9`**

```
**Board** (priority-sorted)
P1 MC-4179 Cliente bloquee pour son echange — Triage
P1 NMC-601 Rotate the leaked GoCardless E1 production token (prod log leak follow-through) — In Progress
P1 NMC-622 Staging deploys are silently broken: Deploy Hasura + Deploy Hasura Admin fail ECS stabilisation, so no api-v2 code or H… — In Progress
P2 GCN-3 engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard — In Progress
P2 GCN-4 daily-work-brief: pull Linear board + render priority-sorted — In Progress
P2 GCN-5 Tars default-team rule: create tickets in GCN unless told otherwise — In Progress
P2 GCN-7 WF6 E2E verification + docs/facts + status log updates — In Progress
P2 MC-4112 Datadog observability E2E for Verdict (ECS + Vercel + Cleaq distributed tracing) — Todo
P2 NMC-497 Block manual "stop invoicing" on Loop for non-tech operators (167 leaking cases) — In Progress
P2 NMC-498 Gate manual stop-invoicing: Hasura permission revoke + api-v2 resolver guard + admin UI — In Progress
P2 NMC-606 A low-confidence investigation automatically triggers an adversarial review pass — In Review
P2 NMC-615 SE never opened a single metarepo learnings PR despite the capability being Done — investigate + fix — In Review
+46 more
Coverage: 2 views, both complete (58 issues)
```

## 5. Trigger through the real path (step 5)

```
$ date -u +%FT%TZ ; ~/.local/bin/hermes cron run 2734e0e5f2fa
2026-08-10T21:53:15Z
  ⟳ compacting context…
Triggered job: wf6-verify-temp daily brief mirror (2734e0e5f2fa)
  Ran now: succeeded.
```

Real cron session in `~/.hermes/logs/agent.log`: `cron_2734e0e5f2fa_20260810_235316`
(75 log lines). The run **compacted its context mid-flight** —
`context compression started: session=cron_2734e0e5f2fa_20260810_235316
messages=65 tokens=~136,530` — which is exactly the condition the anti-fabrication
clause exists for.

Job state after the run: `state=completed status=ok
last_run=2026-08-10T23:59:47.718838+02:00 err=none deliv_err=none`.
Durable artifact: `~/.hermes/cron/output/2734e0e5f2fa/2026-08-10_23-59-47.md`
(24406 B; scanned for `lin_api_|xox[bpade]-|sk-…` → 0 hits).

## 6. What landed in `C0BP2GZUFSR` (step 6)

**Top-level message, ts `1786399187.527289`**, from Tars `U0BBH85NAKH`,
2026-08-10 23:59:47 CEST. Returned by `conversations.history` on the channel
(which excludes thread replies) — i.e. it is a channel-level post, not a reply.

The pre-delivery agent output (section `## Response` of the cron output file,
35 lines / 3311 bytes) verbatim:

```
**Board** (priority-sorted)
P1 MC-4179 Cliente bloquee pour son echange — Triage
P1 NMC-601 Rotate the leaked GoCardless E1 production token (prod log leak follow-through) — In Progress
P1 NMC-622 Staging deploys are silently broken: Deploy Hasura + Deploy Hasura Admin fail ECS stabilisation, so no api-v2 code or H… — In Progress
P2 GCN-3 engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard — In Progress
P2 GCN-4 daily-work-brief: pull Linear board + render priority-sorted — In Progress
P2 GCN-5 Tars default-team rule: create tickets in GCN unless told otherwise — In Progress
P2 GCN-7 WF6 E2E verification + docs/facts + status log updates — In Progress
P2 MC-4112 Datadog observability E2E for Verdict (ECS + Vercel + Cleaq distributed tracing) — Todo
P2 NMC-497 Block manual "stop invoicing" on Loop for non-tech operators (167 leaking cases) — In Progress
P2 NMC-498 Gate manual stop-invoicing: Hasura permission revoke + api-v2 resolver guard + admin UI — In Progress
P2 NMC-606 A low-confidence investigation automatically triggers an adversarial review pass — In Review
P2 NMC-615 SE never opened a single metarepo learnings PR despite the capability being Done — investigate + fix — In Review
+46 more
Coverage: 2 views, both complete (58 issues)

**Since the last daily**
- **Monitoring/security:** The weekly production review led to a merged nested-secret sanitizer for GoCardless, HelpScout and Yousign and a monitoring-rules update. NMC-601 remains open for GoCardless rotation follow-through.
- **Infrastructure/support:** Loop development jobs were restored and their ingress fix merged; ArgoCD automation is back. The development cluster then lost schedulable capacity, leaving workloads pending and the webhook at 503. Datahub RDS access was made Terraform-durable, and the HT console submodule-permission fix was deployed.
- **Tars:** Slack tool access/visibility was corrected and live-verified. Native Linear MCP, GCN defaults, deterministic daily-board rendering, report threading and Himalaya fixes landed.
- **Coordination:** The AI-subscription pricing report was delivered to Oli; a Hermes/life-mirror session is booked for Wednesday.
- **Claude/Orca:** `Tars/wf6-vm-skills` is still working; `graphify-benchmark` is parked after 93/93 cells with T5–T7 judging and the report left; `mc-kestra` verified seven Gbrain flows and both processing chains; the HT-console and Loop-job worktrees completed through PRs #169 and #121.

**Stats**
- PRs merged: **21** — Tars #35–43; datahub #72; helm-charts #121; metarepo #193/#197; support-engineer #163–167/#169; workspace #12287/#12291.
- Linear tickets completed: **7** — GCN-1, GCN-2, GCN-6, GCN-9, SE-680, SE-681, SE-683.

**Today**
1. Restore development-cluster capacity; this is the active blocker behind the 503 and pending workloads.
2. Close NMC-601's token-rotation follow-through and investigate the unresolved Google suspicious-login alert.
3. Finish WF6 via GCN-3/4/5/7; keep the Graphify benchmark parked until those production controls are closed.
4. **Oli:** no repeat AI-pricing report; escalate the development-cluster outage if capacity remains unavailable.

Coverage: Calendar unavailable; configured email was checked and contained no explicit PayFit leave notice.
```

As rendered in Slack the only difference is the transport-level mrkdwn
conversion Hermes applies to the whole message on delivery — `**Board**` →
`*Board*`, `**21**` → `*21*` etc. Every board **row**, the `+46 more` line and
the `Coverage:` line are plain text and pass through the Slack layer unchanged
(confirmed against the `conversations.history` copy of ts `1786399187.527289`).
The byte-identity test below is therefore made against the agent's
pre-delivery output, which is what the anti-fabrication clause governs.

## 7. The byte-diff (pass criterion b)

```
$ sha256sum board-ref.txt board-delivered.txt
0392f6923a3bbc0098c0cda84aad8e22b62e47a4d7929bd1f1925c320146ddb9  board-ref.txt
0392f6923a3bbc0098c0cda84aad8e22b62e47a4d7929bd1f1925c320146ddb9  board-delivered.txt

$ wc -c board-ref.txt board-delivered.txt
1299 board-ref.txt
1299 board-delivered.txt

$ diff -u board-ref.txt board-delivered.txt
(no output)

$ cmp board-ref.txt board-delivered.txt
BYTE IDENTICAL
```

Same sha256, same length, empty diff, `cmp` silent — across a mid-run context
compaction, on the real cron path. No character differs.

## 8. Production untouched

### Anchor thread `1786359613.979759` (criterion e)

`conversations.replies` on `C0BP2GZUFSR` / `1786359613.979759` with
`oldest = 1786398752.000000` (a cutoff taken before the temp job was created):

- **Before the run:** `No thread messsages` — parent only.
- **After the run:** `No thread messsages` — parent only.

The production thread received nothing. (Its full pre-existing history, last
reply `• Send Jamil the Verdict Google OAuth client … [EC-A44C]`, is unchanged.)

### `C0BFQ5WFYTB` `#tech-project-support-engineer` (criterion f)

Same cutoff, `conversations.history`:

- **Before:** no messages.
- **After:** no messages.

Nothing landed there. It was never a delivery target of the temp job.

### The three production jobs (step 8)

Full JSON records of `e231e5faf180`, `62e8cd9db637`, `759e08c598e3` were dumped
from `jobs.json` before the temp job was created and again after it was removed,
then diffed field-for-field:

```
$ diff -u prod-jobs-before.json prod-jobs-after.json
IDENTICAL (all fields, all 3 jobs)
```

That covers every field, not just the named ones — `prompt`, `deliver`,
`schedule`/`schedule_display`, `skill`/`skills`, `enabled`, `state`, `repeat`,
`model`/`provider`, `origin`, `next_run_at`, `last_run_at`, `last_status`,
`no_agent`, `script`, `workdir`, `monitor_*`, `paused_*`, `last_error`,
`last_delivery_error`. All three still `[active]`, delivering to
`slack:C0BP2GZUFSR:1786359613.979759`.

## 9. Cleanup (step 7)

```
$ ~/.local/bin/hermes cron remove 2734e0e5f2fa
Removed job: wf6-verify-temp daily brief mirror (2734e0e5f2fa)

$ ~/.local/bin/hermes cron list | grep -c 2734e0e5f2fa
0

# jobs.json cross-check
TEMP_PRESENT False
```

Gone from both `cron list` and the on-disk store.

## 10. Gateway (step 9)

```
before:  ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
after:   ActiveState=active
         SubState=running
         ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

Identical `ActiveEnterTimestamp` → **no restart** across the probe.
(`NRestarts` deliberately not cited — it is known unreliable here.)

## 11. Verdict

| # | criterion | result |
|---|---|---|
| a | delivered brief OPENS with the board block | **PASS** — `**Board** (priority-sorted)` is the first content of the message |
| b | board block byte-identical to `linear_board.py` stdout | **PASS** — sha256 `0392f69…` both sides, 1299 B, empty `diff`, `cmp` clean |
| c | `Coverage:` line present | **PASS** — `Coverage: 2 views, both complete (58 issues)` in the board (plus the trailing source-coverage line) |
| d | Orca/Claude-history section still present | **PASS** — `- **Claude/Orca:** …` closes *Since the last daily* |
| e | anchor thread `1786359613.979759` received nothing | **PASS** — parent-only before and after |
| f | nothing landed in `C0BFQ5WFYTB` | **PASS** — empty before and after |

Extra hardening this run gives beyond spec check 3: the board survived a
**mid-run context compaction on the real cron path**, with cron prompt delivery
and a real Slack delivery target — the exact conditions `hermes chat` could not
reproduce.

## 12. Note for Gaetan

The verification post is **`C0BP2GZUFSR` ts `1786399187.527289`** (2026-08-10
23:59:47 CEST, top level of `#gcn-tars-reporting`). It is a real, correct daily
brief; posts there are sanctioned, so it was left in place. Delete it if you
don't want a second brief for 2026-08-10 in the channel.
