# Probe — P6 mandate-coverage audit (harsh, read-only)

Reviewed 2026-08-07. Inputs: `docs/plans/knowledge-bases-proposal.md` (mandate),
`docs/proposals/P6-knowledge-bases.md` (deliverable, 190 lines),
`P3-orca-v2-skill.md` + `P5-a2a.md` (house format), and the six `kb-*.md` evidence
files. Nothing was edited; no ssh, no VM access. Line refs are P6 line numbers.

**Headline:** the design is right, lazy, and measured. The proposal *under-reports its
own evidence*: the push section drops four binding rules of the target repo that its
own probe file quotes verbatim, and the scoreboard never tests the single laziest
candidate (fix the dead SOUL pointer, install nothing).

## 1 — Coverage table

### Mandate §1 — retrieval mechanisms actually available (measure, rank by simplicity)

| Mandate item | Verdict | P6 line | What to add |
|---|---|---|---|
| shell tool over the clone (ripgrep / git grep) | COVERED | L17–18 (rows a, b, b′) | — |
| a skill that teaches Tars where things are | COVERED | L19 (row c), §1 L52–61 | — |
| an MCP server (filesystem / docs / RAG) | PARTIAL | L20 (row e), L38–40 | Row (e) is "on paper" for all 5 questions. Justified — `node`/`npx` absent, measured — but say "not runnable on this VM, so scored by inspection" *in the table cell*, not only in §What-is-NOT-proposed. |
| anything Hermes ships natively (`--help`, source, config keys, `toolsets:`) | COVERED | L38–43 (`_HERMES_CORE_TOOLS`, `lcm_*`, §7 "already live") | — |
| **rank by simplicity — the laziest thing that answers wins** | **PARTIAL** | L14–21 ranks by hit rate only | Add the simplicity axis (artifacts installed: b′ = 0, c = 1 skill + 1 script + 1 cron job) and, critically, the missing row below. |
| **the laziest candidate was never tested** | **MISSING** | — | Mechanism (c) bundles TWO changes: the SOUL pointer fix *and* the map. The map text itself names `~/dev/mc-metarepo`, so every (c) win is confounded with the L31–33 blocking finding. **"corrected SOUL line, no skill" has no row in the scoreboard.** Either measure it (2 live turns, Q1 + Q3) or state the confound in one line and let Gaetan decide whether to buy the map on 2 live data points. The decision box (L188) offers "search only (map + SOUL line)" but not "SOUL line only, re-measure" — the cheapest option is not on the ballot. |

### Mandate §2 — refresh

| Mandate item | Verdict | P6 line | What to add |
|---|---|---|---|
| `hermes cron` vs systemd `--user` timer vs plain cron | COVERED | L63–75 (+ `kb-refresh.md` §1,§6) | — |
| who owns the pull | COVERED | L101 (`cron create --no-agent --script`) | — |
| failure behaviour: dirty tree / conflict / rewritten history | COVERED | L69–74, script L86–90 | — |
| failure behaviour: auth expiry, network | COVERED | L74–75 (`GIT_TERMINAL_PROMPT=0`, `timeout 120`) | — |
| **is Tars *told* a refresh happened, or does it benefit silently?** | **PARTIAL** | L104–105 answers *Gaetan's* channel ("silent on success, alerted on failure") | The mandate asks about **Tars**, not Gaetan. Add one line: "Tars is never told — the map is a file it reads on demand; staleness surfaces to Gaetan's Slack, and Tars can be wrong for up to one hour without knowing it." |
| regeneration failure is unhandled | MISSING (script defect) | L91–98 | If `SKILL.head.md` is missing or `ls`/`paste` fails, the `{ … } > "$S/.new"` block still produces a partial file, `cmp` differs, `mv` installs it, and the tick exits **0 → silent**. Add `[ -s "$S/.new" ] \|\| { echo "KB refresh FAILED: map regen"; exit 1; }` before the `cmp`. One line, and it is the only path where the design fails quietly. |

### Mandate §3 — what "search" must actually answer (3–5 real questions, show outputs)

| Mandate item | Verdict | P6 line | What to add |
|---|---|---|---|
| 3–5 real questions, ground truth first | COVERED | L14, `kb-retrieval-trials.md` (ground truth read before trials) | — |
| **each candidate mechanism, outputs shown** | **PARTIAL** | L14–21, honesty note L26–29 | Winner (c) = **2 fired live, 3 reasoned from file sizes**; (e) = 0 live. The prose says so; the table does not. Mark the 3 estimated cells (`✓*`) and footnote `* estimated, not fired`. As printed, a reader scanning the table sees a 5/5 measured result. |
| a design that looks elegant but answers badly is a fail | COVERED | L22–29 (baseline dissection: 422 s + Datadog prod + a stray worktree to miss a 767-B note) | — |

### Mandate §4 — the push phase (designed, not built)

| Mandate item | Verdict | P6 line | What to add |
|---|---|---|---|
| brief shape | PARTIAL | L127–129 (one-sentence paraphrase; literal text in `kb-push-design.md` §2) | House format is "exact commands or config". Inline the 6-line brief skeleton, or at minimum the `worker-start … --spec "$(cat /tmp/tars-brief-<slug>.md)"` invocation. Today P6 §3 contains **zero commands** while §2 contains a full script — the asymmetry reads as a thinner design than it is. |
| which repo / paths | COVERED | L130–132, P6-f L181 | — |
| PR vs branch | COVERED | L120–124 (PR-only, quoted + branch protection) | — |
| who reviews | COVERED | L121–124 (human merge, `required_approving_review_count: 1`, `enforce_admins: true`) | — |
| **how Tars verifies the result** | **MISSING** | L114 says only "verifies the PR" | `kb-push-design.md` §5 has four `gh pr view --json …` assertions incl. the path-scope check and `state: MERGED` = incident. Two lines of that belong in the proposal — verification is the half of the design that is *Tars'* job. |
| what stops it becoming a slow leak of noise | COVERED | L115–118 (human-only trigger, periodic harvest rejected), P6-g L182 | — |

### Mandate — the four binding constraints

| Constraint | Surfaced? | P6 line | Gap |
|---|---|---|---|
| **Tars does not author the deliverable** (P4 rule 1, *"not because it's only markdown"*) | COVERED — quoted, with the quoting-Gaetan carve-out named | L111–114 | — |
| **mc-metarepo is a team repo; honour its own rules incl. PR-only / directory-scoped** | **PARTIAL** | L120–124 quotes `learnings/README.md` rule 1 + branch protection; P6-f raises directory scope | Four rules the probe quotes verbatim never reach the proposal: (a) `CONTRIBUTING.md` §Language — **English + ASD-STE100, CI-gated (`ste-lint` fails on an STE error in an added line)**; (b) `learnings/` rule 5 — **no PII, no secrets, ever** (the material is a Slack quote; this is the rule most likely to bite); (c) the write-rule table — a `knowledge/` note is incomplete without its `@import` line in `CLAUDE.md` **and** its row in `knowledge/README.md`, which is the actual cost of P6-f's "yes"; (d) the `--repo`/cwd-reset trap that **already merged the wrong PR on 2026-08-05**. One four-item list, ~4 lines. Without them P6-f's recommendation looks cheaper than it is. |
| **context budget — no prompt stuffing** | COVERED | L41–42 (1.44 MB ≈ 370 K tokens vs 0.85 trim threshold vs ~1 KB/question), L60 (map = 0.22 %) | — |
| **Hindsight skipped; a memory/RAG provider is a decision, not an assumption** | COVERED — and better: it caught the live config contradicting the mandate | L40–41, P6-b L178 | — |

### Mandate — sources in scope / out of scope

| Source | Verdict | P6 line |
|---|---|---|
| mc-metarepo (in: search + refresh now, push later) | COVERED | L45–46 |
| gaetan-metarepo (out, and *why*) | COVERED | L46–48 |
| Notion (out — already live via MCP, no change) | COVERED | L48 |
| the 10 submodules (out, with the reason made load-bearing) | COVERED | L49–50, P6-c |

### House format (P1–P5 shape)

| Section | Present | Note |
|---|---|---|
| Status header (status / source / evidence / decide) | yes | L3–6 |
| "What changes" | PARTIAL | No single heading; §1/§2/§3 carry it and §"What is NOT proposed" (L35–43) does the negative half well. P3 leads with "What changes, and why v2 is smaller than v1" — P6 leads with the scoreboard, which is a defensible inversion. Not a required fix. |
| Exact commands / config | PARTIAL | §2 is exact and runnable. **The SOUL edit is not**: L158–159 says "drop the pointer, name `~/dev/mc-metarepo`" with no literal before/after line and no `flock`/`.bak` command, although it is step 2 of Order and the L31–33 blocking finding. §3 has no commands (defensible — designed, not built — see above). |
| Order | yes | L155–161 |
| Restart | yes | L162–163 |
| Verification | yes | L164–169 — strong: names Q3 as the discriminator and requires one Slack fire |
| Risks / where it fails | yes | L134–152 — unusually honest (auto-injection untested, 21 KB governance questions, `reset --hard` forever) |
| Rollback | PARTIAL | L170–171 leaves `~/.hermes/scripts/kb-refresh.sh` behind. Add `rm ~/.hermes/scripts/kb-refresh.sh`. Trivial, but rollback lists are read under stress. |
| Open decisions table | PARTIAL | L175–182 runs **a, b, c, e, f, g — there is no P6-d.** Either a dropped decision or a renumber artifact; a reviewer cannot tell which. Worse: `kb-push-design.md` L509–514 has its **own** P6-a…P6-f meaning different things, and P6-e (L180) cites "kb-push-design.md P6-a/b" — two live namespaces, same labels. Renumber the evidence file's to P6-push-a… or add "(push-design numbering)". |
| Decision box | PARTIAL | L186–191: no checkbox for P6-g, and no "SOUL line only, then re-measure" option (see §1). |

### Presented as done that was only proposed?

**Nothing material.** Header L3 says "proposed, NOT applied"; L6 warns against concurrent
application; L61 says the map is applied "minus its two generated lists"; L106 admits
`cron create --script` dir creation is untested; L135–137 admits skill auto-injection was
never installed. Cross-checked against `kb-refresh.md` (`--no-agent`/`--script`/`--deliver`
all verbatim from `hermes cron create --help`), `kb-mc-metarepo.md` (branch protection from
a real `gh api` read), `kb-retrieval-trials.md` (SOUL:54 from a real `grep`). Claims match
evidence.

Two soft over-claims only: the ✓ marks for (c) Q2/Q4/Q5 are inferences printed like
measurements (fix above), and L19's "**958 B / 19 s** live" bolding next to unbolded
estimates is the only signal distinguishing them.

## 2 — Cut this (over-engineering)

The design is already lazy where it counts, and the refusals are the best part of it:
**no vector DB, no MCP server, no memory provider, no `config.yaml` edit, no prompt
stuffing** (L35–43) — all argued from measurement, not taste. What is left to cut is small
and all inside the script:

| Cut | Lines | Why |
|---|---|---|
| `cmp -s "$S/.new" "$S/SKILL.md" \|\| mv …` + `rm -f "$S/.new"` + the `.new` temp (L96–98) | 3 → 1 | Write straight to `SKILL.md`. The only thing it buys is not bumping mtime in the path→(mtime,size) skills manifest — i.e. it avoids re-indexing a 4 KB file once an hour. That is not a cost. `} > "$S/SKILL.md"` and delete two lines. (Counter-argument, if kept: it makes "did the KB actually change?" observable via mtime — which the verify step at L169 relies on. Then keep it and say *that*, because the current text says "rewrite only on change" as if the write were expensive.) |
| the `learnings/` sidecar regeneration pipeline (L94–95) | 3 → 0 | 8 entries (`cleaq/risk`, `mobileclub/{4}`, `next/{2}`) that change on the order of never, and L140–142 already concedes slug routing does not work for `learnings/` anyway. Hardcode the list in `SKILL.head.md`. |
| `date -u -r "$R/.git/FETCH_HEAD" …` inside the failure string (L89) | ~1 | The alert already says STALE and the tick timestamp is in `cron runs`. Cosmetic sub-shell in an error path. |
| `~/.hermes/scripts/` + a script file at all | keep | Justified by `--help`: `--script` requires a path under `~/.hermes/scripts/`. Not optional. |
| the generated map / `SKILL.head.md` split | keep | Earns its place: the repo's own `knowledge/README.md` indexes 35 of 40 notes and the Q3 note is one of the 5 missing (L57–58). A hand-written list would rot the same way. |

Net: the script goes from ~20 lines to ~14, and gains the one guard it is missing
(regen failure, §2 table above). Nothing else in P6 is worth deleting.

## 3 — Verdict

**FIX FIRST** — 8 MISSING/blocking items, none of which change the design. This is a
day's-end polish pass on a proposal whose shape and evidence are sound, not a rework.
Every fix below is 1–4 lines in P6 itself; no new probing is required except #1, which
can also be discharged by stating the confound.

Ordered required fixes:

1. **Add the laziest candidate to the scoreboard or name the confound.** "Corrected
   SOUL line, no skill" is untested and mechanism (c) bundles both changes. Either fire
   Q1 + Q3 that way (2 live turns) or add one line: "(c)'s wins include the SOUL fix; the
   pointer fix alone was not measured separately." Add the matching decision-box option.
2. **Mark the 3 estimated cells** in the L14–21 table (`✓*` + footnote), so the table
   carries the honesty that currently lives only in the prose below it.
3. **Surface the four dropped mc-metarepo rules** in §3, as one list: English +
   ASD-STE100 (CI-gated), `learnings/` rule 5 no-PII/no-secrets, `knowledge/` ⇒
   `@import` + README row (the real cost of P6-f), and the `--repo`/cwd trap that merged
   the wrong PR on 2026-08-05.
4. **Add "how Tars verifies"** to §3 — the `gh pr view --json number,url,state,labels,files`
   assertions, the path-scope check, and `state: MERGED` = incident.
5. **Add the regen guard** to `kb-refresh.sh`: `[ -s "$S/.new" ] || { echo "KB refresh
   FAILED: map regen"; exit 1; }`. It closes the design's only silent-failure path.
6. **Write the SOUL edit literally** — the current line 54, the replacement line, and the
   `.bak` + `flock` command, since it is step 2 of Order and the blocking finding.
7. **Fix the decision numbering**: no P6-d exists; `kb-push-design.md` runs a colliding
   P6-a…f namespace. Renumber one of them, and add the P6-g checkbox to the decision box.
8. **Answer "is Tars told?"** in one line (it is not — up to 1 h of unknowing staleness),
   and add `rm ~/.hermes/scripts/kb-refresh.sh` to Rollback.

Not required, noted: §"What changes" is folded into the measurement lead-in rather than
given its own heading, unlike P3. Defensible for a measurement-led proposal — leave it.
