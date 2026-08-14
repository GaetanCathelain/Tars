# Shared cross-agent skill installation

Use this pattern when one skill package must serve more than one agent runtime on the same host or across a controlled shared checkout.

## Canonical-source pattern

1. Keep one versioned package directory as the source of truth. The package includes `SKILL.md` and every durable `references/`, `templates/`, `scripts/`, and `assets/` file.
2. Install consumers with a directory symlink when their loader supports symlinked skills. A directory link avoids two copies drifting after later updates.
3. Never overwrite an existing consumer path. Inspect it first; if it exists, compare and stop for disambiguation.
4. Use absolute links only when the canonical checkout path is stable on that host. For portable bundles or relocatable checkouts, prefer a safe relative link or a checksum-gated copy.
5. Keep secrets and generated files outside the package. A shared skill directory must be safe for every consumer to read.

## Verification

Prove all of the following before reporting readiness:

- The consumer path is a symlink and its resolved real path equals the canonical package directory.
- `SKILL.md` and every linked support file are readable through the consumer path.
- A deterministic manifest of relative paths and SHA-256 hashes matches between the canonical and consumer views.
- Each runtime's own discovery command lists the skill as enabled. A filesystem link alone is not discovery proof.
- Already-running sessions are reloaded or the report states that the skill becomes active only in a fresh session.

## When not to symlink

Use a checksum-gated copy instead when the consumer runs in a container, on another machine without the shared checkout, under a sandbox that rejects symlinks, or when the canonical path can move. In that case, keep one documented source of truth and repeat the manifest comparison after every synchronization.