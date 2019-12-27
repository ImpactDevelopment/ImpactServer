package newWeb

import (
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"

	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
)

func Server() (e *echo.Echo) {
	e = echo.New()

	e.GET("/ImpactInstaller.*", redirect(http.StatusFound, "https://impactclient.net/"), mid.NoCache())
	e.Any("/*", proxy("https://impact-web.herokuapp.com/"))

	return
}

func proxy(address string) func(echo.Context) error {
	return func(c echo.Context) error {
		addr := c.Request().URL

		// Pull the bits we need from the heroku addr
		newAddr, err := url.Parse(address)
		if err != nil {
			return err
		}
		addr.Scheme = newAddr.Scheme
		addr.Host = newAddr.Host

		// Proxy to heroku
		util.Proxy(c, addr)
		return nil
	}
}

func redirect(code int, address string) func(echo.Context) error {
	return func(c echo.Context) error {
		addr := c.Request().URL

		// Pull the bits we need from the address
		newAddr, err := url.Parse(address)
		if err != nil {
			return err
		}
		addr.Scheme = newAddr.Scheme
		addr.Host = newAddr.Host

		// 302 to the current location
		return c.Redirect(code, addr.String())
	}
}
