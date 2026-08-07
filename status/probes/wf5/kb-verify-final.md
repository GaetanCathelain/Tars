# Adversarial verification — P6 after the overturn rewrite (final pass)

Verifier pass 2026-08-08, read-only. Target: `docs/proposals/P6-knowledge-bases.md`
(321 lines, post-rewrite). Cross-read: `kb-confound-trial.md`, `kb-verify-claims.md`,
`kb-verify-coverage.md`, `kb-retrieval-trials.md`, `kb-refresh.md`, `kb-mc-metarepo.md`,
`kb-push-design.md`, `docs/plans/knowledge-bases-proposal.md` (mandate), `CLAUDE.md`,
`docs/facts.md`, `docs/proposals/README.md`, `status/probes/wf5/apply-p4-soul.md`.
Live re-checks on the VM (**read-only**, no writes, no sops, no git writes):
`grep -n metarepo ~/.hermes/SOUL.md`, `grep -ic mc-metarepo`, `sed -n '1,60p'` and
`'85,92p'` of SOUL.md, `ls ~/.hermes/skills`, `ls -d ~/.hermes/scripts`, `stat SOUL.md`.
The proposal was **not** edited; nothing was committed.

**Verdict: FIX FIRST.** 2 blocking · 4 major · 7 minor. The overturn itself is
argued honestly and the numbers are exact; what breaks is the *apply step*, which
still edits the wrong sentence, and four staleness/label leftovers.

---

## Blocking

### 1. The `sed` in §Order edits a sentence that is not the KB pointer — CONTRADICTED

> L264: ``flock ~/.hermes/.wf3.lock -c "sed -i 's|~/dev/gaetan-metarepo|~/dev/mc-metarepo|' ~/.hermes/SOUL.md"``
> L65-67: "Tars' only KB pointer is `~/dev/gaetan-metarepo` — **a path that does not exist on the VM**"
> L88: "its dead SOUL pointer is **deleted**, not repaired by cloning it"

**Evidence — measured live today** (`sed -n '85,92p' ~/.hermes/SOUL.md`, VM):

```
## Phase 2 — Gaetan's knowledge and preferences
…
This section will carry Gaetan's own working knowledge and preferences, so that
"what would Gaetan do" is something I answer from what he actually decided … his
code style, tooling, workflow, communication and security preferences … Its
source of truth is `~/dev/gaetan-metarepo` on cooper; the live mechanism on my
side is `~/.hermes/memories/USER.md`. Until this section is filled in, I ask
instead of assuming.
```

That sentence is **provenance for Gaetan's personal preferences**, not a knowledge-base
pointer, and its claim is *true*: `gaetan-metarepo` really is on cooper and really is the
source of truth for `USER.md`. The measured harm (baseline Q4 aiming `search_files` at it,
`kb-retrieval-trials.md:650-651`) is the agent **misreading** it as a KB, not the sentence
being wrong.

The proposed `sed` therefore (a) destroys a correct statement about the preferences repo,
(b) asserts that `mc-metarepo` is the source of truth for Gaetan's code style and security
preferences — false, and (c) leaves the resulting sentence saying **"on cooper"**, which is
the exact delegate-to-cooper reflex the proposal is trying to kill (`mc-metarepo` also
exists on cooper — `kb-push-design.md` repo `id:8099e312-…`, path `/home/gaetan/dev/mc-metarepo`).
L267-268 half-notices this ("reword it to name the VM clone, by hand") but treats it as
cosmetic, and it contradicts L88's "deleted, not repaired": the `sed` neither deletes nor
adds, it renames.

**Correction.** Drop the `sed`. Step (2) becomes an **added** sentence (and, optionally,
"on cooper — not on this VM" appended to the Phase-2 sentence, which is then still true).
Write it under `flock` with a `.bak`, e.g.:

```bash
cp ~/.hermes/SOUL.md ~/.hermes/SOUL.md.bak-p6
grep -n metarepo ~/.hermes/SOUL.md   # locate; it was 54, then 81, now 88 — it moves
# append to the "Hard rules"/tools section, by hand, after re-reading the file
```

### 2. The one sentence the whole proposal now recommends is never written down — UNSUPPORTED

> L97-98: "**Ship:** one sentence in `SOUL.md` naming `~/dev/mc-metarepo` as the knowledge
> base on this machine."
> L92-93: "the SOUL sentence **should say so**" (submodules empty)

`kb-verify-coverage.md` fix #6 ("Write the SOUL edit literally — the current line, the
replacement line") is the one fix the rewrite did not discharge; the header's "all folded
in" (L7-8) is therefore slightly over-claimed. With the map deferred, this sentence **is**
the deliverable, and no reviewer can approve text that does not exist. The measured proxy
wording (`kb-confound-trial.md:32-36`) is *"The knowledge base is a git clone at
`~/dev/mc-metarepo` on this machine"* — "on this machine" is load-bearing and the current
apply step loses it.

**Correction.** Put the literal replacement in §1, e.g.:

> My knowledge base on this machine is the git clone at `~/dev/mc-metarepo` — search it
> with `search_files`/`read_file` before delegating anything. Its submodules are empty, so
> product source code is **not** there.

(Second clause = P6-b's routing hint; third = the anti-delegation rule §1 L116 already
says is two lines.)

---

## Major

### 3. "Provisional on P3+P4 landing (both unchecked today)" — CONTRADICTED

> L179: "Provisional on P3+P4 landing (**both unchecked today**)."
> L9: "**Must NOT be applied concurrently with P3/P4**"
> L298 (P6-a): "**Fold into P4 if P4 lands first**, else edit standalone"

**Evidence.** P4 is applied: `status/probes/wf5/apply-p4-soul.md` §"Post-apply state" —
*"sha256 `c0405ea6…d482` … equals `artifacts/SOUL-proposed.md` exactly → the installed
file is the approved text"*, backup `SOUL.md.bak-p4`. Live SOUL is now 5,708 B (was 1,472 B
pre-P4) and carries the P4 rule set verbatim (rule 1 "I never produce the deliverable
myself", rule 2 "I never merge, approve or push — with exactly one exception"). P3 is
applied too (git log `7c3dd8c`, `432a2c0`, `eca0ec3`, `d0d7091` — live `delegate-to-cooper`
skill edits; `status/probes/wf5/apply-p3-orca-v2.md` records a live E2E).

Consequence, and this is the part that matters: **P4 landed and kept the dead pointer**
(`grep -ic mc-metarepo ~/.hermes/SOUL.md` → `0`, re-measured today). So P6-a's recommended
branch ("fold into P4") is dead, and the fold is no longer available.

**Correction.** L179 → "P3 and P4 **have landed** (`apply-p3-orca-v2.md`, `apply-p4-soul.md`)".
P6-a → "**Standalone.** P4 landed on 2026-08-07 and its replacement text kept
`~/dev/gaetan-metarepo`; there is nothing left to fold into." Keep the "never in the same
session as another `~/.hermes/` writer" caution (L9, L257-258) but state it as a
concurrency rule, not as "P3/P4 pending".

### 4. Open-decisions namespace header contradicts its own table — INCONSISTENT

> L293: "Numbered **P6-a…P6-e** here."

The table immediately below contains **P6-f** (L300), and the Decision box offers a P6-f
checkbox (L318). Row order is also `a, b, f, c, d, e`. This is the same class of defect
`kb-verify-coverage.md` fix #7 raised (then: "no P6-d exists").

**Correction.** L293 → "Numbered P6-a…P6-f here", and move the P6-f row to the end (or
renumber the map decision P6-f → P6-c and shift). Cross-refs are otherwise clean:
P6-a (L259) = SOUL fold, P6-b (L93, L244) = submodules, P6-f (L300, L315, L318) = the map,
P6-c/d/e map correctly onto `kb-push-design.md`'s P6-a/b (trigger + rule-8 beat), P6-c
(`knowledge/` ceiling) and P6-d/e/f (brakes) — the "(push-design's own P6-a/b)" annotation
at L301 is accurate.

### 5. "the map buys ~3 KB of tool output **per question**" — UNSUPPORTED

> L114: "Absent that, the map buys ~3 KB of tool output per question and costs an hourly
> regeneration job"

`kb-confound-trial.md:204-207`: *"on Q3, **2.8× fewer tool-output bytes** (1,717 vs 4,808)
and **17 s** … **On Q1 it bought nothing measurable** — identical 958 bytes, 19 s vs 23 s,
inside run-to-run noise."* One of two questions, n=1. P6's own P6-f row states it correctly
("~3 KB on one question, n=1"); L114 generalises to "per question" and contradicts it.
Second half also overstates: per L165-170 the map does **not** cost a new job, it costs an
**eight-line tail** on the hourly job §2 installs anyway.

**Correction.** "the map saves ~3 KB of tool output on the one question where it helped
(Q3; nothing measurable on Q1, n=1) and costs an eight-line regeneration tail on the hourly
job".

### 6. §1's "routing beats searching" omits the measurement that undercuts it — INCOMPLETE

> L105-108: "routing beats searching, because the repo's one-note-one-fact slug convention
> does the semantic indexing a vector store would — `sudo-argv-secrets-authlog` is a handle
> a model recognises from 'API key inline in a command' although `rg -il 'api key'` hits 13
> other files and not that one."

The 13-file fact is SUPPORTED (`kb-retrieval-trials.md:98-113`), but it is about **naive
`rg`**, and the very trial that overturned this proposal measured the opposite for the
agent: `kb-confound-trial.md:170-174` — *"no map, so no slug to recognise — it reached the
right note via `search_files` (4,697 chars) … the vocabulary mismatch that killed naive
`rg` did **not** kill the agent's own `search_files`: given the correct root path, the
model searched on meaning, not on 'api key'."* §1 presents the strongest surviving argument
for the map using evidence the follow-up trial already weakened, in the one section a
reader turns to when deciding whether to override the deferral.

**Correction.** One clause after "…and not that one": "— though with the correct root path
and no map, Tars' own `search_files` still found that note on meaning
(`kb-confound-trial.md` §Q3), which is why this argument is *unproven*, not merely
unmeasured."

---

## Minor

### 7. §Verify's "each under ~40 s" pass bar — UNSUPPORTED

> L279: "Pass = ≥4/5 correct, **each under ~40 s**, none delegating to Orca."

No measurement supports ~40 s for Q2/Q4/Q5. Measured: (c⁻) Q3 = **38 s** (2 s of margin on
a nondeterministic model), baseline Q4 = **113 s** *and correct from the KB*, baseline Q5 =
**62 s** *and correct*. A re-fire where Q4 answers correctly in 70 s would be scored a
FAIL. **Correction.** "each well under the 422 s/414 s baselines; ≤~40 s on Q1/Q3, the two
questions measured under (c⁻)".

### 8. Sample size stated only outside §0 — precision

> L237-238 (§Where this fails): "n=2."
> §0 states "n=1" only for the map's efficiency gain (L58).

`kb-confound-trial.md:68`: *"**n=1 per cell**, a nondeterministic model, and no repeats."*
"n=2" is defensible as "two questions" but reads as "two runs". §0 — the section that
carries the overturn — never states the sample size at all. **Correction.** §0, after the
table: "Two questions, one run each, no repeats (n=1 per cell)". The proxy limitation
itself *is* stated where a reader sees it (L60-63 in §0, again at L235-238) — SUPPORTED,
no change needed there.

### 9. The abort guard's own failure is unchecked — the one silent path left in the script

> L149-150: `[ -d "$R/.git/rebase-merge" ] || [ -d "$R/.git/rebase-apply" ] && git -C "$R" rebase --abort 2>/dev/null`

Shell is **correct**: `(A||B) && abort` matches the verified recipe (`kb-refresh.md:429`);
`set -u` without `set -e` is right and L163 says why; both failure legs `exit 1`; the last
command is the `reset`, so a healthy tick is exit 0 with empty stdout (`--quiet` on both
git calls) — the comment "Exit 0 + silent = healthy tick" (L145) is now **true**, which
was `kb-verify-claims.md` finding #10 and is properly discharged by deleting the regen tail.
Residual: if `rebase --abort` itself fails, the script proceeds and `kb-refresh.md` §6b
proves `reset --hard` then **reports exit 0 while `.git/rebase-merge` and a detached HEAD
survive** — a healthy-looking tick over a stuck mirror. **Correction (optional, one line
after the reset):** `[ "$(git -C "$R" rev-parse --abbrev-ref HEAD)" = main ] || { echo "KB refresh FAILED: mirror not on main"; exit 1; }`.
Also inert-but-harmless: `chmod +x` + shebang — `kb-refresh.md` §1 records *"no `#!/…`
shebang honoured; `.sh`/`.bash` run via `bash`"*.

### 10. "every `~/.hermes/` edit under `flock`" vs §2, which writes without it — INCONSISTENT

> L274: "`.bak` first, every `~/.hermes/` edit under `flock ~/.hermes/.wf3.lock`."

§2 (L141-155) creates `~/.hermes/scripts/kb-refresh.sh` with neither. `CLAUDE.md`'s hard
rule is scoped: *"Every `~/.hermes/config.yaml` or `.env` edit"*. **Correction.** L274 →
"`.bak` first, and every edit to a **live-reloaded** file (`config.yaml`, `.env`,
`SOUL.md`) under `flock ~/.hermes/.wf3.lock`" — a new file in a new directory has no
concurrent writer.

### 11. Rollback order pages Gaetan for nothing — MINOR

> L286-287: "`rm ~/.hermes/scripts/kb-refresh.sh`, `hermes cron remove <job_id>`"

A tick between the two fires a missing script ⇒ non-zero exit ⇒ error alert to Slack
(the very delivery §Verify sets up). **Correction.** Swap: `hermes cron remove <job_id>`
first, then `rm`.

### 12. "`done-is-not-exercised` carries **no** signal" — overstated

> L113. Source says "much less signal" (`kb-retrieval-trials.md:807-808`); P6's own L242
> says "little signal". **Correction.** "little signal", consistently.

### 13. `docs/proposals/README.md` P6 row is pre-overturn — CONTRADICTED

> README L14: "| [P6](P6-knowledge-bases.md) | … | **one SOUL line + a generated `SKILL.md`
> + a cron job** — not while P3/P4 are being applied | yes, but **serialize** |"
> README L35: "P6 writes under `~/.hermes/` (`SOUL.md`, **`skills/`**, `scripts/`, a cron job)"

Both still describe the installed map. Row is otherwise correct and **not duplicated**
(one P6 row, one reading-order entry); the reading-order paragraph (L23-27) survives the
rewrite unchanged and is accurate. **Correction.** L14 needs → "one SOUL line + a refresh
script + a cron job (skill map **deferred** — see §0/P6-f)"; L35 → drop `skills/`.

### 14. §3's "Tars never authors it" vs the new `CLAUDE.md` skill-mirror carve-out — no contradiction, but scope the wording

> L180: "**Tars never authors it** — SOUL rule 1 (P4) forbids Tars producing documentation
> *'not because it's only markdown'*."

`CLAUDE.md` (peer-updated) now reads: *"Tars never merges, approves or pushes — one scoped
exception: its own skill mirror, which SOUL rule 2 has it update by PR + automatic
squash-merge … A PR authored by a session Tars delegated to is fine; Tars authoring one for
anything but its own skills is not."* Verified against the **live** SOUL (read-only): rule 1
= "I never produce the deliverable myself … not because 'it's only markdown'"; rule 2 = "I
never merge, approve or push — with exactly one exception … **This exception covers my own
skills and nothing else.** It is not licence to push in a repo I was delegated work in —
there, rule 2 stands whole." So P6's citations are accurate and the exception is
categorically outside §3's subject (a `learnings/` note in a team repo). **Only the bare
absolute needs narrowing** — one word plus an optional parenthetical:

- L180: "**Tars never authors it**" → "**Tars never authors the note**"
- optional, same line: "(rule 2's one exception is Tars' own `skills/<n>/SKILL.md` mirror,
  self-merged — it does not reach mc-metarepo, and rule 2 'stands whole' in a repo Tars
  delegated work in)"

No other P6 sentence overstates: L209-210 ("reading a PR is verification, rule 2's own
carve-out"), L220 ("a merged PR is an incident"), L303 (P6-e "Tars **asks**") and L231
("a push NEVER goes through the VM's `~/dev/mc-metarepo`") all match the live rules.

---

## Checked and SUPPORTED (no change)

15. **§0 and scoreboard numbers vs `kb-confound-trial.md` — exact, all four cells.**

| P6 | `kb-confound-trial.md` |
|---|---|
| L54 "(c⁻) … **CORRECT** — 23 s, 958 B, 2 tools" / "**CORRECT**, incl. the 'do not erase logs' caveat — 38 s, 4,808 B, 3 tools" | L184 "\| **(c-minus)** \| Q1 \| **CORRECT** \| **23** \| **958** \| **2**"; L187 "Q3 \| **CORRECT** \| **38** \| **4,808** \| **3**"; L166-169 "it states the forbidden action as forbidden ('rotate the credential; do not erase the logs')" |
| L55 "(c) map + `read_file` \| CORRECT — 19 s, 958 B, 1 tool \| CORRECT — 21 s, 1,717 B, 2 tools" | L185 "(c) \| Q1 \| CORRECT \| 19 \| 958 \| 1"; L188 "(c) \| Q3 \| CORRECT \| 21 \| 1,717 \| 2" |
| L26 scoreboard "(c⁻) ✓ **958 B / 23 s** live \| — \| ✓ **4,808 B / 38 s** live \| — \| — \| **2/2 on the two failures**" | same figures; Q2/Q4/Q5 correctly left blank — "**Q2, Q4, Q5 under (c-minus)** … Only the two failing baseline questions were re-fired" (§Not tested) |
| L58 "2.8× fewer tool bytes, 17 s, n=1 … 3.3 KB artifact" | L204-207 "2.8× fewer tool-output bytes (1,717 vs 4,808) and **17 s** (21 vs 38)"; map = 3,279 B |
| L57 "**Every correctness win credited to the map belongs to the pointer fix**" | L200-203 "Every correctness win previously credited to mechanism (c) on Q1 and Q3 is attributable to the pointer" |

16. **Rows (a)/(b)/(b′)/(c)/(d)/(e) and hit rates** — match `kb-retrieval-trials.md`
    §Scoreboard cell for cell; the `✓*` marks and "5/5 — 2 measured, 3 reasoned" discharge
    `kb-verify-claims.md` #11 and `kb-verify-coverage.md` #2. The simplicity axis (L32-34)
    discharges coverage's "rank by simplicity" gap.
17. **The overturn is argued honestly.** §0 names the confound, names the review that found
    it, keeps the map's efficiency delta rather than burying it, and states the proxy limit
    in §0 itself (L60-63) *and* in §Where-this-fails (L235-238) — not buried. The deferral
    is falsifiable: §1 names the one experiment, §Verify names the condition that puts the
    map back on the table (L282-283).
18. **No leftover "the map is installed" sentence.** Every map mention is conditional:
    L6 "(deferred, see §1)", L103 "**Deferred:**", L165 "*Only if the map is later adopted
    (§1)*", L244-246 "if the map is later adopted", L288 "If the map was adopted". Order
    (L256-269) creates no skill dir — §2's block is `mkdir -p ~/.hermes/scripts` only, and
    step (3)/(4)/(5) are script → hand-run → cron. Rollback matches the reduced scope.
19. **Live state re-verified today** (supports L65-71): `grep -ic mc-metarepo
    ~/.hermes/SOUL.md` → `0`; `ls -d ~/.hermes/scripts` → "No such file or directory"
    (supports L162); `ls ~/.hermes/skills` → 17 dirs, no `mc-metarepo` (supports the
    deferral being un-applied). The pointer is now at **line 88**, `SOUL.md` 5,708 B,
    mtime 23:36 — i.e. it moved again since the "line 81" note, which **vindicates** L69-71
    ("Do not trust a line number here … Locate it at apply time"). Only cosmetic: L70 and
    the comment at L265 should read "54, then 81, then 88 — it moves".
20. **Refresh evidence** — `fetch`+`reset --hard` as the only exit-0-everywhere strategy,
    `git pull` stuck forever, `--rebase` leaving `.git/rebase-merge` under an exit-0
    `reset`, the `timeout` need, `GIT_TERMINAL_PROMPT=0` — all match `kb-refresh.md` §4-5.
    The hedges added by the last pass are all present and correctly worded: doc-only error
    alert (L125-128), artifact observed for agent-mode only (L124), restart-survival not
    exercised (L123-124), `--script` basename ambiguity + retry (L160-162), "the invocation
    itself has never been run" (L159).
21. **Push section** — the four mc-metarepo rules (STE/`ste-lint`, `learnings/` rule 5 PII,
    three `knowledge/` surfaces, `--repo`/`-C` cwd trap + the 2026-08-05 wrong-merge) and
    the four `gh pr` assertions match `kb-mc-metarepo.md` and `kb-push-design.md` §5;
    "MERGED = incident" is verbatim from §5's table.
22. **Mandate still fully answered after the cuts.** §1 search (mechanisms measured + ranked
    by simplicity, incl. "install nothing" candidates); §2 hourly refresh (`hermes cron`
    vs systemd `--user` timer vs crontab — the comparison survives as one clause at
    L126-128 plus `kb-refresh.md` §1-3, thin but decided, with the deciding feature pushed
    into §Verify; who owns the pull; dirty/conflict/rewritten-history/auth/network; "is
    Tars told?" answered explicitly at L172-176); §3 push designed-not-built (brief shape,
    repo/paths, PR-not-branch, who reviews, how Tars verifies, noise brakes); sources in/out
    at L86-93 (mc-metarepo in; `gaetan-metarepo` out; Notion no-change, correctly hedged to
    "registered … registry-only" per `docs/facts.md:12`; 10 submodules out); four binding
    constraints all present (author rule L180; team-repo rules L189-207; context budget
    L80-82 — 1.44 MB vs the 0.85 trim threshold, `docs/facts.md:83`; Hindsight decision
    L79-81 + the withdrawn-finding paragraph L305-308).

---

## Verdict

**FIX FIRST** — 2 blocking, 4 major, 7 minor. The rewrite is internally consistent about
the deferral and its numbers are exact; the defects are in the apply step and in four
stale labels/rows.

Ordered minimal fix list:

1. **§Order step (2): delete the `sed`.** Add a sentence instead of renaming the Phase-2
   preferences sentence (finding 1).
2. **Write the literal SOUL sentence in §1** — "on this machine" + "submodules are empty"
   (finding 2).
3. **P3/P4 have landed**: fix L179, L9, L257-258, and change P6-a's recommendation to
   **standalone** (finding 3).
4. **L293 "P6-a…P6-e" → "P6-a…P6-f"**, and move the P6-f row (finding 4).
5. **L114**: "~3 KB on the one question where it helped (Q3), nothing measurable on Q1,
   n=1" + "eight-line tail", not "an hourly regeneration job" (finding 5).
6. **§1**: one clause conceding that map-less `search_files` reached the Q3 note on meaning
   (finding 6).
7. **§Verify**: scope "~40 s" to Q1/Q3 (finding 7); **§0**: state "2 questions, one run
   each, no repeats" (finding 8).
8. **`docs/proposals/README.md`**: P6 row → "one SOUL line + a refresh script + a cron job
   (skill map deferred)"; drop `skills/` from L35 (finding 13).
9. **L180**: "Tars never authors **the note**" + the skills-mirror parenthetical
   (finding 14).
10. Minor, batch: L274 flock scope (10), rollback order (11), "little signal" (12), line
    "54 → 81 → 88" (19), optional `rev-parse --abbrev-ref HEAD` post-check in the script (9).
