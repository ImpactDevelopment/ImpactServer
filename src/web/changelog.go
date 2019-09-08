package web

import (
	"github.com/labstack/echo"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const github = "https://impactdevelopment.github.io/Impact/"

func Changelog(c echo.Context) error {
	// Forward to the changelog hosted by github

	target, err := url.Parse(github + "changelog")
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
	// 302 non-caching redirect
	return c.Redirect(http.StatusFound, "https://impactdevelopment.github.io"+c.Request().RequestURI)
}
