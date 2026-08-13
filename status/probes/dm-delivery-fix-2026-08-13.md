# DM surface fixes — tool-progress narration off, cron deliveries top-level (2026-08-13, UTC)

Two Slack-surface defects in the home DM `D0BBYNM01BL`, both approved by Gaetan.
Applied live on the Tars VM (192.168.0.9). No gateway restart (config live-reloads);
`SOUL.md` untouched; `hermes-agent` checkout untouched.

Gateway `ActiveEnterTimestamp=Thu 2026-08-13 12:27:20 UTC` — unchanged across this work
(that restart predates it).

## Task A — tool-progress narration removed from the home DM

### A.1 Live state before (NOT what the brief predicted)

The brief expected `tool_progress_conversations: {C0BP2GZUFSR: all, D0BBYNM01BL: all}`.
The live file had only `D0BBYNM01BL`. A **dm-cutover already ran at 12:14 today**
(`~/.hermes/config.yaml.bak-20260813-dm-cutover`), which had already dropped
`C0BP2GZUFSR` from three places:

```
$ ssh gaetan@192.168.0.9 'diff ~/.hermes/config.yaml.bak-20260813-dm-cutover ~/.hermes/config.yaml'
105d104
<         C0BP2GZUFSR: all
137d135
<   free_response_channels: C0BP2GZUFSR
245c243
<       chat_id: C0BP2GZUFSR
---
>       chat_id: D0BBYNM01BL
```

So `platforms.slack.home_channel.chat_id` is now `D0BBYNM01BL`, and there was no
`C0BP2GZUFSR: all` left to keep. **Nothing was re-added** — restoring it would have
undone part of that deliberate cutover. Flagged for Gaetan: if the reporting channel
should keep narration, `display.platforms.slack.tool_progress_conversations:
{C0BP2GZUFSR: all}` is the one-line re-add.

### A.2 The edit

`.bak` first, edit under `flock ~/.hermes/.wf3.lock`, targeted two-line removal (no YAML
round-trip, so the rest of the file is byte-identical):

```
$ ssh gaetan@192.168.0.9 'cp ~/.hermes/config.yaml ~/.hermes/config.yaml.bak-20260813-narration \
    && flock ~/.hermes/.wf3.lock -c "python3 /tmp/edit_cfg.py"'
removed tool_progress_conversations block at line 104

$ diff ~/.hermes/config.yaml.bak-20260813-narration ~/.hermes/config.yaml
104,105d103
<       tool_progress_conversations:
<         D0BBYNM01BL: all
```

Resulting YAML, and it parses:

```
display:
  platforms:
    slack:
      tool_progress: 'off'
  tool_progress: all

$ python3 -c 'import yaml,json; …'
{"slack": {"tool_progress": "off"}}
home: {"platform": "slack", "chat_id": "D0BBYNM01BL", "name": "Home"}
```

`tool_progress_dm` deliberately NOT set (the adapter collapses MPIMs into `chat_type dm`,
so it would over-narrate group DMs). The whole key was removed rather than left as an
empty mapping — a bare `tool_progress_conversations:` parses to `None` and would break
`.items()`.

Narration that this kills, sampled from the DM thread before the fix
(ts `1786624981.167329`): `:book: Reading subagent-summary-0-20260813_124212_46...` /
`:alarm_clock: Scheduling create`.

## Task B — cron/reminder deliveries land top-level, never as thread replies

### B.1 Symptom

Engagement-checker output landed as **reply 6** inside the 8 August thread
`1786191017.826669` in `D0BBYNM01BL`:

```
Message TS: 1786626263.708849  (2026-08-13 15:04:23 CEST, from Tars U0BBH85NAKH)
• Configurer le paiement de l'abonnement Claude x20 de Lucas — Lucas attend depuis le
  déjeuner et l'échéance de 15h vient de passer. Next: finaliser le paiement avec la
  carte MC et confirmer à Lucas. [GCN-46]
```

Matches cron job `62e8cd9db637` `Last run: 2026-08-13T15:04:24 ok`. (The other ts the
brief cited, `1786624982.469059`, is Gaetan's own message in the same thread, not a
reminder.)

### B.2 Root cause — cron delivery inherits the job's origin `thread_id`

NOT the skill: `engagement-checker` never posts to Slack itself ("Never a Slack message,
email, or reaction from a source integration") — it returns text and cron delivers it.

`~/.hermes/hermes-agent/cron/scheduler.py`, `_resolve_single_delivery_target()`, the
`":" in deliver_value` branch:

```python
if (
    thread_id is None
    and platform_key == "slack"
    and origin
    and str(origin.get("platform") or "").lower() == platform_key
    and str(origin.get("chat_id")) == str(chat_id)
    and origin.get("thread_id")
):
    thread_id = origin.get("thread_id")
```

`--deliver slack:D0BBYNM01BL` carries no thread, so it **inherits the job's origin
thread**. `deliver=origin` does the same unconditionally, one branch above. And the
jobs were created inside that 8 August thread:

```
62e8cd9db637  deliver='slack:D0BBYNM01BL'  origin: chat_id=D0BBYNM01BL thread_id=1786191017.826669
e231e5faf180  deliver='slack:D0BBYNM01BL'  origin: chat_id=D0BBYNM01BL thread_id=1786191017.826669
759e08c598e3  deliver='slack:D0BBYNM01BL'  origin: chat_id=D0BBYNM01BL thread_id=1786191017.826669
dd647f879010  deliver='origin'             origin: chat_id=C0BP2GZUFSR thread_id=1786434385.110149
0e96ef800465  deliver='origin'             origin: chat_id=D0BBYNM01BL thread_id=1786191017.826669
```

Corroborated by `gateway_routing` in `state.db` (read-only, `mode=ro` URI): session key
`agent:main:slack:dm:T7V1UGJ82:D0BBYNM01BL:1786191017.826669` carries
`origin.thread_id = "1786191017.826669"`, `updated_at 2026-08-13T12:43:18`.

Bare `--deliver slack` takes a different branch: chat_id from `SLACK_HOME_CHANNEL`,
thread from `_get_home_target_thread_id("slack")` = env `SLACK_HOME_CHANNEL_THREAD_ID`.
Origin is never consulted. Verified without printing values:

```
$ grep -c SLACK_HOME_CHANNEL_THREAD_ID ~/.hermes/.env   → 0        (unset ⇒ thread_id None)
$ grep -q '^SLACK_HOME_CHANNEL=D0BBYNM01BL$' ~/.hermes/.env && echo MATCH
ENV SLACK_HOME_CHANNEL == D0BBYNM01BL
```

The "origin thread_id lost" WARNING the brief mentioned is the *inverse* diagnostic
(`scheduler.py` ~line 1673) — it fires when a target correctly drops the origin thread.
Not present in the current logs.

### B.3 What changed

`jobs.json` backed up to `~/.hermes/cron/jobs.json.bak-20260813-topleveldelivery`, then
six active jobs repointed with `hermes cron edit <id> --deliver slack`:

| job | before | after |
|---|---|---|
| `e231e5faf180` Gaetan daily work brief | `slack:D0BBYNM01BL` | `slack` |
| `62e8cd9db637` Gaetan engagement checker | `slack:D0BBYNM01BL` | `slack` |
| `759e08c598e3` engagement checker final pass | `slack:D0BBYNM01BL` | `slack` |
| `3eb53a322510` mc-metarepo-refresh | `slack:D0BBYNM01BL` | `slack` |
| `dd647f879010` NM/MC Brique 3 — Gaetan actions | `origin` (→ C0BP2GZUFSR thread) | `slack` |
| `0e96ef800465` GoCardless rotation | `origin` (→ DM thread) | `slack` |

Every active job now delivers `slack` (bare) or `local`; `Repeat` counters survived the
edit (`dd647f879010` still `2/47`, `0e96ef800465` still `0/1`). Inactive/one-shot-completed
jobs were left alone — several deliberately target other conversations (e.g.
`slack:C0BFQ5WFYTB:…` for Lucas) and rewriting them would be wrong.

Skills amended so the jobs cannot be recreated with the broken flag, mirrored
byte-identical into this repo:

- `skills/orchestration/engagement-checker/SKILL.md` 1.8.0 → **1.9.0** — "Scheduling
  contract" now mandates `--deliver slack`, bans `--deliver origin` and
  `--deliver slack:<chat_id>`, and states a reminder is never a thread reply and never
  goes to `C0BP2GZUFSR`.
- `skills/orchestration/daily-work-brief/SKILL.md` 1.5.0 → **1.6.0** — same fix. Its §6
  said "delivery to the configured reporting conversation", and that section is explicitly
  reconciler-managed ("do not let the two drift"), so leaving it would have re-broken the
  cron job on the next reconcile.

```
$ md5sum (VM)                                   $ md5sum (repo)
e00bf332aeeea729b60d2c94abcf9253 engagement-checker/SKILL.md   e00bf332aeeea729b60d2c94abcf9253
aaf33c2daf4b637075cafd0c8599330f daily-work-brief/SKILL.md     aaf33c2daf4b637075cafd0c8599330f
```

VM copies backed up as `SKILL.md.bak-20260813` in each skill dir.

### B.4 LIVE VERIFICATION — PASS

```
$ hermes cron create "1m" "Reply with exactly this text and nothing else, no preamble,
    no formatting: TOPLEVEL-PROBE-8f3c1d" --deliver slack --repeat 1 \
    --name "top-level delivery probe"
Created job: e8573a984585   Next run: 2026-08-13T15:16:24+02:00

$ hermes cron runs e8573a984585
3e459f36285e44c0a7c7d97feab97e0f  completed  job=e8573a984585  source=builtin  2026-08-13T15:16:27+02:00
```

Read back through the claude.ai Slack connector — `slack_read_channel D0BBYNM01BL`, i.e.
the **channel** surface, which only shows top-level messages:

```
=== Message from Tars (U0BBH85NAKH) at 2026-08-13 15:16:36 CEST ===
Message TS: 1786626996.701989
TOPLEVEL-PROBE-8f3c1d
```

No `Thread:` annotation on it, and `slack_read_thread D0BBYNM01BL 1786191017.826669`
re-read after the run still ends at `1786626263.708849` — the marker is not in that thread.
Probe job auto-removed after its single run (`--repeat 1`); `hermes cron list` no longer
lists it.

## Residual

- `C0BP2GZUFSR` narration is off because of the 12:14 cutover, not because of this work —
  re-add `tool_progress_conversations: {C0BP2GZUFSR: all}` if Gaetan wants it back.
- The scheduler's thread-inheritance is upstream behaviour in `hermes-agent`, untouched by
  design. Any future job created with `--deliver origin` or `--deliver slack:<chat_id>`
  from inside a thread will thread again; the two SKILL.md amendments are the guard.
