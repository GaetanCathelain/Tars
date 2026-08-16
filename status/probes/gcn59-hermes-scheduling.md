# GCN-59 — Does Hermes v0.20.0 have built-in scheduling? (read-only probe)

**Verdict: YES — Hermes has a native cron scheduler (`hermes cron`), already
live on the Tars VM with 11 active jobs. No external systemd timer is needed
for periodic Linear polling; a `hermes cron create` job does it natively.**

## Evidence

### `hermes cron` is a first-class subcommand

```
$ hermes cron --help
usage: hermes cron [-h] [--accept-hooks]
                   {list,create,add,edit,pause,resume,run,remove,rm,delete,status,runs,history,notepad,tick}
Manage scheduled tasks
```

Schedules accept cron syntax (`0 9 * * *`), interval syntax (`every 60m`,
`30m`), or one-shot (`once at 2026-09-11 18:10`). `create` supports
`--skill`, `--script` (+`--no-agent` for script-only, no-LLM runs),
`--monitor-script` / `--monitor-url` (hash-suppressed diff-triggered runs),
`--deliver <target>`, `--workdir`, `--model`/`--provider` pinning.

### The scheduler is live and firing on Tars right now

```
$ hermes cron status
✓ Gateway is running — cron jobs will fire automatically
  PID: 1807357
  Ticker heartbeat: 13s ago
  11 active job(s)
  Next run: 2026-08-16T10:36:12.787243+02:00
```

The ticker runs **inside the gateway process** (systemd `--user`
`hermes-gateway.service`), not as a separate OS-level timer. `cron tick`
exists as a manual "run due jobs once and exit" escape hatch, but normal
operation is gateway-embedded — confirmed by 11 jobs with real `Last run` /
`Next run` timestamps, e.g.:

- "Gaetan daily work brief" — `30 8 * * 1-5`, skill `daily-work-brief`, deliver slack
- "mc-metarepo-refresh" — `every 60m`, script-mode `--no-agent`
- "Gmail triage" — `0 7-21 * * *`, skill `gmail-triage`
- "GoCardless rotation" — `once at 2026-09-11 18:10`

This is a **directly transferable pattern for the GCN-59 ask**: a Linear-poll
job would look like
`hermes cron create "every 15m" --skill <linear-poller> --deliver slack --name "Linear ticket poll"`
— no external timer infra needed.

### config.yaml confirms cron is a config-driven subsystem

```
$ grep -inE 'schedul|cron|timer|interval|heartbeat|poll' ~/.hermes/config.yaml
125:  nudge_interval: 10
130:  creation_nudge_interval: 15
139:cron:
```

`cron:` section present (`wrap_response: false` — only key set beyond
default; per-job model/schedule config lives in `~/.hermes/cron/`, not
grepped further per read-only scope). No separate `schedul*`/`timer`/`poll`
keys elsewhere — cron is the only built-in scheduling primitive.

### Other wake paths that exist alongside cron (for completeness)

- `hermes webhook` — event-driven activation via dynamic webhook
  subscriptions (inbound HTTP trigger, not periodic).
- `hermes kanban` — task board; dispatch now runs inside the gateway too
  (`daemon` subcommand is DEPRECATED, folded into `gateway start`), but
  kanban dispatch is claim/event-driven, not time-based — not a poller by
  itself.
- Neither is a scheduler substitute; `cron` is the only periodic/timer
  primitive.

### External OS-level timer infra check

```
$ systemctl --user list-timers --all
NEXT                        LEFT LAST                         PASSED UNIT                           ACTIVATES
Sun 2026-08-16 16:27:51 UTC   8h Sat 2026-08-15 16:27:51 UTC 15h ago launchpadlib-cache-clean.timer launchpadlib-cache-clean.service
1 timers listed.
```

Only a stock Ubuntu `launchpadlib` cache-clean timer — **no
Hermes-related systemd `--user` timer exists**. All periodic Hermes work
(daily brief, engagement checks, kb-refresh, gmail triage, etc.) already runs
through the built-in `hermes cron` scheduler embedded in the gateway
process, not via any external timer unit.

## Answer to the original question

Hermes does **not** need an external `systemd --user` timer calling
`hermes chat` to poll Linear periodically. `hermes cron create` (schedule +
optional `--skill`) is the native, already-proven mechanism — 11 jobs are
running this way in production on Tars today, including ones that poll
external systems (Gmail triage every hour during the day, mc-metarepo
refresh every 60m). The only genuinely external-trigger-only path is
`hermes webhook` (inbound HTTP event), which is not periodic by nature.

## Surprise

None of the config-file scheduling knobs matter for this decision —
`hermes cron` is a proper subcommand with its own list/status/create/edit
UX, and it's already the load-bearing mechanism for 11 real recurring jobs
on this VM. A Linear-poll job is a one-liner away, no new infra.
