# Deep read: engagement-checker, daily-work-brief, SOUL.md, cooper reachability, other Linear-touching skills

Repo checkout: `/home/gaetan/dev/Tars/.claude/worktrees/kanban-research`
VM: `ssh gaetan@192.168.0.9` (read-only, no secrets printed)

## 0. Repo vs VM diff verdict

```
diff repo/skills/engagement-checker/SKILL.md  vm:~/.hermes/skills/orchestration/engagement-checker/SKILL.md  -> 0 lines
diff repo/skills/daily-work-brief/SKILL.md    vm:~/.hermes/skills/orchestration/daily-work-brief/SKILL.md    -> 0 lines
diff repo/SOUL.md                              vm:~/.hermes/SOUL.md                                          -> 0 lines
```
**Both mirrors are byte-identical to the VM right now.** Neither side has moved
ahead of the other — the repo commit history (`#33`–`#36`, latest `b12d802`
"engagement-checker skill: read decisions from reporting conversation (#36)")
already matches what's live. This confirms the documented flow in
`Tars/CLAUDE.md`: Tars edits `~/.hermes/skills/**/SKILL.md` live via
`skill_manage`, then self-PRs the same content into this repo in the same
turn (SOUL rule 2) — so a deep-read right now finds no reconciliation work,
just a clean baseline to build the new-team change on top of.

Sources read:
- `/home/gaetan/dev/Tars/.claude/worktrees/kanban-research/skills/engagement-checker/SKILL.md`
- `/home/gaetan/dev/Tars/.claude/worktrees/kanban-research/skills/daily-work-brief/SKILL.md`
- `/home/gaetan/dev/Tars/.claude/worktrees/kanban-research/SOUL.md`
- VM: `~/.hermes/skills/orchestration/engagement-checker/SKILL.md`
- VM: `~/.hermes/skills/orchestration/daily-work-brief/SKILL.md`
- VM: `~/.hermes/SOUL.md`

## 1. engagement-checker — pipeline summary

`~/.hermes/skills/orchestration/engagement-checker/SKILL.md`, v1.2.0. Half-hourly
weekday monitor (10:00–17:00 Europe/Paris) for Gaetan's open loops. Tars runs it
directly — never delegates, never spawns another agent for it.

**Sources scanned** (§2–4): Slack (`mcp__slack__conversations_search_messages`,
two views: authored-by-Gaetan and involving-Gaetan, plus thread replies for new
candidates), Gmail (via `google-workspace`'s `google_api.py gmail search`,
read-only), Linear (local GraphQL, see below). Workday gate first checks
weekday + `calendrier.api.gouv.fr` holidays + PayFit leave email, same logic as
`daily-work-brief`.

**How Linear is queried** (§4): NOT through an MCP tool and NOT through cooper.
A hardcoded Python collector script (embedded verbatim in the SKILL.md) is run
locally on the VM with `python3`, talking directly to
`https://api.linear.app/graphql` via stdlib `urllib`, authenticated with the
inherited `LINEAR_API_KEY` env var. It is allowlisted to exactly three GraphQL
query operations — `AssignedIssues` (filter: `assignee: {isMe: {eq: true}}`,
`updatedAt: {gte: $since}`), `IssueComments`, `IssueHistory` — rejects anything
containing `mutation`, rejects redirects, pins the final response URL, caps
response bytes, and redacts credential-shaped strings before any output. It
only ever asks for issues assigned to the authenticated viewer ("isMe") — it
does **not** currently scope by team, so a personal team + "read every ticket
across every team" default will need the query surface widened (today it's
narrower: viewer-assigned only, not "every ticket").

**How items enter/leave the JSON state** (durable state at
`~/.hermes/state/engagement-checker.json`, §"Durable state"): each source
(slack/email/linear) has its own `cursor`/`last_success`/`failures`/`seen[]`.
Linear items get stable id `linear:<issue-key>`. New/updated issues from the
collector become `items{}` entries with `status` open/snoozed/waiting/done/
dismissed, classified in §5 by kind (explicit promise, direct unanswered ask,
review/decision, blocker/security/customer/incident, FYI/noise=discarded), then
scored in §6 (kind weight + overdue/stakeholder/waiting/time-of-day bonuses,
threshold 60, source-specific cooldowns). Items leave the state 14 days after
done/dismissed, or auto-dismiss after 30 days untouched-open. Never persists
raw descriptions/comments/credentials — bounded 240-char snippets only.

**How it reports** (§7): delivery channel is whatever conversation the cron job
is scheduled to deliver to (VM's live cron: `slack:C0BP2GZUFSR:...`, i.e. the
Slack "reporting conversation" — same channel used for daily-work-brief). At
most six items, one line each: `• What — who; why now. Next: action. [EC-XXXX]`.
Most runs return exactly `[SILENT]`.

**Where "decisions from reporting conversation" (commit #36) come in**: §2,
last paragraph: "Also inspect new messages in the configured Slack reporting
conversation for decisions about existing engagement items. A reply in a
reminder thread is authoritative when the parent reminder names one `short_id`;
when it contains several items, require the reply to name a short ID, person,
or unique topic." This feeds §5's reconciliation step (done/dismissed/waiting/
snoozed transitions applied only when the reply resolves unambiguously to one
item; otherwise stored as a bounded `ambiguous_instructions` record, never
mutated silently).

## 2. daily-work-brief — pipeline summary

`~/.hermes/skills/orchestration/daily-work-brief/SKILL.md`, v1.0.0. One
weekday morning brief; cron `30 8 * * 1-5` (VM `hermes cron list`, job
`e231e5faf180`, next run 2026-08-11T08:30:00+02:00, delivers to
`slack:C0BP2GZUFSR:1786359613.979759` — same channel/thread anchor as the
engagement-checker jobs).

**Exactly which sources it pulls today** (§3):
- **Slack** — exhaustive within the window via `mcp__slack__conversations_search_messages` (`filter_users_from=U08BDJAMSRZ` + messages addressed to/about him), with thread-context recovery.
- **Cooper — Claude, shell, repos** — YES, it already reads cooper Orca/Claude session history, **directly over `ssh cooper`, read-only**, no delegation:
  > "Cooper inspection is read-only analysis. Do it directly over `ssh cooper`; do not spawn a coding agent merely to read activity." (§"Load only the branches used")
  Concretely (§3 "Cooper — Claude, shell, repos"):
  - inspect Orca's registered repos/worktrees and current agent/session state (implies `orca repo list --json` / `orca worktree list --json` style calls, per the `delegate-to-cooper` skill's documented `orca` CLI at `/usr/local/bin/orca` over non-interactive ssh);
  - locate Claude Code transcript files modified in the window under the user's Claude data directories (i.e. `~/.claude/projects/**` style transcript files), reading only bounded relevant turns;
  - inspect timestamped shell history (untimestamped `.bash_history` is supporting-only, never proof of window membership);
  - `git log` across Orca-registered repos for Gaetan-authored commits/merges in the window.
  SOUL/skill explicitly forbids spawning/prompting/resuming/modifying an Orca session for this — evidence-gathering only.
- **GitHub** — `gh` (already authenticated) for PRs merged by Gaetan in the window.
- **Linear** — "any already-configured Linear integration, CLI, API, email, or Slack links" to find tickets completed/merged in the window, counting only explicit terminal states; if unreachable, it must state "Linear coverage unavailable" rather than guess from branch names. Unlike engagement-checker, daily-work-brief does **not** embed a hardcoded GraphQL collector — it's looser/optional ("any already-configured... integration"), so this is the weaker of the two Linear touchpoints and likely needs the most rework for the new personal-team SSoT.
- **Email/calendar** — via `google-workspace` (fallback `himalaya`); PayFit-leave check is mandatory, broader mail is supporting evidence only.

**Output format** (§5): fixed compact shape — `**Since the last daily**` outcome
bullets, `**Stats**` (PRs merged N + ids, Linear tickets completed N + ids,
≤2 other counts), `**Today**` (1. highest-leverage action, 2. follow-ups w/
owner, 3. work in motion, 4. `Oli:` report/no-report line), optional trailing
`Coverage:` line. Cap ~350 words. Delivered to the Slack reporting conversation
(cron `deliver: slack:...`).

**Cron time**: `30 8 * * 1-5` Europe/Paris (08:30 workdays), confirmed live via
`hermes cron list` on the VM (job name "Gaetan daily work brief",
`e231e5faf180`, last run 2026-08-10T08:35:24+02:00 ok).

## 3. How the VM reaches cooper for Orca/Claude history

**`~/.ssh/config` on the VM** (hosts/aliases only):
```
Host cooper
    HostName 192.168.0.4
    User gaetan
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

Host pve
    HostName 192.168.0.3
    User root
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

Host mac
    HostName macbook-pro-de-gaetan.tail6e788b.ts.net
    User gcath
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

Host phermes
    HostName 192.168.0.8
    User hermes
    IdentityFile ~/.ssh/id_ed25519
    BatchMode yes
    ConnectTimeout 5
```
So `ssh cooper` from the VM is a direct, key-based, non-interactive hop to
`192.168.0.4` as `gaetan` — no jump host, no bastion.

**grep for ssh/cooper/orca across `~/.hermes/skills/`** (excluding `.bak`
files, schema/reference noise, and the skills index cache):
- `skills/delegate-to-cooper/SKILL.md` — the general-purpose "drive Orca on
  cooper" skill (see below).
- `skills/orchestration/daily-work-brief/SKILL.md` — the only other skill that
  references cooper, and it does so for **read-only evidence gathering**, not
  delegation (see §2 above).
- `skills/orchestration/engagement-checker/SKILL.md` — mentions `cooper`/`orca`
  only indirectly via SOUL's rule-2 skill-mirror sequence text quoted inside
  its own metadata description; engagement-checker itself never touches
  cooper — it is explicitly Tars-local (Slack/email/Linear only, "Do not
  spawn, prompt, resume, modify, or delegate through Claude, a worktree/
  session, or another non-Tars agent system").

**Orca/Claude history surfaces on cooper, per the skills' own text**:
1. **Orca CLI** (`/usr/local/bin/orca`, confirmed on non-interactive-ssh PATH,
   `delegate-to-cooper/SKILL.md`, "[help-verified]" 2026-08-07): `repo list
   --json`, `worktree list --repo <selector> --json`, `terminal list
   --worktree <selector> --json`, `orchestration run-create` / `task-create` /
   `check --peek` / `worker-start` / `worker-show`. This is the live
   session/dispatch state surface — repos, worktrees, terminals, durable
   run/task rows.
2. **Claude Code transcript files** — `daily-work-brief` says "locate Claude
   Code transcript files modified in the window under the user's Claude data
   directories" (i.e. `~/.claude/projects/**/*.jsonl`-style transcripts on
   cooper, matching this very session's own `~/.claude/projects/...` layout),
   read for bounded relevant turns only.
3. **Shell history** — timestamped shell history when available; raw
   `.bash_history` is explicitly downgraded to "supporting context only, never
   proof of window membership."
4. **git log** — across every Orca-registered repo, for Gaetan-authored
   commits/merges in the window.

No dedicated "Orca session list for Linear ticket creation" surface exists yet
in either skill — `delegate-to-cooper` is a *write* path (spawns/tracks Orca
sessions to do coding work), while `daily-work-brief`'s cooper inspection is a
*read* path (evidence for the brief). The default-Linear-team behaviour the
task describes ("same default for Claude/Orca sessions on cooper") is not
currently encoded anywhere in these three skills — it would need either a new
rule in `delegate-to-cooper` (brief template default) or a cooper-side default
(e.g. this repo's own `CLAUDE.md`/a cooper skill), not something
engagement-checker or daily-work-brief currently control.

## 4. SOUL.md — diff + where a "default Linear team" rule fits

Diff (repo vs VM): **0 lines — identical.**

**Where it would naturally live**: not in SOUL.md. SOUL.md's Hard Rules are
about *authority/delegation boundaries* (who may implement, merge, answer,
touch which machine) — rules 1–9 — plus a placeholder "Phase 2 — Gaetan's
knowledge and preferences" section (currently empty, reserved for
`~/.hermes/memories/USER.md`-sourced content). A "default team for new Linear
tickets" is an *operational default for a specific tool call*, not an
authority boundary — it belongs in the **skill(s) that create Linear issues**
(most directly `to-tickets/SKILL.md`, and potentially a new/extended
`engagement-checker`/`daily-work-brief` Linear-write path if one gets added),
or in a small shared reference doc those skills point to. SOUL is also
explicitly "not yet written" for Phase 2 personal-preference content — adding
a Linear-team default there would be premature scaffolding against a section
marked "do not invent content here."

Relevant existing SOUL rules to slot the new rule beside without conflict:
- **Rule 1** (implementation deliverables never Tars's to produce; analysis is
  Tars's to do directly) — creating a Linear ticket is a **write**, so it is
  either (a) something Tars does itself if ticket creation is treated as
  Tars's own operational bookkeeping (like `skill_manage` in rule 2), or (b)
  something a delegated agent does as part of its own deliverable. The
  plan needs to state explicitly which of these ticket-creation is, because
  rule 1 currently only carves out *analysis* and *the skill-mirror PR* (rule
  2) as things Tars may write.
- **Rule 2** (Tars never merges/approves/pushes, one exception: its own skill
  mirror via PR+squash-merge) — a new Linear team is not a git artifact, so
  this rule doesn't directly gate it, but the *pattern* (a narrow, explicit,
  named exception to a broad "don't write" rule) is the template to follow if
  Tars is granted a scoped Linear-write exception.
- **Rule 6** (never read/print/echo/pass a credential) — governs
  `LINEAR_API_KEY` handling; the existing engagement-checker collector's
  redaction/no-argv discipline (§4 script) is the pattern any new Linear write
  path must match.
- **Rule 9** (knowledge base at `~/dev/mc-metarepo`, searched before
  delegating) — analogous pointer pattern for where a "personal team ID /
  workload labels / priority" reference could live if it needs to be shared
  across skills rather than duplicated in each SKILL.md.

## 5. Other skills under `~/.hermes/skills/` mentioning "linear"

`grep -rli linear ~/.hermes/skills/`, filtered to skills plausibly relevant
(dropping unrelated hits — ML/creative-skill references to the geometric/
mathematical sense of "linear", template examples that happen to reference
linear.app as a design inspiration, PDF/W&B/eval docs, and non-SKILL reference
files):

- `skills/orchestration/engagement-checker/SKILL.md` — covered above.
- `skills/orchestration/daily-work-brief/SKILL.md` — covered above.
- `skills/to-tickets/SKILL.md` — **genuinely relevant, not yet covered above.**
  A Claude-Code-plugin-style skill ("Break a plan, spec, or conversation into
  tracer-bullet tickets... run `/setup-matt-pocock-skills` if not [configured]")
  that can publish tickets to "a real issue tracker (GitHub, Linear, …)... one
  issue per ticket in dependency order." This is the generic ticket-creation
  skill that would run on **cooper** (via Claude Code / matt-pocock-skills
  plugin), not on the Tars VM — it's the other half of "same default for
  Claude/Orca sessions on cooper" from the goal. It currently has no team
  default logic at all; it defers entirely to whatever
  `/setup-matt-pocock-skills` configured. This is the clearest candidate for
  needing the new "default to the personal team" rule on the cooper side.
- `skills/creative/popular-web-designs/templates/linear.app.md` — a design
  reference template (Linear's marketing site as a UI style example) — not
  ticket-tracker-related, no update needed.
- `skills/setup-matt-pocock-skills/SKILL.md` — mentions Linear only insofar as
  it's one of the trackers `to-tickets` can target; worth a quick look if the
  personal-team default needs to be wired in at setup/config time rather than
  per-invocation in `to-tickets` itself, but not read in full for this pass
  (time-boxed; flagging for the follow-up planning step rather than reading
  now).
- All other hits (`mlops/*`, `research/research-paper-writing/*`,
  `creative/ascii-video/*`, `creative/manim-video/*`, `creative/p5js/*`,
  `creative/touchdesigner-mcp/*`, `creative/comfyui/*`,
  `creative/baoyu-infographic/*`, `productivity/pdf/reference.md`,
  `last30days/*`, `autonomous-ai-agents/hermes-agent/references/*`,
  `improve-codebase-architecture/HTML-REPORT.md`) use "linear" in the
  mathematical/generic-English sense (linear layers, linear progression,
  linear regression, linear interpolation, "linearly" as an adverb) — no
  connection to the Linear issue tracker, no update needed.

## Cron facts (from `hermes cron list` on the VM, live)

```
e231e5faf180  Gaetan daily work brief         30 8 * * 1-5     -> slack:C0BP2GZUFSR:1786359613.979759
62e8cd9db637  Gaetan engagement checker       */30 10-16 * * 1-5 -> slack:C0BP2GZUFSR:1786359613.979759
759e08c598e3  Gaetan engagement checker final pass  0 17 * * 1-5 -> slack:C0BP2GZUFSR:1786359613.979759
```
All three deliver to the same Slack thread anchor — confirms "the configured
reporting conversation" in both skills' text is this one thread.

## Surprises / things that don't match the naive assumption

- Repo and VM are perfectly in sync right now for all three files diffed —
  there was no "which side moved" reconciliation needed, contrary to what the
  task framing implied might be found.
- `engagement-checker`'s Linear read is scoped to **issues assigned to the
  authenticated viewer only** (`assignee: {isMe: {eq: true}}`), not "every
  ticket across every team" — the goal's "Tars reads every ticket across every
  team" requirement is a real gap against the current `AssignedIssues` query,
  not just a config tweak.
- `daily-work-brief`'s Linear touchpoint is much weaker than
  engagement-checker's — no hardcoded collector, just "any already-configured
  integration," with an explicit escape hatch to say "Linear coverage
  unavailable." It will likely need the most new code/spec to become a first-
  class Linear puller.
- The "default team for ticket creation" behavior has **zero existing
  ownership** — none of the three skills read here create Linear tickets at
  all today. The only ticket-creating skill found in the whole `~/.hermes/
  skills/` tree is `to-tickets`, which is cooper-side (matt-pocock-skills
  plugin) and defers to `/setup-matt-pocock-skills` config, not to a Tars
  skill.
