---
name: mobile-club-okr-report
description: Recreate the mobile.club executive OKR report design.
version: 0.1.0
author: Gaetan Cathelain, Hermes Agent
license: Proprietary
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [design, okr, report, html, mobile-club]
    related_skills: [claude-design]
---

# mobile.club OKR Report

Recreate the exact visual language of the mobile.club executive OKR report as a self-contained HTML page. The frozen source template is `templates/report-okr-template.html`; preserve its layout and CSS and replace only the explicitly marked `<!-- SLOT: ... -->` content.

## When to Use

- Build or refresh an OKR committee report in the established mobile.club board design.
- Reuse this design for a structurally similar executive report.
- Give Cooper or another implementation agent an exact visual starting point.

Don't use for unrelated dashboards, product UI, or generic presentation decks.

## Prerequisites

- Load this skill before drafting or implementing the report.
- Read the full template with `read_file` before changing anything.
- Gather the current reporting period, measured values, statuses, milestones, decisions, risks, and source timestamp.
- Treat business data as inputs; never infer missing metrics.

## Design Contract

### Visual language

- Dark board canvas: `#08081a`, layered radial indigo glows, white primary copy.
- Surfaces: translucent indigo cards, hairline violet borders, subtle deep shadows, 10–12 px corner radius.
- Accent: electric violet `#7e7aff`; hero highlight: gold `#ffc000`.
- Status colors: green `#4ade80`, yellow `#fcd34d`, red `#f87171`, neutral lavender `#b0aecc`.
- Typography: DM Sans/Avenir/system for reading; Monument Extended/Arial Black for display; SF Mono/ui-monospace for metadata and dates.
- Section rhythm: uppercase compact heading, long horizontal rule, then generous vertical spacing.
- Composition: branded masthead, executive thesis, quarter clock, three key figures, three objective cards, detailed KR board, milestones, committee decisions, risks, provenance.
- Preserve inline SVG branding and keep the page self-contained: no scripts and no remote assets.

### Layout rules

- Desktop content stays within the template's centered `1084px` wrapper.
- Three-column card groups collapse to one column below `900px`.
- Two-column month/risk groups collapse to one column below `900px`.
- KR rows preserve the status stripe, narrative block, gauge, and right-hand verdict; on mobile they stack without removing information.
- Use metric-aligned numerals and monospace metadata exactly as the template does.
- Do not add navigation, gradients outside the existing background, illustrations, glass blur, or decorative UI chrome.

## Procedure

1. **Clone the frozen template.** Copy `templates/report-okr-template.html` to the requested output path. Completion criterion: the output is byte-identical before slot edits.
2. **Inventory every slot.** Account for `TITLE`, `META`, `EYEBROW+H1`, `THESIS`, `CLOCK`, `KEYS`, `OBJECTIFS`, `BOARD`, `BOARD-FOOT`, `JALONS`, `DECISIONS`, `RISQUES`, and `PROVENANCE`. Completion criterion: each slot has an explicit input or is omitted only where the slot comment allows omission.
3. **Replace content only.** Change HTML inside slot boundaries and allowed inline progress values; do not edit the CSS, class names, wrapper hierarchy, SVG paths, or responsive rules. Completion criterion: a diff shows no CSS changes.
4. **Maintain executive altitude.** Lead with the quarter narrative, measured outcomes, dated tension, decisions, and structural risks. Exclude operational logs and vanity volumes. Completion criterion: every key figure carries a decision-relevant implication.
5. **Render statuses honestly.** `On Track`, `Progressing`, `Off Track`, and neutral states use their existing classes. Uninstrumented work is neutral and shown as unavailable, never silently as zero. Completion criterion: every displayed number traces to a named source and measurement date.
6. **Update time mechanics.** Set quarter-clock fill/current-marker percentages and per-KR gauge widths/ticks inline. For a closed quarter, follow the template comments: 100% clock, no “today” marker, no expected-at-date ticks. Completion criterion: percentages match the supplied dates and values.
7. **Verify at desktop and mobile sizes.** Open the local HTML with `browser_navigate`, inspect with `browser_vision`, and test at approximately 1200px and 390px widths. Completion criterion: no horizontal overflow, clipped text, overlapping labels, or missing sections.
8. **Check semantic integrity.** Use `browser_snapshot(full=true)` to confirm headings, metrics, statuses, and provenance remain readable without visual interpretation. Completion criterion: all report facts appear in the accessibility tree.

## Pitfalls

- The template is frozen: “improving” the CSS destroys design continuity.
- Monument Extended may fall back to Arial Black where the font is unavailable; preserve the stack rather than adding a network font.
- Progress widths and the quarter position are inline data, not CSS redesigns.
- The decision section is omitted when there are no open arbitrations or the quarter is closed.
- Off Track does not automatically mean failure when a KR is intentionally back-loaded; explain the governing milestone.
- Preserve “zero declarative numbers”: unmeasured means uninstrumented, not 0%.
- Never publish the output or upload it to an unlisted host unless the user explicitly asks for that destination.

## Verification

- Template loaded from `templates/report-okr-template.html`.
- All 13 named slots accounted for.
- CSS, SVG branding, class names, and responsive rules unchanged.
- No remote asset or JavaScript dependency introduced.
- Desktop and mobile renders inspected visually.
- Accessibility snapshot contains the complete report.
- Every metric has a source and last-measured date.
