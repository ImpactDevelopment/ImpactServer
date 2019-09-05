package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
)

var port string

func init() {
	// Get the specified port
	port = os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
}

func main() {
	// Echo is cool https://echo.labstack.com
	e := echo.New()

	// Enforce URL style
	// We don't need to do any http->https or www->root stuff here 'coi cloudflare
	e.Pre(middleware.NonWWWRedirect())
	e.Pre(middleware.RemoveTrailingSlash())

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

	// Start the server
	e.Logger.Fatal(e.Start(":" + port))
}
