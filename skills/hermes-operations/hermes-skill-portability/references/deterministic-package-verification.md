# Deterministic Package Verification and Provenance

Use this procedure when a migration is gated by an expected whole-package checksum or must preserve registry provenance.

## Framed package hash

For each durable regular file, sorted by normalized POSIX relative path, update SHA-256 with:

1. `len(path_utf8)` as an unsigned 8-byte **big-endian** integer
2. UTF-8 path bytes
3. `len(raw_content)` as an unsigned 8-byte **big-endian** integer
4. Raw file content bytes

Do not include modes in this package hash; compare modes separately. Exclude symlinks unless explicitly accepted as safe package-relative links. Also exclude caches, bytecode, logs, editor/temp/generated outputs, `.env`, auth/token/cookie stores, private keys, and credential-like files.

A checksum-gated migration should compare four states: source-before, staging, destination-after-install, and source-after. Require identical normalized path sets, per-file hashes, modes, and framed package hashes. A source-after mismatch means the source changed during migration and the affected package must not be certified.

## URL-sourced skills

The active package path is authoritative, not a guessed category path. Resolve it from frontmatter plus `$HERMES_HOME/skills/.hub/lock.json` when present. A URL-installed record commonly supplies `install_path`, `identifier`, source, trust, and owned file paths.

If preserving URL provenance at the destination is required, merge only the selected installed records into the destination lock after the package bytes pass verification. Preserve every pre-existing destination record and validate the resulting lock with `hermes doctor` or the current Skills Hub diagnostic. A direct file copy without selected registry metadata may correctly discover as a local skill but lose URL/update provenance.

## Scope proof

Snapshot the destination skills tree before installation as relative path → content hash + mode. Repeat after all discovery and smoke checks. Classify changes into:

- selected package files,
- expected registry metadata such as `.hub/lock.json`,
- usage metadata such as `.usage.json`, and
- unselected package files.

Require zero removals and zero unselected-package changes unless separately authorized. Usage metadata changed by `skill_view` is operational metadata, not package drift, but should still be disclosed.

## Slash-command smoke test

For each installed frontmatter name, verify that the generated `/<name>` command maps to the expected destination directory and can build a side-effect-free invocation message. In Hermes's Python API, `resolve_skill_command_key()` takes the **bare command name** (no leading slash), while invocation-message builders take the canonical slash key. Also document `/skill <name>` if a core-command collision suppresses auto-registration.

Packages containing only instructions and references may have no executable help command. In that case, successful `skill_view`, generated command resolution, and side-effect-free invocation expansion are the applicable package diagnostics; do not invent an executable test.
