package stripe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/zeromicro/go-zero/core/logx"
)

type StripeProcessor struct {
	apiKey        string
	successURL    string
	cancelURL     string
	webhookSecret string
	client        session.Client
}

func NewStripeProcessor(cfg config.StripeConfig) *StripeProcessor {
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("STRIPE_API_KEY"))
	}
	if apiKey == "" {
		logx.Info("Stripe API key is empty, Stripe payment will not work")
	}
	stripe.Key = apiKey

	// Increase default timeout to 8 seconds for better reliability
	requestTimeout := time.Duration(cfg.RequestTimeoutMs) * time.Millisecond
	if requestTimeout <= 0 {
		requestTimeout = 8 * time.Second
	}

	// Enable network retries by default (2 retries)
	maxNetworkRetries := cfg.MaxNetworkRetries
	if maxNetworkRetries < 0 {
		maxNetworkRetries = 2
	}

	backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		HTTPClient: &http.Client{
			Timeout: requestTimeout,
		},
		MaxNetworkRetries: stripe.Int64(maxNetworkRetries),
	})

	webhookSecret := strings.TrimSpace(cfg.WebhookSecret)
	if webhookSecret == "" {
		webhookSecret = strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	}

	// Resolve URLs from config or environment
	successURL := strings.TrimSpace(cfg.SuccessURL)
	if successURL == "" {
		successURL = strings.TrimSpace(os.Getenv("STRIPE_SUCCESS_URL"))
	}
	cancelURL := strings.TrimSpace(cfg.CancelURL)
	if cancelURL == "" {
		cancelURL = strings.TrimSpace(os.Getenv("STRIPE_CANCEL_URL"))
	}

	return &StripeProcessor{
		apiKey:        apiKey,
		successURL:    successURL,
		cancelURL:     cancelURL,
		webhookSecret: webhookSecret,
		client:        session.Client{B: backend, Key: apiKey},
	}
}

// CreatePaymentLink creates a Stripe Checkout payment link
func (s *StripeProcessor) CreatePaymentLink(ctx context.Context, orderID string, amount int64, items []*PaymentItem, metadata map[string]string) (string, error) {
	if strings.TrimSpace(s.apiKey) == "" || strings.Contains(strings.ToLower(s.apiKey), "placeholder") {
		return "", errors.New("Stripe API key is not configured")
	}

	var lineItems []*stripe.CheckoutSessionLineItemParams

	// If no items provided, create a default line item for the order
	if len(items) == 0 {
		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String("cny"),
				UnitAmount: stripe.Int64(amount),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String("Order Payment"),
				},
			},
			Quantity: stripe.Int64(1),
		})
	} else {
		for _, item := range items {
			lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String("cny"),
					UnitAmount: stripe.Int64(item.Price),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(item.Name),
						Description: stripe.String(item.Description),
					},
				},
				Quantity: stripe.Int64(item.Quantity),
			})
		}
	}

	// Marshal items to JSON for metadata
	itemsJSON, _ := json.Marshal(items)
	sessionMetadata := map[string]string{
		"order_id": orderID,
		"items":    string(itemsJSON),
	}
	for key, value := range metadata {
		if value == "" {
			continue
		}
		sessionMetadata[key] = value
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:   sessionMetadata,
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(buildReturnURL(s.successURL, sessionMetadata)),
		CancelURL:  stripe.String(buildReturnURL(s.cancelURL, sessionMetadata)),
	}
	params.Context = ctx

	result, err := s.client.New(params)
	if err != nil {
		logx.Errorw("Failed to create Stripe payment link", logx.Field("error", err))
		return "", err
	}

	// Log the created URLs for debugging
	logx.Infow("Created Stripe payment link",
		logx.Field("order_id", orderID),
		logx.Field("success_url", buildReturnURL(s.successURL, sessionMetadata)),
		logx.Field("cancel_url", buildReturnURL(s.cancelURL, sessionMetadata)),
		logx.Field("url", result.URL))
	return result.URL, nil
}

// GetWebhookSecret returns the webhook secret
func (s *StripeProcessor) GetWebhookSecret() string {
	return s.webhookSecret
}

func buildReturnURL(baseURL string, metadata map[string]string) string {
	if baseURL == "" {
		return ""
	}

	query := fmt.Sprintf("?order_id=%s", metadata["order_id"])
	if paymentID := metadata["payment_id"]; paymentID != "" {
		query += fmt.Sprintf("&payment_id=%s", paymentID)
	}
	return baseURL + query
}

// PaymentItem represents an item in the payment
type PaymentItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	Price       int64  `json:"price"`
}

func UserFacingCreatePaymentError(err error) string {
	if err == nil {
		return "Stripe 支付创建失败，请稍后重试"
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "Stripe 请求超时，请稍后重试"
	case strings.Contains(message, "api key is not configured"):
		return "Stripe API key 未配置"
	case strings.Contains(message, "lookup api.stripe.com"),
		strings.Contains(message, "no such host"),
		strings.Contains(message, "dial tcp"),
		strings.Contains(message, "i/o timeout"):
		return "Stripe 服务暂时不可用，请检查当前开发机到 api.stripe.com 的网络和 DNS 后重试"
	default:
		return "Stripe 支付创建失败，请稍后重试"
	}
}
