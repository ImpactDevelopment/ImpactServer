package middleware

import (
	"strings"
	"fmt"

	"github.com/labstack/echo"
)

// Redirect trailing index.html's.
// code must be in the 300 range
func RemoveIndexHTML(code int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Copy URL struct
			address := c.Request().URL
			slice := deleteEmpty(strings.Split(address.Path, "/"))

			// If last path element is index.html
			if i := len(slice) - 1; i >= 0 && strings.ToLower(slice[i]) == "index.html" {
				// re-build the path without the last element
				address.Path = strings.Join(slice[:i], "/")

				// Echo tends to set the Request URL to just the path+query
				if address.Host == "" {
					address.Host = c.Request().Host
				}
				if address.Scheme == "" {
					address.Scheme = c.Scheme()
				}

				// Redirect
				return c.Redirect(code, address.String())
			}
			return next(c)
		}
	}
}

// Remove empty elements from a slice
func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func EnforceHTTPS(code int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for name, headers := range c.Request().Header {
				for _, h := range headers {
					fmt.Printf("%v: %v\n", name, h)
				}
			}
			if c.Request().Header.Get("X-Forwarded-Proto") != "http" {
				// this header is set by cloudflare
				// it won't be set on localhost
				return next(c)
			}
			// it is http
			addr := c.Request().URL
			if addr.Path == "/releases.json" {
				// don't break 4.7 and 4.8 update checker
				return next(c)
			}
			addr.Scheme = "https"
			addr.Host = c.Request().Host
			//return c.Redirect(code, addr.String())
			return next(c)
		}
	}
}
