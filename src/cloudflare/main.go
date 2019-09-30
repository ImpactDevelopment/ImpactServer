package cloudflare

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func Purge() {
	zone := os.Getenv("CLOUDFLARE_ZONE_IDENTIFIER")
	key := os.Getenv("CLOUDFLARE_API_KEY")
	if zone == "" || key == "" {
		fmt.Println("NOT purging cloudflare cache since I don't have an API key!")
		return
	}
	fmt.Println("Purging cloudflare cache")

	url := "https://api.cloudflare.com/client/v4/zones/" + zone + "/purge_cache"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"purge_everything":true}`)))
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
