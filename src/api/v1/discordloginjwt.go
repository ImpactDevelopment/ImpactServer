package v1

import (
	"errors"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"net/http"
	"regexp"

	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
)

type discordLoginRequest struct {
	Token string `json:"access_token" form:"access_token" query:"access_token"`
}

// Discord's OAuth tokens are alphanumeric
var discordTokenPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func discordLoginJWT(c echo.Context) error {
	var body discordLoginRequest
	if err := c.Bind(&body); err != nil {
		return err
	}
	if body.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "access_token must be provided")
	}

	// Validate the token, prevent trying to auth with discord using some completely invalid token
	if !discordTokenPattern.MatchString(body.Token) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid access_token "+body.Token)
	}

	// Create a discord session using the provided token. Does not verify the token is valid in any way.
	// Using discordgo here is massively overkill, but who cares
	session, err := discordgo.New("Bearer " + body.Token)
	if err != nil {
		return err
	}
	defer session.Close()

	// Get the user's identity
	discordUser, err := session.User("@me")
	if err != nil {
		var restErr *discordgo.RESTError
		if errors.As(err, &restErr) {
			// TODO should this be a 500? Or maybe a 401?
			return echo.NewHTTPError(http.StatusUnprocessableEntity, fmt.Sprintf(`error authenticating with discord "%s"`, restErr.Message.Message))
		}
		return err
	}
	if discordUser.ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	user := database.LookupUserByDiscordID(discordUser.ID)
	if user == nil || len(user.Roles) <= 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "no premium user found")
	}
	return c.JSONBlob(http.StatusOK, jwt.CreateJWT(user, ""))
}
