# secrets

Encrypted Tars-build credentials — SOPS+age, the single canonical store (PLAN.md Amendments, `docs/recon/DECISION.md` D5).
Add/update a key: `scripts/tars-secret KEY_NAME` (prompts silently, upserts into `tars.sops.yaml`).
Read a key: `sops -d secrets/tars.sops.yaml | grep KEY_NAME`.
Recipients: cooper's age key today; the Tars VM's own age key is added as a second recipient + `sops updatekeys` once lane B reports it.
