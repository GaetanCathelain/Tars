# Checksum-Backed Active Skill Diff Procedure

Use this procedure when comparing a source Hermes installation against a destination without modifying either side.

## Discovery model

Hermes discovery is not equivalent to recursively listing directories. Reuse the live installation's own discovery/configuration logic where practical so the audit honors:

- active `HERMES_HOME` and profile;
- global and platform-specific disabled names;
- platform and environment eligibility in frontmatter;
- local-before-external directory precedence;
- excluded directories such as `.archive`;
- canonical frontmatter names and duplicate suppression.

The CLI's enabled-only listing is the user-visible baseline. For checksums, map each canonical entry back to its selected package root while retaining all eligible candidate roots for ambiguity reporting.

## Provenance

Where available, combine the installed-skill lock/index with the bundled manifest:

1. A canonical name in the installed registry/URL lock uses that recorded source and trust/provenance.
2. Otherwise a name in the bundled manifest is builtin.
3. Otherwise classify it as local.

Do not infer provenance solely from the directory category or path.

## Durable-file exclusions

Use a documented, symmetric exclusion set on both hosts. Typical transient directories include:

- `.git`, `.svn`, `.hg`;
- `__pycache__`, `.pytest_cache`, `.mypy_cache`, `.ruff_cache`, `.tox`, `.nox`, `.cache`;
- `node_modules`, build/dist trees, coverage output;
- log, temp, and generated output directories.

Typical transient files include `*.pyc`, `*.pyo`, `*.log`, `*.tmp`, editor swap/backup files, `.DS_Store`, and coverage artifacts.

Exclude credential-like files without reading them: `.env*`, authentication/token stores, credential/secret files, private keys, and browser-cookie databases. Report the count excluded, not names or contents when that could leak information.

Do not follow symlinks. Either include safe symlink metadata under a separately declared hash contract or omit and flag them; never silently read an external target.

## Stable snapshot loop

A useful stability signature is the sorted tuple:

```text
(canonical_name, selected_root, package_sha256)
```

Take the full signature twice with a short delay. If it differs, wait and repeat until two consecutive reads match or report that a coherent snapshot could not be obtained. Do not combine inventory from one read with hashes from another.

## Classification

For every active canonical name on the directional source:

- **source-only**: no active canonical match on destination;
- **modified**: active on both, roots resolved, hashes differ;
- **identical**: active on both, roots resolved, hashes equal.

Keep inactive-but-present destination candidates out of the active comparison, but mention them separately if the request distinguishes “absent” from “disabled.” Never count duplicate roots as extra active skills; flag the ambiguity.

## Output contract

Include:

- timestamp/scope if requested;
- source and destination active totals;
- source-only, modified, and identical counts;
- only the requested visible classifications;
- category and source/provenance when available;
- exact source and destination SHA-256 values for every modified match;
- duplicate canonical names and unresolved roots;
- excluded sensitive/symlink counts;
- the reconciliation equation.

A valid directional result satisfies:

```text
source_only + modified + identical = source_active_total
```

If it does not reconcile, stop and fix identity/root resolution before reporting.

## Output transport pitfall

Large JSON or base64 snapshots may be truncated by command-output summarizers even when the remote command succeeds. Run the transport through a raw/pass-through execution mode, or compress and encode the payload and decode it locally before reducing it to the requested rows. Treat a truncation marker as incomplete evidence, not valid JSON.
