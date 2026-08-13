# hermes-lcm removal + GCN-49 job verification — 2026-08-13

Standing order from Gaetan: switch `context.engine` from `lcm` back to
`compressor` and remove `hermes-lcm` entirely. Executed on the Tars VM
(192.168.0.9) at 15:15–15:25 UTC. Nothing was sent to Slack by this session.

## 1. Incident verdict — the 23:55 GCN-49 job (`982c896036bc`)

**job_claim_true = yes. Job left untouched — it already matches Gaetan's
instruction.** Tars' 15:13 UTC claim ("j'ai corrigé et relu le job") is true.

Verified independently of the recon, from `~/.hermes/cron/jobs.json`:

| Check | Result |
|---|---|
| Created | `2026-08-13T17:11:21.239326+02:00` = 15:11:21 UTC — inside the 15:03–15:13 correction window |
| Names the **Gaetan–Damien** DM by id | `D08BJ8CQLP6` present in prompt ✓ |
| Excludes the origin thread and the Tars DM | "pas seulement le thread" ✓ / "pas le DM entre Gaetan et Tars" ✓ |
| Gates the write on Damien's user id | `U7UC03YV6` ✓, after ts `1786633197.554209` ✓ |
| Write target | `GCN-49`, state `0434e579-7b85-487a-8cf9-5aed6caaf41b` (GCN Done) ✓ |
| Delivery | `deliver: slack` (home-channel token, not `origin`); no `continuable` / `mirror_delivery` key → fire-and-forget → **new top-level message in `D0BBYNM01BL`**, not a reply into thread `1786633431.681329` ✓ |

The wrong-conversation aim that probing caught earlier was genuinely fixed
before the job was left to run. **No edit made.** Prompt integrity is pinned
by hash so the survival check below is exact: `len=1032`,
`sha256[:16]=07107ada3e58068c`.

Engagement-checker conflict: **none**. `~/.hermes/state/engagement-checker.json`
key `slack:D08BJ8CQLP6:1786633197.554209` is `status: open`, no snooze; the
checker's next run is 2026-08-14T10:00 (+02:00), after tonight's job. Per
engagement-checker SKILL.md v2.0.1 §5/§5a it reconciles an already-terminal
issue to `done` and skips the write — no double-report, no re-open.

## 2. Config switch

`.bak` first, then edited under `flock ~/.hermes/.wf3.lock`, tmp+mv, line-precise
(no YAML round-trip, so nothing else could be reformatted or reordered).

Backup: `~/.hermes/config.yaml.bak-2026-08-13-lcm-removal` (md5 of pre-edit
file `6274567b504122a07beb37012090033d`).

```
118c118
<   engine: lcm
---
>   engine: compressor
195d194
<     - hermes-lcm
```

That is the **entire** diff — exactly the two lines the recon predicted. The
`context:` block holds no other key, so there were no further lcm-specific keys
to strip.

Re-read + re-parse proves the merge is clean:

```
context     = {'engine': 'compressor'}
plugins     = {'enabled': ['rtk-rewrite'], 'disabled': []}
compression = {'enabled': True, 'threshold': 0.5, 'protect_last_n': 20}   # untouched, as ordered
top-level keys: 33 before -> 33 after
differing top-level keys: {'plugins', 'context'}
mcp_servers identical: True
"lcm" anywhere in parsed config: False
```

**Permission regression caught and fixed**: tmp+mv created the new file under
the default umask (`0664`); `config.yaml` holds credentials and was `0600`.
Restored with `chmod 600` — verified `-rw------- gaetan gaetan`. Worth
remembering: the repo's tmp+mv convention does not preserve mode, so any
tmp+mv onto a secret-bearing file needs a `chmod --reference` after it.

## 3. Plugin + state moved aside (reversible, nothing deleted)

| Was | Now | Size |
|---|---|---|
| `~/.hermes/plugins/hermes-lcm` | `~/.hermes/plugins-removed-hermes-lcm-2026-08-13` | 25M |
| `~/.hermes/lcm.db` | `~/.hermes/lcm.db-removed-2026-08-13` | 143M |
| `~/.hermes/lcm.db-wal` | `~/.hermes/lcm.db-wal-removed-2026-08-13` | 4.0M |
| `~/.hermes/lcm.db-shm` | `~/.hermes/lcm.db-shm-removed-2026-08-13` | 32K |
| `~/.hermes/lcm-large-outputs` | `~/.hermes/lcm-large-outputs-removed-2026-08-13` | 4.4M |
| `~/.hermes/skills/hermes-lcm` (symlink) | `~/.hermes/skills-removed-hermes-lcm-symlink-2026-08-13` | — |

**Correction to the recon**: it predicted the skill symlink would
"auto-vanish when the plugin dir is removed". It does not — a symlink outlives
its target and becomes dangling, which would have left a broken entry inside
the skills tree. It was moved aside as well. Verified after:
`find ~/.hermes/skills -xtype l` → empty, and `find -L ~/.hermes/skills -name
SKILL.md | wc -l` → 130 with no traversal error.

Compressor needs no migration — it starts fresh per session and never read
`lcm.db`. The 147M of lcm state is kept only so the removal is reversible and
so past runs stay forensically readable.

## 4. Stale references

- `~/.hermes/SOUL.md` line 261, rule 10 — the `lcm_grep` tool disappears with
  the plugin. Minimal edit, `.bak` first
  (`~/.hermes/SOUL.md.bak-2026-08-13-lcm-removal`), flock + tmp+mv on the live
  file, mode restored via `chmod --reference`:

```
261c261
<     own, the session that performed it (`session_search`, `lcm_grep`) — the
---
>     own, the session that performed it (`session_search`) — the
```

  Rule 10's substance is unchanged: still "fetch the session that performed it
  before narrating a past act", just without the tool that no longer exists.
  Live file and repo mirror were byte-identical before the edit
  (`b5a2df9f132d257dfbec020917ccac89`, 317 lines) and are byte-identical after
  (`f808270f95153c405025af2f81d1c681`, 317 lines) — so no Tars-side drift was
  clobbered.

- **No skill needed a change**, so no skill version bumps. `grep -rn lcm` over
  `~/.hermes/skills/` returns only generated caches
  (`.hub/index-cache/hermes-index.json`, `.usage.json` stat block) and an
  unrelated substring inside vendored OOXML `.xsd` files — none are functional
  references.
- **No cron job references lcm** — full substring scan across all 25 jobs:
  `"lcm" in json.dumps(jobs).lower()` → `False`. Tonight's job included.
- Repo `status/`, `PLAN.md`, `PITCH.md` hits are historical evidence of past
  state (WF4/B5 probe records), not live references — left as written.

## 5. Restart

Required: plugin discovery is a process-wide singleton
(`hermes_cli/plugins.py: discover_and_load(force=False)`, line 1316 —
`if self._discovered and not force: return`), so a running gateway never
rescans disk. Config live-reload (~30 s) covers config only, never the cached
plugin object. Verified in-tree on the VM.

Quiet window checked on **both** axes before restarting (a restart earlier
today cut a run mid-task):

- inbound silence: last real inbound `15:12:39` UTC, answered `15:13:10`;
  everything after it is `Early reject of unauthorized user` (not Gaetan)
- no in-flight agent run: zero processing/turn/tool markers in the last 200
  lines of `agent.log`; last cron activity was job `3eb53a322510` completing
  at `15:18:54`; next scheduled cron ≈ `16:18` UTC

```
ActiveEnterTimestamp BEFORE : Thu 2026-08-13 14:46:45 UTC   MainPID=1562573
restart issued              : 2026-08-13 15:24:10 UTC
ActiveEnterTimestamp AFTER  : Thu 2026-08-13 15:24:11 UTC   MainPID=1578393
ActiveState=active   NRestarts=0
```

The unit sent its usual shutdown notification to the home DM
(`15:24:10,487 gateway.run: Sent shutdown notification to home channel
slack:D0BBYNM01BL`) — expected behaviour of the restart, not a message from
this session. Drain was clean: `drain done at +0.21s (drain took 0.00s,
timed_out=False)` — nothing was cut mid-run.

## 6. Verification (no Slack traffic)

| Check | Result |
|---|---|
| `hermes config get context.engine` | `compressor` ✓ |
| `hermes plugins list \| grep -ci lcm` | `0` ✓ |
| `rtk-rewrite` still loaded | `enabled 0.1.0 user` ✓ |
| lcm mentions in `agent.log` since restart | `0` ✓ (pre-change runs logged 4 lines: "Plugin 'hermes-lcm' registered context engine: lcm", "LCM plugin loaded — lossless context management active", …) |
| lcm/plugin errors in `errors.log` since restart | none ✓ |
| Plugin discovery count | `56 found, 49 enabled` — was `57 found, 50 enabled` at 15:18:35, i.e. exactly one plugin gone ✓ |
| 23:55 job `982c896036bc` | `[active]`, `state=scheduled`, next run `2026-08-13T23:55:00+02:00`, `deliver=slack`, prompt `len=1032 sha=07107ada3e58068c` — **byte-identical to before the work** ✓ |
| `daily-work-brief` `e231e5faf180` | active, next `2026-08-14T08:30+02:00` ✓ |
| `engagement-checker` `62e8cd9db637` / final pass `759e08c598e3` | active, next `2026-08-14T10:00` / `17:00` +02:00 ✓ |
| Enabled job count | 11 before, 11 after ✓ |

Note on log evidence: there is no positive "using compressor" startup line —
`agent_init.py` only logs `Using context engine: %s` on the non-default branch;
the built-in `ContextCompressor` path is silent by design. The proof is the
three-way agreement of `config get` → `compressor`, zero lcm load attempts, and
the plugin count dropping by exactly one.

## Rollback (if ever needed)

```
mv ~/.hermes/plugins-removed-hermes-lcm-2026-08-13 ~/.hermes/plugins/hermes-lcm
mv ~/.hermes/skills-removed-hermes-lcm-symlink-2026-08-13 ~/.hermes/skills/hermes-lcm
for f in lcm.db lcm.db-wal lcm.db-shm lcm-large-outputs; do mv ~/.hermes/$f-removed-2026-08-13 ~/.hermes/$f; done
cp -a ~/.hermes/config.yaml.bak-2026-08-13-lcm-removal ~/.hermes/config.yaml   # restores engine: lcm + plugins entry
systemctl --user restart hermes-gateway.service
```

The unpatched resurfacing bug (#485) that motivated the removal comes back with
it — rollback is a diagnostic step, not a destination.
