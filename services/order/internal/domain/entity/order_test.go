package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderEntityLifecycle(t *testing.T) {
	t.Parallel()

	order := NewOrder("ord-1", "pre-1", 1, "", 1000, 100, 900, 30)
	order.AddItem(NewOrderItem("ord-1", 1001, 2, 300, "product", "desc"))
	order.SetAddress(NewOrderAddress("ord-1", 10, "fan", "13800138000", "Zhejiang", "Hangzhou", "No.1"))

	require.NoError(t, order.Pay(1, "tx-1"))
	require.NoError(t, order.Ship())
	require.NoError(t, order.Complete())
	require.Equal(t, OrderStatus_ORDER_STATUS_COMPLETED, order.OrderStatus)
}

func TestOrderEntityCancelAndRefund(t *testing.T) {
	t.Parallel()

	order := NewOrder("ord-2", "pre-2", 1, "", 1000, 0, 1000, 30)
	require.NoError(t, order.Cancel("cancel"))
	require.Equal(t, OrderStatusCanceled, order.OrderStatus)

	paid := NewOrder("ord-3", "pre-3", 1, "", 1000, 0, 1000, 30)
	require.NoError(t, paid.Pay(1, "tx-2"))
	require.NoError(t, paid.Refund())
	require.Equal(t, PaymentStatus_PAYMENT_STATUS_REFUNDed, paid.PaymentStatus)
}
