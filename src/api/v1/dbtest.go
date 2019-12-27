package v1

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
)

func dbTest(c echo.Context) error {
	var value int
	err := database.DB.QueryRow("SELECT test FROM test").Scan(&value)
	if err != nil {
		return err
	}
	_, err = database.DB.Exec("UPDATE test SET test=test+1")
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, value)
}
