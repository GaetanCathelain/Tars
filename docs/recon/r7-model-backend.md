# R7 — model backend: ChatGPT-subscription OAuth in Hermes

Recon for DECISION.md Blocker 1. Read-only throughout. No secret values printed anywhere below —
token/credential fields are redacted to shape (`prefix...(N chars)`) or omitted. Course-corrected
mid-probe: the personal Hermes deployment is **already** running production traffic on the
`gaetan.cathelain@mobile.club` ChatGPT subscription today, so this doc documents the live,
working mechanism rather than a hypothetical.

## Verdict

**Yes — proven live, not hypothetical.** Hermes Agent has a first-class provider `openai-codex`
that authenticates via ChatGPT OAuth device-code flow and bills against the ChatGPT
subscription (not the pay-per-token API). It is live on p-Hermes **today**, backing the
`default` profile (Telegram) and at least the `tars`, `justine`, `immichbackup`,
`immichresearch`, and `papainventory` profiles, all resolving to the **same one** underlying
OAuth credential. The credential lands in `auth.json` under `credential_pool.openai-codex[]`
(current schema) — confirmed directly, not inferred from docs. Blocker 1 is resolved: lane A's
job is to run the same one command on the new VM, not to invent a new integration.

The `default:` model id in config.yaml is literally `gpt-5.6-sol` on four of the five profiles
checked — PLAN.md's target model is already the production default elsewhere in this Hermes
install. Only `tars` itself is pinned one rung down at `gpt-5.5` (a pre-existing quirk of the
profile being retired at cutover, not a new-VM concern — see Open questions).

## Supported provider paths

From the live `v0.19.0` binary's own `--help` (authoritative over docs, which describe the
newer `v0.20.0` pinned for the Tars VM — flag shapes should be stable across a one-minor-version
gap but re-verify per R1's standing caveat):

```
$ hermes auth --help
usage: hermes auth [-h] {add,list,remove,reset,status,logout,spotify} ...
  add       Add a pooled credential
  list      List pooled credentials
  remove    Remove a pooled credential by index, id, or label
  reset     Clear exhaustion status for all credentials for a provider
  status    Show auth status for a provider
  logout    Log out a provider and clear stored auth state

$ hermes auth add --help
usage: hermes auth add [-h] [--type {oauth,api-key,api_key}] [--label LABEL]
                        [--api-key API_KEY] [--portal-url PORTAL_URL]
                        [--inference-url INFERENCE_URL] [--client-id CLIENT_ID]
                        [--scope SCOPE] [--no-browser] [--timeout TIMEOUT]
                        [--insecure] [--ca-bundle CA_BUNDLE]
                        provider
  provider  Provider id (for example: anthropic, openai-codex, openrouter)
```

So `openai-codex` is one named provider among several (`anthropic`, `openrouter`, others per
vendor docs: Google Gemini, xAI Grok, DeepSeek, MiniMax, Qwen, Bedrock, Azure Foundry, GitHub
Copilot, Nous Portal). Three ways to reach an OpenAI model, confirmed a mix of live-probed and
vendor-doc:

1. **`openai-codex` — ChatGPT subscription OAuth (the one in use).** `hermes auth add
   openai-codex --type oauth` (or the interactive `hermes model` picker, though that command's
   own `--help` exposes only Nous-Portal-flavored flags — `--portal-url`, `--inference-url`,
   `--client-id` — so `hermes auth add` is the more direct, provider-explicit path). Opens a
   browser device-code flow against `chatgpt.com`; no API key. Confirmed live (`hermes auth
   status openai-codex` → `openai-codex: logged in`).
2. **Plain `OPENAI_API_KEY`** — standard pay-per-token API key in `.env`, `hermes auth add
   openai-codex --type api-key --api-key ...` or an `openai`-style provider entry (vendor docs
   only; not observed live on p-Hermes — no profile checked uses this path).
3. **GitHub Copilot OAuth** — `hermes auth add copilot`-style, reaches "GPT-5.x, Claude, Gemini"
   through a Copilot subscription instead of ChatGPT's. Vendor-doc only, not live anywhere
   probed, not relevant given Gaetan's ChatGPT (not Copilot) subscription is the target.

Nous Portal (`hermes setup --portal`, R1) is a *different* provider (`nous`) — free OAuth to
Nous's own hosted models, unrelated to OpenAI/ChatGPT despite superficially similar "one OAuth,
no API key" framing. Don't conflate the two when scripting lane A.

## ChatGPT-sub specifics — how it lands in auth.json / credential_pool

Live-probed, all six profiles + the base install, via `qm guest exec 103` as user `hermes`
(read-only `stat`/`jq`/`grep`, values piped through a redact-to-shape filter before ever leaving
the VM):

**`auth.json` top level** (`~/.hermes/auth.json`, the base/`default` profile — 5,967 bytes,
mtime 2026-08-05 01:41:37 UTC):
```json
{
  "version": 1,
  "providers": {
    "openai-codex": {
      "tokens": { "access_token": "eyJhbG...(1656 chars)", "refresh_token": "rt.1.A...(211 chars)" },
      "last_refresh": "2026-0...(27 chars)",
      "auth_mode": "chatgpt"
    }
  },
  "active_provider": "openai...(12 chars)",   // == "openai-codex"
  "updated_at": "2026-0...(32 chars)",
  "credential_pool": {
    "openai-codex": [{
      "id": "39c133", "label": "device...(11 chars)", "auth_type": "oauth", "priority": 0,
      "source": "device...(11 chars)",   // == "device_code" per `hermes auth list` (below)
      "access_token": "eyJhbG...(1656 chars)", "refresh_token": "rt.1.A...(211 chars)",
      "last_status": "ok", "last_error_code": null, "last_error_reason": null,
      "base_url": "https:...(37 chars)",   // == "https://chatgpt.com/backend-api/codex", byte-length-confirmed
      "last_refresh": "2026-0...(27 chars)", "request_count": 0
    }],
    "anthropic": [ /* 2 entries, unrelated to this recon — Claude Code + a separate Anthropic OAuth,
                      one shown "exhausted (ready to retry)" */ ]
  }
}
```
Two storage shapes coexist: the legacy singular `providers.openai-codex.tokens.{access_token,
refresh_token}` + `auth_mode: "chatgpt"` (only present on the `default` profile — likely an
older-schema leftover from before pooling existed), and the current
`credential_pool.openai-codex[]` array shape (array because a provider can hold multiple pooled
credentials, rotated per `credential_pool_strategies`: `fill_first`/`round_robin`/`least_used`/
`random`, per vendor docs). `active_provider` and `providers.*` are what a given profile's model
calls actually resolve through; `credential_pool` is the auth store `hermes auth {add,list}`
manages. `auth_mode: "chatgpt"` is the literal tag confirming this specific credential came from
the ChatGPT-subscription OAuth mode, not a plain API key or a ChatGPT *Enterprise*/API-org login.

**`config.yaml` → `model:` block, per profile** (non-secret, confirms provider selection is
config-level, credential is auth.json-level):

| Profile | `model.provider` | `model.default` | `model.base_url` | own `auth.json`? |
|---|---|---|---|---|
| `default` (`~/.hermes`) | `openai-codex` | `gpt-5.6-sol` | `https://chatgpt.com/backend-api/codex` | yes (5,967 B) |
| `tars` | `openai-codex` | **`gpt-5.5`** | same | yes (2,573 B, mtime 2026-08-04 — older than default's) |
| `justine` | `openai-codex` | `gpt-5.6-sol` | (not set — inherits) | yes (1,411 B) but **no `openai-codex` entry inside it**, only `anthropic` |
| `immichbackup` / `immichresearch` / `papainventory` | `openai-codex` | `gpt-5.6-sol` | same | **no auth.json file at all** |
| `privatelocal` | `custom` | `hermes-3-llama-3.2-3b` | local :18082 | n/a — unrelated, local model |

**Credential fallback, confirmed by direct observation, not by reading source:** `hermes auth
list` run under four different `HERMES_HOME`s (`default`, `tars`, `justine`,
`immichbackup`) returns **byte-identical** results for `openai-codex`:
```
openai-codex (1 credentials):
  #1  device_code          oauth   device_code
```
even for `justine` (whose own `auth.json` has no `openai-codex` key at all) and
`immichbackup`/`immichresearch`/`papainventory` (no `auth.json` file present, period). The only
place a real, distinct `openai-codex` credential blob exists on disk is `default`'s and `tars`'s
`auth.json` — and those two are identical in every redacted-shape field (`id: "39c133"`, same
token lengths, same `base_url` length, same `last_refresh` length). **Conclusion: Hermes falls
back to the base profile's (`~/.hermes/auth.json`) `credential_pool` entry for a provider when a
named profile's own `auth.json` lacks (or omits) that provider** — a per-profile `auth.json`
only needs to exist if it's *overriding* the base credential, which is presumably what happened
for `tars` at some point (its file predates `default`'s most recent refresh, so it's a stale
clone/manual-copy, not a live symlink — refreshing on one side does not appear to write through
to the other). This is inferred from repeated, consistent CLI behavior across 4 profiles, not
confirmed against Hermes source — treat as high-confidence, not vendor-documented.

**No Codex-CLI import happened.** `~/.codex/` exists on the VM (Codex CLI's own state dir:
`config.toml`, `hooks.json`, session/memory sqlite files) but **contains no `auth.json`** —
ruling out the "import existing `~/.codex/auth.json`" path the vendor docs mention as an
alternative. The credential was obtained by Hermes's own device-code flow directly (matches
`source: "device_code"` / `label: "device_code"` in the pool entry, and `hermes auth add
openai-codex --type oauth`'s own `--help` surface).

## GPT 5.6 Sol availability

- Real, current OpenAI model (per OpenAI's own docs/pages, WebSearch-confirmed): flagship of the
  GPT-5.6 family (Luna/Terra/Sol, ascending capability), "workhorse"/"best coding model yet",
  1.05M token context, 128K max output. Standard API pricing $5/$30 per M in/out tokens — but
  that's the pay-per-token lane, irrelevant to the OAuth path being used here.
- **Config on p-Hermes literally sets `model.default: gpt-5.6-sol`** on 4 of 5 profiles routed
  through `openai-codex` — i.e., this exact model id is already the declared default against the
  live ChatGPT-subscription credential. That is stronger evidence than any vendor doc: it's the
  production config of a working deployment.
- A third-party guide (`openclawlaunch.com/guides/hermes-chatgpt-subscription`, not primary
  vendor doc, treat as corroborating rather than authoritative) states plainly: *"Connect your
  ChatGPT Plus or Pro subscription to Hermes Agent and use GPT-5.6 Sol at zero API cost"* —
  consistent with the `base_url: https://chatgpt.com/backend-api/codex` observed live, which is
  ChatGPT's own internal Codex backend endpoint (the same one the ChatGPT web/desktop app and
  Codex CLI use when signed in with "Sign in with ChatGPT"), not `api.openai.com`.
- **Not independently verified here:** an actual inference call was not made (out of read-only
  recon scope, and not needed — `hermes auth status` + live production config.yaml is sufficient
  proof of wiring). Which ChatGPT plan tier (Plus vs Pro vs Team) is required for Sol
  specifically, and whether that tier is what's active on `gaetan.cathelain@mobile.club`, is not
  visible from redacted token shapes (no `account_id`/`id_token`/tier field populated in this
  schema instance) — the fact that `default:` is already set to `gpt-5.6-sol` and the provider
  shows `logged in` is the strongest available evidence it works, short of placing a live call.

## What lane A must do, step by step

Given the credential is already live at the base-profile level and D2 already decided **Tars IS
the new VM's base `~/.hermes` profile** (no `hermes profile create`), this is a single command,
not a multi-step integration:

1. Install Hermes on the new VM (lane B, B4, unaffected by this — `hermes setup` deferred/blank
   per DECISION.md D2, so B4 does **not** need this step).
2. Lane A, on the new VM, in Gaetan's browser session (same "one browser" constraint as every
   other lane-A credential): `hermes auth add openai-codex --type oauth`. This opens a
   device-code URL; Gaetan authorizes it against `gaetan.cathelain@mobile.club`'s ChatGPT
   session in the browser tab. **Do not pass `--api-key`** (that's the API-key path, wrong for
   this decision) and default `--type oauth` unless the flag turns out required explicitly on
   `v0.20.0`.
3. Set `model.provider: openai-codex` and `model.default: gpt-5.6-sol` in the new profile's
   `config.yaml` (mirrors `default`/`immichbackup`/etc., **not** `tars`'s stale `gpt-5.5` —
   PLAN.md's target is Sol, use it explicitly rather than whatever the picker defaults to).
   `base_url` appears to be set automatically by `hermes auth add`/`hermes model` to
   `https://chatgpt.com/backend-api/codex`; confirm it's present after step 2 rather than hand
   -writing it.
4. Probe: `hermes auth status openai-codex` → expect `openai-codex: logged in`; `hermes auth
   list` → expect one `openai-codex` credential, `source: device_code`. Async probe agent per
   PLAN.md's pattern — this is the credential's live-proof step, not a curl-based API check like
   the other 8 D5 targets (there is no bearer-token curl probe for an OAuth-managed provider;
   `hermes auth status`/`hermes model` output *is* the probe).
5. **Do not copy p-Hermes's `auth.json` blob to the new VM, and do not attempt to mirror it into
   SOPS as a static secret.** Two independent reasons, both load-bearing:
   - It's a rotating-refresh-token OAuth credential (evidence: `default`'s `auth.json` mtime is
     newer than `tars`'s despite both once holding matching data — a refresh happened on one
     side and did not propagate). If a copied refresh token got used from two machines, standard
     OAuth rotation semantics mean whichever host refreshes first can invalidate the other's
     copy — risking breaking the **live production** p-Hermes Telegram/other-profile traffic
     that depends on it, for zero benefit (the new VM doesn't need p-Hermes's specific token; it
     needs its own).
   - Hermes already self-manages refresh (`last_refresh` fields update automatically); there is
     nothing to durably store beyond "lane A ran `hermes auth add openai-codex --type oauth` as
     `gaetan.cathelain@mobile.club`, verified `logged in`" — record that fact (account, provider
     name, verification command + result) in `status/lane-a.md`, not a credential blob or SOPS
     key. This is a deliberate departure from D5's SOPS-for-everything default, scoped narrowly
     to this one OAuth-managed provider — everything else in D5's table (static API keys,
     harvested cookies, Slack bot tokens) is still SOPS as decided.
6. Update DECISION.md's D5 table row `A1b` from "TBD (`auth.json`, not `.env`)" to: harvest =
   `hermes auth add openai-codex --type oauth` (browser device-code, Gaetan's ChatGPT session);
   probe = `hermes auth status openai-codex`; storage = **none in SOPS** — status-file note only,
   per point 5 above.

## Open questions

- **Why does `tars`'s `config.yaml` pin `gpt-5.5` instead of `gpt-5.6-sol`** like every sibling
  profile? Not investigated further — it's the profile being deleted at cutover (D2), not
  carried forward, and PLAN.md's target for the new VM is explicitly Sol regardless of what the
  old profile happened to pin. Flagging only so lane A doesn't accidentally copy `tars`'s
  `config.yaml` model block verbatim as a shortcut.
- **The fallback mechanism (profile-less-`auth.json` → base `auth.json`) is inferred from
  behavior, not from source or docs.** High confidence (4/4 profiles tested agree, including one
  with zero local `auth.json`), but if D2's "Tars is the base profile" plan ever changes to a
  *named* profile under a multi-profile install later, re-verify this fallback still applies
  before assuming a single base-level OAuth login covers a named Tars profile too.
  Not a concern under the current decision (no named profile planned).
  **This point is now moot for lane A's step 2** because the new VM's Tars *is* the base profile
  — there's no second profile for the credential to need to "fall back" from.
- **`v0.19.0` (probed) vs `v0.20.0` (DECISION.md's pinned install target) flag drift**: `hermes
  auth add --help`'s exact flags should be re-checked with `--help` on the actual `v0.20.0`
  binary before lane A scripts anything non-interactively — standing caveat, same as R1's for
  every other CLI surface.
- **Which ChatGPT plan tier is active on `gaetan.cathelain@mobile.club` and whether it's
  sufficient for Sol specifically** (vs. e.g. only unlocking Terra) is not visible from this
  probe. Not blocking — the live config already declares Sol as default and the provider is
  `logged in`; if Sol turns out unreachable in practice, that will surface as an inference-time
  error the first time Tars is actually used, not as an auth-time failure. Worth a real chat
  turn as part of WF4's verify pass rather than resolving here.
- **Whether `hermes model` (the Nous-portal-flavored interactive picker) can *also* drive the
  `openai-codex` OAuth flow**, vs. `hermes auth add openai-codex` being the only path — not
  fully resolved; `hermes model --help`'s flags are all Nous-specific, suggesting `hermes auth
  add` is the correct, more explicit command, but the picker's interactive menu wasn't run live
  (would require driving an actual OAuth browser flow, out of read-only recon scope).

## Sources

- Live, read-only, `ssh root@192.168.0.3` → `qm guest exec 103` → `runuser -u hermes`: `stat`/
  `jq`(via piped-stdin script, no files written) over all 6 profiles' + base `auth.json` and
  `config.yaml`; `hermes --version`, `hermes auth --help`, `hermes auth add --help`, `hermes auth
  status openai-codex`, `hermes auth list` (under 4 different `HERMES_HOME`s); `ls ~/.codex/`.
- `hermes-agent.nousresearch.com/docs/integrations/providers`,
  `/docs/user-guide/features/codex-app-server-runtime`,
  `/docs/user-guide/configuration`, `/docs/getting-started/quickstart` (WebFetch, LLM-summarized
  reads).
- `github.com/NousResearch/hermes-agent/blob/main/website/docs/integrations/providers.md`,
  `github.com/NousResearch/hermes-agent/issues/15167` (`credential_pool` schema corroboration).
- WebSearch: `openclawlaunch.com/guides/hermes-chatgpt-subscription` (third-party, corroborating
  only); `developers.openai.com/api/docs/models/gpt-5.6-sol`, `openai.com/index/gpt-5-6/` (model
  facts).
- `~/dev/se-nmc496/config.yaml`, `.env.example` — checked, uses `provider: anthropic` for that
  deployment; no OpenAI/Codex config present, not directly relevant beyond confirming the
  `model: {provider, default}` config.yaml shape is consistent across independent Hermes
  deployments.
- `which hermes` / `hermes --help` on cooper — not installed (`command not found`); cooper has
  only the `~/.hermes` stub described in R1 (Orca's status-hook plugin dir), no real profile.
- This repo: `PLAN.md` §Model backend, `docs/recon/DECISION.md` Blocker 1, `docs/recon/r1-hermes.md`,
  `docs/recon/r3-p-hermes.md`, `docs/recon/r6-credentials.md` (grepped, confirms zero prior
  mention of openai/codex/chatgpt/gpt — Blocker 1's claim that R6 omitted this target is
  accurate).
