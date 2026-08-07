# Tars orchestrator repo

Build + ops repo for **Tars**: Gaetan's personal Hermes agent on a dedicated VM,
live in Slack since 2026-08-07. Operational truth sheet: `docs/facts.md`.
Current state: `status/lane-a.md` §Log (newest first) · verification:
`status/wf4-report.md` · run graph + triage: `PLAN.md` §Phases.

## Environment

- **Tars VM:** `ssh gaetan@192.168.0.9` (also `tars` / 100.116.31.76 on tailnet).
  `hermes` is NOT on PATH over non-interactive ssh — use `~/.local/bin/hermes`.
  VM clock is UTC (cooper local = UTC+2).
- **p-Hermes** (Gaetan's personal Hermes, VM 103): `ssh phermes` from the Tars VM
  (user hermes, 192.168.0.8). Never touch the Proxmox host `192.168.0.3` as root.
- **Secrets:** `secrets/tars.sops.yaml` (SOPS+age, 2 recipients). Ingest:
  `scripts/tars-secret KEY` (stdin). On the VM, everything lives in
  `~/.hermes/.env` (0600).

## Hard rules (violations have been incident-logged)

- **NEVER whole-file `sops -d`** in any form (incl. `--output-type dotenv`).
  Only `sops -d --extract '["KEY"]' | ssh … 'umask 077; cat > file'`. Never a
  secret on argv, echoed, or in an evidence file. Two agents violated this in
  one day; both exposures are logged in `status/lane-a.md`.
- Every `~/.hermes/config.yaml` or `.env` edit on the VM: `.bak` copy first,
  then `flock ~/.hermes/.wf3.lock -c '<edit>'` — Hermes live-reloads (~30 s),
  and shared-file writes must merge, never append (a raw append once nearly
  dropped a sibling agent's `mcp_servers` stanza).
- Status files are single-writer: `status/lane-a.md` belongs to the orchestrator
  session only. Evidence goes in `status/probes/` (one file per probe/agent).
- This worktree cannot check out `main`: push with
  `git pull --rebase origin main && git push origin HEAD:main`.
- CLI facts in specs are LLM-summarized vendor docs — run `--help` on the VM
  before scripting any non-interactive invocation, and measure before believing
  recon claims (a recon agent's central claim was disproven by measurement the
  same evening it was written).
