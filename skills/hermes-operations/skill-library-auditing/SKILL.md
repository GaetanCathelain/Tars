---
name: skill-library-auditing
description: "Use when auditing or comparing active skill libraries."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [skills, audit, inventory, checksums, provenance, fleet]
---

# Skill Library Auditing

Audit one or more Hermes skill libraries without changing them. Use this for host/profile comparisons, drift reports, duplicate-name detection, active-only inventories, provenance reviews, and checksum-backed migration planning.

Read `references/checksum-diff-procedure.md` before implementing a cross-host package comparison. For byte-exact recovery after an installer or category collision, read `references/exact-collision-recovery.md`; never restore a package whose declared support files remain unproven.

## Core Workflow

1. **Resolve live scope.** Identify the active Hermes home/profile independently on every host. Use the live Hermes CLI/state rather than a pasted or historical inventory.
2. **Establish active inventory.** Include builtin, official, registry/URL/community, external, and local skills that are actually enabled and environment/platform eligible. Exclude disabled skills and archive trees.
3. **Match canonical identity.** Parse the canonical `name` from skill frontmatter, falling back to the package directory only when frontmatter omits it. Never equate directory path or category with identity.
4. **Resolve package roots.** Follow the same local-before-external precedence as Hermes discovery. Record every eligible candidate root before selecting one; flag duplicate canonical names instead of silently treating them as independent skills.
5. **Hash durable package content.** Hash every durable regular file under the selected package root using sorted normalized relative paths and raw bytes. Use explicit framing so path/content boundaries cannot collide. Exclude caches, bytecode, logs, temporary/editor files, generated outputs, credentials, and secret stores.
6. **Wait for stability.** When installs, imports, or curator work may be running, take two complete inventory-plus-checksum reads separated by a short interval. Report only after both reads match; retry rather than mixing states.
7. **Classify from one coherent snapshot.** For a directional comparison, partition every source-side active canonical name exactly once into source-only, same-name modified, or identical. A same-name package is comparable only when active on both sides and both roots resolve.
8. **Reconcile and report.** Show requested visible classes, exact checksums for modified matches, source/category when available, ambiguity/unresolved warnings, and counts proving the partition sums to the source active total.

## Deterministic Hash Contract

For each durable file in lexical order by POSIX-style relative path:

- encode the relative path as UTF-8;
- append its length as an unsigned 8-byte big-endian integer, then the path bytes;
- append the raw content length with the same framing, then raw content bytes;
- feed the resulting stream to SHA-256.

The contract hashes package identity by both path and content and avoids ambiguous concatenation. Do not hash absolute roots, mtimes, ownership, or host-specific metadata.

## Safety Rules

- Operate read-only unless the user separately authorizes migration or repair.
- Do not inspect or hash credential files merely to make two packages compare equal.
- Do not follow symlinks outside a package root. Flag nontrivial symlinks if their omission could affect portability.
- Do not let shell-output summarizers truncate machine-readable snapshots; use a raw execution mode or compact encoding, then parse and reduce locally.
- Do not call a package “identical” from `SKILL.md` alone when it contains references, templates, scripts, or assets.

## Verification Checklist

- Live active counts were obtained for every scope.
- Canonical-name candidates and package roots were resolved.
- Two reads were stable when concurrent mutation was possible.
- Every modified match includes exact checksums for both sides.
- `only + modified + identical == source active total`.
- Duplicate names, unresolved roots, excluded sensitive files, and excluded symlinks are explicitly reported.
