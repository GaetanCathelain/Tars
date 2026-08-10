# P6 verify — independent re-measurement of mc-metarepo KB wiring

Independent verification, not the deploying session. All times UTC.
VM: `ssh gaetan@192.168.0.9`. No sops decrypt, no config/env read, no cron
edits, no config.yaml/.env writes performed by this probe.

## 1. SOUL.md pointer

```
$ ssh gaetan@192.168.0.9 "grep -n -A3 -B1 'mc-metarepo' ~/.hermes/SOUL.md"
108-
109:9. My knowledge base on this machine is the git clone at `~/dev/mc-metarepo` —
110-   search it with `search_files`/`read_file` before delegating anything. Its
111-   submodules are empty, so product source code is **not** there.
112-
```

Pointer present, rule 9, as expected.

## 2. Cron list

```
$ ssh gaetan@192.168.0.9 "export XDG_RUNTIME_DIR=/run/user/\$(id -u); ~/.local/bin/hermes cron list"
```

5 jobs total: `mc-metarepo-refresh` (`3eb53a322510`, every 60m, no-agent,
script `kb-refresh.sh`, next run 2026-08-10T13:47:29+02:00, active) +
4 pre-existing jobs, all present/active/unmodified: `Gaetan daily work brief`
(`e231e5faf180`), `Gaetan engagement checker` (`62e8cd9db637`), `Gaetan
engagement checker final pass` (`759e08c598e3`), `Forward new Claude invoices
to Olivier` (`137e6c516ddf`).

Note: the check brief said "3 pre-existing jobs"; the deploying session's own
evidence (`status/probes/wf5/p6-apply.md`, §Result) already recorded the live
baseline as 4 both before and after its cron-create run — so this is a stale
number in the brief, not a fault. Re-confirmed here: nothing beyond the new
job changed.

## 3. Clone freshness / cleanliness

```
$ ssh gaetan@192.168.0.9 "git -C ~/dev/mc-metarepo log -1 --format='%h %ad %s' --date=iso"
2d3d014 2026-08-10 09:26:42 +0200 chore(helptaker): seal duty day 2026-08-10

$ ssh gaetan@192.168.0.9 "git -C ~/dev/mc-metarepo status -sb"
## main...origin/main
```

Fresh (2026-08-10, today), clean working tree, on `main` tracking
`origin/main`. PASS.

## 4. Live retrieval trials

### Trial 1 — Next Mobiles vs Mobile.Club order discriminant

UTC before: `2026-08-10T10:48:19Z` (host clock check) / query launched
`2026-08-10T10:48:4x`.

```
$ ssh gaetan@192.168.0.9 "export XDG_RUNTIME_DIR=/run/user/\$(id -u); time ~/.local/bin/hermes chat -Q -q 'In api-v2, how do I tell a Next Mobiles order apart from a Mobile.Club order?'"
session_id: 20260810_104846_b4d8a2
real	0m32.362s
```

Answer:
> Check `next_mobile_order_mappings` by `order_id`:
> - Mapping exists → **Next Mobiles**
> - No mapping → **Mobile.Club**
> This is what `IsNextMobilesOrderService` uses. Don't rely on
> `orders.business_entity_id`: NM order creation leaves it at the
> Mobile.Club default, so it's unreliable.

Matches ground truth (`knowledge/nm-vs-mc-order-source.md`) on both halves:
(a) mapping-table discriminant, correct table+column, correct service name;
(b) the `business_entity_id` trap, correctly framed as never-set-by-NM-path.
**Trial 1: PASS, correct.**

### Trial 2 — inline API key on shared VM, rotate?

UTC before: `2026-08-10T10:48:42Z`. Query launched `2026-08-10T10:49:35Z`,
ended `10:49:51Z` (`real 0m17.618s`).

```
$ ssh gaetan@192.168.0.9 "export XDG_RUNTIME_DIR=/run/user/\$(id -u); time ~/.local/bin/hermes chat -Q -q 'Last week I ran a one-off command on the shared agents VM with an API key inline in the command. Nobody was watching the terminal. Do I actually need to rotate the key?'"
session_id: 20260810_104935_3b3723
real	0m17.618s
```

Answer (full, via local rtk tee capture — terminal display truncates but
the tee file is unredacted):
> Yes—rotate it.
> "Nobody watched the terminal" only rules out shoulder-surfing. An inline
> API key may persist in: shell history, process/audit logs, monitoring or
> EDR telemetry, terminal/session recordings, `/proc` or process listings
> while the command ran.
> Treat it as potentially exposed, not definitely stolen. Revoke or rotate
> the old key, then verify the replacement works. Deleting the command from
> shell history is useful cleanup, but not a substitute for rotation.
> For future one-offs, pass secrets via a protected environment file, stdin,
> or the service's credential store—not as command-line arguments.

Verdict on content: reaches the right yes/rotate conclusion but is **generic
security reasoning**, not the specific grounded claim. It never names
`sudo`, `/var/log/auth.log`, or the systemd journal, and never says anything
about NOT erasing/truncating logs (the ground-truth file's explicit warning)
— the closest analogue ("deleting shell history is useful cleanup, not a
substitute") is about shell history, not log erasure, and isn't the same
claim. **Trial 2: content is plausible but not grounded in
`knowledge/sudo-argv-secrets-authlog.md`'s specific mechanism.**

## 5. Log evidence — mc-metarepo file access per trial

Both sessions' full `agent.log` lines for the tool-call window (session IDs
`20260810_104846_b4d8a2` for trial 1, `20260810_104935_3b3723` for trial 2):

```
$ ssh gaetan@192.168.0.9 "sed -n '40766,40778p' ~/.hermes/logs/agent.log"
2026-08-10 10:48:53,311 INFO [20260810_104846_b4d8a2] agent.conversation_loop: API call #1 ...
2026-08-10 10:48:53,325 WARNING [20260810_104846_b4d8a2] agent.tool_executor: Tool kanban_show returned error ...
2026-08-10 10:48:59,918 INFO [20260810_104846_b4d8a2] agent.conversation_loop: API call #2 ...
2026-08-10 10:49:00,008 INFO agent.tool_executor: tool skill_view failed (0.08s) ...
2026-08-10 10:49:00,042 INFO agent.tool_executor: tool search_files completed (0.11s, 8749 chars)
2026-08-10 10:49:09,024 INFO [20260810_104846_b4d8a2] agent.conversation_loop: API call #3 ...
2026-08-10 10:49:09,079 INFO [20260810_104846_b4d8a2] agent.tool_executor: tool read_file completed (0.05s, 958 chars)
2026-08-10 10:49:16,311 INFO [20260810_104846_b4d8a2] agent.conversation_loop: API call #4 ...
2026-08-10 10:49:16,329 INFO [20260810_104846_b4d8a2] agent.conversation_loop: Turn ended: ... tool_turns=3 ...
```

Trial 1: `search_files` (8749 chars result) then `read_file` (958 chars
result) both executed mid-turn, 4 API calls, `tool_turns=3`. Consistent with
a KB lookup landing on the right file. (Caveat: this log level records tool
name + result size, not the path argument, so the exact path
`~/dev/mc-metarepo/knowledge/nm-vs-mc-order-source.md` cannot be quoted
verbatim from `agent.log` — the file's own repo-relative grep below
corroborates the file's presence and content instead.)

```
$ ssh gaetan@192.168.0.9 "grep -n '20260810_104935_3b3723' ~/.hermes/logs/agent.log"
2026-08-10 10:49:36,778 INFO [...] agent.turn_context: conversation turn: session=20260810_104935_3b3723 ...
2026-08-10 10:49:44,944 INFO [...] agent.conversation_loop: API call #1 ...
2026-08-10 10:49:44,965 WARNING [...] agent.tool_executor: Tool kanban_show returned error (0.00s): {"error": "task_id is required..."}
2026-08-10 10:49:51,232 INFO [...] agent.conversation_loop: API call #2 ...
2026-08-10 10:49:51,247 INFO [...] agent.conversation_loop: Turn ended: reason=text_response ...
```

Trial 2: **only** tool call in the whole turn is `kanban_show`, which
errored (`task_id is required`). **No `search_files` or `read_file` call at
all.** Answer was produced from 2 API calls with zero KB access.
**Trial 2: FAIL — ungrounded, no logged mc-metarepo file access.**

Sanity: grepped the whole trial window (`sed -n '40760,40830p'`) for the
literal string `mc-metarepo` — zero matches in either session's log lines,
confirming the path never surfaces at this log level for either trial (a
tooling limitation, not signal either way) — the *tool-name* evidence above
is what decides grounding, and it splits 1-for-1 between trials.

## 6. errors.log — new entries in trial window

```
$ ssh gaetan@192.168.0.9 "grep -n '10:4[6-9]\|10:5[0-2]' ~/.hermes/logs/errors.log | grep '2026-08-10'"
21199-21202: WARNING tools.registry: check_fn check_{bfl,computer_use,image_generation}_requirements / check_web_api_key returned False; dependent tools ... (x2, once per session startup — routine, optional-tool capability probes, unrelated to this task)
21203: WARNING [20260810_104846_b4d8a2] agent.tool_executor: Tool kanban_show returned error (already covered above)
21204: WARNING [20260810_104846_b4d8a2] agent.tool_executor: Tool skill_view returned error (already covered above)
21209: WARNING [20260810_104935_3b3723] agent.tool_executor: Tool kanban_show returned error (already covered above)
```

No ERROR-level entries in the window; only routine startup WARNINGs and the
already-explained tool errors (`kanban_show` needs `HERMES_KANBAN_TASK`,
unrelated to KB; `skill_view` looked for a skill name that doesn't exist).
Nothing indicates a KB-wiring fault.

## Verdict: PARTIAL

- Pointer: PASS (rule 9 present, correct wording).
- Cron: PASS (`mc-metarepo-refresh` active; all 4 live pre-existing jobs
  intact — brief's "3" is a stale count already flagged as informational in
  the deploying session's own evidence, `p6-apply.md`).
- Clone freshness/cleanliness: PASS (`2d3d014`, 2026-08-10, clean, main).
- Trial 1: PASS — correct on both required halves, and grounded
  (`search_files` + `read_file` both fired mid-turn).
- Trial 2: FAIL — answer is a plausible generic security take (correct
  yes/rotate conclusion) but does not contain the ground-truth file's
  specific claims (sudo logs full argv to `/var/log/auth.log` AND the
  systemd journal; explicitly does not advise erasing logs), and **zero**
  `search_files`/`read_file` calls were logged for that turn — the agent
  never touched the KB for this query.

Per the brief's own bar ("BOTH trials ... AND at least one logged
mc-metarepo file access per trial" / "a plausible but ungrounded answer is a
FAIL for that trial"), one of two trials fails outright → overall **PARTIAL**,
not PASS. Infrastructure (pointer, cron, clone) is solid; retrieval is not
reliably triggered — trial 2's phrasing (a scenario/judgment question, no
literal "api key"/file-system keyword overlap with the KB doc title) didn't
prompt the agent to search the KB at all, whereas trial 1 (a direct "how do
I do X in api-v2" question) did.
