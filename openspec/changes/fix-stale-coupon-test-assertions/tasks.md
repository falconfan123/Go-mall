> 单文件测试断言修正（test/rpc/coupons/coupons_use_test.go）；不改服务代码。验证命令在 test/rpc 下执行。

## 1. 断言对齐服务幂等契约

- [ ] 1.1 更新子用例 `优惠券状态非锁定`：UseCoupon(USED 券) 断言由 `90011 CouponStatusInvalid` 改为 `Success(0)`（对齐 usecouponlogic.go:99-104 USED 幂等容忍），并加注释标注依据（幂等契约 + 服务日志 `use coupon success`）
- [ ] 1.2 更新子用例 `重复使用优惠券`：第二次 UseCoupon 断言由 `90011` 改为 `Success(0)`（同幂等契约），加注释
- [ ] 1.3 **强制**保持 `90011 CouponStatusInvalid` 分支覆盖不降级：必须新增一个 AVAILABLE 券（未 Lock 直接 Use）子用例覆盖该分支（依据 explore-brief D2；已实证 90011 分支全仓仅本文件两处断言覆盖，不得用"说明"替代新增）；使用 seed.go 中未被本测试函数消费的 AVAILABLE 券（LOCK/RELEASE，status=1），子用例顺序即执行顺序

## 2. 验证

- [ ] 2.1 coupons 包全绿：`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 ./coupons/...` → 0 失败（附输出）
- [ ] 2.2 断言改动核查：`git diff` 仅 `coupons_use_test.go` 的断言/注释，无服务代码改动、无 t.Skip、无删除用例；`git diff --name-only` 白名单 = test/rpc/coupons/
- [ ] 2.3 门禁（Rule 1）：`make lint` 通过（测试代码格式/vet）；确认未影响其它包（`go test ./coupons/...` 已覆盖）