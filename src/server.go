package main

import (
	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/s3proxy"
	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
	"os"
	"strconv"
)

var port = 3000

func init() {
	// Check if $PORT has been set to an int
	if p, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		port = p
	}
}

func main() {
	hosts := map[string]*echo.Echo{
		"impactclient.net":       web.Server(),
		"files.impactclient.net": s3proxy.Server(),
	}

	e := echo.New()

	// Enforce URL style
	// We don't need to do any http->https stuff here 'cos cloudflare
	e.Pre(middleware.NonWWWRedirect())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(mid.RemoveIndexHTML(http.StatusMovedPermanently))

	e.Any("/*", func(c echo.Context) (err error) {
		req := c.Request()
		res := c.Response()
		host := hosts[req.Host]

		if host == nil {
			err = echo.ErrNotFound
		} else {
			host.ServeHTTP(res, req)
		}

		return
	})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Start the server
	e.Logger.Fatal(StartServer(e, port))
}

func StartServer(e *echo.Echo, port int) error {
	return e.Start(":" + strconv.Itoa(port))
}
