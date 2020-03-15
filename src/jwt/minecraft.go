package jwt

import (
	"github.com/ImpactDevelopment/ImpactServer/src/minecraft"
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
)

func MinecraftLoginHandler(c echo.Context) error {
	var body struct {
		Username string `json:"username" form:"username" query:"username"`
		Hash     string `json:"hash" form:"hash" query:"hash"`
	}
	if err := c.Bind(&body); err != nil {
		return err
	}
	if body.Username == "" || body.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "both username and hash must be provided")
	}
	profile, err := minecraft.HasJoinedServer(body.Username, body.Hash)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed authentication with mojang").SetInternal(err)
	}
	user := database.LookupUserByMinecraftID(profile.ID)
	if user == nil || len(user.Roles) <= 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "no premium user found")
	}

	return respondWithToken(user, c)
}
