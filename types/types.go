package types

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cbergoon/merkletree"
	"github.com/dustinxie/ecc"
	"github.com/near/borsh-go"
)

// Kind: ["call", "view", "deploy", "genesis"]
// FunctionName: "any"
// Args: []string
type Action struct {
	Kind         string   `json:"kind"`
	FunctionName string   `json:"function_name"`
	Args         []string `json:"args"`
}

type RpcReply struct {
	BlockHash   string `json:"block_hash"`
	BlockHeight uint64 `json:"block_height"`
	Result      []byte `json:"result"`
}

// ID hex, Signature hex, Transaction hex
type SignedTransaction struct {
	ID          string `json:"id"`
	Signature   string `json:"signed"`
	Transaction string `json:"transaction"`
}

func (st *SignedTransaction) Unpack() Transaction {
	txUnpacked := new(Transaction)

	txBytes, err := hex.DecodeString(st.Transaction)
	if err != nil {
		panic(err)
	}

	err = borsh.Deserialize(txUnpacked, txBytes)
	if err != nil {
		panic(err)
	}

	return *txUnpacked
}

func (st *SignedTransaction) IsValid() bool {
	txUnpacked := new(Transaction)

	txBytes, err := hex.DecodeString(st.Transaction)
	if err != nil {
		panic(err)
	}

	err = borsh.Deserialize(txUnpacked, txBytes)
	if err != nil {
		panic(err)
	}

	pubKeyBytes, err := hex.DecodeString(txUnpacked.Signer)
	if err != nil {
		panic(err)
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		panic(err)
	}

	sigBytes, err := hex.DecodeString(st.Signature)
	if err != nil {
		panic(err)
	}

	hashedMsg := sha256.Sum256(txBytes)
	verified := ecc.VerifyBytes(pubKey.ToECDSA(), hashedMsg[:], sigBytes, ecc.Normal)

	return verified
}

// Signer str, Receiver str, Actions hex
type Transaction struct {
	Nonce    uint64 `json:"nonce"`
	Signer   string `json:"signer"`
	Receiver string `json:"receiver"`
	Actions  string `json:"actions"`
}

// func (t *Transaction) serialize() []byte {

// }

// func (t *Transaction) deserialize() []byte {

// }

type CommonServerQuery struct {
	Method string   `json:"method"`
	Args   []string `json:"args"`
}

type MerkleTreeContent struct {
	Value string
}

func (t MerkleTreeContent) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.Value)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (t MerkleTreeContent) Equals(other merkletree.Content) (bool, error) {
	return t.Value == other.(MerkleTreeContent).Value, nil
}

type BlockHeader struct {
	ChainID     string
	Height      uint64
	Time        int64
	LastBlockID []byte
	DataHash    []byte
	StorageHash []byte
}
type Block struct {
	Header BlockHeader
	Data   []byte
}
type BlockHash []byte
