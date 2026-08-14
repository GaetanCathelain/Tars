# Deleted Skill Package Forensics

Use this only when an authorized skill-library repair requires exact recovery and ordinary sources (git, backups, upstream, another host) are exhausted. It is a last-resort evidence procedure, not a substitute for backups.

## Recovery standard

A package is recoverable only when every durable file is obtained byte-for-byte from an authoritative remnant and its original path set and executable modes are known. If a recovered `SKILL.md` points to a missing reference, script, template, or asset, the package is incomplete: stage the proven files, but do not install or call the package restored.

## Proven ext4 approach

1. Preserve the target tree and avoid unnecessary writes. Confirm the filesystem and device read-only before forensic work.
2. Record historical package evidence first: prior discovery snapshots can establish paths, sizes, mtimes, frontmatter identity, and whether support files existed. Logs and usage metadata establish provenance but are not substitutes for file bytes.
3. Check inexpensive authoritative sources: git history and unreachable objects, backups/trash, open deleted handles, session tool results, remote sibling hosts, and exact upstream packages.
4. On ext4, inspect directory metadata and deleted inodes with `debugfs`. Modern ext4 may report zero deleted inodes because inode metadata is cleared promptly; that does not prove data blocks are gone.
5. With explicit authorization and read-only access to the block device, scan raw bytes for a long, package-specific literal. A frontmatter line such as `name: <canonical-name>` is better than a generic description. Implement the scan in fixed-size chunks with `len(needle)-1` overlap so boundary-spanning matches are not missed.
6. For each hit, inspect bounded surrounding bytes. Accept a candidate only when a complete file boundary can be established independently—for example, historical exact byte length plus intact frontmatter/body and a plausible terminator. Hash the exact recovered byte range immediately and stage it outside the active skill tree.
7. Search separately for every linked support path and executable. A raw filename hit may occur only inside another file's prose or logs; it does not prove the referenced file was recovered. Require complete bytes, path, and mode for each package member.
8. Compare recovered evidence against historical manifests, usage timestamps, tool logs, and adjacent metadata. These can corroborate creation time and identity, but never fill missing prose.
9. Install only after the complete durable path set is proven. Preserve modes, run the deterministic package hash and credential-name scan, then verify `skill_view` and discovery.

## Safety and reporting

- Never reconstruct missing prose from a description, model memory, or neighboring fragments.
- Do not scan credential stores or copy secret-bearing artifacts into staging.
- Raw-device reads may surface unrelated sensitive data. Keep extraction narrowly bounded, do not print unrelated blocks, and delete forensic dumps after the repair decision.
- Report exact source device/offset or source object, byte length, SHA-256, staged path, missing members, and why installation was or was not performed.
- A byte-exact `SKILL.md` plus an unrecovered linked reference is a useful partial recovery, but the correct outcome remains **blocked**, not “restored.”
