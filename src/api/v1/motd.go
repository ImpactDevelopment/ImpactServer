package v1

import (
	"github.com/labstack/echo"
	"net/http"
)

var motdText = "testing123"

func init() {
	// TODO load motd text from github
}

func motd(c echo.Context) error {
	return c.String(http.StatusOK, motdText)
}
