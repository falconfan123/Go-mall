package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/services/inventory/internal/svc"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DecreaseInventoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDecreaseInventoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DecreaseInventoryLogic {
	return &DecreaseInventoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DecreaseInventory 扣减库存（Saga branch-02）。
//
// 携带 dtm barrier 上下文（真实 saga 调用）时：经 barrier 幂等执行，业务规则失败
// （产品不存在/库存不足）返回 codes.Aborted 触发补偿，基础设施失败原样透传触发
// dtm 退避重试；"已扣减"由模型锁表幂等 + barrier 双保险按成功处理。
// 无 barrier 上下文（直连 RPC，兼容调用）时：保持既有语义（业务失败随 body 返回）。
func (l *DecreaseInventoryLogic) DecreaseInventory(in *inventory.InventoryReq) (*inventory.InventoryResp, error) {

	var res = new(inventory.InventoryResp)

	//将id和数量分别存入数组
	productId := make([]int32, len(in.Items))
	quantity := make([]int32, len(in.Items))
	for i, item := range in.Items {
		productId[i] = item.ProductId
		quantity[i] = item.Quantity
	}

	if bar, berr := dtmgrpc.BarrierFromGrpc(l.ctx); berr == nil {
		bar.DBType = "postgres"
		err := bar.CallWithDB(l.svcCtx.DtmDB, func(tx *sql.Tx) error {
			return l.svcCtx.InventoryModel.BatchDecreaseInventoryAtomWithSession(
				l.ctx, sqlx.NewSessionFromTx(tx), productId, quantity, int64(in.UserId), in.PreOrderId)
		})
		switch {
		case err == nil:
			return res, nil
		case errors.Is(err, sqlx.ErrNotFound):
			l.Logger.Infow("product not in inventory", logx.Field("product_id", productId))
			return nil, status.Error(codes.Aborted, code.ProductNotFoundInventoryMsg)
		case errors.Is(err, biz.ErrInventoryNotEnough):
			l.Logger.Infow("product inventory not enough", logx.Field("product_id", productId))
			return nil, status.Error(codes.Aborted, code.InventoryNotEnoughMsg)
		default:
			// 基础设施失败：未知结果 → dtm 退避重试（不触发补偿）
			l.Logger.Errorw("product inventory decrease failed", logx.Field("product_id", productId), logx.Field("err", err))
			return nil, err
		}
	}

	// 事务
	err := l.svcCtx.InventoryModel.BatchDecreaseInventoryAtom(l.ctx, productId, quantity, int64(in.UserId), in.PreOrderId)

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

	return res, nil
}
