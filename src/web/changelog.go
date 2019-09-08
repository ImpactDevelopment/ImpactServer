package web

import (
	"bytes"
	"github.com/labstack/echo"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

const github = "https://impactdevelopment.github.io/Impact/"

func Changelog(c echo.Context) error {
	// Forward to the changelog hosted by github

	target, err := url.Parse(github + "changelog")
	if err != nil {
		return err //wtf
	}

	proxy := httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// Change the URL
			req.URL = target
			req.Header.Set("X-Forwarded-Host", req.Host)
			req.Host = target.Host

			// Don't send our cookies to github
			req.Header.Del("Cookie")
			req.Header.Del("Authorization")

			// Ask github not to compress so we can do string-replace in the response body
			req.Header.Set("Accept-Encoding", "identity")
		},
		// Epic string replace meme
		ModifyResponse: replaceLinks,
	}

	proxy.ServeHTTP(c.Response(), c.Request())
	return nil
}

func replaceLinks(res *http.Response) error {
	b, err := ioutil.ReadAll(res.Body) // TODO stream instead of reading whole body into memory?
	if err != nil {
		return err
	}
	err = res.Body.Close()
	if err != nil {
		return err
	}

	// Replace gh-pages links with relative links
	b = bytes.Replace(b, []byte(github), []byte("/"), -1) // replace html

	// Write our changes to the response
	res.Body = ioutil.NopCloser(bytes.NewReader(b))

	// Update content-length
	res.ContentLength = int64(len(b))
	res.Header.Set("Content-Length", strconv.Itoa(len(b)))

	return nil
}

func ImpactRedirect(c echo.Context) error {
	// 302 non-caching redirect
	return c.Redirect(http.StatusFound, "https://impactdevelopment.github.io"+c.Request().RequestURI)
}
