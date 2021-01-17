package middleware

import (
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"regexp"
)

const userCtxKey = "user"

var authBearerRegx = regexp.MustCompile(`^Bearer\s+(\S+)`)

// GetUser returns the User object attached to the context, presumably by the RequireAuth middleware.
// Otherwise it returns nil.
func GetUser(c echo.Context) (user *users.User) {
	// Try to cast to *user.User, ignore if it failed, user probably just isn't set
	user, _ = c.Get(userCtxKey).(*users.User)
	return
}

var Auth = func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if header := c.Request().Header.Get(echo.HeaderAuthorization); header != "" {
			if m := authBearerRegx.FindStringSubmatch(header); len(m) == 2 && m[1] != "" {
				token := m[1]

				// Verify the JWT
				user, err := jwt.Verify(token)
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token").SetInternal(err)
				}
				// Set the user context userCtxKey
				c.Set(userCtxKey, user)
			}
		}
		return next(c)
	}
}

// RequireAuth requires that the Authorization header be set, correctly formatted, and the token to be valid
// For invalid token, it sends “401 - Unauthorized” response.
// For missing or invalid Authorization header, it sends “400 - Bad Request”.
var RequireAuth = func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get(echo.HeaderAuthorization)
		if header == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "authentication is required")
		}
		if m := authBearerRegx.FindStringSubmatch(header); len(m) != 2 || m[1] == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid "+echo.HeaderAuthorization+" header")
		}
		return Auth(func(c echo.Context) error {
			// Require Auth to have succeeded
			if user := GetUser(c); user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "no user found")
			}
			return next(c)
		})(c)
	}
}

// RequireRoles returns a middleware that requires the user to have at least one of the provided role IDs.
// This middleware automatically calls RequireAuth
func RequireRole(roles ...string) echo.MiddlewareFunc {
	// Figure out how best to print the roles in http errors
	var rolesString string
	if len(roles) == 1 {
		rolesString = fmt.Sprintf("%v", roles[0])
	} else {
		rolesString = fmt.Sprintf("%v", roles)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return RequireAuth(func(c echo.Context) error {
			user := GetUser(c)
			if user == nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Roles required but no user: %v", rolesString))
			}

			// If any role matches, continue with request
			for _, role := range user.Roles {
				for _, required := range roles {
					if role.ID == required {
						return next(c)
					}
				}
			}

			// The user doesn't have any matching roles
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Required at least one role from %v", rolesString))
		})
	}
}

func AuthGetParam() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.QueryParam("auth") + "0"
			if auth != os.Getenv("API_AUTH_SECRET") {
				return c.JSON(http.StatusForbidden, "auth wrong im sowwy")
			}
			return next(c)
		}
	}
}
