# Design conversation — 2026-08-07 (Gaetan × Claude, cooper)

The conversation that produced PLAN.md, condensed turn by turn. Gaetan's messages are quoted
(lightly trimmed); Claude's answers are summarized with their substance intact. The raw session
transcript is deliberately **not** committed: it embeds the whole mc-metarepo knowledge context,
which does not belong in this personal repo.

---

## Turn 1 — "graph engineering?"

> There's the next big thing after loop engineering, it's graph engineering. How could we set you
> up with the capability to have graph runs? Maybe there's a GH repo that already works or you can
> do it natively. Do some research.

**Claude:** Researched (web sweep + local audit). Conclusion: no repo needed — Claude Code's
native **Workflow tool** is already a graph runtime (`agent()` nodes, `pipeline()`/`parallel()`
edges, schema outputs, checkpoint-resume, 16∥/1000 agents). Kestra on cooper covers durable/cron
graphs (gbrain flows prove it, incl. headless `claude -p` nodes). Third-party orchestrators
(ruflo ~67k★, oh-my-claudecode ~38k★) are increasingly wrappers over native features — skip.
The term itself: coined ~July 2026 (loop engineering June 2026; Steinberger post mid-July;
LangChain claimed prior art within days); real but mid-hype. Full report:
[graph-engineering-research.md](graph-engineering-research.md).

## Turn 2 — the pitch

> Tell me (don't implement) which graph/workflow you would build (ideally a visual) for this
> project: [PITCH.md](../PITCH.md). Small help: we can fully create the VM while the
> authentication shenanigan is running — these could be parallel in our graph.

**Claude:** Designed the five-workflow graph (WF1 recon fan-out → lane A creds ∥ lane B
provision → WF3 wire → gated cutover → WF4 verify → WF5 later) and published the visual
([graph.html](graph.html), artifact
<https://claude.ai/code/artifact/eadbfef8-7526-41e1-b27b-0703a61b8693>). Shape driven by three
house constraints: the logged-in browser is main-session-only → auth is a serial human lane with
async probe fan-outs; workflows can't pause for a human → chain of workflows with the session as
decision point; the Slack-app cutover is destructive → explicit gate, evidence-first, nothing
deleted until the new side is smoke-tested. WF4 exists because "done is not exercised": 14
probes incl. the negative test (non-Gaetan message ignored).

## Turn 3 — two sessions?

> Plan looks good. I'm guessing we could work on Lane A in one session and Lane B in another?

**Claude:** Yes, and better than one session: context isolation (lane A's browser noise will hit
the 50% handoff threshold, and background workflows don't survive a handoff — same-session
resume) and failure isolation. Rendezvous design: lane A session = orchestrator (owns join,
WF3, cutover); lane B pings done via SendMessage; durable state committed to this repo, sessions
ephemeral.

## Turn 4 — Orca setup, shared files, best practices

> How to set that up with Orca? One session per prompt? Common TARS-PLAN.md / TARS-TODO.md /
> TARS-ROADMAP.md? Best practices?

**Claude:** **The prompt is a pointer, the plan lives in the repo.** Three files only: PLAN.md
(written once, then read-only) + `status/lane-a.md` + `status/lane-b.md` — **single writer per
file** makes merge conflicts structurally impossible; both commit to `main` with pull-rebase.
No shared TODO.md (two-writer conflict machine), no ROADMAP.md (WF5 is one section in PLAN.md).
Orca note: the browser tab is resolved by worktree, which enforces "browser belongs to lane A"
at the tooling level. Kickoff order: lane A first (WF1 → commit PLAN.md), then lane B points at
a file that exists.

## Turn 5 — single managing session? options?

> How about a single Claude session spawning and managing lanes, subworkflows if necessary? The
> manual 2 sessions seem too hands-on and don't cover the pre-work (WF1 recon). We have
> /handoff-orca and /start-orca; Orca can do subworktrees.

**Claude:** The browser constraint settles half: lane A can never be delegated — the
orchestrator IS lane A's executor. Four options weighed:

1. Solo session, everything in-process — simplest, but handoff-fragile (orphans an in-flight
   WF2).
2. **Hub-and-spoke (chosen):** one orchestrator session runs WF1 in-process (covers the recon
   pre-work), spawns lane B itself via start-orca with a pointer prompt, does lane A
   interactively, receives lane B's ping, runs WF3/WF4 in-process, holds at the cutover gate.
   Handoff-safe both directions.
3. One mega-workflow — deadlocks on the human (WF3 needs lane A's creds; workflows can't wait).
4. Herdr-pane fleet — re-adds manual multiplexing for zero gain here.

## Turn 6 — this repo

> Clone the Tars project, move your findings, the pitch and the graph artifact there. Structure
> the findings; you can even include this full conversation. Then open the project on Orca.

**Claude:** Done — this commit. Repo scorched-earth as pitched (one leftover remote branch
`tars/v2-build`, deletion deferred). Orca has its own orchestration layer (runs, tasks, workers,
decision **gates**) — flagged in PLAN.md as a candidate mechanism for the cutover gate, to
evaluate during WF1 rather than assume.

## Open items when the build starts

- Run WF1 (recon fan-out) → commit the decision doc → spawn lane B.
- Delete the `tars/v2-build` remote branch (pitch authorizes it).
- WF1's Orca-control node decides: Orca orchestration gates/tasks vs plain SendMessage for the
  lane rendezvous.
