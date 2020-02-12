package jwt

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/discord"
	"github.com/labstack/echo/v4"
)

type discordRequest struct {
	Token string `json:"access_token" form:"access_token" query:"access_token"`
}

func DiscordLoginHandler(c echo.Context) error {
	var body discordRequest
	if err := c.Bind(&body); err != nil {
		return err
	}
	if body.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "access_token must be provided")
	}

	// Get the user's identity
	discordId, err := discord.GetUserId(body.Token)
	if err != nil {
		return err
	}

	user := database.LookupUserByDiscordID(discordId)
	if user == nil || len(user.Roles) <= 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "no premium user found")
	}

	return respondWithToken(user, c)
}
