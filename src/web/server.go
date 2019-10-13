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
	e.GET("/references.json", references, mid.CacheUntilPurge())

	e.GET("/ImpactInstaller.jar", installerForJar, mid.NoCache())
	e.GET("/ImpactInstaller.exe", installerForExe, mid.NoCache())

	staticEcho := echo.New()
	staticEcho.Use(mid.CacheUntilRestart(604800)) // 1 week
	staticEcho.Static("/", "static")
	e.Any("/*", func(c echo.Context) error {
		staticEcho.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	return
}
