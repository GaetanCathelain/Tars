# WF6 check 3 — deterministic `linear_board.py` exercised against LIVE Linear, then deployed

Run date: 2026-08-10, 20:36–20:56 UTC, on the Tars VM (`gaetan@192.168.0.9`, clock UTC).
Operator: Claude Code subagent (cooper), incremental — step N only after step N−1 passed.

**Verdict: STEP 1 PASS · STEP 2 PASS · STEP 3 PASS.** Live at the end: `daily-work-brief`
v1.5.0 + `scripts/linear_board.py`, enabled, gateway not restarted, nothing posted to Slack.

Secret handling: `LINEAR_API_KEY` was never printed, echoed, logged, put on argv, or written
anywhere. The only path used was `set -a; . ~/.hermes/.env 2>/dev/null; set +a` in the ssh
shell that then `exec`s python3 — the sourcing's own stderr was suppressed precisely so a
malformed line could not echo a value. No `sops` was run. No captured output contained a
key-like string.

---

## STEP 1 — script only, no skill deploy

### 1.1 Copy + hash

```
$ ssh gaetan@192.168.0.9 'mkdir -p ~/.hermes/skills/orchestration/daily-work-brief/scripts'
$ scp .../skills/daily-work-brief/scripts/linear_board.py \
      gaetan@192.168.0.9:/home/gaetan/.hermes/skills/orchestration/daily-work-brief/scripts/

repo sha256: b504ef98c160735ce75bdc5de028afe393916dec8be478cb2fcc2939c80fa308
VM   sha256: b504ef98c160735ce75bdc5de028afe393916dec8be478cb2fcc2939c80fa308   MATCH
VM python3:  3.12.3
```

### 1.2 Offline self-test — PASS

```
$ ssh gaetan@192.168.0.9 'python3 ~/.hermes/skills/orchestration/daily-work-brief/scripts/linear_board.py --selftest; echo "EXIT=$?"'
selftest OK
EXIT=0
```

### 1.3 THE SPECIFIC UNKNOWN — the server-side `state: {type: {nin: [...]}}` filter

**It works. First exercise of this comparator on this transport, and it was accepted.**

```
$ ssh gaetan@192.168.0.9 'set -a; . ~/.hermes/.env 2>/dev/null; set +a; \
    python3 ~/.hermes/skills/orchestration/daily-work-brief/scripts/linear_board.py'
EXIT=0
--- stdout (verbatim) ---
**Board** (priority-sorted)
P1 MC-4179 Cliente bloquee pour son echange — Triage
P1 NMC-601 Rotate the leaked GoCardless E1 production token (prod log leak follow-through) — In Progress
P1 NMC-622 Staging deploys are silently broken: Deploy Hasura + Deploy Hasura Admin fail ECS stabilisation, so no api-v2 code or H… — In Progress
P2 GCN-3 engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard — In Progress
P2 GCN-4 daily-work-brief: pull Linear board + render priority-sorted — In Progress
P2 GCN-5 Tars default-team rule: create tickets in GCN unless told otherwise — In Progress
P2 GCN-7 WF6 E2E verification + docs/facts + status log updates — In Progress
P2 MC-4112 Datadog observability E2E for Verdict (ECS + Vercel + Cleaq distributed tracing) — Todo
P2 NMC-497 Block manual "stop invoicing" on Loop for non-tech operators (167 leaking cases) — In Progress
P2 NMC-498 Gate manual stop-invoicing: Hasura permission revoke + api-v2 resolver guard + admin UI — In Progress
P2 NMC-606 A low-confidence investigation automatically triggers an adversarial review pass — In Review
P2 NMC-615 SE never opened a single metarepo learnings PR despite the capability being Done — investigate + fix — In Review
+45 more
Coverage: 2 views, both complete (57 issues)
--- stderr (verbatim) ---
(empty, 0 bytes)
```

sha256 of stdout: `730ca330fd8bb57da6528b796b32c49563bffb3b9dff1e896506958f8d2f2e1c`

**No repo change is needed.** The fallback branch (drop the `state` clause from the
server-side filter) was NOT exercised because it was not needed — the filter is accepted as
written. Both queries (`BoardTeamIssues` with `team.id.eq` + `state.type.nin`, and
`BoardAssignedIssues` with `assignee.isMe.eq` + `state.type.nin`) returned `data` with no
`errors`, so the script never reached the `graphql` sentinel.

### 1.4 Row-by-row reality check — the fabrication test

Independent source (NOT the script, NOT the agent): the claude.ai Linear MCP connector from
cooper, authenticated as Gaëtan Cathelain `4951b192-e49c-4b7e-b491-58c89e66043c` — the same
person `assignee.isMe` resolves to on the VM key (proved below by the exact set match). Two
reads, both paged to `hasNextPage == false`, `includeArchived: false`:

* `list_issues(team=81e7b769-…GCN, limit 250)` → 12 issues, 1 page.
  Non-terminal: **5** — GCN-3, GCN-4, GCN-5, GCN-7, GCN-10.
  Terminal: GCN-1, GCN-2, GCN-6, GCN-9, GCN-12 (completed); GCN-8, GCN-11 (canceled).
* `list_issues(assignee=me, limit 250)` → page 1: 250 rows (178 completed / 25 canceled /
  2 duplicate / 15 unstarted / 15 started / 14 backlog / 1 triage) → 45 non-terminal;
  page 2 (cursor `48c515a9-…`): 47 rows → 12 non-terminal. **Total assigned non-terminal: 57.**
* Union (dedup on key): **57** — every non-terminal GCN issue is also assigned to Gaetan, so
  the GCN view adds 0 new keys. That is exactly the number the script printed.

The 57-row truth set was then fed through the script's own `normalize`/`sort_key`/`render`
(data independent, rendering already proven by the self-test) and diffed against the live
stdout: **`diff` clean, 0 differences.**

| Check | Result | Numbers |
|---|---|---|
| Every rendered identifier exists | **PASS** | 12/12 |
| Every rendered title matches the real issue byte-for-byte | **PASS** | 12/12 (0 invented, 0 paraphrased — previous version invented 7 of 12) |
| Every rendered state name matches | **PASS** | 12/12 (Triage, In Progress ×8, Todo, In Review ×2) |
| No terminal issue rendered | **PASS** | 0 terminal rows. 205 terminal issues were in the raw reads (178 completed + 25 canceled + 2 duplicate) and none survived. Notably GCN-9 is **P1 Urgent and Done** — if the terminal drop had leaked it would have been the top row; it is absent. SE-409 (`duplicate`) also absent. |
| `+N more` = true total − rendered rows | **PASS** | 57 − 12 = **45**, printed `+45 more` |
| Priority ordering 1,2,3,4 then 0-last | **PASS** | rows 1–3 = P1, rows 4–12 = P2. No P3/P4/P0 reached the 12-row cap; the P0-last rule is covered by self-test §4. |
| GCN before other teams within a priority | **PASS** | P2 block: GCN-3, GCN-4, GCN-5, GCN-7, then MC-4112, NMC-497, NMC-498, NMC-606, NMC-615 (team key MC < NMC). P1 block has no GCN issue; MC-4179 precedes NMC-601/622, same rule. |
| `Coverage:` counts match true totals | **PASS** | `2 views, both complete (57 issues)`; independently computed union = 57 |
| Title clipping is faithful, not a paraphrase | **PASS** | NMC-622 raw title is 168 chars; rendered as the first 119 chars + `…` = 120 (`TITLE_CAP`). Prefix identical to source. |

### 1.5 Determinism — two runs, byte-identical

```
run 1 sha256: 730ca330fd8bb57da6528b796b32c49563bffb3b9dff1e896506958f8d2f2e1c  (20:37 UTC)
run 2 sha256: 730ca330fd8bb57da6528b796b32c49563bffb3b9dff1e896506958f8d2f2e1c  (20:38 UTC)
$ diff run1.out run2.out   ->  no output.  stderr 0 bytes on both.  exit 0 on both.
```

Two further runs bracketing the step-3 agent run (20:41 UTC and 20:56 UTC) produced the same
sha256 — 4 runs over 19 minutes, all identical.

**STEP 1: PASS.**

---

## STEP 2 — deploy the skill (v1.5.0)

### 2.1 Backup, recorded

```
$ flock ~/.hermes/.wf3.lock -c 'cd ~/.hermes/skills/orchestration/daily-work-brief && \
    cp -a SKILL.md SKILL.md.bak-check3-20260810T203842Z && chmod 600 .SKILL.md.new && mv .SKILL.md.new SKILL.md'
```

**Rollback file: `~/.hermes/skills/orchestration/daily-work-brief/SKILL.md.bak-check3-20260810T203842Z`**
(9362 bytes, sha256 `69a3c1285e6b54da02c2d5e07f521dc51f7130176c5c9b4dd72d47c3cdd7f992`
— the safe rolled-back v1.0.0, board-free). Pre-existing backups left untouched:
`SKILL.md.bak-wf6-20260810T193012Z` (v1.0.0) and
`SKILL.md.wf6-fabricating-board-20260810T201445Z` (the fabricating version, frozen).

### 2.2 Deployed file — sha256 both sides

```
repo SKILL.md sha256: e55278b0322c799d94d19171c28a48712917353c6c767ee8ad79f853e741314c
VM   SKILL.md sha256: e55278b0322c799d94d19171c28a48712917353c6c767ee8ad79f853e741314c   MATCH
size 18304 bytes, mode 0600
```

### 2.3 `required_environment_variables` did NOT break loading — PASS

This is the first skill in the repo to carry that frontmatter field. 40 s after the write
(Hermes live-reload window):

```
$ hermes skills list
┏━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━━━┳━━━━━━━━━┓
┃ Name                 ┃ Category            ┃ Source    ┃ Trust     ┃ Status  ┃
┡━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━╇━━━━━━━━━━━╇━━━━━━━━━┩
│ daily-work-brief     │ orchestration       │ local     │ local     │ enabled │
```

`enabled`, category still `orchestration` (that value comes from `metadata.hermes.category`
in the frontmatter, so the YAML parsed). No skill-load error in `~/.hermes/logs/*.log` at or
after 20:38 UTC (`grep -i skill` over 20:38–20:49 returned nothing; `errors.log` shows no new
line after the deploy other than pre-existing unrelated `tools.registry` warnings).

Caveat, stated plainly: **`hermes skills list` has no version column**, and
`hermes skills inspect daily-work-brief` fails with `No skill named 'daily-work-brief' found
in any source` (that subcommand resolves hub/tap sources, not local skills). The v1.5.0
evidence is therefore (a) the on-disk frontmatter `version: 1.5.0`, (b) the sha256 match with
the repo file, and (c) step 3 below, where `hermes chat -s daily-work-brief` preloaded the
skill and the run followed v1.5.0's script-driven board instruction. No stronger version
readout exists on this Hermes.

### 2.4 Gateway — did NOT restart

```
before (20:36 UTC)   NRestarts=0  ActiveState=active  ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
after  (20:39 UTC)   NRestarts=0  ActiveState=active  ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
after  (20:56 UTC)   NRestarts=0  ActiveState=active  ActiveEnterTimestamp=Mon 2026-08-10 20:00:38 UTC
```

`ActiveEnterTimestamp` unchanged across the whole exercise → **no restart**. (`NRestarts` is
unreliable and is recorded only for completeness.)

**STEP 2: PASS.**

---

## STEP 3 — spec check 3, the real cron prompt through the agent loop

### 3.1 The prompt

Job `e231e5faf180` "Gaetan daily work brief", `30 8 * * 1-5`, skills `[daily-work-brief]`,
deliver `slack:C0BP2GZUFSR:1786359613.979759`, `enabled_toolsets: null`,
`provider_snapshot: openai-codex`, `model_snapshot: gpt-5.6-sol`. Its `prompt` field verbatim
(407 bytes, read from `~/.hermes/cron/jobs.json`):

> Prepare Gaetan's morning daily by following the attached daily-work-brief skill completely. Use Europe/Paris for all date/time decisions. For the Monday 2026-08-10 run only, include in the Today section a reminder to try the idea from this Instagram reel: https://www.instagram.com/reel/DY42w69iOb4/?igsh=MXh4Z3kyYjE3ZTQ0ZA== . On later runs, do not repeat this reminder unless Gaetan explicitly asks again.

Note for the record: this live cron prompt does **not** match the §6 prompt in SKILL.md
v1.5.0 — the reconciler has not (re)written it. Out of scope for this check; flagged.

Appended for this run only:

> VERIFICATION-RUN OVERRIDE (Gaetan asked for this run): do NOT post, send, schedule or draft anything to Slack — no slack_send_message, no slack_schedule_message, no DM, no channel message, no thread reply, no canvas, under any circumstance. Reading Slack is allowed and required. Return the finished brief text to me as your final message, and nothing else.

### 3.2 How delivery was prevented

The run went through `hermes chat`, **not** `hermes cron run`. `hermes chat` has no delivery
target — the `deliver: slack:…` field belongs to the cron job and is never consulted — so the
only way anything could have reached Slack was an explicit `slack_send_message` tool call,
which the override forbade and which the logs prove did not happen (§3.6).

```
$ nohup setsid ~/.local/bin/hermes chat -Q --accept-hooks -s daily-work-brief \
      -q "$(cat ~/check3-prompt.txt)" >~/check3-agent.out 2>~/check3-agent.err </dev/null &
launched 20:41:46 UTC   session_id: 20260810_204146_6276b8   model=gpt-5.6-sol provider=openai-codex
finished ~20:49:46 UTC (8m00s, 13 API calls)
```

### 3.3 The agent's output, verbatim

```
  ⟳ compacting context…
**Board** (priority-sorted)
P1 MC-4179 Cliente bloquee pour son echange — Triage
P1 NMC-601 Rotate the leaked GoCardless E1 production token (prod log leak follow-through) — In Progress
P1 NMC-622 Staging deploys are silently broken: Deploy Hasura + Deploy Hasura Admin fail ECS stabilisation, so no api-v2 code or H… — In Progress
P2 GCN-3 engagement-checker: Linear push+pull, SSoT=GCN, nag-loop guard — In Progress
P2 GCN-4 daily-work-brief: pull Linear board + render priority-sorted — In Progress
P2 GCN-5 Tars default-team rule: create tickets in GCN unless told otherwise — In Progress
P2 GCN-7 WF6 E2E verification + docs/facts + status log updates — In Progress
P2 MC-4112 Datadog observability E2E for Verdict (ECS + Vercel + Cleaq distributed tracing) — Todo
P2 NMC-497 Block manual "stop invoicing" on Loop for non-tech operators (167 leaking cases) — In Progress
P2 NMC-498 Gate manual stop-invoicing: Hasura permission revoke + api-v2 resolver guard + admin UI — In Progress
P2 NMC-606 A low-confidence investigation automatically triggers an adversarial review pass — In Review
P2 NMC-615 SE never opened a single metarepo learnings PR despite the capability being Done — investigate + fix — In Review
+45 more
Coverage: 2 views, both complete (57 issues)

**Since the last daily**
- **Monitoring/security:** shipped nested HTTP-secret masking and monitoring rulings; GoCardless rotation remains open. A genuine Google Workspace suspicious-login alert is still uninvestigated.
- **Infrastructure/support:** restored Quentin's Datahub production access; fixed and deployed the HT console submodule-permission regression. Loop development jobs were repaired, the ingress fix merged, and ArgoCD automation restored; the development cluster then lost schedulable capacity and remained `Progressing`/HTTP 503. The production return-label alert still lacks a Datadog investigation.
- **Tars/operations:** corrected Slack tool visibility without restricting capabilities, merged and deployed it; advanced daily threading and native Linear/GCN integration. Forwarded 65 historical Anthropic invoices to Oli's billing alias, delivered the AI-pricing report, and completed Jamil's Verdict OAuth client.
- **Claude/Orca:** completed HT console (`run_96604c35a15f`), Loop jobs (`run_2bdb3d17c898`), ingress durability (`run_7e53c1c9e1da`), and Slack visibility (`run_11159f96b51a`). Active: `Tars/wf6-vm-skills` and `Tars/improvements`; Graphify benchmark is parked.

**Stats**
- PRs merged: 19 — Tars #35–41; support-engineer #163–167/#169; workspace #12287/#12291; metarepo #193/#197; datahub #72; helm-charts #121.
- Linear tickets completed: 8 — GCN-1/2/6/9/12, SE-680/681/683.

**Today**
1. Investigate the suspicious Workspace login and confirm the old GoCardless token is revoked.
2. Close the production return-label alert; follow up with Ahmad/Alexis on workspace PR #12050 and investigate development-cluster capacity.
3. Resume WF6's deterministic Linear-board verification; leave Graphify parked.
4. Try the idea from https://www.instagram.com/reel/DY42w69iOb4/?igsh=MXh4Z3kyYjE3ZTQ0ZA==
5. **Oli:** no new report; the pricing report is delivered. Update him when GoCardless revocation is confirmed.

Coverage: Calendar unavailable; email, PayFit leave, Slack, Cooper, GitHub and Linear were checked.
```

(stderr held only `session_id: 20260810_204146_6276b8`.)

### 3.4 THE BYTE COMPARISON — agent block vs script's direct output

Reference: the script run directly at 20:41:10 UTC, 36 s before the agent was launched
(`run3`), and again at 20:56 UTC after it finished (`run4`). Both identical to each other, so
the board did not move under the agent run.

```
agent block (extracted from the agent's message, 1298 bytes)
  sha256 730ca330fd8bb57da6528b796b32c49563bffb3b9dff1e896506958f8d2f2e1c
script direct, 20:41:10 UTC (run3, 1298 bytes)
  sha256 730ca330fd8bb57da6528b796b32c49563bffb3b9dff1e896506958f8d2f2e1c
script direct, 20:56 UTC   (run4, 1298 bytes)
  sha256 730ca330fd8bb57da6528b796b32c49563bffb3b9dff1e896506958f8d2f2e1c

diff agent_block script_direct  ->  no output.  BYTE-IDENTICAL, 0 differing characters.
```

**This is the fabrication test and it passed.** Stronger than asked for: the run compacted its
context **twice** mid-flight — `agent.conversation_compression` at 20:46:05 and 20:47:28 UTC,
and the agent even printed `⟳ compacting context…` — which is the exact condition that
produced the 2026-08-10 fabrication (7 of 12 invented titles, 2 terminal rows, `+2 more`
against a true remainder of ~46). Under the same condition, the v1.5.0 script path emitted the
board unchanged to the byte.

### 3.5 Other PASS criteria

| Criterion | Result |
|---|---|
| Brief OPENS with the board block | **PASS** — `**Board** (priority-sorted)` is the first line of the message |
| Block byte-identical to the script's direct output | **PASS** — §3.4 |
| Carries the `Coverage:` line | **PASS** — `Coverage: 2 views, both complete (57 issues)` closes the block |
| Brief still includes the Orca/Claude-history section | **PASS** — the labelled `**Claude/Orca:**` line closes *Since the last daily*, with four `run_…` handles and the two active worktrees |

### 3.6 Nothing posted to Slack — PASS

* Reporting conversation `C0BP2GZUFSR` thread `1786359613.979759`, read from cooper with the
  claude.ai Slack connector, before **and** after the run with `oldest=1786390000` (19:26 UTC):
  `No thread messsages` both times. Nothing was posted there in the window.
* Session log `20260810_204146_6276b8` (110 lines in `agent.log`): the only Slack tool names
  that appear are `slack__conversations_history` (3), `slack__conversations_replies` (6),
  `slack__conversations_search_messages` (2), `slack__users_search` (3) — **all read-only**.
  No `chat_postMessage`, no `slack_send_message`, no `slack_schedule_message`, no draft.

**STEP 3: PASS.**

---

## Rollback trigger — not fired

None of the trigger conditions occurred: the skill loaded and stayed `enabled`; the gateway's
`ActiveEnterTimestamp` never moved; the agent's block matched the script byte for byte; no
terminal issue rendered; nothing posted to Slack. The `.bak` was therefore not restored.

If it ever needs to be:

```
ssh gaetan@192.168.0.9 "flock ~/.hermes/.wf3.lock -c \
  'cd ~/.hermes/skills/orchestration/daily-work-brief && cp -a SKILL.md.bak-check3-20260810T203842Z SKILL.md'"
# expect sha256 69a3c1285e6b54da02c2d5e07f521dc51f7130176c5c9b4dd72d47c3cdd7f992
```

## Live state at the end of this run

| Path (VM) | State |
|---|---|
| `~/.hermes/skills/orchestration/daily-work-brief/SKILL.md` | **v1.5.0**, 18304 B, sha256 `e55278b0…`, mode 0600, enabled |
| `~/.hermes/skills/orchestration/daily-work-brief/scripts/linear_board.py` | present, sha256 `b504ef98…` (= repo), selftest OK, live run exit 0 |
| `~/.hermes/skills/orchestration/daily-work-brief/SKILL.md.bak-check3-20260810T203842Z` | v1.0.0 rollback, sha256 `69a3c128…` |
| `hermes-gateway.service` | active since 2026-08-10 20:00:38 UTC, not restarted |
| cron job `e231e5faf180` | untouched — still active, next run 2026-08-11T08:30+02:00, deliver unchanged |

Scratch artifacts (not committed, session scratchpad):
`run1.out`/`run2.out`/`run3.out`/`run4.out`, `expected.txt`, `agent.out`, `agent_block.txt`,
`check3-prompt.txt`. VM-side: `~/check3-agent.out`, `~/check3-agent.err`, `~/check3-prompt.txt`.

## Open items surfaced, not acted on

1. The live cron prompt for `e231e5faf180` (§3.1) predates SKILL.md v1.5.0 §6 and does not
   carry the "board is the script's stdout, byte for byte" instruction. The skill body carries
   it, and the run proved the skill body is enough — but the two have drifted and §6 says the
   reconciler must keep them aligned.
2. `hermes skills list` exposes no version and `hermes skills inspect` cannot resolve a local
   skill, so "the loaded skill is v1.5.0" is only provable by file hash plus behaviour.
3. The repo working tree still holds `linear_board.py` and `SKILL.md` v1.5.0 uncommitted; this
   probe ran no git command and touched nothing under `skills/`.
