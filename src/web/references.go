package web

import (
	"net/url"

	"github.com/ImpactDevelopment/ImpactServer/src/util"

	"github.com/labstack/echo"
)

func references(c echo.Context) error {
	target, err := url.Parse("https://impactdevelopment.github.io/Resources/data/references.json")
	if err != nil {
		return err //wtf
	}
	util.Proxy(c, target)
	return nil
}
