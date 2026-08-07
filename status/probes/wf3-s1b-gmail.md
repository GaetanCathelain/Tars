# WF3 agent 1b (gmail) — evidence

Re-run of the aborted `wf3-s1-gmail.md` attempt. Clean run: no whole-file `sops -d` in any form,
only per-key `--extract` piped straight over ssh. No secret value appears in this file, in any
command line on cooper, or in the transcript.

- Target: `gaetan@192.168.0.9` (VM 102 `tars`, Ubuntu 24.04, x86_64) — reachable, key auth, exit 0.
- Date: 2026-08-07.
- Verdict: **PASS_WITH_DEVIATIONS** (probe passes; the spec's config block and probe flags are
  v1-era and had to be re-derived against the installed v2.0.0 — see Deviations).

---

## Step 1 — install himalaya

```sh
ssh gaetan@192.168.0.9 'sudo -n true'                  # rc=0 → passwordless sudo available
ssh gaetan@192.168.0.9 'sudo -n snap install himalaya' # rc=1: error: snap "himalaya" not found
```

Snap loses. Official release binary wins:

```sh
ssh gaetan@192.168.0.9 'curl -sSL https://api.github.com/repos/pimalaya/himalaya/releases/latest | grep tag_name'
# "tag_name": "v2.0.0"
ssh gaetan@192.168.0.9 'mkdir -p ~/.local/bin ~/tmp-hima && cd ~/tmp-hima &&
  curl -sSL -o h.tgz https://github.com/pimalaya/himalaya/releases/download/v2.0.0/himalaya.x86_64-linux.tgz &&
  tar xzf h.tgz && install -m 0755 ~/tmp-hima/himalaya ~/.local/bin/himalaya && rm -rf ~/tmp-hima'
ssh gaetan@192.168.0.9 'sudo -n ln -sf /home/gaetan/.local/bin/himalaya /usr/local/bin/himalaya'
ssh gaetan@192.168.0.9 'himalaya --version'
```

```
himalaya v2.0.0 +smtp +gmail +imap +m2dir +msgraph +jmap +rustls-ring
build: linux musl x86_64
git: nix-flake-20260726001142, rev b1f6dece32c3afc97d44adeb148bf87e89fde140
```

**Which won: release binary, v2.0.0**, `~/.local/bin/himalaya` + symlink `/usr/local/bin/himalaya`.
The symlink is why bare `himalaya` resolves above — a non-interactive ssh shell does NOT get
`~/.local/bin` on PATH (Ubuntu's `~/.profile` only runs for login shells; measured
`PATH_has_local_bin=no`), and Tars' shell tool is a non-interactive shell.

**PASS.**

## Step 2 — deliver the two 0600 files

```sh
sops -d --extract '["GMAIL_APP_PASSWORD"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 'umask 077; mkdir -p ~/.hermes/.secrets; tr -d "\r\n" > ~/.hermes/.secrets/gmail_app_password'
sops -d --extract '["GMAIL_ADDRESS"]' secrets/tars.sops.yaml \
  | ssh gaetan@192.168.0.9 'umask 077; tr -d "\r\n" > ~/.hermes/.secrets/gmail_address'
```

Both rc=0. Verification (metadata only, values never read back):

```
-rw------- 1 gaetan gaetan 28 Aug  7 18:09 gmail_address
-rw------- 1 gaetan gaetan 16 Aug  7 18:09 gmail_app_password
gmail_address       bytes=28  sha256_first8=1b7a9d8c
gmail_app_password  bytes=16  sha256_first8=5003e966
```

16 bytes = a Gmail app password with the spaces stripped, as expected. `tr -d "\r\n"` is deliberate:
`cat` must return exactly the credential, with no trailing newline for the SASL layer to choke on.

**PASS.** Only two `sops` invocations happened this run, both per-key `--extract`, both piped
directly into `ssh`. Nothing decrypted landed on cooper's disk, stdout, or argv.

## Step 3 — verify config schema, then write `~/.config/himalaya/config.toml`

Schema verified BEFORE writing, per the standing R1/R7 caveat. `himalaya account configure` does
not exist in v2 (subcommands are `list` / `check` only), and `himalaya manual` emits CLI man pages
with no config section — so the authority used was the shipped sample for the exact tag:

```sh
ssh gaetan@192.168.0.9 'curl -sSL https://raw.githubusercontent.com/pimalaya/himalaya/v2.0.0/config.sample.toml -o ~/tmp-hima-sample.toml'  # 200
```

The spec §1 block **does not exist in v2.0.0**. Mapping actually used (all keys present in the
v2.0.0 sample):

| spec §1 (v1-era) | himalaya v2.0.0 |
|---|---|
| `backend.type/host/port` | `imap.server = "imaps://host:port"` (scheme carries TLS + port) |
| `backend.login` | `imap.sasl.plain.username` |
| `backend.auth.type = "password"` | `imap.sasl.plain` (SASL mechanism *is* the auth type) |
| `backend.auth.cmd` | `imap.sasl.plain.password.command` |
| `message.send.backend.*` | `smtp.server`, `smtp.sasl.plain.*` |

Parse acceptance of `email` (absent from the v2 sample) was pre-tested with a throwaway config
holding a dummy `@example.com` address — no secret involved:

```sh
ssh gaetan@192.168.0.9 'himalaya -c ~/tmp-test-config.toml account list'
# ┌──────┬────────────┬─────────┐  NAME=tars  BACKENDS=imap, smtp  DEFAULT=yes   rc=0
```

`email` is accepted, so it was kept (spec intent preserved). The real file was then written in a
remote shell that substitutes the address **from the VM's own 0600 file** — the address never
transits cooper:

```sh
ssh gaetan@192.168.0.9 'umask 077; mkdir -p ~/.config/himalaya;
  ADDR=$(cat ~/.hermes/.secrets/gmail_address);
  PWCMD="cat /home/gaetan/.hermes/.secrets/gmail_app_password";
  cat > ~/.config/himalaya/config.toml <<EOF
  … "$ADDR" … "$PWCMD" …
EOF
  chmod 600 ~/.config/himalaya/config.toml'
```

Resulting file (`-rw------- 744 bytes`), local-part redacted:

```toml
mailbox.alias.inbox = "INBOX"

[accounts.tars]
default = true
email = "<redacted>@mobile.club"

imap.server = "imaps://imap.gmail.com:993"
imap.sasl.plain.username = "<redacted>@mobile.club"
imap.sasl.plain.password.command = "cat /home/gaetan/.hermes/.secrets/gmail_app_password"

smtp.server = "smtps://smtp.gmail.com:465"
smtp.sasl.plain.username = "<redacted>@mobile.club"
smtp.sasl.plain.password.command = "cat /home/gaetan/.hermes/.secrets/gmail_app_password"
```

No password inline anywhere — a rotation is a re-delivery of `~/.hermes/.secrets/gmail_app_password`
and nothing else. Config validation:

```sh
ssh gaetan@192.168.0.9 'himalaya account check -a tars'
```
```
Account: tars
  imap: OK
  smtp: OK
```
rc=0 — both legs authenticate against Gmail with the file-sourced password.

**PASS.**

## Step 4 — nothing registered

No MCP server, no plugin, no `config.yaml` stanza, no `.env` write. `~/.hermes/config.yaml` was not
read, copied, or modified by this agent. No `flock` was needed: this agent edits no shared file.

**PASS.**

## Probe (ends the agent)

Flags verified first against `himalaya envelope list --help` on the VM: v2 has **no `-f`**; the
folder flag is `-m/--mailbox`, and `-s/--page-size` is the page size the spec intended.

```sh
ssh gaetan@192.168.0.9 'himalaya envelope list -a tars -m INBOX -s 1'
```
```
plain_probe_rc=0
table_lines=6
┌───────┬───────┬────────────────┬────────┬────────────────────────┬───────────┐
│ ID    ┆ FLAGS ┆ SUBJECT        ┆ FROM   ┆ DATE                   ┆ SIZE      │
╞═══════╪═══════╪════════════════╪════════╪════════════════════════╪═══════════╡
│ 12204 ┆ *     ┆ Security alert ┆ Google ┆ 2026-08-07 17:17+00:00 ┆ 12.03 KiB │
└───────┴───────┴────────────────┴────────┴────────────────────────┴───────────┘
```

Same run in `--json`, machine-checked and redacted (subject truncated to 20 chars):

```
top_keys=['envelopes']  rows=1
row keys = date, flags, from, has-attachment, id, message-id, size, subject, to
sender_domain=? | subject=Security alert | date=2026-08-07T17:17:27Z | flags=[] | size=12323
```

(`sender_domain=?` because the `from` object carries the display name `Google` and no bare addr in
that row — not a probe failure; the plain-table run above shows the same row.)

**Pass criteria met: exactly one envelope row on stdout, exit 0.** `INBOX` resolved directly, so the
`\All`-tag fallback in r6 was not needed. No message fetched, nothing sent.

**PASS.** The fallback `curl imaps://` discriminator was therefore never run — it is only a
credential-vs-config discriminator and there is nothing to discriminate.

Note, not a failure: the one envelope in INBOX is a Google **"Security alert"** timestamped 17:17Z,
i.e. Google noticing this very IMAP sign-in. Expected for a first app-password login from a new host.

## Cleanup

`~/tmp-hima`, `~/tmp-hman`, `~/tmp-hima-sample.toml`, `~/tmp-test-config.toml`, `~/tmp-env.json`
removed from the VM. Nothing written on cooper except this file.

---

## Deviations

1. **snap lost.** `sudo -n snap install himalaya` → `error: snap "himalaya" not found`. Installed the
   official v2.0.0 `x86_64-linux` release tarball into `~/.local/bin` instead, per spec's own
   alternative.
2. **Extra `/usr/local/bin/himalaya` symlink** (`sudo -n ln -sf`). Not in the spec. Without it, bare
   `himalaya` is not on PATH for the non-interactive/non-login shell Tars' shell tool will use.
3. **Spec §1's config block does not parse under himalaya v2.0.0** — `backend.*`,
   `message.send.backend.*` and `auth.cmd` are all v1 keys. Re-derived from
   `config.sample.toml @ v2.0.0` (table in Step 3). The spec block should be updated.
4. **Spec's probe flag `-f INBOX` does not exist in v2** — it is `-m/--mailbox`. `-s 1` was correct.
5. **`himalaya account configure --help` does not exist in v2** (only `account list` / `account
   check`). Used `account check -a tars` as the config validator; it exercises both IMAP and SMTP.
6. **`GMAIL_ADDRESS` also delivered as a 0600 file** (`~/.hermes/.secrets/gmail_address`); the spec
   names only the password file. Required so the address is substituted into `config.toml` on the VM
   and never transits cooper.
7. **Trailing newlines stripped** from both delivered files (`tr -d "\r\n"`), so
   `cat <file>` hands the SASL layer exactly the credential.
8. Added `mailbox.alias.inbox = "INBOX"` so a bare `himalaya envelope list -a tars` (no `-m`) also
   lands on INBOX — v2 falls back to the `inbox` alias, which is otherwise unset.

## Followups

- Update `docs/specs/wf3-wiring.md` §1 to the v2.0.0 schema and the `-m` probe flag, or a future
  re-run will repeat this derivation.
- Rotation decision on the Gmail pair is still pending with Gaetan. When it lands, the only action
  on this wiring is re-delivering `~/.hermes/.secrets/gmail_app_password` with the same
  `--extract | ssh` one-liner. No config edit, no reinstall. `~/.hermes/.env` (agent 2's copy of the
  same pair) has to be re-delivered separately — that is not this agent's file.
- SMTP is configured and authenticates (`smtp: OK`) but was deliberately not exercised by a send.
