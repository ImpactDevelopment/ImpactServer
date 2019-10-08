package v1

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
)

func dbTest(c echo.Context) error {
	var value string
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
