package v1

import (
	"errors"
	"github.com/ImpactDevelopment/ImpactServer/src/paypal"
	"github.com/labstack/echo/v4"
	"net/http"
)

type donationRequest struct {
	ID string `json:"orderID" form:"orderID"`
}

// TODO
//type donationResponse struct {
//	Amount int           `json:"amount"`
//	Order  *paypal.Order `json:"order,omitempty"`
//}

func donate(c echo.Context) error {
	body := &donationRequest{}
	err := c.Bind(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if body.ID == "" {
		return c.JSON(http.StatusBadRequest, errors.New("order ID missing"))
	}
	order, err := paypal.GetAndAuthorizeOrder(body.ID)
	if err != nil {
		// TODO check the error before choosing a code?
		return c.JSON(http.StatusUnprocessableEntity, err)
	}
	perks := order.Amount >= 500

	if perks {
		// TODO
	}

	// TODO
	return c.JSONPretty(http.StatusOK, order, "    ")
}
