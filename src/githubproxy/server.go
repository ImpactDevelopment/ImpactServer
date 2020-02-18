package githubproxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

var cached = make(map[string]string)
var cachedLock sync.Mutex

func Server() (e *echo.Echo) {
	e = echo.New()

	e.Match([]string{http.MethodHead, http.MethodGet}, "/*", proxyHandler)

	return
}

func getIfCached(url string) (string, bool) {
	cachedLock.Lock()
	defer cachedLock.Unlock()
	val, ok := cached[url]
	return val, ok
}

func proxyHandler(c echo.Context) error {
	file := c.Request().URL.Path
	url, ok := getProxiedUrl(file)
	if !ok {
		return c.String(404, "Unknown")
	}
	str, ok := getIfCached(url)
	if ok {
		return c.String(200, str)
	}
	str, err := fetch(url)
	if err != nil {
		return err
	}
	cachedLock.Lock()
	defer cachedLock.Unlock()
	cached[url] = str
	return c.String(200, str)
}

func getProxiedUrl(url string) (string, bool) {
	if strings.HasPrefix(url, "/Impact-") && (strings.HasSuffix(url, ".json") || strings.HasSuffix(url, ".json.asc")) {
		version := strings.Split(url, "Impact-")[1]
		version = strings.Split(version, ".json")[0]
		url := "https://github.com/ImpactDevelopment/ImpactReleases/releases/download/" + version + url
		return url, true
	}
	if url == "/maven.refmap.json" {
		return "https://impactdevelopment.github.io/Resources/data/maven.refmap.json", true
	}
	return "", false
}

func fetch(url string) (string, error) { // pasted from src/api/v1/motd.go
	log.Println("Fetching", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
