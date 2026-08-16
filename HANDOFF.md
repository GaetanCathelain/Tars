# HANDOFF — GCN-50 Tars V2 wayfinder (successor hub brief)

**Date:** 2026-08-16 · Written by the outgoing hub session (context-exhausted).
**The canonical coordination surface is the Linear wayfinder map: GCN-50**
(https://linear.app/mobile-club/issue/GCN-50, label `wayfinder:map`). Load it
first; every decision lives in its closed child tickets' resolution comments.
Operate via the `/wayfinder` skill (work-through mode). One decision ticket per
session. Lean-model rule + evidence conventions are in the map's Notes.

## State (2026-08-16 ~10:15 UTC)

- **Closed (see Decisions-so-far on the map):** Knowledge ledger (GCN-57 →
  `docs/gcn50-knowledge-ledger.md`), End goal restated (GCN-58), Architecture
  re-derived (GCN-59), plus probes GCN-60/61/66/67/68/69 (busy-receiver,
  resume→wake, A2A/interop, Orca surface, herdr, deepseek-harness) — all with
  artifacts in `status/probes/gcn63-*` / `gcn59-*` on main.
- **In flight:** GCN-70 "A2A Claude-lane wrapper — live spike" — an Opus agent
  attached to the OUTGOING session. The outgoing session will verify + record
  its resolution and close GCN-70, then stop. If GCN-70 is still open with no
  resolution comment when you look: the artifact is
  `status/probes/gcn63-a2a-sdk-spike.md` on main — verify and record it yourself.
- **Frontier after GCN-70 closes:** GCN-63 "Pick the messaging transports"
  (grilling, the critical path — decision shape: A2A wrapper w/ Agent-SDK-hosted
  lanes as target vs herdr/Orca-send as pragmatic start; Hermes-side A2A already
  live on the VM; lane→Tars = two-rung `hermes send`/`--resume`, unbeaten) and
  GCN-62 "Lane lifecycle & failure handling" (grilling, independent).
  Then GCN-64 (security posture, blocked on 63) → GCN-65 (Assemble SPEC-V2,
  blocked on 62+63+64) finishes the map.

## Ops

- Push flow: `git pull --rebase origin main && git push origin HEAD:main`.
- PROPOSED-SPEC.md is superseded-pending-respec — never build from it.
- Hard rules (sops, VM edits) in CLAUDE.md. Codex on cooper currently 401s
  (token refresh) — unrelated, unfixed.
