# Context-max evidence — thread ts 1786635368.803689 (D0BBYNM01BL)

Investigation date: 2026-08-13. READ-ONLY on the VM — no edits, no restarts.

## Verdict (one line)

**Warning, not a real limit.** The message is a one-time informational banner
about the compaction *trigger threshold* (not actual token usage), and it
fired on this thread only because it was the first agent-init after a 15:24 UTC
gateway restart that cut the compression engine over from the `hermes-lcm`
plugin to the built-in `compressor` engine (GCN-49, commit `0fae076`). The
session backing this thread was brand new (created the same second as the
first message) and nowhere near its context window.

## 1. Verbatim Slack text

Thread parent (Gaetan, 1786635368.803689, 2026-08-13 17:36:08 CEST /
15:36:08 UTC):
> Add/Update tasks to start tomorrow:
> • Engineer Tars <> Orca orchestration
> • On Kestra ticket: benchmark Gbrain + finalize + push

Reply 1 of 11 — **Tars bot U0BBH85NAKH**, ts 1786635370.992139, 2026-08-13
17:36:10 CEST / 15:36:10 UTC (2s after the parent message):

> :information_source: Codex gpt-5.6-sol caps context at 272K, so
> auto-compaction was raised to 85% (from 50%) to use more of the window
> before summarizing.
>   Opt back out: `hermes config set compression.codex_gpt55_autoraise false`

This is the ONLY message in the thread that resembles a "context max"
statement. Nothing in the thread literally says "context max" — Gaetan's
framing was a reasonable read of an ℹ (info) banner about a threshold being
*raised*, not a warning about being *at* a limit. The rest of the thread
(replies 2–11) is normal task-add / Linear-ticket conversation, unrelated to
context/compression.

Surrounding channel messages (D0BBYNM01BL, same window):
- 2026-08-13 17:24:10 CEST / 15:24:10 UTC — Tars: ":warning: Gateway shutting
  down — Your current task will be interrupted." (this is the gateway
  restart that flipped the compression engine, see §3)
- 2026-08-13 17:03:51 CEST / 15:03:51 UTC — unrelated Damien/1Password note
  tagged [GCN-49] (ticket-number coincidence with the compressor-restore
  commit's GCN-49, not the same event)

## 2. Log excerpts (VM: ~/.hermes/logs/)

Gateway restart, ~/.hermes/logs/gateway.log:
```
2026-08-13 15:24:10,283 WARNING gateway.run: Shutdown context: signal=SIGTERM under_systemd=yes parent_pid=27136...
2026-08-13 15:24:10,635 INFO hermes_plugins.slack_platform.adapter: [Slack] Disconnected
2026-08-13 15:24:15,489 INFO hermes_plugins.slack_platform.adapter: [Slack] Authenticated as @tars in workspace Mo...
2026-08-13 15:24:15,543 INFO hermes_plugins.slack_platform.adapter: [Slack] Socket Mode connected (1 workspace(s))
...
2026-08-13 15:36:02,601 WARNING hermes_plugins.slack_platform.adapter: [Slack] Early reject of unauthorized user U...
2026-08-13 15:36:09,840 INFO gateway.run: inbound message: platform=slack user=U08BDJAMSRZ chat=D0BBYNM01BL msg='A...
```

Last `hermes_lcm` (old plugin compressor) log line before the restart,
~/.hermes/logs/gateway.log:
```
2026-08-13 15:09:52,180 INFO hermes_plugins.hermes_lcm.engine: LCM condensation: d0 × 4 → d1 (L1, 5475→1789 tokens)
2026-08-13 15:09:52,183 INFO hermes_plugins.hermes_lcm.compaction: LCM leaf compaction finished in 75142.6ms
2026-08-13 15:09:52,195 INFO hermes_plugins.hermes_lcm.compaction: LCM compaction #4: 48 messages → 34 (1 leaf p...
```
No `hermes_lcm` lines appear anywhere after the 15:24 restart (checked the
full 15:24–15:40 window) — confirms the engine swap took effect at restart.

Session/agent init for the thread's own session, ~/.hermes/logs/agent.log:
```
2026-08-13 15:36:10,391 INFO run_agent: OpenAI client created (agent_init, shared=True) thread=hermes-gateway_0:...
2026-08-13 15:36:10,521 INFO agent.auxiliary_client: Vision auto-detect: using main provider openai-codex (gpt-5.6...
2026-08-13 15:36:10,849 INFO [20260813_153609_4fae1992] agent.turn_context: conversation turn: session=20260813_15...
2026-08-13 15:36:17,355 INFO [20260813_153609_4fae1992] agent.conversation_loop: API call #1: model=gpt-5.6-sol provider=openai-codex in=30656 out=94 total=30750 latency=6.4s
```
(A later turn in the same session, 15:41:15: `API call #1: model=gpt-5.6-sol
provider=openai-codex in=75234 out=1037 total=76271 latency=24.2s
cache=73216/75234 (97%)` — still far under the 272K cap / 231K (85%)
threshold.)

## 3. Source of the message — file:line

`/home/gaetan/.hermes/hermes-agent/agent/agent_init.py:244-273`
(`_build_codex_gpt5_autoraise_notice`):

```python
def _build_codex_gpt5_autoraise_notice(
    autoraise: Dict[str, Any], context_length: Optional[int] = None
) -> str:
    """Build the one-time notice shown when Codex gpt-5.x raises compaction.
    ``autoraise`` is ``{"model": <slug>, "from": <old_ratio>, "to": <new_ratio>}``.
    ...
    """
    model = str(autoraise.get("model") or "gpt-5.4/5.5").strip().lower().rsplit("/", 1)[-1]
    if isinstance(context_length, int) and context_length > 0:
        cap = f"{round(context_length / 1000)}K"
    else:
        cap = "128K" if model.startswith("gpt-5.3-codex-spark") else "272K"
    from_pct = int(round(autoraise["from"] * 100))
    to_pct = int(round(autoraise["to"] * 100))
    return (
        f"ℹ Codex {model} caps context at {cap}, so auto-compaction was raised "
        f"to {to_pct}% (from {from_pct}%) to use more of the window before "
        f"summarizing.\n"
        f"  Opt back out: hermes config set compression.codex_gpt55_autoraise false"
    )
```

Gating logic that decides whether/when to show it, same file, lines
2716–2772 (`_show_autoraise_notice` block): the notice is shown **at most
once per profile**, tracked by a marker file
(`_codex_gpt55_autoraise_notice_marker()` → `agent_init.py:311-318`, path
`get_hermes_home() / ".codex_gpt55_autoraise_notice"`). Comment at line
2716-2719 explicitly documents the mechanism this incident hit: *"the gateway
rebuilds the agent per inbound message ... without the persisted marker the
notice re-fires on every agent init"*, and *"a change in the raised threshold
(or the autoraised model) updates the marker state and re-notifies once."*

Threshold computation: `agent/auxiliary_client.py:659-694`
(`_compression_threshold_for_model`) — gpt-5.4/5.5/5.6 on the Codex OAuth
route get `0.85` instead of the global `0.5` because "Codex caps all three
families at 272K and the default 50% trigger would compact at ~136K."

## 4. Marker-file evidence — why it fired NOW, not on any of the prior gpt-5.6-sol turns

`~/.hermes/.codex_gpt55_autoraise_notice` — content `gpt-5.6-sol:50:85`,
`stat`:
```
Modify: 2026-08-13 15:36:10.824194996 +0000
Change: 2026-08-13 15:36:10.824194996 +0000
 Birth: 2026-08-13 15:36:10.824194996 +0000
```
Birth == Modify == Change, all at 15:36:10 — the file was **created for the
first time ever** at that exact moment (same second as the thread's agent
init). Yet gpt-5.6-sol had already been used in many sessions since at least
2026-08-10 (`agent.log.1: 2026-08-10 11:41:37 ... model=gpt-5.6-sol`, and
earlier the same day at `2026-08-13 14:55:24 [20260813_140745_980067c0]
API call #21: model=gpt-5.6-sol`). The notice never fired on any of those
turns because, per the code comment above, **a plugin compression engine
(`hermes_lcm`) bypasses the host `compression_threshold`/autoraise code path
entirely** — the host-level notice logic only runs when
`compression.engine: compressor` (the built-in engine) is active. Before the
15:24 UTC restart, `hermes_lcm` was the active engine (still logging
compactions at 15:09:52, see §2); the restart flipped the live config to the
`compressor` engine (GCN-49 restore), and the very next agent init — which
happened to be this thread's first message — was the first time the
autoraise/notice code path ever ran, so it fired unconditionally (no marker
existed yet) and posted the one-time banner.

## 5. Config (non-secret keys only)

`~/.hermes/config.yaml`:
```
compression:
  enabled: true
  progress_notices: false
  threshold: 0.5
  target_ratio: 0.2
  protect_last_n: 20
  codex_gpt55_autoraise: true
  codex_app_server_auto: native
  idle_compact_after_seconds: 0
  engine: compressor
memory:
  memory_enabled: true
```
`engine: compressor` confirms the GCN-49 restore (commit `0fae076`,
"compressor engine restored, hermes-lcm removed") is live in config. Global
threshold is the default 50%; the per-model Codex autoraise pushes the
*effective* threshold for gpt-5.6-sol turns to 85% (not user-set — computed
by `_compression_threshold_for_model`).

## 6. Session-per-thread mapping (Hermes thread↔session model)

`docs/session-lifecycle.md` (in the hermes-agent source tree) confirms
`SessionSource.chat_type` distinguishes `"dm"` / `"group"` / `"channel"` /
`"thread"`, and `thread_id` "differentiates threads" for session-key
purposes. Empirically confirmed from `~/.hermes/sessions/sessions.json`
(single 115,795-byte file, 91 entries, mtime 2026-08-13 15:36): session keys
for this DM are of the form
`agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:<thread_ts>` — **one session key
per Slack thread**, not one per channel. The key for this investigation's
thread is literally
`agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786635368.803689`, i.e. it is
freshly minted for this thread; there is no shared/bloated channel-level
session that a "first message of a fresh thread" could be secretly riding on.

Session metadata for that exact key (from `sessions.json`, metadata fields
only — no message content read or printed):
```
session_id          = 20260813_153609_4fae1992
created_at           = 2026-08-13T15:36:09.841107
updated_at           = 2026-08-13T15:36:09.841107   (identical to created_at)
input_tokens         = 0
output_tokens        = 0
cache_read_tokens    = 0
cache_write_tokens   = 0
total_tokens         = 0
last_prompt_tokens   = 0
chat_type            = dm
```
`created_at == updated_at` and all token counters at `0` (snapshot taken at
the moment the notice fired, before the entry was updated post-turn) proves
this session had not made a single API call yet when the banner posted. The
actual first API call moments later (`agent.log`, §2) used `in=30656` tokens
— ~11% of the 272K Codex cap and ~13% of the 85%-raised (231K) compaction
trigger. Not remotely close to any limit.

## 7. Compressor engine activity/errors since GCN-49 (commit 0fae076)

- `hermes_lcm` plugin logged routine no-op/compaction activity continuously
  through 2026-08-12 and up to 2026-08-13 15:09:52 — no errors observed.
- No `hermes_lcm` log lines after the 15:24:10 UTC restart — clean cutover,
  no crash/error logged for the old engine's shutdown.
- No compressor-engine errors found in `~/.hermes/logs/errors.log` for the
  15:24–15:52 UTC window (only one unrelated `skill_manage` tool warning at
  15:41:15 in `agent.log`, orthogonal to compression).
- No further autoraise-notice re-fires observed after 15:36:10 (marker file
  present, single write) — behaving as designed ("shown at most once per
  profile").

## Answer to "why context-max on a fresh thread — real limit or just a warning?"

Just a warning, and a mistimed/misleading one at that: it's an ℹ info notice
that the compaction *trigger percentage* was auto-raised (50%→85%) for the
Codex gpt-5.6-sol route, not a statement that the session is near or at its
limit. It coincided with the very first message of a brand-new thread purely
because that message happened to be the first agent-init after the 15:24 UTC
gateway restart that cut compression over from the retired `hermes_lcm`
plugin to the restored built-in `compressor` engine (GCN-49) — the
one-time-per-profile marker for this notice had never existed before that
moment, regardless of how many prior gpt-5.6-sol turns had already run under
the old engine. Session token counters confirm the thread's session was
genuinely brand-new and empty (0 tokens) when the notice posted.
