package cloudflare

import (
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"os"
)

var zone string
var key string

func init() {
	zone = os.Getenv("CLOUDFLARE_ZONE_IDENTIFIER")
	key = os.Getenv("CLOUDFLARE_API_KEY")
	if zone == "" || key == "" {
		fmt.Println("WARNING: Not purging cloudflare cache since I don't have an API key!")
	}
}

func Purge() {
	fmt.Println("Purging cloudflare cache of everything")
	purgeWithData(struct {
		PurgeEverything bool `json:"purge_everything"`
	}{true})
}

func PurgeURLs(urls []string) {
	fmt.Println("Purging cloudflare cache of URLs", urls)
	purgeWithData(struct {
		Files []string `json:"files"`
	}{urls})
}

func purgeWithData(jsonData interface{}) {
	if zone == "" {
		fmt.Println("WARNING: Not purging cloudflare cache since a zone is not specified!")
		return
	}
	if key == "" {
		fmt.Println("WARNING: Not purging cloudflare cache since a key is not specified!")
		return
	}

	url := "https://api.cloudflare.com/client/v4/zones/" + zone + "/purge_cache"
	req, err := util.JSONRequest(url, jsonData)
	if err != nil {
		fmt.Println("Cloudflare error building request", err)
		return
	}
	req.Authorization("Bearer", key)

	resp, err := req.Do()
	if err != nil {
		fmt.Println("Cloudflare purge error", err)
		return
	}

	fmt.Println("Cloudflare: code: "+resp.Status()+", body: ", resp.String())
}
