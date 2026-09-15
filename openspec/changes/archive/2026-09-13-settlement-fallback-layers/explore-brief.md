# Explore Brief — settlement-fallback-layers

> 探索落盘（2026-09-05）。本文是后续 proposal/design/specs/tasks 过反思审核时的**需求基线**，
> 审核方须以本 Checklist 逐项核对，防止细节在上下文丢失后走样。
> 前置 change：harden-payment-settlement-chain（Saga 化，已实施 22/23）。本 change 是其
> 明确声明的"后续增强"（探针/对账兜底），落地后**推翻该 change 的一条已接受残余风险**
> （"PAID 成功但建 saga 失败且 Stripe 不重试时无自动自愈，人工承接"——由 A 废除）。

## 背景 — 现状调查结论（含实证）

### 残余缺口全景（前序 change 收尾时的遗留清单）

| # | 失败模式 | 现状兜底 | 实证出处 |
|---|---|---|---|
| ② | payment PAID 落库 → EnsureSaga 之间窗口丢失（payment.go:194 裸 Update 无事务；崩溃/重启 + Stripe 不重试 = 永久悬空） | Stripe 重试 + 快路径 EnsureSaga（机会性） | design 决策 1 附注"已接受残余风险" |
| ③ | saga failed 已补偿 → 订单回 PendingPayment + payment PAID（gid 墓碑挡住重试，Stripe 重试永远 500） | error 告警 → 纯人工 | design 决策 1 附注 |
| ④ | delay 弃单（3 次重试 Reject(false)）→ 订单该关没关 | 弃单告警 → 纯人工 | explore-brief C4（第一周期） |
| ⑤ | webhook 永久丢失：Stripe 已扣款、本地 payment 未 PAID（本地两库自洽，无信号） | **无**（完全盲区） | 本轮新识别 |
| ⑥ | 本地 PAID、Stripe 无此账 | 无 | 本轮新识别（告警勿发货级） |

### 关键实证（本轮）

- `payment.go:194`：`PaymentModel.Update` 为裸单行更新（无事务）——A 的同事务写入是 payment
  服务第一次本地多写事务（需 model WithSession 变体，参照 inventory 先例）。
- payments 表已有全部 join 键：order_id/pre_order_id/user_id/transaction_id(varchar255)/
  status/paid_at；Stripe 对账金额单位一致（cny、分），`normalizeStripeTransactionID`（≤64）
  即 join 键。
- order/delay 延迟 = 30min（mq/delay/init.go:24）→ scan2 窗口须 > 30min（定 35min）。
- dtm HTTP query API 现成：`GET /api/dtmsvr/query?gid=`（dtm v1.19.0 api_http.go:37）——
  扫描器查 saga 状态**不需要**碰 dtm 库表。
- delay consumer 的关单业务（行锁+状态机 Created→Closed+ReleaseCoupon+ReturnPreInventory）
  目前内联在 handler 中，无独立 RPC——scan2 复用需抽业务入口（实现细节归 design）。
- `PaymentStatus_FULLY_REFUNDED(4)` 枚举早已预留但全仓无退款实现——人工层是空头支票。
- Stripe SDK 已接（stripe-go v81，APIBackend + MaxNetworkRetries，processor.go）——C 的
  Sessions/BalanceTransaction list 是自然扩展。
- 时序事实：harden-payment-settlement-chain 仍 active（22/23，探针⑤待人工复核），其 spec
  **尚未归档进主 specs**（premature 同步已删）。本 change 的能力 delta 以归档后主 spec 为基准。

## 目标 / 非目标

### 目标（用户视角）

1. 洞②由"依赖 Stripe 施舍重试"升级为**落盘必达**：payment PAID 与结算待办同事务落库，
   relay 接力调 EnsureSaga。
2. 洞③④由"纯人工"升级为**定时扫描自愈/分流**：孤儿单自动补结算（含墓碑单有限自动重试）、
   超龄单自动关单，仍失败才转人工。
3. 洞⑤⑥由"盲区"升级为**外部账本裁决**：Stripe 日终双向对账，丢单自动重放（信任决策已拍板），
   空账告警勿发货。
4. 一切自动化的尽头有**人工产出**：告警聚合去重 + 异常单待办产出（不做退款实现）。

### 非目标（明确排除）

1. **退款实现另立项**（Stripe Refund API + FULLY_REFUNDED 状态流转）——本次只做异常单产出。
2. 不复活 PaymentDelayMQ（其"超时对账"设计意图被 B 扫描器实质取代）。
3. 不做 dtm 多实例/高可用改造（保持 compose 单实例；扫描器自带容忍）。
4. 不改既有幂等机制（状态机/barrier/确定性 gid/two-state key 全部保持）。
5. C 不引入新的对账数据源（只对 Stripe；不加银行/微信/支付宝通道）。

## 待决策点（均已拍板，2026-09-05）

| # | 决策点 | 结论 |
|---|---|---|
| D1 | 范围 | **三板斧一个 change**：A Outbox + B 扫描器（scan1/scan2）+ C Stripe 对账 + 人工产出层 |
| D2 | scan1 墓碑单 | **自动递增 gid 重试**：查 dtm status=failed → 新 gid `settle:{orderId}:r{n}`（上限 3），仍失败告警转人工 |
| D3 | C 方向1（Stripe 有账本地无） | **自动重放** webhook 处理逻辑（安全垫：支付状态机 + EnsureSaga 幂等 + outbox/扫描）；重放失败转告警 |
| D4 | 人工层 | **只做产出**（告警聚合去重 + 异常单待办），退款另立项 |
| D5 | 时序 | **先归档 harden-payment-settlement-chain → 再 propose 本 change**（delta 以归档后主 spec 为基准） |
| D6 | scanner/relay 部署位置 | **倾向 payment 服务内常驻监督循环**（relay 与 scan 同宿主；payment 已有 webhook HTTP + 常驻循环先例），独立 reconciler cmd 留为备选——design 阶段定 |
| D7 | 人工待办落点 | 倾向异常单表（可查可关）+ 结构化告警，纯日志留为备选——design 阶段定 |

## 设计要点（逐层）

### A 事务性发件箱（治②）

- 新表 `payment_outbox`（mall 库，与 payments 同库事务）：event_id/order_id/payload/
  status(pending|done|dead)/attempts/next_retry_at。
- webhook 同事务：`TransactCtx(UPDATE payments PAID + INSERT outbox pending)`——payment 服务
  首个本地多写事务。
- relay：payment 服务内监督循环（5~10s tick），`FOR UPDATE SKIP LOCKED` 批量取 pending，
  调 **EnsureSaga**（幂等已备），成功置 done；失败指数退避（2^n）+ attempts 上限 → dead + 告警。
- **done 语义 = 已提交给编排器**（不等于结算完成——saga 异步，职责到 EnsureSaga 为止）。
- 快路径（Stripe 重试）保持不动：outbox 系统性保证，重试机会性保证，两层叠加。

### B 定时扫描器（治③④）

- scan1（paid_at < now-10min，10min 给正常链路留窗口）：
  - order PendingPayment → 查 dtm status：无 saga → EnsureSaga；in-flight → 干等；failed →
    递增后缀 gid 重试（r2/r3，上限 3）→ 仍失败告警转人工。
  - order Closed/Cancelled → **不补结算**（补偿语义正确）→ 人工产出（退款队列）。
- scan2（created_at < now-35min，> 30min delay）：复用关单业务（行锁+状态机），与 delay
  consumer 竞态由既有状态机语义收敛（关了但钱到 → 变成 scan1 的 Closed+PAID 分支 → 退款产出）。
- scan3（可选增强）：dtm in-flight 超龄告警。
- 告警去重：同一 orderId 三层（outbox dead/scan1/scan2/对账）重复发现必须聚合。

### C Stripe 对账（治⑤⑥）

- 日终拉 Checkout Sessions list（窗口按 created），join 本地 transaction_id。
- 方向1（Stripe 有、本地无 PAID）→ **自动重放** webhook 处理逻辑；失败 → 告警人工。
- 方向2（本地 PAID、Stripe 无）→ 告警 + 冻结（勿自动回滚——结算已发生，乱回滚放大伤害）。

### 人工层

- 异常单产出：退款队列性质的待办（Closed+PAID、对账方向2、重试超限单），聚合去重。
- 退款实现（Stripe Refund API + FULLY_REFUNDED 流转）**另立项**。

## Checklist — 提案审核基线

- [ ] **C1 outbox 同事务**：payment PAID 翻转与 outbox pending 写入必须同一事务（TransactCtx
      + model WithSession 变体）；不存在"翻转了但无待办"的窗口。
- [ ] **C2 relay 幂等接力**：relay 调 EnsureSaga（不重复造幂等）；指数退避 + attempts 上限 →
      dead + error 告警；relay 自身崩溃重启可续扫（SKIP LOCKED + 状态落库）。
- [ ] **C3 done 语义**：outbox done = 已提交编排器，不得等待/依赖 saga 终态。
- [ ] **C4 scan1 分流**：PendingPayment → 补结算（tombstone 递增 gid r2/r3 上限 3）；Closed/
      Cancelled → 人工产出不补结算。时间阈值 10min 必须存在（防误扫正常单）。
- [ ] **C5 scan2 竞态**：关单复用既有状态机（Created→Closed 才执行），窗口 35min > 30min
      delay；与迟到支付的竞态语义与 delay consumer 一致。
- [ ] **C6 对账双向**：方向1 自动重放（失败转告警）；方向2 告警冻结不自动回滚。
- [ ] **C7 人工层只产出**：不做退款实现（Non-goal）。
- [ ] **C8 告警去重**：同一异常单跨层重复发现必须聚合（orderId 维度）。
- [ ] **C9 能力增补**：async-settlement-reliability 增加 outbox 必达 + 扫描自愈需求（delta 于
      归档后的主 spec）。
- [ ] **C10 时序前提**：propose 前须先归档 harden-payment-settlement-chain（前提：探针⑤人工复核）。
- [ ] **C11 测试**：mock 先行；关键场景——relay 崩溃续扫、outbox dead 路径、scan1 tombstone
      递增、scan2 与迟到支付竞态、C 重放幂等、对账差额生成。
- [ ] **C12 DDL**：payment_outbox（+人工待办表如采用 D7 表方案）走 construct/depend/sql 目录。
- [ ] **C13 调度窗口可配置**：扫描间隔/时间阈值/重试上限进 yaml，不写死。
