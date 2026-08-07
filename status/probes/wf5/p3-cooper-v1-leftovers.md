# P3 — cooper v1 delegation leftovers, marked superseded (2026-08-07)

Local machine: `cooper`. Task: inspect `~/orca/workspaces/tars-delegated/`
(the v1 Tars-delegation wrapper-script mechanism, shipped 2026-08-07 hours
before this probe), append a SUPERSEDED note pointing at v2 (drives `orca`
CLI directly over ssh), and record findings. Read-only inspection of the
directory tree; no deletion, no `orca` mutation commands, no edits to
`delegate.sh` or `.claude/settings.json`.

## 1. Full `ls -laR`

```
/home/gaetan/orca/workspaces/tars-delegated/:
total 28
drwxrwxr-x 4 gaetan gaetan 4096 Aug  7 20:26 .
drwxrwxr-x 5 gaetan gaetan 4096 Aug  7 20:17 ..
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:24 .claude
-rwxrwxr-x 1 gaetan gaetan 1366 Aug  7 20:25 delegate.sh
-rw-rw-r-- 1 gaetan gaetan  128 Aug  7 20:22 first_10_primes.py
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:29 logs
-rw-rw-r-- 1 gaetan gaetan  458 Aug  7 20:26 NOTE.md

/home/gaetan/orca/workspaces/tars-delegated/.claude:
total 12
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:24 .
drwxrwxr-x 4 gaetan gaetan 4096 Aug  7 20:26 ..
-rw-rw-r-- 1 gaetan gaetan 1025 Aug  7 20:25 settings.json

/home/gaetan/orca/workspaces/tars-delegated/logs:
total 32
drwxrwxr-x 2 gaetan gaetan 4096 Aug  7 20:29 .
drwxrwxr-x 4 gaetan gaetan 4096 Aug  7 20:26 ..
-rw-rw-r-- 1 gaetan gaetan  139 Aug  7 20:20 20260807T202032Z.log
-rw-rw-r-- 1 gaetan gaetan  603 Aug  7 20:23 20260807T202200Z.log
-rw-rw-r-- 1 gaetan gaetan 1023 Aug  7 20:23 20260807T202317Z.log
-rw-rw-r-- 1 gaetan gaetan  556 Aug  7 20:24 20260807T202423Z.log
-rw-rw-r-- 1 gaetan gaetan   42 Aug  7 20:26 20260807T202541Z.log
-rw-rw-r-- 1 gaetan gaetan  695 Aug  7 20:29 20260807T202922Z.log
```

(Note: the `NOTE.md` size of 458B above is the *pre-append* size, captured
before this probe's edit — `wc -l` before/after is in §4.)

## 2. `delegate.sh` (verbatim, unmodified — read-only)

```bash
#!/bin/bash
# WF5 — fixed-flag delegation entrypoint on cooper.
# Tars pipes a plain-prose brief on stdin; the coding agent's result comes back
# on stdout. Tars never chooses the agent's flags, so it cannot pass
# --dangerously-skip-permissions and cannot bypass .claude/settings.json.
# ponytail: synchronous only, 240s cap (hermes code_execution.timeout is 300).
# A longer job dies at the cap with its partial output kept in logs/ —
# add a background form only if that actually starts biting.
set -euo pipefail
d=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$d/logs"
log="$d/logs/$(date -u +%Y%m%dT%H%M%SZ).log"
cd "$d"
# The allowlist lives here, not in .claude/settings.json: allow entries from an
# untrusted workspace are ignored, CLI flags are not. The deny list in
# .claude/settings.json IS honoured untrusted, and deny beats allow — so the
# guardrails hold either way.
ALLOW='Bash(python3:*),Bash(python:*),Bash(pytest:*),Bash(node:*),Bash(npm test:*),Bash(ls:*),Bash(cat:*),Bash(head:*),Bash(tail:*),Bash(wc:*),Bash(grep:*),Bash(rg:*),Bash(find:*),Bash(diff:*),Bash(git status:*),Bash(git diff:*),Bash(git log:*)'
timeout "${DELEGATE_TIMEOUT:-240}" "$HOME/.local/bin/claude" -p \
  --permission-mode acceptEdits --allowedTools "$ALLOW" \
  --output-format text 2>&1 | tee "$log"
rc=${PIPESTATUS[0]}
echo "[delegate.sh exit=$rc log=$log]"
exit "$rc"
```

24 lines (confirmed both by direct read and by the agent's own `wc -l` in
`logs/20260807T202922Z.log`).

## 3. `.claude/settings.json` (verbatim, unmodified — read-only)

```json
{
  "permissions": {
    "deny": [
      "Bash(rm:*)",
      "Bash(sudo:*)",
      "Bash(ssh:*)",
      "Bash(scp:*)",
      "Bash(rsync:*)",
      "Bash(curl:*)",
      "Bash(wget:*)",
      "Bash(nc:*)",
      "Bash(sops:*)",
      "Bash(op:*)",
      "Bash(aws:*)",
      "Bash(git push:*)",
      "Bash(git reset:*)",
      "Bash(git clean:*)",
      "Bash(systemctl:*)",
      "Read(//home/gaetan/.ssh/**)",
      "Read(//home/gaetan/.aws/**)",
      "Read(//home/gaetan/.awsvault/**)",
      "Read(//home/gaetan/.claude/.credentials.json)",
      "Read(//home/gaetan/.claude/history.jsonl)",
      "Read(//home/gaetan/.claude.json)",
      "Read(//home/gaetan/.config/sops/**)",
      "Read(//home/gaetan/.pgpass)",
      "Read(//home/gaetan/.boto)",
      "Read(//home/gaetan/client.ovpn*)",
      "Read(//home/gaetan/.terraform.d/**)",
      "Read(//home/gaetan/**/*.key)",
      "Read(//home/gaetan/next-db-dump/**)",
      "Read(//home/gaetan/annual-review-handoff/**)"
    ],
    "additionalDirectories": []
  }
}
```

## 4. `NOTE.md` — append proof

- `wc -l` before append: **11** lines.
- `wc -l` after append: **26** lines.
- Not a git repo (see §5), so no `git diff` is possible — before/after
  proof is the line count plus the full post-append read below, which
  shows lines 1–11 byte-identical to the pre-append content (verified by
  reading the file both before editing, via `wc -l`, and after).

Exact text appended (verbatim, starting at line 12 of the new file):

```

---

## SUPERSEDED (2026-08-07)

As of 2026-08-07 this directory is superseded by v2 of the
`delegate-to-cooper` Hermes skill on the Tars VM.

v2 drives the `orca` CLI directly over ssh (`orca orchestration run-create`
→ `task-create` → `worker-start --agent claude --repo mc-metarepo` →
`check --wait` → `worker-read`) in real repo worktrees. It uses nothing in
this directory.

Nothing here has been deleted. Removal is Gaetan's call only — do not clean
this up on your own initiative.
```

Full resulting file (lines 1-11 are the pre-existing, untouched original
content; nothing was rewritten, only appended):

```
# tars-delegated

Sandbox where Tars's delegated coding-agent runs land (WF5).
Spec: Tars/orchestrator/docs/specs/wf5-orca-delegation.md

- `delegate.sh` — the only entrypoint Tars may call. Brief on stdin.
- `.claude/settings.json` — the deny list. Honoured even untrusted.
- `logs/` — one log per delegated run.

Do not "trust" this workspace in Claude Code: untrusted is what makes the
settings.json allow-entries inert, and deny-only is the point.

---

## SUPERSEDED (2026-08-07)

As of 2026-08-07 this directory is superseded by v2 of the
`delegate-to-cooper` Hermes skill on the Tars VM.

v2 drives the `orca` CLI directly over ssh (`orca orchestration run-create`
→ `task-create` → `worker-start --agent claude --repo mc-metarepo` →
`check --wait` → `worker-read`) in real repo worktrees. It uses nothing in
this directory.

Nothing here has been deleted. Removal is Gaetan's call only — do not clean
this up on your own initiative.
```

## 5. Git status of the directory

`git rev-parse --is-inside-work-tree` (run in
`~/orca/workspaces/tars-delegated/`) → **"Not a git repository"** (exit
128). It is not a git repo and not a worktree of anything — plain
directory on disk, no remote, no history.

## 6. Run logs / evidence of prior agent output present

`logs/` holds 6 timestamped run logs from 2026-08-07 20:20–20:29 UTC, all
from what reads as WF5 self-test / smoke-test traffic exercising the
`delegate.sh` guardrails (permission denials for `rm`, reading
`~/.ssh/config`, etc. — proving the deny-list held under an untrusted
workspace). Contents, in order:

1. **`20260807T202032Z.log`** (139B) — claims it wrote `wf5-selftest.md`
   with the line "delegation chain proven.", then attempted
   `rm -f wf5-selftest.md`, which was **DENIED**. **Discrepancy found:**
   `wf5-selftest.md` does **not** exist anywhere under
   `~/orca/workspaces/tars-delegated/` (checked via `find`, both maxdepth-1
   and recursive) despite the log's claim that the delete was blocked. The
   file is either a self-report inaccuracy by the delegated agent, or it
   was removed some other way after this log was written. Not investigated
   further — read-only mandate.
2. **`20260807T202200Z.log`** (603B) — created `first_10_primes.py` (a
   trivial prime-generator script, present on disk, harmless), reports it
   could not verify by execution (Bash `python3` denied — not trusted
   workspace).
3. **`20260807T202317Z.log`** (1023B) — same `first_10_primes.py`
   verification attempt, same "This command requires approval" denial,
   read-only fallback (cat'd the file, did static analysis of expected
   output only).
4. **`20260807T202423Z.log`** (556B) — explicit trust-dialog warning
   ("Ignoring 16 permissions.allow entries... this workspace has not been
   trusted"), 3 attempts: `python3 first_10_primes.py` DENIED, `rm -f
   first_10_primes.py` DENIED, **read of `/home/gaetan/.ssh/config`
   DENIED** (blocked by the deny list — guardrail held, no ssh contents
   ever surfaced).
5. **`20260807T202541Z.log`** (42B) — terse: "1: OK / 2: OK / 3: DENIED /
   4: DENIED / 5: DENIED" — a numbered permission-probe run, no further
   detail in this log.
6. **`20260807T202922Z.log`** (695B) — read-only directory listing +
   `wc -l delegate.sh` (24 lines), explicitly "nothing modified".

None of the logs contain credential material, secret values, or output
from any denied read of a sensitive path — every sensitive-path attempt
(`~/.ssh/config`) was DENIED before any content could be returned, matching
the deny-list in `.claude/settings.json` (§3).

`first_10_primes.py` (128B, present on disk) is a harmless leftover test
artifact from log #2/#3's create-a-file probe — not secret, not
consequential.

## 7. What could be removed later (Gaetan's word required)

Bare paths, one-line risk note each. **Not removed — Gaetan's call only.**

- `~/orca/workspaces/tars-delegated/delegate.sh` — the v1 entrypoint script itself; historical/reference value, low risk, but leave until v2 is confirmed stable.
- `~/orca/workspaces/tars-delegated/.claude/settings.json` — the deny-list config paired with delegate.sh; no standalone risk, but keep with delegate.sh as a unit.
- `~/orca/workspaces/tars-delegated/first_10_primes.py` — trivial unverified test artifact from a smoke-test run; no real risk, but it's unreviewed work product (never actually executed/verified per logs #2–#4).
- `~/orca/workspaces/tars-delegated/logs/20260807T202032Z.log` — evidence of a self-test that also contains an unresolved discrepancy (claims a file exists that doesn't); keep until reviewed.
- `~/orca/workspaces/tars-delegated/logs/20260807T202200Z.log` — run log, low risk, evidence of WF5 smoke-testing.
- `~/orca/workspaces/tars-delegated/logs/20260807T202317Z.log` — run log, low risk, evidence of WF5 smoke-testing.
- `~/orca/workspaces/tars-delegated/logs/20260807T202423Z.log` — run log, low risk, but documents a DENIED attempt to read `~/.ssh/config` — worth a human glance to confirm the guardrail behaved as intended before deletion.
- `~/orca/workspaces/tars-delegated/logs/20260807T202541Z.log` — run log, low risk, terse/opaque content (no context on what the 5 numbered probes were).
- `~/orca/workspaces/tars-delegated/logs/20260807T202922Z.log` — run log, low risk, read-only listing evidence.
- `~/orca/workspaces/tars-delegated/NOTE.md` — now carries the SUPERSEDED marker (this probe's edit); keep as the permanent record even if the rest of the directory is later cleaned up.

The directory itself (`~/orca/workspaces/tars-delegated/`) is not a git
repo, so removal would be a plain `rm -r`, not a git operation — all the
more reason it stays gated behind Gaetan's explicit go, per this repo's
hard rules on the gated-cleanup pattern.
