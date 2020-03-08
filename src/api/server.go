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
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${status} ${method} ${uri} latency=${latency_human} error=${error}\n",
	}))

	v1.API(e.Group("/v1"))

	return
}
