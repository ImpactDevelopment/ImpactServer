package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/api"
	"net/http"
	"os"
	"strconv"

	"github.com/ImpactDevelopment/ImpactServer/src/cloudflare"
	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/s3proxy"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
		"":      web.Server(),
		"files": s3proxy.Server(),
		"api":   api.Server(),
	}

	e := echo.New()

	// Enforce URL style
	// We don't need to do any http->https stuff here 'cos cloudflare
	e.Pre(middleware.NonWWWRedirect())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(mid.RemoveIndexHTML(http.StatusMovedPermanently))

	e.Any("/*", func(c echo.Context) error {
		req := c.Request()
		res := c.Response()

		server := hosts[util.GetSubdomains(req.Host)]

		if server == nil {
			return echo.ErrNotFound
		}

		server.ServeHTTP(res, req)
		return nil
	})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	go cloudflare.Purge()
	// Start the server
	e.Logger.Fatal(StartServer(e, port))
}

func StartServer(e *echo.Echo, port int) error {
	return e.Start(":" + strconv.Itoa(port))
}
