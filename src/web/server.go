package web

import (
	"github.com/labstack/echo/middleware"
	"net/http"

	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo"
)

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", changelog)
	e.Any("/Impact/*", impactRedirect)
	e.GET("/prereleases.json", prereleases, mid.Cache(3600))

	e.GET("/ImpactInstaller.jar", installerForJar, mid.Cache(0))
	e.GET("/ImpactInstaller.exe", installerForExe, mid.Cache(0))

	// Use e.Group since e.Static doesn't allow setting middleware, see labstack/echo#1407
	_ = e.Group("/", middleware.Static("static"), mid.Cache(86400))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return
}
