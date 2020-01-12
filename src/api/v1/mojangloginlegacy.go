package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func mojangLoginLegacy(c echo.Context) error {
	uuidStr, err := util.HasJoinedServer(c.QueryParam("username"), c.QueryParam("hash"))
	if err != nil {
		return err
	}
	uuidVal, err := uuid.Parse(uuidStr) // we do this to add the dashes btw
	if err != nil {
		return err
	}
	user := database.LookupUserByMinecraftID(uuidVal)
	if user != nil && len(user.Roles) > 0 {
		return c.JSON(http.StatusOK, struct {
			Roles []users.Role `json:"roles"`
		}{Roles: user.Roles})
	}
	return c.JSON(http.StatusForbidden, []struct{}{})
}
