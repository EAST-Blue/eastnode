package main

import (
	"eastnode/indexer/peer"
	"eastnode/indexer/store"
	storeDB "eastnode/utils/store"

	"github.com/btcsuite/btcd/chaincfg"
)

func main() {
	s := storeDB.GetInstance(storeDB.IndexerDB)

	str := store.NewStorage(&chaincfg.RegressionNetParams, s.Gorm)
	p, err := peer.NewPeer("localhost:18444", str)
	if err != nil {
		panic(err)
	}
	err = p.Run()
	if err != nil {
		panic(err)
	}
}
