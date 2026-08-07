# Probe — is the skill map load-bearing, or is the pointer enough? (P6/kb)

Run 2026-08-07 23:32–23:34 UTC (VM clock) from cooper, read-only on the VM.
Two `~/.local/bin/hermes chat -Q -q` oneshots, nothing written under
`~/.hermes/`, no `git` write of any kind on `~/dev/mc-metarepo`.

## Why this trial exists

`kb-retrieval-trials.md` ranks mechanism **(c) — a `SKILL.md` map of the repo
layout → one targeted `read_file`** as the winner, 5/5 (2 fired live, 3
reasoned). An adversarial read caught that every (c) run bundles **two**
changes at once:

1. **the pointer fix** — Tars' only knowledge-base pointer is
   `~/dev/gaetan-metarepo` (`~/.hermes/SOUL.md`), a path that does not exist on
   the VM; baseline Q4 produced two consecutive
   `search_files … {"error": "Path not found: /home/gaetan/dev/gaetan-metarepo"}`.
   The map prompt named `~/dev/mc-metarepo`, so (c) silently fixed this.
2. **the map itself** — the layout table + the 40-slug list + the routing rules.

So every (c) win is confounded, and the laziest candidate of all —
**Tars simply knows the right path, no map installed** — was never measured.
If the pointer alone answers, the map is not needed and P6 should propose
almost nothing.

## Method + the proxy's limitation

SOUL.md may not be edited (a peer is rewriting it right now), so the pointer is
delivered as a **sentence appended to the question**:

```
~/.local/bin/hermes chat -Q -q '<the original question> The knowledge base is a
git clone at ~/dev/mc-metarepo on this machine.'
```

Call this **(c-minus)**: pointer, no map, no slug list, no routing rules, no
"never read whole" rules, no "the submodules are empty" rule.

Re-fired **Q1** and **Q3** — the two questions the baseline got wrong
(Q1: 422 s, delegated to an Orca worker, missed the trap · Q3: 22 s, zero
retrieval, advice the note explicitly forbids).

**State of the VM at run time, verified read-only** — the confound is genuinely
still absent from the agent's own context:

```
$ grep -in "metarepo" ~/.hermes/SOUL.md
81:source of truth is `~/dev/gaetan-metarepo` on cooper; the live mechanism on my
$ grep -ic "mc-metarepo" ~/.hermes/SOUL.md
0
$ ls ~/.hermes/skills/
apple  autonomous-ai-agents  creative  delegate-to-cooper  email  github
hermes-lcm  hermes-operations  i-have-adhd  media  mlops  note-taking
productivity  research  smart-home  social-media  software-development
```

SOUL.md still names the dead path, still never names `mc-metarepo`, and no KB
skill is installed. The path in the prompt is the agent's only pointer.

**Limitations, stated plainly — do not overstate this result:**

- **A path in the user's question is not a path in SOUL.md.** Different
  position in the prompt (end of the user turn vs. system persona), different
  salience, and above all: SOUL is present on *every* turn while a question
  carrying the path is present only when Gaetan happens to type it. (c-minus)
  measures *"does knowing the path suffice?"*, not *"will a SOUL.md line
  deliver that knowledge?"*. The second question stays untested.
- **n=1 per cell**, a nondeterministic model, and no repeats.
- **SOUL.md changed between the baseline and this run** — mtime
  `2026-08-07 23:10:42`, i.e. 22 min before these turns, by the peer session.
  Baseline (22:50) and (c-minus) (23:32) did not see byte-identical personas.
  The dead KB pointer survived the rewrite, but other text moved (the metarepo
  line is at `:81` now, it was `:54` during the baseline).
- Tool-output bytes below are the `chars` figures Hermes logs per tool result;
  the failed `kanban_show` (a per-turn artifact of every CLI oneshot on this
  VM, ~60 chars of error) is counted as a call but not as bytes.

## Q1 — "In api-v2, how do I tell a Next Mobiles order apart from a Mobile.Club order?"

Ground truth (`knowledge/nm-vs-mc-order-source.md`, 767 B): a row in
`next_mobile_order_mappings` (`order_id`) ⇒ Next Mobiles, none ⇒ Mobile.Club,
as in `IsNextMobilesOrderService` — **and** the trap: do **not** use
`orders.business_entity_id`, `CreateNmOrderService` never sets it.

Command:

```
$ ~/.local/bin/hermes chat -Q -q "In api-v2, how do I tell a Next Mobiles order apart from a Mobile.Club order? The knowledge base is a git clone at ~/dev/mc-metarepo on this machine."
```

Verbatim output:

```
session_id: 20260807_233240_5a3191
Check `next_mobile_order_mappings` by `order_id`:

- Mapping exists → Next Mobiles order
- No mapping → Mobile.Club order

This is the logic used by `IsNextMobilesOrderService`.

Do not use `orders.business_entity_id`: `CreateNmOrderService` leaves it at the default Mobile.Club value, so it is unreliable.

Source: `~/dev/mc-metarepo/knowledge/nm-vs-mc-order-source.md`
=== WALLCLOCK: 23s ===
```

Tool trace (`~/.hermes/logs/agent.log`, verbatim):

```
2026-08-07 23:32:46,850 WARNING [20260807_233240_5a3191] agent.tool_executor: Tool kanban_show returned error (0.00s): {"error": "task_id is required (or set HERMES_KANBAN_TASK in the env)"}
2026-08-07 23:32:55,743 INFO [20260807_233240_5a3191] agent.tool_executor: tool read_file completed (0.03s, 958 chars)
2026-08-07 23:33:01,603 INFO [20260807_233240_5a3191] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=4/500 budget=4/500 tool_turns=3 last_msg_role=assistant response_len=370 session=20260807_233240_5a3191
```

- 23 s wall clock, **958 chars** of tool output, **2 logged tool calls** (one
  failed `kanban_show`, one `read_file`). `tool_turns=3` / `api_calls=4`; only
  those two tool results are logged, and no `search_files`, no `terminal`, no
  Orca delegation appears anywhere in the session's 11 log lines.
- **Verdict: CORRECT.** Both halves, including the `business_entity_id` trap
  and the `CreateNmOrderService` reason for it. Byte-for-byte the same
  `read_file` size (958 chars) as the map-guided (c) run, i.e. it opened the
  same 767-byte note.

## Q3 — "…API key inline in the command… Do I actually need to rotate the key?"

Ground truth (`knowledge/sudo-argv-secrets-authlog.md`, 1509 B): yes, rotate;
`sudo` writes the full argv to `/var/log/auth.log` **and** the journal,
readable by every admin, surviving in rotated copies; and the remediation is
rotation, **not** editing/truncating the logs. The vocabulary trap: the words
"api key" do not occur in that file. The baseline failed here by advising
"remove the old command from history/logs where feasible" — the one action the
note forbids.

Command:

```
$ ~/.local/bin/hermes chat -Q -q "Last week I ran a one-off command on the shared agents VM with an API key inline in the command. Nobody was watching the terminal. Do I actually need to rotate the key? The knowledge base is a git clone at ~/dev/mc-metarepo on this machine."
```

Verbatim output:

```
session_id: 20260807_233303_1b0b27
Yes—rotate it.

“No one was watching” is irrelevant: an inline key can persist in shell transcripts/history and process arguments. If the command used `sudo`, rotation is mandatory: `sudo` records the full command line—including inline environment values—in both `/var/log/auth.log` and the systemd journal, including rotated copies readable by VM admins.

The documented remediation is explicitly “rotate the credential; do not erase the logs”:

`~/dev/mc-metarepo/knowledge/sudo-argv-secrets-authlog.md`, lines 5–22.
=== WALLCLOCK: 38s ===
```

Tool trace (verbatim):

```
2026-08-07 23:33:12,340 WARNING [20260807_233303_1b0b27] agent.tool_executor: Tool kanban_show returned error (0.00s): {"error": "task_id is required (or set HERMES_KANBAN_TASK in the env)"}
2026-08-07 23:33:31,603 INFO [20260807_233303_1b0b27] agent.tool_executor: tool search_files completed (0.04s, 4697 chars)
2026-08-07 23:33:31,769 INFO [20260807_233303_1b0b27] agent.tool_executor: tool terminal completed (0.16s, 111 chars)
2026-08-07 23:33:38,972 INFO [20260807_233303_1b0b27] agent.conversation_loop: Turn ended: reason=text_response(finish_reason=stop) model=gpt-5.6-sol api_calls=5/500 budget=5/500 tool_turns=4 last_msg_role=assistant response_len=518 session=20260807_233303_1b0b27
```

- 38 s, **4,808 chars** of tool output (4697 + 111), **3 logged tool calls**
  (failed `kanban_show`, `search_files`, `terminal`).
- **Verdict: CORRECT.** It carries the auth.log + journal pair, the rotated
  copies, the admin readership, and — the discriminating part — it states the
  forbidden action as forbidden ("rotate the credential; do not erase the
  logs"), which is exactly the sentence the baseline contradicted.
- Note the *route*: no map, so no slug to recognise — it reached the right note
  via `search_files` (4,697 chars of results) rather than a targeted read. The
  vocabulary mismatch that killed naive `rg` (mechanism (a)) did **not** kill
  the agent's own `search_files`: given the correct root path, the model
  searched on meaning, not on "api key".

## Comparison table

Bytes = tool-output chars the model ingested. `(d)` and `(c)` figures are as
measured in `kb-retrieval-trials.md`; `(c-minus)` is this trial.

| Mechanism | Q | Verdict | Seconds | Tool-output bytes | Tool calls |
|---|---|---|---|---|---|
| (d) baseline — no pointer, no map | Q1 | PARTIAL (right discriminant, **missed the trap**) | 422 | not itemised; Orca worker read product source, wrote 2 `/tmp` files, left a worktree + branch | 25 |
| **(c-minus) path in prompt, no map** | Q1 | **CORRECT** | **23** | **958** | **2** (1 = failed `kanban_show`) |
| (c) map + targeted read | Q1 | CORRECT | 19 | 958 | 1 (`read_file`) |
| (d) baseline — no pointer, no map | Q3 | **WRONG** (from priors; advises the forbidden log-cleanup) | 22 | 0 (zero retrieval) | 1 (failed `kanban_show`) |
| **(c-minus) path in prompt, no map** | Q3 | **CORRECT** | **38** | **4,808** | **3** (1 = failed `kanban_show`) |
| (c) map + targeted read | Q3 | CORRECT | 21 | 1,717 | 2 (1 = failed `kanban_show`) |

## Verdict

**The pointer alone answers both questions. On this evidence the map is not
justified — the confound resolves in favour of the one-sentence fix.**

- Both previously-failing questions flip to CORRECT with **nothing but the
  correct path**: no slug list, no layout table, no routing rules. Q1's
  422 s / 25-tool / Orca-delegation / git-branch failure collapses to 23 s and
  one `read_file` of the same 958 chars the map-guided run consumed. Q3's
  confidently-wrong 22 s answer becomes correct in 38 s.
- **Every correctness win previously credited to mechanism (c) on Q1 and Q3 is
  attributable to the pointer**, not to the map. (c)'s 5/5 must therefore be
  read as "5/5 for *a correct path*, of which the map contributed no measured
  correctness on the 2 questions that were fired live".
- What the map *did* buy, measured: on Q3, **2.8× fewer tool-output bytes**
  (1,717 vs 4,808) and **17 s** (21 vs 38), because a slug hit replaces a
  `search_files` sweep with a direct `read_file`. On Q1 it bought **nothing
  measurable** — identical 958 bytes, 19 s vs 23 s, inside run-to-run noise.
  That is an efficiency delta on one of two questions, at n=1, against a
  hand-maintained artifact that `kb-retrieval-trials.md` itself shows decays
  (`knowledge/README.md` is missing 5 of 40 notes).
- **P6 should lead with the pointer fix and treat the map as unproven.** The
  first item is one sentence in SOUL.md naming `~/dev/mc-metarepo` (and
  deleting or correcting the dead `~/dev/gaetan-metarepo` reference). Ship
  that, then re-measure before proposing anything else.

**What the map would need to earn its keep** — none of it measured today:

1. **A question the pointer alone gets wrong and the map gets right.** That is
   the only evidence that settles it. Candidates: a note whose slug carries
   little signal (`done-is-not-exercised`, `ops-list-reconcile-before-bulk`),
   where `search_files` on meaning has nothing to grab; a Q4-style question
   whose answer is spread over `learnings/README.md` + `CONTRIBUTING.md` +
   `CLAUDE.md`; a question routing to `schemas/` or `index/`.
2. **A measured cost blow-up the map's *rules* prevent, not its slug list.**
   The two rules with real teeth are "the submodules are empty — do not
   delegate a code-reading session" (this is precisely what caused baseline
   Q1's 422 s Orca delegation) and "never read whole: `next-bdc.fiches.md`
   152 KB, `schemas/*.full.md`, `learnings/mobileclub/workspace.md` 70 KB".
   Neither fired here: with the pointer, Q1 did not delegate and Q3 read
   nothing large. A question that *looks* like it needs product source is the
   test that could still vindicate the map — and if it does, the justification
   is the **anti-delegation rule**, which is 2 lines, not a 3.3 KB generated
   slug map.
3. **A repeat count that survives noise** — ≥3 runs per cell per mechanism,
   given a 19-vs-23 s difference is currently indistinguishable from jitter.

If (1) never produces a case, the honest P6 is: fix the pointer, add at most
the two rules from (2) to SOUL.md, and drop the skill map entirely.

## Not tested

- **The real thing: a path in SOUL.md.** Forbidden this session (a peer owns
  that file). The proxy puts the path at the end of the user turn; SOUL puts it
  in the persona, on every turn, whether or not Gaetan mentions the KB. A
  SOUL-borne pointer could be weaker (buried in 5 KB of persona) or stronger
  (always present). **Untested, and it is the change P6 would actually make.**
- **Q2, Q4, Q5 under (c-minus).** Only the two failing baseline questions were
  re-fired. Q4 in particular is the case where the map's routing was argued to
  matter most, and it remains unfired under *every* agent mechanism except the
  baseline.
- **Repeats.** n=1 per cell; no variance estimate.
- **The map as an installed skill.** Still never written to
  `~/.hermes/skills/`; whether Hermes' relevance judgement would even load it
  unprompted is untested (carried over from `kb-retrieval-trials.md`).
- **Slack surface.** Both turns were `platform=cli`.
- **Whether the pointer sentence changes behaviour on a question the KB cannot
  answer** — e.g. whether it now reads the clone instead of correctly saying "I
  don't know" or correctly delegating. No negative control was run.
- **LCM precedence.** Q5's baseline answered from `lcm.db` without touching the
  KB; nothing here tests whether a pointer changes that.
