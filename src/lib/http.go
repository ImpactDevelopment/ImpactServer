package lib

import "github.com/labstack/echo"

// Implemented by echo.Echo
type HttpServer interface {
	Pre(middleware ...echo.MiddlewareFunc)
	Use(middleware ...echo.MiddlewareFunc)
	Start(address string) error
}
