# GCN-50 Q12 probe A — headless `claude -p` ListAgents/SendMessage delivery to a peer session

**Question:** can a headless `claude -p` invocation use ListAgents/SendMessage to
deliver a peer message to a running interactive Claude Code session?

**Target receiver:** `improvements-c8` (expects this probe; no other session messaged).

**Date:** 2026-08-14. Machine: cooper.

## 1. Flag discovery

`claude --help` (relevant excerpts):

```
-p, --print                           Print response and exit (useful for pipes)...
--model <model>                       Model for the current session. Provide an alias
                                       (e.g. 'fable', 'opus', or 'sonnet') or a model's full name...
--dangerously-skip-permissions        Bypass all permission checks. Recommended only
                                       for sandboxes with no internet access.
--allow-dangerously-skip-permissions  Enable bypassing all permission checks as an
                                       option, without it being enabled by default.
--permission-mode <mode>              Permission mode to use for the session
                                       (choices: "acceptEdits", "auto",
                                       "bypassPermissions", "manual", "dontAsk", "plan")
--output-format <format>              (only works with --print): "text" (default),
                                       "json" (single result), or "stream-json"
```

`~/.local/bin/orc-sonnet` (read in full):

```sh
#!/bin/sh
# Sonnet 5 orchestrator, ultracode on, standing ORCHESTRATION POLICY.
POLICY="$HOME/.claude/handoff/ORCHESTRATION-POLICY.md"
[ -r "$POLICY" ] || { echo "orc-sonnet: missing $POLICY" >&2; exit 1; }
exec claude \
  --model 'sonnet[1m]' \
  --settings '{"ultracode":true,"crossSessionInbound":"accept"}' \
  --append-system-prompt-file "$POLICY" \
  --dangerously-skip-permissions \
  "$@"
```

Not used verbatim (interactive-only per its own header comment: "Never pipe or
redirect this command"); its flags (`--model`, `--dangerously-skip-permissions`)
were reused directly on `claude -p` instead. First attempt succeeded, so the
fallback-retry step (task step 3) was not needed.

## 2. Attempt 1 — combined ListAgents + SendMessage, text output

Command (verbatim, timeboxed):

```
timeout 240 claude -p "Call ListAgents. Report the full list. Then call SendMessage with to: 'improvements-c8' and message: '[PROBE gcn50-q12-A] headless -p SendMessage delivery test — reply not needed'. Report the SendMessage tool result verbatim, including any error." --model sonnet --dangerously-skip-permissions
```

Exit code: `0`. stderr: empty.

stdout (verbatim, full):

```
SendMessage tool result (verbatim):

```json
{"success":true,"message":"“[PROBE gcn50-q12-A] headless -p SendMessage delivery test — reply not needed” → improvements-c8 (another Claude session on this machine)","msg_id":"3983426e-0e11-4d4b-8406-0c0d10b15c83"}
```
```

Note: the model's final text reply only surfaced the SendMessage tool result,
not a transcript of the ListAgents call (default `--output-format text` prints
only the last assistant message, not intermediate tool-call bodies). No
permission-hold, no error, no "tool unavailable" — `msg_id` returned means the
runtime accepted and queued/delivered the message.

## 3. Attempt 2 — ListAgents only, json output (to get verbatim tool evidence without re-sending to the target)

Command (verbatim, timeboxed):

```
timeout 240 claude -p "Call ListAgents. Report the full raw tool result verbatim (do not summarize, do not call any other tool)." --model sonnet --dangerously-skip-permissions --output-format json
```

Exit code: `0`. stderr: empty.

stdout (verbatim, full JSON result, `result` field reproduced below):

```
Peer sessions (12):
  helptechs-7a [f355c1]  ·  interactive  ·  idle  ·  started 2h ago
  mc-metarepo-mc-kestra-ed [79dd72]  ·  interactive  ·  idle  ·  started 1h ago
  mc-metarepo-ht-console-gang-59 [e2f93d]  ·  interactive  ·  idle  ·  started 1h ago
  mc-metarepo-everything-else-14 [5ba77b]  ·  interactive  ·  idle  ·  started 1h ago
  nmc-678-console-pointers-44 [805a21]  ·  interactive  ·  idle  ·  started 1h ago
  improvements-c8 [e0ffa8]  ·  interactive  ·  idle  ·  started 1h ago
  gaetan-8e [9acb87]  ·  interactive  ·  idle  ·  started 2h ago
  improvements-18 [c8a801]  ·  interactive  ·  busy  ·  started 1h ago
  nmc-682-precedent-flag-2a [30d16f]  ·  interactive  ·  idle  ·  started 1h ago
  nmc-680-review-gate-c3 [82864a]  ·  interactive  ·  idle  ·  started 1h ago
  mc-4228-09 [506fba]  ·  interactive  ·  idle  ·  started 36m ago
  mc-metarepo-everything-else-7e [603a1d]  ·  interactive  ·  idle  ·  started 1h ago
```

(full JSON envelope: `is_error:false`, `stop_reason:"end_turn"`,
`session_id:"5c5589c4-1d23-4473-84a0-af36333c87ac"`, `subtype:"success"`,
`terminal_reason:"completed"`, `permission_denials:[]`.)

`improvements-c8` is present and listed `idle` — confirms ListAgents from a
headless `-p` process enumerates real peer sessions on this machine, including
the intended target, and that no send happened to it in this attempt (single
tool call, per the constrained prompt).

## 4. Analysis

- **ListAgents**: worked from headless `-p`, no approval prompt, returned the
  live peer roster (12 sessions incl. `improvements-c8`). Tool is available
  and functional non-interactively.
- **SendMessage**: worked from headless `-p`, returned
  `{"success":true, ..., "msg_id":"3983426e-..."}`  — no error, no
  hold-for-approval, no "tool unavailable" message. `permission_denials:[]` in
  the attempt-2 JSON envelope corroborates no permission gate was hit.
- Both calls ran with `--dangerously-skip-permissions`, so this probe shows
  delivery succeeds *when permissions are bypassed* — it does not establish
  whether SendMessage would prompt/hold under default permission mode from a
  headless invocation (untested; `-p` without the flag would have no TTY to
  answer a prompt, which is a separate open question, not covered here).
- Whether `improvements-c8` actually *received/surfaced* the message inside
  its own transcript was not independently verified from the receiver side —
  out of scope for this probe (receiver was told to expect it, no reply
  required).

## 5. Verdict

**VERIFIED-SENT**

1. `claude -p ... --dangerously-skip-permissions` successfully called both
   ListAgents (got the full live peer list, `improvements-c8` included) and
   SendMessage (got `success:true` + a `msg_id`), on the first attempt — no
   retry needed.
2. No error, no permission hold, no "tool unavailable" — `permission_denials`
   was empty and both processes exited 0.
3. Caveat: only demonstrated under `--dangerously-skip-permissions`; behavior
   under default/interactive permission mode from a headless process is
   untested and remains open.
