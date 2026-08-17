#!/usr/bin/env python3
"""GCN-73 helpers for bounded Slack collection and atomic engagement state.

Stdlib only. Secrets stay inside the process; stdout is compact JSON.
"""
from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
from pathlib import Path
import re
import ssl
import stat
import sys
import tempfile
import time
import unittest
import urllib.error
import urllib.parse
import urllib.request
from zoneinfo import ZoneInfo

SLACK_ENDPOINT = "https://slack.com/api/"
SLACK_USER = "U08BDJAMSRZ"
TARS_USER = "U0BBH85NAKH"
MAX_RESPONSE_BYTES = 2 * 1024 * 1024
PAGE_SIZE = 100
MAX_PAGES_PER_VIEW_DATE = 25
MAX_EVENTS = 100
SNIPPET_CHARS = 240
ALLOWED_METHODS = {"auth.test", "search.messages"}
SECRET_ASSIGNMENT = re.compile(r"(?i)\b(authorization|api[_-]?key|access[_-]?token|token|secret|password)(\s*[:=]\s*)(?:bearer\s+)?(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)")
TOKEN_PATTERN = re.compile(r"\b(?:xox[baprcds]-|xoxe\.xoxp-|sk-|gh[pousr]_)[A-Za-z0-9._~+/=-]{8,}")


def instant(value: str) -> dt.datetime:
    parsed = dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    if parsed.tzinfo is None:
        raise ValueError("timezone_required")
    return parsed.astimezone(dt.timezone.utc)


def iso(value: dt.datetime) -> str:
    return value.astimezone(dt.timezone.utc).isoformat().replace("+00:00", "Z")


def compact(value) -> None:
    print(json.dumps(value, ensure_ascii=False, separators=(",", ":"), sort_keys=True))


def redact(text: str, credentials=()) -> str:
    out = text
    for credential in credentials:
        if credential:
            out = out.replace(credential, "[REDACTED]")
            out = out.replace(urllib.parse.quote(credential, safe=""), "[REDACTED]")
    out = SECRET_ASSIGNMENT.sub(lambda m: m.group(1) + m.group(2) + "[REDACTED]", out)
    return TOKEN_PATTERN.sub("[REDACTED]", out)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise urllib.error.HTTPError(req.full_url, code, "redirect_rejected", headers, fp)


def default_opener():
    context = ssl.create_default_context()
    return urllib.request.build_opener(urllib.request.HTTPSHandler(context=context), NoRedirect())


def load_credentials(path: Path) -> tuple[str, str]:
    mode = stat.S_IMODE(path.stat().st_mode)
    if mode & 0o077 or path.stat().st_uid != os.getuid():
        raise RuntimeError("credential_permissions")
    values = {}
    for line in path.read_text(encoding="utf-8").splitlines():
        stripped = line.strip()
        if stripped and not stripped.startswith("#") and "=" in stripped:
            key, value = stripped.split("=", 1)
            values[key.strip()] = value.strip().strip("\"'")
    xoxc = values.get("SLACK_MCP_XOXC_TOKEN", "")
    xoxd = values.get("SLACK_MCP_XOXD_TOKEN", "")
    if not xoxc.startswith("xoxc-") or not xoxd.startswith("xoxd-"):
        raise RuntimeError("credential_shape")
    return xoxc, xoxd


class SlackClient:
    def __init__(self, xoxc: str, xoxd: str, opener=None):
        self.xoxc, self.xoxd = xoxc, xoxd
        self.opener = opener or default_opener()

    def call(self, method: str, params: dict) -> dict:
        if method not in ALLOWED_METHODS:
            raise RuntimeError("method_not_allowed")
        endpoint = SLACK_ENDPOINT + method
        request = urllib.request.Request(
            endpoint + "?" + urllib.parse.urlencode(params),
            headers={"Authorization": "Bearer " + self.xoxc, "Cookie": "d=" + self.xoxd},
            method="GET",
        )
        try:
            with self.opener.open(request, timeout=20) as response:
                if response.geturl() != endpoint + "?" + urllib.parse.urlencode(params):
                    raise RuntimeError("endpoint_mismatch")
                raw = response.read(MAX_RESPONSE_BYTES + 1)
        except urllib.error.HTTPError as exc:
            if exc.code == 429:
                raise RuntimeError("rate_limited")
            raise RuntimeError("http_error")
        except (urllib.error.URLError, TimeoutError, OSError):
            raise RuntimeError("transport_error")
        if len(raw) > MAX_RESPONSE_BYTES:
            raise RuntimeError("response_too_large")
        try:
            payload = json.loads(raw.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError):
            raise RuntimeError("malformed_json")
        if not isinstance(payload, dict) or not payload.get("ok"):
            code = payload.get("error") if isinstance(payload, dict) else None
            if code in {"invalid_auth", "not_authed", "token_revoked", "account_inactive"}:
                raise RuntimeError("auth_failed")
            raise RuntimeError("api_error")
        return payload


def _channel_id(match: dict) -> str | None:
    if isinstance(match.get("channel_id"), str):
        return match["channel_id"]
    channel = match.get("channel")
    return channel.get("id") if isinstance(channel, dict) and isinstance(channel.get("id"), str) else None


def _safe_permalink(value: object) -> str | None:
    if not isinstance(value, str):
        return None
    parsed = urllib.parse.urlsplit(value)
    if parsed.scheme != "https" or parsed.hostname not in {"slack.com", "mobileclub-squad.slack.com"}:
        return None
    return urllib.parse.urlunsplit((parsed.scheme, parsed.netloc, parsed.path, "", ""))


def _event(match: dict, credentials) -> dict | None:
    ts = match.get("ts")
    channel = _channel_id(match)
    user = match.get("user_id") or match.get("user")
    if not isinstance(ts, str) or not isinstance(channel, str) or not isinstance(user, str):
        return None
    if match.get("bot_id") or user == TARS_USER:
        return None
    try:
        moment = dt.datetime.fromtimestamp(float(ts), tz=dt.timezone.utc)
    except (ValueError, TypeError, OverflowError):
        return None
    text = " ".join(str(match.get("text") or "").split())
    text = redact(text, credentials)[:SNIPPET_CHARS]
    thread_ts = match.get("thread_ts") if isinstance(match.get("thread_ts"), str) else ts
    return {
        "id": f"slack:{channel}:{ts}",
        "channel_id": channel,
        "thread_ts": thread_ts,
        "message_ts": ts,
        "timestamp": iso(moment),
        "author_id": user,
        "evidence_handle": _safe_permalink(match.get("permalink")),
        "snippet": text,
    }


def _search_view(client: SlackClient, query: str, credentials) -> tuple[list[dict], int]:
    rows, page, cursor, pages = [], 1, "", 0
    seen_positions = set()
    while True:
        if pages >= MAX_PAGES_PER_VIEW_DATE:
            raise RuntimeError("page_cap")
        params = {"query": query, "count": PAGE_SIZE, "sort": "timestamp", "sort_dir": "asc"}
        if cursor:
            params["cursor"] = cursor
        else:
            params["page"] = page
        payload = client.call("search.messages", params)
        messages = payload.get("messages")
        if not isinstance(messages, dict) or not isinstance(messages.get("matches"), list):
            raise RuntimeError("malformed_search")
        rows.extend(messages["matches"])
        pages += 1
        metadata = payload.get("response_metadata")
        next_cursor = metadata.get("next_cursor", "") if isinstance(metadata, dict) else ""
        paging = messages.get("paging")
        if next_cursor:
            position = ("cursor", next_cursor)
            if position in seen_positions:
                raise RuntimeError("repeated_cursor")
            seen_positions.add(position)
            cursor = next_cursor
            continue
        if isinstance(paging, dict):
            current = paging.get("page", page)
            total = paging.get("pages", current)
            if not isinstance(current, int) or not isinstance(total, int):
                raise RuntimeError("malformed_paging")
            if current < total:
                position = ("page", current + 1)
                if position in seen_positions:
                    raise RuntimeError("repeated_page")
                seen_positions.add(position)
                page = current + 1
                continue
        break
    events = []
    for row in rows:
        if isinstance(row, dict):
            event = _event(row, credentials)
            if event:
                events.append(event)
    return events, pages


def collect_slack(client: SlackClient, cursor: dt.datetime, run_end: dt.datetime, credentials=()) -> dict:
    if run_end < cursor:
        raise RuntimeError("run_end_before_cursor")
    paris = ZoneInfo("Europe/Paris")
    day = cursor.astimezone(paris).date()
    final_day = run_end.astimezone(paris).date()
    exact, pages = {}, 0
    while day <= final_day:
        date_text = day.isoformat()
        for modifier in (f"from:<@{SLACK_USER}>", f"with:<@{SLACK_USER}>"):
            batch, used = _search_view(client, f"{modifier} on:{date_text}", credentials)
            pages += used
            for event in batch:
                moment = instant(event["timestamp"])
                if cursor < moment <= run_end:
                    exact[event["id"]] = event
        day += dt.timedelta(days=1)
    ordered = sorted(exact.values(), key=lambda e: (e["timestamp"], e["id"]))
    retained = ordered[:MAX_EVENTS]
    if len(ordered) > MAX_EVENTS:
        edge = retained[-1]["timestamp"]
        retained.extend(event for event in ordered[MAX_EVENTS:] if event["timestamp"] == edge)
        proposed = edge
        truncated = True
    else:
        proposed = iso(run_end)
        truncated = False
    return {
        "coverage": {
            "source": "slack",
            "start": iso(cursor),
            "end": iso(run_end),
            "complete": True,
            "truncated": truncated,
            "incomplete": False,
            "fail_closed": False,
            "pages": pages,
            "exact_matches": len(ordered),
            "returned": len(retained),
            "proposed_cursor": proposed,
            "caps": {"events": MAX_EVENTS, "pages_per_view_date": MAX_PAGES_PER_VIEW_DATE},
        },
        "events": retained,
    }


def state_hash(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read_state(path: Path) -> dict:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict) or not isinstance(data.get("sources"), dict) or not isinstance(data.get("items"), dict):
        raise RuntimeError("state_shape")
    return data


def _set_path(data: dict, dotted: str, value) -> None:
    parts = dotted.split(".")
    if not parts or any(not part for part in parts):
        raise RuntimeError("mutation_path")
    node = data
    for part in parts[:-1]:
        child = node.get(part)
        if child is None:
            child = {}
            node[part] = child
        if not isinstance(child, dict):
            raise RuntimeError("mutation_path")
        node = child
    node[parts[-1]] = value


def atomic_write(path: Path, data: dict) -> str:
    path.parent.mkdir(parents=True, exist_ok=True)
    raw = (json.dumps(data, ensure_ascii=False, indent=2, sort_keys=False) + "\n").encode("utf-8")
    fd, tmp_name = tempfile.mkstemp(prefix=path.name + ".", suffix=".tmp", dir=path.parent)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "wb") as handle:
            handle.write(raw)
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(tmp_name, path)
        dir_fd = os.open(path.parent, os.O_RDONLY)
        try:
            os.fsync(dir_fd)
        finally:
            os.close(dir_fd)
    finally:
        if os.path.exists(tmp_name):
            os.unlink(tmp_name)
    reread = read_state(path)
    if reread != data:
        raise RuntimeError("state_verify")
    return hashlib.sha256(raw).hexdigest()


def atomic_write_ndjson(path: Path, rows: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    raw = b"".join((json.dumps(row, ensure_ascii=False, separators=(",", ":"), sort_keys=True) + "\n").encode("utf-8") for row in rows)
    fd, tmp_name = tempfile.mkstemp(prefix=path.name + ".", suffix=".tmp", dir=path.parent)
    try:
        os.fchmod(fd, 0o600)
        with os.fdopen(fd, "wb") as handle:
            handle.write(raw)
            handle.flush()
            os.fsync(handle.fileno())
        os.replace(tmp_name, path)
    finally:
        if os.path.exists(tmp_name):
            os.unlink(tmp_name)


def apply_mutations(path: Path, expected_hash: str, mutations: dict) -> dict:
    before = state_hash(path)
    if before != expected_hash:
        raise RuntimeError("state_changed")
    data = read_state(path)
    for dotted, value in mutations.get("set", {}).items():
        _set_path(data, dotted, value)
    for source, values in mutations.get("append_seen", {}).items():
        source_state = data["sources"].get(source)
        if not isinstance(source_state, dict) or not isinstance(values, list):
            raise RuntimeError("seen_shape")
        ring = source_state.setdefault("seen", [])
        for value in values:
            if isinstance(value, str) and value not in ring:
                ring.append(value)
        source_state["seen"] = ring[-500:]
    for key, value in mutations.get("upsert_items", {}).items():
        if not isinstance(value, dict) or value.get("id") != key:
            raise RuntimeError("item_shape")
        data["items"][key] = value
    for key in mutations.get("delete_items", []):
        data["items"].pop(key, None)
    recovered = []
    auth = mutations.get("auth", {})
    notices = data.setdefault("last_failure_notice", {}).setdefault("auth", {})
    for source, record in auth.items():
        ok = record.get("ok")
        at = record.get("at")
        if not isinstance(ok, bool) or not isinstance(at, str):
            raise RuntimeError("auth_shape")
        prior = notices.get(source)
        if ok:
            if isinstance(prior, dict) and prior.get("failed"):
                recovered.append(source)
            notices[source] = {"failed": False, "recovered_at": at}
        else:
            notices[source] = {"failed": True, "failed_at": at}
    after = atomic_write(path, data)
    return {"before_sha256": before, "after_sha256": after, "size": path.stat().st_size, "recovered_sources": recovered}


def cmd_slack(args):
    try:
        credential_path = Path(args.credentials).expanduser() if args.credentials else Path.home() / "tars/slack-mcp/.env"
        xoxc, xoxd = load_credentials(credential_path)
        client = SlackClient(xoxc, xoxd)
        if args.probe:
            client.call("auth.test", {})
            compact({"coverage": {"source": "slack", "authenticated": True, "complete": True}, "events": []})
            return 0
        result = collect_slack(client, instant(args.cursor), instant(args.run_end), (xoxc, xoxd))
        if args.output:
            rows = [{"coverage": result["coverage"]}] + [{"event": event} for event in result["events"]]
            atomic_write_ndjson(Path(args.output).expanduser(), rows)
            compact({"output": str(Path(args.output).expanduser()), "coverage": result["coverage"]})
        else:
            compact(result)
        return 0
    except Exception as exc:
        code = str(exc) if isinstance(exc, RuntimeError) else "unexpected"
        compact({"coverage": {"source": "slack", "complete": False, "incomplete": True, "fail_closed": True, "errors": [code]}, "events": []})
        return 1


def cmd_snapshot(args):
    path = Path(args.state).expanduser()
    data = read_state(path)
    compact({"sha256": state_hash(path), "size": path.stat().st_size, "items": len(data["items"]), "sources": {k: {"cursor": v.get("cursor"), "last_success": v.get("last_success"), "seen": len(v.get("seen", []))} for k, v in data["sources"].items()}})
    return 0


def cmd_export(args):
    path = Path(args.state).expanduser()
    data = read_state(path)
    header = {key: value for key, value in data.items() if key not in {"items", "sources"}}
    source_meta = {
        source: {key: value for key, value in record.items() if key != "seen"}
        for source, record in data["sources"].items()
    }
    header["sources"] = source_meta
    rows = [{"state": header, "sha256": state_hash(path), "size": path.stat().st_size}]
    seen_lines = 0
    for source, record in sorted(data["sources"].items()):
        seen = record.get("seen", [])
        for offset in range(0, len(seen), 100):
            rows.append({"seen": {"source": source, "ids": seen[offset:offset + 100]}})
            seen_lines += 1
    rows.extend({"item": value} for _, value in sorted(data["items"].items()))
    output = Path(args.output).expanduser()
    atomic_write_ndjson(output, rows)
    compact({"output": str(output), "sha256": state_hash(path), "items": len(data["items"]), "seen_lines": seen_lines, "lines": len(rows)})
    return 0


def cmd_apply(args):
    path = Path(args.state).expanduser()
    mutations = json.loads(Path(args.mutations).read_text(encoding="utf-8"))
    compact(apply_mutations(path, args.expected_hash, mutations))
    return 0


class _Response:
    def __init__(self, url, payload):
        self.url = url
        self.raw = json.dumps(payload).encode()
    def __enter__(self): return self
    def __exit__(self, *args): return False
    def geturl(self): return self.url
    def read(self, amount): return self.raw[:amount]


class _FixtureOpener:
    def __init__(self, responses): self.responses = list(responses)
    def open(self, request, timeout=20):
        if not self.responses: raise AssertionError("fixture_exhausted")
        return _Response(request.full_url, self.responses.pop(0))


def _matches(start, count):
    return [{"ts": f"{start+i}.000001", "channel_id": "C1", "user_id": SLACK_USER, "text": f"event {start+i}", "permalink": f"https://mobileclub-squad.slack.com/archives/C1/p{start+i}"} for i in range(count)]


class SelfTests(unittest.TestCase):
    def test_slack_drains_more_than_100(self):
        base = 1_786_700_000
        page1 = {"ok": True, "messages": {"matches": _matches(base, 60), "paging": {"page": 1, "pages": 2}}}
        page2 = {"ok": True, "messages": {"matches": _matches(base+60, 60), "paging": {"page": 2, "pages": 2}}}
        empty = {"ok": True, "messages": {"matches": [], "paging": {"page": 1, "pages": 1}}}
        client = SlackClient("xoxc-test", "xoxd-test", _FixtureOpener([page1, page2, empty]))
        result = collect_slack(client, dt.datetime.fromtimestamp(base-1, dt.timezone.utc), dt.datetime.fromtimestamp(base+200, dt.timezone.utc))
        self.assertEqual(result["coverage"]["exact_matches"], 120)
        self.assertEqual(result["coverage"]["returned"], 100)
        self.assertTrue(result["coverage"]["truncated"])
        self.assertEqual(result["coverage"]["proposed_cursor"], iso(dt.datetime.fromtimestamp(base+99.000001, dt.timezone.utc)))

    def test_large_state_atomic_and_auth_recovery(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "state.json"
            data = {"version": 1, "sources": {"slack": {"cursor": "a", "last_success": "a", "seen": []}, "email": {"cursor": "a", "last_success": "a", "seen": []}, "linear": {"cursor": "a", "last_success": "a", "seen": []}}, "items": {}, "future": "x" * 130_000}
            atomic_write(path, data)
            first = apply_mutations(path, state_hash(path), {"auth": {"email": {"ok": False, "at": "2026-08-17T10:00:00Z"}}, "append_seen": {"slack": [f"e{i}" for i in range(600)]}})
            self.assertGreater(first["size"], 100_000)
            self.assertEqual(read_state(path)["sources"]["email"]["cursor"], "a")
            second = apply_mutations(path, state_hash(path), {"auth": {"email": {"ok": True, "at": "2026-08-17T10:30:00Z"}}, "set": {"sources.email.cursor": "b", "sources.email.last_success": "b"}})
            self.assertEqual(second["recovered_sources"], ["email"])
            saved = read_state(path)
            self.assertEqual(saved["future"], data["future"])
            self.assertEqual(saved["sources"]["email"]["cursor"], "b")
            self.assertEqual(len(saved["sources"]["slack"]["seen"]), 500)
            self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)

    def test_hash_guard(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "state.json"
            atomic_write(path, {"sources": {}, "items": {}})
            with self.assertRaisesRegex(RuntimeError, "state_changed"):
                apply_mutations(path, "0" * 64, {"set": {}})


def cmd_self_test(_args):
    suite = unittest.defaultTestLoader.loadTestsFromTestCase(SelfTests)
    result = unittest.TextTestRunner(stream=sys.stderr, verbosity=1).run(suite)
    compact({"tests_run": result.testsRun, "failures": len(result.failures), "errors": len(result.errors), "ok": result.wasSuccessful()})
    return 0 if result.wasSuccessful() else 1


def parser():
    root = argparse.ArgumentParser()
    sub = root.add_subparsers(dest="command", required=True)
    slack = sub.add_parser("slack")
    slack.add_argument("--credentials")
    slack.add_argument("--cursor")
    slack.add_argument("--run-end")
    slack.add_argument("--output")
    slack.add_argument("--probe", action="store_true")
    slack.set_defaults(func=cmd_slack)
    snapshot = sub.add_parser("state-snapshot")
    snapshot.add_argument("--state", default="~/.hermes/state/engagement-checker.json")
    snapshot.set_defaults(func=cmd_snapshot)
    export = sub.add_parser("state-export")
    export.add_argument("--state", default="~/.hermes/state/engagement-checker.json")
    export.add_argument("--output", required=True)
    export.set_defaults(func=cmd_export)
    apply = sub.add_parser("state-apply")
    apply.add_argument("--state", default="~/.hermes/state/engagement-checker.json")
    apply.add_argument("--expected-hash", required=True)
    apply.add_argument("--mutations", required=True)
    apply.set_defaults(func=cmd_apply)
    tests = sub.add_parser("self-test")
    tests.set_defaults(func=cmd_self_test)
    return root


if __name__ == "__main__":
    args = parser().parse_args()
    if args.command == "slack" and not args.probe and (not args.cursor or not args.run_end):
        parser().error("slack collection requires --cursor and --run-end")
    raise SystemExit(args.func(args))
