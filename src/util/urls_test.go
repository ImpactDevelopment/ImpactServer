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
