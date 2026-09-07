# attach_to_session root-cause investigation (read-only)

Investigation date: 2026-09-07. Access: `ssh gaetan@192.168.0.9`, binary
`~/.local/bin/hermes` (not on PATH over non-interactive ssh). Read-only
throughout — no edits made, no secret/token values printed.

## Section 1 — CLI surface: does `attach_to_session` exist as a flag?

### `hermes cron --help`

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron --help'
usage: hermes cron [-h] [--accept-hooks]
                   {list,create,add,edit,pause,resume,run,remove,rm,delete,status,runs,history,incidents,notepad,doctor,tick}
                   ...

Manage scheduled tasks

positional arguments:
  {list,create,add,edit,pause,resume,run,remove,rm,delete,status,runs,history,incidents,notepad,doctor,tick}
    list                List scheduled jobs
    create (add)        Create a scheduled job
    edit                Edit an existing scheduled job
    pause               Pause a scheduled job
    resume              Resume a paused job
    run                 Run a job on the next scheduler tick
    remove (rm, delete)
                        Remove a scheduled job
    status              Check if cron scheduler is running
    runs (history)      Show durable execution attempts
    incidents           List or acknowledge durable cron failure incidents
    notepad             Read/write a job's durable notepad (persistent KV
                          across runs)
    doctor              Check scheduled jobs for common health issues
    tick                Run due jobs once and exit

options:
  -h, --help            show this help message and exit
  --accept-hooks        Auto-approve unseen shell hooks without a TTY prompt
                        (equivalent to HERMES_ACCEPT_HOOKS=1 /
                        hooks_auto_accept: true).
```

No `attach_to_session`/`attach-to-session` flag or subcommand anywhere.

### `hermes cron create --help`

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron create --help'
usage: hermes cron create [-h] --deliver DELIVER
                           [--failure-deliver FAILURE_DELIVER]
                           [--repeat REPEAT] [--skill SKILL]
                           [--script SCRIPT] [--no-agent]
                           [--monitor-script MONITOR_SCRIPT]
                           [--monitor-url MONITOR_URL] [--workdir WORKDIR]
                           [--model MODEL] [--provider PROVIDER]
                           [--reasoning-effort REASONING_EFFORT]
                           [--continuity] --name NAME
                           schedule [prompt]

positional arguments:
  schedule              Cron expression, e.g. '0 9 * * 1-5'
  prompt                Prompt/task for the agent to run

options:
  -h, --help            show this help message and exit
  --deliver DELIVER     Delivery target: origin, local, telegram, discord,
                        signal, platform:chat_id, or bot-chat[:profile]
  --failure-deliver FAILURE_DELIVER
                        Delivery target used only on failure (defaults to
                        --deliver)
  --repeat REPEAT       Repeat count, or 'inf' for indefinite (default: 1)
  --skill SKILL         Attach a skill to this job (repeatable)
  --script SCRIPT       Run a shell script instead of the agent (implies
                        --no-agent)
  --no-agent            Do not invoke the agent; script/monitor only
  --monitor-script MONITOR_SCRIPT
                        Script to run before the agent to gate execution
  --monitor-url MONITOR_URL
                        URL to poll before the agent to gate execution
  --workdir WORKDIR     Working directory for script/monitor execution
  --model MODEL         Model override for this job (user-owned; the
                        agent's cronjob tool cannot set this)
  --provider PROVIDER   Provider override for this job (user-owned; the
                        agent's cronjob tool cannot set this)
  --reasoning-effort REASONING_EFFORT
                        Reasoning effort override for this job
  --continuity          Reuse the most recent session as context (continuity
                        thread)
  --name NAME           Human-readable name for this job
```

No `attach_to_session`/`attach-to-session` flag present.

### `hermes cron edit --help`

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron edit --help'
usage: hermes cron edit [-h] [--schedule SCHEDULE] [--prompt PROMPT]
                         [--name NAME] [--deliver DELIVER]
                         [--failure-deliver FAILURE_DELIVER]
                         [--repeat REPEAT] [--skill SKILL]
                         [--add-skill ADD_SKILL] [--remove-skill REMOVE_SKILL]
                         [--clear-skills] [--script SCRIPT] [--no-agent]
                         [--agent] [--continuity] [--no-continuity]
                         [--monitor-script MONITOR_SCRIPT]
                         [--monitor-url MONITOR_URL] [--workdir WORKDIR]
                         [--model MODEL] [--provider PROVIDER]
                         [--reasoning-effort REASONING_EFFORT]
                         job_id

positional arguments:
  job_id                Job ID to edit

options:
  -h, --help            show this help message and exit
  --schedule SCHEDULE   New cron expression
  --prompt PROMPT       New prompt/task
  --name NAME           New human-readable name
  --deliver DELIVER     New delivery target
  --failure-deliver FAILURE_DELIVER
                        New failure-only delivery target
  --repeat REPEAT       New repeat count, or 'inf'
  --skill SKILL         Replace skills with this single skill
  --add-skill ADD_SKILL
                        Attach a skill to this job (repeatable)
  --remove-skill REMOVE_SKILL
                        Detach a skill from this job (repeatable)
  --clear-skills        Remove all skills from this job
  --script SCRIPT       Run a shell script instead of the agent
  --no-agent            Do not invoke the agent; script/monitor only
  --agent               Re-enable agent invocation (undo --no-agent)
  --continuity          Reuse the most recent session as context
  --no-continuity       Disable continuity for this job
  --monitor-script MONITOR_SCRIPT
                        Script to run before the agent to gate execution
  --monitor-url MONITOR_URL
                        URL to poll before the agent to gate execution
  --workdir WORKDIR     Working directory for script/monitor execution
  --model MODEL         Model override for this job (user-owned; the
                        agent's cronjob tool cannot set this)
  --provider PROVIDER   Provider override for this job (user-owned; the
                        agent's cronjob tool cannot set this)
  --reasoning-effort REASONING_EFFORT
                        Reasoning effort override for this job
```

No `attach_to_session`/`attach-to-session` flag present.

**Conclusion — Section 1:** `attach_to_session` is **not a documented,
user-settable CLI flag** anywhere in `hermes cron`, `cron create`, or
`cron edit`. There is also **no thread/session-delivery flag** beyond
`--deliver <platform:chat_id>` and `--continuity`/`--no-continuity` (which
reuses/disables the most-recent-session context, a different concept from
`attach_to_session`). There is **no unfurl/link-unfurling flag** anywhere in
this help text. Both `--model` and `--provider` help text carry the identical
note "(user-owned; the agent's cronjob tool cannot set this)" — hinting at an
internal agent-facing "cronjob tool" distinct from this CLI, a plausible
(unconfirmed) mechanism for how `attach_to_session` gets set outside the
documented flag surface.

There is also no per-job detail viewer: `hermes cron list --help` only takes
`[--all]` (no `--json`/`-o json`); `hermes cron status --help` and
`hermes cron doctor --help` take no job_id and are not per-job viewers. The
only way to inspect `attach_to_session` is direct, field-scoped inspection of
`~/.hermes/cron/jobs.json`.

## Section 2 — Daily-work-brief job `e231e5faf180`: live state

`hermes cron list` (default, human-readable):

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron list'
e231e5faf180 [active]  Gaetan daily work brief
  Schedule    30 8 * * 1-5
  Repeat      ∞
  Next run    2026-09-08T08:30:00+02:00
  Deliver     slack:D0BBYNM01BL
  Skills      daily-work-brief
  Last run    2026-09-07T09:30:28.385558+02:00 error: Interrupted by shutdown before terminal completion.
```

(No `--json`/`show`/`get` subcommand exists — see Section 1 — so the
authoritative per-field detail below comes from a field-scoped `jq` read of
the store, sensitive fields `prompt`/`script`/`monitor_script`/
`provider_snapshot`/`model_snapshot` excluded from the query.)

```
$ ssh gaetan@192.168.0.9 "jq -r '.jobs[] | select(.id==\"e231e5faf180\") | {id, deliver, attach_to_session, origin, context_from, state}' ~/.hermes/cron/jobs.json"
{
  "id": "e231e5faf180",
  "deliver": "slack:D0BBYNM01BL",
  "attach_to_session": false,
  "origin": {
    "platform": "slack",
    "chat_id": "D0BBYNM01BL",
    "chat_name": "D0BBYNM01BL",
    "thread_id": "1786191017.826669",
    "user_id": "U08BDJAMSRZ"
  },
  "context_from": null,
  "state": "scheduled"
}
```

**`attach_to_session` for the daily-work-brief job is explicitly `false`
(present in the store, not absent).** Deliver target is `slack:D0BBYNM01BL`
(a channel/DM-scoped target, not bare `slack`). `origin.thread_id` is
`1786191017.826669`. No separate unfurl-related field exists among this job's
fields — the full key list for a job object (enumerated via `jq keys[]`, no
values exposed) is: `attach_to_session, base_url, context_from, created_at,
deliver, enabled, enabled_toolsets, failure_streak, fire_claim, id,
last_delivery_error, last_dispatch, last_error, last_run_at, last_status,
model, model_snapshot, monitor_script, monitor_state, monitor_url, name,
next_run_at, no_agent, origin, paused_at, paused_reason, prompt, provider,
provider_snapshot, repeat, schedule, schedule_display, script, skill, skills,
state, workdir` — none of these names relate to link unfurling.

## Section 3 — Engagement-checker jobs `62e8cd9db637` / `759e08c598e3`

No `--json` mode exists on `hermes cron list` (see Section 1), so these were
located via `hermes cron list --all` (both are `[paused]`, hidden by the
default listing) and then detailed via the same field-scoped `jq` read used
above.

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes cron list --all'
...
62e8cd9db637 [paused]  Gaetan engagement checker
  Schedule    */30 10-16 * * 1-5
  Deliver     slack
  Last run    2026-08-19T16:34:52.329409+02:00 ok
...
759e08c598e3 [paused]  Gaetan engagement checker final pass
  Schedule    0 17 * * 1-5
  Deliver     slack
  Last run    2026-08-19T17:04:39.671760+02:00 ok
...
(other unrelated job IDs omitted)
```

```
$ ssh gaetan@192.168.0.9 "jq -r '.jobs[] | select(.id==\"62e8cd9db637\" or .id==\"759e08c598e3\") | {id, deliver, attach_to_session, origin, context_from, state}' ~/.hermes/cron/jobs.json"
{
  "id": "62e8cd9db637",
  "deliver": "slack",
  "attach_to_session": null,
  "origin": {
    "platform": "slack",
    "chat_id": "D0BBYNM01BL",
    "chat_name": "D0BBYNM01BL",
    "thread_id": "1786191017.826669",
    "user_id": "U08BDJAMSRZ"
  },
  "context_from": null,
  "state": "paused"
}
{
  "id": "759e08c598e3",
  "deliver": "slack",
  "attach_to_session": null,
  "origin": {
    "platform": "slack",
    "chat_id": "D0BBYNM01BL",
    "chat_name": "D0BBYNM01BL",
    "thread_id": "1786191017.826669",
    "user_id": "U08BDJAMSRZ"
  },
  "context_from": null,
  "state": "paused"
}
```

Plain `.attach_to_session` field access returns `null` for both — but `jq`
returns `null` both for an explicit `null` value AND for a genuinely absent
key, so this alone is ambiguous. Confirmed with `has()`:

```
$ ssh gaetan@192.168.0.9 'jq "[.jobs[] | select(.id==\"62e8cd9db637\" or .id==\"759e08c598e3\") | {id, has_attach: has(\"attach_to_session\")}]" ~/.hermes/cron/jobs.json'
[
  {
    "id": "62e8cd9db637",
    "has_attach": false
  },
  {
    "id": "759e08c598e3",
    "has_attach": false
  }
]
```

**Both engagement-checker jobs genuinely LACK the `attach_to_session` key
entirely** — it is not stored as `false` or explicit `null`, it is simply
absent from their job objects. Both share the same `origin.thread_id`
(`1786191017.826669`) and `origin.chat_id` (`D0BBYNM01BL`) as the daily job,
but use a bare `deliver: "slack"` (no channel/DM id) rather than the
channel-scoped `slack:D0BBYNM01BL` used by jobs that do carry the key. This
deliver-format difference (bare `slack` vs. `slack:<chat_id>`) is a strong
candidate correlate for whether `attach_to_session` gets populated in the
store at job-creation time.

## Section 4 — Mandated grep of the jobs store

```
$ ssh gaetan@192.168.0.9 "grep -o '\"attach_to_session\"[^,]*' ~/.hermes/cron/jobs.json | sort | uniq -c"
      2 "attach_to_session": false
```

Reconciling `jq` query (structural fields only, no sensitive values) showing
which 2 jobs these are:

```
$ ssh gaetan@192.168.0.9 "jq -r '.jobs[] | select(has(\"attach_to_session\")) | {id, name, deliver, attach_to_session, origin}' ~/.hermes/cron/jobs.json"
{
  "id": "e231e5faf180",
  "name": "Gaetan daily work brief",
  "deliver": "slack:D0BBYNM01BL",
  "attach_to_session": false,
  "origin": {
    "platform": "slack",
    "chat_id": "D0BBYNM01BL",
    "chat_name": "D0BBYNM01BL",
    "thread_id": "1786191017.826669",
    "user_id": "U08BDJAMSRZ"
  }
}
{
  "id": "e3c7a767c872",
  "name": "GCN-99 monitored completion and verification",
  "deliver": "slack:D0BBYNM01BL",
  "attach_to_session": false,
  "origin": {
    "platform": "slack",
    "chat_id": "D0BBYNM01BL",
    "chat_name": "Gaëtan",
    "thread_id": "1788589737.362879",
    "user_id": "U08BDJAMSRZ",
    "scope_id": "T7V1UGJ82"
  }
}
```

**Across the entire jobs store (~20+ jobs total, confirmed via
`hermes cron list --all`), only 2 jobs carry the `attach_to_session` key at
all — `e231e5faf180` (Gaetan daily work brief) and `e3c7a767c872` (GCN-99
monitored completion and verification, not originally named in this
investigation's scope) — and both are set to `false`.** This matches the
mandated grep's count exactly (2 occurrences, both `false`). Every other job
in the store, including the two engagement-checker jobs, lacks the key
entirely.

## Overall root cause

`attach_to_session` cannot be set or inspected through any documented
`hermes cron` CLI flag or subcommand in this Hermes version. It exists only
as an internal field in `~/.hermes/cron/jobs.json`, present (and `false`) on
exactly 2 of ~20+ jobs — both of which use a channel/DM-scoped `--deliver
platform:chat_id` target — and absent on every job using a bare-platform
`--deliver` target (including the two engagement-checker jobs). No
link-unfurling setting exists anywhere in the cron job schema, CLI-exposed or
otherwise. The `--model`/`--provider` help text's identical footnote
("user-owned; the agent's cronjob tool cannot set this") suggests an internal
agent-facing job-creation path distinct from this human CLI that may be
responsible for populating `attach_to_session`, but this is not directly
confirmed by any evidence gathered here.
