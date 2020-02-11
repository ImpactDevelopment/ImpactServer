package jwt

import (
	"crypto/rsa"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/paypal"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/gbrlsnchs/jwt/v3"
)

type impactUserJWT struct {
	jwt.Payload
	Roles       []string `json:"roles"`
	Legacy      bool     `json:"legacy"`
	MinecraftID string   `json:"mcuuid,omitempty"`
}
type donationJWT struct {
	jwt.Payload
	OrderID string `json:"order"`
	Amount  int    `json:"amount"`
}

var rs512 jwt.Algorithm

var jwtIssuerURL string

func init() {
	var key *rsa.PrivateKey

	if env := os.Getenv("JWT_KEY"); env != "" {
		var err error
		key, err = util.StrToRsa(env)
		if err != nil {
			fmt.Println("WARNING: Unable to load JWT_KEY from the environment", err)
		}
	}

	addr := util.GetServerURL()
	jwtIssuerURL = addr.Scheme + "://api." + addr.Host + "/v1"
	fmt.Println("JWT Issuer URL is", jwtIssuerURL)

	if key == nil {
		fmt.Println("WARNING: JWT_KEY not specified, generating a temporary one")
		key = util.GenerateRsa()
		fmt.Println("Printing private key since this is temporary")
		fmt.Println("Private key is", util.RsaToStr(key))
	}

	fmt.Println("Public key is", util.RsaPubToStr(&key.PublicKey))

	// rs512 can be used to sign and verify tokens, e.g. jtw.Sign(payload []byte, rs512 Algorithm)
	rs512 = jwt.NewRS512(jwt.RSAPrivateKey(key), jwt.RSAPublicKey(&key.PublicKey))
}

// CreateDonationJWT returns a jwt token for a paypal order which can then be used
// to register for an Impact Account.
func CreateDonationJWT(order *paypal.Order) string {
	now := time.Now()

	return createJWT(donationJWT{
		Payload: jwt.Payload{
			Issuer:         jwtIssuerURL,
			Subject:        "",
			Audience:       jwt.Audience{"impact_account"},
			ExpirationTime: jwt.NumericDate(now.Add(90 * 24 * time.Hour)),
			NotBefore:      jwt.NumericDate(now),
			IssuedAt:       jwt.NumericDate(now),
		},
		OrderID: order.ID,
		Amount:  order.Total(),
	})
}

// CreateUserJWT returns a jwt token for the user with the subject (if not empty).
// The client can then use this to verify that the user has authenticated
// with a valid Impact server by checking the signature and issuer.
// If the client chooses, it could cache the token and reuse it until its
// expiration time.
func CreateUserJWT(user *users.User) string {
	now := time.Now()

	return createJWT(impactUserJWT{
		Payload: jwt.Payload{
			Issuer:         jwtIssuerURL,
			Subject:        user.ID.String(),
			Audience:       jwt.Audience{"impact_client", "impact_account"},
			ExpirationTime: jwt.NumericDate(now.Add(24 * time.Hour)),
			IssuedAt:       jwt.NumericDate(now),
		},
		Roles:       user.RoleIDs(),
		MinecraftID: user.MinecraftID.String(),
		Legacy:      user.Legacy,
	})
}

func createJWT(payload interface{}) string {
	token, err := jwt.Sign(payload, rs512)
	if err != nil {
		return ""
	}
	return string(token)
}

// respondWithToken responds to a http request with the token or returns a HTTPError
func respondWithToken(user *users.User, c echo.Context) error {
	token := CreateUserJWT(user)
	if token == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "error creating jwt token")
	}

	// TODO respect Accept header
	return c.String(http.StatusOK, token)
}
