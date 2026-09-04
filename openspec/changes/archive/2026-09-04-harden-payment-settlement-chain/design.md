# Design — harden-payment-settlement-chain（第 2 版，Saga 路线）

## Context

探针实证（explore-brief 实证 0-10）确立了链路缺陷全貌；实证 11 确认 DTM 地基已备而未接线；
实证 12-13 补充 delay/audit 细节。第 1 版骨架已落地代码库（未提交），本次修订建立在其之上：

- **已落地（保持/重构）**：`common/mq/consumerx` 单态骨架（手动确认、QoS(1)、监督重启、
  panic recovery、优雅停机、有界重试、毒消息即弃）、notify/delay 薄适配层、
  `idempotency.Release`、全服务 consumer 配置。
- **未动（本次改造）**：payment 快路径（payment.go:210-213 附近仍直接 RPC）、
  `updateorder2paymentsuccesslogic.go:108` 的 MQ publish。
- **关键既有事实**：
  - `UpdateOrder2PaymentSuccess` 对非 PendingPayment 订单（含"已 Paid"重放）在 :102-105
    返回 `codes.Aborted`——作 Saga 分支会被 DTM 误判失败触发补偿，必须调整（决策 2）。
  - DTM 地基：dtm server 在 `configs/docker-compose.dev.yml:109`（**optional profile**，
    `docker compose --profile optional up` 显式启用；镜像 `yedf/dtm:latest`），
    `construct/depend/conf/dtm.yaml` 仅配 MicroService（gozero driver + consul），
    **未配 Store**；order 三补偿 RPC 完整；`ReturnInventory`/`ReleaseCoupon` 存量在位；
    所有 go.mod 无 dtmclient；三库无 `dtm_barrier` 表。
  - streadway/amqp 已随第 1 版进入 common 依赖。

## Goals / Non-Goals

**Goals:**

- 结算链路 Saga 化：三分支 + 补偿接入 DTM，消灭"DB 提交 + MQ 发布"双写。
- 幂等补建：快路径 EnsureSaga，Stripe 重试天然驱动。
- 分支幂等双保险：barrier 表 + 业务幂等容忍。
- consumerx 从单态升级为两态键 lease 语义（PROCESSING → SUCCESS / 释放 + 重试计数）。
- notify MQ 整条退役；delay 链路保持第 1 版可靠性语义。

**Non-Goals:**

- 不引入事务性 Outbox、DLX、延迟对账扫描、PaymentDelayMQ 复活。
- 不重构 webhook 与支付单模型；不改订单状态机前置校验（PendingPayment→Paid）。
- 不做 Stripe event.ID 去重；不修 audit consumer（独立后续）。
- 不更换 AMQP 库；不引入 Kafka。

## Decisions

### 决策 1：Saga 接入形态与 gid 策略

payment 服务在本地 PAID 事务成功后，经 `dtmclient`（gRPC 模式，配 `dtm-driver-gozero`）
发起 Saga。**gid 采用确定性构造 `settle:{orderId}`**：orderId 全局唯一 ⇒ 同一订单重复发起
必撞同一 gid，dtm 对重复 gid 报错 ⇒ 发起方将"gid 已存在"视为成功，天然幂等。saga 分支数据
（orderId/userId/paymentResult、items、couponId 等）随分支请求下发。

- 发起逻辑封装在 payment `svc` 层（`SettlementSaga` 封装：Initiator + Ensure 两个方法），
  `payment.go`（main 包）只调用封装，pb 转换不外泄（Rule 4）。
- **备选**：payment-PAID 作为 saga 分支 0（Shape B，零双写）——需 payment 自分支 RPC +
  补偿 + webhook 重构，工程量数倍；方案 A 的残余风险（PAID 成功但发起失败且 Stripe 不重试）
  用户已接受，否决。
- **备选**：Msg/Workflow 模式——本链路是典型的 rollback 型事务，Saga 语义最贴合，否决其他模式。
- API 形状以 dtmclient 实际版本为准（实施期以最小 spike 验证 `NewSagaGrpc` + `Add` +
  `Submit` 及重复 gid 错误语义），设计不锁定具体函数签名。

### 决策 2：分支/补偿映射与"已 Paid"语义调整

| 分支 | 前向（存量） | 补偿（存量） |
|---|---|---|
| 1 订单 | `UpdateOrder2PaymentSuccess` | `UpdateOrder2PaymentSuccessRollback` |
| 2 库存 | `DecreaseInventory` | `ReturnInventory` |
| 3 优惠券 | `UseCoupon` | `ReleaseCoupon` |

分支 1 的状态机语义调整为三分支判定（在 barrier 之后执行）：

- 订单 PendingPayment → 正常流转 Paid（原路径）；
- 订单**已 Paid** → 返回成功（幂等重放，不 Aborted）——消除 DTM 误判失败的资损风险；
- 订单其他状态（Closed/Cancelled 等真实冲突）→ 维持 `codes.Aborted` 失败，Saga 走补偿 +
  error 告警，属正确业务行为。

两个既有 Aborted 分支在 Saga 语境下的行为声明（不改其判定，仅声明预期）：

- `:44-48` 订单不存在（`OrderNotExist` → 外层 `codes.Aborted`）：作为分支失败处理。
  **已知悬挂风险**：若该分支已成功后进入补偿、而订单行其间被删除，补偿
  `UpdateOrder2PaymentSuccessRollback`（:42-49 同样对不存在返回 Aborted）会持续重试失败，
  Saga 补偿链悬挂直至人工介入。靠 barrier 主保险（先挡重放）+ dtm 补偿重试告警承接；
  实施测试须覆盖该边界（异常数据触发补偿失败的告警路径）。
- `:64-69` payment status 非 Paying（`PaymentStatusInvalid` → 外层 `codes.Aborted`）：
  订单状态判定已覆盖绝大多数重放（已 Paid 在其前短路）；该分支保持失败语义，仅在
  "订单 PendingPayment 但支付单状态异常"的组合下触发，属真实冲突，失败 + 告警正确。

分支 2/3 依赖下游既有容忍语义（"已扣减"/"已使用"按成功，spec 已契约化）+ barrier 双保险。
分支 payload 一致性：DecreaseInventory 与 ReturnInventory 使用同一 items/PreOrderId 集合，
保证补偿可精确逆操作。

- **备选**：仅靠 barrier 不改"已 Paid"语义——barrier 只挡"已执行过"的重放，挡不住
  barrier 记录缺失（迁移前已 Paid 的存量订单、barrier 写入失败边缘）时的 Aborted 误判，
  两层都要，否决单一依赖。

### 决策 3：barrier 基建（必须项，先于代码上线）

- 三库（order/inventory/coupons）执行 dtm 官方 `dtm_barrier.barrier` 表 SQL（迁移脚本入库，
  遵循项目 SQL 目录约定）。
- 分支 handler 接入 dtm gRPC barrier：barrier 命中（分支已执行）直接返回成功，业务逻辑不执行。
- **dtm Store 配置补齐**：当前 `construct/depend/conf/dtm.yaml` 无 Store，默认容器内 bolt
  文件，重启即丢 Saga 进度。MUST 补 MySQL Store（复用 compose 现有 MySQL），生产
  `manifests/dtm.yaml` 同步。

### 决策 4：consumerx 两态键重构（洞 2 完整解）

`Store` 接口从布尔语义升级为状态语义（key 值即状态字符串）：

```
Claim(ctx, key, leaseTTL) → (State, error)   // SETNX key=PROCESSING PX=leaseTTL
                                              // State: New | Processing | Done
MarkDone(ctx, key, successTTL) error          // SET key=SUCCESS PX=successTTL(24h)
Release(ctx, key) error                       // DEL（失败路径；沿用已落地实现）
IncrWithTTL(ctx, key, ttl) (int64, error)     // 重试计数（保留）
```

`Process` 流程调整：decode → Claim →（Processing/Done 则 Ack 跳过）→ Handle 成功 →
MarkDone + Ack；失败 → Release + IncrWithTTL（超限弃单）。**PX 过期即 stale 检测**：
租约到期后 SETNX 自然成功，无需 timestamp 比对。Processing 状态 Ack 跳过的取舍：避免
重复消息阻塞队列头部；极端情形（原消费者崩溃于 Handle 中、租约未过期、重投被跳过）的
丢失窗口由 lease TTL 界定（到期后重放收敛），delay 链路低频且有告警日志可查。

- **旧单态 key 兼容**：已落地代码经 `CheckAndSet` 写入的 key 值为 `"1"`。新 Claim 的
  Lua 脚本 MUST 将旧值 `"1"` 识别为 Processing（跳过、等租约自然过期），不得按未知值
  报错或误判 New 重复处理；旧 key 最长 1h 自然过期，上线窗口内无长残留。

- `common/utils/idempotency` 新增状态原语；**保留 `CheckAndSet`**——audit consumer 仍在
  使用（C13：audit 不在本次范围，避免为不修的代码制造破坏性变更）。
- 配置：`IdempotencyTtlSeconds` 语义收窄为 lease TTL（PROCESSING），新增
  `SuccessTtlSeconds`（默认 24h）。
- notify 薄适配层不单独适配两态键——它随决策 5 整体删除；delay 适配层与 payment/mq 骨架
  随新接口编译期适配。
- **备选**：时间戳 + stale 判定——PX 过期已原子覆盖该需求，否决（更复杂、非原子）。

### 决策 5：notify MQ 退役清单

删除：`services/order/internal/mq/notify/`（consumer.go 薄适配层、init.go 队列/交换机声明、
product.go）、`servicecontext.go` 的 `OrderNotifyMQ` 字段与装配、
`updateorder2paymentsuccesslogic.go` 的 publish 调用与 notify import。清理后全仓 grep 无
notify 队列残留引用（`make build && make lint` 守门）。

- **部署窗口残留消息**：退役上线时 notify 队列可能残留未消费消息。结算已改走 Saga，残留
  消息仅代表历史重复通知，消费与否不影响正确性（业务幂等收敛）；窗口内排空或直接弃队列
  （随 exchange/queue 声明删除，队列消失即消息消失），不保留僵尸消费者（C12 教训）。
- **备选**：保留 notify 为"结算完成事件广播"——无消费方，YAGNI，否决。

### 决策 6：payment/mq 骨架同步重构，不接线（延续第 1 版决策 7）

`services/payment/internal/mq` 随两态键接口编译期适配（消除 autoAck 混用，实证 0）；
`PaymentMQ` 保持 nil、Init 保持注释。复活是独立变更。

### 决策 7：已落地骨架的保持项

第 1 版已实现并经第 1 周期审核的语义原样保持：监督重启（指数退避 cap 30s）、graceful
shutdown（cancel → 停取 → timer 限时 → 超时关连接）、panic recovery（按一次业务失败计重试）、
毒消息一次即弃、QoS(1) 串行、重试计数 key（`settlement:retry:{queue}:{identity}`）。
本次仅按决策 4 调整幂等键相关流程，其余不动，避免无谓回归。

## Risks / Trade-offs

- [dtm Store 未配置，默认容器内 bolt，重启丢 Saga 进度] → 决策 3 强制补 MySQL Store，
  迁移步骤第 1 项；生产 manifests 同步。
- [`yedf/dtm:latest` 镜像漂移] → 迁移时 pin 具体版本（与 dtmclient 版本匹配），实施期确定。
- [dtm server 单点（compose 单实例）] → Saga 进度持久化在 MySQL，进程重启续跑；可用性影响
  结算时效而非正确性；后续增强：多实例 + 存储共享。
- ["已 Paid"放宽语义的边界] → 仅放宽到"已 Paid 即成功"；Closed/Cancelled 等真实冲突仍失败
  补偿 + 告警。与快路径重放、通知重放的既有语义一致。
- [补偿正确性依赖分支 payload 一致] → Decrease/Return 使用同一 items 集合；Saga 分支数据
  固化在发起时快照，不重读可变状态。
- [dtmclient 重复 gid 错误语义依赖] → 发起方将"gid 已存在"错误映射为成功；实施期 spike
  验证该错误可与其他错误区分（错误信息含 gid）。
- [迁移窗口双路径] → 部署顺序（先分支服务后 payment）保证 saga 发起时分支已就位；回滚
  payment 二进制即停新 saga，已提交 saga 由 dtm 续跑完成。
- [audit consumer 仍用单态 CheckAndSet] → 保留旧原语，互不影响；audit 修复独立立项。

## Migration Plan

1. dtm Store 配置（MySQL）——dev（`construct/depend/conf/dtm.yaml`）与生产
   （`manifests/dtm.yaml`）两份同步补齐——并 pin 镜像版本；三库执行 `dtm_barrier` 表迁移。
2. common：consumerx 两态键重构 + idempotency 状态原语；`go mod tidy`（common 及四个服务）。
3. order/inventory/coupons：分支 barrier 接线 + 分支 1 语义调整 + notify 退役 + delay 适配；
   `make mock && make lint && make test-unit && make build` 全绿后**先部署**。
4. payment：saga 发起封装 + EnsureSaga + 快路径改造（svc 层，Rule 4）+ payment/mq 适配；
   门禁全绿后**最后部署**。
5. 开发环境：`docker compose --profile optional up` 启动 dtm；生产按 manifests（补 Store 后）。
6. 回滚：回退 payment 二进制即停止新 saga 发起；已提交 saga 由 dtm 依 MySQL 进度续跑完成；
   分支服务回退需在无进行中 saga 时进行（运维窗口）；notify 代码随分支服务版本回退即恢复
   MQ 路径（回滚窗口内双路径并存，靠三重业务幂等收敛）。

## Open Questions

（无阻塞项。dtmclient 具体版本与 gRPC API 形状、dtm 生产部署 Store 落点，实施期以最小
spike 确认；重试上限 3、退避基值、lease/SUCCESS TTL 均已配置化，可调。）
