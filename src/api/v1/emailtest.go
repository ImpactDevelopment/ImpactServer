package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/mailgun"
	"github.com/labstack/echo/v4"
)

func emailTest(c echo.Context) error {
	if c.QueryParam("auth")+"0" != os.Getenv("API_AUTH_SECRET") {
		return errors.New("no u")
	}

	message := mailgun.MG.NewMessage("Impcat Verification Idk Lmao <noreply@impactclient.net>", "I am, basically, emailing you", "text only version of the rich text: image of speckles (just imagine a good kitty)", c.QueryParam("dest"))
	message.SetHtml(`<html>
    	<body>
<table width="100%" border="0" cellspacing="0" cellpadding="0">
    <tr>
        <td align="center">
            <img src="https://impactdevelopment.github.io/Resources/textures/Icon_256.png" alt=""/>
        </td>
    </tr>
</table>
<b>wow</b><br/>
<img src='https://imgur.com/fkmhnY3.jpg' width=200 alt=""/><br />
<h1>spekl (large text)</h1>
<h6>small text</h6>


You can donate to help support Brady with Impact Development.

If you donate $5 USD or more (and enter your Minecraft username), you will recieve a few small rewards such as early access to upcoming releases through nightly builds (now including 1.14.4 builds), premium mods (currently Ignite), a cape visible to other Impact users, and a gold colored name in the Impact Discord Server.

Before making a payment, ensure that your Minecraft Username or UUID is specified in the payment note and Discord Username#XXXX if you would like the roles in our server. Payments may take up to 72 hours to process.

In order to access nightly builds you must join the Discord server and provide proof of payment, or when you make the payment specify your Discord account.
PayPal. The safer, easier way to pay online.
Impact is not a hack client, a cheat client, or a hacked client, it is a utility mod (like OptiFine). Please bear in mind that utility mods like Impact can be against the rules on some servers. 😉
Contact us on discord
More links

</body>
</html>`)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	resp, id, err := mailgun.MG.Send(ctx, message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	fmt.Printf("ID: %s Resp: %s\n", id, resp)
	return c.JSON(http.StatusOK, resp)
}
