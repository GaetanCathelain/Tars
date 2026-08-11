---
name: daily-work-brief
description: "Use for concise workday dailies from all activity sources."
version: 1.5.0
required_environment_variables: [LINEAR_API_KEY]
metadata:
  hermes:
    tags: [daily, status, slack, github, linear, cooper, payfit]
    category: orchestration
---

# Daily work brief

Produce Gaetan's source-backed morning daily: everything material since the last successful daily, compact stats, and a judgmental plan for today. The unit is work accomplished and coordination moved forward—not an activity log.

## Load only the branches used

- Load `stakeholder-communication` before synthesis.
- For Gmail and Calendar, load `google-workspace`; if its auth check fails, check the already-configured `himalaya` skill before treating email as unavailable.
- Load a GitHub skill only if `gh --help` is insufficient for the read-only query.
- Load `hermes-agent` only when installing or troubleshooting the cron job.
- Cooper inspection is read-only analysis. Do it directly over `ssh cooper`; do not spawn a coding agent merely to read activity.

Tars owns this workflow and its final judgment. It may parallelize collection and bounded analysis through Hermes/Tars-native subagents or workflows, while Tars verifies and synthesizes the final daily. Claude and Orca are evidence sources only: Tars may read Claude histories and inspect Orca sessions/worktrees/state, but must never spawn, prompt, resume, modify, or delegate work through Claude, Orca, an Orca worktree/session, or another non-Tars agent system. External testing of the skill may also use Hermes/Tars-native subagents when Gaetan asks for it. A scheduled run has a tight runtime, so collect independent sources in parallel and keep raw results out of the final message.

## 1. Establish the window

Use `Europe/Paris` for all dates and labels.

1. Find the timestamp of the last **successful daily-work-brief delivery** in the configured Slack reporting conversation or cron history. The window starts immediately after it.
2. If there is no prior daily, use 08:30 Europe/Paris on the previous actual workday. On a Monday this normally covers Friday 08:30 through Monday morning, including weekend work signals.
3. End at the current instant. State the window only when the fallback was used or source coverage is partial.
4. Deduplicate the same work mentioned by several sources. Prefer the strongest source: merged PR or ticket state over shell command; direct Slack statement over inference.

Completion criterion: every source query uses the same explicit start/end instants, with date-only APIs widened then filtered locally by timestamp.

## 2. Decide whether today is a daily day

Check all three gates before doing the expensive collection:

1. **Calendar:** today is Monday–Friday in Europe/Paris.
2. **France holiday:** query the official metropolitan France calendar at `https://calendrier.api.gouv.fr/jours-feries/metropole/<YEAR>.json`; match the Paris-local ISO date.
3. **PayFit leave:** search existing email access for messages from or about PayFit that explicitly cover today's date. Read the relevant message body; a subject/snippet alone is not proof. Accepted/approved leave, RTT, or absence is off evidence. A generic PayFit notification is not.

If it is a weekend, France bank holiday, or explicit PayFit day off, make one bounded activity probe (Gaetan-authored Slack, merged PRs, and Cooper Claude activity). If there is no substantive work signal, return exactly `[SILENT]`; Hermes suppresses delivery for that marker. If there is substantive work, continue: “actually worked” overrides the nominal day-off gate.

If holiday or leave evidence cannot be checked, continue rather than silently skipping, and add one short coverage note at the end.

Completion criterion: a skip is supported by both a day-off reason and absence of substantive work signals.

## 3. Collect the evidence ledger

Build a private ledger with columns `time | source | outcome | people | follow-up | evidence handle`. Never place credentials, tokens, private key material, or raw private transcript dumps in it.

### Slack — exhaustive within the window

Use Slack search with `filter_users_from=U08BDJAMSRZ` and the widened date range. Paginate until `next_cursor` is empty. For every hit, recover enough surrounding channel/DM/thread context with conversation history or replies to understand what Gaetan did and whether anyone owes a response. Include DMs, group DMs, public/private channels, thread replies, and access/help actions (for example Tailscale setup or AWS access), not only project channels.

Also search messages addressed to or discussing Gaetan when needed to identify unresolved asks. Resolve people by Slack lookup; do not guess display-name identities. Keep sensitive operational details generalized in the final brief.

Completion criterion: all pages of Gaetan-authored messages are covered and every material conversation has its outcome or open loop recorded.

### Cooper — Claude, shell, repos

Run read-only commands over `ssh cooper`:

- inspect Orca's registered repos/worktrees and current agent/session state;
- locate Claude Code transcript files modified in the window under the user's Claude data directories, then read only the bounded relevant turns needed to identify goals, decisions, implementation outcomes, tests, and pending work;
- inspect timestamped shell history when timestamps exist; treat untimestamped `.bash_history` only as supporting context, never as proof that a command fell in the window;
- query git logs across Orca-registered repos for Gaetan-authored commits/merges in the window and inspect relevant branch/PR state.

Exclude secret files and credential contents. Commands that merely install, inspect, or navigate are evidence of an outcome only when corroborated.

Completion criterion: each active Claude session and authored git outcome in the window is represented once; unsupported shell-history timing is labelled or omitted.

### GitHub and Linear

- With the existing `gh` authentication, list PRs authored by Gaetan that were merged in the window. Record repository, PR number/title, merge time, and count. Also record reviews/comments only when they materially moved work.

#### The board block is produced by a script. You do not write it.

Run this LAST, after every other source is collected, immediately before §5 is written:

```bash
python3 ${HERMES_HOME:-$HOME/.hermes}/skills/orchestration/daily-work-brief/scripts/linear_board.py
```

**Run it through the shell/terminal tool — never `execute_code`.** Measured: the `execute_code` sandbox scrubs every environment variable whose name contains `KEY`, so `LINEAR_API_KEY` is absent there and the script can only fail; the terminal tool keeps it. That is also why the frontmatter declares `required_environment_variables`. Never put the key on a command line, never read it, never echo it — the script takes it from its own environment and no argument.

The script prints the finished block and nothing else: the two views (every non-terminal GCN issue; every non-terminal issue assigned to Gaetan on any team), deduplicated on the issue key, priority-sorted with GCN first inside a priority, capped at 12 rows with a **computed** `+N more`, closed by a mandatory `Coverage:` line (`Board: clear.` plus that line when nothing is open — that is a healthy run, not a failure) — or, if either view failed, errored, or could not be paged to completion, exactly the single line `board unavailable (coverage unproven)` and a non-zero exit. There is no partial board and no third outcome. **Whatever it printed is what goes in the brief, byte for byte** (§5).

**Why a script and not a tool call — do not "simplify" this back into the prompt.** On 2026-08-10 a live run rendered the board from the agent's own recall after context compacted twice: 7 of 12 titles were invented or paraphrased, two rows were terminal issues, and `+2 more` stood against a true remainder of ~46 (`status/probes/wf6/deploy-and-verify.md`). Every structural check passed. The fix is that no board row passes through the model at all.

**This script is the one raw-HTTP exception in the system, and it is scoped to this block only.** Every other Linear read below, and every Linear read and write in every other skill, stays on the native `mcp__linear__*` tools per `linear-ticketing`. The reason is measured, not preference: no `hermes` command invokes a named MCP tool outside the LLM agent loop — `hermes mcp` has no `call`, and `mcp test` takes a server name and nothing else (`status/probes/wf6/deterministic-render-options.md`). A deterministic step therefore cannot use the native tool. If a future Hermes ships a non-agentic MCP invocation, retire the script and move the block back to `mcp__linear__*`.

#### `Linear tickets completed: N` — one native read

The board script does not produce this stat. Read it with the native `mcp__linear__list_issues` tool — a tool call, never a script, a raw HTTP request, an ad-hoc integration, an email guess, or a Slack link. Load `linear-ticketing` for the team id, the measured response shape (§9) and the `fields` enum (§8). Its measured facts bind this read: the tool result is a JSON **string** under `result` and must be parsed a second time; **`id` is the issue key (`GCN-42`) — there is no `identifier` field**, and `identifier`, `state`, `stateId`, `labelIds`, `cycle`, `relations`, `comments`, `attachments` and `documents` are **hard errors that kill the entire call**; **empty values are field-specific** — `assignee` is omitted entirely when unset, `completedAt` comes back present and `null`. Read every field defensively and never use key presence as a value test.

**Completed in window** — `assignee` `4951b192-e49c-4b7e-b491-58c89e66043c` (Gaetan), `state` `"completed"`, `updatedAt` set to the window start as a **bare string** (a lower bound; the object form `{"gte": …}` is a hard error), `limit` 250, `fields` `["id","title","completedAt"]`; keep only issues whose `completedAt` **value** falls inside the window. Test the value, never the key: with this projection **every** row carries `completedAt`, `null` on the ones that are not completed, so a `"completedAt" in row` test keeps all of them. There is no upper time bound and no `completedAt` parameter — that filter is client-side. Titles only, never descriptions or comments, so nothing beyond the length cap has to be redacted.

**Assert `hasNextPage == false` before using it.** It is the only completeness signal: there is no total count and no truncation flag, the default page is 50 and `limit` maxes at 250 (251 errors). If it returns `true`, re-issue the identical call with `cursor` set to the returned value until it is false, unioning and deduplicating on `id`, bounded by a page cap of **8 pages** (`linear-ticketing` §8). If the cap is reached with `hasNextPage` still true, name the stat incomplete in `Coverage:` rather than printing a number you cannot justify. Never narrow the filter to make `hasNextPage` go false — narrowing proves a different, smaller read complete. Do not reach for `query`: it is fuzzy and its `hasNextPage` is always `true`. If this read fails, drop the `Linear tickets completed` line and add `Linear completed-count unavailable` to `Coverage:`; never estimate it from branch names, and never invent an id.

Completion criterion: every numeric stat is reproducible from ledger rows with explicit identifiers and terminal state, and the board block in the delivery is byte-identical to the script's stdout.

### Email and calendar

Using existing authenticated access, inspect sent/received work mail and calendar events in the window when they add work outcomes or follow-ups. PayFit leave checking is mandatory; broader email is supporting evidence. Never send or modify anything.

## 4. Synthesize judgments

For each ledger row, classify it as:

- **Done:** an outcome landed.
- **Moved:** coordination/help/review advanced but is not fully done.
- **Open loop:** Gaetan or someone else owes a response, review, access check, decision, or report.
- **Noise:** automation, repeated alerts, passive reads, and commands without a work outcome.

Then decide today's plan from evidence, in this order:

1. time-sensitive open loops and promises;
2. blockers Gaetan can clear for someone else;
3. unfinished work already in motion on Cooper/Claude/PRs;
4. planned Linear/Slack/calendar work;
5. reporting or escalation to Oli when a meaningful result, risk, delay, or decision warrants it.

Name who to follow up with and why. Do not invent a priority from weak activity signals. Separate “someone owes Gaetan” from “Gaetan owes someone.”

## 5. Write the delivery

If the §2 gate returned `[SILENT]`, produce no user-facing report. Otherwise use this compact shape, board first. The brief OPENS with §3's board block — the script's stdout, unedited — and looks like this:

**Board** (priority-sorted)
P1 GCN-42 Ship the thing — In Progress
P2 MC-4179 Review the other thing — Todo
+7 more
Coverage: 2 views, both complete (58 issues)

- **Paste the script's stdout byte for byte, as the first element of the brief.** No row may be written, corrected, re-ordered, re-titled, shortened, re-typed, summarised, merged, annotated or filtered by you. Not one character. The sort, the 12-row cap, the computed `+N more`, the terminal-row drop and the `Coverage:` line are the script's output and are already correct; changing any of them can only make the board wrong. You narrate **around** the block, never inside it.
- **If the script printed `board unavailable (coverage unproven)`, print that one line as the board and nothing else** — no rows from a previous run, no rows from memory, no partial list, no explanation stitched into it. Add the reason to the brief's own `Coverage:` line at the end if you know it.
- **If you do not have the script's stdout in front of you — it was never run, the output scrolled away, or the context compacted — run it again.** One shell call. Never reconstruct a board row from recall, a branch name, a Slack mention, or an issue you remember working on. If the exact bytes are not in front of you, the board does not exist.
- The board block is excluded from the ~350-word prose cap below—it must never squeeze out the judgment sections.

**Since the last daily**
- Outcome-first bullets grouped by project/topic. Fold related Slack help, Claude work, commits, PRs, and tickets into one bullet.
- **`Claude/Orca:`** one required labelled line closing the block: the Cooper Claude sessions and Orca worktrees that moved in the window, with their evidence handles, whenever §3's Cooper collection produced anything. `Claude/Orca: none.` when it produced nothing. It is a named line, not an optional fold — the brief must be checkable for it.

**Stats**
- `PRs merged: N` and compact identifiers.
- `Linear tickets completed: N` and IDs, from the completed-in-window read only.
- At most two other useful counts (reviews, access/help requests resolved, incidents), never vanity counts such as shell commands or messages sent.

**Today**
1. The highest-leverage next action.
2. Follow-ups, with owner and reason.
3. Work already in motion to resume.
4. `Oli:` report/no-report recommendation with one-line rationale when relevant.

End with `Coverage: ...` only for a missing or partial mandatory source. Do not include cron headers, job IDs, collection logs, source dumps, or a generic motivational close. Keep the prose under 350 words, board block excluded, unless the evidence genuinely cannot fit. Use cautious wording for inferred outcomes and omit noise.

Completion criterion: every bullet maps to ledger evidence; stats reconcile; every “Today” item is actionable; the message reads as a brief, not a transcript.

## 6. Scheduled-run behavior

One recurring weekday job at 08:30 Europe/Paris, with this skill attached and delivery to the configured reporting conversation. The job is reconciler-managed on the VM; reconcile it against this prompt and do not let the two drift.

> Run `daily-work-brief` end to end. Use Europe/Paris. Establish the window from the last successful delivery, then apply the workday gate — weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off — and return exactly `[SILENT]` when the gate says not to run. Otherwise collect Slack, Cooper (Orca state, Claude transcripts, shell history, git), GitHub, Linear, email and calendar evidence in parallel, keep raw results out of the message, and deliver the compact brief. Do Linear LAST. The brief must OPEN with the board block, and **the board block is the stdout of `python3 ${HERMES_HOME:-$HOME/.hermes}/skills/orchestration/daily-work-brief/scripts/linear_board.py`, run through the shell/terminal tool (never `execute_code`, which scrubs `LINEAR_API_KEY`) and included byte for byte.** Do not write, correct, re-order, re-title, filter or re-type a single row: the script already sorted, capped, computed `+N more` and emitted the `Coverage:` line. If it printed `board unavailable (coverage unproven)`, print exactly that one line as the board and nothing else. If its output is not in front of you — never run, scrolled away, or the context compacted — run it again rather than recalling rows; a fabricated board is worse than no board. Take `Linear tickets completed: N` from a separate native `mcp__linear__list_issues` read (`state` `"completed"`, `updatedAt` at the window start, `fields` `["id","title","completedAt"]`), proven complete by paging `cursor` until `hasNextPage == false` — never by narrowing the filter. `Since the last daily` must end with the labelled `Claude/Orca:` line covering the Cooper Claude/Orca history, or `Claude/Orca: none.` Every source is read-only: write nothing to Linear, and do not spawn, prompt, or delegate to another agent. Return only the brief.

Delivery should return only the brief. On `[SILENT]`, remain silent. On a source or run failure that prevents a trustworthy daily, send one concise failure message naming the failed source and the next repair action—no cron wrapper prose and no fabricated substitute.
