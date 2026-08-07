# WF5 Kanban lane — E2E test evidence

Date: 2026-08-07, Tars VM (192.168.0.9), all times UTC (VM clock).
Spec: `docs/specs/wf5-kanban.md`. Implement evidence: `status/probes/wf5/kanban-implement.md`.

Verdict: **PASS** — both legs green. One spec correction below (§5): cards do
**not** stay in `todo`; they auto-promote to `ready` and the live gateway
dispatcher spawns a worker on them within one tick.

No config edit was made by this test. `~/.hermes/config.yaml` mtime still
`Aug 7 20:21` (the implement-lane edit), md5 `2633c9d6db0bc6347a8d5bed4ad83bed`.
Gateway never restarted: `NRestarts=0`, `ActiveEnterTimestamp=Fri 2026-08-07
19:40:45 UTC`, still active at 20:30:37.

## 0. Baseline

```
$ hermes --version      → Hermes Agent v0.20.0 (2026.8.3)
$ grep -A3 '^toolsets:' ~/.hermes/config.yaml
147:toolsets:
148:- hermes-cli
149:- kanban
$ hermes kanban stats   → triage 0  todo 0  scheduled 0  ready 0  running 0  blocked 0  done 0
```

## 1. Leg 1 CLI — turn A creates a card

```
$ hermes chat -Q -q "Use the kanban_create tool to create exactly one card on the
    default board titled: WF5 kanban smoke card. Description: E2E probe, safe to
    delete. Then report the task id you got back. Do not create anything else."
session_id: 20260807_202702_316ced
t_4f772ab0
```

Real execution, not narration — `~/.hermes/logs/agent.log`:

```
20:27:08 WARNING [20260807_202702_316ced] agent.tool_executor: Tool kanban_show returned error (0.00s)...
20:27:14 INFO    [20260807_202702_316ced] agent.tool_executor: tool terminal completed (0.67s, 308 chars)
20:27:18 INFO    [20260807_202702_316ced] agent.tool_executor: tool kanban_create completed (0.01s, 150 c...
```

(`kanban_show` erroring first is the same benign no-task-id probe seen in the
implement lane.)

Board immediately after:

```
$ hermes kanban stats
  ready 1        By assignee: default ready=1        Oldest ready task age: 10s
```

## 2. Leg 1 CLI — turn B lists the board and sees it

```
$ hermes chat -Q -q "Use the kanban_list tool to list the default board and report
    every task you see as: id | title | status. Do not create, modify or delete anything."
session_id: 20260807_202739_4bd848
t_4f772ab0 | WF5 kanban smoke card | ready
```

```
20:27:49 WARNING [20260807_202739_4bd848] agent.tool_executor: Tool kanban_show returned error (0.00s)...
20:27:52 INFO    [20260807_202739_4bd848] agent.tool_executor: tool kanban_list completed (0.00s, 529 chars)
```

The reported id/title/status match the on-disk board exactly.

## 3. Unplanned but real: the dispatcher worked the card

Within one dispatcher tick the card went `ready → running → done` on its own,
worked by a spawned worker — first time `~/.hermes/kanban/logs/` has ever been
populated (implement lane recorded it absent).

```
$ hermes kanban list
● t_4f772ab0  running   default   WF5 kanban smoke card
$ ps -ef | grep kanban
gaetan 87426 76255 26 20:27 ? .../venv/bin/python ...
$ cat ~/.hermes/kanban/logs/t_4f772ab0.log
Query: work kanban task t_4f772ab0
Initializing agent...
  ┊ ⚡ kanban_sh   0.0s
  ┊ ⚡ kanban_co   0.0s
╭─ ⚕ Hermes ─╮
Kanban task t_4f772ab0 completed successfully. E2E lifecycle passed: claimed, read, and finalized.
╰─╯
Session: 20260807_202800_1a199e   Duration: 16s   Messages: 6 (1 user, 4 tool calls)
```

So the full lane — model tool → DB → gateway dispatcher → spawned worker →
terminal state — is proven working, not just the tool surface.

## 4. Leg 2 Slack — one DM as Gaetan, in-thread reply, board changed

One message only, sent as Gaetan (U08BDJAMSRZ) to the Tars home DM D0BBYNM01BL
with the VM's personal xoxc/xoxd creds (`curl -K` from stdin; no secret in argv,
none printed). `ok=True ts=1786134530.441039`.

`conversations.replies` on that ts (never bare history):

```
--- U08BDJAMSRZ 'kanban test: please use the kanban_create tool to add one card on the default
                 board titled "WF5 kanban slack smoke card" (description: E2E probe, safe to
                 delete), then reply with the task id it returned.'
--- U0BBH85NAKH ':clipboard: kanban_show...\n:computer: terminal\n```\nhermes profile list\n```\n:heavy_plus_sign: kanban_create...'
--- U0BBH85NAKH 't_f525e54b'
```

Board state change verified on the VM afterwards, independently of Slack:

```
$ hermes kanban show t_f525e54b
Task t_f525e54b: WF5 kanban slack smoke card
  status:    ready
  assignee:  default
  workspace: scratch
  created:   2026-08-07 20:29 by worker
Body:
E2E probe, safe to delete
Events (1):
  [2026-08-07 20:29] created {'assignee': 'default', 'status': 'ready', 'parents': [], ...}
```

That card too was picked up and worked by a spawned worker:

```
$ tail ~/.hermes/kanban/logs/t_f525e54b.log
Kanban task t_f525e54b completed successfully. WF5 Slack E2E smoke probe passed.
Session: 20260807_203000_8d6313   Duration: 12s   Messages: 6 (1 user, 4 tool calls)
```

## 5. Spec correction — `kanban_create` does not leave cards in `todo`

`docs/specs/wf5-kanban.md` ("What was deliberately NOT enabled") states new
tasks land in `todo` and that nothing runs unless created with `--triage`. The
`triage`/`auto_decompose` half is right; the `todo` half is not, and the
practical consequence is the opposite of "nothing happens":

- The tool schema does say a new task "lands in 'todo'" (`tools/kanban_tools.py:1958,2007`).
- But `kanban_db.py:4135 recompute_ready()` — *"Promote `todo` tasks to `ready`
  when all parents are `done` or `archived`"* — promotes any **parentless** task
  on the next pass. It runs on the dispatcher tick and also inside `kanban_list`
  (`kanban_tools.py:600`, schema note at `:1592`).
- The dispatcher then applies `default_assignee` to unassigned ready tasks
  (`gateway/kanban_watchers.py:1122-1129`) and spawns a worker.

Measured: card created 20:27:18 → `ready` by 20:27:28 (`Oldest ready task age:
10s`) → `running` by 20:27:52 → `done` by 20:28. Same for the Slack card.

Not a defect — this is the kanban lane doing its job — but the spec's "nothing
auto-decomposes" should not be read as "nothing auto-runs". **Any card created
without parents gets an autonomous worker within ~60s.** The brake is
`kanban.auto_decompose` only for triage fan-out; to stop dispatch itself the
knob is `kanban.dispatch_in_gateway: false`, currently at its default `true`.

## 6. Cleanup — cleanup turn, then archive, board empty

Cleanup chat turn (both cards had already self-completed by then):

```
$ hermes chat -Q -q "Cleanup: use kanban_list to find every task whose title starts
    with WF5 kanban, then use kanban_complete on each one that is not already done.
    Report each id and its final status. Do not create anything."
session_id: 20260807_203002_80fc79
- t_4f772ab0 — done
- t_f525e54b — done
Both were already done; no tasks were modified or created.

20:30:11 WARNING [20260807_203002_80fc79] agent.tool_executor: Tool kanban_show returned error (0.00s)...
20:30:14 INFO    [20260807_203002_80fc79] agent.tool_executor: tool kanban_list completed (0.00s, 1088 ch...
```

There is no model-facing delete/archive tool — the 12 exposed `kanban_*` tools
stop at `kanban_complete`; `archive` is CLI-only. Final removal:

```
$ hermes kanban archive t_4f772ab0 t_f525e54b
Archived t_4f772ab0
Archived t_f525e54b
$ hermes kanban list
(no matching tasks)
$ hermes kanban stats
  triage 0  todo 0  scheduled 0  ready 0  running 0  blocked 0  done 0
```

Board empty, exactly as at baseline.

## 7. Residue

- Two archived rows remain in `~/.hermes/kanban.db` (archive is soft-delete; they
  are invisible to `list`/`stats`).
- `~/.hermes/kanban/logs/t_4f772ab0.log`, `t_f525e54b.log` (974 B / 956 B) kept
  deliberately as worker-spawn evidence; delete if a clean-slate board is wanted.
- Worker sessions `20260807_202800_1a199e` and `20260807_203000_8d6313` exist in
  session history; both ran in `workspace: scratch`.
- VM scratch files `/tmp/wf5_*` removed (`ls: cannot access '/tmp/wf5_*'`); they
  held no secrets — tokens were sourced from `~/.hermes/.env` into env and piped
  to `curl -K` on stdin.
- Slack: exactly one human-visible message sent, to D0BBYNM01BL only.
