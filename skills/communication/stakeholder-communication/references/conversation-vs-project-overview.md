# Conversation overview vs project overview

## Failure pattern

A user asked for a quick overview of “our conversations” to send to a stakeholder. The first draft instead listed the implementation scope, architecture, security choices, and acceptance checks. Those facts were accurate, but the object was wrong: it was a project brief rather than a recap of the exchange.

## Correct interpretation

A conversation overview should preserve the interaction arc:

1. The user opened by asking how a tool would be installed.
2. The assistant investigated and proposed a deployment approach.
3. The user expanded the request to include credentials, history, and integrations.
4. The assistant explained feasibility and limitations.
5. The user narrowed one UI scope and asked for delegation.
6. A later question exposed that two different UIs had been conflated.
7. The assistant acknowledged the misunderstanding and amended the active brief.
8. The recap ends with the work running and verification pending.

The architecture may appear only as context for a conversational turn. It should not become the organizing structure.

## Contrast

**Project-oriented:**

> The POC contains Docker Compose, a CLI container, copied authentication state, MCP configuration, and a local web UI.

**Conversation-oriented:**

> The discussion began with an installation question. After reviewing the proposed container setup, the user asked whether existing authentication, history, and MCPs could be carried over. The user then removed one dashboard, later clarified that a separate Goal Planner UI was still required, and the active implementation brief was corrected accordingly.

## Review heuristic

Remove the names of the participants and reread the draft:

- If it still reads like README content, it is probably a project overview.
- If it shows what each turn changed about the shared understanding, it is a conversation overview.
