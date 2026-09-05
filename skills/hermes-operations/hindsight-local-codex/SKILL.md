---
name: hindsight-local-codex
description: Use when deploying Hindsight with dedicated Codex OAuth.
---
# Hindsight with dedicated Codex OAuth
Load hermes-agent and systemd-service-operations first and read current provider documentation.

- Isolate Hindsight in its own venv, HOME, work directory, database instance and loopback ports. Native LLM provider is `openai-codex`; use dedicated authenticated CODEX_HOME. Keep credentials 0600 and enclosing dirs 0700.
- Set localhost binding, API-key tenant extension and disable unused MCP. Prefer local embeddings/reranking. Never mix subscription credentials with the normal Codex home.
- Upstream startup output prints database DSNs even at warning level. Redirect stdout/stderr to a PRIVATE file from the first start and redact logs before display. Do not print auth/env files.
- Prove synthetic retain, separate-process recall, and reflect in a separate fixture bank. Close SDK clients explicitly. Production bank must not contain fixture data.
- Use supported Hermes `local_external` provider config; preserve unrelated config/.env. Test a fresh AIAgent using `_memory_manager`, tool schema presence and a real tool recall; distinguish this from a full model turn.
- Verify API unauthorized rejection, database bind/auth, readiness and persistence AFTER restart, and unchanged gateway PID. Systemd active is not readiness. Fresh-agent success is not live gateway activation.
- For pg0 credential rotation, rotate BOTH the application role and `postgres` in the same transaction: pg0 bootstrap authenticates as `postgres` using the configured application password. Rotating only `hindsight` leaves the app credential valid but breaks the next start. For recovery, start the existing cluster with bundled `pg_ctl` and its installation `lib` in `LD_LIBRARY_PATH`, loopback-only; authenticate with the current application credential, make a private logical backup, align the bootstrap role password, stop the manual instance, then restart the owning service. Never reinitialize data or weaken HBA authentication. Prove changed API PID and DB start time, preserved fact IDs, post-restart recall, and wrong-password rejection.
- If post-credential-rotation startup stalls, preserve sanitized logs and exact error; stop at the agreed stop-loss, disable only the new service and restore prior provider config. Preserve isolated data/auth for diagnosis; do not guess the root cause or claim completion.
