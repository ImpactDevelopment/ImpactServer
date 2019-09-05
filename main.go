package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"os"
)

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

	// Compress responses
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))

	// Log all the things TODO formatting https://echo.labstack.com/middleware/logger
	e.Use(middleware.Logger())

	// Don't crash
	e.Use(middleware.Recover())

	// Get the specified port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start the server
	e.Logger.Fatal(e.Start(":" + port))
}
