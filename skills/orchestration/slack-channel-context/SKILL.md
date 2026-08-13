---
name: slack-channel-context
description: "Map Slack channels before interpreting routed requests."
version: 0.2.0
metadata:
  hermes:
    tags: [slack, context, routing, channels]
    category: orchestration
---

# Slack Channel Context

Use this skill whenever a Slack request depends on what the current channel is for, especially when creating or routing work. Channel metadata is context, not decoration: read the current channel ID, look it up here, and interpret the request against that purpose before acting.

## Rules

1. Match on immutable channel ID, never the display name alone; names can change.
2. Use `channels_me` to verify that the channel is still accessible and to refresh its name, topic, and purpose.
3. Treat a channel's topic/purpose as the strongest source. Use recent history only when metadata is absent or ambiguous. This rations history reads **for classifying a channel's purpose only** — it never limits the reading SOUL rule 10 requires before answering a question about what happened.
4. A project-specific channel supplies project context for the request. Apply any explicit routing recorded below before a generic default.
5. Do not infer a Linear team merely from a broad department or product channel. Only mappings explicitly recorded as `Linear routing` authorize that routing.
6. Never post in `#general`; Gaetan posts company-wide announcements himself.
7. If a channel is new or has materially changed, follow the refresh procedure and update this skill before acting on channel-dependent routing.

## Explicit routing

- `C0BFQ5WFYTB` `#tech-project-support-engineer` — dedicated Support Engineer project channel. **Linear routing: NMC / Support Engineer.** Tickets originating here go directly there unless Gaetan says otherwise.
- `C0BP2GZUFSR` `#gcn-tars-reporting` — Tars reporting/home channel. No inferred company-team Linear routing.

## Channel map

Descriptions marked **metadata** come from Slack name/topic/purpose. **Observed** means recent messages were sampled because metadata was insufficient. **Inferred** means the name is the only available evidence; use it for orientation, not irreversible routing.

### Company and social

- `C7V603M9T` `#general` — company-wide announcements and work information; never post. **Metadata**
- `C7V603MK7` `#aleatoire` — non-work chat and miscellaneous social sharing. **Metadata**
- `CK91GFZPT` `#office-paris` — Paris office coordination. **Inferred**
- `C08R0FAA1HA` `#spotify-jam-01` — shared Spotify jam. **Inferred**
- `C09E33A5HL0` `#all_hands-show_and_tell` — all-hands demos/show-and-tell. **Inferred**
- `C0ADLM742VA` `#cleaq-linkedin-raid` — support Cleaq LinkedIn posts to increase visibility. **Metadata**

### Core tech collaboration

- `GQ07CQXT7` `#tech` — main engineering coordination: dailies, environments, releases, technical discussion and team availability. **Observed**
- `C068AJYLXN3` `#aleatoire-tech` — technical watch and random technical topics kept out of `#tech`. **Metadata**
- `C02R0BFT29Y` `#tech-pr` — pull-request discussions, one thread per PR. **Metadata**
- `C081A0CTT8T` `#tech-pr-gh` — GitHub PR automation/notifications. **Inferred**
- `C04LZBBNVNY` `#ci-notifs` — GitHub Actions and CI notifications. **Metadata**
- `C08P9Q67AKS` `#ci-ganesh` — CI notifications for Ganesh. **Inferred**
- `C07LS428F09` `#tech-product-retreat-2025-1` — 2025 tech/product retreat coordination or archive. **Inferred**
- `C09HMUZ6U2W` `#cybersecurity` — cybersecurity discussion and coordination. **Inferred**
- `C0AG8P87RQQ` `#tech-infra-aws` — AWS infrastructure work. **Inferred**
- `C0ARN6WS13N` `#tech-llm-devtools` — discussion of LLM development tools. **Metadata**
- `C0A2RHQEEHH` `#genai-hub` — generative-AI use cases, tools, prompts, automation ideas and practices. **Metadata**

### Support, Care and Ops

- `CC397R0HY` `#help-tech` — incoming Help Tech issues from Linear and their operational threads. **Metadata + observed**
- `C0BLPCP0APN` `#help-help-tech` — Help Tech duty rota, daily assignee, backup and coordination. **Observed**
- `CJSC5KNPP` `#help-ops` — operational help requests: shipments, disputes, accessories, Enviro and parcel exceptions. **Metadata + observed**
- `C09DM6EDX7E` `#care_ai` — Care AI project/workflow. **Inferred**
- `C0800LWF9EU` `#cleaq-care-b2b` — Cleaq B2B Care coordination. **Inferred**
- `C08JQUDH4BT` `#ops-product` — Ops/Product catalogue work, especially SKU creation, product updates and Loop/Quable sync. **Observed**
- `C0AUUF1TW83` `#cleaq-loop-questions` — questions about Cleaq/Loop behavior or data. **Inferred**
- `C0BAMV28A5P` `#help-1password` — 1Password access and usage help. **Inferred**

### Monitoring, alerts and incidents

- `CJ2BFM7JT` `#tech-alerts` — primary Datadog production alerts for API, jobs and database health. **Observed**
- `C04P2SQ8AH0` `#this-is-fire` — urgent alerts that require action or escalation. **Metadata**
- `C04P59RQT28` `#status-checks` — SaaS status aggregation. **Metadata**
- `C093AALCM61` `#tech-monito` — recurring technical monitoring session and generated monitoring reports. **Observed**
- `C0AHMHH62SG` `#tech-alert-quieter` — filtered/test Datadog alert stream intended to reduce noise. **Observed**
- `C0AF8C70F1N` `#tech-sonarly-alerts` — Sonarly incident analyses and RCA alerts. **Observed**
- `C0AET8C5QMP` `#sonarly-exchange` — shared channel with Sonarly for product updates, monitoring feedback and coordination. **Observed**
- `C0AF92PFKB2` `#nextmobiles-alerts` — NextMobiles alert stream. **Inferred**
- `C09GTLX2NG1` `#tech-tamet-transition-errors` — Tamet transition-error monitoring. **Inferred**
- `C08LHPSRK4J` `#cleaq-jobs-alert` — Cleaq background-job alerts. **Inferred**
- `C08D1S1F8RJ` `#cleaq-cron` — Cleaq scheduled-job/cron notifications and coordination. **Inferred**
- `C08LK6N95MX` `#cleaq-from-loop-automations` — Loop-to-Cleaq workflow errors. **Metadata**

### Products, migrations and data

- `C03R4LF3PGF` `#mobile-club` — Mobile Club product/business channel. **Inferred**
- `C09SNESS24A` `#nextmobiles-mobileclub` — NextMobiles/Mobile Club cross-product coordination. **Inferred**
- `C08MR2ZSVPB` `#migration-b2b-mc-to-cleaq` — B2B Mobile Club to Cleaq migration. **Inferred**
- `C0A9628QZGA` `#rebura-mobileclub` — Rebura and Mobile Club project coordination. **Inferred**
- `C0ABAUSBFKK` `#merge-nextmobile-mobileclub-data` — NextMobiles/Mobile Club data merge. **Inferred**
- `C0ACWUE2T7U` `#transfo-bc` — business-core transformation work involving Loop, Vecna, events, billing and deployment. **Observed**
- `C0AN8STFU9L` `#retro-nmc` — NMC retrospectives. **Inferred**
- `C0AM84EJPPS` `#data-quality-watch` — data-quality investigations, stock/device reconciliation and anomaly follow-up. **Observed**
- `C08SYLSUNGY` `#feedback-search` — feedback and bugs for Loop/global search behavior. **Observed**
- `C09MMPV7QH3` `#pim` — PIM/Quable catalogue import and sync with Loop/Contentful. **Metadata + observed**
- `C0A7E16QQFP` `#pim-sync` — PIM synchronization operations/notifications. **Inferred**
- `C09C619A8GL` `#ie` — Intelligent Experience/Care AI strategy, Rebura/Fin/Soupy/Ganesh and related data/infra choices. **Observed**
- `C05SVNPJU4R` `#ganesh` — Ganesh analytics/data platform, Tableau access and MCP usage. **Observed**

### Cleaq operations and commercial flows

- `C07TQHQ6ZPH` `#cleaq-contract-signed` — urgent signed-contract alerts; stated SLA under 24 hours. **Metadata**
- `C07UHP0MJSD` `#cleaq-contract-created` — Cleaq contract-created alerts. **Inferred**
- `C087EGNCCKC` `#cleaq-new-order` — Cleaq new-order alerts. **Inferred**
- `C097HS66AAK` `#cleaq-admin` — Cleaq admin product/operations. **Inferred**
- `C09GG2HGRML` `#qonto-cleaq` — Qonto/Cleaq partnership or integration. **Inferred**
- `C0BN2FL640K` `#pmi-brique3-new-site` — Brique 3 NM/MC transformation website project run with Tech Tribe. **Metadata**

### Dedicated projects and sandboxes

- `C0BFQ5WFYTB` `#tech-project-support-engineer` — Support Engineer project. **Metadata; explicit routing above**
- `C0BK3AGFQFR` `#verdict-project` — Verdict project delivery: product questions, Auth0, HubSpot/Webflow, scoring and access dependencies. **Observed**
- `C08RWSTU9LK` `#gcn-sandbox` — Gaetan's notification playground. **Metadata**
- `C0BP2GZUFSR` `#gcn-tars-reporting` — Tars reports and progress. **Explicit routing above**
- `C09J4BKHTBL` `#ia-projects` — cross-department AI project portfolio, prioritisation and vendor collaboration. **Observed**

### Partners, vendors and specialised work

- `G01FZT6K9F1` `#afone_project` — Afone mobile partner integration: APIs, activations, portability, consumption files and operational incidents. **Observed**
- `CKA509YAZ` `#kill-them-all` — competitor/market watch for device rental, financing, refurbishing and partnerships. **Observed**

## Refresh procedure

Use when Tars gains channel access, a channel is renamed, or observed use contradicts this map.

1. Call `mcp__slack__channels_me` with `channel_types: "public_channel,private_channel"` and `limit: 999`.
2. Diff returned channel IDs against every ID in this file. New IDs must be added; inaccessible IDs must be marked inaccessible rather than silently deleted.
3. For each new or ambiguous channel, prefer non-empty topic/purpose. If still unclear, call `mcp__slack__conversations_history` with `limit: "20"` and classify from recurring content, not one exceptional message.
4. Label the evidence as Metadata, Observed or Inferred. Do not promote an inferred purpose to explicit Linear routing without Gaetan's correction or an unambiguous project/team handle in channel metadata.
5. Remove duplicate rows and verify each accessible channel ID appears exactly once in the canonical map (references in other sections do not count).
6. Update this skill with `skill_manage`, read the complete resulting file, mirror it to the Tars repository, merge the PR, and prove the repository copy is byte-identical to the live skill.

## Pitfalls

- Channel names are mutable; IDs are stable.
- Recent history may reflect a temporary incident rather than the channel's purpose.
- A product name is not automatically a Linear team key.
- Notification channels are evidence sources, not implicit permission to acknowledge, mutate or reply.
- Access to a channel does not override the rule that Tars answers only Gaetan.

## Verification

A refresh is complete only when the `channels_me` result is exhausted, every accessible public/private channel ID is represented in the canonical map, every explicit routing has a source stronger than inference, and the mirrored skill is byte-identical after merge.
