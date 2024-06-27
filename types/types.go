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

	hashedSignature := utils.SHA256(sigBytes)

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

// WasmRuntime

type BitcoinBlockHeader struct {
	ID            int64  `json:"id"`
	Version       int64  `json:"version"`
	Height        int64  `json:"height"`
	PreviousBlock string `json:"previous_block"`
	MerkleRoot    string `json:"merkle_root"`
	Hash          string `json:"hash"`
	Time          int64  `json:"time"`
	Nonce         int64  `json:"nonce"`
	Bits          int64  `json:"bits"`
}

type BitcoinTransaction struct {
	ID        int64  `json:"id"`
	Hash      string `json:"hash"`
	BlockHash string `json:"block_hash"`
	BlockId   int64  `json:"block_id"`
	LockTime  int64  `json:"lock_time"`
	Version   int64  `json:"version"`
	Safe      int64  `json:"safe"`
}

type BitcoinOutpoint struct {
	ID              int64  `json:"id"`
	Value           int64  `json:"value"`
	SpendingTxId    int64  `json:"spending_tx_id"`
	SpendingTxHash  string `json:"spending_tx_hash"`
	SpendingTxIndex int64  `json:"spending_tx_index"`
	Sequence        int64  `json:"sequence"`
	FundingTxId     int64  `json:"funding_tx_id"`
	FundingTxHash   string `json:"funding_tx_hash"`
	FundingTxIndex  int64  `json:"funding_tx_index"`
	SignatureScript string `json:"signature_script"`
	PkScript        string `json:"pk_script"`
	Witness         string `json:"witness"`
	Spender         string `json:"spender"`
	Type            string `json:"type"`
}
