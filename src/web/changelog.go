package web

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const github = "https://impactdevelopment.github.io"

func Changelog(c echo.Context) error {
	// Forward to the changelog hosted by github

	target, err := url.Parse(github + "/Impact/changelog")
	if err != nil {
		return err //wtf
	}

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
	return nil
}

// var func to allow overriding in tests
var serveProxy = func(proxy *httputil.ReverseProxy, req *http.Request, res http.ResponseWriter) {
	proxy.ServeHTTP(res, req)
}

func ImpactRedirect(c echo.Context) error {
	path := c.Request().URL.Path

	// Special case: 301 /Impact/changelog → /changelog
	if path == "/Impact/changelog" {
		return c.Redirect(http.StatusMovedPermanently, strings.Replace(strings.ToLower(c.Request().URL.String()), "/impact/", "/", 1))
	}

	// Redirect with the query string intact
	if query := c.Request().URL.RawQuery; query != "" {
		path += "?" + query
	}

	// 302 to github.io
	return c.Redirect(http.StatusFound, github+path)
}
