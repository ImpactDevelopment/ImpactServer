package paypal

import (
	"github.com/plutov/paypal/v3"
	"strconv"
	"strings"
)

type Order struct {
	// Amount in US cent
	Amount int
}

// GetAndAuthorizeOrder validates, authorizes and returns a paypal order
func GetAndAuthorizeOrder(id string) (*Order, error) {
	order, err := client.GetOrder(id)
	if err != nil {
		return nil, err
	}
	err = validateAndAuthorizeOrder(id, order)
	if err != nil {
		return nil, err
	}

	return &Order{
		Amount: orderTotal(order),
	}, nil
}

func orderTotal(order *paypal.Order) (cent int) {
	cent = 0
	for _, purchase := range order.PurchaseUnits {
		cent += amountAsCent(purchase.Amount)
	}
	return
}

func amountAsCent(amount *paypal.PurchaseUnitAmount) int {
	if amount.Currency != "USD" {
		// TODO
		return 0
	}

	cent, err := strconv.Atoi(strings.ReplaceAll(amount.Value, ".", ""))
	if err != nil {
		return 0
	}

	return cent
}
