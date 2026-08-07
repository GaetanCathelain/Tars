# Probe — hourly refresh mechanisms (P6/kb)

All commands run on `gaetan@192.168.0.9` (tars). Read-only against
`~/.hermes/` and `~/dev/mc-metarepo` throughout; all failure-mode tests ran in
throwaway sandboxes under `/tmp/kb-probe-*`, cleaned up at the end (verified
empty — see §4 closing command).

> **Corrected 2026-08-08** after re-verification: the submodule count was **5**
> in three places; the true count is **10** (all SSH). Measured:
> `git -C ~/dev/mc-metarepo config -f .gitmodules --get-regexp "submodule\..*\.url"`
> → 10 lines, all `git@github.com:` (monorepo-loop, vecna-api, vecna-front,
> cleaq-api, cleaq-front, return-app, wordpress, support-engineer, risk,
> pim-migration-sync). This matches `kb-mc-metarepo.md` ("10 registered
> submodules, all EMPTY"). Nothing else in this file changes.

## Summary

- **Recommended mechanism: `hermes cron` in `--no-agent --script` mode.**
  Zero LLM tokens per tick (script IS the job), persists across gateway
  restart (`~/.hermes/cron/jobs.json`, reloaded every tick), the gateway
  `--user` unit is already `enabled` + `Linger=yes` so the 60 s ticker runs
  headless today, and failure is auto-delivered (non-zero exit / timeout →
  error alert on the `--deliver` target) with no extra code.
- **Recommended pull shape: `git fetch` + `git reset --hard origin/main`**,
  never plain `git pull` or `git pull --rebase` unattended. Measured:
  `fetch+reset--hard` is the *only* strategy that succeeds (exit 0) in every
  dirty-tree, conflict, and rewritten-history case tested — it always
  converges to origin's exact state. The other three all have an unattended
  dead-end (§Failure matrix).
- **Worst failure mode measured: a silently-dropped network path hangs `git
  fetch` with zero output and no built-in timeout** — confirmed by killing it
  externally at 8 s (exit 124) after `http.lowSpeedTime` did *not* bound the
  TCP-connect phase. An hourly job **must** wrap the pull in `timeout <n>`.
- **Second-worst: `git pull --rebase` on conflict leaves `.git/rebase-merge`
  + a detached HEAD that a bare `git reset --hard` does NOT clear** —
  measured directly (§Test 6b). A guard (`git rebase --abort` if
  `.git/rebase-merge`/`.git/rebase-apply` exists) before `fetch+reset--hard`
  is required for full self-healing; confirmed that guard recovers fully.
- **`hermes cron`'s `--script` mode needs `~/.hermes/scripts/`, which does
  not exist yet** (`ls`: No such file or directory) — first job creation
  needs to create the dir (or `hermes cron create` may do it; not tested,
  see Not tested).
- mc-metarepo has **10 submodules on `git@github.com:...` SSH URLs**, and
  `submodule.recurse`/`fetch.recursesubmodules` are both unset — a plain
  superproject `git pull`/`reset --hard` does **not** touch submodules or
  their SSH auth. Out of scope for "pull the top-level clone" but a real gap
  for a KB that includes submodule content — flag for the design doc, not
  solved here.
- Neither systemd `--user` timers nor plain cron have anything configured
  for this today (`list-timers` shows only a stock Ubuntu timer; `crontab -l`
  is empty) — greenfield either way. `Linger=yes` means a `--user` timer
  *would* fire with nobody logged in, same as the gateway already does.
- Observability today, if wired via `hermes cron`: (1) `hermes cron list` /
  `runs` show `last_status`/`last_error`; (2) every run — success or
  failure — writes `~/.hermes/cron/output/<job_id>/<ts>.md` regardless of
  delivery outcome; (3) `~/.hermes/logs/agent.log` gets `cron.scheduler`
  INFO lines (`Job 'X' completed successfully` / delivery lines); (4) a
  non-zero exit or timeout auto-delivers an error alert to `--deliver`. A
  plain systemd timer or crontab job has none of this for free — only
  journald/syslog, which would need to be found and read manually.

---

## 1. `hermes cron` — syntax, persistence, cost model

`~/.local/bin/hermes cron --help` and every subcommand's `--help` (verbatim,
condensed to the load-bearing parts — full transcripts were captured, this
is the syntax that matters):

```
usage: hermes cron [-h] [--accept-hooks]
                   {list,create,add,edit,pause,resume,run,remove,rm,delete,status,runs,history,notepad,tick}
```

```
usage: hermes cron create [-h] [--name NAME] [--deliver DELIVER]
                          [--repeat REPEAT] [--skill SKILLS] [--script SCRIPT]
                          [--no-agent] [--monitor-script MONITOR_SCRIPT]
                          [--monitor-url MONITOR_URL] [--workdir WORKDIR]
                          [--model MODEL] [--provider MODEL_PROVIDER]
                          schedule [prompt]
  schedule              Schedule like '30m', 'every 2h', or '0 9 * * *'
  --deliver DELIVER     Delivery target: origin, local, telegram, discord,
                        signal, or platform:chat_id
  --script SCRIPT       Path to a script under ~/.hermes/scripts/. Default
                        mode: script stdout is injected into the agent's
                        prompt each run. With --no-agent: the script IS the
                        job and its stdout is delivered verbatim.
  --no-agent            Skip the LLM entirely — run --script on schedule and
                        deliver its stdout directly. Empty stdout = silent.
```

**Exact syntax for the KB refresh job** (NOT created — CLI syntax only,
composed from `--help` + the doc pages below):

```
hermes cron create "every 1h" \
  --no-agent \
  --script kb-refresh.sh \
  --deliver slack \
  --name "mc-metarepo-refresh"
```

`--repeat` is omitted (recurs forever until paused/removed); `--deliver
slack` with no target = `SLACK_HOME_CHANNEL`, per known fact and confirmed
live below.

**Schedule format** — 4 kinds, from
`~/.hermes/hermes-agent/website/docs/developer-guide/cron-internals.md`
("Scheduling Model" table) and cross-checked against the live job's stored
JSON:

| Format | Example | Behavior |
|---|---|---|
| Relative delay | `30m`, `2h`, `1d` | One-shot after the duration |
| Interval | `every 2h`, `every 30m` | Recurring |
| Cron expression | `0 9 * * *` | Standard 5-field cron |
| ISO timestamp | `2025-01-15T09:00:00` | One-shot at exact time |

**Cost model — confirmed the "no-agent" doc claim is real, not aspirational**:
default mode creates a fresh `AIAgent` session and runs the prompt through
the model every tick (`cron/scheduler.py` tick cycle, per the internals doc:
"Create fresh AIAgent session... Run the job prompt through the agent").
`--no-agent` skips that entirely — "Zero tokens, zero agent loop, zero model
spend" (`cron-script-only.md`). For an hourly git-pull, `--no-agent
--script` is the only mode that doesn't burn a model call every hour just to
run `git fetch`.

**Persistence** — jobs live in `~/.hermes/cron/jobs.json`, plain JSON, atomic
write (temp file + rename). Read-only cat of the one existing job:

```
$ cat ~/.hermes/cron/jobs.json
{
  "jobs": [{
    "id": "004adebba8ee", "name": "wf4-p13",
    "schedule": {"kind": "once", "run_at": "...", "display": "once in 1m"},
    "no_agent": false, "script": null,
    "enabled": false, "state": "completed", "last_status": "ok",
    "deliver": "slack", ...
  }],
  "updated_at": "2026-08-07T19:06:01.624848+00:00"
}
```

Survives a gateway restart: the scheduler tick reloads `jobs.json` from disk
every cycle (internals doc, Tick Cycle step 2: "Load all jobs from
jobs.json"), and the gateway `--user` unit is `enabled` (starts at boot) with
`loginctl show-user gaetan` → `Linger=yes` (confirmed live, §2) — so the
60 s in-process ticker (`InProcessCronScheduler`, confirmed in
`agent.log`: `"In-process cron scheduler started (interval=60s)"`) comes back
up unattended after any restart or reboot, no login required.

**What runs the job**: an agent TURN (costs tokens) by default; a shell
script with **zero** agent turn under `--no-agent`. Script rules (doc +
`--help`): must live under `~/.hermes/scripts/` (enforced at create and run
time — no arbitrary paths, no `#!/...` shebang honoured); `.sh`/`.bash` run
via `bash`, everything else via the current Python interpreter; default
timeout 3600 s, tunable via `HERMES_CRON_SCRIPT_TIMEOUT` /
`cron.script_timeout_seconds`.

**Output/failure → delivery mapping** (`cron-script-only.md`, "How Script
Output Maps to Delivery"):

| Script behavior | Result |
|---|---|
| Exit 0, non-empty stdout | stdout delivered verbatim |
| Exit 0, empty stdout | Silent tick, no delivery |
| Exit 0, stdout ends `{"wakeAgent": false}` | Silent tick |
| Non-zero exit | **Error alert delivered** |
| Script timeout | **Error alert delivered** |

**Existing jobs / list / status** (read-only, nothing created):

```
$ hermes cron list
No scheduled jobs.

$ hermes cron list --all
  004adebba8ee [completed]
    Name:      wf4-p13
    Schedule:  once in 1m
    Last run:  2026-08-07T19:06:01.624648+00:00  ok

$ hermes cron status
✓ Gateway is running — cron jobs will fire automatically
  PID: 76255
  Ticker heartbeat: 39s ago
  No active jobs

$ hermes cron runs --limit 10
f99a1bdc74b74f89944499c67d237506  completed  job=004adebba8ee  source=builtin  2026-08-07T19:05:58.086734+00:00
```

Only that one prior test job exists (from wf4-p13, confirms the known
`--deliver slack` bare-platform → home-channel fact live). No KB job present.

**Per-run local artifact** (independent of delivery success — a durable
audit trail even if Slack delivery fails):

```
$ find ~/.hermes/cron/output/004adebba8ee -type f
~/.hermes/cron/output/004adebba8ee/2026-08-07_19-06-01.md
```
contains job name, run time, schedule, full prompt, and full response —
written every run, regardless of `--deliver` target.

**`~/.hermes/scripts/` does not exist yet**:
```
$ ls -la ~/.hermes/scripts/
ls: cannot access '/home/gaetan/.hermes/scripts/': No such file or directory
```
Not tested whether `hermes cron create --script` auto-creates it — see
"Not tested".

---

## 2. systemd `--user` timers

```
$ export XDG_RUNTIME_DIR=/run/user/$(id -u)
$ systemctl --user list-timers --all
NEXT                        LEFT LAST                        PASSED UNIT                           ACTIVATES
Sat 2026-08-08 16:19:51 UTC  17h Fri 2026-08-07 16:19:51 UTC 6h ago launchpadlib-cache-clean.timer launchpadlib-cache...
1 timers listed.
```
Only a stock Ubuntu timer exists — nothing KB-related. Created/enabled
nothing.

```
$ loginctl show-user gaetan | grep -i linger
Linger=yes
```
Linger is on, so a `--user` timer **would** fire with nobody logged in — and
this is already proven in practice, not just theoretical: the gateway itself
is a `--user` unit that's `enabled` and running unattended:
```
$ systemctl --user is-enabled hermes-gateway.service
enabled
$ systemctl --user status hermes-gateway.service | head -6
● hermes-gateway.service - Hermes Agent Gateway - Messaging Platform Integration
     Loaded: loaded (...; enabled; preset: enabled)
     Active: active (running) since Fri 2026-08-07 19:40:45 UTC; 2h 57min ago
```

A hand-rolled `--user` timer is a viable fallback if `hermes cron` turned out
unsuitable, but it buys nothing `hermes cron --no-agent` doesn't already give
for free (persistence, delivery-on-failure, a local output artifact) — plus
it's a second unit to babysit outside Hermes' own state.

---

## 3. plain cron

```
$ crontab -l
no crontab for gaetan

$ ps aux | grep -i cron
root       24145  0.0  0.0   7224  2756 ?        Ss   16:13   0:00 /usr/sbin/cron -f -P

$ systemctl status cron | head -8
● cron.service - Regular background program processing daemon
     Loaded: loaded (...; enabled; preset: enabled)
     Active: active (running) since Fri 2026-08-07 16:13:33 UTC; 6h ago
```
`crond` is running system-wide (root-owned `cron.service`) but gaetan has no
crontab. Available as a fallback (no `hermes` dependency at all, survives
even a broken Hermes install) but has zero of the observability hooks in §6
— you'd be building delivery/logging from scratch. Created nothing.

---

## 4. Failure modes for an hourly `git pull` (measured in throwaway sandboxes)

Real clone's actual transport, read-only inspection (informs what "auth
expiry" would really look like):

```
$ cd ~/dev/mc-metarepo && git remote -v
origin  https://github.com/mobile-club/metarepo.git (fetch/push)
$ git config --list --show-origin | grep -i credential
file:/home/gaetan/.gitconfig  credential.https://github.com.helper=!/usr/bin/gh auth git-credential
$ gh auth status
  ✓ Logged in to github.com account GaetanCathelain
  Token scopes: 'gist', 'read:org', 'repo'
```
Auth is `gh`'s stored OAuth token via its git-credential helper, scoped to
`github.com` HTTPS only. Real "auth expiry" = `gh`'s token expiring/being
revoked, which surfaces to `git fetch` exactly like the 404 proxy tested
below (`gh auth git-credential` returns nothing usable → GitHub answers
"Repository not found" for private repos with a bad/expired token — GitHub
does not distinguish "doesn't exist" from "you can't see it").

Also noted: **10 submodules on SSH URLs** (`git@github.com:mobile-club/...`
etc.), `submodule.recurse`/`fetch.recursesubmodules` both unset — a plain
top-level `git pull`/`reset --hard` never touches them, so submodule SSH
auth is a separate, untested surface (see Not tested).

Sandbox construction (throwaway, `/tmp` only, one fresh bare `origin.git` +
seed commit per test group — never touched `~/dev/mc-metarepo`):
```
git init --bare -q origin.git
git init -q -b main seed && cd seed && ... && git push -q origin main
git clone -q ../origin.git clone
```

**Test 1 — dirty tree, upstream advances a file the dirty edit doesn't
touch:**
- `git pull` (merge): fast-forwards clean, **keeps the unrelated dirty
  edit**. `EXIT=0`.
- `git pull --ff-only`: same — fast-forward succeeds, dirty edit survives.
  `EXIT=0`.
- `git fetch && git reset --hard origin/main`: **silently discards the dirty
  edit** (file reverts to committed content). `EXIT=0`.
- `git pull --rebase`: refuses outright, dirty or not:
  `error: cannot pull with rebase: You have unstaged changes. error: Please
  commit or stash them.` `EXIT=128`.

**Test 2 — dirty tree, upstream advances the SAME line the dirty edit
touches:**
- `git pull` (merge): fetches, then aborts *before* touching files:
  `error: Your local changes to the following files would be overwritten by
  merge: file.txt / Please commit your changes or stash them before you
  merge. / Aborting`. `EXIT=1`. Dirty edit intact, tree stays clean (no
  stuck merge state).
- `git pull --ff-only`: identical abort, same message. `EXIT=1`.
- `git fetch && git reset --hard origin/main`: succeeds, dirty edit
  discarded, file now matches upstream. `EXIT=0`.
- `git pull --rebase`: same unconditional refusal as Test 1. `EXIT=128`.

**Test 3 — network down (connection refused, `127.0.0.1:1`):**
```
$ git fetch
fatal: unable to access 'https://127.0.0.1:1/does-not-exist.git/': Failed to connect to 127.0.0.1 port 1 after 0 ms: Couldn't connect to server
EXIT=128
$ git pull --ff-only
(same fatal line)
EXIT=1
```
Instant, clean failure — `git pull` and `git fetch` return **different exit
codes** for the identical underlying error (128 vs 1); don't gate on a
specific code, gate on non-zero.

**Test 3b — network down, silently dropped (non-routable `10.255.255.1`, no
RST):**
```
$ timeout 8 git fetch
EXIT=124   # killed externally — git itself produced no output and did not time out on its own
```
`http.lowSpeedTime`/`http.lowSpeedLimit` do **not** bound the TCP-connect
phase, only a stalled-but-connected transfer. **Without an external
`timeout` wrapper, a dropped-not-refused network path hangs the job
indefinitely.** This is the single most important operational finding here.

**Test 4 — auth-failure proxy (private/nonexistent repo, headless):**
```
$ GIT_TERMINAL_PROMPT=0 git fetch
remote: Repository not found.
fatal: repository 'https://github.com/mobile-club/definitely-nonexistent-private-repo-xyz123.git/' not found
EXIT=128
```
With `GIT_TERMINAL_PROMPT=0` this fails fast and clean (0.4 s). **Without
it**, a bad-credential HTTPS remote can block on an interactive credential
prompt that never resolves headless — always set this env var for the
unattended job.

**Test 5 — rewritten upstream history (force-push / amended tip; clean local
tree, local == old tip, no ancestor relationship to new tip):**
- `git pull` (merge): fetch force-updates the local `origin/main` ref, then
  the merge step refuses: `hint: You have divergent branches... fatal: Need
  to specify how to reconcile divergent branches.` `EXIT=128`.
- `git pull --ff-only`: `hint: Diverging branches can't be fast-forwarded...
  fatal: Not possible to fast-forward, aborting.` `EXIT=128`.
- `git fetch && git reset --hard origin/main`: succeeds, lands exactly on
  the new tip. `EXIT=0`.
- `git pull --rebase`: **succeeds silently** — `Successfully rebased and
  updated refs/heads/main.` `EXIT=0`. (Local had zero commits ahead of the
  old tip, so there was nothing to replay; this only stays this cheap
  because the mirror never carries local commits — see Test 6b for what
  happens when it does.)

**Test 6 — can an unattended NEXT tick self-heal after a failed run, with no
operator and no manual `stash`/`abort`?**

*6a — plain `git pull` (merge) after a real conflict leaves a clean-but-stuck
state:*
```
$ git pull      # first tick
fatal: Need to specify how to reconcile divergent branches.
EXIT=128
$ git pull      # second tick, unattended, identical command
fatal: Need to specify how to reconcile divergent branches.
EXIT=128        # STUCK — same error forever, no self-heal
$ git fetch && git reset --hard origin/main   # recovery via the OTHER strategy
HEAD is now at f3424b5 ...
EXIT=0          # fully recovers
```
`git pull` alone never digs itself out; `fetch+reset--hard` does, on the
first try, from that exact stuck state.

*6b — `git pull --rebase` hitting a real conflict leaves the clone
detached-HEAD with a live rebase in progress, and `reset --hard` alone does
NOT clear it:*
```
$ git pull --rebase
CONFLICT (content): Merge conflict in file.txt
error: could not apply 7aa59df...
EXIT=1
$ git status --short --branch
## HEAD (no branch)
UU file.txt
$ ls -d .git/rebase-merge
.git/rebase-merge
$ git fetch && git reset --hard origin/main    # naive unattended recovery attempt
HEAD is now at 62b76fc ...
EXIT=0                                          # exit 0 — LOOKS fine
$ ls -d .git/rebase-merge
.git/rebase-merge                               # ← still there
$ git status --short --branch
## HEAD (no branch)                             # ← still detached, still "rebasing"
```
**`git reset --hard` reports success (exit 0) while the repo is still stuck
mid-rebase and HEAD is still detached.** Confirmed separately that
`git checkout -B main origin/main` does **not** fix it either — it refuses
with `error: you need to resolve your current index first` while
`.git/rebase-merge` exists. The only recipe that actually recovers:
```
[ -d .git/rebase-merge ] || [ -d .git/rebase-apply ] || [ -f .git/MERGE_HEAD ] && git rebase --abort  # or merge --abort
git fetch origin && git reset --hard origin/main
```
Verified this exact sequence fully recovers: `.git/rebase-merge` gone, `git
status` back to `## main...origin/main`, content matches origin.

Sandbox cleanup verified:
```
$ rm -rf /tmp/kb-probe-* /tmp/probe-git-failures*.sh
$ ls /tmp/ | grep -i kb-probe
(no output)
```

---

## 5. Safest pull shape — verdict

**`git fetch origin && git reset --hard origin/main`**, guarded by a
pre-check that aborts any leftover `rebase-merge`/`rebase-apply`/`MERGE_HEAD`
state, wrapped in `timeout <n>`, with `GIT_TERMINAL_PROMPT=0` set. Measured
justification, directly from the tests above:

- It is the **only** strategy that returns `EXIT=0` in every single test
  (dirty non-conflicting, dirty conflicting, network-refused N/A since it
  never runs mid-failure, rewritten history, and — critically — recovering
  from *another* strategy's stuck state in 6a).
- It's the only strategy that's **unconditionally idempotent**: every tick
  converges the working tree to exactly `origin/main`, dirty state or not,
  history-rewrite or not. That is the correct semantics for a read-only KB
  mirror — nobody should ever be editing `~/dev/mc-metarepo` by hand, so
  "discards local state" is a feature, not a risk, here.
- Plain `git pull` is explicitly the wrong choice: it's the strategy that
  gets stuck forever on divergent-history/conflict (6a) — the retry-next-hour
  design this proposal wants does not save it, because the same command
  produces the same fatal error every tick.
- `--ff-only` behaves identically to plain merge-pull on every conflict/
  divergence case tested (both refuse) but adds no benefit over
  `fetch+reset--hard` and still leaves the tree stuck.
- `--rebase` is the only strategy that can leave the clone in a state a bare
  `reset --hard` cannot fully clean (6b) — never use it unattended on a
  mirror that might ever gain a stray local commit (e.g. someone testing by
  hand, or a bug).

---

## 6. Observability — how would anyone know a 03:00 pull failed?

Ranked by what's already free vs. what would need building:

1. **`hermes cron`, `--no-agent --script`, non-zero exit or timeout →
   automatic error alert delivered to `--deliver` target** (confirmed
   doc behavior, consistent with the live delivery mechanics already
   verified for `wf4-p13`: `Job '004adebba8ee': delivered to
   slack:D0BBYNM01BL via live adapter` in `agent.log`). This is the
   headline reason to use `hermes cron` over a bare systemd timer or
   crontab — failure notification is the default, not something to build.
2. **Local per-run artifact, always written**:
   `~/.hermes/cron/output/<job_id>/<timestamp>.md` — exists independent of
   whether Slack delivery itself succeeds, so it survives a simultaneous
   Slack outage.
3. **`~/.hermes/logs/agent.log`** carries `cron.scheduler` INFO lines per
   run (`Job 'X' completed successfully`, delivery confirmation lines) —
   confirmed live for the existing job. A failed script run would show the
   matching failure/error line here (not tested — no failing job exists
   yet to observe the exact failure log line shape; see Not tested).
   `errors.log` is presumably where an exception traceback would land
   (empty today — no failures to date).
4. **`hermes cron list --all` / `hermes cron runs`** — `last_status` /
   `last_error` fields on the job record, and a durable execution-attempt
   ledger, queryable anytime without grepping logs.
5. journald: per project facts, journald carries **WARNINGs only** — a
   cron script failure logged at INFO would not appear there at all; don't
   rely on `journalctl --user` for this.

A bare systemd `--user` timer or plain crontab job has **none** of 1–4 for
free — failure surfaces only via `systemctl --user status <unit>` /
`journalctl --user -u <unit>` (if you remember to check) or whatever the
script itself is coded to do. `hermes cron --no-agent` gets a
failure-delivers-itself default at zero extra engineering cost, which is the
main reason it beats the OS-level options for this use case.

---

## Failure matrix

| Failure | `git pull` (merge) | `git pull --ff-only` | `git fetch && git reset --hard` | `git pull --rebase` |
|---|---|---|---|---|
| Dirty tree, non-conflicting upstream change | OK, keeps dirty edit — exit 0 | OK, keeps dirty edit — exit 0 | OK, **discards** dirty edit — exit 0 | Refuses unconditionally — exit 128 |
| Dirty tree, conflicting upstream change | Clean abort, no file touched — exit 1 | Clean abort, no file touched — exit 1 | OK, discards dirty edit — exit 0 | Refuses unconditionally — exit 128 |
| Network down, connection refused | fatal, instant — exit 128 (fetch) / 1 (pull) | same | same (fetch leg) | same (fetch leg) |
| Network down, silently dropped (no RST) | **hangs**, no built-in timeout — killed externally at 124 | same | same | same |
| Auth/token expiry (proxied via 404 private repo) | `remote: Repository not found.` fatal — exit 128, fast (0.4 s) *with* `GIT_TERMINAL_PROMPT=0`; can hang on a credential prompt without it | same | same (fetch leg) | same (fetch leg) |
| Force-push / rewritten upstream history | `fatal: Need to specify how to reconcile divergent branches.` — exit 128 | `fatal: Not possible to fast-forward, aborting.` — exit 128 | OK, lands on new tip — exit 0 | OK if local has 0 unique commits (rebase is a no-op) — exit 0; **untested with real local commits** beyond 6b's conflict case |
| Merge conflict → next unattended tick, same command, no operator | **stuck forever**, exit 128 every time (6a) | not tested identically, but same refusal class expected (untested combo) | **self-heals**, exit 0 (6a) | Leaves `.git/rebase-merge` + detached HEAD; a bare `reset --hard` does **not** clear it (still exit 0 but repo still broken) — needs an explicit `rebase --abort` guard first (6b) |

---

## Not tested

- Real GitHub token expiry/revocation (only proxied via a 404 private-repo
  response, which produces the same class of error GitHub gives for bad
  auth on a private repo — not a bitwise-identical reproduction).
- Real network partition / packet loss mid-transfer (only tested TCP-connect
  hang on a non-routable IP and instant connection-refused; a stall
  *mid-download* against `http.lowSpeedTime` was not exercised).
- Whether `hermes cron create --script <name>` auto-creates
  `~/.hermes/scripts/` if it doesn't exist, or requires the operator to
  `mkdir` it first (directory doesn't exist today; didn't create a job to
  find out, per the create-nothing constraint).
- The exact `agent.log` / `errors.log` line shape for a **failed** cron run
  (no failing job has ever run on this VM to observe it against).
- `git pull --rebase` behavior with real local commits diverging from a
  rewritten upstream tip (Test 5's rebase case had zero local commits ahead,
  so no actual patch replay/conflict was exercised there — only Test 6b
  exercised a real rebase conflict, and that was against a *non-rewritten*,
  simply-advanced upstream).
- Submodule refresh/auth (10 submodules on SSH URLs; a plain top-level pull
  never touches them — confirmed only that `submodule.recurse` /
  `fetch.recursesubmodules` are unset, not what pulling them would look
  like).
- Chronos managed-cron provider (`cron.provider: chronos`) — not configured
  on this VM (`cron:` section absent from `config.yaml`), irrelevant since
  the gateway already runs as an always-on `--user` service, not
  scale-to-zero.
- Whether a `--workdir ~/dev/mc-metarepo` cron job auto-injects
  `AGENTS.md`/`CLAUDE.md` context that could confuse a `--no-agent` script
  run (doc says `--workdir` affects agent-mode context injection and cwd;
  not exercised since no job was created).
