#!/usr/bin/env bash
# =============================================================================
# openspec-review.sh — OpenSpec 提案反思审核（Reflection Harness）
#
# 实现自彭骞知乎回答《别再让愚蠢的spec文件降低你的代码质量了》：
#   - 反思智能体 = opencode run headless（只读 agent: openspec-review-headless，
#     model: opencode-go/kimi-k2.6），与主智能体使用不同模型，提供独立视角（reflection 模式）
#   - 作用位置: /opsx-propose 之后、/opsx-apply 之前（可多轮循环）
#   - 按 🔴 阻断 / 🟡 应修复 / 💡 建议 分级，冻结判据 = "### 🔴 遗留" 为空
#
# 用法:
#   ./openspec/reviewer/openspec-review.sh <change-name> [options]
#
#   options:
#     --log            把本轮审核结果追加到 openspec/changes/<name>/review-log.md
#     --batch <file>   只审查单个制品（proposal.md / design.md / tasks.md /
#                      specs/<cap>/spec.md），默认审查整个 change 目录全部制品
#     -m <model>       覆盖审查模型（默认 opencode-go/kimi-k2.6，格式 provider/model）
#
# 退出码:
#   0 = 无 🔴 遗留（本批可冻结）   1 = 存在 🔴（需修复后重审）
#   2 = 用法错误 / change 不存在
# =============================================================================
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OPENCODE="${OPENCODE_BIN:-opencode}"
REVIEW_AGENT="${REVIEW_AGENT:-openspec-review-headless}"
REVIEW_MODEL="${REVIEW_MODEL:-opencode-go/kimi-k2.6}"

if [[ $# -lt 1 ]]; then
  echo "用法: $0 <change-name> [--log] [--batch <file>] [-m <model>]" >&2
  exit 2
fi

CHANGE_NAME="$1"; shift
CHANGE_DIR="$ROOT/openspec/changes/$CHANGE_NAME"
if [[ ! -d "$CHANGE_DIR" ]]; then
  echo "❌ 找不到变更目录: ${CHANGE_DIR}（先 openspec propose 生成提案）" >&2
  exit 2
fi

LOG=0
BATCH=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --log) LOG=1; shift ;;
    --batch) BATCH="$2"; shift 2 ;;
    -m|--model) REVIEW_MODEL="$2"; shift 2 ;;
    *) echo "未知参数: $1" >&2; exit 2 ;;
  esac
done

# ---------- 收集制品文件（按依赖序: proposal -> design -> specs -> tasks） ----------
FILES=()
if [[ -n "$BATCH" ]]; then
  if [[ -f "$CHANGE_DIR/$BATCH" ]]; then
    FILES=("$CHANGE_DIR/$BATCH")
  else
    echo "❌ 制品不存在: $CHANGE_DIR/$BATCH" >&2
    exit 2
  fi
else
  for f in proposal.md design.md tasks.md; do
    [[ -f "$CHANGE_DIR/$f" ]] && FILES+=("$CHANGE_DIR/$f")
  done
  while IFS= read -r f; do FILES+=("$f"); done \
    < <(find "$CHANGE_DIR/specs" -name 'spec.md' -type f 2>/dev/null | sort)
  [[ -f "$CHANGE_DIR/.openspec.yaml" ]] && FILES+=("$CHANGE_DIR/.openspec.yaml")
fi

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "❌ 变更目录里没有任何制品文件: $CHANGE_DIR" >&2
  exit 2
fi

# ---------- 轮次计数（review-log.md 已有"审查轮次"段落数 + 1） ----------
ROUND=1
if [[ -f "$CHANGE_DIR/review-log.md" ]]; then
  ROUND=$(( $(grep -c '^## 审查轮次' "$CHANGE_DIR/review-log.md" 2>/dev/null || echo 0) + 1 ))
fi
if [[ $ROUND -gt 5 ]]; then
  echo "⚠️ 本批已累计 $((ROUND-1)) 轮仍未通过（MAX_ROUNDS=5）——停止自动循环，交人工决策:"
  echo "   1) Mandatory freeze（用户接受现状 → 强制冻结本批，进下一批）"
  echo "   2) Fallback design（缩小范围 → 修改当前制品降级，重审，轮次重置）"
fi

# ---------- 构造审查提示 ----------
REL_FILES=()
for f in "${FILES[@]}"; do REL_FILES+=("${f#"$ROOT"/}"); done

PROMPT=$(cat <<EOF
你是一名 **OpenSpec 变更审查员**——一个具备批判性思维、聚焦实质问题的审计者。
你的任务是在 OpenSpec 变更进入实施之前，审查变更中的每一份产出物，找出那些会真正导致实施失败或返工的问题。

## 核心原则：区分实质性缺陷和形式问题
**实质性缺陷 = 会导致实施方向错误、遗漏关键场景、产生矛盾、或无法验收的问题。**
**形式问题 = 不影响实施质量的格式/写法差异。**
你的首要任务是找出前者。后者可以提，但必须标注为可选建议且放在末尾。

## 定位
你工作在 /opsx-propose 和 /opsx-apply 之间的阶段：explore → /opsx-propose → ⬅ 你在这里（可能多次） → /opsx-apply → verify → archive
规范尚未冻结。实施尚未开始。你的使命是：**在代码被写出来之前，找出真正会导致返工或事故的缺陷**。因为发现规范错误只需几分钟，而修正错误代码则需要数小时。

## 本次审查范围
- 变更目录: openspec/changes/$CHANGE_NAME/（相对仓库根）
- 本轮待审制品: ${REL_FILES[*]}
- 基线: openspec/specs/ 下的主 specs（若存在，先阅读再评判增量）
- 若存在 explore-brief.md，以它为需求基线 checklist；若存在 review-log.md，参考历史审核结论，不要重复报告已确认修复的问题

## 执行步骤
1. 阅读 openspec/specs/ 基线（了解现有系统契约）。
2. 阅读本轮待审制品全文。
3. 对照基线逐条检查：需求是否完整可验收？与已冻结前序制品是否一致？有没有遗漏边界/异常场景？有没有自相矛盾？任务是否可追踪到每个 requirement？场景是否可用 Given/When/Then 验证？
4. 输出分级结论。

## 原则
- **有建设性且严格。** 每一个问题不仅要指出「是什么」，更要说明「为什么会导致返工或事故」。
- **具体，不笼统。** 精确指出文件路径、需求名称、任务编号。
- **按严重度分级。** 🔴 阻断 vs 🟡 应修复 vs 💡 建议，不混淆。
- **有上下文意识。** 对照现有系统（openspec/specs/）进行评判，而非脱离实际。
- **只读。** 绝不修改文件。你照亮问题，主智能体执行修复。

## 需要避免的反模式
- 盖章放行：没有深度审视就说「看起来不错！」
- 吹毛求疵：只顾格式问题，忽略架构性缺陷。
- 急于给方案：在用户认可问题存在之前就跳到「应该这样修复」。
- 忽略现有规范：不了解基线就审查增量。
- 模糊反馈：「这里可以更好」——要说清楚具体哪里、为什么。

## 输出格式（必须严格遵守）
以 Markdown 输出，包含以下小节，顺序不变：

### 🔴 遗留（阻断，必须全部修复后才能冻结本批）
- 每条格式: [相对文件路径] 问题描述（以及为什么会导致返工或事故）
- 没有则写: 无

### 🟡 应修复（不影响冻结，但建议本轮修复）
- 每条格式同上

### 💡 建议（可选）
- 每条格式同上

### 已确认无问题的点
- 简要列出你检查过且确认无问题的方面（说明覆盖范围）
EOF
)

echo "🔍 审查 openspec/changes/$CHANGE_NAME/（第 $ROUND 轮）— 审查员: opencode run [$REVIEW_AGENT / $REVIEW_MODEL] (read-only)"
echo "   制品: ${REL_FILES[*]}"
echo ""

OUTPUT="$(cd "$ROOT" && "$OPENCODE" run --agent "$REVIEW_AGENT" -m "$REVIEW_MODEL" "$PROMPT" 2>&1)"
echo "$OUTPUT"

# ---------- 写 review-log ----------
if [[ $LOG -eq 1 ]]; then
  {
    echo ""
    echo "## 审查轮次 $ROUND"
    echo ""
    echo "_审查时间: $(date '+%Y-%m-%d %H:%M:%S') | 审查员: opencode run [$REVIEW_AGENT / $REVIEW_MODEL] (read-only) | 范围: ${REL_FILES[*]}_"
    echo ""
    echo '```'
    echo "$OUTPUT"
    echo '```'
  } >> "$CHANGE_DIR/review-log.md"
  echo ""
  echo "📝 已追加到 review-log.md（第 $ROUND 轮）"
fi

# ---------- 冻结判据: 最后一个 "### 🔴 遗留" 小节为空/为"无" ----------
# 注意: opencode run 输出不回显 prompt，但模型可能读取 review-log.md（含旧轮次 🔴 头），
# 必须取最后一个 🔴 小节（模型最终结论在最后）
if echo "$OUTPUT" | grep -q '^### 🔴 遗留'; then
  LAST="$(echo "$OUTPUT" | grep -n '^### 🔴 遗留' | tail -1 | cut -d: -f1)"
  # 从最后一个 🔴 头开始，跳过头行，遇到下一个 "### " 小节头停止
  SECTION="$(echo "$OUTPUT" | tail -n +"$LAST" | awk 'NR==1{next} /^### /{exit} {print}')"
  if echo "$SECTION" | grep -q '^[[:space:]]*无[[:space:]]*$'; then
    echo ""
    echo "✅ 本批 🔴 清零 → 可冻结（按批次顺序进入下一制品，或进入 /openspec-reflect 整体复查）"
    exit 0
  else
    echo ""
    echo "❌ 存在 🔴 遗留问题 → 本批不能冻结。主智能体修复后重跑本脚本。"
    exit 1
  fi
else
  echo ""
  echo "⚠️ 审查输出未找到 '### 🔴 遗留' 小节（按文章规则视为通过），请人工确认输出格式合规"
  exit 0
fi
