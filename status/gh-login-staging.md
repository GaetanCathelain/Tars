# status — gh device-flow login staging (Tars VM)

Append-only. No secrets in this file, ever.

## 2026-08-07 — gh CLI installed + device-flow login started

**Installed:** `gh` 2.97.0 on VM `192.168.0.9` (`gaetan@192.168.0.9`, key auth from cooper),
via the official GitHub apt repo (`/etc/apt/sources.list.d/github-cli.list`,
keyring at `/etc/apt/keyrings/githubcli-archive-keyring.gpg`). `gh --version` confirms
`gh version 2.97.0 (2026-07-31)`. tmux was already present on the VM (no install needed).

**Login started:** `gh auth login --hostname github.com --git-protocol https --web` launched
inside a detached tmux session named **`ghlogin`** on the VM, stdout+stderr tee'd to
`/tmp/ghlogin.log` on the VM (survives SSH disconnect/exit; not copied off-box).

- Start time: **2026-08-07T17:11:43Z** (VM clock, `Etc/UTC`).
- Poll deadline (GitHub device-flow default 900s window): **2026-08-07T17:26:43Z**.
- One-time code and URL captured from `/tmp/ghlogin.log` (see report to caller — not
  duplicated here since the code is single-use and time-boxed, not a durable artifact).

**Follow-up verify (not run yet):** once authorized, check with
`ssh gaetan@192.168.0.9 tmux capture-pane -t ghlogin -p` or `tail /tmp/ghlogin.log` for the
"Logged in as ..." confirmation line. `gh auth setup-git --help` confirmed the subcommand
exists on this `gh` build; it was **not run** — auth is not complete yet.

**Untouched, per hard rules:** nothing under `~/.hermes`; `~/.config/gh/hosts.yml` was not
read (does not exist until auth completes, and never will be cat'd by this session either way);
hermes gateway not stopped/started/enabled.

## A1b codex OAuth

**Command shape confirmed on VM (`v0.20.0`, matches R7's doc exactly, no drift):** `hermes auth
add openai-codex --type oauth`. Full path used: `~/.local/bin/hermes` (not on non-interactive
SSH `$PATH` by default — `hermes`, `hermes-acp`, `hermes-agent` all live there as wrapper
scripts). `hermes auth add --help` on this build confirms `--type {oauth,api-key,api_key}`,
`provider` positional, plus `--no-browser`/`--label`/`--timeout`/etc. No interactive menu before
the device code — R1's caveat didn't apply here, one command straight to the code screen.

**Login started:** `PYTHONUNBUFFERED=1 hermes auth add openai-codex --type oauth --no-browser`
launched inside a detached tmux session named **`codexauth`** on the VM, stdout+stderr tee'd to
`/tmp/codexauth.log` on the VM (survives SSH disconnect/exit; not copied off-box). `--no-browser`
used deliberately — the VM has no `$DISPLAY`/`xdg-open`, so this forces the pure device-code path
instead of hermes attempting (and failing) to spawn a local browser.

**Buffering gotcha (new, not present in the gh case):** first launch attempt with plain
`tmux new-session -d ... | tee log` produced zero bytes for 45+ seconds even though the process
was alive and sleeping on network I/O — Python fully-buffers stdout when it isn't a tty (piped
into `tee`), unlike `gh`'s Go binary which flushed fine unbuffered. Fix: prefix the command with
`PYTHONUNBUFFERED=1`. Worth carrying forward to any future Python-CLI tmux staging in this repo.

- Start time: **2026-08-07T17:24:45Z** (VM clock, `Etc/UTC`).
- Device/user code and verification URL captured from `/tmp/codexauth.log` (see report to
  caller — not duplicated here per the same single-use/time-boxed reasoning as the gh entry
  above; unlike gh's flow, this CLI's own output states no explicit expiry/countdown — R7's doc
  doesn't cite one either, so none is recorded).
- Session left polling ("Waiting for sign-in... (press Ctrl+C to cancel)") — not driven further,
  per instructions: completion/authorization is Gaetan's step, not this session's.

**Untouched, per hard rules:** `~/.hermes/auth.json` was never read, before or during this run
(the device code came from the tmux pane/log, not from auth.json); `config.yaml` `model.*` not
touched (WF3 owns that per R7); gateway not started/enabled/touched.
