package mailgun

import (
	"os"

	mailgun "github.com/mailgun/mailgun-go/v3"
)

var MG = mailgun.NewMailgun(os.Getenv("MAILGUN_DOMAIN_IDENTIFIER"), os.Getenv("MAILGUN_API_KEY"))
