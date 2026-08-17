import datetime as dt
import json
import os
import re
import ssl
import sys
import urllib.error
import urllib.parse
import urllib.request

ENDPOINT = "https://api.linear.app/graphql"
MAX_RESPONSE_BYTES = 2 * 1024 * 1024
ISSUE_PAGE = 50
MAX_ISSUE_PAGES = 4
MAX_CANDIDATES = 100
SUBWINDOWS = 16
CONTEXT_PAGE = 20
MAX_CONTEXT_PAGES = 2
GCN_TEAM_ID = "81e7b769-2a46-4e2a-8db5-c165a7963b0e"

class NoRedirectHandler(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise urllib.error.HTTPError(req.full_url, code, "redirect rejected", headers, fp)

TLS_CONTEXT = ssl.create_default_context()
HTTP = urllib.request.build_opener(
    urllib.request.HTTPSHandler(context=TLS_CONTEXT), NoRedirectHandler())
SECRET_ASSIGNMENT = re.compile(
    r"(?i)\b(authorization|api[_-]?key|access[_-]?token|token|secret|password)"
    r"(\s*[:=]\s*)(?:bearer\s+)?(?:\"[^\"]*\"|'[^']*'|[^\s,;]+)")
BEARER_TOKEN = re.compile(r"(?i)\bbearer\s+[A-Za-z0-9._~+/=-]{12,}")
COMMON_TOKEN = re.compile(r"\b(?:lin_api_|sk-|gh[pousr]_|xox[baprs]-)[A-Za-z0-9._~+/=-]{8,}")
OPS = {
    "AssignedIssues": """query AssignedIssues($since: DateTimeOrDuration!, $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after, includeArchived: true,
        filter: {assignee: {isMe: {eq: true}}, updatedAt: {gte: $since}}) {
        nodes { id identifier title url description priority dueDate createdAt updatedAt
          canceledAt completedAt state { id name type } assignee { id name }
          team { id key name } }
        pageInfo { hasNextPage endCursor }
      }
    }""",
    "TeamIssues": """query TeamIssues($teamId: ID!, $since: DateTimeOrDuration!, $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after, includeArchived: true,
        filter: {team: {id: {eq: $teamId}}, updatedAt: {gte: $since}}) {
        nodes { id identifier title url description priority dueDate createdAt updatedAt
          canceledAt completedAt state { id name type } assignee { id name }
          team { id key name } }
        pageInfo { hasNextPage endCursor }
      }
    }""",
    "AssignedIssuesWindow": """query AssignedIssuesWindow($since: DateTimeOrDuration!, $until: DateTimeOrDuration!, $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after, includeArchived: true,
        filter: {assignee: {isMe: {eq: true}}, updatedAt: {gte: $since, lte: $until}}) {
        nodes { id identifier title url description priority dueDate createdAt updatedAt
          canceledAt completedAt state { id name type } assignee { id name }
          team { id key name } }
        pageInfo { hasNextPage endCursor }
      }
    }""",
    "TeamIssuesWindow": """query TeamIssuesWindow($teamId: ID!, $since: DateTimeOrDuration!, $until: DateTimeOrDuration!, $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after, includeArchived: true,
        filter: {team: {id: {eq: $teamId}}, updatedAt: {gte: $since, lte: $until}}) {
        nodes { id identifier title url description priority dueDate createdAt updatedAt
          canceledAt completedAt state { id name type } assignee { id name }
          team { id key name } }
        pageInfo { hasNextPage endCursor }
      }
    }""",
    "IssueComments": """query IssueComments($id: String!, $first: Int!, $after: String) {
      issue(id: $id) { id comments(first: $first, after: $after) {
        nodes { id body createdAt updatedAt user { id name } }
        pageInfo { hasNextPage endCursor }
      } }
    }""",
    "IssueHistory": """query IssueHistory($id: String!, $first: Int!, $after: String) {
      issue(id: $id) { id history(first: $first, after: $after) {
        nodes { id createdAt updatedAt actor { id name }
          fromState { id name type } toState { id name type }
          fromAssignee { id name } toAssignee { id name }
          fromPriority toPriority fromDueDate toDueDate }
        pageInfo { hasNextPage endCursor }
      } }
    }""",
}

def instant(value):
    parsed = dt.datetime.fromisoformat(value.replace("Z", "+00:00"))
    if parsed.tzinfo is None:
        raise ValueError("timestamp must include a timezone")
    return parsed.astimezone(dt.timezone.utc)

def _post(operation, query, variables):
    if "mutation" in query.lower():
        raise RuntimeError("non-allowlisted GraphQL operation")
    request = urllib.request.Request(
        ENDPOINT,
        data=json.dumps({"operationName": operation, "query": query, "variables": variables}).encode("utf-8"),
        headers={"Content-Type": "application/json", "Authorization": os.environ["LINEAR_API_KEY"]},
        method="POST")
    with HTTP.open(request, timeout=20) as response:
        if response.geturl() != ENDPOINT:
            raise RuntimeError("Linear response URL mismatch")
        if response.status != 200:
            raise RuntimeError(f"Linear HTTP status {response.status}")
        raw = response.read(MAX_RESPONSE_BYTES + 1)
        if len(raw) > MAX_RESPONSE_BYTES:
            raise RuntimeError("Linear response exceeds byte limit")
        payload = json.loads(raw.decode("utf-8"))
    if not isinstance(payload, dict) or payload.get("errors") or not isinstance(payload.get("data"), dict):
        raise RuntimeError(f"Linear GraphQL/shape failure in {operation}")
    return payload["data"]

def graphql(operation, variables):
    query = OPS.get(operation)
    if query is None or "mutation" in query.lower():
        raise RuntimeError("non-allowlisted GraphQL operation")
    return _post(operation, query, variables)

def connection(value, label):
    if not isinstance(value, dict) or not isinstance(value.get("nodes"), list):
        raise RuntimeError(f"malformed {label} connection")
    page = value.get("pageInfo")
    if not isinstance(page, dict) or not isinstance(page.get("hasNextPage"), bool):
        raise RuntimeError(f"malformed {label} pageInfo")
    if page["hasNextPage"] and not page.get("endCursor"):
        raise RuntimeError(f"missing {label} pagination cursor")
    return value["nodes"], page

def bounded_context(operation, field, issue_id):
    nodes, after, source_has_more, pages = [], None, False, 0
    for _ in range(MAX_CONTEXT_PAGES):
        data = graphql(operation, {"id": issue_id, "first": CONTEXT_PAGE, "after": after})
        issue = data.get("issue")
        if not isinstance(issue, dict) or issue.get("id") != issue_id:
            raise RuntimeError(f"malformed {field} issue context")
        batch, page = connection(issue.get(field), field)
        nodes.extend(batch)
        pages += 1
        source_has_more = page["hasNextPage"]
        if not source_has_more:
            break
        next_after = page["endCursor"]
        if next_after == after:
            raise RuntimeError(f"stalled {field} pagination")
        after = next_after
    return nodes, {"pages_succeeded": pages, "items": len(nodes),
                   "bound": CONTEXT_PAGE * MAX_CONTEXT_PAGES, "source_has_more": source_has_more}

def iso(moment):
    return moment.isoformat().replace("+00:00", "Z")

def scan_issues(operation, extra, since, until=None):
    name = operation + "Window" if until is not None else operation
    nodes, after, pages, truncated = [], None, 0, False
    while True:
        if pages >= MAX_ISSUE_PAGES:
            truncated = True
            break
        variables = {"since": since, "first": ISSUE_PAGE, "after": after}
        if until is not None:
            variables["until"] = until
        variables.update(extra)
        data = graphql(name, variables)
        if not isinstance(data.get("viewer"), dict) or not data["viewer"].get("id"):
            raise RuntimeError("authenticated viewer shape is unavailable")
        batch, page = connection(data.get("issues"), "issues")
        nodes.extend(batch)
        pages += 1
        if not page["hasNextPage"]:
            break
        next_after = page["endCursor"]
        if next_after == after:
            raise RuntimeError(f"stalled {name} pagination")
        after = next_after
    return nodes, pages, truncated

def drain(operation, extra, start, end):
    nodes, pages, truncated = scan_issues(operation, extra, iso(start))
    if not truncated:
        return nodes, pages, end, False
    span = (end - start) / SUBWINDOWS
    edges = [start + span * n for n in range(SUBWINDOWS)] + [end]
    nodes, complete_through = [], start
    for lo, hi in zip(edges, edges[1:]):
        batch, used, cut = scan_issues(operation, extra, iso(lo), iso(hi))
        pages += used
        if cut:
            return nodes, pages, complete_through, True
        nodes.extend(batch)
        complete_through = hi
    return nodes, pages, end, False

def sanitized(value, limit):
    if not isinstance(value, str):
        return value
    result = value
    credential = os.environ.get("LINEAR_API_KEY", "")
    if credential:
        for form in {credential, urllib.parse.quote(credential, safe="")}:
            if form:
                result = result.replace(form, "[REDACTED]")
    result = SECRET_ASSIGNMENT.sub(lambda match: match.group(1) + match.group(2) + "[REDACTED]", result)
    result = BEARER_TOKEN.sub("Bearer [REDACTED]", result)
    result = COMMON_TOKEN.sub("[REDACTED]", result)
    return result[:limit]

try:
    if not os.environ.get("LINEAR_API_KEY"):
        raise RuntimeError("LINEAR_API_KEY is unavailable")
    cursor = instant(os.environ["LINEAR_CURSOR"])
    run_end = instant(os.environ["LINEAR_RUN_END"])
    if run_end < cursor:
        raise RuntimeError("run end precedes cursor")
    overlap = cursor - dt.timedelta(minutes=5)
    assigned, assigned_pages, assigned_through, assigned_cut = drain("AssignedIssues", {}, overlap, run_end)
    team, team_pages, team_through, team_cut = drain("TeamIssues", {"teamId": GCN_TEAM_ID}, overlap, run_end)
    issue_pages = assigned_pages + team_pages
    pages_truncated = assigned_cut or team_cut
    window_end = min(assigned_through, team_through, run_end)
    scanned, seen_ids = [], set()
    for issue in assigned + team:
        if not isinstance(issue, dict) or not issue.get("id") or not issue.get("identifier"):
            raise RuntimeError("malformed issue node")
        if issue["id"] in seen_ids:
            continue
        seen_ids.add(issue["id"])
        scanned.append(issue)
    candidates = []
    for issue in scanned:
        updated = instant(issue.get("updatedAt", ""))
        if cursor < updated <= window_end:
            candidates.append(issue)
    candidates.sort(key=lambda item: item["updatedAt"])
    candidates_dropped = 0
    if len(candidates) > MAX_CANDIDATES:
        edge = candidates[MAX_CANDIDATES - 1]["updatedAt"]
        kept = [item for item in candidates if item["updatedAt"] <= edge]
        candidates_dropped = len(candidates) - len(kept)
        candidates = kept
    truncated = pages_truncated or candidates_dropped > 0
    proposed_cursor = instant(candidates[-1]["updatedAt"]) if candidates_dropped else max(cursor, window_end)
    output_items = []
    for issue in sorted(candidates, key=lambda item: (item["updatedAt"], item["identifier"])):
        comments, comments_meta = bounded_context("IssueComments", "comments", issue["id"])
        history, history_meta = bounded_context("IssueHistory", "history", issue["id"])
        issue["description"] = sanitized(issue.get("description"), 2000)
        for comment in comments:
            comment["body"] = sanitized(comment.get("body"), 1000)
        output_items.append({"id": "linear:" + issue["identifier"], "issue": issue,
            "comments": comments, "activity": history,
            "context_completeness": {"all_required_calls_succeeded": True,
                "comments": comments_meta, "activity": history_meta}})
    print(json.dumps({"source": "linear", "retrieval": {
        "overlap_start": iso(overlap), "exact_cursor": iso(cursor),
        "proposed_cursor": iso(proposed_cursor), "cursor_advances": proposed_cursor > cursor,
        "complete_through": iso(window_end), "views": ["AssignedIssues", "TeamIssues"],
        "deduplicated_by_issue_id": True, "issue_connection_exhausted": not pages_truncated,
        "issue_pages_succeeded": issue_pages, "issues_scanned": len(scanned),
        "exact_candidates": len(output_items), "candidates_dropped": candidates_dropped,
        "truncated": truncated, "all_required_context_calls_succeeded": True,
        "context_is_bounded": True}, "items": output_items}, ensure_ascii=False))
except (KeyError, ValueError, RuntimeError, OSError, json.JSONDecodeError) as exc:
    print(f"linear collector failed closed: {type(exc).__name__}: {exc}", file=sys.stderr)
    raise SystemExit(1)
