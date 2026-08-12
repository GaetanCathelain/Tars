# WF7 (GCN-44) — independent adversarial verification

Verifier: separate Orca subagent, 2026-08-12 ~15:00–15:12 UTC. Read-only on the
VM and on Gmail. No `sops -d`, no secret value read or printed, nothing written
on 192.168.0.9, 192.168.0.3 and 192.168.0.8 untouched. Every claim below was
re-derived from a primary source, not from `deploy.md`.

## 1. Gmail — 4/4 PASS

**1.1 PASS — all 8 taxonomy labels exist with the exact expected names.**
`mcp__claude_ai_Gmail__list_labels` (live API):

```
Label_6  🛠️ Outils & SaaS              522 msgs / 429 threads
Label_7  💳 Finance & Facturation       262 / 255
Label_8  👥 Interne Mobile Club         206 /  96
Label_9  🤝 Partenaires & Prestataires   34 /  19
Label_10 📅 Réunions & Agenda            83 /  78
Label_11 🎯 Recrutement                   8 /   6
Label_12 🔐 Sécurité & Comptes           68 /  62
Label_13 📣 Newsletters & Pub            64 /  60
```

Byte-exact match with the expected list, including the VS16 in `🛠️` and the
accents. IDs match the map in `~/.hermes/state/gmail-triage.json` one-for-one.

**1.2 PASS — 3 labels spot-checked, plausible and correctly routed.**
(`label:Label_N` returns `{}` — the connector needs the *display name*, not the
ID, contrary to its own tool doc. Noted; not a WF7 defect.)

- `label:"📣 Newsletters & Pub" after:2026/05/15` → 60 threads; senders are
  gifteo, dynatrace, neon changelog, atlassian, cloudflare, 1password,
  quable, wispr, sncf-voyageurs. All genuinely newsletters.
- `label:"🤝 Partenaires & Prestataires" after:2026/05/15` → 19 threads; senders
  rebura.com, techtribe.fr, adactim.com, modelabs.com, premaccess.com,
  nextmobiles.com, egym-wellpass. All outside vendors/partners.
- `label:"🛠️ Outils & SaaS" after:2026/05/15` → first page cloudflare.com,
  mail.anthropic.com, linear.app, vercel.com — all in the row-8 domain list.

**1.3 PASS — the backfill actually ran; the inbox residue is 4 threads.**
`in:inbox after:2026/05/15 has:nouserlabels` → `resultCountEstimate: 4`:
`support@fivetran.com` (arrived 15:03:26 UTC, i.e. **after** the 15:00 tick, so
correctly not yet triaged), plus exactly the three the leftovers agent declared
unlabelled: `noreply@infos.sncf-voyageurs.com`, `tyioposti@gmail.com`,
`scott-gordon@noreply.link`. Independent confirmation of that claim.

**1.4 PASS — pre-existing labels present and unrenamed.**
`Label_3 github` (7243), `Label_5175022515725755344 tech-mailgroup` (5760),
`Label_7907540109249447966 staging-recouvrement` (17868),
`Label_1/2/4/5 Hermes/Claude-Invoice-{Forwarded,Pending,Backfill-Forwarded,Backfill-Pending}`.
None renamed, none deleted, counts consistent with the skill's own §safety text
(which cites `staging-recouvrement` at 17868).

## 2. VM (read-only) — 5/5 PASS

**2.1 PASS — cron job exists exactly as claimed.** `~/.hermes/cron/jobs.json`,
job `191332ae6a24`: `"schedule": {"kind":"cron","expr":"0 7-21 * * *"}`,
`"skills":["gmail-triage"]`, `"deliver":"slack"`, `"enabled":true`,
`"state":"scheduled"`, `"repeat":{"completed":3}`,
`"created_at":"2026-08-12T16:31:44+02:00"`,
`"next_run_at":"2026-08-12T18:00:00+02:00"`, `"last_status":"ok"`. The prompt
begins `Run \`gmail-triage\` end to end… never \`--remove-labels\`…` — matches
SKILL.md § Scheduling contract. `next_run_at` is 18:00+02:00 at 15:04 UTC,
which independently re-confirms the expr is read in Europe/Paris.

**2.2 PASS — repo == live, md5 identical.**

```
8eaa93e425c3b5a4ed1ffb9f7107b84b  /home/gaetan/.hermes/skills/email/gmail-triage/SKILL.md   (20562 B, Aug 12 14:30)
8eaa93e425c3b5a4ed1ffb9f7107b84b  <repo>/skills/email/gmail-triage/SKILL.md
```

**2.3 PASS — state file exists, schema and cursor as specified.**
`~/.hermes/state/gmail-triage.json` (891 B, 15:01 UTC): `version:1`,
`cursor:"2026-08-12T17:00:51+02:00"` == `last_completed_run`, `seen` = 3 ids,
`labels` = the 8-entry name→ID map (matches §1.1 above exactly),
`last_run:{candidates:2,labelled:2,unrouted:0,overflow:0,by_label:{💳:1,🤝:1}}`,
`consecutive_failures:0`, `unrouted_senders:{}`. No lock left behind
(`gmail-triage.lock` absent).

**2.4 PASS — the quoted run outputs are real.**
`~/.hermes/cron/output/191332ae6a24/` holds exactly the three files quoted:
`2026-08-12_16-33-16.md`, `_16-37-03.md`, `_17-01-27.md`. Their `## Response`
sections are verbatim what `deploy.md` quotes: run 1 the single bullet
"State file was missing and recreated; cursor reset to 2026-08-12 16:32 Paris.
No historical mail was backfilled.", runs 2 and 3 `[SILENT]`.
`~/.hermes/logs/agent.log`:

```
22872 14:33:16,308 cron.scheduler: Job '191332ae6a24': delivered to slack:D0BBYNM01BL
22972 14:37:03,722 cron.scheduler: Job '191332ae6a24': agent returned [SILENT] — skipping delivery
23070 15:01:27,969 cron.scheduler: Job '191332ae6a24': agent returned [SILENT] — skipping delivery
```

Run 3's session id is `cron_191332ae6a24_20260812_170028` with the first
scheduler line at **15:00:28,900 UTC = 17:00 Paris** — the schedule fired it,
one DM total, no duplicate.

**2.5 PASS — both engagement-checker jobs present and unmodified.**
`62e8cd9db637` (`*/30 10-16 * * 1-5`, `deliver slack:C0BP2GZUFSR:1786516435.025349`,
enabled, `completed:42`, `last_status:ok`, `created_at 2026-08-08T15:49:26.187+02`)
and `759e08c598e3` (`0 17 * * 1-5`, same deliver target, `completed:7`,
`last_status:ok`, same creation timestamp). Both still carry their original
`origin` block and their WF6 prompt; neither was paused, retargeted or
rescheduled. Their `deliver` is indeed the rotating report thread, not a DM —
so the deploy agent's decision not to copy that literal string was correct.

## 3. Repo — 3/3 PASS

**3.1 PASS — both artifacts exist with version and provenance.**
`docs/specs/wf7-gmail-triage.md` (11.5 K, line 3: "Ticket: **GCN-44**") and
`skills/email/gmail-triage/SKILL.md` (front matter `version: 1.0.0`, line 25:
"Provenance: GCN-44 / WF7").

**3.2 PASS — safety rails forbid what they must.** § Safety rails states
`--remove-labels` is never passed ("never archive, never mark read/unread,
never un-label"), `--add-labels` carries only the eight table IDs (no `TRASH`,
`SPAM`, `STARRED`, `IMPORTANT`, `UNREAD`, no `Hermes/*`), `github`,
`tech-mailgroup`, `staging-recouvrement` and every `Hermes/*` are untouchable,
no `gmail send`/`reply`, ≤100 `modify` per run. Behaviour matches: the four
messages labelled during the runs all still carry `INBOX`, and the two from the
17:00 tick still carry `UNREAD` (verified live in §1 searches —
`19ff67c620facaea` `['UNREAD','IMPORTANT','Label_9','INBOX']`).

**3.3 PASS — the 🛠️ negative subject filters survive.** Row 8, line 223:
"…**and** the subject matches no row-1 and no row-2 cue. Those two negative
filters are not redundant with the ordering: keep both…".

## 4. Side-effect sweep — 3/3 PASS

**4.1 PASS — `config.yaml` untouched.** mtime `2026-08-12 08:43:24 UTC`, i.e.
~6 h **before** the 14:30 deploy; `.env` mtime `2026-08-07 20:13:49 UTC`. No
new `.bak` after 2026-08-11. Gateway not restarted:
`ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC`, `NRestarts=0`.

**4.2 PASS — no stray files under `~/.hermes/skills/email/`.** Contents:
`DESCRIPTION.md` (Aug 7), `durable-email-automation/` (Aug 10), `himalaya/`
(Aug 11), `rich-email-composition/` (Aug 2), and the new `gmail-triage/`
holding a single `SKILL.md`. No `.new`, no `.bak`, no leftovers.

**4.3 PASS — no state writes outside gmail-triage's own files.**
`find ~/.hermes/state -newermt "2026-08-12 14:25"` returns only
`gmail-triage.json`, `gmail-triage.json.bak-gcn44-20260812T143556Z`,
`gateway.heartbeat` (gateway's own), and `engagement-checker.json` +
`engagement-checker.lock` — the latter two written at 14:33:28 / 15:01:10 UTC by
the engagement checker's **own** ticks (`last_run_at 16:33:47+02:00`), not by
WF7.

## Notes (not failures)

- The `deploy.md` §7.1 off-by-one is **real and reproduced**: after run 2, state
  held `labelled:1`, `seen:["19ff456661cc3ad3"]`, yet both `19ff456661cc3ad3`
  and `19ff4c90892113cb` carry `Label_13` live. Under-reports volume; cannot
  cause a double-label (a labelled message drops out of `-has:userlabels`).
- `docs/specs/wf7-gmail-triage.md` line 3 still reads "Status: ACTIVE — authored
  2026-08-12, **not yet deployed**". Stale since 14:30 UTC.
- `docs/specs/wf7-gmail-triage.md`, `skills/email/gmail-triage/` and
  `status/probes/wf7/` are **untracked** in git — the evidence is on disk but
  not committed or pushed.
- Backfill dual-labelling is common and looks intentional (many 📣 threads also
  carry `Label_6`); it is not forbidden by the skill.

## Verdict

**PASS-with-notes** — 15/15 checks pass. Every load-bearing claim in
`deploy.md` (md5 identity, cron record, delivery target `D0BBYNM01BL`,
Paris-interpreted expr, unassisted 17:00 tick, `[SILENT]` de-duplication,
label correctness, no `--remove-labels`, no `config.yaml` edit, engagement
checker untouched) was independently re-derived from the live VM and the live
Gmail API and holds. The three notes are cosmetic/hygiene, not defects in the
shipped workflow.
