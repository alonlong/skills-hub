#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
PASS=0
FAIL=0
NAMESPACE_SLUG="smoke-$(date +%s)"
SKILL_SLUG="email-helper"
VERSION="1.0.0"
ARCHIVE_PATH="$(mktemp "/tmp/skillhub-smoke-archive-XXXXXX.zip")"

cleanup() {
  rm -f "$ARCHIVE_PATH"
}

trap cleanup EXIT

pass() {
  echo "PASS: $1"
  PASS=$((PASS + 1))
}

fail() {
  echo "FAIL: $1"
  FAIL=$((FAIL + 1))
}

check_status() {
  local desc="$1"
  local actual="$2"
  local expected="$3"
  if [[ "$actual" == "$expected" ]]; then
    pass "$desc (HTTP $actual)"
  else
    fail "$desc (expected $expected, got $actual)"
  fi
}

ADMIN_USERNAME="${BOOTSTRAP_ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${BOOTSTRAP_ADMIN_PASSWORD:-ChangeMe!2026}"

json_field() {
  local json="$1"
  local expr="$2"
  JSON_INPUT="$json" python3 - "$expr" <<'PY'
import json
import os
import sys

expr = sys.argv[1]
value = json.loads(os.environ["JSON_INPUT"])
for part in expr.split("."):
    if part.isdigit():
        value = value[int(part)]
    else:
        value = value[part]
if isinstance(value, (dict, list)):
    print(json.dumps(value))
else:
    print(value)
PY
}

create_skill_archive() {
  python3 - "$ARCHIVE_PATH" <<'PY'
import json
import sys
import zipfile

archive_path = sys.argv[1]
with zipfile.ZipFile(archive_path, "w") as zf:
    zf.writestr("manifest.json", json.dumps({
        "version": "1.0.0",
        "displayName": "Email Helper",
        "summary": "Send and summarize email",
    }))
    zf.writestr("SKILL.md", "# Email Helper\n\nSmoke test package.\n")
PY
}

echo "=== SkillHub Smoke Test ==="
echo "Target: $BASE_URL"
echo "Namespace: $NAMESPACE_SLUG"
echo

HEALTH_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 10 -s -o /dev/null -w "%{http_code}" "$BASE_URL/healthz" || true)"
status="$HEALTH_STATUS"; check_status "Health endpoint" "$HEALTH_STATUS" "200"

UNAUTHORIZED_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 10 -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/auth/me" || true)"
status="$UNAUTHORIZED_STATUS"; check_status "Auth endpoint requires a token" "$UNAUTHORIZED_STATUS" "401"

LOGIN_RESPONSE="$(curl --retry 3 --retry-delay 1 --max-time 10 -sS \
  -H "Content-Type: application/json" \
  -X POST "$BASE_URL/api/v1/auth/login" \
  -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" || true)"
LOGIN_TOKEN="$(json_field "$LOGIN_RESPONSE" "accessToken" 2>/dev/null || true)"
if [[ -n "$LOGIN_TOKEN" ]]; then
  pass "Bootstrap admin can log in"
else
  fail "Bootstrap admin login failed"
fi

AUTH_HEADER=(-H "Authorization: Bearer $LOGIN_TOKEN")

ME_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 10 -s -o /dev/null -w "%{http_code}" "${AUTH_HEADER[@]}" "$BASE_URL/api/v1/auth/me" || true)"
status="$ME_STATUS"; check_status "Authenticated user can fetch /auth/me" "$ME_STATUS" "200"

CREATE_NAMESPACE_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 10 -s -o /tmp/skillhub-smoke-namespace.json -w "%{http_code}" \
  "${AUTH_HEADER[@]}" \
  -H "Content-Type: application/json" \
  -X POST "$BASE_URL/api/v1/namespaces" \
  -d "{\"slug\":\"$NAMESPACE_SLUG\",\"displayName\":\"Smoke Namespace\",\"type\":\"TEAM\",\"description\":\"Smoke test namespace\"}" || true)"
status="$CREATE_NAMESPACE_STATUS"; check_status "Create namespace" "$CREATE_NAMESPACE_STATUS" "201"

GET_NAMESPACE_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 10 -s -o /tmp/skillhub-smoke-get-namespace.json -w "%{http_code}" \
  "$BASE_URL/api/v1/namespaces/$NAMESPACE_SLUG" || true)"
status="$GET_NAMESPACE_STATUS"; check_status "Fetch namespace by slug" "$GET_NAMESPACE_STATUS" "200"

create_skill_archive
PUBLISH_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 20 -s -o /tmp/skillhub-smoke-publish.json -w "%{http_code}" \
  "${AUTH_HEADER[@]}" \
  -F "file=@$ARCHIVE_PATH;type=application/zip" \
  -X POST "$BASE_URL/api/v1/skills/$NAMESPACE_SLUG/$SKILL_SLUG/versions" || true)"
status="$PUBLISH_STATUS"; check_status "Publish skill version" "$PUBLISH_STATUS" "201"

DOWNLOAD_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 20 -s -o /tmp/skillhub-smoke-download.zip -w "%{http_code}" \
  "$BASE_URL/api/v1/skills/$NAMESPACE_SLUG/$SKILL_SLUG/versions/$VERSION/download" || true)"
status="$DOWNLOAD_STATUS"; check_status "Download published archive" "$DOWNLOAD_STATUS" "200"

SEARCH_STATUS="$(curl --retry 3 --retry-delay 1 --max-time 10 -s -o /tmp/skillhub-smoke-search.json -w "%{http_code}" \
  "$BASE_URL/api/v1/search?q=email" || true)"
status="$SEARCH_STATUS"; check_status "Search endpoint responds" "$SEARCH_STATUS" "200"

echo
echo "Results: $PASS passed, $FAIL failed"
if [[ "$FAIL" -ne 0 ]]; then
  exit 1
fi
