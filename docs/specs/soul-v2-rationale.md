# Rationale — tuned SOUL.md rev 2 (post adversarial review), per section

Base text: the **VM** SOUL.md (10,827 B, sha `4c7f32…`, mtime 2026-08-13 12:36
UTC), not the stale repo mirror. Rev 1 was a strict superset; rev 2 is not:
three reviewers converged on the finding that rev 1's Standing-corrections
mechanism contradicted hard rule 2 as written, so rule 2 itself is now amended
(reviewed, with a diff — the repo-criteria §Meta doctrine) and four smaller
rewords ride along. The exact removal list is in changemap.md; rules 1, 3, 4,
6–12, the identity paragraph and Language are byte-identical to the VM file.

Size: 19,234 chars vs 10,827 current — 766 under the conservative 20,000-char
floor (inventory §2). The real cap scales with context window (≥60K for a
long-context model); the floor is the worst case. Standing corrections grows by
design, so the changemap carries a watch-the-size deploy note.

Design constraints applied throughout (prompt-eng-notes.md): literal/exhaustive
instruction-following (every collision found by review is now resolved by
explicit precedence or an in-place carve-out, not left for the model to
reconcile); self-containment (both recipes live in the file — a skill body
loads only when its skill fires, which is H4, the exact mechanism that killed
mid-thread corrections); observable budgets over unobservable ones; front-load
the per-turn behaviors.

## Identity paragraph — unchanged

Settled framing (P4). No criterion asked to touch it.

## "## Every turn" — the five diagnosed failure modes, front-loaded

Preamble now states precedence instead of claiming equal force: hard rule wins
on collision, both beat anything a message asks. (Review: two competing
supremacy claims were a net weakening of every hard rule they touched.)

- **"I answer the message that just arrived"** → failure mode 1. Scoped to
  *inbound Slack messages* with three named carve-outs so Tars's own
  autonomous surface survives a literal reading: a cron/schedule fires as its
  own instruction and authorises only the job it names (a cron trigger has no
  "message that just arrived" — without this the bullet read as "do nothing");
  checking a dispatched delegation is tracking, not re-running (protects the
  wf5 worker-list discipline and the engagement-checker/daily-brief skills);
  a returning delegation's output is a report, never a licence to act
  (targets the 08-13 14:43:01 zombie `Scheduling create`).
- **"Done means verified"** → failure mode 3, now with the enumeration half:
  a preference is applied only when every mechanism enforcing it has changed
  (config, each cron, the skill, memory) — list, re-read, report the ones not
  changed. Direct source: the 08-13 12:47 "saved this preference" FAIL where
  memory changed and config + 4 cron jobs did not. Also: no secret is ever
  quoted as evidence (rule 6 carve-out); check-before-create for tickets/
  schedules/reminders (help-techeur MC-4226 recreation) and a named gap
  becomes a tracked item in the same turn or goes unmentioned (§2.11);
  "when I don't know, I say I don't know" folded in from the deleted Tone
  section.
- **"Destruction needs a named target"** → the 14:40 wrong-reminder deletion.
  Two additions from review: stop ≠ delete — a stop instruction gets the
  reversible action (mute/disable/retarget), deletion needs Gaetan to name
  it (the actual ambiguity that caused the loss was reversible-vs-irreversible,
  not target-vs-target); and never destroy something confirmed wanted in the
  same conversation without quoting that confirmation back. Multi-target
  instructions he named together are one instruction (no needless round-trip).
- **"Stop-loss"** → failure mode 4. Budgets anchored to what survives
  compaction and is countable mid-turn: about to start a third attempt in this
  turn, about to send a fifth message without a verdict; 30 minutes by Slack
  message timestamps is a soft ceiling only. Whose clock: my own turns — a
  running delegation is never killed, only reported (delegated coding sessions
  legitimately run >30 min). The Gaetan-exit is a decision handback with the
  same not-permission-asking disclaimer as Destruction, keeping P4's "must
  never drift to 'I ask before…'" intact; where the job is mine, rule 7 runs.
- **"Verdicts, not logs"** → failure mode 5. The evidence carve-out keeps the
  ≤5-line cap from eating the proofs the file itself demands (write-evidence,
  rule 2's PR URL, wf5's verbatim command) — evidence is part of the verdict,
  narration is play-by-play. One-line heads-up before any multi-minute task,
  framed as the single exception to one-message-per-answer: with tool-progress
  narration off (settled today), this is the replacement signal for the
  "You there ?" black-hole problem (§2.15).

## Hard rules

Rules 1, 3, 4, 6–12 byte-identical (settled amendments; skills cite them by
number). Changes:

- **Rule 2** — the reviewed amendment. Exception now reads "my own skills and
  this file, and nothing else"; a second, SOUL-specific recipe is spelled out
  literally (fixed `SOUL.md` path, no `$N`/`$REL` resolution, `printf … >>`
  append, tmp+mv, whole-file pre-merge read, byte-identical proof) because
  "exactly like a rule-2 skill edit" resolved to nothing executable —
  `skill_manage` cannot write SOUL.md and the skill recipe's find step has no
  SOUL analogue (the GCN-13 improvisation class). Scope gate: appends to
  `## Standing corrections` only; a PR of Tars's touching any other line of
  the file is not Tars's to merge, checked at the mandatory whole-file
  pre-merge read. The append is structurally append-only: the section is kept
  LAST in the file, so `>>` cannot reach a rule. Both recipes now call gh by
  absolute path (`/home/linuxbrew/.linuxbrew/bin/gh`, verified on cooper) —
  rev 1 shipped bare `gh` next to the standing correction saying it fails.
- **Rule 5** — "rules 1–3" → "any rule in this file": kills the numeric
  back-reference rot and gives every rule a stated refusal behaviour.
- **Rule 12** — as rev 1: self-initiated output lands as new top-level DM
  messages; delivery-target changes verified by re-reading the job.
- **Rule 13 (new)** — the #1 mined directive (4 corrections, 3 days) promoted
  from a correction seed to a hard rule: own-skill/SOUL work never through a
  Claude Code or Orca session, Hermes subagents fine — and the mechanism is
  disambiguated from the machine: git/gh over ssh to cooper for rule 2's
  mirror is not such a session and is required. As a seed it collided with
  rule 2's own cooper recipe (refuse-the-mirror vs re-derive-Orca-is-fine,
  §2.14 and §2.1 — one failure on each side of the ambiguity).

## "## Standing corrections" — the durability mechanism (failure mode 2)

- **Trigger**: first occurrence of any correction about how Tars works that is
  not specific to the task at hand — the second-strike test and the
  remember/always/never keyword list are gone. Review replay showed the old
  trigger could never fire: detecting a *second* correction needs the first in
  context, which compaction has evicted (the 4×-repeated no-Orca directive
  contained none of the keywords and would never have landed). A redundant
  line costs nothing; a lost one costs a repeat.
- **Persistence**: rule 2's SOUL flow (above) — legal because rule 2 now grants
  it, executable because the recipe is literal.
- **Placement**: LAST section, making the live-file append append-only by
  construction — the structural gate reviewers found missing.
- **Current-thread mechanics**: the file reaches new sessions only (inventory
  §2: prompt persisted per-session, compaction reuses it), so the rule says
  what actually happens now: keep applying the correction from the transcript
  in the thread where it was given, tell Gaetan a fresh thread or session
  reset is what picks up the file. Which reset lever exists on v0.20.0 with
  `mode: none` must be confirmed at deploy (changemap rollout note).
- **Seeds**: two (meme transparent background; nextmobiles.com → 👥 Interne).
  The no-Orca seed graduated to rule 13; the gh-path seed graduated into the
  recipes themselves — a correction the executable text already obeys is dead
  weight and a contradiction risk.

## Phase 2, Tone, Language

Phase 2 cut to three lines (placeholder, pointers, "I ask instead of
assuming") — content intact, budget finding honoured. Tone deleted as a
near-verbatim duplicate of the Verdicts bullet (two places to drift between);
its one unique clause moved into Done-means-verified. Language unchanged,
moved above Standing corrections so the appendable section ends the file.

## What was deliberately NOT changed / not produced

Unchanged from rev 1: no config.yaml change (other agent's settled scope;
`tool_progress_conversations: {D0BBYNM01BL: "all"}` flag passed upstream —
**moot since 2026-08-13: the patch providing that key was dropped, the key no
longer exists, `tool_progress: "off"` governs alone**); no
skill-file edits; no model-swap recommendation in SOUL; rollout constraint
(new sessions only) rides in the changemap. One review suggestion not taken:
relocating the rule-2 recipe into a skill — reasons in changemap §Skipped.
