#!/usr/bin/env python3
"""Deterministic Linear board renderer for `daily-work-brief`.

Prints the finished `**Board**` block on stdout, ready to be pasted into the
brief byte for byte. Same board state in, same bytes out.

WHY THIS IS RAW HTTP AND NOT `mcp__linear__*`
---------------------------------------------
Measured on the VM (`status/probes/wf6/deterministic-render-options.md`):
Hermes v0.20.0 has NO command that invokes one named MCP tool with arguments
outside the LLM agent loop. `hermes mcp` has no `call`/`invoke`; `hermes mcp
test` takes a server name and nothing else; `hermes tools` only enables and
disables. So a *deterministic* step physically cannot use the native Linear
tools — every `mcp__linear__*` call goes through the model, and the model under
context compaction is exactly what fabricated 7 of 12 board rows on 2026-08-10
(`status/probes/wf6/deploy-and-verify.md`).

**This raw-HTTP exception is scoped to the board block and nothing else.**
Every other Linear read, and every Linear write, in every skill stays on the
native `mcp__linear__*` tools per `linear-ticketing`. There is no write path
here: `_post()` rejects any query containing `mutation` before it builds a
request, and the four queries below are the only ones that exist. If a future
Hermes ships a non-agentic MCP invocation, delete this script and move the
block back to the native tool.

CREDENTIAL
----------
`LINEAR_API_KEY` is read from `os.environ` only. It is NEVER placed on argv,
never printed, never logged, never written to a file, and no exception text
reaches any stream: every failure is re-raised as a short fixed code (see
`BoardError`), because a raw urllib/ssl exception can echo request headers.
The key reaches this process only on the shell/terminal-tool path — the
`execute_code` sandbox scrubs any var whose name contains `KEY` (measured).

Usage:  python3 linear_board.py            # fetch + render
        python3 linear_board.py --selftest # offline logic check, no network
"""

import json
import os
import re
import ssl
import sys
import urllib.error
import urllib.request

ENDPOINT = "https://api.linear.app/graphql"
GCN_TEAM_ID = "81e7b769-2a46-4e2a-8db5-c165a7963b0e"

MAX_RESPONSE_BYTES = 2 * 1024 * 1024
PAGE = 100                 # nodes per request; measured-safe (the probe ran 50)
MAX_PAGES = 12             # explicit cap: 1200 issues per view, then fail closed
MAX_ROWS = 12              # rows rendered before `+N more`
TITLE_CAP = 120

# Client-side terminal assert. `duplicate` is in the set because it was measured
# as a live terminal state (NMC-499) and one was rendered as a live row in the
# fabrication incident. Only the two the GraphQL enum certainly carries are sent
# server-side.
TERMINAL = frozenset({"completed", "canceled", "duplicate"})
TERMINAL_SERVER_SIDE = ["completed", "canceled"]

# Linear priority ints: 0 No priority, 1 Urgent, 2 High, 3 Medium, 4 Low.
# Display order 1,2,3,4 then 0 last.
PRIORITY_RANK = {1: 0, 2: 1, 3: 2, 4: 3, 0: 4}

UNAVAILABLE = "board unavailable (coverage unproven)"

# Node selection is deliberately minimal: exactly the five things a row prints,
# plus the team id the GCN-first sort keys on. No description, no url, no
# timestamps — nothing this script does not render may enter the payload.
_NODES = """nodes { identifier title priority state { name type } team { id key } }
        pageInfo { hasNextPage endCursor }"""

QUERIES = {
    # Server-side filter, asserted client-side regardless (see normalize()).
    # NOTE: the `state.type` `nin` comparator is not yet measured on this
    # transport — `team.id.eq` and `assignee.isMe.eq` are. It is safe to use
    # here anyway: on GraphQL a bad filter returns `errors` and this script
    # prints the unavailable sentinel, it can NEVER silently return zero rows
    # the way the native tool's `state` param does. The client-side drop is the
    # real guarantee; the server-side filter only keeps the page budget small.
    "BoardTeamIssues": """query BoardTeamIssues($teamId: ID!, $types: [String!], $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after,
        filter: {team: {id: {eq: $teamId}}, state: {type: {nin: $types}}}) {
        %s
      }
    }""" % _NODES,
    "BoardAssignedIssues": """query BoardAssignedIssues($types: [String!], $first: Int!, $after: String) {
      viewer { id }
      issues(first: $first, after: $after,
        filter: {assignee: {isMe: {eq: true}}, state: {type: {nin: $types}}}) {
        %s
      }
    }""" % _NODES,
}

VIEWS = (
    ("BoardTeamIssues", {"teamId": GCN_TEAM_ID}),
    ("BoardAssignedIssues", {}),
)

_SUFFIX = re.compile(r"-(\d+)$")


class BoardError(Exception):
    """Carries a short fixed code only. Never an upstream exception's text."""

    def __init__(self, code):
        super().__init__(code)
        self.code = code


class NoRedirectHandler(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise urllib.error.HTTPError(req.full_url, code, "redirect rejected", headers, fp)


HTTP = urllib.request.build_opener(
    urllib.request.HTTPSHandler(context=ssl.create_default_context()),
    NoRedirectHandler(),
)


def _post(operation, variables):
    query = QUERIES.get(operation)
    if query is None or "mutation" in query.lower():
        raise BoardError("op")
    body = json.dumps(
        {"operationName": operation, "query": query, "variables": variables}
    ).encode("utf-8")
    request = urllib.request.Request(
        ENDPOINT,
        data=body,
        headers={
            "Content-Type": "application/json",
            "Authorization": _credential(),
        },
        method="POST",
    )
    try:
        with HTTP.open(request, timeout=20) as response:
            if response.geturl() != ENDPOINT:
                raise BoardError("url")
            if response.status != 200:
                raise BoardError("http")
            raw = response.read(MAX_RESPONSE_BYTES + 1)
    except BoardError:
        raise
    except Exception:
        # `from None`: no upstream exception may survive to a traceback — an
        # HTTPError repr can carry request headers, i.e. the credential.
        raise BoardError("net") from None
    if len(raw) > MAX_RESPONSE_BYTES:
        raise BoardError("size")
    try:
        payload = json.loads(raw.decode("utf-8"))
    except Exception:
        raise BoardError("parse") from None
    if not isinstance(payload, dict) or payload.get("errors"):
        raise BoardError("graphql")
    data = payload.get("data")
    if not isinstance(data, dict):
        raise BoardError("shape")
    return data


def _credential():
    key = os.environ.get("LINEAR_API_KEY")
    if not key:
        raise BoardError("nokey")
    return key


def normalize(node):
    """One payload node -> one row dict, or None if the row is terminal.

    Raises on a malformed node: a board that cannot prove a row is better
    withheld than guessed.
    """
    if not isinstance(node, dict):
        raise BoardError("node")
    identifier = node.get("identifier")
    title = node.get("title")
    state = node.get("state")
    team = node.get("team") or {}
    if not isinstance(identifier, str) or not isinstance(title, str):
        raise BoardError("node")
    if not isinstance(state, dict) or not isinstance(state.get("name"), str):
        raise BoardError("node")
    if state.get("type") in TERMINAL:  # the assert, independent of the filter
        return None
    priority = node.get("priority")
    if not isinstance(priority, int) or isinstance(priority, bool):
        priority = 0
    suffix = _SUFFIX.search(identifier)
    return {
        "identifier": identifier,
        # The one normalisation applied to a title: control whitespace becomes a
        # space, because a newline would break the row structure of the block.
        "title": title.translate({0x0A: " ", 0x0D: " ", 0x09: " "}),
        "priority": priority,
        "state_name": state["name"],
        "team_id": team.get("id") if isinstance(team, dict) else None,
        "team_key": (team.get("key") or "") if isinstance(team, dict) else "",
        "number": int(suffix.group(1)) if suffix else 0,
    }


def sort_key(row):
    return (
        PRIORITY_RANK.get(row["priority"], 4),
        0 if row["team_id"] == GCN_TEAM_ID else 1,
        row["team_key"],
        row["number"],
        row["identifier"],
    )


def priority_label(priority):
    return "P%d" % priority if priority in (1, 2, 3, 4) else "P-"


def clip(title):
    return title if len(title) <= TITLE_CAP else title[: TITLE_CAP - 1] + "…"


def render(rows):
    """rows: already sorted. Returns the finished block, no trailing newline."""
    lines = []
    if not rows:
        lines.append("Board: clear.")
    else:
        lines.append("**Board** (priority-sorted)")
        for row in rows[:MAX_ROWS]:
            lines.append(
                "%s %s %s — %s"
                % (
                    priority_label(row["priority"]),
                    row["identifier"],
                    clip(row["title"]),
                    row["state_name"],
                )
            )
        if len(rows) > MAX_ROWS:
            # Computed from the rows actually in hand, never estimated. It is
            # only ever printed on the path where every view proved complete.
            lines.append("+%d more" % (len(rows) - MAX_ROWS))
    lines.append(
        "Coverage: %d views, both complete (%d issues)" % (len(VIEWS), len(rows))
    )
    return "\n".join(lines)


def fetch_view(operation, extra, post=_post):
    """Paginate one view to completion. Raises BoardError if it cannot."""
    rows = {}
    after = None
    for _ in range(MAX_PAGES):
        variables = {"first": PAGE, "after": after, "types": TERMINAL_SERVER_SIDE}
        variables.update(extra)
        data = post(operation, variables)
        viewer = data.get("viewer")
        if not isinstance(viewer, dict) or not viewer.get("id"):
            raise BoardError("viewer")
        connection = data.get("issues")
        if not isinstance(connection, dict) or not isinstance(connection.get("nodes"), list):
            raise BoardError("conn")
        page = connection.get("pageInfo")
        if not isinstance(page, dict) or not isinstance(page.get("hasNextPage"), bool):
            raise BoardError("page")
        for node in connection["nodes"]:
            row = normalize(node)
            if row is not None:
                rows[row["identifier"]] = row
        if not page["hasNextPage"]:
            return rows
        nxt = page.get("endCursor")
        if not nxt or nxt == after:
            raise BoardError("stall")
        after = nxt
    # Page cap reached with hasNextPage still true: coverage is unproven.
    raise BoardError("pagecap")


def build(fetch=fetch_view):
    """Returns (block, exit_code, reason). Never raises.

    THE COVERAGE GATE: if either view fails, errors, or exhausts its page cap
    with more pages outstanding, the block is the unavailable sentinel and
    nothing else. A partial board is never emitted.
    """
    rows = {}
    try:
        for operation, extra in VIEWS:
            rows.update(fetch(operation, extra))
    except BoardError as exc:
        return UNAVAILABLE, 1, exc.code
    except Exception:
        return UNAVAILABLE, 1, "internal"
    return render(sorted(rows.values(), key=sort_key)), 0, ""


def main(argv):
    if "--selftest" in argv:
        return selftest()
    block, code, reason = build()
    sys.stdout.write(block + "\n")
    if code:
        sys.stderr.write("linear_board: %s\n" % reason)
    return code


# --------------------------------------------------------------------------
# Self-test: the one runnable check. Fixtures only, no network, no credential.
# --------------------------------------------------------------------------

def _node(identifier, title, priority, name, type_, team_id=GCN_TEAM_ID, key="GCN"):
    return {
        "identifier": identifier,
        "title": title,
        "priority": priority,
        "state": {"name": name, "type": type_},
        "team": {"id": team_id, "key": key},
    }


_OTHER = ("11a6f6b1-bf49-4c5f-8426-190474ce22e2", "NMC")

FIXTURE = [
    _node("GCN-10", "Prune the Linear MCP tool surface", 3, "Todo", "unstarted"),
    _node("GCN-9", "Wire Linear MCP server natively", 3, "Todo", "unstarted"),
    _node("GCN-42", "Ship the thing", 1, "In Progress", "started"),
    _node("GCN-51", "Scope the migration", 0, "Backlog", "backlog"),
    _node("NMC-4179", "Review the other thing", 2, "Todo", "unstarted", *_OTHER),
    _node("GCN-3", "engagement-checker: Linear push+pull", 2, "In Progress", "started"),
    _node("NMC-499", "Terminal duplicate row", 2, "Duplicate", "duplicate", *_OTHER),
    _node("NMC-514", "Terminal done row", 2, "Done", "completed", *_OTHER),
    _node("GCN-8", "Terminal canceled row", 4, "Canceled", "canceled"),
]
FIXTURE += [
    _node("GCN-%d" % n, "Filler %d" % n, 4, "Todo", "unstarted") for n in range(60, 70)
]


def _fixture_fetch(operation, extra):
    if operation == "BoardTeamIssues":
        rows = {}
        for node in FIXTURE:
            row = normalize(node)
            if row is not None:
                rows[row["identifier"]] = row
        return rows
    return {}  # the assigned view returns the same issues; dedup by identifier


def selftest():
    # 1. normalize(): terminal rows never survive, whatever the filter did.
    assert normalize(_node("X-1", "t", 1, "Done", "completed")) is None
    assert normalize(_node("X-2", "t", 1, "Canceled", "canceled")) is None
    assert normalize(_node("X-3", "t", 1, "Duplicate", "duplicate")) is None
    assert normalize(_node("X-4", "t", 1, "Todo", "unstarted"))["identifier"] == "X-4"

    # 2. a malformed node fails closed rather than rendering a guess.
    for bad in ({"identifier": "X-5"}, {"title": "t"}, "not-a-dict"):
        try:
            normalize(bad)
            raise AssertionError("malformed node accepted")
        except BoardError:
            pass

    # 3. full render: sort order, P- for 0, GCN-first, numeric suffix, cap+more.
    block, code, reason = build(_fixture_fetch)
    assert (code, reason) == (0, ""), (code, reason)
    expected = "\n".join([
        "**Board** (priority-sorted)",
        "P1 GCN-42 Ship the thing — In Progress",
        "P2 GCN-3 engagement-checker: Linear push+pull — In Progress",
        "P2 NMC-4179 Review the other thing — Todo",
        "P3 GCN-9 Wire Linear MCP server natively — Todo",
        "P3 GCN-10 Prune the Linear MCP tool surface — Todo",
        "P4 GCN-60 Filler 60 — Todo",
        "P4 GCN-61 Filler 61 — Todo",
        "P4 GCN-62 Filler 62 — Todo",
        "P4 GCN-63 Filler 63 — Todo",
        "P4 GCN-64 Filler 64 — Todo",
        "P4 GCN-65 Filler 65 — Todo",
        "P4 GCN-66 Filler 66 — Todo",
        "+4 more",
        "Coverage: 2 views, both complete (16 issues)",
    ])
    assert block == expected, "\n--- got ---\n%s\n--- want ---\n%s" % (block, expected)

    # 4. P0 renders P- and sorts last; it is inside the +N remainder above.
    tail = sorted(
        (r for r in (normalize(n) for n in FIXTURE) if r is not None), key=sort_key
    )[-1]
    assert tail["identifier"] == "GCN-51", tail
    assert priority_label(tail["priority"]) == "P-"

    # 5. empty board.
    empty, code, _ = build(lambda op, extra: {})
    assert code == 0
    assert empty == "Board: clear.\nCoverage: 2 views, both complete (0 issues)", empty

    # 6. long title is clipped, not dropped or paraphrased.
    long_row = normalize(_node("GCN-1", "x" * 200, 1, "Todo", "unstarted"))
    clipped = clip(long_row["title"])
    assert len(clipped) == TITLE_CAP and clipped.endswith("…") and clipped[:10] == "x" * 10

    # 7. THE COVERAGE GATE: any view failure yields the sentinel and nothing else.
    def _boom(op, extra):
        raise BoardError("http")

    assert build(_boom) == (UNAVAILABLE, 1, "http")

    def _partial(op, extra):
        if op == "BoardTeamIssues":
            return _fixture_fetch(op, extra)
        raise BoardError("graphql")

    assert build(_partial) == (UNAVAILABLE, 1, "graphql"), "partial board leaked"

    # 8. page cap with hasNextPage still true is a coverage failure, not a board.
    def _never_ends(op, variables):
        return {
            "viewer": {"id": "u"},
            "issues": {
                "nodes": [_node("GCN-%d" % variables["first"], "t", 1, "Todo", "unstarted")],
                "pageInfo": {"hasNextPage": True, "endCursor": "c%s" % variables["after"]},
            },
        }

    try:
        fetch_view("BoardTeamIssues", {}, post=_never_ends)
        raise AssertionError("page cap not enforced")
    except BoardError as exc:
        assert exc.code == "pagecap", exc.code

    # 9. a stalled cursor is a failure, not an infinite loop.
    def _stalls(op, variables):
        return {
            "viewer": {"id": "u"},
            "issues": {"nodes": [], "pageInfo": {"hasNextPage": True, "endCursor": None}},
        }

    try:
        fetch_view("BoardTeamIssues", {}, post=_stalls)
        raise AssertionError("stall not caught")
    except BoardError as exc:
        assert exc.code == "stall", exc.code

    # 10. the sentinel text is contractual — the skill matches on it verbatim.
    assert UNAVAILABLE == "board unavailable (coverage unproven)"

    sys.stdout.write("selftest OK\n")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
