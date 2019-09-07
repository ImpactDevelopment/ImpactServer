package main

import (
	"github.com/ImpactDevelopment/ImpactServer/src/lib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerMiddleware(t *testing.T) {
	s := &lib.MockServer{}
	AddMiddleware(s)

	// TODO check the right middleware is added
	assert.Equal(t, 5, len(s.PreMiddleware)+len(s.UseMiddleware))
}

func TestServerStart(t *testing.T) {
	s := &lib.MockServer{}
	port := 100
	err := StartServer(s, port)
	assert.Nil(t, err)
	assert.Equal(t, ":"+string(port), s.StartAddress)
}
