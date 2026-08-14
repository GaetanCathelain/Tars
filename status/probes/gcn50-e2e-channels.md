# GCN-50 E2E — round-trip with wire-3 delivered as an MCP CHANNEL (file-inbox), not `claude -p` SendMessage

**Date:** 2026-08-14 (UTC) · **Host:** cooper · **Claude Code:** 2.1.232 · **VM:** `gaetan@192.168.0.9` Hermes v0.20.0
**DM:** `D0BBYNM01BL` · **Lane thread anchor ts:** `1786723101.384829`
**Run:** autonomous — the operator (Tars) role was played by this session; the FRENCH answer was supplied by me.

## VERDICT: **CHANNELS-E2E-PROVEN**

spawn (interactive Orca terminal, filewatch channel MCP loaded) → **wire-2** question (session→operator, Slack) →
operator answer (FRENCH) → **wire-3-via-CHANNEL** (plain file dropped into the watched `inbox/` → filewatch MCP
pushes a `<channel source="filewatch">` event into the live session) → session acted (`greeting.txt` = `Bonjour !`,
French) → completion posted (`[lane:gcn50-e2e-ch] done — wrote greeting.txt in FRENCH`). Causation locked by the
pre-drop `greeting.txt`-absent baseline; provenance airtight — nonce `E2E-CH-f1da94` existed **only** in the file the
server deleted before the session could read it, so the channel is the sole possible source.

- **greeting.txt correctly FRENCH:** YES — 10 bytes, exactly `Bonjour !\n`.
- **completion landed in thread:** YES — `[lane:gcn50-e2e-ch] done — wrote greeting.txt in FRENCH`, ts `1786723323.638989`.
- **channel-delivery latency:** file-drop → greeting.txt written **4.20 s**; file-drop → completion post **8.18 s**.

---

## 1. The wire under test (server + flags)

Reused the q12 filewatch channel server **byte-for-byte** (`status/probes/gcn50-q12-channels.md` §2): stdlib python3,
watches `inbox/`, pushes each new file as `notifications/claude/channel`, then **deletes** it. Fresh copy +
fresh inbox/log at `scratchpad/e2e-ch-channel/` (server.py, inbox/, server.log, mcp.json).

`mcp.json`: `{"mcpServers":{"filewatch":{"command":"python3","args":["…/e2e-ch-channel/server.py"]}}}`

Receiver flags (the mandatory ones from q12): `--mcp-config …/mcp.json --dangerously-load-development-channels server:filewatch`.
Delivered via a launcher wrapper `orc-opus-ch` = `orc-opus` (opus[1m], ultracode, ORCHESTRATION-POLICY appended,
`crossSessionInbound:accept`, `--dangerously-skip-permissions`) **plus** those two channel flags, prompt positional-last.

Server self-check before the run: PASS (capability `experimental.claude/channel` declared; a nonce file in inbox
emitted one `notifications/claude/channel` and was deleted).

Live-run `server.log` (verbatim, the whole file):
```
IN  {"method":"initialize",...,"clientInfo":{"name":"claude-code",...,"version":"2.1.232"},...,"id":0}
OUT {"jsonrpc":"2.0","id":0,"result":{...,"capabilities":{"experimental":{"claude/channel":{}}},"serverInfo":{"name":"filewatch",...}}}
IN  {"method":"notifications/initialized","jsonrpc":"2.0"}
OUT {"jsonrpc":"2.0","method":"notifications/claude/channel","params":{"content":"operator answer: FRENCH\nnonce: E2E-CH-f1da94\nWrite greeting.txt with exactly the single line: Bonjour !\n","meta":{"filename":"answer.txt"}}}
```
Handshake (initialize→capability→initialized) proves the channel was **registered**; the second OUT is the actual
wire-3 push. Server ran as PID of a python3 child of the interactive claude session.

## 2. Spawn handling — the wire-1 gate finding, VALIDATED

The receiver **must** be a real interactive terminal (q12: `claude -p` drops channel events silently). Spawned via the
**Orca launcher path**, not a bare pty:

1. Scratch git repo created **OUTSIDE $HOME**: `…/scratchpad/e2e-ch-repo` (under /tmp), no CLAUDE.md in-tree — the
   deliberate wire-1 mitigation.
2. `orca repo add --path <repo>` (terminal create rejects an unregistered path: `selector_not_found`), then
   `orca terminal create --worktree path:<repo> --title tars-gcn50-e2e-ch --command <abs>/orc-opus-ch` (no prompt).
3. **Startup gates observed, cleared with `orca terminal send --enter`:**
   - **Folder-trust** ("Is this a project you trust? ❯ 1. Yes, I trust this folder") — fired. Accepted.
   - **Channel-dev warning** ("WARNING: Loading development channels … Channels: server:filewatch … ❯ 1. I am using
     this for local development") — fired. Accepted.
   - **"Allow external CLAUDE.md imports?"** — **DID NOT FIRE.** ← this is the validated finding: spawning the repo
     outside $HOME's CLAUDE.md inheritance removes the external-import gate entirely (in the e2e-roundtrip wire-1
     finding it fired because the worktree lived under $HOME and inherited `~/CLAUDE.md`'s external @imports).
4. Because the seed can be dropped while a dialog is up, I spawned **bare** (no `--command` prompt), cleared the two
   gates, then seeded with a one-line brief-pointer via `orca terminal send … --enter` (the orc-tab reseed pattern).
   Seed submitted cleanly; no drop.

Post-startup the session showed: `Opus 5 (1M context) … ultracode`, cwd `…/e2e-ch-repo`, and the banner
`Channels (experimental) messages from server:filewatch inject directly in this session`. (The cosmetic q12 warning
`server:filewatch · no MCP server configured with that name` reappeared, and a transient `⚠ 1 MCP server needs
authentication` — both non-fatal; delivery worked regardless.)

## 3. The round-trip, timed

| step | when (UTC) | evidence |
|---|---|---|
| anchor posted | 16:00 | `hermes send -t slack:D0BBYNM01BL` → ts `1786723101.384829` |
| seed sent to session | 16:00:46 | `orca terminal send --enter` |
| **wire-2** question posted | 16:01:13 | session ran `ssh vm 'hermes chat -Q -q "…compose…" \| hermes send -t slack:D0BBYNM01BL:<ts>'` → reply ts `1786723273.498729`: `[E2E gcn50-ch] Should greeting.txt be written in FRENCH or ENGLISH?` |
| pre-drop baseline | 16:01:55 | `greeting.txt` **ABSENT** in repo (only BRIEF.md, LANE.md, README.md) |
| **wire-3 drop** (as Tars) | **16:01:55.459** | `printf 'operator answer: FRENCH\nnonce: E2E-CH-f1da94\n…' > inbox/answer.txt` |
| inbox consumed | +≤0.5s | server deleted answer.txt, emitted the `notifications/claude/channel` push |
| session acted | **16:01:59.656** | `← filewatch: operator answer: FRENCH nonce: E2E-CH-f1da94 …` → wrote `greeting.txt` = `Bonjour !` |
| **completion** posted | **16:02:03.639** | thread reply ts `1786723323.638989`: `[lane:gcn50-e2e-ch] done — wrote greeting.txt in FRENCH` |

**Latencies:** drop → greeting.txt **4.20 s**; drop → completion **8.18 s** (well within the 120 s budget).

Slack thread (final), verbatim:
```
[E2E gcn50-ch] channel lane                                            (anchor, Tars)
> [E2E gcn50-ch] Should greeting.txt be written in FRENCH or ENGLISH?  (wire-2 Q, Tars)
> [lane:gcn50-e2e-ch] done — wrote greeting.txt in FRENCH              (completion, Tars)
```

`greeting.txt` content (`cat -A`): `Bonjour !$`  → French, correct (`Bonjour !` + one `\n`, 10 bytes).

## 4. Channel-as-wire-3 vs SendMessage-as-wire-3 (test-1, `gcn50-e2e-roundtrip.md`)

| | **CHANNEL (this run)** | **SendMessage `claude -p` (test-1)** |
|---|---|---|
| wire-3 "send" action | a plain file write into `inbox/` (`printf > answer.txt`) | a throwaway `claude -p` that calls the `SendMessage` tool once |
| send-side cost | ~nil — no model, no tokens | ~**6278 tokens / 6.83 s** (a whole model turn) to relay one line |
| latency (send → completion) | **8.18 s** | **16.5 s** |
| latency (send → file written) | **4.20 s** | 12.4 s |
| receiver constraint | **INTERACTIVE ONLY** — headless `claude -p` silently drops channel events (q12, 3/3) | works headless **or** interactive; delivery is via the peer socket |
| persistent server per session | **YES** — one filewatch MCP subprocess lives for the whole session | **NO** — nothing persistent; the relay process is spun on demand and exits |
| moving parts | file write + 1 persistent stdio MCP watcher + the `--dangerously-load-development-channels` opt-in | on-demand `claude -p` + cross-session peer registry (`crossSessionInbound:accept`) |
| prompt-injection surface | **wide** — anything dropped in `inbox/` lands **verbatim** in the model's context as a `<channel>` tag; the drop dir must be access-controlled and the server must gate senders | **narrower** — delivery gated by peer identity + `crossSessionInbound` policy; payload still attacker-controlled text if the sender is untrusted |
| caller gets an ack | no (fire-and-forget; delivery proven receiver-side only) | envelope `"Message sent successfully"` (but not a machine msg_id in the lean config) |

### 3-line bottom-line
1. **Speed/cost:** the channel wins — its "send" is a bare file write (0 tokens, 8.18 s to completion) vs SendMessage
   burning a ~6.8 s / 6278-token `claude -p` per relay (16.5 s to completion).
2. **Constraint/moving parts:** the channel pays for it with an **interactive-only** receiver **and** a persistent
   per-session MCP subprocess plus a `--dangerously-load-development-channels` opt-in; SendMessage needs neither and
   works headless, at the price of a model call per message.
3. **Attack surface:** the channel's `inbox/` is a raw prompt-injection surface (file bytes reach the model verbatim)
   that must be locked down and sender-gated; SendMessage delivery is gated by the peer registry, a narrower surface.

**Takeaway for GCN-50:** the channel is the cheaper/faster inbound (operator→session) transport *when the lane session
is already a long-lived interactive terminal* (which the orchestration model here provides) and the inbox is
access-controlled; SendMessage remains the right choice for headless receivers or when a per-message model turn and
peer-identity gating are acceptable/desirable.

## 5. Cleanup

- Closed the `tars-gcn50-e2e-ch` Orca terminal tab (handle `term_5004f6ba-…`) → killed the interactive claude and its
  child filewatch MCP (PID verified gone).
- Best-effort `orca worktree rm` on the scratch repo; removed scratch dirs `e2e-ch-repo`, `e2e-ch-channel`.
  (`orca repo add` has no `repo rm` counterpart — the repo-registration metadata for the now-deleted /tmp path is left
  dangling and harmless.)
- `rm ~/.hermes/lanes/1786723101.384829.md` on the VM.
- **Left in place:** the Slack `[E2E gcn50-ch]` thread (anchor + wire-2 Q + completion) and this evidence file.
- **No** sops/secret, **no** `~/.hermes/config.yaml`/`.env` edit, no gateway restart. This file was not committed.
