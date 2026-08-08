---
name: daily-work-brief
description: "Use for concise workday dailies from all activity sources."
version: 1.0.0
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

This workflow's daily collection, analysis, judgment, and drafting are Tars's own work. Run those steps directly: never delegate them to a subagent, Claude, Orca, an Orca worktree/session, or another agent. External testing of the skill may be delegated when Gaetan explicitly asks for it. A scheduled run has a tight runtime, so collect independent sources in parallel and keep raw results out of the final message.

## 1. Establish the window

Use `Europe/Paris` for all dates and labels.

1. Find the timestamp of the last **successful daily-work-brief delivery** in the origin Slack conversation or cron history. The window starts immediately after it.
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
- Use any already-configured Linear integration, CLI, API, email, or Slack links to identify tickets completed/merged in the window. Count only explicit terminal states and preserve ticket IDs. If no integration is reachable, omit the metric and add `Linear coverage unavailable`—do not estimate from branch names.

Completion criterion: every numeric stat is reproducible from ledger rows with explicit identifiers and terminal state.

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

If the gate returned `NO_DAILY`, produce no user-facing report. Otherwise use this compact shape:

**Since the last daily**
- Outcome-first bullets grouped by project/topic. Fold related Slack help, Claude work, commits, PRs, and tickets into one bullet.

**Stats**
- `PRs merged: N` and compact identifiers.
- `Linear tickets completed: N` and IDs, only with explicit evidence.
- At most two other useful counts (reviews, access/help requests resolved, incidents), never vanity counts such as shell commands or messages sent.

**Today**
1. The highest-leverage next action.
2. Follow-ups, with owner and reason.
3. Work already in motion to resume.
4. `Oli:` report/no-report recommendation with one-line rationale when relevant.

End with `Coverage: ...` only for a missing or partial mandatory source. Do not include cron headers, job IDs, collection logs, source dumps, or a generic motivational close. Keep it under 350 words unless the evidence genuinely cannot fit. Use cautious wording for inferred outcomes and omit noise.

Completion criterion: every bullet maps to ledger evidence; stats reconcile; every “Today” item is actionable; the message reads as a brief, not a transcript.

## 6. Scheduled-run behavior

The cron prompt must be self-contained and attach this skill. Delivery should return only the brief. On `NO_DAILY`, remain silent. On a source or run failure that prevents a trustworthy daily, send one concise failure message naming the failed source and the next repair action—no cron wrapper prose and no fabricated substitute.
