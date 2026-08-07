# Probe — retrieval trials, 5 real questions x N mechanisms (P6/kb)

All commands run 2026-08-07 22:45–23:15 UTC. Filesystem work is read-only over
`ssh gaetan@192.168.0.9`; live-agent turns are `~/.local/bin/hermes chat -Q -q`
oneshots (the project's sanctioned probe pattern). **No writes** were made under
`~/.hermes/` or `~/dev/mc-metarepo`. Scratch files live in `/tmp/kbtrial/` on
the VM and one map file at `/tmp/kbtrial/map.md`.

Inputs: `kb-hermes-retrieval.md` (mechanisms) and `kb-mc-metarepo.md` (the KB).

> **Side effects to flag, caused by baseline turns, not by me.** Baselines Q1
> and Q2 each made Tars **delegate to an Orca worker on cooper**, creating two
> worktrees and two branches:
> `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/api-v2-order-origin` on branch
> `GaetanCathelain/api-v2-order-origin` (run `run_33222715552e`), and
> `…/vecna-post-order-status` on branch `GaetanCathelain/vecna-post-order-status`
> (run `run_9a6a5e637304`). **Baseline Q2's worker also queried live production
> Datadog** (spans, logs and 5 monitors) to answer a question a 971-byte note
> already answers. `write_file` calls by these turns landed in `/tmp` only
> (`/tmp/verify-api-v2-order-origin.py`, `/tmp/tars-brief-api-v2-order-origin.md`,
> `/tmp/tars-brief-vecna-post-order-status.md`) — verified nothing under
> `~/.hermes/` or the clone itself was written by my turns.

---

## The 5 questions + ground truth

Ground truth was established **before** any trial, by reading the files
directly. Each is a question Gaetan would plausibly DM Tars.

### Q1 — easy, single file, vocabulary matches

> "In api-v2, how do I tell a Next Mobiles order apart from a Mobile.Club order?"

**Ground truth** — `knowledge/nm-vs-mc-order-source.md` (767 B), verbatim:

```
To tell a **Next Mobiles** order apart from a **Mobile.Club** order, check for
the existence of a row in the `next_mobile_order_mappings` table (column
`order_id`):
- a mapping exists ⇒ **Next Mobiles**
- no mapping ⇒ **Mobile.Club**
This is the pattern already used by `IsNextMobilesOrderService`.
## What not to do
**Do NOT** use `orders.business_entity_id` as the source discriminant:
`CreateNmOrderService` does not set it when creating an NM order, so it stays at
the default Mobile.Club value and is not reliable.
```

A correct answer must carry **both** halves — the discriminant *and* the
`business_entity_id` trap. The trap is the operational value of the note.

### Q2 — easy, single file, vocabulary matches

> "Datadog shows `VecnaApi.postOrderStatus` at about 92% errors. Is the service
> actually broken?"

**Ground truth** — `knowledge/datadog-apm-error-sampling.md` (971 B), verbatim:

```
APM ingestion keeps **100% of error spans** (`ingestion_reason:error`) but
**samples `ok` spans** (`diversity_sampling`). As a result, a raw span search
shows massively inflated error *ratios*.
...
Real case: `VecnaApi.postOrderStatus` looked ~92% in error in the spans, while
the unsampled logs showed only a handful of errors over 2 days.
```

Answer: **no** — it is a sampling artifact; read rates from the logs
(`service:(api-v2 OR api-v2-jobs)`), never from raw spans.

### Q3 — HARD: vocabulary mismatch (the obvious keyword is absent from the answering file)

> "Last week I ran a one-off command on the shared agents VM with an API key
> inline in the command. Nobody was watching the terminal. Do I actually need
> to rotate the key?"

**Ground truth** — `knowledge/sudo-argv-secrets-authlog.md` (1509 B), verbatim:

```
`sudo` logs the **full argv** of every command it runs, at auth priority, into
`/var/log/auth.log` **and** the systemd journal — inline `env VAR=value`
assignments included. One `sudo -u <user> env SECRET=… <cmd>` invocation
persists the secret in plaintext to two log stores, readable by every admin
(on mc-agents: every `mc-agents-ssh` member via sudo), and surviving in rotated
copies.
...
- If it happens anyway: the remediation is **rotating the credential**, not
  editing the logs — truncating auth.log / vacuuming the journal erases the
  audit trail along with the evidence, and rotated copies remain.
```

Two things make this the discriminating case:

1. **The mismatch is measured, not assumed.** The user's own words do not occur
   in the answering file:
   ```
   $ cd ~/dev/mc-metarepo && rg -il 'api key|api_key|apikey' .
   ./learnings/next/wordpress-monorepo.md
   ./learnings/mobileclub/support-engineer.md
   ./knowledge/linear-agent-write-surfaces.md
   ./schemas/next/wordpress-mysql.full.md
   ./schemas/next/wordpress-mysql.core.md
   ./next/helm-charts.md
   ./schemas/next/vecna-mysql.full.md
   ./next/terraform-infrastructure.md
   ./schemas/next/vecna-mysql.core.md
   ./next/INFRASTRUCTURE.md
   ./mobileclub/strapi.md
   ./index/next/terraform-infrastructure.json
   ./mobileclub/return-app.md
   ```
   13 files hit; **`knowledge/sudo-argv-secrets-authlog.md` is not one of them.**
2. **There is a wrong-but-plausible answer available from priors**, and it is
   the one the KB explicitly forbids: "clean the logs". A mechanism that answers
   from the model's general knowledge will produce it.

**Bonus KB-hygiene finding**, measured while building this question — the
`knowledge/README.md` index table is **incomplete**, and the Q3 note is one of
the missing entries:

```
$ ls knowledge/*.md | ... | wc -l   → 40 notes
$ rg -o '^\| \[([a-z0-9-]+)\]' -r '$1' knowledge/README.md | wc -l   → 35 indexed
$ comm -23 notes.txt indexed.txt
loop-backfilled-items-missing-active
return-app-datadog-logs
sudo-argv-secrets-authlog
tech-product-okr-q3-2026
workspace-commit-conventions
```
The complete manifest is instead the `@knowledge/*.md` import list in the repo's
`CLAUDE.md` (lines 83–120, 38 entries), which *does* include the Q3 note.
**Any design that routes through `knowledge/README.md` inherits a 5-note blind
spot; routing through the filenames does not.**

### Q4 — HARD: answer spread across 3 files

> "If one of my sessions learns something durable about the Vecna codebase, what
> exactly has to happen for it to land in the team metarepo, and what would get
> it rejected?"

**Ground truth** — three files, none sufficient alone:

| File | Bytes | Contributes |
|---|---|---|
| `learnings/README.md` | 5350 | The contract: PR labeled `agent-learning`, **"A human MUST review and merge every learnings PR; agents MUST NEVER self-merge"**; no PII/secrets (rule 5); tech debt ⇒ Linear issue, not an entry (rule 7); regeneration never touches `learnings/` (rule 2) |
| `CONTRIBUTING.md` | 8261 | **English only**, plus **ASD-STE100 Simplified Technical English**, CI-gated: "CI fails a pull request on an STE **error** in an added line" |
| `CLAUDE.md` | 6858 | The write-rule table: `learnings/` = "append via PR labeled `agent-learning`, **human-merged only**"; `cleaq/ mobileclub/ next/` pinned docs "**never hand-edit** — regenerate" |

Total ground-truth surface: **20,469 bytes across 3 files.**

### Q5 — medium: multi-file, vocabulary matches

> "A Next Mobiles customer says they never received their bon de retour. Where
> do I start looking, and is there anything I must not do?"

**Ground truth** — `knowledge/vecna-owns-returns.md` (3993 B), plus its
cross-reference `knowledge/vecna-order-error-manual-relaunch.md` (5196 B):

- Start in **Vecna** MySQL (`event`, `subscription`, `order_line`,
  `notification_history`), not Loop — Vecna owns the BR; Loop only mints the
  Chronopost label once and "deliberately skips every NM customer email".
- **Never** "Regenerate tracking code" (`recreateChronopostTracking`) on an NM
  return parcel — "it invalidates the reservation without telling Vecna and
  kills the customer's live BR link". Use admin **"Download return label"**.
- Check `wp_nxt_tasks` (`done=0, retry_count=5` = dead task) before declaring
  the `RETURN_CHANGE` event was never fired — it is WordPress cron lag, not a
  Loop bug.

---

## Mechanism (a) — naive ripgrep

The laziest thing: one `rg -i` with the keywords a model would lift straight
out of the question, over the whole clone.

```
$ cd ~/dev/mc-metarepo && rg -i -- "<pattern>" .
```

Timings from `date +%s.%N` around the call; bytes/lines from `wc` on the
captured output.

| Q | pattern | seconds | bytes | lines | files hit | GT file located? | **answer in the output?** |
|---|---|---|---|---|---|---|---|
| Q1 | `next mobiles order` | 0.007 | 337 | 3 | 3 | yes | **no** |
| Q2 | `error rate` | 0.006 | 369 | 3 | 3 | yes | **no** |
| Q3 | `api key` | 0.007 | 1419 | 6 | 5 | **no** | **no** |
| Q4 | `learning` | 0.007 | 15457 | 121 | 53 | yes | partial |
| Q5 | `bon de retour` | 0.007 | 1349 | 11 | 7 | yes | **no** |

Verbatim outputs (Q1, Q2, Q5 shown in full; Q4 truncated to its first 12 lines):

```
===== q1 =====
./learnings/mobileclub/workspace.md:  refused on a Next Mobiles order.
./knowledge/README.md:| [nm-vs-mc-order-source](nm-vs-mc-order-source.md) | MobileClub | Tell a Next Mobiles order ap…
./knowledge/nm-vs-mc-order-source.md:# Telling a Next Mobiles order apart from a Mobile.Club order (api-v2)
===== q2 =====
./MCP.md:- "What's the gotcha about reading error rates from Datadog spans?" → finds
./knowledge/README.md:| [datadog-apm-error-sampling](datadog-apm-error-sampling.md) | MobileClub | Datadog EU samples…
./knowledge/datadog-apm-error-sampling.md:# Datadog (EU): never read an error rate from raw spans
===== q3 =====
./learnings/mobileclub/support-engineer.md:## 2026-08-05 — applying a Datadog monitor with no API key, from the logge…
./next/terraform-infrastructure.md:- **datadog** — IRSA role + Secrets Manager container for the Datadog Agent API ke…
./next/terraform-infrastructure.md:- Never store secrets in Terraform: the Datadog API key pattern (Secrets Manager c…
./next/INFRASTRUCTURE.md:- **Datadog API key = Secrets Manager** (`datadog/<env>/agent`). Terraform creates only the …
./schemas/next/wordpress-mysql.core.md:- `wp_woocommerce_api_keys`, `wp_wc_webhooks` — REST API keys & outbound webho…
./index/next/terraform-infrastructure.json:    "notes": "datadog stack provisions only the IRSA role + Secrets Manage…
===== q5 =====
./learnings/next/vecna.md:- **Claim:** The NextMobiles "bon de retour" link
./learnings/mobileclub/workspace.md:- **Claim:** Loop ops facts for "reprendre un appareil / bon de retour":
./next/next-bdc.fiches.md:Le client peut télécharger le bon de retour depuis le lien
./next/next-bdc.fiches.md:     déverrouillage de l'ancien, le bon de retour est téléchargeable via un
./next/next-bdc.fiches.md:Il  peut télécharger le bon de retour via un lien envoyé par mail, par sms ou depuis son es…
./next/next-bdc.fiches.md:que le bon de retour.
./next/next-bdc.fiches.md:Répondre au client que la restitution se fait exclusivement avec le bon de retour via le tr…
./SKILLS.md:| [`triage-missing-vecna-return-event`](.claude/skills/triage-return-label-not-delivered/SKILL.md) | Patt…
./knowledge/vecna-owns-returns.md:For a **Next Mobiles (NM) order**, the customer-facing `bon de retour` (BR) is
./knowledge/README.md:| [vecna-owns-returns](vecna-owns-returns.md) | Next / MobileClub | For an NM order the `bon de…
./knowledge/wp-feature-flags-helm.md:Real case (MC-3639): the "return slip" (`bon de retour`) button didn't show on
===== q4 head -12 =====
./learnings/README.md:# learnings/ — agent-contributed operational learnings
./learnings/README.md:learnings/<stack>/<repo>.md      # sidecar of <stack>/<repo>.md
./learnings/README.md:   requests labeled `agent-learning`, opened by an agent against `main`.
./learnings/README.md:   **A human MUST review and merge every learnings PR; agents MUST NEVER
./learnings/README.md:2. **The regeneration pipeline MUST NEVER touch `learnings/`.** The pinned
./learnings/README.md:   regenerated from a fresh clone; `learnings/` is additive, human-curated
./learnings/README.md:3. **Learnings never patch pinned docs.** If an entry shows a pinned doc has
./learnings/README.md:4. **Durable facts only — never ticket-specific ones.** A learning must still
./learnings/README.md:7. **Tech debt gets a Linear issue, not an entry.** If the "learning" is really
./learnings/README.md:   merged — see [../CONTRIBUTING.md](../CONTRIBUTING.md#tech-debt-is-a-linear-issue-not-a-learn…
./learnings/README.md:Each sidecar file is a `# <repo> — learnings` header followed by appended
./learnings/README.md:  metarepo docs are ...
```

**Reading.** Naive rg is fast and cheap but it **locates without answering**.
For Q1, Q2 and Q5 the only match in the correct file is the **title line** — the
answer is 3–20 lines further down and never enters the output. Naive rg is a
*router*, not a retriever. On Q3 it does not even route: 1419 bytes of Datadog
and Terraform noise, zero hits on the answering file. Q4 is the one case where
grepping a broad term happens to surface real contract lines, because
`learnings/README.md` is written in exactly that vocabulary — and it costs
15 KB across 53 files to get there.

## Mechanism (b) — scoped ripgrep, file-list first

`rg -li` for the file list, then read the top file(s).

```
$ cd ~/dev/mc-metarepo && rg -li -- "<pattern>" .
```

Verbatim lists (order is ripgrep's traversal order, **not** relevance):

```
### q1  rg -li [next mobiles order]     list 95 B / 3 files
./learnings/mobileclub/workspace.md
./knowledge/README.md
./knowledge/nm-vs-mc-order-source.md

### q2  rg -li [error rate]             list 73 B / 3 files
./MCP.md
./knowledge/datadog-apm-error-sampling.md
./knowledge/README.md

### q3  rg -li [api key]                list 185 B / 5 files
./learnings/mobileclub/support-engineer.md
./next/terraform-infrastructure.md
./schemas/next/wordpress-mysql.core.md
./next/INFRASTRUCTURE.md
./index/next/terraform-infrastructure.json

### q4  rg -li [learning]               list 1757 B / 53 files
./learnings/README.md
./index/README.md
./learnings/next/vecna.md
./learnings/next/wordpress-monorepo.md
./learnings/cleaq/risk.md
(… 48 more)

### q5  rg -li [bon de retour]           list 193 B / 7 files
./learnings/next/vecna.md
./learnings/mobileclub/workspace.md
./next/next-bdc.fiches.md
./SKILLS.md
./knowledge/vecna-owns-returns.md
./knowledge/wp-feature-flags-helm.md
./knowledge/README.md
```

### (b) as literally specified — "read the top file"

This is a **trap**, and it is worth measuring because it is the obvious
implementation:

```
$ for f in $(head -1 q1.list) $(head -1 q2.list) $(head -1 q5.list); do wc -c $f; done
70381 ./learnings/mobileclub/workspace.md      # q1 top file — WRONG file
 6119 ./MCP.md                                 # q2 top file — WRONG file
49711 ./learnings/next/vecna.md                # q5 top file — partial
```

Reading the first file rg lists costs **70 KB** on Q1 and lands on the wrong
document. rg's output order is filesystem order.

Ranking by hit count does not rescue it:

```
$ rg -c -i "bon de retour" .   | sort -t: -k2 -rn | head -3
./next/next-bdc.fiches.md:5      ← 152 KB auto-generated CS script, the worst possible read
./SKILLS.md:1
./learnings/next/vecna.md:1
$ rg -c -i "next mobiles order" . | sort -t: -k2 -rn
./learnings/mobileclub/workspace.md:1
./knowledge/README.md:1
./knowledge/nm-vs-mc-order-source.md:1     ← all tied at 1
```

### (b′) with a `knowledge/`-first directory heuristic

The fix is a **directory** heuristic, not a scoring one: prefer a `knowledge/`
hit, because that directory's contract is "one note = one fact".

| Q | list bytes | file read | read bytes | **total** | correct? |
|---|---|---|---|---|---|
| Q1 | 95 | `knowledge/nm-vs-mc-order-source.md` | 767 | **862** | yes (both halves) |
| Q2 | 73 | `knowledge/datadog-apm-error-sampling.md` | 971 | **1044** | yes |
| Q3 | 185 | — no `knowledge/` hit exists | — | 185 | **NO — not routable** |
| Q4 | 1757 | `learnings/README.md` | 5350 | **7107** | partial (misses English/STE + write-rule table) |
| Q5 | 193 | `knowledge/vecna-owns-returns.md` | 3993 | **4186** | yes |

(Q3's 1509-byte read was performed for the byte measurement only; the file is
**not** in the candidate list, so the mechanism cannot reach it. Scored as a
failure.)

## Mechanism (c) — skill-guided map

The map a `SKILL.md` would carry. Written from B's inventory plus the slug list.
Full text: **`status/probes/wf5/kb-map-draft-SKILL.md`** in this repo (and
`/tmp/kbtrial/map.md` on the VM, which is what the live turns below actually
consumed). **3,279 bytes.** Its load-bearing parts:

```markdown
| Gaetan asks about | Read |
|---|---|
| an operational gotcha, "why does X behave like that", "is it safe to Y" | `knowledge/<slug>.md` — one file = one fact. Pick the slug below. |
| deep per-repo detail, ticket history, real queries | `learnings/<stack>/<repo>.md` (append-only, 20–70 KB — grep, don't read whole) |
| what a system *is* (stack, deploy, seams) | `<stack>/<repo>.md` (pinned) or `index/<stack>/<repo>.json` |
| DB tables/columns | `schemas/<stack>/<db>.core.md` |
| repo rules, contributing, language, PR flow | `CLAUDE.md`, `CONTRIBUTING.md`, `learnings/README.md` |

## knowledge/ — 40 notes, the filename IS the index
admin-codegen-offline · agent-browser-evidence · … · sudo-argv-secrets-authlog ·
tech-product-okr-q3-2026 · vecna-order-error-manual-relaunch · vecna-owns-returns · …

## Rules
1. Match the question's **meaning** to a slug, then `read_file` that one file.
   The answering note almost never uses Gaetan's words — route on meaning, not
   on keyword match.
2. Only if no slug fits: `search_files` file-list-first, scoped to `knowledge/`
   and `learnings/`.
3. Never read whole: `next/next-bdc.fiches.md` (152 KB), `schemas/*.full.md`,
   `learnings/mobileclub/workspace.md` (70 KB). Grep them.
4. The submodules are **empty** in this clone. Product source code is NOT here —
   do not delegate a code-reading session to find what a note already states.
```

The slug list is the whole trick: `sudo-argv-secrets-authlog` is a *semantic*
handle. A model that cannot keyword-match "API key" can still recognise
"secrets on a command line ⇒ argv ⇒ auth.log". The full map of every markdown
file in the repo — `find . -name '*.md'` — is only **4,289 bytes**, so the map
is cheap at any granularity.

### (c) fired live on the agent (Q1 and Q3)

The map was prepended to the prompt of a oneshot, simulating the skill being in
context. No skill was installed (that would be a write to `~/.hermes/skills/`,
forbidden).

**Q3 — the hard case, map-guided.** 21 s, verbatim:

```
$ ~/.local/bin/hermes chat -Q -q "$(cat /tmp/kbtrial/map.md)

---
Using the knowledge base described above, answer: Last week I ran a one-off command on the shared agents VM with an API key inline in the command. Nobody was watching the terminal. Do I actually need to rotate the key? Cite the file you used."

session_id: 20260807_230042_122c15
Yes—rotate the key.

Nobody watching the terminal doesn't matter: an inline secret can persist in transcripts; if the command used `sudo`, its full arguments were also written in plaintext to `/var/log/auth.log` and the systemd journal, including rotated copies readable by admins. The prescribed remediation is rotation, not deleting logs.

Source: `/home/gaetan/dev/mc-metarepo/knowledge/sudo-argv-secrets-authlog.md`, especially lines 5–10 and 20–22.
=== WALLCLOCK: 21s ===
```

Tool trace (`~/.hermes/logs/agent.log`) — **one** retrieval call:

```
agent.tool_executor: Tool kanban_show returned error (0.00s): {"error": "task_id is required (o…
agent.tool_executor: tool read_file completed (0.05s, 1717 chars)
agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol
```

**Q1 — map-guided.** 19 s, verbatim:

```
session_id: 20260807_230119_7e10e7
Check `next_mobile_order_mappings.order_id`:

- Mapping exists → Next Mobiles order
- No mapping → Mobile.Club order

Do not use `orders.business_entity_id`: Next order creation leaves it at the Mobile.Club default. This is the logic used by `IsNextMobilesOrderService`.

Source: `~/dev/mc-metarepo/knowledge/nm-vs-mc-order-source.md`
=== WALLCLOCK: 19s ===
```

Tool trace — again **one** call:

```
agent.tool_executor: tool read_file completed (0.05s, 958 chars)
```

### (c) byte costs

| Q | routed to | file bytes | tool-output bytes | fired live? |
|---|---|---|---|---|
| Q1 | `knowledge/nm-vs-mc-order-source.md` | 767 | **958** (measured) | yes |
| Q2 | `knowledge/datadog-apm-error-sampling.md` | 971 | ~1.2 K (est.) | **no** |
| Q3 | `knowledge/sudo-argv-secrets-authlog.md` | 1509 | **1717** (measured) | yes |
| Q4 | `learnings/README.md` + `CONTRIBUTING.md` + `CLAUDE.md` | 20,469 | ~21 K (est.) | **no** |
| Q5 | `knowledge/vecna-owns-returns.md` | 3993 | ~4.2 K (est.) | **no** |

Plus the map itself, **3,279 bytes, once per session**, not once per question.

## Mechanism (d) — what Hermes ships natively

Per recon A, the retrieval surface already live on both CLI and Slack is
`search_files` + `read_file` + `terminal` (all in `_HERMES_CORE_TOOLS`), and
`lcm_*` over the agent's own conversation store. There is **nothing else** —
no document RAG, no index. Mechanism (d) with no map is therefore identical to
the Baseline section below; mechanism (d) with a map is mechanism (c). The two
observations worth recording separately:

- **`search_files` works, but the agent aims it at the wrong path.** Baseline
  Q4's first two tool calls, verbatim:
  ```
  Tool search_files returned error (0.06s): {"total_count": 0, "error": "Path not found: /home/gaetan/dev/gaetan-metarepo"}
  Tool search_files returned error (0.05s): {"total_count": 0, "error": "Path not found: /home/gaetan/dev/gaetan-metarepo"}
  ```
  `gaetan-metarepo` is the path named in `~/.hermes/SOUL.md:54` — and it is on
  **cooper**, not on the VM:
  ```
  $ grep -in "metarepo" ~/.hermes/SOUL.md
  54:source of truth is `~/dev/gaetan-metarepo` on cooper; the live mechanism on my
  ```
  SOUL.md names a metarepo that does not exist locally and never mentions
  `~/dev/mc-metarepo`. **The agent's only pointer to a knowledge base is a dead
  path on the wrong machine.**
- **LCM answers KB questions from conversation memory, not from the KB** — see
  Baseline Q5.

## Mechanism (e) — a stock filesystem MCP server, on paper only

Not installable and not registered (recon A: `node`/`npx` absent from the VM;
no filesystem/RAG server in `hermes mcp catalog`). Evaluated on paper against
these 5 questions:

- The official `@modelcontextprotocol/server-filesystem` exposes `read_file`,
  `list_directory` and a `search_files` — i.e. **the same three primitives
  Hermes already has natively**, plus a Node runtime as a new dependency. It
  would score identically to mechanism (b′): it fixes nothing on Q3, because
  the failure is semantic routing, not tool availability.
- The team reached the same conclusion independently. `mc-metarepo/MCP.md`
  §"What it is (KISS, by design)", verbatim:
  ```
  - It exposes the repo's files for **read + built-in filename/content search**
    (`search_files`), complemented by the MCP client's own ripgrep …
  - **No vector DB, no embeddings, no index service, no credentials.** …
  This is the explicit KISS decision: the metarepo is small, already
  well-structured, and git-versioned — grep + structured JSON records beat a
  vector store here, and cost nothing to operate.
  ```
  Their proposed Knowledge MCP is a wrapper around `search_files`. Tars has
  `search_files`. **Adding the MCP server would buy Tars nothing it does not
  already have** — the missing piece is the map, not the tool.

---

## Baseline — what Tars answers TODAY

Five oneshots, no map, no skill, exactly as Gaetan would DM them. Wall clock
from `date +%s` around the call; tool traces from `~/.hermes/logs/agent.log`.
No turn hit the known "incomplete after 3 continuation attempts" intermittent,
so no retry was needed.

| Q | session | seconds | tool turns | verdict | answered *from the KB*? |
|---|---|---|---|---|---|
| Q1 | `20260807_225044_b6d355` | **422** | 25 | partial — right discriminant, **missed the trap** | no — Orca worker read product source |
| Q2 | `20260807_225810_7d8488` | **414** | 16+ | correct | no — Orca worker queried **live production Datadog** |
| Q3 | `20260807_225812_fd1da1` | 22 | 1 (a failed `kanban_show`) | **WRONG** — from priors, contradicts the KB | no — zero retrieval |
| Q4 | `20260807_225816_2a3dac` | 113 | 9 | **correct**, via ~42 KB of `terminal` output | **yes** — the only one |
| Q5 | `20260807_225817_764c60` | 62 | 8 | correct | no — from `lcm.db` conversation memory |

**3/5 correct, but 1/5 from the knowledge base.** Tars mostly gets the right
answer by re-deriving it from the world — source code, production telemetry, or
its own memory — never by reading the note that already says it. Two of the five
turns cost ~7 minutes, a git branch and a delegated Claude Code session each.

### Baseline Q1 — 422 s, delegated to an Orca worker

```
$ ~/.local/bin/hermes chat -Q -q "In api-v2, how do I tell a Next Mobiles order apart from a Mobile.Club order?"

session_id: 20260807_225044_b6d355
Use the `source` field returned by the `nest_search` GraphQL query:

- `NEXT_MOBILES` → Next Mobiles order
- `MOBILE_CLUB` → Mobile.Club order
- `CLEAQ` → Cleaq order

The underlying rule is:

    cleaq_order_id exists                  → CLEAQ
    next_mobile_order_mappings row exists → NEXT_MOBILES
    otherwise                              → MOBILE_CLUB

Verified in:
- `packages/api-v2/src/modules/search/domain/search-result.ts:29-33`
- `packages/api-v2/src/modules/search/infrastructure/search-provider-pg.ts:111-117`
- `packages/api-v2/src/modules/search/infrastructure/search.resolver.ts:22`
…
Delegated inspection command:
orca orchestration worker-start --run "run_33222715552e" --task "task_4f9e05d18e4e" … --agent claude …
Run `run_33222715552e`; dispatch `ctx_dca005ed6e6c`; outcome `succeeded`. Worktree:
`/home/gaetan/dev/mc-metarepo/null/mc-metarepo/api-v2-order-origin`, branch `GaetanCathelain/api-v2-order-origin`.
=== WALLCLOCK: 422s ===
```

Tars **never opened the 767-byte note**. It spawned a Claude Code session that
checked out the real product source and re-derived the answer from code — seven
minutes, a git worktree and a branch for a fact that was already written down.
It also **missed the `business_entity_id` trap entirely**, which is the half of
the note that actually prevents a bug.

Confirmation that it *did* know the clone existed but went hunting for source
code inside it (`/tmp/tars-brief-api-v2-order-origin.md`, written by the turn):

```
Goal: determine exactly how callers of api-v2 can distinguish a Next Mobiles order from a Mobile.Club order.
Where: inspect the mc-metarepo repository, especially api-v2 order schemas, serializers, endpoints, domain models, te…
```

The submodules in that clone are **empty** (recon B). The agent constructed
`/home/gaetan/dev/mc-metarepo/null/mc-metarepo/api-v2-order-origin/monorepo-loop`
and delegated a fresh checkout to fill the gap.

### Baseline Q2 — 414 s, correct, via a live production investigation

```
$ ~/.local/bin/hermes chat -Q -q "Datadog shows VecnaApi.postOrderStatus at about 92% errors. Is the service actually broken?"

session_id: 20260807_225810_7d8488
No — the service is not broken. The ~92% is a misleading Datadog APM sampling/retention artifact, not the real failure rate.

Fresh production evidence:
- Last 15 days: 594 retained spans OK, 1 error — 0.17%, not 92%.
- Last 2 days: 68 OK, 0 errors.
- Unsampled logs show ~196 Vecna calls/day continuing normally.
- Only 2 "Vecna API error" logs in 15 days; both were deterministic 404 "order not found" responses …
- All five Vecna monitors currently report OK.
…
Verdict: no incident. Fix the dashboard/error-rate query so it doesn't divide preferentially retained error spans by sampled successes; separately repair monitor `110136431`.

Command:
ssh cooper '~/.local/bin/orca orchestration worker-start --run "run_9a6a5e637304" … --agent claude …'
Run `run_9a6a5e637304`; outcome `succeeded`. Worktree `/home/gaetan/dev/mc-metarepo/null/mc-metarepo/vecna-post-order-status`, branch `GaetanCathelain/vecna-post-order-status`.
=== WALLCLOCK: 414s ===
```

Tool trace: 16 calls, all `terminal` plus one `browser_navigate` and one
`write_file`. The answer is **right, and independently derived** — it even found
a broken monitor as a bonus. But it was obtained by **delegating a live
production Datadog investigation** — real span queries, real log queries, five
monitor checks, a git worktree and a branch — to re-discover the sentence
`knowledge/datadog-apm-error-sampling.md` already states in 971 bytes:
*"APM ingestion keeps 100% of error spans but samples `ok` spans. As a result,
a raw span search shows massively inflated error ratios."* The note even names
`VecnaApi.postOrderStatus` as its worked example. **Seven minutes and a
production investigation to reproduce a note written about this exact endpoint.**

### Baseline Q3 — 22 s, wrong, and it contradicts house doctrine

```
session_id: 20260807_225812_fd1da1
Yes—rotate it.

"Nobody watched the terminal" doesn't prevent exposure. An inline API key may persist in:

- shell history
- process/audit logs
- agent execution logs
- monitoring/telemetry
- VM snapshots or backups

Treat the key as compromised, revoke it, issue a replacement with minimum permissions, and remove the old command from history/logs where feasible. Deleting traces is cleanup, not a substitute for rotation.
=== WALLCLOCK: 22s ===
```

Tool trace — the only tool call was a failure:

```
$ grep "20260807_225812_fd1da1" ~/.hermes/logs/agent.log | grep tool_executor
… Tool kanban_show returned error (0.00s): {"error": "task_id is required (or set HE…
… tool_turns=1 api_calls=2/500 response_len=423
```

**Zero retrieval.** The headline verdict is right by luck; everything specific
is absent (auth.log + journal, `mc-agents-ssh` readership, rotated copies), and
"remove the old command from history/logs where feasible" is the exact action
the note forbids: *"truncating auth.log / vacuuming the journal erases the audit
trail along with the evidence, and rotated copies remain."* A confident,
fast, plausible, wrong answer is worse than a refusal.

### Baseline Q4 — correct, 113 s, ~42 KB of tool output

Answer (excerpt — it is long and it is good):

```
session_id: 20260807_225816_2a3dac
For a Vecna-specific durable learning, the landing zone is:
`mobile-club/metarepo/learnings/next/vecna.md`
…
6. Run `node scripts/ste-lint.mjs --diff`, push the branch, open a PR against `main`, and add `agent-learning`.
7. A human reviews and merges it. The agent must not push to `main`, merge, or "chase" the PR.
…
- Weak evidence, speculation presented as fact … PR #160 currently has changes requested for exactly that.
- Human scope/value rejection even when mechanical rules pass. PR #182 passed its checks and followed the format, but was closed because the repo owner did not plan to contribute that local-environment knowledge …
=== WALLCLOCK: 113s ===
```

Tool trace — `search_files` failed twice on the SOUL.md path, then it fell back
to shell:

```
Tool search_files returned error (0.06s): {"total_count": 0, "error": "Path not found: /home/gaetan/dev/gaetan-metarepo"}
Tool search_files returned error (0.05s): {"total_count": 0, "error": "Path not found: /home/gaetan/dev/gaetan-metarepo"}
tool terminal completed (0.13s, 491 chars)
Tool terminal returned error (0.13s)
tool terminal completed (1.04s, 5366 chars)
tool terminal completed (0.16s, 126 chars)
tool terminal completed (0.84s, 5728 chars)
tool terminal completed (0.85s, 9934 chars)
tool terminal completed (0.86s, 12265 chars)
tool terminal completed (0.85s, 108 chars)
tool terminal completed (0.45s, 7868 chars)
api_calls=10/500 tool_turns=9
```

Successful `terminal` output totals **41,886 chars**. This is the only baseline
question that was genuinely answered *from the KB*, and it cost 42 KB of
unstructured shell output to get 20 KB of ground truth — because the model was
grepping blind.

### Baseline Q5 — correct, but not from the KB

```
session_id: 20260807_225817_764c60
Start in Vecna MySQL, not Loop.
1. Query `nextmobiles_prd.notification_history`, joined with `notification` …
Never:
- use admin "Regenerate tracking code" on a Next Mobiles return parcel;
- force-create a Loop BR.
Loop gives Vecna the label only once. Regenerating tracking invalidates the old Chronopost reservation without updating Vecna, killing the customer's stored BR link …
Source: `knowledge/vecna-owns-returns.md` and `.claude/skills/triage-return-label-not-delivered/SKILL.md`.
=== WALLCLOCK: 62s ===
```

Correct, and the cited paths are real (I verified
`~/dev/mc-metarepo/.claude/skills/triage-return-label-not-delivered/SKILL.md`
exists). **But the tool trace shows no file was read:**

```
agent.tool_executor: Tool kanban_show returned error (0.00s)…
agent.tool_executor: tool skill_view completed (0.03s, 3207 chars)
agent.tool_executor: tool lcm_recall completed (0.02s, 1786 chars)
agent.tool_executor: tool lcm_grep completed (0.01s, 2183 chars)
agent.tool_executor: tool lcm_expand completed (0.00s, 1529 chars)
agent.tool_executor: tool lcm_load_session completed (0.00s, 13333 chars)
api_calls=9/500 tool_turns=8
```

No `search_files`, no `read_file`, no `terminal`. The answer came out of
**`lcm.db` — the agent's own conversation history** (13.3 KB loaded from a prior
session). It is right today because someone already discussed this; it is not
reproducible on a cold agent, and it will silently go stale when the note
changes, because nothing re-reads the file. **A correct answer sourced from
memory instead of the KB is a false positive for any retrieval design.**

Why Q5 and not Q1/Q3: Gaetan's phrase "bon de retour" is *literally* the KB's
phrase. Vocabulary match is what fires retrieval today. That is the whole
finding.

---

## Scoreboard

`✓` correct and complete · `~` partial · `✗` wrong, absent, or not routable.
Bytes = tool/command output the model must ingest. Seconds for filesystem
mechanisms are the shell command; for agent mechanisms, wall clock of the turn.

| | Q1 NM/MC order | Q2 Datadog 92% | Q3 API-key rotate (vocab mismatch) | Q4 learning→metarepo (3 files) | Q5 bon de retour | **hit rate** |
|---|---|---|---|---|---|---|
| **(a) naive `rg`** | ✗ 337 B / 0.007 s | ✗ 369 B / 0.006 s | ✗ 1419 B / 0.007 s | ~ 15457 B / 0.007 s | ✗ 1349 B / 0.007 s | **0/5** (0.5/5 with partials) |
| **(b) scoped `rg -l` → top file** | ✗ 70476 B | ✗ 6192 B | ✗ 185 B (not listed) | ~ 7107 B | ~ 49904 B | **0/5** |
| **(b′) scoped + `knowledge/`-first** | ✓ 862 B | ✓ 1044 B | ✗ 185 B (not listed) | ~ 7107 B | ✓ 4186 B | **3/5** |
| **(c) skill map → `read_file`** | ✓ **958 B / 19 s** (live) | ✓ ~1.2 KB (not fired) | ✓ **1717 B / 21 s** (live) | ✓ ~21 KB (not fired) | ✓ ~4.2 KB (not fired) | **5/5** (2/5 fired live) |
| **(d) Hermes native, no map = baseline** | ~ 422 s, 25 tools, Orca worker | ✓ 414 s, Orca worker, **prod Datadog** | ✗ 22 s, 0 tools, wrong | ✓ 113 s, 41886 B | ✓ 62 s, from LCM not KB | **3/5 correct, 1/5 from the KB** |
| **(e) filesystem MCP server** | on paper = (b′) | on paper = (b′) | on paper ✗ | on paper ~ | on paper = (b′) | **3/5 on paper**, node absent, adds a dependency, fixes nothing (b′) doesn't |

---

## Context-cost measurement

The number that proves "do not stuff it in the prompt":

| Thing | Bytes | × the winning read |
|---|---|---|
| Whole KB, markdown only (recon B) | **1,478,742** | 861× |
| Four largest files (`schemas/*.full.md` + `next-bdc.fiches.md`) | ~551,000 | 321× |
| One `learnings/` sidecar (`mobileclub/workspace.md`) | 70,381 | 41× |
| `find . -name '*.md'` — the complete path map | 4,289 | 2.5× |
| **The `SKILL.md` map, once per session** | **3,279** | 1.9× |
| Q4 baseline, blind shell grepping | 41,886 | 24× |
| **(c) typical single-fact answer, measured live** | **958–1,717** | **1×** |

- The map is **0.22%** of the KB. A typical answer is **0.10%** of the KB.
- Map + one answer ≈ **5 KB** ⇒ **~0.3%** of the corpus for a complete turn.
- The worst case measured for (c) is Q4 at ~21 KB (three governance files read
  whole) — still **1.4%** of the corpus, and reducible by grepping
  `CONTRIBUTING.md` for the language section instead of reading it whole.
- Against the baseline's actual behaviour the saving is not 100× in bytes, it is
  in **time and blast radius**: 422 s + a git worktree + a delegated Claude Code
  session → **19 s and one `read_file`**.

Stuffing the KB into the prompt was never on the table anyway: 1.44 MB is
roughly 370 K tokens, and the gateway trims at 0.85 of context
(`docs/facts.md`). But the measured cost of *not* stuffing it is ~1 KB per
question, so the design has no tension to resolve.

---

## Verdict

**Ranked:**

1. **(c) skill map → targeted `read_file` — the winner. 5/5, ~1 KB per answer,
   ~20 s per turn, one tool call.** It is also the only mechanism that survives
   the vocabulary-mismatch case, and it needs **zero new machinery**: the map is
   a `SKILL.md` file, and `read_file`/`search_files` are already in
   `_HERMES_CORE_TOOLS` on both CLI and Slack. The mechanism is *routing*, not
   *searching* — the win comes from the repo's slug-per-fact convention doing
   the semantic indexing that a vector store would otherwise do, for free.
2. **(b′) scoped `rg -l` + a `knowledge/`-first heuristic — 3/5, ~1–4 KB.** The
   correct fallback when the map has no slug for the question. Must be
   file-list-first *and* directory-biased; both halves are load-bearing.
3. **(d) what Tars does today — 3/5 correct, but only 1/5 from the knowledge
   base.** This is the number the proposal should lead with, because "3/5
   correct" flatters it badly: two of the three correct answers were
   re-derived from the world (product source; **live production Datadog**) at
   ~7 minutes, one Orca-delegated Claude Code session and one git branch each,
   and the third came from `lcm.db`. Tars is not bad at answering — it is bad
   at *knowing the answer is already written down*. The cost is not bytes, it
   is minutes, branches, and production queries.
4. **(a) naive `rg` — 0/5 as a retriever.** Useful only as the first half of
   (b′). It locates and does not answer.
5. **(b) scoped `rg -l` → read the top file — 0/5.** Actively harmful: ripgrep's
   order is traversal order, so "the top file" is 70 KB of the wrong document.
6. **(e) a filesystem MCP server — ruled out.** `node`/`npx` are absent from the
   VM; it would re-expose `search_files`, which Hermes already has; and the
   team's own `MCP.md` already concluded filesystem+grep with no vector DB.
   Nothing to add.

**Where the winner fails — named, not hedged:**

- **The map must be maintained by hand, and the repo already proves that
  decays.** `knowledge/README.md` is missing 5 of 40 notes. A slug list in a
  `SKILL.md` is the same class of artifact and will rot the same way — a note
  added tomorrow is invisible to Tars until the map is regenerated. **The map
  must be generated from `ls knowledge/*.md`, not written**, and regenerated on
  the same hourly tick as the `git pull` (P6 item 2). This is the single
  highest-risk part of the design and it is cheap to fix.
- **Q4 is where (c) is weakest, and it was not fired live.** Multi-file
  governance questions cost ~21 KB because the map routes to three whole files.
  The map has no mechanism for "read §Language of `CONTRIBUTING.md`" —
  section-level routing is not expressible in a slug list. Baseline Q4 got the
  right answer via 42 KB of blind grepping; (c) would get it via 21 KB of
  targeted reading. That is a 2× saving, not the 24× saving the single-fact
  questions show. **Any P6 claim of "~1 KB per question" is true for
  `knowledge/` notes and false for `learnings/` and governance questions.**
- **Slug routing degrades exactly where the slug vocabulary is thin.** Q3 worked
  because `sudo-argv-secrets-authlog` happens to encode the mechanism. A note
  slugged `done-is-not-exercised` or `ops-list-reconcile-before-bulk` carries
  much less signal, and a question that maps to one of those will fall through
  to (b′) — which is 3/5. **The map's hit rate is bounded by the quality of
  someone else's filenames, in a repo Tars does not control.**
- **I only fired (c) live on 2 of 5 questions.** Q2, Q4 and Q5 under (c) are
  computed from file sizes and my own routing decision, and I authored the map
  after reading the ground truth — so my routing is not blind. The two questions
  I did fire are the two that matter (the 422 s failure and the wrong-answer
  failure), and both were solved by the model itself, not by me. But the 5/5 is
  **2 measured, 3 reasoned**, and should be re-measured before P6 asserts it.
- **The winner does not fix `hindsight`/LCM shadowing.** Baseline Q5 shows the
  agent will happily answer from `lcm.db` without touching the KB. A map that
  says "read the file" does not stop LCM from getting there first with a stale
  copy. Nothing in this trial tested precedence between the two.

**One finding that is not about mechanisms and blocks all of them:** the only
knowledge-base pointer in Tars' SOUL.md is `~/dev/gaetan-metarepo`, which does
not exist on the VM — measured as two consecutive `search_files` "Path not
found" errors in baseline Q4. `~/dev/mc-metarepo` is named nowhere in SOUL.md.
Whatever P6 chooses, that line is the first fix, and it is one sentence.

---

## Not tested:

- **(c) fired live on Q2, Q4, Q5** — only Q1 and Q3 were run through the agent
  with the map. The other three rows in the (c) scoreboard are file sizes plus
  my own routing decision, made after I had read the ground truth. Not blind,
  not measured.
- **Baseline Q2's production figures were not independently verified.** Its
  worker reported 594 OK spans / 1 error over 15 days and a broken monitor
  `110136431`; I did not re-run those Datadog queries. The verdict ("sampling
  artifact, not an incident") matches the KB note, so it is scored correct, but
  the supporting numbers are the agent's, not mine.
- **The map as a real installed skill.** It was prepended to the prompt, not
  written to `~/.hermes/skills/` (forbidden — peers are editing that tree).
  Whether Hermes' skill auto-injection actually fires this skill on these
  questions — i.e. whether the *description* frontmatter is good enough to get
  the map into context unprompted — is **untested and is a real risk**: the map
  only helps if it loads. `docs/facts.md` says skills load "when relevant"; that
  relevance judgement was not exercised.
- **Any Slack-surface turn.** All 7 agent turns went through the `cli` platform
  (`platform=cli` in `turn_context`). Recon A established the toolsets are
  identical by source, but no question was asked over Slack.
- **Whether `search_files` honours `.gitignore`**, and its exact behaviour on
  the 152 KB / 156 KB files — recon A flagged this as unresolved and I did not
  resolve it either.
- **Hourly-refresh failure behaviour (P6 item 2)** — out of scope here; no
  `git fetch`/`pull` was run (banned), so nothing about staleness detection or
  auth-lapse behaviour was measured in this probe.
- **Precedence between LCM recall and file retrieval.** Baseline Q5 answered
  from `lcm.db`; I did not test whether a map-guided turn would still prefer LCM
  over `read_file`, or how a stale LCM copy competes with a fresh file.
- **Questions routing to `schemas/`, `index/` or the pinned `<stack>/` docs.**
  All 5 questions land in `knowledge/` or governance files. The map claims
  routes for `schemas/*.core.md` and `index/*.json`; those routes are unexercised.
- **Cold-cache timings.** Every `rg` ran against a warm page cache on a 2.2 MB
  tree; the 6–8 ms figures are lower bounds, though the tree is small enough
  that this cannot change any ranking.
