## Why

集成测试 `Test_UseCouponLogic_UseCoupon` 的两个子用例（`优惠券状态非锁定`、`重复使用优惠券`）断言过时：期望 `90011 CouponStatusInvalid`，但服务已实现"已使用（USED）券幂等容忍为成功"的契约（`services/coupons/internal/logic/usecouponlogic.go:99-104`，specs 业务幂等兜底契约），服务日志实证 `use coupon success cid=USED20250525001`。这 2 个过时断言使 Integration 无法全绿，阻塞后续"阶段 2（Integration 转 required）"。

## What Changes

- 更新 `test/rpc/coupons/coupons_use_test.go` 的 2 个子用例断言，对齐服务"USED 幂等容忍成功"契约（期望 `Success(0)` 而非 `90011`），并注释标注依据（usecouponlogic.go:99-104 + 服务日志）。
- **保持分支覆盖不降级**：`90011 CouponStatusInvalid` 分支（非 USED 且非 LOCKED，如 AVAILABLE 直接 Use）改用 AVAILABLE 券（未 Lock 直接 Use）**新增子用例**覆盖（不得以"说明"替代新增），不因本次修改丢失该分支的测试覆盖。
- 不改服务代码、不改其它用例断言、不删/skip 用例、不放宽检查。

## Capabilities

### New Capabilities
<!-- 无：本变更为测试断言修正，不改任何系统运行时行为（服务契约不变）。 -->

### Modified Capabilities
<!-- 无：服务"UseCoupon 对 USED 幂等容忍成功"契约已由既有 specs（async-settlement-reliability 业务幂等兜底）覆盖，本变更仅使测试对齐该既有契约，无 spec 级行为变化。 -->

## Impact

- **测试代码**：`test/rpc/coupons/coupons_use_test.go`（2 个子用例断言 + 可能的 AVAILABLE 券分支覆盖用例）。
- **服务代码**：无（`usecouponlogic.go` 契约正确，不改）。
- **门禁**：Integration 测试套 coupons 包转绿（为阶段 2 观察前提之一）。
- **风险**：低（仅测试断言对齐既有契约；需确保分支覆盖不降级）。