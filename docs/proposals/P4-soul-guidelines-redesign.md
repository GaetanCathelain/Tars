# P4 — SOUL / guidelines redesign: separation of concerns, not confinement

**Status:** proposed, NOT applied · **Source:** orca-a2a peer, 2026-08-07 ·
**Full text:** `artifacts/SOUL-proposed.md` (56 lines, complete replacement for
`~/.hermes/SOUL.md`) · **Change list:** `artifacts/guidelines-changelist.md`
(12 changes, A–F, each with current/proposed text + risk both ways) ·
**Reviews:** `artifacts/safety-review-{deletions,timidity}.md` ·
**Decide:** Gaetan · **Must land together with [P3](P3-orca-v2-skill.md)** — see Order.

## Why

Your ruling: Tars is a *secretary of work* — an orchestrator that doesn't do the
work. The old rules enforced that by making Tars **unable** to act (no sudo, deny
lists, a sandboxed wrapper). Under a separation-of-concerns frame that machinery
doesn't earn its keep: Tars may hold your access on cooper because *you* hold it.

12 changes: 6 reframes · 4 adds · 1 drop · 1 stale rewrite, across `SOUL.md`,
`config.yaml`, one skill, and repo docs.

## The four rules worth reading closely

**Rule 2 — was the real blocker.** It read "I never open, review or merge a pull
request, and I never push" — which forbids *reading* a PR, diff or CI log, i.e.
exactly the verification the redesign requires, under a heading that overrides
everything else. Now:

> 2. I never merge, approve or push. Reading a pull request, a diff or a CI log is
>    how I verify what I delegated — that is my job, not a breach of it.

**Rule 1 — closed the documentation loophole.** It enumerated file types ("no
implementation, patch, script, config"), so asked for *documentation* Tars would
write it itself and believe it was compliant. Now scoped by **role**, with
"it's only markdown" explicitly refused, and two carve-outs: writing the brief
isn't doing the work, and quoting an agent's output back to you is evidence owed,
not a breach.

**Rule 7 — grants sudo positively** and names the timidity failure mode as a
failure: *"refusing to act is as much a failure as doing the implementation
myself."* p-Hermes is given its address (192.168.0.8, read-only) so it can't be
conflated with the pve hypervisor.

**Rule 8 — NEW, and the one line worth arguing about:**

> 8. Before an action that cannot be undone, I say what I am about to do and leave
>    Gaetan the beat to stop it. This is not asking permission — on cooper I act
>    with his access — it is not surprising him with something he cannot reverse.

Announce-then-act: a **statement, not a request**. It names no forbidden command,
no path list, no capability Tars lacks — those three properties are what keep it
prudence rather than confinement. My review: the wording holds. If it ever drifts
to "I ask before…", it becomes the thing you deleted. **You dropped confinement,
not prudence** — this is the line that encodes the difference.

Rule 4 (the identity frame + `·`) is byte-identical to live — verified
programmatically. No `docs/facts.md` change needed.

## B1 — the approvals config: PARKED, your call

**This is the real sudo gate. No prose sentence ever was** — dropping the
doc wording alone changes nothing.

Today there is no `approvals:` key, so the code default `manual` applies, and
`DANGEROUS_PATTERNS` matches `sudo` including flags. Proposed:

```yaml
approvals:
  mode: off   # Tars acts with Gaetan's own access on cooper by design (WF5 redesign,
              # 2026-08-07). Separation of concerns, not confinement, is the guardrail.
```

**The nuance that makes `off` less drastic than it sounds** — Hermes has two
tiers, and the peer read the source directly:

| Tier | Contents | Under `off` |
|---|---|---|
| **Hardline** (source: *"deliberately tiny — only things with no recovery path"*) | `rm -rf /`, recursive delete of system/home dirs, `mkfs`, `dd` to a raw device, redirect to a block device, fork bomb, shutdown/reboot | **Still blocked. Cannot be bypassed.** |
| **Dangerous patterns** (source: *"that's what yolo is for"*) | `git reset --hard`, `rm -rf /tmp/x`, `chmod -R 777`, `curl\|sh`, writes to `/etc`, `~/.ssh`, hermes `.env`/`config.yaml`, credential files | Released |

**Four options, not two:** `manual` (current — but on Slack it's block-and-hang,
there's no interactive human to approve), `smart` (an auxiliary LLM judges each
dangerous command, with a consecutive-denial circuit breaker), `off`, or **route
approvals into Slack** — the programmatic hooks already exist
(`resolve_gateway_approval` / `has_blocking_approval` / `submit_pending`), so
that's wiring-up rather than building-from-scratch.

Peer recommends `off` (unrecoverable tier stays blocked; Tars's own command
surface is narrow — ssh, orca, hermes — and the code-touching happens inside a
spawned Claude Code session with its own permission system). I'd add: `smart`
or Slack-routed approvals are the options worth a second thought before `off`,
precisely because rule 8 is a *prompt-level* promise and `off` removes the
*code-level* backstop underneath it.

## Order · restart · verify · rollback

- **Order: P4 (SOUL) and P3 (skill) must land together.** The law review caught
  it: the sudo prohibition you ordered dropped is **still live right now** at
  `~/.hermes/skills/delegate-to-cooper/SKILL.md:47`. SOUL alone = rule 7 granting
  sudo while a loaded skill forbids it, contradicting itself in the same context
  window. B1 lands separately, only on your ruling. Repo docs D1–D5 anytime.
- **Restart: none needed** (verified in source) — `SOUL.md` is read fresh at every
  system-prompt build; `config.yaml` is mtime-cached and re-read per turn (~30 s).
  Do **not** restart while the other peer is live on the gateway.
- **Verify:** one Slack turn confirms Tars still answers you (identity frame
  intact) **and** that a non-Gaetan sender still trips the early-reject WARNING —
  grep the log, never settle for absence-of-reply.
- **Rollback:** `.bak` restore under `flock`, no restart either way. Note the v1
  `SKILL.md` exists **only on the VM, never in git** — its `.bak` is the only
  rollback path, so take it before touching anything.

## Risk to state plainly

With confinement dropped, **`SLACK_ALLOWED_USERS` + the adapter early-reject
becomes the sole gate** on transitive Orca access — promoted from
defence-in-depth. And P1 (thread-follow) widens *what wakes* Tars while this
widens *what a wake can do*: together the blast radius is the **product**, not
the sum. Recommend not landing both unreviewed on the same day.

## Decision

- [ ] Apply SOUL replacement + P3 skill together
- [ ] Apply with edits (mark them on `artifacts/SOUL-proposed.md`)
- [ ] Not yet
- B1 approvals mode: [ ] `off` · [ ] `smart` · [ ] keep `manual` · [ ] route approvals into Slack first
