package web

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/labstack/echo/v4"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

var captcha recaptcha.ReCAPTCHA

func init() {
	recaptchaSecret := os.Getenv("RECAPTCHA_SECRET_KEY")
	if recaptchaSecret == "" {
		fmt.Println("WARNING: No recaptcha secret; discord verification is disabled!")
		return
	}
	var err error
	captcha, err = recaptcha.NewReCAPTCHA(recaptchaSecret, recaptcha.V2, 10*time.Second)
	if err != nil {
		panic(err)
	}
}

type verification struct {
	Recaptcha string `json:"g-recaptcha-response" form:"g-recaptcha-response" query:"g-recaptcha-response"`
	Token     string `json:"token" form:"token" query:"token"`
	Id        string `json:"discord" form:"discord" query:"discord"`
}

func discordVerify(c echo.Context) error {
	body := &verification{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	if body.Recaptcha == "" || (body.Id == "" && body.Token == "") {
		return echo.NewHTTPError(http.StatusBadRequest, "recapture and discord id (or access token) must be provided")
	}

	remoteIP := strings.Split(c.Request().Header.Get("X-FORWARDED-FOR"), ",")[0]
	// remoteIP is empty string if not present, which is exactly what this library expects
	err = captcha.VerifyWithOptions(body.Recaptcha, recaptcha.VerifyOption{RemoteIP: remoteIP, Hostname: util.GetServerURL().Hostname()})
	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Recaptcha failed").SetInternal(err)
	}

	// Get the user's identity
	var discordId string
	if body.Id != "" {
		discordId = body.Id
	} else {
		discordId, err = discord.GetUserId(body.Token)
		if err != nil {
			return err
		}
	}

	if !discord.CheckServerMembership(discordId) {
		return echo.NewHTTPError(http.StatusBadRequest, discordId+" doesn't appear to be a member of our discord?")
	}
	err = discord.GiveVerified(discordId)
	if err != nil {
		return err
	}
	return c.String(200, "Success")
}
