package v1

import (
	"net/http"

	"github.com/ImpactDevelopment/ImpactServer/src/jwt"
	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/google/uuid"
	"github.com/labstack/echo"
)

func mojangLoginJWT(c echo.Context) error {
	uuidStr, err := util.HasJoinedServer(c.QueryParam("username"), c.QueryParam("hash"))
	if err != nil {
		return err
	}
	uuidVal, err := uuid.Parse(uuidStr) // we do this to add the dashes btw
	if err != nil {
		return err
	}
	user := users.LookupUserByUUID(uuidVal)
	if user != nil && len(user.Roles()) > 0 {
		return c.JSONBlob(http.StatusOK, jwt.CreateJWT(user, uuidVal.String()))
	}
	return c.JSON(http.StatusForbidden, []struct{}{})
}
