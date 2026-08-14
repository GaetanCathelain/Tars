# Slack MCP write-policy allowlists

Use this when Hermes exposes Slack through an external MCP server and a write tool returns an error such as `conversations_add_message tool is not allowed for channel …, applied policy: …`.

## Separate the three gates

1. **Hermes sender authorization** — `SLACK_ALLOWED_USERS` (normally in the active profile `.env`) controls which Slack users may trigger Hermes. It is not the MCP posting policy.
2. **Hermes inbound conversation routing** — `slack.allowed_channels` in `config.yaml` controls which shared Slack conversations Hermes accepts incoming events from. An empty value means unrestricted; setting a one-item list narrows all other shared conversations. Do not change this merely to fix an outbound MCP posting denial unless the user explicitly requested inbound restriction.
3. **External MCP outbound write policy** — for `korotovsky/slack-mcp-server`, `SLACK_MCP_ADD_MESSAGE_TOOL` controls where `conversations_add_message` may post. Supported values include `true`/`1` for all conversations or a comma-separated ID allowlist. Prefer appending the exact requested conversation ID to the existing list.

## Inspection without secret exposure

- Identify the MCP command from gateway service status/process arguments. An `--env-file` may point outside `HERMES_HOME`.
- Parse and print only named allowlist/policy variables. Never dump the full env file or container environment.
- Inspect the installed/pinned MCP binary or its authoritative docs for the exact policy variable and semantics. The error's `applied policy` value is strong evidence of the live process snapshot.

## Safe writes

- Use `hermes config set` only for values that genuinely live in `config.yaml`.
- Update env-file policy lines with a targeted atomic replacement that preserves every unrelated line, file mode, and credential.
- For scalar-or-CSV Hermes settings such as `slack.allowed_channels`, prefer a plain scalar or comma-separated value. On Hermes v0.20.0, passing JSON list text to `hermes config set` can persist the JSON syntax as a string rather than a YAML list. Always re-read the value and verify its YAML type/effective bridge.

## Reload and verification

- YAML-to-env bridges and MCP `--env-file` values are startup snapshots. Editing the files does not update an already-running gateway or MCP child.
- Prove both states separately: (a) a fresh loader/new throwaway container sees the new policy, and (b) the live process still has the old policy until restarted.
- If the current task runs inside the gateway service tree, do not self-restart. Report the exact external restart command and leave activation pending.
- Perform the requested live write test only after proving the live MCP process has the new policy. A direct `hermes send` is not an equivalent test because it bypasses the MCP policy gate.
