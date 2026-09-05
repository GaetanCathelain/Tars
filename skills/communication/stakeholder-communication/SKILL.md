---
name: stakeholder-communication
description: "Use for stakeholder DMs, handoffs, and conversation recaps."
version: 1.0.0
metadata:
  hermes:
    tags: [communication, handoff, summary, slack, stakeholder]
    category: communication
---

# Stakeholder Communication

Draft concise stakeholder messages that preserve the right object: the conversation, the project, the decision, or the current status. Use for Slack DMs, handoffs, executive recaps, and messages written on someone else's behalf.

## First classify the requested overview

Before drafting, identify what the user actually asked to summarize:

| Requested object | Include | Do not substitute |
|---|---|---|
| **Conversation** | Questions asked, answers given, how understanding changed, corrections, decisions, and unresolved points, in sequence | Static project description or task checklist |
| **Project** | Goal, scope, architecture, deliverables, owners, risks | A transcript-like chronology |
| **Status** | What is done, running, blocked, and next | Full history unless it explains a blocker |
| **Decision** | Options considered, trade-offs, choice, rationale | General background without the decision path |

If the user says “our conversation,” default to the **interaction arc**, not the resulting project state.

## Conversation recap procedure

1. **Establish the opening question.** State what initiated the discussion.
2. **Preserve the sequence.** Summarize the material turns in chronological order.
3. **Name changes in understanding.** Include clarifications, rejected assumptions, and scope corrections; these are often the most useful part.
4. **Separate discussion from action.** Say what was merely proposed versus what was actually delegated, sent, changed, or verified.
5. **End at the current state.** Close with the latest decision, remaining uncertainty, or promised follow-up.
6. **Keep operational secrets out.** Avoid credentials, tokens, private transcript content, and unnecessary internal paths in stakeholder messages.

## Cross-conversation audits and ranked recommendations

When asked to analyse interactions across many conversations, use this procedure rather than expanding a single-thread recap:

1. **Establish coverage before interpreting patterns.** Record the requested time window, earliest retrievable evidence, cutoff and source boundaries. Combine live platform history with retained local sessions where useful; distinguish an available-history audit from a complete platform export.
2. **Separate actual turns from injected context.** Treat quoted roots, rehydrated messages, summaries and worker notifications as context, not new user requests. Preserve platform event IDs where present; label content-based deduplication as heuristic because intentionally repeated messages can be identical.
3. **Build a coverage manifest before dividing the reading.** Partition large corpora by non-overlapping message-time windows, retain source/session/message identifiers, and reconcile the union of reviewed sessions. Do not add per-window session counts: one conversation can span several windows. Report stored rows separately from unique messages or human turns.
4. **Redact before emitting or delegating excerpts.** Inspect sensitive source material locally and pass only the minimum safe evidence; redacting the final report does not undo exposure in intermediate tool output or worker context.
5. **Read both successful and difficult interaction arcs.** Preserve clarifications, rejected assumptions, recovery and positive feedback. Detailed requested reports can work well; do not turn complaints about noisy status messages into a universal brevity rule.
6. **Verify representative causes against primary records.** Check actual tool calls, results and platform messages before attributing a failure to the assistant, worker or transport. An apology proves an acknowledgment, not its explanation; a successful retry proves recovery, not root-cause resolution. Distinguish historical incidents from defects verified to remain active.
7. **Rank recommendations by harm, recurring supervision cost and practical leverage.** Give each recommendation its evidence, proposed change, responsible layer and observable acceptance criterion. Prefer repairing existing mechanisms over introducing another orchestration layer. Separate execution enforcement, durable task state, evidence-backed completion and communication policy instead of prescribing more instructions for every failure.
8. **Deliver the synthesis with explicit limits.** Lead with the main finding, then a short chronology, strengths worth preserving and ranked recommendations. Keep detailed source IDs and methodology in an evidence appendix. State which recommendations are proposed versus implemented; an analytical request does not authorize the fixes.

## Draft-before-send rule

When asked to “first draft” a message:

- Produce only the draft.
- Explicitly state that it has not been sent.
- Do not send until the user approves the wording or explicitly instructs sending.
- Resolve recipient identity before sending, but do not clutter the draft with lookup details.
- Treat known personal details as acceptance criteria: match the recipient or subject’s gender, role, and requested form of address consistently throughout the draft.
- Do not echo delegation briefs, agent prompts, exact commands, or internal execution logs into chat unless the user asks. Keep operational detail in the work system (for example Linear) at most; report the verdict, status, unresolved decisions, and useful handles.
- Distinguish drafted text from a native platform draft object. If the user asks to save a Slack Draft, use a verified draft-capable integration or authenticated UI; otherwise provide the text and state precisely that it was not saved or sent, without turning a temporary capability gap into a permanent claim.

## Apply editorial revisions literally

When the user revises a stakeholder draft, treat each instruction as an acceptance criterion rather than a suggestion:

1. Build a short add/drop checklist from the latest instruction.
2. Delete requested material completely; do not preserve it through a paraphrase elsewhere.
3. Add requested source links using the canonical upstream URL, placed where the named project first appears.
4. Preserve requested meta-context about **how the interaction happened** when it is part of the story—for example, that the user stayed in Slack while the assistant autonomously inspected a machine, delegated work, tracked it, and verified results.
5. Reread the entire revised draft against the checklist before presenting or sending it. Later revisions supersede earlier wording constraints where they conflict.

Operational autonomy belongs in a conversation recap when it demonstrates the interaction model; it should be one concise sentence, not a self-congratulatory implementation log.

## Quality check

Before presenting a recap, ask:

- Does this describe the **interaction**, or did it drift into describing the artifact?
- Can the recipient see how the conclusion evolved?
- Are the user's corrections represented fairly rather than silently absorbed?
- Did I apply every requested addition and deletion literally?
- Are requested project/source links present and canonical?
- Does it distinguish verified facts from plans and work still in progress?
- Is it short enough for the destination while retaining the decision path?

## Message shape

A useful stakeholder conversation recap usually follows:

1. Why the conversation started.
2. What was initially proposed or understood.
3. What the user clarified or changed.
4. What action followed.
5. Where the matter stands now.

For the specific failure mode and corrected example, read `references/conversation-vs-project-overview.md`.
