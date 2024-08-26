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

func Test_SignWithoutDKG(t *testing.T) {
	privkey, _ := hexStringToBytes("c1a341dec8519eccd2ec92ec0f4fca08cc18ea3c630ebafb70cfffa9dd1a72ef")
	pubkey, _ := hexStringToBytes("0306b653d26799f418ee43af52358d8e1a35586f393517f1e217b6d32ab5fa3c6b")
	groupPubkey, _ := hexStringToBytes("038313e927d056b3a2a91e2faaf16b7b6e69cada44fc83e11fab3dfd61cc867b8e")
	signer1 := signer.NewFromStaticKeys("2NEpo7TZRRrLZSi2U", 2, 3, pubkey, groupPubkey, privkey)

	privkey, _ = hexStringToBytes("660799dfb1cf5654c8af08bb8bb130aedc9424bfb97e7ce3a629377f134e9155")
	pubkey, _ = hexStringToBytes("0237e88194f7bfe00106a65ea924a8efc1e499028e6aae0ba14a7b0747ad9ca660")
	signer2 := signer.NewFromStaticKeys("6J8jK8MZYvJbTW8HD", 2, 3, pubkey, groupPubkey, privkey)

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
