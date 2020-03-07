package web

import (
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/ImpactDevelopment/ImpactServer/src/recaptcha"
	"github.com/labstack/echo/v4"
	"net/http"
)

func simpleRecaptchaCheck(c echo.Context) error {
	err := recaptcha.Verify(c)
	if err != nil {
		return err
	}
	return c.String(200, "Success")
}

type verification struct {
	Token string `json:"token" form:"token" query:"token"`
	Id    string `json:"discord" form:"discord" query:"discord"`
}

func discordVerify(c echo.Context) error {
	body := &verification{}
	err := c.Bind(body)
	if err != nil {
		return err
	}
	err = recaptcha.Verify(c)
	if err != nil {
		return err
	}

	if body.Id == "" && body.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, " discord id (or access token) must be provided")
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

	if discord.CheckServerMembership(discordId) {
		err = discord.GiveVerified(discordId)
		if err != nil {
			return err
		}
	} else {
		if body.Token == "" {
			// they arent in, and we cant join since no token :(
			return echo.NewHTTPError(http.StatusBadRequest, discordId+" doesn't appear to be a member of our discord?")
		} else {
			err = discord.JoinOurServer(body.Token, discordId, false)
			if err != nil {
				return err
			}
		}
	}

	return c.Redirect(http.StatusFound, "https://discordapp.com/channels/208753003996512258/222120655594848256")
}
