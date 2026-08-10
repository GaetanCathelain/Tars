# Proposals awaiting Gaetan's review

Nothing here is applied except **P7 (applied 2026-08-08)** and **P6 (applied
2026-08-10** — SOUL line + refresh cron; skill map still deferred). Both VMs are
otherwise untouched (one inert empty orchestration row on cooper,
`run_24d000a3a6b6`, noted in P3). Each file is self-contained:
what changes, exact text, order, risks, verification, rollback, decision box.

| # | Proposal | Needs | Independent? |
|---|---|---|---|
| [P1](P1-thread-follow.md) | **DEAD as written (2026-08-10)** — it proposed loosening `strict_mention`; Gaetan's directive after the `C0BFQ5WFYTB` regression deployed the opposite (`strict_mention: true`, no `allowed_channels`). See `docs/plans/apply-slack-channel-boundaries.md` §Supersedes | — | — |
| [P2](P2-emoji-trigger.md) | Emoji reaction as a trigger (e.g. 📋 → Kanban) + repair the broken 👀 ack. **Must be reconciled with the mention boundary before landing** — `reaction_triggers` is a `force_process` path that re-opens the gate (apply plan, Preconditions §3) | **your browser** (Slack scopes) + 2 lines | yes |
| [P3](P3-orca-v2-skill.md) | Tars drives real Orca sessions on cooper (v2 replaces tonight's v1) | one `SKILL.md` — **must land with P4** | no |
| [P4](P4-soul-guidelines-redesign.md) | SOUL redesign: separation of concerns, not confinement (12 changes) | `SOUL.md` replacement — **must land with P3**; `approvals:` mode is a separate call | no |
| [P5](P5-a2a.md) | A2A investigation — reduces to one go/no-go | p-Hermes **upgrade** blocks everything | yes |
| [P6](P6-knowledge-bases.md) | **APPLIED 2026-08-10** — Knowledge bases: Tars searches `mc-metarepo`, re-pulls hourly; push-back designed not built | one `SOUL.md` sentence + a refresh script + a cron job (skill map **deferred**, see §0/P6-c) | applied in its own session, as required |
| [P7](P7-direct-analysis-exception.md) | Direct-analysis exception: Tars writes read-only analyses/reports itself; build work stays delegated | `SOUL.md` rule 1 + `delegate-to-cooper` skill + repo docs, no restart | yes, but never in the same session as another `~/.hermes/` writer |

## Reading order

1. **P4** — it carries the frame (and rule 8, the announce-then-act line worth
   arguing about) and the `approvals:` mode question, which is the real sudo gate.
2. **P3** — the mechanism P4's frame enables; they apply as one unit.
3. **P2** — independent, but only after reconciling with the deployed
   mention boundary (P1 is dead as written, 2026-08-10).
4. **P5** — one decision: upgrade p-Hermes or leave A2A parked.
5. **P6** — independent of all of them in *effect*, but it writes under `~/.hermes/` like
   P3/P4 do, so it must not be applied on the same pass. Its blocking finding stands alone
   and is worth reading first: Tars' only knowledge-base pointer names a path that does not
   exist on the VM. Measurement-led (a scoreboard of 5 real questions); its push phase is
   design-only and presumes P3+P4 land.

## Cross-cutting risk, stated once

P1 is dead; the deployed boundary *narrows* what wakes Tars (fresh-@mention
outside the home channel). P3+P4 widen **what a wake can do** — reviewable
alone, but landing them together with any wake-widening change (e.g. a
revived P2) multiplies the blast radius: the *product*, not the sum.

P6 writes under `~/.hermes/` (`SOUL.md`, `scripts/`, a cron job) exactly where P3+P4 did —
reviewable alone, but **never applied in the same session as another `~/.hermes/` writer**:
a concurrent edit there has already nearly dropped a sibling's config stanza.

With confinement dropped (P3+P4), `SLACK_ALLOWED_USERS` + the adapter early-reject
becomes the **sole** gate on transitive Orca access, promoted from
defence-in-depth. It is proven working (`status/probes/wf4/p16-negative-rerun.md`)
and must not lapse.
