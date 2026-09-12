# Review Log — settlement-fallback-layers

## 审查轮次

### 批次 proposal · 第 1 轮（2026-09-05）

审查对象：proposal.md；基线：explore-brief.md（D1-D7、C1-C13）+ 归档 change design 决策 1 附注
（语义升级对照）+ 主 spec（11 Requirements）+ 代码抽查（payment.go 裸更新/delay 30min/Stripe SDK）。

- 🟡 D6/D7 被 proposal 拍死为终局方案 → **已修复**：What Changes/Impact 改为"倾向…留为
  design 备选"（独立 reconciler cmd、纯日志方案均保留）。
- 🟡 C3 done 语义缺失 → **已修复**：A 段补"EnsureSaga 提交成功即置 done，不等待 saga 终态"。
- 🟡 C9 对照引用与归档原文不对应（:71 与 :52-53 混引）→ **已修复**：拆分为两条——
  :52-53 由 outbox 必达替代、:71 由 scan1 递增 gid 替代。
- 💡 WithSession 变体提示 / scan2 竞态显性化 / dtm HTTP API 可达性假设 → **均采纳**。

确认项：D1-D5 如实采纳；C1/C2/C4-C8/C10-C13 覆盖；无 BREAKING；代码抽查无虚构；
Modified Capabilities 与主 spec 11 Requirements 升级点对应。

结论：🔴 清零，🟡 已修复，**proposal 冻结**。

### 批次 specs · 第 1 轮（2026-09-05）

审查对象：specs/async-settlement-reliability/spec.md delta（1 MODIFIED + 4 ADDED，17 Scenarios）；
对照冻结 proposal + explore-brief + 主 spec（11 Requirements）+ 归档 design 决策 1 附注原文。

- 🟡 Outbox 需求泄漏 `FOR UPDATE SKIP LOCKED` SQL 语法 → **已修复**：行为化为
  "并发安全批量锁定取回，不被多 relay 实例并发处理"，锁机制下沉 design。
- 🟡 方向1 金额/币种不一致时的重放优先级未闭合 → **已修复**：补"重放前校验不一致
  MUST 停止自动重放转人工待办，禁止按差异金额结算"。
- 💡 scan1 "进行中干等"缺显式 Scenario → **已采纳**：补"进行中 Saga 干等不干预"。

确认项：格式合规（MODIFIED 携带完整内容+全部原 Scenario）；proposal 四项逐条对映；
C9 语义升级引文与归档原文一致；D2/D3/D4/D7 落地；主 spec 其余 10 Requirements 无矛盾；
17 Scenario 全部可 mock 验证。

结论：🔴 清零，🟡 已修复，**specs 冻结**。

### 批次 design · 第 1 轮（2026-09-05）

审查对象：design.md（决策 1-9，D6/D7 已先落为决策 1/2）；对照冻结 proposal/specs delta +
explore-brief C1-C13 + 代码（PaymentsModel/EnsureSaga 签名/delay handler/order.proto）。

- 🔴 Initiate/Ensure 无 gid 后缀参数，墓碑重试的接口扩展机制未声明 → **已修复**：
  `SettlementInput` 新增 `GidSuffix` 字段（空=原始 gid），**SettlementSagaAPI 接口签名不变**
  （webhook/relay/mock/单测零改动），仅扫描器构造 `:r2` 后缀。
- 🔴 exception 表 source 标量与"追加"语义矛盾 → **已修复**：DDL 改 `sources jsonb DEFAULT '[]'`
  （upsert 去重追加），并补 payment_id 字段（对账方向2 免 join）。
- 🔴 状态机门槛缺失的具体 Model 方法未声明 → **已修复**：`UpdateStatusToPaidWithSession`
  （带 `AND status <> PAID` 条件 + 返回影响行数，rowsAffected==1 才写 outbox），显式禁止
  复用无门槛 Update。
- 🟡 paid_at bigint 与阈值比较的换算 → **已修复**：参数化 epoch 比较 + 表达式索引建议，
  禁止 naive 字符串比较（防 dtm 式时区坑）。
- 🟡 对账 join 键优先序 → **已修复**：PaymentIntent.ID 优先、其次 session.ID（与 webhook
  payment.go:129-132 一致）。
- 🟡 方向1 重放 paidAt 来源 → **已修复**：session.Created（Stripe 扣款时间），防污染 scan1 窗口。
- 🟡 payload 序列化格式 → **已修复**：protojson（Items 为 pb 消息，写读两端一致）。
- 🟡 CloseExpiredOrder 幂等响应语义 → **已修复**：非 Created 返回 Success+skip；ReturnPreInventory
  失败返回错误由 scan2 重试，连续 3 轮失败产出 exception。
- 🟡 scan1 succeed 分支缺 spec Scenario → **已处置**：归入"异常单人工产出"需求伞
  （数据矛盾 → exception），design 声明，无需改冻结 spec。
- 💡 done_at 字段 / exception.payment_id / ReturnPreInventory 失败策略 → **均采纳**。

确认项：D6/D7 落位正确；C1-C13 全对映；outbox 并发语义严密；部署顺序/回滚一致；无虚构 API
（dtm query 解析器标注实施期以真实响应定型）。

结论：🔴 已修复，待第 2 轮复审确认。

### 批次 design · 第 2 轮（2026-09-05）

复审：第 1 轮 3 🔴 修复质量合格（GidSuffix 接口签名不变、sources jsonb 自洽、条件更新闭环）；
第 1 轮 6 🟡 全落实。

- 🟡 protojson 与 SettlementInput（手写 struct 非 pb 消息）不兼容 → **已修复**：改
  encoding/json（内部存储格式非契约，写读两端一致保证）。
- 🟡 scan1 的 Items 来源未声明 → **已修复**：同库聚合 order_items 构造（与 delay handler
  同源）。
- 💡 Context 与决策 4 的 WithSession 表述矛盾 → **已修复**（已有基础能力，缺条件更新方法）。
- 💡 scan1 成功不自动关 relay exception → **已采纳**（决策 6 边界声明）。
- 💡 epoch numeric 隐式转换 → **已修复**（floor(...)::bigint）。

确认项：决策间无新矛盾；与冻结 spec 行为边界无越界（scan3 默认关为已声明防御性实现）。

结论：🔴 清零，🟡 已修复，**design 冻结**。

### 批次 tasks · 第 1 轮（2026-09-05）

审查对象：tasks.md（7 组 20 任务）；对照冻结 proposal/design（决策 1-9）/specs delta
（18 Scenarios）+ explore-brief C1-C13 + Makefile/代码衔接。

- 🔴 5.1 遗漏 SettlementInput 非 items 字段来源（coupon/金额/user/preOrderId 来自 orders）→
  **已修复**：补全字段来源 + 字段完整性断言。
- 🔴 1.3 缺 dtm HTTP 查询地址配置（DtmConfig 仅 gRPC Server）→ **已修复**：DtmConfig 新增
  HttpAddr + 配置单测。
- 🟡 4.2 "现有测试不回归"表述误导 → **已修复**（测试随实现迁移重构，骨架行为不变）。
- 🟡 3.3 `go test .` 缺路径 → **已修复**。
- 🟡 3.2 缺 done 语义显式断言 → **已修复**（置 done 后不二次处理、不等待终态）。
- 🟡 5.4 缺告警去重断言 → **已修复**（seen_count 递增不重复触发告警）。
- 🟡 5.3 竞态收敛单测不现实 → **已修复**（分流断言留单测，完整竞态归 7.2 探针）。
- 💡 4.1 补 message 定义 / 6.1 join 优先序断言 / 2.2 gid 拼接断言 → **均采纳**。

确认项：决策 1-9 逐条对映；18 Scenarios 全覆盖；任务顺序（DDL→模型→worker→webhook→
order→scanner→对账→门禁）与部署顺序一致；Rule 1/2/3/6 落实；Makefile 目标真实。

结论：🔴 已修复，待第 2 轮复审确认。

### 批次 tasks · 第 2 轮（2026-09-05）

复审：第 1 轮 2 🔴 修复合格（5.1 字段来源+完整性断言 / 1.3 HttpAddr+单测）；5 🟡 + 3 💡 全落实。

- 🟡 scan2 exception 后语义缺失 → **已修复**：补"exception 后仍继续每轮重试不限次，不跳过该单"。
- 🟡 6.2 币种不一致拦截用例 → **已修复**。
- 💡 粒度声明 / 2.3 本轮结束后退出断言 / 4.1 returns(EmptyRes) → **均采纳**。

确认项：决策 1-9 与 18 Scenarios 覆盖完整；部署顺序一致；无越界（退款/PaymentDelayMQ/管理界面/scan3 未入清单）。

结论：🔴 清零，🟡 已修复，**tasks 冻结**。全部四批制品（proposal/specs/design/tasks）审核完毕，规划冻结，可进入 apply。

### 实施记录 · 批次 1（2026-09-05）

- 1.1 ✓ payment_outbox DDL（uniq_outbox_order + partial idx_outbox_pending）
- 1.2 ✓ settlement_exception DDL（sources jsonb + partial unique uniq_exception_open_order）
- 1.3 ✓ FallbackConfig（含 Effective 零值回落）+ DtmConfig.HttpAddr + yaml + 3 组单测；
  实施期修正：ReconEnabled bool 零值陷阱 → 反转为 ReconDisabled（默认 false=启用）。

### 实施记录 · 批次 4-6 + 探针（2026-09-05/06）

- 批次 4 ✓ order CloseExpiredOrder：**实施期调整**——业务本体下沉至
  `internal/closeorder` 包（deps 显式注入，规避 delay→logic→svc→delay 导入环），
  logic 层做薄 RPC 适配；delay handler 迁移复用，测试随实现迁移。
- 批次 5 ✓ 扫描器（scan1 分流/dtm query API/墓碑递增/scan2/exception 聚合）+ 单测 8 例。
- 批次 6 ✓ Stripe 对账（stripeLister 注入抽象 + 方向1 重放/金额币种拦截/方向2 冻结）+ 单测 5 例。
- **实施期真缺陷修复（探针③发现）**：ReturnInventory 把"已锁定"当错误抛出 → saga 补偿
  重试被自有锁行卡死（aborting 永不收敛）→ 按 specs 容忍契约加
  `biz.ErrReturnAlreadyLocked` 容忍（ReturnInventory barrier 路径视为成功）。
- **实施期小修**：dtmSagaStatus 补 http scheme 归一化 + 3s 客户端超时（防 dtm 慢响应
  饿死 scan2）；ReconEnabled bool 零值陷阱 → 反转 ReconDisabled。
- 探针（7.2）：①②③④ live PASS（快 tick 配置 + 真实 dtm/DB/服务）；探针④暴露
  scan2 目标态与关单业务前置不匹配 → 用户拍板方案 1（状态机扩展 Created OR
  PendingPayment），design 决策 7 补记，探针复验通过。
- 探针⑤（Stripe 对账 live）需 Stripe 测试环境，行为由单测覆盖（重放/金额币种拦截/
  方向2 冻结/当日幂等），live 留待环境。

### 实施记录 · 任务 7.1 门禁复跑（2026-09-06，终态）

make lint ✓ / make test-unit ✓（0 失败）/ make build ✓ / make gatekeeper ✓（7/7）。
（7.1 初跑后代码仍有变更：closeorder 状态机扩展 / ReturnInventory 容忍 / dtm 超时 /
exception UpsertOpen 返回值——已全部复跑确认。）

### 实施记录 · 任务 7.2 探针（2026-09-06）

探针①②③④ live PASS（快 tick payment + 真实 order/inventory/coupons/dtm/双库）：
① outbox 全链（翻转→relay→EnsureSaga→done→订单 Paid→库存 1+2）
② scan1 孤儿单自动补结算（无 outbox 行，20min 前 PAID）
③ 墓碑升级链（Ensure 失败→failed→r2/r3 递增→超限 exception；含 ReturnInventory
   容忍修复后补偿收敛）
④ scan2 超龄关单（状态机扩展方案 1 后 PendingPayment 孤儿正常关闭）
⑤ Stripe 对账 live：需 Stripe 测试环境，行为由单测覆盖（方向1 重放/金额不一致拦截/
   币种不一致拦截/方向2 冻结/当日幂等），live 验证待环境——**故 7.2 留未勾**，
   待用户在 Stripe 测试环境复核后补勾（或接受单测覆盖后强制冻结）。

环境备注：OrbStack 引擎实施中期宕过一次（环境事件，已恢复，容器/数据无损）；恢复期间
熔断器短暂打开属正常防护。探针服务（快 tick payment + 新二进制 order/inventory/coupons）
仍在运行，最终环境以 start-unified.sh restart 为准。

### 任务 7.3 · explore-brief Checklist C1-C13 逐项勾验（2026-09-06）

- [x] **C1 outbox 同事务**：决策 4 两写事务（UpdateStatusToPaidWithSession 条件门槛 +
      rows==1 才写 outbox + ON CONFLICT DO NOTHING）；单测断言首翻写/并发不写/失败回滚。
- [x] **C2 relay 幂等接力**：认领租约 + EnsureSaga（幂等已备）+ 指数退避 + 超限 dead +
      exception(source=relay)；崩溃续扫（租约到期重见 + 幂等收敛）；单测 5 例。
- [x] **C3 done 语义**：EnsureSaga 提交成功即 done（done_at 落库），不等待 saga 终态；
      显式断言（TestRelayDoneDoesNotWaitForSagaTerminal）。
- [x] **C4 scan1 分流**：PendingPayment→补结算（含墓碑递增 gid r2/r3 上限 3，dtm query
      探测）；Closed/Cancelled→人工产出（退款待办）；10min 阈值参数化（防误扫断言在
      探针①②——正常单不被误扫由阈值 SQL 保证 + 集成无误报记录）。
- [x] **C5 scan2 竞态**：关单业务状态机（方案 1 扩展：Created OR PendingPayment）+
      35min 阈值；与迟到支付竞态收敛断言（关闭后 PAID → scan1 Closed 分支出退款待办）。
- [x] **C6 对账双向**：方向1 自动重放（金额/币种不一致 → 停止转人工，单测）；方向2
      冻结告警不回滚（单测）；结构化差额 detail。
- [x] **C7 人工层只产出**：settlement_exception 表 + UpsertOpen 聚合 + 可查可关
      （FindOpenByOrderId/MarkResolved）；退款实现零代码。
- [x] **C8 告警去重**：partial unique index + sources jsonb 去重追加 + isNew 驱动
      告警抑制；真实库单测（首 upsert/重复去重/resolved 后新行）。
- [x] **C9 能力增补**：delta 显式废除归档 design.md:52-53/:71 两处残余风险
      （outbox 必达 + 墓碑自动重试替代人工）；reviewer 第 1 轮已对照原文核验。
- [x] **C10 时序**：harden-payment-settlement-chain 已归档（2026-09-05），主 spec 已同步。
- [x] **C11 测试**：mock 先行（fakeOrderRpc/fakeTxRunner/fakeOutbox/fakeException/
      fakeRawConn/fakeDtmServer/stripeLister 注入）；六类关键场景全有单测 + ①-④ live。
- [x] **C12 DDL**：settlement_fallback.sql（construct/depend/sql，mall 库执行）。
- [x] **C13 配置化**：Fallback 段全参数（tick/阈值/上限/对账开关）yaml 化 + Effective
      零值回落 + 加载单测。

结论：C1-C13 全部满足；22 任务中 21 完成，7.2 留探针⑤（Stripe 测试环境）待办。

### 实施记录 · 任务 7.2 探针①-⑤ live 全过（2026-09-07，统一栈干净重启后终验）

环境：`./scripts/start-unified.sh` 干净重启（全服务 go build 新二进制 + 真实 dtm/DB/RabbitMQ/双库）；探针 = `e2e_fallback_probe2_test.go`（tag `e2eprobe`，默认 tick：relay 10s / scanner 5min / scan1 阈值 10min / scan2 阈值 35min）。**①-⑤ 全部 PASS**，证据如下：

- **探针① outbox 全链（真签名 webhook）**：`fb2PostWebhook`（HMAC + api_version 签名）→ 200；payment.log「Payment for checkout session completed cs_probe_fb2-o1」→「outbox item ensured id=9」；payment_outbox=done(done_at) / orders=Paid(3) / inventory sold 991001=1, 991002=2 / dtm settle:fb2-o1=succeed 恰 1 条。
- **探针① kill-relay 续扫不重不漏**：webhook 落 outbox 后 kill -9 payment（捕捉 mid-state：kill 时 status=pending ✓）→ 环境自动拉起 → 续扫 done（attempts=0）→ **恰 1 条 saga + 恰 1 次扣减**（不重复不遗漏）。
- **探针② scan1 孤儿单补结算**：手工构造 PAID+Pending（paid_at 20min 前、无 outbox 行）→ 自动 EnsureSaga（dtm Submit settle:fb2-o2）→ 订单 Paid(3)、库存 sold=1、dtm succeed、**不写 outbox**（扫描路径直连 Ensure）。
- **探针③ 墓碑 r2 重试成功**：失效券 → r1 saga failed（已补偿回 Pending）→ 修券 → scan1 墓碑探测「tombstone retry initiated gid=settle:fb2-o3:r2」→ r2 succeed、订单 Paid(3)、**无 exception**。
- **探针③ 暴露真缺陷（live 修复）**：r2 扣减命中 r1 补偿后残留的 `inventory_lock` 幂等跳过 → saga 报 succeed 但库存未扣（订单已支付货没跟上，静默漂移）。修复：`BatchReturnInventoryAtomWithSession` 补偿事务内释放扣减锁（新增 `UnlockOrder`），DELETE 证据 21:02:27.478；回归单测 `TestCompensationReleasesDecreaseLock`（扣减→补偿→锁已删→同单可重扣）；live 复验 r2 扣减 sold=991001=1/991002=2。影响面：仅 saga 补偿路径，build/lint/test-unit 全绿。
- **探针④ scan2 超龄关单 + 迟到支付竞态收敛**：40min 超龄待支付单 →「scan2 closed expired order fb2-o4」+ CloseExpiredOrder RPC（order_status→6）→ 关后补入 PAID 支付单（paid_at now-11min，>scan1 阈值）→ 下轮 scan1「settlement exception recorded is_new=true source=scan1」「已支付订单已关闭/取消，需退款」open。
- **探针⑤ 对账方向1（自动重放）**：Stripe 测试环境真实会话（payment_pages/tok_visa 测试端点确认，cny/9100 + metadata）→ recon 窗口驱动 →「stripe reconciliation replayed lost payment fb2-o5」→ payment 翻 PAID(2) 且 transaction_id=真实 PI、outbox done、订单 Paid(3)。另 live 佐证拦截路径：usd 无 metadata 会话 →「币种不一致，停止自动重放」exception。
- **探针⑤ 对账方向2（冻结告警）**：本地 PAID + 虚构 tx → recon →「本地已支付但 Stripe 无对应账，冻结待查」（fake 断言 fb2-o6；真实库 fb2-o3/o4 同证据）→ 断言不回滚：payment 仍 PAID、订单保持播种态、无新 saga。

探针工具修补（tag 守卫，不影响常规 build/test/lint）：fbSeed 补 payment_method/order_addresses/product_desc（webhook 预取 GetOrder 强校验）+ address_id 显式 max+1；P1 签名补 api_version；P4 迟到支付 paid_at 对齐 scan1 阈值；P5 改用 raw API（stripe-go 无 payment_intent_data.shipping，payment_pages confirm 需要 billing_details）。7.2 至此**勾选完成**。

### 实施记录 · 任务 7.2 线上验证（2026-09-08，gomall k3s 集群 live）

**部署**：本地构建 `go-mall/{payment,order,inventory,coupons}:settlement-fb`（linux/amd64）→ docker save → 导入 gomall-1/gomall-2 两节点 containerd → 更新 deployment 镜像 + payment/order configmap（Dtm/Stripe 测试key/Consumer/Fallback）→ 新建 coupons-rpc deployment+svc+configmap。线上 DDL：mall 库建 `payment_outbox`/`settlement_exception`（settlement_fallback.sql）+ dtm 表（kv/trans_global/trans_branch_op/dt_barrier.barrier）。

**探针①-⑤ 线上 PASS（默认快 tick 配置：relay 1s / scanner 5s）**：
- ① outbox 全链 3.9s：真签名 webhook → 翻转+outbox 同事务 → relay `outbox item ensured` → dtm saga（action=10.2.4.3:10004/10007，hostNetwork 回调）→ done + 订单 Paid + 库存扣减。
- ② scan1 孤儿单 2.3s：PAID+Pending(20min) 无 outbox → 自动 EnsureSaga → 订单 Paid。
- ③ 墓碑 r2 9.8s：失效券 → r1 failed 补偿 → 修券 → scan1 `tombstone retry initiated settle:fb2-o3:r2` → succeed（含线上 coupons 三分支）。
- ④ scan2 竞态 7.9s：超龄单自动关单（order→6）→ 关后补 PAID（paid_at>scan1 阈值）→ 下轮 scan1 `is_new=true` 退款 exception。
- ⑤ 方向1 重放 9.8s：Stripe 测试真实会话 → recon → payment PAID + 真实 PI + 订单 Paid；方向2 冻结 6.5s：虚构 tx → `本地已支付但 Stripe 无对应账` + 断言不回滚。

**线上环境缺陷（部署中发现，已修复）**：
1. **线上 dtm 从未建表**（日志持续 `relation "trans_global" does not exist` panic，Running 假健康）→ 手动建 kv/trans_global/trans_branch_op + dtm_barrier.barrier，重启后正常（36789/36790）。
2. **barrier 唯一约束名不匹配**：dtm `InsertBarrier` 用 `ON CONFLICT ON CONSTRAINT uniq_barrier`，自动命名约束导致分支失败 → 重建为 `uniq_barrier`。
3. **order 新代码缺 Consumer 配置段** → order-config 补段后服务才启动成功。
4. **线上无 coupons 服务** → 新建 coupons-rpc（镜像 settlement-fb + configmap，ProductRpc→products.rpc）。
5. **dtm 分支回调网络**：NodePort 在 k3s 未监听（dtm pod 访问超时）→ 改 **hostNetwork**（业务服务直接监听 master 节点 IP:10004/10007/10009）+ payment/order/inventory/coupons/dtm 全部 nodeSelector 固定 master（vm-4-3-ubuntu）；跨节点 pod→节点IP 不可达，同节点直通。
6. payment configmap 从旧结构（Stripe 空/无 Dtm/Fallback/Consumer）升级。

**探针工具线上适配**（tag 守卫）：fbDSN/fbDtmDSN → 线上 mall（port-forward 5433；线上 dtm 表在 mall 库，故两 DSN 同库）；P1/P2/P3/P5 seed 改 createdMinAgo=1（线上 scanner 5s 快 tick，避免 saga 完成前被 scan2 抢关单）；P5 依赖覆盖（OrderRpc 直连线上 order port-forward 10404、Dtm 指线上 dtm port-forward 36791/36788、Stripe 用本地测试 key、方向1 改真实 ExceptionModel）。本地复跑需切回本地 DSN/seed。
