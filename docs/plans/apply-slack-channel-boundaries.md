# Apply plan — Slack channel boundaries

**Source:** `patches/`, `status/probes/wf5/slack-channel-boundaries.md`
(mention-routing evidence),
`status/probes/wf5/slack-tool-visibility-correction.md` (display evidence).

**Status, 2026-08-10:**
- The **mention-routing half is deployed** (14:37 UTC, record:
  `status/probes/wf5/deploy-2026-08-10.md`). It is correct and stays.
- The same deploy also shipped **per-conversation toolset gating**, which is
  **withdrawn** — it restricted capability where Gaetan asked for a display
  change. Removing it and applying the corrected display patch is the
  "what to apply" below. **PERFORMED 2026-08-10 19:57–20:01 UTC** (Orca
  dispatch on Gaetan's go — he asked Tars in Slack whether the fix was
  pushed; the dispatch instruction covered deploy). VM checkout `4d10183`
  on `d615ca8`, config edited under flock with .bak, gateway restarted
  (PID 774992). Deploy record + all verification:
  `status/probes/wf6/slack-tool-visibility-restore.md`. The plan
  originally sat on a local lineage that never reached `main`; that
  dispatch restored and re-based it onto `main`.

> The `verify-mention.md` / `verify-bypass.md` / `impl-report` documents this
> plan cites are session scratchpad artifacts, not in the repo;
> `status/probes/wf5/slack-channel-boundaries.md` is their summary of record.

## Goal

Four invariants, verbatim-precise:

1. **Capability is identical in every validly-invoked Slack conversation.**
   Whatever conversation Tars is legitimately woken in, the model is offered
   the full, identical tool schema — no conversation is toolless. A stable
   schema is also what keeps the prompt cache warm.
2. **Interim tool-call / progress display is per-conversation.** The bubbles
   ("⚙️ Reading …") are shown in the home channel `C0BP2GZUFSR` and in
   Gaetan's 1:1 IM `D0BBYNM01BL`, hidden in every other public or shared
   conversation.
3. **Final answers are delivered everywhere.** Hiding the narration never
   suppresses the answer.
4. **Outside `C0BP2GZUFSR`, fresh-@mention-only — on every shared surface,
   channels and group DMs (MPIMs) alike.** No auto-follow wake on an
   unmentioned message, in a thread or otherwise; a fresh explicit @mention
   *on that message* is always processed. There is **no channel whitelist**:
   `allowed_channels` is removed (corrected 2026-08-10 after self-review — a
   whitelist contradicted "an explicit mention must be processed" for
   unlisted channels and MPIMs). `SLACK_ALLOWED_USERS` remains the who-gate.
   `C0BP2GZUFSR` itself stays free-response (unmentioned top-level and thread
   messages both processed). The 1:1 DM (`channel_type: im`) is exempt from
   mention gating entirely, as it already was. **Config-only, already
   deployed, unchanged by this plan.**

## Root cause

### The wake regression (fixed, deployed)

`strict_mention` was not simply unset — the 2026-08-07 live-config snapshot
(`status/probes/wf5/live-config-snapshot.md:153-158`) shows it configured
under `gateway.platforms.slack: {require_mention: true, strict_mention: true,
unauthorized_dm_behavior: ignore}`. It has no effect there: a top-level
`slack:` block silently shadows it. `gateway/config.py:1511-1530` (the
shared-key bridge loop) and `:1652-1664` (the `apply_yaml_config_fn` dispatch
that stamps `SLACK_STRICT_MENTION` etc.) both prefer `yaml_cfg.get("slack")`
and only fall back to `gateway.platforms.slack` / `platforms.slack` when that
key is absent or not a dict; `PlatformConfig.from_dict` (`:661-707`) keeps
only the `extra:` sub-key, so nested `strict_mention`/`require_mention`
written as plain keys never reach `config.extra` by any other route.
Measured (verify-mention.md §3): VM-shaped YAML (keys under
`gateway.platforms.slack`, no top-level `slack:` block) → effective
`strict=False`, incident reproduces; the identical keys moved to a top-level
`slack:` block → effective `strict=True`, regression is IGNORED.

`_handle_slack_message` therefore fell past the strict gate at
`adapter.py:5707` into `elif not is_mentioned:` at `adapter.py:5721`, which
woke the bot on an unmentioned message via `_has_active_session_for_thread`
(`adapter.py:5203` → `adapter.py:7972`).

> **Warning — do not "tidy" the new keys back under `gateway.platforms.slack`.**
> The keys go in the top-level `slack:` block below. `docs/specs/tars-profile.md`
> documents `gateway.platforms.slack` as their home (corrected there
> post-WF5) — that placement is the shadowed one and silently reverts this
> entire fix. At apply time, **remove** any
> `strict_mention`/`require_mention`/`unauthorized_dm_behavior` keys still
> sitting under `gateway.platforms.slack`: two contradictory copies is what
> let this regression hide in the first place.

### The misread (what this plan undoes)

Gaetan's instruction — "You shouldn't have tool calling enabled in channels
other than this one and the DM between you and me" (`C0BP2GZUFSR`, thread
`1786365348.026439`) — was implemented as a *capability* gate:
`platform_tool_conversations` stripped every toolset outside
`{C0BP2GZUFSR, D0BBYNM01BL}`. He meant the cosmetic tool-call/progress
**display**. The gate's cost showed up immediately: a fresh mention in
`GQ07CQXT7` opened a toolless session that could not delegate anything.
Full incident + corrected implementation:
`status/probes/wf5/slack-tool-visibility-correction.md`.

## Preconditions

1. **Proxy mode expected off** (verified off 2026-08-10). This was a hard
   precondition for the withdrawn gate — proxy mode routes agent runs to a
   remote server that resolves tools with no `chat_id`
   (`gateway/run.py:24694` → `gateway/platforms/api_server.py:2868`), which
   bypassed it (verify-bypass.md F1). With the gate withdrawn there is
   nothing to bypass; kept as a cheap sanity check, no longer a blocker.
   ```bash
   ~/.local/bin/hermes config get gateway.proxy_url        # expect empty
   grep -c '^GATEWAY_PROXY_URL=' ~/.hermes/.env             # expect 0 — grep -c only, never cat .env
   ```
2. **These env vars must stay ABSENT from `~/.hermes/.env`** — each would
   silently override the YAML forever (verify-mention.md §6, F3):
   `SLACK_STRICT_MENTION`, `SLACK_REQUIRE_MENTION`,
   `SLACK_FREE_RESPONSE_CHANNELS`, `SLACK_ALLOWED_CHANNELS`,
   `SLACK_REQUIRE_MENTION_CHANNELS`, `SLACK_THREAD_REQUIRE_MENTION`,
   `SLACK_MENTION_PATTERNS`, `SLACK_REACTION_TRIGGERS`.
3. **`slack.reaction_triggers` and `slack.mention_patterns` must stay
   unset.** A `reaction_triggers` list is a `force_process` path that runs a
   full turn on *any* unmentioned message in an allowed channel
   (`adapter.py:5680-5681`, `:4858-4859` — verify-mention.md §5, F2).
   `docs/proposals/P2-emoji-trigger.md` as written
   (`reaction_triggers: [kanban]`) would re-open the mention gate in
   `C0BFQ5WFYTB` — **P2 must be reconciled with this boundary before it
   lands.**
4. **`SLACK_ALLOWED_USERS` must be PRESENT and non-empty in
   `~/.hermes/.env`** — after this change it is the sole who-gate in front
   of a fully-armed Tars in every conversation.
   `grep -c '^SLACK_ALLOWED_USERS=' ~/.hermes/.env` → expect 1 (grep -c
   only, never cat). `slack.allow_bots` must stay unset/`none` — it is what
   keeps connector posts from triggering Tars (`docs/facts.md`, allow_bots
   row). Both are asserted again in the effective-value check below.

## What to apply

### (a) Code — roll back the withdrawn gate, apply the corrected patch

The VM checkout `~/.hermes/hermes-agent` is at **`5af9e1e`** (the withdrawn
gating commit) on top of `d615ca8` on top of upstream `6e87d43`. Reset to
`d615ca8` — this discards `5af9e1e` and nothing else — then apply the new
patch:

```bash
scp patches/0002-tool-progress-conversations.patch gaetan@192.168.0.9:/tmp/
ssh gaetan@192.168.0.9 '
  cd ~/.hermes/hermes-agent &&
  git log --oneline -3 &&                      # expect 5af9e1e, d615ca8, 6e87d43
  git status --short &&                        # expect only the known JS lockfile drift
  git diff > /tmp/js-drift.patch &&            # preserve the drift across the reset
  git reset --hard d615ca8 &&
  git -c user.name="Claude (Tars deploy)" -c user.email="noreply@anthropic.com" \
    am --3way /tmp/0002-tool-progress-conversations.patch &&
  git apply /tmp/js-drift.patch ;              # restore drift; non-fatal if empty or
                                               # conflicting — it is regenerable npm churn
  git log --oneline -2 &&
  venv/bin/python -c "import ast; ast.parse(open(\"gateway/run.py\").read()); print(\"AST_OK\")"
'
# afterwards, clean up: ssh … 'rm -f /tmp/0002-tool-progress-conversations.patch /tmp/js-drift.patch'
```

Expected HEAD subject: `feat(display): per-conversation tool_progress
override (display.platforms.<platform>.tool_progress_conversations)`, then
`d615ca8`'s DM-scoped subject. The `-c` identity is inline because the VM has
none configured — without it `git am` aborts mid-chain, after the reset. The
replayed sha differs from the kit's recorded commit — the tree does not.

**If `git am` fails:** `git am --abort`, then `git apply --3way` the patch
and commit manually; verify with
`git diff --stat HEAD~1` → `gateway/run.py | 10 +` and
`tests/gateway/test_slack_tool_visibility.py | 584 +`. **If the reset is
wrong** (checkout was not at `5af9e1e`, or carried unexpected local work):
stop, do not force it — `git reflog` still holds the old tip.

### (b) Config

Merge into `~/.hermes/config.yaml` under `flock ~/.hermes/.wf3.lock`, `.bak`
copy first, **merge into the existing structure, never append** — a raw
append once nearly dropped a sibling agent's `mcp_servers` stanza. Use a
ruamel round-trip script, the same mechanism the 2026-08-10 deploy used
(`hermes config set` is rejected: it rewrites the file with plain PyYAML and
drops comments).

Two changes, both in this one edit:

```yaml
# REMOVE this whole top-level block — the key it feeds is gone from the code:
platform_tool_conversations:
  slack:
    - C0BP2GZUFSR
    - D0BBYNM01BL

# Narration policy becomes chat-ID-exact, via the map ONLY:
display:
  platforms:
    slack:
      tool_progress: "off"               # SET — live value re-measured 2026-08-10
                                         # as `all` (gate-era config left display
                                         # broad). QUOTED on purpose: bare `off`
                                         # parses as YAML-1.1 boolean False under
                                         # PyYAML; the string "off" is the shape
                                         # the test suite validates.
      # tool_progress_dm: REMOVED (was `all`) — the adapter collapses MPIMs
      # into chat_type "dm" (is_dm = channel_type in {"im","mpim"}), so this
      # chat-TYPE override narrates group DMs, a shared surface. The chat-ID
      # map below subsumes it; patch 0001 stays in the kit (code path +
      # upstream compat) but no config key uses it.
      tool_progress_conversations:       # NEW (patch 0002) — the only narration source
        C0BP2GZUFSR: all                 # home channel narrates
        D0BBYNM01BL: all                 # the 1:1 IM, by id — MPIMs do not inherit
```

Already live and **not re-applied** by this plan (verify only, deployed
2026-08-10): the top-level `slack:` block —
`require_mention: true`, `strict_mention: true`,
`free_response_channels: C0BP2GZUFSR`, `unauthorized_dm_behavior: ignore`,
no `allowed_channels`; and the removal of the three shadowed duplicates from
`gateway.platforms.slack`.

Untouched by design: `allow_bots`, `allowed_users`/`SLACK_ALLOWED_USERS`,
`thread_require_mention`, `require_mention_channels`,
`platform_toolsets.slack`, global `display.tool_progress: all`.

**Nothing goes into `.env`.** `SLACK_STRICT_MENTION` /
`SLACK_FREE_RESPONSE_CHANNELS` must stay absent there: "the YAML→env bridge
only writes an env var when it is unset (`adapter.py:8992-8995`), so an
`.env` entry would silently win forever" (impl-report §3).

### (c) Gateway restart — REQUIRED

`systemctl --user restart hermes-gateway.service`

This is a **code** change: `config.yaml` is re-read on every turn, but the
running process never reloads `gateway/run.py`. The display key alone would
take effect on the next message (it is read from the per-turn raw config),
but the code that reads it does not exist in the running process until the
restart.

(Separately, and unchanged: `strict_mention`, `free_response_channels` and
`allowed_channels` land in `PlatformConfig.extra` / process env, a snapshot
built exactly once at gateway startup — `load_gateway_config()` →
`SlackAdapter.__init__`, never rebuilt in-process. Their env stamp is
set-if-absent and never cleared, so *removing* such a key also requires a
restart. That half is already deployed and restarted.)

## Pre-restart verification

1. **YAML parses, and the obsolete key is gone.**
   ```bash
   python3 -c "
   import yaml, os
   d = yaml.safe_load(open(os.path.expanduser('~/.hermes/config.yaml')))
   assert 'platform_tool_conversations' not in d, 'obsolete key still present'
   s = d['display']['platforms']['slack']
   print(s)
   assert s['tool_progress_conversations'] == {'C0BP2GZUFSR': 'all', 'D0BBYNM01BL': 'all'}, s
   assert s['tool_progress'] == 'off', s   # the STRING — quoted in the YAML on purpose
   assert 'tool_progress_dm' not in s, s   # removed — MPIMs collapse to chat_type dm
   print('CONFIG_OK')
   "
   ```
2. **Effective-value check** — replaces a raw `hermes config get
   slack.strict_mention` read: that reads YAML text and prints `true` in
   *both* the working top-level layout and the shadowed nested layout, so it
   cannot detect the regression (verify-mention.md §3). This reads the
   value the adapter will actually use, and confirms the deployed
   mention-routing half is still intact after the config edit:
   ```bash
   cd ~/.hermes/hermes-agent && venv/bin/python -c "
   from gateway.config import load_gateway_config, Platform
   from plugins.platforms.slack.adapter import SlackAdapter
   pc = load_gateway_config().platforms[Platform.SLACK]
   a = SlackAdapter(pc)
   print('strict', a._slack_strict_mention())
   print('require', a._slack_require_mention())
   print('free', sorted(a._slack_free_response_channels()))
   print('allowed', sorted(a._slack_allowed_channels()))
   print('reaction_triggers', a._slack_reaction_triggers())
   print('mention_patterns', a._slack_mention_patterns())
   import os
   au = getattr(a, '_slack_allowed_users', lambda: None)() or os.environ.get('SLACK_ALLOWED_USERS')
   print('allowed_users_nonempty', bool(au), 'gaetan_in', 'U08BDJAMSRZ' in str(au))
   print('allow_bots', a.config.extra.get('allow_bots'))
   "
   # Expect (unchanged): strict True | require True | free ['C0BP2GZUFSR'] |
   #         allowed [] (empty = no channel whitelist) | reaction_triggers None | mention_patterns []
   # and the who-gate, now the sole gate in front of a fully-armed Tars:
   #         allowed_users_nonempty True gaetan_in True | allow_bots None
   ```
   Run this read-only, VM-runnable check before the restart (it re-runs the
   loader in a throwaway process — env is only stamped inside it) and again
   after, to confirm the restart actually picked it up.

## Post-restart verification

1. **Capability is uniform — ASSERT, don't eyeball.** After the reset,
   `_get_platform_tools` is upstream again: it takes no `chat_id`, which is
   *why* a toolset cannot vary by conversation. Pin both the signature and a
   non-empty resolution:
   ```bash
   cd ~/.hermes/hermes-agent && venv/bin/python -c "
   import inspect
   from hermes_cli.config import load_config
   from hermes_cli.tools_config import _get_platform_tools
   from gateway.display_config import _normalise
   params = inspect.signature(_get_platform_tools).parameters
   assert 'chat_id' not in params, ('withdrawn gate still present', list(params))
   n = len(_get_platform_tools(load_config(), 'slack'))
   print('slack toolsets', n)
   assert n > 0, n
   # display override resolves per conversation, and only where configured
   conv = {'C0BP2GZUFSR': 'all', 'D0BBYNM01BL': 'all'}
   print('home', _normalise('tool_progress', conv['C0BP2GZUFSR']))
   assert _normalise('tool_progress', conv['C0BP2GZUFSR']) == 'all'
   assert _normalise('tool_progress', conv['D0BBYNM01BL']) == 'all'
   assert conv.get('C0BFQ5WFYTB') is None and conv.get('GQ07CQXT7') is None
   print('ASSERTS_OK')
   "
   ```
   The signature assert is the deployed form of the patch's own guard test
   (`test_toolsets_never_vary_by_conversation`): if someone re-applies the
   withdrawn gate, `chat_id` reappears and this fails.
2. **The code is actually running.** `git log --oneline -1` in the checkout
   shows the corrected subject, and the restart timestamp in
   `~/.hermes/logs/` post-dates the `git am`.
3. Live Slack probes (native creds — the claude.ai connector cannot trigger
   Tars):
   - Fresh `@mention` in a non-home channel (`C0BFQ5WFYTB` or `GQ07CQXT7`)
     asking for something that **requires a tool** → the answer arrives and
     reflects real tool output, and **no interim progress bubble** appears.
     This is the invariant the withdrawn patch broke: capability yes,
     narration no.
   - Unmentioned reply in an existing thread in the same channel → silence
     (mention routing unchanged — regression check).
   - Unmentioned top-level message in `C0BP2GZUFSR` that runs a tool → reply
     **with** the progress bubble (free-response + verbose).
   - Message in `D0BBYNM01BL` that runs a tool → reply **with** the progress
     bubble (via its `tool_progress_conversations` map entry —
     `tool_progress_dm` is no longer set).
   - Per operating norms Tars replies in-thread, so poll
     `conversations.replies` on the probe message's `ts`, not
     `conversations.history`, or a same-thread reply will read as silence.
   - Cross-check the quiet-channel probe in `~/.hermes/logs/agent.log`:
     `tool_executor` lines for that session prove the tool really ran even
     though Slack showed no bubble.

## Expected side effects

- **Behavioral change, deliberate:** Tars answers a fresh @mention in *any*
  channel or group DM it is a member of, **with full tools**; only the
  running commentary is hidden outside `C0BP2GZUFSR`/`D0BBYNM01BL`.
  `SLACK_ALLOWED_USERS` still limits *who* can trigger it to Gaetan.
- Any in-flight session is interrupted by the restart; auto-resume recovers
  it (~1 min lost — same pattern as the `d615ca8` restart,
  `status/lane-a.md`).
- `platforms.slack.home_channel.gateway_restart_notification: true` posts a
  restart notice into `C0BP2GZUFSR`. Expected, not a boundary violation — home
  channel.
- First turn in each conversation after the restart is a cold prompt-cache
  miss.

## Rollback

**Never roll back to `5af9e1e`** — that resurrects the withdrawn capability
gate (`patches/README.md`: "never re-apply it"). The safe floor is
`d615ca8`:

```bash
git -C ~/.hermes/hermes-agent reset --hard d615ca8
systemctl --user restart hermes-gateway.service
```

The corrected config is harmless on `d615ca8` code: the map key is unread
there, so with `tool_progress: "off"` and no `tool_progress_dm`, narration
is simply quiet everywhere (capability untouched, full) until roll-forward.
Only if the config edit itself must come off, restore the
`config.yaml.bak-*` copy taken by this apply **after deleting its
`platform_tool_conversations` block** (the .bak predates the fix and still
carries the gate key) — under the same flock, then restart. The two halves
remain independently revertible.

## Risks

- **Environment**: the corrected patch was validated on Python 3.14.6 in a
  scratch venv (20 tests green, 135-test regression across 12 files); the VM
  runs 3.11.15. Nothing in the 10-line hunk is version-sensitive, but the
  suite has not run on 3.11 — pytest is not installed there.
- **Malformed per-conversation *values* normalise to verbose.** A
  non-mapping `tool_progress_conversations`, or a chat id absent from it, is
  ignored. A garbage *value* (`{C0…: {"nested": 1}}`) is **not**: shared
  `_normalise` maps anything unrecognised to `"all"`. That is inherited
  verbatim from `tool_progress_dm` — same call, same semantics — and is
  cosmetic-only in both cases. A typo makes a channel chattier, never less
  capable. Left as-is rather than made stricter than its sibling.
- **Order does not matter.** The corrected code ignores a stale
  `platform_tool_conversations` key entirely (nothing reads it; pinned by the
  patch's guard test, which leaves the key in the fixture config on purpose),
  so the config-key removal and the code deploy need not be atomic and can be
  done in either order.

## Supersedes P1-thread-follow

`docs/proposals/P1-thread-follow.md` proposed `strict_mention: false` (with
`require_mention: true` unchanged) specifically to *loosen* thread-follow —
letting Tars answer unmentioned replies in threads it was already active in.
This change sets `strict_mention: true` instead: the opposite direction, by
Gaetan's directive after the `C0BFQ5WFYTB` regression (an unmentioned
"Mrc le goat" woke a tool call in a shared channel — the exact failure mode
P1's own risk table flagged as open blast radius, never resolved). P1 is
dead as written; its `allowed_channels` cap (P1-b) is dropped entirely —
the corrected boundary is mention-gating, and nothing else: not channel
membership, and not per-conversation toolsets.
`docs/proposals/README.md` should be updated to reflect this at apply time —
not done here, per the write-only scope of this session.
