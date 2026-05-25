#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

REPO="${GITHUB_REPOSITORY:-falconfan123/Go-mall}"
BASE_BRANCH="${BASE_BRANCH:-main}"
COMMIT_MESSAGE="${MSG:-${1:-}}"
BRANCH_NAME="${BRANCH_NAME:-auto/submit-$(date +%Y%m%d-%H%M%S)}"
ALLOW_TMP_FILES="${SUBMIT_WITH_CI_ALLOW_TMP_FILES:-0}"
POLL_INTERVAL="${SUBMIT_WITH_CI_POLL_INTERVAL:-10}"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

require_clean_index() {
  if ! git diff --cached --quiet; then
    echo "staged changes detected; please unstage or commit them before running submit-with-ci" >&2
    exit 1
  fi
}

detect_tmp_files() {
  python3 - <<'PY'
import os
import subprocess
import sys

tracked = subprocess.check_output(["git", "ls-files"], text=True).splitlines()
others = subprocess.check_output(["git", "ls-files", "--others", "--exclude-standard"], text=True).splitlines()
hits = []
for path in tracked + others:
    base = os.path.basename(path)
    if base.startswith(".tmp"):
        hits.append(path)
for item in hits:
    print(item)
PY
}

run_preflight() {
  echo "running preflight checks"
  go work sync
  make test-unit
  make ci-build
}

ensure_commit_message() {
  if [[ -z "${COMMIT_MESSAGE}" ]]; then
    echo "commit message is required: MSG=\"...\" make submit-ci or scripts/submit-with-ci.sh \"...\"" >&2
    exit 1
  fi
}

create_pr_body() {
  local body_file="$1"
  cat >"${body_file}" <<EOF
## Automated Submission

- Source: \`scripts/submit-with-ci.sh\`
- Base branch: \`${BASE_BRANCH}\`
- Local branch: \`${BRANCH_NAME}\`
- Commit: \`${COMMIT_MESSAGE}\`

## Local Preflight

- \`go work sync\`
- \`make test-unit\`
- \`make ci-build\`

This PR was created automatically and auto-merge has been enabled. Merge will happen only after required GitHub checks pass.
EOF
}

wait_for_required_checks() {
  local pr_ref="$1"
  local output

  while true; do
    set +e
    output="$(gh pr checks "${pr_ref}" --required --watch=false --json name,bucket,link 2>&1)"
    local status=$?
    set -e

    if [[ ${status} -eq 0 ]]; then
      echo "${output}" | python3 - <<'PY'
import json
import sys
checks = json.load(sys.stdin)
for item in checks:
    print(f"{item['name']}: {item['bucket']}")
PY
      return 0
    fi

    if [[ ${status} -eq 8 ]]; then
      echo "required checks still pending; waiting ${POLL_INTERVAL}s"
      sleep "${POLL_INTERVAL}"
      continue
    fi

    echo "failed to read PR checks:" >&2
    echo "${output}" >&2
    return "${status}"
  done
}

ensure_pr_merged() {
  local pr_ref="$1"
  local state merged_at
  state="$(gh pr view "${pr_ref}" --json state --jq '.state')"
  merged_at="$(gh pr view "${pr_ref}" --json mergedAt --jq '.mergedAt')"
  if [[ "${state}" != "MERGED" && "${merged_at}" == "null" ]]; then
    echo "PR is not merged after checks completed" >&2
    return 1
  fi
}

sync_local_main() {
  local original_branch="$1"
  git fetch origin
  if git show-ref --verify --quiet "refs/heads/${BASE_BRANCH}"; then
    git switch "${BASE_BRANCH}"
    git reset --hard "origin/${BASE_BRANCH}"
    git switch "${original_branch}"
  fi
}

require_cmd git
require_cmd gh
require_cmd go
require_cmd make
require_cmd python3

gh auth status >/dev/null
ensure_commit_message
require_clean_index

if [[ "${ALLOW_TMP_FILES}" != "1" ]]; then
  tmp_hits="$(detect_tmp_files)"
  if [[ -n "${tmp_hits}" ]]; then
    echo "temporary files detected; refusing automatic submission:" >&2
    echo "${tmp_hits}" >&2
    echo "override with SUBMIT_WITH_CI_ALLOW_TMP_FILES=1 only if these files are intentional" >&2
    exit 1
  fi
fi

if [[ -z "$(git status --porcelain)" ]]; then
  echo "no local changes to submit" >&2
  exit 1
fi

original_branch="$(git branch --show-current)"
git fetch origin

run_preflight

git switch -c "${BRANCH_NAME}"
git add -A
git commit -m "${COMMIT_MESSAGE}"
git rebase "origin/${BASE_BRANCH}"
git push -u origin "${BRANCH_NAME}"

body_file="$(mktemp)"
trap 'rm -f "${body_file}"' EXIT
create_pr_body "${body_file}"

pr_url="$(
  gh pr create \
    --repo "${REPO}" \
    --base "${BASE_BRANCH}" \
    --head "${BRANCH_NAME}" \
    --title "${COMMIT_MESSAGE}" \
    --body-file "${body_file}"
)"

echo "created PR: ${pr_url}"

gh pr merge "${pr_url}" --repo "${REPO}" --auto --squash --delete-branch
wait_for_required_checks "${pr_url}"
ensure_pr_merged "${pr_url}"
sync_local_main "${original_branch}"

echo "submission completed and merged: ${pr_url}"
