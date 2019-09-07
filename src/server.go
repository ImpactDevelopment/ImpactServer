package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/lib"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
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
	// Echo is cool https://echo.labstack.com
	server := echo.New()
	AddMiddleware(server)

	// Start the server
	server.Logger.Fatal(StartServer(server, port))
}

func AddMiddleware(s lib.HttpServer) {
	// Enforce URL style
	// We don't need to do any http->https stuff here 'cos cloudflare
	s.Pre(middleware.NonWWWRedirect())
	s.Pre(middleware.RemoveTrailingSlash())

	// Fall back to static files
	s.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:   "static", // Root directory from where the static content is served.
		Browse: false,    // Don't enable directory browsing.
		HTML5:  false,    // Don't forward everything to root (would allow client-side routing)
	}))

	// Compression not required because the CDN does that for us

	// Log all the things TODO formatting https://echo.labstack.com/middleware/logger
	s.Use(middleware.Logger())

	// Don't crash
	s.Use(middleware.Recover())
}

func StartServer(s lib.HttpServer, port int) error {
	return s.Start(":" + strconv.FormatInt(int64(port), 10))
}
