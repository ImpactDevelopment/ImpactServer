package web

import (
	"net/http"

	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo"
)

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", changelog)
	e.Any("/Impact/*", impactRedirect)
	e.GET("/releases.json", releases, mid.Cache(3600))

	e.GET("/ImpactInstaller.jar", installerForJar, mid.Cache(0))
	e.GET("/ImpactInstaller.exe", installerForExe, mid.Cache(0))

	staticEcho := echo.New()
	staticEcho.Use(mid.Cache(86400))
	staticEcho.Static("/", "static")
	e.Any("/*", func(c echo.Context) error {
		staticEcho.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	return
}
