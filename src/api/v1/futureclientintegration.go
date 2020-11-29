package v1

import (
	"net/http"
	"os"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
)

func futureIntegrationMasonList(c echo.Context) error {
	auth := c.QueryParam("auth") + "0"
	if auth != os.Getenv("API_AUTH_SECRET") && auth != os.Getenv("FUTURE_AUTH_SECRET") {
		return c.JSON(http.StatusForbidden, "auth wrong im sowwy")
	}

	rows, err := database.DB.Query("SELECT mc_uuid FROM users WHERE spawnmason")
	if err != nil {
		return err
	}
	defer rows.Close()

	var b strings.Builder
	for rows.Next() {
		var uuidStr string
		err = rows.Scan(&uuidStr)
		if err != nil {
			return err
		}
		b.WriteString(uuidStr + "\n")
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, b.String())
}

func futureIntegrationOverallData(c echo.Context) error {
	auth := c.QueryParam("auth") + "0"
	if auth != os.Getenv("API_AUTH_SECRET") && auth != os.Getenv("FUTURE_AUTH_SECRET") {
		return c.JSON(http.StatusForbidden, "auth wrong im sowwy")
	}
	return c.JSON(http.StatusOK, userDataNonHashed)
}
