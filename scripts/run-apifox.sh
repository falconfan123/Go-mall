#!/bin/bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCENARIO_FILE="${1:-$PROJECT_ROOT/test/apifox/scenarios/go-mall-smoke.apifox-cli.json}"
REPORT_DIR="${2:-$PROJECT_ROOT/test/apifox/reports}"

if ! command -v apifox >/dev/null 2>&1; then
  echo "错误: 未找到 apifox CLI"
  exit 1
fi

if [ ! -f "$SCENARIO_FILE" ]; then
  echo "未找到场景文件: $SCENARIO_FILE"
  echo "先执行以下步骤："
  echo "1. node scripts/generate-apifox-openapi.mjs"
  echo "2. 在 Apifox App 中导入 test/apifox/go-mall-gateway.openapi.json"
  echo "3. 创建测试场景并导出为 Apifox CLI 格式到:"
  echo "   test/apifox/scenarios/go-mall-smoke.apifox-cli.json"
  echo "4. 重新运行本脚本"
  exit 1
fi

mkdir -p "$REPORT_DIR"
if [ $# -gt 0 ]; then
  shift
fi
if [ $# -gt 0 ]; then
  shift
fi

apifox run "$SCENARIO_FILE" \
  -r cli,html,json \
  --out-dir "$REPORT_DIR" \
  "$@"
