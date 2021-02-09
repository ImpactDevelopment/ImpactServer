package stripe

import (
	"errors"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/account"
	"github.com/stripe/stripe-go/v71/balancetransaction"
	"github.com/stripe/stripe-go/v71/paymentintent"
	"github.com/stripe/stripe-go/v71/reversal"
	"github.com/stripe/stripe-go/v71/transfer"
	"github.com/stripe/stripe-go/v71/webhook"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var PublicKey string
var webhookSecret string

// A list of accounts to distribute donations to
var connectedAccounts []stripe.Account
var accountsLock sync.Mutex

// The amount to remain in Impact's balance after distributing (plus any remainder from division)
var targetLeftover int64

func init() {
	// Set values from environment
	PublicKey = os.Getenv("STRIPE_PUBLIC_KEY")
	stripe.Key = os.Getenv("STRIPE_PRIVATE_KEY")
	webhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
	if target, err := strconv.ParseInt(os.Getenv("STRIPE_TARGET_LEFTOVER"), 10, 64); err == nil {
		targetLeftover = target
	} else {
		println("Error reading STRIPE_TARGET_LEFTOVER:", err.Error())
		targetLeftover = 0
	}

	// Fetch connected accounts
	accountsLock.Lock()
	defer accountsLock.Unlock()
	var err error
	connectedAccounts, err = getConnectedAccounts()

	// If no error on initial fetch, do it again every hour
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

func makePaymentStruct(intent *stripe.PaymentIntent) *Payment {
	return &Payment{
		PaymentIntent: intent,
		ClientSecret:  intent.ClientSecret,
		Amount:        intent.Amount,
		Currency:      intent.Currency,
		Email:         intent.Metadata["email"],
	}
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
		params.AddMetadata("email", email)
	}
	payment, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}

	return makePaymentStruct(payment), nil
}

func GetPayment(id string) (*Payment, error) {
	payment, err := paymentintent.Get(id, nil)
	if err != nil {
		return nil, err
	}

	return makePaymentStruct(payment), nil
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

// SendReceipt updates the payment description if a token is provided and sends a receipt if an email is associated with the payment
func SendReceipt(payment *stripe.PaymentIntent, token *uuid.UUID) error {
	var params stripe.PaymentIntentParams
	if email, ok := payment.Metadata["email"]; ok {
		params.ReceiptEmail = &email
	}
	if token != nil {
		params.AddMetadata("token", token.String())
		params.Description = stripe.String(payment.Description + "\nRegistration token: " + token.String())
	}
	_, err := paymentintent.Update(payment.ID, &params)
	return err
}

// DistributeDonation splits a paid charge evenly between the cached connected accounts
// Any leftovers remain in the Impact stripe account
func DistributeDonation(charge *stripe.Charge) error {
	accountsLock.Lock()
	defer accountsLock.Unlock()

	if !charge.Paid {
		return fmt.Errorf("cannot distribute payments from unpaid charge %s with status %s", charge.ID, charge.Status)
	}

	// We need to get the actual balance transaction
	// - the charge may be in a different currency to the final balance transaction
	// - the charge amount will be higher than the net amount on the balance transaction
	var bt *stripe.BalanceTransaction
	if charge.BalanceTransaction == nil {
		return fmt.Errorf("charge %s has no balance_transaction", charge.ID)
	} else {
		var err error
		bt, err = balancetransaction.Get(charge.BalanceTransaction.ID, nil)
		if err != nil {
			return fmt.Errorf("error getting balance_transaction %s for charge %s: %s", charge.BalanceTransaction.ID, charge.ID, err.Error())
		}
	}

	// Calculate number of shares
	shares := len(connectedAccounts)
	if shares < 1 {
		return errors.New("unable to distribute shares, zero shareholders")
	}

	// Calculate the value of each share
	share := (bt.Net - targetLeftover) / int64(shares)

	// Don't transfer negative values 😂
	// This could happen, for example, if targetLeftover > bt.Net
	if share <= 0 {
		return fmt.Errorf("calculated share (%.2f %s)is less than zero", float64(share)/100, bt.Currency)
	}

	// Distribute the shares
	// Keep track of created transfers so we can verify they were all created successfully
	var transfers []stripe.Transfer
	for _, acct := range connectedAccounts {

		// Do the transfer
		t, err := transfer.New(&stripe.TransferParams{
			Amount:            stripe.Int64(share),
			Currency:          stripe.String(string(bt.Currency)),
			Destination:       stripe.String(acct.ID),
			SourceTransaction: stripe.String(charge.ID),
		})

		if err == nil {
			transfers = append(transfers, *t)
		} else {
			fmt.Printf("Error distributing %.2f %s to %s\n", float64(share)/100, bt.Currency, acct.Email)
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

func ReverseDistribution(charge *stripe.Charge) error {
	// We're not actually touching connectedAccounts, but we don't want to create any transfers while reversing them...
	//accountsLock.Lock()
	//defer accountsLock.Unlock()

	if charge.TransferGroup == "" {
		// No transfers to reverse
		return nil
	}

	// Get all transfers related to this charge
	iter := transfer.List(&stripe.TransferListParams{
		TransferGroup: stripe.String(charge.TransferGroup),
	})

	// Keep track of the number of errors and transfers in case anything goes wrong
	var errs []error

	// Reverse each transfer
	for iter.Next() {
		t := iter.Transfer()
		_, err := reversal.New(&stripe.ReversalParams{
			Transfer: stripe.String(t.ID),
		})
		if err != nil {
			fmt.Printf("Error reversing transfer %s: %s\n", t.ID, err.Error())
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d errors encountered reversing %d transfers for charge %s", len(errs), iter.Meta().TotalCount, charge.ID)
	}

	return nil
}
