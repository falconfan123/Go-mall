package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/falconfan123/Go-mall/services/payment/internal/config"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/zeromicro/go-zero/core/logx"
)

type StripeProcessor struct {
	apiKey     string
	successURL string
	cancelURL  string
}

func NewStripeProcessor(cfg config.StripeConfig) *StripeProcessor {
	if cfg.APIKey == "" {
		logx.Info("Stripe API key is empty, Stripe payment will not work")
	}
	stripe.Key = cfg.APIKey
	return &StripeProcessor{
		apiKey:     cfg.APIKey,
		successURL: cfg.SuccessURL,
		cancelURL:  cfg.CancelURL,
	}
}

// CreatePaymentLink creates a Stripe Checkout payment link
func (s *StripeProcessor) CreatePaymentLink(ctx context.Context, orderID string, amount int64, items []*PaymentItem) (string, error) {
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
	metadata := map[string]string{
		"order_id": orderID,
		"items":    string(itemsJSON),
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:   metadata,
		LineItems:  lineItems,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("%s?order_id=%s", s.successURL, orderID)),
		CancelURL:  stripe.String(fmt.Sprintf("%s?order_id=%s", s.cancelURL, orderID)),
	}

	result, err := session.New(params)
	if err != nil {
		logx.Errorw("Failed to create Stripe payment link", logx.Field("error", err))
		return "", err
	}

	logx.Infow("Created Stripe payment link", logx.Field("order_id", orderID), logx.Field("url", result.URL))
	return result.URL, nil
}

// PaymentItem represents an item in the payment
type PaymentItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	Price       int64  `json:"price"`
}
