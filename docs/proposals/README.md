# Proposals awaiting Gaetan's review

Nothing here is applied. Both VMs are untouched (one inert empty orchestration
row on cooper, `run_24d000a3a6b6`, noted in P3). Each file is self-contained:
what changes, exact text, order, risks, verification, rollback, decision box.

| # | Proposal | Needs | Independent? |
|---|---|---|---|
| [P1](P1-thread-follow.md) | Tars answers unmentioned replies in threads it's in | SOUL amendment + config + **restart** | yes |
| [P2](P2-emoji-trigger.md) | Emoji reaction as a trigger (e.g. 📋 → Kanban) + repair the broken 👀 ack | **your browser** (Slack scopes) + 2 lines | yes |
| [P3](P3-orca-v2-skill.md) | Tars drives real Orca sessions on cooper (v2 replaces tonight's v1) | one `SKILL.md` — **must land with P4** | no |
| [P4](P4-soul-guidelines-redesign.md) | SOUL redesign: separation of concerns, not confinement (12 changes) | `SOUL.md` replacement — **must land with P3**; `approvals:` mode is a separate call | no |
| [P5](P5-a2a.md) | A2A investigation — reduces to one go/no-go | p-Hermes **upgrade** blocks everything | yes |

## Reading order

1. **P4** — it carries the frame (and rule 8, the announce-then-act line worth
   arguing about) and the `approvals:` mode question, which is the real sudo gate.
2. **P3** — the mechanism P4's frame enables; they apply as one unit.
3. **P1** / **P2** — independent Slack behavior, either can land alone.
4. **P5** — one decision: upgrade p-Hermes or leave A2A parked.

## Cross-cutting risk, stated once

P1 widens **what wakes** Tars. P3+P4 widen **what a wake can do**. Either alone is
reviewable; together the blast radius is the *product*, not the sum — recommend
not landing both unreviewed on the same day.

With confinement dropped (P3+P4), `SLACK_ALLOWED_USERS` + the adapter early-reject
becomes the **sole** gate on transitive Orca access, promoted from
defence-in-depth. It is proven working (`status/probes/wf4/p16-negative-rerun.md`)
and must not lapse.
