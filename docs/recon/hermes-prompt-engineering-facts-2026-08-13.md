# Hermes-agent prompt-engineering facts (v0.20.0, pinned checkout on Tars VM)

Sources: `~/.hermes/hermes-agent` on 192.168.0.9, git remote
`https://github.com/NousResearch/hermes-agent.git` (`origin`), HEAD
`4d10183d6160411b52a673244f1fd22ee9fbc07c` (2026-08-10), `pyproject.toml` version
`0.20.0` — matches Tars. Upstream docs fetched live at
https://hermes-agent.nousresearch.com/docs/... (same version family) but the
**authoritative quotes below are from the local checked-out copy** of
`website/docs/guides/use-soul-with-hermes.md` and
`website/docs/developer-guide/prompt-assembly.md`, which is pinned to Tars'
exact commit — more trustworthy than the live site for version-exactness.

---

## 1. HARD constraints (format / size / parse)

### 1.1 Truncation — corrects the ~20K assumption

`agent/prompt_builder.py`:
```python
CONTEXT_FILE_MAX_CHARS = 20_000          # floor
CONTEXT_TRUNCATE_HEAD_RATIO = 0.7
CONTEXT_TRUNCATE_TAIL_RATIO = 0.2
_CONTEXT_FILE_CHARS_PER_TOKEN = 4
_CONTEXT_FILE_WINDOW_FRACTION = 0.06
_CONTEXT_FILE_DYNAMIC_CEILING = 500_000
```
`_dynamic_context_file_max_chars(context_length)` (line 1297): when no explicit
`context_file_max_chars` is set in `config.yaml` (Tars has none — `grep
context_file_max_chars ~/.hermes/config.yaml` empty), the cap is
`max(20_000, min(context_length * 4 * 0.06, 500_000))`.

Tars' model cache, `~/.hermes/context_length_cache.yaml`:
```yaml
context_lengths:
  gpt-5.6-sol@https://chatgpt.com/backend-api/codex: 272000
```
→ **live cap ≈ 272000 × 4 × 0.06 = 65,280 chars**, not 20,000. The 20K figure
is only the floor for small-context models / when the model's window is
unknown. SOUL v2 (19,234 bytes) is under *both* the conservative 20K floor
and the real ~65K dynamic cap by a wide margin — headroom is larger than the
task brief assumed, but see §3 on why that headroom isn't free to spend.

**Truncation mechanism when it does fire** (`_truncate_content`, line 1964):
**head+tail, not a hard cutoff** — keeps `head_chars = 0.7 × cap` from the
start and `tail_chars = 0.2 × cap` from the end, drops the middle, and splices
in a marker: `[...truncated SOUL.md: kept {head}+{tail} of {total} chars. The
middle is omitted — read_file: {path}]`. A truncation warning is also
surfaced to the chat status channel (`drain_truncation_warnings()` in
`agent/system_prompt.py`). No truncation warnings appear in
`~/.hermes/logs/*.log` for SOUL.md — confirms it isn't firing today.

Same cap/mechanism applies to `AGENTS.md`/`.hermes.md`/`CLAUDE.md`/
`.cursorrules` context files, each independently — SOUL.md is not competing
with those for one shared budget, each file gets its own 65K-ish cap.

### 1.2 Structural parsing — SOUL.md is opaque text, only security-scanned

`load_soul_md()` (prompt_builder.py:2004): read → `.strip()` →
`_scan_context_content()` (prompt-injection/promptware/role-play-hijack scan,
"context" scope — NOT the stricter "strict" scope that flags SSH-backdoor/
persistence/exfil patterns, so SOUL v2's own `ssh cooper "..."` runbooks don't
risk self-block) → `_truncate_content()` → returned as one opaque string.
**Hermes parses nothing structurally out of SOUL.md** — no front-matter, no
section-name extraction, no directive syntax. Headings (`# Identity`, etc.)
are a human/model-readability convention only, not machine-read.

### 1.3 Per-session vs per-turn loading

`agent/system_prompt.py` docstring + `build_system_prompt()`: **the whole
system prompt (including SOUL.md) is built ONCE per session and cached** on
`agent._cached_system_prompt`; it is only rebuilt after a **context
compression event** (batch compaction at `compression.threshold` — Tars:
`0.5`, or micro-compaction if `micro_compact: true` — Tars: `false`, i.e.
Tars only rebuilds on the batch-compaction path) or on session
restore/failover (`reconstruct_static_prefix`). It is **not** reloaded fresh
on every message, despite the legacy SOUL.md scaffold comment ("This file is
loaded fresh each message — no restart needed") that shipped in old
installers — that comment is gone from `DEFAULT_SOUL_MD` and from SOUL v2,
and SOUL v2's own "Standing corrections" section states the correct, narrower
behavior itself ("An edit here reaches new sessions only... in the thread
where the correction was given I keep applying it from the transcript... a
fresh thread or after a session reset") — this matches source exactly.
`session-lifecycle.md`: a session is a `chat_id`/`thread_id`-scoped
conversation lane that resets on idle/daily policy or `/new`/`/reset`; a
Slack thread session persists (SOUL v2 rule: "A Slack thread is one session
that never expires") until compaction/reset triggers a rebuild.

### 1.4 Where SOUL.md sits in assembly order

`build_system_prompt_parts()` (`agent/system_prompt.py`), three tiers joined
`stable → context → volatile`:

**stable** (cache-critical prefix, in order):
1. `SOUL.md` content (or `DEFAULT_AGENT_IDENTITY` fallback) — **slot #1,
   always first**
2. `HERMES_AGENT_HELP_GUIDANCE` (pointer to hermes-agent docs/skill)
3. `TASK_COMPLETION_GUIDANCE` (anti-fabrication/finish-the-job — gated on
   `agent.task_completion_guidance`, default true, only if tools present)
4. `PARALLEL_TOOL_CALL_GUIDANCE` (batch independent tool calls)
5. tool-conditional guidance block: `MEMORY_GUIDANCE` +
   `SESSION_SEARCH_GUIDANCE` + `SKILLS_GUIDANCE` + `KANBAN_GUIDANCE` (each
   only if the matching tool is loaded)
6. `STEER_CHANNEL_NOTE` (mid-turn steering marker semantics)
7. computer-use guidance (if tool present)
8. Nous subscription block
9. `TOOL_USE_ENFORCEMENT_GUIDANCE` + model-family block — for Tars
   (`gpt-5.6-sol` via `openai-codex` → matches `"codex"`/`"gpt"` substrings)
   this injects **`OPENAI_MODEL_EXECUTION_GUIDANCE`** (see §3, full text
   quoted there)
10. environment hints, coding posture, environment probe line, active-profile
    hint, platform hint

**context**: coding workspace snapshot → caller `system_message` → discovered
project context files (`.hermes.md`/`AGENTS.md`/`CLAUDE.md`/`.cursorrules`
under `TERMINAL_CWD`, first-match-wins priority, SOUL.md excluded here via
`skip_soul=True` once already loaded in stable).

**volatile** (rendered last on purpose, most likely to change turn-to-turn):
skills index → memory snapshot (`MEMORY.md`) → `USER.md` profile → external
memory-provider block → timestamp/session/model/provider line.

Caching consequence: because SOUL.md is the very first byte of the very
first tier, **any edit to SOUL.md — including an append at its own tail —
invalidates the prefix cache for the entire rest of the system prompt** on
the next rebuild (all later stable blocks + context + volatile). This is a
real cost of SOUL v2's growing, append-only "Standing corrections" section
sitting inside the one tier upstream designed to be the most stable.

---

## 2. Conventions — what upstream prescribes for a good SOUL.md

Verbatim from `website/docs/guides/use-soul-with-hermes.md` (local checkout,
pinned to Tars' v0.20.0 commit):

> **What SOUL.md is for:** tone, personality, communication style, how
> direct or warm Hermes should be, what Hermes should avoid stylistically,
> how Hermes should relate to uncertainty, disagreement, and ambiguity.
>
> **What SOUL.md is not for:** repo-specific coding conventions, file paths,
> commands, service ports, architecture notes, project workflow
> instructions. **Those belong in AGENTS.md.**
>
> A good rule: if it should apply everywhere, put it in SOUL.md; if it only
> belongs to one project, put it in AGENTS.md.
>
> A strong SOUL.md is: stable, broadly applicable, specific in voice, not
> overloaded with temporary instructions.
> A weak SOUL.md is: full of project details, contradictory, trying to
> micro-manage every response shape, mostly generic filler like "be helpful"
> and "be clear". **Hermes already tries to be helpful and clear — SOUL.md
> should add real personality and style, not restate obvious defaults.**
>
> Suggested structure: `# Identity` / `# Style` / `# Avoid` / `# Defaults`.
>
> Troubleshooting — "My SOUL.md became too project-specific": Move project
> instructions into AGENTS.md and keep SOUL.md focused on identity and
> style.

`SOUL.md` vs `/personality`: SOUL.md = durable baseline voice; `/personality
<name>` = temporary session overlay (config `agent.personalities`, Tars has
the stock set — helpful/concise/technical/creative/teacher/kawaii/catgirl/
pirate/shakespeare/surfer — unused). Not a substitute for SOUL content.

---

## 3. Interactions — what SOUL.md should NOT duplicate (Hermes injects it anyway)

Full text of the guidance blocks Hermes auto-injects around SOUL.md,
regardless of what SOUL.md says (`agent/prompt_builder.py`):

- **`TASK_COMPLETION_GUIDANCE`** (universal, injected whenever `tools` are
  loaded — always true for Tars): "the deliverable is a working artifact
  backed by real tool output — not a description of one... NEVER substitute
  plausible-looking fabricated output... Reporting a blocker honestly is
  always better than inventing a result."
- **`OPENAI_MODEL_EXECUTION_GUIDANCE`** (injected because Tars' model string
  matches `"codex"`/`"gpt"`) — six tagged blocks, most relevant two quoted in
  full:
  - `<verification>`: "Before finalizing your response: Correctness — does
    the output satisfy every stated requirement? Grounding — are factual
    claims backed by tool outputs or provided context? ... Safety — if the
    next step has side effects (file writes, commands, API calls), confirm
    scope before executing."
  - `<missing_context>`: "If required context is missing, do NOT guess or
    hallucinate an answer. Use the appropriate lookup tool... Ask a
    clarifying question only when the information cannot be retrieved by
    tools. If you must proceed with incomplete information, label
    assumptions explicitly."
- **`MEMORY_GUIDANCE`** (injected whenever `memory` tool loaded — true for
  Tars): full rules on what to save, declarative-not-imperative phrasing,
  "if a fact will be stale in 7 days, it does not belong in memory."
- **`SKILLS_GUIDANCE`** (injected whenever skill tools loaded — true for
  Tars): "save the approach as a skill with skill_manage... patch it
  immediately with skill_manage(action='patch') — don't wait to be asked."
- **`SESSION_SEARCH_GUIDANCE`**, **`PARALLEL_TOOL_CALL_GUIDANCE`**,
  **`TOOL_USE_ENFORCEMENT_GUIDANCE`**, **`STEER_CHANNEL_NOTE`**,
  **`HERMES_AGENT_HELP_GUIDANCE`** — all likewise auto-present, all
  independent of SOUL.md content.

**These are config/tool-gated, not SOUL-gated** — nothing in SOUL v2
disables or weakens them; they fire purely off `agent.valid_tool_names` /
`config.yaml` flags. Confirmed live for Tars: `task_completion_guidance:
true`, `parallel_tool_call_guidance: true`, `tool_use_enforcement: auto`
(`hermes config get agent`).

---

## 4. SOUL v2 vs these sources — top violations found

1. **Genre mismatch (upstream's #1 documented mistake, at the extreme end).**
   SOUL v2 (317 lines / 19,234 bytes) is ~90% hard rules, an escalation/
   delegation protocol, Slack channel IDs, machine IPs, and **two full
   verbatim bash runbooks** (the skill-mirror and SOUL-mirror PR flows,
   ≈90 lines / ~5.5KB combined) — exactly the "repo-specific... commands...
   project workflow instructions" category upstream says belongs in
   `AGENTS.md`, not `SOUL.md`, taken further than the guide's own worst-case
   example ("full of project details"). This is a deliberate, load-bearing
   choice (SOUL v2 IS Tars' governance doc, and `skill_manage` structurally
   *cannot* write SOUL.md — see repo's `CLAUDE.md` §"Tars edits its own
   skills" — so there's no other injectable-and-editable-by-Tars surface for
   the self-merge recipe), but it's worth naming plainly: this is not what
   upstream calls a SOUL.md, it's an AGENTS.md-shaped document occupying
   SOUL.md's slot, and it pays the "always in the cache-critical prefix,
   busts the whole downstream prompt on every edit" cost that a real
   AGENTS.md-tier file (loaded per-cwd, not globally) wouldn't.
2. **Duplicates Hermes' own auto-injected anti-fabrication/verify-before-done
   guidance.** SOUL v2's "Done means verified" bullet ("What I have not
   verified I report as 'not verified', never as done... Facts follow the
   same law: I state only what I have read") restates, in different words,
   content Hermes already guarantees structurally and unconditionally for
   Tars via `TASK_COMPLETION_GUIDANCE` + `OPENAI_MODEL_EXECUTION_GUIDANCE`'s
   `<verification>`/`<missing_context>` blocks (§3) — the same "no
   fabrication, verify before claiming done, label assumptions" principle
   appears up to 3 times in one prompt. Bytes spent re-litigating a
   structural guarantee, contra "SOUL.md should add real personality and
   style, not restate obvious defaults."
3. **Growth mechanism sits in the most cache-sensitive slot.** The
   "Standing corrections" section is explicitly designed to grow by
   append over time and is the very last content of SOUL.md — but SOUL.md
   is also the very *first* item in the stable tier (§1.4), so every append
   invalidates the prefix cache for literally everything downstream (rest
   of stable + all of context + all of volatile) on the next rebuild. Not a
   correctness bug, but a real, uncosted tradeoff of using SOUL.md instead
   of, say, a skill or a `MEMORY.md`-style file for a log meant to grow
   indefinitely.

(Positive finding, not a violation: SOUL v2's Phase-2 section correctly
defers to `~/.hermes/memories/USER.md` rather than duplicating it —
"Not yet written; do not invent content for it... live mechanism:
`~/.hermes/memories/USER.md`" — this is exactly right per §3.)

---

## 5. THE HYPOTHESIS — verbatim shipped defaults + self-improvement mechanism audit

### 5.1 Shipped default SOUL.md, verbatim (`hermes_cli/default_soul.py`)

```
You are Hermes Agent, an intelligent AI assistant created by Nous Research.
You are helpful, knowledgeable, and direct. You assist users with a wide
range of tasks including answering questions, writing and editing code,
analyzing information, creative work, and executing actions via your tools.
You communicate clearly, admit uncertainty when appropriate, and prioritize
being genuinely useful over being verbose unless otherwise directed below.
Be targeted and efficient in your exploration and investigations.
```
(6 sentences, ≈600 bytes.) Seeded by `_ensure_default_soul_md()` in
`hermes_cli/config.py:833` on first run, and used to auto-upgrade legacy
comment-only scaffolds (`is_legacy_template_soul()`) — never touches a
SOUL.md a user actually wrote content into.

**Legacy pre-`DEFAULT_SOUL_MD` scaffold** (what `install.sh`/`install.ps1`/
`docker/SOUL.md` used to seed, now auto-upgraded away, quoted for context —
this is the "loaded fresh each message" text referenced in the task):
```
# Hermes Agent Persona

<!--
This file defines the agent's personality and tone.
The agent will embody whatever you write here.
Edit this to customize how Hermes communicates with you.

Examples:
  - "You are a warm, playful assistant who uses kaomoji occasionally."
  - "You are a concise technical expert. No fluff, just facts."
  - "You speak like a friendly coworker who happens to know everything."

This file is loaded fresh each message -- no restart needed.
Delete the contents (or this file) to use the default personality.
-->
```

### 5.2 No default AGENTS.md / IDENTITY.md / USER.md / MEMORY.md templates exist

Grepped `hermes_bootstrap.py`, `hermes_cli/config.py`, `hermes_cli/
config_defaults.py`, `agent/*.py` for any seeded template for these
filenames — **none found**. `SOUL.md` is the *only* persona/identity file
hermes-agent auto-seeds. `MEMORY.md`/`USER.md` are not templates at all —
they're runtime-populated artifacts the memory subsystem writes to as it
learns (confirmed: Tars' `~/.hermes/memories/MEMORY.md` and `USER.md` are
plain fact lists with no boilerplate header, written by the `memory` tool /
Hindsight plugin, not by a bootstrap step). `~/.hermes/AGENTS.md` (77,543
bytes) that exists in `~/.hermes/hermes-agent/` is the **hermes-agent
project's own contributor guide** — irrelevant to a running profile's
persona surface, not a per-profile template.

### 5.3 What actually makes a stock Hermes agent "get better with time"

`README.md` (local, matches origin): tagline **"The self-improving AI agent
built by Nous Research"**, feature table row: *"A closed learning loop —
Agent-curated memory with periodic nudges. Autonomous skill creation..."*
This is realized entirely as **runtime code + config, not SOUL.md prose**:

| Mechanism | Source | Tars config | Live status |
|---|---|---|---|
| Agent-curated memory (`MEMORY.md`) | `agent/turn_context.py`, `MEMORY_GUIDANCE` | `memory.memory_enabled: true`, `memory.nudge_interval: 10` | **ACTIVE** — `~/.hermes/memories/MEMORY.md` has 3 recent, well-formed declarative facts (last touched 2026-08-12) |
| User profile (`USER.md`) | same memory subsystem | `memory.user_profile_enabled: true`, `memory.user_char_limit: 1375` | **ACTIVE** — `USER.md` has 18 lines, last touched 2026-08-13 (today), style matches `MEMORY_GUIDANCE`'s declarative-fact convention exactly |
| Periodic memory nudge (every N turns, reminds agent to save) | `agent/turn_context.py:556-603` | `nudge_interval: 10` | **ACTIVE** |
| Autonomous skill creation/patching | `SKILLS_GUIDANCE`, `agent/turn_finalizer.py:708`, `skill_manage` tool | `skills.creation_nudge_interval: 15` | **ACTIVE** — Tars self-authors/patches its own `SKILL.md` files per this repo's `CLAUDE.md` ("did so 7 times during the first live Orca run"); `~/.hermes/skills/` has many `SKILL.md` files with mtimes newer than the hermes-agent README, consistent with live edits |
| Cross-session recall (`session_search`) | `SESSION_SEARCH_GUIDANCE` | tool present in Tars' toolset (referenced explicitly in SOUL v2 rule 10) | **ACTIVE** |
| Batch/micro-compaction (keeps long sessions usable without losing user intent) | `docs/micro-compaction.md` | `compression.enabled: true`, `threshold: 0.5`, `micro_compact: false` | **ACTIVE** (batch only; micro off — default) |
| Trajectory compression / batch training-data generation ("Research-ready") | `trajectory_compressor.py`, `batch_runner.py` | n/a — not a personal-assistant feature | **N/A for Tars** — unrelated to "getting better," it's Nous' model-training pipeline |

**Conclusion on the hypothesis:** the default `SOUL.md` contributes
approximately **nothing** to "getting better with time" — it's 6 generic
sentences with zero mention of memory, skills, learning, or self-editing.
Every mechanism the README credits for the "self-improving" claim
(memory-with-nudges, autonomous skill authoring, session recall) is
**structural**: gated by `config.yaml` flags and tool availability, injected
into the prompt automatically by `agent/prompt_builder.py` regardless of
SOUL.md's content (§3), and **none of it was touched, disabled, or
displaced** by replacing the default SOUL.md with Tars' custom one — the
config keys and injected guidance blocks are entirely independent of SOUL
content. All six checked mechanisms are live and populated for Tars today.
**The hypothesis as stated ("our SOUL.md lowered the smartness/capabilities")
is not supported by the mechanism-level evidence**: nothing SOUL v2 does
turns off or weakens the self-improvement loop. The more defensible version
of the concern is the one in §4 — SOUL v2 spends its highest-priority,
always-loaded prompt real estate on governance/procedure that upstream's own
docs say degrades what SOUL.md is *for* (voice/identity), and duplicates
~150-200 words of structurally-guaranteed anti-fabrication guidance — a
prompt-quality cost, not a disabled-capability cost.
