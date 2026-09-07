## 1. DTM 基建（决策 1/3，先于一切分支代码）

- [x] 1.1 dtm Store 补齐与版本 pin：`construct/depend/conf/dtm.yaml` 与 `manifests/dtm.yaml` 增加 MySQL Store 段（复用 compose 现有 MySQL）；compose/manifests 的 `yedf/dtm:latest` pin 为与 dtmclient 匹配的具体版本。验证：`docker compose -f configs/docker-compose.dev.yml --profile optional up dtm` 启动成功且日志显示 MySQL 存储就绪，重启容器后 Saga 进度不丢
- [x] 1.2 `dtm_barrier.barrier` 表迁移：`construct/depend/sql/` 下新增 `dtm_barrier.sql`（文件头注释注明须对 order/inventory/coupons 三库分别执行）。验证：三库 `dtm_barrier.barrier` 表存在且结构与 dtm 官方 SQL 一致
- [x] 1.3 dtmclient 依赖接入：order/inventory/coupons/payment 四服务 go.mod 新增 `dtm-labs/dtmclient`（含 dtm-driver-gozero），版本与 1.1 pin 的 server 匹配；执行 `go mod tidy`（common + 四服务）。验证：`go mod tidy && go build ./...`（四模块 + common）通过
- [x] 1.4 发起端最小 spike（临时代码/测试，结论记录到 review-log 不进主干）：本地 dtm 上用 `NewSagaGrpc` 跑通两分支 Saga（一成功一补偿），并验证重复 gid 提交的错误语义（错误可区分、"gid 已存在"可映射成功）。验证：spike 脚本/测试运行通过，结论写入 review-log
- [x] 1.5 配置扩展：order/inventory/coupons/payment yaml 与 config struct 新增 DTM target；`IdempotencyTtlSeconds` 语义收窄为 lease TTL（注释明确），新增 `SuccessTtlSeconds`（默认 86400）。验证：conf.MustLoad 加载成功 + config 单测

## 2. consumerx 两态键重构（决策 4，洞 2 完整解）

- [x] 2.1 `common/utils/idempotency` 新增状态原语：`Claim(ctx, rds, key, leaseTTL) (State, error)`（Lua SETNX PROCESSING PX，旧值 `"1"` 识别为 Processing）、`MarkDone(ctx, rds, key, successTTL)`（SET SUCCESS PX）；**保留 `CheckAndSet`/`Release` 不动**（audit 仍在用，C13）。验证：`go test ./common/utils/idempotency/...`（含旧值兼容、租约过期重入、MarkDone 后跳过用例）
- [x] 2.2 `common/mq/consumerx` Store 接口状态化：`Claim → (State, error)`、新增 `MarkDone`，`Release`/`IncrWithTTL` 保留；`Process` 流程调整——Processing/Done → Ack 跳过，Handle 成功 → MarkDone + Ack，失败 → Release + IncrWithTTL；**重试未超限时输出可恢复级 Info 结构化日志（含幂等标记删除事件）并 Reject(true) 重投，超限时 Reject(false) 弃单并输出 error 日志**（specs"幂等标记删除"与"有界重投"Scenario）。验证：`go test ./common/mq/consumerx/...` 覆盖四条路径（毒消息/重复跳过/成功/失败重投与超限弃单）+ 旧值 `"1"` 兼容 + 日志级别可区分断言
- [x] 2.3 消费者编译适配：验证 `services/order/internal/mq/delay/consumer.go`、`services/payment/internal/mq/consumer.go` 经 consumerx 重构后编译通过（两者为 Handler[T] 实现，预计无需改代码；若编译失败优先检查 consumerx 回调签名是否变化）；如 Config 新增字段（如 SuccessTTL）需装配则一并补全。payment 侧 Handle 仍为空实现保留扩展点，`PaymentMQ` 保持 nil 不接线（决策 6）。验证：`go test ./services/order/... ./services/payment/...` + `make build`

## 3. Saga 分支接线（决策 2/3，三服务）

- [x] 3.1 分支 1（order）：`updateorder2paymentsuccesslogic.go` 接 dtm gRPC barrier；`:57-61` 订单状态前置检查拆分为三分支判定——PendingPayment 正常流转、**已 Paid 返回成功（幂等重放）**、其他状态维持 Aborted；`:44-48` 订单不存在与 `:64-69` payment status 分支保持失败语义并补悬挂风险注释（design 决策 2）。本任务不触及 MQ publish 删除（归 4.1）。验证：`go test ./services/order/internal/logic/...`（mock OrderModel/barrier：PendingPayment 流转、已 Paid 重放成功、Closed 冲突失败、订单不存在失败）
- [x] 3.2 分支 2/3（inventory/coupons）：`decreaseinventorylogic.go`、`returninventorylogic.go`、`usecouponlogic.go`、`releasecouponlogic.go` 接 barrier；确认"已扣减/已使用"容忍响应在分支语境按成功返回（specs"业务幂等兜底"契约）。验证：`go test ./services/inventory/... ./services/coupons/...`（mock Model：重放 barrier 命中短路、容忍响应成功、真实失败返回错误）
- [x] 3.3 补偿分支可用性核验：`UpdateOrder2PaymentSuccessRollback`/`ReturnInventory`/`ReleaseCoupon` 以 Saga 分支调用方式联调（barrier 接线后），确认对不存在/未生效场景的返回语义与 design 决策 2 的悬挂风险声明一致。验证：`go test ./services/order/... ./services/inventory/... ./services/coupons/...` 覆盖补偿路径
- [x] 3.4 `make mock` 更新 order/inventory/coupons mock（若 5.3 由同一人连续执行，可合并执行一次）。验证：`make mock && make test-unit`

## 4. order/notify MQ 退役（决策 5，BREAKING）

- [x] 4.1 删除 publish：`updateorder2paymentsuccesslogic.go:108-118` 调用与 `notify` 包 import（**本任务执行，3.1 不触及**）。验证：`rg notify services/order/internal/logic/updateorder2paymentsuccesslogic.go` 无结果 + `go test ./services/order/internal/logic/...` 通过
- [x] 4.2 删除 `services/order/internal/mq/notify/`（consumer.go/init.go/product.go）与 `svc/servicecontext.go` 的 `OrderNotifyMQ` 字段、`initOrderNotifyMQ` 装配及相关 config 项。验证：`make build` 通过
- [x] 4.3 残留扫描：全仓 grep `order-notify`/`OrderNotifyMQ`/notify 队列声明无业务残留（历史文档/注释除外）。验证：`make lint` 绿 + grep 记录

## 5. payment Saga 发起与快路径（决策 1，Rule 4）

- [x] 5.1 svc 层 `SettlementSaga` 封装：`Initiate(orderId, userId, paymentResult, items, couponId)`（三分支 + 补偿注册、gid=`settle:{orderId}`、Submit）与 `Ensure(...)`（重复 gid 错误映射为成功）；pb 转换内聚 svc，不外泄 main 包。验证：`go test ./services/payment/internal/svc/...`（mock dtm client：首次发起、重复发起幂等、发起失败透传错误）
- [x] 5.2 `payment.go` 接入：本地 PAID 成功后调用 `Initiate`，失败返回错误（webhook 500 触发 Stripe 重试）；快路径（订单已 Paid）改为调用 `Ensure` 后返回。验证：`go test ./services/payment/...`（依赖注入 mock OrderRpc/SettlementSaga：首次路径、快路径 Ensure 必达、发起失败 500 语义）
- [x] 5.3 `make mock` 更新 payment mock（若 3.4 已合并执行则跳过重复部分）。验证：`make mock && make test-unit`
- [x] 5.4 SettleOrder 退役（实施期发现，design 决策 1 附注）：order proto 删除 `SettleOrder` 方法后重新生成 pb/server（Rule 3，注册随生成自动移除，不手编 generated 文件）；删除 `settleorderlogic.go`；payment 快路径已随 5.2 改调 EnsureSaga，全仓无 SettleOrder 残留引用。验证：`rg -i settleorder services/ --type go` 无业务残留 + `make build && make lint` 绿

## 6. 门禁与集成验证

- [x] 6.1 全仓门禁：`make mock && make lint && make test-unit && make build && make gatekeeper` 全绿（Rule 1）。验证：命令全绿记录
- [x] 6.2 本地集成验证（探针式，对照 specs Scenario，`--profile optional` 启动 dtm）：① 正常支付 → 三分支全成功无补偿；② mock 库存瞬时故障 → 分支重试成功不补偿；③ 优惠券不可恢复失败 → 逆序补偿（回补库存/释放券/订单回滚）+ error 日志（构造方式：mock `UseCoupon` 返回业务级错误如 `codes.Aborted`"券不适用"，而非网络超时——网络超时应走分支重试而非补偿）；④ 同订单重复发起 EnsureSaga 幂等；⑤ delay 链路：毒消息一次即弃、失败 3 次弃单告警、`kill -TERM` 优雅退出、管理端强杀 channel（先 `rabbitmqctl list_connections` 定位连接，再 `rabbitmqctl close_connection <name>` 或管理 UI Close）后监督循环自动重连恢复。验证：手工执行并记录日志证据
- [x] 6.3 对照 explore-brief Checklist（第二轮修订版 C0-C13）逐项勾验，结论追加 review-log.md。验证：review-log 更新完成
