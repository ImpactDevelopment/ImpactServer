package v1

import (
	"time"

	"github.com/labstack/echo/v4"
)

const alteningReferral = "impact"

var alteningInfoStruct = TheAlteningInfo{
	Dashboard: &Dashboard{
		GenerateUrl: "https://panel.thealtening.com/?ref=" + alteningReferral + "#generator",
		AccountUrl:  "https://panel.thealtening.com/?ref=" + alteningReferral + "#account",
	},
	Generator: &Generator{
		FreeUrl: "https://thealtening.com/?ref=" + alteningReferral + "&type=transit&destination=https://thealtening.com/free/free-minecraft-alt",
		PaidUrl: "https://panel.thealtening.com/?ref=" + alteningReferral + "#generator",
	},
	// TODO load a list of `Promo`s at runtime
	Promos: &[]Promo{
		{
			Code:     alteningReferral,
			Discount: "20%",
		},
	},
	Enabled: true,
}

type TheAlteningInfo struct {
	Dashboard *Dashboard `json:"dashboard,omitempty"`
	Generator *Generator `json:"generate,omitempty"`
	Promos    *[]Promo   `json:"promotions,omitempty"`
	Enabled   bool       `json:"enabled"`
}

type Dashboard struct {
	GenerateUrl string `json:"generate,omitempty"`
	AccountUrl  string `json:"account,omitempty"`
}

type Generator struct {
	FreeUrl string `json:"free,omitempty"`
	PaidUrl string `json:"premium,omitempty"`
}

type Promo struct {
	Code     string     `json:"promo_code,omitempty"`
	Discount string     `json:"discount,omitempty"`
	Expiry   *time.Time `json:"expiry,omitempty"`
}

func getTheAlteningInfo(c echo.Context) error {
	return c.JSON(200, alteningInfoStruct)
}
