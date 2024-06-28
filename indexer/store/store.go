package store

import (
	"eastnode/indexer/peer"

	"github.com/btcsuite/btcd/chaincfg"
	"gorm.io/gorm"
)

// TODO: test reorgs
// TODO: test pending transactions

type Storage interface {
	peer.Storage
}

type storage struct {
	params *chaincfg.Params
	db     *gorm.DB
}

func NewStorage(params *chaincfg.Params, db *gorm.DB) Storage {
	return &storage{
		params: params,
		db:     db,
	}
}
