---
name: hermes-skill-portability
description: "Use when migrating Hermes skills between hosts/profiles."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [hermes, skills, migration, portability, verification]
---

# Hermes Skill Portability

Migrate or synchronize a Hermes skill between hosts, users, or profiles without losing package files, copying credentials, or assuming that a `SKILL.md` alone is the whole installation.

For the detailed audit and verification checklist, read `references/migration-checklist.md`. For checksum-gated migrations, URL provenance handling, scope-diff proof, and slash-command diagnostics, also read `references/deterministic-package-verification.md`. When one canonical package must serve several agent runtimes, read `references/shared-cross-agent-install.md` for the verified symlink-or-copy pattern. When an authorized repair requires byte-exact recovery of a deleted package after normal sources are exhausted, read `references/deleted-package-forensics.md`; never install a partial package whose linked support files or modes remain unproven.

## Workflow

1. **Resolve both homes.** Check `HERMES_HOME` and the active profile on each side. Do not hardcode `~/.hermes` when a profile or explicit Hermes home may be active.
2. **Inspect before installing.** Inventory the complete source skill directory, including hidden files, metadata, scripts, references, templates, assets, executability, and symlinks. Read frontmatter and identify required commands, Python/runtime versions, optional credentials, and first-run behavior.
3. **Check external wiring.** Search the source Hermes plugin, command, bundle, and configuration locations for references to the skill name. Distinguish real runtime registration from logs, snapshots, and history artifacts. Hermes normally creates `/skill-name` automatically from the skill frontmatter, so do not invent a plugin when none exists.
4. **Stage safely.** Copy to a temporary local staging directory first. Exclude caches and transient outputs (`__pycache__/`, `*.pyc`, `*.pyo`, logs, temp files, editor backups). Reject credential artifacts such as `.env`, auth/token stores, private keys, and browser-cookie databases.
5. **Prove staging fidelity.** Compute a deterministic manifest of relative path plus SHA-256 for durable regular files on source and staging. Compare exact path sets and hashes before writing the destination.
6. **Protect an existing destination.** If present, compare first. Update matching package paths safely; do not use destructive mirroring or delete unrelated files unless explicitly authorized. Preserve executable bits and safe relative symlinks. Before using a third-party multi-skill installer, compare every exposed skill basename with every existing top-level directory under the active skills home: a basename such as `research` may collide with a category directory and cause the installer to replace all nested skills. Back up any colliding directory before installation, or install/relocate the incoming skill to a collision-free folder and verify its frontmatter name still resolves.
7. **Install explicit prerequisites only.** Install non-secret dependencies clearly required by metadata or executable preflight. Do not invent API keys, copy another host's credentials, or silently enable browser-cookie access. Report optional coverage that remains unavailable.
8. **Verify the real package.** Check destination inventory, source-to-destination durable-file checksums, `hermes skills list`, `skill_view`, slash registration/invocation support, and an executable help/doctor/mock or harmless smoke path where supplied.
9. **Clean after testing.** Smoke tests may regenerate `__pycache__` and bytecode inside the installed package. Remove those transient files, then perform the final checksum comparison again.
10. **Report operational use.** Give source and absolute destination paths, imported file groups/counts, prerequisites changed, verification evidence, readiness, invocation syntax, and remaining consent/credential caveats. Mention `/reload-skills` for Hermes processes that were already running before installation.

## Verification Standard

A migration is complete only when all applicable checks pass:

- Durable source and destination manifests have identical paths and hashes.
- Hermes discovers the skill as enabled in the intended profile.
- `skill_view(name=...)` loads the installed directory and reports readiness.
- The generated slash command is present, or a documented collision/fallback invocation is identified.
- The package's own side-effect-free diagnostic or mock path executes successfully.
- No credential or transient artifacts remain in the destination.

## Pitfalls

- `python3` in frontmatter may be less specific than the engine's actual minimum version. Trust the executable preflight and help/doctor result.
- A successful copy is not proof of discovery; profile scope and disabled-skill configuration can still hide it.
- Running tests before the final manifest can create cache files and make the destination appear to contain files that were never imported.
- Searching for the skill name can find logs or achievement snapshots; these are evidence of history, not plugin dependencies.
- Avoid running first-run flows that read browser cookies or authorize accounts without explicit consent, even when the skill recommends automatic setup.
