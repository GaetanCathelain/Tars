# Design V2 — UI Overhaul

Complete visual redesign of the TARS WebUI frontend. No functionality changes.

## What Changed

### app.css
- Base font size: 13px → 14px, line-height 1.4 → 1.5
- Added `--shadow-sm`, `--shadow-md`, `--shadow-lg` custom properties
- Added `.surface-gradient` utility class for elevated surfaces

### Login & Register Pages
- Card: `bg-[#18181b]` on `bg-[#111113]` background — visually distinct
- Card width: `w-[380px]`, padding `p-8`, border `border-zinc-800/50`, `shadow-lg`
- Added 🤖 emoji icon above title
- Title: `text-xl font-semibold` (was text-base)
- Inputs: `h-10`, `rounded-lg`, `bg-[#1c1c20]`, indigo focus ring
- Button: `h-10`, `bg-indigo-500`, `rounded-lg`, `shadow-sm`
- Field spacing: `space-y-5` (was space-y-4)
- Labels: `text-sm text-zinc-400 mb-2`

### Sidebar (+layout.svelte)
- Title: `text-base font-semibold` with "Orchestrator" subtitle
- Header padding: `px-5 py-5` with bottom border
- Section label: `text-[11px] uppercase tracking-[0.1em]`
- Task items: `py-2.5 rounded-lg`, `gap-3`, status dots `w-2 h-2`
- Active state: `bg-indigo-500/10 text-zinc-100`
- Hover: `bg-zinc-800/50`
- New Task button: dashed border style (`border-dashed border-zinc-700`)
- User section: larger text, wider dot

### Chat View (tasks/[id]/+page.svelte)
- Header: `px-6 py-4`, title `text-base font-medium`
- Status badge: `rounded-full px-3 py-1`
- Avatars: `w-8 h-8` (was w-6 h-6)
- Message spacing: `space-y-6`, content gap `gap-3.5`
- Content: `text-sm leading-relaxed`
- System messages: left border accent (`border-l-2 border-zinc-800 pl-4`)
- Input area: `h-10` input + button, `px-6 py-4`, `gap-3`

### WorkerCard
- Container: `bg-zinc-900/50 border-zinc-800 rounded-xl shadow-md`
- Header: `px-4 py-3`, text at `text-sm`
- Status dots: `w-2 h-2`, running uses indigo with `animate-pulse`
- Chevron: `w-4 h-4`

### Empty State (tasks/+page.svelte)
- 🤖 emoji at `text-4xl`
- "No tasks yet" at `text-lg font-medium`
- Proper `space-y-4` spacing

### Global
- All transitions: `transition-colors duration-150` (not transition-all)
- All borders: `border-zinc-800/50` for subtlety
- All focus: `focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20`
- No text smaller than 11px
