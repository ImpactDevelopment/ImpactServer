package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

func checkDonator(c echo.Context) error {
	var premium bool
	err := database.DB.QueryRow("SELECT premium FROM users WHERE discord_id = $1", c.Param("discordid")).Scan(&premium)
	if err != nil {
		log.Println(err)
	}
	if premium {
		return c.String(http.StatusOK, "yes")
	} else {
		return c.String(http.StatusOK, "no")
	}
}

func genkey(c echo.Context) error {
	var token string
	err := database.DB.QueryRow("INSERT INTO pending_donations(amount) VALUES(0) RETURNING token").Scan(&token)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, token)
}
