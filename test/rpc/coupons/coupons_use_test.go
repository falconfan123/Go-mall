package coupons

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	coupons "github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	lockCouponCode      = "LOCK20250525001"
	releaseCouponCode   = "RELEASE20250525001"
	useCouponCode       = "USE20250525001"
	usedCouponCode      = "USED20250525001"
	duplicateCouponCode = "DUP20250525001"
)

func Test_LockCouponLogic_LockCoupon(t *testing.T) {
	seed.SeedUserCoupons(t)
	uci := lockCouponCode
	pid := "pre-lock-coupon"
	t.Run("正常情况", func(t *testing.T) {
		res, err := couponsClient.LockCoupon(context.Background(), &coupons.LockCouponReq{
			UserId:       1,
			UserCouponId: uci,
			PreOrderId:   pid,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), res.StatusCode)
	})
	t.Run("优惠卷已经锁定", func(t *testing.T) {
		res, err := couponsClient.LockCoupon(context.Background(), &coupons.LockCouponReq{
			UserId:       1,
			UserCouponId: uci,
			PreOrderId:   pid,
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log(res)
		lock, err := couponsClient.LockCoupon(context.Background(), &coupons.LockCouponReq{
			UserId:       1,
			UserCouponId: uci,
			PreOrderId:   pid,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.CouponsAlreadyLocked), lock.StatusCode)

	})
}

func Test_UnlockCouponLogic_UnlockCoupon(t *testing.T) {
	seed.SeedUserCoupons(t)
	uci := releaseCouponCode
	pid := "pre-release-coupon"
	t.Run("正常情况", func(t *testing.T) {
		res, err := couponsClient.LockCoupon(context.Background(), &coupons.LockCouponReq{
			UserId:       1,
			UserCouponId: uci,
			PreOrderId:   pid,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), res.StatusCode)
		unlock, err := couponsClient.ReleaseCoupon(context.Background(), &coupons.ReleaseCouponReq{
			UserId:       1,
			UserCouponId: uci,
			PreOrderId:   pid,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), unlock.StatusCode)
	})
	t.Run("优惠卷已经释放", func(t *testing.T) {
		res, err := couponsClient.ReleaseCoupon(context.Background(), &coupons.ReleaseCouponReq{
			UserId:       1,
			UserCouponId: uci,
			PreOrderId:   pid,
		})
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, int32(code.CouponsAlreadyReleased), res.StatusCode)
	})
}

// 用户优惠券使用情况
func Test_ListCouponsUsageLogic_ListCouponsUsage(t *testing.T) {
	res, err := couponsClient.ListCouponUsages(context.Background(), &coupons.ListCouponUsagesReq{
		Pagination: &coupons.PaginationReq{
			Page: 1,
			Size: 10,
		},
		UserId: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(res.Usages)
}

// 记录使用优惠券
func Test_UseCouponLogic_UseCoupon(t *testing.T) {
	seed.SeedUserCoupons(t)
	uid := 1

	t.Run("正常情况", func(t *testing.T) {
		preOrderID := "pre-use-coupon"
		orderID := "order-use-coupon"
		lock, err := couponsClient.LockCoupon(context.Background(), &coupons.LockCouponReq{
			UserId:       int32(uid),
			UserCouponId: useCouponCode,
			PreOrderId:   preOrderID,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), lock.StatusCode)

		res, err := couponsClient.UseCoupon(context.Background(), &coupons.UseCouponReq{
			UserId:         1,
			CouponId:       useCouponCode,
			OrderId:        orderID,
			DiscountAmount: 100,
			OriginAmount:   100,
			PreOrderId:     preOrderID,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), res.StatusCode)
	})
	t.Run("优惠券不存在", func(t *testing.T) {
		invalidCid := "INVALID_CID"
		res, err := couponsClient.UseCoupon(context.Background(), &coupons.UseCouponReq{
			UserId:         int32(uid),
			CouponId:       invalidCid,
			OrderId:        "order-invalid-coupon",
			DiscountAmount: 100,
			OriginAmount:   100,
		})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, int32(code.CouponsNotExist), res.StatusCode)
		t.Log(res)
	})
	// 已使用（USED）券再次 UseCoupon：服务按幂等契约容忍为成功
	// （usecouponlogic.go:99-104：status==USED → tolerate as success，返回 nil）
	t.Run("已使用券幂等容忍成功", func(t *testing.T) {
		res, err := couponsClient.UseCoupon(context.Background(), &coupons.UseCouponReq{
			UserId:         int32(uid),
			CouponId:       usedCouponCode,
			OrderId:        "order-used-coupon",
			DiscountAmount: 100,
			OriginAmount:   100,
		})
		if err != nil {
			t.Fatal(err)
		} // 事务错误处理方式特殊
		assert.Equal(t, int32(code.Success), res.StatusCode)

	})
	// 可用（AVAILABLE）券未锁定直接 UseCoupon → 非 USED 且非 LOCKED → CouponStatusInvalid(90011)
	// 保持 90011 分支覆盖（USED 券已改判为幂等成功，该分支需由 AVAILABLE 券覆盖）
	t.Run("可用券未锁定直接使用", func(t *testing.T) {
		res, err := couponsClient.UseCoupon(context.Background(), &coupons.UseCouponReq{
			UserId:         int32(uid),
			CouponId:       lockCouponCode,
			OrderId:        "order-available-coupon",
			DiscountAmount: 100,
			OriginAmount:   100,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.CouponStatusInvalid), res.StatusCode)

	})
	t.Run("重复使用优惠券", func(t *testing.T) {
		preOrderID := "pre-duplicate-coupon"
		orderID := "order-duplicate-coupon"
		lock, err := couponsClient.LockCoupon(context.Background(), &coupons.LockCouponReq{
			UserId:       int32(uid),
			UserCouponId: duplicateCouponCode,
			PreOrderId:   preOrderID,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), lock.StatusCode)

		res, err := couponsClient.UseCoupon(context.Background(), &coupons.UseCouponReq{
			PreOrderId:     preOrderID,
			UserId:         int32(uid),
			CouponId:       duplicateCouponCode,
			OrderId:        orderID,
			DiscountAmount: 100,
			OriginAmount:   100,
		})
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, int32(code.Success), res.GetStatusCode())
		// 第二次使用
		res2, err := couponsClient.UseCoupon(context.Background(), &coupons.UseCouponReq{
			UserId:         int32(uid),
			CouponId:       duplicateCouponCode,
			OrderId:        orderID,
			DiscountAmount: 100,
			OriginAmount:   100,
		})
		if err != nil {
			t.Fatal(err)
		}
		// 第一次使用后券状态为 USED；再次使用按幂等契约容忍为成功
		// （usecouponlogic.go:99-104：status==USED → tolerate as success）
		assert.Equal(t, int32(code.Success), res2.GetStatusCode())
	})

}
