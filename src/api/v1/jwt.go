package v1

import (
	"crypto/rsa"
	"fmt"
	"github.com/ImpactDevelopment/ImpactServer/src/util"
	"github.com/gbrlsnchs/jwt/v3"
	"os"
)

var rs512 jwt.Algorithm

func init() {
	var key *rsa.PrivateKey

	if env := os.Getenv("JWT_KEY"); env != "" {
		var err error
		key, err = util.StrToRsa(env)
		if err != nil {
			fmt.Println("WARNING: Unable to load JWT_KEY from the environment", err)
		}
	}

	if key == nil {
		fmt.Println("WARNING: JWT_KEY not specified, generating a temporary one")
		key = util.GenerateRsa()
	}

	// rs512 can be used to sign and verify tokens, e.g. jtw.Sign(payload []byte, rs512 Algorithm)
	rs512 = jwt.NewRS512(jwt.RSAPrivateKey(key), jwt.RSAPublicKey(&key.PublicKey))
}

func sign(payload interface{}) (token []byte, err error) {
	token, err = jwt.Sign(payload, rs512)
	return
}

// TODO get token from user account object
