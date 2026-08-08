#!/usr/bin/env python3
"""Bounded, read-only Slack Web API delta and selected-thread collector."""
from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import re
import ssl
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from collections.abc import Mapping
from dataclasses import dataclass
from decimal import Decimal, InvalidOperation
from pathlib import Path
from typing import Any, Callable

ENV_FILE = Path("/home/gaetan/tars/slack-mcp/.env")
API_BASE = "https://slack.com/api/"
DEFAULT_USER = "U08BDJAMSRZ"
MAX_PAGES_PER_VIEW = 20
MAX_MATCHES_SCANNED = 2000
MAX_EVENTS = 100
MAX_SNIPPET_CHARS = 240
MAX_WINDOW_DAYS = 31
MAX_CONTEXT_MESSAGES = 12
MAX_CONTEXT_PAGES = 20
MAX_CONTEXT_SCANNED = 2000
HTTP_TIMEOUT_SECONDS = 15
MAX_HTTP_ATTEMPTS = 3
MAX_TOTAL_RETRY_WAIT_SECONDS = 30.0
READ_ONLY_METHODS = frozenset({"auth.test", "search.messages", "conversations.replies"})
TOKEN_RE = re.compile(r"(?i)\b(?:xox[a-z]-)[A-Za-z0-9%._-]+")
SENSITIVE_ASSIGNMENT_RE = re.compile(
    r"(?i)\b(authorization|cookie|token|secret|password)\s*[:=]\s*\S+"
)
ID_RE = re.compile(r"^[CGD][A-Z0-9]{8,20}$")
USER_RE = re.compile(r"^U[A-Z0-9]{8,20}$")
SLACK_TS_RE = re.compile(r"^\d{10,16}(?:\.\d{1,9})?$")

# These patterns are deliberately conservative: discovery is broad, retention is local.
AUTHORED_COMMITMENT_RE = re.compile(
    r"\b(?:i\s+will|i\s*['’]\s*ll|i(?:\s+can)?\s+(?:take|handle|do|own)\b|"
    r"let\s+me\b|je\s+vais\b|je\s+m\s*['’]\s*(?:en\s+)?(?:occupe|charge)\b|"
    r"je\s+(?:prends|ferai|vais\s+faire)\b)", re.I
)
AUTHORED_RESOLUTION_RE = re.compile(
    r"\b(?:done|fixed|resolved|completed|shipped|closed|termin(?:é|e|ée|és|ées)|"
    r"fait|faite|régl(?:é|e|ée|és|ées)|résolu(?:e|s|es)?|corrig(?:é|e|ée|és|ées))\b", re.I
)
AUTHORED_DEFERRAL_RE = re.compile(
    r"\b(?:tomorrow|later|next\s+week|defer(?:red|ring)?|postpone(?:d)?|"
    r"demain|plus\s+tard|semaine\s+prochaine|report(?:é|e|ée|és|ées))\b", re.I
)
INBOUND_ASK_RE = re.compile(
    r"\b(?:can|could|would)\s+you\b|\bplease\b|\b(?:peux|pourrais)(?:-tu|\s+tu)\b|"
    r"\btu\s+(?:peux|pourrais)\b|\bstp\b|\bmerci\s+de\b", re.I
)
REVIEW_DECISION_RE = re.compile(
    r"\b(?:review|approve|approval|validation|decision|décision|avis|relire|revue|valider|approbation)\w*\b",
    re.I,
)
BLOCKER_RE = re.compile(
    r"\b(?:blocked|blocker|blocking|incident|security|customer|bloqu(?:é|e|ée|és|ées|ant|er)|"
    r"sécurité|client)\w*\b", re.I
)


class SafeError(Exception):
    def __init__(self, code: str, status: int | None = None):
        super().__init__(code)
        self.code = code
        self.status = status


class ParseFailure(Exception):
    pass


class JsonArgumentParser(argparse.ArgumentParser):
    def error(self, message: str) -> None:
        raise ParseFailure("invalid_arguments")


@dataclass(frozen=True)
class Credentials:
    xoxc: str
    xoxd: str


def load_credentials(path: Path = ENV_FILE) -> Credentials:
    try:
        stat = path.stat()
        if stat.st_mode & 0o077:
            raise SafeError("credential_file_permissions")
        values: dict[str, str] = {}
        with path.open("r", encoding="utf-8") as handle:
            for raw in handle:
                line = raw.strip()
                if not line or line.startswith("#") or "=" not in line:
                    continue
                key, value = line.split("=", 1)
                key, value = key.strip(), value.strip()
                if len(value) >= 2 and value[0] == value[-1] and value[0] in "\"'":
                    value = value[1:-1]
                if key in {"SLACK_MCP_XOXC_TOKEN", "SLACK_MCP_XOXD_TOKEN"}:
                    values[key] = value
    except SafeError:
        raise
    except OSError:
        raise SafeError("credential_file_unavailable") from None
    xoxc = values.get("SLACK_MCP_XOXC_TOKEN", "")
    xoxd = values.get("SLACK_MCP_XOXD_TOKEN", "")
    if not xoxc or not xoxd:
        raise SafeError("credentials_missing")
    if not xoxc.startswith("xoxc-") or not xoxd.startswith("xoxd-"):
        raise SafeError("credentials_invalid")
    return Credentials(xoxc=xoxc, xoxd=xoxd)


class NoRedirectHandler(urllib.request.HTTPRedirectHandler):
    """Refuse all redirects: following one could forward Authorization/Cookie
    headers to another origin or downgrade to plain HTTP."""

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


class SlackClient:
    def __init__(self, credentials: Credentials, opener: Callable[..., Any] | None = None,
                 sleeper: Callable[[float], None] | None = None) -> None:
        self._credentials = credentials
        self._sleeper = sleeper or time.sleep
        self._ssl_context = ssl.create_default_context()
        if opener is None:
            director = urllib.request.build_opener(
                NoRedirectHandler(), urllib.request.HTTPSHandler(context=self._ssl_context))

            def opener(request, timeout, context=None):
                return director.open(request, timeout=timeout)
        self._opener = opener

    @staticmethod
    def _retry_after_seconds(exc: urllib.error.HTTPError) -> float | None:
        value = exc.headers.get("Retry-After") if exc.headers is not None else None
        if not isinstance(value, str):
            return None
        value = value.strip()
        if not re.fullmatch(r"\d+", value):
            return None
        try:
            return float(value)
        except (ValueError, OverflowError):
            return None

    def get(self, method: str, params: Mapping[str, Any] | None = None) -> dict[str, Any]:
        if method not in READ_ONLY_METHODS:
            raise SafeError("endpoint_not_allowed")
        query = urllib.parse.urlencode(params or {})
        request = urllib.request.Request(
            API_BASE + method + (("?" + query) if query else ""), method="GET",
            headers={
                "Authorization": "Bearer " + self._credentials.xoxc,
                "Cookie": "d=" + self._credentials.xoxd,
                "Accept": "application/json",
                "User-Agent": "engagement-checker-slack-delta/1.1",
            },
        )
        total_wait = 0.0
        for attempt in range(1, MAX_HTTP_ATTEMPTS + 1):
            try:
                response = self._opener(request, timeout=HTTP_TIMEOUT_SECONDS, context=self._ssl_context)
                with response:
                    raw = response.read(2_000_001)
                if len(raw) > 2_000_000:
                    raise SafeError("response_too_large")
                payload = json.loads(raw.decode("utf-8"))
                break
            except urllib.error.HTTPError as exc:
                status = int(exc.code)
                retry_after = self._retry_after_seconds(exc) if status == 429 else None
                can_retry = (
                    retry_after is not None
                    and attempt < MAX_HTTP_ATTEMPTS
                    and total_wait + retry_after <= MAX_TOTAL_RETRY_WAIT_SECONDS
                )
                if not can_retry:
                    raise SafeError("http_error", status) from None
                total_wait += retry_after
                self._sleeper(retry_after)
            except (urllib.error.URLError, TimeoutError, OSError):
                raise SafeError("network_error") from None
            except (UnicodeDecodeError, json.JSONDecodeError, TypeError, ValueError):
                raise SafeError("invalid_response") from None
        if not isinstance(payload, dict):
            raise SafeError("invalid_response")
        if payload.get("ok") is not True:
            api_error = payload.get("error")
            allowed = {
                "invalid_auth", "not_authed", "account_inactive", "token_revoked", "missing_scope",
                "ratelimited", "invalid_arguments", "not_allowed_token_type", "channel_not_found",
                "thread_not_found",
            }
            code = api_error if isinstance(api_error, str) and api_error in allowed else "api_error"
            raise SafeError("slack_" + code)
        return payload


def parse_iso(value: str) -> dt.datetime:
    try:
        parsed = dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        raise SafeError("invalid_timestamp") from None
    if parsed.tzinfo is None:
        raise SafeError("timestamp_requires_timezone")
    return parsed.astimezone(dt.timezone.utc)


def iso_z(value: dt.datetime) -> str:
    return value.astimezone(dt.timezone.utc).isoformat().replace("+00:00", "Z")


def ts_decimal(value: Any) -> Decimal | None:
    try:
        return Decimal(str(value))
    except (InvalidOperation, ValueError):
        return None


def iso_to_slack_ts(value: dt.datetime) -> Decimal:
    return Decimal(str(value.timestamp()))


def paging_int(value: Any) -> int | None:
    # Strict positive decimal integer: real int (bool excluded) or digit string
    # for legacy schema compatibility; floats/None/other strings are rejected.
    if isinstance(value, bool):
        return None
    if isinstance(value, str) and value.isdecimal():
        value = int(value)
    if isinstance(value, int) and value > 0:
        return value
    return None


def next_cursor_strict(payload: Mapping[str, Any], seen_cursors: set[str],
                       invalid_code: str, loop_code: str) -> str | None:
    """Extract response_metadata.next_cursor for strict cursor advancement:
    fail closed on non-string cursors and on any cursor already seen."""
    metadata = payload.get("response_metadata")
    cursor = metadata.get("next_cursor", "") if isinstance(metadata, Mapping) else ""
    if not cursor:
        return None
    if not isinstance(cursor, str):
        raise SafeError(invalid_code)
    if cursor in seen_cursors:
        raise SafeError(loop_code)
    seen_cursors.add(cursor)
    return cursor


def safe_snippet(text: Any, secrets: tuple[str, ...] = ()) -> str:
    if not isinstance(text, str):
        return ""
    clean = text
    for secret in secrets:
        if secret:
            clean = clean.replace(secret, "[REDACTED]")
            clean = clean.replace(urllib.parse.quote(secret, safe=""), "[REDACTED]")
    clean = TOKEN_RE.sub("[REDACTED]", clean)
    clean = SENSITIVE_ASSIGNMENT_RE.sub(lambda m: m.group(1) + "=[REDACTED]", clean)
    return " ".join(clean.split())[:MAX_SNIPPET_CHARS]


def author_id_from(message: Mapping[str, Any]) -> str:
    author = message.get("user_id") or message.get("user")
    if isinstance(author, str):
        return author
    if isinstance(author, Mapping) and isinstance(author.get("id"), str):
        return str(author["id"])
    return ""


def candidate_hints(text: str, authored: bool, inbound_mention: bool = False) -> list[str]:
    hints: list[str] = []
    if authored:
        if AUTHORED_COMMITMENT_RE.search(text):
            hints.append("possible_commitment")
        if AUTHORED_RESOLUTION_RE.search(text):
            hints.append("possible_resolution")
        if AUTHORED_DEFERRAL_RE.search(text):
            hints.append("possible_deferral")
    else:
        if inbound_mention:
            hints.append("direct_mention")
        if INBOUND_ASK_RE.search(text) or "?" in text:
            hints.append("possible_ask")
        if REVIEW_DECISION_RE.search(text):
            hints.append("possible_review_or_decision")
        if BLOCKER_RE.search(text):
            hints.append("possible_blocker_or_urgent")
    return hints[:4]


def classification_hints(text: str, authored: bool) -> list[str]:
    """Compatibility helper; collection uses candidate_hints with mention context."""
    return candidate_hints(text, authored)


def extract_channel_id(match: Mapping[str, Any]) -> str:
    direct = match.get("channel_id")
    if isinstance(direct, str):
        return direct
    channel = match.get("channel")
    if isinstance(channel, str):
        return channel
    if isinstance(channel, Mapping) and isinstance(channel.get("id"), str):
        return str(channel["id"])
    return ""


def safe_permalink(value: Any, fallback: str) -> str:
    if not isinstance(value, str):
        return fallback
    try:
        parsed = urllib.parse.urlsplit(value)
    except ValueError:
        return fallback
    host = (parsed.hostname or "").casefold()
    if parsed.scheme != "https" or not host.endswith(".slack.com"):
        return fallback
    return urllib.parse.urlunsplit(("https", parsed.netloc, parsed.path, "", ""))


def event_from_match(
    match: Mapping[str, Any], view: str, credentials: Credentials, hints: list[str] | None = None
) -> dict[str, Any] | None:
    channel_id = extract_channel_id(match)
    message_ts = str(match.get("ts", ""))
    if not ID_RE.fullmatch(channel_id) or not SLACK_TS_RE.fullmatch(message_ts) or ts_decimal(message_ts) is None:
        return None
    raw_thread_ts = str(match.get("thread_ts") or message_ts)
    thread_ts = raw_thread_ts if SLACK_TS_RE.fullmatch(raw_thread_ts) else message_ts
    author_id = author_id_from(match)
    snippet = safe_snippet(match.get("text"), (credentials.xoxc, credentials.xoxd))
    evidence = safe_permalink(match.get("permalink"), f"slack:{channel_id}:{message_ts}")
    return {
        "id": f"slack:{channel_id}:{message_ts}", "channel_id": channel_id,
        "thread_ts": thread_ts, "message_ts": message_ts, "author_id": author_id,
        "permalink": evidence, "evidence_handle": f"slack:{channel_id}:{thread_ts}",
        "classification_hints": list(hints if hints is not None else classification_hints(snippet, view == "authored")),
        "snippet": snippet, "_views": [view],
    }


def base_coverage(mode: str, user_id: str) -> dict[str, Any]:
    return {
        "source": "slack", "mode": mode, "user_id": user_id, "complete": False,
        "truncated": False, "incomplete": True, "fail_closed": True, "pages": 0,
        "matches_scanned": 0, "exact_matches": 0, "candidate_matches": 0,
        "filtered_noise": 0, "deduplicated": 0, "returned": 0, "errors": [],
        "caps": {"pages_per_view": MAX_PAGES_PER_VIEW, "matches_scanned": MAX_MATCHES_SCANNED,
                 "events": MAX_EVENTS, "snippet_chars": MAX_SNIPPET_CHARS},
    }


def search_views(user_id: str, after_date: str, before_date: str) -> tuple[tuple[str, str], ...]:
    """Return the two Slack search views used for delta discovery."""
    return (
        ("authored", f"from:<{user_id}> after:{after_date} before:{before_date}"),
        ("involving", f"with:<@{user_id}> after:{after_date} before:{before_date}"),
    )


def collect(client: SlackClient, credentials: Credentials, start: dt.datetime, end: dt.datetime,
            user_id: str) -> dict[str, Any]:
    coverage = base_coverage("delta", user_id)
    coverage["start"], coverage["end"] = iso_z(start), iso_z(end)
    if end <= start or end - start > dt.timedelta(days=MAX_WINDOW_DAYS):
        coverage["errors"] = ["invalid_window"]
        return {"coverage": coverage, "events": []}
    start_ts, end_ts = iso_to_slack_ts(start), iso_to_slack_ts(end)
    after_date = (start - dt.timedelta(days=1)).date().isoformat()
    before_date = (end + dt.timedelta(days=1)).date().isoformat()
    views = search_views(user_id, after_date, before_date)
    events: dict[str, dict[str, Any]] = {}
    try:
        for view, query in views:
            cursor, page, seen_cursors = "", 1, set()
            while True:
                if page > MAX_PAGES_PER_VIEW:
                    coverage["truncated"] = True
                    raise SafeError("page_cap_reached")
                params: dict[str, Any] = {"query": query, "count": 100, "sort": "timestamp", "sort_dir": "asc"}
                if cursor:
                    params["cursor"] = cursor
                elif page > 1:
                    params["page"] = page
                payload = client.get("search.messages", params)
                coverage["pages"] += 1
                messages = payload.get("messages")
                if not isinstance(messages, Mapping) or not isinstance(messages.get("matches", []), list):
                    raise SafeError("invalid_search_response")
                for raw_match in messages.get("matches", []):
                    coverage["matches_scanned"] += 1
                    if coverage["matches_scanned"] > MAX_MATCHES_SCANNED:
                        coverage["truncated"] = True
                        raise SafeError("scan_cap_reached")
                    if not isinstance(raw_match, Mapping):
                        continue
                    mts = ts_decimal(raw_match.get("ts"))
                    if mts is None or not (start_ts < mts <= end_ts):
                        continue
                    coverage["exact_matches"] += 1
                    author_id = author_id_from(raw_match)
                    authored = author_id == user_id or (view == "authored" and not author_id)
                    inbound = view == "involving" and bool(author_id) and author_id != user_id
                    clean_text = safe_snippet(raw_match.get("text"), (credentials.xoxc, credentials.xoxd))
                    hints = candidate_hints(clean_text, authored=authored, inbound_mention=inbound)
                    if not authored and not inbound:
                        hints = []
                    if not hints:
                        coverage["filtered_noise"] += 1
                        continue
                    coverage["candidate_matches"] += 1
                    event = event_from_match(raw_match, view, credentials, hints)
                    if event is None:
                        continue
                    prior = events.get(event["id"])
                    if prior:
                        coverage["deduplicated"] += 1
                        if view not in prior["_views"]:
                            prior["_views"].append(view)
                        prior["classification_hints"] = sorted(
                            set(prior["classification_hints"] + event["classification_hints"])
                        )[:4]
                    else:
                        if len(events) >= MAX_EVENTS:
                            coverage["truncated"] = True
                            raise SafeError("event_cap_reached")
                        events[event["id"]] = event
                next_cursor = next_cursor_strict(
                    payload, seen_cursors, "invalid_search_response", "pagination_loop")
                if next_cursor:
                    cursor, page = next_cursor, page + 1
                    continue
                if "paging" in messages:
                    paging = messages["paging"]
                    if not isinstance(paging, Mapping):
                        raise SafeError("invalid_search_response")
                    current = paging_int(paging.get("page"))
                    total_pages = paging_int(paging.get("pages"))
                    # Strict progress: explicit positive page/pages, the reported
                    # page must match the one we requested, and the total can't
                    # sit behind it (zero/repeated/mismatched pages loop forever).
                    if current is None or total_pages is None or current != page or total_pages < current:
                        raise SafeError("invalid_search_response")
                    if current < total_pages:
                        cursor, page = "", page + 1
                        continue
                break
    except SafeError as exc:
        coverage["errors"], coverage["returned"] = [exc.code], 0
        return {"coverage": coverage, "events": []}
    ordered = sorted(events.values(), key=lambda event: (Decimal(event["message_ts"]), event["id"]))
    for event in ordered:
        event.pop("_views", None)
    coverage.update({"complete": True, "incomplete": False, "fail_closed": False, "returned": len(ordered)})
    return {"coverage": coverage, "events": ordered}


def context_coverage() -> dict[str, Any]:
    return {
        "source": "slack", "mode": "thread_context", "complete": False, "pagination_complete": False,
        "truncated": False, "incomplete": True, "fail_closed": True, "pages": 0,
        "messages_scanned": 0, "returned": 0, "target_found": False, "errors": [],
        "caps": {"pages": MAX_CONTEXT_PAGES, "messages_scanned": MAX_CONTEXT_SCANNED,
                 "messages_returned": MAX_CONTEXT_MESSAGES, "snippet_chars": MAX_SNIPPET_CHARS},
    }


def collect_thread_context(client: SlackClient, credentials: Credentials, channel_id: str,
                           thread_ts: str, message_ts: str) -> dict[str, Any]:
    coverage = context_coverage()
    context = {"channel_id": channel_id, "thread_ts": thread_ts, "message_ts": message_ts, "messages": []}
    if not ID_RE.fullmatch(channel_id) or not SLACK_TS_RE.fullmatch(thread_ts) or not SLACK_TS_RE.fullmatch(message_ts):
        coverage["errors"] = ["invalid_context_ids"]
        return {"coverage": coverage, "context": context}
    sanitized: dict[str, dict[str, Any]] = {}
    cursor, seen_cursors, page = "", set(), 1
    try:
        while True:
            if page > MAX_CONTEXT_PAGES:
                coverage["truncated"] = True
                raise SafeError("context_page_cap_reached")
            params: dict[str, Any] = {"channel": channel_id, "ts": thread_ts, "limit": 100, "inclusive": "true"}
            if cursor:
                params["cursor"] = cursor
            payload = client.get("conversations.replies", params)
            coverage["pages"] += 1
            messages = payload.get("messages")
            if not isinstance(messages, list):
                raise SafeError("invalid_replies_response")
            for raw in messages:
                coverage["messages_scanned"] += 1
                if coverage["messages_scanned"] > MAX_CONTEXT_SCANNED:
                    coverage["truncated"] = True
                    raise SafeError("context_scan_cap_reached")
                if not isinstance(raw, Mapping):
                    continue
                mts = str(raw.get("ts", ""))
                if not SLACK_TS_RE.fullmatch(mts) or ts_decimal(mts) is None:
                    continue
                sanitized[mts] = {
                    "message_ts": mts, "author_id": author_id_from(raw),
                    "snippet": safe_snippet(raw.get("text"), (credentials.xoxc, credentials.xoxd)),
                    "target": mts == message_ts,
                }
            next_cursor = next_cursor_strict(
                payload, seen_cursors, "invalid_replies_response", "context_pagination_loop")
            if next_cursor:
                cursor, page = next_cursor, page + 1
                continue
            # Some deployments include has_more; without a cursor it cannot be completed safely.
            if payload.get("has_more") is True:
                raise SafeError("context_pagination_incomplete")
            break
        ordered = sorted(sanitized.values(), key=lambda item: Decimal(item["message_ts"]))
        target_index = next((i for i, item in enumerate(ordered) if item["message_ts"] == message_ts), None)
        if target_index is None:
            raise SafeError("context_target_not_found")
        start = max(0, target_index - (MAX_CONTEXT_MESSAGES // 2))
        end = min(len(ordered), start + MAX_CONTEXT_MESSAGES)
        start = max(0, end - MAX_CONTEXT_MESSAGES)
        selected = ordered[start:end]
    except SafeError as exc:
        coverage["errors"] = [exc.code]
        return {"coverage": coverage, "context": context}
    context["messages"] = selected
    coverage.update({"complete": True, "pagination_complete": True, "incomplete": False,
                     "fail_closed": False, "target_found": True, "returned": len(selected)})
    return {"coverage": coverage, "context": context}


def probe(client: SlackClient, user_id: str) -> dict[str, Any]:
    coverage = base_coverage("probe", user_id)
    coverage.pop("caps", None)
    coverage.update({"authenticated": False, "events_returned": 0})
    try:
        client.get("auth.test")
    except SafeError as exc:
        coverage["errors"] = [exc.code]
        return {"coverage": coverage, "events": []}
    coverage.update({"authenticated": True, "complete": True, "incomplete": False, "fail_closed": False})
    return {"coverage": coverage, "events": []}


def parser() -> argparse.ArgumentParser:
    result = JsonArgumentParser(add_help=True)
    result.add_argument("--start")
    result.add_argument("--end")
    result.add_argument("--user", default=DEFAULT_USER)
    modes = result.add_mutually_exclusive_group()
    modes.add_argument("--probe", action="store_true")
    modes.add_argument("--context", action="store_true")
    result.add_argument("--channel")
    result.add_argument("--thread")
    result.add_argument("--message")
    return result


def safe_failure(code: str, mode: str = "delta", user_id: str = DEFAULT_USER) -> dict[str, Any]:
    coverage = context_coverage() if mode == "thread_context" else base_coverage(mode, user_id)
    coverage["errors"] = [code]
    return {"coverage": coverage, "context": {"messages": []}} if mode == "thread_context" else {"coverage": coverage, "events": []}


def main(argv: list[str] | None = None) -> int:
    mode, user_id = "delta", DEFAULT_USER
    try:
        args = parser().parse_args(argv)
        user_id = args.user
        mode = "probe" if args.probe else ("thread_context" if args.context else "delta")
        if not USER_RE.fullmatch(args.user):
            raise SafeError("invalid_user_id")
        if args.context:
            if not args.channel or not args.thread or not args.message or args.start or args.end:
                raise SafeError("context_ids_required")
        elif not args.probe and (not args.start or not args.end):
            raise SafeError("start_end_required")
        credentials = load_credentials()
        client = SlackClient(credentials)
        if args.probe:
            result = probe(client, args.user)
        elif args.context:
            result = collect_thread_context(client, credentials, args.channel, args.thread, args.message)
        else:
            result = collect(client, credentials, parse_iso(args.start), parse_iso(args.end), args.user)
    except ParseFailure as exc:
        result = safe_failure(str(exc), mode, user_id)
    except SafeError as exc:
        result = safe_failure(exc.code, mode, user_id)
    except Exception:
        result = safe_failure("internal_error", mode, user_id)
    sys.stdout.write(json.dumps(result, ensure_ascii=False, separators=(",", ":"), sort_keys=True) + "\n")
    return 0 if result["coverage"].get("complete") else 1


if __name__ == "__main__":
    raise SystemExit(main())
