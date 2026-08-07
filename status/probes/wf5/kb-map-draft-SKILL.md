---
name: mc-metarepo
description: The team knowledge base (mobile-club/metarepo). Use whenever Gaetan
  asks an operational question about Cleaq, Mobile.Club, Next Mobiles, Loop,
  Vecna, WordPress, HelpTech, the group's OKRs, or "how do we do X here".
---

# mc-metarepo — the team knowledge base

Clone: `~/dev/mc-metarepo` on this VM. Private repo, 120 markdown files,
~1.4 MB. **Never read the whole thing** — route, then read one file.

## Route first, grep second

| Gaetan asks about | Read |
|---|---|
| an operational gotcha, "why does X behave like that", "is it safe to Y" | `knowledge/<slug>.md` — one file = one fact. Pick the slug below. |
| deep per-repo detail, ticket history, real queries | `learnings/<stack>/<repo>.md` (append-only, 20–70 KB — grep, don't read whole) |
| what a system *is* (stack, deploy, seams) | `<stack>/<repo>.md` (pinned) or `index/<stack>/<repo>.json` |
| DB tables/columns | `schemas/<stack>/<db>.core.md` |
| repo rules, contributing, language, PR flow | `CLAUDE.md`, `CONTRIBUTING.md`, `learnings/README.md` |

## knowledge/ — 40 notes, the filename IS the index

admin-codegen-offline · agent-browser-evidence · agent-prompt-trust-lanes ·
claude-cli-pin-is-file-ownership · datadog-apm-error-sampling ·
datadog-logs-source-not-env · datadog-monitors-as-code-api-traps ·
done-is-not-exercised · enviro-primary-warehouse · gha-oidc-subject-has-ids ·
hasura-prod-no-mutation-attribution · helptech-console-accepted-risks ·
helptech-console-deploy · hermes-agents-core-design ·
hermes-platform-primitives · linear-agent-write-surfaces ·
loop-b2c-invoicing-handover · loop-backfilled-items-missing-active ·
loop-lost-parcel-cancel-leak · loop-returned-device-stays-rented ·
mc-agents-ssh-access · mc-bastion-naming-trap ·
mutating-verifier-agents-need-isolation ·
next-customer-x-reference-not-wordpress-id · next-repos-investigate-on-release ·
nm-carrier-lost-device · nm-vs-mc-order-source · ops-list-reconcile-before-bulk ·
return-app-datadog-logs · route53-adopting-live-records ·
sudo-argv-secrets-authlog · tech-product-okr-q3-2026 ·
vecna-order-error-manual-relaunch · vecna-owns-returns ·
vecna-prod-portainer-console · vercel-prebuilt-needs-standalone ·
workspace-commit-conventions · wp-date-delivered-retractation ·
wp-feature-flags-helm · wt-load-shallow-submodules

## learnings/ sidecars

cleaq/risk · mobileclub/{return-app, support-engineer, terraform, workspace} ·
next/{vecna, wordpress-monorepo}

## Rules

1. Match the question's **meaning** to a slug, then `read_file` that one file.
   The answering note almost never uses Gaetan's words — route on meaning, not
   on keyword match.
2. Only if no slug fits: `search_files` with a file-list-first pass scoped to
   `knowledge/` and `learnings/`.
3. Never read whole: `next/next-bdc.fiches.md` (152 KB), `schemas/*.full.md`
   (100–156 KB), `learnings/mobileclub/workspace.md` (70 KB). Grep them.
4. The submodules (`vecna-api`, `wordpress`, `cleaq-api`, …) are **empty** in
   this clone. Product source code is NOT here — do not go looking for it and
   do not delegate a code-reading session to find what a note already states.
5. The clone is read-only reference. Never commit, never push from here.
