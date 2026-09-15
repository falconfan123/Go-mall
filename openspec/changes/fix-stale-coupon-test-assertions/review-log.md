# Review Log · fix-stale-coupon-test-assertions

## 整体反思审核（proposal + tasks；specs skip_specs、design 条件性跳过）

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
