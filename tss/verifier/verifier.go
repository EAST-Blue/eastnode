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

	signatureCopy := frost.Signature{
		R: cs.Group.NewElement(),
		Z: cs.Group.NewScalar(),
	}

	signatureCopy.R = signature.R.Copy()
	signatureCopy.Z = signature.Z.Copy()

	return frost.Verify(cs, message, &signatureCopy, publicKey)
}

// func Find(
// 	message []byte,
// 	commitments frost.CommitmentList,
// 	signatureShares []*frost.SignatureShare,
// 	encodedPublicKeys [][]byte,
// 	encodedGroupPublicKey []byte,
// ) (bool, group.Scalar) {

// 	cs := frost.Secp256k1.Configuration().Ciphersuite

// 	groupPublicKey := cs.Group.Base()
// 	err := groupPublicKey.Decode(encodedGroupPublicKey)
// 	if err != nil {
// 		panic(err)
// 	}

// 	configuration := frost.Secp256k1.Configuration(groupPublicKey)
// 	checker := configuration.Participant(nil, nil)

// 	for i, signatureShare := range signatureShares {
// 		commitmentI := commitments.Get(signatureShare.Identifier)
// 		if commitmentI == nil {
// 			panic("commitment not found")
// 		}

// 		pki := encodedPublicKeys[i]
// 		publicKey := cs.Group.Base()
// 		err := publicKey.Decode(pki)
// 		if err != nil {
// 			return false, *signatureShare.Identifier
// 		}

// 		if !checker.VerifySignatureShare(commitmentI, publicKey, signatureShare.SignatureShare, commitments, message) {
// 			return false, *signatureShare.Identifier
// 		}
// 	}

// 	return true, *cs.Group.NewScalar().Zero()
// }
