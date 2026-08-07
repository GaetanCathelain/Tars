# WF3 section 4 — Slack personal via `korotovsky/slack-mcp-server` (Docker, stdio)

Agent: wiring agent 4 ("slack"). Target: `ssh gaetan@192.168.0.9` (VM `tars`, VMID 102).
Section: `docs/specs/wf3-wiring.md` §4, with the orchestrator's section-specific corrections.
Date: 2026-08-07.

**No secret value appears in this file.** Every credential check below is a boolean, a hash
comparison result, an HTTP code or a redacted field. The only `sops` calls made were per-key
`sops -d --extract '["KEY"]'`, each piped straight over `ssh` into a `0600` file on the VM. No
whole-file `sops -d` in any form, no `--output-type dotenv`.

## Placeholder resolution

| Token | Value | Source |
|---|---|---|
| `<TARS_IP>` | `192.168.0.9` | `status/lane-b.md` B1 (`vmbr0`, VMID 102) |
| `<TARS_USER>` | `gaetan` | `status/lane-b.md` B2 |
| `<TAG>` | `v1.3.0` | resolved live, see Wire 2 |

SOPS keys used: `SLACK_TOKEN`, `SLACK_COOKIE` (both present in `secrets/tars.sops.yaml`, confirmed
by reading the plaintext key names only).

## Preconditions

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes --version'
Hermes Agent v0.20.0 (2026.8.3)   Install directory: /home/gaetan/.hermes/hermes-agent

$ ssh gaetan@192.168.0.9 'docker --version'
Docker version 29.1.3, build 29.1.3-0ubuntu3~24.04.2   (rootless-free, no sudo needed)

$ ssh gaetan@192.168.0.9 'grep -c "^mcp_servers:" ~/.hermes/config.yaml'
0        (no pre-existing block — the edit is a pure append, no key collision)
```

## `--help` verification before scripting (standing caveat R1/R7)

Run on the VM before any non-interactive invocation was scripted:

| Command | Fact established | Spec said |
|---|---|---|
| `hermes mcp --help` | subcommands `serve add remove list test configure login reauth picker catalog install` | `hermes mcp list` — **correct** |
| `hermes mcp list --help` | no flags; **prints configured servers only, does not connect** | spec assumed it shows "connected" — it does not |
| `hermes mcp test --help` | `hermes mcp test <name>` — this is the one that actually connects + discovers tools | not in spec |
| `hermes chat --help` | **`-q/--query` for a single non-interactive query, `-Q/--quiet` for programmatic output. There is no `--oneshot` flag.** | spec wrote `hermes chat --oneshot '…'` — **wrong flag** |
| `hermes tools --help` | `hermes tools list`; MCP tools use `server:tool` notation in the CLI | — |
| `docker run --rm …:v1.3.0 --help` | see Wire 2 | — |

Schema for the `mcp_servers` block verified against the live binary's own source
(`~/.hermes/hermes-agent/tools/mcp_tool.py`, module docstring): stdio = `command` + `args` (+ optional
`env`), HTTP/SSE = `url` (+ `transport: sse`). The spec's stdio stanza shape is correct; the
**`tars-profile.md` §2 draft is not** — it proposes `url: http://127.0.0.1:13080/sse` for a server
named `slack-personal`. This section deliberately follows `wf3-wiring.md` §4 (stdio, name `slack`),
which is what §4's pass-evidence line targets.

## Wire 1 — container env file `~/tars/slack-mcp/.env` (0600)

Built by two per-key extract pipes. `tr -d '\n'` strips any trailing newline from the decrypted
value so the mapping is byte-exact; the value never lands on argv and is never echoed.

```
$ sops -d --extract '["SLACK_TOKEN"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 'umask 077; mkdir -p ~/tars/slack-mcp;
      { printf "SLACK_MCP_XOXC_TOKEN="; tr -d "\n"; printf "\n"; } > ~/tars/slack-mcp/.env'
xoxc_write_exit=0

$ sops -d --extract '["SLACK_COOKIE"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 'umask 077;
      { printf "SLACK_MCP_XOXD_TOKEN="; tr -d "\n"; printf "\n"; } >> ~/tars/slack-mcp/.env'
xoxd_write_exit=0
```

Shape check (keys only, no values):

```
$ ssh gaetan@192.168.0.9 'ls -l ~/tars/slack-mcp/.env; wc -l < …; cut -d= -f1 …'
-rw------- 1 gaetan gaetan 390 Aug  7 17:50 /home/gaetan/tars/slack-mcp/.env
lines=2
SLACK_MCP_XOXC_TOKEN
SLACK_MCP_XOXD_TOKEN
```

Byte-identity check — `sha256` of the SOPS value piped over `ssh` on **stdin** (not argv) and
compared against the `sha256` of the value in the env file on the VM. Only the verdict is printed;
neither the value nor the hash is ever displayed:

```
$ sops -d --extract '["SLACK_TOKEN"]' … | tr -d '\n' | sha256sum | cut -d' ' -f1 \
  | ssh gaetan@192.168.0.9 'read -r h; v=$(grep "^SLACK_MCP_XOXC_TOKEN=" … | cut -d= -f2- \
      | tr -d "\n" | sha256sum | cut -d" " -f1); [ "$h" = "$v" ] && echo XOXC_BYTE_IDENTICAL …'
XOXC_BYTE_IDENTICAL

$ (same for SLACK_COOKIE → SLACK_MCP_XOXD_TOKEN)
XOXD_BYTE_IDENTICAL
```

Value-class checks (booleans only):

```
xoxc_prefix=ok                          (matches ^SLACK_MCP_XOXC_TOKEN=xoxc-)
xoxd_prefix=ok                          (matches ^SLACK_MCP_XOXD_TOKEN=xoxd-)
xoxd_contains_percent_escapes=true      (stored URL-ENCODED, confirming the A4 correction to D1)
xoxd_urlsafe_per_slackdump_regex=true   (no character outside [-._~%A-Za-z0-9] — see Wire 3)
```

`~/.hermes/.env` was **not** touched by this agent (agent 2 owns it). Its mtime `17:50:50` and size
`1423` are agent 2's writes, not mine — every write from this section went to
`~/tars/slack-mcp/.env` only.

**PASS.**

## Wire 2 — image pin + the eager-cache switch (D1 binding constraint, korotovsky #86)

```
$ ssh gaetan@192.168.0.9 'gh api repos/korotovsky/slack-mcp-server/releases/latest --jq .tag_name'
v1.3.0

$ ssh gaetan@192.168.0.9 'docker pull ghcr.io/korotovsky/slack-mcp-server:v1.3.0'
Status: Downloaded newer image for ghcr.io/korotovsky/slack-mcp-server:v1.3.0

$ ssh gaetan@192.168.0.9 'docker image inspect … --format "{{index .RepoDigests 0}}"'
ghcr.io/korotovsky/slack-mcp-server@sha256:35cbc988d9282409e27b755957e48a6096fcf037dee72118e97177fe38b1a1b3
```

**Tag: `v1.3.0`. Digest: `sha256:35cbc988d9282409e27b755957e48a6096fcf037dee72118e97177fe38b1a1b3`.**

The image's own surface (no tokens needed for `--help`):

```
$ ssh gaetan@192.168.0.9 'docker run --rm ghcr.io/korotovsky/slack-mcp-server:v1.3.0 --help'
Usage of mcp-server:
  -e string           Comma-separated list of enabled tools (empty = all tools)
  -enabled-tools string
  -no-cache           Skip user/channel cache loading on startup for faster initialization.
                      Lookups by #channel-name or @username will not work; use channel/user IDs instead.
  -rod string         Set the default value of options used by rod.
  -t string           Transport type (stdio, sse or http) (default "stdio")
  -transport string
exit=0
```

### THE FLAG: `--no-cache` (equivalently `-no-cache`). Command-line flag only — there is no env-var equivalent at v1.3.0.

Where I looked, and what the source at `v1.3.0` says (shallow clone `git clone --depth 1 --branch
v1.3.0` on the VM, `/tmp/smcp`):

- `cmd/slack-mcp-server/main.go:31` — the flag declaration quoted above.
- `cmd/slack-mcp-server/main.go:74-82` — **the eager crawl is the DEFAULT**:
  ```go
  if noCache {
      p.SkipCache()
      logger.Info("Cache loading disabled via --no-cache flag", …)
  } else {
      go func() { var once sync.Once
          newUsersWatcher(p, &once, logger)()      // → p.RefreshUsers(...)
          newChannelsWatcher(p, &once, logger)() } // → p.RefreshChannels(...)
  }
  ```
- `pkg/provider/api.go:874-930` — `refreshUsersInternal` falls through to `fetchAndStoreUsers(ctx)`
  ("No usable cache: fetch fresh data synchronously (first run or force)") whenever no cache file
  exists. On a `--rm` container there is never a cache file, so **every single start would do a
  full workspace user enumeration** — precisely the #86 burn path.
- `main.go:86-95` — on `stdio` transport the server **blocks** until `IsReady()`, i.e. until that
  crawl finishes. `SkipCache()` (`pkg/provider/api.go:1360-1365`) marks both caches ready without
  fetching anything, so `--no-cache` also removes the startup stall.
- README env-var table at `v1.3.0` (`docs/` + `README.md`, fetched via `gh api …?ref=v1.3.0`):
  `SLACK_MCP_USERS_CACHE` / `SLACK_MCP_CHANNELS_CACHE` are **cache file *paths*, not on/off
  switches** — they cannot disable the crawl. No other cache-related env var exists at this tag.

**Recorded trade-off** (README "Limitations matrix & Cache", and confirmed in
`pkg/handler/channels.go:150+`): with `--no-cache`, `channels_list` reads
`ProvideChannelsMaps().Channels`, which is empty, so that tool returns nothing.
`channels_me` (`pkg/handler/channels.go:237+`) calls `GetConversationsForUserContext` **live** against
the Slack API and is unaffected — which is why probe (b) below still returns a full channel list.
Lookups by `#name`/`@name` are also gone; callers must use IDs. D1 is binding, so the cache stays
off and this limitation is accepted, documented in the config comment.

**PASS** — flag identified by name, from the image's own `--help` and the source at the pinned tag.

## Wire 3 — cookie encoding: does the server re-encode the xoxd value?

Per the orchestrator's correction: `SLACK_COOKIE` is stored **URL-encoded** and works as-stored
(A4). The risk to check is the **server** encoding it a second time.

Chain traced at `v1.3.0`:

- `pkg/provider/api.go:685` → `auth.NewValueAuth(xoxcToken, xoxdToken)` from
  `github.com/rusq/slackdump/v3 v3.1.13` (`go.mod:14`).
- `slackdump/auth/value.go` → `NewValueAuth` → `makeCookie("d", cookie)`:
  ```go
  func makeCookie(key, val string) *http.Cookie {
      if !urlsafe(val) { val = url.QueryEscape(val) }   // CONDITIONAL, not unconditional
      …
  }
  var reURLsafe = regexp.MustCompile(`[-._~%a-zA-Z0-9]+`)
  func urlsafe(s string) bool { return len(reURLsafe.ReplaceAllString(s, "")) == 0 }
  ```

**`%` is inside the url-safe character class.** An already-percent-encoded xoxd value therefore
consists exclusively of safe characters, `urlsafe()` returns true, and `url.QueryEscape` is
**skipped**. Verified against the actual stored value on the VM without printing it:
`xoxd_urlsafe_per_slackdump_regex=true` (Wire 1). Hence **no double-encoding**.

**No flip applied, and none needed.** The one-flip rule was not triggered — the 401 signature it
guards against cannot arise on this code path, and probes (a)–(c) all passed with the as-stored
value. Store stays URL-encoded.

## Wire 4 — `config.yaml` stanza (flock-guarded, backup taken)

One-time backup before any edit:

```
$ ssh gaetan@192.168.0.9 'cp ~/.hermes/config.yaml ~/.hermes/config.yaml.bak-slack'
backup_taken=ok
-rw------- 1 gaetan gaetan 5512 Aug  7 17:50 /home/gaetan/.hermes/config.yaml.bak-slack
```

Stanza staged in a separate file (contains no secrets), then applied by a single flock-guarded
copy → append → **YAML parse validation** → atomic `mv`. The parse gate means a syntax error can
never reach the live-reloading file:

```
$ ssh gaetan@192.168.0.9 'flock ~/.hermes/.wf3.lock -c "
    cp -p ~/.hermes/config.yaml ~/.hermes/config.yaml.wf3new
    && cat ~/tars/slack-mcp/stanza.yaml >> ~/.hermes/config.yaml.wf3new
    && python3 -c \"<yaml.safe_load + assert mcp_servers keys == [slack]>\" ~/.hermes/config.yaml.wf3new
    && mv ~/.hermes/config.yaml.wf3new ~/.hermes/config.yaml && echo CONFIG_APPLIED"'
YAML_OK
CONFIG_APPLIED
exit=0
```

Appended block:

```yaml
# WF3 section 4 — Slack personal via korotovsky/slack-mcp-server (stdio, Docker).
# Image pinned: ghcr.io/korotovsky/slack-mcp-server:v1.3.0
# Secrets live ONLY in /home/gaetan/tars/slack-mcp/.env (0600) — never inline here.
# --no-cache is the D1 binding constraint (korotovsky #86): it skips the eager
# users.list / conversations.list workspace crawl at startup, which is what
# invalidates xoxc/xoxd. Trade-off (upstream README limitations matrix): the
# cache-backed `channels_list` tool returns nothing; use `channels_me`, and
# address channels/users by ID, not by #name / @name.
mcp_servers:
  slack:
    command: "docker"
    args: ["run", "-i", "--rm", "--env-file", "/home/gaetan/tars/slack-mcp/.env",
           "ghcr.io/korotovsky/slack-mcp-server:v1.3.0",
           "--transport", "stdio", "--no-cache"]
```

`--transport stdio` is passed explicitly because the image's `CMD` is `["--transport", "sse"]`
(`Dockerfile`, production stage) — passing args replaces `CMD`, and the flag default alone would
not have been reached. **Stdio was used; no SSE/HTTP fallback and no exposed port were needed.**

Read back through Hermes itself:

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes config get mcp_servers'
slack:
  command: docker
  args: [run, -i, --rm, --env-file, /home/gaetan/tars/slack-mcp/.env,
         ghcr.io/korotovsky/slack-mcp-server:v1.3.0, --transport, stdio, --no-cache]
```

**PASS.**

## Probe (a) — server sees the workspace

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes mcp list'
  Name             Transport          Tools   Status
  slack            docker run -i      all     ✓ enabled
exit=0
```

`hermes mcp list` is registry-layer only (verified via `--help`: no flags, no connection). The
connection + tool discovery the spec's pass-evidence line actually asks for is `hermes mcp test`:

```
$ ssh gaetan@192.168.0.9 '~/.local/bin/hermes mcp test slack'
  Testing 'slack'...
  Transport: stdio → docker
  Auth: none
  ✓ Connected (855ms)
  ✓ Tools discovered: 18
    channels_list  channels_me  conversations_history  conversations_join
    conversations_leave  conversations_mark  conversations_replies
    conversations_search_messages  conversations_unreads  saved_clear_completed
    saved_list  saved_update  usergroups_create  usergroups_list  usergroups_me
    usergroups_update  usergroups_users_update  users_search
exit=0
```

The **855 ms** connect is independent runtime proof that `--no-cache` is in effect: on the stdio
path the server blocks until `IsReady()`, so a cold full-workspace user+channel crawl (this
workspace has 40+ channels and a far larger user table) could not have completed in under a second.

Tool-name form — the spec's `mcp_slack_*` is stale. The live binary uses
`mcp__<server>__<tool>` (`tools/mcp_tool.py:5762-5770`, `MCP_TOOL_NAME_PREFIX = "mcp__"`,
`_MCP_NAME_DELIM = "__"`), so the 18 tools above surface as `mcp__slack__channels_me` etc. Same
server, same tools, double-underscore delimiter.

**PASS** (connected, 18 `mcp__slack__*` tools discovered).

## Probe (b) — one real, minimal tool call through Hermes

Spec's `--oneshot` does not exist; the verified equivalent is `-q` + `-Q`.

```
$ ssh gaetan@192.168.0.9 'timeout 300 ~/.local/bin/hermes chat -Q -q "list my Slack channels"'
#ops-product
#cleaq-jobs-alert
#cleaq-from-loop-automations
#migration-b2b-mc-to-cleaq
#ci-ganesh
… (40 channels returned; full list withheld here only for length, nothing secret in it) …
#help-help-tech
#mobile-club
exit=0
```

Non-empty channel list, real workspace names, through the Hermes model loop and the MCP server.
The model reached for the live `channels_me` path (the one that survives `--no-cache`) rather than
the cache-backed `channels_list`, which is exactly the behaviour the Wire 2 trade-off predicted.
No provider/model/auth error — agent 5's `openai-codex` backend override was live
(`hermes config get model.provider` → `openai-codex`).

**PASS.**

## Probe (c) — the token-survival guard, run AFTER (b)

The spec's inline form puts both secrets on `curl`'s argv. Replaced with the equivalent `-K`
config-file delivery (`cookie =` ≡ `-b`, `form =` ≡ `-F`) written by a shell heredoc, so neither
value ever appears in a process argument list. The config file is `mktemp`'d under `umask 077` and
`shred -u`'d in an `EXIT` trap. Script: `~/tars/slack-mcp/authtest.sh` (contains no secrets — it
sources the 0600 env file).

```
$ ssh gaetan@192.168.0.9 'sh ~/tars/slack-mcp/authtest.sh'
http_status=200
ok=True
error=None
user=gaetan.cathelain
user_id=U08BDJAMSRZ
team=Mobile Club
team_id=T7V1UGJ82
url=https://mobileclub-squad.slack.com/
```

**`ok:true` after (b). No `invalid_auth`. The pair survived the MCP call.**

Identity captured for DECISION Blocker 2 / `SLACK_ALLOWED_USERS` at cutover:

| Field | Value | Expected | Match |
|---|---|---|---|
| `user_id` | `U08BDJAMSRZ` | `U08BDJAMSRZ` | **yes** |
| `user` (handle) | `gaetan.cathelain` | — | matches A4 |
| `team_id` | `T7V1UGJ82` | — | matches A4 |

No mismatch to flag. (Note on wording: Slack's `auth.test` returns the *handle* in `user` and the
member ID in `user_id`; the orchestrator's "expected `U08BDJAMSRZ` in `user`" is satisfied by
`user_id`, identical to the A4 evidence.)

**PASS.**

### Container-lifetime recheck

`--rm` stdio containers exit when the calling Hermes session ends, so nothing should persist.
Immediately after probe (c):

```
$ ssh gaetan@192.168.0.9 'docker ps -a'
gifted_yalow        mcp/notion                                    Up 5 seconds
eloquent_lumiere    ghcr.io/korotovsky/slack-mcp-server:v1.3.0    Up 5 seconds
```

Both are **transient containers spawned by a concurrently running Hermes session** (agent 2's
Notion work — its own `mcp/notion` container is up alongside; any `hermes` invocation starts every
configured stdio MCP server for the life of that session). Because a slack container *was* running
at that moment, the spec's 5-minute recheck was executed rather than skipped:

```
$ ssh gaetan@192.168.0.9 'sleep 300; docker ps -a; sh ~/tars/slack-mcp/authtest.sh'
=== T+5min docker ps -a ===
(nothing above = drained)
=== T+5min auth.test ===
http_status=200
ok=True
error=None
user=gaetan.cathelain
user_id=U08BDJAMSRZ
team=Mobile Club
team_id=T7V1UGJ82
url=https://mobileclub-squad.slack.com/
```

**No slack-mcp container remains** (`docker ps -a` empty — the transient ones exited with `--rm`
when the concurrent session ended; nothing was left running by this section). **`auth.test` still
`ok:true` five minutes later, same `user_id`** — the credential pair is intact after a full
connect + tool-discovery + real tool call cycle. This is the #86 guard passing, not a formality.

## Invariants re-checked

| Invariant | Evidence |
|---|---|
| p-Hermes (`192.168.0.3`) never contacted | no command in this section addresses it; every `ssh` target is `gaetan@192.168.0.9` |
| Gateway unit never enabled/started | `systemctl is-enabled hermes` → `not-found`, `is-active` → `inactive`; this agent ran no `systemctl` write |
| `~/.hermes/auth.json` never read or copied | present, untouched; only `hermes config get model.provider` was read |
| No browser / OAuth / device flow | none driven; `hermes mcp test` reported `Auth: none` for this server |
| `~/.hermes/.env` untouched by this agent | all writes went to `~/tars/slack-mcp/.env`; agent 2 owns and wrote `.env` at 17:50:50 |
| No `git add/commit/push` | none run |
| `status/lane-a.md`, `status/lane-b.md` not edited | not opened for write |
| No whole-file `sops -d` | only two `--extract` calls, each piped straight into `ssh` |

## Deviations

1. **`hermes chat --oneshot` does not exist** at v0.20.0. Used the verified `-Q -q '<query>'`.
2. **`hermes mcp list` does not connect** — it is registry-layer only. Added `hermes mcp test slack`
   to satisfy the "connected + tools listed" half of the pass-evidence line. Both are reported.
3. **Tool prefix is `mcp__slack__<tool>`, not `mcp_slack_<tool>`** (double underscore, per
   `MCP_TOOL_NAME_PREFIX` in the live source). Spec text is stale; substance unchanged.
4. **Probe (c) uses `curl -K` instead of inline `-b`/`-F`.** Same request, but keeps both secrets
   off the VM's process argument list. Strictly safer, identical semantics.
5. **Server registered as `slack`, not `slack-personal`, and as stdio, not SSE.** Follows
   `wf3-wiring.md` §4 (whose pass-evidence line names `slack`); `tars-profile.md` §2's
   `slack-personal` + `url: …:13080/sse` draft is superseded — that file itself flags the transport
   as "VERIFY".
6. **Explicit `--transport stdio` argument** beyond the spec stanza, required because the image's
   `CMD` is `--transport sse`.
7. **`--no-cache` disables `channels_list`.** Accepted (D1 is binding); documented in the config
   comment. `channels_me` covers the use case live.

## Surprises

- The eager workspace crawl is **on by default and unconditional per container start**. With
  `--rm` there is never a cache file to short-circuit it, so an unflagged stdio stanza would have
  re-enumerated the whole workspace on *every* Hermes session — the #86 burn path, at high
  frequency. The `--no-cache` flag is not an optimisation here, it is the thing that keeps the
  credential alive.
- The double-encoding worry resolves **structurally, not empirically**: slackdump's `makeCookie`
  encodes *conditionally*, and `%` is inside its url-safe class, so an already-encoded value is
  passed through untouched. This is why the as-stored form works for both raw `curl` (A4) and this
  server, and it means the store can stay as it is.
- A concurrent agent's Hermes session spawns this container too — the containers seen in
  `docker ps` were not leaks from this section.

## Verdict

**PASS.**

| Probe | Pass evidence required by §4 | Got | Verdict |
|---|---|---|---|
| (a) | `slack` connected, `mcp_slack_*` tools listed | connected in 855 ms, 18 tools, surfacing as `mcp__slack__*` | **PASS** (naming form corrected) |
| (b) | non-empty channel list in the reply | 40 channels, real workspace names | **PASS** |
| (c) | `{"ok":true,…}` **after** (b), `user` captured | `ok:true`, HTTP 200, `user_id=U08BDJAMSRZ` (expected), again at T+5min | **PASS** |

No rollback triggered: no stanza removed, no `docker rm -f` needed, no cookie flip applied, no
re-harvest. `~/.hermes/config.yaml.bak-slack` is the restore point if the orchestrator wants the
stanza gone.
