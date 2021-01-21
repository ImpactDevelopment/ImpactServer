package stripe

import (
	"errors"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/account"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"github.com/stripe/stripe-go/v71/transfer"
	"github.com/stripe/stripe-go/v71/webhook"
	"net/http"
	"os"
	"sync"
	"time"
)

var PublicKey string
var webhookSecret string

// A list of accounts to distribute donations to
// TODO combine mutex lock & slice into one strut
var connectedAccounts []stripe.Account
var accountsLock sync.Mutex

func init() {
	// Set values from environment
	PublicKey = os.Getenv("STRIPE_PUBLIC_KEY")
	stripe.Key = os.Getenv("STRIPE_PRIVATE_KEY")
	webhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Fetch connected accounts
	var err error
	connectedAccounts, err = getConnectedAccounts()
	if err != nil {
		println("Error fetching Stripe connected accounts: ", err.Error())
	} else {
		util.DoRepeatedly(time.Hour, func() {
			accountsLock.Lock()
			defer accountsLock.Unlock()

			accounts, err := getConnectedAccounts()
			if err != nil {
				println("Error updating Stripe connected accounts: ", err.Error())
			}

			connectedAccounts = accounts
		})
	}
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

// DistributeDonation splits a payment evenly between the cached connected accounts
// Any leftovers remain in the Impact stripe account
func DistributeDonation(payment *stripe.PaymentIntent) error {
	accountsLock.Lock()
	defer accountsLock.Unlock()

	// Get charge id
	// There should only be one charge
	var chargeID *string
	for _, charge := range payment.Charges.Data {
		chargeID = &charge.ID
		break
	}
	if chargeID == nil {
		return fmt.Errorf("unable to get charge ID for payment %s\n", payment.ID)
	}

	// Calculate the value of each share
	shares := len(connectedAccounts)
	if shares < 1 {
		return errors.New("unable to distribute shares, zero shareholders")
	}

	// TODO get the target leftover amount from env?
	const targetLeftover = 50
	amount := payment.Amount
	share := (amount - targetLeftover) / int64(shares)
	leftover := amount - share*int64(shares)

	// Don't transfer negative values 😂
	if share <= 0 {
		return fmt.Errorf("calculated share (%.2f %s)is less than zero", float64(share)/100, payment.Currency)
	}

	// FIXME remove this debugging code?
	fmt.Printf("Distributing %d shares of %.2f %s, with a leftover of %.2f %s\n", shares, float64(share)/100, payment.Currency, float64(leftover)/100, payment.Currency)

	// Distribute the shares
	// Keep track of created transfers so we can verify they were all created successfully
	var transfers []stripe.Transfer
	for _, acct := range connectedAccounts {

		// Do the transfer
		t, err := transfer.New(&stripe.TransferParams{
			Amount:            stripe.Int64(share),
			Currency:          stripe.String(payment.Currency),
			Destination:       stripe.String(acct.ID),
			SourceTransaction: chargeID,
		})

		if err == nil {
			transfers = append(transfers, *t)
		} else {
			fmt.Printf("Error distributing %.2f %s to %s\n", float64(share)/100, payment.Currency, acct.Email)
		}
	}

	// Sanity check
	//TODO is this really necessary? If it is, should we attempt to reverse successful transfers when some fail?
	if len(transfers) != shares {
		return fmt.Errorf("not all transfers were created successfully, %d succeeded out of %d", len(transfers), shares)
	}

	return nil
}

// getConnectedAccounts returns a list of up to 10 connected accounts
func getConnectedAccounts() ([]stripe.Account, error) {
	// Fetch up to 10 connected accounts
	params := &stripe.AccountListParams{}
	params.Filters.AddFilter("limit", "", "10")
	iter := account.List(params)
	if iter.Err() != nil {
		return nil, iter.Err()
	}

	// Copy accounts to a new slice
	var accounts []stripe.Account
	for iter.Next() {
		accounts = append(accounts, *iter.Account())
	}

	return accounts, nil
}
