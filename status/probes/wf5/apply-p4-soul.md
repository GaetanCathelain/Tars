# WF5 — apply P4 (SOUL / guidelines redesign)

Peer session `apply-p4-soul`, worktree `/home/gaetan/orca/workspaces/Tars/apply-p4-soul`.
Mandate: `docs/plans/apply-P4-soul.md`. Proposal: `docs/proposals/P4-soul-guidelines-redesign.md`.
Source text: `artifacts/SOUL-proposed.md`. Change list: `artifacts/guidelines-changelist.md`.

All timestamps are the VM clock (**UTC**).

## Scope actually applied

| Item | Status |
|---|---|
| `~/.hermes/SOUL.md` — complete replacement (change-list §A, A1–A5) | see §Apply |
| Repo docs D1 (README), D2+D3 (PLAN.md), D4 (CLAUDE.md), D5 (`docs/specs/tars-profile.md`) | applied, repo-side only |
| **B1 `approvals:` config — PARKED, NOT APPLIED** | deliberately untouched; Gaetan has not ruled between `manual` / `smart` / `off` / route-into-Slack |
| `~/.hermes/skills/` (sibling peer owns `delegate-to-cooper/SKILL.md`) | untouched |
| `~/.hermes/config.yaml` | untouched |
| Gateway restart | **not performed — none is needed** (`SOUL.md` is read fresh at every system-prompt build) and it would have disrupted the sibling peer |
| `status/lane-a.md` | untouched (single writer: the hub) |

## Pre-apply state

Live file before the change:

```
-rw-rw-r-- 1 gaetan gaetan 1472 Aug  7 19:39 /home/gaetan/.hermes/SOUL.md
sha256 86a61caa16239e8d926960f29ba3bf6cb365dca5631e9cd54f1c968fd05e20c2
32 lines, 1472 bytes, mode 664
```

Backup (mandate step 2), taken with `cp -p` so mode and mtime survive:

```
~/.hermes/SOUL.md.bak-p4
-rw-rw-r-- 1 gaetan gaetan 1472 Aug  7 19:39
sha256 86a61caa16239e8d926960f29ba3bf6cb365dca5631e9cd54f1c968fd05e20c2   ← identical to live
```

Replacement staged on the VM ahead of the gate, deliberately **not** moved into place:

```
~/.hermes/.SOUL.md.p4new
-rw-rw-r-- 1 gaetan gaetan 3849 Aug  7 22:25   ← mode 664 via chmod --reference=SOUL.md
sha256 c0405ea6d73dc9950ac58c174fc891d6a0807ec4c0228a4c252b79db7606d482
       == sha256 of local artifacts/SOUL-proposed.md
```

Staging in the same directory as the target means the apply is a same-filesystem
`mv`, i.e. atomic, and the mode travels with the file. This is the explicit fix
for the earlier incident where an agent's `os.replace` silently dropped a file
from 600 to 664.

## Diff summary (live → proposed)

Verified with `diff -u` against a fresh `cat` of the live file, not by eye.

**Changed**

- Intro, +5 lines: the "secretary of his work" paragraph — holds the queue, schedules,
  writes briefs, prompts the agents, tracks, double-checks, reports; "a mirror of how he
  works himself … in his place, with his access". (A1)
- Rule 1: rescoped from *"I never write code"* (a file-type list) to *"I never produce the
  deliverable myself"* (role scope), explicitly refusing the "it's only markdown" loophole,
  and adding two carve-outs — writing the brief is not doing the work, and quoting an
  agent's output back to Gaetan is evidence owed, not a breach. (A2, widened)
- Rule 2: was *"I never open, review or merge a pull request, and I never push"* — which
  forbade the verification the redesign requires. Now *"I never merge, approve or push.
  Reading a pull request, a diff or a CI log is how I verify what I delegated."*
- Rule 3: "Work that needs **code written**" → "Work that needs **something built**".

**Added**

- Rule 6 — credentials: never read, print, echo or pass along a credential, token or key;
  name the secret instead. (A3 — promoted into SOUL so it does not die with the v1 skill.)
- Rule 7 — cooper is Tars's to act on with Gaetan's access, sudo included, and *"refusing to
  act is as much a failure as doing the implementation myself"*; p-Hermes (192.168.0.8)
  read-only, pve (192.168.0.3, a different machine) off-limits entirely. (A4)
- Rule 8 — announce before an irreversible action, as a statement, not a request.
- `## Phase 2 — Gaetan's knowledge and preferences`, a marked placeholder that invents no
  content. (A5)

**Unchanged** — `# Tars` / "I live on Slack" opening, rule 4, rule 5, `## Tone`, `## Language`.

### Rule 4 byte-identity (mandate step 1 — verified programmatically, not by eye)

Rule 4 is the identity frame plus the minimal reply `·`; it is the fix for the
empty-response incident, so it had to survive untouched. Extracted the rule-4 block from
both files and hashed each:

```
$ awk '/^4\. I answer Gaetan and no one else\./{p=1} p{print} p&&/no content, no reaction, no explanation\.$/{exit}' <file> | sha256sum
43af5d846e388dbdfa5529cd1d51be77c863a318f030bb38d72efa885c8f54e6   ← live ~/.hermes/SOUL.md
43af5d846e388dbdfa5529cd1d51be77c863a318f030bb38d72efa885c8f54e6   ← artifacts/SOUL-proposed.md
$ cmp <live block> <proposed block>   →  IDENTICAL (no output, exit 0)
```

And the minimal reply character itself, hexdumped from both blocks:

```
00000000  22 c2 b7 22 0a    |"..".|
```

`c2 b7` = U+00B7 MIDDLE DOT — correct in both, and specifically **not** an ASCII `.`
(`2e`). This is the point `guidelines-changelist.md` §D7 flags as owed to `docs/facts.md`'s
owner.

## Pre-apply behavioural baseline

The same question is asked before and after, so the post-apply answer can only come from
the new text. Asked on the VM via `~/.local/bin/hermes chat -Q -q`:

> Am I allowed to read a pull request, its diff and a CI log in order to verify work I
> delegated? Answer yes or no, in one sentence, and say which of your hard rules decides it.

**Before (session `20260807_222747_c52cc2`):**

```
No—hard rule 2 says I never open or review a pull request, which includes reading its diff
and CI log to verify it.
```

That is precisely the blocker P4 removes.

`errors.log` `Empty response` baseline before the change: **4 occurrences**, most recent
`2026-08-07 19:02:10` (`model=gpt-5.6-sol`), i.e. more than three hours before this apply
and unrelated to it. Any new occurrence after the apply would be attributable.

## Joint-apply gate

The live skill `~/.hermes/skills/delegate-to-cooper/SKILL.md:47` still reads *"Never `rm`,
`rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper"*, which contradicts new SOUL rule 7.
SOUL alone would put two opposing rules in one context window, so this apply was held at
the gate: everything above was prepared, then `P4 READY` was sent to the hub
(`orchestrator-16`) and the write waited for its `GO`, with the sibling peer applying P3
concurrently.

## Repo docs D1–D5 (applied — repo side, no VM impact)

Applied while held at the gate; the proposal marks these as landable "anytime".
D2/D3/D5 were delegated and then re-verified independently here. Detail:
`status/probes/wf5/apply-p4-repo-docs.md`.

| # | File | Change |
|---|---|---|
| D1 | `README.md` | separation-of-concerns reframe + PR reconciliation (see below) |
| D2 | `PLAN.md` §Amendments | **additive** — new "SOUL.md v2 (WF5 redesign, 2026-08-07)" bullet; the v1 bullet is left byte-intact so the record survives |
| D3 | `PLAN.md` §Guardrails | "enforced later in its profile/system prompt" → "a separation of concerns, not a security confinement; stated in…" |
| D4 | `CLAUDE.md` | same reframe as D1 + PR reconciliation |
| D5 | `docs/specs/tars-profile.md` | (a) superseded-by note on the "v1 is deliberately minimal" sentence; (b) the fenced SOUL v1 copy replaced with SOUL v2; (c) "Rules 1–3 are the PLAN.md guardrail" → "…the PLAN.md separation-of-concerns rule … job scope, not confinement" |

D5(b) verified independently, not taken on the subagent's word:

```
$ awk '/^```/{n++; next} n==1' docs/specs/tars-profile.md | diff - artifacts/SOUL-proposed.md
   → empty, exit 0: BYTE-IDENTICAL
$ grep -o '"·"' docs/specs/tars-profile.md | hexdump -C
   → 22 c2 b7 22 0a   (U+00B7 survived the splice)
```

### Two deviations from the authored change list, both deliberate

1. **The PR absolute (D1, and D4 by extension).** The mandate directs reconciling
   `README.md`'s stale absolute *"No PR is ever created by Tars"* to the settled reading:
   Tars does not do the work itself, but a PR resulting from a Tars-initiated Orca task is
   fine. D1's authored text kept that sentence verbatim, so it was rewritten; D4 left the
   identical claim standing in `CLAUDE.md` (*"and Tars never creates a PR"*), and reconciling
   one but not the other would have defeated the point, so the same fix was applied there and
   flagged to the hub in the `P4 READY` message.
   **Converged on the hub's wording.** The rebase before pushing hit a conflict on
   `CLAUDE.md`: the hub had landed its own reconciliation of the same sentence upstream, and a
   better one — it tracks new SOUL rule 2's actual verbs (*"never merges, approves or pushes.
   A PR authored by a session Tars delegated to is fine; Tars authoring one is not"*) instead
   of the vaguer "never authors a PR itself". The hub's text was taken wholesale for
   `CLAUDE.md`, and `README.md` was re-aligned to match it rather than left in a third
   phrasing. One settled reading now, in both files.
2. **D5(c) drops a trailing clause.** The authored replacement for `tars-profile.md:63-64` is a
   single sentence, so *"WF4 deliberately does **not** probe coding."* is not carried over. Kept
   as authored rather than re-adding it: the fact is unrelated to the reframe and still stated
   at `PLAN.md` §Guardrails (*"WF4 does not probe coding on purpose"*), so nothing is lost from
   the repo.

### How the gate actually resolved

The hub's `GO` never arrived. **Gaetan gave the order directly in the peer session**
("Go") after the `P4 READY` report — a human override of a peer-coordination gate, not a
gate bypass. The hub was told immediately, in the same message that reported the apply.

## Apply

```
$ flock ~/.hermes/.wf3.lock -c 'mv ~/.hermes/.SOUL.md.p4new ~/.hermes/SOUL.md'
mv rc=0
```

**`applied_at = 2026-08-07T22:32:45Z`** (VM clock, UTC).

Post-apply state, read back from the VM:

```
-rw-rw-r-- 1 gaetan gaetan 3849 Aug  7 22:25 /home/gaetan/.hermes/SOUL.md
sha256 c0405ea6d73dc9950ac58c174fc891d6a0807ec4c0228a4c252b79db7606d482
66 lines, 3849 bytes, mode 664
```

- sha256 equals `artifacts/SOUL-proposed.md` exactly → the installed file is the approved text.
- **Mode preserved at 664**, as the mandate requires. The staged file carried the mode
  (`chmod --reference`) and the apply was a same-filesystem `mv`, so there was no
  `os.replace`-style silent 664→600 (or 600→664) drift.
- Staging file `~/.hermes/.SOUL.md.p4new` consumed by the `mv` (no leftover).
- Backup re-verified *after* the apply: `~/.hermes/SOUL.md.bak-p4`, 1472 B,
  sha256 `86a61caa…e20c2` — intact, so rollback is one command.
- No gateway restart. None is needed (`SOUL.md` is read fresh per system-prompt build) and
  a restart would have disrupted the sibling peer.

### Open risk at apply time — the contradiction window

Checked in the same breath as the apply:

```
$ grep -n 'sudo' ~/.hermes/skills/delegate-to-cooper/SKILL.md
47:- Never `rm`, `rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper.
$ ls -l ~/.hermes/skills/delegate-to-cooper/SKILL.md
-rw-rw-r-- 1 gaetan gaetan 3304 Aug  7 20:21    ← mtime unchanged, sibling had not applied P3
```

So from 22:32:45Z the exact condition the joint-apply gate exists to prevent was live: SOUL
rule 7 grants sudo on cooper at prompt band 1 while the loaded skill forbids it at band 3.
Reported to the hub immediately with the rollback command. **This peer does not own that
file and did not touch it.** Whether the window has since closed is P3's record, not this
one's.

## Verification

### 1. The model actually sees the new text (mandate step 4)

Same question before and after, so the answer can only come from the changed file.
`~/.local/bin/hermes chat -Q -q`, on the VM:

| | Session | Answer |
|---|---|---|
| **Before** | `20260807_222747_c52cc2` | `No—hard rule 2 says I never open or review a pull request, which includes reading its diff and CI log to verify it.` |
| **After** | `20260807_223314_5ec652` | `Yes—Hard Rule 2 explicitly allows reading a pull request, diff, or CI log to verify delegated work.` |

**PASS.** The polarity flips on the exact clause P4 rewrote, and no restart was involved —
which also confirms in passing that `SOUL.md` is re-read per prompt build.

### 2. Identity frame survived — live Slack turn as Gaetan

Sent **natively** with the VM's Slack user credentials, because the claude.ai connector
cannot trigger Tars (`docs/facts.md` §"The claude.ai Slack connector CANNOT trigger Tars").
Tokens were read from `~/.hermes/.env` inside the remote shell and passed to `curl -K` on
stdin — never on argv, never echoed, never written to disk, and not reproduced here.

```
auth.test  →  ok=True  user_id=U08BDJAMSRZ  team=T7V1UGJ82      ← genuinely Gaetan
chat.postMessage → D0BBYNM01BL (home DM), ts=1786142043.792249
conversations.replies(ts) → REPLY from U0BBH85NAKH after ~12 s
```

Question asked: *"P4 check, one sentence please: which machine may you act on with sudo,
and which one is read-only to you?"* — Tars:

> I may act with sudo on **cooper**; **p-Hermes (192.168.0.8)** is read-only to me.

**PASS, twice over.** Tars still answers Gaetan, so the rule 4 identity frame survived the
replacement; and the content is new rule 7 verbatim, from the live Slack surface rather than
the CLI. Polled via `conversations.replies` on the sent ts, per `docs/facts.md` — Tars replies
in-thread and `conversations.history` is blind to it.

### 3. No new `Empty response` warnings

```
grep -c "Empty response" ~/.hermes/logs/errors.log  →  4      (baseline was 4)
occurrences timestamped at/after 22:32:45Z          →  none
```

**PASS.** The count did not move; the newest of the 4 is `19:02:10`, over three hours before
the apply. This is the failure mode a botched rule-4 edit produces, and it did not appear.

### 4. Allowlist early-reject — the control that must not lapse

With confinement dropped, `SLACK_ALLOWED_USERS` + the adapter early-reject is the **sole**
gate on transitive Orca access. State after the apply:

```
SLACK_ALLOWED_USERS present in ~/.hermes/.env, exactly 1 entry   (unchanged, not widened)
"[Slack] Early reject of unauthorized user …"  → 33 occurrences in ~/.hermes/logs/
code path intact: adapter.py:5546
```

**Not weakened — but honestly scoped.** The most recent firing is `20:46:05`, i.e. *before*
this apply, so **no post-apply firing of the control was observed**. That specific check is
**unverified**, and it cannot be closed from this session: producing one needs a genuine
non-Gaetan Slack sender, and the only other surface available (the claude.ai connector)
authenticates *as* Gaetan (`U08BDJAMSRZ`) and is in any case dropped as a bot sender before
the allowlist check is ever reached. What *is* established: the control is adapter-level,
pre-model and pre-dispatch, so a `SOUL.md` edit cannot affect it by construction; its config
and its code path are verified unchanged; and it has fired 33 times on this gateway. A fresh
negative test needs a second Slack identity and is owed to whoever can supply one.

## Verdict

| Check | Result |
|---|---|
| Rule 4 byte-identical (programmatic) | **PASS** — sha256 match + `c2 b7` |
| Installed file == approved text | **PASS** — sha256 `c0405ea6…d482` |
| Mode preserved (664) | **PASS** |
| Backup exists and is restorable | **PASS** — `SOUL.md.bak-p4`, sha verified post-apply |
| Model sees the new text | **PASS** — answer polarity flipped, no restart |
| Identity frame survived (live Slack) | **PASS** — replied in-thread in ~12 s |
| New rule 7 reaches the Slack surface | **PASS** |
| No new `Empty response` | **PASS** — still 4, none since apply |
| Allowlist config + code path unchanged | **PASS** |
| Allowlist fires post-apply on a real stranger | **UNVERIFIED** — needs a non-Gaetan sender this session cannot produce |
| B1 `approvals:` left parked | **PASS** — not applied |
| No gateway restart | **PASS** |
| Sibling P3 skill contradiction closed | **NOT THIS PEER'S** — still open at apply time, hub notified |

## Rollback

One command, no restart, ~5 s:

```
flock ~/.hermes/.wf3.lock -c 'cp -p ~/.hermes/SOUL.md.bak-p4 ~/.hermes/SOUL.md'
```

`cp -p` rather than `mv`, so the backup survives a second rollback attempt.
