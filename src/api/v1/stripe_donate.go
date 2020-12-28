package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/stripe"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	upstreamstripe "github.com/stripe/stripe-go/v71"
)

type redeemRequest struct {
	ID    string `json:"payment_id" form:"payment_id" query:"payment_id"`
	Email string `json:"email" form:"email" query:"email"`
}

type redeemResponse struct {
	Token string `json:"token" form:"token" query:"token"`
}

type createRequest struct {
	Amount int64  `json:"amount" form:"amount" query:"amount"`
	Email  string `json:"email" form:"email" query:"email"`
}

type createResponse struct {
	*stripe.Payment
}

func handleStripeWebhook(c echo.Context) error {
	// Read body payload
	const maxBodyBytes = int64(65536)
	var payload []byte
	var err error
	if c.Request().Body != nil {
		payload, err = ioutil.ReadAll(io.LimitReader(c.Request().Body, maxBodyBytes))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "unable to read request body").SetInternal(err)
		}
	}

	// Get & validate webhook event
	event, err := stripe.GetWebhookEvent(payload, c.Request().Header.Get("Stripe-Signature"))
	if err != nil {
		return err
	}

	// Choose a handler for the webhook event
	switch event.Type {
	// TODO do something with refund events:
	//charge.refund.updated
	//charge.refunded
	default:
		return echo.NewHTTPError(http.StatusNotImplemented, "unknown webhook type "+event.Type)
	}
}

func createStripePayment(c echo.Context) error {
	var body createRequest
	err := c.Bind(&body)
	if err != nil {
		return err
	}
	if body.Amount == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "order amount is empty")
	}

	// Validate email
	if body.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is empty")
	}
	if !util.IsValidEmail(body.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid email: "+body.Email)
	}

	payment, err := stripe.CreatePayment(body.Amount, "usd", "Donation", body.Email)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &createResponse{payment})
}

func redeemStripePayment(c echo.Context) error {
	var body redeemRequest
	err := c.Bind(&body)
	if err != nil {
		return err
	}
	if body.ID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "payment_id is empty")
	}
	payment, err := stripe.GetPayment(body.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Error getting details for payment "+body.ID).SetInternal(err)
	}

	// Validate order
	// Check email matches payment
	if payment.Email != body.Email {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Payment Email: Payment + "+body.ID+" was not made by "+body.Email)
	}
	// Check payment has succeeded
	if payment.PaymentIntent.Status != upstreamstripe.PaymentIntentStatusSucceeded {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Payment Status: Payment "+body.ID+" is "+string(payment.PaymentIntent.Status)+", expected status "+string(upstreamstripe.PaymentIntentStatusSucceeded))
	}
	// Check payment was in USD
	if payment.PaymentIntent.Currency != string(upstreamstripe.CurrencyUSD) {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Payment Currency: Payment "+body.ID+" is in "+payment.PaymentIntent.Currency+", expected "+string(upstreamstripe.CurrencyUSD))
	}
	// Check payment was over $5
	if payment.PaymentIntent.Amount < 500 {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Payment Amount: Payment "+body.ID+" totals "+strconv.FormatInt(payment.PaymentIntent.Amount, 10)+", expected 500 or more")
	}

	// INSERT if no conflict or simply SELECT if already exists
	var token uuid.UUID
	err = database.DB.QueryRow(`
		WITH new_pending_donation AS (
    		INSERT INTO pending_donations(stripe_payment_id, stripe_payer_email, amount, premium)
    		VALUES ($1, $2, $3, TRUE)
    		ON CONFLICT(stripe_payment_id) DO NOTHING
    		RETURNING token
		) SELECT COALESCE (
		    (SELECT token FROM new_pending_donation),
		    (SELECT token FROM pending_donations WHERE NOT used AND stripe_payment_id = $1)
		)`,
		body.ID, payment.Email, payment.Amount).Scan(&token)
	if err != nil {
		log.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving pending donation").SetInternal(err)
	}

	return c.JSON(http.StatusOK, &redeemResponse{
		Token: token.String(),
	})
}
