package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/falconfan123/Go-mall/common/types/coupons"
	"github.com/falconfan123/Go-mall/dal/model/coupons/coupon_usage"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/falconfan123/Go-mall/services/coupons/internal/svc"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
)

type UseCouponLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUseCouponLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UseCouponLogic {
	return &UseCouponLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UseCoupon 使用优惠券（支付成功确认，Saga branch-03）。
// 携带 dtm barrier 上下文时经 barrier 幂等执行；无 barrier 上下文保持既有事务语义。
// "已使用"类幂等响应按成功处理（specs 业务幂等兜底契约）；业务规则失败（券不存在/
// 状态无效）在 barrier 路径返回 codes.Aborted 触发补偿，直连路径保持既有 body 语义。
func (l *UseCouponLogic) UseCoupon(in *couponspb.UseCouponReq) (*couponspb.EmptyResp, error) {
	res := &couponspb.EmptyResp{}

	if bar, berr := dtmgrpc.BarrierFromGrpc(l.ctx); berr == nil {
		bar.DBType = "postgres"
		err := bar.CallWithDB(l.svcCtx.DtmDB, func(tx *sql.Tx) error {
			if err := l.useCouponTx(l.ctx, sqlx.NewSessionFromTx(tx), in, res); err != nil {
				return err
			}
			if res.StatusCode != code.Success {
				return status.Error(codes.Aborted, res.StatusMsg)
			}
			return nil
		})
		if err != nil {
			l.Logger.Errorw("saga use coupon branch failed", logx.Field("err", err),
				logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
				logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
			return nil, err
		}
		return res, nil
	}

	// 非 saga 直连路径：既有事务语义
	if err := l.svcCtx.Model.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		return l.useCouponTx(ctx, session, in, res)
	}); err != nil {
		l.Logger.Errorw("insert coupon usage error", logx.Field("err", err),
			logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
			logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
		return nil, err
	}

	if res.StatusCode != code.Success {
		return res, nil
	}
	l.Logger.Infow("use coupon success", logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
		logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
	return res, nil
}

// useCouponTx 在给定事务会话内执行使用优惠券业务（barrier 路径与直连路径共用）。
// 返回 error 仅代表基础设施失败（dtm 退避重试）；业务规则失败通过 res.StatusCode 表达。
func (l *UseCouponLogic) useCouponTx(ctx context.Context, session sqlx.Session, in *couponspb.UseCouponReq, res *couponspb.EmptyResp) error {
	// --------------- check ---------------
	// 判断用户优惠券状态是否已经是已使用，支付成功后，修改优惠券状态为已使用
	st, err := l.svcCtx.UserCouponsModel.WithSession(session).GetStatusByUserIdCouponId(ctx, in.UserId, in.CouponId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Logger.Infow("user coupon not exist", logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId))
			res.StatusCode = code.CouponsNotExist
			res.StatusMsg = code.CouponsNotExistMsg
			return nil
		}
		l.Logger.Errorw("get user coupon status error", logx.Field("err", err),
			logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
			logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
		return err
	}
	// 2. 状态校验（幂等容忍：已使用按成功处理，specs 业务幂等兜底契约）
	if coupons.CouponStatus(st.Status) == coupons.CouponStatusUsed {
		l.Logger.Infow("coupon already used, tolerate as success", logx.Field("user_id", in.UserId),
			logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId))
		return nil
	}
	if coupons.CouponStatus(st.Status) != coupons.CouponStatusLocked {
		res.StatusCode = code.CouponStatusInvalid
		res.StatusMsg = code.CouponStatusInvalidMsg
		l.Logger.Infow("coupon status invalid", logx.Field("user_id", in.UserId),
			logx.Field("coupon_id", in.CouponId), logx.Field("order_id", in.OrderId),
			logx.Field("pre_order_id", in.PreOrderId), logx.Field("status", st.Status))
		return nil
	}

	// --------------- query ---------------
	tp, err := l.svcCtx.CouponsModel.GetCouponTypeByID(ctx, session, in.CouponId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			l.Logger.Infow("coupon not exist", logx.Field("coupon_id", in.CouponId))
			res.StatusCode = code.CouponsNotExist
			res.StatusMsg = code.CouponsNotExistMsg
			return nil
		}
		l.Logger.Errorw("get coupon error", logx.Field("err", err),
			logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
			logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
		return err
	}

	// --------------- update and record ---------------
	// update
	if err := l.svcCtx.UserCouponsModel.WithSession(session).UpdateStatusOrderById(ctx,
		in.OrderId, int(st.ID), coupons.CouponStatusUsed); err != nil {
		l.Logger.Errorw("update user coupon status error", logx.Field("err", err),
			logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
			logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
		return err
	}
	// record
	if _, err := l.svcCtx.CouponUsageModel.WithSession(session).Insert(ctx, &coupon_usage.CouponUsage{
		OrderId:        in.OrderId,
		CouponId:       in.CouponId,
		UserId:         uint64(in.UserId),
		CouponType:     tp,
		DiscountAmount: in.DiscountAmount,
		OriginValue:    in.OriginAmount,
		AppliedAt:      time.Now(),
	}); err != nil {
		l.Logger.Errorw("insert coupon usage error", logx.Field("err", err),
			logx.Field("user_id", in.UserId), logx.Field("coupon_id", in.CouponId),
			logx.Field("order_id", in.OrderId), logx.Field("pre_order_id", in.PreOrderId))
		return err
	}
	return nil
}
