# WF7 deploy — gmail-triage shipped to the Tars VM, live run proven (GCN-44)

Session: Orca subagent on cooper, 2026-08-12. VM clock is UTC; Hermes'
configured timezone is `Europe/Paris` (`config.yaml: timezone: Europe/Paris`),
so both are quoted below. **No `sops -d` was run, no secret value was read or
printed, `config.yaml` was not edited, `status/lane-a.md` was not touched.**

## 1. Ship — `~/.hermes/skills/email/gmail-triage/SKILL.md`

Pre-state (14:29:38 UTC): `~/.hermes/skills/email/gmail-triage/` did not exist
(`ls: cannot access … No such file or directory`), and neither did
`~/.hermes/state/gmail-triage.json`. **No live file to compare against** — the
first deploy, so no `.bak` and no stop-to-compare.

```
ssh gaetan@192.168.0.9 'mkdir -p ~/.hermes/skills/email/gmail-triage'
cat skills/email/gmail-triage/SKILL.md | ssh gaetan@192.168.0.9 \
  'umask 077; cat > ~/.hermes/skills/email/gmail-triage/SKILL.md.new'
ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c \
  "mv …/SKILL.md.new …/SKILL.md"'
```

md5 verification, 14:30:36 UTC:

```
repo  8eaa93e425c3b5a4ed1ffb9f7107b84b  skills/email/gmail-triage/SKILL.md
live  8eaa93e425c3b5a4ed1ffb9f7107b84b  /home/gaetan/.hermes/skills/email/gmail-triage/SKILL.md
-rw------- 1 gaetan gaetan 20562 Aug 12 14:30 SKILL.md
```

**MATCH.** Repo copy == live copy, and stayed so (0 skill iterations, see §6).

Gateway not restarted by the live-reload, before and after:
`ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC`, `NRestarts=0`.

## 2. Cron job — id `191332ae6a24`

Delivery target resolved from source before creating anything, so that
"same target as the engagement-checker" did not mean "same *literal string*":

- engagement-checker (`62e8cd9db637`) delivers to
  `slack:C0BP2GZUFSR:1786516435.025349` — **not** a DM: that is the *daily
  rotating report thread*, whose `parent_ts` is rewritten every day by the
  "Daily report thread reconciler" job (`~/.hermes/state/daily-report-thread.json`
  → `{"channel": "C0BP2GZUFSR", "date": "2026-08-12", "parent_ts": "1786516435.025349"}`).
  Copying that literal value would have pinned Gmail triage to *today's* thread
  forever and posted into a channel.
- `cron/scheduler.py:268` — `_HOME_TARGET_ENV_VARS = { … "slack": "SLACK_HOME_CHANNEL" … }`,
  and `~/.hermes/.env` has `SLACK_HOME_CHANNEL=D0BBYNM01BL` = Gaetan's DM.
- So `--deliver slack` **is** the Slack DM to Gaetan, exactly as
  `SKILL.md § Scheduling contract` and the spec state. Confirmed at runtime:
  `cron.scheduler: Job '191332ae6a24': delivered to slack:D0BBYNM01BL`.

Creation command (2026-08-12 16:31:44+02:00). **`hermes cron create` needs both
positionals adjacent** — `create SCHED --name … --skill … --deliver … PROMPT`
fails with `hermes: error: unrecognized arguments: Run \`gmail-triage\`…`
(argparse eats the optional `prompt` positional in the first block). Prompt was
staged to `/tmp/wf7-prompt.txt` (md5 `23ca5c218ad55e741884b0931c68c536`, 814 B,
verbatim from `SKILL.md § Scheduling contract`) and passed as `"$(cat …)"` so
its backticks are not re-evaluated:

```
~/.local/bin/hermes cron create "0 7-21 * * *" "$(cat /tmp/wf7-prompt.txt)" \
  --name "Gmail triage" --skill gmail-triage --deliver slack
→ Created job: 191332ae6a24
  Schedule: 0 7-21 * * *   Skills: gmail-triage
  Next run: 2026-08-12T17:00:00+02:00
```

On-disk record, `~/.hermes/cron/jobs.json` (prompt elided — it is the
SKILL.md block verbatim):

```json
{
 "id": "191332ae6a24",
 "name": "Gmail triage",
 "prompt": "Run `gmail-triage` end to end. Use Europe/Paris and the run's fixed timestamp. …",
 "skills": ["gmail-triage"],
 "skill": "gmail-triage",
 "model": null, "provider": null,
 "provider_snapshot": "openai-codex", "model_snapshot": "gpt-5.6-sol",
 "base_url": null, "script": null, "no_agent": false,
 "monitor_script": null, "monitor_url": null, "monitor_state": null,
 "context_from": null,
 "schedule": {"kind": "cron", "expr": "0 7-21 * * *", "display": "0 7-21 * * *"},
 "schedule_display": "0 7-21 * * *",
 "repeat": {"times": null, "completed": 0},
 "enabled": true, "state": "scheduled",
 "paused_at": null, "paused_reason": null,
 "created_at": "2026-08-12T16:31:44.046105+02:00",
 "next_run_at": "2026-08-12T17:00:00+02:00",
 "last_run_at": null, "last_status": null, "last_error": null,
 "last_delivery_error": null,
 "deliver": "slack",
 "origin": null,
 "enabled_toolsets": null, "workdir": null
}
```

`next_run_at` = `17:00:00+02:00` from a `0 7-21 * * *` expr created at
`16:31+02:00` — **the expression is interpreted in Europe/Paris**, as the spec
claims. No timezone field exists on the job; none was invented.

`hermes cron list` shows `191332ae6a24 [active] … Deliver: slack … Skills: gmail-triage`.

## 3. Trigger method

`hermes cron run 191332ae6a24` (the probe's mechanism; it flags the job for the
next scheduler tick — ticker heartbeat was 24 s at the time, so it fires within
a minute and returns `Ran now: succeeded.`). Runs took **~80 s**, not minutes.

## 4. Run 1 — 16:31:57 → 16:33:16 +02:00 (14:31:57 → 14:33:16 UTC)

Execution `dfac81783b474d6a89c3b88626401d5d`, `status=completed`,
output `~/.hermes/cron/output/191332ae6a24/2026-08-12_16-33-16.md`.

State file **created from the skill's literal shape** —
`~/.hermes/state/gmail-triage.json`, 730 B, all eight label IDs present:

```json
{"version": 1,
 "cursor": "2026-08-12T16:32:20+02:00",
 "last_completed_run": "2026-08-12T16:32:20+02:00",
 "seen": [],
 "labels": {"🛠️ Outils & SaaS": "Label_6", "💳 Finance & Facturation": "Label_7",
            "👥 Interne Mobile Club": "Label_8", "🤝 Partenaires & Prestataires": "Label_9",
            "📅 Réunions & Agenda": "Label_10", "🎯 Recrutement": "Label_11",
            "🔐 Sécurité & Comptes": "Label_12", "📣 Newsletters & Pub": "Label_13"},
 "last_run": {"at": "2026-08-12T16:32:20+02:00", "candidates": 0, "labelled": 0,
              "unrouted": 0, "overflow": 0, "by_label": {}},
 "consecutive_failures": 0, "last_failure_notice": {}, "unrouted_senders": {}}
```

Delivered response (whole of it):

```
## Response
• State file was missing and recreated; cursor reset to 2026-08-12 16:32 Paris. No historical mail was backfilled.
```

Scheduler: `2026-08-12 14:33:16,308 INFO cron.scheduler: Job '191332ae6a24': delivered to slack:D0BBYNM01BL`.

**Did it actually look at the mailbox?** `agent.log` records tool completions
without the command text, so the proof is by output size. Run 1's tool trace:

```
tool terminal completed (0.12s, 116 chars)     ← lock dir + timestamp
tool write_file completed (0.06s, 235 chars)   ← state file
tool terminal completed (0.83s, 1362 chars)    ← the §2 search
tool patch completed (0.05s, 564 chars)        ← state fields
tool terminal completed (0.08s, 53 chars)      ← lock release
```

Independent read-only re-run of the skill's own §2 query at 14:34 UTC:

```
$GAPI gmail search "in:inbox -has:userlabels after:2026/08/12" --max 100
real 0m0.745s   exit=0   → 1296 bytes, 2 messages
  19ff4c90892113cb | Wed, 12 Aug 2026 02:04:03 -0500 | …@dynatrace.ai
  19ff456661cc3ad3 | Wed, 12 Aug 2026 04:58:50 +0000 | …@email.claude.com
```

1296 B of stdout + the terminal tool's command echo ≈ the **1362 chars** logged,
and 0.745 s ≈ the **0.83 s** logged. Run 1 ran that search. It found the same
two messages, both dated **before** the freshly self-healed cursor (09:04 and
06:58 Paris vs 16:32), and correctly labelled **nothing** — the skill's "do not
backfill after a state loss" rule. `candidates: 0` is that rule, not a dead run.

## 5. Run 2 — the labelling path, 16:36:00 → 16:37:03 +02:00

Run 1 left an empty window (cursor = now), so the `modify` path was still
unexercised. To exercise it without a backfill: the state file was **backed up**
(`gmail-triage.json.bak-gcn44-20260812T143556Z`) and its `cursor` rolled back to
`2026-08-12T00:00:00+02:00` — today only, the two known candidates, no history.
Nothing else was changed.

Execution `9f419252bea941728ac796e6f5c03c12`, `status=completed`. Tool trace:

```
tool terminal completed  (0.14s, 133 chars)   ← lock + timestamp
tool read_file completed (0.05s, 982 chars)   ← existing state read
tool terminal completed  (0.63s, 677 chars)   ← §2 search
tool terminal completed  (0.83s, 185 chars)   ← gmail modify --add-labels
tool write_file completed(0.04s, 235 chars)   ← state persisted before delivery
tool terminal completed  (0.07s, 58 chars)    ← lock release
```

Delivered response: `[SILENT]`, and the scheduler honoured it —
`2026-08-12 14:37:03,722 INFO cron.scheduler: Job '191332ae6a24': agent returned [SILENT] — skipping delivery`.
Correct: two newsletters is exactly the routine case §5 says to stay silent on.

State after (cursor advanced to the run's fixed end timestamp, no overflow):

```json
"cursor": "2026-08-12T16:36:26+02:00",
"seen": ["19ff456661cc3ad3"],
"last_run": {"at": "2026-08-12T16:36:26+02:00", "candidates": 1, "labelled": 1,
             "unrouted": 0, "overflow": 0, "by_label": {"📣 Newsletters & Pub": 1}},
"consecutive_failures": 0, "unrouted_senders": {}
```

**Mailbox ground truth** (read-only, 14:39–14:40 UTC). The §2 query now returns
`No messages found.` and both messages carry `Label_13`:

```
== 19ff456661cc3ad3  {'labels': ['IMPORTANT','CATEGORY_UPDATES','INBOX','Label_13']}
   subject: [No Action Needed]: Sonnet pricing update            (…@email.claude.com → row 7 📣)
== 19ff4c90892113cb  {'labels': ['CATEGORY_PROMOTIONS','IMPORTANT','INBOX','Label_13']}
   subject: L'avis de plus de 800 responsables IT…                (…@dynatrace.ai   → row 7 📣)
```

Both routings are correct per §3 row 7 (`no-reply@email.claude.com`,
`updates@dynatrace.ai` are both named senders). No `--remove-labels`, no system
label, no message left the inbox — `INBOX` still present on both.

Lock released (`~/.hermes/state/gmail-triage.lock` absent). Gateway still
`ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC`, `NRestarts=0`.

## 6. Iterations

**0.** The skill was never edited; repo md5 == live md5 == `8eaa93e…` throughout.
No `config.yaml` edit was needed or made.

## 7. Two discrepancies worth a future fix (neither blocked the run)

1. **State under-counted by one.** Two messages provably carry `Label_13` after
   run 2, but state records `labelled: 1`, `by_label: {📣: 1}` and only one id in
   `seen`. Harmless in operation — the second message now carries a user label,
   so `-has:userlabels` excludes it from every future run and it cannot be
   double-labelled — but the §5 report would under-report volume. Cause not
   isolated: the search's 677-char output looks like one record while the
   185-char `modify` output looks like two chained calls.
2. **`gmail search` does not always return JSON.** On an empty result set it
   prints the plain string `No messages found.` (19 bytes), not `[]`. SKILL.md §2
   says "`search` returns a JSON array"; a stricter future implementation that
   pipes it into a JSON parser would crash on the common case. The agent handled
   it fine in runs 1 and 3, so this is a doc correction, not a live bug.

## 8. Spec acceptance status (`docs/specs/wf7-gmail-triage.md`)

1. `hermes cron list` shows `[active]` + Paris `next_run_at` — **PASS**.
2. One `hermes cron run`: state file with the 8-entry map + cursor, ≥1 message
   labelled in Gmail, delivery is `[SILENT]` or ≤6 bullets — **PASS** (2
   messages labelled; run 1 delivered 1 bullet, run 2 `[SILENT]`).
3. The next tick labels only new mail, no re-labelling, no duplicate DM —
   **PASS**, §9.
4. Spot-check 10 labelled messages by hand — **partial**: all **4** messages
   labelled so far were checked against §3 and all 4 are right (§5, §9). Carry
   the full 10-message check into the first full day.

## 9. Scheduled tick — 17:00 Paris, unassisted (acceptance #3)

Not a manual trigger: the job's own schedule fired it.
`2026-08-12 15:00:28,827 INFO cron.scheduler: Running job 'Gmail triage' (ID: 191332ae6a24)`
— 15:00:28 **UTC** = 17:00 Paris, which is the third independent confirmation
that `0 7-21 * * *` is read in Europe/Paris.

Output `2026-08-12_17-01-27.md`; state after:

```json
"cursor": "2026-08-12T17:00:51+02:00",
"seen": ["19ff456661cc3ad3", "19ff67c620facaea", "19ff67d0471ff43e"],
"last_run": {"at": "2026-08-12T17:00:51+02:00", "candidates": 2, "labelled": 2,
             "unrouted": 0, "overflow": 0,
             "by_label": {"💳 Finance & Facturation": 1, "🤝 Partenaires & Prestataires": 1}},
"consecutive_failures": 0, "unrouted_senders": {}
```

`2026-08-12 15:01:27,969 INFO cron.scheduler: Job '191332ae6a24': agent returned [SILENT] — skipping delivery`
— no duplicate DM, and the two messages run 2 had labelled were **not** touched
again (`seen` grew by exactly the two new ids; `by_label` holds no 📣).

Mailbox ground truth, read-only:

```
19ff67c620facaea | ['UNREAD','IMPORTANT','Label_9','CATEGORY_PERSONAL','INBOX'] | from …@rebura…
19ff67d0471ff43e | ['UNREAD','IMPORTANT','Label_7','CATEGORY_UPDATES','INBOX']  | from …@mail.anthropic.com
in:inbox -has:userlabels after:2026/08/12 → No messages found.
```

`Label_7` on the Anthropic statement is §3 row 2, deterministic. `Label_9` on
the Rebura sender matches no row — it is the **judgment path**, and it landed on
🤝 Partenaires & Prestataires, which is right for an outside consultancy. Both
kept `UNREAD` and `INBOX`: nothing was read, archived or removed.

Accounting was exact on this tick (2 candidates → 2 labelled → 2 new `seen`
ids), which makes the §7.1 off-by-one a one-run anomaly rather than a
systematic miscount.
