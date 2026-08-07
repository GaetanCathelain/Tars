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
