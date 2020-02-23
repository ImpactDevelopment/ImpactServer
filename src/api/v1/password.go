package v1

import (
	"context"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/mailgun"
	"github.com/ImpactDevelopment/ImpactServer/src/recaptcha"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Let's be real, I should be using text/template and html/template here instead of format strings
	text = `Here's your password reset link: %s`
	html = `<p>
<a href="%s">Click here to reset your password</a> or copy the following link if that doesn't work:
</p>
<pre>
%s
</pre>`
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
	resetUrl := util.GetServerURL()
	resetUrl.Path = "/forgotpassword.html"
	resetUrl.RawQuery = url.Values{"token": {token.String()}}.Encode()

	// Send user an email, don't just give anyone a token lol
	message := mailgun.MG.NewMessage("Impact <noreply@impactclient.net>", "Password reset", fmt.Sprintf(text, resetUrl.String()), user.Email)
	message.SetHtml(fmt.Sprintf(html, resetUrl.String(), resetUrl.String()))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, id, err := mailgun.MG.Send(ctx, message)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset email")
	}
	fmt.Printf("Successful password reset email sent to %s, mailgun id %s, resp %s", user.Email, id, resp)
	return c.JSONBlob(http.StatusOK, []byte(`{"message":"success"}`))
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
