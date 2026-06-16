#!/bin/bash
# =============================================================================
# Go-mall Master Gatekeeper Script
# 验证 AI 提交的代码是否真正合格
# 信脚本不信 AI —— 脚本过了才算过
# =============================================================================

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
PASS=0
FAIL=0

pass() { echo -e "  ${GREEN}✅ $1${NC}"; PASS=$((PASS+1)); }
fail() { echo -e "  ${RED}❌ $1${NC}"; FAIL=$((FAIL+1)); }
warn() { echo -e "  ${YELLOW}⚠️  $1${NC}"; }

echo -e "${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Go-mall Gatekeeper — 主守门人检查${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════${NC}"
echo ""

# ── 1. 编译检查 ──────────────────────────────────────────────
echo -e "${BLUE}[1/7] 编译检查${NC}"
if make build 2>&1; then
  pass "全部服务编译通过"
else
  fail "编译失败 — 请修复后再提交"
fi
echo ""

# ── 2. 静态分析 / Lint ───────────────────────────────────────
echo -e "${BLUE}[2/7] 静态分析 (make lint)${NC}"
if make lint 2>&1; then
  pass "Lint 检查通过"
else
  fail "Lint 失败 — 请运行 make fmt 或 ./scripts/check.sh --auto-fix"
fi
echo ""

# ── 3. 单元测试 + 覆盖率 ─────────────────────────────────────
echo -e "${BLUE}[3/7] 单元测试 + 覆盖率${NC}"
COVERAGE_THRESHOLD=63.6
if make test-unit 2>&1; then
  pass "单元测试全部通过"
  if COVERAGE_OUTPUT=$(COVERAGE_THRESHOLD="$COVERAGE_THRESHOLD" make coverage-ci 2>&1); then
    printf '%s\n' "$COVERAGE_OUTPUT"
    COV=$(printf '%s\n' "$COVERAGE_OUTPUT" | awk -F= '/TOTAL_COVERAGE=/{print $2}' | tail -n 1)
    if [ -n "$COV" ]; then
      pass "覆盖率 ${COV}% ≥ ${COVERAGE_THRESHOLD}%"
    else
      warn "覆盖率检查通过，但未解析到 TOTAL_COVERAGE"
    fi
  else
    printf '%s\n' "$COVERAGE_OUTPUT"
    fail "覆盖率检查失败 — 请运行 make coverage-ci"
  fi
else
  fail "单元测试失败"
fi
echo ""

# ── 4. Mock 完整性 ────────────────────────────────────────────
echo -e "${BLUE}[4/7] Mock 完整性${NC}"
if make mock 2>&1; then
  pass "Mock 生成成功"
  # 检查是否有未提交的 mock 变更
  UNCOMMITTED_MOCKS=$(git diff --name-only -- '**/mock*.go' 2>/dev/null || true)
  if [ -n "$UNCOMMITTED_MOCKS" ]; then
    warn "以下 mock 文件有未提交变更："
    echo "$UNCOMMITTED_MOCKS" | sed 's/^/    /'
  fi
else
  fail "Mock 生成失败 — 检查 mockgen 配置"
fi
echo ""

# ── 5. 硬编码密钥检查 ─────────────────────────────────────────
echo -e "${BLUE}[5/7] 敏感信息扫描${NC}"
SENSITIVE=$(rg -n 'sk-[A-Za-z0-9]{20,}' \
  --glob '*.go' --glob '*.yaml' --glob '*.json' --glob '*.env*' \
  --glob '!vendor/**' --glob '!**/pb/**' --glob '!.git/**' --glob '!frontend/node_modules/**' . || true)
if [ -z "$SENSITIVE" ]; then
  pass "未发现硬编码密钥"
else
  fail "发现可能硬编码的密钥："
  echo "$SENSITIVE" | head -5 | sed 's/^/    /'
fi
echo ""

# ── 6. 调试代码检查 ───────────────────────────────────────────
echo -e "${BLUE}[6/7] 调试代码检查${NC}"
DEBUG=$(rg -n '\bfmt\.Print(f|ln)?\b|\bprintln\b' \
  --glob '*.go' \
  --glob '!**/*_test.go' \
  --glob '!vendor/**' \
  --glob '!**/pb/**' \
  --glob '!.git/**' \
  --glob '!frontend/node_modules/**' \
  --glob '!test/**' \
  --glob '!cmd/**' \
  --glob '!tools/**' \
  --glob '!run.go' \
  --glob '!services/*/*.go' . || true)
if [ -z "$DEBUG" ]; then
  pass "未发现调试输出残留"
else
  fail "发现可能的调试输出："
  echo "$DEBUG" | head -10 | sed 's/^/    /'
fi
echo ""

# ── 7. TODO/FIXME 审计 ────────────────────────────────────────
echo -e "${BLUE}[7/7] TODO/FIXME 审计${NC}"
TODOS=$(rg -n '//.*\b(TODO|FIXME|HACK|XXX|BUG)\b|/\*.*\b(TODO|FIXME|HACK|XXX|BUG)\b' \
  --glob '*.go' --glob '*.jsx' --glob '*.js' \
  --glob '!vendor/**' --glob '!**/pb/**' --glob '!.git/**' --glob '!**/node_modules/**' . || true)
TODO_COUNT=$(echo "$TODOS" | grep -c . || true)
if [ "$TODO_COUNT" -eq 0 ]; then
  pass "零 TODO/FIXME 残留"
elif [ "$TODO_COUNT" -le 5 ]; then
  warn "发现 $TODO_COUNT 个 TODO/FIXME（建议清理但非强制）："
  echo "$TODOS" | head -5 | sed 's/^/    /'
else
  warn "发现 $TODO_COUNT 个 TODO/FIXME（建议清理）："
  echo "$TODOS" | head -10 | sed 's/^/    /'
fi
echo ""

# ── 汇总 ──────────────────────────────────────────────────────
echo -e "${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Gatekeeper 结果汇总${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "  ${GREEN}通过: $PASS${NC}"
echo -e "  ${RED}失败: $FAIL${NC}"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo -e "${RED}❌ Gatekeeper 未通过 — 请修复后再提交${NC}"
  exit 1
else
  echo -e "${GREEN}✅ Gatekeeper 全部通过 — 代码合格${NC}"
fi
