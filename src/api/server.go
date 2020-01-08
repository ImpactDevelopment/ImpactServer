package api

import (
	v1 "github.com/ImpactDevelopment/ImpactServer/src/api/v1"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server returns an echo server that handles api requests for each version
func Server() (e *echo.Echo) {
	e = echo.New()

	// Allow browser clients to use the API
	e.Use(middleware.CORS())

	v1.API(e.Group("/v1"))

	return
}
