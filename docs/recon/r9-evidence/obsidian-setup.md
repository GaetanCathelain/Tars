# Obsidian setup — ground truth (2026-08-10)

Scope: read-only recon on cooper + Tars VM per task brief. Privacy note: only
paths, plugin config, and structural formats are reported below — no personal
note content is quoted anywhere in this file.

## 1. Where the vault lives

- `~/.config/obsidian/obsidian.json` does **not exist on cooper** — the
  Obsidian desktop app is not installed/registered on this machine
  (`which obsidian` → not found; no `obsidian.json`, no `*obsidian*` app
  files under `~/` besides recon artifacts below).
- No `.obsidian` directory exists anywhere on cooper's live filesystem
  (`find ~/dev ~/Documents ~/ -name .obsidian` → empty).
- The vault's canonical home is **Gaetan's Mac** (`mc-mbp` /
  `MacBook-Pro-de-Gaetan.local`), mirrored via Syncthing to the hub
  (`syncthing.home.gcath.tech`, folder id `c5efu-kvd7c`, label "Obsidian",
  path `/data-backups/Obsidian` on the hub) and to a Pixel 6. **Cooper is not
  currently a sync target for this folder.**
- Two prior-session artifacts on cooper give an actual (slightly stale,
  2026-07-30) snapshot of the vault's contents, used below for grounding:
  - `~/obsidian-vault-20260730-prereconcile.tgz` (11.6 MB tarball, macOS
    tar with AppleDouble `._*` sidecar files — a full vault export)
  - `~/vault-recon-20260730/` (a prior reconciliation task's
    manifests/diffs/reports comparing mac-live vs the Syncthing hub
    ("bigboy") vs this tarball — evidence the vault has had sync-divergence
    issues before; see `reports/divergence-analysis.md`,
    `reports/reconciliation-plan-draft.md` in that dir if more detail is
    ever needed)
  - These are **not live** — treat plugin/config facts as reliable (low
    churn) but don't assume the note tree is current.

## 2. Is the Obsidian vault the Syncthing "vault share"? Which devices?

Yes. Per `~/dev/gaetan-metarepo/machine/syncthing.md`:

- Folder `c5efu-kvd7c` / label `Obsidian` is the **only** Syncthing folder
  in the whole 13-folder mesh with a `.obsidian` directory at its root
  (verified there via `GET /rest/db/browse`).
- Shared by 4 devices: `mc-mbp`, `Pixel 6`, `MacBook-Pro-de-Gaetan.local`,
  and the hub (`syncthing`, device id `WUW6B2P…`, aka `bigboy` in cooper's
  own config). Type `sendreceive`, 441 files / 58 dirs / 26.7 MB, idle,
  0 errors at time of that recon.
- **cooper (`cooper-vm`) is NOT on the share list for this folder.**
  Confirmed independently from cooper's own live Syncthing config:
  `~/.local/state/syncthing/config.xml` (Syncthing v2.1.2 running via
  Homebrew, `pgrep -a syncthing` shows it live) lists exactly one folder —
  `rvqys-nnmxw` / "dev-mc" (`~/dev/mobile-club`). No Obsidian folder is
  configured on cooper today.
- Cooper's Syncthing already knows the hub device (`bigboy` =
  `WUW6B2P…`) and `mc-mbp` as trusted devices (`<device>` entries present
  in config.xml) — so adding the Obsidian folder to cooper would be a
  config-only change (share the existing `c5efu-kvd7c` folder from the hub
  to the `cooper` device id), not a new pairing. Not done; nobody has asked
  for it before now.

## 3. What's actually in the vault (from the 2026-07-30 snapshot)

### Kanban plugin: installed AND enabled, never used

- `Obsidian/.obsidian/plugins/obsidian-kanban/manifest.json`:
  ```json
  { "id": "obsidian-kanban", "name": "Kanban", "version": "2.0.51",
    "author": "mgmeyers", "isDesktopOnly": false }
  ```
- `Obsidian/.obsidian/community-plugins.json` lists it as **enabled**:
  `["obsidian-excalidraw-plugin","obsidian-importer","excalibrain",
  "dataview","obsidian-kanban","obsidian-tasks-plugin"]`
- **No `data.json` exists for the kanban plugin** (tar extraction of that
  path fails — file absent) — i.e. it has default settings, never
  configured.
- `grep -rl 'kanban-plugin' Obsidian/**/*.md` → **zero hits**. No Kanban
  board file exists anywhere in the vault today. The plugin is installed
  and turned on, but Gaetan has never actually created a board with it.
- Other plugins installed/enabled alongside it: `dataview`,
  `obsidian-tasks-plugin` (a checkbox/task query plugin — relevant, since
  it can aggregate `- [ ]` items across files without a dedicated board),
  `obsidian-excalidraw-plugin`, `excalibrain`, `obsidian-importer`,
  `obsidian-textgenerator-plugin`, `syncthing-integration` (a plugin that
  shows Syncthing sync status inside Obsidian).

### Confirmed file format (from the plugin's own `main.js`, not memory)

Grepped directly out of
`Obsidian/.obsidian/plugins/obsidian-kanban/main.js`:
```
"kanban-plugin";  ...  n["kanban-plugin"]=r["kanban-plugin"]||"board"
```
This confirms the on-disk format: a board is one `.md` file whose
frontmatter contains `kanban-plugin: board` (that literal key/value,
`"board"` is the default written when a new board is created). Standard
plugin behavior (matches public docs, now confirmed against Gaetan's own
installed binary): columns are `##` headings under that frontmatter, cards
are `- [ ]` / `- [x]` list items under each heading, optionally with
`#tags`, `@{date}`, or `**text**` for due-date/priority the plugin
renders specially. No board files exist yet to show a populated example.

### "To check / to test" style lists

No file matching that description was found by name. Filename search
(`*todo*`, `*a faire*`, `*checklist*`) turned up only personal-life items,
all under `Perso/Google Keep/` (an old Google Keep import folder — e.g.
`Todo.md`, `Todo general.md`, `Camping car small todo_.md`, etc. — names
only, contents not opened). The professional side of the vault
(`Pro/MC/`) is organized as `day-to-day/` (dated notes), `project notes/`
(per-project folders), `Meeting Notes/`, `Misc/`, `Onboarding/` — no
kanban-shaped or generic-todo file among the top two levels. Gaetan's
"to check / to test" list (from the task brief) is evidently **not** kept
in Obsidian today — it lives elsewhere (his own account of it, not found
here).

## 4. Tars VM (192.168.0.9) — Syncthing status

```
$ ssh gaetan@192.168.0.9 'which syncthing; pgrep -a syncthing; \
    export XDG_RUNTIME_DIR=/run/user/$(id -u); systemctl --user list-units | grep -i sync; \
    dpkg -l | grep -i syncthing'
```
All four checks returned **nothing** (exit 1, empty output for each).
Syncthing is not installed, not running, and has no systemd unit
(user or system) on the Tars VM.

### What syncing the vault to the VM would take

Nothing exotic — same pattern as `dev-mc` (folder `rvqys-nnmxw`), which
already syncs cooper ↔ hub ↔ mc-mbp:
1. Install Syncthing on the VM (it's Debian/Ubuntu-class per
   `docs/facts.md`; apt package `syncthing`, or the Homebrew build like
   cooper's if Homebrew is on the VM — not checked here, out of scope for
   read-only recon).
2. Generate its device ID, add it as a device on the hub
   (`syncthing.home.gcath.tech`), and share folder `c5efu-kvd7c`
   (Obsidian) to it — a GUI/REST action on the hub, one-time.
3. Decide `sendreceive` vs `receiveonly` on the VM. Given Tars would be
   *editing* board files (moving cards), it needs `sendreceive` — same
   as the Mac and Pixel, unlike the hub's own `receiveonly` stance on
   `dev-mc`.
4. No credentials land in this repo or on argv for this — Syncthing
   device pairing is a device-ID exchange, not a secret (consistent with
   `secrets/security.md`: device ids are public identifiers, confirmed
   in `syncthing.md`'s own "Method note").
This is a real option, not a stretch — the mesh, the hub, and the folder
already exist; only the VM-side leg is missing. Estimated effort: install
+ pair + wait for first sync, no new infra class introduced (reuses the
exact same Syncthing mesh already carrying `dev-mc` to cooper).

## Sources

- `~/.config/obsidian/obsidian.json` (absent — checked, not found)
- `~/dev/gaetan-metarepo/machine/syncthing.md` (folder/device inventory)
- `~/.local/state/syncthing/config.xml` (cooper's live Syncthing config)
- `~/obsidian-vault-20260730-prereconcile.tgz` (vault snapshot, tar listing
  + selective extraction of `.obsidian/plugins/*/manifest.json`,
  `community-plugins.json`, and `obsidian-kanban/main.js` only — no note
  bodies read)
- `~/vault-recon-20260730/` (prior recon task's manifests/reports, referenced
  not re-analyzed)
- `ssh gaetan@192.168.0.9` — `which syncthing`, `pgrep -a syncthing`,
  `systemctl --user list-units`, `dpkg -l | grep syncthing` (all empty)
