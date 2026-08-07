# WF4 probes 4-8 — credential re-runs from the Tars VM

Agent: WF4 credential probe subagent. All commands run over `ssh gaetan@192.168.0.9`.
No `sops` invoked anywhere; all secrets sourced from `~/.hermes/.env` on the VM via
`set -a; . ~/.hermes/.env; set +a` and passed to `curl` through `-K` config files on
`/tmp` (written, used, then `shred -u`'d), never on argv, never echoed.

VM sanity (2026-08-07T19:00:10+00:00): `hermes --version` → Hermes Agent v0.20.0
(2026.8.3); `himalaya --version` → v2.0.0 (+imap +smtp +gmail...).

---

## §4 — Gmail read via Himalaya

**Deviation from spec:** spec text says `-p INBOX`; confirmed via
`himalaya envelope list --help` that v2 syntax is `-m, --mailbox <NAME>` (no `-p`
mailbox flag exists — `-p` is page number). Ran the corrected form.

```
$ date -Is
2026-08-07T19:00:18+00:00
$ himalaya account list
┌──────┬────────────┬─────────┐
│ NAME ┆ BACKENDS   ┆ DEFAULT │
╞══════╪════════════╪═════════╡
│ tars ┆ imap, smtp ┆ yes     │
└──────┴────────────┴─────────┘
$ himalaya envelope list -a tars -m INBOX -s 1
┌───────┬───────┬────────────────┬────────┬────────────────────────┬───────────┐
│ ID    ┆ FLAGS ┆ SUBJECT        ┆ FROM   ┆ DATE                   ┆ SIZE      │
╞═══════╪═══════╪════════════════╪════════╪════════════════════════╪═══════════╡
│ 12204 ┆ *     ┆ Security alert ┆ Google ┆ 2026-08-07 17:17+00:00 ┆ 12.03 KiB │
└───────┴───────┴────────────────┴────────┴────────────────────────┴───────────┘
$ EXIT=0
```

**Verdict: PASS.** One envelope returned, exit 0. Spec deviation noted (`-m` not `-p`) —
no owner action needed, CLI syntax correction only, consistent with the WF3-era override
in the task brief.

---

## §5 — Calendar read (CalDAV)

Credential: `GMAIL_ADDRESS` / `GMAIL_APP_PASSWORD` sourced from `~/.hermes/.env`, passed
via `-K` stdin config (`user = "addr:apppass"` — curl config-file syntax, not argv).

First attempt against `www.google.com/calendar/dav/...` using a two-line `-K` file
(`--user\n<value>`) mis-parsed by curl's config-file grammar (400/401 loginRequired) —
corrected to the documented `user = "value"` config syntax; re-ran, same URL, real
result below (temp file shredded after use):

```
$ date -Is
2026-08-07T19:00:36+00:00
$ curl -sS -o resp.xml -w "HTTP=%{http_code}\n" -K p5b.cfg -X PROPFIND -H "Depth: 0" \
    "https://www.google.com/calendar/dav/${GMAIL_ADDRESS}/events"
HTTP=207
$ head -c 300 resp.xml
<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:caldav="urn:ietf:params:xml:ns:caldav" ...>
 <D:response ...
```

**Verdict: PASS.** 207 Multi-Status, never a bare 200. Confirms §4's app-password
credential doubles for CalDAV cleanly (per spec, no separate credential).

---

## §6 — Linear query

Credential: `LINEAR_API_KEY`, sent as bare key (no `Bearer` prefix) via `-K` header
config, GraphQL `{ viewer { id name } }` POST to `api.linear.app/graphql`.

```
$ date -Is
2026-08-07T19:00:41+00:00
$ curl -sS -o p6.json -w "HTTP=%{http_code}\n" -K p6.cfg -X POST \
    -d '{"query":"{ viewer { id name } }"}' https://api.linear.app/graphql
HTTP=200
$ cat p6.json
{"data": {"viewer": {"id": "4951b192-e49c-4b7e-b491-58c89e66043c", "name": "Gaëtan Cathelain"}}}
```

**Verdict: PASS.** 200, `viewer.id` + `viewer.name` present and match Gaetan's account.

---

## §7 — Notion query (spec deviation applied per task brief)

Spec §7 calls for replaying the harvested Quick-Find `search` XHR. Per WF3 finding
(`status/probes/notion.md`), `/api/v3/search` 400s regardless of token validity — not a
useful liveness signal. Substituted the corrected auth-shaped probe per this task's
instructions: `POST /api/v3/loadUserContent` with cookie `token_v2=$NOTION_TOKEN_V2`,
empty JSON body. Bare 200 = live token, 401 = dead token.

```
$ date -Is
2026-08-07T19:00:45+00:00
$ curl -sS -o p7.json -w "HTTP=%{http_code}\n" -K p7.cfg -X POST -d "{}" \
    https://www.notion.so/api/v3/loadUserContent
HTTP=200
$ python3 -c "import json;print(list(json.load(open('p7.json')).keys()))"
['recordMap']
```

**Verdict: PASS** (bare 200, `recordMap` key present — token_v2 alive). **Spec
deviation logged**: used `loadUserContent` in place of the spec's `search` XHR replay,
per WF3's prior proof that `search` 400s unconditionally. `NOTION_FILE_TOKEN` (export
path) was **not exercised** — noted only, not gated on, per spec §7's own instruction.

---

## §8 — GitHub query + metarepo clone freshness

**Spec deviation #1:** spec mentions a PAT header for the GitHub API call. VM auth is
device-flow `gh` (already authenticated), not a bare PAT header file — used `gh api`
directly, which is the live auth path on this VM.

**Spec deviation #2:** spec says clone lives at `~/mc-metarepo`; confirmed the real path
is `~/dev/mc-metarepo` (WF3 cloned there). `~/mc-metarepo` does not exist.

```
$ date -Is
2026-08-07T19:00:48+00:00
$ gh auth status
github.com
  ✓ Logged in to github.com account GaetanCathelain (/home/gaetan/.config/gh/hosts.yml)
  - Token scopes: 'gist', 'read:org', 'repo'
$ gh api repos/mobile-club/metarepo --jq .private
true
$ ls -d ~/mc-metarepo
ls: cannot access '/home/gaetan/mc-metarepo': No such file or directory
$ ls -d ~/dev/mc-metarepo
/home/gaetan/dev/mc-metarepo

$ date -Is
2026-08-07T19:00:53+00:00
$ cd ~/dev/mc-metarepo && git fetch
exit=0
$ git status -sb
## main...origin/main
$ git log -1 --format="local HEAD  %H %cI"
local HEAD  4831253a24928e57b76fcb2870df95705706b6f2 2026-08-07T18:17:48+02:00
$ git log -1 --format="origin/main %H %cI" origin/main
origin/main 4831253a24928e57b76fcb2870df95705706b6f2 2026-08-07T18:17:48+02:00
$ git rev-list --left-right --count HEAD...origin/main
0	0
```

**Verdict: PASS.** `.private == true`; `git fetch` clean, no auth error; local HEAD ==
origin/main exactly (0 commits ahead/behind, well within the "one commit" tolerance).
Both spec deviations (gh device-flow vs stale PAT-header instruction; `~/dev/mc-metarepo`
vs stale `~/mc-metarepo` path) logged, no owner action needed — documentation corrections
only.

---

## Summary table

| Probe | Result | Notes |
|---|---|---|
| §4 Gmail (Himalaya) | PASS | `-m` not `-p` (v2 syntax correction) |
| §5 Calendar (CalDAV) | PASS | 207 Multi-Status; `-K` config-file syntax correction needed (`user = "..."`) |
| §6 Linear | PASS | bare key, `viewer.id`+`viewer.name` returned |
| §7 Notion | PASS | corrected probe (`loadUserContent`, not `search`); spec deviation logged; `NOTION_FILE_TOKEN` not exercised |
| §8 GitHub + metarepo | PASS | `gh` device-flow (not PAT header); path is `~/dev/mc-metarepo` (not `~/mc-metarepo`); HEAD == origin/main |

No FAILs — no owner-bucket routing needed for any of the five assigned probes.
