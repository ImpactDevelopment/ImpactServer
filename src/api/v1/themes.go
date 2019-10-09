package v1

import (
	"net/http"

	"github.com/labstack/echo"
)

type (
	theme struct {
		Background  *background `json:"background,omitempty"`
		DefaultFont *font       `json:"default_font,omitempty"`
		TitleFont   *font       `json:"title_font,omitempty"`
		MOTDFont    *font       `json:"motd_font,omitempty"`
		// TODO add useful things to themes
	}
	background struct {
		// Epic meme make the client support data:image/png;base64 URIs
		URL string `json:"url,omitempty"`
	}
	font struct {
		Color uint32 `json:"color,omitempty"`
	}
)

// API Handler
func getThemes(c echo.Context) error {
	return c.JSON(http.StatusOK, themes)
}

var themes = map[string]theme{
	"Impact": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/Pink_sunset_at_Visevnik.jpg"},
	},
	"alexandre-godreau": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/alexandre-godreau-203580-unsplash.jpg"},
	},
	"andrew-ruiz": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/andrew-ruiz-406374-unsplash.jpg"},
	},
	"aniket-deole": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/aniket-deole-294646-unsplash.jpg"},
	},
	"bailey-zindel": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/bailey-zindel-396398-unsplash.jpg"},
	},
	"benjamin-voros": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/benjamin-voros-575800-unsplash.jpg"},
	},
	"casey-horner": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/casey-horner-1265505-unsplash.jpg"},
	},
	"daniel-leone": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/daniel-leone-185834-unsplash.jpg"},
	},
	"eberhard-grossgasteiger": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/eberhard-grossgasteiger-299348-unsplash.jpg"},
	},
	"gabriele-garanzelli": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/gabriele-garanzelli-529492-unsplash.jpg"},
	},
	"james-donovan": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/james-donovan-180375-unsplash.jpg"},
	},
	"john-westrock": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/john-westrock-638048-unsplash.jpg"},
	},
	"julian-zett": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/julian-zett-643140-unsplash.jpg"},
	},
	"juskteez-vu": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/juskteez-vu-3824-unsplash.jpg"},
	},
	"martin-jernberg": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/martin-jernberg-197949-unsplash.jpg"},
	},
	"nasa": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/nasa-53884-unsplash.jpg"},
	},
	"olivier-miche": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/olivier-miche-508901-unsplash.jpg"},
	},
	"pascal-debrunner": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/pascal-debrunner-634122-unsplash.jpg"},
	},
	"patrick-fore": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/patrick-fore-562304-unsplash.jpg"},
	},
	"stephan-seeber": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/stephan-seeber-507791-unsplash.jpg"},
	},
	"stephen-wheeler": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/stephen-wheeler-732168-unsplash.jpg"},
	},
	"tanya-nevidoma": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/tanya-nevidoma-1085291-unsplash.jpg"},
	},
	"vashishtha-jogi": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/vashishtha-jogi-101218-unsplash.jpg"},
	},
	"wolfgang-hasselmann": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/wolfgang-hasselmann-1403514-unsplash.jpg"},
	},
	"yuriy-garnaev": {
		Background: &background{URL: "https://impactdevelopment.github.io/Resources/textures/backgrounds/yuriy-garnaev-395879-unsplash.jpg"},
	},
}
