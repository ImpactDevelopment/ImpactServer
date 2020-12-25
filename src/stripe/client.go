package stripe

import (
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/client"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"os"
)

var sc *client.API
var clientID string

func init() {
	clientID = os.Getenv("STRIPE_CLIENT_ID")
	stripe.Key = os.Getenv("STRIPE_CLIENT_SECRET")
	//apiBase := os.Getenv("STRIPE_API_BASE")
}

type Payment struct {
	ClientSecret string
	Amount       int64
	Currency     string
	//USDAmount int64 // TODO fetch the USD equivalent to allow granting premium on other currencies
}

func CreatePayment(amount int64, currency string) (*Payment, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
	}
	intent, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	return &Payment{
		ClientSecret: intent.ClientSecret,
		Amount:       intent.Amount,
		Currency:     intent.Currency,
	}, nil

}
