package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSubdomain(t *testing.T) {
	// Get an empty string if no sub-domain
	assert.Equal(t, "", GetSubdomains("example.com"))
	assert.Equal(t, "", GetSubdomains("example.co.uk"))
	assert.Equal(t, "", GetSubdomains("localhost"))

	// Get a sub-domain if present
	assert.Equal(t, "foo", GetSubdomains("foo.bar.com"))
	assert.Equal(t, "foo", GetSubdomains("foo.bar.co.uk"))
	assert.Equal(t, "foo", GetSubdomains("foo.localhost"))

	// Support sub-sub domains
	assert.Equal(t, "foo.bar", GetSubdomains("foo.bar.example.com"))
	assert.Equal(t, "foo.bar", GetSubdomains("foo.bar.example.co.uk"))
	assert.Equal(t, "foo.bar", GetSubdomains("foo.bar.localhost"))

	// Don't break if a port is included
	assert.Equal(t, "", GetSubdomains("example.com:3000"))
	assert.Equal(t, "", GetSubdomains("localhost:321"))
	assert.Equal(t, "abc", GetSubdomains("abc.example.com:3000"))
	assert.Equal(t, "abc", GetSubdomains("abc.localhost:321"))
	assert.Equal(t, "abc.def", GetSubdomains("abc.def.example.com:3000"))
	assert.Equal(t, "abc.def", GetSubdomains("abc.def.localhost:321"))
}
