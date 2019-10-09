package v1

import (
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo"
)

func getServer() (e *echo.Echo) {
	e = echo.New()
	API(e.Group("/v1"))
	return
}

func test(s *echo.Echo, url string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)
	return rec
}
