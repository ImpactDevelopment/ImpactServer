package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo"
)

// TODO API Doc
func API(api *echo.Group) {
	api.GET("/motd", getMotd)
	api.GET("/minecraft/user/info", getUserInfo)

	// Cache everything, at least for now
	api.Use(middleware.Cache(60 * 10))
}
