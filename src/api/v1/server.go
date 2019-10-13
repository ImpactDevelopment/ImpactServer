package v1

import (
	"github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo"
)

// API configures the Group to implement v1 of the API
func API(api *echo.Group) {
	// TODO API Doc

	api.GET("/motd", getMotd, middleware.CacheUntilPurge())
	api.GET("/themes", getThemes, middleware.CacheUntilPurge())
	api.GET("/minecraft/user/info", getUserInfo, middleware.Cache(3600))
	api.GET("/dbtest", dbTest, middleware.NoCache())
	api.GET("/minecraft/login", mojangLogin, middleware.NoCache())
	api.GET("/emailtest", emailTest, middleware.NoCache())
}
