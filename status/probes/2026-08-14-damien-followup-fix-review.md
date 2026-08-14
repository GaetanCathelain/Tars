# Synthesis judgement — conversation-state fix draft (2026-08-14)

Role: synthesis judge over four red-team lenses (binding, replay, regression, ops).
Target: `status/probes/2026-08-14-damien-followup-fix-draft.md` (H1 / H2 / H3 / RB).
**Nothing modified.** I wrote only this file. No `sops -d`, no VM write.

## Independent checks I ran (not taken from the lenses)

| Claim | Result |
|---|---|
| repo `SOUL.md` md5 / lines | `0f2d817607f813c363fe9338f17edd81`, 319 lines |
| VM `~/.hermes/SOUL.md` md5 | `0f2d817607f813c363fe9338f17edd81` — **matches** |
| VM `engagement-checker/SKILL.md` md5 | `4d73d985c52c5929d6831fe8e889d129` — **matches repo** (draft called this unverified; it is verified now) |
| `cbc504a` SOUL standing correction | 2026-08-14 **11:46:45 UTC** — real, and after verify.md's `c32c3a82…` baseline |
| SOUL append mechanism | `SOUL.md:185` = `printf … >> ~/.hermes/SOUL.md` — **no flock**, confirmed |
| `hermes chat --help` on the VM | `usage: hermes chat [-h] [-q QUERY] …` — **no positional prompt**. The runbook's probe is an invalid command. Confirmed. |
| `CONTEXT_FILE_MAX_CHARS` | `prompt_builder.py:1282` = `20_000`; head/tail ratios `0.7`/`0.2`; `_dynamic…` returns the flat 20K whenever `context_length` is unknown (`system_prompt.py:175-184`) |
| `wc -m ~/.hermes/SOUL.md` | **19,385** chars — 615 under the 20K floor today |
| `engagement-checker/SKILL.md:112` ban | exists verbatim, untouched by H3 |
| `daily-work-brief` board-first + byte-for-byte | confirmed (`:139`), 350-word cap excludes the board block (`:159`) |

So: both mirrors agree with the VM **right now**, and the ops lens's two most consequential
claims (invalid probe command, 20K truncation floor) are true.

---

## Adjudication table

Severity is mine, not the lens's. "×2/×3" = independently raised by that many lenses (extra weight).

### H1

| # | Finding | Lens | Verdict | Why |
|---|---|---|---|---|
| A1 | H1's anchor silently flips the job's **write** predicate: "close it unless something moves" anchored on Gaetan's last message fires the close branch whenever Gaetan spoke last — the normal end state of any exchange | replay | **VALID — blocker** | draft:61-62 anchors *the prompt*, not the narration. Ground truth in vm-run-trace:46 shows the incident job carried `save_issue(Done)` on that branch. A report-side fix that moves a write predicate is a strictly worse bug than the one it closes. |
| A2 | H1 (compile time) and H2 (run time) disagree on when "Gaetan's last message" is resolved; the literal H1 reading bakes a second constant ts | binding ×2 (replay) | **VALID — major** | Same defect class as the `Source:` ts it replaces. Job compiled Monday, Gaetan writes Tuesday, Wednesday's run reports the stale state. |
| A3 | H1 says "the DM"; H2 says "the DM **or channel**" — for a colleague Gaetan only meets in a channel, H1 anchors on a DM that does not exist | replay ×2 (regression) | **VALID — major** | Two hunks in one file must not name different targets for the same read. |
| A4 | H1 makes "the prompt names the DM" unconditional with no resolve step and no fallback; the incident needed two live human corrections to get the channel right | replay | **VALID — major** | vm-run-trace:33-36. H1 has no authoring-time equivalent of H2's say-so clause. |
| A5 | "H1 is decorative for enforcement; H2 is the whole fix" | binding | **PARTLY VALID** | The *conclusion* is wrong — A1 proves H1 has teeth (too many). What is valid inside it: the draft's stated rationale (draft:25-30, "rule 10 alone never fires on that path") is **stale**, because H2's trigger was deliberately widened to cron deliveries. Fix the rationale paragraph, not the hunk. |
| A6 | Legacy jobs: every already-scheduled follow-up keeps its wrong-anchored prompt; no `hermes cron list` audit anywhere | binding ×2 (regression) | **VALID — major** (RB item) | Against `SOUL.md:41-47` "a preference is not applied until every mechanism that enforces it has changed". Belongs in the runbook, not the hunk. |

### H2

| # | Finding | Lens | Verdict | Why |
|---|---|---|---|---|
| B1 | Trigger fires on every engagement-checker run (15/day, ≤6 person-keyed bullets) and every daily brief — one 30d DM read per named person — contradicting `engagement-checker:112` which H3 never amends | binding ×3 (replay, regression) | **VALID — blocker** | Verified: engagement-checker bullets are person-keyed (`• What — who`, `:704`); `:112` ban intact; cron runtime is short (`:106`). |
| B2 | Mandatory preamble has **no delivery scope**; rule 4 (`SOUL.md:216-228`) lets Tars post substantive content into group conversations that include Gaetan, so Gaetan's private 1:1 DM timing leaks to third parties | binding ×2 (regression) | **VALID — blocker** | The read is fine (Gaetan's own token); the *disclosure* is the regression. No carve-out to `D0BBYNM01BL`. |
| B3 | "widening until I have both" has no terminator and is unreachable past the measured `90d` free-tier cap; `limit` "must be empty when `cursor` is provided", so the paginated escape is forbidden by the same sentence | binding ×4 (all) | **VALID — blocker** | Live schema cache confirms. Scenario B — genuine silence, the case the fix exists for — dead-ends into the escape hatch instead of producing "Gaetan asked 5 days ago, no reply". The terminator is simply the wrong one: what the report needs is Gaetan's last message **and everything after it**, not both parties' last messages. |
| B4 | "whether it is still waiting on an answer" has no criterion separating a closing acknowledgment from an open ask — so on the incident's own ground truth the text licenses two opposite compliant reports ("he hasn't answered" vs "he answered, you closed it") | binding ×2 (replay) | **VALID — major (root cause of B4/B5)** | vm-run-trace:79. A rule that permits both readings of the incident it was written for has not fixed the incident. |
| B5 | Ordering ("before I state the ticket's state") collides with `daily-work-brief:139` byte-frozen board-first block; hard rule wins, so every morning Tars must displace the frozen board or break a hard rule | regression | **VALID — major** | Verified. Note the 350-word cap explicitly excludes the board block, so the *cap* half of that finding is weaker; the *ordering* half is real. |
| B6 | Escape hatch is self-declared with no obligation to name the call or the error — cheapest compliant path is to skip the read and print "unreachable" | binding | **VALID — major** | Rule 10's parent text names exact tools (`SOUL.md:258-261`); the new paragraph does not. One clause fixes it. |
| B7 | Silently rewrites engagement-checker's one-line delivery format (`SKILL.md:704`) | regression | **VALID — major, subsumed** | Evaporates once B1 narrows the trigger. |
| B8 | 30d read: truncation / ordering / tool-result budget unmeasured; `engagement-checker:617` documents the silent-drop failure for unprojected large reads | replay ×2 (regression) | **VALID as a gap — major** | Not a proven defect; it is exactly the "measure before believing recon claims" hard rule going unsatisfied for the one mechanism the fix rests on. Mitigate by allowing the message-count form of `limit`. |
| B9 | The patch takes SOUL from 19,385 → 21,099 chars, crossing the hard 20,000 `CONTEXT_FILE_MAX_CHARS` floor; under that floor the dropped middle (14,000–17,099) is **the new paragraph itself**, while H1 survives | ops | **VALID — major (latent)** | Verified constants. Does not fire today (cap 65,280 from a 272k model), but this repo swapped engines twice this week. My rewrite is not materially shorter — brevity cannot solve this. The answer is a size assertion + `prompt-size` check in the runbook, not a shorter rule. |
| B10 | Paragraph sits under rule 10's narrow lead ("I read the record before I answer **for the past**") and directly after the unscoped exemption "A ping that asks for something new needs none of this" | binding | **VALID — minor** | Two textual routes for a compressing reader to escape the duty. Fixed by an explicit opening clause in the rewrite. |
| B11 | Blast radius: it loosens, for every named person, the one control keeping colleague DM content out of Tars' context | replay | **NEEDS-GAETAN** | A deliberate trade, correctly identified. Not a defect — a call. |
| B12 | Contradicts `SOUL.md:77-79` "Verdicts, not logs. The outcome is my first line" | binding ×2 (replay) | **PARTLY VALID — minor** | Precedence (`SOUL.md:14-16`) resolves the collision cleanly; what survives is a style cost, and the "state it only when the loop is open" escape is worth having. Folded into the rewrite. |

### H3

| # | Finding | Lens | Verdict | Why |
|---|---|---|---|---|
| C1 | Carve-out is planted in §2 (per-candidate delta collection, up to 100 events/run, 15 runs/day) but worded for a §8 reporting concept ("an item being reported") — inert where written, and licenses candidate-scale reads where read | binding ×2 (regression) | **VALID — blocker** | Verified §2 retains up to 100 events (`SKILL.md:607`). The "~16×" multiplier is an upper bound, not an estimate — the placement defect stands on its own. |
| C2 | Leaves `SKILL.md:112` ("Do not rescan the whole day, inbox, Slack history") intact, so the contradiction survives one section higher | binding ×3 (replay, regression) | **VALID — major** | Verified verbatim. skill-audit.md:66 already named both caps. |
| C3 | H2 mandates "the DM **or channel**"; H3 carves out only "the DM" | replay | **VALID — major** | Channel read stays banned for anyone Gaetan does not DM. |
| C4 | The file's own precedent two lines below (`SKILL.md:131`) is "**one bounded** read … **only when** it cannot otherwise be resolved"; H3 copies none of it | regression | **VALID — major** | Best available fix shape: copy the neighbour. |
| C5 | Contradiction audit ran over a 5-of-55 mirror; sibling bans in 50 unmirrored VM skills are unaudited | replay | **VALID — minor** | Cannot claim "the contradictions are removed", only "the visible one is". State it as unverified rather than fixing it. |

### RB

| # | Finding | Lens | Verdict | Why |
|---|---|---|---|---|
| D1 | `test -s` cannot detect a truncated transfer; a dropped ssh stream installs a partial-but-non-empty constitution that Tars picks up on its next prompt build | ops | **VALID — blocker** | One-line fix, same round-trip. |
| D2 | Step-0 md5 gate is advisory (two outputs, no scripted compare, no non-zero exit) and separated from the write by an unbounded manual editing interval — a TOCTOU race against an agent that appends to this exact file without a lock, proved by `cbc504a` today | binding ×4 (all) | **VALID — blocker** | I verified `cbc504a` (11:46:45 UTC) and the lock-free `>>` at `SOUL.md:185`. A standing correction landing in the window is destroyed by the whole-file `mv`, and step 3's commit hides the loss. |
| D3 | No STOP defined for a **SOUL** mismatch (only for SKILL), and no tail-merge procedure | replay | **VALID — blocker** | draft:170-172 pre-announces SOUL as matching, on a stale measurement. |
| D4 | `flock ~/.hermes/.wf3.lock` serializes nothing: zero references to `.wf3.lock` in the installed Hermes source; `skill_manage` writes via `atomic_write_text` with no locking | ops | **VALID — major** | Both sides write atomically ⇒ no torn file, but last-writer-wins silently. Keep the flock (cheap, matches CLAUDE.md convention) — just stop treating it as the protection. |
| D5 | Verification probe `hermes chat "…"` is an invalid invocation | ops ×2 (binding, on different grounds) | **VALID — major, verified by me** | `hermes chat` takes `-q/--query` only. Binding's sub-claim that rule 4 would answer `·` on a CLI turn is **unproven speculation — INVALID as stated**; the finding survives entirely on the syntax. |
| D6 | Verification proves file presence, not behaviour: no probe compiles a fresh follow-up job, no scenario-B read | binding ×2 (replay) | **VALID — major** | Honest post-runbook status is "applied, unverified" unless one authoring dry-run is added. |
| D7 | Rollback uses non-atomic `cp -p` in place while the apply uses `mv` | ops | **VALID — major** | A prompt build racing the rollback reads a half-written constitution. |
| D8 | VM is written (step 2) before the repo is committed and pushed (step 3); a rebase conflict lands content on main that differs from what is running | ops | **VALID — major** | Free to fix: reorder. |
| D9 | No cron/scheduled-job re-audit | binding ×2 (regression) | **VALID — major** | See A6. |
| D10 | No size guard, no post-apply truncation check | ops | **VALID — minor→major given B9** | `wc -m` assertion + `prompt-size --json` + one log grep. |
| D11 | Runs from a Claude Code session the work rule 13 reserves to Tars, and pushes the repo half straight to main, bypassing rule 2's PR flow | regression ×2 (binding) | **NEEDS-GAETAN** | Rule 13 binds Tars, not Gaetan; rule 2 says every non-standing-correction line changes "only by Gaetan's reviewed amendment" — which is what this is. The *process* question (Tars applies it vs Claude Code; PR vs direct push for the mirror half) is his call. Draft flags it as risk 3 but the command sequence has no go-gate. |
| D12 | `chmod $SOUL_MODE` unguarded and after the `mv`; `$STAMP`/`$MODE` ephemeral so the rollback is unusable from a fresh shell; no `.bak` prune; orphan `.new` on local failure | ops ×2 (replay, regression) | **VALID — minor** | All one-liners. |
| D13 | No `hermes pause`; the exact skill being edited is loaded by a cron every 30 min, and the two transfers are separate ssh calls ⇒ one turn under new SOUL + old skill | ops | **VALID — minor** | Cheap and sanctioned lever. |
| D14 | Draft's provenance sentence ("`c32c3a82…` equal … no repo-side change since") is false | binding ×2 (replay) | **VALID — minor** | Verified. Ironically the *conclusion* is true today (both pairs match); the justification is not. |
| D15 | Only 5 of 55 live skill dirs are mirrored ⇒ the transfer could revert a `skill_manage` edit | regression | **VALID but defused today** | Ops lens checked: all six `engagement-checker` backups are ours (`-wf6-`, `-gcn30-32-`, `-20260813`); the only `-tars-selfedit` backup is on `delegate-to-cooper`. Live mechanism, not an observed event on this file — and both hashes match right now. |

---

## Verdicts

| Hunk | Verdict |
|---|---|
| **H1** | **FIX-WORDING** — the diagnosis is right and H1 has real teeth, but as drafted it moves a write predicate (A1) and bakes a stale anchor (A2). |
| **H2** | **FIX-WORDING** — correct contract, wrong bounds: trigger too wide (B1), no delivery scope (B2), unbounded window (B3), no ack-vs-ask criterion (B4). |
| **H3** | **FIX-WORDING** — right idea, wrong section, wrong width, and it leaves the stronger ban standing. |
| **RB** | **FIX-WORDING** — the file-handling *shape* is sound (`cp -p`, `umask 077`, mode capture/restore, same-device `mv`, byte-level `diff`, no `sops`) and "reload: none needed" is verified correct; the gates are not. Rewritten below. |

Nothing here argues for keeping the current behaviour. All four verdicts are "land it, worded differently".

---

## Rewrites — ready to paste

### H1 — replaces the appended sentences (draft:58-64)

> … if it recommends an action, I say so and wait. Because the job only ever
> runs what its prompt says, writing that prompt is where this bites: when I
> compile a follow-up, a reminder or a "close it unless something moves" check
> about a named person, the prompt names the conversation between Gaetan and
> that person — the DM, or the channel they actually use — and tells the run to
> report its state per rule 10, reading **at run time** who spoke last, never
> from a timestamp I bake in and never from a ticket's `Source:` ts. Any
> close-or-leave-it condition stays exactly as Gaetan worded it: what I read
> decides what I report, never what I write. If I cannot identify that
> conversation, I say so before scheduling the job rather than guess a channel.
> A boolean I invented is not a state I read.

Closes A1 (write predicate explicitly out of scope), A2 (run-time resolution stated),
A3 (matches H2's "DM or channel"), A4 (authoring-time say-so).

### H2 — replaces the drafted paragraph (draft:97-113), 4-space indent

>     **Conversation state, not just events.** This holds even when nothing was
>     asked of me. When what I am sending turns on whether a named person has
>     answered — a follow-up, a reminder, a "where does this stand", a cron or a
>     brief I deliver unasked — I do not answer it from a ticket. I read
>     Gaetan's conversation with that person (their DM, or the channel they
>     actually use) and state **who sent the last message, when, and whether
>     anything is still owed** before I draw any conclusion about the ticket; a
>     block a skill pastes byte-for-byte keeps its place, this goes in the prose
>     around it. Merely listing an item is not this: it fires when whether the
>     answer came is the point of what I am sending. An acknowledgment that
>     closes a loop is not an unanswered ask — "he answered and Gaetan closed it
>     himself" and "he has not answered" are different findings and I never turn
>     one into the other. I pass an explicit `limit`, `30d`, and once more at
>     `90d` — the cap — if Gaetan's own last message is not yet in view; the
>     default is a relative `1d` window, which returns nothing in exactly the
>     case worth reporting, a silence older than a day. I never anchor on a
>     ticket's `Source:` ts: that is the person's own opening message, so "did
>     they post after it" is trivially yes and proves nothing. This state I give
>     to Gaetan in our DM and nowhere else — never in a conversation that
>     includes anyone but him. If I cannot reach the conversation, or cannot
>     tell who spoke last, I say so and name the call I made and what it
>     returned.

Closes B1 (trigger keyed to the question, not to any named person; listing excluded),
B2 (DM-only delivery), B3 (`30d` → `90d` once → say so; anchor is Gaetan's last message
and everything after it, not both parties' last messages), B4 (ack ≠ open ask),
B5 (pasted-block exemption), B6 (name the call), B10 (opening override clause).
Residual, accepted: B8 stays a measurement gap — see must-fix 8 below; B9 is unchanged
in kind (this rewrite is ~the same length) and is handled in the runbook, not by brevity.

### H3 — two edits, not one

`SKILL.md:129`, replacing `Do not fetch unrelated channel history.`:

> Do not fetch unrelated channel history — with one exception: **one bounded**
> `conversations_history` read of Gaetan's conversation (DM or channel) with a
> person named in an item **this run is delivering**, and only when the delivery
> owes conversation state.

`SKILL.md:112`, replacing `Do not rescan the whole day, inbox, Slack history, or all Linear issues for context.`:

> Do not rescan the whole day, inbox, Slack history, or all Linear issues for
> context; the single exception is §2's bounded delivery-time read.

Closes C1 (keyed to *delivering*, not to candidates), C2 (:112 amended in the same edit),
C3 (DM or channel), C4 (copies `SKILL.md:131`'s "one bounded … only when" shape).
C5 stands: the 50 unmirrored skills remain unaudited — say so, do not claim otherwise.

### RB — corrected runbook

Order changed (repo first), gates made executable, probe replaced.

```bash
cd /home/gaetan/dev/orca-worktrees/Tars/improvements
set -eu
VM=gaetan@192.168.0.9
SOUL_R=~/.hermes/SOUL.md
SKILL_R=~/.hermes/skills/orchestration/engagement-checker/SKILL.md

# ── Step 0: pre-flight, scripted (no eyeballing) ────────────────────────────
git pull --rebase origin main
OLD_SOUL=$(md5sum < SOUL.md | cut -d' ' -f1)
OLD_SKILL=$(md5sum < skills/orchestration/engagement-checker/SKILL.md | cut -d' ' -f1)
VM_SOUL=$(ssh $VM "md5sum < $SOUL_R"  | cut -d' ' -f1)
VM_SKILL=$(ssh $VM "md5sum < $SKILL_R" | cut -d' ' -f1)
[ "$OLD_SOUL" = "$VM_SOUL" ]   || { echo "STOP: SOUL drift  repo=$OLD_SOUL vm=$VM_SOUL";   exit 1; }
[ "$OLD_SKILL" = "$VM_SKILL" ] || { echo "STOP: SKILL drift repo=$OLD_SKILL vm=$VM_SKILL"; exit 1; }
SOUL_MODE=$(ssh $VM "stat -c %a $SOUL_R");  [ -n "$SOUL_MODE" ]  || exit 1   # expect 664
SKILL_MODE=$(ssh $VM "stat -c %a $SKILL_R"); [ -n "$SKILL_MODE" ] || exit 1   # expect 600

# On drift: do NOT overwrite. Diff both sides, rebase the hunks onto the VM copy
# (SOUL drift is almost always a `## Standing corrections` tail line), re-run step 0.

# ── Step 0b: audit the jobs the fix is supposed to bind ─────────────────────
ssh $VM 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes cron list'
# Any existing follow-up/"close unless something moves" job anchored on a ticket
# `Source:` ts must be re-authored or deleted, or the fix is not applied
# (SOUL.md:41-47). Record the verdict per job in the status file.

# ── Step 1: edit repo copies, assert shape AND size ─────────────────────────
# (apply H1+H2 to SOUL.md, both H3 edits to the skill)
wc -l SOUL.md                                  # 319 + hunk lines
CH=$(wc -m < SOUL.md); echo "SOUL chars=$CH"
[ "$CH" -lt 20000 ] || echo "WARN: over the 20,000-char CONTEXT_FILE_MAX_CHARS floor \
(prompt_builder.py:1282) — safe only while the dynamic cap applies; step 4 must prove it"
grep -n 'Conversation state, not just events' SOUL.md
grep -c 'one bounded' skills/orchestration/engagement-checker/SKILL.md
NEW_SOUL=$(md5sum < SOUL.md | cut -d' ' -f1)
NEW_SKILL=$(md5sum < skills/orchestration/engagement-checker/SKILL.md | cut -d' ' -f1)

# ── Step 2: repo FIRST (main becomes the record of what gets installed) ─────
git add SOUL.md skills/orchestration/engagement-checker/SKILL.md status/probes/
git commit -m "SOUL rules 1+10: conversation state before ticket state; engagement-checker carve-out — GCN-49"
git pull --rebase origin main && git push origin HEAD:main
# if the rebase changed either file, recompute NEW_SOUL/NEW_SKILL before continuing

# ── Step 3: quiesce, back up, transfer with the gate INSIDE the write ───────
ssh $VM 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes pause --reason "SOUL/skill apply"'
STAMP=$(date -u +%Y%m%dT%H%M%SZ); echo "STAMP=$STAMP" | tee -a status/probes/2026-08-14-apply-log.txt
ssh $VM "cp -p $SOUL_R $SOUL_R.bak-convstate-$STAMP && cp -p $SKILL_R $SKILL_R.bak-convstate-$STAMP"

cat SOUL.md | ssh $VM "umask 077; cat > $SOUL_R.new && \
  [ \"\$(md5sum < $SOUL_R.new | cut -d' ' -f1)\" = '$NEW_SOUL' ] && \
  [ \"\$(md5sum < $SOUL_R     | cut -d' ' -f1)\" = '$OLD_SOUL' ] && \
  flock ~/.hermes/.wf3.lock -c 'mv $SOUL_R.new $SOUL_R' && \
  chmod $SOUL_MODE $SOUL_R && stat -c '%a %n' $SOUL_R || \
  { rm -f $SOUL_R.new; echo ABORT_SOUL; exit 1; }"
# same block for $SKILL_R with $NEW_SKILL / $OLD_SKILL / $SKILL_MODE

# ── Step 4: prove it landed, whole, and is being read ───────────────────────
ssh $VM "cat $SOUL_R"  | diff - SOUL.md && echo SOUL_MIRRORED
ssh $VM "cat $SKILL_R" | diff - skills/orchestration/engagement-checker/SKILL.md && echo SKILL_MIRRORED
ssh $VM "stat -c '%a %n' $SOUL_R $SKILL_R"          # expect 664 / 600
ssh $VM 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes prompt-size --json' | head -20
#   ^ offline, no API call. `stable` must grow by the patch size ⇒ SOUL loaded WHOLE.
ssh $VM 'export XDG_RUNTIME_DIR=/run/user/$(id -u); ~/.local/bin/hermes resume'
# after the first cron run:
ssh $VM "grep -rl 'Context file SOUL.md TRUNCATED' ~/.hermes/logs/ || echo NO_TRUNCATION"
```

Rollback (atomic, and the stamp is on disk from step 3):

```bash
ssh $VM "cp -p $SOUL_R.bak-convstate-$STAMP $SOUL_R.rb && \
  flock ~/.hermes/.wf3.lock -c 'mv $SOUL_R.rb $SOUL_R' && chmod $SOUL_MODE $SOUL_R"
# same for the skill; then git revert the repo commit.
```

Delete from the draft: the `hermes chat "…"` probe (invalid syntax, and it starts a
paid live agent session to prove a file read). Keep the flock — it is cheap and matches
the CLAUDE.md convention — but the md5 assertions inside the write are the actual gate.

Even with all of this, the honest status after the runbook is **"applied; behaviour
unverified"** until one follow-up job is compiled and its anchor read back (D6).

---

## NEEDS-GAETAN

1. **Who applies it, and how.** Rule 13 reserves SOUL/skill work to Tars; rule 2 makes
   every non-standing-correction line "Gaetan's reviewed amendment". Applying from a
   Claude Code session and pushing the mirror half straight to `main` is defensible as
   *his* amendment, but it is his call — and if he wants the mirror half to go by PR
   like Tars' own flow, step 2 changes. (D11)
2. **Which report he actually wants** in the incident case. My H2 rewrite deliberately
   produces "Damien answered « done » at 17:04, you closed it 15 s later — nothing is
   owed", **not** "Damien hasn't answered since your last message". His original
   complaint asked for the second sentence, and the ground truth says it is false
   (verify §Premise). Confirm the rewrite picks the right one. (B4 / replay §1.3)
3. **Conversation state in group posts.** My rewrite forbids it outright ("our DM and
   nowhere else"). If he wants Tars to be able to say "you last heard from X on the 3rd"
   in a group channel he is in, that clause must be relaxed. (B2)
4. **Widening the DM-read control.** H2 loosens, for named people, the one thing keeping
   colleague DM content out of Tars' context. Deliberate trade, worth an explicit yes. (B11)
5. **Existing scheduled jobs.** Step 0b will list them; re-authoring or deleting a live
   recurring job is a change to something he relies on. (A6/D9)

## Not defects — for the record

- "Reload: none needed" is **correct**, verified live against the installed source
  (`load_soul_md` re-reads uncached at every prompt build).
- Line arithmetic, indentation and all three BEFORE blocks match the current files.
- Backup/mode/atomicity shape of the runbook is right; no hard rule in `CLAUDE.md` is
  breached (no `sops -d`, no secret on argv, no `config.yaml`/`.env` touched).
