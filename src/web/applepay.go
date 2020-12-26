package web

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
)

var applePayVerification string

func init() {
	applePayVerification = os.Getenv("APPLE_PAY_VERIFICATION")
}

func applePayVerify(c echo.Context) error {
	return c.String(http.StatusOK, applePayVerification)
}
