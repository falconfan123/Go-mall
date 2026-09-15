# Proposal — harden-payment-settlement-chain（第 2.1 版，Saga 路线）

> 修订说明：第 1 版（2026-09-03）按"consumer 止血 + SettleOrder 补发"路线冻结，**且其骨架部分已落地代码库**（未提交：`common/mq/consumerx` 单态语义骨架 + notify/delay consumer 迁移 + `idempotency.Release` + 全服务 consumer 配置；payment 快路径与 MQ publish 未动）。探索第二轮（2026-09-04，见 explore-brief.md）实证 DTM 基础设施已备而未接线（实证 11），用户拍板改为 Saga 化路线（D2/D3 推翻，D7-D11 新增）。本版基于**已落地的第 1 版代码现状**重写：consumerx 从"新建"改为"重构"，其余按新基线展开。

## Why

支付成功结算链路（订单状态 → MQ → 扣库存/用券）存在系统性可靠性缺陷（explore-brief 实证 0-10）：MQ consumer 因 autoAck=true 与手动确认混用而"一次性死亡"、幂等键失败不释放、MQ 发布失败后订单已 Paid 而副作用永久悬空。第 1 版以 consumerx 骨架 + 有界重试止了血（已落地），但洞 1 的结构性根源——"DB 提交与 MQ 发布"双写——仍在：结算靠手工 MQ 编排串联三个跨服务写操作。第二轮探索（实证 11）确认项目早已为 DTM Saga 备好地基（dtm server 在 dev compose、order/inventory/coupons 三服务前向与补偿 RPC 均已实现），只是从未接线。按 AGENTS.md Rule 7 回归 Saga 正解：从结构上消灭洞 1（双写），同时把第 1 版落地的 consumerx 升级为两态键语义（洞 2 的完整解），并让 notify MQ 整条退役（洞 2/洞 3 的结算版本随之消灭）。

## What Changes

- **结算链路 Saga 化（洞 1，结构正解）**：payment webhook 本地 PAID 后发起 DTM Saga（gid = `settle:{orderId}` 确定性），三个分支将存量 RPC 纳入 DTM 事务编排：`UpdateOrder2PaymentSuccess` ⇄ `UpdateOrder2PaymentSuccessRollback`、`DecreaseInventory` ⇄ `ReturnInventory`、`UseCoupon` ⇄ `ReleaseCoupon`。DTM 保证分支自动重试 + 失败逆序补偿。**BREAKING**（结算编排从 MQ 手工编排改为 DTM 驱动）。
- **webhook 快路径幂等补建（洞 1 兜底）**：`payment.go` 快路径（订单已 Paid 时当前直接返回）改为 EnsureSettlementSaga——确定性 gid 重复提交被 dtm 拒绝即视为已存在，幂等补建；Stripe 重试天然驱动补建。已知残余风险：PAID 成功但建 saga 失败且 Stripe 不重试时无自动自愈，靠 error 日志 + 人工承接（用户已接受）。saga 发起与 EnsureSaga 的实现落在 payment `svc` 层封装（pb 转换不外泄 main 包，Rule 4）。
- **order/notify MQ 整条退役（洞 2/洞 3 在结算链路上消灭）**：删除 publish（updateorder2paymentsuccesslogic.go:108-118）、notify consumer（现为 consumerx 薄适配层）、notify 包的 exchange/queue 声明与 product、svc 装配。**BREAKING**（结算不再经过 MQ）。
- **barrier 基建（Saga 分支幂等，必须项）**：order/inventory/coupons 三库新增 `dtm_barrier` 表迁移；分支 handler 接入 dtm barrier 防分支重复执行；"已 Paid"重放的判定语义（业务码 vs gRPC error）在 design 中明确，确保 DTM 不会误判分支失败触发错误补偿——与库存"已扣减"/券"已使用"容忍构成双保险。
- **consumerx 重构为两态幂等键 lease 版（洞 2 完整解）**：现有单态 Claim/Release 语义升级为 `SETNX key=PROCESSING PX=leaseTTL` → 成功 `SET key=SUCCESS PX=successTTL(24h)` / 失败 `DEL key` + `INCR retry:{key}`（超限弃单 + 告警）；PX 过期即 stale 检测（无需 timestamp 比对）。**BREAKING**（幂等键语义变更）。毒消息一次即弃 + 告警保持现状。
- **delay consumer 适配**：`services/order/internal/mq/delay` 已在第 1 版迁移至 consumerx（确认语义与 `Ack(true)` 误用已随之消除），本次仅随骨架重构适配两态键语义。`services/payment/internal/mq` 骨架同步重构，消除其 autoAck=true 与手动确认混用（实证 0 的 channel 死亡模式）；`PaymentMQ` 保持 nil 不接线（消除带病休眠）。
- **保护既有业务幂等**：全部修复不得破坏三重业务幂等——订单行锁+状态机（Paid 短路）、DecreaseInventory"已扣减"容忍、UseCoupon"已使用"容忍；saga 分支与 EnsureSaga 自身必须幂等。
- **失败可观测**：弃单/毒消息/supervisor 重启/saga 补偿路径输出结构化错误日志（对齐现有 observability 规范），不再静默。
- **测试策略（原则性声明）**：外部依赖（RPC/Redis/AMQP/DTM）一律 mock（Rule 6）；测试必须覆盖四类关键场景——saga 分支重试不误补偿、EnsureSaga 幂等、channel 死亡后 supervisor 可恢复、单条消息失败路径不双扣；测试基于 API 规范编写（Rule 2）。

明确不做：不引入事务性 Outbox；不引入 DLX/DLQ（弃单 + 告警承接）；不做 Stripe event.ID 接口层去重；不做延迟对账扫描；不复活 PaymentDelayMQ（svc 中 `PaymentMQ` 保持 nil、不恢复 Init，骨架随本次重构）；不修 audit consumer 同款问题（实证 13，记为已知，后续独立修）；不采用 payment-PAID 入 Saga 的 Shape B（D7 定为方案 A：本地 PAID + 后发 saga）；不重构 webhook 与支付单模型；不改订单状态机既有语义与前置校验逻辑（PendingPayment→Paid 保留，"已 Paid"重放的返回语义调整属分支幂等适配，见下）。

## Capabilities

### New Capabilities

- `async-settlement-reliability`：支付成功结算链路的可靠性契约——结算 Saga（分支/补偿/barrier 幂等/EnsureSaga 幂等补建）、delay consumer 确认语义与生命周期监督、两态幂等键生命周期、单条消息失败处理（重试上限/毒消息/无 fall-through）、失败路径可观测。（第 1 版定义随 SettleOrder 方案作废，本版重构：MQ 消费可靠性部分收窄至 delay 链路，新增 Saga 化契约。）

### Modified Capabilities

（无——现有 `observability` spec 的需求不变，本次仅按其规范输出日志，不改变其要求。）

## Impact

- **代码**：`common/mq/consumerx`（**重构**既有骨架：单态 → 两态键，Store 接口扩展 SUCCESS 转换语义）、`common/utils/idempotency`（已有 Release；新增两态键状态原语）、`services/order/internal/mq/notify/`（**整目录删除**，含 init.go 声明与 product）、`services/order/internal/svc/servicecontext.go`（删除 `OrderNotifyMQ` 字段与 `initOrderNotifyMQ` 装配）、`services/order/internal/mq/delay/`（适配两态键）、`services/payment/internal/mq/`（骨架同步重构，不接线）、`services/payment/payment.go`（快路径 EnsureSaga）、`services/payment/internal/svc`（dtm client 接入 + saga 发起封装）、`services/order/internal/logic/updateorder2paymentsuccesslogic.go`（删除 MQ publish 与 notify 包 import、接 barrier、"已 Paid"重放返回语义调整——现返回 `codes.Aborted`，作 Saga 分支会被 DTM 误判失败触发错误补偿，须改为分支幂等成功语义，详见 design）、`services/inventory/internal/logic/{decreaseinventorylogic.go,returninventorylogic.go}`（barrier）、`services/coupons/internal/logic/{usecouponlogic.go,releasecouponlogic.go}`（barrier）。
- **依赖**：order/inventory/coupons/payment 服务 go.mod 新增 `github.com/dtm-labs/dtmclient`（含 dtm-driver-gozero）；common 模块 amqp 依赖已随第 1 版落地（无需新增）。
- **配置**：order/inventory/coupons/payment yaml 新增 DTM server target；consumer 参数已随第 1 版落地（RetryLimit/BackoffBaseMs/ShutdownTimeoutMs/IdempotencyTtlSeconds），本次将 `IdempotencyTtlSeconds` 语义拆分为 lease TTL（PROCESSING，配置化）与 SUCCESS TTL（默认 24h）两个配置项。
- **DB/基建**：order/inventory/coupons 三库 `dtm_barrier` 表迁移；dtm server + consul 为部署前置——注意 `configs/docker-compose.dev.yml` 中 dtm 服务在 `optional` profile 下，本地开发需 `docker compose --profile optional up` 显式启用。
- **API**：无新增 RPC；**删除**第 1 版遗留的 `SettleOrder` RPC（并行实施期已落地：order proto/logic + payment 快路径调用，本 change 以 EnsureSaga 取代并退役之，见 design 决策 1 附注）。
- **行为**：结算路径不再经过 MQ（**BREAKING**）；幂等键两态语义变更（**BREAKING**）；delay consumer 确认语义变更（投递即确认 → 处理完成才确认）已在第 1 版落地，本次仅随骨架重构适配两态键，无新增行为变更。
- **风险**：saga 补偿正确性依赖分支幂等（barrier 必须先行落地）；dtm server 单点可用性影响结算时效（重试由 DTM 自带，服务恢复后续跑）；已落地的第 1 版骨架需重构而非回退，存在迁移窗口。
- **部署**：三库 barrier 表迁移 → 部署 order/inventory/coupons（分支+barrier 就位）→ 最后部署 payment（saga 发起）。
