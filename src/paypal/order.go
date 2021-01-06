package paypal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/plutov/paypal/v3"
)

type Order struct {
	*paypal.Order

	PayerID    string
	PayerEmail string

	// Remember the requested order id for validation purposes
	// Overkill tbh
	requestedId string
}

// GetOrder returns a paypal order
func GetOrder(id string) (*Order, error) {
	if client == nil {
		return nil, errors.New("no paypal client is setup")
	}

	order, err := client.GetOrder(id)
	if err != nil {
		return nil, err
	}

	return &Order{Order: order, requestedId: id}, nil
}

// Total returns the total amount in US cent
func (o Order) Total() (cent int64) {
	cent = 0
	for _, purchase := range o.PurchaseUnits {
		cent += amountAsCent(purchase.Amount)
	}
	return
}

// Currency returns the first currency found in the order
func (o Order) Currency() string {
	for _, purchase := range o.PurchaseUnits {
		if currency := purchase.Amount.Currency; currency != "" {
			return currency
		}
	}
	return ""
}

// Capture will capture "approved" orders and return an error if it fails
func (o *Order) Capture() error {
	switch o.Status {
	case "COMPLETED":
		// Nothing to do, already captured
	case "APPROVED":
		capture, err := client.CaptureOrder(o.ID, paypal.CaptureOrderRequest{})
		if err != nil {
			return err
		}

		// Check the capture was a success
		if capture.Status != "COMPLETED" {
			return errors.New("intent was CAPTURE but status is still not COMPLETED after capturing")
		}

		// Update the Order struct
		newOrder, err := client.GetOrder(o.ID)
		if err != nil {
			return err
		}
		o.Order = newOrder

		if capture.Payer != nil {
			o.PayerID = capture.Payer.PayerID
			o.PayerEmail = capture.Payer.EmailAddress
		}
	default:
		return errors.New("Unknown order status " + o.Status)
	}

	return nil
}

// Validate tries to check for invalid orders. If something is wrong with the order it returns an error
func (o Order) Validate() error {
	if o.requestedId != o.ID {
		return fmt.Errorf("requested order ID %s does not match %s", o.requestedId, o.ID)
	}
	if o.Intent != "CAPTURE" && o.Intent != "AUTHORIZE" {
		return fmt.Errorf("intent was %s not CAPTURE or AUTHORIZE", o.Intent)
	}
	if o.Status != "COMPLETED" {
		return fmt.Errorf("intent was %s but status is %s (not COMPLETED)", o.Intent, o.Status)
	}
	if len(o.PurchaseUnits) < 1 {
		return errors.New("no purchase units")
	}

	return nil
}

// amountAsCent is a helper method used by Order.Total()
func amountAsCent(amount *paypal.PurchaseUnitAmount) int64 {
	if amount.Currency != "USD" {
		// We can only sum up USD amounts since we don't know what the exchange rate is
		// FIXME this means foreign donations don't qualify for perks
		return 0
	}

	cent, err := strconv.Atoi(strings.ReplaceAll(amount.Value, ".", ""))
	if err != nil {
		return 0
	}

	return int64(cent)
}
