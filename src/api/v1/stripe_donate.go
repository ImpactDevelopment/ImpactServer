package v1

import (
	"encoding/json"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/ImpactDevelopment/ImpactServer/src/stripe"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"

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
	Premium bool `json:"premium" form:"premium" query:"premium"`
}

// donationLock should be used while editing the DB or discord messages related to a donation
var donationLock sync.Mutex

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

	return c.JSON(http.StatusOK, &createResponse{
		Payment: payment,
		Premium: payment.Amount >= 500,
	})
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

	// Now that we are interacting with the DB we should lock
	donationLock.Lock()

	// Store the donation in the DB - or fetch it if it already exists
	token, logID, err := getOrCreateDonation(body.ID, payment.Email, payment.Amount)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving pending donation").SetInternal(err)
	}

	// Log the donation to discord
	go func() {
		_ = editOrCreateDonationLog("Someone just donated and generated a token", payment.Amount, token, logID)
		donationLock.Unlock() // Done messing with donation logging
	}()

	return c.JSON(http.StatusOK, &redeemResponse{
		Token: token.String(),
	})
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
	case "payment_intent.succeeded":
		var paymentIntent upstreamstripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error parsing webhook JSON").SetInternal(err)
		}
		return handlePaymentSucceeded(c, event, paymentIntent)
	case "charge.refunded":
		var refund upstreamstripe.Refund
		err := json.Unmarshal(event.Data.Raw, &refund)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error parsing webhook JSON").SetInternal(err)
		}
		return handleRefund(c, event, refund)
	// TODO: Handle failed refunds; charge.refund.updated with status:failed along with a failure_reason and failure_balance_transaction
	//       https://stripe.com/docs/refunds#failed-refunds
	//case "charge.refund.updated":
	default:
		return echo.NewHTTPError(http.StatusNotImplemented, "unknown webhook type "+event.Type)
	}
}

func handlePaymentSucceeded(c echo.Context, event *stripe.WebhookEvent, payment upstreamstripe.PaymentIntent) error {
	donationLock.Lock()
	defer donationLock.Unlock()

	// Check the DB to see if a pending_donation already exists, create one if not
	token, logID, err := getOrCreateDonation(payment.ID, payment.ReceiptEmail, payment.Amount)
	if err != nil {
		return err
	}

	_ = editOrCreateDonationLog("Someone just donated", payment.Amount, token, logID)

	return c.NoContent(http.StatusOK)
}

func handleRefund(c echo.Context, event *stripe.WebhookEvent, refund upstreamstripe.Refund) error {
	donationLock.Lock()
	defer donationLock.Unlock()

	payment := refund.PaymentIntent
	if payment == nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "unable to handle refunds not associated with a PaymentIntent")
	}

	// TODO Check the DB to see if a pending_donation already exists, create one if not

	if payment.Amount >= 500 {
		// TODO Mark the token as refunded

		// TODO check if the token was used_by anyone - if so disable or otherwise remove their account
		// If they had registered, populate userinfo vars above so we can use them in the donate log message
	}

	// TODO log refund to discord
	// consider also DMing the devs or posting something somewhere like #staff-anouncements or #senior-cisizens?

	return c.NoContent(http.StatusOK)
}

// Helper func to add a donation to pending_donations - or fetch the token if it already exists
func getOrCreateDonation(paymentID string, email string, amount int64) (token uuid.UUID, logID string, err error) {
	// INSERT if no conflict or simply SELECT if already exists
	err = database.DB.QueryRow(`
		WITH new_pending_donation AS (
    		INSERT INTO pending_donations(stripe_payment_id, stripe_payer_email, amount, premium)
    		VALUES ($1, $2, $3, TRUE)
    		ON CONFLICT(stripe_payment_id) DO NOTHING
    		RETURNING token, log_msg_id
		) SELECT COALESCE (
		    (SELECT token, log_msg_id FROM new_pending_donation),
		    (SELECT token, log_msg_id FROM pending_donations WHERE NOT used AND stripe_payment_id = $1)
		)`,
		paymentID, email, amount).Scan(&token, &logID)
	if err != nil {
		log.Println(err)
	}
	return
}

// Helper func to edit the donation discord log - or create on if it doesn't exist
func editOrCreateDonationLog(message string, amount int64, token uuid.UUID, logID string) error {
	newLogMsg := logID == ""
	logID, err := discord.LogDonationEvent(logID, message, "", nil, amount)
	if newLogMsg && err == nil {
		database.DB.Exec(`UPDATE pending_donations SET log_msg_id = $2 WHERE token = $1`, token, logID)
	}
	return err
}
