package v1

import (
	"net/http"
	"os"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func premiumCheck(c echo.Context) error {
	if c.QueryParam("auth")+"0" != os.Getenv("API_AUTH_SECRET") {
		return c.JSON(http.StatusForbidden, "auth wrong im sowwy")
	}
	uuidStr := c.QueryParam("uuid")
	minecraftID, err := uuid.Parse(uuidStr)
	if err != nil {
		return c.JSON(http.StatusForbidden, "uuid is bad?")
	}
	user := database.LookupUserByMinecraftID(minecraftID)
	if user == nil || len(user.Roles) <= 0 {
		return echo.NewHTTPError(http.StatusForbidden, "no premium user found for uuid "+minecraftID.String())
	}
	return c.JSON(http.StatusOK, user.RoleIDs())
}
