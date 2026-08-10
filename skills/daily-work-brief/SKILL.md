---
name: daily-work-brief
description: "Use for concise workday dailies from all activity sources."
version: 1.3.0
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

Linear is read with the native `mcp__linear__list_issues` tool — a tool call, never a script, a raw HTTP request, an ad-hoc integration, an email guess, or a Slack link. No endpoint, no key, no cursor loop belongs in this skill. Load `linear-ticketing` for the team id, the priority integers, the priority sort rule, the measured response shape (§9) and the `fields` enum (§8). Five of its measured facts bind every read here: the tool result is a JSON **string** under `result` and must be parsed a second time; **`id` is the issue key (`GCN-42`) — there is no `identifier` field**; **there is no `state` object** — the state is flat `status` (display name) plus `statusType`; **`priority` comes back as an object** `{"value":3,…}` and is normalised to an int before any sort or render (`linear-ticketing` §2); and **empty values are field-specific** — `assignee` is omitted entirely when unset, `completedAt` comes back present and `null`, `labels` present and `[]`. Read every field defensively for absent, null and empty-collection alike, and never use key presence as a value test.

Project every read with `fields` — an unprojected large read exceeds Hermes' own tool-result budget and never reaches the agent — and request only measured-valid members. `identifier`, `state`, `stateId`, `labelIds`, `cycle`, `relations`, `comments`, `attachments` and `documents` are **hard errors that kill the entire call**. Non-terminal means `statusType ∈ {backlog, unstarted, started}`. **Never build a negated (`"!Done"`) or comma-list (`"Todo,In Progress"`) `state` filter, and never pass `state` an array**: an array is a hard error, and the other two return a normal response carrying zero issues — an empty board that looks legitimate. Three reads:

1. **GCN board** — `team` `81e7b769-2a46-4e2a-8db5-c165a7963b0e` (GCN), `limit` 250, `fields` `["id","title","priority","status","statusType","teamId"]`, **one call per non-terminal type**: `state` `"started"`, then `"unstarted"`, then `"backlog"`. Union the three. Scoping by type keeps every Done and Canceled GCN issue — including every issue `engagement-checker` ever filed and closed — out of the page budget entirely, so the board does not start failing its own completeness test the month GCN crosses 250 lifetime issues. Do not read the team unfiltered and drop terminal rows client-side.
2. **Cross-team assigned** — `assignee` `4951b192-e49c-4b7e-b491-58c89e66043c` (Gaetan), no `team`, same `fields`, `limit` and three type-scoped calls. `state` is single-valued but accepts a state *type*, and the type filter is measured to work cross-team, so no other team's state ids are needed. `"started"` and `"unstarted"` are measured; `"backlog"` is the same enum but was never exercised, and a bad `state` value returns zero issues with no error — if a backlog call returns zero on a day the other reads show backlog work, name it in `Coverage:` rather than reading it as an empty backlog. An unfiltered assignee-wide read is not a substitute — it overflows one page.
3. **Completed in window** — `assignee` Gaetan, `state` `"completed"`, `updatedAt` set to the window start as a **bare string** (a lower bound; the object form `{"gte": …}` is a hard error), `limit` 250, `fields` `["id","title","completedAt"]`; keep only issues whose `completedAt` **value** falls inside the window. Test the value, never the key: with this projection **every** row carries `completedAt`, `null` on the ones that are not completed, so a `"completedAt" in row` test keeps all of them. There is no upper time bound and no `completedAt` parameter — that filter is client-side. This feeds `Linear tickets completed: N` with real data.

**Assert `hasNextPage == false` on every read before using it.** It is the only completeness signal: there is no total count and no truncation flag, the default page is 50 and `limit` maxes at 250 (251 errors). If a read returns `true`, **page it**: re-issue the identical call with `cursor` set to the returned value and repeat until `hasNextPage` is false, unioning the pages and deduplicating on `id`, bounded by an explicit page cap of **8 pages per call** (`linear-ticketing` §8). If the cap is reached while `hasNextPage` is still true, the read is **incomplete**: render what arrived and name it in `Coverage:` as a partial board. Never render a partial board as if it were whole.

**Never narrow a read to make `hasNextPage` go false** — no `updatedAt` bound on reads 1 and 2, ever. Narrowing changes the question, so the retry's `false` proves a different, smaller read complete: every non-terminal issue outside the narrowed bound vanishes and the brief asserts a whole board while doing it. Read 3 is the only read with an `updatedAt` bound, and there it is the filter the read is *for*, not a remedy. Do not reach for `query` here either: it is fuzzy, and its `hasNextPage` is always `true`.

Select titles only — never descriptions or comments — so nothing beyond the length cap has to be redacted; that is a deliberate assumption, not an oversight. Cap each title at 120 characters. Deduplicate reads 1 and 2 on `id`. No tool returns an issue UUID, so `id` — which Linear re-keys when an issue changes team — is the only key available; that is a known ceiling of the native path, not something to work around.

Any Linear read failure — tool `error`, an unparseable `result`, or a field the render needs absent from every row — omits the board block and the completed metric and adds `Linear coverage unavailable` to `Coverage:`. Never estimate from branch names and never fabricate a row.

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

If the §2 gate returned `[SILENT]`, produce no user-facing report. Otherwise use this compact shape, board first:

**Board** (priority-sorted)
P1 GCN-42 Ship the thing — In Progress
P2 MC-4179 Review the other thing — Todo
P3 GCN-51 Scope the migration — Todo

- One line per issue, `P<n> <id> <title> — <status>`: `id` is the issue key (`GCN-42`), `status` the flat display name, and `<n>` is the priority **normalised to an int** — `priority` comes back as an object `{"value":3,"name":"Medium"}`, so read `p["value"] if isinstance(p, dict) else p` and sort on that; comparing the object directly sorts everything last. Sorted by the `linear-ticketing` priority rule: Urgent, High, Medium, Low, then No-priority last; within one priority, GCN first, keyed on `teamId == 81e7b769-2a46-4e2a-8db5-c165a7963b0e` rather than a parsed prefix; within one team, by the numeric suffix of `id` ascending, so `GCN-9` precedes `GCN-10`. A `priority` of 0 or absent renders `P–` and sorts last.
- Cap at 12 issue lines, then one final `+N more`.
- Empty board: emit the single line `Board: clear.` Read failure: omit the block and name it in `Coverage:`. Incomplete read: render what arrived and name it in `Coverage:` as partial. Never drop the block silently and never pass a partial board off as whole.
- The board is excluded from the ~350-word prose cap below—it must never squeeze out the judgment sections.

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

> Run `daily-work-brief` end to end. Use Europe/Paris. Establish the window from the last successful delivery, then apply the workday gate — weekday, official metropolitan-France holiday, explicit PayFit leave, then the bounded actual-work override for a nominal day off — and return exactly `[SILENT]` when the gate says not to run. Otherwise collect Slack, Cooper (Orca state, Claude transcripts, shell history, git), GitHub, Linear, email and calendar evidence in parallel, keep raw results out of the message, and deliver the compact brief. Pull Linear with `mcp__linear__list_issues` per the skill's three reads, projected with `fields`, reads 1 and 2 scoped by `state` type, and each proven complete by paging `cursor` until `hasNextPage == false` — never by narrowing the filter; the brief must OPEN with the `Board` block — every non-terminal GCN issue plus every non-terminal issue assigned to Gaetan on any team, deduplicated on `id`, priority-sorted with priority normalised from its object form to an int, capped at 12 lines then `+N more`, `Board: clear.` when the board is empty, and omitted with `Linear coverage unavailable` in `Coverage:` only when a read failed — never dropped silently. `Since the last daily` must end with the labelled `Claude/Orca:` line covering the Cooper Claude/Orca history, or `Claude/Orca: none.` Every source is read-only: write nothing to Linear, and do not spawn, prompt, or delegate to another agent. Return only the brief.

Delivery should return only the brief. On `[SILENT]`, remain silent. On a source or run failure that prevents a trustworthy daily, send one concise failure message naming the failed source and the next repair action—no cron wrapper prose and no fabricated substitute.
