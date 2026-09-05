---
name: hermes-context-diagnostics
description: "Use when measuring Hermes prompt or context overhead."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [hermes, context, tokens, prompt, tools, diagnostics]
---

# Hermes Context Diagnostics

Measure where a Hermes session's input context goes. Use this when startup feels unexpectedly large, compaction seems ineffective, or the user wants token counts for tools, operating instructions, skills, memory, profile, and conversation history.

Read `references/startup-context-audit.md` for the measured database/log procedure and reporting contract.

## Core workflow

1. Identify the active session and its first model call.
2. Run `hermes prompt-size --platform <platform> --json` to reconstruct the current platform's system prompt and model-tool assembly; use its character and JSON-byte breakdown for orientation, never as token counts.
3. Take the provider-reported request input-token count as the exact total.
4. Resolve the immutable stored system prompt through the session's prompt hash; prefer it over the current `prompt-size` reconstruction when auditing a past session because configuration may have changed.
5. Partition the prompt by stable section markers into operating record, runtime policy, workflow protocols, host/session metadata, skill index, memory, and user profile.
6. Tokenize prompt partitions and compactly serialized active tools with the closest model tokenizer.
7. Report component values as estimates and leave wrappers/current inbound context/tokenizer mismatch as an explicit residual.
8. Check the live history count before blaming conversation history or compaction.

## Interpretation

- Compaction reduces accumulated conversation history. It does not remove fixed startup scaffolding such as tool schemas, operating instructions, the active skill index, memory, or profile.
- Tool count is not tool cost. Long descriptions, JSON schemas, and catalogs often dominate.
- File size is not token count. Resolve the exact stored prompt used by the session.
- A provider total may be exact while the component split remains estimated.

## Output format

Lead with the exact first-call input total, then a compact table of component tokens and percentages. Label exact, estimated, and residual figures clearly. Include the observed history count when it answers whether old conversation was responsible.

## Verification

- Provider total came from the matching session and first API call.
- Prompt came from the hash referenced by that session.
- Tool definitions match the active profile/platform assembly.
- Component estimates plus residual equal the provider total.
- No per-component estimate is presented as provider-exact accounting.
