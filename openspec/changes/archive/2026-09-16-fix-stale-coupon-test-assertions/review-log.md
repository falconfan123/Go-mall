# Review Log · fix-stale-coupon-test-assertions

## 审查轮次 1 — 2026-09-15（原"整体反思审核"节，内容未改）

**审核方式**：@openspec-reviewer 对照 explore-brief 与 F1 溯源（归档 review-log F1 节）逐条核，含服务代码 usecouponlogic.go:99-104 逐行核对。

### 🔴 遗留
🔴 无（冻结）。

### 🟡 修复（本轮一次修完）
1. tasks 1.3「或显式说明」退路 → 删除，改为**强制新增 AVAILABLE 券子用例**覆盖 90011 分支（已实证该分支全仓仅本文件两处断言覆盖，不得以说明替代）。
2. 行号引用 98-102 → **99-104**（explore-brief/proposal/tasks 统一；指向 usecouponlogic.go:99-104 幂等容忍块）。
3. proposal「或显式说明」退路 → 收紧为"必须新增子用例"。

### 💡（本轮采纳/记录）
1. 子用例名 `优惠券状态非锁定` 在断言改 Success 后名实不符 → 实施时改名为 `已使用券幂等容忍成功`（属该子用例自身修正范围）。
2. tasks 2.1 补充"子用例声明顺序即执行顺序，新增 90011 用例须用未被消费的 AVAILABLE 券"。

### 已确认无问题
- 论证方向正确（服务 USED 幂等契约正确，测试旧语义；有代码+日志证据），非反向放宽服务。
- skip_specs 合理（async-settlement-reliability 已覆盖幂等契约，无 spec delta）；design 条件性跳过合理（单文件测试修正）。
- 红线守住：不改服务代码/不改其它断言/不删 skip/不放宽；白名单 test/rpc/coupons/。

## Verify — 2026-09-16（apply 后一致性核对）

**一致性四项**：
1. **proposal → tasks**：proposal 3 点（2 子用例断言对齐 + 90011 分支覆盖）→ tasks 1.1/1.2/1.3 对应 ✓。
2. **tasks 全完成且有证据**：6/6 勾选；证据 = `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` → `ok`（1.081s）+ `make lint` exit 0 + PR #74 合并（commit 0e56bb6）。
3. **无越界**：`git diff --name-only` = 仅 `test/rpc/coupons/coupons_use_test.go`（白名单 ✓；无服务代码/CI/依赖改动）。
4. **explore-brief C 系列红线逐条成立**：C1 断言对齐 + 新增 AVAILABLE 用例保持 90011 覆盖 ✓；C2 coupons 全绿 ✓；C3 不改服务/其它断言/不删 skip ✓；C4 注释标注依据（usecouponlogic.go:99-104）✓；C5 新增用例用 LOCK 券（已有种子，未改种子语义）✓。

**Integration 修复前后对比（归因）**：
- 前 3 次失败 run：`34946062980` / `34953758435` / `34955862416` —— 签名全部 `audit mq init attempt N/30 failed: 127.0.0.1:9200`（ES 连接拒绝）+ `service failed to listen: audit`。
- 后最新 run：`35089155705`（completed/failure）——签名 `audit mq init attempt` ×7 + `service failed to listen: audit`（与修复前完全一致）。
- **归因**：Integration 红 = **CI 中 ES 栈启动失败（audit 无法 listen）**，与 F1 无关（F1 只改 coupons 断言；栈未起、测试断言未执行，失败计数 = 0）。此归因作为第 3 组（fix-integration-stack-readiness）立案依据。
- **未写成"已转绿/应转绿"**：Integration 当前仍红（ES 栈就绪问题，另案）。
