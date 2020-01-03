package util

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

// testRoundTripper implements http.RoundTripper
type testRoundTripper func(req *http.Request, body string) *http.Response

func (f testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read body so we can pass it on as a string
	defer req.Body.Close()
	body, _ := ioutil.ReadAll(req.Body)

	return f(req, string(body)), nil
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

type (
	testStruct1 struct {
		It string `json:"it" xml:"it" form:"it"`
	}
	testStruct2 struct {
		That string `json:"that" xml:"that" form:"that"`
	}
)

func TestGet(t *testing.T) {
	request, err := GetRequest("http://example.com/some/path")
	assert.NoError(t, err)

	// Override client to avoid actually sending request
	request.client = testClient(func(req *http.Request, body string) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "http://example.com/some/path")
		assert.Equal(t, req.Method, http.MethodGet)
		assert.Empty(t, body)

		assert.Equal(t, "ImpactServer", req.UserAgent())
		assert.Equal(t, "ImpactServer", req.Header.Get("User-Agent"))

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

func TestJSON(t *testing.T) {
	reqBody := &testStruct1{"Hello, world"}
	request, err := JSONRequest("http://example.com/some/other/path", reqBody)
	assert.NoError(t, err)

	// Override client to avoid actually sending request
	request.client = testClient(func(req *http.Request, body string) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "http://example.com/some/other/path")
		assert.Equal(t, req.Method, http.MethodPost)
		assert.Equal(t, `{"it":"Hello, world"}`, body)

		assert.Equal(t, "ImpactServer", req.UserAgent())
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		length, err := strconv.Atoi(req.Header.Get("Content-Length"))
		assert.NoError(t, err)
		assert.Equal(t, len(body), length)

		return testResponse(http.StatusOK, `{"that":"thing"}`, nil)
	})

	response, err := request.Do()
	assert.NoError(t, err)

	assert.True(t, response.Ok())
	assert.Equal(t, 200, response.Code())
	assert.Equal(t, "200 OK", response.Status())

	body := &testStruct2{}
	err = response.JSON(body)
	assert.NoError(t, err)
	assert.Equal(t, &testStruct2{"thing"}, body)
	assert.Equal(t, "thing", body.That)
}

func TestXML(t *testing.T) {
	reqBody := &testStruct1{"Hello, world"}
	request, err := XMLRequest("http://example.com/some/other/path", reqBody)
	assert.NoError(t, err)

	// Override client to avoid actually sending request
	request.client = testClient(func(req *http.Request, body string) *http.Response {
		// Test request parameters
		assert.Equal(t, req.URL.String(), "http://example.com/some/other/path")
		assert.Equal(t, req.Method, http.MethodPost)
		assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+`<testStruct1><it>Hello, world</it></testStruct1>`, body)

		assert.Equal(t, "ImpactServer", req.UserAgent())
		assert.Equal(t, "application/xml", req.Header.Get("Content-Type"))
		length, err := strconv.Atoi(req.Header.Get("Content-Length"))
		assert.NoError(t, err)
		assert.Equal(t, len(body), length)

		return testResponse(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+`<testStruct2><that>thing</that></testStruct2>`, nil)
	})

	response, err := request.Do()
	assert.NoError(t, err)

	assert.True(t, response.Ok())
	assert.Equal(t, 200, response.Code())
	assert.Equal(t, "200 OK", response.Status())

	body := &testStruct2{}
	err = response.XML(body)
	assert.NoError(t, err)
	assert.Equal(t, &testStruct2{"thing"}, body)
	assert.Equal(t, "thing", body.That)
}
