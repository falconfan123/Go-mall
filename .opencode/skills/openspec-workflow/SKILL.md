---
name: openspec-workflow
description: Mandatory prerequisite for ALL OpenSpec operations — load this BEFORE any openspec-* skill. Use when running /opsx-apply, /opsx-propose, /opsx-verify, /opsx-archive, /opsx-explore, /opsx-sync; running `openspec status`, `openspec list`, `openspec instructions`; reading files under `openspec/changes/`; or doing any OpenSpec stage (propose, apply, verify, archive, explore, sync).
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: Go-mall (adapted from Peng Qian's openspec-workflow)
  version: "1.1"
---

# OpenSpec Workflow（强制）

**所有代码变更必须先有提案，后写代码。**

## 流程

1. **Explore**
   - 用户说"想想"、"讨论"、"explore"时，只讨论不编码。
   - 需求讨论结束若产生 change 需求 → 落盘 `explore-brief.md`（checklist baseline）。
2. **Propose**
   - 创建 `openspec/changes/<name>/` 下的提案文件（proposal → specs → design → tasks，按 schema 依赖序逐个生成，生成时对照已生成制品保持一致性；**不逐批审核**）。
   - 全部制品齐备后，做**一次整体反思审核**（`@openspec-reviewer` 审完整变更包，2026-09-08 起单门禁策略，见下"反思流程"）→ 🔴 清零 → 全部冻结 → 写入 `review-log.md`。
3. **Apply**
   - 按提案任务实施，**没有提案不得修改任何文件**。propose 末尾已做过整体审核；若 apply 前制品被手工或 /opsx-update 修改过，可先跑 `/openspec-reflect` 补一次整体复查（可选，非必做）。
4. **Verify**
   - 实现完成后验证，检查实现与提案的一致性。
5. **Archive**
   - 归档变更，delta specs 合并回 `openspec/specs/`。

## 硬性规则

- **bug修复无需编辑或创建提案**，此硬性规则只对功能变更或新增功能时有效
- **无提案不变更**：如果用户要求修改代码，必须先确认已有对应提案，或先创建提案
- **无提案不编辑**：编辑文件前，检查当前是否在 `openspec/changes/` 下有匹配的变更目录
- **Explore 模式不编码**：当用户在 explore 模式下，**禁止**创建提案、**禁止**编辑文件、**禁止**写测试
- **变更完成后**：必须执行 verify 流程，检查实现与提案的一致性

## 反思流程（Reflection Harness — 单门禁，每个 change 只审一次）

**提案制品在进入 apply 之前，必须且只需经过一次整体审核。** 反思智能体 = `@openspec-reviewer` 子智能体（OpenCode 原生，只读：edit/write/bash 全部 deny；与主智能体使用不同模型，提供独立视角）。备选通道：`bash ./openspec/reviewer/openspec-review.sh <change-name>`（opencode run headless 只读 agent，用于手动/CI 场景）。

> 2026-09-08 起从"逐批审核 + reflect 整体复查"改为**单门禁**：全部制品一次生成，然后只对完整变更包做一次整体审核（原逐批冻结 + 独立 Reflect 环节已废弃——审核次数太多，拖慢 propose）。修复循环只发生在这一道门禁内。

### 反思循环（单门禁）

1. **全部制品（proposal/specs/design/tasks）创建完成后**，调起 `@openspec-reviewer` 对完整变更包审核一次（备选：运行反思审核脚本）：
   ```bash
   ./openspec/reviewer/openspec-review.sh <change-name>
   ```
2. 主智能体根据审核反馈修改对应制品文件。
3. 有 🔴 则再次整体审核（同一门禁内循环），直到通过。
4. **审查通过标准**：
   - 4a. **单轮通过原则**：该轮次的 "### 🔴 遗留" 不存在或为空 → 全部制品冻结。
   - 4b. **修复循环**：若该轮仍有 🔴 问题 → 主智能体修复 → 下一轮审核 → 回到 4a。重复直至通过或触发 4c。
   - 4c. **硬上界保护**（MAX_ROUNDS = 5）：累计审查轮次达到 5 轮仍未按 4a 通过 → 停止循环，交由人工决策。

### 制品生成顺序（不逐批审核，但生成时保持一致性）

制品按 schema 依赖序逐个生成（以 `openspec status --change <name> --json` 的 requires 为准，官方顺序：proposal → specs → design → tasks）。每个制品生成时**对照已生成的制品与 explore-brief** 保证衔接一致；一致性校验统一交给末尾那一次整体审核，不在中途打断。

> 为什么不逐批审核了：逐批冻结每批都要调一次审核智能体（一个 change 通常 4 批 × 每批若干轮），加上末尾整体复查，审核次数过多、propose 被拖得很慢。单门禁只在全部制品齐备后审一次；有 🔴 才进入修复-复审循环，多数情况一次通过。

### 反思日志

每轮审核结果追加记录到 `openspec/changes/<change-name>/review-log.md`，保证多轮反思的连续性（防上下文丢失）：
```bash
./openspec/reviewer/openspec-review.sh <change-name> --log
```
（脚本会把当前轮次的审核输出追加到 review-log.md）

### 反思状态机（round 计数 + 人工决策）

单门禁内的反思循环严格按以下状态机执行（对应文章图 2，批次维度换成 change 维度）：

```
Start reviewing the change
        │
        ▼
     round += 1        ← round 从 review-log.md 已有"审查轮次"计数 +1 得出
        │
        ▼
    round > 5? ──Yes──▶ Human decision（停止自动化，交用户）
        │No                ├─▶ Mandatory freeze（用户接受现状 → 强制冻结，进 apply）
        ▼                  └─▶ Fallback design（用户要求缩小范围 → 修改制品降级，重审）
  @openspec-reviewer
        │
        ▼
   Check for 🔴? ──No──▶ Freeze all（冻结全部制品）──▶ 进入 apply
        │Yes
        ▼
Primary agent repair issues（修 🔴）──▶ 回到 round += 1 重审
```

**round 计数规则**：
- round 从 `review-log.md` 中已有的审查轮次数 +1 得出（`grep -c '^## 审查轮次'` 脚本已实现）
- 累计超过 5 轮仍未通过 → **停止自动化循环**，向用户报告当前状态并给两个选项：
  1. **Mandatory freeze（强制冻结）**：用户认可当前制品质量可接受 → 标记"人工强制冻结"写入 review-log，进入 apply
  2. **Fallback design（降级设计）**：用户认为范围过大/设计过复杂 → 修改制品缩小范围，然后重审（round 重置）
- 用户未选择前，不得自动继续循环，也不得跳过审核

**进入 apply 的判定**：
- 整体审核 🔴 清零（或人工强制冻结）→ 全部制品冻结 → 反思阶段结束，进入 apply

### 越权处理

反思智能体（`@openspec-reviewer`）权限 deny edit/write/bash，**不可能修改文件**。审核反馈的修复由主智能体执行。

## 提案创建要求

每次创建变更必须包含：
- `proposal.md` - 变更原因和范围
- `design.md` - 设计方案
- `specs/<capability>/spec.md` - 行为规格（delta）
- `tasks.md` - 具体任务清单
- `.openspec.yaml` - 变更元数据

附加要求：
- **任务粒度**：`tasks.md` 中每个任务不超过 2 小时

## Explore 落盘

explore 阶段的需求讨论必须落盘为 `openspec/changes/<change-name>/explore-brief.md`，作为整体审核的 checklist 基线（防会话上下文丢失）。简单需求未经过 explore 讨论时可不生成。

## 违规处理

如果检测到以下情况，立即停止并提醒用户：
- 没有提案就开始修改代码
- 在 explore 模式下编辑文件
- 修改了不属于当前提案范围的文件
