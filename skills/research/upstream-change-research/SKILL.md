---
name: upstream-change-research
description: "Use when checking if a bug was reported or fixed upstream."
version: 1.0.0
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [github, upstream, issues, pull-requests, releases, regression, research]
---

# Upstream Change Research

Determine whether an observed bug, regression, or behavior has already been reported or fixed in an upstream GitHub repository. Produce an evidence-backed answer that distinguishes an exact match from merely related work and identifies the first released or still-unreleased version containing any fix.

Use primary GitHub evidence: issue/PR bodies and status, merge commits, repository commits, annotated tags, release pages, compare results, and discussion metadata. Do not treat search snippets, a closed PR, or a plausible commit message as proof by itself.

For reusable GitHub commands and API recipes, read `references/github-evidence-workflow.md`.

## Procedure

1. **Normalize the observed signature.** Record the user-visible symptom, platform/component, timing, event IDs/session keys, installed version and commit, and leading causal hypotheses. Separate facts from hypotheses.
2. **Search issue and PR text broadly.** Search open and closed issues and PRs using symptom phrases, code identifiers, acknowledgement/error strings, transport terms, and mechanism terms. Run both narrow exact-phrase and broader conjunctive searches. Save the best queries so a negative conclusion is auditable.
3. **Inspect candidate bodies.** Retrieve full issue/PR bodies and metadata. Classify each candidate as exact, close, or incidental. Explain the distinguishing evidence rather than grouping every duplicate-message report together.
4. **Trace PRs to landed code.** A closed PR may be merged, rejected, superseded, or salvaged into another commit. Check `merged`, `merged_at`, `merge_commit_sha`, and commit history. Follow references such as “salvaged from,” “supersedes,” and “fixes.”
5. **Inspect current and recent source history.** Search commits by mechanism and by the affected path since the relevant release. Verify whether an apparent fix actually touches the live code path after refactors or plugin migrations.
6. **Prove version containment.** Resolve annotated tags to commit objects and use GitHub’s compare API to establish whether the fix is an ancestor of each candidate release. Identify the first release containing it. Do not infer containment from dates alone.
7. **Check discussions availability.** Query `hasDiscussionsEnabled`; search discussions when enabled and state explicitly when disabled.
8. **Reconcile against the observed version.** If the reporter reproduced the bug on a version that already contains the supposed fix, do not call it fixed. Treat that as evidence of a distinct path, incomplete fix, or regression unless an exact later fix is proven.
9. **Recommend one action.** Comment on an existing issue only when the observed signature fits its scope and adds useful evidence. Otherwise recommend a new issue and list the closest precedents to reference.

## Evidence Standards

- Every external conclusion gets a direct GitHub URL.
- Report issue/PR title and current status exactly.
- Distinguish “PR merged” from “PR closed but salvaged elsewhere.”
- Distinguish inbound duplicate processing, outbound duplicate delivery, content duplicated inside one prompt, session-key concurrency, and cosmetic acknowledgement suppression.
- A fix that merely hides an acknowledgement is not a fix for duplicate delivery or incorrect busy-state activation.
- “No exact report found” must include the best search queries checked and the closest non-matches.

## Output Contract

Lead with the verdict, then provide:

1. Exact or closely related issue/PR URLs with title, status, and relevance.
2. Existing-fix assessment, including first released version or “open/unreleased.”
3. An explicit negative statement plus search queries when no exact report exists.
4. A concise recommendation: comment on an existing issue or file a new one.
5. Files modified and investigation blockers, if any.

Keep the report concise. Do not replay every search result; include only candidates that materially affect the verdict.

## Pitfalls

- **Quoting the whole search expression as one CLI argument incorrectly.** Prefer `gh search issues <terms> --repo OWNER/REPO`; validate the resulting query when the CLI reports an invalid repository qualifier.
- **Trusting local shallow history.** A packaged checkout may be grafted, shallow, locally generated, or have synthetic commits absent upstream. Use GitHub tag and compare APIs for release ancestry.
- **Equating dates with releases.** A commit before a release date is not proof it is in that release.
- **Calling a related duplicate bug exact.** Compare event cardinality, direction (inbound/outbound), timestamp spacing, session identity, and whether the busy handler fired.
- **Stopping at the original PR.** Maintainers may close an external PR and reapply it in a different path while preserving attribution; locate the actual landed commit.
- **Overstating negative search.** Say what surfaces and queries were checked; never imply exhaustive proof beyond GitHub’s indexed evidence.