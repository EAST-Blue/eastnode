package tss_test

import (
	"eastnode/tss/signer"
	"eastnode/tss/verifier"
	"testing"

	"github.com/bytemare/frost"
)

func Test_FullSign3of3(t *testing.T) {
	signer1 := signer.New(1337, 2, 3)
	signer2 := signer.New(15623516, 2, 3)
	signer3 := signer.New(1, 2, 3)
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
	signer1 := signer.New(1337, 2, 3)
	signer2 := signer.New(15623516, 2, 3)
	signer3 := signer.New(1, 2, 3)
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
