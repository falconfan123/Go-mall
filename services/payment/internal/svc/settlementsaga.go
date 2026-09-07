package svc

import (
	"context"
	"fmt"
	"regexp"

	"github.com/dtm-labs/dtm/client/dtmgrpc"
	couponspb "github.com/falconfan123/Go-mall/services/coupons/pb"
	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	orderpb "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

// settlementGidPattern 从 dtm 重复提交错误中解析既有 saga 状态
// （spike 实测：终态重复提交返回 `current status '<status>', cannot sumbmit. FAILURE`）
var settlementGidPattern = regexp.MustCompile(`current status '(\w+)'`)

// SettlementSagaAPI saga 发起/补建的抽象（便于 webhook 单测注入 mock，Rule 6）。
type SettlementSagaAPI interface {
	Initiate(ctx context.Context, in *SettlementInput) error
	Ensure(ctx context.Context, in *SettlementInput) error
}

// SettlementSaga 结算 saga 发起/幂等补建（design 决策 1）。
// gid = settle:{orderId} 确定性构造：重复发起撞同一 gid，由 dtm 幂等收敛。
// pb 转换内聚本封装，main 包（payment.go）不得直接构造分支 payload（Rule 4）。
type SettlementSaga struct {
	cfg config.DtmConfig
	log logx.Logger
}

func NewSettlementSaga(cfg config.DtmConfig, log logx.Logger) *SettlementSaga {
	return &SettlementSaga{cfg: cfg, log: log}
}

// gid 确定性编排标识；GidSuffix 为空 = 原始 gid（settle:{orderId}），
// 墓碑重试时由扫描器构造 ":r2"/":r3" 后缀（design 决策 5 前置声明）。
func (s *SettlementSaga) gid(orderID, suffix string) string { return "settle:" + orderID + suffix }

// 分支 URL：dtm server（容器）回调分支服务的地址（host 前缀来自配置，dev 为 host.docker.internal）
func (s *SettlementSaga) url(port int, method string) string {
	return fmt.Sprintf("%s:%d/%s", s.cfg.BusiHost, port, method)
}

func (s *SettlementSaga) orderAction() string {
	return s.url(s.cfg.OrderPort, "order.OrderService/UpdateOrder2PaymentSuccess")
}

func (s *SettlementSaga) orderCompensate() string {
	return s.url(s.cfg.OrderPort, "order.OrderService/UpdateOrder2PaymentSuccessRollback")
}

func (s *SettlementSaga) inventoryAction() string {
	return s.url(s.cfg.InventoryPort, "inventory.Inventory/DecreaseInventory")
}

func (s *SettlementSaga) inventoryCompensate() string {
	return s.url(s.cfg.InventoryPort, "inventory.Inventory/ReturnInventory")
}

func (s *SettlementSaga) couponAction() string {
	return s.url(s.cfg.CouponsPort, "coupons.Coupons/UseCoupon")
}

func (s *SettlementSaga) couponRollback() string {
	return s.url(s.cfg.CouponsPort, "coupons.Coupons/UseCouponRollback")
}

// SettlementInput 发起结算所需的业务数据（由 webhook 从支付单 + 订单快照提取）；
// 亦作为 outbox payload 的序列化载体（encoding/json，内部存储格式）。
type SettlementInput struct {
	OrderID        string
	UserID         int32
	PreOrderID     string
	TransactionID  string
	PaidAmount     int64
	PaidAt         int64
	Items          []*inventorypb.InventoryReq_Items // 订单行（productId+quantity），扣/回补共用同一集合
	CouponID       string                            // 无券为空（跳过 branch-03）
	DiscountAmount int64
	OriginAmount   int64
	// GidSuffix 编排标识后缀（墓碑重试用，如 ":r2"；空 = 原始 gid）
	GidSuffix string
}

// Initiate 发起结算 saga（新结算）。dtm 保证分支自动重试 + 失败逆序补偿。
func (s *SettlementSaga) Initiate(ctx context.Context, in *SettlementInput) error {
	sg := dtmgrpc.NewSagaGrpc(s.cfg.Server, s.gid(in.OrderID, in.GidSuffix))

	// branch-01 订单：PendingPayment → Paid；补偿：Paid → PendingPayment
	sg.Add(s.orderAction(), s.orderCompensate(), &orderpb.UpdateOrder2PaymentSuccessRequest{
		OrderId: in.OrderID,
		UserId:  in.UserID,
		PaymentResult: &orderpb.PaymentResult{
			TransactionId: in.TransactionID,
			PaidAmount:    in.PaidAmount,
			PaidAt:        in.PaidAt,
		},
	})

	// branch-02 库存：真实扣减；补偿：等量回补（同一 items 集合，design 决策 2）
	invReq := &inventorypb.InventoryReq{
		Items:      in.Items,
		PreOrderId: in.PreOrderID,
		UserId:     in.UserID,
	}
	sg.Add(s.inventoryAction(), s.inventoryCompensate(), invReq)

	// branch-03 优惠券：使用；补偿：回滚（同类型 payload，见 UseCouponRollback）
	if in.CouponID != "" {
		couponReq := &couponspb.UseCouponReq{
			UserId:         in.UserID,
			PreOrderId:     in.PreOrderID,
			OrderId:        in.OrderID,
			CouponId:       in.CouponID,
			DiscountAmount: in.DiscountAmount,
			OriginAmount:   in.OriginAmount,
		}
		sg.Add(s.couponAction(), s.couponRollback(), couponReq)
	}

	return sg.Submit()
}

// Ensure 幂等补建（webhook 快路径）：对同一 gid 重复发起由 dtm 收敛——
//   - 新 gid → 创建 saga；
//   - in-flight（prepared/submitted）→ dtm 静默接受（spike 实测 Submit 返回 nil）；
//   - 终态 succeed → 结算已完成，视为成功（specs"结算发起幂等补建"）；
//   - 终态 failed → 补偿已回滚订单（不可能经快路径触达），返回错误 + error 告警
//     （人工介入，design 决策 1 附注的已接受残余风险姿态）。
func (s *SettlementSaga) Ensure(ctx context.Context, in *SettlementInput) error {
	err := s.Initiate(ctx, in)
	if err == nil {
		return nil
	}
	st := settlementGidPattern.FindStringSubmatch(err.Error())
	if len(st) == 2 {
		switch st[1] {
		case "succeed":
			logx.Infow("settlement saga already succeeded, ensure treated as success",
				logx.Field("order_id", in.OrderID), logx.Field("gid", s.gid(in.OrderID, in.GidSuffix)))
			return nil
		case "failed":
			logx.Errorw("settlement saga previously failed and compensated, manual intervention required",
				logx.Field("order_id", in.OrderID), logx.Field("gid", s.gid(in.OrderID, in.GidSuffix)))
			return err
		}
	}
	return err
}
