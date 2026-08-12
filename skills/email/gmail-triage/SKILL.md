---
name: gmail-triage
description: "Use to label new inbox mail against Gaetan's 8-label taxonomy."
version: 1.0.0
metadata:
  hermes:
    tags: [email, gmail, triage, labels, automation, cron]
    category: email
---

# Gmail triage

An hourly, incremental auto-labeller for `gaetan.cathelain@mobile.club`. Each
run takes the inbox messages that arrived since the cursor and carry no user
label, routes each one to one of eight business labels, and applies that label.
It is not an inbox assistant: it never reads to Gaetan, never answers, never
tidies. Labelling is the whole job, and most runs return exactly `[SILENT]`.

Tars owns and executes this workflow directly. Do not spawn, prompt, resume, or
delegate through Claude, a worktree/session, or another agent system. Outside
Tars's own state file the permitted write is **exactly the one in "Permitted
writes" below** — that section is the authority, and nothing else in this file
or in the cron prompt restates or extends it.

All human times use `Europe/Paris`. Provenance: GCN-44 / WF7
(`docs/specs/wf7-gmail-triage.md`).

**Empty is a value, not a failure.** A search returning zero messages, a run
labelling nothing, an absent `seen` ring — all are legitimate readings of calls
that worked. Only three things are failures: a non-zero exit from
`google_api.py`, a call that could not be made at all, and a candidate list that
hit the per-run cap. Nothing else may hold the cursor or break silence.

## Permitted writes

One, and no others:

1. **Add one or two of the eight taxonomy label IDs to a Gmail message** —
   `google_api.py gmail modify <message_id> --add-labels <ID[,ID]>`.

Plus the two Tars-local writes that are not mailbox writes: this skill's state
file, and the Slack report of §5 when §5 says to send one.

## Safety rails — absolute, structural, never relaxed

These are not guidance. A run that cannot honour one of them stops instead.

- **`--remove-labels` is never passed. Not once, not with `UNREAD`, not with
  `INBOX`, not to fix a mistake.** That single ban is what makes "never
  archive, never mark read/unread, never un-label" true structurally rather
  than by intention — every one of those actions is a `--remove-labels` call.
  A wrong label is corrected by Gaetan in the UI, or by an explicitly
  instructed one-off, never by a cron tick.
- **`--add-labels` carries only IDs from the eight-row table in §3.** No
  system label (`TRASH`, `SPAM`, `STARRED`, `IMPORTANT`, `UNREAD`), no
  `Hermes/*`, no id resolved at runtime from a name that is not in that table.
  Adding `TRASH` is deletion; the ID allowlist is what makes deletion
  unreachable.
- **`github`, `tech-mailgroup`, `staging-recouvrement` and every `Hermes/*`
  label are untouchable** — never added, never removed, never counted, never
  proposed for cleanup by a cron run. `github` and `tech-mailgroup` are
  Gaetan's own skip-inbox filters and work correctly;
  `staging-recouvrement` is a 17868-message fossil whose disposal is Gaetan's
  call (§ Open items in the spec). They are outside the ID allowlist above, so
  this is already structurally true; it is restated because an agent that
  "helpfully tidies" would reach for them first.
- **Never `gmail send`, `gmail reply`, or any Gmail write other than the one
  permitted `modify` shape.**
- **At most 100 `modify` calls per run.** Overflow is normal, not an error:
  record it in state, report it, and drain the rest next run (§4).
- **Ambiguous means unlabelled.** An item the rules miss and about which the
  evidence in hand does not support a confident single call is left bare. An
  unlabelled mail costs Gaetan nothing; a wrongly-labelled one costs him trust
  in every other label. Never spread a label across a class to "cover" it.
- **Threads are not the unit — messages are.** `gmail modify` takes one
  `message_id` and Gmail aggregates labels up to the thread in the UI. Never
  claim thread-level coverage.

## Load only what is needed

- `google-workspace` supplies the transport. Everything below runs through its
  `scripts/google_api.py`, invoked with the **hermes-agent venv interpreter** —
  bare `python3` fails with `ModuleNotFoundError: googleapiclient`:

  ```bash
  GAPI="${HERMES_HOME:-$HOME/.hermes}/hermes-agent/venv/bin/python ${HERMES_HOME:-$HOME/.hermes}/skills/productivity/google-workspace/scripts/google_api.py"
  ```

- Do **not** load `himalaya`. It is the engagement-checker's read fallback and
  has no place here: its IMAP `message move/copy` is folder semantics, not
  Gmail's multi-label semantics, and this skill must never move a message out
  of the inbox. If `google_api.py` is unavailable, this skill does nothing and
  reports it (§6) — there is no fallback labelling path.
- Do not load `engagement-checker` or `daily-work-brief`. They read the same
  mailbox for a different purpose, hold their own cursors, and share no state
  with this skill.

## Durable state

The source of truth for the run is `~/.hermes/state/gmail-triage.json`. Create
the parent directory if needed. Read it at the start; overwrite it after the
labelling loop, before delivery, so a delivery failure cannot replay a run.

```json
{
  "version": 1,
  "cursor": "ISO-8601 timestamp with offset",
  "last_completed_run": "ISO timestamp or null",
  "seen": [],
  "labels": {
    "🛠️ Outils & SaaS": "Label_6",
    "💳 Finance & Facturation": "Label_7",
    "👥 Interne Mobile Club": "Label_8",
    "🤝 Partenaires & Prestataires": "Label_9",
    "📅 Réunions & Agenda": "Label_10",
    "🎯 Recrutement": "Label_11",
    "🔐 Sécurité & Comptes": "Label_12",
    "📣 Newsletters & Pub": "Label_13"
  },
  "last_run": {
    "at": "ISO timestamp",
    "candidates": 0,
    "labelled": 0,
    "unrouted": 0,
    "overflow": 0,
    "by_label": {}
  },
  "consecutive_failures": 0,
  "last_failure_notice": {},
  "unrouted_senders": {}
}
```

- `seen`: stable Gmail **message ids** already processed, most recent last, at
  most 500, oldest evicted. It is what makes the day-granular retrieval overlap
  in §1 idempotent, and nothing else writes it.
- `unrouted_senders`: `{"<sender domain>": {"count": N, "last_seen": "ISO",
  "sample_subject": "≤120 chars"}}`, at most 20 domains, lowest-count evicted
  first. It is the evidence behind the "worth a rule?" bullet in §5 — a domain
  seen once is noise, one seen five times across runs is a missing rule.
- `last_failure_notice`: `{"<condition>": "ISO timestamp"}`, the rate-limit key
  of §6. Written only where §6 says.
- Never persist message bodies, subjects beyond the bounded sample above,
  addresses beyond the domain, tokens, or headers. State writes preserve
  unknown future top-level keys.

**Self-healing.** If the file is missing or unparseable, recreate it from the
shape above with `labels` exactly as written, `cursor` = now, `seen` = `[]`,
`consecutive_failures` = 0, and report one line under §5. Do **not** backfill
after a state loss: a fresh cursor means the gap is unlabelled mail, which is
the pre-skill status quo and harms nothing, whereas a self-authorised backfill
is up to 2843 unreviewed label writes.

## Single-writer lock

Atomically create `~/.hermes/state/gmail-triage.lock` as a directory. If it
exists and is less than 15 minutes old, return exactly `[SILENT]`. If it is
older, treat it as a crashed run, remove only that empty lock directory, and
retry once. Always remove it in final cleanup. Do not remove any other path.

## 1. Establish the incremental window

Get the current timestamp with a tool, convert to Europe/Paris, and fix it as
the run's end timestamp. Read `cursor` from state.

Gmail search is **date-granular only** — `after:` takes `YYYY/MM/DD`, not a
time. So: widen the query to the cursor's Paris-local date, then filter the
returned messages locally by their exact `date` field, keeping only those
strictly newer than the cursor **or** whose id is absent from `seen`. That
widening is transport overlap, not permission to reprocess: a message both
older than the cursor and present in `seen` is dropped without a second look.

Advance the cursor only where §4 says, and never past a message this run did
not finish.

## 2. Collect candidates

One search, bounded:

```bash
$GAPI gmail search "in:inbox -has:userlabels after:YYYY/MM/DD" --max 100
```

`search` returns a JSON array of `{id, from, subject, date, snippet}`. Those
four fields plus the sender's domain are the classification input for the vast
majority of messages.

- `in:inbox` excludes `github` and `tech-mailgroup`, which bypass the inbox by
  Gmail filter, so this skill never sees them.
- `-has:userlabels` excludes anything already carrying a user label —
  including messages a previous run labelled and the `Hermes/Claude-Invoice-*`
  accounting labels, which stay this skill's business never. **If that operator
  is rejected or the result set is visibly not filtered by it, drop it from the
  query and rely on `cursor` + `seen` alone**, which are sufficient; say so
  once under §5 and continue. Never treat it as a failure.
- `gmail get <message_id>` exists and returns the full body. Use it **only**
  for a message the §3 rules miss *and* whose sender, subject and snippet are
  genuinely insufficient, capped at **5 `get` calls per run**. Past that cap,
  leave the remainder unlabelled — reading forty bodies to place forty
  newsletters is not worth the turn.

Append every processed message id to `seen` as it is decided, capped at 500.

## 3. Classify — deterministic rules first, judgment second

Evaluate the table **top to bottom and stop at the first match**. The order is
load-bearing: the specific subject rules (🔐, 💳) must be tested before the
broad domain rule (🛠️) or 🛠️ swallows them by domain, and the
address-specific rules (📣, 🤝) before it for the same reason —
`info@e.atlassian.com` is marketing while `atlassian.net` is tooling, and
`ramzi@cloudflare.com` is an account manager while `em@em1.cloudflare.com` is a
drip campaign.

| # | Label | ID | Matches when |
|---|---|---|---|
| 1 | 🔐 Sécurité & Comptes | `Label_12` | subject contains `security alert`, `alerte de sécurité`, `suspicious login`, `connexion suspecte`, `API key`, `API token`, `verifying it's you`, `2FA`, `password reset`; **or** sender is `google-workspace-alerts-noreply@google.com`, `no-reply@accounts.google.com`, `security@updates.linear.app` |
| 2 | 💳 Finance & Facturation | `Label_7` | subject contains `receipt`, `invoice`, `facture`, `spend limit`, `on-demand spend`, `cost report`, `billing`, `payment`, `abonnement`; **or** sender is `invoice+statements@mail.anthropic.com`, `accounts@1password.eu`, `no-reply@sns.amazonaws.com`, `notifications@doit.com` |
| 3 | 📅 Réunions & Agenda | `Label_10` | subject starts with `Invitation:`, `Updated invitation:`, `Invitation annulée:`, `Canceled event:`, `Accepted:`, `Declined:`; **or** subject contains `meeting recap`, `notes de réunion`, `transcript`; **or** sender is `meetings-noreply@google.com`, `calendar-notification@google.com`, or a Fireflies/Granola address **and** the subject names a meeting rather than onboarding |
| 4 | 🎯 Recrutement | `Label_11` | subject or snippet contains `candidature`, `spontaneous application`, `alternance`, `stage`, `CV`, `entretien`, `recrutement`, `application for` — and the sender is **not** `@mobile.club` |
| 5 | 👥 Interne Mobile Club | `Label_8` | sender domain is `mobile.club` or `cleaq.com`, excluding `tech@mobile.club` |
| 6 | 🤝 Partenaires & Prestataires | `Label_9` | sender domain is `nextmobiles.com`, `techtribe.fr`, `egym-wellpass.com`; **or** sender is `ramzi@cloudflare.com`; **or** a DoIT address whose subject is relationship traffic rather than the monthly cost digest (the digest is row 2) |
| 7 | 📣 Newsletters & Pub | `Label_13` | sender is `learn@send.zapier.com`, `updates@dynatrace.ai`, `info@e.atlassian.com`, `em@em1.cloudflare.com`, `changelog@neon.tech`, `news@gifteo.fr`, `no-reply@email.claude.com`, `noreply@tm.openai.com`; **or** subject is a plain product-announcement/changelog/webinar/promo with no transaction and no account event |
| 8 | 🛠️ Outils & SaaS | `Label_6` | sender domain is one of `mail.anthropic.com`, `anthropic.com`, `vercel.com`, `linear.app`, `github.com`, `dtdg.eu`, `datadoghq.com`, `1password.com`, `1password.eu`, `mail.notion.so`, `atlassian.com`, `atlassian.net`, `id.atlassian.com`, `cloudflare.com`, `openai.com`, `tm.openai.com`, `fivetran.com`, `zapier.com`, `neon.tech`, `granola.ai`, `fireflies.ai` — **and** the subject matches no row-1 and no row-2 cue. Those two negative filters are not redundant with the ordering: keep both, they are what stops a domain rule from eating a receipt or a security alert if a row above is ever edited |

Matching is case-insensitive, on the sender address and its domain, the
subject, and — for row 4 only — the snippet.

**When no row matches**, judge it. Assign a label only when one of the eight is
clearly right on the evidence in hand; otherwise leave it unlabelled and count
it in `unrouted`, recording the sender's **domain** (never the full address) in
`unrouted_senders`. Judgment fills the gaps the rules cannot see — a new SaaS
vendor, a partner writing from a personal address, a French subject the cue
list misses. It never overrides a row that matched.

**Multi-label** is permitted where a message is genuinely both, and only then:
pass the two IDs comma-separated in a single `--add-labels`. The canonical case
is the DoIT monthly cost digest, which is `Label_7,Label_9`. Prefer one label.
Never three.

**`nextmobiles.com` routes to 🤝 Partenaires & Prestataires provisionally** —
its sister-brand-versus-third-party status is unresolved (spec § Open items).
Apply the rule; do not re-litigate it in a run; do not raise it in a report.

## 4. Apply labels

Order candidates by their `date` ascending — oldest first, so a batch drains
from the cursor forward — and label them in that order.

```bash
$GAPI gmail modify <message_id> --add-labels <ID>
```

- One message per call: `google_api.py` has no batch flag and no
  `batchModify`. Up to **10 such calls may be chained with `&&`** in one
  terminal command to cut turn count; on a non-zero exit re-run the remainder
  individually so one failure does not silently drop nine messages.
- **Run these through the shell/terminal tool.** Never `execute_code`, never
  `python3 -c`, never a heredoc, never a pipe into an interpreter: a scheduled
  run has no one to answer an approval prompt, so `approvals.cron_mode: deny`
  turns a flagged shape into a blocked tool error that never reaches the
  report — the labelling simply does not happen and the run goes `[SILENT]`
  looking healthy. A bare interpreter path followed by a script path, with
  `&&` chaining, is the safe shape. **This skill needs no script of its own**;
  if a future version does, materialise it with `write_file` and execute it by
  absolute path, exactly as `engagement-checker` §4 does, and leave the file
  in place rather than deleting it.
- **Stop at 100 `modify` calls.** Record `overflow` = the number of remaining
  candidates, set the cursor to the `date` of the **last message actually
  labelled**, and report the overflow under §5. The retained batch is
  contiguous from the cursor, so nothing is skipped and the backlog drains one
  bounded batch per hour.
- With no overflow, advance the cursor to the run's fixed end timestamp.
- A message that was classified but whose `modify` failed is **not** counted as
  labelled, and the cursor does not advance past it. Its id still enters
  `seen` only after a successful call.

Write state — cursor, `seen`, `last_run`, `unrouted_senders`,
`consecutive_failures` = 0 — before delivering.

## 5. Deliver or remain silent

Most runs label a handful of routine notifications and are of no interest.
Persist state, release the lock, and return exactly:

`[SILENT]`

Break silence only for one of these four, and then send **at most six
bullets**, no preamble, no per-message list, no "no action needed" filler:

1. **🔐 seen** — one or more messages routed to `Label_12` this run.
2. **A new sender class worth a rule** — a domain in `unrouted_senders` whose
   count has reached 3 or more.
3. **An error or scope failure** (§6).
4. **Overflow** — more than 100 candidates in the window.

One line each:

```
• 🔐 2 security items labelled — Google login alert, Linear API key created.
• Unrouted: acme-corp.com ×4 over 3 runs (subj. "Contract renewal") — rule worth adding?
• Overflow: 137 candidates, 100 labelled, 37 carry to the next run.
• Error: gmail modify exited 1 on 3 messages — <first error line, ≤120 chars>.
```

Counts and label names only. **Never quote a message body, a full subject line
of a 🔐 item, or a sender's full address** — the report travels to Slack, the
mailbox does not.

Three states must never be hidden and are reported on any run that delivers,
whatever the rate limit: a `modify` that failed, an authentication/scope
failure, and an overflow. Say plainly what was not done; never let a run read
as if the labelling had worked.

## 6. Failure handling

- **`google_api.py` exits non-zero on auth or scope** (`invalid_grant`,
  `insufficient authentication scopes`, a 401/403): **stop the run
  immediately.** Do not retry, do not fall back to himalaya, do not re-run the
  search. Increment `consecutive_failures`, persist state, release the lock,
  and send one bullet naming the error. Retry is next hour's tick and nothing
  sooner — a retry storm against Google's auth endpoint is how a token gets
  rate-limited into a longer outage.
- **A single `modify` fails** (transient, 5xx, one bad id): skip that message,
  leave it unlabelled and out of `seen`, continue the batch, count it, hold the
  cursor at the last success. It retries next run for free.
- **Rate limit on the report, one exemption.** A persisting failure is
  reported at most once every four hours per distinct condition, keyed on
  `last_failure_notice[<condition>]`. The three must-never-hide states of §5
  are exempt: they are reported on every delivering run. A rate limit that can
  suppress "the labelling did not happen" makes a broken run read as healthy.
- `consecutive_failures` resets to 0 on any run that completes its labelling
  loop, including one that labelled nothing.
- **Never fabricate.** If the run cannot say what it did, it says that.

## Scheduling contract

Hermes' configured timezone is `Europe/Paris` and a cron job carries no
timezone field of its own — the expression is interpreted in that timezone.
One job, no second pass:

- `0 7-21 * * *` — hourly on the hour, 07:00 through 21:00 Paris, every day.

```bash
hermes cron create '0 7-21 * * *' \
  --name "Gmail triage" \
  --skill gmail-triage \
  --deliver slack \
  '<the prompt below>'
```

`--deliver slack` (bare platform name) delivers to `SLACK_HOME_CHANNEL`,
Gaetan's DM. Verify with `hermes cron list` after creating.

The job's self-contained prompt:

> Run `gmail-triage` end to end. Use Europe/Paris and the run's fixed
> timestamp. Acquire the single-writer lock; read durable state; search the
> inbox for messages newer than the cursor carrying no user label; classify
> each against the eight-label table, deterministic rows first and judgment
> only for what they miss; apply labels with `google_api.py gmail modify
> --add-labels` and **never** `--remove-labels`; cap the run at 100 modify
> calls; persist state; release the lock; then return the report or exactly
> `[SILENT]`. Labelling is the only mailbox write permitted — never send,
> reply, archive, delete, or mark read. Break silence only for a security-
> labelled item, a repeat unroutable sender, an error, or an overflow. If a
> label could not be applied or authentication failed, say so rather than
> going silent.

## Backfill — separately authorised, never on a tick

The cursor starts at deployment time, so the pre-existing inbox is untouched by
design. Labelling that history is a one-off of up to ~2800 writes and needs
**Gaetan's explicit go, per run, naming the window**. When authorised: same
classification table, same 100-per-batch cap, oldest first, `in:inbox
-has:userlabels before:<deployment date>`, one batch per invocation with a
report of what each batch labelled — not a single unattended sweep. A cron tick
never backfills, whatever the state file says.
