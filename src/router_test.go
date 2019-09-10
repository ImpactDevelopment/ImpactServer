package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/web"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestRouter(t *testing.T) {
	e := echo.New()
	Router(e)

	// Test the handlers are assigned to the right paths
	assert.Equal(t, "/changelog", e.Reverse(getName(web.Changelog)))
	assert.Equal(t, "/Impact/*", e.Reverse(getName(web.ImpactRedirect)))
}

// copied from echo.handlerName
func getName(h echo.HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}
