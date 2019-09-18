package main

import (
	"net/http"
	"os"
	"strconv"

	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
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
	// Echo is cool https://echo.labstack.com
	server := echo.New()
	AddMiddleware(server)
	Router(server)

	// Start the server
	server.Logger.Fatal(StartServer(server, port))
}

func AddMiddleware(e *echo.Echo) {
	// Enforce URL style
	// We don't need to do any http->https stuff here 'cos cloudflare
	e.Pre(middleware.NonWWWRedirect())
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(mid.RemoveIndexHTML(http.StatusMovedPermanently))

	// Fall back to static files
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "static", // Root directory from where the static content is served.
		Browse: false,    // Don't enable directory browsing.
		HTML5:  false,    // Don't forward everything to root (would allow client-side routing)
	}))

	// Compression not required because the CDN does that for us

	// Log all the things TODO formatting https://echo.labstack.com/middleware/logger
	e.Use(middleware.Logger())

	// Don't crash
	e.Use(middleware.Recover())
}

func StartServer(e *echo.Echo, port int) error {
	return e.Start(":" + strconv.Itoa(port))
}
