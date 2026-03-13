# Frontend Changes — TARS v2 Phase 1

## Summary
Complete SvelteKit frontend implementation for TARS v2 Phase 1, ready to integrate with the Go backend.

## Tech Stack
- **SvelteKit 2** with **Svelte 5** (runes: `$state`, `$derived`, `$effect`)
- **shadcn-svelte** v1.1.1 (New York style, Zinc color scheme)
- **Tailwind CSS v4** via `@tailwindcss/vite`
- **TypeScript** strict mode
- **Dark mode only** (class `dark` on html root)

## What Was Built

### Infrastructure
- SvelteKit project initialization with TypeScript
- Tailwind CSS v4 with shadcn zinc dark theme (CSS custom properties)
- Vite proxy config: `/api` → `localhost:8080`, `/ws` → WebSocket
- 12 shadcn-svelte components installed: button, card, dialog, dropdown-menu, input, label, separator, skeleton, table, tooltip, badge, avatar

### Auth Flow
- `hooks.server.ts` — validates `tars_session` cookie against backend `/api/v1/auth/me`
- `+layout.server.ts` — auth guard redirects unauthenticated users to `/login` (except `/login` and `/auth/callback`)
- `/login` — centered card with "Sign in with GitHub" button → redirects to `/api/v1/auth/github`
- `/auth/callback` — loading spinner, redirects to `/` (backend sets cookie during OAuth flow)
- User data passed through layout load function

### Layout
- Sidebar (224px) with TARS branding + nav: Dashboard, Repos, Tasks (disabled, "Coming soon" tooltip)
- Top bar with user avatar dropdown (sign out action)
- Main content area with scrollable container

### Repos Page (`/repos`)
- Server-side data loading from `/api/v1/repos`
- Table with columns: Name, URL, Branch, Added
- "Add Repository" dialog: name, URL, local path (optional), default branch
- Delete confirmation dialog per row
- Empty state: icon + "No repositories yet"

### Stores
- `auth.svelte.ts` — Svelte 5 runes store: `user`, `isAuthenticated`, `logout()`
- `ws.svelte.ts` — WebSocket store scaffold: auto-reconnect with exponential backoff (3s → 30s max), `connected` state, `send()`, `onMessage()`

### API Client (`lib/api.ts`)
- Typed fetch wrapper: `get<T>`, `post<T>`, `patch<T>`, `del`
- Credentials included (cookies)
- Auto 401 → redirect to `/login`
- Server-side variant for SvelteKit load functions

## Build Status
- `svelte-check`: **0 errors, 0 warnings**
- `vite build`: **succeeds** (client + server)

## Files Created/Modified
All changes are under `frontend/` — no backend or root files touched.
