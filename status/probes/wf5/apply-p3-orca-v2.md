# P3 — apply Orca v2 delegation skill + first live E2E

Peer session: `apply-p3-orca` (worktree
`/home/gaetan/orca/workspaces/Tars/apply-p3-orca`, branch
`GaetanCathelain/apply-p3-orca`). Mandate: `docs/plans/apply-P3-orca-v2.md`.
Hub: the Tars orchestrator session. Sibling peer applied P4 (SOUL) concurrently
under a joint-apply gate.

All times UTC (the VM clock is UTC).

---

## 1. Backup — taken FIRST, before anything was touched

The v1 `SKILL.md` existed **only on the VM, never in git**. It is the sole
rollback path, so it was copied before any other action.

```
$ ssh gaetan@192.168.0.9 'cp -p ~/.hermes/skills/delegate-to-cooper/SKILL.md \
                             ~/.hermes/skills/delegate-to-cooper/SKILL.md.bak-p3'
```

| | |
|---|---|
| Backup path | `~/.hermes/skills/delegate-to-cooper/SKILL.md.bak-p3` (on the VM, 192.168.0.9) |
| v1 md5 | `d61888ecb34a2342663a7122414aecdb` |
| v1 size / lines | 3304 bytes / 77 lines |
| Mode / owner | `-rw-rw-r-- gaetan gaetan`, preserved by `cp -p` |
| `diff SKILL.md SKILL.md.bak-p3` | empty — **IDENTICAL** |

The write refused to proceed if `.bak-p3` already existed (guard in the command),
so an existing backup could not be clobbered by a re-run.

**Index-pollution check (not in the mandate — verified anyway).** A stray file in
a skills directory could in principle be picked up as a second skill. It cannot
be: every scanner in the Hermes source matches the filename exactly —
`agent/learning_graph.py:80`, `agent/curator_backup.py:185`,
`tools/skill_usage.py:359,551,1235,1318`, `tools/skills_sync_client.py:533,1713`,
`tools/skills_hub.py:3376,3391,3701` all use `rglob("SKILL.md")`, and
`scripts/build_skills_index.py:178` tests `path.endswith("/SKILL.md")`.
`SKILL.md.bak-p3` matches none of them. The backup is safe where it sits.

**Second rollback copy (beyond the mandate).** Because "only on the VM" is a
single point of failure, the v1 file was also pulled into git, byte-identical:

```
artifacts/delegate-to-cooper-SKILL-v1-live.md   md5 d61888ecb34a2342663a7122414aecdb
```

Rollback is therefore possible from either the VM `.bak-p3` or from git.

### Rollback command (if ever needed)

```bash
ssh gaetan@192.168.0.9 'D=~/.hermes/skills/delegate-to-cooper
flock ~/.hermes/.wf3.lock -c "cp -p $D/SKILL.md.bak-p3 $D/.SKILL.md.tmp && mv $D/.SKILL.md.tmp $D/SKILL.md"'
# then verify: md5sum must be d61888ecb34a2342663a7122414aecdb
```

---

## 2. What was replaced, and the contradiction that forced the joint gate

v1 `SKILL.md:47`, verbatim:

> `- Never `rm`, `rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper.`

This is what P4's new SOUL rule 7 contradicts: SOUL would grant `sudo` on cooper
while a loaded skill forbade it. Confirmed live before the apply — the gate was
justified, not ceremonial.

v1 also confined Tars to `~/orca/workspaces/tars-delegated/` (`SKILL.md:54-56`)
and to a four-command allowlist (`:33-40`). v2 drops all of it.

---

## 3. Staging + rehearsal (before the gate)

v2 source: `artifacts/delegate-to-cooper-SKILL.md`, 437 lines,
md5 `929b23de94fd0697bf3ae2709f2dca42`.

Staged to the VM over stdin (never through argv or a heredoc, so nothing in the
file — it contains `$(cat …)` and a `BRIEF` heredoc — could be shell-expanded):

```
$ ssh gaetan@192.168.0.9 'cat > /tmp/p3-skill.new' < artifacts/delegate-to-cooper-SKILL.md
$ ssh gaetan@192.168.0.9 'md5sum /tmp/p3-skill.new'
929b23de94fd0697bf3ae2709f2dca42  /tmp/p3-skill.new    # matches source exactly
```

The apply command was then **rehearsed on a scratch directory** so that the GO
would be instant and already proven:

```
before: -rw-rw-r-- gaetan gaetan 3304  md5 d61888ecb34a2342663a7122414aecdb
after:  -rw-rw-r-- gaetan gaetan 25050 md5 929b23de94fd0697bf3ae2709f2dca42
=== REHEARSAL OK, scratch removed ===
```

Mode and owner preserved; the rename is atomic; the content is verified *before*
it goes live. Scratch dir removed.

---

## 4. Baseline before the apply

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes skills list'
7 hub-installed, 66 builtin, 7 local — 80 enabled, 0 disabled
```

**Spec correction.** `docs/specs/wf5-orca-delegation.md` (v2 section, §Superseded
v1 artifacts) predicts *"Skill count should stay 77 / 6-local"*. That number is
**stale** — measured baseline on 2026-08-07 is **80 enabled / 7 local / 66
builtin / 7 hub-installed**. The prediction's substance still holds: this is a
content replacement, not an add, so the count must not move.

The count therefore cannot by itself prove the new file parsed. The signal that
does: v1 frontmatter carries `category: delegate-to-cooper`, v2 carries
`category: orchestration`. A category move proves Hermes re-read the file.

---

## 5. E2E preflight (read-only, run before the gate)

| Check | Result |
|---|---|
| `ssh cooper` from the VM | OK — user `gaetan`, `192.168.0.4`, no prompt |
| `orca` on non-interactive PATH | `/usr/local/bin/orca` — confirms the spec's correction |
| `orca status --json` | `ok:true`, `runtime.reachable: **true**`, app running pid 218650 |
| `orca repo list --json` | `8099e312-3232-46f2-83a9-97aeaf5de5a2` → `/home/gaetan/dev/mc-metarepo` — **matches the id hardcoded in the skill** |
| | `a71db494-6e32-437d-be39-acff38061723` → `/home/gaetan/dev/Tars` |
| `orca account list --json` | claude: 2 accounts, active `009bfa67-…` (`gaetan.cathelain-claude02@mobile.club`, subscription-oauth) |

**Correction to a spec claim.** The v2 spec says *"Only `claude` and `codex` are
verified-usable `--agent` values on cooper."* Measured: `result.codex.accounts`
is **empty** and `activeAccountId` is `null`. `--agent codex` has no account to
run under right now — `claude` is the only presently usable value.

---

## 6. Joint-apply gate

Everything above was completed with **nothing written to `~/.hermes/`**.
`P3 READY` was sent to the hub, reporting the backup, the staged file, the
rehearsal, the baseline, and the confirmed v1/SOUL contradiction.

Not touched, per mandate: `~/.hermes/SOUL.md` (sibling peer),
`~/.hermes/config.yaml`, `status/lane-a.md` (hub single-writer). No gateway
restart at any point.

<!-- SECTIONS 7+ (apply, verification, E2E) APPENDED AFTER GO -->
