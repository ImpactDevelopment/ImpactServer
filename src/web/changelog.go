package web

import (
	"net/http"
	"net/url"

	"github.com/ImpactDevelopment/ImpactServer/src/util"

	"github.com/labstack/echo"
)

const github = "https://impactdevelopment.github.io"

func changelog(c echo.Context) error {
	// Forward to the changelog hosted by github

	target, err := url.Parse(github + "/Impact/changelog")
	if err != nil {
		return err //wtf
	}
	util.Proxy(c, target)
	return nil
}

func impactRedirect(c echo.Context) error {
	address := c.Request().URL

	// Echo tends to set the Request URL to just the path+query
	if address.Host == "" {
		address.Host = c.Request().Host
	}
	if address.Scheme == "" {
		address.Scheme = c.Scheme()
	}

	// Special case: 301 /Impact/changelog → /changelog
	if address.Path == "/Impact/changelog" {
		address.Path = "/changelog"
		return c.Redirect(http.StatusMovedPermanently, address.String())
	}

	// Pull the bits we need from the github url
	ghAddr, err := url.Parse(github)
	if err != nil {
		return err
	}
	address.Scheme = ghAddr.Scheme
	address.Host = ghAddr.Host

	// 302 to github.io
	return c.Redirect(http.StatusFound, address.String())
}
