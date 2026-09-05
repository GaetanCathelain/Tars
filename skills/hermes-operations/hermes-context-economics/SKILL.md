---
name: hermes-context-economics
description: "Use when auditing Hermes context size and compaction cost."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [hermes, context, tokens, compaction, toolsets, prompt-cache]
---

# Hermes Context Economics

Measure and reduce Hermes startup context, accumulated history, tool-schema overhead, and compaction stalls without trading away required capabilities or silently breaking prompt caching.

Read `references/context-audit-procedure.md` for the deterministic measurement procedure, interpretation rules, and optimization order.

## Core workflow

1. **Measure the real request.** Record the provider-reported input tokens on the first API call. Use the session's stored system prompt and the active serialized tool definitions to apportion that total; label tokenizer-derived component counts as estimates.
2. **Separate fixed from accumulated cost.** Report startup components independently from conversation messages and tool results. A large immutable prefix may be expensive without being the event that triggered compaction.
3. **Rank schemas against use.** Pair every direct tool's schema-token size with its recorded call count over an explicit date range. Flag zero-use and high-cost/low-use tools. Call them tools, not skills.
4. **Read compression telemetry.** Capture effective context limit and trigger, request size at compression, duration, before/after message or token estimates, and the largest tool-result producers in the affected session.
5. **Optimize without reducing reach.** Start with bounded extraction and compact skill essentials/on-demand references. Measure result growth and cache behavior before proposing pruning or threshold changes. Gaetan declined profile splitting and MCP pruning; do not reintroduce either, remove tools, or alter provider/model defaults without new explicit authority and measured justification.
6. **Verify on a fresh session.** Tool schemas are session-stable for prompt-cache safety. Re-read config after writes, run the config validator, and measure a new session's first request before claiming savings.

## Decision rules

- Direct schemas must remain stable within a session; changing them mid-thread breaks the cached prefix and does not retroactively shrink earlier requests.
- Deferred catalogs are cheaper than loading every deferred schema, but their names and descriptions still have a measurable fixed cost.
- Keep worker lifecycle tools available to actual workers without enabling them in every chat. For Kanban, dispatcher-spawned workers can receive their lifecycle surface through `HERMES_KANBAN_TASK`; ordinary profiles do not need the explicit full toolset.
- Prefer deterministic proactive pruning with a meaningful minimum-reclaim gate when old, large tool results drive growth.
- Do not enable micro-compaction merely to remove visible stalls: it rewrites history repeatedly and can invalidate prompt caching every turn. Treat it as a measured latency-versus-cache trade, never a free optimization.
- Raising the batch threshold is low value when telemetry already shows a high effective threshold. Fix what fills the window first.
- Never remove a capability solely because a short observation window shows zero calls. Authorized channels must remain tool-capable; display quieting and context savings are not permission changes.
- For transcript/log/database audits, sanitize inside the reading process **before** excerpts reach stdout, tool context or scratch artifacts. Project IDs/counts first; mask credential/token/header/cookie values and bearer/magic-link URLs before clipping, since clipping can hide their markers. When uncertain, omit the content and retain only the evidence handle. Never print raw historical excerpts to inspect whether they contain secrets; runtime redaction is defense in depth, not the extraction plan.
- Measure time to useful verdict, avoidable user corrections, duplicate dispatches, premature completion claims and useful versus redundant notifications separately from startup tokens and compaction. Compare representative before/after tasks; byte reductions alone prove neither token savings nor behavior improvement.

## Reporting contract

Give the verdict first, then:

- exact observation window and session count;
- fixed token table;
- schema-token versus call-count table;
- compaction count, thresholds, durations, and dominant outputs;
- ranked optimizations with estimated savings and trade-offs;
- changes applied versus recommendations only.

Do not claim a configuration change is active from a successful write alone. Re-read it, validate configuration, and distinguish current-session, fresh-session, and restarted-daemon effects.
