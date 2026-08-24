---
name: claude-account-usage
description: Fetch usage across Gaetan's Claude accounts.
version: 1.1.0
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
- A Claude magic link is a bearer credential: consume it inside CloakBrowser without printing it, returning it from a tool, copying it into chat, saving it, or passing it to a subagent.
- Never expose magic-link fragments, verification codes, cookies, session storage, authorization headers, message source, or Gmail contents beyond the minimum usage result.
- Use only a magic-link email generated during the current account step. Ignore older matching mail.
- Verify the authenticated Claude email before recording usage. If it does not exactly match the requested account, log out and mark that account failed; do not infer or relabel the result.
- Do not change plan, billing, organization, settings, or account membership.
- Use the Personal Organization for `gaetan.cathelain@mobile.club`; do not report MobileClub organization usage for that account.
- Keep one result per account. A failure on one account does not erase successful results, but do not retry an account more than twice.

## 1. Verify prerequisites

The expected persistent browser is exposed over CDP at `http://127.0.0.1:9223` and uses `~/.hermes/cloakbrowser-profile`.

1. Call `browser_cdp` with `Target.getTargets` and confirm it returns at least one target.
2. If CDP is unavailable, check `http://127.0.0.1:9223/json/version` without printing cookies or page data.
3. If no process is serving it, start this existing helper as a long-running background process:

   `~/.hermes/tools/cloakbrowser/.venv/bin/python ~/.hermes/tools/cloakbrowser/start_cloakbrowser_cdp.py`

4. Verify `/json/version` responds and retry `Target.getTargets`. Stop if the endpoint still cannot be reached; do not fall back to another browser backend.
5. Verify the existing Google API token without asking Gaetan to log into Gmail:

   ```bash
   PY="$HOME/.hermes/hermes-agent/venv/bin/python"
   BASE="$HOME/.hermes/skills/productivity/google-workspace/scripts"
   "$PY" "$BASE/setup.py" --check
   ```

6. The check must report `AUTHENTICATED`. If it fails, report the Google API authentication failure exactly; never replace this with an interactive Gmail-login request.

Do not restart a healthy CloakBrowser.

## 2. Prepare the Claude tab

Keep one top-level Claude tab in CloakBrowser at `https://claude.ai/login`. Use `Target.getTargets` to retain its page target ID. Use tab-specific CDP calls when ordinary browser tools would make the active tab ambiguous.

Do not open or depend on a Gmail browser tab. Gmail retrieval is through the authenticated Google API token so the skill never asks Gaetan to log into Gmail interactively.

## 3. Process one Claude account

Repeat this whole section for exactly one target email before moving to the next.

### A. Establish a clean Claude session

1. In the Claude tab, log out of the current Claude account using the account menu. If the UI is inaccessible, navigate to Claude's official logout route only after confirming the `claude.ai` origin.
2. Verify the tab shows the logged-out sign-in flow.
3. Navigate to `https://claude.ai/login` if needed.
4. Enter the exact target email and request an email/magic-link login.
5. Record the request time and the newest matching Gmail message ID before the request for freshness matching only; do not expose either unless diagnosing a timeout.
6. Verify Claude confirms that a login email was sent.

### B. Retrieve and consume the fresh magic link without exposing it

1. Poll Gmail through the authenticated Google API, never through an interactive browser login:

   ```bash
   PY="$HOME/.hermes/hermes-agent/venv/bin/python"
   API="$HOME/.hermes/skills/productivity/google-workspace/scripts/google_api.py"
   "$PY" "$API" gmail search 'to:<TARGET> from:(mail.anthropic.com) newer_than:1d' --max 5
   ```

2. Poll for at most ten minutes, roughly every 10–20 seconds. Select only a message newer than the pre-request message ID and received after the request. Confirm its `to` value exactly matches the target account and that the fetched body contains a Claude magic link. Do not depend on the subject language; Anthropic localizes it.
3. Fetch that exact message with `gmail get` inside a local process whose stdout is captured, not printed. Extract the `https://claude.ai/magic-link#...` href in process memory. Never return or log the URL, fragment, body, or code.
4. In that same local process, connect directly to the existing CloakBrowser CDP WebSocket and send `Page.navigate` to the retained Claude target with the full fresh magic-link URL. Print only a non-secret status such as `FRESH_MAGIC_LINK_NAVIGATED`.
5. Never call `Runtime.evaluate` on `location.href`, `Network.getCookies`, or any other command whose result would expose the bearer link or cookie values.
6. If Anthropic presents hCaptcha, solve the visible challenge through browser screenshot/vision and normal browser or CDP input events. Bound this to two challenge attempts. Do not ask Gaetan to solve it unless both attempts fail.
7. Wait for the authenticated Claude application. Do not paste the magic-link fragment into the verification-code box; it is not the displayed verification code.

If no correct fresh message appears within ten minutes, mark that account `magic-link timeout` and continue only after restoring the Claude tab to logged-out state.

### C. Verify identity and fetch usage

1. Wait until the authenticated Claude application loads.
2. Open the account/profile menu and verify the signed-in email exactly equals the target email.
3. For `gaetan.cathelain@mobile.club`, select and verify **Personal Organization** before reading usage. Do not use the MobileClub organization.
4. Navigate to `https://claude.ai/new#settings/usage`.
5. Wait for the Usage settings panel to finish loading. If the hash route does not open directly, use Settings → Usage in the UI.
6. Read every visible usage category, including its label, consumed percentage or remaining allowance, and reset/renewal time when shown. Do not calculate unavailable values.
7. Take a snapshot after the panel is loaded. Treat the snapshot as the evidence for that account.
8. Record the verified account email, organization when shown, each usage category exactly as displayed, reset times exactly as displayed, and the check time.
9. If the page shows an error or no usage data after one refresh, record the exact visible state instead of guessing.

### D. Close the account step

1. Log out of Claude.
2. Verify the logged-out sign-in flow before processing the next account.
3. Close duplicate Claude tabs created by the magic link, keeping one clean Claude tab.

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
- [ ] Authenticated Google API access was used; Gaetan was not asked to log into Gmail.
- [ ] Accounts were processed serially.
- [ ] Every successful result was tied to an exact email identity check.
- [ ] `gaetan.cathelain@mobile.club` used Personal Organization.
- [ ] Every magic link was fresh and no bearer link, fragment, code, or cookie was exposed.
- [ ] Every successful row came from `https://claude.ai/new#settings/usage` or the equivalent in-app Usage panel.
- [ ] Claude is logged out at the end.
