---
name: helptech-duty-intake
description: Automate Gaëtan's B2C HelpTech duty intake.
version: 0.1.0
author: Gaëtan Cathelain, Tars
license: MIT
platforms: [linux]
metadata:
  hermes:
    tags: [HelpTech, Linear, Orca, Cron, Orchestration]
    category: orchestration
    related_skills: [linear-ticketing, delegate-to-cooper]
---

# HelpTech Duty Intake

Set up one HelpTech duty day: verify the rota, assign only eligible B2C Slack
Asks in Linear, create an Orca parent with one child per ticket, and keep intake
running until local midnight. This is coordination only; do not start coding
agents or implement the tickets.

## When to Use

- Gaëtan says he is the HelpTech/help-techer/helptaker of the day.
- He asks for a HelpTech Orca worktree, ticket assignment, and same-day polling.
- He wants the previous successful HelpTech setup repeated.

Do not use this skill to compute or announce the duty rotation; that belongs to
the metarepo's `helptaker-du-jour` workflow. Do not include B2B tickets.

## Prerequisites

- Load `linear-ticketing` before every Linear read or write.
- Load `delegate-to-cooper` for Orca access conventions, but do not launch an
  implementation worker: worktrees are the requested coordination artifact.
- Search `~/dev/mc-metarepo` for current HelpTech/Orca operational notes before
  acting. Product source submodules there may be empty; they are not needed.
- Orca must be reachable on `cooper`.
- Default Orca repository is mc-metarepo:
  `id:8099e312-3232-46f2-83a9-97aeaf5de5a2`.
- Incoming HelpTech Slack Asks channel: `CC397R0HY` (`#help-tech`).
- Rotation/announcement channel: `C0BLPCP0APN` (`#help-help-tech`).
- Linear team must be `Mobile club`, team id
  `cb4eb28b-ab4b-4507-b235-a002323bb0d4`.
- Gaëtan's Linear assignee id is
  `4951b192-e49c-4b7e-b491-58c89e66043c`.

## Procedure

### 1. Verify the duty claim

Read the announcement link Gaëtan supplied, or the current day's announcement
in `C0BLPCP0APN`. Require it to name Gaëtan as today's HelpTech owner. Record
the source message timestamp and backup, if present.

If duty cannot be verified, report that and stop before any Linear or Orca
write.

### 2. Get per-message Linear approval

Linear company-team writes require Gaëtan's explicit approval naming the team in
the current conversation. If he has not said **Mobile club**, ask once:

> Confirmes-tu que je peux assigner les tickets B2C HelpTech de l'équipe Linear
> Mobile club à toi aujourd'hui ?

His answer is approval for this duty-day run only. Never treat this skill or a
past run as standing authorization.

### 3. Establish the local-day window

Use `terminal` to read the current Europe/Paris time. Convert local midnight at
the start and end of the duty date to explicit ISO-8601 UTC bounds. For example,
2026-08-11 in CEST is
`[2026-08-10T22:00:00Z, 2026-08-11T22:00:00Z)`.

Never use an unexamined UTC midnight for a France-local duty day. Never process
the whole opening backlog merely because the announcement includes its size.

### 4. Find eligible tickets completely

Call `mcp__linear__list_issues` once per allowed label:

- `B2C (MC)`
- `B2C (Next)`

Use the lower UTC bound as `createdAt`, `includeArchived: false`, `limit: 250`,
`orderBy: createdAt`, and an explicit `fields` projection containing at least
`id`, `title`, `createdAt`, `status`, `statusType`, `labels`, `assignee`,
`assigneeId`, `team`, and `teamId`.

Page unchanged queries until `hasNextPage == false`, with the `linear-ticketing`
cap of eight pages. Apply the upper time bound client-side. Exclude
`completed`, `canceled`, and `duplicate`. If coverage is incomplete, perform no
writes and report it.

Never query or act on B2B labels. Labels alone are not sufficient proof that a
ticket is HelpTech.

### 5. Prove each ticket came from the HelpTech Slack Ask flow

Call `mcp__linear__get_issue` for every candidate. Require all of:

- team is `Mobile club` and the team id matches the prerequisite;
- an attachment title begins with `Ask from `;
- that attachment URL contains `/archives/CC397R0HY/`;
- created time is inside the local-day window;
- status is non-terminal;
- at least one allowed B2C label is present and no B2B label is present.

Skip any candidate that fails one condition. The exact Slack attachment is the
source proof; descriptions, similar titles, and triage suggestions are not.

### 6. Assign safely

For each verified ticket:

- unassigned: call `mcp__linear__save_issue` with only `id` and
  `assignee: 4951b192-e49c-4b7e-b491-58c89e66043c`;
- already assigned to Gaëtan: leave it unchanged;
- assigned to someone else: do not overwrite; record a conflict for Gaëtan.

Omit `labels`, `state`, priority, description, and every unrelated field so the
update preserves them. A write succeeds only when the parsed response carries
Gaëtan's exact `assigneeId`; re-read when needed.

### 7. Create the Orca parent once

Run Orca preflight on cooper:

```bash
ssh cooper '~/.local/bin/orca status --json'
ssh cooper '~/.local/bin/orca worktree list --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json'
```

Reuse an existing active root worktree named `HelpTechs`. If absent, create it:

```bash
ssh cooper '~/.local/bin/orca worktree create \
  --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 \
  --name "HelpTechs" --no-parent --setup run --activate --json'
```

Capture the returned worktree id, instance id, path, branch, base ref, and head.
Require `parentWorktreeId: null`.

### 8. Ensure exactly one child per eligible ticket

Before each create, inspect the full worktree list. Treat either an exact
`linkedLinearIssue` match or the expected child name/path as an existing child;
verify it belongs to `HelpTechs`. Do not duplicate a partially completed run.

For an assigned ticket with no child, run:

```bash
ssh cooper '~/.local/bin/orca worktree create \
  --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 \
  --name "<ISSUE_ID>" \
  --linear-issue "<ISSUE_ID>" \
  --parent-worktree "<HELPTECHS_WORKTREE_ID>" \
  --setup run --json'
```

Require the returned `linkedLinearIssue` to equal the ticket id and
`parentWorktreeId` to equal the captured `HelpTechs` id. Orca children are
siblings on disk; parentage is Orca lineage, not directory nesting or git
ancestry.

### 9. Poll until local midnight

Create a Hermes cron every five minutes. Its prompt must be self-contained and
carry:

- the exact local-day UTC bounds;
- the two allowed B2C labels and explicit B2B exclusion;
- the Slack Ask proof for `CC397R0HY`;
- Mobile club and Gaëtan ids;
- the captured HelpTechs worktree id and repository id;
- complete Linear paging rules;
- assignment conflict behavior;
- pre-create duplicate checks and post-create lineage verification;
- prohibition on code changes and implementation agents;
- local-only delivery with a concise delta/coverage record.

Attach `linear-ticketing` and `delegate-to-cooper` to the cron. Keep each run
bounded for Hermes' three-minute cron limit.

Then create a separate one-shot cron at the next Europe/Paris midnight. Embed
the poller's returned job id and instruct the one-shot job to pause that exact
job, list jobs, and verify it is disabled. Deliver only the shutdown verdict to
the originating conversation. A prose promise inside the poller is not a
shutdown mechanism.

Trigger one immediate manual poller run with known tickets named in the run
context. Its success criterion is idempotence: no assignments or worktrees are
created for already-complete tickets.

### 10. Verify and report

Re-read Linear and Orca independently. Report:

- duty-source message;
- eligible ticket ids and their source Slack Ask links;
- confirmed assignee ids;
- parent path/branch/id;
- each child path/branch, Linear link, and parent id;
- poller and shutdown cron job ids, schedules, timezone, and enabled state;
- immediate-run coverage, conflicts, and duplicate counts.

Lead with the verdict. Include exact Orca commands used. State explicitly that
no B2B ticket and no implementation agent was included.

## Pitfalls

- The announcement's opening queue count is not the set created today.
- `B2C (Next)` or `B2C (MC)` does not by itself prove a HelpTech Ask.
- Linear `query` search is fuzzy and cannot prove complete absence.
- An unassigned issue may omit the `assignee` key entirely.
- `labels` on `save_issue` replaces all labels; omit it for assignment-only
  updates.
- Parallel reruns can race. List immediately before worktree creation and verify
  after it; if a duplicate appears, report it rather than deleting anything.
- Do not infer France-local dates from UTC or from the scheduler's display.
- Do not remove worktrees or branches automatically after duty. Cleanup is a
  separate, potentially destructive request.

## Verification Checklist

- [ ] Gaëtan's duty is source-verified.
- [ ] Current-message approval explicitly names Mobile club.
- [ ] Europe/Paris date bounds are explicit and correct.
- [ ] Both allowed B2C scans end with `hasNextPage: false`.
- [ ] Every acted-on ticket has a Slack Ask attachment from `CC397R0HY`.
- [ ] No B2B or terminal ticket was acted on.
- [ ] Every assignment response confirms Gaëtan's assignee id.
- [ ] `HelpTechs` is a root Orca worktree.
- [ ] Exactly one linked child exists per eligible assigned ticket.
- [ ] Every child points to `HelpTechs` through `parentWorktreeId`.
- [ ] Poller is enabled every five minutes and delivers locally.
- [ ] One-shot shutdown is enabled for next local midnight.
- [ ] Immediate run is idempotent and coverage-complete.
- [ ] No implementation agent was launched.
