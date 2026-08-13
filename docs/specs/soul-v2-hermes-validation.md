# SOUL v2 — Hermes best-practice validation & default-SOUL hypothesis (2026-08-13)

Follow-up to `soul-v2-rationale.md` / `soul-v2-changemap.md`. Evidence:
`docs/recon/hermes-prompt-engineering-facts-2026-08-13.md` (source-cited, pinned
to the VM's hermes-agent commit `4d10183`, v0.20.0).

## Gaetan's hypothesis

> "Hermes agent is famously known for getting better with time. Maybe the
> default SOUL.md is a big part of that and us defining our own SOUL.md lowered
> the smartness and capabilities of Hermes/Tars too much."

## Verdict: REFUTED at the mechanism level

- The shipped default SOUL.md (`hermes_cli/default_soul.py`) is 6 generic
  sentences, ~600 bytes. It contains nothing about memory, skills, learning or
  self-editing — nothing capability-bearing to displace.
- Every mechanism upstream credits for "self-improving" is structural — gated
  by `config.yaml` flags and tool availability, injected by
  `agent/prompt_builder.py` regardless of SOUL.md content: agent-curated
  MEMORY.md + nudges, USER.md profiling, autonomous skill creation/patching,
  cross-session recall, compaction.
- All six mechanisms verified ACTIVE and populated on Tars (2026-08-13):
  MEMORY.md last touched 08-12, USER.md touched 08-13, live skill self-edits,
  `memory_enabled/user_profile_enabled: true`, nudge intervals set.
- Conclusion: replacing the default SOUL.md disabled nothing. No restored-text
  delta is warranted; SOUL v2 stands as applied (commit `c785b13`).

## Residual (true but accepted, litigated in the v2 review round)

1. Genre mismatch vs upstream ("commands/workflow → AGENTS.md, SOUL = voice"):
   deliberate — SOUL.md is the ONLY always-injected surface Tars can edit
   itself (`skill_manage` cannot write it; per-profile AGENTS.md is not a
   Hermes surface — discovery is TERMINAL_CWD-scoped, unreliable for a
   Slack-first agent). The governance content has nowhere better to live.
2. ~150–200 words overlap with auto-injected TASK_COMPLETION /
   OPENAI_MODEL_EXECUTION guidance: kept — persona-voiced, incident-derived,
   and insurance if `task_completion_guidance` is ever flipped off; ~300 bytes
   of a measured 65,280-char budget.
3. Standing-corrections appends bust the downstream prompt cache (SOUL is
   byte-0 of the stable tier): accepted at current append frequency.

Size guidance corrected in `soul-v2-changemap.md` §Size (measured 65,280-char
live cap; middle-drop truncation spares the last-position corrections section).
