package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
)

func GenerateRsa() (key *rsa.PrivateKey) {
	// Should never error
	key, _ = rsa.GenerateKey(rand.Reader, 4096)
	return
}

func RsaToStr(key *rsa.PrivateKey) string {
	bytes := x509.MarshalPKCS1PrivateKey(key)
	return base64.StdEncoding.EncodeToString(bytes)
}

func StrToRsa(keyString string) (key *rsa.PrivateKey, err error) {
	bytes, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return
	}
	key, err = x509.ParsePKCS1PrivateKey(bytes)
	if err != nil {
		return
	}

	return
}
