package v1

import (
	"github.com/labstack/echo"
)

// TODO API Doc
func API(api *echo.Group) {
	api.GET("/motd", getMotd)
	api.GET("/minecraft/user/info", getUserInfo)
}
