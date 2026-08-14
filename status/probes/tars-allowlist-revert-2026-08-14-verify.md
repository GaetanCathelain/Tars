# Independent verification — allowlist revert (2026-08-14)

Fresh, read-only re-measurement of the claims in
`status/probes/tars-allowlist-revert-2026-08-14.md`. All commands run live
against **192.168.0.9** at ~13:05 UTC (restart was 13:02:18 UTC, ~3 min prior).
Nothing decrypted, no whole-file `sops -d`, no secret value printed.

**Overall: 6/7 checklist items PASS. Item 7 partially verified — the specific
negative-test leg (Oli/Ilo actually messaging post-revert) has not happened
yet; noted as unverified, same as the original log flagged.**

---

## 1. `SLACK_ALLOWED_USERS` — PASS

```
$ grep -n "^SLACK_ALLOWED_USERS" ~/.hermes/.env
8:SLACK_ALLOWED_USERS=U08BDJAMSRZ
```
Matches expected (Gaetan only, matches the Aug 7 pre-widen backups cited in
the revert log).

## 2. `allowed_channels` in config.yaml — PASS

```
$ grep -n allowed_channels ~/.hermes/config.yaml
(no match, exit=1)
```
Key absent, as claimed.

## 3. File permissions — PASS

```
$ stat -c '%a %n' ~/.hermes/.env ~/.hermes/config.yaml
600 /home/gaetan/.hermes/.env
600 /home/gaetan/.hermes/config.yaml
```
Both 600.

## 4. Timestamped `.bak` files from today — PASS

```
$ ls -la --time-style=full-iso ~/.hermes/.env.bak-20260814-revert ~/.hermes/config.yaml.bak-20260814-revert
-rw------- 1 gaetan gaetan 8115 2026-08-14 11:31:01 config.yaml.bak-20260814-revert
-rw------- 1 gaetan gaetan  908 2026-08-14 11:30:27 .env.bak-20260814-revert
```
Both present, 600, mtimes preserved from the pre-revert (widened) state per
`cp -p` — consistent with the claimed edit method.

## 5. Unit active / restart timing / clean logs — PASS

```
$ systemctl --user show hermes-gateway -p ActiveEnterTimestamp -p MainPID -p ActiveState
MainPID=1807357
ActiveState=active
ActiveEnterTimestamp=Fri 2026-08-14 13:02:18 UTC
$ date -u
Fri Aug 14 01:05:08 PM UTC 2026
```
Active, restarted ~3 min before this check — consistent with "in the last
hour." Last 50 `gateway.log` lines: clean shutdown (SIGTERM →
drain-timeout-with-1-cron-job-interrupted → disconnected) followed by clean
start (`Secret redaction: ENABLED` → Slack authenticated → socket connected →
a2a connected → `Gateway running with 2 platform(s)` → channel directory 107
targets → housekeeping/kanban started). **No startup errors.** No
`allowed_channels` reference anywhere in the tail (config load reflects its
absence — Hermes doesn't log the key at all when unset, which is the expected
silence, not a gap).

## 6. Sibling `mcp_servers` stanza intact — PASS

```
current mcp_servers keys: ['linear', 'notion', 'slack']
bak13   mcp_servers keys: ['linear', 'notion', 'slack']
current count: 3   bak13 count: 3
```
Identical to the newest pre-revert backup (`config.yaml.bak-2026-08-13-lcm-removal`,
15:15 Aug 13) — no sibling key dropped, YAML parses on both.

## 7. Post-restart inbound events from non-Gaetan senders — PARTIAL / one leg unverified

Non-Gaetan inbound Slack events **did** occur after the 13:02:18 restart:

```
13:04:38  Early reject of unauthorized user U05LKHLDV0A in channel GQ07CQXT7
13:04:44  Early reject of unauthorized user U05LKHLDV0A in channel GQ07CQXT7
13:05:04  Early reject of unauthorized user U0352LDJ7PD in channel C0BFQ5WFYTB
13:05:18  Early reject of unauthorized user U05LKHLDV0A in channel GQ07CQXT7
13:05:33  Early reject of unauthorized user U0352LDJ7PD in channel C0BFQ5WFYTB
```

Checked: for both `U05LKHLDV0A` and `U0352LDJ7PD`, **every** occurrence
anywhere in `gateway.log` (not just post-restart) is an `Early reject` WARNING
— there is no `inbound message:` or `response ready:` line for either id at
any timestamp, and no `Sending response` to either of their channels at these
times. `errors.log` since the restart shows only pre-existing/benign noise
(a cron tool_executor error from the interrupted job, capability-check
warnings, the known `client`/`token` bolt warning) — nothing tied to these
rejections. **Confirmed: early-rejected before any model call, no outbound
reply.**

However — these two ids were never part of the widened allowlist either (not
`U7XJ4K631`/Oli, not `U0BQTJUK2F2`/Ilo); they're unrelated randoms who've
always been rejected. Grepped explicitly for the two users the revert actually
removed:

```
$ awk '/^2026-08-14 13:02:18/{f=1} f' gateway.log | grep -E "U7XJ4K631|U0BQTJUK2F2"
(no match)
```

**No message from Oli or Ilo has arrived since the revert.** The specific
regression test — "a previously-widened user now gets Early-rejected" — stays
unverified, exactly as the original log's §7 flagged. This item is not a
failure of the revert; it's an absence of a test signal. Closes when either
of them next `@Tars`s.

---

## Redaction note

No `.env` or `config.yaml` value beyond `SLACK_ALLOWED_USERS` (Slack user IDs,
not secrets) was printed. No `sops` invoked, no whole-file dump, no other key
name/value pair shown beyond the `mcp_servers` key list already public in the
original log.
