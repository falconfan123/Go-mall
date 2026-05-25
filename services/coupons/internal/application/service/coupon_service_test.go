package service

import (
	"context"
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/coupons/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/valueobject"
	"github.com/stretchr/testify/require"
)

func TestCouponAppServiceCreateClaimAndCalculate(t *testing.T) {
	t.Parallel()

	repo := &stubCouponRepository{}
	userRepo := &stubUserCouponRepository{}
	service := NewCouponAppService(repo, userRepo)

	start := time.Now().Add(-time.Hour).Unix()
	end := time.Now().Add(time.Hour).Unix()
	createResp, err := service.CreateCoupon(context.Background(), &dto.CreateCouponReq{
		Name:       "coupon",
		CouponType: int64(valueobject.CouponTypeFullReduction),
		Value:      100,
		MinAmount:  0,
		StartTime:  start,
		EndTime:    end,
		TotalCount: 10,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, createResp.StatusCode)

	repo.coupon = repo.savedCoupon
	claimResp, err := service.ClaimCoupon(context.Background(), &dto.ClaimCouponReq{
		UserID:   1,
		CouponID: repo.coupon.ID,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, claimResp.StatusCode)

	calcResp, err := service.CalculateCoupon(context.Background(), &dto.CalculateCouponReq{
		CouponID:    0,
		OrderAmount: 1000,
	})
	require.NoError(t, err)
	require.EqualValues(t, code.CouponsNotExist, calcResp.StatusCode)
}

func TestCouponAppServiceUseCoupon(t *testing.T) {
	t.Parallel()

	userCoupon := entity.NewUserCoupon(1, "coupon-1")
	service := NewCouponAppService(&stubCouponRepository{}, &stubUserCouponRepository{userCoupon: userCoupon})

	resp, err := service.UseCoupon(context.Background(), &dto.UseCouponReq{
		CouponID: 1,
		OrderID:  "ord-1",
	})
	require.NoError(t, err)
	require.EqualValues(t, code.Success, resp.StatusCode)
}

type stubCouponRepository struct {
	coupon      *aggregate.Coupon
	savedCoupon *aggregate.Coupon
}

func (s *stubCouponRepository) GetByID(_ context.Context, id string) (*aggregate.Coupon, error) {
	if s.coupon == nil || s.coupon.ID != id {
		return nil, entity.ErrUserCouponExpired
	}
	return s.coupon, nil
}

func (s *stubCouponRepository) Save(_ context.Context, coupon *aggregate.Coupon) error {
	s.savedCoupon = coupon
	s.coupon = coupon
	return nil
}

func (s *stubCouponRepository) Update(context.Context, *aggregate.Coupon) error {
	return nil
}

func (s *stubCouponRepository) Delete(context.Context, string) error {
	return nil
}

func (s *stubCouponRepository) List(context.Context, int, int) ([]*aggregate.Coupon, int64, error) {
	return nil, 0, nil
}

func (s *stubCouponRepository) ListAvailable(context.Context, int, int) ([]*aggregate.Coupon, int64, error) {
	if s.coupon == nil {
		return nil, 0, nil
	}
	return []*aggregate.Coupon{s.coupon}, 1, nil
}

func (s *stubCouponRepository) DecreaseStock(context.Context, string, int) error {
	return nil
}

func (s *stubCouponRepository) IncreaseStock(context.Context, string, int) error {
	return nil
}

type stubUserCouponRepository struct {
	userCoupon *entity.UserCoupon
}

func (s *stubUserCouponRepository) GetByID(context.Context, int64) (*entity.UserCoupon, error) {
	return s.userCoupon, nil
}

func (s *stubUserCouponRepository) GetByUserIDAndCouponID(context.Context, int64, string) (*entity.UserCoupon, error) {
	return s.userCoupon, nil
}

func (s *stubUserCouponRepository) ListByUserID(context.Context, int64, *entity.UserCouponStatus, int, int) ([]*entity.UserCoupon, int64, error) {
	if s.userCoupon == nil {
		return nil, 0, nil
	}
	return []*entity.UserCoupon{s.userCoupon}, 1, nil
}

func (s *stubUserCouponRepository) Save(context.Context, *entity.UserCoupon) error {
	return nil
}

func (s *stubUserCouponRepository) Update(context.Context, *entity.UserCoupon) error {
	return nil
}

func (s *stubUserCouponRepository) Delete(context.Context, int64) error {
	return nil
}

func (s *stubUserCouponRepository) CountByUserIDAndCouponID(context.Context, int64, string) (int64, error) {
	return 0, nil
}

func (s *stubUserCouponRepository) FindAvailableByUserID(context.Context, int64, int64) ([]*entity.UserCoupon, error) {
	return nil, nil
}

var _ repository.CouponRepository = (*stubCouponRepository)(nil)
var _ repository.UserCouponRepository = (*stubUserCouponRepository)(nil)
