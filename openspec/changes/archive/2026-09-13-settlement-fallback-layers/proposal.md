# Proposal — settlement-fallback-layers

## Why

harden-payment-settlement-chain（已归档）把结算主干升级为幂等 Saga 后，链路可靠性缺口收敛为三类**静默失败**——它们没有链路内错误信号，幂等/重试/Saga 均无能为力，只有对账与扫描能兜住（explore-brief 实证）：

1. **PAID→EnsureSaga 窗口丢失**（洞②）：`payment.go` 的支付单翻转是裸单行更新（无事务），翻转成功但 saga 未发起时（进程崩溃/dtm 宕机）只能依赖 Stripe 重试；Stripe 不重试即永久悬空——归档 change design 决策 1 附注声明为"已接受残余风险（人工承接）"。
2. **无信号状态漂移**（洞③④）：saga failed 已补偿的墓碑单（订单回 PendingPayment + payment PAID，重试被确定性 gid 挡死）、delay 弃单后订单该关没关——现状纯人工。
3. **本地自洽的盲区**（洞⑤⑥）：webhook 永久丢失时 Stripe 已扣款而本地 payment 未 PAID，本地两库自洽无信号，只有外部账本知道真相；反向（本地有账 Stripe 无钱）同样只有对账能发现。

本次为结算链路补齐兜底三层 + 人工产出，使"支付成功的副作用最终必然发生"从**过程保证**（Saga）升级为**结果保证**（发件箱必达 + 扫描自愈 + 外部对账）。

## What Changes

- **A 事务性发件箱（治②）**：新增 `payment_outbox` 表（mall 库，与 payments 同库）；webhook 把"支付单状态翻转"与"结算待办写入"放进同一事务；payment 服务内 relay 监督循环以 `FOR UPDATE SKIP LOCKED` 批量接力调 EnsureSaga（幂等已备）；失败指数退避 + attempts 上限 → dead + 告警；**EnsureSaga 提交成功即置 done，不等待 saga 终态**（done = 已提交编排器，非结算完成）。**语义升级（C9）**：废除归档 change 两处"已接受残余风险"——design.md:52-53"PAID 成功但发起失败且 Stripe 不重试"由本项 outbox 必达替代；design.md:71"tombstone 纯人工承接"由 B scan1 递增 gid 自动重试替代（specs delta 显式声明）。
- **B 定时对账扫描器（治③④）**：scan1——`payments PAID + paid_at < now-10min` 联查订单分流：PendingPayment → 补结算（saga failed 墓碑单查 dtm status 后以递增后缀 gid `settle:{orderId}:r{n}` 有限自动重试，上限 3，仍失败告警转人工）；Closed/Cancelled → 不补结算（补偿语义正确），产出人工待办（退款另立项）。scan2——`orders PendingPayment + created_at < now-35min`（> 30min delay 留缓冲）自动关单，复用既有状态机业务（与 delay consumer 的竞态由状态机语义一致收敛）。scan1 查 saga 状态经 dtm HTTP API（须与 payment 服务网络可达，compose/集群同网络）。scan3（可选）——dtm in-flight 超龄告警。
- **C Stripe 双向对账（治⑤⑥）**：日终拉取 Checkout Sessions（窗口按 created），以 normalized transaction_id join 本地 payments；方向1（Stripe 有账、本地未 PAID）→ **自动重放** webhook 处理逻辑（安全垫：支付状态机 + EnsureSaga 幂等 + outbox + 扫描，全部已建），重放失败转告警；方向2（本地 PAID、Stripe 无账）→ 告警 + 冻结，**不自动回滚**（结算已发生，乱回滚放大伤害）。
- **D 人工产出层**：**倾向**异常单待办表（可查可关）+ 结构化告警（纯日志方案留为 design 备选）；同一 orderId 跨层（outbox dead/scan1/scan2/对账）重复发现必须聚合去重。退款实现（Stripe Refund API + FULLY_REFUNDED 流转）**另立项**（Non-goal）。
- **保护既有机制**：订单状态机/barrier/确定性 gid/two-state 幂等键/consumerx 全部不动；PaymentDelayMQ 继续休眠（其"超时对账"设计意图由 B 扫描器实质取代）。
- **测试策略（原则性声明）**：外部依赖（DB/RPC/Stripe/dtm）一律 mock（Rule 6）；关键场景——relay 崩溃续扫、outbox dead 路径、scan1 墓碑递增、scan2 与迟到支付竞态、C 重放幂等、对账差额生成；测试基于能力规范编写（Rule 2）。

## Capabilities

### New Capabilities

（无——兜底三层与人工产出同属结算可靠性域，全部并入既有能力。）

### Modified Capabilities

- `async-settlement-reliability`：增补四组需求——结算待办必达投递（Outbox + relay）、定时对账扫描自愈（scan1/scan2 分流）、Stripe 双向对账（方向1 自动重放/方向2 告警冻结）、异常单人工产出（聚合去重）；**"结算发起幂等补建"升级为 outbox 驱动的必达语义**（废除归档 change design 决策 1 附注"无自动自愈、人工承接"的残余风险声明，审核对照点：归档 design.md:52-53、:71）；"失败路径可观测"扩展 outbox dead/扫描发现/对账差额/墓碑重试等新来源。

## Impact

- **代码**：payment 服务（webhook 写路径改造为 TransactCtx 两写——需新增 PaymentModel 的 WithSession 事务变体，参照 inventory 先例；**倾向** relay + 扫描器以常驻监督循环同宿主于 payment，独立 reconciler cmd 留为 design 备选；EnsureSaga 复用、Stripe Sessions 拉取与重放）；order 服务（关单业务入口从 delay handler 抽取为可复用入口——供 scan2 调用）；`common/utils` 无改动。
- **DB/DDL**：`payment_outbox` 表 + 异常单待办表（D7 倾向表方案，纯日志留为 design 备选），走 `construct/depend/sql` 目录；均在 mall 库。
- **配置**：扫描间隔、时间阈值（10min/35min）、重试上限（relay attempts / 墓碑 r 上限 3）进 yaml（C13），不写死。
- **依赖**：无新外部依赖（stripe-go v81 已接、dtm v1.19.0 已接、postgres/rabbitmq 已有）。
- **行为**：payment 服务新增两个常驻后台循环（relay、scanner）；webhook 写路径单行更新变为两写事务；无接口删除、无既有语义变更。
- **风险**：outbox 写失败会回滚整个 PAID 翻转（Stripe 重试兜底，语义安全）；扫描器误扫由时间阈值防（10min/35min 窗口）；自动重放的信任边界（Stripe 账本为准）已由用户拍板（D3）；扫描器跨域读（payments+orders 同库）为对账作业惯例，写路径全部走服务 RPC。
- **部署**：先迁移 DDL；order 服务先发（关单入口抽取）；payment 后发（relay/scanner/对账）。上线后"无自动自愈"残余风险声明即失效。
