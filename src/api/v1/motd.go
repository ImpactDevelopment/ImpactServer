package v1

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/cloudflare"
	"github.com/labstack/echo"
)

const motdURL = "https://impactdevelopment.github.io/Resources/data/motd.txt"

var motd string

func init() {
	var err error
	motd, err = fetchMotd()
	if err != nil {
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		for range ticker.C {
			newer, err := fetchMotd()
			if err != nil {
				log.Println("MOTD ERROR", err)
			}
			newMotd(newer)
		}
	}()
}

func newMotd(newer string) {
	if newer != motd {
		motd = newer
		log.Println("MOTD UPDATE from", motd, "to", newer)
		cloudflare.PurgeURLs([]string{"https://api.impactclient.net/v1/motd"})
	}
}

func fetchMotd() (string, error) {
	resp, err := http.Get(motdURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getMotd(c echo.Context) error {
	return c.String(http.StatusOK, motd)
}
