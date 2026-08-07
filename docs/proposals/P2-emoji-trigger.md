# P2 — Emoji reaction as a Tars trigger (e.g. react to add to Kanban)

**Status:** proposed, NOT applied · **Source:** thread-behavior peer, 2026-08-07 ·
**Evidence:** `status/probes/wf5/emoji-trigger.md` · **Decide:** Gaetan
· **Independent of [P1](P1-thread-follow.md)** — either can land without the other.

## Verdict: the feature ships with Hermes and is switched off, not missing

| Piece | State |
|---|---|
| `reaction_added` handler | Registered `adapter.py:1996` → `_handle_slack_reaction` `:4773` |
| Config key | `slack.reaction_triggers` exists |
| The action (Kanban) | Already live — 12 `kanban_*` tools, dispatcher + workers proven (`wf5/kanban-test.md`) |
| Threading | Already correct: the handler resolves `item.ts` to the true thread parent via `conversations.replies` (`:4861-4877`), verified live — so Tars answers **in the right thread** |
| Allowlist | Covers it: the reaction synthesizes a message with the **reactor** as user through the same chokepoint (`:5528-5549`). Only the *mention* gate is bypassed (`_hermes_force_process`) |
| Agent-class app | Not a blocker — reactions are in the agent base scope set (Slack AI-app docs; Hermes' own manifest generator `hermes_cli/slack_cli.py:91,104`) |
| **Blocker** | **`reactions:read` is not granted.** `auth.test` shows 17 scopes, without it |

## What to do

**(a) Slack app dashboard — human/browser action, Gaetan only.** On app
`A0BC0GXH78R`: add scope **`reactions:read`** and event subscription
**`reaction_added`** (skip `reaction_removed`), then re-install the app.
Slack scopes are additive → **no token change, no SOPS rotation expected**;
confirm with `auth.test` afterwards (expect 18+ scopes, same tokens).

**(b) Config**, one line under `gateway.platforms.slack` (flock + `.bak`):

```yaml
      reaction_triggers: [kanban]
```

**(c) SOUL**, one line mapping `reaction:added:kanban` to the Kanban tool so the
model knows what the trigger means.

## Bundled repair — `reactions:write` (recommended, same dashboard visit)

`reactions:write` is **also absent**, so Tars's 👀 acknowledgement reaction has
been **silently failing since cutover**. Pre-existing lapse, unrelated to the
emoji trigger, but it is the same browser visit and the same additive-scope
mechanics.

## Risk

Low. The allowlist still gates *who* may trigger (the reactor is checked), the
action surface is Kanban only (`reaction_triggers` is an explicit list), and no
gateway restart is implied by the scope grant itself — the config line follows
Hermes' usual reload path (confirm at apply time; `extra`-backed keys need a
restart, see P1 §3).

Note the interaction with P1: an emoji trigger is a *new* wake path, but a
narrow and explicit one. It does not widen what unmentioned messages do.

## Verification after applying

React with the chosen emoji on a message in `C08RWSTU9LK` → a Kanban card is
created and Tars confirms **in that message's thread**; `hermes kanban show`
on the VM confirms the card independently. Then react as a non-Gaetan user (or
inspect the log) → allowlist WARNING, no action.

## Decision

- [ ] Grant `reactions:read` + `reactions:write`, then apply (b) and (c)
- [ ] Only repair `reactions:write` (fix the broken 👀), skip the trigger
- [ ] Skip both for now
- Trigger emoji + meaning: ________ (e.g. 📋 → add to Kanban)
