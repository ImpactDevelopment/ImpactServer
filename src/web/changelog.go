package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo"
)

const github = "https://impactdevelopment.github.io"

func Changelog(c echo.Context) error {
	// Forward to the changelog hosted by github

	target, err := url.Parse(github + "/Impact/changelog")
	if err != nil {
		return err //wtf
	}
	doProxy(c, target)
	return nil
}

func doProxy(c echo.Context, target *url.URL) {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Change the URL
			req.URL = target
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Host = target.Host

			// Don't send our cookies to github
			req.Header.Del(echo.HeaderCookie)
			req.Header.Del(echo.HeaderAuthorization)
		},
	}

	serveProxy(proxy, c.Request(), c.Response())
}

// var func to allow overriding in tests
var serveProxy = func(proxy *httputil.ReverseProxy, req *http.Request, res http.ResponseWriter) {
	proxy.ServeHTTP(res, req)
}

func ImpactRedirect(c echo.Context) error {
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
