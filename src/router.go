package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"github.com/labstack/echo"
	"net/http"
)

func Router(e *echo.Echo) {
	e.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", web.Changelog)
	e.Any("/Impact/*", web.ImpactRedirect)
}
