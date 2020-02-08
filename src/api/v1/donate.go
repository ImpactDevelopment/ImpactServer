package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/ImpactDevelopment/ImpactServer/src/paypal"
	"github.com/labstack/echo/v4"
	"net/http"
)

type (
	donationRequest struct {
		ID string `json:"orderID" form:"orderID" query:"orderID"`
	}
	donationResponse struct {
		Amount int    `json:"amount"`
		Token  string `json:"token,omitempty"`
	}
)

func afterDonation(c echo.Context) error {
	body := &donationRequest{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	if body.ID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "order ID missing")
	}

	order, err := paypal.GetOrder(body.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Error getting order details for id "+body.ID).SetInternal(err)
	}

	err = order.Authorize()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error authorizing order").SetInternal(err)
	}

	err = order.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Error validating order").SetInternal(err)
	}

	// This token can be used to register for an impact account, assuming amount is high enough
	token := jwt.CreateDonationJWT(order)
	if token == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error creating donation token")
	}

	// TODO instead of returning a token, should we store it in the database and return a short token id instead?
	// The jwt is rather long if users are planning to share it as a gift...
	// Another option would be to compress it somehow maybe 🤔
	return c.JSON(http.StatusOK, donationResponse{
		Amount: order.Total(),
		Token:  token,
		// TODO return expiry too?
	})
}
