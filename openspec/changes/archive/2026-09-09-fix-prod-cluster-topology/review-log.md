# Review Log · fix-prod-cluster-topology

## 审查轮次 1（proposal）

### 🔴 遗留
1. **能力定义"零停机"泛化**（proposal.md Capabilities）——零停机仅承诺网关，其余服务允许资源受限下的短暂窗口；定义需限定。

### 🟡 建议（本轮已修复）
1. 未申明 hostNetwork 保留约束（dtm 单 BusiHost 回调依赖）→ 已补"保留约束"。
2. C5 缺实证数字记录 → 回归验证已补"记录可达性/延迟/QPS/错误率/连接中断计数"。
3. C6 缺回滚承诺 → Impact 已补"回滚"段。
4. 数量口径 12+ deployment vs ~12 pod → 统一为"~12 个业务服务 deployment（含网关双副本）"。
5. Why 补 hostNetwork 成因 → 已补一句前序变更背景。

### 复核（轮次 2 待跑）

## 审查轮次 2（proposal 复核）

🔴 无。上轮 1🔴 + 5🟡 全部确认修复（能力定义限定、hostNetwork 保留约束、实证数字、回滚、口径、成因）。proposal 冻结。

🟡 遗留建议：design 阶段须产出 R1/R2 资源取舍结论（网关双副本可行性 / 降级预案）；"连接中断计数=0"的观测手段留 design/tasks 落地。

## 审查轮次 3（specs）

### 审核中

## 审查轮次 3（specs）

🔴 无。5 条 🟡 全部修复（下单链路 Scenario、maxSurge/maxUnavailable 声明固定、R4 基础设施可达 Scenario、C6 回滚/数据保护 Requirement、coupons 分支）。复核 🔴 无，specs 冻结。补修 worker 字面措辞（"不承载业务服务负载"）。

🟡 遗留（design/tasks 落地）：① 下单链路 saga 终态判定窗口需明确时限；② "worker 仅承载基础设施"验收措辞按"不承载业务服务负载"执行。

## 审查轮次 4（design）

### 审核中

## 审查轮次 4（design）

🔴 1 项（postgres/redis 节点归属矛盾）→ 实证重核：postgres/redis/etcd 在 master，ES/rabbitmq/minio 在 worker；spec/explore-brief 措辞修正，design 正确。4 条 🟡 全部修复（gateway readinessProbe 前置、CPU 2C 压测边界、nodeSelector 标签验证留证、中断计数观测手段）。复核 🔴 无，design 冻结；顺手修残留 2 条（explore-brief D2 清单、design D2 补 minio）。

## 审查轮次 5（tasks）

### 审核中

## 审查轮次 5（tasks）

审核循环达 5 轮硬上界（脚本按 review-log 累计计数），交人工决策。上轮（Codex 备选通道）🔴 3 条 + 🟡 8 条已全部修复核对：
- 🔴 dtm 遗漏 RollingUpdate → 4.3 已纳入（含 dtm）
- 🔴 基础设施 endpoint 未具体化 → 2.3 补全各 Service IP:port
- 🔴 零停机缺可执行命令 → 5.1 给出 while-curl 探活 + awk 统计
- 🟡 nodeSelector 验证 / dtm 口径 / 非容量标准 / 文档路径 / gateway 地址解析 / 资源基线 / spec 映射 / 探活时长 → 全部补齐

**用户决策（2026-09-08）：Mandatory freeze** —— 接受当前 tasks.md 质量，人工强制冻结本批。全部制品（proposal/specs/design/tasks）冻结完成，进入 apply 前置状态。
