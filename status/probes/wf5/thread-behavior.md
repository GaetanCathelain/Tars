# WF5 — Thread behavior: mention-follow + context depth

Investigation peer session (worktree `thread-behavior`), 2026-08-07.
Mandate: `docs/plans/thread-behavior-investigation.md`.

Companion evidence files in this directory:

- `source-gating.md` — Slack adapter inbound gate, gate order, knob inventory
- `source-context.md` — session keying + what the model sees for a thread turn
- `live-config-snapshot.md` — live `config.yaml` / `SLACK_*` vars / session store
- `log-forensics-reference-thread.md` — gateway log lines for the reference thread

---

## 0. Baseline: the reference thread, reconstructed from Slack

Read via the claude.ai Slack connector (authenticated as Gaetan `U08BDJAMSRZ`),
`conversations.replies` on channel `C08RWSTU9LK`, parent ts `1786135149.072829`.
Permalink: <https://mobileclub-squad.slack.com/archives/C08RWSTU9LK/p1786135283424339?thread_ts=1786135149.072829&cid=C08RWSTU9LK>

Times below are **CEST as Slack renders them**; UTC = CEST − 2 h (the VM clock is UTC).

| # | Slack ts | CEST | UTC | Sender | Mention? | Tars replied? |
|---|---|---|---|---|---|---|
| parent | 1786135149.072829 | 22:39:09 | 20:39:09 | Gaetan `U08BDJAMSRZ` | **yes** | yes — 22:39:15 (6 s) |
| r1 | 1786135155.707859 | 22:39:15 | 20:39:15 | Tars `U0BBH85NAKH` | — | (tool-trace message) |
| r2 | 1786135178.649919 | 22:39:38 | 20:39:38 | Tars | — | (answer) |
| r3 | 1786135219.118499 | 22:40:19 | 20:40:19 | Gaetan | **yes** | yes — 22:40:25 (6 s) |
| r4 | 1786135225.606029 | 22:40:25 | 20:40:25 | Tars | — | (answer) |
| r5 | 1786135242.240479 | 22:40:42 | 20:40:42 | Gaetan | **yes** | yes — 22:40:59 (17 s) |
| r6 | 1786135259.991959 | 22:40:59 | 20:40:59 | Tars | — | (answer) |
| **r7** | **1786135270.459679** | **22:41:10** | **20:41:10** | **Gaetan** | **NO** | **NO — silently dropped** |
| r8 | 1786135283.424339 | 22:41:23 | 20:41:23 | Gaetan | **yes** | yes — 22:41:29 (6 s) |
| r9 | 1786135289.769869 | 22:41:29 | 20:41:29 | Tars | — | (answer) |
| **r10** | **1786135310.012819** | **22:41:50** | **20:41:50** | **Gaetan** | **NO** | **NO — silently dropped** |
| **r11** | **1786135439.766629** | **22:43:59** | **20:43:59** | **Nans `U05LKHLDV0A`** | **NO** | **NO** |

### What this establishes with zero new messages sent

1. **Q1 symptom confirmed live.** In a thread where Tars is an active participant
   and had just answered 13 s earlier, Gaetan's next message without a re-mention
   (r7) produced no reply. Same again at r10. The three mentioned messages in the
   same thread each got an answer in 6–17 s, so the silence is not latency —
   `require_mention` applies to thread replies exactly as it does to top-level
   channel messages. Thread participation buys nothing.

2. **A non-Gaetan is already in this thread.** Nans `U05LKHLDV0A` posted r11.
   This is not a hypothetical: any thread-follow change makes r11-shaped messages
   reach the gate that the mention requirement is currently absorbing. The
   allowlist must therefore be proven to run *before* (or independently of) the
   mention/thread branch before any change is applied — see `source-gating.md`
   §gate order and `log-forensics-reference-thread.md` §Nans.

3. r1 shows Tars emitting a tool-trace message (`kanban_show`,
   `mcp__slack__conversations_replies`, `conversations_history`) as a *separate
   Slack message* before its answer. Noted: the model has Slack MCP tools and can
   fetch thread history *as a tool call* — which is a different mechanism from the
   adapter pre-loading thread context. `source-context.md` disambiguates the two.

<!-- sections below filled in as agents land -->

## 1. Source: how inbound thread replies are gated

_pending — see `source-gating.md`_

## 2. Source: what the model sees for a thread turn

_pending — see `source-context.md`_

## 3. Live configuration

_pending — see `live-config-snapshot.md`_

## 4. Gateway log forensics

_pending — see `log-forensics-reference-thread.md`_

## 5. Live measurement

_pending_

## 6. Change applied / proposed

_pending_
