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
func (o Order) Total() (cent int) {
	cent = 0
	for _, purchase := range o.PurchaseUnits {
		cent += amountAsCent(purchase.Amount)
	}
	return
}

// Authorize will authorize "approved" orders and return an error if it fails
func (o *Order) Authorize() error {
	if o.Intent == paypal.OrderIntentAuthorize {
		switch o.Status {
		case "COMPLETED":
			// Nothing to do, already authorized
		case "APPROVED":
			authorization, err := client.AuthorizeOrder(o.ID, paypal.AuthorizeOrderRequest{})
			if err != nil {
				return err
			}

			// Check the authorization was a success
			if authorization.Status != "COMPLETED" {
				return errors.New("intent was AUTHORIZE but status is still not COMPLETED after authorizing")
			}

			// Update the Order struct
			newOrder, err := client.GetOrder(o.ID)
			if err != nil {
				return err
			}
			o.Order = newOrder
		default:
			return errors.New("Unknown order status " + o.Status)
		}
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
func amountAsCent(amount *paypal.PurchaseUnitAmount) int {
	if amount.Currency != "USD" {
		// We can only sum up USD amounts since we don't know what the exchange rate is
		// FIXME this means foreign donations don't qualify for perks
		return 0
	}

	cent, err := strconv.Atoi(strings.ReplaceAll(amount.Value, ".", ""))
	if err != nil {
		return 0
	}

	return cent
}
