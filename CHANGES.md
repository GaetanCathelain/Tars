# Design Overhaul: Linear + Obsidian Aesthetic

## Summary
Complete visual redesign of the TARS WebUI from SpacetimeDB-inspired dark/cyan theme to a refined Linear + Obsidian inspired aesthetic. Zero functionality changes.

## Files Changed

| File | What Changed |
|------|-------------|
| `frontend/src/app.css` | New theme palette, 13px base font, Inter font features, ultra-thin scrollbars, focus-visible rings |
| `frontend/src/app.html` | Added Google Fonts import for Inter and JetBrains Mono |
| `frontend/src/routes/+layout.svelte` | Sidebar: subtle wordmark, tiny uppercase section labels, smaller status dots, ghost new-task button, 150ms transitions |
| `frontend/src/routes/tasks/[id]/+page.svelte` | Chat: comment-style messages (no bubbles), small round avatars with initials, system messages as italic text, tighter spacing |
| `frontend/src/lib/components/WorkerCard.svelte` | Elevated surface card, refined status badges with icons, subtle shadow |
| `frontend/src/lib/components/Terminal.svelte` | Matched ANSI colors to new palette (indigo blues, muted reds/greens), #0a0a0a background |
| `frontend/src/routes/login/+page.svelte` | Centered card layout with border + subtle shadow, smaller type |
| `frontend/src/routes/register/+page.svelte` | Same card treatment as login |
| `frontend/src/routes/tasks/+page.svelte` | Cleaner empty state with avatar circle instead of emoji |

## Color Palette

| Token | Old | New |
|-------|-----|-----|
| bg-primary | `#0a0a0f` | `#111113` |
| bg-secondary | `#12121a` | `#18181b` |
| bg-tertiary | `#1a1a2e` | `#1c1c20` |
| border | `#2a2a3e` | `#27272a` |
| text-primary | `#e0e0e8` | `#fafafa` |
| text-secondary | `#8888a0` | `#a1a1aa` |
| accent | `#00d4ff` (cyan) | `#818cf8` (indigo) |
| success | `#00ff88` | `#34d399` |
| warning | `#ffaa00` | `#fbbf24` |
| danger | `#ff4466` | `#f87171` |

**New tokens:** `bg-elevated`, `border-subtle`, `text-tertiary`, `accent-muted`, `running`

## Build
`npm run build` passes cleanly.
