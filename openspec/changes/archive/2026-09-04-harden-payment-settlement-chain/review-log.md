# Review Log — harden-payment-settlement-chain

## 审查轮次

### 批次 proposal · 第 1 轮（2026-09-03）

审查对象：proposal.md；基线：explore-brief.md（C0~C12）+ 代码抽查。

- 🔴 D1=b 未如实采纳：What Changes 遗漏 payment/delay consumer 骨架修复，与"骨架可复用、消除半死"的非目标表述自相矛盾 → **已修复**：三 consumer 统一加固写入 What Changes；非目标改为"不接线启用"（PaymentMQ 保持 nil），骨架照修。
- 🟡 C2 未覆盖 `context.TODO()` 消除与 graceful shutdown → **已修复**：监督循环条目补充两点。
- 🟡 C10 测试策略缺失 → **已修复**：新增测试策略原则性声明（mock 先行 + 两类关键单测场景）。
- 🟡 C5 DLX 取舍未表态 → **已修复**：显式声明本次不引入 DLX，取舍记入 design.md。
- 🟡 C7 TTL 风险文档化未指定制品 → **已修复**：明确落在 design.md 量化。
- 💡 C8 业务幂等保护未显式声明 → **已修复**：新增保护既有三重业务幂等条目。
- 💡 What Changes 文件路径写法不一致 → **已修复**：统一为完整相对路径。

其余确认项：D2/D3/D4/D5/D6 如实采纳；事实陈述与代码一致；Capabilities/Impact 合理。

结论：🔴 已清零，proposal 冻结。

### 批次 proposal · 第 2 轮（2026-09-03）

复审：逐条核对第 1 轮 8 项修复，全部 ✅；payment/delay 修复与"不接线启用"非目标自洽；Capability 覆盖一致；C0~C12 全覆盖。

结论：🔴 清零，**proposal 冻结**。

### 批次 specs · 第 1 轮（2026-09-03）

审查对象：specs/async-settlement-reliability/spec.md；对照已冻结 proposal + explore-brief。

- 🔴 业务幂等容忍（库存"已扣减"/券"已使用"返回视为成功）未落成显式契约，实施者可能按字面把容忍路径判失败 → **已修复**：新增 Requirement"业务幂等兜底的副作用返回视为成功"+ 2 个 Scenario，并在"有界重投"语义上明确不覆盖此类响应。
- 🟡 "失败路径可观测"声明四类日志仅 1 个 Scenario → **已修复**：补充毒消息丢弃、幂等标记释放两个 Scenario。
- 🟡 Scenario 泄漏 AMQP 具体错误码 406 → **已修复**：改为通用"协议异常"描述。
- 🟡 幂等 TTL 过期边缘行为缺 Scenario → **已修复**：补充"标记过期后重投按新消息处理"Scenario。
- 💡 Purpose "恰好生效一次"表述与 at-least-once 语义混淆 → **已修复**：改为"业务层面恰好生效一次（at-least-once + 业务幂等收敛）"。

结论：待第 2 轮复审确认。

### 批次 specs · 第 2 轮（2026-09-03）

复审：第 1 轮 5 项修复全部 ✅；业务幂等容忍 Requirement 与确认语义自洽；10 Requirements / 21 Scenarios 格式合规；与 proposal 八项变更点一一对应。

结论：🔴 清零，**specs 冻结**。

### 批次 design · 第 1 轮（2026-09-03）

审查对象：design.md；对照已冻结 proposal/specs + explore-brief C0~C12。

- 🟡 决策6 补发语义缺口（"均可补发"vs 实际间接自愈）→ **已修复**：明确"间接自愈"性质、Stripe 不重试时的残余风险与日志承接、新增"直接自愈"备选否决。
- 🟡 决策5 panic recovery 未说明 → **已修复**：骨架内建 recover，panic 按业务失败处理，不产生悬挂 unacked。
- 🟡 决策5 graceful shutdown 机制不完整 → **已修复**：cancel + 主动关闭连接打断 range + shutdown timeout（默认 10s 配置化）。
- 🟡 决策1 骨架与回调边界模糊 → **已修复**：职责清单 + Decode/IdempotencyKey/Handle 回调接口 + 幂等容忍归回调。
- 🟡 common 新增 amqp 依赖未声明 → **已修复**：Migration 步骤 3 显式化 go mod tidy。
- 🟡 Migration "旧消息按新语义消费"不准确 → **已修复**：改为 autoAck 无未确认残留、新 consumer 消费就绪消息。
- 🟡 决策6 响应结构未定义 → **已修复**：复用 EmptyRes + Aborted/Internal 错误语义。
- 💡 包名暂定 → **已修复**：定名 common/mq/consumerx。
- 💡 Release key 构造 → **已修复**：经 BuildKey，杜绝格式漂移。
- 💡 配置变更清单 → **已修复**：Migration 步骤 4。

确认项：C0~C12 设计层全覆盖；范围无蔓延；prefetch=1/NotifyClose/SettleOrder 幂等依据可行。

结论：🔴 首轮即为零，🟡 已全部修复，待第 2 轮复审确认。

### 批次 design · 第 2 轮（2026-09-03）

复审：第 1 轮 10 项修复全部 ✅；决策 6 与 specs 快路径要求一致；决策 5 与 specs 优雅退出语义一致；无编辑残留。

- 🟡 决策5 graceful shutdown 时序颠倒（先关连接再等待，timeout 形同虚设且完成后无法 Ack）→ **已修复**：改为"cancel → 当前消息完成后停取 → 骨架内 timer 限时等待（默认 10s）→ 超时才关连接强制退出"。
- 💡 "变为 unacked"术语不严谨 → **已修复**：改为"broker 端重新入队待重投"。

结论：🔴 清零，🟡 已修复，**design 冻结**。

### 批次 tasks · 第 1 轮（2026-09-03）

审查对象：tasks.md；对照全部已冻结制品 + explore-brief C0~C12 + Makefile 核对。

- 🔴 5.3 测试路径错误（processStripePaymentSuccess 非导出，test/rpc 是集成测试无法 mock）→ **已修复**：改为 services/payment 包内单元测试 + 依赖注入重构说明，验证命令改为 go test ./services/payment/...
- 🟡 2.2/3.2 生命周期对接缺 go-zero 机制说明 → **已修复**：显式 AddWrapUpListener/等价机制 + 实例保留要求 + 禁止仅换 Background。
- 🟡 consumerx 连接管理策略未明确 → **已修复**：1.3 显式"接收调用方传入连接复用，骨架不 Dial"。
- 🟡 缺 common go mod tidy 任务 → **已修复**：新增 1.6。
- 🟡 5.1 缺 PaymentDelayMQ 回调实现 → **已修复**：补 Decode/IdempotencyKey + Handle 空实现。
- 💡 2.2/3.2/5.1 验证方式含集成范畴 → **已修复**：限定 go test + make build，吞吐验证归 6.2。
- 💡 5.2 paymentResult 映射提示 → **已修复**：补映射说明。

确认项：10 Requirements / 7 Decisions 逐条有承接；C0~C12 全覆盖；Makefile 目标名真实；无越出冻结范围；粒度适中。

结论：🔴 已修复，待第 2 轮复审确认。

### 批次 tasks · 第 2 轮（2026-09-03）

复审：第 1 轮 8 项修复全部 ✅；编号有序、引用一致（3.2→2.2 成立）；5.3 与 design 决策 6 无冲突；决策 5 时序在 1.5/2.2/3.2 一致承接；22 条任务全部 `- [ ] X.Y` 格式且含验证方式；10 Requirements / 21 Scenarios 全部有对应任务。

结论：🔴 清零，**tasks 冻结**。全部四批制品（proposal / specs / design / tasks）审核完毕，规划冻结，可进入 apply。

---

# 修订周期（2026-09-04）— Saga 路线（第 2 版）

修订理由：探索第二轮（explore-brief.md）实证 DTM 地基已备未接线（实证 11），用户拍板
推翻 D2/D3（SettleOrder → DTM Saga）、新增 D7-D11。四批制品按新基线全部重开修订，
轮次从第一周期计数续起。

### 批次 proposal · 第 3 轮（2026-09-04）

审查对象：proposal.md 第 2 版；基线：explore-brief 第二轮（D2/D3/D7-D11、C0-C13）+ 代码抽查。

- 🔴 对代码现状认知失真：第 1 版实施已部分落地未提交（common/mq/consumerx 单态骨架 359 行、
  notify/delay 已迁移为薄适配层、idempotency.Release 已加、common amqp 依赖已存在），
  proposal 按"新建/替换"表述将导致实施路径错误与 D9 两态键无法兑现 → **已修复**：
  主智能体核实 git 工作区现状后重写为第 2.1 版——consumerx 改"重构（单态→两态键）"、
  delay 改"适配"、amqp 依赖改"已存在"，并声明重构策略。
- 🟡 What Changes 的 consumerx/幂等键条目缺 BREAKING → **已修复**。
- 🟡 非目标不自包含（缺 webhook 不重构/状态机不改/延迟对账不做）→ **已修复**：显式补入。
- 🟡 lease TTL 与 SUCCESS TTL 拆分未明确（含既有 IdempotencyTtlSeconds 去向）→ **已修复**：
  保留为 lease TTL（配置化）+ 新增 SUCCESS TTL（默认 24h）。
- 🟡 dtm 在 compose 的 optional profile 未提及 → **已修复**：Impact 注明
  `docker compose --profile optional up`。
- 🟡 C9 saga 驱动放置未声明 → **已修复**：saga 发起与 EnsureSaga 落 payment svc 层封装。
- 💡 "激活存量 RPC"措辞 → **已修复**：改为"纳入 DTM 事务编排"。
- 💡 "已 Paid"响应语义细节 → 转入 design 明确（第 4 轮已核码：返回 codes.Aborted）。

结论：🔴 1 项已修复，待第 4 轮复审。

### 批次 proposal · 第 4 轮（2026-09-04）

复审：第 3 轮 🔴/🟡 全部 ✅；D2/D3/D7-D11 如实映射；第 1 版残留（SettleOrder/Release 单态/
三 consumer 旧范围）清除干净；事实陈述与代码一致（快路径未动、publish 在 :108、dtmclient
未引入、optional profile 属实）。

- 🟡 Impact 遗漏 order svc/servicecontext.go（OrderNotifyMQ 字段与装配删除）与
  updateorder2paymentsuccesslogic.go 的 notify import 清理 → **已修复**。
- 🟡 Impact 遗漏 updateorder2paymentsuccesslogic.go "已 Paid"返回语义调整（现返回
  codes.Aborted，作 Saga 分支会被 DTM 误判失败触发错误补偿）→ **已修复**（主智能体核码
  确认 :58 + :102-105 语义后写入）。
- 🟡 delay 条目"修 Ack(true) 误用"措辞与现状偏差（第 1 版已消除；payment/mq 实际问题是
  autoAck 混用，实证 0）→ **已修复**：改为"消除 autoAck 与手动确认混用"。
- 💡 delay 确认语义 BREAKING 已随第 1 版落地 → **已采纳**：Impact 行为条目注明无新增变更。
- 💡 非目标补"不调整状态机前置校验逻辑" → **已采纳**。

结论：🔴 清零，🟡 已修复，**proposal 冻结**（第 2.1 版）。

### 批次 specs · 第 3 轮（2026-09-04）

审查对象：specs/async-settlement-reliability/spec.md 修订版（11 Requirements / 33 Scenarios）；
对照已冻结 proposal 第 2.1 版 + explore-brief 第二轮基线；核对第 1 版 specs 修复无退化。

- 🟡 两态键缺"处理中租约未过期时重投"边界 Scenario → **已修复**：补充跳过语义 Scenario。
- 🟡 "所有活跃 MQ 消费者"措辞有 audit 范围蔓延风险（C13）→ **已修复**：Purpose 增加范围限定句。
- 🟡 结算发起缺"首次发起即失败"Scenario（网关重试驱动断言无直接覆盖）→ **已修复**。
- 🟡 两态键计数与有界重投计数关系未显式统一 → **已修复**：声明为同一计数器。
- 💡 "已生效类响应"通用判据 → **已采纳**（副作用已生效即成功；超时/连接/校验失败除外）。
- 💡 "副作用失败时不确认"THEN 改可观测断言 → **已采纳**。

确认项：proposal 八项变更逐条有 Requirement 承接；barrier 以"编排层分支执行记录"抽象表述
不泄漏表名；C0-C13 spec 层全覆盖；第 1 版修复（容忍契约/TTL 过期/无 AMQP 错误码）无退化。

结论：🔴 清零，🟡 已修复，**specs 冻结**。

## 实施期 Checklist 勾验（apply · 6.3，2026-09-04）

对照 explore-brief.md Checklist C0~C12 逐项核验（门禁 + 探针证据）：

- [x] **C0 channel 修复**：三 consumer 全部 autoAck=false、业务成功后才 Ack（common/mq/consumerx + notify/delay/payment 迁移）。探针 1-4 验证手动确认与真实 broker 交互正常。
- [x] **C1 fall-through**：失败即 return，由骨架统一进入失败处理；notify/delay 回调单测（查单失败→不执行后续副作用）+ fall-through 回归测试通过。
- [x] **C2 监督循环**：Run/consumeOnce + 指数退避（cap 30s）；`context.TODO()` 全部消除；svc 经 `proc.AddWrapUpListener` 注册 cancel，wrap-up 触发优雅退出。探针 3（cancel→限时等待→超时强制关连接）通过。
- [x] **C3 幂等释放**：重投前 Release（idempotency.Release DEL），单测断言释放序列 + 探针 2 验证。
- [x] **C4 重试上限**：Redis 计数（settlement:retry:{queue}:{identity}），默认 3 次弃单 + error 级结构化日志（identity/cause/attempts）。探针 2：恰好 3 次尝试后弃单、队列清空。
- [x] **C5 毒消息**：解码失败一次即弃（Reject(false)）+ error 日志含 body 摘要；不引入 DLX（design 决策 3 记录取舍）。探针 1 通过。
- [x] **C6 发布失败兜底**：SettleOrder RPC（proto goctl 生成）+ webhook 快路径调用；logic 单测覆盖未支付拒绝/已支付补发/重复调用等效/发布失败 Internal。
- [x] **C7 TTL 语义**：TTL=1h 维持并在 design 决策 4 记录窗口风险（下游三步业务幂等收敛）。
- [x] **C8 业务幂等不回归**：库存"已扣减"/券"已使用"容忍语义保留（notify 回调单测）；订单已 Paid 跳过更新仍结算的既有语义保留。
- [x] **C9 DDL 红线**：proto 经 goctl 生成（settleorderlogic.go 为生成模板内填实现）；admin 服务未涉及。
- [x] **C10 测试策略**：mock 先行（全部 fake 注入，无外部依赖）；channel 死亡可恢复（探针 4 + 单测）与失败不双扣（fall-through/重试序列单测）两类场景齐备。test-unit 白名单新增 7 个包。
- [x] **C11 非目标**：未引入 DTM/Outbox/event.ID 去重；PaymentMQ 保持 nil。
- [x] **C12 PaymentDelayMQ 处置**：骨架统一加固（consumerx + Start(ctx) 模式），svc 保持禁用，无"带病休眠"残留。

### 实施期发现的额外缺陷（已修复）

1. **deliveries closed-channel 热循环（严重，探针 4 发现）**：consumeOnce 忽略 `<-deliveries` 的 ok 标志，通道关闭后零值 Delivery 被当毒消息处理 → 死 channel 上无限快速循环（毫秒级刷日志）。已修复：ok=false 即退出本轮交监督循环。该缺陷正是本 change 要消灭的 bug 类型，集成探针物有所值。
2. **goctl 重生成风险处置**：server 文件含手写 Seckill 元数据逻辑，goctl 生成新文件后已将定制代码迁移至新生成文件并删除旧文件；`createorderlogic.go` 重复脚手架与 `github.com/` pb 树已清理。
3. **订单服务 orderservice/orderservice.go（未跟踪文件）在 goctl 重生成时被覆盖**：原内容不可恢复；该目录为遗留客户端脚手架、无任何构建/部署引用（Dockerfile 构建根 order.go），已从产物中移除该文件，恢复为仅含已跟踪的 order_service.go。请 owner 知悉。

### 门禁终态

- make lint ✅ / make test-unit ✅（44 包 134 测试 0 失败）/ make build ✅ / make gatekeeper ✅（7/7）
- 集成探针 4/4 PASS（真实 RabbitMQ + Redis，证据见 review-log 与探针输出）

### 批次 design · 第 3 轮（2026-09-04）

审查对象：design.md 修订版（决策 1-7）；对照已冻结 proposal/specs + explore-brief 第二轮；
核对第 1 版 design 修复（graceful shutdown 时序/panic recovery/职责边界）无退化。

- 🟡 决策 2 遗漏 :44-48 订单不存在与 :64-69 payment status 分支的 Saga 语义声明（含补偿
  悬挂风险：rollback 对不存在同样 Aborted → 补偿链悬挂）→ **已修复**：显式声明两分支
  预期行为 + 悬挂风险靠 barrier 主保险 + dtm 补偿重试告警承接 + 测试须覆盖该边界。
- 🟡 决策 4 旧单态 key（值 "1"）迁移兼容未提及 → **已修复**：新 Lua 将 "1" 识别为
  Processing（保守正确：不能映射 Done，否则崩溃消息被误跳过），1h 自然过期。
- 💡 Processing 状态 Ack 跳过的取舍依据 → **已采纳**（避免阻塞队头；丢失窗口由 lease
  TTL 界定 + delay 低频可查）。
- 💡 dev dtm.yaml 补 Store 显式入迁移第 1 步 → **已采纳**（dev/生产两份同步）。

确认项：决策 1 gid 幂等依据成立、svc 封装满足 Rule 4；决策 3 bolt 风险论断与代码属实
（dtm.yaml 无 Store、compose 无持久卷）；决策 4 与既有 consumerx 兼容、与 specs 两态键
Requirement 语义一致；决策 5 退役清单与 proposal Impact 一致；部署顺序/回滚可行。

结论：🔴 清零，🟡 已修复，**design 冻结**。

### 批次 tasks · 第 3 轮（2026-09-04）

审查对象：tasks.md 修订版（6 组 22 任务）；对照已冻结 proposal/design/specs + explore-brief
C0-C13 + Makefile/代码现状核对。

- 🔴 3.1 与 4.1 对 notify publish/import 删除职责双向踢皮球（3.1 写"归入 4.1"、4.1 写
  "3.1 已含"）→ **已修复**：单向化——3.1 明确"不触及"，4.1 明确"本任务执行"。
- 🟡 2.2 缺"幂等标记删除产生可恢复级日志"承接 → **已修复**。
- 🟡 3.1 缺 :57-61（已 Paid 三分支核心修改点）位置引用 → **已修复**。
- 🟡 2.3 "按新 Store 接口适配"表述失真（Handler[T] 不直接依赖 Store，预计无需改代码）→
  **已修复**：改为"编译适配 + Config 新增字段装配"。
- 🟡 1.2 多库迁移脚本组织无说明 → **已修复**：明确 construct/depend/sql/dtm_barrier.sql +
  文件头注明三库分别执行。
- 🟡 3.x/5.x 验证缺具体命令 → **已修复**：补 go test 包路径。
- 💡 compose -f 全路径 / rabbitmqctl 示例 / make mock 合并注记 → **已采纳**。

结论：🔴 已修复，待第 4 轮复审。

### 批次 tasks · 第 4 轮（2026-09-04）

复审：第 3 轮 🔴/🟡 全部 ✅（职责单向化、位置引用、命令补全均到位）。

- 🟡 2.2 日志描述"成功时"指代歧义（Release 本身在失败路径调用）→ **已修复**：改为
  "重试未超限 Info + 重投 / 超限 error + 弃单"，对齐冻结 specs。
- 🟡 6.2 探针③"不可恢复失败"缺场景构造说明 → **已修复**：注明业务级错误（codes.Aborted）
  而非网络超时（超时应走重试）。
- 🟡 缺显式 go mod tidy 任务（design Migration 步骤 2 有要求）→ **已修复**：并入 1.3。
- 💡 "supervisor"术语 → **已采纳**："监督循环自动重连恢复"。
- 💡 rabbitmqctl 连接定位 → **已采纳**：补 list_connections。
- 💡 2.3 编译失败排查提示 → **已采纳**。

确认项：C0-C13 全承接；11 Requirements/33 Scenarios 全覆盖；与冻结 design 决策 1-7 无矛盾；
Makefile 目标真实；粒度 ≤2h；scope 无蔓延（audit 不触及、PaymentMQ 保持 nil）。

结论：🔴 清零，🟡 已修复，**tasks 冻结**。四个制品（proposal/specs/design/tasks）修订周期
全部审核完毕，规划重新冻结，可进入 apply。
