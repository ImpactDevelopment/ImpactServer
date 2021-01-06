package stripe

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"github.com/stripe/stripe-go/v71/webhook"
	"net/http"
	"os"
)

var PublicKey string
var webhookSecret string

func init() {
	PublicKey = os.Getenv("STRIPE_PUBLIC_KEY")
	stripe.Key = os.Getenv("STRIPE_PRIVATE_KEY")
	webhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
}

type Payment struct {
	PaymentIntent *stripe.PaymentIntent `json:"-" form:"-" query:"-" xml:"-"`
	ClientSecret  string                `json:"client_secret" form:"client_secret" query:"client_secret"`
	Amount        int64                 `json:"amount" form:"amount" query:"amount"`
	Currency      string                `json:"currency" form:"currency" query:"currency"`
	Email         string                `json:"email" form:"email" query:"email"`
	//USDAmount int64 // TODO fetch the USD equivalent to allow granting premium on other currencies
}

type WebhookEvent struct {
	stripe.Event
}

type CurrencyInfo struct {
	Amount      int64  `json:"premium_amount" form:"premium_amount" query:"premium_amount"`
	DisplayName string `json:"display_name" form:"display_name" query:"display_name"`
	Symbol      string `json:"symbol" form:"symbol" query:"symbol"`
}

// Supported currencies and the respective amount required for premium perks
var stripeCurrencyMap = map[string]CurrencyInfo{
	"usd": {
		Amount:      500,
		DisplayName: "$ USD",
		Symbol:      "$",
	},
	"eur": {
		Amount:      500,
		DisplayName: "€ EUR",
		Symbol:      "€",
	},
	"gbp": {
		Amount:      500,
		DisplayName: "£ GBP",
		Symbol:      "£",
	},
}

func GetWebhookEvent(payload []byte, signature string) (*WebhookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "incorrect signature").SetInternal(err)
	}
	return &WebhookEvent{event}, nil
}

func CreatePayment(amount int64, currency string, description string, email string) (*Payment, error) {
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(currency),
		Description: stripe.String(description),
	}
	if email != "" {
		params.ReceiptEmail = stripe.String(email)
	}
	payment, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	return &Payment{
		PaymentIntent: payment,
		ClientSecret:  payment.ClientSecret,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Email:         payment.ReceiptEmail,
	}, nil
}

func GetPayment(id string) (*Payment, error) {
	payment, err := paymentintent.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return &Payment{
		PaymentIntent: payment,
		ClientSecret:  payment.ClientSecret,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Email:         payment.ReceiptEmail,
	}, nil
}

func GetCurrencySymbol(currency string) string {
	if it, ok := stripeCurrencyMap[currency]; ok {
		return it.Symbol
	}
	return "¤"
}

func GetCurrencyInfo(currency string) (*CurrencyInfo, error) {
	if it, ok := stripeCurrencyMap[currency]; ok {
		return &it, nil
	}
	return nil, errors.New("invalid or unsupported currency \"" + currency + "\"")
}

func GetCurrencyMap() map[string]CurrencyInfo {
	return stripeCurrencyMap
}
