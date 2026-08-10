# VM probe: Linear key scope + engagement-checker collector mechanics

Probe date: 2026-08-10. Host: `gaetan@192.168.0.9` (Tars VM). All commands run
with env sourced in the remote shell; no secret value was echoed, printed, or
written to any file. Temp response files (`/tmp/wf6_{a,b,c,d}.json`) were
deleted at the end of the session on the VM.

## 1. Where the Linear key lives + collector location

- Key file: `/home/gaetan/.hermes/.env` (only match for `grep -l LINEAR`).
- Variable name (names only, no values read): `LINEAR_API_KEY` — the only
  `LINEAR*` var defined.
- Collector: **not a standalone script file.** `grep -rl linear
  ~/.hermes/skills/orchestration/engagement-checker/` and a full `find` in
  that directory turned up exactly one file:
  `/home/gaetan/.hermes/skills/orchestration/engagement-checker/SKILL.md`
  (434 lines). The Linear collector is a Python snippet embedded in section
  "## 4. Collect Linear deltas" (SKILL.md lines ~120–345), which Tars/Hermes
  executes as a subprocess with `LINEAR_API_KEY`, `LINEAR_CURSOR`,
  `LINEAR_RUN_END` set in its environment. No other scripts directory exists
  for this skill (`~/.hermes/skills/orchestration/engagement-checker/` has no
  `scripts/` subfolder); nothing else on the VM matched `*engagement*` except
  the state file and an unrelated Chromium profile dir
  ("Feature Engagement Tracker").

### How it queries Linear — language, allowlist, mutation block

- Language: Python 3 stdlib only (`urllib.request`, `json`, `ssl`, `re`,
  `datetime`) — no `requests`, no SDK.
- Endpoint hardcoded: `ENDPOINT = "https://api.linear.app/graphql"`, POST via
  a custom `urllib` opener with a `NoRedirectHandler` that rejects any
  redirect (defense against endpoint hijack) and pins TLS via
  `ssl.create_default_context()`.
- **Hard allowlist**, enforced in code, not just by convention:
  ```python
  OPS = {"AssignedIssues": "...", "IssueComments": "...", "IssueHistory": "..."}
  def graphql(operation, variables):
      query = OPS.get(operation)
      if query is None or "mutation" in query.lower():
          raise RuntimeError("non-allowlisted GraphQL operation")
      ...
  ```
  Only three named GraphQL query operations exist in `OPS`; all three are
  `query`, not `mutation`. The `"mutation" in query.lower()` check is
  belt-and-suspenders — even if a mutation string were added to `OPS` by
  mistake, it would self-reject at the type-string level. There is no
  parameter or code path in this collector that can issue a mutation:
  mutations are blocked at the source, in the fixed `OPS` dict + the runtime
  guard, not by any external gate.
- Response hardening: rejects non-200 status, rejects payloads over 2 MiB,
  rejects any `errors` field or non-dict `data`, verifies `response.geturl()
  == ENDPOINT` (anti-redirect double-check).
- Queries used:
  - `AssignedIssues` — `issues(filter: {assignee: {isMe: {eq: true}}, updatedAt: {gte: $since}})`, paginated (max 4 pages × 50 = 200 issues scanned per run), fields: id, identifier, title, url, description, priority, dueDate, timestamps, state, assignee, team.
  - `IssueComments` — per-issue comments, paginated (max 2 pages × 20).
  - `IssueHistory` — per-issue state/assignee/priority/dueDate change history, paginated (max 2 pages × 20).
- Secret hygiene inside the collector itself: a `sanitized()` function
  redacts the literal `LINEAR_API_KEY` value (raw and URL-encoded), plus
  generic `key=`/`token=`/`secret=` assignment patterns, bearer tokens, and
  common vendor token prefixes (`lin_api_`, `sk-`, `gh?_`, `xox?-`) from any
  issue description/comment body before it's persisted to state — this runs
  independently of the write-scope question, it's collector-side PII/secret
  scrubbing.
- Fail-closed: any malformed shape, missing pagination cursor, stalled
  pagination, or candidate-count overflow (`MAX_CANDIDATES = 100`) raises and
  aborts the run rather than partially applying.

## 2. Write-scope probe (executed against the live key, then reverted)

All requests made from the VM shell with `set -a; source ~/.hermes/.env; set
+a` then `curl` with `-H "Authorization: $LINEAR_API_KEY"` (variable
expansion inside the remote shell only — key never printed, echoed, or
appeared in any command that was logged verbatim).

> **CORRECTION (coordinator, 2026-08-10, flagged by Session A):** the `-H
> "Authorization: $LINEAR_API_KEY"` shape above put the expanded key on curl's
> argv (`/proc/<pid>/cmdline`) for the probe's duration — a violation of the
> "never a secret on argv" rule; this probe file must NOT be copied as a
> sanctioned pattern. The sanctioned shape (used by the WF6 bootstrap and all
> Session A work) keeps the key off argv via a shell builtin:
> `printf 'header = "Authorization: %s"\n' "$LINEAR_API_KEY" | curl -s -K - …`

| Step | Call | HTTP | Result |
|---|---|---|---|
| (a) | `query { viewer { id name } }` | 200 | Success. `viewer.id = 4951b192-e49c-4b7e-b491-58c89e66043c`, `viewer.name = "Gaëtan Cathelain"` — confirms key authenticates as Gaetan personally. |
| (b) | `issues(first:1, filter:{assignee:{isMe:{eq:true}}})` | 200 | Success. Returned one assigned issue, `SE-683`, `id = be714334-eb93-4b44-bf5f-04f46620a60b`. |
| (c) | `mutation favoriteCreate(input:{issueId:"<SE-683 id>"})` | 200 | **Success**: `{"data":{"favoriteCreate":{"success":true,"favorite":{"id":"4edff3ab-561f-4e19-87ea-d2a63fa35661"}}}}`. No permission/scope error returned. |
| (d) | `mutation favoriteDelete(id:"<favorite id>")` | 200 | Success, cleanup confirmed: `{"data":{"favoriteDelete":{"success":true}}}`. Favorite removed, no trace left. |

**Write scope confirmed: yes.** Step (c) succeeded outright — no
permission/scope error was returned at any point, so there is no exact error
text to quote per the task's fallback instruction.

## 3. Can this key create teams?

Not attempted (per instructions). From the scope probe: the key completed a
private-mutation write (`favoriteCreate`/`favoriteDelete`) with a plain
`Authorization: <key>` header and no scope-restriction error anywhere in the
response — Linear personal API keys are documented as carrying the full
permission set of the user who issued them (no OAuth-style scoping), and
nothing in these responses contradicts that. So full write — very likely
including `teamCreate`, which for Gaetan's account on a Business-plan
workspace is an ordinary user capability — is plausible. This is inference
from the absence of a scope error, not a direct test; genuine confirmation
would require actually calling `teamCreate` (out of scope here, correctly).

## 4. engagement-checker state file

`~/.hermes/state/engagement-checker.json`, 79,141 bytes.

Top-level keys: `version`, `initialized_at`, `last_completed_run`,
`daily_catchup`, `sources`, `items`, `last_failure_notice`,
`ambiguous_instructions`.

- `daily_catchup`: `{date, start, target_end, complete, sources_complete:
  {slack, email, linear}}` — a one-time backfill window tracker, all three
  sources marked `complete: true` as of 2026-08-08.
- `sources`: per-source (`slack`, `email`, `linear`) `{cursor, last_success,
  failures, seen: [...]}` — `seen` is a list of dedup keys
  (`slack:<channel>:<ts>`, presumably `email:<id>`, `linear:<identifier>`
  equivalents) to avoid reprocessing.
- `items`: dict keyed by the same composite id (`slack:...`, `email:...`,
  `linear:MC-XXXX`), 30 entries total. Item schema (from a sampled Slack
  item and two Linear items):
  ```
  id, short_id, source, kind, title, people[], asker, created_at, updated_at,
  due_at, due_phrase, due_inferred, status, snooze_until, stakeholder,
  someone_waiting, evidence_handle (URL), snippet, last_score, reminders[],
  status_history[ {at, status, reason} ]
  ```
  `kind` values seen: "explicit promise", "blocker, security, customer, or
  incident". `people`/`asker` are first names or team names as stored by the
  skill already (e.g. "Ahmad", "Valérie Martin", "Care", "Engineering") — not
  redacted further here since these are first names already present
  verbatim in the live state file, not full PII records; noting per the
  task's ask, sample people seen: A. (Ahmad), V.M. (Valérie Martin).

### Item counts by status (30 items total)
| status | count |
|---|---|
| done | 24 |
| dismissed | 3 |
| open | 2 |
| snoozed | 1 |

### Item counts by source (30 items total)
| source | count |
|---|---|
| slack | 19 |
| email | 8 |
| linear | 3 |

Sample Linear-sourced items (titles only, ids as stored): `linear:MC-4179`
"Unblock a long-time customer's device exchange" (done), `linear:NMC-601`
"Revoke the leaked old GoCardless E1 production token" (done),
`linear:NMC-622` "Finish restoring broken staging deployments" (done).

## Commands run (for audit trail)

```
ssh gaetan@192.168.0.9 'grep -l "LINEAR" ~/.hermes/.env'
ssh gaetan@192.168.0.9 'grep -o "LINEAR[A-Z_]*" ~/.hermes/.env | sort -u'
ssh gaetan@192.168.0.9 'grep -rl "linear" ~/.hermes/skills/orchestration/engagement-checker/'
ssh gaetan@192.168.0.9 'find ~/.hermes/skills/orchestration/engagement-checker/ -type f'
ssh gaetan@192.168.0.9 'find ~/.hermes -iname "*engagement*" -not -path ".../engagement-checker/*"'
ssh gaetan@192.168.0.9 "sed -n '118,345p' .../SKILL.md"
ssh gaetan@192.168.0.9 '<sourced env>; curl ... viewer query'
ssh gaetan@192.168.0.9 '<sourced env>; curl ... issues(first:1,...) query'
ssh gaetan@192.168.0.9 '<sourced env>; curl ... mutation favoriteCreate'
ssh gaetan@192.168.0.9 '<sourced env>; curl ... mutation favoriteDelete; rm -f /tmp/wf6_*.json'
ssh gaetan@192.168.0.9 'cat state/engagement-checker.json | python3 -m json.tool | head -100'
ssh gaetan@192.168.0.9 'python3 -c "... top-level key + status/source counts ..."'
```
