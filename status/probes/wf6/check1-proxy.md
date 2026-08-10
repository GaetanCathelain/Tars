# WF6 spec check 1 — proxy (hermes chat CLI)

**Goal**: a bare ask to create a ticket via the `linear-ticketing` skill must
produce a Linear issue in team GCN with exactly one workload label, an
explicit priority, explicit non-Backlog state, assignee Gaetan, and a reply
containing the issue key/URL.

**Transport**: `ssh gaetan@192.168.0.9 "~/.local/bin/hermes chat -Q -q '<prompt>'"`
(oneshot pattern, non-interactive). VM clock UTC. No secrets printed, no
files edited.

**Result ticket**: GCN-14

---

## Step 1 — creation

Command:
```
~/.local/bin/hermes chat -Q -q 'Create a ticket: verify the WF6 check-1 creation path end to end. This is a deliberate test ticket; pick the label and priority you judge right.'
```

Reply:
```
session_id: 20260810_222845_92a98c
Created GCN-14: Verify the WF6 check-1 creation path end to end
https://linear.app/mobile-club/issue/GCN-14
- State: Todo
- Label: test-check
- Priority: Low — deliberate verification with no deadline or blocker
- Assignee: Gaetan
All fields were verified from Linear's returned issue state.
```

PASS — reply names key GCN-14 and includes the issue URL.

---

## Step 2 — independent verification (post-creation, pre-cleanup)

Command:
```
~/.local/bin/hermes chat -Q -q 'Show issue GCN-14: team, state, stateHistory, labels, priority, assignee. Just the facts.'
```

Reply:
```
session_id: 20260810_222926_82cada
Team: Gaetan (GCN)
State: Todo (unstarted)
State history:
- Todo — started 2026-08-10 22:29:11.903 UTC; current
Labels: test-check
Priority: Low (P4)
Assignee: Gaëtan Cathelain
```

### Per-criterion verdict

| Criterion | Observed | Verdict |
|---|---|---|
| Team GCN | `Team: Gaetan (GCN)` | PASS |
| Exactly one workload label | `Labels: test-check` (single) | PASS |
| Explicit priority (1–4) | `Priority: Low (P4)` | PASS |
| Explicit non-Backlog state | `State: Todo (unstarted)` | PASS |
| stateHistory never shows Backlog | Single entry: `Todo — started 2026-08-10 22:29:11.903 UTC; current` — no prior Backlog entry | PASS |
| Assignee Gaetan | `Assignee: Gaëtan Cathelain` | PASS |
| Reply contains issue key/URL (step 1) | Key `GCN-14` + URL `https://linear.app/mobile-club/issue/GCN-14` | PASS |

**All 7 criteria: PASS.**

---

## Step 3 — cleanup (cancel)

Command:
```
~/.local/bin/hermes chat -Q -q 'Cancel ticket GCN-14 — it was a deliberate WF6 test ticket you filed a moment ago.'
```

Reply:
```
session_id: 20260810_222952_61a81e
Canceled GCN-14. Verified status: Canceled.
```

### Fourth read — confirm cleanup

Command:
```
~/.local/bin/hermes chat -Q -q 'Show issue GCN-14: team, state, stateHistory, labels, priority, assignee. Just the facts.'
```

Reply:
```
session_id: 20260810_223033_aaabf1
Team: Gaetan (GCN)
State: Canceled (`canceled`)
State history:
- Todo (`unstarted`): 2026-08-10 22:29:11.903Z → 22:30:23.761Z
- Canceled (`canceled`): since 2026-08-10 22:30:23.761Z
Labels: test-check
Priority: P4 — Low
Assignee: Gaëtan Cathelain
```

**Cleanup verdict: PASS** — state is Canceled, full history confirms
Todo → Canceled transition, never Backlog at any point.

---

## Overall verdict: PASS (7/7 criteria, cleanup confirmed)

---

## Caveat (verbatim, as required)

Proxy transport: hermes chat CLI, not a Slack DM. The Slack DM surface
(allowlist, threading, delivery) is separately proven by live daily use,
including Gaetan's deploy-check exchange at ts 1786372813/1786372819 on
2026-08-10. The literal DM remains Gaetan's 30-second gold acceptance.
