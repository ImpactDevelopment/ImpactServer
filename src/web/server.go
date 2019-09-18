package web

import (
	"github.com/labstack/echo/middleware"
	"net/http"

	"github.com/labstack/echo"
)

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Match([]string{http.MethodHead, http.MethodGet}, "/changelog", changelog)
	e.Any("/Impact/*", impactRedirect)
	e.Static("/", "static")

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return
}
