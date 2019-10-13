package v1

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo"
)

func mojangLogin(c echo.Context) error {
	uuidStr, err := util.HasJoinedServer(c.QueryParam("username"), c.QueryParam("hash"))
	if err != nil {
		return err
	}
	uuidVal, err := uuid.Parse(uuidStr) // we do this to add the dashes btw
	if err != nil {
		return err
	}
	data, ok := loginData[hashUUID(uuidVal.String())]
	if ok && data != nil && len(data.Roles) > 0 {
		return c.JSON(http.StatusOK, data)
	}
	return c.JSON(http.StatusForbidden, []struct{}{})
}