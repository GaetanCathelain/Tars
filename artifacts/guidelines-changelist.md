# WF5 — Tars redesign change-list (decision-ready)

Synthesis of the four WF5 inventory probes (`status/probes/wf5/guidelines-{soul,skills,repo-docs,runtime}.md`).
Decision it implements (Gaetan, 2026-08-07): **Tars is an orchestrator and secretary of work, a
mirror of how Gaetan himself works — not a restricted assistant.** Restraint from implementing
stays as *separation of concerns*; every framing of that restraint as *security confinement*
(sudo prohibition, command allowlist, path restriction, sandbox) is dropped.

Nothing here has been applied. This session was read-only on the VM and wrote only the two files
under `artifacts/`. No commit, no push.

**Legend** — REFRAME (text keeps its rule, justification changes) · DROP (confinement framing or
control removed) · STALE-REWRITE (describes the dropped v1 design as current) · ADD (new text).

**Apply mechanics** — every `~/.hermes/*` change: `.bak` copy first, then
`flock ~/.hermes/.wf3.lock -c '<edit>'`, merge never append. **No change below needs a gateway
restart**: `SOUL.md` is read fresh at every `_build_system_prompt()` and `config.yaml` is
mtime-cached and re-read per turn (`hermes_cli/config.py:236-249`) — ~30 s to take effect. A
sibling session is live on that gateway; do not restart it.

**Change counts:** REFRAME 6 · ADD 4 · DROP 1 · STALE-REWRITE 1 → **12 changes to apply**, plus
**2 files flagged for the other workstream** (`delegate-to-cooper/SKILL.md`,
`docs/specs/wf5-orca-delegation.md`) for which no replacement text is authored here.

---

## A. `~/.hermes/SOUL.md` (VM, highest blast radius — prompt slot #1, every turn, every surface)

Needs **`.bak` + `flock`**. No restart. Complete replacement text: `artifacts/SOUL-proposed.md`
(install whole-file; the entries below are the diff a reviewer should see).

### A1 — L5–6, intro: name the job (ADD)

CURRENT
```
Gaetan asks me where things stand and what happens next. I dispatch work to
other agents and machines, follow it, and report back.
```
PROPOSED
```
Gaetan asks me where things stand and what happens next. I dispatch work to
other agents and machines, follow it, and report back. I am the secretary of his
work: I hold the queue, I schedule it, I write the briefs, I prompt the agents
that do the work, I track them, I double-check the facts they report, and I tell
Gaetan where things stand. That is a mirror of how he works himself — he prompts
coding agents and judges the result; I do it in his place, with his access.
```
Rationale: Gaetan's decision adds scheduling, real Orca-session driving and "a copy of me" to the
job description; the old two lines were compatible with it but said none of it.
Risk if applied: SOUL.md grows ~20 lines; `docs/specs/tars-profile.md:18` claims SOUL is "truncated
if long" (recon claim, **never measured** — see §E1). Risk if not applied: Tars keeps declining
scheduling/driving work as out of character.

### A2 — L12–13, hard rule 1: say *why* (REFRAME)

CURRENT
```
1. I never write code. No implementation, no patch, no script, no config file —
   not in a message, not to disk, not "just as an example".
```
PROPOSED
```
1. I never write code. No implementation, no patch, no script, no config file —
   not in a message, not to disk, not "just as an example". Not because I am not
   trusted with it: writing the implementation is the coding agent's job. Mine is
   the brief, the tracking, the verification and the verdict.
```
Rationale: the rule survives verbatim; only its justification moves from implied distrust to
job scope, exactly as the decision requires.
Risk if applied: none substantive — two sentences, same prohibition. Risk if not applied: an
unjustified prohibition sitting next to newly-granted root reads as a security rule, and the model
generalises it into refusing sudo and other powerful-but-on-mission actions.

### A3 — new hard rule 6: credentials (ADD)

CURRENT: absent from SOUL.md (the clause lives only in `delegate-to-cooper/SKILL.md:50-52`, a file
the other workstream is rewriting).
PROPOSED
```
6. I never read, print, echo or pass along a credential, token or key — not in a
   message, not into a file, not on a command line. If work needs a secret, I name
   which one and let Gaetan or the agent that owns it supply it.
```
Classification: ADD (promotion of a LOAD-BEARING rule to a durable surface).
Rationale: secret-handling is explicitly out of scope of the redesign and must survive — it must
not survive *only* inside the one skill file being retired.
Risk if applied: duplication with whatever the sibling skill rewrite keeps (harmless; needs
Gaetan's nod — §E3). Risk if not applied: the rule can vanish silently with the skill rewrite,
right as Tars gains real machine access.

### A4 — new hard rule 7: cooper yes, other machines no (ADD)

CURRENT: absent from SOUL.md (nearest equivalent is `delegate-to-cooper/SKILL.md:53`, "Never touch
`192.168.0.3` / pve, and never the p-Hermes host").
PROPOSED
```
7. Cooper is mine to act on with Gaetan's own access, sudo included — standing in
   for him is the point. Other machines are not: p-Hermes is read-only to me, and
   the pve hypervisor at 192.168.0.3 is not mine to touch.
```
Classification: ADD (states the granted power positively **and** keeps the LOAD-BEARing
non-cooper boundary that would otherwise be lost with the skill).
Rationale: Gaetan's "it's ok if he even has root on cooper, it's by design" needs to be *stated*,
or the model supplies its own confinement; the second sentence keeps the only machine boundary the
decision did not drop.
Risk if applied: an empowerment sentence in a hard-rules list is unusual and reads oddly out of
context. Risk if not applied: Tars refuses sudo on cooper from trained caution (no config change
fixes that — see A/B split with B1), or over-generalises the new freedom onto p-Hermes/pve.

### A5 — new marked section: Phase 2 (ADD)

CURRENT: absent.
PROPOSED (see `artifacts/SOUL-proposed.md`, section `## Phase 2 — Gaetan's knowledge and
preferences`, opening with `<!-- PLACEHOLDER, not yet written. Do not invent content here. -->`).
Rationale: Gaetan's decision defers his knowledge/preferences corpus to Phase 2 — the slot is
reserved, its source (`~/dev/gaetan-metarepo`) and its live mechanism
(`~/.hermes/memories/USER.md`, already enabled and populated) named, no content invented.
Risk if applied: a placeholder in the live prompt costs ~8 lines of context. Risk if not applied:
Phase 2 lands as an ad-hoc edit with no agreed home, or someone fills it by guessing.

**Not changed in SOUL.md:** rules 2, 3, 4, 5, `## Tone`, `## Language`, and the `# Tars` /
"I live on Slack" opening — see §D.

---

## B. `~/.hermes/config.yaml` (VM, runtime — the only behavioural change in this list)

Needs **`.bak` + `flock`**. No restart (mtime-cache re-read per turn).

### B1 — `approvals.mode` (DROP)

CURRENT: **no `approvals:` key exists in the file at all** — so the code default applies:
```
# tools/approval.py:2936-2939  →  _get_approval_mode() → cfg.get("mode", "manual")
# effective: approvals.mode = "manual"
# DANGEROUS_PATTERNS (approval.py:381-516) explicitly matches `sudo` (with flags)
#   and blocks `sudo -S` stdin password guessing.
```
PROPOSED (merge as a new top-level block, do not reorder the file):
```yaml
approvals:
  mode: off   # Tars acts with Gaetan's own access on cooper by design (WF5 redesign,
              # 2026-08-07). Separation of concerns, not confinement, is the guardrail.
```
Rationale: this — not any doc sentence — is the actual sudo guardrail on Tars today; Gaetan said
to drop the sudo prohibition, and prose alone cannot.
Risk if applied: **highest-risk change in this list.** `mode: off` removes the approval gate for
*every* dangerous pattern, not only `sudo` — `rm -rf`, disk writes, `dd`, `curl | sh` included, on
a box where `terminal.backend: local` means commands run directly on the host. It applies to a
gateway a sibling session is currently using. Risk if not applied: Tars hits an approval prompt on
sudo-class commands with no interactive human on the Slack path, and the redesign's "root on
cooper by design" is text-only.
**Needs Gaetan's explicit ruling — see §E2.** Two options: `mode: off` (full drop) or leave the
default and route approvals to Slack (Gaetan approves in-band, which is arguably what "mirror of
me" means anyway).

**Not changed in config.yaml:** `gateway.platforms.slack.require_mention: true`,
`strict_mention: true`, `unauthorized_dm_behavior: ignore` — see §D.

---

## C. Skills (`~/.hermes/skills/`) — nothing of ours to change

The tree-wide grep across 76 skill packages found **exactly one** file carrying redesign-relevant
rules, and it belongs to another workstream.

### C1 — `~/.hermes/skills/delegate-to-cooper/SKILL.md` (STALE-REWRITE — **NOT OURS, no text authored**)

Owner: the WF5 Orca-delegation workstream. Flagged content (verbatim, from the skills inventory):
L15 "This is how SOUL rule 3 actually gets done."; L20 `ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh"`;
L33-40 the 4-command allowlist; L44-46 "`--dangerously-skip-permissions` and friends are not mine
to type."; **L47 "Never `rm`, `rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper." (DROP — the
one literal sudo prohibition anywhere in Tars's authored text)**; L54-56 the
`~/orca/workspaces/tars-delegated/` path restriction; L57-58 "Never edit `delegate.sh` or
`.claude/settings.json` — those are the guardrails"; L75-76 the `DENIED` deny-rule protocol.
What it implies for **our** files: L50-52 (credentials) and L53 (pve/p-Hermes) are LOAD-BEARING and
must not die with this file — that is why A3 and A4 promote them into SOUL.md. If the rewrite keeps
them too, the duplication is harmless.
Risk if that file is left as-is: SOUL.md and the skill contradict each other in the same context
window (SOUL grants sudo on cooper at band 1, the skill forbids it at band 3) — the model will pick
one unpredictably. **This is the sequencing dependency between the two workstreams.**

### C2 — `autonomous-ai-agents/hermes-agent/references/security-privacy.md:54-62` (no change)

Vendor doc describing Hermes's own `command_allowlist` / approvals mechanism. Synced content —
editing it gets clobbered by the next bundle sync. Informational only; the live setting it
describes is handled by B1.

No other skill in the tree references `cooper`, `tars`, `tars-delegated`, `delegate.sh` or
`SOUL rule` — retiring the v1 design touches nothing else.

---

## D. Repo docs (lowest blast radius — not in any prompt; no flock, no restart)

### D1 — `README.md:8-9` (REFRAME)

CURRENT
```
Tars orchestrates and reports; it never implements. Coding work is delegated to cooper/Orca.
No PR is ever created by Tars.
```
PROPOSED
```
Tars orchestrates and reports; it never implements — that is the coding agent's job, not a limit
on what Tars is trusted with. Coding work is delegated to a real Orca session on cooper, which
Tars drives, tracks and verifies. No PR is ever created by Tars.
```
Rationale: same rule, justification made explicit, and "cooper/Orca" sharpened to the real
mechanism the redesign grants. Risk if applied: none (README). Risk if not applied: the repo's
front door keeps implying confinement.

### D2 — `PLAN.md:201-203` (REFRAME)

CURRENT
```
- **SOUL.md v1: minimal + bilingual** — identity, the hard rules (no code, no PRs,
  orchestrate/report only, answer only Gaetan), concise tone, mirror the incoming language
  (FR/EN). Behaviors grow in WF5. WF4 probe set confirmed as planned (14 + critic).
```
PROPOSED
```
- **SOUL.md v1: minimal + bilingual** — identity, the hard rules (no code, no PRs,
  orchestrate/report only, answer only Gaetan), concise tone, mirror the incoming language
  (FR/EN). Behaviors grow in WF5. WF4 probe set confirmed as planned (14 + critic).
- **SOUL.md v2 (WF5 redesign, 2026-08-07):** Tars is the secretary of Gaetan's work and a mirror
  of how he works — it schedules, briefs, drives real Orca sessions on cooper, tracks, verifies
  and reports. "No code, no PRs" stays, restated as separation of concerns, not confinement.
  Credentials and the non-cooper machine boundary promoted into SOUL; sudo prohibition dropped.
```
Rationale: PLAN §Amendments is where settled decisions land; this one is settled. Additive so the
v1 record stays intact. Risk if applied: none. Risk if not applied: the plan of record contradicts
the running agent.

### D3 — `PLAN.md:216-217` (REFRAME)

CURRENT
```
- Tars never opens PRs, never implements — enforced later in its profile/system prompt (WF2
  scaffolds it, WF4 does not probe coding on purpose).
```
PROPOSED
```
- Tars never opens PRs, never implements — a separation of concerns, not a security confinement;
  stated in its profile/system prompt (WF2 scaffolds it, WF4 does not probe coding on purpose).
```
Rationale: this line sits under a `## Guardrails` heading whose other entries are genuine security
controls; without the qualifier it reads as one of them. Risk if applied: none. Risk if not
applied: the strongest remaining source of the wrong framing in the repo.

### D4 — `CLAUDE.md:9-10` (REFRAME)

CURRENT
```
Tars orchestrates and reports; it never implements — coding work is delegated
to cooper/Orca, and Tars never creates a PR.
```
PROPOSED
```
Tars orchestrates and reports; it never implements — that is the coding agent's
job, not a trust limit. Coding work is delegated to a real Orca session on
cooper that Tars drives and verifies, and Tars never creates a PR.
```
Rationale: this file is loaded by every Claude Code session in this repo; it is the highest-leverage
repo-side sentence. Risk if applied: none. Risk if not applied: every future session inherits the
old framing.

### D5 — `docs/specs/tars-profile.md:23-56` + `:20` + `:63-65` (STALE-REWRITE)

CURRENT: lines 23-56 are a fenced copy of SOUL v1, verbatim identical to the live file; line 20 says
"v1 is deliberately minimal — behaviors … are WF5's job and are *not* pre-declared here"; lines
63-64 say "Rules 1–3 are the PLAN.md guardrail"; line 65 says "Keep it at this length. Growing it is
WF5's amendment, with a diff, not a drive-by edit."
PROPOSED: replace the fenced block with the exact content of `artifacts/SOUL-proposed.md`; append to
line 20 "(v2, WF5 2026-08-07, supersedes the block below — see PLAN §Amendments)"; change 63-64 to
"Rules 1–3 are the PLAN.md separation-of-concerns rule 'Tars never opens PRs, never implements' —
job scope, not confinement." Line 65 needs no edit: this *is* the WF5 amendment, with a diff.
Rationale: this spec is the documented mirror of the live file; leaving it at v1 makes the repo's
own spec the most convincing wrong source. Risk if applied: two copies of SOUL text to keep in sync
(already the case). Risk if not applied: the next session reinstalls v1 from the spec.

### D6 — `docs/specs/wf5-orca-delegation.md` (STALE-REWRITE — **NOT OURS, no text authored**)

Self-describes the dropped v1 design ("Dedicated sandbox. Not a repo, not a checkout, no git
remote… `delegate.sh` — the only entrypoint Tars may call… deny-only permission rules…") as
"Authoritative spec. Status: **implemented and verified 2026-08-07**". Owned by the Orca-delegation
workstream. Flagged only. Risk if left: it is the most authoritative-looking stale document in the
repo and reads as current shipped fact.

### D7 — `docs/facts.md` (NO CHANGE, deliberately)

A sibling session is appending to it concurrently, and its "SOUL / persona design" section is
LOAD-BEARING. One factual correction is owed to its owner, not applied here: the minimal reply
character is `·` (U+00B7 MIDDLE DOT), not `.` — verified byte-for-byte against the live
`~/.hermes/SOUL.md` (`c2 b7`) and reproduced exactly in `artifacts/SOUL-proposed.md`.

---

## E. LOAD-BEARING rules NOT touched — and why

Nothing protective was removed. Every item below was inspected and deliberately left alone.

| Rule | Where it lives | Why untouched |
|---|---|---|
| **Slack identity frame, hard rule 4** — "I answer Gaetan and no one else. Gaetan is Slack user U08BDJAMSRZ — … sender prefix reads \"[U08BDJAMSRZ \| …]\" … I reply with the single character \"·\" and nothing else" | `~/.hermes/SOUL.md:17-20` → `SOUL-proposed.md` rule 4, **byte-identical**, incl. U+00B7 | Names Gaetan's platform identity per surface **and** prescribes a minimal literal reply instead of silence. Hermes has **no silence terminal**: deliberate model silence becomes N empty retries → 61 s stall → user-facing error. Weakening this caused a live incident. With confinement dropped, it is also the model-layer half of the only trust boundary left. |
| **`SLACK_ALLOWED_USERS`** (set, `.env`), `require_mention: true`, `strict_mention: true`, `unauthorized_dm_behavior: ignore` (config.yaml) — adapter-level early reject | VM `.env` + `config.yaml` | This is the *real* enforcement; rule 4 is defence in depth. Post-redesign anything that can message Tars can transitively drive Orca on cooper, so this control matters **more**, not less. Not touched, not weakened, not made conditional. |
| **Secrets handling** — never whole-file `sops -d`, per-key `--extract` only, never a secret on argv/echoed/written | repo `CLAUDE.md:49-56`, `PLAN.md:211-212`, skill L50-52 | About secrets, not about Tars's autonomy. Kept, and **strengthened**: promoted into SOUL as hard rule 6 (A3) so it survives the skill rewrite. |
| **Non-cooper machines** — p-Hermes read-only, never root on `192.168.0.3`/pve, gated old-profile delete needs Gaetan's explicit go | repo `CLAUDE.md:60`, `PLAN.md:213-215`, `wf3-wiring.md:279-282`, skill L53 | The decision grants root on **cooper**, explicitly by analogy to Gaetan's own access there. It says nothing about other machines. Kept, and **strengthened**: promoted into SOUL as hard rule 7 (A4). |
| **`.bak` + `flock ~/.hermes/.wf3.lock` on every VM config/SOUL edit; merge never append** | repo `CLAUDE.md` | Concurrency protection against sibling sessions, unrelated to the redesign. Every change in §A and §B is marked as requiring it. |
| **Hard rules 2, 3, 5** — no PR/push; orchestrate-and-report; refuse-in-one-line-and-offer-delegation | `SOUL.md:14,15-16,21-22` → `SOUL-proposed.md`, verbatim | Pure separation of concerns; the decision explicitly keeps them. Rule 5's "rules 1–3" reference stays valid because numbering was preserved and additions were appended as 6 and 7. |
| **Tone / Language sections** | `SOUL.md:24-32` → `SOUL-proposed.md`, verbatim | Unrelated to the redesign. |
| **D1–D6** (`docs/recon/DECISION.md`) | repo | None assert sudo/root confinement on Tars; D2's only "root" is the orchestrator session's access to p-Hermes/pve. No amendment needed. |

---

## F. Open questions Gaetan's decision did not anticipate

1. **E1 — SOUL truncation is an unverified recon claim.** `tars-profile.md:18` says SOUL.md "is
   truncated if long"; nothing in the VM source read this session confirms a length cap
   (`load_soul_md`, `prompt_builder.py:2005-2031`, reads the file whole). The proposed file is
   ~55 lines vs 32. Cheap to verify before installing; not verified here (read-only session).
2. **E2 — `approvals.mode: off` is broader than "drop the sudo rule".** It removes the gate for
   every dangerous pattern on a host where `terminal.backend: local`. Needs Gaetan's explicit
   ruling: full `off`, or keep the default gate and route approval prompts to Slack so he answers
   them in-band.
3. **E3 — promoting the credentials + non-cooper clauses into SOUL** (A3/A4) duplicates text the
   sibling skill rewrite may also keep. Harmless, but it is a judgement call made here, not one
   Gaetan asked for.
4. **E4 — "help me schedule my work" has no model-facing tool.** No `cron` toolset exists in
   `toolsets` or the tool registry; `hermes cron` is a CLI command. Today Tars can only schedule by
   shelling out through `terminal()`. Either that is accepted, or a scheduling toolset is a WF5
   work item.
5. **E5 — Phase 2's landing surface is self-curated, not authored.** `~/.hermes/memories/USER.md`
   exists, is enabled and already holds 2 model-written entries — but it is the memory tool's own
   store (compacted under `user_char_limit: 1375`), so hand-seeding Gaetan's preferences there will
   be rewritten. A static surface (`~/.hermes/preferences/`, a valid per-profile path) does not
   exist yet. Phase 2 needs that mechanism decision before content.
6. **E6 — sequencing, and a gap the sibling draft leaves.** §A (SOUL grants sudo on cooper) and
   §C1 (live skill still forbids it, L47) must land together, or the model reads two contradictory
   rules in one context window. Checked the sibling's replacement draft
   (`artifacts/delegate-to-cooper-SKILL.md`): it is consistent on framing — "This is a division of
   labour, **not** a security boundary. There is no sandbox, no command allowlist, no forbidden
   path list, no wrapper script" (L37-38) — but it contains **no credentials clause and no
   pve/p-Hermes clause** (grep: zero hits). Both LOAD-BEARING clauses die with the v1 skill unless
   A3 and A4 land. That makes A3/A4 not optional.
