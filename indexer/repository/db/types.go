package db

import (
	"gorm.io/gorm"
)

const INDEXER_LAST_HEIGHT_KEY = "INDEXER_LAST_HEIGHT_KEY"

type Indexer struct {
	gorm.Model

	Key   string `gorm:"index:idx_key,unique"`
	Value string
}

type P2shAsmScripts struct {
	LockScripts   []string `json:"lock_scripts"`
	UnlockScripts []string `json:"unlock_scripts"`
}

type Block struct {
	Hash string `gorm:"index:idx_hash" json:"hash"`
	// TODO: need to add index for height
	Height   uint64 `json:"height"`
	IsOrphan bool   `json:"is_orphan"`

	PreviousBlock string `json:"previous_block"`
	Version       int32 `json:"version"`
	Nonce         uint32 `json:"nonce"`
	Timestamp     uint32 `json:"timestamp"`
	Bits          string `json:"bits"`
	MerkleRoot    string `json:"merkle_root"`
}

type Transaction struct {
	Hash     string `gorm:"index:idx_hash;unique" json:"hash"`
	LockTime uint32 `json:"lock_time"`
	Version  int32 `json:"version"`
	// TODO: what's definition of safe?
	// - safe utxo: safe to spend
	Safe bool `json:"safe"`

	BlockID     uint   `json:"block_id"`
	BlockHash   string `json:"block_hash"`
	BlockHeight uint64 `json:"block_height"`

	// TODO: need to add index for block_index
	BlockIndex uint32 `json:"block_index"`
}

type OutPoint struct {
	SpendingTxHash       string `json:"spending_tx_hash"`
	SpendingTxIndex      uint32 `json:"spending_tx_index"`
	SpendingBlockHash    string `json:"spending_block_hash"`
	SpendingBlockHeight  uint64 `json:"spending_block_height"`
	SpendingBlockTxIndex uint32 `json:"spending_block_tx_index"`
	Sequence             uint32 `json:"sequence"`
	SignatureScript      string `json:"signature_script"`
	Witness              string `json:"witness"`

	FundingTxHash       string `json:"funding_tx_hash"`
	FundingTxIndex      uint32 `json:"funding_tx_index"`
	FundingBlockHash    string `json:"funding_block_hash"`
	FundingBlockHeight  uint64 `json:"funding_block_height"`
	FundingBlockTxIndex uint32 `json:"funding_block_tx_index"`

	PkScript string `json:"pk_script"`
	Value    int64 `json:"value"`
	Spender  string `json:"spender"`
	Type     string `json:"type"`

	P2shAsmScripts    *P2shAsmScripts `json:"p2sh_asm_scripts"`
	PkAsmScripts      *[]string       `json:"pk_asm_scripts"`
	WitnessAsmScripts *[]string       `json:"witness_asm_scripts"`
}

type Vin struct {
	TxHash          string `gorm:"index:idx_tx_hash" json:"tx_hash"`
	TxIndex         uint32 `json:"tx_index"`
	BlockHash       string `json:"block_hash"`
	BlockHeight     uint64 `json:"block_height"`
	BlockTxIndex    uint32 `json:"block_tx_index"`
	Sequence        uint32 `json:"sequence"`
	SignatureScript string `json:"signature_script"`
	Witness         string `json:"witness"`

	FundingTxHash  string `json:"funding_tx_hash"`
	FundingTxIndex uint32 `json:"funding_tx_index"`

	PkScript string `json:"pk_script"`
	Value    int64 `json:"value"`
	Spender  string `json:"spender"`
	Type     string `json:"type"`

	P2shAsmScripts    *P2shAsmScripts `json:"p2sh_asm_scripts"`
	PkAsmScripts      *[]string       `json:"pk_asm_scripts"`
	WitnessAsmScripts *[]string       `json:"witness_asm_scripts"`
}

type Vout struct {
	TxHash       string `gorm:"index:idx_tx_hash" json:"tx_hash"`
	TxIndex      uint32 `json:"tx_index"`
	BlockHash    string `json:"block_hash"`
	BlockHeight  uint64 `json:"block_height"`
	BlockTxIndex uint32 `json:"block_tx_index"`

	PkScript string `json:"pk_script"`
	Value    int64 `json:"value"`
	Spender  string `json:"spender"`
	Type     string `json:"type"`

	P2shAsmScripts *P2shAsmScripts `json:"p2sh_asm_scripts" gorm:"-"`
	PkAsmScripts   *[]string       `json:"pk_asm_scripts" gorm:"-"`
}

type VinV1 struct {
	TxHash string `json:"tx_hash"`
	Index  uint32 `json:"index"`
	Value  uint64 `json:"value"`
}

type VoutV1 struct {
	TxHash   string `json:"tx_hash"`
	Index    uint32 `json:"index"`
	Address  string `json:"address"`
	PkScript string `json:"pk_script"`
	Value    uint64 `json:"value"`
}

type TransactionV1 struct {
	Hash     string `json:"hash"`
	LockTime uint32 `json:"lock_time"`
	Version  uint32 `json:"version"`

	Vins  []VinV1  `json:"vins"`
	Vouts []VoutV1 `json:"vouts"`
}

func NewDB(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Indexer{}, &Block{}, &Transaction{}, &Vin{}, &Vout{})
	if err != nil {
		panic(err)
	}

	return db, nil
}
