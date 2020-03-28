package middleware

import (
	"github.com/labstack/echo/v4"
	"strings"
)

// StripExt redirects any request ending with a provided extension to one without it, using the supplied http status code
func StripExt(code int, extension ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := req.URL.Path

			// Redirect
			for _, ext := range extension {
				if strings.HasSuffix(path, "."+ext) {
					url := *req.URL
					if url.Host == "" {
						url.Host = req.Host
					}
					if url.Scheme == "" {
						url.Scheme = c.Scheme()
					}
					url.Path = path[:len(path)-len(ext)-1]
					return c.Redirect(code, url.String())
				}
			}
			return next(c)
		}
	}
}
