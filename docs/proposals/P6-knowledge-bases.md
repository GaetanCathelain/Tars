# P6 — knowledge bases: the fix is one line, not an index

**Status:** proposed, NOT applied · **Source:** kb peer team, 2026-08-07/08 ·
**Evidence:** `status/probes/wf5/kb-{retrieval-trials,confound-trial,hermes-retrieval,mc-metarepo,refresh,push-design,push-inputs,probe-side-effects}.md`;
skills-manifest + prompt-rebuild claims rest on `wf5/orca-implement.md` §66 and `wf5/guidelines-runtime.md` §3–4 ·
**Map draft (deferred, see §1):** `status/probes/wf5/kb-map-draft-SKILL.md` · **Adversarially reviewed:**
three passes, `kb-verify-{claims,coverage,final}.md` (11 + 8 + 13 findings, all folded in) —
the coverage pass **overturned this proposal's original recommendation** (§0) and the final
pass caught an apply step that would have destroyed a correct SOUL sentence · **Decide:** Gaetan ·
[P3](P3-orca-v2-skill.md)/[P4](P4-soul-guidelines-redesign.md) **have landed**; P6 edits the
same `SOUL.md`, so apply it **in its own session**, never alongside another `~/.hermes/` writer.

**The recommendation in one line:** Tars has **no** pointer to its knowledge base — the clone
sits on the VM at `~/dev/mc-metarepo` and `SOUL.md` never names it. Add one sentence, keep the
clone fresh hourly, and **install nothing else** — measured, that is the whole win.

## Lead with the measurement

Five real questions Gaetan would DM, fired against the clone through each candidate
(`kb-retrieval-trials.md`, full transcripts). Bytes = tool output the model ingests;
`✓` correct+complete · `~` partial · `✗` wrong/absent.

| Mechanism | Q1 NM/MC | Q2 Datadog 92% | Q3 rotate key (vocab mismatch) | Q4 learning→metarepo (3 files) | Q5 bon de retour | hit rate |
|---|---|---|---|---|---|---|
| **(d) Tars TODAY** | ~ 422 s, 25 tools, Orca worker | ✓ 414 s, Orca worker, **live prod Datadog** | ✗ 22 s, 0 tools, contradicts the note | ✓ 113 s, 41,886 B blind grep | ✓ 62 s, from `lcm.db` | 3/5 correct, **1/5 from the KB** |
| (a) naive `rg` · (b) `rg -l` → top file | ✗ 337 B · ✗ 70,476 B | ✗ 369 B · ✗ 6,192 B | ✗ 1,419 B · ✗ not listed | ~ 15,457 B · ~ 7,107 B | ✗ 1,349 B · ~ 49,904 B | 0/5 each — (a) locates and never answers, (b)'s "top file" is traversal order |
| (b′) (b) + `knowledge/`-first | ✓ 862 B | ✓ 1,044 B | ✗ not listed | ~ 7,107 B | ✓ 4,186 B | 3/5 — the fallback |
| **(c⁻) right path, no map** | ✓ **958 B / 23 s** live | — | ✓ **4,808 B / 38 s** live | — | — | **2/2 on the two failures** |
| (c) skill map → one `read_file` | ✓ **958 B / 19 s** live | ✓* ~1.2 KB | ✓ **1,717 B / 21 s** live | ✓* ~21 KB | ✓* ~4.2 KB | 5/5 — 2 measured, 3 reasoned |
| (e) filesystem MCP server | = (b′) | = (b′) | ✗ | ~ | = (b′) | 3/5 **by inspection only — not runnable here** (`node`/`npx` absent) |

`✓*` = **estimated, never fired**: file size plus the trial agent's own routing choice, made
after it had read ground truth. Not blind, not measured (`kb-retrieval-trials.md` §Not-tested).
**Simplicity axis**, since the mandate ranks by it: (a)/(b)/(b′)/(c⁻) install **nothing**;
(c) installs 1 skill + 1 script + 1 cron job; (e) needs an MCP server *and* a node runtime the
VM does not have.

"3/5 correct" flatters today badly. Q1 spent 422 s delegating an Orca worker to read
product source and **still missed the `business_entity_id` trap** the note exists to
prevent; Q2 spent 414 s delegating a **live production Datadog investigation** to
rediscover a 971-byte note written about that exact endpoint — both left a local branch and
a worktree behind (audited: nothing reached a remote, `kb-probe-side-effects.md`); Q5 came
from conversation memory, not the file. Tars is not bad at answering; it is bad at knowing
the answer is already written down.

## 0 — The finding that overturned this proposal

The first draft recommended the skill map on a 5/5 score. Adversarial review caught that
**row (c) bundled two changes** — fixing the dead pointer *and* installing the map — so every
win was confounded. The laziest candidate of all, *"Tars simply knows the right path"*, had
never been fired. It was measured afterwards (`kb-confound-trial.md`), by naming the clone in
the question with no map installed:

| | Q1 (baseline PARTIAL) | Q3 (baseline WRONG) |
|---|---|---|
| (c⁻) right path, no map | **CORRECT** — 23 s, 958 B, 2 tools | **CORRECT**, incl. the "do not erase logs" caveat — 38 s, 4,808 B, 3 tools |
| (c) map + `read_file` | CORRECT — 19 s, 958 B, 1 tool | CORRECT — 21 s, 1,717 B, 2 tools |

Two questions, one run each, no repeats (**n=1 per cell**, nondeterministic model).

**Every correctness win credited to the map belongs to the pointer fix.** The map's only
measured gain is efficiency on one question (2.8× fewer tool bytes, 17 s), n=1, in exchange
for a 3.3 KB artifact that must be regenerated hourly (and nothing measurable on Q1).
**So the map is deferred, not
rejected** — §1 states the single experiment that would justify it. Proxy limit, stated
plainly: a path in the *question* is not a path in `SOUL.md` (different prompt position, and
SOUL is always present while a question is not); the SOUL fix should be re-measured after it
lands, which §Verify requires.

**Blocking finding, one sentence to add:** `~/dev/mc-metarepo` is named **nowhere** in
`SOUL.md` (`grep -ic mc-metarepo ~/.hermes/SOUL.md` → 0, re-checked after P4 landed). The
only `metarepo` Tars can see is in the Phase-2 preferences section — *"Its source of truth is
`~/dev/gaetan-metarepo` on cooper"* — which is a **true statement about Gaetan's preferences
repo**, not a knowledge-base pointer. Measured harm: baseline Q4 aimed `search_files` at it
twice and got `"Path not found"` (it is on cooper, not the VM). So the fix is to **add** a
sentence, never to rename that one. **Do not trust a line number** — P4's rewrite moved it
54 → 81 → 88 across this review. Locate at apply time: `grep -n metarepo ~/.hermes/SOUL.md`.

**What is NOT proposed.** **No vector DB / RAG index** — the team ruled it out first,
`mc-metarepo/MCP.md` verbatim: *"No vector DB, no embeddings, no index service … grep +
structured JSON records beat a vector store here"*. **No MCP server** — their proposed
Knowledge MCP re-exposes `search_files`, already in `_HERMES_CORE_TOOLS` on CLI **and**
Slack (`kb-hermes-retrieval.md` §2), and `node`/`npx` are absent from the VM (§6). **No
memory/RAG provider** — LCM's 15 `lcm_*` tools index `lcm.db`, the agent's own past
conversations, never a directory (§8). Hindsight stays as specified: provider set, keys
deliberately absent, `hermes memory status` → `not available` (`wf4-probes.md` §11) — P6
needs no provider and changes nothing there. **No prompt stuffing** — 1.44 MB ≈ 370 K tokens
against a 0.85 trim threshold, versus ~1 KB/question measured. **No `config.yaml` edit** —
`read_file`/`search_files`/`terminal` are already live (`kb-hermes-retrieval.md` §2–3,
proven live in §7).

**Sources covered.** **mc-metarepo only** (120 md, 1.44 MB, clean, tracks `origin/main`):
search + hourly refresh now, push later. **`gaetan-metarepo`: no** — deliberately not
Orca-registered (P3) and absent from the VM; the SOUL sentence naming it stays as-is —
it is about preferences, and it is true. **Notion: no change** — the `notion` MCP server is *registered*
(`hermes mcp list`); liveness was not re-tested for this proposal, and `hermes mcp list` is
registry-only (`docs/facts.md`). **The 10 submodules: no** — uninitialized here, so product
source is simply not in the KB; the SOUL sentence should say so, which is what kills the Q1
"delegate a code read" reflex (P6-b).

## 1 — Search: name the clone in SOUL. The map waits for evidence.

**Ship — this is the entire deliverable, so here it is literally.** Add to `SOUL.md`:

> My knowledge base on this machine is the git clone at `~/dev/mc-metarepo` — search it with
> `search_files`/`read_file` before delegating anything. Its submodules are empty, so product
> source code is **not** there.

*"on this machine"* is load-bearing: it is what separates this from the Phase-2 sentence
about `gaetan-metarepo` **on cooper**, and it is the wording the trial actually measured
(`kb-confound-trial.md`). Sentence 2 is P6-b's routing hint plus the anti-delegation rule —
the two lines that kill Q1's "spawn an Orca worker to read product source" reflex.

Nothing else: Tars already holds `search_files` + `read_file` + `terminal` on CLI **and**
Slack — nothing to enable, no `config.yaml` edit, no new tool. Measured, this turns a 422 s
Orca delegation that missed the point into a 23 s local read, and a confidently wrong answer
into a correct one.

**Deferred:** the generated skill map (`kb-map-draft-SKILL.md`, 3,279 B = 0.22 % of the KB).
The argument for it is real but unproven: routing beats searching, because the repo's
one-note-one-fact slug convention does the semantic indexing a vector store would —
`sudo-argv-secrets-authlog` is a handle a model recognises from "API key inline in a command"
although `rg -il 'api key'` hits 13 other files and not that one — **though that is naive
`rg`'s failure, not the agent's**: given the correct root path and no map, Tars' own
`search_files` still reached that note on meaning (`kb-confound-trial.md` §Q3), which is
exactly why this argument is *unproven* rather than merely unmeasured. And it would have to be
**generated, never hand-written**: the repo's own `knowledge/README.md` indexes 35 of 40
notes, the Q3 note among the 5 missing.

**The one experiment that would justify it** — a question the corrected pointer gets *wrong*
and the map gets *right*. Candidates from the trials: a thin-slug note (`done-is-not-exercised`
carries little signal), a Q4-style governance question spread over 3 files, or
`schemas/`/`index/` routing. Absent that, the map saves ~3 KB of tool output on the one
question where it helped (Q3; **nothing measurable on Q1**, n=1) and costs an eight-line
regeneration tail on the hourly job — pay it when a measurement asks for it, not before. Its
*rules* (never read the whole repo, do not delegate a code read for a documented fact) are two
lines, and §1's sentence 2 already carries them.

## 2 — Refresh: `hermes cron --no-agent`, `fetch` + `reset --hard`, guarded

`--no-agent --script` costs **zero model tokens per tick**; the scheduler re-reads
`~/.hermes/cron/jobs.json` every tick (internals doc — restart-survival was *not* exercised,
no gateway restart while peers are live; the unit is `enabled` with `Linger=yes`, measured);
it writes a per-run artifact (observed for an **agent-mode** job, not yet for `--no-agent`);
and, **per the vendor doc only, never observed on this VM**, it auto-delivers an error alert
on non-zero exit or timeout. That last one is the whole reason to prefer it over a systemd
`--user` timer or crontab, so §Verify forces a failing tick before trusting it
(`kb-refresh.md` §1,§6 and its §Not-tested). `fetch`
+ `reset --hard origin/main` is the **only** strategy returning exit 0 in every dirty /
conflicting / rewritten-history test, *including self-healing from another strategy's
stuck state*; plain `git pull` gets stuck forever on divergence, same fatal error every
tick, and `--rebase` leaves a rebase-in-progress + detached HEAD that a bare `reset --hard`
**does not clear while reporting exit 0** ⇒ abort guard. A dropped (not refused) network
path hangs `git fetch` **indefinitely** ⇒ `timeout`; bad credentials block on a headless
prompt ⇒ `GIT_TERMINAL_PROMPT=0`.

**Proposed, never executed** — no write under `~/.hermes/` was made by this investigation.
Run it *from an interactive session on the VM* (`ssh gaetan@192.168.0.9`, then the block);
written as one local `&&` chain it would `mkdir` on cooper.

```bash
mkdir -p ~/.hermes/scripts
cat > ~/.hermes/scripts/kb-refresh.sh <<'EOF'
#!/usr/bin/env bash
# Hourly mc-metarepo mirror. Exit 0 + silent = healthy tick.
set -u; export GIT_TERMINAL_PROMPT=0
R="$HOME/dev/mc-metarepo"
# a leftover rebase/merge from a hand-run command wedges every future tick
[ -d "$R/.git/rebase-merge" ] || [ -d "$R/.git/rebase-apply" ] && git -C "$R" rebase --abort 2>/dev/null
[ -f "$R/.git/MERGE_HEAD" ] && git -C "$R" merge --abort 2>/dev/null
timeout 120 git -C "$R" fetch --quiet origin \
  || { echo "KB refresh FAILED: git fetch. mc-metarepo is STALE."; exit 1; }
git -C "$R" reset --hard --quiet origin/main || { echo "KB refresh FAILED: reset --hard"; exit 1; }
# reset --hard reports exit 0 even from a stuck rebase + detached HEAD — measured, kb-refresh.md §6b
[ "$(git -C "$R" rev-parse --abbrev-ref HEAD)" = main ] || { echo "KB refresh FAILED: mirror not on main"; exit 1; }
EOF
chmod +x ~/.hermes/scripts/kb-refresh.sh
~/.local/bin/hermes cron create "every 1h" --no-agent --script kb-refresh.sh --deliver slack --name "mc-metarepo-refresh"
```

Flags verified against live `--help`; **the invocation itself has never been run**
(`kb-refresh.md` §1). `--script` is documented as *a path under* `~/.hermes/scripts/` and we
pass a bare basename — if rejected, retry with the full path. That directory does not exist
today (whether `cron create --script` creates it: **not tested**), hence the `mkdir`.
`set -e` is deliberately absent — the two abort guards must be allowed to fail.

*Only if the map is later adopted (§1)*, the same script grows a regeneration tail: `cat` the
hand-written `SKILL.head.md`, append `ls knowledge/*.md` and `ls learnings/*/*.md` as slug
lists, write via `$S/.new`, and **guard both ends** — `[ -f "$S/SKILL.head.md" ]` before and
`[ -s "$S/.new" ]` after, or a failed regen installs a truncated map and still exits 0. Full
tail is eight lines. Both lists generated rather than hand-written: hand lists rot, and this
repo proves it (its own index carries 35 of 40 notes).

**Is Tars told a refresh happened? No** — the clone is just files it reads on demand, no event,
no message. Staleness surfaces only to *Gaetan's* Slack, and only on failure, so Tars can answer
from an hour-old KB (or arbitrarily older, if an alert is ignored) without knowing. Deliberate:
telling Tars costs a turn an hour and buys nothing it can act on.

## 3 — Push (DESIGNED, NOT BUILT)

Full design + literal brief: `status/probes/wf5/kb-push-design.md`. P3 and P4 **have landed**
(`apply-p3-orca-v2.md`, `apply-p4-soul.md`), so the mechanism this rests on exists.
**Tars never authors the note** — live SOUL rule 1 forbids Tars producing the deliverable
*"not because it's only markdown"*. (Rule 2's one exception is Tars' own `skills/<n>/SKILL.md`
mirror, self-merged; it does not reach mc-metarepo, and rule 2 "stands whole" in a repo Tars
delegated work in.) A push is: Tars quotes Gaetan,
hands the quote to a delegated Claude Code session via Orca on cooper, verifies the PR —
quoting Gaetan back is evidence, not authorship, rule 1's own carve-out. **Trigger:
human-only**, Gaetan says "remember this"; periodic harvest **rejected** — Hindsight is
off, so Tars has no record of what it learned, and a scheduled harvest with no source
produces filler on a cadence into a team repo. The trigger *is* the primary brake: a rate
bounded by deliberate human asks cannot leak.

**The binding rule of the target repo**, `learnings/README.md` rule 1, verbatim:
> **A human MUST review and merge every learnings PR; agents MUST NEVER self-merge.**

Also enforced by GitHub (`required_approving_review_count: 1`, `enforce_admins: true`,
force-push blocked) ⇒ **direct-to-main is not an open decision**, closed twice over.

Four more of its rules bind the brief, each with a real cost (`kb-mc-metarepo.md`
§Contribution rules quotes them verbatim):

1. **English + ASD-STE100 Simplified Technical English, CI-gated** — `ste-lint` fails the PR
   on an STE error in an added line. A session writing normal prose gets a red check.
2. **`learnings/` rule 5: no PII, no secrets — ever.** The material is a Slack quote, which
   is exactly where tokens and customer identifiers live. Most likely rule to bite: Tars must
   **refuse to dispatch** a credential-shaped quote, not pass it through.
3. **A `knowledge/` note needs three surfaces**, not one: the file, its `@import` line in
   `CLAUDE.md`, its row in `knowledge/README.md`. That is the real price of P6-e's "yes".
4. **`--repo` on every `gh` call, `-C <dir>` on every `git` call** — an agent shell's cwd
   resets to the metarepo root between commands; on **2026-08-05 skipping this merged an
   unrelated PR**. Their `CONTRIBUTING.md` records the incident.

**How Tars verifies, without doing the work** (reading a PR is verification, rule 2's own
carve-out; fixing it is not) — `kb-push-design.md` §5:

```bash
gh pr view <n> --repo mobile-club/metarepo --json number,url,state,labels,files,additions,deletions
gh pr diff <n> --repo mobile-club/metarepo --name-only
gh pr checks <n> --repo mobile-club/metarepo        # ste-lint
```

Four hard assertions: PR is `OPEN` · `labels` contains `agent-learning` · **every** changed
path is under `learnings/` (or one of the three `knowledge/` surfaces the brief allows) ·
`state != MERGED` — **a merged PR is an incident, not a success**, and gets reported as one.
Out-of-scope paths: Tars reports and asks, it does not fix. Junk-but-in-scope: Tars says so
plainly and leaves it to the human reviewer; it is not the quality gate.

**Nothing to build:** mc-metarepo ships the procedure as its own skill —
`.claude/skills/retro/SKILL.md` (harvest → map coverage → place → PR), and
`ship-pr/SKILL.md:20` already stops before merge on a `learnings/` diff. The brief is *"run
this repo's own `retro` on this one fact; skip Extract, skill-pass and the AskUserQuestion
gate; stop at the PR"*. Label `agent-learning` exists; prior art PRs #182/#177/#162/#160.
**Rule stated here because no doc states it:** a push NEVER goes through the VM's
`~/dev/mc-metarepo` — that clone is a read-only mirror; every write happens in an Orca
worktree on cooper (repo `id:8099e312-…`, a *different* checkout).

## Where this fails, and the risk

**The SOUL fix is measured through a proxy, not in place** — the path was named in the
*question*, not in `SOUL.md`, and SOUL sits in a different prompt position and is always
present. n=2. §Verify re-fires all five questions after the edit lands; if the win does not
reproduce, that is the signal the map (or the map's *rules*) is doing real work after all.
**LCM can still answer first from stale memory** — baseline Q5 was correct and read no file;
precedence between `lcm_*` recall and a fresh `read_file` is untested. **Hit rate is bounded
by someone else's filenames** in a repo Tars does not control (`done-is-not-exercised` carries
little signal). **Governance questions are not ~1 KB** — Q4-style answers spread over 3 files
cost ~7–21 KB, still far under any prompt-stuffing design, but not the headline number.
**Submodules are never refreshed** (10, all SSH URLs) — P6-b. And if the map is later adopted,
its biggest unknown is untouched: **skill auto-injection was never tested**, the map was only
prepended to a prompt, so whether `description:` frontmatter loads it unprompted is unproven.

The refresh runs `git reset --hard` unattended, forever — correct **only** because that clone
is a mirror nobody edits by hand; the hour anyone does, their work is silently gone and the
script cannot tell. Pointing Tars at a **team** repo means a wrong note routes straight into
its answers, and mc-metarepo's human-merge gate is the only filter — it is not ours. With
P3+P4 a KB question can still end in a delegated cooper session with sudo; the fix *reduces*
that reflex (measured: 422 s + a branch → 23 s + a local read), it does not gate it.

## Order · restart · verify · rollback

- **Order:** (1) **in its own session** — P3/P4 have landed, but `SOUL.md` has more than one
  author now, and a raw append here once nearly dropped a sibling's config stanza. Re-read
  before writing.
  (2) **SOUL sentence first** — standalone (P6-a). **Add**, never `sed`: the existing
  `gaetan-metarepo` sentence is correct and must survive untouched.
  ```bash
  cp ~/.hermes/SOUL.md ~/.hermes/SOUL.md.bak-p6
  grep -n metarepo ~/.hermes/SOUL.md     # locate; 54 → 81 → 88 across this review — it moves
  # then, BY HAND under flock, append §1's two sentences to the tools/hard-rules section:
  #   flock ~/.hermes/.wf3.lock -c 'cat >> ~/.hermes/SOUL.md'   (or an editor under the lock)
  grep -ic mc-metarepo ~/.hermes/SOUL.md # 1+ ; and confirm the Phase-2 sentence is unchanged
  ```
  Optionally append *"— on cooper, not on this VM"* to the Phase-2 sentence, which keeps it
  true and removes the misreading that produced Q4's two `"Path not found"` errors.
  (3) `kb-refresh.sh`, (4) run it once by hand, (5) then `hermes cron create`.
- **Restart:** none expected — SOUL and skill files are re-read from disk at every
  `_build_system_prompt()` (`guidelines-runtime.md` §3–4), skills rebuild from the
  path→(mtime,size) manifest (`orca-implement.md` §66), and the cron scheduler re-reads
  `jobs.json` each tick (internals doc; **not exercised** — no restart was performed).
  `.bak` first, and every edit to a **live-reloaded** file (`SOUL.md`, `config.yaml`, `.env`)
  under `flock ~/.hermes/.wf3.lock` — a new file in a new directory has no concurrent writer.
- **Verify — this is the experiment, not a formality.** Re-fire all 5 trial questions as
  `hermes chat -Q -q` oneshots, **with no path named in the question** (that is the whole
  point: the pointer must now come from SOUL), and grep `~/.hermes/logs/agent.log` by session
  id. Pass = ≥4/5 correct, each well under the 422 s/414 s baselines (≤~40 s on Q1 and Q3,
  the two actually measured under (c⁻)), **none delegating to Orca**. Q3 is the
  discriminator (wrong today), Q1 the cost one (422 s today). Fire at least one over
  **Slack** — untested surface. If a question fails that (c⁻) got right, the proxy misled us
  and §1's map goes back on the table. Refresh: force one tick with `hermes cron run <job_id>`,
  check `hermes cron runs`, the per-run artifact under `~/.hermes/cron/output/<job>/`, and
  `git -C ~/dev/mc-metarepo status -sb` clean. **Then force one *failing* tick** (point `R` at
  a nonexistent path) and confirm the alert lands in Slack — that delivery is doc-only today
  and is the sole reason `hermes cron` beat a systemd timer.
- **Rollback:** `hermes cron remove <job_id>` **first**, then `rm ~/.hermes/scripts/kb-refresh.sh`
  (the other order lets a tick fire a missing script and page you for nothing); restore
  `SOUL.md.bak` (**check first** that no peer edited SOUL after your `.bak`, or you revert
  their work too). If the map was adopted: `rm -rf ~/.hermes/skills/mc-metarepo` — generated,
  nothing lost.

## Open decisions

Numbered P6-a…P6-c here. `kb-push-design.md` runs **its own** P6-a…P6-c namespace for
push-internal choices — where this table cites it, it says so.

| # | Question | Recommendation |
|---|---|---|
| P6-a | The missing SOUL sentence: standalone edit, or folded into a SOUL replacement? | **Standalone.** P4 landed on 2026-08-07 and its replacement text kept `~/dev/gaetan-metarepo` and named no KB — there is nothing left to fold into |
| P6-b | Submodules: leave uninitialized, or init them into the KB? | **Leave empty.** 10 product repos on SSH auth; their absence is itself the routing hint ("the code is not here, do not go looking") |
| P6-c | The skill map: install it now anyway, or hold until a question justifies it? | **Hold.** §0 shows its correctness win was the pointer fix; the residual gain is ~3 KB on one question, n=1, against an hourly regeneration job. Revisit if §Verify's re-fire disappoints |
| P6-d | Push trigger and SOUL rule 8's beat (push-design's own P6-a/b) | **Human-only ("remember this"), announce before dispatch** — one Slack line, no wait |
| P6-e | May a push target `knowledge/`, or is `learnings/**` the ceiling? | **`knowledge/` allowed with written justification** — it is auto-imported into every teammate's context *and* costs three surfaces (file + `@import` + README row), so the bar is high, not infinite |
| P6-f | Brakes: one open Tars-originated learnings PR at a time, keep push worktrees, and may Tars close an out-of-scope PR? | **Yes / keep / no** — one `gh pr list` preflight, no state; worktrees are evidence at v1 volume; closing a PR is a write against a team repo, so Tars **asks** (rule 2) |

Withdrawn after review: an earlier draft raised `memory.provider: hindsight` being set in the
live config as a contradiction. It is not — provider set with keys deliberately absent is the
*specified* state (`tars-profile.md` §87, acceptance probe `wf4-probes.md` §236, live
`hermes memory status` → `not available`). Nothing to reconcile; P6 touches no provider.

## Decision

- [ ] **Apply P6 as recommended** — SOUL line + hourly refresh job, no skill map — after
      P3/P4 have settled
- [ ] SOUL line only; refresh job later
- [ ] Apply *with* the skill map anyway (overrides P6-c)
- [ ] Apply with edits · [ ] Not yet
- P6-a SOUL sentence: [ ] standalone (recommended — P4 already landed) · [ ] defer to a later SOUL pass
- P6-b submodules: [ ] leave empty · [ ] init
- P6-c map: [ ] hold (recommended) · [ ] install now
- P6-d trigger: [ ] human-only · [ ] also end-of-task — P6-e target: [ ] `learnings/**` + justified `knowledge/` · [ ] `learnings/**` only
- P6-f brakes: [ ] one open PR at a time · [ ] keep worktrees · [ ] Tars may close an out-of-scope PR
