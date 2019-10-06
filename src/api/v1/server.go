package v1

import (
	"github.com/labstack/echo"
)

// TODO API Doc
func API(api *echo.Group) {
	api.GET("/motd", getMotd)
	//api.GET("/version/:project/:version", getVersionInfo) // Get info about mc or impact versions, e.g. which versions of the former the latter supports.
	api.GET("/minecraft/user/info", getUserInfo)
	//api.GET("/minecraft/user/legacy/:type", legacyMCUserInfo) //DEPRECATED: Get a list of UUIDs for the given type
}