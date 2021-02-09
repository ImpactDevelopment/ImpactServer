package v1

import (
	"database/sql"
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
	"strings"
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
	Currency string `json:"currency" form:"currency" query:"currency"`
	Amount   int64  `json:"amount" form:"amount" query:"amount"`
	Email    string `json:"email" form:"email" query:"email"`
}

type createResponse struct {
	*stripe.Payment
	Premium bool `json:"premium" form:"premium" query:"premium"`
}

type stripeInfoReqponse struct {
	Version         string                         `json:"stripe_api_version" form:"stripe_api_version" query:"stripe_api_version"`
	PubKey          string                         `json:"stripe_public_key" form:"stripe_public_key" query:"stripe_public_key"`
	DefaultCurrency string                         `json:"default_currency" form:"default_currency" query:"default_currency"`
	Currencies      map[string]stripe.CurrencyInfo `json:"currencies" form:"currencies" query:"currencies"`
}

const defaultCurrency = "usd"

// donationLock should be used while editing the DB or discord messages related to a donation
var donationLock sync.Mutex

func getStripeInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, &stripeInfoReqponse{
		Version:         upstreamstripe.APIVersion,
		PubKey:          stripe.PublicKey,
		DefaultCurrency: defaultCurrency,
		Currencies:      stripe.GetCurrencyMap(),
	})
}

func createStripePayment(c echo.Context) error {
	var body createRequest
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	// Default currency
	if body.Currency == "" {
		body.Currency = defaultCurrency
	} else {
		body.Currency = strings.ToLower(strings.TrimSpace(body.Currency))
	}

	// Validate currency
	currency, err := stripe.GetCurrencyInfo(body.Currency)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
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

	payment, err := stripe.CreatePayment(body.Amount, body.Currency, "Donation", body.Email)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &createResponse{
		Payment: payment,
		Premium: payment.Amount >= currency.Amount,
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
	// Check payment is a valid currency
	currency, err := stripe.GetCurrencyInfo(payment.Currency)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Payment Currency: Payment "+body.ID+" is in "+payment.Currency+", which isn't supported")
	}
	// Check payment was enough for perks
	if payment.PaymentIntent.Amount < currency.Amount {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad Payment Amount: Payment "+body.ID+" totals "+strconv.FormatInt(payment.Amount, 10)+", expected "+strconv.FormatInt(currency.Amount, 10)+" or more")
	}

	// Now that we are interacting with the DB we should lock
	donationLock.Lock()

	// Store the donation in the DB - or fetch it if it already exists
	token, err := getOrCreateDonation(body.ID, payment.Email, payment.Currency, payment.Amount)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving pending donation").SetInternal(err)
	}

	// Log the donation to discord
	go func() {
		_ = editOrCreateDonationLog("Someone just donated and generated a token", payment.PaymentIntent, token)
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
		if err := unmarshal(event, &paymentIntent); err != nil {
			return err
		}
		return handlePaymentSucceeded(c, event, &paymentIntent)
	case "charge.succeeded":
		var charge upstreamstripe.Charge
		if err := unmarshal(event, &charge); err != nil {
			return err
		}
		return handleChargeSucceeded(c, event, &charge)
	case "charge.refunded":
		var refund upstreamstripe.Charge
		if err := unmarshal(event, &refund); err != nil {
			return err
		}
		return handleRefund(c, event, &refund)
	// TODO: Handle failed refunds; charge.refund.updated with status:failed along with a failure_reason and failure_balance_transaction
	//       https://stripe.com/docs/refunds#failed-refunds
	//case "charge.refund.updated":
	default:
		return echo.NewHTTPError(http.StatusNotImplemented, "unknown webhook type "+event.Type)
	}
}

func unmarshal(event *stripe.WebhookEvent, it interface{}) error {
	err := json.Unmarshal(event.Data.Raw, it)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Error parsing webhook JSON").SetInternal(err)
	}
	return nil
}

func handlePaymentSucceeded(c echo.Context, event *stripe.WebhookEvent, payment *upstreamstripe.PaymentIntent) error {
	donationLock.Lock()
	defer donationLock.Unlock()

	// Check the DB to see if a pending_donation already exists, create one if not
	token, err := getOrCreateDonation(payment.ID, payment.Metadata["email"], payment.Currency, payment.Amount)
	if err != nil {
		return err
	}

	// Update the payment with the token and send the email receipt
	err = stripe.SendReceipt(payment, &token)
	if err != nil {
		return err
	}

	_ = editOrCreateDonationLog("Someone just donated", payment, token)

	return c.NoContent(http.StatusOK)
}

func handleChargeSucceeded(c echo.Context, event *stripe.WebhookEvent, charge *upstreamstripe.Charge) error {
	// Distribute charge amount between connected accounts
	// We do this on charge succeeded instead of payment succeeded so we don't have to sort through successful and failed charges
	err := stripe.DistributeDonation(charge)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Error distributing charge").SetInternal(err)
	}
	return c.NoContent(http.StatusOK)
}

func handleRefund(c echo.Context, event *stripe.WebhookEvent, charge *upstreamstripe.Charge) error {
	// First things first, lets reverse any associated transfers (i.e. share distributions to connected accounts)
	err := stripe.ReverseDistribution(charge)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error reversing transfers for refunded charge "+charge.ID).SetInternal(err)
	}

	// Next, revoke any perks granted by this donation
	if charge.PaymentIntent != nil {
		err = revokeDonation(charge.PaymentIntent)
		if err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

// Helper func to add a donation to pending_donations - or fetch the token if it already exists
func getOrCreateDonation(paymentID string, email string, currency string, amount int64) (token uuid.UUID, err error) {
	// INSERT if no conflict or simply SELECT if already exists
	err = database.DB.QueryRow(`
		WITH new_pending_donation AS (
    		INSERT INTO pending_donations(stripe_payment_id, stripe_payer_email, currency, amount, premium)
    		VALUES ($1, $2, $3, $4, TRUE)
    		ON CONFLICT(stripe_payment_id) DO NOTHING
    		RETURNING token
		) SELECT COALESCE (
		    (SELECT token FROM new_pending_donation),
		    (SELECT token FROM pending_donations WHERE NOT used AND stripe_payment_id = $1)
		)`,
		paymentID, email, currency, amount).Scan(&token)
	if err != nil {
		log.Println(err)
	}
	return
}

// Helper func to edit the donation discord log - or create on if it doesn't exist
func editOrCreateDonationLog(message string, payment *upstreamstripe.PaymentIntent, token uuid.UUID) error {
	// Get logID if it exitst
	var logID sql.NullString
	database.DB.QueryRow(`SELECT log_msg_id FROM pending_donations WHERE token = $1`, token).Scan(&logID)

	newLogID, err := discord.LogDonationEvent(logID.String, message, "", nil, payment.Currency, payment.Amount)
	if !logID.Valid && err == nil {
		database.DB.Exec(`UPDATE pending_donations SET log_msg_id = $2 WHERE token = $1`, token, newLogID)
	}
	return err
}

// revokeDonation mark's the associated token as used, and deletes any user created with the token.
// It also updates the discord log message accordingly
func revokeDonation(payment *upstreamstripe.PaymentIntent) error {
	donationLock.Lock()
	defer donationLock.Unlock()

	var token uuid.UUID
	var user *uuid.UUID
	err := database.DB.QueryRow(`SELECT token, used_by FROM pending_donations WHERE stripe_payment_id=$1`, payment.ID).Scan(&token, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			// No token has been generated with this payment, nothing to do
			return nil
		} else {
			// Some other SQL error - uh oh
			return echo.NewHTTPError(http.StatusInternalServerError, "sql error finding tokens associated with payment").SetInternal(err)
		}
	}

	// Make DB changes in a transaction
	tx, err := database.DB.Begin()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "sql error creating transaction").SetInternal(err)
	}
	defer tx.Rollback()

	// TODO add a specific refunded field instead of just setting used=true
	_, err = tx.Exec(`UPDATE pending_donations SET used = true WHERE token=$1`, token)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "sql error marking token as used").SetInternal(err)
	}

	if user != nil {
		_, err = tx.Exec(`DELETE FROM users WHERE user_id=$1`, user)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "sql error deleting refunded user").SetInternal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Print(err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "sql error committing transaction").SetInternal(err)
	}

	// log refund to discord
	// TODO consider also DMing the devs or posting something somewhere like #staff-announcements or #senior-citizens?
	err = editOrCreateDonationLog("This donation was refunded", payment, token)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error logging refund to discord").SetInternal(err)
	}

	return nil
}
