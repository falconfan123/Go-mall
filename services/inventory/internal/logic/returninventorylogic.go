package logic

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/inventory/internal/svc"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReturnInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReturnInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReturnInventoryLogic {
	return &ReturnInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ReturnInventory 退还库存（支付失败时）
func (l *ReturnInventoryLogic) ReturnInventory(in *inventory.InventoryReq) (*inventory.InventoryResp, error) {
	var res = new(inventory.InventoryResp)

	//将id和数量分别存入数组
	productId := make([]int32, len(in.Items))
	quantity := make([]int32, len(in.Items))
	for i, item := range in.Items {
		productId[i] = item.ProductId
		quantity[i] = item.Quantity
	}

	// saga 补偿分支路径（barrier 幂等；业务规则失败 → Aborted；基础设施失败 → 退避重试）
	if bar, berr := dtmgrpc.BarrierFromGrpc(l.ctx); berr == nil {
		bar.DBType = "postgres"
		err := bar.CallWithDB(l.svcCtx.DtmDB, func(tx *sql.Tx) error {
			return l.svcCtx.InventoryModel.BatchReturnInventoryAtomWithSession(
				l.ctx, sqlx.NewSessionFromTx(tx), productId, quantity, in.PreOrderId, int64(in.UserId))
		})
		switch {
		case err == nil:
			for _, item := range in.Items {
				if _, cerr := l.svcCtx.AdjustInventoryCacheCtx(l.ctx, int64(item.ProductId), int64(item.Quantity)); cerr != nil {
					l.Logger.Errorw("return inventory cache adjust failed",
						logx.Field("err", cerr),
						logx.Field("product_id", item.ProductId),
						logx.Field("pre_order_id", in.PreOrderId),
					)
				}
			}
			return res, nil
		case errors.Is(err, sqlx.ErrNotFound):
			l.Logger.Infow("product not in inventory", logx.Field("product_id", productId))
			return nil, status.Error(codes.Aborted, code.ProductNotFoundInventoryMsg)
		case errors.Is(err, biz.ErrReturnAlreadyLocked):
			// 幂等容忍（specs"业务幂等兜底"契约）：该订单回补已在进行/已完成，
			// 视为成功——否则 saga 补偿重试会被自有锁行卡死（实施期发现）
			l.Logger.Infow("return already locked, tolerate as success", logx.Field("pre_order_id", in.PreOrderId))
			for _, item := range in.Items {
				if _, cerr := l.svcCtx.AdjustInventoryCacheCtx(l.ctx, int64(item.ProductId), int64(item.Quantity)); cerr != nil {
					logx.Errorw("return inventory cache adjust failed",
						logx.Field("err", cerr),
						logx.Field("product_id", item.ProductId),
						logx.Field("pre_order_id", in.PreOrderId),
					)
				}
			}
			return res, nil
		default:
			l.Logger.Errorw("return inventory failed", logx.Field("product_id", productId), logx.Field("err", err))
			return nil, err
		}
	}

	// 事务
	err := l.svcCtx.InventoryModel.BatchReturnInventoryAtom(l.ctx, productId, quantity, in.PreOrderId, int64(in.UserId))

	switch {
	case errors.Is(err, sqlx.ErrNotFound):
		l.Logger.Infow("product not in inventory", logx.Field("product_id", productId))
		res.StatusCode = code.ProductNotFoundInventory
		res.StatusMsg = code.ProductNotFoundInventoryMsg
		return res, nil

	case errors.Is(err, biz.ErrInventoryNotEnough):
		l.Logger.Infow("product inventory not enough", logx.Field("product_id", productId))
		res.StatusCode = code.InventoryNotEnough
		res.StatusMsg = code.InventoryNotEnoughMsg
		return res, nil

	case errors.Is(err, biz.ErrInventoryDecreaseFailed):
		l.Logger.Errorw("product inventory decrease failed", logx.Field("product_id", productId))
		return nil, err
	}
	if err != nil {
		l.Logger.Errorw("product inventory decrease failed", logx.Field("product_id", productId))
		return nil, err
	}

	for _, item := range in.Items {
		if _, err := l.svcCtx.AdjustInventoryCacheCtx(l.ctx, int64(item.ProductId), int64(item.Quantity)); err != nil {
			l.Logger.Errorw("return inventory cache adjust failed",
				logx.Field("err", err),
				logx.Field("product_id", item.ProductId),
				logx.Field("pre_order_id", in.PreOrderId),
			)
		}
	}

	return res, nil
}
