# GitHub Evidence Workflow

Reusable commands for determining whether an upstream bug is reported or fixed. Replace `OWNER/REPO`, terms, paths, tags, and SHAs.

## 1. Establish repository and release state

```bash
gh auth status
gh repo view OWNER/REPO \
  --json nameWithOwner,url,defaultBranchRef,latestRelease

gh api repos/OWNER/REPO/releases --paginate \
  --jq '.[] | {name,tag_name,published_at,html_url,target_commitish}'
```

`target_commitish: main` is not the release commit. Resolve the tag object before testing ancestry.

## 2. Search issues and PRs

Use several queries; GitHub search is conjunctive and wording-sensitive.

```bash
gh search issues 'platform duplicate' \
  --repo OWNER/REPO --include-prs --state open --limit 100 \
  --json number,title,state,url,updatedAt,isPullRequest

gh search issues '"exact acknowledgement text"' \
  --repo OWNER/REPO --include-prs --state closed --limit 100 \
  --json number,title,state,url,updatedAt,isPullRequest
```

For direct API search and result counts:

```bash
gh api /search/issues \
  -f q='repo:OWNER/REPO platform mechanism symptom' \
  -f per_page=100 --method GET \
  --jq '"COUNT \(.total_count)", (.items[] | "\(.html_url) | \(.state) | \(.title)")'
```

Useful query families:

- visible symptom and exact strings;
- component/platform plus `duplicate`, `replay`, `redelivery`, `interrupt`, or `busy`;
- code identifiers and config names;
- transport mechanism (`reconnect`, event type, webhook, Socket Mode);
- identity/routing terms (`timestamp`, event ID, session key, thread concurrency).

Search both open and closed results. Record the best query strings in the final negative finding.

## 3. Inspect full candidate evidence

The Issues API returns issues and PR stubs. Check whether `pull_request` exists, then inspect PR merge state separately.

```bash
gh api repos/OWNER/REPO/issues/NUMBER \
  --jq '{title,state,state_reason,html_url,created_at,closed_at,body}'

gh api repos/OWNER/REPO/pulls/NUMBER \
  --jq '{title,state,merged,merged_at,merge_commit_sha,html_url,body,head:.head.sha}'
```

Never report a closed PR as merged without the PR endpoint. A closed-unmerged PR may still have been manually reapplied or salvaged; search its number, title, author, and key phrases in commits.

## 4. Search commits

```bash
gh search commits 'component dedup reconnect' \
  --repo OWNER/REPO --limit 100 --json sha,commit,url

gh api /search/commits \
  -f q='repo:OWNER/REPO component mechanism' \
  -f per_page=100 --method GET \
  --jq '.items[] | {sha,html_url,message:.commit.message,date:.commit.committer.date}'
```

Recent commits touching the actual path are often more reliable than keyword search:

```bash
gh api repos/OWNER/REPO/commits \
  -f path=path/to/component.py \
  -f since=YYYY-MM-DDTHH:MM:SSZ \
  -f per_page=100 --method GET \
  --jq '.[] | {sha,html_url,date:.commit.committer.date,message:.commit.message}'
```

Confirm that the commit touches the current live path, especially after refactors or plugin migrations.

## 5. Resolve tags and prove containment

A Git tag ref may point to an annotated tag object, which then points to the commit.

```bash
TAG_OBJ=$(gh api repos/OWNER/REPO/git/ref/tags/TAG \
  --jq '.object.sha')
TAG_TYPE=$(gh api repos/OWNER/REPO/git/ref/tags/TAG \
  --jq '.object.type')

if [ "$TAG_TYPE" = tag ]; then
  RELEASE_SHA=$(gh api repos/OWNER/REPO/git/tags/$TAG_OBJ \
    --jq '.object.sha')
else
  RELEASE_SHA=$TAG_OBJ
fi

gh api repos/OWNER/REPO/compare/FIX_SHA...$RELEASE_SHA \
  --jq '{status,ahead_by,behind_by,total_commits,html_url}'
```

Interpretation for `FIX_SHA...RELEASE_SHA`:

- `status: ahead`, `behind_by: 0`: the release is ahead of and contains the fix.
- `status: behind`: the release predates the fix.
- `diverged`: ancestry is not established; inspect branches/cherry-picks.

Check the release immediately before and after the fix to identify the first containing version. Use the compare URL as evidence.

A packaged/local commit may not exist upstream. If GitHub says no commit exists, map the installed package to its official release tag and use that upstream release commit for containment.

## 6. Check Discussions

```bash
gh api graphql -f query='query {
  repository(owner:"OWNER", name:"REPO") {
    hasDiscussionsEnabled
    discussions(first:100, orderBy:{field:UPDATED_AT,direction:DESC}) {
      nodes { number title url body closedAt }
    }
  }
}'
```

If `hasDiscussionsEnabled` is false, state that there were no repository discussions to search.

## 7. Classify candidates precisely

Compare each candidate against a small signature table:

| Dimension | Observed | Candidate |
|---|---|---|
| Direction | inbound / outbound / prompt content | |
| Cardinality | one event, two callbacks; one callback, two sends | |
| Timing | immediate / delayed replay | |
| Identity | same or different event timestamp / session key | |
| Busy behavior | ack only / actual cancellation / parallel run | |
| Version | contains prior fix? | |

This prevents conflating:

- transport redelivery with generic double-send;
- duplicate rich-text content with duplicate agent turns;
- per-thread concurrency with same-session re-entry;
- acknowledgement suppression with root-cause correction.

## 8. Recommendation rule

Recommend **commenting** when an existing open issue has the same signature and the new evidence narrows or reproduces it. Recommend **a new issue** when the closest issue is closed as fixed, the reporter reproduces on a containing release, or the direction/timing/session identity differs materially. Reference the related closed issues and commits in the new report.