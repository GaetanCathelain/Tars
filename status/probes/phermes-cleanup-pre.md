# p-Hermes gated cleanup — pre-delete inventory evidence

Date: 2026-08-13. Read-only probe, no writes/deletes performed by this agent.
Gaetan gave explicit go for the gated cleanup this session; the hub session
executes the delete itself using this evidence.

## Spec basis (what "old profile" means, and the gate history)

- `PLAN.md:14` — "the profile *delete* is deferred to a gated post-WF4
  cleanup ... It is a gated inline step, never inside an unattended graph."
- `PLAN.md:193-194` (Amendments) — "Cutover: disable-only (the Slack token
  pair frees when the old gateway stops). Profile deletion is a separate,
  gated cleanup after WF4 passes + soak."
- `PLAN.md:224-226` (Guardrails) — "Lane B never writes to personal Hermes
  (read-only probe only); the cutover only *disables* the old tars gateway
  (rollback = re-enable); the profile delete happens exactly once, in a
  gated post-WF4 cleanup, with before-evidence captured first."
- `docs/specs/tars-profile.md:272-276` — ordering constraint: old
  `hermes-gateway-tars.service` on p-Hermes disabled+stopped before the new
  gateway starts (done at cutover); "The p-Hermes profile *delete* is a
  separate gated cleanup after WF4 + soak, not part of the cutover."
- `docs/recon/r3-p-hermes.md` §5 (lines 137-201) — inventory-level recon:
  profile dir listing, unit file + override.conf verbatim, `.env` key names
  (values never read), `config.yaml` guardrail values, `auth.json` top-level
  keys, Slack manifest identifiers, sizes and modes.
- `docs/specs/wf4-probes.md:507-527` (§23) — originally required **byte-level**
  before-evidence (tar + sha256 manifest off-host) before the delete.
- `status/wf4-report.md:113-120` — §4 verdict PARTIAL: only inventory-level
  evidence existed; byte-level capture did not; "not a blocker for declaring
  Tars live; it **is** a blocker for the delete → probe 23."
- `status/wf4-report.md:174-175` — **§23 formally DROPPED**: "the gated
  p-Hermes profile delete will rest on the inventory-level evidence in
  `r3-p-hermes.md` §3–5 only (Gaetan's call)." This supersedes the
  byte-level/archive requirement — no tar/checksum manifest is needed.
- `status/lane-a.md:341` — pre-drop status note ("needs §23 byte-level
  before-evidence first"), superseded by the `wf4-report.md:174-175` ruling
  above.
- `status/lane-a.md:343-351`, `status/probes/cutover.md:49` — cutover
  executed 2026-08-07: old `hermes-gateway-tars.service` on p-Hermes
  disabled+inactive (confirmed by the orchestrator before the WF3-verify run;
  re-confirmed live by this probe below).

**"Old profile" = exactly one directory**, no other doc extends the gated
delete's scope: `/home/hermes/.hermes/profiles/tars/` on p-Hermes (Proxmox
VM 103, `192.168.0.8`, user `hermes`, Hermes v0.19.0). The unit file and
override.conf (below) are **not** named as in-scope for this delete by any
spec revision — flagged as a leftover, left untouched pending Gaetan's call.

## Target path

```
/home/hermes/.hermes/profiles/tars/
```
on p-Hermes, reached only via the Tars VM: `ssh gaetan@192.168.0.9 "ssh
phermes '<cmd>'"` (p-Hermes refuses direct SSH from elsewhere; never touch
`192.168.0.3` / pve, a different machine).

## Pre-delete inventory (measured 2026-08-13, this probe)

```
$ ssh gaetan@192.168.0.9 "ssh phermes 'ls -la /home/hermes/.hermes/profiles/tars/'"
total 11984
drwx------ 22 hermes hermes    4096 Aug  7 18:36 .
drwxrwxr-x  8 hermes hermes    4096 Aug  4 14:20 ..
drwx------  2 hermes hermes    4096 Jun 20 08:58 audio_cache
-rw-------  1 hermes hermes    2573 Aug  4 10:12 auth.json
-rw-rw-r--  1 hermes hermes       0 Jun 20 08:58 auth.lock
drwxrwxr-x  2 hermes hermes    4096 Jun 20 09:37 bin
drwxrwxr-x  4 hermes hermes    4096 Jul 13 15:45 cache
-rw-------  1 hermes hermes    7101 Aug  7 18:31 channel_directory.json
-rw-rw-r--  1 hermes hermes       0 Aug  7 18:36 .clean_shutdown
-rw-------  1 hermes hermes   12499 Aug  4 12:57 config.yaml
-rw-------  1 hermes hermes   12542 Jun 20 13:25 config.yaml.bak-20260620132605
-rw-------  1 hermes hermes   13126 Jul 12 21:48 config.yaml.pre-tars-security-audit-20260712
-rw-rw-r--  1 hermes hermes      73 Jun 20 09:37 context_length_cache.yaml
drwx------  3 hermes hermes    4096 Aug  7 18:36 cron
-rw-------  1 hermes hermes     518 Jun 20 10:28 .env
-rw-rw-r--  1 hermes hermes     197 Aug  5 22:41 gateway.lock
-rw-rw-r--  1 hermes hermes     107 Aug  5 22:41 gateway-starts.log
-rw-------  1 hermes hermes     403 Aug  7 18:36 gateway_state.json
drwxrwxr-x  2 hermes hermes    4096 Jun 20 08:58 home
drwx------  2 hermes hermes    4096 Jun 20 08:58 hooks
drwx------  2 hermes hermes    4096 Jun 21 13:06 image_cache
drwx------  3 hermes hermes    4096 Jul 28 16:35 logs
drwx------  2 hermes hermes    4096 Jun 20 08:58 memories
-rw-------  1 hermes hermes 3491174 Aug  4 14:14 models_dev_cache.json
-rw-rw-r--  1 hermes hermes     151 Jun 20 08:58 .no-bundled-skills
drwx------  2 hermes hermes    4096 Jun 20 08:58 pairing
drwxrwxr-x  2 hermes hermes    4096 Jul 28 16:35 pending_messages
drwxrwxr-x  2 hermes hermes    4096 Jun 20 08:58 plans
drwxrwxr-x  3 hermes hermes    4096 Jul 13 14:46 platforms
drwxrwxr-x  2 hermes hermes    4096 Aug  1 18:36 plugins
-rw-rw-r--  1 hermes hermes     167 Jun 20 08:58 profile.yaml
drwxrwxr-x  3 hermes hermes    4096 Jun 20 09:37 sandboxes
drwx------  2 hermes hermes    4096 Aug  4 10:12 sessions
drwx------  5 hermes hermes    4096 Aug  4 14:14 skills
-rw-------  1 hermes hermes    3105 Aug  4 12:56 .skills_prompt_snapshot.json
drwxrwxr-x  2 hermes hermes    4096 Jun 20 08:58 skins
-rw-rw-r--  1 hermes hermes   13326 Jun 20 10:41 slack-manifest.json
-rw-rw-r--  1 hermes hermes   13326 Jun 20 10:43 slack-manifest.tars.redacted.json
-rw-rw-r--  1 hermes hermes    2046 Jun 20 18:51 SOUL.md
drwxrwxr-x  2 hermes hermes    4096 Aug  7 18:36 state
-rw-r--r--  1 hermes hermes 8515584 Aug  4 14:30 state.db
-rw-rw-r--  1 hermes hermes      69 Aug  4 12:56 .update_check
-rw-r--r--  1 hermes hermes   32768 Aug  4 14:26 verification_evidence.db
drwxrwxr-x  2 hermes hermes    4096 Jun 20 08:58 workspace
```

Aggregate:
```
$ du -sh /home/hermes/.hermes/profiles/tars/
57M

$ find /home/hermes/.hermes/profiles/tars/ -type f | wc -l
79 files

$ find /home/hermes/.hermes/profiles/tars/ -type d | wc -l
49 dirs (incl. root)
```

Full recursive `find -printf` (name/type/mode/size/mtime only, no content
reads) was captured in-session; largest entries are `bin/tirith` (22.7 MB
helper binary), `state.db` (8.5 MB), `models_dev_cache.json` (3.5 MB),
`cache/openrouter_model_metadata.json` (215 KB), `verification_evidence.db`
(32 KB) — the rest (skills tree, cron, sessions, memories, logs, curator
backups under `skills/.curator_backups/`) is small config/state files.

**Adjacent systemd paths (not part of the delete target — flagged only):**
```
$ ls -la /home/hermes/.config/systemd/user/hermes-gateway-tars.service
-rw-rw-r-- 1 hermes hermes 1000 Jul 24 13:57 .../hermes-gateway-tars.service

$ ls -la /home/hermes/.config/systemd/user/hermes-gateway-tars.service.d/
-rw------- 1 hermes hermes 55 Jun 20 10:29 override.conf

$ ls -la /home/hermes/.config/systemd/user/default.target.wants/hermes-gateway-tars.service
No such file or directory   <- the enable-symlink was already removed by
                                `disable --now` at cutover; expected.
```
The unit file and its `.d/override.conf` still exist on disk but are inert
(disabled, no enable-symlink). No spec revision names them as in scope for
this gated cleanup (`PLAN.md` Amendments/Guardrails and `tars-profile.md:275`
both say "the profile *delete*", singular, meaning the directory). Left
untouched — a follow-up call for Gaetan if he wants them removed too.

**Gateway state, re-confirmed live (not from stale recon):**
```
$ systemctl --user is-active hermes-gateway-tars.service
inactive
$ systemctl --user is-enabled hermes-gateway-tars.service
disabled
$ ps -eo pid,cmd | grep -i "profile tars" | grep -v grep
(no output — no live process under this profile)
```

## Divergence from recon

**None found.** Every file r3-p-hermes.md §5 recorded by size matches this
probe exactly: `.env` 518 bytes, `auth.json` 2,573 bytes, `config.yaml`
12,499 bytes, `slack-manifest.json` / `slack-manifest.tars.redacted.json`
13,326 bytes each. `.env`'s mtime (Jun 20) is older than the cutover
(Aug 7) — consistent with the cutover harvesting the Slack token pair by
copy, not rotation, per `status/lane-a.md:343-345` ("harvested at cutover;
files→SOPS→shred" refers to the shredded copy on the Tars side, not this
file). Files newer than recon (`gateway_state.json`, `.clean_shutdown`,
`channel_directory.json`, all Aug 7 18:3x) are the expected artifacts of the
cutover's `disable --now` and clean shutdown — not a surprise.

No secret values were read at any point (names/sizes/modes/mtimes only, per
the hard rule).

## Recommended delete command

Run from the Tars VM (never direct — p-Hermes refuses direct SSH):

```
ssh gaetan@192.168.0.9 "ssh phermes 'rm -rf /home/hermes/.hermes/profiles/tars/'"
```

Pre-conditions already verified true above: gateway `inactive`+`disabled`,
no live process, no enable-symlink. No further readiness check needed.
