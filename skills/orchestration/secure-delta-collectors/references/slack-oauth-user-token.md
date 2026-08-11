# Slack OAuth user tokens for read-only collectors

Use this instead of browser-session `xoxc`/`xoxd` authentication when a Slack collector must call `search.messages` and `conversations.replies` on a member's behalf.

## Recommended credential

Create an **internal, undistributed, single-workspace Slack app** and install it as the target member. Use its OAuth V2 **user access token**:

- non-rotating: `xoxp-…` (Slack says it does not expire);
- rotating: `xoxe.xoxp-…` plus a one-time refresh token and app client credentials (12-hour access-token lifetime).

Request only these **User Token Scopes**:

- `search:read`
- `channels:history`
- `groups:history`
- `im:history`
- `mpim:history`

Do not request bot scopes merely for this collector. `search.messages` and `search:read` are user-token-only. `conversations.replies` supports user or bot tokens, but a bot token cannot satisfy the combined workflow.

## Visibility

A user token represents the member who authenticated the app and has the same workspace visibility that member has. Install as the person whose search visibility the collector requires. Local filtering does not narrow the OAuth grant.

## Installation and lifecycle

1. Create the app for the target workspace and leave distribution disabled.
2. Add the five user scopes under **OAuth & Permissions**.
3. Install/reinstall while signed in as the target member.
4. If app approval is enabled, request owner/app-manager approval for those exact scopes.
5. Put the User OAuth Token directly in secure runtime secret storage; never place tokens or client secrets in chat, logs, source control, durable workflow state, or this reference.
6. Enable token rotation only after refresh handling is implemented. Rotation cannot be turned off once enabled.

Uninstalling revokes all installation tokens; `auth.revoke` can revoke a single token. Apps with these scopes may be automatically uninstalled if the installing user leaves or becomes a guest, so reauthorization is a normal recovery path.

## Distribution and rate limits

- Internal customer-built app: `search.messages` is Tier 2 (20+ requests/minute); `conversations.replies` is Tier 3 (50+ requests/minute).
- New commercially distributed non-Marketplace app/install: `conversations.replies` is 1 request/minute with limit/default 15.
- Marketplace is not a practical route for this exact contract: Slack's current submission guide explicitly identifies `search:read` among legacy/restricted scopes or methods it does not accept.
- `search.messages` is officially marked legacy; plan a separate future migration to Slack's Real-time Search API, but do not silently change an existing collector contract.

## Official Slack sources

- `search.messages`: https://docs.slack.dev/reference/methods/search.messages/
- `search:read`: https://docs.slack.dev/reference/scopes/search.read/
- `conversations.replies`: https://docs.slack.dev/reference/methods/conversations.replies/
- Token types and user visibility: https://docs.slack.dev/authentication/tokens/
- OAuth V2 installation: https://docs.slack.dev/authentication/installing-with-oauth/
- Token rotation and revocation: https://docs.slack.dev/authentication/using-token-rotation/
- App distribution: https://docs.slack.dev/distribution/
- Rate limits: https://docs.slack.dev/apis/web-api/rate-limits/
- App approval: https://slack.com/help/articles/222386767-Manage-app-approval-for-your-workspace
- Marketplace restrictions: https://docs.slack.dev/slack-marketplace/distributing-your-app-in-the-slack-marketplace/
- Credential storage: https://docs.slack.dev/authentication/best-practices-for-security/
