# Peer mandate — P6 proposal: knowledge bases in Hermes

From Gaetan, 2026-08-07. **Deliverable is a PROPOSAL, not an implementation.**
Nothing is applied to the VM, cooper, or any repo beyond your own evidence and
proposal files. Output: `docs/proposals/P6-knowledge-bases.md`, review-shaped like
the existing P1–P5 (read one first for the format: what changes, exact commands or
config, order, risks, verification, rollback, decision box).

## The question

How should Tars integrate knowledge bases — `mc-metarepo` first — so it can:

1. **Search them** — answer from their content, on demand, without stuffing them
   into the prompt.
2. **Re-pull hourly** — stay current automatically, and behave sanely when a pull
   fails, conflicts, or the repo has moved on.
3. **Later, push knowledge back** — write learnings into the KB. Gaetan's own
   framing: *"via Orca + cooper I guess."* Design this phase, do not build it.

Treat it generically: `mc-metarepo` is the first source, not the only one.
`gaetan-metarepo` exists and is **deliberately not registered in Orca**; Notion is
already live through its MCP server; the VM already holds a working clone of
mc-metarepo at `~/dev/mc-metarepo` (probe 8, `gh` device-flow auth, HEAD tracking
`origin/main`). Say clearly which sources your design covers and which it doesn't.

## Constraints that bind the design (do not design around them silently)

- **Tars does not author the deliverable.** Under the redesigned SOUL (see
  `docs/proposals/P4-soul-guidelines-redesign.md`, rule 1) Tars must not write
  documentation itself — *"not because it's only markdown"*. So the push phase
  cannot be "Tars writes a doc and commits it"; it must route through a delegated
  Claude Code session (the P3 v2 Orca skill). Say how, and what the brief looks like.
- **`mc-metarepo` is a team repo.** Read its own contribution rules (`AGENTS.md`,
  `CLAUDE.md`, `CONTRIBUTING.md` in the clone) and honour them — including any
  PR-only or directory-scoped restriction. Do not propose anything that writes
  straight to `main` without saying so explicitly as a decision.
- **Context budget is real.** Hermes runs an LCM context engine; a design that
  pastes documents into the prompt is wrong by construction. Prefer tool-driven
  retrieval (the model asks, the tool answers).
- Hindsight memory is deliberately **skipped for v1** (no keys). If your design
  wants a memory/RAG provider, that is a decision to surface, not to assume.

## What to investigate (evidence, not opinion)

1. **Retrieval mechanisms actually available in this Hermes build** — measure, do
   not guess: the shell tool over the clone (ripgrep/git grep), a skill that
   teaches Tars where things are, an MCP server (filesystem/docs/RAG), and
   anything Hermes ships natively (`hermes --help`, source at
   `/home/gaetan/.hermes/hermes-agent/`, config keys, existing `toolsets:`).
   Rank by simplicity — the laziest thing that answers real questions wins.
2. **Refresh**: `hermes cron create` (already proven: `docs/facts.md` p13) vs a
   systemd `--user` timer vs plain cron. Decide who owns the pull, what happens on
   failure (dirty tree, auth expiry, network), and whether Tars should be *told* a
   refresh happened or just benefit from it silently.
3. **What "search" must actually answer.** Try 3–5 real questions against the
   clone with each candidate mechanism and show the outputs — a design that looks
   elegant but answers badly is a fail. Cheap probes are fine; you may run
   `hermes chat -Q -q` turns on the VM (read-only, no config edits).
4. **The push phase**, designed not built: brief shape, which repo/paths, PR vs
   branch, who reviews, how Tars verifies the result, and what stops it becoming a
   slow leak of noise into a team repo.

## Hard rules

- Read `CLAUDE.md` and `docs/facts.md` first; they bind you (secrets, `flock`,
  never root on `192.168.0.3`, `--help` before scripting, measure before believing).
- **No VM config edits, no `~/.hermes/*` writes, no gateway restart.** Two sibling
  peers are applying P3/P4 to the live agent right now — stay out of
  `~/.hermes/SOUL.md`, `~/.hermes/skills/`, and `~/.hermes/config.yaml` entirely.
  Read-only inspection is fine.
- No pushes to `mc-metarepo`. No `sops` decryption of any kind — you need no secrets.
- Do not edit `status/lane-a.md` (single-writer: the hub), `CLAUDE.md`, or
  `README.md` (hub-owned).
- Commit your own files and push with
  `git pull --rebase origin main && git push origin HEAD:main`.

## Reporting — one-way channel

Hub→peer messaging is currently broken (peers are not addressable from the hub),
so: **do not wait for a reply from anyone.** SendMessage the hub your milestones
and your final verdict, and treat the committed repo as the real channel. If you
hit something that needs Gaetan's ruling, write it into the proposal as an open
decision with a recommendation — do not block on it.

Evidence → `status/probes/wf5/kb-*.md`. Proposal → `docs/proposals/P6-knowledge-bases.md`.
Update `docs/proposals/README.md`'s table with the P6 row when you land it.
