# npm audit fix — Tars VM Hermes npm workspaces

Follow-up to B7's flagged item (`status/lane-b.md` §B7 "Flagged, outside lane B's mandate"):
`hermes doctor` reported 6 npm vulnerabilities across bundled workspaces (agent-browser 2, web 3,
ui-tui 1). Gaetan's call: "fix all of those." Done, no `--force`, nothing outside scope touched.

VM: `gaetan@192.168.0.9`. Install tree: `~/.hermes/hermes-agent` (git checkout, HEAD
`6e87d43a5765d4480359cfa57b574e97bc5eacf4` — matches `hermes doctor`'s `version 0.20.0 [6e87d43a]`).
`node`/`npm` are symlinked into `~/.local/bin`, not on the default non-interactive SSH `PATH` — every
command below was run with `PATH="$HOME/.local/bin:$PATH"` exported first.

## Before

`hermes doctor` (verbatim from the affected lines):

```
⚠ Browser tools (agent-browser) deps (0 critical, 2 high, 0 moderate — run: cd /home/gaetan/.hermes/hermes-agent && npm audit fix --workspaces=false)
⚠ web workspace deps (0 critical, 3 high, 0 moderate — build-tool advisory; clears via lockfile bump)
⚠ ui-tui workspace deps (0 critical, 1 high, 0 moderate — build-tool advisory; clears via lockfile bump)
...
Found 5 issue(s) to address:
  1. Run 'hermes setup' to configure API keys
  2. Browser tools (agent-browser) has 2 npm vulnerabilities
  3. web workspace has 3 npm vulnerabilities
  4. ui-tui workspace has 1 npm vulnerability
  5. Run 'hermes setup' to configure missing API keys for full tool access
```

Scoped `npm audit` per workspace (same scoping doctor uses) confirmed the counts and gave advisory
IDs:

| Workspace | Command | Vulns | Package(s) | Advisory |
|---|---|---|---|---|
| root (agent-browser dep) | `npm audit --workspaces=false` | 2 high | `brace-expansion` 4.0.0–5.0.8, `minimatch` (depends on it) | [GHSA-rgw5-rvv9-x895](https://github.com/advisories/GHSA-rgw5-rvv9-x895) — DoS via unbounded intermediate arrays, bypasses the CVE-2026-14257 mitigation |
| `web` | `npm audit --workspace web` | 3 high | `brace-expansion`, `minimatch` (same as above, hoisted copy) + `undici` 7.0.0–7.28.0 (transitive via `vitest → jsdom`, dev/build-time only) | GHSA-rgw5-rvv9-x895, plus undici: [GHSA-8xcm-r25x-g524](https://github.com/advisories/GHSA-8xcm-r25x-g524), [GHSA-4cwx-7wf7-3272](https://github.com/advisories/GHSA-4cwx-7wf7-3272), [GHSA-m8rv-5g2x-5cg5](https://github.com/advisories/GHSA-m8rv-5g2x-5cg5), [GHSA-jr45-8vmc-qm54](https://github.com/advisories/GHSA-jr45-8vmc-qm54), [GHSA-v3r7-h72x-cjcm](https://github.com/advisories/GHSA-v3r7-h72x-cjcm) |
| `ui-tui` | `npm audit --workspace ui-tui` | 1 high | `undici` — combined range `<=6.27.0 \|\| 7.0.0-7.28.0` (its own direct pin `6.27.0` nested at `ui-tui/node_modules/undici`, plus the same hoisted 7.x copy as `web`) | same undici advisories as above |

Total: **6 vulnerabilities** (0 critical, 6 high, 0 moderate), all with `fixAvailable` non-major
per `npm audit --json`.

## Root cause of why plain `npm audit fix` (no edits) did nothing

`npm audit fix --workspaces=false` (the exact command `hermes doctor` suggests) ran clean
("up to date, 0 packages changed") and **left the two root vulnerabilities in place**. Root
package.json pins an `overrides` block:

```
"overrides": { "lodash": "4.18.1", "yauzl": "^3.3.1", "protobufjs": "^8.7.1", "brace-expansion": "5.0.8" }
```

`5.0.8` is the exact last **vulnerable** version (the advisory title says it "bypasses the
CVE-2026-14257 mitigation" — 5.0.8 *was* that earlier mitigation, now itself found insufficient).
`npm audit fix` will not rewrite an `overrides` pin without `--force`, so it can't move past 5.0.8
on its own. Same shape for `ui-tui`: `undici` is pinned as an **exact** version (`"6.27.0"`, no
range) in `ui-tui/package.json`'s own `dependencies`, so audit fix can't bump it either without
touching the manifest.

Both are one-line **patch** bumps to the exact version the advisory says fixes it — not the
`--force`/major case the task said to hold off on and report back.

## Fix applied (no `--force` anywhere)

1. `~/.hermes/hermes-agent/package.json` — bumped the `overrides` pin:
   `"brace-expansion": "5.0.8"` → `"5.0.9"` (latest published version; fixes the advisory).
2. `~/.hermes/hermes-agent/ui-tui/package.json` — bumped the direct dependency pin:
   `"undici": "6.27.0"` → `"6.28.0"` (patch; clears ui-tui's own nested copy).
3. Ran, in order:
   - `npm audit fix --workspaces=false` → `changed 1 package` (brace-expansion), root audit → **0 vulnerabilities**.
   - `npm audit fix --workspace web` → `added 1 package, changed 1 package` (bumped the hoisted `undici` within `jsdom`'s accepted range, plus picked up the already-fixed brace-expansion), web audit → **0 vulnerabilities**.
   - `npm install --workspace ui-tui` (apply the new pin) then `npm audit fix --workspace ui-tui` → **0 vulnerabilities**.

Files touched on the VM (verified via `git status --short` in `~/.hermes/hermes-agent`, which is
itself a git checkout): `package.json`, `package-lock.json`, `ui-tui/package.json`. Nothing else.

## After

Scoped audits, all three:

```
$ npm audit --workspaces=false      → found 0 vulnerabilities
$ npm audit --workspace web         → found 0 vulnerabilities
$ npm audit --workspace ui-tui      → found 0 vulnerabilities
```

`hermes doctor`, same three lines, before → after:

```
⚠ Browser tools (agent-browser) deps (0 critical, 2 high, 0 moderate ...)   →  ✓ Browser tools (agent-browser) deps (no known vulnerabilities)
⚠ web workspace deps (0 critical, 3 high, 0 moderate ...)                   →  ✓ web workspace deps (no known vulnerabilities)
⚠ ui-tui workspace deps (0 critical, 1 high, 0 moderate ...)                →  ✓ ui-tui workspace deps (no known vulnerabilities)
```

Doctor's issue count: **5 → 2**. Both remaining issues are the pre-existing, credential-gated
"Run 'hermes setup' to configure API keys" lines from B7 — unchanged, not npm-related, not this
task's scope. No new `⚠`/`✗` lines appeared anywhere else in doctor's output (Python env, SSL,
required packages, config files, directory structure, gateway service, command installation,
external tools, tool availability, skills hub, memory provider — all identical to the B7
baseline modulo the three vuln lines).

## Verified nothing else broke

| Check | Result |
|---|---|
| `hermes status` | unchanged: gateway `✗ stopped` (installed but stopped), 0 sessions, 0 scheduled jobs |
| `systemctl --user is-enabled hermes-gateway` | `disabled` (exit 1) — **unchanged, never touched** |
| `systemctl --user is-active hermes-gateway` | `inactive` (exit 3) — **unchanged, never touched** |
| `hermes plugins list` | `hermes-lcm` 0.21.0-rc2 **enabled**, `rtk-rewrite` 0.1.0 **enabled** — same as B5/B7 |
| `systemctl --user is-enabled/is-active hermes-cloakbrowser` | `enabled` / `active` — unchanged from B6 |

## Out of scope, correctly left alone

The monorepo has a fourth workspace, `apps/desktop` (an Electron app), which shares the same
hoisted `node_modules` tree. `npm audit --workspace apps/desktop` still shows **5 vulnerabilities**
(dompurify, electron, js-yaml, mermaid, undici — several needing `--force`/major bumps). This is
not one of the 6 `hermes doctor` flagged, was never touched, and `git status` in the install dir
confirms no `apps/desktop` file changed. Not this task's call; flagging it here for visibility only,
same spirit as B7's original flag.

## Remaining vulns needing `--force`

None among the 6 originally flagged. All 6 cleared via non-major patch bumps. (The `apps/desktop`
vulnerabilities above are pre-existing, out of scope, and were not part of the 6 — not fixed, not
attempted.)
