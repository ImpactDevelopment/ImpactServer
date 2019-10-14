package util

import (
	"encoding/asn1"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Integration test
func TestRSA(t *testing.T) {
	// Create the keys
	prv := GenerateRsa()
	assert.NotNil(t, prv, "GenerateRsa should not return nil, maybe an error was ignored")
	pub := &prv.PublicKey

	// Export the keys to pem string
	prvStr := RsaToStr(prv)

	// Import the keys from pem string
	prvParsed, err := StrToRsa(prvStr)
	if assert.NoError(t, err) {
		pubParsed := &prvParsed.PublicKey

		// Export the newly imported keys
		prvParsedStr := RsaToStr(prvParsed)

		assert.Equal(t, prvStr, prvParsedStr, "Export and Import should result in the same private key")
		assert.Equal(t, prv, prvParsed, "Export and Import should result in the same private key")
		assert.Equal(t, pub, pubParsed, "Export and Import should result in the same public key")
	}
}

func TestStrToRsa(t *testing.T) {
	validKey := "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"
	invalidKey := "invalid/key="
	invalidBase64 := "invalid_base64"

	parsed, err := StrToRsa(validKey)
	if assert.NoError(t, err, "A valid key should not error") {
		assert.NotNil(t, parsed)
		assert.Equal(t, 512, parsed.Size(), "Expect size to be %d bytes (%d bits)", 4096/8, 4096)
	}

	parsed, err = StrToRsa(invalidKey)
	assert.Nil(t, parsed)
	assert.Error(t, err, "Should error out when an invalid key is passed")
	assert.IsType(t, asn1.StructuralError{}, err, "Should return StructuralError when passed an invalid key")

	err = nil
	parsed, err = StrToRsa(invalidBase64)
	assert.Nil(t, parsed)
	assert.Error(t, err, "Should error out when keyString is invalid base64")
	assert.IsType(t, base64.CorruptInputError(0), err, "Should return CorruptInputError when passed invalid base64")
}
