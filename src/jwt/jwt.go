package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/ImpactDevelopment/ImpactServer/src/users"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	jwt "github.com/gbrlsnchs/jwt/v3"
	"github.com/google/uuid"
)

type impactUserJWT struct {
	jwt.Payload
	Roles  []string `json:"roles"`
	Legacy bool     `json:"legacy"`
	MCUUID string   `json:"mcuuid,omitempty"`
	HWID   string   `json:"hwid,omitempty"`
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

	// TODO perhaps this should instead be the static https://api.impactclient.net/v1 when APP_ENV is heroku
	// and otherwise, http://api.localhost:PORT/? idk
	jwtIssuerURL = os.Getenv("JWT_ISSUER_URL")
	if jwtIssuerURL == "" {
		fmt.Println("WARNING: JWT_ISSUER_URL is empty, all tokens will be invalid!")
	}
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

// roles is the list of roles that we should sign them as having
// mcUuid or hwid is something to verify that this token is intended for them specifically
// i.e. to prevent duplication attacks
// i.e. i.e. basically this JWT should **only work for one client**, the one that asked for it
// for example, if logging in using mojang, "mcUuid" would be their UUID
//  the client verifies that this is the UUID that they are
//  if not, the JWT isn't for them, so it's invalid
// for another example, if logging in using email / pass / hwid, hwid is the hwid
//  this allows things like "only two hwids per account"
//  the client would verify that that is their hwid
// auth cannot be the same as their impact account identifier since they won't already know it
// (like they can't check its value against something they already know)
func CreateJWT(user *users.User, subject string, mcUuid *uuid.UUID) []byte {
	now := time.Now()
	roles := user.Roles()
	rolesArr := make([]string, len(roles))
	for i := range roles {
		rolesArr[i] = roles[i].ID
	}

	mcUuidStr := "" // go doesn't have ? : ternary :rage:
	if mcUuid != nil {
		mcUuidStr = mcUuid.String()
	}

	payload := impactUserJWT{
		Payload: jwt.Payload{
			Issuer:         jwtIssuerURL,
			Subject:        subject,
			ExpirationTime: jwt.NumericDate(now.Add(24 * time.Hour)),
			IssuedAt:       jwt.NumericDate(now),
		},
		Roles:  rolesArr,
		MCUUID: mcUuidStr,
		Legacy: user.IsLegacy(),
	}

	token, err := jwt.Sign(payload, rs512)
	if err != nil {
		panic(err)
	}
	return token
}
