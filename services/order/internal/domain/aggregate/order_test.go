package aggregate

import (
	"testing"
	"time"

	"github.com/falconfan123/Go-mall/services/order/internal/domain/entity"
	"github.com/stretchr/testify/require"
)

func TestOrderAggregateLifecycle(t *testing.T) {
	t.Parallel()

	order := NewOrderAggregate("ord-1", "pre-1", 1, "coupon-1", 1000, 100, 900, 30)
	require.NoError(t, order.AddItem(1001, 2, 300, "product", "desc"))
	order.SetAddress(10, "fan", "13800138000", "Zhejiang", "Hangzhou", "No.1")

	require.NoError(t, order.Pay(1, "tx-1"))
	require.Equal(t, entity.OrderStatus_ORDER_STATUS_PAID, order.GetStatus())
	require.Equal(t, entity.PaymentStatus_PAYMENT_STATUS_PAID, order.GetPaymentStatus())

	require.NoError(t, order.Ship())
	require.NoError(t, order.Complete())
	require.Equal(t, entity.OrderStatus_ORDER_STATUS_COMPLETED, order.GetStatus())
}

func TestOrderAggregateCancelAndRefundRules(t *testing.T) {
	t.Parallel()

	order := NewOrderAggregate("ord-2", "pre-2", 1, "", 1000, 0, 1000, 30)
	require.NoError(t, order.Cancel("user canceled"))
	require.Equal(t, entity.OrderStatusCanceled, order.GetStatus())

	paidOrder := NewOrderAggregate("ord-3", "pre-3", 1, "", 1000, 0, 1000, 30)
	require.NoError(t, paidOrder.Pay(1, "tx-2"))
	require.NoError(t, paidOrder.Refund())
	require.Equal(t, entity.PaymentStatus_PAYMENT_STATUS_REFUNDed, paidOrder.GetPaymentStatus())

	notPaidOrder := NewOrderAggregate("ord-4", "pre-4", 1, "", 1000, 0, 1000, 30)
	require.ErrorIs(t, notPaidOrder.Ship(), entity.ErrOrderNotPaid)
	require.ErrorIs(t, notPaidOrder.Refund(), entity.ErrOrderNotPaid)
}

func TestOrderAggregateLoadAndGetters(t *testing.T) {
	t.Parallel()

	order := NewOrderAggregate("ord-5", "pre-5", 9, "coupon-9", 1200, 100, 1100, 30)
	loaded := LoadOrder(order.GetOrder())

	require.Same(t, order.GetOrder(), loaded.GetOrder())
	require.Equal(t, "ord-5", loaded.GetOrderID())
	require.Equal(t, "pre-5", loaded.GetPreOrderID())
	require.Equal(t, int64(9), loaded.GetUserID())
	require.Equal(t, int64(1100), loaded.GetTotalAmount())
}

func TestOrderAggregateIsExpired(t *testing.T) {
	t.Parallel()

	order := NewOrderAggregate("ord-6", "pre-6", 1, "", 1000, 0, 1000, 30)
	order.GetOrder().ExpireTime = time.Now().Add(-time.Minute)
	require.True(t, order.IsExpired())
}
