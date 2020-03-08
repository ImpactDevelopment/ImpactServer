package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/ImpactDevelopment/ImpactServer/src/api"

	"github.com/ImpactDevelopment/ImpactServer/src/cloudflare"
	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/ImpactDevelopment/ImpactServer/src/newWeb"
	"github.com/ImpactDevelopment/ImpactServer/src/s3proxy"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
		"new":   newWeb.Server(),
		"files": s3proxy.Server(),
		"api":   api.Server(),
	}

	e := echo.New()

	// Enforce URL style
	e.Pre(middleware.NonWWWRedirect())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(mid.RemoveIndexHTML(http.StatusMovedPermanently))
	e.Pre(mid.EnforceHTTPS(http.StatusMovedPermanently))

	e.Use(middleware.BodyLimit("1M"))

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

	e.Use(middleware.Recover())

	go cloudflare.PurgeIfNeeded() // "go" as a vague halfhearted attempt to make this occur only after we start listening and serving, to prevent long blocking requests
	// Start the server
	e.Logger.Fatal(StartServer(e, port))
}

func StartServer(e *echo.Echo, port int) error {
	return e.Start(":" + strconv.Itoa(port))
}
