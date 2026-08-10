# Slack tool visibility — restore, land on main, deploy

2026-08-10 (evening), Orca-dispatched worker, worktree
`orca-worktrees/Tars/slack-tool-visibility-correction`, branch
`GaetanCathelain/slack-tool-visibility-correction`. All times UTC.

## Why a restore

The correction lineage (`8466fbd` → `ea774fa` → `a99a186` → `de9f73d`,
"retract tool-capability gate, correct to per-conversation display") diverged
from `main` at `b12d802` and **never landed on a remote branch** — current
`origin/main` (`e4dda26`) carries none of it: no `patches/`, no apply plan, no
deploy record. Meanwhile the VM still runs the withdrawn gate. Gaetan asked
Tars in Slack whether the fix was pushed; this dispatch restores, lands, and
deploys it.

## What was restored, and how

Net final state of the lineage (`git diff b12d802 de9f73d`, +1620 lines,
purely additive except `docs/specs/tars-profile.md`) applied onto `e4dda26`
via `git apply --index` — clean, no conflicts (main's 17 commits since
`b12d802` touch none of these paths). Individual cherry-pick of `de9f73d` was
rejected: 4 of its 8 files never existed on main (modify/delete conflicts).
Intermediate wrong states (the gate patch itself) are not resurrected; the
withdrawn `0002-per-conversation-toolset-gating.patch` exists only in the
local lineage history, recoverable via `git show a99a186:patches/…`.

Adaptations on top of the verbatim restore:

- `docs/plans/apply-slack-channel-boundaries.md` — Status updated (apply
  performed this dispatch); the config section's
  `tool_progress: off  # unchanged` corrected to **SET**: live value
  re-measured 2026-08-10 as `all` (the gate-era deploy left display broad —
  capability was gated instead, so narration showed wherever tools existed).
- `docs/proposals/README.md` — P1 marked dead as written, P2 flagged
  must-reconcile (both required "at apply time" by the plan's §Supersedes and
  Preconditions §3; the plan's authoring session was write-scoped and could
  not do it).

## VM pre-state (measured 2026-08-10, read-only, before any change)

Raw grep of `~/.hermes/config.yaml` for the display/gate keys (line-numbered,
verbatim — the platform key at :95 is distinct from the global sibling
at :104):

```
95:      tool_progress: all          # display.platforms.slack.tool_progress
96:      tool_progress_dm: all       # display.platforms.slack.tool_progress_dm
104:  tool_progress: all             # display.tool_progress (global, untouched by design)
234:platform_tool_conversations:
235:  slack:
236:  - C0BP2GZUFSR
237:  - D0BBYNM01BL
```

`tool_progress_conversations` absent everywhere.

**Reconciliation — this contradicts two checked-in records.** `lane-a.md`
(2026-08-08 `d615ca8` entry) says the config was set to
`{tool_progress: off, tool_progress_dm: all}`; the 14:37 deploy record
(`wf5/deploy-2026-08-10.md`) lists a config diff with no `display` edits at
all, and the `config.yaml.bak-boundaries-20260810-143638` it names no longer
exists on the VM (only `SOUL.md.bak` remains). So the platform
`tool_progress` moved `off` → `all` somewhere between 2026-08-08 and this
measurement with no record; archaeology is inconclusive without the .bak.
The measured state is treated as authoritative for the edit — the apply
**sets** `tool_progress` rather than assuming it. Both prior records now
carry pointer annotations.
- Checkout `~/.hermes/hermes-agent` (shallow clone of
  `github.com/NousResearch/hermes-agent`, editable install — the gateway runs
  this tree): HEAD `5af9e1e` (the withdrawn gate), `d615ca8` in history,
  working-tree drift only the known JS lockfile files.
- `hermes-gateway.service` (--user) active, PID 677155, up since 14:37:08 UTC.
- No `config.yaml.bak`, no patch kit staged on the VM.

## Test evidence (pre-merge, scratch clone on cooper)

Full reproduction of the wf5 RED/GREEN/mutation methodology, fresh scratch
clone of `NousResearch/hermes-agent` at `6e87d43`, push disabled, venv
outside the repo. **Python 3.12.13** (repo caps `<3.14`; closer to the VM's
3.12 than the prior 3.14.6 validation), pytest 9.1.1, pytest-asyncio 1.3.0,
`pip install -e .[dev,slack]`.

- **Fidelity** — `git am` of `patches/0001-tool-progress-dm.patch` onto
  `6e87d43` yields tree `3f63c14…`, byte-identical to the VM's
  `d615ca8^{tree}` (read-only `git rev-parse` over ssh). The kit exactly
  reproduces the VM's local lineage.
- **Sanity** at the d615ca8-equivalent: `test_run_progress_topics.py`
  18 passed.
- **RED** (gate re-created from the withdrawn 743-line patch + new test file
  overlaid): `test_slack_tool_visibility.py` **5 failed, 15 passed** — the
  exact five predicted (three public-channel capability tests,
  `test_workshop_channel_shows_interim_progress`,
  `test_toolsets_never_vary_by_conversation`).
- **GREEN** (`git am --3way patches/0002-tool-progress-conversations.patch`
  on the d615ca8-equivalent): **20 passed**.
- **Regression**, broader than the wf5 run (46 files — 40 `tests/gateway/`
  + 6 `tests/hermes_cli/` matching progress/display/slack/tool patterns, vs
  12 files/135 tests then): **507 passed, 1 skipped, 0 failures** (16
  pre-existing async-mock RuntimeWarnings).
- **Mutation A** (production hunk reverted, tests kept): exactly 1 failure,
  `test_workshop_channel_shows_interim_progress` — the hunk is display-only.
- **Mutation B** (withdrawn gate diff re-applied on corrected): exactly 4
  failures — the three capability tests + the guard. The two axes stay
  orthogonal; the guard is a live anti-regression against the gate.

Deviation from prior evidence: regression scope only (46 files/507 vs
12/135 — superset, discovered by literal grep rather than a hand-picked
list). No test failures anywhere; nothing massaged.

**Extended suite after the adversarial review** (kit commit `ef35fd1`,
production hunk byte-identical to `0427536` — same post-image blob
`7fcfcf25f`; +5 tests, 20 → 25):

- New tests: `test_group_dm_stays_quiet_without_tool_progress_dm` (MPIM
  chat_type `dm`, map-only config: tool runs, zero narration),
  `test_dm_narrates_via_conversations_map` (D0BBYNM01BL loud via the map,
  not `tool_progress_dm`), fail-open ×2 (no `display` key at all /
  non-mapping map: turn completes, toolset non-empty),
  `test_bare_off_in_yaml_resolves_to_quiet` (measured: `_normalise` has a
  `value is False` branch → bare `off` ≡ `"off"`; the deploy still quotes
  it since the equivalence rests on that one branch), and the guard now
  also pins `'chat_id' not in signature(_get_platform_tools)`.
- Extended suite on `corrected`: **25 passed**, first run.
- Mutation A (production hunk reverted): **3 failed** — workshop display +
  the two new map-dependent tests; group-DM-quiet correctly survives (the
  hunk makes listed conversations loud, not unlisted ones quiet).
- Mutation B (withdrawn gate re-applied): **4 failed** — the three
  capability tests + the guard, now tripping on the signature pin
  (`chat_id: Optional[str] = None` visible in the assert), the
  mechanism-independent form.
- Regression re-run, same 46 files: **512 passed, 1 skipped** (Δ exactly
  the +5 new), 16 pre-existing async-mock warnings.

## Adversarial self-review (pre-PR, independent Opus pass over the full diff)

1 BLOCKER, 6 MAJOR, 14 MINOR, 3 NIT. Everything actionable fixed on the
branch before the PR opened; the full report is a session artifact, findings
of record:

- **BLOCKER — MPIM narration leak (fixed).** The adapter collapses MPIMs
  into chat_type `dm` (`is_dm = channel_type in {"im","mpim"}`, measured in
  `wf5/audit-slack-allowlist-bypass.md`), so the planned
  `tool_progress_dm: all` would narrate interim tool calls in group DMs — a
  shared surface, the exact original incident. The withdrawn gate had been
  masking it (no tools in MPIMs → nothing to narrate); removing the gate
  un-masks it. **Fix:** narration is chat-ID-exact via the map only —
  `tool_progress_conversations: {C0BP2GZUFSR: all, D0BBYNM01BL: all}`,
  `tool_progress_dm` removed from config (patch 0001 stays in the kit for
  the code path). MPIM + DM-map tests added to the shipped test file.
- **MAJOR — untested `off` shape (fixed):** bare YAML `off` parses as
  boolean `False`, a shape the suite never validated; the config now writes
  the quoted string `"off"` (the tested shape) and the pre-restart assert
  requires exactly the string. A False-shape pin test was added.
- **MAJOR — `git am` without committer identity (fixed):** the runbook
  command now carries inline `-c` identity; it previously failed mid-chain,
  after the irreversible-looking reset.
- **MAJOR — "Performed" vs pending contradiction (fixed):** plan Status now
  says authorized/dispatched, and flips to performed only with the deploy
  record in this file.
- **MAJOR — measurement vs records (fixed):** raw measured lines pasted
  above; both contradicted records annotated; drift documented as
  unrecorded.
- **MAJOR — rollback resurrected the gate (fixed):** rollback floor is now
  `d615ca8`, never `5af9e1e`; .bak restore only with the gate key stripped.
- **MAJOR — who-gate unverified (fixed):** `SLACK_ALLOWED_USERS`
  present/non-empty and `allow_bots` unset are now a positive precondition
  and asserted in the pre/post effective-value check.
- MINORs fixed: P1 cross-cutting paragraph rewritten; stale lane-a line
  citation; patches/README apply block labelled post-`/upgrade`-only with
  identity; JS-drift preservation across the reset; scp + /tmp cleanup
  steps; dangling `verify-*.md` refs disclosed in the plan header; deploy
  record's test-name/suite citations corrected; facts.md row added;
  fail-open + signature-assert tests added to the shipped suite; commit
  message inverted-additivity claim corrected.
- Accepted as-is (disclosed, deliberate): the test fixture keeps a stale
  `platform_tool_conversations` key as a tripwire; final-answer delivery is
  pinned at return-value level (live probe covers the send path); no
  negative-sender test (env-side control, covered by the live allowlist
  WARNING evidence); Python-version caveat (an `ast.parse` sanity runs at
  deploy).
- Review categories explicitly clean: secret safety; sender
  authorization/mention routing untouched; no unrelated changes; commit
  message otherwise accurate.

## Deploy record — performed 2026-08-10 19:57–20:01 UTC

Landed on main first: PR #39, squash-merged as `a420035` (parent `3c8c8f3`,
PR #40 from a concurrent lane — no path overlap), verified as fresh
`origin/main` HEAD post-fetch. No CI is configured in this repo; the merge
gate was the test evidence + adversarial review above.

Preconditions, all re-verified live immediately before the deploy: no
logged-in users, gateway idle since 19:25:39; `gateway.proxy_url` unset and
`GATEWAY_PROXY_URL` count 0; all eight forbidden `SLACK_*` env vars count 0;
`SLACK_ALLOWED_USERS` count 1 (grep -c only); checkout at `5af9e1e` with
only the known JS lockfile drift.

- **Code** (19:58): `scp` of the kit 0002; drift preserved via
  `git diff > /tmp/js-drift.patch`; `git reset --hard d615ca8`;
  `git -c user.name="Claude (Tars deploy)" -c user.email=… am --3way` →
  landed as VM commit **`4d10183`** (replay of kit `ef35fd1` — tree
  identical, committer differs); drift restored (exit 0, same three JS
  files dirty after); `ast.parse(gateway/run.py)` OK.
- **Config** (19:59:50): backup
  `~/.hermes/config.yaml.bak-toolvis-20260810-195950`, then the ruamel
  round-trip under `flock ~/.hermes/.wf3.lock`: removed
  `platform_tool_conversations`, set `tool_progress: 'off'` (quoted),
  removed `tool_progress_dm`, added
  `tool_progress_conversations: {C0BP2GZUFSR: all, D0BBYNM01BL: all}`.
  Unified diff captured; ruamel reflowed long `personalities` strings
  (cosmetic, precedented in the 14:37 deploy record) — semantic equality of
  everything outside the three intended keys ASSERTED
  (`yaml.safe_load(bak) == yaml.safe_load(new)` after popping the intended
  keys: `SEMANTIC_EQUAL_OK`). Top-level `slack:` mention-routing block
  untouched by the diff.
- **Pre-restart checks**: `CONFIG_OK` (obsolete key gone, map exact,
  `tool_progress == 'off'` the string, `tool_progress_dm` absent);
  effective adapter values unchanged — `strict True | require True |
  free ['C0BP2GZUFSR'] | allowed [] | reaction_triggers None |
  mention_patterns []`; who-gate positive — `allowed_users_nonempty True,
  gaetan_in True` (with `.env` sourced in the remote shell; the first,
  un-sourced run printed False — throwaway processes never see `.env`, the
  known gotcha), `allow_bots None`, and 42 historical
  `[Slack] Early reject of unauthorized user` WARNINGs in gateway.log.
- **Restart** (20:00:38): `systemctl --user restart hermes-gateway.service`
  — the supported external lifecycle path. Old PID 677155 (up since 14:37),
  new **PID 774992**, `ActiveState=active`; Slack Socket Mode connected
  20:00:43; nothing was in flight (last turn ended 19:25).
- **Post-restart asserts** (`ASSERTS_OK`): `chat_id` NOT in
  `_get_platform_tools` signature (the withdrawn gate is gone from the
  running code), **22 slack toolsets** resolve non-empty, `_normalise`
  resolves both map entries to `all` and unlisted conversations to nothing;
  effective mention-routing values identical pre/post; VM checkout
  `git log -1` = `4d10183`.

## Post-deploy verification

**Source/config-level evidence:** the config diff + `CONFIG_OK` +
`SEMANTIC_EQUAL_OK` above; kit patch on main (`a420035`) byte-applied by
`git am` (tree of `4d10183` ≡ kit `ef35fd1`).

**Runtime/service-level evidence:** PID 774992 active, socket connected,
`ASSERTS_OK` (signature + toolset count + map normalisation), effective
values unchanged — all listed above.

**Live Slack-message evidence** (2026-08-10 ~20:03–20:12 UTC, sent as
Gaetan via the VM-stored native creds — secrets never on argv; replies
polled via `conversations.replies` on each probe's ts; probe file
`~/tars-probe.txt` with a unique timestamp payload, removed after):

| # | Probe | Expected | Observed |
|---|---|---|---|
| A | `C0BFQ5WFYTB` fresh `@tars`, tool-requiring ask (read a file) | reply reflects real tool output, **no** interim bubble | thread = request + final answer only (**0** interim messages); answer = the probe file's exact 24-char content; `agent.log` session `20260810_200318_17b418ab`: `read_file completed (0.06s, 139 chars)` at 20:03:23 — **the tool ran, Slack showed no narration** |
| B | unmentioned `merci !` in A's thread | silence | no new message after 90 s+; thread count unchanged (regression check: mention routing intact) |
| C | `C0BP2GZUFSR` top-level, unmentioned, same ask | reply **with** bubble | thread = request + `⚙️` progress bubble + final answer (free-response + map narration) |
| D | `D0BBYNM01BL` 1:1 DM, same ask | reply **with** bubble | thread = request + progress bubble + final answer (narration via the map entry — `tool_progress_dm` is unset) |

Probe A is the invariant the withdrawn gate broke, now live-proven from
both sides: capability yes (log-verified tool execution), narration no.
Anomalies, both non-blocking and disclosed: A's turn first tried
`kanban_show` (errored, never surfaced to Slack) before `read_file`; one
local JSON poll capture was corrupted by shell command substitution
(local capture path, not Slack) and re-captured via file redirect.

**Not live-probed:** an MPIM (creating one would ping a third party —
same stance as the 14:37 deploy). Covered synthetically by
`test_group_dm_stays_quiet_without_tool_progress_dm` (MPIM chat_type
`dm`, map-only config: tool runs, zero narration) plus the fact that no
MPIM has a map entry; Tars is a member of zero MPIMs today
(`docs/facts.md`).

## Verdict

All six matrix points hold, each with config-, runtime- AND
live-message-level evidence (MPIM leg synthetic, as stated). The
withdrawn capability gate is out of the code, out of the config, and
pinned out by guard + signature tests; mention routing and the who-gate
are measurably unchanged.

## Post-deploy verification

_Pending — config-level, runtime-level, live-message evidence, each labelled._
