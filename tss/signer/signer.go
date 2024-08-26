package signer

import (
	"bytes"
	"errors"
	"io"
	"math/big"

	group "github.com/bytemare/crypto"

	"github.com/bytemare/frost"
	"github.com/bytemare/frost/dkg"
	"github.com/bytemare/secp256k1"
)

type Signer struct {
	n              int
	t              int
	SecretKey      *group.Scalar
	dkgData        dkg.Participant
	participant    frost.Participant
	Identifier     *group.Scalar
	PublicKey      *group.Element
	GroupPublicKey *group.Element
}

func New(identifier string, threshold int, maximumAmountOfParticipants int) Signer {
	configuration := frost.Secp256k1.Configuration()

	// Convert base58 encoded string to scalar
	// following libp2p's standard of peer ID
	var idBigInt big.Int
	idBigInt.SetString(identifier, 58)
	participantIdentifier := configuration.Ciphersuite.Group.NewScalar()
	participantIdentifier.SetInt(&idBigInt)

	dkgData := dkg.NewParticipant(
		configuration.Ciphersuite,
		participantIdentifier,
		maximumAmountOfParticipants,
		threshold,
	)

	return Signer{
		Identifier: participantIdentifier,
		dkgData:    *dkgData,
		n:          maximumAmountOfParticipants,
		t:          threshold,
	}
}

func NewFromStaticKeys(
	identifier string,
	threshold int,
	maximumAmountOfParticipants int,
	publicKey []byte,
	groupPublicKey []byte,
	secretKey []byte,
) Signer {
	configuration := frost.Secp256k1.Configuration()

	var idBigInt big.Int
	idBigInt.SetString(identifier, 58)
	participantIdentifier := configuration.Ciphersuite.Group.NewScalar()
	participantIdentifier.SetInt(&idBigInt)

	privkey := configuration.Ciphersuite.Group.NewScalar()
	if err := privkey.Decode(secretKey); err != nil {
		panic(err)
	}

	pubkey := configuration.Ciphersuite.Group.NewElement()
	if err := pubkey.Decode(publicKey); err != nil {
		panic(err)
	}

	groupPubkey := configuration.Ciphersuite.Group.NewElement()
	if err := groupPubkey.Decode(groupPublicKey); err != nil {
		panic(err)
	}

	configuration = frost.Secp256k1.Configuration(groupPubkey)
	participant := *configuration.Participant(participantIdentifier, privkey)

	return Signer{
		Identifier:     participantIdentifier,
		n:              maximumAmountOfParticipants,
		t:              threshold,
		participant:    participant,
		SecretKey:      privkey,
		PublicKey:      pubkey,
		GroupPublicKey: groupPubkey,
	}
}

func accumulateRound1Data(round1dataBytes [][]byte, n int) ([]*dkg.Round1Data, error) {
	var allRound1data []*dkg.Round1Data

	elementLength := secp256k1.ElementLength()
	scalarLength := secp256k1.ScalarLength()
	pokLength := elementLength + scalarLength
	identifierLength := scalarLength

	cs := frost.Secp256k1.Configuration().Ciphersuite

	for _, value := range round1dataBytes {
		round1data := dkg.Round1Data{
			ProofOfKnowledge: frost.Signature{
				R: cs.Group.Base(),
				Z: cs.Group.NewScalar(),
			},
			SenderIdentifier: cs.Group.NewScalar(),
		}

		if err := round1data.ProofOfKnowledge.Decode(group.Secp256k1, value[:pokLength]); err != nil {
			return allRound1data, err
		}
		if err := round1data.SenderIdentifier.Decode(value[pokLength : pokLength+identifierLength]); err != nil {
			return allRound1data, err
		}

		reader := bytes.NewReader(value[pokLength+identifierLength:])
		for {

			buf := make([]byte, elementLength)
			n, err := reader.Read(buf)

			if err != nil {
				if err == io.EOF {
					break
				}
				return allRound1data, err
			}

			comm := group.Secp256k1.NewElement()
			if err := comm.Decode(buf[:n]); err != nil {
				return allRound1data, err
			}

			round1data.Commitment = append(round1data.Commitment, comm)
		}

		allRound1data = append(allRound1data, &round1data)
	}

	accumulatedRound1Data := make([]*dkg.Round1Data, 0, n)
	accumulatedRound1Data = append(accumulatedRound1Data, allRound1data...)

	return accumulatedRound1Data, nil
}

func accumulateRound2Data(round2dataBytes [][]byte, identifier *group.Scalar) ([]*dkg.Round2Data, error) {
	var allRound2data []*dkg.Round2Data

	scalarLength := secp256k1.ScalarLength()
	cs := frost.Secp256k1.Configuration().Ciphersuite

	for _, data := range round2dataBytes {

		reader := bytes.NewReader(data)
		for {

			buf := make([]byte, scalarLength*3)
			n, err := reader.Read(buf)

			if err != nil {
				if err == io.EOF {
					break
				}
				return allRound2data, err
			}

			if n != scalarLength*3 {
				return allRound2data, errors.New("corrupted serialized round2data")
			}

			round2data := dkg.Round2Data{
				SenderIdentifier:   cs.Group.NewScalar(),
				ReceiverIdentifier: cs.Group.NewScalar(),
				SecretShare:        cs.Group.NewScalar(),
			}

			if err := round2data.SenderIdentifier.Decode(buf[:scalarLength]); err != nil {
				return allRound2data, err
			}

			if err := round2data.ReceiverIdentifier.Decode(buf[scalarLength : scalarLength*2]); err != nil {
				return allRound2data, err
			}

			if err := round2data.SecretShare.Decode(buf[scalarLength*2:]); err != nil {
				return allRound2data, err
			}

			if round2data.ReceiverIdentifier.Equal(identifier) == 1 {
				allRound2data = append(allRound2data, &round2data)
				break
			}

		}
	}

	return allRound2data, nil
}

func (s *Signer) DKGRound1() []byte {
	round1data := s.dkgData.Init()

	var serializedData []byte
	serializedData = append(serializedData, round1data.ProofOfKnowledge.Encode()...)
	serializedData = append(serializedData, round1data.SenderIdentifier.Encode()...)

	for _, value := range round1data.Commitment {
		serializedData = append(serializedData, value.Encode()...)
	}

	return serializedData
}

func (s *Signer) DKGRound2(round1dataBytes [][]byte) ([]byte, error) {

	decodedRound1Data, err := accumulateRound1Data(round1dataBytes, s.n)
	if err != nil {
		return []byte{}, err
	}
	round1Data := make([]*dkg.Round1Data, 0, s.n)
	round1Data = append(round1Data, decodedRound1Data...)

	round2Data, err := s.dkgData.Continue(round1Data)
	if err != nil {
		return []byte{}, err
	}

	var serializedData []byte
	for _, data := range round2Data {
		serializedData = append(serializedData, data.SenderIdentifier.Encode()...)
		serializedData = append(serializedData, data.ReceiverIdentifier.Encode()...)
		serializedData = append(serializedData, data.SecretShare.Encode()...)
	}

	return serializedData, nil
}

func (s *Signer) DKGFinalize(round1dataBytes [][]byte, round2dataBytes [][]byte) error {

	accumulatedRound1Data, err := accumulateRound1Data(round1dataBytes, s.n)
	if err != nil {
		return err
	}
	accumulatedRound2Data, err := accumulateRound2Data(round2dataBytes, s.Identifier)
	if err != nil {
		return err
	}

	var participantsSecretKey *group.Scalar
	var participantsPublicKey *group.Element
	participantsSecretKey, participantsPublicKey, groupPublicKeyGeneratedInDKG, err := s.dkgData.Finalize(
		accumulatedRound1Data,
		accumulatedRound2Data,
	)
	if err != nil {
		return err
	}

	s.PublicKey = participantsPublicKey
	s.GroupPublicKey = groupPublicKeyGeneratedInDKG
	s.SecretKey = participantsSecretKey

	configuration := frost.Secp256k1.Configuration(s.GroupPublicKey)
	s.participant = *configuration.Participant(s.dkgData.Identifier, s.SecretKey)

	return nil
}

func (s *Signer) Commit() frost.Commitment {
	commitment := s.participant.Commit()
	return *commitment
}

// Sign as Participant
func (s *Signer) SignAsParticipant(message []byte, commitments frost.CommitmentList) (frost.SignatureShare, error) {
	commitments.Sort()

	signatureShare, err := s.participant.Sign(message, commitments)
	if err != nil {
		return frost.SignatureShare{}, err
	}

	return *signatureShare, nil
}

// Sign as Coordinator and aggregate final signature
func (s *Signer) SignAsCoordinator(message []byte, commitments frost.CommitmentList, signatureShares []*frost.SignatureShare) (frost.Signature, error) {
	commitments.Sort()

	signatureShare, err := s.participant.Sign(message, commitments)
	if err != nil {
		return frost.Signature{}, err
	}

	signatureShares = append(signatureShares, signatureShare)

	signature := s.participant.Aggregate(commitments, message, signatureShares)

	cs := frost.Secp256k1.Configuration().Ciphersuite
	if !frost.Verify(cs, message, signature, s.GroupPublicKey) {
		panic("TODO: find the malicious signer by verifying each share")
	}

	return *signature, nil
}
