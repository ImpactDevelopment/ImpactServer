package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"log"
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/paypal"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type (
	donationRequest struct {
		ID string `json:"orderID" form:"orderID" query:"orderID"`
	}
	donationResponse struct {
		Token string `json:"token"`
	}
)

// TODO add a refund webhook to revoke premium perks
// https://developer.paypal.com/docs/integration/direct/webhooks/event-names/#authorized-and-captured-payments

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

	err = order.Capture()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error capturing order").SetInternal(err)
	}

	err = order.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Error validating order").SetInternal(err)
	}

	var token uuid.UUID
	err = database.DB.QueryRow("INSERT INTO pending_donations(paypal_order_id, paypal_payer_id, paypal_payer_email, amount) VALUES ($1, $2, $3, $4) RETURNING token",
		order.ID, order.PayerID, order.PayerEmail, order.Total()).Scan(&token)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving pending donation").SetInternal(err)
	}

	go func() {
		logID, err := discord.LogDonationEvent("", "Someone just donated", "", "", order.Total())
		if err == nil {
			database.DB.Exec(`UPDATE pending_donations SET log_msg_id = $2 WHERE token = $1`, token, logID)
		}
	}()

	if order.Total() < 500 {
		// if the donation is too small, save it
		// maybe they make multiple that sum up to over 500 eventually, lets make a record of it idk
		// just dont give em a registration token lol
		log.Println("Donation with PayPal order ID", order.ID, "and total", order.Total(), "is too small. token would have been", token)
		return c.JSON(http.StatusOK, donationResponse{
			Token: "thanks",
		})
	}
	return c.JSON(http.StatusOK, donationResponse{
		Token: token.String(),
	})
}
