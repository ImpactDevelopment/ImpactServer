package util

import (
	"github.com/labstack/echo/v4"
	"strings"
)

func RealIPIfUnambiguous(c echo.Context) string {
	xForwardedFor := getSplitSlice(c)
	if len(xForwardedFor) == 2 {
		// We are behind exactly two proxies (heroku and cloudflare)
		return strings.TrimSpace(xForwardedFor[0])
	}
	return ""
}

func RealIPBestGuess(c echo.Context) string {
	xForwardedFor := getSplitSlice(c)
	if l := len(xForwardedFor); l >= 2 {
		// We are behind two proxies (heroku and cloudflare) and the user has either proxied or lied in their header
		// Return the ip that cloudflare got
		return strings.TrimSpace(xForwardedFor[l-2])
	}
	// We probably aren't behind any proxies
	return c.Request().RemoteAddr
}

func getSplitSlice(c echo.Context) []string {
	return strings.Split(c.Request().Header.Get(echo.HeaderXForwardedFor), ",")
}
