---
name: mobile-club-okr-report
description: Apply the mobile.club board design to HTML reports.
version: 0.2.0
author: Gaetan Cathelain, Hermes Agent
license: Proprietary
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [design, report, template, html, mobile-club]
    related_skills: [claude-design]
---

# mobile.club Report Template

Apply the visual language of the mobile.club board report to any executive HTML report. The source file `templates/report-okr-template.html` is a visual specimen and component library, not a fixed OKR content schema: reuse its design tokens, typography, surfaces, grids, status treatments, gauges, and responsive behavior while adapting the information architecture to the report being produced.

## When to Use

- Create an executive, project, operational, audit, strategy, or KPI report in the mobile.club board style.
- Restyle an existing HTML report with this visual system.
- Give Tars, Cooper, or another implementation agent a shared report design reference.

Don't use for product interfaces, dashboards intended for continuous interaction, or presentation slides.

## Source Template

Read `templates/report-okr-template.html` in full before designing. It is the canonical rendered example for visual decisions.

Preserve from it:

- the palette and layered dark background;
- the typography roles and scale;
- the centered `1084px` content wrapper;
- the compact uppercase section headings and horizontal rules;
- the translucent cards, subtle borders, shadows, and corner radii;
- the metric, status, progress, milestone, risk, and provenance components;
- the `900px` responsive breakpoint and mobile stacking behavior;
- the inline SVG brand treatment;
- the self-contained HTML approach with no script or remote asset dependency.

Do not preserve by default:

- the OKR-specific section order;
- quarter clocks, objectives, KRs, gauges, or status pills when the new report does not need them;
- the original business copy, metrics, dates, sources, or number of cards;
- the original `<!-- SLOT: ... -->` inventory.

## Design System

### Palette

- Canvas: `#08081a` with soft radial indigo glows.
- Primary text: `#ffffff`; secondary text: `#b0aecc`; muted text: `#7a78a0`.
- Accent: `#7e7aff`; deep accent: `#3b45c7`; hero highlight: `#ffc000`.
- Positive: `#4ade80`; warning: `#fcd34d`; critical: `#f87171`.
- Surfaces: translucent violet/indigo fills with hairline violet borders.

### Typography

- Reading: DM Sans, Avenir Next, Helvetica Neue, system sans-serif.
- Display: Monument Extended, Arial Black, Avenir Next, sans-serif.
- Metadata and numbers: SF Mono, ui-monospace, Menlo, monospace.
- Use display type sparingly for report title, section labels, hero metrics, and short identifiers.

### Composition

- Start with a compact brand/meta row, a decisive title, and a short executive thesis.
- Build a report-specific section sequence from the content; do not force the OKR example's structure.
- Use three-column grids for comparable headline cards and two-column grids for paired narratives; collapse both to one column on narrow screens.
- Keep one dominant insight per card. Use gold only for the single strongest takeaway.
- Use status color as a narrow signal, never as a full-card fill.
- End with provenance: sources, measurement date, and report owner or context.

## Procedure

1. **Read the source report and the new content.** Identify the audience, decision, hierarchy, and required sections. Completion criterion: every content block has a purpose before layout starts.
2. **Choose components, not sections.** Reuse the template's masthead, section header, cards, pills, gauges, timelines, risk rows, and provenance only where they fit. Completion criterion: no OKR-only component remains without a content reason.
3. **Build a self-contained HTML report.** Copy the CSS and component patterns from the source template, then compose a new semantic document structure. Completion criterion: the result has no JavaScript or remote asset dependency.
4. **Keep the visual contract.** Preserve palette, type roles, spacing rhythm, wrapper width, border language, shadows, status signals, and responsive breakpoint. Completion criterion: a side-by-side view is recognizably the same design family without sharing the same report outline.
5. **Protect information quality.** Use supplied facts only; show unavailable or unmeasured values explicitly instead of inventing zeroes. Completion criterion: every metric and status has a named source and measurement date when applicable.
6. **Verify desktop and mobile.** Inspect at about `1200px` and `390px` with `browser_navigate` and `browser_vision`. Completion criterion: no horizontal overflow, clipping, overlap, or unreadable density.
7. **Verify semantics.** Use `browser_snapshot(full=true)` to confirm the report remains complete and understandable without visual styling. Completion criterion: headings, facts, statuses, and provenance are present in the accessibility tree.

## Pitfalls

- Treating the OKR report's content structure as the template instead of its design system.
- Copying old business content as placeholder text into a new report.
- Adding generic dashboard chrome, navigation, glass blur, illustrations, or extra gradients.
- Overusing gold or status colors until hierarchy disappears.
- Loading a remote display font when the established fallback stack already preserves the look.
- Publishing or uploading the result without an explicit destination request.

## Verification

- The new report uses the mobile.club palette, typography roles, surfaces, spacing, and responsive rules.
- Its information architecture fits the report rather than the OKR specimen.
- The original business content is absent unless explicitly requested.
- No remote asset or JavaScript dependency was introduced.
- Desktop and mobile renders were inspected.
- The accessibility snapshot contains the complete report.
