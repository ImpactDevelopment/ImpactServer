package web

import (
	"net/http"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/s3proxy"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
)

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type Release struct {
	TagName    string  `json:"tag_name"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

func prereleases(c echo.Context) error {
	objs, err := s3.New(s3proxy.AWSSession).ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("impactclient-files"),
		Prefix: aws.String("artifacts/Impact/"),
	})
	if err != nil {
		return err
	}

	keys := make(map[string]bool)

	for _, item := range objs.Contents {
		keys[*item.Key] = true
	}

	resp := make([]Release, 0)
	for k, _ := range keys {
		// e.g. artifacts/Impact/dev/dev-856f3ad-1.13.2/Impact-dev-856f3ad-1.13.2.jar
		parts := strings.Split(k, "/")
		fileName := parts[len(parts)-1] // Impact-dev-856f3ad-1.13.2.jar
		if !strings.HasPrefix(fileName, "Impact-") || !strings.HasSuffix(fileName, ".jar") {
			continue
		}

		tagName := parts[len(parts)-2]    // dev-856f3ad-1.13.2
		jsonFile := k[:len(k)-3] + "json" // artifacts/Impact/dev/dev-856f3ad-1.13.2/Impact-dev-856f3ad-1.13.2.json
		if _, ok := keys[jsonFile]; !ok {
			continue
		}
		jsonFileName := fileName[:len(fileName)-3] + "json" // Impact-dev-856f3ad-1.13.2.json

		resp = append(resp, Release{
			TagName:    tagName,
			Draft:      strings.Contains(tagName, "dev"),
			Prerelease: true,
			Assets: []Asset{
				Asset{
					Name: fileName,
					URL:  "https://files.impactclient.net/" + k,
				},
				Asset{
					Name: jsonFileName,
					URL:  "https://files.impactclient.net/" + jsonFile,
				},
			},
		})

	}
	c.Response().Header().Set("Cache-Control", "max-age=3600")
	return c.JSON(http.StatusOK, resp)
}
