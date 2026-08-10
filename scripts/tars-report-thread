#!/usr/bin/env python3
"""Point the three Tars report cron jobs at today's Paris reporting thread.

Runs as a no-agent `--script` cron job. It posts NOTHING to Slack: it reads
conversations.history, picks the day's parent (earliest top-level message
authored by the Tars bot), and rewrites each report job's `deliver` value via
`hermes cron edit`. Empty stdout is the scheduler's silence marker, so every
diagnostic goes to stderr and any failure exits nonzero (the scheduler then
alerts the reconciler job's own deliver target, the home DM).

Spec: docs/specs/daily-report-threading.md
"""

import argparse
import fcntl
import json
import os
import subprocess
import sys
import tempfile
import urllib.parse
import urllib.request
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo


class Fail(Exception):
    """A failure worth exiting nonzero for."""


def log(msg):
    print(msg, file=sys.stderr)


def parse_args(argv):
    p = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    p.add_argument("--channel", default="C0BP2GZUFSR")
    p.add_argument("--bot-user", default="U0BBH85NAKH")
    p.add_argument("--jobs", default="e231e5faf180,62e8cd9db637,759e08c598e3")
    p.add_argument("--hermes-home", default=os.path.expanduser("~/.hermes"))
    p.add_argument("--hermes-bin", default=os.path.expanduser("~/.local/bin/hermes"))
    p.add_argument("--slack-base", default="https://slack.com/api")
    p.add_argument("--tz", default="Europe/Paris")
    p.add_argument("--dry-run", action="store_true",
                   help="report intended actions on stderr, change nothing")
    return p.parse_args(argv)


def read_token(hermes_home):
    """Env first (future-proof for terminal.env_passthrough), then ~/.hermes/.env.

    The value is only ever returned, never logged, persisted or put on argv.
    """
    token = os.environ.get("SLACK_BOT_TOKEN")
    if token:
        return token
    env_path = hermes_home / ".env"
    try:
        lines = env_path.read_text().splitlines()
    except OSError as exc:
        raise Fail(f"cannot read {env_path}: {exc.strerror}") from None
    for line in lines:
        line = line.strip()
        if line.startswith("export "):
            line = line[len("export "):].strip()
        key, sep, value = line.partition("=")
        if not sep or key.strip() != "SLACK_BOT_TOKEN":
            continue
        value = value.strip()
        if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
            value = value[1:-1]
        if value:
            return value
    raise Fail(f"SLACK_BOT_TOKEN not in os.environ nor {env_path}")


def load_json(path):
    try:
        return json.loads(path.read_text())
    except (OSError, ValueError) as exc:
        raise Fail(f"cannot read {path}: {exc}") from None


def load_jobs(jobs_path):
    """jobs.json is a list of job dicts (tolerate {"jobs": [...]} / {id: job})."""
    data = load_json(jobs_path)
    if isinstance(data, dict):
        data = data.get("jobs", list(data.values()))
    return {j["id"]: j for j in data if isinstance(j, dict) and "id" in j}


def find_parent(messages, bot_user):
    """Earliest top-level, subtype-free message authored by the bot, or None."""
    tops = [
        m for m in messages
        if m.get("user") == bot_user
        and not m.get("subtype")
        and m.get("ts")
        and (m.get("thread_ts") or m["ts"]) == m["ts"]
    ]
    return min((m["ts"] for m in tops), key=float, default=None)


def slack_history(base, token, channel, oldest):
    query = urllib.parse.urlencode({
        "channel": channel,
        "oldest": f"{oldest:.6f}",
        "limit": "999",
        "inclusive": "true",
    })
    req = urllib.request.Request(
        f"{base.rstrip('/')}/conversations.history?{query}",
        headers={"Authorization": f"Bearer {token}"},
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            payload = json.loads(resp.read().decode("utf-8"))
    except Exception as exc:  # never let a header-carrying frame escape
        raise Fail(f"conversations.history request failed: {type(exc).__name__}: {exc}") from None
    if not payload.get("ok"):
        raise Fail(f"conversations.history returned ok=false: {payload.get('error')}")
    return payload.get("messages", [])


def write_state(state_path, payload):
    fd, tmp = tempfile.mkstemp(dir=str(state_path.parent), prefix=".daily-report-thread.")
    try:
        with os.fdopen(fd, "w") as fh:
            json.dump(payload, fh, indent=2, sort_keys=True)
            fh.write("\n")
        os.replace(tmp, state_path)
    except BaseException:
        os.unlink(tmp)
        raise


def main(argv=None):
    args = parse_args(argv)
    home = Path(args.hermes_home)
    job_ids = [j.strip() for j in args.jobs.split(",") if j.strip()]
    state_dir = home / "state"
    state_dir.mkdir(parents=True, exist_ok=True)

    lock_fd = os.open(state_dir / "daily-report-thread.lock", os.O_CREAT | os.O_RDWR, 0o600)
    try:
        fcntl.flock(lock_fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except OSError:
        log("lock held by another reconciler, skipping tick")
        return 0

    now = datetime.now(ZoneInfo(args.tz))
    today = now.date().isoformat()
    state_path = state_dir / "daily-report-thread.json"
    try:
        state = json.loads(state_path.read_text())
    except (OSError, ValueError):
        state = {}  # state is a Slack-derived cache: missing or corrupt self-heals via discovery
    if not isinstance(state, dict):
        state = {}
    jobs = load_jobs(home / "cron" / "jobs.json")

    missing = [j for j in job_ids if j not in jobs]
    if missing:
        raise Fail(f"job id(s) not in jobs.json: {', '.join(missing)}")

    if state.get("date") == today and state.get("parent_ts"):
        target = f"slack:{args.channel}:{state['parent_ts']}"
        if all(jobs[j].get("deliver") == target for j in job_ids):
            log(f"fast path: {today} jobs already on {target}")
            return 0

    oldest = now.replace(hour=0, minute=0, second=0, microsecond=0).timestamp()
    messages = slack_history(args.slack_base, read_token(home), args.channel, oldest)
    parent_ts = find_parent(messages, args.bot_user)
    desired = f"slack:{args.channel}:{parent_ts}" if parent_ts else f"slack:{args.channel}"
    log(f"{today}: {len(messages)} message(s) since Paris midnight, "
        f"parent={parent_ts or 'none yet'}")

    failures = []
    for job_id in job_ids:
        current = jobs[job_id].get("deliver")
        if current == desired:
            continue
        if args.dry_run:
            log(f"dry-run: would set {job_id} deliver {current} -> {desired}")
            continue
        done = subprocess.run(
            [args.hermes_bin, "cron", "edit", job_id, "--deliver", desired],
            capture_output=True, text=True,
        )
        if done.returncode != 0:
            failures.append(f"{job_id}: cron edit rc={done.returncode} {done.stderr.strip()}")
            continue
        log(f"{job_id}: deliver {current} -> {desired}")

    if failures:
        raise Fail("; ".join(failures))

    if args.dry_run:
        log("dry-run: state not written")
        return 0

    write_state(state_path, {
        "channel": args.channel,
        "date": today,
        "parent_ts": parent_ts,
        "updated_at": now.isoformat(),
    })
    log(f"state: date={today} parent_ts={parent_ts}")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Fail as err:
        log(f"tars-report-thread: {err}")
        sys.exit(1)
