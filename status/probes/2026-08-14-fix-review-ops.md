# Red-team review — OPS / RUNBOOK SAFETY lens (2026-08-14)

Target: `status/probes/2026-08-14-damien-followup-fix-draft.md` §"Apply runbook".
Mode: adversarial. **Read-only throughout.** No `sops -d`, no wholesale
`config.yaml`/`.env` read, no credential printed, no VM write, no file in this
repo modified except this artifact.

Live checks ran over `ssh gaetan@192.168.0.9` at 2026-08-14 13:46–13:52 UTC
(`stat`, `md5sum`, `ls`, `df`, `find`, `grep`, `sed`, `python3 -c` over a
non-secret JSON cache, `--help`, and one offline `hermes prompt-size --json`).

---

## 0. Live baseline (what the VM actually looks like right now)

```
$ ssh gaetan@192.168.0.9 'date -u; stat -c "%a %U:%G %s %n" ~/.hermes/SOUL.md \
    ~/.hermes/skills/orchestration/engagement-checker/SKILL.md; md5sum <same two>'
Fri Aug 14 01:46:14 PM UTC 2026
664 gaetan:gaetan 19532 /home/gaetan/.hermes/SOUL.md
600 gaetan:gaetan 81873 /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md
0f2d817607f813c363fe9338f17edd81  /home/gaetan/.hermes/SOUL.md
4d73d985c52c5929d6831fe8e889d129  /home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md

$ md5sum SOUL.md skills/orchestration/engagement-checker/SKILL.md   # repo worktree
0f2d817607f813c363fe9338f17edd81  SOUL.md
4d73d985c52c5929d6831fe8e889d129  skills/orchestration/engagement-checker/SKILL.md
```

Re-checked at 13:51:17 UTC — unchanged. **Both pairs currently MATCH**, so the
draft's md5-identity premise holds as of this review (it did NOT hold for
`SKILL.md` at draft time — the draft correctly marked it unverified; it is now
verified, by this probe and by `2026-08-14-skills-mirror-drift.md`).

Filesystem: `/dev/sda1 ext4` for both paths — same device, so `mv` within a
directory is a `rename(2)`, atomic. Directory modes: `~/.hermes` = 700,
`~/.hermes/skills` = 700, `.../engagement-checker` = 775. SOUL.md at **664** is
group/other-readable but shielded by the 700 parent.

Litter already present: **9** `SOUL.md.bak-*` in `~/.hermes/`, **6**
`SKILL.md.bak-*` in the engagement-checker dir. No stale `*.new` files.

---

## 1. Per-command verdicts

### Step 0 — pre-flight md5 gate

| Command | Verdict |
|---|---|
| `cd …/improvements` | OK. Only `cd` in the runbook; steps 1–3 silently assume the same shell and cwd. |
| `git pull --rebase origin main` | OK. 4 untracked probe files present; rebase is unaffected. |
| `md5sum` local ×2 | OK |
| `ssh … 'md5sum' ×2` | **Runs, but the gate is advisory only.** Two separate outputs, no scripted comparison, no non-zero exit on mismatch. The draft's own Risk note #1 calls this "the only thing preventing this apply from silently reverting a `skill_manage` edit" — a human eyeballing four hex strings is not a gate. |

**Worse: the gate is not co-located with the write.** Between step 0 and step 2
the operator hand-applies three hunks to two files. That interval is unbounded.

### Step 1 — mode capture

| Command | Verdict |
|---|---|
| `SOUL_MODE=$(ssh … 'stat -c %a …')` | OK — yields `664`. Local expansion into the later double-quoted ssh string is correct. |
| `SKILL_MODE=$(…)` | OK — yields `600`. |
| `echo "SOUL_MODE=… SKILL_MODE=…"` + "must be non-empty" | **Advisory comment, not a check.** If the ssh fails the variable is empty and the failure only surfaces *after* the `mv`, as `chmod  <path>` → "invalid mode", leaving the file at umask-077's 0600. For SOUL.md that is silently more restrictive than 664, and nothing in step 4 flags it (step 4 prints modes but states no expectation). |
| `wc -l SOUL.md` expect 343 | **Arithmetic verified.** I applied the hunks to a scratch copy: 319 → **343** lines. Correct. |
| `grep -n` ×2, `grep -c` ×1 | OK |

### Step 2 — VM transfer

| Command | Verdict |
|---|---|
| `cp -p … .bak-convstate-$STAMP` ×2 | OK, mode-preserving. `$STAMP` expands locally. Taken outside the lock and in a separate ssh connection from the write — a small but real window. |
| `cat SOUL.md \| ssh … "umask 077; cat > …new && test -s …new && flock … -c 'mv …' && chmod $MODE … && stat"` | **Quoting is correct.** Verified `sh -c "echo ~/.hermes"` → `/home/gaetan/.hermes` on the VM, so the tilde inside `flock … -c '…'` (which runs the string through `sh`) expands. `flock --help` confirms `-c` semantics. `mv` is same-dir/same-fs ⇒ atomic. **The integrity gate is wrong** — see §2.1. |
| second block, SKILL.md | Same shape, same defect. |
| `ssh … 'md5sum' ×2` | OK, but compared by eye against nothing; step 4 repeats it properly. |
| "Reload: none needed" | **CORRECT — now verified live, not just from repo docs.** See §2.3. |
| `hermes chat "Quote rule 10s …"` | **INVALID COMMAND.** See §2.4. |

### Step 3 — repo commit + push

`git add … && git commit … && git pull --rebase origin main && git push origin HEAD:main`
matches this repo's documented push flow (worktree cannot check out `main`). OK
mechanically. **Ordering is backwards** — see §2.6.

### Step 4 — verification

`md5sum` both sides + `diff` both files + `stat` modes. This is the strongest
part of the runbook: the `diff` is byte-level and would catch a truncated
transfer *after the fact*. It compares VM against the **local worktree**, not
against `origin/main`, so a failed/rebased push is not caught here.
No expected values are stated for the `stat` output.

### Rollback line

`flock … -c 'cp -p …bak… ~/.hermes/SOUL.md'` — **non-atomic**, see §2.2.

---

## 2. Findings, with evidence

### 2.1 `test -s` cannot detect a truncated transfer (blocker)

The only integrity check before the atomic `mv` is `test -s file` — "non-empty".
A dropped ssh stream mid-transfer leaves a *partial but non-empty* `.new`, which
passes, and the `mv` then atomically installs a truncated constitution. Because
`SOUL.md` is re-read fresh at every prompt build (§2.3), Tars would immediately
start running on a SOUL that is missing its tail hard rules, with **no warning
anywhere** — `_truncate_content` only fires on the size cap, not on this.

Fix (one line, same ssh round-trip, no new step):

```bash
LOCAL_MD5=$(md5sum < SOUL.md | cut -d' ' -f1)
cat SOUL.md | ssh gaetan@192.168.0.9 "umask 077; cat > ~/.hermes/SOUL.md.new && \
  [ \"\$(md5sum < ~/.hermes/SOUL.md.new | cut -d' ' -f1)\" = '$LOCAL_MD5' ] && \
  flock ~/.hermes/.wf3.lock -c 'mv ~/.hermes/SOUL.md.new ~/.hermes/SOUL.md' && \
  chmod $SOUL_MODE ~/.hermes/SOUL.md && stat -c '%a %n' ~/.hermes/SOUL.md"
```

### 2.2 Rollback is non-atomic while apply is atomic (major)

Apply uses `mv` (rename, atomic). Rollback uses `cp -p src dst`, which
**truncates and rewrites in place**. A prompt build racing the rollback reads a
half-written `SOUL.md`. The lock does not help (§2.5). Rollback must mirror the
apply path:

```bash
ssh gaetan@192.168.0.9 "cp -p ~/.hermes/SOUL.md.bak-convstate-$STAMP ~/.hermes/SOUL.md.rb && \
  flock ~/.hermes/.wf3.lock -c 'mv ~/.hermes/SOUL.md.rb ~/.hermes/SOUL.md'"
```

### 2.3 SOUL reload semantics — VERIFIED LIVE, draft's claim is correct

The draft asserted "read fresh at every system-prompt build" from repo docs. I
confirmed it against the installed source
(`/home/gaetan/.hermes/hermes-agent`, launched by `~/.local/bin/hermes` →
`hermes-agent/venv/bin/python hermes-agent/hermes`):

- `agent/prompt_builder.py:2004` `def load_soul_md(...)` → `:2017` builds
  `get_hermes_home()/"SOUL.md"`, `:2023` `soul_path.read_text(...)` on **every
  call**. `grep -n "lru_cache|_soul_cache|@cache" agent/system_prompt.py` → no
  hits. No module-level cache, no watcher, no restart needed.
- `agent/system_prompt.py:192-197` calls `_r.load_soul_md(_ctx_len)` inside the
  per-build "stable tier" assembly.
- Independent live proof: `hermes prompt-size --json` (offline, no API call)
  reports `stable (identity/guidance/skills) = 35,563 chars` for a **fresh**
  session, of which SOUL.md's 19,385 chars are present whole.

**So: no restart, and `systemctl --user restart` is correctly marked optional.**
The draft's caveat that a *running* Slack thread keeps its original prompt also
holds (stable-tier is built once per session).

### 2.4 The verification probe is a command that does not exist (major)

```
$ hermes chat --help
usage: hermes chat [-h] [-q QUERY] [--image IMAGE] [-m MODEL] ...
  -q QUERY, --query QUERY   Single query (non-interactive mode)
```

`hermes chat` takes **no positional prompt**. The runbook's
`hermes chat "Quote rule 10s conversation-state paragraph verbatim."` will
either be rejected as an unrecognized argument or drop into an interactive TUI
over a non-TTY ssh. An operator seeing that fail could mis-read it as "the
reload did not work".

It is also the wrong tool: it starts a **real agent session on a live production
agent**, with tools, costing tokens, to prove a file read. The right check is
free, offline and already installed:

```bash
ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); \
  ~/.local/bin/hermes prompt-size --json' | head -20
```

`hermes prompt-size --help`: *"Report the fixed prompt budget for a fresh
session… Runs offline (no API call)."* Its `stable` section must grow by ~1,714
chars, and it is the only check that proves SOUL landed **whole** (§2.7).

### 2.5 `flock ~/.hermes/.wf3.lock` is not the lock Tars' own writes take (major)

Grepped the installed Hermes source, excluding vendored `node_modules`:

```
$ grep -rn "fcntl\|flock" --include=*.py agent gateway cron hermes_cli tools | wc -l
134
```

Every one of those 134 is a *different* lock — `gateway/status.py:732`
(`gateway.lock`), `agent/shell_hooks.py:712` (approvals), `agent/credential_pool.py`
(auth store), `gateway/kanban_watchers.py`. **Zero reference to `.wf3.lock`
anywhere in Hermes.** `.wf3.lock` is a WF3-era convention of *ours*; Hermes has
never heard of it.

Specifically for the skill file: `skill_manage` is
`tools/skill_manager_tool.py:1513`, and every write path in it goes through
`atomic_write_text` (`:942, :1006, :1135, :1314`, defined `utils.py:139` —
"temp file + fsync + atomic rename"). `grep -n "flock\|fcntl"
tools/skill_manager_tool.py` → **no hits**.

Consequence: both sides write atomically, so there is no torn file — but
**last-writer-wins, silently**. The `flock` in the runbook creates false
assurance that the apply is serialized against Tars. It is also pointless for
`SOUL.md`, which is not live-reloaded by a config watcher (unlike `config.yaml`
/ `.env`, which is what the CLAUDE.md hard rule was written for) and whose swap
is already atomic.

The real protection is a **re-check of the md5 inside the same ssh command as
the mv**, which also fixes §2.1:

```bash
… "cat > …new && [ \"\$(md5sum < ~/.hermes/…/SKILL.md | cut -d' ' -f1)\" = '$EXPECTED_OLD_MD5' ] \
     && [ \"\$(md5sum < …new | cut -d' ' -f1)\" = '$NEW_MD5' ] && mv …"
```

Fairness note: the draft's own Risk #1 and
`2026-08-14-skills-mirror-drift.md:34` cite `SKILL.md.bak-tars-selfedit` as
evidence Tars self-edits mirrored skills. That backup exists — but only on
**`delegate-to-cooper`**. All six backups on `engagement-checker` are ours
(`-wf6-*`, `-gcn30-32-*`, `-20260813`). So there is no direct evidence Tars has
ever self-edited *this* file; the race is a live mechanism, not an observed
event.

### 2.6 VM is written before the repo is pushed (major)

Step 2 (VM) precedes step 3 (commit + `git pull --rebase` + push). If the rebase
conflicts — plausible, this repo takes concurrent pushes from peer sessions —
the resolved content can differ from what is already installed on the VM. Step 4
would catch it (it diffs VM against the local worktree), but only after the
production agent has been running the un-landed version. Commit and push first,
then transfer from the pushed tree; the mirror is then the durable record of
exactly what was installed.

### 2.7 The patch removes all headroom against the SOUL truncation floor (major)

Measured, not estimated. Applying the three hunks to a scratch copy:

```
orig  319 lines  19,385 chars
new   343 lines  21,099 chars   (delta +1,714)
```

(`wc -c -m ~/.hermes/SOUL.md` → 19,385 chars / 19,532 bytes — 147 multibyte
chars, the French accents. `.strip()` in `load_soul_md` removes 1 more.)

SOUL.md is passed through `_truncate_content` (`agent/prompt_builder.py:2025`),
capped by `_get_context_file_max_chars` (`:1312`):

1. `context_file_max_chars` in config.yaml — **not set** (`grep -n
   "context_file_max_chars" ~/.hermes/config.yaml` → no match; key name only, no
   value read).
2. dynamic: `context_length * 4 * 0.06`, floored at 20,000, ceiled at 500,000
   (`:1292-1294, :1297-1309`).
3. fallback `CONTEXT_FILE_MAX_CHARS = 20_000` (`:1282`) whenever `_ctx_len` is
   `None` — and `_ctx_len` is `None` unless `agent.context_compressor` exists
   *and* exposes a positive int `context_length` (`agent/system_prompt.py:179-184`).

Today: `~/.hermes/context_length_cache.yaml` → `gpt-5.6-sol…: 272000`, so the
cap is `272000*4*0.06 = 65,280`. **21,099 < 65,280 — truncation does not fire
today.** Confirmed: `hermes prompt-size --json` shows SOUL included whole, and
`grep -rhoE "Context file .* TRUNCATED" ~/.hermes/logs/` returns nothing, ever.

But the patch takes SOUL from **615 chars under** the 20,000 floor to **1,099
chars over** it. Any of: a model swap to a ≤83,333-token window
(20000 / (4*0.06)), a run where the compressor is not attached, or an explicit
low `context_file_max_chars`, and `_truncate_content` keeps head 14,000 + tail
4,000 and drops chars **14,000–17,099** with only a `logger.warning`. I checked
what is in that window on the patched file:

```
dropped-middle starts: "…dy seen or asked for; when the text is mine and he has not seen
   it, I show him the final text and send after his go…"
dropped-middle ends:   "…ents.** Whenever I report on a ticket, a task or a status that
   involves a named person other than Gaetan…"
'Conversation state, not just events'      -> survives truncation? False
'A boolean I invented is not a state I read' -> survives truncation? True
```

i.e. under the floor **the fix deletes itself**, along with the approval rules
around it, while H1 survives — the worst possible partial state. This repo has
had model/engine churn as recently as this week (GCN-49 cutover, compressor
restore, context-max banner), so it is not hypothetical.

The runbook has **no size guard and no post-apply truncation check**. Both are
one line: assert `wc -m SOUL.md` and add `hermes prompt-size --json` to step 4.

### 2.8 Ephemeral variables make the rollback unusable from a fresh shell (minor)

`$STAMP`, `$SOUL_MODE`, `$SKILL_MODE` live only in the applying shell. The
rollback recipe at the bottom of the runbook interpolates `$STAMP`; run from any
other shell it becomes `…SOUL.md.bak-convstate-` and fails (safely, but
uselessly). The actual stamp is never written down. Record the resolved `.bak`
paths and modes in the status file at apply time.

### 2.9 No cleanup, and orphan `.new` on failure (minor)

`~/.hermes` already carries 9 `SOUL.md.bak-*`; the skill dir 6 `SKILL.md.bak-*`.
Two more are added with no prune. And if `cat SOUL.md` fails locally (wrong cwd
— steps 1–3 never re-`cd`), ssh gets empty stdin, `test -s` correctly aborts,
and a zero-byte `~/.hermes/SOUL.md.new` is left behind.

### 2.10 No `hermes pause`, and a cron fires every 30 minutes (minor)

```
$ hermes cron list
  62e8cd9db637 [active] Gaetan engagement checker
    Schedule: */30 10-16 * * 1-5   Next run: 2026-08-14T16:00:00+02:00
    Skills: engagement-checker      Last run: 2026-08-14T15:36:57 ok
  759e08c598e3 [active] Gaetan engagement checker final pass
    Schedule: 0 17 * * 1-5          Next run: 2026-08-14T17:00:00+02:00
```

The exact skill being edited is loaded by a cron every 30 minutes on weekday
afternoons. The two transfers in step 2 are separate ssh calls, so a run that
starts between them sees **new SOUL + old skill**. Not corrupting (an in-flight
read holds the old inode across the rename), but it is one turn under a mixed
contract. `hermes pause --reason …` / `hermes resume` exist and are the
sanctioned lever ("Halts NEW work only — cron dispatch, kanban dispatch, and new
gateway turns… In-flight work is never killed"). The runbook uses neither.

### 2.11 `30d` is valid — but H2's widening rule has no bound (minor)

Verified from the live schema cache, not from a doc. `python3` over
`~/.hermes/cache/mcp_schema_cache.json`, `conversations_history`:

```
REQUIRED: ['channel_id']
--- limit  string | default= 1d
    Limit of messages to fetch in format of maximum ranges of time (e.g. 1d - 1 day,
    1w - 1 week, 30d - 30 days, 90d - 90 days which is a default limit for free tier
    history) or number of messages (e.g. 50). Must be empty when 'cursor' is provided.
--- cursor string | default= None
--- include_activity_messages boolean | default= False
```

`30d` is explicitly enumerated ⇒ **H2's format is correct**, and the `1d`
default is confirmed as a relative time window (the latent defect the incident
exposed). Two residual gaps in H2's wording:

- *"widening until I have both"* has **no termination bound**. `90d` is named as
  the free-tier history cap; past it, widening returns the same result forever.
  H2 should stop at `90d` and fall through to its own say-so-if-unreachable
  clause.
- The schema says `limit` **must be empty when `cursor` is provided**. H2's
  unconditional "I pass an explicit `limit`" contradicts the paginated path.

---

## 3. What the runbook gets right

- Backups with `cp -p` before touching anything, distinct stamped names.
- `umask 077` on the incoming temp file, mode captured *before* the write and
  restored *after* the `mv` — correctly encodes the 2026-08-13 tmp+mv mode-drop
  lesson, and the captured values (664 / 600) are right.
- `mv` within the same directory on one ext4 device — genuinely atomic.
- Byte-level `diff` in step 4, not hashes alone.
- No `sops -d` anywhere, no secret on argv, no config/.env touched — clean
  against every hard rule in CLAUDE.md.
- "Reload: none needed" is **correct**, and I could not refute it.
- The `wc -l` expectation (343) is arithmetically exact.

## 4. Minimum changes before this is safe to run

1. Replace `test -s` with an md5 equality check against the local file, inside
   the same ssh command as the `mv`. (§2.1)
2. Move the drift gate into that same command — assert the *old* md5 on the VM
   immediately before the `mv`, not minutes earlier in step 0. (§2.5)
3. Make the rollback use tmp+mv, not `cp` in place. (§2.2)
4. Replace `hermes chat "…"` with `hermes prompt-size --json`, and add it to
   step 4 as the proof SOUL landed whole. (§2.4, §2.7)
5. Assert `wc -m SOUL.md` < the effective cap, and grep the logs for
   `Context file SOUL.md TRUNCATED` after the first cron run. (§2.7)
6. Commit and push before transferring. (§2.6)
7. `hermes pause` before step 2, `hermes resume` after step 4. (§2.10)
8. Bound H2's widening at `90d` and exempt the cursor path. (§2.11)
