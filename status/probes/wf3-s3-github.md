# WF3 section 3 — GitHub via `gh` CLI + mc-metarepo clone

Agent: wiring agent 3 ("github"). Target: `ssh gaetan@192.168.0.9` (VM `tars`, VMID 102).
Section: docs/specs/wf3-wiring.md §3 (REWRITTEN 2026-08-07 — no PAT, no SOPS entry; A1 device-flow
already authorized in lane A).

## Placeholder resolution

| Token | Value | Source |
|---|---|---|
| `<TARS_IP>` | `192.168.0.9` | status/lane-b.md B1 |
| `<TARS_USER>` | `gaetan` | status/lane-b.md B2 |

No SOPS keys involved in this section (per the 2026-08-07 rewrite note). No config.yaml or .env
touched — mechanical section only.

## Precondition checks

```
$ ssh gaetan@192.168.0.9 'gh auth status --hostname github.com'
github.com
  ✓ Logged in to github.com account GaetanCathelain (/home/gaetan/.config/gh/hosts.yml)
  - Active account: true
  - Git operations protocol: https
  - Token: gho_************************************
  - Token scopes: 'gist', 'read:org', 'repo'
```
PASS — already logged in (A1, lane A device-flow), scopes include `repo` + `read:org`.

```
$ ssh gaetan@192.168.0.9 'which gh; gh --version'
/usr/bin/gh
gh version 2.97.0 (2026-07-31)
```
git present (confirmed earlier in B2 evidence, lane-b.md).

## Wire steps

**1. `gh auth setup-git`**
```
$ ssh gaetan@192.168.0.9 'gh auth setup-git'
exit=0
```
PASS.

**2. `git clone https://github.com/mobile-club/metarepo.git ~/dev/mc-metarepo`**
```
$ ssh gaetan@192.168.0.9 'git clone https://github.com/mobile-club/metarepo.git ~/dev/mc-metarepo'
Cloning into '/home/gaetan/dev/mc-metarepo'...
exit=0
```
PASS — no pre-existing directory (`ls` confirmed absent first).

**3. No `credential.helper store`, no askpass files added.** Verified below (git global config
inspection) — the only credential helper registered for `github.com`/`gist.github.com` is
`!/usr/bin/gh auth git-credential`; nothing was added by this agent.

## Probe (exact 4 lines from spec)

```
$ ssh gaetan@192.168.0.9 'gh auth status --hostname github.com'
github.com
  ✓ Logged in to github.com account GaetanCathelain (/home/gaetan/.config/gh/hosts.yml)
  - Active account: true
  - Git operations protocol: https
  - Token: gho_************************************
  - Token scopes: 'gist', 'read:org', 'repo'

$ ssh gaetan@192.168.0.9 'gh api repos/mobile-club/metarepo --jq .private'
true

$ ssh gaetan@192.168.0.9 'git -C ~/dev/mc-metarepo rev-parse --short HEAD'
4831253

$ ssh gaetan@192.168.0.9 'grep -c "ghp_\|github_pat_\|x-access-token" ~/dev/mc-metarepo/.git/config'
0
(grep exit code 1 — expected per spec note: exit 1 on zero matches, judged by the printed count "0", not exit code)
```

## Pass evidence check (spec: "logged-in status · `true` · a commit SHA · `0`")

| Expected | Got | Verdict |
|---|---|---|
| logged-in status | `✓ Logged in to github.com account GaetanCathelain`, scopes `repo`+`read:org` | PASS |
| `true` | `true` | PASS |
| a commit SHA | `4831253` | PASS |
| `0` (credential grep) | `0` (grep -c printed count; exit code 1 is the documented zero-match behavior, not a failure signal) | PASS |

## Supplementary check — credential helper hygiene

```
$ ssh gaetan@192.168.0.9 'git config --global --list | grep -i credential'
credential.https://github.com.helper=
credential.https://github.com.helper=!/usr/bin/gh auth git-credential
credential.https://gist.github.com.helper=
credential.https://gist.github.com.helper=!/usr/bin/gh auth git-credential
```
Confirms gh IS the sole credential helper (URL-scoped), no `credential.helper store`. `.git/config`
in the clone contains no `[credential]` section — clean.

```
$ ssh gaetan@192.168.0.9 'cat ~/dev/mc-metarepo/.git/config'
[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[remote "origin"]
	url = https://github.com/mobile-club/metarepo.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	merge = refs/heads/main
```
No token, no credential helper override.

## Deviations / surprises

None. Section proceeded exactly per Wire 1-3, no rollback triggered. No SOPS decrypt performed
(none required — hard rule about per-key sops -d / whole-file ban is N/A to this section since it
carries no secrets). config.yaml and .env untouched, as instructed. p-Hermes (192.168.0.3) not
touched. No browser/OAuth/device flow driven by this agent — A1's device flow was already
completed and authorized by Gaetan in lane A prior to this run.

## Verdict

**PASS.** All four probe lines match spec's pass evidence exactly.
