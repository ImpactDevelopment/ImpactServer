package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo"
)

// API configures the Group to implement v1 of the API
func API(api *echo.Group) {
	// TODO API Doc

	api.GET("/motd", getMotd, middleware.Cache(60*30))
	api.GET("/minecraft/user/info", getUserInfo, middleware.Cache(60*30))
}
