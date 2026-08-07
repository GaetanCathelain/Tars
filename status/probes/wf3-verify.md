# WF3 — independent verification pass

Verifier: independent WF3 verifier agent (adversarial, read-only). Date: 2026-08-07.
Target VM: `gaetan@192.168.0.9` (`tars`, VMID 102). Hermes always as `~/.local/bin/hermes`.
Standard: each section's **Pass evidence** line in `docs/specs/wf3-wiring.md`.

**No secret value appears in this file.** Every credential check below is an HTTP code, a boolean,
a length, a key *name*, or a public identifier (member ID, team ID, host-key fingerprint, image
digest).

**Constraints honoured:** no `sops` invocation of any form (whole-file or `--extract`); no `git`
commands; no file edited other than this one. VM interaction was read-only with one disclosed
exception: the two credential re-probes wrote a `curl` response body to `/tmp/.n.$$` / `/tmp/.c.$$`
on the VM and `rm -f`'d it in the same command (needed to count `recordMap` / inner `200`s without
printing the body). No config, `.env`, `.ssh`, repo or systemd state was written.

---

## Summary table

| § | Agent verdict | Verifier verdict | One-line reason |
|---|---|---|---|
| 1 Gmail/Himalaya | BLOCKED | **REFUTED (section not delivered; the BLOCKED self-report is accurate)** | pass line never run; re-probe confirms nothing was wired |
| 2 Linear+Notion+Cal | PASS_WITH_DEVIATIONS | **CONFIRMED** | all 3 credential probes + MCP liveness independently reproduced |
| 3 GitHub | PASS | **CONFIRMED** | all 4 probe lines independently reproduced |
| 4 Slack personal | PASS_WITH_DEVIATIONS | **CONFIRMED** | connect + 18 tools + `auth.test ok:true` independently reproduced |
| 5 Mesh + model | PASS_WITH_DEVIATIONS | **CONFIRMED** | 3 legs exit 0, model triple correct, live inference reproduced |

**Overall: ISSUES** — not because a wired section failed re-probe (none did), but because of two
open items carried below (§ Open issues): the Gmail credential incident is unlogged and the
declared-compromised pair is now live on the VM, and `mcp/notion` runs unpinned `:latest`.

---

## Independent re-probes (raw)

### R1 — `config.yaml` parses, no top-level `platforms:` shadow

```
$ ssh gaetan@192.168.0.9 'python3 -c "import yaml,...; d=yaml.safe_load(open(\"~/.hermes/config.yaml\")) ..."'
PARSE_OK top_level_keys= 25
HAS_TOP_LEVEL_platforms= False
mcp_servers= ['notion', 'slack']
gateway.platforms= ['a2a', 'slack']

$ ssh ... 'grep -c "^mcp_servers:" ~/.hermes/config.yaml'   -> 1     (no duplicate top-level key)
$ ssh ... 'grep -c "^platforms:"   ~/.hermes/config.yaml'   -> 0     (lane B B4 trap: clear)
```
Parses clean; the known top-level-`platforms:` footgun is **not** present; both MCP servers live
under a single `mcp_servers:` key (agent 2's merge-not-append did save agent 4's stanza).
Top-level key count is 25, not the 24 agent 5 reported — the difference is `mcp_servers`, added by
agents 2/4 after agent 5's snapshot. Not a discrepancy.

### R2 — model backend

```
$ ssh ... 'hermes config get model.provider'  -> openai-codex
$ ssh ... 'hermes config get model.default'   -> gpt-5.6-sol
$ ssh ... 'hermes config get model.base_url'  -> https://chatgpt.com/backend-api/codex
```
No `openrouter.ai` anywhere in the three. Matches §5 exactly.

```
$ ssh ... 'hermes auth status openai-codex'   -> openai-codex: logged in
$ ssh ... 'hermes auth list'
copilot (1 credentials):      #1  gh auth token        api_key gh_cli ←
openai-codex (1 credentials): #1  openai-codex-oauth-1 oauth   device_code ←
```
Exactly one `openai-codex` credential, `source: device_code`. The `copilot` row is a different
provider (side effect of §3's `gh` login), as agent 5 said.

Live inference re-run (the §5 pass condition, not just `auth status`):
```
$ ssh ... 'hermes chat -Q -q "Reply with exactly one word: sol"'
session_id: 20260807_180211_893d8f
sol
```
Reproduced independently ~20 min after agent 5's run. ChatGPT tier reaches Sol.

### R3 — MCP servers registered and live

```
$ ssh ... 'hermes mcp list'
  Name     Transport        Tools   Status
  slack    docker run -i    all     ✓ enabled
  notion   docker run -i    all     ✓ enabled

$ ssh ... 'hermes mcp test notion'   ✓ Connected (601ms)   ✓ Tools discovered: 24
$ ssh ... 'hermes mcp test slack'    ✓ Connected (801ms)   ✓ Tools discovered: 18
```
Both counts match the evidence files (24 / 18). The 801 ms slack connect independently corroborates
that `--no-cache` is in effect — a cold full-workspace crawl cannot finish sub-second.

Stanzas as live YAML (verbatim, no secret present):
```json
"slack":  {"command":"docker","args":["run","-i","--rm","--env-file",
           "/home/gaetan/tars/slack-mcp/.env",
           "ghcr.io/korotovsky/slack-mcp-server:v1.3.0","--transport","stdio","--no-cache"]}
"notion": {"command":"docker","args":["run","-i","--rm","-e","NOTION_TOKEN","mcp/notion"],
           "env":{"NOTION_TOKEN":"${NOTION_API_TOKEN}"}}
```
`--no-cache` present · slack image tag-pinned · `-e NOTION_TOKEN` carries no value on argv
(structural proof, independent of agent 2's `ps` snapshot).

Image identity:
```
ghcr.io/korotovsky/slack-mcp-server@sha256:35cbc988d9282409e27b755957e48a6096fcf037dee72118e97177fe38b1a1b3   (matches §4 evidence)
mcp/notion:latest  mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9        (UNPINNED tag — see Open issues)
```

### R4 — gateway unit still disabled AND inactive

```
$ ssh ... 'systemctl --user is-enabled hermes-gateway.service'  -> disabled
$ ssh ... 'systemctl --user is-active  hermes-gateway.service'  -> inactive
$ ssh ... 'systemctl        is-enabled hermes-gateway.service'  -> not-found   (system scope: correct, it is a --user unit)
```
Invariant holds. Confirms agent 5's surprise: the system-scope query is the misleading one; WF4
probes must use `systemctl --user`.

### R5 — `~/.hermes/.env` shape (names only, never values)

```
$ ssh ... 'ls -l ~/.hermes/.env'
-rw------- 1 gaetan gaetan 1423 Aug  7 17:50 /home/gaetan/.hermes/.env      (0600 ✓)

$ ssh ... 'cut -d= -f1 ~/.hermes/.env'
HINDSIGHT_MODE  LINEAR_API_KEY  NOTION_TOKEN_V2  NOTION_FILE_TOKEN  NOTION_SPACE_ID
NOTION_API_TOKEN  GMAIL_ADDRESS  GMAIL_APP_PASSWORD  SLACK_MCP_XOXC_TOKEN  SLACK_MCP_XOXD_TOKEN
```
10 keys, exactly the set agent 2 claimed; lane B's `HINDSIGHT_MODE` preserved; the four
cutover-held Slack keys (`SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`, `SLACK_HOME_CHANNEL`,
`SLACK_ALLOWED_USERS`) absent as specced.

```
$ ssh ... 'ls -l ~/.hermes/.env.new'   -> No such file or directory      (no leftover staging file ✓)
```
Also present and expected: `.env.bak-linear-notion-cal` (21 B — the pre-WF3 `HINDSIGHT_MODE`-only
file), `.env.installer-default`, and four `config.yaml.bak-*` restore points. Nothing stray.

```
$ ssh ... 'ls -l ~/tars/slack-mcp/.env; cut -d= -f1 …'
-rw------- 1 gaetan gaetan 390 …    SLACK_MCP_XOXC_TOKEN  SLACK_MCP_XOXD_TOKEN     (0600 ✓, 2 keys)
```

**Not verifiable by me:** agent 2's and agent 4's "byte-identical to SOPS" hash claims. Confirming
them requires a `sops` decrypt, which is banned for this pass. Recorded as UNVERIFIED-by-design;
the four live credential probes below are the functional substitute and all pass.

### R6 — mesh legs and the `mac` alias

```
$ ssh ... 'awk "/^Host /{h=\$2} h==\"mac\"" ~/.ssh/config'
Host mac
    HostName macbook-pro-de-gaetan.tail6e788b.ts.net       (tailnet name, NOT 192.168.0.x ✓)
    User gcath
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

# from the VM
ssh -o BatchMode=yes -o ConnectTimeout=5 cooper true -> cooper:0
ssh -o BatchMode=yes -o ConnectTimeout=5 mac    true -> mac:0
# from cooper
ssh -o BatchMode=yes -o ConnectTimeout=5 gaetan@192.168.0.9 true -> tars:0
```
All three legs exit 0, BatchMode, no prompt, no output. Mandatory repoint is in place.
Nothing under cooper's `~/.ssh/` was read or written by this pass.

### R7 — `~/dev/mc-metarepo`

```
$ ssh ... 'git -C ~/dev/mc-metarepo rev-parse --short HEAD'          -> 4831253      (matches §3)
$ ssh ... 'git -C ~/dev/mc-metarepo status --porcelain | head -3'    -> (empty, clean tree)
$ ssh ... 'grep -c "ghp_\|github_pat_\|x-access-token\|://.*:.*@" ~/dev/mc-metarepo/.git/config'  -> 0
$ ssh ... 'git -C ~/dev/mc-metarepo config --get remote.origin.url'  -> https://github.com/mobile-club/metarepo.git
$ ssh ... 'gh auth status --hostname github.com'                     -> ✓ Logged in … scopes 'gist','read:org','repo'
```
I widened the credential grep beyond the spec's three patterns to also catch a `user:pass@` URL —
still `0`. Clone is credential-free; gh is the helper.

### R8 — Docker hygiene

```
$ ssh ... 'docker ps -a --format …'   -> (empty, before and after my two mcp tests)
```
No lingering `slack-mcp-server` (or `mcp/notion`) container. `--rm` + stdio lifecycle behaves as
agent 4 described.

### R9 — §2's three credential probes, re-run from the VM by me

Values sourced from `~/.hermes/.env` **on the VM** and handed to `curl` through `-K -` (stdin),
so nothing ever reached an argv or left the VM:

```
--- P1 Linear ---     viewer_id_present= True  len= 36  errors= False        (HTTP 200)
--- P2 Notion loadUserContent ---  http_code=200   recordMap_present=true
--- P3 CalDAV PROPFIND ---         http_code=207   inner_200_count=1  bytes=945
```
Identical to agent 2's numbers, including the 945-byte CalDAV body. §2's pass line
(200+`viewer.id` · bare 200 · 207 with inner `HTTP/1.1 200 OK`) independently reproduced.

### R10 — §4's token-survival guard, re-run by me

```
--- slack auth.test (post-everything) ---
ok= True  error= None  user_id= U08BDJAMSRZ  team= Mobile Club  team_id= T7V1UGJ82
```
`ok:true` again, ~25 min after agent 4's T+5min recheck and after two further MCP container starts
(my `mcp test slack`). The xoxc/xoxd pair is intact — #86 guard holds under repeated use, which is
stronger evidence than a single post-probe check. `user_id` matches DECISION Blocker 2.

### R11 — §1: confirm nothing was wired

```
$ ssh ... 'command -v himalaya'                      -> himalaya: NOT INSTALLED
$ ssh ... 'ls -l ~/.config/himalaya/config.toml'      -> No such file or directory
$ ssh ... 'ls -la ~/.hermes/.secrets'                 -> No such file or directory
```
Agent 1's "no install, no config, no password file, no probe" is true. The section's rollback
hygiene ("delete `~/.hermes/.secrets/gmail_app_password` on the way out") is moot — it was never
created.

---

## Per-section verdicts

### §1 — Gmail via Himalaya → **REFUTED** (as a delivered section)

The evidence file does not claim a pass and does not fake one; it is an honest incident report and
R11 corroborates every negative claim in it. But the section's Pass-evidence line — *one envelope
row on stdout, exit 0* — was never produced, and no fallback probe was run either. Gmail is **not
wired**. The agent's own `BLOCKED` verdict is the accurate label; I mark the section REFUTED so
that nothing downstream reads "confirmed" and assumes mail works.

Carry-forward: the incident's own follow-up list is **not** discharged — see Open issue 1.

### §2 — Linear + Notion + Calendar (+ Notion MCP) → **CONFIRMED**

Pass line demanded: Linear 200 with `viewer.id`; Notion `loadUserContent` bare 200; Calendar 207
with inner `HTTP/1.1 200 OK` — **all three from the VM**. R9 reproduced all three from the VM with
matching values. The MCP layer the agent added beyond the spec (r8 gate) is real: R3 shows
`✓ Connected (601ms)`, 24 tools, and a stanza that keeps the token off argv structurally.

The agent's own honesty checks hold up: it flagged its first "Cleaq" chat reply as weak evidence
(prompt echo) and settled it with an `API-get-self` call returning a workspace name absent from the
prompt. That is the right standard. `NOTION_FILE_TOKEN` is correctly declared unprobed rather than
claimed.

Deviations accepted, all substantive not cosmetic: `-Q -q` instead of the non-existent `--oneshot`;
`mcp test` (not `mcp list`) as the liveness probe; merge-not-append on `config.yaml` — R1 confirms
that merge is what saved agent 4's stanza from a duplicate-key silent drop.

### §3 — GitHub via `gh` + clone → **CONFIRMED**

All four pass-evidence lines reproduced verbatim in R7: logged-in with `repo`+`read:org` · `true`
(via the agent; I re-confirmed the login and the clone) · `4831253` · `0`. The agent's reading of
`grep -c` exit 1 as "zero matches, judged by the printed count" is correct, not a fudge. My widened
grep (adding `://user:pass@`) also returns 0, and the working tree is clean.

### §4 — Slack personal → **CONFIRMED**

Pass line (a)/(b)/(c) all stand. (a) reproduced: connected, 18 tools. (c) reproduced by me at a
third point in time, still `ok:true`, same `user_id`. (b) — the 40-channel list — I did not re-run
(it costs a model turn and a workspace call for no new information); the evidence file's list is
consistent with `channels_me` being the surviving path under `--no-cache`, which R3's stanza and
801 ms connect corroborate.

The eager-cache answer is the most load-bearing deliverable here and it is *sourced*, not asserted:
flag name from the image's own `--help`, plus `main.go:74-82` showing the crawl is the default
branch and `--rm` guaranteeing no cache file to short-circuit it. Same for the cookie question —
resolved structurally from slackdump's `makeCookie` (`%` inside the url-safe class ⇒ no
double-encode) rather than by trial-and-error flipping, which is exactly what the spec's "one flip,
not a search" rule was trying to prevent. Both are correctly recorded as required by D1.

Naming corrections (`mcp__slack__*` not `mcp_slack_*`; server named `slack`, stdio not SSE) are
spec-text staleness, not deviations of substance.

### §5 — Mesh + model → **CONFIRMED**

Pass line: exit 0 on every configured leg with no prompt (R6: 3/3), `openai-codex: logged in` with
one `device_code` credential (R2), and a live model reply (R2, reproduced independently). The
last one is the part that actually closes R7's open question and it is real, not inferred from
`auth status`.

Two findings in this file are worth more than the verdict: the OpenRouter `base_url` **survived**
lane A's OAuth (so the spec's "add the block" phrasing would have silently left Tars pointed at a
keyless OpenRouter), and `TARS_VM_SSH_PRIVATE_KEY` is not in the SOPS store at all — executing
Wire step 1 as written would have overwritten a working key with nothing. Both are corrected in
place, both re-verified above. The `mac` repoint is done and proven same-host by fingerprint.

---

## Secret-leak sweep

Patterns swept over the five evidence files **and over this file**: `xoxc-`, `xoxd-`, `ghp_`,
`github_pat_`, `lin_api`, `ntn_`, `secret_`, `BEGIN OPENSSH PRIVATE KEY`, and any
`[A-Za-z0-9+/=_-]{40,}` run.

```
$ grep -nE 'xoxc-|xoxd-|ghp_|github_pat_|lin_api|ntn_|secret_|BEGIN OPENSSH PRIVATE KEY' wf3-s*.md
wf3-s3-github.md:75   ← the literal `grep -c "ghp_\|github_pat_\|x-access-token"` probe command
wf3-s4-slack.md:100   ← the literal regex `^SLACK_MCP_XOXC_TOKEN=xoxc-` in a boolean check
wf3-s4-slack.md:101   ← the literal regex `^SLACK_MCP_XOXD_TOKEN=xoxd-` in a boolean check
```
All three are pattern text with **no value following**. Not incidents.

```
$ grep -nEo '[A-Za-z0-9+/=_-]{40,}' wf3-s*.md
wf3-s5:131,159,161,163  ← SSH host-key SHA256 fingerprints (public by construction)
wf3-s4:115              ← a GitHub API path
wf3-s4:122,125          ← the slack-mcp-server image digest (public)
```
No secret value in any of the five files. `gho_************` in §3 is gh's own masking, not a
leak. This file re-swept after writing: clean by the same patterns (its only long runs are the same
public digest and fingerprint strings).

**Verdict: no leak in the WF3 evidence files.** Note the distinct, separate incident: §1's leak was
into the *agent's transcript*, not into a file — no file-level remediation applies, and rotation is
still the only fix.

---

## Open issues (neither is a failed re-probe)

**1 — The Gmail credential incident is unlogged, and the declared-compromised pair is live on the
VM.** Agent 1 declared `GMAIL_ADDRESS` / `GMAIL_APP_PASSWORD` transcript-exposed and asked for
(a) a re-mint, (b) re-encryption, (c) a `status/lane-a.md` incident entry. None has happened:
`grep -niE 'gmail|incident|rotat' status/lane-a.md` shows only the pre-existing `token_v2`
incident (closed, risk accepted) and A3's original row. Meanwhile agent 2 delivered the
**unrotated** pair to `~/.hermes/.env` on the VM (its own deviation 4 says so plainly) and used it
for the passing CalDAV probe. So the exposed app password is now provisioned on the VM and in
active use. This is a decision for Gaetan, not a defect I can fix: either rotate (then re-deliver
`.env` per-key — agent 2's merge is idempotent) or accept the risk and log it the way the `token_v2`
exposure was logged. Right now it is neither.

**2 — `mcp/notion` runs unpinned `:latest`.** Digest today is
`sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9`, but the stanza names no
tag, so any `docker pull` silently changes what the gateway executes. Agent 4 pinned its image to
`v1.3.0` + digest; agent 2 did not, and flagged it as its own surprise 3. Pin before cutover.

**Lower-severity, already flagged by the agents, re-confirmed by me and worth carrying to WF4:**
- `tirith` security scanner is enabled but unavailable — a security control running in degraded
  (pattern-matching) mode, predating WF3. Visible on every `hermes chat` run, including mine.
- `rtk` plugin warns on every non-interactive run; the rtk-rewrite plugin is a no-op under the
  gateway's non-login PATH.
- Multi-step/delegated chat turns died with `Codex response remained incomplete after 3
  continuation attempts` while single-tool turns succeed — one WF4 probe before the gateway carries
  real work.
- The hermes units are `--user` units. Any WF4 probe using system-scope `systemctl` will report
  `not-found` and silently mean nothing (R4 reproduces both answers side by side).
- Tars → p-Hermes is deliberately unwired; §5's flag to Gaetan stands before WF4 assumes the leg.
