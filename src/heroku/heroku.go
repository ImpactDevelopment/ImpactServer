package heroku

import (
	"context"
	"fmt"
	"os"
	"time"

	heroku "github.com/heroku/heroku-go/v5"
)

const RELEASE_TIME = 24 * time.Hour

var token string
var app_id string

func init() {
	token = os.Getenv("HEROKU_API_TOKEN")
	app_id = os.Getenv("HEROKU_APP_ID")
	if token == "" || app_id == "" {
		fmt.Println("WARNING: Not checking Heroku I don't have an API key!")
	}
}

func NoRecentReleases() bool {
	if token == "" || app_id == "" {
		fmt.Println("Assuming potentially recent releases, since no Heroku api access!")
		return false
	}
	heroku.DefaultTransport.BearerToken = token

	h := heroku.NewService(heroku.DefaultClient)
	apps, err := h.AppList(context.TODO(), nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Invalid Heroku API token? Assuming potentially recent releases!")
		return false
	}
	for _, app := range apps {
		if app.ID == app_id {
			fmt.Println("Found myself on Heroku, app ID is", app.ID, "and app name is", app.Name)
			fmt.Println("Most recent release is", app.ReleasedAt.Format(time.RFC3339))
			dur := time.Since(*app.ReleasedAt)
			fmt.Println("Duration since last release is", dur)
			return dur > RELEASE_TIME
		}
	}
	panic("I was provided with a heroku api token and heroku app id, but the api token did not grant me access to that app?")
}
