#!/usr/bin/env bash
# smoke-test.sh — TARS v2 end-to-end smoke test
#
# Usage:
#   ./scripts/smoke-test.sh [BASE_URL]
#
# Defaults:
#   BASE_URL=http://localhost:8080
#
# Prerequisites:
#   - curl, jq (both must be in PATH)
#   - A running TARS stack (make up, or docker compose up -d)
#   - A valid tars_session cookie, OR set TARS_SESSION env var
#     (obtain by completing GitHub OAuth and copying from browser devtools)
#
# Exit codes:
#   0  — all checks passed
#   1  — one or more checks failed
#
# Environment variables:
#   BASE_URL       API base URL (default: http://localhost:8080)
#   TARS_SESSION   Pre-existing session cookie value (skips auth check if set)
#   REPO_PATH      Absolute path to a real git repo on the server host
#                  (default: /tmp/tars-smoke-repo, created if missing)
#   VERBOSE        Set to 1 for full response bodies

set -euo pipefail

# ── Config ────────────────────────────────────────────────────────────────────
BASE_URL="${1:-${BASE_URL:-http://localhost:8080}}"
API="${BASE_URL}/api/v1"
REPO_PATH="${REPO_PATH:-/tmp/tars-smoke-repo}"
VERBOSE="${VERBOSE:-0}"

# ── Colour helpers ────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

PASS=0
FAIL=0
SKIP=0

pass() { echo -e "  ${GREEN}✓${RESET} $1" >&2; ((PASS++)) || true; }
fail() { echo -e "  ${RED}✗${RESET} $1" >&2; ((FAIL++)) || true; }
skip() { echo -e "  ${YELLOW}⊘${RESET} $1 (skipped)" >&2; ((SKIP++)) || true; }
section() { echo -e "\n${BOLD}${CYAN}▶ $1${RESET}" >&2; }

summarise() {
  local total=$((PASS + FAIL + SKIP))
  echo "" >&2
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}" >&2
  echo -e "${BOLD}Smoke test results${RESET}" >&2
  echo -e "  ${GREEN}Passed${RESET}:  ${PASS}" >&2
  echo -e "  ${RED}Failed${RESET}:  ${FAIL}" >&2
  echo -e "  ${YELLOW}Skipped${RESET}: ${SKIP}" >&2
  echo -e "  Total:   ${total}" >&2
  echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}" >&2
  if [[ "${FAIL}" -gt 0 ]]; then
    echo -e "${RED}${BOLD}FAILED${RESET} — ${FAIL} check(s) did not pass." >&2
    exit 1
  else
    echo -e "${GREEN}${BOLD}PASSED${RESET} — all ${PASS} check(s) passed (${SKIP} skipped)." >&2
    exit 0
  fi
}

# ── Dependency checks ─────────────────────────────────────────────────────────
section "Dependency checks"

if ! command -v curl &>/dev/null; then
  fail "curl not found — install curl to run smoke tests"
  exit 1
fi
pass "curl available"

if ! command -v jq &>/dev/null; then
  fail "jq not found — install jq to run smoke tests"
  exit 1
fi
pass "jq available"

# ── Helper: HTTP request with session cookie ──────────────────────────────────
SESSION_COOKIE="${TARS_SESSION:-}"

api_get() {
  local path="$1"
  local extra="${2:-}"
  curl -s -w '\n%{http_code}' \
    ${SESSION_COOKIE:+-H "Cookie: tars_session=${SESSION_COOKIE}"} \
    ${extra} \
    "${API}${path}"
}

api_post() {
  local path="$1"
  local body="${2:-{}}"
  curl -s -w '\n%{http_code}' \
    -X POST \
    -H "Content-Type: application/json" \
    ${SESSION_COOKIE:+-H "Cookie: tars_session=${SESSION_COOKIE}"} \
    -d "${body}" \
    "${API}${path}"
}

api_patch() {
  local path="$1"
  local body="${2:-{}}"
  curl -s -w '\n%{http_code}' \
    -X PATCH \
    -H "Content-Type: application/json" \
    ${SESSION_COOKIE:+-H "Cookie: tars_session=${SESSION_COOKIE}"} \
    -d "${body}" \
    "${API}${path}"
}

api_delete() {
  local path="$1"
  curl -s -w '\n%{http_code}' \
    -X DELETE \
    ${SESSION_COOKIE:+-H "Cookie: tars_session=${SESSION_COOKIE}"} \
    "${API}${path}"
}

# Parse last line as HTTP status, rest as body
http_status() { tail -1 <<< "$1"; }
http_body()   { head -n -1 <<< "$1"; }

check_status() {
  local label="$1"
  local response="$2"
  local expected="$3"
  local actual
  actual=$(http_status "${response}")
  local body
  body=$(http_body "${response}")
  if [[ "${actual}" == "${expected}" ]]; then
    pass "${label} → HTTP ${actual}"
    if [[ "${VERBOSE}" == "1" ]]; then
      echo "    $(echo "${body}" | jq -c '.' 2>/dev/null || echo "${body}")" >&2
    fi
  else
    fail "${label} → expected HTTP ${expected}, got HTTP ${actual}"
    if [[ "${VERBOSE}" == "1" || "${actual}" == 4* || "${actual}" == 5* ]]; then
      echo "    $(echo "${body}" | jq -c '.' 2>/dev/null || echo "${body}")" >&2
    fi
  fi
  echo "${body}"
}

# ── 1. Health check ───────────────────────────────────────────────────────────
section "1. Health check"

resp=$(curl -s -w '\n%{http_code}' "${BASE_URL}/health")
body=$(http_body "${resp}")
status=$(http_status "${resp}")
if [[ "${status}" == "200" ]]; then
  pass "GET /health → 200 OK"
  if [[ "${VERBOSE}" == "1" ]]; then echo "    ${body}" >&2; fi
else
  fail "GET /health → HTTP ${status} (is the stack running?)"
  echo -e "\n${RED}Backend is not reachable at ${BASE_URL}. Start the stack and retry.${RESET}" >&2
  exit 1
fi

# ── 2. Auth endpoints ─────────────────────────────────────────────────────────
section "2. Auth endpoints"

# 2a. Unauthenticated /me → 401 (explicitly no cookie)
resp=$(curl -s -w '\n%{http_code}' "${API}/auth/me")
status=$(http_status "${resp}")
if [[ "${status}" == "401" ]]; then
  pass "GET /auth/me without session → 401 UNAUTHORIZED"
else
  fail "GET /auth/me without session → expected 401, got ${status}"
fi

# 2b. If a session is provided, verify /me returns the user object
if [[ -n "${SESSION_COOKIE}" ]]; then
  resp=$(check_status "GET /auth/me with session" "$(api_get '/auth/me')" "200")
  ME_ID=$(echo "${resp}" | jq -r '.id // empty' 2>/dev/null)
  ME_LOGIN=$(echo "${resp}" | jq -r '.login // empty' 2>/dev/null)
  if [[ -n "${ME_ID}" ]]; then
    pass "User object has id=${ME_ID} login=${ME_LOGIN}"
  else
    fail "GET /auth/me response missing id field"
  fi

  # 2c. OAuth redirect endpoint accessible (no-auth, just check it returns 302)
  REDIRECT_STATUS=$(curl -s -o /dev/null -w '%{http_code}' --max-redirs 0 "${API}/auth/github/login")
  if [[ "${REDIRECT_STATUS}" == "302" || "${REDIRECT_STATUS}" == "301" ]]; then
    pass "GET /auth/github/login → ${REDIRECT_STATUS} redirect"
  else
    fail "GET /auth/github/login → expected 302, got ${REDIRECT_STATUS}"
  fi
else
  skip "GET /auth/me with session (no TARS_SESSION set)"
  skip "GET /auth/github/login redirect check"
  echo -e "  ${YELLOW}Set TARS_SESSION=<cookie> to run authenticated tests${RESET}" >&2
  echo -e "\n${YELLOW}Remaining tests require authentication. Exiting.${RESET}" >&2
  summarise
  exit 0
fi

# ── 3. Repo CRUD ──────────────────────────────────────────────────────────────
section "3. Repo CRUD"

# Ensure a local git repo exists at REPO_PATH (server-side test)
# The repo path is used only as metadata — actual disk presence isn't validated at create time
SMOKE_REPO_PATH="${REPO_PATH}"
SMOKE_REPO_NAME="smoke-test-$(date +%s)"

# 3a. Create repo
resp=$(api_post "/repos" "{
  \"name\": \"${SMOKE_REPO_NAME}\",
  \"github_url\": \"https://github.com/smoke/test\",
  \"path\": \"${SMOKE_REPO_PATH}\"
}")
body=$(check_status "POST /repos (create)" "${resp}" "201")
REPO_ID=$(echo "${body}" | jq -r '.id // empty' 2>/dev/null)
if [[ -n "${REPO_ID}" ]]; then
  pass "Repo created with id=${REPO_ID}"
else
  fail "Repo creation response missing id"
  REPO_ID=""
fi

# 3b. List repos
resp=$(api_get "/repos")
body=$(check_status "GET /repos (list)" "${resp}" "200")
REPO_COUNT=$(echo "${body}" | jq '.repos | length' 2>/dev/null)
if [[ -n "${REPO_COUNT}" && "${REPO_COUNT}" -ge 1 ]]; then
  pass "GET /repos returns ${REPO_COUNT} repo(s)"
else
  fail "GET /repos returned empty or invalid repos array"
fi

if [[ -z "${REPO_ID}" ]]; then
  fail "No repo ID — skipping remaining repo/task/agent tests"
  summarise
  exit 1
fi

# 3c. Get repo
resp=$(api_get "/repos/${REPO_ID}")
body=$(check_status "GET /repos/${REPO_ID}" "${resp}" "200")
GOT_NAME=$(echo "${body}" | jq -r '.name // empty' 2>/dev/null)
if [[ "${GOT_NAME}" == "${SMOKE_REPO_NAME}" ]]; then
  pass "GET repo returns correct name"
else
  fail "GET repo name mismatch: got '${GOT_NAME}', expected '${SMOKE_REPO_NAME}'"
fi

# 3d. Update repo
resp=$(api_patch "/repos/${REPO_ID}" '{"default_branch": "develop"}')
body=$(check_status "PATCH /repos/${REPO_ID}" "${resp}" "200")
UPDATED_BRANCH=$(echo "${body}" | jq -r '.default_branch // empty' 2>/dev/null)
if [[ "${UPDATED_BRANCH}" == "develop" ]]; then
  pass "PATCH repo updates default_branch"
else
  fail "PATCH repo default_branch mismatch: got '${UPDATED_BRANCH}'"
fi

# 3e. Conflict: create duplicate name
resp=$(api_post "/repos" "{
  \"name\": \"${SMOKE_REPO_NAME}\",
  \"github_url\": \"https://github.com/smoke/test2\",
  \"path\": \"/tmp/other\"
}")
status=$(http_status "${resp}")
if [[ "${status}" == "409" ]]; then
  pass "POST /repos duplicate name → 409 CONFLICT"
else
  fail "POST /repos duplicate name → expected 409, got ${status}"
fi

# ── 4. Task CRUD ──────────────────────────────────────────────────────────────
section "4. Task CRUD"

# 4a. Create task
resp=$(api_post "/repos/${REPO_ID}/tasks" '{
  "title": "Smoke test task",
  "description": "Created by smoke-test.sh",
  "priority": 2
}')
body=$(check_status "POST /repos/${REPO_ID}/tasks" "${resp}" "201")
TASK_ID=$(echo "${body}" | jq -r '.id // empty' 2>/dev/null)
TASK_STATUS=$(echo "${body}" | jq -r '.status // empty' 2>/dev/null)
if [[ -n "${TASK_ID}" ]]; then
  pass "Task created with id=${TASK_ID} status=${TASK_STATUS}"
else
  fail "Task creation response missing id"
  TASK_ID=""
fi

# 4b. List tasks
resp=$(api_get "/repos/${REPO_ID}/tasks")
body=$(check_status "GET /repos/${REPO_ID}/tasks" "${resp}" "200")
TASK_COUNT=$(echo "${body}" | jq '.tasks | length' 2>/dev/null)
if [[ -n "${TASK_COUNT}" && "${TASK_COUNT}" -ge 1 ]]; then
  pass "GET tasks returns ${TASK_COUNT} task(s)"
else
  fail "GET tasks returned empty or invalid tasks array"
fi

if [[ -n "${TASK_ID}" ]]; then
  # 4c. Get task
  resp=$(api_get "/repos/${REPO_ID}/tasks/${TASK_ID}")
  body=$(check_status "GET /repos/${REPO_ID}/tasks/${TASK_ID}" "${resp}" "200")
  GOT_TITLE=$(echo "${body}" | jq -r '.title // empty' 2>/dev/null)
  if [[ "${GOT_TITLE}" == "Smoke test task" ]]; then
    pass "GET task returns correct title"
  else
    fail "GET task title mismatch: '${GOT_TITLE}'"
  fi

  # 4d. Update task status
  resp=$(api_patch "/repos/${REPO_ID}/tasks/${TASK_ID}" '{"status": "in_progress"}')
  body=$(check_status "PATCH task status → in_progress" "${resp}" "200")
  NEW_STATUS=$(echo "${body}" | jq -r '.status // empty' 2>/dev/null)
  if [[ "${NEW_STATUS}" == "in_progress" ]]; then
    pass "PATCH task updates status to in_progress"
  else
    fail "PATCH task status mismatch: '${NEW_STATUS}'"
  fi

  # 4e. Invalid status → 400
  resp=$(api_patch "/repos/${REPO_ID}/tasks/${TASK_ID}" '{"status": "invalid_status"}')
  status=$(http_status "${resp}")
  if [[ "${status}" == "400" ]]; then
    pass "PATCH task invalid status → 400 VALIDATION_ERROR"
  else
    fail "PATCH task invalid status → expected 400, got ${status}"
  fi
fi

# ── 5. Auth guard checks ──────────────────────────────────────────────────────
section "5. Auth guard checks"

# Protected routes must return 401 without a session
for path in "/repos" "/repos/fake_id" "/repos/fake_id/tasks" "/repos/fake_id/agents"; do
  status=$(curl -s -o /dev/null -w '%{http_code}' "${API}${path}")
  if [[ "${status}" == "401" ]]; then
    pass "GET ${path} without auth → 401"
  else
    fail "GET ${path} without auth → expected 401, got ${status}"
  fi
done

# ── 6. Events API ─────────────────────────────────────────────────────────────
section "6. Events API"

resp=$(api_get "/repos/${REPO_ID}/events?limit=10")
body=$(check_status "GET /repos/${REPO_ID}/events" "${resp}" "200")
HAS_MORE=$(echo "${body}" | jq -r '.has_more // empty' 2>/dev/null)
EVENTS_ARR=$(echo "${body}" | jq '.events' 2>/dev/null)
if [[ "${EVENTS_ARR}" != "null" && -n "${EVENTS_ARR}" ]]; then
  EVENTS_COUNT=$(echo "${body}" | jq '.events | length' 2>/dev/null)
  pass "Events API returns array (${EVENTS_COUNT} event(s)), has_more=${HAS_MORE}"
else
  fail "Events API missing events array"
fi

# Pagination params accepted without error
resp=$(api_get "/repos/${REPO_ID}/events?limit=5&type=agent.spawned")
status=$(http_status "${resp}")
if [[ "${status}" == "200" ]]; then
  pass "GET events with type filter → 200"
else
  fail "GET events with type filter → ${status}"
fi

# ── 7. Presence API ───────────────────────────────────────────────────────────
section "7. Presence API"

resp=$(api_get "/repos/${REPO_ID}/presence")
body=$(check_status "GET /repos/${REPO_ID}/presence" "${resp}" "200")
PRESENCE_REPO=$(echo "${body}" | jq -r '.repo_id // empty' 2>/dev/null)
PRESENCE_USERS=$(echo "${body}" | jq '.users' 2>/dev/null)
if [[ "${PRESENCE_REPO}" == "${REPO_ID}" && "${PRESENCE_USERS}" != "null" ]]; then
  pass "Presence endpoint returns repo_id and users array"
else
  fail "Presence endpoint response malformed (repo_id='${PRESENCE_REPO}')"
fi

# ── 8. Agent listing ──────────────────────────────────────────────────────────
section "8. Agent listing"

resp=$(api_get "/repos/${REPO_ID}/agents")
body=$(check_status "GET /repos/${REPO_ID}/agents" "${resp}" "200")
AGENTS_ARR=$(echo "${body}" | jq '.agents' 2>/dev/null)
if [[ "${AGENTS_ARR}" != "null" && -n "${AGENTS_ARR}" ]]; then
  pass "GET agents returns array"
else
  fail "GET agents missing agents array"
fi

# 8b. Spawn agent — skipped unless SMOKE_SPAWN_AGENT=1 (requires real git repo + Claude CLI)
if [[ "${SMOKE_SPAWN_AGENT:-0}" == "1" ]]; then
  section "8b. Agent spawn (live)"

  AGENT_NAME="smoke-agent-$(date +%s)"
  resp=$(api_post "/repos/${REPO_ID}/agents" "{
    \"name\": \"${AGENT_NAME}\",
    \"persona\": \"general\",
    \"model\": \"claude-opus-4-5\"
  }")
  body=$(check_status "POST /repos/${REPO_ID}/agents (spawn)" "${resp}" "201")
  AGENT_ID=$(echo "${body}" | jq -r '.id // empty' 2>/dev/null)
  AGENT_STATUS=$(echo "${body}" | jq -r '.status // empty' 2>/dev/null)
  if [[ -n "${AGENT_ID}" ]]; then
    pass "Agent spawned id=${AGENT_ID} status=${AGENT_STATUS}"

    # Wait briefly for agent to reach running/stopped
    sleep 3

    # Get logs
    resp=$(api_get "/repos/${REPO_ID}/agents/${AGENT_ID}/logs?limit=50")
    body=$(check_status "GET agent logs" "${resp}" "200")
    LINES_COUNT=$(echo "${body}" | jq '.lines | length' 2>/dev/null)
    pass "Agent logs returned ${LINES_COUNT} line(s)"

    # Get diff (agent just started, may be empty)
    resp=$(api_get "/repos/${REPO_ID}/agents/${AGENT_ID}/diff")
    status=$(http_status "${resp}")
    if [[ "${status}" == "200" ]]; then
      pass "GET agent diff → 200"
      diff_body=$(http_body "${resp}")
      FILES_CHANGED=$(echo "${diff_body}" | jq '.stats.files_changed // 0' 2>/dev/null)
      pass "Agent diff stats.files_changed=${FILES_CHANGED}"
    else
      fail "GET agent diff → ${status}"
    fi

    # Stop agent
    resp=$(api_post "/repos/${REPO_ID}/agents/${AGENT_ID}/stop")
    body=$(check_status "POST agent stop" "${resp}" "200")
    STOPPED_STATUS=$(echo "${body}" | jq -r '.status // empty' 2>/dev/null)
    if [[ "${STOPPED_STATUS}" == "stopped" ]]; then
      pass "Agent stopped successfully"
    else
      fail "Agent stop status: '${STOPPED_STATUS}'"
    fi
  else
    fail "Agent spawn response missing id"
    AGENT_ID=""
  fi
else
  skip "Agent spawn/stop (set SMOKE_SPAWN_AGENT=1 to enable — requires git repo + Claude CLI on server)"
fi

# ── 9. WebSocket connectivity ─────────────────────────────────────────────────
section "9. WebSocket connectivity"

if command -v websocat &>/dev/null; then
  # Test WS upgrade + ping/pong
  WS_URL="${BASE_URL/http/ws}/ws"
  PONG=$(echo '{"type":"ping","payload":{}}' | \
    websocat -n1 --header "Cookie: tars_session=${SESSION_COOKIE}" "${WS_URL}" 2>/dev/null | \
    jq -r '.type // empty' 2>/dev/null)
  if [[ "${PONG}" == "pong" ]]; then
    pass "WebSocket ping → pong received"
  else
    fail "WebSocket ping → no pong (got: '${PONG}')"
  fi
else
  skip "WebSocket ping/pong (websocat not installed — 'cargo install websocat' or 'brew install websocat')"
fi

# ── 10. Frontend reachability ─────────────────────────────────────────────────
section "10. Frontend reachability"

FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
FRONTEND_STATUS=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "${FRONTEND_URL}" 2>/dev/null || echo "000")
if [[ "${FRONTEND_STATUS}" == "200" || "${FRONTEND_STATUS}" == "302" ]]; then
  pass "Frontend reachable at ${FRONTEND_URL} → ${FRONTEND_STATUS} (auth redirect expected)"
elif [[ "${FRONTEND_STATUS}" == "000" ]]; then
  skip "Frontend not reachable at ${FRONTEND_URL} (not started or different port)"
else
  fail "Frontend at ${FRONTEND_URL} → HTTP ${FRONTEND_STATUS}"
fi

# SPA routing: /login should return 200 (SPA fallback)
LOGIN_STATUS=$(curl -s -o /dev/null -w '%{http_code}' --max-time 5 "${FRONTEND_URL}/login" 2>/dev/null || echo "000")
if [[ "${LOGIN_STATUS}" == "200" ]]; then
  pass "Frontend SPA route /login → 200 (SPA fallback working)"
elif [[ "${LOGIN_STATUS}" == "000" ]]; then
  skip "Frontend SPA route /login (frontend not reachable)"
else
  fail "Frontend SPA route /login → ${LOGIN_STATUS} (expected 200)"
fi

# ── 11. Cleanup ───────────────────────────────────────────────────────────────
section "11. Cleanup"

if [[ -n "${TASK_ID}" ]]; then
  resp=$(api_delete "/repos/${REPO_ID}/tasks/${TASK_ID}")
  status=$(http_status "${resp}")
  if [[ "${status}" == "204" ]]; then
    pass "DELETE task → 204"
    # Verify 404 on re-fetch
    resp=$(api_get "/repos/${REPO_ID}/tasks/${TASK_ID}")
    status=$(http_status "${resp}")
    if [[ "${status}" == "404" ]]; then
      pass "GET deleted task → 404 (confirmed gone)"
    else
      fail "GET deleted task → ${status} (expected 404)"
    fi
  else
    fail "DELETE task → ${status}"
  fi
fi

if [[ -n "${REPO_ID}" ]]; then
  resp=$(api_delete "/repos/${REPO_ID}")
  status=$(http_status "${resp}")
  if [[ "${status}" == "204" ]]; then
    pass "DELETE repo → 204"
    # Verify 404 on re-fetch
    resp=$(api_get "/repos/${REPO_ID}")
    status=$(http_status "${resp}")
    if [[ "${status}" == "404" ]]; then
      pass "GET deleted repo → 404 (confirmed gone)"
    else
      fail "GET deleted repo → ${status} (expected 404)"
    fi
  else
    fail "DELETE repo → ${status}"
  fi
fi

summarise
