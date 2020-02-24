package paypal

import (
	"fmt"
	"os"

	"github.com/plutov/paypal/v3"
)

var client *paypal.Client

func init() {
	// Load our secrets from the environment
	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	secret := os.Getenv("PAYPAL_CLIENT_SECRET")
	apiBase := os.Getenv("PAYPAL_API_BASE")

	// API base defaults
	switch apiBase {
	case "":
		apiBase = paypal.APIBaseLive
	case "sb":
		fallthrough
	case "sandbox":
		apiBase = paypal.APIBaseSandBox
	}

	if clientID == "" {
		fmt.Println(`WARNING: no PAYPAL_CLIENT_ID specified, falling back to "sb", things may not function as intended`)
		clientID = "sb"
		if apiBase == paypal.APIBaseLive {
			apiBase = paypal.APIBaseSandBox
		}
	}

	if secret == "" && clientID != "sb" {
		fmt.Println(`WARNING no PAYPAL_CLIENT_SECRET is specified, falling back to "sb"`)
		secret = "sb"
	}

	var err error
	client, err = paypal.NewClient(clientID, secret, apiBase)
	if err != nil {
		fmt.Println("Error loggin into paypal:", err.Error())
		return
	}
	token, err := client.GetAccessToken()
	if err != nil {
		fmt.Println("Error getting paypal token:", err.Error())
		return
	}

	// Only print token when not using the "live" api, i.e. sandbox mode
	if apiBase == paypal.APIBaseLive {
		fmt.Println("Logged into paypal successfully")
	} else {
		fmt.Println("Logged into paypal successfully and received token: ", token.Token)
	}
}
