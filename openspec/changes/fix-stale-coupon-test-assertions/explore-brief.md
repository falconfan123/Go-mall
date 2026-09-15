# Explore Brief · fix-stale-coupon-test-assertions

> 探索结论落盘（来源：fix-ci-integration-pipeline 的发现项 F1），作为提案审核需求基线。

## 背景（实证）

- 集成测试 `test/rpc/coupons/coupons_use_test.go` 的 `Test_UseCouponLogic_UseCoupon` 两个子用例失败：
  - `优惠券状态非锁定`：UseCoupon(uid=1, CouponId=USED20250525001，user_coupons 种子 status=USED) 期望 `90011 CouponStatusInvalid`。
  - `重复使用优惠券`：UseCoupon(lock→use→再用) 第二次期望 `90011 CouponStatusInvalid`。
- **服务实际行为**（正确契约）：`services/coupons/internal/logic/usecouponlogic.go:99-104` —— `if status == CouponStatus_USED { 容忍为成功 return nil }`（幂等兜底契约，specs 业务幂等兜底：支付成功确认的重复到达不应报错）。服务日志实证：`use coupon success cid=USED20250525001`。
- 即：**服务行为正确（USED 幂等容忍成功），测试断言为旧语义（期望 90011）**。
- 种子修复（前序 change D8）已提供 user_coupons 种子；Lock/Unlock 组已绿，仅剩此 2 子用例。

## 目标 / 非目标

- **目标**：更新这 2 个子用例的断言，对齐服务"USED 幂等容忍成功"契约（期望 `Success(0)`），使 `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` 全绿。
- **非目标**：不改服务行为（契约正确，不改 usecouponlogic.go）；不改其它测试用例/断言；不放宽任何检查；不删/跳过用例。

## 待决策点

- **D1 断言口径**：`优惠券状态非锁定`（USED 券直接 UseCoupon）→ 期望改为 `Success(0)`（幂等容忍）；`重复使用优惠券`（第二次 UseCoupon）→ 期望改为 `Success(0)`。需在测试注释中标注依据（usecouponlogic.go:99-104 幂等契约）。
- **D2 是否保留"状态无效"的负例覆盖**：`优惠券状态非锁定` 原意是覆盖"非 LOCKED 状态"分支。USED 已被幂等容忍（成功），真正走 90011 的是**非 USED 且非 LOCKED** 的状态（如 AVAILABLE 直接 Use）。是否新增/改用 AVAILABLE 券覆盖 90011 分支（保持分支覆盖不降级）——倾向：**保留分支覆盖**，用 AVAILABLE 券（种子的 USE 券在 UseCoupon 前未 Lock）覆盖 90011，同时把 USED 券用例改为断言 Success。

## Checklist（反思审核基线）

- [ ] C1 两个子用例断言对齐服务幂等契约（USED → Success），且**不降低分支覆盖**（90011 分支仍需被某用例覆盖，或用例显式说明）。
- [ ] C2 `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` 全绿（附输出）。
- [ ] C3 不改服务代码、不改其它用例断言、不删/skip 用例、不放宽检查。
- [ ] C4 测试注释标注依据（usecouponlogic.go:99-104 幂等契约 + 服务日志证据）。
- [ ] C5 若 D2 判定需新增覆盖，新增用例基于已有种子（不改种子语义）。