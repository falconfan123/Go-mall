package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
	inventorypb "github.com/falconfan123/Go-mall/services/inventory/pb"
	order "github.com/falconfan123/Go-mall/services/order/pb"
	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/falconfan123/Go-mall/services/payment/internal/server"
	"github.com/falconfan123/Go-mall/services/payment/internal/svc"
	payment "github.com/falconfan123/Go-mall/services/payment/pb"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/payment.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		payment.RegisterPaymentServer(grpcServer, server.NewPaymentServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	paymentSvc := NewPaymentService(ctx)
	paymentSvc.startHTTPServer()
	paymentSvc.startFallbackWorker()

	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

type PaymentService struct {
	ctx *svc.ServiceContext
	// recon 对账状态（进程内存）：当日已跑则跳过
	reconMu      sync.Mutex
	lastReconDay string
	// stripeLister Stripe 会话拉取抽象（nil=默认实现；测试注入 fake，Rule 6）
	stripeLister func(ctx context.Context, startSec, endSec int64) ([]stripeSessionData, error)
}

func NewPaymentService(ctx *svc.ServiceContext) *PaymentService {
	return &PaymentService{ctx: ctx}
}

// handleStripeWebhook 处理 Stripe Webhook 回调
func (s *PaymentService) handleStripeWebhook(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if recovered := recover(); recovered != nil {
			logx.Errorw("Stripe webhook handler panicked",
				logx.Field("panic", recovered),
				logx.Field("stack", string(debug.Stack())))
			http.Error(writer, "stripe webhook panic", http.StatusInternalServerError)
		}
	}()

	logx.Info("Got webhook from Stripe")

	const MaxBodyBytes = int64(65536)
	request.Body = http.MaxBytesReader(writer, request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		logx.Infow("Error reading request body", logx.Field("err", err))
		http.Error(writer, err.Error(), http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, request.Header.Get("Stripe-Signature"),
		s.ctx.StripeProcessor.GetWebhookSecret())
	if err != nil {
		logx.Infow("Error verifying webhook signature", logx.Field("err", err))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	// 处理支付成功事件
	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logx.Infow("Error unmarshaling event", logx.Field("err", err))
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			logx.Infow("Payment for checkout session completed", logx.Field("session_id", session.ID))

			orderID := session.Metadata["order_id"]
			paymentID := session.Metadata["payment_id"]
			userID, err := strconv.ParseUint(session.Metadata["user_id"], 10, 32)
			if err != nil || orderID == "" || paymentID == "" || userID == 0 {
				logx.Errorw("Stripe session metadata is incomplete",
					logx.Field("order_id", orderID),
					logx.Field("payment_id", paymentID),
					logx.Field("user_id", session.Metadata["user_id"]),
					logx.Field("err", err))
				http.Error(writer, "invalid stripe metadata", http.StatusBadRequest)
				return
			}

			paidAmount, err := parseStripePaidAmount(session)
			if err != nil {
				logx.Errorw("Failed to parse Stripe paid amount", logx.Field("err", err))
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}

			transactionID := normalizeStripeTransactionID(session.ID)
			if session.PaymentIntent != nil && session.PaymentIntent.ID != "" {
				transactionID = normalizeStripeTransactionID(session.PaymentIntent.ID)
			}

			if err := s.processStripePaymentSuccess(
				request.Context(),
				orderID,
				paymentID,
				uint32(userID),
				transactionID,
				paidAmount,
				time.Now().Unix(),
			); err != nil {
				logx.Errorw("Failed to process Stripe payment success", logx.Field("err", err))
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	writer.WriteHeader(http.StatusOK)
}

func parseStripePaidAmount(session stripe.CheckoutSession) (int64, error) {
	if raw := session.Metadata["pay_amount"]; raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	if session.AmountTotal > 0 {
		return session.AmountTotal, nil
	}
	return 0, fmt.Errorf("missing pay_amount metadata for session %s", session.ID)
}

func normalizeStripeTransactionID(value string) string {
	const maxLen = 64
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen]
}

func (s *PaymentService) processStripePaymentSuccess(
	ctx context.Context,
	orderID string,
	paymentID string,
	userID uint32,
	transactionID string,
	paidAmount int64,
	paidAt int64,
) error {
	paymentRecord, err := s.ctx.PaymentModel.FindOne(ctx, paymentID)
	if err != nil {
		return err
	}

	// 结算输入快照预取（事务外；outbox payload 与结算输入共用，design 决策 4）：
	// 金额/优惠券来自订单行，items 来自订单项（Initiate 与 Ensure 共用同一集合）。
	detail, err := s.ctx.OrderRpc.GetOrder(ctx, &order.GetOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
	if err != nil {
		return err
	}
	if detail.StatusCode != code.Success {
		return fmt.Errorf("query order detail failed: %s", detail.StatusMsg)
	}
	items := make([]*inventorypb.InventoryReq_Items, 0, len(detail.Items))
	for _, it := range detail.Items {
		items = append(items, &inventorypb.InventoryReq_Items{
			ProductId: int32(it.ProductId),
			Quantity:  int32(it.Quantity),
		})
	}
	input := &svc.SettlementInput{
		OrderID:        orderID,
		UserID:         int32(userID),
		PreOrderID:     detail.Order.PreOrderId,
		TransactionID:  transactionID,
		PaidAmount:     paidAmount,
		PaidAt:         paidAt,
		Items:          items,
		CouponID:       detail.Order.CouponId,
		DiscountAmount: detail.Order.DiscountAmount,
		OriginAmount:   detail.Order.OriginalAmount,
	}

	if payment.PaymentStatus(paymentRecord.Status) != payment.PaymentStatus_PAYMENT_STATUS_PAID {
		// 两写事务（design 决策 4）：支付单翻转与结算待办同事务落库——
		// "翻转了但未触发编排"只可能存在于 outbox pending 行，由 relay 必达接力。
		// 条件更新带状态机门槛（status <> PAID），rows==1（本次翻转成功）才写待办。
		payloadBytes, err := json.Marshal(input)
		if err != nil {
			return err
		}
		err = s.ctx.Model.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
			rows, err := s.ctx.PaymentModel.UpdateStatusToPaidWithSession(
				ctx, session, paymentID, transactionID, paidAmount, paidAt)
			if err != nil {
				return err
			}
			if rows == 1 {
				return s.ctx.OutboxModel.InsertPendingWithSession(
					ctx, session, orderID, paymentID, int64(userID), string(payloadBytes))
			}
			return nil
		})
		if err != nil {
			logx.Errorw("payment flip transaction failed (rolled back, stripe will retry)",
				logx.Field("err", err), logx.Field("order_id", orderID), logx.Field("payment_id", paymentID))
			return err
		}
	}

	orderState, err := s.ctx.OrderRpc.GetOrder2Payment(ctx, &order.GetOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
	if err != nil {
		return err
	}
	if orderState.StatusCode != code.Success {
		return fmt.Errorf("query order state failed: %s", orderState.StatusMsg)
	}

	// 结算统一走 Saga（design 决策 1/2）：本地支付单已 PAID，此后由 dtm 编排
	// 订单/库存/优惠券三分支；快路径（订单已 Paid）与正常路径共用同一确定性 gid，
	// 重复发起由 dtm 幂等收敛（Ensure），outbox relay 为系统性必达保证（本 change）。
	orderPaid := ordertypes.OrderStatus(orderState.Order.OrderStatus) == ordertypes.OrderStatusPaid &&
		ordertypes.PaymentStatus(orderState.Order.PaymentStatus) == ordertypes.PaymentStatusPaid

	if orderPaid {
		// 快路径：订单已支付 → 幂等补建（saga in-flight/已完成均收敛为成功）
		return s.ctx.SettlementSaga.Ensure(ctx, input)
	}
	// 正常路径：发起结算 saga（发起失败返回错误，webhook 500 触发 Stripe 重试）
	return s.ctx.SettlementSaga.Initiate(ctx, input)
}

// 封装HTTP服务启动
func (s *PaymentService) startHTTPServer() {
	// 注册 Stripe Webhook
	if s.ctx.Config.Stripe.WebhookPort > 0 {
		http.HandleFunc("/stripe/webhook", s.handleStripeWebhook)
		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", s.ctx.Config.Stripe.WebhookPort), nil); err != nil {
				logx.Errorw("Stripe webhook server error", logx.Field("err", err))
			}
		}()
	}
}
