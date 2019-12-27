package util

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
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

	target, _ := url.Parse("https://impactdevelopment.github.io/Impact/changelog")

	// Run the handler
	Proxy(c, target)
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
	assert.Equal(t, target.String(), servedReq.URL.String())
}
