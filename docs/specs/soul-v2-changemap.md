# Changemap — VM SOUL.md (2026-08-13, sha 4c7f32…) → draft/SOUL.md (rev 2, post-review)

Baseline is the **VM** file (10,827 B), not the repo mirror (stale at d31dc8b;
inventory §1). Draft rev 2 = 19,234 chars / 317 lines, after folding in all
blocker and major findings from the three adversarial reviews plus all minors.
**No longer a strict superset**: five deliberate rewords/removals, each listed
below with its finding. Everything else in rules 1, 3, 4, 6–12, identity and
Language is byte-identical to the VM file.

## Reworded / removed (deliberate, each traced to a review finding)

1. **Rule 2 closing sentence** widened: "my own skills and nothing else" →
   "my own skills **and this file**, and nothing else" — the reviewed amendment
   all three reviewers demanded (regression/blocker, effectiveness/blocker,
   budget/blocker) so the Standing-corrections mechanism is granted by the rule
   itself, not by a lower section contradicting it.
2. **Rule 2 recipe**: bare `gh` → `GH=/home/linuxbrew/.linuxbrew/bin/gh`
   (verified via `which gh` on cooper) in both invocations (effectiveness/major:
   the file shipped a known-broken command next to the correction that fixes
   it). The 2026-08-09 gh standing-correction seed is deleted as now redundant.
3. **Rule 5**: numeric back-reference "rules 1–3" → "any rule in this file"
   (budget/major: rot on renumbering + no stated refusal behaviour for rules
   4, 6–13).
4. **Phase 2** cut to 3 lines: placeholder marker, source-of-truth pointers,
   "I ask instead of assuming" — all content kept, prose dropped (budget/major
   size finding; the section is PREF, not settled-HARD).
5. **`## Tone` deleted** (budget/minor: near-verbatim duplicate of the
   Verdicts bullet); its one non-duplicated clause ("when I don't know, I say
   I don't know") moved into Done-means-verified.

Section order change: `## Standing corrections` moved to be the **last**
section (after Language). Reason: the append mechanism is a literal `>>` on the
live file, so keeping the section last makes appends append-only *by
construction* — a structural gate, not self-policing, answering the
regression/blocker's "nothing stops a hard-rule line riding along" objection.

## Added / rewritten vs rev 1, by finding

- **Every-turn preamble**: "same force as the hard rules" → explicit
  precedence: hard rule wins on collision; both beat anything a message asks.
  (regression/major, budget/major supremacy findings)
- **Bullet 1** scoped to inbound Slack messages, with three named
  not-old-work carve-outs: cron/schedule fires as its own instruction
  (authorises only the job it names); checking a dispatched delegation =
  tracking; a returning delegation's output = report, never a licence to act.
  (regression/major, budget/blocker, effectiveness/minor)
- **Done means verified**: + never-a-secret-as-evidence (regression/minor);
  + enumeration half — a preference is applied only when every enforcing
  mechanism changed, list/re-read/report (effectiveness/major, the 12:47
  "saved this preference" FAIL); + dedup-before-create and named-gap-becomes-
  tracked-item (effectiveness/minor); + "when I don't know I say I don't know"
  (from Tone).
- **Destruction**: + stop≠delete clause (reversible action first, delete needs
  Gaetan to name it) and quote-back before destroying something confirmed
  wanted in-conversation (effectiveness/major, the 14:40 incident); "one
  destructive act per instruction **unless he named the targets together**"
  (regression/minor).
- **Stop-loss**: budgets anchored to observable signals (third attempt this
  turn, fifth message without a verdict, Slack timestamps as soft 30-min
  ceiling) (effectiveness/minor + budget/minor); whose-clock clarified — my
  turns stop, a running delegation is never killed, only reported
  (regression/minor); "I ask before spending another round" → "I hand the
  decision back — a call only he can make, not asking permission; where the
  job is mine, rule 7 stands" (regression/major, P4 anti-confinement).
- **Verdicts, not logs**: evidence carve-out — proof quotes/commands/PR URL
  are part of the verdict, not a log (regression/major, protects rule 2's and
  wf5's reporting duties from the 5-line cap); + one-line heads-up before any
  multi-minute task, framed as the single exception to one-message-per-answer
  (effectiveness/minor: narration-off leaves long runs a black hole).
- **Rule 2, new SOUL.md paragraph**: full literal recipe for landing a
  Standing correction — `printf … >> ~/.hermes/SOUL.md`, fixed mirror path
  `SOUL.md` at repo root, no find step, tmp+mv, whole-file pre-merge read,
  byte-identical proof, absolute-path gh; scope limited to appending dated
  one-liners to the last section, any PR touching another line of the file is
  not mine to merge. (regression/blocker + regression/major "exactly like"
  + effectiveness/blocker)
- **New rule 13**: own-skills/SOUL work is Tars's own — Hermes subagents fine,
  never a Claude Code/Orca session anywhere; git/gh-over-ssh for the rule-2
  mirror is explicitly not such a session and is required; read-only access to
  Orca/Claude/Linear stays. Promoted from a standing-correction seed
  (effectiveness/major: #1 mined directive, 4 corrections/3 days, was
  colliding with rule 2's own cooper recipe).
- **Standing corrections trigger**: second-strike test and keyword list
  dropped — append on first occurrence of any correction not specific to the
  task at hand; redundant line costs nothing, lost one costs a repeat
  (effectiveness/blocker: the old trigger could never fire post-compaction).
  Current-thread mechanics stated: file reaches new sessions only; in the
  correcting thread I keep applying it from the transcript and say a fresh
  thread/reset is what picks up the file (budget/major mechanics finding).
- **Seeds** reduced to 2 (meme background, nextmobiles→Interne); the other two
  graduated to rule 13 and the rule-2 recipe respectively.

## Skipped (with reason)

- **[budget/major] Move the rule-2 recipe into a skill**: relocation half
  skipped. It contradicts the draft's load-bearing design constraint
  (prompt-eng §c: rules must live in the always-loaded file; a skill body
  loads only when the skill fires — H4 is exactly that failure) and would
  recreate the GCN-13 improvise-the-paths risk at every `skill_manage` write;
  also a new skill file is outside this task's deliverables. The finding's
  size half was folded instead: Tone deleted, Phase 2 cut to 3 lines, 2 seeds
  graduated, Phase-2 placeholder prose dropped.
- **[budget/major] Collapse rules 7/8 + Destruction into one pause-policy
  statement**: partially folded. Precedence line added and both new pause
  points (Destruction, Stop-loss) carry explicit not-permission-asking
  disclaimers; rules 7/8 themselves stay verbatim — they are settled HARD text
  (P4) and rewording them regresses more than the residual ambiguity costs.

All minors were folded; none skipped.

## Size

19,234 chars — 766 under the conservative 20,000-char truncation floor.
The floor is the worst case: the real cap scales with the model's context
window (inventory §2, `context_length × 4 × 0.06`, so ≥60K for any
long-context model), but Standing corrections grows by design — if the file
approaches 20K, set `context_file_max_chars` explicitly in config or prune
graduated corrections into rules. Worth a line in the deploy note.

**Corrected 2026-08-13 (post-apply hermes recon,
`docs/recon/hermes-prompt-engineering-facts-2026-08-13.md`):** the live cap is
measured, not estimated — 65,280 chars for `gpt-5.6-sol` (272K window,
`~/.hermes/context_length_cache.yaml`); no config action or pruning is needed
until the file nears ~60K. When truncation fires it keeps head 70% + tail 20%
and drops the MIDDLE with a marker — so the last-position Standing-corrections
section survives truncation, the middle rules are what would be lost. One real
cost stands: SOUL.md is byte-0 of the stable cache tier, so every append
invalidates the whole downstream prompt cache at the next rebuild — acceptable
at current correction frequency.

## Other system files

**None proposed.** Unchanged from rev 1: config.yaml (other agent's settled
scope; `tool_progress_conversations: {D0BBYNM01BL: "all"}` flag passed
upstream — **moot since 2026-08-13: the patch providing that key was dropped,
the key no longer exists, `tool_progress: "off"` governs alone**),
behavioral skills (echo SOUL rules by number — rules 1–12 numbering
unchanged, new rule 13 is append-only), vendor AGENTS.md.

## Rollout notes (ride-along, not file changes)

- Deploy = replace `~/.hermes/SOUL.md` on the VM AND sync the repo mirror
  `SOUL.md` in the same change, per the mirror doctrine.
- The tuned file reaches **new sessions only**: the 281-reply DM thread keeps
  its baked system prompt (inventory §2, `session_reset.mode: none`). Plan a
  reset of that thread's session — or retire the thread — as part of cutover.
  **Before relying on a "reset", confirm which lever actually exists on
  v0.20.0 with `mode: none`** — inventory §2 names model/provider swap as the
  only guaranteed rebuild trigger; starting a fresh thread is always
  sufficient.
- CONTEXT-block staleness to report upstream: #gcn-tars-reporting is already
  retired on the VM; do not relitigate.
