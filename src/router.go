package main

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"github.com/labstack/echo"
)

func Router(e *echo.Echo) {
	e.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", web.Changelog)
	e.GET("/s3proxy/:file", web.S3Proxy)
	e.Any("/Impact/*", web.ImpactRedirect)
}
