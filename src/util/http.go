package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/ImpactDevelopment/ImpactServer/src/util/mime"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// HttpRequest wraps http.Request so that we can provide custom methods
type HttpRequest struct {
	Req *http.Request
}

// HttpResponse wraps http.Response so that we can provide custom methods
type HttpResponse struct {
	Resp *http.Response
}

func GetRequest(address string) (*HttpRequest, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	request := &HttpRequest{req}

	return request, nil
}

func JSONRequest(address string, body interface{}) (*HttpRequest, error) {
	post, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(post))
	if err != nil {
		return nil, err
	}
	request := &HttpRequest{req}

	request.setContentType(mime.JSON)
	request.setLength(len(post))

	return request, nil
}

func FormRequest(address string, form map[string]string) (*HttpRequest, error) {
	post := formValues(form)

	req, err := http.NewRequest(http.MethodPost, address, strings.NewReader(post))
	if err != nil {
		return nil, err
	}
	request := &HttpRequest{req}

	request.setContentType(mime.Form)
	request.setLength(len(post))

	return request, nil
}

func (r HttpRequest) URL() *url.URL {
	return r.Req.URL
}

func (r *HttpRequest) SetQuery(key, value string) {
	SetQuery(r.URL(), key, value)
}

func (r *HttpRequest) SetHeader(key, value string) {
	r.Req.Header.Set(key, value)
}

func (r *HttpRequest) setLength(length int) {
	r.SetHeader("Content-Length", strconv.Itoa(length))
}

func (r *HttpRequest) setContentType(mimeType mime.MimeType) {
	r.SetHeader("Content-Type", mimeType.String())
}

func (r *HttpRequest) Accept(mimeType mime.MimeType) {
	r.SetHeader("Accept", mimeType.String())
}

func (r *HttpRequest) Authorization(authType string, authKey string) {
	r.SetHeader("Authorization", authType+" "+authKey)
}

func (r *HttpRequest) Do() (*HttpResponse, error) {
	resp, err := http.DefaultClient.Do(r.Req)
	return &HttpResponse{resp}, err
}

func (r HttpResponse) Ok() bool {
	return r.Code() == http.StatusOK
}

func (r HttpResponse) Code() int {
	return r.Resp.StatusCode
}

// Returns the full status string, e.g. "400 Bad Request"
func (r HttpResponse) Status() string {
	code := r.Code()
	return strconv.Itoa(code) + " " + http.StatusText(code)
}

func (r HttpResponse) GetHeader(key string) string {
	return r.Resp.Header.Get(key)
}

func (r HttpResponse) ContentType() mime.MimeType {
	return mime.MimeType(r.GetHeader("Content-Type"))
}

// Returns the body as a string
func (r *HttpResponse) String() (string, error) {
	defer r.Resp.Body.Close()

	str, err := ioutil.ReadAll(r.Resp.Body)
	return string(str), err
}

// Decodes the body into the provided interface{}
func (r *HttpResponse) JSON(v interface{}) error {
	defer r.Resp.Body.Close()

	return json.NewDecoder(r.Resp.Body).Decode(v)
}

// Decodes the body into the provided interface{}
func (r *HttpResponse) XML(v interface{}) error {
	defer r.Resp.Body.Close()

	return xml.NewDecoder(r.Resp.Body).Decode(v)
}

func formValues(v map[string]string) string {
	values := &url.Values{}
	for key, value := range v {
		values.Set(key, value)
	}
	return values.Encode()
}
