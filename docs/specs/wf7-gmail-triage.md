# WF7 — Gmail auto-triage

Status: LIVE — authored + deployed 2026-08-12 (cron job `191332ae6a24`,
verified `status/probes/wf7/verification.md`). Ticket: **GCN-44**.
Author: WF7 session (Orca on cooper). Decision basis: Gaetan's Slack DM of
2026-08-12, the 90-day mailbox pattern analysis, and a read-only VM probe of
`google_api.py` + the Hermes cron store (evidence:
`status/probes/wf7/`). Artifact: `skills/email/gmail-triage/SKILL.md` v1.1.0,
mirroring to `~/.hermes/skills/email/gmail-triage/SKILL.md`.

## Motivation

Gaetan's inbox has **no human categorisation at all**: of ~2843 inbox threads,
essentially 0% carry a usable business label — the only user labels that exist
are `github` and `tech-mailgroup` (skip-inbox filters that work fine),
`staging-recouvrement` (a dead fossil), and four `Hermes/Claude-Invoice-*`
labels Tars maintains for invoice accounting. Everything else is triaged by eye,
every day, against a stream that is 55-65% SaaS notification noise. Olivier set
up a labelled taxonomy on his own mailbox and Gaetan asked (Slack DM,
2026-08-12) for the same thing on his, driven by Tars rather than by hand-built
Gmail filters — filters cannot tell an Anthropic *receipt* from an Anthropic
*product notification*, and that distinction is most of the value.

## Taxonomy — 8 labels, IDs already created in Gmail

Evaluated **top to bottom, first match wins**. The order is load-bearing:
subject-specific rows (🔐, 💳) precede the broad domain row (🛠️) or the domain
row swallows receipts and security alerts; address-specific rows (📣, 🤝)
precede it for the same reason — `info@e.atlassian.com` is marketing while
`atlassian.net` is tooling, `ramzi@cloudflare.com` is an account manager while
`em@em1.cloudflare.com` is a drip campaign.

| # | Label | ID | Routing rule (summary — SKILL.md §3 is authoritative) |
|---|---|---|---|
| 1 | 🔐 Sécurité & Comptes | `Label_12` | subject: `security alert`, `suspicious login`, `API key/token`, `verifying it's you`, `2FA`, `password reset`; senders `google-workspace-alerts-noreply@google.com`, `no-reply@accounts.google.com`, `security@updates.linear.app` |
| 2 | 💳 Finance & Facturation | `Label_7` | subject: `receipt`, `invoice`, `facture`, `spend limit`, `cost report`, `billing`; senders `invoice+statements@mail.anthropic.com`, `accounts@1password.eu`, `no-reply@sns.amazonaws.com`, `notifications@doit.com` |
| 3 | 📅 Réunions & Agenda | `Label_10` | `Invitation:` / `Updated invitation:` / `Accepted:` subjects, meeting recaps and transcripts, `meetings-noreply@google.com`, Fireflies/Granola **when the subject names a meeting** |
| 4 | 🎯 Recrutement | `Label_11` | `candidature`, `alternance`, `stage`, `CV`, `entretien`, `application for` — sender not `@mobile.club` |
| 5 | 👥 Interne Mobile Club | `Label_8` | sender domain `mobile.club`, `cleaq.com` or `nextmobiles.com`, excluding `tech@mobile.club` |
| 6 | 🤝 Partenaires & Prestataires | `Label_9` | `techtribe.fr`, `egym-wellpass.com`, `ramzi@cloudflare.com`, DoIT relationship traffic |
| 7 | 📣 Newsletters & Pub | `Label_13` | `learn@send.zapier.com`, `updates@dynatrace.ai`, `info@e.atlassian.com`, `em@em1.cloudflare.com`, `changelog@neon.tech`, `news@gifteo.fr`, `no-reply@email.claude.com`, `noreply@tm.openai.com`; plain announcements/changelogs/webinars |
| 8 | 🛠️ Outils & SaaS | `Label_6` | ~21 SaaS domains (Anthropic, Vercel, Linear, GitHub, Datadog, 1Password, Notion, Atlassian, Cloudflare, OpenAI, Fivetran, Zapier, Neon, Granola, Fireflies) **minus** any row-1 or row-2 subject cue |

Rows the deterministic table misses go to LLM judgment, which may assign one of
the same eight or — the default — leave the message unlabelled. **An unlabelled
mail costs nothing; a wrongly-labelled one costs trust in every other label.**

Multi-label is allowed only where a message is genuinely both (canonical case:
the DoIT monthly cost digest → `Label_7,Label_9`), passed as one
comma-separated `--add-labels`. Never three.

`github`, `tech-mailgroup`, `staging-recouvrement` and `Hermes/*` are outside
the ID allowlist and therefore structurally untouchable.

## Architecture — why `google_api.py`

Three substrates exist on the VM. Only one can do this job.

- **`google_api.py` (`skills/productivity/google-workspace/scripts/`) — chosen.**
  `gmail search` takes a full Gmail query and returns `{id, from, subject,
  date, snippet}`; `gmail modify <id> --add-labels <IDs>` applies true Gmail
  multi-label semantics. The OAuth token at `~/.hermes/google_token.json`
  already carries `gmail.modify`, and the scope is **proven exercised, not just
  granted**: the live `Hermes/Claude-Invoice-Forwarded` label was created and
  applied under it by `~/.hermes/scripts/claude_invoice_forward.py`.
- **himalaya — rejected.** Its build has no `mailbox create`, so it cannot
  create a label at all, and `message move/copy` is IMAP folder semantics: a
  move takes the message *out of the inbox*, which this skill must never do.
- **Gmail MCP — unavailable.** `~/.hermes/config.yaml` `mcp_servers:` carries
  `linear`, `slack`, `notion` and nothing else; the only "google" hit is an
  unconfigured `google_chat:` platform toolset. There is no Gmail MCP server on
  this VM to call, headless or otherwise.

Two operational constraints follow from the probe and are baked into the skill:

1. **The interpreter is the hermes-agent venv**
   (`~/.hermes/hermes-agent/venv/bin/python`), never bare `python3` —
   `googleapiclient` is not on the system path.
2. **No batch.** `google_api.py` wraps a single
   `users().messages().modify()`; there is no `batchModify` anywhere in it. N
   messages means N calls, chained ≤10 per terminal command with `&&`. Under
   `approvals.cron_mode: deny`, `python3 -c`, heredocs and pipes-into-an-interpreter
   are blocked shapes that degrade a cron run to a healthy-looking `[SILENT]`;
   a bare interpreter path plus a script path is the safe shape. **v1.0.0 needs
   no script of its own** — plain `google_api.py` invocations only.

**State**: `~/.hermes/state/gmail-triage.json` — cursor, the name→ID map above,
a 500-entry `seen` ring of message ids (Gmail search is date-granular, so the
run widens to the cursor's date and filters locally; the ring is what makes
that overlap idempotent), per-run stats, `unrouted_senders`, and
`consecutive_failures`. Recreated from the skill's own literal shape if lost —
without a backfill, because a self-authorised backfill after state loss is
~2800 unreviewed writes.

**Report**: Slack DM, ≤6 bullets, or exactly `[SILENT]`. Silence breaks only
for a 🔐-labelled item, an unroutable sender domain seen 3+ times, an error or
scope failure, or a >100-item overflow. Failed writes, auth failures and
overflow are never rate-limited away.

## Schedule

One Hermes cron job — `0 7-21 * * *`, hourly on the hour, 07:00–21:00, every
day. Hermes' configured timezone is `Europe/Paris` and **a cron job carries no
timezone field of its own** (`schedule.kind: "cron"` + a raw 5-field
`schedule.expr`), so the expression is Paris time as written; no adaptation
needed. Unlike `engagement-checker` there is no weekday restriction and no
second end-of-day pass: mail arrives at weekends and labelling it is
consequence-free.

```bash
hermes cron create '0 7-21 * * *' --name "Gmail triage" \
  --skill gmail-triage --deliver slack '<prompt from SKILL.md § Scheduling contract>'
hermes cron list        # confirm [active] and the next_run_at offset
```

`--deliver slack` (bare platform name) resolves to `SLACK_HOME_CHANNEL`,
Gaetan's DM. Manual exercise is `hermes cron run <job_id>` — it **delivers to
the real target**, there is no dry-run flag.

## Deployment recipe

Repo mirror is `skills/email/gmail-triage/SKILL.md`; the VM path is the same
relative path under `~/.hermes/skills/`. The skill is new, so there is no
`.bak` to take on the first deploy — take one on every subsequent one.

```bash
# 1. destination dir
ssh gaetan@192.168.0.9 'mkdir -p ~/.hermes/skills/email/gmail-triage'

# 2. (upgrades only) dated backup of the live file, first
ssh gaetan@192.168.0.9 'cp ~/.hermes/skills/email/gmail-triage/SKILL.md \
  ~/.hermes/skills/email/gmail-triage/SKILL.md.bak-gcn44-$(date -u +%Y%m%dT%H%M%SZ)'

# 3. stage
cat skills/email/gmail-triage/SKILL.md | ssh gaetan@192.168.0.9 \
  'umask 077; cat > ~/.hermes/skills/email/gmail-triage/SKILL.md.new'

# 4. atomic swap under the shared lock
ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c \
  "mv ~/.hermes/skills/email/gmail-triage/SKILL.md.new ~/.hermes/skills/email/gmail-triage/SKILL.md"'

# 5. verify: live == repo
md5sum skills/email/gmail-triage/SKILL.md
ssh gaetan@192.168.0.9 'md5sum ~/.hermes/skills/email/gmail-triage/SKILL.md'

# 6. create the cron job (§ Schedule), then hermes cron list
```

Hermes live-reloads the skill (~30 s). Confirm no restart happened by comparing
`ActiveEnterTimestamp` on `hermes-gateway.service` — **not** `NRestarts`, which
provably misses in-band respawns.

Acceptance, in order, evidence into `status/probes/wf7/`:

1. `hermes cron list` shows the job `[active]` with a Paris `next_run_at`.
2. One `hermes cron run <job_id>`: the state file exists with the 8-entry label
   map and a cursor; ≥1 message carries a taxonomy label in Gmail; the delivery
   is either `[SILENT]` or ≤6 conforming bullets.
3. The **next** tick labels only new mail — no re-labelling, no duplicate DM.
4. Spot-check 10 labelled messages by hand against the table. A wrong label is
   a rule bug to fix in §3, not a reason to un-label (nothing un-labels).

## Rollback

```bash
hermes cron remove <job_id>                              # stops the schedule
ssh gaetan@192.168.0.9 'rm -rf ~/.hermes/skills/email/gmail-triage'
```

Optionally `rm ~/.hermes/state/gmail-triage.json`. **Labels already applied
stay** — removing them is a mailbox mutation the skill is forbidden to make and
a rollback has no business making either. They are inert: a Gmail label on a
message changes nothing except what Gaetan can filter on, and he can bulk-remove
any of them in the UI in seconds if he wants them gone.

## Open items — Gaetan's call, not a run's

- **`nextmobiles.com` routing — RESOLVED (Gaetan, 2026-08-12).** The sister
  brand is "us": it routes to 👥 Interne Mobile Club (`Label_8`, row 5), not
  🤝 Partenaires & Prestataires. Applied in SKILL.md v1.1.0 — rows 5 and 6 both
  changed. Labels already applied under v1.0.0 stay (nothing un-labels).
- **`staging-recouvrement` deletion.** 445 threads / 17868 messages, zero
  activity in 90+ days — the largest label on the account and a dead one.
  Deleting the label (not the mail) is the obvious cleanup, and it is **gated on
  Gaetan's explicit go**; this skill will never touch it.
- **DoIT double identity.** Relationship mail is 🤝, the monthly cost digest is
  💳 — the one sanctioned multi-label case. Revisit if the digest volume grows.
- **Fireflies / Granola.** The same sender emits onboarding noise (🛠️) and real
  meeting recaps (📅); only the subject separates them. Row 3 encodes that;
  expect a few misses.
- **`category:promotions` is not usable locally.** It is a server-side Gmail
  operator, and `gmail search` returns no label field, so 📣 relies on the
  sender list plus judgment. A second bounded search per run would sharpen it —
  deferred as not worth the extra call until 📣 is measurably weak.
- **`google-workspace` is a vendored upstream skill and is not mirrored in this
  repo** (GCN-31). A `hermes` upgrade can silently revert its venv-interpreter
  fix and break this skill's only transport. Same exposure as
  `engagement-checker` §3; tracked there, inherited here.
