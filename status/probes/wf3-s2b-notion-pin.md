# WF3-S2B — pin mcp/notion to digest in ~/.hermes/config.yaml — evidence

VM: `ssh gaetan@192.168.0.9` (key auth, non-interactive). `hermes` not on PATH
over ssh — used `~/.local/bin/hermes` throughout.

## 1. Current image + digest

```
$ docker images --digests | grep -i notion
mcp/notion   latest   sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9   df0d6781d03f   6 weeks ago   533MB

$ docker image inspect mcp/notion --format '{{json .RepoDigests}}'
["mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9"]

$ docker image inspect mcp/notion --format '{{.Id}}'
sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9
```

RepoDigest: `sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9`

## 2. Backup

```
$ cp ~/.hermes/config.yaml ~/.hermes/config.yaml.bak-notionpin
$ sha256sum ~/.hermes/config.yaml ~/.hermes/config.yaml.bak-notionpin
78cb57e942851bb56d44fa23eec375e9d5cee322f1520ade20be1d2a30742c82  config.yaml
78cb57e942851bb56d44fa23eec375e9d5cee322f1520ade20be1d2a30742c82  config.yaml.bak-notionpin
```

## 3. Edit

Only PyYAML is installed on the VM (no ruamel.yaml), and a full
`yaml.safe_load` → `yaml.safe_dump` round-trip would reformat the file's
multi-line personality block scalars and quoting elsewhere — violating "change
ONLY that one value." Instead used a python script (`~/wf3_notion_pin.py`,
deleted after use) that:

1. Loads the file with `yaml.safe_load` to verify it parses and to confirm
   the notion stanza's last arg is `mcp/notion` and the slack stanza is
   intact (guard against acting on an unexpected file).
2. Locates the exact line `    - mcp/notion\n` in the raw text — confirmed
   unique beforehand (`grep -c mcp/notion config.yaml` → 1).
3. Replaces only that line with `    - mcp/notion@sha256:<digest>\n`.
4. Diffs old vs new line-by-line and asserts exactly one line index differs.
5. Re-parses the new text with `yaml.safe_load`, asserts the notion args
   changed only in the pinned image and the rest of the parsed structure is
   identical to the original (`new_doc == doc` after normalizing the pinned
   value back).
6. Writes to `config.yaml.tmp-wf3-notionpin` then `os.replace()`s onto
   `config.yaml` (atomic rename — no half-written file visible to Hermes'
   live-reload).

Wrapped in the lock:

```
$ flock ~/.hermes/.wf3.lock -c 'python3 ~/wf3_notion_pin.py'
OK: patched line 195
old: - mcp/notion
new: - mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9
```

**Side effect caught and fixed:** `os.replace()` from a tmp file created via
Python's `open(..., "w")` picks up the process umask, which dropped
`config.yaml`'s mode from `600` to `664` (all the `config.yaml.bak-*` files
on the VM are `600`). Restored under the same lock:

```
$ flock ~/.hermes/.wf3.lock -c 'chmod 600 ~/.hermes/config.yaml'
$ stat -c '%a %n' ~/.hermes/config.yaml
600 /home/gaetan/.hermes/config.yaml
```

## 4. Diff — only the one line changed

```
$ diff -u ~/.hermes/config.yaml.bak-notionpin ~/.hermes/config.yaml
--- config.yaml.bak-notionpin
+++ config.yaml
@@ -192,6 +192,6 @@
     - --rm
     - -e
     - NOTION_TOKEN
-    - mcp/notion
+    - mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9
     env:
       NOTION_TOKEN: ${NOTION_API_TOKEN}
```

Slack stanza (lines 175-186), the personality block scalars, and everything
else byte-identical to the backup — confirmed by the diff above showing only
the `@@ -192,6 +192,6 @@` hunk.

## 5. Verify

```
$ python3 -c "import yaml; d=yaml.safe_load(open('/home/gaetan/.hermes/config.yaml')); print('parsed OK'); print(d['mcp_servers']['notion']['args']); print(d['mcp_servers']['slack']['args'])"
parsed OK
['run', '-i', '--rm', '-e', 'NOTION_TOKEN', 'mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9']
['run', '-i', '--rm', '--env-file', '/home/gaetan/tars/slack-mcp/.env', 'ghcr.io/korotovsky/slack-mcp-server:v1.3.0', '--transport', 'stdio', '--no-cache']

$ time ~/.local/bin/hermes mcp test notion
  Testing 'notion'...
  Transport: stdio → docker
  Auth: none
  ✓ Connected (577ms)
  ✓ Tools discovered: 24
    API-get-user, API-get-users, API-get-self, API-post-search,
    API-get-block-children, API-patch-block-children, API-retrieve-a-block,
    API-update-a-block, API-delete-a-block, API-retrieve-a-page,
    API-patch-page, API-post-page, API-retrieve-a-page-property,
    API-retrieve-a-comment, API-create-a-comment, API-query-data-source,
    API-retrieve-a-data-source, API-update-a-data-source,
    API-create-a-data-source, API-list-data-source-templates,
    API-retrieve-a-database, API-move-page, API-retrieve-page-markdown,
    API-update-page-markdown
real  0m1.217s
```

Re-run after the permission fix, for the record:

```
$ ~/.local/bin/hermes mcp test notion
  Testing 'notion'...
  Transport: stdio → docker
  Auth: none
  ✓ Connected (595ms)
  ✓ Tools discovered: 24
```

24 tools, sub-second connect — matches expectation, no rollback needed.

## Old vs new image reference

- Old: `mcp/notion` (unpinned, resolves to `latest` at pull time)
- New: `mcp/notion@sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9`

## Scope respected

- Slack stanza: untouched (byte-identical, confirmed by diff + parsed-args
  check above).
- Gateway unit: not touched, stays disabled — no systemctl/service commands
  run this session.
- p-Hermes (192.168.0.3): not touched — this task only ever connected to
  192.168.0.9.
- No secrets decrypted, no `.env` values read (only key existence implied by
  `${NOTION_API_TOKEN}` already present in the untouched `env:` line).
- No `git add`/commit/push run.
- `status/lane-a.md` / `status/lane-b.md` not edited.

## Verdict

**PASS** — `mcp/notion` in `~/.hermes/config.yaml` on the Tars VM is now
digest-pinned to `sha256:df0d6781d03f37bd5b962c85ae1f288382f31b7108c489473641ffc372f43dc9`,
edited under `flock` via an atomic load-verify-replace, with only that one
line changed (confirmed by diff) and `hermes mcp test notion` reporting
Connected with 24 tools post-edit.
