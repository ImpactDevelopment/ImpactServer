package middleware

import (
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

// RequireAuth returns a middleware that sets the user key in context and calls next handler.
// For invalid token, it sends “401 - Unauthorized” response.
// For missing or invalid Authorization header, it sends “400 - Bad Request”.
func RequireAuth() echo.MiddlewareFunc {
	// Use a KeyAuth instead of a JWTAuth since we want to verify the JWT ourselves
	return middleware.KeyAuth(func(token string, c echo.Context) (bool, error) {
		// Verify the JWT
		user, err := jwt.Verify(token)
		if err != nil {
			return false, err
		}
		// Set the user context key
		c.Set("user", user)
		return true, nil
	})
}

// RequireRoles returns a middleware that requires the user to have at least one of the provided role IDs.
// This middleware must come after RequireAuth
func RequireRole(roles ...string) echo.MiddlewareFunc {
	// Figure out how best to print the roles in http errors
	var rolesString string
	if len(roles) == 1 {
		rolesString = fmt.Sprintf("%v", roles[0])
	} else {
		rolesString = fmt.Sprintf("%v", roles)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if user, ok := c.Get("user").(*users.User); ok {
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
			} else {
				// Can't cast c.Get("user") from interface{} to *User
				return echo.NewHTTPError(http.StatusInternalServerError, "Role required but cannot get user from context")
			}
		}
	}
}
