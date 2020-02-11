package web

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/discord"
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
	Discord   string `json:"discord" form:"discord" query:"discord"`
}

func discordVerify(c echo.Context) error {
	body := &verification{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	opt := recaptcha.VerifyOption{}
	forward := strings.Split(c.Request().Header.Get("X-FORWARDED-FOR"), ",")[0]
	if forward != "" {
		opt.RemoteIP = forward
	}
	fmt.Println("Passing remote IP to recaptcha:", opt.RemoteIP)
	err = captcha.VerifyWithOptions(body.Recaptcha, opt)
	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Recaptcha failed").SetInternal(err)
	}

	if !discord.CheckServerMembership(body.Discord) {
		return echo.NewHTTPError(http.StatusBadRequest, body.Discord+" doesn't appear to be a member of our discord?")
	}
	err = discord.GiveVerified(body.Discord)
	if err != nil {
		return err
	}
	return c.String(200, "Success")
}
