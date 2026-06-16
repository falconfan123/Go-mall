package entity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaymentEntityLifecycle(t *testing.T) {
	t.Parallel()

	payment := NewPayment("pay-1", "pre-1", "ord-1", 1, 1000, 900, PaymentMethodStripe, "url", 30)
	require.True(t, payment.CanPay())

	require.NoError(t, payment.Pay("tx-1"))
	require.Equal(t, PaymentStatusPaid, payment.Status)

	require.NoError(t, payment.Refund())
	require.Equal(t, PaymentStatusRefunded, payment.Status)
}

func TestPaymentMethodFromString(t *testing.T) {
	t.Parallel()

	require.Equal(t, PaymentMethodStripe, PaymentMethodFromString("stripe"))
	require.Equal(t, PaymentMethodUnknown, PaymentMethodFromString("unknown"))
	require.Equal(t, "stripe", PaymentMethodStripe.String())
}
