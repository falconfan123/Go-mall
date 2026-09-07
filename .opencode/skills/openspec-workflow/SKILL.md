---
name: openspec-workflow
description: Mandatory prerequisite for ALL OpenSpec operations — load this BEFORE any openspec-* skill. Use when running /opsx-apply, /opsx-propose, /opsx-verify, /opsx-archive, /opsx-explore, /opsx-sync; running `openspec status`, `openspec list`, `openspec instructions`; reading files under `openspec/changes/`; or doing any OpenSpec stage (propose, apply, verify, archive, explore, sync).
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: Go-mall (adapted from Peng Qian's openspec-workflow)
  version: "1.0"
---

# OpenSpec Workflow（强制）

**所有代码变更必须先有提案，后写代码。**

## 流程

1. **Explore**
   - 用户说"想想"、"讨论"、"explore"时，只讨论不编码。
   - 需求讨论结束若产生 change 需求 → 落盘 `explore-brief.md`（checklist baseline）。
2. **Propose**
   - 创建 `openspec/changes/<name>/` 下的提案文件（proposal → design → specs → tasks 逐个生成，逐个过反思审核，通过后冻结）。
   - 每批制品冻结时写入 `review-log.md`。
3. **Reflect**（独立整体复查）
   - 全部批次冻结后，运行 `/openspec-reflect` 对完整变更包做整体反思（调 Codex read-only）。
   - 🔴 清零 → 通过；超 5 轮 → 交人工决策（Mandatory freeze / Fallback design）。
4. **Apply**
   - 按提案任务实施，**没有提案不得修改任何文件**。
5. **Verify**
   - 实现完成后验证，检查实现与提案的一致性。
6. **Archive**
   - 归档变更，delta specs 合并回 `openspec/specs/`。

## 硬性规则

- **bug修复无需编辑或创建提案**，此硬性规则只对功能变更或新增功能时有效
- **无提案不变更**：如果用户要求修改代码，必须先确认已有对应提案，或先创建提案
- **无提案不编辑**：编辑文件前，检查当前是否在 `openspec/changes/` 下有匹配的变更目录
- **Explore 模式不编码**：当用户在 explore 模式下，**禁止**创建提案、**禁止**编辑文件、**禁止**写测试
- **变更完成后**：必须执行 verify 流程，检查实现与提案的一致性

## 反思流程（Reflection Harness）

**提案制品在进入 apply 之前，必须经过反思智能体审核。** 反思智能体 = `@openspec-reviewer` 子智能体（OpenCode 原生，只读：edit/write/bash 全部 deny；与主智能体使用不同模型，提供独立视角）。备选通道：`bash ./openspec/reviewer/openspec-review.sh <change-name>`（调 Codex read-only，用于手动/CI 场景）。

### 反思循环

1. **每批制品创建后**都**必须**调起 `@openspec-reviewer` 对该批次内的制品文件进行审核（备选：运行反思审核脚本）：
   ```bash
   ./openspec/reviewer/openspec-review.sh <change-name>
   ```
2. 主智能体根据审核反馈修改**当前批次的制品文件**。
3. 再次运行反思审核，对**当前批次的制品文件**进行审核。
4. **审查通过标准**：
   - 4a. **单轮通过原则**：当前轮次审查后，若该轮次的 "### 🔴 遗留" 不存在或为空，进入下一批次。
   - 4b. **修复循环**：若当前轮仍有 🔴 问题 → 主智能体修复 → 下一轮审查 → 回到 4a。重复直至通过或触发 4c。
   - 4c. **硬上界保护**（MAX_ROUNDS = 5）：同一批次累计审查轮次达到 5 轮仍未按 4a 通过 → 停止循环，交由人工决策。

### 批次顺序（逐个生成 + 冻结）

制品**逐个生成、逐个审核、通过后冻结**，严禁一次生成全部制品再统一审核：

1. `proposal.md` → 反思审核 → 🔴 清零 → **冻结**（后续不得随意修改，修改需说明理由）
2. `design.md` → 反思审核（对照已冻结的 proposal）→ 🔴 清零 → **冻结**
3. `specs/<capability>/spec.md` → 反思审核（对照 proposal + design）→ 🔴 清零 → **冻结**
4. `tasks.md` → 反思审核（对照全部已冻结制品）→ 🔴 清零 → **冻结**

> 为什么要这样：一次性审核 N 个文件时，修复 A 文件会让 B 文件不一致，一致性验证成本是 N×N 组合爆炸。逐个冻结后，每轮审核只需检查「新制品 vs 已冻结制品」一组对比。

### 反思日志

每轮审核结果追加记录到 `openspec/changes/<change-name>/review-log.md`，保证多轮反思的连续性（防上下文丢失）：
```bash
./openspec/reviewer/openspec-review.sh <change-name> --log
```
（脚本会把当前轮次的审核输出追加到 review-log.md）

### 反思状态机（round 计数 + 人工决策）

单个批次内的反思循环严格按以下状态机执行（对应文章图 2）：

```
Start reviewing the batch
        │
        ▼
     round += 1        ← round 从 review-log.md 中该批次已有的"审查轮次"计数 +1 得出
        │
        ▼
    round > 5? ──Yes──▶ Human decision（停止自动化，交用户）
        │No                ├─▶ Mandatory freeze（用户接受现状 → 强制冻结本批，进下一批）
        ▼                  └─▶ Fallback design（用户要求缩小范围 → 修改当前制品降级，重审）
  @openspec-reviewer
        │
        ▼
   Check for 🔴? ──No──▶ Batch freeze（冻结本批）──▶ 下一批次 / 全部完成
        │Yes
        ▼
Primary agent repair issues（修 🔴）──▶ 回到 round += 1 重审
```

**round 计数规则**：
- 每批次的 round 从 `review-log.md` 中该批次已有的审查轮次数 +1 得出（`grep -c '^## 审查轮次'` 脚本已实现）
- 同一批次内累计超过 5 轮仍未通过 → **停止自动化循环**，向用户报告当前状态并给两个选项：
  1. **Mandatory freeze（强制冻结）**：用户认可当前制品质量可接受 → 标记"人工强制冻结"写入 review-log，进入下一批次
  2. **Fallback design（降级设计）**：用户认为范围过大/设计过复杂 → 修改当前制品缩小范围，然后重审（round 重置）
- 用户未选择前，不得自动继续循环，也不得跳过该批次

**进入下一批次的判定**：
- 当前批次 🔴 清零 → 冻结 → 进入下一批次（proposal → design → specs → tasks 顺序）
- 所有批次冻结完成 → 反思阶段结束，进入 apply

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

explore 阶段的需求讨论必须落盘为 `openspec/changes/<change-name>/explore-brief.md`，作为提案审核的持续性 checklist 基线（防会话上下文丢失）。简单需求未经过 explore 讨论时可不生成。

## 违规处理

如果检测到以下情况，立即停止并提醒用户：
- 没有提案就开始修改代码
- 在 explore 模式下编辑文件
- 修改了不属于当前提案范围的文件
