package util

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

// testRoundTripper implements http.RoundTripper
type testRoundTripper func(req *http.Request) *http.Response

func (f testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//testClient returns *http.Client with Transport replaced to avoid making real calls
func testClient(roundTripFunc testRoundTripper) *http.Client {
	return &http.Client{
		Transport: roundTripFunc,
	}
}

func TestGetRequest(t *testing.T) {
	// Override httpClient()
	httpClient = func() *http.Client {
		return testClient(func(req *http.Request) *http.Response {
			// Test request parameters
			assert.Equal(t, req.URL.String(), "http://example.com/some/path")
			assert.Equal(t, req.Method, http.MethodGet)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				Header:     make(http.Header), // Must be set to non-nil value or it panics
			}
		})
	}

	request, err := GetRequest("http://example.com/some/path")
	assert.NoError(t, err)

	response, err := request.Do()
	assert.NoError(t, err)

	assert.True(t, response.Ok())
	assert.Equal(t, 200, response.Code())
	assert.Equal(t, "200 OK", response.Status())

	body, err := response.String()
	assert.NoError(t, err)
	assert.Equal(t, "OK", body)

	assert.NoError(t, err)
	assert.Equal(t, "OK", body)
}
