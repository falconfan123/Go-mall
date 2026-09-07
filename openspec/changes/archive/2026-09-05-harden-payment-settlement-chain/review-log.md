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

### 实施期增量 · SettleOrder 退役登记（2026-09-04，apply 启动时）

apply 启动核查发现：change 曾被并行会话在 0/21 任务时归档（17:31），且并行会话已完整实施
第 1 版计划（consumerx 迁移/SettleOrder RPC/探针/门禁全绿，review-log 实施期三段记录）。
经用户确认：单 change 实施现有 Saga 计划，并登记实施期发现——SettleOrder 退役
（design 决策 1 附注 / proposal Impact API / tasks 5.4）。聚焦复审 🔴=0 可合入。
仓库操作：还原点 f122f76（git add openspec/ 快照归档态）→ mv 回 changes/ → 删 premature
主 spec（openspec/specs/async-settlement-reliability，真归档时重新同步）。

### 实施记录 · 任务 1.1-1.4（2026-09-04）

- 1.1 ✓ dtm Store=postgres（业务栈实为 PG，compose 的 mysql 是 gorse 用——修正 design/tasks
  的 MySQL 表述）、镜像 pin yedf/dtm:1.19.0、compose 修挂载相对路径（configs/ 下需 ../，
  原配置为带病路径）+ 挂 depend_mall 外部网络、TimeZoneOffset: 8（时区坑实测：不配则 cron
  永远捞不到任务）。验证：容器启动/ping/重启 pong ✓。
- 1.2 ✓ construct/depend/sql/dtm_barrier.sql（两段：dtm 库创建 + mall 库 barrier schema，
  postgres 方言取自 dtm v1.19.0 sqls/dtmcli.barrier.postgres.sql）。验证：三库对象就绪 ✓。
- 1.3 ✓ 依赖修正：github.com/dtm-labs/dtmclient 不存在——实为 github.com/dtm-labs/dtm
  v1.19.0（client/dtmgrpc）。四服务 go get + tidy + build 全绿 ✓。
- 1.4 ✓ spike（临时容器，不进主干）五场景全过：正常提交/重试至成功（指数退避）/确定性失败
  逆序补偿/重复 gid 状态依赖语义/dtm 重启续跑。结论固化进 design 决策 1 附注（分支契约：
  业务失败=Aborted；不可用=其他错误码；已 Paid 重放=成功）。

### 实施记录 · 任务 1.5 / 2.1-2.3（2026-09-05）

- 1.5 ✓ common ConsumerConfig 增加 SuccessTtlSeconds（默认 86400）+ 测试三段断言更新；
  payment 新增 DtmConfig（Server/BusiHost/三分支端口）+ payment.yaml Dtm 段 + 配置加载
  单测。偏差记录：任务原文写"四服务新增 DTM target"，实际只有 payment 发起 saga 需要
  （分支服务不外呼 dtm，barrier 走上下文头+本地库），inventory/coupons 未加 Dtm 配置。
- 2.1 ✓ idempotency 新增 Claim（三态 Lua：NEW/PROCESSING/SUCCESS，历史单态 "1" 与未知
  值保守映射 Processing）+ MarkDone（SUCCESS PX）；CheckAndSet/Release 保留（audit，C13）。
  单测 5 组（生命周期/旧值/租约过期重入/Release 重占）真实 Redis 全绿。
- 2.2 ✓ consumerx Store 状态化 + Process 两态流程（Processing/Done 跳过、MarkDone 失败按
  业务失败重投）+ Config.SuccessTTL（默认 24h）。测试重写 12 项含日志级别区分断言
  （重投=info / 弃单=error）。既有约定保留：BuildKey 输出含 idempotency: 前缀（audit
  双前缀键兼容，C13）。
- 2.3 ✓ notify/delay/payment 三个适配层 SuccessTTL 装配 + seqStore fake 适配新接口；
  order/payment 全部构建+测试绿。

### 实施记录 · 任务 3.1-3.4（2026-09-05）

- 3.1 ✓ order 分支：三分支判定（PendingPayment 流转 / **已 Paid 重放=成功** / 冲突=Aborted）
  + barrier 接线（dtmgrpc.BarrierFromGrpc + BranchBarrier.CallWithDB，业务经
  sqlx.NewSessionFromTx 注入 barrier 自管事务）；svc 新增 DtmDB（lib/pq，共用 DSN）。
  实施期修复既有缺陷：原 rollback 逻辑 DB 错误后缺 return → orderRes 空指针 panic。
  测试 9 例（5 单元 + 4 barrier 真库集成：首次结算/重放短路/冲突 Aborted/空补偿 NoOp，
  含 barrier 行清理防跨运行污染）。
- 3.2 ✓ inventory：模型抽 BatchDecrease/ReturnInventoryAtomWithSession（嵌套事务不可行，
  errCantNestTx）；decrease/return logic barrier 分支（业务规则失败=Aborted，基础设施=退避
  重试；直连路径语义原样保留）。coupons：**实施期发现** —— Saga 补偿回调与 action 共用
  同类型 payload（UseCouponReq），既有 ReleaseCoupon(ReleaseCouponReq) 签名不兼容；
  新增薄适配 RPC `UseCouponRollback(UseCouponReq)`（proto + pb 重生成 [package pb 映射修正]
  + server 注册 + logic barrier 实现：Used→回滚可用+撤销流水 / Available·Locked·不存在→
  容忍成功）；UseCoupon 抽 useCouponTx 共用体 + barrier 分支（"已使用"容忍提升为分支语义）；
  CouponUsageModel 新增 DeleteByOrderId。
- 3.3 ✓ 三补偿分支核验：order rollback（容忍语义重写：不存在/非 Paid→成功，补偿必须
  始终成功；补偿执行 error 级日志 per specs 可观测契约）、inventory Return、coupons
  Rollback 全部 barrier 化并单测/集成验证。
- 3.4 ✓ make mock + 四服务 build 全绿。

### 实施记录 · 任务 4.1-4.3 / 5.1-5.4（2026-09-05）

- 任务顺序微调：5.4（SettleOrder 拆除）的 order 侧提前至 4.2 执行（settleorderlogic 依赖
  notify 包，阻塞构建），5.2（payment 接线）随后插入——构建全程保持绿（Rule 1）。
- 4.1 ✓ publish 已随 3.1 重写移除（无 notify import）。
- 4.2 ✓ notify 目录（consumer/init/product）+ svc 装配（OrderNotifyMQ 字段/NotifyPublisher
  接口/initOrderNotifyMQ/Start）全删。
- 4.3 ✓ 全仓 rg：SettleOrder/OrderNotifyMQ/order-notify 零业务残留；四服务 build 绿。
- 5.1 ✓ payment svc 新增 SettlementSaga（gid=settle:{orderId}；三分支+补偿 URL 由
  DtmConfig(BusiHost+端口) 拼装；Initiate/Ensure；Ensure 语义：in-flight→dtm 静默接受、
  终态 succeed→映射成功、终态 failed→error+告警（人工，已接受姿态））。
  抽 SettlementSagaAPI 接口（Rule 6 可 mock）。
- 5.2 ✓ payment.go 接线：本地 PAID 后——订单未 Paid → Initiate；订单已 Paid（快路径）
  → Ensure。结算输入快照经既有 GetOrder RPC（items/金额/coupon_id）；
  **实施期新增**：order pb Order 增补字段 coupon_id=18（additive，convert 填充）——
  saga 分支 payload 需要券 ID 而原 Order 消息未暴露。
- 5.3 ✓ make mock；5.4 ✓ SettleOrder 全链拆除（proto/pb/server/logic/test）。
  payment 测试重写 6 例（快路径 Ensure 必达+payload 映射/Ensure 失败传播/正常路径
  Initiate/Initiate 失败传播/两个查询失败传播）全绿。

### 实施记录 · 任务 6.1 门禁（2026-09-05）

make lint ✓ / make test-unit ✓（43 包 137+ 测试 0 失败；白名单移除已删除的 notify 包）
/ make build ✓ / make gatekeeper ✓（7/7）。

### 实施记录 · 任务 6.2 本地集成探针（2026-09-05）

环境：真实 dtm 容器（yedf/dtm:1.19.0，postgres Store）+ 新二进制 order/inventory/coupons
（10004/10007/10009）+ mall/dtm 双库 + 真实 rabbitmq。探针代码：services/payment/e2e_probe_test.go
（build tag e2eprobe，不进白名单/CI）。

- 探针① ✓ 正常结算三分支全成功：订单 Paid/Paid、库存 990001 sold=1 / 990002 sold=2、
  券 Used + coupon_usage 流水 1 条、saga succeed。
- 探针②③ ✓ 券分支确定性失败（用户券不存在 → coupons 业务失败 → codes.Aborted）→
  saga failed → **逆序补偿实测**：订单回滚 PendingPayment(2)、库存回补 sold=0、无使用流水。
- 探针④ ✓ 终态 saga 重复提交 → `Aborted: current status 'succeed', cannot sumbmit. FAILURE`
  （Ensure 映射成功/failed→告警的原始依据，与单测一致）。
- 探针⑤ ⚠️ 部分完成：delay 链路四项（毒消息一次即弃/失败 3 次弃单/优雅退出/channel 恢复）
  框架级已由 consumerx 单测 + 并行会话真实 broker 探针（第一周期 C0-C5 勾验记录）背书，
  本期未改动该代码路径；现场重验被 dev 环境陈旧消费者干扰（audit tmp/main 陈旧进程 +
  start-unified 看护与手动探针实例抢端口，毒消息被无痕消费）。**待办**：start-unified.sh
  干净重启后人工复核（rabbitmqadmin publish 到 order-delay-dlx-queue + 观察 error 日志）。
  故 6.2 未勾选，留待人工复核后勾选。
- 环境恢复：探针实例已停；spike 临时容器已清理；start-unified.sh restart 已触发
  （后台推进中，卡在 audit 就绪检查——audit 为未改动服务，与本 change 无关）。

### 任务 6.3 · explore-brief 第二轮 Checklist 逐项勾验（2026-09-05）

- [x] C0 channel 修复（delay）：consumerx autoAck=false + 成功后 Ack，本期未动该核心。
- [x] C1 fall-through：骨架统一失败处理 + 3.1 修复 rollback nil-deref（同款复制 bug）。
- [x] C2 监督循环：第 1 版已落地保持（Run/consumeOnce/指数退避/优雅停机）。
- [x] C3 幂等 → 两态键 lease 版（D9）：idempotency.Claim/MarkDone + consumerx 流程
      （Processing/Done 跳过、失败 Release+INCR）+ 5 组单测。
- [x] C4 重试上限：Redis INCR + 超限弃单告警保持，测试断言（含日志级别可区分）。
- [x] C5 毒消息：一次即弃 + error 日志保持（单测；现场复核见 6.2 探针⑤ 待办）。
- [x] C6 Saga 化：三分支/补偿/逆序补偿/EnsureSaga —— 3.x/5.x + e2e 探针①②③④。
- [x] C6a dtm 依赖（github.com/dtm-labs/dtm v1.19.0）/barrier 三库（mall：order/inventory/
      coupons 共用 barrier schema；dtm 库：server 存储表）/确定性 gid/barrier 接线/
      已 Paid 不误判失败（barrier + 三分支判定 + 容忍，单测+集成+e2e 三层验证）。
- [x] C6b notify 全链删除：目录/svc 装配/publish/import 零残留（rg + build）。
- [x] C7 两态 TTL：IdempotencyTtlSeconds=lease（1h）+ SuccessTtlSeconds=24h 配置化。
- [x] C8 业务幂等不回归：订单状态机/库存锁表容忍/券容忍保留；分支幂等双保险
      （barrier + 容忍），e2e 探针③ 补偿全量回滚实测。
- [x] C9 DDL 红线：pb 经 protoc 生成（order/coupons，package 映射修正）；SettlementSaga
      封装于 svc 层，main 包不构造 pb。
- [x] C10 测试：mock 先行（SettlementSagaAPI/OrdersModel/PaymentModel 全 fake）；四类关键
      场景齐备（重试不误补偿/EnsureSaga 幂等/channel 恢复/失败不双扣）。
- [x] C11 不做清单：Outbox/DLX/event.ID/延迟对账/PaymentDelayMQ 复活均未做。
- [x] C12 PaymentDelayMQ 保持 nil（骨架随 consumerx 同步，决策 6）。
- [x] C13 audit 不在 scope：CheckAndSet/Release 原语保留，audit 语义未破坏。

结论：C0-C13 全部满足；21+1 任务中 22 项完成、6.2 留探针⑤人工复核待办。

### 任务 6.2 · 探针⑤ 人工复核通过（2026-09-05，用户确认）

环境：`./scripts/start-unified.sh` 干净重启后（消除此前陈旧消费者干扰），delay 链路四项
复核结果均符合 specs 预期：

- 毒消息投递 → **一次即弃**（Reject(false)，不再重投）+ error 级结构化告警日志（含消息体摘要）；
- 业务失败重投 3 次达上限 → **弃单**（Reject(false)）+ 需人工介入级结构化告警日志（含业务
  标识、失败原因、累计次数），日志级别与可恢复事件可区分。

覆盖 specs 场景："非法 JSON 消息被弃""超过重试上限后弃单""弃单产生需人工介入级别的日志"。
至此探针①-⑤全部通过，**任务 6.2 完成**。

- [x] 6.2 本地集成验证（探针①②③④由自动化 e2e 覆盖见上文记录，探针⑤人工复核通过）

（勘误）任务 1.5 勾选补记：工作已于实施期完成并验证（common ConsumerConfig.SuccessTtlSeconds
+ payment DtmConfig + yaml + conf 加载单测全绿，见"实施记录 · 任务 1.5 / 2.1-2.3"），勾选
批次时遗漏。补勾后 22/22 全部完成。
