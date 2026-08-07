# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Build + ops repo for **Tars**: Gaetan's personal Hermes agent on a dedicated
Proxmox VM, **live in Slack since 2026-08-07** (DM-first, answers only Gaetan).
Tars orchestrates and reports; it never implements — coding work is delegated
to cooper/Orca, and Tars never creates a PR. This repo holds no application
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
```

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
- p-Hermes (VM 103): `ssh phermes` from the Tars VM. Never root on `192.168.0.3`.
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
