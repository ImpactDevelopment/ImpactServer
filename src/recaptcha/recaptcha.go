package recaptcha

import (
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/labstack/echo/v4"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
	"net/http"
	"os"
	"time"
)

var captcha recaptcha.ReCAPTCHA

func init() {
	secret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secret == "" {
		fmt.Println("WARNING: No recaptcha secret; discord verification is disabled!")
		return
	}
	var err error
	captcha, err = recaptcha.NewReCAPTCHA(secret, recaptcha.V2, 10*time.Second)
	if err != nil {
		panic(err)
	}
}

type request struct {
	Recaptcha string `json:"g-recaptcha-response" form:"g-recaptcha-response" query:"g-recaptcha-response"`
}

// Verify returns a HTTP error if the provided g-recaptcha-response is invalid
func Verify(c echo.Context) error {
	body := &request{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	if body.Recaptcha == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "recaptcha must be provided")
	}

	// remoteIP is empty string if not present, which is exactly what this library expects
	err = captcha.VerifyWithOptions(body.Recaptcha, recaptcha.VerifyOption{
		RemoteIP: util.RealIPIfUnambiguous(c),
		Hostname: util.GetServerURL().Hostname(),
	})
	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Recaptcha failed").SetInternal(err)
	}

	return nil
}
