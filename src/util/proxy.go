package util

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
)

// var func to allow overriding in tests
var serveProxy = func(proxy *httputil.ReverseProxy, req *http.Request, res http.ResponseWriter) {
	proxy.ServeHTTP(res, req)
}

func Proxy(c echo.Context, target *url.URL) {
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
