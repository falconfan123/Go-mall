package stripe

import (
	"context"
	"errors"
	"testing"
)

func TestUserFacingCreatePaymentError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
			want: "Stripe 请求超时，请稍后重试",
		},
		{
			name: "dns lookup failed",
			err:  errors.New(`Post "https://api.stripe.com/v1/checkout/sessions": dial tcp: lookup api.stripe.com: no such host`),
			want: "Stripe 服务暂时不可用，请检查当前开发机到 api.stripe.com 的网络和 DNS 后重试",
		},
		{
			name: "api key missing",
			err:  errors.New("Stripe API key is not configured"),
			want: "Stripe API key 未配置",
		},
		{
			name: "fallback",
			err:  errors.New("unexpected stripe error"),
			want: "Stripe 支付创建失败，请稍后重试",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := UserFacingCreatePaymentError(tt.err); got != tt.want {
				t.Fatalf("UserFacingCreatePaymentError() = %q, want %q", got, tt.want)
			}
		})
	}
}
