# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Build + ops repo for **Tars**: Gaetan's personal Hermes agent on a dedicated
Proxmox VM, **live in Slack since 2026-08-07** (DM-first, answers only Gaetan).
Tars orchestrates and reports; it never produces the deliverable itself — the
work is delegated to Claude Code sessions driven through Orca on cooper, and
Tars never merges, approves or pushes. A PR authored by a session Tars
delegated to is fine; Tars authoring one is not. This repo holds no application
code: it is the plan, the specs, the secrets store, and the exercised evidence
that the wiring works. Operational truth sheet: `docs/facts.md`.

## Repo map — how the documents flow

```
PITCH.md → PLAN.md (run graph, §Amendments = settled decisions, §Phases = triage)
  → docs/recon/     r1..r8 evidence → DECISION.md (D1–D6)
  → docs/specs/     executable specs: tars-profile (SOUL/config/env map),
                    wf3-wiring, wf4-probes (+post-WF4 corrections & stubs §16–23),
                    cutover-notes, wf5-orca-delegation, wf5-kanban
  → status/         lane-a.md = orchestrator log & credential checklist (newest first),
                    lane-b.md = VM provisioning verdicts, wf4-report.md = verification
                    state, probes/ = one evidence file per probe/agent
  → secrets/        tars.sops.yaml (SOPS+age, 2 recipients: cooper + the VM)

  SOUL.md           live mirror of ~/.hermes/SOUL.md on the VM
  skills/<n>/SKILL.md  live mirror of ~/.hermes/skills/<n>/SKILL.md
```

**Tars edits its own skills, and commits them here.** `skill_manage` lets Tars
rewrite its own `SKILL.md` when a run teaches it something the file gets wrong —
it did so 7 times during the first live Orca run. **SOUL rule 2 requires it to
commit that edit to `skills/<name>/SKILL.md` and push, in the same turn.** So
this mirror is authored by Tars as well as by us: when it and the VM disagree,
find out which side moved before overwriting either. `artifacts/*-SKILL*.md` are
frozen historical snapshots (what a proposal approved, what a rollback needs) —
never the live copy.

Specs are corrected in place when probes disprove them (marked "corrected
post-WF4"); `status/lane-a.md` §Log is the authoritative timeline. Single
writer per status file — `lane-a.md` belongs to the orchestrator session only;
agents write new files under `status/probes/`.

## Commands

- Secret ingest: `scripts/tars-secret KEY` (silent prompt or stdin) — upserts
  into `secrets/tars.sops.yaml`; plaintext never touches argv or disk.
- Secret delivery to the VM (the ONLY sanctioned decrypt):
  `sops -d --extract '["KEY"]' secrets/tars.sops.yaml | ssh gaetan@192.168.0.9 'umask 077; cat > …'`
- VM access: `ssh gaetan@192.168.0.9` (tailnet alias `tars`; clock is UTC).
  Hermes CLI: `~/.local/bin/hermes` (NOT on PATH over non-interactive ssh).
  Units are `--user` (`export XDG_RUNTIME_DIR=/run/user/$(id -u)`); logs in
  `~/.hermes/logs/` (journald carries WARNINGs only).
- **p-Hermes** (Gaetan's personal Hermes, Proxmox VM 103) = **192.168.0.8**,
  user `hermes`: reach it with `ssh phermes` from the Tars VM. It runs Hermes
  v0.19.0 (older than Tars' v0.20.0). Read-only unless Gaetan says otherwise.
- **`192.168.0.3` is `pve`, the Proxmox HYPERVISOR — a different machine.**
  Never root on it. (These two were previously adjacent here and read as one
  host; the confusion misled an agent on 2026-08-07.)
- Push flow (this worktree cannot check out `main`):
  `git pull --rebase origin main && git push origin HEAD:main`

## Hard rules (violations have been incident-logged)

- **NEVER whole-file `sops -d`** in any form (incl. `--output-type dotenv`) —
  per-key `--extract` piped over ssh only. Never a secret on argv, echoed, or
  in an evidence file. Two agent violations in one day are logged in
  `status/lane-a.md`.
- Every `~/.hermes/config.yaml` or `.env` edit on the VM: `.bak` copy first,
  then `flock ~/.hermes/.wf3.lock -c '<edit>'` — Hermes live-reloads (~30 s) —
  and merge, never append (a raw append once nearly dropped a sibling agent's
  `mcp_servers` stanza).
- CLI facts in specs are LLM-summarized vendor docs: run `--help` on the VM
  before scripting any non-interactive invocation, and measure before believing
  recon claims.
- The gated cleanup (p-Hermes old-profile delete) needs Gaetan's explicit go.
