package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	url := "https://api.cloudflare.com/client/v4/zones/" + zone + "/purge_cache"
	data, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println("Cloudflare marshal error", err)
		return
	}
	fmt.Println("Cloudflare purging data:", string(data))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println("Cloudflare purge error", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Cloudflare response body:", string(body))
}
