package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/util/mediatype"

	"github.com/ImpactDevelopment/ImpactServer/src/cloudflare"
	"github.com/ImpactDevelopment/ImpactServer/src/s3proxy"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
)

var rels map[string]Release

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

func init() {
	githubToken = os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubToken == "" {
		fmt.Println("WARNING: No GitHub access token to bypass ratelimiting!")
	}
	var err error
	rels, err = allReleases()
	if err != nil {
		panic(err)
	}
	util.DoRepeatedly(15*time.Minute, func() {
		newRel, err := allReleases()
		if err != nil {
			log.Println("RELEASES ERROR", err)
			return
		}
		if !reflect.DeepEqual(rels, newRel) {
			rels = newRel

			cloudflare.PurgeURLs([]string{"http://impactclient.net/releases.json"})
		}
	})
}

func releases(c echo.Context) error {
	relsCopy := rels // vague multithreading protection idk lmao
	resp := make([]Release, 0, len(rels))
	for _, v := range relsCopy {
		resp = append(resp, v)
	}
	return c.JSON(http.StatusOK, resp)
}

func allReleases() (map[string]Release, error) {
	resp := make(map[string]Release)
	err := githubReleases(resp)
	if err != nil {
		return nil, err
	}
	err = s3Releases(resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func githubReleases(rels map[string]Release) error {
	// not strictly necessary given that we won't be querying all that often
	// but we have no idea who else is on this IP (shared host from heroku)
	// so to guard against posssible "noisy neighbors" who are spamming github's api
	// we provoide an authorization token so that we get our own rate limit regardless of IP
	req, err := util.GetRequest("https://api.github.com/repos/ImpactDevelopment/ImpactReleases/releases")
	if err != nil {
		fmt.Println("Github error building request", err)
		return err
	}
	req.SetQuery("per_page", "100")
	req.Accept(mediatype.JSON)
	if githubToken != "" {
		req.Authorization("Basic", githubToken)
	}

	resp, err := req.Do()
	if err != nil {
		fmt.Println("Github error", err)
		return err
	}

	var releasesData []Release
	err = resp.JSON(&releasesData)
	if err != nil || len(releasesData) == 0 {
		fmt.Println("Github returned invalid json reply!!")
		fmt.Println(err)
		fmt.Println(resp.String())
		return err
	}

	for _, rel := range releasesData {
		rels[rel.TagName] = rel
	}
	return nil
}

func s3Releases(resp map[string]Release) error {
	objs := make([]*s3.Object, 0)
	err := s3.New(s3proxy.AWSSession).ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String("impactclient-files"),
		Prefix: aws.String("artifacts/Impact/"),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) (shouldContinue bool) {
		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			objs = append(objs, obj)
		}
		return true
	})
	if err != nil {
		fmt.Println("s3 error but let's not break the client for everyone since this only affects premium")
		fmt.Println(err)
		return nil
	}

	keys := make(map[string]bool)

	for _, item := range objs {
		if *item.StorageClass != "STANDARD" {
			continue
		}
		keys[*item.Key] = true
	}

	for k := range keys {
		// e.g. artifacts/Impact/dev/dev-856f3ad-1.13.2/Impact-dev-856f3ad-1.13.2.jar
		parts := strings.Split(k, "/")
		fileName := parts[len(parts)-1] // Impact-dev-856f3ad-1.13.2.jar
		if !strings.HasPrefix(fileName, "Impact-") || !strings.HasSuffix(fileName, ".jar") {
			continue
		}

		tagName := parts[len(parts)-2]             // dev-856f3ad-1.13.2
		fullPath := k[:len(k)-3]                   // artifacts/Impact/dev/dev-856f3ad-1.13.2/Impact-dev-856f3ad-1.13.2.
		internalName := fileName[:len(fileName)-3] // Impact-dev-856f3ad-1.13.2.

		if _, ok := keys[fullPath+"json"]; !ok {
			continue
		}

		rel := Release{
			TagName:    tagName,
			Draft:      strings.Contains(tagName, "dev"),
			Prerelease: !strings.Contains(tagName, "release"),
			Assets: []Asset{
				{
					Name: fileName,
					URL:  "https://files.impactclient.net/" + k,
				},
				{
					Name: internalName + "json",
					URL:  "https://files.impactclient.net/" + fullPath + "json",
				},
			},
		}

		if _, ok := keys[fullPath+"json.asc"]; ok {
			rel.Assets = append(rel.Assets, Asset{
				Name: internalName + "json.asc",
				URL:  "https://files.impactclient.net/" + fullPath + "json.asc",
			})
		}
		resp[tagName] = rel
	}
	return nil
}
