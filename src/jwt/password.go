package jwt

import (
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type passwordRequest struct {
	Email    string `json:"email" form:"email" query:"email"`
	Password string `json:"password" form:"password" query:"password"`
}

func PasswordLoginHandler(c echo.Context) error {
	var body passwordRequest
	if err := c.Bind(&body); err != nil {
		return err
	}
	if body.Email == "" || body.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password must both be provided")
	}

	// Get the user
	user := database.LookupUserByEmail(strings.TrimSpace(body.Email))
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no user found")
	}

	// Check the password
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password))
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "incorrect password")
	}

	return respondWithToken(user, c)
}
