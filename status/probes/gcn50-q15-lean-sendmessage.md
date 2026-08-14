# GCN-50 Q15 — leanest headless `claude -p` that still delivers via SendMessage

Machine: cooper. Date: 2026-08-14. Goal: find the leanest `claude -p` invocation
that can still call SendMessage to reach a peer session, to shrink the
Tars→session throwaway-per-message spawn from ~full orchestrator context
toward near-zero.

## Method

- Receiver: interactive `claude` (sonnet, `--settings '{"crossSessionInbound":"accept"}'`,
  `-n q15-receiver`) launched under a tmux pty in a scratch dir with no
  CLAUDE.md, instructed to append any inbound cross-session message to
  `received.txt`. Registered in `~/.claude/sessions/<pid>.json` with
  `"status":"idle"` before any measurement run.
- Each config: `claude -p '<prompt calling SendMessage to:"q15-receiver">'
  --output-format json --dangerously-skip-permissions <config flags>`, timed
  with `/usr/bin/time -v` (wall clock) and read back for `usage.*` token
  counts. Delivery verified per-run by grepping `received.txt` for a unique
  nonce embedded in that run's message — not by trusting the CLI's own
  `"result":"Sent."` text.
- Token metric used: `usage.input_tokens + usage.cache_creation_input_tokens +
  usage.cache_read_input_tokens` ("TOTAL" below) — the actual total prompt
  size processed for that call. The raw `input_tokens` field alone is
  misleading here: heavier configs (A/B/C) hit large warm prompt caches from
  other live orc-* sessions sharing the same prefix, so their raw
  `input_tokens` looks artificially tiny (6–38) while they still paid for
  (and occupied) 150k+ tokens of cache read. Leaner configs (G/H) are small
  enough that no caching kicked in at all, so their full prompt shows up as
  raw `input_tokens`.

## Results

| cfg | model | flags added vs previous | TOTAL tokens | raw input | cache_creation | cache_read | turns | wall time | delivered |
|---|---|---|---|---|---|---|---|---|---|
| A | sonnet[1m] | orc-sonnet baseline (full CLAUDE.md, ultracode+crossSessionInbound settings, ORCHESTRATION-POLICY.md appended, all MCP servers) | 220,054 | 6 | 53,232 | 166,816 | 3 | 17.39s | yes |
| B | sonnet[1m] | + `--strict-mcp-config --mcp-config {mcpServers:{}}` | 180,281 | 6 | 28,002 | 152,273 | 3 | 8.02s | yes |
| C | haiku | B + `--model haiku` | 159,379 | 38 | 22,119 | 137,222 | 4 | 13.16s | yes |
| D | haiku | C + scratch cwd (no CLAUDE.md) | 105,345 | 28 | 17,543 | 87,774 | 3 | 9.13s | yes |
| E | haiku | D, drop `--settings`/`--append-system-prompt-file`, add `--system-prompt` 1-line full override | 80,734 | 28 | 27,453 | 53,253 | 3 | 9.43s | yes |
| F | haiku | E + `--setting-sources ""` (drops user/project/local settings incl. this repo's SubagentStart hooks) | 57,980 | 28 | 3,998 | 53,954 | 3 | 7.38s | yes |
| **G** | haiku | F + `--tools "SendMessage"` (only tool loaded) | **3,870** | 3,870 | 0 | 0 | 2 | **4.73s** | **yes** |
| H | haiku | G + `--effort low` (extra probe, not asked for) | 3,887 | 3,887 | 0 | 0 | 2 | 4.92s | yes |

G is the floor found: every flag stacks additively and SendMessage kept
delivering all the way down. H (effort low) showed no further win — wall
time here is dominated by process/model cold-start, not reasoning depth, so
effort tuning doesn't help this task shape. Did not test `--tools ""`
(known-negative control: would remove SendMessage itself) or `--bare`
(`ANTHROPIC_API_KEY` is unset on this box — this account authenticates via
OAuth/subscription, and `--bare`'s docs state it accepts only
`ANTHROPIC_API_KEY`/`apiKeyHelper`, i.e. it would break auth here).

## Recommended leanest-working invocation (config G)

```sh
claude -p '<prompt instructing SendMessage>' \
  --output-format json \
  --dangerously-skip-permissions \
  --model haiku \
  --strict-mcp-config --mcp-config /path/to/empty-mcp.json \
  --system-prompt "You are a message relay. Call the tool exactly as the user instructs, then stop." \
  --setting-sources "" \
  --tools "SendMessage"
```
run from a cwd with no CLAUDE.md, where `empty-mcp.json` contains
`{"mcpServers":{}}`.

## Reduction vs baseline (A)

- Tokens: 220,054 → 3,870 = **-216,184 tokens, -98.2%**
- Wall time: 17.39s → 4.73s = **-12.66s, -72.8%**

## Caveats

- `--tools "SendMessage"` is the single biggest lever (F→G alone: -54,110
  tokens, drops all built-in tool schemas — Bash/Edit/Read/Grep/etc. — since
  none are needed for a pure message relay). Second biggest: `--setting-sources
  ""` (E→F: -22,754 tokens), which incidentally strips this repo's own
  SubagentStart Ponytail hook injection along with any other project/user
  hooks — fine for a throwaway relay, would NOT be fine for a session meant
  to do real work under this repo's rules.
- Haiku reasoning was not a bottleneck for this fixed-format single-tool-call
  task in either accuracy (100% delivery, 8/8 runs including the sanity
  check) or, once `--tools` was restricted, latency. Before the `--tools`
  restriction, haiku (C, 4 turns) was slower than sonnet (B, 3 turns) at the
  same MCP-stripped config — haiku "thought" more turns/tokens with the full
  tool surface still exposed. That inversion disappears once `--tools` is
  narrowed.
- All 8 runs (A–H) confirmed delivered by nonce match in the receiver's
  `received.txt`; raw file not reproduced here as it contains only the
  synthetic probe payloads.
