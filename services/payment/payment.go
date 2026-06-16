package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/code"
	ordertypes "github.com/falconfan123/Go-mall/common/types/order"
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

	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

type PaymentService struct {
	ctx *svc.ServiceContext
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

	if payment.PaymentStatus(paymentRecord.Status) != payment.PaymentStatus_PAYMENT_STATUS_PAID {
		paymentRecord.TransactionId = sql.NullString{String: transactionID, Valid: transactionID != ""}
		paymentRecord.PaidAmount = sql.NullInt64{Int64: paidAmount, Valid: true}
		paymentRecord.PaidAt = sql.NullInt64{Int64: paidAt, Valid: true}
		paymentRecord.Status = int64(payment.PaymentStatus_PAYMENT_STATUS_PAID)
		paymentRecord.UpdatedAt = time.Now()
		if err := s.ctx.PaymentModel.Update(ctx, paymentRecord); err != nil {
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

	if ordertypes.OrderStatus(orderState.Order.OrderStatus) == ordertypes.OrderStatusPaid &&
		ordertypes.PaymentStatus(orderState.Order.PaymentStatus) == ordertypes.PaymentStatusPaid {
		return nil
	}

	orderRes, err := s.ctx.OrderRpc.UpdateOrder2PaymentSuccess(ctx, &order.UpdateOrder2PaymentSuccessRequest{
		OrderId: orderID,
		PaymentResult: &order.PaymentResult{
			TransactionId: transactionID,
			PaidAmount:    paidAmount,
			PaidAt:        paidAt,
		},
		UserId: int32(userID),
	})
	if err != nil {
		return err
	}
	if orderRes.StatusCode != code.Success {
		return fmt.Errorf("update order status failed: %s", orderRes.StatusMsg)
	}
	return nil
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
