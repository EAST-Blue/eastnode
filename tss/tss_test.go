package tss_test

import (
	"eastnode/tss/signer"
	"eastnode/tss/verifier"
	"encoding/hex"
	"testing"

	"github.com/bytemare/frost"
)

func hexStringToBytes(hexStr string) ([]byte, error) {
	decodedBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	return decodedBytes, nil
}

func Test_FullSign3of3(t *testing.T) {
	signer1 := signer.New("2NEpo7TZRRrLZSi2U", 2, 3)
	signer2 := signer.New("6J8jK8MZYvJbTW8HD", 2, 3)
	signer3 := signer.New("4gBdyMDWCBmUz9xwK", 2, 3)
	serializedData := signer1.DKGRound1()
	serializedData2 := signer2.DKGRound1()
	serializedData3 := signer3.DKGRound1()

	var round1data [][]byte
	round1data = append(round1data, serializedData)
	round1data = append(round1data, serializedData2)
	round1data = append(round1data, serializedData3)

	encodedRound2data1, err := signer1.DKGRound2(round1data)
	if err != nil {
		t.Error(err)
	}
	encodedRound2data2, err := signer2.DKGRound2(round1data)
	if err != nil {
		t.Error(err)
	}
	encodedRound2data3, err := signer3.DKGRound2(round1data)
	if err != nil {
		t.Error(err)
	}

	var round2data1 [][]byte
	var round2data2 [][]byte
	var round2data3 [][]byte
	round2data1 = append(round2data1, encodedRound2data2)
	round2data1 = append(round2data1, encodedRound2data3)

	round2data2 = append(round2data2, encodedRound2data1)
	round2data2 = append(round2data2, encodedRound2data3)

	round2data3 = append(round2data3, encodedRound2data1)
	round2data3 = append(round2data3, encodedRound2data2)

	signer1.DKGFinalize(round1data, round2data1)
	signer2.DKGFinalize(round1data, round2data2)
	signer3.DKGFinalize(round1data, round2data3)

	message := []byte("example")
	comm := frost.CommitmentList{}

	comm1 := signer1.Commit()
	comm2 := signer2.Commit()
	comm3 := signer3.Commit()
	comm = append(comm, &comm1, &comm2, &comm3)

	sig := []*frost.SignatureShare{}
	sigShare2, err := signer2.SignAsParticipant(message, comm)
	if err != nil {
		t.Error(err)
	}
	sigShare3, err := signer3.SignAsParticipant(message, comm)
	if err != nil {
		t.Error(err)
	}
	sig = append(sig, &sigShare2, &sigShare3)

	signature, err := signer1.SignAsCoordinator(message, comm, sig)
	if err != nil {
		t.Error(err)
	}

	if !verifier.Verify(message, &signature, signer1.GroupPublicKey.Encode()) {
		t.Error("Should verify true")
	}
}

func Test_PartialSign2of3(t *testing.T) {
	signer1 := signer.New("2NEpo7TZRRrLZSi2U", 2, 3)
	signer2 := signer.New("6J8jK8MZYvJbTW8HD", 2, 3)
	signer3 := signer.New("4gBdyMDWCBmUz9xwK", 2, 3)
	serializedData := signer1.DKGRound1()
	serializedData2 := signer2.DKGRound1()
	serializedData3 := signer3.DKGRound1()

	var round1data [][]byte
	round1data = append(round1data, serializedData)
	round1data = append(round1data, serializedData2)
	round1data = append(round1data, serializedData3)

	encodedRound2data1, err := signer1.DKGRound2(round1data)
	if err != nil {
		t.Error(err)
	}
	encodedRound2data2, err := signer2.DKGRound2(round1data)
	if err != nil {
		t.Error(err)
	}
	encodedRound2data3, err := signer3.DKGRound2(round1data)
	if err != nil {
		t.Error(err)
	}

	var round2data1 [][]byte
	var round2data2 [][]byte
	var round2data3 [][]byte
	round2data1 = append(round2data1, encodedRound2data2)
	round2data1 = append(round2data1, encodedRound2data3)

	round2data2 = append(round2data2, encodedRound2data1)
	round2data2 = append(round2data2, encodedRound2data3)

	round2data3 = append(round2data3, encodedRound2data1)
	round2data3 = append(round2data3, encodedRound2data2)

	signer1.DKGFinalize(round1data, round2data1)
	signer2.DKGFinalize(round1data, round2data2)
	signer3.DKGFinalize(round1data, round2data3)

	message := []byte("example")
	comm := frost.CommitmentList{}

	comm1 := signer1.Commit()
	comm2 := signer2.Commit()
	comm = append(comm, &comm1, &comm2)

	sig := []*frost.SignatureShare{}
	sigShare2, err := signer2.SignAsParticipant(message, comm)
	if err != nil {
		t.Error(err)
	}
	sig = append(sig, &sigShare2)

	signature, err := signer1.SignAsCoordinator(message, comm, sig)
	if err != nil {
		t.Error(err)
	}

	if !verifier.Verify(message, &signature, signer1.GroupPublicKey.Encode()) {
		t.Error("Should verify true")
	}
}

func Test_SignWithoutDKG(t *testing.T) {
	privkey, _ := hexStringToBytes("460866b60b005caff1ab5114533c4404552fef116f2de26c3948333b673aa82f")
	pubkey, _ := hexStringToBytes("02d9e571f49b0e1c0d43b4620334b0d522056965b957d71e1b3c5f8cc934254f09")
	groupPubkey, _ := hexStringToBytes("036c3a1567535bdbad025cb27f872bf3d63efcc85d491feb4e7a9d012c048a7df7")
	signer1 := signer.NewFromStaticKeys("12D3KooWS5u5JCdm9eZCzDhV1tCNgHzwnd86xfWFTAzicuUjEaSN", 2, 2, privkey, pubkey, groupPubkey)

	privkey, _ = hexStringToBytes("936d7b0a22568ce17c2ff409aa02d06ea086c8e9233ee2585d6ebcfdd3d22bfe")
	pubkey, _ = hexStringToBytes("035fa39e4a8d3add2a6528861e68089e0343652a825e3be3ac45076bc0e1ab5119")
	signer2 := signer.NewFromStaticKeys("12D3KooW9wuqbsTYjxBGihtyU674SgHKvFBZMjLsPFcLE3CjRSeK", 2, 2, privkey, pubkey, groupPubkey)

	message := []byte("example")
	comm := frost.CommitmentList{}

	comm1 := signer1.Commit()
	comm2 := signer2.Commit()
	comm = append(comm, &comm1, &comm2)

	sig := []*frost.SignatureShare{}
	sigShare2, err := signer2.SignAsParticipant(message, comm)

	if err != nil {
		t.Error(err)
	}
	sig = append(sig, &sigShare2)

	signature, err := signer1.SignAsCoordinator(message, comm, sig)
	if err != nil {
		t.Error(err)
	}

	if !verifier.Verify(message, &signature, signer1.GroupPublicKey.Encode()) {
		t.Error("Should verify true")
	}
}
