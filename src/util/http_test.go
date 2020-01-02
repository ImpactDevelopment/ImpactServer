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

// testResponse reduces some of the boilerplate needed in a testClient's testRoundTripper callback
func testResponse(status int, body string, headers map[string]string) *http.Response {
	h := http.Header{}
	if headers != nil {
		for key, value := range headers {
			h.Set(key, value)
		}
	}

	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
		Header:     h,
	}
}

func TestGetRequest(t *testing.T) {
	request, err := GetRequest("http://example.com/some/path")
	assert.NoError(t, err)

	// Override client to avoid actually sending request
	request.client = testClient(func(req *http.Request) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "http://example.com/some/path")
		assert.Equal(t, req.Method, http.MethodGet)

		return testResponse(http.StatusOK, "OK", nil)
	})

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
