#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

OWNER_REPO="${GITHUB_REPOSITORY:-falconfan123/Go-mall}"
DEFAULT_BRANCH="${DEFAULT_BRANCH:-main}"
REQUIRED_CHECKS=("${@:-Build Quality Integration}")

if [[ "${#REQUIRED_CHECKS[@]}" -eq 1 && "${REQUIRED_CHECKS[0]}" == "Build Quality Integration" ]]; then
  REQUIRED_CHECKS=(Build Quality Integration)
fi

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_cmd gh
require_cmd python3

gh auth status >/dev/null

repo_id="$(gh api "repos/${OWNER_REPO}" --jq '.id')"
current_ruleset_id="$(
  gh api "repos/${OWNER_REPO}/rulesets" --jq '.[] | select(.name == "main-ci-gate") | .id' 2>/dev/null || true
)"

ruleset_payload="$(python3 - "$repo_id" "$DEFAULT_BRANCH" "${REQUIRED_CHECKS[@]}" <<'PY'
import json
import sys

repo_id = int(sys.argv[1])
branch = sys.argv[2]
checks = sys.argv[3:]

payload = {
    "name": "main-ci-gate",
    "target": "branch",
    "enforcement": "active",
    "bypass_actors": [],
    "conditions": {
        "ref_name": {
            "include": [f"refs/heads/{branch}"],
            "exclude": [],
        }
    },
    "rules": [
        {
            "type": "pull_request",
            "parameters": {
                "dismiss_stale_reviews_on_push": False,
                "require_code_owner_review": False,
                "require_last_push_approval": False,
                "required_approving_review_count": 0,
                "required_review_thread_resolution": False,
            },
        },
        {
            "type": "required_status_checks",
            "parameters": {
                "strict_required_status_checks_policy": True,
                "required_status_checks": [
                    {"context": check, "integration_id": 15368}
                    for check in checks
                ],
            },
        },
        {"type": "deletion"},
        {"type": "non_fast_forward"},
    ],
}

print(json.dumps(payload))
PY
)"

gh api \
  --method PATCH \
  -H "Accept: application/vnd.github+json" \
  "repos/${OWNER_REPO}" \
  -f allow_auto_merge=true \
  -f delete_branch_on_merge=true >/dev/null

if [[ -n "${current_ruleset_id}" ]]; then
  gh api \
    --method PUT \
    -H "Accept: application/vnd.github+json" \
    "repos/${OWNER_REPO}/rulesets/${current_ruleset_id}" \
    --input - <<<"${ruleset_payload}" >/dev/null
  echo "updated ruleset main-ci-gate on ${OWNER_REPO}:${DEFAULT_BRANCH}"
else
  gh api \
    --method POST \
    -H "Accept: application/vnd.github+json" \
    "repos/${OWNER_REPO}/rulesets" \
    --input - <<<"${ruleset_payload}" >/dev/null
  echo "created ruleset main-ci-gate on ${OWNER_REPO}:${DEFAULT_BRANCH}"
fi
