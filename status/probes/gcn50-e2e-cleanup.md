# GCN-50 E2E test-1 infra teardown (cooper)

Date: 2026-08-14. Authorized by Gaetan. All targets confirmed throwaway
before removal (self-checked each step; no destructive action taken blind).

## Pre-checks (step 1)

- `git status`/`cat greeting.txt` in
  `/home/gaetan/orca/workspaces/e2e-scratch-repo/gcn50-e2e`: only untracked
  `greeting.txt` containing `Hello!`, tracked `README.md` from the single
  `init scratch repo` commit, working tree otherwise clean. Nothing worth
  keeping.
- Scratch repo main worktree (`.../scratchpad/e2e-scratch-repo`): clean,
  single `init scratch repo` commit, only a throwaway `README.md`.
- VM `~/.hermes/lanes/1786721723.183519.md`: confirmed content matches the
  GCN-50 E2E lane (`[E2E gcn50]`, ticket GCN-50, lane `gcn50-e2e`, peer
  `tars-gcn50-e2e`) and was the ONLY file in `~/.hermes/lanes/`.

## Removed

1. **Terminals** (worktree `id:637ab5e3-6a68-4074-9f07-919c0d3bd0b4::/home/gaetan/orca/workspaces/e2e-scratch-repo/gcn50-e2e`):
   - `term_c523b614-8ddf-4430-af8c-e70debb32c1a` (title `✳ tars-gcn50-e2e`,
     the lane session) — closed via `orca terminal close --terminal <handle>
     --tab`. First call SIGHUPed the underlying claude process
     (`ptyKilled: false`, preview showed "1 jobs SIGHUPed .../claude
     --resume"); a second `terminal close` on the same handle finished the
     job (`ptyKilled: true`).
   - `term_e915d752-674f-4dce-b7d1-a8d698e4b90f` (plain shell tab, same
     worktree) — closed the same way so the worktree had zero live terminals
     before removal.
   - Verified: `orca terminal list --worktree id:637ab5e3-...::.../gcn50-e2e`
     → `totalCount: 0`.

2. **Orca worktree**:
   `orca worktree rm --worktree "id:637ab5e3-6a68-4074-9f07-919c0d3bd0b4::/home/gaetan/orca/workspaces/e2e-scratch-repo/gcn50-e2e" --force --json`
   → `{"removed": true}`.
   Verified: `orca worktree list --repo id:637ab5e3-...` no longer lists it
   (only the repo's own `main` worktree, the scratch-repo dir, remains); path
   `/home/gaetan/orca/workspaces/e2e-scratch-repo/gcn50-e2e` gone from disk
   (moved into Orca's own `.orca-worktree-trash`, not touched further).

3. **Scratch repo dir**:
   `rm -rf /tmp/claude-1000/-home-gaetan-dev-orca-worktrees-Tars-improvements/86148edc-827d-4f2c-9864-855a2a981415/scratchpad/e2e-scratch-repo`
   — confirmed gone (`ls` → No such file or directory).

4. **VM lane-index file**:
   `ssh gaetan@192.168.0.9 'rm -v ~/.hermes/lanes/1786721723.183519.md'` —
   removed, confirmed via `ls ~/.hermes/lanes/` now empty. No other file
   under `~/.hermes` touched (only `ls`/`cat`/`rm` on this one path).

## Not removed / not possible

- **Orca repo registration** (`id: 637ab5e3-6a68-4074-9f07-919c0d3bd0b4`,
  `e2e-scratch-repo`) still appears in `orca repo list` — this Orca CLI
  build has no `repo rm`/`repo remove` verb (`orca repo <command> --help`
  lists list/add/show/set-base-ref/search-refs only; `orca repo rm` errors
  "Unknown command"). Per task instructions this step is skipped. The
  registration is now a harmless dangling entry pointing at a deleted path;
  no worktree can be created under it since the directory is gone.
- **Slack `[E2E gcn50]` messages**: left in place per instructions (harmless
  markers, MCP can't cleanly delete).

## Unrelated infra — untouched, confirmed

- Only worktree id `637ab5e3-6a68-4074-9f07-919c0d3bd0b4` (repo
  `e2e-scratch-repo`) and its `gcn50-e2e` child worktree/terminals were
  touched. `orca worktree list --json` (full) and `orca repo list --json`
  were read-only greps for confirmation, no other selector was passed to any
  mutating command.
- Did not touch `gcn50-e2e-ch` / `gcn50-e2e-go` or any other session,
  worktree, or repo. Final `orca worktree list --json` grep for `gcn50`
  returns no matches at all (the removed worktree is gone; no other
  gcn50-named worktree existed at time of cleanup).
- No `ListAgents`-equivalent tool was available in this session to
  cross-check; verification relied on `orca terminal list` (0 terminals for
  the worktree) and `orca worktree list` (worktree absent) instead.
