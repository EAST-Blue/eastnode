package model

import (
	"time"

	"gorm.io/gorm"
)

type P2shAsmScripts struct {
	LockScripts   []string `json:"lock_scripts"`
	UnlockScripts []string `json:"unlock_scripts"`
}

type Block struct {
	gorm.Model

	ID uint `json:"id"`

	Hash     string `json:"hash"`
	Height   int32  `json:"height"`
	IsOrphan bool   `json:"is_orphan"`

	PreviousBlock string    `json:"previous_block"`
	Version       int32     `json:"version"`
	Nonce         uint32    `json:"nonce"`
	Timestamp     time.Time `json:"timestamp"`
	Bits          uint32    `json:"bits"`
	MerkleRoot    string    `json:"merkle_root"`
}

type Transaction struct {
	gorm.Model

	ID uint `json:"id"`

	Hash     string `json:"hash"`
	LockTime uint32 `json:"lock_time"`
	Version  int32  `json:"version"`
	Safe     bool   `json:"safe"`

	BlockID    uint   `json:"block_id"`
	BlockHash  string `json:"block_hash"`
	BlockIndex uint32 `json:"block_index"`
}

type OutPoint struct {
	gorm.Model

	ID uint `json:"id"`

	SpendingTxID    uint   `json:"spending_tx_id"`
	SpendingTxHash  string `json:"spending_tx_hash"`
	SpendingTxIndex uint32 `json:"spending_tx_index"`
	Sequence        uint32 `json:"sequence"`
	SignatureScript string `json:"signature_script"`
	Witness         string `json:"witness"`

	FundingTxID    uint   `json:"funding_tx_id"`
	FundingTxHash  string `json:"funding_tx_hash"`
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

	err = db.AutoMigrate(&Block{}, &Transaction{}, &OutPoint{})
	if err != nil {
		panic(err)
	}

	return db, nil
}
