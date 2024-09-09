package types

import (
	"crypto/sha256"
	"eastnode/utils"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cbergoon/merkletree"
	"github.com/dustinxie/ecc"
)

// Kind: ["call", "view", "deploy", "genesis"]
// FunctionName: "any"
// Args: []string
type Action struct {
	Kind         string   `json:"kind"`
	FunctionName string   `json:"function_name"`
	Args         []string `json:"args"`
}

type ActionKind int

const (
	Call ActionKind = 0
	View ActionKind = 1
)

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

type JsonArray struct {
	Array []string
}

func (st *SignedTransaction) Unpack() Transaction {
	txUnpacked := new(Transaction)

	utils.DecodeHexAndBorshDeserialize(txUnpacked, st.Transaction)

	return *txUnpacked
}

func (st *SignedTransaction) IsValid() bool {
	txUnpacked := st.Unpack()

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

	txBytes, _ := hex.DecodeString(st.Transaction)
	hashedMsg := sha256.Sum256(txBytes)
	verified := ecc.VerifyBytes(pubKey.ToECDSA(), hashedMsg[:], sigBytes, ecc.Normal)

	hashedSignature := utils.SHA256([]byte(st.Signature))

	if hashedSignature != st.ID {
		panic("TxId invalid")
	}

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

type RuntimeServerQuery struct {
	Target       string   `json:"target"`
	FunctionName string   `json:"function_name"`
	Args         []string `json:"args"`
}

type CommonServerQuery struct {
	FunctionName string   `json:"function_name"`
	Args         []string `json:"args"`
}

type BitcoinServerQuery struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type ServerQueryReply struct {
	BlockHash   string `json:"block_hash"`
	BlockHeight uint64 `json:"block_height"`
	Result      string `json:"result"`
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
	BitcoinHash string
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
