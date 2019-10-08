package v1

import (
	"github.com/aws/aws-sdk-go/private/util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetThemes(t *testing.T) {
	// Override the returned themes
	themes = map[string]theme{
		"theme-one": {
			DefaultFont: font{Color: 0xff00cc},
		},
		"theme-two": {
			Background: background{URL: "hello, world"},
		},
	}
	expected := `{"theme-one":{"background":{},"default_font":{"color":16711884},"title_font":{"color":0},"motd_font":{"color":0}},"theme-two":{"background":{"url":"hello, world"},"default_font":{"color":0},"title_font":{"color":0},"motd_font":{"color":0}}}`

	e := getServer()
	res := test(e, "/v1/themes")

	assert.Equal(t, http.StatusOK, res.Code)
	body, err := ioutil.ReadAll(res.Body)
	if assert.NoError(t, err) {
		assert.Equal(t, expected, util.Trim(string(body)))
	}
}
