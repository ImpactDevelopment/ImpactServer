package v1

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/database"
	"github.com/ImpactDevelopment/ImpactServer/src/mailgun"
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/recaptcha"
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
	var body struct {
		Email string `json:"email" form:"email" query:"email"`
	}
	err := c.Bind(&body)
	if err != nil {
		return err
	}
	if body.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email is required")
	}

	err = recaptcha.Verify(c)
	if err != nil {
		return err
	}

	user := database.LookupUserByEmail(strings.TrimSpace(body.Email))
	if user == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user not found")
	}

	token, err := genToken(user.ID)
	if err != nil {
		return err
	}
	resetURL := util.GetServerURL()
	resetURL.Path = "/forgotpassword.html"
	resetURL.RawQuery = url.Values{"token": {token.String()}}.Encode()

	// Send user an email, don't just give anyone a token lol
	message := mailgun.MG.NewMessage("Impact <noreply@impactclient.net>", "Password reset", fmt.Sprintf(text, resetURL.String()), user.Email)
	message.SetHtml(fmt.Sprintf(html, resetURL.String(), resetURL.String()))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, err = mailgun.MG.Send(ctx, message)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to send reset email")
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{"success"})
}

func putPassword(c echo.Context) error {
	var body struct {
		Password string `json:"password" form:"password" query:"password"`
	}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	if user := middleware.GetUser(c); user != nil {
		// We are authenticated so trust the user
		err = setPassword(user.ID, body.Password)
		if err != nil {
			return err
		}
	} else {
		// We are not authenticated... They should provide a reset token
		token := strings.TrimSpace(c.Param("token"))
		var userID uuid.UUID
		var createdAt int64
		err = database.DB.QueryRow(`DELETE FROM password_resets WHERE token = $1 RETURNING user_id, created_at`, token).Scan(&userID, &createdAt)

		// Check the token actually exists
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "invalid reset token").SetInternal(err)
		}

		// If the error  is some other error, that probably means we failed to delete the token
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "unable to delete reset token").SetInternal(err)
		}

		// Also check when the token was created: should be in the past, but not too far in the past...
		if now, then := time.Now(), time.Unix(createdAt, 0); now.Before(then) || now.After(then.Add(24*time.Hour)) {
			return echo.NewHTTPError(http.StatusUnauthorized, "expired reset token")
		}

		// OK, valid token so we can trust them now, I guess
		err = setPassword(userID, body.Password)
		if err != nil {
			return err
		}
	}
	return c.JSON(http.StatusOK, struct {
		Message string `json:"message"`
	}{"success"})
}

func genToken(userID uuid.UUID) (token uuid.UUID, err error) {
	err = database.DB.QueryRow(`INSERT INTO password_resets (user_id) VALUES ($1) RETURNING token`, userID).Scan(&token)
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
