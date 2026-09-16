# Design · fix-stale-coupon-test-assertions

## Context

见 `proposal.md` — Why。单文件测试断言修正（`test/rpc/coupons/coupons_use_test.go`），对齐服务 `usecouponlogic.go:99-104` 的 USED 幂等容忍契约。

## Goals / Non-Goals

- **Goals**：2 个子用例断言对齐契约（90011 → Success(0)）；90011 分支覆盖不降级（新增 AVAILABLE 券用例）；`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` 全绿。
- **Non-Goals**：不改服务代码；不改其它用例断言；不删/skip 用例；不放宽检查。

## Decisions

无架构/设计决策（单文件断言修正，无跨服务/无新依赖/无安全性能迁移复杂度；specs skip_specs）。修复直接按 `explore-brief.md` D1/D2 执行：断言改 Success + 新增 AVAILABLE 券用例覆盖 90011 分支（子用例声明顺序=执行顺序，使用本测试函数内未被消费的 LOCK 券）。

## Risks / Trade-offs

- 低：断言对齐既有契约（有代码 + 服务日志证据）；90011 分支覆盖由新增用例保持。
- 无回滚负担（git revert 单文件改动）。

## Migration Plan

1. 改 2 子用例断言 + 注释依据；2. 新增 AVAILABLE 用例；3. coupons 全绿验证 + make lint + 白名单核查；4. 提 PR 合并。

## Open Questions

无。