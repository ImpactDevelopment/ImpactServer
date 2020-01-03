package util

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetQuery(t *testing.T) {
	url1, err := url.Parse("http://example.com/foo?that=thing")
	assert.NoError(t, err)

	assert.Equal(t, "thing", url1.Query().Get("that"))
	SetQuery(url1, "that", "random thing")
	assert.Equal(t, "random thing", url1.Query().Get("that"))
	assert.Equal(t, "http://example.com/foo?that=random+thing", url1.String())
}

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
