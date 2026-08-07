# Cutover runbook notes — collected gotchas (seed, finalized before the gate)

Decision base: PLAN §Amendments (disable-only) + DECISION D2 (paths) + D5 (A5 tokens).

1. **Order**: evidence bundle → Gaetan's explicit "go" in the orchestrator session →
   `systemctl --user disable --now hermes-gateway-tars.service` on p-Hermes (via
   `ssh root@192.168.0.3` → `qm guest exec 103` → `runuser -u hermes`) → copy `xoxb-`/`xapp-`
   into the new VM's `.env` (from SOPS; Socket Mode, static, host-independent) → set
   `SLACK_ALLOWED_USERS` → start gateway on new VM → `/sethome` in the live DM.
2. **`SLACK_ALLOWED_USERS=U08BDJAMSRZ`** — resolved by A4's probe 2026-08-07 (team `T7V1UGJ82`).
3. **Rollback** = stop new gateway, re-enable old unit. Old profile is NOT deleted at cutover —
   deletion is a separate gated cleanup after WF4 + soak.
4. **CloakBrowser stale-lock trap (lane B B7)**: `Restart=always` does not self-heal a stale
   Chromium profile lock — an orphaned Chromium on 9223 makes the unit restart-loop (chrome
   exit 21). Fix = kill the orphan process, not restarting harder.
5. **Config shadowing check** before debugging any guardrail: no stray top-level `platforms:`
   block may exist (it silently overrides `gateway.platforms.*`). Verified absent at B7.
6. **Bots don't auto-join**: re-invite the Tars app to each needed channel post-attach.
7. **RESOLVED 2026-08-07**: the 6 npm vulnerabilities are fixed on the VM (patch bumps only, no
   `--force`; doctor clean; evidence in `status/npm-vulns-fix.md`). Residual, accepted: the
   unused `apps/desktop` Electron workspace carries 5 unrelated vulns needing major bumps —
   out of Tars' runtime surface, left untouched.
