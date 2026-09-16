# Review Log · fix-govulncheck-grpc-cve

## 审查轮次 1 — 2026-09-16

### 🔴 遗留
🔴 无（冻结）。

### 🟡 修复（本轮一次修完）
1. tasks 1.1 模块范围：`tools/rag` 不含 grpc 依赖 → 明确"仅 18 个含 grpc 的模块升级，tools/rag 只验证不升级"（避免结构性无法满足验证）。
2. tasks 1.1 direct/indirect：`dal`/`services/gateway` 的 grpc 为 `// indirect`（go get 可升级，验证 require 行版本即可）。
3. tasks 1.6 回落需重跑：明确"回落 v1.83.1 后必须重跑 1.1–1.5 并附新输出"。

### 已确认无问题
- 影响面 18 模块与仓库 grep 事实一致；go-ci-vulncheck.sh 无豁免机制（真升级）；Non-goals 三处一致；skip_specs 合理；回滚可执行。

## 审查轮次 2 — 2026-09-16

### 🔴 遗留
🔴 无（冻结）。

### 🟡 修复（本轮一次修完）
1. tasks 1.3 计数修正：17 → **18 模块（go_ci_modules.txt 全部）**（含 tools/rag），与 proposal"全部模块"一致。
2. tasks 1.3 验证字段补 `GO-2026-6372`（与"两个 CVE 均清"一致）。
3. explore-brief C1/C2 补 amqp091-go / GO-2026-6372 条目（需求基线同步）。
4. tasks 1.1 补 amqp091-go indirect 验证（require 行保留 v1.13.0 且 tidy 不回落）。

### 💡 修复
1. proposal 调用点表述修正为"经 RabbitMQEventPublisher.publish → go-queue/rabbitmq 间接到达"。
2. tasks 加 1.9a：实施期间新 CVE 按同口径纳入并更新漏洞矩阵。

### 目标重述（本 round）
本变更目标 = `scripts/go-ci-vulncheck.sh` 在全部模块 exit 0（当前漏洞库快照）；Non-goal = 不升与本目标无关的依赖。
