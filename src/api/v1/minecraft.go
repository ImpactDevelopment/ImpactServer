package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/labstack/echo"
	"net/http"
)

type role struct {
	ID string `json:"id"`
}

type userinfo struct {
	Roles []role `json:"roles"`
	Icon  string `json:"icon,omitempty"`
	Cape  string `json:"cape,omitempty"`
}

func hashUUID(uuid string) string {
	hash := sha256.Sum256([]byte(uuid))
	return hex.EncodeToString(hash[:])
}

func userInfo(c echo.Context) error {
	// Test data
	res := map[string]userinfo{
		hashUUID("a4f77739d15e4dc2b957219a2f6f9244"): {
			Roles: []role{
				{ID: "pepsi"},
			},
			Icon: "https://raw.githubusercontent.com/ImpactDevelopment/Resources/master/textures/Pepsi_32.png",
			Cape: "http://i.imgur.com/SKjRGbH.png",
		},
	}

	return c.JSON(http.StatusOK, res)
}
