package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	// Helper function
	test := func(age int) *httptest.ResponseRecorder {
		e := echo.New()
		e.Use(Cache(age))
		e.Any("/*", func(c echo.Context) error {
			return c.String(http.StatusOK, "Ok cowboy")
		})
		req := httptest.NewRequest(http.MethodGet, "http://foobar.net/cached.meme/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	// Cache-Control header should equal max-age=[age]
	rec := test(200)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, max-age=200", rec.Header().Get("Cache-Control"))

	rec = test(150)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, max-age=150", rec.Header().Get("Cache-Control"))
}
