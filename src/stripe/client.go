package stripe

import (
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"os"
)

func init() {
	stripe.Key = os.Getenv("STRIPE_PRIVATE_KEY")
}

type Payment struct {
	ClientSecret string `json:"client_secret" form:"client_secret" query:"client_secret"`
	Amount       int64  `json:"amount" form:"amount" query:"amount"`
	Currency     string `json:"currency" form:"currency" query:"currency"`
	//USDAmount int64 // TODO fetch the USD equivalent to allow granting premium on other currencies
}

func CreatePayment(amount int64, currency string, description string) (*Payment, error) {
	params := &stripe.PaymentIntentParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(currency),
		Description: stripe.String(description),
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
