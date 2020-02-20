package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/labstack/echo/v4"
	"net/http"
)

func getUser(c echo.Context) error {
	if user, ok := c.Get("user").(*users.User); ok {
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "no user found")
		}
		return c.JSON(http.StatusOK, user)
	} else {
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to cast user")
	}
}
