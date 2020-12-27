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
	PaymentIntent *stripe.PaymentIntent `json:"-" form:"-" query:"-" xml:"-"`
	ClientSecret  string                `json:"client_secret" form:"client_secret" query:"client_secret"`
	Amount        int64                 `json:"amount" form:"amount" query:"amount"`
	Currency      string                `json:"currency" form:"currency" query:"currency"`
	Email         string                `json:"email" form:"email" query:"email"`
	//USDAmount int64 // TODO fetch the USD equivalent to allow granting premium on other currencies
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
