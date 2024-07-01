package main

import (
	"eastnode/indexer/peer"
	"eastnode/indexer/store"
	storeDB "eastnode/utils/store"

	"github.com/btcsuite/btcd/chaincfg"
	// "eastnode/indexer/repository"
	// storeDB "eastnode/utils/store"
	// "fmt"
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

	// s := storeDB.GetInstance(storeDB.IndexerDB)
	// indexerRepo := repository.NewIndexerRepository(s.Gorm)
	//
	// outpoints, err := indexerRepo.GetOutpointsByTransactionHash("19b0b492b03bfec6a318069706d087f9fffd247b7eb625af18305b7ae6f241be")
	// if err != nil {
	// 	panic(err)
	// }
	//
	// for _, v := range outpoints {
	// 	fmt.Println(v.PkAsmScripts)
	// }
}
