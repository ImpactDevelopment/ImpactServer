package paypal

import (
	"errors"
	"github.com/plutov/paypal/v3"
)

// validateAndAuthorizeOrder returns an error if the order isn't valid. If the order is APPROVED
// but not COMPLETED then it will try to Authorize the order.
func validateAndAuthorizeOrder(id string, order *paypal.Order) (err error) {
	if id != order.ID {
		err = errors.New("Order ID mis-match, " + id + " does not match " + order.ID)
		return
	}

	switch order.Intent {
	case "CAPTURE":
		{
			if order.Status != "COMPLETED" {
				err = errors.New("intent was CAPTURE but status is not COMPLETED")
				return
			}
		}
	case "AUTHORIZE":
		{
			switch order.Status {
			case "COMPLETED":
				{
					// Fair enough, already approved; nothing to do
				}
			case "APPROVED":
				{
					var authorization *paypal.Authorization
					authorization, err = client.AuthorizeOrder(order.ID, paypal.AuthorizeOrderRequest{
						PaymentSource:      nil,
						ApplicationContext: paypal.ApplicationContext{},
					})
					if err != nil {
						return
					}
					if authorization.Status != "COMPLETED" {
						err = errors.New("intent was AUTHORIZE but status is still not COMPLETED after authorizing")
						return
					}
				}
			default:
				{
					err = errors.New("intent was AUTHORIZE but status was neither COMPLETED nor APPROVED")
					return
				}
			}
		}
	}

	if len(order.PurchaseUnits) < 1 {
		err = errors.New("no purchase units")
		return
	}

	return
}
