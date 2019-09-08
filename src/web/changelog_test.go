package web

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func TestConst(t *testing.T) {
	assert.Equal(t, "https://impactdevelopment.github.io/Impact/", github)
}

func TestChangelog(t *testing.T) {
	// Override serveProxy and store what's passed into it
	var (
		servedCount = 0
		servedProxy *httputil.ReverseProxy
		servedReq   *http.Request
		servedRes   http.ResponseWriter
	)
	serveProxy = func(proxy *httputil.ReverseProxy, req *http.Request, res http.ResponseWriter) {
		servedCount++
		servedProxy = proxy
		servedReq = req
		servedRes = res
	}

	// Setup the request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/changelog", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/changelog")

	// Run the handler
	err := Changelog(c)
	if assert.NoError(t, err) {
		assert.Equal(t, 1, servedCount)
		assert.NotNil(t, servedProxy)
		assert.NotNil(t, servedReq)
		assert.NotNil(t, servedRes)
	}

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
