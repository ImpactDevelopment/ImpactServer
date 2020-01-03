package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/ImpactDevelopment/ImpactServer/src/util/mediatype"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const userAgent = "ImpactServer"

// HTTPRequest wraps http.Request so that we can provide custom methods
type HTTPRequest struct {
	Req *http.Request

	// client is the http.Client which will do the request
	//
	// it is set here so it can be overridden by tests
	client *http.Client
}

// HTTPResponse wraps http.Response so that we can provide custom methods
type HTTPResponse struct {
	Resp *http.Response
}

// NewRequest wraps http.NewRequest but returns HTTPRequest instead of http.Request. The ImpactServer User Agent is automatically added
//
// You probably want to use one of its wrappers like GetRequest or JSONRequest instead
func NewRequest(method, url string, body io.Reader) (*HTTPRequest, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", userAgent)

	return &HTTPRequest{
		Req:    request,
		client: http.DefaultClient,
	}, nil
}

// GetRequest returns a HTTPRequest using method GET with no body
func GetRequest(address string) (*HTTPRequest, error) {
	return NewRequest(http.MethodGet, address, nil)
}

// JSONRequest returns a HTTPRequest using method POST with a JSON marshalled body
func JSONRequest(address string, body interface{}) (*HTTPRequest, error) {
	post, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	request, err := NewRequest(http.MethodPost, address, bytes.NewReader(post))
	if err != nil {
		return nil, err
	}

	request.setContentType(mediatype.JSON)
	request.setLength(len(post))

	return request, nil
}

// XMLRequest returns a HTTPRequest using method POST with an XML marshalled body
func XMLRequest(address string, body interface{}) (*HTTPRequest, error) {
	// Strip the trailing newline, since we add our own later.
	doctype := xml.Header[:len(xml.Header)-1]
	return XMLRequestWithDoctype(address, doctype, body)
}

// XMLRequestWithDoctype returns a HTTPRequest using method POST with an XML marshalled body with the specified doctype
func XMLRequestWithDoctype(address, doctype string, body interface{}) (*HTTPRequest, error) {
	postBody, err := xml.Marshal(body)
	if err != nil {
		return nil, err
	}
	post := bytes.Join([][]byte{[]byte(doctype), postBody}, []byte("\n"))

	request, err := NewRequest(http.MethodPost, address, bytes.NewReader(post))
	if err != nil {
		return nil, err
	}

	request.setContentType(mediatype.XML)
	request.setLength(len(post))

	return request, nil
}

// FormRequest returns a HTTPRequest using method POST with a x-www-form-urlencoded marshalled body
func FormRequest(address string, form map[string]string) (*HTTPRequest, error) {
	post := urlValues(form).Encode()

	request, err := NewRequest(http.MethodPost, address, strings.NewReader(post))
	if err != nil {
		return nil, err
	}

	request.setContentType(mediatype.Form)
	request.setLength(len(post))

	return request, nil
}

// URL returns the url.URL associated with this HTTPRequest
func (r HTTPRequest) URL() *url.URL {
	return r.Req.URL
}

// SetQuery sets a url query value on the HTTPRequest's URL
func (r *HTTPRequest) SetQuery(key, value string) {
	SetQuery(r.URL(), key, value)
}

// SetHeader sets a header on the HTTPRequest
func (r *HTTPRequest) SetHeader(key, value string) {
	r.Req.Header.Set(key, value)
}

func (r *HTTPRequest) setLength(length int) {
	r.SetHeader("Content-Length", strconv.Itoa(length))
}

func (r *HTTPRequest) setContentType(mediaType mediatype.MediaType) {
	r.SetHeader("Content-Type", mediaType.String())
}

// Accept sets the Accept header on the HTTPRequest to indicate what content-type we expect
func (r *HTTPRequest) Accept(mediaType mediatype.MediaType) {
	r.SetHeader("Accept", mediaType.String())
}

// Authorization sets the Authorization header on the HTTPRequest for token-based auth
// e.g. request.Authorization("Bearer", token)
func (r *HTTPRequest) Authorization(authType string, authKey string) {
	r.SetHeader("Authorization", authType+" "+authKey)
}

// Do does a request and returns the response, as a HTTPResponse
func (r *HTTPRequest) Do() (*HTTPResponse, error) {
	resp, err := r.client.Do(r.Req)
	return &HTTPResponse{resp}, err
}

// Ok returns true if the status code is "200 OK"
func (r HTTPResponse) Ok() bool {
	return r.Code() == http.StatusOK
}

// Code returns the status code as an int
func (r HTTPResponse) Code() int {
	return r.Resp.StatusCode
}

// Status returns the full status string, e.g. "400 Bad Request"
func (r HTTPResponse) Status() string {
	code := r.Code()
	return strconv.Itoa(code) + " " + http.StatusText(code)
}

// GetHeader returns the value of the given header key
func (r HTTPResponse) GetHeader(key string) string {
	return r.Resp.Header.Get(key)
}

// ContentType returns the MediaType of the response body, according to the Content-Type header
func (r HTTPResponse) ContentType() mediatype.MediaType {
	return mediatype.MediaType(r.GetHeader("Content-Type"))
}

// String returns the body as a string
func (r *HTTPResponse) String() (string, error) {
	defer r.Resp.Body.Close()

	str, err := ioutil.ReadAll(r.Resp.Body)
	return string(str), err
}

// JSON decodes the body into the provided interface{}
func (r *HTTPResponse) JSON(v interface{}) error {
	defer r.Resp.Body.Close()

	return json.NewDecoder(r.Resp.Body).Decode(v)
}

// XML decodes the body into the provided interface{}
func (r *HTTPResponse) XML(v interface{}) error {
	defer r.Resp.Body.Close()

	return xml.NewDecoder(r.Resp.Body).Decode(v)
}

// urlValues converts a map of strings to url Values for use in forms or query strings
func urlValues(values map[string]string) *url.Values {
	v := &url.Values{}
	for key, value := range values {
		v.Set(key, value)
	}
	return v
}
