# Skill Migration Checklist

Use this as a concise evidence sheet for a host-to-host or profile-to-profile migration.

## Discovery

- Record source SSH target, user, resolved Hermes home, active profile, and absolute skill path.
- Record destination resolved Hermes home, active profile, and absolute skill path.
- Inventory hidden and regular files, modes, symlinks, and total durable-file count.
- Read all frontmatter fields, especially `name`, version, platform/environment constraints, required commands, optional environment variables, and `user-invocable`.
- Search source plugin/command/bundle/config trees for the skill name without reading or printing secret stores.

## Safe Transfer Rules

Exclude at minimum:

```text
__pycache__/
*.pyc
*.pyo
*.log
*.tmp
*.temp
.DS_Store
*~
```

Do not transfer `.env`, OAuth/auth stores, browser-cookie databases, private keys, password-store contents, or machine-specific credentials. Preserve executable modes. Accept symlinks only when they are safe and package-relative.

Stage before destination installation. A useful durable manifest row is:

```text
<SHA-256><two spaces><relative/path>
```

Sort paths identically on source, staging, and destination. Compare both path sets and checksums.

## Existing Destination

If a destination exists:

1. Produce source-only, destination-only, and changed-path lists.
2. Back up or snapshot changed destination package files when appropriate.
3. Overlay intended package files without `--delete` by default.
4. Never remove destination-only files until their ownership and relevance are known.

## Runtime Verification

Run all applicable checks:

1. Package help or syntax/import smoke test with the runtime version the engine actually requires.
2. Package doctor/diagnostic in a mode that masks credential values.
3. Harmless mock, dry-run, or fixture invocation.
4. `hermes skills list` and confirm category/source/status.
5. `skill_view(name='<name>')` and record `skill_dir`, readiness, missing requirements, and linked files.
6. Inspect Hermes's generated skill-command map or interactive autocomplete and confirm `/<normalized-name>` resolves to the installed skill.
7. Remove test-generated caches, rebuild the destination durable manifest, and compare it with the source manifest one final time.

## Slash Invocation Interpretation

Hermes generally scans installed `SKILL.md` files and generates `/skill-name` commands from frontmatter names. A standalone plugin is only needed when source inspection proves additional behavior exists outside the skill directory. Core-command name collisions may suppress auto-registration; in that case document the supported explicit skill-loading fallback rather than claiming the slash command exists.

For a process started before installation, use `/reload-skills`; a new Hermes process should scan the package automatically.

## Final Report Fields

- Source host and absolute path
- Destination absolute path and profile
- Imported durable-file count and grouped inventory
- Existing-destination handling
- Source/staging/destination checksum results
- Installed non-secret prerequisites
- `hermes skills list` result
- `skill_view` readiness result
- Slash invocation result and exact syntax
- Package smoke-test result
- Missing optional credentials, first-run consent, or degraded source coverage
- Confirmation that source was not modified and no credentials were copied
