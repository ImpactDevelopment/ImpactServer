package web

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConst(t *testing.T) {
	assert.Equal(t, "https://impactdevelopment.github.io/Impact/", github)
}

func TestChangelog(t *testing.T) {
	// Should proxy to github
	//e := echo.New()
	//req := httptest.NewRequest(http.MethodGet, "/changelog")
}

func TestImpactRedirect(t *testing.T) {
	const route = "/Impact/"
	const path = "assets/css/style.css?v=foobar"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, route+path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(route + "*")
	err := ImpactRedirect(c)

	if assert.NoError(t, err) {
		// Expect 302
		assert.Equal(t, http.StatusFound, rec.Code)
		// Expect the correct target
		assert.Equal(t, github+path, rec.Header().Get(echo.HeaderLocation))
	}
}
