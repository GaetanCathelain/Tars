# Slack tool visibility — correction of the withdrawn toolset gate

2026-08-10 (after the 14:37 UTC deploy), orchestrator-delegated Claude Code
session. Scratch clone only — **no ssh, no VM write**. All times UTC.

Supersedes the tool-gating half of
`status/probes/wf5/slack-channel-boundaries.md`; the deploy it corrects is
`status/probes/wf5/deploy-2026-08-10.md`; the plan to land it is
`docs/plans/apply-slack-channel-boundaries.md`. Not deployed — gated on
Gaetan's go.

> **Amended at apply time (2026-08-10 evening, restore dispatch):** the
> config half now ships chat-ID-map-only — `tool_progress_dm` removed and
> `D0BBYNM01BL: all` added to the map — because the adapter collapses MPIMs
> into chat_type `dm`, so the chat-type override would have narrated group
> DMs (adversarial-review blocker). The `tool_progress: off` this file calls
> "existing" was also re-measured as drifted to `all` and is SET at apply.
> Deploy status and record:
> `status/probes/wf6/slack-tool-visibility-restore.md`.

## Incident

After the boundary deploy, a **fresh @mention from Gaetan in `GQ07CQXT7`**
(a channel outside `{C0BP2GZUFSR, D0BBYNM01BL}`) was routed correctly by the
mention gate and opened a session with **zero tools**. Tars could not
delegate, could not read a file, could not do anything the request needed —
it was awake and disarmed. The gate did exactly what it was written to do;
what it was written to do was wrong.

## Root cause — a display request implemented as a capability restriction

Gaetan, in `C0BP2GZUFSR` thread `1786365348.026439`:

> You shouldn't have tool calling enabled in channels other than this one and
> the DM between you and me

Read as: *strip the toolset outside those two conversations.* Shipped as
`platform_tool_conversations` (VM lineage `94a3414` → `302ed2e` → deployed
`5af9e1e`), an early `return set()` in `_get_platform_tools`.

Meant: *don't narrate your tool calls in front of other people.* The subject
of the sentence is what the channel **shows**, not what the agent **can do** —
consistent with the `d615ca8` / `tool_progress_dm` work that immediately
preceded it (`status/lane-a.md`, commit `d1b68e2`), which split exactly the
same axis and is cosmetic. A capability reading also contradicts the reason
Tars is in those channels at all: it is asked to act there.

## Corrected policy

1. **Full, identical tool schema in every validly-invoked Slack
   conversation.** No conversation is toolless. (A stable schema also keeps
   the prompt cache warm.)
2. **Interim tool-call / progress display is per-conversation**: hidden in
   public and shared conversations, shown in the home channel `C0BP2GZUFSR`
   and the 1:1 DM `D0BBYNM01BL`.
3. **Final answers are delivered everywhere** — hiding narration never
   suppresses the answer.
4. **Mention routing is unchanged** (`strict_mention`,
   `free_response_channels: [C0BP2GZUFSR]`, no `allowed_channels`) —
   config-only, deployed 2026-08-10, correct as deployed.

## Implementation

Commit `0427536` on `d615ca8` (base `6e87d43`), exported as
`patches/0002-tool-progress-conversations.patch`. Subject: `feat(display):
per-conversation tool_progress override
(display.platforms.<platform>.tool_progress_conversations)`.

| File | Change |
|---|---|
| `gateway/run.py` | **+10 lines**, immediately after `d615ca8`'s `tool_progress_dm` block: read `display.platforms.<platform>.tool_progress_conversations`, a `chat_id → mode` map; if the map is a dict and holds this `source.chat_id`, `progress_mode = _normalise("tool_progress", value)`. Most specific wins — over `tool_progress_dm` and over the platform setting |
| `tests/gateway/test_slack_tool_visibility.py` | **new, 584 lines, 20 tests** — incl. 10 mention-routing tests ported from the withdrawn suite, and a guard test that toolsets never vary by conversation |

Cosmetic only, and deliberately the **opposite** of the withdrawn patch on
malformed input: a non-mapping value or an unlisted chat id is *ignored*, not
fail-closed. A typo in a display key must never change what the agent can do.

Config side (not applied): remove `platform_tool_conversations` entirely; add
`display.platforms.slack.tool_progress_conversations: {C0BP2GZUFSR: all}`
next to the existing `tool_progress: off` / `tool_progress_dm: all`.

## Verification evidence

Environment: scratch clone with push disabled, venv outside the repo,
Python 3.14.6, `pytest==9.1.1` + `pytest-asyncio==1.3.0`. Sanity run at
HEAD (`5af9e1e`) before writing anything: `test_run_progress_topics.py`
**18 passed**.

A behaviour probe at HEAD pinned the seams first: DM and workshop channel
resolve 18 toolsets, `C0BFQ5WFYTB` resolves **[]** (the withdrawn gate,
observed); with progress disabled the gateway passes **no**
`tool_progress_callback` at all; progress text is friendly-labelled
(`⚙️ Reading tars-probe.txt`), so display assertions match the preview.

**RED** — new test file on branch `mistaken` at `5af9e1e`:

```
$ venv/bin/python -m pytest tests/gateway/test_slack_tool_visibility.py -q --tb=line
5 failed, 15 passed in 7.97s        # re-run with the final file: 5 failed, 15 passed
```

| Test | Assertion that fired | Why |
|---|---|---|
| `test_public_channel_gets_the_same_toolset_as_the_workshop` | `public channel session was handed no tools` / `assert []` | gate empties `enabled_toolsets` |
| `test_public_channel_offers_the_full_tool_schema_to_the_model` | `assert {'read_file','terminal'} <= set()` | no schema reaches the wire |
| `test_public_channel_can_actually_run_a_tool` | `assert set() >= {'read_file'}` | nothing to dispatch |
| `test_workshop_channel_shows_interim_progress` | `the workshop channel narrated nothing` / `assert []` | no per-conversation display override exists yet |
| `test_toolsets_never_vary_by_conversation` | `{'C0BP2GZUFSR': [18 toolsets], 'C0BFQ5WFYTB': [], 'GQ07CQXT7': [], 'G0GROUPDM01': [], …}` | the guard, failing on the deployed code |

The 15 that passed at HEAD are the ones that should: public channel silent,
DM narrates, all three `final_answer` cases, and the 10 ported mention tests.
No test was massaged to fail.

**GREEN** — branch `corrected` from `d615ca8`, +10 lines in `gateway/run.py`:

```
$ venv/bin/python -m pytest tests/gateway/test_slack_tool_visibility.py -q
20 passed in 6.24s
```

Regression, 12 files (progress/display, Slack adapter, tools_config /
tool-definition):

```
135 passed, 1 warning in 22.41s
```

**Mutation (A) — revert only the production hunk, keep the tests:**

```
1 failed, 19 passed
FAILED …::test_workshop_channel_shows_interim_progress
E   AssertionError: the workshop channel narrated nothing
```

Exactly the one display test the hunk exists for; **no capability test
moved** — the change is display-only, proven, not asserted.

**Mutation (B) — re-apply the withdrawn gating diff on top of `corrected`
(`git diff 5af9e1e~1 5af9e1e -- gateway/run.py hermes_cli/tools_config.py`,
applied clean):**

```
4 failed, 16 passed
FAILED …::test_public_channel_gets_the_same_toolset_as_the_workshop
FAILED …::test_public_channel_offers_the_full_tool_schema_to_the_model
FAILED …::test_public_channel_can_actually_run_a_tool
FAILED …::test_toolsets_never_vary_by_conversation
```

The three capability tests and the guard fail; **every display test
survives** — the guard is a live anti-regression against re-introducing
capability gating, and the two concerns are orthogonal. Dropped → 20 passed.

**Pristine `git am`** — the patch applied to a bare `d615ca8` worktree:

```
$ git am --3way out/0002-tool-progress-conversations.patch
Applying: feat(display): per-conversation tool_progress override (…)
$ venv/bin/python -m pytest tests/gateway/test_slack_tool_visibility.py -q
20 passed in 8.27s
```

(The replayed sha differs — `git am` re-commits with a different committer;
tree and author are identical.)

## Caveats

- **Python version**: validated on 3.14.6 locally; the VM runs **3.11.15**.
  Nothing in the 10-line hunk is version-sensitive, but the suite has not run
  on 3.11 — pytest is not installed there and this session did no ssh.
- **Malformed *values* normalise to verbose.** An unrecognised value
  (`{C0…: {"nested": 1}}`) is not ignored: shared `_normalise` maps anything
  it does not recognise to `"all"`. Inherited verbatim from
  `tool_progress_dm` — same call, same semantics, cosmetic in both. A typo
  makes a conversation chattier, never less capable. Flagged, not fixed:
  making it stricter than its sibling buys nothing.
- **A stale `platform_tool_conversations` key is inert.** The corrected code
  never reads it — the test fixture deliberately keeps the key, pointing at
  two of five conversations, and the guard test asserts every conversation
  still resolves the same toolset. So the config-key removal and the code
  deploy need not be atomic, in either order.
- `_run_agent` returns the final answer rather than sending it; assertion (3)
  pins `result["final_response"]` for all three conversations, which is the
  value the delivery layer above returns. Driving
  `_handle_message_with_agent` would have added session plumbing without
  adding a fact the silent-channel + final-answer pair does not already give.
