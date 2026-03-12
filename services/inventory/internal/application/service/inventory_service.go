package service

import (
	"context"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/inventory/internal/application/dto"
	appevent "github.com/falconfan123/Go-mall/services/inventory/internal/application/event"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/aggregate"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/event"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/repository"
	"github.com/falconfan123/Go-mall/services/inventory/internal/domain/valueobject"
	"github.com/google/uuid"
)

// InventoryAppService 库存应用服务
type InventoryAppService struct {
	inventoryRepo repository.InventoryRepository
	eventPub      appevent.InventoryEventPublisher
}

// NewInventoryAppService 创建库存应用服务
func NewInventoryAppService(
	inventoryRepo repository.InventoryRepository,
	eventPub appevent.InventoryEventPublisher,
) *InventoryAppService {
	return &InventoryAppService{
		inventoryRepo: inventoryRepo,
		eventPub:      eventPub,
	}
}

// GetInventory 查询库存
func (s *InventoryAppService) GetInventory(ctx context.Context, req *dto.GetInventoryRequest) (*dto.GetInventoryResponse, error) {
	inventory, err := s.inventoryRepo.GetByProductID(ctx, int64(req.ProductID))
	if err != nil {
		return &dto.GetInventoryResponse{
			StatusCode: code.ServerError,
			StatusMsg:  "failed to get inventory: " + err.Error(),
		}, err
	}

	return &dto.GetInventoryResponse{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
		Inventory:  inventory.TotalStock.Value(),
		SoldCount:  inventory.SoldCount,
	}, nil
}

// PreDecreaseInventory 预扣减库存
func (s *InventoryAppService) PreDecreaseInventory(ctx context.Context, req *dto.PreDecreaseInventoryRequest) (*dto.InventoryResponse, error) {
	for _, item := range req.Items {
		// 1. 查询库存
		inventory, err := s.inventoryRepo.GetByProductID(ctx, int64(item.ProductID))
		if err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ProductNotFound,
				StatusMsg:  fmt.Sprintf("product %d not found: %v", item.ProductID, err),
			}, err
		}

		// 2. 创建预扣记录
		expireTime := time.Now().Add(15 * time.Minute) // 预扣15分钟过期
		record, err := valueobject.NewPreInventoryRecord(
			int64(item.ProductID),
			int64(item.Quantity),
			req.PreOrderID,
			int64(req.UserID),
			expireTime,
		)
		if err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.Fail,
				StatusMsg:  err.Error(),
			}, err
		}

		// 3. 领域层处理预扣
		if err := pb.PreDecrease(record); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.Fail,
				StatusMsg:  fmt.Sprintf("product %d pre decrease failed: %v", item.ProductID, err),
			}, err
		}

		// 4. 保存库存
		if err := s.inventoryRepo.Save(ctx, inventory); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ServerError,
				StatusMsg:  fmt.Sprintf("product %d save failed: %v", item.ProductID, err),
			}, err
		}

		// 5. 发布预扣事件
		evt := &event.InventoryPreDecreasedEvent{
			BaseEvent: event.BaseEvent{
				EventID:    uuid.New().String(),
				EventType:  "inventory.pre_decreased",
				OccurredAt: time.Now(),
			},
			ProductID:  int64(item.ProductID),
			Quantity:   int64(item.Quantity),
			PreOrderID: req.PreOrderID,
			UserID:     int64(req.UserID),
		}
		if err := s.eventPub.PublishInventoryPreDecreased(evt); err != nil {
			fmt.Printf("failed to publish pre decreased event: %v\n", err)
		}
	}

	return &dto.InventoryResponse{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
	}, nil
}

// DecreaseInventory 真实扣减库存
func (s *InventoryAppService) DecreaseInventory(ctx context.Context, req *dto.DecreaseInventoryRequest) (*dto.InventoryResponse, error) {
	for _, item := range req.Items {
		// 1. 查询库存
		inventory, err := s.inventoryRepo.GetByProductID(ctx, int64(item.ProductID))
		if err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ProductNotFound,
				StatusMsg:  fmt.Sprintf("product %d not found: %v", item.ProductID, err),
			}, err
		}

		// 2. 确认扣减
		if err := pb.ConfirmDecrease(req.PreOrderID, int64(item.ProductID)); err != nil {
			// 如果没有预扣记录，尝试直接扣减
			if err.Error() == "pre inventory record not found" {
				if err := pb.DirectDecrease(int64(item.Quantity)); err != nil {
					return &dto.InventoryResponse{
						StatusCode: code.Fail,
						StatusMsg:  fmt.Sprintf("product %d decrease failed: %v", item.ProductID, err),
					}, err
				}
			} else {
				return &dto.InventoryResponse{
					StatusCode: code.Fail,
					StatusMsg:  fmt.Sprintf("product %d confirm decrease failed: %v", item.ProductID, err),
				}, err
			}
		}

		// 3. 保存库存
		if err := s.inventoryRepo.Save(ctx, inventory); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ServerError,
				StatusMsg:  fmt.Sprintf("product %d save failed: %v", item.ProductID, err),
			}, err
		}

		// 4. 发布扣减事件
		evt := &event.InventoryDecreasedEvent{
			BaseEvent: event.BaseEvent{
				EventID:    uuid.New().String(),
				EventType:  "inventory.decreased",
				OccurredAt: time.Now(),
			},
			ProductID: int64(item.ProductID),
			Quantity:  int64(item.Quantity),
			OrderID:   req.PreOrderID,
		}
		if err := s.eventPub.PublishInventoryDecreased(evt); err != nil {
			fmt.Printf("failed to publish decreased event: %v\n", err)
		}
	}

	return &dto.InventoryResponse{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
	}, nil
}

// ReturnPreInventory 退还预扣库存
func (s *InventoryAppService) ReturnPreInventory(ctx context.Context, req *dto.ReturnPreInventoryRequest) (*dto.InventoryResponse, error) {
	for _, item := range req.Items {
		// 1. 查询库存
		inventory, err := s.inventoryRepo.GetByProductID(ctx, int64(item.ProductID))
		if err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ProductNotFound,
				StatusMsg:  fmt.Sprintf("product %d not found: %v", item.ProductID, err),
			}, err
		}

		// 2. 退还预扣
		if err := pb.ReturnPreInventory(req.PreOrderID, int64(item.ProductID)); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.Fail,
				StatusMsg:  fmt.Sprintf("product %d return pre inventory failed: %v", item.ProductID, err),
			}, err
		}

		// 3. 保存库存
		if err := s.inventoryRepo.Save(ctx, inventory); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ServerError,
				StatusMsg:  fmt.Sprintf("product %d save failed: %v", item.ProductID, err),
			}, err
		}

		// 4. 发布退还事件
		evt := &event.InventoryPreReturnedEvent{
			BaseEvent: event.BaseEvent{
				EventID:    uuid.New().String(),
				EventType:  "inventory.pre_returned",
				OccurredAt: time.Now(),
			},
			ProductID:  int64(item.ProductID),
			Quantity:   int64(item.Quantity),
			PreOrderID: req.PreOrderID,
		}
		if err := s.eventPub.PublishInventoryPreReturned(evt); err != nil {
			fmt.Printf("failed to publish pre returned event: %v\n", err)
		}
	}

	return &dto.InventoryResponse{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
	}, nil
}

// ReturnInventory 退还库存
func (s *InventoryAppService) ReturnInventory(ctx context.Context, req *dto.ReturnInventoryRequest) (*dto.InventoryResponse, error) {
	for _, item := range req.Items {
		// 1. 查询库存
		inventory, err := s.inventoryRepo.GetByProductID(ctx, int64(item.ProductID))
		if err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ProductNotFound,
				StatusMsg:  fmt.Sprintf("product %d not found: %v", item.ProductID, err),
			}, err
		}

		// 2. 退还库存
		if err := pb.ReturnInventory(int64(item.Quantity)); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.Fail,
				StatusMsg:  fmt.Sprintf("product %d return inventory failed: %v", item.ProductID, err),
			}, err
		}

		// 3. 保存库存
		if err := s.inventoryRepo.Save(ctx, inventory); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ServerError,
				StatusMsg:  fmt.Sprintf("product %d save failed: %v", item.ProductID, err),
			}, err
		}

		// 4. 发布退还事件
		evt := &event.InventoryReturnedEvent{
			BaseEvent: event.BaseEvent{
				EventID:    uuid.New().String(),
				EventType:  "inventory.returned",
				OccurredAt: time.Now(),
			},
			ProductID: int64(item.ProductID),
			Quantity:  int64(item.Quantity),
			OrderID:   req.OrderID,
		}
		if err := s.eventPub.PublishInventoryReturned(evt); err != nil {
			fmt.Printf("failed to publish returned event: %v\n", err)
		}
	}

	return &dto.InventoryResponse{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
	}, nil
}

// UpdateInventory 更新库存
func (s *InventoryAppService) UpdateInventory(ctx context.Context, req *dto.UpdateInventoryRequest) (*dto.InventoryResponse, error) {
	for _, item := range req.Items {
		// 1. 查询库存
		inventory, err := s.inventoryRepo.GetByProductID(ctx, int64(item.ProductID))
		if err != nil {
			// 如果不存在，创建新库存
			newStock, err := valueobject.NewStock(int64(item.Quantity))
			if err != nil {
				return &dto.InventoryResponse{
					StatusCode: code.Fail,
					StatusMsg:  err.Error(),
				}, err
			}
			inventory = aggregate.NewInventory(int64(item.ProductID), newStock)
		} else {
			// 更新库存
			newStock, err := valueobject.NewStock(int64(item.Quantity))
			if err != nil {
				return &dto.InventoryResponse{
					StatusCode: code.Fail,
					StatusMsg:  err.Error(),
				}, err
			}
			oldStock := pb.TotalStock.Value()
			inventory.UpdateStock(newStock)

			// 发布更新事件
			evt := &event.InventoryUpdatedEvent{
				BaseEvent: event.BaseEvent{
					EventID:    uuid.New().String(),
					EventType:  "inventory.updated",
					OccurredAt: time.Now(),
				},
				ProductID:     int64(item.ProductID),
				OldStock:      oldStock,
				NewStock:      newStock.Value(),
				ChangedAmount: newStock.Value() - oldStock,
			}
			if err := s.eventPub.PublishInventoryUpdated(evt); err != nil {
				fmt.Printf("failed to publish updated event: %v\n", err)
			}
		}

		// 2. 保存库存
		if err := s.inventoryRepo.Save(ctx, inventory); err != nil {
			return &dto.InventoryResponse{
				StatusCode: code.ServerError,
				StatusMsg:  fmt.Sprintf("product %d save failed: %v", item.ProductID, err),
			}, err
		}
	}

	return &dto.InventoryResponse{
		StatusCode: code.Success,
		StatusMsg:  code.SuccessMsg,
	}, nil
}
