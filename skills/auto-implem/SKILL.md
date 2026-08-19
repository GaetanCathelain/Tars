---
name: auto-implem
description: Run a Linear ticket through verified closure.
version: 0.1.0
author: Gaetan Cathelain, Hermes Agent
license: MIT
platforms: [linux]
metadata:
  hermes:
    tags: [linear, orchestration, delegation, orca, verification]
    related_skills: [linear-ticketing, delegate-to-cooper]
---

# Autonomous Ticket Lifecycle

Own one named Linear ticket from intake to a verified outcome. Tars coordinates, delegates implementation, keeps Linear current, verifies the real source, cleans up execution resources, and reports the verdict. This skill does not weaken SOUL rules: implementation belongs to a worker, deletion still needs a named target and the required confirmation beat, and agent self-reports are never proof.

## When to Use

Use when Gaetan says to start, handle, finish, or take ownership of a named Linear ticket and expects end-to-end follow-through.

Typical triggers:

- “Start working on GCN-84.”
- “Take this ticket through completion.”
- “Delegate this on Cooper and keep Linear updated.”
- “Do the ticket, clean up the worktree, and report back.”

Do not use for:

- creating or editing a ticket without executing its work;
- read-only status questions about work Tars did not initiate;
- a multi-ticket program that needs a separate plan or dependency graph;
- daily briefs or engagement reminders.

## Prerequisites

Load and follow:

- `orchestration/linear-ticketing` for Linear lookup, comments, states, and read-back;
- `delegate-to-cooper` when the ticket requires an implementation deliverable;
- the task-specific skill when one matches the ticket domain.

Required access depends on the ticket:

- connected Linear tools;
- `ssh cooper` and the Orca CLI for delegated implementation;
- authenticated source tools needed for verification, such as Notion, GitHub, Slack, or a browser.

Before delegating, search Tars’s knowledge base at `~/dev/mc-metarepo`. Its submodules may be empty; do not pretend product source is present there.

## Operating Contract

1. The current message defines the current action. Do not resume an old ticket merely because it appears in thread history.
2. Tars writes the brief, tracks the worker, verifies the result, updates Linear, manages cleanup, and reports. The worker produces implementation deliverables.
3. A worker’s “done” message is a claim. Re-read the changed source and its acceptance evidence before closing Linear.
4. Keep Gaetan’s DM and Linear aligned with actual state. Never call a ticket done while Linear still says otherwise or while required cleanup is unverified.
5. Never expose credentials, tokens, keys, or secret-bearing evidence in Linear or Slack.

## Procedure

### 1. Establish ticket truth

1. Retrieve the exact issue and its comments, relations, status, assignee, project, labels, acceptance criteria, and linked sources.
2. If the request names a ticket or project but the identifier is ambiguous, search all Linear teams assigned to Gaetan; do not assume GCN.
3. Read the source referenced by the ticket when accessible. A `Source:` Slack timestamp is context, not proof of current conversation state.
4. Search `~/dev/mc-metarepo` for ticket, system, and environment context before writing a delegation brief.
5. Check for an existing active run, branch, PR, worktree, or duplicate ticket before creating anything.

Completion criterion: the exact ticket, current status, acceptance criteria, source handles, and existing execution artifacts are known.

### 2. Decide who performs the work

- If the requested deliverable is analysis or a report and Tars can produce it through read-only inspection, Tars does it directly.
- If the ticket requires code, files, configuration, migration, documentation in a product or repository, production mutation, or any other implementation artifact, delegate it.
- If verifying a hypothesis requires a mutation, delegate that investigation too.

Completion criterion: ownership is explicit and SOUL rule 1 is satisfied.

### 3. Prepare a bounded worker brief

The brief must include:

- exact ticket ID, title, source URL, and acceptance criteria;
- authoritative context already found;
- allowed repositories, systems, and mutation scope;
- required tests or read-back checks;
- prohibited unrelated changes;
- the evidence the worker must return;
- instruction to report through Orca orchestration and not close Linear itself.

Do not paste secrets. Do not ask the worker to decide ticket closure or cleanup policy.

Completion criterion: another agent can execute without guessing scope, proof, or handoff format.

### 4. Dispatch on Cooper

Follow `delegate-to-cooper` exactly. In particular:

1. Preflight Orca status, accounts, repository, worktrees, and the coordinator terminal.
2. Create a Run with `--from <coordinator-terminal>`.
3. Create the task with both `--run <run-id>` and `--from <coordinator-terminal>`.
4. Start the worker with both `--run` and `--from`; omitting them can produce `selector_not_found` even when the task exists.
5. `/auto-implem` work on Cooper always uses Claude **Opus 4.8**. Explicitly pass Orca's Opus 4.8 model selector to `worker-start`; never inherit Orca's default model. If Orca cannot resolve Opus 4.8 exactly, stop before dispatch and report the blocker.
6. Verify both `launch.requested` and `launch.effective` identify Claude Opus 4.8. Any fallback or different effective model is a failed dispatch, not an acceptable substitute.
7. Record run ID, task ID, dispatch ID, worker terminal handle, exact worktree ID/path, branch, and model.

Completion criterion: `worker-start` returns `state=ready`, `stage=input_accepted`, and both requested and effective launch are Claude Opus 4.8.

### 5. Prove the agent actually started

`input_accepted` can still leave a pasted prompt waiting at Claude’s `❯` prompt.

1. Read the worker terminal after dispatch.
2. If the full brief is visible but there is no assistant response or tool activity, send one Enter with `orca terminal send --enter`.
3. Read again or wait briefly for terminal activity.
4. Treat a worker sentence, tool call, thinking indicator, or changed terminal cursor/output as start evidence.

Do not repeatedly press Enter. One accepted byte plus visible agent activity is enough.

Completion criterion: the worker is visibly reasoning or using tools, not merely displaying the prompt.

### 6. Update Linear at kickoff

After the worker is genuinely active:

1. Move the ticket to `In Progress` if it is not already started.
2. Add a concise comment containing the model, run/task/dispatch handles, worker terminal, exact worktree path, and current state.
3. Re-read the ticket and comments. Quote only non-secret evidence.

If dispatch fails, do not leave the ticket falsely marked In Progress without a blocker comment.

Completion criterion: Linear’s state and comment match the live dispatch.

### 7. Track the right completion signal

Poll exact handles, not names or UI labels.

Use these as authoritative orchestration completion fields:

- `dispatch.status=completed`;
- `worker.state=succeeded` or a terminal failure state;
- `worker.stage=settled`;
- task `status=completed` with a worker report.

Do **not** gate completion on `observation.status`. It can remain `running` after success when the terminal resource is `user_owned` with `retainedReason=user_takeover`. A monitor that watches only `observation.status` will time out after a successful job.

If using a background monitor, stop when the dispatch or worker reaches a terminal state. Give it a bounded timeout and report timeout honestly. Do not create an endless sleep loop.

Post material blockers or scope changes to Linear as they happen, then re-read the comment. Do not narrate routine polling.

Completion criterion: the exact task has settled, or a specific blocker is recorded with evidence.

### 8. Verify the deliverable independently

Read the real changed system after the worker settles.

Examples:

- Notion: retrieve the page after the change and quote the exact rendered entry.
- GitHub: inspect the PR diff, checks, branch, and merge state.
- Repository: inspect the diff and run the requested tests without modifying the worker’s deliverable.
- Production or infrastructure: query the actual service or control plane and check the intended state.
- Slack or another conversation: fetch the thread and surrounding history before reporting what happened.

Map each acceptance criterion to one source-backed fact. If an item cannot be verified, mark it `not verified`; do not close the ticket as Done merely because the worker says it succeeded.

Completion criterion: every acceptance criterion is either proven or explicitly blocked.

### 9. Settle Linear

When verification succeeds:

1. Add a completion comment with the outcome, exact source evidence, run/task/dispatch handles, tests or read-back, repository/PR state, and cleanup state.
2. Move the issue to `Done` only after the completion comment has landed.
3. Re-read the issue and comments. Verify `status=Done`, `statusType=completed`, `completedAt`, and the exact completion comment.

When verification fails or work is blocked:

- keep the truthful non-terminal status;
- comment with the blocker and evidence;
- do not invent a successful outcome.

Completion criterion: Linear reflects the verified result and contains a durable evidence trail.

### 10. Clean up execution resources

Cleanup scope must be explicit.

1. Identify the exact terminal handle, worktree ID/path, and branch.
2. A request to stop or close a session is not a request to delete a branch. Close or release the terminal without deleting unrelated artifacts.
3. Worktree removal is destructive. Name the exact path before removal. If Gaetan previously asked for deletion and Tars confirmed it in the same conversation, quote that instruction and obtain the required one-line confirmation before deleting.
4. If confirmed, remove only the named worktree. Use `--force` only when Gaetan explicitly authorizes forced removal or normal removal is blocked and he names that remedy.
5. Do not delete the branch, PR, Run, task, or other worktrees unless Gaetan named them too.

Verify worktree removal three ways:

- Orca lookup returns `selector_not_found`;
- the exact filesystem path is absent;
- `git worktree list --porcelain` no longer contains it.

A retained terminal after `user_takeover` does not mean the task failed; it means cleanup still needs explicit handling.

Completion criterion: every authorized resource is absent or released, and every unauthorized resource is untouched.

### 11. Report the verdict

The first line is the outcome. Stay within five lines unless Gaetan requested a report.

Include only useful handles:

- ticket link and final Linear state;
- one-line deliverable result;
- source-backed verification;
- PR or deployment handle when relevant;
- worktree cleanup result, or the exact remaining action requiring Gaetan.

Do not echo the worker brief or full command log unless asked.

## Failure Handling

- Three failed attempts on one artifact: stop and report the exact state and options.
- A second research round with no new fact: stop researching.
- Agent dispatch accepted but no activity: inspect the terminal and send one Enter if the prompt is staged.
- Worker succeeded but monitor timed out: inspect `dispatch.status`, `worker.state`, and `worker.stage`; ignore stale `observation.status`.
- Linear write returns no error but no state change: re-read, report the no-op, and retry only with corrected identifiers.
- Cleanup fails: keep Linear’s implementation result truthful, record cleanup as pending, and never claim the worktree is gone.
- Source verification is inaccessible: say what call failed and keep the ticket non-terminal unless Gaetan explicitly decides otherwise.

## Verification Checklist

- [ ] Exact ticket and acceptance criteria read
- [ ] Existing work and duplicates checked
- [ ] Knowledge base searched before delegation
- [ ] Ownership classified correctly
- [ ] Effective model/context/effort verified
- [ ] Worker visibly started, not just `input_accepted`
- [ ] Linear kickoff status and comment read back
- [ ] Completion judged from dispatch/worker/task, not stale observation
- [ ] Every acceptance criterion independently verified
- [ ] Completion comment read back
- [ ] Linear `Done` state read back
- [ ] Exact cleanup target named and authorized
- [ ] Worktree absence verified through Orca, filesystem, and Git
- [ ] Final verdict gives status, evidence, and handles without logs
