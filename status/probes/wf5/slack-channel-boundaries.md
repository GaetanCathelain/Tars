# Slack channel boundaries — recon + implementation evidence

2026-08-10, orchestrator-delegated Claude Code session (scratch dir only, no
VM write). All times UTC.

> **RETRACTION, 2026-08-10 (later the same day): this probe's tool-gating
> half is withdrawn.** The mention-routing half — root cause, wake-path
> trace, MPIM verdict, restart verdict, `allowed_channels` removal — stands
> as written and is deployed. Everything this file said about *toolset*
> gating (commit `302ed2e`, the `platform_tool_conversations` key, the
> 39-test suite and its mutation/collateral numbers) implemented a misread
> instruction: Gaetan asked for tool-call **display** to be limited to the
> home channel and his DM, not tool **capability**. Those sections have been
> cut back to a retraction below. Correction evidence, corrected patch and
> the incident it caused:
> `status/probes/wf5/slack-tool-visibility-correction.md`.

## Observed regression

Channel `C0BFQ5WFYTB`, thread `1786362584.259959`. Gaetan's unmentioned
in-thread reply "Mrc le goat" woke the bot, which ran `kanban_show` and
replied — in a channel that should have neither auto-follow wake nor tools.
Full unredacted (non-secret) log lines (`recon-vm.md` §4):

```
2026-08-10 12:34:07,768 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ
chat=C0BFQ5WFYTB msg='Mrc le goat' reply_to_id=1786362584.259959
reply_to_text='apparemment petit problème de permissions sur la HT console [image: image.png]'
```

Follow-on trace, session `20260810_123407_b4b91a87` (`agent.log`):

```
12:34:07,857  agent.turn_context: conversation turn: session=20260810_123407_b4b91a87 ...
12:34:07,840  run_agent: OpenAI client created (agent_init, shared=True)
12:34:13,755  agent.conversation_loop: API call #1: model=gpt-5.6-sol ...
12:34:13,760  WARNING agent.tool_executor: Tool kanban_show returned error (0.00s) ...
12:34:18,118  agent.conversation_loop: API call #2: model=gpt-5.6-sol ...
12:34:18,128  agent.conversation_loop: Turn ended: reason=text_response(...)
12:34:18,369  gateway.run: response ready: platform=slack chat=C0BFQ5WFYTB time=10.6s api_calls=2
12:34:18,394  gateway.platforms.base: [Slack] Sending response (14 chars) to C0BFQ5WFYTB
```

`kanban_show` errored (no board in that context) but ran at all — on a turn
that should never have started. The bug is the wake: an unmentioned message
in a shared channel must not open a turn. (The original version of this file
read the tool call itself as the defect and concluded the toolset should have
been empty — retracted; see the banner above.)

## Root-cause trace

- **Wake path — corrected:** the original phrasing here ("`slack.strict_mention`
  unset on the VM") was imprecise. Adversarial verification
  (verify-mention.md §3) found `strict_mention: true` **is** written on the
  VM, under `gateway.platforms.slack`, and is silently shadowed there by the
  presence of a top-level `slack:` block: `gateway/config.py:1511-1530` (the
  shared-key bridge loop) and `:1652-1664` (the `apply_yaml_config_fn`
  dispatch that stamps `SLACK_STRICT_MENTION`) both prefer
  `yaml_cfg.get("slack")` and only fall back to `gateway.platforms.slack` /
  `platforms.slack` when that key is absent or not a dict;
  `PlatformConfig.from_dict` (`:661-707`) keeps only the `extra:` sub-key, so
  nested `strict_mention`/`require_mention` written as plain keys never
  reach `config.extra` by any other route. Measured: VM-shaped YAML (keys
  under `gateway.platforms.slack`, no top-level `slack:` block) → effective
  `strict=False`; the identical keys moved to a top-level `slack:` block →
  effective `strict=True`. Configured-but-shadowed, not unset. Given that,
  `_handle_slack_message` fell past the strict gate at `adapter.py:5707`
  into `elif not is_mentioned:` at `adapter.py:5721` →
  `_should_wake_on_unmentioned_message` (`adapter.py:5164`) → check 3,
  `_has_active_session_for_thread` (`adapter.py:5203` → `adapter.py:7972`),
  found the live session for this thread and woke on the unmentioned
  message.
- **Toolset gap — retracted.** This file originally listed a second root
  cause: that `_get_platform_tools()` resolves per *platform* only, so every
  Slack conversation gets the identical `hermes-slack` toolset. That is a
  true description of the code and **not a defect** — it is the intended
  behaviour (capability is uniform; display is what varies). It is now pinned
  as an invariant by `test_toolsets_never_vary_by_conversation`.

## What was built — tool-gating half RETRACTED

The mention-routing fix is **config-only**: no code change at all (see
§Restart verdict). It is deployed and unaffected by this retraction.

The code half this section originally described — commit `302ed2e`
(`platform_tool_conversations`, amending `94a3414`, deployed to the VM as
`5af9e1e`), its `_tools_allowed_in_conversation` fail-closed gate, the two
gateway call sites that threaded `chat_id`, and
`tests/gateway/test_slack_channel_boundaries.py` (39 tests) — **is
withdrawn**. It restricted capability; Gaetan had asked for display. The
corrected, display-only replacement and the full red/green/mutation evidence
are in `status/probes/wf5/slack-tool-visibility-correction.md`; the deploy
that carried the withdrawn code is `status/probes/wf5/deploy-2026-08-10.md`,
and its removal is planned in `docs/plans/apply-slack-channel-boundaries.md`.

Retracted with it: the 39-test matrix, the mutation and collateral tables and
the empty-toolset assertions this file's §Test evidence originally reported.
None of those numbers describe anything that should ship.

## Test evidence — mention routing (the surviving half)

Python 3.14.6 in a scratch venv (the VM runs 3.11.15 — not re-run there).
These cases drive the real `SlackAdapter._handle_slack_message`; **no
`allowed_channels` anywhere in the fixture** — the mention gate is the
channel boundary. They need no code change, only the corrected config
placement.

| Case | Expected | Result |
|---|---|---|
| `C0BP2GZUFSR` top-level + thread, unmentioned | PROCESSED | PASS |
| `D0BBYNM01BL` 1:1 IM, unmentioned | PROCESSED | PASS |
| `C0BFQ5WFYTB` thread reply, unmentioned, all wake paths armed | IGNORED | PASS |
| control: same armed thread, `strict_mention: false` | PROCESSED | PASS (proves the armed test isn't vacuous) |
| `C0BFQ5WFYTB` fresh @mention | PROCESSED | PASS |
| MPIM `G0GROUPDM01`, unmentioned thread reply, wake paths armed | IGNORED | PASS |
| MPIM `G0GROUPDM01`, fresh @mention | PROCESSED | PASS |
| never-configured channel `C_RANDOM_9999`, fresh @mention | PROCESSED | PASS |
| `C_RANDOM_9999`, unmentioned thread reply, wake paths armed | IGNORED | PASS |

Ten mention-routing tests were ported from the withdrawn suite into the
corrected patch's `tests/gateway/test_slack_tool_visibility.py` and pass
there, so this coverage survives the retraction rather than being lost with
the file it lived in.

## MPIM verdict — corrected 2026-08-10 (self-review round 2)

**An MPIM is mention-gated exactly like a channel: unmentioned messages are
dropped by `strict_mention`, a fresh @-mention is processed.**
(The original wording added "and its toolset is empty" — retracted with the
tool-gating half; a mentioned MPIM turn gets the same toolset as anywhere
else, only its progress display is quiet.) `adapter.py:5512` treats an MPIM as a DM only for
*DM-disable* purposes (`is_dm = channel_type in {"im", "mpim"}`);
`adapter.py:5528` `is_one_to_one_dm = channel_type == "im"` explicitly
excludes MPIMs from mention-exemption — the comment at `:5520-5527` states
"An MPIM (group DM) is a SHARED surface … it must obey the same operator
controls as a channel". `adapter.py:5654` `if not is_one_to_one_dm and
bot_uid:` puts an MPIM through the same gate ladder as a channel. The first
round kept the `allowed_channels` whitelist, which dropped MPIMs *even when
mentioned* (`:5656-5661`) — that contradicted the invariant "an explicit
mention must be processed on every shared surface" and is why
`allowed_channels` is now removed: with it absent,
`_slack_allowed_channels()` returns the empty set and the whitelist gate
short-circuits (`:5657` — "empty set means no channel restriction"), so the
`strict_mention` elif at `:5707` does the gating. No code change was needed
for any of this. Pinned by `test_group_dm_unmentioned_thread_reply_is_ignored` and
`test_group_dm_fresh_mention_is_processed` (both ported into the corrected
suite); `test_group_dm_gets_no_toolsets` is retracted with the tool gate.

## Restart verdict

**Editing the `slack:` block does NOT take effect without a gateway
restart.** `_slack_strict_mention()` (`adapter.py:8263`) reads
`self.config.extra.get("strict_mention")` then falls back to
`os.getenv("SLACK_STRICT_MENTION", "false")`; `strict_mention` is never
bridged into `PlatformConfig.extra` (absent from `gateway/config.py:1533-1642`)
— it only reaches the process as an env var written once by
`_apply_yaml_config` (`adapter.py:8995`), itself dispatched exactly once
from `load_gateway_config()` at gateway startup (`gateway/config.py:1668`).
The per-turn re-read (`gateway/run.py:3169`) returns a raw dict via
`read_raw_config()` and never rebuilds `PlatformConfig` or re-runs the hook;
`SlackAdapter.__init__` runs once. `free_response_channels` *is* bridged
into `extra` (`gateway/config.py:1561-1562`) but that `extra` is built by the
same startup call and frozen on the adapter instance — same verdict.
`allowed_channels` follows the same env-stamp path as `strict_mention`
(`adapter.py:9031-9035`, set-if-absent, never cleared), so *removing* it
from the YAML also only takes effect at a restart. (A paragraph here about
the tool-gate key's live-reload behaviour is retracted with that key. The
corrected display key `tool_progress_conversations` is likewise read from the
raw per-turn config, but its *code* ships in a patch, so its deploy needs a
restart regardless — `docs/plans/apply-slack-channel-boundaries.md` §(c).)

## Side-find for Gaetan

The repo's `CLAUDE.md` says Hermes "live-reloads (~30 s)". That wording is
imprecise on two counts: (1) there is no polling watcher — `config.yaml` is
re-read fresh on every turn (`read_raw_config`, mtime/size cache-busted,
`hermes_cli/config.py:2933`), so a plain config-level change is closer to
"next message" than a fixed 30 s cadence; (2) for Slack specifically, the
`PlatformConfig`/env snapshot (`strict_mention`, `free_response_channels`,
and anything else bridged into `extra`) is startup-only regardless of how
fast the raw config re-read is — no amount of waiting substitutes for a
restart. Flagged here; `CLAUDE.md` not edited (out of this session's scope).

## Verification (adversarial, 3 lenses)

Three independent adversarial passes (bypass hunt, mention-gate refutation,
test-integrity + rerun) ran against commit `94a3414` before it touched the
VM; overall PASS at the time. Their tool-gate findings — proxy-mode bypass,
the toolless-channel system prompt, cron scoping, the fail-open malformed
config, the AST call-site pin — and this round's test numbers are
**retracted with the gate**: they only ever described the withdrawn code.
The mention-routing findings stand:

| Severity | Finding | Source |
|---|---|---|
| minor | `unauthorized_dm_behavior: ignore` is inert via the same shadowing as the root cause — restored in the apply plan's config block; harmless today since the allowlist early-reject fires first | verify-mention.md §7, F4 |
| minor, standing interaction | `slack.reaction_triggers` / `slack.mention_patterns` must stay unset — a `reaction_triggers` list is a `force_process` path that wakes the bot on any unmentioned message; `docs/proposals/P2-emoji-trigger.md` as written would re-open the gate in `C0BFQ5WFYTB` and needs reconciling first | verify-mention.md §5, F2 |

Full evidence: `verify-bypass.md`, `verify-mention.md`, `verify-tests.md` —
session scratchpad artifacts, not checked into this repo. This probe file is
their summary of record.

## Self-review round 2 — corrections to `8466fbd` (2026-08-10)

An orchestrator self-review of the committed `8466fbd` raised two blockers
and one test-strength item. **Only the first survives.** The other two were
corrections *to* the withdrawn tool gate — fail-closed semantics for a
malformed `platform_tool_conversations`, and strengthening the AST pin on the
`chat_id` call sites — and are retracted with it, along with that round's
re-run, collateral and mutation numbers.

1. **`allowed_channels` contradicted the invariant.** The first round kept
   `allowed_channels: C0BP2GZUFSR,C0BFQ5WFYTB`, so channels outside that
   list and every MPIM were dropped *even when explicitly mentioned*. The
   required invariant is mention-gated, not membership-gated: outside
   `C0BP2GZUFSR`, every channel and group DM is fresh-@mention-only and an
   explicit mention must be processed. Fix: `allowed_channels` removed from
   the boundary config and the test fixture; new handler-level tests pin
   mention-processed / unmentioned-ignored for an MPIM and for a
   never-configured channel (see matrix). `SLACK_ALLOWED_USERS` unchanged.
   Deployed and live.

Neither self-review round caught the real defect — that the entire
tool-gating half answered a question Gaetan had not asked. That surfaced
later on 2026-08-10, when a fresh mention in `GQ07CQXT7` opened a session
with no tools: `status/probes/wf5/slack-tool-visibility-correction.md`.
