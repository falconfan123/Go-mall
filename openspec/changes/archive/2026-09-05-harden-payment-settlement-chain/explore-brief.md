# Explore Brief — harden-payment-settlement-chain

> 探索落盘（2026-09-03）。本文是后续 proposal/design/specs/tasks 过反思审核时的**需求基线**，
> 审核方须以本 Checklist 逐项核对，防止细节在上下文丢失后走样。

## 背景 — 现状调查结论（含实证）

### 支付结算链路现状

```
Stripe webhook (payment.go:66 handleStripeWebhook)
  → processStripePaymentSuccess (payment.go:174)
      → payment DB 更新为 PAID（payment.go:183-197，无锁 check-then-act）
      → 快路径：订单已 Paid+Paid 直接 return nil（payment.go:210-213）
      → RPC UpdateOrder2PaymentSuccess（order 服务）
          → 订单事务：行锁 + 状态机前置校验（updateorder2paymentsuccesslogic.go:41-69）
          → 订单更新为 Paid（同文件 70-80）
          → MQ Product 发布 order-notify 消息（同文件 108-118，失败返回 Internal）
  → order/notify consumer 消费
      → 订单事务兜底（notify/consumer.go:72-93）
      → DecreaseInventory（116）→ UseCoupon（134）→ Ack（154）
```

### 实证结论

**实证 0（探针实验，本轮新发现，最高严重级）**：在本地 broker（localhost:5672，admin/admin）
用 streadway/amqp v1.1.0 实测：autoAck=true 的 consumer 手动调用 `Ack`/`Reject`，
调用本身返回 nil（fire-and-forget），但 broker 回 406 PRECONDITION_FAILED
（unknown delivery tag）→ **channel 死亡 → deliveries channel 关闭 → `for range` 退出 →
consumer goroutine 静默死亡**。连接仍存活，无任何告警。
**推论：三个 consumer（order/notify、order/delay、payment/delay）都是"一次性"的——
每条消息（无论成功路径的 Ack 还是失败路径的 Reject）都会杀死 channel，服务从此永久失聪，
直到进程重启。** 每次重启只能再处理一条。此前未大面积暴露是因为手动测试每次只付一单。

**实证 1（洞1）**：MQ 发布失败返回 Internal（updateorder2paymentsuccesslogic.go:108-118），
但订单已提交 Paid。Stripe 重试 webhook 撞快路径（payment.go:210-213）直接 200，
MQ 永不补发 → 扣库存/用券永久缺失。结构性根源：`UpdateOrder2PaymentSuccess` 的状态机
校验只放行 PendingPayment→Paid（同文件 :57），已 Paid 的重试在 :102 被 Aborted 挡掉，
"挡并发"与"允许结算重试"耦合在同一入口，无法自然重放。

**实证 2（洞2）**：幂等 key SETNX 占坑后失败不释放。CheckAndSet 在业务前占坑
（notify/consumer.go:54，common/utils/idempotency/idempotency.go Lua SETNX+EXPIRE 原子），
所有 Reject(true) 分支（96/123/144）均无 Del → 重投消息撞 isProcessed=true 被 Ack 跳过 → 永久丢失。

**实证 3（洞3，被实证 0 取代并升级）**：autoAck=true（notify/consumer.go:30）+ 全手动
Ack/Reject 混用 → 见实证 0，后果是 consumer 一次性死亡而非单纯消息丢失。

**实证 4（洞4）**：事务失败分支 Reject(true) 后无 continue/return（notify/consumer.go:93-99），
代码继续执行：
- GetOrder 失败 → orderModelRes 为 nil → :101 `orderModelRes.OrderId` 空指针 panic
  （Go panic 崩整个进程，不分 goroutine）；
- Update 失败 → orderModelRes 有值 → 继续执行 DecreaseInventory/UseCoupon 重复副作用，
  且消息已被（无效地）Reject。
order/delay/consumer.go:96-101 → 109 存在一模一样的复制 bug。

**实证 5（洞5，新发现）**：consumer 无监督循环。三个 consumer 均为
`go xxx.consumer(context.TODO())` 裸跑一次（notify/init.go:101、payment/mq/init.go:83），
goroutine 退出 = 永久失聪，无重连、无告警。即使修好 autoAck，网络抖动/任何 channel
异常仍会导致永久停摆。

**实证 6（洞6，新发现，潜伏）**：业务失败 Reject(true) 无限重投 + 永久性失败（如订单
ErrNotFound）= 死循环热烧 CPU。**当前该洞不存在（channel 反正会死），修好实证 0 后即诞生**，
修洞必须成套。另：order-notify-queue 未声明 DLX（notify/init.go:67 args=nil），
Reject(false) 毒消息静默丢弃，无计数、无告警。

**实证 7（洞7，新发现）**：幂等 key TTL=1h（notify/consumer.go:54）且语义为"占坑即已处理"：
崩在 CheckAndSet 之后 → key 挡 1 小时 → 期间重投全被跳过。缓解因素：链路三步全部自带
业务幂等——订单行锁+状态机（Paid 短路）、DecreaseInventory"已扣减"不算失败
（notify/consumer.go:129）、UseCoupon 同理（:149）。Redis 仅是省重复功的优化，
不是正确性的唯一防线。

**实证 8（新发现）**：`PaymentDelayMQ`（x-delayed-message、30min 延迟、创建支付单时已
Product，consumer 里幂等检查已写但业务逻辑为空）整个被禁用：
`svc/servicecontext.go:28` init 被注释、`:41` `PaymentMQ: nil`。设计意图疑似超时对账，写一半废弃。
order/delay MQ 的 x-delayed-message 基础设施可用（delay 退避重试可复用该机制）。

**实证 9（新发现）**：payment.go:96 只 switch event.Type，全程未使用 event.ID，
接口层无 Stripe 事件去重；payment.go:183-197 支付单更新为无锁 check-then-act
（FindOne → 非 PAID 才 Update）。

**实证 10（DTM 骨架已存在）**：`updateorder2paymentsuccessrollbacklogic.go`（完整补偿实现：
行锁+状态检查+回写）、`createorderrollbacklogic.go`、`updateorder2paymentstatusrollbacklogic.go`
均为 RPC 形式的 Saga 补偿分支；coupons/lockcouponlogic.go:86 有 dtm barrier 痕迹
（注释"一般数据库不会错误不需要dtm回滚，就让他一直重试"）。未发现活跃 saga 驱动。
AGENTS.md Rule 7 要求跨服务事务必须 DTM Saga + 补偿。

### 失败模型总览（修订后）

```
第0层 consumer 生命周期    洞3'(实证0) channel 死亡   洞5(实证5) 无监督
第1层 单消息控制流          洞4(实证4) fall-through
第2层 重试语义             洞2(实证2) key 不释放  洞6(实证6) 无限重投  洞7(实证7) TTL/语义
第3层 发布端               洞1(实证1) dual-write
加分项                     实证9 event.ID 去重
```

## 目标 / 非目标

### 目标（用户视角）
1. 支付成功后，扣库存/用券副作用**最终必然发生**（at-least-once + 业务幂等收敛），不再依赖"碰巧重启服务"。
2. MQ 消费者的生命周期可靠：channel 异常/网络抖动后自动恢复，恢复有日志可查。
3. 单条消息失败不会崩进程（nil panic）、不会双扣、不会死循环烧 CPU。
4. MQ 发布失败不再造成订单已 Paid 而副作用永久悬空。
5. 一切异常路径有日志/告警可观测，不再静默。

### 非目标（明确排除）
1. **不在本次引入 DTM Saga 运行时**（档2/档3 架构升级单独立项）——本次为止血修复，保持现有 MQ 编排。
2. 不重构 webhook 与支付单模型（payment.go:183-197 无锁区靠行锁兜底已够，event.ID 去重为加分项非必做）。
3. 不改订单状态机本身的语义（PendingPayment→Paid 前置校验保留，重试能力另开入口解决）。
4. 不引入新中间件（不加 Kafka/不改 broker）。

## 待决策点

| # | 决策点 | 选项 | 倾向 |
|---|---|---|---|
| D1 | 修复范围：三 consumer 一起修还是只修 notify | (a) 只修 order/notify（支付主链路）；(b) 三个一起（bug 是复制的，order/delay 还有同款 nil-deref） | **(b)**，同一套骨架统一替换 |
| D2 | 洞1 发布可靠性机制 | (A) 事务性 Outbox（新表+relay，最正统）；(B) 快路径补发（webhook 发现已 Paid 时调新 `SettleOrder` RPC 补发，改动最小）；(C) 复活 PaymentDelayMQ 做 30min 延迟对账扫描（自愈式，基建半现成但当前整个被禁用） | **B**（止血优先），C 可作为后续增强单独评估 |
| D3 | DTM 本次上不上 | 档1 纯 consumer 修固 / 档2 MQ 触发 Saga / 档3 全 Saga 化 | **档1 本次，档2 立项另议** |
| D4 | 幂等语义 | (a) 最简：Reject 前 Del key；(b) done-marker：全成功后 SET；(c) 两态 key claim(短TTL)→done(长TTL) | **(a)**，三步业务幂等已兜底，不必过度设计 |
| D5 | 重试上限计数方式 | (a) Redis INCR per-order attempt key；(b) 引入 DLX + x-delayed 延迟退避 | 倾向 **(a) 简单版起步**，DLX 退避留作 C 方案配套 |
| D6 | event.ID 接口层去重（实证9） | 做或不做 | **本次不做**（P2 加分项，行锁+状态机已兜住资金路径） |

## Checklist — 提案审核基线

后续 proposal/design/specs/tasks 审核时逐项核对：

- [ ] **C0 channel 修复**：三个 consumer 全部 autoAck=false；手动 Ack 仅在业务全部成功后发送（实证 0）。
- [ ] **C1 fall-through**：所有失败分支 Nack/Reject 后必须 `continue`；order/delay 的同款 nil-deref（delay/consumer.go:109）一并修复（实证 4）。
- [ ] **C2 监督循环**：consumer 外层有 supervisor（崩溃/断连后重连+退避+日志），`context.TODO()` 消除，接入服务生命周期（实证 5）。
- [ ] **C3 幂等释放**：失败重投前 Del 幂等 key（或等价机制），重投消息不再被旧 key 跳过（实证 2）。
- [ ] **C4 重试上限**：重试次数有上限（D5 方案），超限弃单+告警，禁止无限 requeue；毒消息（解析失败）一次即弃+告警（实证 6）。
- [ ] **C5 毒消息策略**：JSON 解析失败不再 Reject(false) 静默丢弃——至少结构化日志+可观测（notify 队列无 DLX 是现状，需明确取舍）。
- [ ] **C6 发布失败兜底**：洞1 按 D2 选型落地；选 B 时需解决"已 Paid 状态机挡住结算重试"的结构问题（新入口，不复用 UpdateOrder2PaymentSuccess 前置校验）（实证 1）。
- [ ] **C7 幂等 key 语义**：TTL 1h 过短问题有明确交代（至少文档化风险；两态 key 不强制）（实证 7）。
- [ ] **C8 不回归业务幂等**：修复不得破坏现有三重业务幂等（订单状态机/库存已扣减容忍/券已使用容忍），且新增路径（如 SettleOrder）自身幂等。
- [ ] **C9 DDL 红线**：admin 服务 pb 结构体不泄漏 logic/service 层；新增 RPC 走 proto 生成（Rule 3/4）。
- [ ] **C10 测试策略**：基于 API 规范编写；外部依赖（RPC/Redis/MQ）用 mock（Rule 2/6）；consumer 修复需有"channel 死亡可恢复"与"失败不双扣"两类单测。
- [ ] **C11 明确不做**：DTM 运行时、event.ID 去重、Outbox 全家桶若排除，需在 proposal 的 Non-goals 显式声明。
- [ ] **C12 PaymentDelayMQ 处置**：保持禁用或明确其复活条件，不留"半死"状态（实证 8）。

---

# 第二轮探索（2026-09-04）— DTM Saga 路线锁定

> 本轮结论**修订并取代**第一轮的 D2/D3/D4 决策与 Checklist 相关项。
> 用户拍板记录：洞1=直接上 DTM；payment-PAID=方案 A；notify MQ=彻底删除；change=修订现有。

## 新实证（第二轮）

**实证 11（DTM 地基，"备好但从未通电"）**：
- dtm server 已在基础设施中：`configs/docker-compose.dev.yml:109`（yedf/dtm:latest +
  dtm-driver-gozero + consul 注册，`construct/depend/conf/dtm.yaml`、`manifests/dtm.yaml` 同款）。
- 前向分支 RPC 全部存在：`UpdateOrder2PaymentSuccess`（order）、`DecreaseInventory`（inventory）、
  `UseCoupon`（coupons）。
- 补偿分支 RPC 全部存在：`UpdateOrder2PaymentSuccessRollback`（order，完整实现：
  行锁+状态机，rollbacklogic.go:69）、`ReturnInventory`（inventory，注释明示
  "退还库存（支付失败时）"）、`ReleaseCoupon`（coupons）；checkout 亦备
  `UpdateStatus2OrderRollback`。
- **缺失三件**：所有 go.mod 无 `dtm-labs/dtmclient`；SQL（construct/depend/sql/）无
  `dtm_barrier` 表；saga 驱动代码零行（lockcouponlogic.go:86 注释可证当初刻意绕开）。

**实证 12（delay consumer 补充雷点）**：`delay/consumer.go:104` 用 `Ack(true)`（multi=true 误用）；
同款 fall-through（:96-101 后无 continue）→ :109 nil-deref；幂等键同样失败不释放。

**实证 13（audit consumer，同款问题但不在 scope）**：autoAck 已是 false（相对健康），
但幂等键失败不释放（同洞 2，TTL 24h）、无监督循环。独立服务，列为已知同款问题，
本次不修（避免 scope 稀释），后续可复用 consumerx 原语单独修。

## 修订决策（最终版）

| # | 决策点 | 第一轮 | 第二轮（最终） |
|---|---|---|---|
| D2 | 洞1 发布可靠性 | B 快路径补发 SettleOrder | **DTM Saga 化**（激活存量分支+补偿）；SettleOrder 决策作废 |
| D3 | DTM 档位 | 档1 本次 | **档3 全 Saga 化**（结算链路 MQ 编排退役） |
| D7 | payment-PAID 位置 | — | **A：本地 PAID + 后发 saga**。快路径 payment.go:210-213 变 EnsureSaga：确定性 gid `settle:{orderId}`，重复提交被 dtm 拒绝 = 幂等补建。残余风险（PAID 成功但建 saga 失败且 Stripe 不重试）用户已知并接受 |
| D8 | order/notify MQ | （原计划骨架改造） | **彻底删除**：publish（updateorder2paymentsuccesslogic.go:108-118）+ consumer + 队列/交换机声明。C12 教训：不留带病休眠 |
| D9 | 幂等键生命周期 | (a) 失败 Del | **两态键 lease 版**：`SETNX key=PROCESSING PX=leaseTTL` → 成功 `SET key=SUCCESS PX=24h` / 失败 `DEL key` + `INCR retry:{key}`（超限弃单+告警）。**PX 过期即 stale 检测**，无需 timestamp 比对。落为 consumerx 通用原语，服务 order/delay 与未来消费者 |
| D10 | DLX/DLQ | 否决 | 维持不加：弃单+结构化告警承接（delay 链路低频低风险） |
| D11 | change 处置 | — | **修订现有 change**（本文件即新基线），不拆分 |

## 新架构基线

```
Stripe webhook (payment.go)
  → payment DB → PAID（本地事务，现状不动 = D7 方案 A）
  → 发起 Saga（gid = settle:{orderId}，确定性）
      ├─ 分支1 UpdateOrder2PaymentSuccess   ⇄ UpdateOrder2PaymentSuccessRollback
      ├─ 分支2 DecreaseInventory            ⇄ ReturnInventory
      └─ 分支3 UseCoupon                    ⇄ ReleaseCoupon
  DTM 保证：分支自动重试 + barrier 挡分支重放 + 失败逆序补偿
快路径（:210-213）→ EnsureSettlementSaga（幂等补建）
order/notify MQ → 整条退役（D8）
consumerx 骨架范围缩至：order/delay（必须）+ payment/delay（休眠骨架）
```

**分支幂等关键点**：`UpdateOrder2PaymentSuccess` 状态机把"已 Paid"判为 Aborted 拒绝，
作为 saga 分支重试时会被误判失败触发补偿 → **barrier 表是必须项不是可选项**（挡分支重放），
叠加库存"已扣减"容忍、券"已使用"容忍构成双保险。分支函数须接 dtm barrier 中间件。

## Checklist 修订（审核基线以本节为准）

- [ ] **C0/C1/C2**（channel/fall-through/监督）：范围缩至 **order/delay**（notify 退役，payment/delay 仅骨架）；实证 12 的 Ack(true) 一并修。
- [ ] **C3** 幂等：升级为 **D9 两态键 lease 版**，consumerx 通用原语。
- [ ] **C4** 重试上限：保留（Redis INCR + 超限弃单告警），适用于 delay 链路。
- [ ] **C5** 毒消息：一次即弃+结构化告警维持；DLX 不加（D10）。
- [ ] **C6**（洞1 兜底）：改为 **Saga 化**——结算由 DTM 必达保证；快路径 EnsureSaga 幂等补建（D7）；SettleOrder 相关任务作废。
- [ ] **C6a（新增）**：dtmclient 依赖接入全部涉及服务 go.mod；`dtm_barrier` 表迁移（order/inventory/coupons 三库）；saga 驱动代码 + 确定性 gid 策略；分支函数接 barrier；已 Paid 不得误判分支失败（barrier+业务容忍双保险）。
- [ ] **C6b（新增）**：order/notify MQ 全链删除（D8）后，`go build`/`make lint` 确认无残留引用；queue/exchange 声明清理。
- [ ] **C7** 幂等 TTL 风险：notify 退役后按两态键语义重新表述（PROCESSING lease TTL 配置化；SUCCESS 24h）。
- [ ] **C8** 业务幂等不回归：订单状态机/库存容忍/券容忍 + **新增分支幂等要求**（见 C6a）。
- [ ] **C9** DDL 红线：saga 驱动若经 payment 服务调用，pb 不进 logic 前置转换。
- [ ] **C10** 测试：mock 先行；新增"saga 分支重试不误补偿""EnsureSaga 幂等"两类场景；delay consumer 保留原两类场景（channel 恢复/失败不双扣）。
- [ ] **C11** 不做清单（更新）：~~DTM~~（移入本次核心）；Outbox、DLX、event.ID 去重、延迟对账扫描、PaymentDelayMQ 复活 仍不做。
- [ ] **C12** PaymentDelayMQ：维持休眠决策不变（决策 7 延续，payment/delay 骨架加固不接线）。
- [ ] **C13（新增）**：audit consumer 同款问题（实证 13）记录为已知问题，显式不在 scope，后续独立修。
