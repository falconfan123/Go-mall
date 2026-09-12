# Design — settlement-fallback-layers

## Context

前序 change（harden-payment-settlement-chain，已归档）完成结算 Saga 化：webhook 本地 PAID →
EnsureSaga → dtm 编排三分支（含 barrier 幂等）。残余缺口为四类静默失败（explore-brief 实证），
本 change 以 Outbox/扫描器/Stripe 对账/人工产出补齐。关键既有事实：

- `payment.go:194` 支付单翻转为裸单行更新（`PaymentsModel.Update`，无事务）；`PaymentsModel`
  接口现状无 WithSession（需按 inventory 先例补事务变体）。
- payment 状态枚举（pb）：UNPAID=1 / PAID=2 / FAILED=3 / FULLY_REFUNDED=4 / EXPIRED=5，
  DB 存 bigint。order：PendingPayment=2 / Paid=3 / Cancelled=5 / Closed=6。
- delay handler 实测关单业务 = **状态机翻转（Created→Closed/Expired）+ ReturnPreInventory**
  （不含 ReleaseCoupon——修正 explore-brief 旧表述），内联于 delay handler。
- dtm HTTP query API 现成：`GET {dtm}:36789/api/dtmsvr/query?gid=`，返回含 status；
  dtm 默认 http=36789 / grpc=36790。
- Stripe SDK（stripe-go v81）已接（processor.go），CheckoutSession 带
  `AmountTotal/Currency/Metadata(order_id/payment_id)`；`normalizeStripeTransactionID`
  产出 ≤64 字符交易号，即对账 join 键（payments.transaction_id varchar(255)）。
- PaymentsModel 已有 `WithSession` 基础能力（paymentsmodel.go:50-52），缺的是带状态机
  门槛并返回影响行数的条件更新方法（决策 4）。
- 所有业务表同库 mall；无 cron/调度基建；payment 已有常驻循环先例
  （webhook HTTP、PaymentDelayMQ 消费骨架）与优雅停机先例（proc.AddWrapUpListener）。

## Goals / Non-Goals

**Goals:** outbox 必达接力（治②）；扫描自愈/分流（治③④）；Stripe 双向对账（治⑤⑥）；
异常单聚合产出（人工层）。全部常驻任务接入优雅停机与结构化日志。

**Non-Goals:** 退款实现（另立项）；PaymentDelayMQ 复活；dtm 多实例改造；新对账数据源；
管理界面（异常单处置用表语义 + 运维 SQL，面板另立项）；不改既有幂等机制。

## Decisions

### 决策 1（D6 落定）：relay/scanner/对账宿主 = payment 服务内常驻循环

三类后台任务（relay、scanner、对账）以独立 goroutine 常驻于 payment 服务，统一由一个
fallback worker 聚合生命周期：`ctx` 经 `proc.AddWrapUpListener` 接入优雅停机（先例照抄），
每轮执行外层 `defer recover()`（panic 记 error 日志后本轮回收，循环不死——参照 consumerx
safeHandle 模式）。

- **备选**：独立 reconciler cmd（独立部署单元）——多一套部署/配置/观测成本，任务量级
  （支付频次级）不值，留为后续拆分备选，否决。
- relay tick 默认 10s；scanner tick 默认 5min；对账每日一次（默认 02:00 本地时区），
  全部 yaml 配置化（决策 9）。

### 决策 2（D7 落定）：人工待办 = `settlement_exception` 表 + 结构化告警双写

```sql
CREATE TABLE settlement_exception (
  id bigserial PRIMARY KEY,
  order_id varchar(64) NOT NULL,
  payment_id varchar(64),
  sources jsonb NOT NULL DEFAULT '[]',  -- 发现来源集合（relay|scan1|scan2|recon），upsert 去重追加
  reason varchar(255) NOT NULL,
  detail jsonb,
  status varchar(16) NOT NULL DEFAULT 'open',   -- open|resolved
  first_seen_at timestamptz DEFAULT now(),
  last_seen_at timestamptz DEFAULT now(),
  seen_count int NOT NULL DEFAULT 1,
  resolved_at timestamptz,
  resolved_note varchar(255)
);
CREATE UNIQUE INDEX uniq_exception_open_order ON settlement_exception (order_id) WHERE status='open';
```

- **聚合去重靠 partial unique index**：同 order_id 的 open 待办 upsert（seen_count+1、
  last_seen_at、sources jsonb 数组去重追加），三层重复发现天然收敛为一条——这是选表方案
  的核心理由（纯日志无法去重聚合，留备选否决）。表另含 payment_id 字段（对账方向2 直接
  关联支付单，免 join）。
- 处置语义：`UPDATE ... SET status='resolved'`（运维 SQL，文档化；管理面板另立项）；
  resolved 后同单新异常会再开新待行（符合"处置后不重复告警、复发再告"直觉）。

### 决策 3：`payment_outbox` 表结构与写入语义

```sql
CREATE TABLE payment_outbox (
  id bigserial PRIMARY KEY,
  order_id varchar(64) NOT NULL,
  payment_id varchar(64) NOT NULL,
  user_id bigint NOT NULL,
  payload jsonb NOT NULL,               -- SettlementInput 快照（发起时数据定格）
  status varchar(16) NOT NULL DEFAULT 'pending',  -- pending|done|dead
  attempts int NOT NULL DEFAULT 0,
  next_retry_at timestamptz,
  done_at timestamptz,                  -- 审计 relay 接力延迟
  last_error text,
  created_at/updated_at timestamptz DEFAULT now()
);
CREATE UNIQUE INDEX uniq_outbox_order ON payment_outbox (order_id);
CREATE INDEX idx_outbox_pending ON payment_outbox (next_retry_at) WHERE status='pending';
```

- **payload 快照**：relay 调 EnsureSaga 所需全部输入（orderId/userId/transactionId/
  paidAmount/paidAt + items/couponId/金额，webhook 发起时已从 GetOrder 取得），以
  **encoding/json** 序列化（`SettlementInput` 为手写 Go struct 非 pb 消息，protojson 不适用；
  这是内部存储格式而非对外契约，字段命名混合 CamelCase/pb 导出名由写读两端一致用
  encoding/json 保证）。快照而非 relay 时重查——relay 不依赖
  order 服务可用性；订单行不可变，快照不会过期。
- **gid 后缀扩展（墓碑重试的基础设施，先于决策 6 声明）**：`SettlementInput` 新增
  `GidSuffix string` 字段（空串 = 原始 `settle:{orderId}`）；`SettlementSaga.gid()` 拼接
  后缀。**SettlementSagaAPI 接口签名不变**（Initiate/Ensure 仍只收 SettlementInput）——
  webhook/relay/mock/单测零改动；仅扫描器墓碑重试时构造 `GidSuffix=":r2"`。
- 写入条件：**仅翻转事务内写**（payment 本次由非 PAID→PAID 时 INSERT，`ON CONFLICT
  (order_id) DO NOTHING` 兜并发重复 webhook）；重复 webhook（已 PAID）不写——outbox 行
  在首次翻转时已存在。
- done 行永久保留（审计）；dead 行由人工/后续工具重置为 pending 重放（本期不做重放工具）。

### 决策 4：webhook 两写事务改造

`processStripePaymentSuccess` 的翻转段改为：

```
TransactCtx:
  UPDATE payments SET ...PAID... WHERE payment_id=? AND status <> PAID   -- 状态机门槛
  INSERT INTO payment_outbox ... ON CONFLICT (order_id) DO NOTHING
```

- **状态机门槛为条件更新**：`PaymentsModel` 新增 `UpdateStatusToPaidWithSession(ctx,
  session, paymentId, transactionId, paidAmount, paidAt) (rowsAffected int64, err error)`
  ——SQL 带 `AND status <> PAID` 门槛并返回影响行数（WithSession 基础能力已有，缺的是这个
  带条件方法，**不得**复用现有无门槛 `Update`，否则并发 webhook 丢失状态机门槛）。
  `rowsAffected==1` 才 INSERT outbox（翻转者负责写待办）；==0 说明并发已翻转，跳过。
  payload 所需的 order 快照（items/coupon/金额）在**事务外**经 GetOrder 预取后传入
  （事务内不跨服务）。
- GetOrder 预取失败 → 不进入事务、不翻转（返回错误 → Stripe 重试，幂等安全）。
- 快路径（已 PAID）不写 outbox、不进事务——EnsureSaga 直接兜底（现状语义保持）。
- outbox 写失败回滚整个翻转 → webhook 500 → Stripe 重试 → 重写，安全。

### 决策 5：relay 循环（治②）

每 tick：`SELECT ... WHERE status='pending' AND next_retry_at <= now() ORDER BY id LIMIT 50
FOR UPDATE SKIP LOCKED`（同实例/多实例均安全，决策 1 的行为化落点）→ 逐条调
`EnsureSaga(payload)` → 成功置 done（`done_at` 落库，便于审计 relay 延迟）；失败 `attempts+1`、`next_retry_at=now()+2^attempts
基值`、超过上限（默认 5）置 dead + error 告警 + 产出 `settlement_exception(source=relay)`。
**EnsureSaga 提交成功即 done，不等待 saga 终态**（与 webhook 快路径并发时靠幂等收敛）。

### 决策 6：scanner 循环（治③④）

单循环 5min tick，每 tick 依序执行。**时间比较基础**：`payments.paid_at` 为 bigint
（unix 秒），扫描 SQL 统一以 `p.paid_at < floor(extract(epoch from (now() - interval '10 minutes')))::bigint`
参数化比较（bigint 对 bigint，规避 numeric 隐式转换；禁止 naive timestamptz 字符串比较
——防 dtm 式时区坑；如需索引建表达式索引）：

- **scan1**：`payments(status=PAID, paid_at 超过 10min 阈值) JOIN orders`（补结算所需的
  items 由扫描器同库按 order_id 聚合 `order_items` 构造——与 delay handler 的
  QueryOrderItemsByOrderID 同源），按订单状态分流：
  - PendingPayment → 查 dtm status（HTTP query API）：无记录 → EnsureSaga(原始 gid)；
    submitted/prepared → 记录干等；**succeed → 数据矛盾（订单非 Paid 却 succeed，理论
    不可达）→ 按"异常单人工产出"需求处理（exception, source=scan1, reason=数据矛盾）**；
    failed（墓碑）→ 依次探测 `settle:{orderId}` / `:r2` / `:r3` 找最大
    已用后缀 → 以 `:r(n+1)` 调 Initiate（上限 3，超限 exception 转人工）。
  - Closed/Cancelled → `settlement_exception(source=scan1, reason=退款)` 产出，不补结算。
- **scan2**：`orders(order_status=PendingPayment, payment_status=Paying, created_at<now-35min)`
  → 调 order 服务关单入口（决策 7 的 RPC）；关单后钱到账的竞态由状态机语义收敛（转 scan1
  下轮的 Closed+PAID 分支产出退款待办）。
- **边界声明**：scan1 补结算成功（含墓碑 r2/r3）**不自动关闭** relay 已产出的 dead
  exception 待办——异常处置以订单最终状态为准（人工核实时订单已正常，标记 resolved 即可）。
- **scan3（可选增强，本期实现从简）**：dtm in-flight 超龄告警——经 `/api/dtmsvr/all` 按状态
  过滤（实现从简 = 本期仅留配置开关，默认关）。

### 决策 7：order 关单入口抽取（scan2 依赖，order 服务先发）

order proto 新增 `rpc CloseExpiredOrder(CloseExpiredOrderRequest) returns (EmptyRes)`
（Rule 3 生成；入参 orderId/userId），logic 封装现有 delay handler 的两步业务（状态机
Created→Closed/Expired + ReturnPreInventory，幂等容忍语义原样保持）。**delay handler 迁移为
调用同一 logic**（消除内联复制，consumerx 骨架不动）。payment scanner 经 zrpc 调用该 RPC。
幂等响应语义：订单非 Created（已支付/已关闭）→ 返回 Success + StatusMsg="already closed,
skip"（与 delay handler 的跳过语义一致，scan2 重试安全）；ReturnPreInventory 失败 → 返回
错误（scan2 下轮重试，天然间隔为 scanner tick；连续 3 轮失败（配置化）→ 产出 exception
(source=scan2) 并告警，本单仍继续尝试）。

**实施期扩展（方案 1，2026-09-06，探针④暴露）**：scan2 的目标孤儿形态是
PendingPayment(2)（支付链接已开出但超时未付，其 payment 超时兜底原由休眠的
PaymentDelayMQ 负责），而原状态机只关 Created(1)——语义缺口经用户拍板以**方案 1**闭合：
状态机扩展为 `Created OR PendingPayment → Closed/Expired`（一行前置扩展）。delay 队列
只发 Created 单，行为不受影响；skip 语义仅限其他状态（已支付/已关闭）。

### 决策 8：Stripe 双向对账（治⑤⑥）

- 拉取：`CheckoutSessionList(created >= T-1 00:00 && < T 00:00, status=complete)` 分页遍历；
  join 键 = normalizeStripeTransactionID(**PaymentIntent.ID 优先、其次 session.ID**——与
  webhook 取号顺序 payment.go:129-132 一致，防止对账与写路径取号不一致导致 join 失败) ==
  payments.transaction_id。
- **方向1**（Stripe 有、本地未 PAID）：校验 `session.AmountTotal == order.PayableAmount`
  （经 GetOrder；GetOrder 不可得/金额/币种不一致 → exception 转人工，**禁止按差异金额重放**）→
  复用 `processStripePaymentSuccess` 同逻辑重放（注入态上下文，内部幂等）→ 失败 exception。
  重放 `paidAt` 取 **session.Created**（Stripe 侧扣款时间，unix 秒）而非本地接收时间——
  与真实支付时间一致，避免污染 scan1 的时间窗判断。
- **方向2**（本地 PAID、Stripe 无账）：exception（source=recon，reason=Stripe 无账）+ 告警，
  不自动回滚。
- 对账运行窗口可配（默认 T-1 日终 02:00）；重复发现由 exception 表聚合（决策 2）。
- 结构化对账产出：差额明细进 detail jsonb + error 日志（含订单标识、方向、两侧金额）。

### 决策 9：配置与生命周期

payment yaml 新增 `Fallback` 段：relayTickMs(10000)/scannerTickMs(300000)/scan1ThresholdMinutes(10)/
scan2ThresholdMinutes(35)/relayMaxAttempts(5)/tombstoneMaxRetry(3)/reconHour(2)/reconEnabled(true)。
全部零值回落默认。生命周期：worker ctx 挂 AddWrapUpListener；停机时当前轮跑完即退出
（循环任务非消息语义，无 unacked 概念）；每轮 panic recover。

## Risks / Trade-offs

- [outbox 写失败回滚 PAID 翻转] → webhook 500 + Stripe 重试重写，语义安全（决策 4）。
- [relay/scanner 与 webhook 并发触发 EnsureSaga] → 确定性 gid + dtm 静默接受收敛，幂等已备。
- [scan2 与迟到支付竞态] → 行锁 + 状态机收敛；关单后钱到 → scan1 下轮 Closed+PAID 分支
  产出退款待办（不丢钱）。
- [墓碑 r2/r3 重试间隔] → scanner tick 5min 为天然间隔，连续失败 3 次即 ~15min 内三次
  全量尝试，下游故障时压力可接受（结算分支幂等）；如需加大间隔进配置。
- [dtm query API 不可达] → scan1 保守跳过该单 + error 日志，不误判（不无中生有发 r2）。
- [对账窗口时钟偏差/首日回扫] → 窗口按 created 分页幂等，重复发现被 exception 聚合；
  首日回扫天数实施期定（默认 7 天）。
- [跨域读（scanner 读 orders/payments 同库）] → 对账作业惯例；写路径全部走服务 RPC
  （关单走 order RPC、补结算走本服务 EnsureSaga），不越域写。
- [异常单表膨胀] → resolved 行保留（审计），量级低频；清理策略后续增强。

## Migration Plan

1. DDL：`payment_outbox` + `settlement_exception`（construct/depend/sql，mall 库）。
2. order 服务：CloseExpiredOrder RPC（proto 生成 Rule 3）+ delay handler 迁移复用 →
   `make mock && make lint && make test-unit && make build` → **先部署**。
3. payment 服务：两写事务 + relay/scanner/对账 + Fallback 配置 → 门禁全绿 → **后部署**。
4. 上线后观察项：outbox pending 峰值、dead 计数、exception 产出量、首日对账差额。
5. 回滚：回退 payment 二进制（relay/scanner 停，outbox 表残留无害——新写入停止，旧语义
   Stripe 重试 + EnsureSaga 仍兜底）；order 回退需在无扫描器运行时进行（运维窗口）。

## Open Questions

（无阻塞项——对账首日回扫天数、scan3 开关默认值进配置，实施期可调；dtm query 响应的
解析字段以 v1.19.0 实测为准（spike 已证 HTTP API 可用），实施期以真实响应定型解析器。）
