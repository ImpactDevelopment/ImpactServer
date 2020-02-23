package v1

import (
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/recaptcha"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

func resetPassword(c echo.Context) error {
	email := strings.TrimSpace(c.Param("email"))
	if email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err := recaptcha.Verify(c)
	if err != nil {
		return err
	}

	user := database.LookupUserByEmail(email)
	if user == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user not found")
	}

	token, err := getToken(user.ID)
	if err != nil {
		return err
	}

	// TODO send user an email, don't just give anyone a token!!
	//return c.JSONBlob(http.StatusOK, []byte(`{"message":"success"}`))
	// FIXME DO NOT MERGE WITH THIS AS IS!!!!!!
	return c.JSON(http.StatusOK, struct {
		Token string
	}{
		Token: token.String(),
	})
}

func putPassword(c echo.Context) error {
	var body struct {
		Password string `json:"password" form:"password" query:"password"`
	}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	if user, ok := c.Get("user").(*users.User); ok {
		// We are authenticated so trust the user
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "no user found")
		}
		err = setPassword(user.ID, body.Password)
		if err != nil {
			return err
		}
		return c.JSONBlob(http.StatusOK, []byte(`{"message":"success"}`))
	} else {
		// We are not authenticated... They should provide a reset token
		token := strings.TrimSpace(c.Param("token"))
		var userID uuid.UUID
		err = database.DB.QueryRow(`SELECT user_id FROM password_resets WHERE token = $1`, token).Scan(&userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid reset token").SetInternal(err)
		}

		return c.String(http.StatusOK, token)
	}
}

func getToken(userID uuid.UUID) (token uuid.UUID, err error) {
	err = database.DB.QueryRow(`INSERT INTO password_resets (user_id) VALUES ($1) RETURNING token`).Scan(&token)
	if err != nil {
		err = echo.NewHTTPError(http.StatusInternalServerError, "failed to create reset token").SetInternal(err)
	}
	return
}

func setPassword(userID uuid.UUID, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password").SetInternal(err)
	}

	// Set the new hash
	result, err := database.DB.Exec(`UPDATE users SET password_hash = $2 WHERE user_id = $1`, userID, hashedPassword)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to update password hash")
	}

	// Check it set correctly
	rows, err := result.RowsAffected()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "DB driver doesn't support RowsAffected()")
	}
	if rows != 1 {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Incorrect number of users affected: %d", rows))
	}

	return nil
}
