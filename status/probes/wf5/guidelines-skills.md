# Surface 2 inventory — Tars's skills tree (`~/.hermes/skills/` on the VM)

Read-only, over `ssh gaetan@192.168.0.9`. Nothing on the VM was changed —
every command below is a list/grep/cat/`hermes skills list`. No `sops -d`,
no config edit, no gateway restart.

## 1. Tree summary

```
~/.hermes/skills/
├── apple/               (4 skills: apple-notes, apple-reminders, findmy, imessage)
├── autonomous-ai-agents/ (5: claude-code, codex, computer-use, hermes-agent, opencode)
├── creative/             (~18: architecture-diagram, ascii-art, ascii-video,
│                           baoyu-infographic, claude-design, comfyui, design-md,
│                           excalidraw, humanizer, manim-video, meme-generator,
│                           p5js, popular-web-designs, pretext, sketch,
│                           songwriting-and-ai-music, touchdesigner-mcp, …)
├── delegate-to-cooper/   (1 — LOCAL, no sub-category)
├── email/                (himalaya, …)
├── github/               (github-auth, github-code-review, github-issues,
│                           github-pr-workflow, github-repo-management)
├── hermes-operations/    (2 — LOCAL: hermes-skill-portability, skill-library-auditing)
├── i-have-adhd/          (1 — community-installed, no sub-category)
├── media/, mlops/, note-taking/, productivity/, research/, smart-home/,
│   social-media/, software-development/  (remaining vendor categories)
├── .bundled_manifest     (70 name:md5 lines — the vendor/bundle sync ledger)
└── .hub/                 (lock.json, taps.json, scan-cache/, index-cache/, audit.log —
                            skill-hub package-manager state, not model-facing content)
```

Totals: 674 files under the tree; `find -name SKILL.md` → 76 skill directories;
`hermes skills list` (live, read-only) reports **74** registered skills
(the "roughly 73" in the task — confirmed).

## 2. Locally-authored vs synced/vendor — evidence

Three independent signals agree:

1. **`.bundled_manifest` diff.** It lists 70 `name:md5` pairs — the vendor
   bundle's own integrity ledger. `comm -23` between every `SKILL.md` parent-dir
   name and this list leaves exactly 6 directories unaccounted for:
   `delegate-to-cooper`, `i-have-adhd`, `hermes-skill-portability`,
   `skill-library-auditing`, `last30days`, `meme-generator`.

2. **mtimes.** Every vendor category dir and its children share one batch
   timestamp, `2026-08-07 16:17:59–16:19:08` (the sync that laid the bundle
   down). The 6 outliers split into two groups:
   - **Built today, after the sync** — `delegate-to-cooper` (20:21),
     `hermes-operations/hermes-skill-portability` (21:13),
     `hermes-operations/skill-library-auditing` (21:20). These are genuinely
     hand-authored for this Tars build.
   - **Pre-date the sync entirely** — `research/last30days` (2026-07-08),
     `creative/meme-generator` (2026-07-31). Pre-existing personal skills,
     untouched by the bundle sync (additive, not overwritten). No
     redesign-relevant content in either (verified by grep, §3).

3. **`hermes skills list` source/category columns** (live, read-only —
   confirms rather than contradicts #1/#2, with one correction):

   | name | category | source | provenance |
   |---|---|---|---|
   | delegate-to-cooper | local | **local** | hand-authored |
   | hermes-skill-portability | hermes-operations | **local** | hand-authored |
   | skill-library-auditing | hermes-operations | **local** | hand-authored |
   | i-have-adhd | community | **skills.sh** | **not** hand-authored — installed from the community skill registry. `~/.hermes/skills/.hub/audit.log` confirms: `2026-08-07T16:32:30Z INSTALL i-have-adhd skills.sh:community safe sha256:6d1514c3f8…`. Its recency and manifest-absence made it look local; it isn't. |
   | meme-generator | creative | local | pre-existing, not part of this build |
   | last30days | research | local | pre-existing, not part of this build |

**Correction to the task's own example**: `i-have-adhd` is *not* a good example
of "locally authored" — it's a community skill pulled from `skills.sh` the
same day. `delegate-to-cooper` is the clean example of hand-authored.

## 3. Grep hits for redesign fingerprints — full table

Pattern run: `sudo|\broot\b|never run|must not|forbidden|not allowed|allowlist|
denylist|\bdeny\b|sandbox|tars-delegated|delegate\.sh|write code|never write|
implement|confine|restrict|permission`, tree-wide first (530 raw hits — almost
all noise: LaTeX conference-template boilerplate in
`research/research-paper-writing/`, generic code comments in `research/last30days/`,
docx/pptx symlink-safety code, `google-workspace` admin-allowlist help text,
`computer-use`'s click-idempotency note). Narrowed to hits that are actually
about Tars/cooper/security-of-Tars:

| file:line | verbatim | classification | editable? |
|---|---|---|---|
| `delegate-to-cooper/SKILL.md:15` | "This is how SOUL rule 3 actually gets done." | **STALE** | local — but **another workstream owns this file; do not edit** |
| `delegate-to-cooper/SKILL.md:20` | `ssh cooper "~/orca/workspaces/tars-delegated/delegate.sh" <<'BRIEF'` | **STALE** — the dropped v1 sandbox path/wrapper | local, not ours |
| `delegate-to-cooper/SKILL.md:33-40` | "Everything I am allowed to run on cooper" — 4-item allowlist | **STALE→REFRAME** (redesign grants real Orca sessions, not a 4-command allowlist) | local, not ours |
| `delegate-to-cooper/SKILL.md:44-46` | "Only `delegate.sh`, which fixes them. `--dangerously-skip-permissions` and friends are not mine to type." | **STALE** | local, not ours |
| `delegate-to-cooper/SKILL.md:47` | "Never `rm`, `rmdir`, `mv`, `chmod`, `sudo`, `systemctl` on cooper." | **DROP** — textbook security-confinement framing the redesign explicitly kills | local, not ours |
| `delegate-to-cooper/SKILL.md:48-49` | "Never `git push`... Never open, review or merge a pull request (SOUL rule 2)." | **KEEP-AS-IS / REFRAME** — git-write and PR restraint is mission-scope ("Tars never creates a PR" per this repo's own CLAUDE.md), not a trust guardrail — but justification line needs rewording once owned | local, not ours |
| `delegate-to-cooper/SKILL.md:50-52` | "Never read, print or pass along credentials…" | **LOAD-BEARING** — keep verbatim, this is the secrets rule that must survive | local, not ours |
| `delegate-to-cooper/SKILL.md:53` | "Never touch `192.168.0.3` / pve, and never the p-Hermes host." | **LOAD-BEARING** — explicitly named as must-survive in the task brief | local, not ours |
| `delegate-to-cooper/SKILL.md:54-56` | "Never a path outside `~/orca/workspaces/tars-delegated/`…" | **STALE** — that sandbox is dropped; redesign works "in a worktree of a registered repo" | local, not ours |
| `delegate-to-cooper/SKILL.md:57-58` | "Never edit `delegate.sh` or `.claude/settings.json` — those are the guardrails…" | **STALE** — both files no longer exist as designed | local, not ours |
| `delegate-to-cooper/SKILL.md:59` | "Never write the code myself (SOUL rule 1). I write the brief. That is all." | **KEEP-AS-IS / REFRAME** — the one rule the redesign explicitly preserves; only its justification changes from "can't be trusted" to "not my job" | local, not ours |
| `delegate-to-cooper/SKILL.md:75-76` | "The agent answers `DENIED` — a guardrail fired... do not look for a way around it." | **STALE** — refers to the dropped `.claude/settings.json` deny-rule mechanism | local, not ours |
| `autonomous-ai-agents/hermes-agent/references/security-privacy.md:54-62` | "Shell-command allowlist: `hermes config set command_allowlist '[]'`" / "Shell hooks allowlist" | **KEEP-AS-IS** — vendor doc describing a *generic Hermes platform feature* (the shell-command approval/YOLO-mode gate), not a Tars-authored rule. Informational only: if this `command_allowlist` or `approvals.mode` is non-`off` in Tars's own `config.yaml`, sudo-class commands would still hit an approval prompt regardless of skill text — **flag to the orchestrator as a config check, out of scope for this skills-only pass** | **synced/vendor** — editing risks being clobbered by the next `hermes-agent` bundle sync |
| `i-have-adhd/SKILL.md:111,113,115` | "Forbidden openers…", "Forbidden recaps…", "Forbidden closers…" | **N/A** — communication-style guardrails (banned filler phrases), unrelated to the redesign | community (`skills.sh`) — do not edit, next `skills update` overwrites |
| `hermes-operations/skill-library-auditing/SKILL.md:41` | "Operate read-only unless the user separately authorizes migration or repair." | **N/A** — generic tool-safety default for a skill-*auditing* meta-skill, not a Tars-autonomy rule | local, editable, but out of scope (different subject) |
| `autonomous-ai-agents/computer-use/SKILL.md:156` | "A confirmed effect is done and must not be duplicated." | **N/A** — UI-click idempotency note, false positive | synced/vendor |
| `research/research-paper-writing/templates/**` (15+ hits) | conference-submission style rules ("Authors must not…", "confined within a rule…") | **N/A** — LaTeX template boilerplate, false positive | synced/vendor |
| `research/last30days/**`, `productivity/docx`, `productivity/powerpoint`, `creative/popular-web-designs/templates/apple.md`, `productivity/google-workspace/SKILL.md:311` | assorted code comments / admin-allowlist help text | **N/A** — false positives, unrelated subject matter | synced/vendor |

**Net count: 1 file carries real redesign-relevant rules — `delegate-to-cooper/SKILL.md`** (11 of its ~78 lines hit the grep, 9 of those are DROP/STALE, 2 are LOAD-BEARING/KEEP that must survive whatever replaces it). Everything else that matched is either a vendor false-positive or the one generic vendor doc about Hermes's own shell-approval mechanism (informational, not a rule to edit).

## 4. `delegate-to-cooper/SKILL.md` — inventory only, not touched

Per the mandate, this file (and `docs/specs/wf5-orca-delegation.md`) belong to
a separate workstream — no replacement text authored here. What it implies
for other files, for that workstream to pick up:

- It is the *only* skill file that encodes the dropped v1 design end-to-end:
  the `tars-delegated/` sandbox path, the `delegate.sh` wrapper, the
  `.claude/settings.json` deny-rule reference, and a hardcoded 4-command
  allowlist. All of it is STALE against the "real Orca sessions in a worktree
  of a registered repo" redesign.
- Two clauses in it are explicitly load-bearing and must carry forward
  unchanged into whatever replaces it: the credentials never-read list
  (lines 50-52) and the p-Hermes/pve no-touch rule (line 53) — both named
  verbatim in this task's "must survive" list.
- One clause changes justification, not substance: "never write the code
  myself" (line 59) stays as a rule, but its *reason* moves from
  "guardrail I can't be trusted to bypass" to "not my job, that's cooper's."
- No other skill file references `delegate-to-cooper`, `cooper`, `tars`,
  `SOUL rule`, `tars-delegated`, or `delegate.sh` — confirmed by tree-wide
  grep (§3). The blast radius of retiring this file is contained to itself;
  nothing else in the skills tree needs touching as a consequence.

## 5. How skills reach model context, and reload mechanics

**Injection.** Skills are injected into the model's context the same way
`AGENTS.md`/`SOUL.md`/memory are — automatically, per session — per
`hermes-agent/references/cli-reference.md:22`: `--ignore-rules  Skip
AGENTS.md/SOUL.md/memory/skill injection`. A skill can also be force-loaded
for one session via `--skills/-s NAME` (preload) or invoked mid-session by
its own `/<name>` slash command (`cli-reference.md:18`, `:57-62`).

**Discovery/index.** `~/.hermes/skills/.hub/index-cache/hermes-index.json`
(generated 2026-07-20, 90,605 entries) is the **community catalog** used for
`hermes skills browse/search` — it is not the live "what's loaded" registry.
The live registry is filesystem-backed: Hermes scans `~/.hermes/skills/`
directly, matching each `SKILL.md`'s frontmatter `name`. There is no separate
"model-facing index" file to go stale — the source of truth is the directory
tree itself.

**Invalidation / pickup.** Two paths, per
`hermes-operations/hermes-skill-portability/references/migration-checklist.md:63`
and `autonomous-ai-agents/hermes-agent/references/slash-commands.md:73`:
- A **running** Hermes process: send `/reload-skills` ("Re-scan skills
  directory") to the live session.
- A **new** Hermes process/session: scans the directory fresh on start — no
  explicit reload needed.

**Gateway restart: NOT required either way.** Both mechanisms above operate
without restarting `hermes-gateway.service`; `/restart` in the slash-command
reference is listed separately, scoped to "Restart gateway after draining
active runs," and is not documented as a precondition for skill pickup.

**Proof command (read-only, actually run against the live VM this session):**

```
$ ssh gaetan@192.168.0.9 'export XDG_RUNTIME_DIR=/run/user/$(id -u); \
    ~/.local/bin/hermes skills list' | grep -iE "delegate-to-cooper|hermes-operations|i-have-adhd"

│ delegate-to-cooper   │                     │ local     │ local     │ enabled │
│ i-have-adhd          │                     │ skills.sh │ community │ enabled │
│ hermes-skill-portab… │ hermes-operations   │ local     │ local     │ enabled │
│ skill-library-audit… │ hermes-operations   │ local     │ local     │ enabled │
```

This confirms presence/source/category/enabled status directly from the live
process — the CLI-level way to prove a *new or renamed* skill was picked up.
For proving a *content edit* inside an existing `SKILL.md` was picked up
(not tested here — would require sending `/reload-skills` to the live
session, which is a gateway interaction, not read-only, so out of scope for
this probe), the documented check is `hermes skills check [name]`
(validates package structure) followed by a behavioral check that the model's
next reply reflects the new text.
