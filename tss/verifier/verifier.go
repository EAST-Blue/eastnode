package verifier

import (
	"github.com/bytemare/frost"
)

func Verify(message []byte, signature *frost.Signature, encodedPublicKey []byte) bool {
	cs := frost.Secp256k1.Configuration().Ciphersuite

	publicKey := cs.Group.Base()
	err := publicKey.Decode(encodedPublicKey)
	if err != nil {
		return false
	}

	return frost.Verify(cs, message, signature, publicKey)
}
