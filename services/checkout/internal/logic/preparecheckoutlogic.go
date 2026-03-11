package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/code"
	checkout2 "github.com/falconfan123/Go-mall/dal/model/checkout"
	"github.com/falconfan123/Go-mall/services/checkout/checkout"
	"github.com/falconfan123/Go-mall/services/checkout/internal/svc"
	"github.com/falconfan123/Go-mall/services/coupons/coupons"
	"github.com/falconfan123/Go-mall/services/inventory/inventory"
	"github.com/falconfan123/Go-mall/services/product/product"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

type PrepareCheckoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPrepareCheckoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PrepareCheckoutLogic {
	return &PrepareCheckoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func generatePreOrderID() (string, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// PrepareCheckout 处理预结算
func (l *PrepareCheckoutLogic) PrepareCheckout(in *checkout.CheckoutReq) (*checkout.CheckoutResp, error) {
	// 1. 生成 pre_order_id
	preOrderId, err := generatePreOrderID()
	if err != nil {
		l.Logger.Errorw("生成 preOrderId 失败",
			logx.Field("err", err),
			logx.Field("user_id", in.UserId))
		return &checkout.CheckoutResp{
			StatusCode: code.GenerateOrderFailed,
			StatusMsg:  code.GenerateOrderFailedMsg,
		}, nil
	}

	// 2. 使用 Redis 锁来保证幂等性
	cacheKey := fmt.Sprintf("checkout:preorder:%d", in.UserId)
	luaScript := `
		if redis.call("EXISTS", KEYS[1]) == 1 then
			return redis.call("GET", KEYS[1])
		else
			redis.call("SETEX", KEYS[1], ARGV[1], ARGV[2])
			return ARGV[2]
		end
	`
	result, err := l.svcCtx.RedisClient.EvalCtx(l.ctx, luaScript, []string{cacheKey}, []any{300, preOrderId})
	if err != nil {
		l.Logger.Errorw("Redis Lua 执行失败",
			logx.Field("err", err),
			logx.Field("user_id", in.UserId))
		return &checkout.CheckoutResp{
			StatusCode: code.InternalFailed,
			StatusMsg:  code.InternalFailedMsg,
			PreOrderId: preOrderId,
		}, nil
	}
	// 如果结果不是 ARGV[2]，说明是已存在的预订单
	if result != preOrderId {
		l.Logger.Infof("用户 %d 的预订单已存在，使用已有的预订单: %v", in.UserId, result)
		preOrderId = result.(string)
	}

	// 3. 检查是否有商品信息
	if len(in.OrderItems) == 0 {
		// 释放 Redis 锁
		if _, err := l.svcCtx.RedisClient.Del(cacheKey); err != nil {
			l.Logger.Errorw("删除 Redis 锁失败",
				logx.Field("err", err),
				logx.Field("user_id", in.UserId))
		}
		return &checkout.CheckoutResp{
			StatusCode: code.OrderProductEmpty,
			StatusMsg:  code.OrderProductEmptyMsg,
		}, nil
	}
	// 4. 调用库存预扣接口
	inventoryItems := make([]*inventory.InventoryReq_Items, 0)
	for _, item := range in.OrderItems {
		inventoryItems = append(inventoryItems, &inventory.InventoryReq_Items{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		})
	}
	res := &checkout.CheckoutResp{}
	inventoryRes, err := l.svcCtx.InventoryRpc.DecreasePreInventory(l.ctx, &inventory.InventoryReq{
		Items:      inventoryItems,
		PreOrderId: preOrderId,
		UserId:     int32(in.UserId),
	})

	if err != nil {
		l.Logger.Errorw("库存预扣失败，执行同步库存回滚",
			logx.Field("err", err),
			logx.Field("user_id", in.UserId),
			logx.Field("pre_order_id", preOrderId))

		l.rollback(preOrderId, int32(in.UserId), inventoryItems, cacheKey)

		return &checkout.CheckoutResp{
			StatusCode: code.OutOfInventory,
			StatusMsg:  code.OutOfInventoryMsg,
		}, nil
	}
	if inventoryRes.StatusCode != code.Success {
		// 释放 Redis 锁
		l.svcCtx.RedisClient.Del(cacheKey)
		res.StatusCode = inventoryRes.StatusCode
		res.StatusMsg = inventoryRes.StatusMsg
		return res, nil
	}
	// 5. 异步处理结算信息
	ctx := context.TODO()
	var totalPrice uint64
	var finalPrice uint64
	items := make([]*checkout2.CheckoutItems, len(in.OrderItems))
	couponsItems := make([]*couponsclient.Items, len(in.OrderItems))
	expireTime := time.Now().Add(10 * time.Minute).Unix()
	for i, item := range in.OrderItems {
		productResp, err := l.svcCtx.ProductRpc.GetProduct(ctx, &product.GetProductReq{
			Id: uint32(item.ProductId),
		})
		if err != nil {
			l.Logger.Errorw("获取商品详情失败",
				logx.Field("err", err),
				logx.Field("product_id", item.ProductId))
			l.rollback(preOrderId, int32(in.UserId), inventoryItems, cacheKey)
			return nil, err
		}
		snapshotData := map[string]interface{}{"name": productResp.Product.Name, "desc": productResp.Product.Description}
		snapshotJSON, _ := json.Marshal(snapshotData)
		items[i] = &checkout2.CheckoutItems{
			PreOrderId: preOrderId,
			ProductId:  uint64(item.ProductId),
			Quantity:   uint64(item.Quantity),
			Price:      productResp.Product.Price,
			Snapshot:   string(snapshotJSON),
		}
		couponsItems[i] = &couponsclient.Items{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		}
		totalPrice += uint64(productResp.Product.Price) * uint64(item.Quantity)

	}
	finalPrice = totalPrice
	if in.CouponId != "" {
		resp, err := l.svcCtx.CouponsRpc.CalculateCoupon(ctx, &couponsclient.CalculateCouponReq{
			CouponId: in.CouponId,
			UserId:   int32(in.UserId),
			Items:    couponsItems,
		})
		if err != nil {
			l.Logger.Errorw("计算优惠券失败",
				logx.Field("err", err),
				logx.Field("user_id", in.UserId))
			l.rollback(preOrderId, int32(in.UserId), inventoryItems, cacheKey)
			return nil, err
		}
		if resp.StatusCode != code.Success {
			res.StatusCode = int32(resp.StatusCode)
			res.StatusMsg = resp.StatusMsg
			l.rollback(preOrderId, int32(in.UserId), inventoryItems, cacheKey)
			return res, nil
		}
		finalPrice = uint64(resp.FinalAmount)
	}

	if err := l.svcCtx.Mysql.TransactCtx(ctx, func(context context.Context, session sqlx.Session) error {
		// 2. 获取商品信息，计算原始总金额并插入 checkout_items
		for _, item := range items {
			if _, err := l.svcCtx.CheckoutItemsModel.WithSession(session).Insert(ctx, item); err != nil {
				return err
			}
		}
		if _, err := l.svcCtx.CheckoutModel.Insert(ctx, &checkout2.Checkouts{
			PreOrderId:     preOrderId,
			UserId:         uint64(in.UserId),
			CouponId:       sql.NullString{String: in.CouponId, Valid: in.CouponId != ""},
			OriginalAmount: int64(totalPrice),
			FinalAmount:    int64(finalPrice),
			ExpireTime:     expireTime,
			Status:         int64(checkout.CheckoutStatus_RESERVING),
			AddressId:      in.AddressId,
		}); err != nil {
			return err
		}
		return nil

	}); err != nil {
		l.Logger.Errorw("处理结算信息失败",
			logx.Field("err", err))
		l.rollback(preOrderId, int32(in.UserId), inventoryItems, cacheKey)
		return nil, err
	}
	// 6. 返回预结算信息
	return &checkout.CheckoutResp{
		PreOrderId: preOrderId,
		ExpireTime: expireTime,
		PayMethod:  []int64{1, 2},
	}, nil
}

func (l *PrepareCheckoutLogic) rollback(preOrderId string, userId int32, items []*inventory.InventoryReq_Items, cacheKey string) {
	// 释放 Redis 锁
	if _, err := l.svcCtx.RedisClient.Del(cacheKey); err != nil {
		l.Logger.Errorw("删除 Redis 锁失败",
			logx.Field("err", err),
			logx.Field("user_id", userId))
	}

	// 回滚库存
	_, errRollback := l.svcCtx.InventoryRpc.ReturnPreInventory(l.ctx, &inventory.InventoryReq{
		Items:      items,
		PreOrderId: preOrderId,
		UserId:     userId,
	})
	if errRollback != nil {
		l.Logger.Errorw("库存回滚失败",
			logx.Field("err", errRollback),
			logx.Field("user_id", userId),
			logx.Field("pre_order_id", preOrderId))
	}
}
