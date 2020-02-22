package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func premiumCheck(c echo.Context) error {
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
