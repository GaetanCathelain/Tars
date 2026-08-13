# Local patches on the VM's hermes-agent checkout

`~/.hermes/hermes-agent` on the Tars VM is a clone of upstream hermes-agent
carrying two local commits, not upstream. **Standing rule: both must be
re-applied (`git am --3way`) after any `git pull` / `/upgrade` of
hermes-agent.** Previously this was only a warning in `status/lane-a.md`
(the dated `d615ca8` entry — lane-a is newest-first, line numbers shift);
these files are the tracked, re-appliable form.

## Patches (apply in order — 0002 depends on 0001's tree)

| # | File | Commit (base `6e87d43`) | What it does |
|---|---|---|---|
| 1 | `0001-tool-progress-dm.patch` | `d615ca8` | DM-scoped `display.platforms.<platform>.tool_progress_dm` override. **Config no longer sets its key** (2026-08-10): the adapter collapses MPIMs into chat_type `dm`, so the chat-type override would narrate group DMs — 0002's chat-ID map subsumes it. Kept for the code path and upstream compat |
| 2 | `0002-tool-progress-conversations.patch` | `ef35fd1` (supersedes `0427536` — same production hunk byte-for-byte, +5 tests from the pre-PR adversarial review: MPIM-quiet, DM-via-map, fail-open ×2, bare-`off`/False pin, and a `_get_platform_tools` signature pin in the guard) | Per-conversation `display.platforms.<platform>.tool_progress_conversations` (chat_id → mode map); wins over `tool_progress_dm` and the platform setting. **Cosmetic only** — toolsets are identical in every conversation. Ships `tests/gateway/test_slack_tool_visibility.py` (25 tests). The test fixture's `HOME_CHANNEL = "C0BP2GZUFSR"` is deliberately KEPT despite the channel's 2026-08-13 retirement: it is test-only (never imported by the gateway; live narration routing comes from config `tool_progress_conversations`), and pointing it at the DM would collide with `OWNER_DM = "D0BBYNM01BL"` in the same file, collapsing the group-vs-DM discrimination the suite exists to test |

The ordering claim is verified, not inherited: 0002's `gateway/run.py` hunk
sits directly below 0001's `tool_progress_dm` block and quotes it as context,
and its pre-image blob (`index d7ecd22e7..`) is 0001's post-image
(`index 83342c3..d7ecd22`). 0002 does not apply to a bare `6e87d43`.

## WITHDRAWN — the previous 0002 (per-conversation *toolset* gating)

`0002-per-conversation-toolset-gating.patch` (VM lineage `94a3414` →
`302ed2e` → deployed as `5af9e1e`) is **withdrawn and deleted from this
directory. Never re-apply it.** It added `platform_tool_conversations.<platform>`
and returned an empty toolset for any conversation not on the list — a
*capability* restriction, where Gaetan had asked only that tool-call
*display* be limited to the home channel and his DM. The cost was real: a
fresh mention in `GQ07CQXT7` got a toolless session that could not delegate.

**The corrected deploy ran 2026-08-10 20:00 UTC**: the VM checkout was
reset `5af9e1e` → `d615ca8` and the new 0002 applied as VM commit `4d10183`
(tree ≡ kit `ef35fd1`). This directory and the VM checkout now match.
Evidence: `status/probes/wf5/slack-tool-visibility-correction.md` (the
correction), `status/probes/wf6/slack-tool-visibility-restore.md` (the
deploy).

## Apply + verify — the post-`/upgrade` re-apply path

**This block is for after a `git pull` / `/upgrade` that dropped BOTH
patches — it is NOT today's deploy** (which resets `5af9e1e` → `d615ca8`
and applies 0002 only; see `docs/plans/apply-slack-channel-boundaries.md`).
`git am 0001` fails as already-applied while `d615ca8` is in the history.

```bash
scp patches/000*.patch gaetan@192.168.0.9:/tmp/
ssh gaetan@192.168.0.9 '
  cd ~/.hermes/hermes-agent &&
  git -c user.name="Claude (Tars deploy)" -c user.email="noreply@anthropic.com" \
    am --3way /tmp/0001-tool-progress-dm.patch &&
  git -c user.name="Claude (Tars deploy)" -c user.email="noreply@anthropic.com" \
    am --3way /tmp/0002-tool-progress-conversations.patch &&
  git log --oneline -3
'
# Expect HEAD = ef35fd1's subject ("feat(display): per-conversation
# tool_progress override …"), then d615ca8's, then the upstream base.
# The replayed shas differ from ef35fd1 / d615ca8 — git am re-commits with the
# local committer; the trees are identical (verified for 0002 by a pristine
# `git am` onto bare d615ca8, correction probe §Verification evidence).
# If git am fails: git am --abort, then git apply --3way + manual commit.
```
