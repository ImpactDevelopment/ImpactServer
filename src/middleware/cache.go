package middleware

import (
	"strconv"

	"github.com/labstack/echo"
)

func Cache(maxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
			return next(c)
		}
	}
}

func CacheUntilRestart(browserMaxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// https://stackoverflow.com/questions/7071763/max-value-for-cache-control-header-in-http/25201898
			c.Response().Header().Set("Cache-Control", "public, s-maxage=31536000, max-age="+strconv.Itoa(browserMaxAge))
			// not a typo! s-maxage really has no hyphen between max and age, while max-age does!
			return next(c)
		}
	}
}

func NoCache() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "private, max-age=0")
			return next(c)
		}
	}
}
