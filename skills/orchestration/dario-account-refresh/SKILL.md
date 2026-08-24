---
name: dario-account-refresh
description: Refresh one or more Dario OAuth pool accounts safely.
version: 0.1.0
author: Gaetan Cathelain, Hermes Agent
license: MIT
platforms: [linux]
metadata:
  hermes:
    tags: [Dario, OAuth, Claude, Operations]
    related_skills: [containerized-service-operations, google-workspace, computer-use]
---

# Dario Account Refresh

Refresh one or more named Claude OAuth accounts in the Dario pool on `cooper`, using isolated Chrome profiles on `mac` and magic links from Gaetan's authenticated Gmail API. This workflow changes credentials and may optionally remove explicitly named aliases; keep every secret out of output and verify each write.

## When to Use

- Refresh one or more expired Dario pool aliases.
- Replace explicitly named temporary or backup aliases.
- Reload and verify the Dario proxy after account changes.

Do not use this for single-login `dario refresh`; pool accounts have a separate credential store.

## Prerequisites

- SSH aliases `cooper` and `mac` are reachable.
- Docker container `dario` is running on `cooper`; the working tree is `/root/dario-fork`.
- Google Workspace auth is valid. Use the existing authenticated Gmail API; never require interactive Gmail login.
- Chrome is installed on the Mac and can be launched with remote debugging.
- Load `containerized-service-operations`, its `references/interactive-oauth-refresh.md` and `references/dario-pool-reauth.md`, plus `google-workspace` and `computer-use`.

## Procedure

### 1. Establish the exact account plan

1. Run `bun dist/cli.js accounts list` inside the container and record only aliases and expiry states.
2. Map every requested alias to its exact email before starting.
3. If deletion was requested, name every exact alias and leave the destructive-action beat. After Gaetan confirms, run `accounts remove <alias>` once per named alias and re-read the full list. Never infer an alias from a wildcard.
4. Preserve existing credentials unless deletion was explicitly requested. A failed OAuth exchange must not erase an otherwise usable account.

Completion: every target email has one alias, and any requested deletions are confirmed absent.

### 2. Prepare isolated browser control

1. Create one SSH tunnel from a free local port to Chrome's remote-debugging port on `mac`.
2. For each account, launch a new Chrome instance with a unique `--user-data-dir`. Never reuse another account's profile or a profile from a failed OAuth state/verifier pair.
3. Close and remove only that exact profile after the alias is verified, then start the next account.

Completion: CDP identifies exactly the dedicated page for the current account.

### 3. Start a secret-safe Dario OAuth exchange

Use Dario's `startAddAccount(alias)` and `completeAddAccount(alias, code, codeVerifier, state)` from `/root/dario-fork/dist/accounts.js` in a temporary script inside the container.

- Run the script with stdin attached; do not pipe JavaScript source into `bun -`, because the source pipe consumes the later authorization input.
- Read the code with a non-echo line reader (`terminal: false`).
- Keep the process alive while the browser flow runs.
- Capture `pending.authorizeUrl` internally. Never print or report it.
- Before `completeAddAccount`, pass only the code portion: strip the `#state` suffix. Supply `pending.codeVerifier` and `pending.state` separately.

Completion: the same live process owns the authorization URL, verifier, state, and final exchange.

### 4. Sign in through the fresh Chrome profile

1. Navigate internally to the new authorization URL.
2. Before requesting the magic link, baseline Gmail message IDs with:
   `to:<exact-email> from:(mail.anthropic.com) newer_than:1d`
3. Submit the exact account email in Claude.
4. Poll Gmail for a message newer than the baseline and addressed to the exact target. Do not depend on an English subject; Anthropic localizes subjects.
5. Extract the magic link in memory and navigate to it in the same Chrome profile. Never output the link or email body.
6. If Authorize remains disabled, ask Gaetan to click the black area. Once enabled, click Authorize and inspect the resulting callback page.

Completion: the callback's state matches the live Dario process's state.

### 5. Exchange and verify each alias

1. Pass the callback code directly to the waiting Dario process without printing or echoing it.
2. Require a successful process exit.
3. Re-run `accounts list` and require the exact alias to show a future expiry; browser success alone is insufficient.
4. Close and delete the exact Chrome profile before continuing to the next account.

Completion: every requested alias is present and non-expired.

### 6. Restart the active proxy

1. Discover the unique running `bun dist/cli.js proxy --port=<port>` process and its working directory. Do not assume the user's TUI port is the proxy port.
2. Re-read the PID, command, port, and cwd immediately before sending `TERM`.
3. Wait for the old PID to exit, then start the exact same proxy command from the same cwd.
4. Require a different PID, HTTP 200 from `/health`, `status: ok`, and `oauth: healthy`.
5. Read `DARIO_API_KEY` only from the new listener's environment in memory and make one minimal `/v1/messages` request. Report only HTTP status, response type, and stop reason.

Completion: the new listener is healthy and one authenticated upstream message succeeds.

### 7. Clean up

Close the SSH tunnel, remove only the temporary browser profiles and temporary OAuth scripts created by this run, and verify they are gone. Do not kill unrelated Chrome, SSH, Docker, or Dario processes.

## Pitfalls

- `dario accounts add <alias> --manual` refuses existing aliases; the lower-level start/complete functions can safely replace the exact alias only after a successful exchange.
- `completeAddAccount` takes four arguments, not a pending object.
- Passing `code#state` as the code produces `invalid_grant`; strip the suffix.
- A restarted OAuth process invalidates every earlier URL, magic link, code, verifier, and state.
- A process named `dario tui --port=3475` does not prove the proxy listens on 3475; discover the proxy command and probe it.
- `/health` is necessary but not sufficient. Require a real message request.
- Authorization URLs, magic links, codes, tokens, account files, and API keys are secrets. Keep them out of tool output, logs, screenshots, files, and chat.

## Verification

Report only:

- aliases removed and their absence from the re-read list;
- each refreshed alias and its future expiry;
- old and new proxy PIDs plus the discovered port;
- health HTTP status, non-secret OAuth state, and version;
- real message HTTP status, response type, and stop reason;
- cleanup verification.
