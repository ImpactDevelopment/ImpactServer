package v1

import (
	"net/http"
	"os"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
)

func futureIntegration(c echo.Context) error {
	if c.QueryParam("auth")+"0" != os.Getenv("API_AUTH_SECRET") {
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
