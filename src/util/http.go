package util

import (
	"bytes"
	"encoding/json"
	"github.com/ImpactDevelopment/ImpactServer/src/util/mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// HttpRequest wraps http.Request so that we can provide custom methods
type HttpRequest struct {
	Req *http.Request
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

func (r *HttpRequest) Do() (*http.Response, error) {
	return http.DefaultClient.Do(r.Req)
}

func formValues(v map[string]string) string {
	values := &url.Values{}
	for key, value := range v {
		values.Set(key, value)
	}
	return values.Encode()
}
