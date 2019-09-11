package lib

import "github.com/labstack/echo"

type MockServer struct {
	StartAddress  string
	PreMiddleware []echo.MiddlewareFunc
	UseMiddleware []echo.MiddlewareFunc
}

func (s *MockServer) Pre(middleware ...echo.MiddlewareFunc) {
	s.PreMiddleware = append(s.PreMiddleware, middleware...)
}

func (s *MockServer) Use(middleware ...echo.MiddlewareFunc) {
	s.UseMiddleware = append(s.UseMiddleware, middleware...)
}

func (s *MockServer) Start(address string) error {
	s.StartAddress = address
	return nil
}

// HTTP methods
func (s *MockServer) Group(prefix string, m ...echo.MiddlewareFunc) (g *echo.Group) {
	return nil //TODO
}

func (s *MockServer) Any(path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route {
	return nil //TODO
}

func (s *MockServer) Match(methods []string, path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route {
	return nil //TODO
}
