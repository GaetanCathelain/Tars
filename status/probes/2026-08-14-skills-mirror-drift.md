# Skills mirror drift — VM vs repo (2026-08-14)

Read-only inventory. No `sops -d`, no config.yaml/.env reads, no VM writes. Source: `ssh gaetan@192.168.0.9` (`find` + `md5sum` over `~/.hermes/skills/`, VM clock UTC) vs this repo's `skills/` mirror (`find` + `md5sum` + `git log -1 --format=%cI`).

## Classification counts

| Class | Count |
|---|---|
| Total live SKILL.md dirs on VM | 132 |
| Total SKILL.md dirs mirrored in repo | 11 |
| MIRRORED-IDENTICAL (md5 match, all mirrored files) | 11 |
| MIRRORED-DIVERGED (md5 differs) | 0 |
| REPO-ONLY (in mirror, not live) | 0 |
| LIVE-ONLY (on VM, not in mirror) | 121 |

## MIRRORED dirs — full detail (11)

All 11 mirrored skill dirs are **MIRRORED-IDENTICAL**: every SKILL.md, and every reference/template file the mirror also carries, matches VM byte-for-byte by md5 (16 files checked total across the 11 dirs — see repo file list below). 0 MIRRORED-DIVERGED.

| skill dir | md5 match | VM mtime (UTC) | VM size | repo last commit (UTC) | repo→VM commit lag |
|---|---|---|---|---|---|
| `creative/mobile-club-okr-report` | MATCH | 2026-08-14T11:27:49+0000 | 5966 | 2026-08-14T11:28:49+00:00 | 0:01:00 |
| `delegate-to-cooper` | MATCH | 2026-08-08T12:11:15+0000 | 34794 | 2026-08-08T12:12:04+00:00 | 0:00:49 |
| `email/gmail-triage` | MATCH | 2026-08-12T15:55:56+0000 | 20538 | 2026-08-12T15:58:04+00:00 | 0:02:08 |
| `email/himalaya` | MATCH | 2026-08-11T18:19:56+0000 | 7632 | 2026-08-11T18:21:52+00:00 | 0:01:56 |
| `helptech-duty-intake` | MATCH | 2026-08-11T15:28:29+0000 | 10475 | 2026-08-11T15:28:48+00:00 | 0:00:19 |
| `hermes-operations/hermes-orchestration` | MATCH | 2026-08-08T10:52:08+0000 | 9753 | 2026-08-08T10:52:40+00:00 | 0:00:32 |
| `orchestration/daily-work-brief` | MATCH | 2026-08-13T13:16:09+0000 | 18571 | 2026-08-13T13:18:07+00:00 | 0:01:58 |
| `orchestration/engagement-checker` | MATCH | 2026-08-13T14:55:48+0000 | 81873 | 2026-08-13T14:56:32+00:00 | 0:00:44 |
| `orchestration/linear-ticketing` | MATCH | 2026-08-13T15:18:19+0000 | 27619 | 2026-08-13T15:20:19+00:00 | 0:02:00 |
| `orchestration/secure-delta-collectors` | MATCH | 2026-08-11T12:35:50+0000 | 11296 | 2026-08-11T16:38:32+00:00 | 4:02:42 |
| `orchestration/slack-channel-context` | MATCH | 2026-08-13T12:36:06+0000 | 11086 | 2026-08-13T12:36:43+00:00 | 0:00:37 |

Note on `.bak-*` litter: `delegate-to-cooper`, `email/gmail-triage`, `orchestration/{daily-work-brief,engagement-checker,linear-ticketing,slack-channel-context}` each carry VM-only `SKILL.md.bak-*` files (Tars' own pre-edit backups per the hard-rule `.bak` copy-before-edit workflow — confirmed by name, e.g. `SKILL.md.bak-tars-selfedit`). These are edit-history litter, not un-mirrored skill content — expected, not a mirror gap.

## LIVE-ONLY — by mtime cluster (provenance)

- **Cluster A — `2026-08-07T16:17:59+0000`, n=68**: exact-second match across dozens of unrelated dirs → single bulk copy/install event. Shipped with Hermes, not agent-authored.
- **Cluster B — `2026-08-08T12:00:37+0000`, n=35**: same signature, a second bulk install/upgrade batch (matches the vanilla-skills reset referenced in memory `hermes-upgrade-reapply-local-patches.md`). Not agent-authored.
- **Cluster C — `2026-06-20T18:44:59` .. `18:46:50`, n=6** (ponytail family): tight 2-minute window, predates Tars' 2026-08-07 go-live entirely → pre-Tars install, not agent-authored.

Clusters A+B+C account for 109 of 121 live-only dirs. The remaining 12 have a lone, non-clustered mtime each — listed individually below, oldest first.

| skill dir | VM mtime (UTC) | size | non-SKILL.md files | pre/post Tars go-live (2026-08-07) | provenance read |
|---|---|---|---|---|---|
| `creative/meme-generator` | 2026-07-31T16:31:01+0000 | 7837 | 0 | pre | pre-Tars — not a candidate |
| `email/rich-email-composition` | 2026-08-02T15:24:18+0000 | 6476 | 0 | pre | pre-Tars — not a candidate |
| `i-have-adhd` | 2026-08-07T16:32:30+0000 | 6848 | 0 | post | same-day as Cluster A (15 min after) — ambiguous, likely still install-time |
| `last30days` | 2026-08-08T12:05:11+0000 | 222241 | 1 | post | 5 min after Cluster B — ambiguous, likely large-skill install straggler |
| `hermes-operations/skill-library-auditing` | 2026-08-08T12:47:48+0000 | 4232 | 0 | post | 47 min after Cluster B, hermes-ops category — MEDIUM confidence suspect |
| `email/durable-email-automation` | 2026-08-10T09:51:44+0000 | 6748 | 0 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |
| `research/upstream-change-research` | 2026-08-10T12:37:22+0000 | 5286 | 0 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |
| `communication/stakeholder-communication` | 2026-08-11T16:26:39+0000 | 5029 | 0 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |
| `productivity/google-workspace` | 2026-08-11T19:46:43+0000 | 12883 | 1 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |
| `software-development/local-ipc-probes` | 2026-08-14T10:53:23+0000 | 6713 | 0 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |
| `hermes-operations/hermes-skill-portability` | 2026-08-14T11:27:00+0000 | 5440 | 0 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |
| `hermes-operations/hermes-gateway-operations` | 2026-08-14T11:32:55+0000 | 9648 | 0 | post | isolated, well outside both bulk windows — STRONG suspect (skill_manage self-edit, never landed by PR) |

## Suspected Tars-authored-but-never-landed skills

Per SOUL rule 2, every self-edit via `skill_manage` must land here by self-merged PR in the same turn. These live-only dirs have a lone mtime, postdate Tars' 2026-08-07 go-live, and sit well outside both bulk-install windows — i.e. no bulk-install explanation fits, and no repo commit exists.

**Strong confidence (7):**

- `email/durable-email-automation` — mtime 2026-08-10T09:51:44+0000, 6748 bytes
- `research/upstream-change-research` — mtime 2026-08-10T12:37:22+0000, 5286 bytes
- `communication/stakeholder-communication` — mtime 2026-08-11T16:26:39+0000, 5029 bytes
- `productivity/google-workspace` — mtime 2026-08-11T19:46:43+0000, 12883 bytes
- `software-development/local-ipc-probes` — mtime 2026-08-14T10:53:23+0000, 6713 bytes
- `hermes-operations/hermes-skill-portability` — mtime 2026-08-14T11:27:00+0000, 5440 bytes
- `hermes-operations/hermes-gateway-operations` — mtime 2026-08-14T11:32:55+0000, 9648 bytes

**Medium confidence (1, close to a bulk window so not ruled out as install-time):**

- `hermes-operations/skill-library-auditing` — mtime 2026-08-08T12:47:48+0000, 4232 bytes

Note: `productivity/google-workspace` (mtime 2026-08-11T19:46:43Z, live-only) is already tracked separately in memory `hermes-upgrade-reapply-local-patches.md` as a known local customization (venv-python) that must survive each `/upgrade` — its drift has a different, already-documented cause, distinguish it from the skill_manage-authored suspects above.

## REPO-ONLY

None. Every skill dir in the repo mirror exists live on the VM with an identical SKILL.md.

## Cluster A / B directory listing (for completeness)

<details><summary>Cluster A — 68 dirs, all mtime 2026-08-07T16:17:59+0000</summary>

`apple/apple-notes`, `apple/apple-reminders`, `apple/findmy`, `apple/imessage`, `autonomous-ai-agents/claude-code`, `autonomous-ai-agents/codex`, `autonomous-ai-agents/computer-use`, `autonomous-ai-agents/hermes-agent`, `autonomous-ai-agents/opencode`, `creative/architecture-diagram`, `creative/ascii-art`, `creative/ascii-video`, `creative/baoyu-infographic`, `creative/claude-design`, `creative/comfyui`, `creative/design-md`, `creative/excalidraw`, `creative/humanizer`, `creative/manim-video`, `creative/p5js`, `creative/popular-web-designs`, `creative/pretext`, `creative/sketch`, `creative/songwriting-and-ai-music`, `creative/touchdesigner-mcp`, `github/codebase-inspection`, `github/github-auth`, `github/github-code-review`, `github/github-issues`, `github/github-pr-workflow`, `github/github-repo-management`, `media/gif-search`, `media/songsee`, `media/youtube-content`, `mlops/evaluation/evaluating-llms-harness`, `mlops/evaluation/weights-and-biases`, `mlops/huggingface-hub`, `mlops/inference/llama-cpp`, `mlops/inference/serving-llms-vllm`, `note-taking/obsidian`, `productivity/airtable`, `productivity/docx`, `productivity/maps`, `productivity/nano-pdf`, `productivity/notion`, `productivity/ocr-and-documents`, `productivity/pdf`, `productivity/powerpoint`, `productivity/teams-meeting-pipeline`, `productivity/xlsx`, `research/arxiv`, `research/blogwatcher`, `research/grounded-citations`, `research/llm-wiki`, `research/research-paper-writing`, `smart-home/openhue`, `social-media/xurl`, `software-development/dogfood`, `software-development/hermes-agent-skill-authoring`, `software-development/inspecting-hermes-desktop-dom`, `software-development/node-inspect-debugger`, `software-development/plan`, `software-development/python-debugpy`, `software-development/requesting-code-review`, `software-development/simplify-code`, `software-development/spike`, `software-development/systematic-debugging`, `software-development/test-driven-development`

</details>

<details><summary>Cluster B — 35 dirs, all mtime 2026-08-08T12:00:37+0000</summary>

`ask-matt`, `claude-handoff`, `code-review`, `codebase-design`, `diagnosing-bugs`, `domain-modeling`, `git-guardrails-claude-code`, `grill-me`, `grill-with-docs`, `grilling`, `handoff`, `implement`, `improve-codebase-architecture`, `loop-me`, `mattpocock-research`, `migrate-to-shoehorn`, `prototype`, `resolving-merge-conflicts`, `scaffold-exercises`, `setup-matt-pocock-skills`, `setup-pre-commit`, `setup-ts-deep-modules`, `tdd`, `teach`, `to-questionnaire`, `to-spec`, `to-tickets`, `triage`, `wait-what`, `wayfinder`, `wizard`, `writing-beats`, `writing-for-agents`, `writing-fragments`, `writing-shape`

</details>

<details><summary>Cluster C (ponytail family) — 6 dirs, 2026-06-20T18:44:59..18:46:50, pre-Tars</summary>

`software-development/ponytail`, `software-development/ponytail-audit`, `software-development/ponytail-debt`, `software-development/ponytail-gain`, `software-development/ponytail-help`, `software-development/ponytail-review`

</details>

## Recommendation

One-way VM→repo sync for the 11 already-mirrored dirs needs nothing (all MIRRORED-IDENTICAL, 0 diverged, so no which-side-moved investigation required there); land the 7 strong-confidence live-only suspects (`email/durable-email-automation`, `research/upstream-change-research`, `communication/stakeholder-communication`, `productivity/google-workspace`, `software-development/local-ipc-probes`, `hermes-operations/hermes-skill-portability`, `hermes-operations/hermes-gateway-operations`) into `skills/` by PR next — they are Tars' own un-landed skill_manage edits per SOUL rule 2, not shipped content, so pulling VM→repo is the only direction that makes sense (no repo-side edits exist to conflict with); treat the 2 same-day-of-a-bulk-window entries (`i-have-adhd`, `last30days`) and the 1 medium-confidence entry (`hermes-operations/skill-library-auditing`) as needs-owner-input before syncing, since a bulk-install origin can't be ruled out from mtime alone (note `productivity/google-workspace` is in the strong-7 by mtime evidence but has a documented alternate cause — a manually-applied venv-python patch tracked in memory, not necessarily a `skill_manage` edit — so confirm provenance before PR-ing it, not before mirroring it); leave the 68+35+6 clustered dirs out of this sync entirely.

