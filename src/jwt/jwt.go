package jwt

import (
	"crypto/rsa"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/ImpactDevelopment/ImpactServer/src/util/mediatype"
	"github.com/gbrlsnchs/jwt/v3"
)

type impactUserJWT struct {
	jwt.Payload
	Roles       []string `json:"roles"`
	Legacy      bool     `json:"legacy"`
	MinecraftID string   `json:"mcuuid,omitempty"`
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

// CreateJWT returns a jwt token for the user with the subject (if not empty).
// The client can then use this to verify that the user has authenticated
// with a valid Impact server by checking the signature and issuer.
// If the client chooses, it could cache the token and reuse it until its
// expiration time.
func CreateJWT(user *users.User) string {
	now := time.Now()

	payload := impactUserJWT{
		Payload: jwt.Payload{
			Issuer:         jwtIssuerURL,
			Subject:        "", // TODO maybe user email? or id?
			ExpirationTime: jwt.NumericDate(now.Add(24 * time.Hour)),
			IssuedAt:       jwt.NumericDate(now),
		},
		Roles:       user.RoleIDs(),
		MinecraftID: user.MinecraftID.String(),
		Legacy:      user.Legacy,
	}

	token, err := jwt.Sign(payload, rs512)
	if err != nil {
		panic(err)
	}
	return string(token)
}

// respondWithToken responds to a http request with the token or returns a HTTPError
func respondWithToken(user *users.User, c echo.Context) error {
	token := CreateJWT(user)
	if token == "" {
		return echo.NewHTTPError(http.StatusInternalServerError, "error creating jwt token")
	}

	// TODO more mediatypes
	accepts := util.Accepts(*c.Request(), "text/plain", mediatype.JSON)
	if accepts == nil {
		return echo.NewHTTPError(http.StatusNotAcceptable)
	}
	switch *accepts {
	case mediatype.JSON:
		return c.JSON(http.StatusOK, map[string]string{
			"token": token,
		})
	default:
		return c.String(http.StatusOK, token)
	}
}
