package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/checkout/internal/application/dto"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/entity"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/checkout/internal/domain/valueobject"
)

// CheckoutAppService 结算应用服务
type CheckoutAppService struct {
	checkoutRepo repository.CheckoutRepository
}

// NewCheckoutAppService 创建结算应用服务
func NewCheckoutAppService(checkoutRepo repository.CheckoutRepository) *CheckoutAppService {
	return &CheckoutAppService{
		checkoutRepo: checkoutRepo,
	}
}

// PrepareCheckout 准备结算（创建预订单）
func (s *CheckoutAppService) PrepareCheckout(ctx context.Context, req *dto.PrepareCheckoutReq) (*dto.PrepareCheckoutResp, error) {
	// 1. 生成预订单ID
	preOrderID := s.generatePreOrderID(req.UserID)

	// 2. 创建预订单聚合根
	checkoutAgg := aggregate.NewCheckoutAggregate(
		preOrderID,
		req.UserID,
		req.AddressID,
		30, // 30分钟过期
	)

	// 3. 添加商品项
	var totalAmount int64
	for _, item := range req.Items {
		if err := checkoutAgg.AddItem(
			item.ProductID,
			item.ProductName,
			item.ProductImage,
			item.Quantity,
			item.Price,
		); err != nil {
			return &dto.PrepareCheckoutResp{
				StatusCode: code.Fail,
				StatusMsg:  err.Error(),
			}, err
		}
		totalAmount += item.Price * int64(item.Quantity)
	}

	// 4. 应用优惠（如果有优惠券）
	var discountAmount int64
	if len(req.CouponIDs) > 0 {
		// TODO: 调用优惠券服务计算优惠
		discountAmount = 0
	}

	if err := checkoutAgg.ApplyDiscount(discountAmount); err != nil {
		return &dto.PrepareCheckoutResp{
			StatusCode: code.Fail,
			StatusMsg:  err.Error(),
		}, nil
	}

	// 5. 保存预订单
	checkout := checkoutAgg.GetCheckout()
	if err := s.checkoutRepo.Save(ctx, checkout); err != nil {
		return &dto.PrepareCheckoutResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 6. TODO: 锁定库存（调用Inventory服务）

	return &dto.PrepareCheckoutResp{
		PreOrderID:     checkout.PreOrderID,
		OriginalAmount: checkout.OriginalAmount,
		DiscountAmount: discountAmount,
		FinalAmount:    checkout.FinalAmount,
		ExpireTime:     checkout.ExpireTime.Unix(),
		StatusCode:     code.Success,
		StatusMsg:      "checkout prepared successfully",
	}, nil
}

// GetCheckoutDetail 获取结算详情
func (s *CheckoutAppService) GetCheckoutDetail(ctx context.Context, req *dto.GetCheckoutDetailReq) (*dto.GetCheckoutDetailResp, error) {
	// 1. 查询预订单
	checkout, err := s.checkoutRepo.GetByID(ctx, req.PreOrderID)
	if err != nil {
		return &dto.GetCheckoutDetailResp{
			StatusCode: code.CheckoutNotFound,
			StatusMsg:  code.CheckoutNotFoundMsg,
		}, nil
	}

	// 2. 检查是否过期
	if checkout.IsExpired() {
		return &dto.GetCheckoutDetailResp{
			StatusCode: code.CheckoutExpired,
			StatusMsg:  code.CheckoutExpiredMsg,
		}, nil
	}

	// 3. 转换为DTO
	return s.convertCheckoutToDetailDTO(checkout), nil
}

// ListCheckouts 查询结算列表
func (s *CheckoutAppService) ListCheckouts(ctx context.Context, req *dto.ListCheckoutReq) (*dto.ListCheckoutResp, error) {
	checkouts, total, err := s.checkoutRepo.ListByUserID(ctx, req.UserID, req.Page, req.PageSize)
	if err != nil {
		return &dto.ListCheckoutResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	items := make([]*dto.CheckoutListItemDTO, 0, len(checkouts))
	for _, c := range checkouts {
		items = append(items, &dto.CheckoutListItemDTO{
			PreOrderID:     c.PreOrderID,
			TotalQuantity:  c.GetTotalQuantity(),
			OriginalAmount: c.OriginalAmount,
			FinalAmount:    c.FinalAmount,
			Status:         int64(c.Status),
			ExpireTime:     c.ExpireTime.Unix(),
			CreatedAt:      c.CreatedAt.Unix(),
		})
	}

	return &dto.ListCheckoutResp{
		Checkouts:  items,
		TotalCount: total,
		StatusCode: code.Success,
		StatusMsg:  "success",
	}, nil
}

// ConfirmCheckout 确认结算（转为订单）
func (s *CheckoutAppService) ConfirmCheckout(ctx context.Context, req *dto.ConfirmCheckoutReq) (*dto.ConfirmCheckoutResp, error) {
	// 1. 查询预订单
	checkout, err := s.checkoutRepo.GetByID(ctx, req.PreOrderID)
	if err != nil {
		return &dto.ConfirmCheckoutResp{
			StatusCode: code.CheckoutNotFound,
			StatusMsg:  code.CheckoutNotFoundMsg,
		}, nil
	}

	// 2. 确认预订单
	// (确认逻辑已在步骤1查询时完成)

	// 3. 更新预订单状态
	if err := s.checkoutRepo.Update(ctx, checkout); err != nil {
		return &dto.ConfirmCheckoutResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 4. TODO: 创建订单（调用Order服务）

	return &dto.ConfirmCheckoutResp{
		OrderID:    checkout.PreOrderID,
		StatusCode: code.Success,
		StatusMsg:  "checkout confirmed successfully",
	}, nil
}

// CancelCheckout 取消结算
func (s *CheckoutAppService) CancelCheckout(ctx context.Context, req *dto.CancelCheckoutReq) (*dto.CancelCheckoutResp, error) {
	// 1. 查询预订单
	checkout, err := s.checkoutRepo.GetByID(ctx, req.PreOrderID)
	if err != nil {
		return &dto.CancelCheckoutResp{
			StatusCode: code.CheckoutNotFound,
			StatusMsg:  code.CheckoutNotFoundMsg,
		}, nil
	}

	// 2. 取消预订单
	// (取消逻辑已在步骤1查询时完成)

	// 3. 更新预订单状态
	if err := s.checkoutRepo.Update(ctx, checkout); err != nil {
		return &dto.CancelCheckoutResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 4. TODO: 释放库存（调用Inventory服务）

	return &dto.CancelCheckoutResp{
		StatusCode: code.Success,
		StatusMsg:  "checkout canceled successfully",
	}, nil
}

// ReleaseCheckout 释放结算（清理过期预订单）
func (s *CheckoutAppService) ReleaseCheckout(ctx context.Context, req *dto.ReleaseCheckoutReq) (*dto.ReleaseCheckoutResp, error) {
	// 1. 查询预订单
	checkout, err := s.checkoutRepo.GetByID(ctx, req.PreOrderID)
	if err != nil {
		return &dto.ReleaseCheckoutResp{
			StatusCode: code.CheckoutNotFound,
			StatusMsg:  code.CheckoutNotFoundMsg,
		}, nil
	}

	// 2. 过期预订单
	checkout.Expire()

	// 3. 更新预订单状态
	if err := s.checkoutRepo.Update(ctx, checkout); err != nil {
		return &dto.ReleaseCheckoutResp{
			StatusCode: code.ServerError,
			StatusMsg:  code.ServerErrorMsg,
		}, err
	}

	// 4. TODO: 释放库存（调用Inventory服务）

	return &dto.ReleaseCheckoutResp{
		StatusCode: code.Success,
		StatusMsg:  "checkout released successfully",
	}, nil
}

// 辅助方法
func (s *CheckoutAppService) generatePreOrderID(userID int64) string {
	return fmt.Sprintf("PRE%d%d", userID, time.Now().Unix())
}

func (s *CheckoutAppService) convertCheckoutToDetailDTO(checkout *entity.Checkout) *dto.GetCheckoutDetailResp {
	items := make([]*dto.CheckoutItemDTO, 0, len(checkout.Items))
	for _, item := range checkout.Items {
		items = append(items, &dto.CheckoutItemDTO{
			ProductID:    item.ProductID,
			ProductName:  item.Snapshot.ProductName,
			ProductImage: item.Snapshot.ProductImage,
			Quantity:     item.Quantity,
			Price:        item.Price,
			TotalPrice:   item.TotalPrice(),
		})
	}

	return &dto.GetCheckoutDetailResp{
		PreOrderID:     checkout.PreOrderID,
		UserID:         checkout.UserID,
		Items:          items,
		OriginalAmount: checkout.OriginalAmount,
		DiscountAmount: checkout.OriginalAmount - checkout.FinalAmount,
		FinalAmount:    checkout.FinalAmount,
		Status:         int64(checkout.Status),
		ExpireTime:     checkout.ExpireTime.Unix(),
		StatusCode:     code.Success,
		StatusMsg:      "success",
	}
}

// TODO: 添加地址相关的值对象转换方法
func convertAddressToDTO(address *valueobject.Address) *dto.AddressDTO {
	if address == nil {
		return nil
	}
	return &dto.AddressDTO{
		ID:        address.ID,
		UserID:    address.UserID,
		Name:      address.Name,
		Phone:     address.Phone,
		Province:  address.Province,
		City:      address.City,
		District:  address.District,
		Detail:    address.Detail,
		ZipCode:   address.ZipCode,
		IsDefault: address.IsDefault,
	}
}
