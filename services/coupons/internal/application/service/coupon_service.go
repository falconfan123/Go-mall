package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/coupons/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/coupons/internal/domain/valueobject"
)

// CouponAppService 优惠券应用服务
type CouponAppService struct {
	couponRepo     repository.CouponRepository
	userCouponRepo repository.UserCouponRepository
}

// NewCouponAppService 创建优惠券应用服务
func NewCouponAppService(
	couponRepo repository.CouponRepository,
	userCouponRepo repository.UserCouponRepository,
) *CouponAppService {
	return &CouponAppService{
		couponRepo:     couponRepo,
		userCouponRepo: userCouponRepo,
	}
}

// CreateCoupon 创建优惠券
func (s *CouponAppService) CreateCoupon(ctx context.Context, req *dto.CreateCouponReq) (*dto.CreateCouponResp, error) {
	// 1. 创建优惠券聚合根
	couponType := valueobject.CouponType(req.CouponType)
	coupon, err := aggregate.NewCoupon(
		fmt.Sprintf("coupon_%d", time.Now().Unix()),
		req.Name,
		couponType,
		req.Value,
		req.MinAmount,
		time.Unix(req.StartTime, 0),
		time.Unix(req.EndTime, 0),
		req.TotalCount,
	)
	if err != nil {
		return &dto.CreateCouponResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, err
	}

	// 2. 保存优惠券
	if err := s.couponRepo.Save(ctx, coupon); err != nil {
		return &dto.CreateCouponResp{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to save coupon: " + err.Error(),
		}, err
	}

	return &dto.CreateCouponResp{
		CouponID:   coupon.ID,
		StatusCode: code.Success,
		StatusMsg:  "coupon created successfully",
	}, nil
}

// ClaimCoupon 用户领取优惠券
func (s *CouponAppService) ClaimCoupon(ctx context.Context, req *dto.ClaimCouponReq) (*dto.ClaimCouponResp, error) {
	// 1. 检查用户是否已领取
	count, err := s.userCouponRepo.CountByUserIDAndCouponID(ctx, req.UserID, req.CouponID)
	if err != nil {
		return &dto.ClaimCouponResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}
	if count > 0 {
		return &dto.ClaimCouponResp{
			StatusCode: code.CouponsAlreadyClaimed,
			StatusMsg:  code.CouponsAlreadyClaimedMsg,
		}, nil
	}

	// 2. 检查优惠券库存
	coupon, err := s.couponRepo.GetByID(ctx, req.CouponID)
	if err != nil {
		return &dto.ClaimCouponResp{
			StatusCode: code.CouponsNotExist,
			StatusMsg:  code.CouponsNotExistMsg,
		}, nil
	}

	// 3. 检查是否可领取
	if err := coupon.CanClaim(); err != nil {
		return &dto.ClaimCouponResp{
			StatusCode: code.CouponsOutOfStock,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 4. 扣减库存
	if err := s.couponRepo.DecreaseStock(ctx, req.CouponID, 1); err != nil {
		return &dto.ClaimCouponResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 5. 创建用户优惠券
	userCoupon := entity.NewUserCoupon(req.UserID, req.CouponID)
	if err := s.userCouponRepo.Save(ctx, userCoupon); err != nil {
		// 归还库存
		_ = s.couponRepo.IncreaseStock(ctx, req.CouponID, 1)
		return &dto.ClaimCouponResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.ClaimCouponResp{
		Coupon:     s.convertCouponToDTO(coupon),
		StatusCode: code.Success,
		StatusMsg:  "coupon claimed successfully",
	}, nil
}

// UseCoupon 使用优惠券
func (s *CouponAppService) UseCoupon(ctx context.Context, req *dto.UseCouponReq) (*dto.UseCouponResp, error) {
	// 1. 查询用户优惠券
	userCoupon, err := s.userCouponRepo.GetByID(ctx, req.CouponID)
	if err != nil {
		return &dto.UseCouponResp{
			StatusCode: code.CouponsNotExist,
			StatusMsg:  code.CouponsNotExistMsg,
		}, nil
	}

	// 2. 检查是否可使用
	if err := userCoupon.CanUse(); err != nil {
		return &dto.UseCouponResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 3. 使用优惠券
	if err := userCoupon.Use(req.OrderID); err != nil {
		return &dto.UseCouponResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 4. 更新用户优惠券
	if err := s.userCouponRepo.Update(ctx, userCoupon); err != nil {
		return &dto.UseCouponResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	return &dto.UseCouponResp{
		StatusCode: code.Success,
		StatusMsg:  "coupon used successfully",
	}, nil
}

// CalculateCoupon 计算优惠金额
func (s *CouponAppService) CalculateCoupon(ctx context.Context, req *dto.CalculateCouponReq) (*dto.CalculateCouponResp, error) {
	// 1. 查询优惠券
	couponID := fmt.Sprintf("%d", req.CouponID)
	coupon, err := s.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		return &dto.CalculateCouponResp{
			StatusCode: code.CouponsNotExist,
			StatusMsg:  code.CouponsNotExistMsg,
		}, nil
	}

	// 2. 检查是否可使用
	if err := coupon.CanUse(req.OrderAmount); err != nil {
		return &dto.CalculateCouponResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 3. 计算优惠金额
	discountAmount, err := coupon.CalculateDiscount(req.OrderAmount)
	if err != nil {
		return &dto.CalculateCouponResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	return &dto.CalculateCouponResp{
		DiscountAmount: discountAmount,
		FinalAmount:    req.OrderAmount - discountAmount,
		StatusCode:     code.Success,
		StatusMsg:      "calculated successfully",
	}, nil
}

// ListCoupons 查询优惠券列表
func (s *CouponAppService) ListCoupons(ctx context.Context, req *dto.ListCouponsReq) (*dto.ListCouponsResp, error) {
	coupons, total, err := s.couponRepo.ListAvailable(ctx, req.Page, req.PageSize)
	if err != nil {
		return &dto.ListCouponsResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	dtos := make([]*dto.CouponDTO, 0, len(coupons))
	for _, c := range coupons {
		dtos = append(dtos, s.convertCouponToDTO(c))
	}

	return &dto.ListCouponsResp{
		Coupons:    dtos,
		TotalCount: total,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ListUserCoupons 查询用户优惠券列表
func (s *CouponAppService) ListUserCoupons(ctx context.Context, req *dto.ListUserCouponsReq) (*dto.ListUserCouponsResp, error) {
	var status *entity.UserCouponStatus
	if req.Status != nil {
		s := entity.UserCouponStatus(*req.Status)
		status = &s
	}

	userCoupons, total, err := s.userCouponRepo.ListByUserID(ctx, req.UserID, status, req.Page, req.PageSize)
	if err != nil {
		return &dto.ListUserCouponsResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	dtos := make([]*dto.UserCouponDTO, 0, len(userCoupons))
	for _, uc := range userCoupons {
		dtos = append(dtos, s.convertUserCouponToDTO(uc))
	}

	return &dto.ListUserCouponsResp{
		UserCoupons: dtos,
		TotalCount:  total,
		StatusCode:  code.Success,
		StatusMsg:   "success",
	}, nil
}

// GetCoupon 获取优惠券详情
func (s *CouponAppService) GetCoupon(ctx context.Context, req *dto.GetCouponReq) (*dto.GetCouponResp, error) {
	coupon, err := s.couponRepo.GetByID(ctx, req.CouponID)
	if err != nil {
		return &dto.GetCouponResp{
			StatusCode: code.CouponsNotExist,
			StatusMsg:  code.CouponsNotExistMsg,
		}, nil
	}

	return &dto.GetCouponResp{
		Coupon:     s.convertCouponToDTO(coupon),
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// 转换方法
func (s *CouponAppService) convertCouponToDTO(coupon *aggregate.Coupon) *dto.CouponDTO {
	return &dto.CouponDTO{
		ID:             coupon.ID,
		Name:           coupon.Name,
		CouponType:     int64(coupon.Discount.CouponType()),
		Value:          coupon.Discount.Value(),
		MinAmount:      coupon.Discount.MinAmount(),
		TotalCount:     coupon.TotalCount,
		RemainingCount: coupon.RemainingCount,
		StartTime:      coupon.ValidPeriod.StartTime().Unix(),
		EndTime:        coupon.ValidPeriod.EndTime().Unix(),
		Status:         int64(coupon.Status),
	}
}

func (s *CouponAppService) convertUserCouponToDTO(userCoupon *entity.UserCoupon) *dto.UserCouponDTO {
	dto := &dto.UserCouponDTO{
		ID:       userCoupon.ID,
		UserID:   userCoupon.UserID,
		CouponID: userCoupon.CouponID,
		Status:   int64(userCoupon.Status),
		GetTime:  userCoupon.GetTime.Unix(),
	}

	if userCoupon.UseTime != nil {
		useTime := userCoupon.UseTime.Unix()
		dto.UseTime = &useTime
	}

	if userCoupon.OrderID != nil {
		dto.OrderID = *userCoupon.OrderID
	}

	return dto
}
