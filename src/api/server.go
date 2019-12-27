package api

import (
	"github.com/ImpactDevelopment/ImpactServer/src/api/v1"
	"github.com/labstack/echo/v4"
)

// Server returns an echo server that handles api requests for each version
func Server() (e *echo.Echo) {
	e = echo.New()

	v1.API(e.Group("/v1"))

	return
}
