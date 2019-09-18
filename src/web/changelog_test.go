package web

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestConst(t *testing.T) {
	assert.Equal(t, "https://impactdevelopment.github.io", github)
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
	req := httptest.NewRequest(http.MethodGet, "http://foobar.host/changelog", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/changelog")

	// Run the handler
	err := Changelog(c)
	if assert.NoError(t, err) {
		// Basic checks
		assert.Equal(t, 1, servedCount)
		assert.NotNil(t, servedProxy)
		assert.NotNil(t, servedReq)
		assert.NotNil(t, servedRes)

		// Request should be unchanged
		assert.Equal(t, "", servedReq.Header.Get("X-Forwarded-Host"))
		assert.Equal(t, "foobar.host", servedReq.Host)
		assert.Equal(t, "foobar.host", servedReq.URL.Host)
		assert.Equal(t, "/changelog", servedReq.URL.Path)
		assert.Equal(t, "", servedReq.URL.RawQuery)
		assert.Equal(t, "http://foobar.host/changelog", servedReq.URL.String())

		// The Director function should mutate the request
		servedProxy.Director(servedReq)
		assert.Equal(t, "foobar.host", servedReq.Header.Get("X-Forwarded-Host"))
		assert.Equal(t, "impactdevelopment.github.io", servedReq.Host)
		assert.Equal(t, "https", servedReq.URL.Scheme)
		assert.Equal(t, "impactdevelopment.github.io", servedReq.URL.Host)
		assert.Equal(t, "/Impact/changelog", servedReq.URL.Path)
		assert.Equal(t, "", servedReq.URL.RawQuery)
		assert.Equal(t, github+"/Impact/changelog", servedReq.URL.String())
	}
}

func TestImpactRedirect(t *testing.T) {
	const route = "/Impact/"
	const path = "assets/css/style.css?v=foobar"

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "http://foobar.cool"+route+path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(route + "*")
	err := ImpactRedirect(c)

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
	err := ImpactRedirect(c)

	if assert.NoError(t, err) {
		// Expect 301
		assert.Equal(t, http.StatusMovedPermanently, rec.Code)
		// Expect the correct target
		assert.Equal(t, "http://foobar.cool/changelog", rec.Header().Get(echo.HeaderLocation))
	}
}
