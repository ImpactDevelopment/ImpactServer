package middleware

import (
	"strconv"

	"github.com/labstack/echo"
)

// instruct the browser and cloudflare to cache for this amount of time
func Cache(maxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
			return next(c)
		}
	}
}

// override an existing max-age header (e.g. from github) with a s-maxage that only cloudflare will respect instead
func CacheCloudflare(cloudflareMaxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, s-maxage="+strconv.Itoa(cloudflareMaxAge)+", max-age="+strconv.Itoa(cloudflareMaxAge))
			return next(c)
		}
	}
}

// cache indefinitely in cloudflare (until this server restarts and purges cloudflare), and for the defined amount of time in the browser
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

// do not cache anywhere
func NoCache() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "private, max-age=0")
			return next(c)
		}
	}
}
