package middleware

import (
	"github.com/labstack/echo"
	"strings"
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
