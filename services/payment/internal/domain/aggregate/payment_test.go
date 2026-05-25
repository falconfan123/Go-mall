package aggregate

import (
	"testing"

	"github.com/falconfan123/Go-mall/services/payment/internal/domain/entity"
	"github.com/stretchr/testify/require"
)

func TestPaymentAggregateLifecycle(t *testing.T) {
	t.Parallel()

	payment := NewPaymentAggregate("pay-1", "pre-1", "ord-1", 1, 1000, 900, entity.PaymentMethodAlipay, "url", 30)
	require.True(t, payment.CanPay())
	require.NoError(t, payment.Pay("tx-1"))
	require.Equal(t, entity.PaymentStatusPaid, payment.GetStatus())

	require.NoError(t, payment.Refund())
	require.Equal(t, entity.PaymentStatusRefunded, payment.GetStatus())
}

func TestPaymentAggregateCancelAndInvalidTransitions(t *testing.T) {
	t.Parallel()

	payment := NewPaymentAggregate("pay-2", "pre-2", "ord-2", 1, 1000, 900, entity.PaymentMethodWechat, "url", 30)
	require.NoError(t, payment.Cancel())
	require.Equal(t, entity.PaymentStatusFailed, payment.GetStatus())
	require.ErrorIs(t, payment.Pay("tx-1"), entity.ErrInvalidPaymentStatus)
}
