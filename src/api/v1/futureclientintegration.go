package v1

import (
	"net/http"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
)

func futureIntegration(c echo.Context) error {
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
