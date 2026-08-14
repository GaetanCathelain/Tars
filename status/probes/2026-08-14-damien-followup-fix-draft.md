# Fix draft — conversation-state reporting (Damien follow-up, 2026-08-14)

Ready-to-apply hunks + apply runbook. **Nothing applied. Nothing executed.**

Sources read in full: `SOUL.md` (repo), `skills/orchestration/engagement-checker/SKILL.md`,
`status/probes/2026-08-14-damien-followup-rootcause.md` §5,
`status/probes/2026-08-14-damien-followup-verify.md`.

Mirror/VM agreement at draft time: verify.md §B measured
`md5sum ~/.hermes/SOUL.md` on the VM (`c32c3a82…`) **equal** to the repo
`SOUL.md` (2026-08-14, read-only). This session re-hashed the repo file:
`0f2d817607f813c363fe9338f17edd81` — no repo-side change since, and runbook
step 0 re-checks both sides before anything is touched.
`engagement-checker/SKILL.md` VM-vs-mirror equality is **unverified** — step 0
settles it and gates the apply.

Line numbers are repo `SOUL.md` (319 lines) and
`skills/orchestration/engagement-checker/SKILL.md` (739 lines) as of this commit.

Design: **one contract, two references.** Rule 10 states the conversation-state
contract (H2). Rule 1's cron clause only points at it for the job-compilation
step (H1). The skill line only carves itself out of the way (H3). No paragraph
is duplicated.

Why both H1 and H2 are needed (verify §"Better-fitting explanation"): rule 10
triggers on an inbound message about the past, and the incident ran as a cron
with `history=0` — rule 10 alone never fires on that path. Rule 1's cron clause
("authorises exactly the job it names") is what narrowed the run to its
prompt's boolean, and the prompt was compiled at 17:11–17:13, hours before the
run. So the compile step has to be bound where that clause lives.

---

## H1 — `SOUL.md`, "Every turn" bullet 1, cron clause

Append two sentences to the end of the bullet. Continuation lines are indented
2 spaces, matching the bullet.

### BEFORE (lines 28–33, verbatim)

```
28	  Three things are not "old work": a cron, schedule or reminder of mine fires
29	  as its own instruction — it authorises exactly the job it names, nothing
30	  found in the thread around it; checking on a delegation I already dispatched
31	  is tracking it, not re-running it; and when a delegation I started earlier
32	  returns, its output is something I report — it never authorises a write on
33	  its own; if it recommends an action, I say so and wait.
```

### AFTER (replaces lines 28–33)

```
  Three things are not "old work": a cron, schedule or reminder of mine fires
  as its own instruction — it authorises exactly the job it names, nothing
  found in the thread around it; checking on a delegation I already dispatched
  is tracking it, not re-running it; and when a delegation I started earlier
  returns, its output is something I report — it never authorises a write on
  its own; if it recommends an action, I say so and wait. Because the job only
  ever runs what its prompt says, writing that prompt is where this bites:
  when I compile a follow-up, a reminder or a "close it unless something
  moves" check about a named person, the prompt names the DM between Gaetan
  and that person and anchors on Gaetan's last message in it — never a
  ticket's `Source:` ts — and it requires the delivery to state conversation
  state as rule 10 does. A boolean I invented is not a state I read.
```

Effect: the defect's actual locus — job-authoring time — is now constrained, on
the one path (cron/schedule/reminder) that rule 10's trigger does not reach.

---

## H2 — `SOUL.md`, hard rule 10, appended paragraph

Insert one blank line after line 265 (end of rule 10) and then the paragraph,
indented 4 spaces to match rule 10's continuation lines. Line 266 is currently
blank and line 267 starts rule 11 — insert **between 265 and 266**.

### BEFORE (lines 262–267, verbatim)

```
262	    tool call that ran, not my memory of what I meant to do. A ping that asks
263	    for something new needs none of this; an account of something past always
264	    does. I never narrate a cause I have not read: either I checked, or I say
265	    I have not checked.
266	
267	11. A rejected or failed delivery is reported as failed, error included —
```

### AFTER (replaces lines 262–267)

```
    tool call that ran, not my memory of what I meant to do. A ping that asks
    for something new needs none of this; an account of something past always
    does. I never narrate a cause I have not read: either I checked, or I say
    I have not checked.

    **Conversation state, not just events.** Whenever I report on a ticket, a
    task or a status that involves a named person other than Gaetan — whether
    I am answering him or delivering a cron, a schedule or a brief nobody
    asked for — I read the DM or channel between Gaetan and that person
    first, and I state **who sent the last message, when, and whether it is
    still waiting on an answer** before I state the ticket's state. That read
    is anchored on Gaetan's last message there, never on a ticket's `Source:`
    ts — that ts is the person's own opening message, so "did they post after
    it" is trivially yes and proves nothing. I pass an explicit `limit` wide
    enough to reach the last message of each side, starting at `30d` and
    widening until I have both: the tool's default is a relative `1d` window,
    which returns nothing in exactly the case worth reporting, a silence
    older than a day. I report the state I read and nothing past it — "he
    answered and Gaetan closed it himself" and "he has not answered" are
    different findings and I never turn one into the other. If the
    conversation is unreachable, or I cannot establish who spoke last, I say
    that in the report instead of dropping it.

11. A rejected or failed delivery is reported as failed, error included —
```

Deltas vs the rootcause §5 draft, each closing a verify-file flaw:

- Trigger widened past "answering a question about the past" to cover cron /
  schedule / brief deliveries — rule 10's own trigger otherwise excludes the
  path the incident took (verify §"Better-fitting explanation", bullet 2).
- "still unanswered" → "still waiting on an answer", plus an explicit ban on
  upgrading "answered, Gaetan closed it" into "he went silent": the premise of
  the complaint was false for this exchange (verify §Premise) and the rule must
  not manufacture the silence it looks for.
- Dropped the draft's "the ~20-message cap does not apply to this read" —
  rule 10's cap is about the channel Tars was pinged on, a different read; the
  explicit `limit` sentence carries the whole point without amending a cap.
- Kept, unchanged in substance: anchor on Gaetan's last message, `30d` start
  widening until both sides are found, say-so-when-unreachable.

---

## H3 — `skills/orchestration/engagement-checker/SKILL.md`, line 129

One-line carve-out so §2 stops contradicting the new SOUL paragraph.

### BEFORE (line 129, verbatim — final sentence-pair of the paragraph)

```
129	Filter every result by exact Slack `ts` against the cursor before classifying it. Deduplicate the two views by channel and timestamp. For a new candidate only, use `mcp__slack__conversations_replies(channel_id, thread_ts)` to recover enough thread context to decide whether there is a commitment, unanswered ask, resolution, or user instruction. Do not fetch unrelated channel history. Resolve names from source data rather than guessing.
```

### AFTER (replaces line 129)

```
Filter every result by exact Slack `ts` against the cursor before classifying it. Deduplicate the two views by channel and timestamp. For a new candidate only, use `mcp__slack__conversations_replies(channel_id, thread_ts)` to recover enough thread context to decide whether there is a commitment, unanswered ask, resolution, or user instruction. Do not fetch unrelated channel history, except the DM between Gaetan and a person named in an item being reported. Resolve names from source data rather than guessing.
```

Diff is exactly `history.` → `history, except the DM between Gaetan and a person named in an item being reported.` — the file's own register (imperative, Gaetan named, one long line per paragraph) is preserved.

---

## Apply runbook (commands only — NOT executed)

No `sops -d` anywhere; no secret is read, printed or transferred. The repo copy
is the source of truth for the transfer (byte-identical by construction; no
`sed` on the VM).

### Step 0 — pre-flight: repo and VM must already agree

```bash
cd /home/gaetan/dev/orca-worktrees/Tars/improvements
git pull --rebase origin main
md5sum SOUL.md skills/orchestration/engagement-checker/SKILL.md
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/SOUL.md ~/.hermes/skills/orchestration/engagement-checker/SKILL.md'
```

Both pairs must match. `SOUL.md` matched on 2026-08-14 (verify §B). **If the
`SKILL.md` pair differs, STOP** — Tars edits its own skills with `skill_manage`,
so the VM may hold the newer copy; reconcile before overwriting either side.

### Step 1 — capture modes, then edit the repo copies

```bash
SOUL_MODE=$(ssh gaetan@192.168.0.9 'stat -c %a ~/.hermes/SOUL.md')
SKILL_MODE=$(ssh gaetan@192.168.0.9 'stat -c %a ~/.hermes/skills/orchestration/engagement-checker/SKILL.md')
echo "SOUL_MODE=$SOUL_MODE SKILL_MODE=$SKILL_MODE"   # must be non-empty before continuing
```

Apply H1 + H2 to `SOUL.md` and H3 to
`skills/orchestration/engagement-checker/SKILL.md` in the repo (editor or
`Edit` tool — hunks above are exact). Then sanity-check the shape:

```bash
wc -l SOUL.md                                    # expect 319 + 24 = 343  (H1 +6, H2 +18)
grep -n 'Conversation state, not just events' SOUL.md
grep -n 'A boolean I invented is not a state I read' SOUL.md
grep -c 'except the DM between Gaetan' skills/orchestration/engagement-checker/SKILL.md   # expect 1
git diff --stat
```

### Step 2 — VM side: backup, transfer under the lock, restore mode

```bash
STAMP=$(date -u +%Y%m%dT%H%M%SZ)

# backups (mode-preserving)
ssh gaetan@192.168.0.9 "cp -p ~/.hermes/SOUL.md ~/.hermes/SOUL.md.bak-convstate-$STAMP && \
  cp -p ~/.hermes/skills/orchestration/engagement-checker/SKILL.md \
        ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-convstate-$STAMP && \
  ls -l ~/.hermes/SOUL.md.bak-convstate-$STAMP \
        ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.bak-convstate-$STAMP"

# SOUL.md — tmp+mv under flock, then restore the captured mode (tmp+mv drops it)
cat SOUL.md | ssh gaetan@192.168.0.9 "umask 077; cat > ~/.hermes/SOUL.md.new && \
  test -s ~/.hermes/SOUL.md.new && \
  flock ~/.hermes/.wf3.lock -c 'mv ~/.hermes/SOUL.md.new ~/.hermes/SOUL.md' && \
  chmod $SOUL_MODE ~/.hermes/SOUL.md && stat -c '%a %n' ~/.hermes/SOUL.md"

# engagement-checker SKILL.md — same shape
cat skills/orchestration/engagement-checker/SKILL.md | ssh gaetan@192.168.0.9 \
  "umask 077; cat > ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new && \
  test -s ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new && \
  flock ~/.hermes/.wf3.lock -c 'mv ~/.hermes/skills/orchestration/engagement-checker/SKILL.md.new \
      ~/.hermes/skills/orchestration/engagement-checker/SKILL.md' && \
  chmod $SKILL_MODE ~/.hermes/skills/orchestration/engagement-checker/SKILL.md && \
  stat -c '%a %n' ~/.hermes/skills/orchestration/engagement-checker/SKILL.md"

ssh gaetan@192.168.0.9 'md5sum ~/.hermes/SOUL.md ~/.hermes/skills/orchestration/engagement-checker/SKILL.md'
```

**Reload: none needed.** `SOUL.md` is read fresh at every system-prompt build —
`_r.load_soul_md()` in `agent/system_prompt.py:41-46`
(`status/probes/wf5/guidelines-runtime.md:61`; applied without restart and
recorded as such in `status/probes/wf5/apply-p4-soul.md:18,212` and
`docs/plans/apply-P4-soul.md:22`). Skill files are likewise re-read from disk
per run (`docs/proposals/P6-knowledge-bases.md:300`,
`P7-direct-analysis-exception.md:255`). Caveat from `SOUL.md:309-313`: a
**running** Slack thread keeps the prompt it started with — the change lands in
a fresh thread or after a session reset, not in a live conversation. Check it
that way:

```bash
# new thread, not an existing one; expect it to quote the new clause
ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); \
  ~/.local/bin/hermes chat "Quote rule 10s conversation-state paragraph verbatim."'
```

(If a restart is ever wanted anyway: `systemctl --user restart <hermes unit>` —
not required here, and it would disturb any live session.)

### Step 3 — repo side: commit and push

```bash
cd /home/gaetan/dev/orca-worktrees/Tars/improvements
git add SOUL.md skills/orchestration/engagement-checker/SKILL.md \
        status/probes/2026-08-14-damien-followup-fix-draft.md
git commit -m "SOUL rules 1+10: report conversation state (who spoke last, unanswered) before ticket state; engagement-checker carve-out — GCN-49 Damien follow-up fix"
git pull --rebase origin main && git push origin HEAD:main
```

### Step 4 — verification: VM vs mirror, both files

```bash
cd /home/gaetan/dev/orca-worktrees/Tars/improvements
md5sum SOUL.md skills/orchestration/engagement-checker/SKILL.md
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/SOUL.md ~/.hermes/skills/orchestration/engagement-checker/SKILL.md'
# byte-level proof, not just hashes:
ssh gaetan@192.168.0.9 'cat ~/.hermes/SOUL.md' | diff - SOUL.md && echo SOUL_MIRRORED
ssh gaetan@192.168.0.9 'cat ~/.hermes/skills/orchestration/engagement-checker/SKILL.md' \
  | diff - skills/orchestration/engagement-checker/SKILL.md && echo SKILL_MIRRORED
ssh gaetan@192.168.0.9 'stat -c "%a %n" ~/.hermes/SOUL.md ~/.hermes/skills/orchestration/engagement-checker/SKILL.md'
```

Rollback: `ssh gaetan@192.168.0.9 "flock ~/.hermes/.wf3.lock -c 'cp -p ~/.hermes/SOUL.md.bak-convstate-$STAMP ~/.hermes/SOUL.md'"` (same for the skill), then `git revert` the repo commit.

---

## Risk note

1. The `engagement-checker` mirror may be stale (5 of 55 live skill dirs are
   mirrored, verify §B) — step 0's md5 gate is the only thing preventing this
   apply from silently reverting a `skill_manage` edit Tars made on the VM.
2. H2 costs one extra Slack `conversations_history` call per person-scoped
   report with a `30d` window, and widens what Tars pulls into context; the
   H3 carve-out is deliberately narrow (only a person named in the item) so
   engagement-checker's "don't rescan the day" discipline still holds.
3. SOUL rule 13 makes Tars' own operating record Tars' to edit, never a Claude
   Code session's — applying this from cooper is Gaetan's own reviewed
   amendment (rule 2's "every other line changes only by Gaetan's reviewed
   amendment"), so it needs his go, and the live thread where the correction
   was discussed keeps the old prompt until it is reset.
