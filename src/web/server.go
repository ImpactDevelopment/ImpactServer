package web

import (
	"net/http"

	"github.com/labstack/echo/middleware"

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

	staticEcho := echo.New()
	staticEcho.Use(mid.Cache(86400))
	staticEcho.Static("/", "static")
	e.Any("/*", func(c echo.Context) error {
		staticEcho.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return
}
