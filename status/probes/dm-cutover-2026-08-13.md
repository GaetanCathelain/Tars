# DM cutover — C0BP2GZUFSR retired, everything to D0BBYNM01BL

2026-08-13 ~12:30 UTC, orchestrator session (Fable), Gaetan's explicit go:
"Make sure we don't use #gcn-tars-reporting anymore, full DMs." Implemented by
workflow `wf_dd7b5487-8d4` (opus VM implementer + sonnet docs + independent
sonnet verifier), reconciler removal + SOUL edits + mirror sync by the
orchestrator. Related: `soul-rule10-context-before-answering.md` (same day).

## VM changes (all backed up, config edit flock-wrapped, merge not append)

- Cron: reconciler `3cee7db96367` paused first (won the race 3 min before its
  next tick), then REMOVED (`hermes cron remove`) — it hardcoded
  `--channel C0BP2GZUFSR` and would have silently reverted the job edits.
  `e231e5faf180` / `62e8cd9db637` / `759e08c598e3` / `3eb53a322510` repointed to
  flat `slack:D0BBYNM01BL` (threading retired with the channel — in a DM the
  reconciler's "day's earliest bot message" parent selection is meaningless).
- `config.yaml` (`.bak-20260813-dm-cutover`): `tool_progress_conversations`
  C0BP2GZUFSR entry removed; `free_response_channels` key removed —
  measured first: DMs are mention-exempt via `is_one_to_one_dm`
  (`adapter.py:5654` guard), the list only ever gated channels;
  `home_channel.chat_id` → `D0BBYNM01BL`.
- Gateway restarted (required: `free_response_channels` is startup-bound, not
  live-reloaded). Lifecycle notice delivered to `slack:D0BBYNM01BL` — doubling
  as home-delivery proof. ActiveState=active.
- `~/.hermes/memories/USER.md` l.15 rewritten: DM is the only progress surface
  (was "his DM and Tars's home channel", contradicting l.13).
- SOUL.md: rule 4 gains the surface paragraph (DM is the interaction surface;
  C0BP2GZUFSR retired 2026-08-13); new rule 11: a rejected/failed delivery is
  reported as failed, never rerouted (the Oli-misroute lesson). Mirrors synced,
  `diff` empty both files.
- Deliberate non-change: `HOME_CHANNEL = "C0BP2GZUFSR"` in
  `tests/gateway/test_slack_tool_visibility.py` (live + patch 0002) stays — it
  is a test-only fixture (gateway never imports it; narration routing comes
  from config), and pointing it at the DM collides with `OWNER_DM` in the same
  file, breaking the group-vs-DM tests for zero runtime benefit.

## Verification (independent agent, read-only, all PASS)

- `grep -n C0BP2GZUFSR ~/.hermes/config.yaml` → zero hits.
- `hermes cron list --all` → zero C0BP2GZUFSR deliver targets across all jobs
  (28 at verify time; reconciler since removed, re-confirmed 0 hits).
- USER.md l.15 names only the DM. Gateway active (entered 12:27:20 UTC).
- Full outputs archived in session scratchpad (`vm-dm-cutover.md`,
  `vm-dm-final-verify.md`).

## Still true / follow-ups

- `SLACK_MCP_ADD_MESSAGE_TOOL=D08K34MA3QT` (Oli DM allowlist) KEPT — Gaetan's
  explicit call, same day. Any slack-mcp `conversations_add_message` to another
  destination still fails; SOUL rule 11 now governs what Tars does about it.
- Existing Slack threads keep the pre-cutover SOUL forever (system prompt
  builds once per session, sessions never expire) — behavior tests need a NEW
  thread or DM turn.
- Channel archive itself is Gaetan's Slack-side action, sensible after one
  clean weekday of zero channel deliveries (nothing can deliver there anymore).
- `scripts/tars-report-thread.py` kept in repo as history (its cron job is
  gone; the retired spec `docs/specs/daily-report-threading.md` carries the
  banner).
