# p-Hermes gated cleanup — post-delete verification

Date: 2026-08-13 ~15:52Z. Executed by the hub session itself (destructive gate
stays with the hub), on Gaetan's explicit go given this session. Pre-delete
inventory: `phermes-cleanup-pre.md` (same directory).

## Command executed

```
ssh gaetan@192.168.0.9 "ssh phermes 'rm -rf /home/hermes/.hermes/profiles/tars/'"
```

Exit 0, no output.

## Post-delete verification (measured immediately after)

```
$ ssh gaetan@192.168.0.9 "ssh phermes 'ls /home/hermes/.hermes/profiles/tars/ 2>&1; echo ---; ls /home/hermes/.hermes/profiles/'"
ls: cannot access '/home/hermes/.hermes/profiles/tars/': No such file or directory
---
immichbackup
immichresearch
justine
papainventory
privatelocal
```

- Target directory gone.
- All five sibling profiles present and untouched.

## Out of scope, still on disk (flagged in pre-evidence)

`~/.config/systemd/user/hermes-gateway-tars.service` + its
`.d/override.conf` on p-Hermes — inert (disabled, no enable-symlink, unit's
WorkingDirectory now gone). No spec revision named them in the gated delete's
scope; removal is a separate call for Gaetan if ever wanted.
