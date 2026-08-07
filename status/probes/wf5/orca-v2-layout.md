# Orca v2 layout model — Projects, Repos, Worktrees, Tabs, Agent sessions

Read-only recon on cooper, runtime `e55eb94b-470d-490a-9f61-74c640329261`, app
`1.4.176`, CLI `/home/gaetan/.local/bin/orca`. Every claim below is either a
verbatim `--help` capture or a live `--json` read against real objects (this
very orca-a2a worktree, its sibling `thread-behavior`, their shared parent
`orchestrator`, and the 9-worktree mc-metarepo repo). Nothing was created,
removed, or sent. Full commands run are listed in **Sources** at the end.

## 0. Containment hierarchy (text diagram)

```
Project  (durable, host-independent identity — "github:org/repo")
  └─ Project Host Setup  ("repo" in everyday CLI use — one concrete checkout
      │                   of that project on ONE host: id, path, git username)
      │
      └─ Worktree  (a git checkout: one branch, one directory, on disk)
           │  · has parentWorktreeId / childWorktreeIds (Orca's OWN lineage
           │    tracking — see §3, this is NOT git ancestry)
           │  · has a `lineage` object recording HOW the parent was captured
           │
           ├─ Terminal(s)  (a live PTY: shell OR an agent like claude/codex
           │    │            running inside it — same object either way)
           │    │  · each terminal has a Tab (tabId) + a Pane/leaf (leafId)
           │    │    inside that worktree's terminal-tab strip
           │    │  · a Tab can hold >1 Pane via `terminal split`
           │    │  · orchestration identifies a running agent by
           │    │    paneKey = "<tabId>:<leafId>"  (confirmed byte-for-byte
           │    │    equal between `worktree ps` and `terminal list`)
           │    │
           │    └─ Orchestration Dispatch (optional) — `worker-start` binds
           │         a Task to a terminal handle and returns a durable
           │         dispatch_id; the terminal is the runtime seat, the
           │         dispatch is the durable tracking row (SQLite)
           │
           └─ Browser Tab(s)  (a SEPARATE object family: `orca tab *`,
                addressed by browserPageId, scoped to worktreeId — does
                NOT appear in the terminal-tab/pane tree at all)

(Editor / diff views: `orca file open` / `file diff` open a file in the
 Orca app's editor pane. There is no `file list`/`file show`/`file close`
 — fire-and-forget from the CLI, not a listed/addressable object.)
```

## 1. PROJECT vs REPO — Gaetan's "project" is Orca's "repo" in practice

Orca actually has **two** superimposed concepts here, and the CLI's own
"Projects:" help section is a false friend for Gaetan's vocabulary:

- **Project** (`orca project *`) — a durable, host-independent identity,
  keyed by provider+owner+repo (e.g. `github:mobile-club/metarepo`), derived
  from the git remote. `orca project list --json` on this box returns 2
  projects (mc-metarepo, Tars), each with `sourceRepoIds: [<repoId>]`.
  A project can have multiple **Project Host Setups** — e.g. one on cooper,
  one on a different Orca host — via `project setup-clone` /
  `setup-existing-folder` / `setup-create`.
- **Project Host Setup** (`orca project setups`) — the concrete registration
  of a project on ONE host: local path, git username, `kind: git|folder`,
  `worktreeBasePath`. On this box both setups have `setupMethod:
  "legacy-repo"` and their `id` is IDENTICAL to the repo's id — because both
  repos here were registered the old way, via `repo add`, before the
  Project/ProjectHostSetup split existed.
- **Repo** (`orca repo *`) — this is what you actually use day to day.
  `repo list/show/add/set-base-ref/search-refs`, and every `--repo <selector>`
  flag on `worktree create/list`, all key off `repoId`. `repo add --path
  <path>` is literally titled "Add a **project** to Orca by filesystem path"
  in its own help text — Orca's docs use the words loosely, but the flag
  that matters is `--repo`.

**Practical rule for the skill:** use `orca repo list --json` /
`--repo id:<repoId>` for everything. Only reach for `orca project *` if you
need multi-host routing (`worktree create --project <id> --host <host-id>`)
or are registering a brand-new codebase Orca hasn't seen on ANY host before.

Verified live (`orca repo list --json`):

```json
{
  "id": "8099e312-3232-46f2-83a9-97aeaf5de5a2",
  "path": "/home/gaetan/dev/mc-metarepo",
  "displayName": "mc-metarepo",
  "kind": "git",
  "gitRemoteIdentity": {
    "canonicalKey": "github.com/mobile-club/metarepo",
    "remoteUrl": "https://github.com/mobile-club/metarepo.git"
  }
}
```
2 repos registered on cooper: `mc-metarepo` (id `8099e312-…`) and `Tars`
(id `a71db494-…`, the repo this very worktree belongs to).

## 2. WORKTREE — a git checkout, one branch, one directory

`orca worktree create --name <name> [--repo <selector>] [--base-branch <ref>]
[--parent-worktree <selector>|--no-parent] [--agent <id>] [--prompt <text>]
[--activate] [--json]`. Identity is the compound id `<repoId>::<absPath>`;
selectors also accept `name:`, `branch:`, `issue:`, `path:`, or
`active`/`current`.

- **Landing path**: not fixed — `mc-metarepo`'s 9 live worktrees land in at
  least 3 different places (`~/dev/mc-metarepo-<slug>`, `~/dev/mc-metarepo/
  worktrees/mc-metarepo/<slug>`, `~/orca/workspaces/mc-metarepo/<slug>`)
  depending on `worktreeBasePath` config and creation era — do not assume a
  fixed pattern, read `path` from the JSON.
- **Branch**: defaults to a new branch off the repo's base ref
  (`repo set-base-ref` persists that default; `--base-branch` overrides
  per-call). Observed `baseRef` values: `"refs/remotes/origin/main"` for a
  worktree created straight off the registered remote, plain `"main"` for
  worktrees created from inside another worktree (see §3 — this is
  independent of the parent/child relationship).
- **List**: `orca worktree list [--repo <selector>] [--limit <n>] --json` →
  `{worktrees:[…], totalCount, truncated}`. Per-worktree fields include
  `id, instanceId, repoId, path, head, branch, isMainWorktree, displayName,
  comment, linkedIssue/PR/LinearIssue/GitLabMR/…, isArchived, isPinned,
  workspaceStatus (todo|in-progress|in-review|completed|custom),
  parentWorktreeId, childWorktreeIds, lineage, git{…}`. `worktree show
  --worktree <selector>` and `worktree current` return the same single-object
  shape.
- **Update metadata**: `worktree set --worktree <sel> [--display-name]
  [--issue] [--linear-issue] [--comment] [--workspace-status]
  [--parent-worktree|--no-parent]`. This is Orca bookkeeping only — it does
  not touch git.
- **Remove**: `worktree rm --worktree <sel> [--force] [--run-hooks]` —
  removes from both Orca and git; repo-defined `orca.yaml` archive hooks are
  skipped unless `--run-hooks`.
- **Orchestration summary**: `worktree ps [--limit n] --json` is the
  richest single read — per worktree it adds `liveTerminalCount,
  hasAttachedPty, lastOutputAt, preview, status (active|working|permission|
  idle…), agents[]` where each agent has `paneKey, parentPaneKey, state,
  agentType (claude|codex), prompt, lastAssistantMessage, toolName,
  toolInput, stateStartedAt, updatedAt` — this is the fastest way to answer
  "what is every worktree doing right now" in one call.

## 3. PARENT/CHILD WORKTREES — the part Gaetan flagged as least understood

**Ground truth, established from a live triple**: this session's own
worktree (`Tars::…/orca-a2a`) and its sibling `Tars::…/thread-behavior` are
BOTH children of `Tars::…/orchestrator`. Verified with
`orca worktree list --repo id:a71db494-… --json`:

```json
{
  "id": "a71db494-…::/home/gaetan/orca/workspaces/Tars/orca-a2a",
  "parentWorktreeId": "a71db494-…::/home/gaetan/orca/workspaces/Tars/orchestrator",
  "childWorktreeIds": [],
  "baseRef": "main",
  "cliProvenance": {"kind": "created-by-cli", "callerTerminalHandle": "term_479c7b7b-…"},
  "lineage": {
    "worktreeId": "…orca-a2a",
    "parentWorktreeId": "…orchestrator",
    "parentWorktreeInstanceId": "55b488a0-…",
    "origin": "cli",
    "capture": {"source": "env-workspace", "confidence": "inferred"},
    "createdByTerminalHandle": "term_479c7b7b-…",
    "createdAt": 1786135839506
  }
}
```
and, symmetrically, the orchestrator worktree carries
`"childWorktreeIds": ["…orca-a2a", "…thread-behavior"]`, `parentWorktreeId:
null`.

**What "parent/child" actually IS — three things it is NOT, first:**

1. **NOT a git relationship.** `orca-a2a`'s `baseRef` is plain `"main"` —
   the same base a totally unrelated top-level worktree would use, not
   `"orchestrator"`. Parent/child does not set the merge target or the
   branch-off point; `--base-branch` (or the repo's default base ref) does
   that, completely independently.
2. **NOT nesting on disk.** `orchestrator` lives at
   `~/orca/workspaces/Tars/orchestrator`; its child `orca-a2a` lives as a
   *sibling* directory at `~/orca/workspaces/Tars/orca-a2a` — not underneath
   it.
3. **NOT set by an explicit flag in the observed case.** `lineage.capture
   = {"source": "env-workspace", "confidence": "inferred"}` — Orca inferred
   the parent from the calling terminal's own worktree context
   (`createdByTerminalHandle` = the terminal that ran `worktree create`,
   itself living in the `orchestrator` worktree), not from an explicit
   `--parent-worktree` flag on that call.

**What it IS**: pure Orca-side lineage/organizational metadata — "which
worktree's terminal spawned this one" — used for UI nesting (indent a
worktree under the one that begat it) and for orchestration bookkeeping
(`worktree ps`'s tree, `worker-start --worktree new-child` routing). It is
recorded whether the spawn happened via `worktree create` or via
`orchestration worker-start`.

**How the default/opt-out actually works** (from `worktree create --help`,
confirmed by the evidence above matching it exactly):
- Default: *"Orca records the new worktree as a child of the caller context
  when it can infer one from the Orca terminal or current directory."* — no
  flag needed, inference is automatic and silent.
- `--parent-worktree <selector>` makes an inferred-or-not relationship
  **explicit** — selector forms: `active`, `id:<repoId>::<path>`,
  `branch:<b>`, `issue:<n>`, `path:<p>`, `folder:<folderId>` (a non-git
  "folder"-kind project setup can also be a parent), or `worktree:<id>`.
- `--no-parent` is the true opt-out — *"Force no parent lineage for
  unrelated work"* — used for independent top-level worktrees. Every
  mc-metarepo worktree inspected (`worktree list --repo id:8099e312-… --json`,
  9 worktrees) had `parentWorktreeId: null, lineage: null` — i.e. all were
  created top-level / independent, or predate the lineage feature; only the
  Tars repo's set (created more recently, from inside other worktrees'
  terminals) shows real lineage. So expect `lineage: null` to be the common
  case, not the exception, on any repo whose worktrees were mostly created
  by hand from outside Orca-managed terminals.
- `orchestration worker-start --worktree <current|selector|new-child|
  new-top-level>` exposes the SAME choice at the orchestration layer with
  explicit enum values — `new-child` vs `new-top-level` — rather than a
  flag/no-flag pair. (`new-child` and `current` are invalid when `--on`
  targets a *remote* server — "discover an exact remote selector or use
  new-top-level" per its own `--help`.)
- **Seeing the tree**: `childWorktreeIds` on the parent + `parentWorktreeId`
  on the child is the whole graph; `worktree ps --json` additionally
  surfaces `lineageWorktreeInstanceId` and `parentWorktreeInstanceId` (the
  live *instance* ids, vs. the durable path-based ids) for the same edges.
  There is no dedicated "worktree tree" command — walk the parent/child
  pointers yourself from one `worktree list --repo … --json` call.

## 4. TABS — Orca overloads this word for THREE distinct things

Gaetan's framing ("tabs of type browser, Claude Code, terminal, …") does
not match a single Orca object. There are three separate, non-overlapping
CLI surfaces:

| What a human would call a "tab" | Orca CLI family | Addressed by | Lives in |
|---|---|---|---|
| A terminal running a shell OR an agent (claude/codex) | `orca terminal *` | `--terminal <handle>` (`term_…`), plus `tabId`+`leafId` | the worktree's terminal-tab strip (`visualLayouts`) |
| A browser page | `orca tab *` | `--page <browserPageId>` or `--index` | a flat per-worktree list, `orca tab list --worktree <sel|all>` |
| A file/diff view in the editor | `orca file open` / `orca file diff` | nothing — fire-and-forget | the Orca app's editor pane, NOT listed/closable via CLI |

**Terminal-tab/pane topology** (`orca terminal list --worktree <sel>
--include-visual-layouts --json` — plain `terminal list` omits this):

```json
"visualLayouts": [{
  "worktreeId": "…orca-a2a",
  "root": {
    "type": "group",
    "activeTabId": "8843bf18-…",
    "tabs": [{
      "tabId": "8843bf18-8e6e-41d7-9d33-fec87d21f524",
      "title": "orca-a2a",
      "activeLeafId": "27a70e0c-7c79-4989-90b7-8e0c04866c97",
      "panes": {
        "type": "terminal",
        "handle": "term_37db3f88-8b04-4527-bac6-3889baf09b5f",
        "tabId": "8843bf18-…", "leafId": "27a70e0c-…",
        "title": "⠐ Orca-integration-v2 and A2A investigation for Tars",
        "connected": true, "active": true
      }
    }]
  }
}]
```
Each worktree owns one `root` group; a `tab` has one `activeLeafId` and a
`panes` tree (`orca terminal split` turns a single `"type":"terminal"` leaf
into a `"type":"split"` node with 2+ terminal-leaf children — not observed
live here since no split terminals existed in any inspected worktree, but
implied by the split/close-pane-vs-close-tab distinction in `terminal
close --help`: *"Without --tab, preserves the existing pane/session close
behavior. With --tab, waits until the whole tab is durably removed."*). One
tab = one or more terminal panes; a `--command "codex"`/`"claude"` terminal
IS the "Claude Code tab" — there's no separate agent-tab type, an agent
session is just a terminal whose process happens to be an agent CLI.

**Terminal verbs**: `create [--worktree] [--title] [--command] [--focus]`
· `list [--worktree] [--limit] [--include-visual-layouts]` · `show
[--terminal]` · `read [--terminal] [--cursor] [--limit]` · `send [--terminal]
[--text] [--enter] [--interrupt]` · `wait [--terminal] --for exit|tui-idle
[--timeout-ms]` · `split [--terminal] [--direction h|v] [--command]` ·
`switch`/`focus` (alias) `[--terminal]` · `close [--terminal] [--tab]` ·
`stop --worktree <sel>` (kills every terminal in a worktree) · `rename
[--terminal] [--title]`.

**Browser-tab verbs** (fully separate subsystem, confirmed empty for this
worktree, non-empty and live on `orchestrator`'s two Tailscale/GitHub
tabs — `orca tab list --worktree all --json`):
`create [--url] [--worktree] [--profile]` · `list [--worktree|all]
[--show-profile]` · `show --page <id>` · `current [--worktree|all]` ·
`switch [--index|--page]` · `close [--index|--page]` · plus a `profile
list|create|delete|set|show|use-default|clone` family and the full
`snapshot/click/fill/goto/…` page-automation verb set (all machine-drivable,
none of it terminal-shaped).

**Editor/diff**: `file open <path> [--worktree]`, `file diff <path>
[--staged] [--worktree]`, `file open-changed [--mode edit|diff|both]
[--worktree]` — all write-only from the CLI's point of view: they tell the
app to open something, return success/failure, and give you nothing back to
list, read, or close later. **Human-only surface** in practice — a CLI-only
agent cannot re-discover or manage editor/diff tabs it opened.

**CLI-drivable vs human-only, summary**: terminal panes and browser tabs are
both fully CLI-drivable (create/list/show/read-or-snapshot/send-or-interact/
close). Editor/diff tabs are CLI-triggerable but not CLI-observable —
effectively human-only once opened.

## 5. AGENT SESSIONS — the containment graph a skill reader most needs

An "agent session" (claude or codex actually thinking/working) is **not** a
distinct Orca object with its own id. It is a **terminal** whose live
process is an agent CLI, observed through three different lenses that all
key off the same underlying pane:

```
Worktree  a71db494-…::/home/gaetan/orca/workspaces/Tars/orca-a2a
   │
   ├── Terminal   handle: term_37db3f88-8b04-4527-bac6-3889baf09b5f
   │      tabId:  8843bf18-8e6e-41d7-9d33-fec87d21f524
   │      leafId: 27a70e0c-7c79-4989-90b7-8e0c04866c97
   │      (this triple is returned by `terminal list`/`terminal show`)
   │
   └── same pane, seen from `worktree ps`'s agents[]:
          paneKey = "8843bf18-8e6e-41d7-9d33-fec87d21f524:27a70e0c-7c79-4989-90b7-8e0c04866c97"
                      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                                 = tabId                                = leafId
          parentPaneKey: null        (non-null only for a split-off pane)
          agentType: "claude"
          state: working|waiting|done|permission|idle
```
Verified byte-for-byte: the `paneKey` string in `worktree ps --json`'s
`agents[]` for this worktree is the literal concatenation
`"<tabId>:<leafId>"` from `terminal list --include-visual-layouts --json`
for the same worktree — this is the load-bearing join key between the two
views.

**Orchestration adds a fourth, optional lens** — only present when the
agent was launched via `orchestration worker-start` rather than a bare
`terminal create --command claude`:
```
Task (task_id)  →  worker-start --task <id> --worktree <current|sel|new-child|new-top-level> --agent claude
                        │
                        ▼
                Dispatch (dispatch_id)  — durable, SQLite-backed row, survives
                        │                 terminal churn (worker_dispatches table,
                        │                 per prior recon in orca-v2-capabilities.md)
                        ▼
                bound to one Terminal (the terminal IS the runtime seat;
                the dispatch is the durable pointer to it — `worker-show
                --dispatch <id>`, `worker-read`, `worker-stop`,
                `worker-release` all operate by dispatch_id, not by the
                terminal handle, precisely so a stale/dead terminal doesn't
                strand the tracking)
```
`worker-start`'s own notes make the worktree-choice/parent-choice explicit
at this layer too: *"Current and existing worktrees never rerun setup; a
fresh agent terminal is created unless --terminal is explicit"* and *"New
worktrees use agent-first creation"* — i.e. `new-child`/`new-top-level`
under `worker-start` go through the exact same lineage machinery as
`worktree create --parent-worktree`/`--no-parent` in §3, just spelled with
an enum instead of a flag pair.

**Net**: Worktree → Terminal (⊂ Tab ⊂ Pane/leaf) is the durable containment;
Dispatch is an optional durable POINTER to a Terminal, not a container of
one. A terminal can exist with no dispatch (hand-created via `terminal
create --command claude`); a dispatch always points at exactly one terminal
handle at a time (a `--retry-of` replacement gets a NEW dispatch_id pointed
at a new terminal, "does not inherit placement").

## 6. Vocabulary mapping

| Gaetan's word | Orca's word | CLI flag / selector | Notes |
|---|---|---|---|
| Project | **Repo** (practically) / Project (formally) | `--repo <selector>` (id:/name:/path:) | Use `repo`; `project` only for multi-host or brand-new registration |
| (repo's remote identity) | Project | `--project <id>` (`github:org/repo`) | rarely needed directly |
| (repo on this machine) | Project Host Setup / Repo | `--host <host-id>`, `--project-host-setup <id>` | `id` collides with `repoId` for legacy-registered repos |
| Worktree | Worktree | `--worktree <selector>` (id:/name:/branch:/issue:/path:/active) | git checkout, 1 branch, 1 dir |
| Parent/child worktree | "lineage" (parentWorktreeId/childWorktreeIds) | `--parent-worktree <sel>`, `--no-parent` | Orca-only bookkeeping, NOT git ancestry, NOT disk nesting |
| Tab (browser) | Tab | `--page <browserPageId>`, `--index` | `orca tab *` family, worktree-scoped |
| Tab (agent/terminal) | Tab + Pane(leaf) inside a Terminal's `visualLayouts` | `--terminal <handle>` | `orca terminal *` family |
| Tab (editor/diff) | (unnamed — "workspace file"/"diff" opened "in the Orca editor") | `orca file open/diff <path>` | no list/show/close verb exists |
| Agent session | (no dedicated noun — a Terminal running an agent) | `--terminal <handle>` or `--dispatch <dispatch_id>` | see §5 graph |
| A run of delegated work | Task / Dispatch / Run | `--task <task_id>`, `--dispatch <dispatch_id>`, `--run <run_id>` | orchestration layer, separate from worktree/terminal identity |

## 7. `--help` appendix (verbatim, commands most load-bearing for layout)

<details><summary>orca --help (top level)</summary>

```
orca

Usage: orca <command> [options]
[... full text captured live, structure matches §0-§6; abbreviated here to
keep this file's length sane — the complete top-level --help output is
reproduced once, verbatim, at the top of this recon's raw tool transcript.
Sections present: Startup, Diagnostics, Agent Discovery, Accounts, Skills,
Environments, Environment Recipes, Automations, Projects, Repos, Worktrees,
Files, Terminals, Orchestration, Computer Use, Linear, Mobile Emulator,
Browser Automation, plus Common Commands / Selectors / Examples blocks.]
```
</details>

**`orca worktree create --help`**
```
Usage: orca worktree create --name <name> [--repo <selector>|--project <id> [--host <host-id>]|--project-host-setup <id>] [--agent <id>] [--prompt <text>] [--setup run|skip|inherit] [--base-branch <ref>] [--issue <number>] [--linear-issue <identifier-or-url>] [--comment <text>] [--parent-worktree <selector>] [--no-parent] [--run-hooks] [--activate] [--json]

Notes:
  This creates a new checkout. For a fresh agent in an existing worktree, use `orca terminal create --worktree active --command "codex"` instead.
  By default, Orca records the new worktree as a child of the caller context when it can infer one from the Orca terminal or current directory.
  If --repo is omitted, Orca infers the repo from the current Orca-managed worktree.
  Use --project with --host to create on a ready project host setup without spelling the backing repo id.
  For related work, use the inferred parent or pass --parent-worktree active, folder:<id>, or worktree:<worktreeId> to make the relationship explicit. Worktree ids are the full <repo-id>::<path> values returned by `orca worktree list --json`.
  Use --no-parent when the new worktree should be independent of the current context.
  --no-parent only affects Orca lineage; omit --base-branch to use the repo default base, or pass the default base ref explicitly for independent top-level work.
  By default this creates the worktree and its first terminal without switching the active Orca view.
  Pass --agent to launch an agent in the first terminal; --prompt sends initial work to that agent.
  With --agent --json, read the new agent handle from result.agentTerminalHandle; older runtimes return only result.startupTerminal.handle, and may return neither for folder-based repos.
  Repo-defined setup hooks follow the repository setup policy; pass --setup run to force them.
  Pass --activate when the CLI caller intentionally wants to reveal the new worktree in the app.
  Passing --run-hooks is kept as a legacy alias for --setup run and reveals the worktree.
```

**`orca worktree set --help`** — `--worktree <sel> [--display-name] [--issue]
[--linear-issue] [--comment] [--workspace-status] [--parent-worktree|
--no-parent] [--json]`. Notes: workspace-status ids match board columns
(todo/in-progress/in-review/completed, or custom); `--linear-issue null`
clears the link.

**`orca worktree rm --help`** — `--worktree <sel> [--force] [--run-hooks]
[--json]`. Note: repo-defined `orca.yaml` archive hooks skipped unless
`--run-hooks`.

**`orca terminal create --help`** — `[--worktree <sel>] [--title <name>]
[--command <text>] [--focus] [--json]`. Notes: creates a visible terminal
tab without switching focus unless `--focus`; "Use this, not worktree
create, for a fresh agent in the current checkout."

**`orca terminal list --help`** — `[--worktree <sel>] [--limit <n>]
[--include-visual-layouts] [--json]`. Note: JSON omits `visualLayouts` by
default.

**`orca terminal split --help`** — `[--terminal <handle>] [--direction
horizontal|vertical] [--command <text>] [--json]`.

**`orca terminal close --help`** — `[--terminal <handle>] [--tab] [--json]`.
Note: without `--tab` closes just the pane/session; with `--tab` waits for
durable removal of the whole tab.

**`orca terminal switch --help`** (`focus` is an alias) — `[--terminal
<handle>] [--json]` — "Switch to a terminal tab in the UI".

**`orca tab create/list/show/current --help`** — worktree-scoped browser
tab CRUD; `list` accepts `--worktree <sel|all>`; `show` needs `--page <id>`
("Stable browser page id from `orca tab list --json`").

**`orca file open/diff/open-changed --help`** — path may be relative to
the worktree or absolute inside it; `open-changed` defaults to `--mode diff`
and, per its own note, is "v1" (git-status-driven) and skips deleted files
in edit mode.

**`orca orchestration worker-start --help`**
```
Usage: orca orchestration worker-start --task <task_id> [--on <saved-environment>] [--worktree <current|selector|new-child|new-top-level>] (--agent <agent> | --terminal <handle>) [--model <id>] [--effort <level>] [--name <name>] [--repo <selector>] [--base-branch <ref>] [--display-name <text>] [--comment <text>] [--setup <run|skip|inherit>] [--retry-of <dispatch_id>] [--timeout-ms <n>] [--run <run_id>] [--from <handle>] [--retry-request <id>] [--json]

Notes:
  Current and existing worktrees never rerun setup; a fresh agent terminal is created unless --terminal is explicit.
  --model supports Claude, Codex, and Cursor opaque provider model ids; --effort requires --model. Neither can combine with --terminal.
  New worktrees use agent-first creation and default --setup to run.
  Creation flags are rejected for current/existing worktrees.
  --on selects only the worker server; the Run and this command remain on the current Orca server.
  Remote current and new-child are invalid; discover an exact remote selector or use new-top-level.
  --retry-of links the replacement attempt but does not inherit placement.
  The call exits 0 only for ready. Failed or outcome_unknown exits 1 and JSON includes stage/failedStage, setup, effects, residualResources, and recovery commands when needed.
```

**`orca orchestration worker-show --help`** — `--dispatch <dispatch_id>
[--json]` — "Inspect one supervised worker Dispatch".

Everything else asked for (`repo list/show/add/set-base-ref/search-refs`,
`worktree list/show/current/ps`, `terminal show/read/send/wait/stop/rename`,
`tab switch/close/profile *`) was captured live and quoted inline in §1-§5;
omitted here only to keep this appendix from duplicating the whole CLI
surface. Raw transcripts for all of them are in this task's tool-call
history if a byte-exact re-check is ever needed.

## 8. Sources (every command run, in order)

```
orca --help
orca repo list --help ; orca repo show --help
orca worktree create/list/show/set/rm/ps --help
orca status --json
orca repo list --json
orca worktree list --repo id:8099e312-3232-46f2-83a9-97aeaf5de5a2 --json     # mc-metarepo, 9 worktrees
orca worktree ps --json                                                     # cross-repo live snapshot
orca worktree list --repo id:a71db494-6e32-437d-be39-acff38061723 --json    # Tars, parent/child evidence
orca terminal list --worktree active --include-visual-layouts --json
orca tab list --worktree active --json
orca worktree current --json
orca terminal list --worktree path:/home/gaetan/orca/workspaces/Tars/thread-behavior --include-visual-layouts --json
orca terminal list --worktree path:/home/gaetan/dev/mc-metarepo-mc-kestra --include-visual-layouts --json
orca tab list --worktree all --json
orca terminal split/switch/close/stop --help
orca file open/diff/open-changed --help
orca agent-context --json      # schemaVersion 1, commandCount 223
orca project list --json
orca project setups --json
orca repo add --help ; orca repo set-base-ref --help ; orca repo search-refs --help
orca orchestration worker-start/worker-show/worker-list --help
Cross-read: status/probes/wf5/orca-v2-capabilities.md (prior session, §3/§6/§8 —
  confirmed no contradiction; that recon did not cover parent/child worktrees
  in depth, which is the gap this file fills)
```
