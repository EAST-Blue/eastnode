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

type UpdateOutpointSpendingData struct {
	PreviousTxHash  string
	PreviousTxIndex uint32

	SpendingTxHash       string
	SpendingTxIndex      uint32
	SpendingBlockHash    string
	SpendingBlockHeight  uint64
	SpendingBlockTxIndex uint32
	Sequence             uint32
	SignatureScript      string
	Witness              string
}

type Block struct {
	Hash     string `gorm:"index:idx_hash" json:"hash"`
	Height   uint64 `gorm:"index:idx_height" json:"height"`
	IsOrphan bool   `json:"is_orphan"`

	PreviousBlock string `json:"previous_block"`
	Version       int32  `json:"version"`
	Nonce         uint32 `json:"nonce"`
	Timestamp     uint32 `json:"timestamp"`
	Bits          string `json:"bits"`
	MerkleRoot    string `json:"merkle_root"`
}

type Transaction struct {
	Hash     string `gorm:"index:idx_hash;unique" json:"hash"`
	LockTime uint32 `json:"lock_time"`
	Version  int32  `json:"version"`
	Safe     bool   `json:"safe"`

	BlockID     uint   `json:"block_id"`
	BlockHash   string `gorm:"index:idx_block_hash" json:"block_hash"`
	BlockHeight uint64 `gorm:"index:idx_block_height" json:"block_height"`
	BlockIndex  uint32 `gorm:"index:idx_block_index" json:"block_index"`
}

type OutPoint struct {
	SpendingTxHash       string `gorm:"index:idx_spending_tx_hash" json:"spending_tx_hash"`
	SpendingTxIndex      uint32 `json:"spending_tx_index"`
	SpendingBlockHash    string `json:"spending_block_hash"`
	SpendingBlockHeight  uint64 `json:"spending_block_height"`
	SpendingBlockTxIndex uint32 `json:"spending_block_tx_index"`
	Sequence             uint32 `json:"sequence"`
	SignatureScript      string `json:"signature_script"`
	Witness              string `json:"witness"`

	FundingTxHash       string `gorm:"index:idx_funding_tx_hash_funding_tx_index,priority:1" json:"funding_tx_hash"`
	FundingTxIndex      uint32 `gorm:"index:idx_funding_tx_hash_funding_tx_index,priority:2" json:"funding_tx_index"`
	FundingBlockHash    string `json:"funding_block_hash"`
	FundingBlockHeight  uint64 `json:"funding_block_height"`
	FundingBlockTxIndex uint32 `json:"funding_block_tx_index"`

	PkScript string `json:"pk_script"`
	Value    int64  `json:"value"`
	Spender  string `json:"spender"`
	Type     string `json:"type"`

	P2shAsmScripts    *P2shAsmScripts `json:"p2sh_asm_scripts" gorm:"-"`
	PkAsmScripts      *[]string       `json:"pk_asm_scripts" gorm:"-"`
	WitnessAsmScripts *[]string       `json:"witness_asm_scripts" gorm:"-"`
}

func NewDB(dialector gorm.Dialector, opts ...gorm.Option) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, opts...)
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Indexer{}, &Block{}, &Transaction{}, &OutPoint{})
	if err != nil {
		panic(err)
	}

	return db, nil
}
