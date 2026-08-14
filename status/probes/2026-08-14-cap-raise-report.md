# Cap raise applied — Option A (`context_file_max_chars: 40000`), 2026-08-14

**Status: DONE.** All four gates passed. One line added to `~/.hermes/config.yaml`
on the VM (`gaetan@192.168.0.9`). No SOUL/skill edit here — this only unblocks the
415-char budget in `2026-08-14-damien-followup-fix-draft-v2.md` §3/§7 (new cap
39,800 usable per that runbook's `CAP` variable).

Secret hygiene: the file was never `cat`/`less`/`sops -d`'d. Everything below is
either a key name, a size, a mode, an md5, or the single line I added. No other
value from `config.yaml` was printed at any point.

---

## Mechanism, re-read from the installed source (not remembered)

`~/.hermes/hermes-agent/agent/prompt_builder.py`

- `_get_context_file_max_chars` (`:1312`) — resolution order confirmed verbatim:
  1. explicit **top-level** `context_file_max_chars` in `config.yaml`, "user knows
     best, always wins (including over the dynamic cap)";
  2. `_dynamic_context_file_max_chars` = `context_length × 4 × 0.06`, floor 20,000,
     ceiling 500,000;
  3. flat `CONTEXT_FILE_MAX_CHARS = 20_000` (`:1280`).
- Read path is `load_config_readonly().get("context_file_max_chars")` — a **root-level
  scalar**, *not* a member of the existing `context:` mapping. That is where the line
  went.
- Context-file set = `build_context_files_prompt` (`:2132`): SOUL.md from HERMES_HOME
  (always) **plus at most one** project file discovered from the configured cwd, first
  match wins: `.hermes.md`/`HERMES.md` (walk to git root) → `AGENTS.md`/`agents.md` →
  `CLAUDE.md`/`claude.md` → `.cursorrules` + `.cursor/rules/*.mdc`. cwd comes from
  `terminal.cwd` via `resolve_context_cwd()` (`agent/runtime_cwd.py:76`).

## Gate 1 — PRECONDITION: every configured context file < 40,000 ✅

Enumerated by replicating that discovery against the live `terminal.cwd`:

| Context file | chars (`wc -m`) | < 40,000 |
|---|---|---|
| `~/.hermes/SOUL.md` | **19,385** | yes |
| project context file | **none** | — |

`terminal.cwd` resolves to an existing directory that contains **none** of
`.hermes.md`, `HERMES.md`, `AGENTS.md`, `agents.md`, `CLAUDE.md`, `claude.md`,
`.cursorrules`, `.cursor/rules/` (each checked individually; path itself not printed).
So exactly **one** context file is configured. Largest = 19,385; SOUL patched with the
full v2 contract = 20,751 (draft §3) — both well under 40,000.

## Gate 2 — baseline ✅

| | value |
|---|---|
| mode before | `600` (as expected) |
| size before | 8,083 bytes |
| **md5 before** | `b6212f80bcee0d3a1f23877613c897ca` |
| backup | `~/.hermes/config.yaml.bak-capraise-20260814T145644Z` (`cp -p`, mode 600, md5 identical to baseline — verified) |
| key already present? | **no** — `grep -n 'context_file_max_chars' config.yaml` → no match anywhere (so: insert, not update-in-place) |

## Gate 3 — merged edit under flock ✅

Line added, at **column 0 (root level)**, appended as the file's last line:

```yaml
context_file_max_chars: 40000
```

Procedure actually run:

1. `python3` built `config.yaml.new` in place of a hand edit, with three asserts that
   abort before writing anything:
   - `"context_file_max_chars" not in src` (no duplicate/shadowed key);
   - **`yaml.safe_load(new) == {**yaml.safe_load(old), "context_file_max_chars": 40000}`**
     — this is the real merge gate: the parsed config must be the old one *plus exactly
     this key*, nothing dropped, nothing re-nested (result: `PARSE_EQUAL_PLUS_KEY ok`);
   - byte delta == length of the added line (**+30 bytes**, i.e. the file already ended
     with a newline).
2. `umask 077` + `chmod 600` on the temp file.
3. Baseline md5 re-asserted immediately before the swap (drift guard), then
   `flock ~/.hermes/.wf3.lock -c 'mv ~/.hermes/config.yaml.new ~/.hermes/config.yaml'`.
4. `chmod 600` after the `mv` (tmp+mv drops the mode — incident-logged 2026-08-13),
   then `stat` re-verified.

## Gate 4 — post-checks ✅

| Check | Result |
|---|---|
| `grep -c '^context_file_max_chars: 40000$'` | **1** |
| mode after | **600** |
| size after | 8,113 bytes (+30) |
| **md5 after** | `6efb21ef424d9e35af19ee29ae518365` — differs from baseline ✅ |
| edit completed at | `2026-08-14T14:57:08Z` |
| **effective cap** (installed `_get_context_file_max_chars`, called after the edit) | `f()` → **40000**, `f(200000)` → **40000**, `f(32000)` → **40000** — pinned, confirming the "always wins over the dynamic cap" cost |
| `hermes-gateway.service` (user unit) | `active/running`, **`NRestarts=0`**, `ExecMainStartTimestamp = 13:02:18 UTC` — unchanged across the edit (live reload, no restart, no crash) |
| `hermes-cloakbrowser.service` | `active/running` |
| journald `--user -u hermes-gateway --since 14:57:00` | **no entries** — zero new WARNING/error after the edit (checked at 14:59:00Z, i.e. ≈2 min > the ~30 s reload window) |
| `~/.hermes/logs/errors.log` in the 14:5x window | 2 lines, both **before** the edit (14:52:50, 14:53:42) and both the pre-existing `[Slack] Early reject of unauthorized user` WARNING. Nothing after 14:57. |
| `grep -rl TRUNCATED ~/.hermes/logs/` | `NO_TRUNCATION_MARKERS` |
| `hermes prompt-size --json` (offline, no API call) | system_prompt 52,069 chars / skills_index 12,636 / memory 733 / user_profile 1,501 / tools 41 (71,467 json bytes). It does **not** expose the cap itself, so it is corroboration only — the direct resolver call above is the evidence for the cap. |

**Journald verdict: clean.** No new warnings or errors attributable to the edit; the
gateway kept the same main PID start time, so it live-reloaded rather than restarted.

---

## Rollback recipe (not used — documented)

```bash
ssh gaetan@192.168.0.9 '
  cp -p ~/.hermes/config.yaml.bak-capraise-20260814T145644Z ~/.hermes/config.yaml.rb && \
  flock ~/.hermes/.wf3.lock -c "mv ~/.hermes/config.yaml.rb ~/.hermes/config.yaml" && \
  chmod 600 ~/.hermes/config.yaml && \
  [ "$(md5sum < ~/.hermes/config.yaml | cut -d" " -f1)" = b6212f80bcee0d3a1f23877613c897ca ] && \
  stat -c "%a %n" ~/.hermes/config.yaml'
```

Restores md5 `b6212f80bcee0d3a1f23877613c897ca`, mode 600. The gateway live-reloads
within ~30 s; no restart needed.

## Known consequence (already priced in by Option A)

The cap is now **pinned at 40,000 for every context file**, replacing today's dynamic
65,280. Harmless at current sizes (largest context file 19,385; 20,751 patched), and it
is exactly the trade the draft named. If a context file ever needs > 40,000, this line
is the single place to raise.

## Not verified

- Behaviour of the v2 SOUL contract itself — nothing in SOUL or the skill was touched
  here; the draft's runbook §9 is still un-run (its `CAP` may now be set to 39,800).
- No live Hermes turn was exercised post-edit (no `hermes chat` — that would start a
  paid session). Health evidence is unit state + journald + logs only.
