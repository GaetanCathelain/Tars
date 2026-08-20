#!/usr/bin/env python3
"""Atomically apply a bounded Gmail-triage state patch from raw JSON."""
import argparse
import json
import os
import stat
import tempfile
from pathlib import Path

MUTABLE_KEYS = {
    "cursor",
    "last_completed_run",
    "last_run",
    "consecutive_failures",
    "last_failure_notice",
    "unrouted_senders",
}


def atomic_json_write(path: Path, value: dict) -> None:
    payload = json.dumps(value, ensure_ascii=False, indent=2) + "\n"
    mode = stat.S_IMODE(path.stat().st_mode) if path.exists() else 0o600
    fd, temp_name = tempfile.mkstemp(prefix=f".{path.name}.", suffix=".tmp", dir=path.parent)
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as handle:
            handle.write(payload)
            handle.flush()
            os.fsync(handle.fileno())
        os.chmod(temp_name, mode)
        os.replace(temp_name, path)
        dir_fd = os.open(path.parent, os.O_RDONLY)
        try:
            os.fsync(dir_fd)
        finally:
            os.close(dir_fd)
    finally:
        if os.path.exists(temp_name):
            os.unlink(temp_name)


def main() -> None:
    default_home = Path(os.environ.get("HERMES_HOME", Path.home() / ".hermes"))
    parser = argparse.ArgumentParser()
    parser.add_argument("--state", type=Path, default=default_home / "state/gmail-triage.json")
    parser.add_argument("--patch-file", type=Path, required=True)
    args = parser.parse_args()

    lock = args.state.with_name("gmail-triage.lock")
    if not lock.is_dir():
        raise SystemExit(f"gmail-triage lock is not held: {lock}")

    with args.state.open("r", encoding="utf-8") as handle:
        state = json.load(handle)
    with args.patch_file.open("r", encoding="utf-8") as handle:
        patch = json.load(handle)
    if not isinstance(state, dict) or not isinstance(patch, dict):
        raise SystemExit("state and patch must be JSON objects")

    unexpected = set(patch) - MUTABLE_KEYS - {"seen_append"}
    if unexpected:
        raise SystemExit(f"unsupported patch keys: {sorted(unexpected)}")

    seen_append = patch.pop("seen_append", [])
    if not isinstance(seen_append, list) or not all(isinstance(item, str) for item in seen_append):
        raise SystemExit("seen_append must be a JSON array of message-id strings")
    seen = state.get("seen", [])
    if not isinstance(seen, list) or not all(isinstance(item, str) for item in seen):
        raise SystemExit("state.seen must be a JSON array of strings")

    # Keep the most recent occurrence of every existing id, then append this run's ids.
    seen = list(reversed(list(dict.fromkeys(reversed(seen)))))
    for message_id in seen_append:
        seen = [item for item in seen if item != message_id]
        seen.append(message_id)
    state["seen"] = seen[-500:]
    state.update(patch)

    atomic_json_write(args.state, state)
    with args.state.open("r", encoding="utf-8") as handle:
        verified = json.load(handle)
    if verified != state:
        raise SystemExit("post-write verification mismatch")
    print(json.dumps({
        "persisted": True,
        "cursor": verified.get("cursor"),
        "seen_count": len(verified.get("seen", [])),
        "appended_seen_count": len(seen_append),
    }, ensure_ascii=False))


if __name__ == "__main__":
    main()
