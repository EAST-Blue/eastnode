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
	gorm.Model

	ID uint `gorm:"primarykey" json:"id"`

	Hash     string `json:"hash"`
	Height   int32  `gorm:"index:idx_height" json:"height"`
	IsOrphan bool   `json:"is_orphan"`

	PreviousBlock string `json:"previous_block"`
	Version       int32  `json:"version"`
	Nonce         uint32 `json:"nonce"`
	Timestamp     uint32 `json:"timestamp"`
	Bits          string `json:"bits"`
	MerkleRoot    string `json:"merkle_root"`
}

type Transaction struct {
	gorm.Model

	ID uint `gorm:"primarykey" json:"id"`

	Hash     string `gorm:"index:idx_hash" json:"hash"`
	LockTime uint32 `json:"lock_time"`
	Version  int32  `json:"version"`
	Safe     bool   `json:"safe"`

	BlockID    uint   `json:"block_id"`
	BlockHash  string `json:"block_hash"`
	BlockIndex uint32 `json:"block_index"`
}

type OutPoint struct {
	gorm.Model

	ID uint `gorm:"primarykey" json:"id"`

	SpendingTxID    uint   `json:"spending_tx_id"`
	SpendingTxHash  string `gorm:"index:idx_spending_tx_hash" json:"spending_tx_hash"`
	SpendingTxIndex uint32 `json:"spending_tx_index"`
	Sequence        uint32 `json:"sequence"`
	SignatureScript string `json:"signature_script"`
	Witness         string `json:"witness"`

	FundingTxID    uint   `json:"funding_tx_id"`
	FundingTxHash  string `gorm:"index:idx_funding_tx_hash" json:"funding_tx_hash"`
	FundingTxIndex uint32 `json:"funding_tx_index"`
	PkScript       string `json:"pk_script"`
	Value          int64  `json:"value"`
	Spender        string `json:"spender"`
	Type           string `json:"type"`

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
