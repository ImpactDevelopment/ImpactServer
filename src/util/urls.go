package util

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func init() {
	// Run on init to panic early if we typo our hardcoded urls
	addr := GetServerURL()
	fmt.Println("Server URL is", addr)
}

func GetServerURL() *url.URL {
	var (
		addr *url.URL
		err  error
	)

	if env := os.Getenv("SERVER_URL"); env != "" {
		addr, err = url.Parse(env)
	} else {
		port := strings.TrimSpace(os.Getenv("PORT"))
		switch port {
		case "80":
			port = ""
		case "":
			port = ":3000"
		default:
			port = ":" + port
		}

		addr, err = url.Parse("http://localhost" + port)
	}

	if err != nil {
		panic("Failed to parse server url")
	}

	return addr
}

// SetQuery changes the query parameters on the given url
func SetQuery(address *url.URL, key, value string) {
	query := address.Query()
	query.Set(key, value)
	address.RawQuery = query.Encode()
}