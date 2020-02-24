package jwt

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type minecraftRequest struct {
	Username string `json:"username" form:"username" query:"username"`
	Hash     string `json:"hash" form:"hash" query:"hash"`
}

func MinecraftLoginHandler(c echo.Context) error {
	var body minecraftRequest
	if err := c.Bind(&body); err != nil {
		return err
	}
	if body.Username == "" || body.Hash == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "both username and hash must be provided")
	}
	uuidStr, err := util.HasJoinedServer(body.Username, body.Hash)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed authentication with mojang").SetInternal(err)
	}
	minecraftID, err := uuid.Parse(uuidStr)
	if err != nil {
		return err
	}
	user := database.LookupUserByMinecraftID(minecraftID)
	if user == nil || len(user.Roles) <= 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "no premium user found")
	}

	return respondWithToken(user, c)
}
