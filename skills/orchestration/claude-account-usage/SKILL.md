---
name: claude-account-usage
description: Fetch usage across Gaetan's Claude accounts.
version: 1.0.0
author: Tars
license: private
platforms: [linux]
metadata:
  hermes:
    tags: [claude, usage, cloakbrowser, gmail, account-rotation]
---

# Claude Account Usage

Use this skill when Gaetan asks for Claude usage, limits, or remaining capacity across his five Claude Max accounts. Drive the existing persistent CloakBrowser session through Hermes browser tools. Process accounts strictly one at a time so Claude cookies and magic links cannot cross.

## Accounts

Run in this order unless Gaetan names a subset or another order:

1. `gaetan.cathelain@mobile.club` — Personal Organization, not the MobileClub organization.
2. `gaetan.cathelain-claude02@mobile.club`
3. `eirik.valdstrom@mobile.club`
4. `sigrun.havelund@mobile.club`
5. `torvald.bjornkvist@mobile.club`

All login messages arrive in the Gmail inbox for `gaetan.cathelain@mobile.club`.

## Safety invariants

- Never run accounts in parallel.
- A Claude magic link is a bearer credential: click it inside CloakBrowser. Never extract, print, copy into chat, save, or pass its URL to a subagent.
- Never expose cookies, session storage, authorization headers, message source, or Gmail contents beyond the minimum usage result.
- Use only a magic-link email generated during the current account step. Ignore older matching mail.
- Verify the authenticated Claude email before recording usage. If it does not exactly match the requested account, log out and mark that account failed; do not infer or relabel the result.
- Do not change plan, billing, organization, settings, or account membership.
- Use the Personal Organization for `gaetan.cathelain@mobile.club`; do not report MobileClub organization usage for that account.
- Keep one result per account. A failure on one account does not erase successful results, but do not retry an account more than twice.

## 1. Verify CloakBrowser

The expected persistent browser is exposed over CDP at `http://127.0.0.1:9223` and uses `~/.hermes/cloakbrowser-profile`.

1. Call `browser_cdp` with `Target.getTargets`.
2. Confirm it returns at least one target. This proves Hermes browser tools are attached to the live CloakBrowser CDP endpoint.
3. If CDP is unavailable, check `http://127.0.0.1:9223/json/version` without printing cookies or page data.
4. If no process is serving it, start this existing helper as a long-running background process:

   `~/.hermes/tools/cloakbrowser/.venv/bin/python ~/.hermes/tools/cloakbrowser/start_cloakbrowser_cdp.py`

5. Verify `/json/version` responds and retry `Target.getTargets`. Stop if the endpoint still cannot be reached; do not fall back to another browser backend.

Do not restart a healthy CloakBrowser. Its persistent profile is what retains the central Gmail login.

## 2. Prepare two tabs

Keep two top-level tabs in the same CloakBrowser context:

- **Gmail tab:** `https://mail.google.com/`
- **Claude tab:** `https://claude.ai/`

Use `Target.getTargets` to retain both page target IDs. Use `browser_cdp` with the target ID for tab-specific inspection when ordinary browser tools would make the active tab ambiguous.

In Gmail, verify the inbox is already authenticated as `gaetan.cathelain@mobile.club`. If Google asks for credentials or MFA, stop and tell Gaetan the Gmail session needs interactive login. Do not ask for or handle his Google password.

## 3. Process one Claude account

Repeat this whole section for exactly one target email before moving to the next.

### A. Establish a clean Claude session

1. In the Claude tab, log out of the current Claude account using the account menu. If the UI is inaccessible, navigate to Claude's logout route only after confirming it is the official `claude.ai` origin.
2. Verify the tab shows the logged-out sign-in flow.
3. Navigate to `https://claude.ai/login` if needed.
4. Enter the exact target email and request an email/magic-link login.
5. Record the request time for matching only; do not expose it unless diagnosing a timeout.
6. Verify Claude confirms that a login email was sent.

### B. Wait for the correct fresh magic link

1. Switch to the Gmail tab.
2. Search for a fresh Claude login email that arrived after the request and contains the exact target account address. Prefer Gmail search terms combining the target email, Claude sender/domain, and a narrow time window.
3. Poll by refreshing the search, with a bounded wait of ten minutes. Do not busy-loop; wait roughly 10–20 seconds between checks.
4. Open only the newest matching message received after this account's login request.
5. Confirm the message is an Anthropic/Claude login email and refers to the exact target account.
6. Click the login button/link in the rendered email. Never inspect or return the link URL.
7. If Gmail warns about leaving the site, continue only when the destination shown is an official Claude/Anthropic domain.

If no correct fresh message appears within ten minutes, mark that account `magic-link timeout` and continue only after restoring the Claude tab to logged-out state.

### C. Verify identity and fetch usage

1. Return to the Claude tab or the newly opened Claude tab produced by the magic link.
2. Wait until the authenticated Claude application loads.
3. Open the account/profile menu and verify the signed-in email exactly equals the target email.
4. For `gaetan.cathelain@mobile.club`, select and verify **Personal Organization** before reading usage. Do not use the MobileClub organization.
5. Navigate to `https://claude.ai/new#settings/usage`.
6. Wait for the Usage settings panel to finish loading. If the hash route does not open directly, use Settings → Usage in the UI.
7. Read every visible usage category, including its label, consumed percentage or remaining allowance, and reset/renewal time when shown. Do not calculate unavailable values.
8. Take a snapshot after the panel is loaded. Treat the snapshot as the evidence for that account.
9. Record the verified account email, organization when shown, each usage category exactly as displayed, reset times exactly as displayed, and the check time.
10. If the page shows an error or no usage data after one refresh, record the exact visible state instead of guessing.

### D. Close the account step

1. Log out of Claude.
2. Verify the logged-out sign-in flow before processing the next account.
3. Close duplicate Claude tabs created by the magic link, keeping one clean Claude tab and the Gmail tab.

## 4. Report

Return a compact table in the requested language, one row or section per account:

- Account
- Organization, when relevant
- Usage categories and displayed values
- Reset/renewal times
- Checked time
- Status/error if unavailable

Lead with how many accounts were successfully verified, for example: `Verified usage for 4/5 accounts.` Distinguish `not checked`, `login failed`, `magic-link timeout`, `identity mismatch`, and `usage unavailable`. Never claim complete success unless every row was identity-verified and read from the Usage page during this run.

## Verification checklist

Before reporting completion, confirm:

- [ ] CloakBrowser CDP was used; no fallback browser was used.
- [ ] Accounts were processed serially.
- [ ] Every successful result was tied to an exact email identity check.
- [ ] `gaetan.cathelain@mobile.club` used Personal Organization.
- [ ] Every magic link was fresh and was never exposed.
- [ ] Every successful row came from `https://claude.ai/new#settings/usage` or the equivalent in-app Usage panel.
- [ ] Claude is logged out at the end and Gmail remains open.
