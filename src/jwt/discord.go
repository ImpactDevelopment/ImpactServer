package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/bwmarrin/discordgo"
	"github.com/labstack/echo/v4"
)

type discordRequest struct {
	Token string `json:"access_token" form:"access_token" query:"access_token"`
}

// Discord's OAuth tokens are alphanumeric
var discordOAuthToken = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func DiscordLoginHandler(c echo.Context) error {
	var body discordRequest
	if err := c.Bind(&body); err != nil {
		return err
	}
	if body.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "access_token must be provided")
	}

	// Validate the token, prevent trying to auth with discord using some completely invalid token
	if !discordOAuthToken.MatchString(body.Token) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid access_token "+body.Token)
	}

	// Create a discord session using the provided token. Does not verify the token is valid in any way.
	// Using discordgo here is massively overkill, but who cares
	// This won't use websockets unless we call session.Open(), so there's no need to call Close() either.
	session, err := discordgo.New("Bearer " + body.Token)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "error setting up discord session").SetInternal(err)
	}

	// Get the user's identity
	discordUser, err := session.User("@me")
	if err != nil {
		var restErr *discordgo.RESTError
		if errors.As(err, &restErr) {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf(`error authenticating with discord "%s"`, restErr.Message.Message))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "error authenticating with discord").SetInternal(err)
	}
	if discordUser.ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "no discord user found")
	}

	user := database.LookupUserByDiscordID(discordUser.ID)
	if user == nil || len(user.Roles) <= 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "no premium user found")
	}

	return respondWithToken(user, c)
}
