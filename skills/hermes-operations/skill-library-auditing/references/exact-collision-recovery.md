# Exact Skill-Package Collision Recovery

Use this procedure after an installer, migration, or category collision removes a pre-existing skill package. The goal is byte-exact recovery, not prose reconstruction.

## Evidence hierarchy

1. **Persisted tool-call arguments:** inspect the canonical Hermes session store and any active context-engine store for the original `skill_manage(create|write_file|edit)` arguments. Extract decoded string values as UTF-8 bytes and hash them immediately.
2. **Durable copies and manifests:** check backups, source repositories, package locks, prior inventories, remote hosts, and manifests that record relative path, size, mode, and checksum.
3. **Filesystem metadata and raw blocks:** only after higher-level sources are exhausted, use inode/extent metadata, journal records, validated directory entries, and bounded raw-block searches. Treat logs and summaries as search coordinates, not file-content evidence.

A timestamp proving that a write occurred does not prove its arguments. A fragment, plausible heading, or summary of the missing file is not a recoverable file.

## Completeness standard

Restore a package only when every declared durable file is exact. Accept completeness from one of:

- original tool-call bytes plus the successful tool result;
- filesystem metadata establishing exact size/boundaries plus intact bytes;
- multiple independently located, byte-identical complete copies;
- a checksum-backed authoritative source.

For each file record relative path, byte length, SHA-256, mode, and provenance. If `SKILL.md` references a missing support file, do not install `SKILL.md` alone.

## Repair and verification

After exact recovery is established:

1. Restore only the authorized package root and preserve modes.
2. Load the recovered skill and each required support file through `skill_view`.
3. Re-hash unaffected packages against their pinned source revision, including paths, bytes, and modes.
4. Confirm no other profile changed and no credential artifacts entered the package set.
5. Remove raw images, journal dumps, candidate fragments, and investigation scripts after recording checksums and provenance.

If exact recovery cannot be proved, leave the package absent and report the precise missing evidence rather than installing a reconstructed or partial package.
