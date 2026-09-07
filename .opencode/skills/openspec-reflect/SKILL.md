---
name: openspec-reflect
description: Run the reflection phase on a complete OpenSpec change — after all artifacts are proposed and batch-frozen, do a final holistic review across proposal/design/specs/tasks with the @openspec-reviewer agent, and optionally update review-log.md. Use after /openspec-propose completes all batches, before /openspec-apply-change.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: Go-mall (adapted from Peng Qian's reflection harness)
  version: "1.0"
---

# OpenSpec Reflect（反思阶段）

对**已完成全部批次冻结的 change** 做一次**整体复查**。这是 propose 内逐批反思之后的最后一道整体门禁（对应主流程 Explore → Propose → **Reflect** → Apply → Archive 中的 Reflect 环节）。

## 触发时机

- 用户在 propose 全部批次完成后说"反思一下 / reflect / 复查一遍"
- 或 apply 之前要求对完整 change 做整体检查

## 执行步骤

1. **确认变更存在且制品齐全**：
   ```bash
   openspec status --change "<name>"
   ```
   制品不全（缺 proposal/design/specs/tasks 任一）→ 先补，不进入整体复查。

2. **运行整体反思审核**（调起 `@openspec-reviewer` 只读子智能体，审查完整变更包；备选：bash 脚本调 Codex read-only）：
   ```bash
   ./openspec/reviewer/openspec-review.sh "<name>" --log
   ```

3. **处理审核结果**（状态机同 workflow skill，MAX_ROUNDS = 5）：
   - 🔴 存在 → 主智能体修复对应制品 → 重跑步骤 2，直至 🔴 清零或轮次超限转人工
   - 🔴 清零 → 记录结论到 review-log.md（脚本已带 --log）
   - 超 5 轮 → 交用户：Mandatory freeze（强制放行）/ Fallback design（缩小范围重审）

4. **整体一致性检查**（人工视角补充，子智能体可能漏的）：
   - proposal 的每个能力点是否在 specs 里有对应 requirement？
   - design 是否覆盖 specs 全部 requirement？
   - tasks 是否可追溯到 specs 全部 requirement？
   - 有没有超出 proposal Non-goals 的越界设计？

5. **输出结论**：
   - ✅ 通过 → 告知用户可进入 `/openspec-apply-change`
   - 🔴 未清零 → 列明剩余问题，等待用户决策

## Guardrails

- **只读反思**：审核由 `@openspec-reviewer` 只读执行（edit/write/bash deny）；主智能体只修改制品文件，不碰业务代码。
- **不越权 apply**：反思通过 ≠ 开始实施。等待用户明确要求进入 apply。
- **review-log 必记**：每轮审核结论必须追加到 review-log.md（防上下文丢失）。
