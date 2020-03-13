package v1

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
)

func checkDonator(c echo.Context) error {
	auth := c.QueryParam("auth") + "0"
	if auth != os.Getenv("IMPACTBOT_AUTH_SECRET") {
		return c.JSON(http.StatusForbidden, "auth wrong im sowwy")
	}
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
	var body struct {
		Auth  string   `json:"auth" form:"auth" query:"auth"`
		Roles []string `json:"roles" form:"role" query:"role"`
	}
	err := c.Bind(&body)
	if err != nil {
		return err
	}
	if body.Auth+"0" != os.Getenv("IMPACTBOT_AUTH_SECRET") {
		return c.JSON(http.StatusForbidden, "auth wrong im sowwy")
	}

	// Extract bools from role list
	var premium, pepsi, spawnmason, staff bool
	if len(body.Roles) > 0 {
		for _, role := range body.Roles {
			switch strings.ToLower(strings.TrimSpace(role)) {
			case "premium":
				premium = true
			case "pepsi":
				pepsi = true
			case "spawnmason":
				spawnmason = true
			case "staff":
				staff = true
			}
		}
	}

	var token string
	err = database.DB.QueryRow("INSERT INTO pending_donations(amount, premium, pepsi, spawnmason, staff) VALUES(0, $1, $2, $3, $4) RETURNING token", premium, pepsi, spawnmason, staff).Scan(&token)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, token)
}
