package lib

import "github.com/labstack/echo"

// Implemented by echo.Echo
type HttpServer interface {
	Pre(middleware ...echo.MiddlewareFunc)
	Use(middleware ...echo.MiddlewareFunc)
	Start(address string) error

	// HTTP methods
	Group(prefix string, m ...echo.MiddlewareFunc) (g *echo.Group)
	Any(path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
	Match(methods []string, path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
}
