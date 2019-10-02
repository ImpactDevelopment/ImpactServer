package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ImpactDevelopment/ImpactServer/src/s3proxy"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
)

var githubToken string

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

type ReleaseSource func() ([]Release, error)

func init() {
	githubToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubToken == "" {
		fmt.Println("WARNING: No GitHub access token to bypass ratelimiting!")
	}
}

func githubReleases() ([]Release, error) {
	// not strictly necessary given that we won't be querying all that often
	// but we have no idea who else is on this IP (shared host from heroku)
	// so to guard against posssible "noisy neighbors" who are spamming github's api
	// we provoide an authorization token so that we get our own rate limit regardless of IP
	req, err := http.NewRequest("GET", "https://api.github.com/repos/ImpactDevelopment/ImpactReleases/releases?per_page=100", nil)
	if githubToken != "" {
		req.Header.Set("Authorization", "Basic "+githubToken)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println("Github error", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	releasesData := make([]Release, 0)
	err = json.Unmarshal(body, &releasesData)
	if err != nil {
		fmt.Println("Github returned invalid json reply!!")
		fmt.Println(err)
		fmt.Println(string(body))
		return nil, err
	}
	return releasesData, nil
}

func s3Releases() ([]Release, error) {
	objs, err := s3.New(s3proxy.AWSSession).ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String("impactclient-files"),
		Prefix: aws.String("artifacts/Impact/"),
	})
	if err != nil {
		fmt.Println("s3 error but let's not break the client for everyone since this only affects premium")
		fmt.Println(err)
		return make([]Release, 0), nil
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
			Prerelease: !strings.Contains(tagName, "release"),
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
	return resp, nil
}

var releaseSources = []ReleaseSource{githubReleases, s3Releases}

func releases(c echo.Context) error {
	errCh := make(chan error)
	dataCh := make(chan []Release)

	for _, elem := range releaseSources {
		go func(source ReleaseSource) {
			data, err := source()
			if err != nil {
				errCh <- err
				return
			}
			dataCh <- data
		}(elem)
	}

	resp := make([]Release, 0)
	for _ = range releaseSources {
		select {
		case data := <-dataCh:
			resp = append(resp, data...)
		case err := <-errCh:
			return err
		}
	}

	return c.JSON(http.StatusOK, resp)
}
