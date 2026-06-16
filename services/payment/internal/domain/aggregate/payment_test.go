package aggregate

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
	"github.com/stretchr/testify/require"
)

func TestPaymentAggregateLifecycle(t *testing.T) {
	t.Parallel()

	payment := NewPaymentAggregate("pay-1", "pre-1", "ord-1", 1, 1000, 900, entity.PaymentMethodStripe, "url", 30)
	require.True(t, payment.CanPay())
	require.NoError(t, payment.Pay("tx-1"))
	require.Equal(t, entity.PaymentStatusPaid, payment.GetStatus())

	require.NoError(t, payment.Refund())
	require.Equal(t, entity.PaymentStatusRefunded, payment.GetStatus())
}

func TestPaymentAggregateCancelAndInvalidTransitions(t *testing.T) {
	t.Parallel()

	payment := NewPaymentAggregate("pay-2", "pre-2", "ord-2", 1, 1000, 900, entity.PaymentMethodStripe, "url", 30)
	require.NoError(t, payment.Cancel())
	require.Equal(t, entity.PaymentStatusFailed, payment.GetStatus())
	require.ErrorIs(t, payment.Pay("tx-1"), entity.ErrInvalidPaymentStatus)
}

func TestPaymentAggregateLoadAndGetters(t *testing.T) {
	t.Parallel()

	payment := NewPaymentAggregate("pay-3", "pre-3", "ord-3", 7, 1000, 900, entity.PaymentMethodStripe, "pay-url", 30)
	loaded := LoadPayment(payment.GetPayment())

	require.Same(t, payment.GetPayment(), loaded.GetPayment())
	require.Equal(t, "pay-3", loaded.GetPaymentID())
	require.Equal(t, "ord-3", loaded.GetOrderID())
	require.Equal(t, "pre-3", loaded.GetPreOrderID())
	require.Equal(t, int64(7), loaded.GetUserID())
	require.Equal(t, int64(900), loaded.GetAmount())
	require.Equal(t, entity.PaymentMethodStripe, loaded.GetPaymentMethod())
	require.Equal(t, "pay-url", loaded.GetPayURL())
	require.Greater(t, loaded.GetExpireTime(), int64(0))
}

func TestPaymentAggregateExpiration(t *testing.T) {
	t.Parallel()

	payment := NewPaymentAggregate("pay-4", "pre-4", "ord-4", 1, 1000, 900, entity.PaymentMethodStripe, "url", -1)
	require.True(t, payment.IsExpired())
	require.False(t, payment.CanPay())
}
