# Context audit procedure

Use this reference to answer “what consumes startup context?”, “why did this thread compact?”, and “which tools should leave the default profile?”

## 1. Fix the scope

Record:

- Hermes version, active profile, provider, and model;
- observation window for usage counts;
- number of sessions represented;
- one target session for compaction diagnosis;
- whether counts include cron and worker sessions.

Do not mix a ten-day usage sample with lifetime language.

## 2. Measure the fixed prefix

Use the provider/API log's first-call `input_tokens` as the exact aggregate. Resolve the session's `system_prompt_hash` through the session database's `system_prompts` table, then partition the stored prompt by stable headings such as:

- operating/SOUL rules;
- generic Hermes policies;
- Kanban or other lifecycle protocol blocks;
- execution policy;
- host/runtime metadata;
- skill instructions and active-skill index;
- memory and user profile.

Serialize the active direct tool definitions in the same shape the provider receives. Tokenize prompt components and individual schemas with the closest available model encoding. State that component counts are estimates; reconcile their sum to the exact provider total with a residual bucket for message framing, provider instructions, and the current user turn.

## 3. Rank tool cost against use

For each startup-loaded tool, report:

- schema tokens;
- calls in the explicit observation window;
- zero-use or low-use status;
- the toolset that supplies it;
- whether a worker/runtime gate can add it only when needed.

Count completed tool-result rows by `tool_name`; one row corresponds to one invocation. Keep deferred tools separate: only `tool_search`, `tool_describe`, and `tool_call` are direct schemas, while the catalog text inside `tool_search` is fixed startup cost. Full deferred schemas are paid only after description/loading.

Useful interpretations:

- high tokens + zero calls: removal candidate, subject to capability review;
- high tokens + frequent calls: schema-compression candidate, not removal candidate;
- low tokens + zero calls: negligible savings;
- worker lifecycle tools in normal chats: profile/toolset routing problem.

## 4. Diagnose actual compaction

From compression logs capture:

- request-token estimate at trigger;
- effective threshold and model context limit;
- compression duration;
- message/token estimate before and after;
- retry or ineffective-compression signals.

Then aggregate target-session tool-result characters by `tool_name`, with count and maximum result size. This distinguishes a fixed-prefix problem from a few enormous file, Slack, issue-list, or skill reads.

Never infer that the configured global ratio was active: model/provider-specific overrides may raise it. Report the effective threshold shown in telemetry.

## 5. Optimization order

1. Bound and project tool extraction; sanitize transcripts inside the reader before output or persistence, and mask secrets before clipping. Prefer IDs/counts over excerpts.
2. Keep skill triggers, invariants and main workflow compact; move long specialist detail to on-demand references when a concrete read-cost measurement warrants it.
3. Measure episodic pruning's reclaim and cache cost before proposing it; do not silently enable history-rewriting controls.
4. Preserve the user's current tool reach, profiles and provider/model defaults. Profile splitting and MCP pruning were declined; a zero-use sample does not reopen that decision.
5. Distinguish fresh-session measurements from the active thread. Do not reset a session merely to claim a smaller prompt.
6. Tune thresholds last and only with authorization. A trigger already near the model limit is not the cause of rapid growth.

## 6. Cache trade-offs

- Batch compaction: occasional large stall and occasional prefix rewrite.
- Proactive prune: episodic deterministic rewrite, bounded by reclaim gates.
- Micro-compaction: smaller recurring work but repeated history rewrites; can invalidate the cached prefix every cadence interval.

Prefer measurement over slogans. Compare compaction frequency and duration, cache-read ratios, and input-token growth before and after a change.

## 7. Verification

After a toolset/config change:

1. re-read every changed key;
2. run the supported config validator;
3. determine whether the setting is per-session or daemon-startup scoped;
4. start a fresh session when schemas are session-stable;
5. record its first-call token total and active tool names;
6. verify lifecycle workers still receive required tools;
7. compare measured rather than projected savings.

A successful config write is not proof of activation, and activation in a fresh process is not proof that the current thread shrank.
