package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type MimeType string

const (
	JSON MimeType = "application/json"
	Form MimeType = "application/x-www-form-urlencoded"
)

func (t MimeType) String() string {
	return string(t)
}

func GetRequest(address string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func JSONRequest(address string, body interface{}) (*http.Request, error) {
	data := jsonData(body)

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", JSON.String())
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))

	return req, nil
}

func FormRequest(address string, form map[string]string) (*http.Request, error) {
	data := formData(form)

	req, err := http.NewRequest(http.MethodPost, address, strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", Form.String())
	req.Header.Add("Content-Length", strconv.Itoa(len(data)))

	return req, nil
}

func Accept(request *http.Request, mimeType MimeType) {
	request.Header.Set("Content-Type", JSON.String())
}

func jsonData(v interface{}) []byte {
	byteData, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	return byteData
}

func formData(v map[string]string) string {
	values := &url.Values{}
	for key, value := range v {
		values.Set(key, value)
	}

	return values.Encode()
}
