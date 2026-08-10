# P6 apply evidence — mc-metarepo KB refresh (SOUL + script + cron)

All times UTC. VM: `ssh gaetan@192.168.0.9`.

## Part 1 — landed by the interrupted 2026-08-10 ~09:58-09:59 UTC run

Verified from that run's transcript (not re-verified byte-for-byte here except
where noted in Part 2 §Sanity):

- `SOUL.md` rule 9 appended (mc-metarepo KB pointer); backup at
  `~/.hermes/SOUL.md.bak-p6-20260810`; diff = only the addition.
- `~/.hermes/scripts/kb-refresh.sh` created, 838 bytes, executable,
  `bash -n` OK, md5 `03610d5e925229b264b6bf37277d7718`.
- Manual refresh ran once, rc=0: clone `4831253` (2026-08-07) ->
  `2d3d014` (2026-08-10 09:26 +0200), branch `main`, clean.

## Part 2 — this run: cron job creation (2026-08-10, ~10:44-10:48 UTC)

### Sanity check

```
$ ssh gaetan@192.168.0.9 'ls -la ~/.hermes/scripts/kb-refresh.sh && md5sum ~/.hermes/scripts/kb-refresh.sh'
-rwxrwxr-x 1 gaetan gaetan 838 Aug 10 09:59 /home/gaetan/.hermes/scripts/kb-refresh.sh
03610d5e925229b264b6bf37277d7718  /home/gaetan/.hermes/scripts/kb-refresh.sh
```

Matches prescribed checksum. Proceeded.

### `hermes cron create --help` (read before scripting, per hard rule)

Relevant flags confirmed: `schedule` positional, `--name`, `--deliver`
(`origin, local, telegram, discord, signal, or platform:chat_id`), `--script`
(`Path to a script under ~/.hermes/scripts/`), `--no-agent` (script IS the
job, stdout delivered verbatim). No contradiction with the prescribed
invocation — used as-is, no adaptation needed.

### Cron job creation

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron create "every 1h" --no-agent --script kb-refresh.sh --deliver slack --name "mc-metarepo-refresh"'
Created job: 3eb53a322510
  Name: mc-metarepo-refresh
  Schedule: every 60m
  Script: kb-refresh.sh
  Mode: no-agent (script stdout delivered directly)
  Next run: 2026-08-10T13:47:29.368623+02:00
```

Bare basename `kb-refresh.sh` and bare `--deliver slack` (not
`slack:<chat_id>`) were both accepted first try — no retry with the full
path or `platform:chat_id` form was needed.

### Verification — `hermes cron list`

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron list'
┌─────────────────────────────────────────────────────────────────────────┐
│                         Scheduled Jobs                                  │
└─────────────────────────────────────────────────────────────────────────┘
  e231e5faf180 [active]
    Name:      Gaetan daily work brief
    Schedule:  30 8 * * 1-5
    Repeat:    ∞
    Next run:  2026-08-11T08:30:00+02:00
    Deliver:   slack:D0BBYNM01BL
    Skills:    daily-work-brief
    Last run:  2026-08-10T08:35:24.328233+02:00  ok
    Execution: completed  798ba336412c475e9dfc0971933f6748
  62e8cd9db637 [active]
    Name:      Gaetan engagement checker
    Schedule:  */30 10-16 * * 1-5
    Repeat:    ∞
    Next run:  2026-08-10T13:00:00+02:00
    Deliver:   slack:D0BBYNM01BL
    Skills:    engagement-checker
    Last run:  2026-08-10T12:35:29.387516+02:00  ok
    Execution: completed  be47530fae6c4ce981d48a760f418109
  759e08c598e3 [active]
    Name:      Gaetan engagement checker final pass
    Schedule:  0 17 * * 1-5
    Repeat:    ∞
    Next run:  2026-08-10T17:00:00+02:00
    Deliver:   slack:D0BBYNM01BL
    Skills:    engagement-checker
  137e6c516ddf [active]
    Name:      Forward new Claude invoices to Olivier
    Schedule:  30 9 * * *
    Repeat:    ∞
    Next run:  2026-08-11T09:30:00+02:00
    Deliver:   local
    Script:    claude_invoice_forward.py
    Mode:      no-agent (script stdout delivered directly)
    Last run:  2026-08-10T10:45:57.761032+02:00  ok
    Execution: completed  2c5204792ba14147a21b75318e2239ad
  3eb53a322510 [active]
    Name:      mc-metarepo-refresh
    Schedule:  every 60m
    Repeat:    ∞
    Next run:  2026-08-10T13:47:29.368623+02:00
    Deliver:   slack
    Script:    kb-refresh.sh
    Mode:      no-agent (script stdout delivered directly)
```

New job `3eb53a322510` (`mc-metarepo-refresh`) is listed, active, `every 60m`,
no-agent script mode, matching the prescribed invocation.

## Result

- New job listed: yes — `3eb53a322510`, `mc-metarepo-refresh`.
- Pre-existing jobs intact: **4** jobs (`e231e5faf180`, `62e8cd9db637`,
  `759e08c598e3`, `137e6c516ddf`) all present, unmodified, same
  schedule/deliver/skills/script fields as before creation. Note: the task
  brief said "3 pre-existing" jobs; the live count on the VM was 4 both
  before and after this run — none were touched, so the discrepancy is
  informational only, not a fault in this run.
- No secret value read, echoed, or written anywhere in this evidence file or
  in any command output above.
