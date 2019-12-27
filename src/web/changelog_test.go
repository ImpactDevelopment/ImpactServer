package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestConst(t *testing.T) {
	assert.Equal(t, "https://impactdevelopment.github.io", github)
}

func TestImpactRedirect(t *testing.T) {
	const route = "/Impact/"
	const path = "assets/css/style.css?v=foobar"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://foobar.cool"+route+path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(route + "*")
	err := impactRedirect(c)

	if assert.NoError(t, err) {
		// Expect 302
		assert.Equal(t, http.StatusFound, rec.Code)
		// Expect the correct target
		assert.Equal(t, github+route+path, rec.Header().Get(echo.HeaderLocation))
	}
}

func TestChangelogRedirect(t *testing.T) {
	const route = "/Impact/"
	const path = "changelog"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://foobar.cool"+route+path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(route + "*")
	err := impactRedirect(c)

	if assert.NoError(t, err) {
		// Expect 301
		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		// Expect the correct target
		assert.Equal(t, "http://foobar.cool/changelog", rec.Header().Get(echo.HeaderLocation))
	}
}
