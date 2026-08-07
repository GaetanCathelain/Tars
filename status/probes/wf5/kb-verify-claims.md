# Adversarial verification — P6 knowledge-bases proposal (claim-by-claim)

Verifier pass, 2026-08-08 UTC. Target: `docs/proposals/P6-knowledge-bases.md`
(191 lines). Evidence checked: `status/probes/wf5/kb-{retrieval-trials,
hermes-retrieval,mc-metarepo,refresh,push-design,push-inputs,map-draft-SKILL}.md`,
`docs/facts.md`, `docs/specs/{tars-profile,wf4-probes}.md`, `CLAUDE.md`, plus
**read-only** re-checks on the VM (`ssh gaetan@192.168.0.9`): `hermes cron
create --help`, `hermes skills list`, `hermes memory status`, `ls
~/.hermes/scripts`, `git -C ~/dev/mc-metarepo config -f .gitmodules
--get-regexp url`, `grep -in metarepo ~/.hermes/SOUL.md`, `wc -lc
~/.hermes/SOUL.md`, `ls ~/dev/mc-metarepo/knowledge/*.md | wc -l`.
No writes anywhere. The proposal was **not** edited.

**Score: 11 findings — 4 CONTRADICTED, 7 UNSUPPORTED.** Everything else spot-
checked came back SUPPORTED (§B).

---

## A. CONTRADICTED / UNSUPPORTED

### 1. `SOUL.md:54` — CONTRADICTED (most serious)

> L158-159: "(2) SOUL line first: drop the `gaetan-metarepo` pointer at
> `SOUL.md:54`, name `~/dev/mc-metarepo`"
> L31-32: "Tars' only KB pointer is `~/dev/gaetan-metarepo` at `SOUL.md:54`"
> L177 (P6-a): "Dead `SOUL.md:54` pointer"

**Evidence.** The line number came from `kb-retrieval-trials.md:458-459`
(`grep -in metarepo ~/.hermes/SOUL.md` → `54:source of truth is
~/dev/gaetan-metarepo on cooper`), measured 22:45–23:15 UTC. SOUL.md has been
rewritten since — measured now:

```
$ grep -in "metarepo" ~/.hermes/SOUL.md
81:source of truth is `~/dev/gaetan-metarepo` on cooper; the live mechanism on my
$ sed -n '54p' ~/.hermes/SOUL.md
4. I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — …
$ wc -lc ~/.hermes/SOUL.md ; stat -c %y ~/.hermes/SOUL.md
93 5272 /home/gaetan/.hermes/SOUL.md
2026-08-07 23:10:42 +0000        # was 1472 B @ 19:39 per live-config-snapshot.md:393
```

The pointer is at **line 81**. Line 54 is now the hard rule "I answer Gaetan
and no one else" — an apply step that edits `SOUL.md:54` would delete the
allowlist backstop rule and leave the dead pointer in place.

**Correction.** Replace every `SOUL.md:54` with "the `gaetan-metarepo`
sentence in SOUL.md (line 81 as of 2026-08-07 23:10 UTC; **locate it with
`grep -n metarepo ~/.hermes/SOUL.md` at apply time — a peer is rewriting this
file and the line number moves**)". Add the `.bak` + `flock` note to that step
specifically, and re-read the file before editing.

### 2. "submodules … (5 SSH URLs)" — CONTRADICTED

> L144: "and submodules are never refreshed (5 SSH URLs) — P6-c"

**Evidence.** Inherited from `kb-refresh.md:36` ("5 submodules on
`git@github.com:...` SSH URLs"), which is itself wrong. Measured now:

```
$ git -C ~/dev/mc-metarepo config -f .gitmodules --get-regexp 'submodule\..*\.url'
… 10 entries, ALL git@github.com: (monorepo-loop, vecna-api, vecna-front,
cleaq-api, cleaq-front, return-app, wordpress, support-engineer, risk,
pim-migration-sync)
```

`kb-mc-metarepo.md:34,70,339` also says **10 registered submodules**. The
proposal contradicts itself: L49-50 says "The 10 submodules", L144 says 5.

**Correction.** L144: "(10 submodules, all on SSH URLs)". Also correct the
source probe `kb-refresh.md` §Summary and §4 if it is ever reused.

### 3. P6-b "hindsight contradicts the brief" — CONTRADICTED (a manufactured open decision)

> L178: "`memory.provider: hindsight` is **actively set** in the live
> `config.yaml` …, contradicting 'Hindsight skipped for v1, no keys' →
> **Reconcile before applying.**"

**Evidence.** This is the designed state, not a discrepancy.
`docs/specs/tars-profile.md:87` specifies `provider: hindsight` in the
profile; `docs/specs/wf4-probes.md:236` defines the acceptance probe as
"*`hermes memory status` → provider `hindsight`, status `not available`*"
because "both HINDSIGHT_* keys deliberately absent". Measured live:

```
$ hermes memory status
  Provider:  hindsight
  Plugin:    installed ✓
  Status:    not available ✗
  Missing:  ✗ HINDSIGHT_API_KEY  ✗ HINDSIGHT_LLM_API_KEY
```

Exactly what the spec predicts. Nothing to reconcile, and nothing about P6
depends on it. It also sits oddly next to the proposal's own L115 ("Hindsight
is **off**"), which is the correct reading.

**Correction.** Delete row P6-b from the Open-decisions table and the
`P6-b hindsight: [ ] off/unset · [ ] keep as configured` item from §Decision.
If anything is kept, demote it to one line under §risk: "provider is set,
keys absent, provider reports `not available` — as specified in
`wf4-probes.md` §11; no action for P6."

### 4. `ssh gaetan@192.168.0.9 && mkdir -p …` — CONTRADICTED (the command does not do what the block implies)

> L78: "`ssh gaetan@192.168.0.9 && mkdir -p ~/.hermes/skills/mc-metarepo ~/.hermes/scripts`"

**Evidence.** `&&` sequences two *local* commands: this opens an interactive
ssh session, and only after it exits (status 0) does `mkdir -p` run — **on
cooper**, creating `~/.hermes/skills/mc-metarepo` and `~/.hermes/scripts` on
the wrong machine. The rest of the block (`cat > ~/.hermes/scripts/…`,
`chmod`, `hermes cron create`) is written as if it continued inside the ssh
session; it does not. Nobody ran any of this — `kb-refresh.md` §1 and
`kb-retrieval-trials.md` both state no writes were made under `~/.hermes/`.
Re-verified: `ls ~/.hermes/scripts` → "No such file or directory".

**Correction.** Make it one remote invocation and mark the block as
**proposed, never executed**, e.g.
`ssh gaetan@192.168.0.9 'mkdir -p ~/.hermes/skills/mc-metarepo ~/.hermes/scripts'`
followed by the heredoc piped over ssh (or run the whole block from an
interactive session on the VM, stated as such).

### 5. The `hermes cron create` line quoted as a working recipe — UNSUPPORTED

> L101: "`~/.local/bin/hermes cron create "every 1h" --no-agent --script
> kb-refresh.sh --deliver slack --name "mc-metarepo-refresh"`"

**Evidence.** `kb-refresh.md:85-87` is explicit: *"**Exact syntax for the KB
refresh job** (NOT created — CLI syntax only, composed from `--help` + the doc
pages below)"*. No cron job has ever been created on this VM other than the
one-shot `wf4-p13` agent-mode job. The flag names are real (re-verified
`hermes cron create --help`: `--no-agent`, `--script`, `--deliver`, `--name`,
positional `schedule [prompt]`), but the invocation as a whole is untested,
and one part is actively ambiguous: `--script` is documented as *"Path to a
script under `~/.hermes/scripts/`"* while the proposal passes a bare basename
(`kb-refresh.sh`). Whether a basename, a relative path or an absolute path is
accepted was never exercised. `docs/../CLAUDE.md` hard rule: *"CLI facts in
specs are LLM-summarized vendor docs: run `--help` on the VM before scripting
any non-interactive invocation, and measure before believing recon claims."*

**Correction.** Annotate the line: "`hermes cron create` — flags verified
against `--help`; **the invocation itself has never been run** (`kb-refresh.md`
§1). At apply time, create it, then confirm with `hermes cron list` and force
one run with `hermes cron run <job_id>` before trusting the schedule; if
`--script kb-refresh.sh` is rejected, retry with the full path."

### 6. "auto-delivers an error alert on non-zero exit or timeout" — UNSUPPORTED

> L67-68: "and **auto-delivers an error alert on non-zero exit or timeout** —
> none of which a systemd `--user` timer or crontab gives free"

**Evidence.** `kb-refresh.md` §6 grades this as *"confirmed **doc** behavior"*
and its §Not-tested says: *"The exact `agent.log` / `errors.log` line shape for
a **failed** cron run (no failing job has ever run on this VM to observe it
against)."* The only live delivery evidence is a *successful* agent-mode job.
So the single feature the proposal uses to justify `hermes cron` over a
systemd timer is vendor-doc text, never observed.

**Correction.** "…and, **per the vendor doc (`cron-script-only.md`), not yet
observed on this VM**, auto-delivers an error alert on non-zero exit or
timeout." Add to §Verify: "force a failing tick once (e.g. temporarily point
`R` at a non-existent path, or `hermes cron run <job_id>` with the remote
unreachable) and confirm the alert actually lands in Slack."

### 7. "writes a per-run artifact" (under `--no-agent`) — UNSUPPORTED

> L68: "persists in `~/.hermes/cron/jobs.json` across restarts …, writes a
> per-run artifact"

**Evidence.** `kb-refresh.md` §1 shows exactly one artifact,
`~/.hermes/cron/output/004adebba8ee/2026-08-07_19-06-01.md`, from the
**agent-mode** job `wf4-p13` — its stated contents are "job name, run time,
schedule, full prompt, and full response". No `--no-agent` job has ever run
here, so "writes a per-run artifact" for a script-only job is an
extrapolation.

**Correction.** "…writes a per-run artifact (observed for an agent-mode job;
not yet observed for `--no-agent`)". Add the artifact path to §Verify.

### 8. "persists in `~/.hermes/cron/jobs.json` across restarts" — UNSUPPORTED

> L68 (same sentence) and L162: "cron reloads `jobs.json` each tick"

**Evidence.** `kb-refresh.md` §1 derives per-tick reload from the internals
doc ("Tick Cycle step 2: Load all jobs from jobs.json") and measures only the
*preconditions* (`hermes-gateway.service` `enabled`, `Linger=yes`, ticker
heartbeat 39 s ago). No gateway restart was performed — correctly, since peers
are live on it. So restart-survival is inference, not measurement.

**Correction.** "…persists in `~/.hermes/cron/jobs.json`, which the scheduler
re-reads every tick (internals doc; restart-survival not exercised — the unit
is `enabled` with `Linger=yes`, measured)".

### 9. "Notion: no change — already live via MCP, Tars queries it today" — UNSUPPORTED

> L48: "**Notion: no change** — already live via MCP, Tars queries it today."

**Evidence.** `kb-hermes-retrieval.md` §6 shows the `notion` server
**registered** (`hermes mcp list`), and its own §Not-tested says
`hermes mcp test <name>` was **not** run. `docs/facts.md`: *"`hermes mcp list`
is registry-only (never connects); liveness = `hermes mcp test <name>`."* No
probe anywhere shows a Notion query by Tars.

**Correction.** "**Notion: no change** — the `notion` MCP server is registered
(`hermes mcp list`); liveness not re-tested for this proposal."

### 10. "Exit 0 + silent = healthy tick" — UNSUPPORTED, and the script contradicts it

> L82 (script comment): "# Hourly mc-metarepo mirror + skill-map regen. Exit 0
> + silent = healthy tick."
> L104: "Silent on success (empty stdout = silent tick), alerted on failure"

**Evidence.** Only the `fetch` and `reset` legs `exit 1`. The map-regeneration
half has no failure path: `set -u` is set but `set -e` is not, so if
`$S/SKILL.head.md` is missing or `$S` was removed (which the documented
rollback at L170 does: `rm -rf ~/.hermes/skills/mc-metarepo`), then
`{ cat …; } > "$S/.new"` fails, `cmp` differs, and the script still ends on
`rm -f` ⇒ **exit 0, empty stdout, "healthy tick"** — while `SKILL.md` has been
replaced by an empty or truncated file, or silently stopped being updated.
Nothing in any probe measures this; it is a defect in the proposed script, not
a measured behaviour.

**Correction.** Add `[ -f "$S/SKILL.head.md" ] || { echo "KB refresh FAILED:
SKILL.head.md missing"; exit 1; }` before the regeneration block, and gate the
`mv` on a non-empty `.new` (`[ -s "$S/.new" ] || { echo "KB refresh FAILED:
empty map"; exit 1; }`). Then the "exit 0 + silent = healthy" claim becomes
true instead of aspirational.

### 11. The `(c)` row's bare "**5/5**" in the lead table — UNSUPPORTED as presented

> L19: "| **(c) skill map → one `read_file`** | ✓ **958 B / 19 s** live | ✓
> ~1.2 KB | ✓ **1,717 B / 21 s** live | ✓ ~21 KB | ✓ ~4.2 KB | **5/5** |"

**Evidence.** The source scoreboard writes the same cell as
`**5/5** (2/5 fired live)` (`kb-retrieval-trials.md:721`), and its §Not-tested
opens with: *"(c) fired live on Q2, Q4, Q5 — only Q1 and Q3 were run through
the agent … The other three rows … are file sizes plus my own routing
decision, made after I had read the ground truth. **Not blind, not
measured.**"* The proposal does disclose this at L27-29, but the table cell —
the thing a reader takes away — drops the qualifier, and the Q2/Q4/Q5 cells
carry no "(est.)" marker while the (d) row's numbers are all measured.

**Correction.** Write the cell as `**5/5 (2 measured, 3 reasoned)**` and mark
the three estimated cells `~1.2 KB (est.)`, `~21 KB (est.)`, `~4.2 KB (est.)`,
matching the source table.

---

## B. Spot-checked and SUPPORTED (no change needed)

12. L16-20 scoreboard numbers — every cell matches `kb-retrieval-trials.md`
    §Scoreboard: (a) 337/369/1419/15457/1349 B; (b) 70476/6192/not-listed/7107/
    49904 B; (b′) 862/1044/-/7107/4186 B; (c) 958 B/19 s and 1717 B/21 s live;
    (d) 422 s·25 tools, 414 s, 22 s·0 tools, 113 s·41886 B, 62 s from `lcm.db`.
    Hit rates 0/5, 0/5, 3/5, 5/5, 3/5-correct-1/5-from-KB all match. SUPPORTED.
13. L31-33 "dead pointer, measured as two `search_files` 'Path not found'
    errors in baseline Q4" — SUPPORTED (`kb-retrieval-trials.md:452-453`).
    Only the *line number* is wrong (finding 1).
14. L37-43 "no MCP server / no vector DB / no LCM / no prompt stuffing / no
    `config.yaml` edit" — SUPPORTED: `_HERMES_CORE_TOOLS` is shared by
    `hermes-cli` and `hermes-slack` verbatim (`kb-hermes-retrieval.md` §2);
    `node`/`npx` absent (§6); 15 `lcm_*` tools over `lcm.db` only (§8);
    1,478,742 B of markdown (`kb-mc-metarepo.md`) vs the 0.85 gateway trim
    threshold (`docs/facts.md:81`). Two citation slips worth fixing in place:
    node absence is §6 only (L39 cites "§2,§6"), and the toolset facts are
    §2/§3 (L43 cites "§7", which carries the live proof but not the registry).
15. L45 "120 md, 1.44 MB, clean, tracks `origin/main`" — SUPPORTED
    (`kb-mc-metarepo.md` §Shape, §Git state).
16. L47 "`gaetan-metarepo` deliberately not Orca-registered" — SUPPORTED
    (`P3-orca-v2-skill.md:80`).
17. L58 "`knowledge/README.md` indexes 35 of 40 notes" — SUPPORTED
    (`kb-retrieval-trials.md:123-131`); re-measured live: 41 files in
    `knowledge/` = 40 notes + README. (`kb-mc-metarepo.md`'s "41 notes" is the
    off-by-one, not the proposal.)
18. L63-75 refresh strategy (fetch+`reset --hard` only strategy at exit 0 in
    every case; `git pull` stuck forever on divergence; `--rebase` leaves
    `.git/rebase-merge` + detached HEAD that `reset --hard` reports exit 0
    without clearing; dropped path hangs `git fetch`; `GIT_TERMINAL_PROMPT=0`)
    — all SUPPORTED, measured in `kb-refresh.md` §4 tests 1-6b.
19. L86-87 the abort guard — SUPPORTED and shell-correct: `[ -d A ] || [ -d B ]
    && abort` groups as `(A||B) && abort`, matching the verified recipe at
    `kb-refresh.md:421`.
20. L106 "`~/.hermes/scripts/` does not exist today … not tested whether
    `cron create --script` creates it" — SUPPORTED, re-verified live.
21. L120-132 push section (`learnings/README.md` rule 1 verbatim;
    `required_approving_review_count: 1` + `enforce_admins: true` + force-push
    blocked; `.claude/skills/retro/SKILL.md`; `ship-pr/SKILL.md:20` stopping
    before merge on a `learnings/` diff; label `agent-learning`; PRs
    #182/#177/#162/#160; Orca repo `id:8099e312-3232-46f2-83a9-97aeaf5de5a2`)
    — SUPPORTED by `kb-mc-metarepo.md` §Branch protection and
    `kb-push-design.md:13,17,31-33,42,58-63,299-300`. The section's own
    "(DESIGNED, NOT BUILT)" header is honest: `kb-push-design.md:523` records
    that the whole `run-create → … → worker-release` sequence is untested.
22. L136-152 the "where the winner fails" section (auto-injection untested,
    ~21 KB for governance questions, hit rate bounded by someone else's
    filenames, LCM shadowing, unattended `reset --hard`) — SUPPORTED, and it
    matches `kb-retrieval-trials.md` §"Where the winner fails" one-for-one.
    This is the strongest part of the proposal.
23. L162-163 "skills rebuild from the path→(mtime,size) manifest … SOUL
    live-reloads (~30 s)" — SUPPORTED, but by probes **not listed in the
    Evidence header**: `orca-implement.md:66` (skills prompt snapshot
    invalidated by a path→(mtime,size) manifest) and `guidelines-runtime.md`
    §3/§4 ("SOUL.md and skill files are read from disk fresh at every
    `_build_system_prompt()` call"). Add both to the Evidence line.
24. L164 "`hermes skills list` shows `mc-metarepo`" — SUPPORTED, verified live:
    hand-placed skills under `~/.hermes/skills/<name>/SKILL.md` do appear
    (`delegate-to-cooper … local … enabled`).
25. L168 "Fire at least one over **Slack** — untested surface" — SUPPORTED and
    correctly hedged (`kb-hermes-retrieval.md` §Not-tested;
    `kb-retrieval-trials.md` §Not-tested: all 7 turns were `platform=cli`).
26. Nothing in the proposal presents a write under `~/.hermes/` or a push to
    `mc-metarepo` as already done — every apply step is stated as proposed, the
    status line says "proposed, NOT applied", and both source probes assert no
    writes were made. SUPPORTED.
