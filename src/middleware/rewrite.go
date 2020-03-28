package middleware

import (
	"github.com/labstack/echo/v4"
	"regexp"
	"strconv"
	"strings"
)

// RegexRewrite is based on echo's Rewrite but expects a regex rather than trying to create one from a glob
func RegexRewrite(rules map[string]string) echo.MiddlewareFunc {
	// Convert the string:string map to a regex:string map
	rulesRegex := map[*regexp.Regexp]string{}
	for k, v := range rules {
		rulesRegex[regexp.MustCompile(k)] = v
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			url := c.Request().URL

			// Rewrite
			for k, v := range rulesRegex {
				replacer := captureTokens(k, url.Path)
				if replacer != nil {
					url.Path = replacer.Replace(v)
					break
				}
			}

			return next(c)
		}
	}
}

// captureTokens is based on echo/v4@v4.1.11/middleware/middleware.go but supports replacing $0
func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	matches := groups[0]
	replace := make([]string, 2*len(matches))
	for i, match := range matches {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i)
		replace[j+1] = match
	}
	return strings.NewReplacer(replace...)
}
