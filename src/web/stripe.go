package web

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// Redirects the old testing URI /stripe from PR#53 to /donate
func stripe(c echo.Context) error {
	address := c.Request().URL

	// Echo tends to set the Request URL to just the path+query
	if address.Host == "" {
		address.Host = c.Request().Host
	}
	if address.Scheme == "" {
		address.Scheme = c.Scheme()
	}

	// 301 /stripe → /donate
	address.Path = "/donate"
	return c.Redirect(http.StatusMovedPermanently, address.String())
}
