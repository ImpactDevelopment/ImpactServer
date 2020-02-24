package web

import (
	"net/http"

	mid "github.com/ImpactDevelopment/ImpactServer/src/middleware"
	"github.com/labstack/echo/v4"
)

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", changelog)
	e.Any("/Impact/*", impactRedirect)
	e.GET("/releases.json", releases, mid.CacheUntilPurge())

	e.GET("/ImpactInstaller.jar", installerForJar, mid.NoCache())
	e.GET("/ImpactInstaller.exe", installerForExe, mid.NoCache())

	e.POST("/discordverify", discordVerify)
	e.POST("/recaptchaverify", simpleRecaptchaCheck)

	staticEcho := echo.New()
	staticEcho.Use(mid.CacheUntilRestart(3600)) // 1 hour
	staticEcho.Static("/", "static")
	e.Any("/*", func(c echo.Context) error {
		staticEcho.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	return
}
