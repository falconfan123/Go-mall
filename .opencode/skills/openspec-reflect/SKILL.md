---
name: openspec-reflect
description: Optional holistic re-review of a complete OpenSpec change with the @openspec-reviewer agent — since 2026-09-08 the single mandatory review gate lives at the END of /opsx-propose (whole-package review after all artifacts exist). Use this skill only when artifacts were modified after propose (e.g. /opsx-update, manual edits) or the user explicitly asks to re-check a change before apply.
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: Go-mall (adapted from Peng Qian's reflection harness)
  version: "1.1"
---

# OpenSpec Reflect（可选整体复查）

对**已完成 propose（含其末尾单次整体审核）的 change** 提供可选的整体复查通道。**2026-09-08 起不再是主流程必做环节**——单次整体审核已并入 `/opsx-propose` 末尾（propose 生成全部制品后审一次，🔴 清零即冻结）。若 propose 之后又跑一次本 skill，等于对同一 change 审第二次，正常流程**不要**这样做。

## 触发时机（仅这些情况用）

- 制品在 propose 冻结后被修改过：`/opsx-update` 调整、用户手工编辑制品 → apply 前想确认改完仍一致
- 用户明确要求 apply 前对完整 change 再做一次整体检查（"反思一下 / reflect / 复查一遍"）
- 备选通道场景：propose 时没走反思（例如旧 change 的制品早已生成、未经过审核），补一次审核

## 执行步骤

1. **确认变更存在且制品齐全**：
   ```bash
   openspec status --change "<name>"
   ```
   制品不全（缺 proposal/design/specs/tasks 任一）→ 先补，不进入整体复查。

2. **运行整体复查**（调起 `@openspec-reviewer` 只读子智能体，审查完整变更包；备选：bash 脚本）：
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
- **勿重复审核**：propose 已含单次整体审核，除非触发时机命中，否则不主动调本 skill。
