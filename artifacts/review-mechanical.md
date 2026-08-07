# Adversarial review — LENS 1: mechanical correctness

Target: `artifacts/delegate-to-cooper-SKILL.md`
Method: live `--help` + `orca agent-context --json` + `orca skills get orchestration`
on cooper (appVersion **1.4.176**, `runtime.reachable: true`), plus real
double-hop `ssh gaetan@192.168.0.9 → ssh cooper` quoting tests.
Nothing was created, sent, spawned or removed. One scratch file was written to
`/tmp` on cooper for the quoting test and deleted in the same command.

**Score: 6 commands/instructions that would fail or misbehave if executed
literally (1 possibly fatal to the whole design), 4 non-fatal doc errors.**

---

## 0. PATH over non-interactive ssh — VERDICT: NOT A PROBLEM, and the skill's
## claim about it is factually wrong

Measured, VM → cooper, non-interactive:

```
$ ssh gaetan@192.168.0.9 "ssh cooper 'echo \$0; echo \$PATH; command -v orca'"
zsh
/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin
/usr/local/bin/orca
```

`orca` **is** on `PATH` over non-interactive ssh, at `/usr/local/bin/orca` →
`/opt/Orca/resources/bin/orca-ide`. `~/.local/bin/orca` is a 3-line wrapper
`exec '/opt/Orca/resources/bin/orca-ide' "$@"` — same target, and is NOT on the
non-interactive PATH (`~/.local/bin` is absent from it). So:

- Skill line 20–21: *"The CLI is at `~/.local/bin/orca` and is **not** reliably
  on `PATH` over a non-interactive ssh — always spell the full path."*
  The premise is wrong (orca IS on PATH; it's `~/.local/bin` that is not), the
  remedy is right by accident. Tilde expansion is done by cooper's remote zsh,
  verified live: `ssh cooper '~/.local/bin/orca status --json'` returned
  `result.runtime.reachable: true` through the double hop. **No change required
  — but the reasoning is wrong and the skill should not teach a false rule that
  a later author will "fix".** Corrected prose:

  > The CLI is on `PATH` over non-interactive ssh as `/usr/local/bin/orca`;
  > `~/.local/bin/orca` is an equivalent wrapper and also works (the remote shell
  > expands `~`). Either is safe.

- Also verified: cooper's `~/.ssh` alias hop is real — the VM's `~/.ssh/config`
  has `Host cooper → 192.168.0.4, User gaetan, IdentitiesOnly yes`. `ssh cooper`
  works BatchMode, no prompt.

### Shell mechanics over ssh — all VERIFIED WORKING (no findings)

- Hermes runs a command as `bash -c "<whole string>"`
  (`tools/environments/local.py:1500`), so the step-1 **heredoc is fine** — it is
  one string, not line-by-line. `code_execution.timeout: 300` confirmed in
  `~/.hermes/config.yaml:115-116`.
- Backslash-continuation *inside* the single-quoted remote command (steps 2, 4, 5)
  is fine: ssh joins argv with spaces, cooper's shell then treats `\`+newline as
  a continuation.
- `"$(cat /tmp/tars-brief.md)"` expands on cooper and delivers the brief as ONE
  argv element, byte-exact. Tested with a brief containing `"`, `'`, backticks,
  `$HOME`, `$5`, `%`, an em-dash and a stray backslash — round-tripped verbatim.
  Command-substitution output is not re-scanned, so no injection from brief text.
- The sequence is correct as **separate stateless ssh invocations**: the brief
  file persists in cooper's `/tmp` between calls; every other step carries its
  ids on argv.

---

## 1. WOULD FAIL — ranked

### F1 (WORST, possibly fatal to the design) — `check --run <id>` may be unreachable from a terminal-less caller; the skill's response is "STOP" with no working fallback

Skill lines 149–153 and 168–172 build the entire completion signal on
`orca orchestration check --run "<RUN_ID>" --wait`.

Orca's own bundled skill (`orca skills get orchestration`) says, verbatim:

> line 25: *"the coordinator must create or **bind** a Run"*
> line 104: *"New orchestration messages and tasks belong to one explicitly **bound** Run."*
> line 130: `orca orchestration check [--terminal <handle>] [--ack <delivery_id>] …`
>   — **the vendor's canonical `check` usage has no `--run` at all**
> line 139: *"A coordinator `check` returns **the bound Run's** oldest FIFO Delivery"*
> line 138: *"Omit `--from` unless impersonating another terminal; Orca **auto-resolves it from the current terminal**."*

`run-create`'s own summary is *"Create **and bind** a lightweight orchestration
Run"* — bind to the invoking coordinator terminal. **Tars invokes over stateless
ssh with no Orca terminal**, so there is nothing to bind to. This is the exact
shape of the anomaly `orca-v2-sequence-validation.md` §1 item 4 measured live:
`check --run run_legacy_local` → `run_not_found` on a run that `task-list` /
`worker-list` / `gate-list` all resolve with the identical `--run`, and whose
`coordinator_handle` is `null`.

This directly contradicts the task brief's "KEY ESTABLISHED FACT": *"`--run <id>`
works standalone; there is NO prior bind/run-use step required."* That fact is
**not evidenced anywhere** — the only live measurement of it is a failure.

I could not settle it read-only: there is no non-legacy Run on the host and
`run-create` is a mutation. It stays the #1 gate on first E2E.

The skill's handling is inadequate. Line 152 says:

```
#   Expect ok:true (count 0 is fine). If it returns run_not_found, STOP: `check`
#   cannot resolve this run and the blocking wait below will fail the same way.
```

"STOP" is not a plan — it strands the whole skill with no way to know a worker
finished. **Correction: keep the checkpoint, but give it a real fallback**, and
put the fallback in the sequence rather than only in the failure table:

```bash
# ── 3. Cheap checkpoint BEFORE committing to a long wait. --peek: read-only.
ssh cooper '~/.local/bin/orca orchestration check --run "<RUN_ID>" --peek --json'
#   ok:true (count 0 fine) => the blocking wait in step 5 is usable.
#   run_not_found, but `task-list --run <RUN_ID> --json` resolves
#     => `check` cannot bind this Run from a terminal-less ssh caller.
#        DO NOT stop. Fall back to POLLED completion for this whole run:
#          worker-show --dispatch <DISPATCH_ID> --json  → result.state
#        and treat state succeeded|failed|stopped|abandoned as terminal.
#        Say in the reply that the push path is unavailable and I am polling.
```

`worker-show`'s 9-value `state` enum is schema-verified (`CHECK` constraint on
`worker_dispatches.state`), so the polled path is sound. The vendor skill also
blesses it: *"`worker-show --dispatch <id>` says `ready`: keep waiting … proves
`failed` or `stopped` …"* (line 265-267).

### F2 — answering a `question` with `orchestration send` does not unblock the worker

Skill lines 224–228:

```
`orca orchestration send --to dispatch:<DISPATCH_ID> --subject "…" --body "…" --json`
(run `--help` on this one first — its flags are the least verified in this skill)
```

Flags verified: `--to`, `--subject` (**required**), `--body` all exist. So it
will not error. **It will silently not work.** A `question` message is produced
by the worker calling `orca orchestration ask`, which *blocks* until a reply
addressed to that message id. Vendor skill line 240 and 281:

> *"reply to `question` messages with `orca orchestration reply --id <msg_id> --body <answer> --json`"*
> *"Use `ask` for worker-to-coordinator questions; it creates a `question` message that the coordinator answers with `reply`."*

`send --to dispatch:<id>` is *coordinator guidance mail* — the worker only sees
it on its **next** `orchestration check`, which a worker blocked inside `ask`
will never issue. The agent stalls until its `ask --timeout-ms` expires.

Corrected line:

```bash
# The wait returned a `question`: take message.id from the check payload.
ssh cooper '~/.local/bin/orca orchestration reply --id "<MSG_ID>" --body "<answer>" --json'
# `send --to dispatch:<id>` is unsolicited guidance mail, delivered on the
# worker's NEXT check — it does NOT release a worker blocked in `ask`.
```

Verified live: `orca orchestration reply --help` →
`orca orchestration reply --id <msg_id> --body <text> [--from <handle>] [--json]`.

### F3 — the re-attach `check --wait` replays the same Delivery forever (no `--ack`)

`orca orchestration check --help`, Notes, verbatim:

> *"A bound Run replays the same Delivery until `--ack`; process every message before acknowledging."*
> *"default: return the bound Run's oldest unacknowledged FIFO batch."*

The skill **never uses `--ack`, anywhere**. Consequences, in order of damage:

1. Line 216–217 ("block again" on a later turn) will return **instantly** with
   the batch already processed on the previous turn — Tars will report the same
   `worker_done` twice, or report a stale `question` as if it were new, and will
   never see the *next* message.
2. Step 3's checkpoint (line 150) uses the **default** mode, which marks the
   oldest batch read. Harmless at t=0 (empty), but the same command shape is what
   an agent will reach for on re-attach, where it consumes a batch it has not
   processed. Use `--peek` (documented: *"return only unread messages without
   marking them read"*).

Corrected step 5 / re-attach pair:

```bash
# ── 5. Block for completion (first wait — no delivery to ack yet).
ssh cooper '~/.local/bin/orca orchestration check --run "<RUN_ID>" --wait \
  --types worker_done,escalation,question --timeout-ms 240000 --json' 2>/dev/null
#   Capture the Delivery id from the response BEFORE anything else.

# ── 5b. Every subsequent wait MUST acknowledge the batch it already handled,
#        or Orca replays it and the new one never surfaces.
ssh cooper '~/.local/bin/orca orchestration check --run "<RUN_ID>" --ack "<DELIVERY_ID>" --wait \
  --types worker_done,escalation,question --timeout-ms 240000 --json' 2>/dev/null
```

The Delivery id joins RUN_ID / TASK_ID / DISPATCH_ID as a value the skill must
carry in its reply between turns. It is currently not mentioned once.

### F4 — `worker-start` has no `--timeout-ms`; a 300s Hermes kill loses a durable id created by a mutation

Skill lines 156–160 omit `--timeout-ms`. `worker-start --help` lists
`[--timeout-ms <n>]`; **no default is documented and `agent-context --json`
exposes none.** `worker-start` blocks until the worker reaches `ready` (verified:
*"The call exits 0 only for ready"*), and it does worktree creation + setup +
agent launch. If it outruns Hermes' hard 300s `code_execution.timeout`, the tool
call is killed — but the mutation already happened on cooper, so a dispatch
exists that Tars never learned the id of. Every downstream step in the skill is
keyed on that id.

mc-metarepo's setup script is empty (`repo show` → `hookSettings.scripts.setup:
""`), so today this is unlikely — but the skill must not depend on that, and it
must state the recovery.

Corrected:

```bash
ssh cooper '~/.local/bin/orca orchestration worker-start \
  --run "<RUN_ID>" --task "<TASK_ID>" \
  --worktree new-top-level \
  --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 \
  --name <slug> --agent claude --setup run --timeout-ms 200000 --json'
#   --timeout-ms 200000 keeps this inside my 300s cap so I always see the JSON.
#   If the call is killed anyway, DO NOT re-run worker-start (it would spawn a
#   second worker). Recover the id first:
#     ~/.local/bin/orca orchestration worker-list --run "<RUN_ID>" --json
```

### F5 — the failure table tells the agent to read `error.data.validFlags`, which does not exist for a bad *value*

Skill line 256: *"`error.code: invalid_argument` … `error.data.validFlags` is the
fix list — read it, correct the flag, re-run. Do not improvise."*

Measured live on a bad flag **value** (not a bad flag name):

```
$ orca orchestration check --run run_zzz_nope --types totally_bogus_type --json
{"code":"invalid_argument","message":"Invalid --types: totally_bogus_type",
 "data":{"nextSteps":["Fix the command flags or RPC params exactly as described
 by the error message.","Do not retry the same command unchanged."]}}
```

No `validFlags`. An agent following the skill's instruction literally finds the
field absent and has no rule. Corrected row:

| `error.code: invalid_argument` | unknown flag **or** rejected flag value | Unknown flag ⇒ `error.data.validFlags` is the fix list. Rejected value ⇒ there is no `validFlags`; `error.message` names the flag (`Invalid --types: …`) and `error.data.nextSteps` says do not retry unchanged. Re-read `--help` for that command; do not improvise the value. |

(Positive control, same session: `--types worker_done,escalation,question`
passes validation — all three are real message types. `--types` **is** validated
before dispatch, so a typo there fails fast rather than silently never waking.)

### F6 — `/tmp/tars-brief.md` is a fixed path; two delegations in flight corrupt each other

Skill line 135. Two Tars turns spawning work concurrently (or one retried) both
write the same file, and step 2's `$(cat …)` picks up whichever landed last.
Silent wrong-brief, not an error. One-word fix:

```bash
ssh cooper 'cat > /tmp/tars-brief-<slug>.md' <<'BRIEF'
```

and the same `<slug>` in the `task-create` `$(cat …)`.

---

## 2. Non-fatal doc errors (nothing breaks, but the text is wrong)

| # | Line | Claim | Measured |
|---|---|---|---|
| N1 | 20–21 | `orca` is *"not reliably on PATH over non-interactive ssh"* | False. `/usr/local/bin/orca` IS on the non-interactive PATH. See §0. |
| N2 | 131 | `account list --json` → *"an active claude … account with usage headroom"* | No such flat field. Active id is `result.claude.activeAccountId`; usage lives under `result.rateLimits.claude` (`usedPercent`). Name the paths or the agent will not know what to read. |
| N3 | 163–165 | *"Predicted landing: `/home/gaetan/orca/workspaces/mc-metarepo/<slug>`, branch `GaetanCathelain/<slug>`"* | Unverified: `repo show` returns `worktreeBasePath: "null"` (the literal string). Contradicts the skill's own line 53 rule *"Landing path is NOT a fixed pattern — read `path` from the JSON."* Keep the rule, drop the prediction, or mark it clearly as a guess to be overwritten from the `worker-start` response. |
| N4 | 160 | `--name <slug>` unquoted in the template | Fine for a slug, breaks if the model substitutes a multi-word name (splits into a stray positional). Quote it: `--name "<slug>"`. |

## 3. Verified correct — no action

`status --json` → `result.runtime.reachable` (measured `true`, appVersion
1.4.176) · `run-create --objective` · `task-create --run/--task-title/--spec`
(`--spec` required, `--task-title` optional — order in the skill is fine) ·
`worker-start --task/--run/--worktree new-top-level/--repo/--name/--agent/--setup`
(all present; `new-top-level` is a literal enum member; `--agent`/`--terminal`
are the required either-or and the skill picks `--agent claude`) ·
`worker-read --dispatch --limit` · `worker-release --dispatch` ·
`worker-show --dispatch` · `worker-list --run` (no `--status`, no `--all` — skill
is right) · `run-show --id` not `--run` (skill is right) · `inbox` rejects
`--run` (skill is right) · `worktree ps [--limit] [--json]` (no `--repo`; skill
does not pass one) · `worktree list --repo` · `worktree rm --worktree --force` ·
`terminal list --worktree` · `repo add --path` · `repo list` ·
repo `8099e312-3232-46f2-83a9-97aeaf5de5a2` still registered at
`/home/gaetan/dev/mc-metarepo` · `2>/dev/null` on step 5 does drop the stderr
keepalives correctly (they are emitted every 15s per `check --help`; the result
JSON is on stdout) — though it also swallows ssh's own connection errors, so
`| jq 'select(._keepalive|not)'` is the more diagnosable form the vendor
recommends.
