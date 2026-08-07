# WF5 test — Orca delegation lane, live Slack E2E

**Verdict: PASS.** One Slack DM from Gaetan → Tars loaded the playbook skill
unprompted → ssh cooper → `delegate.sh` → coding agent ran the inspection →
real output back in-thread with the exact command and the
`[delegate.sh exit=… log=…]` line.

Spec: `docs/specs/wf5-orca-delegation.md`. Implementation evidence:
`status/probes/wf5/orca-implement.md`. Date 2026-08-07, window 20:28Z → 20:30Z.
VM `tars` @ 192.168.0.9, cooper local. Total elapsed inbound→reply: **43.4s**.

**Messages sent: exactly one**, as Gaetan (U08BDJAMSRZ) to the home DM
D0BBYNM01BL. No other human messaged. No `config.yaml` edit (so no `.bak`, no
flock, no restart — none needed). No sops. No git. No `status/lane-*.md` edit.
192.168.0.3 never touched. Gateway `active`, `NRestarts=0` before and after.
The known Codex "incomplete after 3 continuation attempts" bug **did not
occur**; no retry was needed.

## The turn

| | |
|---|---|
| Prompt ts | `1786134543.728879` (2026-08-07 20:29:03Z), user U08BDJAMSRZ |
| Prompt permalink | https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786134543728879?thread_ts=1786134543.728879&cid=D0BBYNM01BL |
| Reply ts | `1786134589.874459` (20:29:49Z), user U0BBH85NAKH (Tars), in-thread |
| Reply permalink | https://mobileclub-squad.slack.com/archives/D0BBYNM01BL/p1786134589874459?thread_ts=1786134543.728879&cid=D0BBYNM01BL |
| Progress message | ts `1786134550.972269`, edited in place (kanban_show → skill read → terminal) |
| Hermes session | `20260807_202906_947ad4e5`, model gpt-5.6-sol via openai-codex |

Prompt sent (verbatim):

> Tars: per the delegate-to-cooper playbook, delegate a trivial read-only
> inspection to cooper — in ~/orca/workspaces/tars-delegated/, list the
> directory and report the line count of delegate.sh. Reply with the output
> plus the exact command you ran.

## Evidence 1 — the Slack reply carries the real result

Fetched via `conversations.replies` on the prompt ts (never bare history).
Reply text, unedited:

```
Cooper completed the read-only inspection. `delegate.sh` has *24 lines*.

*Output*
total 28
drwxrwxr-x 4 gaetan gaetan 4096 Aug  7 20:26 .
drwxrwxr-x 5 gaetan gaetan 4096 Aug  7 20:17 ..
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:24 .claude
-rwxrwxr-x 1 gaetan gaetan 1366 Aug  7 20:25 delegate.sh
-rw-rw-r-- 1 gaetan gaetan  128 Aug  7 20:22 first_10_primes.py
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:29 logs
-rw-rw-r-- 1 gaetan gaetan  458 Aug  7 20:26 NOTE.md

24 /home/gaetan/orca/workspaces/tars-delegated/delegate.sh

*Exact command run*
ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF'
Perform a trivial read-only inspection confined to ~/orca/workspaces/tars-delegated/.
1. List the directory contents.
2. Report the line count of ~/orca/workspaces/tars-delegated/delegate.sh.
Do not modify any files. Return the exact inspection output and state the commands you used.
BRIEF

[delegate.sh exit=0 log=/home/gaetan/orca/workspaces/tars-delegated/logs/20260807T202922Z.log]
```

Ground truth on cooper, checked independently after the fact:
`wc -l delegate.sh` → `24`; the directory listing matches file-for-file,
byte-for-byte. The reply is not a plausible hallucination — it is the output.

## Evidence 2 — VM agent.log shows the terminal/tool execution

`~/.hermes/logs/agent.log`, session `20260807_202906_947ad4e5`:

```
20:29:06,290 [947ad4e5] agent.turn_context: conversation turn: session=20260807_202906_947ad4e5 model=gpt-5.6-sol provider=openai-codex platform=slack history=0 msg='Tars: per the delegate-to-cooper playbook, delegate a trivial read-only inspecti...'
20:29:10,460 [947ad4e5] API call #1: in=29319 out=53 latency=4.1s
20:29:10,470 [947ad4e5] Tool kanban_show returned error (0.00s): {"error": "task_id is required (or set HERMES_KANBAN_TASK in the env)"}
20:29:16,362 [947ad4e5] tool skill_view completed (0.02s, 4456 chars)     <-- read delegate-to-cooper SKILL.md
20:29:40,129 [947ad4e5] tool terminal completed (18.28s, 850 chars)       <-- the ssh cooper delegate.sh call
20:29:49,335 [947ad4e5] API call #4: in=31252 out=390 latency=9.2s
20:29:49,350 [947ad4e5] Turn ended: reason=text_response(finish_reason=stop) api_calls=4 tool_turns=3 response_len=1041
20:29:49,628 gateway.run: response ready: platform=slack chat=D0BBYNM01BL time=43.4s api_calls=4 response=1041 chars
20:29:49,652 gateway.platforms.base: [Slack] Sending response (1041 chars) to D0BBYNM01BL
```

Exactly **one** `terminal` call, 18.28s — one delegation, no flailing. The
`skill_view` at 20:29:16 is the model pulling the playbook before acting; the
Slack progress message rendered it as ":books: Reading skill delegate-to-cooper".

`kanban_show` erroring is a sibling lane's toolset (missing `task_id`), unrelated
to WF5 and harmless — the model moved on and delegated.

## Evidence 3 — cooper-side proof the command actually ran

New log file, created by `delegate.sh` during the turn (baseline before the
probe was 5 logs, newest `20260807T202541Z.log`):

```
$ ls -la --time-style=+%Y%m%dT%H%M%SZ /home/gaetan/orca/workspaces/tars-delegated/logs/
-rw-rw-r-- 1 gaetan gaetan  695 20260807T202939Z 20260807T202922Z.log
```

Filename stamp 20:29:22Z, mtime 20:29:39Z — brackets the VM's
`tool terminal completed` at 20:29:40, and its path is the one quoted verbatim
in the Slack reply. Contents (`cat 20260807T202922Z.log`):

```
Done — read-only, nothing modified.

**1. Directory listing** — `ls -la /home/gaetan/orca/workspaces/tars-delegated/`:
total 28
drwxrwxr-x 4 gaetan gaetan 4096 Aug  7 20:26 .
drwxrwxr-x 5 gaetan gaetan 4096 Aug  7 20:17 ..
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:24 .claude
-rwxrwxr-x 1 gaetan gaetan 1366 Aug  7 20:25 delegate.sh
-rw-rw-r-- 1 gaetan gaetan  128 Aug  7 20:22 first_10_primes.py
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:29 logs
-rw-rw-r-- 1 gaetan gaetan  458 Aug  7 20:26 NOTE.md

**2. Line count** — `wc -l /home/gaetan/orca/workspaces/tars-delegated/delegate.sh`:
24 /home/gaetan/orca/workspaces/tars-delegated/delegate.sh

delegate.sh is 24 lines.
```

The workspace is otherwise unchanged — no new files, nothing modified. The task
was read-only and stayed read-only.

## Evidence 4 — the guardrails doc reached the model

Not inferred from the reply's tone; four independent signals:

| Playbook clause | Observed |
|---|---|
| Skill is model-invocable, no prompt needed | `skill_view` fired on API call #2, before any terminal use — Gaetan named the playbook but never told it to read a file |
| Only the four whitelisted command shapes | One `terminal` call, shape (1): `ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF' … BRIEF`. Never typed `claude`, never passed agent flags |
| Tars writes the brief, not the work | The brief is Tars's own prose, restating scope ("confined to ~/orca/workspaces/tars-delegated/", "Do not modify any files") that Gaetan did not dictate |
| Reporting duty: verdict, then exact command verbatim incl. brief, then the exit/log line | All three present, in that order (§1) |

The scope sentences Tars added to its own brief are the clearest tell: that
constraint exists in the SKILL.md, not in Gaetan's message.

## Result matrix

| Criterion | Result |
|---|---|
| Reply contains the real result (matches cooper ground truth) | PASS (§1, §3) |
| VM agent.log shows terminal/tool execution | PASS (§2) |
| Cooper-side evidence the command actually ran | PASS (§3, new log 20260807T202922Z.log) |
| Guardrails doc reached the model / playbook behaviour reflected | PASS (§4) |
| Task completed within 5 min | PASS — 43.4s |
| Codex multi-step bug | Did not occur; no retry needed |
| Exactly one Slack message sent, no other human contacted | PASS |

## Notes and residual ceilings

- The progress message (ts `…550.972269`) is edited in place; polling
  `conversations.replies` and breaking on "first new message" catches the
  progress stub, not the answer. Break on the `[delegate.sh exit=` marker.
- Ceilings from the spec are untested here and remain untested: the 240s
  timeout path, concurrent delegations sharing the one working directory, and
  repo access (still absent by design). This probe exercised the happy path of
  a read-only brief; §5 of `orca-implement.md` remains the guardrail matrix of
  record.
