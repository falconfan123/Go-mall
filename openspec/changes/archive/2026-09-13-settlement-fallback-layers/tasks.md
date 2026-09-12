> 全部任务粒度 ≤2h（审查基线约束）；验证命令均在对应服务目录执行。

## 1. DDL 与基建（先于代码）

- [x] 1.1 `payment_outbox` 表迁移：construct/depend/sql 新增脚本（字段/唯一索引 uniq_outbox_order/partial 索引 idx_outbox_pending，照 design 决策 3 DDL）。验证：mall 库表结构核对
- [x] 1.2 `settlement_exception` 表迁移：sources jsonb / payment_id / partial unique index uniq_exception_open_order（决策 2 DDL）。验证：mall 库表结构核对
- [x] 1.3 payment yaml `Fallback` 配置段 + config struct：relayTickMs/scannerTickMs/scan1ThresholdMinutes/scan2ThresholdMinutes/relayMaxAttempts/tombstoneMaxRetry/reconHour/reconEnabled（决策 9，零值回落默认）；`DtmConfig` 新增 `HttpAddr`（dtm query API :36789，墓碑探测依赖）。验证：`go test ./services/payment/internal/config/...`

## 2. payment 模型与骨架（webhook 两写的基础）

- [x] 2.1 `PaymentsModel.UpdateStatusToPaidWithSession`：带 `AND status <> PAID` 门槛 + 返回影响行数（决策 4）；补单测（首翻 rows=1 / 并发已翻 rows=0）。验证：`go test ./dal/model/payment/...`
- [x] 2.2 `SettlementInput.GidSuffix` 字段 + `SettlementSaga.gid()` 拼接（接口签名不变，决策 5 前置声明）；outbox payload 序列化（encoding/json）roundtrip 单测；gid 拼接断言（空后缀/`:r2`/`:r3`）。验证：`go test ./services/payment/internal/svc/...`
- [x] 2.3 fallback worker 骨架：relay/scanner/对账三循环统一宿主（ctx 挂 AddWrapUpListener、每轮 panic recover、tick 配置化，决策 1/9）；优雅停机单测（cancel 后**当前轮跑完即退出**，不立即中断）。验证：`go test ./services/payment/...`

## 3. webhook 两写事务（C1，治②）

- [x] 3.1 `processStripePaymentSuccess` 翻转段改造：TransactCtx 内条件更新（rowsAffected==1 才 INSERT outbox，ON CONFLICT DO NOTHING）；GetOrder 预取移到事务外（决策 4）。验证：单测（首翻写 outbox/并发翻转只一条待办/outbox 失败回滚翻转/已 PAID 不写）
- [x] 3.2 relay 循环：SKIP LOCKED 批量取回 → EnsureSaga → done（done_at）；失败退避 2^attempts、超限 dead + exception(source=relay) + 告警（决策 5）。验证：单测（接力成功/失败退避/超限 dead/崩溃续扫——状态落库断言；**done 语义显式断言**：置 done 后不再二次处理、不等待 saga 终态）
- [x] 3.3 快路径与 relay 并发幂等回归：EnsureSaga 双发收敛断言（复用现有 fake Saga 单测扩展）。验证：`go test ./services/payment/...`

## 4. order 关单入口（C5 前置，order 先发）

- [x] 4.1 order proto 新增 message `CloseExpiredOrderRequest`（order_id/user_id）+ `rpc CloseExpiredOrder(CloseExpiredOrderRequest) returns (EmptyRes)`（Rule 3 生成）；logic 封装 delay handler 两步业务（状态机 Created→Closed/Expired + ReturnPreInventory，幂等响应语义照 design 决策 7）；server 注册。验证：logic 单测（首关/重复关 skip/ReturnPre 失败透传）+ `make build`
- [x] 4.2 delay handler 迁移调用同一 logic（消除内联复制，consumerx 骨架不动）。验证：delay handler 测试随实现迁移重构（consumerx 骨架行为保持不变，断言改为面向 CloseExpiredOrder 的幂等/失败透传）+ `go test ./services/order/internal/mq/delay/...` + `make mock`

## 5. 扫描器（C4/C5，治③④）

- [x] 5.1 scan1 主流程：epoch 参数化 SQL（floor(...)::bigint）JOIN orders 分流（PendingPayment/Closed+Cancelled）；SettlementInput 全字段构造——orders 表取 coupon_id/original_amount/discount_amount/user_id/pre_order_id，order_items 同库聚合 items（决策 6）。验证：单测（mock 行数据：三类分流各一 + SettlementInput 字段完整性断言）
- [x] 5.2 墓碑探测与递增重试：dtm query API 解析器（HTTP GET，响应解析以 v1.19 实测定型）→ 无记录 EnsureSaga / in-flight 干等 / failed → GidSuffix=r{n+1}（上限 3）/ succeed → exception(数据矛盾)；dtm 不可达保守跳过+日志（决策 6）。验证：单测（fake dtm HTTP：五分支各一）
- [x] 5.3 scan2：35min 超龄单 → CloseExpiredOrder RPC；失败重试 3 轮超限 → exception(source=scan2) 并告警，**exception 后仍继续每轮重试（不限次），不跳过该单**；竞态分流断言（决策 6/7）。验证：单测（mock RPC：关闭成功/幂等 skip/连续失败产 exception；"关单后 payment 已 PAID → scan1 下轮产出 exception"的分流断言；完整竞态收敛归 7.2 探针）
- [x] 5.4 exception 产出聚合：upsert（sources 去重追加/seen_count+1/resolved 后新开行）+ 结构化告警（决策 2）。验证：单测（跨层重复发现聚合断言 + seen_count 递增时不重复触发告警的断言）

## 6. Stripe 对账（C6，治⑤⑥）

- [x] 6.1 拉取与 join：CheckoutSessionList 分页（T-1 窗口）+ join 键 PaymentIntent.ID 优先（与 webhook 一致）+ normalizeStripeTransactionID（决策 8）。验证：单测（mock Stripe：分页/键归一/join 键优先序与 webhook 一致断言）
- [x] 6.2 方向1 自动重放：金额/币种校验（AmountTotal vs PayableAmount，不一致停止转人工）→ 复用 processStripePaymentSuccess 重放（paidAt=session.Created）→ 失败 exception（决策 8）。验证：单测（重放成功/金额不一致拦截/**币种不一致拦截**/重放失败转人工）
- [x] 6.3 方向2 冻结告警：本地 PAID − Stripe 无账 → exception(source=recon) + 告警不回滚（决策 8）。验证：单测（差集生成/不触发回滚断言）

## 7. 门禁与验证

- [x] 7.1 全仓门禁：`make mock && make lint && make test-unit && make build && make gatekeeper` 全绿（Rule 1）。验证：命令全绿记录
- [x] 7.2 本地集成探针（对照 specs Scenario）：① outbox 全链（翻转→relay→EnsureSaga→done，kill relay 续扫）② scan1 孤儿单自动补结算（手工构造 PAID+Pending 超 10min）③ 墓碑单 r2 重试成功 ④ scan2 超龄单自动关单 + 迟到支付竞态 ⑤ 对账方向1/方向2 差额产出（Stripe 测试环境构造）。验证：手工执行并记录日志证据
- [x] 7.3 对照 explore-brief Checklist C1-C13 逐项勾验，结论追加 review-log.md。验证：review-log 更新完成
