package middleware

import (
	"strconv"

	"github.com/labstack/echo"
)

func Cache(maxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "max-age="+strconv.Itoa(maxAge))
			return next(c)
		}
	}
}
