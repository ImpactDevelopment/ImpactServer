package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestIndexToSlash(t *testing.T) {
	// Helper func to setup the test
	setup := func(code int) (e *echo.Echo) {
		e = echo.New()
		e.Pre(RemoveIndexHTML(code))
		e.Any("/*", func(c echo.Context) error {
			return c.String(http.StatusOK, "Ok cowboy")
		})
		return
	}

	// Helper function to run a request and return the response
	test := func(s *echo.Echo, url string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, req)
		return rec
	}

	e := setup(http.StatusMovedPermanently)

	// index.html/ should produce 301
	rec := test(e, "http://foobar.net/foo/bar/index.html/")
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "http://foobar.net/foo/bar", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "", rec.Body.String())

	// index.html should produce 301
	rec = test(e, "http://foobar.net/foo/bar/index.html")
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "http://foobar.net/foo/bar", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "", rec.Body.String())

	// / should not redirect
	rec = test(e, "http://foobar.net/foo/bar/")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "Ok cowboy", rec.Body.String())

	// neither should bare path
	rec = test(e, "http://foobar.net/foo/bar")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "Ok cowboy", rec.Body.String())

	// root/
	rec = test(e, "http://foobar.net/")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "Ok cowboy", rec.Body.String())

	// root
	rec = test(e, "http://foobar.net")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "Ok cowboy", rec.Body.String())

	// index.html/ should produce 301
	rec = test(e, "http://foobar.net/index.html/")
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "http://foobar.net", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "", rec.Body.String())

	// index.html should produce 301
	rec = test(e, "http://foobar.net/index.html")
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "http://foobar.net", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "", rec.Body.String())

	// Try with a different status code
	e = setup(http.StatusFound)
	rec = test(e, "http://foobar.net/index.html")
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "http://foobar.net", rec.Header().Get(echo.HeaderLocation))
	assert.Equal(t, "", rec.Body.String())
}
