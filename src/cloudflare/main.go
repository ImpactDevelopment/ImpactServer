package cloudflare

import (
	"bytes"
	"encoding/json"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
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
		util.LogWarn("Not purging cloudflare cache since I don't have an API key!")
	}
}

func Purge() {
	util.LogInfo("Purging cloudflare cache of everything")
	purgeWithData(struct {
		PurgeEverything bool `json:"purge_everything"`
	}{true})
}

func PurgeURLs(urls []string) {
	util.LogInfo("Purging cloudflare cache of URLs ")
	util.LogInfo(urls)
	purgeWithData(struct {
		Files []string `json:"files"`
	}{urls})
}

func purgeWithData(jsonData interface{}) {
	url := "https://api.cloudflare.com/client/v4/zones/" + zone + "/purge_cache"
	data, err := json.Marshal(jsonData)
	if err != nil {
		util.LogWarn("Cloudflare marshal error " + err.Error())
		return
	}
	util.LogInfo("Cloudflare purging data:" + string(data))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	// shouldn't you be error handling here?!
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		util.LogWarn("Cloudflare purge error " + err.Error())
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			util.LogWarn("Error closing body. " + err.Error())
		}
	}()
	body, _ := ioutil.ReadAll(resp.Body)
	util.LogInfo("Cloudflare response body: " + string(body))
}
