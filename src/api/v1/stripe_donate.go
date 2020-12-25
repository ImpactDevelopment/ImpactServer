package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/stripe"
	"github.com/labstack/echo/v4"
	"net/http"
)

type createRequest struct {
	Amount int64
}

type createResponse struct {
	*stripe.Payment
}

func createStripePayment(c echo.Context) error {
	// TODO
	var body createRequest
	err := c.Bind(&body)
	if err != nil {
		return err
	}
	if body.Amount == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "order amount is empty")
	}

	payment, err := stripe.CreatePayment(body.Amount, "usd")
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &createResponse{
		payment,
	})
}

func afterStripePayment(c echo.Context) error {
	// TODO
	return echo.NewHTTPError(http.StatusNotImplemented)
}
